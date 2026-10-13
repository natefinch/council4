package parser

import "testing"

func TestPermanentUnionSharedTailScopeRefuses(t *testing.T) {
	for _, tail := range []string{
		"with flying",
		"without flying",
		"with power 3 or less",
		"with toughness 3 or less",
		"with lesser power",
		"with greater power",
		"with power 3 and toughness 3",
		"with mana value 3 and power 3 and toughness 3",
	} {
		for _, nouns := range []string{"creature or Vehicle", "Vehicle or creature"} {
			t.Run(nouns+" "+tail, func(t *testing.T) {
				document, diagnostics := Parse("Tap target "+nouns+" "+tail+".", Context{InstantOrSorcery: true})
				target := document.Abilities[0].Sentences[0].Targets[0]
				if target.Exact || len(diagnostics) != 1 || diagnostics[0].Summary != "unsupported permanent union scope" {
					t.Fatalf("shared tail became last-branch-only: target=%#v diagnostics=%#v", target, diagnostics)
				}
			})
		}
	}
}

func TestPermanentUnionExplicitTailScopeControls(t *testing.T) {
	for _, phrase := range []string{
		"creature with flying or Vehicle with flying",
		"creature without flying or Vehicle without flying",
		"creature with power 3 or less or Vehicle with power 3 or less",
		"creature with toughness 3 or less or Vehicle with toughness 3 or less",
		"creature with lesser power or Vehicle with lesser power",
		"creature with greater power or Vehicle with greater power",
		"creature with power 3 and toughness 3 or Vehicle with power 3 and toughness 3",
		"creature with flying or Vehicle",
		"creature without flying or Vehicle",
		"creature with power 3 or less or Vehicle",
		"creature with toughness 3 or less or Vehicle",
		"blue creature or artifact",
		"creature or tapped Vehicle",
		"enchanted creature or enchantment creature",
	} {
		t.Run(phrase, func(t *testing.T) {
			if target := singleTarget(t, "Tap target "+phrase+"."); !target.Exact {
				t.Fatalf("explicit independent branch scope lost: %#v", target)
			}
		})
	}
}

func TestPermanentUnionUnrepresentedQualifierStaysClosed(t *testing.T) {
	for _, phrase := range []string{
		"1/1 creature or Vehicle",
		"creature or 1/1 Vehicle",
		"creature or Vehicle with a +1/+1 counter on it",
		"creature or Vehicle with ward {2}",
		"creature or Vehicle with flying and power 3",
		"creature or Vehicle with mana value X",
		"creature or Vehicle of the chosen type",
	} {
		t.Run(phrase, func(t *testing.T) {
			document, _ := Parse("Tap target "+phrase+".", Context{InstantOrSorcery: true})
			if target := document.Abilities[0].Sentences[0].Targets[0]; target.Exact {
				t.Fatalf("unrepresented mixed-branch qualifier admitted: %#v", target)
			}
		})
	}
}

func TestPermanentUnionTypedPrefixCannotProveTailScope(t *testing.T) {
	prefix := SelectionSyntax{
		PowerToughnessPrefix: true, MatchPower: true, MatchToughness: true,
	}
	for _, tail := range []SelectionSyntax{
		{Keyword: KeywordFlying},
		{ExcludedKeyword: KeywordFlying},
		{MatchManaValue: true},
		{MatchPower: true},
		{MatchToughness: true},
		{PowerLessThanSource: true},
		{PowerGreaterThanSource: true},
	} {
		if permanentUnionScopeProven(SelectionSyntax{}, tail) || permanentUnionScopeProven(prefix, tail) {
			t.Fatalf("a prefix or empty branch proved trailing scope: %#v", tail)
		}
		if !permanentUnionScopeProven(tail, tail) {
			t.Fatalf("fully explicit same-dimension tails lost: %#v", tail)
		}
	}
	if !permanentUnionScopeProven(SelectionSyntax{}, prefix) {
		t.Fatal("explicit right prefix was treated as a shared trailing bound")
	}
}
