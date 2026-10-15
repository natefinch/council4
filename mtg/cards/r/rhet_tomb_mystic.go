package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RhetTombMystic is the card definition for Rhet-Tomb Mystic.
//
// Type: Creature — Zombie Bird Wizard
// Cost: {1}{U}
//
// Oracle text:
//
//	Flying
//	Each creature card in your hand has cycling {1}{U}. ({1}{U}, Discard that card: Draw a card.)
var RhetTombMystic = newRhetTombMystic

func newRhetTombMystic() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Rhet-Tomb Mystic",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Zombie, types.Bird, types.Wizard},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.StaticAbility{
					Text: "Each creature card in your hand has cycling {1}{U}.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectGrantHandCardAbility,
							Duration:       game.DurationPermanent,
							AffectedPlayer: game.PlayerYou,
							CardSelection:  game.Selection{RequiredTypes: []types.Card{types.Creature}},
							GrantedAbility: game.CyclingActivatedAbility(cost.Mana{cost.O(1), cost.U}),
						},
					},
				},
			},
			OracleText: `
			Flying
			Each creature card in your hand has cycling {1}{U}. ({1}{U}, Discard that card: Draw a card.)
		`,
		},
	}
}
