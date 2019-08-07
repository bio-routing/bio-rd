package types

import "strconv"

type ClusterList []uint32

func (cl *ClusterList) String() string {
	if cl == nil {
		return ""
	}

	ret := ""
	for _, x := range *cl {
		ret += strconv.Itoa(int(x)) + " "
	}

	return ret
}
