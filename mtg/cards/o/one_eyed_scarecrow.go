package o

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// OneEyedScarecrow is the card definition for One-Eyed Scarecrow.
//
// Type: Artifact Creature — Scarecrow
// Cost: {3}
//
// Oracle text:
//
//	Defender
//	Creatures with flying your opponents control get -1/-0.
var OneEyedScarecrow = newOneEyedScarecrow

func newOneEyedScarecrow() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "One-Eyed Scarecrow",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Scarecrow},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.DefenderStaticBody,
				game.StaticAbility{
					Text: "Creatures with flying your opponents control get -1/-0.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:   game.DurationPermanent,
							Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerOpponent, Keyword: game.Flying}),
							Layer:      game.LayerPowerToughnessModify,
							PowerDelta: -1,
						},
					},
				},
			},
			OracleText: `
			Defender
			Creatures with flying your opponents control get -1/-0.
		`,
		},
	}
}
