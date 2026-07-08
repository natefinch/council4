package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// PrismaticUndercurrents is the card definition for Prismatic Undercurrents.
//
// Type: Enchantment
// Cost: {3}{G}
//
// Oracle text:
//
//	Vivid — When this enchantment enters, search your library for up to X basic land cards, where X is the number of colors among permanents you control. Reveal those cards, put them into your hand, then shuffle.
//	You may play an additional land on each of your turns.
var PrismaticUndercurrents = newPrismaticUndercurrents

func newPrismaticUndercurrents() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Prismatic Undercurrents",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Enchantment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:                game.RuleEffectAdditionalLandPlays,
							AffectedPlayer:      game.PlayerYou,
							AdditionalLandPlays: 1,
						},
					},
				},
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
						Sequence: []game.Instruction{
							{
								Primitive: game.Search{
									Player: game.ControllerReference(),
									Spec: game.SearchSpec{
										SourceZone:  zone.Library,
										Destination: zone.Hand,
										Filter:      game.Selection{RequiredTypes: []types.Card{types.Land}, Supertypes: []types.Super{types.Basic}},
										Reveal:      true,
									},
									Amount: game.Dynamic(game.DynamicAmount{
										Kind:       game.DynamicAmountColorCountInGroup,
										Multiplier: 1,
										Group:      game.BattlefieldGroup(game.Selection{Controller: game.ControllerYou}),
									}),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Vivid — When this enchantment enters, search your library for up to X basic land cards, where X is the number of colors among permanents you control. Reveal those cards, put them into your hand, then shuffle.
			You may play an additional land on each of your turns.
		`,
		},
	}
}
