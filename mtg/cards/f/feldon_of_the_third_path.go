package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// FeldonOfTheThirdPath is the card definition for Feldon of the Third Path.
//
// Type: Legendary Creature — Human Artificer
// Cost: {1}{R}{R}
//
// Oracle text:
//
//	{2}{R}, {T}: Create a token that's a copy of target creature card in your graveyard, except it's an artifact in addition to its other types. It gains haste. Sacrifice it at the beginning of the next end step.
var FeldonOfTheThirdPath = newFeldonOfTheThirdPath

func newFeldonOfTheThirdPath() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Feldon of the Third Path",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
				cost.R,
			}),
			Colors:     []color.Color{color.Red},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Artificer},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:            "{2}{R}, {T}: Create a token that's a copy of target creature card in your graveyard, except it's an artifact in addition to its other types. It gains haste. Sacrifice it at the beginning of the next end step.",
					ManaCost:        opt.Val(cost.Mana{cost.O(2), cost.R}),
					AdditionalCosts: cost.Tap,
					ZoneOfFunction:  zone.Battlefield,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target creature card in your graveyard",
								Allow:      game.TargetAllowCard,
								TargetZone: zone.Graveyard,
								Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenCopyOf(game.TokenCopySpec{
										Source:      game.TokenCopySourceObject,
										Object:      game.TargetCardReference(0),
										AddTypes:    []types.Card{types.Artifact},
										AddKeywords: []game.Keyword{game.Haste},
									}),
									PublishLinked: game.LinkedKey("delayed-sacrifice-1"),
								},
							},
							{
								Primitive: game.CreateDelayedTrigger{
									Trigger: game.DelayedTriggerDef{
										Timing:         game.DelayedAtBeginningOfNextEndStep,
										CapturedObject: opt.Val(game.LinkedObjectReference("delayed-sacrifice-1")),
										Content: game.Mode{
											Sequence: []game.Instruction{
												{
													Primitive: game.Sacrifice{
														Object: game.CapturedObjectReference(),
													},
												},
											},
										}.Ability(),
									},
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			{2}{R}, {T}: Create a token that's a copy of target creature card in your graveyard, except it's an artifact in addition to its other types. It gains haste. Sacrifice it at the beginning of the next end step.
		`,
		},
	}
}
