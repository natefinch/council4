package compiler

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
)

func TestCompileSelfCannotBlockStaticAbility(t *testing.T) {
	t.Parallel()
	source := "This creature can't block."
	compilation, diagnostics := compileSource(source, pipelineContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}

	ability := compilation.Abilities[0]
	if ability.Kind != AbilityStatic ||
		len(ability.Content.Effects) != 1 ||
		ability.Content.Effects[0].Kind != EffectCantBlock ||
		!ability.Content.Effects[0].Negated {
		t.Fatalf("ability = %#v", ability)
	}
	if len(ability.Content.References) != 1 ||
		ability.Content.References[0].Kind != ReferenceThisObject {
		t.Fatalf("references = %#v", ability.Content.References)
	}
}

func TestCompileSelfCannotBeBlockedStaticAbility(t *testing.T) {
	t.Parallel()
	source := "This creature can't be blocked."
	compilation, diagnostics := compileSource(source, pipelineContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}

	ability := compilation.Abilities[0]
	if ability.Kind != AbilityStatic ||
		len(ability.Content.Effects) != 1 ||
		ability.Content.Effects[0].Kind != EffectCantBeBlocked ||
		!ability.Content.Effects[0].Negated {
		t.Fatalf("ability = %#v", ability)
	}
	if len(ability.Content.References) != 1 ||
		ability.Content.References[0].Kind != ReferenceThisObject {
		t.Fatalf("references = %#v", ability.Content.References)
	}
}

func TestCompileSelfMustAttackStaticAbility(t *testing.T) {
	t.Parallel()
	source := "This creature attacks each combat if able."
	compilation, diagnostics := compileSource(source, pipelineContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}

	ability := compilation.Abilities[0]
	if ability.Kind != AbilityStatic ||
		len(ability.Content.Effects) != 1 ||
		ability.Content.Effects[0].Kind != EffectMustAttack ||
		ability.Content.Effects[0].Negated {
		t.Fatalf("ability = %#v", ability)
	}
	if len(ability.Content.References) != 1 ||
		ability.Content.References[0].Kind != ReferenceThisObject {
		t.Fatalf("references = %#v", ability.Content.References)
	}
	if len(ability.Content.Conditions) != 0 {
		t.Fatalf("intrinsic if-able text became a separate condition: %#v", ability.Content.Conditions)
	}
}

