package parser

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/shared"
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

func TestParseStaticNoMaximumHandSizeDeclarationMeaning(t *testing.T) {
	t.Parallel()
	declarations := parseStaticDeclarationSyntax(t, "You have no maximum hand size.", Context{})
	if len(declarations) != 1 {
		t.Fatalf("declarations = %#v, want one", declarations)
	}
	declaration := declarations[0]
	if declaration.Kind != StaticDeclarationPlayerRule {
		t.Fatalf("kind = %v, want player rule", declaration.Kind)
	}
	if declaration.Subject.Kind != StaticDeclarationSubjectController {
		t.Fatalf("subject = %#v, want controller", declaration.Subject)
	}
	if declaration.PlayerRule != StaticDeclarationPlayerRuleNoMaximumHandSize {
		t.Fatalf("player rule = %v, want no maximum hand size", declaration.PlayerRule)
	}
}

func TestParseStaticAttackTaxDeclarationMeaning(t *testing.T) {
	t.Parallel()
	declarations := parseStaticDeclarationSyntax(t,
		"Creatures can't attack you unless their controller pays {2} for each creature they control that's attacking you.",
		Context{})
	if len(declarations) != 1 {
		t.Fatalf("declarations = %#v, want one", declarations)
	}
	declaration := declarations[0]
	if declaration.Kind != StaticDeclarationPlayerRule {
		t.Fatalf("kind = %v, want player rule", declaration.Kind)
	}
	if declaration.Subject.Kind != StaticDeclarationSubjectController {
		t.Fatalf("subject = %#v, want controller", declaration.Subject)
	}
	if declaration.PlayerRule != StaticDeclarationPlayerRuleAttackTax || declaration.AttackTaxGeneric != 2 {
		t.Fatalf("player rule = %v, generic = %d, want attack tax {2}", declaration.PlayerRule, declaration.AttackTaxGeneric)
	}
	if declaration.Span == (shared.Span{}) || declaration.OperationSpan == (shared.Span{}) {
		t.Fatalf("spans = declaration %#v operation %#v, want source spans", declaration.Span, declaration.OperationSpan)
	}
}

func TestParseStaticAttackTaxDeclarationFailsClosed(t *testing.T) {
	t.Parallel()
	tests := map[string]string{
		"conditional":  "As long as you control an artifact, creatures can't attack you unless their controller pays {2} for each creature they control that's attacking you.",
		"planeswalker": "Creatures can't attack you or planeswalkers you control unless their controller pays {2} for each creature they control that's attacking you or a planeswalker you control.",
		"life payment": "Creatures can't attack you unless their controller pays 2 life for each creature they control that's attacking you.",
		"per combat":   "Creatures can't attack you unless their controller pays {2} for each combat.",
		"zero mana":    "Creatures can't attack you unless their controller pays {0} for each creature they control that's attacking you.",
		"signed mana":  "Creatures can't attack you unless their controller pays {+2} for each creature they control that's attacking you.",
		"spaced mana":  "Creatures can't attack you unless their controller pays { 2} for each creature they control that's attacking you.",
		"leading zero": "Creatures can't attack you unless their controller pays {02} for each creature they control that's attacking you.",
	}
	for name, source := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			for _, declaration := range parseStaticDeclarationSyntax(t, source, Context{}) {
				if declaration.PlayerRule == StaticDeclarationPlayerRuleAttackTax {
					t.Fatalf("declaration = %#v, want unsupported near-miss to fail closed", declaration)
				}
			}
		})
	}
}

