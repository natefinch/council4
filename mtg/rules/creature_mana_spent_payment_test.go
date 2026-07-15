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

// creatureManaDork adds a ready (non-summoning-sick) creature permanent that
// taps for one mana of the given color, standing in for a mana creature like
// Llanowar Elves. Mana it produces must be tagged as "from a creature".
func creatureManaDork(g *game.Game, controller game.PlayerID, name string, m mana.Color) *game.Permanent {
	dork := addManaAbilityPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:      name,
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	}}, m, 1)
	dork.SummoningSick = false
	return dork
}

// noncreatureManaRock adds a ready artifact permanent that taps for one mana of
// the given color, standing in for a mana rock like Mind Stone. Mana it produces
// must NOT be tagged as "from a creature".
func noncreatureManaRock(g *game.Game, controller game.PlayerID, name string, m mana.Color) *game.Permanent {
	return addManaAbilityPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:  name,
		Types: []types.Card{types.Artifact},
	}}, m, 1)
}

// castThreeManaCreatureAndReadCreatureMana casts a vanilla {G}{G}{G} creature,
// letting the payment planner auto-tap the battlefield sources already present,
// then returns how much of the mana spent came from creatures as recorded on the
// stack object. This exercises the full source -> pool -> payment -> accounting
// provenance chain that feeds Inga and Esika's draw trigger.
func castThreeManaCreatureAndReadCreatureMana(t *testing.T, g *game.Game, engine *Engine) int {
	t.Helper()
	def := creatureSpellDef("Provenance Beast", types.Beast)
	def.ManaCost = opt.Val(cost.Mana{cost.G, cost.G, cost.G})
	def.Power = opt.Val(game.PT{Value: 3})
	def.Toughness = opt.Val(game.PT{Value: 3})
	spellID := addCardToHand(g, game.Player1, def)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("applyAction(cast three-mana creature) = false, want true")
	}
	obj, ok := g.Stack.Peek()
	if !ok {
		t.Fatal("stack empty after casting creature spell")
	}
	return obj.ManaFromCreaturesSpentToCast
}

// TestCreatureManaSpentAllFromCreatures proves that paying a creature spell
// entirely from creature mana sources records all three mana as creature mana,
// meeting Inga and Esika's "three or more" threshold.
func TestCreatureManaSpentAllFromCreatures(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creatureManaDork(g, game.Player1, "Elf One", mana.G)
	creatureManaDork(g, game.Player1, "Elf Two", mana.G)
	creatureManaDork(g, game.Player1, "Elf Three", mana.G)
	if got := castThreeManaCreatureAndReadCreatureMana(t, g, engine); got != 3 {
		t.Fatalf("creature mana spent = %d, want 3 (all from creatures)", got)
	}
}

// TestCreatureManaSpentExactlyTwoBelowThreshold proves that paying with two
// creature dorks and one non-creature rock records exactly two creature mana,
// which is below Inga and Esika's three-or-more threshold.
func TestCreatureManaSpentExactlyTwoBelowThreshold(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creatureManaDork(g, game.Player1, "Elf One", mana.G)
	creatureManaDork(g, game.Player1, "Elf Two", mana.G)
	noncreatureManaRock(g, game.Player1, "Mana Rock", mana.G)
	if got := castThreeManaCreatureAndReadCreatureMana(t, g, engine); got != 2 {
		t.Fatalf("creature mana spent = %d, want 2 (two creatures, one rock)", got)
	}
}

// TestCreatureManaSpentNoneFromCreatures proves that paying entirely from
// non-creature sources records zero creature mana.
func TestCreatureManaSpentNoneFromCreatures(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	noncreatureManaRock(g, game.Player1, "Rock One", mana.G)
	noncreatureManaRock(g, game.Player1, "Rock Two", mana.G)
	noncreatureManaRock(g, game.Player1, "Rock Three", mana.G)
	if got := castThreeManaCreatureAndReadCreatureMana(t, g, engine); got != 0 {
		t.Fatalf("creature mana spent = %d, want 0 (no creature sources)", got)
	}
}

// TestCreatureManaSpentAcrossMultipleColors proves that creature-mana
// provenance is tracked independently of color: two creature dorks of different
// colors both count toward the creature-mana total.
func TestCreatureManaSpentAcrossMultipleColors(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creatureManaDork(g, game.Player1, "White Elf", mana.W)
	creatureManaDork(g, game.Player1, "Blue Elf", mana.U)

	def := creatureSpellDef("Two-Color Beast", types.Beast)
	def.ManaCost = opt.Val(cost.Mana{cost.W, cost.U})
	def.Power = opt.Val(game.PT{Value: 2})
	def.Toughness = opt.Val(game.PT{Value: 2})
	spellID := addCardToHand(g, game.Player1, def)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("applyAction(cast two-color creature) = false, want true")
	}
	obj, ok := g.Stack.Peek()
	if !ok {
		t.Fatal("stack empty after casting creature spell")
	}
	if obj.ManaFromCreaturesSpentToCast != 2 {
		t.Fatalf("creature mana spent = %d, want 2 (white + blue creature mana)", obj.ManaFromCreaturesSpentToCast)
	}
}
