package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// StellaLeeWildCard is the card definition for Stella Lee, Wild Card.
//
// Type: Legendary Creature — Human Rogue
// Cost: {1}{U}{R}
//
// Oracle text:
//
//	Whenever you cast your second spell each turn, exile the top card of your library. Until the end of your next turn, you may play that card.
//	{T}: Copy target instant or sorcery spell you control. You may choose new targets for the copy. Activate only if you've cast three or more spells this turn.
var StellaLeeWildCard = newStellaLeeWildCard

func newStellaLeeWildCard() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Red),
		CardFace: game.CardFace{
			Name: "Stella Lee, Wild Card",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
				cost.R,
			}),
			Colors:     []color.Color{color.Red, color.Blue},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Rogue},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 4}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:            "{T}: Copy target instant or sorcery spell you control. You may choose new targets for the copy. Activate only if you've cast three or more spells this turn.",
					AdditionalCosts: cost.Tap,
					ZoneOfFunction:  zone.Battlefield,
					ActivationCondition: opt.Val(game.Condition{
						EventHistory: opt.Val(game.EventHistoryCondition{Pattern: game.TriggerPattern{
							Event:      game.EventSpellCast,
							Controller: game.TriggerControllerYou,
						}, Window: game.EventHistoryCurrentTurn, MinCount: 3}),
					}),
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target instant or sorcery spell you control",
								Allow:      game.TargetAllowStackObject,
								Predicate: game.TargetPredicate{
									SpellCardTypesAny: []types.Card{types.Instant, types.Sorcery},
									StackObjectKinds:  []game.StackObjectKind{game.StackSpell},
									Controller:        game.ControllerYou,
								},
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.CopyStackObject{
									Object:              game.TargetStackObjectReference(0),
									MayChooseNewTargets: true,
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
							Event:                      game.EventSpellCast,
							Controller:                 game.TriggerControllerYou,
							PlayerEventOrdinalThisTurn: 2,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ImpulseExile{
									Player:   game.ControllerReference(),
									Amount:   game.Fixed(1),
									Duration: game.DurationUntilEndOfYourNextTurn,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Whenever you cast your second spell each turn, exile the top card of your library. Until the end of your next turn, you may play that card.
			{T}: Copy target instant or sorcery spell you control. You may choose new targets for the copy. Activate only if you've cast three or more spells this turn.
		`,
		},
	}
}
