package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ThorAsgardSAvenger is the card definition for Thor, Asgard's Avenger.
//
// Type: Legendary Creature — God Warrior Hero
// Cost: {2}{R}{R}
//
// Oracle text:
//
//	Flying
//	If another source you control would deal damage to an opponent or a permanent an opponent controls, it deals that much damage plus 1 instead.
var ThorAsgardSAvenger = newThorAsgardSAvenger

func newThorAsgardSAvenger() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Thor, Asgard's Avenger",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
				cost.R,
			}),
			Colors:     []color.Color{color.Red},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.God, types.Warrior, types.Hero},
			Power:      opt.Val(game.PT{Value: 4}),
			Toughness:  opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.DamageReplacementFiltered("If another source you control would deal damage to an opponent or a permanent an opponent controls, it deals that much damage plus 1 instead.", &game.DamageReplacementSpec{Multiplier: 0, Addend: 1, SourceColors: nil, SourceTypes: nil, ExcludeSource: true, RecipientOpponent: true, RecipientOpponentPlayerOnly: false, NoncombatOnly: false, Controller: game.TriggerControllerYou}),
			},
			OracleText: `
			Flying
			If another source you control would deal damage to an opponent or a permanent an opponent controls, it deals that much damage plus 1 instead.
		`,
		},
	}
}
