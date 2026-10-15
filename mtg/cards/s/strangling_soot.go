package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// StranglingSoot is the card definition for Strangling Soot.
//
// Type: Instant
// Cost: {2}{B}
//
// Oracle text:
//
//	Destroy target creature with toughness 3 or less.
//	Flashback {5}{R} (You may cast this card from your graveyard for its flashback cost. Then exile it.)
var StranglingSoot = newStranglingSoot

func newStranglingSoot() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Red),
		CardFace: game.CardFace{
			Name: "Strangling Soot",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Instant},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Flashback {5}{R}",
					KeywordAbilities: []game.KeywordAbility{
						game.FlashbackKeyword{
							Cost: cost.Mana{
								cost.O(5),
								cost.R,
							},
						},
					},
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target creature with toughness 3 or less",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Toughness: opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 3})}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Destroy{
							Object: game.TargetPermanentReference(0),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Destroy target creature with toughness 3 or less.
			Flashback {5}{R} (You may cast this card from your graveyard for its flashback cost. Then exile it.)
		`,
		},
	}
}
