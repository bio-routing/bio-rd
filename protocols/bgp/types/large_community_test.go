package types

import (
	"errors"
	"testing"

	"fmt"
	"math"

	"strconv"

	"github.com/bio-routing/bio-rd/route/api"
	"github.com/stretchr/testify/assert"
)

func TestLargeCommunityFromProtoCommunity(t *testing.T) {
	input := &api.LargeCommunity{
		GlobalAdministrator: 1,
		DataPart1:           100,
		DataPart2:           200,
	}

	expected := LargeCommunity{
		GlobalAdministrator: 1,
		DataPart1:           100,
		DataPart2:           200,
	}

	result := LargeCommunityFromProtoCommunity(input)
	assert.Equal(t, expected, result)
}

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
			err:      errors.New("can not parse large community 1,2"),
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
			err:      &strconv.NumError{Func: "ParseUint", Num: "[1", Err: strconv.ErrSyntax},
		},
		{
			name:     "missing digit",
			in:       "(,2,3)",
			expected: LargeCommunity{},
			err:      &strconv.NumError{Func: "ParseUint", Num: "", Err: strconv.ErrSyntax},
		},
		{
			name:     "too big global administrator",
			in:       fmt.Sprintf("(%d,1,2)", math.MaxInt64),
			expected: LargeCommunity{},
			err:      &strconv.NumError{Func: "ParseUint", Num: fmt.Sprintf("%d", math.MaxInt64), Err: strconv.ErrRange},
		},
		{
			name:     "too big data part 1",
			in:       fmt.Sprintf("(1,%d,2)", math.MaxInt64),
			expected: LargeCommunity{1, 0, 0},
			err:      &strconv.NumError{Func: "ParseUint", Num: fmt.Sprintf("%d", math.MaxInt64), Err: strconv.ErrRange},
		},
		{
			name:     "too big data part 2",
			in:       fmt.Sprintf("(1,2,%d)", math.MaxInt64),
			expected: LargeCommunity{1, 2, 0},
			err:      &strconv.NumError{Func: "ParseUint", Num: fmt.Sprintf("%d", math.MaxInt64), Err: strconv.ErrRange},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			com, err := ParseLargeCommunityString(test.in)
			if test.err != nil {
				assert.EqualError(t, err, test.err.Error())
			} else {
				assert.Nil(t, err)
			}

			assert.Equal(t, test.expected, com)
		})
	}
}
