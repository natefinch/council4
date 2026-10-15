package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DwarvenMine is the card definition for Dwarven Mine.
//
// Type: Land — Mountain
//
// Oracle text:
//
//	({T}: Add {R}.)
//	This land enters tapped unless you control three or more other Mountains.
//	When this land enters untapped, create a 1/1 red Dwarf creature token.
var DwarvenMine = newDwarvenMine

func newDwarvenMine() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name:     "Dwarven Mine",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Mountain},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.R),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When this land enters untapped, create a 1/1 red Dwarf creature token.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentEnteredBattlefield,
							Source:           game.TriggerSourceSelf,
							SubjectSelection: game.Selection{Tapped: game.TriFalse},
							StackObjectKind:  game.StackSpell,
							CounterKind:      counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Fixed(1),
											Source: game.TokenDef(dwarvenMineToken),
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
				game.EntersTappedIfReplacement("This land enters tapped unless you control three or more other Mountains.", &game.Condition{
					Negate: true,
					ControlsMatching: opt.Val(game.SelectionCount{
						Selection: game.Selection{SubtypesAny: []types.Sub{types.Sub("Mountain")}, ExcludeSource: true},
						MinCount:  3,
					}),
				}),
			},
			OracleText: `
			({T}: Add {R}.)
			This land enters tapped unless you control three or more other Mountains.
			When this land enters untapped, create a 1/1 red Dwarf creature token.
		`,
		},
	}
}

var dwarvenMineToken = newDwarvenMineToken()

func newDwarvenMineToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Dwarf",
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Dwarf},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
