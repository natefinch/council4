package cardgen

import (
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

func TestGenerateExecutableCardSourceSelfCannotAttack(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Pacifist Bear",
		Layout:     "normal",
		ManaCost:   "{1}{W}",
		TypeLine:   "Creature — Bear",
		OracleText: "This creature can't attack.",
		Colors:     []string{"W"},
		Power:      new("3"),
		Toughness:  new("3"),
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "p")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if !strings.Contains(source, "game.CantAttackStaticBody") {
		t.Fatalf("source missing cannot-attack static body:\n%s", source)
	}
}

func TestGenerateExecutableCardSourceSelfMustBeBlocked(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Provoking Bear",
		Layout:     "normal",
		ManaCost:   "{2}{R}",
		TypeLine:   "Creature — Bear",
		OracleText: "This creature must be blocked if able.",
		Colors:     []string{"R"},
		Power:      new("3"),
		Toughness:  new("2"),
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "p")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if !strings.Contains(source, "game.MustBeBlockedStaticBody") {
		t.Fatalf("source missing must-be-blocked static body:\n%s", source)
	}
}

func TestGenerateExecutableCardSourceRejectsConditionalCannotAttack(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Conditional Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		OracleText: "This creature can't attack unless defending player controls an Island.",
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

func TestGenerateExecutableCardSourceRejectsMustBeBlockedWithoutIfAble(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Conditional Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		OracleText: "This creature must be blocked by two or more creatures.",
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

func TestGenerateExecutableCardSourcePacifismAttachedCantAttackOrBlock(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Pacifism",
		Layout:     "normal",
		ManaCost:   "{1}{W}",
		TypeLine:   "Enchantment — Aura",
		OracleText: "Enchant creature\nEnchanted creature can't attack or block.",
		Colors:     []string{"W"},
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "p")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Kind:             game.RuleEffectCantAttack,",
		"Kind:             game.RuleEffectCantBlock,",
		"AffectedAttached: true,",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
	if strings.Contains(source, "AffectedSource: true,") {
		t.Fatalf("attached can't-attack-or-block rule must not set AffectedSource:\n%s", source)
	}
}

func TestGenerateExecutableCardSourceAttackTax(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Propaganda",
		Layout:     "normal",
		ManaCost:   "{2}{U}",
		TypeLine:   "Enchantment",
		OracleText: "Creatures can't attack you unless their controller pays {2} for each creature they control that's attacking you.",
		Colors:     []string{"U"},
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "p")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Kind:             game.RuleEffectAttackTax,",
		"AffectedPlayer:   game.PlayerYou,",
		"AttackTaxGeneric: 2,",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceSelfCantAttackOrBlock(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Subdued Bear",
		Layout:     "normal",
		ManaCost:   "{2}{W}",
		TypeLine:   "Creature — Bear",
		OracleText: "This creature can't attack or block.",
		Colors:     []string{"W"},
		Power:      new("3"),
		Toughness:  new("3"),
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "s")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if !strings.Contains(source, "game.CantAttackOrBlockStaticBody") {
		t.Fatalf("source missing can't-attack-or-block static body:\n%s", source)
	}
}

