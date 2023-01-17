package ethernet

type EthernetInterfaceFactoryI interface {
	New(name string, bpf *BPF, llc LLC) (EthernetInterfaceI, error)
}

type EthernetInterfaceFactory struct{}

func NewEthernetInterfaceFactory() *EthernetInterfaceFactory {
	return &EthernetInterfaceFactory{}
}

func (eif *EthernetInterfaceFactory) New(name string, bpf *BPF, llc LLC) (EthernetInterfaceI, error) {
	return NewEthernetInterface(name, bpf, llc)
}

type MockEthernetInterfaceFactory struct{}

func NewMockEthernetInterfaceFactory() *MockEthernetInterfaceFactory {
	return &MockEthernetInterfaceFactory{}
}

func (meif *MockEthernetInterfaceFactory) New(name string, bpf *BPF, llc LLC) (EthernetInterfaceI, error) {
	return NewMockEthernetInterface(), nil
}
