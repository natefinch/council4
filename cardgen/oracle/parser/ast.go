// Package parser recognizes the grammatical structure of Oracle text.
package parser

import "github.com/natefinch/council4/cardgen/oracle/shared"

// AbilityKind is the syntactic category of an Oracle-text ability.
type AbilityKind string

// Ability kinds recognized by the syntax parser.
const (
	AbilityUnknown     AbilityKind = ""
	AbilitySpell       AbilityKind = "AbilitySpell"
	AbilityActivated   AbilityKind = "AbilityActivated"
	AbilityLoyalty     AbilityKind = "AbilityLoyalty"
	AbilityChapter     AbilityKind = "AbilityChapter"
	AbilityTriggered   AbilityKind = "AbilityTriggered"
	AbilityReplacement AbilityKind = "AbilityReplacement"
	AbilityStatic      AbilityKind = "AbilityStatic"
	AbilityReminder    AbilityKind = "AbilityReminder"
	// AbilitySpellAdditionalCost is a spell paragraph that declares an additional
	// cost to cast ("As an additional cost to cast this spell, <cost>."). Its
	// cost phrase is recognized through the shared cost machinery (CostSyntax);
	// it has no resolving body of its own.
	AbilitySpellAdditionalCost AbilityKind = "AbilitySpellAdditionalCost"
)

// Context supplies card-face facts that Oracle text alone cannot express.
type Context struct {
	InstantOrSorcery bool `json:",omitempty"`
	Planeswalker     bool `json:",omitempty"`
	Saga             bool `json:",omitempty"`
	// CardName is the card's own name. The parser uses it to recognize explicit
	// self-name references so the compiler need not inspect name spelling.
	CardName string `json:",omitempty"`
}

// Document is a lossless syntax tree for one card face's Oracle text.
type Document struct {
	Source    string      `json:",omitempty"`
	CardName  string      `json:",omitempty"`
	Span      shared.Span `json:"-"`
	Abilities []Ability   `json:",omitempty"`
}

// Ability is one Oracle-text paragraph, or one modal header and its options.
type Ability struct {
	Kind        AbilityKind        `json:",omitempty"`
	Span        shared.Span        `json:"-"`
	Text        string             `json:",omitempty"`
	Tokens      []shared.Token     `json:"-"`
	AbilityWord *AbilityWordClause `json:",omitempty"`
	Chapters    []int              `json:",omitempty"`
	ChapterSpan shared.Span        `json:"-"`
	// costPhrase is the source cost phrase recognized before the typed cost is
	// emitted. It is parser-internal: the compiler consumes the typed cost via
	// CostSyntax and only ever needs the cost's presence, not its tokens.
	costPhrase *Phrase
	Trigger    *TriggerClause `json:",omitempty"`
	// BodySpan is the source span of the ability's resolving body: the tokens
	// after the activated/loyalty cost colon or the triggered event comma (and
	// after any ability-word or chapter prefix). It is the zero span when the
	// body is empty.
	BodySpan shared.Span `json:"-"`
	// BodySeparatorSpan is the source span of the single separator token that
	// introduces the resolving body (the cost colon, the triggered event comma,
	// or a Saga chapter heading's em dash). It is the zero span when the body has
	// no separator (the whole ability is its body). Consumers that must account
	// for the separator punctuation read this typed span instead of locating the
	// separator token by kind.
	BodySeparatorSpan shared.Span `json:"-"`
	// CostSyntax is the typed cost recognized from the ability's cost phrase, or
	// nil when no cost was parsed.
	CostSyntax *Cost `json:",omitempty"`
	// Optional reports that a triggered ability's resolving body begins with the
	// optional "you may" choice; OptionalSpan covers those two words.
	Optional     bool        `json:",omitempty"`
	OptionalSpan shared.Span `json:"-"`
	// ConditionBoundaries lists every condition introducer the parser finds in
	// the ability's condition-scan token stream, in source order, including
	// introducers whose predicate is unrecognized (so the compiler can fail
	// closed).
	ConditionBoundaries []ConditionBoundary `json:",omitempty"`
	// EventHistoryConditions are the ability's parsed event-history conditions.
	EventHistoryConditions []EventHistoryCondition `json:",omitempty"`
	// ConditionClauses are the ability's typed condition clauses.
	ConditionClauses []ConditionClause `json:",omitempty"`
	// StaticDeclarations are the ability's typed static declarations.
	StaticDeclarations []StaticDeclarationSyntax `json:",omitempty"`
	// ConditionSegments are the ability's condition clauses, pre-segmented over
	// the same semantic token stream the compiler historically scanned.
	ConditionSegments []ConditionSegment `json:",omitempty"`
	// TriggerConditionSegments are the ability's condition clauses segmented over
	// its raw tokens, used to locate a triggered ability's intervening-if.
	TriggerConditionSegments []ConditionSegment `json:",omitempty"`
	// SemanticReferences are the explicit references recognized in the ability's
	// semantic tokens.
	SemanticReferences []Reference `json:",omitempty"`
	// SemanticKeywords are the keywords recognized in the ability's semantic body
	// tokens.
	SemanticKeywords []Keyword `json:",omitempty"`
	// ContentSpan is the source span of the ability's resolving content.
	ContentSpan            shared.Span             `json:"-"`
	ActivationRestrictions []ActivationRestriction `json:",omitempty"`
	// TriggerFrequency is the recognized trailing "This ability triggers only
	// once/twice each turn." qualifier on a triggered ability, or nil when
	// absent.
	TriggerFrequency *TriggerFrequencyRestriction `json:",omitempty"`
	Sentences        []Sentence                   `json:",omitempty"`
	Reminders        []Delimited                  `json:"-"`
	Quoted           []Delimited                  `json:"-"`
	Modal            *Modal                       `json:",omitempty"`
	// ReadAheadSacrificeChapter is the final lore chapter named by a recognized
	// "Read ahead" reminder ("Sacrifice after <chapter>"), or 0 when the reminder
	// omits the sacrifice clause. The chapter is a typed semantic value derived
	// from the parser's roman-numeral grammar.
	ReadAheadSacrificeChapter int `json:",omitempty"`
	// SagaReminder reports that this ability is a Saga's intrinsic lore-counter
	// reminder ("(As this Saga enters and after your draw step, add a lore
	// counter[. Sacrifice after <chapter>].)"). It is recognized only in Saga
	// context. Reminder text carries no game meaning, so downstream stages
	// consume this typed flag instead of re-reading the reminder wording.
	SagaReminder bool `json:",omitempty"`
	// ReadAheadRecognized reports that this ability's text is the canonical
	// "Read ahead" keyword line and reminder.
	ReadAheadRecognized bool `json:",omitempty"`
	// DevoidRecognized reports that this ability is exactly the canonical
	// "Devoid (This card has no color.)" keyword line. The Devoid reminder is
	// fixed boilerplate; downstream stages consume this typed flag instead of
	// re-reading the reminder wording.
	DevoidRecognized bool `json:",omitempty"`
	// Atoms holds the source-spanned typed semantic atoms recognized within this
	// ability's semantic tokens. Downstream stages consume these typed values by
	// span instead of re-recognizing Oracle spelling.
	Atoms Atoms `json:",omitzero"`
	// reminderInner is the parsed inner content of a fully-parenthesized reminder
	// ability ("(...)"). The parser parses the text inside the outer parentheses
	// once so a consumer lowers a reminder mana ability from typed data instead of
	// re-parsing the reminder wording itself. It is nil for non-reminder abilities
	// and for reminder text that is not fully parenthesized.
	reminderInner *reminderInner
}

