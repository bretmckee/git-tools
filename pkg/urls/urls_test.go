package urls

import (
	"testing"
)

func TestGet(t *testing.T) {
	tests := []struct {
		name          string
		baseURL       string
		uploadURL     string
		wantBaseURL   string
		wantUploadURL string
		wantErr       bool
	}{
		{
			name:          "Both URLs provided",
			baseURL:       "https://github.example.com/api/v3",
			uploadURL:     "https://github.example.com/api/uploads",
			wantBaseURL:   "https://github.example.com/api/v3",
			wantUploadURL: "https://github.example.com/api/uploads",
			wantErr:       false,
		},
		{
			name:          "Both URLs empty",
			baseURL:       "",
			uploadURL:     "",
			wantBaseURL:   defaultAPIURL,
			wantUploadURL: defaultUploadURL,
			wantErr:       false,
		},
		{
			name:          "Upload URL without base URL",
			baseURL:       "",
			uploadURL:     "https://github.example.com/api/uploads",
			wantBaseURL:   "",
			wantUploadURL: "",
			wantErr:       true,
		},
		{
			name:          "Base URL without upload URL",
			baseURL:       "https://github.example.com/api/v3",
			uploadURL:     "",
			wantBaseURL:   "https://github.example.com/api/v3",
			wantUploadURL: "https://github.example.com/api/v3",
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotBaseURL, gotUploadURL, err := Get(tt.baseURL, tt.uploadURL)

			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if gotBaseURL != tt.wantBaseURL {
				t.Errorf("Get() gotBaseURL = %v, want %v", gotBaseURL, tt.wantBaseURL)
			}
			if gotUploadURL != tt.wantUploadURL {
				t.Errorf("Get() gotUploadURL = %v, want %v", gotUploadURL, tt.wantUploadURL)
			}
		})
	}
}
