package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CrashTheParty is the card definition for Crash the Party.
//
// Type: Instant
// Cost: {5}{G}
//
// Oracle text:
//
//	Create a tapped 4/4 green Rhino Warrior creature token for each tapped creature you control.
var CrashTheParty = newCrashTheParty()

func newCrashTheParty() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Crash the Party",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:       game.DynamicAmountCountSelector,
								Multiplier: 1,
								Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou, Tapped: game.TriTrue}),
							}),
							Source:      game.TokenDef(crashThePartyToken),
							EntryTapped: true,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create a tapped 4/4 green Rhino Warrior creature token for each tapped creature you control.
		`,
		},
	}
}

var crashThePartyToken = newCrashThePartyToken()

func newCrashThePartyToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Rhino Warrior",
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Rhino, types.Warrior},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
		},
	}
}
