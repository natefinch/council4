package rules

import (
	"fmt"
	"reflect"
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/mtg/rules/payment"
	"github.com/natefinch/council4/opt"
)

// Rules source: official MagicCompRules 20261106.txt, effective 2026-11-06.
// CR 118.3, 601.2b/f/h, 602.2b, 107.4f, 702.51, 702.66, 702.126.
func requireCompletePayment(t *testing.T, g *game.Game, req payment.Request, partial payment.Partial) payment.Plan {
	t.Helper()
	plan, ok, reason := paymentOrch.planner(g).Complete(req, partial)
	if !ok {
		t.Fatalf("Complete: %s", reason)
	}
	return plan
}

func requireApplyPayment(t *testing.T, g *game.Game, plan payment.Plan) payment.ApplyResult {
	t.Helper()
	result, ok, reason := paymentOrch.planner(g).Apply(plan)
	if !ok {
		t.Fatalf("Apply: %s", reason)
	}
	return result
}

func TestPaymentCompletionJointCountersAndFilteredOptions(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Counter source", Types: []types.Card{types.Artifact}}})
	source.Counters.Add(counter.Charge, 1)
	source.Counters.Add(counter.Defense, 1)
	req := payment.ForAbility(payment.AbilityRequest{PlayerID: game.Player1, Source: source, AdditionalCosts: []cost.Additional{
		{Kind: cost.AdditionalRemoveCounter, AnyCounterKind: true, Amount: 1},
		{Kind: cost.AdditionalRemoveCounter, CounterKind: counter.Charge, Amount: 1},
	}})
	plan := requireCompletePayment(t, g, req, payment.Partial{})
	if got := plan.Components()[0].Counters; !slices.Equal(got, []payment.CounterRemovalChoice{{Permanent: source.ObjectID, Kind: counter.Defense}}) {
		t.Fatalf("first removal = %v, want defense reserved before mandatory charge", got)
	}
	options := paymentOrch.planner(g).Options(req, payment.Partial{}, payment.Step{Kind: payment.StepComponent, Index: 0})
	if len(options) != 1 || options[0].Component.Counters[0].Kind != counter.Defense {
		t.Fatalf("completable counter choices = %+v, want only defense", options)
	}
	requireApplyPayment(t, g, plan)
	if !source.Counters.IsEmpty() {
		t.Fatalf("counters after payment = %v", source.Counters)
	}
	if _, ok, _ := paymentOrch.planner(g).Complete(req, payment.Partial{}); ok {
		t.Fatal("exhausted counter costs were offered")
	}
}

func TestPaymentCompletionGroupedCounterKeepsSelectedKind(t *testing.T) {
	for _, spell := range []bool{false, true} {
		t.Run(fmt.Sprint("spell=", spell), func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			source := addCombatPermanent(g, game.Player1, registryCardDef(t, "Invasion of Segovia"))
			source.Counters.Add(counter.Charge, 1)
			source.Counters.Add(counter.Defense, 1)
			costs := []cost.Additional{
				{Kind: cost.AdditionalRemoveCounterAmong, ChoiceGroup: 1, AnyCounterKind: true, Amount: 1},
				{Kind: cost.AdditionalPayLife, ChoiceGroup: 1, Amount: 2},
			}
			req := payment.ForAbility(payment.AbilityRequest{PlayerID: game.Player1, Source: source, AdditionalCosts: costs})
			if spell {
				def := &game.CardDef{CardFace: game.CardFace{Name: "Grouped cost", Types: []types.Card{types.Instant}, AdditionalCosts: costs}}
				card := addCardToHand(g, game.Player1, def)
				req = payment.ForSpell(payment.SpellRequest{PlayerID: game.Player1, CardID: card, Card: def, SourceZone: zone.Hand})
			}
			partial := payment.Partial{Groups: map[uint8]int{1: 0}, Components: map[int]payment.ComponentChoice{
				0: {Counters: []payment.CounterRemovalChoice{{Permanent: source.ObjectID, Kind: counter.Defense}}, Finished: true},
			}}
			plan := requireCompletePayment(t, g, req, partial)
			requireApplyPayment(t, g, plan)
			if source.Counters.Get(counter.Charge) != 1 || source.Counters.Get(counter.Defense) != 0 {
				t.Fatalf("grouped selected counter was lost: %v", source.Counters)
			}
		})
	}
}

