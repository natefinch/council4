package cardgen

import "testing"

func TestSequenceCounterTaxEventSpell(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name: "Tax Watcher", Layout: "normal", TypeLine: "Enchantment",
		OracleText: "Whenever this enchantment becomes the target of a spell or ability, counter that spell or ability unless its controller pays {3}. You gain 2 life.",
	}
	assertCardPaths(t, card,
		"TriggeredAbilities[0].Content.Modes[0].Sequence[0].Primitive.(game.Pay).Payment.Payer.Val.object.Val.kind = game.ObjectReferenceEventStackObject",
		"TriggeredAbilities[0].Content.Modes[0].Sequence[1].Primitive.(game.CounterObject).Object.kind = game.ObjectReferenceEventStackObject",
		"TriggeredAbilities[0].Content.Modes[0].Sequence[2].Primitive.(game.GainLife).Amount.fixed = 2",
	)
}

func TestSequenceCounterTaxRealCards(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, mana, typeLine, text string
	}{
		{"Offering to Asha", "{2}{W}{U}", "Instant", "Counter target spell unless its controller pays {4}. You gain 4 life."},
		{"Condescend", "{X}{U}", "Instant", "Counter target spell unless its controller pays {X}. Scry 2. (Look at the top two cards of your library, then put any number of them on the bottom and the rest on top in any order.)"},
		{"Mindswipe", "{X}{U}{R}", "Instant", "Counter target spell unless its controller pays {X}. Mindswipe deals X damage to that spell's controller."},
		{"Frightful Delusion", "{2}{U}", "Instant", "Counter target spell unless its controller pays {1}. That player discards a card."},
		{"Sage's Dousing", "{2}{U}", "Kindred Instant — Wizard", "Counter target spell unless its controller pays {3}. If you control a Wizard, draw a card."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{Name: tt.name, ManaCost: tt.mana, Layout: "normal", TypeLine: tt.typeLine, OracleText: tt.text}
			assertCardPaths(t, card,
				"SpellAbility.Val.Modes[0].Sequence[0].PublishResult = \"unless-paid\"",
				"SpellAbility.Val.Modes[0].Sequence[1].ResultGate.Val.Succeeded = game.TriFalse",
			)
		})
	}
}
