package types

import (
	"strconv"
	"strings"
)

type ClusterList []uint32

func (cl *ClusterList) String() string {
	if cl == nil {
		return ""
	}

	clStrings := make([]string, len(*cl))
	for i, x := range *cl {
		clStrings[i] = strconv.Itoa(int(x))
	}

	return strings.Join(clStrings, " ")
}
