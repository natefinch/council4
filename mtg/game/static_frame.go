package game

// staticSourceFrame memoizes a pure-read computation for the duration of a
// "frame". The rules layer rebuilds the set of permanents and graveyard cards
// that carry static continuous abilities every time it evaluates a permanent's
// effective characteristics; that scan is the dominant cost of long games. A
// frame lets the rules layer build that set once and reuse it while it is
// repeatedly evaluating permanents.
//
// Frames must wrap reads only. While a frame is open the caller guarantees the
// game state that the cached value depends on does not change, so the cached
// value can never be stale within the frame. The value is held as an opaque
// any owned by the rules layer.
type staticSourceFrame struct {
	depth int
	value any
	set   bool
}

// BeginStaticSourceFrame opens (or re-enters) a static-source frame. Frames are
// reentrant: nested Begin/End calls share one frame and the cache is cleared
// only when the outermost frame ends. A frame must wrap reads that do not
// mutate the game state the cached value depends on, and every Begin must be
// paired with an End (typically via defer).
func (g *Game) BeginStaticSourceFrame() {
	if g.staticFrame == nil {
		g.staticFrame = &staticSourceFrame{}
	}
	g.staticFrame.depth++
}

// EndStaticSourceFrame closes one level of the static-source frame, discarding
// the cache when the outermost frame ends.
func (g *Game) EndStaticSourceFrame() {
	if g.staticFrame == nil {
		return
	}
	g.staticFrame.depth--
	if g.staticFrame.depth <= 0 {
		g.staticFrame = nil
	}
}

// InStaticSourceFrame reports whether a static-source frame is currently open.
func (g *Game) InStaticSourceFrame() bool {
	return g.staticFrame != nil
}

// StaticSourceFrameValue returns the cached frame value and whether a value has
// been cached in the current frame. It returns false when no frame is open.
func (g *Game) StaticSourceFrameValue() (any, bool) {
	if g.staticFrame == nil || !g.staticFrame.set {
		return nil, false
	}
	return g.staticFrame.value, true
}

// SetStaticSourceFrameValue caches a value for the current frame. It is a no-op
// when no frame is open, so callers outside a frame always recompute.
func (g *Game) SetStaticSourceFrameValue(value any) {
	if g.staticFrame == nil {
		return
	}
	g.staticFrame.value = value
	g.staticFrame.set = true
}
