package breaker

import (
	"math/rand"
	"sync"
	"time"
)

const (
	// StateClosed when circuit breaker closed, request allowed, the breaker
	// calc the succeed ratio, if request num greater request setting and
	// ratio lower than the setting ratio, then reset state to open.
	StateClosed int32 = iota
	// StateOpen when circuit breaker open, request not allowed, after sleep
	// some duration, allow one single request for testing the health, if ok
	// then state reset to closed, if not continue the step.
	StateOpen
)

type Breaker interface {
	Allow() error
	Accept()
	Reject()
	Do(func() error, func(error) bool) error
}

type Proba struct {
	r    *rand.Rand
	lock sync.Mutex
}

func NewProba() *Proba {
	return &Proba{
		r: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (p *Proba) TrueOnProba(proba float64) bool {
	p.lock.Lock()
	reject := p.r.Float64() < proba
	p.lock.Unlock()
	return reject
}

type BreakerGroup struct {
	mutex    sync.RWMutex
	breakers map[string]Breaker
}

func NewBreakerGroup() *BreakerGroup {
	return &BreakerGroup{
		breakers: make(map[string]Breaker),
	}
}

func (bg *BreakerGroup) Get(name string) Breaker {
	bg.mutex.RLock()
	breaker, ok := bg.breakers[name]
	bg.mutex.RUnlock()
	if ok {
		return breaker
	}

	bg.mutex.Lock()
	breaker = newGoogleSreBreaker(nil)
	if _, ok := bg.breakers[name]; !ok {
		bg.breakers[name] = breaker
	}
	bg.mutex.Unlock()

	return breaker
}

func (bg *BreakerGroup) Do(name string, run func() error, accept func(error) bool) error {
	breaker := bg.Get(name)
	return breaker.Do(run, accept)
}
