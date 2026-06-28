package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// AshiokSForerunner is the card definition for Ashiok's Forerunner.
//
// Type: Creature — Human Wizard
// Cost: {3}{U}{B}
//
// Oracle text:
//
//	Flash
//	When this creature enters, you may search your library and/or graveyard for a card named Ashiok, Sculptor of Fears, reveal it, and put it into your hand. If you search your library this way, shuffle.
var AshiokSForerunner = newAshiokSForerunner()

func newAshiokSForerunner() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black),
		CardFace: game.CardFace{
			Name: "Ashiok's Forerunner",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
				cost.B,
			}),
			Colors:    []color.Color{color.Black, color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Wizard},
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
										Name:          "Ashiok, Sculptor of Fears",
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
			When this creature enters, you may search your library and/or graveyard for a card named Ashiok, Sculptor of Fears, reveal it, and put it into your hand. If you search your library this way, shuffle.
		`,
		},
	}
}
