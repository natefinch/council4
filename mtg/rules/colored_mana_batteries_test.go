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

// Source review: Red Mana Battery, 4ED #343, pinned printing revision
// e4e507bc-441d-4f44-85d4-cf93ca199d2e (Oracle ID
// 8ceb295c-a611-4ac7-940b-557bf60f8a8b); exact Oracle text is retained in the
// pinned corpus and .cardwork/colored-mana-batteries/red-card.json. Its
// https://api.scryfall.com/cards/e4e507bc-441d-4f44-85d4-cf93ca199d2e/rulings
// response in .cardwork/colored-mana-batteries/red-rulings.json has one WotC
// ruling (2004-10-04): "Can be tapped even if it has no counters."
//
// Read the official Comprehensive Rules effective 2026-09-25, matching this
// workflow's rules_effective pin, at
// https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.pdf;
// the PDF is retained at
// .cardwork/colored-mana-batteries/comprehensive-rules-2026-09-25.pdf, and
// text extracted from the official DOCX is retained at
// .cardwork/colored-mana-batteries/comprehensive-rules-2026-09-25.txt (source
// DOCX: .cardwork/colored-mana-batteries/comprehensive-rules-2026-09-25.docx).
// CR 106.3-4 governs adding mana to the player's pool; 107.1c makes "any
// number" include zero; 118.1-3 and 602.2a-b govern paying the activation costs;
// 122.1 defines counters; 605.1a classifies the untargeted ability that adds
// mana; and 605.3a-b permits its timing and immediate resolution. Thus removing
// N chosen charge counters as a cost adds one mana of the printed color plus N
// more, while the two-mana ability resolves to add one charge counter.
//
// Source review: Green Mana Battery, 4ED #323, pinned printing revision
// d0a6e224-72df-4f1f-93d1-5114779aed2c (Oracle ID
// b953d19f-6ba7-4aee-b353-c60dc9572610); exact Oracle text and the live
// Scryfall card response are retained in the pinned corpus and
// .cardwork/colored-mana-batteries/green-card.json. Its
// https://api.scryfall.com/cards/d0a6e224-72df-4f1f-93d1-5114779aed2c/rulings
// response in .cardwork/colored-mana-batteries/green-rulings.json has one WotC
// ruling (2004-10-04): "Can be tapped even if it has no counters."
// The task header revision d0a6e224-72df-4f44-85d4-cf93ca199d2e differs from
// the saved queue and pinned corpus ID above; the exact-name 4ED lookup
// confirmed the saved queue's Oracle ID, text, and printing.
//
// Read the official Comprehensive Rules effective 2026-09-25, matching this
// workflow's rules_effective pin, at
// https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.pdf;
// the PDF is retained at
// .cardwork/colored-mana-batteries/comprehensive-rules-2026-09-25.pdf, and
// text extracted from the official DOCX is retained at
// .cardwork/colored-mana-batteries/comprehensive-rules-2026-09-25.txt (source
// DOCX: .cardwork/colored-mana-batteries/comprehensive-rules-2026-09-25.docx).
// CR 106.3-4 governs adding mana to the player's pool; 107.1c makes "any
// number" include zero; 118.1-3 and 602.2a-b govern paying the activation costs;
// 122.1 defines counters; 605.1a classifies the untargeted ability that adds
// mana; and 605.3a-b permits its timing and immediate resolution. Thus removing
// N chosen charge counters as a cost adds one green mana plus N more, while the
// two-mana ability resolves to add one charge counter.
//
// Source review: White Mana Battery, 4ED #357, pinned printing revision
// b622e694-858d-482f-a67d-0e52d268708c (Oracle ID
// cf58682f-c305-4803-b54b-37f0841788e9); exact Oracle text and the live
// Scryfall card response are retained in the pinned corpus and
// .cardwork/colored-mana-batteries/white-card.json. Its
// https://api.scryfall.com/cards/b622e694-858d-482f-a67d-0e52d268708c/rulings
// response in .cardwork/colored-mana-batteries/white-rulings.json has one WotC
// ruling (2004-10-04): "Can be tapped even if it has no counters."
//
// Read the official Comprehensive Rules effective 2026-09-25, matching this
// workflow's rules_effective pin, at
// https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.pdf;
// the PDF is retained at
// .cardwork/colored-mana-batteries/comprehensive-rules-2026-09-25.pdf, and
// text extracted from the official DOCX is retained at
// .cardwork/colored-mana-batteries/comprehensive-rules-2026-09-25.txt (source
// DOCX: .cardwork/colored-mana-batteries/comprehensive-rules-2026-09-25.docx).
// CR 106.3-4 governs adding mana to the player's pool; 107.1c makes "any
// number" include zero; 118.1-3 and 602.2a-b govern paying the activation costs;
// 122.1 defines counters; 605.1a classifies the untargeted ability that adds
// mana; and 605.3a-b permits its timing and immediate resolution. Thus removing
// N chosen charge counters as a cost adds one white mana plus N more, while the
// two-mana ability resolves to add one charge counter.
//
// Source review: Black Mana Battery, 4ED #298, pinned printing revision
// a81f7a0f-183f-438b-b252-738d8d30c245 (Oracle ID
// dc3b2bab-755d-4b2f-97c0-70a74d721a79); exact Oracle text and the live
// Scryfall card response are retained in the pinned corpus and
// .cardwork/colored-mana-batteries/black-card.json. Its
// https://api.scryfall.com/cards/a81f7a0f-183f-438b-b252-738d8d30c245/rulings
// response in .cardwork/colored-mana-batteries/black-rulings.json has one WotC
// ruling (2004-10-04): "Can be tapped even if it has no counters."
//
// Read the official Comprehensive Rules effective 2026-09-25, matching this
// workflow's rules_effective pin, at
// https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.pdf;
// the PDF is retained at
// .cardwork/colored-mana-batteries/comprehensive-rules-2026-09-25.pdf, and
// text extracted from the official DOCX is retained at
// .cardwork/colored-mana-batteries/comprehensive-rules-2026-09-25.txt (source
// DOCX: .cardwork/colored-mana-batteries/comprehensive-rules-2026-09-25.docx).
// CR 106.3-4 governs adding mana to the player's pool; 107.1c makes "any
// number" include zero; 118.1-3 and 602.2a-b govern paying the activation costs;
// 122.1 defines counters; 605.1a classifies the untargeted ability that adds
// mana; and 605.3a-b permits its timing and immediate resolution. Thus removing
// N chosen charge counters as a cost adds one black mana plus N more, while the
// two-mana ability resolves to add one charge counter.
const (
	redManaBatteryTestOracleID   = "8ceb295c-a611-4ac7-940b-557bf60f8a8b"
	greenManaBatteryTestOracleID = "b953d19f-6ba7-4aee-b353-c60dc9572610"
	whiteManaBatteryTestOracleID = "cf58682f-c305-4803-b54b-37f0841788e9"
	blackManaBatteryTestOracleID = "dc3b2bab-755d-4b2f-97c0-70a74d721a79"
)

