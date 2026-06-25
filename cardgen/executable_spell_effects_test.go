package cardgen

import (
	"strings"
	"testing"
)

func TestGenerateExecutableHeroicIntervention(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Heroic Intervention",
		Layout:     "normal",
		TypeLine:   "Instant",
		ManaCost:   "{1}{G}",
		Colors:     []string{"G"},
		OracleText: "Permanents you control gain hexproof and indestructible until end of turn.",
	}, "h")
	if err != nil {
		t.Fatal(err)
	}

	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"game.ApplyContinuous",
		"game.BattlefieldGroup(game.Selection{Controller: game.ControllerYou})",
		"AddKeywords: []game.Keyword{",
		"game.Hexproof",
		"game.Indestructible",
		"Duration: game.DurationUntilEndOfTurn",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated Heroic Intervention missing %q:\n%s", want, source)
		}
	}
}

func TestGenerateExecutableSwanSong(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Swan Song",
		Layout:     "normal",
		TypeLine:   "Instant",
		ManaCost:   "{U}",
		Colors:     []string{"U"},
		OracleText: "Counter target enchantment, instant, or sorcery spell. Its controller creates a 2/2 blue Bird creature token with flying.",
	}, "s")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"SpellCardTypesAny: []types.Card{types.Enchantment, types.Instant, types.Sorcery}",
		"Primitive: game.CounterObject{",
		"Primitive: game.CreateToken{",
		"Recipient: opt.Val(game.ObjectControllerReference(game.TargetStackObjectReference(0)))",
		"Name:      \"Bird\"",
		"Colors:    []color.Color{color.Blue}",
		"Subtypes:  []types.Sub{types.Bird}",
		"Power:     opt.Val(game.PT{Value: 2})",
		"Toughness: opt.Val(game.PT{Value: 2})",
		"game.FlyingStaticBody",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated Swan Song missing %q:\n%s", want, source)
		}
	}
}

func TestLowerMassBounceSpellToGroup(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		oracleText string
		wantGroup  string
	}{
		{
			name:       "all creatures",
			oracleText: "Return all creatures to their owners' hands.",
			wantGroup:  "Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}}),",
		},
		{
			name:       "all permanents",
			oracleText: "Return all permanents to their owners' hands.",
			wantGroup:  "Group: game.BattlefieldGroup(game.Selection{}),",
		},
		{
			name:       "all artifacts you control",
			oracleText: "Return all artifacts you control to their owner's hand.",
			wantGroup:  "Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Artifact}, Controller: game.ControllerYou}),",
		},
		{
			name:       "all attacking creatures",
			oracleText: "Return all attacking creatures to their owner's hand.",
			wantGroup:  "Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, CombatState: game.CombatStateAttacking}),",
		},
		{
			name:       "all blocking creatures",
			oracleText: "Return all blocking creatures to their owners' hands.",
			wantGroup:  "Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, CombatState: game.CombatStateBlocking}),",
		},
		{
			name:       "each creature without a +1/+1 counter",
			oracleText: "Return each creature without a +1/+1 counter on it to its owner's hand.",
			wantGroup:  "Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, MatchExcludedCounter: true, ExcludedCounter: counter.PlusOnePlusOne}),",
		},
		{
			name:       "each permanent",
			oracleText: "Return each permanent to its owner's hand.",
			wantGroup:  "Group: game.BattlefieldGroup(game.Selection{}),",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
				Name:       "Test Wipe",
				Layout:     "normal",
				TypeLine:   "Sorcery",
				ManaCost:   "{3}{U}",
				OracleText: test.oracleText,
			}, "t")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) != 0 {
				t.Fatalf("mass bounce %q unexpectedly failed: %v", test.oracleText, diagnostics)
			}
			if !strings.Contains(source, "Primitive: game.Bounce{") || !strings.Contains(source, test.wantGroup) {
				t.Fatalf("mass bounce %q did not lower to the expected group Bounce:\n%s", test.oracleText, source)
			}
		})
	}
}

func TestLowerControlledChoiceBounceSpell(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		oracleText    string
		wantSelection string
	}{
		{
			name:          "any permanent you control",
			oracleText:    "Return a permanent you control to its owner's hand.",
			wantSelection: "Group:            game.BattlefieldGroup(game.Selection{Controller: game.ControllerYou}),",
		},
		{
			name:          "creature you control",
			oracleText:    "Return a creature you control to its owner's hand.",
			wantSelection: "Group:            game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),",
		},
		{
			name:          "another permanent you control excludes source",
			oracleText:    "Return another creature you control to its owner's hand.",
			wantSelection: "Group:            game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou, ExcludeSource: true}),",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
				Name:       "Test Bouncer",
				Layout:     "normal",
				TypeLine:   "Sorcery",
				ManaCost:   "{1}{U}",
				OracleText: test.oracleText,
			}, "t")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) != 0 {
				t.Fatalf("choice bounce %q unexpectedly failed: %v", test.oracleText, diagnostics)
			}
			if !strings.Contains(source, "Primitive: game.Bounce{") ||
				!strings.Contains(source, "ControlledChoice: true,") ||
				!strings.Contains(source, "Amount:           game.Fixed(1),") ||
				!strings.Contains(source, test.wantSelection) {
				t.Fatalf("choice bounce %q did not lower to the expected ControlledChoice Bounce:\n%s", test.oracleText, source)
			}
		})
	}
}

func TestLowerMassBounceFailsClosedForUnexpressibleGroups(t *testing.T) {
	t.Parallel()
	for _, oracleText := range []string{
		"Return all permanents of the color of your choice to their owners' hands.",
		"Return all creatures to their owners' hands except for Krakens, Leviathans, Octopuses, and Serpents.",
		"Return all but one creature to their owners' hands.",
	} {
		_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
			Name:       "Test Wipe",
			Layout:     "normal",
			TypeLine:   "Sorcery",
			ManaCost:   "{3}{U}",
			OracleText: oracleText,
		}, "t")
		if err != nil {
			t.Fatal(err)
		}
		if len(diagnostics) == 0 {
			t.Fatalf("mass bounce %q was expected to fail closed but lowered cleanly", oracleText)
		}
	}
}

