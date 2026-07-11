package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Burgeoning is the card definition for Burgeoning.
//
// Type: Enchantment
// Cost: {G}
//
// Oracle text:
//
//	Whenever an opponent plays a land, you may put a land card from your hand onto the battlefield.
var Burgeoning = newBurgeoning

func newBurgeoning() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Burgeoning",
			ManaCost: opt.Val(cost.Mana{
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Enchantment},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:  game.EventLandPlayed,
							Player: game.TriggerPlayerOpponent,
						},
					},
					Optional: true,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ChooseFromZone{
									Player:     game.ControllerReference(),
									SourceZone: zone.Hand,
									Filter:     game.Selection{RequiredTypes: []types.Card{types.Land}},
									Quantity:   game.Fixed(1),
									Destination: game.ChooseDestination{
										Zone: zone.Battlefield,
									},
									Prompt: "Choose a card to put onto the battlefield",
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Whenever an opponent plays a land, you may put a land card from your hand onto the battlefield.
		`,
		},
	}
}
