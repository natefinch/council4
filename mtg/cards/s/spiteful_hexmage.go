package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SpitefulHexmage is the card definition for Spiteful Hexmage.
//
// Type: Creature — Human Warlock
// Cost: {B}
//
// Oracle text:
//
//	When this creature enters, create a Cursed Role token attached to target creature you control. (If you control another Role on it, put that one into the graveyard. Enchanted creature is 1/1.)
var SpitefulHexmage = newSpitefulHexmage

func newSpitefulHexmage() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Spiteful Hexmage",
			ManaCost: opt.Val(cost.Mana{
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Warlock},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 2}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentEnteredBattlefield,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target creature you control",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerYou}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount:          game.Fixed(1),
									Source:          game.TokenDef(spitefulHexmageToken),
									EntryAttachedTo: opt.Val(game.TargetObjectReference(0)),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			When this creature enters, create a Cursed Role token attached to target creature you control. (If you control another Role on it, put that one into the graveyard. Enchanted creature is 1/1.)
		`,
		},
	}
}

var spitefulHexmageToken = newSpitefulHexmageToken()

func newSpitefulHexmageToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "Cursed Role",
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Aura, types.Role},
			StaticAbilities: []game.StaticAbility{
				game.EnchantStaticAbility(&game.TargetSpec{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "creature",
					Allow:      game.TargetAllowPermanent,
					Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
				}),
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:        game.LayerPowerToughnessSet,
							Group:        game.AttachedObjectGroup(game.SourcePermanentReference()),
							SetPower:     opt.Val(game.PT{Value: 1}),
							SetToughness: opt.Val(game.PT{Value: 1}),
						},
					},
				},
			},
			OracleText: `
			Enchant creature
			Enchanted creature has base power and toughness 1/1.
		`,
		},
	}
}
