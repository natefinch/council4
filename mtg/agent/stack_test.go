package agent

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/rules"
)

// counterspellDef is a card the agent casts targeting a spell on the stack; the
// reactive scorer classifies it as a counter by its stack-object target, so its
// own type does not matter.
func counterspellDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:   "Counterspell",
		Types:  []types.Card{types.Instant},
		Colors: []color.Color{color.Blue},
	}}
}

func addStackSpell(g *game.Game, controller game.PlayerID, def *game.CardDef) id.ID {
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{ID: cardID, Def: def, Owner: controller}
	objectID := g.IDGen.Next()
	g.Stack.Push(&game.StackObject{
		ID:         objectID,
		Kind:       game.StackSpell,
		SourceID:   cardID,
		Controller: controller,
	})
	return objectID
}

// TestCounterTheBombNotTheCantrip checks the agent counters a high-impact spell
// (it beats passing) but lets a cheap spell resolve (countering scores below
// Pass), conserving the counter.
func TestCounterTheBombNotTheCantrip(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	counterID := addObservedHandCard(g, game.Player1, counterspellDef())
	bomb := addStackSpell(g, game.Player2, creatureWithCost("Eldrazi", 9, 9, 7))
	cantrip := addStackSpell(g, game.Player2, instantDef("Opt", 1, color.Blue))
	obs := rules.NewObservation(g, game.Player1)
	strategy := GenericStrategy{}

	counterBomb := strategy.ScoreAction(obs, action.CastSpell(counterID, []game.Target{game.StackObjectTarget(bomb)}, 0, nil))
	counterCantrip := strategy.ScoreAction(obs, action.CastSpell(counterID, []game.Target{game.StackObjectTarget(cantrip)}, 0, nil))

	if counterBomb <= scorePass {
		t.Errorf("countering the bomb (%v) should beat passing (%v)", counterBomb, scorePass)
	}
	if counterCantrip >= scorePass {
		t.Errorf("countering the cantrip (%v) should score below passing (%v) so the counter is held", counterCantrip, scorePass)
	}
}

// TestNeverCounterOwnSpell checks the agent does not counter its own spell.
func TestNeverCounterOwnSpell(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	counterID := addObservedHandCard(g, game.Player1, counterspellDef())
	ownSpell := addStackSpell(g, game.Player1, creatureWithCost("My Bomb", 8, 8, 7))
	obs := rules.NewObservation(g, game.Player1)
	strategy := GenericStrategy{}

	score := strategy.ScoreAction(obs, action.CastSpell(counterID, []game.Target{game.StackObjectTarget(ownSpell)}, 0, nil))
	if score >= scorePass {
		t.Errorf("countering own spell (%v) should score well below passing (%v)", score, scorePass)
	}
}

// TestDontWasteRemovalOnSmallCreature checks instant-speed removal is held for a
// worthy target: aiming it at a 1/1 scores below Pass while aiming it at a 6/6
// scores above Pass.
func TestDontWasteRemovalOnSmallCreature(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	// Score on an opponent's turn — removal's natural window — so this exercises
	// the target-value threshold without the own-turn hold-for-later timing.
	g.Turn.ActivePlayer = game.Player2
	removalID := addObservedHandCard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Doom Blade",
		Types: []types.Card{types.Instant},
	}})
	small := addObservedPermanent(g, game.Player2, creatureCardDef("Token", 1, 1))
	big := addObservedPermanent(g, game.Player2, creatureCardDef("Dragon", 6, 6))
	obs := rules.NewObservation(g, game.Player1)
	strategy := GenericStrategy{}

	removalSmall := strategy.ScoreAction(obs, action.CastSpell(removalID, []game.Target{game.PermanentTarget(small.ObjectID)}, 0, nil))
	removalBig := strategy.ScoreAction(obs, action.CastSpell(removalID, []game.Target{game.PermanentTarget(big.ObjectID)}, 0, nil))

	if removalSmall >= scorePass {
		t.Errorf("removing the 1/1 (%v) should score below passing (%v) so removal is held", removalSmall, scorePass)
	}
	if removalBig <= scorePass {
		t.Errorf("removing the 6/6 (%v) should beat passing (%v)", removalBig, scorePass)
	}
}

// TestBeneficialOwnInstantNotHeldAsRemoval checks that an instant targeting only
// the agent's own permanent (a combat trick / protection) stays castable: it is
// scored as a beneficial own-board instant and beats passing, rather than being
// held with a removal self-penalty.
func TestBeneficialOwnInstantNotHeldAsRemoval(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	trickID := addObservedHandCard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Giant Growth",
		Types: []types.Card{types.Instant},
	}})
	mine := addObservedPermanent(g, game.Player1, creatureCardDef("My Bear", 2, 2))
	obs := rules.NewObservation(g, game.Player1)
	strategy := GenericStrategy{}

	score := strategy.ScoreAction(obs, action.CastSpell(trickID, []game.Target{game.PermanentTarget(mine.ObjectID)}, 0, nil))
	if score <= scorePass {
		t.Errorf("a beneficial own-creature instant (%v) should still beat passing (%v), not be held as removal", score, scorePass)
	}
}

func TestCounterPrefersBiggerThreat(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	counterID := addObservedHandCard(g, game.Player1, counterspellDef())
	bigger := addStackSpell(g, game.Player2, creatureWithCost("Wurm", 7, 7, 7))
	smaller := addStackSpell(g, game.Player2, creatureWithCost("Bear", 2, 2, 2))
	obs := rules.NewObservation(g, game.Player1)
	strategy := GenericStrategy{}

	counterBigger := strategy.ScoreAction(obs, action.CastSpell(counterID, []game.Target{game.StackObjectTarget(bigger)}, 0, nil))
	counterSmaller := strategy.ScoreAction(obs, action.CastSpell(counterID, []game.Target{game.StackObjectTarget(smaller)}, 0, nil))
	if counterBigger <= counterSmaller {
		t.Errorf("countering the 7-drop (%v) should outscore countering the 2-drop (%v)", counterBigger, counterSmaller)
	}
}
