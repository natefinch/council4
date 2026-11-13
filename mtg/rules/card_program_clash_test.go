package rules

import (
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review: CR effective 2026-11-06 from the verified official copy
// (MagicCompRules 20261106), sections 701.30a-d read: clashing reveals the top
// card of the library and may put it on the bottom (701.30a); "clash with an
// opponent" chooses an opponent and both clash (701.30b); reveals are
// simultaneous, placement decisions follow APNAP order, then the cards move at
// the same time (701.30c); the winner revealed a card with a higher mana value
// than every other revealed card (701.30d). CR 101.4 (APNAP choices) and 709.4b
// (split-card mana value off the stack) also apply. Pinned Oracle corpus
// 6da9ced0...: Research the Deep 02f35851-c1be-4162-a6f8-7897325555a7 ("Draw a
// card. Clash with an opponent. If you win, return Research the Deep to its
// owner's hand."; Scryfall rulings 2008-04-01: a win moves it from the stack to
// hand, never the graveyard; a countered spell does not clash), Pulling Teeth
// 9c966ba3-3408-4bdd-a4cf-97c9183e1115 ("Clash with an opponent. If you win,
// target player discards two cards. Otherwise, that player discards a card.";
// ruling 2008-04-01: the clashed opponent need not be the target), and Oaken
// Brawler 819351c6-b0d6-42dc-9bd9-d1a3851aad7c ("When this creature enters,
// clash with an opponent. If you win, put a +1/+1 counter on this creature.";
// no rulings). These test-local programs prove the shared Clash operation; the
// cards themselves are admitted by a later card unit.

var (
	researchTheDeepTestReference = game.ProgramReference{ID: "test/research-the-deep", Version: "1"}
	pullingTeethTestReference    = game.ProgramReference{ID: "test/pulling-teeth", Version: "1"}
	oakenBrawlerTestReference    = game.ProgramReference{ID: "test/oaken-brawler", Version: "1"}
	fixedClashTestReference      = game.ProgramReference{ID: "test/fixed-clash", Version: "1"}
)

type clashFixture struct {
	t       *testing.T
	g       *game.Game
	engine  *Engine
	results *[]cardprogram.ClashResult
}

func newClashFixture(t *testing.T) *clashFixture {
	t.Helper()
	results := &[]cardprogram.ClashResult{}
	catalog, err := cardprogram.NewCatalog(clashTestPrograms(results))
	if err != nil {
		t.Fatal(err)
	}
	return &clashFixture{
		t: t, g: game.NewGame([game.NumPlayers]game.PlayerConfig{}),
		engine: NewEngineWithPrograms(nil, catalog), results: results,
	}
}

// clashTestPrograms are the clashing test programs, recording each result.
func clashTestPrograms(results *[]cardprogram.ClashResult) []cardprogram.Program {
	record := func(result cardprogram.ClashResult) { *results = append(*results, result) }
	return []cardprogram.Program{
		{
			Reference: researchTheDeepTestReference,
			Resolve: func(ctx cardprogram.Context) error {
				if _, err := ctx.Execute(game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}); err != nil {
					return err
				}
				result, err := ctx.Clash(opt.V[game.PlayerReference]{})
				if err != nil {
					return err
				}
				record(result)
				if !result.Won(ctx.View().Source().Controller) {
					return nil
				}
				_, err = ctx.Execute(game.MoveResolvingSpell{Destination: zone.Hand})
				return err
			},
		},
		{
			Reference: pullingTeethTestReference,
			Resolve: func(ctx cardprogram.Context) error {
				result, err := ctx.Clash(opt.V[game.PlayerReference]{})
				if err != nil {
					return err
				}
				record(result)
				amount := 1
				if result.Won(ctx.View().Source().Controller) {
					amount = 2
				}
				_, err = ctx.Execute(game.Discard{Amount: game.Fixed(amount), Player: game.TargetPlayerReference(0)})
				return err
			},
		},
		{
			Reference: oakenBrawlerTestReference,
			Resolve: func(ctx cardprogram.Context) error {
				result, err := ctx.Clash(opt.V[game.PlayerReference]{})
				if err != nil {
					return err
				}
				record(result)
				if !result.Won(ctx.View().Source().Controller) {
					return nil
				}
				_, err = ctx.Execute(game.AddCounter{
					Object: game.SourcePermanentReference(), Amount: game.Fixed(1), CounterKind: counter.PlusOnePlusOne,
				})
				return err
			},
		},
		{
			Reference: fixedClashTestReference,
			Resolve: func(ctx cardprogram.Context) error {
				result, err := ctx.Clash(opt.Val(game.TargetPlayerReference(0)))
				record(result)
				return err
			},
		},
	}
}

