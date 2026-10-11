package cardgen

import "testing"

func TestSequencePaymentDoesNotStealControllerResult(t *testing.T) {
	card := &ScryfallCard{
		Name: "Controller Result Beside Payment", Layout: "normal", TypeLine: "Instant",
		OracleText: "Counter target spell unless its controller pays {3}. You gain 1 life. If you do, draw a card.",
	}
	assertCardPaths(t, card,
		"SpellAbility.Val.Modes[0].Sequence[0].PublishResult = \"unless-paid\"",
		"SpellAbility.Val.Modes[0].Sequence[1].ResultGate.Val.Key = \"unless-paid\"",
		"SpellAbility.Val.Modes[0].Sequence[2].Primitive.(game.GainLife).Amount.fixed = 1",
		"SpellAbility.Val.Modes[0].Sequence[2].PublishResult = \"if-you-do\"",
		"SpellAbility.Val.Modes[0].Sequence[3].Primitive.(game.Draw)",
		"SpellAbility.Val.Modes[0].Sequence[3].ResultGate.Val.Key = \"if-you-do\"",
	)
}
