// Package domain holds the business entities and rules of Farmi. It has no
// dependencies on frameworks, storage or transport: adapters depend on it,
// never the other way around.
package domain

import "fmt"

// Money is an amount in US cents. Integer cents keep totals exact.
type Money int64

// Cents builds a Money value from an integer amount of cents.
func Cents(c int64) Money { return Money(c) }

// Times multiplies the amount by a quantity.
func (m Money) Times(qty int) Money { return m * Money(qty) }

// Float renders the amount in dollars for JSON payloads (12.5).
func (m Money) Float() float64 { return float64(m) / 100 }

// Format renders the amount the way Farmi writes it in chat: "USD 12,50".
func (m Money) Format() string {
	sign := ""
	if m < 0 {
		sign = "-"
		m = -m
	}
	return fmt.Sprintf("%sUSD %d,%02d", sign, m/100, m%100)
}
