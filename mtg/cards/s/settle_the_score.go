package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SettleTheScore is the card definition for Settle the Score.
//
// Type: Sorcery
// Cost: {2}{B}{B}
//
// Oracle text:
//
//	Exile target creature. Put two loyalty counters on a planeswalker you control.
var SettleTheScore = newSettleTheScore

func newSettleTheScore() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Settle the Score",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target creature",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.MovePermanent{
							Object:      game.TargetPermanentReference(0),
							Destination: zone.Exile,
						},
					},
					{
						Primitive: game.AddCounter{
							Amount:      game.Fixed(2),
							Group:       game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Planeswalker}, Controller: game.ControllerYou}),
							CounterKind: counter.Loyalty,
							ChooseOne:   true,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Exile target creature. Put two loyalty counters on a planeswalker you control.
		`,
		},
	}
}
