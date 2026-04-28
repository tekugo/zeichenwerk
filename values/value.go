package values

import (
	"sync"

	. "github.com/tekugo/zeichenwerk/core"
	. "github.com/tekugo/zeichenwerk/widgets"
)

// Value is a reactive, generic Value.
type Value[T any] struct {
	mu          sync.RWMutex
	value       T
	subscribers []func(T)
}

// NewValue creates a reactive value with an initial value.
func NewValue[T any](initial T) *Value[T] {
	return &Value[T]{value: initial}
}

// Get returns the current value.
func (v *Value[T]) Get() T {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.value
}

// Set sets the new value and notifies all subscribers.
func (v *Value[T]) Set(val T) {
	v.mu.Lock()
	v.value = val
	subs := make([]func(T), len(v.subscribers))
	copy(subs, v.subscribers)
	v.mu.Unlock()

	for _, fn := range subs {
		fn(val)
	}
}

// Bind binds a value to a widget setter.
func (v *Value[T]) Bind(s Setter[T]) *Value[T] {
	v.mu.Lock()
	v.subscribers = append(v.subscribers, s.Set)
	current := v.value
	v.mu.Unlock()
	s.Set(current)
	return v
}

func (v *Value[T]) Observe(widget Widget, convert ...func(any) (T, bool)) {
	widget.On(EvtChange, func(_ Widget, _ Event, params ...any) bool {
		if len(params) == 0 {
			return false
		}
		var val T
		var ok bool
		if len(convert) > 0 {
			val, ok = convert[0](params[0])
		} else {
			val, ok = params[0].(T)
		}
		if ok {
			v.Set(val)
		}
		// We return false, so other EvtChange handlers still get called
		return false
	})
}

// Subscribe adds a new callback function for receiving updates.
func (v *Value[T]) Subscribe(fn func(T)) *Value[T] {
	v.mu.Lock()
	v.subscribers = append(v.subscribers, fn)
	current := v.value
	v.mu.Unlock()
	fn(current)
	return v
}

func Derived[A, B any](source *Value[A], convert func(A) B) *Value[B] {
	derived := NewValue(convert(source.Get()))
	source.Subscribe(func(a A) {
		derived.Set(convert(a))
	})
	return derived
}