func TestCompileSelfUncounterableStaticAbility(t *testing.T) {
	t.Parallel()
	source := "This spell can't be countered."
	compilation, diagnostics := compileSource(source, pipelineContext{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}

	ability := compilation.Abilities[0]
	if ability.Kind != AbilityStatic ||
		len(ability.Content.Effects) != 1 ||
		ability.Content.Effects[0].Kind != EffectCantBeCountered ||
		!ability.Content.Effects[0].Negated {
		t.Fatalf("ability = %#v", ability)
	}
	if len(ability.Content.References) != 1 ||
		ability.Content.References[0].Kind != ReferenceThisObject {
		t.Fatalf("references = %#v", ability.Content.References)
	}
}

func TestCompileComposedSimpleStaticRuleWordingVariants(t *testing.T) {
	t.Parallel()
	tests := map[string]StaticRuleKind{
		"This creature cannot block.":                    StaticRuleCantBlock,
		"This creature cannot be blocked.":               StaticRuleCantBeBlocked,
		"This creature must attack each combat if able.": StaticRuleMustAttack,
		"This spell cannot be countered.":                StaticRuleCantBeCountered,
	}
	for source, want := range tests {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := compileSource(source, pipelineContext{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			ability := compilation.Abilities[0]
			if ability.Static == nil ||
				len(ability.Static.Declarations) != 1 ||
				ability.Static.Declarations[0].Rule == nil ||
				ability.Static.Declarations[0].Rule.Kind != want {
				t.Fatalf("static semantics = %#v, want rule %v", ability.Static, want)
			}
		})
	}
}

func TestCompileConstructedTypedStaticRulesWithoutOracleWording(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		syntax parser.StaticRuleSyntax
		want   StaticRuleKind
		zone   StaticZone
	}{
		"active block prohibition": {
			syntax: parser.StaticRuleSyntax{
				Subject:    parser.StaticRuleSubject{Kind: parser.StaticRuleSubjectSourceCreature},
				Constraint: parser.StaticRuleConstraint{Kind: parser.StaticRuleConstraintProhibition},
				Operation:  parser.StaticRuleOperation{Kind: parser.StaticRuleOperationBlock, Voice: parser.StaticRuleVoiceActive},
			},
			want: StaticRuleCantBlock,
			zone: StaticZoneBattlefield,
		},
		"passive block prohibition": {
			syntax: parser.StaticRuleSyntax{
				Subject:    parser.StaticRuleSubject{Kind: parser.StaticRuleSubjectSourceCreature},
				Constraint: parser.StaticRuleConstraint{Kind: parser.StaticRuleConstraintProhibition},
				Operation:  parser.StaticRuleOperation{Kind: parser.StaticRuleOperationBlock, Voice: parser.StaticRuleVoicePassive},
			},
			want: StaticRuleCantBeBlocked,
			zone: StaticZoneBattlefield,
		},
		"attack requirement": {
			syntax: parser.StaticRuleSyntax{
				Subject:    parser.StaticRuleSubject{Kind: parser.StaticRuleSubjectSourceCreature},
				Constraint: parser.StaticRuleConstraint{Kind: parser.StaticRuleConstraintRequirement},
				Operation:  parser.StaticRuleOperation{Kind: parser.StaticRuleOperationAttack, Voice: parser.StaticRuleVoiceActive},
				Qualifiers: []parser.StaticRuleQualifier{
					{Kind: parser.StaticRuleQualifierEachCombat},
					{Kind: parser.StaticRuleQualifierIfAble},
				},
			},
			want: StaticRuleMustAttack,
			zone: StaticZoneBattlefield,
		},
		"passive counter prohibition": {
			syntax: parser.StaticRuleSyntax{
				Subject:    parser.StaticRuleSubject{Kind: parser.StaticRuleSubjectSourceSpell},
				Constraint: parser.StaticRuleConstraint{Kind: parser.StaticRuleConstraintProhibition},
				Operation:  parser.StaticRuleOperation{Kind: parser.StaticRuleOperationCounter, Voice: parser.StaticRuleVoicePassive},
			},
			want: StaticRuleCantBeCountered,
			zone: StaticZoneStack,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			document := parser.Document{
				Source: "unrelated source metadata",
				Abilities: []parser.Ability{{
					Kind: parser.AbilityStatic,
					Text: "not Oracle wording",
					Sentences: []parser.Sentence{{
						Text:       "also not Oracle wording",
						StaticRule: &test.syntax,
					}},
				}},
			}
			compilation, diagnostics := Compile(document, Context{})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			ability := compilation.Abilities[0]
			if ability.Static == nil || len(ability.Static.Declarations) != 1 {
				t.Fatalf("static semantics = %#v, want one declaration", ability.Static)
			}
			declaration := ability.Static.Declarations[0]
			if declaration.Rule == nil ||
				declaration.Rule.Kind != test.want ||
				declaration.Rule.Zone != test.zone ||
				declaration.Group.Domain != StaticGroupSource {
				t.Fatalf("declaration = %#v", declaration)
			}
		})
	}
}

func TestCompileSimpleStaticRuleNearMissesFailClosed(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"This creature attacks each combat.",
		"This creature must attack if able.",
		"This spell can't be countered by spells.",
	} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := compileSource(source, pipelineContext{InstantOrSorcery: true})
			if len(diagnostics) == 0 || diagnostics[0].Span != compilation.Syntax.Abilities[0].Span {
				t.Fatalf("diagnostics = %#v, want source-spanned unsupported diagnostic", diagnostics)
			}
			if static := compilation.Abilities[0].Static; static != nil && len(static.Declarations) != 0 {
				t.Fatalf("static semantics = %#v, want no declarations", static)
			}
		})
	}
}