// reminderInner carries a reminder ability's parsed inner document together with
// the diagnostics that inner parse produced, so a consumer reproduces the exact
// fail-closed behavior of re-parsing the reminder text without doing so itself.
type reminderInner struct {
	document    Document
	diagnostics []shared.Diagnostic
}

// ReminderInner returns the parsed inner content of a fully-parenthesized
// reminder ability, the diagnostics its inner parse produced, and whether such
// inner content exists. Consumers lower the typed inner document instead of
// re-parsing the reminder's Oracle text.
func (a *Ability) ReminderInner() (Document, []shared.Diagnostic, bool) {
	if a.reminderInner == nil {
		return Document{}, nil, false
	}
	return a.reminderInner.document, a.reminderInner.diagnostics, true
}

// ActivationRestrictionKind identifies a typed trailing activation restriction.
type ActivationRestrictionKind string

// Activation restriction syntax recognized by the parser.
const (
	ActivationRestrictionUnknown       ActivationRestrictionKind = ""
	ActivationRestrictionUnsupported   ActivationRestrictionKind = "ActivationRestrictionUnsupported"
	ActivationRestrictionSorceryTiming ActivationRestrictionKind = "ActivationRestrictionSorceryTiming"
	ActivationRestrictionFrequency     ActivationRestrictionKind = "ActivationRestrictionFrequency"
	ActivationRestrictionPhaseStep     ActivationRestrictionKind = "ActivationRestrictionPhaseStep"
)

// ActivationFrequencyCountKind identifies how many activations are permitted.
type ActivationFrequencyCountKind string

// Activation frequency counts recognized by the parser.
const (
	ActivationFrequencyCountUnknown ActivationFrequencyCountKind = ""
	ActivationFrequencyCountOnce    ActivationFrequencyCountKind = "ActivationFrequencyCountOnce"
)

// ActivationFrequencyPeriodKind identifies the period over which a frequency
// applies.
type ActivationFrequencyPeriodKind string

// Activation frequency periods recognized by the parser.
const (
	ActivationFrequencyPeriodUnknown ActivationFrequencyPeriodKind = ""
	ActivationFrequencyPeriodTurn    ActivationFrequencyPeriodKind = "ActivationFrequencyPeriodTurn"
)

// ActivationFrequencyCount is a source-spanned activation count.
type ActivationFrequencyCount struct {
	Kind ActivationFrequencyCountKind `json:",omitempty"`
	Span shared.Span                  `json:"-"`
}

// ActivationFrequencyPeriod is a source-spanned activation period.
type ActivationFrequencyPeriod struct {
	Kind ActivationFrequencyPeriodKind `json:",omitempty"`
	Span shared.Span                   `json:"-"`
}

// ActivationFrequencyRestriction is a composable typed activation frequency.
type ActivationFrequencyRestriction struct {
	Span   shared.Span               `json:"-"`
	Count  ActivationFrequencyCount  `json:",omitzero"`
	Period ActivationFrequencyPeriod `json:",omitzero"`
}

// ActivationPhaseStepRestriction is a composable typed phase or step
// restriction.
type ActivationPhaseStepRestriction struct {
	Span       shared.Span           `json:"-"`
	Quantifier PhaseStepQuantifier   `json:",omitzero"`
	Player     TriggerPlayerSelector `json:",omitzero"`
	Name       PhaseStepName         `json:",omitzero"`
}

// ActivationRestriction is source-spanned typed syntax for one trailing
// "Activate only" sentence. Unsupported preserves a restriction sentence that
// has recognized framing but unavailable or ambiguous inner grammar.
type ActivationRestriction struct {
	Kind        ActivationRestrictionKind      `json:",omitempty"`
	Span        shared.Span                    `json:"-"`
	SorcerySpan shared.Span                    `json:"-"`
	Frequency   ActivationFrequencyRestriction `json:",omitzero"`
	PhaseStep   ActivationPhaseStepRestriction `json:",omitzero"`
}

