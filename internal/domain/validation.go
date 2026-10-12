// Package domain contains the core entities and their validation rules.
// It has no dependency on SQL, Wails or the UI.
package domain

import (
	"encoding/json"
	"math"
	"strings"
	"time"
)

// ValidationError lists field-level problems, suitable for showing in a form.
type ValidationError struct {
	Fields map[string]string `json:"fields"`
}

func (v *ValidationError) Error() string {
	parts := make([]string, 0, len(v.Fields))
	for f, m := range v.Fields {
		parts = append(parts, f+": "+m)
	}
	return "validação: " + strings.Join(parts, "; ")
}

// validator accumulates field errors.
type validator struct{ fields map[string]string }

func (v *validator) add(field, msg string) {
	if v.fields == nil {
		v.fields = map[string]string{}
	}
	if _, exists := v.fields[field]; !exists {
		v.fields[field] = msg
	}
}

func (v *validator) err() error {
	if len(v.fields) == 0 {
		return nil
	}
	return &ValidationError{Fields: v.fields}
}

const DateLayout = "2006-01-02"

// Now is the clock used by the domain; tests may replace it.
var Now = func() time.Time { return time.Now().UTC() }

// Timestamp formats a time as RFC 3339 UTC.
func Timestamp(t time.Time) string { return t.UTC().Format(time.RFC3339) }

func finite(f float64) bool { return !math.IsNaN(f) && !math.IsInf(f, 0) }

func oneOf(s string, set ...string) bool {
	for _, x := range set {
		if s == x {
			return true
		}
	}
	return false
}

// checkDate validates a YYYY-MM-DD date; futureOK allows dates after today.
func checkDate(v *validator, field, s string, futureOK bool) {
	t, err := time.Parse(DateLayout, s)
	if err != nil {
		v.add(field, "Data inválida (use AAAA-MM-DD).")
		return
	}
	// A calendar date has no zone: allow one day ahead of UTC so that a date that is "today"
	// for the user (e.g. UTC+9 in the morning) is never rejected as future.
	if !futureOK && t.After(Now().AddDate(0, 0, 1)) {
		v.add(field, "A data não pode estar no futuro.")
	}
	if t.Year() < 1900 {
		v.add(field, "Data anterior a 1900.")
	}
}

func required(v *validator, field, s, msg string) {
	if strings.TrimSpace(s) == "" {
		v.add(field, msg)
	}
}

func jsonValid(s string) bool { return json.Valid([]byte(s)) }

func jsonObject(s string) bool {
	var m map[string]any
	return json.Unmarshal([]byte(s), &m) == nil && m != nil
}
