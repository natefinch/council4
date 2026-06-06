package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ContestOfClaws is the card definition for Contest of Claws.
//
// Type: Sorcery
// Cost: {1}{G}
//
// Oracle text:
//
//	Target creature you control deals damage equal to its power to another target creature. If excess damage was dealt this way, discover X, where X is that excess damage.
var ContestOfClaws = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Green),
	CardFace: game.CardFace{
		Name: "Contest of Claws",
		ManaCost: opt.Val(cost.Mana{
			cost.O(1),
			cost.G,
		}),
		Colors: []color.Color{color.Green},
		Types:  []types.Card{types.Sorcery},
		OracleText: `
			Target creature you control deals damage equal to its power to another target creature. If excess damage was dealt this way, discover X, where X is that excess damage. (Exile cards from the top of your library until you exile a nonland card with that mana value or less. Cast it without paying its mana cost or put it into your hand. Put the rest on the bottom in a random order.)
		`,
		SpellAbility: opt.Val(
			game.SpellAbilityBody{
				Text: `
					Target creature you control deals damage equal to its power to another target creature. If excess damage was dealt this way, discover X, where X is that excess damage.
				`,
				Content: game.PlainAbilityContent{
					Targets: []game.TargetSpec{
						{
							MinTargets: 1,
							MaxTargets: 1,
							Constraint: "creature you control",
							Allow:      game.TargetAllowPermanent,
							Predicate: game.TargetPredicate{
								PermanentTypes: []types.Card{
									types.Creature,
								},
								Controller: game.ControllerYou,
							},
						},
						{
							MinTargets: 1,
							MaxTargets: 1,
							Constraint: "another target creature",
							Allow:      game.TargetAllowPermanent,
							Predicate: game.TargetPredicate{
								PermanentTypes: []types.Card{
									types.Creature,
								},
								Another: true,
							},
						},
					},
					Sequence: []game.Instruction{
						{
							Primitive: game.Damage{
								Amount: game.Dynamic(game.DynamicAmount{
									Kind:        game.DynamicAmountTargetPower,
									TargetIndex: 0,
								}),
								Recipient: game.TargetRecipient(1),
								DamageSource: opt.Val(game.ObjectReference{
									Kind:        game.ObjectReferenceTargetPermanent,
									TargetIndex: 0,
								}),
								ResultAmountKind: game.EffectResultAmountExcessDamage,
							},
							PublishResult: game.ResultKey("excess"),
						},
						{
							Primitive: game.DiscoverCards{
								Amount: game.Dynamic(game.DynamicAmount{
									Kind:      game.DynamicAmountPreviousEffectExcessDamage,
									ResultKey: game.ResultKey("excess"),
								}),
							},
							ResultGate: opt.Val(game.InstructionResultGate{
								Key:       game.ResultKey("excess"),
								Succeeded: game.TriTrue,
							}),
						},
					},
				},
			},
		),
	},
}
