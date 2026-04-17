package feed 

type EventQueue struct{
	ch chan *PostCreatedEvent
}

func NewEventQueue(bufferSize int) *EventQueue{
	return &EventQueue{
		ch: make(chan *PostCreatedEvent, bufferSize),
	}
}

func (q *EventQueue) Publish(event *PostCreatedEvent){
	q.ch <- event
}

func (q *EventQueue) Subscribe() <- chan *PostCreatedEvent {
	return q.ch
}