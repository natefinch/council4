package cardgen

import (
	"testing"
)

// TestResolvingGateEventPlayerConsequenceReferenceStrip locks the fix that lets
// an affirmative resolving gate's consequence reference the event player without
// failing closed. The gate clause ("If that player does") carries its own "that
// player" anaphor inside the consequence's clause span; that reference belongs to
// the consumed gate, not the consequence body, so the ordered-sequence lowering
// must strip it before lowering the consequence. Otherwise "they lose 2 life" or
// "you draw a card" would see a phantom second reference and fail closed. Each
// case previously failed with an "unsupported ordered effect sequence" sub-effect
// diagnostic; the "you lose"/"that player loses" spellings, which never leaked a
// second reference, are included to prove the fix did not regress them.
func TestResolvingGateEventPlayerConsequenceReferenceStrip(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		oracle string
	}{
		{
			name:   "they lose life",
			oracle: "Target opponent sacrifices a creature of their choice. If that player does, they lose 2 life.",
		},
		{
			name:   "you draw",
			oracle: "Target opponent sacrifices a creature of their choice. If that player does, you draw a card.",
		},
		{
			name:   "you lose life",
			oracle: "Target opponent sacrifices a creature of their choice. If that player does, you lose 2 life.",
		},
		{
			name:   "that player loses life",
			oracle: "Target opponent sacrifices a creature of their choice. If that player does, that player loses 2 life.",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			_, diagnostics := lowerExecutableFaces(&ScryfallCard{
				Name:       "X",
				Layout:     "normal",
				TypeLine:   "Sorcery",
				OracleText: test.oracle,
			})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
		})
	}
}
