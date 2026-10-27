package rules

import (
	"fmt"
	"testing"

	"github.com/natefinch/council4/internal/carddata"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func redirectResultCardToHand(g *game.Game, cardID id.ID) {
	g.ReplacementEffects = append(g.ReplacementEffects, game.ReplacementEffect{
		ID: g.IDGen.Next(), Controller: game.Player1, MatchEvent: game.EventZoneChanged,
		MatchFromZone: true, FromZone: zone.Library,
		MatchToZone: true, ToZone: zone.Graveyard,
		AffectedCardID: cardID, ReplaceToZone: zone.Hand,
	})
}

func TestResultObjectAvailabilityKnownEmptyAndIncomplete(t *testing.T) {
	for _, tc := range []struct {
		name     string
		count    int
		hidden   bool
		matching bool
		missing  bool
	}{
		{"known zero", 0, false, false, false},
		{"known filter zero", 2, false, false, false},
		{"known qualifying", 2, false, true, false},
		{"hidden whole", 1, true, true, false},
		{"hidden partial", 2, true, true, false},
		{"missing capture whole", 1, false, true, true},
		{"missing capture partial", 2, false, true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			for i := range tc.count {
				cardType := types.Land
				if tc.matching {
					cardType = types.Creature
				}
				cardID := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
					Name: fmt.Sprintf("Member %d", i), Types: []types.Card{cardType},
				}})
				if tc.hidden && i == 0 {
					redirectResultCardToHand(g, cardID)
				}
				if tc.missing && i == 0 {
					g.CardInstances[cardID].Def = nil
				}
			}
			sourceID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Publisher"}})
			obj := &game.StackObject{Controller: game.Player1, SourceID: sourceID}
			key := game.ResultKey("available")
			publisher := game.Instruction{
				Primitive: game.MoveTopOfLibrary{Player: game.ControllerReference(), Amount: game.Fixed(3),
					Destination: zone.Graveyard, PublishLinked: "graveyard-only"},
				PublishResult: key,
			}
			engine := NewEngine(nil)
			engine.resolveInstructionWithChoices(g, obj, &publisher, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
			snapshots, available := obj.ResolutionResultObjects[string(key)]
			unavailable := tc.hidden || tc.missing
			if available == unavailable || !unavailable && len(snapshots) != tc.count {
				t.Fatalf("snapshot available=%t size=%d; unavailable=%t actual=%d", available, len(snapshots), unavailable, tc.count)
			}
			receipt := obj.ResolutionResults[string(key)]
			if !receipt.Accepted || receipt.Succeeded != (tc.count > 0) ||
				receipt.Amount != tc.count || obj.ResolvedAmounts[string(key)] != tc.count {
				t.Fatalf("characteristic availability corrupted action receipt: %#v", receipt)
			}
			agent := &choiceOnlyAgent{choices: [][]int{{1}}}
			reader := actualCardQuantityReader(key, 1)
			reader.Optional, reader.PublishResult = true, "consumer"
			before := g.Players[game.Player1].Life
			engine.resolveInstructionWithChoices(g, obj, &reader,
				[game.NumPlayers]PlayerAgent{game.Player1: agent}, &TurnLog{})
			_, executed := obj.ResolutionResults["consumer"]
			wantLife, wantPrompts := before, 0
			if !unavailable {
				wantLife++
				if tc.matching {
					wantLife += tc.count
				}
				wantPrompts = 1
			}
			if executed == unavailable || agent.next != wantPrompts || g.Players[game.Player1].Life != wantLife {
				t.Fatalf("executed=%t prompts=%d life=%d, want executed=%t prompts=%d life=%d",
					executed, agent.next, g.Players[game.Player1].Life, !unavailable, wantPrompts, wantLife)
			}
			linked := linkedObjects(g, linkedObjectSourceKey(g, obj, "graveyard-only"))
			wantLinked := tc.count
			if tc.hidden {
				wantLinked--
			}
			if len(linked) != wantLinked {
				t.Fatal("characteristic completeness changed the strict graveyard-only link batch")
			}
		})
	}
}

