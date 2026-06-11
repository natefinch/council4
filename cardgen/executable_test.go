package cardgen

import (
	"fmt"
	"strings"
	"testing"
)

func TestGenerateExecutableCardSourceKeywords(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Keyword Bear",
		Layout:     "normal",
		ManaCost:   "{1}{G}",
		TypeLine:   "Creature — Bear",
		OracleText: "Flying, vigilance",
		Power:      new("2"),
		Toughness:  new("2"),
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "k")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"game.FlyingStaticBody",
		"game.VigilanceStaticBody",
		"StaticAbilities: []game.StaticAbility",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
	if strings.Contains(source, "TODO") {
		t.Fatalf("executable source contains TODO:\n%s", source)
	}
}

func TestGenerateExecutableCardSourceSpellCastTrigger(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Blue Spell Watcher",
		Layout:     "normal",
		ManaCost:   "{2}{U}",
		TypeLine:   "Creature — Wizard",
		OracleText: "Whenever an opponent casts a blue spell, draw a card.",
		Colors:     []string{"U"},
		Power:      new("2"),
		Toughness:  new("2"),
	}

	source, diagnostics, err := GenerateExecutableCardSource(card, "b")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"game.EventSpellCast",
		"game.TriggerControllerOpponent",
		"CardSelection: game.Selection{ColorsAny: []color.Color{color.Blue}}",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceReadAhead(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Read Ahead Saga",
		Layout:     "saga",
		ManaCost:   "{2}{U}",
		TypeLine:   "Enchantment — Saga",
		OracleText: "Read ahead (Choose a chapter and start with that many lore counters. Add one after your draw step. Skipped chapters don't trigger.)\nI — Draw a card.\nII — Draw a card.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "r")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if !strings.Contains(source, "game.ReadAheadStaticBody") {
		t.Fatalf("source missing Read ahead static body:\n%s", source)
	}
}

func TestGenerateExecutableCardSourceDevoid(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:          "Colorless Bear",
		Layout:        "normal",
		ManaCost:      "{1}{R}",
		TypeLine:      "Creature — Bear",
		OracleText:    "Devoid (This card has no color.)",
		Colors:        []string{"R"},
		ColorIdentity: []string{"R"},
		Power:         new("2"),
		Toughness:     new("2"),
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "c")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"game.DevoidStaticBody",
		"ColorIdentity: color.NewIdentity(color.Red)",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
	if strings.Contains(source, "Colors:") {
		t.Fatalf("Devoid card face is not colorless:\n%s", source)
	}
}

func TestGenerateExecutableCardSourceRejectsNoncanonicalDevoid(t *testing.T) {
	t.Parallel()
	for _, oracleText := range []string{
		"Devoid",
		"Devoid (This card is colorless.)",
	} {
		t.Run(oracleText, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{
				Name:       "Almost Colorless Bear",
				Layout:     "normal",
				TypeLine:   "Creature — Bear",
				OracleText: oracleText,
				Power:      new("2"),
				Toughness:  new("2"),
			}
			source, diagnostics, err := GenerateExecutableCardSource(card, "a")
			if err != nil {
				t.Fatal(err)
			}
			if source != "" || len(diagnostics) == 0 {
				t.Fatalf("source = %q, diagnostics = %#v", source, diagnostics)
			}
			if got := diagnostics[0].Summary; got != "unsupported Devoid ability" {
				t.Fatalf("summary = %q", got)
			}
		})
	}
}

func TestGenerateExecutableCardSourceSelfCannotBlock(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Reluctant Bear",
		Layout:     "normal",
		ManaCost:   "{1}{G}",
		TypeLine:   "Creature — Bear",
		OracleText: "This creature can't block.",
		Power:      new("3"),
		Toughness:  new("3"),
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "r")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if !strings.Contains(source, "game.CantBlockStaticBody") {
		t.Fatalf("source missing cannot-block static body:\n%s", source)
	}
}

func TestGenerateExecutableCardSourceSelfCannotBeBlocked(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Elusive Bear",
		Layout:     "normal",
		ManaCost:   "{1}{U}",
		TypeLine:   "Creature — Bear",
		OracleText: "This creature can't be blocked.",
		Colors:     []string{"U"},
		Power:      new("2"),
		Toughness:  new("2"),
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "e")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if !strings.Contains(source, "game.CantBeBlockedStaticBody") {
		t.Fatalf("source missing cannot-be-blocked static body:\n%s", source)
	}
}

func TestGenerateExecutableCardSourceSelfMustAttack(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Reckless Bear",
		Layout:     "normal",
		ManaCost:   "{1}{R}",
		TypeLine:   "Creature — Bear",
		OracleText: "This creature attacks each combat if able.",
		Colors:     []string{"R"},
		Power:      new("3"),
		Toughness:  new("2"),
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "r")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if !strings.Contains(source, "game.MustAttackStaticBody") {
		t.Fatalf("source missing must-attack static body:\n%s", source)
	}
}

func TestGenerateExecutableCardSourceRejectsConditionalMustAttack(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Conditional Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		OracleText: "This creature attacks each combat if able unless you control an artifact.",
		Power:      new("3"),
		Toughness:  new("2"),
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "c")
	if err != nil {
		t.Fatal(err)
	}
	if source != "" || len(diagnostics) == 0 {
		t.Fatalf("source = %q, diagnostics = %#v", source, diagnostics)
	}
}

func TestGenerateExecutableCardSourceRejectsConditionalCannotBeBlocked(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Conditional Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		OracleText: "This creature can't be blocked as long as you control an artifact.",
		Power:      new("2"),
		Toughness:  new("2"),
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "c")
	if err != nil {
		t.Fatal(err)
	}
	if source != "" || len(diagnostics) == 0 {
		t.Fatalf("source = %q, diagnostics = %#v", source, diagnostics)
	}
}

func TestGenerateExecutableCardSourceSelfUncounterable(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Certain Doom",
		Layout:     "normal",
		ManaCost:   "{1}{B}",
		TypeLine:   "Sorcery",
		OracleText: "This spell can't be countered.\nDestroy target creature.",
		Colors:     []string{"B"},
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "c")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if !strings.Contains(source, "game.CantBeCounteredStaticBody") {
		t.Fatalf("source missing uncounterable static body:\n%s", source)
	}
}

func TestGenerateExecutableCardSourceRejectsNoncanonicalSelfUncounterable(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Certain Doom",
		Layout:     "normal",
		ManaCost:   "{1}{B}",
		TypeLine:   "Sorcery",
		OracleText: "Certain Doom can't be countered.\nDestroy target creature.",
		Colors:     []string{"B"},
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "c")
	if err != nil {
		t.Fatal(err)
	}
	if source != "" || len(diagnostics) == 0 {
		t.Fatalf("source = %q, diagnostics = %#v", source, diagnostics)
	}
}

