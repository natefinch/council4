package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Realmwalker is the card definition for Realmwalker.
//
// Type: Creature — Shapeshifter
// Cost: {2}{G}
//
// Oracle text:
//
//	Changeling (This card is every creature type.)
//	As this creature enters, choose a creature type.
//	You may look at the top card of your library any time.
//	You may cast creature spells of the chosen type from the top of your library.
var Realmwalker = newRealmwalker

func newRealmwalker() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Realmwalker",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Shapeshifter},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.ChangelingStaticBody,
				game.LookAtTopCardAnyTimeStaticBody,
				game.StaticAbility{
					Text: "You may cast creature spells of the chosen type from the top of your library.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:                   game.RuleEffectCastSpellsFromZone,
							Duration:               game.DurationPermanent,
							AffectedPlayer:         game.PlayerYou,
							SpellTypes:             []types.Card{types.Creature},
							CastFromZone:           zone.Library,
							TopCardOnly:            true,
							SpellChosenSubtypeFrom: game.EntryTypeChoiceKey,
						},
					},
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntryTypeChoiceReplacement("As this creature enters, choose a creature type."),
			},
			OracleText: `
			Changeling (This card is every creature type.)
			As this creature enters, choose a creature type.
			You may look at the top card of your library any time.
			You may cast creature spells of the chosen type from the top of your library.
		`,
		},
	}
}
