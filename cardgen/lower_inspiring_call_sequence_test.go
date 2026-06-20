package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
)

const inspiringCallOracleText = "Draw a card for each creature you control with a +1/+1 counter on it. " +
	"Those creatures gain indestructible until end of turn. " +
	"(Damage and effects that say \"destroy\" don't destroy them.)"

// TestLowerInspiringCallSequence verifies the ordered pair "Draw a card for each
// <group>. Those creatures gain <keyword> until end of turn." lowers to a
// dynamic-count draw followed by a keyword grant over that same counted group.
func TestLowerInspiringCallSequence(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Inspiring Call",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: inspiringCallOracleText,
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %#v, want draw then keyword grant", mode.Sequence)
	}
	draw, ok := mode.Sequence[0].Primitive.(game.Draw)
	if !ok ||
		draw.Player.Kind() != game.PlayerReferenceController ||
		!draw.Amount.DynamicAmount().Exists ||
		draw.Amount.DynamicAmount().Val.Kind != game.DynamicAmountCountSelector {
		t.Fatalf("draw = %#v, want controller draw for each counted creature", mode.Sequence[0].Primitive)
	}
	apply, ok := mode.Sequence[1].Primitive.(game.ApplyContinuous)
	if !ok || apply.Object.Exists || apply.Duration != game.DurationUntilEndOfTurn {
		t.Fatalf("apply = %#v, want unanchored group grant until end of turn", mode.Sequence[1].Primitive)
	}
	if len(apply.ContinuousEffects) != 1 {
		t.Fatalf("continuous effects = %d, want 1", len(apply.ContinuousEffects))
	}
	effect := apply.ContinuousEffects[0]
	if effect.Layer != game.LayerAbility ||
		len(effect.AddKeywords) != 1 ||
		effect.AddKeywords[0] != game.Indestructible {
		t.Fatalf("effect = %+v, want indestructible keyword layer", effect)
	}
	// The grant's group must be exactly the draw's counted selection so "those
	// creatures" resolves to the just-counted set.
	wantSelection := draw.Amount.DynamicAmount().Val.Group.Selection()
	selection := effect.Group.Selection()
	if effect.Group.Domain() != game.GroupDomainBattlefield ||
		selection.Controller != game.ControllerYou ||
		len(selection.RequiredTypes) != 1 ||
		selection.RequiredTypes[0] != types.Creature ||
		!selection.MatchCounter ||
		selection.RequiredCounter != counter.PlusOnePlusOne ||
		selection.RequiredCounter != wantSelection.RequiredCounter {
		t.Fatalf("group selection = %+v, want creatures you control with a +1/+1 counter", selection)
	}
}

// TestLowerInspiringCallVariantKeywords verifies the back-reference grant
// generalizes over keyword and dynamic-count group beyond Inspiring Call's exact
// wording.
func TestLowerInspiringCallVariantKeywords(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Those Grant",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Draw a card for each creature you control with a +1/+1 counter on it. Those creatures gain hexproof until end of turn.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	apply, ok := mode.Sequence[1].Primitive.(game.ApplyContinuous)
	if !ok || len(apply.ContinuousEffects) != 1 {
		t.Fatalf("apply = %#v, want one group grant", mode.Sequence[1].Primitive)
	}
	effect := apply.ContinuousEffects[0]
	if len(effect.AddKeywords) != 1 || effect.AddKeywords[0] != game.Hexproof {
		t.Fatalf("keywords = %v, want [Hexproof]", effect.AddKeywords)
	}
}
