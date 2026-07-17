package parser

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/types"
)

const nightmareShepherdOracle = "Flying\nWhenever another nontoken creature you control dies, you may exile it. If you do, create a token that's a copy of that creature, except it's 1/1 and it's a Nightmare in addition to its other types."

func TestParseNightmareShepherdReusableMechanics(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(nightmareShepherdOracle, Context{CardName: "Nightmare Shepherd"})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(document.Abilities) != 2 {
		t.Fatalf("abilities = %d, want Flying and dies trigger", len(document.Abilities))
	}
	ability := document.Abilities[1]
	var effects []EffectSyntax
	for _, sentence := range ability.Sentences {
		effects = append(effects, sentence.Effects...)
	}
	if len(effects) != 2 {
		t.Fatalf("effects = %#v, want exile and token copy", effects)
	}
	if effects[0].Kind != EffectExile || !effects[0].Optional {
		t.Fatalf("exile effect = %#v, want optional exile", effects[0])
	}
	copyEffect := effects[1]
	if copyEffect.Kind != EffectCreate || !copyEffect.TokenCopyOfReference ||
		!copyEffect.TokenCopyOverride || !copyEffect.TokenCopyOverridePTKnown ||
		copyEffect.TokenCopyOverridePower != 1 || copyEffect.TokenCopyOverrideToughness != 1 ||
		!copyEffect.TokenCopyOverrideAdditiveTypes ||
		len(copyEffect.TokenCopyOverrideSubtypes) != 1 ||
		copyEffect.TokenCopyOverrideSubtypes[0] != types.Nightmare {
		t.Fatalf("copy effect = %#v", copyEffect)
	}
}
