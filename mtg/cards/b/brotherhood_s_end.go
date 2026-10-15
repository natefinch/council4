package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BrotherhoodSEnd is the card definition for Brotherhood's End.
//
// Type: Sorcery
// Cost: {1}{R}{R}
//
// Oracle text:
//
//	Choose one —
//	• Brotherhood's End deals 3 damage to each creature and each planeswalker.
//	• Destroy all artifacts with mana value 3 or less.
var BrotherhoodSEnd = newBrotherhoodSEnd

func newBrotherhoodSEnd() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Brotherhood's End",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Brotherhood's End deals 3 damage to each creature and each planeswalker.",
						Sequence: []game.Instruction{
							{
								Primitive: game.Damage{
									Amount:    game.Fixed(3),
									Recipient: game.GroupDamageRecipient(game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}})),
								},
							},
							{
								Primitive: game.Damage{
									Amount:    game.Fixed(3),
									Recipient: game.GroupDamageRecipient(game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Planeswalker}})),
								},
							},
						},
					},
					game.Mode{
						Text: "Destroy all artifacts with mana value 3 or less.",
						Sequence: []game.Instruction{
							{
								Primitive: game.Destroy{
									Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Artifact}, ManaValue: opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 3})}),
								},
							},
						},
					},
				},
				MinModes: 1,
				MaxModes: 1,
			}),
			OracleText: `
			Choose one —
			• Brotherhood's End deals 3 damage to each creature and each planeswalker.
			• Destroy all artifacts with mana value 3 or less.
		`,
		},
	}
}
