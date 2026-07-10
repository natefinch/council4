package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ShefetMonitor is the card definition for Shefet Monitor.
//
// Type: Creature — Lizard
// Cost: {5}{G}
//
// Oracle text:
//
//	Cycling {3}{G} ({3}{G}, Discard this card: Draw a card.)
//	When you cycle this card, you may search your library for a basic land card or a Desert card, put it onto the battlefield, then shuffle. (Do this before you draw.)
var ShefetMonitor = newShefetMonitor

func newShefetMonitor() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Shefet Monitor",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Lizard},
			Power:     opt.Val(game.PT{Value: 6}),
			Toughness: opt.Val(game.PT{Value: 5}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.CyclingActivatedAbility(cost.Mana{cost.O(3), cost.G}),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventCycled,
							Source: game.TriggerSourceSelf,
							Player: game.TriggerPlayerYou,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Search{
									Player: game.ControllerReference(),
									Spec: game.SearchSpec{
										SourceZone:  zone.Library,
										Destination: zone.Battlefield,
										Filter:      game.Selection{AnyOf: []game.Selection{game.Selection{RequiredTypes: []types.Card{types.Land}, Supertypes: []types.Super{types.Basic}}, game.Selection{SubtypesAny: []types.Sub{types.Sub("Desert")}}}},
									},
									Amount: game.Fixed(1),
								},
								Optional: true,
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Cycling {3}{G} ({3}{G}, Discard this card: Draw a card.)
			When you cycle this card, you may search your library for a basic land card or a Desert card, put it onto the battlefield, then shuffle. (Do this before you draw.)
		`,
		},
	}
}
