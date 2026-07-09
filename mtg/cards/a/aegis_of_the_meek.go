package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// AegisOfTheMeek is the card definition for Aegis of the Meek.
//
// Type: Artifact
// Cost: {3}
//
// Oracle text:
//
//	{1}, {T}: Target 1/1 creature gets +1/+2 until end of turn.
var AegisOfTheMeek = newAegisOfTheMeek

func newAegisOfTheMeek() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Aegis of the Meek",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:            "{1}, {T}: Target 1/1 creature gets +1/+2 until end of turn.",
					ManaCost:        opt.Val(cost.Mana{cost.O(1)}),
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
			OracleText: `
			{1}, {T}: Target 1/1 creature gets +1/+2 until end of turn.
		`,
		},
	}
}
