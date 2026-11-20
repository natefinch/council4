package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Source review: pinned Oracle corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4;
// Unbounded Potential printing revision 732c3b63-5e15-4510-813c-58e20cd9833f,
// Oracle ID d9ae08cf-eef0-407a-ad2c-08b4c89f99f0, Instant {1}{W}, exact text:
// "Choose one —
// • Put a +1/+1 counter on each of up to two target creatures.
// • Proliferate. (Choose any number of permanents and/or players, then give
// each another counter of each kind already there.)
// Entwine {3}{W} (Choose both if you pay the entwine cost.)"
// Scryfall rulings URI https://api.scryfall.com/cards/732c3b63-5e15-4510-813c-58e20cd9833f/rulings
// retrieved 2026-11-20 returned six WotC rulings dated 2021-06-18: the earlier
// entwined mode can supply eligible counters; only battlefield permanents
// and players with counters qualify, including opponents; zero is legal;
// every existing kind is added; no priority interrupts resolution; opposite
// +1/+1 and -1/-1 counters cancel as a state-based action.
// Official Wizards Comprehensive Rules effective 2026-11-06, verified through
// https://magic.wizards.com/en/rules and read from
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt
// (SHA-256 524f5195a1e5f54bb13c735edf399ac717a4458e2713f398d1a3e6877186b829):
// 601.2b/c/f-h and 700.2a/c/h govern modal choice, chosen-mode targets and
// additional-cost payment; 701.34a requires any-number selection and one
// additional counter of each existing kind; 115.10a and 608.2c-d make it a
// nontarget choice during ordered resolution; 702.26b excludes phased-out
// permanents; 702.42a-b allows all modes for the entwine cost and resolves
// them in printed order. 101.4 and 616.1 order simultaneous replacement
// choices by affected seat. These are shared-operation probes, not a registration.
func TestUnboundedPotentialProliferateRulesProbe(t *testing.T) {
	tests := []struct {
		name         string
		selection    []int
		wantDelta    int
		wantFallback bool
	}{
		{name: "choosing no permanents or players leaves counters unchanged"},
		{name: "mixed subset gets every kind on both recipients", selection: []int{3, 0}, wantDelta: 1},
		{name: "duplicate recipient is not a valid selection", selection: []int{0, 0}, wantFallback: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			chosen := addCombatCreaturePermanentWithPower(g, game.Player2, 2, game.Hexproof)
			chosen.Counters.Add(counter.PlusOnePlusOne, 1)
			chosen.Counters.Add(counter.Charge, 2)
			skipped := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
			skipped.Counters.Add(counter.Charge, 1)
			counterless := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
			phasedOut := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
			phasedOut.Counters.Add(counter.Charge, 1)
			phasedOut.PhasedOut = true
			g.Players[game.Player1].ExperienceCounters = 1
			opponent := g.Players[game.Player2]
			opponent.PoisonCounters, opponent.EnergyCounters, opponent.ExperienceCounters = 1, 2, 3
			g.Players[game.Player3].Eliminated = true
			g.Players[game.Player3].PoisonCounters = 1
			addEffectSpellToStack(g, game.Player1, game.Proliferate{}, nil)

			log := &TurnLog{}
			NewEngine(nil).resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{
				game.Player1: &choiceOnlyAgent{choices: [][]int{tt.selection}},
			}, log)

			if len(log.Choices) != 1 {
				t.Fatalf("choices = %v, want one decision", log.Choices)
			}
			decision := log.Choices[0]
			request := decision.Request
			if request.Kind != game.ChoiceProliferate || request.Player != game.Player1 ||
				request.MinChoices != 0 || request.MaxChoices != 4 || len(request.Options) != 4 {
				t.Fatalf("request = %+v, want one optional mixed choice for the resolving controller", request)
			}
			if info := request.Options[0].Card; !info.Exists || info.Val.CardID != chosen.CardInstanceID {
				t.Fatalf("first option = %+v, want the opponent's hexproof permanent", request.Options[0])
			}
			for _, option := range request.Options {
				if len(option.Targets) != 0 {
					t.Fatalf("proliferate should not target: %+v", option)
				}
			}
			wantSelection := tt.selection
			if tt.wantFallback {
				wantSelection = nil
			}
			if decision.UsedFallback != tt.wantFallback || !slices.Equal(decision.Selected, wantSelection) {
				t.Fatalf("choice = %+v, want selection %v, fallback %v", decision, wantSelection, tt.wantFallback)
			}

			permanentCounts := [2]int{chosen.Counters.Get(counter.PlusOnePlusOne), chosen.Counters.Get(counter.Charge)}
			if want := [2]int{1 + tt.wantDelta, 2 + tt.wantDelta}; permanentCounts != want {
				t.Errorf("selected permanent's +1/+1 and charge counts = %v, want %v", permanentCounts, want)
			}
			playerCounts := [3]int{opponent.PoisonCounters, opponent.EnergyCounters, opponent.ExperienceCounters}
			if want := [3]int{1 + tt.wantDelta, 2 + tt.wantDelta, 3 + tt.wantDelta}; playerCounts != want {
				t.Errorf("selected player's poison, energy and experience counts = %v, want %v", playerCounts, want)
			}
			if skipped.Counters.Get(counter.Charge) != 1 || !counterless.Counters.IsEmpty() ||
				phasedOut.Counters.Get(counter.Charge) != 1 || g.Players[game.Player1].ExperienceCounters != 1 ||
				g.Players[game.Player3].PoisonCounters != 1 {
				t.Error("proliferate changed counters on an unselected or ineligible recipient")
			}

			var events []game.Event
			for _, event := range g.Events {
				if event.Kind == game.EventCountersAdded {
					events = append(events, event)
				}
			}
			if len(events) != 5*tt.wantDelta {
				t.Fatalf("placement events = %v, want %d", events, 5*tt.wantDelta)
			}
			for _, event := range events {
				if event.SimultaneousID == 0 || event.SimultaneousID != events[0].SimultaneousID ||
					event.Controller != game.Player1 || event.Amount != 1 {
					t.Errorf("placement event = %+v, want one simultaneous placement by Player1", event)
				}
				if event.CounterKind == counter.Poison &&
					(event.Player != game.Player2 || event.PreviousCounterAmount != 1) {
					t.Errorf("poison event = %+v, want Player2 with one previous counter", event)
				}
			}
		})
	}
}

