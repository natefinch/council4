package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TectonicHazard is the card definition for Tectonic Hazard.
//
// Type: Sorcery
// Cost: {R}
//
// Oracle text:
//
//	Tectonic Hazard deals 1 damage to each opponent and each creature they control.
var TectonicHazard = newTectonicHazard

func newTectonicHazard() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Tectonic Hazard",
			ManaCost: opt.Val(cost.Mana{
				cost.R,
			}),
			Colors: []color.Color{color.Red},
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
						Primitive: game.Damage{
							Amount:    game.Fixed(1),
							Recipient: game.GroupDamageRecipient(game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerOpponent})),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Tectonic Hazard deals 1 damage to each opponent and each creature they control.
		`,
		},
	}
}