func TestParseStaticNoMaximumHandSizeRejectsVariants(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"Each player has no maximum hand size.",
		"You have no maximum hand size of seven.",
		"You have a maximum hand size.",
	} {
		document, diagnostics := Parse(source, Context{})
		if len(diagnostics) != 0 || len(document.Abilities) != 1 {
			continue
		}
		for _, declaration := range document.Abilities[0].StaticDeclarations {
			if declaration.Kind == StaticDeclarationPlayerRule {
				t.Fatalf("source %q unexpectedly produced a player-rule declaration", source)
			}
		}
	}
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
		"controlled creature tokens": {
			source: "Creature tokens you control get +1/+1.",
			kind:   EffectStaticSubjectControlledCreatureTokens,
		},
		"battlefield creature tokens": {
			source: "Creature tokens get -1/-1.",
			kind:   EffectStaticSubjectBattlefieldCreatureTokens,
		},
		"controlled legendary creatures": {
			source: "Legendary creatures you control get +2/+2.",
			kind:   EffectStaticSubjectControlledLegendaryCreatures,
		},
		"controlled untapped creatures": {
			source: "Untapped creatures you control get +0/+2.",
			kind:   EffectStaticSubjectControlledUntappedCreatures,
		},
		"other controlled tapped creatures": {
			source: "Other tapped creatures you control have hexproof.",
			kind:   EffectStaticSubjectOtherControlledTappedCreatures,
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

func TestParseStaticPermanentManaAbilityGrantMeaning(t *testing.T) {
	t.Parallel()
	declarations := parseStaticDeclarationSyntax(
		t,
		`Lands you control have "{T}: Add one mana of any color."`,
		Context{},
	)
	if len(declarations) != 1 {
		t.Fatalf("declarations = %#v, want one", declarations)
	}
	declaration := declarations[0]
	if declaration.Kind != StaticDeclarationPermanentAbilityGrant ||
		declaration.Subject.Kind != StaticDeclarationSubjectGroup ||
		declaration.Subject.Group.Kind != EffectStaticSubjectControlledLands {
		t.Fatalf("declaration = %#v, want controlled-land permanent ability grant", declaration)
	}
	granted := declaration.GrantedManaAbility
	if granted == nil || !granted.TapCost || granted.Amount != 1 || !granted.AnyColor {
		t.Fatalf("granted ability = %#v, want tap for one mana of any color", granted)
	}
}

func TestParseStaticPermanentManaAbilityGrantCreatureSubject(t *testing.T) {
	t.Parallel()
	declarations := parseStaticDeclarationSyntax(
		t,
		`Creatures you control have "{T}: Add one mana of any color."`,
		Context{},
	)
	if len(declarations) != 1 {
		t.Fatalf("declarations = %#v, want one", declarations)
	}
	declaration := declarations[0]
	if declaration.Kind != StaticDeclarationPermanentAbilityGrant ||
		declaration.Subject.Kind != StaticDeclarationSubjectGroup ||
		declaration.Subject.Group.Kind != EffectStaticSubjectControlledCreatures {
		t.Fatalf("declaration = %#v, want controlled-creature permanent ability grant", declaration)
	}
	granted := declaration.GrantedManaAbility
	if granted == nil || !granted.TapCost || granted.Amount != 1 || !granted.AnyColor {
		t.Fatalf("granted ability = %#v, want tap for one mana of any color", granted)
	}
}

func TestParseStaticPermanentManaAbilityGrantTreasureSacrifice(t *testing.T) {
	t.Parallel()
	declarations := parseStaticDeclarationSyntax(
		t,
		`Treasures you control have "{T}, Sacrifice this artifact: Add three mana of any one color."`,
		Context{},
	)
	if len(declarations) != 1 {
		t.Fatalf("declarations = %#v, want one", declarations)
	}
	declaration := declarations[0]
	if declaration.Kind != StaticDeclarationPermanentAbilityGrant ||
		declaration.Subject.Group.Kind != EffectStaticSubjectControlledArtifacts ||
		declaration.Subject.Group.Subtype != types.Treasure ||
		!declaration.Subject.Group.SubtypeKnown {
		t.Fatalf("declaration = %#v, want controlled-Treasure permanent ability grant", declaration)
	}
	granted := declaration.GrantedManaAbility
	if granted == nil || !granted.TapCost || granted.Amount != 3 ||
		!granted.Sacrifice || !granted.AnyOneColor || granted.AnyColor ||
		granted.Text != "{T}, Sacrifice this artifact: Add three mana of any one color." {
		t.Fatalf("granted ability = %#v, want tap-sacrifice for three mana of any one color", granted)
	}
}

func TestParseStaticPermanentManaAbilityGrantNearMissesFailClosed(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		`Lands you control have "{T}: Draw a card."`,
		`Lands you control have "{T}: Add two mana of any color."`,
		`Lands you control have "{1}, {T}: Add one mana of any color."`,
		`Lands you control have "{T}: Add one mana of any color." and "{T}: Add {C}."`,
		`As long as you control an artifact, lands you control have "{T}: Add one mana of any color."`,
		`Land cards in your hand have "{T}: Add one mana of any color."`,
		`Enchantments you control have "{T}: Add one mana of any color."`,
		`Treasures you control have "{T}, Sacrifice this artifact: Add one mana of any one color."`,
		`Lands your opponents control have "{T}: Add one mana of any color."`,
	} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			declarations := parseStaticDeclarationSyntax(t, source, Context{})
			for _, declaration := range declarations {
				if declaration.Kind == StaticDeclarationPermanentAbilityGrant {
					t.Fatalf("source %q unexpectedly produced a permanent ability grant: %#v", source, declaration)
				}
			}
		})
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

func TestParseStaticRuleThenKeywordGrantComposition(t *testing.T) {
	t.Parallel()
	declarations := parseStaticDeclarationSyntax(
		t,
		"Equipped creature can't be blocked and has shroud.",
		Context{},
	)
	if len(declarations) != 2 {
		t.Fatalf("declarations = %#v, want two", declarations)
	}
	if declarations[0].Kind != StaticDeclarationRule ||
		declarations[1].Kind != StaticDeclarationKeywordGrant {
		t.Fatalf("declarations = %#v, want rule then keyword", declarations)
	}
	if declarations[0].Rule.Operation.Kind != StaticRuleOperationBlock ||
		declarations[0].Rule.Operation.Voice != StaticRuleVoicePassive {
		t.Fatalf("rule = %#v, want passive block prohibition", declarations[0].Rule)
	}
	for i, declaration := range declarations {
		if declaration.Subject.Kind != StaticDeclarationSubjectGroup ||
			declaration.Subject.Group.Kind != EffectStaticSubjectAttachedObject {
			t.Fatalf("declaration %d = %#v, want attached-object subject", i, declaration)
		}
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
		qualifier StaticRuleQualifierKind
		amount    int
		color     Color
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
		"cannot be blocked by flying": {
			source:    "This creature can't be blocked by creatures with flying.",
			subject:   StaticDeclarationSubjectSourceCreature,
			operation: StaticRuleOperationBlock,
			voice:     StaticRuleVoicePassive,
			qualifier: StaticRuleQualifierBlockerFlying,
		},
		"cannot be blocked by power or less": {
			source:    "This creature can't be blocked by creatures with power 2 or less.",
			subject:   StaticDeclarationSubjectSourceCreature,
			operation: StaticRuleOperationBlock,
			voice:     StaticRuleVoicePassive,
			qualifier: StaticRuleQualifierBlockerPowerOrLess,
			amount:    2,
		},
		"cannot be blocked by power or greater": {
			source:    "This creature can't be blocked by creatures with power 3 or greater.",
			subject:   StaticDeclarationSubjectSourceCreature,
			operation: StaticRuleOperationBlock,
			voice:     StaticRuleVoicePassive,
			qualifier: StaticRuleQualifierBlockerPowerOrGreater,
			amount:    3,
		},
		"cannot be blocked by color": {
			source:    "This creature can't be blocked by black creatures.",
			subject:   StaticDeclarationSubjectSourceCreature,
			operation: StaticRuleOperationBlock,
			voice:     StaticRuleVoicePassive,
			qualifier: StaticRuleQualifierBlockerColor,
			color:     ColorBlack,
		},
		"cannot be blocked by artifact": {
			source:    "This creature can't be blocked by artifact creatures.",
			subject:   StaticDeclarationSubjectSourceCreature,
			operation: StaticRuleOperationBlock,
			voice:     StaticRuleVoicePassive,
			qualifier: StaticRuleQualifierBlockerArtifact,
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
			if test.qualifier != "" && !staticRuleQualifiersAre(rule.Qualifiers, test.qualifier) {
				t.Fatalf("qualifiers = %#v, want %s", rule.Qualifiers, test.qualifier)
			}
			if test.amount != 0 && (len(rule.Qualifiers) != 1 || rule.Qualifiers[0].Amount != test.amount) {
				t.Fatalf("qualifiers = %#v, want amount %d", rule.Qualifiers, test.amount)
			}
			if test.color != "" && (len(rule.Qualifiers) != 1 || rule.Qualifiers[0].Color != test.color) {
				t.Fatalf("qualifiers = %#v, want color %s", rule.Qualifiers, test.color)
			}
		})
	}
}

