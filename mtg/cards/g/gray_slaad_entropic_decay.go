package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// GraySlaad is the card definition for Gray Slaad // Entropic Decay.
//
// Type: Creature — Frog Horror // Sorcery — Adventure
// Cost: {2}{B} // {1}{B}
// Face: Entropic Decay — Sorcery — Adventure ({1}{B})
//
// Oracle text:
//
//	As long as there are four or more creature cards in your graveyard, this creature has menace and deathtouch.
var GraySlaad = newGraySlaad

func newGraySlaad() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Gray Slaad",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Frog, types.Horror},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "As long as there are four or more creature cards in your graveyard, this creature has menace and deathtouch.",
					Condition: opt.Val(game.Condition{
						Text:              "As long as there are four or more creature cards in your graveyard",
						SourceCounterKind: counter.PlusOnePlusOne,
						ControllerGraveyardCardOfTypeCountAtLeast: 4,
						ControllerGraveyardCountCardType:          types.Creature,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerAbility,
							AddKeywords:    []game.Keyword{game.Menace, game.Deathtouch},
						},
					},
				},
			},
			OracleText: `
			As long as there are four or more creature cards in your graveyard, this creature has menace and deathtouch.
		`,
		},
		Layout: game.LayoutAdventure,
		Alternate: opt.Val(game.CardFace{
			Name: "Entropic Decay",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
			}),
			Colors:   []color.Color{color.Black},
			Types:    []types.Card{types.Sorcery},
			Subtypes: []types.Sub{types.Adventure},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.MoveTopOfLibrary{
							Amount:      game.Fixed(4),
							Player:      game.ControllerReference(),
							Destination: zone.Graveyard,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Mill four cards. (Then exile this card. You may cast the creature later from exile.)
		`,
		}),
	}
}
