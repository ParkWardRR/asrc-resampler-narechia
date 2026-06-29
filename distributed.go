package asrc

import (
	"context"
	"fmt"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// DistributedOrchestrator acts as a client to a fleet of remote ASRC gRPC workers.
type DistributedOrchestrator struct {
	connections []*grpc.ClientConn
	mu          sync.Mutex
	roundRobin  int
}

// NewDistributedOrchestrator connects to multiple remote worker endpoints.
func NewDistributedOrchestrator(endpoints []string) (*DistributedOrchestrator, error) {
	var conns []*grpc.ClientConn
	for _, ep := range endpoints {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		conn, err := grpc.DialContext(ctx, ep, grpc.WithTransportCredentials(insecure.NewCredentials()))
		cancel()
		if err != nil {
			return nil, fmt.Errorf("failed to connect to %s: %w", ep, err)
		}
		conns = append(conns, conn)
	}

	return &DistributedOrchestrator{
		connections: conns,
	}, nil
}

// Submit distributed task (stub implementation).
func (d *DistributedOrchestrator) Submit(task StreamTask) error {
	d.mu.Lock()
	if len(d.connections) == 0 {
		d.mu.Unlock()
		return fmt.Errorf("no workers available")
	}
	// Select connection via round-robin
	// conn := d.connections[d.roundRobin%len(d.connections)]
	d.roundRobin++
	d.mu.Unlock()

	// In a real implementation, we would use the generated protobuf client here to stream
	// the `task.Input` to the `conn`. Since we haven't compiled the proto to Go code yet
	// to avoid protoc dependency issues on the user's machine, we leave this as a stub.

	return nil
}

// Close tears down gRPC connections.
func (d *DistributedOrchestrator) Close() {
	d.mu.Lock()
	defer d.mu.Unlock()
	for _, c := range d.connections {
		_ = c.Close()
	}
}
