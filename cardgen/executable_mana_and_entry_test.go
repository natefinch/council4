package cardgen

import (
	"strings"
	"testing"
)

func TestGenerateExecutableCardSourceTapManaAbility(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Mana Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		OracleText: "{T}: Add {G}.",
		Power:      new("1"),
		Toughness:  new("1"),
	}

	source, diagnostics, err := GenerateExecutableCardSource(card, "m")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		`"github.com/natefinch/council4/mtg/game/mana"`,
		"ManaAbilities: []game.ManaAbility",
		"game.TapManaAbility(mana.G)",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceManaCostActivatedAbility(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Tome",
		Layout:     "normal",
		TypeLine:   "Artifact",
		OracleText: "{1}: Draw a card.",
	}

	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"ActivatedAbilities: []game.ActivatedAbility",
		"ManaCost:",
		"opt.Val(cost.Mana{cost.O(1)})",
		"ZoneOfFunction: zone.Battlefield",
		"game.Draw",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceTapCostActivatedAbility(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Icy",
		Layout:     "normal",
		TypeLine:   "Artifact",
		OracleText: "{T}: Tap target creature.",
	}

	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"AdditionalCosts: cost.Tap",
		"ZoneOfFunction:",
		"zone.Battlefield",
		"game.TargetPermanentReference(0)",
		"Primitive: game.Tap",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceManaAndTapCostActivatedAbility(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Engine",
		Layout:     "normal",
		TypeLine:   "Artifact",
		OracleText: "{2}, {T}: Draw a card.",
	}

	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"opt.Val(cost.Mana{cost.O(2)})",
		"AdditionalCosts: cost.Tap",
		"game.Draw",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceNonManaActivatedCosts(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:     "Test Engine",
		Layout:   "normal",
		TypeLine: "Artifact",
		OracleText: "{2}, {T}, Sacrifice a creature: Draw a card.\n" +
			"Discard two creature cards: Draw a card.\n" +
			"Pay 2 life: Draw a card.\n" +
			"Exile this artifact: Draw a card.\n" +
			"Exile a creature card from your graveyard: Draw a card.\n" +
			"{Q}: Draw a card.\n" +
			"Remove a charge counter from this artifact: Draw a card.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"cost.AdditionalSacrifice",
		"MatchPermanentType: true",
		"PermanentType:",
		"cost.AdditionalDiscard",
		"Amount: 2",
		"MatchCardType: true",
		"CardType:",
		"types.Creature",
		"zone.Hand",
		"cost.AdditionalPayLife",
		"cost.AdditionalExileSource",
		"zone.Battlefield",
		"cost.AdditionalExile",
		"zone.Graveyard",
		"cost.AdditionalUntap",
		"cost.AdditionalRemoveCounter",
		"counter.Charge",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceRejectsUnsupportedActivatedCost(t *testing.T) {
	t.Parallel()
	for _, oracleText := range []string{
		"Remove a +1/+1 counter from target creature: Draw a card.",
		"Sacrifice a nontoken creature: Draw a card.",
		"Discard a nonblack card: Draw a card.",
		"Discard a permanent card: Draw a card.",
		"Exile a card: Draw a card.",
	} {
		t.Run(oracleText, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{
				Name:       "Test Altar",
				Layout:     "normal",
				TypeLine:   "Artifact",
				OracleText: oracleText,
			}

			source, diagnostics, err := GenerateExecutableCardSource(card, "t")
			if err != nil {
				t.Fatal(err)
			}
			if source != "" {
				t.Fatalf("source = %q, want no partial card", source)
			}
			if len(diagnostics) != 1 || diagnostics[0].Summary != "unsupported activation cost" {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
		})
	}
}

func TestGenerateExecutableCardSourceEntersTapped(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Land",
		Layout:     "normal",
		TypeLine:   "Land",
		OracleText: "This land enters tapped.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"ReplacementAbilities: []game.ReplacementAbility",
		`game.EntersTappedReplacement("This land enters tapped.")`,
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
	if strings.Count(source, "game.EntersTappedReplacement") != 1 {
		t.Fatalf("source has duplicate enters-tapped replacements:\n%s", source)
	}
}

func TestGenerateExecutableCardSourceArtifactEntersTapped(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Relic",
		Layout:     "normal",
		TypeLine:   "Artifact",
		OracleText: "This artifact enters tapped.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if !strings.Contains(source, `game.EntersTappedReplacement("This artifact enters tapped.")`) {
		t.Fatalf("source missing artifact enters-tapped replacement:\n%s", source)
	}
}

func TestGenerateExecutableCardSourceEntersTappedUnlessTwoBasicLands(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Vista",
		Layout:     "normal",
		TypeLine:   "Land — Forest Plains",
		OracleText: "This land enters tapped unless you control two or more basic lands.",
	}

	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	// gofmt aligns PermanentFilter fields; check value literals rather than
	// aligned field assignments to stay resilient to column-width changes.
	for _, wanted := range []string{
		"game.EntersTappedIfReplacement",
		"Negate: true",
		"[]types.Card{types.Land}",
		"[]types.Super{types.Basic}",
		"MinCount:",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceEntersTappedUnlessGeneralizedLandConditions(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		oracleText string
		wanted     []string
	}{
		{
			name:       "singular basic land",
			oracleText: "This land enters tapped unless you control a basic land.",
			wanted: []string{
				"game.EntersTappedIfReplacement",
				"Negate: true",
				"[]types.Card{types.Land}",
				"[]types.Super{types.Basic}",
			},
		},
		{
			name:       "generic land count",
			oracleText: "This land enters tapped unless you control two or more lands.",
			wanted: []string{
				"game.EntersTappedIfReplacement",
				"Negate: true",
				"[]types.Card{types.Land}",
				"MinCount:",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
				Name:       "Test Vista",
				Layout:     "normal",
				TypeLine:   "Land",
				OracleText: tt.oracleText,
			}, "t")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			for _, wanted := range tt.wanted {
				if !strings.Contains(source, wanted) {
					t.Fatalf("source missing %q:\n%s", wanted, source)
				}
			}
		})
	}
}

