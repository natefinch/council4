package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// TriumphantReckoning is the card definition for Triumphant Reckoning.
//
// Type: Sorcery
// Cost: {6}{W}{W}{W}
//
// Oracle text:
//
//	Return all artifact, enchantment, and planeswalker cards from your graveyard to the battlefield.
var TriumphantReckoning = newTriumphantReckoning

func newTriumphantReckoning() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Triumphant Reckoning",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
				cost.W,
				cost.W,
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.MassReturnFromGraveyard{
							Player:      game.ControllerReference(),
							Selection:   game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Enchantment, types.Planeswalker}, Controller: game.ControllerYou},
							Destination: zone.Battlefield,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Return all artifact, enchantment, and planeswalker cards from your graveyard to the battlefield.
		`,
		},
	}
}
