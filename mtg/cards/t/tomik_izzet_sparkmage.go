package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TomikIzzetSparkmage is the card definition for Tomik, Izzet Sparkmage.
//
// Type: Legendary Creature — Human Wizard
// Cost: {1}{R}
//
// Oracle text:
//
//	Prowess (Whenever you cast a noncreature spell, this creature gets +1/+1 until end of turn.)
//	If a source you control would deal noncombat damage to an opponent or a permanent an opponent controls, it deals that much damage plus 1 instead.
var TomikIzzetSparkmage = newTomikIzzetSparkmage

func newTomikIzzetSparkmage() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Tomik, Izzet Sparkmage",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors:     []color.Color{color.Red},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Wizard},
			Power:      opt.Val(game.PT{Value: 1}),
			Toughness:  opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.ProwessStaticBody,
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.DamageReplacementFiltered("If a source you control would deal noncombat damage to an opponent or a permanent an opponent controls, it deals that much damage plus 1 instead.", &game.DamageReplacementSpec{Multiplier: 0, Addend: 1, SourceColors: nil, SourceTypes: nil, ExcludeSource: false, RecipientOpponent: true, RecipientOpponentPlayerOnly: false, NoncombatOnly: true, Controller: game.TriggerControllerYou}),
			},
			OracleText: `
			Prowess (Whenever you cast a noncreature spell, this creature gets +1/+1 until end of turn.)
			If a source you control would deal noncombat damage to an opponent or a permanent an opponent controls, it deals that much damage plus 1 instead.
		`,
		},
	}
}