// TestParseStaticQualifiedRuleDeclarationMeaning covers the bounded-exception
// rule operations that carry a fixed qualifier: the defender-scoped attack
// prohibition and the single-blocker block prohibition. These appear in printed
// cards only as the trailing operation of a composed declaration, so the test
// drives the composed-declaration parser and asserts on the rule node.
func TestParseStaticQualifiedRuleDeclarationMeaning(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		source    string
		operation StaticRuleOperationKind
		voice     StaticRuleVoice
		qualifier StaticRuleQualifierKind
		amount    int
	}{
		"cannot attack you or planeswalkers": {
			source:    "Enchanted creature gets +2/+2 and can't attack you or planeswalkers you control.",
			operation: StaticRuleOperationAttack,
			voice:     StaticRuleVoiceActive,
			qualifier: StaticRuleQualifierDefenderYou,
		},
		"cannot be blocked by more than one": {
			source:    "Enchanted creature gets +1/+2 and can't be blocked by more than one creature.",
			operation: StaticRuleOperationBlock,
			voice:     StaticRuleVoicePassive,
			qualifier: StaticRuleQualifierByMoreThanOne,
		},
		"cannot be blocked by flying": {
			source:    "Enchanted creature gets +1/+2 and can't be blocked by creatures with flying.",
			operation: StaticRuleOperationBlock,
			voice:     StaticRuleVoicePassive,
			qualifier: StaticRuleQualifierBlockerFlying,
		},
		"cannot be blocked by power 2 or less": {
			source:    "Enchanted creature gets +1/+2 and can't be blocked by creatures with power 2 or less.",
			operation: StaticRuleOperationBlock,
			voice:     StaticRuleVoicePassive,
			qualifier: StaticRuleQualifierBlockerPowerOrLess,
			amount:    2,
		},
		"cannot be blocked by power 3 or greater": {
			source:    "Enchanted creature gets +1/+2 and can't be blocked by creatures with power 3 or greater.",
			operation: StaticRuleOperationBlock,
			voice:     StaticRuleVoicePassive,
			qualifier: StaticRuleQualifierBlockerPowerOrGreater,
			amount:    3,
		},
		"cannot be blocked by color": {
			source:    "Enchanted creature gets +1/+2 and can't be blocked by green creatures.",
			operation: StaticRuleOperationBlock,
			voice:     StaticRuleVoicePassive,
			qualifier: StaticRuleQualifierBlockerColor,
		},
		"cannot be blocked by artifact": {
			source:    "Enchanted creature gets +1/+2 and can't be blocked by artifact creatures.",
			operation: StaticRuleOperationBlock,
			voice:     StaticRuleVoicePassive,
			qualifier: StaticRuleQualifierBlockerArtifact,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			declarations := parseStaticDeclarationSyntax(t, test.source, Context{})
			if len(declarations) != 2 || declarations[1].Kind != StaticDeclarationRule {
				t.Fatalf("declarations = %#v, want power/toughness then rule", declarations)
			}
			rule := declarations[1].Rule
			if rule.Operation.Kind != test.operation || rule.Operation.Voice != test.voice {
				t.Fatalf("rule = %#v, want operation %s voice %s", rule, test.operation, test.voice)
			}
			if !staticRuleQualifiersAre(rule.Qualifiers, test.qualifier) {
				t.Fatalf("qualifiers = %#v, want %s", rule.Qualifiers, test.qualifier)
			}
			if test.amount != 0 && (len(rule.Qualifiers) != 1 || rule.Qualifiers[0].Amount != test.amount) {
				t.Fatalf("qualifiers = %#v, want amount %d", rule.Qualifiers, test.amount)
			}
		})
	}
}

