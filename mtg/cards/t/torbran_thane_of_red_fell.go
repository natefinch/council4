package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TorbranThaneOfRedFell is the card definition for Torbran, Thane of Red Fell.
//
// Type: Legendary Creature — Dwarf Noble
// Cost: {1}{R}{R}{R}
//
// Oracle text:
//
//	If a red source you control would deal damage to an opponent or a permanent an opponent controls, it deals that much damage plus 2 instead.
var TorbranThaneOfRedFell = newTorbranThaneOfRedFell

func newTorbranThaneOfRedFell() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Torbran, Thane of Red Fell",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
				cost.R,
				cost.R,
			}),
			Colors:     []color.Color{color.Red},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Dwarf, types.Noble},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 4}),
			ReplacementAbilities: []game.ReplacementAbility{
				game.DamageReplacementFiltered("If a red source you control would deal damage to an opponent or a permanent an opponent controls, it deals that much damage plus 2 instead.", &game.DamageReplacementSpec{Multiplier: 0, Addend: 2, SourceColors: []color.Color{color.Red}, SourceTypes: nil, ExcludeSource: false, RecipientOpponent: true, RecipientOpponentPlayerOnly: false, NoncombatOnly: false, Controller: game.TriggerControllerYou}),
			},
			OracleText: `
			If a red source you control would deal damage to an opponent or a permanent an opponent controls, it deals that much damage plus 2 instead.
		`,
		},
	}
}
