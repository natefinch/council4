package rules

import (
	"testing"

	cardsg "github.com/natefinch/council4/mtg/cards/g"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TestGoreclawReducesOnlyHighPowerCreatureSpells drives the real generated
// Goreclaw, Terror of Qal Sisma card def to prove its "Creature spells you cast
// with power 4 or greater cost {2} less to cast." static: the power-filtered
// CardSelection reduces a power-4 creature spell by {2} but leaves a power-3
// creature spell untouched, confirming the new Selection.Power filter gates which
// creature spells qualify.
func TestGoreclawReducesOnlyHighPowerCreatureSpells(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCombatPermanent(g, game.Player1, cardsg.GoreclawTerrorOfQalSisma())

	bigCreature := &game.CardDef{CardFace: game.CardFace{
		Name:      "Big Creature",
		Types:     []types.Card{types.Creature},
		ManaCost:  opt.Val(cost.Mana{cost.O(6)}),
		Power:     opt.Val(game.PT{Value: 4}),
		Toughness: opt.Val(game.PT{Value: 4}),
	}}
	if got := sourceSpellGenericReduction(g, game.Player1, bigCreature); got != 2 {
		t.Fatalf("power-4 creature spell reduction = %d, want 2", got)
	}

	smallCreature := &game.CardDef{CardFace: game.CardFace{
		Name:      "Small Creature",
		Types:     []types.Card{types.Creature},
		ManaCost:  opt.Val(cost.Mana{cost.O(6)}),
		Power:     opt.Val(game.PT{Value: 3}),
		Toughness: opt.Val(game.PT{Value: 3}),
	}}
	if got := sourceSpellGenericReduction(g, game.Player1, smallCreature); got != 0 {
		t.Fatalf("power-3 creature spell reduction = %d, want 0", got)
	}
}

// TestGoreclawAttackBuffsOnlyHighPowerCreaturesYouControl exercises the runtime
// shape generated for Goreclaw's attack trigger — "each creature you control with
// power 4 or greater gets +1/+1 and gains trample until end of turn." — as an
// ApplyContinuous over a Power-filtered, controller-scoped battlefield group. The
// power-4 and power-5 creatures the resolving player controls each gain +1/+1 and
// trample, while a power-3 creature and an opponent's power-4 creature are
// unaffected.
func TestGoreclawAttackBuffsOnlyHighPowerCreaturesYouControl(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	powerFour := addCombatCreaturePermanentWithPower(g, game.Player1, 4)
	powerFive := addCombatCreaturePermanentWithPower(g, game.Player1, 5)
	powerThree := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
	opponentFour := addCombatCreaturePermanentWithPower(g, game.Player2, 4)

	obj := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackTriggeredAbility,
		Controller: game.Player1,
	}
	r := &effectResolver{engine: NewEngine(nil), game: g, obj: obj, agents: [game.NumPlayers]PlayerAgent{}, log: &TurnLog{}}

	filter := game.Selection{
		RequiredTypes: []types.Card{types.Creature},
		Controller:    game.ControllerYou,
		Power:         opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 4}),
	}
	resolved := handleApplyContinuous(r, game.ApplyContinuous{
		ContinuousEffects: []game.ContinuousEffect{
			{Layer: game.LayerPowerToughnessModify, Group: game.BattlefieldGroup(filter), PowerDelta: 1, ToughnessDelta: 1},
			{Layer: game.LayerAbility, Group: game.BattlefieldGroup(filter), AddKeywords: []game.Keyword{game.Trample}},
		},
		Duration: game.DurationUntilEndOfTurn,
	})
	if !resolved.succeeded {
		t.Fatal("Goreclaw attack ApplyContinuous did not apply")
	}

	for _, tc := range []struct {
		name     string
		creature *game.Permanent
		wantPow  int
		wantTou  int
	}{
		{"power 4", powerFour, 5, 5},
		{"power 5", powerFive, 6, 6},
	} {
		if got := effectivePower(g, tc.creature); got != tc.wantPow {
			t.Fatalf("%s creature effective power = %d, want %d", tc.name, got, tc.wantPow)
		}
		if got, ok := effectiveToughness(g, tc.creature); !ok || got != tc.wantTou {
			t.Fatalf("%s creature effective toughness = %d (ok=%v), want %d", tc.name, got, ok, tc.wantTou)
		}
		if !hasKeyword(g, tc.creature, game.Trample) {
			t.Fatalf("%s creature you control did not gain trample", tc.name)
		}
	}

	if got := effectivePower(g, powerThree); got != 3 {
		t.Fatalf("power-3 creature effective power = %d, want 3 (unbuffed)", got)
	}
	if hasKeyword(g, powerThree, game.Trample) {
		t.Fatal("power-3 creature wrongly gained trample")
	}
	if got := effectivePower(g, opponentFour); got != 4 {
		t.Fatalf("opponent power-4 creature effective power = %d, want 4 (unbuffed)", got)
	}
	if hasKeyword(g, opponentFour, game.Trample) {
		t.Fatal("opponent's power-4 creature wrongly gained trample")
	}
}
