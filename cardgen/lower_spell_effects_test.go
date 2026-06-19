package cardgen

import (
	"reflect"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestLowerSpellDamage(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Bolt",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Test Bolt deals 3 damage to any target.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 {
		t.Fatalf("got %d targets, want 1", len(mode.Targets))
	}
	damage, ok := mode.Sequence[0].Primitive.(game.Damage)
	if !ok {
		t.Fatalf("primitive = %T, want game.Damage", mode.Sequence[0].Primitive)
	}
	if damage.Amount.Value() != 3 {
		t.Fatalf("damage amount = %d, want 3", damage.Amount.Value())
	}
}

func TestLowerSpellDamageQualifiedTarget(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Bolt",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Test Bolt deals 3 damage to target attacking or blocking creature.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if got := mode.Targets[0].Predicate.CombatState; got != game.CombatStateAttackingOrBlocking {
		t.Fatalf("combat state = %v, want attacking or blocking", got)
	}
}

func TestLowerSpellXAmounts(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		cardName   string
		oracleText string
		quantity   func(game.AbilityContent) game.Quantity
	}{
		{
			name:       "damage",
			cardName:   "Test Blaze",
			oracleText: "Test Blaze deals X damage to any target.",
			quantity: func(content game.AbilityContent) game.Quantity {
				primitive, ok := content.Modes[0].Sequence[0].Primitive.(game.Damage)
				if !ok {
					return game.Fixed(0)
				}

				return primitive.Amount
			},
		},
		{
			name:       "draw",
			cardName:   "Test Insight",
			oracleText: "Draw X cards.",
			quantity: func(content game.AbilityContent) game.Quantity {
				primitive, ok := content.Modes[0].Sequence[0].Primitive.(game.Draw)
				if !ok {
					return game.Fixed(0)
				}
				return primitive.Amount
			},
		},
		{
			name:       "life",
			cardName:   "Test Life",
			oracleText: "You gain X life.",
			quantity: func(content game.AbilityContent) game.Quantity {
				primitive, ok := content.Modes[0].Sequence[0].Primitive.(game.GainLife)
				if !ok {
					return game.Fixed(0)
				}
				return primitive.Amount
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       test.cardName,
				Layout:     "normal",
				TypeLine:   "Sorcery",
				OracleText: test.oracleText,
			})
			dynamic := test.quantity(face.SpellAbility.Val).DynamicAmount()
			if !dynamic.Exists || dynamic.Val.Kind != game.DynamicAmountX {
				t.Fatalf("dynamic amount = %+v, want X", dynamic)
			}
		})
	}
}

