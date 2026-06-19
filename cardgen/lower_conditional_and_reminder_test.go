package cardgen

import (
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestLowerConditionalEntersTappedReplacement(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Vista",
		Layout:     "normal",
		TypeLine:   "Land — Forest Plains",
		OracleText: "This land enters tapped unless you control two or more basic lands.",
	})
	if len(face.ReplacementAbilities) != 1 {
		t.Fatalf("got %d replacement abilities, want 1", len(face.ReplacementAbilities))
	}
	repl := face.ReplacementAbilities[0]
	if !repl.Replacement.EntersTapped {
		t.Fatal("replacement is not an enters-tapped replacement")
	}
	if !repl.Replacement.Condition.Exists {
		t.Fatal("conditional replacement has no condition")
	}
	cond := repl.Replacement.Condition.Val
	if !cond.Negate {
		t.Fatal("condition should be negated (unless)")
	}
	if !cond.ControlsMatching.Exists {
		t.Fatal("condition has no matching-selection count")
	}
	filter := cond.ControlsMatching.Val.Selection
	if len(filter.RequiredTypes) != 1 || filter.RequiredTypes[0] != types.Land {
		t.Fatalf("filter types = %#v, want [types.Land]", filter.RequiredTypes)
	}
	if len(filter.Supertypes) != 1 || filter.Supertypes[0] != types.Basic {
		t.Fatalf("filter supertypes = %#v, want [types.Basic]", filter.Supertypes)
	}
	if cond.ControlsMatching.Val.MinCount != 2 {
		t.Fatalf("filter MinCount = %d, want 2", cond.ControlsMatching.Val.MinCount)
	}
}

func TestLowerEntersTappedUnlessLegendaryCreatureReplacement(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Minas Tirith",
		Layout:     "normal",
		TypeLine:   "Legendary Land",
		OracleText: "Minas Tirith enters tapped unless you control a legendary creature.",
	})
	if len(face.ReplacementAbilities) != 1 {
		t.Fatalf("got %d replacement abilities, want 1", len(face.ReplacementAbilities))
	}
	repl := face.ReplacementAbilities[0]
	if !repl.Replacement.EntersTapped {
		t.Fatal("replacement is not an enters-tapped replacement")
	}
	if !repl.Replacement.Condition.Exists {
		t.Fatal("conditional replacement has no condition")
	}
	cond := repl.Replacement.Condition.Val
	if !cond.Negate {
		t.Fatal("condition should be negated (unless)")
	}
	if !cond.ControlsMatching.Exists {
		t.Fatal("condition has no matching-selection count")
	}
	filter := cond.ControlsMatching.Val.Selection
	if len(filter.RequiredTypes) != 1 || filter.RequiredTypes[0] != types.Creature {
		t.Fatalf("filter types = %#v, want [types.Creature]", filter.RequiredTypes)
	}
	if len(filter.Supertypes) != 1 || filter.Supertypes[0] != types.Legendary {
		t.Fatalf("filter supertypes = %#v, want [types.Legendary]", filter.Supertypes)
	}
	// "a legendary creature" carries no explicit count, so MinCount defaults to 1
	// at evaluation time (an empty MinCount with a non-empty Selection).
	if cond.ControlsMatching.Val.MinCount != 0 {
		t.Fatalf("filter MinCount = %d, want 0 (defaults to 1)", cond.ControlsMatching.Val.MinCount)
	}
}

func TestLowerEntersTappedUnlessWorldSupertypeFailsClosed(t *testing.T) {
	t.Parallel()
	// "world" is outside the closed condition supertype vocabulary, so the
	// conditional enters-tapped replacement must not lower to a controls
	// predicate and the card stays unsupported rather than guessing a meaning.
	face := lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:       "Test Land",
		Layout:     "normal",
		TypeLine:   "Land",
		OracleText: "This land enters tapped unless you control a world enchantment.",
	})
	if len(face.ReplacementAbilities) != 0 {
		t.Fatalf("got %d replacement abilities, want 0", len(face.ReplacementAbilities))
	}
}

