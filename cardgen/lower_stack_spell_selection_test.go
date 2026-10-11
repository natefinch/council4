package cardgen

import "testing"

func TestStackSpellSelectionCompositions(t *testing.T) {
	for _, tc := range []struct {
		text string
		want []string
	}{
		{"Counter target enchantment spell.", []string{
			"Modes[0].Targets[0].Predicate.SpellCardTypes[0] = types.Enchantment",
			"Primitive.(game.CounterObject).Object.kind = game.ObjectReferenceTargetStackObject",
		}},
		{"Counter target blue creature spell you don't control with mana value 3 or less.", []string{
			"Modes[0].Targets[0].Predicate.SpellCardTypes[0] = types.Creature",
			"Modes[0].Targets[0].Predicate.SpellColors[0] = color.Blue",
			"Modes[0].Targets[0].Predicate.Controller = game.ControllerNotYou",
			"Modes[0].Targets[0].Predicate.ManaValue.Val.Op = compare.LessOrEqual",
			"Modes[0].Targets[0].Predicate.ManaValue.Val.Value = 3",
		}},
		{"Counter target nonblue enchantment spell unless its controller pays {2}.", []string{
			"Modes[0].Targets[0].Predicate.SpellCardTypes[0] = types.Enchantment",
			"Modes[0].Targets[0].Predicate.SpellExcludedColors[0] = color.Blue",
			"Sequence[0].Primitive.(game.Pay).Payment.ManaCost.Val[0] = {2}",
			"Sequence[1].Primitive.(game.CounterObject).",
		}},
		{"Exile any number of target artifact or enchantment spells an opponent controls with mana value 3 or less.", []string{
			"Modes[0].Targets[0].MaxTargets = 99",
			"Modes[0].Targets[0].Predicate.SpellCardTypesAny[0] = types.Artifact",
			"Modes[0].Targets[0].Predicate.SpellCardTypesAny[1] = types.Enchantment",
			"Modes[0].Targets[0].Predicate.Controller = game.ControllerOpponent",
			"Primitive.(game.ExileTargetSpells).",
		}},
		{"Copy target colorless enchantment spell you control. You may choose new targets for the copy.", []string{
			"Modes[0].Targets[0].Predicate.SpellCardTypes[0] = types.Enchantment",
			"Modes[0].Targets[0].Predicate.SpellColorless = true",
			"Modes[0].Targets[0].Predicate.Controller = game.ControllerYou",
			"Primitive.(game.CopyStackObject).",
		}},
		{"Gain control of target multicolored noncreature spell.", []string{
			"Modes[0].Targets[0].Predicate.ExcludedSpellCardTypes[0] = types.Creature",
			"Modes[0].Targets[0].Predicate.SpellMulticolored = true",
			"Primitive.(game.ChangeStackObjectController).",
		}},
		{"Choose new targets for target noncreature enchantment spell.", []string{
			"Modes[0].Targets[0].Predicate.SpellCardTypes[0] = types.Enchantment",
			"Modes[0].Targets[0].Predicate.ExcludedSpellCardTypes[0] = types.Creature",
			"Primitive.(game.ChooseNewTargets).",
		}},
	} {
		t.Run(tc.text, func(t *testing.T) {
			card := &ScryfallCard{Name: "Selector Composition", Layout: "normal", TypeLine: "Instant", OracleText: tc.text}
			assertCardPaths(t, card, tc.want...)
			assertCardPathsAbsent(t, card, "Predicate.SpellSupertypes", "Selection.Val")
		})
	}
}

func TestStackSpellSelectionRefusals(t *testing.T) {
	for _, selector := range []string{
		"blue enchantment or instant spell",
		"enchantment or blue instant spell",
		"blue enchantment or red instant spell",
		"noncreature artifact or enchantment spell",
		"legendary enchantment spell",
		"Aura spell",
		"enchantment spell with flying",
		"enchantment spell without flying",
		"enchantment spell you own",
		"enchantment spell from your graveyard",
		"enchantment spell with mana value X",
	} {
		for _, verb := range []string{"Counter target ", "Exile any number of target "} {
			text := verb + selector + "."
			t.Run(text, func(t *testing.T) {
				assertCardUnsupported(t, &ScryfallCard{
					Name: "Selector Refusal", Layout: "normal", TypeLine: "Instant", OracleText: text,
				})
			})
		}
	}
	assertCardUnsupported(t, &ScryfallCard{
		Name: "Bounded Counter", Layout: "normal", TypeLine: "Instant",
		OracleText: "Counter two target blue enchantment spells.",
	})
}

func TestStackSpellSelectionArensonsAura(t *testing.T) {
	card := &ScryfallCard{
		Name: "Arenson's Aura", Layout: "normal", TypeLine: "Enchantment", ManaCost: "{2}{W}",
		OracleText: "{W}, Sacrifice an enchantment: Destroy target enchantment.\n" +
			"{3}{U}{U}: Counter target enchantment spell.",
	}
	assertCardPaths(t, card,
		"ActivatedAbilities[0].ManaCost.Val[0] = {W}",
		"ActivatedAbilities[0].AdditionalCosts[0].Kind = cost.AdditionalSacrifice",
		"ActivatedAbilities[0].AdditionalCosts[0].Amount = 1",
		"ActivatedAbilities[0].AdditionalCosts[0].PermanentType = types.Enchantment",
		"ActivatedAbilities[0].Content.Modes[0].Sequence[0].Primitive.(game.Destroy).",
		"ActivatedAbilities[1].ManaCost.Val[0] = {3}",
		"ActivatedAbilities[1].ManaCost.Val[1] = {U}",
		"ActivatedAbilities[1].ManaCost.Val[2] = {U}",
		"ActivatedAbilities[1].Content.Modes[0].Targets[0].Predicate.SpellCardTypes[0] = types.Enchantment",
		"ActivatedAbilities[1].Content.Modes[0].Sequence[0].Primitive.(game.CounterObject).",
	)
}
