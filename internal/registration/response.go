package registration

import (
	"encoding/json"
	"log"
	"net/http"
)

// -- shared ---
type ErrorResponse struct {
	Error string `json:"error"`
}

func (b *BaseHandler) writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("handlers: encode JSON response: %v", err)
	}
}

func (b *BaseHandler) writeError(w http.ResponseWriter, status int, msg string) {
	b.writeJSON(w, status, ErrorResponse{Error: msg})
}
