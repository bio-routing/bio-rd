package routingtable

import "fmt"

// PrefixLimitHitError represents an error when a prefix limit (on a Adj-RIB was hit)
type PrefixLimitHitError struct {
	limit uint
}

// NewPrefixLimitHitError creates a new error
func NewPrefixLimitHitError(limit uint) *PrefixLimitHitError {
	return &PrefixLimitHitError{limit}
}

func (err *PrefixLimitHitError) Error() string {
	return fmt.Sprintf("Prefix if %d limit was hit", err.limit)
}

// Limit gives a notion about the threashold hit
func (err *PrefixLimitHitError) Limit() uint {
	return err.limit
}