func TestCompiledHiddenMilledQuantityUnavailable(t *testing.T) {
	for _, noun := range []string{"creature card", "card"} {
		t.Run(noun, func(t *testing.T) {
			def := legacyCard(t, carddata.ScryfallCard{
				Name: "Unavailable quantity", Layout: "normal", TypeLine: "Sorcery",
				OracleText: "Mill a card. For each " + noun + " milled this way, you gain 2 life.",
			})
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			cardID := addCardToLibrary(g, game.Player1, vanillaCreature("Hidden", 1, 1))
			redirectResultCardToHand(g, cardID)
			obj := &game.StackObject{Controller: game.Player1}
			sequence := append([]game.Instruction(nil), def.SpellAbility.Val.Modes[0].Sequence...)
			sequence[1].PublishResult = "consumer"
			restore := resolveSequenceInOwningFrame(g, obj, sequence)
			defer restore()
			key := string(sequence[0].PublishResult)
			if _, present := obj.ResolutionResultObjects[key]; present {
				t.Fatal("hidden batch was certified complete")
			}
			if _, executed := obj.ResolutionResults["consumer"]; executed {
				t.Fatal("unavailable characteristic count executed as zero")
			}
			if obj.ResolvedAmounts[key] != 1 || !obj.ResolutionResults[key].Succeeded ||
				g.Players[game.Player1].Life != 40 || !g.Players[game.Player1].Hand.Contains(cardID) {
				t.Fatal("unavailable quantity corrupted the actual action or scalar receipt")
			}
		})
	}
}

func TestResultObjectUnavailableDoesNotInvertOrSuppressActorReceipts(t *testing.T) {
	for _, present := range []bool{false, true} {
		for _, succeeded := range []bool{false, true} {
			t.Run(fmt.Sprintf("present=%t succeeded=%t", present, succeeded), func(t *testing.T) {
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				obj := &game.StackObject{
					Controller:        game.Player1,
					ResolutionResults: map[string]game.InstructionResolutionResult{"result": {Accepted: true, Succeeded: succeeded, Amount: 1}},
					ResolvedAmounts:   map[string]int{"result": 1},
				}
				if present {
					obj.ResolutionResultObjects = map[string][]game.ObjectSnapshot{"result": nil}
				}
				filter := game.InstructionResultGate{Key: "result", Succeeded: game.TriTrue,
					ObjectSelection: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}), Negate: true}
				if got := instructionResultGateSatisfied(g, obj, filter); got != present {
					t.Fatalf("negated empty filter=%t, want known complete=%t", got, present)
				}
				for _, gate := range []game.InstructionResultGate{
					{Key: "result", Accepted: game.TriTrue},
					{Key: "result", Succeeded: map[bool]game.TriState{true: game.TriTrue, false: game.TriFalse}[succeeded]},
				} {
					if !instructionResultGateSatisfied(g, obj, gate) {
						t.Fatal("snapshot availability suppressed an unfiltered actor/action gate")
					}
				}
				reader := game.Instruction{Primitive: game.GainLife{Player: game.ControllerReference(), Amount: game.Dynamic(game.DynamicAmount{
					Kind: game.DynamicAmountPreviousEffectResult, ResultKey: "result",
				})}}
				NewEngine(nil).resolveInstructionWithChoices(g, obj, &reader, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
				if g.Players[game.Player1].Life != 41 {
					t.Fatal("nil-Selection scalar stopped reading its independent action amount")
				}
			})
		}
	}
}

