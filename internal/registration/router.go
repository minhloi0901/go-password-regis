package registration

import "net/http"

func NewRouter(rh *RegisterHandler) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /register", rh.HandleRegister)
	mux.HandleFunc("POST /login", rh.HandleLogin)
	mux.HandleFunc("GET /health", rh.HandleHealth)
	mux.HandleFunc("POST /verify-email", rh.HandleVerifyEmail)
	mux.HandleFunc("POST /resend-verification", rh.HandleResendVerification)

	return mux
}
