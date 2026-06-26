package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// insteadTokenCounts extracts the two mutually-exclusive token counts from a
// "create A tokens. If <condition>, create N of those tokens instead." spell
// ability, asserting the structural shape the family lowers to: exactly two
// CreateToken instructions, the first gated on the negated condition and the
// second gated on the same condition unnegated.
func insteadTokenCounts(t *testing.T, mode game.Mode) (base, replacement game.CreateToken) {
	t.Helper()
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %#v, want two gated token creations", mode.Sequence)
	}
	base, ok := mode.Sequence[0].Primitive.(game.CreateToken)
	if !ok {
		t.Fatalf("instruction 0 = %#v, want CreateToken", mode.Sequence[0].Primitive)
	}
	replacement, ok = mode.Sequence[1].Primitive.(game.CreateToken)
	if !ok {
		t.Fatalf("instruction 1 = %#v, want CreateToken", mode.Sequence[1].Primitive)
	}
	baseCond := mode.Sequence[0].Condition
	replCond := mode.Sequence[1].Condition
	if !baseCond.Exists || !replCond.Exists ||
		!baseCond.Val.Condition.Exists || !replCond.Val.Condition.Exists {
		t.Fatalf("both creations must be gated, got base=%#v repl=%#v", baseCond, replCond)
	}
	if !baseCond.Val.Condition.Val.Negate {
		t.Fatalf("base condition must be negated, got %#v", baseCond.Val.Condition.Val)
	}
	if replCond.Val.Condition.Val.Negate {
		t.Fatalf("replacement condition must not be negated, got %#v", replCond.Val.Condition.Val)
	}
	return base, replacement
}

// TestLowerRiteOfReplicationInsteadCount verifies Rite of Replication's
// "Create a token that's a copy of target creature. If this spell was kicked,
// create five of those tokens instead." lowers to two gated copy-of-target
// token creations: one of the target unkicked and five of the same target
// kicked, both reusing the single target.
func TestLowerRiteOfReplicationInsteadCount(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Rite of Replication",
		Layout:   "normal",
		ManaCost: "{2}{U}{U}",
		TypeLine: "Sorcery",
		OracleText: "Kicker {5} (You may pay an additional {5} as you cast this spell.)\n" +
			"Create a token that's a copy of target creature. If this spell was kicked, create five of those tokens instead.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("no spell ability lowered")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 {
		t.Fatalf("targets = %d, want the single copied creature", len(mode.Targets))
	}
	base, replacement := insteadTokenCounts(t, mode)
	if got := base.Amount.Value(); got != 1 {
		t.Fatalf("base count = %d, want 1", got)
	}
	if got := replacement.Amount.Value(); got != 5 {
		t.Fatalf("kicked count = %d, want 5", got)
	}
	baseCopy, ok := base.Source.TokenCopy()
	if !ok {
		t.Fatalf("base source = %#v, want a copy-of-target source", base.Source)
	}
	replCopy, ok := replacement.Source.TokenCopy()
	if !ok {
		t.Fatalf("replacement source = %#v, want a copy-of-target source", replacement.Source)
	}
	if baseCopy.Object != replCopy.Object {
		t.Fatalf("counts must copy the same object: base=%#v repl=%#v", baseCopy.Object, replCopy.Object)
	}
	if k := mode.Sequence[0].Condition.Val.Condition.Val.SpellWasKicked; !k {
		t.Fatalf("base must be gated on SpellWasKicked, got %#v", mode.Sequence[0].Condition.Val.Condition.Val)
	}
}

// TestLowerSaprolingMigrationInsteadCount verifies Saproling Migration's typed
// token wording lowers to two gated creations of the same synthesized Saproling
// token def — two unkicked and four kicked.
func TestLowerSaprolingMigrationInsteadCount(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Saproling Migration",
		Layout:   "normal",
		ManaCost: "{1}{G}",
		TypeLine: "Sorcery",
		OracleText: "Kicker {4} (You may pay an additional {4} as you cast this spell.)\n" +
			"Create two 1/1 green Saproling creature tokens. If this spell was kicked, create four of those tokens instead.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("no spell ability lowered")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 0 {
		t.Fatalf("targets = %d, want none for a typed-token create", len(mode.Targets))
	}
	base, replacement := insteadTokenCounts(t, mode)
	if got := base.Amount.Value(); got != 2 {
		t.Fatalf("base count = %d, want 2", got)
	}
	if got := replacement.Amount.Value(); got != 4 {
		t.Fatalf("kicked count = %d, want 4", got)
	}
	baseDef, ok := base.Source.TokenDefRef()
	if !ok {
		t.Fatalf("base source = %#v, want a synthesized token def", base.Source)
	}
	replDef, ok := replacement.Source.TokenDefRef()
	if !ok {
		t.Fatalf("replacement source = %#v, want a synthesized token def", replacement.Source)
	}
	if baseDef == nil || replDef == nil {
		t.Fatalf("both creations need a token def: base=%v repl=%v", baseDef, replDef)
	}
	if baseDef.Name != "Saproling" || replDef.Name != baseDef.Name {
		t.Fatalf("counts must create the same token: base=%q repl=%q", baseDef.Name, replDef.Name)
	}
	if baseDef.Power != replDef.Power || baseDef.Toughness != replDef.Toughness {
		t.Fatalf("token P/T must match: base=%v/%v repl=%v/%v",
			baseDef.Power, baseDef.Toughness, replDef.Power, replDef.Toughness)
	}
}
