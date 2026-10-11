package rules

import (
	"fmt"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func TestResultObjectGroupAvailabilityStickyAndActorReceiptIndependent(t *testing.T) {
	for _, optional := range []bool{false, true} {
		for _, hiddenPlayer := range []game.PlayerID{game.Player2, game.Player3} {
			t.Run(fmt.Sprintf("optional=%t hidden=%d", optional, hiddenPlayer), func(t *testing.T) {
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				for _, player := range []game.PlayerID{game.Player2, game.Player3} {
					cardID := addCardToLibrary(g, player, vanillaCreature("Member", 1, 1))
					if player == hiddenPlayer {
						redirectResultCardToHand(g, cardID)
					}
				}
				obj := &game.StackObject{Controller: game.Player1,
					Targets: []game.Target{game.PlayerTarget(game.Player2), game.PlayerTarget(game.Player3)}}
				publisher := game.Instruction{
					Primitive: game.MoveTopOfLibrary{Player: game.GroupOfferMemberReference(),
						Amount: game.Fixed(1), Destination: zone.Graveyard},
					PublishResult: "group",
				}
				if optional {
					publisher.Optional = true
					publisher.OptionalActorGroup = opt.Val(game.TargetedPlayersReference())
				} else {
					publisher.ForEachPlayerGroup = opt.Val(game.TargetedPlayersReference())
				}
				agents := [game.NumPlayers]PlayerAgent{
					game.Player2: &choiceOnlyAgent{choices: [][]int{{1}}},
					game.Player3: &choiceOnlyAgent{choices: [][]int{{1}}},
				}
				engine := NewEngine(nil)
				engine.resolveInstructionWithChoices(g, obj, &publisher, agents, &TurnLog{})
				if _, available := obj.ResolutionResultObjects["group"]; available {
					t.Fatal("group publication exposed a partial batch after an unavailable member")
				}
				receipt := obj.ResolutionResults["group"]
				if !receipt.Accepted || !receipt.Succeeded || receipt.Amount != 2 ||
					receipt.AcceptedActors.Count() != 2 || obj.ResolvedAmounts["group"] != 2 {
					t.Fatalf("completeness replaced acted/accepting-player receipt: %#v", receipt)
				}
				reader := actualCardQuantityReader("group", 1)
				reader.PublishResult = "consumer"
				engine.resolveInstructionWithChoices(g, obj, &reader, agents, &TurnLog{})
				if _, executed := obj.ResolutionResults["consumer"]; executed {
					t.Fatal("group's unavailable characteristic quantity ran as a partial count")
				}
				actorReader := game.Instruction{
					Primitive:  game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(3)},
					ResultGate: opt.Val(game.InstructionResultGate{Key: "group", Succeeded: game.TriTrue}),
				}
				engine.resolveInstructionWithChoices(g, obj, &actorReader, agents, &TurnLog{})
				if g.Players[game.Player1].Life != 43 {
					t.Fatal("group characteristic absence suppressed its independent action gate")
				}
				empty := game.Instruction{Primitive: game.MoveTopOfLibrary{Player: game.ControllerReference(),
					Amount: game.Fixed(1), Destination: zone.Graveyard}, PublishResult: "next"}
				engine.resolveInstructionWithChoices(g, obj, &empty, agents, &TurnLog{})
				if snapshots, present := obj.ResolutionResultObjects["next"]; !present || len(snapshots) != 0 {
					t.Fatal("group incompleteness leaked across instruction boundary")
				}
			})
		}
	}
}

func TestResultObjectFrameRestoresPresenceAcrossRepeat(t *testing.T) {
	for _, outer := range []string{"absent", "empty", "positive"} {
		t.Run(outer, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			cardID := addCardToLibrary(g, game.Player1, vanillaCreature("Hidden inner", 1, 1))
			redirectResultCardToHand(g, cardID)
			obj := &game.StackObject{Controller: game.Player1,
				ResolutionResults: map[string]game.InstructionResolutionResult{"scope": {Accepted: true, Succeeded: true}},
				ResolvedAmounts:   map[string]int{"scope": 99},
			}
			if outer != "absent" {
				obj.ResolutionResultObjects = map[string][]game.ObjectSnapshot{"scope": nil}
			}
			if outer == "positive" {
				obj.ResolutionResultObjects["scope"] = []game.ObjectSnapshot{{CardID: 100, Types: []types.Card{types.Creature}}}
			}
			body := []game.Instruction{
				{Primitive: game.MoveTopOfLibrary{Player: game.ControllerReference(), Amount: game.Fixed(1), Destination: zone.Graveyard},
					PublishResult: "scope", LocalProducts: game.LocalProducts{Results: []game.ResultKey{"scope"}}},
				actualCardQuantityReader("scope", 1),
			}
			NewEngine(nil).resolveInstructionSequence(g, obj, []game.Instruction{
				{Primitive: game.RepeatProcess{Times: game.Fixed(2), Body: game.Mode{Sequence: body}.Ability()}},
			}, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
			snapshots, present := obj.ResolutionResultObjects["scope"]
			if present != (outer != "absent") || outer == "empty" && len(snapshots) != 0 ||
				outer == "positive" && (len(snapshots) != 1 || snapshots[0].CardID != 100) {
				t.Fatal("repeat did not restore exact outer present-empty/positive/absent cells")
			}
			if obj.ResolvedAmounts["scope"] != 99 || g.Players[game.Player1].Life != 41 {
				t.Fatal("repeat borrowed unavailable first iteration or suppressed second known-zero addend")
			}
		})
	}
}

