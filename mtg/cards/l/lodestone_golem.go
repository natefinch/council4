package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// LodestoneGolem is the card definition for Lodestone Golem.
//
// Type: Artifact Creature — Golem
// Cost: {4}
//
// Oracle text:
//
//	Nonartifact spells cost {1} more to cast.
var LodestoneGolem = newLodestoneGolem()

func newLodestoneGolem() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Lodestone Golem",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Golem},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind: game.RuleEffectCostModifier,
							CostModifier: game.CostModifier{
								Kind:            game.CostModifierSpell,
								CardSelection:   game.Selection{ExcludedTypes: []types.Card{types.Artifact}},
								GenericIncrease: 1,
							},
						},
					},
				},
			},
			OracleText: `
			Nonartifact spells cost {1} more to cast.
		`,
		},
	}
}