func (f *clashFixture) result() cardprogram.ClashResult {
	f.t.Helper()
	if len(*f.results) != 1 {
		f.t.Fatalf("clash ran %d times, want 1", len(*f.results))
	}
	return (*f.results)[0]
}

func researchTheDeepTestDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:         "Research the Deep",
		ManaCost:     opt.Val(cost.Mana{cost.O(1), cost.U}),
		Types:        []types.Card{types.Sorcery},
		SpellAbility: opt.Val(game.Mode{Program: researchTheDeepTestReference}.Ability()),
	}}
}

func pullingTeethTestDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     "Pulling Teeth",
		ManaCost: opt.Val(cost.Mana{cost.O(1), cost.B}),
		Types:    []types.Card{types.Sorcery},
		SpellAbility: opt.Val(game.Mode{
			Targets: []game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Constraint: "target player", Allow: game.TargetAllowPlayer}},
			Program: pullingTeethTestReference,
		}.Ability()),
	}}
}

func oakenBrawlerTestDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:      "Oaken Brawler",
		ManaCost:  opt.Val(cost.Mana{cost.O(3), cost.W}),
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Treefolk, types.Warrior},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 4}),
		TriggeredAbilities: []game.TriggeredAbility{{
			Text: "When this creature enters, clash with an opponent. If you win, put a +1/+1 counter on this creature.",
			Trigger: game.TriggerCondition{Type: game.TriggerWhen, Pattern: game.TriggerPattern{
				Event: game.EventPermanentEnteredBattlefield, Source: game.TriggerSourceSelf,
			}},
			Content: game.Mode{Program: oakenBrawlerTestReference}.Ability(),
		}},
	}}
}

// clashLibraryCard is a library card with the given mana value; zero is a land.
func clashLibraryCard(name string, manaValue int) *game.CardDef {
	if manaValue == 0 {
		return &game.CardDef{CardFace: game.CardFace{Name: name, Types: []types.Card{types.Land}}}
	}
	return &game.CardDef{CardFace: game.CardFace{
		Name: name, Types: []types.Card{types.Creature}, ManaCost: opt.Val(cost.Mana{cost.O(manaValue)}),
		Power: opt.Val(game.PT{Value: 1}), Toughness: opt.Val(game.PT{Value: 1}),
	}}
}

// clashAgents answers the controller's opponent choice with opponent and each
// placement prompt with bottom[player], recording every request in order.
func clashAgents(t *testing.T, opponent game.PlayerID, bottom map[game.PlayerID]bool, requests *[]game.ChoiceRequest) [game.NumPlayers]PlayerAgent {
	agents := simPassPolicies()
	for player := range game.PlayerID(game.NumPlayers) {
		agents[player] = authoredChoicePolicy{
			simPassPolicy: agents[player].(simPassPolicy),
			choose: func(_ PlayerObservation, request game.ChoiceRequest) []int {
				if requests != nil {
					*requests = append(*requests, request)
				}
				switch request.Kind {
				case game.ChoicePlayer:
					return chooseAuthoredTargets(t, request, game.PlayerTarget(opponent))
				case game.ChoiceResolution:
					if strings.HasPrefix(request.Prompt, "Clash:") && bottom[request.Player] {
						return []int{1}
					}
					return []int{0}
				}
				return slices.Clone(request.DefaultSelection)
			},
		}
	}
	return agents
}