func TestGenerateExecutableCardSourceRejectsConditionalCannotBlock(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Conditional Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		OracleText: "This creature can't block unless you control an artifact.",
		Power:      new("3"),
		Toughness:  new("3"),
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "c")
	if err != nil {
		t.Fatal(err)
	}
	if source != "" {
		t.Fatalf("source = %q, want no partial card", source)
	}
	if len(diagnostics) == 0 {
		t.Fatal("expected unsupported diagnostic")
	}
}

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
			if len(diagnostics) != 1 || diagnostics[0].Summary != "unsupported activated ability" {
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
			name:       "unless single basic land",
			oracleText: "This land enters tapped unless you control a basic land.",
		},
		{
			name:       "unless two or more non-basic lands",
			oracleText: "This land enters tapped unless you control two or more lands.",
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

func TestGenerateExecutableCardSourceCycling(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Card",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Cycling {1}{U} ({1}{U}, Discard this card: Draw a card.)",
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
		"game.CyclingActivatedAbility(cost.Mana",
		"cost.O(1)",
		"cost.U",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceCyclingTrigger(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Cycler",
		Layout:     "normal",
		TypeLine:   "Creature — Fox",
		ManaCost:   "{W}",
		Colors:     []string{"W"},
		OracleText: "Whenever you cycle another card, draw a card.",
		Power:      new("1"),
		Toughness:  new("1"),
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"TriggeredAbilities: []game.TriggeredAbility",
		"game.EventCycled",
		"game.TriggerPlayerYou",
		"ExcludeSelf: true",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceHandCyclingGrant(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Reformation",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		ManaCost:   "{1}{R}",
		Colors:     []string{"R"},
		OracleText: "Each land card in your hand has cycling {R}.",
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
		"game.RuleEffectGrantHandCardAbility",
		"game.PlayerYou",
		"RequiredTypes: []types.Card{types.Land}",
		"game.CyclingActivatedAbility(cost.Mana{cost.R})",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceEquip(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Equipment",
		Layout:     "normal",
		TypeLine:   "Artifact — Equipment",
		OracleText: "Equip {2}",
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
		"game.EquipActivatedAbility(cost.Mana{cost.O(2)})",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceEnchantCreature(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Aura",
		Layout:     "normal",
		TypeLine:   "Enchantment — Aura",
		OracleText: "Enchant creature",
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
		"game.EnchantStaticAbility(&game.TargetSpec{",
		"PermanentTypes: []types.Card{types.Creature}",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceProtectionFromColor(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		OracleText: "Protection from red",
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
		"game.ProtectionFromColorsStaticAbility(color.Red)",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceProtectionFromEverything(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Angel",
		Layout:     "normal",
		TypeLine:   "Creature — Angel",
		OracleText: "Flying\nProtection from everything",
		Power:      new("5"),
		Toughness:  new("5"),
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if !strings.Contains(source, "game.ProtectionFromEverythingStaticAbility()") {
		t.Fatalf("source missing ProtectionFromEverythingStaticAbility:\n%s", source)
	}
}

func TestGenerateExecutableCardSourceProtectionFromTypes(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		OracleText: "Protection from artifacts",
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
	if !strings.Contains(source, "game.ProtectionFromTypesStaticAbility(types.Artifact)") {
		t.Fatalf("source missing ProtectionFromTypesStaticAbility:\n%s", source)
	}
}

func TestGenerateExecutableCardSourceProtectionFromSubtypes(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Dragon Hunter",
		Layout:     "normal",
		TypeLine:   "Creature — Human Warrior",
		OracleText: "Protection from Dragons",
		Power:      new("2"),
		Toughness:  new("1"),
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if !strings.Contains(source, "game.ProtectionFromSubtypesStaticAbility(types.Dragon)") {
		t.Fatalf("source missing ProtectionFromSubtypesStaticAbility:\n%s", source)
	}
}

func TestGenerateExecutableCardSourceProtectionFromEachColor(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Etched Champion",
		Layout:     "normal",
		TypeLine:   "Artifact Creature — Soldier",
		OracleText: "Metalcraft — As long as you control three or more artifacts, this creature has protection from all colors.",
		Power:      new("2"),
		Toughness:  new("2"),
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %#v", diagnostics)
	}
	for _, want := range []string{
		"AffectedSource: true",
		"game.ProtectionFromEachColorStaticAbility()",
		"AddAbilities:",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}

func TestGenerateExecutableCardSourceProtectionGrantFromEnchant(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Aura",
		Layout:     "normal",
		TypeLine:   "Enchantment — Aura",
		OracleText: "Enchant creature\nEnchanted creature has protection from black.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"AddAbilities:",
		"game.ProtectionFromColorsStaticAbility(color.Black)",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceProtectionGrantWithSourcePTBuff(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Guardian",
		Layout:     "normal",
		TypeLine:   "Creature — Guardian",
		OracleText: "This creature gets +1/+1 and has protection from creatures.",
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
		"PowerDelta:",
		"ToughnessDelta:",
		"AddAbilities:",
		"game.ProtectionFromTypesStaticAbility(types.Creature)",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
	if got := strings.Count(source, "AffectedSource: true"); got != 2 {
		t.Fatalf("AffectedSource count = %d, want 2:\n%s", got, source)
	}
}

func TestGenerateExecutableCardSourceChosenColorProtectionFails(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Shield",
		Layout:     "normal",
		TypeLine:   "Enchantment — Aura",
		OracleText: "Enchant creature\nEnchanted creature has protection from the chosen color.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if source != "" {
		t.Fatalf("unexpected source for chosen-color protection:\n%s", source)
	}
	if len(diagnostics) == 0 {
		t.Fatal("expected unsupported diagnostic for chosen-color protection")
	}
}

func TestGenerateExecutableCardSourceFixedDamage(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Bolt",
		Layout:     "normal",
		ManaCost:   "{R}",
		TypeLine:   "Instant",
		OracleText: "Test Bolt deals 3 damage to any target.",
		Colors:     []string{"R"},
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"SpellAbility: opt.Val(game.Mode{",
		"MinTargets: 1",
		"MaxTargets: 1",
		`Constraint: "any target"`,
		"game.TargetAllowPermanent | game.TargetAllowPlayer",
		"Primitive: game.Damage",
		"game.Fixed(3)",
		"Recipient: game.AnyTargetDamageRecipient(0)",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceOrderedEffects(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Spell",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Destroy target artifact. Draw a card.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	destroy := strings.Index(source, "Primitive: game.Destroy")
	draw := strings.Index(source, "Primitive: game.Draw")
	if destroy < 0 || draw < 0 || destroy >= draw {
		t.Fatalf("instructions are not rendered in Oracle order:\n%s", source)
	}
}

func TestGenerateExecutableCardSourceOrderedEffectsWithMultipleTargets(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Spell",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Destroy target artifact. Test Spell deals 2 damage to any target.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Object: game.TargetPermanentReference(0)",
		"Recipient: game.AnyTargetDamageRecipient(1)",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceSurveil(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Surveil",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Surveil 2.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if !strings.Contains(source, "Primitive: game.Surveil") {
		t.Fatalf("source missing Surveil primitive:\n%s", source)
	}
}

func TestGenerateExecutableCardSourceInvestigate(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Investigate",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Investigate.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if !strings.Contains(source, "Primitive: game.Investigate") {
		t.Fatalf("source missing Investigate primitive:\n%s", source)
	}
}

func TestGenerateExecutableCardSourceProliferate(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Proliferate",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Proliferate.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if !strings.Contains(source, "Primitive: game.Proliferate{") ||
		!strings.Contains(source, "Amount: game.Fixed(1)") {
		t.Fatalf("source missing Proliferate primitive:\n%s", source)
	}
}

func TestGenerateExecutableCardSourceFixedCounterSpell(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Counter",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Put a -1/-1 counter on target creature.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		`"github.com/natefinch/council4/mtg/game/counter"`,
		"Primitive: game.AddCounter",
		"Amount:      game.Fixed(1)",
		"Object:      game.TargetPermanentReference(0)",
		"CounterKind: counter.MinusOneMinusOne",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("source missing %q:\n%s", want, source)
		}
	}
}

func TestGenerateExecutableCardSourceNamedPlayerCounterSpell(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Counter",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Put X poison counters on target player, where X is the number of lands you control.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"Primitive: game.AddPlayerCounter",
		"game.DynamicAmountCountSelector",
		"game.TargetPlayerReference(0)",
		"CounterKind: counter.Poison",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("source missing %q:\n%s", want, source)
		}
	}
}

func TestGenerateExecutableCardSourceRejectsCountersWithoutRuntimeMechanics(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"stun", "finality"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
				Name:       "Test Counter",
				Layout:     "normal",
				TypeLine:   "Sorcery",
				OracleText: "Put a " + name + " counter on target creature.",
			}, "t")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) == 0 {
				t.Fatalf("%s placement generated without diagnostics:\n%s", name, source)
			}
			if source != "" {
				t.Fatalf("%s placement generated source:\n%s", name, source)
			}
		})
	}
}

func TestGenerateExecutableCardSourceRegenerate(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Regenerate",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Regenerate target creature.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if !strings.Contains(source, "Primitive: game.Regenerate") {
		t.Fatalf("source missing Regenerate primitive:\n%s", source)
	}
}

func TestGenerateExecutableCardSourceFight(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Fight",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Target creature you control fights target creature you don't control. (Each deals damage equal to its power to the other.)",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if !strings.Contains(source, "Primitive: game.Fight") ||
		!strings.Contains(source, "RelatedObject: game.TargetPermanentReference(1)") {
		t.Fatalf("source missing Fight primitive:\n%s", source)
	}
}

func TestGenerateExecutableCardSourceFixedDamageTargets(t *testing.T) {
	t.Parallel()
	tests := []struct {
		target string
		wanted string
	}{
		{target: "creature", wanted: "PermanentTypes: []types.Card{types.Creature}"},
		{target: "planeswalker", wanted: "PermanentTypes: []types.Card{types.Planeswalker}"},
		{target: "player", wanted: "game.TargetAllowPlayer"},
		{target: "opponent", wanted: "Player: game.PlayerOpponent"},
	}
	for _, test := range tests {
		t.Run(test.target, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{
				Name:       "Test Bolt",
				Layout:     "normal",
				ManaCost:   "{R}",
				TypeLine:   "Instant",
				OracleText: "Test Bolt deals 2 damage to target " + test.target + ".",
				Colors:     []string{"R"},
			}
			source, diagnostics, err := GenerateExecutableCardSource(card, "t")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			if !strings.Contains(source, test.wanted) {
				t.Fatalf("source missing %q:\n%s", test.wanted, source)
			}
		})
	}
}

func TestGenerateExecutableCardSourceGroupDamage(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		oracleText string
		wantedSnip string
	}{
		{
			name:       "each opponent",
			oracleText: "Test Bolt deals 3 damage to each opponent.",
			wantedSnip: "game.PlayerGroupDamageRecipient(game.OpponentsReference())",
		},
		{
			name:       "each player",
			oracleText: "Test Bolt deals 3 damage to each player.",
			wantedSnip: "game.PlayerGroupDamageRecipient(game.AllPlayersReference())",
		},
		{
			name:       "each creature",
			oracleText: "Test Bolt deals 3 damage to each creature.",
			wantedSnip: "game.GroupDamageRecipient(game.BattlefieldGroup(",
		},
		{
			name:       "each other creature",
			oracleText: "Test Bolt deals 3 damage to each other creature.",
			wantedSnip: "game.GroupDamageRecipient(game.BattlefieldGroupExcluding(",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{
				Name:       "Test Bolt",
				Layout:     "normal",
				ManaCost:   "{R}",
				TypeLine:   "Instant",
				OracleText: test.oracleText,
				Colors:     []string{"R"},
			}
			source, diagnostics, err := GenerateExecutableCardSource(card, "t")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			if source == "" {
				t.Fatal("expected non-empty source")
			}
			for _, wanted := range []string{
				"Primitive: game.Damage",
				"game.Fixed(3)",
				test.wantedSnip,
			} {
				if !strings.Contains(source, wanted) {
					t.Fatalf("source missing %q:\n%s", wanted, source)
				}
			}
		})
	}
}

