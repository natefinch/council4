package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
)

// Source review: pinned Oracle corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4;
// Out of Bounds Oracle ID 431b2e52-1290-47b1-920d-2c1817553d96, printing
// revision 8b457672-902b-42c0-9d53-a3c21be2f500 (Battlebond #38), mana cost
// {3}{U}, Instant; exact Oracle text:
// "Assist (Another player can pay up to {3} of this spell's cost.)
// Counter target spell."
// Its rulings URI
// https://api.scryfall.com/cards/8b457672-902b-42c0-9d53-a3c21be2f500/rulings
// was retrieved 2026-11-21 and returned four WotC rulings dated 2018-06-08:
// targets are chosen before the assisting player; only generic mana can be
// paid by that player; colored costs stay with the caster even when mana can be
// spent as any color; and cost increases/reductions change the amount Assist
// permits. Raw response: .cardwork/source-review/out-of-bounds-rulings.json.
// Read the run-pinned official Wizards Comprehensive Rules effective
// 2026-09-25: https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.txt,
// verified local copy .cardwork/source-review/MagicCompRules-20260925.txt
// (SHA-256 8d860e451f20f38865b725b42d82feb714c725373dd8f3b32b8652b3eeb070ca).
// Read 601.2a-i and 702.132a for spell targets, locked total costs, and the
// helper's generic payment before the caster pays the remainder; 608.2b-c and
// 701.6a-b for rechecking the spell target and moving a countered spell to its
// owner's graveyard without refunding its costs.
const outOfBoundsTestOracleID = "431b2e52-1290-47b1-920d-2c1817553d96"

func TestOutOfBoundsAssistCountersTargetSpell(t *testing.T) {
	registry := cards.NewDefaultRegistry()
	def, err := registry.LookupOracleID(outOfBoundsTestOracleID)
	if err != nil {
		t.Fatalf("exact Out of Bounds registry lookup failed: %v", err)
	}
	if def.Name != "Out of Bounds" {
		t.Fatalf("Oracle ID lookup returned %q", def.Name)
	}

	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	spell := addCardToHand(g, game.Player1, def)
	g.Players[game.Player1].ManaPool.Add(mana.U, 1)
	g.Players[game.Player2].ManaPool.Add(mana.C, 3)
	setMainPhasePriority(g, game.Player1)

	engine := NewEngineWithPrograms(nil, registry.Programs())
	ability := productionCounterSubject(g, false)
	ability.Kind = game.StackActivatedAbility
	if engine.applyAction(g, game.Player1, action.CastSpell(
		spell, []game.Target{game.StackObjectTarget(ability.ID)}, 0, nil,
	)) {
		t.Fatal("Out of Bounds illegally targeted an activated ability")
	}
	if !g.Players[game.Player1].Hand.Contains(spell) ||
		g.Players[game.Player1].ManaPool.Total() != 1 || g.Players[game.Player2].ManaPool.Total() != 3 {
		t.Fatal("illegal ability target changed Out of Bounds or either player's mana")
	}
	g.Stack.Pop()

	target := productionCounterSubject(g, false)
	if !engine.applyAction(g, game.Player1, action.CastSpell(
		spell, []game.Target{game.StackObjectTarget(target.ID)}, 0, nil,
	)) {
		t.Fatal("Out of Bounds failed to cast targeting a spell with its {U} paid by its controller and {3} paid with Assist")
	}
	if g.Players[game.Player1].ManaPool.Total() != 0 || g.Players[game.Player2].ManaPool.Total() != 0 {
		t.Fatal("Out of Bounds did not charge the assisting player for generic mana and its controller for {U}")
	}

	engine.resolveTopOfStack(g, &TurnLog{})
	if !g.Stack.IsEmpty() || !g.Players[game.Player2].Graveyard.Contains(target.SourceID) ||
		!g.Players[game.Player1].Graveyard.Contains(spell) {
		t.Fatal("Out of Bounds did not counter the target spell and finish in its owner's graveyard")
	}
}
