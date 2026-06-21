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
		"CardType:    opt.Val(types.Land),",
		"Supertype:   opt.Val(types.Basic),",
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
		"SubtypesAny: []types.Sub{types.Swamp}",
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
		"PermanentTypes: []types.Card{types.Creature}",
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
		"PermanentTypes: []types.Card{types.Artifact, types.Creature}",
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
		`Subtypes: []types.Sub{types.Sub("Equipment")}`,
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
		"Enchant creature you control",             // controller qualifier
		"Enchant artifact or creature you control", // controller-qualified union
		"Enchant creature or",                      // dangling separator
		"Enchant artifact creature",                // conjunctive type line, no separator
		"Enchant nonland permanent",                // negated type qualifier
		"Enchant creatures",                        // plural form
		"Enchant instant",                          // non-permanent card type
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