func TestGenerateExecutableCardSourceFixedDraw(t *testing.T) {
	t.Parallel()
	for _, oracleText := range []string{"Draw a card.", "Draw two cards."} {
		t.Run(oracleText, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{
				Name:       "Test Draw",
				Layout:     "normal",
				ManaCost:   "{2}{U}",
				TypeLine:   "Sorcery",
				OracleText: oracleText,
				Colors:     []string{"U"},
			}
			source, diagnostics, err := GenerateExecutableCardSource(card, "t")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			amount := 1
			if oracleText == "Draw two cards." {
				amount = 2
			}
			for _, wanted := range []string{
				"Primitive: game.Draw",
				fmt.Sprintf("game.Fixed(%d)", amount),
				"Player: game.ControllerReference()",
			} {
				if !strings.Contains(source, wanted) {
					t.Fatalf("source missing %q:\n%s", wanted, source)
				}
			}
		})
	}
}

func TestGenerateExecutableCardSourceTargetPlayerDraw(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Draw",
		Layout:     "normal",
		ManaCost:   "{2}{U}",
		TypeLine:   "Sorcery",
		OracleText: "Target player draws two cards.",
		Colors:     []string{"U"},
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		`Constraint: "target player"`,
		"game.TargetAllowPlayer",
		"Primitive: game.Draw",
		"game.Fixed(2)",
		"Player: game.TargetPlayerReference(0)",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceDestroyCreature(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Doom",
		Layout:     "normal",
		ManaCost:   "{1}{B}",
		TypeLine:   "Instant",
		OracleText: "Destroy target creature.",
		Colors:     []string{"B"},
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		`Constraint: "target creature"`,
		"PermanentTypes: []types.Card{types.Creature}",
		"Primitive: game.Destroy",
		"Object: game.TargetPermanentReference(0)",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceDestroyPermanentTypes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		target string
		wanted string
	}{
		{target: "artifact", wanted: "types.Artifact"},
		{target: "enchantment", wanted: "types.Enchantment"},
		{target: "land", wanted: "types.Land"},
		{target: "permanent", wanted: `Constraint: "target permanent"`},
	}
	for _, test := range tests {
		t.Run(test.target, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{
				Name:       "Test Doom",
				Layout:     "normal",
				TypeLine:   "Sorcery",
				OracleText: "Destroy target " + test.target + ".",
			}
			source, diagnostics, err := GenerateExecutableCardSource(card, "t")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			if !strings.Contains(source, test.wanted) {
				t.Fatalf("source missing %q:\n%s", test.wanted, source)
			}
		})
	}
}

