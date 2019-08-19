package types

import (
	"fmt"
	"strings"
)

type ClusterList []uint32

func (cl *ClusterList) String() string {
	if cl == nil {
		return ""
	}

	var b strings.Builder
	for _, x := range *cl {
		fmt.Fprintf(&b, "%d ", x)
	}

	return b.String()
}
