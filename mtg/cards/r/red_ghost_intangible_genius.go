package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RedGhostIntangibleGenius is the card definition for Red Ghost, Intangible Genius.
//
// Type: Legendary Creature — Human Scientist Villain
// Cost: {2}{U}{R}
//
// Oracle text:
//
//	Ward {2}
//	Red Ghost can't be blocked.
//	Whenever you draw your second card each turn, create a 3/3 red Ape Villain creature token with haste.
var RedGhostIntangibleGenius = newRedGhostIntangibleGenius

func newRedGhostIntangibleGenius() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Red),
		CardFace: game.CardFace{
			Name: "Red Ghost, Intangible Genius",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
				cost.R,
			}),
			Colors:     []color.Color{color.Red, color.Blue},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Scientist, types.Villain},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.WardStaticAbility(cost.Mana{cost.O(2)}),
				game.CantBeBlockedStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever you draw your second card each turn, create a 3/3 red Ape Villain creature token with haste.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                      game.EventCardDrawn,
							Player:                     game.TriggerPlayerYou,
							StackObjectKind:            game.StackSpell,
							PlayerEventOrdinalThisTurn: 2,
							CounterKind:                counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Fixed(1),
											Source: game.TokenDef(redGhostIntangibleGeniusToken),
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
			Ward {2}
			Red Ghost can't be blocked.
			Whenever you draw your second card each turn, create a 3/3 red Ape Villain creature token with haste.
		`,
		},
	}
}

var redGhostIntangibleGeniusToken = newRedGhostIntangibleGeniusToken()

func newRedGhostIntangibleGeniusToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Ape Villain",
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Ape, types.Villain},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.HasteStaticBody,
			},
		},
	}
}
