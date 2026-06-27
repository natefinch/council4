package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PowerDepot is the card definition for Power Depot.
//
// Type: Artifact Land
//
// Oracle text:
//
//	This land enters tapped.
//	{T}: Add {C}.
//	{T}: Add one mana of any color. Spend this mana only to cast artifact spells or activate abilities of artifacts.
//	Modular 1
var PowerDepot = newPowerDepot()

func newPowerDepot() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "Power Depot",
			Types: []types.Card{types.Artifact, types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.C),
				game.ManaAbility{
					AdditionalCosts: cost.Tap,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Choose{
									Choice: game.ResolutionChoice{
										Kind:   game.ResolutionChoiceMana,
										Prompt: "Choose a color",
										Colors: []mana.Color{mana.W, mana.U, mana.B, mana.R, mana.G},
									},
									PublishChoice: game.ChoiceKey("oracle-mana-color"),
								},
							},
							{
								Primitive: game.AddMana{
									Amount:     game.Fixed(1),
									ChoiceFrom: game.ChoiceKey("oracle-mana-color"),
									SpendRider: opt.Val(game.ManaSpendRider{
										Condition:   game.ManaSpendCastOrActivateArtifact,
										Restriction: game.ManaSpendRestrictedToCondition,
									}),
								},
							},
						},
					}.Ability(),
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentDied,
							Source:           game.TriggerSourceSelf,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
						},
					},
					Optional: true,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target artifact creature",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Artifact, types.Creature}}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.MoveCounters{
									Object:   game.TargetPermanentReference(0),
									Source:   game.CounterSourceSpec{Kind: game.CounterSourceSelf},
									AllKinds: true,
								},
							},
						},
					}.Ability(),
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This land enters tapped."),
				game.EntersWithCountersReplacement("This creature enters with a +1/+1 counter on it.", game.CounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 1}),
			},
			OracleText: `
			This land enters tapped.
			{T}: Add {C}.
			{T}: Add one mana of any color. Spend this mana only to cast artifact spells or activate abilities of artifacts.
			Modular 1
		`,
		},
	}
}
