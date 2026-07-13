package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// InspiringStatuary is the card definition for Inspiring Statuary.
//
// Type: Artifact
// Cost: {3}
//
// Oracle text:
//
//	Nonartifact spells you cast have improvise. (Your artifacts can help cast those spells. Each artifact you tap after you're done activating mana abilities pays for {1}.)
var InspiringStatuary = newInspiringStatuary

func newInspiringStatuary() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Inspiring Statuary",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types: []types.Card{types.Artifact},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:               game.RuleEffectGrantSpellKeyword,
							AffectedController: game.ControllerYou,
							CardSelection:      game.Selection{ExcludedTypes: []types.Card{types.Artifact}},
							GrantedKeyword:     game.Improvise,
						},
					},
				},
			},
			OracleText: `
			Nonartifact spells you cast have improvise. (Your artifacts can help cast those spells. Each artifact you tap after you're done activating mana abilities pays for {1}.)
		`,
		},
	}
}
