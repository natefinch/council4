package parser

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

type cachedParserCard struct {
	Name       string       `json:"name"`
	OracleText string       `json:"oracle_text"`
	CardFaces  []cachedFace `json:"card_faces"`
	TypeLine   string       `json:"type_line"`
}

type cachedFace struct {
	Name       string `json:"name"`
	TypeLine   string `json:"type_line"`
	OracleText string `json:"oracle_text"`
}

func TestParseAbilityKinds(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		source  string
		context Context
		want    AbilityKind
	}{
		"spell": {
			source:  "Destroy target creature.",
			context: Context{InstantOrSorcery: true},
			want:    AbilitySpell,
		},
		"activated": {
			source: "{T}: Add {G}.",
			want:   AbilityActivated,
		},
		"loyalty": {
			source:  "−2: Target creature you control fights target creature you don't control.",
			context: Context{Planeswalker: true},
			want:    AbilityLoyalty,
		},
		"variable loyalty": {
			source:  "+X: Draw X cards.",
			context: Context{Planeswalker: true},
			want:    AbilityLoyalty,
		},
		"numeric activated": {
			source: "2: Draw a card.",
			want:   AbilityActivated,
		},
		"triggered": {
			source: "Whenever you attack, draw a card.",
			want:   AbilityTriggered,
		},
		"ability word trigger": {
			source: "Formidable — Whenever you attack, draw a card.",
			want:   AbilityTriggered,
		},
		"saga chapter": {
			source: "I, II — Draw a card.",
			context: Context{
				Saga: true,
			},
			want: AbilityChapter,
		},
		"replacement": {
			source: "This land enters tapped.",
			want:   AbilityReplacement,
		},
		"static": {
			source: "Creatures you control have haste.",
			want:   AbilityStatic,
		},
		"reminder": {
			source: "(This creature can block creatures with flying.)",
			want:   AbilityReminder,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(test.source, test.context)
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			if len(document.Abilities) != 1 {
				t.Fatalf("abilities = %d", len(document.Abilities))
			}
			if got := document.Abilities[0].Kind; got != test.want {
				t.Fatalf("kind = %s, want %s", got, test.want)
			}
		})
	}
}

func TestParseTypedActivationRestrictions(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		restriction string
		kind        ActivationRestrictionKind
		count       ActivationFrequencyCountKind
		period      ActivationFrequencyPeriodKind
		quantifier  PhaseStepQuantifierKind
		player      TriggerPlayerSelectorKind
		phaseStep   PhaseStepNameKind
	}{
		{"sorcery timing", "Activate only as a sorcery.", ActivationRestrictionSorceryTiming, ActivationFrequencyCountUnknown, ActivationFrequencyPeriodUnknown, PhaseStepQuantifierUnknown, TriggerPlayerSelectorUnknown, PhaseStepNameUnknown},
		{"once each turn", "Activate only once each turn.", ActivationRestrictionFrequency, ActivationFrequencyCountOnce, ActivationFrequencyPeriodTurn, PhaseStepQuantifierUnknown, TriggerPlayerSelectorUnknown, PhaseStepNameUnknown},
		{"combat", "Activate only during combat.", ActivationRestrictionPhaseStep, ActivationFrequencyCountUnknown, ActivationFrequencyPeriodUnknown, PhaseStepQuantifierNone, TriggerPlayerSelectorAny, PhaseStepNameCombat},
		{"controller upkeep", "Activate only during your upkeep.", ActivationRestrictionPhaseStep, ActivationFrequencyCountUnknown, ActivationFrequencyPeriodUnknown, PhaseStepQuantifierSingle, TriggerPlayerSelectorYou, PhaseStepNameUpkeep},
		{"typed unsupported phase", "Activate only during your end step.", ActivationRestrictionPhaseStep, ActivationFrequencyCountUnknown, ActivationFrequencyPeriodUnknown, PhaseStepQuantifierSingle, TriggerPlayerSelectorYou, PhaseStepNameEndStep},
		{"explicit unsupported", "Activate only before combat.", ActivationRestrictionUnsupported, ActivationFrequencyCountUnknown, ActivationFrequencyPeriodUnknown, PhaseStepQuantifierUnknown, TriggerPlayerSelectorUnknown, PhaseStepNameUnknown},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			source := "{1}: Draw a card. " + test.restriction
			document, diagnostics := Parse(source, Context{})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			restrictions := document.Abilities[0].ActivationRestrictions
			if len(restrictions) != 1 {
				t.Fatalf("restrictions = %#v, want one", restrictions)
			}
			restriction := restrictions[0]
			if restriction.Kind != test.kind ||
				restriction.Frequency.Count.Kind != test.count ||
				restriction.Frequency.Period.Kind != test.period ||
				restriction.PhaseStep.Quantifier.Kind != test.quantifier ||
				restriction.PhaseStep.Player.Kind != test.player ||
				restriction.PhaseStep.Name.Kind != test.phaseStep {
				t.Fatalf("restriction = %#v", restriction)
			}
			assertTextSpan(t, "activation restriction", source, restriction.Span, test.restriction)
			switch restriction.Kind {
			case ActivationRestrictionSorceryTiming:
				assertSpanContains(t, "sorcery timing", restriction.Span, restriction.SorcerySpan)
			case ActivationRestrictionFrequency:
				assertSpanContains(t, "frequency count", restriction.Span, restriction.Frequency.Count.Span)
				assertSpanContains(t, "frequency period", restriction.Span, restriction.Frequency.Period.Span)
			case ActivationRestrictionPhaseStep:
				assertSpanContains(t, "phase/step name", restriction.Span, restriction.PhaseStep.Name.Span)
			default:
			}
		})
	}
}

