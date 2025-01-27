package vertexai

import (
	"context"
	"fmt"

	"golang.org/x/oauth2/google"
)

func getAccessToken() (string, error) {
	// Use the default application credentials
	ctx := context.Background()
	credentials, err := google.FindDefaultCredentials(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to find default credentials: %w", err)
	}

	// Get the token from the credentials
	token, err := credentials.TokenSource.Token()
	if err != nil {
		return "", fmt.Errorf("failed to retrieve token: %w", err)
	}

	return token.AccessToken, nil
}
