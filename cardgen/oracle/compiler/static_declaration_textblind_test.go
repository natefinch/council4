package compiler

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
)

// These tests drive the static-declaration recognizers with constructed typed
// parser nodes and compiled semantic content only. The CompiledAbility carries
// no Oracle wording, proving the compiler derives every static-declaration
// meaning from typed syntax rather than from source text or tokens.

func staticTextBlindPTEffect() CompiledEffect {
	return CompiledEffect{
		Kind:           EffectModifyPT,
		PowerDelta:     CompiledSignedAmount{Value: 1, Known: true},
		ToughnessDelta: CompiledSignedAmount{Value: 1, Known: true},
	}
}

func sourceReference() CompiledReference {
	return CompiledReference{Binding: ReferenceBindingSource}
}

func TestRecognizeStaticPowerToughnessFromTypedNodes(t *testing.T) {
	t.Parallel()
	ability := CompiledAbility{
		Kind: AbilityStatic,
		Content: AbilityContent{
			Effects:    []CompiledEffect{staticTextBlindPTEffect()},
			References: []CompiledReference{sourceReference()},
		},
	}
	statics := []parser.StaticDeclarationSyntax{{Kind: parser.StaticDeclarationContinuousPowerToughness}}
	declarations, ok := recognizeStaticPowerToughnessDeclarations(ability, statics)
	if !ok || len(declarations) != 1 {
		t.Fatalf("declarations = %#v ok = %v, want one", declarations, ok)
	}
	if declarations[0].Continuous == nil ||
		declarations[0].Continuous.Layer != StaticLayerPowerToughnessModify ||
		declarations[0].Group.Domain != StaticGroupSource {
		t.Fatalf("declaration = %#v", declarations[0])
	}
}

func TestRecognizeStaticPowerToughnessWithKeywordFromTypedNodes(t *testing.T) {
	t.Parallel()
	ability := CompiledAbility{
		Kind: AbilityStatic,
		Content: AbilityContent{
			Effects:    []CompiledEffect{staticTextBlindPTEffect()},
			Keywords:   []CompiledKeyword{{Kind: parser.KeywordFlying}},
			References: []CompiledReference{sourceReference()},
		},
	}
	statics := []parser.StaticDeclarationSyntax{
		{Kind: parser.StaticDeclarationContinuousPowerToughness},
		{Kind: parser.StaticDeclarationKeywordGrant},
	}
	declarations, ok := recognizeStaticPowerToughnessDeclarations(ability, statics)
	if !ok || len(declarations) != 2 {
		t.Fatalf("declarations = %#v ok = %v, want two", declarations, ok)
	}
	if declarations[1].Continuous == nil ||
		declarations[1].Continuous.Operation != StaticContinuousGrantKeywords {
		t.Fatalf("declaration = %#v, want keyword grant", declarations[1])
	}
}

func TestRecognizeStaticPowerToughnessDynamicMismatchFailsClosed(t *testing.T) {
	t.Parallel()
	ability := CompiledAbility{
		Kind: AbilityStatic,
		Content: AbilityContent{
			Effects:    []CompiledEffect{staticTextBlindPTEffect()},
			References: []CompiledReference{sourceReference()},
		},
	}
	statics := []parser.StaticDeclarationSyntax{
		{Kind: parser.StaticDeclarationContinuousPowerToughness, Dynamic: true},
	}
	if _, ok := recognizeStaticPowerToughnessDeclarations(ability, statics); ok {
		t.Fatal("recognized dynamic PT against a static effect, want fail closed")
	}
}

func TestRecognizeStaticKeywordGrantGroupFromTypedNodes(t *testing.T) {
	t.Parallel()
	ability := CompiledAbility{
		Kind: AbilityStatic,
		Content: AbilityContent{
			Effects: []CompiledEffect{{
				Kind:          EffectGrantKeyword,
				StaticSubject: StaticSubjectControlledCreatures,
			}},
			Keywords: []CompiledKeyword{{Kind: parser.KeywordTrample}},
		},
	}
	statics := []parser.StaticDeclarationSyntax{{Kind: parser.StaticDeclarationKeywordGrant}}
	declarations, ok := recognizeStaticKeywordGrantDeclarations(ability, statics)
	if !ok || len(declarations) != 1 {
		t.Fatalf("declarations = %#v ok = %v, want one", declarations, ok)
	}
	if declarations[0].Group.Domain != StaticGroupSourceControllerPermanents {
		t.Fatalf("declaration = %#v, want controlled-creature group", declarations[0])
	}
}

