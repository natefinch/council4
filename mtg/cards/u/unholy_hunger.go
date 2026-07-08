package u

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// UnholyHunger is the card definition for Unholy Hunger.
//
// Type: Instant
// Cost: {3}{B}{B}
//
// Oracle text:
//
//	Destroy target creature.
//	Spell mastery — If there are two or more instant and/or sorcery cards in your graveyard, you gain 2 life.
var UnholyHunger = newUnholyHunger

func newUnholyHunger() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Unholy Hunger",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target creature",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Destroy{
							Object: game.TargetPermanentReference(0),
						},
					},
					{
						Primitive: game.GainLife{
							Amount: game.Fixed(2),
							Player: game.ControllerReference(),
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								ControllerGraveyardInstantOrSorceryCountAtLeast: 2,
							}),
						}),
					},
				},
			}.Ability()),
			OracleText: `
			Destroy target creature.
			Spell mastery — If there are two or more instant and/or sorcery cards in your graveyard, you gain 2 life.
		`,
		},
	}
}
