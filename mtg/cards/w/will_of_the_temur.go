package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// WillOfTheTemur is the card definition for Will of the Temur.
//
// Type: Sorcery
// Cost: {5}{U}
//
// Oracle text:
//
//	Choose one. If you control a commander as you cast this spell, you may choose both instead.
//	• Create a token that's a copy of target permanent, except it's a 4/4 Dragon creature with flying in addition to its other types.
//	• Target player draws cards equal to the greatest mana value among permanents you control.
var WillOfTheTemur = newWillOfTheTemur()

func newWillOfTheTemur() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Will of the Temur",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Create a token that's a copy of target permanent, except it's a 4/4 Dragon creature with flying in addition to its other types.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target permanent",
								Allow:      game.TargetAllowPermanent,
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenCopyOf(game.TokenCopySpec{
										Source:       game.TokenCopySourceObject,
										Object:       game.TargetPermanentReference(0),
										SetPower:     opt.Val(game.PT{Value: 4}),
										SetToughness: opt.Val(game.PT{Value: 4}),
										AddTypes:     []types.Card{types.Creature},
										AddSubtypes:  []types.Sub{types.Dragon},
										AddKeywords:  []game.Keyword{game.Flying},
									}),
								},
							},
						},
					},
					game.Mode{
						Text: "Target player draws cards equal to the greatest mana value among permanents you control.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target player",
								Allow:      game.TargetAllowPlayer,
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Draw{
									Amount: game.Dynamic(game.DynamicAmount{
										Kind:       game.DynamicAmountGreatestManaValueInGroup,
										Multiplier: 1,
										Group:      game.BattlefieldGroup(game.Selection{Controller: game.ControllerYou}),
									}),
									Player: game.TargetPlayerReference(0),
								},
							},
						},
					},
				},
				MinModes:        1,
				MaxModes:        1,
				ModeChoiceBonus: game.ModeChoiceBonus{Condition: game.ModeChoiceConditionControlsCommander, AdditionalMaxModes: 1},
			}),
			OracleText: `
			Choose one. If you control a commander as you cast this spell, you may choose both instead.
			• Create a token that's a copy of target permanent, except it's a 4/4 Dragon creature with flying in addition to its other types.
			• Target player draws cards equal to the greatest mana value among permanents you control.
		`,
		},
	}
}