func TestGenerateExecutableCardSourceSelfDoesntUntap(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Farmstead Gleaner",
		Layout:     "normal",
		ManaCost:   "{3}",
		TypeLine:   "Artifact Creature — Scarecrow",
		OracleText: "This creature doesn't untap during your untap step.",
		Power:      new("2"),
		Toughness:  new("2"),
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "f")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Kind:           game.RuleEffectDoesntUntap,",
		"AffectedSource: true,",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceAttachedDoesntUntap(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Frozen Solid",
		Layout:     "normal",
		ManaCost:   "{2}{U}",
		TypeLine:   "Enchantment — Aura",
		OracleText: "Enchant creature\nEnchanted creature doesn't untap during its controller's untap step.",
		Colors:     []string{"U"},
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "f")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Kind:             game.RuleEffectDoesntUntap,",
		"AffectedAttached: true,",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceComposedSimpleStaticRuleVariants(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		typeLine   string
		oracleText string
		want       string
	}{
		"cannot block": {
			typeLine:   "Creature — Bear",
			oracleText: "This creature cannot block.",
			want:       "game.CantBlockStaticBody",
		},
		"cannot be blocked": {
			typeLine:   "Creature — Bear",
			oracleText: "This creature cannot be blocked.",
			want:       "game.CantBeBlockedStaticBody",
		},
		"explicit must attack": {
			typeLine:   "Creature — Bear",
			oracleText: "This creature must attack each combat if able.",
			want:       "game.MustAttackStaticBody",
		},
		"cannot be countered": {
			typeLine:   "Sorcery",
			oracleText: "This spell cannot be countered.\nDestroy target creature.",
			want:       "game.CantBeCounteredStaticBody",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
				Name:       "Test Card",
				Layout:     "normal",
				TypeLine:   test.typeLine,
				OracleText: test.oracleText,
				Power:      new("2"),
				Toughness:  new("2"),
			}, "t")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) != 0 || !strings.Contains(source, test.want) {
				t.Fatalf("source = %q, diagnostics = %#v, want %q", source, diagnostics, test.want)
			}
		})
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

