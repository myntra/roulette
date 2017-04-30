package roulette

// Result ...
type Result interface {
	Put(val interface{}, prevVal bool) bool
	Get() interface{}
}

// ResultCallback holds the out channel
type ResultCallback struct {
	fn func(result interface{})
}

// Put receives a value and put's it on the parser's out channel
func (q *ResultCallback) Put(val interface{}, prevVal bool) bool {
	if !prevVal {
		return false
	}

	q.fn(val)
	return true
}

// Get ...
func (q *ResultCallback) Get() interface{} {
	return q.fn
}

// NewResultCallback ...
func NewResultCallback(fn func(interface{})) *ResultCallback {
	return &ResultCallback{fn: fn}
}

// ResultQueue holds the out channel
type ResultQueue struct {
	get chan interface{}
}

// Put receives a value and put's it on the parser's out channel
func (q *ResultQueue) Put(val interface{}, prevVal bool) bool {
	//fmt.Println(val, prevVal)
	if !prevVal {
		return false
	}
	go func(val interface{}) {
		q.get <- val
	}(val)

	return true
}

type empty struct{}

// Get ...
func (q *ResultQueue) Get() interface{} {
	return q.get
}

func (q *ResultQueue) block() {
	// nil channel blocks forever
	//	blocks := make(chan interface{})
	for {
		select {
		case q.get <- empty{}:

		}
	}
}

// NewResultQueue ...
func NewResultQueue() *ResultQueue {
	q := &ResultQueue{get: make(chan interface{})}
	go q.block()
	return q
}
