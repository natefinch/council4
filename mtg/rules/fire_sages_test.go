package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/cards/authored"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
)

// Source review: pinned corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4 and
// the live Scryfall card response retrieved 2026-11-22 agree on Fire Sages,
// Oracle ID 08d1940a-ddc6-4b6d-b6c9-9d2aa470b906, printing revision
// 71a9fa44-306f-417a-a9d9-991aff95025c, {1}{R}, Creature — Human Cleric,
// 2/2, and the full text: “Firebending 1 (Whenever this creature attacks,
// add {R}. This mana lasts until end of combat.) {1}{R}{R}: Put a +1/+1
// counter on this creature.” Its rulings URI,
// https://api.scryfall.com/cards/71a9fa44-306f-417a-a9d9-991aff95025c/rulings,
// returned all five WotC rulings on 2026-11-22, dated 2025-10-02: Firebending
// triggers separately for each instance, uses the stack, and retains its red
// mana through combat steps (including after damage) but not after combat;
// other combat mana still empties normally. Official WotC CR TXT effective
// 2026-11-06, https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt,
// read 106.3–4, 118.2–3, 122.1a, 500.4–5a, 508.1m–508.2a, 602.1–2,
// 603.1–3, 605.1a–b, and 702.189a: attacking triggers a normal stack ability,
// Firebending is not a mana ability, its red mana survives step boundaries
// until combat ends, and paying {1}{R}{R} adds a +1/+1 counter that increases
// the creature's power and toughness.
func TestFireSagesAttackSpendsFirebendingManaForCounter(t *testing.T) {
	const oracleID = "08d1940a-ddc6-4b6d-b6c9-9d2aa470b906"
	if authored.FireSages.OracleID != oracleID {
		t.Fatalf("Fire Sages Oracle ID = %q, want %q", authored.FireSages.OracleID, oracleID)
	}

	registry := cards.NewDefaultRegistry()
	def, err := registry.LookupOracleID(oracleID)
	if err != nil || def == nil || def.Name != "Fire Sages" {
		t.Fatalf("checked Fire Sages identity lookup = %v, %v", def, err)
	}
	engine := NewEngineWithPrograms(nil, registry.Programs())
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	setMainPhasePriority(g, game.Player1)

	sage := addCombatPermanent(g, game.Player1, def)
	for range 2 {
		addBasicLandPermanent(g, game.Player1, types.Mountain)
	}

	attack := attackPlayer2(sage)
	probe := &combatManaProbe{
		attack:      action.DeclareAttackers(attack.Attackers),
		paymentStep: game.StepEndOfCombat,
		red:         make(map[game.Step]int),
		payment: func(candidate action.Action) bool {
			activation, ok := candidate.ActivateAbilityPayload()
			return ok && activation.SourceID == sage.ObjectID
		},
	}
	policies := simPassPolicies()
	policies[game.Player1] = probe
	spentBeforeActivation := g.Players[game.Player1].ManaPool.Spent()
	engine.runCombatPhase(g, policies, &TurnLog{})

	if !probe.triggerOnStack {
		t.Fatal("Firebending did not put a triggered ability on the stack when Fire Sages attacked")
	}
	if got, seen := probe.red[game.StepEndOfCombat]; !seen || got != 1 {
		t.Fatalf("red mana from Firebending at end-of-combat priority = %d (seen %v), want 1", got, seen)
	}
	if !probe.paid {
		t.Fatal("Fire Sages' {1}{R}{R} ability did not activate during end-of-combat priority")
	}
	if got := g.Players[game.Player1].ManaPool.Spent() - spentBeforeActivation; got != 3 {
		t.Fatalf("mana spent on +1/+1 counter activation = %d, want 3", got)
	}
	if got := sage.Counters.Get(counter.PlusOnePlusOne); got != 1 {
		t.Fatalf("Fire Sages has %d +1/+1 counters, want 1", got)
	}
	if got := effectivePower(g, sage); got != 3 {
		t.Fatalf("Fire Sages power after counter = %d, want 3", got)
	}
	if got := effectiveToughnessMust(t, g, sage); got != 3 {
		t.Fatalf("Fire Sages toughness after counter = %d, want 3", got)
	}
}
