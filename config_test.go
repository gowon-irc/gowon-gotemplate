package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseConfig(t *testing.T) {
	cases := []struct {
		name     string
		testFile string
		errMsg   string
	}{
		{
			name:     "valid config",
			testFile: "valid.yaml",
			errMsg:   "",
		},
		{
			name:     "not yaml",
			testFile: "invalid.txt",
			errMsg:   "line 1: cannot unmarshal !!str `abc` into main.Config",
		},
		{
			name:     "empty file",
			testFile: "empty.txt",
			errMsg:   "Commands: cannot be blank",
		},
		{
			name:     "empty yaml",
			testFile: "empty.yaml",
			errMsg:   "Commands: cannot be blank",
		},
		{
			name:     "missing fields",
			testFile: "missing_fields.yaml",
			errMsg:   "Commands: cannot be blank",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cf := openTestFile(t, "TestParseConfig", tc.testFile)

			_, err := parseConfig(cf)

			if tc.errMsg == "" {
				assert.Nil(t, err)
			} else {
				assert.ErrorContains(t, err, tc.errMsg)
			}
		})
	}
}
