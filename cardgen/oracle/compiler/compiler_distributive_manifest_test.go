package compiler

import "testing"

func TestCompileDistributiveExileCloak(t *testing.T) {
	t.Parallel()
	const source = "For each player, exile up to one target nonland permanent that player controls. For each permanent exiled this way, its controller cloaks the top card of their library."
	compilation, diagnostics := compileSource(source, pipelineContext{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	effects := compilation.Abilities[0].Content.Effects
	if len(effects) != 2 ||
		!effects[0].Exact ||
		!effects[0].ExileForEachPlayer ||
		!effects[1].Exact ||
		!effects[1].CloakForEachExiledThisWay {
		t.Fatalf("effects = %#v", effects)
	}
}
