package compiler

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
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
		"This creature can't attack.":                    StaticRuleCantAttack,
		"This creature must attack each combat if able.": StaticRuleMustAttack,
		"This creature must be blocked if able.":         StaticRuleMustBeBlocked,
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

func TestCompileStaticPermanentManaAbilityGrant(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource(
		`Lands you control have "{T}: Add one mana of any color."`,
		pipelineContext{},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := compilation.Abilities[0]
	if ability.Static == nil || len(ability.Static.Declarations) != 1 {
		t.Fatalf("static semantics = %#v, want one declaration", ability.Static)
	}
	declaration := ability.Static.Declarations[0]
	if declaration.Kind != StaticDeclarationContinuous ||
		declaration.Group.Domain != StaticGroupSourceControllerPermanents ||
		!slices.Equal(declaration.Group.Selection.RequiredTypes, []StaticCardType{StaticCardTypeLand}) ||
		declaration.Continuous == nil ||
		declaration.Continuous.Layer != StaticLayerAbility ||
		declaration.Continuous.Operation != StaticContinuousGrantManaAbility {
		t.Fatalf("declaration = %#v, want controlled-land mana-ability grant", declaration)
	}
	granted := declaration.Continuous.GrantedMana
	if granted == nil || !granted.TapCost || granted.Amount != 1 || !granted.AnyColor {
		t.Fatalf("granted mana ability = %#v, want tap for one mana of any color", granted)
	}
}

func TestCompileAttachedAndUntapStaticRules(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		source string
		rule   StaticRuleKind
		group  StaticGroupDomain
		zone   StaticZone
	}{
		"enchanted creature can't attack or block": {
			source: "Enchanted creature can't attack or block.",
			rule:   StaticRuleCantAttackOrBlock,
			group:  StaticGroupAttachedObject,
			zone:   StaticZoneBattlefield,
		},
		"equipped creature can't be blocked": {
			source: "Equipped creature can't be blocked.",
			rule:   StaticRuleCantBeBlocked,
			group:  StaticGroupAttachedObject,
			zone:   StaticZoneBattlefield,
		},
		"this creature can't attack or block": {
			source: "This creature can't attack or block.",
			rule:   StaticRuleCantAttackOrBlock,
			group:  StaticGroupSource,
			zone:   StaticZoneBattlefield,
		},
		"this creature doesn't untap": {
			source: "This creature doesn't untap during your untap step.",
			rule:   StaticRuleDoesntUntap,
			group:  StaticGroupSource,
			zone:   StaticZoneBattlefield,
		},
		"enchanted creature doesn't untap": {
			source: "Enchanted creature doesn't untap during its controller's untap step.",
			rule:   StaticRuleDoesntUntap,
			group:  StaticGroupAttachedObject,
			zone:   StaticZoneBattlefield,
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
			if ability.Static == nil || len(ability.Static.Declarations) != 1 {
				t.Fatalf("static semantics = %#v, want one declaration", ability.Static)
			}
			declaration := ability.Static.Declarations[0]
			if declaration.Rule == nil ||
				declaration.Rule.Kind != test.rule ||
				declaration.Rule.Zone != test.zone ||
				declaration.Rule.Domain != staticRuleDomain(test.rule) ||
				declaration.Group.Domain != test.group {
				t.Fatalf("declaration = %#v, want rule %v group %v", declaration, test.rule, test.group)
			}
		})
	}
}

