package u

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// UrborgSkeleton is the card definition for Urborg Skeleton.
//
// Type: Creature — Skeleton
// Cost: {B}
//
// Oracle text:
//
//	Kicker {3} (You may pay an additional {3} as you cast this spell.)
//	{B}: Regenerate this creature.
//	If this creature was kicked, it enters with a +1/+1 counter on it.
var UrborgSkeleton = newUrborgSkeleton()

func newUrborgSkeleton() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Urborg Skeleton",
			ManaCost: opt.Val(cost.Mana{
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Skeleton},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					KeywordAbilities: []game.KeywordAbility{
						game.KickerKeyword{Cost: cost.Mana{cost.O(3)}},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:           "{B}: Regenerate this creature.",
					ManaCost:       opt.Val(cost.Mana{cost.B}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Regenerate{
									Object: game.SourcePermanentReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersIfReplacement("If this creature was kicked, it enters with a +1/+1 counter on it.", &game.Condition{
					EventPermanentWasKicked: true,
				}, game.CounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 1}),
			},
			OracleText: `
			Kicker {3} (You may pay an additional {3} as you cast this spell.)
			{B}: Regenerate this creature.
			If this creature was kicked, it enters with a +1/+1 counter on it.
		`,
		},
	}
}
