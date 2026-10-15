package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// DeathbellowWarCry is the card definition for Deathbellow War Cry.
//
// Type: Sorcery
// Cost: {5}{R}{R}{R}
//
// Oracle text:
//
//	Search your library for up to four Minotaur creature cards with different names, put them onto the battlefield, then shuffle.
var DeathbellowWarCry = newDeathbellowWarCry

func newDeathbellowWarCry() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Deathbellow War Cry",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.R,
				cost.R,
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Search{
							Player: game.ControllerReference(),
							Spec: game.SearchSpec{
								SourceZone:     zone.Library,
								Destination:    zone.Battlefield,
								Filter:         game.Selection{RequiredTypes: []types.Card{types.Creature}, SubtypesAny: []types.Sub{types.Minotaur}},
								DifferentNames: true,
							},
							Amount: game.Fixed(4),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Search your library for up to four Minotaur creature cards with different names, put them onto the battlefield, then shuffle.
		`,
		},
	}
}
