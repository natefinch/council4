package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// LaeZelVlaakithSChampion is the card definition for Lae'zel, Vlaakith's Champion.
//
// Type: Legendary Creature — Gith Warrior
// Cost: {2}{W}
//
// Oracle text:
//
//	If you would put one or more counters on a creature or planeswalker you control or on yourself, put that many plus one of each of those kinds of counters on that permanent or player instead.
//	Choose a Background (You can have a Background as a second commander.)
var LaeZelVlaakithSChampion = newLaeZelVlaakithSChampion

func newLaeZelVlaakithSChampion() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Lae'zel, Vlaakith's Champion",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
			}),
			Colors:     []color.Color{color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Gith, types.Warrior},
			Power:      opt.Val(game.PT{Value: 3}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.ChooseABackgroundStaticBody,
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.ControlledPermanentTypesOrControllerCounterPlacementReplacement("If you would put one or more counters on a creature or planeswalker you control or on yourself, put that many plus one of each of those kinds of counters on that permanent or player instead.", 0, 1, []types.Card{types.Creature, types.Planeswalker}, game.TriggerControllerYou),
			},
			OracleText: `
			If you would put one or more counters on a creature or planeswalker you control or on yourself, put that many plus one of each of those kinds of counters on that permanent or player instead.
			Choose a Background (You can have a Background as a second commander.)
		`,
		},
	}
}