func TestLowerDynamicEffectAmounts(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		oracleText string
		quantity   func(game.AbilityContent) game.Quantity
		kind       game.DynamicAmountKind
		multiplier int
		cardType   types.Card
		controller game.ControllerRelation
	}{
		{"controlled creatures damage", "Test Swarm deals damage equal to the number of creatures you control to any target.", func(content game.AbilityContent) game.Quantity {
			primitive, ok := content.Modes[0].Sequence[0].Primitive.(game.Damage)
			if !ok {
				return game.Fixed(0)
			}
			return primitive.Amount
		}, game.DynamicAmountCountSelector, 1, types.Creature, game.ControllerYou},
		{"twice battlefield lands damage", "Test Swarm deals damage equal to twice the number of lands on the battlefield to any target.", func(content game.AbilityContent) game.Quantity {
			primitive, ok := content.Modes[0].Sequence[0].Primitive.(game.Damage)
			if !ok {
				return game.Fixed(0)
			}
			return primitive.Amount
		}, game.DynamicAmountCountSelector, 2, types.Land, game.ControllerAny},
		{"life for opponents", "You gain 2 life for each opponent you have.", func(content game.AbilityContent) game.Quantity {
			primitive, ok := content.Modes[0].Sequence[0].Primitive.(game.GainLife)
			if !ok {
				return game.Fixed(0)
			}
			return primitive.Amount
		}, game.DynamicAmountOpponentCount, 2, "", game.ControllerAny},
		{"controller life", "You gain life equal to your life total.", func(content game.AbilityContent) game.Quantity {
			primitive, ok := content.Modes[0].Sequence[0].Primitive.(game.GainLife)
			if !ok {
				return game.Fixed(0)
			}
			return primitive.Amount
		}, game.DynamicAmountControllerLife, 1, "", game.ControllerAny},
		{"draw for controlled lands", "Draw a card for each land you control.", func(content game.AbilityContent) game.Quantity {
			primitive, ok := content.Modes[0].Sequence[0].Primitive.(game.Draw)
			if !ok {
				return game.Fixed(0)
			}
			return primitive.Amount
		}, game.DynamicAmountCountSelector, 1, types.Land, game.ControllerYou},
		{"power for opponents", "Target creature gets +1/+0 for each opponent you have until end of turn.", func(content game.AbilityContent) game.Quantity {
			primitive, ok := content.Modes[0].Sequence[0].Primitive.(game.ModifyPT)
			if !ok {
				return game.Fixed(0)
			}
			return primitive.PowerDelta
		}, game.DynamicAmountOpponentCount, 1, "", game.ControllerAny},
		{"power after duration", "Target creature gets +1/+0 until end of turn for each opponent you have.", func(content game.AbilityContent) game.Quantity {
			primitive, ok := content.Modes[0].Sequence[0].Primitive.(game.ModifyPT)
			if !ok {
				return game.Fixed(0)
			}
			return primitive.PowerDelta
		}, game.DynamicAmountOpponentCount, 1, "", game.ControllerAny},
		{"counters for controlled lands", "Put X +1/+1 counters on target creature, where X is the number of lands you control.", func(content game.AbilityContent) game.Quantity {
			primitive, ok := content.Modes[0].Sequence[0].Primitive.(game.AddCounter)
			if !ok {
				return game.Fixed(0)
			}
			return primitive.Amount
		}, game.DynamicAmountCountSelector, 1, types.Land, game.ControllerYou},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Swarm",
				Layout:     "normal",
				TypeLine:   "Sorcery",
				OracleText: test.oracleText,
			})
			dynamic := test.quantity(face.SpellAbility.Val).DynamicAmount()
			if !dynamic.Exists ||
				dynamic.Val.Kind != test.kind ||
				dynamic.Val.Multiplier != test.multiplier {
				t.Fatalf("dynamic amount = %+v", dynamic)
			}
			if test.cardType != "" {
				selection := dynamic.Val.Group.Selection()
				if len(selection.RequiredTypes) != 1 ||
					selection.RequiredTypes[0] != test.cardType ||
					selection.Controller != test.controller {
					t.Fatalf("selection = %+v", selection)
				}
			}
		})
	}
}

// TestLowerSubtypeCountDamage verifies that single-target damage spells whose
// amount counts a subtype group ("equal to the number of <subtype> you
// control", or the "where X is the number of ..." form) lower to a
// DynamicAmountCountSelector carrying that subtype. The runtime count selector
// already supports SubtypesAny, so these reuse the existing amount kind.
func TestLowerSubtypeCountDamage(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		oracleText string
		subtype    types.Sub
		multiplier int
	}{
		{
			name:       "equal to subtype",
			oracleText: "Test Swarm deals damage to target creature equal to the number of Goblins you control.",
			subtype:    types.Goblin,
			multiplier: 1,
		},
		{
			name:       "equal to land subtype",
			oracleText: "Test Swarm deals damage to any target equal to the number of Mountains you control.",
			subtype:    types.Mountain,
			multiplier: 1,
		},
		{
			name:       "where X is subtype",
			oracleText: "Test Swarm deals X damage to any target, where X is the number of Wizards you control.",
			subtype:    types.Wizard,
			multiplier: 1,
		},
		{
			name:       "leading count clause",
			oracleText: "Test Swarm deals damage equal to the number of Swamps you control to any target.",
			subtype:    types.Swamp,
			multiplier: 1,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Swarm",
				Layout:     "normal",
				TypeLine:   "Sorcery",
				OracleText: test.oracleText,
			})
			damage, ok := face.SpellAbility.Val.Modes[0].Sequence[0].Primitive.(game.Damage)
			if !ok {
				t.Fatalf("primitive = %T, want game.Damage", face.SpellAbility.Val.Modes[0].Sequence[0].Primitive)
			}
			dynamic := damage.Amount.DynamicAmount()
			if !dynamic.Exists ||
				dynamic.Val.Kind != game.DynamicAmountCountSelector ||
				dynamic.Val.Multiplier != test.multiplier {
				t.Fatalf("dynamic amount = %+v", dynamic)
			}
			selection := dynamic.Val.Group.Selection()
			if len(selection.SubtypesAny) != 1 ||
				selection.SubtypesAny[0] != test.subtype ||
				selection.Controller != game.ControllerYou {
				t.Fatalf("selection = %+v, want subtype %v controlled by you", selection, test.subtype)
			}
		})
	}
}

