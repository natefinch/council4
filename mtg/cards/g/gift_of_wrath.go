package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// GiftOfWrath is the card definition for Gift of Wrath.
//
// Type: Enchantment — Aura
// Cost: {3}{R}
//
// Oracle text:
//
//	Enchant artifact or creature
//	As long as enchanted permanent is a creature, it gets +2/+2 and has menace. (It can't be blocked except by two or more creatures.)
//	When this Aura leaves the battlefield, create a 2/2 red Spirit creature token with menace.
var GiftOfWrath = newGiftOfWrath

func newGiftOfWrath() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Gift of Wrath",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
			}),
			Colors:   []color.Color{color.Red},
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Aura},
			StaticAbilities: []game.StaticAbility{
				game.EnchantStaticAbility(&game.TargetSpec{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "artifact or creature",
					Allow:      game.TargetAllowPermanent,
					Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Creature}}),
				}),
				game.StaticAbility{
					Condition: opt.Val(game.Condition{
						Object:        opt.Val(game.SourceAttachedPermanentReference()),
						ObjectMatches: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:          game.LayerPowerToughnessModify,
							Group:          game.AttachedObjectGroup(game.SourcePermanentReference()),
							PowerDelta:     2,
							ToughnessDelta: 2,
						},
						game.ContinuousEffect{
							Layer: game.LayerAbility,
							Group: game.AttachedObjectGroup(game.SourcePermanentReference()),
							AddKeywords: []game.Keyword{
								game.Menace,
							},
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:         game.EventZoneChanged,
							Source:        game.TriggerSourceSelf,
							MatchFromZone: true,
							FromZone:      zone.Battlefield,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(giftOfWrathToken),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Enchant artifact or creature
			As long as enchanted permanent is a creature, it gets +2/+2 and has menace. (It can't be blocked except by two or more creatures.)
			When this Aura leaves the battlefield, create a 2/2 red Spirit creature token with menace.
		`,
		},
	}
}

var giftOfWrathToken = newGiftOfWrathToken()

func newGiftOfWrathToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Spirit",
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Spirit},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.MenaceStaticBody,
			},
		},
	}
}
