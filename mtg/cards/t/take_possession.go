package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TakePossession is the card definition for Take Possession.
//
// Type: Enchantment — Aura
// Cost: {5}{U}{U}
//
// Oracle text:
//
//	Split second (As long as this spell is on the stack, players can't cast spells or activate abilities that aren't mana abilities.)
//	Enchant permanent
//	You control enchanted permanent.
var TakePossession = newTakePossession

func newTakePossession() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Take Possession",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.U,
				cost.U,
			}),
			Colors:   []color.Color{color.Blue},
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Aura},
			StaticAbilities: []game.StaticAbility{
				game.SplitSecondStaticBody,
				game.EnchantStaticAbility(&game.TargetSpec{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "permanent",
					Allow:      game.TargetAllowPermanent,
				}),
				game.StaticAbility{
					Text: "You control enchanted permanent.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:      game.DurationPermanent,
							Group:         game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:         game.LayerControl,
							NewController: opt.Val(game.Player1),
						},
					},
				},
			},
			OracleText: `
			Split second (As long as this spell is on the stack, players can't cast spells or activate abilities that aren't mana abilities.)
			Enchant permanent
			You control enchanted permanent.
		`,
		},
	}
}