func (f *clashFixture) cast(def *game.CardDef, color mana.Color, generic int, targets []game.Target) id.ID {
	f.t.Helper()
	cardID := addCardToHand(f.g, game.Player1, def)
	f.g.Players[game.Player1].ManaPool.Add(color, 1)
	f.g.Players[game.Player1].ManaPool.Add(mana.C, generic)
	setMainPhasePriority(f.g, game.Player1)
	if !f.engine.applyAction(f.g, game.Player1, action.CastSpell(cardID, targets, 0, nil)) {
		f.t.Fatalf("%s cast failed", def.Name)
	}
	return cardID
}

func clashRevealedCards(g *game.Game) map[game.PlayerID][]id.ID {
	revealed := map[game.PlayerID][]id.ID{}
	for _, event := range g.Events {
		if event.Kind == game.EventCardRevealed {
			revealed[event.Player] = append(revealed[event.Player], event.CardID)
		}
	}
	return revealed
}

func TestResearchTheDeepReturnsToHandOnlyWhenItWinsTheClash(t *testing.T) {
	for _, tc := range []struct {
		name         string
		mine, theirs int
		split        bool
		wantHand     bool
	}{
		{name: "higher mana value wins", mine: 3, theirs: 2, wantHand: true},
		{name: "tie does not win", mine: 2, theirs: 2},
		{name: "lower mana value loses", mine: 1, theirs: 2},
		{name: "split card uses combined library mana value", split: true, theirs: 4, wantHand: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newClashFixture(t)
			mine := clashLibraryCard("Clash Card", tc.mine)
			if tc.split {
				mine = &game.CardDef{
					CardFace:  game.CardFace{Name: "Assault", Types: []types.Card{types.Sorcery}, ManaCost: opt.Val(cost.Mana{cost.R})},
					Layout:    game.LayoutSplit,
					Alternate: opt.Val(game.CardFace{Name: "Battery", Types: []types.Card{types.Sorcery}, ManaCost: opt.Val(cost.Mana{cost.O(3), cost.G})}),
				}
			}
			myClash := addCardToLibrary(f.g, game.Player1, mine)
			drawn := addCardToLibrary(f.g, game.Player1, clashLibraryCard("Drawn Card", 5))
			theirClash := addCardToLibrary(f.g, game.Player2, clashLibraryCard("Opponent Card", tc.theirs))
			spell := f.cast(researchTheDeepTestDef(), mana.U, 1, nil)

			f.engine.resolveTopOfStackWithChoices(f.g, clashAgents(t, game.Player2, nil, nil), &TurnLog{})

			result := f.result()
			if !f.g.Players[game.Player1].Hand.Contains(drawn) {
				t.Fatal("Research the Deep did not draw before clashing")
			}
			revealed := clashRevealedCards(f.g)
			if !slices.Equal(revealed[game.Player1], []id.ID{myClash}) || !slices.Equal(revealed[game.Player2], []id.ID{theirClash}) {
				t.Fatalf("revealed = %v, want the post-draw top cards", revealed)
			}
			if got := result.Won(game.Player1); got != tc.wantHand {
				t.Fatalf("won = %v (result %+v), want %v", got, result, tc.wantHand)
			}
			inHand := f.g.Players[game.Player1].Hand.Contains(spell)
			inGraveyard := f.g.Players[game.Player1].Graveyard.Contains(spell)
			if inHand != tc.wantHand || inGraveyard == tc.wantHand {
				t.Fatalf("spell in hand=%v graveyard=%v, want hand=%v", inHand, inGraveyard, tc.wantHand)
			}
			if top, _ := f.g.Players[game.Player1].Library.Top(); top != myClash {
				t.Fatal("a card kept on top moved")
			}
		})
	}
}

