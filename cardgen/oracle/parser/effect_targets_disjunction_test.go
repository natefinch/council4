package parser

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/compare"
)

// TestParseQualifiedDisjunctivePermanentTarget proves an Oxford-comma
// disjunction whose members carry per-member qualifiers (a tapped state, a
// keyword, or a power comparison) parses into a permanent target with one
// Selection.Alternatives entry per member, lifting a shared controller clause to
// the parent so the qualifier stays attached only to its own member.
func TestParseQualifiedDisjunctivePermanentTarget(t *testing.T) {
	t.Parallel()
	target := singleTarget(t, "Exile target artifact, enchantment, or tapped creature an opponent controls.")
	if !target.Exact ||
		target.Selection.Kind != SelectionPermanent ||
		target.Selection.Controller != SelectionControllerOpponent ||
		len(target.Selection.Alternatives) != 3 {
		t.Fatalf("target = %#v, want exact opponent-controlled three-member disjunction", target)
	}
	if target.Selection.Alternatives[0].Kind != SelectionArtifact ||
		target.Selection.Alternatives[1].Kind != SelectionEnchantment {
		t.Fatalf("alternatives = %#v, want artifact then enchantment", target.Selection.Alternatives)
	}
	creature := target.Selection.Alternatives[2]
	if creature.Kind != SelectionCreature || !creature.Tapped ||
		creature.Controller != SelectionControllerAny {
		t.Fatalf("creature alternative = %#v, want controller-free tapped creature", creature)
	}
}

// TestParseDisjunctivePermanentTargetKeywordMember proves a keyword qualifier on
// the final disjunction member ("creature with flying") attaches to that member
// only.
func TestParseDisjunctivePermanentTargetKeywordMember(t *testing.T) {
	t.Parallel()
	target := singleTarget(t, "Destroy target artifact, enchantment, or creature with flying.")
	if !target.Exact || len(target.Selection.Alternatives) != 3 {
		t.Fatalf("target = %#v, want exact three-member disjunction", target)
	}
	creature := target.Selection.Alternatives[2]
	if creature.Kind != SelectionCreature || creature.Keyword != KeywordFlying {
		t.Fatalf("creature alternative = %#v, want creature with flying", creature)
	}
}

// TestParseDisjunctivePermanentTargetPowerMember proves a "with power N or
// greater" comparison on the final member attaches to that member only.
func TestParseDisjunctivePermanentTargetPowerMember(t *testing.T) {
	t.Parallel()
	target := singleTarget(t, "Exile target artifact, enchantment, or creature with power 4 or greater.")
	if !target.Exact || len(target.Selection.Alternatives) != 3 {
		t.Fatalf("target = %#v, want exact three-member disjunction", target)
	}
	creature := target.Selection.Alternatives[2]
	if creature.Kind != SelectionCreature || !creature.MatchPower ||
		creature.Power.Op != compare.GreaterOrEqual || creature.Power.Value != 4 {
		t.Fatalf("creature alternative = %#v, want creature with power 4 or greater", creature)
	}
}

// TestParseDisjunctivePermanentTargetUpToOne proves the optional "up to one"
// cardinality reuses the disjunction reconstruction.
func TestParseDisjunctivePermanentTargetUpToOne(t *testing.T) {
	t.Parallel()
	target := singleTarget(t, "Destroy up to one target artifact, enchantment, or tapped creature.")
	if !target.Exact ||
		target.Cardinality != (TargetCardinalitySyntax{Min: 0, Max: 1}) ||
		len(target.Selection.Alternatives) != 3 {
		t.Fatalf("target = %#v, want exact up-to-one three-member disjunction", target)
	}
}

// TestParseBareTypeUnionStaysFlattened proves a disjunction of bare card types
// keeps its existing flattened RequiredTypesAny parse, so the qualified
// disjunctive path never alters an already-supported plain union.
func TestParseBareTypeUnionStaysFlattened(t *testing.T) {
	t.Parallel()
	target := singleTarget(t, "Destroy target artifact, enchantment, or creature.")
	if len(target.Selection.Alternatives) != 0 ||
		len(target.Selection.RequiredTypesAny) != 3 {
		t.Fatalf("target = %#v, want flattened three-type union", target)
	}
}

func singleTarget(t *testing.T, source string) TargetSyntax {
	t.Helper()
	document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("Parse(%q) diagnostics = %#v", source, diagnostics)
	}
	if len(document.Abilities) != 1 || len(document.Abilities[0].Sentences) != 1 {
		t.Fatalf("Parse(%q) shape = %#v", source, document.Abilities)
	}
	targets := document.Abilities[0].Sentences[0].Targets
	if len(targets) != 1 {
		t.Fatalf("Parse(%q) targets = %#v, want one", source, targets)
	}
	return targets[0]
}
