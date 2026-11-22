package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/cards/authored"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// Source review: pinned corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4 and
// current Scryfall card response retrieved 2026-11-22 agree on Loyal Fire Sage,
// Oracle ID 0c052420-2d8b-4bbf-8197-b8a36e7208c9, printing revision
// 1126b572-1cac-44aa-8fec-e2c72a9fea47, {2}{R}, Creature — Human Cleric Ally,
// 3/3, and the full text: “Firebending 1 (Whenever this creature attacks, add
// {R}. This mana lasts until end of combat.) {5}: Create a 1/1 white Ally
// creature token.” Its rulings URI,
// https://api.scryfall.com/cards/1126b572-1cac-44aa-8fec-e2c72a9fea47/rulings,
// returned all five WotC rulings on 2026-11-22, dated 2025-10-02: Firebending
// uses the stack, triggers on attack, and retains only its own mana through
// combat step boundaries and after damage until leaving combat; separate mana
// empties normally, and multiple instances trigger separately. Official WotC
// CR TXT effective 2026-11-06,
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt,
// read 106.4, 500.5/500.5a, 508.3a, 511.1–3, 602.2, 603.1–3, 605.1b,
// 702.189a and 111.3–4: attack declaration creates a normal stack trigger,
// the {5} activation pays its cost, Firebending mana remains available through
// end-of-combat priority then empties, and the created token is a 1/1 white Ally.
func TestLoyalFireSageAttackSpendsFirebendingManaForAlly(t *testing.T) {
	const oracleID = "0c052420-2d8b-4bbf-8197-b8a36e7208c9"
	if authored.LoyalFireSage.OracleID != oracleID {
		t.Fatalf("Loyal Fire Sage Oracle ID = %q, want %q", authored.LoyalFireSage.OracleID, oracleID)
	}

	registry := cards.NewDefaultRegistry()
	def, err := registry.LookupOracleID(oracleID)
	if err != nil || def == nil || def.Name != "Loyal Fire Sage" {
		t.Fatalf("checked Loyal Fire Sage identity lookup = %v, %v", def, err)
	}
	engine := NewEngineWithPrograms(nil, registry.Programs())
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	setMainPhasePriority(g, game.Player1)

	cardID := addCardToHand(g, game.Player1, def)
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)
	g.Players[game.Player1].ManaPool.Add(mana.R, 1)
	spentBeforeCast := g.Players[game.Player1].ManaPool.Spent()
	if !engine.applyAction(g, game.Player1, action.CastSpell(cardID, nil, 0, nil)) {
		t.Fatal("casting Loyal Fire Sage for {2}{R} failed")
	}
	engine.resolveTopOfStackWithChoices(g, simPassPolicies(), &TurnLog{})
	sage := permanentForCard(g, cardID)
	if sage == nil || g.Players[game.Player1].ManaPool.Spent()-spentBeforeCast != 3 {
		t.Fatal("Loyal Fire Sage did not resolve after paying {2}{R}")
	}
	sage.SummoningSick = false

	for range 4 {
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
		t.Fatal("Firebending did not put a triggered ability on the stack when Loyal Fire Sage attacked")
	}
	if got, seen := probe.red[game.StepEndOfCombat]; !seen || got != 1 {
		t.Fatalf("red mana at end-of-combat priority = %d (seen %v), want 1", got, seen)
	}
	if !probe.paid {
		t.Fatal("Loyal Fire Sage's {5} ability did not activate during end-of-combat priority")
	}
	if got := g.Players[game.Player1].ManaPool.Spent() - spentBeforeActivation; got != 5 {
		t.Fatalf("mana spent on Ally activation = %d, want {5}", got)
	}

	if got := countTokenNamed(g, "Ally", game.Player1); got != 1 {
		t.Fatalf("Loyal Fire Sage created %d Ally tokens, want 1", got)
	}
	for _, permanent := range g.Battlefield {
		if !permanent.Token || permanentEffectiveName(g, permanent) != "Ally" {
			continue
		}
		if effectivePower(g, permanent) != 1 || effectiveToughnessMust(t, g, permanent) != 1 ||
			!permanentHasType(g, permanent, types.Creature) ||
			!permanentHasSubtype(g, permanent, types.Sub("Ally")) ||
			!hasEffectiveColor(g, permanent, color.White) {
			t.Fatal("Loyal Fire Sage's token is not a 1/1 white Ally creature")
		}
	}
	if got := g.Players[game.Player1].ManaPool.Total(); got != 0 {
		t.Fatalf("mana remaining after end of combat = %d, want 0", got)
	}
}
