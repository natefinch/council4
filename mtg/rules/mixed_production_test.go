package rules

import (
	"math/rand/v2"
	"reflect"
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/cards/authored"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
)

// The same fixture/seed/policy is copied unchanged into the step-11 baseline.
// Scenario completion is stable priority, not a complete-game rules draw.
const mixedProductionSeed = 20261014

type mixedPolicy struct {
	decisions int
}

func (p *mixedPolicy) ChooseAction(PlayerObservation, []action.Action) action.Action {
	p.decisions++
	return action.Pass()
}

func (p *mixedPolicy) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	p.decisions++
	if request.Kind == game.ChoiceMay {
		return []int{1}
	}
	if request.Kind == game.ChoiceTarget {
		best, count := -1, -1
		for _, option := range request.Options {
			if len(option.Targets) > count {
				best, count = option.Index, len(option.Targets)
			}
		}
		if best >= 0 && count > 1 {
			return []int{best}
		}
	}
	return fallbackChoice(request)
}

//nolint:revive // The concrete benchmark fixture returns its state, engine, seats and independently counted policy.
func mixedProductionGame(registry *cards.Registry) (*game.Game, *Engine, [game.NumPlayers]PlayerAgent, *mixedPolicy) {
	engine := NewEngineWithPrograms(rand.New(rand.NewPCG(mixedProductionSeed, mixedProductionSeed^7)), registry.Programs())
	var configs [game.NumPlayers]game.PlayerConfig
	for i := range configs {
		configs[i].Deck = repeatedCard(registry.Lookup("Forest"), 40)
	}
	g := engine.NewGame(configs)
	setMainPhasePriority(g, game.Player1)
	policy := &mixedPolicy{}
	var policies [game.NumPlayers]PlayerAgent
	for i := range policies {
		policies[i] = policy
	}
	return g, engine, policies, policy
}

func mixedResolve(tb testing.TB, engine *Engine, g *game.Game, actor game.PlayerID, act action.Action, policies [game.NumPlayers]PlayerAgent) *game.Game {
	tb.Helper()
	setMainPhasePriority(g, actor)
	applied, ok := engine.Simulator().Apply(g, actor, act, policies)
	if !ok {
		tb.Fatalf("production action not accepted: %+v", act)
	}
	return engine.Simulator().ResolvePriority(applied, policies)
}

type mixedScenario struct {
	name string
	run  func(testing.TB, *cards.Registry) (*game.Game, int)
}

func mixedScenarios() []mixedScenario {
	return []mixedScenario{
		{"HiveMind-Bolt", func(tb testing.TB, registry *cards.Registry) (*game.Game, int) {
			g, engine, policies, policy := mixedProductionGame(registry)
			addCombatPermanent(g, game.Player1, registry.Lookup("Hive Mind"))
			card := addCardToHand(g, game.Player1, registry.Lookup("Lightning Bolt"))
			g.Players[game.Player1].ManaPool.Add(mana.R, 1)
			g = mixedResolve(tb, engine, g, game.Player1, action.CastSpell(card, []game.Target{game.PlayerTarget(game.Player2)}, 0, nil), policies)
			if g.Players[game.Player2].Life != 28 || !g.Stack.IsEmpty() {
				tb.Fatal("production copies did not deal four Bolts' damage")
			}
			return g, policy.decisions
		}},
		{"Terastodon-Collar", func(tb testing.TB, registry *cards.Registry) (*game.Game, int) {
			g, engine, policies, policy := mixedProductionGame(registry)
			ring := addCombatPermanent(g, game.Player2, registry.Lookup("Basilisk Collar"))
			land := addCombatPermanent(g, game.Player3, registry.Lookup("Forest"))
			card := addCardToHand(g, game.Player1, registry.Lookup("Terastodon"))
			g.Players[game.Player1].ManaPool.Add(mana.C, 6)
			g.Players[game.Player1].ManaPool.Add(mana.G, 2)
			g = mixedResolve(tb, engine, g, game.Player1, action.CastSpell(card, nil, 0, nil), policies)
			// Trigger target choices are made by the engine policy; require the
			// actual resulting destruction/token payoff, not CardDef structure.
			if !g.Players[game.Player2].Graveyard.Contains(ring.CardInstanceID) ||
				!g.Players[game.Player3].Graveyard.Contains(land.CardInstanceID) ||
				namedTokenCount(g, game.Player2, "Elephant")+namedTokenCount(g, game.Player3, "Elephant") != 2 {
				tb.Fatal("production destruction did not create controller-specific tokens")
			}
			return g, policy.decisions
		}},
		{"Worldgorger-BeastWithin", func(tb testing.TB, registry *cards.Registry) (*game.Game, int) {
			g, engine, policies, policy := mixedProductionGame(registry)
			ring := addCombatPermanent(g, game.Player1, registry.Lookup("Basilisk Collar"))
			card := addCardToHand(g, game.Player1, registry.Lookup("Worldgorger Dragon"))
			g.Players[game.Player1].ManaPool.Add(mana.C, 3)
			g.Players[game.Player1].ManaPool.Add(mana.R, 3)
			g = mixedResolve(tb, engine, g, game.Player1, action.CastSpell(card, nil, 0, nil), policies)
			dragon := permanentByCardID(g, card)
			if dragon == nil {
				tb.Fatal("production Dragon did not enter")
				return g, policy.decisions
			}
			if !g.Players[game.Player1].Exile.Contains(ring.CardInstanceID) {
				tb.Fatal("production Dragon did not exile linked legacy permanent")
			}
			swords := addCardToHand(g, game.Player2, registry.Lookup("Beast Within"))
			g.Players[game.Player2].ManaPool.Add(mana.C, 2)
			g.Players[game.Player2].ManaPool.Add(mana.G, 1)
			g = mixedResolve(tb, engine, g, game.Player2, action.CastSpell(swords, []game.Target{game.PermanentTarget(dragon.ObjectID)}, 0, nil), policies)
			returned := permanentByCardID(g, ring.CardInstanceID)
			if returned == nil || returned.ObjectID == ring.ObjectID || !g.Players[game.Player1].Graveyard.Contains(card) {
				tb.Fatal("production linked return did not survive legacy removal")
			}
			return g, policy.decisions
		}},
	}
}