func TestLowerCommonConditionalEntersTappedReplacements(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		oracleText    string
		negate        bool
		minCount      int
		excludeSource bool
		subtypes      []types.Sub
	}{
		{
			name:          "two or more other lands",
			oracleText:    "This land enters tapped unless you control two or more other lands.",
			negate:        true,
			minCount:      2,
			excludeSource: true,
		},
		{
			name:          "two or fewer other lands",
			oracleText:    "This land enters tapped unless you control two or fewer other lands.",
			minCount:      3,
			excludeSource: true,
		},
		{
			name:       "basic land subtype pair",
			oracleText: "This land enters tapped unless you control a Plains or an Island.",
			subtypes:   []types.Sub{types.Plains, types.Island},
			negate:     true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Land",
				Layout:     "normal",
				TypeLine:   "Land",
				OracleText: test.oracleText,
			})
			condition := face.ReplacementAbilities[0].Replacement.Condition.Val
			if !condition.ControlsMatching.Exists {
				t.Fatal("condition has no matching-selection count")
			}
			filter := condition.ControlsMatching.Val.Selection
			if condition.Negate != test.negate ||
				condition.ControlsMatching.Val.MinCount != test.minCount ||
				filter.ExcludeSource != test.excludeSource ||
				!slices.Equal(filter.SubtypesAny, test.subtypes) {
				t.Fatalf("condition = %+v, want negate=%v min=%d exclude=%v subtypes=%v",
					condition, test.negate, test.minCount, test.excludeSource, test.subtypes)
			}
		})
	}
}

func TestLowerLifeAndOpponentConditionalEntersTappedReplacements(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		condition string
		assert    func(*testing.T, game.Condition)
	}{
		{
			name:      "controller life",
			condition: "unless you have 10 or more life",
			assert: func(t *testing.T, condition game.Condition) {
				if condition.ControllerLifeAtLeast != 10 {
					t.Fatalf("ControllerLifeAtLeast = %d, want 10", condition.ControllerLifeAtLeast)
				}
			},
		},
		{
			name:      "any player life",
			condition: "unless a player has 13 or less life",
			assert: func(t *testing.T, condition game.Condition) {
				if condition.AnyPlayerLifeAtMost != 13 {
					t.Fatalf("AnyPlayerLifeAtMost = %d, want 13", condition.AnyPlayerLifeAtMost)
				}
			},
		},
		{
			name:      "opponent count",
			condition: "unless you have two or more opponents",
			assert: func(t *testing.T, condition game.Condition) {
				if condition.OpponentCountAtLeast != 2 {
					t.Fatalf("OpponentCountAtLeast = %d, want 2", condition.OpponentCountAtLeast)
				}
			},
		},
		{
			name:      "one opponent land count",
			condition: "unless an opponent controls two or more lands",
			assert: func(t *testing.T, condition game.Condition) {
				if !condition.AnyOpponentControls.Exists ||
					condition.AnyOpponentControls.Val.MinCount != 2 {
					t.Fatalf("AnyOpponentControls = %+v, want two lands", condition.AnyOpponentControls)
				}
			},
		},
		{
			name:      "collective opponent land count",
			condition: "unless your opponents control eight or more lands",
			assert: func(t *testing.T, condition game.Condition) {
				if !condition.OpponentsControl.Exists ||
					condition.OpponentsControl.Val.MinCount != 8 {
					t.Fatalf("OpponentsControl = %+v, want eight lands", condition.OpponentsControl)
				}
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Land",
				Layout:     "normal",
				TypeLine:   "Land",
				OracleText: "This land enters tapped " + test.condition + ".",
			})
			condition := face.ReplacementAbilities[0].Replacement.Condition.Val
			if !condition.Negate {
				t.Fatal("unless condition was not negated")
			}
			test.assert(t, condition)
		})
	}
}

