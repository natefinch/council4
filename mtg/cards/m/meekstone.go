package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Meekstone is the card definition for Meekstone.
//
// Type: Artifact
// Cost: {1}
//
// Oracle text:
//
//	Creatures with power 3 or greater don't untap during their controllers' untap steps.
var Meekstone = newMeekstone

func newMeekstone() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Meekstone",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
			}),
			Types: []types.Card{types.Artifact},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Creatures with power 3 or greater don't untap during their controllers' untap steps.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectDoesntUntap,
							Duration:       game.DurationPermanent,
							PermanentTypes: []types.Card{types.Creature},
							AffectedSelection: game.Selection{Power: opt.Val(compare.Int{
								Op:    compare.GreaterOrEqual,
								Value: 3,
							})},
						},
					},
				},
			},
			OracleText: `
			Creatures with power 3 or greater don't untap during their controllers' untap steps.
		`,
		},
	}
}
