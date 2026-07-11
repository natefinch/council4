package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ScurryOfSquirrels is the card definition for Scurry of Squirrels.
//
// Type: Creature — Squirrel Scout
// Cost: {2}{G}
//
// Oracle text:
//
//	Myriad, myriad (Whenever this creature attacks, for each opponent other than defending player, you may create a token that's a copy of this creature that's tapped and attacking that player or a planeswalker they control. Then do it again. Exile the tokens at end of combat.)
//	Whenever this creature deals combat damage to a player, put a +1/+1 counter on target creature you control.
var ScurryOfSquirrels = newScurryOfSquirrels

func newScurryOfSquirrels() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Scurry of Squirrels",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Squirrel, types.Scout},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.MyriadTriggeredBody,
				game.MyriadTriggeredBody,
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:               game.EventDamageDealt,
							Source:              game.TriggerSourceSelf,
							Subject:             game.TriggerSubjectDamageSource,
							RequireCombatDamage: true,
							DamageRecipient:     game.DamageRecipientPlayer,
						},
					},
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target creature you control",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerYou}),
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
			OracleText: `
			Myriad, myriad (Whenever this creature attacks, for each opponent other than defending player, you may create a token that's a copy of this creature that's tapped and attacking that player or a planeswalker they control. Then do it again. Exile the tokens at end of combat.)
			Whenever this creature deals combat damage to a player, put a +1/+1 counter on target creature you control.
		`,
		},
	}
}