func TestResultObjectUnavailableProvidersAndCaptures(t *testing.T) {
	for _, primitive := range []game.Primitive{
		game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(1)},
		game.Draw{Player: game.ControllerReference(), Amount: game.Fixed(1)},
		nil,
	} {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		obj := &game.StackObject{Controller: game.Player1}
		r := newEffectResolver(NewEngine(nil), g, obj, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
		r.resultObjectsKnown = true
		r.publishInstructionResult(&game.Instruction{Primitive: primitive, PublishResult: "scalar"},
			effectResolved{accepted: true, succeeded: true, amount: 1})
		if _, minted := obj.ResolutionResultObjects["scalar"]; minted {
			t.Fatal("untracked/scalar/nil provider minted an available empty snapshot batch")
		}
		if obj.ResolvedAmounts["scalar"] != 1 {
			t.Fatal("untracked scalar receipt changed")
		}
	}
	for _, ref := range []game.LinkedObjectRef{{CardID: 100}, {ObjectID: 100}} {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		obj := &game.StackObject{Controller: game.Player1}
		publisher := game.Instruction{Primitive: game.MoveTopOfLibrary{Player: game.ControllerReference(),
			Amount: game.Fixed(1), Destination: zone.Graveyard}, PublishResult: "capture"}
		r := newEffectResolver(NewEngine(nil), g, obj, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
		r.currentInstruction = &publisher
		r.rememberResultObject(ref)
		r.publishInstructionResult(&publisher, effectResolved{accepted: true, succeeded: true, amount: 1})
		if _, present := obj.ResolutionResultObjects["capture"]; present || !r.resultObjectsUnavailable {
			t.Fatal("missing card/LKI capture was certified as known empty")
		}
	}
	for _, faceDown := range []bool{false, true} {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		cardID := addCardToGraveyard(g, game.Player1, vanillaCreature("Unavailable actual capture", 1, 1))
		if faceDown {
			g.Players[game.Player1].Graveyard.Remove(cardID)
			g.Players[game.Player1].Exile.Add(cardID)
			g.Players[game.Player1].Exile.SetFaceDown(cardID, true)
		} else {
			g.CardInstances[cardID].Def = nil
		}
		obj := &game.StackObject{Controller: game.Player1}
		publisher := game.Instruction{Primitive: game.MoveTopOfLibrary{Destination: zone.Graveyard}, PublishResult: "capture"}
		r := newEffectResolver(NewEngine(nil), g, obj, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
		r.currentInstruction = &publisher
		r.rememberResultCard(cardID)
		r.publishInstructionResult(&publisher, effectResolved{accepted: true, succeeded: true, amount: 1})
		if _, present := obj.ResolutionResultObjects["capture"]; present || !r.resultObjectsUnavailable {
			t.Fatal("missing definition or face-down actual card capture minted known-empty availability")
		}
		if obj.ResolvedAmounts["capture"] != 1 || !obj.ResolutionResults["capture"].Succeeded {
			t.Fatal("unavailable capture erased its independent actual action receipt")
		}
	}
}

func TestResultObjectAvailabilityClearedBeforeSkippedAndDeclinedProducers(t *testing.T) {
	for _, skipped := range []bool{false, true} {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		obj := &game.StackObject{Controller: game.Player1,
			ResolutionResultObjects: map[string][]game.ObjectSnapshot{"result": {{CardID: 100, Types: []types.Card{types.Creature}}}}}
		publisher := game.Instruction{Primitive: game.Discard{Player: game.ControllerReference(), Amount: game.Fixed(1)},
			PublishResult: "result", Optional: !skipped}
		if skipped {
			publisher.ConditionGate = "absent"
		}
		NewEngine(nil).resolveInstructionWithChoices(g, obj, &publisher,
			[game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{0}}}}, &TurnLog{})
		snapshots, present := obj.ResolutionResultObjects["result"]
		if present == skipped || len(snapshots) != 0 {
			t.Fatal("skipped producer borrowed older snapshots or explicit decline lost its known-empty batch")
		}

		if skipped {
			if _, present := obj.ResolutionResults["result"]; present {
				t.Fatal("skipped producer minted an action receipt")
			}
		} else if obj.ResolutionResults["result"].Accepted {
			t.Fatal("declined action receipt lost its acceptance state")
		}
		reader := actualCardQuantityReader("result", 1)
		reader.PublishResult, reader.Optional = "consumer", true
		agent := &choiceOnlyAgent{choices: [][]int{{1}}}
		NewEngine(nil).resolveInstructionWithChoices(g, obj, &reader,
			[game.NumPlayers]PlayerAgent{game.Player1: agent}, &TurnLog{})
		if _, executed := obj.ResolutionResults["consumer"]; executed || agent.next != 0 {
			t.Fatal("skipped or declined producer allowed the addend quantity to prompt or execute")
		}
		if !skipped {
			filter := game.InstructionResultGate{Key: "result", Succeeded: game.TriTrue,
				ObjectSelection: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}})}
			if instructionResultGateSatisfied(g, obj, filter) {
				t.Fatal("explicit declined empty batch matched a positive characteristic predicate")
			}
			filter.Negate = true
			if !instructionResultGateSatisfied(g, obj, filter) {
				t.Fatal("explicit declined known-empty batch lost its filtered predicate complement")
			}
		}
	}
}

