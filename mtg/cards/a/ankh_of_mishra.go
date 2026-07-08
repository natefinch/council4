package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AnkhOfMishra is the card definition for Ankh of Mishra.
//
// Type: Artifact
// Cost: {2}
//
// Oracle text:
//
//	Whenever a land enters, this artifact deals 2 damage to that land's controller.
var AnkhOfMishra = newAnkhOfMishra

func newAnkhOfMishra() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Ankh of Mishra",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
			Types: []types.Card{types.Artifact},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentEnteredBattlefield,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Land}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Damage{
									Amount:       game.Fixed(2),
									Recipient:    game.PlayerDamageRecipient(game.ObjectControllerReference(game.EventPermanentReference())),
									DamageSource: opt.Val(game.SourcePermanentReference()),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Whenever a land enters, this artifact deals 2 damage to that land's controller.
		`,
		},
	}
}
