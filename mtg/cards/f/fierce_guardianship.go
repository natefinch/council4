package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FierceGuardianship is the card definition for Fierce Guardianship.
//
// Type: Instant
// Cost: {2}{U}
//
// Oracle text:
//
//	If you control a commander, you may cast this spell without paying its mana cost.
//	Counter target noncreature spell.
var FierceGuardianship = newFierceGuardianship

func newFierceGuardianship() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Fierce Guardianship",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			AlternativeCosts: []cost.Alternative{
				cost.Alternative{
					Label:     "Cast without paying mana cost",
					Condition: cost.AlternativeConditionControlsCommander,
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
						Primitive: game.CounterObject{
							Object: game.TargetStackObjectReference(0),
						},
					},
				},
			}.Ability()),
			OracleText: `
			If you control a commander, you may cast this spell without paying its mana cost.
			Counter target noncreature spell.
		`,
		},
	}
}
