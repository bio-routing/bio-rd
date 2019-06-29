package risserver

import (
	"sync"

	pb "github.com/bio-routing/bio-rd/cmd/ris/api"
)

type updateFIFO struct {
	dataArrived chan struct{}
	data        []*pb.RIBUpdate
	mu          sync.Mutex
}

func newUpdateFIFO() *updateFIFO {
	return &updateFIFO{
		dataArrived: make(chan struct{}, 1),
		data:        make([]*pb.RIBUpdate, 0),
	}
}

func (uf *updateFIFO) queue(r *pb.RIBUpdate) {
	uf.mu.Lock()
	uf.data = append(uf.data, r)
	select {
	case uf.dataArrived <- struct{}{}:
	default:
	}

	uf.mu.Unlock()
}

func (uf *updateFIFO) empty() bool {
	return len(uf.data) == 0
}

// dequeue get's all elements in the queue unless the queue is empty. Then it blocks until it is not empty.
func (uf *updateFIFO) dequeue() []*pb.RIBUpdate {
	<-uf.dataArrived

	uf.mu.Lock()
	data := uf.data
	uf.data = make([]*pb.RIBUpdate, 0, 128)

	uf.mu.Unlock()
	return data
}
