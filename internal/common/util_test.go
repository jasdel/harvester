package common

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

// Verifies GuessURLsMime works as expected
func TestGuessURLsMime(t *testing.T) {
	kind := GuessURLsMime("https://www.google.com/something.jpg")
	assert.Equal(t, "image/jpeg", kind, "Expect kind to match jpg image.")

	kind = GuessURLsMime("http://ecx.images-amazon.com/images/I/41YNP8xxwsL._AC_SX75_.jpg")
	assert.Equal(t, "image/jpeg", kind, "Expect kind to match jpg image.")

	kind = GuessURLsMime("http://ecx.images-amazon.com/images/I/41YNP8xxwsL._AC_SX75_.css")
	assert.Equal(t, "text/css", kind, "Expect kind to match css.")

	kind = GuessURLsMime("http://ecx.images-amazon.com/images/I/41YNP8xxwsL._AC_SX75_.js")
	assert.Equal(t, "text/javascript", kind, "Expect kind to match javascript.")

	kind = GuessURLsMime("http://ecx.images-amazon.com/")
	assert.Equal(t, "text/html", kind, "Expect kind to match html page.")
}
