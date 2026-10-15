package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TwinflameTyrant is the card definition for Twinflame Tyrant.
//
// Type: Creature — Dragon
// Cost: {3}{R}{R}
//
// Oracle text:
//
//	Flying
//	If a source you control would deal damage to an opponent or a permanent an opponent controls, it deals double that damage instead.
var TwinflameTyrant = newTwinflameTyrant

func newTwinflameTyrant() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Twinflame Tyrant",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Dragon},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.DamageReplacementFiltered("If a source you control would deal damage to an opponent or a permanent an opponent controls, it deals double that damage instead.", &game.DamageReplacementSpec{Multiplier: 2, Addend: 0, SourceColors: nil, SourceTypes: nil, ExcludeSource: false, RecipientOpponent: true, RecipientOpponentPlayerOnly: false, NoncombatOnly: false, Controller: game.TriggerControllerYou}),
			},
			OracleText: `
			Flying
			If a source you control would deal damage to an opponent or a permanent an opponent controls, it deals double that damage instead.
		`,
		},
	}
}
