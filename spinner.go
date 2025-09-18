package zeichenwerk

import "time"

var Spinners = map[string]string{
	"bar":     "|/-\\",
	"dots":    ".oOo",
	"dot":     "⠁⠂⠄⡀⢀⠠⠐⠈",
	"arrow":   "←↖↑↗→↘↓↙",
	"circle":  "◐◓◑◒",
	"bounce":  "⠁⠂⠄⠂",
	"braille": "⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏",
}

type Spinner struct {
	BaseWidget
	runes  []rune        // Spinning runes to cycle through
	index  int           // current rune index
	ticker *time.Ticker  // ticker
	stop   chan struct{} // stop channel
}

func NewSpinner(id string, runes []rune) *Spinner {
	spinner := &Spinner{
		BaseWidget: BaseWidget{id: id, focusable: false},
		runes:      runes,
		stop:       make(chan struct{}),
	}
	return spinner
}

func (s *Spinner) Hint() (int, int) {
	return 1, 1
}

func (s *Spinner) Refresh() {
	Redraw(s)
}

func (s *Spinner) Start(interval time.Duration) {
	// Starting the spinner will block, so we start it as a separate go routine
	go func() {
		if s.ticker != nil {
			panic("ticker for spinner already started")
		}

		s.ticker = time.NewTicker(interval)
		defer s.ticker.Stop()

		for {
			select {
			case <-s.stop:
				s.Log(s, "debug", "Spinner stopped")
				s.ticker = nil
				return
			case <-s.ticker.C:
				s.index++
				if s.index >= len(s.runes) {
					s.index = 0
				}
				s.Refresh()
			}
		}
	}()
}

func (s *Spinner) Rune() rune {
	return s.runes[s.index]
}

func (s *Spinner) Stop() {
	select {
	case s.stop <- struct{}{}:
	default:
	}
}
