package rules

import (
	"testing"

	cards "github.com/natefinch/council4/mtg/cards/c"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
)

// newCyclonusFront puts the real Cyclonus card onto controller's battlefield as
// its front face (Cyclonus, the Saboteur) so its "Whenever Cyclonus deals combat
// damage to a player, it connives. Then if Cyclonus's power is 5 or greater,
// convert it." trigger runs through the real engine path.
func newCyclonusFront(g *game.Game, controller game.PlayerID) *game.Permanent {
	return addCombatPermanent(g, controller, cards.CyclonusTheSaboteur)
}

// newCyclonusBack puts the real Cyclonus card onto controller's battlefield as
// its back face (Cyclonus, Cybertronian Fighter) so its "Whenever Cyclonus deals
// combat damage to a player, convert it. If you do, there is an additional
// beginning phase after this phase." trigger runs through the real engine path.
func newCyclonusBack(g *game.Game, controller game.PlayerID) *game.Permanent {
	permanent := addCombatPermanent(g, controller, cards.CyclonusTheSaboteur)
	permanent.Face = game.FaceBack
	permanent.Transformed = true
	return permanent
}

// TestCyclonusFrontConnivesWithoutTransformBelowFive proves the front face's
// combat-damage trigger connives but does not convert when Cyclonus's power
// stays below five: the power-gated "convert it" fails, so the permanent remains
// its front face.
func TestCyclonusFrontConnivesWithoutTransformBelowFive(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cyclonus := newCyclonusFront(g, game.Player1)
	// A single land in the library is drawn and then discarded by connive, so no
	// +1/+1 counter is placed and Cyclonus's power stays at its base 2 (< 5).
	land := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Island",
		Types: []types.Card{types.Land},
	}})

	dealPlayerDamage(g, cyclonus.ObjectID, cyclonus.ObjectID, game.Player1, game.Player2, 2, true)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("Cyclonus front combat-damage trigger was not put on stack")
	}
	engine.resolveTopOfStackWithChoices(g, allFirstLegalAgents(), &TurnLog{})

	if !g.Players[game.Player1].Graveyard.Contains(land) {
		t.Fatal("connive did not discard the drawn card")
	}
	if cyclonus.Transformed || cyclonus.Face != game.FaceFront {
		t.Fatalf("Cyclonus converted with power below 5 (Face=%v, Transformed=%v)", cyclonus.Face, cyclonus.Transformed)
	}
}

// TestCyclonusFrontConnivesAndTransformsAtFive proves the front face's
// combat-damage trigger connives and then converts when Cyclonus's power is five
// or greater: the power-gated "convert it" resolves, flipping the permanent to
// its back face.
func TestCyclonusFrontConnivesAndTransformsAtFive(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cyclonus := newCyclonusFront(g, game.Player1)
	// Boost Cyclonus's base power of 2 to 6 so the "power is 5 or greater" gate
	// is satisfied when the convert instruction resolves.
	cyclonus.Counters.Add(counter.PlusOnePlusOne, 4)
	spell := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Lightning Bolt",
		Types: []types.Card{types.Instant},
	}})

	dealPlayerDamage(g, cyclonus.ObjectID, cyclonus.ObjectID, game.Player1, game.Player2, 2, true)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("Cyclonus front combat-damage trigger was not put on stack")
	}
	engine.resolveTopOfStackWithChoices(g, allFirstLegalAgents(), &TurnLog{})

	if !g.Players[game.Player1].Graveyard.Contains(spell) {
		t.Fatal("connive did not discard the drawn card")
	}
	if !cyclonus.Transformed || cyclonus.Face != game.FaceBack {
		t.Fatalf("Cyclonus did not convert with power >= 5 (Face=%v, Transformed=%v)", cyclonus.Face, cyclonus.Transformed)
	}
}

// TestCyclonusBackConvertsAndAddsBeginningPhase proves the back face's
// combat-damage trigger converts Cyclonus and, because it did, queues an
// additional beginning phase after the current phase.
func TestCyclonusBackConvertsAndAddsBeginningPhase(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cyclonus := newCyclonusBack(g, game.Player1)
	g.Turn.Phase = game.PhasePostcombatMain

	dealPlayerDamage(g, cyclonus.ObjectID, cyclonus.ObjectID, game.Player1, game.Player2, 5, true)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("Cyclonus back combat-damage trigger was not put on stack")
	}
	engine.resolveTopOfStackWithChoices(g, allFirstLegalAgents(), &TurnLog{})

	if cyclonus.Transformed || cyclonus.Face != game.FaceFront {
		t.Fatalf("Cyclonus did not convert to its front face (Face=%v, Transformed=%v)", cyclonus.Face, cyclonus.Transformed)
	}
	if len(g.Turn.ExtraPhases) != 1 || g.Turn.ExtraPhases[0] != game.PhaseBeginning {
		t.Fatalf("queued extra phases = %#v, want one beginning phase", g.Turn.ExtraPhases)
	}

	// Draining the queue actually runs the additional beginning phase: the active
	// player draws for the extra draw step.
	drawn := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Mountain",
		Types: []types.Card{types.Land},
	}})
	engine.runExtraPhases(g, allFirstLegalAgents(), &TurnLog{})

	if len(g.Turn.ExtraPhases) != 0 {
		t.Fatalf("extra phases not drained: %#v", g.Turn.ExtraPhases)
	}
	if !g.Players[game.Player1].Hand.Contains(drawn) {
		t.Fatal("additional beginning phase did not run its draw step")
	}
}
