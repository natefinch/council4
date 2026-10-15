package n

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// NalaarAetherjet is the card definition for Nalaar Aetherjet.
//
// Type: Token Artifact — Vehicle
//
// Oracle text:
//   Flying
//   Crew 2 (Tap any number of creatures you control with total power 2 or more: This token becomes an artifact creature until end of turn.)

// NalaarAetherjetTokenb51ade1bbe734318b513c99bcf9b2b71 is the card definition for Nalaar Aetherjet.
var NalaarAetherjetTokenb51ade1bbe734318b513c99bcf9b2b71 = newNalaarAetherjetTokenb51ade1bbe734318b513c99bcf9b2b71()

func newNalaarAetherjetTokenb51ade1bbe734318b513c99bcf9b2b71() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Nalaar Aetherjet",
			Types:     []types.Card{types.Artifact},
			Subtypes:  []types.Sub{types.Vehicle},
			Power:     opt.Val(game.PT{IsStar: true}),
			Toughness: opt.Val(game.PT{IsStar: true}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.CrewActivatedAbility(2),
			},
			OracleText: `
			Flying
			Crew 2 (Tap any number of creatures you control with total power 2 or more: This token becomes an artifact creature until end of turn.)
		`,
		},
	}
}