// Phrase is a meaningful contiguous token range.
type Phrase struct {
	Span   shared.Span    `json:"-"`
	Text   string         `json:",omitempty"`
	Tokens []shared.Token `json:"-"`
}

// AbilityWordClause is the recognized ability-word label at the start of an
// ability with its source span. A nil pointer means the ability has no ability
// word. The label is the rendered source text; recognition of which labels are
// meaningful belongs to consumers, not to this syntax node.
type AbilityWordClause struct {
	Label string      `json:",omitempty"`
	Span  shared.Span `json:"-"`
	// SeparatorSpan is the source span of the em dash that separates the ability
	// word label from the ability's body. Consumers slice the body off the token
	// stream at this typed boundary instead of scanning for the em dash.
	SeparatorSpan shared.Span `json:"-"`
}

// TriggerIntroductionKind identifies a trigger clause's leading word.
type TriggerIntroductionKind string

// Trigger introductions recognized by the syntax parser.
const (
	TriggerIntroductionUnknown  TriggerIntroductionKind = ""
	TriggerIntroductionWhen     TriggerIntroductionKind = "TriggerIntroductionWhen"
	TriggerIntroductionWhenever TriggerIntroductionKind = "TriggerIntroductionWhenever"
	TriggerIntroductionAt       TriggerIntroductionKind = "TriggerIntroductionAt"
)

// TriggerIntroduction is the source-spanned leading word of a trigger clause.
type TriggerIntroduction struct {
	Kind TriggerIntroductionKind `json:",omitempty"`
	Span shared.Span             `json:"-"`
}

// TriggerClause is the source-spanned syntax before a triggered ability's first
// top-level body comma. Event preserves unrecognized syntax as source metadata;
// typed event-family clauses carry recognized grammar.
type TriggerClause struct {
	Span         shared.Span         `json:"-"`
	Text         string              `json:",omitempty"`
	Tokens       []shared.Token      `json:"-"`
	Introduction TriggerIntroduction `json:",omitzero"`
	// Event is the rendered event text following the introduction word. It is
	// diagnostic source metadata; the recognized grammar lives in the typed
	// event-family clauses. EventSpan locates it in source.
	Event     string      `json:",omitempty"`
	EventSpan shared.Span `json:"-"`
	// eventTokens preserves the event phrase's tokens for parser-internal
	// recognition; the compiler consumes only the rendered Event/EventSpan.
	eventTokens  []shared.Token
	PhaseStep    *PhaseStepTriggerClause   `json:",omitempty"`
	PlayerEvent  *PlayerEventTriggerClause `json:",omitempty"`
	TriggerEvent *TriggerEventClause       `json:",omitempty"`
	// Order is the trigger clause's dense source-order rank, used downstream to
	// bind references within the trigger body without byte offsets.
	Order shared.SourceOrder `json:"-"`
}

// TriggerEventKind identifies a typed trigger event clause family.
type TriggerEventKind string

// Trigger event families recognized by the syntax parser.
const (
	TriggerEventKindUnknown          TriggerEventKind = ""
	TriggerEventKindZoneChange       TriggerEventKind = "TriggerEventKindZoneChange"
	TriggerEventKindSpellCast        TriggerEventKind = "TriggerEventKindSpellCast"
	TriggerEventKindAbilityActivated TriggerEventKind = "TriggerEventKindAbilityActivated"
	TriggerEventKindAttack           TriggerEventKind = "TriggerEventKindAttack"
	TriggerEventKindBlock            TriggerEventKind = "TriggerEventKindBlock"
	TriggerEventKindBecameBlocked    TriggerEventKind = "TriggerEventKindBecameBlocked"
	TriggerEventKindDamageDealt      TriggerEventKind = "TriggerEventKindDamageDealt"
	TriggerEventKindCounterAdded     TriggerEventKind = "TriggerEventKindCounterAdded"
	TriggerEventKindBecomesTapped    TriggerEventKind = "TriggerEventKindBecomesTapped"
	TriggerEventKindBecomesUntapped  TriggerEventKind = "TriggerEventKindBecomesUntapped"
	TriggerEventKindTurnedFaceUp     TriggerEventKind = "TriggerEventKindTurnedFaceUp"
	TriggerEventKindSacrificed       TriggerEventKind = "TriggerEventKindSacrificed"
	TriggerEventKindMutated          TriggerEventKind = "TriggerEventKindMutated"
	TriggerEventKindBecameTarget     TriggerEventKind = "TriggerEventKindBecameTarget"
)

// TriggerEventSubjectKind identifies the grammatical subject in a trigger event.
type TriggerEventSubjectKind string

// Trigger event subject kinds recognized by the syntax parser.
const (
	TriggerEventSubjectUnknown      TriggerEventSubjectKind = ""
	TriggerEventSubjectSelf         TriggerEventSubjectKind = "TriggerEventSubjectSelf"
	TriggerEventSubjectAttached     TriggerEventSubjectKind = "TriggerEventSubjectAttached"
	TriggerEventSubjectDamageSource TriggerEventSubjectKind = "TriggerEventSubjectDamageSource"
	TriggerEventSubjectSelection    TriggerEventSubjectKind = "TriggerEventSubjectSelection"
)

// TriggerEventAttachKind identifies the attachment relation in an attached subject.
type TriggerEventAttachKind string

// Attachment relations recognized in trigger-event subjects.
const (
	TriggerEventAttachUnknown   TriggerEventAttachKind = ""
	TriggerEventAttachEnchanted TriggerEventAttachKind = "TriggerEventAttachEnchanted"
	TriggerEventAttachEquipped  TriggerEventAttachKind = "TriggerEventAttachEquipped"
	TriggerEventAttachFortified TriggerEventAttachKind = "TriggerEventAttachFortified"
)

