package cardgen

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestLowerTargetDoublePowerUntilEndOfTurn(t *testing.T) {
	t.Parallel()
	tests := []string{
		"Double target creature's power until end of turn.",
		"Double the power of target creature until end of turn.",
	}
	for _, oracle := range tests {
		t.Run(oracle, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Double",
				Layout:     "normal",
				TypeLine:   "Instant",
				ManaCost:   "{R}",
				OracleText: oracle,
			})
			mode := face.SpellAbility.Val.Modes[0]
			if len(mode.Targets) != 1 {
				t.Fatalf("targets = %d, want 1", len(mode.Targets))
			}
			if len(mode.Sequence) != 1 {
				t.Fatalf("sequence = %#v, want 1 instruction", mode.Sequence)
			}
			modify, ok := mode.Sequence[0].Primitive.(game.ModifyPT)
			if !ok {
				t.Fatalf("primitive = %T, want game.ModifyPT", mode.Sequence[0].Primitive)
			}
			if modify.Object != game.TargetPermanentReference(0) {
				t.Fatalf("object = %#v, want target 0", modify.Object)
			}
			if modify.Duration != game.DurationUntilEndOfTurn {
				t.Fatalf("duration = %v, want until end of turn", modify.Duration)
			}
			power := modify.PowerDelta.DynamicAmount()
			if !power.Exists ||
				power.Val.Kind != game.DynamicAmountObjectPower ||
				power.Val.Object != game.TargetPermanentReference(0) {
				t.Fatalf("power delta = %#v, want target's power at resolution", modify.PowerDelta)
			}
		})
	}
}

func TestLowerTargetDoublePowerAndKeywordUntilEndOfTurn(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Legion Leadership",
		Layout:     "normal",
		TypeLine:   "Instant",
		ManaCost:   "{1}{R/W}",
		OracleText: "Until end of turn, double target creature's power and it gains first strike.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 {
		t.Fatalf("targets = %d, want 1", len(mode.Targets))
	}
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %#v, want modify and keyword instructions", mode.Sequence)
	}
	modify, ok := mode.Sequence[0].Primitive.(game.ModifyPT)
	if !ok {
		t.Fatalf("primitive[0] = %T, want game.ModifyPT", mode.Sequence[0].Primitive)
	}
	if modify.Object != game.TargetPermanentReference(0) ||
		modify.Duration != game.DurationUntilEndOfTurn {
		t.Fatalf("modify = %#v, want target 0 until end of turn", modify)
	}
	power := modify.PowerDelta.DynamicAmount()
	if !power.Exists ||
		power.Val.Kind != game.DynamicAmountObjectPower ||
		power.Val.Object != game.TargetPermanentReference(0) {
		t.Fatalf("power delta = %#v, want target's power at resolution", modify.PowerDelta)
	}
	apply, ok := mode.Sequence[1].Primitive.(game.ApplyContinuous)
	if !ok {
		t.Fatalf("primitive[1] = %T, want game.ApplyContinuous", mode.Sequence[1].Primitive)
	}
	if apply.Object != opt.Val(game.TargetPermanentReference(0)) {
		t.Fatalf("object = %#v, want target 0", apply.Object)
	}
	if apply.Duration != game.DurationUntilEndOfTurn {
		t.Fatalf("duration = %v, want until end of turn", apply.Duration)
	}
	if len(apply.ContinuousEffects) != 1 {
		t.Fatalf("continuous effects = %#v, want keyword only", apply.ContinuousEffects)
	}
	keywords := apply.ContinuousEffects[0]
	if keywords.Layer != game.LayerAbility || len(keywords.AddKeywords) != 1 || keywords.AddKeywords[0] != game.FirstStrike {
		t.Fatalf("keyword effect = %#v, want first strike", keywords)
	}
}

func TestLowerSourceDoublePowerAndToughnessUntilEndOfTurn(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Reckless Amplimancer",
		Layout:     "normal",
		TypeLine:   "Creature — Elf Druid",
		ManaCost:   "{1}{G}",
		OracleText: "{4}{G}: Double this creature's power and toughness until end of turn.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("activated abilities = %d, want 1", len(face.ActivatedAbilities))
	}
	mode := face.ActivatedAbilities[0].Content.Modes[0]
	if len(mode.Sequence) != 1 {
		t.Fatalf("sequence = %#v, want 1 instruction", mode.Sequence)
	}
	modify, ok := mode.Sequence[0].Primitive.(game.ModifyPT)
	if !ok {
		t.Fatalf("primitive = %T, want game.ModifyPT", mode.Sequence[0].Primitive)
	}
	if modify.Object != game.SourcePermanentReference() {
		t.Fatalf("object = %#v, want source permanent", modify.Object)
	}
	if modify.Duration != game.DurationUntilEndOfTurn {
		t.Fatalf("duration = %v, want until end of turn", modify.Duration)
	}
	for _, quantity := range []game.Quantity{modify.PowerDelta, modify.ToughnessDelta} {
		dynamic := quantity.DynamicAmount()
		if !dynamic.Exists ||
			dynamic.Val.Object != game.SourcePermanentReference() {
			t.Fatalf("delta = %#v, want source characteristic at resolution", quantity)
		}
	}
}

