package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// LosheelClockworkScholar is the card definition for Losheel, Clockwork Scholar.
//
// Type: Legendary Creature — Elephant Artificer
// Cost: {2}{W}
//
// Oracle text:
//
//	Prevent all combat damage that would be dealt to attacking artifact creatures you control.
//	Whenever one or more artifact creatures you control enter, draw a card. This ability triggers only once each turn.
var LosheelClockworkScholar = newLosheelClockworkScholar

func newLosheelClockworkScholar() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Losheel, Clockwork Scholar",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
			}),
			Colors:     []color.Color{color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Elephant, types.Artificer},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 4}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentEnteredBattlefield,
							Controller:       game.TriggerControllerYou,
							OneOrMore:        true,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Artifact, types.Creature}},
						},
					},
					MaxTriggersPerTurn: 1,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Draw{
									Amount: game.Fixed(1),
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.CombatDamagePreventionToGroupReplacement("Prevent all combat damage that would be dealt to attacking artifact creatures you control.", game.Selection{RequiredTypes: []types.Card{types.Artifact, types.Creature}, Controller: game.ControllerYou, CombatState: game.CombatStateAttacking}),
			},
			OracleText: `
			Prevent all combat damage that would be dealt to attacking artifact creatures you control.
			Whenever one or more artifact creatures you control enter, draw a card. This ability triggers only once each turn.
		`,
		},
	}
}
