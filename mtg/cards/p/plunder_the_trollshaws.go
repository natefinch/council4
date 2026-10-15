package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// PlunderTheTrollshaws is the card definition for Plunder the Trollshaws.
//
// Type: Instant
// Cost: {1}{U}
//
// Oracle text:
//
//	Draw a card. If this spell was cast from a graveyard, draw two cards instead.
//	Flashback {3}{U} (You may cast this card from your graveyard for its flashback cost. Then exile it.)
var PlunderTheTrollshaws = newPlunderTheTrollshaws

func newPlunderTheTrollshaws() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Plunder the Trollshaws",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Flashback {3}{U}",
					KeywordAbilities: []game.KeywordAbility{
						game.FlashbackKeyword{
							Cost: cost.Mana{
								cost.O(3),
								cost.U,
							},
						},
					},
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Draw{
							Amount: game.Fixed(1),
							Player: game.ControllerReference(),
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								Negate:       true,
								CastFromZone: opt.Val(zone.Graveyard),
							}),
						}),
						PublishCondition: "condition-0",
					},
					{
						Primitive: game.Draw{
							Amount: game.Fixed(2),
							Player: game.ControllerReference(),
						},
						ConditionGate:       "condition-0",
						ConditionGateNegate: true,
					},
				},
			}.Ability()),
			OracleText: `
			Draw a card. If this spell was cast from a graveyard, draw two cards instead.
			Flashback {3}{U} (You may cast this card from your graveyard for its flashback cost. Then exile it.)
		`,
		},
	}
}
