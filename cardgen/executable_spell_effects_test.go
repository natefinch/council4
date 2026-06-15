package cardgen

import (
	"strings"
	"testing"
)

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
	}{
		{"Destroy target creature. Its controller draws a card.", "game.Draw{"},
		{"Destroy target creature. Its controller discards a card.", "game.Discard{"},
	} {
		source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
			Name:       "Test Spell",
			Layout:     "normal",
			ManaCost:   "{B}",
			TypeLine:   "Sorcery",
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
			"Player: game.ObjectControllerReference(game.TargetPermanentReference(0)),",
		} {
			if !strings.Contains(source, want) {
				t.Fatalf("%q: source missing %q:\n%s", tc.oracle, want, source)
			}
		}
	}
}

func TestGenerateExecutableCardSourceReferencedControllerLoseLife(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Drain Destroy",
		Layout:     "normal",
		ManaCost:   "{B}",
		TypeLine:   "Sorcery",
		OracleText: "Destroy target creature. Its controller loses 2 life.",
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"Primitive: game.Destroy{",
		"Primitive: game.LoseLife{",
		"Player: game.ObjectControllerReference(game.TargetPermanentReference(0)),",
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
