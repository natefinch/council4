package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// WaywardSwordtooth is the card definition for Wayward Swordtooth.
//
// Type: Creature — Dinosaur
// Cost: {2}{G}
//
// Oracle text:
//
//	Ascend (If you control ten or more permanents, you get the city's blessing for the rest of the game.)
//	You may play an additional land on each of your turns.
//	This creature can't attack or block unless you have the city's blessing.
var WaywardSwordtooth = newWaywardSwordtooth

func newWaywardSwordtooth() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Wayward Swordtooth",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Dinosaur},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.AscendStaticBody,
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:                game.RuleEffectAdditionalLandPlays,
							AffectedPlayer:      game.PlayerYou,
							AdditionalLandPlays: 1,
						},
					},
				},
				game.StaticAbility{
					Condition: opt.Val(game.Condition{
						Negate:                    true,
						ControllerHasCityBlessing: true,
					}),
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCantAttack,
							AffectedSource: true,
						},
						game.RuleEffect{
							Kind:           game.RuleEffectCantBlock,
							AffectedSource: true,
						},
					},
				},
			},
			OracleText: `
			Ascend (If you control ten or more permanents, you get the city's blessing for the rest of the game.)
			You may play an additional land on each of your turns.
			This creature can't attack or block unless you have the city's blessing.
		`,
		},
	}
}