// TriggerEventSubject is a source-spanned trigger-event subject.
type TriggerEventSubject struct {
	Kind       TriggerEventSubjectKind `json:",omitempty"`
	Span       shared.Span             `json:"-"`
	AttachKind TriggerEventAttachKind  `json:",omitempty"`
	Selection  TriggerSelection        `json:",omitzero"`
}

// TriggerEventActorKind identifies an acting player in a trigger event.
type TriggerEventActorKind string

// Acting-player kinds recognized by the syntax parser.
const (
	TriggerEventActorUnknown  TriggerEventActorKind = ""
	TriggerEventActorYou      TriggerEventActorKind = "TriggerEventActorYou"
	TriggerEventActorPlayer   TriggerEventActorKind = "TriggerEventActorPlayer"
	TriggerEventActorOpponent TriggerEventActorKind = "TriggerEventActorOpponent"
)

// TriggerEventActor is a source-spanned acting player.
type TriggerEventActor struct {
	Kind TriggerEventActorKind `json:",omitempty"`
	Span shared.Span           `json:"-"`
}

// TriggerEventZoneKind identifies a zone mentioned by trigger-event syntax.
type TriggerEventZoneKind string

// Trigger-event zones recognized by the syntax parser.
const (
	TriggerEventZoneNone        TriggerEventZoneKind = ""
	TriggerEventZoneBattlefield TriggerEventZoneKind = "TriggerEventZoneBattlefield"
	TriggerEventZoneGraveyard   TriggerEventZoneKind = "TriggerEventZoneGraveyard"
	TriggerEventZoneHand        TriggerEventZoneKind = "TriggerEventZoneHand"
	TriggerEventZoneExile       TriggerEventZoneKind = "TriggerEventZoneExile"
	TriggerEventZoneLibrary     TriggerEventZoneKind = "TriggerEventZoneLibrary"
	TriggerEventZoneStack       TriggerEventZoneKind = "TriggerEventZoneStack"
	TriggerEventZoneCommand     TriggerEventZoneKind = "TriggerEventZoneCommand"
)

// TriggerEventZone is a source-spanned zone.
type TriggerEventZone struct {
	Kind TriggerEventZoneKind `json:",omitempty"`
	Span shared.Span          `json:"-"`
}

// TriggerEventZoneChangeKind identifies the rules event represented by a
// permanent zone-change production.
type TriggerEventZoneChangeKind string

// Permanent zone-change productions recognized by trigger-event syntax.
const (
	TriggerEventZoneChangeUnknown            TriggerEventZoneChangeKind = ""
	TriggerEventZoneChangeEnteredBattlefield TriggerEventZoneChangeKind = "TriggerEventZoneChangeEnteredBattlefield"
	TriggerEventZoneChangeDied               TriggerEventZoneChangeKind = "TriggerEventZoneChangeDied"
	TriggerEventZoneChangeMoved              TriggerEventZoneChangeKind = "TriggerEventZoneChangeMoved"
)

// TriggerEventZoneChange is the source-spanned operation in a zone-change
// event.
type TriggerEventZoneChange struct {
	Kind TriggerEventZoneChangeKind `json:",omitempty"`
	Span shared.Span                `json:"-"`
}

// TriggerEventZoneContext is composable zone-change context.
type TriggerEventZoneContext struct {
	Span          shared.Span      `json:"-"`
	MatchFromZone bool             `json:",omitempty"`
	FromZone      TriggerEventZone `json:",omitzero"`
	MatchToZone   bool             `json:",omitempty"`
	ToZone        TriggerEventZone `json:",omitzero"`
	ExcludeToZone bool             `json:",omitempty"`
}

// TriggerEventTappedStateKind identifies an ETB tapped-state qualifier.
type TriggerEventTappedStateKind string

// Tapped-state qualifiers recognized by trigger-event syntax.
const (
	TriggerEventTappedStateAny      TriggerEventTappedStateKind = ""
	TriggerEventTappedStateTapped   TriggerEventTappedStateKind = "TriggerEventTappedStateTapped"
	TriggerEventTappedStateUntapped TriggerEventTappedStateKind = "TriggerEventTappedStateUntapped"
)

// TriggerEventTappedState is a source-spanned ETB tapped-state qualifier.
type TriggerEventTappedState struct {
	Kind TriggerEventTappedStateKind `json:",omitempty"`
	Span shared.Span                 `json:"-"`
}

// TriggerEventCombatQualifierKind identifies a damage qualifier.
type TriggerEventCombatQualifierKind string

// Combat qualifiers recognized by the syntax parser.
const (
	TriggerEventCombatQualifierAny       TriggerEventCombatQualifierKind = ""
	TriggerEventCombatQualifierCombat    TriggerEventCombatQualifierKind = "TriggerEventCombatQualifierCombat"
	TriggerEventCombatQualifierNoncombat TriggerEventCombatQualifierKind = "TriggerEventCombatQualifierNoncombat"
)

// TriggerEventCombatQualifier is a source-spanned damage qualifier.
type TriggerEventCombatQualifier struct {
	Kind TriggerEventCombatQualifierKind `json:",omitempty"`
	Span shared.Span                     `json:"-"`
}

// TriggerEventDamageRecipientKind identifies a damage recipient category.
type TriggerEventDamageRecipientKind uint8

// Damage recipient categories recognized by the syntax parser.
const (
	TriggerEventDamageRecipientNone      TriggerEventDamageRecipientKind = 0
	TriggerEventDamageRecipientPlayer    TriggerEventDamageRecipientKind = 1
	TriggerEventDamageRecipientPermanent TriggerEventDamageRecipientKind = 2
)

