package cardgen

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/types"
)

// dominateCard is the Scryfall shape of Dominate, whose gain-control target is
// bounded by the spell's chosen X ("target creature with mana value X or less").
func dominateCard() *ScryfallCard {
	return &ScryfallCard{
		Name:       "Dominate",
		Layout:     "normal",
		ManaCost:   "{X}{1}{U}{U}",
		TypeLine:   "Instant",
		OracleText: "Gain control of target creature with mana value X or less.",
		Colors:     []string{"U"},
	}
}

// TestLowerDominateManaValueAtMostX proves the X-bounded mana value target filter
// lowers to a TargetSpec.ManaValueAtMostX flag (not a fixed Selection.ManaValue
// bound) alongside a permanent gain-control ApplyContinuous.
func TestLowerDominateManaValueAtMostX(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, dominateCard())
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 {
		t.Fatalf("targets = %d, want 1", len(mode.Targets))
	}
	spec := mode.Targets[0]
	if !spec.ManaValueAtMostX {
		t.Fatal("target spec must set ManaValueAtMostX for \"mana value X or less\"")
	}
	if spec.Allow != game.TargetAllowPermanent {
		t.Fatalf("target allow = %v, want TargetAllowPermanent", spec.Allow)
	}
	if !spec.Selection.Exists || spec.Selection.Val.ManaValue.Exists {
		t.Fatalf("selection must carry no fixed mana value bound, got %+v", spec.Selection)
	}
	if len(spec.Selection.Val.RequiredTypesAny) != 1 || spec.Selection.Val.RequiredTypesAny[0] != types.Creature {
		t.Fatalf("selection types = %v, want [Creature]", spec.Selection.Val.RequiredTypesAny)
	}
	prim, ok := mode.Sequence[0].Primitive.(game.ApplyContinuous)
	if !ok {
		t.Fatalf("sequence[0] = %T, want game.ApplyContinuous", mode.Sequence[0].Primitive)
	}
	if len(prim.ContinuousEffects) != 1 || prim.ContinuousEffects[0].Layer != game.LayerControl {
		t.Fatalf("continuous effects = %+v, want one LayerControl effect", prim.ContinuousEffects)
	}
	if got := prim.ContinuousEffects[0].NewController; !got.Exists || got.Val != game.Player1 {
		t.Fatalf("NewController = %v, want Player1 (controller gains control)", got)
	}
	if prim.Duration != game.DurationPermanent {
		t.Fatalf("duration = %v, want DurationPermanent", prim.Duration)
	}
}

// TestGenerateDominateSource proves Dominate generates executable source with no
// diagnostics and carries the ManaValueAtMostX flag.
func TestGenerateDominateSource(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(dominateCard(), "d")
	if err != nil {
		t.Fatalf("GenerateExecutableCardSource error = %v", err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v, want none", diagnostics)
	}
	if !strings.Contains(source, "ManaValueAtMostX: true") {
		t.Fatalf("source missing ManaValueAtMostX flag:\n%s", source)
	}
}

// blueSunsTwilightCard is the Scryfall shape of Blue Sun's Twilight, whose
// gain-control target is X-mana-value-bounded and whose "If X is 5 or more,
// create a token that's a copy of that creature." follow-on is gated on the
// spell's chosen X and copies the just-gained creature.
func blueSunsTwilightCard() *ScryfallCard {
	return &ScryfallCard{
		Name:       "Blue Sun's Twilight",
		Layout:     "normal",
		ManaCost:   "{X}{U}{U}",
		TypeLine:   "Sorcery",
		OracleText: "Gain control of target creature with mana value X or less. If X is 5 or more, create a token that's a copy of that creature.",
		Colors:     []string{"U"},
	}
}

// TestLowerBlueSunsTwilightGatedCopy proves the conditional token-copy follow-on
// in a gain-control sequence lowers correctly: the gain-control runs
// unconditionally against the X-bounded target, and the "If X is 5 or more"
// gate applies only to the CreateToken that copies that same target (slot 0).
func TestLowerBlueSunsTwilightGatedCopy(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, blueSunsTwilightCard())
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 || !mode.Targets[0].ManaValueAtMostX {
		t.Fatalf("target = %+v, want one ManaValueAtMostX permanent target", mode.Targets)
	}
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %d instructions, want 2", len(mode.Sequence))
	}

	control, ok := mode.Sequence[0].Primitive.(game.ApplyContinuous)
	if !ok {
		t.Fatalf("sequence[0] = %T, want game.ApplyContinuous", mode.Sequence[0].Primitive)
	}
	if len(control.ContinuousEffects) != 1 || control.ContinuousEffects[0].Layer != game.LayerControl {
		t.Fatalf("gain-control effects = %+v, want one LayerControl effect", control.ContinuousEffects)
	}
	if got := control.ContinuousEffects[0].NewController; !got.Exists || got.Val != game.Player1 {
		t.Fatalf("NewController = %v, want Player1", got)
	}
	if mode.Sequence[0].Condition.Exists {
		t.Fatal("gain-control instruction must be unconditional, not gated on X")
	}

	token, ok := mode.Sequence[1].Primitive.(game.CreateToken)
	if !ok {
		t.Fatalf("sequence[1] = %T, want game.CreateToken", mode.Sequence[1].Primitive)
	}
	spec, ok := token.Source.TokenCopy()
	if !ok || spec.Source != game.TokenCopySourceObject ||
		spec.Object != game.TargetPermanentReference(0) {
		t.Fatalf("token copy source = %+v (ok=%v), want copy of target permanent slot 0", spec, ok)
	}
	gate := mode.Sequence[1].Condition
	if !gate.Exists || !gate.Val.Condition.Exists {
		t.Fatal("token-copy instruction must be gated on the X threshold")
	}
	if got := gate.Val.Condition.Val.Aggregates; len(got) != 1 ||
		got[0].Aggregate != game.AggregateSpellX ||
		got[0].Op != compare.GreaterOrEqual ||
		got[0].Value != 5 {
		t.Fatalf("gate aggregate = %+v, want spell-X >= 5", gate.Val.Condition.Val.Aggregates)
	}
}

// TestGenerateBlueSunsTwilightSource proves Blue Sun's Twilight generates
// executable source with no diagnostics now that the X-gated copy-of-gained-
// creature follow-on is supported.
func TestGenerateBlueSunsTwilightSource(t *testing.T) {
	t.Parallel()
	_, diagnostics, err := GenerateExecutableCardSource(blueSunsTwilightCard(), "b")
	if err != nil {
		t.Fatalf("GenerateExecutableCardSource error = %v", err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v, want none", diagnostics)
	}
}
