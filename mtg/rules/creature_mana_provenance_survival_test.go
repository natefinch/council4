package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TestCreatureManaProvenanceSurvivesSourceLeaving proves that creature-mana
// provenance is captured at production time on the mana unit itself, so it
// survives the producing creature leaving the battlefield before the mana is
// spent. A creature dork taps for floating mana, then leaves; the mana it left
// in the pool is still "from a creature" when later spent to cast a spell.
func TestCreatureManaProvenanceSurvivesSourceLeaving(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	dork := creatureManaDork(g, game.Player1, "Doomed Elf", mana.G)

	if !engine.applyAction(g, game.Player1, action.ActivateAbility(dork.ObjectID, 0, nil, 0)) {
		t.Fatal("activating creature dork mana ability = false, want true")
	}
	pool := &g.Players[game.Player1].ManaPool
	if got := pool.CreatureAmount(); got != 1 {
		t.Fatalf("floating creature mana = %d, want 1 (produced by a creature)", got)
	}

	// The creature leaves the battlefield after producing its mana.
	g.Battlefield = removePermanent(g.Battlefield, dork)

	def := creatureSpellDef("One-Drop Beast", types.Beast)
	def.ManaCost = opt.Val(cost.Mana{cost.G})
	def.Power = opt.Val(game.PT{Value: 1})
	def.Toughness = opt.Val(game.PT{Value: 1})
	spellID := addCardToHand(g, game.Player1, def)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("applyAction(cast from left-behind creature mana) = false, want true")
	}
	obj, ok := g.Stack.Peek()
	if !ok {
		t.Fatal("stack empty after casting creature spell")
	}
	if obj.ManaFromCreaturesSpentToCast != 1 {
		t.Fatalf("creature mana spent = %d, want 1 (provenance survived source leaving)", obj.ManaFromCreaturesSpentToCast)
	}
}

// TestCreatureManaProvenanceSurvivesTypeChange proves the same provenance is
// fixed at production time even if the source stops being a creature afterward:
// the mana already in the pool keeps its creature provenance regardless of the
// source's current types.
func TestCreatureManaProvenanceSurvivesTypeChange(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	dork := creatureManaDork(g, game.Player1, "Shifting Elf", mana.G)

	if !engine.applyAction(g, game.Player1, action.ActivateAbility(dork.ObjectID, 0, nil, 0)) {
		t.Fatal("activating creature dork mana ability = false, want true")
	}
	pool := &g.Players[game.Player1].ManaPool
	if got := pool.CreatureAmount(); got != 1 {
		t.Fatalf("floating creature mana = %d, want 1 (produced by a creature)", got)
	}

	// The source loses its creature type after producing mana. Provenance on the
	// already-produced mana must not change.
	card, ok := permanentCardDef(g, dork)
	if !ok {
		t.Fatal("dork card definition not found")
	}
	card.Types = []types.Card{types.Artifact}

	if got := pool.CreatureAmount(); got != 1 {
		t.Fatalf("floating creature mana after type change = %d, want 1 (provenance is fixed at production)", got)
	}

	def := creatureSpellDef("One-Drop Beast", types.Beast)
	def.ManaCost = opt.Val(cost.Mana{cost.G})
	def.Power = opt.Val(game.PT{Value: 1})
	def.Toughness = opt.Val(game.PT{Value: 1})
	spellID := addCardToHand(g, game.Player1, def)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("applyAction(cast from now-artifact source's creature mana) = false, want true")
	}
	obj, ok := g.Stack.Peek()
	if !ok {
		t.Fatal("stack empty after casting creature spell")
	}
	if obj.ManaFromCreaturesSpentToCast != 1 {
		t.Fatalf("creature mana spent = %d, want 1 (provenance survived type change)", obj.ManaFromCreaturesSpentToCast)
	}
}
