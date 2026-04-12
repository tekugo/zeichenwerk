package zeichenwerk

import (
	"sync"
	"time"
)

// ==== AI ===================================================================

// crtPhase represents the current lifecycle state of the CRT container.
type crtPhase int

const (
	crtPhaseOn   crtPhase = iota // power-on animation running
	crtPhaseIdle                 // fully on — child renders normally
	crtPhaseOff                  // power-off animation running
)

// crtFadeDuration is the number of frames a row takes to fade from green to
// true colour after being revealed (power-on) or before being swallowed by
// the matrix (power-off).
const crtFadeDuration = 8

// crtGreen is the brightness ramp used for the matrix rain: index 0 is
// near-black, index 7 is bright phosphor green.
var crtGreen = [8]string{
	"#001400", "#002200", "#003500", "#004d00",
	"#006600", "#009900", "#00cc00", "#00ff41",
}

// crtTintColors are the two alternating greens used for the pulsating child
// overlay: bright phosphor green and a lighter near-white flash.
var crtTintColors = [2]string{"#00ff41", "#88ffaa"}

// crtCells is the character pool for the matrix rain.  Spaces appear more
// often than glyphs to keep the field sparse and irregular.
var crtCells = [24]string{
	"0", "1", "!", "@", "#", "$", "|", "\\",
	"/", "%", "*", "?", "+", "-", "~", "^",
	" ", " ", " ", " ", " ", "0", "1", "|",
}

// crtHash mixes x, y, and a time counter into a single pseudo-random integer
// that is both irregular spatially and temporally. Using multiplicative
// constants from LCG theory breaks up the linear patterns produced by naive
// sums.
func crtHash(x, y, p int) int {
	h := x*1664525 + y*1013904223 + p*22695477
	h ^= h >> 11
	h ^= h << 15
	h &= 0x7fffffff
	return h
}

// CRT is an animated root container that simulates a CRT monitor powering on
// and off. On startup the child widget is revealed by expanding a horizontal
// band from the screen centre outward (symmetrically up and down) until the
// full height is visible. During normal operation CRT is an invisible
// pass-through wrapper. Calling PowerOff contracts the band back to a line
// and then invokes the provided callback (typically ui.Quit).
//
// The animation areas are filled with a Matrix-style green character rain that
// grows brighter toward the scan edge, flanked by a flashing phosphor scanline.
// Rows of the child content fade from green monochrome to true colour after
// they are revealed (power-on) or before they are swallowed (power-off).
// The final frames of power-on flicker between green and true colour.
//
// Typical usage:
//
//	crt := NewCRT("crt", "")
//	crt.Add(myRootContainer)
//	ui := NewUI(theme, crt)
//	crt.Start(30 * time.Millisecond)
//	// … wire PowerOff to a quit shortcut:
//	// crt.PowerOff(30*time.Millisecond, ui.Quit)
//	ui.Run()
type CRT struct {
	Component
	mu    sync.Mutex
	child Widget

	ticker *time.Ticker
	stop   chan struct{} // buffered(1); created fresh for each animation run
	done   chan struct{} // closed by goroutine defer when it exits

	phase crtPhase
	step  int // half-band height in rows (centre to top/bottom edge)
	end   int // target half-band height = (height+1)/2, set in Layout
	pulse int // frame counter; drives character and colour variation
}

// NewCRT creates a CRT container. Call Start to begin the power-on animation
// after the UI layout has run (i.e. after screen dimensions are known).
func NewCRT(id, class string) *CRT {
	return &CRT{
		Component: Component{id: id, class: class},
		phase:     crtPhaseOn,
		step:      1,
	}
}

// ---- Widget Methods -------------------------------------------------------

// Apply applies a theme's styles to the component.
func (c *CRT) Apply(theme *Theme) {
	theme.Apply(c, c.Selector("crt"))
}

// Hint returns the preferred size of the child widget including its style overhead.
func (c *CRT) Hint() (int, int) {
	if c.hwidth != 0 || c.hheight != 0 {
		return c.hwidth, c.hheight
	}
	if c.child == nil {
		return 0, 0
	}
	w, h := c.child.Hint()
	style := c.child.Style()
	w += style.Horizontal()
	h += style.Vertical()
	return w, h
}

// Layout positions the child to fill the CRT widget's full bounds and records
// the animation end position.
func (c *CRT) Layout() error {
	if c.child != nil {
		cx, cy, cw, ch := c.Bounds()
		c.child.SetBounds(cx, cy, cw, ch)
		c.mu.Lock()
		c.end = (ch + 1) / 2
		c.mu.Unlock()
	}
	return Layout(c)
}

