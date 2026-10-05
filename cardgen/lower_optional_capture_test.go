package cardgen

import "testing"

func TestOptionalCaptureUsesOriginalTypedPublication(t *testing.T) {
	for _, tt := range []struct{ name, text, producer, rider, capture, key string }{
		{"ineffective first", "You may discard a card and return target creature card from your graveyard to the battlefield. It gains haste until end of turn. Exile it at the beginning of the next end step.", "Sequence[1]", "Sequence[2]", "Sequence[3]", "sequence-effect-1-product"},
		{"expanded first", "You may add {R}{G} and return target creature card from your graveyard to the battlefield. It gains haste until end of turn. Exile it at the beginning of the next end step.", "Sequence[2]", "Sequence[3]", "Sequence[4]", "sequence-effect-1-product"},
		{"independent intervener", "You may return target creature card from your graveyard to the battlefield and gain 1 life. You gain 2 life. It gains haste until end of turn. Exile it at the beginning of the next end step.", "Sequence[0]", "Sequence[3]", "Sequence[4]", "sequence-effect-0-product"},
		{"nonzero target", "Tap target artifact. You may gain 1 life and return target creature card from your graveyard to the battlefield. It gains haste until end of turn. Exile it at the beginning of the next end step.", "Sequence[2]", "Sequence[3]", "Sequence[4]", "sequence-effect-2-product"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			card := &ScryfallCard{Name: "Optional Actual Capture", Layout: "normal", TypeLine: "Sorcery", OracleText: tt.text}
			assertCardPaths(t, card,
				tt.producer+".Primitive.(game.PutOnBattlefield).PublishLinked = \""+tt.key+"\"",
				tt.capture+".Primitive.(game.CreateDelayedTrigger).Trigger.CapturedObject.Val.linkID = \""+tt.key+"\"",
				tt.capture+".Primitive.(game.CreateDelayedTrigger).Trigger.Content.Modes[0].Sequence[0].Primitive.(game.MovePermanent).Object.kind = game.ObjectReferenceCapturedObject")
			assertCardPathsAbsent(t, card, tt.capture+".Optional = true", tt.capture+".OptionalDecisionGate",
				tt.capture+".Primitive.(game.CreateDelayedTrigger).Trigger.Optional = true",
				tt.rider+".Primitive.(game.ApplyContinuous).PublishLinked")
		})
	}
}

func TestOptionalCaptureMoveIsTransientAndOriginal(t *testing.T) {
	card := &ScryfallCard{Name: "Optional Move Capture", Layout: "normal", TypeLine: "Sorcery",
		OracleText: "You may gain 1 life and exile target creature. You gain 2 life. Return that card to its owner's hand at the beginning of the next end step."}
	assertCardPaths(t, card,
		"Sequence[0].PublishOptionalDecision",
		"Sequence[1].OptionalDecisionGate",
		"Sequence[1].ClearLinkedBeforeGate = true",
		"Sequence[1].Primitive.(game.MovePermanent).PublishLinked = \"sequence-effect-1-product\"",
		"Sequence[3].Primitive.(game.CreateDelayedTrigger).Trigger.CapturedCard.Val.linkID = \"sequence-effect-1-product\"")
}