func TestLowerOptionalEntryPayments(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		oracleText string
		assert     func(*testing.T, game.ResolutionPayment)
	}{
		{
			name:       "pay life",
			oracleText: "As this land enters, you may pay 2 life. If you don't, it enters tapped.",
			assert: func(t *testing.T, payment game.ResolutionPayment) {
				if len(payment.AdditionalCosts) != 1 ||
					payment.AdditionalCosts[0].Kind != cost.AdditionalPayLife ||
					payment.AdditionalCosts[0].Amount != 2 {
					t.Fatalf("payment = %+v, want pay 2 life", payment)
				}
			},
		},
		{
			// The dual-land cycle also includes a three-life variant; the life
			// amount is read from the parsed clause rather than fixed at two.
			name:       "pay three life",
			oracleText: "As this land enters, you may pay 3 life. If you don't, it enters tapped.",
			assert: func(t *testing.T, payment game.ResolutionPayment) {
				if payment.Prompt != "Pay 3 life?" {
					t.Fatalf("prompt = %q, want %q", payment.Prompt, "Pay 3 life?")
				}
				if len(payment.AdditionalCosts) != 1 ||
					payment.AdditionalCosts[0].Kind != cost.AdditionalPayLife ||
					payment.AdditionalCosts[0].Amount != 3 {
					t.Fatalf("payment = %+v, want pay 3 life", payment)
				}
			},
		},
		{
			name:       "reveal land subtype",
			oracleText: "As this land enters, you may reveal a Mountain or Forest card from your hand. If you don't, this land enters tapped.",
			assert: func(t *testing.T, payment game.ResolutionPayment) {
				if len(payment.AdditionalCosts) != 1 {
					t.Fatalf("payment = %+v, want one reveal cost", payment)
				}
				additional := payment.AdditionalCosts[0]
				if additional.Kind != cost.AdditionalReveal ||
					additional.Source != zone.Hand ||
					additional.SubtypesAny != (cost.SubtypeSet{types.Mountain, types.Forest}) {
					t.Fatalf("additional cost = %+v, want Mountain-or-Forest reveal from hand", additional)
				}
			},
		},
		{
			name:       "reveal creature subtype",
			oracleText: "As this land enters, you may reveal a Giant card from your hand. If you don't, this land enters tapped.",
			assert: func(t *testing.T, payment game.ResolutionPayment) {
				if len(payment.AdditionalCosts) != 1 ||
					payment.AdditionalCosts[0].SubtypesAny != (cost.SubtypeSet{types.Giant}) {
					t.Fatalf("payment = %+v, want Giant reveal", payment)
				}
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Land",
				Layout:     "normal",
				TypeLine:   "Land",
				OracleText: test.oracleText,
			})
			if len(face.ReplacementAbilities) != 1 ||
				!face.ReplacementAbilities[0].UnlessPaid.Exists {
				t.Fatalf("replacement abilities = %+v, want one paid replacement", face.ReplacementAbilities)
			}
			test.assert(t, face.ReplacementAbilities[0].UnlessPaid.Val)
		})
	}
}

func TestLowerReminderManaAbilitySingleColor(t *testing.T) {
	t.Parallel()
	// Basic lands express their mana ability as a parenthesized reminder.
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Forest",
		Layout:     "normal",
		TypeLine:   "Basic Land — Forest",
		OracleText: "({T}: Add {G}.)",
	})
	if len(face.ManaAbilities) != 1 {
		t.Fatalf("got %d mana abilities, want 1", len(face.ManaAbilities))
	}
	mode := face.ManaAbilities[0].Content.Modes[0]
	if len(mode.Sequence) != 1 {
		t.Fatalf("got %d instructions, want 1", len(mode.Sequence))
	}
	addMana, ok := mode.Sequence[0].Primitive.(game.AddMana)
	if !ok {
		t.Fatalf("primitive = %T, want game.AddMana", mode.Sequence[0].Primitive)
	}
	if addMana.ManaColor != mana.G {
		t.Fatalf("mana color = %q, want mana.G", addMana.ManaColor)
	}
}