func TestGenerateExecutableCardSourceDestroyAllCreatures(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Wrath",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Destroy all creatures.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Primitive: game.Destroy",
		"Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}})",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceDestroyAllOtherCreatures(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test One-Sided Wrath",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Destroy all other creatures.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Primitive: game.Destroy",
		"Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, ExcludeSource: true})",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceModifyTargetCreature(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Weakness",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Target creature gets -4/-4 until end of turn.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		`Constraint: "target creature"`,
		"Primitive: game.ModifyPT",
		"PowerDelta:",
		"game.Fixed(-4)",
		"ToughnessDelta:",
		"game.DurationUntilEndOfTurn",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourcePumpTargetCreature(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Growth",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Target creature gets +2/+0 until end of turn.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{"game.Fixed(2)", "game.Fixed(0)"} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceTemporaryContinuousEffects(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		oracleText string
		wants      []string
	}{
		{
			name:       "group power toughness",
			oracleText: "Creatures you control get +1/+1 until end of turn.",
			wants: []string{
				"game.ApplyContinuous",
				"game.BattlefieldGroup",
				"Controller: game.ControllerYou",
				"game.LayerPowerToughnessModify",
				"PowerDelta:",
				"ToughnessDelta:",
			},
		},
		{
			name:       "target keyword",
			oracleText: "Target creature gains flying until end of turn.",
			wants: []string{
				"game.ApplyContinuous",
				"Object: opt.Val(game.TargetPermanentReference(0))",
				"game.LayerAbility",
				"game.Flying",
			},
		},
		{
			name:       "target power toughness and keyword",
			oracleText: "Target creature gets +2/+2 and gains trample until end of turn.",
			wants: []string{
				"game.ApplyContinuous",
				"game.LayerPowerToughnessModify",
				"PowerDelta:",
				"ToughnessDelta:",
				"game.LayerAbility",
				"game.Trample",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
				Name:       "Test Effect",
				Layout:     "normal",
				TypeLine:   "Instant",
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

func TestGenerateExecutableCardSourceNegativeZeroToughness(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Shrink",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Target creature gets -5/-0 until end of turn.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{"game.Fixed(-5)", "game.Fixed(0)"} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceExileCreature(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Exile",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Exile target creature.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		`Constraint: "target creature"`,
		"PermanentTypes: []types.Card{types.Creature}",
		"Primitive: game.Exile",
		"Object: game.TargetPermanentReference(0)",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceBounceCreature(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Bounce",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Return target creature to its owner's hand.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		`Constraint: "target creature"`,
		"PermanentTypes: []types.Card{types.Creature}",
		"Primitive: game.Bounce",
		"Object: game.TargetPermanentReference(0)",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceGainLife(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Salve",
		Layout:     "normal",
		ManaCost:   "{W}",
		TypeLine:   "Instant",
		OracleText: "You gain 3 life.",
		Colors:     []string{"W"},
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Primitive: game.GainLife",
		"game.Fixed(3)",
		"Player: game.ControllerReference()",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceLifeRecipients(t *testing.T) {
	t.Parallel()
	tests := []struct {
		text      string
		primitive string
		recipient string
	}{
		{text: "You lose 2 life.", primitive: "game.LoseLife", recipient: "game.ControllerReference()"},
		{text: "Target player gains 4 life.", primitive: "game.GainLife", recipient: "Player: game.TargetPlayerReference(0)"},
		{text: "Target opponent loses 3 life.", primitive: "game.LoseLife", recipient: "Player: game.PlayerOpponent"},
	}
	for _, test := range tests {
		t.Run(test.text, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{
				Name:       "Test Life",
				Layout:     "normal",
				TypeLine:   "Sorcery",
				OracleText: test.text,
			}
			source, diagnostics, err := GenerateExecutableCardSource(card, "t")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			for _, wanted := range []string{test.primitive, test.recipient} {
				if !strings.Contains(source, wanted) {
					t.Fatalf("source missing %q:\n%s", wanted, source)
				}
			}
		})
	}
}

func TestGenerateExecutableCardSourceLifeGroupRecipients(t *testing.T) {
	t.Parallel()
	tests := []struct {
		text    string
		wanteds []string
	}{
		{
			text: "Each opponent loses 3 life.",
			wanteds: []string{
				"game.LoseLife",
				"game.Fixed(3)",
				"PlayerGroup: game.OpponentsReference()",
			},
		},
		{
			text: "Each player gains 2 life.",
			wanteds: []string{
				"game.GainLife",
				"game.Fixed(2)",
				"PlayerGroup: game.AllPlayersReference()",
			},
		},
		{
			text: "Each opponent loses 1 life and you gain 1 life.",
			wanteds: []string{
				"game.LoseLife",
				"PlayerGroup: game.OpponentsReference()",
				`PublishResult: game.ResultKey("life-change")`,
				"game.GainLife",
				"Player: game.ControllerReference()",
			},
		},
		{
			text: "Each opponent loses 1 life and you gain that much life.",
			wanteds: []string{
				"game.LoseLife",
				"PlayerGroup: game.OpponentsReference()",
				`PublishResult: game.ResultKey("life-change")`,
				"game.GainLife",
				"game.DynamicAmountPreviousEffectResult",
				`ResultKey: game.ResultKey("life-change")`,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.text, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{
				Name:       "Test Life",
				Layout:     "normal",
				TypeLine:   "Sorcery",
				OracleText: test.text,
			}
			source, diagnostics, err := GenerateExecutableCardSource(card, "t")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			for _, wanted := range test.wanteds {
				if !strings.Contains(source, wanted) {
					t.Fatalf("source missing %q:\n%s", wanted, source)
				}
			}
		})
	}
}

func TestGenerateExecutableCardSourceScry(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		text   string
		amount int
	}{
		{text: "Scry 2.", amount: 2},
		{text: "Scry three.", amount: 3},
	} {
		t.Run(test.text, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{
				Name:       "Test Vision",
				Layout:     "normal",
				ManaCost:   "{U}",
				TypeLine:   "Sorcery",
				OracleText: test.text,
				Colors:     []string{"U"},
			}
			source, diagnostics, err := GenerateExecutableCardSource(card, "t")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			for _, wanted := range []string{
				"Primitive: game.Scry",
				fmt.Sprintf("game.Fixed(%d)", test.amount),
				"Player: game.ControllerReference()",
			} {
				if !strings.Contains(source, wanted) {
					t.Fatalf("source missing %q:\n%s", wanted, source)
				}
			}
		})
	}
}

func TestGenerateExecutableCardSourceDiscard(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Mind",
		Layout:     "normal",
		ManaCost:   "{2}{B}",
		TypeLine:   "Sorcery",
		OracleText: "Target player discards two cards.",
		Colors:     []string{"B"},
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		`Constraint: "Target player"`,
		"game.TargetAllowPlayer",
		"Primitive: game.Discard",
		"game.Fixed(2)",
		"Player: game.TargetPlayerReference(0)",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceControllerDiscard(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Mind",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Discard a card.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Primitive: game.Discard",
		"game.Fixed(1)",
		"Player: game.ControllerReference()",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceTapTarget(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Sleep",
		Layout:     "normal",
		ManaCost:   "{U}",
		TypeLine:   "Instant",
		OracleText: "Tap target creature.",
		Colors:     []string{"U"},
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		`Constraint: "target creature"`,
		"PermanentTypes: []types.Card{types.Creature}",
		"Primitive: game.Tap",
		"Object: game.TargetPermanentReference(0)",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceTapUntapTargets(t *testing.T) {
	t.Parallel()
	tests := []struct {
		text      string
		cardType  string
		primitive string
	}{
		{text: "Tap target artifact.", cardType: "types.Artifact", primitive: "game.Tap"},
		{text: "Untap target land.", cardType: "types.Land", primitive: "game.Untap"},
		{text: "Untap target permanent.", cardType: `Constraint: "target permanent"`, primitive: "game.Untap"},
	}
	for _, test := range tests {
		t.Run(test.text, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{
				Name:       "Test Twiddle",
				Layout:     "normal",
				TypeLine:   "Instant",
				OracleText: test.text,
			}
			source, diagnostics, err := GenerateExecutableCardSource(card, "t")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			for _, wanted := range []string{test.cardType, test.primitive} {
				if !strings.Contains(source, wanted) {
					t.Fatalf("source missing %q:\n%s", wanted, source)
				}
			}
		})
	}
}

func TestGenerateExecutableCardSourceMill(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Mill",
		Layout:     "normal",
		ManaCost:   "{1}{U}",
		TypeLine:   "Sorcery",
		OracleText: "Target player mills three cards.",
		Colors:     []string{"U"},
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"game.TargetAllowPlayer",
		"Primitive: game.Mill",
		"game.Fixed(3)",
		"Player: game.TargetPlayerReference(0)",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceControllerMill(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Mill",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Mill four cards.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Primitive: game.Mill",
		"game.Fixed(4)",
		"Player: game.ControllerReference()",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceEnterDrawTrigger(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Drawing Bear",
		Layout:     "normal",
		ManaCost:   "{2}{U}",
		TypeLine:   "Creature — Bear",
		OracleText: "When this creature enters, draw a card.",
		Colors:     []string{"U"},
		Power:      new("2"),
		Toughness:  new("2"),
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "d")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"TriggeredAbilities: []game.TriggeredAbility",
		"Type: game.TriggerWhen",
		"game.EventPermanentEnteredBattlefield",
		"game.TriggerSourceSelf",
		"Primitive: game.Draw",
		"game.Fixed(1)",
		"Player: game.ControllerReference()",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceKickedEnterTrigger(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Kicked Bear",
		Layout:     "normal",
		ManaCost:   "{2}{G}",
		TypeLine:   "Creature — Bear",
		OracleText: "Kicker {1}{G}\nWhen this creature enters, if it was kicked, you gain 4 life.",
		Colors:     []string{"G"},
		Power:      new("2"),
		Toughness:  new("2"),
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "k")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"InterveningIf:",
		`"if it was kicked"`,
		"InterveningIfEventPermanentWasKicked: true",
		"Primitive: game.GainLife",
		"game.Fixed(4)",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceAdditionalEnterConditions(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		condition string
		wants     []string
	}{
		{
			name:      "was cast",
			condition: "if it was cast",
			wants:     []string{"InterveningIfEventPermanentWasCast: true"},
		},
		{
			name:      "controls artifact",
			condition: "if you control an artifact",
			wants:     []string{"InterveningCondition: opt.Val", "RequiredTypes: []types.Card{types.Artifact}"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
				Name:       "Conditional Bear",
				Layout:     "normal",
				TypeLine:   "Creature — Bear",
				OracleText: "When this creature enters, " + test.condition + ", draw a card.",
				Power:      new("2"),
				Toughness:  new("2"),
			}, "c")
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

func TestGenerateExecutableCardSourceEnterMultipleEffectTrigger(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Resourceful Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		OracleText: "When this creature enters, draw a card. You gain 2 life.",
		Power:      new("2"),
		Toughness:  new("2"),
	}

	source, diagnostics, err := GenerateExecutableCardSource(card, "r")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	draw := strings.Index(source, "Primitive: game.Draw")
	gain := strings.Index(source, "Primitive: game.GainLife")
	if draw < 0 || gain < 0 || draw >= gain {
		t.Fatalf("trigger sequence is not draw then gain life:\n%s", source)
	}
}

func TestGenerateExecutableCardSourceThenJoinedSpellEffects(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		oracleText string
		first      string
		second     string
	}{
		{
			name:       "draw then discard",
			oracleText: "Draw two cards, then discard a card.",
			first:      "game.Draw",
			second:     "game.Discard",
		},
		{
			name:       "scry then draw",
			oracleText: "Scry 2, then draw a card.",
			first:      "game.Scry",
			second:     "game.Draw",
		},
		{
			name:       "discard then draw",
			oracleText: "Discard a card, then draw a card.",
			first:      "game.Discard",
			second:     "game.Draw",
		},
		{
			name:       "targeted mill then draw",
			oracleText: "Target player mills three cards, then draws a card.",
			first:      "game.Mill",
			second:     "game.Draw",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{
				Name:       "Test Spell",
				Layout:     "normal",
				TypeLine:   "Sorcery",
				OracleText: test.oracleText,
			}
			source, diagnostics, err := GenerateExecutableCardSource(card, "t")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			first := strings.Index(source, "Primitive: "+test.first)
			second := strings.Index(source, "Primitive: "+test.second)
			if first < 0 || second < 0 || first >= second {
				t.Fatalf("sequence is not %s then %s:\n%s", test.first, test.second, source)
			}
		})
	}
}

func TestGenerateExecutableCardSourceLibrarySearches(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		oracleText string
		wants      []string
	}{
		{
			name:       "Diabolic Tutor",
			oracleText: "Search your library for a card, put that card into your hand, then shuffle.",
			wants:      []string{"zone.Hand", "Amount: game.Fixed(1)"},
		},
		{
			name:       "Rampant Growth",
			oracleText: "Search your library for a basic land card, put it onto the battlefield tapped, then shuffle.",
			wants: []string{
				"zone.Battlefield",
				"opt.Val(types.Land)",
				"opt.Val(types.Basic)",
				"EntersTapped",
			},
		},
		{
			name:       "Three Visits",
			oracleText: "Search your library for a Forest card, put it onto the battlefield, then shuffle.",
			wants:      []string{"zone.Battlefield", "SubtypesAny: []types.Sub{types.Forest}"},
		},
		{
			name:       "Farseek",
			oracleText: "Search your library for a Plains, Island, Swamp, or Mountain card, put it onto the battlefield tapped, then shuffle.",
			wants: []string{
				"[]types.Sub{types.Plains, types.Island, types.Swamp, types.Mountain}",
				"zone.Battlefield",
				"EntersTapped",
			},
		},
		{
			name:       "Safewright Quest",
			oracleText: "Search your library for a Forest or Plains card, reveal it, put it into your hand, then shuffle.",
			wants: []string{
				"SubtypesAny: []types.Sub{types.Forest, types.Plains}",
				"zone.Hand",
				"Reveal",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
				Name:       test.name,
				Layout:     "normal",
				TypeLine:   "Sorcery",
				OracleText: test.oracleText,
			}, "s")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			for _, want := range append([]string{
				"Primitive: game.Search",
				"zone.Library",
				"Player: game.ControllerReference()",
			}, test.wants...) {
				if !strings.Contains(source, want) {
					t.Fatalf("source missing %q:\n%s", want, source)
				}
			}
		})
	}
}

func TestGenerateExecutableCardSourceRejectsUnsupportedLibrarySearches(t *testing.T) {
	t.Parallel()
	tests := []string{
		"Search your library for a card, put that card into your graveyard, then shuffle.",
		"Search your library for a green creature card, put it into your hand, then shuffle.",
		"Search your library for up to two basic land cards, put them onto the battlefield tapped, then shuffle.",
		"Search target opponent's library for a card, put that card into their hand, then shuffle.",
		"Search your library for that card, reveal it, put it into your hand, then shuffle.",
		"Search your library for a creature card, reveal it, put it into your hand or graveyard, then shuffle.",
	}
	for _, oracleText := range tests {
		t.Run(oracleText, func(t *testing.T) {
			t.Parallel()
			source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
				Name:       "Unsupported Search",
				Layout:     "normal",
				TypeLine:   "Sorcery",
				OracleText: oracleText,
			}, "u")
			if err != nil {
				t.Fatal(err)
			}
			if source != "" {
				t.Fatalf("source = %q, want no partial card", source)
			}
			if len(diagnostics) == 0 {
				t.Fatal("expected unsupported search diagnostic")
			}
			if diagnostics[0].Summary != "unsupported search effect" {
				t.Fatalf("diagnostics = %#v, want unsupported search effect", diagnostics)
			}
		})
	}
}

func TestGenerateExecutableCardSourceOptionalEnterTrigger(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Thoughtful Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		OracleText: "When this creature enters, you may draw a card.",
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
		"Optional: true",
		"Primitive: game.Draw",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceHostCreature(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Eager Beaver",
		Layout:     "host",
		TypeLine:   "Host Creature — Beaver",
		OracleText: "When this creature enters, you may untap target permanent.",
		Power:      new("3"),
		Toughness:  new("2"),
	}

	source, diagnostics, err := GenerateExecutableCardSource(card, "e")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if !strings.Contains(source, "Supertypes: []types.Super{types.Host}") {
		t.Fatalf("source does not preserve Host supertype:\n%s", source)
	}
}

func TestGenerateExecutableCardSourceEnterLifeTrigger(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Healing Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		OracleText: "When this permanent enters, you gain 3 life.",
		Power:      new("2"),
		Toughness:  new("2"),
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "h")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"TriggeredAbilities: []game.TriggeredAbility",
		"Primitive: game.GainLife",
		"game.Fixed(3)",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceEnterTargetTrigger(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Breaking Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		OracleText: "When this creature enters, destroy target artifact.",
		Power:      new("2"),
		Toughness:  new("2"),
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "b")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"TriggeredAbilities: []game.TriggeredAbility",
		"PermanentTypes: []types.Card{types.Artifact}",
		"Primitive: game.Destroy",
		"Object: game.TargetPermanentReference(0)",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceLandEnterTrigger(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Refuge",
		Layout:     "normal",
		TypeLine:   "Land",
		OracleText: "When this land enters, you gain 1 life.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"game.EventPermanentEnteredBattlefield",
		"game.TriggerSourceSelf",
		"Primitive: game.GainLife",
		"game.Fixed(1)",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceNonSelfEnterTrigger(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		cardName   string
		typeLine   string
		oracleText string
		wants      []string
	}{
		{
			name:       "another creature any controller",
			cardName:   "Test Bear",
			typeLine:   "Creature — Bear",
			oracleText: "Whenever another creature enters, draw a card.",
			wants: []string{
				"game.TriggerWhenever",
				"game.EventPermanentEnteredBattlefield",
				"ExcludeSelf:",
				"RequirePermanentTypes: []types.Card{types.Creature}",
				"Primitive: game.Draw",
			},
		},
		{
			name:       "a creature any controller",
			cardName:   "Test Bear",
			typeLine:   "Creature — Bear",
			oracleText: "Whenever a creature enters, draw a card.",
			wants: []string{
				"game.TriggerWhenever",
				"game.EventPermanentEnteredBattlefield",
				"RequirePermanentTypes: []types.Card{types.Creature}",
			},
		},
		{
			name:       "another creature you control",
			cardName:   "Test Bear",
			typeLine:   "Creature — Bear",
			oracleText: "Whenever another creature you control enters, draw a card.",
			wants: []string{
				"ExcludeSelf:",
				"game.TriggerControllerYou",
				"RequirePermanentTypes: []types.Card{types.Creature}",
			},
		},
		{
			name:       "another creature opponent controls",
			cardName:   "Test Bear",
			typeLine:   "Creature — Bear",
			oracleText: "Whenever another creature an opponent controls enters, draw a card.",
			wants: []string{
				"ExcludeSelf:",
				"game.TriggerControllerOpponent",
				"RequirePermanentTypes: []types.Card{types.Creature}",
			},
		},
		{
			name:       "another land enters",
			cardName:   "Test Shaman",
			typeLine:   "Creature — Human Shaman",
			oracleText: "Whenever another land enters, you gain 1 life.",
			wants: []string{
				"ExcludeSelf:",
				"RequirePermanentTypes: []types.Card{types.Land}",
				"Primitive: game.GainLife",
			},
		},
		{
			name:       "a land you control enters",
			cardName:   "Test Shaman",
			typeLine:   "Creature — Human Shaman",
			oracleText: "Whenever a land you control enters, draw a card.",
			wants: []string{
				"game.TriggerControllerYou",
				"RequirePermanentTypes: []types.Card{types.Land}",
			},
		},
		{
			name:       "another artifact enters",
			cardName:   "Test Bear",
			typeLine:   "Creature — Bear",
			oracleText: "Whenever another artifact enters, draw a card.",
			wants:      []string{"RequirePermanentTypes: []types.Card{types.Artifact}"},
		},
		{
			name:       "another enchantment enters",
			cardName:   "Test Bear",
			typeLine:   "Creature — Bear",
			oracleText: "Whenever another enchantment enters, draw a card.",
			wants:      []string{"RequirePermanentTypes: []types.Card{types.Enchantment}"},
		},
		{
			name:       "another planeswalker enters",
			cardName:   "Test Bear",
			typeLine:   "Creature — Bear",
			oracleText: "Whenever another planeswalker enters, draw a card.",
			wants:      []string{"RequirePermanentTypes: []types.Card{types.Planeswalker}"},
		},
		{
			name:       "another permanent enters any type",
			cardName:   "Test Bear",
			typeLine:   "Creature — Bear",
			oracleText: "Whenever another permanent enters, draw a card.",
			wants: []string{
				"ExcludeSelf:",
				"game.EventPermanentEnteredBattlefield",
			},
		},
		{
			name:       "a nontoken creature enters",
			cardName:   "Test Bear",
			typeLine:   "Creature — Bear",
			oracleText: "Whenever a nontoken creature enters, draw a card.",
			wants: []string{
				"RequireNonToken:",
				"RequirePermanentTypes: []types.Card{types.Creature}",
			},
		},
		{
			name:       "another nontoken creature you control enters",
			cardName:   "Test Bear",
			typeLine:   "Creature — Bear",
			oracleText: "Whenever another nontoken creature you control enters, draw a card.",
			wants: []string{
				"ExcludeSelf:",
				"game.TriggerControllerYou",
				"RequireNonToken:",
				"RequirePermanentTypes: []types.Card{types.Creature}",
			},
		},
		{
			name:       "one or more artifacts you control enter",
			cardName:   "Test Bear",
			typeLine:   "Creature — Bear",
			oracleText: "Whenever one or more artifacts you control enter, draw a card.",
			wants: []string{
				"game.TriggerWhenever",
				"OneOrMore:",
				"RequirePermanentTypes: []types.Card{types.Artifact}",
				"game.TriggerControllerYou",
			},
		},
		{
			name:       "ability word is ignored",
			cardName:   "Test Shaman",
			typeLine:   "Creature — Human Shaman",
			oracleText: "Landfall — Whenever a land you control enters, draw a card.",
			wants: []string{
				"RequirePermanentTypes: []types.Card{types.Land}",
				"game.TriggerControllerYou",
			},
		},
		{
			name:       "optional trigger",
			cardName:   "Test Shaman",
			typeLine:   "Creature — Human Shaman",
			oracleText: "Whenever a creature you control enters, you may draw a card.",
			wants: []string{
				"Optional: true",
				"game.TriggerControllerYou",
				"RequirePermanentTypes: []types.Card{types.Creature}",
			},
		},
		{
			name:       "event permanent pronoun gets buff",
			cardName:   "Test Druid",
			typeLine:   "Creature — Elf Druid",
			oracleText: "Whenever another creature you control enters, it gets +2/+0 until end of turn.",
			wants: []string{
				"game.EventPermanentReference()",
				"game.ModifyPT{",
				"ExcludeSelf:",
				"game.TriggerControllerYou",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{
				Name:       test.cardName,
				Layout:     "normal",
				TypeLine:   test.typeLine,
				OracleText: test.oracleText,
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
			for _, want := range test.wants {
				if !strings.Contains(source, want) {
					t.Fatalf("source missing %q:\n%s", want, source)
				}
			}
		})
	}
}

func TestGenerateExecutableCardSourceRejectsUnsupportedNonSelfEnterTriggers(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		cardName   string
		typeLine   string
		oracleText string
	}{
		{name: "subtype filter", cardName: "Test Bear", typeLine: "Creature — Bear", oracleText: "Whenever another zombie enters, draw a card."},
		{name: "power filter", cardName: "Test Bear", typeLine: "Creature — Bear", oracleText: "Whenever a creature with power 2 or less enters, draw a card."},
		{name: "compound type", cardName: "Test Bear", typeLine: "Creature — Bear", oracleText: "Whenever another creature or artifact enters, draw a card."},
		{name: "missing article", cardName: "Test Bear", typeLine: "Creature — Bear", oracleText: "Whenever creature enters, draw a card."},
		{name: "unknown suffix", cardName: "Test Bear", typeLine: "Creature — Bear", oracleText: "Whenever another creature that you control enters, draw a card."},
		{name: "caster-relative condition", cardName: "Test Bear", typeLine: "Creature — Bear", oracleText: "Whenever a creature enters, if you cast it, draw a card."},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{
				Name:       test.cardName,
				Layout:     "normal",
				TypeLine:   test.typeLine,
				OracleText: test.oracleText,
				Power:      new("2"),
				Toughness:  new("2"),
			}
			source, diagnostics, err := GenerateExecutableCardSource(card, "t")
			if err != nil {
				t.Fatal(err)
			}
			if source != "" {
				t.Fatalf("unexpected source:\n%s", source)
			}
			if len(diagnostics) == 0 {
				t.Fatal("expected unsupported diagnostic")
			}
		})
	}
}

func TestGenerateExecutableCardSourceDiesTrigger(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		OracleText: "When this creature dies, draw two cards.",
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
		"game.EventPermanentDied",
		"game.TriggerSourceSelf",
		"Primitive: game.Draw",
		"game.Fixed(2)",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceSelfDiesDamageTrigger(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Devil",
		Layout:     "normal",
		TypeLine:   "Creature — Devil",
		OracleText: "When this creature dies, it deals 3 damage to any target.",
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
		"game.EventPermanentDied",
		"Primitive: game.Damage",
		"DamageSource: opt.Val(game.EventPermanentReference())",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceSelfDiesCounterAbsence(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Undying Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		OracleText: "When this creature dies, if it had no +1/+1 counters on it, draw a card.",
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
		`InterveningIf: "if it had no +1/+1 counters on it"`,
		"InterveningIfEventPermanentHadNoCounterKind: opt.Val(counter.PlusOnePlusOne)",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceSelfDiesEventCardReturn(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Phoenix",
		Layout:     "normal",
		TypeLine:   "Creature — Phoenix",
		OracleText: "When this creature dies, return it to its owner's hand.",
		Power:      new("2"),
		Toughness:  new("2"),
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Primitive: game.MoveCard",
		"game.CardReference{Kind: game.CardReferenceEvent}",
		"zone.Graveyard",
		"zone.Hand",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceSelfDiesAdventurePermission(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:   "Test Dreadknight // Test Whispers",
		Layout: "adventure",
		CardFaces: []ScryfallCardFace{
			{
				Name:       "Test Dreadknight",
				ManaCost:   "{1}{G}",
				TypeLine:   "Creature — Human Knight",
				OracleText: "When Test Dreadknight dies, you may cast it from your graveyard as an Adventure until the end of your next turn.",
				Power:      new("2"),
				Toughness:  new("1"),
			},
			{
				Name:       "Test Whispers",
				ManaCost:   "{1}{B}",
				TypeLine:   "Sorcery — Adventure",
				OracleText: "Draw a card.",
			},
		},
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Optional: true",
		"Primitive: game.GrantCastPermission",
		"game.CardReference{Kind: game.CardReferenceEvent}",
		"zone.Graveyard",
		"Face:     game.FaceAlternate",
		"Duration: game.DurationUntilEndOfYourNextTurn",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceDynamicSelfDiesDamage(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Devil",
		Layout:     "normal",
		TypeLine:   "Creature — Devil",
		OracleText: "When this creature dies, it deals damage equal to its power to any target.",
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
		"game.DynamicAmountObjectPower",
		"Object:     game.EventPermanentReference()",
		"DamageSource: opt.Val(game.EventPermanentReference())",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceDynamicCountDamage(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Swarm",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Test Swarm deals damage equal to twice the number of creatures you control to any target.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"game.DynamicAmountCountSelector",
		"Multiplier: 2",
		"game.BattlefieldGroup",
		"types.Creature",
		"Controller: game.ControllerYou",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
	if strings.Contains(source, "/counter\"") {
		t.Fatalf("source has unused counter import:\n%s", source)
	}
}

func TestGenerateExecutableCardSourceDynamicSourcePowerCounters(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Druid",
		Layout:     "normal",
		TypeLine:   "Creature — Druid",
		Power:      new("2"),
		Toughness:  new("2"),
		OracleText: "{T}: Put X +1/+1 counters on target creature, where X is Test Druid's power.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"game.DynamicAmountObjectPower",
		"Object:     game.SourcePermanentReference()",
		"CounterKind: counter.PlusOnePlusOne",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceDynamicSourcePowerDamage(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Devil",
		Layout:     "normal",
		TypeLine:   "Creature — Devil",
		Power:      new("2"),
		Toughness:  new("2"),
		OracleText: "{T}: Test Devil deals damage equal to Test Devil's power to any target.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"game.DynamicAmountObjectPower",
		"Object:     game.SourcePermanentReference()",
		"DamageSource: opt.Val(game.SourcePermanentReference())",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceRejectsAmbiguousDynamicAmount(t *testing.T) {
	t.Parallel()
	for _, oracleText := range []string{
		"Test Swarm deals damage equal to the number of creatures you control plus one to any target.",
		"Test Swarm deals damage equal to creatures you control to any target.",
		"You gain X life, where X is opponent.",
	} {
		t.Run(oracleText, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{
				Name:       "Test Swarm",
				Layout:     "normal",
				TypeLine:   "Sorcery",
				OracleText: oracleText,
			}

			source, diagnostics, err := GenerateExecutableCardSource(card, "t")
			if err != nil {
				t.Fatal(err)
			}
			if source != "" || len(diagnostics) == 0 {
				t.Fatalf("source = %q, diagnostics = %#v, want rejection", source, diagnostics)
			}
		})
	}
}

func TestGenerateExecutableCardSourceRejectsDynamicAmountNumberDisagreement(t *testing.T) {
	t.Parallel()
	for _, oracleText := range []string{
		"Draw a card for each creatures you control.",
		"Test Swarm deals damage equal to the number of creature you control to any target.",
	} {
		t.Run(oracleText, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{
				Name:       "Test Swarm",
				Layout:     "normal",
				TypeLine:   "Sorcery",
				OracleText: oracleText,
			}

			source, diagnostics, err := GenerateExecutableCardSource(card, "t")
			if err != nil {
				t.Fatal(err)
			}
			if source != "" || len(diagnostics) == 0 {
				t.Fatalf("source = %q, diagnostics = %#v, want rejection", source, diagnostics)
			}
		})
	}
}

func TestGenerateExecutableCardSourceRejectsAmbiguousDynamicPowerReference(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Druid",
		Layout:     "normal",
		TypeLine:   "Creature — Druid",
		Power:      new("2"),
		Toughness:  new("2"),
		OracleText: "{T}: Put X +1/+1 counters on target creature, where X is its power.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if source != "" || len(diagnostics) == 0 {
		t.Fatalf("source = %q, diagnostics = %#v, want rejection", source, diagnostics)
	}
}

func TestGenerateExecutableCardSourceDiesMultipleEffectTrigger(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Generous Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		OracleText: "When this creature dies, draw a card. You gain 2 life.",
		Power:      new("2"),
		Toughness:  new("2"),
	}

	source, diagnostics, err := GenerateExecutableCardSource(card, "g")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	draw := strings.Index(source, "Primitive: game.Draw")
	gain := strings.Index(source, "Primitive: game.GainLife")
	if draw < 0 || gain < 0 || draw >= gain {
		t.Fatalf("trigger sequence is not draw then gain life:\n%s", source)
	}
}

func TestGenerateExecutableCardSourceRejectsPartiallyOptionalTrigger(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Unclear Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		OracleText: "When this creature enters, you may draw a card. You gain 2 life.",
		Power:      new("2"),
		Toughness:  new("2"),
	}

	source, diagnostics, err := GenerateExecutableCardSource(card, "u")
	if err != nil {
		t.Fatal(err)
	}
	if source != "" {
		t.Fatalf("source = %q, want no partial card", source)
	}
	if len(diagnostics) != 1 || diagnostics[0].Summary != "unsupported enter trigger" {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
}

func TestGenerateExecutableCardSourceRejectsOptionalKickedEnterTrigger(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Unclear Kicker",
		Layout:     "normal",
		TypeLine:   "Creature — Wizard",
		OracleText: "Kicker {1}{U}\nWhen this creature enters, if it was kicked, you may draw a card.",
		Power:      new("2"),
		Toughness:  new("2"),
	}

	source, diagnostics, err := GenerateExecutableCardSource(card, "u")
	if err != nil {
		t.Fatal(err)
	}
	if source != "" {
		t.Fatalf("source = %q, want no partial card", source)
	}
	if len(diagnostics) != 1 || diagnostics[0].Summary != "unsupported enter trigger effect" {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
}

func TestGenerateExecutableCardSourceRejectsUnsupportedMechanicVariants(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		cardName   string
		typeLine   string
		oracleText string
	}{
		{name: "restricted mana choice", cardName: "Mana Bear", typeLine: "Creature — Bear", oracleText: "{T}: Add one mana of any color in your commander's color identity."},
		{name: "unsupported conditional tapped entry", cardName: "Test Land", typeLine: "Land", oracleText: "This land enters tapped unless you gained life this turn."},
		{name: "nonmana ward", cardName: "Test Bear", typeLine: "Creature — Bear", oracleText: "Ward—Pay 2 life."},
		{name: "typecycling", cardName: "Test Card", typeLine: "Sorcery", oracleText: "Plainscycling {2}"},
		{name: "nonmana equip", cardName: "Test Equipment", typeLine: "Artifact — Equipment", oracleText: "Equip—Pay {3} or discard a card."},
		{name: "qualified equip", cardName: "Test Equipment", typeLine: "Artifact — Equipment", oracleText: "Equip creature token {1}"},
		{name: "qualified enchant", cardName: "Test Aura", typeLine: "Enchantment — Aura", oracleText: "Enchant creature you control"},
		{name: "noncolor protection (replaced by support test)", cardName: "Test Bear", typeLine: "Creature — Bear", oracleText: "Protection from a chosen color"},
		{name: "divided damage", cardName: "Test Bolt", typeLine: "Instant", oracleText: "Test Bolt deals 3 damage divided as you choose among any number of targets."},
		{name: "variable surveil", cardName: "Test Surveil", typeLine: "Sorcery", oracleText: "Surveil X."},
		{name: "repeated proliferate", cardName: "Test Proliferate", typeLine: "Sorcery", oracleText: "Proliferate X times."},
		{name: "another fight target", cardName: "Test Fight", typeLine: "Sorcery", oracleText: "Target creature fights another target creature."},
		{name: "conditional draw", cardName: "Test Draw", typeLine: "Sorcery", oracleText: "If you control a creature, draw two cards."},
		{name: "conditional destroy", cardName: "Test Doom", typeLine: "Instant", oracleText: "If it is tapped, destroy target creature."},
		{name: "regeneration destroy", cardName: "Test Doom", typeLine: "Instant", oracleText: "Destroy target creature. It can't be regenerated."},
		{name: "restricted destroy", cardName: "Test Doom", typeLine: "Instant", oracleText: "Destroy target nonblack creature."},
		{name: "linked exile", cardName: "Test Exile", typeLine: "Instant", oracleText: "Exile target creature, then return it to the battlefield."},
		{name: "graveyard exile", cardName: "Test Exile", typeLine: "Instant", oracleText: "Exile target card from a graveyard."},
		{name: "bounce to your hand", cardName: "Test Bounce", typeLine: "Instant", oracleText: "Return target creature to your hand."},
		{name: "variable power toughness", cardName: "Test Growth", typeLine: "Instant", oracleText: "Target creature gets +X/+X until end of turn."},
		{name: "permanent power toughness", cardName: "Test Growth", typeLine: "Sorcery", oracleText: "Target creature gets +2/+2."},
		{name: "dynamic group power toughness", cardName: "Test Growth", typeLine: "Instant", oracleText: "Creatures you control get +X/+X until end of turn."},
		{name: "unsupported group power toughness duration", cardName: "Test Growth", typeLine: "Instant", oracleText: "Creatures you control get +2/+2 until your next turn."},
		{name: "group keyword grant", cardName: "Test Flight", typeLine: "Instant", oracleText: "Creatures you control gain flying until end of turn."},
		{name: "parameterized temporary keyword", cardName: "Test Ward", typeLine: "Instant", oracleText: "Target creature gains ward {2} until end of turn."},
		{name: "set life total", cardName: "Test Life", typeLine: "Sorcery", oracleText: "Your life total becomes 10."},
		{name: "compound life", cardName: "Test Life", typeLine: "Sorcery", oracleText: "You gain 3 life and draw a card."},
		{name: "variable scry", cardName: "Test Vision", typeLine: "Sorcery", oracleText: "Scry X."},
		{name: "conditional scry", cardName: "Test Vision", typeLine: "Sorcery", oracleText: "If you control a creature, scry 2."},
		{name: "targeted scry", cardName: "Test Vision", typeLine: "Sorcery", oracleText: "Target player scries 2."},
		{name: "random discard", cardName: "Test Mind", typeLine: "Sorcery", oracleText: "Target player discards a card at random."},
		{name: "named discard", cardName: "Test Mind", typeLine: "Sorcery", oracleText: "Target player discards a creature card."},
		{name: "hand discard", cardName: "Test Mind", typeLine: "Sorcery", oracleText: "Target player discards their hand."},
		{name: "optional discard", cardName: "Test Mind", typeLine: "Sorcery", oracleText: "You may discard a card."},
		{name: "mass tap", cardName: "Test Sleep", typeLine: "Sorcery", oracleText: "Tap all creatures."},
		{name: "optional tap", cardName: "Test Sleep", typeLine: "Instant", oracleText: "You may tap target creature."},
		{name: "unsupported tap qualifier", cardName: "Test Sleep", typeLine: "Instant", oracleText: "Tap target creature with flying."},
		{name: "freeze tap", cardName: "Test Sleep", typeLine: "Instant", oracleText: "Tap target creature. It doesn't untap during its controller's next untap step."},
		{name: "conditional untap", cardName: "Test Sleep", typeLine: "Instant", oracleText: "If it is tapped, untap target creature."},
		{name: "until mill", cardName: "Test Mill", typeLine: "Sorcery", oracleText: "Target player mills cards until they mill a land card."},
		{name: "reveal mill", cardName: "Test Mill", typeLine: "Sorcery", oracleText: "Target player reveals and mills three cards."},
		{name: "mass mill", cardName: "Test Mill", typeLine: "Sorcery", oracleText: "Each opponent mills three cards."},
		{name: "leave trigger", cardName: "Test Bear", typeLine: "Creature — Bear", oracleText: "When this creature leaves the battlefield, draw a card."},
		{name: "cast trigger", cardName: "Test Bear", typeLine: "Creature — Bear", oracleText: "When you cast this spell, draw a card."},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{
				Name:       test.cardName,
				Layout:     "normal",
				TypeLine:   test.typeLine,
				OracleText: test.oracleText,
			}
			source, diagnostics, err := GenerateExecutableCardSource(card, "t")
			if err != nil {
				t.Fatal(err)
			}
			if source != "" {
				t.Fatalf("unexpected source:\n%s", source)
			}
			if len(diagnostics) == 0 {
				t.Fatal("expected unsupported diagnostic")
			}
		})
	}
}

func TestGenerateExecutableCardSourceThenJoinedEnterTrigger(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Drawing Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		OracleText: "When this creature enters, draw a card, then discard a card.",
		Power:      new("2"),
		Toughness:  new("2"),
	}

	source, diagnostics, err := GenerateExecutableCardSource(card, "d")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	draw := strings.Index(source, "Primitive: game.Draw")
	discard := strings.Index(source, "Primitive: game.Discard")
	if draw < 0 || discard < 0 || draw >= discard {
		t.Fatalf("trigger sequence is not draw then discard:\n%s", source)
	}
}

func TestGenerateExecutableCardSourceRejectsPartiallyRecognizedKeywordLine(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Bayou Dragonfly",
		Layout:     "normal",
		TypeLine:   "Creature — Insect",
		OracleText: "Flying; swampwalk (This creature can't be blocked as long as defending player controls a Swamp.)",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "b")
	if err != nil {
		t.Fatal(err)
	}
	if source != "" || len(diagnostics) == 0 {
		t.Fatalf("source = %q, diagnostics = %#v", source, diagnostics)
	}
	if got := diagnostics[0].Summary; got != "unsupported mixed keyword ability" {
		t.Fatalf("summary = %q", got)
	}
}

