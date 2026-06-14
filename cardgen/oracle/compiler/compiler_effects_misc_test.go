package compiler

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/zone"
)

func TestCompileReturnToOwnersHand(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource(
		"Return target creature to its owner's hand.",
		pipelineContext{InstantOrSorcery: true},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := compilation.Abilities[0]
	if len(ability.Content.Effects) != 1 || ability.Content.Effects[0].Kind != EffectReturn {
		t.Fatalf("effects = %#v", ability.Content.Effects)
	}
	if len(ability.Content.Targets) != 1 ||
		ability.Content.Targets[0].Selector.Kind != SelectorCreature ||
		ability.Content.Targets[0].Text != "target creature" {
		t.Fatalf("targets = %#v", ability.Content.Targets)
	}
	if len(ability.Content.References) != 1 ||
		ability.Content.References[0].Kind != ReferencePronoun ||
		ability.Content.References[0].Text != "its" {
		t.Fatalf("references = %#v", ability.Content.References)
	}
	if len(ability.Content.Conditions) != 0 ||
		len(ability.Content.Keywords) != 0 ||
		len(ability.Content.Modes) != 0 ||
		ability.Content.Effects[0].Negated ||
		ability.Content.Targets[0].Cardinality.Min != 1 ||
		ability.Content.Targets[0].Cardinality.Max != 1 {
		t.Fatalf("ability = %#v", ability)
	}
}

func TestCompileGraveyardReturnZones(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		text     string
		fromZone zone.Type
		toZone   zone.Type
	}{
		{
			name:     "target card to hand",
			text:     "Return target instant or sorcery card from your graveyard to your hand.",
			fromZone: zone.Graveyard,
			toZone:   zone.Hand,
		},
		{
			name:     "target card to library",
			text:     "Put target card from your graveyard on the bottom of your library.",
			fromZone: zone.Graveyard,
			toZone:   zone.Library,
		},
		{
			name:     "opponents graveyard",
			text:     "Return target creature card from an opponent's graveyard to your hand.",
			fromZone: zone.Graveyard,
			toZone:   zone.Hand,
		},
		{
			name:     "self to battlefield",
			text:     "Return this card from your graveyard to the battlefield tapped.",
			fromZone: zone.Graveyard,
			toZone:   zone.Battlefield,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := compileSource(tc.text, pipelineContext{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			ability := compilation.Abilities[0]
			if len(ability.Content.Effects) != 1 {
				t.Fatalf("effects = %#v", ability.Content.Effects)
			}
			effect := ability.Content.Effects[0]
			if effect.FromZone != tc.fromZone || effect.ToZone != tc.toZone {
				t.Fatalf("zones = %v -> %v, want %v -> %v", effect.FromZone, effect.ToZone, tc.fromZone, tc.toZone)
			}
		})
	}
}

func TestCompileSurveilEffect(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource("Surveil 2.", pipelineContext{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	effects := compilation.Abilities[0].Content.Effects
	if len(effects) != 1 ||
		effects[0].Kind != EffectSurveil ||
		effects[0].Amount.Value != 2 ||
		!effects[0].Amount.Known {
		t.Fatalf("effects = %#v, want surveil 2", effects)
	}
}

func TestCompileInvestigateEffect(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource("Investigate.", pipelineContext{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	effects := compilation.Abilities[0].Content.Effects
	if len(effects) != 1 || effects[0].Kind != EffectInvestigate {
		t.Fatalf("effects = %#v, want investigate", effects)
	}
}

func TestCompileProliferateEffect(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource("Proliferate.", pipelineContext{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	effects := compilation.Abilities[0].Content.Effects
	if len(effects) != 1 || effects[0].Kind != EffectProliferate {
		t.Fatalf("effects = %#v, want proliferate", effects)
	}
}

func TestCompileRegenerateEffect(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource(
		"Regenerate target creature.",
		pipelineContext{InstantOrSorcery: true},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	effects := compilation.Abilities[0].Content.Effects
	if len(effects) != 1 || effects[0].Kind != EffectRegenerate {
		t.Fatalf("effects = %#v, want regenerate", effects)
	}
}

func TestCompileCounterVerbAndNoun(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		wantKinds []EffectKind
	}{
		"Counter target spell.": {
			wantKinds: []EffectKind{EffectCounter},
		},
		"This spell counters target spell.": {
			wantKinds: []EffectKind{EffectCounter},
		},
		"Put two +1/+1 counters on target creature.": {
			wantKinds: []EffectKind{EffectPut},
		},
		"Remove a counter from this permanent: Draw a card.": {
			wantKinds: []EffectKind{EffectDraw},
		},
	}
	for source, test := range tests {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := compileSource(source, pipelineContext{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			effects := compilation.Abilities[0].Content.Effects
			if len(effects) != len(test.wantKinds) {
				t.Fatalf("effects = %#v, want kinds %v", effects, test.wantKinds)
			}
			for i, want := range test.wantKinds {
				if effects[i].Kind != want {
					t.Fatalf("effect %d = %v, want %v", i, effects[i].Kind, want)
				}
			}
		})
	}
}

func TestCompileExactCounterAbilityTargets(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source string
		text   string
		kind   SelectorKind
	}{
		{"Counter target activated ability.", "target activated ability", SelectorActivatedAbility},
		{"Counter target triggered ability.", "target triggered ability", SelectorTriggeredAbility},
		{"Counter target activated or triggered ability.", "target activated or triggered ability", SelectorActivatedOrTriggeredAbility},
		{"Counter target spell, activated ability, or triggered ability.", "target spell, activated ability, or triggered ability", SelectorSpellActivatedOrTriggeredAbility},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := compileSource(test.source, pipelineContext{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			targets := compilation.Abilities[0].Content.Targets
			if len(targets) != 1 || targets[0].Text != test.text || targets[0].Selector.Kind != test.kind {
				t.Fatalf("targets = %#v, want text %q kind %v", targets, test.text, test.kind)
			}
		})
	}
}

func TestCompileNegatedEffect(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource("Players can't gain life.", pipelineContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	effects := compilation.Abilities[0].Content.Effects
	if len(effects) != 1 || effects[0].Kind != EffectGain || !effects[0].Negated {
		t.Fatalf("effects = %#v", effects)
	}
}
