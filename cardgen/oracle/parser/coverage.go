package parser

import (
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// CoverageBlocker names the ability component that left an ability
// parser-incomplete. A blocker is recorded for the work queue so parser grammar
// work can be ranked by the component family that fails closed most often.
type CoverageBlocker string

// Coverage blocker families.
const (
	CoverageBlockerEffect    CoverageBlocker = "effect"
	CoverageBlockerTrigger   CoverageBlocker = "trigger"
	CoverageBlockerCost      CoverageBlocker = "cost"
	CoverageBlockerCondition CoverageBlocker = "condition"
	CoverageBlockerStatic    CoverageBlocker = "static"
	CoverageBlockerModal     CoverageBlocker = "modal"
	CoverageBlockerOther     CoverageBlocker = "other"
)

// UncoveredSpan is one contiguous run of must-cover tokens that no recognized
// parser semantic span accounted for. Text is the rendered run, used to cluster
// the corpus's unrepresented grammar into a ranked work queue.
type UncoveredSpan struct {
	Span shared.Span
	Text string
}

// UncoveredComponent is the grammatical component (a trigger clause, condition,
// cost, effect clause, or modal header) that owns one or more uncovered runs. It
// is the work-queue unit: recognized atoms embedded inside an unrepresented
// clause (a self-reference, a keyword) can fragment the raw token runs, so the
// queue clusters by the owning component's full text instead, yielding
// actionable Oracle constructs rather than stray connectives.
type UncoveredComponent struct {
	Span    shared.Span
	Text    string
	Blocker CoverageBlocker
}

// AbilityCoverageReport is the parser-only round-trip coverage of one ability:
// whether every must-cover token span is accounted for by a recognized typed
// element, the uncovered runs that block completeness, the blocking component
// families, and the resolving/exact effect tallies.
type AbilityCoverageReport struct {
	Complete         bool
	ResolvingEffects int
	ExactEffects     int
	Uncovered        []UncoveredSpan
	Components       []UncoveredComponent
	Blockers         []CoverageBlocker
}

// DocumentCoverageReport aggregates a face's ability coverage.
type DocumentCoverageReport struct {
	Complete         bool
	ResolvingEffects int
	ExactEffects     int
	Abilities        []AbilityCoverageReport
	Uncovered        []UncoveredSpan
	Components       []UncoveredComponent
}

// DocumentCoverage classifies every ability of a parsed face. The face is
// parser-complete when every ability is parser-complete.
func DocumentCoverage(doc Document) DocumentCoverageReport {
	report := DocumentCoverageReport{Complete: true}
	for i := range doc.Abilities {
		ability := AbilityCoverage(&doc.Abilities[i])
		report.ResolvingEffects += ability.ResolvingEffects
		report.ExactEffects += ability.ExactEffects
		if !ability.Complete {
			report.Complete = false
		}
		report.Uncovered = append(report.Uncovered, ability.Uncovered...)
		report.Components = append(report.Components, ability.Components...)
		report.Abilities = append(report.Abilities, ability)
	}
	return report
}

// AbilityCoverage classifies one ability. An ability is parser-complete when
// every must-cover token span (Ability.CoverageSpans) is covered by a span built
// only from recognized parser typed output, every condition introducer it owns
// resolves to a recognized clause, and — for modal abilities — its choice header
// and every option are recognized.
func AbilityCoverage(a *Ability) AbilityCoverageReport {
	if a.Modal != nil {
		return modalAbilityCoverage(a)
	}
	return flatAbilityCoverage(a)
}

func flatAbilityCoverage(a *Ability) AbilityCoverageReport {
	report := AbilityCoverageReport{Complete: true}
	recognized := abilityRecognizedSpans(a)

	report.Uncovered = uncoveredRuns(a.CoverageSpans(), a.Tokens, recognized)
	report.ResolvingEffects, report.ExactEffects = effectTallies(a.Sentences, report.Uncovered)

	if len(report.Uncovered) > 0 {
		report.Complete = false
		report.Blockers = appendBlockers(report.Blockers, uncoveredBlockers(a, report.Uncovered))
	}
	report.Components = abilityUncoveredComponents(a, report.Uncovered)

	for i := range a.ConditionSegments {
		if conditionSegmentRecognized(&a.ConditionSegments[i], a.ConditionClauses, a.EventHistoryConditions, a.Sentences) {
			continue
		}
		report.Complete = false
		report.Blockers = appendBlocker(report.Blockers, CoverageBlockerCondition)
		report.Components = addComponent(report.Components, a.Tokens, a.ConditionSegments[i].Span,
			CoverageBlockerCondition, UncoveredSpan{Span: a.ConditionSegments[i].Span})
	}
	return report
}

func modalAbilityCoverage(a *Ability) AbilityCoverageReport {
	report := AbilityCoverageReport{Complete: true}
	if !a.Modal.ChoiceKnown {
		report.Complete = false
		report.Blockers = appendBlocker(report.Blockers, CoverageBlockerModal)
		report.Components = append(report.Components, UncoveredComponent{
			Span:    a.Modal.header.Span,
			Text:    a.Modal.header.Text,
			Blocker: CoverageBlockerModal,
		})
	}
	for i := range a.Modal.Options {
		mode := modeCoverage(&a.Modal.Options[i])
		report.ResolvingEffects += mode.ResolvingEffects
		report.ExactEffects += mode.ExactEffects
		report.Uncovered = append(report.Uncovered, mode.Uncovered...)
		report.Components = append(report.Components, mode.Components...)
		if !mode.Complete {
			report.Complete = false
			report.Blockers = appendBlockers(report.Blockers, mode.Blockers)
		}
	}
	return report
}

func modeCoverage(m *Mode) AbilityCoverageReport {
	report := AbilityCoverageReport{Complete: true}
	recognized := modeRecognizedSpans(m)

	report.Uncovered = uncoveredRuns(m.CoverageSpans(), m.Tokens, recognized)
	report.ResolvingEffects, report.ExactEffects = effectTallies(m.Sentences, report.Uncovered)

	if len(report.Uncovered) > 0 {
		report.Complete = false
		report.Blockers = appendBlocker(report.Blockers, CoverageBlockerEffect)
	}
	report.Components = modeUncoveredComponents(m, report.Uncovered)

	for i := range m.ConditionSegments {
		if conditionSegmentRecognized(&m.ConditionSegments[i], m.ConditionClauses, m.EventHistoryConditions, m.Sentences) {
			continue
		}
		report.Complete = false
		report.Blockers = appendBlocker(report.Blockers, CoverageBlockerCondition)
		report.Components = addComponent(report.Components, m.Tokens, m.ConditionSegments[i].Span,
			CoverageBlockerCondition, UncoveredSpan{Span: m.ConditionSegments[i].Span})
	}
	return report
}

// abilityRecognizedSpans builds the union of source spans an ability's recognized
// typed output accounts for. The set mirrors the spans the lowering coverage
// consumers assert against, but is reconstructed only from parser data: no
// compiler or lowering output participates.
func abilityRecognizedSpans(a *Ability) []shared.Span {
	var spans []shared.Span
	add := func(span shared.Span) {
		if span != (shared.Span{}) {
			spans = append(spans, span)
		}
	}

	if a.AbilityWord != nil {
		add(a.AbilityWord.Span)
		add(a.AbilityWord.SeparatorSpan)
	}
	add(a.BodySeparatorSpan)
	if a.Optional {
		add(a.OptionalSpan)
	}
	if a.Trigger != nil && triggerRecognized(a.Trigger) {
		spans = appendTriggerSpans(spans, a.Trigger)
	}
	if a.CostSyntax != nil && costRecognized(a.CostSyntax) {
		add(a.CostSyntax.Span)
	}
	if a.Kind == AbilitySpellAdditionalCost && a.CostSyntax != nil && costRecognized(a.CostSyntax) {
		add(a.Span)
	}
	if a.Kind == AbilitySpellAlternativeCost && a.AlternativeCost != nil {
		add(a.AlternativeCost.Span)
	}
	if len(a.Chapters) > 0 {
		add(a.ChapterSpan)
	}
	for i := range a.ActivationRestrictions {
		if activationRestrictionRecognized(a.ActivationRestrictions[i].Kind) {
			add(a.ActivationRestrictions[i].Span)
		}
	}
	if a.TriggerFrequency != nil && a.TriggerFrequency.Kind != TriggerFrequencyUnknown {
		add(a.TriggerFrequency.Span)
	}
	for i := range a.StaticDeclarations {
		spans = appendStaticDeclarationSpans(spans, &a.StaticDeclarations[i])
	}
	for i := range a.ConditionSegments {
		segment := &a.ConditionSegments[i]
		if conditionSegmentRecognized(segment, a.ConditionClauses, a.EventHistoryConditions, a.Sentences) {
			add(segment.Span)
			add(segment.ActivationKeyword)
			spans = appendConditionClauseSpan(spans, segment, a.ConditionClauses)
		}
	}
	spans = appendCommonRecognizedSpans(spans, a.SemanticKeywords, a.SemanticReferences, a.Reminders, a.Quoted)
	spans = appendEffectSpans(spans, a.Sentences)
	spans = appendConstructRecognizedSpans(spans, a)
	return spans
}

// appendTriggerSpans credits a recognized trigger clause. The clause Span can be
// clipped at the first comma of a coordinated event list ("Whenever you cast an
// instant, sorcery, or Wizard spell"), so the typed event clause's own span and
// the rendered event span are added too: each is recognized grammar the parser
// fully typed, and together they cover the whole event phrase the clause Span
// alone may stop short of.
func appendTriggerSpans(spans []shared.Span, trigger *TriggerClause) []shared.Span {
	for _, span := range []shared.Span{trigger.Span, trigger.EventSpan} {
		if span != (shared.Span{}) {
			spans = append(spans, span)
		}
	}
	if trigger.TriggerEvent != nil && trigger.TriggerEvent.Span != (shared.Span{}) {
		spans = append(spans, trigger.TriggerEvent.Span)
	}
	if trigger.TriggerEvent != nil {
		for _, span := range []shared.Span{trigger.TriggerEvent.SpellSelection.Span, trigger.TriggerEvent.DamageSourceSpellSelection.Span} {
			if span != (shared.Span{}) {
				spans = append(spans, span)
			}
		}
	}
	if trigger.PhaseStep != nil && trigger.PhaseStep.Span != (shared.Span{}) {
		spans = append(spans, trigger.PhaseStep.Span)
	}
	if trigger.PlayerEvent != nil && trigger.PlayerEvent.Span != (shared.Span{}) {
		spans = append(spans, trigger.PlayerEvent.Span)
	}
	return spans
}

// appendConditionClauseSpan credits the typed condition clause a recognized
// segment links to. The pre-segmented ConditionSegment.Span is split on commas,
// so it can stop at the first item of a coordinated subject list ("If you control
// a Fish, Octopus, ... or Whale"); the typed clause spans the whole predicate.
func appendConditionClauseSpan(spans []shared.Span, segment *ConditionSegment, clauses []ConditionClause) []shared.Span {
	if segment.ClauseIndex < 0 || segment.ClauseIndex >= len(clauses) {
		return spans
	}
	if span := clauses[segment.ClauseIndex].Span; span != (shared.Span{}) {
		spans = append(spans, span)
	}
	return spans
}

func modeRecognizedSpans(m *Mode) []shared.Span {
	var spans []shared.Span
	for i := range m.ConditionSegments {
		segment := &m.ConditionSegments[i]
		if conditionSegmentRecognized(segment, m.ConditionClauses, m.EventHistoryConditions, m.Sentences) {
			if segment.Span != (shared.Span{}) {
				spans = append(spans, segment.Span)
			}
			if segment.ActivationKeyword != (shared.Span{}) {
				spans = append(spans, segment.ActivationKeyword)
			}
		}
	}
	spans = appendCommonRecognizedSpans(spans, m.SemanticKeywords, m.SemanticReferences, m.Reminders, m.Quoted)
	spans = appendEffectSpans(spans, m.Sentences)
	return spans
}

func appendCommonRecognizedSpans(
	spans []shared.Span,
	keywords []Keyword,
	references []Reference,
	reminders, quoted []Delimited,
) []shared.Span {
	for i := range keywords {
		spans = append(spans, keywords[i].Span)
	}
	for i := range references {
		spans = append(spans, references[i].Span)
	}
	for i := range reminders {
		spans = append(spans, reminders[i].Span)
	}
	for i := range quoted {
		spans = append(spans, quoted[i].Span)
	}
	return spans
}

// appendEffectSpans adds the source spans of a body's recognized resolving
// instructions. A recognized static-rule sentence contributes its whole span; a
// resolving sentence contributes each accounted-for effect's clause and connector
// span.
//
// An exact effect fully consumes its clause: the parser round-tripped every token
// (including compound objects, mana lists, and keyword lists joined by internal
// "and"/","/"or"), so its whole ownership clause is credited. A typed-but-inexact
// effect may have absorbed an adjacent clause that produced no effect (a predicate
// the parser could not type), because resolvingClauseStart/End extend an effect's
// ownership to the sentence edges; such an effect only credits its tight clause,
// clipped to the nearest top-level boundary around its verb, so an unrepresented
// sibling clause leaves its own tokens uncovered and fails closed.
func appendEffectSpans(spans []shared.Span, sentences []Sentence) []shared.Span {
	for i := range sentences {
		sentence := &sentences[i]
		if sentence.StaticRule != nil {
			spans = append(spans, sentence.Span)
			continue
		}
		if sentence.PaymentPrelude != nil {
			spans = append(spans, sentence.Span)
			continue
		}
		tokens := semanticEffectTokens(sentence.Tokens)
		if len(sentence.Effects) > 0 {
			spans = appendSentenceConnectorSpan(spans, tokens)
		}
		for j := range sentence.Effects {
			effect := &sentence.Effects[j]
			if !effectRepresented(effect) {
				continue
			}
			spans = append(spans, effectCreditSpan(tokens, effect, effectExact(effect)))
			if effect.ConnectionSpan != (shared.Span{}) {
				spans = append(spans, effect.ConnectionSpan)
			}
		}
	}
	return spans
}

// appendSentenceConnectorSpan credits a leading "then" that sequences a resolving
// sentence after the preceding one ("... Then this creature deals damage ..."). It
// is a recognized structural connector, like an in-sentence "then"/"and" between
// effects, but it falls before the first effect's subject so effectCreditSpan does
// not reach it; without crediting it the sentence would fail closed on the lone
// connector. Only a leading "then" is credited, so an unrelated leading word still
// leaves its tokens uncovered.
func appendSentenceConnectorSpan(spans []shared.Span, tokens []shared.Token) []shared.Span {
	if len(tokens) > 0 && equalWord(tokens[0], "then") {
		spans = append(spans, tokens[0].Span)
	}
	return spans
}

// effectCreditSpan returns the source span a recognized effect accounts for. The
// parser's ownership clause (resolvingClauseStart/End) extends to the previous and
// next recognized effect verbs or the sentence edges, so a clause that produced no
// effect verb is absorbed into a neighbor's ownership. The credited span is always
// clipped backward to the top-level boundary before the verb so a leading
// unrepresented clause ("Goad target creature and draw a card") is not absorbed.
// The forward extent is kept whole only when the effect round-tripped exactly:
// an exact effect consumed every token after its verb (a compound object, mana
// list, or keyword list joined by internal "and"/","/"or"), whereas an inexact
// effect may have swept in a trailing unrepresented clause and is clipped to the
// next top-level boundary so that clause stays uncovered. semanticEffectTokens
// drops parenthetical and quoted tokens, so every boundary token is top-level.
func effectCreditSpan(tokens []shared.Token, effect *EffectSyntax, fullForward bool) shared.Span {
	verbIndex, startIndex, endIndex := -1, -1, -1
	for i := range tokens {
		if tokens[i].Span == effect.VerbSpan {
			verbIndex = i
		}
		if startIndex == -1 && tokens[i].Span.Start == effect.ClauseSpan.Start {
			startIndex = i
		}
		if tokens[i].Span.End == effect.ClauseSpan.End {
			endIndex = i + 1
		}
	}
	if verbIndex < 0 || startIndex < 0 || endIndex < 0 || verbIndex < startIndex || verbIndex >= endIndex {
		return effect.ClauseSpan
	}
	clauseStart := startIndex
	for i := verbIndex - 1; i >= startIndex; i-- {
		if effectClauseBoundary(tokens[i]) {
			clauseStart = i + 1
			break
		}
	}
	clauseEnd := endIndex
	if !fullForward {
		for i := verbIndex + 1; i < endIndex; i++ {
			if effectClauseBoundary(tokens[i]) {
				clauseEnd = i
				break
			}
		}
	}
	if clauseStart >= clauseEnd {
		return effect.ClauseSpan
	}
	return shared.SpanOf(tokens[clauseStart:clauseEnd])
}

// effectClauseBoundary reports whether a token separates two top-level resolving
// clauses, matching the boundaries resolvingClauseStart and resolvingClauseEnd use
// to segment effects. A recognized effect credits only the tight clause around its
// own verb, clipped to the nearest such boundary in both directions, so a sibling
// clause that produced no effect ("Goad target creature, then draw a card") leaves
// its tokens uncovered instead of being absorbed into the neighbor effect's span.
func effectClauseBoundary(token shared.Token) bool {
	return token.Kind == shared.Comma || token.Kind == shared.Semicolon ||
		equalWord(token, "then") || equalWord(token, "and")
}

// effectRepresented reports whether the parser assigned the effect a recognized
// type. A typed effect is one the downstream compiler can attempt to lower, so a
// typed effect contributes its clause to the recognized-span union; an
// EffectUnknown clause contributes nothing and fails closed as unrepresented
// grammar. This mirrors lowering, whose coverage credits every compiled effect's
// sentence span regardless of the parser's stricter Exact round-trip flag.
func effectRepresented(effect *EffectSyntax) bool {
	return effect.Kind != EffectUnknown
}

// effectExact reports whether the parser fully round-tripped the effect. Most
// effects record this in Exact; add-mana bodies record it in Mana.LegacyBodyExact,
// mirroring legacyExactManaBody. It drives the exact-effect headline metric only,
// not completeness.
func effectExact(effect *EffectSyntax) bool {
	return effect.Exact || effect.Mana.LegacyBodyExact
}

func appendStaticDeclarationSpans(spans []shared.Span, declaration *StaticDeclarationSyntax) []shared.Span {
	if declaration.Span != (shared.Span{}) {
		spans = append(spans, declaration.Span)
	}
	if declaration.Subject.Span != (shared.Span{}) {
		spans = append(spans, declaration.Subject.Span)
	}
	if declaration.OperationSpan != (shared.Span{}) {
		spans = append(spans, declaration.OperationSpan)
	}
	if declaration.HasCondition && declaration.ConditionSpan != (shared.Span{}) {
		spans = append(spans, declaration.ConditionSpan)
	}
	spans = append(spans, declaration.KeywordSpans...)
	return spans
}

// conditionSegmentRecognized reports whether a condition segment resolves to a
// recognized clause: a linked typed ConditionClause, a linked EventHistory
// condition, or an "unless ... pays" payment carried by a resolving effect (the
// reconciliation the compiler performs in applyEffectPaymentsToConditions).
func conditionSegmentRecognized(
	segment *ConditionSegment,
	clauses []ConditionClause,
	histories []EventHistoryCondition,
	sentences []Sentence,
) bool {
	if segment.ClauseIndex >= 0 && segment.ClauseIndex < len(clauses) {
		return true
	}
	if segment.EventHistoryIndex >= 0 && segment.EventHistoryIndex < len(histories) {
		return true
	}
	return segmentCoveredByEffectPayment(segment, sentences)
}

func segmentCoveredByEffectPayment(segment *ConditionSegment, sentences []Sentence) bool {
	for i := range sentences {
		for j := range sentences[i].Effects {
			payment := &sentences[i].Effects[j].Payment
			if payment.Payer != EffectPaymentPayerTargetController || len(payment.ManaCost) == 0 {
				continue
			}
			if segment.Order.Contains(payment.Order) {
				return true
			}
		}
	}
	return false
}

func triggerRecognized(trigger *TriggerClause) bool {
	return trigger.TriggerEvent != nil || trigger.PhaseStep != nil || trigger.PlayerEvent != nil
}

func costRecognized(cost *Cost) bool {
	if len(cost.Components) == 0 {
		return false
	}
	for i := range cost.Components {
		if cost.Components[i].Kind == CostComponentUnknown {
			return false
		}
	}
	return true
}

func activationRestrictionRecognized(kind ActivationRestrictionKind) bool {
	return kind != ActivationRestrictionUnknown && kind != ActivationRestrictionUnsupported
}

// effectTallies counts the resolving effects and how many round-trip exactly. An
// effect counts as exact only when its sentence has no uncovered must-cover token:
// a sentence with an unrepresented clause does not round-trip as a whole, so no
// effect drawn from it can be claimed as an exact reconstruction even if the
// parser flagged that individual clause exact.
func effectTallies(sentences []Sentence, uncovered []UncoveredSpan) (resolving, exact int) {
	for i := range sentences {
		if sentences[i].StaticRule != nil {
			continue
		}
		sentenceExact := !spanHasUncovered(sentences[i].Span, uncovered)
		for j := range sentences[i].Effects {
			resolving++
			if sentenceExact && effectExact(&sentences[i].Effects[j]) {
				exact++
			}
		}
	}
	return resolving, exact
}

// spanHasUncovered reports whether any uncovered run falls within the given span.
func spanHasUncovered(span shared.Span, uncovered []UncoveredSpan) bool {
	for i := range uncovered {
		if spanCovers(span, uncovered[i].Span) {
			return true
		}
	}
	return false
}

// uncoveredRuns returns the contiguous runs of must-cover tokens that no
// recognized span accounts for. Adjacent uncovered tokens are clustered into one
// run so the work queue groups whole unrepresented phrases.
func uncoveredRuns(coverage []shared.Span, tokens []shared.Token, recognized []shared.Span) []UncoveredSpan {
	mustCover := make(map[shared.Span]bool, len(coverage))
	for _, span := range coverage {
		mustCover[span] = true
	}
	var runs []UncoveredSpan
	var run []shared.Token
	flush := func() {
		if len(run) == 0 {
			return
		}
		runs = append(runs, UncoveredSpan{
			Span: shared.SpanOf(run),
			Text: joinTokens(run),
		})
		run = nil
	}
	for i := range tokens {
		token := tokens[i]
		if !mustCover[token.Span] {
			flush()
			continue
		}
		if spanInUnion(token.Span, recognized) {
			flush()
			continue
		}
		run = append(run, token)
	}
	flush()
	return runs
}

func spanInUnion(span shared.Span, union []shared.Span) bool {
	for _, candidate := range union {
		if spanCovers(candidate, span) {
			return true
		}
	}
	return false
}

// uncoveredBlockers classifies which component families left tokens uncovered, so
// the report can rank parser work by component. The classification is heuristic
// reporting metadata only; completeness itself is decided by the coverage gate.
func uncoveredBlockers(a *Ability, uncovered []UncoveredSpan) []CoverageBlocker {
	var blockers []CoverageBlocker
	for i := range uncovered {
		span := uncovered[i].Span
		switch {
		case a.Trigger != nil && spanCovers(a.Trigger.Span, span):
			blockers = appendBlocker(blockers, CoverageBlockerTrigger)
		case a.CostSyntax != nil && spanCovers(a.CostSyntax.Span, span):
			blockers = appendBlocker(blockers, CoverageBlockerCost)
		case spanInSentences(span, a.Sentences):
			blockers = appendBlocker(blockers, CoverageBlockerEffect)
		default:
			blockers = appendBlocker(blockers, CoverageBlockerOther)
		}
	}
	return blockers
}

func spanInSentences(span shared.Span, sentences []Sentence) bool {
	for i := range sentences {
		if spanCovers(sentences[i].Span, span) {
			return true
		}
	}
	return false
}

func appendBlocker(blockers []CoverageBlocker, blocker CoverageBlocker) []CoverageBlocker {
	if slices.Contains(blockers, blocker) {
		return blockers
	}
	return append(blockers, blocker)
}

func appendBlockers(blockers, more []CoverageBlocker) []CoverageBlocker {
	for _, blocker := range more {
		blockers = appendBlocker(blockers, blocker)
	}
	return blockers
}

// abilityUncoveredComponents maps each uncovered run to the grammatical component
// that owns it (trigger, cost, condition, or effect clause) and returns one
// deduplicated work-queue entry per owning component. Mapping to the owning
// component reunites runs that recognized embedded atoms fragmented, so the queue
// reads as whole Oracle constructs.
func abilityUncoveredComponents(a *Ability, runs []UncoveredSpan) []UncoveredComponent {
	var components []UncoveredComponent
	for i := range runs {
		if a.Kind == AbilitySpellAdditionalCost || a.Kind == AbilitySpellAlternativeCost {
			components = addComponent(components, a.Tokens, a.Span, CoverageBlockerCost, runs[i])
			continue
		}
		span, blocker := ownerForRun(runs[i].Span, a.Trigger, a.CostSyntax, a.ConditionSegments, a.Sentences)
		components = addComponent(components, a.Tokens, span, blocker, runs[i])
	}
	return components
}

func modeUncoveredComponents(m *Mode, runs []UncoveredSpan) []UncoveredComponent {
	var components []UncoveredComponent
	for i := range runs {
		span, blocker := ownerForRun(runs[i].Span, nil, nil, m.ConditionSegments, m.Sentences)
		components = addComponent(components, m.Tokens, span, blocker, runs[i])
	}
	return components
}

func ownerForRun(
	run shared.Span,
	trigger *TriggerClause,
	cost *Cost,
	segments []ConditionSegment,
	sentences []Sentence,
) (shared.Span, CoverageBlocker) {
	if trigger != nil && spanCovers(trigger.Span, run) {
		return trigger.Span, CoverageBlockerTrigger
	}
	if cost != nil && spanCovers(cost.Span, run) {
		return cost.Span, CoverageBlockerCost
	}
	for i := range segments {
		if spanCovers(segments[i].Span, run) {
			return segments[i].Span, CoverageBlockerCondition
		}
	}
	for i := range sentences {
		sentence := &sentences[i]
		if !spanCovers(sentence.Span, run) {
			continue
		}
		for j := range sentence.Effects {
			effect := &sentence.Effects[j]
			if !effectRepresented(effect) && spanCovers(effect.ClauseSpan, run) {
				return effect.ClauseSpan, CoverageBlockerEffect
			}
		}
		return sentence.Span, CoverageBlockerEffect
	}
	return run, CoverageBlockerOther
}

func addComponent(
	components []UncoveredComponent,
	tokens []shared.Token,
	span shared.Span,
	blocker CoverageBlocker,
	run UncoveredSpan,
) []UncoveredComponent {
	for i := range components {
		if components[i].Span == span {
			return components
		}
	}
	text := joinTokens(TokensInSpan(tokens, span))
	if text == "" {
		text = run.Text
		span = run.Span
	}
	if text == "" {
		return components
	}
	return append(components, UncoveredComponent{Span: span, Text: text, Blocker: blocker})
}