func TestGenerateExecutableCardSourceRendersParameterizedKeywords(t *testing.T) {
	t.Parallel()
	for _, oracleText := range []string{
		"Kicker {1}{G}",
		"Madness {2}{B}",
		"Morph {3}{U}",
		"Disguise {4}{W}",
		"Toxic 2",
	} {
		t.Run(oracleText, func(t *testing.T) {
			t.Parallel()
			source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
				Name:       "Parameterized Creature",
				Layout:     "normal",
				TypeLine:   "Creature — Test",
				OracleText: oracleText,
				Power:      new("2"),
				Toughness:  new("2"),
			}, "p")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) != 0 || source == "" {
				t.Fatalf("source = %q, diagnostics = %#v", source, diagnostics)
			}
		})
	}
}

func TestGenerateExecutableCardSourceExplainsUnsupportedAbility(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		typeLine   string
		oracleText string
		summary    string
		detail     string
	}{
		"spell": {
			typeLine:   "Sorcery",
			oracleText: "Create a Treasure token.",
			summary:    "unsupported spell ability",
			detail:     "does not yet lower this spell ability",
		},
		"activated": {
			typeLine:   "Creature — Bear",
			oracleText: "Remove a +1/+1 counter from target creature: Draw a card.",
			summary:    "unsupported activated ability",
			detail:     "supports only exact typed costs",
		},
		"parameterized keyword": {
			typeLine:   "Creature — Snake",
			oracleText: "Annihilator 1",
			summary:    "unsupported parameterized keyword",
			detail:     `Annihilator with parameter "1"`,
		},
		"keyword without template": {
			typeLine:   "Creature — Dinosaur",
			oracleText: "Ward",
			summary:    "unsupported keyword ability",
			detail:     "no reusable game template for Ward",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{
				Name:       "Unsupported Example",
				Layout:     "normal",
				TypeLine:   test.typeLine,
				OracleText: test.oracleText,
			}
			source, diagnostics, err := GenerateExecutableCardSource(card, "u")
			if err != nil {
				t.Fatal(err)
			}
			if source != "" || len(diagnostics) == 0 {
				t.Fatalf("source = %q, diagnostics = %#v", source, diagnostics)
			}
			if got := diagnostics[0].Summary; got != test.summary {
				t.Fatalf("summary = %q, want %q", got, test.summary)
			}
			if got := diagnostics[0].Detail; !strings.Contains(got, test.detail) {
				t.Fatalf("detail = %q, want substring %q", got, test.detail)
			}
		})
	}
}