func TestResultObjectNestedInstructionRestoresCompleteness(t *testing.T) {
	for _, outerUnavailable := range []bool{false, true} {
		t.Run(fmt.Sprintf("outerUnavailable=%t", outerUnavailable), func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			obj := &game.StackObject{Controller: game.Player1}
			parent := game.Instruction{Primitive: game.MoveTopOfLibrary{Destination: zone.Graveyard}, PublishResult: "parent"}
			r := newEffectResolver(NewEngine(nil), g, obj, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
			r.currentInstruction = &parent
			r.resultObjectsKnown = true
			r.resultObjectsUnavailable = outerUnavailable
			if !outerUnavailable {
				cardID := addCardToLibrary(g, game.Player1, vanillaCreature("Hidden child", 1, 1))
				redirectResultCardToHand(g, cardID)
			}

			r.resolveInstruction(&game.Instruction{
				Primitive:     game.MoveTopOfLibrary{Player: game.ControllerReference(), Amount: game.Fixed(1), Destination: zone.Graveyard},
				PublishResult: "child",
			})
			if r.currentInstruction != &parent || r.resultObjectsUnavailable != outerUnavailable {
				t.Fatal("nested instruction lost the parent's private completeness state")
			}
			r.publishInstructionResult(&parent, effectResolved{accepted: true})
			if _, present := obj.ResolutionResultObjects["parent"]; present == outerUnavailable {
				t.Fatal("nested incompleteness infected or cleared the parent's availability")
			}
		})
	}
}

func TestResultObjectGroupZeroAndDeclineAvailability(t *testing.T) {
	for _, accept := range []bool{false, true} {
		t.Run(fmt.Sprintf("accept=%t", accept), func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			obj := &game.StackObject{Controller: game.Player1,
				Targets: []game.Target{game.PlayerTarget(game.Player2), game.PlayerTarget(game.Player3)}}
			publisher := game.Instruction{
				Primitive: game.MoveTopOfLibrary{Player: game.GroupOfferMemberReference(),
					Amount: game.Fixed(1), Destination: zone.Graveyard},
				PublishResult: "group", Optional: true,
				OptionalActorGroup: opt.Val(game.TargetedPlayersReference()),
			}
			choice := 0
			if accept {
				choice = 1
			}
			agents := [game.NumPlayers]PlayerAgent{
				game.Player2: &choiceOnlyAgent{choices: [][]int{{choice}}},
				game.Player3: &choiceOnlyAgent{choices: [][]int{{choice}}},
			}
			engine := NewEngine(nil)
			engine.resolveInstructionWithChoices(g, obj, &publisher, agents, &TurnLog{})
			snapshots, present := obj.ResolutionResultObjects["group"]
			receipt := obj.ResolutionResults["group"]
			if !present || len(snapshots) != 0 || receipt.Accepted != accept ||
				receipt.Succeeded || receipt.AcceptedActors.Count() != 2*choice || receipt.Amount != 2*choice {
				t.Fatalf("presence=%t receipt=%#v, want complete no-action or accepted zero group", present, receipt)
			}
			reader := actualCardQuantityReader("group", 1)
			reader.PublishResult = "consumer"
			engine.resolveInstructionWithChoices(g, obj, &reader, agents, &TurnLog{})
			if _, executed := obj.ResolutionResults["consumer"]; executed != accept {
				t.Fatal("accepted known-zero group and declined group shared a quantity behavior")
			}
			if g.Players[game.Player1].Life != 40+choice {
				t.Fatal("known-zero group addend or declined quantity availability changed")
			}
		})
	}
}
