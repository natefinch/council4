package y

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// YouMeetInATavern is the card definition for You Meet in a Tavern.
//
// Type: Sorcery
// Cost: {2}{G}{G}
//
// Oracle text:
//
//	Choose one —
//	• Form a Party — Look at the top five cards of your library. You may reveal any number of creature cards from among them and put them into your hand. Put the rest on the bottom of your library in a random order.
//	• Start a Brawl — Creatures you control get +2/+2 until end of turn.
var YouMeetInATavern = newYouMeetInATavern()

func newYouMeetInATavern() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "You Meet in a Tavern",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Form a Party — Look at the top five cards of your library. You may reveal any number of creature cards from among them and put them into your hand. Put the rest on the bottom of your library in a random order.",
						Sequence: []game.Instruction{
							{
								Primitive: game.Dig{
									Player:    game.ControllerReference(),
									Look:      game.Fixed(5),
									Take:      game.Fixed(5),
									Remainder: game.DigRemainderLibraryBottom,
									Filter:    opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
									TakeUpTo:  true,
									Reveal:    true,
								},
							},
						},
					},
					game.Mode{
						Text: "Start a Brawl — Creatures you control get +2/+2 until end of turn.",
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer:          game.LayerPowerToughnessModify,
											Group:          game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
											PowerDelta:     2,
											ToughnessDelta: 2,
										},
									},
									Duration: game.DurationUntilEndOfTurn,
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
			• Form a Party — Look at the top five cards of your library. You may reveal any number of creature cards from among them and put them into your hand. Put the rest on the bottom of your library in a random order.
			• Start a Brawl — Creatures you control get +2/+2 until end of turn.
		`,
		},
	}
}