func TestPaymentCompletionMillAffordabilityAndLibraryReservations(t *testing.T) {
	for _, tc := range []struct {
		library int
		amounts []int
		exile   bool
		want    bool
	}{
		{0, []int{2}, false, false}, {1, []int{2}, false, false}, {2, []int{2}, false, true},
		{3, []int{2, 2}, false, false}, {4, []int{2, 2}, false, true},
		{2, []int{2}, true, false}, {3, []int{2}, true, true},
	} {
		t.Run(fmt.Sprintf("%d/%v/exile=%v", tc.library, tc.amounts, tc.exile), func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			for i := range tc.library {
				addCardToLibrary(g, game.Player1, discardableCard(fmt.Sprintf("Library %d", i)))
			}
			var costs []cost.Additional
			for _, amount := range tc.amounts {
				costs = append(costs, cost.Additional{Kind: cost.AdditionalMill, Amount: amount})
			}
			if tc.exile {
				costs = append(costs, cost.Additional{Kind: cost.AdditionalExile, Source: zone.Library, Amount: 1})
			}
			req := payment.GenericRequest{PlayerID: game.Player1, AdditionalCosts: costs}
			planner := paymentOrch.planner(g)
			plan, ok, _ := planner.Complete(payment.ForGeneric(req), payment.Partial{})
			if ok != tc.want || planner.CanPayGenericCost(req) != tc.want {
				t.Fatalf("completion/compatibility = %v/%v, want %v", ok, planner.CanPayGenericCost(req), tc.want)
			}
			if tc.want {
				requireApplyPayment(t, g, plan)
				if g.Players[game.Player1].Library.Size() != 0 {
					t.Fatal("affordable library costs were not fully paid")
				}
			} else {
				before := g.Clone()
				if _, paid := planner.PayGenericCost(req); paid || !reflect.DeepEqual(g.Clone(), before) {
					t.Fatal("rejected library payment mutated the game")
				}
			}
		})
	}
	t.Run("mill destination replacement still pays", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		colossus := addCardToLibrary(g, game.Player1, registryCardDef(t, "Blightsteel Colossus"))
		req := payment.ForGeneric(payment.GenericRequest{PlayerID: game.Player1, AdditionalCosts: []cost.Additional{{Kind: cost.AdditionalMill, Amount: 1}}})
		requireApplyPayment(t, g, requireCompletePayment(t, g, req, payment.Partial{}))
		if !g.Players[game.Player1].Library.Contains(colossus) {
			t.Fatal("the mill payment bypassed Blightsteel's replacement")
		}
	})
}

func TestPaymentCompletionSeparateRandomDiscardCostsAndSubjects(t *testing.T) {
	for _, bound := range []bool{false, true} {
		t.Run(fmt.Sprint("bound=", bound), func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			addCombatPermanent(g, game.Player1, registryCardDef(t, "Rielle, the Everwise"))
			for i := range 3 {
				addCardToHand(g, game.Player1, discardableCard(fmt.Sprintf("Random fuel %d", i)))
				addCardToLibrary(g, game.Player1, discardableCard("Draw fuel"))
			}
			costs := []cost.Additional{{Kind: cost.AdditionalDiscard, Random: true, Amount: 2}, {Kind: cost.AdditionalDiscard, Random: true, Amount: 1}}
			if bound {
				costs[0].SubjectKey, costs[1].SubjectKey = "first", "second"
			}
			req := payment.ForGeneric(payment.GenericRequest{PlayerID: game.Player1, AdditionalCosts: costs})
			start := len(g.Events)
			result := requireApplyPayment(t, g, requireCompletePayment(t, g, req, payment.Partial{}))
			var order []id.ID
			sizes := map[id.ID]int{}
			for _, event := range cohortDiscardEvents(g, start) {
				if !slices.Contains(order, event.SimultaneousID) {
					order = append(order, event.SimultaneousID)
				}
				sizes[event.SimultaneousID]++
			}
			if len(order) != 2 || sizes[order[0]] != 2 || sizes[order[1]] != 1 {
				t.Fatalf("random discard cohorts = %v/%v, want [2,1]", order, sizes)
			}
			if bound {
				subjects := result.Ability.PaidCostSubjects
				if len(subjects) != 3 || subjects[0].Key != "first" || subjects[1].Key != "first" || subjects[2].Key != "second" {
					t.Fatalf("random paid subjects = %+v", subjects)
				}
			}
			if !engine.putTriggeredAbilitiesOnStack(g) {
				t.Fatal("Rielle did not trigger for the first random discard")
			}
			engine.resolveTopOfStack(g, &TurnLog{})
			if g.Players[game.Player1].Hand.Size() != 2 || engine.putTriggeredAbilitiesOnStack(g) {
				t.Fatal("Rielle did not draw exactly two for the first component")
			}
		})
	}
}

