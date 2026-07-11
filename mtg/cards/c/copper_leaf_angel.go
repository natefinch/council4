package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// CopperLeafAngel is the card definition for Copper-Leaf Angel.
//
// Type: Artifact Creature — Angel
// Cost: {5}
//
// Oracle text:
//
//	Flying
//	{T}, Sacrifice X lands: Put X +1/+1 counters on this creature.
var CopperLeafAngel = newCopperLeafAngel

func newCopperLeafAngel() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Copper-Leaf Angel",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Angel},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{T}, Sacrifice X lands: Put X +1/+1 counters on this creature.",
					AdditionalCosts: []cost.Additional{
						{
							Kind: cost.AdditionalTap,
						},
						{
							Kind:               cost.AdditionalSacrifice,
							Text:               "Sacrifice X lands",
							AmountFromX:        true,
							MatchPermanentType: true,
							PermanentType:      types.Land,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddCounter{
									Amount: game.Dynamic(game.DynamicAmount{
										Kind: game.DynamicAmountX,
									}),
									Object:      game.SourcePermanentReference(),
									CounterKind: counter.PlusOnePlusOne,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Flying
			{T}, Sacrifice X lands: Put X +1/+1 counters on this creature.
		`,
		},
	}
}
