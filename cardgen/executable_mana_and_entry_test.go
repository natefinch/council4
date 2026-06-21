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

func TestGenerateExecutableCardSourceOptionalEntryDiscardElseGraveyard(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Mox Diamond",
		Layout:     "normal",
		TypeLine:   "Artifact",
		OracleText: "If this artifact would enter, you may discard a land card instead. If you do, put this artifact onto the battlefield. If you don't, put it into its owner's graveyard.\n{T}: Add one mana of any color.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"game.EntersUnlessPaidElseZoneReplacement(",
		"cost.AdditionalDiscard,",
		"types.Land,",
		"zone.Hand,",
		"zone.Graveyard),",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
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
			name:       "pay three life",
			oracleText: "As this land enters, you may pay 3 life. If you don't, it enters tapped.",
			wants: []string{
				"game.EntersTappedUnlessPaidReplacement",
				"cost.AdditionalPayLife",
				"Amount: 3",
				`Prompt: "Pay 3 life?"`,
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
			oracleText: "This land enters tapped unless you control two or more artifacts with banding.",
		},
		{
			name:       "unsupported keyword creature selection",
			oracleText: "This land enters tapped unless you control a creature with deathtouch and flying.",
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

// TestGenerateExecutableCardSourceAnyOneColorCountTapMana covers the "Add <N>
// mana of any one color" family (Gilded Lotus): N mana, all of one chosen color.
func TestGenerateExecutableCardSourceAnyOneColorCountTapMana(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name       string
		oracleText string
		want       string
	}{
		{
			name:       "Gilded Lotus",
			oracleText: "{T}: Add three mana of any one color.",
			want:       "game.TapManaChoiceCountAbility(\"{T}: Add three mana of any one color.\", 3, mana.W, mana.U, mana.B, mana.R, mana.G)",
		},
		{
			name:       "Two Lotus",
			oracleText: "{T}: Add two mana of any one color.",
			want:       "game.TapManaChoiceCountAbility(\"{T}: Add two mana of any one color.\", 2, mana.W, mana.U, mana.B, mana.R, mana.G)",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{
				Name:       tc.name,
				Layout:     "normal",
				TypeLine:   "Artifact",
				OracleText: tc.oracleText,
			}
			source, diagnostics, err := GenerateExecutableCardSource(card, "t")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			if !strings.Contains(source, tc.want) {
				t.Fatalf("source missing %q:\n%s", tc.want, source)
			}
		})
	}
}

func TestGenerateExecutableChromaticLanternStaticManaGrant(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Chromatic Lantern",
		Layout:     "normal",
		ManaCost:   "{3}",
		TypeLine:   "Artifact",
		OracleText: "Lands you control have \"{T}: Add one mana of any color.\"\n{T}: Add one mana of any color.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"game.ObjectControlledGroup(",
		"game.SourcePermanentReference()",
		"RequiredTypes: []types.Card{types.Land}",
		"AddAbilities: []game.Ability",
		"game.TapManaChoiceAbility(mana.W, mana.U, mana.B, mana.R, mana.G)",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceCommanderIdentityTapMana(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Arcane Signet",
		Layout:     "normal",
		TypeLine:   "Artifact",
		OracleText: "{T}: Add one mana of any color in your commander's color identity.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if !strings.Contains(source, "game.TapManaCommanderIdentityAbility()") {
		t.Fatalf("source missing commander-identity mana ability:\n%s", source)
	}
}

// TestGenerateExecutableCardSourceUnscopedAnyColorTapManaFailsClosed asserts the
// commander-identity recognition does not over-match other "any color" wordings.
// "any one color" adds three mana of a single freely-chosen color, which lowers
// to the choose-then-add count ability rather than a commander-identity ability.
func TestGenerateExecutableCardSourceUnscopedAnyColorTapManaFailsClosed(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Rock",
		Layout:     "normal",
		TypeLine:   "Artifact",
		OracleText: "{T}: Add three mana of any one color.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if strings.Contains(source, "game.TapManaCommanderIdentityAbility()") {
		t.Fatalf("unscoped any-color wording wrongly lowered to commander identity:\n%s", source)
	}
}

// TestGenerateExecutableCardSourceAmongControlledColorsTapMana covers the "Add
// one mana of any color among <permanents> you control" family (Mox Amber, Plaza
// of Heroes). The body lowers to the among-controlled-colors mana ability whose
// choosable colors are recomputed at resolution from the matching permanents'
// colors. The permanent filter stays generic over the group: a disjunctive type
// union with a shared supertype (Mox Amber) and a bare supertype permanent
// group (Plaza of Heroes).
func TestGenerateExecutableCardSourceAmongControlledColorsTapMana(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name       string
		typeLine   string
		oracleText string
		want       string
	}{
		{
			"Mox Amber", "Legendary Artifact",
			"{T}: Add one mana of any color among legendary creatures and planeswalkers you control.",
			"game.TapManaAmongControlledColorsAbility(\"{T}: Add one mana of any color among legendary creatures and planeswalkers you control.\", game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Planeswalker}, Supertypes: []types.Super{types.Legendary}, Controller: game.ControllerYou})",
		},
		{
			"Mana Rock", "Artifact",
			"{T}: Add one mana of any color among creatures you control.",
			"game.TapManaAmongControlledColorsAbility(\"{T}: Add one mana of any color among creatures you control.\", game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerYou})",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{Name: tc.name, Layout: "normal", TypeLine: tc.typeLine, OracleText: tc.oracleText}
			source, diagnostics, err := GenerateExecutableCardSource(card, "t")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			if !strings.Contains(source, tc.want) {
				t.Fatalf("source missing %q:\n%s", tc.want, source)
			}
		})
	}
}