func TestGenerateExecutableCardSourceAuraBasePowerToughness(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Shrink Aura",
		Layout:     "normal",
		ManaCost:   "{U}",
		TypeLine:   "Enchantment — Aura",
		OracleText: "Enchant creature\nEnchanted creature has base power and toughness 0/2.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "s")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Layer:        game.LayerPowerToughnessSet,",
		"SetPower:     opt.Val(game.PT{Value: 0}),",
		"SetToughness: opt.Val(game.PT{Value: 2}),",
		"game.AttachedObjectGroup(game.SourcePermanentReference())",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceAuraCharacteristicAddition(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Rotten Aura",
		Layout:     "normal",
		ManaCost:   "{B}",
		TypeLine:   "Enchantment — Aura",
		OracleText: "Enchant creature\nEnchanted creature gets -1/-1 and is a black Zombie in addition to its other colors and types.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "r")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Layer:     game.LayerColor,",
		"AddColors: []color.Color{color.Black},",
		"Layer:       game.LayerType,",
		"AddSubtypes: []types.Sub{types.Sub(\"Zombie\")},",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceAuraSetColor(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Darken Aura",
		Layout:     "normal",
		ManaCost:   "{B}",
		TypeLine:   "Enchantment — Aura",
		OracleText: "Enchant creature\nEnchanted creature gets +3/+1 and is black.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "d")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Layer:     game.LayerColor,",
		"SetColors: []color.Color{color.Black},",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceComposedPowerToughnessRule(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		typeLine   string
		oracleText string
		ruleKind   string
		affected   string
	}{
		"attached cant block": {
			typeLine:   "Enchantment — Aura",
			oracleText: "Enchant creature\nEnchanted creature gets +2/+2 and can't block.",
			ruleKind:   "game.RuleEffectCantBlock",
			affected:   "AffectedAttached: true",
		},
		"attached cant be blocked": {
			typeLine:   "Enchantment — Aura",
			oracleText: "Enchant creature\nEnchanted creature gets +1/+0 and can't be blocked.",
			ruleKind:   "game.RuleEffectCantBeBlocked",
			affected:   "AffectedAttached: true",
		},
		"attached must attack": {
			typeLine:   "Enchantment — Aura",
			oracleText: "Enchant creature\nEnchanted creature gets +2/+2 and attacks each combat if able.",
			ruleKind:   "game.RuleEffectMustAttack",
			affected:   "AffectedAttached: true",
		},
		"equipped cant block": {
			typeLine:   "Artifact — Equipment",
			oracleText: "Equipped creature gets +2/+2 and can't block.\nEquip {3}",
			ruleKind:   "game.RuleEffectCantBlock",
			affected:   "AffectedAttached: true",
		},
		"source cant block": {
			typeLine:   "Creature — Bear",
			oracleText: "This creature gets +2/+2 and can't block.",
			ruleKind:   "game.RuleEffectCantBlock",
			affected:   "AffectedSource: true",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{
				Name:       "Composed Test",
				Layout:     "normal",
				TypeLine:   test.typeLine,
				OracleText: test.oracleText,
			}
			if strings.Contains(test.typeLine, "Creature") {
				card.Power = new("1")
				card.Toughness = new("1")
			}
			source, diagnostics, err := GenerateExecutableCardSource(card, "c")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			for _, wanted := range []string{
				"game.LayerPowerToughnessModify",
				test.ruleKind,
				test.affected,
			} {
				if !strings.Contains(source, wanted) {
					t.Fatalf("source missing %q:\n%s", wanted, source)
				}
			}
		})
	}
}

// TestGenerateExecutableCardSourceComposedQualifiedRule covers the two
// bounded-exception rule operations: the defender-scoped "can't attack you or
// planeswalkers you control" (lowered to a can't-attack effect restricted to the
// controller) and "can't be blocked by more than one creature". Both appear only
// as the trailing operation of a composed declaration, so the test exercises the
// power/toughness and keyword-grant compound shapes the printed cards use.
func TestGenerateExecutableCardSourceComposedQualifiedRule(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		typeLine   string
		oracleText string
		wanted     []string
	}{
		"vow style cant attack you": {
			typeLine:   "Enchantment — Aura",
			oracleText: "Enchant creature\nEnchanted creature gets +1/+3 and can't attack you or planeswalkers you control.",
			wanted: []string{
				"game.LayerPowerToughnessModify",
				"game.RuleEffectCantAttack",
				"DefendingPlayer:",
				"game.PlayerYou,",
				"AffectedAttached: true",
			},
		},
		"power/toughness cant be blocked by more than one": {
			typeLine:   "Enchantment — Aura",
			oracleText: "Enchant creature\nEnchanted creature gets +1/+2 and can't be blocked by more than one creature.",
			wanted: []string{
				"game.LayerPowerToughnessModify",
				"game.RuleEffectCantBeBlockedByMoreThanOne",
				"AffectedAttached: true",
			},
		},
		"keyword cant be blocked by more than one": {
			typeLine:   "Artifact — Equipment",
			oracleText: "Equipped creature has trample and can't be blocked by more than one creature.\nEquip {2}",
			wanted: []string{
				"game.RuleEffectCantBeBlockedByMoreThanOne",
				"AffectedAttached: true",
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{
				Name:       "Qualified Test",
				Layout:     "normal",
				TypeLine:   test.typeLine,
				OracleText: test.oracleText,
			}
			source, diagnostics, err := GenerateExecutableCardSource(card, "q")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			for _, wanted := range test.wanted {
				if !strings.Contains(source, wanted) {
					t.Fatalf("source missing %q:\n%s", wanted, source)
				}
			}
		})
	}
}

