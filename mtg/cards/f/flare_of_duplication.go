package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FlareOfDuplication is the card definition for Flare of Duplication.
//
// Type: Instant
// Cost: {1}{R}{R}
//
// Oracle text:
//
//	You may sacrifice a nontoken red creature rather than pay this spell's mana cost.
//	Copy target instant or sorcery spell. You may choose new targets for the copy.
var FlareOfDuplication = newFlareOfDuplication

func newFlareOfDuplication() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Flare of Duplication",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Instant},
			AlternativeCosts: []cost.Alternative{
				cost.Alternative{
					Label: "Sacrifice a nontoken red creature",
					AdditionalCosts: []cost.Additional{
						{
							Kind:               cost.AdditionalSacrifice,
							Text:               "sacrifice a nontoken red creature",
							Amount:             1,
							MatchPermanentType: true,
							PermanentType:      types.Creature,
							MatchCardColor:     true,
							CardColor:          color.Red,
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
						Constraint: "target instant or sorcery spell",
						Allow:      game.TargetAllowStackObject,
						Predicate: game.TargetPredicate{
							SpellCardTypesAny: []types.Card{types.Instant, types.Sorcery},
							StackObjectKinds:  []game.StackObjectKind{game.StackSpell},
						},
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.CopyStackObject{
							Object:              game.TargetStackObjectReference(0),
							Count:               1,
							MayChooseNewTargets: true,
						},
					},
				},
			}.Ability()),
			OracleText: `
			You may sacrifice a nontoken red creature rather than pay this spell's mana cost.
			Copy target instant or sorcery spell. You may choose new targets for the copy.
		`,
		},
	}
}
