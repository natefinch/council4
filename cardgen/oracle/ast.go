package oracle

// AbilityKind is the syntactic category of an Oracle-text ability.
type AbilityKind uint8

// Ability kinds recognized by the syntax parser.
const (
	AbilityUnknown AbilityKind = iota
	AbilitySpell
	AbilityActivated
	AbilityLoyalty
	AbilityChapter
	AbilityTriggered
	AbilityReplacement
	AbilityStatic
	AbilityReminder
)

var abilityKindNames = [...]string{
	AbilityUnknown:     "unknown",
	AbilitySpell:       "spell",
	AbilityActivated:   "activated",
	AbilityLoyalty:     "loyalty",
	AbilityChapter:     "chapter",
	AbilityTriggered:   "triggered",
	AbilityReplacement: "replacement",
	AbilityStatic:      "static",
	AbilityReminder:    "reminder",
}

func (k AbilityKind) String() string {
	if int(k) >= len(abilityKindNames) {
		return "unknown"
	}
	return abilityKindNames[k]
}

// ParseContext supplies card-face facts that Oracle text alone cannot express.
type ParseContext struct {
	CardName         string
	InstantOrSorcery bool
	Planeswalker     bool
	Saga             bool
}

// Document is a lossless syntax tree for one card face's Oracle text.
type Document struct {
	Source    string
	Span      Span
	Abilities []Ability
}

// Ability is one Oracle-text paragraph, or one modal header and its options.
type Ability struct {
	Kind        AbilityKind
	Span        Span
	Text        string
	Tokens      []Token
	AbilityWord *Phrase
	Chapters    []int
	ChapterSpan Span
	Cost        *Phrase
	Trigger     *TriggerClause
	Sentences   []Sentence
	Reminders   []Delimited
	Quoted      []Delimited
	Modal       *Modal
}

// Phrase is a meaningful contiguous token range.
type Phrase struct {
	Span   Span
	Text   string
	Tokens []Token
}

// TriggerIntroductionKind identifies a trigger clause's leading word.
type TriggerIntroductionKind uint8

// Trigger introductions recognized by the syntax parser.
const (
	TriggerIntroductionUnknown TriggerIntroductionKind = iota
	TriggerIntroductionWhen
	TriggerIntroductionWhenever
	TriggerIntroductionAt
)

// TriggerIntroduction is the source-spanned leading word of a trigger clause.
type TriggerIntroduction struct {
	Kind TriggerIntroductionKind
	Span Span
}

// TriggerClause is the source-spanned syntax before a triggered ability's first
// top-level body comma. Event preserves unrecognized syntax as source metadata;
// PhaseStep carries typed grammar when the event is a recognized phase or step.
type TriggerClause struct {
	Span         Span
	Text         string
	Tokens       []Token
	Introduction TriggerIntroduction
	Event        Phrase
	PhaseStep    *PhaseStepTriggerClause
}

// PhaseStepQuantifierKind identifies a phase or step clause's grammatical
// cardinality.
type PhaseStepQuantifierKind uint8

// Phase and step quantifiers recognized by the syntax parser.
const (
	PhaseStepQuantifierUnknown PhaseStepQuantifierKind = iota
	PhaseStepQuantifierNone
	PhaseStepQuantifierSingle
	PhaseStepQuantifierEach
	PhaseStepQuantifierEachOf
)

// PhaseStepQuantifier is a source-spanned phase or step quantifier.
type PhaseStepQuantifier struct {
	Kind PhaseStepQuantifierKind
	Span Span
}

// PhaseStepPlayerRelationKind identifies whose phase or step a trigger uses.
type PhaseStepPlayerRelationKind uint8

// Phase and step player/controller relations recognized by the syntax parser.
const (
	PhaseStepPlayerRelationUnknown PhaseStepPlayerRelationKind = iota
	PhaseStepPlayerRelationAny
	PhaseStepPlayerRelationYou
	PhaseStepPlayerRelationOpponent
	PhaseStepPlayerRelationSourceController
	PhaseStepPlayerRelationAttachedController
)

// PhaseStepAttachedSubject is the typed subject in an attached-controller
// relation.
type PhaseStepAttachedSubject struct {
	Span      Span
	Selection TriggerSelection
}

// PhaseStepPlayerRelation is a source-spanned player/controller relation.
type PhaseStepPlayerRelation struct {
	Kind            PhaseStepPlayerRelationKind
	Span            Span
	AttachedSubject PhaseStepAttachedSubject
}

// PhaseStepNameKind identifies a literal phase or step name.
type PhaseStepNameKind uint8

