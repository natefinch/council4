package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// GrafdiggerSCage is the card definition for Grafdigger's Cage.
//
// Type: Artifact
// Cost: {1}
//
// Oracle text:
//
//	Creature cards in graveyards and libraries can't enter the battlefield.
//	Players can't cast spells from graveyards or libraries.
var GrafdiggerSCage = newGrafdiggerSCage

func newGrafdiggerSCage() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Grafdigger's Cage",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
			}),
			Types: []types.Card{types.Artifact},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Creature cards in graveyards and libraries can't enter the battlefield.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCantEnterFromZones,
							Duration:       game.DurationPermanent,
							PermanentTypes: []types.Card{types.Creature},
							EnterFromZones: []zone.Type{zone.Graveyard, zone.Library},
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
			Creature cards in graveyards and libraries can't enter the battlefield.
			Players can't cast spells from graveyards or libraries.
		`,
		},
	}
}
