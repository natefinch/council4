package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BoneCairnButcher is the card definition for Bone-Cairn Butcher.
//
// Type: Creature — Demon
// Cost: {1}{R}{W}{B}
//
// Oracle text:
//
//	Mobilize 2 (Whenever this creature attacks, create two tapped and attacking 1/1 red Warrior creature tokens. Sacrifice them at the beginning of the next end step.)
//	Attacking tokens you control have deathtouch.
var BoneCairnButcher = newBoneCairnButcher

func newBoneCairnButcher() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Black, color.Red),
		CardFace: game.CardFace{
			Name: "Bone-Cairn Butcher",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
				cost.W,
				cost.B,
			}),
			Colors:    []color.Color{color.Black, color.Red, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Demon},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer: game.LayerAbility,
							Group: game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}, CombatState: game.CombatStateAttacking, TokenOnly: true}),
							AddKeywords: []game.Keyword{
								game.Deathtouch,
							},
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.MobilizeTriggeredBody(game.MobilizeAmount{Fixed: 2}),
			},
			OracleText: `
			Mobilize 2 (Whenever this creature attacks, create two tapped and attacking 1/1 red Warrior creature tokens. Sacrifice them at the beginning of the next end step.)
			Attacking tokens you control have deathtouch.
		`,
		},
	}
}
