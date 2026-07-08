package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RushOfBattle is the card definition for Rush of Battle.
//
// Type: Sorcery
// Cost: {3}{W}
//
// Oracle text:
//
//	Creatures you control get +2/+1 until end of turn. Warrior creatures you control gain lifelink until end of turn. (Damage dealt by those Warriors also causes their controller to gain that much life.)
var RushOfBattle = newRushOfBattle

func newRushOfBattle() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Rush of Battle",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.ApplyContinuous{
							ContinuousEffects: []game.ContinuousEffect{
								game.ContinuousEffect{
									Layer:          game.LayerPowerToughnessModify,
									Group:          game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
									PowerDelta:     2,
									ToughnessDelta: 1,
								},
							},
							Duration: game.DurationUntilEndOfTurn,
						},
					},
					{
						Primitive: game.ApplyContinuous{
							ContinuousEffects: []game.ContinuousEffect{
								game.ContinuousEffect{
									Layer: game.LayerAbility,
									Group: game.BattlefieldGroup(game.Selection{SubtypesAny: []types.Sub{types.Sub("Warrior")}, Controller: game.ControllerYou}),
									AddKeywords: []game.Keyword{
										game.Lifelink,
									},
								},
							},
							Duration: game.DurationUntilEndOfTurn,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Creatures you control get +2/+1 until end of turn. Warrior creatures you control gain lifelink until end of turn. (Damage dealt by those Warriors also causes their controller to gain that much life.)
		`,
		},
	}
}
