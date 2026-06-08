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
		"AdditionalCosts: cost.Tap",
		"Primitive: game.AddMana",
		"game.Fixed(1)",
		"ManaColor: mana.G",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
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
		"Recipient: game.TargetRecipient(0)",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
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
				"TargetIndex: game.TargetIndexController",
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
		"TargetIndex: 0",
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
		"TargetIndex: 0",
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
		"TargetIndex: game.TargetIndexController",
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
		{text: "You lose 2 life.", primitive: "game.LoseLife", recipient: "game.TargetIndexController"},
		{text: "Target player gains 4 life.", primitive: "game.GainLife", recipient: "TargetIndex: 0"},
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
				"TargetIndex: game.TargetIndexController",
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
		"TargetIndex: 0",
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
		"TargetIndex: game.TargetIndexController",
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
		"TargetIndex: 0",
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
		"TargetIndex: 0",
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
		"TargetIndex: game.TargetIndexController",
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
		"TargetIndex: game.TargetIndexController",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
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
		"TargetIndex: 0",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
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
		{name: "multiple mana", cardName: "Mana Bear", typeLine: "Creature — Bear", oracleText: "{T}: Add {G}{G}."},
		{name: "mana choice", cardName: "Mana Bear", typeLine: "Creature — Bear", oracleText: "{T}: Add one mana of any color."},
		{name: "variable damage", cardName: "Test Bolt", typeLine: "Instant", oracleText: "Test Bolt deals X damage to any target."},
		{name: "divided damage", cardName: "Test Bolt", typeLine: "Instant", oracleText: "Test Bolt deals 3 damage divided as you choose among any number of targets."},
		{name: "mass damage", cardName: "Test Bolt", typeLine: "Instant", oracleText: "Test Bolt deals 3 damage to each opponent."},
		{name: "variable draw", cardName: "Test Draw", typeLine: "Sorcery", oracleText: "Draw X cards."},
		{name: "conditional draw", cardName: "Test Draw", typeLine: "Sorcery", oracleText: "If you control a creature, draw two cards."},
		{name: "compound draw", cardName: "Test Draw", typeLine: "Sorcery", oracleText: "Draw two cards, then discard a card."},
		{name: "conditional destroy", cardName: "Test Doom", typeLine: "Instant", oracleText: "If it is tapped, destroy target creature."},
		{name: "mass destroy", cardName: "Test Doom", typeLine: "Sorcery", oracleText: "Destroy all creatures."},
		{name: "regeneration destroy", cardName: "Test Doom", typeLine: "Instant", oracleText: "Destroy target creature. It can't be regenerated."},
		{name: "restricted destroy", cardName: "Test Doom", typeLine: "Instant", oracleText: "Destroy target nonblack creature."},
		{name: "variable life", cardName: "Test Life", typeLine: "Sorcery", oracleText: "You gain X life."},
		{name: "each opponent life", cardName: "Test Life", typeLine: "Sorcery", oracleText: "Each opponent loses 3 life."},
		{name: "set life total", cardName: "Test Life", typeLine: "Sorcery", oracleText: "Your life total becomes 10."},
		{name: "compound life", cardName: "Test Life", typeLine: "Sorcery", oracleText: "You gain 3 life and draw a card."},
		{name: "variable scry", cardName: "Test Vision", typeLine: "Sorcery", oracleText: "Scry X."},
		{name: "conditional scry", cardName: "Test Vision", typeLine: "Sorcery", oracleText: "If you control a creature, scry 2."},
		{name: "compound scry", cardName: "Test Vision", typeLine: "Sorcery", oracleText: "Scry 2, then draw a card."},
		{name: "targeted scry", cardName: "Test Vision", typeLine: "Sorcery", oracleText: "Target player scries 2."},
		{name: "random discard", cardName: "Test Mind", typeLine: "Sorcery", oracleText: "Target player discards a card at random."},
		{name: "named discard", cardName: "Test Mind", typeLine: "Sorcery", oracleText: "Target player discards a creature card."},
		{name: "hand discard", cardName: "Test Mind", typeLine: "Sorcery", oracleText: "Target player discards their hand."},
		{name: "optional discard", cardName: "Test Mind", typeLine: "Sorcery", oracleText: "You may discard a card."},
		{name: "compound discard", cardName: "Test Mind", typeLine: "Sorcery", oracleText: "Discard a card, then draw a card."},
		{name: "mass tap", cardName: "Test Sleep", typeLine: "Sorcery", oracleText: "Tap all creatures."},
		{name: "optional tap", cardName: "Test Sleep", typeLine: "Instant", oracleText: "You may tap target creature."},
		{name: "qualified tap", cardName: "Test Sleep", typeLine: "Instant", oracleText: "Tap target untapped creature."},
		{name: "freeze tap", cardName: "Test Sleep", typeLine: "Instant", oracleText: "Tap target creature. It doesn't untap during its controller's next untap step."},
		{name: "conditional untap", cardName: "Test Sleep", typeLine: "Instant", oracleText: "If it is tapped, untap target creature."},
		{name: "variable mill", cardName: "Test Mill", typeLine: "Sorcery", oracleText: "Target player mills X cards."},
		{name: "until mill", cardName: "Test Mill", typeLine: "Sorcery", oracleText: "Target player mills cards until they mill a land card."},
		{name: "reveal mill", cardName: "Test Mill", typeLine: "Sorcery", oracleText: "Target player reveals and mills three cards."},
		{name: "compound mill", cardName: "Test Mill", typeLine: "Sorcery", oracleText: "Target player mills three cards, then draws a card."},
		{name: "mass mill", cardName: "Test Mill", typeLine: "Sorcery", oracleText: "Each opponent mills three cards."},
		{name: "optional enter", cardName: "Test Bear", typeLine: "Creature — Bear", oracleText: "When this creature enters, you may draw a card."},
		{name: "intervening enter", cardName: "Test Bear", typeLine: "Creature — Bear", oracleText: "When this creature enters, if you control an artifact, draw a card."},
		{name: "other enter", cardName: "Test Bear", typeLine: "Creature — Bear", oracleText: "Whenever another creature enters, draw a card."},
		{name: "leave trigger", cardName: "Test Bear", typeLine: "Creature — Bear", oracleText: "When this creature leaves the battlefield, draw a card."},
		{name: "cast trigger", cardName: "Test Bear", typeLine: "Creature — Bear", oracleText: "When you cast this spell, draw a card."},
		{name: "compound enter", cardName: "Test Bear", typeLine: "Creature — Bear", oracleText: "When this creature enters, draw a card, then discard a card."},
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

