package j

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// JointExploration is the card definition for Joint Exploration.
//
// Type: Instant
// Cost: {1}{U}
//
// Oracle text:
//
//	Kicker {G} (You may pay an additional {G} as you cast this spell.)
//	Scry 2, then draw a card. If this spell was kicked, you may put a land card from your hand onto the battlefield.
var JointExploration = newJointExploration

func newJointExploration() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Green),
		CardFace: game.CardFace{
			Name: "Joint Exploration",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Kicker {G}",
					KeywordAbilities: []game.KeywordAbility{
						game.KickerKeyword{
							Cost: cost.Mana{
								cost.G,
							},
						},
					},
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Scry{
							Amount: game.Fixed(2),
							Player: game.ControllerReference(),
						},
					},
					{
						Primitive: game.Draw{
							Amount: game.Fixed(1),
							Player: game.ControllerReference(),
						},
					},
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
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								SpellWasKicked: true,
							}),
						}),
						Optional: true,
					},
				},
			}.Ability()),
			OracleText: `
			Kicker {G} (You may pay an additional {G} as you cast this spell.)
			Scry 2, then draw a card. If this spell was kicked, you may put a land card from your hand onto the battlefield.
		`,
		},
	}
}
