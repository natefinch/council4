package cardgen

import "testing"

func TestPostfixTargetTypeConditionSubjects(t *testing.T) {
	t.Parallel()
	for _, keyword := range []string{"deathtouch", "hexproof"} {
		t.Run(keyword, func(t *testing.T) {
			t.Parallel()
			assertCardPaths(t, &ScryfallCard{
				Name: "Postfix Subject Probe", Layout: "normal", TypeLine: "Creature — Human",
				ManaCost: "{1}{W}", Power: new("1"), Toughness: new("1"),
				OracleText: "When Postfix Subject Probe enters, target permanent you control gains " + keyword +
					" until end of turn. Put a +1/+1 counter on it if it's a creature. Put a loyalty counter on it if it's a planeswalker.",
			},
				"TriggeredAbilities[0].Content.Modes[0].Sequence[1].Condition.Val.Condition.Val.Object.Val.kind = game.ObjectReferenceTargetPermanent",
				"TriggeredAbilities[0].Content.Modes[0].Sequence[1].Condition.Val.Condition.Val.ObjectMatches.Val.RequiredTypes[0] = types.Creature",
				"TriggeredAbilities[0].Content.Modes[0].Sequence[2].Condition.Val.Condition.Val.Object.Val.kind = game.ObjectReferenceTargetPermanent",
				"TriggeredAbilities[0].Content.Modes[0].Sequence[2].Condition.Val.Condition.Val.ObjectMatches.Val.RequiredTypes[0] = types.Planeswalker",
			)
		})
	}
}

func TestPostfixReturnedTypeConditionSubject(t *testing.T) {
	t.Parallel()
	assertCardPaths(t, &ScryfallCard{
		Name: "Postfix Return Probe", Layout: "normal", TypeLine: "Sorcery",
		OracleText: "Return target creature card from your graveyard to the battlefield. Put a +1/+1 counter on it if it's a creature.",
	},
		"SpellAbility.Val.Modes[0].Sequence[0].Primitive.(game.PutOnBattlefield).PublishLinked = \"sequence-effect-0-product\"",
		"SpellAbility.Val.Modes[0].Sequence[1].Condition.Val.Condition.Val.Object.Val.kind = game.ObjectReferenceLinkedObject",
		"SpellAbility.Val.Modes[0].Sequence[1].Condition.Val.Condition.Val.Object.Val.linkID = \"sequence-effect-0-product\"",
		"SpellAbility.Val.Modes[0].Sequence[1].Condition.Val.Condition.Val.ObjectMatches.Val.RequiredTypes[0] = types.Creature",
		"SpellAbility.Val.Modes[0].Sequence[1].Primitive.(game.AddCounter).Object.kind = game.ObjectReferenceLinkedObject",
		"SpellAbility.Val.Modes[0].Sequence[1].Primitive.(game.AddCounter).Object.linkID = \"sequence-effect-0-product\"",
	)
}