// TestGenerateExecutableCardSourceAmongControlledColorsFailsClosed asserts the
// among-controlled-colors recognition does not over-match related wordings: a
// bare "permanents you control" with no narrowing predicate, an opponent-
// controlled group, and the unrelated "among" subjects must fail closed rather
// than lower to a mislabeled ability.
func TestGenerateExecutableCardSourceAmongControlledColorsFailsClosed(t *testing.T) {
	t.Parallel()
	for _, oracleText := range []string{
		"{T}: Add one mana of any color among permanents you control.",
		"{T}: Add one mana of any color among creatures an opponent controls.",
		"{T}: Add two mana of any color among creatures you control.",
	} {
		card := &ScryfallCard{Name: "Test Source", Layout: "normal", TypeLine: "Artifact", OracleText: oracleText}
		source, diagnostics, err := GenerateExecutableCardSource(card, "t")
		if err != nil {
			t.Fatal(err)
		}
		if len(diagnostics) == 0 {
			t.Fatalf("expected fail-closed diagnostic for %q, got source:\n%s", oracleText, source)
		}
		if strings.Contains(source, "TapManaAmongControlledColorsAbility") {
			t.Fatalf("unmodeled wording wrongly lowered to among-controlled-colors ability: %q", oracleText)
		}
	}
}

// TestGenerateExecutableCardSourceEachColorAmongControlledTapMana covers Bloom
// Tender's "For each color among permanents you control, add one mana of that
// color" body, which lowers to the each-color production over the controller's
// permanents. Unlike the among-controlled-colors choice ability, a bare
// "permanents you control" group is accepted because the whole board
// contributes its colors. The "Vivid —" ability word is stripped before the
// activated ability is lowered.
func TestGenerateExecutableCardSourceEachColorAmongControlledTapMana(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name       string
		oracleText string
		want       string
	}{
		{
			"Bloom Tender",
			"Vivid — {T}: For each color among permanents you control, add one mana of that color.",
			"game.TapManaEachControlledColorAbility(\"Vivid — {T}: For each color among permanents you control, add one mana of that color.\", game.Selection{Controller: game.ControllerYou})",
		},
		{
			"Creature Group",
			"{T}: For each color among creatures you control, add one mana of that color.",
			"game.TapManaEachControlledColorAbility(\"{T}: For each color among creatures you control, add one mana of that color.\", game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerYou})",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{Name: tc.name, Layout: "normal", TypeLine: "Creature — Elf Druid", OracleText: tc.oracleText}
			source, diagnostics, err := GenerateExecutableCardSource(card, "t")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			if !strings.Contains(source, tc.want) {
				t.Fatalf("source missing %q:\n%s", tc.want, source)
			}
		})
	}
}

