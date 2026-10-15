package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FleetFootedMonk is the card definition for Fleet-Footed Monk.
//
// Type: Creature — Human Monk
// Cost: {1}{W}
//
// Oracle text:
//
//	This creature can't be blocked by creatures with power 2 or greater.
var FleetFootedMonk = newFleetFootedMonk

func newFleetFootedMonk() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Fleet-Footed Monk",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Monk},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "This creature can't be blocked by creatures with power 2 or greater.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCantBeBlockedByCreaturesWith,
							AffectedSource: true,
							Duration:       game.DurationPermanent,
							BlockerRestriction: game.BlockerRestriction{
								Kind:  game.BlockerRestrictionPowerGreaterOrEqual,
								Power: 2,
							},
						},
					},
				},
			},
			OracleText: `
			This creature can't be blocked by creatures with power 2 or greater.
		`,
		},
	}
}
