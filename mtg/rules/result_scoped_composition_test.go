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
			obj := &game.StackObject{Controller: game.Player1}
			NewEngine(nil).resolveAbilityContentWithChoices(g, obj, content, agents, &TurnLog{})
			if got := g.Players[game.Player1].Life; got != tt.want {
				t.Fatalf("life = %d, want %d", got, tt.want)
			}
			if len(obj.ResolutionResults) != 2 {
				t.Fatalf("independent publications = %#v", obj.ResolutionResults)
			}
			if g.Players[game.Player2].Life != 40 {
				t.Fatal("result consequence escaped the frozen controller")
			}
		})
	}
}

func TestScopedFilteredResultsSharePublication(t *testing.T) {
	content := compiledScopedResultContent(t,
		"Discard a card. You gain 2 life. If a land card was discarded this way, you gain 3 life. If a creature card was discarded this way, you gain 5 life.")
	for _, tt := range []struct {
		name  string
		types []types.Card
		want  int
	}{
		{"land", []types.Card{types.Land}, 45},
		{"creature", []types.Card{types.Creature}, 47},
		{"both characteristics", []types.Card{types.Land, types.Creature}, 50},
		{"neither characteristic", []types.Card{types.Artifact}, 42},
		{"failed empty hand", nil, 42},
	} {
		t.Run(tt.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			if tt.types != nil {
				addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Types: tt.types}})
			}
			addCardToHand(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Types: []types.Card{types.Land, types.Creature}}})
			obj := &game.StackObject{Controller: game.Player1}
			NewEngine(nil).resolveAbilityContentWithChoices(g, obj, content, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
			if g.Players[game.Player1].Life != tt.want || g.Players[game.Player2].Hand.Size() != 1 ||
				len(obj.ResolutionResults) != 1 {
				t.Fatalf("life=%d hand=%d results=%#v", g.Players[game.Player1].Life, g.Players[game.Player2].Hand.Size(), obj.ResolutionResults)
			}
		})
	}
}

func TestScopedMissingSkippedAndFailedPublication(t *testing.T) {
	content := compiledScopedResultContent(t,
		"If you control a creature, discard a card. If a land card was discarded this way, you gain 3 life. Otherwise, you gain 1 life. You gain 2 life.")
	for _, tt := range []struct {
		name           string
		creature, hand bool
		want           int
		published      bool
	}{
		{"skipped is unavailable even to complement", false, true, 42, false},
		{"failed available result", true, false, 43, true},
		{"succeeded result", true, true, 45, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			if tt.creature {
				addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Types: []types.Card{types.Creature}}})
			}
			if tt.hand {
				addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Types: []types.Card{types.Land}}})
			}
			obj := &game.StackObject{Controller: game.Player1}
			NewEngine(nil).resolveAbilityContentWithChoices(g, obj, content, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
			_, published := obj.ResolutionResults["if-you-do"]
			if g.Players[game.Player1].Life != tt.want || published != tt.published {
				t.Fatalf("life=%d published=%t, want %d/%t", g.Players[game.Player1].Life, published, tt.want, tt.published)
			}
		})
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
	for _, tt := range []struct {
		name              string
		discard, creature bool
		accept            []bool
		wantLife, results int
	}{
		{"both accepted", true, true, []bool{true, true}, 40, 2},
		{"inner declined", true, true, []bool{true, false}, 42, 2},
		{"inner failed", true, false, []bool{true, true}, 42, 2},
		{"outer declined inner unavailable", true, true, []bool{false}, 40, 1},
		{"outer failed inner unavailable", false, true, []bool{true}, 40, 1},
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
			NewEngine(nil).resolveAbilityContentWithChoices(g, obj, content, agents, &TurnLog{})
			if g.Players[game.Player1].Life != tt.wantLife || len(obj.ResolutionResults) != tt.results {
				t.Fatalf("life=%d results=%#v", g.Players[game.Player1].Life, obj.ResolutionResults)
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
		NewEngine(nil).resolveAbilityContentWithChoices(g, obj, content, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
		wantCount, wantLife := 1, 41
		if land {
			wantCount, wantLife = 0, 43
		}
		if creature.Counters.Get(counter.PlusOnePlusOne) != wantCount ||
			instructionResultGateSatisfied(g, obj, content.Modes[0].Sequence[3].ResultGate.Val) == land ||
			g.Players[game.Player1].Life != wantLife {
			t.Fatalf("otherwise leaked: counters=%d life=%d", creature.Counters.Get(counter.PlusOnePlusOne), g.Players[game.Player1].Life)
		}
	}
}
