package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SeedbornMuse is the card definition for Seedborn Muse.
//
// Type: Creature — Spirit
// Cost: {3}{G}{G}
//
// Oracle text:
//
//	Untap all permanents you control during each other player's untap step.
var SeedbornMuse = newSeedbornMuse

func newSeedbornMuse() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Seedborn Muse",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Spirit},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Untap all permanents you control during each other player's untap step.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:               game.RuleEffectUntapDuringOtherPlayersUntapStep,
							Duration:           game.DurationPermanent,
							AffectedController: game.ControllerYou,
						},
					},
				},
			},
			OracleText: `
			Untap all permanents you control during each other player's untap step.
		`,
		},
	}
}
