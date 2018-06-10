package packet

import (
	"errors"
	"testing"

	"fmt"
	"math"

	"strconv"

	"github.com/stretchr/testify/assert"
)

func TestParseLargeCommunityString(t *testing.T) {
	tests := []struct {
		name     string
		in       string
		expected LargeCommunity
		err      error
	}{
		{
			name: "normal large community",
			in:   "(1,2,3)",
			expected: LargeCommunity{
				GlobalAdministrator: 1,
				DataPart1:           2,
				DataPart2:           3,
			},
			err: nil,
		},
		{
			name:     "too short community",
			in:       "(1,2)",
			expected: LargeCommunity{},
			err:      errors.New("malformed large community string (1,2)"),
		},
		{
			name: "missing parentheses large community",
			in:   "1,2,3",
			expected: LargeCommunity{
				GlobalAdministrator: 1,
				DataPart1:           2,
				DataPart2:           3,
			},
			err: nil,
		},
		{
			name:     "malformed large community",
			in:       "[1,2,3]",
			expected: LargeCommunity{},
			err:      errors.New("malformed large community string [1,2,3]"),
		},
		{
			name:     "missing digit",
			in:       "(,2,3)",
			expected: LargeCommunity{},
			err:      errors.New("malformed large community string (,2,3)"),
		},
		{
			name:     "to big global administrator",
			in:       fmt.Sprintf("(%d,1,2)", math.MaxInt64),
			expected: LargeCommunity{},
			err:      &strconv.NumError{Func: "ParseUint", Num: fmt.Sprintf("%d", math.MaxInt64), Err: strconv.ErrRange},
		},
		{
			name:     "to big data part 1",
			in:       fmt.Sprintf("(1,%d,2)", math.MaxInt64),
			expected: LargeCommunity{1, 0, 0},
			err:      &strconv.NumError{Func: "ParseUint", Num: fmt.Sprintf("%d", math.MaxInt64), Err: strconv.ErrRange},
		},
		{
			name:     "to big data part 2",
			in:       fmt.Sprintf("(1,2,%d)", math.MaxInt64),
			expected: LargeCommunity{1, 2, 0},
			err:      &strconv.NumError{Func: "ParseUint", Num: fmt.Sprintf("%d", math.MaxInt64), Err: strconv.ErrRange},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			com, err := ParseLargeCommunityString(test.in)
			assert.Equal(t, test.err, err)
			assert.Equal(t, test.expected, com)
		})
	}
}
