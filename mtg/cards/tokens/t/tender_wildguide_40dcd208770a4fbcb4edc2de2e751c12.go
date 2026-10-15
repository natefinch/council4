package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// TenderWildguide is the card definition for Tender Wildguide.
//
// Type: Token Creature — Possum Druid
// Cost: {1}{G}
//
// Oracle text:
//   {T}: Add one mana of any color.
//   {T}: Put a +1/+1 counter on this creature.
//   (This token's mana cost is {1}{G}.)

// TenderWildguideToken40dcd208770a4fbcb4edc2de2e751c12 is the card definition for Tender Wildguide.
var TenderWildguideToken40dcd208770a4fbcb4edc2de2e751c12 = newTenderWildguideToken40dcd208770a4fbcb4edc2de2e751c12()

func newTenderWildguideToken40dcd208770a4fbcb4edc2de2e751c12() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Tender Wildguide",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Possum, types.Druid},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{T}: Put a +1/+1 counter on this creature.",
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.AddCounter{
											Amount:      game.Fixed(1),
											Object:      game.SourcePermanentReference(),
											CounterKind: counter.PlusOnePlusOne,
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
			},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.W, mana.U, mana.B, mana.R, mana.G),
			},
			OracleText: `
			{T}: Add one mana of any color.
			{T}: Put a +1/+1 counter on this creature.
			(This token's mana cost is {1}{G}.)
		`,
		},
	}
}