func TestClashControllerChoosesOneOfSeveralOpponents(t *testing.T) {
	f := newClashFixture(t)
	addCardToLibrary(f.g, game.Player1, clashLibraryCard("Mine", 2))
	addCardToLibrary(f.g, game.Player1, clashLibraryCard("Drawn", 1))
	before := map[game.PlayerID][]id.ID{}
	for _, player := range []game.PlayerID{game.Player2, game.Player3, game.Player4} {
		addCardToLibrary(f.g, player, clashLibraryCard("Theirs", 1))
		addCardToLibrary(f.g, player, clashLibraryCard("Theirs Top", 4))
		before[player] = f.g.Players[player].Library.All()
	}
	f.cast(researchTheDeepTestDef(), mana.U, 1, nil)
	var requests []game.ChoiceRequest
	bottom := map[game.PlayerID]bool{game.Player3: true}

	f.engine.resolveTopOfStackWithChoices(f.g, clashAgents(t, game.Player3, bottom, &requests), &TurnLog{})

	result := f.result()
	if !result.Available || result.Opponent != game.Player3 || len(result.Reveals) != 2 {
		t.Fatalf("result = %+v, want a clash with Player3", result)
	}
	if requests[0].Kind != game.ChoicePlayer || requests[0].Player != game.Player1 || len(requests[0].Options) != 3 {
		t.Fatalf("first request = %+v, want the controller choosing among three opponents", requests[0])
	}
	for i, opponent := range []game.PlayerID{game.Player2, game.Player3, game.Player4} {
		if !slices.Equal(requests[0].Options[i].Targets, []game.Target{game.PlayerTarget(opponent)}) {
			t.Fatalf("option %d = %+v, want Player%d", i, requests[0].Options[i], opponent+1)
		}
	}
	if !result.Won(game.Player3) {
		t.Fatalf("result = %+v, want Player3's mana value 4 to beat 2", result)
	}
	revealed := clashRevealedCards(f.g)
	if len(revealed[game.Player2]) != 0 || len(revealed[game.Player4]) != 0 || len(revealed[game.Player3]) != 1 {
		t.Fatalf("revealed = %v, want only the controller and the chosen opponent", revealed)
	}
	for _, player := range []game.PlayerID{game.Player2, game.Player4} {
		if !slices.Equal(f.g.Players[player].Library.All(), before[player]) {
			t.Fatalf("Player%d's library changed", player+1)
		}
	}
	want := append(slices.Clone(before[game.Player3][1:]), before[game.Player3][0])
	if !slices.Equal(f.g.Players[game.Player3].Library.All(), want) {
		t.Fatal("the chosen opponent's card did not go to the bottom")
	}
}