func TestGenerateExecutableCardSourceChooseTwo(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Command",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Choose two —\n• Draw a card.\n• Destroy target creature.\n• You gain 3 life.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{"MinModes: 2,", "MaxModes: 2,"} {
		if !strings.Contains(source, want) {
			t.Fatalf("source missing %q:\n%s", want, source)
		}
	}
}

func TestGenerateExecutableCardSourceVanilla(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:      "Vanilla Bear",
		Layout:    "normal",
		TypeLine:  "Creature — Bear",
		Power:     new("2"),
		Toughness: new("2"),
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "v")
	if err != nil {
		t.Fatal(err)
	}
	if source == "" || len(diagnostics) != 0 || strings.Contains(source, "TODO") {
		t.Fatalf("source = %q, diagnostics = %#v", source, diagnostics)
	}
}

func TestGenerateExecutableCardSourceRejectsUnknownTypeLine(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:     "Decklist",
		Layout:   "token",
		TypeLine: "Card",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "d")
	if err != nil {
		t.Fatal(err)
	}
	if source != "" || len(diagnostics) == 0 {
		t.Fatalf("source = %q, diagnostics = %#v", source, diagnostics)
	}
}

func TestGenerateExecutableCardSourceKeepsSameNamedFacesPositional(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:   "Insect // Insect",
		Layout: "reversible_card",
		CardFaces: []ScryfallCardFace{
			{Name: "Insect", TypeLine: "Creature — Insect", OracleText: "Flying"},
			{Name: "Insect", TypeLine: "Creature — Insect", OracleText: "Haste"},
		},
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "i")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if strings.Count(source, "game.FlyingStaticBody") != 1 ||
		strings.Count(source, "game.HasteStaticBody") != 1 {
		t.Fatalf("face abilities were not kept positional:\n%s", source)
	}
}

