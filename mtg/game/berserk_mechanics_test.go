package game

import (
	"testing"

	"github.com/natefinch/council4/opt"
)

func TestObjectAttackedThisTurnConditionIsNonempty(t *testing.T) {
	t.Parallel()
	condition := Condition{
		Object:                 opt.Val(CapturedObjectReference()),
		ObjectAttackedThisTurn: true,
	}
	if condition.Empty() {
		t.Fatal("object-attacked-this-turn condition reported empty")
	}
}

func TestValidateDelayedCapturedObjectAttackedCondition(t *testing.T) {
	t.Parallel()
	captured := CapturedObjectReference()
	delayed := CreateDelayedTrigger{Trigger: DelayedTriggerDef{
		Timing:         DelayedAtBeginningOfNextEndStep,
		CapturedObject: opt.Val(TargetPermanentReference(0)),
		Content: Mode{Sequence: []Instruction{{
			Primitive: Destroy{Object: captured},
			Condition: opt.Val(EffectCondition{Condition: opt.Val(Condition{
				Object:                 opt.Val(captured),
				ObjectAttackedThisTurn: true,
			})}),
		}}}.Ability(),
	}}
	def := CardDef{CardFace: CardFace{
		Name: "Delayed Attack Test",
		SpellAbility: opt.Val(Mode{
			Targets: []TargetSpec{{
				MinTargets: 1,
				MaxTargets: 1,
				Allow:      TargetAllowPermanent,
			}},
			Sequence: []Instruction{{Primitive: delayed}},
		}.Ability()),
	}}
	if issues := ValidateCardDef(&def); len(issues) != 0 {
		t.Fatalf("ValidateCardDef: %+v", issues)
	}
}
