package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SolKanarTheSwampKing is the card definition for Sol'kanar the Swamp King.
//
// Type: Legendary Creature — Demon
// Cost: {2}{U}{B}{R}
//
// Oracle text:
//
//	Swampwalk (This creature can't be blocked as long as defending player controls a Swamp.)
//	Whenever a player casts a black spell, you gain 1 life.
var SolKanarTheSwampKing = newSolKanarTheSwampKing

func newSolKanarTheSwampKing() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black, color.Red),
		CardFace: game.CardFace{
			Name: "Sol'kanar the Swamp King",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
				cost.B,
				cost.R,
			}),
			Colors:     []color.Color{color.Black, color.Red, color.Blue},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Demon},
			Power:      opt.Val(game.PT{Value: 5}),
			Toughness:  opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.SwampwalkStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever a player casts a black spell, you gain 1 life.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:           game.EventSpellCast,
							CardSelection:   game.Selection{ColorsAny: []color.Color{color.Black}},
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.GainLife{
											Amount: game.Fixed(1),
											Player: game.ControllerReference(),
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
			Swampwalk (This creature can't be blocked as long as defending player controls a Swamp.)
			Whenever a player casts a black spell, you gain 1 life.
		`,
		},
	}
}
