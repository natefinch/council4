package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ChordOfCalling is the card definition for Chord of Calling.
//
// Type: Instant
// Cost: {X}{G}{G}{G}
//
// Oracle text:
//
//	Convoke (Your creatures can help cast this spell. Each creature you tap while casting this spell pays for {1} or one mana of that creature's color.)
//	Search your library for a creature card with mana value X or less, put it onto the battlefield, then shuffle.
var ChordOfCalling = newChordOfCalling

func newChordOfCalling() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Chord of Calling",
			ManaCost: opt.Val(cost.Mana{
				cost.X,
				cost.G,
				cost.G,
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Instant},
			StaticAbilities: []game.StaticAbility{
				game.ConvokeStaticBody,
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Search{
							Player: game.ControllerReference(),
							Spec: game.SearchSpec{
								SourceZone:        zone.Library,
								Destination:       zone.Battlefield,
								Filter:            game.Selection{RequiredTypes: []types.Card{types.Creature}},
								MaxManaValueFromX: true,
							},
							Amount: game.Fixed(1),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Convoke (Your creatures can help cast this spell. Each creature you tap while casting this spell pays for {1} or one mana of that creature's color.)
			Search your library for a creature card with mana value X or less, put it onto the battlefield, then shuffle.
		`,
		},
	}
}
