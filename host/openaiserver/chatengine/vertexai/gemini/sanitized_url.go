package gemini

import (
	"fmt"
	"net/url"
	"sort"
	"strings"
)

// sanitizeURL parses a raw URL string, sanitizes its path and query parameters
// ensuring spaces are encoded as %20, and returns the reconstructed URL string.
func sanitizeURL(rawURL string) (string, error) {
	// 1. Parse the raw URL
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse URL: %w", err)
	}

	// 2. Sanitize Path: Re-encode the decoded path using PathEscape
	// PathEscape encodes spaces as %20 and handles other path-specific chars.
	sanitizedPath := url.PathEscape(u.Path)
	// If the original path was empty or just "/", PathEscape might return empty.
	// url.Parse usually ensures Path starts with "/", preserve that if needed.
	if strings.HasPrefix(u.Path, "/") && !strings.HasPrefix(sanitizedPath, "/") && sanitizedPath != "" {
		// This check might be needed depending on how strict you need to be
		// For simple cases url.PathEscape usually does the right thing.
		// If u.Path was "/" url.PathEscape("/") is "/", so it's often fine.
		// If u.Path was "/ /", url.PathEscape gives "/%20/", which is correct.
	}
	if u.Path == "/" && sanitizedPath == "" { // Handle root path explicitly if needed
		sanitizedPath = "/"
	}

	// 3. Sanitize Query Parameters: Manually re-encode using QueryEscape
	// We cannot use u.Query().Encode() because it uses '+' for spaces.
	queryParams := u.Query() // Decodes RawQuery (treats + and %20 as space)
	var sanitizedQuery strings.Builder

	// Sort keys for deterministic output (optional but good practice)
	keys := make([]string, 0, len(queryParams))
	for k := range queryParams {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for i, k := range keys {
		values := queryParams[k]
		// QueryEscape encodes spaces as %20 and other necessary chars.
		escapedKey := strings.ReplaceAll(url.QueryEscape(k), "+", "%20")
		for j, v := range values {
			if i > 0 || j > 0 { // Add "&" separator before subsequent pairs/values
				sanitizedQuery.WriteString("&")
			}
			escapedValue := strings.ReplaceAll(url.QueryEscape(v), "+", "%20")
			sanitizedQuery.WriteString(escapedKey)
			sanitizedQuery.WriteString("=")
			sanitizedQuery.WriteString(escapedValue)
		}
	}

	// 4. Reconstruct the URL
	// Create a new URL struct with the sanitized components
	reconstructedURL := url.URL{
		Scheme:   u.Scheme,
		Opaque:   u.Opaque,                // Usually empty for standard URLs
		User:     u.User,                  // Userinfo (username:password)
		Host:     u.Host,                  // Hostname and port
		Path:     sanitizedPath,           // Use the re-encoded path
		RawPath:  "",                      // Let the String() method use the Path field
		RawQuery: sanitizedQuery.String(), // Use the manually built query string
		Fragment: u.Fragment,              // Keep the original fragment
		// RawFragment could also be considered if strict % encoding needed there too
	}

	return reconstructedURL.String(), nil
}
