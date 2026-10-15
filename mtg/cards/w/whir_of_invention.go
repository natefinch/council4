package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// WhirOfInvention is the card definition for Whir of Invention.
//
// Type: Instant
// Cost: {X}{U}{U}{U}
//
// Oracle text:
//
//	Improvise (Your artifacts can help cast this spell. Each artifact you tap after you're done activating mana abilities pays for {1}.)
//	Search your library for an artifact card with mana value X or less, put it onto the battlefield, then shuffle.
var WhirOfInvention = newWhirOfInvention

func newWhirOfInvention() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Whir of Invention",
			ManaCost: opt.Val(cost.Mana{
				cost.X,
				cost.U,
				cost.U,
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			StaticAbilities: []game.StaticAbility{
				game.ImproviseStaticBody,
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Search{
							Player: game.ControllerReference(),
							Spec: game.SearchSpec{
								SourceZone:        zone.Library,
								Destination:       zone.Battlefield,
								Filter:            game.Selection{RequiredTypes: []types.Card{types.Artifact}},
								MaxManaValueFromX: true,
							},
							Amount: game.Fixed(1),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Improvise (Your artifacts can help cast this spell. Each artifact you tap after you're done activating mana abilities pays for {1}.)
			Search your library for an artifact card with mana value X or less, put it onto the battlefield, then shuffle.
		`,
		},
	}
}
