package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// finalePumpInstruction builds the third instruction the generated Finale of
// Devastation produces: the resolving-spell continuous group effect that grants
// creatures the controller controls +X/+X (dynamic, from the spell's chosen X)
// and haste until end of turn, gated on "If X is 10 or more". It mirrors the
// generated card field-for-field so these tests exercise the exact shape.
func finalePumpInstruction() game.Instruction {
	pumpGroup := game.BattlefieldGroup(game.Selection{
		RequiredTypes: []types.Card{types.Creature},
		Controller:    game.ControllerYou,
	})
	return game.Instruction{
		Primitive: game.ApplyContinuous{
			ContinuousEffects: []game.ContinuousEffect{
				{
					Layer:                 game.LayerPowerToughnessModify,
					Group:                 pumpGroup,
					PowerDeltaDynamic:     opt.Val(game.DynamicAmount{Kind: game.DynamicAmountX, Multiplier: 1}),
					ToughnessDeltaDynamic: opt.Val(game.DynamicAmount{Kind: game.DynamicAmountX, Multiplier: 1}),
				},
				{
					Layer:       game.LayerAbility,
					Group:       pumpGroup,
					AddKeywords: []game.Keyword{game.Haste},
				},
			},
			Duration: game.DurationUntilEndOfTurn,
		},
		Condition: opt.Val(game.EffectCondition{Condition: opt.Val(game.Condition{
			Aggregates: []game.AggregateComparison{{
				Aggregate: game.AggregateSpellX,
				Op:        compare.GreaterOrEqual,
				Value:     10,
			}},
		})}),
	}
}

// addFinalePumpCreature puts a vanilla 2/2 creature onto the battlefield under
// the given controller, so pump deltas and haste grants are observable.
func addFinalePumpCreature(t *testing.T, g *game.Game, controller game.PlayerID, name string) *game.Permanent {
	t.Helper()
	return addReplacementPermanent(t, g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:      name,
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}})
}