func TestRecognizeStaticKeywordGrantCounterFilterFromTypedNodes(t *testing.T) {
	t.Parallel()
	ability := CompiledAbility{
		Kind: AbilityStatic,
		Content: AbilityContent{
			Effects: []CompiledEffect{{
				Kind:          EffectGrantKeyword,
				StaticSubject: StaticSubjectControlledCreatures,
				Details: &CompiledEffectDetails{
					StaticSubjectCounter: &CompiledStaticSubjectCounter{Kind: counter.PlusOnePlusOne},
				},
			}},
			Keywords: []CompiledKeyword{{Kind: parser.KeywordFlying}},
			// The "with a +1/+1 counter on it" qualifier names the affected
			// creature with the pronoun "it"; the recognizer tolerates that
			// self-reference rather than treating it as a separate antecedent.
			References: []CompiledReference{{Pronoun: ReferencePronounIt}},
		},
	}
	statics := []parser.StaticDeclarationSyntax{{Kind: parser.StaticDeclarationKeywordGrant}}
	declarations, ok := recognizeStaticKeywordGrantDeclarations(ability, statics)
	if !ok || len(declarations) != 1 {
		t.Fatalf("declarations = %#v ok = %v, want one", declarations, ok)
	}
	if declarations[0].Group.Domain != StaticGroupSourceControllerPermanents ||
		!declarations[0].Group.Selection.MatchCounter ||
		declarations[0].Group.Selection.RequiredCounter != counter.PlusOnePlusOne {
		t.Fatalf("declaration = %#v, want controlled-creature group requiring a +1/+1 counter", declarations[0])
	}
}

func TestRecognizeStaticKeywordGrantAnyCounterFilterFromTypedNodes(t *testing.T) {
	t.Parallel()
	ability := CompiledAbility{
		Kind: AbilityStatic,
		Content: AbilityContent{
			Effects: []CompiledEffect{{
				Kind:          EffectGrantKeyword,
				StaticSubject: StaticSubjectControlledCreatures,
				Details: &CompiledEffectDetails{
					StaticSubjectCounter: &CompiledStaticSubjectCounter{Any: true},
				},
			}},
			Keywords: []CompiledKeyword{{Kind: parser.KeywordFlying}},
			// The kind-agnostic "with a counter on it" qualifier (Rishkar)
			// names the affected creature with the pronoun "it"; the
			// recognizer tolerates that self-reference.
			References: []CompiledReference{{Pronoun: ReferencePronounIt}},
		},
	}
	statics := []parser.StaticDeclarationSyntax{{Kind: parser.StaticDeclarationKeywordGrant}}
	declarations, ok := recognizeStaticKeywordGrantDeclarations(ability, statics)
	if !ok || len(declarations) != 1 {
		t.Fatalf("declarations = %#v ok = %v, want one", declarations, ok)
	}
	if declarations[0].Group.Domain != StaticGroupSourceControllerPermanents ||
		!declarations[0].Group.Selection.MatchAnyCounter ||
		declarations[0].Group.Selection.MatchCounter {
		t.Fatalf("declaration = %#v, want controlled-creature group requiring any counter", declarations[0])
	}
}

func TestRecognizeStaticChosenTypePowerToughnessGroupFromTypedNodes(t *testing.T) {
	t.Parallel()
	for name, subject := range map[string]StaticSubjectKind{
		"controlled":       StaticSubjectControlledCreaturesChosenType,
		"other controlled": StaticSubjectOtherControlledCreaturesChosenType,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ability := CompiledAbility{
				Kind: AbilityStatic,
				Content: AbilityContent{
					Effects: []CompiledEffect{{
						Kind:           EffectModifyPT,
						PowerDelta:     CompiledSignedAmount{Value: 1, Known: true},
						ToughnessDelta: CompiledSignedAmount{Value: 1, Known: true},
						StaticSubject:  subject,
					}},
				},
			}
			statics := []parser.StaticDeclarationSyntax{{Kind: parser.StaticDeclarationContinuousPowerToughness}}
			declarations, ok := recognizeStaticPowerToughnessDeclarations(ability, statics)
			if !ok || len(declarations) != 1 {
				t.Fatalf("declarations = %#v ok = %v, want one", declarations, ok)
			}
			if declarations[0].Group.Domain != StaticGroupSourceControllerPermanents ||
				!declarations[0].Group.Selection.SubtypeFromEntryChoice {
				t.Fatalf("declaration = %#v, want chosen-type controlled group", declarations[0])
			}
			wantExclude := subject == StaticSubjectOtherControlledCreaturesChosenType
			if declarations[0].Group.ExcludeSource != wantExclude {
				t.Fatalf("ExcludeSource = %v, want %v", declarations[0].Group.ExcludeSource, wantExclude)
			}
		})
	}
}

func TestRecognizeStaticKeywordGrantSourceRequiresConditionFailsClosed(t *testing.T) {
	t.Parallel()
	ability := CompiledAbility{
		Kind: AbilityStatic,
		Content: AbilityContent{
			Effects:    []CompiledEffect{{Kind: EffectGrantKeyword}},
			Keywords:   []CompiledKeyword{{Kind: parser.KeywordFlying}},
			References: []CompiledReference{sourceReference()},
		},
	}
	statics := []parser.StaticDeclarationSyntax{{Kind: parser.StaticDeclarationKeywordGrant}}
	if _, ok := recognizeStaticKeywordGrantDeclarations(ability, statics); ok {
		t.Fatal("recognized unconditional source keyword grant, want fail closed")
	}
}

