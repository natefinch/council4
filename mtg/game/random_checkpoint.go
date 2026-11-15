package game

import "math/rand/v2"

// randomSource preserves an arbitrary caller's random stream. Words are
// journaled only inside an opaque payment checkpoint, or while replaying it.
type randomSource struct {
	upstream *rand.Rand
	words    []uint64
	cursor   int
	active   bool
}

func (s *randomSource) Uint64() uint64 {
	if s.cursor < len(s.words) {
		word := s.words[s.cursor]
		s.cursor++
		if !s.active && s.cursor == len(s.words) {
			s.words, s.cursor = nil, 0
		}
		return word
	}
	word := s.upstream.Uint64()
	if s.active {
		s.words = append(s.words, word)
		s.cursor++
	}
	return word
}

func (g *Game) setRandom(rng *rand.Rand) {
	g.random = &randomSource{upstream: rng}
	g.RNG = rand.New(g.random)
	g.randomRand = g.RNG
}

// BookmarkRandom starts the one opaque-payment journal. Its result must be
// finished exactly once, with restore true only for a reversible action.
// Clone deliberately continues to fork an independent search stream.
func (g *Game) BookmarkRandom() func(restore bool) {
	if g.random == nil || g.randomRand != g.RNG {
		g.setRandom(g.RNG)
	}
	source := g.random
	if source.active {
		panic("nested opaque random checkpoint")
	}
	source.active = true
	cursor := source.cursor
	return func(restore bool) {
		if !source.active {
			panic("opaque random checkpoint already finished")
		}
		source.active = false
		if restore {
			source.cursor = cursor
		}
		if source.cursor == len(source.words) {
			source.words, source.cursor = nil, 0
		}
	}
}

// OpaquePaymentActive lets a nested mana payment join its caller's checkpoint
// instead of cloning the world or bookmarking the random stream again.
func (g *Game) OpaquePaymentActive() bool {
	return g.random != nil && g.random.active
}

// RestoreOpaquePayment consumes a whole-world snapshot while preserving live
// synchronous scopes and contexts. The caller repairs live object references
// and finishes the random bookmark (CR 734.1–2); IDGen's atomic value is never
// copied.
// Execution diagnostics intentionally continue, so retries consume work budget.
func (g *Game) RestoreOpaquePayment(before *Game) {
	g.restoreOpaquePaymentWorld(before)
	g.Mode = before.Mode
	g.IDGen.Restore(before.IDGen.Current())
}
