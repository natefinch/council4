package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MesmerizingBenthid is the card definition for Mesmerizing Benthid.
//
// Type: Creature — Octopus
// Cost: {3}{U}{U}
//
// Oracle text:
//
//	When this creature enters, create two 0/2 blue Illusion creature tokens with "Whenever this token blocks a creature, that creature doesn't untap during its controller's next untap step."
//	This creature has hexproof as long as you control an Illusion.
var MesmerizingBenthid = newMesmerizingBenthid

func newMesmerizingBenthid() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Mesmerizing Benthid",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Octopus},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Condition: opt.Val(game.Condition{
						ControlsMatching: opt.Val(game.SelectionCount{
							Selection: game.Selection{SubtypesAny: []types.Sub{types.Sub("Illusion")}},
						}),
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:          game.LayerAbility,
							AffectedSource: true,
							AddKeywords: []game.Keyword{
								game.Hexproof,
							},
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentEnteredBattlefield,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(2),
									Source: game.TokenDef(mesmerizingBenthidToken),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			When this creature enters, create two 0/2 blue Illusion creature tokens with "Whenever this token blocks a creature, that creature doesn't untap during its controller's next untap step."
			This creature has hexproof as long as you control an Illusion.
		`,
		},
	}
}

var mesmerizingBenthidToken = newMesmerizingBenthidToken()

func newMesmerizingBenthidToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Illusion",
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Illusion},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 2}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                   game.EventBlockerDeclared,
							Source:                  game.TriggerSourceSelf,
							RelatedSubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.SkipNextUntap{
									Object: game.EventRelatedPermanentReference(),
								},
							},
						},
					}.Ability(),
				},
			},
		},
	}
}
