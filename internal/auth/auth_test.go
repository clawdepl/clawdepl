package auth

import "testing"

func TestDeriveAPIKeyUserID(t *testing.T) {
	tests := []struct {
		name   string
		apiKey string
		want   string
	}{
		{
			name:   "long API key uses first 8 chars",
			apiKey: "abcdefgh123456",
			want:   "user_api_abcdefgh",
		},
		{
			name:   "short API key does not panic and uses full key",
			apiKey: "short",
			want:   "user_api_short",
		},
		{
			name:   "empty API key produces base prefix",
			apiKey: "",
			want:   "user_api_",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := deriveAPIKeyUserID(tt.apiKey)
			if got != tt.want {
				t.Fatalf("deriveAPIKeyUserID(%q) = %q, want %q", tt.apiKey, got, tt.want)
			}
		})
	}
}
