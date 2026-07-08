package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// EsperBattlemage is the card definition for Esper Battlemage.
//
// Type: Artifact Creature — Human Wizard
// Cost: {2}{U}
//
// Oracle text:
//
//	{W}, {T}: Prevent the next 2 damage that would be dealt to you this turn.
//	{B}, {T}: Target creature gets -1/-1 until end of turn.
var EsperBattlemage = newEsperBattlemage

func newEsperBattlemage() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue, color.Black),
		CardFace: game.CardFace{
			Name: "Esper Battlemage",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Wizard},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:            "{W}, {T}: Prevent the next 2 damage that would be dealt to you this turn.",
					ManaCost:        opt.Val(cost.Mana{cost.W}),
					AdditionalCosts: cost.Tap,
					ZoneOfFunction:  zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.PreventDamage{
									Player: game.ControllerReference(),
									Amount: game.Fixed(2),
								},
							},
						},
					}.Ability(),
				},
				game.ActivatedAbility{
					Text:            "{B}, {T}: Target creature gets -1/-1 until end of turn.",
					ManaCost:        opt.Val(cost.Mana{cost.B}),
					AdditionalCosts: cost.Tap,
					ZoneOfFunction:  zone.Battlefield,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target creature",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.ModifyPT{
									Object:         game.TargetPermanentReference(0),
									PowerDelta:     game.Fixed(-1),
									ToughnessDelta: game.Fixed(-1),
									Duration:       game.DurationUntilEndOfTurn,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			{W}, {T}: Prevent the next 2 damage that would be dealt to you this turn.
			{B}, {T}: Target creature gets -1/-1 until end of turn.
		`,
		},
	}
}
