package rules

import (
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func compiledScopedResultContent(t *testing.T, text string) game.AbilityContent {
	t.Helper()
	defs, diagnostics, err := cardgen.CompileCardDefs(&cardgen.ScryfallCard{
		Name: "Scoped Runtime Probe", Layout: "normal", TypeLine: "Sorcery", OracleText: text,
	})
	if err != nil || len(diagnostics) != 0 || len(defs) != 1 || !defs[0].SpellAbility.Exists {
		t.Fatalf("compile = %v, %v", diagnostics, err)
	}
	return defs[0].SpellAbility.Val
}

func TestScopedIndependentActualResults(t *testing.T) {
	content := compiledScopedResultContent(t,
		"You may discard a card. If you do, you gain 3 life. You gain 2 life. You may sacrifice a creature. If you do, you gain 5 life.")
	published := map[game.ResultKey]int{}
	for _, instruction := range content.Modes[0].Sequence {
		if instruction.PublishResult != "" {
			published[instruction.PublishResult]++
		}
	}
	if len(published) != 2 || published["result-clause-1"] != 1 || published["result-clause-4"] != 1 {
		t.Fatalf("publications=%v, want exactly one publication for each independent producer", published)
	}
	// Each producer owns its own receipt; success of one never answers the other.
	probes := []resultProbe{
		{key: "result-clause-1", life: 100},
		{key: "result-clause-4", life: 1000},
		{key: "result-clause-1", succeeded: game.TriTrue, life: 10_000},
		{key: "result-clause-4", succeeded: game.TriTrue, life: 100_000},
	}
	for _, tt := range []struct {
		name               string
		discard, sacrifice bool
		decline            bool
		want               int
	}{
		{"both succeed", true, true, false, 50},
		{"first fails second succeeds", false, true, false, 47},
		{"first succeeds second fails", true, false, false, 45},
		{"both fail", false, false, false, 42},
		{"both declined", true, true, true, 42},
	} {
		t.Run(tt.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			if tt.discard {
				addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Types: []types.Card{types.Land}}})
			}
			if tt.sacrifice {
				addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Types: []types.Card{types.Creature}}})
			}
			agents := [game.NumPlayers]PlayerAgent{}
			if tt.decline {
				agents[game.Player1] = &choiceOnlyAgent{choices: [][]int{{0}, {0}}}
			}
			visible := 1100
			if tt.discard && !tt.decline {
				visible += 10_000
			}
			if tt.sacrifice && !tt.decline {
				visible += 100_000
			}
			obj := &game.StackObject{Controller: game.Player1}
			NewEngine(nil).resolveAbilityContentWithChoices(g, obj, withResultProbes(content, probes), agents, &TurnLog{})
			if got := g.Players[game.Player1].Life; got != tt.want+visible {
				t.Fatalf("life = %d, want %d plus both independent in-resolution receipts %d", got, tt.want, visible)
			}
			if g.Players[game.Player2].Life != 40 {
				t.Fatal("result consequence escaped the frozen controller")
			}
			assertResultKeysCleaned(t, obj, "result-clause-1", "result-clause-4")
			before := g.Players[game.Player1].Life
			decline := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{0}, {0}}}}
			NewEngine(nil).resolveAbilityContentWithChoices(g, obj, withResultProbes(content, probes), decline, &TurnLog{})
			if got := g.Players[game.Player1].Life - before; got != 2+1100 {
				t.Fatalf("declined repeat observed %d, want its own two unsuccessful receipts and 2 life", got)
			}
		})
	}
}

