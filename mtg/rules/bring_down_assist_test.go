package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
)

const bringDownAssistOracleID = "05d2b89b-acc4-4289-95a3-a09ebb23a5d6"

// Source review: pinned Oracle corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4;
// Bring Down Oracle ID 05d2b89b-acc4-4289-95a3-a09ebb23a5d6, printing
// e18146f9-369c-41c8-8a1d-7737edd2c18e (Battlebond #26), exact text
// "Assist (Another player can pay up to {3} of this spell's cost.)
// Destroy target creature with power 4 or greater." Scryfall rulings URI
// https://api.scryfall.com/cards/e18146f9-369c-41c8-8a1d-7737edd2c18e/rulings
// retrieved 2026-11-21: four WotC rulings, all dated 2018-06-08; targets are
// chosen before the helper; Assist can pay only generic mana (cost changes
// adjust the available amount); colored costs remain the caster's responsibility.
// Read the official Wizards Comprehensive Rules pinned to effective 2026-09-25:
// https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.docx
// (downloaded under .cardwork/sources/MagicCompRules-20260925.docx,
// SHA-256 201037368d45407f271733a0575b756b3a868ce5e8024cf24aa1f30d48ecb960).
// Read 115.1a and 601.2c for target selection; 208.1 and 613.1g/4c for
// current power; 601.2f-h, 605.3a and 702.132a for locked cost, mana windows,
// helper payment of generic mana before the caster pays the remainder; 608.2b
// for resolution legality; and 701.8a-c for destruction and replacement.
func TestBringDownAssistCastAndResolves(t *testing.T) {
	registry := cards.NewDefaultRegistry()
	def, err := registry.LookupOracleID(bringDownAssistOracleID)
	if err != nil {
		t.Fatal(err)
	}
	if def.Name != "Bring Down" {
		t.Fatalf("Oracle ID lookup returned %q", def.Name)
	}

	for _, tc := range []struct {
		name       string
		helperPays bool
	}{
		{name: "helper pays generic mana", helperPays: true},
		{name: "caster pays full cost"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			target := addCombatPermanent(g, game.Player2, vanillaCreature("Four-power target", 4, 4))
			spell := addCardToHand(g, game.Player1, def)
			g.Players[game.Player1].ManaPool.Add(mana.W, 1)
			if tc.helperPays {
				g.Players[game.Player2].ManaPool.Add(mana.C, 3)
			} else {
				g.Players[game.Player1].ManaPool.Add(mana.C, 3)
			}
			setMainPhasePriority(g, game.Player1)
			engine := NewEngineWithPrograms(nil, registry.Programs())

			if !engine.applyAction(g, game.Player1, action.CastSpell(
				spell, []game.Target{game.PermanentTarget(target.ObjectID)}, 0, nil,
			)) {
				t.Fatal("Bring Down failed to cast with a legal 4-power target and sufficient total mana")
			}
			if g.Players[game.Player2].ManaPool.Total() != 0 || g.Players[game.Player1].ManaPool.Total() != 0 {
				t.Fatal("wrong player paid the generic or white cost")
			}
			engine.resolveTopOfStack(g, &TurnLog{})
			if !g.Players[game.Player2].Graveyard.Contains(target.CardInstanceID) ||
				!g.Players[game.Player1].Graveyard.Contains(spell) {
				t.Fatal("Bring Down did not destroy its legal target and finish in the caster's graveyard")
			}
		})
	}
}