func TestProliferateSeesCountersFromEarlierPaidMode(t *testing.T) {
	// A test-local modal spell exercises the authored Execute seam without
	// adding Unbounded Potential to any registry.
	add := game.ProgramReference{ID: "test/proliferate/add", Version: "1"}
	proliferate := game.ProgramReference{ID: "test/proliferate/choose", Version: "1"}
	catalog, err := cardprogram.NewCatalog([]cardprogram.Program{
		{Reference: add, Resolve: func(ctx cardprogram.Context) error {
			_, err := ctx.Execute(game.AddCounter{
				Amount: game.Fixed(1), CounterKind: counter.PlusOnePlusOne, Object: game.TargetPermanentReference(0),
			})
			return err
		}},
		{Reference: proliferate, Resolve: func(ctx cardprogram.Context) error {
			_, err := ctx.Execute(game.Proliferate{})
			return err
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	def := &game.CardDef{CardFace: game.CardFace{
		Name: "Paid Proliferate Fixture", Types: []types.Card{types.Instant},
		ManaCost: opt.Val(cost.Mana{cost.O(1), cost.W}),
		SpellAbility: opt.Val(game.AbilityContent{
			MinModes: 1, MaxModes: 2, EscalateCost: opt.Val(cost.Mana{cost.O(3), cost.W}),
			Modes: []game.Mode{
				{Program: add, Targets: []game.TargetSpec{{
					Allow: game.TargetAllowPermanent, MinTargets: 1, MaxTargets: 1,
					Selection: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
				}}},
				{Program: proliferate},
			},
		}),
	}}
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	spell := addCardToHand(g, game.Player1, def)
	setMainPhasePriority(g, game.Player1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 4)
	g.Players[game.Player1].ManaPool.Add(mana.W, 2)
	e := NewEngineWithPrograms(nil, catalog)

	if !e.applyAction(g, game.Player1, action.CastSpell(spell,
		[]game.Target{game.PermanentTarget(creature.ObjectID)}, 0, []int{0, 1})) {
		t.Fatal("could not cast both modes")
	}
	if g.Players[game.Player1].ManaPool.Total() != 0 || !creature.Counters.IsEmpty() {
		t.Fatal("casting should pay the full cost but not place any counters yet")
	}
	log := &TurnLog{}
	e.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{0}}},
	}, log)
	if len(log.Choices) != 1 || log.Choices[0].Request.Kind != game.ChoiceProliferate ||
		len(log.Choices[0].Request.Options) != 1 || !log.Choices[0].Request.Options[0].Card.Exists ||
		log.Choices[0].Request.Options[0].Card.Val.CardID != creature.CardInstanceID || log.Choices[0].UsedFallback {
		t.Fatalf("choices = %+v, want the creature enabled by the first mode", log.Choices)
	}
	if got := creature.Counters.Get(counter.PlusOnePlusOne); got != 2 {
		t.Fatalf("+1/+1 counters = %d, want 2 after both modes", got)
	}
}
