package parser

import "testing"

func tributeEffect(t *testing.T, source string) EffectSyntax {
	t.Helper()
	document, diagnostics := Parse(source, Context{CardName: "Tribute Bearer"})
	for _, diagnostic := range diagnostics {
		t.Fatalf("Parse(%q) produced diagnostic: %s", source, diagnostic.Summary)
	}
	for i := range document.Abilities {
		ability := &document.Abilities[i]
		if ability.Kind != AbilityReplacement {
			continue
		}
		for j := range ability.Sentences {
			sentence := &ability.Sentences[j]
			for k := range sentence.Effects {
				if sentence.Effects[k].Kind == EffectTribute {
					return sentence.Effects[k]
				}
			}
		}
	}
	t.Fatalf("Parse(%q) produced no Tribute replacement effect", source)
	return EffectSyntax{}
}

func TestExpandTributeKeywordCount(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		source string
		want   int
	}{
		{"tribute 1", "Tribute 1 (As this creature enters, an opponent of your choice may put a +1/+1 counter on it.)", 1},
		{"tribute 2", "Tribute 2 (As this creature enters, an opponent of your choice may put two +1/+1 counters on it.)", 2},
		{"tribute 6", "Tribute 6 (As this creature enters, an opponent of your choice may put six +1/+1 counters on it.)", 6},
		{"bare keyword", "Tribute 3", 3},
		{"after other keywords", "Flying\nTribute 2", 2},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			effect := tributeEffect(t, test.source)
			if !effect.EntersTribute {
				t.Fatal("EntersTribute = false, want true")
			}
			if !effect.Exact {
				t.Fatal("Exact = false, want true")
			}
			if effect.EntersTributeCount != test.want {
				t.Fatalf("EntersTributeCount = %d, want %d", effect.EntersTributeCount, test.want)
			}
		})
	}
}

func TestExpandTributeKeywordLeavesOtherTextAlone(t *testing.T) {
	t.Parallel()
	cases := []string{
		"When this creature enters, if tribute wasn't paid, you gain 4 life.",
		"Tribute to the World Tree",
	}
	for _, source := range cases {
		if got := expandTributeKeyword(source); got != source {
			t.Fatalf("expandTributeKeyword(%q) = %q, want unchanged", source, got)
		}
	}
}

func TestTributeWasNotPaidCondition(t *testing.T) {
	t.Parallel()
	clause := parseSingleConditionClause(t, "tribute wasn't paid")
	if clause.Predicate != ConditionPredicateSourceTributeNotPaid {
		t.Fatalf("predicate = %q, want ConditionPredicateSourceTributeNotPaid", clause.Predicate)
	}
}
