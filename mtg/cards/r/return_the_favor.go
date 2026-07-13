package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ReturnTheFavor is the card definition for Return the Favor.
//
// Type: Instant
// Cost: {R}{R}
//
// Oracle text:
//
//	Spree (Choose one or more additional costs.)
//	+ {1} — Copy target instant spell, sorcery spell, activated ability, or triggered ability. You may choose new targets for the copy.
//	+ {1} — Change the target of target spell or ability with a single target.
var ReturnTheFavor = newReturnTheFavor

func newReturnTheFavor() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Return the Favor",
			ManaCost: opt.Val(cost.Mana{
				cost.R,
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "{1} — Copy target instant spell, sorcery spell, activated ability, or triggered ability. You may choose new targets for the copy.",
						Cost: opt.Val(cost.Mana{cost.O(1)}),
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target instant spell, sorcery spell, activated ability, or triggered ability",
								Allow:      game.TargetAllowStackObject,
								Predicate: game.TargetPredicate{
									SpellCardTypesAny: []types.Card{types.Instant, types.Sorcery},
									StackObjectKinds:  []game.StackObjectKind{game.StackSpell, game.StackActivatedAbility, game.StackTriggeredAbility},
								},
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.CopyStackObject{
									Object:              game.TargetStackObjectReference(0),
									MayChooseNewTargets: true,
								},
							},
						},
					},
					game.Mode{
						Text: "{1} — Change the target of target spell or ability with a single target.",
						Cost: opt.Val(cost.Mana{cost.O(1)}),
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target spell or ability",
								Allow:      game.TargetAllowStackObject,
								Predicate: game.TargetPredicate{
									StackObjectKinds: []game.StackObjectKind{game.StackSpell, game.StackActivatedAbility, game.StackTriggeredAbility},
								},
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.ChooseNewTargets{
									Object: game.TargetStackObjectReference(0),
								},
							},
						},
					},
				},
				MinModes: 1,
				MaxModes: 2,
			}),
			OracleText: `
			Spree (Choose one or more additional costs.)
			+ {1} — Copy target instant spell, sorcery spell, activated ability, or triggered ability. You may choose new targets for the copy.
			+ {1} — Change the target of target spell or ability with a single target.
		`,
		},
	}
}
