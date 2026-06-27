package cardgen

import (
	"strings"
	"testing"
)

func TestGenerateExecutableCardSourceEternalize(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Eternalizer",
		Layout:     "normal",
		TypeLine:   "Creature — Snake Druid",
		ManaCost:   "{1}{G}",
		Power:      new("1"),
		Toughness:  new("4"),
		OracleText: "Eternalize {2}{G}{G} ({2}{G}{G}, Exile this card from your graveyard: Create a token that's a copy of it, except it's a 4/4 black Zombie Snake Druid with no mana cost. Eternalize only as a sorcery.)",
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
		`game.EternalizeActivatedBody(cost.Mana{cost.O(2), cost.G, cost.G}, types.Sub("Snake"), types.Sub("Druid"))`,
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceEmbalm(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Embalmer",
		Layout:     "normal",
		TypeLine:   "Creature — Human Cleric",
		ManaCost:   "{2}{W}",
		Power:      new("1"),
		Toughness:  new("4"),
		OracleText: "Embalm {3}{W} ({3}{W}, Exile this card from your graveyard: Create a token that's a copy of it, except it's a white Zombie Human Cleric with no mana cost. Embalm only as a sorcery.)",
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
		`game.EmbalmActivatedBody(cost.Mana{cost.O(3), cost.W}, types.Sub("Human"), types.Sub("Cleric"))`,
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

func TestGenerateExecutableCardSourceScavenge(t *testing.T) {
	t.Parallel()
	power := "4"
	toughness := "2"
	card := &ScryfallCard{
		Name:       "Deadbridge Goliath",
		Layout:     "normal",
		ManaCost:   "{2}{G}{G}",
		TypeLine:   "Creature — Insect",
		OracleText: "Scavenge {4}{G}{G} ({4}{G}{G}, Exile this card from your graveyard: Put a number of +1/+1 counters equal to this card's power on target creature. Scavenge only as a sorcery.)",
		Power:      &power,
		Toughness:  &toughness,
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
		"game.ScavengeActivatedAbility(cost.Mana",
		"cost.O(4)",
		"cost.G",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceDredge(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Dredger",
		Layout:     "normal",
		TypeLine:   "Land",
		OracleText: "Dredge 2 (If you would draw a card, you may mill two cards instead. If you do, return this card from your graveyard to your hand.)",
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
		"game.DredgeStaticAbility(2)",
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

func TestGenerateExecutableCardSourceBasicLandcycling(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Ash Barrens",
		Layout:     "normal",
		TypeLine:   "Land",
		OracleText: "{T}: Add {C}.\nBasic landcycling {1} ({1}, Discard this card: Search your library for a basic land card, reveal it, put it into your hand, then shuffle.)",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "a")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"ActivatedAbilities: []game.ActivatedAbility",
		"ManaAbilities: []game.ManaAbility",
		"game.TapManaAbility(mana.C)",
		"ZoneOfFunction: zone.Hand,",
		"Kind:   cost.AdditionalDiscard,",
		"game.CyclingKeyword{Cost: cost.Mana{cost.O(1)}}",
		"Primitive: game.Search{",
		"Filter:      game.Selection{RequiredTypes: []types.Card{types.Land}, Supertypes: []types.Super{types.Basic}},",
		"Reveal:      true,",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceTypedLandcycling(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Snatcher",
		Layout:     "normal",
		TypeLine:   "Creature — Spirit",
		ManaCost:   "{4}{B}",
		Colors:     []string{"B"},
		OracleText: "Swampcycling {2} ({2}, Discard this card: Search your library for a Swamp card, reveal it, put it into your hand, then shuffle.)",
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
		"Primitive: game.Search{",
		`SubtypesAny: []types.Sub{types.Sub("Swamp")}`,
		"Destination: zone.Hand,",
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

func TestGenerateExecutableCardSourceRestrictedEquip(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		oracle string
		wanted []string
		unwant []string
	}{
		{
			name:   "legendary supertype",
			oracle: "Equipped creature gets +1/+1 for each land you control.\nEquip legendary creature {3}\nEquip {7}",
			wanted: []string{"game.EquipRestrictedActivatedAbility(cost.Mana{cost.O(3)}, []types.Super{types.Legendary}, nil)"},
		},
		{
			name:   "single subtype",
			oracle: "Equipped creature gets +2/+0.\nEquip Knight {1}\nEquip {3}",
			wanted: []string{"game.EquipRestrictedActivatedAbility(cost.Mana{cost.O(1)}, nil, []types.Sub{types.Knight})"},
		},
		{
			name:   "subtype list",
			oracle: "Equip Shaman, Warlock, or Wizard {2}\nEquip {6}",
			wanted: []string{"game.EquipRestrictedActivatedAbility(cost.Mana{cost.O(2)}, nil, []types.Sub{types.Shaman, types.Warlock, types.Wizard})"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{
				Name:       "Test Equipment",
				Layout:     "normal",
				TypeLine:   "Artifact — Equipment",
				OracleText: tc.oracle,
			}
			source, diagnostics, err := GenerateExecutableCardSource(card, "t")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			for _, wanted := range tc.wanted {
				if !strings.Contains(source, wanted) {
					t.Fatalf("source missing %q:\n%s", wanted, source)
				}
			}
		})
	}
}

func TestGenerateExecutableCardSourceUnsupportedEquipRestriction(t *testing.T) {
	t.Parallel()
	for _, oracle := range []string{
		"Equip commander {3}\nEquip {5}",
		"Equip planeswalker {5}\nEquip {5}",
	} {
		card := &ScryfallCard{
			Name:       "Test Equipment",
			Layout:     "normal",
			TypeLine:   "Artifact — Equipment",
			OracleText: oracle,
		}
		_, diagnostics, err := GenerateExecutableCardSource(card, "t")
		if err != nil {
			t.Fatal(err)
		}
		if len(diagnostics) == 0 {
			t.Fatalf("expected diagnostics for %q", oracle)
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
		"RequiredTypesAny: []types.Card{types.Creature}",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceEnchantTypeUnion(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Aura",
		Layout:     "normal",
		TypeLine:   "Enchantment — Aura",
		OracleText: "Enchant artifact or creature",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"game.EnchantStaticAbility(&game.TargetSpec{",
		`Constraint: "artifact or creature"`,
		"RequiredTypesAny: []types.Card{types.Artifact, types.Creature}",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceEnchantSubtype(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Aura",
		Layout:     "normal",
		TypeLine:   "Enchantment — Aura",
		OracleText: "Enchant Equipment",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"game.EnchantStaticAbility(&game.TargetSpec{",
		`Constraint: "equipment"`,
		`SubtypesAny: []types.Sub{types.Sub("Equipment")}`,
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceEnchantYouControl(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Aura",
		Layout:     "normal",
		TypeLine:   "Enchantment — Aura",
		OracleText: "Enchant creature you control",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"game.EnchantStaticAbility(&game.TargetSpec{",
		`Constraint: "creature you control"`,
		"RequiredTypesAny: []types.Card{types.Creature}",
		"Controller: game.ControllerYou",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceEnchantSubtypeYouControl(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Aura",
		Layout:     "normal",
		TypeLine:   "Enchantment — Aura",
		OracleText: "Enchant Mountain you control",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		`Constraint: "mountain you control"`,
		`SubtypesAny: []types.Sub{types.Sub("Mountain")}`,
		"Controller: game.ControllerYou",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceEnchantOpponent(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Aura",
		Layout:     "normal",
		TypeLine:   "Enchantment — Aura",
		OracleText: "Enchant opponent",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"game.EnchantStaticAbility(&game.TargetSpec{",
		`Constraint: "opponent"`,
		"Allow:      game.TargetAllowPlayer",
		"Player: game.PlayerOpponent",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceEnchantMixedUnion(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Aura",
		Layout:     "normal",
		TypeLine:   "Enchantment — Aura",
		OracleText: "Enchant creature or Vehicle",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"game.EnchantStaticAbility(&game.TargetSpec{",
		"Selection:  opt.Val(game.Selection{AnyOf: []game.Selection{",
		"RequiredTypesAny: []types.Card{types.Creature}",
		`SubtypesAny: []types.Sub{types.Sub("Vehicle")}`,
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
	// A non-empty Constraint would let the runtime permanent-type matcher
	// re-parse "creature or vehicle" and reject a non-creature Vehicle, so the
	// mixed-union spec must rely solely on the Selection.
	if strings.Contains(source, `Constraint: "creature or vehicle"`) {
		t.Fatalf("mixed-union spec must not set a Constraint:\n%s", source)
	}
}

func TestGenerateExecutableCardSourceEnchantUnsupportedTargets(t *testing.T) {
	t.Parallel()
	for _, oracle := range []string{
		"Enchant creature or",       // dangling separator
		"Enchant artifact creature", // conjunctive type line, no separator
		"Enchant nonland permanent", // negated type qualifier
		"Enchant creatures",         // plural form
		"Enchant instant",           // non-permanent card type
	} {
		card := &ScryfallCard{
			Name:       "Test Aura",
			Layout:     "normal",
			TypeLine:   "Enchantment — Aura",
			OracleText: oracle,
		}
		_, diagnostics, err := GenerateExecutableCardSource(card, "t")
		if err != nil {
			t.Fatalf("%q: %v", oracle, err)
		}
		if len(diagnostics) == 0 {
			t.Errorf("%q: expected an unsupported diagnostic, got none", oracle)
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

func TestGenerateExecutableCardSourceChosenColorProtection(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Shield",
		Layout:     "normal",
		TypeLine:   "Enchantment — Aura",
		OracleText: "Enchant creature\nAs this Aura enters, choose a color.\nEnchanted creature has protection from the chosen color.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"new(game.ProtectionFromChosenColorStaticAbility())",
		"game.EntryColorChoiceReplacement(",
		"Group: game.AttachedObjectGroup(game.SourcePermanentReference())",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceChosenColorProtectionSelf(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Voice",
		Layout:     "normal",
		TypeLine:   "Creature — Human Cleric",
		OracleText: "As Test Voice enters, choose a color.\nTest Voice has protection from the chosen color.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"new(game.ProtectionFromChosenColorStaticAbility())",
		"game.EntryColorChoiceReplacement(",
		"AffectedSource: true",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceChosenColorProtectionGroup(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Ward",
		Layout:     "normal",
		TypeLine:   "Creature — Sliver",
		OracleText: "As Test Ward enters, choose a color.\nAll Slivers have protection from the chosen color.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"new(game.ProtectionFromChosenColorStaticAbility())",
		"game.EntryColorChoiceReplacement(",
		"game.BattlefieldGroup(",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceDelveSpellWithEffect(t *testing.T) {
	t.Parallel()
	// A spell keyword (Delve) on its own paragraph is a static keyword ability
	// even on a sorcery; the resolving effect lowers as the spell ability. The
	// keyword-only paragraph must not be rejected as unsupported spell content.
	card := &ScryfallCard{
		Name:     "Test Cruise",
		Layout:   "normal",
		TypeLine: "Sorcery",
		OracleText: "Delve (Each card you exile from your graveyard while casting this spell pays for {1}.)\n" +
			"Draw three cards.",
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
		"game.DelveStaticBody",
		"SpellAbility: opt.Val(game.Mode{",
		"game.Draw{",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceLandwalk(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		oracleText string
		wantVar    string
	}{
		"forestwalk": {
			oracleText: "Forestwalk (This creature can't be blocked as long as defending player controls a Forest.)",
			wantVar:    "game.ForestwalkStaticBody",
		},
		"islandwalk": {
			oracleText: "Islandwalk (This creature can't be blocked as long as defending player controls an Island.)",
			wantVar:    "game.IslandwalkStaticBody",
		},
		"swampwalk": {
			oracleText: "Swampwalk (This creature can't be blocked as long as defending player controls a Swamp.)",
			wantVar:    "game.SwampwalkStaticBody",
		},
		"mountainwalk": {
			oracleText: "Mountainwalk (This creature can't be blocked as long as defending player controls a Mountain.)",
			wantVar:    "game.MountainwalkStaticBody",
		},
		"plainswalk": {
			oracleText: "Plainswalk (This creature can't be blocked as long as defending player controls a Plains.)",
			wantVar:    "game.PlainswalkStaticBody",
		},
		"desertwalk": {
			oracleText: "Desertwalk (This creature can't be blocked as long as defending player controls a Desert.)",
			wantVar:    "game.DesertwalkStaticBody",
		},
		"generic landwalk": {
			oracleText: "Landwalk (This creature can't be blocked as long as defending player controls a land.)",
			wantVar:    "game.LandwalkStaticBody",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{
				Name:       "Test Walker",
				Layout:     "normal",
				TypeLine:   "Creature — Spirit",
				ManaCost:   "{1}{G}",
				Power:      new("1"),
				Toughness:  new("1"),
				OracleText: tc.oracleText,
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
				tc.wantVar,
			} {
				if !strings.Contains(source, wanted) {
					t.Fatalf("source missing %q:\n%s", wanted, source)
				}
			}
		})
	}
}

// TestGenerateExecutableCardSourceQualifiedLandwalkUnsupported confirms that
// qualified landwalk forms the backend does not model fail closed instead of
// being silently lowered as plain landwalk.
func TestGenerateExecutableCardSourceQualifiedLandwalkUnsupported(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Snow Walker",
		Layout:     "normal",
		TypeLine:   "Creature — Spirit",
		ManaCost:   "{1}{U}",
		Power:      new("1"),
		Toughness:  new("1"),
		OracleText: "Snow swampwalk (This creature can't be blocked as long as defending player controls a snow Swamp.)",
	}
	_, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) == 0 {
		t.Fatal("expected diagnostics for unsupported snow swampwalk, got none")
	}
}

func TestGenerateExecutableCardSourceTraining(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Trainee",
		Layout:     "normal",
		TypeLine:   "Creature — Human Soldier",
		ManaCost:   "{W}",
		Power:      new("1"),
		Toughness:  new("1"),
		OracleText: "Training (Whenever this creature attacks with another creature with greater power, put a +1/+1 counter on this creature.)",
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
		"game.TrainingTriggeredBody",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceSaddle(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Mount",
		Layout:     "normal",
		TypeLine:   "Creature — Horse Mount",
		ManaCost:   "{1}{R}",
		Power:      new("2"),
		Toughness:  new("2"),
		OracleText: "Whenever this creature attacks while saddled, it gets +1/+0 until end of turn.\nSaddle 2 (Tap any number of other creatures you control with total power 2 or more: This Mount becomes saddled until end of turn. Saddle only as a sorcery.)",
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
		"game.SaddleActivatedAbility(2)",
		"AttackWhileSaddled: true,",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}
