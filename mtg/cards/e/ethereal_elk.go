package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// EtherealElk is the card definition for Ethereal Elk.
//
// Type: Creature — Elk Spirit
// Cost: {3}{G}{G}
//
// Oracle text:
//
//	Trample
//	When this creature enters, you may search your library and/or graveyard for a card named Vivien, Nature's Avenger, reveal it, and put it into your hand. If you search your library this way, shuffle.
var EtherealElk = newEtherealElk()

func newEtherealElk() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Ethereal Elk",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elk, types.Spirit},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
			},
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
										SourceZone:    zone.Library,
										Destination:   zone.Hand,
										Name:          "Vivien, Nature's Avenger",
										Reveal:        true,
										AlsoGraveyard: true,
									},
									Amount: game.Fixed(1),
								},
								Optional:      true,
								OptionalActor: opt.Val(game.ControllerReference()),
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Trample
			When this creature enters, you may search your library and/or graveyard for a card named Vivien, Nature's Avenger, reveal it, and put it into your hand. If you search your library this way, shuffle.
		`,
		},
	}
}
