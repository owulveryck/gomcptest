package gcp

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path"

	"github.com/google/uuid"
	"golang.org/x/oauth2/google"
)

type imagenAPI struct {
	// ex https://europe-west9-aiplatform.googleapis.com/v1/projects/%s/locations/europe-west9/publishers/google/models/imagen-3.0-generate-002:predict
	endpoint    string
	credentials *google.Credentials
}

func newImagenAPI(ctx context.Context, config Configuration, model string) (*imagenAPI, error) {
	credentials, err := google.FindDefaultCredentials(ctx, "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		return nil, err
	}
	return &imagenAPI{
		endpoint:    fmt.Sprintf("https://%s-aiplatform.googleapis.com/v1/projects/%s/locations/%s/publishers/google/models/%s:predict", config.GCPRegion, config.GCPProject, config.GCPRegion, model),
		credentials: credentials,
	}, nil
}

func (i *imagenAPI) generateImage(ctx context.Context, prompt string, imageBaseDir string) (string, error) {
	slog.Debug("generating image", "endpoint", i.endpoint, "prompt", prompt)
	var ret string
	requestBody := map[string]interface{}{
		"instances": []map[string]interface{}{
			{
				"prompt": prompt,
			},
		},
		"parameters": map[string]interface{}{
			"sampleCount": 1,
		},
	}
	requestBodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return ret, errors.New("marshaling request body: %w" + err.Error())
	}

	// 3. Compress the request body with gzip.
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	_, err = zw.Write(requestBodyBytes)
	if err != nil {
		return ret, errors.New("compressing request body: %w" + err.Error())
	}
	if err := zw.Close(); err != nil {
		return ret, errors.New("closing gzip writer: %w" + err.Error())
	}

	token, err := i.credentials.TokenSource.Token()
	if err != nil {
		return ret, errors.New("getting access token: %w" + err.Error())
	}

	// 5. Make the HTTP request.
	req, err := http.NewRequest("POST", i.endpoint, &buf) // Use the gzip-compressed buffer
	if err != nil {
		return ret, errors.New("creating request: %w" + err.Error())
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Content-Encoding", "gzip") // Set the Content-Encoding header

	client := &http.Client{}
	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return ret, errors.New("making request: %w" + err.Error())
	}
	defer resp.Body.Close()

	// Handle potential gzip encoding of the response
	var reader io.ReadCloser
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewReader(resp.Body)
		if err != nil {
			return ret, errors.New("creating gzip reader: %w" + err.Error())
		}
		defer reader.Close()
	default:
		reader = resp.Body
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		return ret, errors.New("reading response body: %w" + err.Error())
	}

	// 6. Parse the response.  Using a generic map to avoid defining structs.
	var responseData map[string]interface{}
	err = json.Unmarshal(body, &responseData)
	if err != nil {
		return ret, errors.New("unmarshaling response: %w" + err.Error())
	}

	predictions, ok := responseData["predictions"].([]interface{})
	if !ok || len(predictions) == 0 {
		return ret, errors.New(" No predictions found in response")
	}

	prediction, ok := predictions[0].(map[string]interface{})
	if !ok {
		return ret, fmt.Errorf(" Invalid prediction format %T", predictions[0])
	}

	base64Encoded, ok := prediction["bytesBase64Encoded"].(string)
	if !ok {
		return ret, errors.New(" bytesBase64Encoded not found or invalid format")
	}
	imageData, err := base64.StdEncoding.DecodeString(base64Encoded)
	if err != nil {
		return ret, fmt.Errorf("error decoding base64 image: %w", err)
	}

	imageID := uuid.New()
	imageName := imageID.String() + ".png"
	imagePath := path.Join(imageBaseDir, imageName)

	err = os.WriteFile(imagePath, imageData, 0644)
	if err != nil {
		return ret, fmt.Errorf("error saving image to file: %w", err)
	}

	return "/images/" + imageName, nil
}