func TestCompileStaticPTBuffSubjects(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		source          string
		wantSubject     StaticSubjectKind
		wantSubjectText string
		wantPower       CompiledSignedAmount
		wantToughness   CompiledSignedAmount
	}{
		"enchanted creature": {
			source:          "Enchanted creature gets +2/+2.",
			wantSubject:     StaticSubjectAttachedObject,
			wantSubjectText: "Enchanted creature",
			wantPower:       CompiledSignedAmount{Value: 2, Known: true},
			wantToughness:   CompiledSignedAmount{Value: 2, Known: true},
		},
		"equipped creature": {
			source:          "Equipped creature gets -3/-1.",
			wantSubject:     StaticSubjectAttachedObject,
			wantSubjectText: "Equipped creature",
			wantPower:       CompiledSignedAmount{Value: 3, Known: true, Negative: true},
			wantToughness:   CompiledSignedAmount{Value: 1, Known: true, Negative: true},
		},
		"other creatures you control": {
			source:          "Other creatures you control get +1/+1.",
			wantSubject:     StaticSubjectOtherControlledCreatures,
			wantSubjectText: "Other creatures you control",
			wantPower:       CompiledSignedAmount{Value: 1, Known: true},
			wantToughness:   CompiledSignedAmount{Value: 1, Known: true},
		},
		"creatures you control": {
			source:          "Creatures you control get +0/+2.",
			wantSubject:     StaticSubjectControlledCreatures,
			wantSubjectText: "Creatures you control",
			wantPower:       CompiledSignedAmount{Value: 0, Known: true},
			wantToughness:   CompiledSignedAmount{Value: 2, Known: true},
		},
		"each wall you control": {
			source:          "Each Wall you control gets +0/+2.",
			wantSubject:     StaticSubjectControlledWalls,
			wantSubjectText: "Each Wall you control",
			wantPower:       CompiledSignedAmount{Value: 0, Known: true},
			wantToughness:   CompiledSignedAmount{Value: 2, Known: true},
		},
		"artifacts you control": {
			source:          "Artifacts you control get +1/+1.",
			wantSubject:     StaticSubjectControlledArtifacts,
			wantSubjectText: "Artifacts you control",
			wantPower:       CompiledSignedAmount{Value: 1, Known: true},
			wantToughness:   CompiledSignedAmount{Value: 1, Known: true},
		},
		"tokens you control": {
			source:          "Tokens you control get +1/+1.",
			wantSubject:     StaticSubjectControlledTokens,
			wantSubjectText: "Tokens you control",
			wantPower:       CompiledSignedAmount{Value: 1, Known: true},
			wantToughness:   CompiledSignedAmount{Value: 1, Known: true},
		},
		"creatures your opponents control": {
			source:          "Creatures your opponents control get -1/-0.",
			wantSubject:     StaticSubjectOpponentControlledCreatures,
			wantSubjectText: "Creatures your opponents control",
			wantPower:       CompiledSignedAmount{Value: 1, Known: true, Negative: true},
			wantToughness:   CompiledSignedAmount{Value: 0, Known: true, Negative: true},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := compileSource(test.source, pipelineContext{})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			if len(compilation.Abilities) != 1 {
				t.Fatalf("abilities = %d, want 1", len(compilation.Abilities))
			}
			ability := compilation.Abilities[0]
			if len(ability.Content.Effects) != 1 || ability.Content.Effects[0].Kind != EffectModifyPT {
				t.Fatalf("effects = %#v", ability.Content.Effects)
			}
			effect := ability.Content.Effects[0]
			if effect.StaticSubject != test.wantSubject {
				t.Fatalf("static subject = %v, want %v", effect.StaticSubject, test.wantSubject)
			}
			if got := test.source[effect.StaticSubjectSpan.Start.Offset:effect.StaticSubjectSpan.End.Offset]; got != test.wantSubjectText {
				t.Fatalf("subject span text = %q, want %q", got, test.wantSubjectText)
			}
			if effect.PowerDelta != test.wantPower || effect.ToughnessDelta != test.wantToughness {
				t.Fatalf("PT = %+v / %+v, want %+v / %+v", effect.PowerDelta, effect.ToughnessDelta, test.wantPower, test.wantToughness)
			}
		})
	}
}

