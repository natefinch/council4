//go:build card_support_gap_probe

package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// Source review: pinned corpus SHA-256 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4;
// Evincar's Justice Oracle ID c0d3dfc0-1ff2-4ac3-98eb-32eaa5b84c05, printing
// revision 2e0fffc6-4395-441e-95a9-c9cf33af7907 (CMA 58, 2017-06-09): "Buyback
// {3} (You may pay an additional {3} as you cast this spell. If you do, put
// this card into your hand as it resolves.) Evincar's Justice deals 2 damage to
// each creature and each player." Its rulings endpoint
// https://api.scryfall.com/cards/2e0fffc6-4395-441e-95a9-c9cf33af7907/rulings
// returned 0 records on 2026-11-11. Read CR 2026-09-25 (the corpus rules pin)
// from the official PDF https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.pdf
// and current official TXT https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt:
// 118.8a-b, 601.2b/f-h, 608.2f/n, 702.27a, 120.3a/e, and 120.4b-d require the
// creature and player damage actions to be processed as one simultaneous event.
func TestEvincarsJusticeCombinedDamageProbe(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	target := addCombatPermanent(g, game.Player2, vanillaCreature("Evincar probe creature", 3, 3))
	startingLife := g.Players[game.Player1].Life

	addEffectSpellToStack(g, game.Player1, game.Damage{
		Amount:    game.Fixed(2),
		Recipient: game.GroupDamageRecipient(game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}})),
	}, nil)
	NewEngine(nil).resolveTopOfStack(g, &TurnLog{})

	if target.MarkedDamage != 2 {
		t.Fatalf("creature damage = %d, want 2", target.MarkedDamage)
	}
	if g.Players[game.Player1].Life != startingLife-2 {
		t.Fatalf("one available creature-group damage operation left players untouched (life %d, want %d); issue #3762 tracks the missing combined creature/player damage batch", g.Players[game.Player1].Life, startingLife-2)
	}
}
