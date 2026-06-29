package asrc

import (
	"sync"
)

// StreamTask represents a single audio stream to be resampled.
type StreamTask struct {
	ID       string
	Input    []float64
	Ratio    float64
	Callback func(output []float64, err error)
}

// Orchestrator manages a pool of ASRC workers for handling multiple concurrent streams.
// This is designed for cloud-native microservices where many short clips or streams are processed.
type Orchestrator struct {
	quality  ASRCQuality
	channels int
	workers  int
	tasks    chan StreamTask
	wg       sync.WaitGroup
	quit     chan struct{}
}

// NewOrchestrator creates a worker pool of CGO-backed resamplers.
func NewOrchestrator(workers int, quality ASRCQuality, channels int) *Orchestrator {
	o := &Orchestrator{
		quality:  quality,
		channels: channels,
		workers:  workers,
		tasks:    make(chan StreamTask, 100),
		quit:     make(chan struct{}),
	}
	o.start()
	return o
}

func (o *Orchestrator) start() {
	for i := 0; i < o.workers; i++ {
		o.wg.Add(1)
		go o.worker()
	}
}

func (o *Orchestrator) worker() {
	defer o.wg.Done()
	
	// Each worker gets its own resampler to avoid lock contention
	resampler := NewASRCResampler(o.quality, o.channels)
	defer resampler.Close()

	for {
		select {
		case task, ok := <-o.tasks:
			if !ok {
				return
			}
			
			resampler.Reset()
			resampler.SetRatio(task.Ratio)
			out := resampler.Process(task.Input)
			
			if task.Callback != nil {
				task.Callback(out, nil)
			}
		case <-o.quit:
			return
		}
	}
}

// Submit enqueues a resampling task to the worker pool.
func (o *Orchestrator) Submit(task StreamTask) {
	o.tasks <- task
}

// Shutdown gracefully stops the worker pool.
func (o *Orchestrator) Shutdown() {
	close(o.tasks)
	o.wg.Wait()
}