func TestGenerateExecutableCardSourceSpellAdditionalExileXCost(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Pyre",
		Layout:     "normal",
		ManaCost:   "{X}{R}",
		TypeLine:   "Instant",
		OracleText: "As an additional cost to cast this spell, exile X cards from your graveyard.\nTest Pyre deals X damage to target creature.",
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
		"AdditionalCosts: []cost.Additional{",
		"Kind:        cost.AdditionalExile,",
		"AmountFromX: true,",
		"Source:      zone.Graveyard,",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceSpellAdditionalExileTypedCardCostOnPermanent(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Skaab",
		Layout:     "normal",
		ManaCost:   "{3}{U}",
		TypeLine:   "Creature — Zombie",
		OracleText: "As an additional cost to cast this spell, exile a creature card from your graveyard.\nFlying",
		Colors:     []string{"U"},
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
	for _, wanted := range []string{
		"AdditionalCosts: []cost.Additional{",
		"Kind:          cost.AdditionalExile,",
		"MatchCardType: true,",
		"CardType:      types.Creature,",
		"Source:        zone.Graveyard,",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceMassReturnFromGraveyardImportsZone(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Replenish",
		Layout:     "normal",
		ManaCost:   "{3}{W}",
		TypeLine:   "Sorcery",
		OracleText: "Return all enchantment cards from your graveyard to the battlefield.",
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
		"game.MassReturnFromGraveyard",
		"Destination: zone.Battlefield,",
		// The zone literal above requires the zone import; without it the
		// generated card fails to compile with "undefined: zone".
		`"github.com/natefinch/council4/mtg/game/zone"`,
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceVanillaPermanentWithAdditionalCost(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Mauler",
		Layout:     "normal",
		ManaCost:   "{3}{U}",
		TypeLine:   "Creature — Zombie",
		OracleText: "As an additional cost to cast this spell, exile a creature card from your graveyard.",
		Colors:     []string{"U"},
		Power:      new("6"),
		Toughness:  new("5"),
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if !strings.Contains(source, "AdditionalCosts: []cost.Additional{") {
		t.Fatalf("source missing additional cost on vanilla permanent:\n%s", source)
	}
}

func TestGenerateExecutableCardSourceSpellAdditionalSacrificeCost(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Splinters",
		Layout:     "normal",
		ManaCost:   "{B}",
		TypeLine:   "Sorcery",
		OracleText: "As an additional cost to cast this spell, sacrifice a creature.\nDestroy target creature.",
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
		"AdditionalCosts: []cost.Additional{",
		"Kind:               cost.AdditionalSacrifice,",
		"MatchPermanentType: true,",
		"PermanentType:      types.Creature,",
		"Primitive: game.Destroy{",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceSpellAdditionalTwoTypeSacrificeCost(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Dispute",
		Layout:     "normal",
		ManaCost:   "{1}{B}",
		TypeLine:   "Instant",
		OracleText: "As an additional cost to cast this spell, sacrifice an artifact or creature.\nDraw two cards.",
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
		"AdditionalCosts: []cost.Additional{",
		"Kind:               cost.AdditionalSacrifice,",
		"MatchPermanentType: true,",
		"PermanentType:      types.Artifact,",
		"PermanentTypeAlt:   types.Creature,",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestLowerSelfBounceReturn(t *testing.T) {
	t.Parallel()
	for _, oracleText := range []string{
		"{U}: Return this creature to its owner's hand.",
		"Sacrifice a land: Return this creature to its owner's hand.",
	} {
		source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
			Name:       "Test Merfolk",
			Layout:     "normal",
			TypeLine:   "Creature — Merfolk",
			Power:      new("2"),
			Toughness:  new("2"),
			OracleText: oracleText,
		}, "t")
		if err != nil {
			t.Fatal(err)
		}
		if len(diagnostics) != 0 {
			t.Fatalf("self-bounce %q unexpectedly failed: %v", oracleText, diagnostics)
		}
		if !strings.Contains(source, "game.Bounce") ||
			!strings.Contains(source, "game.SourcePermanentReference()") {
			t.Fatalf("self-bounce %q did not lower to a source Bounce:\n%s", oracleText, source)
		}
	}
}

func TestLowerSelfNameBounceReturn(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		typeLine   string
		oracleText string
	}{
		{
			name:       "Selenia, Dark Angel",
			typeLine:   "Creature — Angel",
			oracleText: "Flying\nPay 2 life: Return Selenia to its owner's hand.",
		},
		{
			name:       "Oboro, Palace in the Clouds",
			typeLine:   "Land",
			oracleText: "({T}: Add {U}.)\n{1}: Return Oboro to its owner's hand.",
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
			}, "t")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) != 0 {
				t.Fatalf("self-name bounce %q unexpectedly failed: %v", test.oracleText, diagnostics)
			}
			if !strings.Contains(source, "game.Bounce") ||
				!strings.Contains(source, "game.SourcePermanentReference()") {
				t.Fatalf("self-name bounce %q did not lower to a source Bounce:\n%s", test.oracleText, source)
			}
		})
	}
}

func TestGenerateExecutableCardSourceItSourceDamage(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Spellbomb",
		Layout:     "normal",
		TypeLine:   "Artifact",
		OracleText: "{R}, Sacrifice this artifact: It deals 2 damage to any target.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Primitive: game.Damage",
		"game.Fixed(2)",
		"game.AnyTargetDamageRecipient(0)",
		"DamageSource: opt.Val(game.SourcePermanentReference())",
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
		"Recipient: game.AnyTargetDamageRecipient(0)",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceSourcePermanentDamage(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Pinger",
		Layout:     "normal",
		ManaCost:   "{2}{R}",
		TypeLine:   "Creature — Human Shaman",
		OracleText: "{8}: This creature deals 3 damage to any target.",
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
		"Primitive: game.Damage",
		"game.Fixed(3)",
		"game.AnyTargetDamageRecipient(0)",
		"DamageSource: opt.Val(game.SourcePermanentReference())",
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

func TestGenerateExecutableCardSourceExileTopOfLibrary(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Exile Top",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Exile the top card of your library.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if !strings.Contains(source, "Primitive: game.ExileTopOfLibrary") {
		t.Fatalf("source missing ExileTopOfLibrary primitive:\n%s", source)
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

func TestGenerateExecutableCardSourceSourceCounterPlacement(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Self Counter",
		Layout:     "normal",
		TypeLine:   "Creature — Beast",
		OracleText: "{T}: Put a +1/+1 counter on this creature.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"Primitive: game.AddCounter",
		"Amount:      game.Fixed(1)",
		"Object:      game.SourcePermanentReference()",
		"CounterKind: counter.PlusOnePlusOne",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("source missing %q:\n%s", want, source)
		}
	}
}

func TestGenerateExecutableCardSourceReferencedControllerDrawDiscard(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		oracle    string
		primitive string
		recipient string
	}{
		// Permanent antecedent → permanent reference.
		{"Destroy target creature. Its controller draws a card.", "game.Draw{", "game.ObjectControllerReference(game.TargetPermanentReference(0))"},
		{"Destroy target creature. Its controller discards a card.", "game.Discard{", "game.ObjectControllerReference(game.TargetPermanentReference(0))"},
		// Spell antecedent (counterspell) → stack-object reference.
		{"Counter target spell. Its controller draws a card.", "game.Draw{", "game.ObjectControllerReference(game.TargetStackObjectReference(0))"},
		{"Counter target spell. Its controller discards a card.", "game.Discard{", "game.ObjectControllerReference(game.TargetStackObjectReference(0))"},
	} {
		source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
			Name:       "Test Spell",
			Layout:     "normal",
			ManaCost:   "{U}{B}",
			TypeLine:   "Instant",
			OracleText: tc.oracle,
		}, "t")
		if err != nil {
			t.Fatal(err)
		}
		if len(diagnostics) != 0 {
			t.Fatalf("%q: diagnostics = %#v", tc.oracle, diagnostics)
		}
		for _, want := range []string{
			"Primitive: " + tc.primitive,
			"Player: " + tc.recipient + ",",
		} {
			if !strings.Contains(source, want) {
				t.Fatalf("%q: source missing %q:\n%s", tc.oracle, want, source)
			}
		}
	}
}

