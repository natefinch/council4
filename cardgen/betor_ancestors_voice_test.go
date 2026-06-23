package cardgen

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
)

func betorAncestorsVoiceCard() *ScryfallCard {
	power, toughness := "3", "5"
	return &ScryfallCard{
		Name:     "Betor, Ancestor's Voice",
		Layout:   "normal",
		ManaCost: "{2}{W}{B}{G}",
		TypeLine: "Legendary Creature — Spirit Dragon",
		OracleText: "Flying, lifelink\n" +
			"At the beginning of your end step, put a number of +1/+1 counters on up to one other target creature you control equal to the amount of life you gained this turn. Return up to one target creature card with mana value less than or equal to the amount of life you lost this turn from your graveyard to the battlefield.",
		Power:     &power,
		Toughness: &toughness,
	}
}

func TestGenerateExecutableCardSourceBetorAncestorsVoice(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(betorAncestorsVoiceCard(), "b")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"game.FlyingStaticBody",
		"game.LifelinkStaticBody",
		"Event:      game.EventBeginningOfStep",
		"Step:       game.StepEnd",
		"game.AddCounter{",
		"Kind:       game.DynamicAmountLifeGainedThisTurn",
		"CounterKind: counter.PlusOnePlusOne",
		"ManaValueDynamic: opt.Val(game.ManaValueDynamicBound{Kind: game.DynamicAmountLifeLostThisTurn, Multiplier: 1})",
		"game.PutOnBattlefield{",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}

// TestLowerBetorEndStepCounterThenReanimation asserts the end-step trigger
// lowers to an ordered two-step sequence: a dynamic +1/+1 counter placement on
// an optional other creature you control equal to the life gained this turn,
// then a reanimation of an optional graveyard creature whose mana value is
// bounded by the life lost this turn.
func TestLowerBetorEndStepCounterThenReanimation(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, betorAncestorsVoiceCard())
	if len(face.StaticAbilities) != 2 {
		t.Fatalf("static abilities = %d, want 2 (flying, lifelink)", len(face.StaticAbilities))
	}
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	trigger := face.TriggeredAbilities[0]
	if trigger.Trigger.Pattern.Event != game.EventBeginningOfStep ||
		trigger.Trigger.Pattern.Step != game.StepEnd ||
		trigger.Trigger.Pattern.Controller != game.TriggerControllerYou {
		t.Fatalf("trigger pattern = %#v", trigger.Trigger.Pattern)
	}
	if len(trigger.Content.Modes) != 1 {
		t.Fatalf("modes = %d, want 1", len(trigger.Content.Modes))
	}
	mode := trigger.Content.Modes[0]
	if len(mode.Targets) != 2 {
		t.Fatalf("targets = %d, want 2", len(mode.Targets))
	}
	for i, spec := range mode.Targets {
		if spec.MinTargets != 0 || spec.MaxTargets != 1 {
			t.Fatalf("target[%d] cardinality = {%d,%d}, want {0,1}", i, spec.MinTargets, spec.MaxTargets)
		}
	}
	reanimation := mode.Targets[1]
	if !reanimation.Selection.Exists {
		t.Fatal("reanimation target has no selection")
	}
	bound := reanimation.Selection.Val.ManaValueDynamic
	if !bound.Exists || bound.Val.Kind != game.DynamicAmountLifeLostThisTurn {
		t.Fatalf("reanimation mana-value bound = %#v, want life lost this turn", bound)
	}
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence length = %d, want 2", len(mode.Sequence))
	}
	add, ok := mode.Sequence[0].Primitive.(game.AddCounter)
	if !ok {
		t.Fatalf("sequence[0] = %#v, want AddCounter", mode.Sequence[0].Primitive)
	}
	if add.CounterKind != counter.PlusOnePlusOne {
		t.Fatalf("counter kind = %v, want +1/+1", add.CounterKind)
	}
	if add.Object != game.TargetPermanentReference(0) {
		t.Fatalf("counter object = %#v, want target 0", add.Object)
	}
	if _, ok := mode.Sequence[1].Primitive.(game.PutOnBattlefield); !ok {
		t.Fatalf("sequence[1] = %#v, want PutOnBattlefield", mode.Sequence[1].Primitive)
	}
}
