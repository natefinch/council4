package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// FalseMemories is the card definition for False Memories.
//
// Type: Instant
// Cost: {1}{U}
//
// Oracle text:
//
//	Mill seven cards. At the beginning of the next end step, exile seven cards from your graveyard.
var FalseMemories = newFalseMemories()

func newFalseMemories() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "False Memories",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Mill{
							Amount: game.Fixed(7),
							Player: game.ControllerReference(),
						},
					},
					{
						Primitive: game.CreateDelayedTrigger{
							Trigger: game.DelayedTriggerDef{
								Timing: game.DelayedAtBeginningOfNextEndStep,
								Content: game.Mode{
									Sequence: []game.Instruction{
										{
											Primitive: game.ChooseFromZone{
												Player:     game.ControllerReference(),
												SourceZone: zone.Graveyard,
												Filter:     game.Selection{Controller: game.ControllerYou},
												Quantity:   game.Fixed(7),
												Destination: game.ChooseDestination{
													Zone: zone.Exile,
												},
												Prompt: "Choose a card to exile",
											},
										},
									},
								}.Ability(),
							},
						},
					},
				},
			}.Ability()),
			OracleText: `
			Mill seven cards. At the beginning of the next end step, exile seven cards from your graveyard.
		`,
		},
	}
}
