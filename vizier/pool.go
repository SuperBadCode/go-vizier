package vizier

import (
	"errors"
	"fmt"
	"math"
	"sync"
)

var (
	ErrStatesEmpty       = errors.New("there must be at least one state")
	ErrPoolNotRunning    = errors.New("the pool is not running")
	ErrInvalidPoolSize   = errors.New("the pool size must be greater than 0")
	ErrInvalidRetryValue = errors.New("the pool retries value cannot be negative")
)

type Pool struct {
	_          struct{}
	wg         sync.WaitGroup
	states     map[string]IState
	size       int
	run        bool
	stopWorker chan bool
}

func (p *Pool) Create() error {
	for i := 0; i < p.size; i++ {
		p.spawnWorker()
	}

	return nil
}

func (p *Pool) spawnWorker() {
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		defer func() {
			if err := recover(); err != nil {
				fmt.Println(err)
				p.spawnWorker()
			}
		}()
		for p.run {
			select {
			case <-p.stopWorker:
				return
			default:
				for _, state := range p.states {
					state.Run()
				}
			}
		}
	}()
}

func (p *Pool) Stop() error {
	if !p.run {
		return ErrPoolNotRunning
	}
	p.run = false
	return nil
}

func (p *Pool) Wait() error {
	if !p.run {
		return ErrPoolNotRunning
	}
	p.wg.Wait()
	return nil
}

func (p *Pool) SetSize(size int) error {
	if !p.run {
		return ErrPoolNotRunning
	}

	if size <= 0 {
		return ErrInvalidPoolSize
	}

	delta := int(math.Abs(float64(p.size - size)))
	spawn := (size > p.size)
	for i := 0; i < delta; i++ {
		if spawn {
			p.spawnWorker()
			continue
		}
		p.stopWorker <- true
	}

	return nil
}

func NewPool(size int, states map[string]IState) (*Pool, error) {
	if size <= 0 {
		return nil, ErrInvalidPoolSize
	}

	if len(states) <= 0 {
		return nil, ErrStatesEmpty
	}

	pool := Pool{
		states:     states,
		size:       size,
		run:        true,
		stopWorker: make(chan bool),
	}

	return &pool, nil
}