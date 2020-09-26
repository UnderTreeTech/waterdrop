package breaker

type Breaker interface {
	Allow() error
	Accept()
	Reject()
}
