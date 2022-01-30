package trie

type Trie struct {
}

type Route interface {
	SameKey() bool
}

func (t *Trie) AddRoute(r Route) {

}
