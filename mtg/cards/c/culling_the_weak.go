package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CullingTheWeak is the card definition for Culling the Weak.
//
// Type: Instant
// Cost: {B}
//
// Oracle text:
//
//	As an additional cost to cast this spell, sacrifice a creature.
//	Add {B}{B}{B}{B}.
var CullingTheWeak = newCullingTheWeak

func newCullingTheWeak() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Culling the Weak",
			ManaCost: opt.Val(cost.Mana{
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Instant},
			AdditionalCosts: []cost.Additional{
				{
					Kind:               cost.AdditionalSacrifice,
					Text:               "sacrifice a creature",
					Amount:             1,
					MatchPermanentType: true,
					PermanentType:      types.Creature,
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.AddMana{
							Amount:    game.Fixed(1),
							ManaColor: mana.B,
						},
					},
					{
						Primitive: game.AddMana{
							Amount:    game.Fixed(1),
							ManaColor: mana.B,
						},
					},
					{
						Primitive: game.AddMana{
							Amount:    game.Fixed(1),
							ManaColor: mana.B,
						},
					},
					{
						Primitive: game.AddMana{
							Amount:    game.Fixed(1),
							ManaColor: mana.B,
						},
					},
				},
			}.Ability()),
			OracleText: `
			As an additional cost to cast this spell, sacrifice a creature.
			Add {B}{B}{B}{B}.
		`,
		},
	}
}
