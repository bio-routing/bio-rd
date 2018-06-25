package filter

/*func TestNewAcceptAllFilter(t *testing.T) {
	f := NewAcceptAllFilter()

	m := &clientMock{}
	f.Register(m)

	f.AddPath(net.NewPfx(0, 0), &route.Path{})

	if !m.addPathCalled {
		t.Fatalf("expected accepted, but was filtered")
	}
}

func TestNewDrainFilter(t *testing.T) {
	f := NewDrainFilter()

	m := &clientMock{}
	f.Register(m)

	f.AddPath(net.NewPfx(0, 0), &route.Path{})

	if m.addPathCalled {
		t.Fatalf("expected filtered, but was accepted")
	}
}*/
