package main

import (
	"github.com/jasdel/harvester/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteJSON(t *testing.T) {
	w := httptest.NewRecorder()

	err := writeJSON(w, []string{"a", "b", "c"}, http.StatusRequestURITooLong)
	require.Nil(t, err, "No Error should be received")
	assert.Equal(t, http.StatusRequestURITooLong, w.Code, "Code should be set")
	assert.Equal(t, "[\"a\",\"b\",\"c\"]\n", w.Body.String(), "Body should be json encoded")

}

func TestJobIdFromString(t *testing.T) {
	id, err := jobIdFromString("hello")
	assert.NotNil(t, err, "Not valid job id")

	id, err = jobIdFromString("12345")
	assert.Nil(t, err, "Valid job id")
	assert.Equal(t, types.JobId(12345), id, "Correct job id decoded")
}
