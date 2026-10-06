package rules

import (
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestCompiledOptionalGroupDecisionAndTiming(t *testing.T) {
	for _, tt := range []struct {
		name, text         string
		accept             []bool
		initialHand        bool
		wantHand, wantLife int
		wantChoices        int
	}{
		{"accept", "You may draw a card and gain 2 life. You gain 1 life.", []bool{true}, false, 1, 43, 1},
		{"decline", "You may draw a card and gain 2 life. You gain 1 life.", []bool{false}, false, 0, 41, 1},
		{"ineffective first", "You may discard a card and gain 2 life. You gain 1 life.", []bool{true}, false, 0, 43, 1},
		{"conditional once", "If you have no cards in hand, you may draw a card and gain 2 life. You gain 1 life.", []bool{true}, false, 1, 43, 1},
		{"conditional decline", "If you have no cards in hand, you may draw a card and gain 2 life. You gain 1 life.", []bool{false}, false, 0, 41, 1},
		{"false group", "If you have no cards in hand, you may draw a card and gain 2 life. You gain 1 life.", nil, true, 1, 41, 0},
		{"independent groups", "You may draw a card and gain 2 life. You gain 1 life. You may draw a card and gain 3 life.", []bool{false, true}, false, 1, 44, 2},
		{"actual filtered result", "You may discard a card and gain 2 life. If a land card was discarded this way, draw a card.", []bool{true}, true, 1, 42, 1},
		{"failed filtered result", "You may discard a card and gain 2 life. If a land card was discarded this way, draw a card.", []bool{true}, false, 0, 42, 1},
		{"repeat fresh decision", "Repeat the following process two times. You may draw a card and gain 2 life.", []bool{true, false}, false, 1, 42, 2},
	} {
		t.Run(tt.name, func(t *testing.T) {
			content := compiledScopedResultContent(t, tt.text)
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			if tt.initialHand {
				addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Types: []types.Card{types.Land}}})
			}
			for range 3 {
				addCardToLibrary(g, game.Player1, evidenceCard("Drawn", 1))
			}
			agent := &scopedMayAgent{accept: tt.accept}
			agents := [game.NumPlayers]PlayerAgent{}
			agents[game.Player1] = agent
			NewEngine(nil).resolveAbilityContentWithChoices(g, &game.StackObject{Controller: game.Player1}, content, agents, &TurnLog{})
			if g.Players[game.Player1].Hand.Size() != tt.wantHand || g.Players[game.Player1].Life != tt.wantLife ||
				agent.next != tt.wantChoices || g.Players[game.Player2].Life != 40 {
				t.Fatalf("hand=%d life=%d choices=%d, want %d/%d/%d",
					g.Players[game.Player1].Hand.Size(), g.Players[game.Player1].Life, agent.next,
					tt.wantHand, tt.wantLife, tt.wantChoices)
			}
		})
	}
}

func TestCompiledExpandedOptionalActionKeepsNonzeroTargets(t *testing.T) {
	content := compiledScopedResultContent(t,
		"Tap target artifact. If you have no cards in hand, you may put a +1/+1 counter on each of up to two target creatures.")
	for _, tt := range []struct {
		name         string
		accept       bool
		firstMissing bool
		want         int
	}{
		{"accept", true, false, 1}, {"decline", false, false, 0}, {"first missing", true, true, 1},
	} {
		t.Run(tt.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			artifact := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Types: []types.Card{types.Artifact}}})
			first := addCombatCreaturePermanent(g, game.Player1, game.KeywordNone)
			second := addCombatCreaturePermanent(g, game.Player1, game.KeywordNone)
			if tt.firstMissing {
				removePermanentFromBattlefield(g, first.ObjectID)
			}
			agent := &scopedMayAgent{accept: []bool{tt.accept}}
			agents := [game.NumPlayers]PlayerAgent{}
			agents[game.Player1] = agent
			obj := &game.StackObject{Controller: game.Player1, Targets: []game.Target{
				game.PermanentTarget(artifact.ObjectID), game.PermanentTarget(first.ObjectID), game.PermanentTarget(second.ObjectID),
			}, TargetCounts: []int{1, 1, 1}}
			NewEngine(nil).resolveAbilityContentWithChoices(g, obj, content, agents, &TurnLog{})
			if !artifact.Tapped || artifact.Counters.Get(counter.PlusOnePlusOne) != 0 ||
				second.Counters.Get(counter.PlusOnePlusOne) != tt.want || agent.next != 1 {
				t.Fatalf("target isolation/one-choice failed: second=%d choices=%d", second.Counters.Get(counter.PlusOnePlusOne), agent.next)
			}
		})
	}
}