// TriggerEventDamageRecipient is composable typed damage-recipient syntax.
type TriggerEventDamageRecipient struct {
	Kind      TriggerEventDamageRecipientKind `json:",omitempty"`
	Span      shared.Span                     `json:"-"`
	Player    TriggerPlayerSelector           `json:",omitzero"`
	Selection TriggerSelection                `json:",omitzero"`
	IsSource  bool                            `json:",omitempty"`
}

// TriggerEventAttackRecipientKind identifies an attack recipient category.
type TriggerEventAttackRecipientKind uint8

// Attack recipient categories recognized by the syntax parser.
const (
	TriggerEventAttackRecipientAny          TriggerEventAttackRecipientKind = 0
	TriggerEventAttackRecipientPlayer       TriggerEventAttackRecipientKind = 1
	TriggerEventAttackRecipientPlaneswalker TriggerEventAttackRecipientKind = 2
	TriggerEventAttackRecipientBattle       TriggerEventAttackRecipientKind = 4
)

// TriggerEventAttackRecipient is composable typed attack-recipient syntax.
type TriggerEventAttackRecipient struct {
	Kind      TriggerEventAttackRecipientKind `json:",omitempty"`
	Span      shared.Span                     `json:"-"`
	Player    TriggerPlayerSelector           `json:",omitzero"`
	Selection TriggerSelection                `json:",omitzero"`
}

// TriggerEventStackObjectKind identifies a triggering stack object.
type TriggerEventStackObjectKind string

// Stack-object kinds recognized by the syntax parser.
const (
	TriggerEventStackObjectAny   TriggerEventStackObjectKind = ""
	TriggerEventStackObjectSpell TriggerEventStackObjectKind = "TriggerEventStackObjectSpell"
)

// TriggerEventStackObject is a source-spanned stack-object selector.
type TriggerEventStackObject struct {
	Kind TriggerEventStackObjectKind `json:",omitempty"`
	Span shared.Span                 `json:"-"`
}

// TriggerEventCounterKind identifies a supported counter type.
type TriggerEventCounterKind string

// Counter kinds recognized by trigger-event syntax.
const (
	TriggerEventCounterAny              TriggerEventCounterKind = ""
	TriggerEventCounterPlusOnePlusOne   TriggerEventCounterKind = "TriggerEventCounterPlusOnePlusOne"
	TriggerEventCounterMinusOneMinusOne TriggerEventCounterKind = "TriggerEventCounterMinusOneMinusOne"
)

// TriggerEventCounter is a source-spanned counter kind.
type TriggerEventCounter struct {
	Kind TriggerEventCounterKind `json:",omitempty"`
	Span shared.Span             `json:"-"`
}

// TriggerEventSpellSelection is composable typed spell-selection syntax.
type TriggerEventSpellSelection struct {
	Span             shared.Span       `json:"-"`
	Types            []TriggerCardType `json:",omitempty"`
	TypesAny         []TriggerCardType `json:",omitempty"`
	ExcludedTypes    []TriggerCardType `json:",omitempty"`
	ColorsAny        []TriggerColor    `json:",omitempty"`
	SubtypesAny      []TriggerSubtype  `json:",omitempty"`
	Colorless        bool              `json:",omitempty"`
	Multicolored     bool              `json:",omitempty"`
	Kicker           bool              `json:",omitempty"`
	Historic         bool              `json:",omitempty"`
	ManaValueAtLeast int               `json:",omitempty"`
	MatchManaValue   bool              `json:",omitempty"`
	FromZone         TriggerEventZone  `json:",omitzero"`
	// Ordinal records a per-turn spell-cast position from "your Nth spell each
	// turn" wording (1 for first, 2 for second, ...). Zero means no ordinal
	// qualifier. Recognized only with the controller-scoped "you cast" actor.
	Ordinal int `json:",omitempty"`
}

// TriggerEventClause is composable typed syntax for a trigger event.
type TriggerEventClause struct {
	Span                       shared.Span                 `json:"-"`
	Kind                       TriggerEventKind            `json:",omitempty"`
	Subject                    TriggerEventSubject         `json:",omitzero"`
	Actor                      TriggerEventActor           `json:",omitzero"`
	ZoneChange                 TriggerEventZoneChange      `json:",omitzero"`
	Zone                       TriggerEventZoneContext     `json:",omitzero"`
	Tapped                     TriggerEventTappedState     `json:",omitzero"`
	SpellSelection             TriggerEventSpellSelection  `json:",omitzero"`
	DamageSourceSpellSelection TriggerEventSpellSelection  `json:",omitzero"`
	SourceSelection            TriggerSelection            `json:",omitzero"`
	DamageSource               TriggerEventSubject         `json:",omitzero"`
	DamageRecipient            TriggerEventDamageRecipient `json:",omitzero"`
	CombatQualifier            TriggerEventCombatQualifier `json:",omitzero"`
	AttackRecipient            TriggerEventAttackRecipient `json:",omitzero"`
	RelatedSelection           TriggerSelection            `json:",omitzero"`
	Counter                    TriggerEventCounter         `json:",omitzero"`
	StackObject                TriggerEventStackObject     `json:",omitzero"`
	CauseController            TriggerEventActorKind       `json:",omitempty"`
	Controller                 TriggerController           `json:",omitempty"`
	Player                     TriggerPlayerSelector       `json:",omitzero"`
	OneOrMore                  bool                        `json:",omitempty"`
	ExcludeSelf                bool                        `json:",omitempty"`
	// SelfOrAnother marks a zone-change clause whose subject is the union of the
	// ability's own source and another permanent matching Subject's Selection,
	// e.g. "this creature or another Ally you control enters/dies". The trigger
	// fires for the source itself as well as for a matching other permanent.
	SelfOrAnother             bool `json:",omitempty"`
	FaceDown                  bool `json:",omitempty"`
	ExcludeManaAbility        bool `json:",omitempty"`
	DamageSourceIsStackObject bool `json:",omitempty"`
	OneOrMorePerAttackTarget  bool `json:",omitempty"`
	// AttackAlone marks an attacker-declared clause restricted to a creature
	// that attacks alone, i.e. the only attacking creature this combat ("attacks
	// alone", CR 506.5 / the Exalted wording).
	AttackAlone bool `json:",omitempty"`
	// AttackerCountAtLeast restricts a controller-scoped attack clause to combats
	// where the controller attacks with at least this many creatures ("attack
	// with two or more creatures"). Zero imposes no minimum.
	AttackerCountAtLeast int `json:",omitempty"`
	// MatchCopy is set on a spell-cast clause whose "cast or copy" wording also
	// matches spell copies (CR 707, magecraft).
	MatchCopy bool `json:",omitempty"`
}

