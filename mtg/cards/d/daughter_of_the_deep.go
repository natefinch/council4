package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// DaughterOfTheDeep is the card definition for Daughter of the Deep.
//
// Type: Creature — Merfolk Noble
// Cost: {1}{U}
//
// Oracle text:
//
//	Whenever you draw your second card each turn, create a 1/1 blue Merfolk creature token.
//	{U}, {T}: Target Merfolk can't be blocked this turn.
var DaughterOfTheDeep = newDaughterOfTheDeep()

func newDaughterOfTheDeep() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Daughter of the Deep",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Merfolk, types.Noble},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:            "{U}, {T}: Target Merfolk can't be blocked this turn.",
					ManaCost:        opt.Val(cost.Mana{cost.U}),
					AdditionalCosts: cost.Tap,
					ZoneOfFunction:  zone.Battlefield,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target Merfolk",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{SubtypesAny: []types.Sub{types.Sub("Merfolk")}}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyRule{
									Object: opt.Val(game.TargetPermanentReference(0)),
									RuleEffects: []game.RuleEffect{
										game.RuleEffect{
											Kind: game.RuleEffectCantBeBlocked,
										},
									},
									Duration: game.DurationThisTurn,
								},
							},
						},
					}.Ability(),
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                      game.EventCardDrawn,
							Player:                     game.TriggerPlayerYou,
							PlayerEventOrdinalThisTurn: 2,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(daughterOfTheDeepToken),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Whenever you draw your second card each turn, create a 1/1 blue Merfolk creature token.
			{U}, {T}: Target Merfolk can't be blocked this turn.
		`,
		},
	}
}

var daughterOfTheDeepToken = newDaughterOfTheDeepToken()

func newDaughterOfTheDeepToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Merfolk",
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Merfolk},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
