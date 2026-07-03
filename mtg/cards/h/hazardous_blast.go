package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HazardousBlast is the card definition for Hazardous Blast.
//
// Type: Sorcery
// Cost: {3}{R}
//
// Oracle text:
//
//	Hazardous Blast deals 1 damage to each creature your opponents control. Creatures your opponents control can't block this turn.
var HazardousBlast = newHazardousBlast()

func newHazardousBlast() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Hazardous Blast",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Damage{
							Amount:    game.Fixed(1),
							Recipient: game.GroupDamageRecipient(game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerOpponent})),
						},
					},
					{
						Primitive: game.ApplyRule{
							RuleEffects: []game.RuleEffect{
								game.RuleEffect{
									Kind:               game.RuleEffectCantBlock,
									AffectedController: game.ControllerOpponent,
									PermanentTypes:     []types.Card{types.Creature},
								},
							},
							Duration: game.DurationThisTurn,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Hazardous Blast deals 1 damage to each creature your opponents control. Creatures your opponents control can't block this turn.
		`,
		},
	}
}
