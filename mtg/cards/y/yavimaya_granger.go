package y

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// YavimayaGranger is the card definition for Yavimaya Granger.
//
// Type: Creature — Elf
// Cost: {2}{G}
//
// Oracle text:
//
//	Echo {2}{G} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
//	When this creature enters, you may search your library for a basic land card, put that card onto the battlefield tapped, then shuffle.
var YavimayaGranger = newYavimayaGranger

func newYavimayaGranger() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Yavimaya Granger",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elf},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.EchoTriggeredAbility(cost.Mana{cost.O(2), cost.G}),
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
										SourceZone:   zone.Library,
										Destination:  zone.Battlefield,
										Filter:       game.Selection{RequiredTypes: []types.Card{types.Land}, Supertypes: []types.Super{types.Basic}},
										EntersTapped: true,
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
			Echo {2}{G} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
			When this creature enters, you may search your library for a basic land card, put that card onto the battlefield tapped, then shuffle.
		`,
		},
	}
}
