package rt

type Filter struct {
	ClientManager
}

func NewFilter() *Filter {
	return &Filter{}
}

func (f *Filter) Add(r *Route) {
	for client := range f.clients {
		client.AddPath(r)
	}
}

func (f *Filter) Remove(r *Route) {
	for client := range f.clients {
		client.RemovePath(r)
	}
}