// TestGenerateExecutableCardSourceEachColorAmongControlledFailsClosed asserts the
// each-color recognition does not over-match related wordings: an opponent-
// controlled group and a non-"one mana" quantity must fail closed rather than
// lower to a mislabeled ability.
func TestGenerateExecutableCardSourceEachColorAmongControlledFailsClosed(t *testing.T) {
	t.Parallel()
	for _, oracleText := range []string{
		"{T}: For each color among creatures an opponent controls, add one mana of that color.",
		"{T}: For each color among permanents you control, add two mana of that color.",
		"{T}: For each color among monocolored permanents you control, add one mana of that color.",
	} {
		card := &ScryfallCard{Name: "Test Source", Layout: "normal", TypeLine: "Artifact", OracleText: oracleText}
		source, _, err := GenerateExecutableCardSource(card, "t")
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(source, "TapManaEachControlledColorAbility") {
			t.Fatalf("unmodeled wording wrongly lowered to each-color ability: %q", oracleText)
		}
	}
}

// TestGenerateExecutableCardSourceLandsProduceTapMana covers the Exotic Orchard,
// Reflecting Pool, and Fellwar Stone wordings, which lower to the dynamic
// "any color/type a land could produce" mana ability scoped to the matching
// player. The "any type" wording carries the colorless-inclusive flag.
func TestGenerateExecutableCardSourceLandsProduceTapMana(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name       string
		typeLine   string
		oracleText string
		want       string
	}{
		{"Exotic Orchard", "Land", "{T}: Add one mana of any color that a land an opponent controls could produce.", "game.TapManaLandsProduceAbility(game.PlayerOpponent, false)"},
		{"Fellwar Stone", "Artifact", "{T}: Add one mana of any color that a land an opponent controls could produce.", "game.TapManaLandsProduceAbility(game.PlayerOpponent, false)"},
		{"Harvester Druid", "Creature — Elf Druid", "{T}: Add one mana of any color that a land you control could produce.", "game.TapManaLandsProduceAbility(game.PlayerYou, false)"},
		{"Reflecting Pool", "Land", "{T}: Add one mana of any type that a land you control could produce.", "game.TapManaLandsProduceAbility(game.PlayerYou, true)"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{Name: tc.name, Layout: "normal", TypeLine: tc.typeLine, OracleText: tc.oracleText}
			source, diagnostics, err := GenerateExecutableCardSource(card, "t")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			if !strings.Contains(source, tc.want) {
				t.Fatalf("source missing %q:\n%s", tc.want, source)
			}
		})
	}
}

// TestGenerateExecutableCardSourceAnyPlayerTapManaThatPlayer covers the generic
// any-player tapped-for-mana doubler "Whenever a player taps a land for mana,
// that player adds an additional {G}": the trigger leaves the tapped land's
// controller unrestricted (RequireTappedForMana) and routes the produced mana to
// the player who tapped via EventPlayerReference. The opponent-scoped form
// restricts the tapped subject to an opponent.
func TestGenerateExecutableCardSourceAnyPlayerTapManaThatPlayer(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name       string
		oracleText string
	}{
		{"Any Player Doubler", "Whenever a player taps a land for mana, that player adds an additional {G}."},
		{"Opponent Doubler", "Whenever an opponent taps a land for mana, that player adds an additional {U}."},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{Name: tc.name, Layout: "normal", TypeLine: "Enchantment", OracleText: tc.oracleText}
			source, diagnostics, err := GenerateExecutableCardSource(card, "t")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			for _, want := range []string{"RequireTappedForMana: true", "game.EventPlayerReference()"} {
				if !strings.Contains(source, want) {
					t.Fatalf("source missing %q:\n%s", want, source)
				}
			}
		})
	}
}

