package server

type Router interface {
	IsDR() bool
	IsBDR() bool
}

// RFC: https://tools.ietf.org/html/rfc2328#page-95 (Section 10.4)
func shouldFormAdjacency(ifType InterfaceType, self Router, neighbor Router) bool {
	return ifType == IfTypePointToPoint ||
		ifType == IfTypePointToMultipoint ||
		ifType == IfTypeVirtualLink ||
		self.IsDR() || self.IsBDR() ||
		neighbor.IsDR() || neighbor.IsBDR()
}
