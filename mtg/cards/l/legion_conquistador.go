package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// LegionConquistador is the card definition for Legion Conquistador.
//
// Type: Creature — Vampire Soldier
// Cost: {2}{W}
//
// Oracle text:
//
//	When this creature enters, you may search your library for any number of cards named Legion Conquistador, reveal them, put them into your hand, then shuffle.
var LegionConquistador = newLegionConquistador

func newLegionConquistador() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Legion Conquistador",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Vampire, types.Soldier},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentEnteredBattlefield,
							Source: game.TriggerSourceSelf,
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
										Name:        "Legion Conquistador",
										Reveal:      true,
										AnyNumber:   true,
									},
									Amount: game.Fixed(0),
								},
								Optional: true,
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			When this creature enters, you may search your library for any number of cards named Legion Conquistador, reveal them, put them into your hand, then shuffle.
		`,
		},
	}
}