func TestRecognizeStaticPermanentManaAbilityGrantFromTypedNode(t *testing.T) {
	t.Parallel()
	ability := CompiledAbility{Kind: AbilityStatic}
	statics := []parser.StaticDeclarationSyntax{{
		Kind: parser.StaticDeclarationPermanentAbilityGrant,
		Subject: parser.StaticDeclarationSubject{
			Kind: parser.StaticDeclarationSubjectGroup,
			Group: parser.EffectStaticSubjectSyntax{
				Kind: parser.EffectStaticSubjectControlledLands,
			},
		},
		GrantedManaAbility: &parser.StaticGrantedManaAbilitySyntax{
			TapCost:  true,
			Amount:   1,
			AnyColor: true,
		},
	}}
	declaration, ok := recognizeStaticPermanentAbilityGrantDeclaration(ability, statics)
	if !ok {
		t.Fatal("did not recognize typed permanent mana-ability grant")
	}
	if declaration.Continuous == nil ||
		declaration.Continuous.Operation != StaticContinuousGrantManaAbility ||
		declaration.Group.Domain != StaticGroupSourceControllerPermanents ||
		!slices.Equal(declaration.Group.Selection.RequiredTypes, []types.Card{types.Land}) {
		t.Fatalf("declaration = %#v, want controlled-land mana-ability grant", declaration)
	}
}

func TestRecognizeStaticPermanentManaAbilityGrantTreasureSacrifice(t *testing.T) {
	t.Parallel()
	ability := CompiledAbility{Kind: AbilityStatic}
	statics := []parser.StaticDeclarationSyntax{{
		Kind: parser.StaticDeclarationPermanentAbilityGrant,
		Subject: parser.StaticDeclarationSubject{
			Kind: parser.StaticDeclarationSubjectGroup,
			Group: parser.EffectStaticSubjectSyntax{
				Kind:         parser.EffectStaticSubjectControlledArtifacts,
				Subtype:      types.Treasure,
				SubtypeKnown: true,
			},
		},
		GrantedManaAbility: &parser.StaticGrantedManaAbilitySyntax{
			TapCost:     true,
			Amount:      3,
			Sacrifice:   true,
			AnyOneColor: true,
			Text:        "{T}, Sacrifice this artifact: Add three mana of any one color.",
		},
	}}
	declaration, ok := recognizeStaticPermanentAbilityGrantDeclaration(ability, statics)
	if !ok {
		t.Fatal("did not recognize typed Treasure sacrifice mana-ability grant")
	}
	if declaration.Continuous == nil ||
		declaration.Continuous.GrantedMana == nil ||
		!declaration.Continuous.GrantedMana.Sacrifice ||
		!declaration.Continuous.GrantedMana.AnyOneColor ||
		declaration.Continuous.GrantedMana.Amount != 3 ||
		!slices.Equal(declaration.Group.Selection.RequiredTypes, []types.Card{types.Artifact}) ||
		!slices.Equal(declaration.Group.Selection.SubtypesAny, []types.Sub{types.Treasure}) {
		t.Fatalf("declaration = %#v, want controlled-Treasure sacrifice mana-ability grant", declaration)
	}
}

func TestRecognizeStaticPermanentManaAbilityGrantSacrificeAnyColor(t *testing.T) {
	t.Parallel()
	ability := CompiledAbility{Kind: AbilityStatic}
	statics := []parser.StaticDeclarationSyntax{{
		Kind: parser.StaticDeclarationPermanentAbilityGrant,
		Subject: parser.StaticDeclarationSubject{
			Kind: parser.StaticDeclarationSubjectGroup,
			Group: parser.EffectStaticSubjectSyntax{
				Kind: parser.EffectStaticSubjectControlledArtifacts,
			},
		},
		GrantedManaAbility: &parser.StaticGrantedManaAbilitySyntax{
			TapCost:   true,
			Amount:    1,
			AnyColor:  true,
			Sacrifice: true,
			Text:      "{T}, Sacrifice this artifact: Add one mana of any color.",
		},
	}}
	declaration, ok := recognizeStaticPermanentAbilityGrantDeclaration(ability, statics)
	if !ok {
		t.Fatal("did not recognize typed sacrifice any-color mana-ability grant")
	}
	if declaration.Continuous == nil ||
		declaration.Continuous.GrantedMana == nil ||
		!declaration.Continuous.GrantedMana.Sacrifice ||
		!declaration.Continuous.GrantedMana.AnyColor ||
		declaration.Continuous.GrantedMana.AnyOneColor ||
		declaration.Continuous.GrantedMana.Amount != 1 ||
		!slices.Equal(declaration.Group.Selection.RequiredTypes, []types.Card{types.Artifact}) {
		t.Fatalf("declaration = %#v, want controlled-artifact sacrifice any-color mana-ability grant", declaration)
	}
}

func TestRecognizeStaticPermanentManaAbilityGrantTypedNearMissesFailClosed(t *testing.T) {
	t.Parallel()
	base := parser.StaticDeclarationSyntax{
		Kind: parser.StaticDeclarationPermanentAbilityGrant,
		Subject: parser.StaticDeclarationSubject{
			Kind: parser.StaticDeclarationSubjectGroup,
			Group: parser.EffectStaticSubjectSyntax{
				Kind: parser.EffectStaticSubjectControlledLands,
			},
		},
		GrantedManaAbility: &parser.StaticGrantedManaAbilitySyntax{
			TapCost:  true,
			Amount:   1,
			AnyColor: true,
		},
	}
	tests := map[string]parser.StaticDeclarationSyntax{
		"unsupported group": func() parser.StaticDeclarationSyntax {
			node := base
			node.Subject.Group.Kind = parser.EffectStaticSubjectAllCreatures
			return node
		}(),
		"no tap cost": func() parser.StaticDeclarationSyntax {
			node := base
			granted := *base.GrantedManaAbility
			granted.TapCost = false
			node.GrantedManaAbility = &granted
			return node
		}(),
		"different amount": func() parser.StaticDeclarationSyntax {
			node := base
			granted := *base.GrantedManaAbility
			granted.Amount = 2
			node.GrantedManaAbility = &granted
			return node
		}(),
		"sacrifice without any-one-color": func() parser.StaticDeclarationSyntax {
			node := base
			granted := *base.GrantedManaAbility
			granted.AnyColor = false
			granted.Sacrifice = true
			granted.Amount = 3
			node.GrantedManaAbility = &granted
			return node
		}(),
	}
	for name, node := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if _, ok := recognizeStaticPermanentAbilityGrantDeclaration(
				CompiledAbility{Kind: AbilityStatic},
				[]parser.StaticDeclarationSyntax{node},
			); ok {
				t.Fatal("recognized unsupported typed grant, want fail closed")
			}
		})
	}
}