// Literal phase and step names recognized by the syntax parser.
const (
	PhaseStepNameUnknown PhaseStepNameKind = iota
	PhaseStepNameUpkeep
	PhaseStepNameDrawStep
	PhaseStepNameEndStep
	PhaseStepNameCombat
	PhaseStepNameCombatStep
	PhaseStepNameEndOfCombat
	PhaseStepNameEndOfCombatStep
	PhaseStepNamePrecombatMainPhase
	PhaseStepNamePostcombatMainPhase
	PhaseStepNameFirstMainPhase
	PhaseStepNameSecondMainPhase
)

// PhaseStepName is a source-spanned literal phase or step name.
type PhaseStepName struct {
	Kind PhaseStepNameKind
	Span Span
}

// PhaseStepTriggerClause is composable typed syntax for a phase or step event.
type PhaseStepTriggerClause struct {
	Span       Span
	Quantifier PhaseStepQuantifier
	Player     PhaseStepPlayerRelation
	Name       PhaseStepName
}

// Sentence is a top-level sentence in an ability.
type Sentence struct {
	Span       Span
	Text       string
	Tokens     []Token
	StaticRule *StaticRuleSyntax
}

// StaticRuleSubjectKind identifies the source object constrained by a simple
// static rule.
type StaticRuleSubjectKind uint8

// Simple static-rule subjects.
const (
	StaticRuleSubjectUnknown StaticRuleSubjectKind = iota
	StaticRuleSubjectSourceCreature
	StaticRuleSubjectSourceSpell
)

// StaticRuleConstraintKind identifies whether a rule prohibits or requires an
// operation.
type StaticRuleConstraintKind uint8

// Simple static-rule constraints.
const (
	StaticRuleConstraintUnknown StaticRuleConstraintKind = iota
	StaticRuleConstraintProhibition
	StaticRuleConstraintRequirement
)

// StaticRuleOperationKind identifies the rules operation being constrained.
type StaticRuleOperationKind uint8

// Simple static-rule operations.
const (
	StaticRuleOperationUnknown StaticRuleOperationKind = iota
	StaticRuleOperationAttack
	StaticRuleOperationBlock
	StaticRuleOperationCounter
)

// StaticRuleVoice identifies the grammatical role the subject has in an
// operation.
type StaticRuleVoice uint8

// Simple static-rule voices.
const (
	StaticRuleVoiceUnknown StaticRuleVoice = iota
	StaticRuleVoiceActive
	StaticRuleVoicePassive
)

// StaticRuleQualifierKind identifies a composable restriction on an operation.
type StaticRuleQualifierKind uint8

// Simple static-rule qualifiers.
const (
	StaticRuleQualifierUnknown StaticRuleQualifierKind = iota
	StaticRuleQualifierEachCombat
	StaticRuleQualifierIfAble
)

// StaticRuleSubject is a source-spanned simple static-rule subject.
type StaticRuleSubject struct {
	Kind StaticRuleSubjectKind
	Span Span
}

// StaticRuleConstraint is a source-spanned requirement or prohibition.
type StaticRuleConstraint struct {
	Kind StaticRuleConstraintKind
	Span Span
}

// StaticRuleOperation is a source-spanned operation and the subject's
// grammatical role in it.
type StaticRuleOperation struct {
	Kind  StaticRuleOperationKind
	Voice StaticRuleVoice
	Span  Span
}

// StaticRuleQualifier is a source-spanned restriction on a rule operation.
type StaticRuleQualifier struct {
	Kind StaticRuleQualifierKind
	Span Span
}

// StaticRuleSyntax is a composable typed simple static-rule declaration.
// Sentence text and tokens remain available only as source metadata.
type StaticRuleSyntax struct {
	Span       Span
	Subject    StaticRuleSubject
	Constraint StaticRuleConstraint
	Operation  StaticRuleOperation
	Qualifiers []StaticRuleQualifier
}

// Delimited is parenthesized reminder text or a quoted granted ability.
type Delimited struct {
	Span   Span
	Text   string
	Tokens []Token
}

// Modal is a choose header followed by bullet or inline options.
type Modal struct {
	Header  Phrase
	Options []Mode
}

// Mode is one bullet option in a modal ability.
type Mode struct {
	Span      Span
	Text      string
	Tokens    []Token
	Sentences []Sentence
	Reminders []Delimited
	Quoted    []Delimited
}

// Severity is a parser diagnostic severity.
type Severity uint8

// Diagnostic severities.
const (
	SeverityError Severity = iota + 1
	SeverityWarning
)

// Diagnostic describes a localized lexical or syntax problem.
type Diagnostic struct {
	Severity Severity
	Summary  string
	Detail   string
	Span     Span
}
