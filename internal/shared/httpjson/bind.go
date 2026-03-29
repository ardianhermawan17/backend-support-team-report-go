package httpjson

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const DefaultMaxBodyBytes int64 = 1 << 20

type BindingError struct {
	statusCode int
	errorCode  string
	message    string
}

func (e *BindingError) Error() string {
	return e.message
}

func (e *BindingError) StatusCode() int {
	return e.statusCode
}

func (e *BindingError) ErrorCode() string {
	return e.errorCode
}

func (e *BindingError) Message() string {
	return e.message
}

func Bind(c *gin.Context, dst any, maxBodyBytes int64) error {
	if maxBodyBytes <= 0 {
		maxBodyBytes = DefaultMaxBodyBytes
	}

	contentType := strings.TrimSpace(c.GetHeader("Content-Type"))
	if contentType == "" {
		return &BindingError{
			statusCode: http.StatusUnsupportedMediaType,
			errorCode:  "unsupported_media_type",
			message:    "Content-Type must be application/json",
		}
	}

	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil || mediaType != "application/json" {
		return &BindingError{
			statusCode: http.StatusUnsupportedMediaType,
			errorCode:  "unsupported_media_type",
			message:    "Content-Type must be application/json",
		}
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBodyBytes)
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		return classifyDecodeError(err)
	}

	if err := decoder.Decode(&struct{}{}); err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}
		return &BindingError{
			statusCode: http.StatusBadRequest,
			errorCode:  "invalid_request",
			message:    "request body must contain a single JSON object",
		}
	}

	return &BindingError{
		statusCode: http.StatusBadRequest,
		errorCode:  "invalid_request",
		message:    "request body must contain a single JSON object",
	}
}

func WriteError(c *gin.Context, err error) bool {
	bindingErr := &BindingError{}
	if !errors.As(err, &bindingErr) {
		return false
	}

	c.JSON(bindingErr.StatusCode(), gin.H{
		"error":   bindingErr.ErrorCode(),
		"message": bindingErr.Message(),
	})
	return true
}

func classifyDecodeError(err error) error {
	var syntaxErr *json.SyntaxError
	var typeErr *json.UnmarshalTypeError
	var maxBytesErr *http.MaxBytesError

	switch {
	case errors.Is(err, io.EOF):
		return &BindingError{
			statusCode: http.StatusBadRequest,
			errorCode:  "invalid_request",
			message:    "request body is required",
		}
	case errors.As(err, &maxBytesErr):
		return &BindingError{
			statusCode: http.StatusRequestEntityTooLarge,
			errorCode:  "payload_too_large",
			message:    "request body exceeds the maximum allowed size",
		}
	case errors.As(err, &syntaxErr), errors.Is(err, io.ErrUnexpectedEOF):
		return &BindingError{
			statusCode: http.StatusBadRequest,
			errorCode:  "invalid_request",
			message:    "request body must be valid JSON",
		}
	case errors.As(err, &typeErr):
		return &BindingError{
			statusCode: http.StatusBadRequest,
			errorCode:  "invalid_request",
			message:    "request body contains an invalid value",
		}
	case strings.HasPrefix(err.Error(), "json: unknown field "):
		return &BindingError{
			statusCode: http.StatusBadRequest,
			errorCode:  "invalid_request",
			message:    "request body contains unknown fields",
		}
	default:
		return &BindingError{
			statusCode: http.StatusBadRequest,
			errorCode:  "invalid_request",
			message:    "invalid request body",
		}
	}
}
