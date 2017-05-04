package roulette

// Result interface provides methods to get/put an interface{} value from a rule.
type Result interface {
	Put(val interface{}, prevVal ...bool) bool
	Get() interface{}
}

// ResultCallback holds the out channel
type ResultCallback struct {
	fn func(result interface{})
}

// Put receives a value and put's it on the parser's out channel
func (q ResultCallback) Put(val interface{}, prevVal ...bool) bool {
	if len(prevVal) > 0 {
		if !prevVal[0] {
			return false
		}
	}

	q.fn(val)
	return true
}

// Get the callback function.
func (q ResultCallback) Get() interface{} {
	return q.fn
}

// NewResultCallback return a new ResultCallback
func NewResultCallback(fn func(interface{})) *ResultCallback {
	return &ResultCallback{fn: fn}
}

// ResultQueue holds the out channel
type ResultQueue struct {
	get chan interface{}
}

// Put receives a value and put's it on the parser's out channel
func (q ResultQueue) Put(val interface{}, prevVal ...bool) bool {
	//fmt.Println(val, prevVal)
	if len(prevVal) > 0 {
		if !prevVal[0] {
			return false
		}
	}

	go func(val interface{}) {
		q.get <- val
	}(val)

	return true
}

type empty struct{}
type quit struct{}

// Get ...
func (q ResultQueue) Get() interface{} {
	return q.get
}

func (q ResultQueue) block() {
	// nil channel blocks forever
	//	blocks := make(chan interface{})
block:
	for {
		select {
		case v := <-q.get:
			switch v.(type) {
			case quit:
				break block
			default:
				// is not quit, put it back
				q.get <- v
			}

		case q.get <- empty{}:
		}
	}
}

// NewResultQueue returns a new ResultQueue
func NewResultQueue() ResultQueue {
	q := ResultQueue{get: make(chan interface{})}
	go q.block()
	return q
}
