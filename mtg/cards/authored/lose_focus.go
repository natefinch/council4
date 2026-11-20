package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

var LoseFocus = cardset.Entry{
	OracleID: "1cea6439-7ae5-4887-8c33-7da9fb36e2d4",
	Name:     "Lose Focus",
	New:      newLoseFocus,
	Verification: "mtg/rules/lose_focus_test.go: TestLoseFocusReplicateScenario; " +
		"printing 985bdb0c-ce6c-4506-8163-76f3b2fdf5fb; Scryfall WotC rulings 2021-06-18; " +
		"CR 2026-11-06 601.2b/c/f-i, 603.2-3, 608.2b/g/h, 702.56a-b, 707.10/.10c",
}

func newLoseFocus() *game.CardDef {
	const replicateText = "Replicate {U} (When you cast this spell, copy it for each time you paid its replicate cost. You may choose new targets for the copies.)"
	const counterText = "Counter target spell unless its controller pays {2}."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name:       "Lose Focus",
			OracleText: replicateText + "\n" + counterText,
			ManaCost:   opt.Val(cost.Mana{cost.O(1), cost.U}),
			Types:      []types.Card{types.Instant},
			Colors:     []color.Color{color.Blue},
			StaticAbilities: []game.StaticAbility{{
				Text: replicateText,
				KeywordAbilities: []game.KeywordAbility{
					game.ReplicateKeyword{Cost: cost.Mana{cost.U}},
				},
			}},
			SpellAbility: opt.Val(game.Mode{
				Text: counterText,
				Targets: []game.TargetSpec{{
					MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowStackObject,
					Predicate: game.TargetPredicate{StackObjectKinds: []game.StackObjectKind{game.StackSpell}},
				}},
				Sequence: []game.Instruction{
					{
						Primitive: game.Pay{Payment: game.ResolutionPayment{
							Prompt:   "Pay {2}?",
							Payer:    opt.Val(game.ObjectControllerReference(game.TargetStackObjectReference(0))),
							ManaCost: opt.Val(cost.Mana{cost.O(2)}),
						}},
						PublishResult: "lose-focus-paid",
					},
					{
						Primitive:  game.CounterObject{Object: game.TargetStackObjectReference(0)},
						ResultGate: opt.Val(game.InstructionResultGate{Key: "lose-focus-paid", Succeeded: game.TriFalse}),
					},
				},
			}.Ability()),
		},
	}
}
