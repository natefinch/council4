package cardgen

import (
	"reflect"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
)

func TestLowerFlameshadowConjuringPaidTemporaryCopy(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Flameshadow Conjuring",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		ManaCost:   "{3}{R}",
		OracleText: "Whenever a nontoken creature you control enters, you may pay {R}. If you do, create a token that's a copy of that creature. That token gains haste. Exile it at the beginning of the next end step.",
	})
	ability := face.TriggeredAbilities[0]
	if len(ability.Content.Modes[0].Sequence) != 3 {
		t.Fatalf("sequence = %#v, want pay, copy, delayed exile", ability.Content.Modes[0].Sequence)
	}
	pay, ok := ability.Content.Modes[0].Sequence[0].Primitive.(game.Pay)
	if !ok || !pay.Payment.ManaCost.Exists ||
		len(pay.Payment.ManaCost.Val) != 1 || pay.Payment.ManaCost.Val[0] != cost.R {
		t.Fatalf("payment = %#v, want {R}", ability.Content.Modes[0].Sequence[0])
	}
	create, ok := ability.Content.Modes[0].Sequence[1].Primitive.(game.CreateToken)
	wantSource := game.TokenCopyOf(game.TokenCopySpec{
		Source:      game.TokenCopySourceObject,
		Object:      game.EventPermanentReference(),
		AddKeywords: []game.Keyword{game.Haste},
	})
	if !ok || !reflect.DeepEqual(create.Source, wantSource) {
		t.Fatalf("copy = %#v, want hasty token copy", ability.Content.Modes[0].Sequence[1])
	}
	if create.PublishLinked == "" {
		t.Fatal("copy token is not published for delayed cleanup")
	}
	cleanup, ok := ability.Content.Modes[0].Sequence[2].Primitive.(game.CreateDelayedTrigger)
	if !ok || !cleanup.Trigger.CapturedObject.Exists ||
		!reflect.DeepEqual(cleanup.Trigger.CapturedObject.Val, game.LinkedObjectReference(string(create.PublishLinked))) {
		t.Fatalf("cleanup = %#v, want delayed exile", ability.Content.Modes[0].Sequence[2])
	}
	exile, ok := cleanup.Trigger.Content.Modes[0].Sequence[0].Primitive.(game.Exile)
	if !ok || !reflect.DeepEqual(exile.Object, game.CapturedObjectReference()) {
		t.Fatalf("cleanup object = %#v, want captured token", exile.Object)
	}
}
