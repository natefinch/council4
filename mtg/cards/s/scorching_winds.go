package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ScorchingWinds is the card definition for Scorching Winds.
//
// Type: Instant
// Cost: {R}
//
// Oracle text:
//
//	Cast this spell only during the declare attackers step and only if you've been attacked this step.
//	Scorching Winds deals 1 damage to each attacking creature.
var ScorchingWinds = newScorchingWinds

func newScorchingWinds() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Scorching Winds",
			ManaCost: opt.Val(cost.Mana{
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Instant},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text:                          "Cast this spell only during the declare attackers step and only if you've been attacked this step.",
					CastOnlyAfterAttackedThisStep: true,
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Damage{
							Amount:    game.Fixed(1),
							Recipient: game.GroupDamageRecipient(game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, CombatState: game.CombatStateAttacking})),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Cast this spell only during the declare attackers step and only if you've been attacked this step.
			Scorching Winds deals 1 damage to each attacking creature.
		`,
		},
	}
}