// resolveFinalePump resolves the Finale pump rider for a Player1 spell whose
// chosen X is the given value, honoring the "If X is 10 or more" condition gate.
func resolveFinalePump(g *game.Game, x int) {
	obj := &game.StackObject{Kind: game.StackSpell, Controller: game.Player1, XValue: x}
	instr := finalePumpInstruction()
	NewEngine(nil).resolveInstructionWithChoices(g, obj, &instr, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
}

func assertFinalePT(t *testing.T, g *game.Game, permanent *game.Permanent, wantPower, wantToughness int) {
	t.Helper()
	if got := effectivePower(g, permanent); got != wantPower {
		t.Fatalf("effective power = %d, want %d", got, wantPower)
	}
	got, ok := effectiveToughness(g, permanent)
	if !ok || got != wantToughness {
		t.Fatalf("effective toughness = %d ok=%v, want %d true", got, ok, wantToughness)
	}
}

// TestFinalePumpAtXTenPumpsControllerCreaturesAndGrantsHaste proves the X=10
// rider gives creatures the resolving controller controls +X/+X and haste, while
// creatures another player controls are untouched (controller scope).
func TestFinalePumpAtXTenPumpsControllerCreaturesAndGrantsHaste(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	mine := addFinalePumpCreature(t, g, game.Player1, "My Bear")
	theirs := addFinalePumpCreature(t, g, game.Player2, "Their Bear")

	resolveFinalePump(g, 10)

	assertFinalePT(t, g, mine, 12, 12)
	if !hasKeyword(g, mine, game.Haste) {
		t.Fatal("controller's creature did not gain haste")
	}
	assertFinalePT(t, g, theirs, 2, 2)
	if hasKeyword(g, theirs, game.Haste) {
		t.Fatal("another player's creature incorrectly gained haste")
	}
}

// TestFinalePumpBelowThresholdDoesNothing proves the rider is gated on X >= 10:
// at X=9 no creature is pumped or gains haste, and no continuous effect is made.
func TestFinalePumpBelowThresholdDoesNothing(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	mine := addFinalePumpCreature(t, g, game.Player1, "My Bear")

	resolveFinalePump(g, 9)

	assertFinalePT(t, g, mine, 2, 2)
	if hasKeyword(g, mine, game.Haste) {
		t.Fatal("creature gained haste below the X=10 threshold")
	}
	if len(g.ContinuousEffects) != 0 {
		t.Fatalf("continuous effects = %d, want 0 below threshold", len(g.ContinuousEffects))
	}
}

// TestFinalePumpScalesWithX proves the pump is dynamic +X/+X from the resolving
// spell's chosen X: at X=12 a 2/2 becomes 14/14.
func TestFinalePumpScalesWithX(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	mine := addFinalePumpCreature(t, g, game.Player1, "My Bear")

	resolveFinalePump(g, 12)

	assertFinalePT(t, g, mine, 14, 14)
	if !hasKeyword(g, mine, game.Haste) {
		t.Fatal("creature did not gain haste at X=12")
	}
}

// TestFinalePumpSnapshotExcludesLaterEntrants proves the group effect snapshots
// membership at resolution (CR 611.2c): a creature that enters after resolution
// is not pumped and gains no haste, while the creature present at resolution is.
func TestFinalePumpSnapshotExcludesLaterEntrants(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	present := addFinalePumpCreature(t, g, game.Player1, "Present Bear")

	resolveFinalePump(g, 10)

	later := addFinalePumpCreature(t, g, game.Player1, "Later Bear")

	assertFinalePT(t, g, present, 12, 12)
	if !hasKeyword(g, present, game.Haste) {
		t.Fatal("creature present at resolution did not gain haste")
	}
	assertFinalePT(t, g, later, 2, 2)
	if hasKeyword(g, later, game.Haste) {
		t.Fatal("creature entering after resolution incorrectly gained haste")
	}
}

// TestFinalePumpIncludesTokens proves the group effect pumps token creatures the
// controller controls just like nontoken creatures.
func TestFinalePumpIncludesTokens(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	token := addTokenCreaturePermanent(g, game.Player1, "Servo")

	resolveFinalePump(g, 10)

	assertFinalePT(t, g, token, 12, 12)
	if !hasKeyword(g, token, game.Haste) {
		t.Fatal("token creature did not gain haste")
	}
}

// TestFinalePumpExpiresAtEndOfTurn proves both the +X/+X and haste grants last
// only until end of turn: after the cleanup step they are gone.
func TestFinalePumpExpiresAtEndOfTurn(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	mine := addFinalePumpCreature(t, g, game.Player1, "My Bear")

	resolveFinalePump(g, 10)
	assertFinalePT(t, g, mine, 12, 12)
	if !hasKeyword(g, mine, game.Haste) {
		t.Fatal("creature did not gain haste before end of turn")
	}

	expireCleanupDurations(g)

	assertFinalePT(t, g, mine, 2, 2)
	if hasKeyword(g, mine, game.Haste) {
		t.Fatal("haste persisted past end of turn")
	}
}

// TestFinalePumpSurvivesControlChange proves the snapshot binds the buff to the
// affected objects, not to their controller: a creature pumped at resolution
// keeps +X/+X and haste even after another player gains control of it.
func TestFinalePumpSurvivesControlChange(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	mine := addFinalePumpCreature(t, g, game.Player1, "My Bear")

	resolveFinalePump(g, 10)
	assertFinalePT(t, g, mine, 12, 12)

	mine.Controller = game.Player2

	assertFinalePT(t, g, mine, 12, 12)
	if !hasKeyword(g, mine, game.Haste) {
		t.Fatal("buffed creature lost haste after a control change")
	}
}

// TestFinalePumpCopiesRetainOwnX proves each resolving copy reads its own chosen
// X: resolving one copy at X=10 and another at X=12 on the same creature stacks
// +10/+10 and +12/+12 for a total of +22/+22.
func TestFinalePumpCopiesRetainOwnX(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	mine := addFinalePumpCreature(t, g, game.Player1, "My Bear")

	resolveFinalePump(g, 10)
	resolveFinalePump(g, 12)

	assertFinalePT(t, g, mine, 24, 24)
	if !hasKeyword(g, mine, game.Haste) {
		t.Fatal("creature did not have haste after two pumping copies")
	}
}

// TestFinalePumpUnsetXDoesNothing proves an unset or zero chosen X (X < 10) makes
// no group effect, so free/alternate casts that never set X do not pump.
func TestFinalePumpUnsetXDoesNothing(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	mine := addFinalePumpCreature(t, g, game.Player1, "My Bear")

	resolveFinalePump(g, 0)

	assertFinalePT(t, g, mine, 2, 2)
	if hasKeyword(g, mine, game.Haste) {
		t.Fatal("creature gained haste with an unset X")
	}
	if len(g.ContinuousEffects) != 0 {
		t.Fatalf("continuous effects = %d, want 0 with unset X", len(g.ContinuousEffects))
	}
}
