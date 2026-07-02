package cardgen

import (
	"reflect"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestLowerControlledLandsSkipUntapSpell verifies the mass self-stun "Lands you
// control don't untap during your next untap step." lowers to a single group
// SkipNextUntap over the controller's lands, with no target.
func TestLowerControlledLandsSkipUntapSpell(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Lands Stall",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Lands you control don't untap during your next untap step.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 0 || len(mode.Sequence) != 1 {
		t.Fatalf("mode = %+v, want no targets and one instruction", mode)
	}
	stun, ok := mode.Sequence[0].Primitive.(game.SkipNextUntap)
	if !ok {
		t.Fatalf("primitive = %T, want game.SkipNextUntap", mode.Sequence[0].Primitive)
	}
	want := game.BattlefieldGroup(game.Selection{
		RequiredTypes: []types.Card{types.Land},
		Controller:    game.ControllerYou,
	})
	if stun.Object != (game.ObjectReference{}) {
		t.Fatalf("group skip-untap carried an object reference %+v, want the group form", stun.Object)
	}
	if !reflect.DeepEqual(stun.Group, want) {
		t.Fatalf("group = %+v, want lands you control %+v", stun.Group, want)
	}
}

// TestLowerControlledCreaturesSkipUntapSpell verifies the creatures wording lowers
// to the creatures-you-control group, exercising the generalization beyond the
// original hardcoded lands-only recognizer.
func TestLowerControlledCreaturesSkipUntapSpell(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Creatures Stall",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Creatures you control don't untap during your next untap step.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	stun, ok := face.SpellAbility.Val.Modes[0].Sequence[0].Primitive.(game.SkipNextUntap)
	if !ok {
		t.Fatalf("primitive = %T, want game.SkipNextUntap", face.SpellAbility.Val.Modes[0].Sequence[0].Primitive)
	}
	want := game.BattlefieldGroup(game.Selection{
		RequiredTypes: []types.Card{types.Creature},
		Controller:    game.ControllerYou,
	})
	if !reflect.DeepEqual(stun.Group, want) {
		t.Fatalf("group = %+v, want creatures you control %+v", stun.Group, want)
	}
}

// TestLowerGroupSkipUntapInSequence verifies the mass self-stun composes as the
// trailing clause of an ordered sequence ("Destroy all creatures. Lands you
// control don't untap during your next untap step." — Bontu's Last Reckoning): the
// whole card lowers, ending in the group SkipNextUntap.
func TestLowerGroupSkipUntapInSequence(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Reckoning",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Destroy all creatures. Lands you control don't untap during your next untap step.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	sequence := face.SpellAbility.Val.Modes[0].Sequence
	last := sequence[len(sequence)-1].Primitive
	if _, ok := last.(game.SkipNextUntap); !ok {
		t.Fatalf("last instruction = %T, want game.SkipNextUntap", last)
	}
}

// TestLowerGroupSkipUntapFailsClosed verifies shapes the self-controlled group
// skip-untap cannot yet model stay unsupported: a targeted player's permanents
// ("Creatures target player controls ...") whose controller is not the resolving
// player has no group representation here.
func TestLowerGroupSkipUntapFailsClosed(t *testing.T) {
	t.Parallel()
	_, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Test Targeted Stall",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Creatures target player controls don't untap during that player's next untap step.",
	})
	if len(diagnostics) == 0 {
		t.Fatal("expected a diagnostic for the unsupported targeted-player mass stun")
	}
}
