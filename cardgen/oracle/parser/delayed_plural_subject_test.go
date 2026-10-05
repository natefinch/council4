package parser

import "testing"

func TestDelayedImplicitPluralRequiresOneProducer(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name    string
		effects []*EffectSyntax
		kind    DelayedSubjectKind
		clause  int
	}{
		{"one actual batch", []*EffectSyntax{{Kind: EffectCreate, ClauseID: 4}, {Kind: EffectGain}}, DelayedSubjectProduct, 4},
		{"two independent batches", []*EffectSyntax{{Kind: EffectCreate, ClauseID: 4}, {Kind: EffectGain}, {Kind: EffectCreate, ClauseID: 9}}, DelayedSubjectUnsupported, 0},
		{"new grammatical introduction", []*EffectSyntax{{Kind: EffectCreate, ClauseID: 4}, {Kind: EffectExile}, {Kind: EffectCreate, ClauseID: 9}}, DelayedSubjectProduct, 9},
		{"no actual batch", []*EffectSyntax{{Kind: EffectGain}}, DelayedSubjectUnsupported, 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := delayedProductSubject(test.effects)
			if got.Kind != test.kind || got.ProducerClauseID != test.clause {
				t.Fatalf("ownership=%+v; want kind=%v producer=%d", got, test.kind, test.clause)
			}
		})
	}
}
