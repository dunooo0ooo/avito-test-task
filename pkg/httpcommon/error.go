package httpcommon

import (
	"encoding/json"
	"net/http"
)

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type WrappedErrorResponse struct {
	Error ErrorBody `json:"error"`
}

func JSONError(w http.ResponseWriter, statusCode int, code, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	resp := WrappedErrorResponse{
		Error: ErrorBody{
			Code:    code,
			Message: msg,
		},
	}

	_ = json.NewEncoder(w).Encode(resp)
}