func TestParseActivationRestrictionGrammarVariants(t *testing.T) {
	t.Parallel()
	for _, restriction := range []string{
		"Activate only at sorcery speed.",
		"Activate only any time you could cast a sorcery.",
		"Activate only once per turn.",
		"Activate only one time every turn.",
		"Activate only during each combat.",
		"Activate only during each of your upkeeps.",
	} {
		t.Run(restriction, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse("{1}: Draw a card. "+restriction, Context{})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			restrictions := document.Abilities[0].ActivationRestrictions
			if len(restrictions) != 1 || restrictions[0].Kind == ActivationRestrictionUnsupported {
				t.Fatalf("restrictions = %#v, want one supported typed restriction", restrictions)
			}
		})
	}
}

func TestParseComposedActivationRestrictions(t *testing.T) {
	t.Parallel()
	source := "{1}: Draw a card. (Before.) Activate only once per turn. (Between.) Activate only at sorcery speed. (After.)"
	document, diagnostics := Parse(source, Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	restrictions := document.Abilities[0].ActivationRestrictions
	if len(restrictions) != 2 ||
		restrictions[0].Kind != ActivationRestrictionFrequency ||
		restrictions[1].Kind != ActivationRestrictionSorceryTiming {
		t.Fatalf("restrictions = %#v", restrictions)
	}
}

func TestParseActivationRestrictionsFailClosed(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"{1}: Draw a card. Activate only as an instant.",
		"{1}: Draw a card. Activate only once each round.",
		"{1}: Draw a card. Activate only during your next upkeep.",
		"{1}: Draw a card. Activate only during combat on your turn.",
	} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(source, Context{})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			restrictions := document.Abilities[0].ActivationRestrictions
			if len(restrictions) != 1 || restrictions[0].Kind != ActivationRestrictionUnsupported {
				t.Fatalf("restrictions = %#v, want explicit unsupported restriction", restrictions)
			}
		})
	}
	for _, source := range []string{
		"{1}: Draw a card. Activate only if you control a creature.",
		"{1}: Activate only as a sorcery. Draw a card.",
		"Activate only as a sorcery.",
		"{1}: Draw a card. Activate as a sorcery.",
	} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(source, Context{})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			if restrictions := document.Abilities[0].ActivationRestrictions; len(restrictions) != 0 {
				t.Fatalf("restrictions = %#v, want none", restrictions)
			}
		})
	}
}

func TestParsePhaseStepTriggerClauses(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		event      string
		quantifier PhaseStepQuantifierKind
		player     TriggerPlayerSelectorKind
		phaseStep  PhaseStepNameKind
		attached   TriggerSelection
	}{
		{"standalone end of combat", "end of combat", PhaseStepQuantifierNone, TriggerPlayerSelectorAny, PhaseStepNameEndOfCombat, TriggerSelection{}},
		{"source controller upkeep", "the beginning of its controller's upkeep", PhaseStepQuantifierSingle, TriggerPlayerSelectorSourceController, PhaseStepNameUpkeep, TriggerSelection{}},
		{"your draw step", "the beginning of your draw step", PhaseStepQuantifierSingle, TriggerPlayerSelectorYou, PhaseStepNameDrawStep, TriggerSelection{}},
		{"each end step", "the beginning of each end step", PhaseStepQuantifierEach, TriggerPlayerSelectorAny, PhaseStepNameEndStep, TriggerSelection{}},
		{"each player upkeep", "the beginning of each player's upkeep", PhaseStepQuantifierEach, TriggerPlayerSelectorAny, PhaseStepNameUpkeep, TriggerSelection{}},
		{"each opponent draw step", "the beginning of each opponent's draw step", PhaseStepQuantifierEach, TriggerPlayerSelectorOpponent, PhaseStepNameDrawStep, TriggerSelection{}},
		{"combat on your turn", "the beginning of combat on your turn", PhaseStepQuantifierSingle, TriggerPlayerSelectorYou, PhaseStepNameCombat, TriggerSelection{}},
		{"combat on each turn", "the beginning of combat on each turn", PhaseStepQuantifierEach, TriggerPlayerSelectorAny, PhaseStepNameCombat, TriggerSelection{}},
		{"end combat on your turn", "the beginning of the end of combat on your turn", PhaseStepQuantifierSingle, TriggerPlayerSelectorYou, PhaseStepNameEndOfCombat, TriggerSelection{}},
		{"each end combat step", "the beginning of each end of combat step", PhaseStepQuantifierEach, TriggerPlayerSelectorAny, PhaseStepNameEndOfCombatStep, TriggerSelection{}},
		{"each of your first main phases", "the beginning of each of your first main phases", PhaseStepQuantifierEachOf, TriggerPlayerSelectorYou, PhaseStepNameFirstMainPhase, TriggerSelection{}},
		{"your second main phase", "the beginning of your second main phase", PhaseStepQuantifierSingle, TriggerPlayerSelectorYou, PhaseStepNameSecondMainPhase, TriggerSelection{}},
		{"your combat step", "the beginning of your combat step", PhaseStepQuantifierSingle, TriggerPlayerSelectorYou, PhaseStepNameCombatStep, TriggerSelection{}},
		{"attached controller", "the beginning of the upkeep of enchanted legendary white artifact creature's controller", PhaseStepQuantifierSingle, TriggerPlayerSelectorAttachedController, PhaseStepNameUpkeep, TriggerSelection{
			RequiredTypes: []TriggerCardType{TriggerCardTypeArtifact, TriggerCardTypeCreature},
			Supertypes:    []TriggerSupertype{TriggerSupertypeLegendary},
			ColorsAny:     []TriggerColor{TriggerColorWhite},
		}},
		{"attached union controller", "the beginning of the upkeep of enchanted artifact and/or creature's controller", PhaseStepQuantifierSingle, TriggerPlayerSelectorAttachedController, PhaseStepNameUpkeep, TriggerSelection{
			RequiredTypesAny: []TriggerCardType{TriggerCardTypeArtifact, TriggerCardTypeCreature},
		}},
		{"attached constrained controller", "the beginning of the upkeep of enchanted permanent you control's controller", PhaseStepQuantifierSingle, TriggerPlayerSelectorAttachedController, PhaseStepNameUpkeep, TriggerSelection{
			Controller: ControllerYou,
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			source := "At " + test.event + ", draw a card."
			document, diagnostics := Parse(source, Context{})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			trigger := document.Abilities[0].Trigger
			if trigger == nil || trigger.PhaseStep == nil {
				t.Fatalf("trigger = %#v, want typed phase/step clause", trigger)
			}
			if trigger.Introduction.Kind != TriggerIntroductionAt ||
				trigger.PhaseStep.Quantifier.Kind != test.quantifier ||
				trigger.PhaseStep.Player.Kind != test.player ||
				trigger.PhaseStep.Name.Kind != test.phaseStep ||
				!reflect.DeepEqual(trigger.PhaseStep.Player.AttachedSubject.Selection, test.attached) {
				t.Fatalf("trigger = %#v", trigger)
			}
			assertTextSpan(t, "trigger clause", source, trigger.Span, trigger.Text)
			assertTextSpan(t, "trigger event", source, trigger.Event.Span, trigger.Event.Text)
			assertSpanContains(t, "phase/step clause", trigger.Event.Span, trigger.PhaseStep.Span)
			assertSpanContains(t, "phase/step name", trigger.PhaseStep.Span, trigger.PhaseStep.Name.Span)
		})
	}
}

