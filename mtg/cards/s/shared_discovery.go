package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SharedDiscovery is the card definition for Shared Discovery.
//
// Type: Sorcery
// Cost: {U}
//
// Oracle text:
//
//	As an additional cost to cast this spell, tap four untapped creatures you control.
//	Draw three cards.
var SharedDiscovery = newSharedDiscovery

func newSharedDiscovery() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Shared Discovery",
			ManaCost: opt.Val(cost.Mana{
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Sorcery},
			AdditionalCosts: []cost.Additional{
				{
					Kind:               cost.AdditionalTapPermanents,
					Text:               "tap four untapped creatures you control",
					Amount:             4,
					MatchPermanentType: true,
					PermanentType:      types.Creature,
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Draw{
							Amount: game.Fixed(3),
							Player: game.ControllerReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			As an additional cost to cast this spell, tap four untapped creatures you control.
			Draw three cards.
		`,
		},
	}
}
