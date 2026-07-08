package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DoranTheSiegeTower is the card definition for Doran, the Siege Tower.
//
// Type: Legendary Creature — Treefolk Shaman
// Cost: {W}{B}{G}
//
// Oracle text:
//
//	Each creature assigns combat damage equal to its toughness rather than its power.
var DoranTheSiegeTower = newDoranTheSiegeTower

func newDoranTheSiegeTower() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Black, color.Green),
		CardFace: game.CardFace{
			Name: "Doran, the Siege Tower",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
				cost.B,
				cost.G,
			}),
			Colors:     []color.Color{color.Black, color.Green, color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Treefolk, types.Shaman},
			Power:      opt.Val(game.PT{Value: 0}),
			Toughness:  opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectAssignCombatDamageUsingToughness,
							PermanentTypes: []types.Card{types.Creature},
						},
					},
				},
			},
			OracleText: `
			Each creature assigns combat damage equal to its toughness rather than its power.
		`,
		},
	}
}
