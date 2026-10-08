package cardgen

import (
	"fmt"
	"testing"
)

func TestObservedPostfixOptionalCompleteCard(t *testing.T) {
	card := &ScryfallCard{
		Name: "Matter Reshaper", Layout: "normal", TypeLine: "Creature — Eldrazi", ManaCost: "{2}{C}",
		Power: new("3"), Toughness: new("2"),
		OracleText: "({C} represents colorless mana.)\nWhen this creature dies, reveal the top card of your library. You may put that card onto the battlefield if it's a permanent card with mana value 3 or less. Otherwise, put that card into your hand.",
	}
	t.Run("exact observed predicate", func(t *testing.T) {
		assertCardPaths(t, card,
			`TriggeredAbilities[0].Content.Modes[0].Sequence[1].Condition.Val.Condition.Val.Object.Val.linkID = "sequence-effect-0-product"`,
		)
	})
	t.Run("decline uses acceptance not effectiveness", func(t *testing.T) {
		assertCardPaths(t, card,
			"TriggeredAbilities[0].Content.Modes[0].Sequence[1].PublishResult",
			"TriggeredAbilities[0].Content.Modes[0].Sequence[3].ResultGate.Val.Accepted = game.TriFalse",
			"TriggeredAbilities[0].Content.Modes[0].Sequence[3].Primitive.(game.MoveCard).Destination = zone.Hand",
		)
		assertCardPathsAbsent(t, card, "TriggeredAbilities[0].Content.Modes[0].Sequence[3].ResultGate.Val.Succeeded")
	})
}

func TestObservedPostfixOptionalGenericCompositions(t *testing.T) {
	for _, producer := range []string{"Reveal", "Look at"} {
		for _, noun := range []string{"permanent card", "artifact or creature card", "instant or sorcery card"} {
			for _, shell := range []string{"spell", "trigger", "activation", "mode"} {
				t.Run(producer+"/"+noun+"/"+shell, func(t *testing.T) {
					body := producer + " the top card of your library. Scry 1. You may gain 2 life if it's a " +
						noun + " with mana value 3 or less. Otherwise, draw a card. You gain 1 life."
					card := &ScryfallCard{Name: "Postfix Composition", Layout: "normal", TypeLine: "Sorcery", OracleText: body}
					path := "SpellAbility.Val.Modes[0].Sequence"
					switch shell {
					case "spell":
					case "trigger":
						card.TypeLine, card.OracleText = "Artifact", "When this artifact enters, "+body
						path = "TriggeredAbilities[0].Content.Modes[0].Sequence"
					case "activation":
						card.TypeLine, card.OracleText = "Artifact", "{T}: "+body
						path = "ActivatedAbilities[0].Content.Modes[0].Sequence"
					case "mode":
						card.OracleText = "Choose one \u2014\n\u2022 " + body + "\n\u2022 You gain 1 life."
					default:
						t.Fatalf("unknown ability shell %q", shell)
					}
					assertCardPaths(t, card,
						path+`[2].Condition.Val.Condition.Val.Object.Val.linkID = "sequence-effect-0-product"`,
						path+"[2].PublishCondition",
						path+"[2].PublishResult",
						path+"[2].Optional = true",
						path+"[3].ConditionGateNegate = true",
						path+"[4].ResultGate.Val.Accepted = game.TriFalse",
						path+"[4].Primitive.(game.Draw)",
						path+"[5].Primitive.(game.GainLife).Amount.fixed = 1",
					)
					assertCardPathsAbsent(t, card, path+"[5].Condition", path+"[5].ResultGate",
						path+"[4].ResultGate.Val.Succeeded", path+"[4].Optional = true")
				})
			}
		}
	}
}

func TestObservedPostfixOptionalRefusesUnprovenScopes(t *testing.T) {
	for i, text := range []string{
		"Reveal the top two cards of your library. You may put that card onto the battlefield if it's a permanent card with mana value 3 or less. Otherwise, put that card into your hand.",
		"Reveal the top card of your library. You may put the looked-at card onto the battlefield if it's a permanent card with mana value 3 or less. Otherwise, put that card into your hand.",
		"Look at the top card of your library. You may put the revealed card onto the battlefield if it's a permanent card with mana value 3 or less. Otherwise, put that card into your hand.",
		"Reveal the top card of your library. You may reveal it and put that card onto the battlefield if it's a permanent card with mana value 3 or less. Otherwise, put that card into your hand.",
		"Reveal the top card of your library. You may put that card onto the battlefield if it's a permanent card with mana value 3 or less. Otherwise, draw a card and gain 1 life.",
	} {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			assertCardUnsupported(t, &ScryfallCard{
				Name: "Unproven Postfix Composition", Layout: "normal", TypeLine: "Creature \u2014 Eldrazi",
				ManaCost: "{2}{C}", Power: new("3"), Toughness: new("2"),
				OracleText: "When this creature dies, " + text,
			})
		})
	}
}

func TestObservedPredicateLeadingOtherwiseKeepsComplement(t *testing.T) {
	assertCardPaths(t, &ScryfallCard{
		Name: "Predicate Leading", Layout: "normal", TypeLine: "Sorcery",
		OracleText: "Reveal the top card of your library. If it's a permanent card with mana value 3 or less, you may put that card onto the battlefield. Otherwise, put that card into your hand.",
	}, "Sequence[2].ConditionGateNegate = true")
	assertCardPathsAbsent(t, &ScryfallCard{
		Name: "Predicate Leading", Layout: "normal", TypeLine: "Sorcery",
		OracleText: "Reveal the top card of your library. If it's a permanent card with mana value 3 or less, you may put that card onto the battlefield. Otherwise, put that card into your hand.",
	}, "Sequence[1].PublishResult", "Sequence[3].Primitive")
}
