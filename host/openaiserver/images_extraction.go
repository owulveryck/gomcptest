package main

import (
	"encoding/base64"
	"fmt"
	"strings"
)

// ExtractImageData extracts the mime type and decoded byte slice from a data URI string.
func ExtractImageData(dataURI string) (string, []byte, error) {
	// Split the data URI string into parts
	parts := strings.SplitN(dataURI, ",", 2)
	if len(parts) != 2 {
		return "", nil, fmt.Errorf("invalid data URI format")
	}

	metadata := parts[0]
	base64Data := parts[1]

	// Extract the mime type
	mimeTypeParts := strings.SplitN(metadata, ":", 2)
	if len(mimeTypeParts) != 2 {
		return "", nil, fmt.Errorf("invalid data URI metadata format")
	}
	mimeType := mimeTypeParts[1]

	// Extract the encoding type
	encodingParts := strings.SplitN(mimeType, ";", 2)
	if len(encodingParts) > 1 {
		mimeType = encodingParts[0]
	}

	// Decode the base64 data
	imgData, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return "", nil, fmt.Errorf("failed to decode base64  %w", err)
	}

	return mimeType, imgData, nil
}