// TestParseStaticQualifiedRuleDeclarationNearMiss confirms that near-miss
// phrasings of the bounded-exception rule operations fail the whole composed
// declaration closed rather than silently dropping the trailing qualifier.
func TestParseStaticQualifiedRuleDeclarationNearMiss(t *testing.T) {
	t.Parallel()
	for name, source := range map[string]string{
		"attack you only":           "Enchanted creature gets +2/+2 and can't attack you.",
		"attack planeswalkers only": "Enchanted creature gets +2/+2 and can't attack planeswalkers you control.",
		"blocked by more than two":  "Enchanted creature gets +1/+2 and can't be blocked by more than two creatures.",
		"blocked by toughness":      "Enchanted creature gets +1/+2 and can't be blocked by creatures with toughness 2 or less.",
		"blocked by power no bound": "Enchanted creature gets +1/+2 and can't be blocked by creatures with power.",
		"blocked by noncolor word":  "Enchanted creature gets +1/+2 and can't be blocked by enormous creatures.",
		"blocked by color no noun":  "Enchanted creature gets +1/+2 and can't be blocked by black.",
		"blocked by type plural":    "Enchanted creature gets +1/+2 and can't be blocked by artifacts.",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(source, Context{})
			for _, ability := range document.Abilities {
				for _, declaration := range ability.StaticDeclarations {
					if declaration.Kind == StaticDeclarationRule {
						t.Fatalf("near-miss %q produced a rule declaration: %#v", source, declaration.Rule)
					}
				}
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
		source     string
		modifier   StaticDeclarationCostModifierKind
		spellType  StaticDeclarationSpellTypeKind
		spellColor StaticDeclarationSpellColorKind
		amount     int
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
		"red spells reduction": {
			source:     "Red spells you cast cost {1} less to cast.",
			modifier:   StaticDeclarationCostModifierSpellReduction,
			spellColor: StaticDeclarationSpellColorRed,
			amount:     1,
		},
		"white spells reduction": {
			source:     "White spells you cast cost {1} less to cast.",
			modifier:   StaticDeclarationCostModifierSpellReduction,
			spellColor: StaticDeclarationSpellColorWhite,
			amount:     1,
		},
		"colorless spells reduction": {
			source:     "Colorless spells you cast cost {1} less to cast.",
			modifier:   StaticDeclarationCostModifierSpellReduction,
			spellColor: StaticDeclarationSpellColorColorless,
			amount:     1,
		},
		"green spells increase": {
			source:     "Green spells you cast cost {2} more to cast.",
			modifier:   StaticDeclarationCostModifierSpellIncrease,
			spellColor: StaticDeclarationSpellColorGreen,
			amount:     2,
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
				declaration.SpellColor != test.spellColor ||
				declaration.CostReductionAmount != test.amount {
				t.Fatalf("declaration = %#v, want modifier %s spellType %s spellColor %s amount %d", declaration, test.modifier, test.spellType, test.spellColor, test.amount)
			}
		})
	}
}

func TestParseStaticSpellUncounterableDeclarationMeaning(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		source    string
		spellType StaticDeclarationSpellTypeKind
	}{
		"all spells": {
			source:    "Spells you control can't be countered.",
			spellType: StaticDeclarationSpellTypeAll,
		},
		"creature spells": {
			source:    "Creature spells you control can't be countered.",
			spellType: StaticDeclarationSpellTypeCreature,
		},
		"instant spells": {
			source:    "Instant spells you control can't be countered.",
			spellType: StaticDeclarationSpellTypeInstant,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			declarations := parseStaticDeclarationSyntax(t, test.source, Context{})
			if len(declarations) != 1 || declarations[0].Kind != StaticDeclarationSpellUncounterable {
				t.Fatalf("declarations = %#v, want one spell uncounterable", declarations)
			}
			declaration := declarations[0]
			if declaration.SpellType != test.spellType {
				t.Fatalf("spellType = %s, want %s", declaration.SpellType, test.spellType)
			}
			if declaration.Span == (shared.Span{}) || declaration.OperationSpan == (shared.Span{}) {
				t.Fatalf("spans = declaration %#v operation %#v, want source spans", declaration.Span, declaration.OperationSpan)
			}
		})
	}
}

func TestParseStaticUntapDuringOtherUntapStepDeclarationMeaning(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		source string
		group  StaticUntapGroupKind
	}{
		"all permanents": {
			source: "Untap all permanents you control during each other player's untap step.",
			group:  StaticUntapGroupPermanents,
		},
		"all creatures": {
			source: "Untap all creatures you control during each other player's untap step.",
			group:  StaticUntapGroupCreatures,
		},
		"all artifacts": {
			source: "Untap all artifacts you control during each other player's untap step.",
			group:  StaticUntapGroupArtifacts,
		},
		"all lands": {
			source: "Untap all lands you control during each other player's untap step.",
			group:  StaticUntapGroupLands,
		},
		"opponent wording": {
			source: "Untap all permanents you control during each opponent's untap step.",
			group:  StaticUntapGroupPermanents,
		},
		"self form": {
			source: "Untap this artifact during each other player's untap step.",
			group:  StaticUntapGroupSelf,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			declarations := parseStaticDeclarationSyntax(t, test.source, Context{})
			if len(declarations) != 1 || declarations[0].Kind != StaticDeclarationUntapDuringOtherUntapStep {
				t.Fatalf("declarations = %#v, want one untap-during-other-untap-step", declarations)
			}
			if declarations[0].UntapGroup != test.group {
				t.Fatalf("untapGroup = %s, want %s", declarations[0].UntapGroup, test.group)
			}
		})
	}
}

func TestParseStaticUntapDuringOtherUntapStepRejections(t *testing.T) {
	t.Parallel()
	sources := map[string]string{
		"color filter":    "Untap all green creatures you control during each other player's untap step.",
		"subtype filter":  "Untap all Archers you control during each other player's untap step.",
		"own untap step":  "Untap all permanents you control during your untap step.",
		"not you control": "Untap all permanents during each other player's untap step.",
		"missing each":    "Untap all permanents you control during other player's untap step.",
		"wrong step":      "Untap all permanents you control during each other player's upkeep.",
	}
	for name, source := range sources {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			declarations := parseStaticDeclarationSyntax(t, source, Context{})
			for _, declaration := range declarations {
				if declaration.Kind == StaticDeclarationUntapDuringOtherUntapStep {
					t.Fatalf("source %q recognized as untap-during-other-untap-step, want fail closed", source)
				}
			}
		})
	}
}

