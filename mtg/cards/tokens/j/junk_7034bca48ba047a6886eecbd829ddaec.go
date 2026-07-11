package j

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// Junk
//
// Type: Token Artifact — Junk
//
// Oracle text:
//   {T}, Sacrifice this artifact: Exile the top card of your library. You may play that card this turn. Activate only as a sorcery.

// JunkToken7034bca48ba047a6886eecbd829ddaec is the card definition for Junk.
var JunkToken7034bca48ba047a6886eecbd829ddaec = newJunkToken7034bca48ba047a6886eecbd829ddaec()

func newJunkToken7034bca48ba047a6886eecbd829ddaec() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "Junk",
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Junk},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{T}, Sacrifice this artifact: Exile the top card of your library. You may play that card this turn. Activate only as a sorcery.",
					AdditionalCosts: []cost.Additional{
						{
							Kind: cost.AdditionalTap,
						},
						{
							Kind:   cost.AdditionalSacrificeSource,
							Text:   "Sacrifice this artifact",
							Amount: 1,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Timing:         game.SorceryOnly,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ImpulseExile{
									Player:   game.ControllerReference(),
									Amount:   game.Fixed(1),
									Duration: game.DurationThisTurn,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			{T}, Sacrifice this artifact: Exile the top card of your library. You may play that card this turn. Activate only as a sorcery.
		`,
		},
	}
}
