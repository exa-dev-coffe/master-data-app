package upload

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log/slog"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"eka-dev.cloud/master-data/lib"
	"eka-dev.cloud/master-data/utils/response"
	"github.com/deepteams/webp"
	"github.com/google/uuid"
)

type Service interface {
	UploadMenuFoto(fileHeader *multipart.FileHeader) (*UploadResponse, error)
	DeleteMenuFoto(fileName string) error
}

type uploadService struct {
}

func NewUploadService() Service {
	return &uploadService{}
}

func generateFileName(original string, targetExt string) (string, error) {
	extension := targetExt
	if extension == "" {
		extension = filepath.Ext(original)
	}
	if !strings.HasPrefix(extension, ".") {
		extension = "." + extension
	}

	u, err := uuid.NewRandom()
	if err != nil {
		slog.Error("Failed to generate UUID for file name", "error", err)
		return "", response.InternalServerError("Failed to generate file name", nil)
	}

	fileName := fmt.Sprintf("coffe/images/menus/%s-%s%s",
		time.Now().Format("20060102150405"),
		u.String(),
		extension,
	)

	return fileName, nil
}

func extractFileNameFromURL(url string) (string, error) {
	if strings.Contains(url, "project/") {
		parts := strings.SplitAfter(url, "project/")
		if len(parts) < 2 {
			return "", response.BadRequest("Invalid URL format", nil)
		}
		return parts[1], nil
	} else {
		return "", response.BadRequest("URL does not contain expected segment", nil)
	}
}

func (s *uploadService) UploadMenuFoto(fileHeader *multipart.FileHeader) (*UploadResponse, error) {
	file, err := fileHeader.Open()
	if err != nil {
		slog.Error("Failed to open uploaded file", "error", err)
		return nil, response.InternalServerError("Failed to open file", nil)
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		slog.Error("Failed to read uploaded file", "error", err)
		return nil, response.InternalServerError("Failed to read file", nil)
	}

	contentType := fileHeader.Header.Get("Content-Type")
	isWebP := contentType == "image/webp" || strings.EqualFold(filepath.Ext(fileHeader.Filename), ".webp")

	var url string
	if isWebP {
		// File already converted to WebP (e.g. by frontend)
		fileName, err := generateFileName(fileHeader.Filename, ".webp")
		if err != nil {
			return nil, err
		}
		url, err = lib.UploadBytes(fileName, fileBytes, "image/webp")
		if err != nil {
			return nil, err
		}
	} else {
		// Try to decode image and convert to compressed WebP
		img, _, decodeErr := image.Decode(bytes.NewReader(fileBytes))
		if decodeErr == nil {
			var webpBuf bytes.Buffer
			encodeErr := webp.Encode(&webpBuf, img, &webp.Options{Quality: 82})
			if encodeErr == nil {
				fileName, err := generateFileName(fileHeader.Filename, ".webp")
				if err != nil {
					return nil, err
				}
				url, err = lib.UploadBytes(fileName, webpBuf.Bytes(), "image/webp")
				if err != nil {
					return nil, err
				}
			} else {
				slog.Warn("Failed to encode WebP, falling back to original upload", "error", encodeErr)
			}
		} else {
			slog.Warn("Failed to decode image, falling back to raw upload", "error", decodeErr)
		}

		// Fallback if decode or encode failed (e.g., test mock headers)
		if url == "" {
			fileName, err := generateFileName(fileHeader.Filename, "")
			if err != nil {
				return nil, err
			}
			url, err = lib.UploadBytes(fileName, fileBytes, contentType)
			if err != nil {
				return nil, err
			}
		}
	}

	return &UploadResponse{
		URL: url,
	}, nil
}

func (s *uploadService) DeleteMenuFoto(url string) error {
	filepathFromUrl, err := extractFileNameFromURL(url)
	if err != nil {
		return err
	}
	err = lib.DeleteFile(filepathFromUrl)
	if err != nil {
		return err
	}
	return nil
}
