package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards/authored"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/types"
)

// Source review: pinned corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4;
// Mogg Salvage Oracle ID 0923ad85-2ea4-4463-8f71-b7d0a162b486, MH2 #282,
// printing revision 2b93e65e-8a10-444f-8b0a-9327dd30cce2. Current Oracle,
// verified 2026-11-19:
// https://api.scryfall.com/cards/2b93e65e-8a10-444f-8b0a-9327dd30cce2
// {2}{R} Instant. "If an opponent controls an Island and you control a
// Mountain, you may cast this spell without paying its mana cost. Destroy
// target artifact."
// Rulings, retrieved 2026-11-19 from
// https://api.scryfall.com/cards/2b93e65e-8a10-444f-8b0a-9327dd30cce2/rulings:
// any opponent may control the Island, irrespective of the artifact target;
// land control matters only during casting; the spell's mana value remains 3.
// The endpoint returned one unrelated Nevinyrral's Disk ruling.
// Official Wizards Comprehensive Rules effective 2026-11-06:
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt
// Read §§109.5, 110.2, 118.9a-d, 202.3, 205.3i, 305.6, 601.2a-c/e/f/h,
// 608.2b-c, 701.8a-b, 806.1 and 903.2. The would-be caster and current
// controllers determine the two land predicates at cast time; the alternative
// cost leaves mana value unchanged and composes with the normal casting cost
// process. The artifact target is chosen during casting, checked on resolution,
// and destroyed to its owner's graveyard if still legal. Commander multiplayer
// defaults to Free-for-All.
func TestAuthoredMoggSalvageScenario(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, authored.MoggSalvage.New())
	mountain := addBasicLandPermanent(g, game.Player1, types.Mountain)
	mountain.Tapped = true
	artifact := addArtifactPermanent(g, game.Player2)
	setMainPhasePriority(g, game.Player1)

	act := action.CastSpell(spellID, []game.Target{game.PermanentTarget(artifact.ObjectID)}, 0, nil)
	if containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("free-cast alternative was offered without an opponent-controlled Island")
	}
	addBasicLandPermanent(g, game.Player3, types.Island)
	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("free-cast alternative was not offered with a Mountain and an opponent-controlled Island")
	}
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("Mogg Salvage could not be cast for free at its printed cost gate")
	}
	if spell, ok := g.Stack.Peek(); !ok || spell.ManaSpentToCast != 0 {
		t.Fatalf("cast did not use the free alternative: %+v", spell)
	}

	engine.resolveTopOfStack(g, &TurnLog{})
	if !g.Players[game.Player2].Graveyard.Contains(artifact.CardInstanceID) ||
		!g.Players[game.Player1].Graveyard.Contains(spellID) {
		t.Fatal("Mogg Salvage did not destroy the targeted artifact and finish in its owner's graveyard")
	}
}
