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

func TestGenerateExecutableCardSourceControllerAndTargetDraw(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Shared Draw",
		Layout:     "normal",
		ManaCost:   "{2}{U}",
		TypeLine:   "Sorcery",
		OracleText: "You and target opponent each draw a card.",
		Colors:     []string{"U"},
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	// The controller and the player target each draw: two parallel draw
	// instructions, one for the controller and one for the target opponent.
	for _, wanted := range []string{
		`Constraint: "target opponent"`,
		"game.TargetAllowPlayer",
		"Player: game.ControllerReference()",
		"Player: game.TargetPlayerReference(0)",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
	if got := strings.Count(source, "Primitive: game.Draw"); got != 2 {
		t.Fatalf("draw instruction count = %d, want 2:\n%s", got, source)
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

func TestGenerateExecutableCardSourceConditionalEffectInSequence(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Conditional Draw",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Target creature gets -1/-1 until end of turn. If you control a Faerie, draw a card.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Primitive: game.Draw{",
		"Condition: opt.Val(game.EffectCondition{",
		"Condition: opt.Val(game.Condition{",
		"ControlsMatching: opt.Val(game.SelectionCount{",
		`types.Sub("Faerie")`,
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
	// The condition must gate the draw, not the unconditional first effect.
	drawIdx := strings.Index(source, "game.Draw{")
	condIdx := strings.Index(source, "game.EffectCondition{")
	if condIdx < drawIdx {
		t.Fatalf("expected the condition to gate the draw, not the modify:\n%s", source)
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

func TestGenerateExecutableCardSourcePumpSourceCreature(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Self Pump",
		Layout:     "normal",
		TypeLine:   "Creature — Elemental",
		OracleText: "{1}{R}: This creature gets +2/+0 until end of turn.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"game.ModifyPT{",
		"Object:         game.SourcePermanentReference(),",
		"PowerDelta:     game.Fixed(2),",
		"ToughnessDelta: game.Fixed(0),",
		"Duration:       game.DurationUntilEndOfTurn,",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceVariableXPumpTargetCreature(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Untamed Might",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Target creature gets +X/+X until end of turn.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"game.ModifyPT{",
		"Object: game.TargetPermanentReference(0),",
		"game.DynamicAmountX",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceVariableXShrinkTargetCreature(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Death Wind",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Target creature gets -X/-X until end of turn.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"game.DynamicAmountX",
		"Multiplier: -1,",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourcePumpReferencedTarget(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test It Pump",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Untap target creature. It gets +2/+2 until end of turn.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"game.ModifyPT{",
		"Object:         game.TargetPermanentReference(0),",
		"PowerDelta:     game.Fixed(2),",
		"ToughnessDelta: game.Fixed(2),",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceInheritedPowerDamage(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Power Strike",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Target creature you control gets +1/+1 until end of turn. It deals damage equal to its power to target creature you don't control.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"game.Damage{",
		"Kind:       game.DynamicAmountObjectPower,",
		"Object:     game.TargetPermanentReference(0),",
		"Recipient:    game.AnyTargetDamageRecipient(1),",
		"DamageSource: opt.Val(game.TargetPermanentReference(0)),",
	} {
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
		{
			name:       "group power toughness and keyword",
			oracleText: "Creatures you control get +1/+1 and gain trample until end of turn.",
			wants: []string{
				"game.ApplyContinuous",
				"game.BattlefieldGroup",
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

func TestGenerateExecutableCardSourceExileManaValueQualified(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Exile MV",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Exile target permanent with mana value 4 or greater.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		`Constraint: "target permanent with mana value 4 or greater"`,
		"ManaValue: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 4})",
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

func TestGenerateExecutableCardSourceBounceSpell(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		text   string
		wanted []string
	}{
		{
			name: "spell only",
			text: "Return target spell to its owner's hand.",
			wanted: []string{
				`Allow:      game.TargetAllowStackObject`,
				"StackObjectKinds: []game.StackObjectKind{game.StackSpell}",
				"Object: game.TargetObjectReference(0)",
			},
		},
		{
			name: "spell or nonland permanent an opponent controls",
			text: "Return target spell or nonland permanent an opponent controls to its owner's hand.",
			wanted: []string{
				"game.TargetAllowPermanent | game.TargetAllowStackObject",
				"types.Card{types.Land}",
				"game.ControllerOpponent",
				"Object: game.TargetObjectReference(0)",
			},
		},
		{
			name: "spell or creature",
			text: "Return target spell or creature to its owner's hand.",
			wanted: []string{
				"game.TargetAllowPermanent | game.TargetAllowStackObject",
				"types.Card{types.Creature}",
				"Object: game.TargetObjectReference(0)",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{
				Name:       "Test Spell Bounce",
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
			for _, wanted := range test.wanted {
				if !strings.Contains(source, wanted) {
					t.Fatalf("source missing %q:\n%s", wanted, source)
				}
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
			text: "Each opponent loses 1 life for each creature you control.",
			wanteds: []string{
				"game.LoseLife",
				"game.DynamicAmountCountSelector",
				"PlayerGroup: game.OpponentsReference()",
			},
		},
		{
			text: "Each player loses 1 life for each creature you control.",
			wanteds: []string{
				"game.LoseLife",
				"game.DynamicAmountCountSelector",
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

// TestGenerateExecutableCardSourceLifeLostThisWayDrain proves the two-sentence
// drain pattern "Each opponent loses <amount> life. You gain life equal to the
// life lost this way." lowers to a group LoseLife that publishes its total under
// "life-change" followed by a GainLife reading that result, so the controller
// gains exactly the life lost.
func TestGenerateExecutableCardSourceLifeLostThisWayDrain(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		text    string
		wanteds []string
	}{
		{
			name: "fixed",
			text: "Each opponent loses 3 life. You gain life equal to the life lost this way.",
			wanteds: []string{
				"game.LoseLife",
				"game.Fixed(3)",
				"PlayerGroup: game.OpponentsReference()",
				`PublishResult: game.ResultKey("life-change")`,
				"game.GainLife",
				"game.DynamicAmountPreviousEffectResult",
				`ResultKey: game.ResultKey("life-change")`,
				"Player: game.ControllerReference()",
			},
		},
		{
			name: "variableX",
			text: "Each opponent loses X life. You gain life equal to the life lost this way.",
			wanteds: []string{
				"game.LoseLife",
				"game.DynamicAmountX",
				"PlayerGroup: game.OpponentsReference()",
				`PublishResult: game.ResultKey("life-change")`,
				"game.GainLife",
				"game.DynamicAmountPreviousEffectResult",
			},
		},
		{
			name: "dynamicCount",
			text: "Each opponent loses life equal to the number of Vampires you control. You gain life equal to the life lost this way.",
			wanteds: []string{
				"game.LoseLife",
				"game.DynamicAmountCountSelector",
				"PlayerGroup: game.OpponentsReference()",
				`PublishResult: game.ResultKey("life-change")`,
				"game.GainLife",
				"game.DynamicAmountPreviousEffectResult",
			},
		},
		{
			name: "devotion",
			text: "Each opponent loses X life, where X is your devotion to black. You gain life equal to the life lost this way.",
			wanteds: []string{
				"game.LoseLife",
				"game.DynamicAmountDevotion",
				"[]color.Color{color.Black}",
				"PlayerGroup: game.OpponentsReference()",
				`PublishResult: game.ResultKey("life-change")`,
				"game.GainLife",
				"game.DynamicAmountPreviousEffectResult",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{
				Name:       "Test Drain",
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

// TestGenerateExecutableCardSourceLifeLostThisWayFailsClosed proves the drain
// lowerer rejects life-loss clauses whose amount is not faithfully modeled, so
// the "life lost this way" follow-on can never read a wrong total.
func TestGenerateExecutableCardSourceLifeLostThisWayFailsClosed(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		text string
	}{
		{
			// "two times X" is not an exact life-loss amount.
			name: "twiceX",
			text: "Each opponent loses two times X life. You gain life equal to the life lost this way.",
		},
		{
			// A standalone gain with no preceding life loss has nothing to read.
			name: "standalone",
			text: "You gain life equal to the life lost this way.",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{
				Name:       "Test Drain Closed",
				Layout:     "normal",
				TypeLine:   "Sorcery",
				OracleText: test.text,
			}
			_, diagnostics, err := GenerateExecutableCardSource(card, "t")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) == 0 {
				t.Fatalf("expected %s to fail closed, got no diagnostics", test.name)
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

func TestGenerateExecutableCardSourceSelfPowerDamageToItself(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Justice",
		Layout:     "normal",
		ManaCost:   "{R}{W}",
		TypeLine:   "Instant",
		OracleText: "Target creature deals damage to itself equal to its power.",
		Colors:     []string{"R", "W"},
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
		"Primitive: game.Damage",
		"game.DynamicAmountObjectPower",
		"Recipient:    game.AnyTargetDamageRecipient(0)",
		"DamageSource: opt.Val(game.TargetPermanentReference(0))",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceSelfPowerDamageToOther(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Bite",
		Layout:     "normal",
		ManaCost:   "{1}{G}",
		TypeLine:   "Sorcery",
		OracleText: "Target creature you control deals damage equal to its power to target creature you don't control.",
		Colors:     []string{"G"},
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		`Constraint: "target creature you control"`,
		`Constraint: "target creature you don't control"`,
		"game.DynamicAmountObjectPower",
		"Recipient:    game.AnyTargetDamageRecipient(1)",
		"DamageSource: opt.Val(game.TargetPermanentReference(0))",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceSelfPowerDamageToGroupPair(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Ignition",
		Layout:     "normal",
		ManaCost:   "{3}{R}{R}",
		TypeLine:   "Sorcery",
		OracleText: "Target creature you control deals damage equal to its power to each other creature and each opponent.",
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
		`Constraint: "target creature you control"`,
		"game.DynamicAmountObjectPower",
		"game.GroupDamageRecipient(game.BattlefieldGroupExcluding(game.Selection{RequiredTypes: []types.Card{types.Creature}}, game.TargetPermanentReference(0)))",
		"game.PlayerGroupDamageRecipient(game.OpponentsReference())",
		"DamageSource: opt.Val(game.TargetPermanentReference(0))",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceEachOfTwoTargets(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Reprisal",
		Layout:     "normal",
		ManaCost:   "{3}{R}",
		TypeLine:   "Sorcery",
		OracleText: "Test Reprisal deals 2 damage to each of two targets.",
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
		`Constraint: "two targets"`,
		"MaxTargets: 2",
		"Recipient: game.AnyTargetDamageRecipient(0)",
		"Recipient: game.AnyTargetDamageRecipient(1)",
		"Amount:    game.Fixed(2)",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceEachOfUpToTwoTargetCreatures(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Dual",
		Layout:     "normal",
		ManaCost:   "{R}",
		TypeLine:   "Instant",
		OracleText: "Test Dual deals 1 damage to each of up to two target creatures.",
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
		`Constraint: "up to two target creatures"`,
		"MinTargets: 0",
		"MaxTargets: 2",
		"Recipient: game.AnyTargetDamageRecipient(0)",
		"Recipient: game.AnyTargetDamageRecipient(1)",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceMultiTargetUnionDestroy(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		oracleText string
		wanted     []string
	}{
		{
			name:       "up to one artifact or enchantment",
			oracleText: "Destroy up to one target artifact or enchantment.",
			wanted: []string{
				`Constraint: "up to one target artifact or enchantment"`,
				"PermanentTypes: []types.Card{types.Artifact, types.Enchantment}",
				"MinTargets: 0",
				"MaxTargets: 1",
				"Primitive: game.Destroy",
			},
		},
		{
			name:       "up to two creatures or planeswalkers",
			oracleText: "Destroy up to two target creatures or planeswalkers.",
			wanted: []string{
				`Constraint: "up to two target creatures or planeswalkers"`,
				"PermanentTypes: []types.Card{types.Creature, types.Planeswalker}",
				"MinTargets: 0",
				"MaxTargets: 2",
				"Primitive: game.Destroy",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{
				Name:       "Test Sweep",
				Layout:     "normal",
				ManaCost:   "{2}{W}",
				TypeLine:   "Instant",
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
			for _, wanted := range test.wanted {
				if !strings.Contains(source, wanted) {
					t.Fatalf("source missing %q:\n%s", wanted, source)
				}
			}
		})
	}
}

func TestGenerateExecutableCardSourceExcludedTypeManaValueTarget(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		oracleText string
		wanted     []string
	}{
		{
			name:       "nonland permanent mana value 3 or less",
			oracleText: "Destroy target nonland permanent with mana value 3 or less.",
			wanted: []string{
				`Constraint: "target nonland permanent with mana value 3 or less"`,
				"ExcludedTypes: []types.Card{types.Land}",
				"ManaValue:     opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 3})",
				"Primitive: game.Destroy",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{
				Name:       "Test Decay",
				Layout:     "normal",
				ManaCost:   "{B}{G}",
				TypeLine:   "Instant",
				OracleText: test.oracleText,
				Colors:     []string{"B", "G"},
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

func TestGenerateExecutableCardSourceExcludedTypeManaValuePowerFailsClosed(t *testing.T) {
	t.Parallel()
	// Power/toughness exist only on creatures, so a power qualifier on a
	// non-creature noun ("nonland permanent") must fail closed rather than
	// silently drop the qualifier.
	card := &ScryfallCard{
		Name:       "Test Decay",
		Layout:     "normal",
		ManaCost:   "{B}{G}",
		TypeLine:   "Instant",
		OracleText: "Destroy target nonland permanent with power 3 or less.",
		Colors:     []string{"B", "G"},
	}
	_, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) == 0 {
		t.Fatal("expected fail-closed diagnostic, got none")
	}
}
