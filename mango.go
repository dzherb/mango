package mango

import "sync/atomic"

var activeEffect *subscriber

func setActiveEffect(sub *subscriber) {
	activeEffect = sub
}

func getActiveEffect() *subscriber {
	return activeEffect
}

var idCounter atomic.Uint64

type subscriberFn func() subscriberResult
type subscriberResult byte

const (
	keep subscriberResult = iota
	unsubscribe
)

type subscriber struct {
	id uint64
	fn subscriberFn
}

func newSubscriber(fn subscriberFn) *subscriber {
	sub := &subscriber{
		id: idCounter.Add(1),
		fn: fn,
	}

	return sub
}

type subscribable struct {
	subs []*subscriber
}

func (s *subscribable) addSubscriber(newSub *subscriber) {
	if newSub == nil {
		return
	}

	for _, sub := range s.subs {
		if sub.id == newSub.id {
			return
		}
	}

	s.subs = append(s.subs, newSub)
}

func (s *subscribable) removeSubscriber(subToRemove *subscriber) {
	for i, sub := range s.subs {
		if sub.id == subToRemove.id {
			s.subs = append(s.subs[:i], s.subs[i+1:]...)
			return
		}
	}
}

func (s *subscribable) callSubscribers() {
	toUnsubscribe := make([]*subscriber, 0, len(s.subs))

	for _, sub := range s.subs {
		if res := sub.fn(); res == unsubscribe {
			toUnsubscribe = append(toUnsubscribe, sub)
		}
	}

	for _, sub := range toUnsubscribe {
		s.removeSubscriber(sub)
	}
}

type Reactive[T comparable] struct {
	subscribable
	value T
}

func NewReactive[T comparable](value T) *Reactive[T] {
	return &Reactive[T]{value: value}
}

func (r *Reactive[T]) Get() T {
	r.addSubscriber(getActiveEffect())
	return r.value
}

func (r *Reactive[T]) Set(value T) {
	if value == r.value {
		return
	}

	r.value = value
	r.callSubscribers()
}

func WatchEffect(fn func()) (stop func()) {
	active := getActiveEffect()
	stopped := false

	var effect *subscriber

	effect = newSubscriber(func() subscriberResult {
		if stopped {
			return unsubscribe
		}

		setActiveEffect(effect)
		fn()
		setActiveEffect(active)

		return keep
	})

	effect.fn()

	return func() {
		stopped = true
	}
}

type Computed[T comparable] struct {
	subscribable
	value         T
	needUpdate    bool
	getter        func() T
	updateTrigger *subscriber
}

func NewComputed[T comparable](getter func() T) *Computed[T] {
	comp := &Computed[T]{
		getter:     getter,
		needUpdate: true,
	}
	comp.updateTrigger = newSubscriber(func() subscriberResult {
		comp.triggerSubscribers()
		return keep
	})

	return comp
}

func (c *Computed[T]) Get() T {
	active := getActiveEffect()

	if c.needUpdate {
		setActiveEffect(c.updateTrigger)
		c.value = c.getter()
		c.needUpdate = false
	}

	c.addSubscriber(active)
	return c.value
}

func (c *Computed[T]) triggerSubscribers() {
	if c.needUpdate {
		return
	}

	c.needUpdate = true

	c.callSubscribers()
}
