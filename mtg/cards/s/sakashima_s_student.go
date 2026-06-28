package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SakashimaSStudent is the card definition for Sakashima's Student.
//
// Type: Creature — Human Ninja
// Cost: {2}{U}{U}
//
// Oracle text:
//
//	Ninjutsu {1}{U} ({1}{U}, Return an unblocked attacker you control to hand: Put this card onto the battlefield from your hand tapped and attacking.)
//	You may have this creature enter as a copy of any creature on the battlefield, except it's a Ninja in addition to its other creature types.
var SakashimaSStudent = newSakashimaSStudent()

func newSakashimaSStudent() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Sakashima's Student",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Ninja},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 0}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:     "Ninjutsu {1}{U}",
					ManaCost: opt.Val(cost.Mana{cost.O(1), cost.U}),
					AdditionalCosts: []cost.Additional{
						{
							Kind:   cost.AdditionalReturnUnblockedAttacker,
							Text:   "Return an unblocked attacker you control to its owner's hand",
							Amount: 1,
						},
					},
					ZoneOfFunction: zone.Hand,
					Timing:         game.DuringCombat,
					KeywordAbilities: []game.KeywordAbility{
						game.NinjutsuKeyword{Cost: cost.Mana{cost.O(1), cost.U}},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{},
					}.Ability(),
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersAsCopyReplacement("You may have this creature enter as a copy of any creature on the battlefield, except it's a Ninja in addition to its other creature types.", &game.Selection{RequiredTypes: []types.Card{types.Creature}}, true, false, nil, false, nil, []types.Sub{types.Ninja}),
			},
			OracleText: `
			Ninjutsu {1}{U} ({1}{U}, Return an unblocked attacker you control to hand: Put this card onto the battlefield from your hand tapped and attacking.)
			You may have this creature enter as a copy of any creature on the battlefield, except it's a Ninja in addition to its other creature types.
		`,
		},
	}
}