func TestGenerateExecutableCardSourceRejectsPartialAbility(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Drawing Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		OracleText: "When this creature enters, draw a card, then discard a card.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "d")
	if err != nil {
		t.Fatal(err)
	}
	if source != "" || len(diagnostics) == 0 {
		t.Fatalf("source = %q, diagnostics = %#v", source, diagnostics)
	}
	if got := diagnostics[0].Summary; got != "unsupported enter trigger" {
		t.Fatalf("summary = %q", got)
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
			oracleText: "Exile target creature.",
			summary:    "unsupported spell ability",
			detail:     "does not yet lower this spell ability",
		},
		"activated": {
			typeLine:   "Creature — Bear",
			oracleText: "{1}: Draw a card.",
			summary:    "unsupported activated ability",
			detail:     "supports only exact single-color tap mana abilities",
		},
		"static rules text": {
			typeLine:   "Enchantment",
			oracleText: "Creatures you control get +1/+1.",
			summary:    "unsupported static ability",
			detail:     "non-keyword static rules text",
		},
		"parameterized keyword": {
			typeLine:   "Creature — Snake",
			oracleText: "Toxic 1",
			summary:    "unsupported parameterized keyword",
			detail:     `Toxic with parameter "1"`,
		},
		"keyword without template": {
			typeLine:   "Creature — Dinosaur",
			oracleText: "Ward",
			summary:    "unsupported keyword ability",
			detail:     "no reusable game template for Ward",
		},
		"modal": {
			typeLine:   "Sorcery",
			oracleText: "Choose one —\n• Draw a card.\n• Destroy target creature.",
			summary:    "unsupported modal ability",
			detail:     "does not yet lower modal abilities",
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
