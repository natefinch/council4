package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Traumatize is the card definition for Traumatize.
//
// Type: Sorcery
// Cost: {3}{U}{U}
//
// Oracle text:
//
//	Target player mills half their library, rounded down.
var Traumatize = newTraumatize

func newTraumatize() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Traumatize",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "Target player",
						Allow:      game.TargetAllowPlayer,
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Mill{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:      game.DynamicAmountCountCardsInZone,
								Divisor:   2,
								Player:    func() *game.PlayerReference { ref := game.TargetPlayerReference(0); return &ref }(),
								CardZone:  zone.Library,
								Selection: &game.Selection{},
							}),
							Player: game.TargetPlayerReference(0),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Target player mills half their library, rounded down.
		`,
		},
	}
}
