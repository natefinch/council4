package cardgen

import "testing"

func TestCorpusPolicy(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		card   ScryfallCard
		reason CorpusExclusionReason
	}{
		{
			name: "paper legal",
			card: ScryfallCard{Layout: "normal", Games: []string{"paper"}, Legalities: map[string]string{"legacy": "legal"}},
		},
		{
			name: "paper banned",
			card: ScryfallCard{Layout: "normal", Games: []string{"paper"}, Legalities: map[string]string{"commander": "banned"}},
		},
		{
			name: "digital printing of paper identity",
			card: ScryfallCard{
				Layout: "normal", SetType: "masters", Games: []string{"mtgo"}, Digital: true,
				Legalities: map[string]string{"vintage": "legal"},
			},
		},
		{
			name: "paper creature token",
			card: ScryfallCard{
				Layout: "token", SetType: "token", Games: []string{"paper"}, TypeLine: "Creature — Bear",
			},
		},
		{
			name: "paper artifact token face",
			card: ScryfallCard{
				Layout: "double_faced_token", SetType: "promo", Games: []string{"paper"},
				CardFaces: []ScryfallCardFace{{TypeLine: "Token Artifact — Treasure"}},
			},
		},
		{
			name:   "alchemy",
			card:   ScryfallCard{Layout: "normal", SetType: "alchemy", Legalities: map[string]string{"legacy": "legal"}},
			reason: ExcludeAlchemy,
		},
		{
			name:   "memorabilia",
			card:   ScryfallCard{Layout: "normal", SetType: "memorabilia", Games: []string{"paper"}},
			reason: ExcludeMemorabilia,
		},
		{
			name: "illegal funny card",
			card: ScryfallCard{
				Layout: "normal", SetType: "funny", Games: []string{"paper"},
				Legalities: map[string]string{"legacy": "not_legal"},
			},
			reason: ExcludeNoSanctionedPaperFormat,
		},
		{
			name: "illegal funny token",
			card: ScryfallCard{
				Layout: "token", SetType: "funny", Games: []string{"paper"}, TypeLine: "Creature",
			},
			reason: ExcludeNoSanctionedPaperFormat,
		},
		{
			name: "digital-only identity",
			card: ScryfallCard{
				Layout: "normal", SetType: "expansion", Games: []string{"arena"}, Digital: true,
				Legalities: map[string]string{"historic": "legal"},
			},
			reason: ExcludeDigitalOnly,
		},
		{
			name:   "scheme",
			card:   ScryfallCard{Layout: "scheme", SetType: "archenemy", Games: []string{"paper"}},
			reason: ExcludeSpecialFormat,
		},
		{
			name:   "checklist token record",
			card:   ScryfallCard{Layout: "token", SetType: "token", Games: []string{"paper"}, TypeLine: "Card"},
			reason: ExcludeSpecialFormat,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			reason, excluded := (CorpusPolicy{}).Exclusion(test.card)
			if excluded != (test.reason != "") || reason != test.reason {
				t.Fatalf("Exclusion() = (%q, %v), want (%q, %v)", reason, excluded, test.reason, test.reason != "")
			}
		})
	}
}

func TestDisownedCard(t *testing.T) {
	t.Parallel()
	disowned := []string{
		"Invoke Prejudice",
		"Cleanse",
		"Stone-Throwing Devils",
		"Pradesh Gypsies",
		"Jihad",
		"Imprison",
		"Crusade",
		"  crusade  ", // surrounding whitespace and lowercase still match
		"INVOKE PREJUDICE",
	}
	for _, name := range disowned {
		if !DisownedCard(ScryfallCard{Name: name}) {
			t.Errorf("DisownedCard(%q) = false, want true", name)
		}
	}
	allowed := []string{
		"Lightning Bolt",
		"Sol Ring",
		"Crusader of Odric", // shares a prefix with a disowned name but is not disowned
		"",
	}
	for _, name := range allowed {
		if DisownedCard(ScryfallCard{Name: name}) {
			t.Errorf("DisownedCard(%q) = true, want false", name)
		}
	}
}
