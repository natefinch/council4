package j

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// JuntuStakes is the card definition for Juntu Stakes.
//
// Type: Artifact
// Cost: {2}
//
// Oracle text:
//
//	Creatures with power 1 or less don't untap during their controllers' untap steps.
var JuntuStakes = newJuntuStakes

func newJuntuStakes() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Juntu Stakes",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
			Types: []types.Card{types.Artifact},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Creatures with power 1 or less don't untap during their controllers' untap steps.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectDoesntUntap,
							Duration:       game.DurationPermanent,
							PermanentTypes: []types.Card{types.Creature},
							AffectedSelection: game.Selection{Power: opt.Val(compare.Int{
								Op:    compare.LessOrEqual,
								Value: 1,
							})},
						},
					},
				},
			},
			OracleText: `
			Creatures with power 1 or less don't untap during their controllers' untap steps.
		`,
		},
	}
}