func TestGenerateExecutableCardSourceLoyaltyAbility(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Walker",
		Layout:     "normal",
		ManaCost:   "{2}{G}",
		TypeLine:   "Legendary Planeswalker — Test",
		OracleText: "+1: Draw a card.\n\u22122: You gain 3 life.",
		Loyalty:    func() *string { s := "3"; return &s }(),
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"LoyaltyAbilities: []game.LoyaltyAbility",
		"LoyaltyCost: 1,",
		"LoyaltyCost: -2,",
		"game.Draw",
		"game.GainLife",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceModalChooseOneSpell(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Charm",
		Layout:     "normal",
		ManaCost:   "{G}{W}",
		TypeLine:   "Instant",
		OracleText: "Choose one \u2014\n\u2022 Draw a card.\n\u2022 You gain 3 life.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"SpellAbility:",
		"MinModes: 1,",
		"MaxModes: 1,",
		"game.Draw",
		"game.GainLife",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceAdventureFacePreservation(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:          "Pond Guardian // Rippling Insight",
		Layout:        "adventure",
		ColorIdentity: []string{"U"},
		CardFaces: []ScryfallCardFace{
			{
				Name:       "Pond Guardian",
				ManaCost:   "{2}{U}",
				TypeLine:   "Creature — Merfolk Wizard",
				OracleText: "Flying",
				Power:      new("2"),
				Toughness:  new("3"),
			},
			{
				Name:       "Rippling Insight",
				ManaCost:   "{1}{U}",
				TypeLine:   "Instant — Adventure",
				OracleText: "Draw a card.",
			},
		},
	}

	source, diagnostics, err := GenerateExecutableCardSource(card, "p")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		`Name: "Pond Guardian",`,
		"cost.O(2),",
		"cost.U,",
		"types.Creature",
		`Name: "Rippling Insight",`,
		"cost.O(1),",
		"types.Instant",
		"types.Adventure",
		"Alternate: opt.Val(game.CardFace{",
		"Layout: game.LayoutAdventure,",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceAdventureRejectsWhenAnyFaceUnsupported(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:          "Pond Guardian // Impossible Lesson",
		Layout:        "adventure",
		ColorIdentity: []string{"U"},
		CardFaces: []ScryfallCardFace{
			{
				Name:      "Pond Guardian",
				ManaCost:  "{2}{U}",
				TypeLine:  "Creature — Merfolk Wizard",
				Power:     new("2"),
				Toughness: new("3"),
			},
			{
				Name:       "Impossible Lesson",
				ManaCost:   "{1}{U}",
				TypeLine:   "Sorcery — Adventure",
				OracleText: "Start your engines!",
			},
		},
	}

	source, diagnostics, err := GenerateExecutableCardSource(card, "p")
	if err != nil {
		t.Fatal(err)
	}
	if source != "" {
		t.Fatalf("source = %q, want no partial card", source)
	}
	if len(diagnostics) == 0 {
		t.Fatal("expected diagnostics for unsupported adventure face, got none")
	}
}

func TestGenerateExecutableCardSourceAdventureColorsFromManaCost(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:          "Dawn Escort // Guiding Prayer",
		Layout:        "adventure",
		ColorIdentity: []string{"W"},
		CardFaces: []ScryfallCardFace{
			{
				Name:       "Dawn Escort",
				ManaCost:   "{2}{W}",
				TypeLine:   "Creature — Human Knight",
				OracleText: "Vigilance",
				Power:      new("2"),
				Toughness:  new("2"),
			},
			{
				Name:       "Guiding Prayer",
				ManaCost:   "{1}{W}",
				TypeLine:   "Sorcery — Adventure",
				OracleText: "You gain 3 life.",
			},
		},
	}

	source, diagnostics, err := GenerateExecutableCardSource(card, "d")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if got := strings.Count(source, "[]color.Color{color.White}"); got != 2 {
		t.Fatalf("white face colors count = %d, want 2:\n%s", got, source)
	}
}