func TestGenerateExecutableCardSourceReferencedControllerLoseLife(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		oracle    string
		recipient string
	}{
		{"Destroy target creature. Its controller loses 2 life.", "game.ObjectControllerReference(game.TargetPermanentReference(0))"},
		{"Counter target spell. Its controller loses 2 life.", "game.ObjectControllerReference(game.TargetStackObjectReference(0))"},
	} {
		source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
			Name:       "Test Drain",
			Layout:     "normal",
			ManaCost:   "{U}{B}",
			TypeLine:   "Instant",
			OracleText: tc.oracle,
		}, "t")
		if err != nil {
			t.Fatal(err)
		}
		if len(diagnostics) != 0 {
			t.Fatalf("%q: diagnostics = %#v", tc.oracle, diagnostics)
		}
		for _, want := range []string{
			"Primitive: game.LoseLife{",
			"Player: " + tc.recipient + ",",
		} {
			if !strings.Contains(source, want) {
				t.Fatalf("%q: source missing %q:\n%s", tc.oracle, want, source)
			}
		}
	}
}

func TestGenerateExecutableCardSourceReferencedControllerDamage(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name      string
		oracle    string
		recipient string
	}{
		// "that <object>'s controller" names the controller of the prior removal
		// target; a permanent antecedent yields a permanent reference.
		{"Test Burn Land", "Destroy target land. Test Burn Land deals 2 damage to that land's controller.", "game.ObjectControllerReference(game.TargetPermanentReference(0))"},
		{"Test Burn Artifact", "Destroy target artifact. Test Burn Artifact deals 3 damage to that artifact's controller.", "game.ObjectControllerReference(game.TargetPermanentReference(0))"},
		// A spell antecedent (counterspell) yields a stack-object reference.
		{"Test Burn Spell", "Counter target spell. Test Burn Spell deals 2 damage to that spell's controller.", "game.ObjectControllerReference(game.TargetStackObjectReference(0))"},
		// "its controller" resolves the same antecedent.
		{"Test Burn Creature", "Destroy target creature. Test Burn Creature deals 2 damage to its controller.", "game.ObjectControllerReference(game.TargetPermanentReference(0))"},
	} {
		source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
			Name:       tc.name,
			Layout:     "normal",
			ManaCost:   "{2}{R}",
			TypeLine:   "Instant",
			OracleText: tc.oracle,
		}, "t")
		if err != nil {
			t.Fatal(err)
		}
		if len(diagnostics) != 0 {
			t.Fatalf("%q: diagnostics = %#v", tc.oracle, diagnostics)
		}
		for _, want := range []string{
			"Primitive: game.Damage{",
			"Recipient: game.PlayerDamageRecipient(" + tc.recipient + ")",
		} {
			if !strings.Contains(source, want) {
				t.Fatalf("%q: source missing %q:\n%s", tc.oracle, want, source)
			}
		}
	}
}

func TestGenerateExecutableCardSourceEventPlayerDamage(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name   string
		oracle string
	}{
		// "Whenever an opponent draws a card, ~ deals N damage to that player."
		// (Underworld Dreams) targets the triggering event's player.
		{"Test Dream Pain", "Whenever an opponent draws a card, this enchantment deals 1 damage to that player."},
		// "Whenever an opponent discards a card, ~ deals N damage to that player."
		// (Megrim) is the same event-player recipient shape.
		{"Test Discard Pain", "Whenever an opponent discards a card, this enchantment deals 2 damage to that player."},
	} {
		source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
			Name:       tc.name,
			Layout:     "normal",
			ManaCost:   "{2}{B}",
			TypeLine:   "Enchantment",
			OracleText: tc.oracle,
		}, "t")
		if err != nil {
			t.Fatal(err)
		}
		if len(diagnostics) != 0 {
			t.Fatalf("%q: diagnostics = %#v", tc.oracle, diagnostics)
		}
		for _, want := range []string{
			"game.Damage{",
			"game.PlayerDamageRecipient(game.EventPlayerReference())",
		} {
			if !strings.Contains(source, want) {
				t.Fatalf("%q: source missing %q:\n%s", tc.oracle, want, source)
			}
		}
	}
}

func TestGenerateExecutableCardSourceEventPlayerSourcePowerDamage(t *testing.T) {
	t.Parallel()
	// "Whenever an opponent casts a noncreature spell, this creature deals
	// damage equal to its power to that player." (Gleeful Arsonist) reads the
	// source creature's power and deals it to the triggering event's player.
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Spite Arsonist",
		Layout:     "normal",
		ManaCost:   "{1}{R}",
		TypeLine:   "Creature — Goblin",
		OracleText: "Whenever an opponent casts a noncreature spell, this creature deals damage equal to its power to that player.",
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"game.Damage{",
		"game.PlayerDamageRecipient(game.EventPlayerReference())",
		"Kind:       game.DynamicAmountObjectPower",
		"Object:     game.SourcePermanentReference()",
		"DamageSource: opt.Val(game.SourcePermanentReference())",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("source missing %q:\n%s", want, source)
		}
	}
}

func TestGenerateExecutableCardSourceInheritedPronounDestroy(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Tap Destroy",
		Layout:     "normal",
		ManaCost:   "{B}",
		TypeLine:   "Instant",
		OracleText: "Tap target creature. Destroy it.",
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"Primitive: game.Tap{",
		"Primitive: game.Destroy{",
		"Object: game.TargetPermanentReference(0),",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("source missing %q:\n%s", want, source)
		}
	}
}

func TestGenerateExecutableCardSourceCounterOnReferencedTarget(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test On It",
		Layout:     "normal",
		ManaCost:   "{G}",
		TypeLine:   "Instant",
		OracleText: "Target creature gets +2/+2 until end of turn. Put a +1/+1 counter on it.",
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"Primitive: game.AddCounter",
		"Object:      game.TargetPermanentReference(0)",
		"CounterKind: counter.PlusOnePlusOne",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("source missing %q:\n%s", want, source)
		}
	}
}