func TestClashPlacementsFollowAPNAPAndMoveTogether(t *testing.T) {
	f := newClashFixture(t)
	f.g.Turn.ActivePlayer = game.Player3
	addCardToLibrary(f.g, game.Player1, clashLibraryCard("Mine Under", 6))
	mine := addCardToLibrary(f.g, game.Player1, clashLibraryCard("Mine", 3))
	addCardToLibrary(f.g, game.Player3, clashLibraryCard("Theirs Under", 1))
	theirs := addCardToLibrary(f.g, game.Player3, clashLibraryCard("Theirs", 2))
	myBefore := f.g.Players[game.Player1].Library.All()
	theirBefore := f.g.Players[game.Player3].Library.All()
	addCardToHand(f.g, game.Player4, clashLibraryCard("Hand", 1))
	addImplementationSpellToStack(f.g, game.Player1, pullingTeethTestDef(), []game.Target{game.PlayerTarget(game.Player4)})
	var requests []game.ChoiceRequest
	bottom := map[game.PlayerID]bool{game.Player1: true}

	f.engine.resolveTopOfStackWithChoices(f.g, clashAgents(t, game.Player3, bottom, &requests), &TurnLog{})

	var placements []game.ChoiceRequest
	for _, request := range requests {
		if request.Kind == game.ChoiceResolution && strings.HasPrefix(request.Prompt, "Clash:") {
			placements = append(placements, request)
		}
	}
	if len(placements) != 2 || placements[0].Player != game.Player3 || placements[1].Player != game.Player1 {
		t.Fatalf("placement prompts = %+v, want the active opponent first (APNAP)", placements)
	}
	if !placements[0].Subject.Exists || placements[0].Subject.Val.CardID != theirs ||
		!placements[1].Subject.Exists || placements[1].Subject.Val.CardID != mine {
		t.Fatalf("placements = %+v, want each player's own revealed card as the subject", placements)
	}
	theirInfo, myInfo := placements[0].Subject.Val, placements[1].Subject.Val
	if want := []game.ClashDecision{{Player: game.Player3, Card: theirInfo}, {Player: game.Player1, Card: myInfo}}; !reflect.DeepEqual(placements[0].ClashDecisions, want) {
		t.Fatalf("first placement saw %+v, want both revealed cards and no decisions", placements[0].ClashDecisions)
	}
	if want := []game.ClashDecision{{Player: game.Player3, Card: theirInfo, Decided: true}, {Player: game.Player1, Card: myInfo}}; !reflect.DeepEqual(placements[1].ClashDecisions, want) {
		t.Fatalf("second placement saw %+v, want both reveals and the opponent's public top choice", placements[1].ClashDecisions)
	}
	if revealed := clashRevealedCards(f.g); !slices.Equal(revealed[game.Player3], []id.ID{theirs}) ||
		!slices.Equal(revealed[game.Player1], []id.ID{mine}) {
		t.Fatalf("revealed = %v, want both top cards publicly revealed", revealed)
	}
	if want := append(slices.Clone(myBefore[1:]), myBefore[0]); !slices.Equal(f.g.Players[game.Player1].Library.All(), want) {
		t.Fatal("the controller's bottom choice did not move its card to the bottom")
	}
	if !slices.Equal(f.g.Players[game.Player3].Library.All(), theirBefore) {
		t.Fatal("the opponent's top choice moved its card")
	}
	result := f.result()
	if !result.Won(game.Player1) || result.Reveals[0].Player != game.Player3 || !result.Reveals[1].Bottom {
		t.Fatalf("result = %+v, want APNAP reveals and the controller's win", result)
	}
}

func TestClashFirstChooserSeesEveryRevealedCard(t *testing.T) {
	f := newClashFixture(t)
	mine := addCardToLibrary(f.g, game.Player1, clashLibraryCard("Mine", 2))
	theirs := addCardToLibrary(f.g, game.Player2, clashLibraryCard("Theirs", 5))
	addCardToHand(f.g, game.Player4, clashLibraryCard("Hand", 1))
	addImplementationSpellToStack(f.g, game.Player1, pullingTeethTestDef(), []game.Target{game.PlayerTarget(game.Player4)})
	var requests []game.ChoiceRequest
	agents := clashAgents(t, game.Player2, nil, &requests)
	// The active controller chooses first and bottoms its card only after
	// seeing that the opponent's revealed card beats it.
	agents[game.Player1] = authoredChoicePolicy{
		simPassPolicy: simPassPolicy{},
		choose: func(_ PlayerObservation, request game.ChoiceRequest) []int {
			requests = append(requests, request)
			if request.Kind == game.ChoicePlayer {
				return chooseAuthoredTargets(t, request, game.PlayerTarget(game.Player2))
			}
			for _, decision := range request.ClashDecisions {
				if decision.Player != request.Player && decision.Card.ManaValue > request.Subject.Val.ManaValue {
					return []int{1}
				}
			}
			return []int{0}
		},
	}

	f.engine.resolveTopOfStackWithChoices(f.g, agents, &TurnLog{})

	var placements []game.ChoiceRequest
	for _, request := range requests {
		if request.Kind == game.ChoiceResolution && strings.HasPrefix(request.Prompt, "Clash:") {
			placements = append(placements, request)
		}
	}
	if len(placements) != 2 || placements[0].Player != game.Player1 || placements[1].Player != game.Player2 {
		t.Fatalf("placement prompts = %+v, want the active controller first (APNAP)", placements)
	}
	first := placements[0].ClashDecisions
	if len(first) != 2 || first[0].Player != game.Player1 || first[0].Card.CardID != mine ||
		first[1].Player != game.Player2 || first[1].Card.CardID != theirs || first[1].Card.ManaValue != 5 ||
		first[0].Decided || first[1].Decided {
		t.Fatalf("first chooser saw %+v, want both revealed cards before any decision", first)
	}
	second := placements[1].ClashDecisions
	if len(second) != 2 || !second[0].Decided || !second[0].Bottom || second[1].Decided {
		t.Fatalf("second chooser saw %+v, want the controller's public bottom choice", second)
	}
	if bottom, _ := f.g.Players[game.Player1].Library.Bottom(); bottom != mine {
		t.Fatal("the controller's informed bottom choice did not move its card")
	}
	if result := f.result(); !result.Won(game.Player2) {
		t.Fatalf("result = %+v, want the opponent's higher card to win", result)
	}
}

