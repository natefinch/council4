package game

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/counter"
)

func TestValidateObjectCounterDynamicAmount(t *testing.T) {
	t.Parallel()
	valid := []Instruction{{Primitive: Draw{
		Amount: Dynamic(DynamicAmount{
			Kind:        DynamicAmountObjectCounters,
			Object:      SourcePermanentReference(),
			CounterKind: counter.Burden,
		}),
		Player: ControllerReference(),
	}}}
	if err := ValidateInstructionSequence(valid); err != nil {
		t.Fatalf("valid source-counter amount rejected: %v", err)
	}

	invalid := []Instruction{{Primitive: Draw{
		Amount: Dynamic(DynamicAmount{
			Kind:        DynamicAmountObjectCounters,
			Object:      SourcePermanentReference(),
			CounterKind: counter.Kind(255),
		}),
		Player: ControllerReference(),
	}}}
	if err := ValidateInstructionSequence(invalid); err == nil {
		t.Fatal("object-counter amount with invalid counter kind was accepted")
	}
}

func TestValidatePlayerProtectionRuleEffect(t *testing.T) {
	t.Parallel()
	valid := []Instruction{{Primitive: ApplyRule{
		RuleEffects: []RuleEffect{{
			Kind:           RuleEffectPlayerProtection,
			AffectedPlayer: PlayerYou,
			Protection:     ProtectionKeyword{Everything: true},
		}},
		Duration: DurationUntilYourNextTurn,
	}}}
	if err := ValidateInstructionSequence(valid); err != nil {
		t.Fatalf("valid player protection rejected: %v", err)
	}

	for name, effect := range map[string]RuleEffect{
		"missing player": {
			Kind:       RuleEffectPlayerProtection,
			Protection: ProtectionKeyword{Everything: true},
		},
		"unsupported scope": {
			Kind:           RuleEffectPlayerProtection,
			AffectedPlayer: PlayerYou,
			Protection: ProtectionKeyword{
				FromColors: []color.Color{color.Red},
			},
		},
		"permanent scoped": {
			Kind:           RuleEffectPlayerProtection,
			AffectedPlayer: PlayerYou,
			AffectedSource: true,
			Protection:     ProtectionKeyword{Everything: true},
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			sequence := []Instruction{{Primitive: ApplyRule{
				RuleEffects: []RuleEffect{effect},
				Duration:    DurationUntilYourNextTurn,
			}}}
			if err := ValidateInstructionSequence(sequence); err == nil {
				t.Fatalf("invalid player protection accepted: %#v", effect)
			}
		})
	}
}