func TestScopedFilteredResultsSharePublication(t *testing.T) {
	content := compiledScopedResultContent(t,
		"Discard a card. You gain 2 life. If a land card was discarded this way, you gain 3 life. If a creature card was discarded this way, you gain 5 life.")
	// One producer publishes one receipt; both filtered consumers read it.
	published := map[game.ResultKey]bool{}
	consumers := 0
	for _, instruction := range content.Modes[0].Sequence {
		if instruction.PublishResult != "" {
			published[instruction.PublishResult] = true
		}
		if instruction.ResultGate.Exists && instruction.ResultGate.Val.ObjectSelection.Exists {
			if instruction.ResultGate.Val.Key != "if-you-do" {
				t.Fatalf("filtered consumer reads %q, not the shared publication", instruction.ResultGate.Val.Key)
			}
			consumers++
		}
	}
	if len(published) != 1 || !published["if-you-do"] || consumers != 2 {
		t.Fatalf("publications=%v filtered consumers=%d, want one shared publication read twice", published, consumers)
	}
	member := func(typ types.Card, life int) resultProbe {
		return resultProbe{key: "if-you-do", succeeded: game.TriTrue, life: life, selection: opt.Val(game.Selection{RequiredTypes: []types.Card{typ}})}
	}
	probes := []resultProbe{
		{key: "if-you-do", life: 100},
		{key: "if-you-do", succeeded: game.TriTrue, life: 1000},
		member(types.Land, 10_000),
		member(types.Creature, 100_000),
	}
	for _, tt := range []struct {
		name    string
		types   []types.Card
		want    int
		visible int
	}{
		{"land", []types.Card{types.Land}, 45, 11_100},
		{"creature", []types.Card{types.Creature}, 47, 101_100},
		{"both characteristics", []types.Card{types.Land, types.Creature}, 50, 111_100},
		{"neither characteristic", []types.Card{types.Artifact}, 42, 1100},
		{"failed empty hand", nil, 42, 100},
	} {
		t.Run(tt.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			if tt.types != nil {
				addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Types: tt.types}})
			}
			addCardToHand(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Types: []types.Card{types.Land, types.Creature}}})
			obj := &game.StackObject{Controller: game.Player1}
			NewEngine(nil).resolveAbilityContentWithChoices(g, obj, withResultProbes(content, probes), [game.NumPlayers]PlayerAgent{}, &TurnLog{})
			if g.Players[game.Player1].Life != tt.want+tt.visible || g.Players[game.Player2].Hand.Size() != 1 {
				t.Fatalf("life=%d hand=%d, want life %d plus in-resolution receipt/members %d",
					g.Players[game.Player1].Life, g.Players[game.Player2].Hand.Size(), tt.want, tt.visible)
			}
			assertResultKeysCleaned(t, obj, "if-you-do")
			before := g.Players[game.Player1].Life
			NewEngine(nil).resolveAbilityContentWithChoices(g, obj, withResultProbes(content, probes), [game.NumPlayers]PlayerAgent{}, &TurnLog{})
			if got := g.Players[game.Player1].Life - before; got != 2+100 {
				t.Fatalf("empty-hand repeat observed %d, want its own failed receipt and 2 life", got)
			}
		})
	}
}

func TestScopedMissingSkippedAndFailedPublication(t *testing.T) {
	content := compiledScopedResultContent(t,
		"If you control a creature, discard a card. If a land card was discarded this way, you gain 3 life. Otherwise, you gain 1 life. You gain 2 life.")
	probes := []resultProbe{
		{key: "if-you-do", life: 100},
		{key: "if-you-do", succeeded: game.TriTrue, life: 1000},
	}
	for _, tt := range []struct {
		name           string
		creature, hand bool
		want           int
		visible        int
	}{
		{"skipped is unavailable even to complement", false, true, 42, 0},
		{"failed available result", true, false, 43, 100},
		{"succeeded result", true, true, 45, 1100},
	} {
		t.Run(tt.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			var creature *game.Permanent
			if tt.creature {
				creature = addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Types: []types.Card{types.Creature}}})
			}
			if tt.hand {
				addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Types: []types.Card{types.Land}}})
			}
			obj := &game.StackObject{Controller: game.Player1}
			NewEngine(nil).resolveAbilityContentWithChoices(g, obj, withResultProbes(content, probes), [game.NumPlayers]PlayerAgent{}, &TurnLog{})
			if got := g.Players[game.Player1].Life; got != tt.want+tt.visible {
				t.Fatalf("life=%d, want branch life %d plus in-resolution visibility %d", got, tt.want, tt.visible)
			}
			assertResultKeysCleaned(t, obj, "if-you-do")
			if creature != nil {
				if !sacrificePermanent(g, creature) {
					t.Fatal("could not remove the condition creature")
				}
			}
			before := g.Players[game.Player1].Life
			NewEngine(nil).resolveAbilityContentWithChoices(g, obj, withResultProbes(content, probes), [game.NumPlayers]PlayerAgent{}, &TurnLog{})
			if got := g.Players[game.Player1].Life - before; got != 2 {
				t.Fatalf("second skipped invocation gained %d, want only its unconditional 2 life", got)
			}
		})
	}
}

// resultProbe observes one local result while its owning resolution runs.
type resultProbe struct {
	key                 game.ResultKey
	accepted, succeeded game.TriState
	selection           opt.V[game.Selection]
	life                int
}