func TestParseStaticChosenTypeSpellCostModifierDeclarationMeaning(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		source string
		amount int
	}{
		"you cast qualifier": {source: "Creature spells you cast of the chosen type cost {1} less to cast.", amount: 1},
		"no you cast":        {source: "Creature spells of the chosen type cost {2} less to cast.", amount: 2},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			declarations := parseStaticDeclarationSyntax(t, test.source, Context{})
			if len(declarations) != 1 {
				t.Fatalf("declarations = %#v, want one", declarations)
			}
			declaration := declarations[0]
			if declaration.Kind != StaticDeclarationCostModifier ||
				declaration.CostModifier != StaticDeclarationCostModifierSpellReduction ||
				declaration.SpellType != StaticDeclarationSpellTypeCreature ||
				!declaration.ChosenCreatureType ||
				declaration.CostReductionAmount != test.amount {
				t.Fatalf("declaration = %#v, want chosen creature type spell reduction of %d", declaration, test.amount)
			}
			if declaration.Span == (shared.Span{}) || declaration.OperationSpan == (shared.Span{}) {
				t.Fatalf("spans = declaration %#v operation %#v, want source spans", declaration.Span, declaration.OperationSpan)
			}
		})
	}
}

func TestParseStaticSpellCostModifierDeclarationRejections(t *testing.T) {
	t.Parallel()
	sources := map[string]string{
		"subtype filter":       "Dragon spells you cast cost {2} less to cast.",
		"multicolored filter":  "Multicolored spells you cast cost {1} less to cast.",
		"noncreature filter":   "Noncreature spells you cast cost {1} more to cast.",
		"leading condition":    "During turns other than yours, spells you cast cost {1} less to cast.",
		"trailing condition":   "Creature spells you cast cost {1} less to cast if you control a Wizard.",
		"opponents cast":       "Spells your opponents cast cost {1} more to cast.",
		"zero amount":          "Spells you cast cost {0} less to cast.",
		"compound enchantment": "Instant and enchantment spells you cast cost {2} less to cast.",
		"color and type":       "Red creature spells you cast cost {1} less to cast.",
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

func TestParseStaticCharacteristicSetAllColorsMeaning(t *testing.T) {
	t.Parallel()
	allColors := []Color{ColorWhite, ColorBlue, ColorBlack, ColorRed, ColorGreen}
	for name, tc := range map[string]struct {
		source  string
		context Context
		subject StaticDeclarationSubjectKind
	}{
		"this creature": {
			source:  "This creature is all colors.",
			context: Context{},
			subject: StaticDeclarationSubjectSourceCreature,
		},
		"named source": {
			source:  "Transguild Courier is all colors.",
			context: Context{CardName: "Transguild Courier"},
			subject: StaticDeclarationSubjectSourceNamed,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			declarations := parseStaticDeclarationSyntax(t, tc.source, tc.context)
			if len(declarations) != 1 {
				t.Fatalf("declarations = %#v, want one", declarations)
			}
			characteristic := declarations[0]
			if characteristic.Kind != StaticDeclarationContinuousCharacteristic ||
				characteristic.Subject.Kind != tc.subject ||
				characteristic.ColorsAdd ||
				!slices.Equal(characteristic.Colors, allColors) ||
				len(characteristic.CardTypes) != 0 ||
				len(characteristic.Subtypes) != 0 {
				t.Fatalf("characteristic = %#v, want set all colors for subject %s", characteristic, tc.subject)
			}
		})
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

func TestParseStaticChosenCreatureTypeAddition(t *testing.T) {
	t.Parallel()
	declarations := parseStaticDeclarationSyntax(
		t,
		"This creature is the chosen type in addition to its other types.",
		Context{},
	)
	if len(declarations) != 1 {
		t.Fatalf("declarations = %#v, want one", declarations)
	}
	declaration := declarations[0]
	if declaration.Kind != StaticDeclarationContinuousEntryChoiceSubtype ||
		declaration.Subject.Kind != StaticDeclarationSubjectSourceCreature {
		t.Fatalf("declaration = %#v, want source chosen-subtype addition", declaration)
	}
}

func TestParseStaticChosenCreatureTypeTriggerMultiplier(t *testing.T) {
	t.Parallel()
	declarations := parseStaticDeclarationSyntax(
		t,
		"If a triggered ability of another creature you control of the chosen type triggers, it triggers an additional time.",
		Context{},
	)
	if len(declarations) != 1 {
		t.Fatalf("declarations = %#v, want one", declarations)
	}
	if declarations[0].Kind != StaticDeclarationChosenCreatureTypeTriggerMultiplier {
		t.Fatalf("declaration = %#v, want chosen-type trigger multiplier", declarations[0])
	}
}

func TestParseStaticEnteringTriggerMultiplier(t *testing.T) {
	t.Parallel()
	for name, tc := range map[string]struct {
		source string
		filter []CardType
	}{
		"artifact or creature": {
			source: "If an artifact or creature entering causes a triggered ability of a permanent you control to trigger, that ability triggers an additional time.",
			filter: []CardType{CardTypeArtifact, CardTypeCreature},
		},
		"any permanent": {
			source: "If a permanent entering causes a triggered ability of a permanent you control to trigger, that ability triggers an additional time.",
			filter: nil,
		},
		"land": {
			source: "If a land entering causes a triggered ability of a permanent you control to trigger, that ability triggers an additional time.",
			filter: []CardType{CardTypeLand},
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			declarations := parseStaticDeclarationSyntax(t, tc.source, Context{})
			if len(declarations) != 1 {
				t.Fatalf("declarations = %#v, want one", declarations)
			}
			declaration := declarations[0]
			if declaration.Kind != StaticDeclarationEnteringTriggerMultiplier {
				t.Fatalf("declaration kind = %v, want entering-trigger multiplier", declaration.Kind)
			}
			if !slices.Equal(declaration.EnteringFilterTypes, tc.filter) {
				t.Fatalf("filter = %#v, want %#v", declaration.EnteringFilterTypes, tc.filter)
			}
		})
	}
}

func TestParseStaticEnteringTriggerMultiplierFailsClosed(t *testing.T) {
	t.Parallel()
	for name, source := range map[string]string{
		"subtype filter":   "If a Wizard you control entering causes a triggered ability of a permanent you control to trigger, that ability triggers an additional time.",
		"twice wording":    "If an artifact or creature entering causes a triggered ability of a permanent you control to trigger, that ability triggers twice.",
		"opponent control": "If a permanent entering causes a triggered ability of a permanent an opponent controls to trigger, that ability triggers an additional time.",
		"missing filter":   "If entering causes a triggered ability of a permanent you control to trigger, that ability triggers an additional time.",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			for _, declaration := range parseStaticDeclarationSyntax(t, source, Context{}) {
				if declaration.Kind == StaticDeclarationEnteringTriggerMultiplier {
					t.Fatalf("declaration = %#v, want fail-closed near miss", declaration)
				}
			}
		})
	}
}

