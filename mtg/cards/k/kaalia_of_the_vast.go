package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// KaaliaOfTheVast is the card definition for Kaalia of the Vast.
//
// Type: Legendary Creature — Human Cleric
// Cost: {1}{R}{W}{B}
//
// Oracle text:
//
//	Flying
//	Whenever Kaalia attacks an opponent, you may put an Angel, Demon, or Dragon creature card from your hand onto the battlefield tapped and attacking that opponent.
var KaaliaOfTheVast = newKaaliaOfTheVast

func newKaaliaOfTheVast() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Black, color.Red),
		CardFace: game.CardFace{
			Name: "Kaalia of the Vast",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
				cost.W,
				cost.B,
			}),
			Colors:     []color.Color{color.Black, color.Red, color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Cleric},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:           game.EventAttackerDeclared,
							Source:          game.TriggerSourceSelf,
							Player:          game.TriggerPlayerOpponent,
							AttackRecipient: game.AttackRecipientPlayer,
						},
					},
					Optional: true,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ChooseFromZone{
									Player:     game.ControllerReference(),
									SourceZone: zone.Hand,
									Filter:     game.Selection{RequiredTypes: []types.Card{types.Creature}, SubtypesAny: []types.Sub{types.Sub("Angel"), types.Sub("Demon"), types.Sub("Dragon")}},
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
			Flying
			Whenever Kaalia attacks an opponent, you may put an Angel, Demon, or Dragon creature card from your hand onto the battlefield tapped and attacking that opponent.
		`,
		},
	}
}
