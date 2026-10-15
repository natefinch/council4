package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

type productionBatchScenario struct {
	entry       cardset.Entry
	family      string
	counterpart string
}

func productionBatchScenarios() []productionBatchScenario {
	var result []productionBatchScenario
	for _, tc := range productionDestructionCases() {
		counterpart := "Basilisk Collar"
		if tc.noRegrowth {
			counterpart = "Birds of Paradise"
		}
		result = append(result, productionBatchScenario{tc.entry, "destruction", counterpart})
	}
	for _, tc := range productionCounterCases() {
		counterpart := "Lightning Bolt"
		if tc.creature {
			counterpart = "Birds of Paradise"
		}
		result = append(result, productionBatchScenario{tc.entry, "stack-interaction", counterpart})
	}
	for _, tc := range productionTutorCases() {
		counterpart := "Basilisk Collar"
		switch tc.quality {
		case types.Creature:
			counterpart = "Birds of Paradise"
		case types.Enchantment:
			counterpart = "Humility"
		case types.Land:
			counterpart = "Forest"
		default:
		}
		result = append(result, productionBatchScenario{tc.entry, "library-search", counterpart})
	}
	return result
}

func runProductionBatchScenario(tb testing.TB, registry *cards.Registry, scenario productionBatchScenario, simulator bool) (*game.Game, int) {
	tb.Helper()
	g, engine, policies, policy := mixedProductionGame(registry)
	def, err := registry.LookupOracleID(scenario.entry.OracleID)
	if err != nil {
		tb.Fatal(err)
	}
	var subjectCard game.ObjectID
	var targets []game.Target
	librarySize, handSize := g.Players[game.Player1].Library.Size(), g.Players[game.Player1].Hand.Size()
	switch scenario.family {
	case "destruction":
		subject := addCombatPermanent(g, game.Player2, registry.Lookup(scenario.counterpart))
		subjectCard = subject.CardInstanceID
		targets = []game.Target{game.PermanentTarget(subject.ObjectID)}
	case "stack-interaction":
		subjectCard = addCardToHand(g, game.Player2, registry.Lookup(scenario.counterpart))
		subjectTargets := []game.Target{game.PlayerTarget(game.Player3)}
		if scenario.counterpart == "Birds of Paradise" {
			subjectTargets = nil
		}
		g.Players[game.Player2].ManaPool.Add(mana.R, 1)
		g.Players[game.Player2].ManaPool.Add(mana.G, 1)
		setMainPhasePriority(g, game.Player2)
		if !engine.applyAction(g, game.Player2, action.CastSpell(subjectCard, subjectTargets, 0, nil)) {
			tb.Fatal("real legacy counter subject did not cast")
		}
		subject, _ := g.Stack.Peek()
		targets = []game.Target{game.StackObjectTarget(subject.ID)}
	case "library-search":
		addCardToLibrary(g, game.Player1, registry.Lookup(scenario.counterpart))
		librarySize++
	default:
		tb.Fatalf("unknown scenario family %s", scenario.family)
	}
	spell := addCardToHand(g, game.Player1, def)
	for _, c := range []mana.Color{mana.C, mana.W, mana.U, mana.B, mana.R, mana.G} {
		g.Players[game.Player1].ManaPool.Add(c, 5)
	}
	setMainPhasePriority(g, game.Player1)
	act := action.CastSpell(spell, targets, 0, nil)
	if !containsAction(engine.Simulator().LegalActions(g, game.Player1), act) {
		tb.Fatal("production legal actions did not expose the real card/targets")
	}
	beforeStack := g.Stack.Size()
	if simulator {
		branch, accepted := engine.Simulator().Apply(g, game.Player1, act, policies)
		if !accepted || g.Stack.Size() != beforeStack || !g.Players[game.Player1].Hand.Contains(spell) {
			tb.Fatal("Simulator did not accept or mutated its input")
		}
		g = engine.Simulator().ResolvePriority(branch, policies)
	} else {
		if !engine.applyActionWithChoices(g, game.Player1, act, policies, &TurnLog{}) {
			tb.Fatal("live production cast failed")
		}
		engine.runPriorityLoop(g, policies, &TurnLog{})
	}
	if !g.Stack.IsEmpty() || !g.Players[game.Player1].Graveyard.Contains(spell) ||
		g.Execution.ActionFallbacks != 0 || g.Execution.ChoiceFallbacks != 0 {
		tb.Fatal("production scenario did not finish at stable priority without fallbacks")
	}
	switch scenario.family {
	case "destruction", "stack-interaction":
		if !g.Players[game.Player2].Graveyard.Contains(subjectCard) {
			tb.Fatal("actual legacy subject did not reach owner's graveyard")
		}
	case "library-search":
		if g.Players[game.Player1].Library.Size() != librarySize-1 ||
			g.Players[game.Player1].Hand.Size() != handSize+1 {
			tb.Fatal("production tutor did not actually find/place a card")
		}
	default:
		tb.Fatal("unknown outcome family")
	}
	return g, policy.decisions
}

func TestProductionBatchLiveAndSimulator(t *testing.T) {
	registry := cards.NewDefaultRegistry()
	for _, scenario := range productionBatchScenarios() {
		t.Run(scenario.entry.Name, func(t *testing.T) {
			for _, simulator := range []bool{false, true} {
				t.Run(map[bool]string{false: "live", true: "Simulator"}[simulator], func(t *testing.T) {
					_, decisions := runProductionBatchScenario(t, registry, scenario, simulator)
					if decisions == 0 {
						t.Fatal("real priority/choice policy was not exercised")
					}
				})
			}
		})
	}
}

func BenchmarkProductionBatches(b *testing.B) {
	registry := cards.NewDefaultRegistry()
	for _, scenario := range productionBatchScenarios() {
		runProductionBatchScenario(b, registry, scenario, true)
		b.Run(scenario.family+"/"+scenario.entry.Name, func(b *testing.B) {
			b.ReportAllocs()
			decisions := 0
			for range b.N {
				_, decisions = runProductionBatchScenario(b, registry, scenario, true)
			}
			b.ReportMetric(float64(decisions), "decisions/op")
			b.ReportMetric(1, "stable-priority/op")
			b.ReportMetric(1, "workers")
		})
	}
}
