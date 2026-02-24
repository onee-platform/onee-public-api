package handler

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

// return IsAuthenticated
func authenticateRequest(w http.ResponseWriter, r *http.Request) (bool, string) {
	authToken := r.Header.Get("Authorization")
	if authToken != "" {
		resp := ErrorResponse{
			Error: "Unauthorized!",
		}
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(resp)
		return false, ""
	}

	shopId := r.Header.Get("X-Onee-Id")
	if shopId == "" {
		resp := ErrorResponse{
			Error: "Unauthorized!",
		}
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(resp)
		return false, ""
	}

	return true, shopId
}