// TestLowerSubtypeCountDamageFailsClosed verifies that subtype-count damage
// wordings the backend cannot represent exactly stay rejected: a singular head
// after "the number of" is ungrammatical, and a count form with no representable
// group (greatest mana value) is unsupported.
func TestLowerSubtypeCountDamageFailsClosed(t *testing.T) {
	t.Parallel()
	for _, oracleText := range []string{
		"Test Swarm deals damage to any target equal to the number of Goblin you control.",
		"Test Swarm deals damage to any target equal to the greatest mana value among permanents you control.",
	} {
		t.Run(oracleText, func(t *testing.T) {
			t.Parallel()
			faces, diagnostics := lowerExecutableFaces(&ScryfallCard{
				Name:       "Test Swarm",
				Layout:     "normal",
				TypeLine:   "Sorcery",
				OracleText: oracleText,
			})
			if len(faces) != 1 || faces[0].SpellAbility.Exists {
				t.Fatalf("faces = %#v, want face with no lowered spell ability", faces)
			}
			if len(diagnostics) == 0 {
				t.Fatal("expected unsupported diagnostic")
			}
		})
	}
}

func TestLowerSpellDestroyQualifiedTarget(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Destroy",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Destroy target tapped creature an opponent controls.",
	})
	target := face.SpellAbility.Val.Modes[0].Targets[0]
	if target.Predicate.Tapped != game.TriTrue ||
		target.Predicate.Controller != game.ControllerOpponent {
		t.Fatalf("predicate = %+v, want tapped creature an opponent controls", target.Predicate)
	}
}

func TestLowerSpellDestroyPowerToughnessTarget(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		oracleText string
		want       game.TargetPredicate
	}{
		{
			name:       "power at most",
			oracleText: "Destroy target creature with power 2 or less.",
			want: game.TargetPredicate{
				PermanentTypes: []types.Card{types.Creature},
				Power:          opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 2}),
			},
		},
		{
			name:       "toughness at least",
			oracleText: "Destroy target creature with toughness 4 or greater.",
			want: game.TargetPredicate{
				PermanentTypes: []types.Card{types.Creature},
				Toughness:      opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 4}),
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Destroy " + test.name,
				Layout:     "normal",
				TypeLine:   "Instant",
				OracleText: test.oracleText,
			})
			target := face.SpellAbility.Val.Modes[0].Targets[0]
			if !reflect.DeepEqual(target.Predicate, test.want) {
				t.Fatalf("predicate = %+v, want %+v", target.Predicate, test.want)
			}
		})
	}
}

func TestLowerSpellDestroyTypeUnionManaValueTarget(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		oracleText string
		want       game.TargetPredicate
	}{
		{
			name:       "creature or planeswalker",
			oracleText: "Destroy target creature or planeswalker with mana value 3 or less.",
			want: game.TargetPredicate{
				PermanentTypes: []types.Card{types.Creature, types.Planeswalker},
				ManaValue:      opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 3}),
			},
		},
		{
			name:       "artifact or enchantment",
			oracleText: "Destroy target artifact or enchantment with mana value 4 or less.",
			want: game.TargetPredicate{
				PermanentTypes: []types.Card{types.Artifact, types.Enchantment},
				ManaValue:      opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 4}),
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Destroy Union " + test.name,
				Layout:     "normal",
				TypeLine:   "Instant",
				OracleText: test.oracleText,
			})
			target := face.SpellAbility.Val.Modes[0].Targets[0]
			if !reflect.DeepEqual(target.Predicate, test.want) {
				t.Fatalf("predicate = %+v, want %+v", target.Predicate, test.want)
			}
		})
	}
}

