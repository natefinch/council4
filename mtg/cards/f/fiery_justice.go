package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FieryJustice is the card definition for Fiery Justice.
//
// Type: Sorcery
// Cost: {R}{G}{W}
//
// Oracle text:
//
//	Fiery Justice deals 5 damage divided as you choose among any number of targets. Target opponent gains 5 life.
var FieryJustice = newFieryJustice

func newFieryJustice() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Fiery Justice",
			ManaCost: opt.Val(cost.Mana{
				cost.R,
				cost.G,
				cost.W,
			}),
			Colors: []color.Color{color.Green, color.Red, color.White},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 0,
						MaxTargets: 5,
						Constraint: "any number of targets",
						Allow:      game.TargetAllowPermanent | game.TargetAllowPlayer,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Planeswalker, types.Battle}}),
					},
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "Target opponent",
						Allow:      game.TargetAllowPlayer,
						Selection:  opt.Val(game.Selection{Player: game.PlayerOpponent}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Damage{
							Amount:    game.Fixed(5),
							Recipient: game.AnyTargetDamageRecipient(0),
							Divided:   true,
						},
					},
					{
						Primitive: game.GainLife{
							Amount: game.Fixed(5),
							Player: game.TargetPlayerReference(1),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Fiery Justice deals 5 damage divided as you choose among any number of targets. Target opponent gains 5 life.
		`,
		},
	}
}
