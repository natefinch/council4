package cardgen

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestLowerCopyStackObjectSpell(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		oracle    string
		wantKinds []game.StackObjectKind
		wantMay   bool
	}{
		{
			name:      "triggered with new targets",
			oracle:    "{2}, {T}: Copy target triggered ability you control. You may choose new targets for the copy.",
			wantKinds: []game.StackObjectKind{game.StackTriggeredAbility},
			wantMay:   true,
		},
		{
			name:      "activated ability without new targets",
			oracle:    "{2}, {T}: Copy target activated ability you control.",
			wantKinds: []game.StackObjectKind{game.StackActivatedAbility},
			wantMay:   false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Resonator",
				Layout:     "normal",
				TypeLine:   "Artifact",
				OracleText: test.oracle,
			})
			if len(face.ActivatedAbilities) != 1 {
				t.Fatalf("activated abilities = %d, want 1", len(face.ActivatedAbilities))
			}
			content := face.ActivatedAbilities[0].Content
			if len(content.Modes) != 1 {
				t.Fatalf("modes = %d, want 1", len(content.Modes))
			}
			mode := content.Modes[0]
			if len(mode.Targets) != 1 {
				t.Fatalf("targets = %d, want 1", len(mode.Targets))
			}
			target := mode.Targets[0]
			if target.Allow != game.TargetAllowStackObject {
				t.Fatalf("target allow = %v, want stack object", target.Allow)
			}
			if !slices.Equal(target.Predicate.StackObjectKinds, test.wantKinds) {
				t.Fatalf("stack object kinds = %+v, want %+v", target.Predicate.StackObjectKinds, test.wantKinds)
			}
			if target.Predicate.Controller != game.ControllerYou {
				t.Fatalf("controller = %v, want you", target.Predicate.Controller)
			}
			if len(mode.Sequence) != 1 {
				t.Fatalf("sequence = %d, want 1", len(mode.Sequence))
			}
			copyPrim, ok := mode.Sequence[0].Primitive.(game.CopyStackObject)
			if !ok {
				t.Fatalf("primitive = %T, want game.CopyStackObject", mode.Sequence[0].Primitive)
			}
			if copyPrim.Object.Kind() != game.ObjectReferenceTargetStackObject || copyPrim.Object.TargetIndex() != 0 {
				t.Fatalf("copy object = %+v, want target stack object 0", copyPrim.Object)
			}
			if copyPrim.MayChooseNewTargets != test.wantMay {
				t.Fatalf("MayChooseNewTargets = %v, want %v", copyPrim.MayChooseNewTargets, test.wantMay)
			}
		})
	}
}

// TestLowerCopySpellAbility verifies a spell that copies a target instant or
// sorcery spell (Twincast, Reverberate) lowers to a CopyStackObject over a
// spell-only stack-object target, with the "you may choose new targets" rider.
func TestLowerCopySpellAbility(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Twincast",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Copy target instant or sorcery spell. You may choose new targets for the copy.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 || mode.Targets[0].Allow != game.TargetAllowStackObject {
		t.Fatalf("targets = %+v, want one stack-object target", mode.Targets)
	}
	if !slices.Equal(mode.Targets[0].Predicate.StackObjectKinds, []game.StackObjectKind{game.StackSpell}) {
		t.Fatalf("stack object kinds = %+v, want spell only", mode.Targets[0].Predicate.StackObjectKinds)
	}
	if !slices.Equal(mode.Targets[0].Predicate.SpellCardTypesAny, []types.Card{types.Instant, types.Sorcery}) {
		t.Fatalf("spell card types = %+v, want instant or sorcery", mode.Targets[0].Predicate.SpellCardTypesAny)
	}
	copyPrim, ok := mode.Sequence[0].Primitive.(game.CopyStackObject)
	if !ok {
		t.Fatalf("primitive = %T, want game.CopyStackObject", mode.Sequence[0].Primitive)
	}
	if !copyPrim.MayChooseNewTargets {
		t.Fatal("MayChooseNewTargets = false, want true")
	}
}

// TestLowerCopySpellTriggeredAbility verifies Dualcaster Mage's enters-the-
// battlefield triggered ability — "copy target instant or sorcery spell. You
// may choose new targets for the copy." — lowers cleanly. The "you may choose
// new targets for the copy" rider sentence must be span-credited in the
// triggered-ability path or the whole card fails the completeness check.
func TestLowerCopySpellTriggeredAbility(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Dualcaster",
		Layout:     "normal",
		TypeLine:   "Creature — Human Wizard",
		OracleText: "Flash\nWhen Test Dualcaster enters, copy target instant or sorcery spell. You may choose new targets for the copy.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	mode := face.TriggeredAbilities[0].Content.Modes[0]
	if len(mode.Targets) != 1 || mode.Targets[0].Allow != game.TargetAllowStackObject {
		t.Fatalf("targets = %+v, want one stack-object target", mode.Targets)
	}
	if !slices.Equal(mode.Targets[0].Predicate.StackObjectKinds, []game.StackObjectKind{game.StackSpell}) {
		t.Fatalf("stack object kinds = %+v, want spell only", mode.Targets[0].Predicate.StackObjectKinds)
	}
	copyPrim, ok := mode.Sequence[0].Primitive.(game.CopyStackObject)
	if !ok {
		t.Fatalf("primitive = %T, want game.CopyStackObject", mode.Sequence[0].Primitive)
	}
	if !copyPrim.MayChooseNewTargets {
		t.Fatal("MayChooseNewTargets = false, want true")
	}
}

// TestLowerCopyTriggeringSpell verifies that "Whenever you cast a [filter]
// spell, copy that spell." (Reflections of Littjara / Veyran-style) lowers the
// copy clause to a CopyStackObject over the triggering spell — an
// EventStackObject reference with no targets, rather than a target stack object.
func TestLowerCopyTriggeringSpell(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		oracle string
	}{
		{
			name:   "instant or sorcery filter",
			oracle: "Whenever you cast an instant or sorcery spell, copy that spell.",
		},
		{
			name:   "chosen type filter",
			oracle: "As Test Reflections enters, choose a creature type.\nWhenever you cast a spell of the chosen type, copy that spell.",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Reflections " + test.name,
				Layout:     "normal",
				TypeLine:   "Enchantment",
				OracleText: test.oracle,
			})
			var trigger game.AbilityContent
			for _, ability := range face.TriggeredAbilities {
				if len(ability.Content.Modes) == 1 && len(ability.Content.Modes[0].Sequence) == 1 {
					if _, ok := ability.Content.Modes[0].Sequence[0].Primitive.(game.CopyStackObject); ok {
						trigger = ability.Content
					}
				}
			}
			if len(trigger.Modes) != 1 {
				t.Fatalf("no copy triggered ability lowered from %q", test.oracle)
			}
			mode := trigger.Modes[0]
			if len(mode.Targets) != 0 {
				t.Fatalf("targets = %+v, want none (copies the triggering spell)", mode.Targets)
			}
			copyPrim, ok := mode.Sequence[0].Primitive.(game.CopyStackObject)
			if !ok {
				t.Fatalf("primitive = %T, want game.CopyStackObject", mode.Sequence[0].Primitive)
			}
			if copyPrim.Object.Kind() != game.ObjectReferenceEventStackObject {
				t.Fatalf("copy object = %+v, want event stack object", copyPrim.Object)
			}
		})
	}
}