func TestLowerSpellDestroyRegenerationRider(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		oracleText string
		mass       bool
	}{
		{
			name:       "single target",
			oracleText: "Destroy target creature. It can't be regenerated.",
		},
		{
			name:       "excluded color target",
			oracleText: "Destroy target nonblack creature. It can't be regenerated.",
		},
		{
			name:       "mass",
			oracleText: "Destroy all creatures. They can't be regenerated.",
			mass:       true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Regen " + test.name,
				Layout:     "normal",
				TypeLine:   "Sorcery",
				OracleText: test.oracleText,
			})
			destroy, ok := face.SpellAbility.Val.Modes[0].Sequence[0].Primitive.(game.Destroy)
			if !ok {
				t.Fatalf("primitive = %T, want game.Destroy", face.SpellAbility.Val.Modes[0].Sequence[0].Primitive)
			}
			if !destroy.PreventRegeneration {
				t.Fatal("PreventRegeneration = false, want true")
			}
			if test.mass && destroy.Group.Domain() != game.GroupDomainBattlefield {
				t.Fatalf("group domain = %v, want battlefield", destroy.Group.Domain())
			}
		})
	}
}

func TestLowerSpellDestroyWithoutRiderKeepsRegeneration(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Plain Destroy",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Destroy target creature.",
	})
	destroy, ok := face.SpellAbility.Val.Modes[0].Sequence[0].Primitive.(game.Destroy)
	if !ok {
		t.Fatalf("primitive = %T, want game.Destroy", face.SpellAbility.Val.Modes[0].Sequence[0].Primitive)
	}
	if destroy.PreventRegeneration {
		t.Fatal("PreventRegeneration = true, want false for a destroy without a rider")
	}
}

func TestLowerMassDestroyAndExile(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		oracleText string
		selection  game.Selection
		exile      bool
	}{
		{
			name:       "land",
			oracleText: "Destroy all lands.",
			selection:  game.Selection{RequiredTypes: []types.Card{types.Land}},
		},
		{
			name:       "nonland permanent",
			oracleText: "Destroy all nonland permanents.",
			selection:  game.Selection{ExcludedTypes: []types.Card{types.Land}},
		},
		{
			name:       "not controlled by you",
			oracleText: "Destroy all creatures you don't control.",
			selection: game.Selection{
				RequiredTypes: []types.Card{types.Creature},
				Controller:    game.ControllerOpponent,
			},
		},
		{
			name:       "excluded color",
			oracleText: "Destroy all nonwhite creatures.",
			selection: game.Selection{
				RequiredTypes:  []types.Card{types.Creature},
				ExcludedColors: []color.Color{color.White},
			},
		},
		{
			name:       "keyword",
			oracleText: "Destroy all creatures with flying.",
			selection: game.Selection{
				RequiredTypes: []types.Card{types.Creature},
				Keyword:       game.Flying,
			},
		},
		{
			name:       "mana value",
			oracleText: "Destroy all creatures with mana value 3 or less.",
			selection: game.Selection{
				RequiredTypes: []types.Card{types.Creature},
				ManaValue: opt.Val(compare.Int{
					Op:    compare.LessOrEqual,
					Value: 3,
				}),
			},
		},
		{
			name:       "toughness",
			oracleText: "Destroy all creatures with toughness 4 or greater.",
			selection: game.Selection{
				RequiredTypes: []types.Card{types.Creature},
				Toughness: opt.Val(compare.Int{
					Op:    compare.GreaterOrEqual,
					Value: 4,
				}),
			},
		},
		{
			name:       "other",
			oracleText: "Destroy all other creatures.",
			selection: game.Selection{
				RequiredTypes: []types.Card{types.Creature},
				ExcludeSource: true,
			},
		},
		{
			name:       "exile",
			oracleText: "Exile all creatures.",
			selection:  game.Selection{RequiredTypes: []types.Card{types.Creature}},
			exile:      true,
		},
		{
			name:       "subtype",
			oracleText: "Destroy all Islands.",
			selection:  game.Selection{SubtypesAny: []types.Sub{types.Island}},
		},
		{
			name:       "subtype with card type",
			oracleText: "Destroy all Dragon creatures.",
			selection: game.Selection{
				RequiredTypes: []types.Card{types.Creature},
				SubtypesAny:   []types.Sub{types.Dragon},
			},
		},
		{
			name:       "untapped",
			oracleText: "Destroy all untapped creatures.",
			selection: game.Selection{
				RequiredTypes: []types.Card{types.Creature},
				Tapped:        game.TriFalse,
			},
		},
		{
			name:       "noncreature mana value",
			oracleText: "Destroy all nonland permanents with mana value 1 or less.",
			selection: game.Selection{
				ExcludedTypes: []types.Card{types.Land},
				ManaValue: opt.Val(compare.Int{
					Op:    compare.LessOrEqual,
					Value: 1,
				}),
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Mass Effect",
				Layout:     "normal",
				TypeLine:   "Sorcery",
				OracleText: test.oracleText,
			})
			primitive := face.SpellAbility.Val.Modes[0].Sequence[0].Primitive
			var group game.GroupReference
			switch primitive := primitive.(type) {
			case game.Destroy:
				if test.exile {
					t.Fatalf("primitive = %T, want game.Exile", primitive)
				}
				group = primitive.Group
			case game.Exile:
				if !test.exile {
					t.Fatalf("primitive = %T, want game.Destroy", primitive)
				}
				group = primitive.Group
			default:
				t.Fatalf("primitive = %T, want mass destroy or exile", primitive)
			}
			if group.Domain() != game.GroupDomainBattlefield {
				t.Fatalf("group domain = %v, want battlefield", group.Domain())
			}
			if selection := group.Selection(); !reflect.DeepEqual(selection, test.selection) {
				t.Fatalf("selection = %#v, want %#v", selection, test.selection)
			}
		})
	}
}

