package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards/authored"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Source review: pinned Oracle corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4.
// Agony Warp (Oracle ID 377c217b-ddad-4e1a-9122-f347113d3dfa, printing
// revision 5890a23e-1da0-432b-8dc8-85b92dee9ba1) has two target-creature
// paragraphs, -3/-0 and -0/-3 until end of turn. Its Scryfall rulings URI
// https://api.scryfall.com/cards/5890a23e-1da0-432b-8dc8-85b92dee9ba1/rulings
// returned the WotC 2017-03-14 ruling that the targets may be the same or
// different creature.
// Heated Debate (Oracle ID 06c3e85c-17d8-44e1-a0f8-07ef3f1624cb, printing
// revision f40c88e2-28ed-4e5b-bcb3-4397d6a15a6f) can't be countered and deals
// 4 damage to target creature or planeswalker. Its Scryfall rulings URI
// https://api.scryfall.com/cards/f40c88e2-28ed-4e5b-bcb3-4397d6a15a6f/rulings
// returned the WotC 2021-04-16 ruling that an illegal target means no damage.
// Slice from the Shadows (Oracle ID 4d31d3f4-4748-4713-808f-24699fcd7b10,
// printing revision 317800b3-6b2d-4de6-8e44-7e54dd623055) can't be countered and
// gives target creature -X/-X until end of turn. Its Scryfall rulings URI
// https://api.scryfall.com/cards/317800b3-6b2d-4de6-8e44-7e54dd623055/rulings
// returned the WotC 2024-02-02 ruling that ward's cost may be paid, but the
// spell is not countered if it is not.
// Rulings retrieved 2026-11-11. Official Wizards Comprehensive Rules effective
// 2026-11-06, https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt:
// read 107.3a, 601.2b/c/f, 608.2b/c, 611.2a, 613.4c, 701.6a, and 702.21a.
// Capability trace at 0f9cdd7896d79b069b3c8438def369c57cd48a8e: separate
// TargetSpec slots overlap unless DistinctFromPriorTargets is set
// (mtg/game/ability.go, mtg/rules/targets.go); ModifyPT/Quantity resolve through
// mtg/rules/primitive_handlers.go; Damage uses the shared operation in
// mtg/rules/primitive_handlers_damage.go; CantBeCounteredStaticBody is the
// stack-scoped rule effect from mtg/game/mechanic_template.go.
func TestAgonyWarpScenario(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCombatPermanent(g, game.Player2, vanillaCreature("Agony target", 5, 5))
	spell := addCardToHand(g, game.Player1, authored.AgonyWarp.New())
	g.Players[game.Player1].ManaPool.Add(mana.U, 1)
	g.Players[game.Player1].ManaPool.Add(mana.B, 1)
	setMainPhasePriority(g, game.Player1)

	if !engine.applyAction(g, game.Player1, action.CastSpell(spell, []game.Target{
		game.PermanentTarget(target.ObjectID),
		game.PermanentTarget(target.ObjectID),
	}, 0, nil)) {
		t.Fatal("Agony Warp did not cast with the same creature chosen for both target instances")
	}
	if g.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatal("Agony Warp did not pay {U}{B}")
	}

	engine.resolveTopOfStack(g, &TurnLog{})

	if got := effectivePower(g, target); got != 2 {
		t.Fatalf("target power = %d, want 2 after -3/-0", got)
	}
	if got, ok := effectiveToughness(g, target); !ok || got != 2 {
		t.Fatalf("target toughness = %d, want 2 after -0/-3", got)
	}
	if !g.Players[game.Player1].Graveyard.Contains(spell) {
		t.Fatal("resolved Agony Warp did not go to its owner's graveyard")
	}
}

func TestHeatedDebateScenario(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name: "Heated Debate target", Types: []types.Card{types.Planeswalker}, Loyalty: opt.Val(6),
	}})
	target.Counters.Add(counter.Loyalty, 6)
	spell := addCardToHand(g, game.Player1, authored.HeatedDebate.New())
	g.Players[game.Player1].ManaPool.Add(mana.R, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)
	setMainPhasePriority(g, game.Player1)

	if !engine.applyAction(g, game.Player1, action.CastSpell(spell, []game.Target{
		game.PermanentTarget(target.ObjectID),
	}, 0, nil)) {
		t.Fatal("Heated Debate did not cast targeting a planeswalker")
	}
	if g.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatal("Heated Debate did not pay {2}{R}")
	}
	assertAuthoredStatDamageSpellCannotBeCountered(t, g)

	engine.resolveTopOfStack(g, &TurnLog{})

	if got := target.Counters.Get(counter.Loyalty); got != 2 {
		t.Fatalf("planeswalker loyalty = %d, want 2 after 4 damage", got)
	}
	if !g.Players[game.Player1].Graveyard.Contains(spell) {
		t.Fatal("resolved Heated Debate did not go to its owner's graveyard")
	}
}

func TestSliceFromTheShadowsScenario(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCombatPermanent(g, game.Player2, vanillaCreature("Slice target", 4, 4))
	spell := addCardToHand(g, game.Player1, authored.SliceFromTheShadows.New())
	g.Players[game.Player1].ManaPool.Add(mana.B, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)
	setMainPhasePriority(g, game.Player1)

	if !engine.applyAction(g, game.Player1, action.CastSpell(spell, []game.Target{
		game.PermanentTarget(target.ObjectID),
	}, 2, nil)) {
		t.Fatal("Slice from the Shadows did not cast with X=2")
	}
	if g.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatal("Slice from the Shadows did not pay {2}{B}")
	}
	assertAuthoredStatDamageSpellCannotBeCountered(t, g)

	engine.resolveTopOfStack(g, &TurnLog{})

	if got := effectivePower(g, target); got != 2 {
		t.Fatalf("target power = %d, want 2 after -X/-X at X=2", got)
	}
	if got, ok := effectiveToughness(g, target); !ok || got != 2 {
		t.Fatalf("target toughness = %d, want 2 after -X/-X at X=2", got)
	}
	if !g.Players[game.Player1].Graveyard.Contains(spell) {
		t.Fatal("resolved Slice from the Shadows did not go to its owner's graveyard")
	}
}

func assertAuthoredStatDamageSpellCannotBeCountered(t *testing.T, g *game.Game) {
	t.Helper()
	spell, ok := g.Stack.Peek()
	if !ok {
		t.Fatal("spell missing before counterability check")
	}
	if counterStackObject(g, spell.ID) {
		t.Fatal("spell was countered despite its Oracle text")
	}
	if _, ok := stackObjectByID(g, spell.ID); !ok {
		t.Fatal("uncounterable spell left the stack")
	}
}
