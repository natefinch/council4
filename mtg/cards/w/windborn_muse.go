package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// WindbornMuse is the card definition for Windborn Muse.
//
// Type: Creature — Spirit
// Cost: {3}{W}
//
// Oracle text:
//
//	Flying
//	Creatures can't attack you unless their controller pays {2} for each creature they control that's attacking you.
var WindbornMuse = newWindbornMuse

func newWindbornMuse() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Windborn Muse",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Spirit},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.StaticAbility{
					Text: "Creatures can't attack you unless their controller pays {2} for each creature they control that's attacking you.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:             game.RuleEffectAttackTax,
							Duration:         game.DurationPermanent,
							AffectedPlayer:   game.PlayerYou,
							AttackTaxGeneric: 2,
						},
					},
				},
			},
			OracleText: `
			Flying
			Creatures can't attack you unless their controller pays {2} for each creature they control that's attacking you.
		`,
		},
	}
}