func TestGenerateExecutableCardSourceKeywordGrantOnReferencedTarget(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test It Gains",
		Layout:     "normal",
		ManaCost:   "{G}",
		TypeLine:   "Instant",
		OracleText: "Untap target creature. It gains trample until end of turn.",
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"Primitive: game.ApplyContinuous",
		"Object: opt.Val(game.TargetPermanentReference(0))",
		"game.LayerAbility",
		"game.Trample",
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
	for _, name := range []string{"finality"} {
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

// TestGenerateExecutableCardSourceFightAnotherTarget covers the "fights another
// target creature" templating: the second target lowers to a
// DistinctFromPriorTargets spec so the chosen creature must differ from the
// first fighter.
func TestGenerateExecutableCardSourceFightAnotherTarget(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Prey",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Target creature you control fights another target creature.",
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
	if !strings.Contains(source, "DistinctFromPriorTargets: true,") {
		t.Fatalf("source missing distinct second target:\n%s", source)
	}
}

// TestGenerateExecutableCardSourceFightSubtypeTarget covers a bare
// creature-subtype fight target ("Target Mutant"): the subtype names a creature
// even without the "creature" word, so the first fighter lowers to a creature
// target carrying the subtype filter (The Curse of Fenric III).
func TestGenerateExecutableCardSourceFightSubtypeTarget(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Mutant Brawl",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Target Mutant fights target creature you don't control.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if !strings.Contains(source, "Primitive: game.Fight") {
		t.Fatalf("source missing Fight primitive:\n%s", source)
	}
	if !strings.Contains(source, `SubtypesAny: []types.Sub{types.Sub("Mutant")}`) {
		t.Fatalf("source missing Mutant subtype filter:\n%s", source)
	}
}

// TestGenerateExecutableCardSourceFightNamedTarget covers a "named <Name>" fight
// target ("another target creature named Fenric"): the name lowers to a
// RequiredName predicate so the second fighter must be the named creature (The
// Curse of Fenric III).
func TestGenerateExecutableCardSourceFightNamedTarget(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Named Brawl",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Target creature you control fights another target creature named Fenric.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if !strings.Contains(source, "Primitive: game.Fight") {
		t.Fatalf("source missing Fight primitive:\n%s", source)
	}
	if !strings.Contains(source, `Name: "Fenric"`) {
		t.Fatalf("source missing RequiredName predicate:\n%s", source)
	}
	if !strings.Contains(source, "DistinctFromPriorTargets: true,") {
		t.Fatalf("source missing distinct second target:\n%s", source)
	}
}

func TestGenerateExecutableCardSourceFightOptionalSecondTarget(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Optional Fight",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Target creature you control fights up to one target creature you don't control.",
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
	if !strings.Contains(source, "MinTargets: 0,") ||
		!strings.Contains(source, "MinTargets: 1,") {
		t.Fatalf("source missing mandatory + optional fight targets:\n%s", source)
	}
}

func TestGenerateExecutableCardSourceReferencedFightEventPermanent(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Indrik",
		Layout:     "normal",
		TypeLine:   "Creature — Beast",
		OracleText: "When this creature enters, it fights target creature you don't control.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if !strings.Contains(source, "Primitive: game.Fight") ||
		!strings.Contains(source, "Object:        game.EventPermanentReference()") ||
		!strings.Contains(source, "RelatedObject: game.TargetPermanentReference(0)") {
		t.Fatalf("source missing event-permanent fight primitive:\n%s", source)
	}
}

func TestGenerateExecutableCardSourceReferencedFightSourcePermanent(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Mammoth",
		Layout:     "normal",
		TypeLine:   "Creature — Elephant",
		OracleText: "Whenever this creature or another creature you control enters, this creature fights up to one target creature you don't control.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if !strings.Contains(source, "Primitive: game.Fight") ||
		!strings.Contains(source, "Object:        game.SourcePermanentReference()") ||
		!strings.Contains(source, "RelatedObject: game.TargetPermanentReference(0)") ||
		!strings.Contains(source, "MinTargets: 0,") {
		t.Fatalf("source missing source-permanent fight primitive:\n%s", source)
	}
}

// TestGenerateExecutableCardSourceFightAnotherExcludesSource covers "this
// creature fights another target creature" (Brash Taunter): the fighter is the
// source permanent, so "another" excludes the source via a Predicate.Another
// (ExcludeSource) target rather than a DistinctFromPriorTargets second target.
func TestGenerateExecutableCardSourceFightAnotherExcludesSource(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Taunter",
		Layout:     "normal",
		TypeLine:   "Creature — Goblin",
		OracleText: "{2}{R}, {T}: This creature fights another target creature.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if !strings.Contains(source, "Primitive: game.Fight") ||
		!strings.Contains(source, "Object:        game.SourcePermanentReference()") ||
		!strings.Contains(source, "RelatedObject: game.TargetPermanentReference(0)") {
		t.Fatalf("source missing source-permanent fight primitive:\n%s", source)
	}
	if !strings.Contains(source, "ExcludeSource: true") {
		t.Fatalf("source missing source-excluding Another predicate:\n%s", source)
	}
	if strings.Contains(source, "DistinctFromPriorTargets: true,") {
		t.Fatalf("single source fight target must not use DistinctFromPriorTargets:\n%s", source)
	}
}

