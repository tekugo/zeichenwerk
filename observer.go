package zeichenwerk

type Observable interface {
	Subscribe(func(any))
}

type Subject struct {
	observers []func(any)
}

func (s *Subject) Subscribe(fn func(any)) {
	s.observers = append(s.observers, fn)
}

func (s *Subject) Notify(data any) {
	for _, fn := range s.observers {
		fn(data)
	}
}
