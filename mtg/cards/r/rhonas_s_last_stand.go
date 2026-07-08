package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RhonasSLastStand is the card definition for Rhonas's Last Stand.
//
// Type: Sorcery
// Cost: {G}{G}
//
// Oracle text:
//
//	Create a 5/4 green Snake creature token. Lands you control don't untap during your next untap step.
var RhonasSLastStand = newRhonasSLastStand

func newRhonasSLastStand() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Rhonas's Last Stand",
			ManaCost: opt.Val(cost.Mana{
				cost.G,
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(1),
							Source: game.TokenDef(rhonasSLastStandToken),
						},
					},
					{
						Primitive: game.SkipNextUntap{
							Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Land}, Controller: game.ControllerYou}),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create a 5/4 green Snake creature token. Lands you control don't untap during your next untap step.
		`,
		},
	}
}

var rhonasSLastStandToken = newRhonasSLastStandToken()

func newRhonasSLastStandToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Snake",
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Snake},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 4}),
		},
	}
}
