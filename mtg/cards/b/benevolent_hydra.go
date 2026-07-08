package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// BenevolentHydra is the card definition for Benevolent Hydra.
//
// Type: Creature — Hydra
// Cost: {X}{G}{G}
//
// Oracle text:
//
//	This creature enters with X +1/+1 counters on it.
//	If one or more +1/+1 counters would be put on another creature you control, that many plus one +1/+1 counters are put on it instead.
//	{T}, Remove a +1/+1 counter from this creature: Put a +1/+1 counter on another target creature you control.
var BenevolentHydra = newBenevolentHydra

func newBenevolentHydra() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Benevolent Hydra",
			ManaCost: opt.Val(cost.Mana{
				cost.X,
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Hydra},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{T}, Remove a +1/+1 counter from this creature: Put a +1/+1 counter on another target creature you control.",
					AdditionalCosts: []cost.Additional{
						{
							Kind: cost.AdditionalTap,
						},
						{
							Kind:        cost.AdditionalRemoveCounter,
							Text:        "Remove a +1/+1 counter from this creature",
							Amount:      1,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "another target creature you control",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerYou, ExcludeSource: true}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.AddCounter{
									Amount:      game.Fixed(1),
									Object:      game.TargetPermanentReference(0),
									CounterKind: counter.PlusOnePlusOne,
								},
							},
						},
					}.Ability(),
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersReplacement("This creature enters with X +1/+1 counters on it.", game.CounterPlacement{Kind: counter.PlusOnePlusOne, AmountFromX: true}),
				game.ControlledPermanentSelectionCounterKindPlacementReplacement("If one or more +1/+1 counters would be put on another creature you control, that many plus one +1/+1 counters are put on it instead.", 0, 1, counter.PlusOnePlusOne, game.Selection{RequiredTypesAny: []types.Card{types.Creature}, ExcludeSource: true}, game.TriggerControllerYou),
			},
			OracleText: `
			This creature enters with X +1/+1 counters on it.
			If one or more +1/+1 counters would be put on another creature you control, that many plus one +1/+1 counters are put on it instead.
			{T}, Remove a +1/+1 counter from this creature: Put a +1/+1 counter on another target creature you control.
		`,
		},
	}
}
