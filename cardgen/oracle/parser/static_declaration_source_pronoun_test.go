package parser

import "testing"

// The condition-first source-conditional self-anthem "As long as <source-state>,
// it gets +X/+Y and/or has <keyword(s)>" names the source permanent through the
// pronoun "it". The parser must thread that pronoun to the same SourceCreature
// subject the trailing "This creature ... as long as it's equipped" form names,
// so downstream layers resolve the grant to the source.

func TestParseStaticSourcePronounConditionalKeywordGrantMeaning(t *testing.T) {
	t.Parallel()
	declarations := parseStaticDeclarationSyntax(t, "As long as this creature is equipped, it has flying.", Context{})
	if len(declarations) != 1 {
		t.Fatalf("declarations = %#v, want one", declarations)
	}
	declaration := declarations[0]
	if declaration.Kind != StaticDeclarationKeywordGrant {
		t.Fatalf("kind = %v, want keyword grant", declaration.Kind)
	}
	if declaration.Subject.Kind != StaticDeclarationSubjectSourceCreature {
		t.Fatalf("subject = %#v, want source creature", declaration.Subject)
	}
	if !declaration.HasCondition {
		t.Fatalf("declaration = %#v, want a condition", declaration)
	}
}

func TestParseStaticSourcePronounConditionalPowerToughnessMeaning(t *testing.T) {
	t.Parallel()
	declarations := parseStaticDeclarationSyntax(t, "As long as this creature is enchanted, it gets +1/+1.", Context{})
	if len(declarations) != 1 {
		t.Fatalf("declarations = %#v, want one", declarations)
	}
	declaration := declarations[0]
	if declaration.Kind != StaticDeclarationContinuousPowerToughness {
		t.Fatalf("kind = %v, want power/toughness", declaration.Kind)
	}
	if declaration.Subject.Kind != StaticDeclarationSubjectSourceCreature {
		t.Fatalf("subject = %#v, want source creature", declaration.Subject)
	}
	if declaration.PowerDelta.Value != 1 || declaration.ToughnessDelta.Value != 1 || !declaration.HasCondition {
		t.Fatalf("declaration = %#v, want conditional +1/+1", declaration)
	}
}