func TestPaymentCompletionRegisteredPhyrexianGenericCompetition(t *testing.T) {
	for _, exile := range []bool{false, true} {
		t.Run(fmt.Sprint("Evelyn=", exile), func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			e := NewEngine(nil)
			setMainPhasePriority(g, game.Player1)
			def := registryCardDef(t, "Porcelain Legionnaire")
			card := addCardToHand(g, game.Player1, def)
			manaColor := mana.W
			act := action.CastSpell(card, nil, 0, nil)
			if exile {
				g.Players[game.Player1].Hand.Remove(card)
				card = addCardToExile(g, game.Player2, def)
				addCombatPermanent(g, game.Player1, registryCardDef(t, "Evelyn, the Covetous"))
				g.AddExileCounterFromController(card, counter.Collection, 1, game.Player1)
				manaColor = mana.U
				act = action.CastSpellFromZone(card, zone.Exile, nil, 0, nil)
			}
			g.Players[game.Player1].ManaPool.Add(manaColor, 2)
			before := g.Players[game.Player1].Life
			if !containsAction(e.legalActions(g, game.Player1), act) || !e.applyAction(g, game.Player1, act) {
				t.Fatal("registered Legionnaire's complete two-mana/two-life payment was rejected")
			}
			if g.Players[game.Player1].Life != before-2 || g.Players[game.Player1].ManaPool.Total() != 0 {
				t.Fatal("Phyrexian life and generic mana were not accounted exactly")
			}
		})
	}
}

func TestPaymentCompletionStaleAtomicAndCloneSafe(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Clone source", Types: []types.Card{types.Artifact}}})
	source.Counters.Add(counter.Charge, 1)
	manaCost := cost.Mana{cost.W}
	g.Players[game.Player1].ManaPool.Add(mana.W, 1)
	req := payment.ForAbility(payment.AbilityRequest{PlayerID: game.Player1, Source: source, ManaCost: opt.Val(manaCost), AdditionalCosts: []cost.Additional{{Kind: cost.AdditionalRemoveCounter, CounterKind: counter.Charge, Amount: 1}}})
	plan := requireCompletePayment(t, g, req, payment.Partial{})
	manaCost[0] = cost.O(20)
	components := plan.Components()
	delete(components, 0)
	allocation := plan.Mana()
	clear(allocation.PoolSpend)
	clone := g.Clone()
	requireApplyPayment(t, clone, plan)
	if source.Counters.Get(counter.Charge) != 1 || g.Players[game.Player1].ManaPool.Total() != 1 {
		t.Fatal("applying the ID-based witness in a clone mutated the original")
	}
	source.Counters.Remove(counter.Charge, 1)
	before := g.Clone()
	if _, ok, reason := paymentOrch.planner(g).Apply(plan); ok || reason == "" || !reflect.DeepEqual(g.Clone(), before) {
		t.Fatal("stale plan was not rejected atomically with a reason")
	}
}

func TestPaymentCompletionConvokeImproviseDelveAndX(t *testing.T) {
	for _, keyword := range []game.Keyword{game.Convoke, game.Improvise, game.Delve} {
		t.Run(fmt.Sprint(keyword), func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			def := &game.CardDef{CardFace: game.CardFace{Name: "Mixed mana mechanics", ManaCost: opt.Val(cost.Mana{cost.X, cost.W}), Types: []types.Card{types.Sorcery},
				StaticAbilities: []game.StaticAbility{{ZoneOfFunction: zone.Stack, KeywordAbilities: game.SimpleKeywords(keyword)}}}}
			card := addCardToHand(g, game.Player1, def)
			objects := make([]id.ID, 0, 2)
			for i := range 2 {
				permanent := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: fmt.Sprint("Fuel", i), Types: []types.Card{types.Artifact, types.Creature}, Colors: []color.Color{color.White}}})
				if keyword == game.Delve {
					objects = append(objects, addCardToGraveyard(g, game.Player1, discardableCard("Delve fuel")))
				} else {
					objects = append(objects, permanent.ObjectID)
				}

			}
			g.Players[game.Player1].ManaPool.Add(mana.W, 1)
			partial := payment.Partial{X: opt.Val(2)}
			switch keyword {
			case game.Convoke:
				partial.Convoke = opt.Val(objects)
			case game.Improvise:
				partial.Improvise = opt.Val(objects)
			case game.Delve:
				partial.Delve = opt.Val(objects)
			}
			req := payment.ForSpell(payment.SpellRequest{PlayerID: game.Player1, CardID: card, Card: def, SourceZone: zone.Hand})
			plan := requireCompletePayment(t, g, req, partial)
			requireApplyPayment(t, g, plan)
			if g.Players[game.Player1].ManaPool.Total() != 0 || plan.XValue() != 2 {
				t.Fatal("nonmana mechanic did not pay X while mana paid the colored symbol")
			}
			options := paymentOrch.planner(g).Options(req, payment.Partial{}, payment.Step{Kind: payment.StepX, Values: []int{0, 1, 2, 3}})
			if len(options) != 0 {
				t.Fatalf("X choices after exhausting all white mana = %+v, want none", options)
			}
		})
	}
}

