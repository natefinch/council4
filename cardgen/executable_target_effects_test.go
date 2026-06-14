package cardgen

import (
	"fmt"
	"strings"
	"testing"
)

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
				"Player: game.ControllerReference()",
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
		"Player: game.TargetPlayerReference(0)",
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
		"Object: game.TargetPermanentReference(0)",
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

func TestGenerateExecutableCardSourceTypeUnionTarget(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		oracleText string
		wanted     []string
	}{
		{
			name:       "destroy artifact or enchantment",
			oracleText: "Destroy target artifact or enchantment.",
			wanted: []string{
				`Constraint: "target artifact or enchantment"`,
				"PermanentTypes: []types.Card{types.Artifact, types.Enchantment}",
				"Primitive: game.Destroy",
			},
		},
		{
			name:       "damage creature or planeswalker",
			oracleText: "Test Bolt deals 3 damage to target creature or planeswalker.",
			wanted: []string{
				`Constraint: "target creature or planeswalker"`,
				"PermanentTypes: []types.Card{types.Creature, types.Planeswalker}",
				"Allow:      game.TargetAllowPermanent,",
				"Primitive: game.Damage",
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
			for _, wanted := range test.wanted {
				if !strings.Contains(source, wanted) {
					t.Fatalf("source missing %q:\n%s", wanted, source)
				}
			}
		})
	}
}

func TestGenerateExecutableCardSourceExcludedTypeTarget(t *testing.T) {
	t.Parallel()
	tests := []struct {
		oracleText string
		wanted     []string
	}{
		{
			oracleText: "Destroy target nonland permanent.",
			wanted: []string{
				`Constraint: "target nonland permanent"`,
				"ExcludedTypes: []types.Card{types.Land}",
			},
		},
		{
			oracleText: "Destroy target noncreature permanent.",
			wanted: []string{
				`Constraint: "target noncreature permanent"`,
				"ExcludedTypes: []types.Card{types.Creature}",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.oracleText, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{
				Name:       "Test Doom",
				Layout:     "normal",
				ManaCost:   "{2}{B}",
				TypeLine:   "Sorcery",
				OracleText: test.oracleText,
				Colors:     []string{"B"},
			}
			source, diagnostics, err := GenerateExecutableCardSource(card, "t")
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

func TestGenerateExecutableCardSourceDamageEqualToCount(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Raid",
		Layout:     "normal",
		ManaCost:   "{2}{R}",
		TypeLine:   "Sorcery",
		OracleText: "Test Raid deals damage to any target equal to the number of creatures you control.",
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
		`Constraint: "any target"`,
		"Primitive: game.Damage",
		"DynamicAmount",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceDestroyAllCreatures(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Wrath",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Destroy all creatures.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Primitive: game.Destroy",
		"Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}})",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceDestroyAllOtherCreatures(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test One-Sided Wrath",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Destroy all other creatures.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Primitive: game.Destroy",
		"Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, ExcludeSource: true})",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceModifyTargetCreature(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Weakness",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Target creature gets -4/-4 until end of turn.",
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
		"Primitive: game.ModifyPT",
		"PowerDelta:",
		"game.Fixed(-4)",
		"ToughnessDelta:",
		"game.DurationUntilEndOfTurn",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourcePumpTargetCreature(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Growth",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Target creature gets +2/+0 until end of turn.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{"game.Fixed(2)", "game.Fixed(0)"} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceTemporaryContinuousEffects(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		oracleText string
		wants      []string
	}{
		{
			name:       "group power toughness",
			oracleText: "Creatures you control get +1/+1 until end of turn.",
			wants: []string{
				"game.ApplyContinuous",
				"game.BattlefieldGroup",
				"Controller: game.ControllerYou",
				"game.LayerPowerToughnessModify",
				"PowerDelta:",
				"ToughnessDelta:",
			},
		},
		{
			name:       "target keyword",
			oracleText: "Target creature gains flying until end of turn.",
			wants: []string{
				"game.ApplyContinuous",
				"Object: opt.Val(game.TargetPermanentReference(0))",
				"game.LayerAbility",
				"game.Flying",
			},
		},
		{
			name:       "target power toughness and keyword",
			oracleText: "Target creature gets +2/+2 and gains trample until end of turn.",
			wants: []string{
				"game.ApplyContinuous",
				"game.LayerPowerToughnessModify",
				"PowerDelta:",
				"ToughnessDelta:",
				"game.LayerAbility",
				"game.Trample",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
				Name:       "Test Effect",
				Layout:     "normal",
				TypeLine:   "Instant",
				OracleText: test.oracleText,
			}, "t")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			for _, want := range test.wants {
				if !strings.Contains(source, want) {
					t.Fatalf("source missing %q:\n%s", want, source)
				}
			}
		})
	}
}

func TestGenerateExecutableCardSourceNegativeZeroToughness(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Shrink",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Target creature gets -5/-0 until end of turn.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{"game.Fixed(-5)", "game.Fixed(0)"} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceExileCreature(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Exile",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Exile target creature.",
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
		"Primitive: game.Exile",
		"Object: game.TargetPermanentReference(0)",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceBounceCreature(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Bounce",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Return target creature to its owner's hand.",
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
		"Primitive: game.Bounce",
		"Object: game.TargetPermanentReference(0)",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
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
		"Player: game.ControllerReference()",
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
		{text: "You lose 2 life.", primitive: "game.LoseLife", recipient: "game.ControllerReference()"},
		{text: "Target player gains 4 life.", primitive: "game.GainLife", recipient: "Player: game.TargetPlayerReference(0)"},
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

func TestGenerateExecutableCardSourceLifeGroupRecipients(t *testing.T) {
	t.Parallel()
	tests := []struct {
		text    string
		wanteds []string
	}{
		{
			text: "Each opponent loses 3 life.",
			wanteds: []string{
				"game.LoseLife",
				"game.Fixed(3)",
				"PlayerGroup: game.OpponentsReference()",
			},
		},
		{
			text: "Each player gains 2 life.",
			wanteds: []string{
				"game.GainLife",
				"game.Fixed(2)",
				"PlayerGroup: game.AllPlayersReference()",
			},
		},
		{
			text: "Each opponent loses 1 life and you gain 1 life.",
			wanteds: []string{
				"game.LoseLife",
				"PlayerGroup: game.OpponentsReference()",
				`PublishResult: game.ResultKey("life-change")`,
				"game.GainLife",
				"Player: game.ControllerReference()",
			},
		},
		{
			text: "Each opponent loses 1 life and you gain that much life.",
			wanteds: []string{
				"game.LoseLife",
				"PlayerGroup: game.OpponentsReference()",
				`PublishResult: game.ResultKey("life-change")`,
				"game.GainLife",
				"game.DynamicAmountPreviousEffectResult",
				`ResultKey: game.ResultKey("life-change")`,
			},
		},
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
			for _, wanted := range test.wanteds {
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
				"Player: game.ControllerReference()",
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
		"Player: game.TargetPlayerReference(0)",
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
		"Player: game.ControllerReference()",
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
		"Object: game.TargetPermanentReference(0)",
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
		"Player: game.TargetPlayerReference(0)",
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
		"Player: game.ControllerReference()",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}
