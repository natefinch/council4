package cardgen

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

func TestLowerBerserkReusableMechanics(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Test Berserk",
		Layout:   "normal",
		TypeLine: "Instant",
		ManaCost: "{G}",
		OracleText: "Cast this spell only before the combat damage step.\n" +
			"Target creature gains trample and gets +X/+0 until end of turn, where X is its power. " +
			"At the beginning of the next end step, destroy that creature if it attacked this turn.",
	})
	if len(face.StaticAbilities) != 1 ||
		!face.StaticAbilities[0].Body.CastOnlyBeforeCombatDamageStep {
		t.Fatalf("static abilities = %#v, want before-combat-damage restriction", face.StaticAbilities)
	}

	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 || len(mode.Sequence) != 3 {
		t.Fatalf("mode = %#v, want one target and three instructions", mode)
	}
	grant, ok := mode.Sequence[0].Primitive.(game.ApplyContinuous)
	if !ok || len(grant.ContinuousEffects) != 1 ||
		len(grant.ContinuousEffects[0].AddKeywords) != 1 ||
		grant.ContinuousEffects[0].AddKeywords[0] != game.Trample {
		t.Fatalf("grant = %#v, want target trample grant", mode.Sequence[0].Primitive)
	}
	modify, ok := mode.Sequence[1].Primitive.(game.ModifyPT)
	if !ok {
		t.Fatalf("pump = %T, want game.ModifyPT", mode.Sequence[1].Primitive)
	}
	amount := modify.PowerDelta.DynamicAmount()
	if !amount.Exists ||
		amount.Val.Kind != game.DynamicAmountObjectPower ||
		amount.Val.Object != game.TargetPermanentReference(0) ||
		modify.PublishLinked == "" {
		t.Fatalf("pump = %#v, want published target-power snapshot", modify)
	}
	delayed, ok := mode.Sequence[2].Primitive.(game.CreateDelayedTrigger)
	if !ok ||
		delayed.Trigger.Timing != game.DelayedAtBeginningOfNextEndStep ||
		!delayed.Trigger.CapturedObject.Exists {
		t.Fatalf("delayed = %#v, want captured next-end trigger", mode.Sequence[2].Primitive)
	}
	instruction := delayed.Trigger.Content.Modes[0].Sequence[0]
	destroy, ok := instruction.Primitive.(game.Destroy)
	if !ok || destroy.Object.Kind() != game.ObjectReferenceCapturedObject {
		t.Fatalf("destroy = %#v, want captured-object destroy", instruction.Primitive)
	}
	if !instruction.Condition.Exists ||
		!instruction.Condition.Val.Condition.Exists ||
		!instruction.Condition.Val.Condition.Val.ObjectAttackedThisTurn {
		t.Fatalf("destroy condition = %#v, want captured object attacked this turn", instruction.Condition)
	}
}

func TestRenderObjectAttackedThisTurnCondition(t *testing.T) {
	t.Parallel()
	rendered, err := (Renderer{}).renderControllerControlsCondition(newRenderCtx(), &game.Condition{
		Object:                 opt.Val(game.CapturedObjectReference()),
		ObjectAttackedThisTurn: true,
	}, "Berserk delayed destroy")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(rendered, "ObjectAttackedThisTurn: true") {
		t.Fatalf("rendered condition omitted attacked-this-turn predicate:\n%s", rendered)
	}
}

func TestLowerUnconditionalDelayedTargetDestroy(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Blood Frenzy",
		Layout:     "normal",
		TypeLine:   "Instant",
		ManaCost:   "{1}{R}",
		OracleText: "Cast this spell only before the combat damage step.\nTarget attacking or blocking creature gets +4/+0 until end of turn. Destroy that creature at the beginning of the next end step.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence length = %d, want pump and delayed trigger", len(mode.Sequence))
	}
	delayed, ok := mode.Sequence[1].Primitive.(game.CreateDelayedTrigger)
	if !ok || delayed.Trigger.Timing != game.DelayedAtBeginningOfNextEndStep ||
		!delayed.Trigger.CapturedObject.Exists {
		t.Fatalf("second primitive = %#v, want captured next-end-step trigger", mode.Sequence[1].Primitive)
	}
	destroy := delayed.Trigger.Content.Modes[0].Sequence[0]
	if destroy.Condition.Exists {
		t.Fatalf("unconditional delayed destroy has condition %#v", destroy.Condition)
	}
}
