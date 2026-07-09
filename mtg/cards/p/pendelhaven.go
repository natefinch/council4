package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Pendelhaven is the card definition for Pendelhaven.
//
// Type: Legendary Land
//
// Oracle text:
//
//	{T}: Add {G}.
//	{T}: Target 1/1 creature gets +1/+2 until end of turn.
var Pendelhaven = newPendelhaven

func newPendelhaven() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name:       "Pendelhaven",
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Land},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:            "{T}: Target 1/1 creature gets +1/+2 until end of turn.",
					AdditionalCosts: cost.Tap,
					ZoneOfFunction:  zone.Battlefield,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target 1/1 creature",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Power: opt.Val(compare.Int{Op: compare.Equal, Value: 1}), Toughness: opt.Val(compare.Int{Op: compare.Equal, Value: 1})}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.ModifyPT{
									Object:         game.TargetPermanentReference(0),
									PowerDelta:     game.Fixed(1),
									ToughnessDelta: game.Fixed(2),
									Duration:       game.DurationUntilEndOfTurn,
								},
							},
						},
					}.Ability(),
				},
			},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.G),
			},
			OracleText: `
			{T}: Add {G}.
			{T}: Target 1/1 creature gets +1/+2 until end of turn.
		`,
		},
	}
}
