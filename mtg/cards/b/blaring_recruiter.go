package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// BlaringRecruiter is the card definition for Blaring Recruiter.
//
// Type: Creature — Elf Warrior
// Cost: {3}{W}
//
// Oracle text:
//
//	Partner with Blaring Captain (When this creature enters, target player may put Blaring Captain into their hand from their library, then shuffle.)
//	{2}{W}: Create a 1/1 white Warrior creature token.
var BlaringRecruiter = newBlaringRecruiter

func newBlaringRecruiter() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Blaring Recruiter",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elf, types.Warrior},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.PartnerWithStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.PartnerWithTriggeredAbility("Blaring Captain"),
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{2}{W}: Create a 1/1 white Warrior creature token.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(2),
						cost.W,
					}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Fixed(1),
											Source: game.TokenDef(blaringRecruiterToken),
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
			Partner with Blaring Captain (When this creature enters, target player may put Blaring Captain into their hand from their library, then shuffle.)
			{2}{W}: Create a 1/1 white Warrior creature token.
		`,
		},
	}
}

var blaringRecruiterToken = newBlaringRecruiterToken()

func newBlaringRecruiterToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Warrior",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Warrior},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
