package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SteamAugury is the card definition for Steam Augury.
//
// Type: Instant
// Cost: {2}{U}{R}
//
// Oracle text:
//
//	Reveal the top five cards of your library and separate them into two piles. An opponent chooses one of those piles. Put that pile into your hand and the other into your graveyard.
var SteamAugury = newSteamAugury

func newSteamAugury() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Red),
		CardFace: game.CardFace{
			Name: "Steam Augury",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
				cost.R,
			}),
			Colors: []color.Color{color.Red, color.Blue},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.PileSplit{
							Player:          game.ControllerReference(),
							Amount:          game.Fixed(5),
							ChooserOpponent: true,
							Kept:            zone.Hand,
							Other:           zone.Graveyard,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Reveal the top five cards of your library and separate them into two piles. An opponent chooses one of those piles. Put that pile into your hand and the other into your graveyard.
		`,
		},
	}
}
