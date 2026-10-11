package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

func TestLocalProductModeScopePreservesPersistentDomains(t *testing.T) {
	for _, test := range []struct {
		name        string
		key         game.LinkedKey
		conditional bool
		repeat      bool
		want        bool
	}{
		{"unconditional overwrite", sequenceProductKey(0), false, false, true},
		{"conditional local overwrite", sequenceProductKey(0), true, false, false},
		{"repeat may not execute", sequenceProductKey(0), false, true, false},
		{"persistent CR607 link", "imprinted-card", true, false, true},
		{"noncanonical key", "sequence-effect-00-product", true, false, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			first := game.Instruction{Primitive: game.LookAtLibraryTop{Player: game.ControllerReference(), PublishLinked: test.key}}
			second := first
			if test.conditional {
				second.Condition = opt.Val(game.EffectCondition{})
			}
			if test.repeat {
				second = game.Instruction{Primitive: game.RepeatProcess{
					Body: game.Mode{Sequence: []game.Instruction{second}}.Ability(),
				}}
			}
			if got := modeResultScopesCompatible([]game.Mode{
				{Sequence: []game.Instruction{first}}, {Sequence: []game.Instruction{second}},
			}); got != test.want {
				t.Fatalf("scopes compatible=%v, want %v", got, test.want)
			}
		})
	}
}