func TestLowerSpellReturnQualifiedTarget(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Return",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Return target creature you control to its owner's hand.",
	})
	target := face.SpellAbility.Val.Modes[0].Targets[0]
	if target.Predicate.Controller != game.ControllerYou {
		t.Fatalf("controller = %v, want ControllerYou", target.Predicate.Controller)
	}
}

func TestLowerSpellModifyPTQualifiedTarget(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Growth",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Target untapped creature you control gets +2/+2 until end of turn.",
	})
	target := face.SpellAbility.Val.Modes[0].Targets[0]
	if target.Predicate.Tapped != game.TriFalse ||
		target.Predicate.Controller != game.ControllerYou {
		t.Fatalf("predicate = %+v, want untapped creature you control", target.Predicate)
	}
}

func TestLowerTemporaryGroupModifyPTSpell(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Guidance",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Creatures you control get +1/+1 until end of turn.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	primitive, ok := mode.Sequence[0].Primitive.(game.ApplyContinuous)
	if !ok {
		t.Fatalf("primitive = %T, want game.ApplyContinuous", mode.Sequence[0].Primitive)
	}
	if primitive.Object.Exists || primitive.Duration != game.DurationUntilEndOfTurn {
		t.Fatalf("primitive = %+v, want group effect until end of turn", primitive)
	}
	if len(primitive.ContinuousEffects) != 1 {
		t.Fatalf("continuous effects = %d, want 1", len(primitive.ContinuousEffects))
	}
	effect := primitive.ContinuousEffects[0]
	selection := effect.Group.Selection()
	if effect.Layer != game.LayerPowerToughnessModify ||
		effect.PowerDelta != 1 ||
		effect.ToughnessDelta != 1 ||
		effect.Group.Domain() != game.GroupDomainBattlefield ||
		selection.Controller != game.ControllerYou ||
		len(selection.RequiredTypes) != 1 ||
		selection.RequiredTypes[0] != types.Creature {
		t.Fatalf("continuous effect = %+v, want controlled creatures +1/+1", effect)
	}
}

func TestLowerCardNameSelfSubject(t *testing.T) {
	t.Parallel()
	cases := []struct {
		oracleText string
		primitive  string
	}{
		{"{R}: Tester gets +1/+0 until end of turn.", "game.ModifyPT"},
		{"Whenever you gain life, put a +1/+1 counter on Tester.", "game.AddCounter"},
	}
	for _, tc := range cases {
		source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
			Name:       "Tester",
			Layout:     "normal",
			TypeLine:   "Creature — Human",
			Power:      new("2"),
			Toughness:  new("2"),
			OracleText: tc.oracleText,
		}, "t")
		if err != nil {
			t.Fatal(err)
		}
		if len(diagnostics) != 0 {
			t.Fatalf("card-name self %q unexpectedly failed: %v", tc.oracleText, diagnostics)
		}
		if !strings.Contains(source, tc.primitive) ||
			!strings.Contains(source, "game.SourcePermanentReference()") {
			t.Fatalf("card-name self %q did not lower to source %s:\n%s", tc.oracleText, tc.primitive, source)
		}
	}
}