func TestPaymentCompletionEveryAdditionalCostKind(t *testing.T) {
	for _, tc := range []struct {
		name string
		cost cost.Additional
	}{
		{"tap source", cost.Additional{Kind: cost.AdditionalTap}},
		{"untap source", cost.Additional{Kind: cost.AdditionalUntap}},
		{"sacrifice", cost.Additional{Kind: cost.AdditionalSacrifice}},
		{"sacrifice source", cost.Additional{Kind: cost.AdditionalSacrificeSource}},
		{"tap selected", cost.Additional{Kind: cost.AdditionalTapPermanents}},
		{"tap power", cost.Additional{Kind: cost.AdditionalTapPermanents, TotalPowerAtLeast: 2}},
		{"return", cost.Additional{Kind: cost.AdditionalReturnToHand}},
		{"discard", cost.Additional{Kind: cost.AdditionalDiscard}},
		{"reveal", cost.Additional{Kind: cost.AdditionalReveal}},
		{"exile", cost.Additional{Kind: cost.AdditionalExile, Source: zone.Hand}},
		{"exile source", cost.Additional{Kind: cost.AdditionalExileSource}},
		{"evidence", cost.Additional{Kind: cost.AdditionalCollectEvidence, Amount: 2}},
		{"life", cost.Additional{Kind: cost.AdditionalPayLife, Amount: 2}},
		{"energy", cost.Additional{Kind: cost.AdditionalEnergy, Amount: 2}},
		{"mill", cost.Additional{Kind: cost.AdditionalMill}},
		{"exert", cost.Additional{Kind: cost.AdditionalExert}},
		{"add counter", cost.Additional{Kind: cost.AdditionalPutCounter, CounterKind: counter.Charge}},
		{"remove counter", cost.Additional{Kind: cost.AdditionalRemoveCounter, CounterKind: counter.Charge}},
		{"remove any counter", cost.Additional{Kind: cost.AdditionalRemoveCounter, AnyCounterKind: true}},
		{"remove among", cost.Additional{Kind: cost.AdditionalRemoveCounterAmong, CounterKind: counter.Charge}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			def := &game.CardDef{CardFace: game.CardFace{Name: "Cost source", Types: []types.Card{types.Creature, types.Artifact}, Power: opt.Val(game.PT{Value: 2}), Toughness: opt.Val(game.PT{Value: 2})}}
			source := addCombatPermanent(g, game.Player1, def)
			source.Counters.Add(counter.Charge, 1)
			source.Tapped = tc.cost.Kind == cost.AdditionalUntap
			hand := addCardToHand(g, game.Player1, discardableCard("Hand fuel"))
			addCardToLibrary(g, game.Player1, discardableCard("Library fuel"))
			addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Evidence fuel", ManaCost: opt.Val(cost.Mana{cost.O(2)})}})
			g.Players[game.Player1].EnergyCounters = 2
			beforeLife := g.Players[game.Player1].Life
			start := len(g.Events)
			req := payment.ForAbility(payment.AbilityRequest{PlayerID: game.Player1, Source: source, AdditionalCosts: []cost.Additional{tc.cost}})
			plan := requireCompletePayment(t, g, req, payment.Partial{})
			requireApplyPayment(t, g, plan)
			switch tc.cost.Kind {
			case cost.AdditionalTap, cost.AdditionalTapPermanents:
				if !source.Tapped {
					t.Fatal("tap cost did not tap its source")
				}
			case cost.AdditionalUntap:
				if source.Tapped {
					t.Fatal("untap cost left its source tapped")
				}
			case cost.AdditionalSacrifice, cost.AdditionalSacrificeSource:
				if !g.Players[game.Player1].Graveyard.Contains(source.CardInstanceID) {
					t.Fatal("sacrifice cost did not sacrifice")
				}
			case cost.AdditionalReturnToHand:
				if !g.Players[game.Player1].Hand.Contains(source.CardInstanceID) {
					t.Fatal("return cost did not return")
				}
			case cost.AdditionalDiscard:
				if !g.Players[game.Player1].Graveyard.Contains(hand) {
					t.Fatal("discard cost did not discard")
				}
			case cost.AdditionalExile, cost.AdditionalExileSource:
				card := hand
				if tc.cost.Kind == cost.AdditionalExileSource {
					card = source.CardInstanceID
				}
				if !g.Players[game.Player1].Exile.Contains(card) {
					t.Fatal("exile cost did not exile")
				}
			case cost.AdditionalCollectEvidence:
				if g.Players[game.Player1].Graveyard.Size() != 0 || g.Players[game.Player1].Exile.Size() != 1 {
					t.Fatal("evidence cost was not paid")
				}
			case cost.AdditionalPayLife:
				if g.Players[game.Player1].Life != beforeLife-2 {
					t.Fatal("life cost was not paid")
				}
			case cost.AdditionalEnergy:
				if g.Players[game.Player1].EnergyCounters != 0 {
					t.Fatal("energy cost was not paid")
				}
			case cost.AdditionalMill:
				if g.Players[game.Player1].Library.Size() != 0 {
					t.Fatal("mill cost was not paid")
				}
			case cost.AdditionalExert:
				if !source.ExertedBy[game.Player1] || source.SkipsNextUntap {
					t.Fatal("exert cost was not paid")
				}
			case cost.AdditionalPutCounter:
				if source.Counters.Get(counter.Charge) != 2 {
					t.Fatal("counter placement cost was not paid")
				}
			case cost.AdditionalRemoveCounter, cost.AdditionalRemoveCounterAmong:
				if source.Counters.Get(counter.Charge) != 0 {
					t.Fatal("counter removal cost was not paid")
				}
			case cost.AdditionalReveal:
				if !g.Players[game.Player1].Hand.Contains(hand) || len(g.Events) == start {
					t.Fatal("reveal cost did not reveal without moving its card")
				}
			}
		})
	}
}