// Render draws the child clipped to the current animation band. During
// crtPhaseIdle the child is rendered without any clipping.
func (c *CRT) Render(r *Renderer) {
	if c.child == nil {
		return
	}

	c.mu.Lock()
	phase := c.phase
	step := c.step
	ticker := c.ticker
	pulse := c.pulse
	end := c.end
	c.mu.Unlock()

	switch phase {
	case crtPhaseIdle:
		c.child.Render(r)

	case crtPhaseOn:
		if ticker == nil {
			return // Start not yet called — keep screen blank
		}
		c.renderCRTEffect(r, phase, step, pulse, end)

	case crtPhaseOff:
		c.renderCRTEffect(r, phase, step, pulse, end)
	}
}

// renderCRTEffect is the main compositor for both power-on and power-off.
// It draws: matrix rain in non-band areas, pulsating scanlines at band edges,
// child content clipped to the band, and a green phosphor tint on the child.
func (c *CRT) renderCRTEffect(r *Renderer, phase crtPhase, step, pulse, end int) {
	bandH := step * 2
	if bandH > c.height {
		bandH = c.height
	}
	if bandH < 0 {
		bandH = 0
	}

	clipY := c.y + (c.height-bandH)/2
	topEnd := clipY              // top matrix area:    [c.y, topEnd)
	bottomStart := clipY + bandH // bottom matrix area: [bottomStart, c.y+c.height)

	// Matrix rain: brighter near each scanline, darker toward screen edges.
	c.renderMatrix(r, c.y, topEnd, pulse, topEnd-1)
	c.renderMatrix(r, bottomStart, c.y+c.height, pulse, bottomStart)

	// Flashing phosphor scanlines just outside the band edges.
	c.renderScanline(r, topEnd-1, pulse)
	c.renderScanline(r, bottomStart, pulse)

	// Render child content clipped to the visible band.
	if bandH > 0 {
		r.Clip(c.x, clipY, c.width, bandH)
		r.Translate(-c.x, -c.y)
		c.child.Render(r)
		r.Clip(0, 0, 0, 0)
		r.Translate(0, 0)

		// Overlay green phosphor tint on the child content.
		// Power-on end: flicker between green and true colour.
		// All other animation frames: full green tint.
		flickering := phase == crtPhaseOn && end > 0 && step >= end-4
		if flickering {
			if pulse%2 == 1 {
				c.renderChildTint(r, clipY, bandH, pulse)
			}
			// even pulse → true colour, no tint
		} else {
			c.renderChildTint(r, clipY, bandH, pulse)
		}
	}
}

// renderMatrix fills a horizontal strip [y1, y2) with matrix rain characters.
// scanY is the y-coordinate of the adjacent scanline; rows closer to scanY
// receive brighter green shades.  Column speed variation and a good hash
// function keep the pattern irregular and non-repeating.
func (c *CRT) renderMatrix(r *Renderer, y1, y2, pulse, scanY int) {
	const maxDist = 10 // rows over which the brightness gradient runs
	const numCells = len(crtCells)
	const numGreens = len(crtGreen)

	for y := y1; y < y2; y++ {
		distFromScan := y - scanY
		if distFromScan < 0 {
			distFromScan = -distFromScan
		}

		// brightLevel: 7 = right at scanline, 0 = far away.
		brightLevel := 0
		if distFromScan < maxDist {
			brightLevel = 7 - (distFromScan*7)/maxDist
		}

		for x := c.x; x < c.x+c.width; x++ {
			// Each column updates its characters at its own speed so columns
			// appear to "fall" at different rates.
			colRate := 1 + x%5
			h := crtHash(x, y, pulse/colRate)
			ci := h % numCells

			// Colour: base level from scanline proximity, ±1 random variation.
			h2 := crtHash(x, y, pulse/3)
			gi := brightLevel + (h2%3 - 1)
			if gi < 0 {
				gi = 0
			} else if gi >= numGreens {
				gi = numGreens - 1
			}

			r.Set(crtGreen[gi], "black", "")
			r.Put(x, y, crtCells[ci])
		}
	}
}

// renderScanline draws a single bright ━ line at row y, alternating between
// two phosphor-green shades every frame for a pulsating glow.
func (c *CRT) renderScanline(r *Renderer, y, pulse int) {
	if y < c.y || y >= c.y+c.height {
		return
	}
	if pulse&1 == 0 {
		r.Set("#00ff41", "#001a00", "bold")
	} else {
		r.Set("#aaffaa", "#003300", "bold")
	}
	r.Fill(c.x, y, c.width, 1, "━")
}

