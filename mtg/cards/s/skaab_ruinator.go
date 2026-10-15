package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SkaabRuinator is the card definition for Skaab Ruinator.
//
// Type: Creature — Zombie Horror
// Cost: {1}{U}{U}
//
// Oracle text:
//
//	As an additional cost to cast this spell, exile three creature cards from your graveyard.
//	Flying
//	You may cast this card from your graveyard.
var SkaabRuinator = newSkaabRuinator

func newSkaabRuinator() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Skaab Ruinator",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Zombie, types.Horror},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 6}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.StaticAbility{
					Text:           "You may cast this card from your graveyard.",
					ZoneOfFunction: zone.Graveyard,
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCastFromZone,
							AffectedSource: true,
							Duration:       game.DurationPermanent,
							AffectedPlayer: game.PlayerYou,
							CastFromZone:   zone.Graveyard,
						},
					},
				},
			},
			AdditionalCosts: []cost.Additional{
				{
					Kind:          cost.AdditionalExile,
					Text:          "exile three creature cards from your graveyard",
					Amount:        3,
					Source:        zone.Graveyard,
					MatchCardType: true,
					CardType:      types.Creature,
				},
			},
			OracleText: `
			As an additional cost to cast this spell, exile three creature cards from your graveyard.
			Flying
			You may cast this card from your graveyard.
		`,
		},
	}
}
