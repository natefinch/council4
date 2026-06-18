package parser

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game/types"
)

// parseStaticDeclarationSyntax parses a single static-declaration ability and
// returns the typed declarations the parser emitted. It fails the test when the
// source produced anything other than exactly one ability so meaning tests
// assert on fully typed syntax rather than source text.
func parseStaticDeclarationSyntax(t *testing.T, source string, context Context) []StaticDeclarationSyntax {
	t.Helper()
	document, diagnostics := Parse(source, context)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(document.Abilities) != 1 {
		t.Fatalf("abilities = %#v, want exactly one", document.Abilities)
	}
	return document.Abilities[0].StaticDeclarations
}

func TestParseStaticPowerToughnessDeclarationMeaning(t *testing.T) {
	t.Parallel()
	declarations := parseStaticDeclarationSyntax(t, "This creature gets +1/+2.", Context{})
	if len(declarations) != 1 {
		t.Fatalf("declarations = %#v, want one", declarations)
	}
	declaration := declarations[0]
	if declaration.Kind != StaticDeclarationContinuousPowerToughness {
		t.Fatalf("kind = %v, want power/toughness", declaration.Kind)
	}
	if declaration.Subject.Kind != StaticDeclarationSubjectSourceCreature {
		t.Fatalf("subject = %#v, want source creature", declaration.Subject)
	}
	if declaration.PowerDelta.Value != 1 || declaration.PowerDelta.Negative ||
		declaration.ToughnessDelta.Value != 2 || declaration.ToughnessDelta.Negative ||
		declaration.Dynamic {
		t.Fatalf("declaration = %#v, want +1/+2 static", declaration)
	}
}

func TestParseStaticGroupAnthemSubjectKinds(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		source  string
		kind    EffectStaticSubjectKind
		subtype types.Sub
	}{
		"all creatures": {
			source: "All creatures get +1/+1.",
			kind:   EffectStaticSubjectAllCreatures,
		},
		"all other creatures": {
			source: "All other creatures get -1/-1.",
			kind:   EffectStaticSubjectAllOtherCreatures,
		},
		"attacking creatures": {
			source: "Attacking creatures get -1/-0.",
			kind:   EffectStaticSubjectAttackingCreatures,
		},
		"blocking creatures": {
			source: "Blocking creatures get +0/+2.",
			kind:   EffectStaticSubjectBlockingCreatures,
		},
		"all subtype creatures": {
			source:  "All Sliver creatures have flying.",
			kind:    EffectStaticSubjectAllCreatureSubtype,
			subtype: types.Sliver,
		},
		"other subtype creatures": {
			source:  "Other Soldier creatures get +1/+1.",
			kind:    EffectStaticSubjectOtherCreatureSubtype,
			subtype: types.Soldier,
		},
		"attacking creatures you control": {
			source: "Attacking creatures you control get +1/+0.",
			kind:   EffectStaticSubjectControlledAttackingCreatures,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			declarations := parseStaticDeclarationSyntax(t, test.source, Context{})
			if len(declarations) != 1 {
				t.Fatalf("declarations = %#v, want one", declarations)
			}
			subject := declarations[0].Subject
			if subject.Kind != StaticDeclarationSubjectGroup || subject.Group.Kind != test.kind {
				t.Fatalf("subject = %#v, want group %s", subject, test.kind)
			}
			if test.subtype != "" && subject.Group.Subtype != test.subtype {
				t.Fatalf("subtype = %q, want %q", subject.Group.Subtype, test.subtype)
			}
		})
	}
}

func TestParseStaticGroupPowerToughnessDeclarationMeaning(t *testing.T) {
	t.Parallel()
	declarations := parseStaticDeclarationSyntax(t, "Creatures you control get +1/+1.", Context{})
	if len(declarations) != 1 {
		t.Fatalf("declarations = %#v, want one", declarations)
	}
	declaration := declarations[0]
	if declaration.Kind != StaticDeclarationContinuousPowerToughness ||
		declaration.Subject.Kind != StaticDeclarationSubjectGroup {
		t.Fatalf("declaration = %#v, want group power/toughness", declaration)
	}
}

