package routingtable

import "fmt"

// PrefixLimitError represents an error when a prefix limit (on a Adj-RIB was hit)
type PrefixLimitError struct {
	limit uint
}

// NewPrefixLimitError creates a new error
func NewPrefixLimitError(limit uint) *PrefixLimitError {
	return &PrefixLimitError{limit}
}

func (err *PrefixLimitError) Error() string {
	return fmt.Sprintf("Prefix if %d limit was hit", err.limit)
}

// Limit gives a notion about the threashold hit
func (err *PrefixLimitError) Limit() uint {
	return err.limit
}
