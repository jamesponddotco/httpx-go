package httpx_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"testing"

	"git.sr.ht/~jamesponddotco/httpx-go"
)

type TestStruct struct {
	Slideshow TestSlideshow `json:"slideshow"`
}

type TestSlides struct {
	Title string   `json:"title"`
	Type  string   `json:"type"`
	Items []string `json:"items,omitempty"`
}
type TestSlideshow struct {
	Author string       `json:"author"`
	Date   string       `json:"date"`
	Title  string       `json:"title"`
	Slides []TestSlides `json:"slides"`
}

// errReader simulates a read error when reading the response body.
type errReader struct{}

func (*errReader) Read(_ []byte) (n int, err error) {
	return 0, io.ErrUnexpectedEOF
}

type customReadCloser struct {
	data *bytes.Buffer
}

func (c *customReadCloser) Read(p []byte) (n int, err error) {
	return c.data.Read(p)
}

func (*customReadCloser) Close() error {
	return errors.New("mock close error")
}

func TestReadJSON(t *testing.T) {
	t.Parallel()

	var (
		client = httpx.NewClientWithCache(nil)
		header = map[string]string{
			"Accept": "application/json",
		}
	)

	tests := []struct {
		name    string
		give    string
		resp    *http.Response
		want    TestStruct
		wantErr bool
	}{
		{
			name: "valid JSON without request",
			resp: &http.Response{
				Body: io.NopCloser(bytes.NewReader([]byte(`{"slideshow":{"author":"Yours Truly","date":"date of publication","slides":[{"title":"Wake up to WonderWidgets!","type":"all"},{"title":"Overview","type":"all","items":["Why <em>WonderWidgets</em> are great","Who <em>buys</em> WonderWidgets"]}],
"title":"Sample Slide Show"}}`))),
			},
			want: TestStruct{
				Slideshow: TestSlideshow{
					Author: "Yours Truly",
					Date:   "date of publication",
					Slides: []TestSlides{
						{
							Title: "Wake up to WonderWidgets!",
							Type:  "all",
						},
						{
							Items: []string{
								"Why <em>WonderWidgets</em> are great",
								"Who <em>buys</em> WonderWidgets",
							},
							Title: "Overview",
							Type:  "all",
						},
					},
					Title: "Sample Slide Show",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid JSON",
			resp: &http.Response{
				Body: io.NopCloser(bytes.NewReader([]byte(`{"invalid": "json"`))),
			},
			want:    TestStruct{},
			wantErr: true,
		},
		{
			name: "empty JSON",
			resp: &http.Response{
				Body: io.NopCloser(bytes.NewReader([]byte(``))),
			},
			want:    TestStruct{},
			wantErr: true,
		},
		{
			name: "read error",
			resp: &http.Response{
				Body: io.NopCloser(&errReader{}),
			},
			want:    TestStruct{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var got TestStruct

			if tt.give != "" {
				req, err := httpx.NewRequest(context.Background(), http.MethodGet, tt.give, header, http.NoBody)
				if err != nil {
					t.Fatal(err)
				}

				resp, err := client.Do(context.Background(), req)
				if err != nil {
					t.Fatal(err)
				}
				defer resp.Body.Close()

				tt.resp = resp
			}

			err := httpx.ReadJSON(tt.resp, &got)

			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}

			if !tt.wantErr && err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got: %v, want: %v", got, tt.want)
			}
		})
	}
}

func TestDrainResponseBody(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		resp    *http.Response
		wantErr bool
	}{
		{
			name: "valid response body",
			resp: &http.Response{
				Body: io.NopCloser(bytes.NewReader([]byte("valid response body"))),
			},
			wantErr: false,
		},
		{
			name: "empty response body",
			resp: &http.Response{
				Body: io.NopCloser(bytes.NewReader([]byte(""))),
			},
			wantErr: false,
		},
		{
			name: "error response body",
			resp: &http.Response{
				Body: io.NopCloser(&errReader{}),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := httpx.DrainResponseBody(tt.resp)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}

			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestIsSuccess(t *testing.T) {
	t.Parallel()

	client := httpx.NewClientWithCache(nil)

	tests := []struct {
		name string
		give string
		want bool
	}{
		{
			name: "200",
			give: "https://httpstat.us/200",
			want: true,
		},
		{
			name: "400",
			give: "https://httpstat.us/400",
			want: false,
		},
		{
			name: "500",
			give: "https://httpstat.us/500",
			want: false,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req, err := client.Get(context.Background(), tt.give)
			if err != nil {
				t.Fatal(err)
			}

			got := httpx.IsSuccess(req)
			if got != tt.want {
				t.Errorf("got: %v, want: %v", got, tt.want)
			}

			req.Body.Close()
		})
	}
}

func TestDrainResponseBody_ErrorClose(t *testing.T) {
	t.Parallel()

	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       &customReadCloser{data: bytes.NewBufferString("test")},
	}

	err := httpx.DrainResponseBody(resp)
	if err == nil {
		t.Error("expected error, got nil")
	}

	want := fmt.Errorf("%w: %w", httpx.ErrCannotCloseResponse, errors.New("mock close error"))
	if err.Error() != want.Error() {
		t.Errorf("got: %v, want: %v", err, want)
	}
}
