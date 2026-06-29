package asrc

import (
	"fmt"
	"golang.org/x/sys/unix"
)

// DMAStreamer represents a kernel-bypass streaming mechanism
// using memory-mapped I/O (mmap) for ultra-low latency direct memory access.
type DMAStreamer struct {
	fd     int
	memory []byte
	size   int
}

// NewDMAStreamer attempts to open a file or device descriptor and memory-map it.
// This is typically used with specialized driver nodes (e.g. /dev/uio0) or raw hugepages.
func NewDMAStreamer(filepath string, size int) (*DMAStreamer, error) {
	fd, err := unix.Open(filepath, unix.O_RDWR|unix.O_SYNC, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open DMA target: %w", err)
	}

	mem, err := unix.Mmap(fd, 0, size, unix.PROT_READ|unix.PROT_WRITE, unix.MAP_SHARED)
	if err != nil {
		unix.Close(fd)
		return nil, fmt.Errorf("failed to mmap DMA target: %w", err)
	}

	return &DMAStreamer{
		fd:     fd,
		memory: mem,
		size:   size,
	}, nil
}

// Close unmaps the memory and closes the file descriptor.
func (d *DMAStreamer) Close() error {
	if d.memory != nil {
		_ = unix.Munmap(d.memory)
	}
	if d.fd > 0 {
		_ = unix.Close(d.fd)
	}
	return nil
}