func TestGenerateExecutableCardSourceFixedDamageTargets(t *testing.T) {
	t.Parallel()
	tests := []struct {
		target string
		wanted string
	}{
		{target: "creature", wanted: "RequiredTypesAny: []types.Card{types.Creature}"},
		{target: "planeswalker", wanted: "RequiredTypesAny: []types.Card{types.Planeswalker}"},
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

// TestGenerateExecutableCardSourceEachSourceDamage covers the "each <group>
// deals N damage to its controller/owner" shape (Rakdos Charm's third mode):
// every group member is the damage source and the recipient is the player who
// controls (or owns) it, lowered onto a GroupSourceDamage primitive.
func TestGenerateExecutableCardSourceEachSourceDamage(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		oracleText  string
		wantedSnips []string
	}{
		{
			name:       "each creature to its controller",
			oracleText: "Each creature deals 1 damage to its controller.",
			wantedSnips: []string{
				"Primitive: game.GroupSourceDamage",
				"game.Fixed(1)",
				"game.BattlefieldGroup(",
			},
		},
		{
			name:       "each creature to its owner",
			oracleText: "Each creature deals 2 damage to its owner.",
			wantedSnips: []string{
				"Primitive: game.GroupSourceDamage",
				"game.Fixed(2)",
				"ToOwner: true",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{
				Name:       "Test Pulse",
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
			for _, wanted := range test.wantedSnips {
				if !strings.Contains(source, wanted) {
					t.Fatalf("source missing %q:\n%s", wanted, source)
				}
			}
		})
	}
}

// TestGenerateExecutableCardSourceEachSelfPowerDamage covers the group self-power
// damage shape "Each <group> deals damage to itself equal to its power." (Wave of
// Reckoning), where every group member deals its own power to itself, lowered
// onto a GroupSelfPowerDamage primitive. The filtered "each tapped creature"
// variant reuses the SelectionForSelector-backed damage group, so the lowering
// records the tapped filter.
func TestGenerateExecutableCardSourceEachSelfPowerDamage(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		oracleText  string
		wantedSnips []string
	}{
		{
			name:       "each creature",
			oracleText: "Each creature deals damage to itself equal to its power.",
			wantedSnips: []string{
				"Primitive: game.GroupSelfPowerDamage",
				"game.BattlefieldGroup(",
				"types.Creature",
			},
		},
		{
			name:       "each tapped creature",
			oracleText: "Each tapped creature deals damage to itself equal to its power.",
			wantedSnips: []string{
				"Primitive: game.GroupSelfPowerDamage",
				"Tapped:",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{
				Name:       "Test Reckoning",
				Layout:     "normal",
				ManaCost:   "{4}{W}",
				TypeLine:   "Sorcery",
				OracleText: test.oracleText,
				Colors:     []string{"W"},
			}
			source, diagnostics, err := GenerateExecutableCardSource(card, "t")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			for _, wanted := range test.wantedSnips {
				if !strings.Contains(source, wanted) {
					t.Fatalf("source missing %q:\n%s", wanted, source)
				}
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

// TestGenerateExecutableCardSourceGroupDamageVariableX covers X-amount group
// damage: the dealt amount is the spell's X, applied to each member of the
// filtered creature group and, in the Earthquake shape, to every player.
func TestGenerateExecutableCardSourceGroupDamageVariableX(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		oracleText  string
		wantedSnips []string
	}{
		{
			name:       "each creature",
			oracleText: "Test Bolt deals X damage to each creature.",
			wantedSnips: []string{
				"game.Dynamic(game.DynamicAmount{",
				"Kind: game.DynamicAmountX",
				"game.GroupDamageRecipient(game.BattlefieldGroup(",
			},
		},
		{
			name:       "each creature and each player",
			oracleText: "Test Bolt deals X damage to each creature without flying and each player.",
			wantedSnips: []string{
				"Kind: game.DynamicAmountX",
				"game.GroupDamageRecipient(game.BattlefieldGroup(",
				"game.PlayerGroupDamageRecipient(game.AllPlayersReference())",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{
				Name:       "Test Bolt",
				Layout:     "normal",
				ManaCost:   "{X}{R}",
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
			for _, wanted := range append([]string{"Primitive: game.Damage"}, test.wantedSnips...) {
				if !strings.Contains(source, wanted) {
					t.Fatalf("source missing %q:\n%s", wanted, source)
				}
			}
		})
	}
}

// TestGenerateExecutableCardSourceFilteredGroupDamage covers fixed-amount group
// damage spells whose recipients form a single filtered permanent group. Each
// case asserts the recipient Selection captures the exact filter (controller,
// combat, tapped, color, subtype, excluded type, planeswalker, "other") so that
// no filter is silently dropped or approximated.
func TestGenerateExecutableCardSourceFilteredGroupDamage(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		oracleText  string
		wantedSnips []string
	}{
		{
			name:       "controller opponents",
			oracleText: "Test Bolt deals 1 damage to each creature your opponents control.",
			wantedSnips: []string{
				"game.GroupDamageRecipient(game.BattlefieldGroup(",
				"RequiredTypes: []types.Card{types.Creature}",
				"Controller: game.ControllerOpponent",
			},
		},
		{
			name:       "controller you with other",
			oracleText: "Test Bolt deals 1 damage to each other creature you control.",
			wantedSnips: []string{
				"game.GroupDamageRecipient(game.BattlefieldGroupExcluding(",
				"Controller: game.ControllerYou",
				"game.SourcePermanentReference()",
			},
		},
		{
			name:        "controller not you",
			oracleText:  "Test Bolt deals 1 damage to each creature you don't control.",
			wantedSnips: []string{"Controller: game.ControllerNotYou"},
		},
		{
			name:        "attacking",
			oracleText:  "Test Bolt deals 2 damage to each attacking creature.",
			wantedSnips: []string{"CombatState: game.CombatStateAttacking"},
		},
		{
			name:        "blocking",
			oracleText:  "Test Bolt deals 2 damage to each blocking creature.",
			wantedSnips: []string{"CombatState: game.CombatStateBlocking"},
		},
		{
			name:        "tapped",
			oracleText:  "Test Bolt deals 1 damage to each tapped creature.",
			wantedSnips: []string{"Tapped: game.TriTrue"},
		},
		{
			name:        "color",
			oracleText:  "Test Bolt deals 3 damage to each red creature.",
			wantedSnips: []string{"ColorsAny: []color.Color{color.Red}"},
		},
		{
			name:        "subtype",
			oracleText:  "Test Bolt deals 1 damage to each Goblin creature.",
			wantedSnips: []string{`SubtypesAny: []types.Sub{types.Sub("Goblin")}`},
		},
		{
			name:        "excluded type",
			oracleText:  "Test Bolt deals 3 damage to each nonartifact creature.",
			wantedSnips: []string{"ExcludedTypes: []types.Card{types.Artifact}"},
		},
		{
			name:        "planeswalker",
			oracleText:  "Test Bolt deals 1 damage to each planeswalker.",
			wantedSnips: []string{"RequiredTypes: []types.Card{types.Planeswalker}"},
		},
		{
			name:        "keyword flying",
			oracleText:  "Test Bolt deals 1 damage to each creature with flying.",
			wantedSnips: []string{"Keyword: game.Flying"},
		},
		{
			name:       "mana value",
			oracleText: "Test Bolt deals 3 damage to each creature with mana value 3 or less.",
			wantedSnips: []string{
				"RequiredTypes: []types.Card{types.Creature}",
				"ManaValue: opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 3})",
			},
		},
		{
			name:       "power",
			oracleText: "Test Bolt deals 2 damage to each creature with power 2 or greater.",
			wantedSnips: []string{
				"Power: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 2})",
			},
		},
		{
			name:       "toughness",
			oracleText: "Test Bolt deals 4 damage to each creature with toughness 4 or greater.",
			wantedSnips: []string{
				"Toughness: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 4})",
			},
		},
		{
			name:       "controller you with mana value",
			oracleText: "Test Bolt deals 2 damage to each creature you control with mana value 2 or less.",
			wantedSnips: []string{
				"Controller: game.ControllerYou",
				"ManaValue: opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 2})",
			},
		},
		{
			name:       "excluded subtype",
			oracleText: "Test Bolt deals 2 damage to each non-Dragon creature.",
			wantedSnips: []string{
				"RequiredTypes: []types.Card{types.Creature}",
				`ExcludedSubtype: types.Sub("Dragon")`,
			},
		},
		{
			name:       "excluded planeswalker subtype",
			oracleText: "Test Bolt deals 5 damage to each non-Bolas planeswalker.",
			wantedSnips: []string{
				"RequiredTypes: []types.Card{types.Planeswalker}",
				`ExcludedSubtype: types.Sub("Bolas")`,
			},
		},
		{
			name:       "nontoken",
			oracleText: "Test Bolt deals 3 damage to each nontoken creature.",
			wantedSnips: []string{
				"RequiredTypes: []types.Card{types.Creature}",
				"NonToken: true",
			},
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
			for _, wanted := range append([]string{"Primitive: game.Damage"}, test.wantedSnips...) {
				if !strings.Contains(source, wanted) {
					t.Fatalf("source missing %q:\n%s", wanted, source)
				}
			}
		})
	}
}

// TestGenerateExecutableCardSourceDualRecipientGroupDamage covers the dual
// recipient board-sweep wording "deals N damage to each X and each Y". Each
// recipient group is damaged by its own fixed Damage instruction in Oracle
// order, reusing the single-recipient primitives, so the union of the two groups
// is represented without a runtime change.
func TestGenerateExecutableCardSourceDualRecipientGroupDamage(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		oracleText  string
		wantedSnips []string
	}{
		{
			name:       "each creature and each player",
			oracleText: "Test Bolt deals 3 damage to each creature and each player.",
			wantedSnips: []string{
				"game.GroupDamageRecipient(game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}}))",
				"game.PlayerGroupDamageRecipient(game.AllPlayersReference())",
			},
		},
		{
			name:       "each creature and each planeswalker",
			oracleText: "Test Bolt deals 5 damage to each creature and each planeswalker.",
			wantedSnips: []string{
				"game.GroupDamageRecipient(game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}}))",
				"game.GroupDamageRecipient(game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Planeswalker}}))",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{
				Name:       "Test Bolt",
				Layout:     "normal",
				ManaCost:   "{R}",
				TypeLine:   "Sorcery",
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
			// Both recipient groups must be damaged by separate instructions.
			if got := strings.Count(source, "Primitive: game.Damage"); got != 2 {
				t.Fatalf("expected 2 damage instructions, got %d:\n%s", got, source)
			}
			for _, wanted := range test.wantedSnips {
				if !strings.Contains(source, wanted) {
					t.Fatalf("source missing %q:\n%s", wanted, source)
				}
			}
		})
	}
}

// TestGenerateExecutableCardSourceFilteredGroupDamageFailsClosed asserts that
// group damage wordings the executable backend cannot represent exactly stay
// rejected rather than being silently approximated. In particular a multi-color
// filter must never be reduced to plain "each creature".
func TestGenerateExecutableCardSourceFilteredGroupDamageFailsClosed(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		oracleText string
	}{
		{
			name:       "divided damage variable amount",
			oracleText: "Test Bolt deals X damage divided as you choose among any number of target creatures.",
		},
		{
			name:       "multi-color filter not dropped",
			oracleText: "Test Bolt deals 1 damage to each white and blue creature.",
		},
		{
			name:       "dual recipient leading player",
			oracleText: "Test Bolt deals 3 damage to you and each creature you control.",
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
			if len(diagnostics) == 0 {
				t.Fatalf("expected a diagnostic rejecting %q, got source:\n%s", test.oracleText, source)
			}
		})
	}
}

// TestGenerateExecutableCardSourceControllerDamage covers the self-directed
// wording "deals N damage to you", where the recipient is the controller of the
// resolving spell. The single Damage instruction must target the controller
// reference rather than a chosen target.
func TestGenerateExecutableCardSourceControllerDamage(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		oracleText string
		wantedSnip string
	}{
		{
			name:       "fixed amount",
			oracleText: "Test Bolt deals 3 damage to you.",
			wantedSnip: "game.Fixed(3)",
		},
		{
			name:       "variable amount",
			oracleText: "Test Bolt deals X damage to you.",
			wantedSnip: "game.DynamicAmountX",
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
			for _, wanted := range []string{
				"Primitive: game.Damage",
				"game.PlayerDamageRecipient(game.ControllerReference())",
				test.wantedSnip,
			} {
				if !strings.Contains(source, wanted) {
					t.Fatalf("source missing %q:\n%s", wanted, source)
				}
			}
		})
	}
}