func TestLowerTemporarySelfKeywordAbility(t *testing.T) {
	t.Parallel()
	for _, oracleText := range []string{
		"{W}: This creature gains flying until end of turn.",
		"{G}: This creature gains trample and haste until end of turn.",
	} {
		source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
			Name:       "Test Skyfish",
			Layout:     "normal",
			TypeLine:   "Creature — Fish",
			Power:      new("2"),
			Toughness:  new("2"),
			OracleText: oracleText,
		}, "t")
		if err != nil {
			t.Fatal(err)
		}
		if len(diagnostics) != 0 {
			t.Fatalf("self-gain %q unexpectedly failed: %v", oracleText, diagnostics)
		}
		if !strings.Contains(source, "game.ApplyContinuous") ||
			!strings.Contains(source, "game.SourceCardPermanentReference()") {
			t.Fatalf("self-gain %q did not lower to a source ApplyContinuous:\n%s", oracleText, source)
		}
	}
}

func TestLowerTemporaryTargetKeywordSpell(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Flight",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Target creature gains flying until end of turn.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	checkKeywordGrantPrimitive(t, mode, 0, game.Flying)
}

func TestLowerTemporaryTargetPTKeywordSpell(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Growth",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Target creature gets +2/+2 and gains trample until end of turn.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	primitive, ok := mode.Sequence[0].Primitive.(game.ApplyContinuous)
	if !ok {
		t.Fatalf("primitive = %T, want game.ApplyContinuous", mode.Sequence[0].Primitive)
	}
	if len(primitive.ContinuousEffects) != 2 {
		t.Fatalf("continuous effects = %d, want 2", len(primitive.ContinuousEffects))
	}
	pt := primitive.ContinuousEffects[0]
	keyword := primitive.ContinuousEffects[1]
	if pt.Layer != game.LayerPowerToughnessModify || pt.PowerDelta != 2 || pt.ToughnessDelta != 2 {
		t.Fatalf("power/toughness effect = %+v", pt)
	}
	if keyword.Layer != game.LayerAbility ||
		len(keyword.AddKeywords) != 1 ||
		keyword.AddKeywords[0] != game.Trample {
		t.Fatalf("keyword effect = %+v", keyword)
	}
}

func TestLowerSpellDamagePlayerOrPlaneswalker(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Lava Spike",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Lava Spike deals 3 damage to target player or planeswalker.",
	})
	target := face.SpellAbility.Val.Modes[0].Targets[0]
	if target.Allow != game.TargetAllowPlayer|game.TargetAllowPermanent {
		t.Fatalf("allow = %v, want player|permanent", target.Allow)
	}
	if !reflect.DeepEqual(target.Predicate.PermanentTypes, []types.Card{types.Planeswalker}) {
		t.Fatalf("permanent types = %v, want [planeswalker]", target.Predicate.PermanentTypes)
	}
	if target.Predicate.Player != game.PlayerAny {
		t.Fatalf("player = %v, want any player", target.Predicate.Player)
	}
}

func TestLowerSpellDamageOpponentOrPlaneswalker(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Searing Flesh",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Searing Flesh deals 7 damage to target opponent or planeswalker.",
	})
	target := face.SpellAbility.Val.Modes[0].Targets[0]
	if target.Allow != game.TargetAllowPlayer|game.TargetAllowPermanent {
		t.Fatalf("allow = %v, want player|permanent", target.Allow)
	}
	if !reflect.DeepEqual(target.Predicate.PermanentTypes, []types.Card{types.Planeswalker}) {
		t.Fatalf("permanent types = %v, want [planeswalker]", target.Predicate.PermanentTypes)
	}
	if target.Predicate.Player != game.PlayerOpponent {
		t.Fatalf("player = %v, want opponent", target.Predicate.Player)
	}
}

func TestLowerSpellDamageKeywordTarget(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Leaf Arrow",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Leaf Arrow deals 3 damage to target creature with flying.",
	})
	target := face.SpellAbility.Val.Modes[0].Targets[0]
	if target.Predicate.Keyword != game.Flying {
		t.Fatalf("keyword = %v, want flying", target.Predicate.Keyword)
	}
}