func TestRecognizeMixedSourceStaticDeclarationsFromTypedNodes(t *testing.T) {
	t.Parallel()
	ability := CompiledAbility{
		Kind: AbilityStatic,
		Content: AbilityContent{
			Effects:    []CompiledEffect{staticTextBlindPTEffect()},
			Conditions: []CompiledCondition{{Predicate: ConditionPredicateControllerLifeAtLeast, Threshold: 7}},
			Keywords:   []CompiledKeyword{{Kind: parser.KeywordFlying}},
			References: []CompiledReference{sourceReference()},
		},
	}
	statics := []parser.StaticDeclarationSyntax{
		{Kind: parser.StaticDeclarationContinuousPowerToughness},
		{Kind: parser.StaticDeclarationKeywordGrant},
		{Kind: parser.StaticDeclarationRule, Rule: parser.StaticRuleSyntax{
			Subject:    parser.StaticRuleSubject{Kind: parser.StaticRuleSubjectSourceCreature},
			Constraint: parser.StaticRuleConstraint{Kind: parser.StaticRuleConstraintRequirement},
			Operation:  parser.StaticRuleOperation{Kind: parser.StaticRuleOperationAttack, Voice: parser.StaticRuleVoiceActive},
			Qualifiers: []parser.StaticRuleQualifier{
				{Kind: parser.StaticRuleQualifierEachCombat},
				{Kind: parser.StaticRuleQualifierIfAble},
			},
		}},
	}
	declarations, ok := recognizeMixedSourceStaticDeclarations(ability, statics)
	if !ok || len(declarations) != 3 {
		t.Fatalf("declarations = %#v ok = %v, want three", declarations, ok)
	}
	if declarations[2].Rule == nil || declarations[2].Rule.Kind != StaticRuleMustAttack {
		t.Fatalf("declaration = %#v, want must-attack rule", declarations[2])
	}
	for i, declaration := range declarations {
		if declaration.Condition == nil || declaration.Group.Domain != StaticGroupSource {
			t.Fatalf("declaration %d = %#v, want conditional source declaration", i, declaration)
		}
	}
}

func TestRecognizeStaticCostModifierFromTypedNodes(t *testing.T) {
	t.Parallel()
	ability := CompiledAbility{
		Kind: AbilityStatic,
		Content: AbilityContent{
			Keywords: []CompiledKeyword{{Kind: parser.KeywordCycling, ParameterKind: parser.KeywordParameterNone}},
		},
	}
	statics := []parser.StaticDeclarationSyntax{{
		Kind:                parser.StaticDeclarationCostModifier,
		CostModifier:        parser.StaticDeclarationCostModifierAbilityReduction,
		CostReductionAmount: 2,
	}}
	declaration, ok := recognizeStaticCostModifierDeclaration(ability, statics)
	if !ok {
		t.Fatal("did not recognize typed cost modifier")
	}
	if declaration.Cost == nil ||
		declaration.Cost.GenericReduction != 2 ||
		declaration.Group.Domain != StaticGroupControllerHandCards {
		t.Fatalf("declaration = %#v", declaration)
	}
}

func TestRecognizeStaticAbilityCostSetFromTypedNodes(t *testing.T) {
	t.Parallel()
	ability := CompiledAbility{
		Kind: AbilityStatic,
		Content: AbilityContent{
			Keywords: []CompiledKeyword{{Kind: parser.KeywordEquip, ParameterKind: parser.KeywordParameterNone}},
		},
	}
	statics := []parser.StaticDeclarationSyntax{{
		Kind:               parser.StaticDeclarationAbilityCostSet,
		AbilityCostKeyword: parser.KeywordEquip,
		CostReplacement:    "",
	}}
	declaration, ok := recognizeStaticAbilityCostSetDeclaration(ability, statics)
	if !ok {
		t.Fatal("did not recognize typed ability-cost-set declaration")
	}
	if declaration.Kind != StaticDeclarationCostModifier ||
		declaration.Group.Domain != StaticGroupControllerEquipment ||
		declaration.Cost == nil ||
		declaration.Cost.Kind != StaticCostModifierAbility ||
		declaration.Cost.AbilityKeyword != parser.KeywordEquip ||
		!declaration.Cost.ReplaceManaCost ||
		declaration.Cost.SetManaCost != "" {
		t.Fatalf("declaration = %#v cost = %#v", declaration, declaration.Cost)
	}
}

