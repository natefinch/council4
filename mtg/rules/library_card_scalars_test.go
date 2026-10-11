package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func TestLibraryCardCharacteristicPublicationDistinguishesKnownZeroAndUnavailable(t *testing.T) {
	for _, producer := range []string{"look", "reveal", "linked reveal"} {
		for _, outcome := range []string{"known zero", "undefined power", "star power", "missing definition", "empty", "skipped", "declined"} {
			t.Run(producer+"/"+outcome, func(t *testing.T) {
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				obj := &game.StackObject{ID: g.IDGen.Next(), Controller: game.Player1}
				if outcome != "empty" {
					cardID := addCardToLibrary(g, game.Player1, vanillaCreature("Observed", 0, 3))
					card := g.CardInstances[cardID]
					switch outcome {
					case "known zero", "skipped", "declined":
					case "undefined power":
						card.Def.Power = opt.V[game.PT]{}
					case "star power":
						card.Def.Power = opt.Val(game.PT{IsStar: true})
					case "missing definition":
						card.Def = nil
					default:
						t.Fatalf("unknown characteristic outcome %q", outcome)
					}
				}
				outputs := game.LibraryCardCharacteristics{Power: "power", Toughness: "toughness"}
				var primitive game.Primitive = game.LookAtLibraryTop{
					Player: game.ControllerReference(), PublishLinked: "observed", PublishCharacteristics: outputs,
				}
				if producer == "reveal" {
					primitive = game.Reveal{Player: game.ControllerReference(), Amount: game.Fixed(1),
						PublishLinked: "observed", PublishCharacteristics: outputs}
				}
				if producer == "linked reveal" {
					newEffectResolver(NewEngine(nil), g, obj, [game.NumPlayers]PlayerAgent{}, &TurnLog{}).
						resolveInstruction(&game.Instruction{Primitive: game.LookAtLibraryTop{
							Player: game.ControllerReference(), PublishLinked: "observed",
						}})
					primitive = game.Reveal{
						Card:          game.CardReference{Kind: game.CardReferenceLinked, LinkID: "observed"},
						PublishLinked: "observed", PublishCharacteristics: outputs,
					}
				}
				obj.ResolvedAmounts = map[string]int{"power": 91, "toughness": 92, "persistent": 93}
				obj.ResolutionResults = map[string]game.InstructionResolutionResult{"power": {Succeeded: true}, "toughness": {Succeeded: true}}
				instruction := game.Instruction{Primitive: primitive, PublishResult: "observation-count"}
				if outcome == "skipped" {
					instruction.ConditionGate = "absent"
				}
				var agents [game.NumPlayers]PlayerAgent
				if outcome == "declined" {
					instruction.Optional = true
					agents[game.Player1] = &repeatLandChoiceAgent{mayAnswers: []bool{false}}
				}
				newEffectResolver(NewEngine(nil), g, obj, agents, &TurnLog{}).resolveInstruction(&instruction)
				_, powerAvailable := obj.ResolvedAmounts["power"]
				_, powerReceipt := obj.ResolutionResults["power"]
				wantPower := outcome == "known zero"
				if powerAvailable != wantPower || powerReceipt != wantPower ||
					wantPower && obj.ResolvedAmounts["power"] != 0 {
					t.Fatalf("power=%d scalar=%t receipt=%t, want known zero=%t", obj.ResolvedAmounts["power"], powerAvailable, powerReceipt, wantPower)
				}
				_, toughnessAvailable := obj.ResolvedAmounts["toughness"]
				wantToughness := outcome == "known zero" || outcome == "undefined power" || outcome == "star power"
				if toughnessAvailable != wantToughness || wantToughness && obj.ResolvedAmounts["toughness"] != 3 {
					t.Fatalf("toughness availability=%t, want %t", toughnessAvailable, wantToughness)
				}
				if outcome == "undefined power" || outcome == "star power" {
					if receipt := obj.ResolutionResults["observation-count"]; !receipt.Succeeded || receipt.Amount != 1 {
						t.Fatal("unknown characteristic was conflated with failed observation")
					}
				}
				if obj.ResolvedAmounts["persistent"] != 93 {
					t.Fatal("characteristic cleanup cleared unrelated facts")
				}
			})
		}
	}
}

func TestLibraryCardCharacteristicScalarFreezesObservedValueAfterMove(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	cardID := addCardToLibrary(g, game.Player1, vanillaCreature("Observed", 2, 3))
	obj := &game.StackObject{ID: g.IDGen.Next(), Controller: game.Player1}
	resolver := newEffectResolver(NewEngine(nil), g, obj, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	resolver.resolveInstruction(&game.Instruction{Primitive: game.Reveal{
		Player: game.ControllerReference(), Amount: game.Fixed(1), PublishLinked: "observed",
		PublishCharacteristics: game.LibraryCardCharacteristics{Power: "power", Toughness: "toughness"},
	}})
	if !moveCardBetweenZonesWithPlacement(g, game.Player1, cardID, zone.Library, zone.Hand, false) {
		t.Fatal("could not move observed card")
	}
	g.CardInstances[cardID].Def = vanillaCreature("Later Definition", 99, 99)
	if _, available := resolveObjectReference(g, obj, game.LinkedObjectReference("observed")); available {
		t.Fatal("card predicate remained live after departure")
	}
	for key, want := range map[game.ResultKey]int{"power": 2, "toughness": 3} {
		got := dynamicAmountValue(g, obj, obj.Controller, game.DynamicAmount{
			Kind: game.DynamicAmountPreviousEffectResult, ResultKey: key,
		})
		if got != want {
			t.Fatalf("%s=%d, want frozen observed value %d", key, got, want)
		}
	}
}
