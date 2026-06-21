package cardgen

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
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