func TestPaymentCompletionJointSelectionsReserveOnlyConsumedResources(t *testing.T) {
	for _, consumes := range []cost.AdditionalKind{cost.AdditionalSacrifice, cost.AdditionalReturnToHand} {
		t.Run(fmt.Sprint(consumes), func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Reusable source", Types: []types.Card{types.Artifact}}})
			req := payment.ForAbility(payment.AbilityRequest{PlayerID: game.Player1, Source: source, AdditionalCosts: []cost.Additional{
				{Kind: consumes, Amount: 1}, {Kind: cost.AdditionalTapPermanents, Amount: 1},
			}})
			plan := requireCompletePayment(t, g, req, payment.Partial{})
			result := requireApplyPayment(t, g, plan)
			if !slices.Equal(result.Ability.TappedIDs, []id.ID{source.ObjectID}) || permanentForCard(g, source.CardInstanceID) != nil {
				t.Fatal("legal tap-then-zone-move reuse was not paid")
			}
		})
	}
	t.Run("fixed source sacrifice cannot be spent twice", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Fixed source"}})
		req := payment.ForAbility(payment.AbilityRequest{PlayerID: game.Player1, Source: source, AdditionalCosts: []cost.Additional{
			{Kind: cost.AdditionalSacrifice, Amount: 1}, {Kind: cost.AdditionalSacrificeSource, Amount: 1},
		}})
		if _, ok, _ := paymentOrch.planner(g).Complete(req, payment.Partial{}); ok {
			t.Fatal("one source was reserved for two sacrifices")
		}
		other := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Other sacrifice"}})
		plan := requireCompletePayment(t, g, req, payment.Partial{})
		if !slices.Equal(plan.Components()[0].Objects, []id.ID{other.ObjectID}) {
			t.Fatal("joint planning did not reserve the fixed source")
		}
		requireApplyPayment(t, g, plan)
	})
}
