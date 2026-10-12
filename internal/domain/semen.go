package domain

import "strings"

// Stock movement types.
const (
	MovEntry      = "entry"
	MovWithdrawal = "withdrawal"
	MovAdjustment = "adjustment"
)

// Movement reasons per type.
var (
	EntryReasons      = []string{"purchase", "donation", "transfer", "other"}
	WithdrawalReasons = []string{"insemination", "sale", "loss", "transfer", "other"}
)

const ReasonCountCorrection = "count_correction"

// SemenType is a configurable kind of semen (conventional, sexed, ...).
type SemenType struct {
	Code      string `json:"code"`
	Name      string `json:"name"`
	IsBuiltin bool   `json:"isBuiltin"`
}

// SemenMovement is one recorded change of a lot's stock.
type SemenMovement struct {
	ID           string `json:"id"`
	InventoryID  string `json:"inventoryId"`
	BullID       string `json:"bullId"`
	Type         string `json:"type"`
	Delta        int    `json:"delta"`
	BalanceAfter int    `json:"balanceAfter"`
	Reason       string `json:"reason"`
	Note         string `json:"note"`
	MatingID     string `json:"matingId"`
	IsDemo       bool   `json:"isDemo"`
	CreatedAt    string `json:"createdAt"`
}

// ValidateSemenTypeName checks a user-defined type name.
func ValidateSemenTypeName(name string) error {
	var v validator
	n := strings.TrimSpace(name)
	if n == "" {
		v.add("name", "Informe o nome do tipo de sêmen.")
	} else if len(n) > 60 {
		v.add("name", "Nome muito longo.")
	}
	return v.err()
}

// ValidateMovement checks the quantity and reason for a requested movement.
// qty is the absolute number of doses (for adjustments, the signed delta).
func ValidateMovement(movType string, qty int, reason, note string) error {
	var v validator
	switch movType {
	case MovEntry:
		if qty <= 0 {
			v.add("quantity", "Informe uma quantidade maior que zero.")
		}
		if !oneOf(reason, EntryReasons...) {
			v.add("reason", "Motivo inválido para entrada.")
		}
	case MovWithdrawal:
		if qty <= 0 {
			v.add("quantity", "Informe uma quantidade maior que zero.")
		}
		if !oneOf(reason, WithdrawalReasons...) {
			v.add("reason", "Motivo inválido para retirada.")
		}
	case MovAdjustment:
		if qty == 0 {
			v.add("quantity", "O ajuste não pode ser zero.")
		}
		if strings.TrimSpace(note) == "" {
			v.add("note", "Explique o motivo do ajuste.")
		}
	default:
		v.add("type", "Tipo de movimentação inválido.")
	}
	return v.err()
}
