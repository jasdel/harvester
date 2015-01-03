package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLooksLikeImageURL(t *testing.T) {
	kind := guessURLsMime("https://www.google.com/something.jpg")
	assert.Equal(t, "image/jpeg", kind, "Expect kind to match jpg image.")

	kind = guessURLsMime("http://ecx.images-amazon.com/images/I/41YNP8xxwsL._AC_SX75_.jpg")
	assert.Equal(t, "image/jpeg", kind, "Expect kind to match jpg image.")

	kind = guessURLsMime("http://ecx.images-amazon.com/images/I/41YNP8xxwsL._AC_SX75_.css")
	assert.Equal(t, "text/css", kind, "Expect kind to match jpg image.")

	kind = guessURLsMime("http://ecx.images-amazon.com/images/I/41YNP8xxwsL._AC_SX75_.js")
	assert.Equal(t, "text/javascript", kind, "Expect kind to match jpg image.")
}
