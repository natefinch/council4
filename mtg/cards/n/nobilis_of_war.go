package n

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// NobilisOfWar is the card definition for Nobilis of War.
//
// Type: Creature — Spirit Avatar
// Cost: {R/W}{R/W}{R/W}{R/W}{R/W}
//
// Oracle text:
//
//	Flying
//	Attacking creatures you control get +2/+0.
var NobilisOfWar = newNobilisOfWar

func newNobilisOfWar() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Red),
		CardFace: game.CardFace{
			Name: "Nobilis of War",
			ManaCost: opt.Val(cost.Mana{
				cost.HybridMana(mana.R, mana.W),
				cost.HybridMana(mana.R, mana.W),
				cost.HybridMana(mana.R, mana.W),
				cost.HybridMana(mana.R, mana.W),
				cost.HybridMana(mana.R, mana.W),
			}),
			Colors:    []color.Color{color.Red, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Spirit, types.Avatar},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.StaticAbility{
					Text: "Attacking creatures you control get +2/+0.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:   game.DurationPermanent,
							Group:      game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}, CombatState: game.CombatStateAttacking}),
							Layer:      game.LayerPowerToughnessModify,
							PowerDelta: 2,
						},
					},
				},
			},
			OracleText: `
			Flying
			Attacking creatures you control get +2/+0.
		`,
		},
	}
}
