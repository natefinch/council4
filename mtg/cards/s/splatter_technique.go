package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SplatterTechnique is the card definition for Splatter Technique.
//
// Type: Sorcery
// Cost: {1}{U}{U}{R}{R}
//
// Oracle text:
//
//	Choose one —
//	• Draw four cards.
//	• Splatter Technique deals 4 damage to each creature and planeswalker.
var SplatterTechnique = newSplatterTechnique

func newSplatterTechnique() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Red),
		CardFace: game.CardFace{
			Name: "Splatter Technique",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
				cost.U,
				cost.R,
				cost.R,
			}),
			Colors: []color.Color{color.Red, color.Blue},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Draw four cards.",
						Sequence: []game.Instruction{
							{
								Primitive: game.Draw{
									Amount: game.Fixed(4),
									Player: game.ControllerReference(),
								},
							},
						},
					},
					game.Mode{
						Text: "Splatter Technique deals 4 damage to each creature and planeswalker.",
						Sequence: []game.Instruction{
							{
								Primitive: game.Damage{
									Amount:    game.Fixed(4),
									Recipient: game.GroupDamageRecipient(game.BattlefieldGroup(game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Planeswalker}})),
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
			• Draw four cards.
			• Splatter Technique deals 4 damage to each creature and planeswalker.
		`,
		},
	}
}
