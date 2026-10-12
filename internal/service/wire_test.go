package service

import (
	"encoding/json"
	"errors"
	"testing"

	"genetica/internal/apperr"
	"genetica/internal/domain"
)

func decode(t *testing.T, err error) WireError {
	t.Helper()
	var w WireError
	if e := json.Unmarshal([]byte(err.Error()), &w); e != nil {
		t.Fatalf("not JSON: %v", err)
	}
	return w
}

func TestWire(t *testing.T) {
	if Wire(nil, nil) != nil {
		t.Fatal("nil must stay nil")
	}
	w := decode(t, Wire(nil, &domain.ValidationError{Fields: map[string]string{"name": "x"}}))
	if w.Code != apperr.CodeValidation || w.Fields["name"] != "x" {
		t.Fatalf("%+v", w)
	}
	w = decode(t, Wire(nil, apperr.New(apperr.CodeNotFound, "nada")))
	if w.Code != apperr.CodeNotFound || w.Message != "nada" {
		t.Fatalf("%+v", w)
	}
	w = decode(t, Wire(nil, errors.New("secret SQL detail")))
	if w.Code != apperr.CodeInternal || w.Message == "secret SQL detail" {
		t.Fatalf("internal details must not leak: %+v", w)
	}
}
