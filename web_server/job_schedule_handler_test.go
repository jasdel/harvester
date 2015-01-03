package main

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestGetRequestedJobURLs(t *testing.T) {
	reader := strings.NewReader(`https://www.google.com

example.com

http://www.reddit.com
`)
	urls, err := getRequestedJobURLs(reader)
	require.Nil(t, err, "Expect no error")
	assert.Len(t, urls, 3, "Expect lengths to match")
	assert.Equal(t, `https://www.google.com`, urls[0], "URL entry should match")
	assert.Equal(t, `http://example.com`, urls[1], "URL entry should match")
	assert.Equal(t, `http://www.reddit.com`, urls[2], "URL entry should match")
}

func TestGetRequestedJobURLsFail(t *testing.T) {
	reader := strings.NewReader(`/something/not/a/URL`)
	urls, err := getRequestedJobURLs(reader)
	assert.NotNil(t, err, "Expected error to be found")
	assert.Len(t, urls, 0, "Expect no URLs returned")
}

type validateTestCase struct {
	in  string
	out string
	err bool
}

var validateTestCases = []validateTestCase{
	validateTestCase{in: `https://www.google.com`, out: `https://www.google.com`, err: false},
	validateTestCase{in: `http://example.com`, out: `http://example.com`, err: false},
	validateTestCase{in: `reddit.com`, out: `http://reddit.com`, err: false},
	validateTestCase{in: `/something/else`, out: ``, err: true},
}

func TestValidateJobURL(t *testing.T) {
	for _, c := range validateTestCases {
		o, err := validateJobURL(c.in)

		if c.err {
			require.NotNil(t, err, "Validate should fail %s", c.in)
		} else {
			require.Nil(t, err, "Validate should succeed %s", c.in)
		}

		assert.Equal(t, c.out, o, "Expect values to match")
	}
}
