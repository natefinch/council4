package cardgen

import (
	"strings"
	"testing"
)

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
	if len(diagnostics) != 1 || diagnostics[0].Summary != "unsupported optional effect" {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
}

// TestGenerateExecutableCardSourceSupportsOptionalKickedEnterTrigger verifies
// that an enter trigger with a kicker intervening condition whose body is a
// single optional effect ("if it was kicked, you may draw a card") lowers: the
// trigger keeps the kicker intervening condition and the draw instruction is
// marked Optional so the controller is asked whether to draw on resolution.
func TestGenerateExecutableCardSourceSupportsOptionalKickedEnterTrigger(t *testing.T) {
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
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v, want none", diagnostics)
	}
	if source == "" {
		t.Fatal("source = empty, want generated card")
	}
	if !strings.Contains(source, "InterveningIfEventPermanentWasKicked: true") {
		t.Fatalf("source missing kicker intervening condition:\n%s", source)
	}
	if !strings.Contains(source, "game.Draw{") || !strings.Contains(source, "Optional: true") {
		t.Fatalf("source missing optional draw:\n%s", source)
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
		{name: "restricted mana choice", cardName: "Fellwar Stone", typeLine: "Artifact", oracleText: "{T}: Add one mana of any color that a land you control could produce."},
		{name: "unsupported conditional tapped entry", cardName: "Test Land", typeLine: "Land", oracleText: "This land enters tapped unless you gained life this turn."},
		{name: "nonmana ward", cardName: "Test Bear", typeLine: "Creature — Bear", oracleText: "Ward—Pay 2 life."},
		{name: "typecycling", cardName: "Test Card", typeLine: "Sorcery", oracleText: "Plainscycling {2}"},
		{name: "nonmana equip", cardName: "Test Equipment", typeLine: "Artifact — Equipment", oracleText: "Equip—Pay {3} or discard a card."},
		{name: "qualified equip", cardName: "Test Equipment", typeLine: "Artifact — Equipment", oracleText: "Equip creature token {1}"},
		{name: "qualified enchant", cardName: "Test Aura", typeLine: "Enchantment — Aura", oracleText: "Enchant creature you control"},
		{name: "noncolor protection (replaced by support test)", cardName: "Test Bear", typeLine: "Creature — Bear", oracleText: "Protection from a chosen color"},
		{name: "divided damage variable amount", cardName: "Test Bolt", typeLine: "Instant", oracleText: "Test Bolt deals X damage divided as you choose among any number of targets."},
		{name: "variable surveil", cardName: "Test Surveil", typeLine: "Sorcery", oracleText: "Surveil X."},
		{name: "repeated proliferate", cardName: "Test Proliferate", typeLine: "Sorcery", oracleText: "Proliferate X times."},
		{name: "scry with unrecognized sibling", cardName: "Test Vision", typeLine: "Sorcery", oracleText: "Scry 1, then celebrate."},
		{name: "surveil with unrecognized sibling", cardName: "Test Vision", typeLine: "Sorcery", oracleText: "Surveil 1, then celebrate."},
		{name: "investigate with unrecognized sibling", cardName: "Test Clue", typeLine: "Sorcery", oracleText: "Investigate, then celebrate."},
		{name: "proliferate with unrecognized sibling", cardName: "Test Counter", typeLine: "Sorcery", oracleText: "Proliferate, then celebrate."},
		{name: "another fight target", cardName: "Test Fight", typeLine: "Sorcery", oracleText: "Target creature fights another target creature."},
		{name: "conditional draw", cardName: "Test Draw", typeLine: "Sorcery", oracleText: "If you control a creature, draw two cards."},
		{name: "conditional destroy", cardName: "Test Doom", typeLine: "Instant", oracleText: "If it is tapped, destroy target creature."},
		{name: "regeneration destroy", cardName: "Test Doom", typeLine: "Instant", oracleText: "Destroy target creature. It can't be regenerated."},
		{name: "restricted destroy", cardName: "Test Doom", typeLine: "Instant", oracleText: "Destroy target nonblack nonred creature."},
		{name: "linked exile", cardName: "Test Exile", typeLine: "Instant", oracleText: "Exile target creature, then return it to the battlefield."},
		{name: "graveyard exile", cardName: "Test Exile", typeLine: "Instant", oracleText: "Exile target card from a graveyard."},
		{name: "bounce to your hand", cardName: "Test Bounce", typeLine: "Instant", oracleText: "Return target creature to your hand."},
		{name: "variable power toughness", cardName: "Test Growth", typeLine: "Instant", oracleText: "Target creature gets +X/+X until end of turn."},
		{name: "permanent power toughness", cardName: "Test Growth", typeLine: "Sorcery", oracleText: "Target creature gets +2/+2."},
		{name: "dynamic group power toughness", cardName: "Test Growth", typeLine: "Instant", oracleText: "Creatures you control get +X/+X until end of turn."},
		{name: "unsupported group power toughness duration", cardName: "Test Growth", typeLine: "Instant", oracleText: "Creatures you control get +2/+2 until your next turn."},
		{name: "group power toughness with unrecognized sibling", cardName: "Test Growth", typeLine: "Instant", oracleText: "Creatures you control get +2/+2 until end of turn, then celebrate."},
		{name: "group keyword grant", cardName: "Test Flight", typeLine: "Instant", oracleText: "Creatures you control gain flying until end of turn."},
		{name: "parameterized temporary keyword", cardName: "Test Ward", typeLine: "Instant", oracleText: "Target creature gains ward {2} until end of turn."},
		{name: "set life total", cardName: "Test Life", typeLine: "Sorcery", oracleText: "Your life total becomes 10."},
		{name: "variable scry", cardName: "Test Vision", typeLine: "Sorcery", oracleText: "Scry X."},
		{name: "conditional scry", cardName: "Test Vision", typeLine: "Sorcery", oracleText: "If you control a creature, scry 2."},
		{name: "targeted scry", cardName: "Test Vision", typeLine: "Sorcery", oracleText: "Target player scries 2."},
		{name: "random discard", cardName: "Test Mind", typeLine: "Sorcery", oracleText: "Target player discards a card at random."},
		{name: "named discard", cardName: "Test Mind", typeLine: "Sorcery", oracleText: "Target player discards a creature card."},
		{name: "hand discard", cardName: "Test Mind", typeLine: "Sorcery", oracleText: "Target player discards their hand."},
		{name: "mass tap", cardName: "Test Sleep", typeLine: "Sorcery", oracleText: "Tap all creatures."},
		{name: "gain control with unrecognized sibling", cardName: "Test Theft", typeLine: "Sorcery", oracleText: "Gain control of target creature until end of turn, then celebrate."},
		{name: "gain control with negated untap", cardName: "Test Theft", typeLine: "Sorcery", oracleText: "Gain control of target creature until end of turn. Don't untap it."},
		{name: "multiple effects with unrecognized sibling", cardName: "Test Command", typeLine: "Sorcery", oracleText: "Destroy target artifact. Draw a card. Celebrate."},
		{name: "multiple effects with unknown sibling", cardName: "Test Command", typeLine: "Sorcery", oracleText: "Destroy target artifact. Draw a card. Perform a ritual."},
		{name: "mass destroy with unknown keyword filter", cardName: "Test Purge", typeLine: "Sorcery", oracleText: "Destroy all creatures with celebrate."},
		{name: "unsupported tap qualifier", cardName: "Test Sleep", typeLine: "Instant", oracleText: "Tap target creature with flying."},
		{name: "freeze tap", cardName: "Test Sleep", typeLine: "Instant", oracleText: "Tap target creature. It doesn't untap during its controller's next untap step."},
		{name: "conditional untap", cardName: "Test Sleep", typeLine: "Instant", oracleText: "If it is tapped, untap target creature."},
		{name: "until mill", cardName: "Test Mill", typeLine: "Sorcery", oracleText: "Target player mills cards until they mill a land card."},
		{name: "reveal mill", cardName: "Test Mill", typeLine: "Sorcery", oracleText: "Target player reveals and mills three cards."},
		{name: "mass mill", cardName: "Test Mill", typeLine: "Sorcery", oracleText: "Each opponent mills three cards."},
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
			oracleText: "Create a Powerstone token.",
			summary:    "unsupported token creation",
			detail:     "fixed-power/toughness creature token",
		},
		"activated": {
			typeLine:   "Creature — Bear",
			oracleText: "Remove a +1/+1 counter from target creature: Draw a card.",
			summary:    "unsupported activation cost",
			detail:     "cannot lower every typed activation cost component",
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