// TestGenerateExecutableCardSourceSelfDamageRider covers the rider wording
// "deals A damage to <target> and B damage to you" (Char, Psionic Blast). The
// primary target and the controller are each damaged by their own fixed Damage
// instruction in Oracle order, reusing the single-recipient primitives.
func TestGenerateExecutableCardSourceSelfDamageRider(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Bolt",
		Layout:     "normal",
		ManaCost:   "{R}",
		TypeLine:   "Instant",
		OracleText: "Test Bolt deals 4 damage to any target and 2 damage to you.",
		Colors:     []string{"R"},
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if got := strings.Count(source, "Primitive: game.Damage"); got != 2 {
		t.Fatalf("expected 2 damage instructions, got %d:\n%s", got, source)
	}
	for _, wanted := range []string{
		"game.AnyTargetDamageRecipient(0)",
		"game.Fixed(4)",
		"game.PlayerDamageRecipient(game.ControllerReference())",
		"game.Fixed(2)",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

// TestGenerateExecutableCardSourceSelfDamageRiderFailsClosed asserts that
// damage riders the executable backend cannot represent exactly stay rejected
// rather than being silently approximated. A second recipient that is neither
// "you", the target's controller/owner, nor another target, and a variable rider
// pairing, must never collapse onto a supported rider form.
func TestGenerateExecutableCardSourceSelfDamageRiderFailsClosed(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		oracleText string
	}{
		{
			name:       "rider to each opponent",
			oracleText: "Test Bolt deals 4 damage to target creature and 2 damage to each opponent.",
		},
		{
			name:       "variable primary with rider",
			oracleText: "Test Bolt deals X damage to any target and 2 damage to you.",
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
			if len(diagnostics) == 0 {
				t.Fatalf("expected a diagnostic rejecting %q, got source:\n%s", test.oracleText, source)
			}
		})
	}
}

