package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
)

// Source review: 4ED printing cd2cb84e-c079-486e-87ad-d188fe38bc76, Oracle ID
// 79e91e8a-cee7-411b-b0a3-d3383c0d4a43, current Oracle text and printing
// revision verified in the pinned corpus and retained Scryfall card response
// at .cardwork/blue-mana-battery/card.json. The printing's
// https://api.scryfall.com/cards/cd2cb84e-c079-486e-87ad-d188fe38bc76/rulings
// response is retained at .cardwork/blue-mana-battery/rulings.json; its one
// WotC ruling (2004-10-04) confirms the artifact can be tapped with no counters.
//
// Read the official Comprehensive Rules effective 2026-09-25 at
// https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.docx;
// the verified local copy is retained at
// .cardwork/blue-mana-battery/comprehensive-rules-2026-09-25.docx (text
// extraction: .txt). CR 106.3-4 requires produced mana to enter its controller's
// mana pool; 107.1c makes "any number" include zero; 118.1-3 and 602.2a-b govern
// paying the two activated costs; 605.1a classifies the untargeted mana ability,
// and 605.3a-b permits its timing and immediate resolution. The second ability
// must let its controller choose 0..N charge counters as a cost, then add one
// blue plus one for each actually removed. The ruling confirms tapping remains
// legal when the chosen number is zero.
const blueManaBatteryTestOracleID = "79e91e8a-cee7-411b-b0a3-d3383c0d4a43"

func blueManaBatteryTestDefinition(t *testing.T) *game.CardDef {
	t.Helper()
	def, err := cards.NewDefaultRegistry().LookupOracleID(blueManaBatteryTestOracleID)
	if err != nil {
		t.Fatalf("exact Blue Mana Battery default-registry lookup failed: %v", err)
	}
	if def.Name != "Blue Mana Battery" {
		t.Fatalf("exact Blue Mana Battery lookup returned %q", def.Name)
	}
	return def
}

func TestBlueManaBatteryAddsChargeCounter(t *testing.T) {
	def := blueManaBatteryTestDefinition(t)
	g := mainPhaseGame(game.Player1)
	setMainPhasePriority(g, game.Player1)
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, def)
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)

	activation := action.ActivateAbility(source.ObjectID, def.ActivatedAbilityIndex(0), nil, 0)
	if !actionsContain(engine.legalActions(g, game.Player1), activation) ||
		!engine.applyAction(g, game.Player1, activation) {
		t.Fatal("two-mana tap ability was not legally activated")
	}
	if !source.Tapped || g.Players[game.Player1].ManaPool.Total() != 0 ||
		source.Counters.Get(counter.Charge) != 0 || g.Stack.Size() != 1 {
		t.Fatal("activation did not pay {2}, tap the artifact, and put the ability on the stack")
	}

	engine.resolveTopOfStack(g, &TurnLog{})
	if source.Counters.Get(counter.Charge) != 1 {
		t.Fatalf("resolution left %d charge counters, want 1", source.Counters.Get(counter.Charge))
	}
}

func TestBlueManaBatteryProducesBlueManaFromChosenChargeCounters(t *testing.T) {
	def := blueManaBatteryTestDefinition(t)
	g := mainPhaseGame(game.Player1)
	setMainPhasePriority(g, game.Player1)
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, def)
	source.Counters.Add(counter.Charge, 2)
	activation := action.ActivateAbility(source.ObjectID, def.ManaAbilityIndex(0), nil, 0)
	if !actionsContain(engine.legalActions(g, game.Player1), activation) {
		t.Fatal("charge-counter mana ability was not offered")
	}

	prompted := false
	agent := &authoredCopyChoiceAgent{choose: func(req game.ChoiceRequest) []int {
		if req.Kind != game.ChoicePayment {
			return req.DefaultSelection
		}
		prompted = true
		labels := make([]string, len(req.Options))
		for i, option := range req.Options {
			labels[i] = option.Label
		}
		if !slices.Equal(labels, []string{"0", "1", "2"}) {
			t.Fatalf("counter-count options %v, want [0 1 2]", labels)
		}
		return []int{1}
	}}
	if !engine.applyActionWithChoices(g, game.Player1, activation, reboundKickerAgents(agent), &TurnLog{}) {
		t.Fatal("charge-counter mana ability activation rejected")
	}
	if !prompted || !source.Tapped || source.Counters.Get(counter.Charge) != 1 ||
		g.Players[game.Player1].ManaPool.Amount(mana.U) != 2 || g.Stack.Size() != 0 {
		t.Fatalf("choosing one of two left prompted=%v, tapped=%v, counters=%d, blue=%d, stack=%d; want true, true, 1, 2, 0",
			prompted, source.Tapped, source.Counters.Get(counter.Charge),
			g.Players[game.Player1].ManaPool.Amount(mana.U), g.Stack.Size())
	}
}