func TestRecognizeStaticSpellCostModifierFromTypedNodes(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		node       parser.StaticDeclarationSyntax
		reduction  int
		increase   int
		types      []types.Card
		matchColor bool
		color      color.Color
	}{
		"all spells reduction": {
			node: parser.StaticDeclarationSyntax{
				Kind:                parser.StaticDeclarationCostModifier,
				CostModifier:        parser.StaticDeclarationCostModifierSpellReduction,
				CostReductionAmount: 1,
				SpellType:           parser.StaticDeclarationSpellTypeAll,
			},
			reduction: 1,
			types:     nil,
		},
		"creature spells reduction": {
			node: parser.StaticDeclarationSyntax{
				Kind:                parser.StaticDeclarationCostModifier,
				CostModifier:        parser.StaticDeclarationCostModifierSpellReduction,
				CostReductionAmount: 2,
				SpellType:           parser.StaticDeclarationSpellTypeCreature,
			},
			reduction: 2,
			types:     []types.Card{types.Creature},
		},
		"creature spells increase": {
			node: parser.StaticDeclarationSyntax{
				Kind:                parser.StaticDeclarationCostModifier,
				CostModifier:        parser.StaticDeclarationCostModifierSpellIncrease,
				CostReductionAmount: 1,
				SpellType:           parser.StaticDeclarationSpellTypeCreature,
			},
			increase: 1,
			types:    []types.Card{types.Creature},
		},
		"instant and sorcery reduction": {
			node: parser.StaticDeclarationSyntax{
				Kind:                parser.StaticDeclarationCostModifier,
				CostModifier:        parser.StaticDeclarationCostModifierSpellReduction,
				CostReductionAmount: 1,
				SpellType:           parser.StaticDeclarationSpellTypeInstantOrSorcery,
			},
			reduction: 1,
			types:     []types.Card{types.Instant, types.Sorcery},
		},
		"red spells reduction": {
			node: parser.StaticDeclarationSyntax{
				Kind:                parser.StaticDeclarationCostModifier,
				CostModifier:        parser.StaticDeclarationCostModifierSpellReduction,
				CostReductionAmount: 1,
				SpellColor:          parser.StaticDeclarationSpellColorRed,
			},
			reduction:  1,
			matchColor: true,
			color:      color.Red,
		},
		"colorless spells reduction": {
			node: parser.StaticDeclarationSyntax{
				Kind:                parser.StaticDeclarationCostModifier,
				CostModifier:        parser.StaticDeclarationCostModifierSpellReduction,
				CostReductionAmount: 1,
				SpellColor:          parser.StaticDeclarationSpellColorColorless,
			},
			reduction:  1,
			matchColor: true,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ability := CompiledAbility{Kind: AbilityStatic}
			declaration, ok := recognizeStaticSpellCostModifierDeclaration(ability, []parser.StaticDeclarationSyntax{test.node})
			if !ok {
				t.Fatal("did not recognize typed spell cost modifier")
			}
			if declaration.Cost == nil ||
				declaration.Cost.Kind != StaticCostModifierSpell ||
				declaration.Cost.GenericReduction != test.reduction ||
				declaration.Cost.GenericIncrease != test.increase ||
				declaration.Cost.MatchSpellColor != test.matchColor ||
				declaration.Cost.SpellColor != test.color ||
				declaration.Group.Domain != StaticGroupControllerSpells ||
				!slices.Equal(declaration.Cost.SpellTypes, test.types) {
				t.Fatalf("declaration = %#v", declaration)
			}
		})
	}
}

func TestRecognizeStaticChosenTypeSpellCostModifierFromTypedNode(t *testing.T) {
	node := parser.StaticDeclarationSyntax{
		Kind:                parser.StaticDeclarationCostModifier,
		CostModifier:        parser.StaticDeclarationCostModifierSpellReduction,
		CostReductionAmount: 1,
		SpellType:           parser.StaticDeclarationSpellTypeCreature,
		ChosenCreatureType:  true,
	}

	declaration, ok := recognizeStaticSpellCostModifierDeclaration(
		CompiledAbility{Kind: AbilityStatic},
		[]parser.StaticDeclarationSyntax{node},
	)

	if !ok || declaration.Cost == nil ||
		!declaration.Cost.ChosenSubtypeFromEntryChoice {
		t.Fatalf("declaration = %#v ok = %v, want chosen subtype entry-choice provenance", declaration, ok)
	}
}

func TestRecognizeStaticGraveyardZoneSpellCostModifierFromTypedNode(t *testing.T) {
	t.Parallel()
	node := parser.StaticDeclarationSyntax{
		Kind:                parser.StaticDeclarationCostModifier,
		CostModifier:        parser.StaticDeclarationCostModifierSpellReduction,
		CostReductionAmount: 1,
		SpellType:           parser.StaticDeclarationSpellTypeAll,
		SpellCastZone:       parser.StaticDeclarationCastZoneGraveyard,
	}

	declaration, ok := recognizeStaticSpellCostModifierDeclaration(
		CompiledAbility{Kind: AbilityStatic},
		[]parser.StaticDeclarationSyntax{node},
	)

	if !ok || declaration.Cost == nil ||
		declaration.Cost.SourceZone != parser.StaticDeclarationCastZoneGraveyard ||
		declaration.Cost.GenericReduction != 1 {
		t.Fatalf("declaration = %#v ok = %v, want graveyard-scoped reduction", declaration, ok)
	}
}