func TestParsePhaseStepTriggerClausesComposePreviouslyUnsupportedSlots(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"At the beginning of combat on each player's turn, draw a card.",
		"At the beginning of each precombat main phase, draw a card.",
		"At the beginning of your end of combat step, draw a card.",
		"At the beginning of each of your upkeeps, draw a card.",
	} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(source, Context{})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			if trigger := document.Abilities[0].Trigger; trigger == nil || trigger.PhaseStep == nil {
				t.Fatalf("trigger = %#v, want composed phase/step grammar", trigger)
			}
		})
	}
}

func TestParseEveryPreviouslySupportedPhaseStepTriggerClause(t *testing.T) {
	t.Parallel()
	var events []string
	for _, relation := range []string{
		"your",
		"its controller's",
		"the",
		"each",
		"each player's",
		"each opponent's",
	} {
		for _, name := range []string{"upkeep", "draw step", "end step"} {
			events = append(events, "the beginning of "+relation+" "+name)
		}
	}
	events = append(events,
		"end of combat",
		"the end of combat",
		"the beginning of combat on your turn",
		"the beginning of combat on each turn",
		"the beginning of combat on each opponent's turn",
		"the beginning of each combat",
		"the beginning of the end of combat",
		"the beginning of the end of combat on your turn",
		"the beginning of each end of combat step",
		"the beginning of your precombat main phase",
		"the beginning of each player's precombat main phase",
		"the beginning of each opponent's precombat main phase",
		"the beginning of your postcombat main phase",
		"the beginning of each player's postcombat main phase",
		"the beginning of each opponent's postcombat main phase",
		"the beginning of your first main phase",
		"the beginning of each of your first main phases",
		"the beginning of each player's first main phase",
		"the beginning of each opponent's first main phase",
		"the beginning of your second main phase",
		"the beginning of each player's second main phase",
		"the beginning of each opponent's second main phase",
		"the beginning of each of your postcombat main phases",
		"the beginning of your combat step",
		"the beginning of the upkeep of enchanted creature's controller",
		"the beginning of the draw step of enchanted creature's controller",
		"the beginning of the end step of enchanted creature's controller",
		"the beginning of the upkeep of enchanted permanent's controller",
	)
	for _, event := range events {
		t.Run(event, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse("At "+event+", draw a card.", Context{})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			if trigger := document.Abilities[0].Trigger; trigger == nil || trigger.PhaseStep == nil {
				t.Fatalf("trigger = %#v, want typed phase/step clause", trigger)
			}
		})
	}
}

