package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
)

// Source review: exact Oracle ID 372094f8-86dc-4c78-a5d0-444ffbf02630,
// Oath of the Gatewatch #111, printing revision
// 5082e0a1-5c3e-4b8c-ba88-290151564133. The pinned corpus
// (SHA-256 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4)
// and current Scryfall record retrieved 2026-11-21 agree: {3}{R} Creature —
// Goblin Warrior Ally, 3/2; Oracle text: "Surge {1}{R} (You may cast this
// spell for its surge cost if you or a teammate has cast another spell this
// turn.) Menace (This creature can't be blocked except by two or more
// creatures.)" The current record and full rulings response are retained in
// .cardwork/sources/goblin-freerunner-card.json and
// .cardwork/goblin-freerunner-rulings.json. Retrieved from
// https://api.scryfall.com/cards/5082e0a1-5c3e-4b8c-ba88-290151564133 and
// https://api.scryfall.com/cards/5082e0a1-5c3e-4b8c-ba88-290151564133/rulings;
// all four WotC rulings are dated 2016-01-22. They cover Surge as an
// alternative cost, qualifying earlier casts (including resolved/countered
// spells and spells on the stack), copy payment for instants/sorceries, and
// unchanged mana value.
// Read the corpus-pinned official CR effective 2026-09-25 (SHA-256
// 8d860e451f20f38865b725b42d82feb714c725373dd8f3b32b8652b3eeb070ca) and
// current official CR effective 2026-11-06 (SHA-256
// 524f5195a1e5f54bb13c735edf399ac717a4458e2713f398d1a3e6877186b829), both
// from https://magic.wizards.com/en/rules:
// https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.txt
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt.
// Read 102.3-4 (team and free-for-all meaning), 118.9-9d and 601.2b/f-i
// (optional alternative-cost announcement and payment), 500.1 (turn scope),
// 514.2 (end of "this turn"), 202.3 (mana value remains based on mana cost),
// 702.117a (Surge gate), and 702.111a-c (Menace requires at least two blockers).
func TestGoblinFreerunnerSurgeCastHasMenace(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	setMainPhasePriority(g, game.Player1)

	setupSpell := castSurgeSetupSpell(t, engine, g, game.Player1)
	engine.resolveTopOfStack(g, &TurnLog{})
	if !g.Players[game.Player1].Graveyard.Contains(setupSpell) {
		t.Fatal("the qualifying spell did not resolve before Goblin Freerunner")
	}

	registry := cards.NewDefaultRegistry()
	freerunnerDef, err := registry.LookupOracleID("372094f8-86dc-4c78-a5d0-444ffbf02630")
	if err != nil {
		t.Fatalf("exact Goblin Freerunner registry lookup failed: %v", err)
	}
	if freerunnerDef.Name != "Goblin Freerunner" {
		t.Fatalf("exact Goblin Freerunner lookup returned %q", freerunnerDef.Name)
	}

	freerunnerID := addCardToHand(g, game.Player1, freerunnerDef)
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	g.Players[game.Player1].ManaPool.Add(mana.R, 1)
	cast := action.CastSpell(freerunnerID, nil, 0, nil)
	if !containsAction(engine.legalActions(g, game.Player1), cast) ||
		!engine.applyAction(g, game.Player1, cast) {
		t.Fatal("Goblin Freerunner was not castable for its Surge cost after a prior cast")
	}

	spell, ok := g.Stack.Peek()
	if !ok || spell.AlternativeCost != 1 || spell.ManaSpentToCast != 2 {
		t.Fatalf("Goblin Freerunner cast = %+v, want Surge alternative cost and {1}{R} paid", spell)
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	for _, permanent := range g.Battlefield {
		if permanent.CardInstanceID != freerunnerID {
			continue
		}
		if permanent.Controller != game.Player1 || !hasKeyword(g, permanent, game.Menace) {
			t.Fatal("Goblin Freerunner did not enter under its caster's control with menace")
		}
		return
	}
	t.Fatal("Goblin Freerunner did not enter the battlefield")
}
