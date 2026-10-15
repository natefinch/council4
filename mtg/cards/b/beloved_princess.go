package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BelovedPrincess is the card definition for Beloved Princess.
//
// Type: Creature — Human Noble
// Cost: {W}
//
// Oracle text:
//
//	Lifelink
//	This creature can't be blocked by creatures with power 3 or greater.
var BelovedPrincess = newBelovedPrincess

func newBelovedPrincess() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Beloved Princess",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Noble},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.LifelinkStaticBody,
				game.StaticAbility{
					Text: "This creature can't be blocked by creatures with power 3 or greater.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCantBeBlockedByCreaturesWith,
							AffectedSource: true,
							Duration:       game.DurationPermanent,
							BlockerRestriction: game.BlockerRestriction{
								Kind:  game.BlockerRestrictionPowerGreaterOrEqual,
								Power: 3,
							},
						},
					},
				},
			},
			OracleText: `
			Lifelink
			This creature can't be blocked by creatures with power 3 or greater.
		`,
		},
	}
}
