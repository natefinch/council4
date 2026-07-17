package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// DestinySpinner is the card definition for Destiny Spinner.
//
// Type: Enchantment Creature — Human
// Cost: {1}{G}
//
// Oracle text:
//
//	Creature and enchantment spells you control can't be countered.
//	{3}{G}: Target land you control becomes an X/X Elemental creature with trample and haste until end of turn, where X is the number of enchantments you control. It's still a land.
var DestinySpinner = newDestinySpinner

func newDestinySpinner() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Destiny Spinner",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Enchantment, types.Creature},
			Subtypes:  []types.Sub{types.Human},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:               game.RuleEffectCantBeCountered,
							AffectedController: game.ControllerYou,
							SpellTypes:         []types.Card{types.Creature, types.Enchantment},
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:           "{3}{G}: Target land you control becomes an X/X Elemental creature with trample and haste until end of turn, where X is the number of enchantments you control. It's still a land.",
					ManaCost:       opt.Val(cost.Mana{cost.O(3), cost.G}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target land you control",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Land}, Controller: game.ControllerYou}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									Object: opt.Val(game.TargetPermanentReference(0)),
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer:       game.LayerType,
											AddTypes:    []types.Card{types.Creature},
											AddSubtypes: []types.Sub{types.Elemental},
										},
										game.ContinuousEffect{
											Layer: game.LayerAbility,
											AddKeywords: []game.Keyword{
												game.Trample,
												game.Haste,
											},
										},
										game.ContinuousEffect{
											Layer: game.LayerPowerToughnessSet,
											SetPowerDynamic: opt.Val(game.DynamicAmount{
												Kind:       game.DynamicAmountCountSelector,
												Multiplier: 1,
												Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Enchantment}, Controller: game.ControllerYou}),
											}),
											SetToughnessDynamic: opt.Val(game.DynamicAmount{
												Kind:       game.DynamicAmountCountSelector,
												Multiplier: 1,
												Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Enchantment}, Controller: game.ControllerYou}),
											}),
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Creature and enchantment spells you control can't be countered.
			{3}{G}: Target land you control becomes an X/X Elemental creature with trample and haste until end of turn, where X is the number of enchantments you control. It's still a land.
		`,
		},
	}
}
