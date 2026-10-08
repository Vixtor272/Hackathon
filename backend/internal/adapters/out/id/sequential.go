// Package id issues readable sequential identifiers ("ord_1", "pay_2").
package id

import (
	"fmt"
	"sync"
)

// Sequential implements ports.IDGenerator.
type Sequential struct {
	mu       sync.Mutex
	counters map[string]int
}

// NewSequential starts every prefix at 1.
func NewSequential() *Sequential { return &Sequential{counters: map[string]int{}} }

// New returns the next id for a prefix.
func (s *Sequential) New(prefix string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.counters[prefix]++
	return fmt.Sprintf("%s_%d", prefix, s.counters[prefix])
}