func TestGenerateExecutableCardSourceLifeAndOpponentEntersTappedConditions(t *testing.T) {
	t.Parallel()
	tests := []struct {
		condition string
		want      string
	}{
		{condition: "unless you have 10 or more life", want: "ControllerLifeAtLeast: 10"},
		{condition: "unless a player has 13 or less life", want: "AnyPlayerLifeAtMost: 13"},
		{condition: "unless you have two or more opponents", want: "OpponentCountAtLeast: 2"},
		{condition: "unless an opponent controls two or more lands", want: "AnyOpponentControls: opt.Val(game.SelectionCount{"},
		{condition: "unless your opponents control eight or more lands", want: "OpponentsControl: opt.Val(game.SelectionCount{"},
	}
	for _, test := range tests {
		t.Run(test.condition, func(t *testing.T) {
			t.Parallel()
			source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
				Name:       "Test Land",
				Layout:     "normal",
				TypeLine:   "Land",
				OracleText: "This land enters tapped " + test.condition + ".",
			}, "t")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			if !strings.Contains(source, test.want) {
				t.Fatalf("source missing %q:\n%s", test.want, source)
			}
		})
	}
}

// TestGenerateExecutableCardSourceRejectsUnsupportedConditionalEntersTapped
// verifies that near-miss conditions outside the supported wording family are
// rejected. Supported: "unless you control two or more basic lands".
func TestGenerateExecutableCardSourceOptionalEntryPayments(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		oracleText string
		wants      []string
	}{
		{
			name:       "pay life",
			oracleText: "As this land enters, you may pay 2 life. If you don't, it enters tapped.",
			wants: []string{
				"game.EntersTappedUnlessPaidReplacement",
				"cost.AdditionalPayLife",
				"Amount: 2",
			},
		},
		{
			name:       "reveal subtypes",
			oracleText: "As this land enters, you may reveal a Mountain or Forest card from your hand. If you don't, this land enters tapped.",
			wants: []string{
				"game.EntersTappedUnlessPaidReplacement",
				"cost.AdditionalReveal",
				"SubtypesAny: cost.SubtypeSet{types.Mountain, types.Forest}",
				"zone.Hand",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
				Name:       "Test Land",
				Layout:     "normal",
				TypeLine:   "Land",
				OracleText: test.oracleText,
			}, "t")
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

func TestGenerateExecutableCardSourceRejectsUnsupportedConditionalEntersTapped(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		oracleText string
	}{
		{
			name:       "unsupported keyword selection",
			oracleText: "This land enters tapped unless you control two or more artifacts with flying.",
		},
		{
			name:       "unsupported keyword creature selection",
			oracleText: "This land enters tapped unless you control a creature with flying.",
		},
		{
			name:       "if instead of unless",
			oracleText: "This land enters tapped if you control two or more basic lands.",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{
				Name:       "Test Vista",
				Layout:     "normal",
				TypeLine:   "Land",
				OracleText: tt.oracleText,
			}
			source, diagnostics, err := GenerateExecutableCardSource(card, "t")
			if err != nil {
				t.Fatal(err)
			}
			if source != "" {
				t.Fatalf("source = %q, want no partial card for near-miss wording", source)
			}
			if len(diagnostics) == 0 {
				t.Fatal("expected diagnostics for unsupported condition, got none")
			}
		})
	}
}

func TestGenerateExecutableCardSourceBackFaceEntersTappedOnce(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:          "Test Bear // Test Vale",
		Layout:        "modal_dfc",
		ColorIdentity: []string{"G"},
		CardFaces: []ScryfallCardFace{
			{
				Name:      "Test Bear",
				ManaCost:  "{2}{G}",
				TypeLine:  "Creature — Bear",
				Power:     new("2"),
				Toughness: new("2"),
			},
			{
				Name:       "Test Vale",
				TypeLine:   "Land",
				OracleText: "This land enters tapped.",
			},
		},
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if strings.Count(source, "game.EntersTappedReplacement") != 1 {
		t.Fatalf("source has duplicate enters-tapped replacements:\n%s", source)
	}
}

func TestGenerateExecutableCardSourceAnyColorTapMana(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Rock",
		Layout:     "normal",
		TypeLine:   "Artifact",
		OracleText: "{T}: Add one mana of any color.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"game.TapManaChoiceAbility(mana.W, mana.U, mana.B, mana.R, mana.G)",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceColorChoiceTapMana(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Land",
		Layout:     "normal",
		TypeLine:   "Land",
		OracleText: "{T}: Add {R} or {G}.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if !strings.Contains(source, "game.TapManaChoiceAbility(mana.R, mana.G)") {
		t.Fatalf("source has wrong mana choices:\n%s", source)
	}
}

func TestGenerateExecutableCardSourceThreeColorTapMana(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Land",
		Layout:     "normal",
		TypeLine:   "Land",
		OracleText: "{T}: Add {W}, {U}, or {B}.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if !strings.Contains(source, "game.TapManaChoiceAbility(mana.W, mana.U, mana.B)") {
		t.Fatalf("source has wrong mana choices:\n%s", source)
	}
}

func TestGenerateExecutableCardSourceManaWard(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		OracleText: "Ward {2}",
		Power:      new("2"),
		Toughness:  new("2"),
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"StaticAbilities: []game.StaticAbility",
		"game.WardStaticAbility",
		"cost.Mana",
		"cost.O(2)",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}
