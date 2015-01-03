package main

import (
	"encoding/json"
	"fmt"
	"github.com/jasdel/harvester/internal/types"
	"net/http"
	"strconv"
)

type ErrorRsp struct {
	Code string `json:"code"`
	Msg  string `json:"message"`
}

// Encodes the response as a JSON object, and writes it back to the client.
func writeJSON(w http.ResponseWriter, data interface{}, status int) error {
	w.WriteHeader(status)
	e := json.NewEncoder(w)
	return e.Encode(data)
}

// Encodes an error message as a JSON object, and writes it back to the client
func writeJSONError(w http.ResponseWriter, code, msg string, status int) error {
	return writeJSON(w, ErrorRsp{Code: code, Msg: msg}, status)
}

// Converts a string into a Job ID validating that it is a valid value
func jobIdFromString(idStr string) (types.JobId, error) {
	if idStr == "" {
		return types.InvalidJobId, fmt.Errorf("No jobId provided")
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return types.InvalidJobId, fmt.Errorf("Invalid jobId: %s", idStr)
	}

	return types.JobId(id), nil
}
