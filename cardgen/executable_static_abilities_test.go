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
