package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ManaSeism is the card definition for Mana Seism.
//
// Type: Sorcery
// Cost: {1}{R}
//
// Oracle text:
//
//	Sacrifice any number of lands, then add that much {C}.
var ManaSeism = newManaSeism()

func newManaSeism() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Mana Seism",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.SacrificePermanents{
							AnyNumber: true,
							Player:    game.ControllerReference(),
							Selection: game.Selection{RequiredTypes: []types.Card{types.Land}},
						},
						PublishResult: game.ResultKey("sacrificed-this-way"),
					},
					{
						Primitive: game.AddMana{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:      game.DynamicAmountPreviousEffectResult,
								ResultKey: game.ResultKey("sacrificed-this-way"),
							}),
							ManaColor: mana.C,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Sacrifice any number of lands, then add that much {C}.
		`,
		},
	}
}