// TestGenerateExecutableCardSourceSelfQualifiedRuleAndColor covers the
// self/source continuous statics added for issue #360: the source-creature
// "can't be blocked by more than one creature" prohibition (including the
// card-name self-reference subject) and "<source> is all colors", both lowered
// onto the source.
func TestGenerateExecutableCardSourceSelfQualifiedRuleAndColor(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		name       string
		typeLine   string
		oracleText string
		wanted     []string
	}{
		"this creature cant be blocked by more than one": {
			name:       "Norwood Test",
			typeLine:   "Creature — Elf",
			oracleText: "This creature can't be blocked by more than one creature.",
			wanted: []string{
				"game.RuleEffectCantBeBlockedByMoreThanOne",
				"AffectedSource: true",
			},
		},
		"this creature is all colors": {
			name:       "Citizen Test",
			typeLine:   "Creature — Citizen",
			oracleText: "This creature is all colors.",
			wanted: []string{
				"game.LayerColor",
				"AffectedSource: true",
				"SetColors:      []color.Color{color.White, color.Blue, color.Black, color.Red, color.Green}",
			},
		},
		"named source is all colors": {
			name:       "Transguild Courier",
			typeLine:   "Artifact Creature — Golem",
			oracleText: "Transguild Courier is all colors.",
			wanted: []string{
				"game.LayerColor",
				"AffectedSource: true",
				"SetColors:      []color.Color{color.White, color.Blue, color.Black, color.Red, color.Green}",
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{
				Name:       test.name,
				Layout:     "normal",
				TypeLine:   test.typeLine,
				OracleText: test.oracleText,
			}
			source, diagnostics, err := GenerateExecutableCardSource(card, "q")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			for _, wanted := range test.wanted {
				if !strings.Contains(source, wanted) {
					t.Fatalf("source missing %q:\n%s", wanted, source)
				}
			}
		})
	}
}

// TestGenerateExecutableCardSourceCantBeBlockedByCreaturesWith covers the
// bounded blocker-restriction prohibition family on both a source creature and
// an attached object, asserting the rendered rule effect and restriction.
func TestGenerateExecutableCardSourceCantBeBlockedByCreaturesWith(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		typeLine   string
		oracleText string
		power      string
		toughness  string
		wanted     []string
	}{
		"source flying": {
			typeLine:   "Creature — Insect",
			oracleText: "This creature can't be blocked by creatures with flying.",
			power:      "1",
			toughness:  "1",
			wanted: []string{
				"game.RuleEffectCantBeBlockedByCreaturesWith",
				"AffectedSource: true",
				"game.BlockerRestrictionFlying",
			},
		},
		"source power or less": {
			typeLine:   "Creature — Merfolk",
			oracleText: "This creature can't be blocked by creatures with power 2 or less.",
			power:      "2",
			toughness:  "2",
			wanted: []string{
				"game.RuleEffectCantBeBlockedByCreaturesWith",
				"AffectedSource: true",
				"game.BlockerRestrictionPowerLessOrEqual",
				"Power: 2,",
			},
		},
		"source power or greater": {
			typeLine:   "Creature — Kithkin",
			oracleText: "This creature can't be blocked by creatures with power 3 or greater.",
			power:      "1",
			toughness:  "1",
			wanted: []string{
				"game.RuleEffectCantBeBlockedByCreaturesWith",
				"AffectedSource: true",
				"game.BlockerRestrictionPowerGreaterOrEqual",
				"Power: 3,",
			},
		},
		"attached flying": {
			typeLine:   "Artifact — Equipment",
			oracleText: "Equipped creature gets +1/+0 and can't be blocked by creatures with flying.\nEquip {3}",
			wanted: []string{
				"game.RuleEffectCantBeBlockedByCreaturesWith",
				"AffectedAttached: true",
				"game.BlockerRestrictionFlying",
			},
		},
		"attached power or less": {
			typeLine:   "Artifact — Equipment",
			oracleText: "Equipped creature gets +3/+3 and can't be blocked by creatures with power 3 or less.\nEquip {3}",
			wanted: []string{
				"game.RuleEffectCantBeBlockedByCreaturesWith",
				"AffectedAttached: true",
				"game.BlockerRestrictionPowerLessOrEqual",
				"Power: 3,",
			},
		},
		"source color": {
			typeLine:   "Creature — Horse",
			oracleText: "This creature can't be blocked by black creatures.",
			power:      "3",
			toughness:  "3",
			wanted: []string{
				"game.RuleEffectCantBeBlockedByCreaturesWith",
				"AffectedSource: true",
				"game.BlockerRestrictionColor",
				"Color: color.Black,",
			},
		},
		"source artifact": {
			typeLine:   "Creature — Faerie",
			oracleText: "This creature can't be blocked by artifact creatures.",
			power:      "1",
			toughness:  "1",
			wanted: []string{
				"game.RuleEffectCantBeBlockedByCreaturesWith",
				"AffectedSource: true",
				"game.BlockerRestrictionArtifact",
			},
		},
		"attached color": {
			typeLine:   "Enchantment — Aura",
			oracleText: "Enchant creature\nEnchanted creature can't be blocked by white creatures.",
			wanted: []string{
				"game.RuleEffectCantBeBlockedByCreaturesWith",
				"AffectedAttached: true",
				"game.BlockerRestrictionColor",
				"Color: color.White,",
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{
				Name:       "Evasion Test",
				Layout:     "normal",
				TypeLine:   test.typeLine,
				OracleText: test.oracleText,
			}
			if test.power != "" {
				card.Power = new(test.power)
				card.Toughness = new(test.toughness)
			}
			source, diagnostics, err := GenerateExecutableCardSource(card, "e")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			for _, wanted := range test.wanted {
				if !strings.Contains(source, wanted) {
					t.Fatalf("source missing %q:\n%s", wanted, source)
				}
			}
		})
	}
}