func TestPullingTeethClashHandlesEmptyLibrariesAndTies(t *testing.T) {
	for _, tc := range []struct {
		name            string
		mine, theirs    opt.V[int]
		wantWinner      opt.V[game.PlayerID]
		wantDiscarded   int
		wantRevealCount int
	}{
		{name: "opponent library empty and land wins", mine: opt.Val(0), wantWinner: opt.Val(game.Player1), wantDiscarded: 2, wantRevealCount: 1},
		{name: "controller library empty loses", theirs: opt.Val(0), wantWinner: opt.Val(game.Player2), wantDiscarded: 1, wantRevealCount: 1},
		{name: "both libraries empty", wantDiscarded: 1},
		{name: "tie", mine: opt.Val(3), theirs: opt.Val(3), wantDiscarded: 1, wantRevealCount: 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newClashFixture(t)
			if tc.mine.Exists {
				addCardToLibrary(f.g, game.Player1, clashLibraryCard("Mine", tc.mine.Val))
			}
			if tc.theirs.Exists {
				addCardToLibrary(f.g, game.Player2, clashLibraryCard("Theirs", tc.theirs.Val))
			}
			for range 3 {
				addCardToHand(f.g, game.Player4, clashLibraryCard("Hand", 1))
			}
			f.cast(pullingTeethTestDef(), mana.B, 1, []game.Target{game.PlayerTarget(game.Player4)})

			f.engine.resolveTopOfStackWithChoices(f.g, clashAgents(t, game.Player2, nil, nil), &TurnLog{})

			result := f.result()
			if !result.Available || result.Winner != tc.wantWinner {
				t.Fatalf("result = %+v, want winner %+v", result, tc.wantWinner)
			}
			if discarded := 3 - f.g.Players[game.Player4].Hand.Size(); discarded != tc.wantDiscarded {
				t.Fatalf("target discarded %d, want %d", discarded, tc.wantDiscarded)
			}
			count := 0
			for _, cards := range clashRevealedCards(f.g) {
				count += len(cards)
			}
			if count != tc.wantRevealCount {
				t.Fatalf("revealed %d cards, want %d", count, tc.wantRevealCount)
			}
			if len(f.g.FailedDraws) != 0 {
				t.Fatalf("clashing with an empty library recorded a draw attempt: %v", f.g.FailedDraws)
			}
		})
	}
}

