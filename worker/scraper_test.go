package main

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"
)

func TestScrapValidateContent(t *testing.T) {
	mockBody := ioutil.NopCloser(bytes.NewBuffer([]byte("body content")))

	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       mockBody,
	}
	resp.Header.Set("Content-Type", "text/html")

	mime, body, err := validateContent(resp)

	require.Nil(t, err, "Expect no validation error")
	assert.Equal(t, "text/html", mime, "Expected mime to match")
	assert.Equal(t, "body content", string(body), "Expect body to match")
}

func TestScrapValidateContentInvalid(t *testing.T) {
	mockBody := ioutil.NopCloser(bytes.NewBuffer(nil))

	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       mockBody,
	}
	resp.Header.Set("Content-Type", "")

	mime, body, err := validateContent(resp)

	require.Nil(t, err, "Expect no validation error")
	assert.Equal(t, "application/octet-stream", mime, "Expected mime to be subsituted.")
	assert.Len(t, body, 0, "Expect body to be empty")
}

func TestNomralizeURL(t *testing.T) {
	origin, _ := url.Parse("https://example.come/blah/blah")

	u, err := normalizeURL(origin, "http://www.google.com/first/second")
	assert.Nil(t, err, "No error")
	assert.Equal(t, "http://www.google.com/first/second", u, "Expect URLs to match.")

	u, err = normalizeURL(origin, "//example.com/sports")
	assert.Nil(t, err, "No error")
	assert.Equal(t, "https://example.com/sports", u, "Expect URLs to match.")

	u, err = normalizeURL(origin, "sports.png")
	assert.Nil(t, err, "No error")
	assert.Equal(t, "https://example.come/blah/blah/sports.png", u, "Expect URLs to match.")

	u, err = normalizeURL(origin, "sports/")
	assert.Nil(t, err, "No error")
	assert.Equal(t, "https://example.come/blah/blah/sports/", u, "Expect URLs to match.")

	u, err = normalizeURL(origin, "data:image/jpeg;base64,/9j/4AAQSkZJRgABAQAAAQABAAD/2wBDAAoHBwgH")
	assert.NotNil(t, err, "Data URI should be reject")
}
