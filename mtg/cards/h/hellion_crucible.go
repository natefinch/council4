package h

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

// HellionCrucible is the card definition for Hellion Crucible.
//
// Type: Land
//
// Oracle text:
//
//	{T}: Add {C}.
//	{1}{R}, {T}: Put a pressure counter on this land.
//	{1}{R}, {T}, Remove two pressure counters from this land and sacrifice it: Create a 4/4 red Hellion creature token with haste. (It can attack and {T} as soon as it comes under your control.)
var HellionCrucible = newHellionCrucible()

func newHellionCrucible() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name:  "Hellion Crucible",
			Types: []types.Card{types.Land},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:            "{1}{R}, {T}: Put a pressure counter on this land.",
					ManaCost:        opt.Val(cost.Mana{cost.O(1), cost.R}),
					AdditionalCosts: cost.Tap,
					ZoneOfFunction:  zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddCounter{
									Amount:      game.Fixed(1),
									Object:      game.SourcePermanentReference(),
									CounterKind: counter.Pressure,
								},
							},
						},
					}.Ability(),
				},
				game.ActivatedAbility{
					Text:     "{1}{R}, {T}, Remove two pressure counters from this land and sacrifice it: Create a 4/4 red Hellion creature token with haste. (It can attack and {T} as soon as it comes under your control.)",
					ManaCost: opt.Val(cost.Mana{cost.O(1), cost.R}),
					AdditionalCosts: []cost.Additional{
						{
							Kind: cost.AdditionalTap,
						},
						{
							Kind:        cost.AdditionalRemoveCounter,
							Text:        "Remove two pressure counters from this land",
							Amount:      2,
							CounterKind: counter.Pressure,
						},
						{
							Kind:   cost.AdditionalSacrificeSource,
							Text:   "sacrifice it",
							Amount: 1,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(hellionCrucibleToken),
								},
							},
						},
					}.Ability(),
				},
			},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.C),
			},
			OracleText: `
			{T}: Add {C}.
			{1}{R}, {T}: Put a pressure counter on this land.
			{1}{R}, {T}, Remove two pressure counters from this land and sacrifice it: Create a 4/4 red Hellion creature token with haste. (It can attack and {T} as soon as it comes under your control.)
		`,
		},
	}
}

var hellionCrucibleToken = newHellionCrucibleToken()

func newHellionCrucibleToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Hellion",
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Hellion},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.HasteStaticBody,
			},
		},
	}
}
