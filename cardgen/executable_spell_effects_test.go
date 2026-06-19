package cardgen

import (
	"strings"
	"testing"
)

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

func TestLowerMassBounceFailsClosedForUnexpressibleGroups(t *testing.T) {
	t.Parallel()
	for _, oracleText := range []string{
		"Return all permanents of the color of your choice to their owners' hands.",
		"Return all creatures to their owners' hands except for Krakens, Leviathans, Octopuses, and Serpents.",
		"Return all but one creature to their owners' hands.",
		"Return a permanent you control to its owner's hand.",
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
	for _, name := range []string{"stun", "finality"} {
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
			name:       "variable amount",
			oracleText: "Test Bolt deals X damage to each creature.",
		},
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
		{
			name:       "dual recipient variable amount",
			oracleText: "Test Bolt deals X damage to each creature and each player.",
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
// rather than being silently approximated. A non-"you" second recipient or a
// variable rider amount must never collapse onto the controller-self form.
func TestGenerateExecutableCardSourceSelfDamageRiderFailsClosed(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		oracleText string
	}{
		{
			name:       "rider to target controller",
			oracleText: "Test Bolt deals 4 damage to target creature and 2 damage to its controller.",
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
