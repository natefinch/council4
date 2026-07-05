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

// TestGenerateExecutableCardSourceEveryCreatureTypeDropsScopeRider confirms a
// "Creatures you control are every creature type. The same is true for creature
// spells you control and creature cards you own that aren't on the battlefield."
// anthem (Maskwood Nexus) generates: the non-battlefield-zone scope rider is
// dropped and the battlefield every-creature-type continuous effect lowers.
func TestGenerateExecutableCardSourceEveryCreatureTypeDropsScopeRider(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:     "Type Nexus",
		Layout:   "normal",
		ManaCost: "{4}",
		TypeLine: "Artifact",
		OracleText: "Creatures you control are every creature type. " +
			"The same is true for creature spells you control and creature cards you own that aren't on the battlefield.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "n")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if !strings.Contains(source, "AddEveryCreatureType: true") {
		t.Fatalf("source missing every-creature-type continuous effect:\n%s", source)
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

func TestGenerateExecutableCardSourceSelfNameMustAttack(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Toski, Bearer of Secrets",
		Layout:     "normal",
		ManaCost:   "{3}{G}",
		TypeLine:   "Legendary Creature — Squirrel",
		OracleText: "This spell can't be countered.\nIndestructible\nToski attacks each combat if able.\nWhenever a creature you control deals combat damage to a player, draw a card.",
		Colors:     []string{"G"},
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
	if !strings.Contains(source, "game.MustAttackStaticBody") {
		t.Fatalf("source missing must-attack static body for self-name rule:\n%s", source)
	}
	if !strings.Contains(source, "game.CantBeCounteredStaticBody") {
		t.Fatalf("source missing uncounterable static body:\n%s", source)
	}
}

func TestGenerateExecutableCardSourceRejectsSelfNameUncounterable(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Certain Doom",
		Layout:     "normal",
		ManaCost:   "{1}{B}",
		TypeLine:   "Sorcery",
		OracleText: "Certain Doom can't be countered.\nDestroy target creature.",
		Colors:     []string{"B"},
	}
	source, _, err := GenerateExecutableCardSource(card, "c")
	if err != nil {
		t.Fatal(err)
	}
	if source != "" {
		t.Fatalf("expected name-based can't-be-countered to remain unsupported, got source:\n%s", source)
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

func TestGenerateExecutableCardSourceConditionalCannotAttack(t *testing.T) {
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
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if !strings.Contains(source, "Kind:                            game.RuleEffectCantAttack") {
		t.Fatalf("source missing can't-attack rule effect:\n%s", source)
	}
	if !strings.Contains(source, `AttackDefenderControlsSelection: game.Selection{SubtypesAny: []types.Sub{types.Sub("Island")}}`) {
		t.Fatalf("source missing defender-controls selection:\n%s", source)
	}
}

// TestGenerateExecutableCardSourceConditionalCannotAttackMonarch verifies the
// "can't attack unless defending player is the monarch" designation guard
// (Crown-Hunter Hireling) lowers onto the AttackDefenderIsMonarch rule-effect
// flag, the designation counterpart of the defender-controls selection guard.
func TestGenerateExecutableCardSourceConditionalCannotAttackMonarch(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Crown Hunter Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		OracleText: "This creature can't attack unless defending player is the monarch.",
		Power:      new("3"),
		Toughness:  new("2"),
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "c")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if !strings.Contains(source, "Kind:                    game.RuleEffectCantAttack") {
		t.Fatalf("source missing can't-attack rule effect:\n%s", source)
	}
	if !strings.Contains(source, "AttackDefenderIsMonarch: true,") {
		t.Fatalf("source missing defender-is-monarch flag:\n%s", source)
	}
}

func TestGenerateExecutableCardSourceRejectsConditionalCannotAttackUnsupportedSelection(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Conditional Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		OracleText: "This creature can't attack unless defending player controls an enchantment or an enchanted permanent.",
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

func TestGenerateExecutableCardSourceGroupUncounterable(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		oracleText string
		wantTypes  string
	}{
		{
			name:       "creature spells",
			oracleText: "Creature spells you control can't be countered.",
			wantTypes:  "SpellTypes:         []types.Card{types.Creature},",
		},
		{
			name:       "all spells",
			oracleText: "Spells you control can't be countered.",
			wantTypes:  "",
		},
		{
			name:       "instant spells",
			oracleText: "Instant spells you control can't be countered.",
			wantTypes:  "SpellTypes:         []types.Card{types.Instant},",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{
				Name:       "Test Rhythm",
				Layout:     "normal",
				TypeLine:   "Enchantment",
				OracleText: test.oracleText,
			}
			source, diagnostics, err := GenerateExecutableCardSource(card, "c")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			if !strings.Contains(source, "Kind:               game.RuleEffectCantBeCountered,") ||
				!strings.Contains(source, "AffectedController: game.ControllerYou,") {
				t.Fatalf("source missing group uncounterable rule effect:\n%s", source)
			}
			if test.wantTypes == "" {
				if strings.Contains(source, "SpellTypes:") {
					t.Fatalf("unfiltered group should not constrain spell types:\n%s", source)
				}
			} else if !strings.Contains(source, test.wantTypes) {
				t.Fatalf("source missing %q:\n%s", test.wantTypes, source)
			}
		})
	}
}