func withResultProbes(content game.AbilityContent, probes []resultProbe) game.AbilityContent {
	probed := content
	probed.Modes = append([]game.Mode(nil), content.Modes...)
	for i := range probed.Modes {
		sequence := append([]game.Instruction(nil), probed.Modes[i].Sequence...)
		for _, probe := range probes {
			sequence = append(sequence, game.Instruction{
				Primitive: game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(probe.life)},
				ResultGate: opt.Val(game.InstructionResultGate{
					Key: probe.key, Accepted: probe.accepted, Succeeded: probe.succeeded,
					ObjectSelection: probe.selection,
				}),
			})
		}
		probed.Modes[i].Sequence = sequence
	}
	return probed
}

func assertResultKeysCleaned(t *testing.T, obj *game.StackObject, keys ...string) {
	t.Helper()
	for _, key := range keys {
		_, receipt := obj.ResolutionResults[key]
		_, objects := obj.ResolutionResultObjects[key]
		_, amount := obj.ResolvedAmounts[key]
		if receipt || objects || amount {
			t.Fatalf("local result %q leaked past its owning resolution", key)
		}
	}
}

func TestActualMatchingResultCounts(t *testing.T) {
	for _, condition := range []string{
		"two or more land cards were discarded",
		"you discarded two or more land cards",
		"two land cards were discarded",
	} {
		content := compiledScopedResultContent(t,
			"Discard three cards. If "+condition+" this way, you gain 3 life. Otherwise, you gain 1 life.")
		for _, tt := range []struct {
			name string
			hand []types.Card
			want int
		}{
			{"request three actual zero", nil, 41},
			{"request three actual one matching", []types.Card{types.Land}, 41},
			{"two actual matching", []types.Card{types.Land, types.Land}, 43},
			{"three actual only one matching", []types.Card{types.Creature, types.Land, types.Artifact}, 41},
			{"three actual matching", []types.Card{types.Land, types.Land, types.Land}, 43},
		} {
			t.Run(condition+"/"+tt.name, func(t *testing.T) {
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				for _, typ := range tt.hand {
					addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Types: []types.Card{typ}}})
				}
				addCardToHand(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Types: []types.Card{types.Land}}})
				want := tt.want
				if condition == "two land cards were discarded" && len(tt.hand) == 3 && tt.hand[0] == types.Land {
					want = 41
				}
				obj := &game.StackObject{Controller: game.Player1}
				NewEngine(nil).resolveAbilityContentWithChoices(g, obj, content, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
				if got := g.Players[game.Player1].Life; got != want {
					t.Fatalf("life=%d want=%d result=%#v objects=%#v", got, want, obj.ResolutionResults, obj.ResolutionResultObjects)
				}
			})
		}
	}
}

func TestScopedResultTargetSlotsAndExpandedConsumers(t *testing.T) {
	content := compiledScopedResultContent(t,
		"Tap target creature. Destroy target creature. If a creature was destroyed this way, put a +1/+1 counter on each of up to two target creatures. You gain 2 life.")
	for _, indestructible := range []bool{false, true} {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		first := addCombatCreaturePermanent(g, game.Player1, game.KeywordNone)
		keyword := game.KeywordNone
		if indestructible {
			keyword = game.Indestructible
		}
		producerTarget := addCombatCreaturePermanent(g, game.Player2, keyword)
		one := addCombatCreaturePermanent(g, game.Player1, game.KeywordNone)
		two := addCombatCreaturePermanent(g, game.Player2, game.KeywordNone)
		obj := &game.StackObject{Controller: game.Player1, Targets: []game.Target{
			game.PermanentTarget(first.ObjectID), game.PermanentTarget(producerTarget.ObjectID),
			game.PermanentTarget(one.ObjectID), game.PermanentTarget(two.ObjectID),
		}}
		NewEngine(nil).resolveAbilityContentWithChoices(g, obj, content, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
		want := 1
		if indestructible {
			want = 0
		}
		if one.Counters.Get(counter.PlusOnePlusOne) != want || two.Counters.Get(counter.PlusOnePlusOne) != want || first.Counters.Get(counter.PlusOnePlusOne) != 0 ||
			!first.Tapped || g.Players[game.Player1].Life != 42 || g.Players[game.Player2].Life != 40 {
			t.Fatalf("nonzero producer/consumer slots lost: first=%#v one=%#v two=%#v", first, one, two)
		}
	}
}

func TestScopedBareOptionalIndependentRider(t *testing.T) {
	content := compiledScopedResultContent(t, "You may discard a card. You gain 2 life.")
	for _, decline := range []bool{false, true} {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Types: []types.Card{types.Land}}})
		agents := [game.NumPlayers]PlayerAgent{}
		if decline {
			agents[game.Player1] = &choiceOnlyAgent{choices: [][]int{{0}}}
		}
		NewEngine(nil).resolveAbilityContentWithChoices(g, &game.StackObject{Controller: game.Player1}, content, agents, &TurnLog{})
		wantHand := 0
		if decline {
			wantHand = 1
		}
		if g.Players[game.Player1].Life != 42 || g.Players[game.Player1].Hand.Size() != wantHand {
			t.Fatal("independent rider inherited the optional action")
		}
	}
}