// EventHistoryWindowKind identifies the turn window for an event-history
// condition.
type EventHistoryWindowKind string

// Event-history windows recognized by the syntax parser.
const (
	EventHistoryWindowUnknown      EventHistoryWindowKind = ""
	EventHistoryWindowCurrentTurn  EventHistoryWindowKind = "EventHistoryWindowCurrentTurn"
	EventHistoryWindowPreviousTurn EventHistoryWindowKind = "EventHistoryWindowPreviousTurn"
)

// EventHistoryWindow is a source-spanned turn window.
type EventHistoryWindow struct {
	Kind EventHistoryWindowKind `json:",omitempty"`
	Span shared.Span            `json:"-"`
}

// EventHistoryCondition is composable typed syntax for a condition that queries
// whether a supported event occurred in a turn window.
type EventHistoryCondition struct {
	Span         shared.Span               `json:"-"`
	Negated      bool                      `json:",omitempty"`
	NegationSpan shared.Span               `json:"-"`
	Window       EventHistoryWindow        `json:",omitzero"`
	TriggerEvent *TriggerEventClause       `json:",omitempty"`
	PlayerEvent  *PlayerEventTriggerClause `json:",omitempty"`
}

// PhaseStepQuantifierKind identifies a phase or step clause's grammatical
// cardinality.
type PhaseStepQuantifierKind string

// Phase and step quantifiers recognized by the syntax parser.
const (
	PhaseStepQuantifierUnknown PhaseStepQuantifierKind = ""
	PhaseStepQuantifierNone    PhaseStepQuantifierKind = "PhaseStepQuantifierNone"
	PhaseStepQuantifierSingle  PhaseStepQuantifierKind = "PhaseStepQuantifierSingle"
	PhaseStepQuantifierEach    PhaseStepQuantifierKind = "PhaseStepQuantifierEach"
	PhaseStepQuantifierEachOf  PhaseStepQuantifierKind = "PhaseStepQuantifierEachOf"
)

// PhaseStepQuantifier is a source-spanned phase or step quantifier.
type PhaseStepQuantifier struct {
	Kind PhaseStepQuantifierKind `json:",omitempty"`
	Span shared.Span             `json:"-"`
}

// TriggerPlayerSelectorKind identifies a trigger's acting player or controller.
type TriggerPlayerSelectorKind string

// Player/controller selectors recognized by the trigger-clause grammar.
const (
	TriggerPlayerSelectorUnknown            TriggerPlayerSelectorKind = ""
	TriggerPlayerSelectorAny                TriggerPlayerSelectorKind = "TriggerPlayerSelectorAny"
	TriggerPlayerSelectorYou                TriggerPlayerSelectorKind = "TriggerPlayerSelectorYou"
	TriggerPlayerSelectorOpponent           TriggerPlayerSelectorKind = "TriggerPlayerSelectorOpponent"
	TriggerPlayerSelectorSourceController   TriggerPlayerSelectorKind = "TriggerPlayerSelectorSourceController"
	TriggerPlayerSelectorAttachedController TriggerPlayerSelectorKind = "TriggerPlayerSelectorAttachedController"
)

// TriggerAttachedSubject is the typed subject in an attached-controller selector.
type TriggerAttachedSubject struct {
	Span      shared.Span      `json:"-"`
	Selection TriggerSelection `json:",omitzero"`
}

// TriggerPlayerSelector is a source-spanned player/controller selector shared
// across typed trigger families.
type TriggerPlayerSelector struct {
	Kind            TriggerPlayerSelectorKind `json:",omitempty"`
	Span            shared.Span               `json:"-"`
	AttachedSubject TriggerAttachedSubject    `json:",omitzero"`
}

// PhaseStepNameKind identifies a literal phase or step name.
type PhaseStepNameKind string

// Literal phase and step names recognized by the syntax parser.
const (
	PhaseStepNameUnknown             PhaseStepNameKind = ""
	PhaseStepNameUpkeep              PhaseStepNameKind = "PhaseStepNameUpkeep"
	PhaseStepNameDrawStep            PhaseStepNameKind = "PhaseStepNameDrawStep"
	PhaseStepNameEndStep             PhaseStepNameKind = "PhaseStepNameEndStep"
	PhaseStepNameCombat              PhaseStepNameKind = "PhaseStepNameCombat"
	PhaseStepNameCombatStep          PhaseStepNameKind = "PhaseStepNameCombatStep"
	PhaseStepNameEndOfCombat         PhaseStepNameKind = "PhaseStepNameEndOfCombat"
	PhaseStepNameEndOfCombatStep     PhaseStepNameKind = "PhaseStepNameEndOfCombatStep"
	PhaseStepNamePrecombatMainPhase  PhaseStepNameKind = "PhaseStepNamePrecombatMainPhase"
	PhaseStepNamePostcombatMainPhase PhaseStepNameKind = "PhaseStepNamePostcombatMainPhase"
	PhaseStepNameFirstMainPhase      PhaseStepNameKind = "PhaseStepNameFirstMainPhase"
	PhaseStepNameSecondMainPhase     PhaseStepNameKind = "PhaseStepNameSecondMainPhase"
)