func TestLowerReminderManaAbilityChoice(t *testing.T) {
	t.Parallel()
	// Dual lands express their mana ability as a parenthesized reminder.
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Dual",
		Layout:     "normal",
		TypeLine:   "Land — Mountain Forest",
		OracleText: "({T}: Add {R} or {G}.)",
	})
	if len(face.ManaAbilities) != 1 {
		t.Fatalf("got %d mana abilities, want 1", len(face.ManaAbilities))
	}
	mode := face.ManaAbilities[0].Content.Modes[0]
	choose, ok := mode.Sequence[0].Primitive.(game.Choose)
	if !ok {
		t.Fatalf("primitive = %T, want game.Choose", mode.Sequence[0].Primitive)
	}
	if choose.Choice.Kind != game.ResolutionChoiceMana {
		t.Fatalf("choice kind = %v, want ResolutionChoiceMana", choose.Choice.Kind)
	}
	if len(choose.Choice.Colors) != 2 {
		t.Fatalf("choice colors = %#v, want two colors", choose.Choice.Colors)
	}
}

func TestLowerNonManaHybridReminderDoesNotBlockCard(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Hybrid",
		Layout:     "normal",
		TypeLine:   "Creature — Soldier",
		OracleText: "({R/W} can be paid with either {R} or {W}.)\nFirst strike",
		Power:      new("1"),
		Toughness:  new("1"),
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v, want none", diagnostics)
	}
	if source == "" {
		t.Fatal("expected generated source")
	}
}

func TestLowerNonManaReminderDoesNotBlockCard(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Card",
		Layout:     "normal",
		TypeLine:   "Creature — Human",
		OracleText: "(This creature can block as though it had flying.)\nFlying",
		Power:      new("1"),
		Toughness:  new("1"),
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v, want none", diagnostics)
	}
	if source == "" {
		t.Fatal("expected generated source")
	}
}

func TestLowerAbilityWordDoesNotBlockSupportedKeyword(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Threshold",
		Layout:     "normal",
		TypeLine:   "Creature — Human",
		OracleText: "Threshold — Flying",
		Power:      new("1"),
		Toughness:  new("1"),
	}, "t")
	if err != nil {
		t.Fatal(err)
	}

	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v, want none", diagnostics)
	}
	if source == "" {
		t.Fatal("expected generated source")
	}
}

