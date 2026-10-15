package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// BuriedAlive is the card definition for Buried Alive.
//
// Type: Sorcery
// Cost: {2}{B}
//
// Oracle text:
//
//	Search your library for up to three creature cards, put them into your graveyard, then shuffle.
var BuriedAlive = newBuriedAlive

func newBuriedAlive() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Buried Alive",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Search{
							Player: game.ControllerReference(),
							Spec: game.SearchSpec{
								SourceZone:  zone.Library,
								Destination: zone.Graveyard,
								Filter:      game.Selection{RequiredTypes: []types.Card{types.Creature}},
							},
							Amount: game.Fixed(3),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Search your library for up to three creature cards, put them into your graveyard, then shuffle.
		`,
		},
	}
}