// TestGenerateExecutableCardSourceTargetControllerDamageRider covers the rider
// wording "deals A damage to <target> and B damage to that creature's/permanent's
// controller/owner" (First Volley, Chandra's Outrage, Unleash Shell). The primary
// target and the player derived from that target are each damaged by their own
// fixed Damage instruction in Oracle order.
func TestGenerateExecutableCardSourceTargetControllerDamageRider(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		oracleText string
		recipient  string
	}{
		{
			name:       "controller of that creature",
			oracleText: "Test Bolt deals 4 damage to target creature and 2 damage to that creature's controller.",
			recipient:  "game.PlayerDamageRecipient(game.ObjectControllerReference(game.TargetPermanentReference(0)))",
		},
		{
			name:       "controller via its",
			oracleText: "Test Bolt deals 1 damage to target creature and 1 damage to its controller.",
			recipient:  "game.PlayerDamageRecipient(game.ObjectControllerReference(game.TargetPermanentReference(0)))",
		},
		{
			name:       "owner of that permanent",
			oracleText: "Test Bolt deals 5 damage to target creature or planeswalker and 2 damage to that permanent's owner.",
			recipient:  "game.PlayerDamageRecipient(game.ObjectOwnerReference(game.TargetPermanentReference(0)))",
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
			if got := strings.Count(source, "Primitive: game.Damage"); got != 2 {
				t.Fatalf("expected 2 damage instructions, got %d:\n%s", got, source)
			}
			for _, wanted := range []string{
				"game.AnyTargetDamageRecipient(0)",
				test.recipient,
			} {
				if !strings.Contains(source, wanted) {
					t.Fatalf("source missing %q:\n%s", wanted, source)
				}
			}
		})
	}
}

// TestGenerateExecutableCardSourceTargetControllerDamageRiderFailsClosed asserts
// that target-controller riders the backend cannot represent exactly stay
// rejected. A variable primary amount and an unrecognized rider amount must not
// be approximated.
func TestGenerateExecutableCardSourceTargetControllerDamageRiderFailsClosed(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		oracleText string
	}{
		{
			name:       "variable primary with controller rider",
			oracleText: "Test Bolt deals X damage to target creature and 2 damage to that creature's controller.",
		},
		{
			name:       "variable rider amount",
			oracleText: "Test Bolt deals 4 damage to target creature and X damage to that creature's controller.",
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
			if len(diagnostics) == 0 {
				t.Fatalf("expected a diagnostic rejecting %q, got source:\n%s", test.oracleText, source)
			}
		})
	}
}

// TestGenerateExecutableCardSourceTwoTargetDamageRider covers the two-target
// wording "deals A damage to <target0> and B damage to <target1>" (Hungry Flames,
// Lunge, Punish the Enemy, Reckless Rage). Each target receives its own fixed
// Damage instruction keyed to its own target occurrence, in Oracle order.
func TestGenerateExecutableCardSourceTwoTargetDamageRider(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		oracleText string
	}{
		{
			name:       "creature then player or planeswalker",
			oracleText: "Test Bolt deals 3 damage to target creature and 2 damage to target player or planeswalker.",
		},
		{
			name:       "player or planeswalker then creature",
			oracleText: "Test Bolt deals 3 damage to target player or planeswalker and 3 damage to target creature.",
		},
		{
			name:       "two creatures with control predicates",
			oracleText: "Test Bolt deals 4 damage to target creature you don't control and 2 damage to target creature you control.",
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
			if got := strings.Count(source, "Primitive: game.Damage"); got != 2 {
				t.Fatalf("expected 2 damage instructions, got %d:\n%s", got, source)
			}
			for _, wanted := range []string{
				"game.AnyTargetDamageRecipient(0)",
				"game.AnyTargetDamageRecipient(1)",
			} {
				if !strings.Contains(source, wanted) {
					t.Fatalf("source missing %q:\n%s", wanted, source)
				}
			}
		})
	}
}

// TestGenerateExecutableCardSourceTwoTargetAnyOtherDamage covers the "deals A
// damage to any target and B damage to any other target" wording (Boulder Dash,
// Arc Trail). Both targets are "any target" slots; the second carries the
// "other" distinctness qualifier, so its TargetSpec must require a different
// object from the first.
func TestGenerateExecutableCardSourceTwoTargetAnyOtherDamage(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Bolt",
		Layout:     "normal",
		ManaCost:   "{1}{R}",
		TypeLine:   "Sorcery",
		OracleText: "Test Bolt deals 2 damage to any target and 1 damage to any other target.",
		Colors:     []string{"R"},
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if got := strings.Count(source, "Primitive: game.Damage"); got != 2 {
		t.Fatalf("expected 2 damage instructions, got %d:\n%s", got, source)
	}
	for _, wanted := range []string{
		"game.AnyTargetDamageRecipient(0)",
		"game.AnyTargetDamageRecipient(1)",
		"DistinctFromPriorTargets: true",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
	// Only the second ("other") target slot is distinct from the first.
	if got := strings.Count(source, "DistinctFromPriorTargets: true"); got != 1 {
		t.Fatalf("expected exactly one distinct target slot, got %d:\n%s", got, source)
	}
}

// TestGenerateExecutableCardSourceTwoTargetDamageRiderFailsClosed asserts that
// two-target damage shapes the backend cannot represent exactly stay rejected. A
// variable primary or rider amount, and an "any target" second clause (whose
// recipient is not introduced by "target"), must not be approximated.
func TestGenerateExecutableCardSourceTwoTargetDamageRiderFailsClosed(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		oracleText string
	}{
		{
			name:       "variable primary",
			oracleText: "Test Bolt deals X damage to target creature and 2 damage to target player or planeswalker.",
		},
		{
			name:       "variable rider",
			oracleText: "Test Bolt deals 3 damage to target creature and X damage to target player or planeswalker.",
		},
		{
			name:       "any target second clause",
			oracleText: "Test Bolt deals 1 damage to target creature and 1 damage to any target.",
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
			if len(diagnostics) == 0 {
				t.Fatalf("expected a diagnostic rejecting %q, got source:\n%s", test.oracleText, source)
			}
		})
	}
}