func TestLowerAbilityWordConditions(t *testing.T) {
	tests := []struct {
		name       string
		cardName   string
		typeLine   string
		oracleText string
		wants      []string
	}{
		{"threshold static", "Threshold Bear", "Creature — Bear", "Threshold — This creature gets +2/+2 as long as there are seven or more cards in your graveyard.", []string{"ControllerGraveyardCardCountAtLeast: 7"}},
		{"delirium static", "Delirium Bear", "Creature — Bear", "Delirium — This creature gets +1/+1 and has menace as long as there are four or more card types among cards in your graveyard.", []string{"ControllerGraveyardCardTypeCountAtLeast: 4", "AffectedSource: true"}},
		{"domain static", "Domain Bear", "Creature — Bear", "Domain — This creature gets +1/+1 for each basic land type among lands you control.", []string{"PowerDeltaDynamic: opt.Val(game.DynamicAmount{", "ToughnessDeltaDynamic: opt.Val(game.DynamicAmount{", "game.DynamicAmountControllerBasicLandTypeCount"}},
		{"domain spell", "Tribal Flames", "Sorcery", "Domain — Tribal Flames deals X damage to any target, where X is the number of basic land types among lands you control.", []string{"game.DynamicAmountControllerBasicLandTypeCount"}},
		{"metalcraft trigger", "Metalcraft Bear", "Creature — Bear", "Metalcraft — When this creature enters, if you control three or more artifacts, draw a card.", []string{"InterveningCondition: opt.Val(game.Condition{", "MinCount:"}},
		{"hellbent activation", "Hellbent Bear", "Creature — Bear", "Hellbent — {1}: Draw a card. Activate only if you have no cards in hand.", []string{"ActivationCondition: opt.Val(game.Condition{", "ControllerHandEmpty: true"}},
		{"ferocious activation", "Ferocious Bear", "Creature — Bear", "Ferocious — {1}: Draw a card. Activate only if you control a creature with power 4 or greater.", []string{"ActivationCondition: opt.Val(game.Condition{", "Value: 4"}},
		{"coven trigger", "Coven Bear", "Creature — Bear", "Coven — At the beginning of combat on your turn, if you control three or more creatures with different powers, draw a card.", []string{"InterveningCondition: opt.Val(game.Condition{", "ControllerCreaturePowerDiversityAtLeast: 3"}},
		{"morbid trigger", "Morbid Bear", "Creature — Bear", "Morbid — At the beginning of your end step, if a creature died this turn, put a +1/+1 counter on this creature.", []string{"InterveningCondition: opt.Val(game.Condition{", "game.EventPermanentDied", "Window: game.EventHistoryCurrentTurn"}},
		{"survival trigger", "Survival Bear", "Creature — Bear", "Survival — At the beginning of your second main phase, if this creature is tapped, you gain 2 life.", []string{"InterveningCondition: opt.Val(game.Condition{", "Tapped: game.TriTrue"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			card := &ScryfallCard{
				Name:       test.cardName,
				Layout:     "normal",
				TypeLine:   test.typeLine,
				OracleText: test.oracleText,
			}
			if strings.HasPrefix(test.typeLine, "Creature") {
				card.Power = new("2")
				card.Toughness = new("2")
			}
			source, diagnostics, err := GenerateExecutableCardSource(card, "t")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			if source == "" {
				t.Fatal("expected generated source")
			}
			for _, want := range test.wants {
				if !strings.Contains(source, want) {
					t.Fatalf("source missing %q:\n%s", want, source)
				}
			}
		})
	}
}

func TestLowerAbilityWordConditionsFailClosed(t *testing.T) {
	tests := []string{
		"Threshold — This creature gets +2/+2 as long as there are six or more creature cards in your graveyard.",
		"Delirium — This creature gets +2/+2 as long as there are three or more card types among cards in an opponent's graveyard.",
		"Metalcraft — This creature gets +2/+2 as long as you control two or more artifacts with banding.",
		"Hellbent — {1}: Draw a card. Activate only if you have one or fewer cards in hand.",
		"Ferocious — {1}: Draw a card. Activate only if you control a creature with power 3 or less.",
		"Coven — At the beginning of combat on your turn, if you control three or more creatures with the same power, draw a card.",
	}
	for _, oracleText := range tests {
		t.Run(oracleText, func(t *testing.T) {
			source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
				Name:       "Fail Closed Bear",
				Layout:     "normal",
				TypeLine:   "Creature — Bear",
				OracleText: oracleText,
				Power:      new("2"),
				Toughness:  new("2"),
			}, "t")
			if err != nil {
				t.Fatal(err)
			}
			if source != "" || len(diagnostics) == 0 {
				t.Fatalf("source = %q, diagnostics = %#v", source, diagnostics)
			}
		})
	}
}