func TestParseStaticChosenCreatureTypeDeclarationsFailClosedOnNearMisses(t *testing.T) {
	t.Parallel()
	for name, source := range map[string]string{
		"indefinite chosen type": "This creature is a chosen type in addition to its other types.",
		"creature types tail":    "This creature is the chosen type in addition to its other creature types.",
		"missing another":        "If a triggered ability of a creature you control of the chosen type triggers, it triggers an additional time.",
		"permanent source":       "If a triggered ability of another permanent you control of the chosen type triggers, it triggers an additional time.",
		"twice wording":          "If a triggered ability of another creature you control of the chosen type triggers, it triggers twice.",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if declarations := parseStaticDeclarationSyntax(t, source, Context{}); len(declarations) != 0 {
				t.Fatalf("declarations = %#v, want fail-closed near miss", declarations)
			}
		})
	}
}

func TestParseStaticChosenTypeAnthemSubjectKinds(t *testing.T) {
	t.Parallel()
	for name, tc := range map[string]struct {
		source string
		kind   EffectStaticSubjectKind
	}{
		"controlled chosen type": {
			source: "Creatures you control of the chosen type get +1/+1.",
			kind:   EffectStaticSubjectControlledCreaturesChosenType,
		},
		"other controlled chosen type": {
			source: "Other creatures you control of the chosen type get +1/+1.",
			kind:   EffectStaticSubjectOtherControlledCreaturesChosenType,
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
		})
	}
}

func TestParseStaticChosenTypeAnthemFailsClosed(t *testing.T) {
	t.Parallel()
	for name, source := range map[string]string{
		"indefinite chosen type": "Creatures you control of a chosen type get +1/+1.",
		"missing you control":    "Creatures of the chosen type get +1/+1.",
		"plural chosen types":    "Creatures you control of the chosen types get +1/+1.",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			for _, declaration := range parseStaticDeclarationSyntax(t, source, Context{}) {
				if declaration.Subject.Kind == StaticDeclarationSubjectGroup &&
					(declaration.Subject.Group.Kind == EffectStaticSubjectControlledCreaturesChosenType ||
						declaration.Subject.Group.Kind == EffectStaticSubjectOtherControlledCreaturesChosenType) {
					t.Fatalf("source %q unexpectedly parsed as chosen-type group: %#v", source, declaration)
				}
			}
		})
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
		"battlefield leading color": {
			source: "White creatures get +1/+1.",
			kind:   EffectStaticSubjectAllCreatures,
			colors: []Color{ColorWhite},
		},
		"battlefield other color": {
			source: "Other black creatures get -1/-1.",
			kind:   EffectStaticSubjectAllOtherCreatures,
			colors: []Color{ColorBlack},
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

func TestParseStaticGroupKeywordFilterMeaning(t *testing.T) {
	t.Parallel()
	for name, tc := range map[string]struct {
		source          string
		kind            EffectStaticSubjectKind
		keyword         KeywordKind
		excludedKeyword KeywordKind
	}{
		"battlefield with keyword": {
			source:  "Creatures with flying get +1/+1.",
			kind:    EffectStaticSubjectAllCreatures,
			keyword: KeywordFlying,
		},
		"battlefield without keyword": {
			source:          "Creatures without flying get -2/-0.",
			kind:            EffectStaticSubjectAllCreatures,
			excludedKeyword: KeywordFlying,
		},
		"controlled with keyword": {
			source:  "Creatures you control with flying get +1/+1.",
			kind:    EffectStaticSubjectControlledCreatures,
			keyword: KeywordFlying,
		},
		"other controlled with keyword": {
			source:  "Other creatures you control with flying get +1/+1.",
			kind:    EffectStaticSubjectOtherControlledCreatures,
			keyword: KeywordFlying,
		},
		"opponent with keyword": {
			source:  "Creatures with flying your opponents control get -1/-0.",
			kind:    EffectStaticSubjectOpponentControlledCreatures,
			keyword: KeywordFlying,
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
			if subject.Group.Keyword != tc.keyword || subject.Group.ExcludedKeyword != tc.excludedKeyword {
				t.Fatalf("keyword filter = %#v, want keyword %v excluded %v",
					subject.Group, tc.keyword, tc.excludedKeyword)
			}
		})
	}
}

