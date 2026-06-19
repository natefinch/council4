package rules

import (
	"math/rand/v2"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
)

// scenario is a small fluent builder for rules regression tests. It assembles a
// specific board, hand, library, graveyard, and life state without scripting a
// whole game, then exposes the engine so a step or action can be run and the
// outcome asserted. It keeps regressions concise and reproducible.
//
// Example:
//
//	s := newScenario(t)
//	bear := s.permanent(game.Player1, creatureDef("Bear", 2, 2)).tapped()
//	s.life(game.Player2, 1)
//	losses := s.applyStateBasedActions()
//	// assert on losses / bear.permanent() / s.game()
type scenario struct {
	t   *testing.T
	g   *game.Game
	eng *Engine
}

// newScenario starts an empty four-player game with a deterministically seeded
// engine.
func newScenario(t *testing.T) *scenario {
	t.Helper()
	return &scenario{
		t:   t,
		g:   game.NewGame([game.NumPlayers]game.PlayerConfig{}),
		eng: NewEngine(rand.New(rand.NewPCG(1, 2))),
	}
}

// game returns the assembled game so tests can inspect or further configure it.
func (s *scenario) game() *game.Game { return s.g }

// engine returns the engine driving the scenario.
func (s *scenario) engine() *Engine { return s.eng }

// life sets a player's life total.
func (s *scenario) life(player game.PlayerID, life int) *scenario {
	s.g.Players[player].Life = life
	return s
}

// monarch makes a player the monarch.
func (s *scenario) monarch(player game.PlayerID) *scenario {
	for i := range s.g.Players {
		s.g.Players[i].IsMonarch = game.PlayerID(i) == player
	}
	return s
}

// permanent puts a card onto the battlefield under a controller and returns a
// handle for further tweaks (tapped, counters, ...).
func (s *scenario) permanent(controller game.PlayerID, def *game.CardDef) *permanentHandle {
	cardID := s.newInstance(controller, def)
	permanent := &game.Permanent{
		ObjectID:       s.g.IDGen.Next(),
		CardInstanceID: cardID,
		Owner:          controller,
		Controller:     controller,
	}
	s.g.Battlefield = append(s.g.Battlefield, permanent)
	return &permanentHandle{s: s, perm: permanent}
}

// hand adds a card to a player's hand and returns its instance ID.
func (s *scenario) hand(player game.PlayerID, def *game.CardDef) id.ID {
	cardID := s.newInstance(player, def)
	s.g.Players[player].Hand.Add(cardID)
	return cardID
}

// library adds a card to the bottom of a player's library and returns its ID.
func (s *scenario) library(player game.PlayerID, def *game.CardDef) id.ID {
	cardID := s.newInstance(player, def)
	s.g.Players[player].Library.Add(cardID)
	return cardID
}

// graveyard adds a card to a player's graveyard and returns its instance ID.
func (s *scenario) graveyard(player game.PlayerID, def *game.CardDef) id.ID {
	cardID := s.newInstance(player, def)
	s.g.Players[player].Graveyard.Add(cardID)
	return cardID
}

func (s *scenario) newInstance(owner game.PlayerID, def *game.CardDef) id.ID {
	cardID := s.g.IDGen.Next()
	s.g.CardInstances[cardID] = &game.CardInstance{ID: cardID, Def: def, Owner: owner}
	return cardID
}

// --- runners ---

// applyStateBasedActions runs one round of state-based actions and returns the
// losses recorded.
func (s *scenario) applyStateBasedActions() []LossLog {
	return s.eng.applyStateBasedActions(s.g)
}

// legalActions returns the legal priority actions for a player in the current
// state.
func (s *scenario) legalActions(player game.PlayerID) []action.Action {
	return s.eng.legalActions(s.g, player)
}

// resolveTop resolves the top object of the stack.
func (s *scenario) resolveTop() {
	s.eng.resolveTopOfStack(s.g, &TurnLog{})
}

// permanentHandle configures a permanent added to a scenario.
type permanentHandle struct {
	s    *scenario
	perm *game.Permanent
}

// permanent returns the underlying permanent.
func (h *permanentHandle) permanent() *game.Permanent { return h.perm }

// and returns to the scenario for further building.
func (h *permanentHandle) and() *scenario { return h.s }

func (h *permanentHandle) tapped() *permanentHandle {
	h.perm.Tapped = true
	return h
}

func (h *permanentHandle) summoningSick() *permanentHandle {
	h.perm.SummoningSick = true
	return h
}

func (h *permanentHandle) faceDown() *permanentHandle {
	h.perm.FaceDown = true
	return h
}

func (h *permanentHandle) counter(kind counter.Kind, n int) *permanentHandle {
	h.perm.Counters.Add(kind, n)
	return h
}

func (h *permanentHandle) damage(marked int) *permanentHandle {
	h.perm.MarkedDamage = marked
	return h
}
