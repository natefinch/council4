package z

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ZombieApocalypse is the card definition for Zombie Apocalypse.
//
// Type: Sorcery
// Cost: {3}{B}{B}{B}
//
// Oracle text:
//
//	Return all Zombie creature cards from your graveyard to the battlefield tapped, then destroy all Humans.
var ZombieApocalypse = newZombieApocalypse

func newZombieApocalypse() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Zombie Apocalypse",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
				cost.B,
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.MassReturnFromGraveyard{
							Player:      game.ControllerReference(),
							Selection:   game.Selection{RequiredTypes: []types.Card{types.Creature}, SubtypesAny: []types.Sub{types.Zombie}, Controller: game.ControllerYou},
							Destination: zone.Battlefield,
							EntryTapped: true,
						},
					},
					{
						Primitive: game.Destroy{
							Group: game.BattlefieldGroup(game.Selection{SubtypesAny: []types.Sub{types.Sub("Human")}}),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Return all Zombie creature cards from your graveyard to the battlefield tapped, then destroy all Humans.
		`,
		},
	}
}
