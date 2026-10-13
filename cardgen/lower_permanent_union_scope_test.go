package cardgen

import "testing"

func TestPermanentUnionSharedScopeWholeCardRefuses(t *testing.T) {
	assertCardUnsupported(t, &ScryfallCard{
		Name: "Shared Subtype Prevention", Layout: "normal", TypeLine: "Creature", Power: new("2"), Toughness: new("3"),
		OracleText: "{1}{W}: Prevent all damage that would be dealt to target tapped Merfolk or Kithkin creature this turn.",
	})
	for _, phrase := range []string{
		"target tapped Merfolk or Kithkin creature",
		"target Merfolk or Kithkin creature",
		"target blue Merfolk or Kithkin creature",
		"target nontoken artifact or enchantment an opponent controls",
		"target creature token or artifact",
		"target tapped creature or Vehicle",
		"target attacking creature or Vehicle",
		"target multicolored creature or enchantment",
		"target legendary creature or Vehicle",
		"target creature or Vehicle with mana value 3 or less",
	} {
		for _, text := range []string{"Tap " + phrase + ".", "Return " + phrase + " to its owner's hand. Draw a card."} {
			t.Run(text, func(t *testing.T) {
				assertCardUnsupported(t, &ScryfallCard{
					Name: "Union Scope Refusal", Layout: "normal", TypeLine: "Instant", OracleText: text,
				})
			})
		}
	}
}

func TestPermanentUnionIndependentScopeCompiledControls(t *testing.T) {
	for _, text := range []string{
		"Tap target tapped Merfolk creature or tapped Kithkin creature.",
		"Return target Vehicle or creature to its owner's hand.",
		"Destroy target blue creature or artifact.",
		"Untap target creature or tapped Vehicle.",
		"Exile target multicolored creature or multicolored enchantment.",
		"Destroy target enchanted creature or enchantment creature. You gain 3 life.",
		"Return target Saga or enchantment creature you control to its owner's hand.",
	} {
		t.Run(text, func(t *testing.T) {
			assertCardPaths(t, &ScryfallCard{
				Name: "Union Scope Control", Layout: "normal", TypeLine: "Instant", OracleText: text,
			}, "Modes[0].Targets[0].Allow = game.TargetAllowPermanent",
				"Modes[0].Targets[0].Selection.Val.AnyOf[1].")
		})
	}
}
