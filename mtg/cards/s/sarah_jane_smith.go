package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SarahJaneSmith is the card definition for Sarah Jane Smith.
//
// Type: Legendary Creature — Human Detective
// Cost: {1}{W}
//
// Oracle text:
//
//	Whenever you cast a historic spell, investigate. This ability triggers only once each turn. (Artifacts, legendaries, and Sagas are historic. To investigate, create a Clue token. It's an artifact with "{2}, Sacrifice this token: Draw a card.")
//	Doctor's companion (You can have two commanders if the other is the Doctor.)
var SarahJaneSmith = newSarahJaneSmith

func newSarahJaneSmith() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Sarah Jane Smith",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors:     []color.Color{color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Detective},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.CompanionStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever you cast a historic spell, investigate. This ability triggers only once each turn. (Artifacts, legendaries, and Sagas are historic. To investigate, create a Clue token. It's an artifact with \"{2}, Sacrifice this token: Draw a card.\")",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:           game.EventSpellCast,
							Controller:      game.TriggerControllerYou,
							StackObjectKind: game.StackSpell,
							RequireHistoric: true,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					MaxTriggersPerTurn: 1,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Investigate{
											Amount: game.Fixed(1),
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
			Whenever you cast a historic spell, investigate. This ability triggers only once each turn. (Artifacts, legendaries, and Sagas are historic. To investigate, create a Clue token. It's an artifact with "{2}, Sacrifice this token: Draw a card.")
			Doctor's companion (You can have two commanders if the other is the Doctor.)
		`,
		},
	}
}