func TestParseSimpleStaticRulesAsComposedSyntax(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		source         string
		subject        StaticRuleSubjectKind
		subjectText    string
		constraint     StaticRuleConstraintKind
		constraintText string
		operation      StaticRuleOperationKind
		voice          StaticRuleVoice
		operationText  string
		qualifiers     []StaticRuleQualifierKind
	}{
		"cannot block contraction": {
			source:         "This creature can't block.",
			subject:        StaticRuleSubjectSourceCreature,
			subjectText:    "This creature",
			constraint:     StaticRuleConstraintProhibition,
			constraintText: "can't",
			operation:      StaticRuleOperationBlock,
			voice:          StaticRuleVoiceActive,
			operationText:  "block",
		},
		"cannot block": {
			source:         "This creature cannot block.",
			subject:        StaticRuleSubjectSourceCreature,
			subjectText:    "This creature",
			constraint:     StaticRuleConstraintProhibition,
			constraintText: "cannot",
			operation:      StaticRuleOperationBlock,
			voice:          StaticRuleVoiceActive,
			operationText:  "block",
		},
		"cannot be blocked contraction": {
			source:         "This creature can't be blocked.",
			subject:        StaticRuleSubjectSourceCreature,
			subjectText:    "This creature",
			constraint:     StaticRuleConstraintProhibition,
			constraintText: "can't",
			operation:      StaticRuleOperationBlock,
			voice:          StaticRuleVoicePassive,
			operationText:  "be blocked",
		},
		"cannot be blocked": {
			source:         "This creature cannot be blocked.",
			subject:        StaticRuleSubjectSourceCreature,
			subjectText:    "This creature",
			constraint:     StaticRuleConstraintProhibition,
			constraintText: "cannot",
			operation:      StaticRuleOperationBlock,
			voice:          StaticRuleVoicePassive,
			operationText:  "be blocked",
		},
		"implicit attack requirement": {
			source:         "This creature attacks each combat if able.",
			subject:        StaticRuleSubjectSourceCreature,
			subjectText:    "This creature",
			constraint:     StaticRuleConstraintRequirement,
			constraintText: "attacks each combat if able",
			operation:      StaticRuleOperationAttack,
			voice:          StaticRuleVoiceActive,
			operationText:  "attacks",
			qualifiers:     []StaticRuleQualifierKind{StaticRuleQualifierEachCombat, StaticRuleQualifierIfAble},
		},
		"explicit attack requirement": {
			source:         "This creature must attack each combat if able.",
			subject:        StaticRuleSubjectSourceCreature,
			subjectText:    "This creature",
			constraint:     StaticRuleConstraintRequirement,
			constraintText: "must",
			operation:      StaticRuleOperationAttack,
			voice:          StaticRuleVoiceActive,
			operationText:  "attack",
			qualifiers:     []StaticRuleQualifierKind{StaticRuleQualifierEachCombat, StaticRuleQualifierIfAble},
		},
		"cannot be countered contraction": {
			source:         "This spell can't be countered.",
			subject:        StaticRuleSubjectSourceSpell,
			subjectText:    "This spell",
			constraint:     StaticRuleConstraintProhibition,
			constraintText: "can't",
			operation:      StaticRuleOperationCounter,
			voice:          StaticRuleVoicePassive,
			operationText:  "be countered",
		},
		"cannot be countered": {
			source:         "This spell cannot be countered.",
			subject:        StaticRuleSubjectSourceSpell,
			subjectText:    "This spell",
			constraint:     StaticRuleConstraintProhibition,
			constraintText: "cannot",
			operation:      StaticRuleOperationCounter,
			voice:          StaticRuleVoicePassive,
			operationText:  "be countered",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(test.source, Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			ability := document.Abilities[0]
			if ability.Kind != AbilityStatic || len(ability.Sentences) != 1 || ability.Sentences[0].StaticRule == nil {
				t.Fatalf("ability = %#v, want one typed static rule", ability)
			}
			rule := ability.Sentences[0].StaticRule
			if rule.Subject.Kind != test.subject ||
				rule.Constraint.Kind != test.constraint ||
				rule.Operation.Kind != test.operation ||
				rule.Operation.Voice != test.voice {
				t.Fatalf("static rule = %#v", rule)
			}
			assertTextSpan(t, "rule", test.source, rule.Span, test.source)
			assertTextSpan(t, "subject", test.source, rule.Subject.Span, test.subjectText)
			assertTextSpan(t, "constraint", test.source, rule.Constraint.Span, test.constraintText)
			assertTextSpan(t, "operation", test.source, rule.Operation.Span, test.operationText)
			if len(rule.Qualifiers) != len(test.qualifiers) {
				t.Fatalf("qualifiers = %#v, want %v", rule.Qualifiers, test.qualifiers)
			}
			for i, qualifier := range rule.Qualifiers {
				if qualifier.Kind != test.qualifiers[i] {
					t.Fatalf("qualifier %d = %#v, want %v", i, qualifier, test.qualifiers[i])
				}
			}
			if len(rule.Qualifiers) == 2 {
				assertTextSpan(t, "each-combat qualifier", test.source, rule.Qualifiers[0].Span, "each combat")
				assertTextSpan(t, "if-able qualifier", test.source, rule.Qualifiers[1].Span, "if able")
			}
		})
	}
}

func TestParsePhaseStepTriggerClausesFailClosed(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"At the beginning of a player's upkeep, draw a card.",
		"At the beginning of your next upkeep, draw a card.",
		"At the beginning of your declare attackers step, draw a card.",
		"At the beginning of each of your upkeep, draw a card.",
		"At the beginning of each of your main phases, draw a card.",
		"At the beginning of combat on the turn, draw a card.",
		"At the beginning of end of combat on your turn, draw a card.",
		"Whenever the beginning of your upkeep, draw a card.",
	} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(source, Context{})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			trigger := document.Abilities[0].Trigger
			if trigger == nil || trigger.Event.Text == "" || trigger.Event.Span == (shared.Span{}) {
				t.Fatalf("trigger = %#v, want source-spanned unrecognized clause", trigger)
			}
			if trigger.PhaseStep != nil {
				t.Fatalf("trigger = %#v, want unrecognized phase/step grammar", trigger)
			}
		})
	}
}

func TestParseSimpleStaticRulesFailClosedOnNearMisses(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"This creature can block.",
		"This creature can't attack.",
		"This creature can't be countered.",
		"This creature attacks each combat.",
		"This creature attack each combat if able.",
		"This creature attacks each turn if able.",
		"This creature must attack if able.",
		"This creature must attacks each combat if able.",
		"This spell can't block.",
		"This spell can't be blocked.",
		"This spell can't be countered by spells.",
	} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			for _, sentence := range document.Abilities[0].Sentences {
				if sentence.StaticRule != nil {
					t.Fatalf("%q parsed as %#v", source, sentence.StaticRule)
				}
			}
		})
	}
}

