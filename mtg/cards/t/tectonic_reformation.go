package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TectonicReformation is the card definition for Tectonic Reformation.
//
// Type: Enchantment
// Cost: {1}{R}
//
// Oracle text:
//
//	Each land card in your hand has cycling {R}.
//	Cycling {2} ({2}, Discard this card: Draw a card.)
var TectonicReformation = newTectonicReformation

func newTectonicReformation() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Tectonic Reformation",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Enchantment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Each land card in your hand has cycling {R}.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectGrantHandCardAbility,
							Duration:       game.DurationPermanent,
							AffectedPlayer: game.PlayerYou,
							CardSelection:  game.Selection{RequiredTypes: []types.Card{types.Land}},
							GrantedAbility: game.CyclingActivatedAbility(cost.Mana{cost.R}),
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.CyclingActivatedAbility(cost.Mana{cost.O(2)}),
			},
			OracleText: `
			Each land card in your hand has cycling {R}.
			Cycling {2} ({2}, Discard this card: Draw a card.)
		`,
		},
	}
}