func TestOakenBrawlerEnterTriggerClashesThroughTheTriggeredProgramRoute(t *testing.T) {
	for _, tc := range []struct {
		name         string
		mine, theirs int
		wantCounters int
	}{
		{name: "win adds a counter", mine: 5, theirs: 1, wantCounters: 1},
		{name: "loss adds nothing", mine: 1, theirs: 5},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newClashFixture(t)
			addCardToLibrary(f.g, game.Player1, clashLibraryCard("Mine", tc.mine))
			addCardToLibrary(f.g, game.Player2, clashLibraryCard("Theirs", tc.theirs))
			cardID := f.cast(oakenBrawlerTestDef(), mana.W, 3, nil)
			agents := clashAgents(t, game.Player2, nil, nil)
			f.engine.resolveTopOfStackWithChoices(f.g, agents, &TurnLog{})
			if !f.engine.putTriggeredAbilitiesOnStackWithChoices(f.g, agents, &TurnLog{}) {
				t.Fatal("Oaken Brawler's enters trigger was not put on the stack")
			}
			f.engine.resolveTopOfStackWithChoices(f.g, agents, &TurnLog{})

			brawler := permanentForCard(f.g, cardID)
			if brawler == nil {
				t.Fatal("Oaken Brawler is not on the battlefield")
			}
			if got := brawler.Counters.Get(counter.PlusOnePlusOne); got != tc.wantCounters {
				t.Fatalf("+1/+1 counters = %d, want %d (result %+v)", got, tc.wantCounters, f.result())
			}
		})
	}
}

func TestClashReplaysIdenticallyFromAClone(t *testing.T) {
	f := newClashFixture(t)
	for _, player := range []game.PlayerID{game.Player1, game.Player2, game.Player3, game.Player4} {
		addCardToLibrary(f.g, player, clashLibraryCard("Under", 2))
		addCardToLibrary(f.g, player, clashLibraryCard("Top", int(player)+1))
	}
	addCardToHand(f.g, game.Player4, clashLibraryCard("Hand", 1))
	addImplementationSpellToStack(f.g, game.Player1, pullingTeethTestDef(), []game.Target{game.PlayerTarget(game.Player4)})
	clone := f.g.Clone()
	bottom := map[game.PlayerID]bool{game.Player1: true, game.Player4: true}

	f.engine.resolveTopOfStackWithChoices(f.g, clashAgents(t, game.Player4, bottom, nil), &TurnLog{})
	f.engine.resolveTopOfStackWithChoices(clone, clashAgents(t, game.Player4, bottom, nil), &TurnLog{})

	results := *f.results
	if len(results) != 2 || !reflect.DeepEqual(results[0], results[1]) || !results[0].Won(game.Player4) {
		t.Fatalf("results = %+v, want identical Player4 wins", results)
	}
	for player := range game.PlayerID(game.NumPlayers) {
		if !slices.Equal(f.g.Players[player].Library.All(), clone.Players[player].Library.All()) {
			t.Fatalf("Player%d's library diverged after replay", player+1)
		}
	}
	if !reflect.DeepEqual(clashRevealedCards(f.g), clashRevealedCards(clone)) {
		t.Fatal("replayed reveals diverged")
	}
}

func TestClashWithAReferencedOpponentSkipsTheChoice(t *testing.T) {
	def := &game.CardDef{CardFace: game.CardFace{
		Name:  "Fixed Clash",
		Types: []types.Card{types.Sorcery},
		SpellAbility: opt.Val(game.Mode{
			Targets: []game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Constraint: "target player", Allow: game.TargetAllowPlayer}},
			Program: fixedClashTestReference,
		}.Ability()),
	}}
	t.Run("opponent", func(t *testing.T) {
		f := newClashFixture(t)
		addCardToLibrary(f.g, game.Player3, clashLibraryCard("Theirs", 1))
		addImplementationSpellToStack(f.g, game.Player1, def, []game.Target{game.PlayerTarget(game.Player3)})
		var requests []game.ChoiceRequest
		f.engine.resolveTopOfStackWithChoices(f.g, clashAgents(t, game.Player2, nil, &requests), &TurnLog{})
		result := f.result()
		if !result.Available || result.Opponent != game.Player3 || !result.Won(game.Player3) {
			t.Fatalf("result = %+v, want a clash with the referenced Player3", result)
		}
		for _, request := range requests {
			if request.Kind == game.ChoicePlayer {
				t.Fatal("a referenced opponent was chosen again")
			}
		}
	})
	t.Run("controller is not an opponent", func(t *testing.T) {
		f := newClashFixture(t)
		addCardToLibrary(f.g, game.Player1, clashLibraryCard("Mine", 1))
		addImplementationSpellToStack(f.g, game.Player1, def, []game.Target{game.PlayerTarget(game.Player1)})
		f.engine.resolveTopOfStackWithChoices(f.g, clashAgents(t, game.Player2, nil, nil), &TurnLog{})
		if result := f.result(); result.Available || len(clashRevealedCards(f.g)) != 0 {
			t.Fatalf("result = %+v, want no clash with the controller", result)
		}
	})
}

