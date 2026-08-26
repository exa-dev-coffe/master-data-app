package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"log/slog"

	"eka-dev.cloud/master-data/config"
	"eka-dev.cloud/master-data/utils/response"
)

func GenerateHMAC(data interface{}) (string, error) {
	dataJson, err := json.Marshal(data)
	if err != nil {
		slog.Error("Failed to marshal data for HMAC generation", "error", err)
		return "", response.InternalServerError("Internal Server Error", nil)
	}
	h := hmac.New(sha256.New, []byte(config.Config.Secret))
	h.Write(dataJson)
	return string(h.Sum(nil)), nil
}

func CreateSignature(queryString, bodyString, timestamp string) (string, error) {
	message := queryString + timestamp + bodyString
	mac := hmac.New(sha256.New, []byte(config.Config.Secret))
	mac.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil)), nil
}

// VerifySignature checks if signature is valid
func VerifySignature(message string, signatureHeader string) error {
	signatureBytes, err := base64.StdEncoding.DecodeString(signatureHeader)
	if err != nil {
		slog.Error("Failed to decode signature", "error", err)
		return response.InternalServerError("failed to decode signature", nil)
	}

	mac := hmac.New(sha256.New, []byte(config.Config.Secret))
	mac.Write([]byte(message))
	expectedMAC := mac.Sum(nil)

	if !hmac.Equal(signatureBytes, expectedMAC) {
		return response.Unauthorized("invalid signature", nil)
	}

	return nil
}
