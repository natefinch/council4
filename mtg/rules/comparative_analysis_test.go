package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
)

// Source review: Oracle ID 59b869b9-cb48-4152-ad91-52f582e2df5b, Oath of the
// Gatewatch #51, printing revision dd83129b-7e8c-4cc5-a7b3-e0ae221d7ad4;
// pinned corpus SHA-256 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4.
// Current Oracle retrieved 2026-11-20 from
// https://api.scryfall.com/cards/dd83129b-7e8c-4cc5-a7b3-e0ae221d7ad4:
// "{3}{U} Instant. Surge {2}{U} (You may cast this spell for its surge cost if
// you or a teammate has cast another spell this turn.) Target player draws two
// cards."
// Rulings retrieved the same day from
// https://api.scryfall.com/cards/dd83129b-7e8c-4cc5-a7b3-e0ae221d7ad4/rulings:
// four WotC rulings dated 2016-01-22. They clarify that Surge is an alternative
// cost (some cards add a rider, but Comparative Analysis does not); the earlier
// spell can have resolved, been countered, or (for instants) still be on the
// stack; copies of a Surge instant or sorcery inherit the paid Surge cost; and
// paying Surge does not change mana cost or mana value.
// Official Wizards Comprehensive Rules effective 2026-11-06,
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt.
// Read 102.3-4 (teammates and "your team"); 115.1a, 601.2c, and 608.2b (target
// player chosen on casting and rechecked on resolution); 118.9a-d, 601.2b,f-h,
// and 702.117a (optional Surge alternative cost and its eligibility/payment);
// 202.3 and 118.9c (mana value remains based on the printed cost); and 121.1-2
// (draw two cards as two individual draws). The named scenario casts for Surge
// after the caster's earlier resolved spell, pays {2}{U}, targets Player 2,
// and verifies exactly two cards move from that player's library to hand.
func TestComparativeAnalysisSurgeDrawsTargetPlayerTwoCards(t *testing.T) {
	const oracleID = "59b869b9-cb48-4152-ad91-52f582e2df5b"
	registry := cards.NewDefaultRegistry()
	def, err := registry.LookupOracleID(oracleID)
	if err != nil {
		t.Fatalf("lookup Comparative Analysis by Oracle ID: %v", err)
	}
	if def.Name != "Comparative Analysis" {
		t.Fatalf("Oracle ID resolved to %q", def.Name)
	}

	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngineWithPrograms(nil, registry.Programs())
	setMainPhasePriority(g, game.Player1)
	castSurgeSetupSpell(t, engine, g, game.Player1)
	engine.resolveTopOfStack(g, &TurnLog{})

	casterLibraryCardOne := addCardToLibrary(g, game.Player1, &game.CardDef{
		CardFace: game.CardFace{Name: "Caster library card one"},
	})
	casterLibraryCardTwo := addCardToLibrary(g, game.Player1, &game.CardDef{
		CardFace: game.CardFace{Name: "Caster library card two"},
	})
	targetLibraryCardOne := addCardToLibrary(g, game.Player2, &game.CardDef{
		CardFace: game.CardFace{Name: "Target library card one"},
	})
	targetLibraryCardTwo := addCardToLibrary(g, game.Player2, &game.CardDef{
		CardFace: game.CardFace{Name: "Target library card two"},
	})
	spellID := addCardToHand(g, game.Player1, def)
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)
	g.Players[game.Player1].ManaPool.Add(mana.U, 1)

	cast := action.CastSpell(spellID, []game.Target{game.PlayerTarget(game.Player2)}, 0, nil)
	if !containsAction(engine.legalActions(g, game.Player1), cast) ||
		!engine.applyAction(g, game.Player1, cast) {
		t.Fatal("Comparative Analysis was not castable for its Surge cost with a target player after the prior spell")
	}
	spell, ok := g.Stack.Peek()
	if !ok || spell.AlternativeCost != 1 || spell.ManaSpentToCast != 3 ||
		g.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatalf("Comparative Analysis did not pay {2}{U} Surge: spell=%+v, mana=%d",
			spell, g.Players[game.Player1].ManaPool.Total())
	}

	engine.resolveTopOfStack(g, &TurnLog{})
	target := g.Players[game.Player2]
	if target.Hand.Size() != 2 || target.Library.Size() != 0 ||
		!target.Hand.Contains(targetLibraryCardOne) || !target.Hand.Contains(targetLibraryCardTwo) {
		t.Fatalf("target player's hand/library after resolving Comparative Analysis = %d/%d; want both library cards in hand",
			target.Hand.Size(), target.Library.Size())
	}
	caster := g.Players[game.Player1]
	if caster.Library.Size() != 2 || caster.Hand.Contains(casterLibraryCardOne) ||
		caster.Hand.Contains(casterLibraryCardTwo) {
		t.Fatal("Comparative Analysis drew for its caster instead of the targeted player")
	}
}