func TestRecognizeStaticPowerThresholdSpellCostModifierFromTypedNode(t *testing.T) {
	t.Parallel()
	node := parser.StaticDeclarationSyntax{
		Kind:                   parser.StaticDeclarationCostModifier,
		CostModifier:           parser.StaticDeclarationCostModifierSpellReduction,
		CostReductionAmount:    2,
		SpellType:              parser.StaticDeclarationSpellTypeCreature,
		SpellPowerAtLeast:      4,
		MatchSpellPowerAtLeast: true,
	}

	declaration, ok := recognizeStaticSpellCostModifierDeclaration(
		CompiledAbility{Kind: AbilityStatic},
		[]parser.StaticDeclarationSyntax{node},
	)

	if !ok || declaration.Cost == nil ||
		!declaration.Cost.MatchMinPower ||
		declaration.Cost.MinPower != 4 ||
		declaration.Cost.GenericReduction != 2 {
		t.Fatalf("declaration = %#v ok = %v, want power-4 creature reduction", declaration, ok)
	}
}

func TestRecognizeStaticZeroPowerThresholdSpellCostModifierFailsClosed(t *testing.T) {
	t.Parallel()
	node := parser.StaticDeclarationSyntax{
		Kind:                   parser.StaticDeclarationCostModifier,
		CostModifier:           parser.StaticDeclarationCostModifierSpellReduction,
		CostReductionAmount:    2,
		SpellType:              parser.StaticDeclarationSpellTypeCreature,
		MatchSpellPowerAtLeast: true,
	}

	if _, ok := recognizeStaticSpellCostModifierDeclaration(
		CompiledAbility{Kind: AbilityStatic},
		[]parser.StaticDeclarationSyntax{node},
	); ok {
		t.Fatal("recognized a zero power threshold cost modifier, want fail closed")
	}
}

func TestRecognizeStaticZoneScopedChosenTypeSpellCostModifierFailsClosed(t *testing.T) {
	t.Parallel()
	node := parser.StaticDeclarationSyntax{
		Kind:                parser.StaticDeclarationCostModifier,
		CostModifier:        parser.StaticDeclarationCostModifierSpellReduction,
		CostReductionAmount: 1,
		SpellType:           parser.StaticDeclarationSpellTypeCreature,
		ChosenCreatureType:  true,
		SpellCastZone:       parser.StaticDeclarationCastZoneGraveyard,
	}

	if _, ok := recognizeStaticSpellCostModifierDeclaration(
		CompiledAbility{Kind: AbilityStatic},
		[]parser.StaticDeclarationSyntax{node},
	); ok {
		t.Fatal("recognized a zone-scoped chosen-type cost modifier, want fail closed")
	}
}

func TestRecognizeStaticSpellColorDisjunctionCostModifierFromTypedNode(t *testing.T) {
	t.Parallel()
	node := parser.StaticDeclarationSyntax{
		Kind:                parser.StaticDeclarationCostModifier,
		CostModifier:        parser.StaticDeclarationCostModifierSpellReduction,
		CostReductionAmount: 1,
		SpellType:           parser.StaticDeclarationSpellTypeAll,
		SpellColors: []parser.StaticDeclarationSpellColorKind{
			parser.StaticDeclarationSpellColorRed,
			parser.StaticDeclarationSpellColorGreen,
		},
	}
	declaration, ok := recognizeStaticSpellCostModifierDeclaration(
		CompiledAbility{Kind: AbilityStatic},
		[]parser.StaticDeclarationSyntax{node},
	)
	if !ok || declaration.Cost == nil ||
		declaration.Cost.Kind != StaticCostModifierSpell ||
		declaration.Cost.GenericReduction != 1 ||
		declaration.Cost.MatchSpellColor ||
		len(declaration.Cost.SpellTypes) != 0 ||
		declaration.Group.Domain != StaticGroupControllerSpells ||
		!slices.Equal(declaration.Cost.SpellColors, []color.Color{color.Red, color.Green}) {
		t.Fatalf("declaration = %#v ok = %v, want red/green disjunction", declaration, ok)
	}
}

func TestRecognizeStaticSpellColorDisjunctionFailsClosed(t *testing.T) {
	t.Parallel()
	tests := map[string]parser.StaticDeclarationSyntax{
		"single color disjunction": {
			Kind:                parser.StaticDeclarationCostModifier,
			CostModifier:        parser.StaticDeclarationCostModifierSpellReduction,
			CostReductionAmount: 1,
			SpellColors:         []parser.StaticDeclarationSpellColorKind{parser.StaticDeclarationSpellColorRed},
		},
		"colorless in disjunction": {
			Kind:                parser.StaticDeclarationCostModifier,
			CostModifier:        parser.StaticDeclarationCostModifierSpellReduction,
			CostReductionAmount: 1,
			SpellColors: []parser.StaticDeclarationSpellColorKind{
				parser.StaticDeclarationSpellColorRed,
				parser.StaticDeclarationSpellColorColorless,
			},
		},
		"disjunction with type filter": {
			Kind:                parser.StaticDeclarationCostModifier,
			CostModifier:        parser.StaticDeclarationCostModifierSpellReduction,
			CostReductionAmount: 1,
			SpellType:           parser.StaticDeclarationSpellTypeCreature,
			SpellColors: []parser.StaticDeclarationSpellColorKind{
				parser.StaticDeclarationSpellColorRed,
				parser.StaticDeclarationSpellColorGreen,
			},
		},
	}
	for name, node := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if _, ok := recognizeStaticSpellCostModifierDeclaration(
				CompiledAbility{Kind: AbilityStatic},
				[]parser.StaticDeclarationSyntax{node},
			); ok {
				t.Fatalf("recognized malformed color disjunction %#v, want fail closed", node)
			}
		})
	}
}