func TestParsePlayerEventTriggerClauses(t *testing.T) {
	t.Parallel()
	tests := []struct {
		event      string
		player     TriggerPlayerSelectorKind
		action     PlayerEventActionKind
		card       PlayerEventCardKind
		occurrence PlayerEventOccurrenceKind
		ordinal    int
	}{
		{"Whenever you draw a card", TriggerPlayerSelectorYou, PlayerEventActionDraw, PlayerEventCardSingle, PlayerEventOccurrenceAny, 0},
		{"Whenever an opponent discards a card", TriggerPlayerSelectorOpponent, PlayerEventActionDiscard, PlayerEventCardSingle, PlayerEventOccurrenceAny, 0},
		{"Whenever a player cycles a card", TriggerPlayerSelectorAny, PlayerEventActionCycle, PlayerEventCardSingle, PlayerEventOccurrenceAny, 0},
		{"Whenever you discard one or more cards", TriggerPlayerSelectorYou, PlayerEventActionDiscard, PlayerEventCardOneOrMore, PlayerEventOccurrenceAny, 0},
		{"Whenever you cycle another card", TriggerPlayerSelectorYou, PlayerEventActionCycle, PlayerEventCardAnother, PlayerEventOccurrenceAny, 0},
		{"Whenever you cycle or discard another card", TriggerPlayerSelectorYou, PlayerEventActionCycleOrDiscard, PlayerEventCardAnother, PlayerEventOccurrenceAny, 0},
		{"Whenever you scry", TriggerPlayerSelectorYou, PlayerEventActionScry, PlayerEventCardNone, PlayerEventOccurrenceAny, 0},
		{"Whenever a player surveils", TriggerPlayerSelectorAny, PlayerEventActionSurveil, PlayerEventCardNone, PlayerEventOccurrenceAny, 0},
		{"Whenever an opponent gains life", TriggerPlayerSelectorOpponent, PlayerEventActionGainLife, PlayerEventCardNone, PlayerEventOccurrenceAny, 0},
		{"Whenever you lose life", TriggerPlayerSelectorYou, PlayerEventActionLoseLife, PlayerEventCardNone, PlayerEventOccurrenceAny, 0},
		{"Whenever a player draws their fourth card each turn", TriggerPlayerSelectorAny, PlayerEventActionDraw, PlayerEventCardSingle, PlayerEventOccurrenceOrdinalEachTurn, 4},
		{"When you surveil for the first time each turn", TriggerPlayerSelectorYou, PlayerEventActionSurveil, PlayerEventCardNone, PlayerEventOccurrenceFirstEachTurn, 1},
	}
	for _, test := range tests {
		t.Run(test.event, func(t *testing.T) {
			t.Parallel()
			source := test.event + ", draw a card."
			document, diagnostics := Parse(source, Context{})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			trigger := document.Abilities[0].Trigger
			if trigger == nil || trigger.PlayerEvent == nil {
				t.Fatalf("trigger = %#v, want typed player-event clause", trigger)
			}
			if trigger.PlayerEvent.Player.Kind != test.player ||
				trigger.PlayerEvent.Action.Kind != test.action ||
				trigger.PlayerEvent.Card.Kind != test.card ||
				trigger.PlayerEvent.Occurrence.Kind != test.occurrence ||
				trigger.PlayerEvent.Occurrence.Ordinal != test.ordinal {
				t.Fatalf("trigger = %#v", trigger)
			}
			assertTextSpan(t, "trigger clause", source, trigger.Span, trigger.Text)
			assertTextSpan(t, "trigger event", source, trigger.Event.Span, trigger.Event.Text)
			assertSpanContains(t, "player-event clause", trigger.Event.Span, trigger.PlayerEvent.Span)
			assertSpanContains(t, "player selector", trigger.PlayerEvent.Span, trigger.PlayerEvent.Player.Span)
			assertSpanContains(t, "player action", trigger.PlayerEvent.Span, trigger.PlayerEvent.Action.Span)
		})
	}
}

func TestParseEveryPreviouslySupportedSimplePlayerEventTriggerClause(t *testing.T) {
	t.Parallel()
	for _, event := range []string{
		"you draw a card",
		"an opponent draws a card",
		"a player draws a card",
		"you discard a card",
		"an opponent discards a card",
		"a player discards a card",
		"you discard one or more cards",
		"you cycle a card",
		"an opponent cycles a card",
		"a player cycles a card",
		"you cycle another card",
		"you scry",
		"an opponent scries",
		"a player scries",
		"you surveil",
		"an opponent surveils",
		"a player surveils",
		"you cycle or discard a card",
		"you cycle or discard another card",
		"you gain life",
		"an opponent gains life",
		"you lose life",
		"an opponent loses life",
	} {
		t.Run(event, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse("Whenever "+event+", draw a card.", Context{})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			if trigger := document.Abilities[0].Trigger; trigger == nil || trigger.PlayerEvent == nil {
				t.Fatalf("trigger = %#v, want typed player-event clause", trigger)
			}
		})
	}
}

func TestParseEveryPreviouslySupportedPlayerEventOccurrenceTriggerClause(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"Whenever you draw your second card each turn, draw a card.",
		"Whenever an opponent draws their first card each turn, draw a card.",
		"Whenever a player draws their fifth card each turn, draw a card.",
		"Whenever you draw a card for the first time each turn, draw a card.",
		"Whenever an opponent draws a card for the first time each turn, draw a card.",
		"Whenever a player draws a card for the first time each turn, draw a card.",
		"Whenever you scry for the first time each turn, draw a card.",
		"Whenever an opponent scries for the first time each turn, draw a card.",
		"Whenever a player scries for the first time each turn, draw a card.",
		"Whenever you surveil for the first time each turn, draw a card.",
		"Whenever an opponent surveils for the first time each turn, draw a card.",
		"Whenever a player surveils for the first time each turn, draw a card.",
		"Whenever you gain life for the first time each turn, draw a card.",
		"Whenever an opponent gains life for the first time each turn, draw a card.",
		"When you lose life for the first time each turn, draw a card.",
		"When an opponent loses life for the first time each turn, draw a card.",
	} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(source, Context{})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			trigger := document.Abilities[0].Trigger
			if trigger == nil ||
				trigger.PlayerEvent == nil ||
				trigger.PlayerEvent.Occurrence.Kind == PlayerEventOccurrenceAny {
				t.Fatalf("trigger = %#v, want typed player-event occurrence", trigger)
			}
		})
	}
}

