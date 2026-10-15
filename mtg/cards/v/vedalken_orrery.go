package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// VedalkenOrrery is the card definition for Vedalken Orrery.
//
// Type: Artifact
// Cost: {4}
//
// Oracle text:
//
//	You may cast spells as though they had flash.
var VedalkenOrrery = newVedalkenOrrery

func newVedalkenOrrery() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Vedalken Orrery",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
			}),
			Types: []types.Card{types.Artifact},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "You may cast spells as though they had flash.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCastSpellsAsThoughFlash,
							Duration:       game.DurationPermanent,
							AffectedPlayer: game.PlayerYou,
						},
					},
				},
			},
			OracleText: `
			You may cast spells as though they had flash.
		`,
		},
	}
}
