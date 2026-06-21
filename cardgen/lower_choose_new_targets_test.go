package cardgen

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestLowerChooseNewTargetsSpell(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		oracleText   string
		wantKinds    []game.StackObjectKind
		wantOptional bool
	}{
		{
			name:         "target spell",
			oracleText:   "You may choose new targets for target spell.",
			wantKinds:    []game.StackObjectKind{game.StackSpell},
			wantOptional: true,
		},
		{
			name:       "target spell or ability",
			oracleText: "You may choose new targets for target spell or ability.",
			wantKinds: []game.StackObjectKind{
				game.StackSpell, game.StackActivatedAbility, game.StackTriggeredAbility,
			},
			wantOptional: true,
		},
		{
			name:         "change the target of target spell",
			oracleText:   "Change the target of target spell with a single target.",
			wantKinds:    []game.StackObjectKind{game.StackSpell},
			wantOptional: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Retarget",
				Layout:     "normal",
				TypeLine:   "Instant",
				OracleText: test.oracleText,
			})
			if !face.SpellAbility.Exists {
				t.Fatal("spell ability missing")
			}
			ability := face.SpellAbility.Val
			if len(ability.Modes) != 1 {
				t.Fatalf("modes = %d, want 1", len(ability.Modes))
			}
			mode := ability.Modes[0]
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
			if len(mode.Sequence) != 1 {
				t.Fatalf("sequence = %d, want 1", len(mode.Sequence))
			}
			instr := mode.Sequence[0]
			if instr.Optional != test.wantOptional {
				t.Fatalf("instruction Optional = %v, want %v", instr.Optional, test.wantOptional)
			}
			retarget, ok := instr.Primitive.(game.ChooseNewTargets)
			if !ok {
				t.Fatalf("primitive = %T, want game.ChooseNewTargets", instr.Primitive)
			}
			if retarget.Object.Kind() != game.ObjectReferenceTargetStackObject || retarget.Object.TargetIndex() != 0 {
				t.Fatalf("retarget object = %+v, want target stack object 0", retarget.Object)
			}
		})
	}
}
