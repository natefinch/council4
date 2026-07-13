package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PsychicPuppetry is the card definition for Psychic Puppetry.
//
// Type: Instant — Arcane
// Cost: {1}{U}
//
// Oracle text:
//
//	You may tap or untap target permanent.
//	Splice onto Arcane {U} (As you cast an Arcane spell, you may reveal this card from your hand and pay its splice cost. If you do, add this card's effects to that spell.)
var PsychicPuppetry = newPsychicPuppetry

func newPsychicPuppetry() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Psychic Puppetry",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors:   []color.Color{color.Blue},
			Types:    []types.Card{types.Instant},
			Subtypes: []types.Sub{types.Arcane},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					KeywordAbilities: []game.KeywordAbility{
						game.SpliceKeyword{Cost: cost.Mana{cost.U}},
					},
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target permanent",
						Allow:      game.TargetAllowPermanent,
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.TapOrUntap{
							Object: game.TargetPermanentReference(0),
						},
						Optional: true,
					},
				},
			}.Ability()),
			OracleText: `
			You may tap or untap target permanent.
			Splice onto Arcane {U} (As you cast an Arcane spell, you may reveal this card from your hand and pay its splice cost. If you do, add this card's effects to that spell.)
		`,
		},
	}
}