func TestRecognizeStaticSpellCostModifierFailsClosedOnContent(t *testing.T) {
	t.Parallel()
	node := parser.StaticDeclarationSyntax{
		Kind:                parser.StaticDeclarationCostModifier,
		CostModifier:        parser.StaticDeclarationCostModifierSpellReduction,
		CostReductionAmount: 1,
		SpellType:           parser.StaticDeclarationSpellTypeAll,
	}
	ability := CompiledAbility{
		Kind:    AbilityStatic,
		Content: AbilityContent{Conditions: []CompiledCondition{{}}},
	}
	if _, ok := recognizeStaticSpellCostModifierDeclaration(ability, []parser.StaticDeclarationSyntax{node}); ok {
		t.Fatal("recognized spell cost modifier despite a condition in content")
	}
}

func TestRecognizeStaticCardAbilityGrantFromTypedNodes(t *testing.T) {
	t.Parallel()
	ability := CompiledAbility{
		Kind: AbilityStatic,
		Content: AbilityContent{
			Keywords: []CompiledKeyword{{
				Kind:          parser.KeywordCycling,
				Parameter:     "{2}",
				ParameterKind: parser.KeywordParameterManaCost,
			}},
		},
	}
	statics := []parser.StaticDeclarationSyntax{{
		Kind: parser.StaticDeclarationCardAbilityGrant,
		Subject: parser.StaticDeclarationSubject{
			Kind:       parser.StaticDeclarationSubjectControllerHand,
			CardFilter: parser.StaticDeclarationCardFilterLand,
		},
	}}
	declaration, ok := recognizeStaticCardAbilityGrantDeclaration(ability, statics)
	if !ok {
		t.Fatal("did not recognize typed card-ability grant")
	}
	if declaration.CardGrant == nil ||
		declaration.CardGrant.Text != "Each land card in your hand has cycling {2}." ||
		len(declaration.Group.Selection.RequiredTypes) != 1 ||
		declaration.Group.Selection.RequiredTypes[0] != types.Land {
		t.Fatalf("declaration = %#v", declaration)
	}
}

func TestRecognizeStaticDeclarationsFailClosedOnMismatchedKinds(t *testing.T) {
	t.Parallel()
	ability := CompiledAbility{
		Kind: AbilityStatic,
		Content: AbilityContent{
			Effects:    []CompiledEffect{staticTextBlindPTEffect()},
			References: []CompiledReference{sourceReference()},
		},
	}
	// The parser emitted a keyword-grant node, but the compiled effect is a
	// power/toughness change: the PT recognizer must decline rather than guess.
	statics := []parser.StaticDeclarationSyntax{{Kind: parser.StaticDeclarationKeywordGrant}}
	if _, ok := recognizeStaticPowerToughnessDeclarations(ability, statics); ok {
		t.Fatal("PT recognizer matched a keyword-grant node, want fail closed")
	}
	if _, ok := recognizeStaticPowerToughnessDeclarations(ability, nil); ok {
		t.Fatal("PT recognizer matched absent syntax, want fail closed")
	}
}

func TestRecognizeStaticAttackTaxFromTypedNodeWithoutInspectingText(t *testing.T) {
	t.Parallel()
	content := AbilityContent{
		Conditions: []CompiledCondition{{
			Kind:      ConditionUnless,
			Predicate: ConditionPredicateUnsupported,
			Negated:   true,
			Text:      "unrelated retained condition text",
		}},
		References: []CompiledReference{
			{Pronoun: ReferencePronounTheir, Binding: ReferenceBindingAmbiguous, Text: "unrelated"},
			{Pronoun: ReferencePronounThey, Binding: ReferenceBindingAmbiguous, Text: "unrelated"},
		},
	}
	node := parser.StaticDeclarationSyntax{
		Kind:             parser.StaticDeclarationPlayerRule,
		Subject:          parser.StaticDeclarationSubject{Kind: parser.StaticDeclarationSubjectController},
		PlayerRule:       parser.StaticDeclarationPlayerRuleAttackTax,
		AttackTaxGeneric: 2,
	}
	declaration, ok := recognizeStaticPlayerRuleDeclaration(CompiledAbility{Kind: AbilityStatic, Content: content}, []parser.StaticDeclarationSyntax{node})
	if !ok || declaration.Player == nil ||
		declaration.Player.Kind != StaticPlayerRuleAttackTax ||
		declaration.Player.AttackTaxGeneric != 2 {
		t.Fatalf("declaration = %#v, ok = %v, want typed attack tax", declaration, ok)
	}
	node.AttackTaxGeneric = 0
	if _, ok := recognizeStaticPlayerRuleDeclaration(CompiledAbility{Kind: AbilityStatic, Content: content}, []parser.StaticDeclarationSyntax{node}); ok {
		t.Fatal("recognized zero attack tax, want fail closed")
	}
}

