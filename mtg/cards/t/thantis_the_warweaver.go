package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ThantisTheWarweaver is the card definition for Thantis, the Warweaver.
//
// Type: Legendary Creature — Spider
// Cost: {3}{B}{R}{G}
//
// Oracle text:
//
//	Reach, vigilance
//	All creatures attack each combat if able.
//	Whenever a creature attacks you or a planeswalker you control, put a +1/+1 counter on Thantis.
var ThantisTheWarweaver = newThantisTheWarweaver

func newThantisTheWarweaver() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Thantis, the Warweaver",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
				cost.R,
				cost.G,
			}),
			Colors:     []color.Color{color.Black, color.Green, color.Red},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Spider},
			Power:      opt.Val(game.PT{Value: 5}),
			Toughness:  opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.ReachStaticBody,
				game.VigilanceStaticBody,
				game.StaticAbility{
					Text: "All creatures attack each combat if able.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectMustAttack,
							Duration:       game.DurationPermanent,
							PermanentTypes: []types.Card{types.Creature},
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever a creature attacks you or a planeswalker you control, put a +1/+1 counter on Thantis.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                    game.EventAttackerDeclared,
							Player:                   game.TriggerPlayerYou,
							SubjectSelection:         game.Selection{RequiredTypes: []types.Card{types.Creature}},
							StackObjectKind:          game.StackSpell,
							AttackRecipient:          game.AttackRecipientPlayer | game.AttackRecipientPlaneswalker,
							AttackRecipientSelection: game.Selection{RequiredTypes: []types.Card{types.Planeswalker}, Controller: game.ControllerYou},
							CounterKind:              counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.AddCounter{
											Amount:      game.Fixed(1),
											Object:      game.SourcePermanentReference(),
											CounterKind: counter.PlusOnePlusOne,
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
			},
			OracleText: `
			Reach, vigilance
			All creatures attack each combat if able.
			Whenever a creature attacks you or a planeswalker you control, put a +1/+1 counter on Thantis.
		`,
		},
	}
}
