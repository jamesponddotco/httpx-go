package httpx_test

import (
	"testing"

	"git.sr.ht/~jamesponddotco/httpx-go"
)

func TestUserAgent_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		userAgent *httpx.UserAgent
		want      string
	}{
		{
			name:      "default user agent",
			userAgent: httpx.DefaultUserAgent(),
			want:      httpx.DefaultUserAgent().String(),
		},
		{
			name: "custom user agent with no comment",
			userAgent: &httpx.UserAgent{
				Token:   "Custom",
				Version: "1.0.0",
			},
			want: "Custom/1.0.0",
		},
		{
			name: "custom user agent with comments",
			userAgent: &httpx.UserAgent{
				Token:   "Custom",
				Version: "1.0.0",
				Comment: []string{"comment1", "comment2"},
			},
			want: "Custom/1.0.0 (comment1; comment2)",
		},
		{
			name: "custom user agent without token or version",
			userAgent: &httpx.UserAgent{
				Comment: []string{"comment1", "comment2"},
			},
			want: "",
		},
		{
			name: "custom user agent with invalid comment",
			userAgent: &httpx.UserAgent{
				Token:   "Custom",
				Version: "1.0.0",
				Comment: []string{"(invalid)"},
			},
			want: "Custom/1.0.0 (invalid)",
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.userAgent.String()
			if got != tt.want {
				t.Errorf("UserAgent.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
