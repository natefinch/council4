package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DauthiHorror is the card definition for Dauthi Horror.
//
// Type: Creature — Dauthi Horror
// Cost: {1}{B}
//
// Oracle text:
//
//	Shadow (This creature can block or be blocked by only creatures with shadow.)
//	This creature can't be blocked by white creatures.
var DauthiHorror = newDauthiHorror

func newDauthiHorror() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Dauthi Horror",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Dauthi, types.Horror},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.ShadowStaticBody,
				game.StaticAbility{
					Text: "This creature can't be blocked by white creatures.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCantBeBlockedByCreaturesWith,
							AffectedSource: true,
							Duration:       game.DurationPermanent,
							BlockerRestriction: game.BlockerRestriction{
								Kind:  game.BlockerRestrictionColor,
								Color: color.White,
							},
						},
					},
				},
			},
			OracleText: `
			Shadow (This creature can block or be blocked by only creatures with shadow.)
			This creature can't be blocked by white creatures.
		`,
		},
	}
}
