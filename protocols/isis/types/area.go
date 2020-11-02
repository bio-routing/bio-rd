package types

// AreaID is an ISIS Area ID
type AreaID []byte

// Equal checks if area IDs are equal
func (a AreaID) Equal(b AreaID) bool {
	if len(a) != len(b) {
		return false
	}

	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}
