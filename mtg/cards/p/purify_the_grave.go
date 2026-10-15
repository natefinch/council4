package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// PurifyTheGrave is the card definition for Purify the Grave.
//
// Type: Instant
// Cost: {W}
//
// Oracle text:
//
//	Exile target card from a graveyard.
//	Flashback {W} (You may cast this card from your graveyard for its flashback cost. Then exile it.)
var PurifyTheGrave = newPurifyTheGrave

func newPurifyTheGrave() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Purify the Grave",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Instant},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Flashback {W}",
					KeywordAbilities: []game.KeywordAbility{
						game.FlashbackKeyword{
							Cost: cost.Mana{
								cost.W,
							},
						},
					},
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target card from a graveyard",
						Allow:      game.TargetAllowCard,
						TargetZone: zone.Graveyard,
						Selection:  opt.Val(game.Selection{}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.MoveCard{
							Card: game.CardReference{
								Kind: game.CardReferenceTarget,
							},
							FromZone:    zone.Graveyard,
							Destination: zone.Exile,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Exile target card from a graveyard.
			Flashback {W} (You may cast this card from your graveyard for its flashback cost. Then exile it.)
		`,
		},
	}
}
