package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DarkDabbling is the card definition for Dark Dabbling.
//
// Type: Instant
// Cost: {2}{B}
//
// Oracle text:
//
//	Regenerate target creature. Draw a card. (The next time the creature would be destroyed this turn, instead tap it, remove it from combat, and heal all damage on it.)
//	Spell mastery — If there are two or more instant and/or sorcery cards in your graveyard, also regenerate each other creature you control.
var DarkDabbling = newDarkDabbling()

func newDarkDabbling() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Dark Dabbling",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
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
						Primitive: game.Regenerate{
							Object: game.TargetPermanentReference(0),
						},
					},
					{
						Primitive: game.Draw{
							Amount: game.Fixed(1),
							Player: game.ControllerReference(),
						},
					},
					{
						Primitive: game.Regenerate{
							Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou, ExcludeSource: true}),
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
			Regenerate target creature. Draw a card. (The next time the creature would be destroyed this turn, instead tap it, remove it from combat, and heal all damage on it.)
			Spell mastery — If there are two or more instant and/or sorcery cards in your graveyard, also regenerate each other creature you control.
		`,
		},
	}
}
