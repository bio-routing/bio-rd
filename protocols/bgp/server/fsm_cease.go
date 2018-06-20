package server

type ceaseState struct {
}

func newCeaseState() *ceaseState {
	return &ceaseState{}
}

func (s ceaseState) run() (state, string) {
	return newCeaseState(), "Loop"
}
