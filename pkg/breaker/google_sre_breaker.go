package breaker

import (
	"math/rand"
	"sync"
	"time"

	"github.com/UnderTreeTech/waterdrop/pkg/utils/xcollection"
)

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

type googleSreBreaker struct {
	// google accepts multiplier K
	K     float64
	state int32
	rw    *xcollection.RollingWindow
	proba *Proba
}

func NewGoogleSreBreaker() {

}
