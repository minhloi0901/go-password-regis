# API docs

## Overview
A basic single HTTP server to handle registration, login, health, verify-email, resend-verficiation endpoints.

Base URL (local): http://localhost:8080

## Architecture

```
[Frontend] → HTTP → [Registration Server] → gRPC → [Credential Service]
                        ↓
                    [Temporal Worker]
                    ├── Activity: CreateCredential
                    ├── Activity: SendVerificationEmail
                    └── Signal: EmailVerified → ActivateAccount
```

## Endpoints
| Method | Path                 | Description                                                          |
| ------ | -------------------- | -------------------------------------------------------------------- |
| POST   | /register            | Register a new prospect                                              |
| POST   | /login               | Login account by verifying username and password through credentials |
| GET    | /health              | Server liveness check                                                |
| POST   | /verify-email        | Confirm prospect's email by sending code                             |
| POST   | /resend-verification | Resend a new verification code                                       |

### `POST /register`
**Request**:
```
{
    "username": "rudeus",
    "email": "duong.loi@hcltech.com",
    "password": "AbC123456",
}
```
**Response**: `201 Created`
```
{
    "id": "dml-uuid-123",
    "status": "pending",
}
```
**Response**: `400 Bad Request` (validation error - missing required field, invalid username, invalid password, weak password...)
```
{
    "error": "bad_request",
    "message": "username require more than 6 letters",
}
```
**Response**: `409 Conflict` (already exists - username or email is already used)
```
{
    "error": "existed_username",
    "message": "username is already used",
}
```

### `POST /login`
**Request**:
```
{
    "username": "dml_123",
    "password": "AbC123456",
}
```
**Response**: `200 OK`
```
{
    "user": {
        "id": "dml_123"
        "email": "duong.loi@hcltech.com"
        "username": "rudeus"
    },
    "access_token": "<access_token>"
    "refresh_token": "<refresh_token>"
    "expires_in": 3600
}
```
**Response**: `400 Bad Request` (missing required field - email or password)
```
{
    "error": "bad_request",
    "message": "please fill in both emai and password",
}
```
**Response**: `401 Unauthorized` (invalid authentication - invalid username or password)
```
{
    "error": "invalid_credentials",
    "message": "invalid username or password",
}
```
**Response**: `403 Forbidden` 
```
{
    "error": "account_not_verified",
    "message": "account not active yet, please verify email first",
}
```

### `GET /health`
**Request**: no body

**Response**: `200 OK`
```
{
    "status": "healthy 100%"
}
```

### `POST /verify-email`
**Request**:
```
{
    "email": "duong.loi@hcltech.com"
    "code": "806499"
}
```
**Response**: `200 OK`
```
{
    "id": "dml_123"
    "status": "active"
}
```
**Response**: `400 Bad Request` (wrong code or expired)
```
{
    "error": "bad_request",
    "message": "please fill in both emai and password",
}
```
### `POST /resend-verification`