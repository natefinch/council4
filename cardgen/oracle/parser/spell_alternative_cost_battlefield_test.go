package parser

import (
	"testing"
)

// TestParsePermanentsOnBattlefieldManaAlternativeCost proves the parser
// recognizes Blasphemous Edict's board-state gate on a mana-only alternative
// cost in its trailing form ("... rather than pay this spell's mana cost if
// there are N or more <type>s on the battlefield."), carrying the threshold on
// ConditionCount and the counted permanent type on ConditionCardType.
func TestParsePermanentsOnBattlefieldManaAlternativeCost(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name     string
		text     string
		count    int
		cardType CardType
		mana     string
	}{
		{
			name:     "Blasphemous Edict trailing form",
			text:     "You may pay {B} rather than pay this spell's mana cost if there are thirteen or more creatures on the battlefield.",
			count:    13,
			cardType: CardTypeCreature,
			mana:     "{B}",
		},
		{
			name:     "generalized artifact type",
			text:     "You may pay {1} rather than pay this spell's mana cost if there are five or more artifacts on the battlefield.",
			count:    5,
			cardType: CardTypeArtifact,
			mana:     "{1}",
		},
		{
			name:     "leading form",
			text:     "If there are thirteen or more creatures on the battlefield, you may pay {B} rather than pay this spell's mana cost.",
			count:    13,
			cardType: CardTypeCreature,
			mana:     "{B}",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			alternative := parseManaAlternative(t, tc.text+"\nDraw a card.")
			if alternative.Condition != SpellAlternativeCostConditionPermanentsOnBattlefield {
				t.Fatalf("condition = %#v, want permanents-on-battlefield", alternative.Condition)
			}
			if alternative.ConditionCount != tc.count {
				t.Fatalf("count = %d, want %d", alternative.ConditionCount, tc.count)
			}
			if alternative.ConditionCardType != tc.cardType {
				t.Fatalf("card type = %#v, want %#v", alternative.ConditionCardType, tc.cardType)
			}
			if alternative.ConditionExactly {
				t.Fatal("board-state gate must never be an exact-count comparison")
			}
			if alternative.ManaCost.String() != tc.mana {
				t.Fatalf("mana cost = %q, want %q", alternative.ManaCost.String(), tc.mana)
			}
		})
	}
}

// TestParsePermanentsOnBattlefieldManaAlternativeCostFailsClosed proves the
// board-state gate recognizer is strict: only "if there are N or more
// <permanent type>s on the battlefield" (as a lone leading or trailing
// condition) is accepted. Every near-miss wording — a non-permanent or singular
// type, a different zone, an extra qualifier, a non-count word, or a leading and
// trailing condition together — must be left unrecognized rather than
// approximated.
func TestParsePermanentsOnBattlefieldManaAlternativeCostFailsClosed(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name string
		text string
	}{
		{
			name: "non-permanent type",
			text: "You may pay {B} rather than pay this spell's mana cost if there are thirteen or more instants on the battlefield.",
		},
		{
			name: "singular type word",
			text: "You may pay {B} rather than pay this spell's mana cost if there are thirteen or more creature on the battlefield.",
		},
		{
			name: "different zone",
			text: "You may pay {B} rather than pay this spell's mana cost if there are thirteen or more creatures in your graveyard.",
		},
		{
			name: "extra trailing qualifier",
			text: "You may pay {B} rather than pay this spell's mana cost if there are thirteen or more creatures on the battlefield you control.",
		},
		{
			name: "non-count quantity word",
			text: "You may pay {B} rather than pay this spell's mana cost if there are many or more creatures on the battlefield.",
		},
		{
			name: "missing there are",
			text: "You may pay {B} rather than pay this spell's mana cost if thirteen or more creatures are on the battlefield.",
		},
		{
			name: "exact count instead of or more",
			text: "You may pay {B} rather than pay this spell's mana cost if there are exactly thirteen creatures on the battlefield.",
		},
		{
			name: "leading and trailing condition together",
			text: "If there are thirteen or more creatures on the battlefield, you may pay {B} rather than pay this spell's mana cost if there are thirteen or more creatures on the battlefield.",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(tc.text+"\nDraw a card.", Context{InstantOrSorcery: true})
			for _, ability := range document.Abilities {
				if ability.AlternativeCost != nil &&
					ability.AlternativeCost.Kind == SpellAlternativeCostMana {
					t.Fatalf("wording was wrongly recognized as a mana-only alternative cost: %#v",
						ability.AlternativeCost)
				}
			}
		})
	}
}
