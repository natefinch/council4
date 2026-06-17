package cardgen

import (
	"reflect"
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