// renderChildTint overlays a pulsating green phosphor wash on every row of
// the visible band.  Every row is covered — no gaps in true colour.
// A per-row brightness alternation driven by pulse makes the tint shimmer.
func (c *CRT) renderChildTint(r *Renderer, clipY, bandH, pulse int) {
	for y := clipY; y < clipY+bandH; y++ {
		// Alternate between the two tint shades so the green field shimmers.
		r.Set(crtTintColors[(y+pulse)%2], "black", "")
		r.Colorize(c.x, y, c.width, 1)
	}
}

// ---- Container Methods ----------------------------------------------------

// Add sets the single child widget, replacing any previous child.
func (c *CRT) Add(widget Widget, _ ...any) error {
	if c.child != nil {
		c.child.SetParent(nil)
	}
	if widget != nil {
		widget.SetParent(c)
	}
	c.child = widget
	return nil
}

// Children returns the child widget slice (empty if no child has been set).
func (c *CRT) Children() []Widget {
	if c.child != nil {
		return []Widget{c.child}
	}
	return []Widget{}
}

// ---- Animation Control ----------------------------------------------------

// Start begins the power-on animation at the given tick interval. Should be
// called once after the UI layout has run so that the screen height is known.
// If an animation is already running this is a no-op.
func (c *CRT) Start(interval time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.ticker != nil {
		return
	}
	c.phase = crtPhaseOn
	c.step = 1
	c.startAnimation(interval)
}

// PowerOff begins the power-off animation at the given tick interval. When
// the animation completes, onDone is called (typically ui.Quit). If the
// power-on animation is still running, PowerOff interrupts it and begins
// contracting from the current position. Calling PowerOff a second time
// while a power-off animation is already running is a no-op.
func (c *CRT) PowerOff(interval time.Duration, onDone func()) {
	c.mu.Lock()

	// If the power-on goroutine is still running, interrupt it and wait for
	// it to exit so we can start a clean power-off animation.
	if c.ticker != nil && c.phase == crtPhaseOn {
		stopCh := c.stop
		doneCh := c.done
		c.mu.Unlock()

		select {
		case stopCh <- struct{}{}:
		default:
		}
		<-doneCh // block until goroutine defer closes done

		c.mu.Lock()
	}

	if c.ticker != nil {
		c.mu.Unlock()
		return // already animating (double PowerOff guard)
	}

	// Start the power-off band from full height when idle, or from wherever
	// the power-on animation stopped when interrupted.
	if c.phase == crtPhaseIdle {
		c.step = c.end
	}
	c.phase = crtPhaseOff
	c.startAnimation(interval)
	savedDone := c.done
	c.mu.Unlock()

	go func() {
		<-savedDone
		if onDone != nil {
			onDone()
		}
	}()
}

// ---- Internal helpers -----------------------------------------------------

// startAnimation creates fresh channels and starts the animation goroutine.
// Must be called with c.mu held.
func (c *CRT) startAnimation(interval time.Duration) {
	c.stop = make(chan struct{}, 1)
	c.done = make(chan struct{})
	c.ticker = time.NewTicker(interval)

	// Capture locals so the goroutine does not race with PowerOff replacing fields.
	stopCh := c.stop
	doneCh := c.done
	tickerRef := c.ticker

	go func() {
		defer func() {
			tickerRef.Stop()
			c.mu.Lock()
			if c.ticker == tickerRef {
				c.ticker = nil
			}
			c.mu.Unlock()
			close(doneCh)
		}()

		for {
			select {
			case <-stopCh:
				return
			case <-tickerRef.C:
				c.mu.Lock()
				keepGoing := c.tick()
				c.mu.Unlock()
				if !keepGoing {
					return
				}
			}
		}
	}()
}

// tick advances one animation frame. Returns false when the animation is
// complete and the goroutine should exit. Must be called with c.mu held.
func (c *CRT) tick() bool {
	c.pulse++
	switch c.phase {
	case crtPhaseOn:
		c.step++
		if c.step >= c.end {
			c.step = c.end
			c.phase = crtPhaseIdle
			Redraw(c)
			return false
		}
		Redraw(c)
		return true

	case crtPhaseOff:
		c.step--
		if c.step <= 0 {
			c.step = 0
			Redraw(c)
			return false
		}
		Redraw(c)
		return true
	}
	return false
}
