package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CharmedClothier is the card definition for Charmed Clothier.
//
// Type: Creature — Faerie Advisor
// Cost: {4}{W}
//
// Oracle text:
//
//	Flying
//	When this creature enters, create a Royal Role token attached to another target creature you control. (If you control another Role on it, put that one into the graveyard. Enchanted creature gets +1/+1 and has ward {1}.)
var CharmedClothier = newCharmedClothier

func newCharmedClothier() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Charmed Clothier",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Faerie, types.Advisor},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
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
								Constraint: "another target creature you control",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerYou, ExcludeSource: true}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount:          game.Fixed(1),
									Source:          game.TokenDef(charmedClothierToken),
									EntryAttachedTo: opt.Val(game.TargetObjectReference(0)),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Flying
			When this creature enters, create a Royal Role token attached to another target creature you control. (If you control another Role on it, put that one into the graveyard. Enchanted creature gets +1/+1 and has ward {1}.)
		`,
		},
	}
}

var charmedClothierToken = newCharmedClothierToken()

func newCharmedClothierToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "Royal Role",
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
							Layer: game.LayerAbility,
							Group: game.AttachedObjectGroup(game.SourcePermanentReference()),
							AddAbilities: []game.Ability{
								new(game.WardStaticAbility(cost.Mana{cost.O(1)})),
							},
						},
						game.ContinuousEffect{
							Layer:          game.LayerPowerToughnessModify,
							Group:          game.AttachedObjectGroup(game.SourcePermanentReference()),
							PowerDelta:     1,
							ToughnessDelta: 1,
						},
					},
				},
			},
			OracleText: `
			Enchant creature
			Enchanted creature gets +1/+1 and has ward {1}.
			(Whenever this creature becomes the target of a spell or ability an opponent controls, counter it unless that player pays {1}.)
		`,
		},
	}
}
