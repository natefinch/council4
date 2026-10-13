package parser

import "testing"

func TestPermanentUnionSharedScopeRefuses(t *testing.T) {
	for _, phrase := range []string{
		"target tapped Merfolk or Kithkin creature",
		"target Merfolk or Kithkin creature",
		"target blue Merfolk or Kithkin creature",
		"target tapped Merfolk or tapped Kithkin creature",
		"target Forest or Plains land",
		"target nontoken artifact or enchantment",
		"target creature token or artifact",
		"target tapped creature or Vehicle",
		"target attacking creature or Vehicle",
		"target multicolored creature or enchantment",
		"target legendary creature or Vehicle",
		"target creature or Vehicle with mana value 3 or less",
		"target creature or Vehicle with mana value 3",
	} {
		t.Run(phrase, func(t *testing.T) {
			document, diagnostics := Parse("Tap "+phrase+".", Context{InstantOrSorcery: true})
			target := document.Abilities[0].Sentences[0].Targets[0]
			if target.Exact {
				t.Fatal("shared noun/adjective scope certified as independent predicates")
			}
			if len(diagnostics) != 1 || diagnostics[0].Summary != "unsupported permanent union scope" ||
				diagnostics[0].Detail != "shared noun or qualifier scope is not proven by independent permanent predicates" ||
				diagnostics[0].Span != target.Span {
				t.Fatalf("missing precise parser-owned scope diagnostic: %#v", diagnostics)
			}
		})
	}
}

func TestPermanentUnionIndependentScopeControls(t *testing.T) {
	for _, phrase := range []string{
		"target blue creature or artifact",
		"target creature or tapped Vehicle",
		"target tapped Merfolk creature or tapped Kithkin creature",
		"target Vehicle or creature",
		"target Saga or enchantment creature you control",
		"target multicolored creature or multicolored enchantment",
		"target enchanted creature or enchantment creature",
		"target artifact or tapped creature",
		"target nontoken artifact or nontoken enchantment",
		"target tapped creature or untapped Vehicle",
		"target creature with mana value 3 or less or Vehicle with mana value 3 or less",
		"target tapped Merfolk or Kithkin",
		"target Wolf or Werewolf",
		"target artifact or creature",
		"target artifact creature",
	} {
		t.Run(phrase, func(t *testing.T) {
			if target := singleTarget(t, "Tap "+phrase+"."); !target.Exact {
				t.Fatalf("independent or existing exact predicate lost: %#v", target)
			}
		})
	}
}
