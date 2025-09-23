package urls

import "fmt"

const (
	defaultAPIURL    = "https://api.github.com"
	defaultUploadURL = "https://uploads.github.com"
)

func Get(baseURL, uploadURL string) (string, string, error) {
	if baseURL != "" && uploadURL != "" {
		return baseURL, uploadURL, nil
	}

	if baseURL == "" && uploadURL == "" {
		return defaultAPIURL, defaultUploadURL, nil
	}

	if baseURL == "" && uploadURL != "" {
		return "", "", fmt.Errorf("base must be specified if upload is specified")
	}

	// This is a guess based on a comment in github.NewEnterpriseClient
	// which indicates that uploadURL is often the baseURL for enterprise
	// installs. If this turns out to not be true, this should be an error.
	return baseURL, baseURL, nil

}