// PhaseStepName is a source-spanned literal phase or step name.
type PhaseStepName struct {
	Kind PhaseStepNameKind `json:",omitempty"`
	Span shared.Span       `json:"-"`
}

// PhaseStepTriggerClause is composable typed syntax for a phase or step event.
type PhaseStepTriggerClause struct {
	Span       shared.Span           `json:"-"`
	Quantifier PhaseStepQuantifier   `json:",omitzero"`
	Player     TriggerPlayerSelector `json:",omitzero"`
	Name       PhaseStepName         `json:",omitzero"`
}

// PlayerEventActionKind identifies an acting player's event.
type PlayerEventActionKind string

// Player-event actions recognized by the syntax parser.
const (
	PlayerEventActionUnknown        PlayerEventActionKind = ""
	PlayerEventActionDraw           PlayerEventActionKind = "PlayerEventActionDraw"
	PlayerEventActionDiscard        PlayerEventActionKind = "PlayerEventActionDiscard"
	PlayerEventActionCycle          PlayerEventActionKind = "PlayerEventActionCycle"
	PlayerEventActionCycleOrDiscard PlayerEventActionKind = "PlayerEventActionCycleOrDiscard"
	PlayerEventActionScry           PlayerEventActionKind = "PlayerEventActionScry"
	PlayerEventActionSurveil        PlayerEventActionKind = "PlayerEventActionSurveil"
	PlayerEventActionGainLife       PlayerEventActionKind = "PlayerEventActionGainLife"
	PlayerEventActionLoseLife       PlayerEventActionKind = "PlayerEventActionLoseLife"
)

// PlayerEventAction is a source-spanned player-event action.
type PlayerEventAction struct {
	Kind PlayerEventActionKind `json:",omitempty"`
	Span shared.Span           `json:"-"`
}

// PlayerEventCardKind identifies an action's grammatical card object.
type PlayerEventCardKind string

// Player-event card-object modifiers recognized by the syntax parser.
const (
	PlayerEventCardUnknown   PlayerEventCardKind = ""
	PlayerEventCardNone      PlayerEventCardKind = "PlayerEventCardNone"
	PlayerEventCardSingle    PlayerEventCardKind = "PlayerEventCardSingle"
	PlayerEventCardOneOrMore PlayerEventCardKind = "PlayerEventCardOneOrMore"
	PlayerEventCardAnother   PlayerEventCardKind = "PlayerEventCardAnother"
)

// PlayerEventCard is a source-spanned player-event card-object modifier.
type PlayerEventCard struct {
	Kind PlayerEventCardKind `json:",omitempty"`
	Span shared.Span         `json:"-"`

	// RequiredTypes and ExcludedTypes record a card-type filter on the event's
	// card object, such as "a creature card" or "a noncreature, nonland card".
	RequiredTypes []TriggerCardType `json:",omitempty"`
	ExcludedTypes []TriggerCardType `json:",omitempty"`
}

// PlayerEventOccurrenceKind identifies an event's supported turn-relative
// occurrence restriction.
type PlayerEventOccurrenceKind string

// Player-event occurrence restrictions recognized by the syntax parser.
const (
	PlayerEventOccurrenceUnknown         PlayerEventOccurrenceKind = ""
	PlayerEventOccurrenceAny             PlayerEventOccurrenceKind = "PlayerEventOccurrenceAny"
	PlayerEventOccurrenceFirstEachTurn   PlayerEventOccurrenceKind = "PlayerEventOccurrenceFirstEachTurn"
	PlayerEventOccurrenceOrdinalEachTurn PlayerEventOccurrenceKind = "PlayerEventOccurrenceOrdinalEachTurn"
)

// PlayerEventOccurrence is a source-spanned player-event occurrence modifier.
type PlayerEventOccurrence struct {
	Kind    PlayerEventOccurrenceKind `json:",omitempty"`
	Span    shared.Span               `json:"-"`
	Ordinal int                       `json:",omitempty"`
}

// PlayerEventTriggerClause is composable typed syntax for an acting-player event.
type PlayerEventTriggerClause struct {
	Span       shared.Span           `json:"-"`
	Player     TriggerPlayerSelector `json:",omitzero"`
	Action     PlayerEventAction     `json:",omitzero"`
	Card       PlayerEventCard       `json:",omitzero"`
	Occurrence PlayerEventOccurrence `json:",omitzero"`
}

// Sentence is a top-level sentence in an ability.
type Sentence struct {
	Span          shared.Span       `json:"-"`
	Text          string            `json:",omitempty"`
	Tokens        []shared.Token    `json:"-"`
	StaticRule    *StaticRuleSyntax `json:",omitempty"`
	Targets       []TargetSyntax    `json:",omitempty"`
	Effects       []EffectSyntax    `json:",omitempty"`
	LegacyEffects bool              `json:",omitempty"`
	// RegenerationRider reports that this sentence is a credited regeneration
	// rider ("It/They can't be regenerated.") folded onto a preceding destroy
	// effect. Reference and coverage scans treat its pronoun and tokens as
	// belonging to that destroy rather than as an unrecognized sibling.
	RegenerationRider bool `json:",omitempty"`
}

// StaticRuleSubjectKind identifies the source object constrained by a simple
// static rule.
type StaticRuleSubjectKind string

// Simple static-rule subjects.
const (
	StaticRuleSubjectUnknown        StaticRuleSubjectKind = ""
	StaticRuleSubjectSourceCreature StaticRuleSubjectKind = "StaticRuleSubjectSourceCreature"
	StaticRuleSubjectSourceSpell    StaticRuleSubjectKind = "StaticRuleSubjectSourceSpell"
	StaticRuleSubjectAttachedObject StaticRuleSubjectKind = "StaticRuleSubjectAttachedObject"
)

