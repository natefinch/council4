package parser

import "testing"

// TestParseStaticGroupAnthemAlsoAdverb proves the parser drops the "also" adverb
// a stacked threshold anthem prints between a controlled-creature group and its
// power/toughness verb ("Creatures you control also get +1/+0 and have
// trample ...", Jetmir, Nexus of Revels' follow-on clauses). The follow-on clause
// with "also" must parse into the same [PowerToughness, KeywordGrant] node
// sequence over the controlled-creatures group as the identical clause without
// "also", so the emphasis adverb carries no additional meaning.
func TestParseStaticGroupAnthemAlsoAdverb(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"without also": "Creatures you control get +1/+0 and have trample as long as you control six or more creatures.",
		"with also":    "Creatures you control also get +1/+0 and have trample as long as you control six or more creatures.",
	}
	for name, source := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			declarations := parseStaticDeclarationSyntax(t, source, Context{})
			if len(declarations) != 2 {
				t.Fatalf("declarations = %#v, want two (power/toughness and keyword grant)", declarations)
			}
			if declarations[0].Kind != StaticDeclarationContinuousPowerToughness {
				t.Fatalf("declaration[0] kind = %v, want continuous power/toughness", declarations[0].Kind)
			}
			if declarations[1].Kind != StaticDeclarationKeywordGrant {
				t.Fatalf("declaration[1] kind = %v, want keyword grant", declarations[1].Kind)
			}
			for i := range declarations {
				subject := declarations[i].Subject
				if subject.Kind != StaticDeclarationSubjectGroup ||
					subject.Group.Kind != EffectStaticSubjectControlledCreatures {
					t.Fatalf("declaration[%d] subject = %#v, want controlled-creatures group", i, subject)
				}
			}
		})
	}
}