func TestClashRevealsStayPublicForEveryLaterDecisionInTheEffect(t *testing.T) {
	f := newClashFixture(t)
	mine := addCardToLibrary(f.g, game.Player1, clashLibraryCard("Mine", 3))
	theirs := addCardToLibrary(f.g, game.Player2, clashLibraryCard("Theirs", 1))
	for range 3 {
		addCardToHand(f.g, game.Player4, clashLibraryCard("Hand", 1))
	}
	addImplementationSpellToStack(f.g, game.Player1, pullingTeethTestDef(), []game.Target{game.PlayerTarget(game.Player4)})
	agents := clashAgents(t, game.Player2, map[game.PlayerID]bool{game.Player2: true}, nil)
	var discards []game.ChoiceRequest
	var observed [][]game.ClashDecision
	// Player4 neither clashed nor revealed, yet chooses its discards knowing
	// both public reveals, where each went, and so why it discards two.
	agents[game.Player4] = authoredChoicePolicy{
		simPassPolicy: simPassPolicy{},
		choose: func(obs PlayerObservation, request game.ChoiceRequest) []int {
			discards = append(discards, request)
			observed = append(observed, obs.ClashReveals())
			if clone := obs.g.Clone(); !reflect.DeepEqual(clone.PublicClashReveals(), obs.ClashReveals()) {
				t.Error("a clone taken mid-effect lost the public clash reveals")
			}
			return slices.Clone(request.DefaultSelection)
		},
	}
	log := &TurnLog{}

	f.engine.resolveTopOfStackWithChoices(f.g, agents, log)

	if len(discards) != 1 || discards[0].Player != game.Player4 {
		t.Fatalf("discard requests = %+v, want one for the non-clashing target", discards)
	}
	if discarded := 3 - f.g.Players[game.Player4].Hand.Size(); discarded != 2 {
		t.Fatalf("target discarded %d, want 2 after the controller's win", discarded)
	}
	got := discards[0].ClashDecisions
	if len(got) != 2 || got[0].Player != game.Player1 || got[0].Card.CardID != mine || !got[0].Decided || got[0].Bottom ||
		got[1].Player != game.Player2 || got[1].Card.CardID != theirs || !got[1].Decided || !got[1].Bottom {
		t.Fatalf("discard request saw %+v, want both reveals with their final placements", got)
	}
	if !reflect.DeepEqual(observed[0], got) {
		t.Fatalf("observation saw %+v, want the request's public reveals %+v", observed[0], got)
	}
	logged := false
	for _, choice := range log.Choices {
		if choice.Request.Player == game.Player4 && reflect.DeepEqual(choice.Request.ClashDecisions, got) {
			logged = true
		}
	}
	if !logged {
		t.Fatal("the logged discard decision omitted the public clash reveals")
	}
	if len(f.g.ClashReveals) != 0 || len(NewObservation(f.g, game.Player3).ClashReveals()) != 0 {
		t.Fatalf("clash reveals = %+v after the effect finished, want none", f.g.ClashReveals)
	}
}
