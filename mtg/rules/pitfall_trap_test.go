package rules

import (
	"testing"

	cardsp "github.com/natefinch/council4/mtg/cards/p"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TestPitfallTrapManaAlternativeCost exercises the generated Pitfall Trap
// end-to-end: "If exactly one creature is attacking, you may pay {W} rather
// than pay this spell's mana cost." Its normal {2}{W} cost is unpayable from a
// single white mana, so the cast is legal only via the {W} alternative, which is
// gated on exactly one creature attacking. With one attacker the cast pays {W}
// and, on resolution, destroys the attacking creature.
func TestPitfallTrapManaAlternativeCost(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, cardsp.PitfallTrap())
	attacker := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:      "Grizzly Bears",
		Types:     []types.Card{types.Creature},
		Colors:    []color.Color{color.Green},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}})
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareBlockers
	g.Players[game.Player1].ManaPool.Add(mana.W, 1)

	act := action.CastSpell(spellID, []game.Target{game.PermanentTarget(attacker.ObjectID)}, 0, nil)

	// With no attacking creature the exactly-one-attacker condition is false, so
	// the {W} alternative is not offered and the only mana available ({W}) cannot
	// pay the {2}{W} normal cost.
	if containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("Pitfall Trap cast was legal with no creature attacking")
	}

	g.Combat = &game.CombatState{Attackers: []game.AttackDeclaration{
		{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player1}},
	}}
	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("Pitfall Trap cast was not legal with exactly one creature attacking")
	}

	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(cast Pitfall Trap via pay {W}) = false, want true")
	}
	if got := g.Players[game.Player1].ManaPool.Total(); got != 0 {
		t.Fatalf("mana remaining = %d, want 0 after paying the {W} alternative", got)
	}
	if _, ok := g.Stack.Peek(); !ok {
		t.Fatal("Pitfall Trap was not put on the stack")
	}

	engine.resolveTopOfStack(g, &TurnLog{})

	if _, ok := permanentByObjectID(g, attacker.ObjectID); ok {
		t.Fatal("attacking creature was not destroyed")
	}
}
