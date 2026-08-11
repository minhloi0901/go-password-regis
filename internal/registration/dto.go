package registration

// --- POST /register ---

type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=5,max=20"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=80"`
}

type RegisterResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

// --- POST /login ---

type LoginRequest struct {
	Username string `json:"username" validate:"required,min=5,max=20"`
	// Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=80"`
}

type LoginResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

// --- GET /health ---
type HealthResponse struct {
	Status string `json:"status"`
}

// --- POST /verify-email ---
type VerifyEmailRequest struct {
	Email string `json:"email" validate:"required,email"`
	Code  string `json:"code" validate:"required,min=6,max=6"`
}

type VerifyEmailResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

// --- POST /resend-verification ---
type ResendVerificationRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ResendVerificationReponse struct {
	Message string `json:"message"`
}
