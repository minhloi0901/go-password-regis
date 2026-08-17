package prospect

import (
	"crypto/rand"
	"crypto/subtle"
	"log"
	"math/big"
	"time"
)

// use crypto/rand instead of math/rand (entropy pool with sys call)
func GenerateVerificationCode() (string, error) {
	code := make([]byte, VerificationCodeLength)

	for i := range code {
		randNum, err := rand.Int(rand.Reader, big.NewInt(int64(len(VerificationCodeDigits))))
		if err != nil {
			return "", err
		}
		code[i] = VerificationCodeDigits[randNum.Int64()]
	}

	return string(code), nil
}

func ValidateVerificationCode(p Prospect, code string) error {
	// check valid
	if p.VerificationCode == "" || p.CodeExpiresAt.IsZero() {
		log.Println("Verify Email - Validate Code: invalid code in prospect")
		return ErrInvalidCode
	}

	// code already expired
	if time.Now().After(p.CodeExpiresAt) {
		return ErrCodeExpired
	}

	// check matching
	if subtle.ConstantTimeCompare([]byte(p.VerificationCode), []byte(code)) != 1 {
		log.Println("Verify Email - Validate Code: code not match")
		return ErrInvalidCode
	}
	return nil
}
