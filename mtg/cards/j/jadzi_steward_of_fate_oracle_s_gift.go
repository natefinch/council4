package j

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// JadziStewardOfFate is the card definition for Jadzi, Steward of Fate // Oracle's Gift.
//
// Type: Legendary Creature — Human Wizard // Sorcery
// Cost: {2}{U} // {X}{X}{U}
// Face: Oracle's Gift — Sorcery ({X}{X}{U})
//
// Oracle text:
//
//	Jadzi enters prepared. (While it's prepared, you may cast a copy of its spell. Doing so unprepares it.)
//	When Jadzi enters, draw two cards, then discard two cards.
var JadziStewardOfFate = newJadziStewardOfFate

func newJadziStewardOfFate() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Jadzi, Steward of Fate",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
			}),
			Colors:         []color.Color{color.Blue},
			EntersPrepared: true,
			Supertypes:     []types.Super{types.Legendary},
			Types:          []types.Card{types.Creature},
			Subtypes:       []types.Sub{types.Human, types.Wizard},
			Power:          opt.Val(game.PT{Value: 2}),
			Toughness:      opt.Val(game.PT{Value: 4}),
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
								Primitive: game.Draw{
									Amount: game.Fixed(2),
									Player: game.ControllerReference(),
								},
							},
							{
								Primitive: game.Discard{
									Amount: game.Fixed(2),
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Jadzi enters prepared. (While it's prepared, you may cast a copy of its spell. Doing so unprepares it.)
			When Jadzi enters, draw two cards, then discard two cards.
		`,
		},
		Layout: game.LayoutPrepare,
		Alternate: opt.Val(game.CardFace{
			Name: "Oracle's Gift",
			ManaCost: opt.Val(cost.Mana{
				cost.X,
				cost.X,
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind: game.DynamicAmountX,
							}),
							Source: game.TokenDef(jadziStewardOfFateToken),
						},
					},
					{
						Primitive: game.AddCounter{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind: game.DynamicAmountX,
							}),
							Group:       game.BattlefieldGroup(game.Selection{SubtypesAny: []types.Sub{types.Sub("Fractal")}, Controller: game.ControllerYou}),
							CounterKind: counter.PlusOnePlusOne,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create X 0/0 green and blue Fractal creature tokens, then put X +1/+1 counters on each Fractal you control.
		`,
		}),
	}
}

var jadziStewardOfFateToken = newJadziStewardOfFateToken()

func newJadziStewardOfFateToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Fractal",
			Colors:    []color.Color{color.Green, color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Fractal},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 0}),
		},
	}
}
