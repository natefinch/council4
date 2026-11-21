package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
)

// Source review: pinned Oracle corpus SHA-256 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4,
// Unbounded Potential printing 732c3b63-5e15-4510-813c-58e20cd9833f.
// https://api.scryfall.com/cards/732c3b63-5e15-4510-813c-58e20cd9833f/rulings
// returned six rulings; in particular, entwined proliferate can choose creatures
// that received their first counter in the first mode.
// Read the November 6, 2026 Wizards Comprehensive Rules from a verified local
// copy of https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt
// (SHA-256 524f5195a1e5f54bb13c735edf399ac717a4458e2713f398d1a3e6877186b829).
// 115.1a/6, 601.2b/c/f-h, 700.2a/c and 702.42a-b govern optional targets,
// mode choice, the extra entwine payment, and printed resolution order;
// 608.2b-d governs target revalidation and resolution choices; 701.34a and
// 115.10a make proliferate an optional, nontargeted choice of recipients.
// 101.4 and 616.1 order replacement choices in a simultaneous event.
func TestUnboundedPotentialModesAndEntwine(t *testing.T) {
	registry := cards.NewDefaultRegistry()
	def, err := registry.LookupOracleID("d9ae08cf-eef0-407a-ad2c-08b4c89f99f0")
	if err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct {
		name        string
		modes       []int
		numTargets  int
		removeLast  bool
		counterless bool
		generic     int
		white       int
		reject      bool
		want        [2]int
		wantChoices int
		wantPoison  int
	}{
		{name: "counter mode puts one on each of two creatures", modes: []int{0}, numTargets: 2, generic: 1, white: 1, want: [2]int{1, 1}, wantPoison: 1},
		{name: "counter mode can choose no targets", modes: []int{0}, generic: 1, white: 1, wantPoison: 1},
		{name: "proliferate only does not target", modes: []int{1}, generic: 1, white: 1, wantChoices: 1, wantPoison: 2},
		{name: "entwine needs an additional white mana", modes: []int{0, 1}, numTargets: 1, generic: 5, white: 1, reject: true},
		{name: "entwine can proliferate a newly placed counter", modes: []int{0, 1}, numTargets: 1, counterless: true, generic: 4, white: 2, want: [2]int{2, 0}, wantChoices: 1},
		{name: "entwine with zero targets still proliferates", modes: []int{0, 1}, generic: 4, white: 2, wantChoices: 1, wantPoison: 2},
		{name: "one illegal target does not stop the entwined spell", modes: []int{0, 1}, numTargets: 2, removeLast: true, generic: 4, white: 2, want: [2]int{2, 0}, wantChoices: 1, wantPoison: 1},
		{name: "all announced targets illegal stops even proliferate", modes: []int{0, 1}, numTargets: 1, removeLast: true, generic: 4, white: 2, wantPoison: 1},
	} {
		t.Run(tt.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			first := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
			second := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
			permanents := []*game.Permanent{first, second}
			if !tt.counterless {
				g.Players[game.Player2].PoisonCounters = 1
			}
			spell := addCardToHand(g, game.Player1, def)
			setMainPhasePriority(g, game.Player1)
			g.Players[game.Player1].ManaPool.Add(mana.C, tt.generic)
			g.Players[game.Player1].ManaPool.Add(mana.W, tt.white)
			e := NewEngineWithPrograms(nil, registry.Programs())
			var targets []game.Target
			for _, permanent := range permanents[:tt.numTargets] {
				targets = append(targets, game.PermanentTarget(permanent.ObjectID))
			}
			cast := e.applyAction(g, game.Player1, action.CastSpell(spell, targets, 0, tt.modes))
			if tt.reject {
				if cast || g.Players[game.Player1].ManaPool.Total() != tt.generic+tt.white || !g.Players[game.Player1].Hand.Contains(spell) {
					t.Fatal("unpayable entwine should not cast or spend mana")
				}
				return
			}
			if !cast || g.Players[game.Player1].ManaPool.Total() != 0 {
				t.Fatal("spell did not cast for the expected cost")
			}
			if tt.removeLast {
				last := permanents[tt.numTargets-1]
				if _, ok := removePermanentFromBattlefield(g, last.ObjectID); !ok {
					t.Fatal("could not remove the last target")
				}
			}
			log := &TurnLog{}
			e.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{
				game.Player1: &choiceOnlyAgent{choices: [][]int{{0}}},
			}, log)
			if len(log.Choices) != tt.wantChoices || tt.wantChoices != 0 &&
				(log.Choices[0].Request.Kind != game.ChoiceProliferate || log.Choices[0].UsedFallback) {
				t.Fatalf("resolution choices = %+v, want %d successful proliferate choices", log.Choices, tt.wantChoices)
			}
			got := [2]int{first.Counters.Get(counter.PlusOnePlusOne), second.Counters.Get(counter.PlusOnePlusOne)}
			if got != tt.want || g.Players[game.Player2].PoisonCounters != tt.wantPoison {
				t.Fatalf("counters = %v, poison = %d; want %v and %d", got, g.Players[game.Player2].PoisonCounters, tt.want, tt.wantPoison)
			}
			if tt.numTargets == 2 && !tt.removeLast {
				var counters []game.Event
				for _, event := range g.Events {
					if event.Kind == game.EventCountersAdded {
						counters = append(counters, event)
					}
				}
				if len(counters) != 2 || counters[0].SimultaneousID == 0 || counters[0].SimultaneousID != counters[1].SimultaneousID {
					t.Fatalf("expected one simultaneous placement, got %+v", counters)
				}
			}
		})
	}
}
