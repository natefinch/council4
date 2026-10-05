package cardgen

import "testing"

func TestLibraryCardConditionKeepsEventConsumerAndUnconditionalCardRiderDistinct(t *testing.T) {
	card := &ScryfallCard{
		Name: "Lost in the Woods", Layout: "normal", TypeLine: "Enchantment", ManaCost: "{3}{G}{G}",
		OracleText: "Whenever a creature attacks you or a planeswalker you control, reveal the top card of your library. If it's a Forest card, remove that creature from combat. Then put the revealed card on the bottom of your library.",
	}
	assertCardPaths(t, card,
		`Sequence[0].Primitive.(game.Reveal).PublishLinked = "sequence-effect-0-product"`,
		"Sequence[1].Primitive.(game.RemoveFromCombat)",
		`Sequence[1].Condition.Val.Condition.Val.Object.Val.linkID = "sequence-effect-0-product"`,
		`Sequence[2].Primitive.(game.MoveCard).Card.LinkID = "sequence-effect-0-product"`,
		"Sequence[2].Primitive.(game.MoveCard).DestinationBottom = true",
	)
	assertCardPathsAbsent(t, card,
		"Sequence[2].Condition", "Sequence[2].ConditionGate", "Sequence[2].ResultGate",
	)
}