func TestGenerateExecutableCardSourceUntapDuringOtherUntapStep(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		typeLine   string
		oracleText string
		wantFields []string
		wantAbsent []string
	}{
		{
			name:       "all permanents",
			typeLine:   "Creature — Spirit",
			oracleText: "Untap all permanents you control during each other player's untap step.",
			wantFields: []string{
				"Kind:               game.RuleEffectUntapDuringOtherPlayersUntapStep,",
				"AffectedController: game.ControllerYou,",
			},
			wantAbsent: []string{"PermanentTypes:", "AffectedSource:"},
		},
		{
			name:       "all creatures",
			typeLine:   "Creature — Elemental",
			oracleText: "Untap all creatures you control during each other player's untap step.",
			wantFields: []string{
				"Kind:               game.RuleEffectUntapDuringOtherPlayersUntapStep,",
				"AffectedController: game.ControllerYou,",
				"PermanentTypes:     []types.Card{types.Creature},",
			},
		},
		{
			name:       "self form",
			typeLine:   "Artifact",
			oracleText: "Untap this artifact during each other player's untap step.",
			wantFields: []string{
				"Kind:           game.RuleEffectUntapDuringOtherPlayersUntapStep,",
				"AffectedSource: true,",
			},
			wantAbsent: []string{"AffectedController:", "PermanentTypes:"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{
				Name:       "Test Untapper",
				Layout:     "normal",
				TypeLine:   test.typeLine,
				OracleText: test.oracleText,
			}
			source, diagnostics, err := GenerateExecutableCardSource(card, "c")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			for _, want := range test.wantFields {
				if !strings.Contains(source, want) {
					t.Fatalf("source missing %q:\n%s", want, source)
				}
			}
			for _, absent := range test.wantAbsent {
				if strings.Contains(source, absent) {
					t.Fatalf("source unexpectedly contains %q:\n%s", absent, source)
				}
			}
		})
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

func TestGenerateExecutableCardSourceSelfCantBlockAndCantBeBlocked(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Changeling Outcast",
		Layout:     "normal",
		ManaCost:   "{B}",
		TypeLine:   "Creature — Shapeshifter",
		OracleText: "Changeling (This card is every creature type.)\nChangeling Outcast can't block and can't be blocked.",
		Colors:     []string{"B"},
		Power:      new("1"),
		Toughness:  new("1"),
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "c")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Kind:           game.RuleEffectCantBlock,",
		"Kind:           game.RuleEffectCantBeBlocked,",
		"game.ChangelingStaticBody",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
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

// TestGenerateExecutableCardSourceAttachedDoesntUntapUnlessMonarch verifies the
// "doesn't untap during its controller's untap step unless that player is the
// monarch" designation guard (Fall from Favor) lowers onto the
// UntapUnlessControllerIsMonarch rule-effect flag, the untap-step counterpart of
// the can't-attack defender-is-monarch guard.
func TestGenerateExecutableCardSourceAttachedDoesntUntapUnlessMonarch(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:     "Frozen Crown",
		Layout:   "normal",
		ManaCost: "{2}{U}",
		TypeLine: "Enchantment — Aura",
		OracleText: "Enchant creature\n" +
			"Enchanted creature doesn't untap during its controller's untap step unless that player is the monarch.",
		Colors: []string{"U"},
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "f")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Kind:                           game.RuleEffectDoesntUntap,",
		"AffectedAttached:               true,",
		"UntapUnlessControllerIsMonarch: true,",
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

func TestGenerateExecutableCardSourceGroupEveryCreatureType(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Maskwood Nexus",
		Layout:     "normal",
		ManaCost:   "{4}",
		TypeLine:   "Artifact",
		OracleText: "Creatures you control are every creature type.\n{3}, {T}: Create a 2/2 black Shapeshifter creature token with changeling.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "m")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Layer:                game.LayerType,",
		"AddEveryCreatureType: true,",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceSelfEveryCreatureType(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Mistform Ultimus",
		Layout:     "normal",
		ManaCost:   "{4}{U}",
		TypeLine:   "Creature — Illusion",
		OracleText: "Mistform Ultimus is every creature type.",
		Power:      new("3"),
		Toughness:  new("3"),
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "u")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if !strings.Contains(source, "AddEveryCreatureType: true,") {
		t.Fatalf("source missing every-creature-type effect:\n%s", source)
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
		"prohibition first then keyword grant": {
			typeLine:   "Artifact — Equipment",
			oracleText: "Equipped creature can't be blocked and has shroud.\nEquip {2}",
			wanted: []string{
				"game.RuleEffectCantBeBlocked",
				"AffectedAttached: true",
				"game.Shroud",
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

func TestGenerateExecutableCardSourceEveryBasicLandType(t *testing.T) {
	t.Parallel()
	for name, tc := range map[string]struct {
		typeLine   string
		oracleText string
		group      string
	}{
		"dryad of the ilysian grove": {
			typeLine:   "Enchantment Creature — Dryad",
			oracleText: "You may play an additional land on each of your turns.\nLands you control are every basic land type in addition to their other types.",
			group:      "game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Land}})",
		},
		"prismatic omen": {
			typeLine:   "Enchantment",
			oracleText: "Lands you control are every basic land type in addition to their other types.",
			group:      "game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Land}})",
		},
		"all lands": {
			typeLine:   "Legendary Land",
			oracleText: "Each land is every basic land type in addition to its other land types.",
			group:      "game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Land}})",
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{
				Name:       "Every Basic Land Static",
				Layout:     "normal",
				TypeLine:   tc.typeLine,
				OracleText: tc.oracleText,
			}
			source, diagnostics, err := GenerateExecutableCardSource(card, "e")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			for _, wanted := range []string{
				"Layer:                 game.LayerType,",
				tc.group,
				"AddEveryBasicLandType: true,",
			} {
				if !strings.Contains(source, wanted) {
					t.Fatalf("source missing %q:\n%s", wanted, source)
				}
			}
		})
	}
}

func TestGenerateExecutableCardSourceNonlandPermanentsAreArtifacts(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Encroaching Mycosynth",
		Layout:     "normal",
		TypeLine:   "Artifact",
		OracleText: "Nonland permanents you control are artifacts in addition to their other types. The same is true for permanent spells you control and nonland permanent cards you own that aren't on the battlefield.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "e")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Layer:    game.LayerType,",
		"game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{ExcludedTypes: []types.Card{types.Land}})",
		"AddTypes: []types.Card{types.Artifact},",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
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

func TestGenerateExecutableCardSourceCastColorlessSpellsFromLibraryTop(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Forge",
		Layout:     "normal",
		ManaCost:   "{4}",
		TypeLine:   "Artifact",
		OracleText: "You may cast artifact spells and colorless spells from the top of your library.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "f")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"[]types.Card{types.Artifact},",
		"SpellColorless: true,",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
	if strings.Contains(source, "TODO") {
		t.Fatalf("executable source contains TODO:\n%s", source)
	}
}

func TestGenerateExecutableCardSourceCastChosenTypeSpellsFromLibraryTop(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:     "Test Realmwalker",
		Layout:   "normal",
		ManaCost: "{2}{G}",
		TypeLine: "Creature — Shapeshifter",
		OracleText: "Changeling (This card is every creature type.)\n" +
			"As this creature enters, choose a creature type.\n" +
			"You may look at the top card of your library any time.\n" +
			"You may cast creature spells of the chosen type from the top of your library.",
		Power:     new("2"),
		Toughness: new("3"),
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "r")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"[]types.Card{types.Creature},",
		"SpellChosenSubtypeFrom: game.EntryTypeChoiceKey,",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
	if strings.Contains(source, "TODO") {
		t.Fatalf("executable source contains TODO:\n%s", source)
	}
}

func TestGenerateExecutableCardSourcePlayFromLibraryTop(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Seer",
		Layout:     "normal",
		ManaCost:   "{2}{G}",
		TypeLine:   "Creature — Elf",
		OracleText: "Play with the top card of your library revealed.\nYou may play lands from the top of your library.",
		Power:      new("2"),
		Toughness:  new("2"),
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "s")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"game.PlayWithTopCardRevealedStaticBody",
		"game.PlayLandsFromLibraryTopStaticBody",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
	if strings.Contains(source, "TODO") {
		t.Fatalf("executable source contains TODO:\n%s", source)
	}
}

func TestGenerateExecutableCardSourceLookAtTopCardAnyTime(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Seer",
		Layout:     "normal",
		ManaCost:   "{1}{U}",
		TypeLine:   "Creature — Sphinx",
		OracleText: "Flying\nYou may look at the top card of your library any time.",
		Power:      new("4"),
		Toughness:  new("4"),
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "s")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if !strings.Contains(source, "game.LookAtTopCardAnyTimeStaticBody") {
		t.Fatalf("source missing look-at-top static body:\n%s", source)
	}
	if strings.Contains(source, "TODO") {
		t.Fatalf("executable source contains TODO:\n%s", source)
	}
}

func TestGenerateExecutableCardSourceConditionalAttachedKeywordGrant(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Champion's Helm",
		Layout:     "normal",
		ManaCost:   "{2}",
		TypeLine:   "Artifact — Equipment",
		OracleText: "Equipped creature gets +2/+2.\nAs long as equipped creature is legendary, it has hexproof.\nEquip {1}",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "c")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Condition: opt.Val(game.Condition{",
		"Object:        opt.Val(game.SourceAttachedPermanentReference()),",
		"ObjectMatches: opt.Val(game.Selection{Supertypes: []types.Super{types.Legendary}}),",
		"Group: game.AttachedObjectGroup(game.SourcePermanentReference()),",
		"game.Hexproof,",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
	if strings.Contains(source, "TODO") {
		t.Fatalf("executable source contains TODO:\n%s", source)
	}
}

func TestGenerateExecutableCardSourceConditionalEnchantedKeywordGrant(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Conditional Ward Aura",
		Layout:     "normal",
		ManaCost:   "{W}",
		TypeLine:   "Enchantment — Aura",
		OracleText: "Enchant creature\nAs long as enchanted creature is legendary, it has hexproof.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "e")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Object:        opt.Val(game.SourceAttachedPermanentReference()),",
		"ObjectMatches: opt.Val(game.Selection{Supertypes: []types.Super{types.Legendary}}),",
		"Group: game.AttachedObjectGroup(game.SourcePermanentReference()),",
		"game.Hexproof,",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceChosenTypeGroupAnthem(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Type Shaper",
		Layout:     "normal",
		ManaCost:   "{4}",
		TypeLine:   "Enchantment",
		OracleText: "As this enchantment enters, choose a creature type.\nCreatures you control are the chosen type in addition to their other types.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "c")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Layer:                     game.LayerType,",
		"Group:                     game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}}),",
		"AddSubtypeFromEntryChoice: game.EntryTypeChoiceKey,",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}
