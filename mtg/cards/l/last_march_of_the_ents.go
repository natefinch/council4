package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// LastMarchOfTheEnts is the card definition for Last March of the Ents.
//
// Type: Sorcery
// Cost: {6}{G}{G}
//
// Oracle text:
//
//	This spell can't be countered.
//	Draw cards equal to the greatest toughness among creatures you control, then put any number of creature cards from your hand onto the battlefield.
var LastMarchOfTheEnts = newLastMarchOfTheEnts

func newLastMarchOfTheEnts() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Last March of the Ents",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
				cost.G,
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Sorcery},
			StaticAbilities: []game.StaticAbility{
				game.CantBeCounteredStaticBody,
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Draw{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:       game.DynamicAmountGreatestToughnessInGroup,
								Multiplier: 1,
								Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
							}),
							Player: game.ControllerReference(),
						},
					},
					{
						Primitive: game.ChooseFromZone{
							Player:     game.ControllerReference(),
							SourceZone: zone.Hand,
							Filter:     game.Selection{RequiredTypes: []types.Card{types.Creature}},
							Count:      game.ChooseAnyNumber,
							Destination: game.ChooseDestination{
								Zone: zone.Battlefield,
							},
							Prompt: "Choose a card to put onto the battlefield",
						},
					},
				},
			}.Ability()),
			OracleText: `
			This spell can't be countered.
			Draw cards equal to the greatest toughness among creatures you control, then put any number of creature cards from your hand onto the battlefield.
		`,
		},
	}
}