func TestResultObjectUnresolvedOptionalActorIsUnavailable(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCardToHand(g, game.Player1, vanillaCreature("Undiscarded", 1, 1))
	obj := &game.StackObject{Controller: game.Player1}
	publisher := game.Instruction{
		Primitive:     game.Discard{Player: game.ControllerReference(), Amount: game.Fixed(1)},
		PublishResult: "actor", Optional: true,
		OptionalActor: opt.Val(game.TargetPlayerReference(0)),
	}
	engine := NewEngine(nil)
	engine.resolveInstructionWithChoices(g, obj, &publisher, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if _, present := obj.ResolutionResultObjects["actor"]; present {
		t.Fatal("unresolved optional actor was treated as an explicit decline")
	}
	if obj.ResolutionResults["actor"].Accepted || g.Players[game.Player1].Hand.Size() != 1 {
		t.Fatal("unresolved actor performed the action or lost its not-accepted receipt")
	}
	filter := game.InstructionResultGate{Key: "actor", Succeeded: game.TriTrue,
		ObjectSelection: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}), Negate: true}
	if instructionResultGateSatisfied(g, obj, filter) {
		t.Fatal("unavailable optional actor was inverted into known-empty filter truth")
	}
}

func TestResultObjectPreflightRequiresEveryTypedBatch(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	obj := &game.StackObject{Controller: game.Player1,
		ResolutionResults:       map[string]game.InstructionResolutionResult{"known": {Accepted: true}},
		ResolutionResultObjects: map[string][]game.ObjectSnapshot{"known": nil}}
	selection := game.Selection{RequiredTypes: []types.Card{types.Creature}}
	reader := game.Instruction{
		Primitive: game.GainLife{Player: game.ControllerReference(), Amount: game.Dynamic(game.DynamicAmount{
			Kind: game.DynamicAmountMaxOf, Operands: []game.DynamicAmount{
				{Kind: game.DynamicAmountPreviousEffectResult, ResultKey: "known", Selection: &selection},
				{Kind: game.DynamicAmountPreviousEffectResult, ResultKey: "missing", Selection: &selection},
			},
		})},
		PublishResult: "consumer", Optional: true,
		ResultGate: opt.Val(game.InstructionResultGate{Key: "known", Accepted: game.TriTrue}),
	}
	agent := &choiceOnlyAgent{choices: [][]int{{1}}}
	NewEngine(nil).resolveInstructionWithChoices(g, obj, &reader,
		[game.NumPlayers]PlayerAgent{game.Player1: agent}, &TurnLog{})
	if _, executed := obj.ResolutionResults["consumer"]; executed || agent.next != 0 {
		t.Fatal("one available quantity operand allowed another unavailable batch to execute or prompt")
	}
}

func TestResultObjectActedIllegalTargetKnownEmpty(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	obj := &game.StackObject{Controller: game.Player1}
	publisher := game.Instruction{Primitive: game.MoveCard{
		Card: game.CardReference{Kind: game.CardReferenceTarget}, FromZone: zone.Graveyard, Destination: zone.Exile,
	}, PublishResult: "illegal"}
	engine := NewEngine(nil)
	engine.resolveInstructionWithChoices(g, obj, &publisher, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if snapshots, present := obj.ResolutionResultObjects["illegal"]; !present || len(snapshots) != 0 {
		t.Fatal("executed all-illegal card move was unavailable instead of known zero actual members")
	}
	reader := actualCardQuantityReader("illegal", 1)
	engine.resolveInstructionWithChoices(g, obj, &reader, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if g.Players[game.Player1].Life != 41 {
		t.Fatal("known zero illegal move did not evaluate the addend")
	}
}