func TestCountGateMissingPublicationFailsClosed(t *testing.T) {
	for _, negate := range []bool{false, true} {
		gate := game.InstructionResultGate{
			Key: "missing", Succeeded: game.TriTrue, ObjectSelection: opt.Val(game.Selection{}),
			ObjectCountRange: opt.Val(game.IntRange{Min: 2, Max: 3}), Negate: negate,
		}
		if instructionResultGateSatisfied(game.NewGame([game.NumPlayers]game.PlayerConfig{}), &game.StackObject{}, gate) {
			t.Fatal("count/complement admitted unavailable publication")
		}
	}
}

func TestScopedSpecializedSacrificeCountsActualResults(t *testing.T) {
	for _, condition := range []string{"a creature was", "two creatures were"} {
		content := compiledScopedResultContent(t, "You may sacrifice a creature. If "+condition+
			" sacrificed this way, return that card to the battlefield under its owner's control with three +1/+1 counters on it.")
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		addCombatCreaturePermanent(g, game.Player1, game.KeywordNone)
		obj := &game.StackObject{Controller: game.Player1}
		NewEngine(nil).resolveAbilityContentWithChoices(g, obj, content, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
		gate := content.Modes[0].Sequence[1].ResultGate.Val
		if got := instructionResultGateSatisfied(g, obj, gate); got != (condition == "a creature was") {
			t.Fatalf("condition %q matched one actual sacrifice: %t", condition, got)
		}
	}
}

func TestScopedRepeatedOptionalUsesOneChoicePerIteration(t *testing.T) {
	content := compiledScopedResultContent(t, "Repeat the following process two times. You may draw a card.")
	for _, accept := range [][]bool{{false, false}, {true, false}, {false, true}, {true, true}} {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		for range 2 {
			addCardToLibrary(g, game.Player1, &game.CardDef{})
		}
		agent := &scopedMayAgent{accept: accept}
		agents := [game.NumPlayers]PlayerAgent{game.Player1: agent}
		NewEngine(nil).resolveAbilityContentWithChoices(g, &game.StackObject{Controller: game.Player1}, content, agents, &TurnLog{})
		want := 0
		for _, accepted := range accept {
			if accepted {
				want++
			}
		}
		if agent.next != 2 || g.Players[game.Player1].Hand.Size() != want {
			t.Fatalf("accept=%v choices=%d draws=%d, want 2/%d", accept, agent.next, g.Players[game.Player1].Hand.Size(), want)
		}
	}
}

func TestScopedModalUnconditionalPublicationsOverwrite(t *testing.T) {
	content := compiledScopedResultContent(t, "Choose two \u2014\n"+
		"\u2022 Discard a card. If a land card was discarded this way, you gain 1 life.\n"+
		"\u2022 Discard a card. If a land card was discarded this way, you gain 5 life. Otherwise, you gain 2 life.")
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Types: []types.Card{types.Land}}})
	obj := &game.StackObject{Controller: game.Player1, ChosenModes: []int{0, 1}}
	NewEngine(nil).resolveAbilityContentWithChoices(g, obj, content, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if g.Players[game.Player1].Life != 43 {
		t.Fatalf("second mode used first mode's publication instead of its failed action: life=%d", g.Players[game.Player1].Life)
	}
}