// StaticRuleConstraintKind identifies whether a rule prohibits or requires an
// operation.
type StaticRuleConstraintKind string

// Simple static-rule constraints.
const (
	StaticRuleConstraintUnknown     StaticRuleConstraintKind = ""
	StaticRuleConstraintProhibition StaticRuleConstraintKind = "StaticRuleConstraintProhibition"
	StaticRuleConstraintRequirement StaticRuleConstraintKind = "StaticRuleConstraintRequirement"
)

// StaticRuleOperationKind identifies the rules operation being constrained.
type StaticRuleOperationKind string

// Simple static-rule operations.
const (
	StaticRuleOperationUnknown       StaticRuleOperationKind = ""
	StaticRuleOperationAttack        StaticRuleOperationKind = "StaticRuleOperationAttack"
	StaticRuleOperationBlock         StaticRuleOperationKind = "StaticRuleOperationBlock"
	StaticRuleOperationCounter       StaticRuleOperationKind = "StaticRuleOperationCounter"
	StaticRuleOperationAttackOrBlock StaticRuleOperationKind = "StaticRuleOperationAttackOrBlock"
	StaticRuleOperationUntap         StaticRuleOperationKind = "StaticRuleOperationUntap"
)

// StaticRuleVoice identifies the grammatical role the subject has in an
// operation.
type StaticRuleVoice string

// Simple static-rule voices.
const (
	StaticRuleVoiceUnknown StaticRuleVoice = ""
	StaticRuleVoiceActive  StaticRuleVoice = "StaticRuleVoiceActive"
	StaticRuleVoicePassive StaticRuleVoice = "StaticRuleVoicePassive"
)

// StaticRuleQualifierKind identifies a composable restriction on an operation.
type StaticRuleQualifierKind string

// Simple static-rule qualifiers.
const (
	StaticRuleQualifierUnknown    StaticRuleQualifierKind = ""
	StaticRuleQualifierEachCombat StaticRuleQualifierKind = "StaticRuleQualifierEachCombat"
	StaticRuleQualifierIfAble     StaticRuleQualifierKind = "StaticRuleQualifierIfAble"
)

// StaticRuleSubject is a source-spanned simple static-rule subject.
type StaticRuleSubject struct {
	Kind  StaticRuleSubjectKind `json:",omitempty"`
	Span  shared.Span           `json:"-"`
	Order shared.SourceOrder    `json:"-"`
}

// StaticRuleConstraint is a source-spanned requirement or prohibition.
type StaticRuleConstraint struct {
	Kind StaticRuleConstraintKind `json:",omitempty"`
	Span shared.Span              `json:"-"`
}

// StaticRuleOperation is a source-spanned operation and the subject's
// grammatical role in it.
type StaticRuleOperation struct {
	Kind  StaticRuleOperationKind `json:",omitempty"`
	Voice StaticRuleVoice         `json:",omitempty"`
	Span  shared.Span             `json:"-"`
	Order shared.SourceOrder      `json:"-"`
}

// StaticRuleQualifier is a source-spanned restriction on a rule operation.
type StaticRuleQualifier struct {
	Kind StaticRuleQualifierKind `json:",omitempty"`
	Span shared.Span             `json:"-"`
}

// StaticRuleSyntax is a composable typed simple static-rule declaration.
// Sentence text and tokens remain available only as source metadata.
type StaticRuleSyntax struct {
	Span       shared.Span           `json:"-"`
	Subject    StaticRuleSubject     `json:",omitzero"`
	Constraint StaticRuleConstraint  `json:",omitzero"`
	Operation  StaticRuleOperation   `json:",omitzero"`
	Qualifiers []StaticRuleQualifier `json:",omitempty"`
	// Order is the rule's dense source-order rank (of Span), used downstream to
	// order static-rule effects without byte offsets.
	Order shared.SourceOrder `json:"-"`
}

// Delimited is parenthesized reminder text or a quoted granted ability.
type Delimited struct {
	Span   shared.Span    `json:"-"`
	Text   string         `json:",omitempty"`
	Tokens []shared.Token `json:"-"`
}

// Modal is a choose header followed by bullet or inline options.
type Modal struct {
	header  Phrase
	Options []Mode `json:",omitempty"`
	Atoms   Atoms  `json:",omitzero"`
	// MinModes and MaxModes are the recognized choice range of the choose
	// header (e.g. "Choose two —" yields 2/2 and "Choose one or both —" yields
	// 1/2). They are populated only when ChoiceKnown is true; downstream code
	// must consume these typed fields instead of re-reading header tokens.
	MinModes    int  `json:",omitempty"`
	MaxModes    int  `json:",omitempty"`
	ChoiceKnown bool `json:",omitempty"`
}

// Mode is one bullet option in a modal ability.
type Mode struct {
	Span                   shared.Span             `json:"-"`
	Text                   string                  `json:",omitempty"`
	Tokens                 []shared.Token          `json:"-"`
	Sentences              []Sentence              `json:",omitempty"`
	ConditionBoundaries    []ConditionBoundary     `json:",omitempty"`
	EventHistoryConditions []EventHistoryCondition `json:",omitempty"`
	ConditionClauses       []ConditionClause       `json:",omitempty"`
	ConditionSegments      []ConditionSegment      `json:",omitempty"`
	SemanticReferences     []Reference             `json:",omitempty"`
	SemanticKeywords       []Keyword               `json:",omitempty"`
	Reminders              []Delimited             `json:"-"`
	Quoted                 []Delimited             `json:"-"`
	// Atoms holds the source-spanned typed semantic atoms recognized within this
	// mode's semantic tokens.
	Atoms Atoms `json:",omitzero"`
}
