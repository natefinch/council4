package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Dragonstorm is the card definition for Dragonstorm.
//
// Type: Sorcery
// Cost: {8}{R}
//
// Oracle text:
//
//	Search your library for a Dragon permanent card, put it onto the battlefield, then shuffle.
//	Storm (When you cast this spell, copy it for each spell cast before it this turn.)
var Dragonstorm = newDragonstorm

func newDragonstorm() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Dragonstorm",
			ManaCost: opt.Val(cost.Mana{
				cost.O(8),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			StaticAbilities: []game.StaticAbility{
				game.StormStaticBody,
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Search{
							Player: game.ControllerReference(),
							Spec: game.SearchSpec{
								SourceZone:  zone.Library,
								Destination: zone.Battlefield,
								Filter:      game.Selection{SubtypesAny: []types.Sub{types.Dragon}, RequirePermanentCard: true},
							},
							Amount: game.Fixed(1),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Search your library for a Dragon permanent card, put it onto the battlefield, then shuffle.
			Storm (When you cast this spell, copy it for each spell cast before it this turn.)
		`,
		},
	}
}
