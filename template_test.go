package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

type RoundTripFunc func(req *http.Request) *http.Response

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

func NewTestClient(statusCode int, body string) *http.Client {
	f := func(req *http.Request) *http.Response {
		return &http.Response{
			StatusCode: statusCode,
			Body:       ioutil.NopCloser(bytes.NewBufferString(body)),
			Header:     make(http.Header),
		}
	}

	return &http.Client{
		Transport: RoundTripFunc(f),
	}
}

func openTestFile(t *testing.T, test, filename string) []byte {
	fp := filepath.Join("testdata", test, filename)
	out, err := ioutil.ReadFile(fp)

	if err != nil {
		t.Fatalf("failed to read test file: %s", err)
	}

	return out
}

func TestDownloadURL(t *testing.T) {
	cases := []struct {
		name       string
		testFile   string
		body       string
		statusCode int
		errMsg     string
	}{
		{
			name:       "Empty data",
			testFile:   "empty",
			body:       "",
			statusCode: 200,
			errMsg:     "",
		},
		{
			name:       "Non-empty data",
			testFile:   "text",
			body:       "text\n",
			statusCode: 200,
			errMsg:     "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			body := openTestFile(t, "TestDownloadUrl", tc.testFile)
			client := NewTestClient(tc.statusCode, string(body))

			got, err := downloadURL("", client)

			if tc.errMsg == "" {
				assert.Nil(t, err)
			} else {
				assert.ErrorContains(t, err, tc.errMsg)
			}

			assert.Equal(t, tc.body, got)
		})
	}
}

func TestTemplateParse(t *testing.T) {
	cases := []struct {
		name        string
		text        string
		templateMap map[string]interface{}
		expected    string
	}{
		{
			name:        "empty text, empty map",
			text:        "",
			templateMap: map[string]interface{}{},
			expected:    "",
		},
		{
			name:        "no tokens, empty map",
			text:        "text",
			templateMap: map[string]interface{}{},
			expected:    "text",
		},
		{
			name: "token, token in map",
			text: "{{ .a }}",
			templateMap: map[string]interface{}{
				"a": 1,
			},
			expected: "1",
		},
		{
			name:        "token, empty map",
			text:        "{{ .a }}",
			templateMap: map[string]interface{}{},
			expected:    "<no value>",
		},
		{
			name: "tokens, one in map",
			text: "{{ .a }} {{ .b }}",
			templateMap: map[string]interface{}{
				"a": 1,
			},
			expected: "1 <no value>",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := templateParse(tc.text, tc.templateMap)

			assert.Nil(t, err)
			assert.Equal(t, tc.expected, out)
		})
	}
}

func TestHandle(t *testing.T) {
	cases := []struct {
		name         string
		bodyFile     string
		statusCode   int
		templateFile string
		expected     string
		errMsg       string
	}{
		{
			name:         "valid json, valid template",
			bodyFile:     "abc.json",
			statusCode:   200,
			templateFile: "abc.tmpl",
			expected:     "1 2 3",
			errMsg:       "",
		},
		{
			name:         "invalid json",
			bodyFile:     "def.txt",
			statusCode:   200,
			templateFile: "def.tmpl",
			expected:     "",
			errMsg:       "invalid character 'd' looking for beginning of value",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			body := openTestFile(t, "TestHandle", tc.bodyFile)
			client := NewTestClient(tc.statusCode, string(body))

			templ := openTestFile(t, "TestHandle", tc.templateFile)

			got, err := handle("", string(templ), client)

			if tc.errMsg == "" {
				assert.Nil(t, err)
			} else {
				assert.ErrorContains(t, err, tc.errMsg)
			}

			assert.Equal(t, tc.expected, got)
		})
	}
}
