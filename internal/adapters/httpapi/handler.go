package httpapi

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	app "github.com/inceptionlabscorp/sms-otp-factor-service/internal/application/smsotp"
	domain "github.com/inceptionlabscorp/sms-otp-factor-service/internal/domain/smsotp"
)

const maxJSONBodyBytes = 4096

type Handler struct {
	OTP          app.Service
	Session      app.SessionService
	ServiceToken string
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/health" && r.Method == http.MethodGet {
		writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "service": "sms-otp-factor-service"})
		return
	}
	if !h.authorized(r) {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "unauthorized"})
		return
	}
	switch {
	case r.Method == http.MethodPost && (r.URL.Path == "/v1/sms-otp/send" || r.URL.Path == "/v1/admin/sms-otp/send" || r.URL.Path == "/v1/bian/customer-access-entitlement/sms-otp/initiate"):
		h.send(w, r)
	case r.Method == http.MethodPost && (r.URL.Path == "/v1/sms-otp/verify" || r.URL.Path == "/v1/admin/sms-otp/verify" || r.URL.Path == "/v1/bian/customer-access-entitlement/sms-otp/execute"):
		h.verify(w, r)
	case r.Method == http.MethodPost && (r.URL.Path == "/v1/sms-mfa/session/validate" || r.URL.Path == "/v1/admin/sms-mfa/session/validate" || r.URL.Path == "/v1/bian/customer-access-entitlement/sms-mfa-session/evaluate"):
		h.validateSession(w, r)
	default:
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "not_found"})
	}
}

type sendRequest struct {
	SubjectID   string `json:"subject_id"`
	PhoneNumber string `json:"phone_number"`
	Purpose     string `json:"purpose"`
}

type verifyRequest struct {
	SubjectID   string `json:"subject_id"`
	PhoneNumber string `json:"phone_number"`
	Purpose     string `json:"purpose"`
	Code        string `json:"code"`
}

type validateSessionRequest struct {
	SubjectID string `json:"subject_id"`
	Token     string `json:"token"`
}

func (h Handler) send(w http.ResponseWriter, r *http.Request) {
	var input sendRequest
	if !decodeJSON(w, r, &input) {
		return
	}
	err := h.OTP.Send(r.Context(), app.SendInput{
		SubjectID:   input.SubjectID,
		PhoneNumber: input.PhoneNumber,
		Purpose:     input.Purpose,
	})
	if err != nil {
		writeOTPError(w, err)
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]any{"status": "sent"})
}

func (h Handler) verify(w http.ResponseWriter, r *http.Request) {
	var input verifyRequest
	if !decodeJSON(w, r, &input) {
		return
	}
	err := h.OTP.Verify(r.Context(), app.VerifyInput{
		SubjectID:   input.SubjectID,
		PhoneNumber: input.PhoneNumber,
		Purpose:     input.Purpose,
		Code:        input.Code,
	})
	if err != nil {
		writeOTPError(w, err)
		return
	}
	token, expiresIn, err := h.Session.Sign(input.SubjectID)
	if err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "sms mfa session signing is not configured"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"mfa_token":  token,
		"expires_in": expiresIn,
		"method":     domain.SessionMethod,
	})
}

func (h Handler) validateSession(w http.ResponseWriter, r *http.Request) {
	var input validateSessionRequest
	if !decodeJSON(w, r, &input) {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"valid": h.Session.Validate(input.Token, input.SubjectID)})
}

func (h Handler) authorized(r *http.Request) bool {
	expected := strings.TrimSpace(h.ServiceToken)
	if expected == "" {
		return false
	}
	header := strings.TrimSpace(r.Header.Get("Authorization"))
	if !strings.HasPrefix(header, "Bearer ") {
		return false
	}
	actual := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
	return subtle.ConstantTimeCompare([]byte(actual), []byte(expected)) == 1
}

func decodeJSON(w http.ResponseWriter, r *http.Request, out any) bool {
	defer r.Body.Close()
	r.Body = http.MaxBytesReader(w, r.Body, maxJSONBodyBytes)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(out); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid json"})
		return false
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid json"})
		return false
	}
	return true
}

func writeOTPError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrNotConfigured):
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "sms otp provider is not configured"})
	case errors.Is(err, domain.ErrInvalidInput):
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid request"})
	case errors.Is(err, domain.ErrRateLimited):
		writeJSON(w, http.StatusTooManyRequests, map[string]any{"error": "too many sms otp requests"})
	case errors.Is(err, domain.ErrExpiredCode):
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "sms otp code expired"})
	case errors.Is(err, domain.ErrInvalidCode):
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid sms otp code"})
	default:
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "sms otp operation failed"})
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func TimeoutMiddleware(next http.Handler, timeout time.Duration) http.Handler {
	return http.TimeoutHandler(next, timeout, `{"error":"timeout"}`)
}