func TestParseStaticKeywordGrantDeclarationMeaning(t *testing.T) {
	t.Parallel()
	declarations := parseStaticDeclarationSyntax(
		t,
		"As long as you have 7 or more life, this creature has flying.",
		Context{},
	)
	if len(declarations) != 1 {
		t.Fatalf("declarations = %#v, want one", declarations)
	}
	declaration := declarations[0]
	if declaration.Kind != StaticDeclarationKeywordGrant ||
		declaration.Subject.Kind != StaticDeclarationSubjectSourceCreature ||
		!declaration.HasCondition ||
		len(declaration.KeywordSpans) != 1 {
		t.Fatalf("declaration = %#v, want conditional keyword grant", declaration)
	}
}

func TestParseStaticPowerToughnessAndKeywordComposition(t *testing.T) {
	t.Parallel()
	declarations := parseStaticDeclarationSyntax(
		t,
		"Creatures you control get +1/+1 and have trample.",
		Context{},
	)
	if len(declarations) != 2 {
		t.Fatalf("declarations = %#v, want two", declarations)
	}
	if declarations[0].Kind != StaticDeclarationContinuousPowerToughness ||
		declarations[1].Kind != StaticDeclarationKeywordGrant {
		t.Fatalf("declarations = %#v, want PT then keyword", declarations)
	}
	if declarations[0].Dynamic {
		t.Fatalf("composed PT declaration must not be dynamic: %#v", declarations[0])
	}
}

func TestParseStaticMultipleKeywordListComposition(t *testing.T) {
	t.Parallel()
	declarations := parseStaticDeclarationSyntax(
		t,
		"Creatures you control have flying, vigilance, and trample.",
		Context{},
	)
	if len(declarations) != 1 {
		t.Fatalf("declarations = %#v, want one", declarations)
	}
	if declarations[0].Kind != StaticDeclarationKeywordGrant ||
		len(declarations[0].KeywordSpans) != 3 {
		t.Fatalf("declaration = %#v, want three granted keywords", declarations[0])
	}
}

func TestParseStaticMixedSourceDeclarationComposition(t *testing.T) {
	t.Parallel()
	source := "Delirium — As long as there are four or more card types among cards in your graveyard, " +
		"Dragon's Rage Channeler gets +2/+2, has flying, and attacks each combat if able."
	declarations := parseStaticDeclarationSyntax(t, source, Context{CardName: "Dragon's Rage Channeler"})
	if len(declarations) != 3 {
		t.Fatalf("declarations = %#v, want three", declarations)
	}
	if declarations[0].Kind != StaticDeclarationContinuousPowerToughness ||
		declarations[1].Kind != StaticDeclarationKeywordGrant ||
		declarations[2].Kind != StaticDeclarationRule {
		t.Fatalf("declarations = %#v, want PT, keyword, rule", declarations)
	}
	if declarations[2].Rule.Operation.Kind != StaticRuleOperationAttack {
		t.Fatalf("rule = %#v, want attack requirement", declarations[2].Rule)
	}
	for i, declaration := range declarations {
		if declaration.Subject.Kind != StaticDeclarationSubjectSourceNamed || !declaration.HasCondition {
			t.Fatalf("declaration %d = %#v, want conditional self-name subject", i, declaration)
		}
	}
}