func TestCompileAttachedAndUntapStaticRuleNearMissesFailClosed(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"Enchanted creature can't attack and block.",
		"Enchanted creature doesn't untap.",
		"Enchanted permanent doesn't untap during your untap step.",
		"Enchanted creature can't attack or block this turn.",
	} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := compileSource(source, pipelineContext{})
			static := compilation.Abilities[0].Static
			if static != nil {
				for _, declaration := range static.Declarations {
					if declaration.Rule != nil {
						t.Fatalf("declaration = %#v, want no static rule declaration (fail closed); diagnostics = %#v", declaration, diagnostics)
					}
				}
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
		"attack prohibition": {
			syntax: parser.StaticRuleSyntax{
				Subject:    parser.StaticRuleSubject{Kind: parser.StaticRuleSubjectSourceCreature},
				Constraint: parser.StaticRuleConstraint{Kind: parser.StaticRuleConstraintProhibition},
				Operation:  parser.StaticRuleOperation{Kind: parser.StaticRuleOperationAttack, Voice: parser.StaticRuleVoiceActive},
			},
			want: StaticRuleCantAttack,
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
		"block requirement": {
			syntax: parser.StaticRuleSyntax{
				Subject:    parser.StaticRuleSubject{Kind: parser.StaticRuleSubjectSourceCreature},
				Constraint: parser.StaticRuleConstraint{Kind: parser.StaticRuleConstraintRequirement},
				Operation:  parser.StaticRuleOperation{Kind: parser.StaticRuleOperationBlock, Voice: parser.StaticRuleVoicePassive},
				Qualifiers: []parser.StaticRuleQualifier{
					{Kind: parser.StaticRuleQualifierIfAble},
				},
			},
			want: StaticRuleMustBeBlocked,
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
		"permanent untap prohibition": {
			syntax: parser.StaticRuleSyntax{
				Subject:    parser.StaticRuleSubject{Kind: parser.StaticRuleSubjectSourcePermanent},
				Constraint: parser.StaticRuleConstraint{Kind: parser.StaticRuleConstraintProhibition},
				Operation:  parser.StaticRuleOperation{Kind: parser.StaticRuleOperationUntap, Voice: parser.StaticRuleVoiceActive},
			},
			want: StaticRuleDoesntUntap,
			zone: StaticZoneBattlefield,
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
		"This creature can't attack unless you control an artifact.",
		"This creature must be blocked.",
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
		"controlled subtype with creatures noun": {
			source:          "Sliver creatures you control get +2/+0.",
			wantSubject:     StaticSubjectControlledCreatureSubtype,
			wantSubjectText: "Sliver creatures you control",
			wantPower:       CompiledSignedAmount{Value: 2, Known: true},
			wantToughness:   CompiledSignedAmount{Value: 0, Known: true},
		},
		"other controlled subtype with creatures noun": {
			source:          "Other Goblin creatures you control get +1/+1.",
			wantSubject:     StaticSubjectOtherControlledCreatureSubtype,
			wantSubjectText: "Other Goblin creatures you control",
			wantPower:       CompiledSignedAmount{Value: 1, Known: true},
			wantToughness:   CompiledSignedAmount{Value: 1, Known: true},
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
		source              string
		wantSubject         StaticSubjectKind
		wantSubjectSubtype  string
		wantSubjectExcluded bool
		keywords            []string
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
		"excluded controlled subtype": {
			source:              "Non-Human creatures you control have trample.",
			wantSubject:         StaticSubjectControlledCreatureSubtype,
			wantSubjectSubtype:  "Non-Human",
			wantSubjectExcluded: true,
			keywords:            []string{"Trample"},
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
			if got := ability.Content.Effects[0].StaticSubjectSubExcluded(); got != test.wantSubjectExcluded {
				t.Fatalf("static subject excluded = %v, want %v", got, test.wantSubjectExcluded)
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

func TestCompileStaticGroupAnthemSubjects(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		source        string
		domain        StaticGroupDomain
		requireType   []StaticCardType
		subtypesAny   []types.Sub
		supertypes    []types.Super
		colorsAny     []color.Color
		combatState   StaticCombatState
		tapState      StaticTapState
		tokenOnly     bool
		excludeSource bool
	}{
		"all creatures": {
			source:      "All creatures get +1/+1.",
			domain:      StaticGroupBattlefield,
			requireType: []StaticCardType{StaticCardTypeCreature},
		},
		"all other creatures": {
			source:        "All other creatures get -1/-1.",
			domain:        StaticGroupBattlefield,
			requireType:   []StaticCardType{StaticCardTypeCreature},
			excludeSource: true,
		},
		"attacking creatures": {
			source:      "Attacking creatures get -1/-0.",
			domain:      StaticGroupBattlefield,
			requireType: []StaticCardType{StaticCardTypeCreature},
			combatState: StaticCombatStateAttacking,
		},
		"blocking creatures": {
			source:      "Blocking creatures get +0/+2.",
			domain:      StaticGroupBattlefield,
			requireType: []StaticCardType{StaticCardTypeCreature},
			combatState: StaticCombatStateBlocking,
		},
		"all subtype creatures": {
			source:      "All Sliver creatures get +1/+1.",
			domain:      StaticGroupBattlefield,
			subtypesAny: []types.Sub{types.Sliver},
		},
		"other subtype creatures": {
			source:        "Other Soldier creatures get +1/+1.",
			domain:        StaticGroupBattlefield,
			subtypesAny:   []types.Sub{types.Soldier},
			excludeSource: true,
		},
		"attacking creatures you control": {
			source:      "Attacking creatures you control get +1/+0.",
			domain:      StaticGroupSourceControllerPermanents,
			requireType: []StaticCardType{StaticCardTypeCreature},
			combatState: StaticCombatStateAttacking,
		},
		"controlled creature tokens": {
			source:      "Creature tokens you control get +1/+1.",
			domain:      StaticGroupSourceControllerPermanents,
			requireType: []StaticCardType{StaticCardTypeCreature},
			tokenOnly:   true,
		},
		"battlefield creature tokens": {
			source:      "Creature tokens get -1/-1.",
			domain:      StaticGroupBattlefield,
			requireType: []StaticCardType{StaticCardTypeCreature},
			tokenOnly:   true,
		},
		"controlled legendary creatures": {
			source:      "Legendary creatures you control get +2/+2.",
			domain:      StaticGroupSourceControllerPermanents,
			requireType: []StaticCardType{StaticCardTypeCreature},
			supertypes:  []types.Super{types.Legendary},
		},
		"controlled untapped creatures": {
			source:      "Untapped creatures you control get +0/+2.",
			domain:      StaticGroupSourceControllerPermanents,
			requireType: []StaticCardType{StaticCardTypeCreature},
			tapState:    StaticTapStateUntapped,
		},
		"other controlled tapped creatures": {
			source:        "Other tapped creatures you control have hexproof.",
			domain:        StaticGroupSourceControllerPermanents,
			requireType:   []StaticCardType{StaticCardTypeCreature},
			tapState:      StaticTapStateTapped,
			excludeSource: true,
		},
		"battlefield color creatures": {
			source:      "White creatures get +1/+1.",
			domain:      StaticGroupBattlefield,
			requireType: []StaticCardType{StaticCardTypeCreature},
			colorsAny:   []color.Color{color.White},
		},
		"battlefield other color creatures": {
			source:        "Other black creatures get -1/-1.",
			domain:        StaticGroupBattlefield,
			requireType:   []StaticCardType{StaticCardTypeCreature},
			colorsAny:     []color.Color{color.Black},
			excludeSource: true,
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
			if ability.Static == nil || len(ability.Static.Declarations) != 1 {
				t.Fatalf("static semantics = %#v, want one declaration", ability.Static)
			}
			group := ability.Static.Declarations[0].Group
			if group.Domain != test.domain ||
				group.ExcludeSource != test.excludeSource ||
				group.Selection.CombatState != test.combatState ||
				group.Selection.TapState != test.tapState ||
				group.Selection.TokenOnly != test.tokenOnly ||
				!slices.Equal(group.Selection.RequiredTypes, test.requireType) ||
				!slices.Equal(group.Selection.SubtypesAny, test.subtypesAny) ||
				!slices.Equal(group.Selection.Supertypes, test.supertypes) ||
				!slices.Equal(group.Selection.ColorsAny, test.colorsAny) {
				t.Fatalf("group = %#v", group)
			}
		})
	}
}

func TestCompileStaticGroupKeywordAndTypeFilterSelections(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		source          string
		domain          StaticGroupDomain
		requireType     []StaticCardType
		keyword         parser.KeywordKind
		excludedKeyword parser.KeywordKind
		nonToken        bool
		excludeSource   bool
	}{
		"creatures with flying": {
			source:      "Creatures with flying get +1/+1.",
			domain:      StaticGroupBattlefield,
			requireType: []StaticCardType{StaticCardTypeCreature},
			keyword:     parser.KeywordFlying,
		},
		"creatures without flying": {
			source:          "Creatures without flying get -2/-0.",
			domain:          StaticGroupBattlefield,
			requireType:     []StaticCardType{StaticCardTypeCreature},
			excludedKeyword: parser.KeywordFlying,
		},
		"creatures you control with flying": {
			source:      "Creatures you control with flying get +1/+1.",
			domain:      StaticGroupSourceControllerPermanents,
			requireType: []StaticCardType{StaticCardTypeCreature},
			keyword:     parser.KeywordFlying,
		},
		"other creatures you control with flying": {
			source:        "Other creatures you control with flying get +1/+1.",
			domain:        StaticGroupSourceControllerPermanents,
			requireType:   []StaticCardType{StaticCardTypeCreature},
			keyword:       parser.KeywordFlying,
			excludeSource: true,
		},
		"artifact creatures you control": {
			source:      "Artifact creatures you control get +1/+1.",
			domain:      StaticGroupSourceControllerPermanents,
			requireType: []StaticCardType{StaticCardTypeArtifact, StaticCardTypeCreature},
		},
		"nontoken creatures you control": {
			source:      "Nontoken creatures you control get +1/+1.",
			domain:      StaticGroupSourceControllerPermanents,
			requireType: []StaticCardType{StaticCardTypeCreature},
			nonToken:    true,
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
			if ability.Static == nil || len(ability.Static.Declarations) != 1 {
				t.Fatalf("static semantics = %#v, want one declaration", ability.Static)
			}
			group := ability.Static.Declarations[0].Group
			if group.Domain != test.domain ||
				group.ExcludeSource != test.excludeSource ||
				group.Selection.Keyword != test.keyword ||
				group.Selection.ExcludedKeyword != test.excludedKeyword ||
				group.Selection.NonToken != test.nonToken ||
				!slices.Equal(group.Selection.RequiredTypes, test.requireType) {
				t.Fatalf("group = %#v", group)
			}
		})
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

func TestCompileStaticNoMaximumHandSizeDeclaration(t *testing.T) {
	t.Parallel()
	source := "You have no maximum hand size."
	compilation, diagnostics := compileSource(source, pipelineContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := compilation.Abilities[0]
	if ability.Static == nil || len(ability.Static.Declarations) != 1 {
		t.Fatalf("static semantics = %#v, want one declaration", ability.Static)
	}
	declaration := ability.Static.Declarations[0]
	if declaration.Kind != StaticDeclarationPlayerRule ||
		declaration.Player == nil ||
		declaration.Player.Kind != StaticPlayerRuleNoMaximumHandSize {
		t.Fatalf("declaration = %#v, want no-maximum-hand-size player rule", declaration)
	}
	if declaration.Continuous != nil || declaration.Rule != nil || declaration.Cost != nil || declaration.CardGrant != nil {
		t.Fatalf("declaration carries an unexpected payload: %#v", declaration)
	}
}

func TestCompileStaticAttackTaxDeclaration(t *testing.T) {
	t.Parallel()
	source := "Creatures can't attack you unless their controller pays {2} for each creature they control that's attacking you."
	compilation, diagnostics := compileSource(source, pipelineContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v; ability = %#v", diagnostics, compilation.Abilities[0])
	}
	ability := compilation.Abilities[0]
	if ability.Static == nil || len(ability.Static.Declarations) != 1 {
		t.Fatalf("static semantics = %#v, want one declaration", ability.Static)
	}
	declaration := ability.Static.Declarations[0]
	if declaration.Kind != StaticDeclarationPlayerRule ||
		declaration.Player == nil ||
		declaration.Player.Kind != StaticPlayerRuleAttackTax ||
		declaration.Player.AttackTaxGeneric != 2 {
		t.Fatalf("declaration = %#v, want attack-tax player rule", declaration)
	}
	if declaration.Continuous != nil || declaration.Rule != nil || declaration.Cost != nil || declaration.CardGrant != nil {
		t.Fatalf("declaration carries an unexpected payload: %#v", declaration)
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

func TestCompileComposedQualifiedStaticRuleDeclarations(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		source  string
		layer   StaticContinuousLayer
		rule    StaticRuleKind
		domain  StaticRuleDomain
		group   StaticGroupDomain
		blocker StaticBlockerRestriction
	}{
		"power/toughness then can't attack you": {
			source: "Enchanted creature gets +2/+2 and can't attack you or planeswalkers you control.",
			layer:  StaticLayerPowerToughnessModify,
			rule:   StaticRuleCantAttackYou,
			domain: StaticRuleDomainAttack,
			group:  StaticGroupAttachedObject,
		},
		"power/toughness then can't be blocked by more than one": {
			source: "Enchanted creature gets +1/+2 and can't be blocked by more than one creature.",
			layer:  StaticLayerPowerToughnessModify,
			rule:   StaticRuleCantBeBlockedByMoreThanOne,
			domain: StaticRuleDomainBlock,
			group:  StaticGroupAttachedObject,
		},
		"keyword then can't be blocked by more than one": {
			source: "Enchanted creature has trample and can't be blocked by more than one creature.",
			layer:  StaticLayerAbility,
			rule:   StaticRuleCantBeBlockedByMoreThanOne,
			domain: StaticRuleDomainBlock,
			group:  StaticGroupAttachedObject,
		},
		"keyword then can't be blocked by flying": {
			source:  "Enchanted creature has trample and can't be blocked by creatures with flying.",
			layer:   StaticLayerAbility,
			rule:    StaticRuleCantBeBlockedByCreaturesWith,
			domain:  StaticRuleDomainBlock,
			group:   StaticGroupAttachedObject,
			blocker: StaticBlockerRestriction{Kind: StaticBlockerRestrictionFlying},
		},
		"power/toughness then can't be blocked by power 2 or less": {
			source:  "Enchanted creature gets +1/+2 and can't be blocked by creatures with power 2 or less.",
			layer:   StaticLayerPowerToughnessModify,
			rule:    StaticRuleCantBeBlockedByCreaturesWith,
			domain:  StaticRuleDomainBlock,
			group:   StaticGroupAttachedObject,
			blocker: StaticBlockerRestriction{Kind: StaticBlockerRestrictionPowerOrLess, Amount: 2},
		},
		"power/toughness then can't be blocked by power 3 or greater": {
			source:  "Enchanted creature gets +1/+2 and can't be blocked by creatures with power 3 or greater.",
			layer:   StaticLayerPowerToughnessModify,
			rule:    StaticRuleCantBeBlockedByCreaturesWith,
			domain:  StaticRuleDomainBlock,
			group:   StaticGroupAttachedObject,
			blocker: StaticBlockerRestriction{Kind: StaticBlockerRestrictionPowerOrGreater, Amount: 3},
		},
		"power/toughness then can't be blocked by color": {
			source:  "Enchanted creature gets +1/+2 and can't be blocked by black creatures.",
			layer:   StaticLayerPowerToughnessModify,
			rule:    StaticRuleCantBeBlockedByCreaturesWith,
			domain:  StaticRuleDomainBlock,
			group:   StaticGroupAttachedObject,
			blocker: StaticBlockerRestriction{Kind: StaticBlockerRestrictionColor, Color: color.Black},
		},
		"keyword then can't be blocked by artifact": {
			source:  "Enchanted creature has trample and can't be blocked by artifact creatures.",
			layer:   StaticLayerAbility,
			rule:    StaticRuleCantBeBlockedByCreaturesWith,
			domain:  StaticRuleDomainBlock,
			group:   StaticGroupAttachedObject,
			blocker: StaticBlockerRestriction{Kind: StaticBlockerRestrictionArtifact},
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
			if ability.Static == nil || len(ability.Static.Declarations) != 2 {
				t.Fatalf("static semantics = %#v, want two declarations", ability.Static)
			}
			if ability.Static.Declarations[0].Continuous.Layer != test.layer {
				t.Fatalf("first declaration = %#v, want layer %v", ability.Static.Declarations[0], test.layer)
			}
			rule := ability.Static.Declarations[1].Rule
			if rule == nil ||
				rule.Kind != test.rule ||
				rule.Domain != test.domain ||
				rule.Domain != staticRuleDomain(test.rule) ||
				rule.Zone != StaticZoneBattlefield {
				t.Fatalf("rule declaration = %#v, want rule %v domain %v", ability.Static.Declarations[1], test.rule, test.domain)
			}
			if rule.Blocker != test.blocker {
				t.Fatalf("rule blocker = %#v, want %#v", rule.Blocker, test.blocker)
			}
			for i, declaration := range ability.Static.Declarations {
				if declaration.Group.Domain != test.group {
					t.Fatalf("declaration %d group = %#v, want %v", i, declaration.Group, test.group)
				}
			}
		})
	}
}

func TestCompileComposedQualifiedStaticRuleNearMissesFailClosed(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"Enchanted creature gets +2/+2 and can't attack you.",
		"Enchanted creature gets +2/+2 and can't attack planeswalkers you control.",
		"Enchanted creature gets +1/+2 and can't be blocked by more than two creatures.",
		"Enchanted creature has trample and can't be blocked by creatures with toughness 2 or less.",
		"Enchanted creature has trample and can't be blocked by enormous creatures.",
		"Enchanted creature has trample and can't be blocked by artifacts.",
	} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			compilation, _ := compileSource(source, pipelineContext{})
			if static := compilation.Abilities[0].Static; static != nil {
				for _, declaration := range static.Declarations {
					if declaration.Rule != nil {
						t.Fatalf("declaration = %#v, want no static rule declaration (fail closed)", declaration)
					}
				}
			}
		})
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
			source:  "Creatures you control that are enchanted get +1/+1.",
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

func TestCompileStaticBasePowerToughnessDeclaration(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource(
		"Enchanted creature has base power and toughness 0/2.",
		pipelineContext{},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := compilation.Abilities[0]
	if ability.Static == nil || len(ability.Static.Declarations) != 1 {
		t.Fatalf("static semantics = %#v, want one declaration", ability.Static)
	}
	declaration := ability.Static.Declarations[0]
	if declaration.Kind != StaticDeclarationContinuous ||
		declaration.Continuous == nil ||
		declaration.Continuous.Layer != StaticLayerPowerToughnessSet ||
		declaration.Continuous.Operation != StaticContinuousSetBasePowerToughness ||
		declaration.Continuous.SetPower != 0 ||
		declaration.Continuous.SetToughness != 2 {
		t.Fatalf("declaration = %#v, want base 0/2 set declaration", declaration)
	}
	if declaration.Group.Domain != StaticGroupAttachedObject {
		t.Fatalf("group = %#v, want attached-object group", declaration.Group)
	}
}

func TestCompileStaticCharacteristicSetColorComposition(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource(
		"Enchanted creature gets +3/+1 and is black.",
		pipelineContext{},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	declarations := compilation.Abilities[0].Static.Declarations
	if len(declarations) != 2 {
		t.Fatalf("declarations = %#v, want two", declarations)
	}
	if declarations[0].Continuous.Layer != StaticLayerPowerToughnessModify {
		t.Fatalf("declarations[0] = %#v, want PT modify", declarations[0])
	}
	color1 := declarations[1].Continuous
	if color1.Layer != StaticLayerColor ||
		color1.Operation != StaticContinuousSetColors ||
		!slices.Equal(color1.Colors, []color.Color{color.Black}) {
		t.Fatalf("declarations[1] = %#v, want set-color black", declarations[1])
	}
}

func TestCompileSelfCantBeBlockedByMoreThanOneStaticRule(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource(
		"This creature can't be blocked by more than one creature.",
		pipelineContext{},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	declarations := compilation.Abilities[0].Static.Declarations
	if len(declarations) != 1 {
		t.Fatalf("declarations = %#v, want one", declarations)
	}
	declaration := declarations[0]
	if declaration.Kind != StaticDeclarationRule ||
		declaration.Group.Domain != StaticGroupSource ||
		declaration.Rule.Domain != StaticRuleDomainBlock ||
		declaration.Rule.Kind != StaticRuleCantBeBlockedByMoreThanOne ||
		declaration.Condition != nil {
		t.Fatalf("declaration = %#v, want source can't-be-blocked-by-more-than-one rule", declaration)
	}
}

func TestCompileSelfCharacteristicSetAllColors(t *testing.T) {
	t.Parallel()
	allColors := []color.Color{color.White, color.Blue, color.Black, color.Red, color.Green}
	for name, tc := range map[string]struct {
		source  string
		context pipelineContext
	}{
		"this creature": {
			source:  "This creature is all colors.",
			context: pipelineContext{},
		},
		"named source": {
			source:  "Transguild Courier is all colors.",
			context: pipelineContext{CardName: "Transguild Courier"},
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := compileSource(tc.source, tc.context)
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			declarations := compilation.Abilities[0].Static.Declarations
			if len(declarations) != 1 {
				t.Fatalf("declarations = %#v, want one", declarations)
			}
			declaration := declarations[0]
			if declaration.Group.Domain != StaticGroupSource {
				t.Fatalf("declaration group = %#v, want source", declaration.Group)
			}
			colorDecl := declaration.Continuous
			if colorDecl == nil ||
				colorDecl.Layer != StaticLayerColor ||
				colorDecl.Operation != StaticContinuousSetColors ||
				!slices.Equal(colorDecl.Colors, allColors) {
				t.Fatalf("declaration = %#v, want set all colors on source", declaration)
			}
		})
	}
}

func TestCompileStaticCharacteristicInAdditionComposition(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource(
		"Enchanted creature gets -1/-1 and is a black Zombie in addition to its other colors and types.",
		pipelineContext{},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	declarations := compilation.Abilities[0].Static.Declarations
	if len(declarations) != 3 {
		t.Fatalf("declarations = %#v, want three", declarations)
	}
	if declarations[0].Continuous.Layer != StaticLayerPowerToughnessModify {
		t.Fatalf("declarations[0] = %#v, want PT modify", declarations[0])
	}
	colorDecl := declarations[1].Continuous
	if colorDecl.Layer != StaticLayerColor ||
		colorDecl.Operation != StaticContinuousAddColors ||
		!slices.Equal(colorDecl.Colors, []color.Color{color.Black}) {
		t.Fatalf("declarations[1] = %#v, want add-color black", declarations[1])
	}
	typeDecl := declarations[2].Continuous
	if typeDecl.Layer != StaticLayerType ||
		typeDecl.Operation != StaticContinuousAddTypes ||
		!slices.Equal(typeDecl.AddSubtypes, []types.Sub{types.Zombie}) {
		t.Fatalf("declarations[2] = %#v, want add-subtype Zombie", declarations[2])
	}
}

func TestCompileStaticChosenCreatureTypeAddition(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource(
		"This creature is the chosen type in addition to its other types.",
		pipelineContext{},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	declarations := compilation.Abilities[0].Static.Declarations
	if len(declarations) != 1 {
		t.Fatalf("declarations = %#v, want one", declarations)
	}
	declaration := declarations[0]
	if declaration.Group.Domain != StaticGroupSource ||
		declaration.Continuous == nil ||
		declaration.Continuous.Layer != StaticLayerType ||
		declaration.Continuous.Operation != StaticContinuousAddSubtypeFromEntryChoice {
		t.Fatalf("declaration = %#v, want source entry-choice subtype addition", declaration)
	}
}

func TestCompileStaticChosenCreatureTypeTriggerMultiplier(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource(
		"If a triggered ability of another creature you control of the chosen type triggers, it triggers an additional time.",
		pipelineContext{},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	declarations := compilation.Abilities[0].Static.Declarations
	if len(declarations) != 1 {
		t.Fatalf("declarations = %#v, want one", declarations)
	}
	declaration := declarations[0]
	if declaration.Rule == nil ||
		declaration.Rule.Domain != StaticRuleDomainTrigger ||
		declaration.Rule.Kind != StaticRuleAdditionalTriggerForChosenCreatureType {
		t.Fatalf("declaration = %#v, want chosen-type trigger multiplier", declaration)
	}
}

func TestCompileStaticEnteringTriggerMultiplier(t *testing.T) {
	t.Parallel()
	for name, tc := range map[string]struct {
		source string
		types  []types.Card
	}{
		"artifact or creature": {
			source: "If an artifact or creature entering causes a triggered ability of a permanent you control to trigger, that ability triggers an additional time.",
			types:  []types.Card{types.Artifact, types.Creature},
		},
		"any permanent": {
			source: "If a permanent entering causes a triggered ability of a permanent you control to trigger, that ability triggers an additional time.",
			types:  nil,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := compileSource(tc.source, pipelineContext{})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			declarations := compilation.Abilities[0].Static.Declarations
			if len(declarations) != 1 {
				t.Fatalf("declarations = %#v, want one", declarations)
			}
			declaration := declarations[0]
			if declaration.Kind != StaticDeclarationEnteringTriggerMultiplier ||
				declaration.EnteringMultiplier == nil {
				t.Fatalf("declaration = %#v, want entering-trigger multiplier", declaration)
			}
			if !slices.Equal(declaration.EnteringMultiplier.EnteringTypes, tc.types) {
				t.Fatalf("types = %#v, want %#v", declaration.EnteringMultiplier.EnteringTypes, tc.types)
			}
		})
	}
}

func TestCompileStaticComposedContinuousFailClosed(t *testing.T) {
	t.Parallel()
	for name, source := range map[string]string{
		"base pt duration":     "Enchanted creature has base power and toughness 1/1 until end of turn.",
		"unrepresentable type": "Enchanted creature is a planeswalker in addition to its other types.",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			compilation, _ := compileSource(source, pipelineContext{})
			ability := compilation.Abilities[0]
			if ability.Static != nil {
				for _, declaration := range ability.Static.Declarations {
					if declaration.Continuous != nil &&
						(declaration.Continuous.Layer == StaticLayerPowerToughnessSet ||
							declaration.Continuous.Layer == StaticLayerColor ||
							declaration.Continuous.Layer == StaticLayerType) {
						t.Fatalf("declarations = %#v, want no new characteristic declaration (fail closed)",
							ability.Static.Declarations)
					}
				}
			}
		})
	}
}

func TestCompileStaticEnchantedTypeChangeColorlessLandWithMana(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource(
		"Enchanted permanent is a colorless land with \"{T}: Add {C}\" and loses all other card types and abilities.",
		pipelineContext{},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	declarations := compilation.Abilities[0].Static.Declarations
	var sawRemoveAbilities, sawColorless, sawLandType, sawGrantedMana bool
	for _, declaration := range declarations {
		if declaration.Group.Domain != StaticGroupAttachedObject {
			t.Fatalf("declaration domain = %v, want attached object: %#v", declaration.Group.Domain, declaration)
		}
		continuous := declaration.Continuous
		if continuous == nil {
			continue
		}
		if continuous.Layer == StaticLayerAbility && continuous.Operation == StaticContinuousRemoveAllAbilities {
			sawRemoveAbilities = true
		}
		if continuous.Layer == StaticLayerColor && continuous.SetColorless {
			sawColorless = true
		}
		if continuous.Layer == StaticLayerType {
			for _, cardType := range continuous.SetTypes {
				if cardType == StaticCardTypeLand {
					sawLandType = true
				}
			}
		}
		if continuous.GrantedMana != nil && continuous.GrantedMana.Colorless {
			sawGrantedMana = true
		}
	}
	if !sawRemoveAbilities || !sawColorless || !sawLandType || !sawGrantedMana {
		t.Fatalf("declarations = %#v, want remove-abilities + colorless + land type + colorless mana", declarations)
	}
}

func TestCompileStaticEnchantedTypeChangeBareSubtype(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource("Enchanted creature is a Bird.", pipelineContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	declarations := compilation.Abilities[0].Static.Declarations
	if len(declarations) != 1 {
		t.Fatalf("declarations = %#v, want one", declarations)
	}
	declaration := declarations[0]
	if declaration.Group.Domain != StaticGroupAttachedObject ||
		declaration.Continuous == nil ||
		declaration.Continuous.Layer != StaticLayerType ||
		len(declaration.Continuous.SetSubtypes) != 1 ||
		declaration.Continuous.SetSubtypes[0] != types.Bird {
		t.Fatalf("declaration = %#v, want attached-object Bird subtype set", declaration)
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

// TestCompileComposedPowerToughnessRuleDeclarations verifies that a compound
// "<subject> gets +N/+N and <rule>" wording recognizes both the continuous
// power/toughness declaration and the single-subject rule declaration on the
// source or its attached object.
func TestCompileComposedPowerToughnessRuleDeclarations(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		source string
		rule   StaticRuleKind
		group  StaticGroupDomain
	}{
		"enchanted gets and can't block": {
			source: "Enchanted creature gets +2/+2 and can't block.",
			rule:   StaticRuleCantBlock,
			group:  StaticGroupAttachedObject,
		},
		"enchanted gets and can't be blocked": {
			source: "Enchanted creature gets +1/+0 and can't be blocked.",
			rule:   StaticRuleCantBeBlocked,
			group:  StaticGroupAttachedObject,
		},
		"enchanted gets and attacks each combat": {
			source: "Enchanted creature gets +2/+2 and attacks each combat if able.",
			rule:   StaticRuleMustAttack,
			group:  StaticGroupAttachedObject,
		},
		"equipped gets and can't block": {
			source: "Equipped creature gets +2/+2 and can't block.",
			rule:   StaticRuleCantBlock,
			group:  StaticGroupAttachedObject,
		},
		"source gets and can't block": {
			source: "This creature gets +2/+2 and can't block.",
			rule:   StaticRuleCantBlock,
			group:  StaticGroupSource,
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
				t.Fatalf("ability has no static semantics: %#v", ability)
			}
			var continuous, rule *StaticDeclaration
			for i := range ability.Static.Declarations {
				declaration := &ability.Static.Declarations[i]
				switch declaration.Kind {
				case StaticDeclarationContinuous:
					continuous = declaration
				case StaticDeclarationRule:
					rule = declaration
				default:
				}
			}
			if continuous == nil {
				t.Fatalf("missing continuous power/toughness declaration: %#v", ability.Static.Declarations)
			}
			if rule == nil || rule.Rule == nil {
				t.Fatalf("missing rule declaration: %#v", ability.Static.Declarations)
			}
			if rule.Rule.Kind != test.rule {
				t.Fatalf("rule kind = %v, want %v", rule.Rule.Kind, test.rule)
			}
			if rule.Group.Domain != test.group {
				t.Fatalf("rule group = %v, want %v", rule.Group.Domain, test.group)
			}
			if continuous.Group.Domain != test.group {
				t.Fatalf("continuous group = %v, want %v", continuous.Group.Domain, test.group)
			}
		})
	}
}

// TestCompileComposedCharacteristicRuleDeclarations verifies that a compound
// "<subject> ... is a[n] <type> in addition to its other types ... and <rule>"
// wording recognizes both the type-addition characteristic declaration and the
// single-subject rule declaration on the attached object, including alongside a
// keyword grant (Brotherhood Regalia's "has ward {2}, is an Assassin in addition
// to its other types, and can't be blocked").
func TestCompileComposedCharacteristicRuleDeclarations(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		source       string
		rule         StaticRuleKind
		wantSubtype  types.Sub
		wantKeywords bool
	}{
		"type addition and can't be blocked": {
			source:      "Equipped creature is an Assassin in addition to its other types and can't be blocked.",
			rule:        StaticRuleCantBeBlocked,
			wantSubtype: types.Sub("Assassin"),
		},
		"keyword, type addition, and rule": {
			source:       "Equipped creature has ward {2}, is an Assassin in addition to its other types, and can't be blocked.",
			rule:         StaticRuleCantBeBlocked,
			wantSubtype:  types.Sub("Assassin"),
			wantKeywords: true,
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
			if ability.Static == nil || ability.Static.Blocker != StaticDeclarationBlockerNone {
				t.Fatalf("ability static = %#v, want recognized declarations", ability.Static)
			}
			var characteristic, rule, keyword *StaticDeclaration
			for i := range ability.Static.Declarations {
				declaration := &ability.Static.Declarations[i]
				switch {
				case declaration.Kind == StaticDeclarationRule:
					rule = declaration
				case declaration.Continuous != nil && declaration.Continuous.Layer == StaticLayerType:
					characteristic = declaration
				case declaration.Continuous != nil && declaration.Continuous.Operation == StaticContinuousGrantKeywords:
					keyword = declaration
				default:
				}
			}
			if characteristic == nil || len(characteristic.Continuous.AddSubtypes) != 1 ||
				characteristic.Continuous.AddSubtypes[0] != test.wantSubtype {
				t.Fatalf("missing type-addition declaration: %#v", ability.Static.Declarations)
			}
			if characteristic.Group.Domain != StaticGroupAttachedObject {
				t.Fatalf("characteristic group = %v, want attached object", characteristic.Group.Domain)
			}
			if rule == nil || rule.Rule == nil || rule.Rule.Kind != test.rule {
				t.Fatalf("missing rule declaration %v: %#v", test.rule, ability.Static.Declarations)
			}
			if rule.Group.Domain != StaticGroupAttachedObject {
				t.Fatalf("rule group = %v, want attached object", rule.Group.Domain)
			}
			if (keyword != nil) != test.wantKeywords {
				t.Fatalf("keyword grant present = %v, want %v: %#v", keyword != nil, test.wantKeywords, ability.Static.Declarations)
			}
		})
	}
}

// TestCompileComposedPowerToughnessRuleNearMissesFailClosed confirms that
// compound power/toughness and rule wordings outside the single-subject runtime
// model (battlefield groups, conditional rules) produce no rule declaration.
func TestCompileComposedPowerToughnessRuleNearMissesFailClosed(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"Other creatures you control get +1/+1 and can't block.",
		"Enchanted creature gets +2/+2 and can't block as long as you control an artifact.",
	} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			compilation, _ := compileSource(source, pipelineContext{})
			static := compilation.Abilities[0].Static
			if static != nil {
				for _, declaration := range static.Declarations {
					if declaration.Rule != nil {
						t.Fatalf("declaration = %#v, want no rule declaration (fail closed)", declaration)
					}
				}
			}
		})
	}
}

// TestCompileConditionalSelfStaticDeclarations covers self characteristic and
// keyword statics whose conditions exercise the leading possession-clause fix
// and the richer "you control <object>" matchers (token, multicolored, typed
// subtype). Each must compile to a recognized static declaration carrying a
// fully typed condition rather than a generic blocker.
func TestCompileConditionalSelfStaticDeclarations(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		source       string
		predicate    ConditionPredicate
		threshold    int
		tokenOnly    bool
		multicolored bool
		subtypes     []string
	}{
		"leading life": {
			source:    "As long as you have 30 or more life, this creature gets +5/+5 and has flying.",
			predicate: ConditionPredicateControllerLifeAtLeast,
			threshold: 30,
		},
		"leading hand empty": {
			source:    "As long as you have no cards in hand, this creature has double strike.",
			predicate: ConditionPredicateControllerHandEmpty,
		},
		"control token": {
			source:    "This creature gets +2/+0 and has trample as long as you control a token.",
			predicate: ConditionPredicateControllerControls,
			tokenOnly: true,
		},
		"control multicolored": {
			source:       "This creature gets +1/+1 and has first strike as long as you control another multicolored permanent.",
			predicate:    ConditionPredicateControllerControls,
			multicolored: true,
		},
		"control typed subtype": {
			source:    "This creature gets +3/+3 and has flying as long as you control a Griffin creature.",
			predicate: ConditionPredicateControllerControls,
			subtypes:  []string{"Griffin"},
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
			if ability.Static == nil || ability.Static.Blocker != StaticDeclarationBlockerNone {
				t.Fatalf("static = %#v, want no blocker", ability.Static)
			}
			if len(ability.Static.Declarations) == 0 {
				t.Fatalf("declarations = %#v, want at least one", ability.Static.Declarations)
			}
			condition := ability.Static.Declarations[0].Condition
			if condition == nil || condition.Predicate != test.predicate {
				t.Fatalf("condition = %#v, want predicate %v", condition, test.predicate)
			}
			if condition.Threshold != test.threshold {
				t.Fatalf("threshold = %d, want %d", condition.Threshold, test.threshold)
			}
			if condition.Selection.TokenOnly != test.tokenOnly ||
				condition.Selection.Multicolored != test.multicolored ||
				!slices.Equal(condition.Selection.SubtypesAny, test.subtypes) {
				t.Fatalf("selection = %#v", condition.Selection)
			}
		})
	}
}

