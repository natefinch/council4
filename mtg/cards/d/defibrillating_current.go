package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DefibrillatingCurrent is the card definition for Defibrillating Current.
//
// Type: Sorcery
// Cost: {2/R}{2/W}{2/B}
//
// Oracle text:
//
//	Defibrillating Current deals 4 damage to target creature or planeswalker and you gain 2 life.
var DefibrillatingCurrent = newDefibrillatingCurrent

func newDefibrillatingCurrent() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Black, color.Red),
		CardFace: game.CardFace{
			Name: "Defibrillating Current",
			ManaCost: opt.Val(cost.Mana{
				cost.Twobrid(mana.R),
				cost.Twobrid(mana.W),
				cost.Twobrid(mana.B),
			}),
			Colors: []color.Color{color.Black, color.Red, color.White},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target creature or planeswalker",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Planeswalker}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Damage{
							Amount:    game.Fixed(4),
							Recipient: game.AnyTargetDamageRecipient(0),
						},
					},
					{
						Primitive: game.GainLife{
							Amount: game.Fixed(2),
							Player: game.ControllerReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Defibrillating Current deals 4 damage to target creature or planeswalker and you gain 2 life.
		`,
		},
	}
}
