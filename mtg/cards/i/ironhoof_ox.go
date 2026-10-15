package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// IronhoofOx is the card definition for Ironhoof Ox.
//
// Type: Creature — Ox
// Cost: {3}{G}{G}
//
// Oracle text:
//
//	This creature can't be blocked by more than one creature.
var IronhoofOx = newIronhoofOx

func newIronhoofOx() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Ironhoof Ox",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Ox},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "This creature can't be blocked by more than one creature.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCantBeBlockedByMoreThanOne,
							AffectedSource: true,
							Duration:       game.DurationPermanent,
						},
					},
				},
			},
			OracleText: `
			This creature can't be blocked by more than one creature.
		`,
		},
	}
}
