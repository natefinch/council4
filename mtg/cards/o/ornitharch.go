package o

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Ornitharch is the card definition for Ornitharch.
//
// Type: Creature — Archon
// Cost: {3}{W}{W}
//
// Oracle text:
//
//	Flying
//	Tribute 2 (As this creature enters, an opponent of your choice may put two +1/+1 counters on it.)
//	When this creature enters, if tribute wasn't paid, create two 1/1 white Bird creature tokens with flying.
var Ornitharch = newOrnitharch

func newOrnitharch() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Ornitharch",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Archon},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When this creature enters, if tribute wasn't paid, create two 1/1 white Bird creature tokens with flying.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:           game.EventPermanentEnteredBattlefield,
							Source:          game.TriggerSourceSelf,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
						InterveningIf: "if tribute wasn't paid",
						InterveningCondition: opt.Val(game.Condition{
							Text:                 "if tribute wasn't paid",
							SourceCounterKind:    counter.PlusOnePlusOne,
							SourceTributeNotPaid: true,
						}),
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Fixed(2),
											Source: game.TokenDef(ornitharchToken),
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
			ReplacementAbilities: []game.ReplacementAbility{
				game.TributeReplacement("As this creature enters, an opponent of your choice may put 2 +1/+1 counters on it.", 2),
			},
			OracleText: `
			Flying
			Tribute 2 (As this creature enters, an opponent of your choice may put two +1/+1 counters on it.)
			When this creature enters, if tribute wasn't paid, create two 1/1 white Bird creature tokens with flying.
		`,
		},
	}
}

var ornitharchToken = newOrnitharchToken()

func newOrnitharchToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Bird",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Bird},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
		},
	}
}
