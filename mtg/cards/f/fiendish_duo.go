package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FiendishDuo is the card definition for Fiendish Duo.
//
// Type: Creature — Devil
// Cost: {4}{R}{R}
//
// Oracle text:
//
//	First strike (This creature deals combat damage before creatures without first strike.)
//	If a source would deal damage to an opponent, it deals double that damage to that player instead.
var FiendishDuo = newFiendishDuo

func newFiendishDuo() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Fiendish Duo",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.R,
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Devil},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.FirstStrikeStaticBody,
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.DamageReplacementFiltered("If a source would deal damage to an opponent, it deals double that damage to that player instead.", &game.DamageReplacementSpec{Multiplier: 2, Addend: 0, SourceColors: nil, SourceTypes: nil, ExcludeSource: false, RecipientOpponent: true, RecipientOpponentPlayerOnly: true, NoncombatOnly: false, Controller: game.TriggerControllerAny}),
			},
			OracleText: `
			First strike (This creature deals combat damage before creatures without first strike.)
			If a source would deal damage to an opponent, it deals double that damage to that player instead.
		`,
		},
	}
}
