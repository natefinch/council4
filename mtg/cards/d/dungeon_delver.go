package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DungeonDelver is the card definition for Dungeon Delver.
//
// Type: Legendary Enchantment — Background
// Cost: {1}{U}
//
// Oracle text:
//
//	Commander creatures you own have "Room abilities of dungeons you own trigger an additional time."
var DungeonDelver = newDungeonDelver

func newDungeonDelver() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Dungeon Delver",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors:     []color.Color{color.Blue},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Enchantment},
			Subtypes:   []types.Sub{types.Background},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Commander creatures you own have \"Room abilities of dungeons you own trigger an additional time.\"",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration: game.DurationPermanent,
							Group:    game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Owner: game.OwnerYou, MatchCommander: true}),
							Layer:    game.LayerAbility,
							AddAbilities: []game.Ability{
								&game.StaticAbility{
									Text: "Room abilities of dungeons you own trigger an additional time.",
									RuleEffects: []game.RuleEffect{
										game.RuleEffect{
											Kind:     game.RuleEffectAdditionalTriggerForRoomAbility,
											Duration: game.DurationPermanent,
										},
									},
								},
							},
						},
					},
				},
			},
			OracleText: `
			Commander creatures you own have "Room abilities of dungeons you own trigger an additional time."
		`,
		},
	}
}
