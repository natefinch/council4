package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SummonUndead is the card definition for Summon Undead.
//
// Type: Sorcery
// Cost: {4}{B}
//
// Oracle text:
//
//	You may mill three cards. Then return a creature card from your graveyard to the battlefield. (To mill a card, put the top card of your library into your graveyard.)
var SummonUndead = newSummonUndead

func newSummonUndead() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Summon Undead",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.MoveTopOfLibrary{
							Amount:      game.Fixed(3),
							Player:      game.ControllerReference(),
							Destination: zone.Graveyard,
						},
						Optional: true,
					},
					{
						Primitive: game.ChooseFromZone{
							Player:     game.ControllerReference(),
							SourceZone: zone.Graveyard,
							Filter:     game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou},
							Quantity:   game.Fixed(1),
							Destination: game.ChooseDestination{
								Zone: zone.Battlefield,
							},
							Prompt: "Choose a card to return to the battlefield",
						},
					},
				},
			}.Ability()),
			OracleText: `
			You may mill three cards. Then return a creature card from your graveyard to the battlefield. (To mill a card, put the top card of your library into your graveyard.)
		`,
		},
	}
}
