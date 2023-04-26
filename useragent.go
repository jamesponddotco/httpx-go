package httpx

import (
	"strings"

	"git.sr.ht/~jamesponddotco/httpx-go/internal/build"
	"git.sr.ht/~jamesponddotco/httpx-go/internal/separator"
)

const (
	_goClientUserAgent string = "Go-http-client/1.1"
)

// UserAgent represents the User-Agent header value, as defined in [RFC 7231,
// section 5.5.3].
//
// [RFC 7231, section 5.5.3]: https://tools.ietf.org/html/rfc7231#section-5.5.3
type UserAgent struct {
	// Token is the product token.
	Token string

	// Version is the product version.
	Version string

	// Comment is the product comment.
	Comment []string
}

// DefaultUserAgent returns the default User-Agent header value for the httpx
// package.
func DefaultUserAgent() *UserAgent {
	return &UserAgent{
		Token:   build.Name,
		Version: build.Version,
		Comment: []string{
			build.UserAgentURL,
			_goClientUserAgent,
		},
	}
}

// String returns the string representation of the User-Agent header value.
func (ua *UserAgent) String() string {
	if ua == nil || ua.Token == "" || ua.Version == "" {
		return ""
	}

	var builder strings.Builder

	token := ua.Token
	version := ua.Version

	estimatedLength := len(token) + len(version) + 2 // 2 is for the token-version separator

	for _, comment := range ua.Comment {
		estimatedLength += len(comment) + len(separator.Colon+separator.Space) + 2 // 2 is for the parentheses
	}

	builder.Grow(estimatedLength)

	builder.WriteString(token)
	builder.WriteString(separator.ForwardSlash)
	builder.WriteString(version)

	if len(ua.Comment) > 0 {
		builder.WriteString(separator.Space + separator.OpenParenthesis)

		for i, comment := range ua.Comment {
			if i > 0 {
				builder.WriteString(separator.Colon + separator.Space)
			}

			for _, r := range comment {
				if r != '(' && r != ')' {
					builder.WriteRune(r)
				}
			}
		}

		builder.WriteString(separator.CloseParenthesis)
	}

	return builder.String()
}
