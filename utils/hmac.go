package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"

	"eka-dev.cloud/master-data/config"
	"eka-dev.cloud/master-data/utils/response"
	"github.com/gofiber/fiber/v2/log"
)

func GenerateHMAC(data interface{}) (string, error) {
	dataJson, err := json.Marshal(data)
	if err != nil {
		log.Error("Failed to marshal data for HMAC generation:", err)
		return "", response.InternalServerError("Internal Server Error", nil)
	}
	h := hmac.New(sha256.New, []byte(config.Config.Secret))
	h.Write(dataJson)
	return string(h.Sum(nil)), nil
}

// VerifySignature memeriksa apakah signature valid
func VerifySignature(message string, signatureHeader string) error {
	// Decode base64 dari header
	signatureBytes, err := base64.StdEncoding.DecodeString(signatureHeader)
	if err != nil {
		log.Error("Failed to decode signature:", err)
		return response.InternalServerError("failed to decode signature", nil)
	}

	// Buat ulang signature dari body/message
	mac := hmac.New(sha256.New, []byte(config.Config.Secret))
	mac.Write([]byte(message))
	expectedMAC := mac.Sum(nil)

	// Bandingkan dengan waktu konstan (aman dari timing attack)
	if !hmac.Equal(signatureBytes, expectedMAC) {
		return response.Unauthorized("invalid signature", nil)
	}

	return nil
}
