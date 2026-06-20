package cardgen

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestLowerManaDrainEndToEnd(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Mana Drain",
		Layout:     "normal",
		TypeLine:   "Instant",
		ManaCost:   "{U}{U}",
		OracleText: "Counter target spell. At the beginning of your next main phase, add an amount of {C} equal to that spell's mana value.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 || len(mode.Sequence) != 2 {
		t.Fatalf("mode = %#v, want one target and two instructions", mode)
	}
	if _, ok := mode.Sequence[0].Primitive.(game.CounterObject); !ok {
		t.Fatalf("first primitive = %T, want CounterObject", mode.Sequence[0].Primitive)
	}
	delayed, ok := mode.Sequence[1].Primitive.(game.CreateDelayedTrigger)
	if !ok || delayed.Trigger.Timing != game.DelayedAtBeginningOfNextMainPhase {
		t.Fatalf("second primitive = %#v, want next-main delayed trigger", mode.Sequence[1].Primitive)
	}
	add, ok := delayed.Trigger.Content.Modes[0].Sequence[0].Primitive.(game.AddMana)
	if !ok || add.ManaColor != mana.C || !add.Amount.IsDynamic() {
		t.Fatalf("delayed primitive = %#v, want dynamic colorless mana", delayed.Trigger.Content)
	}
	dynamic := add.Amount.DynamicAmount().Val
	if dynamic.Kind != game.DynamicAmountCapturedTargetManaValue ||
		dynamic.Object.Kind() != game.ObjectReferenceCapturedTargetStackObject ||
		dynamic.Object.TargetIndex() != 0 {
		t.Fatalf("dynamic amount = %#v", dynamic)
	}
	card := &game.CardDef{CardFace: game.CardFace{
		Name:         "Mana Drain",
		Types:        []types.Card{types.Instant},
		SpellAbility: face.SpellAbility,
	}}
	if issues := game.ValidateCardDef(card); len(issues) != 0 {
		t.Fatalf("generated Mana Drain validation issues = %#v", issues)
	}
	rendered, err := (Renderer{}).renderPrimitive(newRenderCtx(), delayed)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"game.DelayedAtBeginningOfNextMainPhase",
		"game.DynamicAmountCapturedTargetManaValue",
		"game.CapturedTargetStackObjectReference(0)",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("rendered Mana Drain rider missing %q:\n%s", want, rendered)
		}
	}
}
