package y

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// YouReConfrontedByRobbers is the card definition for You're Confronted by Robbers.
//
// Type: Instant
// Cost: {3}{W}
//
// Oracle text:
//
//	Choose one —
//	• Stall for Time — Tap up to three target creatures.
//	• Call for Aid — Create three 1/1 white Soldier creature tokens.
var YouReConfrontedByRobbers = newYouReConfrontedByRobbers

func newYouReConfrontedByRobbers() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "You're Confronted by Robbers",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Stall for Time — Tap up to three target creatures.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 0,
								MaxTargets: 3,
								Constraint: "up to three target creatures",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Tap{
									Object: game.TargetPermanentReference(0),
								},
							},
							{
								Primitive: game.Tap{
									Object: game.TargetPermanentReference(1),
								},
							},
							{
								Primitive: game.Tap{
									Object: game.TargetPermanentReference(2),
								},
							},
						},
					},
					game.Mode{
						Text: "Call for Aid — Create three 1/1 white Soldier creature tokens.",
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(3),
									Source: game.TokenDef(youReConfrontedByRobbersToken),
								},
							},
						},
					},
				},
				MinModes: 1,
				MaxModes: 1,
			}),
			OracleText: `
			Choose one —
			• Stall for Time — Tap up to three target creatures.
			• Call for Aid — Create three 1/1 white Soldier creature tokens.
		`,
		},
	}
}

var youReConfrontedByRobbersToken = newYouReConfrontedByRobbersToken()

func newYouReConfrontedByRobbersToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Soldier",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Soldier},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