// TestGenerateExecutableCardSourceTriggerLandProducedMana covers the Mirari's
// Wake / Zendikar Resurgent mana-doubler: "Whenever you tap a land for mana, add
// one mana of any type that land produced." It lowers to a tapped-for-mana
// trigger whose mana mirrors the produced type via the dedicated color source.
func TestGenerateExecutableCardSourceTriggerLandProducedMana(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Mirari's Wake",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		OracleText: "Creatures you control get +1/+1.\nWhenever you tap a land for mana, add one mana of any type that land produced.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"RequireTappedForMana: true",
		"game.ResolutionChoiceColorSourceTriggerLandProduced",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("source missing %q:\n%s", want, source)
		}
	}
}

// TestGenerateExecutableCardSourceLandsProduceFailsClosed asserts the
// (basic-land, Gate, sacrificed-land, plural-quantity, non-land, and unscoped
// "a player controls" variants), which must fail closed.
func TestGenerateExecutableCardSourceLandsProduceFailsClosed(t *testing.T) {
	t.Parallel()
	for _, oracleText := range []string{
		"{T}: Add one mana of any color that a basic land you control could produce.",
		"{T}: Add one mana of any color that a Gate you control could produce.",
		"{T}: Add one mana of any type the sacrificed land could produce.",
		"{T}: Add two mana of any color that a land an opponent controls could produce.",
		"{T}: Add one mana of any color that a creature you control could produce.",
		"{T}: Add one mana of any color that a land a player controls could produce.",
	} {
		card := &ScryfallCard{Name: "Test Source", Layout: "normal", TypeLine: "Artifact", OracleText: oracleText}
		source, diagnostics, err := GenerateExecutableCardSource(card, "t")
		if err != nil {
			t.Fatal(err)
		}
		if len(diagnostics) == 0 {
			t.Fatalf("expected fail-closed diagnostic for %q, got source:\n%s", oracleText, source)
		}
		if strings.Contains(source, "TapManaLandsProduceAbility") {
			t.Fatalf("unmodeled wording wrongly lowered to lands-produce ability: %q", oracleText)
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

// TestGenerateExecutableCardSourceUtopiaSprawl covers the "As this Aura enters,
// choose a color." entry choice (named by the source's own subtype, so it
// surfaces no object reference) combined with a tapped-for-mana trigger that
// adds "one mana of the chosen color" to the tapped land's controller.
func TestGenerateExecutableCardSourceUtopiaSprawl(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:     "Utopia Sprawl",
		Layout:   "normal",
		ManaCost: "{G}",
		TypeLine: "Enchantment — Aura",
		OracleText: "Enchant Forest\n" +
			"As this Aura enters, choose a color.\n" +
			"Whenever enchanted Forest is tapped for mana, its controller adds an additional one mana of the chosen color.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "u")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"game.EntryColorChoiceReplacement(",
		"RequireTappedForMana: true",
		"EntryChoiceFrom: game.ChoiceKey(\"oracle-entry-color\")",
		"opt.Val(game.ObjectControllerReference(game.EventPermanentReference()))",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceSupertypeAnthemWithWard(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Flowering of the White Tree",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		OracleText: "Legendary creatures you control get +2/+1 and have ward {1}.\nNonlegendary creatures you control get +1/+1.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Supertypes: []types.Super{types.Legendary}",
		"ExcludedSupertype: types.Legendary",
		"game.WardStaticAbility(cost.Mana{cost.O(1)})",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
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
