package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// AlexiZephyrMage is the card definition for Alexi, Zephyr Mage.
//
// Type: Legendary Creature — Human Spellshaper
// Cost: {3}{U}{U}
//
// Oracle text:
//
//	{X}{U}, {T}, Discard two cards: Return X target creatures to their owners' hands.
var AlexiZephyrMage = newAlexiZephyrMage

func newAlexiZephyrMage() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Alexi, Zephyr Mage",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
				cost.U,
			}),
			Colors:     []color.Color{color.Blue},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Spellshaper},
			Power:      opt.Val(game.PT{Value: 3}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{X}{U}, {T}, Discard two cards: Return X target creatures to their owners' hands.",
					ManaCost: opt.Val(cost.Mana{
						cost.X,
						cost.U,
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
						cost.Additional{
							Kind:        cost.AdditionalDiscard,
							Text:        "Discard two cards",
							Amount:      2,
							Source:      zone.Hand,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MaxTargets:   20,
										Constraint:   "x target creatures",
										Allow:        game.TargetAllowPermanent,
										Selection:    opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
										CountEqualsX: true,
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.MovePermanent{
											Object:      game.AllTargetPermanentsReference(0),
											Destination: zone.Hand,
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
			{X}{U}, {T}, Discard two cards: Return X target creatures to their owners' hands.
		`,
		},
	}
}