func TestOptionalDecisionAvailabilityAndInvalidation(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	agent := &scopedMayAgent{accept: []bool{true, false, true}}
	agents := [game.NumPlayers]PlayerAgent{}
	agents[game.Player1] = agent
	obj := &game.StackObject{Controller: game.Player1}
	resolver := newEffectResolver(NewEngine(nil), g, obj, agents, &TurnLog{})
	publisher := game.Instruction{
		Primitive: game.Discard{Player: game.ControllerReference(), Amount: game.Fixed(1)},
		Optional:  true, PublishOptionalDecision: "group", PublishResult: "actual",
	}
	consumer := game.Instruction{
		Primitive: game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(2)}, OptionalDecisionGate: "group",
	}
	resolver.resolveInstruction(&publisher)
	result := obj.ResolutionResults["actual"]
	if !resolver.optionalDecisions["group"] || result.Succeeded {
		t.Fatal("ineffective accepted action was conflated with success")
	}
	resolver.resolveInstruction(&consumer)
	resolver.resolveInstruction(&publisher)
	if accepted, available := resolver.optionalDecisions["group"]; accepted || !available {
		t.Fatal("actual decline was not a known false decision")
	}
	resolver.resolveInstruction(&consumer)
	resolver.resolveInstruction(&publisher)
	publisher.ResultGate = opt.Val(game.InstructionResultGate{Key: "unavailable", Succeeded: game.TriTrue})
	resolver.resolveInstruction(&publisher)
	if _, available := resolver.optionalDecisions["group"]; available {
		t.Fatal("skipped publisher retained stale acceptance")
	}
	resolver.resolveInstruction(&consumer)
	publisher.ResultGate = opt.V[game.InstructionResultGate]{}
	publisher.OptionalActor = opt.Val(game.ObjectControllerReference(game.TargetPermanentReference(9)))
	resolver.optionalDecisions["group"] = true
	resolver.resolveInstruction(&publisher)
	if _, available := resolver.optionalDecisions["group"]; available {
		t.Fatal("unavailable decider fabricated a declined or accepted answer")
	}
	NewEngine(nil).resolveInstructionSequence(g, obj, []game.Instruction{consumer}, agents, &TurnLog{})
	if g.Players[game.Player1].Life != 42 || agent.next != 3 {
		t.Fatal("unavailable, skipped, or stale decision enabled the group")
	}
}

func TestCompiledGustcloakGroupUsesEventIncarnation(t *testing.T) {
	def := compileUnlessCard(t, cardgen.ScryfallCard{
		Name: "Optional Combat Probe", Layout: "normal", TypeLine: "Creature",
		OracleText: "Whenever this creature becomes blocked, you may untap it and remove it from combat.",
	})
	for _, accept := range []bool{true, false} {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		event := addCombatCreaturePermanent(g, game.Player2, game.KeywordNone)
		other := addCombatCreaturePermanent(g, game.Player1, game.KeywordNone)
		event.Tapped = false
		other.Tapped = true
		g.Combat = &game.CombatState{Attackers: []game.AttackDeclaration{
			{Attacker: event.ObjectID, Target: game.AttackTarget{Player: game.Player3}},
			{Attacker: other.ObjectID, Target: game.AttackTarget{Player: game.Player3}},
		}}
		agent := &scopedMayAgent{accept: []bool{accept}}
		agents := [game.NumPlayers]PlayerAgent{}
		agents[game.Player1] = agent
		obj := &game.StackObject{Controller: game.Player1, SourceID: other.ObjectID, HasTriggerEvent: true,
			TriggerEvent: game.Event{PermanentID: event.ObjectID}}
		NewEngine(nil).resolveAbilityContentWithChoices(g, obj, def.TriggeredAbilities[0].Content, agents, &TurnLog{})
		eventAttacking, otherAttacking := false, false
		for _, declaration := range g.Combat.Attackers {
			eventAttacking = eventAttacking || declaration.Attacker == event.ObjectID
			otherAttacking = otherAttacking || declaration.Attacker == other.ObjectID
		}
		if eventAttacking == accept || !otherAttacking || !other.Tapped || agent.next != 1 {
			t.Fatal("optional combat group lost event identity or required untap effectiveness")
		}
	}
}

func TestOptionalGroupRepeatedSkippedAndSelectedModeIsolation(t *testing.T) {
	body := compiledScopedResultContent(t, "If you have no cards in hand, you may draw a card and gain 2 life.")
	for _, modal := range []bool{false, true} {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		for range 3 {
			addCardToLibrary(g, game.Player1, evidenceCard("Drawn", 1))
		}
		agent := &scopedMayAgent{accept: []bool{true}}
		agents := [game.NumPlayers]PlayerAgent{}
		agents[game.Player1] = agent
		obj := &game.StackObject{Controller: game.Player1}
		content := game.Mode{Sequence: []game.Instruction{{Primitive: game.RepeatProcess{
			Times: game.Fixed(2), Body: body,
		}}}}.Ability()
		if modal {
			content = game.AbilityContent{MinModes: 2, MaxModes: 2, Modes: []game.Mode{body.Modes[0], body.Modes[0]}}
			obj.ChosenModes = []int{0, 1}
		}
		NewEngine(nil).resolveAbilityContentWithChoices(g, obj, content, agents, &TurnLog{})
		if agent.next != 1 || g.Players[game.Player1].Life != 42 || g.Players[game.Player1].Hand.Size() != 1 {
			t.Fatal("skipped later iteration/mode inherited an accepted decision")
		}
	}
}

func TestOptionalGroupSingleActorDecision(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	controller := &scopedMayAgent{accept: []bool{false}}
	decider := &scopedMayAgent{accept: []bool{true}}
	agents := [game.NumPlayers]PlayerAgent{}
	agents[game.Player1], agents[game.Player2] = controller, decider
	sequence := []game.Instruction{
		{Primitive: game.GainLife{Player: game.TargetPlayerReference(0), Amount: game.Fixed(2)},
			Optional: true, OptionalActor: opt.Val(game.TargetPlayerReference(0)), PublishOptionalDecision: "group"},
		{Primitive: game.GainLife{Player: game.TargetPlayerReference(0), Amount: game.Fixed(3)},
			OptionalDecisionGate: "group"},
	}
	NewEngine(nil).resolveInstructionSequence(g, &game.StackObject{
		Controller: game.Player1, Targets: []game.Target{game.PlayerTarget(game.Player2)},
	}, sequence, agents, &TurnLog{})
	if controller.next != 0 || decider.next != 1 || g.Players[game.Player2].Life != 45 || g.Players[game.Player1].Life != 40 {
		t.Fatal("group decision did not use the envelope's single actor")
	}
}
