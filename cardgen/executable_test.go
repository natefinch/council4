package cardgen

import (
	"strings"
	"testing"
)

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

func TestGenerateExecutableCardSourceFuse(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:          "Wear // Tear",
		Layout:        "split",
		ColorIdentity: []string{"R", "W"},
		CardFaces: []ScryfallCardFace{
			{
				Name:       "Wear",
				ManaCost:   "{1}{R}",
				TypeLine:   "Instant",
				OracleText: "Destroy target artifact.\nFuse (You may cast one or both halves of this card from your hand.)",
			},
			{
				Name:       "Tear",
				ManaCost:   "{W}",
				TypeLine:   "Instant",
				OracleText: "Destroy target enchantment.\nFuse (You may cast one or both halves of this card from your hand.)",
			},
		},
	}

	source, diagnostics, err := GenerateExecutableCardSource(card, "w")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if source == "" {
		t.Fatal("source = empty, want generated card")
	}
	// Both halves carry the Fuse keyword static body, and the split layout keeps
	// Tear as the alternate face.
	if strings.Count(source, "game.FuseStaticBody") != 2 {
		t.Fatalf("expected Fuse static body on both halves:\n%s", source)
	}
	for _, wanted := range []string{
		`Name: "Wear",`,
		`Name: "Tear",`,
		"Layout: game.LayoutSplit,",
		"Alternate: opt.Val(game.CardFace{",
		"PermanentTypes: []types.Card{types.Artifact}",
		"PermanentTypes: []types.Card{types.Enchantment}",
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
	// The discard trigger "Whenever an opponent discards a card, put a +1/+1 counter on
	// Tourach, Dread Cantor" lowers correctly: the self-name reference with an internal comma
	// resolves to the source so the counter placement is recognized. The card still has a
	// diagnostic for the unsupported enter-when-kicked trigger.
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
	// The discard trigger phrase is recognized and its self-name counter-placement
	// body now lowers, so neither an unrecognized-trigger diagnostic nor an
	// unsupported-counter-placement diagnostic should appear. The remaining blocker
	// is the enter-when-kicked trigger, which emits no authoritative event.
	for _, d := range diagnostics {
		if d.Summary == "unsupported draw/discard trigger" &&
			strings.Contains(d.Detail, "unrecognized draw/discard trigger event phrase") {
			t.Fatalf("trigger phrase unexpectedly unrecognized: %#v", d)
		}
		if d.Summary == "unsupported counter placement" {
			t.Fatalf("counter placement on comma self-name unexpectedly unsupported: %#v", d)
		}
	}
}

func TestGenerateExecutableCardSourceModalActivatedAbility(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Console",
		Layout:     "normal",
		TypeLine:   "Artifact",
		OracleText: "{1}, Discard a card: Choose one —\n• Draw a card.\n• You gain 3 life.",
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"game.ActivatedAbility{",
		"AdditionalDiscard",
		"MinModes: 1",
		"MaxModes: 1",
		"game.Draw{",
		"game.GainLife{",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}
