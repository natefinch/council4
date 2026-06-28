package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Respite is the card definition for Respite.
//
// Type: Instant
// Cost: {1}{G}
//
// Oracle text:
//
//	Prevent all combat damage that would be dealt this turn. You gain 1 life for each attacking creature.
var Respite = newRespite()

func newRespite() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Respite",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.PreventDamage{
							All:        true,
							CombatOnly: true,
							Global:     true,
						},
					},
					{
						Primitive: game.GainLife{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:       game.DynamicAmountCountSelector,
								Multiplier: 1,
								Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, CombatState: game.CombatStateAttacking}),
							}),
							Player: game.ControllerReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Prevent all combat damage that would be dealt this turn. You gain 1 life for each attacking creature.
		`,
		},
	}
}
