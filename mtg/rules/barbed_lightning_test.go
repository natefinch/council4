package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Source review: pinned Oracle corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4;
// Barbed Lightning, DST 55 printing revision
// 2509482a-68d8-4e94-9d1e-5b069ebdc2e4, Oracle ID
// f6e1eb52-9f13-425c-8309-1b08ed6557db, exact text:
// "Choose one —
// • Barbed Lightning deals 3 damage to target creature.
// • Barbed Lightning deals 3 damage to target player or planeswalker.
// Entwine {2} (Choose both if you pay the entwine cost.)"
// Current Scryfall API record
// https://api.scryfall.com/cards/2509482a-68d8-4e94-9d1e-5b069ebdc2e4
// retrieved 2026-11-20 with the same identity and text. Its rulings URI
// https://api.scryfall.com/cards/2509482a-68d8-4e94-9d1e-5b069ebdc2e4/rulings
// returned an empty data list.
// Official Wizards Comprehensive Rules effective 2026-11-06, verified through
// https://magic.wizards.com/en/rules and read from
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt:
// 700.2 defines modes; 601.2b/c selects modes and only their targets, while
// 601.2f-h determines and pays costs; 702.42a-b permits all modes for the
// additional entwine cost and resolves them in printed order. Rules 115.1 and
// 608.2b govern target announcement and revalidation; 120.3a/c/e makes player
// damage reduce life, planeswalker damage remove loyalty, and creature damage
// be marked. The scenario exercises both entwined modes, the additional cost,
// the player target, and the planeswalker alternative. The declarations reuse
// the existing authored entwine seam in Rain of Rust and the shared Damage
// recipient/target operations used by Heated Debate.
func TestBarbedLightningEntwinedCreatureAndPlayerDamage(t *testing.T) {
	registry := cards.NewDefaultRegistry()
	def, err := registry.LookupOracleID("f6e1eb52-9f13-425c-8309-1b08ed6557db")
	if err != nil {
		t.Fatal(err)
	}
	if def.Name != "Barbed Lightning" {
		t.Fatalf("Oracle ID lookup returned %q", def.Name)
	}

	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creature := addCombatPermanent(g, game.Player2, vanillaCreature("Barbed Lightning target", 5, 5))
	planeswalker := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name: "Barbed Lightning planeswalker", Types: []types.Card{types.Planeswalker},
		Loyalty: opt.Val(5),
	}})
	planeswalker.Counters.Add(counter.Loyalty, 5)
	entwinedSpell := addCardToHand(g, game.Player1, def)
	planeswalkerSpell := addCardToHand(g, game.Player1, def)
	g.Players[game.Player1].ManaPool.Add(mana.C, 4)
	g.Players[game.Player1].ManaPool.Add(mana.R, 6)

	engine := NewEngineWithPrograms(nil, registry.Programs())
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.CastSpell(entwinedSpell, []game.Target{
		game.PermanentTarget(creature.ObjectID),
		game.PlayerTarget(game.Player2),
	}, 0, []int{0, 1})) {
		t.Fatal("Barbed Lightning did not cast with both modes and the entwine cost")
	}
	if got := g.Players[game.Player1].ManaPool.Total(); got != 5 {
		t.Fatalf("mana remaining after entwined cast = %d, want 5", got)
	}
	stackObject, ok := g.Stack.Peek()
	if !ok || len(stackObject.ChosenModes) != 2 ||
		stackObject.ChosenModes[0] != 0 || stackObject.ChosenModes[1] != 1 {
		t.Fatalf("entwined stack modes = %v, want [0 1]", stackObject.ChosenModes)
	}

	engine.resolveTopOfStack(g, &TurnLog{})
	if creature.MarkedDamage != 3 {
		t.Fatalf("creature damage marked = %d, want 3", creature.MarkedDamage)
	}
	if got := g.Players[game.Player2].Life; got != 37 {
		t.Fatalf("player life = %d, want 37 after 3 damage", got)
	}

	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.CastSpell(planeswalkerSpell, []game.Target{
		game.PermanentTarget(planeswalker.ObjectID),
	}, 0, []int{1})) {
		t.Fatal("Barbed Lightning did not cast its player-or-planeswalker mode targeting a planeswalker")
	}
	if got := g.Players[game.Player1].ManaPool.Total(); got != 2 {
		t.Fatalf("mana remaining after single-mode cast = %d, want 2", got)
	}
	stackObject, ok = g.Stack.Peek()
	if !ok || len(stackObject.ChosenModes) != 1 || stackObject.ChosenModes[0] != 1 {
		t.Fatalf("single-mode stack modes = %v, want [1]", stackObject.ChosenModes)
	}

	engine.resolveTopOfStack(g, &TurnLog{})
	if got := planeswalker.Counters.Get(counter.Loyalty); got != 2 {
		t.Fatalf("planeswalker loyalty = %d, want 2 after 3 damage", got)
	}
	if !g.Players[game.Player1].Graveyard.Contains(entwinedSpell) ||
		!g.Players[game.Player1].Graveyard.Contains(planeswalkerSpell) {
		t.Fatal("resolved Barbed Lightning spells did not go to their owner's graveyard")
	}
}
