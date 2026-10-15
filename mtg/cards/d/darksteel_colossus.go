package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// DarksteelColossus is the card definition for Darksteel Colossus.
//
// Type: Artifact Creature — Golem
// Cost: {11}
//
// Oracle text:
//
//	Trample (This creature can deal excess combat damage to the player or planeswalker it's attacking.)
//	Indestructible (Damage and effects that say "destroy" don't destroy this creature.)
//	If Darksteel Colossus would be put into a graveyard from anywhere, reveal Darksteel Colossus and shuffle it into its owner's library instead.
var DarksteelColossus = newDarksteelColossus

func newDarksteelColossus() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Darksteel Colossus",
			ManaCost: opt.Val(cost.Mana{
				cost.O(11),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Golem},
			Power:     opt.Val(game.PT{Value: 11}),
			Toughness: opt.Val(game.PT{Value: 11}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
				game.IndestructibleStaticBody,
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.ReplacementAbility{Text: "If Darksteel Colossus would be put into a graveyard from anywhere, reveal Darksteel Colossus and shuffle it into its owner's library instead.", Replacement: game.ReplacementEffect{
					MatchEvent:         game.EventZoneChanged,
					MatchToZone:        true,
					ToZone:             zone.Graveyard,
					ReplaceToZone:      zone.Library,
					Duration:           game.DurationPermanent,
					ShuffleIntoLibrary: true,
					RevealSource:       true,
				}},
			},
			OracleText: `
			Trample (This creature can deal excess combat damage to the player or planeswalker it's attacking.)
			Indestructible (Damage and effects that say "destroy" don't destroy this creature.)
			If Darksteel Colossus would be put into a graveyard from anywhere, reveal Darksteel Colossus and shuffle it into its owner's library instead.
		`,
		},
	}
}