func TestParseStaticRuleDeclarationMeaning(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		source    string
		subject   StaticDeclarationSubjectKind
		operation StaticRuleOperationKind
		voice     StaticRuleVoice
	}{
		"cannot block": {
			source:    "This creature can't block.",
			subject:   StaticDeclarationSubjectSourceCreature,
			operation: StaticRuleOperationBlock,
			voice:     StaticRuleVoiceActive,
		},
		"cannot be blocked": {
			source:    "This creature can't be blocked.",
			subject:   StaticDeclarationSubjectSourceCreature,
			operation: StaticRuleOperationBlock,
			voice:     StaticRuleVoicePassive,
		},
		"cannot attack": {
			source:    "This creature can't attack.",
			subject:   StaticDeclarationSubjectSourceCreature,
			operation: StaticRuleOperationAttack,
			voice:     StaticRuleVoiceActive,
		},
		"must attack": {
			source:    "This creature attacks each combat if able.",
			subject:   StaticDeclarationSubjectSourceCreature,
			operation: StaticRuleOperationAttack,
			voice:     StaticRuleVoiceActive,
		},
		"must be blocked": {
			source:    "This creature must be blocked if able.",
			subject:   StaticDeclarationSubjectSourceCreature,
			operation: StaticRuleOperationBlock,
			voice:     StaticRuleVoicePassive,
		},
		"cannot be countered": {
			source:    "This spell can't be countered.",
			subject:   StaticDeclarationSubjectSourceSpell,
			operation: StaticRuleOperationCounter,
			voice:     StaticRuleVoicePassive,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			declarations := parseStaticDeclarationSyntax(t, test.source, Context{})
			if len(declarations) != 1 || declarations[0].Kind != StaticDeclarationRule {
				t.Fatalf("declarations = %#v, want one rule", declarations)
			}
			rule := declarations[0].Rule
			if rule.Operation.Kind != test.operation || rule.Operation.Voice != test.voice {
				t.Fatalf("rule = %#v, want operation %s voice %s", rule, test.operation, test.voice)
			}
			if declarations[0].Subject.Kind != test.subject {
				t.Fatalf("subject = %#v, want %s", declarations[0].Subject, test.subject)
			}
		})
	}
}

func TestParseStaticCostModifierDeclarationMeaning(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		source      string
		modifier    StaticDeclarationCostModifierKind
		reduction   int
		replacement string
	}{
		"ability reduction": {
			source:    "Cycling abilities you activate cost up to {2} less to activate.",
			modifier:  StaticDeclarationCostModifierAbilityReduction,
			reduction: 2,
		},
		"replace cost": {
			source:      "As long as you have seven or more cards in hand, you may pay {0} rather than pay cycling costs.",
			modifier:    StaticDeclarationCostModifierReplaceCost,
			replacement: "",
		},
		"replace first cost": {
			source:      "You may pay {1} rather than pay the cycling cost of the first card you cycle each turn.",
			modifier:    StaticDeclarationCostModifierReplaceFirstCost,
			replacement: "{1}",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			declarations := parseStaticDeclarationSyntax(t, test.source, Context{})
			if len(declarations) != 1 || declarations[0].Kind != StaticDeclarationCostModifier {
				t.Fatalf("declarations = %#v, want one cost modifier", declarations)
			}
			declaration := declarations[0]
			if declaration.CostModifier != test.modifier ||
				declaration.CostReductionAmount != test.reduction ||
				declaration.CostReplacement != test.replacement {
				t.Fatalf("declaration = %#v, want modifier %s", declaration, test.modifier)
			}
		})
	}
}