func TestGenerateExecutableCardSourceSplit(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:          "Spark // Shelter",
		Layout:        "split",
		ColorIdentity: []string{"R", "W"},
		CardFaces: []ScryfallCardFace{
			{
				Name:       "Spark",
				ManaCost:   "{R}",
				TypeLine:   "Instant",
				OracleText: "Spark deals 2 damage to any target.",
			},
			{
				Name:       "Shelter",
				ManaCost:   "{1}{W}",
				TypeLine:   "Instant",
				OracleText: "You gain 3 life.",
			},
		},
	}

	source, diagnostics, err := GenerateExecutableCardSource(card, "s")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		`Name: "Spark",`,
		"cost.R,",
		`Name: "Shelter",`,
		"cost.O(1),",
		"cost.W,",
		"Alternate: opt.Val(game.CardFace{",
		"Layout: game.LayoutSplit,",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceSplitRejectsUnsupported(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:          "Spark // Wild Math",
		Layout:        "split",
		ColorIdentity: []string{"R", "U"},
		CardFaces: []ScryfallCardFace{
			{
				Name:       "Spark",
				ManaCost:   "{R}",
				TypeLine:   "Instant",
				OracleText: "Spark deals 2 damage to any target.",
			},
			{
				Name:       "Wild Math",
				ManaCost:   "{X}{U}",
				TypeLine:   "Sorcery",
				OracleText: "Surveil X.",
			},
		},
	}

	source, diagnostics, err := GenerateExecutableCardSource(card, "s")
	if err != nil {
		t.Fatal(err)
	}
	if source != "" {
		t.Fatalf("source = %q, want no partial card", source)
	}
	if len(diagnostics) == 0 {
		t.Fatal("expected diagnostics for unsupported split half, got none")
	}
}

func TestGenerateExecutableCardSourcePrepare(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:          "Shieldmate // Ready Formation",
		Layout:        "prepare",
		ColorIdentity: []string{"W"},
		CardFaces: []ScryfallCardFace{
			{
				Name:       "Shieldmate",
				ManaCost:   "{2}{W}",
				TypeLine:   "Creature — Human Soldier",
				Power:      new("3"),
				Toughness:  new("3"),
				OracleText: "This creature enters prepared. (While it's prepared, you may cast a copy of its spell. Doing so unprepares it.)",
			},
			{
				Name:       "Ready Formation",
				ManaCost:   "{W}",
				TypeLine:   "Sorcery",
				OracleText: "Target creature gets +2/+2 until end of turn.",
			},
		},
	}

	source, diagnostics, err := GenerateExecutableCardSource(card, "s")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Layout: game.LayoutPrepare,",
		"EntersPrepared: true,",
		"Alternate: opt.Val(game.CardFace{",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceRejectsAlternateLayoutsWithMoreThanTwoFaces(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:          "Triple Trouble",
		Layout:        "split",
		ColorIdentity: []string{"U", "R", "W"},
		CardFaces: []ScryfallCardFace{
			{Name: "First", ManaCost: "{U}", TypeLine: "Instant", OracleText: "Draw a card."},
			{Name: "Second", ManaCost: "{R}", TypeLine: "Instant", OracleText: "You gain 3 life."},
			{Name: "Third", ManaCost: "{W}", TypeLine: "Instant", OracleText: "Draw a card."},
		},
	}

	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if source != "" {
		t.Fatalf("source = %q, want no partial card", source)
	}
	if len(diagnostics) != 1 || diagnostics[0].Summary != "unsupported card layout" {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if !strings.Contains(diagnostics[0].Detail, `supports at most 2 faces for "split" layout cards, found 3`) {
		t.Fatalf("diagnostic detail = %q", diagnostics[0].Detail)
	}
}

func TestGenerateExecutableCardSourceSheoldredApocalypse(t *testing.T) {
	t.Parallel()
	// Sheoldred, the Apocalypse: real oracle text uses "they lose 2 life" for opponent draw trigger.
	card := &ScryfallCard{
		Name:       "Sheoldred, the Apocalypse",
		Layout:     "normal",
		ManaCost:   "{2}{B}{B}",
		TypeLine:   "Legendary Creature — Phyrexian Praetor",
		OracleText: "Deathtouch\nWhenever you draw a card, you gain 2 life.\nWhenever an opponent draws a card, they lose 2 life.",
		Colors:     []string{"B"},
		Power:      new("4"),
		Toughness:  new("5"),
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "s")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"game.EventCardDrawn",
		"game.TriggerPlayerYou",
		"game.TriggerPlayerOpponent",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceDyingToServe(t *testing.T) {
	t.Parallel()
	// Dying to Serve uses "discard one or more cards" (OneOrMore: true) trigger.
	// The real card creates tokens, which lowerSpell does not yet support (token
	// creation support is a separate follow-up). This test uses a supported effect
	// to verify the trigger pattern is lowered correctly.
	card := &ScryfallCard{
		Name:       "Dying to Serve",
		Layout:     "normal",
		ManaCost:   "{2}{B}",
		TypeLine:   "Enchantment",
		OracleText: "Whenever you discard one or more cards, you lose 1 life.",
		Colors:     []string{"B"},
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "d")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"game.EventCardDiscarded",
		"game.TriggerPlayerYou",
		"OneOrMore: true",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceTourachDreadCantor(t *testing.T) {
	t.Parallel()
	// Tourach, Dread Cantor: discard trigger for opponent + unsupported enter-when-kicked trigger.
	// The discard trigger "Whenever an opponent discards a card" should lower correctly.
	// The "+1/+1 counter on Tourach" effect is a self-name reference that is not yet supported by
	// lowerCounterPlacementSpell, so the card will still have diagnostics — but none for the discard
	// trigger phrase itself.
	card := &ScryfallCard{
		Name:       "Tourach, Dread Cantor",
		Layout:     "normal",
		ManaCost:   "{1}{B}",
		TypeLine:   "Legendary Creature — Human Cleric",
		OracleText: "Kicker {B}{B}\nProtection from white\nWhen Tourach, Dread Cantor enters, if it was kicked, target opponent discards two cards.\nWhenever an opponent discards a card, put a +1/+1 counter on Tourach, Dread Cantor.",
		Colors:     []string{"B"},
		Power:      new("2"),
		Toughness:  new("1"),
	}
	_, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	// The trigger phrase is recognized even though its self-name counter-placement
	// body remains unsupported.
	foundBodyDiagnostic := false
	for _, d := range diagnostics {
		if d.Summary == "unsupported draw/discard trigger effect" {
			foundBodyDiagnostic = true
			if strings.Contains(d.Detail, "unrecognized draw/discard trigger event phrase") {
				t.Fatalf("trigger phrase unexpectedly unrecognized: %#v", d)
			}
		}
	}
	if !foundBodyDiagnostic {
		t.Fatalf("diagnostics = %#v, want unsupported draw/discard trigger effect", diagnostics)
	}
}
