package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// WeatheredRunestone is the card definition for Weathered Runestone.
//
// Type: Artifact
// Cost: {2}
//
// Oracle text:
//
//	Nonland permanent cards in graveyards and libraries can't enter the battlefield.
//	Players can't cast spells from graveyards or libraries.
var WeatheredRunestone = newWeatheredRunestone

func newWeatheredRunestone() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Weathered Runestone",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
			Types: []types.Card{types.Artifact},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Nonland permanent cards in graveyards and libraries can't enter the battlefield.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:                  game.RuleEffectCantEnterFromZones,
							Duration:              game.DurationPermanent,
							EnterFromZones:        []zone.Type{zone.Graveyard, zone.Library},
							EnterExcludeLandCards: true,
						},
					},
				},
				game.StaticAbility{
					Text: "Players can't cast spells from graveyards or libraries.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:              game.RuleEffectCantCastFromZones,
							Duration:          game.DurationPermanent,
							CantCastFromZones: []zone.Type{zone.Graveyard, zone.Library},
						},
					},
				},
			},
			OracleText: `
			Nonland permanent cards in graveyards and libraries can't enter the battlefield.
			Players can't cast spells from graveyards or libraries.
		`,
		},
	}
}
