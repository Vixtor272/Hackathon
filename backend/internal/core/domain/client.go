package domain

import "strings"

// Client is a person buying through Farmi, identified by their cédula.
type Client struct {
	ID    string // cédula
	Name  string
	Phone string
}

// NormalizeCedula keeps only the digits of a user-typed identifier.
func NormalizeCedula(raw string) string {
	var b strings.Builder
	for _, r := range raw {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// ValidCedula applies the demo rule: an Ecuadorian cédula has 10 digits.
func ValidCedula(id string) bool {
	if len(id) != 10 {
		return false
	}
	for _, r := range id {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// NormalizePhone strips spaces and guarantees the international "+" prefix.
func NormalizePhone(raw string) string {
	p := strings.ReplaceAll(strings.TrimSpace(raw), " ", "")
	if p != "" && p[0] != '+' {
		p = "+" + p
	}
	return p
}