// TestCompileConditionalSelfStaticFailClosed covers self statics whose
// conditions remain outside the supported vocabulary; they must fail closed with
// a condition blocker rather than compiling to an unsupported runtime static.
func TestCompileConditionalSelfStaticFailClosed(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		// HandSize-at-most is parsed but not representable as a static condition.
		"This creature gets +2/+2 as long as you have one or fewer cards in hand.",
		// "more cards in hand than each opponent" is not a typed predicate.
		"This creature gets +2/+2 as long as you have more cards in hand than each opponent.",
	} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			compilation, _ := compileSource(source, pipelineContext{})
			static := compilation.Abilities[0].Static
			if static != nil && static.Blocker == StaticDeclarationBlockerNone && len(static.Declarations) != 0 {
				if static.Declarations[0].Condition != nil &&
					static.Declarations[0].Condition.Predicate != ConditionPredicateUnsupported {
					t.Fatalf("source %q compiled to a supported condition %#v, want fail closed", source, static.Declarations[0].Condition)
				}
			}
		})
	}
}

func TestCompileStaticLoseAbilitiesBecomeDeclaration(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource(
		"Enchanted creature loses all abilities and is a blue Frog creature with base power and toughness 1/1. (It loses all other card types and creature types.)",
		pipelineContext{},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := compilation.Abilities[0]
	if ability.Static == nil || len(ability.Static.Declarations) != 4 {
		t.Fatalf("static semantics = %#v, want four declarations", ability.Static)
	}
	declarations := ability.Static.Declarations
	for _, declaration := range declarations {
		if declaration.Kind != StaticDeclarationContinuous || declaration.Continuous == nil {
			t.Fatalf("declaration = %#v, want continuous", declaration)
		}
		if declaration.Group.Domain != StaticGroupAttachedObject {
			t.Fatalf("group = %#v, want attached-object group", declaration.Group)
		}
	}
	if declarations[0].Continuous.Layer != StaticLayerAbility ||
		declarations[0].Continuous.Operation != StaticContinuousRemoveAllAbilities {
		t.Fatalf("declarations[0] = %#v, want remove-all-abilities", declarations[0])
	}
	colorDecl := declarations[1].Continuous
	if colorDecl.Layer != StaticLayerColor ||
		colorDecl.Operation != StaticContinuousSetColors ||
		!slices.Equal(colorDecl.Colors, []color.Color{color.Blue}) {
		t.Fatalf("declarations[1] = %#v, want set-color blue", declarations[1])
	}
	typeDecl := declarations[2].Continuous
	if typeDecl.Layer != StaticLayerType ||
		typeDecl.Operation != StaticContinuousSetTypes ||
		!slices.Equal(typeDecl.SetTypes, []StaticCardType{StaticCardTypeCreature}) ||
		!slices.Equal(typeDecl.SetSubtypes, []types.Sub{types.Frog}) {
		t.Fatalf("declarations[2] = %#v, want set creature Frog", declarations[2])
	}
	ptDecl := declarations[3].Continuous
	if ptDecl.Layer != StaticLayerPowerToughnessSet ||
		ptDecl.Operation != StaticContinuousSetBasePowerToughness ||
		ptDecl.SetPower != 1 || ptDecl.SetToughness != 1 {
		t.Fatalf("declarations[3] = %#v, want base 1/1 set", declarations[3])
	}
}

func TestCompileStaticLoseAbilitiesBecomeNameFailsClosed(t *testing.T) {
	t.Parallel()
	compilation, _ := compileSource(
		"Enchanted creature loses all abilities and is a green and white Citizen creature with base power and toughness 1/1 named Legitimate Businessperson.",
		pipelineContext{},
	)
	for _, ability := range compilation.Abilities {
		if ability.Static == nil {
			continue
		}
		for _, declaration := range ability.Static.Declarations {
			if declaration.Continuous != nil && declaration.Continuous.Operation == StaticContinuousRemoveAllAbilities {
				t.Fatal("name-setting polymorph unexpectedly produced a remove-all-abilities declaration")
			}
		}
	}
}
