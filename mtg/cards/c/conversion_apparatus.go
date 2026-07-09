package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ConversionApparatus is the card definition for Conversion Apparatus.
//
// Type: Artifact
// Cost: {3}
//
// Oracle text:
//
//	{T}: Add {C}.
//	{3}, {T}: You get {E}{E}{E} (three energy counters).
//	{T}, Pay {E}{E}{E}: Add three mana in any combination of colors.
var ConversionApparatus = newConversionApparatus

func newConversionApparatus() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Conversion Apparatus",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:            "{3}, {T}: You get {E}{E}{E} (three energy counters).",
					ManaCost:        opt.Val(cost.Mana{cost.O(3)}),
					AdditionalCosts: cost.Tap,
					ZoneOfFunction:  zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddPlayerCounter{
									Amount:      game.Fixed(3),
									Player:      game.ControllerReference(),
									CounterKind: counter.Energy,
								},
							},
						},
					}.Ability(),
				},
			},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.C),
				game.ManaAbility{
					AdditionalCosts: []cost.Additional{
						{
							Kind: cost.AdditionalTap,
						},
						{
							Kind:   cost.AdditionalEnergy,
							Text:   "Pay {E}{E}{E}",
							Amount: 3,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddMana{
									Amount:            game.Fixed(3),
									CombinationColors: []mana.Color{mana.W, mana.U, mana.B, mana.R, mana.G},
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			{T}: Add {C}.
			{3}, {T}: You get {E}{E}{E} (three energy counters).
			{T}, Pay {E}{E}{E}: Add three mana in any combination of colors.
		`,
		},
	}
}
