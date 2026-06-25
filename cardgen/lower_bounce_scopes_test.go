package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// bounceInstructions returns the game.Bounce primitives emitted in the first mode
// of a lowered spell ability, in order, so a scope test can assert how many
// permanents/objects the return moves and how each is addressed.
func bounceInstructions(t *testing.T, face loweredFaceAbilities) []game.Bounce {
	t.Helper()
	if !face.SpellAbility.Exists || len(face.SpellAbility.Val.Modes) == 0 {
		t.Fatal("expected a spell ability with at least one mode")
	}
	var bounces []game.Bounce
	for _, instruction := range face.SpellAbility.Val.Modes[0].Sequence {
		if bounce, ok := instruction.Primitive.(game.Bounce); ok {
			bounces = append(bounces, bounce)
		}
	}
	return bounces
}

// TestLowerBounceScopesShareDestinationPrecondition proves that every
// battlefield bounce-to-hand scope still lowers to the same game.Bounce shape
// after they were converged onto the shared plainControllerBounceToHand
// destination/context precondition and the shared bounce-to-hand destination
// constants. One representative card per scope (single target, two independent
// target slots, controlled choose-at-resolution, stack/permanent target union,
// and mass group) is lowered and its emitted Bounce instructions are checked for
// the scope's distinguishing routing, so a regression in the shared precondition
// would drop or mis-shape one of these without slipping past the corpus gate.
func TestLowerBounceScopesShareDestinationPrecondition(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name             string
		oracleText       string
		wantTargets      int
		wantBounces      int
		wantControlled   bool
		wantGroup        bool
		wantObjects      bool
		wantStackAllowed bool
	}{
		{
			name:        "single target",
			oracleText:  "Return target creature to its owner's hand.",
			wantTargets: 1,
			wantBounces: 1,
			wantObjects: true,
		},
		{
			name:        "two independent target slots",
			oracleText:  "Return target artifact and target creature to their owners' hands.",
			wantTargets: 2,
			wantBounces: 2,
			wantObjects: true,
		},
		{
			name:           "controlled choose at resolution",
			oracleText:     "Return a creature you control to its owner's hand.",
			wantTargets:    0,
			wantBounces:    1,
			wantControlled: true,
			wantGroup:      true,
		},
		{
			name:             "stack or permanent target union",
			oracleText:       "Return target spell to its owner's hand.",
			wantTargets:      1,
			wantBounces:      1,
			wantObjects:      true,
			wantStackAllowed: true,
		},
		{
			name:        "mass group",
			oracleText:  "Return all creatures to their owners' hands.",
			wantTargets: 0,
			wantBounces: 1,
			wantGroup:   true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Bounce Scope",
				Layout:     "normal",
				TypeLine:   "Instant",
				OracleText: test.oracleText,
			})
			mode := face.SpellAbility.Val.Modes[0]
			if len(mode.Targets) != test.wantTargets {
				t.Fatalf("targets = %d, want %d", len(mode.Targets), test.wantTargets)
			}
			bounces := bounceInstructions(t, face)
			if len(bounces) != test.wantBounces {
				t.Fatalf("bounce instructions = %d, want %d", len(bounces), test.wantBounces)
			}
			for i, bounce := range bounces {
				if bounce.ControlledChoice != test.wantControlled {
					t.Fatalf("bounce[%d].ControlledChoice = %v, want %v", i, bounce.ControlledChoice, test.wantControlled)
				}
				if bounce.Group.Empty() == test.wantGroup {
					t.Fatalf("bounce[%d].Group empty = %v, want set = %v", i, bounce.Group.Empty(), test.wantGroup)
				}
				objectSet := bounce.Object != game.ObjectReference{}
				if objectSet != test.wantObjects {
					t.Fatalf("bounce[%d].Object set = %v, want %v", i, objectSet, test.wantObjects)
				}
			}
			if test.wantStackAllowed && mode.Targets[0].Allow&game.TargetAllowStackObject == 0 {
				t.Fatalf("target allow = %v, want stack object allowed", mode.Targets[0].Allow)
			}
		})
	}
}
