package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FlareOfDenial is the card definition for Flare of Denial.
//
// Type: Instant
// Cost: {1}{U}{U}
//
// Oracle text:
//
//	You may sacrifice a nontoken blue creature rather than pay this spell's mana cost.
//	Counter target spell.
var FlareOfDenial = newFlareOfDenial

func newFlareOfDenial() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Flare of Denial",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			AlternativeCosts: []cost.Alternative{
				cost.Alternative{
					Label: "Sacrifice a nontoken blue creature",
					AdditionalCosts: []cost.Additional{
						{
							Kind:               cost.AdditionalSacrifice,
							Text:               "sacrifice a nontoken blue creature",
							Amount:             1,
							MatchPermanentType: true,
							PermanentType:      types.Creature,
							MatchCardColor:     true,
							CardColor:          color.Blue,
							RequireNonToken:    true,
						},
					},
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target spell",
						Allow:      game.TargetAllowStackObject,
						Predicate: game.TargetPredicate{
							StackObjectKinds: []game.StackObjectKind{game.StackSpell},
						},
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.CounterObject{
							Object: game.TargetStackObjectReference(0),
						},
					},
				},
			}.Ability()),
			OracleText: `
			You may sacrifice a nontoken blue creature rather than pay this spell's mana cost.
			Counter target spell.
		`,
		},
	}
}
