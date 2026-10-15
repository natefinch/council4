package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ViciousRumors is the card definition for Vicious Rumors.
//
// Type: Sorcery
// Cost: {B}
//
// Oracle text:
//
//	Vicious Rumors deals 1 damage to each opponent. Each opponent discards a card, then mills a card. You gain 1 life.
var ViciousRumors = newViciousRumors

func newViciousRumors() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Vicious Rumors",
			ManaCost: opt.Val(cost.Mana{
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Damage{
							Amount:    game.Fixed(1),
							Recipient: game.PlayerGroupDamageRecipient(game.OpponentsReference()),
						},
					},
					{
						Primitive: game.Discard{
							Amount:      game.Fixed(1),
							PlayerGroup: game.OpponentsReference(),
						},
					},
					{
						Primitive: game.MoveTopOfLibrary{
							Amount:      game.Fixed(1),
							PlayerGroup: game.OpponentsReference(),
							Destination: zone.Graveyard,
						},
					},
					{
						Primitive: game.GainLife{
							Amount: game.Fixed(1),
							Player: game.ControllerReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Vicious Rumors deals 1 damage to each opponent. Each opponent discards a card, then mills a card. You gain 1 life.
		`,
		},
	}
}