func TestParseStaticGroupTypeFilterMeaning(t *testing.T) {
	t.Parallel()
	for name, tc := range map[string]struct {
		source string
		kind   EffectStaticSubjectKind
	}{
		"artifact creatures you control": {
			source: "Artifact creatures you control get +1/+1.",
			kind:   EffectStaticSubjectControlledArtifactCreatures,
		},
		"other artifact creatures you control": {
			source: "Other artifact creatures you control get +1/+1.",
			kind:   EffectStaticSubjectOtherControlledArtifactCreatures,
		},
		"nontoken creatures you control": {
			source: "Nontoken creatures you control get +1/+1.",
			kind:   EffectStaticSubjectControlledNontokenCreatures,
		},
		"other nontoken creatures you control": {
			source: "Other nontoken creatures you control get +1/+1.",
			kind:   EffectStaticSubjectOtherControlledNontokenCreatures,
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
		})
	}
}

func TestParseStaticGroupKeywordFilterFailClosed(t *testing.T) {
	t.Parallel()
	for name, source := range map[string]string{
		"with non keyword word":  "Creatures with sneakiness get +1/+1.",
		"with parametrized form": "Creatures with a flying ability get +1/+1.",
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

// TestParseStaticComposedPowerToughnessRuleAttachedSubject verifies that the
// compound declaration path accepts attached-object subjects for a single
// creature rule operation alongside a continuous power/toughness change.
func TestParseStaticComposedPowerToughnessRuleAttachedSubject(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		source    string
		operation StaticRuleOperationKind
		voice     StaticRuleVoice
		attached  bool
	}{
		"enchanted can't block": {
			source:    "Enchanted creature gets +2/+2 and can't block.",
			operation: StaticRuleOperationBlock,
			voice:     StaticRuleVoiceActive,
			attached:  true,
		},
		"enchanted can't be blocked": {
			source:    "Enchanted creature gets +1/+0 and can't be blocked.",
			operation: StaticRuleOperationBlock,
			voice:     StaticRuleVoicePassive,
			attached:  true,
		},
		"equipped must attack": {
			source:    "Equipped creature gets +2/+2 and attacks each combat if able.",
			operation: StaticRuleOperationAttack,
			voice:     StaticRuleVoiceActive,
			attached:  true,
		},
		"source can't block": {
			source:    "This creature gets +2/+2 and can't block.",
			operation: StaticRuleOperationBlock,
			voice:     StaticRuleVoiceActive,
			attached:  false,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			declarations := parseStaticDeclarationSyntax(t, test.source, Context{})
			if len(declarations) != 2 ||
				declarations[0].Kind != StaticDeclarationContinuousPowerToughness ||
				declarations[1].Kind != StaticDeclarationRule {
				t.Fatalf("declarations = %#v, want [PT, rule]", declarations)
			}
			rule := declarations[1].Rule
			if rule.Operation.Kind != test.operation || rule.Operation.Voice != test.voice {
				t.Fatalf("rule = %#v, want operation %s voice %s", rule, test.operation, test.voice)
			}
			for i, declaration := range declarations {
				if test.attached {
					if declaration.Subject.Kind != StaticDeclarationSubjectGroup ||
						declaration.Subject.Group.Kind != EffectStaticSubjectAttachedObject {
						t.Fatalf("declaration %d subject = %#v, want attached object", i, declaration.Subject)
					}
				} else if declaration.Subject.Kind != StaticDeclarationSubjectSourceCreature {
					t.Fatalf("declaration %d subject = %#v, want source creature", i, declaration.Subject)
				}
			}
		})
	}
}

// TestParseStaticComposedPowerToughnessRuleGroupFailClosed confirms the compound
// path still rejects battlefield-group subjects for rule operations, which need
// runtime group plumbing that does not yet exist.
func TestParseStaticComposedPowerToughnessRuleGroupFailClosed(t *testing.T) {
	t.Parallel()
	document, _ := Parse("Other creatures you control get +1/+1 and can't block.", Context{})
	if len(document.Abilities) == 0 {
		return
	}
	for _, declaration := range document.Abilities[0].StaticDeclarations {
		if declaration.Kind == StaticDeclarationRule {
			t.Fatalf("declaration = %#v, want no rule declaration for battlefield group (fail closed)", declaration)
		}
	}
}

// TestParseStaticLoseAbilitiesBecomeMeaning confirms the polymorph static shape
// "<subject> loses all abilities and is a <colors> <subtype> creature with base
// power and toughness N/N" types fully: the affected object, the loss flag, and
// the set colors, card type, subtype, and base power/toughness.
func TestParseStaticLoseAbilitiesBecomeMeaning(t *testing.T) {
	t.Parallel()
	declarations := parseStaticDeclarationSyntax(t,
		"Enchanted creature loses all abilities and is a blue Frog creature with base power and toughness 1/1.",
		Context{})
	if len(declarations) != 1 {
		t.Fatalf("declarations = %#v, want one", declarations)
	}
	declaration := declarations[0]
	if declaration.Kind != StaticDeclarationLoseAbilitiesBecome {
		t.Fatalf("kind = %v, want lose-abilities-become", declaration.Kind)
	}
	if !declaration.LoseAllAbilities {
		t.Fatal("LoseAllAbilities = false, want true")
	}
	if declaration.Subject.Kind != StaticDeclarationSubjectGroup ||
		declaration.Subject.Group.Kind != EffectStaticSubjectAttachedObject {
		t.Fatalf("subject = %#v, want attached-object group", declaration.Subject)
	}
	if !slices.Equal(declaration.Colors, []Color{ColorBlue}) {
		t.Fatalf("colors = %#v, want blue", declaration.Colors)
	}
	if !slices.Equal(declaration.CardTypes, []CardType{CardTypeCreature}) {
		t.Fatalf("card types = %#v, want creature", declaration.CardTypes)
	}
	if !slices.Equal(declaration.Subtypes, []types.Sub{types.Frog}) {
		t.Fatalf("subtypes = %#v, want Frog", declaration.Subtypes)
	}
	if !declaration.BasePTSet || declaration.BasePower != 1 || declaration.BaseToughness != 1 {
		t.Fatalf("base P/T = set %v %d/%d, want set 1/1", declaration.BasePTSet, declaration.BasePower, declaration.BaseToughness)
	}
}

