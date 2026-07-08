package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SerpentOfTheEndlessSea is the card definition for Serpent of the Endless Sea.
var SerpentOfTheEndlessSea = newSerpentOfTheEndlessSea

func newSerpentOfTheEndlessSea() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Serpent of the Endless Sea",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.U,
			}),
			Colors:           []color.Color{color.Blue},
			Types:            []types.Card{types.Creature},
			Subtypes:         []types.Sub{types.Serpent},
			Power:            opt.Val(game.PT{IsStar: true}),
			Toughness:        opt.Val(game.PT{IsStar: true}),
			DynamicPower:     opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerSubtypeCount, Subtype: types.Island}),
			DynamicToughness: opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerSubtypeCount, Subtype: types.Island}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:                            game.RuleEffectCantAttack,
							AffectedSource:                  true,
							AttackDefenderControlsSelection: game.Selection{SubtypesAny: []types.Sub{types.Sub("Island")}},
						},
					},
				},
			},
			OracleText: `
			Serpent of the Endless Sea's power and toughness are each equal to the number of Islands you control.
			This creature can't attack unless defending player controls an Island.
		`,
		},
	}
}
