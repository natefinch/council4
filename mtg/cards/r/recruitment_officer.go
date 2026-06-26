package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// RecruitmentOfficer is the card definition for Recruitment Officer.
//
// Type: Creature — Human Soldier
// Cost: {W}
//
// Oracle text:
//
//	{3}{W}: Look at the top four cards of your library. You may reveal a creature card with mana value 3 or less from among them and put it into your hand. Put the rest on the bottom of your library in a random order.
var RecruitmentOfficer = newRecruitmentOfficer()

func newRecruitmentOfficer() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Recruitment Officer",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Soldier},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:           "{3}{W}: Look at the top four cards of your library. You may reveal a creature card with mana value 3 or less from among them and put it into your hand. Put the rest on the bottom of your library in a random order.",
					ManaCost:       opt.Val(cost.Mana{cost.O(3), cost.W}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Dig{
									Player:    game.ControllerReference(),
									Look:      game.Fixed(4),
									Take:      game.Fixed(1),
									Remainder: game.DigRemainderLibraryBottom,
									Filter:    opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}, ManaValue: opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 3})}),
									TakeUpTo:  true,
									Reveal:    true,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			{3}{W}: Look at the top four cards of your library. You may reveal a creature card with mana value 3 or less from among them and put it into your hand. Put the rest on the bottom of your library in a random order.
		`,
		},
	}
}
