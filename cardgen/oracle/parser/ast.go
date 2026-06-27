// Package parser recognizes the grammatical structure of Oracle text.
package parser

import (
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
)

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
	// AbilitySpellAlternativeCost is a spell paragraph that declares an
	// optional alternative to its printed mana cost.
	AbilitySpellAlternativeCost AbilityKind = "AbilitySpellAlternativeCost"
	// AbilityLevelBand is a leveler card's "LEVEL lo-hi" / "LEVEL lo+" band
	// header (CR 711). It carries the band's level range and printed P/T through
	// Ability.LevelBand and has no resolving body of its own.
	AbilityLevelBand AbilityKind = "AbilityLevelBand"
)

// Context supplies card-face facts that Oracle text alone cannot express.
type Context struct {
	InstantOrSorcery bool `json:",omitempty"`
	Planeswalker     bool `json:",omitempty"`
	Saga             bool `json:",omitempty"`
	// Class reports that the card is a Class enchantment. The parser uses it to
	// recognize the Class level-up activated abilities ("{cost}: Level N") and
	// the intrinsic Class reminder line, which share wording only Class cards use.
	Class bool `json:",omitempty"`
	// Leveler reports that the card is a leveler card (CR 711, layout "leveler").
	// The parser uses it to recognize the "Level up {cost}" activated ability and
	// the "LEVEL lo-hi" / "LEVEL lo+" band headers with their printed P/T, which
	// share wording only leveler cards use.
	Leveler bool `json:",omitempty"`
	// CardName is the card's own name. The parser uses it to recognize explicit
	// self-name references so the compiler need not inspect name spelling.
	CardName string `json:",omitempty"`
	// Legendary reports that the card is legendary. The parser uses it to admit
	// the legend's "<short name> of <place>" pre-"of" short name as an additional
	// self-name spelling (e.g. "Rosie Cotton" for "Rosie Cotton of South Lane"),
	// a proper-noun self reference only legendary names denote unambiguously.
	Legendary bool `json:",omitempty"`
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
	// ChapterFlavorSpan is the source span of a Saga chapter's flavor-name prefix
	// (the first em dash and the Title-Case proper name set off before the
	// effect, as in "I — Gungnir — Destroy ..."), or the zero span when the
	// chapter has no flavor name. The name is rules-free flavor; consumers account
	// for its tokens through this span so the effect body lowers as if absent.
	ChapterFlavorSpan shared.Span `json:"-"`
	// costPhrase is the source cost phrase recognized before the typed cost is
	// emitted. It is parser-internal: the compiler consumes the typed cost via
	// CostSyntax and only ever needs the cost's presence, not its tokens.
	costPhrase *Phrase
	// wardCostPhrase is the source cost phrase of a "Ward—<cost>" keyword whose
	// payment is a non-mana or composite cost (e.g. "Ward—{2}, Pay 2 life.",
	// "Ward—Pay 3 life.", "Ward—Sacrifice a creature."). It is parser-internal;
	// emitWardKeywordCost parses it into the typed WardCost carried on the Ward
	// keyword so the compiler consumes the cost components there.
	wardCostPhrase *Phrase
	Trigger        *TriggerClause `json:",omitempty"`
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
	CostSyntax                 *Cost                             `json:",omitempty"`
	SourceAbilityCostReduction *SourceAbilityCostReductionSyntax `json:",omitempty"`
	// AlternativeCost is the typed alternative spell-cost declaration, or nil
	// when this paragraph does not declare one.
	AlternativeCost *SpellAlternativeCost `json:",omitempty"`
	// ExactSequence is a parser-owned, exact-vocabulary resolving sequence.
	ExactSequence *ExactSequenceSyntax `json:",omitempty"`
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
	// Companion is the recognized companion keyword ability (CR 702.139), or nil
	// when this paragraph is not a companion ability. The parser owns the
	// companion wording (the standard "Companion — <deckbuilding condition>" form
	// and the "<X>'s companion" partner variant, as on Barbara Wright's "Doctor's
	// companion"); when it is set the paragraph's competing effect, keyword, and
	// declaration semantics are cleared so downstream stages consume only the
	// companion identity.
	Companion *CompanionClause `json:",omitempty"`
	// PartnerWith is the recognized "Partner with <name>" keyword ability (CR
	// 702.124e), or nil when this paragraph is not a partner-with ability. The
	// parser owns the "Partner with <name>" wording; when it is set the
	// paragraph's competing effect, keyword, and declaration semantics are
	// cleared so downstream stages consume only the partner-with identity.
	PartnerWith *PartnerWithClause `json:",omitempty"`
	// ChooseABackground is the recognized "Choose a Background" keyword ability
	// (CR 702.124f), or nil when this paragraph is not a choose-a-background
	// ability. The parser owns the "Choose a Background" wording; when it is set
	// the paragraph's competing effect, keyword, and declaration semantics are
	// cleared so downstream stages consume only the choose-a-background identity.
	ChooseABackground *ChooseABackgroundClause `json:",omitempty"`
	// Partner is the recognized "Partner" keyword ability (CR 702.124a) and its
	// "Partner—<quality>" restricted variants (CR 702.124f), or nil when this
	// paragraph is not a partner ability. The parser owns the "Partner" and
	// "Partner—<quality>" wording; when it is set the paragraph's competing
	// effect, keyword, declaration, and ability-word semantics are cleared so
	// downstream stages consume only the partner identity.
	Partner *PartnerClause `json:",omitempty"`
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
	// DiceTable is the recognized die-roll outcome table that follows this
	// ability's "Roll a d<N>." line ("1—9 | <effect>", "10—19 | <effect>",
	// "20 | <effect>"). It is nil for abilities without an outcome table. Each
	// row carries its inclusive result interval and its own resolving sentences;
	// downstream stages consume these typed values instead of re-reading the
	// row wording or the em-dash/pipe glyphs.
	DiceTable *DiceTable `json:",omitempty"`
	// CoinFlip is the recognized "Flip a coin." outcome on this ability: a flip
	// sentence followed by one or both of "If you win the flip, <effect>." and
	// "If you lose the flip, <effect>." branches. It is nil for abilities without
	// a recognized coin flip. Each branch carries its own freshly parsed
	// resolving sentences so downstream stages consume typed effects rather than
	// re-reading the condition wording; the flip itself lowers to a fair
	// two-sided random draw whose result gates the win and lose branches.
	CoinFlip *CoinFlip `json:",omitempty"`
	// Vote is the recognized "Starting with you, each player votes for <A> or
	// <B>." voting construct (CR 701.32), or nil when the ability holds no vote.
	// Like CoinFlip, the recognizer re-parses each "If <option> gets more
	// votes[ or the vote is tied], ..." arm clause in isolation and sheds the
	// consumed sentences' effects and condition wording, so the construct lowers
	// to a Vote interaction whose tally gates each arm.
	Vote *VoteClause `json:",omitempty"`
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
	// ClassReminder reports that this ability is a Class enchantment's intrinsic
	// level-up reminder ("(Gain the next level as a sorcery to add its
	// ability.)"). Reminder text carries no game meaning, so downstream stages
	// consume this typed flag instead of re-reading the reminder wording.
	ClassReminder bool `json:",omitempty"`
	// ClassLevelGain is the target level of a Class enchantment's level-up
	// activated ability ("{cost}: Level N"), or 0 when this ability is not a
	// level-up. The level number is a typed semantic value the parser reads from
	// the ability body so downstream stages need not re-read the wording.
	ClassLevelGain int `json:",omitempty"`
	// LevelUpCost is the mana cost of a leveler card's "Level up {cost}" ability
	// (CR 711), set only when LevelUpRecognized is true. The leveler activated
	// ability puts a level counter on the source at sorcery speed; the cost is a
	// typed semantic value the parser reads so downstream stages need not re-read
	// the wording.
	LevelUpCost cost.Mana `json:",omitempty"`
	// LevelUpRecognized reports that this ability is a leveler card's intrinsic
	// "Level up {cost}" activated ability line.
	LevelUpRecognized bool `json:",omitempty"`
	// LevelBand carries a leveler card's "LEVEL lo-hi" / "LEVEL lo+" band header
	// with its printed base power/toughness. It is nil for non-band abilities.
	// The abilities printed below a band header belong to that band until the
	// next band header; downstream stages gate them by level-counter count.
	LevelBand *LevelBand `json:",omitempty"`
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

// ExactSequenceKind identifies an exact multi-instruction Oracle sequence.
type ExactSequenceKind uint8

// Exact sequence kinds enumerate the recognized multi-instruction Oracle
// sequences. ExactSequenceUnknown is the zero value for an unrecognized body.
const (
	ExactSequenceUnknown ExactSequenceKind = iota
	ExactSequenceChosenTypeLibraryTopToHand
	ExactSequenceBottomHandThenDraw
	ExactSequenceDiscardHandThenDraw
	ExactSequenceConditionalLookAtTopReveal
)

// ExactSequenceSyntax records an exact sequence and its resolving-body span.
// Bottom and DrawOffset are only meaningful for ExactSequenceBottomHandThenDraw:
// Bottom selects the library end the hand cards move to, and DrawOffset is the
// fixed number added to the "draw that many cards" count ("plus one" => 1).
// LookAtTopCardTypes is only meaningful for
// ExactSequenceConditionalLookAtTopReveal: it lists the card types whose
// disjunction satisfies the "If it's a <type> card" gate before revealing.
type ExactSequenceSyntax struct {
	Kind               ExactSequenceKind
	Span               shared.Span
	Bottom             bool
	DrawOffset         int
	LookAtTopCardTypes []CardType
}

// SourceAbilityCostReductionSyntax is the typed syntax for a source-local
// activated-ability cost reduction.
type SourceAbilityCostReductionSyntax struct {
	Span           shared.Span
	Amount         int
	CountSelection SelectionSyntax
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
	ActivationRestrictionPlayerTurn    ActivationRestrictionKind = "ActivationRestrictionPlayerTurn"
	ActivationRestrictionInstantTiming ActivationRestrictionKind = "ActivationRestrictionInstantTiming"
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

// ActivationPlayerTurnRestriction is a composable typed restriction limiting
// activation to a player's own turn (for example, "Activate only during your
// turn").
type ActivationPlayerTurnRestriction struct {
	Span   shared.Span           `json:"-"`
	Player TriggerPlayerSelector `json:",omitzero"`
}

// ActivationRestriction is source-spanned typed syntax for one trailing
// "Activate only" sentence. Unsupported preserves a restriction sentence that
// has recognized framing but unavailable or ambiguous inner grammar.
type ActivationRestriction struct {
	Kind        ActivationRestrictionKind       `json:",omitempty"`
	Span        shared.Span                     `json:"-"`
	SorcerySpan shared.Span                     `json:"-"`
	Frequency   ActivationFrequencyRestriction  `json:",omitzero"`
	PhaseStep   ActivationPhaseStepRestriction  `json:",omitzero"`
	PlayerTurn  ActivationPlayerTurnRestriction `json:",omitzero"`
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
	TriggerEventKindTokenCreated     TriggerEventKind = "TriggerEventKindTokenCreated"
	// TriggerEventKindDied marks the dies constituent of an event-union trigger
	// such as "enters or dies". It is only used as a union secondary
	// (TriggerEventClause.UnionKind); a standalone dies trigger is a zone-change
	// clause with ZoneChange.Kind == TriggerEventZoneChangeDied.
	TriggerEventKindDied TriggerEventKind = "TriggerEventKindDied"
	// TriggerEventKindAttacksUnblocked marks "this creature attacks and isn't
	// blocked" (CR 509.1h). Unlike a bare attack clause it fires after the
	// declare-blockers step, so it compiles to its own runtime event.
	TriggerEventKindAttacksUnblocked TriggerEventKind = "TriggerEventKindAttacksUnblocked"
	// TriggerEventKindClassBecameLevel marks "When this Class becomes level N"
	// (CR 716), a self-source trigger on a Class enchantment reaching a new
	// level. The target level is carried by TriggerEventClause.ClassBecameLevel.
	TriggerEventKindClassBecameLevel TriggerEventKind = "TriggerEventKindClassBecameLevel"
	// TriggerEventKindDoorUnlocked marks "When you unlock this door" (CR 715),
	// the self-source trigger on a Room enchantment half that fires as that
	// door becomes unlocked. Its subject is the ability's own source.
	TriggerEventKindDoorUnlocked TriggerEventKind = "TriggerEventKindDoorUnlocked"
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

// TriggerCastTurnRelation restricts a spell-cast trigger to the caster's own
// turn or to a turn that isn't theirs ("during your turn", "during an
// opponent's turn").
type TriggerCastTurnRelation string

// Spell-cast turn relations recognized by the syntax parser.
const (
	TriggerCastTurnRelationNone        TriggerCastTurnRelation = ""
	TriggerCastTurnRelationYourTurn    TriggerCastTurnRelation = "TriggerCastTurnRelationYourTurn"
	TriggerCastTurnRelationNotYourTurn TriggerCastTurnRelation = "TriggerCastTurnRelationNotYourTurn"
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
	Span            shared.Span      `json:"-"`
	MatchFromZone   bool             `json:",omitempty"`
	FromZone        TriggerEventZone `json:",omitzero"`
	MatchToZone     bool             `json:",omitempty"`
	ToZone          TriggerEventZone `json:",omitzero"`
	ExcludeToZone   bool             `json:",omitempty"`
	ExcludeFromZone bool             `json:",omitempty"`
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
	TriggerEventCounterLore             TriggerEventCounterKind = "TriggerEventCounterLore"
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
	ManaValueAtMost  int               `json:",omitempty"`
	MatchManaValue   bool              `json:",omitempty"`
	FromZone         TriggerEventZone  `json:",omitzero"`
	// Ordinal records a per-turn spell-cast position from "your Nth spell each
	// turn" wording (1 for first, 2 for second, ...). Zero means no ordinal
	// qualifier. Recognized only with the controller-scoped "you cast" actor.
	Ordinal int `json:",omitempty"`
	// SubtypeFromEntryChoice records the trailing "of the chosen type"
	// restriction ("Whenever you cast a creature spell of the chosen type"),
	// requiring the cast spell to share the creature subtype the source
	// permanent chose as it entered. It lowers to the runtime
	// Selection.SubtypeFromSourceEntryChoice predicate.
	SubtypeFromEntryChoice bool `json:",omitempty"`
	// CastNotFromHand records the trailing "from anywhere other than their hand"
	// (or "your hand") cast-provenance restriction ("Whenever an opponent casts
	// a spell from anywhere other than their hand"). It fires only for spells
	// cast from a zone other than the caster's hand and lowers to the runtime
	// ExcludeFromZone filter against the hand. Unlike FromZone, it is recognized
	// for every caster actor, not only the controller-scoped "you".
	CastNotFromHand bool `json:",omitempty"`
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
	// AttackWhileSaddled marks an attacker-declared clause restricted to combats
	// where the attacking source is saddled ("attacks while saddled", saddle
	// CR 702.166).
	AttackWhileSaddled bool `json:",omitempty"`
	// AttackerCountAtLeast restricts a controller-scoped attack clause to combats
	// where the controller attacks with at least this many creatures ("attack
	// with two or more creatures"). Zero imposes no minimum.
	AttackerCountAtLeast int `json:",omitempty"`
	// MatchCopy is set on a spell-cast clause whose "cast or copy" wording also
	// matches spell copies (CR 707, magecraft).
	MatchCopy bool `json:",omitempty"`
	// TappedForMana restricts a becomes-tapped clause to taps that paid the cost
	// of a mana ability ("is tapped for mana"), CR 106.11a / 605.
	TappedForMana bool `json:",omitempty"`
	// TappedForManaColor narrows a TappedForMana clause to taps that produced a
	// specific type of mana, e.g. "tap a permanent for {C}" restricts to taps
	// that added colorless mana. It is empty for the unrestricted "for mana"
	// wording, which matches a tap that produced any type.
	TappedForManaColor mana.Color `json:"-"`
	// UnionKind names a second trigger event family whose constituent event
	// joins Kind under a shared subject and actor, expressing "Whenever you
	// create or sacrifice a token" (CR 603.2). The trigger fires when either the
	// Kind event or the UnionKind event occurs. It is empty for single-event
	// clauses.
	UnionKind TriggerEventKind `json:",omitempty"`
	// SpellTargetsSource is set on a spell-cast clause whose "that targets this
	// creature" / "that targets <source name>" wording restricts the trigger to
	// spells that target the source permanent (CR 603.2e, the Heroic ability
	// word). It is empty for unrestricted spell-cast clauses.
	SpellTargetsSource bool `json:",omitempty"`
	// SpellTargetSelection restricts a spell-cast clause to spells that target a
	// permanent matching this selection ("...that targets a creature you
	// control" / "...a creature an opponent controls"). It is nil when the clause
	// imposes no such relation. The self-target special case is carried by
	// SpellTargetsSource instead and never co-occurs with this field.
	SpellTargetSelection *TriggerSelection `json:",omitempty"`

	// SpellCastTurnRelation restricts a spell-cast clause to the caster's own
	// turn or to a turn that isn't theirs ("Whenever you cast a spell during
	// your turn" / "during an opponent's turn"). It is empty for spell-cast
	// clauses with no turn restriction.
	SpellCastTurnRelation TriggerCastTurnRelation `json:",omitempty"`

	// ClassBecameLevel carries the target level of a "When this Class becomes
	// level N" clause (TriggerEventKindClassBecameLevel). It is zero for clauses
	// of any other kind.
	ClassBecameLevel int `json:",omitempty"`

	// SelfCast marks a spell-cast clause whose triggering spell is the ability's
	// own source ("When you cast this spell", CR 601.3i). The trigger fires once
	// as the source spell is put on the stack, rather than on every matching
	// spell the controller casts. It is set only on TriggerEventKindSpellCast
	// clauses introduced by "When".
	SelfCast bool `json:",omitempty"`

	// DealtDamageBySourceThisTurn marks a dies clause whose subject is restricted
	// to a permanent that was dealt damage by the ability's own source earlier in
	// the current turn ("Whenever a creature dealt damage by this creature this
	// turn dies", CR 603.2). The relative clause names the ability source via
	// "this creature" or the card's own name; the dying subject itself remains
	// any matching creature. It is set only on TriggerEventKindZoneChange clauses
	// whose ZoneChange.Kind is TriggerEventZoneChangeDied.
	DealtDamageBySourceThisTurn bool `json:",omitempty"`

	// FirstTimeEachTurn marks a became-target clause restricted to the first such
	// targeting each turn ("...for the first time each turn", the Valiant ability
	// word of Bloomburrow and the Glasskite spirits). It caps the ability to one
	// trigger per turn, modeled downstream as a once-per-turn trigger frequency.
	// It is set only on TriggerEventKindBecameTarget clauses.
	FirstTimeEachTurn bool `json:",omitempty"`
	// FirstTimeEachTurnSpan is the source span of the recognized "for the first
	// time each turn" ordinal qualifier. It is zero when FirstTimeEachTurn is
	// false.
	FirstTimeEachTurnSpan shared.Span `json:"-"`
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
	// MinCount is the minimum number of matching events that must have occurred
	// in the window for the condition to hold (e.g. "you attacked with two or
	// more creatures this turn" requires two attacker-declared events). A zero
	// value means a single matching event suffices.
	MinCount int `json:",omitempty"`
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
	// Next marks a one-shot "next" occurrence ("your next upkeep", "the next end
	// step") rather than a recurring phase/step trigger. A spell that resolves
	// sets up such a clause as a delayed triggered ability (CR 603.7).
	Next bool `json:",omitempty"`
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
	PlayerEventActionSearchLibrary  PlayerEventActionKind = "PlayerEventActionSearchLibrary"
	PlayerEventActionCommitCrime    PlayerEventActionKind = "PlayerEventActionCommitCrime"
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

	// PlayerEventCardThis is the self-referential card object "this card",
	// naming the ability's own source ("When you cycle this card", CR 702.29e).
	PlayerEventCardThis PlayerEventCardKind = "PlayerEventCardThis"
)

// PlayerEventCard is a source-spanned player-event card-object modifier.
type PlayerEventCard struct {
	Kind PlayerEventCardKind `json:",omitempty"`
	Span shared.Span         `json:"-"`

	// RequiredTypes and ExcludedTypes record a card-type filter on the event's
	// card object, such as "a creature card" or "a noncreature, nonland card".
	RequiredTypes []TriggerCardType `json:",omitempty"`
	ExcludedTypes []TriggerCardType `json:",omitempty"`

	// RequiredTypesAny and RequiredSubtypesAny record a disjunctive union on the
	// event's card object: a card matching any one of the listed card types
	// ("an artifact or creature card") or subtypes ("an Island, Pirate, or
	// Vehicle card"). The two union dimensions are mutually exclusive because the
	// runtime selection conjoins them, so a mixed type/subtype union fails
	// closed in the parser.
	RequiredTypesAny    []TriggerCardType `json:",omitempty"`
	RequiredSubtypesAny []TriggerSubtype  `json:",omitempty"`
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

	// PlayerEventOccurrenceExceptFirstInDrawStep matches every qualifying draw
	// except the first card a player draws during each of their draw steps
	// ("except the first one they draw in each of their draw steps").
	PlayerEventOccurrenceExceptFirstInDrawStep PlayerEventOccurrenceKind = "PlayerEventOccurrenceExceptFirstInDrawStep"
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
	// PaymentPrelude is an exact standalone event-player payment sentence whose
	// following sentence carries its failure-gated consequence.
	PaymentPrelude *EffectPaymentSyntax `json:",omitempty"`
	// RegenerationRider reports that this sentence is a credited regeneration
	// rider ("It/They can't be regenerated.") folded onto a preceding destroy
	// effect. Reference and coverage scans treat its pronoun and tokens as
	// belonging to that destroy rather than as an unrecognized sibling.
	RegenerationRider bool `json:",omitempty"`
	// TokenCopyGrantRider reports that this sentence is a credited "[That token/
	// It] gains <keyword>." rider folded onto a preceding create-copy-token
	// effect. Reference and coverage scans treat its tokens as belonging to that
	// create effect rather than as an unrecognized sibling.
	TokenCopyGrantRider bool `json:",omitempty"`
	// ReturnAsEnchantmentRider reports that this sentence is a credited "It's an
	// enchantment." rider folded onto a preceding return-to-battlefield effect
	// (the Enduring enchantment-creature cycle). Reference and coverage scans
	// treat its pronoun and tokens as belonging to that return rather than as an
	// unrecognized sibling.
	ReturnAsEnchantmentRider bool `json:",omitempty"`
	// CopyChooseNewTargetsRider reports that this sentence is a credited "You may
	// choose new targets for the copy[ies]." rider folded onto a preceding
	// copy-stack-object effect. Reference and coverage scans treat its tokens as
	// belonging to that copy effect rather than as an unrecognized sibling.
	CopyChooseNewTargetsRider bool `json:",omitempty"`
	// PlayFromTopPayLifeRider reports that this sentence is a credited "If you
	// cast a spell this way, pay life equal to its mana value rather than pay its
	// mana cost." rider folded onto a preceding play-from-library-top grant.
	// Reference and coverage scans treat its tokens as belonging to that grant
	// rather than as an unrecognized sibling.
	PlayFromTopPayLifeRider bool `json:",omitempty"`
	// PileSplitRider reports that this sentence is the credited zero-effect
	// middle sentence of a recognized pile-split sequence ("An opponent separates
	// those cards into two piles." / "An opponent chooses one of those piles.").
	// Reference and coverage scans treat its tokens as belonging to the folded
	// pile-split rather than as an unrecognized sibling.
	PileSplitRider bool `json:",omitempty"`
	// RemoveAuraRider reports that this sentence is the credited inert "This
	// effect doesn't remove this Aura." clarification folded onto a preceding
	// protection (or other continuous) keyword grant on an Aura. The clause only
	// overrides paper Magic's state-based Aura falloff, which this engine never
	// performs, so reference and coverage scans treat its tokens as belonging to
	// the grant rather than as an unrecognized sibling.
	RemoveAuraRider bool `json:",omitempty"`
}

// StaticRuleSubjectKind identifies the source object constrained by a simple
// static rule.
type StaticRuleSubjectKind string

// Simple static-rule subjects.
const (
	StaticRuleSubjectUnknown         StaticRuleSubjectKind = ""
	StaticRuleSubjectSourceCreature  StaticRuleSubjectKind = "StaticRuleSubjectSourceCreature"
	StaticRuleSubjectSourcePermanent StaticRuleSubjectKind = "StaticRuleSubjectSourcePermanent"
	StaticRuleSubjectSourceSpell     StaticRuleSubjectKind = "StaticRuleSubjectSourceSpell"
	StaticRuleSubjectAttachedObject  StaticRuleSubjectKind = "StaticRuleSubjectAttachedObject"
	// StaticRuleSubjectAttachedPermanent scopes a static rule to the noncreature
	// permanent an Aura is attached to ("Enchanted permanent doesn't untap during
	// its controller's untap step.", Ice Over; "Enchanted artifact doesn't untap
	// ...", Inertia Bubble). Unlike StaticRuleSubjectAttachedObject (the attached
	// creature, which also carries combat restrictions) this subject is valid only
	// for the untap prohibition, keeping the wider "enchanted permanent/artifact"
	// nouns from admitting combat rules that name no real card. The compiler maps
	// it to the same attached-object affected group.
	StaticRuleSubjectAttachedPermanent StaticRuleSubjectKind = "StaticRuleSubjectAttachedPermanent"
	// StaticRuleSubjectControlledCreatures scopes a static rule to the creatures
	// the source's controller controls ("Creatures you control can't be
	// blocked."). It is the only group-scoped rule subject; the compiler maps it
	// to the controller-permanents affected group.
	StaticRuleSubjectControlledCreatures StaticRuleSubjectKind = "StaticRuleSubjectControlledCreatures"
	// StaticRuleSubjectBattlefieldCreatures scopes a static rule to every creature
	// on the battlefield, optionally narrowed by a source-relative power filter
	// ("Creatures with power less than this creature's power can't block ..."). The
	// compiler maps it to the battlefield affected group; it backs the conditional
	// "can't block" restrictions whose restricted blockers are any creatures, not
	// only the controller's.
	StaticRuleSubjectBattlefieldCreatures StaticRuleSubjectKind = "StaticRuleSubjectBattlefieldCreatures"
	// StaticRuleSubjectOpponentControlledCreatures scopes a static rule to the
	// creatures the source's controller's opponents control ("Creatures your
	// opponents control attack each combat if able."). The compiler maps it to a
	// battlefield affected group whose affected-permanent Selection scopes the
	// controller to the opponent relation.
	StaticRuleSubjectOpponentControlledCreatures StaticRuleSubjectKind = "StaticRuleSubjectOpponentControlledCreatures"
)

// StaticRuleBlockedObjectKind identifies the protected object an active "can't
// block" restriction shields ("can't block it", "can't block creatures you
// control"). It is the blocked-relationship scope of a block prohibition and is
// unused for every other operation.
type StaticRuleBlockedObjectKind string

// Static-rule blocked-object scopes.
const (
	StaticRuleBlockedObjectNone   StaticRuleBlockedObjectKind = ""
	StaticRuleBlockedObjectSource StaticRuleBlockedObjectKind = "StaticRuleBlockedObjectSource"
	// StaticRuleBlockedObjectControlledCreatures shields the source controller's
	// creatures ("can't block creatures you control").
	StaticRuleBlockedObjectControlledCreatures StaticRuleBlockedObjectKind = "StaticRuleBlockedObjectControlledCreatures"
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
	// StaticRuleOperationTransform constrains transforming the subject ("...
	// can't transform").
	StaticRuleOperationTransform StaticRuleOperationKind = "StaticRuleOperationTransform"
	// StaticRuleOperationBlockAndBeBlocked combines the active "block" and
	// passive "be blocked" prohibitions printed as a single sentence ("can't
	// block and can't be blocked"); it lowers to both block-domain rule effects.
	StaticRuleOperationBlockAndBeBlocked StaticRuleOperationKind = "StaticRuleOperationBlockAndBeBlocked"
	// StaticRuleOperationBlockedByAll is the true-lure requirement printed as
	// "All creatures able to block <subject> do so." Every creature able to block
	// the subject attacker must do so (CR 509.1c). It always pairs with a
	// requirement constraint and passive voice.
	StaticRuleOperationBlockedByAll StaticRuleOperationKind = "StaticRuleOperationBlockedByAll"
	// StaticRuleOperationAssignDamageAsUnblocked is the permission printed as "You
	// may have <subject> assign its combat damage as though it weren't blocked."
	// The subject attacker may deal its combat damage to its attack target rather
	// than to its blockers. It always pairs with a requirement constraint and
	// passive voice.
	StaticRuleOperationAssignDamageAsUnblocked StaticRuleOperationKind = "StaticRuleOperationAssignDamageAsUnblocked"
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
	// StaticRuleQualifierDefenderYou restricts an attack prohibition to the
	// controller and their planeswalkers ("can't attack you or planeswalkers
	// you control").
	StaticRuleQualifierDefenderYou StaticRuleQualifierKind = "StaticRuleQualifierDefenderYou"
	// StaticRuleQualifierByMoreThanOne bounds a "can't be blocked" prohibition
	// to the exceptional case "by more than one creature".
	StaticRuleQualifierByMoreThanOne StaticRuleQualifierKind = "StaticRuleQualifierByMoreThanOne"
	// StaticRuleQualifierBlockerFlying restricts a "can't be blocked" prohibition
	// to blockers with flying ("can't be blocked by creatures with flying").
	StaticRuleQualifierBlockerFlying StaticRuleQualifierKind = "StaticRuleQualifierBlockerFlying"
	// StaticRuleQualifierBlockerPowerOrLess restricts a "can't be blocked"
	// prohibition to blockers whose power is at most the qualifier's Amount
	// ("can't be blocked by creatures with power N or less").
	StaticRuleQualifierBlockerPowerOrLess StaticRuleQualifierKind = "StaticRuleQualifierBlockerPowerOrLess"
	// StaticRuleQualifierBlockerPowerOrGreater restricts a "can't be blocked"
	// prohibition to blockers whose power is at least the qualifier's Amount
	// ("can't be blocked by creatures with power N or greater").
	StaticRuleQualifierBlockerPowerOrGreater StaticRuleQualifierKind = "StaticRuleQualifierBlockerPowerOrGreater"
	// StaticRuleQualifierBlockerColor restricts a "can't be blocked" prohibition
	// to blockers of a single color ("can't be blocked by white creatures"); the
	// color travels in StaticRuleQualifier.Color.
	StaticRuleQualifierBlockerColor StaticRuleQualifierKind = "StaticRuleQualifierBlockerColor"
	// StaticRuleQualifierBlockerArtifact restricts a "can't be blocked"
	// prohibition to artifact-creature blockers ("can't be blocked by artifact
	// creatures").
	StaticRuleQualifierBlockerArtifact StaticRuleQualifierKind = "StaticRuleQualifierBlockerArtifact"
	// StaticRuleQualifierBlockedAttackerFlying restricts a "can block only"
	// permission to attackers with flying ("This creature can block only
	// creatures with flying."): the subject blocker may block only attackers that
	// have flying. Unlike the StaticRuleQualifierBlocker* kinds the matched
	// characteristic describes the attacker the subject blocks, not the subject's
	// own blockers.
	StaticRuleQualifierBlockedAttackerFlying StaticRuleQualifierKind = "StaticRuleQualifierBlockedAttackerFlying"
	// StaticRuleQualifierAlone restricts an active attack, block, or attack-or-
	// block prohibition to the "alone" case ("can't attack alone", "can't block
	// alone", "can't attack or block alone"): the subject can't be the only
	// creature attacking or blocking.
	StaticRuleQualifierAlone StaticRuleQualifierKind = "StaticRuleQualifierAlone"
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
	// Amount carries the power threshold for the blocker power-comparison
	// qualifiers (StaticRuleQualifierBlockerPowerOrLess/OrGreater); it is zero
	// and unused for all other qualifier kinds.
	Amount int `json:",omitempty"`
	// Color carries the stopped blocker color for StaticRuleQualifierBlockerColor;
	// it is the unknown color and unused for all other qualifier kinds.
	Color Color `json:",omitempty"`
}

// StaticRuleSyntax is a composable typed simple static-rule declaration.
// Sentence text and tokens remain available only as source metadata.
type StaticRuleSyntax struct {
	Span       shared.Span           `json:"-"`
	Subject    StaticRuleSubject     `json:",omitzero"`
	Constraint StaticRuleConstraint  `json:",omitzero"`
	Operation  StaticRuleOperation   `json:",omitzero"`
	Qualifiers []StaticRuleQualifier `json:",omitempty"`
	// Guarded marks a static rule that carries a trailing condition clause
	// ("unless you control seven or more lands.") gating the rule. When true, the
	// rule applies only while the separately parsed condition holds; the clause
	// itself is recognized by the condition machinery, not the static-rule
	// parser. False means the rule is unconditional.
	Guarded bool `json:",omitempty"`
	// Order is the rule's dense source-order rank (of Span), used downstream to
	// order static-rule effects without byte offsets.
	Order shared.SourceOrder `json:"-"`
	// BlockedObject names the protected object an active "can't block" restriction
	// shields ("Creatures with power less than this creature's power can't block
	// it.", "... can't block creatures you control."). The empty value means the
	// block prohibition is unconditional ("Creatures can't block."); it is unused
	// for every non-block operation.
	BlockedObject StaticRuleBlockedObjectKind `json:",omitempty"`
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
	// Spree marks a Spree modal (CR 702.171): a "Spree" keyword header whose
	// options are "+ {cost} — effect" lines, each with its own additional mana
	// cost. The controller chooses one or more options and pays each chosen
	// option's cost. It is recognized directly from the Spree header, so its
	// choice range is set without consulting the choose-header vocabulary.
	Spree bool `json:",omitempty"`
	// Escalate marks an Escalate modal (CR 702.121): a "Escalate <cost>" keyword
	// header printed above an ordinary choose-one-or-more modal whose controller
	// pays EscalateCost once for each mode chosen beyond the first. Unlike Spree,
	// every option shares the single escalate cost rather than carrying its own,
	// so the cost lives on the modal rather than on each Mode.
	Escalate bool `json:",omitempty"`
	// EscalateCost is the additional mana cost paid for each mode chosen beyond
	// the first on an Escalate modal. It is set only when Escalate is true.
	EscalateCost cost.Mana `json:",omitempty"`
	// EscalateSpan covers the recognized "Escalate <cost>" keyword header so
	// coverage and rendering can credit its source tokens.
	EscalateSpan shared.Span `json:"-"`
	// MinModes and MaxModes are the recognized choice range of the choose
	// header (e.g. "Choose two —" yields 2/2 and "Choose one or both —" yields
	// 1/2). They are populated only when ChoiceKnown is true; downstream code
	// must consume these typed fields instead of re-reading header tokens.
	MinModes    int                    `json:",omitempty"`
	MaxModes    int                    `json:",omitempty"`
	ChoiceKnown bool                   `json:",omitempty"`
	ChoiceKind  ModalChoiceKind        `json:",omitempty"`
	ChoiceBonus ModalChoiceBonusSyntax `json:",omitzero"`
}

// DiceTable is a recognized die-roll outcome table: a "Roll a d<N>." line
// followed by result rows that each map an inclusive interval of the rolled
// value to a resolving effect.
type DiceTable struct {
	// DieSides is the number of faces of the rolled die (the N in "d<N>"). An
	// open-ended row ("15+ |") uses DieSides as its inclusive upper bound.
	DieSides int            `json:",omitempty"`
	Rows     []DiceTableRow `json:",omitempty"`
}

// DiceTableRow is one outcome row of a DiceTable: an inclusive result interval
// [Min, Max] and the resolving sentences that apply when the roll lands in it.
type DiceTableRow struct {
	Span      shared.Span    `json:"-"`
	Text      string         `json:",omitempty"`
	Tokens    []shared.Token `json:"-"`
	Min       int            `json:",omitempty"`
	Max       int            `json:",omitempty"`
	Sentences []Sentence     `json:",omitempty"`
	Atoms     Atoms          `json:",omitzero"`
}

// CoinFlip is a recognized "Flip a coin." outcome and its win/lose branches.
// The flip resolves to a fair two-sided random draw (CR 705); a win branch
// applies on heads and a lose branch applies on tails. At least one branch is
// present. Each branch holds its own freshly parsed resolving sentences, parsed
// from the clause after "If you win the flip," / "If you lose the flip," so the
// condition wording carries no residual effect tokens.
type CoinFlip struct {
	// Win holds the resolving sentences of the "If you win the flip, ..." branch,
	// or nil when the ability has no win branch.
	Win []Sentence `json:",omitempty"`
	// Lose holds the resolving sentences of the "If you lose the flip, ..."
	// branch, or nil when the ability has no lose branch.
	Lose []Sentence `json:",omitempty"`
	// Spans are the source spans of every sentence the coin-flip recognizer
	// consumed (the flip sentence and each branch sentence). Coverage credits
	// them whole because the recognizer fully accounts for their wording.
	Spans []shared.Span `json:"-"`
	// ConstructSpan covers the whole coin-flip construct, from the "Flip a coin."
	// sentence through the last branch sentence. The compiler stamps it on every
	// branch effect so the position-blind backend's body-span machinery covers
	// the entire construct without reasoning about source offsets itself.
	ConstructSpan shared.Span `json:"-"`
}

// VoteClause is the recognized "Starting with you, each player votes for <A> or
// <B>." voting construct and its majority-gated arms (CR 701.32). Options holds
// the two printed choice labels in printed order; Arms holds one entry per "If
// <option> gets more votes[ or the vote is tied], <effect>." consequence, each
// naming the option index it depends on and whether a tie also satisfies it.
type VoteClause struct {
	// Options are the two named choice labels in printed order.
	Options []string `json:",omitempty"`
	// Arms are the majority-gated consequences, in printed order.
	Arms []VoteArm `json:",omitempty"`
	// Spans are the source spans of every sentence the recognizer consumed (the
	// voting sentence and each arm sentence). Coverage credits them whole.
	Spans []shared.Span `json:"-"`
	// ConstructSpan covers the whole construct, from the voting sentence through
	// the last arm sentence. The compiler stamps it on every arm effect so the
	// position-blind backend's body-span machinery covers the entire construct.
	ConstructSpan shared.Span `json:"-"`
}

// VoteArm is one majority-gated consequence of a VoteClause. Option is the index
// into VoteClause.Options whose vote count the arm depends on; TieInclusive
// reports whether a tied vote also satisfies the arm ("or the vote is tied").
type VoteArm struct {
	Option       int        `json:",omitempty"`
	TieInclusive bool       `json:",omitempty"`
	Sentences    []Sentence `json:",omitempty"`
}

// ModalChoiceKind identifies exact modal header vocabulary whose range alone
// is not sufficient to preserve fail-closed lowering.
type ModalChoiceKind string

const (
	// ModalChoiceKindUnknown marks modal headers without special typed vocabulary.
	ModalChoiceKindUnknown ModalChoiceKind = ""
	// ModalChoiceKindOneOrMore marks the exact "choose one or more" header.
	ModalChoiceKindOneOrMore ModalChoiceKind = "ModalChoiceKindOneOrMore"
	// ModalChoiceKindOneAtRandom marks the exact "choose one at random" header,
	// where the single mode is selected at random rather than by the controller
	// (CR 700.2). Its range stays one/one; the kind preserves that the choice is
	// not made by a player so downstream lowering can fail closed where random
	// selection is unsupported.
	ModalChoiceKindOneAtRandom ModalChoiceKind = "ModalChoiceKindOneAtRandom"
)

// ModalChoiceBonusCondition identifies a cast-time condition that expands a
// modal choice range.
type ModalChoiceBonusCondition string

const (
	// ModalChoiceBonusConditionNone marks a modal header without a bonus.
	ModalChoiceBonusConditionNone ModalChoiceBonusCondition = ""
	// ModalChoiceBonusConditionControlsCommander requires controlling a commander.
	ModalChoiceBonusConditionControlsCommander ModalChoiceBonusCondition = "ModalChoiceBonusConditionControlsCommander"
)

// ModalChoiceBonusSyntax is a typed conditional expansion of a modal choice.
type ModalChoiceBonusSyntax struct {
	Condition          ModalChoiceBonusCondition `json:",omitempty"`
	AdditionalMaxModes int                       `json:",omitempty"`
}

// ModeLabelKind identifies an exact supported label printed before a modal
// option's rules text.
type ModeLabelKind string

const (
	// ModeLabelUnknown marks an unlabeled or unsupported mode label.
	ModeLabelUnknown ModeLabelKind = ""
	// ModeLabelSellContraband marks the exact "Sell Contraband" label.
	ModeLabelSellContraband ModeLabelKind = "ModeLabelSellContraband"
	// ModeLabelBuyInformation marks the exact "Buy Information" label.
	ModeLabelBuyInformation ModeLabelKind = "ModeLabelBuyInformation"
	// ModeLabelHireMercenary marks the exact "Hire a Mercenary" label.
	ModeLabelHireMercenary ModeLabelKind = "ModeLabelHireMercenary"
)

// ModeLabelClause is a recognized modal option label and its separating em dash.
type ModeLabelClause struct {
	Kind          ModeLabelKind `json:",omitempty"`
	Text          string        `json:",omitempty"`
	Span          shared.Span   `json:"-"`
	SeparatorSpan shared.Span   `json:"-"`
}

// Mode is one bullet option in a modal ability.
type Mode struct {
	Span      shared.Span      `json:"-"`
	Text      string           `json:",omitempty"`
	Tokens    []shared.Token   `json:"-"`
	Label     *ModeLabelClause `json:",omitempty"`
	SpreeCost *SpreeCostClause `json:",omitempty"`
	// FlavorSpan covers a Final Fantasy "Summon:" Saga modal option's flavor-name
	// prefix ("Combine Powers! — Put three +1/+1 counters..."), a Title-Case
	// proper name set off by an em dash. The name carries no rules meaning (CR
	// 207.2c) and is stripped from the option body, so it lowers as if absent;
	// the span is retained only so coverage credits the flavor tokens.
	FlavorSpan             shared.Span             `json:"-"`
	FlavorSeparatorSpan    shared.Span             `json:"-"`
	Body                   Phrase                  `json:",omitzero"`
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
