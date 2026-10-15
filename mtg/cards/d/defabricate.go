package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Defabricate is the card definition for Defabricate.
//
// Type: Instant
// Cost: {1}{U}
//
// Oracle text:
//
//	Choose one —
//	• Counter target artifact or enchantment spell. If that spell is countered this way, exile it instead of putting it into its owner's graveyard.
//	• Counter target activated or triggered ability.
var Defabricate = newDefabricate

func newDefabricate() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Defabricate",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Counter target artifact or enchantment spell. If that spell is countered this way, exile it instead of putting it into its owner's graveyard.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target artifact or enchantment spell",
								Allow:      game.TargetAllowStackObject,
								Predicate: game.TargetPredicate{
									SpellCardTypesAny: []types.Card{types.Artifact, types.Enchantment},
									StackObjectKinds:  []game.StackObjectKind{game.StackSpell},
								},
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.CounterObject{
									Object:       game.TargetStackObjectReference(0),
									ExileInstead: true,
								},
							},
						},
					},
					game.Mode{
						Text: "Counter target activated or triggered ability.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target activated or triggered ability",
								Allow:      game.TargetAllowStackObject,
								Predicate: game.TargetPredicate{
									StackObjectKinds: []game.StackObjectKind{game.StackActivatedAbility, game.StackTriggeredAbility},
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
					},
				},
				MinModes: 1,
				MaxModes: 1,
			}),
			OracleText: `
			Choose one —
			• Counter target artifact or enchantment spell. If that spell is countered this way, exile it instead of putting it into its owner's graveyard.
			• Counter target activated or triggered ability.
		`,
		},
	}
}
