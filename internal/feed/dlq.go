package feed

type DeadLetterQueue struct {
	ch chan *PostCreatedEvent
}

func NewDeadLetterQueue(bufferSize int) *DeadLetterQueue {
	return &DeadLetterQueue{
		ch: make(chan *PostCreatedEvent, bufferSize),
	}
}

func (d *DeadLetterQueue) Push(event *PostCreatedEvent){
	d.ch <- event
}

func (d *DeadLetterQueue) Subscribe() <-chan *PostCreatedEvent{
	return d.ch
}