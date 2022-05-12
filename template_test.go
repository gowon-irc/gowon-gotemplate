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
		body := openTestFile(t, "TestDownloadUrl", tc.testFile)
		client := NewTestClient(tc.statusCode, string(body))

		got, err := downloadURL("", client)

		if tc.errMsg == "" {
			assert.Nil(t, err)
		} else {
			assert.ErrorContains(t, err, tc.errMsg)
		}

		assert.Equal(t, got, tc.body)
	}
}
