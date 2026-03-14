package apiurl

import "testing"

func TestBaseURL_Default(t *testing.T) {
	t.Setenv(BaseURLEnvVar, "")

	if got := BaseURL(); got != DefaultBaseURL {
		t.Fatalf("BaseURL() = %q, want %q", got, DefaultBaseURL)
	}
}

func TestBaseURL_Override(t *testing.T) {
	t.Setenv(BaseURLEnvVar, " http://localhost:8787/ ")

	if got := BaseURL(); got != "http://localhost:8787" {
		t.Fatalf("BaseURL() = %q, want %q", got, "http://localhost:8787")
	}
}

func TestResolveURLFromBase_RejectsNonHTTPScheme(t *testing.T) {
	t.Parallel()

	for _, scheme := range []string{"ftp://example.com", "file:///etc/passwd", "ssh://host"} {
		_, err := ResolveURLFromBase(scheme, "/path")
		if err == nil {
			t.Errorf("ResolveURLFromBase(%q, ...) = nil error, want scheme error", scheme)
		}
	}
}

func TestResolveURL(t *testing.T) {
	t.Setenv(BaseURLEnvVar, "http://localhost:8787/")

	got, err := ResolveURL("/oauth/device/code")
	if err != nil {
		t.Fatalf("ResolveURL() error = %v", err)
	}

	if got != "http://localhost:8787/oauth/device/code" {
		t.Fatalf("ResolveURL() = %q, want %q", got, "http://localhost:8787/oauth/device/code")
	}
}
