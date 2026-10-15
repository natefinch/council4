package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ThunderbreakRegent is the card definition for Thunderbreak Regent.
//
// Type: Creature — Dragon
// Cost: {2}{R}{R}
//
// Oracle text:
//
//	Flying
//	Whenever a Dragon you control becomes the target of a spell or ability an opponent controls, this creature deals 3 damage to that player.
var ThunderbreakRegent = newThunderbreakRegent

func newThunderbreakRegent() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Thunderbreak Regent",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Dragon},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever a Dragon you control becomes the target of a spell or ability an opponent controls, this creature deals 3 damage to that player.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventObjectBecameTarget,
							Controller:       game.TriggerControllerYou,
							CauseController:  game.TriggerControllerOpponent,
							SubjectSelection: game.Selection{SubtypesAny: []types.Sub{types.Dragon}},
							StackObjectKind:  game.StackSpell,
							CounterKind:      counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Damage{
											Amount:       game.Fixed(3),
											Recipient:    game.PlayerDamageRecipient(game.EventPlayerReference()),
											DamageSource: opt.Val(game.SourcePermanentReference()),
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
			},
			OracleText: `
			Flying
			Whenever a Dragon you control becomes the target of a spell or ability an opponent controls, this creature deals 3 damage to that player.
		`,
		},
	}
}
