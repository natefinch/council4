package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DesolationTwin is the card definition for Desolation Twin.
//
// Type: Creature — Eldrazi
// Cost: {10}
//
// Oracle text:
//
//	When you cast this spell, create a 10/10 colorless Eldrazi creature token.
var DesolationTwin = newDesolationTwin

func newDesolationTwin() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Desolation Twin",
			ManaCost: opt.Val(cost.Mana{
				cost.O(10),
			}),
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Eldrazi},
			Power:     opt.Val(game.PT{Value: 10}),
			Toughness: opt.Val(game.PT{Value: 10}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:       game.EventSpellCast,
							Source:      game.TriggerSourceSelf,
							Controller:  game.TriggerControllerYou,
							SelfWasCast: true,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(desolationTwinToken),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			When you cast this spell, create a 10/10 colorless Eldrazi creature token.
		`,
		},
	}
}

var desolationTwinToken = newDesolationTwinToken()

func newDesolationTwinToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Eldrazi",
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Eldrazi},
			Power:     opt.Val(game.PT{Value: 10}),
			Toughness: opt.Val(game.PT{Value: 10}),
		},
	}
}
