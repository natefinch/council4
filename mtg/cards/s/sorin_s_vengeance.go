package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SorinSVengeance is the card definition for Sorin's Vengeance.
//
// Type: Sorcery
// Cost: {4}{B}{B}{B}
//
// Oracle text:
//
//	Sorin's Vengeance deals 10 damage to target player or planeswalker and you gain 10 life.
var SorinSVengeance = newSorinSVengeance

func newSorinSVengeance() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Sorin's Vengeance",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.B,
				cost.B,
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target player or planeswalker",
						Allow:      game.TargetAllowPermanent | game.TargetAllowPlayer,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Planeswalker}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Damage{
							Amount:    game.Fixed(10),
							Recipient: game.AnyTargetDamageRecipient(0),
						},
					},
					{
						Primitive: game.GainLife{
							Amount: game.Fixed(10),
							Player: game.ControllerReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Sorin's Vengeance deals 10 damage to target player or planeswalker and you gain 10 life.
		`,
		},
	}
}