func TestScopedActiveActorIsNotAffectedObjectController(t *testing.T) {
	content := compiledScopedResultContent(t,
		"Destroy target creature. If you destroyed a creature this way, you gain 2 life.")
	for _, owner := range []game.PlayerID{game.Player1, game.Player2} {
		for _, indestructible := range []bool{false, true} {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			keyword := game.KeywordNone
			if indestructible {
				keyword = game.Indestructible
			}
			creature := addCombatCreaturePermanent(g, owner, keyword)
			obj := &game.StackObject{Controller: game.Player1, Targets: []game.Target{game.PermanentTarget(creature.ObjectID)}}
			NewEngine(nil).resolveAbilityContentWithChoices(g, obj, content, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
			want := 42
			if indestructible {
				want = 40
			}
			if g.Players[game.Player1].Life != want || g.Players[game.Player2].Life != 40 {
				t.Fatalf("actor/affected-controller confused: owner=%d life=%d", owner, g.Players[game.Player1].Life)
			}
		}
	}
}

type scopedMayAgent struct {
	choiceOnlyAgent

	accept []bool
	next   int
}

func (a *scopedMayAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	if request.Kind != game.ChoiceMay {
		return request.DefaultSelection
	}
	accept := a.next < len(a.accept) && a.accept[a.next]
	a.next++
	if accept {
		return []int{1}
	}
	return []int{0}
}

func TestScopedNestedOptionalFailure(t *testing.T) {
	content := compiledScopedResultContent(t,
		"You may discard a card. If you do, you may sacrifice a creature. If you don't, you gain 2 life.")
	probes := []resultProbe{
		{key: "result-clause-1", life: 100},
		{key: "result-clause-1", succeeded: game.TriTrue, life: 1000},
		{key: "result-clause-1", accepted: game.TriTrue, life: 10_000_000},
		{key: "result-clause-2", life: 10_000},
		{key: "result-clause-2", succeeded: game.TriTrue, life: 100_000},
		{key: "result-clause-2", accepted: game.TriTrue, life: 1_000_000},
	}
	for _, tt := range []struct {
		name              string
		discard, creature bool
		accept            []bool
		wantLife, visible int
	}{
		{"both accepted", true, true, []bool{true, true}, 40, 11_111_100},
		{"inner declined", true, true, []bool{true, false}, 42, 10_011_100},
		{"inner failed", true, false, []bool{true, true}, 42, 11_011_100},
		{"outer declined inner unavailable", true, true, []bool{false}, 40, 100},
		{"outer failed inner unavailable", false, true, []bool{true}, 40, 10_000_100},
	} {
		t.Run(tt.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			if tt.discard {
				addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Types: []types.Card{types.Land}}})
			}
			if tt.creature {
				addCombatCreaturePermanent(g, game.Player1, game.KeywordNone)
			}
			agent := &scopedMayAgent{accept: tt.accept}
			agents := [game.NumPlayers]PlayerAgent{game.Player1: agent}
			obj := &game.StackObject{Controller: game.Player1}
			NewEngine(nil).resolveAbilityContentWithChoices(g, obj, withResultProbes(content, probes), agents, &TurnLog{})
			if got := g.Players[game.Player1].Life; got != tt.wantLife+tt.visible {
				t.Fatalf("life=%d, want branch life %d plus in-resolution visibility %d", got, tt.wantLife, tt.visible)
			}
			assertResultKeysCleaned(t, obj, "result-clause-1", "result-clause-2")
			before := g.Players[game.Player1].Life
			decline := &scopedMayAgent{accept: []bool{false}}
			NewEngine(nil).resolveAbilityContentWithChoices(g, obj, withResultProbes(content, probes),
				[game.NumPlayers]PlayerAgent{game.Player1: decline}, &TurnLog{})
			if got := g.Players[game.Player1].Life - before; got != 100 {
				t.Fatalf("declined second invocation observed %d, want only its own declined outer receipt", got)
			}
		})
	}
}

func TestScopedOtherwiseSubjectContinuation(t *testing.T) {
	content := compiledScopedResultContent(t,
		"Discard a card. If a land card was discarded this way, you gain 2 life. Otherwise, put a +1/+1 counter on target creature. It gains trample until end of turn. You gain 1 life.")
	for _, land := range []bool{false, true} {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		typ := types.Artifact
		if land {
			typ = types.Land
		}
		addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Types: []types.Card{typ}}})
		creature := addCombatCreaturePermanent(g, game.Player2, game.KeywordNone)
		obj := &game.StackObject{Controller: game.Player1, Targets: []game.Target{game.PermanentTarget(creature.ObjectID)}}
		// The continuation's own Otherwise gate is evaluated inside the owning
		// resolution by a probe sharing it, and through its actual trample grant.
		continuation := content.Modes[0].Sequence[3].ResultGate.Val
		probed := withResultProbes(content, nil)
		probed.Modes[0].Sequence = append(probed.Modes[0].Sequence, game.Instruction{
			Primitive:  game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(100)},
			ResultGate: opt.Val(continuation),
		})
		NewEngine(nil).resolveAbilityContentWithChoices(g, obj, probed, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
		wantCount, wantLife := 1, 41+100
		if land {
			wantCount, wantLife = 0, 43
		}
		if creature.Counters.Get(counter.PlusOnePlusOne) != wantCount ||
			hasKeyword(g, creature, game.Trample) == land ||
			g.Players[game.Player1].Life != wantLife {
			t.Fatalf("otherwise leaked: counters=%d life=%d", creature.Counters.Get(counter.PlusOnePlusOne), g.Players[game.Player1].Life)
		}
		assertResultKeysCleaned(t, obj, string(continuation.Key))
	}
}
