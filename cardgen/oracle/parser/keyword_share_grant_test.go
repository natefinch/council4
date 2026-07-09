package parser

import (
	"testing"
)

// odricKeywordShareText is Odric, Lunarch Marshal's exact oracle text: the team
// keyword-sharing construct this recognizer targets.
const odricKeywordShareText = "At the beginning of each combat, creatures you control gain first strike until end of turn if a creature you control has first strike. The same is true for flying, deathtouch, double strike, haste, hexproof, indestructible, lifelink, menace, reach, skulk, trample, and vigilance."

// anyKeywordShareGrant returns the first recognized KeywordShareGrant clause
// across a document's abilities, or nil when none matched.
func anyKeywordShareGrant(abilities []Ability) *KeywordShareGrantClause {
	for i := range abilities {
		if abilities[i].KeywordShareGrant != nil {
			return abilities[i].KeywordShareGrant
		}
	}
	return nil
}

// TestKeywordShareGrantRecognizesOdric proves the recognizer folds Odric's whole
// body onto a KeywordShareGrant clause carrying the lead keyword plus every
// keyword named by "The same is true for ..." in printed order.
func TestKeywordShareGrantRecognizesOdric(t *testing.T) {
	t.Parallel()

	document, diagnostics := Parse(odricKeywordShareText, Context{CardName: "Odric, Lunarch Marshal"})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	share := anyKeywordShareGrant(document.Abilities)
	if share == nil {
		t.Fatalf("Odric was not recognized as a keyword share; abilities = %#v", document.Abilities)
	}
	want := []KeywordKind{
		KeywordFirstStrike,
		KeywordFlying,
		KeywordDeathtouch,
		KeywordDoubleStrike,
		KeywordHaste,
		KeywordHexproof,
		KeywordIndestructible,
		KeywordLifelink,
		KeywordMenace,
		KeywordReach,
		KeywordSkulk,
		KeywordTrample,
		KeywordVigilance,
	}
	if len(share.Keywords) != len(want) {
		t.Fatalf("keywords = %#v, want %d in printed order", share.Keywords, len(want))
	}
	for i, kind := range want {
		if share.Keywords[i] != kind {
			t.Fatalf("keyword %d = %q, want %q", i, share.Keywords[i], kind)
		}
	}
}

// TestKeywordShareGrantRecognizesLeadOnly proves the "The same is true for ..."
// sentence is optional: a lone gated lead sentence recognizes a single keyword.
func TestKeywordShareGrantRecognizesLeadOnly(t *testing.T) {
	t.Parallel()

	source := "At the beginning of each combat, creatures you control gain first strike until end of turn if a creature you control has first strike."
	document, diagnostics := Parse(source, Context{CardName: "Test Lead Only"})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	share := anyKeywordShareGrant(document.Abilities)
	if share == nil {
		t.Fatal("lead-only keyword share was not recognized")
	}
	if len(share.Keywords) != 1 || share.Keywords[0] != KeywordFirstStrike {
		t.Fatalf("keywords = %#v, want [KeywordFirstStrike]", share.Keywords)
	}
}

// TestKeywordShareGrantFailsClosed proves the recognizer discriminates on exact
// wording and keyword grantability: every sibling that shares the "the same is
// true for ..." surface but differs in trigger, subject, gate, or keyword set is
// left untouched so the compiler never lowers a construct it does not model.
func TestKeywordShareGrantFailsClosed(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source string
		reason string
	}{
		{
			name:   "Concerted Effort",
			source: "At the beginning of each upkeep, creatures you control gain flying until end of turn if a creature you control has flying. The same is true for fear, first strike, double strike, landwalk, protection, trample, and vigilance.",
			reason: "landwalk and protection are not grantable",
		},
		{
			name:   "Oddric, Lunar Marquis",
			source: `At the beginning of each combat, creatures you control gain banding until end of turn if a creature you control has banding. The same is true for changeling, devoid, fear, flanking, horsemanship, ingest, intimidate, landwalk, shroud, tantrum, wither, and the activated ability "Sacrifice this creature: Add {C}."`,
			reason: "banding lead keyword is not grantable",
		},
		{
			name:   "Bleeding Effect",
			source: "At the beginning of combat on your turn, creatures you control gain flying until end of turn if a creature card in your graveyard has flying. The same is true for first strike, double strike, deathtouch, hexproof, indestructible, lifelink, menace, reach, trample, and vigilance.",
			reason: "gate is a creature card in your graveyard, not a creature you control",
		},
		{
			name:   "Majestic Myriarch",
			source: "Majestic Myriarch's power and toughness are each equal to twice the number of creatures you control.\nAt the beginning of each combat, this creature gains flying until end of turn if you control a creature with flying. The same is true for first strike, double strike, deathtouch, haste, hexproof, indestructible, lifelink, menace, reach, trample, and vigilance.",
			reason: "subject is this creature, not creatures you control",
		},
		{
			name:   "Thunderous Orator",
			source: "Vigilance\nWhenever this creature attacks, it gains flying until end of turn if you control a creature with flying. The same is true for first strike, double strike, deathtouch, indestructible, lifelink, menace, and trample.",
			reason: "attack trigger, not a phase/step trigger",
		},
		{
			name:   "Sproutwatch Dryad",
			source: "At the beginning of each combat, Sproutwatch Dryad gains flying until end of turn if a creature you control or a card in your hand has flying. The same is true for first strike, double strike, deathtouch, haste, hexproof, indestructible, lifelink, menace, reach, trample, and vigilance.",
			reason: "self subject and expanded gate",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(tc.source, Context{CardName: tc.name})
			if share := anyKeywordShareGrant(document.Abilities); share != nil {
				t.Fatalf("%s recognized as keyword share (%s); keywords = %#v", tc.name, tc.reason, share.Keywords)
			}
		})
	}
}
