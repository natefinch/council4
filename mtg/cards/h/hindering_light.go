package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HinderingLight is the card definition for Hindering Light.
//
// Type: Instant
// Cost: {W}{U}
//
// Oracle text:
//
//	Counter target spell that targets you or a permanent you control.
//	Draw a card.
var HinderingLight = newHinderingLight()

func newHinderingLight() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue),
		CardFace: game.CardFace{
			Name: "Hindering Light",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
				cost.U,
			}),
			Colors: []color.Color{color.Blue, color.White},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target spell that targets you or a permanent you control",
						Allow:      game.TargetAllowStackObject,
						Predicate: game.TargetPredicate{
							StackObjectKinds: []game.StackObjectKind{game.StackSpell},
							SpellTargets: []game.SpellTargetRequirement{game.SpellTargetRequirement{
								Kind:   game.SpellTargetRequirementPlayer,
								Player: game.PlayerYou,
							}, game.SpellTargetRequirement{
								Kind:       game.SpellTargetRequirementPermanent,
								Controller: game.ControllerYou,
							}},
						},
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.CounterObject{
							Object: game.TargetStackObjectReference(0),
						},
					},
					{
						Primitive: game.Draw{
							Amount: game.Fixed(1),
							Player: game.ControllerReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Counter target spell that targets you or a permanent you control.
			Draw a card.
		`,
		},
	}
}