func TestParseStaticSpellCostModifierDeclarationMeaning(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		source    string
		modifier  StaticDeclarationCostModifierKind
		spellType StaticDeclarationSpellTypeKind
		amount    int
	}{
		"all spells reduction": {
			source:    "Spells you cast cost {1} less to cast.",
			modifier:  StaticDeclarationCostModifierSpellReduction,
			spellType: StaticDeclarationSpellTypeAll,
			amount:    1,
		},
		"artifact spells reduction": {
			source:    "Artifact spells you cast cost {1} less to cast.",
			modifier:  StaticDeclarationCostModifierSpellReduction,
			spellType: StaticDeclarationSpellTypeArtifact,
			amount:    1,
		},
		"creature spells reduction": {
			source:    "Creature spells you cast cost {2} less to cast.",
			modifier:  StaticDeclarationCostModifierSpellReduction,
			spellType: StaticDeclarationSpellTypeCreature,
			amount:    2,
		},
		"enchantment spells reduction": {
			source:    "Enchantment spells you cast cost {1} less to cast.",
			modifier:  StaticDeclarationCostModifierSpellReduction,
			spellType: StaticDeclarationSpellTypeEnchantment,
			amount:    1,
		},
		"instant and sorcery reduction": {
			source:    "Instant and sorcery spells you cast cost {1} less to cast.",
			modifier:  StaticDeclarationCostModifierSpellReduction,
			spellType: StaticDeclarationSpellTypeInstantOrSorcery,
			amount:    1,
		},
		"creature spells increase": {
			source:    "Creature spells you cast cost {1} more to cast.",
			modifier:  StaticDeclarationCostModifierSpellIncrease,
			spellType: StaticDeclarationSpellTypeCreature,
			amount:    1,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			declarations := parseStaticDeclarationSyntax(t, test.source, Context{})
			if len(declarations) != 1 || declarations[0].Kind != StaticDeclarationCostModifier {
				t.Fatalf("declarations = %#v, want one cost modifier", declarations)
			}
			declaration := declarations[0]
			if declaration.CostModifier != test.modifier ||
				declaration.SpellType != test.spellType ||
				declaration.CostReductionAmount != test.amount {
				t.Fatalf("declaration = %#v, want modifier %s spellType %s amount %d", declaration, test.modifier, test.spellType, test.amount)
			}
		})
	}
}

func TestParseStaticSpellCostModifierDeclarationRejections(t *testing.T) {
	t.Parallel()
	sources := map[string]string{
		"subtype filter":       "Dragon spells you cast cost {2} less to cast.",
		"color filter":         "Red spells you cast cost {1} less to cast.",
		"noncreature filter":   "Noncreature spells you cast cost {1} more to cast.",
		"leading condition":    "During turns other than yours, spells you cast cost {1} less to cast.",
		"trailing condition":   "Creature spells you cast cost {1} less to cast if you control a Wizard.",
		"opponents cast":       "Spells your opponents cast cost {1} more to cast.",
		"zero amount":          "Spells you cast cost {0} less to cast.",
		"compound enchantment": "Instant and enchantment spells you cast cost {2} less to cast.",
	}
	for name, source := range sources {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			for _, declaration := range parseStaticDeclarationSyntax(t, source, Context{}) {
				if declaration.Kind == StaticDeclarationCostModifier &&
					(declaration.CostModifier == StaticDeclarationCostModifierSpellReduction ||
						declaration.CostModifier == StaticDeclarationCostModifierSpellIncrease) {
					t.Fatalf("source %q unexpectedly produced a spell cost modifier: %#v", source, declaration)
				}
			}
		})
	}
}

func TestParseStaticCardAbilityGrantDeclarationMeaning(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		source string
		filter StaticDeclarationCardFilterKind
	}{
		"land cards": {
			source: "Each land card in your hand has cycling {2}.",
			filter: StaticDeclarationCardFilterLand,
		},
		"creature cards": {
			source: "Each creature card in your hand has cycling {2}.",
			filter: StaticDeclarationCardFilterCreature,
		},
		"historic cards": {
			source: "Each historic card in your hand has cycling {2}.",
			filter: StaticDeclarationCardFilterHistoric,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			declarations := parseStaticDeclarationSyntax(t, test.source, Context{})
			if len(declarations) != 1 || declarations[0].Kind != StaticDeclarationCardAbilityGrant {
				t.Fatalf("declarations = %#v, want one card-ability grant", declarations)
			}
			if declarations[0].Subject.Kind != StaticDeclarationSubjectControllerHand ||
				declarations[0].Subject.CardFilter != test.filter {
				t.Fatalf("subject = %#v, want hand filter %s", declarations[0].Subject, test.filter)
			}
		})
	}
}

