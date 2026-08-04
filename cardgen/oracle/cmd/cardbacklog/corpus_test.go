package main

import (
	"testing"

	"github.com/natefinch/council4/cardgen"
)

// legalCard returns a minimal ScryfallCard that CorpusPolicy.Exclusion accepts,
// so tests can isolate the DisownedCard check in evaluateCard from the ordinary
// corpus-policy exclusions covered elsewhere.
func legalCard(id, name string) cardgen.ScryfallCard {
	return cardgen.ScryfallCard{
		ID:         id,
		OracleID:   id,
		Name:       name,
		Layout:     "normal",
		TypeLine:   "Enchantment",
		OracleText: "Test.",
		Legalities: map[string]string{"commander": "legal"},
	}
}

// TestEvaluateCardDropsDisownedCards pins the reconciliation-guard fix: a card
// Wizards of the Coast has disowned (DisownedCard) must come back ineligible,
// exactly as compilecards and parsercoverage treat it. Before this fix,
// evaluateCard only consulted CorpusPolicy.Exclusion, which does not know about
// disowned cards (they are ordinary legal, sanctioned-format cards by every
// other measure) — that returns them eligible, and applyAuthority read their
// absence from both the compile report's unsupported and excluded sets as an
// authoritative "generated" verdict, when in fact compilecards had omitted them
// entirely. The result was a fabricated reconciliation divergence on every one
// of Wizards' seven disowned cards, on every run, forever.
func TestEvaluateCardDropsDisownedCards(t *testing.T) {
	t.Parallel()
	card := legalCard("b18b9869-8490-4875-a5bb-484c3299f2c5", "Jihad")
	outcome := evaluateCard(job{index: 0, card: card})
	if outcome.eligible {
		t.Fatalf("evaluateCard(%q) eligible = true, want false: a disowned card must never be routed", card.Name)
	}
}

// TestEvaluateCardKeepsOrdinaryCardsEligible guards against overcorrecting
// TestEvaluateCardDropsDisownedCards into dropping every card.
func TestEvaluateCardKeepsOrdinaryCardsEligible(t *testing.T) {
	t.Parallel()
	card := legalCard("00000000-0000-0000-0000-000000000000", "Test Card")
	outcome := evaluateCard(job{index: 0, card: card})
	if !outcome.eligible {
		t.Fatalf("evaluateCard(%q) eligible = false, want true for an ordinary legal card", card.Name)
	}
}