func TestLowerActivationConditionPermanentSelections(t *testing.T) {
	tests := []struct {
		name       string
		typeLine   string
		oracleText string
		wants      []string
	}{
		{
			name:       "color permanents",
			typeLine:   "Land — Forest",
			oracleText: "{G}, {T}: You gain 1 life. Activate only if you control two or more green permanents.",
			wants:      []string{"ActivationCondition: opt.Val(game.Condition{", "ControlsMatching: opt.Val(game.SelectionCount{", "ColorsAny: []color.Color{color.Green}", "MinCount:  2"},
		},
		{
			name:       "snow permanents",
			typeLine:   "Creature — Human Wizard",
			oracleText: "{2}, {T}: Return target permanent to its owner's hand. Activate only if you control four or more snow permanents.",
			wants:      []string{"ActivationCondition: opt.Val(game.Condition{", "Supertypes: []types.Super{types.Snow}", "MinCount:  4"},
		},
		{
			name:       "legendary creature",
			typeLine:   "Legendary Land",
			oracleText: "{1}{U}, {T}: Scry 2. Activate only if you control a legendary creature.",
			wants:      []string{"ActivationCondition: opt.Val(game.Condition{", "ControlsMatching: opt.Val(game.SelectionCount{", "RequiredTypes: []types.Card{types.Creature}", "Supertypes: []types.Super{types.Legendary}"},
		},
		{
			name:       "creatures total power",
			typeLine:   "Creature — Bear",
			oracleText: "{1}{G}: This creature gets +1/+1 until end of turn. Activate only if creatures you control have total power 8 or greater.",
			wants: []string{
				"ActivationCondition: opt.Val(game.Condition{",
				"ControlsMatching: opt.Val(game.SelectionCount{",
				"Selection:  game.Selection{RequiredTypes: []types.Card{types.Creature}}",
				"TotalPower: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 8})",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			card := &ScryfallCard{
				Name:       "Cond " + test.name,
				Layout:     "normal",
				TypeLine:   test.typeLine,
				OracleText: test.oracleText,
			}
			if strings.HasPrefix(test.typeLine, "Creature") {
				card.Power = new("2")
				card.Toughness = new("2")
			}
			source, diagnostics, err := GenerateExecutableCardSource(card, "t")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			for _, want := range test.wants {
				if !strings.Contains(source, want) {
					t.Fatalf("source missing %q:\n%s", want, source)
				}
			}
		})
	}
}

func TestLowerActivationConditionPermanentSelectionsFailClosed(t *testing.T) {
	// These trailing activation conditions are one normalization away from a
	// supported permanent selection but use a predicate the condition model does
	// not represent, so the whole card must stay unsupported.
	tests := []string{
		"{1}: Draw a card. Activate only if you control a world enchantment.",
	}
	for _, oracleText := range tests {
		t.Run(oracleText, func(t *testing.T) {
			source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
				Name:       "Fail Closed Bear",
				Layout:     "normal",
				TypeLine:   "Creature — Bear",
				OracleText: oracleText,
				Power:      new("2"),
				Toughness:  new("2"),
			}, "t")
			if err != nil {
				t.Fatal(err)
			}
			if source != "" || len(diagnostics) == 0 {
				t.Fatalf("source = %q, diagnostics = %#v", source, diagnostics)
			}
		})
	}
}

func TestLowerAbilityWordSurfacesActualUnsupportedKeyword(t *testing.T) {
	t.Parallel()
	_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Threshold",
		Layout:     "normal",
		TypeLine:   "Creature — Human",
		OracleText: "Threshold — Protection from everything",
		Power:      new("1"),
		Toughness:  new("1"),
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) == 0 {
		t.Fatal("expected unsupported keyword diagnostic")
	}
	for _, diagnostic := range diagnostics {
		if diagnostic.Summary == "unsupported ability word" {
			t.Fatalf("diagnostics = %#v, want actual unsupported keyword diagnostic", diagnostics)
		}
	}
}

func TestLowerUnknownEmDashHeaderRemainsUnsupported(t *testing.T) {
	t.Parallel()
	_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Ticketed",
		Layout:     "normal",
		TypeLine:   "Creature — Human",
		OracleText: "{TK}{TK} — Menace",
		Power:      new("1"),
		Toughness:  new("1"),
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) == 0 {
		t.Fatal("expected unknown em-dash header to remain unsupported")
	}
}
