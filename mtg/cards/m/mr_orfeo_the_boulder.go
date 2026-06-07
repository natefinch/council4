package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MrOrfeoTheBoulder is the card definition for Mr. Orfeo, the Boulder.
//
// Type: Legendary Creature — Rhino Warrior
// Cost: {1}{B}{R}{G}
//
// Oracle text:
//
//	Whenever you attack, double target creature's power until end of turn.
var MrOrfeoTheBoulder = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Black, color.Green, color.Red),
	CardFace: game.CardFace{
		Name: "Mr. Orfeo, the Boulder",
		ManaCost: opt.Val(cost.Mana{
			cost.O(1),
			cost.B,
			cost.R,
			cost.G,
		}),
		Colors:     []color.Color{color.Black, color.Green, color.Red},
		Supertypes: []types.Super{types.Legendary},
		Types:      []types.Card{types.Creature},
		Subtypes:   []types.Sub{types.Rhino, types.Warrior},
		Power:      opt.Val(game.PT{Value: 2}),
		Toughness:  opt.Val(game.PT{Value: 4}),
		OracleText: `
			Whenever you attack, double target creature's power until end of turn.
		`,
		TriggeredAbilities: []game.TriggeredAbility{
			{
				Text: `
					Whenever you attack, double target creature's power until end of turn.
				`,
				Trigger: game.TriggerCondition{
					Type: game.TriggerWhenever,
					Pattern: game.TriggerPattern{
						Event:      game.EventAttackerDeclared,
						Controller: game.TriggerControllerYou,
						OneOrMore:  true,
					},
				},
				Content: game.Mode{
					Targets: []game.TargetSpec{
						{
							MinTargets: 1,
							MaxTargets: 1,
							Constraint: "creature",
							Allow:      game.TargetAllowPermanent,
							Predicate: game.TargetPredicate{
								PermanentTypes: []types.Card{
									types.Creature,
								},
							},
						},
					},
					Sequence: []game.Instruction{
						{
							Primitive: game.ModifyPT{
								TargetIndex: 0,
								PowerDelta: game.Dynamic(game.DynamicAmount{
									Kind: game.DynamicAmountObjectPower,
									Object: game.ObjectReference{
										Kind:        game.ObjectReferenceTargetPermanent,
										TargetIndex: 0,
									},
								}),
								Duration: game.DurationUntilEndOfTurn,
							},
						},
					},
				}.Ability(),
			},
		},
	},
}