func TestParsePlayerEventTriggerClausesComposePreviouslyUnsupportedSlots(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"Whenever an opponent discards one or more cards, draw a card.",
		"Whenever a player discards another card, draw a card.",
		"Whenever an opponent cycles or discards another card, draw a card.",
		"Whenever a player gains life, draw a card.",
		"Whenever a player loses life, draw a card.",
	} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(source, Context{})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			if trigger := document.Abilities[0].Trigger; trigger == nil || trigger.PlayerEvent == nil {
				t.Fatalf("trigger = %#v, want composed player-event grammar", trigger)
			}
		})
	}
}

func TestParsePlayerEventTriggerClausesFailClosed(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"Whenever your draw a card, draw a card.",
		"Whenever you draws a card, draw a card.",
		"Whenever a player draw a card, draw a card.",
		"When you draw a card, draw a card.",
		"Whenever you scry a card, draw a card.",
		"Whenever you draw one or more cards, draw a card.",
		"Whenever you cycle one or more cards, draw a card.",
		"Whenever you cycle or discard one or more cards, draw a card.",
		"Whenever you gain a life, draw a card.",
		"Whenever you discard another cards, draw a card.",
		"Whenever you discard a card for the second time each turn, draw a card.",
		"Whenever a player draws your second card each turn, draw a card.",
		"Whenever a player gains life for the first time each turn, draw a card.",
		"When a player loses life for the first time each turn, draw a card.",
		"At you draw a card, draw a card.",
	} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(source, Context{})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			trigger := document.Abilities[0].Trigger
			if trigger == nil || trigger.Event.Text == "" || trigger.Event.Span == (shared.Span{}) {
				t.Fatalf("trigger = %#v, want source-spanned unrecognized clause", trigger)
			}
			if trigger.PlayerEvent != nil {
				t.Fatalf("trigger = %#v, want unrecognized player-event grammar", trigger)
			}
		})
	}
}

