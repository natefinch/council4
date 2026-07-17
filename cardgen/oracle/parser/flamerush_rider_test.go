package parser

import "testing"

const flamerushRiderOracle = "Whenever this creature attacks, create a token that's a copy of another target attacking creature and that's tapped and attacking. Exile the token at end of combat.\n" +
	"Dash {2}{R}{R} (You may cast this spell for its dash cost. If you do, it gains haste, and it's returned from the battlefield to its owner's hand at the beginning of the next end step.)"

func TestParseFlamerushRiderComposableAttackCopy(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(flamerushRiderOracle, Context{CardName: "Flamerush Rider"})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(document.Abilities) != 2 {
		t.Fatalf("abilities = %d, want attack trigger and dash", len(document.Abilities))
	}
	var effects []EffectSyntax
	for _, sentence := range document.Abilities[0].Sentences {
		effects = append(effects, sentence.Effects...)
	}
	if len(effects) != 2 {
		t.Fatalf("effects = %d, want create and delayed exile", len(effects))
	}
	create := effects[0]
	if create.Kind != EffectCreate || !create.Exact ||
		!create.TokenCopyOfTarget ||
		!create.TokenCopyEntersTapped ||
		!create.TokenCopyAttacksWithTarget ||
		len(create.Targets) != 1 ||
		!create.Targets[0].Selection.Another ||
		!create.Targets[0].Selection.Attacking {
		t.Fatalf("create kind=%v exact=%v copy=%v tapped=%v attacking=%v targets=%#v",
			create.Kind, create.Exact, create.TokenCopyOfTarget,
			create.TokenCopyEntersTapped, create.TokenCopyAttacksWithTarget, create.Targets)
	}
	exile := effects[1]
	if exile.Kind != EffectExile || !exile.Exact ||
		!exile.CreatedTokensReference ||
		exile.DelayedTiming != DelayedTimingEndOfCombat {
		t.Fatalf("exile kind=%v exact=%v created=%v timing=%v targets=%#v refs=%#v text=%q",
			exile.Kind, exile.Exact, exile.CreatedTokensReference, exile.DelayedTiming,
			exile.Targets, exile.References, exile.Text)
	}
}

func TestParseAttackingTargetCopyIsCardNameIndependent(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"Whenever this creature attacks, create a token that's a copy of target attacking creature and that's tapped and attacking. Exile the token at end of combat.",
		Context{CardName: "Unrelated Creature"},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	create := document.Abilities[0].Sentences[0].Effects[0]
	if !create.TokenCopyOfTarget || !create.TokenCopyAttacksWithTarget {
		t.Fatalf("create = %#v", create)
	}
}
