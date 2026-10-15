package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Reshape is the card definition for Reshape.
//
// Type: Sorcery
// Cost: {X}{U}{U}
//
// Oracle text:
//
//	As an additional cost to cast this spell, sacrifice an artifact.
//	Search your library for an artifact card with mana value X or less, put it onto the battlefield, then shuffle.
var Reshape = newReshape

func newReshape() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Reshape",
			ManaCost: opt.Val(cost.Mana{
				cost.X,
				cost.U,
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Sorcery},
			AdditionalCosts: []cost.Additional{
				{
					Kind:               cost.AdditionalSacrifice,
					Text:               "sacrifice an artifact",
					Amount:             1,
					MatchPermanentType: true,
					PermanentType:      types.Artifact,
				},
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
			As an additional cost to cast this spell, sacrifice an artifact.
			Search your library for an artifact card with mana value X or less, put it onto the battlefield, then shuffle.
		`,
		},
	}
}