func TestMixedProductionInteractions(t *testing.T) {
	registry := cards.NewDefaultRegistry()
	for _, scenario := range mixedScenarios() {
		t.Run(scenario.name, func(t *testing.T) {
			g, decisions := scenario.run(t, registry)
			if decisions == 0 || g.Execution.Decisions != decisions || g.Execution.ActionFallbacks != 0 || g.Execution.ChoiceFallbacks != 0 {
				t.Fatalf("decision/fallback accounting: %+v, policy %d", g.Execution, decisions)
			}
			again, _ := scenario.run(t, registry)
			if !reflect.DeepEqual(g.Events, again.Events) {
				t.Fatal("fixed-seed mixed scenario is not reproducible")
			}
		})
	}
}

func BenchmarkMixedProduction(b *testing.B) {
	registry := cards.NewDefaultRegistry()
	for _, scenario := range mixedScenarios() {
		scenario.run(b, registry) // warm lazy definitions equally in each baseline
		b.Run(scenario.name, func(b *testing.B) {
			b.ReportAllocs()
			decisions := 0
			for range b.N {
				_, decisions = scenario.run(b, registry)
			}
			b.ReportMetric(float64(decisions), "decisions/op")
			b.ReportMetric(1, "stable-priority/op")
			b.ReportMetric(1, "workers")
		})
	}
}

func TestEveryProvingCardLoadsIntoLiveAndSearchEngines(t *testing.T) {
	registry := cards.NewDefaultRegistry()
	for _, entry := range authored.Cards {
		t.Run(entry.Name, func(t *testing.T) {
			def, err := registry.LookupOracleID(entry.OracleID)
			if err != nil {
				t.Fatal(err)
			}
			g, engine, policies, _ := mixedProductionGame(registry)
			addCombatPermanent(g, game.Player2, registry.Lookup("Birds of Paradise"))
			cardID := addCardToHand(g, game.Player1, def)
			for _, color := range []mana.Color{mana.C, mana.W, mana.U, mana.B, mana.R, mana.G} {
				g.Players[game.Player1].ManaPool.Add(color, 10)
			}
			var cast action.Action
			found := false
			for _, act := range engine.Simulator().LegalActions(g, game.Player1) {
				if payload, ok := act.CastSpellPayload(); ok && payload.CardID == cardID {
					cast, found = act, true
					break
				}
			}
			if !found {
				t.Fatal("proving card not castable through production legal actions")
			}
			world := engine.newSearchContext(g, game.Player1).Determinize()
			search := engine.newSearchContext(g, game.Player1).Simulator()
			branch, ok := search.Apply(world, game.Player1, cast, policies)
			if !ok {
				t.Fatal("determinized production cast not accepted")
			}
			branch = search.ResolvePriority(branch, policies)
			live := mixedResolve(t, engine, g, game.Player1, cast, policies)
			if !branch.Stack.IsEmpty() || !live.Stack.IsEmpty() {
				t.Fatal("live/search catalogs failed to reach stable priority")
			}
			if !live.Players[game.Player1].Graveyard.Contains(cardID) && permanentByCardID(live, cardID) == nil {
				// Gilded Drake can be sacrificed by its actual unsuccessful exchange.
				t.Fatal("proving card vanished after production resolution")
			}
		})
	}
}

