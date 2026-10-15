package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ElvenChorus is the card definition for Elven Chorus.
//
// Type: Enchantment
// Cost: {3}{G}
//
// Oracle text:
//
//	You may look at the top card of your library any time.
//	You may cast creature spells from the top of your library.
//	Creatures you control have "{T}: Add one mana of any color."
var ElvenChorus = newElvenChorus

func newElvenChorus() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Elven Chorus",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Enchantment},
			StaticAbilities: []game.StaticAbility{
				game.LookAtTopCardAnyTimeStaticBody,
				game.StaticAbility{
					Text: "You may cast creature spells from the top of your library.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCastSpellsFromZone,
							Duration:       game.DurationPermanent,
							AffectedPlayer: game.PlayerYou,
							SpellTypes:     []types.Card{types.Creature},
							CastFromZone:   zone.Library,
							TopCardOnly:    true,
						},
					},
				},
				game.StaticAbility{
					Text: "Creatures you control have \"{T}: Add one mana of any color.\"",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration: game.DurationPermanent,
							Group:    game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}}),
							Layer:    game.LayerAbility,
							AddAbilities: []game.Ability{
								new(game.TapManaChoiceAbility(mana.W, mana.U, mana.B, mana.R, mana.G)),
							},
						},
					},
				},
			},
			OracleText: `
			You may look at the top card of your library any time.
			You may cast creature spells from the top of your library.
			Creatures you control have "{T}: Add one mana of any color."
		`,
		},
	}
}
