package cardgen

import (
	"strings"
	"testing"
)

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
		{
			name:       "Ramosian Commander",
			oracleText: "Search your library for a Rebel permanent card with mana value 5 or less, put it onto the battlefield, then shuffle.",
			wants: []string{
				"zone.Battlefield",
				"Permanent:    true",
				"SubtypesAny:  []types.Sub{types.Rebel}",
				"MaxManaValue: opt.Val(5)",
			},
		},
		{
			name:       "Trinket Mage tutor",
			oracleText: "Search your library for an artifact card with mana value 1 or less, reveal it, put it into your hand, then shuffle.",
			wants: []string{
				"zone.Hand",
				"opt.Val(types.Artifact)",
				"MaxManaValue: opt.Val(1)",
				"Reveal",
			},
		},
		{
			name:       "Time of Need",
			oracleText: "Search your library for a legendary creature card, reveal it, put it into your hand, then shuffle.",
			wants: []string{
				"zone.Hand",
				"opt.Val(types.Creature)",
				"opt.Val(types.Legendary)",
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
		"Search your library for a green creature card, put it into your hand, then shuffle.",
		"Search your library for up to two basic land cards with different names, put them onto the battlefield tapped, then shuffle.",
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

// TestGenerateExecutableCardSourceEnterTriggerLibrarySearch covers the
// enter-the-battlefield ramp tutors (Wood Elves, Rune-Scarred Demon, and the
// basic-land monuments) whose library-search clause is embedded after the
// trigger condition and so reaches the parser with a lowercase "search" verb.
// Each must lower to a triggered ability carrying the same Search primitive as a
// sentence-initial controller tutor.
func TestGenerateExecutableCardSourceEnterTriggerLibrarySearch(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		typeLine   string
		oracleText string
		wants      []string
	}{
		{
			name:       "Wood Elves",
			typeLine:   "Creature — Elf Scout",
			oracleText: "When this creature enters, search your library for a Forest card, put that card onto the battlefield, then shuffle.",
			wants: []string{
				"zone.Battlefield",
				"SubtypesAny: []types.Sub{types.Forest}",
			},
		},
		{
			name:       "Rune-Scarred Demon",
			typeLine:   "Creature — Demon",
			oracleText: "When this creature enters, search your library for a card, put it into your hand, then shuffle.",
			wants: []string{
				"zone.Hand",
				"Amount: game.Fixed(1)",
			},
		},
		{
			name:       "Basic Monument",
			typeLine:   "Artifact",
			oracleText: "When this artifact enters, search your library for a basic Island, Mountain, or Plains card, reveal it, put it into your hand, then shuffle.",
			wants: []string{
				"zone.Hand",
				"opt.Val(types.Basic)",
				"[]types.Sub{types.Island, types.Mountain, types.Plains}",
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
				TypeLine:   test.typeLine,
				OracleText: test.oracleText,
				Power:      new("1"),
				Toughness:  new("1"),
			}, "w")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			for _, want := range append([]string{
				"TriggeredAbilities: []game.TriggeredAbility",
				"game.EventPermanentEnteredBattlefield",
				"game.TriggerSourceSelf",
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

// TestGenerateExecutableCardSourceEnterTriggerSearchFailsClosed confirms the
// embedded lowercase search relaxes nothing else: an unmodeled color filter in a
// triggered tutor still fails closed rather than lowering to a wrong predicate.
func TestGenerateExecutableCardSourceEnterTriggerSearchFailsClosed(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Unsupported Trigger Search",
		Layout:     "normal",
		TypeLine:   "Creature — Elf",
		OracleText: "When this creature enters, search your library for a green creature card, put it onto the battlefield, then shuffle.",
		Power:      new("1"),
		Toughness:  new("1"),
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
				want = strings.Replace(want, "RequirePermanentTypes:", "RequiredTypes:", 1)
				want = strings.Replace(want, "RequireNonToken:", "NonToken:", 1)
				if !strings.Contains(source, want) {
					t.Fatalf("source missing %q:\n%s", want, source)
				}
			}
			if !strings.Contains(source, "SubjectSelection: game.Selection{") &&
				!strings.Contains(test.oracleText, "permanent enters") {
				t.Fatalf("source missing semantic SubjectSelection:\n%s", source)
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

func TestGenerateExecutableCardSourceEnterOrAttackUnionTrigger(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Sun Titan",
		Layout:     "normal",
		ManaCost:   "{4}{W}{W}",
		TypeLine:   "Creature — Giant",
		OracleText: "Vigilance\nWhenever this creature enters or attacks, you may return target permanent card with mana value 3 or less from your graveyard to the battlefield.",
		Colors:     []string{"W"},
		Power:      new("6"),
		Toughness:  new("6"),
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "s")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"game.EventPermanentEnteredBattlefield",
		"UnionEvent: game.EventAttackerDeclared",
		"game.TriggerSourceSelf",
		"game.PutOnBattlefield",
		"zone.Graveyard",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}
