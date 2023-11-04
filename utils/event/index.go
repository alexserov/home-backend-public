package event

import (
	"github.com/google/uuid"
)

type eventhandler[TSender, TArgs any] func(sender TSender, args TArgs)
type eventItem[TSender, TArgs any] struct {
	key string
	eventhandler[TSender, TArgs]
}

type event[TSender, TArgs any] struct {
	sender TSender
	items []eventItem[TSender, TArgs]
}

type EventPublic[TSender, TArgs any] interface {
	Add(handler eventhandler[TSender, TArgs]) string
	Remove(key string)
}
type EventPrivate[TSender, TArgs any] interface {
	EventPublic[TSender, TArgs]
	RaiseEvent(args TArgs)
}

func Create[TSender, TArgs any](sender TSender) EventPrivate[TSender, TArgs] {
	return &event[TSender, TArgs]{
		sender: sender, 
		items: make([]eventItem[TSender, TArgs], 0),
	}
}

func (event *event[TSender, TArgs]) Add(handler eventhandler[TSender, TArgs]) string {
	uid := uuid.New().String()
	event.items = append(event.items, eventItem[TSender, TArgs]{uid, handler})

	return uid
}

func (event *event[TSender, TArgs]) Remove(key string) {
	for index, item := range event.items {
		if item.key == key {
			event.items = append(event.items[:index], event.items[index+1:]...)
		}
	}
}

func (event *event[TSender, TArgs]) RaiseEvent(args TArgs) {
	for _, item := range event.items {
		item.eventhandler(event.sender, args)
	}
}