package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CourtCleric is the card definition for Court Cleric.
//
// Type: Creature — Human Cleric
// Cost: {W}
//
// Oracle text:
//
//	Lifelink (Damage dealt by this creature also causes you to gain that much life.)
//	This creature gets +1/+1 as long as you control an Ajani planeswalker.
var CourtCleric = newCourtCleric

func newCourtCleric() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Court Cleric",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Cleric},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.LifelinkStaticBody,
				game.StaticAbility{
					Text: "This creature gets +1/+1 as long as you control an Ajani planeswalker.",
					Condition: opt.Val(game.Condition{
						Text: "as long as you control an Ajani planeswalker",
						ControlsMatching: opt.Val(game.SelectionCount{
							Selection: game.Selection{RequiredTypes: []types.Card{types.Planeswalker}, SubtypesAny: []types.Sub{types.Ajani}},
						}),
						SourceCounterKind: counter.PlusOnePlusOne,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     1,
							ToughnessDelta: 1,
						},
					},
				},
			},
			OracleText: `
			Lifelink (Damage dealt by this creature also causes you to gain that much life.)
			This creature gets +1/+1 as long as you control an Ajani planeswalker.
		`,
		},
	}
}
