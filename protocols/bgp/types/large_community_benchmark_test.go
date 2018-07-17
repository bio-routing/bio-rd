package types

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
)

func BenchmarkParseLargeCommunityString(b *testing.B) {
	for _, i := range []int{1, 2, 4, 8, 16, 32} {
		str := getNNumbers(i)
		input := strings.Join([]string{str, str, str}, ",")
		b.Run(fmt.Sprintf("BenchmarkParseLargeCommunityString-%d", i), func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				ParseLargeCommunityString(input)
			}
		})
	}
}

func getNNumbers(n int) (ret string) {
	var numbers string
	for i := 0; i < n; i++ {
		numbers += strconv.Itoa(i % 10)
	}
	return numbers
}