func TestLowerSpellDamageExcludedKeywordTarget(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Roast",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Roast deals 5 damage to target creature without flying.",
	})
	target := face.SpellAbility.Val.Modes[0].Targets[0]
	if target.Predicate.ExcludedKeyword != game.Flying {
		t.Fatalf("excluded keyword = %v, want flying", target.Predicate.ExcludedKeyword)
	}
	if target.Predicate.Keyword != game.KeywordNone {
		t.Fatalf("keyword = %v, want none", target.Predicate.Keyword)
	}
}

func TestLowerSpellDamageMultiColorTarget(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Rending Volley",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Rending Volley deals 4 damage to target white or blue creature.",
	})
	target := face.SpellAbility.Val.Modes[0].Targets[0]
	if !reflect.DeepEqual(target.Predicate.Colors, []color.Color{color.White, color.Blue}) {
		t.Fatalf("colors = %v, want [white blue]", target.Predicate.Colors)
	}
}

func TestLowerSpellDamageGroupKeyword(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Gale Force",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Gale Force deals 5 damage to each creature with flying.",
	})
	damage, ok := face.SpellAbility.Val.Modes[0].Sequence[0].Primitive.(game.Damage)
	if !ok {
		t.Fatalf("primitive = %T, want game.Damage", face.SpellAbility.Val.Modes[0].Sequence[0].Primitive)
	}
	want := game.GroupDamageRecipient(game.BattlefieldGroup(game.Selection{
		RequiredTypes: []types.Card{types.Creature},
		Keyword:       game.Flying,
	}))
	if !reflect.DeepEqual(damage.Recipient, want) {
		t.Fatalf("recipient = %#v, want %#v", damage.Recipient, want)
	}
	if damage.Amount.Value() != 5 {
		t.Fatalf("amount = %d, want 5", damage.Amount.Value())
	}
}

func TestLowerSpellDamageGroupExcludedKeyword(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Seismic Shudder",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Seismic Shudder deals 1 damage to each creature without flying.",
	})
	damage, ok := face.SpellAbility.Val.Modes[0].Sequence[0].Primitive.(game.Damage)
	if !ok {
		t.Fatalf("primitive = %T, want game.Damage", face.SpellAbility.Val.Modes[0].Sequence[0].Primitive)
	}
	want := game.GroupDamageRecipient(game.BattlefieldGroup(game.Selection{
		RequiredTypes:   []types.Card{types.Creature},
		ExcludedKeyword: game.Flying,
	}))
	if !reflect.DeepEqual(damage.Recipient, want) {
		t.Fatalf("recipient = %#v, want %#v", damage.Recipient, want)
	}
	if damage.Amount.Value() != 1 {
		t.Fatalf("amount = %d, want 1", damage.Amount.Value())
	}
}

func TestLowerSpellDamageUnsupportedGroupKeywordFailsClosed(t *testing.T) {
	t.Parallel()
	// "shadow" is not a runtime-modelable selector keyword, so the group
	// damage spell stays fail-closed at lowering.
	faces, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Shadowflyer",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Shadowflyer deals 1 damage to each creature with shadow.",
	})
	if len(faces) > 0 && faces[0].SpellAbility.Exists {
		t.Fatal("expected fail-closed, got lowered spell ability")
	}
	if len(diagnostics) == 0 {
		t.Fatal("expected an unsupported diagnostic")
	}
}

// modifyPTSlots returns the ModifyPT primitive at each sequence index of the
// spell's first mode, asserting the mode targets one spec with the expected
// cardinality range.
func modifyPTSlots(t *testing.T, oracleText, typeLine string, wantMin, wantMax int) (game.TargetSpec, []game.ModifyPT) {
	t.Helper()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Pump",
		Layout:     "normal",
		TypeLine:   typeLine,
		OracleText: oracleText,
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 {
		t.Fatalf("targets = %d, want 1", len(mode.Targets))
	}
	spec := mode.Targets[0]
	if spec.MinTargets != wantMin || spec.MaxTargets != wantMax {
		t.Fatalf("cardinality = %d..%d, want %d..%d", spec.MinTargets, spec.MaxTargets, wantMin, wantMax)
	}
	if len(mode.Sequence) != wantMax {
		t.Fatalf("sequence = %d instructions, want %d", len(mode.Sequence), wantMax)
	}
	mods := make([]game.ModifyPT, 0, len(mode.Sequence))
	for i := range mode.Sequence {
		mod, ok := mode.Sequence[i].Primitive.(game.ModifyPT)
		if !ok {
			t.Fatalf("instruction %d primitive = %T, want game.ModifyPT", i, mode.Sequence[i].Primitive)
		}
		if mod.Object != game.TargetPermanentReference(i) {
			t.Fatalf("instruction %d object = %+v, want target reference %d", i, mod.Object, i)
		}
		if mod.Duration != game.DurationUntilEndOfTurn {
			t.Fatalf("instruction %d duration = %v, want until end of turn", i, mod.Duration)
		}
		mods = append(mods, mod)
	}
	return spec, mods
}

