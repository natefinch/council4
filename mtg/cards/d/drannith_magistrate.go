package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// DrannithMagistrate is the card definition for Drannith Magistrate.
//
// Type: Creature — Human Wizard
// Cost: {1}{W}
//
// Oracle text:
//
//	Your opponents can't cast spells from anywhere other than their hands.
var DrannithMagistrate = newDrannithMagistrate

func newDrannithMagistrate() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Drannith Magistrate",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Wizard},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Your opponents can't cast spells from anywhere other than their hands.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:              game.RuleEffectCantCastFromZones,
							Duration:          game.DurationPermanent,
							AffectedPlayer:    game.PlayerOpponent,
							CantCastFromZones: []zone.Type{zone.Graveyard, zone.Exile, zone.Library, zone.Command},
						},
					},
				},
			},
			OracleText: `
			Your opponents can't cast spells from anywhere other than their hands.
		`,
		},
	}
}