func TestParseSagaChapterHeading(t *testing.T) {
	t.Parallel()
	source := "I, II, III — Draw a card."
	document, diagnostics := Parse(source, Context{Saga: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := document.Abilities[0]
	if ability.AbilityWord != nil {
		t.Fatalf("ability word = %#v, want nil", ability.AbilityWord)
	}
	if !slices.Equal(ability.Chapters, []int{1, 2, 3}) {
		t.Fatalf("chapters = %v, want [1 2 3]", ability.Chapters)
	}
	assertTextSpan(t, "chapter heading", source, ability.ChapterSpan, "I, II, III")
}

func TestParseDoesNotTreatRomanNumeralsAsChaptersOutsideSagaContext(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse("I — Draw a card.", Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := document.Abilities[0]
	if ability.Kind == AbilityChapter || ability.AbilityWord == nil || ability.AbilityWord.Text != "I" {
		t.Fatalf("ability = %#v, want ordinary ability-word syntax", ability)
	}
}

func TestParseStructures(t *testing.T) {
	t.Parallel()
	source := "Formidable — {1}{G}, {T}: Draw a card. Then discard a card. (Do this once.)"
	document, diagnostics := Parse(source, Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := document.Abilities[0]
	if ability.AbilityWord == nil || ability.AbilityWord.Text != "Formidable" {
		t.Fatalf("ability word = %#v", ability.AbilityWord)
	}
	if ability.Cost == nil || ability.Cost.Text != "{1}{G}, {T}" {
		t.Fatalf("cost = %#v", ability.Cost)
	}
	if len(ability.Sentences) != 3 {
		t.Fatalf("sentences = %#v", ability.Sentences)
	}
	if len(ability.Reminders) != 1 || ability.Reminders[0].Text != "(Do this once.)" {
		t.Fatalf("reminders = %#v", ability.Reminders)
	}
}

func TestParseModalAbility(t *testing.T) {
	t.Parallel()
	source := "Choose one —\n• Draw a card.\n• Target creature fights another target creature. (They deal damage.)"
	document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(document.Abilities) != 1 {
		t.Fatalf("abilities = %d", len(document.Abilities))
	}
	ability := document.Abilities[0]
	if ability.Kind != AbilitySpell || ability.Modal == nil {
		t.Fatalf("modal ability = %#v", ability)
	}
	if ability.Text != source {
		t.Fatalf("text = %q", ability.Text)
	}
	if len(ability.Modal.Options) != 2 {
		t.Fatalf("options = %#v", ability.Modal.Options)
	}
	if got := ability.Modal.Options[1].Text; got != "Target creature fights another target creature. (They deal damage.)" {
		t.Fatalf("second mode = %q", got)
	}
	if ability.Modal.Options[1].Span.Start != ability.Modal.Options[1].Tokens[0].Span.Start {
		t.Fatal("mode span includes syntax outside mode tokens")
	}
}

func TestParseModalActivatedAbility(t *testing.T) {
	t.Parallel()
	source := "{1}, Discard a card: Choose one —\n• Draw a card.\n• You gain 3 life."
	document, diagnostics := Parse(source, Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(document.Abilities) != 1 {
		t.Fatalf("abilities = %#v, want one", document.Abilities)
	}
	ability := document.Abilities[0]
	if ability.Kind != AbilityActivated || ability.Cost == nil || ability.Modal == nil {
		t.Fatalf("ability = %#v, want modal activated ability", ability)
	}
	if ability.Cost.Text != "{1}, Discard a card" || ability.Modal.Header.Text != "Choose one —" || len(ability.Modal.Options) != 2 {
		t.Fatalf("cost/header/options = %q/%q/%d", ability.Cost.Text, ability.Modal.Header.Text, len(ability.Modal.Options))
	}

	withWord, diagnostics := Parse("Hellbent — "+source, Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("ability-word diagnostics = %#v", diagnostics)
	}
	ability = withWord.Abilities[0]
	if ability.AbilityWord == nil || ability.AbilityWord.Text != "Hellbent" ||
		ability.Cost == nil || ability.Cost.Text != "{1}, Discard a card" ||
		ability.Modal == nil {
		t.Fatalf("ability-word modal activated ability = %#v", ability)
	}
}

func TestParseInlineModalAbility(t *testing.T) {
	t.Parallel()
	source := "Choose one — Noxious Hydra Breath deals 5 damage to each player; or destroy each tapped non-Head creature."
	document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(document.Abilities) != 1 {
		t.Fatalf("abilities = %#v", document.Abilities)
	}
	ability := document.Abilities[0]
	if ability.Modal == nil || len(ability.Modal.Options) != 2 {
		t.Fatalf("modal = %#v", ability.Modal)
	}
	if ability.Modal.Header.Text != "Choose one —" {
		t.Fatalf("header = %q", ability.Modal.Header.Text)
	}
	if got := ability.Modal.Options[0].Text; got != "Noxious Hydra Breath deals 5 damage to each player" {
		t.Fatalf("first mode = %q", got)
	}
	if got := ability.Modal.Options[1].Text; got != "destroy each tapped non-Head creature." {
		t.Fatalf("second mode = %q", got)
	}
}

func TestChooseSentenceBeforeVillainousChoiceIsNotModalHeader(t *testing.T) {
	t.Parallel()
	source := "Choose up to four target creatures you don't control. For each of them, that creature's controller faces a villainous choice — That creature becomes a 1/1 white Human creature and loses all abilities, or you create a token that's a copy of it."
	document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(document.Abilities) != 1 {
		t.Fatalf("abilities = %#v", document.Abilities)
	}
	ability := document.Abilities[0]
	if ability.Kind != AbilitySpell || ability.Modal != nil || ability.Text != source {
		t.Fatalf("ability = %#v", ability)
	}
}

func TestParseQuotedAbility(t *testing.T) {
	t.Parallel()
	source := `Equipped creature has "{2}: This creature gets +1/+0 until end of turn."`
	document, diagnostics := Parse(source, Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	quoted := document.Abilities[0].Quoted
	if len(quoted) != 1 || quoted[0].Text != `"{2}: This creature gets +1/+0 until end of turn."` {
		t.Fatalf("quoted = %#v", quoted)
	}
	if len(document.Abilities[0].Sentences) != 1 {
		t.Fatal("sentence split inside quote")
	}
}

func TestParseNestedDelimitedText(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		source        string
		wantReminders int
		wantQuoted    int
	}{
		"reminder inside quote": {
			source:        `Enchanted creature has "Flying (it can't be blocked)."`,
			wantReminders: 0,
			wantQuoted:    1,
		},
		"quote inside reminder": {
			source:        `Flying (This means "can't be blocked.")`,
			wantReminders: 1,
			wantQuoted:    0,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(test.source, Context{})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			ability := document.Abilities[0]
			if len(ability.Reminders) != test.wantReminders || len(ability.Quoted) != test.wantQuoted {
				t.Fatalf("reminders = %#v, quoted = %#v", ability.Reminders, ability.Quoted)
			}
		})
	}
}

func TestParseMultilineReminderOnlyAbility(t *testing.T) {
	t.Parallel()
	source := "(You can cover a face-down creature with this reminder card.\nA card with morph can be turned face up any time for its morph cost.)"
	document, diagnostics := Parse(source, Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(document.Abilities) != 1 {
		t.Fatalf("abilities = %#v", document.Abilities)
	}
	ability := document.Abilities[0]
	if ability.Kind != AbilityReminder || ability.Text != source {
		t.Fatalf("ability = %#v", ability)
	}
	if len(ability.Reminders) != 1 || ability.Reminders[0].Text != source {
		t.Fatalf("reminders = %#v", ability.Reminders)
	}
}

func TestParseUnclosedMultilineReminderRecoversAtNewline(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse("(unclosed\nFlying", Context{})
	if len(document.Abilities) != 2 {
		t.Fatalf("abilities = %#v", document.Abilities)
	}
	if len(diagnostics) != 1 || diagnostics[0].Summary != "unclosed parenthesis" {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
}

func TestParseEmbeddedParenthesisDoesNotJoinAbilities(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse("Flying (gains flying\nTrample)", Context{})
	if len(document.Abilities) != 2 {
		t.Fatalf("abilities = %#v", document.Abilities)
	}
	if len(diagnostics) != 2 ||
		diagnostics[0].Summary != "unclosed parenthesis" ||
		diagnostics[1].Summary != "unmatched parenthesis" {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
}

func TestParseDiagnosticsAndRecovery(t *testing.T) {
	t.Parallel()
	source := "Flying)\nChoose one —\nHaste\n\"unclosed"
	document, diagnostics := Parse(source, Context{})
	if len(document.Abilities) != 4 {
		t.Fatalf("abilities = %d", len(document.Abilities))
	}
	if len(diagnostics) != 3 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if diagnostics[0].Summary != "unmatched parenthesis" ||
		diagnostics[1].Summary != "modal ability has no options" ||
		diagnostics[2].Summary != "unclosed quote" {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
}

func TestParseScryfallCacheLosslessly(t *testing.T) {
	t.Parallel()
	cache := filepath.Join("..", "..", ".cardwork", "deck", "cache", "scryfall")
	paths, err := filepath.Glob(filepath.Join(cache, "*.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) == 0 {
		t.Skip("local Scryfall cache is not present")
	}

	var texts int
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var card cachedParserCard
		if err := json.Unmarshal(data, &card); err != nil {
			t.Fatalf("%s: %v", path, err)
		}
		check := func(name, typeLine, source string) {
			t.Helper()
			if source == "" {
				return
			}
			texts++
			context := Context{
				InstantOrSorcery: typeLine == "Instant" || typeLine == "Sorcery",
				Planeswalker:     typeLine == "Planeswalker" || typeLine == "Legendary Planeswalker",
			}
			document, diagnostics := Parse(source, context)
			if len(diagnostics) != 0 {
				t.Fatalf("%s: diagnostics = %#v", name, diagnostics)
			}
			if document.Source != source ||
				document.Span.Start.Offset != 0 ||
				document.Span.End.Offset != len(source) {
				t.Fatalf("%s: document is not lossless", name)
			}
			for _, ability := range document.Abilities {
				assertAbilitySpans(t, name, source, ability)
			}
		}
		check(card.Name, card.TypeLine, card.OracleText)
		for _, face := range card.CardFaces {
			check(face.Name, face.TypeLine, face.OracleText)
		}
	}
	if texts != 59 {
		t.Fatalf("checked %d non-empty Oracle texts, want 59", texts)
	}
}

func assertAbilitySpans(t *testing.T, name, source string, ability Ability) {
	t.Helper()
	assertTextSpan(t, name+" ability", source, ability.Span, ability.Text)
	assertTokensInSpan(t, name+" ability", ability.Span, ability.Tokens)
	for _, sentence := range ability.Sentences {
		assertTextSpan(t, name+" sentence", source, sentence.Span, sentence.Text)
		assertSpanContains(t, name+" sentence", ability.Span, sentence.Span)
	}
	for _, reminder := range ability.Reminders {
		assertTextSpan(t, name+" reminder", source, reminder.Span, reminder.Text)
		assertSpanContains(t, name+" reminder", ability.Span, reminder.Span)
	}
	for _, quoted := range ability.Quoted {
		assertTextSpan(t, name+" quote", source, quoted.Span, quoted.Text)
		assertSpanContains(t, name+" quote", ability.Span, quoted.Span)
	}
	assertDisjoint(t, name, ability.Reminders, ability.Quoted)
	if ability.AbilityWord != nil {
		assertTextSpan(t, name+" ability word", source, ability.AbilityWord.Span, ability.AbilityWord.Text)
	}
	if ability.Cost != nil {
		assertTextSpan(t, name+" cost", source, ability.Cost.Span, ability.Cost.Text)
	}
	if ability.Trigger != nil {
		assertTextSpan(t, name+" trigger", source, ability.Trigger.Span, ability.Trigger.Text)
		assertTokensInSpan(t, name+" trigger", ability.Trigger.Span, ability.Trigger.Tokens)
		assertSpanContains(t, name+" trigger introduction", ability.Trigger.Span, ability.Trigger.Introduction.Span)
		if ability.Trigger.Event.Text != "" {
			assertTextSpan(t, name+" trigger event", source, ability.Trigger.Event.Span, ability.Trigger.Event.Text)
			assertSpanContains(t, name+" trigger event", ability.Trigger.Span, ability.Trigger.Event.Span)
		}
		if phaseStep := ability.Trigger.PhaseStep; phaseStep != nil {
			assertSpanContains(t, name+" phase/step", ability.Trigger.Event.Span, phaseStep.Span)
			assertSpanContains(t, name+" phase/step name", phaseStep.Span, phaseStep.Name.Span)
			if phaseStep.Quantifier.Span != (shared.Span{}) {
				assertSpanContains(t, name+" phase/step quantifier", phaseStep.Span, phaseStep.Quantifier.Span)
			}
			if phaseStep.Player.Span != (shared.Span{}) {
				assertSpanContains(t, name+" phase/step player", phaseStep.Span, phaseStep.Player.Span)
			}
			if phaseStep.Player.AttachedSubject.Span != (shared.Span{}) {
				assertSpanContains(t, name+" phase/step attached subject", phaseStep.Player.Span, phaseStep.Player.AttachedSubject.Span)
			}
		}
	}
	if ability.Modal == nil {
		return
	}
	assertTextSpan(t, name+" modal header", source, ability.Modal.Header.Span, ability.Modal.Header.Text)
	for _, mode := range ability.Modal.Options {
		assertTextSpan(t, name+" mode", source, mode.Span, mode.Text)
		assertTokensInSpan(t, name+" mode", mode.Span, mode.Tokens)
		assertSpanContains(t, name+" mode", ability.Span, mode.Span)
		for _, sentence := range mode.Sentences {
			assertTextSpan(t, name+" mode sentence", source, sentence.Span, sentence.Text)
			assertSpanContains(t, name+" mode sentence", mode.Span, sentence.Span)
		}
	}
}

func assertTextSpan(t *testing.T, name, source string, span shared.Span, text string) {
	t.Helper()
	if span.Start.Offset < 0 || span.End.Offset < span.Start.Offset || span.End.Offset > len(source) {
		t.Fatalf("%s has invalid span %#v", name, span)
	}
	if got := source[span.Start.Offset:span.End.Offset]; got != text {
		t.Fatalf("%s text = %q, source span = %q", name, text, got)
	}
}

func assertTokensInSpan(t *testing.T, name string, parent shared.Span, tokens []shared.Token) {
	t.Helper()
	for _, token := range tokens {
		assertSpanContains(t, name+" token", parent, token.Span)
	}
}

func assertSpanContains(t *testing.T, name string, parent, child shared.Span) {
	t.Helper()
	if child.Start.Offset < parent.Start.Offset || child.End.Offset > parent.End.Offset {
		t.Fatalf("%s span %#v is outside parent %#v", name, child, parent)
	}
}

func assertDisjoint(t *testing.T, name string, reminders, quoted []Delimited) {
	t.Helper()
	for _, reminder := range reminders {
		for _, quote := range quoted {
			if reminder.Span.Start.Offset < quote.Span.End.Offset &&
				quote.Span.Start.Offset < reminder.Span.End.Offset {
				t.Fatalf("%s reminder %#v overlaps quote %#v", name, reminder.Span, quote.Span)
			}
		}
	}
}