func TestLowerPluralModifyPTEachGet(t *testing.T) {
	t.Parallel()
	spec, mods := modifyPTSlots(t, "Two target creatures each get -1/-1 until end of turn.", "Instant", 2, 2)
	if len(spec.Predicate.PermanentTypes) != 1 || spec.Predicate.PermanentTypes[0] != types.Creature {
		t.Fatalf("predicate = %+v, want creature", spec.Predicate)
	}
	for i, mod := range mods {
		if mod.PowerDelta != game.Fixed(-1) || mod.ToughnessDelta != game.Fixed(-1) {
			t.Fatalf("slot %d delta = %v/%v, want -1/-1", i, mod.PowerDelta, mod.ToughnessDelta)
		}
	}
}

func TestLowerUpToTwoModifyPTEachGet(t *testing.T) {
	t.Parallel()
	spec, mods := modifyPTSlots(t, "Up to two target creatures each get +2/+2 until end of turn.", "Instant", 0, 2)
	if len(spec.Predicate.PermanentTypes) != 1 || spec.Predicate.PermanentTypes[0] != types.Creature {
		t.Fatalf("predicate = %+v, want creature", spec.Predicate)
	}
	for i, mod := range mods {
		if mod.PowerDelta != game.Fixed(2) || mod.ToughnessDelta != game.Fixed(2) {
			t.Fatalf("slot %d delta = %v/%v, want +2/+2", i, mod.PowerDelta, mod.ToughnessDelta)
		}
	}
}

func TestLowerUpToTwoControlledModifyPTEachGet(t *testing.T) {
	t.Parallel()
	spec, _ := modifyPTSlots(t, "Up to two target creatures you control each get +1/+0 until end of turn.", "Instant", 0, 2)
	if spec.Predicate.Controller != game.ControllerYou {
		t.Fatalf("controller = %v, want you control", spec.Predicate.Controller)
	}
}

func TestLowerUpToOneModifyPT(t *testing.T) {
	t.Parallel()
	spec, mods := modifyPTSlots(t, "Up to one target creature gets -3/-3 until end of turn.", "Instant", 0, 1)
	if len(spec.Predicate.PermanentTypes) != 1 || spec.Predicate.PermanentTypes[0] != types.Creature {
		t.Fatalf("predicate = %+v, want creature", spec.Predicate)
	}
	if mods[0].PowerDelta != game.Fixed(-3) || mods[0].ToughnessDelta != game.Fixed(-3) {
		t.Fatalf("delta = %v/%v, want -3/-3", mods[0].PowerDelta, mods[0].ToughnessDelta)
	}
}

func TestLowerTypedSubtypeModifyPT(t *testing.T) {
	t.Parallel()
	spec, mods := modifyPTSlots(t, "Target Human you control gets +2/+2 until end of turn.", "Instant", 1, 1)
	if len(spec.Predicate.Subtypes) != 1 || spec.Predicate.Subtypes[0] != types.Sub("Human") {
		t.Fatalf("subtypes = %+v, want Human", spec.Predicate.Subtypes)
	}
	if spec.Predicate.Controller != game.ControllerYou {
		t.Fatalf("controller = %v, want you control", spec.Predicate.Controller)
	}
	if mods[0].PowerDelta != game.Fixed(2) || mods[0].ToughnessDelta != game.Fixed(2) {
		t.Fatalf("delta = %v/%v, want +2/+2", mods[0].PowerDelta, mods[0].ToughnessDelta)
	}
}

func TestLowerNonCreaturePumpTargetFailsClosed(t *testing.T) {
	t.Parallel()
	_, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Test Land Pump",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Two target lands each get +1/+1 until end of turn.",
	})
	if len(diagnostics) == 0 {
		t.Fatal("expected an unsupported diagnostic for a non-creature pump target")
	}
}
