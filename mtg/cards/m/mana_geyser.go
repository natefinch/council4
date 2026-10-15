package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ManaGeyser is the card definition for Mana Geyser.
//
// Type: Sorcery
// Cost: {3}{R}{R}
//
// Oracle text:
//
//	Add {R} for each tapped land your opponents control.
var ManaGeyser = newManaGeyser

func newManaGeyser() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Mana Geyser",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.AddMana{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:       game.DynamicAmountCountSelector,
								Multiplier: 1,
								Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Land}, Controller: game.ControllerOpponent, Tapped: game.TriTrue}),
							}),
							ManaColor: mana.R,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Add {R} for each tapped land your opponents control.
		`,
		},
	}
}
