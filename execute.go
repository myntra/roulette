package roulette

import "time"

// SimpleExecute interface provides methods to retrieve a parser and a method which executes on the incoming values.
type SimpleExecute interface {
	Execute(vals ...interface{})
}

// QueueExecute interface provides methods to retrieve a parser and a method which executes on the incoming values on the input channel.
type QueueExecute interface {
	Execute(in <-chan interface{}, out chan<- interface{}) // in channel to write, out channel to read.
	CloseResult()
}

// SimpleExecutor implements the SimpleExecute interface
type SimpleExecutor struct {
	Parser Parser
}

// Execute executes rules in order of priority.
// one(true): executes in order of priority until a high priority rule is successful, after which execution stops
func (s *SimpleExecutor) Execute(vals ...interface{}) {
	s.Parser.Execute(vals)
}

// QueueExecutor implements the QueueExecute
type QueueExecutor struct {
	Parser Parser
}

// Execute ...
func (q *QueueExecutor) Execute(in <-chan interface{}, out chan<- interface{}) {

	go q.drainQueue(out)
	go q.fillQueue(in)

}

func (q *QueueExecutor) processWorker(vals interface{}) {
	q.process(vals)
}

func (q *QueueExecutor) process(vals interface{}) error {

	q.Parser.Execute(vals)
	return nil
}

func (q *QueueExecutor) fillQueue(in <-chan interface{}) {
fill:
	for {
		select {
		case v, ok := <-in:
			if !ok {
				break fill
			}

			go q.processWorker(v)
			//TODO: quit the loop clean
			//TODO: Pool of process workers
		}
	}
}

// adapter from github.com/kylelemons/iq
func (q *QueueExecutor) drainQueue(out chan<- interface{}) {
	defer close(out)

	// pending events (this is the "infinite" part)
	var pending []interface{}

recv:
	for {
		// Ensure that pending always has values so the select can
		// multiplex between the receiver and sender properly
		if len(pending) == 0 {
			v, ok := <-q.Parser.GetResult().Get().(chan interface{})
			if !ok {
				// in is closed, flush values
				break recv
			}

			switch v.(type) {
			case empty:
				continue
			}

			pending = append(pending, v)

		}

		select {
		// Queue incoming values
		case v, ok := <-q.Parser.GetResult().Get().(chan interface{}):
			if !ok {
				// in is closed, flush values
				break recv
			}

			switch v.(type) {
			case empty:
				continue
			}

			pending = append(pending, v)

		// Send queued values
		case out <- pending[0]:
			pending = pending[1:]
		}
	}

	// After in is closed, we may still have events to send
	for _, v := range pending {
		out <- v
	}

}

// CloseResult closes the result channel
func (q *QueueExecutor) CloseResult() {
	q.Parser.GetResult().Get().(chan interface{}) <- quit{}
	select {
	case <-q.Parser.GetResult().Get().(chan interface{}):
	case <-time.After(time.Millisecond * 10):
		close(q.Parser.GetResult().Get().(chan interface{}))
		break

	}

}	

// NewSimpleExecutor returns a new SimpleExecutor
func NewSimpleExecutor(parser Parser) SimpleExecute {
	return &SimpleExecutor{Parser: parser}
}

// NewQueueExecutor returns a new QueueExecutor
func NewQueueExecutor(parser Parser) QueueExecute {
	return &QueueExecutor{Parser: parser}
}