// TestGenerateExecutableCardSourceGroupDynamicCountDamage covers single-recipient
// group damage whose amount is a trailing "where X is the number of ..." count
// phrase. The dealt amount is a battlefield count selector resolved once and
// applied to every member of the recipient group; the recipient phrase is scoped
// to the tokens before the count clause, so the count subject's filters (here
// "Gate" controlled by you) bind to the count selector, not the recipient.
func TestGenerateExecutableCardSourceGroupDynamicCountDamage(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		oracleText  string
		wantedSnips []string
	}{
		{
			name:       "count of subtype you control",
			oracleText: "Test Bolt deals X damage to each creature, where X is the number of Gates you control.",
			wantedSnips: []string{
				"Kind:       game.DynamicAmountCountSelector",
				`game.BattlefieldGroup(game.Selection{SubtypesAny: []types.Sub{types.Sub("Gate")}, Controller: game.ControllerYou})`,
				"Recipient: game.GroupDamageRecipient(game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}}))",
			},
		},
		{
			name:       "count of all creatures",
			oracleText: "Test Bolt deals X damage to each creature, where X is the number of creatures on the battlefield.",
			wantedSnips: []string{
				"Kind:       game.DynamicAmountCountSelector",
				"Recipient: game.GroupDamageRecipient(game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}}))",
			},
		},
		{
			name:       "count of tapped permanents you control",
			oracleText: "Test Bolt deals X damage to each creature, where X is the number of tapped creatures you control.",
			wantedSnips: []string{
				"Kind:       game.DynamicAmountCountSelector",
				"Tapped: game.TriTrue",
				"Recipient: game.GroupDamageRecipient(game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}}))",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{
				Name:       "Test Bolt",
				Layout:     "normal",
				ManaCost:   "{X}{R}",
				TypeLine:   "Sorcery",
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
			for _, wanted := range append([]string{"Primitive: game.Damage"}, test.wantedSnips...) {
				if !strings.Contains(source, wanted) {
					t.Fatalf("source missing %q:\n%s", wanted, source)
				}
			}
		})
	}
}

// TestGenerateExecutableCardSourceGroupDynamicCountDamageFailsClosed verifies
// that dynamic group damage outside the supported single-recipient group-wide
// shape stays rejected: a two-recipient spell cannot be modeled as one group, so
// it yields an unsupported diagnostic rather than an approximate lowering.
func TestGenerateExecutableCardSourceGroupDynamicCountDamageFailsClosed(t *testing.T) {
	t.Parallel()
	for _, oracleText := range []string{
		"Test Bolt deals X damage to each creature without flying and each player, where X is the number of Beasts on the battlefield.",
	} {
		t.Run(oracleText, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{
				Name:       "Test Bolt",
				Layout:     "normal",
				ManaCost:   "{X}{R}",
				TypeLine:   "Sorcery",
				OracleText: oracleText,
				Colors:     []string{"R"},
			}
			_, diagnostics, err := GenerateExecutableCardSource(card, "t")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) == 0 {
				t.Fatalf("expected unsupported diagnostic for %q", oracleText)
			}
		})
	}
}

// TestGenerateExecutableCardSourceGroupDynamicEqualToDamage covers single-recipient
// group damage whose amount is a trailing "equal to ..." dynamic phrase, the
// sibling of the "where X is ..." form, and exercises the group-wide dynamic
// amount kinds the executable backend resolves once and deals to every member of
// the recipient group: a battlefield count, controller devotion, and controller
// domain (basic land type count). It also covers the player-group recipients
// ("each opponent", "each player") that the dynamic group damage path shares with
// the fixed group damage path.
func TestGenerateExecutableCardSourceGroupDynamicEqualToDamage(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		oracleText  string
		wantedSnips []string
	}{
		{
			name:       "devotion to each opponent",
			oracleText: "Test Bolt deals damage to each opponent equal to your devotion to red.",
			wantedSnips: []string{
				"Kind:       game.DynamicAmountDevotion",
				"Recipient: game.PlayerGroupDamageRecipient(game.OpponentsReference())",
			},
		},
		{
			name:       "count you control to each opponent",
			oracleText: "Test Bolt deals damage to each opponent equal to the number of creatures you control.",
			wantedSnips: []string{
				"Kind:       game.DynamicAmountCountSelector",
				"game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou})",
				"Recipient: game.PlayerGroupDamageRecipient(game.OpponentsReference())",
			},
		},
		{
			name:       "domain to each player",
			oracleText: "Test Bolt deals damage to each player equal to the number of basic land types among lands you control.",
			wantedSnips: []string{
				"Kind:       game.DynamicAmountControllerBasicLandTypeCount",
				"Recipient: game.PlayerGroupDamageRecipient(game.AllPlayersReference())",
			},
		},
		{
			name:       "domain to each creature where X",
			oracleText: "Test Bolt deals X damage to each creature, where X is the number of basic land types among lands you control.",
			wantedSnips: []string{
				"Kind:       game.DynamicAmountControllerBasicLandTypeCount",
				"Recipient: game.GroupDamageRecipient(game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}}))",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{
				Name:       "Test Bolt",
				Layout:     "normal",
				ManaCost:   "{X}{R}",
				TypeLine:   "Sorcery",
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
			for _, wanted := range append([]string{"Primitive: game.Damage"}, test.wantedSnips...) {
				if !strings.Contains(source, wanted) {
					t.Fatalf("source missing %q:\n%s", wanted, source)
				}
			}
		})
	}
}

// TestGenerateExecutableGreatestCharacteristicDraw verifies that "draw cards
// equal to the greatest <characteristic> among <group>" lowers to a dynamic
// draw whose amount measures the greatest power, toughness, or mana value of a
// battlefield group.
func TestGenerateExecutableGreatestCharacteristicDraw(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		oracleText string
		kind       string
		group      string
	}{
		{
			oracleText: "Draw cards equal to the greatest power among creatures you control.",
			kind:       "game.DynamicAmountGreatestPowerInGroup",
			group:      "game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou})",
		},
		{
			oracleText: "Draw cards equal to the greatest toughness among creatures you control.",
			kind:       "game.DynamicAmountGreatestToughnessInGroup",
			group:      "game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou})",
		},
		{
			oracleText: "Draw cards equal to the greatest mana value among permanents you control.",
			kind:       "game.DynamicAmountGreatestManaValueInGroup",
			group:      "game.BattlefieldGroup(game.Selection{Controller: game.ControllerYou})",
		},
	} {
		t.Run(tc.oracleText, func(t *testing.T) {
			t.Parallel()
			source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
				Name:       "Test Knowledge",
				Layout:     "normal",
				TypeLine:   "Sorcery",
				OracleText: tc.oracleText,
			}, "t")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			for _, wanted := range []string{"Primitive: game.Draw{", tc.kind, tc.group} {
				if !strings.Contains(source, wanted) {
					t.Fatalf("source missing %q:\n%s", wanted, source)
				}
			}
		})
	}
}

func TestGenerateExecutableExileTopOfEachPlayersLibrary(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Library Raider",
		Layout:     "normal",
		TypeLine:   "Creature — Dinosaur",
		ManaCost:   "{4}{R}{R}",
		Power:      new("6"),
		Toughness:  new("6"),
		OracleText: "Whenever Library Raider attacks, exile the top card of each player's library.",
	}, "l")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"Primitive: game.ExileTopOfLibrary{",
		"PlayerGroup: game.AllPlayersReference(),",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated card missing %q:\n%s", want, source)
		}
	}
}
