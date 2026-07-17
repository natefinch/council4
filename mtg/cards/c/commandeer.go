package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Commandeer is the card definition for Commandeer.
//
// Type: Instant
// Cost: {5}{U}{U}
//
// Oracle text:
//
//	You may exile two blue cards from your hand rather than pay this spell's mana cost.
//	Gain control of target noncreature spell. You may choose new targets for it. (If that spell is an artifact, enchantment, or planeswalker, the permanent enters under your control.)
var Commandeer = newCommandeer

func newCommandeer() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Commandeer",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.U,
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			AlternativeCosts: []cost.Alternative{
				cost.Alternative{
					Label: "Exile 2 blue cards",
					AdditionalCosts: []cost.Additional{
						{
							Kind:           cost.AdditionalExile,
							Amount:         2,
							Source:         zone.Hand,
							MatchCardColor: true,
							CardColor:      color.Blue,
						},
					},
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target noncreature spell",
						Allow:      game.TargetAllowStackObject,
						Predicate: game.TargetPredicate{
							ExcludedSpellCardTypes: []types.Card{types.Creature},
							StackObjectKinds:       []game.StackObjectKind{game.StackSpell},
						},
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.ChangeStackObjectController{
							Object:     game.TargetStackObjectReference(0),
							Controller: game.ControllerReference(),
						},
					},
					{
						Primitive: game.ChooseNewTargets{
							Object: game.TargetStackObjectReference(0),
						},
						Optional: true,
					},
				},
			}.Ability()),
			OracleText: `
			You may exile two blue cards from your hand rather than pay this spell's mana cost.
			Gain control of target noncreature spell. You may choose new targets for it. (If that spell is an artifact, enchantment, or planeswalker, the permanent enters under your control.)
		`,
		},
	}
}
