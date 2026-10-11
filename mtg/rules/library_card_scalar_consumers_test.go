package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestLibraryCardScalarAvailabilityDoesNotUseObservationCount(t *testing.T) {
	for _, outcome := range []string{"zero", "unknown", "empty", "declined", "skipped"} {
		for _, negate := range []bool{false, true} {
			t.Run(outcome+map[bool]string{false: "/available", true: "/negated"}[negate], func(t *testing.T) {
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				g.Players[game.Player1].Life = 20
				if outcome != "empty" {
					cardID := addCardToLibrary(g, game.Player1, vanillaCreature("Observed", 0, 3))
					if outcome == "unknown" {
						g.CardInstances[cardID].Def.Power = opt.Val(game.PT{IsStar: true})
					}
				}
				obj := &game.StackObject{ID: g.IDGen.Next(), Controller: game.Player1}
				observe := game.Instruction{
					Primitive: game.Reveal{
						Player: game.ControllerReference(), Amount: game.Fixed(1), PublishLinked: "observed",
						PublishCharacteristics: game.LibraryCardCharacteristics{Power: "power"},
					},
					PublishResult: "observed-count",
				}
				if outcome == "declined" {
					observe.Optional = true
				}
				if outcome == "skipped" {
					observe.ConditionGate = "absent"
				}
				resolver := newEffectResolver(NewEngine(nil), g, obj,
					[game.NumPlayers]PlayerAgent{game.Player1: &libraryPaymentAgent{}}, &TurnLog{})
				resolver.resolveInstruction(&observe)
				resolver.resolveInstruction(&game.Instruction{
					Primitive:  game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(4)},
					ResultGate: opt.Val(game.InstructionResultGate{Key: "power", AmountAvailable: true, Negate: negate}),
				})
				want := 20
				if outcome == "zero" && !negate {
					want += 4
				}
				if g.Players[game.Player1].Life != want {
					t.Fatalf("life=%d, want %d: availability inferred from observation count, missing zero, or negated absence",
						g.Players[game.Player1].Life, want)
				}
				if outcome == "unknown" && !obj.ResolutionResults["observed-count"].Succeeded {
					t.Fatal("unavailable power was conflated with an unsuccessful observation")
				}
			})
		}
	}
}

func TestLibraryCardFrozenScalarsSurviveDrawButLivePredicatesReevaluate(t *testing.T) {
	for _, owner := range []game.PlayerID{game.Player1, game.Player3} {
		t.Run(map[game.PlayerID]string{game.Player1: "controller library", game.Player3: "other library"}[owner], func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			g.Players[game.Player1].Life = 20
			g.Players[game.Player3].Life = 30
			for range 3 {
				addCardToLibrary(g, game.Player1, vanillaCreature("Controller Decoy", 99, 99))
			}
			cardID := addCardToLibrary(g, owner, vanillaCreature("Observed", 2, 3))
			obj := &game.StackObject{
				ID: g.IDGen.Next(), Controller: game.Player1,
				Targets: []game.Target{game.PlayerTarget(game.Player2), game.PlayerTarget(owner)},
			}
			resolver := newEffectResolver(NewEngine(nil), g, obj, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
			resolver.resolveInstruction(&game.Instruction{Primitive: game.Reveal{
				Player: game.TargetPlayerReference(1), Amount: game.Fixed(1), PublishLinked: "observed",
				PublishCharacteristics: game.LibraryCardCharacteristics{Power: "power", Toughness: "toughness"},
			}})
			condition := opt.Val(game.EffectCondition{Condition: opt.Val(game.Condition{
				Object:        opt.Val(game.LinkedObjectReference("observed")),
				ObjectMatches: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
			})})
			resolver.resolveInstruction(&game.Instruction{
				Primitive: game.Draw{Player: game.ControllerReference(), Amount: game.Dynamic(game.DynamicAmount{
					Kind: game.DynamicAmountPreviousEffectResult, ResultKey: "power", Multiplier: 1,
				})},
				Condition: condition, PublishCondition: "creature-group",
				ResultGate: opt.Val(game.InstructionResultGate{Key: "power", AmountAvailable: true}),
			})
			resolver.resolveInstruction(&game.Instruction{
				Primitive: game.GainLife{Player: game.ControllerReference(), Amount: game.Dynamic(game.DynamicAmount{
					Kind: game.DynamicAmountPreviousEffectResult, ResultKey: "toughness", Multiplier: 1,
				})},
				ConditionGate: "creature-group",
				ResultGate:    opt.Val(game.InstructionResultGate{Key: "toughness", AmountAvailable: true}),
			})
			resolver.resolveInstruction(&game.Instruction{
				Primitive: game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(7)},
				Condition: condition,
			})
			want := 23
			if owner != game.Player1 {
				want += 7
			}
			if g.Players[game.Player1].Life != want || g.Players[game.Player1].Hand.Size() != 2 ||
				g.Players[game.Player3].Life != 30 ||
				g.Players[owner].Library.Contains(cardID) != (owner != game.Player1) {
				t.Fatalf("wrong amount subject, draw actor, or predicate timing: life=%d hand=%d observed remains=%v",
					g.Players[game.Player1].Life, g.Players[game.Player1].Hand.Size(), g.Players[owner].Library.Contains(cardID))
			}
		})
	}
}
