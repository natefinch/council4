package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BrightstoneRitual is the card definition for Brightstone Ritual.
//
// Type: Instant
// Cost: {R}
//
// Oracle text:
//
//	Add {R} for each Goblin on the battlefield.
var BrightstoneRitual = newBrightstoneRitual

func newBrightstoneRitual() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Brightstone Ritual",
			ManaCost: opt.Val(cost.Mana{
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.AddMana{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:       game.DynamicAmountCountSelector,
								Multiplier: 1,
								Group:      game.BattlefieldGroup(game.Selection{SubtypesAny: []types.Sub{types.Sub("Goblin")}}),
							}),
							ManaColor: mana.R,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Add {R} for each Goblin on the battlefield.
		`,
		},
	}
}