func TestParseStaticBasePowerToughnessDeclarationMeaning(t *testing.T) {
	t.Parallel()
	declarations := parseStaticDeclarationSyntax(
		t,
		"Enchanted creature has base power and toughness 0/2.",
		Context{},
	)
	if len(declarations) != 1 {
		t.Fatalf("declarations = %#v, want one", declarations)
	}
	declaration := declarations[0]
	if declaration.Kind != StaticDeclarationContinuousBasePowerToughness ||
		!declaration.BasePTSet ||
		declaration.BasePower != 0 ||
		declaration.BaseToughness != 2 {
		t.Fatalf("declaration = %#v, want base 0/2", declaration)
	}
	if declaration.Subject.Kind != StaticDeclarationSubjectGroup ||
		declaration.Subject.Group.Kind != EffectStaticSubjectAttachedObject {
		t.Fatalf("subject = %#v, want attached object", declaration.Subject)
	}
}

func TestParseStaticCharacteristicSetColorMeaning(t *testing.T) {
	t.Parallel()
	declarations := parseStaticDeclarationSyntax(
		t,
		"Enchanted creature gets +3/+1 and is black.",
		Context{},
	)
	if len(declarations) != 2 {
		t.Fatalf("declarations = %#v, want two", declarations)
	}
	if declarations[0].Kind != StaticDeclarationContinuousPowerToughness {
		t.Fatalf("declarations[0] = %#v, want power/toughness", declarations[0])
	}
	characteristic := declarations[1]
	if characteristic.Kind != StaticDeclarationContinuousCharacteristic ||
		characteristic.ColorsAdd ||
		!slices.Equal(characteristic.Colors, []Color{ColorBlack}) ||
		len(characteristic.CardTypes) != 0 ||
		len(characteristic.Subtypes) != 0 {
		t.Fatalf("characteristic = %#v, want set-color black", characteristic)
	}
}

func TestParseStaticCharacteristicInAdditionMeaning(t *testing.T) {
	t.Parallel()
	declarations := parseStaticDeclarationSyntax(
		t,
		"Enchanted creature gets -1/-1 and is a black Zombie in addition to its other colors and types.",
		Context{},
	)
	if len(declarations) != 2 {
		t.Fatalf("declarations = %#v, want two", declarations)
	}
	characteristic := declarations[1]
	if characteristic.Kind != StaticDeclarationContinuousCharacteristic ||
		!characteristic.ColorsAdd ||
		!slices.Equal(characteristic.Colors, []Color{ColorBlack}) ||
		!slices.Equal(characteristic.Subtypes, []types.Sub{types.Zombie}) {
		t.Fatalf("characteristic = %#v, want added black Zombie", characteristic)
	}
}

func TestParseStaticCharacteristicTypeInAdditionMeaning(t *testing.T) {
	t.Parallel()
	declarations := parseStaticDeclarationSyntax(
		t,
		"Equipped creature gets +1/+0, has flying, and is a Bird in addition to its other types.",
		Context{},
	)
	if len(declarations) != 3 {
		t.Fatalf("declarations = %#v, want three", declarations)
	}
	if declarations[0].Kind != StaticDeclarationContinuousPowerToughness ||
		declarations[1].Kind != StaticDeclarationKeywordGrant ||
		declarations[2].Kind != StaticDeclarationContinuousCharacteristic {
		t.Fatalf("declarations = %#v, want PT, keyword, characteristic", declarations)
	}
	characteristic := declarations[2]
	if characteristic.ColorsAdd ||
		len(characteristic.Colors) != 0 ||
		!slices.Equal(characteristic.Subtypes, []types.Sub{types.Bird}) {
		t.Fatalf("characteristic = %#v, want added Bird subtype", characteristic)
	}
}

