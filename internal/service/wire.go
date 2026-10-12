package service

import (
	"encoding/json"
	"errors"
	"log/slog"

	"genetica/internal/apperr"
	"genetica/internal/domain"
)

// WireError is the JSON document the frontend receives as the error string
// of a rejected Wails call: {"code","message","fields"}.
type WireError struct {
	Code    apperr.Code       `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

// Wire converts any error into an error whose text is a WireError JSON.
// Internal details are logged, never sent to the UI.
func Wire(log *slog.Logger, err error) error {
	if err == nil {
		return nil
	}
	w := WireError{Code: apperr.CodeInternal, Message: "Ocorreu um erro inesperado."}
	var ve *domain.ValidationError
	var ae *apperr.Error
	switch {
	case errors.As(err, &ve):
		w = WireError{Code: apperr.CodeValidation, Message: "Verifique os campos destacados.", Fields: ve.Fields}
	case errors.As(err, &ae):
		w = WireError{Code: ae.Code, Message: ae.Message}
		if ae.Code != apperr.CodeNotFound && ae.Code != apperr.CodeConflict && log != nil {
			log.Error("service error", "err", err)
		}
	default:
		if log != nil {
			log.Error("unexpected error", "err", err)
		}
	}
	b, _ := json.Marshal(w)
	return errors.New(string(b))
}
