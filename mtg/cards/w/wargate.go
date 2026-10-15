package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Wargate is the card definition for Wargate.
//
// Type: Sorcery
// Cost: {X}{G}{W}{U}
//
// Oracle text:
//
//	Search your library for a permanent card with mana value X or less, put it onto the battlefield, then shuffle.
var Wargate = newWargate

func newWargate() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue, color.Green),
		CardFace: game.CardFace{
			Name: "Wargate",
			ManaCost: opt.Val(cost.Mana{
				cost.X,
				cost.G,
				cost.W,
				cost.U,
			}),
			Colors: []color.Color{color.Green, color.Blue, color.White},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Search{
							Player: game.ControllerReference(),
							Spec: game.SearchSpec{
								SourceZone:        zone.Library,
								Destination:       zone.Battlefield,
								Filter:            game.Selection{RequirePermanentCard: true},
								MaxManaValueFromX: true,
							},
							Amount: game.Fixed(1),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Search your library for a permanent card with mana value X or less, put it onto the battlefield, then shuffle.
		`,
		},
	}
}
