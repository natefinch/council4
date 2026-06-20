package game

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestCapturedTargetManaValueUsesEnclosingTargetNamespace(t *testing.T) {
	t.Parallel()
	delayed := CreateDelayedTrigger{Trigger: DelayedTriggerDef{
		Timing: DelayedAtBeginningOfNextMainPhase,
		Content: Mode{
			Targets: []TargetSpec{{
				MinTargets: 1,
				MaxTargets: 1,
				Allow:      TargetAllowPermanent,
			}},
			Sequence: []Instruction{{Primitive: AddMana{
				Amount: Dynamic(DynamicAmount{
					Kind:   DynamicAmountCapturedTargetManaValue,
					Object: CapturedTargetStackObjectReference(0),
				}),
				ManaColor: mana.C,
			}}},
		}.Ability(),
	}}
	card := &CardDef{CardFace: CardFace{
		Name:  "Captured Mana",
		Types: []types.Card{types.Instant},
		SpellAbility: opt.Val(Mode{
			Targets: []TargetSpec{{
				MinTargets: 1,
				MaxTargets: 1,
				Allow:      TargetAllowStackObject,
				Predicate:  TargetPredicate{StackObjectKinds: []StackObjectKind{StackSpell}},
			}},
			Sequence: []Instruction{{Primitive: delayed}},
		}.Ability()),
	}}
	if issues := ValidateCardDef(card); len(issues) != 0 {
		t.Fatalf("ValidateCardDef() issues = %#v", issues)
	}
}

func TestCapturedTargetManaValueRejectsInvalidEnclosingTarget(t *testing.T) {
	t.Parallel()
	sequence := []Instruction{{Primitive: CreateDelayedTrigger{Trigger: DelayedTriggerDef{
		Timing: DelayedAtBeginningOfNextMainPhase,
		Content: Mode{Sequence: []Instruction{{Primitive: AddMana{
			Amount: Dynamic(DynamicAmount{
				Kind:   DynamicAmountCapturedTargetManaValue,
				Object: CapturedTargetStackObjectReference(1),
			}),
			ManaColor: mana.C,
		}}}}.Ability(),
	}}}}
	err := ValidateInstructionSequence(sequence, []TargetSpec{{
		MinTargets: 1,
		MaxTargets: 1,
		Allow:      TargetAllowStackObject,
		Predicate:  TargetPredicate{StackObjectKinds: []StackObjectKind{StackSpell}},
	}})
	if err == nil || !strings.Contains(err.Error(), "target index 1") {
		t.Fatalf("ValidateInstructionSequence() error = %v, want captured target bounds error", err)
	}
}

func TestCapturedTargetManaValueRequiresCapturedReferenceNamespace(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name     string
		kind     DynamicAmountKind
		object   ObjectReference
		wantText string
	}{
		{
			name:     "captured amount with local target",
			kind:     DynamicAmountCapturedTargetManaValue,
			object:   TargetStackObjectReference(0),
			wantText: "requires a captured target stack object reference",
		},
		{
			name:     "ordinary amount with captured target",
			kind:     DynamicAmountObjectManaValue,
			object:   CapturedTargetStackObjectReference(0),
			wantText: "requires a captured target mana value amount",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			err := validateQuantity(Dynamic(DynamicAmount{
				Kind:   test.kind,
				Object: test.object,
			}), nil, false)
			if err == nil || !strings.Contains(err.Error(), test.wantText) {
				t.Fatalf("validateQuantity() error = %v, want %q", err, test.wantText)
			}
		})
	}
}
