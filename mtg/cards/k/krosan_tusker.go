package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// KrosanTusker is the card definition for Krosan Tusker.
//
// Type: Creature — Boar Beast
// Cost: {5}{G}{G}
//
// Oracle text:
//
//	Cycling {2}{G} ({2}{G}, Discard this card: Draw a card.)
//	When you cycle this card, you may search your library for a basic land card, reveal that card, put it into your hand, then shuffle. (Do this before you draw.)
var KrosanTusker = newKrosanTusker

func newKrosanTusker() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Krosan Tusker",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Boar, types.Beast},
			Power:     opt.Val(game.PT{Value: 6}),
			Toughness: opt.Val(game.PT{Value: 5}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.CyclingActivatedAbility(cost.Mana{cost.O(2), cost.G}),
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
										Destination: zone.Hand,
										Filter:      game.Selection{RequiredTypes: []types.Card{types.Land}, Supertypes: []types.Super{types.Basic}},
										Reveal:      true,
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
			Cycling {2}{G} ({2}{G}, Discard this card: Draw a card.)
			When you cycle this card, you may search your library for a basic land card, reveal that card, put it into your hand, then shuffle. (Do this before you draw.)
		`,
		},
	}
}