func TestRecognizeStaticAdditionalLandPlaysFromTypedNodeWithoutInspectingText(t *testing.T) {
	t.Parallel()
	node := parser.StaticDeclarationSyntax{
		Kind:                parser.StaticDeclarationPlayerRule,
		Subject:             parser.StaticDeclarationSubject{Kind: parser.StaticDeclarationSubjectController},
		PlayerRule:          parser.StaticDeclarationPlayerRuleAdditionalLandPlays,
		AdditionalLandPlays: 2,
	}
	ability := CompiledAbility{Kind: AbilityStatic}
	declaration, ok := recognizeStaticPlayerRuleDeclaration(ability, []parser.StaticDeclarationSyntax{node})
	if !ok || declaration.Player == nil ||
		declaration.Player.Kind != StaticPlayerRuleAdditionalLandPlays ||
		declaration.Player.AdditionalLandPlays != 2 {
		t.Fatalf("declaration = %#v, ok = %v, want typed additional land plays", declaration, ok)
	}
	node.AdditionalLandPlays = 0
	if _, ok := recognizeStaticPlayerRuleDeclaration(ability, []parser.StaticDeclarationSyntax{node}); ok {
		t.Fatal("recognized zero additional land plays, want fail closed")
	}
}

func TestRecognizeStaticEachPlayerAdditionalLandPlaysFromTypedNodeWithoutInspectingText(t *testing.T) {
	t.Parallel()
	node := parser.StaticDeclarationSyntax{
		Kind:                parser.StaticDeclarationPlayerRule,
		Subject:             parser.StaticDeclarationSubject{Kind: parser.StaticDeclarationSubjectEachPlayer},
		PlayerRule:          parser.StaticDeclarationPlayerRuleAdditionalLandPlays,
		AdditionalLandPlays: 1,
	}
	ability := CompiledAbility{Kind: AbilityStatic}
	declaration, ok := recognizeStaticPlayerRuleDeclaration(ability, []parser.StaticDeclarationSyntax{node})
	if !ok || declaration.Player == nil ||
		declaration.Player.Kind != StaticPlayerRuleAdditionalLandPlays ||
		declaration.Player.AdditionalLandPlays != 1 ||
		!declaration.Player.AffectsAllPlayers {
		t.Fatalf("declaration = %#v, ok = %v, want each-player additional land plays", declaration, ok)
	}
}

func TestRecognizeStaticEachPlayerSubjectRejectedForControllerOnlyRule(t *testing.T) {
	t.Parallel()
	node := parser.StaticDeclarationSyntax{
		Kind:       parser.StaticDeclarationPlayerRule,
		Subject:    parser.StaticDeclarationSubject{Kind: parser.StaticDeclarationSubjectEachPlayer},
		PlayerRule: parser.StaticDeclarationPlayerRuleNoMaximumHandSize,
	}
	ability := CompiledAbility{Kind: AbilityStatic}
	if _, ok := recognizeStaticPlayerRuleDeclaration(ability, []parser.StaticDeclarationSyntax{node}); ok {
		t.Fatal("recognized each-player subject for a controller-only rule, want fail closed")
	}
}

func TestRecognizeStaticLifeForCommanderTaxFromTypedNodeWithoutInspectingText(t *testing.T) {
	t.Parallel()
	node := parser.StaticDeclarationSyntax{
		Kind:       parser.StaticDeclarationPlayerRule,
		Subject:    parser.StaticDeclarationSubject{Kind: parser.StaticDeclarationSubjectController},
		PlayerRule: parser.StaticDeclarationPlayerRuleLifeForCommanderTax,
	}
	content := AbilityContent{
		References: []CompiledReference{
			{Binding: ReferenceBindingSource, Text: "this spell"},
		},
	}
	ability := CompiledAbility{Kind: AbilityStatic, Content: content}
	declaration, ok := recognizeStaticPlayerRuleDeclaration(ability, []parser.StaticDeclarationSyntax{node})
	if !ok || declaration.Player == nil ||
		declaration.Player.Kind != StaticPlayerRuleLifeForCommanderTax {
		t.Fatalf("declaration = %#v, ok = %v, want typed life-for-commander-tax", declaration, ok)
	}
}

func TestRecognizeStaticLifeForCommanderTaxRejectsForeignReference(t *testing.T) {
	t.Parallel()
	node := parser.StaticDeclarationSyntax{
		Kind:       parser.StaticDeclarationPlayerRule,
		Subject:    parser.StaticDeclarationSubject{Kind: parser.StaticDeclarationSubjectController},
		PlayerRule: parser.StaticDeclarationPlayerRuleLifeForCommanderTax,
	}
	content := AbilityContent{
		References: []CompiledReference{
			{Binding: ReferenceBindingEventPlayer, Text: "they"},
		},
	}
	ability := CompiledAbility{Kind: AbilityStatic, Content: content}
	if _, ok := recognizeStaticPlayerRuleDeclaration(ability, []parser.StaticDeclarationSyntax{node}); ok {
		t.Fatal("recognized life-for-commander-tax with a non-source reference, want fail closed")
	}
}