func TestParseStaticGroupColorFilterMeaning(t *testing.T) {
	t.Parallel()
	for name, tc := range map[string]struct {
		source       string
		kind         EffectStaticSubjectKind
		colors       []Color
		colorless    bool
		multicolored bool
	}{
		"leading color": {
			source: "Red creatures you control get +1/+1.",
			kind:   EffectStaticSubjectControlledCreatures,
			colors: []Color{ColorRed},
		},
		"other color": {
			source: "Other white creatures you control get +1/+1.",
			kind:   EffectStaticSubjectOtherControlledCreatures,
			colors: []Color{ColorWhite},
		},
		"colorless qualifier": {
			source:    "Other colorless creatures you control get +0/+1.",
			kind:      EffectStaticSubjectOtherControlledCreatures,
			colorless: true,
		},
		"multicolored qualifier": {
			source:       "Multicolored creatures you control have flying.",
			kind:         EffectStaticSubjectControlledCreatures,
			multicolored: true,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			declarations := parseStaticDeclarationSyntax(t, tc.source, Context{})
			if len(declarations) != 1 {
				t.Fatalf("declarations = %#v, want one", declarations)
			}
			subject := declarations[0].Subject
			if subject.Kind != StaticDeclarationSubjectGroup || subject.Group.Kind != tc.kind {
				t.Fatalf("subject = %#v, want group kind %v", subject, tc.kind)
			}
			if !slices.Equal(subject.Group.Colors, tc.colors) ||
				subject.Group.Colorless != tc.colorless ||
				subject.Group.Multicolored != tc.multicolored {
				t.Fatalf("color filter = %#v, want colors %v colorless %v multicolored %v",
					subject.Group, tc.colors, tc.colorless, tc.multicolored)
			}
		})
	}
}

func TestParseStaticGroupColorFilterFailClosed(t *testing.T) {
	t.Parallel()
	for name, source := range map[string]string{
		"monocolored qualifier": "Monocolored creatures you control get +1/+1.",
		"non color word":        "Sneaky creatures you control get +1/+1.",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(source, Context{})
			if len(document.Abilities) != 1 {
				t.Fatalf("abilities = %#v, want one", document.Abilities)
			}
			for _, declaration := range document.Abilities[0].StaticDeclarations {
				if declaration.Subject.Kind == StaticDeclarationSubjectGroup {
					t.Fatalf("declarations = %#v, want no group declaration (fail closed)",
						document.Abilities[0].StaticDeclarations)
				}
			}
		})
	}
}

func TestParseStaticCharacteristicFailClosed(t *testing.T) {
	t.Parallel()
	for name, source := range map[string]string{
		"bare subtype no in addition":   "Enchanted creature is a Bird.",
		"bare type no in addition":      "Enchanted creature is an artifact.",
		"in addition category mismatch": "Enchanted creature is black in addition to its other types.",
		"base pt negative":              "Enchanted creature has base power and toughness 0/-2.",
		"base pt dynamic":               "Enchanted creature has base power and toughness X/X.",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(source, Context{})
			if len(document.Abilities) != 1 {
				t.Fatalf("abilities = %#v, want one", document.Abilities)
			}
			for _, declaration := range document.Abilities[0].StaticDeclarations {
				if declaration.Kind == StaticDeclarationContinuousBasePowerToughness ||
					declaration.Kind == StaticDeclarationContinuousCharacteristic {
					t.Fatalf("declarations = %#v, want no characteristic declaration (fail closed)",
						document.Abilities[0].StaticDeclarations)
				}
			}
		})
	}
}

func TestParseStaticDeclarationsFailClosed(t *testing.T) {
	t.Parallel()
	for name, source := range map[string]string{
		"unknown verb":             "This creature flickers +1/+1.",
		"dangling connector":       "This creature gets +1/+1 and.",
		"attack missing qualifier": "This creature attacks each combat.",
		"cant attack conditional":  "This creature can't attack unless defending player controls an Island.",
		"must be blocked no able":  "This creature must be blocked.",
		"must be blocked alone":    "This creature must be blocked alone.",
		"unsupported keyword slot": "This creature has +1/+1.",
		"group rule unsupported":   "Creatures you control can't block.",
		"trailing junk":            "This creature gets +1/+1 wobble.",
		"comma without and":        "This creature gets +1/+1, has flying.",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(source, Context{})
			if len(document.Abilities) != 1 {
				t.Fatalf("abilities = %#v, want one", document.Abilities)
			}
			if declarations := document.Abilities[0].StaticDeclarations; len(declarations) != 0 {
				t.Fatalf("declarations = %#v, want none (fail closed)", declarations)
			}
		})
	}
}