func manaBatteryTestDefinition(t *testing.T, oracleID, name string) *game.CardDef {
	t.Helper()
	def, err := cards.NewDefaultRegistry().LookupOracleID(oracleID)
	if err != nil {
		t.Fatalf("exact %s default-registry lookup failed: %v", name, err)
	}
	if def.Name != name {
		t.Fatalf("exact %s lookup returned %q", name, def.Name)
	}
	return def
}

func runManaBatteryScenario(t *testing.T, oracleID, name string, manaColor mana.Color) {
	t.Helper()
	def := manaBatteryTestDefinition(t, oracleID, name)
	engine := NewEngine(nil)

	g := mainPhaseGame(game.Player1)
	setMainPhasePriority(g, game.Player1)
	source := addCombatPermanent(g, game.Player1, def)
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)

	chargeActivation := action.ActivateAbility(source.ObjectID, def.ActivatedAbilityIndex(0), nil, 0)
	if !actionsContain(engine.legalActions(g, game.Player1), chargeActivation) ||
		!engine.applyAction(g, game.Player1, chargeActivation) {
		t.Fatal("two-mana tap ability was not legally activated")
	}
	if !source.Tapped || g.Players[game.Player1].ManaPool.Total() != 0 ||
		source.Counters.Get(counter.Charge) != 0 || g.Stack.Size() != 1 {
		t.Fatal("activation did not pay {2}, tap the artifact, and put the ability on the stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if source.Counters.Get(counter.Charge) != 1 {
		t.Fatalf("charge ability left %d counters, want 1", source.Counters.Get(counter.Charge))
	}

	// A fresh game represents a later untap and starts with the counter the
	// first ability creates; the second ability itself still pays its real costs.
	g = mainPhaseGame(game.Player1)
	setMainPhasePriority(g, game.Player1)
	source = addCombatPermanent(g, game.Player1, def)
	source.Counters.Add(counter.Charge, 1)
	manaActivation := action.ActivateAbility(source.ObjectID, def.ManaAbilityIndex(0), nil, 0)
	if !actionsContain(engine.legalActions(g, game.Player1), manaActivation) {
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
		if !slices.Equal(labels, []string{"0", "1"}) {
			t.Fatalf("charge-counter options %v, want [0 1]", labels)
		}
		return []int{1}
	}}
	if !engine.applyActionWithChoices(g, game.Player1, manaActivation, reboundKickerAgents(agent), &TurnLog{}) {
		t.Fatal("charge-counter mana ability activation rejected")
	}
	if !prompted || !source.Tapped || source.Counters.Get(counter.Charge) != 0 ||
		g.Players[game.Player1].ManaPool.Amount(manaColor) != 2 ||
		g.Players[game.Player1].ManaPool.Total() != 2 || g.Stack.Size() != 0 {
		t.Fatalf("mana ability left prompted=%v, tapped=%v, counters=%d, color mana=%d, total=%d, stack=%d; want true, true, 0, 2, 2, 0",
			prompted, source.Tapped, source.Counters.Get(counter.Charge),
			g.Players[game.Player1].ManaPool.Amount(manaColor),
			g.Players[game.Player1].ManaPool.Total(), g.Stack.Size())
	}
}

func TestRedManaBatteryStoresAndConvertsChargeCounters(t *testing.T) {
	runManaBatteryScenario(t, redManaBatteryTestOracleID, "Red Mana Battery", mana.R)
}

func TestGreenManaBatteryStoresAndConvertsChargeCounters(t *testing.T) {
	runManaBatteryScenario(t, greenManaBatteryTestOracleID, "Green Mana Battery", mana.G)
}

func TestWhiteManaBatteryStoresAndConvertsChargeCounters(t *testing.T) {
	runManaBatteryScenario(t, whiteManaBatteryTestOracleID, "White Mana Battery", mana.W)
}

func TestBlackManaBatteryStoresAndConvertsChargeCounters(t *testing.T) {
	runManaBatteryScenario(t, blackManaBatteryTestOracleID, "Black Mana Battery", mana.B)
}