func TestCompileStaticKeywordGrantSubjects(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		source             string
		wantSubject        StaticSubjectKind
		wantSubjectSubtype string
		keywords           []string
	}{
		"enchanted creature": {
			source:      "Enchanted creature has menace.",
			wantSubject: StaticSubjectAttachedObject,
			keywords:    []string{"Menace"},
		},
		"equipped creature": {
			source:      "Equipped creature has flying and first strike.",
			wantSubject: StaticSubjectAttachedObject,
			keywords:    []string{"Flying", "First strike"},
		},
		"double strike": {
			source:      "Equipped creature has double strike.",
			wantSubject: StaticSubjectAttachedObject,
			keywords:    []string{"Double strike"},
		},
		"other creatures": {
			source:      "Other creatures you control have flying.",
			wantSubject: StaticSubjectOtherControlledCreatures,
			keywords:    []string{"Flying"},
		},
		"controlled creatures": {
			source:      "Creatures you control have haste.",
			wantSubject: StaticSubjectControlledCreatures,
			keywords:    []string{"Haste"},
		},
		"controlled artifacts": {
			source:      "Artifacts you control have indestructible.",
			wantSubject: StaticSubjectControlledArtifacts,
			keywords:    []string{"Indestructible"},
		},
		"controlled subtype": {
			source:             "Zombies you control have flying.",
			wantSubject:        StaticSubjectControlledCreatureSubtype,
			wantSubjectSubtype: "Zombies",
			keywords:           []string{"Flying"},
		},
		"other controlled subtype": {
			source:             "Other Dinosaurs you control have haste.",
			wantSubject:        StaticSubjectOtherControlledCreatureSubtype,
			wantSubjectSubtype: "Dinosaurs",
			keywords:           []string{"Haste"},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := compileSource(test.source, pipelineContext{})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			ability := compilation.Abilities[0]
			if len(ability.Content.Effects) != 1 || ability.Content.Effects[0].Kind != EffectGrantKeyword {
				t.Fatalf("effects = %#v", ability.Content.Effects)
			}
			if got := ability.Content.Effects[0].StaticSubject; got != test.wantSubject {
				t.Fatalf("static subject = %v, want %v", got, test.wantSubject)
			}
			if got := ability.Content.Effects[0].StaticSubjectSubtype(); got != test.wantSubjectSubtype {
				t.Fatalf("static subject subtype = %q, want %q", got, test.wantSubjectSubtype)
			}
			if len(ability.Content.Keywords) != len(test.keywords) {
				t.Fatalf("keywords = %#v, want %v", ability.Content.Keywords, test.keywords)
			}
			for i, keyword := range ability.Content.Keywords {
				if keyword.Name != test.keywords[i] {
					t.Fatalf("keyword %d = %q, want %q", i, keyword.Name, test.keywords[i])
				}
			}
		})
	}
}