func TestLowerSubtypeGroupDoublePowerUntilEndOfTurn(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Thrakkus the Butcher",
		Layout:     "normal",
		TypeLine:   "Legendary Creature — Dragon Peasant",
		ManaCost:   "{3}{R}{G}",
		OracleText: "Trample\nWhenever Thrakkus attacks, double the power of each Dragon you control until end of turn.",
		Power:      new("3"),
		Toughness:  new("4"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	mode := face.TriggeredAbilities[0].Content.Modes[0]
	if len(mode.Sequence) != 1 {
		t.Fatalf("sequence = %#v, want 1 instruction", mode.Sequence)
	}
	apply, ok := mode.Sequence[0].Primitive.(game.ApplyContinuous)
	if !ok {
		t.Fatalf("primitive = %T, want game.ApplyContinuous", mode.Sequence[0].Primitive)
	}
	if apply.Object.Exists || apply.Duration != game.DurationUntilEndOfTurn {
		t.Fatalf("apply = %#v, want group effect until end of turn", apply)
	}
	if len(apply.ContinuousEffects) != 1 {
		t.Fatalf("continuous effects = %#v, want one", apply.ContinuousEffects)
	}
	effect := apply.ContinuousEffects[0]
	if effect.Layer != game.LayerPowerToughnessModify || !effect.DoublePower || effect.DoubleToughness {
		t.Fatalf("continuous effect = %#v, want double power", effect)
	}
	selection := effect.Group.Selection()
	if selection.Controller != game.ControllerYou ||
		len(selection.SubtypesAny) != 1 || selection.SubtypesAny[0] != types.Dragon {
		t.Fatalf("selection = %#v, want controlled Dragons", selection)
	}
}

func TestUnsupportedRepeatedDoublePowerDoesNotPanic(t *testing.T) {
	t.Parallel()
	face := lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:       "Exponential Growth",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		ManaCost:   "{X}{X}{G}{G}",
		OracleText: "Until end of turn, double target creature's power X times.",
	})
	if face.SpellAbility.Exists {
		t.Fatalf("spell ability = %#v, want unsupported repeated doubling to fail closed", face.SpellAbility)
	}
}

// TestRenderDoublePowerToughnessFields guards against the regression where the
// renderer dropped the dynamic object-power/object-toughness amounts, silently
// emitting a no-op power/toughness modify. It asserts the generated source reads
// the affected object's current values at resolution.
func TestRenderDoublePowerToughnessFields(t *testing.T) {
	t.Parallel()
	tests := []struct {
		oracle string
		want   []string
	}{
		{"Double target creature's power until end of turn.", []string{"game.DynamicAmountObjectPower"}},
		{"Double target creature's power and toughness until end of turn.", []string{"game.DynamicAmountObjectPower", "game.DynamicAmountObjectToughness"}},
	}
	for _, test := range tests {
		t.Run(test.oracle, func(t *testing.T) {
			t.Parallel()
			source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
				Name:       "Test Double Render",
				Layout:     "normal",
				TypeLine:   "Instant",
				ManaCost:   "{R}",
				OracleText: test.oracle,
			}, "t")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			for _, want := range test.want {
				if !strings.Contains(source, want) {
					t.Fatalf("source missing %q:\n%s", want, source)
				}
			}
		})
	}
}

// TestGenerateJunkJetEquippedSubject covers the resolving "Double equipped
// creature's power" form (issue #2648): the "equipped creature's" possessive
// subject lowers to the source's attached permanent, doubling its power.
func TestGenerateJunkJetEquippedSubject(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Junk Jet",
		Layout:     "normal",
		TypeLine:   "Artifact — Equipment",
		ManaCost:   "{2}",
		OracleText: "{3}, Sacrifice another artifact: Double equipped creature's power until end of turn.\nEquip {1}",
	}, "j")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"game.SourceAttachedPermanentReference()",
		"game.DynamicAmountObjectPower",
		"Primitive: game.ModifyPT{",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("source missing %q:\n%s", want, source)
		}
	}
}