// TestParseStaticLoseAbilitiesBecomeHasOnly confirms the bare "has base power
// and toughness N/N" tail types as a base-P/T set with no color or type change.
func TestParseStaticLoseAbilitiesBecomeHasOnly(t *testing.T) {
	t.Parallel()
	declarations := parseStaticDeclarationSyntax(t,
		"Enchanted creature loses all abilities and has base power and toughness 1/1.",
		Context{})
	if len(declarations) != 1 {
		t.Fatalf("declarations = %#v, want one", declarations)
	}
	declaration := declarations[0]
	if declaration.Kind != StaticDeclarationLoseAbilitiesBecome || !declaration.LoseAllAbilities {
		t.Fatalf("declaration = %#v, want lose-abilities-become", declaration)
	}
	if len(declaration.Colors) != 0 || len(declaration.CardTypes) != 0 || len(declaration.Subtypes) != 0 {
		t.Fatalf("declaration set characteristics = %#v, want none", declaration)
	}
	if !declaration.BasePTSet || declaration.BasePower != 1 || declaration.BaseToughness != 1 {
		t.Fatalf("base P/T = set %v %d/%d, want set 1/1", declaration.BasePTSet, declaration.BasePower, declaration.BaseToughness)
	}
}

// TestParseStaticLoseAbilitiesBecomeRejectsVariants confirms the polymorph parser
// fails closed for a name-setting tail and a colorless body, neither of which the
// continuous machinery can model as a set characteristic.
func TestParseStaticLoseAbilitiesBecomeRejectsVariants(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"Enchanted creature loses all abilities and is a green and white Citizen creature with base power and toughness 1/1 named Legitimate Businessperson.",
		"Enchanted creature loses all abilities and is a colorless Noggle with base power and toughness 1/1.",
	} {
		document, diagnostics := Parse(source, Context{})
		if len(diagnostics) != 0 || len(document.Abilities) != 1 {
			continue
		}
		for _, declaration := range document.Abilities[0].StaticDeclarations {
			if declaration.Kind == StaticDeclarationLoseAbilitiesBecome {
				t.Fatalf("source %q unexpectedly produced a lose-abilities-become declaration", source)
			}
		}
	}
}

// TestParseStaticAllLandsTypeAdditionMeaning verifies the continuous
// land-type-adding statics printed on Yavimaya, Cradle of Growth and Urborg,
// Tomb of Yawgmoth parse into a single additive characteristic declaration whose
// affected group is every land on the battlefield.
func TestParseStaticAllLandsTypeAdditionMeaning(t *testing.T) {
	t.Parallel()
	for name, tc := range map[string]struct {
		source  string
		subtype types.Sub
	}{
		"each land forest": {
			source:  "Each land is a Forest in addition to its other land types.",
			subtype: types.Forest,
		},
		"each land swamp": {
			source:  "Each land is a Swamp in addition to its other land types.",
			subtype: types.Swamp,
		},
		"all lands islands": {
			source:  "All lands are Islands in addition to their other land types.",
			subtype: types.Island,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			declarations := parseStaticDeclarationSyntax(t, tc.source, Context{})
			if len(declarations) != 1 {
				t.Fatalf("declarations = %#v, want one", declarations)
			}
			declaration := declarations[0]
			if declaration.Kind != StaticDeclarationContinuousCharacteristic {
				t.Fatalf("kind = %v, want continuous characteristic", declaration.Kind)
			}
			if declaration.Subject.Kind != StaticDeclarationSubjectGroup ||
				declaration.Subject.Group.Kind != EffectStaticSubjectAllLands {
				t.Fatalf("subject = %#v, want all-lands group", declaration.Subject)
			}
			if declaration.ColorsAdd || len(declaration.Colors) != 0 || len(declaration.CardTypes) != 0 {
				t.Fatalf("declaration = %#v, want no color or card-type payload", declaration)
			}
			if !slices.Equal(declaration.Subtypes, []types.Sub{tc.subtype}) {
				t.Fatalf("subtypes = %#v, want %v", declaration.Subtypes, tc.subtype)
			}
		})
	}
}

// TestParseStaticAllLandsTypeAdditionRejectsVariants confirms the all-lands
// land-type static fails closed outside the exact "each/all land(s) is/are a
// <basic land type> in addition to its/their other land types" shape: a missing
// "in addition" tail, a non-basic subtype, and a power/toughness operation all
// fail to produce a characteristic declaration with the all-lands group.
func TestParseStaticAllLandsTypeAdditionRejectsVariants(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"Each land is a Forest.",
		"Each land is a Goblin in addition to its other land types.",
		"Each land is a 0/0 creature in addition to its other land types.",
		"All lands get +1/+0.",
	} {
		document, diagnostics := Parse(source, Context{})
		if len(diagnostics) != 0 || len(document.Abilities) != 1 {
			continue
		}
		for _, declaration := range document.Abilities[0].StaticDeclarations {
			if declaration.Kind == StaticDeclarationContinuousCharacteristic &&
				declaration.Subject.Group.Kind == EffectStaticSubjectAllLands &&
				len(declaration.Subtypes) != 0 {
				t.Fatalf("source %q unexpectedly produced an all-lands land-type declaration", source)
			}
		}
	}
}