func TestCompileStaticPTBuffWithKeywordHasOneEffect(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource(
		"Creatures you control get +1/+1 and have vigilance.",
		pipelineContext{},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := compilation.Abilities[0]
	if len(ability.Content.Effects) != 1 || ability.Content.Effects[0].Kind != EffectModifyPT {
		t.Fatalf("effects = %#v", ability.Content.Effects)
	}
}

func TestCompileStaticDeclarationsCarryClosedGroupSelectionAndLayer(t *testing.T) {
	t.Parallel()
	source := "Creatures your opponents control get -1/-0."
	compilation, diagnostics := compileSource(source, pipelineContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := compilation.Abilities[0]
	if ability.Static == nil || len(ability.Static.Declarations) != 1 {
		t.Fatalf("static semantics = %#v, want one declaration", ability.Static)
	}
	declaration := ability.Static.Declarations[0]
	if declaration.Kind != StaticDeclarationContinuous ||
		declaration.Continuous.Layer != StaticLayerPowerToughnessModify ||
		declaration.Continuous.Operation != StaticContinuousModifyPowerToughness {
		t.Fatalf("declaration = %#v, want power/toughness continuous declaration", declaration)
	}
	if declaration.Group.Domain != StaticGroupBattlefield ||
		declaration.Group.Selection.Controller != ControllerOpponent ||
		!slices.Equal(declaration.Group.Selection.RequiredTypes, []StaticCardType{StaticCardTypeCreature}) {
		t.Fatalf("group = %#v, want opponent-controlled battlefield creatures", declaration.Group)
	}
	if got := source[declaration.Group.Span.Start.Offset:declaration.Group.Span.End.Offset]; got != "Creatures your opponents control" {
		t.Fatalf("group span = %q", got)
	}
}

func TestCompileStaticControlGrantDeclaration(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"You control enchanted creature.",
		"You control enchanted permanent.",
	} {
		compilation, diagnostics := compileSource(source, pipelineContext{})
		if len(diagnostics) != 0 {
			t.Fatalf("%q diagnostics = %#v", source, diagnostics)
		}
		ability := compilation.Abilities[0]
		if ability.Static == nil || len(ability.Static.Declarations) != 1 {
			t.Fatalf("%q static semantics = %#v, want one declaration", source, ability.Static)
		}
		declaration := ability.Static.Declarations[0]
		if declaration.Kind != StaticDeclarationContinuous ||
			declaration.Continuous == nil ||
			declaration.Continuous.Layer != StaticLayerControl ||
			declaration.Continuous.Operation != StaticContinuousChangeControl {
			t.Fatalf("%q declaration = %#v, want control-change continuous declaration", source, declaration)
		}
		if declaration.Group.Domain != StaticGroupAttachedObject {
			t.Fatalf("%q group = %#v, want attached-object group", source, declaration.Group)
		}
	}
}

func TestCompileStaticControlGrantDeclarationsCarryClosedGroupSelectionAndLayer(t *testing.T) {
	t.Parallel()
	source := "You control enchanted creature."
	compilation, diagnostics := compileSource(source, pipelineContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	declaration := compilation.Abilities[0].Static.Declarations[0]
	if got := source[declaration.Group.Span.Start.Offset:declaration.Group.Span.End.Offset]; got != "enchanted creature" {
		t.Fatalf("group span = %q", got)
	}
}

func TestCompileStaticControlGrantOnAttachmentStateDurationFailsClosed(t *testing.T) {
	t.Parallel()
	// "for as long as that creature is enchanted" is an attachment-state
	// duration; #225/#324 keep it fail-closed rather than treating it as a
	// static source-tied control grant.
	source := "You control enchanted creature for as long as that creature is enchanted."
	_, diagnostics := compileSource(source, pipelineContext{})
	if len(diagnostics) == 0 {
		t.Fatal("expected the attachment-state duration wording to fail closed")
	}
}

func TestCompileStaticDeclarationsCarryConditionsAndRuleDomains(t *testing.T) {
	t.Parallel()
	source := "As long as you control an artifact, this creature has flying."
	compilation, diagnostics := compileSource(source, pipelineContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	declaration := compilation.Abilities[0].Static.Declarations[0]
	if declaration.Group.Domain != StaticGroupSource ||
		declaration.Condition == nil ||
		declaration.Condition.Predicate != ConditionPredicateControllerControls {
		t.Fatalf("declaration = %#v, want conditional source declaration", declaration)
	}
	if declaration.Continuous.Layer != StaticLayerAbility ||
		declaration.Continuous.Operation != StaticContinuousGrantKeywords {
		t.Fatalf("continuous declaration = %#v", declaration.Continuous)
	}

	compilation, diagnostics = compileSource("This spell can't be countered.", pipelineContext{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	declaration = compilation.Abilities[0].Static.Declarations[0]
	if declaration.Kind != StaticDeclarationRule ||
		declaration.Rule.Domain != StaticRuleDomainCountering ||
		declaration.Rule.Kind != StaticRuleCantBeCountered ||
		declaration.Rule.Zone != StaticZoneStack {
		t.Fatalf("rule declaration = %#v", declaration)
	}
}

func TestCompileMixedStaticParagraphProducesExactDeclarations(t *testing.T) {
	t.Parallel()
	source := "Delirium — As long as there are four or more card types among cards in your graveyard, Dragon's Rage Channeler gets +2/+2, has flying, and attacks each combat if able."
	compilation, diagnostics := compileSource(source, pipelineContext{CardName: "Dragon's Rage Channeler"})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := compilation.Abilities[0]
	if ability.Static == nil || len(ability.Static.Declarations) != 3 {
		t.Fatalf("static semantics = %#v, want three declarations", ability.Static)
	}
	if ability.Static.Declarations[0].Continuous.Layer != StaticLayerPowerToughnessModify ||
		ability.Static.Declarations[1].Continuous.Layer != StaticLayerAbility ||
		ability.Static.Declarations[2].Rule.Domain != StaticRuleDomainAttack ||
		ability.Static.Declarations[2].Rule.Kind != StaticRuleMustAttack {
		t.Fatalf("static declarations = %#v", ability.Static.Declarations)
	}
	for i, declaration := range ability.Static.Declarations {
		if declaration.Group.Domain != StaticGroupSource || declaration.Condition == nil {
			t.Fatalf("declaration %d = %#v, want conditional source declaration", i, declaration)
		}
		if declaration.Span.Start.Offset != 0 || declaration.Span.End.Offset != len(source) {
			t.Fatalf("declaration %d span = %#v, want entire paragraph", i, declaration.Span)
		}
	}
}

func TestCompileStaticDeclarationsFailClosedOnAdjacentSemantics(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		source  string
		blocker StaticDeclarationBlocker
	}{
		"duration": {
			source:  "Creatures you control get +1/+1 until end of turn.",
			blocker: StaticDeclarationBlockerDuration,
		},
		"condition": {
			source:  "As long as the moon is full, creatures you control get +1/+1.",
			blocker: StaticDeclarationBlockerCondition,
		},
		"group": {
			source:  "All creatures get +1/+1.",
			blocker: StaticDeclarationBlockerGroup,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := compileSource(test.source, pipelineContext{})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			ability := compilation.Abilities[0]
			if ability.Static == nil {
				t.Fatal("static semantics = nil, want blocker")
			}
			if len(ability.Static.Declarations) != 0 {
				t.Fatalf("static declarations = %#v, want none", ability.Static.Declarations)
			}
			if ability.Static.Blocker != test.blocker {
				t.Fatalf("static blocker = %v, want %v", ability.Static.Blocker, test.blocker)
			}
		})
	}
}

func TestCompileResolvingPTBuffHasNoStaticSubject(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource(
		"Target creature gets +2/+2 until end of turn.",
		pipelineContext{InstantOrSorcery: true},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	effect := compilation.Abilities[0].Content.Effects[0]
	if effect.StaticSubject != StaticSubjectNone {
		t.Fatalf("static subject = %v, want StaticSubjectNone", effect.StaticSubject)
	}
	if effect.StaticSubjectSpan != (shared.Span{}) {
		t.Fatalf("static subject span = %#v, want zero span", effect.StaticSubjectSpan)
	}
}
