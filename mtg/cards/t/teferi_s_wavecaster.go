package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// TeferiSWavecaster is the card definition for Teferi's Wavecaster.
//
// Type: Creature — Merfolk Wizard
// Cost: {3}{U}{U}
//
// Oracle text:
//
//	Flash
//	When this creature enters, you may search your library and/or graveyard for a card named Teferi, Timeless Voyager, reveal it, and put it into your hand. If you search your library this way, shuffle.
var TeferiSWavecaster = newTeferiSWavecaster()

func newTeferiSWavecaster() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Teferi's Wavecaster",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Merfolk, types.Wizard},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.FlashStaticBody,
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
										Name:          "Teferi, Timeless Voyager",
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
			Flash
			When this creature enters, you may search your library and/or graveyard for a card named Teferi, Timeless Voyager, reveal it, and put it into your hand. If you search your library this way, shuffle.
		`,
		},
	}
}
