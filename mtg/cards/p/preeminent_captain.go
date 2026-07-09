package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// PreeminentCaptain is the card definition for Preeminent Captain.
//
// Type: Creature — Kithkin Soldier
// Cost: {2}{W}
//
// Oracle text:
//
//	First strike (This creature deals combat damage before creatures without first strike.)
//	Whenever this creature attacks, you may put a Soldier creature card from your hand onto the battlefield tapped and attacking.
var PreeminentCaptain = newPreeminentCaptain

func newPreeminentCaptain() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Preeminent Captain",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Kithkin, types.Soldier},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FirstStrikeStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:  game.EventAttackerDeclared,
							Source: game.TriggerSourceSelf,
						},
					},
					Optional: true,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ChooseFromZone{
									Player:     game.ControllerReference(),
									SourceZone: zone.Hand,
									Filter:     game.Selection{RequiredTypes: []types.Card{types.Creature}, SubtypesAny: []types.Sub{types.Sub("Soldier")}},
									Quantity:   game.Fixed(1),
									Destination: game.ChooseDestination{
										Zone: zone.Battlefield,
									},
									Riders: game.ChooseRiders{
										EntersTapped:    true,
										EntersAttacking: true,
									},
									Prompt: "Choose a card to put onto the battlefield",
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			First strike (This creature deals combat damage before creatures without first strike.)
			Whenever this creature attacks, you may put a Soldier creature card from your hand onto the battlefield tapped and attacking.
		`,
		},
	}
}
