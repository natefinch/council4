package cardgen

import (
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
			apply, ok := mode.Sequence[0].Primitive.(game.ApplyContinuous)
			if !ok {
				t.Fatalf("primitive = %T, want game.ApplyContinuous", mode.Sequence[0].Primitive)
			}
			if apply.Object != opt.Val(game.TargetPermanentReference(0)) {
				t.Fatalf("object = %#v, want target 0", apply.Object)
			}
			if apply.Duration != game.DurationUntilEndOfTurn {
				t.Fatalf("duration = %v, want until end of turn", apply.Duration)
			}
			if len(apply.ContinuousEffects) != 1 {
				t.Fatalf("continuous effects = %#v, want one", apply.ContinuousEffects)
			}
			effect := apply.ContinuousEffects[0]
			if effect.Layer != game.LayerPowerToughnessModify || !effect.DoublePower || effect.DoubleToughness {
				t.Fatalf("continuous effect = %#v, want double power only", effect)
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
	if len(mode.Sequence) != 1 {
		t.Fatalf("sequence = %#v, want 1 instruction", mode.Sequence)
	}
	apply, ok := mode.Sequence[0].Primitive.(game.ApplyContinuous)
	if !ok {
		t.Fatalf("primitive = %T, want game.ApplyContinuous", mode.Sequence[0].Primitive)
	}
	if apply.Object != opt.Val(game.TargetPermanentReference(0)) {
		t.Fatalf("object = %#v, want target 0", apply.Object)
	}
	if apply.Duration != game.DurationUntilEndOfTurn {
		t.Fatalf("duration = %v, want until end of turn", apply.Duration)
	}
	if len(apply.ContinuousEffects) != 2 {
		t.Fatalf("continuous effects = %#v, want PT + keyword", apply.ContinuousEffects)
	}
	pt := apply.ContinuousEffects[0]
	if pt.Layer != game.LayerPowerToughnessModify || !pt.DoublePower || pt.DoubleToughness {
		t.Fatalf("PT effect = %#v, want double power only", pt)
	}
	keywords := apply.ContinuousEffects[1]
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
	apply, ok := mode.Sequence[0].Primitive.(game.ApplyContinuous)
	if !ok {
		t.Fatalf("primitive = %T, want game.ApplyContinuous", mode.Sequence[0].Primitive)
	}
	if apply.Object != opt.Val(game.SourcePermanentReference()) {
		t.Fatalf("object = %#v, want source permanent", apply.Object)
	}
	if apply.Duration != game.DurationUntilEndOfTurn {
		t.Fatalf("duration = %v, want until end of turn", apply.Duration)
	}
	if len(apply.ContinuousEffects) != 1 {
		t.Fatalf("continuous effects = %#v, want one", apply.ContinuousEffects)
	}
	effect := apply.ContinuousEffects[0]
	if effect.Layer != game.LayerPowerToughnessModify || !effect.DoublePower || !effect.DoubleToughness {
		t.Fatalf("continuous effect = %#v, want double power/toughness", effect)
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
