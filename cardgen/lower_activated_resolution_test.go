package cardgen

import "testing"

func TestActivatedResolutionOrdinal(t *testing.T) {
	t.Parallel()
	for _, text := range []string{
		"{1}: If this is the second time this ability has resolved this turn, draw a card.",
		"{1}: Choose one \u2014\n\u2022 If this is the second time this ability has resolved this turn, draw a card.\n\u2022 You gain 2 life.",
	} {
		card := &ScryfallCard{Name: "Ordinal", Layout: "normal", TypeLine: "Artifact", OracleText: text}
		assertCardPaths(t, card,
			"ActivatedAbilities[0].CountsResolutionsThisTurn = true",
			"ActivatedAbilities[0].Content.Modes[0].Sequence[0].Condition.Val.Condition.Val.SourceAbilityResolutionOrdinalThisTurn = 2",
		)
		assertCardPathsAbsent(t, card, "ActivatedAbilities[0].ActivationCondition")
	}
	card := &ScryfallCard{
		Name: "Inner-Flame Igniter", Layout: "normal", TypeLine: "Creature - Elemental Warrior",
		OracleText: "{2}{R}: Creatures you control get +1/+0 until end of turn. If this is the third time this ability has resolved this turn, creatures you control gain first strike until end of turn.",
	}
	assertCardPaths(t, card,
		"ActivatedAbilities[0].CountsResolutionsThisTurn = true",
		"ActivatedAbilities[0].Content.Modes[0].Sequence[1].Condition.Val.Condition.Val.SourceAbilityResolutionOrdinalThisTurn = 3",
	)
}

func TestActivatedResolutionOrdinalNearMisses(t *testing.T) {
	t.Parallel()
	for _, text := range []string{
		"{1}: Draw a card. Activate only if this is the second time this ability has resolved this turn.",
		"{1}: If this is the second time this ability has been activated this turn, draw a card.",
		"{1}: If this is the second time an ability has resolved this turn, draw a card.",
		"{T}: If this is the second time this ability has resolved this turn, add {G}.",
		"{T}: Draw a card. If this is the second time this ability has resolved this turn, add {G}.",
	} {
		assertCardUnsupported(t, &ScryfallCard{Name: "Ordinal Near Miss", Layout: "normal", TypeLine: "Artifact", OracleText: text})
	}
	assertCardPathsAbsent(t, &ScryfallCard{
		Name: "Ordinary", Layout: "normal", TypeLine: "Artifact", OracleText: "{1}: Draw a card.",
	}, "ActivatedAbilities[0].CountsResolutionsThisTurn")
}