// TestGenerateExecutableCardSourceRejectsComposedGroupRule confirms that a
// composed power/toughness and rule declaration on a battlefield group (rather
// than the source or its attached object) fails closed, since group-affecting
// rule operations need runtime group plumbing that does not yet exist.
func TestGenerateExecutableCardSourceRejectsComposedGroupRule(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Group Rule Lord",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		OracleText: "Other creatures you control get +1/+1 and can't block.",
		Power:      new("2"),
		Toughness:  new("2"),
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "g")
	if err != nil {
		t.Fatal(err)
	}
	if source != "" || len(diagnostics) == 0 {
		t.Fatalf("source = %q, diagnostics = %#v, want fail-closed", source, diagnostics)
	}
}

func TestGenerateExecutableCardSourceAllLandsTypeAddition(t *testing.T) {
	t.Parallel()
	for name, tc := range map[string]struct {
		typeLine   string
		oracleText string
		subtype    string
	}{
		"yavimaya": {
			typeLine:   "Legendary Land",
			oracleText: "Each land is a Forest in addition to its other land types.",
			subtype:    "Forest",
		},
		"urborg": {
			typeLine:   "Legendary Land",
			oracleText: "Each land is a Swamp in addition to its other land types.",
			subtype:    "Swamp",
		},
		"blanket of night": {
			typeLine:   "Enchantment",
			oracleText: "Each land is a Swamp in addition to its other land types.",
			subtype:    "Swamp",
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{
				Name:       "Land Type Static",
				Layout:     "normal",
				TypeLine:   tc.typeLine,
				OracleText: tc.oracleText,
			}
			source, diagnostics, err := GenerateExecutableCardSource(card, "l")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			for _, wanted := range []string{
				"Layer:       game.LayerType,",
				"game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Land}})",
				"AddSubtypes: []types.Sub{types.Sub(\"" + tc.subtype + "\")},",
			} {
				if !strings.Contains(source, wanted) {
					t.Fatalf("source missing %q:\n%s", wanted, source)
				}
			}
		})
	}
}

func TestGenerateExecutableCardSourceAllLandsTypeAdditionFailsClosed(t *testing.T) {
	t.Parallel()
	for name, oracleText := range map[string]string{
		"missing in addition tail": "Each land is a Forest.",
		"non-basic subtype":        "Each land is a Goblin in addition to its other land types.",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{
				Name:       "Bad Land Static",
				Layout:     "normal",
				TypeLine:   "Legendary Land",
				OracleText: oracleText,
			}
			source, diagnostics, err := GenerateExecutableCardSource(card, "l")
			if err != nil {
				t.Fatal(err)
			}
			if source != "" {
				t.Fatalf("source = %q, want no partial card", source)
			}
			if len(diagnostics) == 0 {
				t.Fatal("expected unsupported diagnostic")
			}
		})
	}
}
