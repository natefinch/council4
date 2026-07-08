package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ViashinoBranchrider is the card definition for Viashino Branchrider.
//
// Type: Creature — Lizard Warrior
// Cost: {R}
//
// Oracle text:
//
//	Kicker {2}{G} (You may pay an additional {2}{G} as you cast this spell.)
//	Haste
//	If this creature was kicked, it enters with two +1/+1 counters on it.
//	{2}{R}: This creature gets +2/+0 until end of turn.
var ViashinoBranchrider = newViashinoBranchrider

func newViashinoBranchrider() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Viashino Branchrider",
			ManaCost: opt.Val(cost.Mana{
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Lizard, types.Warrior},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					KeywordAbilities: []game.KeywordAbility{
						game.KickerKeyword{Cost: cost.Mana{cost.O(2), cost.G}},
					},
				},
				game.HasteStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:           "{2}{R}: This creature gets +2/+0 until end of turn.",
					ManaCost:       opt.Val(cost.Mana{cost.O(2), cost.R}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ModifyPT{
									Object:         game.SourcePermanentReference(),
									PowerDelta:     game.Fixed(2),
									ToughnessDelta: game.Fixed(0),
									Duration:       game.DurationUntilEndOfTurn,
								},
							},
						},
					}.Ability(),
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersIfReplacement("If this creature was kicked, it enters with two +1/+1 counters on it.", &game.Condition{
					EventPermanentWasKicked: true,
				}, game.CounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 2}),
			},
			OracleText: `
			Kicker {2}{G} (You may pay an additional {2}{G} as you cast this spell.)
			Haste
			If this creature was kicked, it enters with two +1/+1 counters on it.
			{2}{R}: This creature gets +2/+0 until end of turn.
		`,
		},
	}
}