func TestDiagnosticsCloneAndHiddenInformation(t *testing.T) {
	registry := cards.NewDefaultRegistry()
	g, _, _, _ := mixedProductionGame(registry)
	addCardToHand(g, game.Player2, registry.Lookup("Mindslaver"))
	g.Execution.Decisions = 42
	clone := g.Clone()
	clone.Execution.Decisions++
	if g.Execution.Decisions != 42 || clone.Execution.Decisions != 43 {
		t.Fatal("execution diagnostics lost clone isolation")
	}
	if len(NewObservation(g, game.Player1).Hand()) != 0 || len(NewObservation(g, game.Player2).Hand()) != 1 {
		t.Fatal("evaluation metadata leaked a hidden hand")
	}
	if !reflect.DeepEqual(NewObservation(g, game.Player1).FidelityLimitations(), NewObservation(g, game.Player2).FidelityLimitations()) {
		t.Fatal("fidelity metadata leaked seat-specific hidden information")
	}
}

type exhaustedSearchPolicy struct{ mixedPolicy }

func (*exhaustedSearchPolicy) ChooseActionBySearch(ctx SearchContext, legal []action.Action) action.Action {
	world := ctx.Determinize()
	if world.Execution.Decisions != 0 {
		panic("hypothetical execution inherited live work")
	}
	world.Execution.Decisions = DefaultSearchDecisionLimit
	ctx.Simulator().ResolvePriority(world, [game.NumPlayers]PlayerAgent{})
	return legal[0]
}

func TestHypotheticalBudgetExhaustionDoesNotAbortLiveGame(t *testing.T) {
	g, engine, policies, _ := mixedProductionGame(cards.NewDefaultRegistry())
	engine.DecisionLimit = 2
	policies[game.Player1] = &exhaustedSearchPolicy{}
	legal := engine.legalActions(g, game.Player1)
	chosen := engine.chooseAction(g, policies, game.Player1, legal, action.Pass())
	if chosen.Kind != action.ActionPass || g.Execution.Decisions != 1 || g.Execution.SearchBudgetStops != 1 ||
		g.Execution.ActionFallbacks != 1 || g.Execution.InvalidActions != 0 {
		t.Fatalf("speculative exhaustion contaminated live result: %+v", g.Execution)
	}
}

type invalidChoicePolicy struct{ mixedPolicy }

func (*invalidChoicePolicy) ChooseChoice(PlayerObservation, game.ChoiceRequest) []int {
	return []int{999}
}

func TestInvalidChoiceFallbackIsExplicit(t *testing.T) {
	g, engine, policies, _ := mixedProductionGame(cards.NewDefaultRegistry())
	policies[game.Player1] = &invalidChoicePolicy{}
	log := &TurnLog{}
	engine.chooseChoice(g, policies, game.ChoiceRequest{
		Kind: game.ChoiceMay, Player: game.Player1, MinChoices: 1, MaxChoices: 1,
		Options: []game.ChoiceOption{{Index: 0}, {Index: 1}}, DefaultSelection: []int{1},
	}, log)
	if g.Execution.InvalidChoices != 1 || g.Execution.ChoiceFallbacks != 1 || !log.Choices[0].UsedFallback {
		t.Fatal("invalid choice/fallback was silent")
	}
}

func TestOpeningSimultaneousLossesAreRulesDrawNotBudget(t *testing.T) {
	engine := NewEngine(nil)
	g := engine.NewGame([game.NumPlayers]game.PlayerConfig{})
	result := engine.RunGameWithTurnLimit(g, [game.NumPlayers]PlayerAgent{}, 1)
	if result.Outcome() != TerminationDraw || result.TurnLimitReached || !result.Completed() {
		t.Fatalf("actual rules draw misclassified: %+v", result)
	}
	for player := range game.NumPlayers {
		if !result.Drew(game.PlayerID(player)) {
			t.Fatal("terminal simultaneous loss set lost its draw")
		}
	}
}

func TestPreviouslyEliminatedSeatDoesNotJoinTerminalDraw(t *testing.T) {
	g, engine, policies, _ := mixedProductionGame(cards.NewDefaultRegistry())
	engine.eliminatePlayer(g, game.Player1)
	for _, player := range g.Players[1:] {
		player.Life = 0
	}
	result := engine.RunGame(g, policies)
	if result.Outcome() != TerminationDraw || result.Drew(game.Player1) || !result.Drew(game.Player2) ||
		!result.Drew(game.Player3) || !result.Drew(game.Player4) {
		t.Fatalf("prior and terminal losses confused: %v", result.DrawParticipants)
	}
	clone := g.Clone()
	clone.DrawParticipants[game.Player2] = false
	if !g.DrawParticipants[game.Player2] {
		t.Fatal("draw membership shared between search branches")
	}
}

func TestGoldfishEliminationIsRulesLossNotDraw(t *testing.T) {
	engine := NewEngine(nil)
	g := engine.NewGoldfishGame(game.PlayerConfig{})
	result := engine.RunGoldfish(g, nil, 1)
	if result.Outcome() != TerminationLoss || !result.Completed() || result.Drew(game.Player1) {
		t.Fatalf("single-player loss misclassified: %s", result.Outcome())
	}
}
