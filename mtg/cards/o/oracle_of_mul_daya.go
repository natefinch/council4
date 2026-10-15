package o

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// OracleOfMulDaya is the card definition for Oracle of Mul Daya.
//
// Type: Creature — Elf Shaman
// Cost: {3}{G}
//
// Oracle text:
//
//	You may play an additional land on each of your turns.
//	Play with the top card of your library revealed.
//	You may play lands from the top of your library.
var OracleOfMulDaya = newOracleOfMulDaya

func newOracleOfMulDaya() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Oracle of Mul Daya",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elf, types.Shaman},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "You may play an additional land on each of your turns.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:                game.RuleEffectAdditionalLandPlays,
							Duration:            game.DurationPermanent,
							AffectedPlayer:      game.PlayerYou,
							AdditionalLandPlays: 1,
						},
					},
				},
				game.PlayWithTopCardRevealedStaticBody,
				game.PlayLandsFromLibraryTopStaticBody,
			},
			OracleText: `
			You may play an additional land on each of your turns.
			Play with the top card of your library revealed.
			You may play lands from the top of your library.
		`,
		},
	}
}
