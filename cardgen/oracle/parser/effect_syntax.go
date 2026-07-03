package parser

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func emitResolvingSyntax(abilities []Ability) {
	for i := range abilities {
		if recognizeChosenTypeLibraryTopSequence(&abilities[i]) {
			continue
		}
		if recognizeConditionalLookAtTopSequence(&abilities[i]) {
			continue
		}
		if recognizeConditionalLookAtTopBattlefieldSequence(&abilities[i]) {
			continue
		}
		if recognizeBottomHandThenDrawSequence(&abilities[i]) {
			continue
		}
		if recognizeDiscardHandThenDrawSequence(&abilities[i]) {
			continue
		}
		if recognizeDrawThenDiscardUnlessSequence(&abilities[i]) {
			continue
		}
		emitSentenceResolvingSyntax(
			abilities[i].Sentences,
			abilities[i].Atoms,
			abilities[i].ActivationRestrictions,
			abilities[i].TriggerFrequency,
			abilities[i].SourceAbilityCostReduction,
			abilities[i].interveningConditionStrip(),
		)
		fuseDiscardThenDrawSentences(abilities[i].Sentences)
		annotateSacrificeThenCountSentences(abilities[i].Sentences)
		attachTokenGrantedAbilities(&abilities[i])
		attachGainGrantedAbilities(&abilities[i])
		attachEmblemEffects(&abilities[i])
		recognizeControllerOptionalPaymentSequence(&abilities[i])
		recognizeOptionalManaPaymentBenefitSequence(&abilities[i])
		recognizeEventPlayerOptionalPaymentSequence(&abilities[i])
		recognizeEventPlayerOptionalPaymentAffirmativeSequence(&abilities[i])
		recognizeDefendingPlayerOptionalPaymentSequence(&abilities[i])
		recognizeEventPlayerPerCreatureUntapPayment(&abilities[i])
		recognizeControllerMandatoryPaymentSequence(&abilities[i])
		recognizeEventPlayerOptionalActionGate(&abilities[i])
		recognizeNonControllerMayHaveActionGate(&abilities[i])
		recognizeGroupMayHaveActionGate(&abilities[i])
		recognizeResolvingCopyPaymentGate(&abilities[i])
		recognizePayRepeatedlyAnimateSequence(&abilities[i])
		foldAnimateSelfStillSentence(&abilities[i])
		foldAnimateTargetStillSentence(&abilities[i])
		if abilities[i].DiceTable != nil {
			for k := range abilities[i].DiceTable.Rows {
				row := &abilities[i].DiceTable.Rows[k]
				emitSentenceResolvingSyntax(row.Sentences, row.Atoms, nil, nil, nil, interveningConditionStrip{})
				fuseDiscardThenDrawSentences(row.Sentences)
				annotateSacrificeThenCountSentences(row.Sentences)
			}
		}
		if abilities[i].Modal == nil {
			continue
		}
		for j := range abilities[i].Modal.Options {
			mode := &abilities[i].Modal.Options[j]
			emitSentenceResolvingSyntax(mode.Sentences, mode.Atoms, nil, nil, nil, interveningConditionStrip{})
			fuseDiscardThenDrawSentences(mode.Sentences)
			annotateSacrificeThenCountSentences(mode.Sentences)
			if sentencesHaveImpulseExile(mode.Sentences) {
				mode.SemanticReferences = nil
				mode.ConditionBoundaries = nil
				mode.EventHistoryConditions = nil
				mode.ConditionClauses = nil
				mode.ConditionSegments = nil
			}
		}
	}
}

func sentencesHaveImpulseExile(sentences []Sentence) bool {
	for len(sentences) > 2 && isReminderSentence(sentences[len(sentences)-1]) {
		sentences = sentences[:len(sentences)-1]
	}
	return len(sentences) == 2 &&
		len(sentences[0].Effects) == 1 &&
		sentences[0].Effects[0].Kind == EffectImpulseExile
}

// attachTokenGrantedAbilities binds each quoted ability captured on the ability
// ("... creature token with \"When this token dies, you gain 1 life.\"") to the
// create-token effect whose clause contains it, parsing the quoted body through
// the same pipeline so downstream layers lower it from the typed inner document.
// It re-evaluates the affected effect's exactness so the reconstructed "... token
// with" rider is byte-checked. A quoted body that fails to parse, or that no
// single create clause contains, is left unattached so the create fails closed.
func attachTokenGrantedAbilities(ability *Ability) {
	if len(ability.Quoted) == 0 {
		return
	}
	for q := range ability.Quoted {
		quoted := ability.Quoted[q]
		effect := createEffectContainingSpan(ability, quoted.Span)
		if effect == nil || effect.TokenGrantedAbility != nil {
			continue
		}
		granted, ok := parseStaticGrantedAbility(quoted)
		if !ok {
			continue
		}
		// A quoted triggered ability ("When this token dies, ...") or activated
		// ability, including a mana ability ("{T}: Add {C}", "Sacrifice this
		// token: Add {C}"), attaches to the created token; the lowerer compiles
		// the inner body and appends the resulting triggered, activated, or mana
		// ability to the token definition. A quoted static ability ("This token
		// can't block.") attaches the same way and the lowerer appends the
		// resulting static ability to the token definition. The create
		// reconstructs only when the rider is an ability the downstream layers
		// handle; a rider whose static body the lowerer cannot represent fails
		// closed there rather than here.
		if len(granted.document.Abilities) != 1 ||
			(granted.document.Abilities[0].Kind != AbilityTriggered &&
				granted.document.Abilities[0].Kind != AbilityActivated &&
				granted.document.Abilities[0].Kind != AbilityStatic) {
			continue
		}
		stored := granted
		effect.TokenGrantedAbility = &stored
		effect.Exact = exactEffectSyntax(effect)
	}
}

// attachGainGrantedAbilities binds each quoted ability captured on the ability
// ("This creature gains \"Whenever this creature deals combat damage to a
// player, that player loses the game.\"") to the gain effect whose clause
// contains it, parsing the quoted body through the same pipeline so downstream
// layers lower it from the typed inner document. It re-evaluates the affected
// effect's exactness so the reconstructed "gains \"...\"" rider is byte-checked.
// A quoted body that fails to parse, or that no single gain clause contains, is
// left unattached so the grant fails closed.
func attachGainGrantedAbilities(ability *Ability) {
	if len(ability.Quoted) == 0 {
		return
	}
	for q := range ability.Quoted {
		quoted := ability.Quoted[q]
		effect := gainEffectContainingSpan(ability, quoted.Span)
		if effect == nil || effect.GainGrantedAbility != nil {
			continue
		}
		granted, ok := parseStaticGrantedAbility(quoted)
		if !ok {
			continue
		}
		// Only a quoted triggered ability ("Whenever this creature deals combat
		// damage to a player, ...") attaches. Static, activated, and mana granted
		// abilities stay fail-closed pending dedicated lowering support, so the
		// grant reconstructs only when the rider is a triggered ability the
		// downstream layers handle.
		if len(granted.document.Abilities) != 1 ||
			granted.document.Abilities[0].Kind != AbilityTriggered {
			continue
		}
		stored := granted
		effect.GainGrantedAbility = &stored
		effect.Exact = exactEffectSyntax(effect)
	}
}

// attachEmblemEffects reclassifies an "You get an emblem with \"...\"" effect,
// which the sentence pipeline first models as a generic clause, into an
// EffectCreateEmblem carrying each quoted ability parsed through the same
// pipeline so downstream layers lower the emblem from the typed inner
// documents. Every quoted ability contained in the emblem clause must parse, or
// the clause is left unchanged so the emblem fails closed.
func attachEmblemEffects(ability *Ability) {
	if len(ability.Quoted) == 0 {
		return
	}
	effect := emblemEffectClause(ability)
	if effect == nil {
		return
	}
	var grants []StaticGrantedAbilitySyntax
	for q := range ability.Quoted {
		quoted := ability.Quoted[q]
		if quoted.Span.Start.Offset < effect.ClauseSpan.Start.Offset ||
			quoted.Span.End.Offset > effect.Span.End.Offset {
			continue
		}
		granted, ok := parseStaticGrantedAbility(quoted)
		if !ok {
			return
		}
		grants = append(grants, granted)
	}
	if len(grants) == 0 {
		return
	}
	effect.Kind = EffectCreateEmblem
	effect.EmblemAbilities = grants
	effect.Exact = exactEffectSyntax(effect)
}

// emblemEffectClause returns the lone effect whose clause begins "You get an
// emblem with". It returns nil when no clause or more than one clause begins
// that way, so an ambiguous emblem binding fails closed.
func emblemEffectClause(ability *Ability) *EffectSyntax {
	var match *EffectSyntax
	for i := range ability.Sentences {
		for j := range ability.Sentences[i].Effects {
			effect := &ability.Sentences[i].Effects[j]
			if _, ok := cutTokenPrefix(effect.Tokens, "you", "get", "an", "emblem", "with"); !ok {
				continue
			}
			if match != nil {
				return nil
			}
			match = effect
		}
	}
	return match
}

// span. It returns nil when no gain clause contains span, or when more than one
// might, so an ambiguous granted-ability binding fails closed.
func gainEffectContainingSpan(ability *Ability, span shared.Span) *EffectSyntax {
	var match *EffectSyntax
	for i := range ability.Sentences {
		for j := range ability.Sentences[i].Effects {
			effect := &ability.Sentences[i].Effects[j]
			if effect.Kind != EffectGain ||
				span.Start.Offset < effect.ClauseSpan.Start.Offset ||
				span.End.Offset > effect.Span.End.Offset {
				continue
			}
			if match != nil {
				return nil
			}
			match = effect
		}
	}
	return match
}

// createEffectContainingSpan returns the lone create-token effect whose clause
// contains span. It returns nil when no create clause contains span, or when
// more than one might, so an ambiguous granted-ability binding fails closed.
func createEffectContainingSpan(ability *Ability, span shared.Span) *EffectSyntax {
	var match *EffectSyntax
	for i := range ability.Sentences {
		for j := range ability.Sentences[i].Effects {
			effect := &ability.Sentences[i].Effects[j]
			if effect.Kind != EffectCreate ||
				span.Start.Offset < effect.ClauseSpan.Start.Offset ||
				span.End.Offset > effect.Span.End.Offset {
				continue
			}
			if match != nil {
				return nil
			}
			match = effect
		}
	}
	return match
}

// interveningConditionStrip removes a triggered ability's intervening-if clause
// from the token stream that feeds the body effect and target parse. The
// intervening "if" gates whether the trigger's effect happens; its own tokens,
// references, and pronouns ("if at least four mana was spent to cast it, ...")
// belong to the condition, not the effect, and are recognized separately as the
// ability's condition. Leaving them in the effect parse would seed a spurious
// leading effect from a condition verb ("cast"/"spent") and leak the condition's
// pronoun into the effect's reference set, defeating the controller-subject and
// single-reference effect recognizers. The strip is keyed to the intervening
// boundary position, so non-trigger leading conditions and replacement effects
// ("If X would happen, ... instead") are untouched.
type interveningConditionStrip struct {
	start shared.Position
	set   bool
}

// strip drops the intervening-if clause, including its trailing comma, when the
// supplied effect tokens begin at the recognized intervening boundary. Tokens
// that do not begin there are returned unchanged.
func (s interveningConditionStrip) strip(tokens []shared.Token) []shared.Token {
	if !s.set || len(tokens) == 0 || tokens[0].Span.Start.Offset != s.start.Offset {
		return tokens
	}
	end := conditionClauseEnd(tokens, 0)
	if end >= len(tokens) || tokens[end].Kind != shared.Comma {
		return tokens
	}
	return tokens[end+1:]
}

// interveningConditionStrip returns the strip for this ability's triggered
// intervening-if condition, or the zero strip when the ability has none.
func (a *Ability) interveningConditionStrip() interveningConditionStrip {
	for _, boundary := range a.ConditionBoundaries {
		if boundary.Intervening {
			return interveningConditionStrip{start: boundary.Start, set: true}
		}
	}
	return interveningConditionStrip{}
}

func emitSentenceResolvingSyntax(
	sentences []Sentence,
	atoms Atoms,
	restrictions []ActivationRestriction,
	triggerFrequency *TriggerFrequencyRestriction,
	sourceCostReduction *SourceAbilityCostReductionSyntax,
	intervening interveningConditionStrip,
) {
	if recognizeImpulseExileSequence(sentences) {
		return
	}
	if recognizeRepeatProcessSequence(sentences, atoms) {
		return
	}
	legacyEffects := 0
	currentEffects := 0
	unrecognizedSibling := false
	var riderCandidates []int
	var chooseColorCandidates []int
	var enchantmentReturnCandidates []int
	var pileSplitMiddleCandidates []int
	for i := range sentences {
		if sentences[i].StaticRule != nil ||
			sourceCostReduction != nil && sentences[i].Span == sourceCostReduction.Span ||
			spanInsideActivationRestriction(sentences[i].Span, restrictions) ||
			spanInsideTriggerFrequency(sentences[i].Span, triggerFrequency) {
			continue
		}
		tokens := stripLeadingConditionClause(
			intervening.strip(semanticEffectTokens(sentences[i].Tokens)), atoms)
		count := orderedEffectCount(tokens, atoms)
		legacyEffects += count
		sentences[i].LegacyEffects = count > 0
		sentences[i].Targets = parseTargets(tokens, atoms)
		sentences[i].Effects = parseEffects(sentences[i], tokens, atoms)
		recognizeTargetOpponentHandManaSentence(&sentences[i])
		recognizeLookAtTargetPlayerHandSentence(&sentences[i])
		reconcileRetargetSentenceTargets(&sentences[i])
		collapseManaSpendRiderSentence(&sentences[i], tokens)
		currentEffects += len(sentences[i].Effects)
		if len(tokens) > 0 && len(sentences[i].Effects) == 0 &&
			len(atoms.KeywordsWithin(tokens)) == 0 && count == 0 &&
			!effectWordsAt(tokens, 0, "activate", "only", "if") {
			switch {
			case isRegenerationRiderTokens(tokens) || isThisWayRegenerationRiderTokens(tokens):
				riderCandidates = append(riderCandidates, i)
			case isChosenColorChooseTokens(tokens):
				chooseColorCandidates = append(chooseColorCandidates, i)
			case isEnchantmentReturnRiderTokens(tokens):
				enchantmentReturnCandidates = append(enchantmentReturnCandidates, i)
			case isPileSplitMiddleTokens(tokens):
				pileSplitMiddleCandidates = append(pileSplitMiddleCandidates, i)
			case isChooseTargetPreambleTokens(tokens) && len(sentences[i].Targets) > 0:
				// A bare "Choose [another] target <object>." preamble declares a
				// target consumed by a following effect's "it" pronoun and emits
				// no standalone effect, so it is a benign sibling, not an
				// unrecognized one.
			default:
				unrecognizedSibling = true
			}
		}
	}
	recognizeShuffleRevealPermanentSequence(sentences)
	recognizeRevealUntilThenPutSequence(sentences)
	recognizeRevealTopPartitionSequence(sentences)
	recognizeRevealChooseHandDiscardSequence(sentences)
	if len(pileSplitMiddleCandidates) > 0 && !recognizePileSplitSequence(sentences) {
		unrecognizedSibling = true
	}
	creditConjoinedCopyChooseNewTargetsRider(sentences)
	if len(chooseColorCandidates) > 0 && !creditChosenColorChoice(sentences, chooseColorCandidates) {
		unrecognizedSibling = true
	}
	if foldedLegacy, foldedEffects, ok := creditTokenCopyGrantRider(sentences, atoms); ok {
		legacyEffects -= foldedLegacy
		currentEffects -= foldedEffects
	}
	if foldedLegacy, foldedEffects, ok := creditCopyChooseNewTargetsRider(sentences, atoms); ok {
		legacyEffects -= foldedLegacy
		currentEffects -= foldedEffects
	}
	if foldedLegacy, foldedEffects, ok := creditPlayFromTopPayLifeRider(sentences, atoms); ok {
		legacyEffects -= foldedLegacy
		currentEffects -= foldedEffects
	}
	if currentEffects == 1 && unrecognizedSibling {
		for i := range sentences {
			for j := range sentences[i].Effects {
				sentences[i].Effects[j].Exact = false
				sentences[i].Effects[j].HasUnrecognizedSibling = true
			}
		}
	}
	if len(riderCandidates) > 0 {
		creditRegenerationRider(sentences, riderCandidates, unrecognizedSibling)
	}
	if len(enchantmentReturnCandidates) > 0 {
		creditEnchantmentReturnRider(sentences, enchantmentReturnCandidates, unrecognizedSibling)
	}
	if legacyEffects <= 1 {
		return
	}
	for i := range sentences {
		for j := range sentences[i].Effects {
			sentences[i].Effects[j].RequiresOrderedLowering = true
		}
	}
}

// creditRegenerationRider folds one or more "It/They can't be regenerated."
// rider sentences onto the ability's lone destroy effect: it sets
// PreventRegeneration plus a coverage span on the destroy and marks the rider
// sentences so reference and coverage scans credit them. It credits only when
// the ability holds exactly one destroy effect, that destroy is exact, and no
// other sentence is unrecognized; otherwise the riders stay uncredited and the
// card fails closed at the lowering coverage check. Sibling effects other than
// the lone destroy are permitted (for example a "...creates a token" rider that
// destruction spells such as Pongify pair with the regeneration clause), because
// the rider's pronoun subject can only denote the destroyed permanent and the
// rider span is covered independently of those siblings.
func creditRegenerationRider(sentences []Sentence, riderCandidates []int, unrecognizedSibling bool) {
	if unrecognizedSibling {
		return
	}
	destroy := loneDestroyEffect(sentences)
	if destroy == nil || !destroy.Exact {
		return
	}
	riderSpan := sentences[riderCandidates[0]].Span
	for _, index := range riderCandidates[1:] {
		if sentences[index].Span.End.Offset > riderSpan.End.Offset {
			riderSpan.End = sentences[index].Span.End
		}
	}
	destroy.PreventRegeneration = true
	destroy.RegenerationRiderSpan = riderSpan
	for _, index := range riderCandidates {
		sentences[index].RegenerationRider = true
	}
}

// isEnchantmentReturnRiderTokens reports whether the sentence tokens are the
// "It's an enchantment." rider of the Enduring enchantment-creature cycle. The
// parenthetical "(It's not a creature.)" reminder is stripped before sentence
// parsing, so only the bare declaration remains. The rider folds onto a
// preceding return-to-battlefield effect, recording that the returned permanent
// enters as an Enchantment (losing its creature type).
func isEnchantmentReturnRiderTokens(tokens []shared.Token) bool {
	if !effectWordsAt(tokens, 0, "it's", "an", "enchantment") {
		return false
	}
	rest := tokens[3:]
	for i := range rest {
		if rest[i].Kind != shared.Period {
			return false
		}
	}
	return true
}

// creditEnchantmentReturnRider folds one or more "It's an enchantment." rider
// sentences onto the ability's lone return-to-battlefield effect: it sets
// ReturnAsEnchantment plus a coverage span on the return and marks the rider
// sentences so reference and coverage scans credit them. It credits only when
// the ability holds exactly one return-to-battlefield effect and no other
// sentence is unrecognized; otherwise the riders stay uncredited and the card
// fails closed at the lowering coverage check.
func creditEnchantmentReturnRider(sentences []Sentence, riderCandidates []int, unrecognizedSibling bool) {
	if unrecognizedSibling {
		return
	}
	ret := loneReturnToBattlefieldEffect(sentences)
	if ret == nil {
		return
	}
	riderSpan := sentences[riderCandidates[0]].Span
	for _, index := range riderCandidates[1:] {
		if sentences[index].Span.End.Offset > riderSpan.End.Offset {
			riderSpan.End = sentences[index].Span.End
		}
	}
	ret.ReturnAsEnchantment = true
	ret.ReturnAsEnchantmentRiderSpan = riderSpan
	for _, index := range riderCandidates {
		sentences[index].ReturnAsEnchantmentRider = true
	}
}

// loneReturnToBattlefieldEffect returns the single return-to-battlefield effect
// across the sentences, or nil when the sentences hold zero or more than one
// such effect. Sibling effects of other kinds are permitted and ignored.
func loneReturnToBattlefieldEffect(sentences []Sentence) *EffectSyntax {
	var found *EffectSyntax
	for i := range sentences {
		for j := range sentences[i].Effects {
			if sentences[i].Effects[j].Kind != EffectReturn ||
				sentences[i].Effects[j].ToZone != zone.Battlefield {
				continue
			}
			if found != nil {
				return nil
			}
			found = &sentences[i].Effects[j]
		}
	}
	return found
}

// parseCounterPlacementScopedToCount recognizes the counter kind a "put …
// counters …" clause places, scoping detection so a trailing "where X is the
// number of … counters on …" count subject does not introduce a second counter
// atom. parseCounterPlacement fails closed when a clause spans more than one
// counter atom, so for a WhereX amount the detection is limited to the tokens
// before the count subject, leaving the count phrase to the amount's own
// dynamic subject.
func parseCounterPlacementScopedToCount(clause []shared.Token, amount EffectAmountSyntax, atoms Atoms) (counter.Kind, bool) {
	counterClause := clause
	if amount.DynamicForm == EffectDynamicAmountFormWhereX && amount.Span != (shared.Span{}) {
		counterClause = tokensBeforeOffset(clause, amount.Span.Start.Offset)
	}
	counterClause = placedCounterTokens(counterClause)
	return parseCounterPlacement(counterClause, atoms)
}

// entersTappedCounterClause strips the leading "tapped and" connector from a
// combined self enters-tapped-with-counters clause ("enters tapped and with N
// <kind> counters on it.", Sphere of the Suns, Noble's Purse) so the
// with-counters portion is detected as a single counter placement rather than
// failing closed on the "and" that joins the tapped qualifier to it. The bare
// "enters tapped with N counters on it." form (the Vivid land cycle) carries no
// "and" and is unaffected. Only EffectEnterTapped clauses are rewritten; every
// other effect kind keeps its clause so genuine compound placements ("a +1/+1
// counter and a shield counter") still fail closed.
func entersTappedCounterClause(kind EffectKind, clause []shared.Token) []shared.Token {
	if kind != EffectEnterTapped {
		return clause
	}
	if len(clause) >= 2 && equalWord(clause[0], "tapped") && equalWord(clause[1], "and") {
		return clause[2:]
	}
	return clause
}

// parseCounterPlacementChoices recognizes a placed-counter noun phrase that lets
// the resolving controller choose between two or more counter kinds ("a +1/+1
// counter or a loyalty counter", Elspeth Conquers Death chapter III). It returns
// the distinct kinds in source order. The phrase must join its kinds with "or"
// (never "and", which is a compound placement), each kind must be a real,
// permanent-placeable kind, and the phrase must carry no other counter
// alternative collapsing to a single atom. Any other shape returns nil so the
// single-kind path and the fail-closed path stay unchanged.
func parseCounterPlacementChoices(clause []shared.Token, atoms Atoms) []counter.Kind {
	tokens := placedCounterTokens(clause)
	hasOr := false
	for _, token := range tokens {
		if equalWord(token, "and") {
			return nil
		}
		if equalWord(token, "or") {
			hasOr = true
		}
	}
	if !hasOr {
		return nil
	}
	span := shared.SpanOf(tokens)
	var kinds []counter.Kind
	seen := make(map[counter.Kind]bool)
	for _, atom := range atoms.Counters() {
		if !spanCovers(span, atom.Span) {
			continue
		}
		if !atom.Kind.Valid() || atom.Kind == counter.Finality || atom.Kind.PlayerOnly() {
			return nil
		}
		if seen[atom.Kind] {
			return nil
		}
		seen[atom.Kind] = true
		kinds = append(kinds, atom.Kind)
	}
	if len(kinds) < 2 {
		return nil
	}
	return kinds
}

// placedCounterTokens narrows a "Put <amount> <kind> counter(s) on <recipient>"
// clause to the placed-counter noun phrase that precedes the placement
// preposition "on". The recipient that follows may itself name a counter ("each
// creature you control with a +1/+1 counter on it") or join subtypes with "and"
// ("each Pest, Bat, and Spider you control"); either would make
// parseCounterPlacement see a second counter atom or an "and" and fail closed,
// so they are excluded here. The placed-counter portion keeps any "and" before
// the preposition so a genuine compound placement ("a +1/+1 counter and a
// shield counter on …") still fails closed. A clause with no "on" is returned
// unchanged.
func placedCounterTokens(tokens []shared.Token) []shared.Token {
	for i := range tokens {
		if equalWord(tokens[i], "on") {
			return tokens[:i]
		}
	}
	return tokens
}

// isChooseTargetPreambleTokens reports whether the sentence is a bare
// "Choose [another] target <object>." preamble that declares a target consumed
// by a following effect's "it" pronoun ("Choose another target creature. Put a
// number of +1/+1 counters on it …"). The choice itself produces no standalone
// effect, so it is a benign sibling rather than an unrecognized one.
func isChooseTargetPreambleTokens(tokens []shared.Token) bool {
	if len(tokens) == 0 || !equalWord(tokens[0], "choose") {
		return false
	}
	return effectContainsWords(normalizedWords(tokens), "target")
}

// isChosenColorChooseTokens reports whether the sentence tokens are exactly
// "Choose a color" (with optional trailing periods). This bare color-choice
// sentence precedes "Add an amount of mana of that color equal to your devotion
// to that color." (Nykthos, Shrine to Nyx); the choice itself produces no
// standalone effect, so it is folded onto the chosen-color devotion add-mana.
func isChosenColorChooseTokens(tokens []shared.Token) bool {
	if !effectWordsAt(tokens, 0, "choose", "a", "color") {
		return false
	}
	rest := tokens[3:]
	for i := range rest {
		if rest[i].Kind != shared.Period {
			return false
		}
	}
	return true
}

// underOwnersControl reports the battlefield-destination ownership rider "under
// their owners' control" / "under its owner's control", under which each moved
// card enters controlled by its own owner rather than the resolving player. It
// is distinct from the "under your control" rider and the bare form.
func underOwnersControl(tokens []shared.Token) bool {
	words := normalizedWords(tokens)
	if !slices.Contains(words, "under") {
		return false
	}
	return effectContainsWords(words, "owners", "control") ||
		effectContainsWords(words, "owner's", "control")
}

// ability's lone chosen-color add-mana effect by widening that effect's span to
// cover the choice sentence, so the mana ability's coverage scan credits the
// choice. It succeeds only when the ability holds exactly one add-mana effect
// that carries a chosen-color body (devotion or dynamic count) and that effect is
// exact; otherwise it reports failure so the choice stays unrecognized and the
// card fails closed.
func creditChosenColorChoice(sentences []Sentence, chooseCandidates []int) bool {
	manaEffect := loneChosenColorManaEffect(sentences)
	if manaEffect == nil || !manaEffect.Exact {
		return false
	}
	for _, index := range chooseCandidates {
		if sentences[index].Span.Start.Offset < manaEffect.Span.Start.Offset {
			manaEffect.Span.Start = sentences[index].Span.Start
		}
	}
	return true
}

// loneChosenColorManaEffect returns the single chosen-color add-mana effect
// (devotion or dynamic count) across the sentences, or nil when the sentences
// hold zero or more than one such effect.
func loneChosenColorManaEffect(sentences []Sentence) *EffectSyntax {
	var found *EffectSyntax
	for i := range sentences {
		for j := range sentences[i].Effects {
			manaSyntax := sentences[i].Effects[j].Mana
			if sentences[i].Effects[j].Kind != EffectAddMana ||
				!manaSyntax.ChosenColorDevotion && !manaSyntax.ChosenColorDynamic {
				continue
			}
			if found != nil {
				return nil
			}
			found = &sentences[i].Effects[j]
		}
	}
	return found
}

// loneDestroyEffect returns the single EffectDestroy across the sentences, or nil
// when the sentences hold zero or more than one destroy effect. Sibling effects
// of other kinds are permitted and ignored; only the count of destroy effects
// constrains the result so a regeneration rider can fold onto a destruction
// clause that is accompanied by a recognized non-destroy effect.
func loneDestroyEffect(sentences []Sentence) *EffectSyntax {
	var found *EffectSyntax
	for i := range sentences {
		for j := range sentences[i].Effects {
			if sentences[i].Effects[j].Kind != EffectDestroy {
				continue
			}
			if found != nil {
				return nil
			}
			found = &sentences[i].Effects[j]
		}
	}
	return found
}

// creditTokenCopyGrantRider folds a "[That token/It] gains <keyword>." rider
// sentence onto the sentences' lone create-copy-token effect. The created token
// gains the keyword(s); the rider sentence's effects are cleared and the
// sentence is marked so reference and coverage scans credit it to the create.
// It credits only when the sentences hold exactly one exact create-copy-token
// effect and a single matching gain-keyword sentence whose subject denotes the
// created token; otherwise nothing is folded and the card fails closed. It
// returns the folded sentence's legacy-effect and current-effect counts so the
// caller can correct its sequence-length bookkeeping.
func creditTokenCopyGrantRider(sentences []Sentence, atoms Atoms) (foldedLegacy, foldedEffects int, ok bool) {
	create := loneCopyTokenCreateEffect(sentences)
	if create == nil || !create.Exact {
		return 0, 0, false
	}
	for i := range sentences {
		if len(sentences[i].Effects) != 1 || sentences[i].Effects[0].Kind != EffectGain {
			continue
		}
		tokens := semanticEffectTokens(sentences[i].Tokens)
		keywords, match := tokenCopyGrantRiderKeywords(tokens, atoms)
		if !match {
			continue
		}
		create.TokenCopyGrantKeywords = keywords
		create.TokenCopyGrantRiderSpan = sentences[i].Span
		foldedEffects = len(sentences[i].Effects)
		if sentences[i].LegacyEffects {
			foldedLegacy = orderedEffectCount(tokens, atoms)
		}
		sentences[i].Effects = nil
		sentences[i].TokenCopyGrantRider = true
		return foldedLegacy, foldedEffects, true
	}
	return 0, 0, false
}

// creditCopyChooseNewTargetsRider folds the optional "You may choose new targets
// for the copy[ies]." rider sentence onto the ability's lone copy-stack-object
// effect: it sets CopyMayChooseNewTargets plus a coverage span on the copy and
// clears the rider sentence's effects so reference and coverage scans credit it.
// It credits only when the ability holds exactly one copy-stack-object effect,
// that copy is exact, and the rider sentence is exactly the recognized retarget
// clause; otherwise the rider stays uncredited and the card fails closed.
func creditCopyChooseNewTargetsRider(sentences []Sentence, atoms Atoms) (foldedLegacy, foldedEffects int, ok bool) {
	copyEffect := loneCopyStackObjectEffect(sentences)
	if copyEffect == nil || !copyEffect.Exact {
		return 0, 0, false
	}
	for i := range sentences {
		if len(sentences[i].Effects) != 1 || sentences[i].Effects[0].Kind != EffectChooseNewTargets {
			continue
		}
		tokens := semanticEffectTokens(sentences[i].Tokens)
		if !isCopyChooseNewTargetsRiderTokens(tokens) {
			continue
		}
		copyEffect.CopyMayChooseNewTargets = true
		copyEffect.CopyChooseNewTargetsRiderSpan = sentences[i].Span
		foldedEffects = len(sentences[i].Effects)
		if sentences[i].LegacyEffects {
			foldedLegacy = orderedEffectCount(tokens, atoms)
		}
		sentences[i].Effects = nil
		sentences[i].LegacyEffects = false
		sentences[i].CopyChooseNewTargetsRider = true
		return foldedLegacy, foldedEffects, true
	}
	return 0, 0, false
}

// loneCopyStackObjectEffect returns the single copy-stack-object effect across
// the sentences, or nil when the sentences hold zero or more than one.
func loneCopyStackObjectEffect(sentences []Sentence) *EffectSyntax {
	var found *EffectSyntax
	for i := range sentences {
		for j := range sentences[i].Effects {
			effect := &sentences[i].Effects[j]
			if effect.Kind != EffectCopyStackObject {
				continue
			}
			if found != nil {
				return nil
			}
			found = effect
		}
	}
	return found
}

// isCopyChooseNewTargetsRiderTokens reports whether the sentence tokens are
// exactly "You may choose new targets for the copy[ies]." The plural "copies"
// form covers multi-copy effects ("Copy ... X times. You may choose new targets
// for the copies."). Any other wording leaves the rider uncredited.
func isCopyChooseNewTargetsRiderTokens(tokens []shared.Token) bool {
	clause := strings.TrimSuffix(strings.ToLower(joinedEffectText(tokens)), ".")
	return clause == "you may choose new targets for the copy" ||
		clause == "you may choose new targets for the copies"
}

// loneCopyTokenCreateEffect returns the single create-copy-token effect across
// the sentences (a copy of a target, reference, or attached permanent), or nil
// when the sentences hold zero or more than one such effect.
func loneCopyTokenCreateEffect(sentences []Sentence) *EffectSyntax {
	var found *EffectSyntax
	for i := range sentences {
		for j := range sentences[i].Effects {
			effect := &sentences[i].Effects[j]
			if !effect.TokenCopyOfTarget && !effect.TokenCopyOfReference && !effect.TokenCopyOfAttached {
				continue
			}
			if found != nil {
				return nil
			}
			found = effect
		}
	}
	return found
}

// tokenCopyGrantRiderKeywords reports whether the sentence tokens are exactly
// "[That token/Those tokens/It/They] gain(s) <keyword>[ and <keyword> ...]." and
// returns the granted keyword kinds in source order. It fails closed for any
// trailing duration ("until end of turn"), quoted ability, or other content so
// only a plain keyword grant on the created token is folded.
func tokenCopyGrantRiderKeywords(tokens []shared.Token, atoms Atoms) ([]KeywordKind, bool) {
	verb := -1
	for i := range tokens {
		if equalWord(tokens[i], "gains") || equalWord(tokens[i], "gain") {
			verb = i
			break
		}
	}
	if verb <= 0 {
		return nil, false
	}
	subject := strings.ToLower(joinedEffectText(tokens[:verb]))
	switch subject {
	case "that token", "those tokens", "it", "they":
	default:
		return nil, false
	}
	keywordAtoms := atoms.KeywordsWithin(tokens)
	if len(keywordAtoms) == 0 {
		return nil, false
	}
	kinds := make([]KeywordKind, 0, len(keywordAtoms))
	texts := make([]string, 0, len(keywordAtoms))
	for _, keyword := range keywordAtoms {
		if keyword.Parameter.Kind != KeywordParameterNone {
			return nil, false
		}
		kinds = append(kinds, keyword.Kind)
		texts = append(texts, keyword.Text)
	}
	remainder := strings.TrimSuffix(joinedEffectText(tokens[verb+1:]), ".")
	expected := strings.Join(texts, " and ")
	if !strings.EqualFold(normalizeApostrophes(remainder), normalizeApostrophes(expected)) {
		return nil, false
	}
	return kinds, true
}

// targets that subject phrases ("that creature", "a creature destroyed this
// way") would otherwise contribute to the compiled target set.
func isRegenerationRiderTokens(tokens []shared.Token) bool {
	if !effectWordsAt(tokens, 0, "it", "can't", "be", "regenerated") &&
		!effectWordsAt(tokens, 0, "they", "can't", "be", "regenerated") {
		return false
	}
	rest := tokens[4:]
	for i := range rest {
		if rest[i].Kind != shared.Period {
			return false
		}
	}
	return true
}

// isThisWayRegenerationRiderTokens reports whether the sentence tokens are a
// regeneration rider of the "destroyed this way" templated form, for example
// "A creature destroyed this way can't be regenerated." (Damn) or "Creatures
// destroyed this way can't be regenerated." Unlike the bare "that
// creature"/"those creatures" subject forms, this indefinite "destroyed this
// way" clause introduces no back-reference, so it contributes no phantom target
// or reference to the compiled effect and can fold onto the lone destroy safely.
// "Dealt damage this way" riders are intentionally excluded: they belong to a
// damage effect, which has no prevent-regeneration lowering yet, so they remain
// fail-closed instead of silently dropping the clause.
func isThisWayRegenerationRiderTokens(tokens []shared.Token) bool {
	end := len(tokens)
	for end > 0 && tokens[end-1].Kind == shared.Period {
		end--
	}
	core := tokens[:end]
	return endsWithWords(core, "destroyed", "this", "way", "can't", "be", "regenerated")
}

// endsWithWords reports whether the trailing tokens match words in order.
func endsWithWords(tokens []shared.Token, words ...string) bool {
	return effectWordsAt(tokens, len(tokens)-len(words), words...)
}

func spanInsideActivationRestriction(span shared.Span, restrictions []ActivationRestriction) bool {
	for i := range restrictions {
		if spanCovers(restrictions[i].Span, span) || spanCovers(span, restrictions[i].Span) {
			return true
		}
	}
	return false
}

func spanInsideTriggerFrequency(span shared.Span, triggerFrequency *TriggerFrequencyRestriction) bool {
	if triggerFrequency == nil {
		return false
	}
	return spanCovers(triggerFrequency.Span, span) || spanCovers(span, triggerFrequency.Span)
}

func semanticEffectTokens(tokens []shared.Token) []shared.Token {
	result := make([]shared.Token, 0, len(tokens))
	depth := 0
	quoted := false
	for _, token := range tokens {
		switch token.Kind {
		case shared.LeftParen:
			if !quoted {
				depth++
			}
		case shared.RightParen:
			if !quoted && depth > 0 {
				depth--
			}
		case shared.Quote:
			quoted = !quoted
		default:
			if depth == 0 && !quoted {
				result = append(result, token)
			}
		}
	}
	return result
}

// tokensBeforeOffset returns the tokens that end at or before the given source
// offset, preserving order. It is used to scope the recipient Selection of a
// trailing-amount damage clause to the tokens before the amount phrase so the
// amount's counted subject does not contaminate the recipient. Returning a
// contiguous prefix (rather than deleting the amount tokens in place) keeps the
// recipient span from bridging across the removed phrase to later punctuation.
func tokensBeforeOffset(tokens []shared.Token, offset int) []shared.Token {
	result := make([]shared.Token, 0, len(tokens))
	for _, token := range tokens {
		if token.Span.End.Offset > offset {
			break
		}
		result = append(result, token)
	}
	return result
}

// trailingDynamicCountInClause reports whether amount is a trailing dynamic
// count phrase ("for each ...", "equal to ...", "where X is ...") whose tokens
// fall inside clause. A leading count prefix ("For each X, create ...") lives
// before the verb, so its span starts before the clause and is excluded here.
func trailingDynamicCountInClause(clause []shared.Token, amount EffectAmountSyntax) bool {
	switch amount.DynamicForm {
	case EffectDynamicAmountFormForEach, EffectDynamicAmountFormEqual, EffectDynamicAmountFormWhereX:
	default:
		return false
	}
	if len(clause) == 0 {
		return false
	}
	return amount.Span.Start.Offset >= clause[0].Span.Start.Offset
}

// tokenAttackDefenderClause recognizes the trailing "... attacking <defender>"
// relative clause of a created attacking token (CR 508.4) and returns the
// matched defender kind together with the span of the defender tokens (the run
// after "attacking", up to the clause-final period). It matches only the
// create-token effect and the exact modern defender wordings; the bare
// "... attacking." clause and every other tail leave it (None, _, false). The
// returned span lets the build loop scope the token Selection to the tokens
// before the defender phrase, so player/planeswalker nouns in the defender
// ("that player or a planeswalker they control") do not fold into the token's
// own type line.
func tokenAttackDefenderClause(kind EffectKind, clause []shared.Token) (AttackDefenderKind, shared.Span, bool) {
	if kind != EffectCreate {
		return AttackDefenderNone, shared.Span{}, false
	}
	attackIndex := -1
	for i := range clause {
		if equalWord(clause[i], "attacking") {
			attackIndex = i
			break
		}
	}
	if attackIndex < 0 {
		return AttackDefenderNone, shared.Span{}, false
	}
	tail := clause[attackIndex+1:]
	// Drop a trailing clause-final period so the defender words match exactly.
	if len(tail) > 0 && tail[len(tail)-1].Text == "." {
		tail = tail[:len(tail)-1]
	}
	if len(tail) == 0 {
		return AttackDefenderNone, shared.Span{}, false
	}
	defenders := []struct {
		kind  AttackDefenderKind
		words []string
	}{
		{AttackDefenderThatPlayerOrPlaneswalker, []string{"that", "player", "or", "a", "planeswalker", "they", "control"}},
		{AttackDefenderThatPlayer, []string{"that", "player"}},
		{AttackDefenderThatOpponent, []string{"that", "opponent"}},
	}
	for _, d := range defenders {
		if len(tail) == len(d.words) && effectWordsAt(tail, 0, d.words...) {
			return d.kind, shared.SpanOf(tail), true
		}
	}
	return AttackDefenderNone, shared.Span{}, false
}

// canAttackDefenderTailSpan returns the span of the trailing "as though it
// didn't have defender" reminder of a "<source> can attack this turn as though
// it didn't have defender." permission. The build loop stores it so the
// semantic scans remove the anaphoric "it" and the trailing "defender" noun
// from reference and keyword scanning. It is the zero span for any other effect.
func canAttackDefenderTailSpan(kind EffectKind, clause []shared.Token) shared.Span {
	if kind != EffectCanAttackAsThoughDefender {
		return shared.Span{}
	}
	for i := range clause {
		if asThoughDidntHaveDefenderTailAt(clause, i) {
			return shared.SpanOf(clause[i : i+6])
		}
	}
	return shared.Span{}
}

// referencesOutsideAttackDefender drops anaphoric references (e.g. the
// "that player" ReferenceThatPlayer in "... attacking that player or a
// planeswalker they control") that fall inside a created attacking token's
// defender phrase. The runtime EntryAttacking model is defender-agnostic, so
// such a reference has no resolving meaning and would otherwise leave the
// create-token body with an unconsumed reference and fail closed. It is a no-op
// unless the effect carries a recognized token-attack defender.
func referencesOutsideAttackDefender(refs []Reference, hasDefender bool, defender shared.Span) []Reference {
	if !hasDefender {
		return refs
	}
	kept := make([]Reference, 0, len(refs))
	for _, ref := range refs {
		if ref.Span.Start.Offset >= defender.Start.Offset && ref.Span.End.Offset <= defender.End.Offset {
			continue
		}
		kept = append(kept, ref)
	}
	return kept
}

// referencesOutsideTargetCounterQualifiers drops the "it"/"them" pronoun that
// closes a modeled counter qualifier inside a target noun phrase. The target's
// typed Selection owns that pronoun; retaining it as a free semantic reference
// would make lowerers treat the qualifier as a second resolving object.
func referencesOutsideTargetCounterQualifiers(refs []Reference, targets []TargetSyntax) []Reference {
	kept := make([]Reference, 0, len(refs))
	for _, ref := range refs {
		internal := false
		for _, target := range targets {
			if targetOwnsCounterQualifierReference(target, ref) {
				internal = true
				break
			}
		}
		if !internal {
			kept = append(kept, ref)
		}
	}
	return kept
}

func targetOwnsCounterQualifierReference(target TargetSyntax, ref Reference) bool {
	return selectionHasCounterQualifier(target.Selection) &&
		ref.Kind == ReferencePronoun &&
		(ref.Pronoun == PronounIt || ref.Pronoun == PronounThem) &&
		spanCovers(target.Span, ref.Span)
}

// stripLeadingConditionClause drops a leading "As long as ..." condition clause
// so the subject grammar sees only the effect's group subject ("creatures you
// control"). The first effect's ownership tokens begin at the sentence start, so
// a leading "As long as <condition>, ..." gate (the Ascension cycle's "As long
// as ~ has seven or more quest counters on it, creatures you control get +X/+X",
// and the Incarnation cycle's graveyard zone-of-function condition) would
// otherwise prevent the group subject from being recognized at token zero. A
// leading "If <source> has [a <kind>] counter[s] on it, ..." gate is likewise
// dropped: it is the counter-conditional mana multiplier rider's condition
// (Incubation Druid's "If this creature has a +1/+1 counter on it, add three
// mana of that type instead."), whose "has" verb would otherwise seed a spurious
// keyword-grant effect from the gate. The condition clause itself is recognized
// separately, so removing it here only affects subject and effect-verb
// recognition of the gated body.
func stripLeadingConditionClause(tokens []shared.Token, atoms Atoms) []shared.Token {
	if len(tokens) == 0 {
		return tokens
	}
	intro, introWidth := conditionIntroAt(tokens, 0)
	end := conditionClauseEnd(tokens, 0)
	if end >= len(tokens) || tokens[end].Kind != shared.Comma {
		return tokens
	}
	switch intro {
	case ConditionIntroAsLongAs:
		return tokens[end+1:]
	case ConditionIntroIf:
		if _, ok := recognizeSourceCounterStateCondition(tokens[introWidth:end], atoms); ok {
			return tokens[end+1:]
		}
	default:
	}
	return tokens
}

// stripLeadingDurationClause removes a sentence-leading duration clause
// ("Until end of turn, ...", "Until your next turn, ...") from the token slice,
// returning the remaining tokens and the duration it names. A duration stated
// once at the front of a sentence applies to every continuous effect the
// sentence produces (CR 611.2: "Until end of turn, creatures you control gain
// trample and get +X/+X ..."), so the parser lifts it off the front where it
// would otherwise derail the group-subject parse and leave the trailing effect
// with no duration.
func stripLeadingDurationClause(tokens []shared.Token, atoms Atoms) ([]shared.Token, EffectDurationKind) {
	if len(tokens) == 0 || !equalWord(tokens[0], "until") {
		return tokens, EffectDurationNone
	}
	comma := -1
	for i := range tokens {
		if tokens[i].Kind == shared.Comma {
			comma = i
			break
		}
	}
	if comma <= 0 {
		return tokens, EffectDurationNone
	}
	duration := parseEffectDuration(tokens[:comma], atoms)
	if duration == EffectDurationNone {
		return tokens, EffectDurationNone
	}
	return tokens[comma+1:], duration
}

// leadingForEachDamagePreventedThisWay reports whether the effect's ownership
// tokens open with a "for each <count> damage prevented this way," iteration
// prefix. The prefix scales the effect by a companion prevention shield's
// prevented total (Brace for Impact's "For each 1 damage prevented this way, put
// a +1/+1 counter on that creature."), a dynamic amount no runtime construct
// tracks. The per-clause loop never folds this leading prefix into the effect's
// amount, so callers use this to fail the effect closed rather than round-trip
// the sentence while silently dropping the multiplier.
func leadingForEachDamagePreventedThisWay(effect *EffectSyntax) bool {
	tokens := effect.Tokens
	if len(tokens) < 2 || !effectWordsAt(tokens, 0, "for", "each") {
		return false
	}
	for i := 2; i+2 < len(tokens); i++ {
		if tokens[i].Kind == shared.Comma {
			return false
		}
		if effectWordsAt(tokens, i, "prevented", "this", "way") {
			return true
		}
	}
	return false
}

// parseSpecialEffects dispatches the sentence to the whole-sentence effect
// recognizers that bypass the per-clause loop in parseEffects. It returns the
// first recognizer's result, or ok=false when none match and the general
// per-clause parsing should run. Order is significant and matches the original
// dispatch sequence.
func parseSpecialEffects(sentence Sentence, tokens []shared.Token, atoms Atoms) ([]EffectSyntax, bool) {
	for _, recognize := range []func() ([]EffectSyntax, bool){
		func() ([]EffectSyntax, bool) { return parsePassiveTokenDoublingEffects(sentence, tokens, atoms) },
		func() ([]EffectSyntax, bool) { return parsePassiveTokenAdditiveEffects(sentence, tokens, atoms) },
		func() ([]EffectSyntax, bool) { return parsePassiveTokenIdentityEffects(sentence, tokens, atoms) },
		func() ([]EffectSyntax, bool) { return parseEntersAsCopyEffect(sentence, tokens, atoms) },
		func() ([]EffectSyntax, bool) { return parseDevourEffect(sentence, tokens, atoms) },
		func() ([]EffectSyntax, bool) { return parseTributeEffect(sentence, tokens, atoms) },
		func() ([]EffectSyntax, bool) { return parseBecomeCopyEffect(sentence, tokens, atoms) },
		func() ([]EffectSyntax, bool) { return parsePolymorphEffect(sentence, tokens, atoms) },
		func() ([]EffectSyntax, bool) { return parseSetBasePowerToughnessEffect(sentence, tokens, atoms) },
		func() ([]EffectSyntax, bool) { return parseAnimateSelfEffect(sentence, tokens, atoms) },
		func() ([]EffectSyntax, bool) { return parseAnimateTargetEffect(sentence, tokens, atoms) },
		func() ([]EffectSyntax, bool) { return parseSwitchPowerToughnessEffect(sentence, tokens, atoms) },
		func() ([]EffectSyntax, bool) { return parseNamedBecomePolymorphEffect(sentence, tokens, atoms) },
		func() ([]EffectSyntax, bool) { return parseBecomeTypeEffect(sentence, tokens, atoms) },
		func() ([]EffectSyntax, bool) { return parseReferencedTypeGrantEffect(sentence, tokens, atoms) },
		func() ([]EffectSyntax, bool) { return parseBecomeColorEffect(sentence, tokens, atoms) },
		func() ([]EffectSyntax, bool) { return parseDrawEmptyLibraryWinReplacement(sentence, tokens, atoms) },
		func() ([]EffectSyntax, bool) { return parseDrawDoublingReplacement(sentence, tokens, atoms) },
		func() ([]EffectSyntax, bool) { return parseDrawReplacementDig(sentence, tokens, atoms) },
		func() ([]EffectSyntax, bool) { return parseLifeGainReplacement(sentence, tokens, atoms) },
		func() ([]EffectSyntax, bool) {
			return parseLeaveBattlefieldExileReplacement(sentence, tokens, atoms)
		},
		func() ([]EffectSyntax, bool) {
			return parseDieThisTurnExileReplacement(sentence, tokens, atoms)
		},
		func() ([]EffectSyntax, bool) { return parseLifeLossReplacement(sentence, tokens, atoms) },
		func() ([]EffectSyntax, bool) { return parseExcessDamageToControllerEffect(sentence, tokens, atoms) },
		func() ([]EffectSyntax, bool) { return parsePunisherEachLoseLifeEffect(sentence, tokens, atoms) },
		func() ([]EffectSyntax, bool) { return parseLibraryTopReorderEffect(sentence, tokens, atoms) },
		func() ([]EffectSyntax, bool) { return parseGroupEntersTappedEffect(sentence, tokens) },
		func() ([]EffectSyntax, bool) { return parseGroupEntersWithCountersEffect(sentence, tokens, atoms) },
		func() ([]EffectSyntax, bool) { return parsePlayerProtectionEffects(sentence, tokens, atoms) },
		func() ([]EffectSyntax, bool) { return parseGroupPhaseOutEffect(sentence, tokens, atoms) },
		func() ([]EffectSyntax, bool) { return parsePhaseOutEffect(sentence, tokens, atoms) },
		func() ([]EffectSyntax, bool) { return parseMassReanimationExchangeEffect(sentence, tokens, atoms) },
		func() ([]EffectSyntax, bool) { return parseAdditionalLandPlaysEffect(sentence, tokens, atoms) },
		func() ([]EffectSyntax, bool) { return parseCastAsThoughFlashEffect(sentence, tokens) },
		func() ([]EffectSyntax, bool) { return parsePlayFromLibraryTopEffect(sentence, tokens, atoms) },
		func() ([]EffectSyntax, bool) { return parsePlayExiledCardEffect(sentence, tokens, atoms) },
		func() ([]EffectSyntax, bool) { return parsePlayThatCardEffect(sentence, tokens, atoms) },
		func() ([]EffectSyntax, bool) { return parseAdditionalCombatPhaseEffect(sentence, tokens) },
		func() ([]EffectSyntax, bool) { return parseRollDieEffect(sentence, tokens) },
		func() ([]EffectSyntax, bool) { return parseRingTemptsEffect(sentence, tokens) },
		func() ([]EffectSyntax, bool) { return parseNoMaximumHandSizeForRestOfGameEffect(sentence, tokens) },
		func() ([]EffectSyntax, bool) { return parseCantCastSpellsEffect(sentence, tokens) },
		func() ([]EffectSyntax, bool) { return parseSpellCostModifierEffect(sentence, tokens) },
		func() ([]EffectSyntax, bool) { return parseGroupMustAttackEffect(sentence, tokens, atoms) },
		func() ([]EffectSyntax, bool) { return parseGroupCantBlockEffect(sentence, tokens, atoms) },
		func() ([]EffectSyntax, bool) { return parseDirectedMustAttackEffect(sentence, tokens, atoms) },
		func() ([]EffectSyntax, bool) { return parseAttackTaxEffect(sentence, tokens, atoms) },
		func() ([]EffectSyntax, bool) { return parseSpellsCantBeCounteredEffect(sentence, tokens) },
		func() ([]EffectSyntax, bool) { return parseChangeTargetRetargetEffect(sentence, tokens, atoms) },
	} {
		if effects, ok := recognize(); ok {
			return effects, true
		}
	}
	return nil, false
}

func parseEffects(sentence Sentence, tokens []shared.Token, atoms Atoms) []EffectSyntax {
	if effects, ok := parseSpecialEffects(sentence, tokens, atoms); ok {
		return effects
	}
	if effects, ok := parsePreventCombatDamageEffect(sentence, tokens, atoms); ok {
		return effects
	}
	if effects, ok := parsePreventAllDamageTargetEffect(sentence, tokens, atoms); ok {
		return effects
	}
	if effects, ok := parsePreventNextDamageFromSourceEffect(sentence, tokens, atoms); ok {
		return effects
	}
	if effects, ok := parsePreventAmountDamageEffect(sentence, tokens, atoms); ok {
		return effects
	}
	if effects, ok := parsePreventThatDamageEffect(sentence, tokens, atoms); ok {
		return effects
	}
	indices := effectIndices(tokens, atoms)
	requiresOrderedLowering := orderedEffectCount(tokens, atoms) > 1
	// Fold a trailing "<effect> unless you <non-mana cost>" controller payment:
	// drop the cost verbs from segmentation so the gated effect parses as a
	// single payment-bearing effect rather than a spurious two-effect sequence,
	// and carry the recognized AdditionalCost forward onto that effect's Payment.
	unlessCost := recognizeUnlessControllerAdditionalCost(sentence, tokens, atoms)
	if unlessCost.ok {
		filtered := indices[:0:0]
		for _, idx := range indices {
			if idx < unlessCost.unlessIndex {
				filtered = append(filtered, idx)
			}
		}
		if len(filtered) == len(indices) || len(filtered) == 0 {
			// No effect precedes the payment, or the payment consumed no effect
			// verb; leave the sentence to its ordinary segmentation.
			unlessCost = recognizedUnlessCost{}
		} else {
			indices = filtered
			requiresOrderedLowering = orderedEffectCount(tokens, atoms) > 1
		}
	}
	effects := make([]EffectSyntax, 0, len(indices))
	_, leadingDuration := stripLeadingDurationClause(tokens, atoms)
	for effectIndex, tokenIndex := range indices {
		clauseEnd := resolvingClauseEnd(tokens, indices, effectIndex)
		ownershipStart := resolvingClauseStart(tokens, indices, effectIndex)
		ownership := tokens[ownershipStart:clauseEnd]
		clause := tokens[tokenIndex+1 : clauseEnd]
		clause, delayed := cutDelayedTiming(clause)
		if delayed == DelayedTimingNone {
			delayed = leadingDelayedTiming(tokens[ownershipStart:tokenIndex])
		}
		power, toughness := parsePTChange(clause)
		span := shared.SpanOf(clause)
		ownershipSpan := shared.SpanOf(ownership)
		toZone := firstZone(atoms, span, ZoneRoleTo)
		if ambiguousZoneChoice(ownership, atoms, span) {
			toZone = zone.None
		}
		subjectTokens := stripLeadingConditionClause(ownership, atoms)
		subjectTokens, _ = stripLeadingDurationClause(subjectTokens, atoms)
		staticSubject := parseEffectStaticSubject(subjectTokens, atoms)
		payment := parseEffectPayment(tokens, atoms)
		if payment.Form == EffectPaymentFormUnknown && unlessCost.ok {
			payment = unlessCost.payment
		}
		connection, connectionSpan := effectConnection(tokens, indices, effectIndex)
		optional, optionalSpan := effectOptional(tokens, tokenIndex)
		context := effectContextAt(tokens, tokenIndex, atoms)
		if effectIndex > 0 && !effectHasExplicitSubject(tokens, tokenIndex, atoms.SelfNameSpans()) &&
			effects[len(effects)-1].Context != EffectContextController {
			context = EffectContextPriorSubject
		}
		durationTokens := ownership
		nextConnection := EffectConnectionNone
		if effectIndex+1 < len(indices) {
			nextConnection, _ = effectConnection(tokens, indices, effectIndex+1)
			if nextConnection == EffectConnectionAnd &&
				durationScopesAcrossAnd(effectKindAt(tokens, tokenIndex), effectKindAt(tokens, indices[effectIndex+1])) {
				durationTokens = tokens
			}
		}
		duration := parseEffectDuration(durationTokens, atoms)
		if duration == EffectDurationNone {
			duration = leadingDuration
		}
		kind := effectKindAt(tokens, tokenIndex)
		if loseGameObject(kind, clause) {
			kind = EffectLoseGame
		}
		entersColorChoice, entersColorChoiceExclude := entersColorChoiceSyntax(kind, clause)
		doublePower, doubleToughness := false, false
		doubleSourceCounters := false
		var doubleSourceCounterKind counter.Kind
		doubleCountersTarget := false
		doubleCountersAllKinds := false
		doubleCountersGroup := false
		if kind == EffectDouble {
			if object, okDouble := parseDoublePTObject(clause, atoms); okDouble {
				staticSubject = object.Subject
				doublePower, doubleToughness = object.DoublePower, object.DoubleToughness
			} else if power, toughness, okTargetPT := parsePossessiveDoublePTObject(clause); okTargetPT {
				doublePower, doubleToughness = power, toughness
			} else if counters, okCounters := parseDoubleCountersObject(clause, atoms); okCounters {
				doubleSourceCounters = true
				doubleSourceCounterKind = counters.Kind
				doubleCountersTarget = counters.Target
				doubleCountersAllKinds = counters.AllKinds
				if counters.Group {
					doubleCountersGroup = true
					staticSubject = counters.Subject
				}
			}
		}
		tokenPower, tokenToughness, tokenPTKnown := parseTokenPowerToughness(kind, clause)
		tokenPTVariableX := parseTokenPTVariableX(kind, clause)
		amount := parseEffectAmount(kind, clause, atoms)
		if forEach, ok := parseCreateForEachAmount(kind, context, tokenPTKnown, tokens[ownershipStart:tokenIndex], amount, atoms); ok {
			amount = forEach
		}
		counterKind, counterKnown := parseCounterPlacementScopedToCount(entersTappedCounterClause(kind, clause), amount, atoms)
		var counterKindChoices []counter.Kind
		if kind == EffectPut && !counterKnown {
			counterKindChoices = parseCounterPlacementChoices(clause, atoms)
		}
		// A deal-damage clause whose amount is a trailing "where X is the number
		// of ..." count phrase ("deals X damage to each creature, where X is the
		// number of Gates you control.") embeds the counted-subject selector in
		// the same clause as the recipient group. parseSelection scans the span
		// of its tokens, so leaving the count phrase in would fold the count
		// subject's filters (here "Gate" and "you control") into the recipient,
		// and merely deleting the count tokens would still bridge the span across
		// the trailing sentence period. The recipient group is exactly the run of
		// tokens before the trailing count phrase, so scope the recipient
		// Selection to those, leaving the count subject to the amount's own
		// selector.
		// A create-token clause whose amount is a trailing "for each <permanent>
		// you control" (or "equal to ...") count phrase ("Create a 0/1 green
		// Plant creature token for each land you control.") embeds the
		// counted-subject selector in the same clause as the token's own
		// characteristics. Like the deal-damage case above, scope the token
		// Selection to the run of tokens before the count phrase so the count
		// subject's filters do not fold into the token's type line.
		selectionClause := clause
		tokenAttackDefender, tokenAttackDefenderSpan, hasAttackDefender := tokenAttackDefenderClause(kind, clause)
		canAttackDefenderSpan := canAttackDefenderTailSpan(kind, clause)
		switch {
		case kind == EffectDealDamage && amount.DynamicForm == EffectDynamicAmountFormWhereX:
			selectionClause = tokensBeforeOffset(clause, amount.Span.Start.Offset)
		case kind == EffectDealDamage && amount.DynamicForm == EffectDynamicAmountFormEqual && groupDamageRecipientFollowsAmount(clause, amount):
			// "deals damage equal to its power to each opponent" embeds the
			// "equal to ..." amount phrase before the recipient group, so the
			// full clause folds the amount's "its power" referent into the
			// recipient selection. Scope the recipient Selection to the run of
			// tokens after the amount (the "each <group>" phrase) so the amount's
			// referent does not bleed into the recipient. The WhereX case above
			// is the mirror image, where the count phrase trails the recipient.
			if recipient, ok := damageRecipientTokensAfterAmount(clause, amount); ok {
				selectionClause = recipient
			}
		case kind == EffectCreate && trailingDynamicCountInClause(clause, amount):
			selectionClause = tokensBeforeOffset(clause, amount.Span.Start.Offset)
		case hasAttackDefender:
			// Scope the token Selection to the tokens before the trailing
			// "... attacking <defender>" phrase so the defender's player and
			// planeswalker nouns do not fold into the token's own type line.
			selectionClause = tokensBeforeOffset(clause, tokenAttackDefenderSpan.Start.Offset)
		case kind == EffectMill && trailingDynamicCountInClause(clause, amount):
			// "mills cards equal to <dynamic>" embeds the count subject in the
			// same clause as the milled-card noun. Scoping the selection to the
			// tokens before the count phrase keeps a competing permanent noun in
			// the amount (e.g. "the sacrificed creature's power") from folding
			// into the milled-card selection.
			selectionClause = tokensBeforeOffset(clause, amount.Span.Start.Offset)
		case kind == EffectPut && amount.DynamicForm == EffectDynamicAmountFormWhereX:
			// "put X +1/+1 counters on each creature you control, where X is the
			// number of Shrines you control" trails the count phrase after the
			// recipient group. Like the deal-damage WhereX case above, scope the
			// recipient Selection to the tokens before the count phrase so the
			// count subject's subtype or referent ("Shrines", "this creature's
			// power") does not fold into the recipient group's type line.
			selectionClause = tokensBeforeOffset(clause, amount.Span.Start.Offset)
		case kind == EffectSearch && amount.DynamicForm == EffectDynamicAmountFormWhereX:
			// "Search your library for up to X basic land cards, where X is the
			// number of lands you control" trails the count phrase after the
			// searched filter. Like the deal-damage and put WhereX cases above,
			// scope the searched-card Selection to the tokens before the count
			// phrase so the count subject's noun ("lands", "creatures") does not
			// fold into the searched card's filter.
			selectionClause = tokensBeforeOffset(clause, amount.Span.Start.Offset)
		default:
		}
		eachSourceDamageGroup, eachSourceDamageRecipient := eachSourceDamageSyntax(kind, tokens[ownershipStart:tokenIndex], clause, amount, atoms)
		fallbackOnInability := effectFallbackOnInability(tokens, ownershipStart, tokenIndex)
		effectSelection := parseSelection(selectionClause, atoms)
		// A group selection naming a creature conjoined with one other permanent
		// type ("each artifact creature you control") records both card types in
		// RequiredTypesAny with no conjunctive marker, exactly as a target does.
		// Mark it so the lowering reads the two-word type as a single all-of
		// filter rather than an any-of union, mirroring conjunctive targets.
		if conjunctiveTypeTarget(effectSelection) {
			effectSelection.ConjunctiveTypes = true
		}
		// "put a Saga card and/or a land card from among them onto the
		// battlefield" joins two singular articled card nouns with "and/or",
		// meaning up to one card of each named type rather than a single card of
		// either. Record it so the lowering realizes one optional pick per named
		// type. The repeated singular "card" noun distinguishes it from a plural
		// union ("artifacts, creatures, and/or lands"), which is a single match.
		if kind == EffectPut && selectionJoinsCardNounsWithAndOr(selectionClause) {
			effectSelection.InclusiveOneOfEach = true
		}
		// "put any number of <type> cards from among them" lets the resolving
		// player choose any quantity from none up to all eligible cards. The
		// literal "any number of" run is the only signal for this unbounded
		// count: "all", "the", and a bare plural noun all compile to the same
		// empty amount, so record the marker positively here rather than
		// inferring it downstream from the absence of a count.
		if kind == EffectPut && effectHasTokenWords(selectionClause, "any", "number", "of") {
			amount.AnyNumber = true
		}
		// A created token's name is printed either as a leading "Create <Name>, a
		// ..." prefix (named legendary tokens) or as a trailing "named <Name>"
		// tail. Prefer the leading form when present and record its placement so
		// the exactness recognizer reconstructs the name in the right position.
		tokenName := parseTokenName(kind, clause)
		tokenNameLeading := false
		if leading := parseLeadingTokenName(kind, clause); leading != "" {
			tokenName = leading
			tokenNameLeading = true
		}
		tokenKeywords := parseTokenKeywords(kind, clause, atoms)
		tokenToxic := parseTokenKeywordToxic(kind, clause, atoms)
		tokenPredefinedName := parsePredefinedTokenName(kind, clause)
		// A multi-token create clause ("Create a 1/1 green Snake creature token, a
		// 2/2 green Wolf creature token, and a 3/3 green Elephant creature token.")
		// merges every named token's subtypes and characteristics into one
		// selection when scanned as a single spec. Split it into one spec per
		// created token: the first spec replaces this effect's own merged token
		// fields, and the rest ride on AdditionalTokens for the lowering to emit
		// as a sequence of CreateToken instructions. Only the controller-creates
		// form is split; every other clause keeps its single-token fields.
		var additionalTokens []EffectSyntax
		if context == EffectContextController {
			if specs, ok := multiTokenCreateSpecs(kind, clause, atoms); ok {
				first := specs[0]
				effectSelection = first.Selection
				tokenPower = first.TokenPower
				tokenToughness = first.TokenToughness
				tokenPTKnown = first.TokenPTKnown
				tokenPTVariableX = false
				tokenKeywords = first.TokenKeywords
				tokenToxic = first.TokenToxic
				tokenName = first.TokenName
				tokenNameLeading = false
				tokenPredefinedName = first.TokenPredefinedName
				additionalTokens = specs[1:]
			}
		}
		effects = append(effects, EffectSyntax{
			Kind:           kind,
			Context:        context,
			Connection:     connection,
			ConnectionSpan: connectionSpan,
			Span:           sentence.Span,
			VerbSpan:       tokens[tokenIndex].Span,
			ClauseSpan:     ownershipSpan,
			Text:           sentence.Text,
			Tokens:         append([]shared.Token(nil), ownership...),
			Duration:       duration,
			DelayedTiming:  delayed,
			Selection:      effectSelection,
			DamageRecipient: DamageRecipientSyntax{
				Groups:          parseDamageRecipientPair(kind, clause, amount, atoms),
				EachSourceGroup: eachSourceDamageGroup,
				EachSourceRole:  eachSourceDamageRecipient,
			},
			Amount:                   amount,
			AmassSubtype:             parseAmassSubtype(kind, clause),
			PowerDelta:               power,
			ToughnessDelta:           toughness,
			TokenPower:               tokenPower,
			TokenToughness:           tokenToughness,
			TokenPTKnown:             tokenPTKnown,
			TokenPTVariableX:         tokenPTVariableX,
			TokenKeywords:            tokenKeywords,
			TokenToxic:               tokenToxic,
			TokenName:                tokenName,
			TokenPredefinedName:      tokenPredefinedName,
			TokenNameLeading:         tokenNameLeading,
			AdditionalTokens:         additionalTokens,
			AttackDefender:           tokenAttackDefender,
			AttackDefenderSpan:       tokenAttackDefenderSpan,
			CanAttackDefenderSpan:    canAttackDefenderSpan,
			TokenChoice:              parseTokenChoice(kind, clause),
			StaticSubject:            staticSubject,
			SubjectSourceAttached:    resolvingAttachedPossessiveSubject(ownership, staticSubject),
			DoublePower:              doublePower,
			DoubleToughness:          doubleToughness,
			DoubleSourceCounters:     doubleSourceCounters,
			DoubleSourceCounterKind:  doubleSourceCounterKind,
			DoubleCountersTarget:     doubleCountersTarget,
			DoubleCountersAllKinds:   doubleCountersAllKinds,
			DoubleCountersGroup:      doubleCountersGroup,
			CounterKind:              counterKind,
			CounterKnown:             counterKnown,
			CounterKindChoices:       counterKindChoices,
			CounterRecipientAttached: counterRecipientAttached(kind, counterKnown, clause),
			FightSubjectAttached:     fightSubjectAttached(kind, tokens[ownershipStart:tokenIndex]),
			MoveCountersAll:          kind == EffectMoveCounters && moveAllCountersClause(clause),
			RemoveCountersAll:        kind == EffectRemoveCounter && removeAllCountersClause(clause),
			MoveCountersDistribute:   kind == EffectMoveCounters && moveCountersDistributeClause(clause),
			MoveThoseCounters:        kind == EffectPut && moveThoseCountersClause(clause),
			FromZone:                 effectFromZone(kind, clause, atoms, span, toZone),
			ToZone:                   toZone,
			Destination:              parseEffectDestination(ownership),
			EntersTapped:             effectWordsAtAny(ownership, "battlefield", "tapped"),
			EntersTappedSelf:         entersTappedSelfSyntax(kind, clause),
			EntersColorChoice:        entersColorChoice,
			EntersColorChoiceExclude: entersColorChoiceExclude,
			EntersTypeChoice:         entersTypeChoiceSyntax(kind, clause),
			EntersWithCounters:       entersWithCountersSyntax(kind, clause),
			UnderYourControl:         effectContainsWords(normalizedWords(ownership), "under", "your", "control"),
			UnderOwnersControl:       underOwnersControl(ownership),
			CastAsAdventure:          effectContainsWords(normalizedWords(clause), "as", "an", "adventure"),
			CastWithoutPayingManaCost: kind == EffectCast &&
				effectContainsWords(normalizedWords(clause), "without", "paying", "its", "mana", "cost"),
			Negated:             effectIsNegated(tokens, tokenIndex) && !fallbackOnInability,
			FallbackOnInability: fallbackOnInability,
			Optional:            optional,
			OptionalSpan:        optionalSpan,
			LifeObject:          gainLoseLifeObject(kind, clause),
			LoseAllAbilities:    loseAllAbilitiesObject(kind, sentence.Text),
			Symbol:              firstEffectSymbol(clause),
			Mana:                parseEffectMana(kind, clause, nextConnection != EffectConnectionNone),
			Replacement:         parseEffectReplacement(ownership, atoms),
			References: referencesOutsideTargetCounterQualifiers(
				referencesOutsideSpan(
					referencesOutsideAttackDefender(
						referencesInSpan(atoms, ownershipSpan),
						hasAttackDefender,
						tokenAttackDefenderSpan,
					),
					canAttackDefenderSpan,
				),
				sentence.Targets,
			),
			SubjectReferences:       referencesInSpan(atoms, shared.SpanOf(tokens[ownershipStart:tokenIndex])),
			Targets:                 targetsInSpan(sentence.Targets, ownershipSpan),
			SubjectTargets:          targetsInSpan(sentence.Targets, shared.SpanOf(tokens[ownershipStart:tokenIndex])),
			Payment:                 payment,
			RequiresOrderedLowering: requiresOrderedLowering,
		})
	}

	for i := range effects {
		finalizeParsedEffect(&effects[i], sentence, atoms)
	}
	return effects
}

func finalizeParsedEffect(effect *EffectSyntax, sentence Sentence, atoms Atoms) {
	effect.Divided = dividedDamageEffect(effect)
	effect.DistributeCounters = distributeCountersEffect(effect)
	effect.DamageRecipient.Reference = damageRecipientReference(effect)
	effect.DamageRiders = parseDamageRiders(effect)
	effect.Dig = parseDigPut(effect)
	effect.HandLibraryPut = parseHandLibraryPut(effect)
	effect.HandDiscard = parseHandDiscard(effect)
	effect.RandomDiscard = exactNonControllerRandomDiscardSyntax(effect)
	effect.DiscardEntireHand = parseDiscardEntireHand(effect)
	effect.SearchSplit = parseSearchSplitPut(effect)
	effect.GraveyardZoneExile = parseGraveyardZoneExile(effect)
	parseExileTopOfLibrary(effect)
	effect.Additional = drawAdditionalCardsQualifier(effect)
	effect.MoveCountersFromTarget = effect.Kind == EffectMoveCounters &&
		!effect.MoveCountersDistribute && len(effect.Targets) == 2
	effect.MoveCountersAnyKind = effect.Kind == EffectMoveCounters &&
		!effect.MoveCountersDistribute && !effect.MoveCountersAll && !effect.CounterKnown
	effect.Exact = exactEffectSyntax(effect)
	// A leading "For each 1 damage prevented this way, <effect>" rider scales the
	// effect by the amount of damage a companion prevention shield stops this turn
	// (Brace for Impact, Test of Faith, Temper). No runtime construct tracks
	// "damage prevented this way", and the leading iteration prefix is not folded
	// into the effect's amount (the counter/life amount stays a fixed 1), so the
	// parser would otherwise round-trip the sentence while silently dropping the
	// multiplier. Fail the effect closed instead so such cards are not generated
	// with a wrong fixed amount.
	if effect.Exact && leadingForEachDamagePreventedThisWay(effect) {
		effect.Exact = false
	}
	effect.KeywordGrantChoice = keywordGrantIsChoice(effect)
	if recognizeTargetOpponentHandMana(effect) {
		effect.Exact = true
	}
	if recognizeDynamicCountMana(effect) {
		effect.Exact = true
	}
	if recognizeColorsAmongControlledMana(effect, atoms) {
		effect.Exact = true
	}
	if recognizeMillHalfLibrary(effect) {
		effect.Exact = exactEffectSyntax(effect)
	}
	if recognizeEachColorAmongControlledMana(effect, atoms) {
		effect.Exact = true
	}
	if recognizeTargetColorIfRider(effect, atoms) {
		effect.Exact = true
	}
	// "<verb> each <permanent group>" selects every matching permanent like the
	// plural "all" form, so flag its selection as a mass group to lower to a
	// battlefield-group effect. Scoped to the recognized mass-each forms of the
	// group verbs (destroy, exile, tap, untap, regenerate) so "each creature"
	// damage recipients and "each player" distributive effects on other effect
	// kinds are untouched.
	if massEachGroupVerbEffectSyntax(effect) {
		effect.Selection.All = true
	}
	// "Return each <group> to its owner's hand" selects every matching permanent
	// like the plural "all" mass bounce, so flag its selection as a mass group to
	// lower to a single group Bounce.
	if exactMassEachBounceEffectSyntax(effect) {
		effect.Selection.All = true
	}
	effect.CounterRecipientSingleChoice = effect.Exact && counterPlacementSingleChoiceRecipient(effect)
	effect.TokenCopyOfTarget = exactCreateCopyTokenEffectSyntax(effect)
	effect.TokenCopyOfReference = exactCreateCopyTokenReferenceEffectSyntax(effect)
	effect.TokenCopyOfTriggeringSet = exactCreateCopyTokenTriggeringSetEffectSyntax(effect)
	effect.TokenCopyOfAttached = exactCreateCopyTokenAttachedEffectSyntax(effect)
	effect.RegenerateAttached = effect.Kind == EffectRegenerate && exactRegenerateAttachedEffectSyntax(effect)
	effect.ExileAttached = effect.Kind == EffectExile && exactExileAttachedEffectSyntax(effect)
	effect.TapAttached = effect.Kind == EffectTap && exactTapAttachedEffectSyntax(effect)
	effect.UntapAttached = effect.Kind == EffectUntap && exactUntapAttachedEffectSyntax(effect)
	// A mass self-stun "<group> you control don't untap during your next untap
	// step." carries its affected group only in its subject noun (no target or
	// reference), so record that controlled-permanent group in StaticSubject for
	// the group skip-untap lowering. The subject tokens are the clause's first
	// three semantic tokens ("<group> you control"); crediting their span keeps
	// the consumed-token accounting exact.
	if kind, ok := negatedControlledGroupNextUntapStep(effect); ok {
		semantic := semanticEffectTokens(effect.Tokens)
		effect.StaticSubject = EffectStaticSubjectSyntax{
			Kind: kind,
			Span: shared.SpanOf(semantic[:3]),
		}
	}
	if group, ok := exactCreateCopyTokenForEachEffectSyntax(effect, atoms); ok {
		effect.TokenCopyOfForEach = true
		effect.TokenCopyForEachGroup = group
		// The per-each copy is the more specific shape: its "For each <group>,"
		// prefix iterates the controlled group and its trailing "that <permanent>"
		// pronoun names each member, not a single fixed source. Clear the
		// single-source copy flags the reference matcher also set so the lowering
		// dispatches to the per-each path.
		effect.TokenCopyOfTarget = false
		effect.TokenCopyOfReference = false
		effect.TokenCopyOfTriggeringSet = false
		effect.TokenCopyOfAttached = false
	}
	effect.Mana.LegacyBodyExact = legacyExactManaBody(effect, sentence)
	if effect.Kind == EffectSearch {
		effect.UnsupportedDetail = searchUnsupportedDetail(effect)
		effect.SearchSharedSubtype = searchSharedSubtypeRider(effect)
		effect.SearchDifferentNames = searchDifferentNamesRider(effect)
		effect.SearchDestination = searchDestinationPosition(effect)
		effect.SearchControl = searchControlRider(effect)
		effect.SearchSlots = searchHeterogeneousSlotSubtypes(effect)
	}
}

func parseLibraryTopReorderEffect(sentence Sentence, tokens []shared.Token, atoms Atoms) ([]EffectSyntax, bool) {
	amount, ok := matchLibraryTopReorder(tokens, atoms)
	if !ok {
		return nil, false
	}
	effect := EffectSyntax{
		Kind:       EffectReorderLibraryTop,
		Context:    EffectContextController,
		Span:       sentence.Span,
		VerbSpan:   tokens[0].Span,
		ClauseSpan: sentence.Span,
		Text:       sentence.Text,
		Tokens:     append([]shared.Token(nil), tokens...),
		Amount:     amount,
		References: referencesInSpan(atoms, sentence.Span),
	}
	effect.Exact = exactEffectSyntax(&effect)
	return []EffectSyntax{effect}, true
}

func matchLibraryTopReorder(tokens []shared.Token, atoms Atoms) (EffectAmountSyntax, bool) {
	if len(tokens) != 18 ||
		!effectWordsAt(tokens, 0, "look", "at", "the", "top") ||
		!effectWordsAt(tokens, 5, "cards", "of", "your", "library") ||
		tokens[9].Kind != shared.Comma ||
		!effectWordsAt(tokens, 10, "then", "put", "them", "back", "in", "any", "order") ||
		tokens[17].Kind != shared.Period {
		return EffectAmountSyntax{}, false
	}
	amount := parseEffectAmount(EffectReorderLibraryTop, tokens[4:5], atoms)
	return amount, amount.Known && amount.Value > 0
}

// parsePassiveTokenDoublingEffects recognizes the passive-voice token-doubling
// replacement "If one or more tokens would be created under your control, twice
// that many of those tokens are created instead." (Mondrak, Adrix and Nev). Its
// active-voice equivalent "If an effect would create one or more tokens under
// your control, it creates twice that many of those tokens instead." (Doubling
// Season, Anointed Procession, Parallel Lives) parses through the ordinary
// create-verb path. The passive wording carries no active "create" verb, so it
// is recognized here and emitted as the same two EffectCreate instructions the
// active form produces: the would-create group and the doubled output marked
// EffectReplacementTwiceThatMany. The matching intervening-if condition is
// recognized separately by recognizeTokenCreationCondition.
func parsePassiveTokenDoublingEffects(sentence Sentence, tokens []shared.Token, atoms Atoms) ([]EffectSyntax, bool) {
	commaIndex, anyController, ok := matchPassiveTokenDoubling(tokens)
	if !ok {
		return nil, false
	}
	condition := tokens[:commaIndex]
	resolving := tokens[commaIndex+1:]
	createdIndex := commaIndex - 4
	if anyController {
		createdIndex = commaIndex - 1
	}
	createEffect := EffectSyntax{
		Kind:             EffectCreate,
		Context:          EffectContextController,
		Span:             shared.SpanOf(condition),
		VerbSpan:         tokens[createdIndex].Span,
		ClauseSpan:       shared.SpanOf(condition),
		Text:             sentence.Text,
		Tokens:           append([]shared.Token(nil), condition...),
		Amount:           EffectAmountSyntax{Value: 1, Known: true},
		UnderYourControl: !anyController,
	}
	doubledIndex := commaIndex + 8
	doubledEffect := EffectSyntax{
		Kind:       EffectCreate,
		Context:    EffectContextReferencedObject,
		Span:       shared.SpanOf(resolving),
		VerbSpan:   tokens[doubledIndex].Span,
		ClauseSpan: shared.SpanOf(resolving),
		Text:       sentence.Text,
		Tokens:     append([]shared.Token(nil), resolving...),
		Amount:     EffectAmountSyntax{Value: 2, Known: true},
		Replacement: EffectReplacementSyntax{
			Kind: EffectReplacementTwiceThatMany,
			Span: tokens[len(tokens)-2].Span,
		},
		References: referencesInSpan(atoms, shared.SpanOf(resolving)),
	}
	return []EffectSyntax{createEffect, doubledEffect}, true
}

// matchPassiveTokenDoubling reports the index of the comma separating the
// would-create condition clause from the doubled output clause when tokens spell
// the passive token-doubling replacement. The controller-only wording ("...under
// your control, ...", Mondrak) and the controller-agnostic wording ("...would be
// created, ...", Primal Vigor) are distinguished by anyController, which the
// any-player runtime node honors via an any-player scope.
func matchPassiveTokenDoubling(tokens []shared.Token) (commaIndex int, anyController, ok bool) {
	if len(tokens) == 22 &&
		effectWordsAt(tokens, 0, "if", "one", "or", "more", "tokens", "would", "be", "created") &&
		effectWordsAt(tokens, 8, "under", "your", "control") &&
		tokens[11].Kind == shared.Comma &&
		effectWordsAt(tokens, 12, "twice", "that", "many", "of", "those", "tokens", "are", "created", "instead") &&
		tokens[21].Kind == shared.Period {
		return 11, false, true
	}
	if len(tokens) == 19 &&
		effectWordsAt(tokens, 0, "if", "one", "or", "more", "tokens", "would", "be", "created") &&
		tokens[8].Kind == shared.Comma &&
		effectWordsAt(tokens, 9, "twice", "that", "many", "of", "those", "tokens", "are", "created", "instead") &&
		tokens[18].Kind == shared.Period {
		return 8, true, true
	}
	return 0, false, false
}

// parsePassiveTokenAdditiveEffects recognizes the passive-voice additive
// token-creation replacement "If one or more [type] tokens would be created
// under your control, those tokens plus <addend> are created instead."
// (Peregrin Took: "... plus an additional Food token ..."; Donatello, the
// Brains: "... plus a Mutagen token ..."; Stridehangar Automaton: "If one or
// more artifact tokens would be created under your control, those tokens plus an
// additional 1/1 colorless Thopter artifact creature token with flying are
// created instead."). Its active-voice equivalent "If you would create one or
// more [type] tokens, instead create those tokens plus an additional <addend>."
// (Worldwalker Helm, Xorn) parses through the ordinary create-verb path. The
// passive wording carries no active "create" verb, so it is recognized here and
// emitted as the same two EffectCreate instructions the active form produces:
// the would-create group (carrying the optional card-type filter in its
// selector) and the addend output marked EffectReplacementPlusAdditional. The
// matching intervening-if condition is recognized separately by
// recognizeTokenCreationCondition.
func parsePassiveTokenAdditiveEffects(sentence Sentence, tokens []shared.Token, atoms Atoms) ([]EffectSyntax, bool) {
	commaIndex, anyController, ok := matchPassiveTokenAdditive(tokens)
	if !ok {
		return nil, false
	}
	condition := tokens[:commaIndex]
	resolving := tokens[commaIndex+1:]
	addendAmount, ok := passiveTokenAddendAmount(resolving, atoms)
	if !ok {
		return nil, false
	}
	// The would-create noun phrase ("one or more [type] tokens") is the
	// condition with its leading "if" and trailing "would be created [under your
	// control]" stripped. parseSelection over just the noun phrase keeps the
	// "would be created" verb words and the controller clause out of the
	// selection, leaving the optional card-type filter the lowering reads.
	nounPhrase := condition
	if len(nounPhrase) > 0 && equalWord(nounPhrase[0], "if") {
		nounPhrase = nounPhrase[1:]
	}
	if trimmed, stripped := stripTokenSuffix(nounPhrase, "would", "be", "created", "under", "your", "control"); stripped {
		nounPhrase = trimmed
	} else if trimmed, stripped := stripTokenSuffix(nounPhrase, "would", "be", "created"); stripped {
		nounPhrase = trimmed
	}
	createdIndex := commaIndex - 1
	for i := range condition {
		if equalWord(condition[i], "created") {
			createdIndex = i
			break
		}
	}
	createEffect := EffectSyntax{
		Kind:             EffectCreate,
		Context:          EffectContextController,
		Span:             shared.SpanOf(condition),
		VerbSpan:         tokens[createdIndex].Span,
		ClauseSpan:       shared.SpanOf(condition),
		Text:             sentence.Text,
		Tokens:           append([]shared.Token(nil), condition...),
		Selection:        parseSelection(nounPhrase, atoms),
		Amount:           EffectAmountSyntax{Value: 1, Known: true},
		UnderYourControl: !anyController,
	}
	if conjunctiveTypeTarget(createEffect.Selection) {
		createEffect.Selection.ConjunctiveTypes = true
	}
	// The addend clause ("those tokens plus <addend>") is the resolving clause
	// with its trailing "are created instead ." stripped. It is parsed by the
	// same token-characteristic helpers the active create-verb path uses, so the
	// addend's subtypes, colors, power/toughness, keyword, and name match the
	// active form exactly.
	addendClause := resolving[:len(resolving)-4]
	tokenPower, tokenToughness, tokenPTKnown := parseTokenPowerToughness(EffectCreate, addendClause)
	plusSpan := resolving[0].Span
	for i := range resolving {
		if equalWord(resolving[i], "plus") {
			plusSpan = resolving[i].Span
			break
		}
	}
	createdResolvingIndex := len(resolving) - 3
	addendEffect := EffectSyntax{
		Kind:                EffectCreate,
		Context:             EffectContextReferencedObject,
		Span:                shared.SpanOf(resolving),
		VerbSpan:            resolving[createdResolvingIndex].Span,
		ClauseSpan:          shared.SpanOf(resolving),
		Text:                sentence.Text,
		Tokens:              append([]shared.Token(nil), resolving...),
		Selection:           parseSelection(addendClause, atoms),
		Amount:              EffectAmountSyntax{Value: 1, Known: true},
		TokenPower:          tokenPower,
		TokenToughness:      tokenToughness,
		TokenPTKnown:        tokenPTKnown,
		TokenKeywords:       parseTokenKeywords(EffectCreate, addendClause, atoms),
		TokenToxic:          parseTokenKeywordToxic(EffectCreate, addendClause, atoms),
		TokenName:           parseTokenName(EffectCreate, addendClause),
		TokenPredefinedName: parsePredefinedTokenName(EffectCreate, addendClause),
		Replacement: EffectReplacementSyntax{
			Kind:   EffectReplacementPlusAdditional,
			Amount: addendAmount,
			Span:   plusSpan,
		},
		References: referencesInSpan(atoms, shared.SpanOf(resolving)),
	}
	if conjunctiveTypeTarget(addendEffect.Selection) {
		addendEffect.Selection.ConjunctiveTypes = true
	}
	return []EffectSyntax{createEffect, addendEffect}, true
}

// matchPassiveTokenAdditive reports the index of the comma separating the
// would-create condition clause from the additive output clause when tokens
// spell the passive additive token-creation replacement. The controller-only
// wording ("...under your control, ...") and the controller-agnostic wording
// ("...would be created, ...") are distinguished by anyController. The optional
// card-type word(s) between "more" and "tokens" are tolerated here and carried
// downstream by the would-create group's selector, mirroring the active form.
func matchPassiveTokenAdditive(tokens []shared.Token) (commaIndex int, anyController, ok bool) {
	if !effectWordsAt(tokens, 0, "if", "one", "or", "more") {
		return 0, false, false
	}
	for i := range tokens {
		if tokens[i].Kind != shared.Comma || !effectWordsAt(tokens, i+1, "those", "tokens", "plus") {
			continue
		}
		body := tokens[1:i]
		controllerOnly := false
		if _, stripped := stripTokenSuffix(body, "tokens", "would", "be", "created", "under", "your", "control"); stripped {
			controllerOnly = true
		} else if _, stripped := stripTokenSuffix(body, "tokens", "would", "be", "created"); !stripped {
			continue
		}
		if !effectWordsAt(body, 0, "one", "or", "more") {
			continue
		}
		resolving := tokens[i+1:]
		n := len(resolving)
		if n < 4 ||
			!effectWordsAt(resolving, n-4, "are", "created", "instead") ||
			resolving[n-1].Kind != shared.Period {
			continue
		}
		return i, !controllerOnly, true
	}
	return 0, false, false
}

// passiveTokenAddendAmount returns the number of additional tokens an additive
// token-creation replacement creates, read from the "those tokens plus <amount>
// ..." clause. The "a"/"an" article and the "an additional" rider both denote a
// single extra token; an explicit number ("two additional ...") denotes that
// many. Any other wording fails closed.
func passiveTokenAddendAmount(tokens []shared.Token, atoms Atoms) (int, bool) {
	for i := range tokens {
		if !equalWord(tokens[i], "plus") || i+1 >= len(tokens) {
			continue
		}
		next := tokens[i+1]
		if equalWord(next, "a") || equalWord(next, "an") {
			return 1, true
		}
		if amount, ok := effectNumber(next, atoms); ok {
			return amount, true
		}
		return 0, false
	}
	return 0, false
}

// parseDrawEmptyLibraryWinReplacement recognizes the draw-from-empty-library win
// replacement "If you would draw a card while your library has no cards in it,
// you win the game instead." (Laboratory Maniac, Jace, Wielder of Mysteries) and
// emits a single win-the-game effect for the resolving clause. The matching
// intervening-if condition is recognized separately by
// recognizeDrawFromEmptyLibraryCondition; the runtime replacement is otherwise
// self-contained, so the would-draw clause needs no effect of its own.
func parseDrawEmptyLibraryWinReplacement(sentence Sentence, tokens []shared.Token, atoms Atoms) ([]EffectSyntax, bool) {
	commaIndex, ok := matchDrawEmptyLibraryWin(tokens)
	if !ok {
		return nil, false
	}
	resolving := tokens[commaIndex+1:]
	winIndex := commaIndex + 2
	return []EffectSyntax{{
		Kind:       EffectWinGame,
		Context:    EffectContextController,
		Span:       shared.SpanOf(tokens),
		VerbSpan:   tokens[winIndex].Span,
		ClauseSpan: shared.SpanOf(tokens),
		Text:       sentence.Text,
		Tokens:     append([]shared.Token(nil), resolving...),
		Replacement: EffectReplacementSyntax{
			Kind: EffectReplacementInstead,
			Span: tokens[len(tokens)-2].Span,
		},
		References: referencesInSpan(atoms, shared.SpanOf(resolving)),
		Exact:      true,
	}}, true
}

// parsePunisherEachLoseLifeEffect recognizes the "punisher" family ("<group>
// punisherUnlessClauseAt reports whether tokens[i:] begins the punisher idiom
// "unless that player sacrifices ... [or discards ...]". This per-player
// alternative-cost clause is consumed wholesale by parsePunisherEachLoseLifeEffect
// and must not be parsed as a game-state condition. It is distinguished from the
// "unless that player pays ..." payment idiom by the sacrifice/discard verb.
func punisherUnlessClauseAt(tokens []shared.Token, i int) bool {
	if !effectWordsAt(tokens, i, "unless", "that", "player") || i+3 >= len(tokens) {
		return false
	}
	return equalWord(tokens[i+3], "sacrifices") || equalWord(tokens[i+3], "discards")
}

// loses N life unless that player sacrifices a <permanent> [of their choice]
// [or discards a card].") as a single EffectPunisherLoseLife effect. The group
// must be each-opponent / each-player / each-other-player; the alternatives are
// a filtered sacrifice and/or a discard-a-card, joined by "or" in either order.
func parsePunisherEachLoseLifeEffect(sentence Sentence, tokens []shared.Token, atoms Atoms) ([]EffectSyntax, bool) {
	verbIndex := -1
	for i, token := range tokens {
		if equalWord(token, "loses") || equalWord(token, "lose") {
			verbIndex = i
			break
		}
	}
	if verbIndex < 0 || verbIndex+2 >= len(tokens) {
		return nil, false
	}
	context := effectContextAt(tokens, verbIndex, atoms)
	switch context {
	case EffectContextEachOpponent, EffectContextEachPlayer, EffectContextEachOtherPlayer:
	default:
		return nil, false
	}
	amount, ok := effectNumber(tokens[verbIndex+1], atoms)
	if !ok || amount < 1 || !equalWord(tokens[verbIndex+2], "life") {
		return nil, false
	}
	rest := tokens[verbIndex+3:]
	if len(rest) < 4 ||
		!equalWord(rest[0], "unless") ||
		!equalWord(rest[1], "that") ||
		!equalWord(rest[2], "player") {
		return nil, false
	}
	options := rest[3:]
	if n := len(options); n > 0 && options[n-1].Kind == shared.Period {
		options = options[:n-1]
	}
	segments := splitPunisherOptions(options)
	if len(segments) == 0 || len(segments) > 2 {
		return nil, false
	}
	effect := EffectSyntax{
		Kind:       EffectPunisherLoseLife,
		Context:    context,
		Span:       shared.SpanOf(tokens),
		VerbSpan:   tokens[verbIndex].Span,
		ClauseSpan: shared.SpanOf(tokens),
		Text:       sentence.Text,
		Tokens:     append([]shared.Token(nil), tokens...),
		Amount:     EffectAmountSyntax{Value: amount, Known: true, Span: tokens[verbIndex+1].Span},
		Exact:      true,
	}
	for _, segment := range segments {
		if len(segment) == 0 {
			return nil, false
		}
		switch {
		case equalWord(segment[0], "sacrifices") || equalWord(segment[0], "sacrifice"):
			if effect.PunisherSacrifice {
				return nil, false
			}
			selectionTokens := stripOfTheirChoice(segment[1:])
			if len(selectionTokens) == 0 {
				return nil, false
			}
			selection := parseSelection(selectionTokens, atoms)
			effect.PunisherSacrifice = true
			effect.Selection = selection
		case equalWord(segment[0], "discards") || equalWord(segment[0], "discard"):
			if effect.PunisherDiscard || !punisherDiscardACard(segment[1:]) {
				return nil, false
			}
			effect.PunisherDiscard = true
		default:
			return nil, false
		}
	}
	if !effect.PunisherSacrifice && !effect.PunisherDiscard {
		return nil, false
	}
	return []EffectSyntax{effect}, true
}

// splitPunisherOptions splits a punisher's avoidance clause on its top-level
// "or" connectives, returning each option's tokens.
func splitPunisherOptions(tokens []shared.Token) [][]shared.Token {
	var segments [][]shared.Token
	start := 0
	for i, token := range tokens {
		if equalWord(token, "or") {
			segments = append(segments, tokens[start:i])
			start = i + 1
		}
	}
	return append(segments, tokens[start:])
}

// stripOfTheirChoice drops a trailing "of their choice" qualifier from a
// sacrifice selection's tokens ("a nonland permanent of their choice").
func stripOfTheirChoice(tokens []shared.Token) []shared.Token {
	if n := len(tokens); n >= 3 &&
		equalWord(tokens[n-3], "of") &&
		equalWord(tokens[n-2], "their") &&
		equalWord(tokens[n-1], "choice") {
		return tokens[:n-3]
	}
	return tokens
}

// punisherDiscardACard reports whether tokens spell the "a card" object of a
// punisher's discard alternative.
func punisherDiscardACard(tokens []shared.Token) bool {
	return len(tokens) == 2 && equalWord(tokens[0], "a") && equalWord(tokens[1], "card")
}

// matchDrawEmptyLibraryWin reports the index of the comma separating the
// would-draw condition clause from the win-the-game result when tokens spell the
// draw-from-empty-library win replacement.
func matchDrawEmptyLibraryWin(tokens []shared.Token) (int, bool) {
	if len(tokens) != 21 ||
		!effectWordsAt(tokens, 0, "if", "you", "would", "draw", "a", "card") ||
		!effectWordsAt(tokens, 6, "while", "your", "library", "has", "no", "cards", "in", "it") ||
		tokens[14].Kind != shared.Comma ||
		!effectWordsAt(tokens, 15, "you", "win", "the", "game", "instead") ||
		tokens[20].Kind != shared.Period {
		return 0, false
	}
	return 14, true
}

// parseDrawDoublingReplacement recognizes the draw-doubling replacement "If you
// would draw a card[ except the first one you draw in each of your draw steps],
// draw <N> cards instead." (Thought Reflection, Teferi's Ageless Insight) and
// emits a single draw effect carrying the replacement multiplier N for the
// resolving clause. The matching intervening-if condition is recognized
// separately by recognizeDrawCardReplacementCondition.
func parseDrawDoublingReplacement(sentence Sentence, tokens []shared.Token, atoms Atoms) ([]EffectSyntax, bool) {
	commaIndex, amount, ok := matchDrawDoubling(tokens)
	if !ok {
		return nil, false
	}
	resolving := tokens[commaIndex+1:]
	drawIndex := commaIndex + 1
	return []EffectSyntax{{
		Kind:       EffectDraw,
		Context:    EffectContextController,
		Span:       shared.SpanOf(tokens),
		VerbSpan:   tokens[drawIndex].Span,
		ClauseSpan: shared.SpanOf(tokens),
		Text:       sentence.Text,
		Tokens:     append([]shared.Token(nil), resolving...),
		Amount:     EffectAmountSyntax{Value: amount, Known: true},
		Replacement: EffectReplacementSyntax{
			Kind: EffectReplacementInstead,
			Span: tokens[len(tokens)-2].Span,
		},
		References: referencesInSpan(atoms, shared.SpanOf(resolving)),
		Exact:      true,
	}}, true
}

// parseDrawReplacementDig recognizes the draw-replacement dig "If you would draw
// a card, instead look at the top <N> cards of your library, then put <M> [of
// them|of those cards] into your hand and the <rest|other> <remainder>."
// (Underrealm Lich) and emits a single dig effect carrying the look count N (in
// Amount), the take count M and remainder destination (in Dig), and the instead
// replacement. The matching intervening-if condition is recognized separately by
// recognizeDrawCardReplacementCondition.
func parseDrawReplacementDig(sentence Sentence, tokens []shared.Token, atoms Atoms) ([]EffectSyntax, bool) {
	look, dig, ok := matchDrawReplacementDig(tokens)
	if !ok {
		return nil, false
	}
	// The would-draw condition is "if you would draw a card", so the separating
	// comma is token index 6 and the resolving "instead ..." clause follows it.
	const commaIndex = 6
	resolving := tokens[commaIndex+1:]
	return []EffectSyntax{{
		Kind:       EffectDig,
		Context:    EffectContextController,
		Span:       shared.SpanOf(tokens),
		VerbSpan:   tokens[commaIndex+2].Span,
		ClauseSpan: shared.SpanOf(tokens),
		Text:       sentence.Text,
		Tokens:     append([]shared.Token(nil), resolving...),
		Amount:     EffectAmountSyntax{Value: look, Known: true},
		Dig:        dig,
		Replacement: EffectReplacementSyntax{
			Kind: EffectReplacementInstead,
			Span: tokens[commaIndex+1].Span,
		},
		References: referencesInSpan(atoms, shared.SpanOf(resolving)),
		Exact:      true,
	}}, true
}

// matchDrawReplacementDig reports the look count and structured dig fields when
// tokens spell the draw-replacement dig "if you would draw a card, instead look
// at the top <N> cards of your library, then put <M> ... instead". The take
// count must be at least one and fewer than the look count, mirroring the
// impulse dig's look-exceeds-take requirement; every other wording fails closed.
func matchDrawReplacementDig(tokens []shared.Token) (look int, dig DigSyntax, ok bool) {
	if len(tokens) < 6 || tokens[len(tokens)-1].Kind != shared.Period {
		return 0, DigSyntax{}, false
	}
	if !effectWordsAt(tokens, 0, "if", "you", "would", "draw", "a", "card") ||
		len(tokens) <= 6 || tokens[6].Kind != shared.Comma {
		return 0, DigSyntax{}, false
	}
	if !effectWordsAt(tokens, 7, "instead", "look", "at", "the", "top") {
		return 0, DigSyntax{}, false
	}
	i := 12
	if i >= len(tokens) {
		return 0, DigSyntax{}, false
	}
	lookCount, lookOK := CardinalWordValue(tokens[i].Text)
	if !lookOK || lookCount < 2 {
		return 0, DigSyntax{}, false
	}
	i++
	if !effectWordsAt(tokens, i, "cards", "of", "your", "library") {
		return 0, DigSyntax{}, false
	}
	i += 4
	if i >= len(tokens) || tokens[i].Kind != shared.Comma {
		return 0, DigSyntax{}, false
	}
	i++
	if !effectWordsAt(tokens, i, "then", "put") {
		return 0, DigSyntax{}, false
	}
	i += 2
	if i >= len(tokens) {
		return 0, DigSyntax{}, false
	}
	takeCount, takeOK := CardinalWordValue(tokens[i].Text)
	if !takeOK || takeCount < 1 || takeCount >= lookCount {
		return 0, DigSyntax{}, false
	}
	i++
	switch {
	case effectWordsAt(tokens, i, "of", "them"):
		dig.Source = DigSourceThem
		i += 2
	case effectWordsAt(tokens, i, "of", "those", "cards"):
		dig.Source = DigSourceThoseCards
		i += 3
	default:
		dig.Source = DigSourceNone
	}
	if !effectWordsAt(tokens, i, "into", "your", "hand", "and", "the") {
		return 0, DigSyntax{}, false
	}
	i += 5
	switch {
	case effectWordsAt(tokens, i, "other"):
		dig.Singular = true
		i++
	case effectWordsAt(tokens, i, "rest"):
		i++
	default:
		return 0, DigSyntax{}, false
	}
	remainder, after, remainderOK := digRemainderAt(tokens, i)
	if !remainderOK || after != len(tokens)-1 {
		return 0, DigSyntax{}, false
	}
	dig.Remainder = remainder
	dig.Take = takeCount
	dig.Put = true
	return lookCount, dig, true
}

// matchDrawDoubling reports the comma index separating the would-draw condition
// from the "draw <N> cards instead" result and the multiplier N when tokens
// spell a draw-doubling replacement. The condition must be the plain would-draw
// or the draw-step exception form, and the result an N (>= 2) card draw.
func matchDrawDoubling(tokens []shared.Token) (commaIndex, amount int, ok bool) {
	if len(tokens) < 6 || tokens[len(tokens)-1].Kind != shared.Period {
		return 0, 0, false
	}
	if !effectWordsAt(tokens, 0, "if", "you", "would", "draw", "a", "card") {
		return 0, 0, false
	}
	comma := -1
	for i := range tokens {
		if tokens[i].Kind == shared.Comma {
			comma = i
			break
		}
	}
	if comma < 0 || !drawDoublingConditionBody(tokens[1:comma]) {
		return 0, 0, false
	}
	result := tokens[comma+1 : len(tokens)-1]
	if len(result) != 4 ||
		!equalWord(result[0], "draw") ||
		!equalWord(result[2], "cards") ||
		!equalWord(result[3], "instead") {
		return 0, 0, false
	}
	n, valueOK := CardinalWordValue(result[1].Text)
	if !valueOK || n < 2 {
		return 0, 0, false
	}
	return comma, n, true
}

// drawDoublingConditionBody reports whether a draw-doubling condition body is one
// of the two supported forms: the plain "you would draw a card" or the draw-step
// exception "you would draw a card except the first one you draw in each of your
// draw steps".
func drawDoublingConditionBody(body []shared.Token) bool {
	if tokenWordsEqual(body, "you", "would", "draw", "a", "card") {
		return true
	}
	return tokenWordsEqual(body,
		"you", "would", "draw", "a", "card",
		"except", "the", "first", "one", "you", "draw",
		"in", "each", "of", "your", "draw", "steps")
}

// parseLifeGainReplacement recognizes the life-gain replacement "If you would
// gain life, you gain twice that much life instead." (multiplier two) and "If
// you would gain life, you gain that much life plus N instead." (additive
// bonus), emitting a single gain effect carrying the replacement so the
// would-gain condition clause does not become a spurious effect of its own. The
// matching intervening-if condition is recognized separately by
// recognizeLifeGainCondition.
func parseLifeGainReplacement(sentence Sentence, tokens []shared.Token, atoms Atoms) ([]EffectSyntax, bool) {
	commaIndex, replacement, ok := matchLifeGainReplacement(tokens, atoms)
	if !ok {
		return nil, false
	}
	resolving := tokens[commaIndex+1:]
	gainIndex := commaIndex + 1
	replacement.Span = tokens[len(tokens)-2].Span
	return []EffectSyntax{{
		Kind:        EffectGain,
		Context:     EffectContextController,
		Span:        shared.SpanOf(tokens),
		VerbSpan:    tokens[gainIndex].Span,
		ClauseSpan:  shared.SpanOf(tokens),
		Text:        sentence.Text,
		Tokens:      append([]shared.Token(nil), resolving...),
		Replacement: replacement,
		References:  referencesInSpan(atoms, shared.SpanOf(resolving)),
		Exact:       true,
	}}, true
}

// matchLifeGainReplacement reports the comma index separating the would-gain
// condition from its "you gain ... life instead" result and the replacement
// (twice-that-much or that-much-plus-N) when tokens spell a life-gain
// replacement.
func matchLifeGainReplacement(tokens []shared.Token, atoms Atoms) (commaIndex int, replacement EffectReplacementSyntax, ok bool) {
	if len(tokens) < 6 || tokens[len(tokens)-1].Kind != shared.Period {
		return 0, EffectReplacementSyntax{}, false
	}
	if !effectWordsAt(tokens, 0, "if") {
		return 0, EffectReplacementSyntax{}, false
	}
	comma := -1
	for i := range tokens {
		if tokens[i].Kind == shared.Comma {
			comma = i
			break
		}
	}
	if comma < 0 || !tokenWordsEqual(tokens[1:comma], "you", "would", "gain", "life") {
		return 0, EffectReplacementSyntax{}, false
	}
	result := tokens[comma+1 : len(tokens)-1]
	if tokenWordsEqual(result, "you", "gain", "twice", "that", "much", "life", "instead") {
		return comma, EffectReplacementSyntax{Kind: EffectReplacementTwiceThatMuch}, true
	}
	if len(result) == 8 &&
		tokenWordsEqual(result[:6], "you", "gain", "that", "much", "life", "plus") &&
		equalWord(result[7], "instead") {
		if n, valueOK := effectNumber(result[6], atoms); valueOK && n > 0 {
			return comma, EffectReplacementSyntax{Kind: EffectReplacementThatMuchPlus, Amount: n}, true
		}
	}
	return 0, EffectReplacementSyntax{}, false
}

// parseLeaveBattlefieldExileReplacement recognizes the leaves-the-battlefield
// self-replacement "If it would leave the battlefield, exile it instead of
// putting it anywhere else." (Whip of Erebos) and its "this <type>" self-applied
// and shorter "...exile it instead." variants, emitting a single
// EffectExileIfLeaveBattlefield effect so the leading would-leave condition does
// not become a spurious activation/intervening condition of its own. The
// matching condition boundary is suppressed by
// conditionLeaveBattlefieldExileReplacementAt.
func parseLeaveBattlefieldExileReplacement(sentence Sentence, tokens []shared.Token, atoms Atoms) ([]EffectSyntax, bool) {
	context, ok := matchLeaveBattlefieldExileReplacement(tokens)
	if !ok {
		return nil, false
	}
	return []EffectSyntax{{
		Kind:       EffectExileIfLeaveBattlefield,
		Context:    context,
		Span:       shared.SpanOf(tokens),
		ClauseSpan: shared.SpanOf(tokens),
		Text:       sentence.Text,
		Tokens:     append([]shared.Token(nil), tokens...),
		References: referencesInSpan(atoms, shared.SpanOf(tokens)),
		Exact:      true,
	}}, true
}

// matchLeaveBattlefieldExileReplacement reports the effect context (referenced
// object for "it", source for "this <type>") when tokens spell the
// leaves-the-battlefield exile replacement "If <subject> would leave the
// battlefield, exile it instead[ of putting it anywhere else]."
func matchLeaveBattlefieldExileReplacement(tokens []shared.Token) (EffectContextKind, bool) {
	if len(tokens) < 9 || tokens[len(tokens)-1].Kind != shared.Period || !equalWord(tokens[0], "if") {
		return EffectContextUnknown, false
	}
	rest := tokens[1:]
	subjectWidth := leaveBattlefieldReplacementSubjectWidth(rest)
	if subjectWidth == 0 {
		return EffectContextUnknown, false
	}
	idx := 1 + subjectWidth
	if !effectWordsAt(tokens, idx, "would", "leave", "the", "battlefield") {
		return EffectContextUnknown, false
	}
	idx += 4
	if idx >= len(tokens) || tokens[idx].Kind != shared.Comma {
		return EffectContextUnknown, false
	}
	result := tokens[idx+1 : len(tokens)-1]
	if !leaveBattlefieldReplacementResult(result) {
		return EffectContextUnknown, false
	}
	if equalWord(rest[0], "it") {
		return EffectContextReferencedObject, true
	}
	return EffectContextSource, true
}

// leaveBattlefieldReplacementResult reports whether result spells the redirect
// "exile it instead" or "exile it instead of putting it anywhere else".
func leaveBattlefieldReplacementResult(result []shared.Token) bool {
	return tokenWordsEqual(result, "exile", "it", "instead") ||
		tokenWordsEqual(result, "exile", "it", "instead", "of", "putting", "it", "anywhere", "else")
}

// parseDieThisTurnExileReplacement recognizes the single-target damage-spell
// rider "If that creature [or planeswalker] would die this turn, exile it
// instead." (Lava Coil, Obliterating Bolt, Magma Spray, Flame-Blessed Bolt,
// Bleed Dry, ...) and emits a single EffectExileIfWouldDieThisTurn effect so the
// leading would-die condition does not become a spurious activation/intervening
// condition of its own. The matching condition boundary is suppressed by
// conditionDieThisTurnExileReplacementAt. The subject ("that creature", "that
// creature or planeswalker", or "it") and the result's "it" are carried as
// references that bind to the spell's single target.
func parseDieThisTurnExileReplacement(sentence Sentence, tokens []shared.Token, atoms Atoms) ([]EffectSyntax, bool) {
	if !matchDieThisTurnExileReplacement(tokens) {
		return nil, false
	}
	return []EffectSyntax{{
		Kind:       EffectExileIfWouldDieThisTurn,
		Context:    EffectContextController,
		Span:       shared.SpanOf(tokens),
		ClauseSpan: shared.SpanOf(tokens),
		Text:       sentence.Text,
		Tokens:     append([]shared.Token(nil), tokens...),
		References: referencesInSpan(atoms, shared.SpanOf(tokens)),
		Exact:      true,
	}}, true
}

// matchDieThisTurnExileReplacement reports whether tokens spell the would-die
// exile rider "If <subject> would die this turn, exile it instead." where the
// subject is "that creature", "that creature or planeswalker", or "it".
func matchDieThisTurnExileReplacement(tokens []shared.Token) bool {
	if len(tokens) < 9 || tokens[len(tokens)-1].Kind != shared.Period || !equalWord(tokens[0], "if") {
		return false
	}
	subjectWidth := dieThisTurnExileSubjectWidth(tokens[1:])
	if subjectWidth == 0 {
		return false
	}
	idx := 1 + subjectWidth
	if !effectWordsAt(tokens, idx, "would", "die", "this", "turn") {
		return false
	}
	idx += 4
	if idx >= len(tokens) || tokens[idx].Kind != shared.Comma {
		return false
	}
	return tokenWordsEqual(tokens[idx+1:len(tokens)-1], "exile", "it", "instead")
}

// dieThisTurnExileSubjectWidth reports the token width of the subject that opens
// a would-die exile rider ("it" → 1, "that creature" → 2, "that creature or
// planeswalker" → 4), or 0 when tokens do not begin with such a subject.
func dieThisTurnExileSubjectWidth(tokens []shared.Token) int {
	if len(tokens) == 0 {
		return 0
	}
	if equalWord(tokens[0], "it") {
		return 1
	}
	if len(tokens) >= 4 && tokenWordsEqual(tokens[:4], "that", "creature", "or", "planeswalker") {
		return 4
	}
	if len(tokens) >= 2 && tokenWordsEqual(tokens[:2], "that", "creature") {
		return 2
	}
	return 0
}

// parseLifeLossReplacement recognizes the life-loss replacement "If an opponent
// would lose life during your turn, they lose twice that much life instead."
// (Bloodletter of Aclazotz) and its untimed/any-player generalizations,
// emitting a single lose effect carrying the replacement so the would-lose
// condition clause does not become a spurious effect of its own. The matching
// intervening-if condition is recognized separately by recognizeLifeLossCondition.
func parseLifeLossReplacement(sentence Sentence, tokens []shared.Token, atoms Atoms) ([]EffectSyntax, bool) {
	commaIndex, replacement, ok := matchLifeLossReplacement(tokens, atoms)
	if !ok {
		return nil, false
	}
	resolving := tokens[commaIndex+1:]
	loseIndex := commaIndex + 1
	replacement.Span = tokens[len(tokens)-2].Span
	return []EffectSyntax{{
		Kind:        EffectLose,
		Context:     EffectContextController,
		Span:        shared.SpanOf(tokens),
		VerbSpan:    tokens[loseIndex].Span,
		ClauseSpan:  shared.SpanOf(tokens),
		Text:        sentence.Text,
		Tokens:      append([]shared.Token(nil), resolving...),
		Replacement: replacement,
		References:  referencesInSpan(atoms, shared.SpanOf(resolving)),
		Exact:       true,
	}}, true
}

// parseExcessDamageToControllerEffect recognizes the standalone follow-on
// sentence "Excess damage is dealt to that creature's controller instead."
// (Flame Spill, Pigment Storm, Gandalf's Sanction). It redirects the overflow of
// the preceding "deals N damage to target creature" clause — the damage beyond
// what was needed to destroy the targeted creature — onto that creature's
// controller (or owner). It emits a deal-damage effect whose amount is the
// excess damage dealt this way and whose recipient is the prior target's
// controller/owner, so the pair lowers to two ordered Damage instructions: the
// targeted creature damage publishes its excess, and this instruction deals that
// excess to its controller. It fails closed for every other wording.
func parseExcessDamageToControllerEffect(sentence Sentence, tokens []shared.Token, atoms Atoms) ([]EffectSyntax, bool) {
	clause := tokens
	if n := len(clause); n > 0 && clause[n-1].Kind == shared.Period {
		clause = clause[:n-1]
	}
	if len(clause) < 7 ||
		!effectWordsAt(clause, 0, "excess", "damage", "is", "dealt", "to") ||
		!equalWord(clause[len(clause)-1], "instead") {
		return nil, false
	}
	recipient := clause[5 : len(clause)-1]
	role, ok := referencedControllerOwnerRecipient(recipient)
	if !ok {
		return nil, false
	}
	synthesized := append([]shared.Token{clause[1], clause[4]}, recipient...)
	return []EffectSyntax{{
		Kind:            EffectDealDamage,
		Context:         EffectContextSource,
		Span:            shared.SpanOf(tokens),
		VerbSpan:        clause[3].Span,
		ClauseSpan:      shared.SpanOf(tokens),
		Text:            sentence.Text,
		Tokens:          synthesized,
		Amount:          EffectAmountSyntax{DynamicKind: EffectDynamicAmountExcessDamageDealtThisWay, Multiplier: 1},
		DamageRecipient: DamageRecipientSyntax{Reference: role},
		References:      referencesInSpan(atoms, shared.SpanOf(recipient)),
		Exact:           true,
	}}, true
}

// matchLifeLossReplacement reports the comma index separating the would-lose
// condition from its "they lose ... instead" result and the replacement
// (twice-that-much or that-much-plus-N) when tokens spell a life-loss
// replacement gated by one of the recognized would-lose conditions.
func matchLifeLossReplacement(tokens []shared.Token, atoms Atoms) (commaIndex int, replacement EffectReplacementSyntax, ok bool) {
	if len(tokens) < 6 || tokens[len(tokens)-1].Kind != shared.Period {
		return 0, EffectReplacementSyntax{}, false
	}
	if !effectWordsAt(tokens, 0, "if") {
		return 0, EffectReplacementSyntax{}, false
	}
	comma := -1
	for i := range tokens {
		if tokens[i].Kind == shared.Comma {
			comma = i
			break
		}
	}
	if comma < 0 {
		return 0, EffectReplacementSyntax{}, false
	}
	condition := tokens[1:comma]
	if !tokenWordsEqual(condition, "an", "opponent", "would", "lose", "life", "during", "your", "turn") &&
		!tokenWordsEqual(condition, "an", "opponent", "would", "lose", "life") &&
		!tokenWordsEqual(condition, "a", "player", "would", "lose", "life") {
		return 0, EffectReplacementSyntax{}, false
	}
	result := tokens[comma+1 : len(tokens)-1]
	if tokenWordsEqual(result, "they", "lose", "twice", "that", "much", "life", "instead") ||
		tokenWordsEqual(result, "they", "lose", "twice", "that", "much", "instead") {
		return comma, EffectReplacementSyntax{Kind: EffectReplacementTwiceThatMuch}, true
	}
	if len(result) == 8 &&
		tokenWordsEqual(result[:6], "they", "lose", "that", "much", "life", "plus") &&
		equalWord(result[7], "instead") {
		if n, valueOK := effectNumber(result[6], atoms); valueOK && n > 0 {
			return comma, EffectReplacementSyntax{Kind: EffectReplacementThatMuchPlus, Amount: n}, true
		}
	}
	return 0, EffectReplacementSyntax{}, false
}

func recognizeImpulseExileSequence(sentences []Sentence) bool {
	// Trailing reminder text ("(If you cast a spell this way…)") is parsed as
	// its own parenthesized sentence; it carries no game meaning and is excluded
	// from coverage, so ignore it when matching the two-sentence impulse shape.
	for len(sentences) > 2 && isReminderSentence(sentences[len(sentences)-1]) {
		sentences = sentences[:len(sentences)-1]
	}
	if len(sentences) != 2 {
		return false
	}
	amount, variableX, ok := matchImpulseExileClause(strings.TrimSpace(sentences[0].Text))
	if !ok {
		return false
	}
	// "the top X cards" agrees with the plural play-permission demonstratives
	// ("those cards"/"them"), so resolve the object phrase as a plural.
	objectAmount := amount
	if variableX {
		objectAmount = 2
	}
	duration, ok := matchImpulsePlayPermissionClause(strings.TrimSpace(sentences[1].Text), objectAmount)
	if !ok {
		return false
	}
	span := shared.Span{Start: sentences[0].Span.Start, End: sentences[1].Span.End}
	sentences[0].Effects = []EffectSyntax{{
		Kind:       EffectImpulseExile,
		Context:    EffectContextController,
		Span:       span,
		ClauseSpan: span,
		Text:       sentences[0].Text + " " + sentences[1].Text,
		Tokens:     append(append([]shared.Token(nil), sentences[0].Tokens...), sentences[1].Tokens...),
		Amount:     EffectAmountSyntax{Value: amount, Known: !variableX, VariableX: variableX},
		Duration:   duration,
		Exact:      true,
	}}
	return true
}

// isReminderSentence reports whether a sentence is wholly reminder text, i.e.
// its trimmed text is fully enclosed in parentheses ("(…)"). Such sentences
// carry no game meaning and are excluded from coverage.
func isReminderSentence(sentence Sentence) bool {
	text := strings.TrimSpace(sentence.Text)
	return strings.HasPrefix(text, "(") && strings.HasSuffix(text, ")")
}

// matchImpulseExileClause recognizes "Exile the top card of your library." and
// its counted plural "Exile the top <N> cards of your library." (N a cardinal
// word two..ten) or variable "Exile the top X cards of your library." It returns
// the fixed number of cards exiled and whether the count is the spell's {X}.
func matchImpulseExileClause(text string) (amount int, variableX bool, ok bool) {
	if strings.EqualFold(text, "Exile the top card of your library.") {
		return 1, false, true
	}
	const prefix = "Exile the top "
	const suffix = " cards of your library."
	if len(text) <= len(prefix)+len(suffix) ||
		!strings.EqualFold(text[:len(prefix)], prefix) ||
		!strings.EqualFold(text[len(text)-len(suffix):], suffix) {
		return 0, false, false
	}
	middle := text[len(prefix) : len(text)-len(suffix)]
	if middle == "X" {
		return 0, true, true
	}
	count, ok := CardinalWordValue(middle)
	if !ok || count < 2 {
		return 0, false, false
	}
	return count, false, true
}

// matchImpulsePlayPermissionClause recognizes the temporary play-permission
// sentence that follows an impulse exile: "You may play <object> this turn.",
// the "until end of turn" variants (trailing or leading "Until end of turn,"),
// the "until the end of your next turn" variants, and the "until your next end
// step" variants (Inti, Seneschal of the Sun), where <object> agrees in number
// with the count exiled ("it"/"that card" for a single card, "them"/"those
// cards" for several). It returns the play window.
func matchImpulsePlayPermissionClause(text string, amount int) (EffectDurationKind, bool) {
	for _, object := range impulsePlayObjects(amount) {
		switch {
		case strings.EqualFold(text, "You may play "+object+" this turn."):
			return EffectDurationThisTurn, true
		case strings.EqualFold(text, "You may play "+object+" until end of turn."),
			strings.EqualFold(text, "Until end of turn, you may play "+object+"."):
			return EffectDurationUntilEndOfTurn, true
		case strings.EqualFold(text, "You may play "+object+" until the end of your next turn."),
			strings.EqualFold(text, "Until the end of your next turn, you may play "+object+"."):
			return EffectDurationUntilEndOfYourNextTurn, true
		case strings.EqualFold(text, "You may play "+object+" until your next end step."),
			strings.EqualFold(text, "Until your next end step, you may play "+object+"."):
			return EffectDurationUntilYourNextEndStep, true
		}
	}
	return EffectDurationNone, false
}

// impulsePlayObjects lists the demonstratives an impulse play-permission sentence
// uses to refer back to the exiled cards, matching grammatical number to the
// count exiled.
func impulsePlayObjects(amount int) []string {
	if amount == 1 {
		return []string{"it", "that card"}
	}
	return []string{"them", "those cards"}
}

func recognizeTargetOpponentHandMana(effect *EffectSyntax) bool {
	if effect.Kind != EffectAddMana ||
		!strings.EqualFold(strings.TrimSpace(exactEffectClauseText(effect)), "Add {R} for each card in target opponent's hand.") {
		return false
	}
	effect.Amount = EffectAmountSyntax{
		DynamicKind: EffectDynamicAmountCount,
		DynamicForm: EffectDynamicAmountFormForEach,
		Multiplier:  1,
		Selection: &SelectionSyntax{
			Kind:       SelectionCard,
			Controller: SelectionControllerOpponent,
			Zone:       zone.Hand,
		},
	}
	effect.Mana = EffectManaSyntax{
		Symbols:     []string{"{R}"},
		Colors:      []mana.Color{mana.R},
		ColorsKnown: true,
	}
	return true
}

func recognizeTargetOpponentHandManaSentence(sentence *Sentence) {
	if len(sentence.Effects) != 1 ||
		!recognizeTargetOpponentHandMana(&sentence.Effects[0]) ||
		len(sentence.Targets) != 1 {
		return
	}
	target := sentence.Targets[0]
	target.Cardinality = TargetCardinalitySyntax{Min: 1, Max: 1}
	target.Selection = SelectionSyntax{Kind: SelectionOpponent}
	target.Exact = true
	sentence.Targets[0] = target
	sentence.Effects[0].Targets = []TargetSyntax{target}
}

// recognizeLookAtTargetPlayerHandSentence rewrites the manufactured possessive
// target of a "Look at target <player>'s hand." effect into a clean runtime
// player target. The generic target parser spans the whole "target player's
// hand" phrase with an untyped selection, which the player-target lowering
// cannot consume. This retypes it to a "target player" (or "target opponent")
// selection so lowering can emit a single player target. The parser owns this
// wording; the rewrite only fires once the effect is classified EffectLookAtHand.
func recognizeLookAtTargetPlayerHandSentence(sentence *Sentence) {
	if len(sentence.Effects) != 1 ||
		sentence.Effects[0].Kind != EffectLookAtHand ||
		len(sentence.Targets) != 1 {
		return
	}
	selection := SelectionSyntax{Kind: SelectionPlayer}
	text := "target player"
	if strings.Contains(strings.ToLower(sentence.Targets[0].Text), "opponent") {
		selection = SelectionSyntax{Kind: SelectionOpponent}
		text = "target opponent"
	}
	target := sentence.Targets[0]
	target.Cardinality = TargetCardinalitySyntax{Min: 1, Max: 1}
	target.Selection = selection
	target.Text = text
	target.Exact = true
	sentence.Targets[0] = target
	sentence.Effects[0].Targets = []TargetSyntax{target}
	sentence.Effects[0].Context = EffectContextTarget
	sentence.Effects[0].Exact = true
}

// recognizeDynamicCountMana types an add-mana body whose produced amount scales
// with a dynamic count: a fixed-color battlefield count ("for each <permanent>
// you control", recognizeControlledCountMana), a chosen-color battlefield count
// ("equal to <count>", recognizeChosenColorCountMana), or a source-counter count
// ("for each <kind> counter on this permanent", recognizeSourceCounterCountMana).
// recognizeMillHalfLibrary types the "mills half their library, rounded up/down"
// amount on a mill effect (Traumatize, Fleet Swallower, Cut Your Losses). The
// amount is half the milling player's library, counted as the effect resolves
// and rounded up or down per the trailing "rounded up"/"rounded down" word
// (CR 107.4). The milling player is the effect's own subject, so the possessive
// "their"/"your" names no free referent; the recognizer leaves no reference of
// its own and the semantic reference scan drops the possessive's span (see
// computeSemanticReferences), so downstream stages see a referent-free
// single-player mill. It matches only the exact six-token "half <their|your>
// library , rounded <up|down>" run on a mill effect; every other mill amount is
// left untouched.
func recognizeMillHalfLibrary(effect *EffectSyntax) bool {
	if effect.Kind != EffectMill {
		return false
	}
	for i := 0; i+5 < len(effect.Tokens); i++ {
		if !equalWord(effect.Tokens[i], "half") ||
			(!equalWord(effect.Tokens[i+1], "their") && !equalWord(effect.Tokens[i+1], "your")) ||
			!equalWord(effect.Tokens[i+2], "library") ||
			effect.Tokens[i+3].Kind != shared.Comma ||
			!equalWord(effect.Tokens[i+4], "rounded") ||
			(!equalWord(effect.Tokens[i+5], "down") && !equalWord(effect.Tokens[i+5], "up")) {
			continue
		}
		matched := effect.Tokens[i : i+6]
		span := shared.SpanOf(matched)
		effect.Amount = EffectAmountSyntax{
			Span:        span,
			Text:        joinedEffectText(matched),
			DynamicKind: EffectDynamicAmountHalfPlayerLibrary,
			DynamicForm: EffectDynamicAmountFormHalfLibrary,
			RoundUp:     equalWord(effect.Tokens[i+5], "up"),
		}
		effect.Selection = SelectionSyntax{Kind: SelectionCard, Text: effect.Selection.Text}
		return true
	}
	return false
}

func recognizeDynamicCountMana(effect *EffectSyntax) bool {
	return recognizeControlledCountMana(effect) ||
		recognizeCardsNamedSelfInGraveyardsMana(effect) ||
		recognizeChosenColorCountMana(effect) ||
		recognizeChosenColorSourceCounterMana(effect) ||
		recognizeSourceCounterCountMana(effect) ||
		recognizeSingleColorDynamicMana(effect) ||
		recognizeAnyOneColorDynamicMana(effect)
}

// recognizeCardsNamedSelfInGraveyardsMana types an "Add <mana> for each card
// named <this card> in each graveyard" add-mana body (Rite of Flame) whose
// produced amount scales with the number of self-named cards across every
// graveyard. parseEffectAmount already typed the trailing count clause as a
// self-named graveyard count; the leading mana symbol is left unparsed because
// parseEffectMana rejects the trailing count clause. This re-parses just the
// symbol before the count phrase and records the produced color, so the lowerer
// can emit one mana per counted card. It fires only for a single fixed produced
// color, so choice, any-color, and multi-symbol outputs stay fail-closed.
func recognizeCardsNamedSelfInGraveyardsMana(effect *EffectSyntax) bool {
	if effect.Kind != EffectAddMana ||
		effect.Amount.DynamicKind != EffectDynamicAmountCardsNamedSelfInGraveyards ||
		effect.Amount.DynamicForm != EffectDynamicAmountFormForEach ||
		effect.Amount.Multiplier < 1 {
		return false
	}
	body := manaBodyBeforeAmount(effect)
	if len(body) == 0 {
		return false
	}
	parsed := parseEffectMana(EffectAddMana, body, true)
	if !parsed.ColorsKnown || len(parsed.Colors) != 1 || parsed.Choice {
		return false
	}
	effect.Mana = parsed
	return true
}

// recognizeTargetColorIfRider reports whether a single-target counter or destroy
// effect carries the trailing color rider "if it's <color>" / "if it is
// <color>" (Pyroblast, Red Elemental Blast). The rider is segmented separately
// as a ConditionPredicateTargetColor condition (which carries the color and is
// consumed by counter/destroy lowering), but it leaves the verb clause text
// reading "<verb> <target> if it's <color>." so exactEffectSyntax cannot
// round-trip the bare "<verb> <target>." shape. This credits the effect as exact
// when the only trailing text past the lone exact target is the recognized color
// rider, so lowering can attach the resolving target-color gate.
func recognizeTargetColorIfRider(effect *EffectSyntax, atoms Atoms) bool {
	if effect.Kind != EffectCounter && effect.Kind != EffectDestroy {
		return false
	}
	if len(effect.Targets) != 1 ||
		!effect.Targets[0].Exact ||
		effect.Targets[0].Cardinality.Min != 1 ||
		effect.Targets[0].Cardinality.Max != 1 {
		return false
	}
	targetEnd := effect.Targets[0].Span.End.Offset
	var rider []shared.Token
	for _, token := range effect.Tokens {
		if token.Span.Start.Offset >= targetEnd {
			rider = append(rider, token)
		}
	}
	for len(rider) > 0 && rider[len(rider)-1].Kind == shared.Period {
		rider = rider[:len(rider)-1]
	}
	rest, ok := cutTokenPrefix(rider, "if", "it's")
	if !ok {
		if rest, ok = cutTokenPrefix(rider, "if", "it", "is"); !ok {
			return false
		}
	}
	if len(rest) != 1 {
		return false
	}
	_, ok = atoms.ColorAt(rest[0].Span)
	return ok
}

// recognizeSingleColorDynamicMana types the add-mana body "an amount of
// <symbol> equal to <dynamic>" (Marwyn, the Nurturer: "Add an amount of {G}
// equal to Marwyn's power."), the fixed-color sibling of
// recognizeAnyOneColorDynamicMana. The produced quantity is the dynamic amount
// already typed onto effect.Amount by parseEffectAmount ("equal to <dynamic>" or
// "where X is <dynamic>"); the leading "an amount of <symbol>" body is left
// unrecognized by parseEffectMana because the trailing amount phrase trips its
// fixed-symbol parsing. This re-parses just the lone produced symbol and records
// it as one fixed color, so the lowerer can emit that color scaled by the
// dynamic amount. It fires only for a single fixed produced color, so choice,
// any-color, and multi-symbol outputs stay fail-closed.
func recognizeSingleColorDynamicMana(effect *EffectSyntax) bool {
	if effect.Kind != EffectAddMana ||
		effect.Amount.DynamicKind == EffectDynamicAmountNone {
		return false
	}
	switch effect.Amount.DynamicForm {
	case EffectDynamicAmountFormWhereX, EffectDynamicAmountFormEqual:
	default:
		return false
	}
	body := manaBodyBeforeAmount(effect)
	for len(body) > 0 && body[len(body)-1].Kind == shared.Comma {
		body = body[:len(body)-1]
	}
	if len(body) != 4 || !effectWordsAt(body, 0, "an", "amount", "of") {
		return false
	}
	parsed := parseEffectMana(EffectAddMana, body[3:], true)
	if !parsed.ColorsKnown || len(parsed.Colors) != 1 ||
		parsed.Choice || parsed.AnyColor || parsed.Colors[0] == mana.C {
		return false
	}
	parsed.Span = shared.SpanOf(body)
	effect.Mana = parsed
	return true
}

// recognizeAnyOneColorDynamicMana types the add-mana body "X mana of any one
// color" or "an amount of mana of any one color" (Kami of Whispered Hopes:
// "Add X mana of any one color, where X is this creature's power."). The
// produced quantity is the dynamic amount already typed onto effect.Amount by
// parseEffectAmount ("where X is <dynamic>" or "equal to <dynamic>"); the body
// itself is left unrecognized by parseEffectMana. This credits the
// freely-chosen single-color output when the body matches exactly and a dynamic
// amount is present, so the lowerer can produce that many mana of one chosen
// color. It fails closed when no dynamic amount is present.
func recognizeAnyOneColorDynamicMana(effect *EffectSyntax) bool {
	if effect.Kind != EffectAddMana ||
		effect.Amount.DynamicKind == EffectDynamicAmountNone {
		return false
	}
	switch effect.Amount.DynamicForm {
	case EffectDynamicAmountFormWhereX, EffectDynamicAmountFormEqual:
	default:
		return false
	}
	body := manaBodyBeforeAmount(effect)
	for len(body) > 0 && body[len(body)-1].Kind == shared.Comma {
		body = body[:len(body)-1]
	}
	if !anyOneColorDynamicManaBody(body) {
		return false
	}
	effect.Mana = EffectManaSyntax{Span: shared.SpanOf(body), AnyOneColorDynamic: true}
	return true
}

// anyOneColorDynamicManaBody reports whether body is the leading add-mana phrase
// of an "any one color" dynamic-amount output: "X mana of any one color" or "an
// amount of mana of any one color".
func anyOneColorDynamicManaBody(body []shared.Token) bool {
	if len(body) == 6 &&
		equalWord(body[0], "x") &&
		effectWordsAt(body, 1, "mana", "of", "any", "one", "color") {
		return true
	}
	if len(body) == 8 &&
		effectWordsAt(body, 0, "an", "amount", "of", "mana", "of", "any", "one", "color") {
		return true
	}
	return false
}

// recognizeControlledCountMana types an "Add <mana> for each <permanent> you
// control" add-mana body (Cabal Coffers, Gaea's Cradle, Serra's Sanctum) whose
// produced amount scales with a battlefield permanent count. The "for each
// <permanent>" suffix is already typed onto effect.Amount as a dynamic
// battlefield count by parseEffectAmount; the leading mana symbol, however, is
// left unparsed because parseEffectMana rejects the trailing count clause. This
// re-parses just the symbols before the count phrase and records the produced
// color, so the lowerer can emit one mana per counted permanent. It fires only
// for a single fixed produced color over a battlefield (non-zone) count, so
// choice, any-color, and multi-symbol outputs stay fail-closed.
func recognizeControlledCountMana(effect *EffectSyntax) bool {
	if effect.Kind != EffectAddMana ||
		effect.Amount.DynamicKind != EffectDynamicAmountCount ||
		effect.Amount.DynamicForm != EffectDynamicAmountFormForEach ||
		effect.Amount.Multiplier < 1 ||
		effect.Amount.Selection == nil ||
		effect.Amount.Selection.Zone != zone.None {
		return false
	}
	body := manaBodyBeforeAmount(effect)
	if len(body) == 0 {
		return false
	}
	parsed := parseEffectMana(EffectAddMana, body, true)
	if !parsed.ColorsKnown || len(parsed.Colors) != 1 || parsed.Choice {
		return false
	}
	effect.Mana = parsed
	return true
}

// recognizeSourceCounterCountMana types an "Add <mana> for each <kind> counter
// on <this permanent>" add-mana body (Everflowing Chalice) whose produced amount
// scales with the number of counters of one kind on the source permanent.
// parseEffectAmount types the trailing "for each ... counter on this artifact"
// suffix as a source-counter-count dynamic amount, but the leading mana symbol is
// left unparsed because parseEffectMana rejects the trailing count clause. This
// re-parses just the symbols before the count phrase and records the produced
// color so the lowerer can emit one mana per counter. It fires only for a single
// fixed produced color over a recognized counter kind, so choice and any-color
// outputs stay fail-closed.
func recognizeSourceCounterCountMana(effect *EffectSyntax) bool {
	if effect.Kind != EffectAddMana ||
		effect.Amount.DynamicKind != EffectDynamicAmountSourceCounterCount ||
		effect.Amount.DynamicForm != EffectDynamicAmountFormForEach ||
		effect.Amount.Multiplier < 1 ||
		!effect.Amount.CounterKind.Valid() {
		return false
	}
	body := manaBodyBeforeAmount(effect)
	if len(body) == 0 {
		return false
	}
	parsed := parseEffectMana(EffectAddMana, body, true)
	if !parsed.ColorsKnown || len(parsed.Colors) != 1 || parsed.Choice {
		return false
	}
	effect.Mana = parsed
	return true
}

// recognizeChosenColorSourceCounterMana types the add-mana body "one mana of
// that color" whose produced quantity scales with the number of counters of one
// kind on the source permanent ("Choose a color. Add one mana of that color for
// each charge counter on this artifact.", Astral Cornucopia). The trailing "for
// each <kind> counter on this permanent" suffix is already typed onto
// effect.Amount as a source-counter-count dynamic amount by parseEffectAmount;
// the leading "one mana of that color" body is left unrecognized by
// parseEffectMana. This credits the freely-chosen single-color output when the
// body matches exactly and the amount is a supported source-counter count, so the
// lowerer can produce one mana of the chosen color per counter. It fails closed
// for any other amount.
func recognizeChosenColorSourceCounterMana(effect *EffectSyntax) bool {
	if effect.Kind != EffectAddMana ||
		effect.Amount.DynamicKind != EffectDynamicAmountSourceCounterCount ||
		effect.Amount.DynamicForm != EffectDynamicAmountFormForEach ||
		effect.Amount.Multiplier < 1 ||
		!effect.Amount.CounterKind.Valid() {
		return false
	}
	body := manaBodyBeforeAmount(effect)
	if len(body) != 5 ||
		!effectWordsAt(body, 0, "one", "mana", "of", "that", "color") {
		return false
	}
	effect.Mana = EffectManaSyntax{Span: shared.SpanOf(body), ChosenColor: true, ChosenColorDynamic: true}
	return true
}

// recognizeChosenColorCountMana types the add-mana body "an amount of mana of
// that color equal to <dynamic count>" (Three Tree City: "...equal to the number
// of creatures you control of the chosen type."). The trailing count phrase is
// already typed onto effect.Amount as a dynamic battlefield count by
// parseEffectAmount; the leading "an amount of mana of that color" body is left
// unrecognized by parseEffectMana. This credits the chosen-color output when the
// body matches exactly and the amount is a supported battlefield (non-zone)
// count, so the lowerer can produce one mana of the chosen color per counted
// permanent. It fails closed for a card-zone count or a missing amount.
func recognizeChosenColorCountMana(effect *EffectSyntax) bool {
	if effect.Kind != EffectAddMana ||
		effect.Amount.DynamicKind != EffectDynamicAmountCount ||
		effect.Amount.DynamicForm != EffectDynamicAmountFormEqual ||
		effect.Amount.Multiplier < 1 ||
		effect.Amount.Selection == nil ||
		effect.Amount.Selection.Zone != zone.None {
		return false
	}
	body := manaBodyBeforeAmount(effect)
	if len(body) != 7 ||
		!effectWordsAt(body, 0, "an", "amount", "of", "mana", "of", "that", "color") {
		return false
	}
	effect.Mana = EffectManaSyntax{Span: shared.SpanOf(body), ChosenColor: true, ChosenColorDynamic: true}
	return true
}

// manaBodyBeforeAmount returns the effect tokens that sit after the add-mana
// verb and before the trailing dynamic-count phrase (the produced mana symbols).
func manaBodyBeforeAmount(effect *EffectSyntax) []shared.Token {
	verbEnd := effect.VerbSpan.End.Offset
	amountStart := effect.Amount.Span.Start.Offset
	var body []shared.Token
	for _, token := range effect.Tokens {
		if token.Span.Start.Offset >= verbEnd && token.Span.End.Offset <= amountStart {
			body = append(body, token)
		}
	}
	return body
}

// manaBodyAfterVerb returns the add-mana body tokens that follow the verb,
// dropping a trailing sentence period.
func manaBodyAfterVerb(effect *EffectSyntax) []shared.Token {
	verbEnd := effect.VerbSpan.End.Offset
	var body []shared.Token
	for _, token := range effect.Tokens {
		if token.Span.Start.Offset >= verbEnd {
			body = append(body, token)
		}
	}
	for len(body) > 0 && body[len(body)-1].Kind == shared.Period {
		body = body[:len(body)-1]
	}
	return body
}

// recognizeColorsAmongControlledMana recognizes the add-mana body "one mana of
// any color among <permanents> you control" (Mox Amber, Plaza of Heroes), whose
// produced color is chosen at resolution from the union of colors of the
// permanents the controller controls matching the filter. The filter after
// "among" is parsed by the shared selection parser so it stays generic over the
// permanent group (legendary creatures and planeswalkers, legendary permanents,
// and so on). It fires only for a "you control" battlefield group carrying a
// type, supertype, subtype, or color filter, so an empty wildcard or a non-
// battlefield wording stays fail-closed.
func recognizeColorsAmongControlledMana(effect *EffectSyntax, atoms Atoms) bool {
	if effect.Kind != EffectAddMana ||
		effect.Mana.AnyColor || effect.Mana.ColorsKnown ||
		effect.Mana.CommanderIdentity || effect.Mana.LandsProduce ||
		effect.Mana.LinkedExileColors || effect.Mana.FilterPair ||
		len(effect.Mana.Symbols) != 0 {
		return false
	}
	body := manaBodyAfterVerb(effect)
	if len(body) <= 6 || !effectWordsAt(body, 0, "one", "mana", "of", "any", "color", "among") {
		return false
	}
	selection := parseSelection(body[6:], atoms)
	if selection.Controller != SelectionControllerYou ||
		selection.Zone != zone.None ||
		!colorsAmongSelectionSupported(selection) {
		return false
	}
	clone := selection
	effect.Mana = EffectManaSyntax{
		Span:                  shared.SpanOf(body),
		ColorsAmongControlled: true,
		ColorsAmongSelection:  &clone,
	}
	return true
}

// recognizeEachColorAmongControlledMana recognizes the add-mana body "For each
// color among <permanents> you control, add one mana of that color" (Bloom
// Tender), which produces one mana of each distinct color found among the
// permanents the controller controls matching the filter. The "for each color
// among <group>" prefix precedes the "add" verb; the body after the verb is
// "one mana of that color". The group after "among" is parsed by the shared
// selection parser so it stays generic over the permanent group, and a bare
// "permanents you control" is accepted because the whole controlled board
// legitimately contributes its colors. It fires only for a "you control"
// battlefield group, so a foreign controller or a non-battlefield wording stays
// fail-closed.
func recognizeEachColorAmongControlledMana(effect *EffectSyntax, atoms Atoms) bool {
	if effect.Kind != EffectAddMana ||
		effect.Mana.AnyColor || effect.Mana.ColorsKnown ||
		effect.Mana.ChosenColor || effect.Mana.CommanderIdentity ||
		effect.Mana.LandsProduce || effect.Mana.LinkedExileColors ||
		effect.Mana.FilterPair || effect.Mana.ColorsAmongControlled ||
		effect.Amount.DynamicKind != "" ||
		len(effect.Mana.Symbols) != 0 {
		return false
	}
	body := manaBodyAfterVerb(effect)
	if len(body) != 5 || !effectWordsAt(body, 0, "one", "mana", "of", "that", "color") {
		return false
	}
	prefix := manaPrefixBeforeVerb(effect)
	for len(prefix) > 0 && prefix[len(prefix)-1].Kind == shared.Comma {
		prefix = prefix[:len(prefix)-1]
	}
	if len(prefix) <= 4 || !effectWordsAt(prefix, 0, "for", "each", "color", "among") {
		return false
	}
	selection := parseSelection(prefix[4:], atoms)
	if selection.Controller != SelectionControllerYou ||
		selection.Zone != zone.None {
		return false
	}
	// Accept either a narrowed group whose predicate the selection parser
	// captures (the colorsAmongControlled facets) or the exact literal bare
	// "permanents you control" group. Any other prefix (e.g. "monocolored
	// permanents you control", whose qualifier the selection parser drops
	// silently) fails closed so it cannot lower to a mislabeled ability.
	bareControlled := len(prefix) == 7 && effectWordsAt(prefix, 4, "permanents", "you", "control")
	if !colorsAmongSelectionSupported(selection) && !bareControlled {
		return false
	}
	clone := selection
	effect.Mana = EffectManaSyntax{
		Span:                     shared.SpanOf(effect.Tokens),
		EachColorAmongControlled: true,
		ColorsAmongSelection:     &clone,
	}
	return true
}

// manaPrefixBeforeVerb returns the effect tokens that precede the add-mana verb,
// such as a "For each color among <group>," distributive prefix.
func manaPrefixBeforeVerb(effect *EffectSyntax) []shared.Token {
	verbStart := effect.VerbSpan.Start.Offset
	var prefix []shared.Token
	for _, token := range effect.Tokens {
		if token.Span.End.Offset <= verbStart {
			prefix = append(prefix, token)
		}
	}
	return prefix
}

// filter carries an exact, supported permanent predicate. It requires a type,
// supertype, subtype, or color filter (so a bare "permanents you control" with
// no narrowing predicate fails closed) and rejects qualifiers the executable
// backend cannot represent for this body (numeric, combat, tapped, "another",
// or excluded-type/keyword qualifiers).
func colorsAmongSelectionSupported(selection SelectionSyntax) bool {
	if selection.All || selection.Another || selection.Other ||
		selection.Attacking || selection.Blocking ||
		selection.Tapped || selection.Untapped ||
		selection.MatchManaValue || selection.MatchPower || selection.MatchToughness ||
		selection.Keyword != KeywordUnknown || selection.ExcludedKeyword != KeywordUnknown ||
		len(selection.ExcludedTypes) != 0 || len(selection.ExcludedSupertypes) != 0 ||
		len(selection.ExcludedColors) != 0 || len(selection.Alternatives) != 0 {
		return false
	}
	return len(selection.RequiredTypesAny) != 0 ||
		len(selection.Supertypes) != 0 ||
		len(selection.SubtypesAny) != 0 ||
		len(selection.ColorsAny) != 0 ||
		selection.Colorless || selection.Multicolored ||
		selectionKindNarrowsPermanent(selection.Kind)
}

// selectionKindNarrowsPermanent reports whether a selection Kind names a concrete
// permanent card type (so "creatures you control" narrows) rather than the
// catch-all permanent/any kinds (so "permanents you control" alone does not).
func selectionKindNarrowsPermanent(kind SelectionKind) bool {
	switch kind {
	case SelectionArtifact, SelectionCreature, SelectionEnchantment,
		SelectionLand, SelectionPlaneswalker:
		return true
	default:
		return false
	}
}

func parseHandDiscard(effect *EffectSyntax) HandDiscardSyntax {
	if effect.Kind != EffectDiscard ||
		effect.Context != EffectContextController ||
		!effect.Amount.Known || effect.Amount.Value < 1 ||
		len(effect.Targets) != 0 ||
		len(effect.References) != 0 {
		return HandDiscardSyntax{}
	}
	if exactCardCountEffectSyntax(effect, "Discard", "discards", false) {
		return HandDiscardSyntax{Present: true}
	}
	if exactControllerRandomDiscardSyntax(effect) {
		return HandDiscardSyntax{Present: true, AtRandom: true}
	}
	return HandDiscardSyntax{}
}

// exactControllerRandomDiscardSyntax reconstructs the canonical "discard <N>
// card(s) at random" wording for a controller-context discard of a known fixed
// count. The "at random" suffix marks a random discard, distinguishing it from
// the player-choice discard exactCardCountEffectSyntax recognizes.
func exactControllerRandomDiscardSyntax(effect *EffectSyntax) bool {
	if !effect.Amount.Known || effect.Amount.Value < 1 || effect.Amount.RangeKnown ||
		effect.Amount.DynamicForm != EffectDynamicAmountFormNone {
		return false
	}
	noun := "cards"
	if effect.Amount.Value == 1 {
		noun = "card"
	}
	text := exactEffectClauseText(effect)
	amountText := effectAmountSourceText(effect)
	for _, prefix := range []string{"Discard", "You discard"} {
		if strings.EqualFold(text, fmt.Sprintf("%s %s %s at random.", prefix, amountText, noun)) {
			return true
		}
	}
	return false
}

// parseDiscardEntireHand recognizes a "discard their hand" clause whose affected
// player discards every card in hand. It accepts the controller ("Discard your
// hand" and the verbose "Discard all the cards in your hand"), each-player,
// each-opponent, and single-target-player forms; the amount is unknown because
// it depends on the player's hand at resolution. The
// optional "You may discard your hand." offer is accepted too: exactEffectClause
// text strips the "you may" prefix, so the controller clause reconstructs
// exactly and the optional wrapper gates the entire-hand discard.
func parseDiscardEntireHand(effect *EffectSyntax) bool {
	if effect.Kind != EffectDiscard ||
		effect.Amount.Known ||
		effect.Negated {
		return false
	}
	text := strings.TrimSpace(exactEffectClauseText(effect))
	switch effect.Context {
	case EffectContextController:
		return len(effect.Targets) == 0 &&
			(strings.EqualFold(text, "Discard your hand.") ||
				strings.EqualFold(text, "You discard your hand.") ||
				strings.EqualFold(text, "Discard all the cards in your hand.") ||
				strings.EqualFold(text, "You discard all the cards in your hand."))
	case EffectContextEachPlayer:
		return len(effect.Targets) == 0 &&
			strings.EqualFold(text, "Each player discards their hand.")
	case EffectContextEachOtherPlayer:
		return len(effect.Targets) == 0 &&
			strings.EqualFold(text, "Each other player discards their hand.")
	case EffectContextEachOpponent:
		return len(effect.Targets) == 0 &&
			strings.EqualFold(text, "Each opponent discards their hand.")
	case EffectContextTarget:
		return len(effect.Targets) == 1 && effect.Targets[0].Exact &&
			exactCardCountTargetPlayer(effect.Targets[0].Selection) &&
			strings.EqualFold(text, titleFirstEffectText(effect.Targets[0].Text)+" discards their hand.")
	default:
		return false
	}
}

func parsePlayerProtectionEffects(sentence Sentence, tokens []shared.Token, _ Atoms) ([]EffectSyntax, bool) {
	if strings.TrimSpace(sentence.Text) != "Until your next turn, your life total can't change and you gain protection from everything." {
		return nil, false
	}
	changeIndex, andIndex, gainIndex := -1, -1, -1
	for i := range tokens {
		switch {
		case strings.EqualFold(tokens[i].Text, "change"):
			changeIndex = i
		case changeIndex >= 0 && andIndex < 0 && strings.EqualFold(tokens[i].Text, "and"):
			andIndex = i
		case strings.EqualFold(tokens[i].Text, "gain"):
			gainIndex = i
		default:
		}
	}
	if changeIndex < 0 || andIndex < 0 || gainIndex < 0 || andIndex+1 >= len(tokens) {
		return nil, false
	}
	base := EffectSyntax{
		Span:                    sentence.Span,
		Text:                    sentence.Text,
		Tokens:                  append([]shared.Token(nil), tokens...),
		Duration:                EffectDurationUntilYourNextTurn,
		Context:                 EffectContextController,
		Exact:                   true,
		RequiresOrderedLowering: true,
	}
	life := base
	life.Kind = EffectLifeTotalCantChange
	life.ClauseSpan = shared.Span{Start: sentence.Span.Start, End: tokens[changeIndex].Span.End}
	life.VerbSpan = tokens[changeIndex].Span
	protection := base
	protection.Kind = EffectProtectionFromEverything
	protection.Connection = EffectConnectionAnd
	protection.ConnectionSpan = tokens[andIndex].Span
	protection.ClauseSpan = shared.Span{Start: tokens[andIndex+1].Span.Start, End: sentence.Span.End}
	protection.VerbSpan = tokens[gainIndex].Span
	return []EffectSyntax{life, protection}, true
}

func parseGroupPhaseOutEffect(sentence Sentence, tokens []shared.Token, atoms Atoms) ([]EffectSyntax, bool) {
	if strings.TrimSpace(sentence.Text) != "All permanents you control phase out." {
		return nil, false
	}
	return []EffectSyntax{{
		Kind:       EffectPhaseOut,
		Span:       sentence.Span,
		ClauseSpan: sentence.Span,
		Text:       sentence.Text,
		Tokens:     append([]shared.Token(nil), tokens...),
		Selection:  parseSelection(tokens, atoms),
		Exact:      true,
	}}, true
}

// parsePhaseOutEffect recognizes the singular-subject "<subject> phases out."
// effect (CR 702.26), such as "This creature phases out." (Blink Dog) and
// "Target creature phases out." The plural mass form "All permanents you control
// phase out." is handled separately by parseGroupPhaseOutEffect; this recognizer
// keys on the singular "phases out" verb that closes the sentence, so the two
// never overlap. The subject's self reference or target is scanned by the shared
// reference and target passes, so the recognizer only identifies the verb and
// covers the sentence; lowering fails closed for any subject it cannot represent.
func parsePhaseOutEffect(sentence Sentence, tokens []shared.Token, atoms Atoms) ([]EffectSyntax, bool) {
	if !sentenceClosesWithPhasesOut(tokens) {
		return nil, false
	}
	return []EffectSyntax{{
		Kind:       EffectPhaseOut,
		Span:       sentence.Span,
		ClauseSpan: sentence.Span,
		Text:       sentence.Text,
		Tokens:     append([]shared.Token(nil), tokens...),
		Selection:  parseSelection(tokens, atoms),
		Exact:      true,
	}}, true
}

// sentenceClosesWithPhasesOut reports whether the sentence's meaningful tokens
// end with the singular "phases out" verb immediately before the closing period.
// It requires a subject before the verb so the bare verb alone never matches.
func sentenceClosesWithPhasesOut(tokens []shared.Token) bool {
	period := shared.TopLevelIndex(tokens, shared.Period)
	if period < 3 {
		return false
	}
	return equalWord(tokens[period-2], "phases") && equalWord(tokens[period-1], "out")
}

// parseMassReanimationExchangeEffect recognizes the symmetric mass-reanimation
// sentence "Each player exiles all <type> cards from their graveyard, then
// sacrifices all <type> they control, then puts all cards they exiled this way
// onto the battlefield." (Living Death, Living End, Scrap Mastery). The leading
// type word is singular ("creature"/"artifact") in the exile clause and plural
// ("creatures"/"artifacts") in the sacrifice clause; both must name the same
// card type. The whole sentence collapses to one EffectMassReanimationExchange
// whose Selection carries the card-type filter (parsed from the "all <type>
// cards from their graveyard" sub-phrase), letting the lowering stay text-blind.
// Any other wording fails closed and flows through the generic effect parser.
func parseMassReanimationExchangeEffect(sentence Sentence, tokens []shared.Token, atoms Atoms) ([]EffectSyntax, bool) {
	words, ok := massReanimationExchangeWords(tokens)
	if !ok {
		return nil, false
	}
	return []EffectSyntax{{
		Kind:       EffectMassReanimationExchange,
		Span:       sentence.Span,
		ClauseSpan: sentence.Span,
		VerbSpan:   words[2].Span,
		Text:       sentence.Text,
		Tokens:     append([]shared.Token(nil), tokens...),
		Context:    EffectContextEachPlayer,
		Selection:  parseSelection(words[3:9], atoms),
		Exact:      true,
	}}, true
}

// massReanimationExchangeWords returns the non-punctuation tokens of a sentence
// when they match the mass-reanimation exchange template (see
// parseMassReanimationExchangeEffect), and reports whether they matched. The
// returned slice is indexable by the template positions, so callers read the
// "all <type> cards from their graveyard" sub-phrase as words[3:9].
func massReanimationExchangeWords(tokens []shared.Token) ([]shared.Token, bool) {
	words := make([]shared.Token, 0, len(tokens))
	for _, token := range tokens {
		if token.Kind == shared.Word {
			words = append(words, token)
		}
	}
	template := []string{
		"each", "player", "exiles", "all", "", "cards", "from", "their", "graveyard",
		"then", "sacrifices", "all", "", "they", "control",
		"then", "puts", "all", "cards", "they", "exiled", "this", "way", "onto", "the", "battlefield",
	}
	if len(words) != len(template) {
		return nil, false
	}
	for offset, want := range template {
		if want == "" {
			continue
		}
		if !equalWord(words[offset], want) {
			return nil, false
		}
	}
	singular := strings.ToLower(words[4].Text)
	if singular != "creature" && singular != "artifact" {
		return nil, false
	}
	if !equalWord(words[12], singular+"s") {
		return nil, false
	}
	return words, true
}

// parseAdditionalLandPlaysEffect recognizes the controller-scoped grant of one
// or more extra land plays for the turn: "Play an additional land this turn.",
// "You may play an additional land this turn.", and the multi-land "... two
// additional lands ..." / "... up to N additional lands ..." variants. The "you
// may" permission is folded into an unconditional allowance (the player is never
// forced to play the extra land). The static "on each of your turns" form is a
// separate static ability and is not matched here.
// parseCastAsThoughFlashEffect recognizes the controller-scoped, turn-scoped
// timing permission "You may cast spells this turn as though they had flash."
// (Borne Upon a Wind, Emergence Zone). The leading "you may" is a permission,
// not a resolving choice, so the effect is modeled unconditionally (like
// parseAdditionalLandPlaysEffect) rather than as an optional cast effect. Any
// other wording fails closed and flows through the generic effect parser.
func parseCastAsThoughFlashEffect(sentence Sentence, tokens []shared.Token) ([]EffectSyntax, bool) {
	words := make([]shared.Token, 0, len(tokens))
	for _, token := range tokens {
		if token.Kind == shared.Period {
			continue
		}
		words = append(words, token)
	}
	if len(words) >= 2 && equalWord(words[0], "you") && equalWord(words[1], "may") {
		words = words[2:]
	}
	if len(words) != 9 || !equalWord(words[0], "cast") {
		return nil, false
	}
	castToken := words[0]
	if !equalWord(words[1], "spells") ||
		!equalWord(words[2], "this") ||
		!equalWord(words[3], "turn") ||
		!equalWord(words[4], "as") ||
		!equalWord(words[5], "though") ||
		!equalWord(words[6], "they") ||
		!equalWord(words[7], "had") ||
		!equalWord(words[8], "flash") {
		return nil, false
	}
	return []EffectSyntax{{
		Kind:       EffectCastAsThoughFlash,
		Span:       sentence.Span,
		ClauseSpan: sentence.Span,
		VerbSpan:   castToken.Span,
		Text:       sentence.Text,
		Tokens:     append([]shared.Token(nil), tokens...),
		Context:    EffectContextController,
		Duration:   EffectDurationThisTurn,
		Exact:      true,
	}}, true
}

// parseAdditionalCombatPhaseEffect recognizes the extra-phase-insertion effect
// in both clause orders: the leading "After this [main/combat] phase, there is
// an additional combat phase[ followed by an additional main phase]."
// (Aggravated Assault, Aurelia the Warleader, World at War, Combat Celebrant)
// and the trailing "There is an additional combat phase after this phase."
// (Raiyuu, Storm's Edge; Moraug, Fury of Akoum, which prints "there's"). It
// inserts an additional combat phase into the current turn, optionally followed
// by an additional main phase. The "after this <phase>" reference to the current
// phase is descriptive; the simplified runtime model drains the inserted phases
// after the postcombat main phase. A leading condition clause ("If it's the
// first combat phase of the turn, ...") is stripped here so the extra-phase
// grammar matches; that condition is recognized separately and re-associated by
// the compiler through this effect's sentence-wide ClauseSpan. Any other wording
// fails closed and flows through the generic parser.
func parseAdditionalCombatPhaseEffect(sentence Sentence, tokens []shared.Token) ([]EffectSyntax, bool) {
	effectTokens := tokens
	if len(tokens) > 0 {
		if intro, _ := conditionIntroAt(tokens, 0); intro != ConditionIntroUnknown {
			end := conditionClauseEnd(tokens, 0)
			if end >= len(tokens) || tokens[end].Kind != shared.Comma {
				return nil, false
			}
			effectTokens = tokens[end+1:]
		}
	}
	words := make([]shared.Token, 0, len(effectTokens))
	for _, token := range effectTokens {
		if token.Kind == shared.Period || token.Kind == shared.Comma {
			continue
		}
		words = append(words, token)
	}
	verbToken, additionalMain, ok := matchAdditionalCombatPhaseWords(words)
	if !ok {
		return nil, false
	}
	return []EffectSyntax{{
		Kind:                  EffectAdditionalCombatPhase,
		Span:                  sentence.Span,
		ClauseSpan:            sentence.Span,
		VerbSpan:              verbToken.Span,
		Text:                  sentence.Text,
		Tokens:                append([]shared.Token(nil), tokens...),
		Context:               EffectContextController,
		AdditionalCombatPhase: true,
		AdditionalMainPhase:   additionalMain,
		Exact:                 true,
	}}, true
}

// matchAdditionalCombatPhaseWords matches the punctuation-stripped words of an
// additional-combat-phase clause in either order and reports the "there is"
// verb token and whether an additional main phase follows. It fails closed for
// any other wording.
func matchAdditionalCombatPhaseWords(words []shared.Token) (verb shared.Token, additionalMain bool, ok bool) {
	if verb, main, ok := matchLeadingAdditionalCombatPhaseWords(words); ok {
		return verb, main, true
	}
	return matchTrailingAdditionalCombatPhaseWords(words)
}

// matchLeadingAdditionalCombatPhaseWords matches "after this [main|combat] phase
// there is an additional combat phase[ followed by an additional main phase]".
func matchLeadingAdditionalCombatPhaseWords(words []shared.Token) (verb shared.Token, additionalMain bool, ok bool) {
	rest, ok := cutTokenPrefix(words, "after", "this")
	if !ok || len(rest) == 0 {
		return shared.Token{}, false, false
	}
	if equalWord(rest[0], "main") || equalWord(rest[0], "combat") {
		rest = rest[1:]
	}
	rest, ok = cutTokenPrefix(rest, "phase")
	if !ok || len(rest) == 0 {
		return shared.Token{}, false, false
	}
	verb = rest[0]
	rest, ok = cutTokenPrefix(rest, "there", "is", "an", "additional", "combat", "phase")
	if !ok {
		return shared.Token{}, false, false
	}
	main, ok := matchAdditionalMainPhaseTail(rest)
	if !ok {
		return shared.Token{}, false, false
	}
	return verb, main, true
}

// matchTrailingAdditionalCombatPhaseWords matches "there is an additional combat
// phase after this [phase|one][ followed by an additional main phase]", also
// accepting the "there's" contraction.
func matchTrailingAdditionalCombatPhaseWords(words []shared.Token) (verb shared.Token, additionalMain bool, ok bool) {
	if len(words) == 0 {
		return shared.Token{}, false, false
	}
	verb = words[0]
	rest, ok := cutTokenPrefix(words, "there's", "an", "additional", "combat", "phase")
	if !ok {
		rest, ok = cutTokenPrefix(words, "there", "is", "an", "additional", "combat", "phase")
	}
	if !ok {
		return shared.Token{}, false, false
	}
	rest, ok = cutTokenPrefix(rest, "after", "this")
	if !ok || len(rest) == 0 || (!equalWord(rest[0], "phase") && !equalWord(rest[0], "one")) {
		return shared.Token{}, false, false
	}
	main, ok := matchAdditionalMainPhaseTail(rest[1:])
	if !ok {
		return shared.Token{}, false, false
	}
	return verb, main, true
}

// matchAdditionalMainPhaseTail consumes an optional "followed by an additional
// main phase" tail, reporting whether it was present. It fails closed if any
// other tokens remain.
func matchAdditionalMainPhaseTail(words []shared.Token) (additionalMain bool, ok bool) {
	if len(words) == 0 {
		return false, true
	}
	rest, ok := cutTokenPrefix(words, "followed", "by", "an", "additional", "main", "phase")
	if !ok || len(rest) != 0 {
		return false, false
	}
	return true, true
}

// parseRollDieEffect recognizes "roll a d<N>" (CR 706), the die-roll mechanic
// whose result a following effect consumes via "...equal to the result." It
// types an EffectRollDie carrying DieSides = N. The "d<N>" lexes as a word "d"
// followed by an integer N; any other wording fails closed and flows through the
// generic effect parser. It backs the Ancient Dragon dice cycle (Ancient Copper
// Dragon et al.) and other "roll a dN" cards.
func parseRollDieEffect(sentence Sentence, tokens []shared.Token) ([]EffectSyntax, bool) {
	words := make([]shared.Token, 0, len(tokens))
	for _, token := range tokens {
		if token.Kind == shared.Period {
			continue
		}
		words = append(words, token)
	}
	if len(words) != 4 ||
		!equalWord(words[0], "roll") ||
		!equalWord(words[1], "a") ||
		!equalWord(words[2], "d") ||
		words[3].Kind != shared.Integer {
		return nil, false
	}
	sides, err := strconv.Atoi(words[3].Text)
	if err != nil || sides < 2 {
		return nil, false
	}
	return []EffectSyntax{{
		Kind:       EffectRollDie,
		Span:       sentence.Span,
		ClauseSpan: sentence.Span,
		VerbSpan:   words[0].Span,
		Text:       sentence.Text,
		Tokens:     append([]shared.Token(nil), tokens...),
		Context:    EffectContextController,
		DieSides:   sides,
		Exact:      true,
	}}, true
}

// parseRingTemptsEffect recognizes the fixed designation effect "The Ring
// tempts you." (CR 701.51). The wording is fully fixed — the resolving
// controller gets the Ring emblem, advances it one level, and chooses a
// creature they control as their Ring-bearer — so the recognizer matches the
// whole sentence and emits a controller-scoped EffectRingTempts carrying no
// targets or amounts. Any other wording fails closed and flows through the
// general effect parser.
func parseRingTemptsEffect(sentence Sentence, tokens []shared.Token) ([]EffectSyntax, bool) {
	words := make([]shared.Token, 0, len(tokens))
	for _, token := range tokens {
		if token.Kind == shared.Period {
			continue
		}
		words = append(words, token)
	}
	expected := []string{"the", "ring", "tempts", "you"}
	if len(words) != len(expected) {
		return nil, false
	}
	for i, word := range expected {
		if !equalWord(words[i], word) {
			return nil, false
		}
	}
	return []EffectSyntax{{
		Kind:       EffectRingTempts,
		Span:       sentence.Span,
		ClauseSpan: sentence.Span,
		VerbSpan:   words[2].Span,
		Text:       sentence.Text,
		Tokens:     append([]shared.Token(nil), tokens...),
		Context:    EffectContextController,
		Exact:      true,
	}}, true
}

// parseNoMaximumHandSizeForRestOfGameEffect recognizes the controller-scoped,
// rest-of-game continuous effect "You have no maximum hand size for the rest of
// the game." (Sea Gate Restoration). As a resolving spell effect it removes the
// controller's maximum hand size for the rest of the game; the "rest of the
// game" duration is fixed by the wording, so the effect carries no parsed
// duration. The permanent static "You have no maximum hand size." form
// (Reliquary Tower) is handled by the static-declaration parser, so this
// recognizer requires the trailing "for the rest of the game" clause. Any other
// wording fails closed and flows through the generic effect parser.
func parseNoMaximumHandSizeForRestOfGameEffect(sentence Sentence, tokens []shared.Token) ([]EffectSyntax, bool) {
	words := make([]shared.Token, 0, len(tokens))
	for _, token := range tokens {
		if token.Kind == shared.Period {
			continue
		}
		words = append(words, token)
	}
	expected := []string{
		"you", "have", "no", "maximum", "hand", "size",
		"for", "the", "rest", "of", "the", "game",
	}
	if len(words) != len(expected) {
		return nil, false
	}
	for i, word := range expected {
		if !equalWord(words[i], word) {
			return nil, false
		}
	}
	return []EffectSyntax{{
		Kind:       EffectNoMaximumHandSize,
		Span:       sentence.Span,
		ClauseSpan: sentence.Span,
		VerbSpan:   words[1].Span,
		Text:       sentence.Text,
		Tokens:     append([]shared.Token(nil), tokens...),
		Context:    EffectContextController,
		Exact:      true,
	}}, true
}

// prohibition "<players> can't cast spells this turn." (Silence: "Your opponents
// can't cast spells this turn."; "Players can't cast spells this turn."). The
// affected players are the controller's opponents ("your opponents", "each
// opponent") or every player ("players"). It is modeled as an unconditional
// turn-scoped restriction reusing the continuous cast-prohibition rule effect.
// Targeted, referenced, defending-player, and spell-type-filtered wordings fail
// closed and flow through the generic effect parser.
func parseCantCastSpellsEffect(sentence Sentence, tokens []shared.Token) ([]EffectSyntax, bool) {
	words := make([]shared.Token, 0, len(tokens))
	for _, token := range tokens {
		if token.Kind == shared.Period {
			continue
		}
		words = append(words, token)
	}
	index := 0
	allPlayers := false
	switch {
	case len(words) >= 2 && equalWord(words[0], "your") && equalWord(words[1], "opponents"):
		index = 2
	case len(words) >= 2 && equalWord(words[0], "each") && equalWord(words[1], "opponent"):
		index = 2
	case len(words) >= 1 && equalWord(words[0], "players"):
		allPlayers = true
		index = 1
	default:
		return nil, false
	}
	if index >= len(words) || (!equalWord(words[index], "can't") && !equalWord(words[index], "cannot")) {
		return nil, false
	}
	index++
	if index >= len(words) || !equalWord(words[index], "cast") {
		return nil, false
	}
	castSpan := words[index].Span
	index++
	// An optional card-type word between "cast" and "spells" filters the
	// prohibition: a bare card type ("creature spells") restricts it to that
	// type, while a "non"-prefixed word ("noncreature spells") exempts that type.
	var requiredTypes, excludedTypes []CardType
	switch len(words) - index {
	case 4:
		requiredType, excludedType, ok := cantCastSpellsFilterType(words[index].Text)
		if !ok {
			return nil, false
		}
		if requiredType != CardTypeUnknown {
			requiredTypes = []CardType{requiredType}
		} else {
			excludedTypes = []CardType{excludedType}
		}
		index++
	case 3:
	default:
		return nil, false
	}
	for offset, want := range []string{"spells", "this", "turn"} {
		if !equalWord(words[index+offset], want) {
			return nil, false
		}
	}
	return []EffectSyntax{{
		Kind:                        EffectCantCastSpells,
		Span:                        sentence.Span,
		ClauseSpan:                  sentence.Span,
		VerbSpan:                    castSpan,
		Text:                        sentence.Text,
		Tokens:                      append([]shared.Token(nil), tokens...),
		Context:                     EffectContextController,
		Duration:                    EffectDurationThisTurn,
		CantCastSpellsAllPlayers:    allPlayers,
		CantCastSpellsRequiredTypes: requiredTypes,
		CantCastSpellsExcludedTypes: excludedTypes,
		Exact:                       true,
	}}, true
}

// cantCastSpellsFilterType maps the optional card-type word in a "<players> can't
// cast <filter> spells this turn" clause to either a required card type (a bare
// type word such as "creature") or an excluded card type (a "non"-prefixed word
// such as "noncreature"). Exactly one of the returned types is set; it fails
// closed for any word that is not a recognized card type.
func cantCastSpellsFilterType(word string) (required, excluded CardType, ok bool) {
	if cardType, found := recognizeCardTypeWord(word); found {
		return cardType, CardTypeUnknown, true
	}
	if rest, found := strings.CutPrefix(strings.ToLower(word), "non"); found {
		if cardType, found := recognizeCardTypeWord(rest); found {
			return CardTypeUnknown, cardType, true
		}
	}
	return CardTypeUnknown, CardTypeUnknown, false
}

// parseGroupCantBlockEffect recognizes the one-shot, group-scoped combat
// restriction "<group> can't block this turn." (Falter, Magmatic Chasm, Seismic
// Stomp: "Creatures without flying can't block this turn."; Cosmotronic Wave:
// "Creatures your opponents control can't block this turn."). The affected
// creature group is recognized through the shared static-subject grammar and
// recorded in StaticSubject so lowering can scope the this-turn can't-block rule
// effect by controller, color, and keyword filter. It deliberately never matches
// the targeted form "<target> can't block this turn." (that keeps flowing to the
// generic per-clause target recognizer), so a leading "target" quantifier fails
// closed here. Any subject the static-subject grammar cannot represent also
// fails closed and flows through the generic effect parser.
func parseGroupCantBlockEffect(sentence Sentence, tokens []shared.Token, atoms Atoms) ([]EffectSyntax, bool) {
	remaining, leadingDuration := stripLeadingDurationClause(tokens, atoms)
	if leadingDuration != EffectDurationNone {
		return nil, false
	}
	words := make([]shared.Token, 0, len(remaining))
	for _, token := range remaining {
		if token.Kind == shared.Period {
			continue
		}
		words = append(words, token)
	}
	tail := []string{"can't", "block", "this", "turn"}
	if len(words) <= len(tail) {
		return nil, false
	}
	verbIndex := len(words) - len(tail)
	for offset, want := range tail {
		if !equalWord(words[verbIndex+offset], want) {
			return nil, false
		}
	}
	subjectWords := words[:verbIndex]
	for _, word := range subjectWords {
		// The targeted form "<target> can't block this turn." is owned by the
		// generic target recognizer; never hijack it here.
		if equalWord(word, "target") {
			return nil, false
		}
	}
	subject, ok := recognizeGroupSubjectWords(subjectWords)
	if !ok {
		return nil, false
	}
	subject.Span = shared.SpanOf(subjectWords)
	return []EffectSyntax{{
		Kind:          EffectCantBlock,
		Span:          sentence.Span,
		ClauseSpan:    sentence.Span,
		VerbSpan:      words[verbIndex].Span,
		Text:          sentence.Text,
		Tokens:        append([]shared.Token(nil), tokens...),
		Context:       EffectContextController,
		Duration:      EffectDurationThisTurn,
		StaticSubject: subject,
		Exact:         true,
	}}, true
}

// recognizeGroupSubjectWords recognizes a static creature-group subject standing
// on its own (without a trailing group verb) by reusing the shared
// parseEffectStaticSubject grammar: it appends a synthetic "have" verb so the
// subject sub-parsers, which require a trailing group verb, engage over exactly
// the subject tokens. It returns the recognized subject with Kind==None cleared
// to a failure so an unrepresentable subject fails closed.
func recognizeGroupSubjectWords(subjectWords []shared.Token) (EffectStaticSubjectSyntax, bool) {
	if len(subjectWords) == 0 {
		return EffectStaticSubjectSyntax{}, false
	}
	synthetic := append(append([]shared.Token(nil), subjectWords...), shared.Token{Kind: shared.Word, Text: "have"})
	subject := parseEffectStaticSubject(synthetic, collectAtoms(synthetic, nil, nil, "", false))
	if subject.Kind == EffectStaticSubjectNone {
		return EffectStaticSubjectSyntax{}, false
	}
	// Only a plain creature group refined by an optional color and/or single
	// keyword filter is representable by the this-turn can't-block rule effect.
	// Subtype-, counter-, power-, and chosen-color-filtered groups drop their
	// refinement when compiled, so they must fail closed here rather than widen
	// to every creature.
	if subject.SubtypeKnown ||
		len(subject.SubtypesAny) != 0 ||
		subject.ExcludedSubtype ||
		len(subject.ExcludedSubtypes) != 0 ||
		len(subject.ExcludedTypes) != 0 ||
		subject.CounterRequired ||
		subject.CounterAny ||
		subject.MatchPower ||
		subject.MatchToughness ||
		subject.PowerOrToughness ||
		subject.PowerLessThanSource ||
		subject.PowerGreaterThanSource ||
		subject.ChosenColorFromEntry {
		return EffectStaticSubjectSyntax{}, false
	}
	return subject, true
}

// parseGroupMustAttackEffect recognizes the one-shot forced-attack effect
// "<group> attack this turn if able." (Bident of Thassa: "Creatures your
// opponents control attack this turn if able."; "Creatures you control attack
// this turn if able."; "All creatures attack this turn if able.") and its
// duration-scoped variant "Until your next turn, <group> attack each combat if
// able." (The Akroan War chapter II: "Until your next turn, creatures your
// opponents control attack each combat if able."). The affected creature group
// is recorded in StaticSubject so lowering can scope the continuous must-attack
// rule effect by controller, and the recognized duration becomes the rule
// effect's lifetime. Any other subject, duration, or trailing clause fails
// closed and flows through the generic effect parser.
func parseGroupMustAttackEffect(sentence Sentence, tokens []shared.Token, atoms Atoms) ([]EffectSyntax, bool) {
	remaining, leadingDuration := stripLeadingDurationClause(tokens, atoms)
	words := make([]shared.Token, 0, len(remaining))
	for _, token := range remaining {
		if token.Kind == shared.Period {
			continue
		}
		words = append(words, token)
	}
	var subject EffectStaticSubjectKind
	index := 0
	switch {
	case len(words) >= 4 && equalWord(words[0], "creatures") &&
		equalWord(words[1], "your") && equalWord(words[2], "opponents") && equalWord(words[3], "control"):
		subject = EffectStaticSubjectOpponentControlledCreatures
		index = 4
	case len(words) >= 3 && equalWord(words[0], "creatures") &&
		equalWord(words[1], "you") && equalWord(words[2], "control"):
		subject = EffectStaticSubjectControlledCreatures
		index = 3
	case len(words) >= 2 && equalWord(words[0], "all") && equalWord(words[1], "creatures"):
		subject = EffectStaticSubjectAllCreatures
		index = 2
	default:
		return nil, false
	}
	var rest []string
	var duration EffectDurationKind
	switch leadingDuration {
	case EffectDurationNone:
		rest = []string{"attack", "this", "turn", "if", "able"}
		duration = EffectDurationThisTurn
	case EffectDurationUntilYourNextTurn:
		rest = []string{"attack", "each", "combat", "if", "able"}
		duration = EffectDurationUntilYourNextTurn
	default:
		return nil, false
	}
	if len(words)-index != len(rest) {
		return nil, false
	}
	for offset, want := range rest {
		if !equalWord(words[index+offset], want) {
			return nil, false
		}
	}
	return []EffectSyntax{{
		Kind:          EffectMustAttack,
		Span:          sentence.Span,
		ClauseSpan:    sentence.Span,
		VerbSpan:      words[index].Span,
		Text:          sentence.Text,
		Tokens:        append([]shared.Token(nil), tokens...),
		Context:       EffectContextController,
		Duration:      duration,
		StaticSubject: EffectStaticSubjectSyntax{Kind: subject, Span: shared.SpanOf(words[:index])},
		Exact:         true,
	}}, true
}

// parseDirectedMustAttackEffect recognizes The Brothers' War chapter II directed
// forced-attack effect "Until your next turn, each creature they control attacks
// the other chosen player each combat if able." It pairs with the preceding
// "Choose two target players." clause: "they" are the two chosen players and "the
// other chosen player" is the reciprocal defender. The recognizer emits a single
// EffectDirectedMustAttack carrying the until-your-next-turn duration; lowering
// reconstructs the reciprocal directed structure from the two player targets. Any
// other group, defender, duration, or trailing clause fails closed.
func parseDirectedMustAttackEffect(sentence Sentence, tokens []shared.Token, atoms Atoms) ([]EffectSyntax, bool) {
	remaining, leadingDuration := stripLeadingDurationClause(tokens, atoms)
	if leadingDuration != EffectDurationUntilYourNextTurn {
		return nil, false
	}
	words := make([]shared.Token, 0, len(remaining))
	for _, token := range remaining {
		if token.Kind == shared.Period {
			continue
		}
		words = append(words, token)
	}
	want := []string{
		"each", "creature", "they", "control",
		"attacks", "the", "other", "chosen", "player",
		"each", "combat", "if", "able",
	}
	if len(words) != len(want) {
		return nil, false
	}
	for i, word := range want {
		if !equalWord(words[i], word) {
			return nil, false
		}
	}
	return []EffectSyntax{{
		Kind:       EffectDirectedMustAttack,
		Span:       sentence.Span,
		ClauseSpan: sentence.Span,
		VerbSpan:   words[4].Span,
		Text:       sentence.Text,
		Tokens:     append([]shared.Token(nil), tokens...),
		Context:    EffectContextController,
		Duration:   EffectDurationUntilYourNextTurn,
		Exact:      true,
	}}, true
}

// parseAttackTaxEffect recognizes the resolving, duration-bounded attack-tax
// effect "Until your next turn, creatures can't attack you unless their
// controller pays {N} for each of those creatures." (Summon: Yojimbo chapters
// II/III). The leading "Until your next turn," duration clause distinguishes this
// resolving installation from the continuous Propaganda-style static attack tax,
// which the static-declaration parser handles. Lowering applies a
// RuleEffectAttackTax for the recognized duration. Any other duration, defending
// scope ("or planeswalkers you control"), or trailing wording fails closed and
// flows through the generic effect parser.
func parseAttackTaxEffect(sentence Sentence, tokens []shared.Token, atoms Atoms) ([]EffectSyntax, bool) {
	remaining, leadingDuration := stripLeadingDurationClause(tokens, atoms)
	if leadingDuration != EffectDurationUntilYourNextTurn {
		return nil, false
	}
	words := make([]shared.Token, 0, len(remaining))
	for _, token := range remaining {
		if token.Kind == shared.Period {
			continue
		}
		words = append(words, token)
	}
	if len(words) != 14 ||
		words[8].Kind != shared.Symbol ||
		!equalWord(words[0], "creatures") ||
		!equalWord(words[1], "can't") ||
		!equalWord(words[2], "attack") ||
		!equalWord(words[3], "you") ||
		!equalWord(words[4], "unless") ||
		!equalWord(words[5], "their") ||
		!equalWord(words[6], "controller") ||
		!equalWord(words[7], "pays") ||
		!equalWord(words[9], "for") ||
		!equalWord(words[10], "each") ||
		!equalWord(words[11], "of") ||
		!equalWord(words[12], "those") ||
		!equalWord(words[13], "creatures") {
		return nil, false
	}
	amount, ok := staticGenericSymbolValue(words[8].Text)
	if !ok || amount <= 0 {
		return nil, false
	}
	return []EffectSyntax{{
		Kind:             EffectAttackTax,
		Span:             sentence.Span,
		ClauseSpan:       sentence.Span,
		VerbSpan:         words[2].Span,
		Text:             sentence.Text,
		Tokens:           append([]shared.Token(nil), tokens...),
		Context:          EffectContextController,
		Duration:         EffectDurationUntilYourNextTurn,
		AttackTaxGeneric: amount,
		Exact:            true,
	}}, true
}

// (Mistrise Village), the all-spells form "Spells you cast this turn can't be
// countered." (Domri, Anarch of Bolas), and the equivalent "Spells you control
// can't be countered this turn." (Veil of Summer). The leading "The next" marks
// the single-next-spell variant; a bare "Spells" marks the every-spell-this-turn
// variant. The subject verb is "cast" or "control" and the duration "this turn"
// may precede or follow "can't be countered"; both order the same
// controller-scoped uncounterable buff. The buff applies to the controller's own
// spells, so any other subject, a type filter, a negation, or extra wording
// fails closed and flows through the generic effect parser.
func parseSpellsCantBeCounteredEffect(sentence Sentence, tokens []shared.Token) ([]EffectSyntax, bool) {
	words := make([]shared.Token, 0, len(tokens))
	for _, token := range tokens {
		if token.Kind == shared.Period {
			continue
		}
		words = append(words, token)
	}
	index := 0
	nextOnly := false
	switch {
	case len(words) >= 3 && equalWord(words[0], "the") && equalWord(words[1], "next") && equalWord(words[2], "spell"):
		nextOnly = true
		index = 3
	case len(words) >= 1 && equalWord(words[0], "spells"):
		index = 1
	default:
		return nil, false
	}
	if index+1 >= len(words) || !equalWord(words[index], "you") {
		return nil, false
	}
	verbToken := words[index+1]
	if !equalWord(verbToken, "cast") && !equalWord(verbToken, "control") {
		return nil, false
	}
	if !spellsCantBeCounteredTail(words[index+2:]) {
		return nil, false
	}
	return []EffectSyntax{{
		Kind:                          EffectSpellsCantBeCountered,
		Span:                          sentence.Span,
		ClauseSpan:                    sentence.Span,
		VerbSpan:                      verbToken.Span,
		Text:                          sentence.Text,
		Tokens:                        append([]shared.Token(nil), tokens...),
		Context:                       EffectContextController,
		Duration:                      EffectDurationThisTurn,
		SpellsCantBeCounteredNextOnly: nextOnly,
		Exact:                         true,
	}}, true
}

// reconcileRetargetSentenceTargets aligns the sentence's target list with the
// single clean spell target that parseChangeTargetRetargetEffect extracted. The
// "Change the target of target spell with a single target." wording makes
// parseTargets manufacture several spurious targets from its "target" nouns, but
// the redirect lowering keys off the ability's target list, so the spurious
// entries are replaced here with the one spell target the recognizer chose.
func reconcileRetargetSentenceTargets(sentence *Sentence) {
	if len(sentence.Effects) != 1 ||
		sentence.Effects[0].Kind != EffectChooseNewTargets ||
		len(sentence.Effects[0].Targets) != 1 ||
		len(sentence.Targets) == 1 {
		return
	}
	sentence.Targets = append([]TargetSyntax(nil), sentence.Effects[0].Targets...)
}

// parseChangeTargetRetargetEffect recognizes the redirect wording "[You may]
// Change the target[s] of <selection>[ with a single target]." (Deflection,
// Swerve, Misdirection, Imp's Mischief, Bolt Bend, Reroute, Goblin Flectomancer)
// and lowers it through the same EffectChooseNewTargets primitive that powers
// "You may choose new targets for target spell." (Redirect). The generic effect
// parser cannot handle this sentence because parseTargets manufactures a spurious
// target for every "target" noun ("the target of", "a single target") and blanks
// the real spell selection, so this dedicated recognizer extracts the single
// clean target selection itself. An optional leading "You may" rides on the
// effect's Optional flag (Goblin Flectomancer's "You may change the targets of
// target instant or sorcery spell"). The selection that follows the fixed
// "Change the target[s] of" lead-in may be a spell, an activated/triggered
// ability, or a mixed spell-or-ability (retargetSelectionKind), each of which the
// EffectChooseNewTargets lowering retargets. The optional trailing "with a single
// target" qualifier is not modeled at resolution, so the lowered effect retargets
// any one targeted object of the selection, which is broader than the printed
// restriction but safe for this family. A "to <object>" redirect-to-a-named
// object (Muck Drubb) is rejected so it stays unsupported, and any other wording
// fails closed through the generic effect parser.
func parseChangeTargetRetargetEffect(sentence Sentence, tokens []shared.Token, atoms Atoms) ([]EffectSyntax, bool) {
	words := make([]shared.Token, 0, len(tokens))
	origin := make([]int, 0, len(tokens))
	for index, token := range tokens {
		if token.Kind == shared.Period {
			continue
		}
		words = append(words, token)
		origin = append(origin, index)
	}
	// An optional leading "you may" makes the retarget optional (Goblin
	// Flectomancer); it is stripped before the fixed lead-in is matched.
	optional := false
	lead := 0
	if len(words) >= 2 && equalWord(words[0], "you") && equalWord(words[1], "may") {
		optional = true
		lead = 2
	}
	// The fixed lead-in is four words ("change the target[s] of") and at least two
	// words of selection ("target spell") must follow it.
	if len(words)-lead < 6 {
		return nil, false
	}
	if !equalWord(words[lead], "change") || !equalWord(words[lead+1], "the") ||
		!equalWord(words[lead+3], "of") {
		return nil, false
	}
	if !equalWord(words[lead+2], "target") && !equalWord(words[lead+2], "targets") {
		return nil, false
	}
	selStart := lead + 4
	selEnd := len(words)
	// An optional trailing "with a single target[s]" qualifier ends the selection
	// four words early. Without it the selection runs to the final word.
	if equalWord(words[selEnd-4], "with") && equalWord(words[selEnd-3], "a") &&
		equalWord(words[selEnd-2], "single") &&
		(equalWord(words[selEnd-1], "target") || equalWord(words[selEnd-1], "targets")) {
		selEnd -= 4
	}
	if selEnd-selStart < 2 {
		return nil, false
	}
	// A trailing "to <object>" redirects to a specific named object rather than
	// freely re-choosing targets, which EffectChooseNewTargets does not model, so
	// reject any "to" inside the selection span.
	for i := selStart; i < selEnd; i++ {
		if equalWord(words[i], "to") {
			return nil, false
		}
	}
	// The selection words must be contiguous in the original tokens (no
	// intervening period) so the slice is exactly the selection phrase.
	for i := selStart; i < selEnd; i++ {
		if origin[i] != origin[selStart]+(i-selStart) {
			return nil, false
		}
	}
	spellTargets := parseTargets(tokens[origin[selStart]:origin[selEnd-1]+1], atoms)
	if len(spellTargets) != 1 || !retargetSelectionKind(spellTargets[0].Selection.Kind) {
		return nil, false
	}
	return []EffectSyntax{{
		Kind:       EffectChooseNewTargets,
		Span:       sentence.Span,
		ClauseSpan: sentence.Span,
		VerbSpan:   words[lead].Span,
		Text:       sentence.Text,
		Tokens:     append([]shared.Token(nil), tokens...),
		Context:    EffectContextController,
		Targets:    spellTargets,
		Optional:   optional,
		Exact:      true,
	}}, true
}

// retargetSelectionKind reports whether a selection between the "Change the
// target[s] of" lead-in and the "with a single target" tail names a stack object
// the EffectChooseNewTargets lowering can retarget. It admits the spell and
// spell-or-ability forms (Deflection, Bolt Bend) plus the activated/triggered
// ability forms (Reroute targets an activated ability), all of which the
// downstream counterTargetSpec resolver already accepts. Any other selection
// fails closed through the generic effect parser.
func retargetSelectionKind(kind SelectionKind) bool {
	switch kind {
	case SelectionSpell,
		SelectionActivatedAbility,
		SelectionTriggeredAbility,
		SelectionActivatedOrTriggeredAbility,
		SelectionSpellActivatedOrTriggeredAbility,
		SelectionTriggeredAbilityOrSpell:
		return true
	default:
		return false
	}
}

// spellsCantBeCounteredTail accepts the two interchangeable orderings of the
// duration and prohibition tail that follow the "Spells you cast/control"
// subject: "this turn can't be countered" (Domri) and "can't be countered this
// turn" (Veil of Summer). "cannot" is accepted as a spelling of "can't".
func spellsCantBeCounteredTail(words []shared.Token) bool {
	cantBeCountered := func(group []shared.Token) bool {
		return len(group) == 3 &&
			(equalWord(group[0], "can't") || equalWord(group[0], "cannot")) &&
			equalWord(group[1], "be") && equalWord(group[2], "countered")
	}
	thisTurn := func(group []shared.Token) bool {
		return len(group) == 2 && equalWord(group[0], "this") && equalWord(group[1], "turn")
	}
	if len(words) != 5 {
		return false
	}
	if thisTurn(words[:2]) && cantBeCountered(words[2:]) {
		return true
	}
	return cantBeCountered(words[:3]) && thisTurn(words[3:])
}

// parsePreventCombatDamageEffect recognizes the one-shot, turn-scoped combat
// damage prevention shield "Prevent all combat damage that would be dealt to
// [and dealt by] <object> this turn." (Maze of Ith, Goblin Snowman, Moonlight
// Geist), where <object> is a back-reference ("that creature", "this creature",
// "it") to a prior target or the source. PreventDamageTo/PreventDamageBy record
// the prevented directions. It also recognizes the global form "Prevent all
// combat damage that would be dealt this turn." (Spike Weaver, Holy Day), which
// prevents every combat damage event for the turn and sets PreventDamageGlobal.
// Wordings without "this turn" (continuous static prevention), with a player or
// group recipient, or with an unrecognized object fail closed and flow through
// the generic effect parser.
func parsePreventCombatDamageEffect(sentence Sentence, tokens []shared.Token, atoms Atoms) ([]EffectSyntax, bool) {
	words := make([]shared.Token, 0, len(tokens))
	for _, token := range tokens {
		if token.Kind == shared.Period {
			continue
		}
		words = append(words, token)
	}
	prefix := []string{"prevent", "all", "combat", "damage", "that", "would", "be", "dealt"}
	if len(words) < len(prefix) {
		return nil, false
	}
	for i, want := range prefix {
		if !equalWord(words[i], want) {
			return nil, false
		}
	}
	idx := len(prefix)
	if idx+2 == len(words) && equalWord(words[idx], "this") && equalWord(words[idx+1], "turn") {
		return []EffectSyntax{{
			Kind:                EffectPreventDamage,
			Span:                sentence.Span,
			ClauseSpan:          sentence.Span,
			VerbSpan:            words[0].Span,
			Text:                sentence.Text,
			Tokens:              append([]shared.Token(nil), tokens...),
			Context:             EffectContextController,
			Duration:            EffectDurationThisTurn,
			PreventDamageGlobal: true,
			Exact:               true,
		}}, true
	}
	preventTo, preventBy := false, false
	switch {
	case idx+3 < len(words) && equalWord(words[idx], "to") &&
		equalWord(words[idx+1], "and") && equalWord(words[idx+2], "dealt") && equalWord(words[idx+3], "by"):
		preventTo, preventBy = true, true
		idx += 4
	case idx+3 < len(words) && equalWord(words[idx], "by") &&
		equalWord(words[idx+1], "and") && equalWord(words[idx+2], "dealt") && equalWord(words[idx+3], "to"):
		preventTo, preventBy = true, true
		idx += 4
	case idx < len(words) && equalWord(words[idx], "to"):
		preventTo = true
		idx++
	case idx < len(words) && equalWord(words[idx], "by"):
		preventBy = true
		idx++
	default:
		return nil, false
	}
	objectStart := idx
	switch {
	case idx+1 < len(words) && (equalWord(words[idx], "that") || equalWord(words[idx], "this")) && equalWord(words[idx+1], "creature"):
		idx += 2
	case idx < len(words) && equalWord(words[idx], "it"):
		idx++
	default:
		return nil, false
	}
	objectSpan := shared.SpanOf(words[objectStart:idx])
	if idx+2 != len(words) || !equalWord(words[idx], "this") || !equalWord(words[idx+1], "turn") {
		return nil, false
	}
	references := referencesInSpan(atoms, objectSpan)
	if len(references) != 1 {
		return nil, false
	}
	return []EffectSyntax{{
		Kind:            EffectPreventDamage,
		Span:            sentence.Span,
		ClauseSpan:      sentence.Span,
		VerbSpan:        words[0].Span,
		Text:            sentence.Text,
		Tokens:          append([]shared.Token(nil), tokens...),
		Context:         EffectContextController,
		Duration:        EffectDurationThisTurn,
		PreventDamageTo: preventTo,
		PreventDamageBy: preventBy,
		References:      references,
		Exact:           true,
	}}, true
}

// parsePreventAllDamageTargetEffect recognizes the turn-scoped all-damage (not
// only combat) prevention shield that names a single target permanent as the
// damage source or recipient:
//
//	Prevent all damage target creature would deal this turn. (Shieldmage Elder,
//	Chain of Silence — active by-source)
//	Prevent all damage that would be dealt by target creature this turn.
//	(by-source, passive voice)
//	Prevent all damage that would be dealt to target creature this turn. (Oriss,
//	Samite Guardian — to-recipient)
//
// Unlike parsePreventCombatDamageEffect it covers "all damage" rather than "all
// combat damage" and names the shielded permanent by a lone target rather than a
// back-reference, so it sets PreventDamageAllTypes and carries the target. The
// target's own selection ("target attacking or blocking creature", "target
// creature") is left to the shared target-spec lowering, which fails closed on
// any selector the backend cannot represent.
func parsePreventAllDamageTargetEffect(sentence Sentence, tokens []shared.Token, _ Atoms) ([]EffectSyntax, bool) {
	words := make([]shared.Token, 0, len(tokens))
	for _, token := range tokens {
		if token.Kind == shared.Period {
			continue
		}
		words = append(words, token)
	}
	prefix := []string{"prevent", "all", "damage"}
	if len(words) < len(prefix)+3 {
		return nil, false
	}
	for i, want := range prefix {
		if !equalWord(words[i], want) {
			return nil, false
		}
	}
	// Every accepted form ends with "this turn".
	if !equalWord(words[len(words)-2], "this") || !equalWord(words[len(words)-1], "turn") {
		return nil, false
	}
	idx := len(prefix)
	preventTo, preventBy := false, false
	var targetStart, spanEnd int
	switch {
	case idx+4 < len(words) && equalWord(words[idx], "that") && equalWord(words[idx+1], "would") &&
		equalWord(words[idx+2], "be") && equalWord(words[idx+3], "dealt") &&
		(equalWord(words[idx+4], "to") || equalWord(words[idx+4], "by")):
		// Passive voice: "... that would be dealt to|by <target> this turn."
		if equalWord(words[idx+4], "to") {
			preventTo = true
		} else {
			preventBy = true
		}
		targetStart = idx + 5
		spanEnd = len(words) - 2
	default:
		// Active by-source: "... <target> would deal this turn." The atomizer
		// tends to extend the target atom through the trailing "would", so the
		// recipient span reaches up to (but excludes) "deal" to contain it.
		if !equalWord(words[len(words)-4], "would") || !equalWord(words[len(words)-3], "deal") {
			return nil, false
		}
		preventBy = true
		targetStart = idx
		spanEnd = len(words) - 3
	}
	if targetStart >= spanEnd || !equalWord(words[targetStart], "target") {
		return nil, false
	}
	targetSpan := shared.SpanOf(words[targetStart:spanEnd])
	targets := targetsInSpan(sentence.Targets, targetSpan)
	if len(targets) != 1 {
		return nil, false
	}
	return []EffectSyntax{{
		Kind:                  EffectPreventDamage,
		Span:                  sentence.Span,
		ClauseSpan:            sentence.Span,
		VerbSpan:              words[0].Span,
		Text:                  sentence.Text,
		Tokens:                append([]shared.Token(nil), tokens...),
		Context:               EffectContextController,
		Duration:              EffectDurationThisTurn,
		PreventDamageTo:       preventTo,
		PreventDamageBy:       preventBy,
		PreventDamageAllTypes: true,
		Targets:               targets,
		Exact:                 true,
	}}, true
}

// parsePreventNextDamageFromSourceEffect recognizes the one-shot "The next time
// a [color] source of your choice would deal damage to you this turn, prevent
// that damage." shield (Circle of Protection, Rune of Protection, Pentagram of
// the Ages). It prevents all of the next damage the controller would take this
// turn from a chosen source matching an optional single-color filter, then
// expires. The color is recorded text-blind as a typed Color; an absent filter
// ("a source of your choice") records no color. Any other recipient ("to any
// target", "to enchanted creature"), amount ("prevent half that damage"),
// source qualifier ("an artifact source", "of the chosen color"), or trailing
// rider fails closed and flows through the generic effect parser.
func parsePreventNextDamageFromSourceEffect(sentence Sentence, tokens []shared.Token, _ Atoms) ([]EffectSyntax, bool) {
	words := make([]shared.Token, 0, len(tokens))
	for _, token := range tokens {
		if token.Kind == shared.Period {
			continue
		}
		words = append(words, token)
	}
	head := []string{"the", "next", "time"}
	if len(words) < len(head)+1 {
		return nil, false
	}
	for i, want := range head {
		if !equalWord(words[i], want) {
			return nil, false
		}
	}
	idx := len(head)
	if !equalWord(words[idx], "a") && !equalWord(words[idx], "an") {
		return nil, false
	}
	idx++
	if idx >= len(words) {
		return nil, false
	}
	var colors []Color
	if colorValue, ok := recognizeColorWord(words[idx].Text); ok {
		colors = []Color{colorValue}
		idx++
	}
	tail := []string{"source", "of", "your", "choice", "would", "deal", "damage", "to", "you", "this", "turn"}
	if idx+len(tail) > len(words) {
		return nil, false
	}
	for i, want := range tail {
		if !equalWord(words[idx+i], want) {
			return nil, false
		}
	}
	idx += len(tail)
	closing := []string{"prevent", "that", "damage"}
	if idx+1+len(closing) != len(words) || words[idx].Kind != shared.Comma {
		return nil, false
	}
	idx++
	for i, want := range closing {
		if !equalWord(words[idx+i], want) {
			return nil, false
		}
	}
	return []EffectSyntax{{
		Kind:                        EffectPreventDamage,
		Span:                        sentence.Span,
		ClauseSpan:                  sentence.Span,
		VerbSpan:                    words[idx].Span,
		Text:                        sentence.Text,
		Tokens:                      append([]shared.Token(nil), tokens...),
		Context:                     EffectContextController,
		Duration:                    EffectDurationThisTurn,
		PreventDamageNextFromSource: true,
		PreventDamageSourceColors:   colors,
		Exact:                       true,
	}}, true
}

func parsePreventAmountDamageEffect(sentence Sentence, tokens []shared.Token, atoms Atoms) ([]EffectSyntax, bool) {
	words := make([]shared.Token, 0, len(tokens))
	for _, token := range tokens {
		if token.Kind == shared.Period {
			continue
		}
		words = append(words, token)
	}
	prefix := []string{"prevent", "the", "next"}
	if len(words) < len(prefix)+1 {
		return nil, false
	}
	for i, want := range prefix {
		if !equalWord(words[i], want) {
			return nil, false
		}
	}
	amountToken := words[len(prefix)]
	if amountToken.Kind != shared.Integer {
		return nil, false
	}
	amount, err := strconv.Atoi(amountToken.Text)
	if err != nil || amount < 1 {
		return nil, false
	}
	middle := []string{"damage", "that", "would", "be", "dealt", "to"}
	idx := len(prefix) + 1
	if idx+len(middle) > len(words) {
		return nil, false
	}
	for i, want := range middle {
		if !equalWord(words[idx+i], want) {
			return nil, false
		}
	}
	idx += len(middle)
	// The recipient run ends at the trailing "this turn".
	if idx+2 > len(words) || !equalWord(words[len(words)-2], "this") || !equalWord(words[len(words)-1], "turn") {
		return nil, false
	}
	recipientStart := idx
	recipientEnd := len(words) - 2
	if recipientStart >= recipientEnd {
		return nil, false
	}
	recipient := words[recipientStart:recipientEnd]
	recipientSpan := shared.SpanOf(recipient)

	var kind PreventDamageRecipientKind
	var refs []Reference
	var recipientTargets []TargetSyntax
	switch {
	case equalWord(recipient[0], "you") && len(recipient) == 1:
		kind = PreventDamageRecipientYou
	case equalWord(recipient[0], "it") && len(recipient) == 1,
		len(recipient) == 2 && (equalWord(recipient[0], "this") || equalWord(recipient[0], "that")) && equalWord(recipient[1], "creature"):
		refs = referencesInSpan(atoms, recipientSpan)
		if len(refs) != 1 {
			return nil, false
		}
		kind = PreventDamageRecipientSource
	case equalWord(recipient[0], "any") || equalWord(recipient[0], "target"):
		recipientTargets = targetsInSpan(sentence.Targets, recipientSpan)
		if len(recipientTargets) != 1 {
			return nil, false
		}
		kind = PreventDamageRecipientTarget
	default:
		return nil, false
	}

	return []EffectSyntax{{
		Kind:                       EffectPreventDamage,
		Span:                       sentence.Span,
		ClauseSpan:                 sentence.Span,
		VerbSpan:                   words[0].Span,
		Text:                       sentence.Text,
		Tokens:                     append([]shared.Token(nil), tokens...),
		Context:                    EffectContextController,
		Duration:                   EffectDurationThisTurn,
		Amount:                     EffectAmountSyntax{Span: amountToken.Span, Text: amountToken.Text, Value: amount, Known: true},
		PreventDamageNextRecipient: kind,
		References:                 refs,
		Targets:                    recipientTargets,
		Exact:                      true,
	}}, true
}

// parsePreventThatDamageEffect recognizes the continuous static damage
// prevention "If <source> would deal damage to you, prevent N of that damage."
// (Sphere of Law, Urza's Armor). The leading "if" damage-source condition is
// recognized separately; this recognizer claims the trailing "prevent N of that
// damage" effect clause after the condition comma and records its fixed amount.
// A clause without the leading "if", without a top-level comma, or whose effect
// run is not exactly "prevent <integer> of that damage" fails closed.
func parsePreventThatDamageEffect(sentence Sentence, tokens []shared.Token, _ Atoms) ([]EffectSyntax, bool) {
	words := make([]shared.Token, 0, len(tokens))
	for _, token := range tokens {
		if token.Kind == shared.Period {
			continue
		}
		words = append(words, token)
	}
	if len(words) == 0 || !equalWord(words[0], "if") {
		return nil, false
	}
	comma := -1
	for i := range words {
		if words[i].Kind == shared.Comma {
			comma = i
			break
		}
	}
	if comma < 0 {
		return nil, false
	}
	effect := words[comma+1:]
	if len(effect) != 5 ||
		!equalWord(effect[0], "prevent") ||
		effect[1].Kind != shared.Integer ||
		!equalWord(effect[2], "of") ||
		!equalWord(effect[3], "that") ||
		!equalWord(effect[4], "damage") {
		return nil, false
	}
	amount, err := strconv.Atoi(effect[1].Text)
	if err != nil || amount < 1 {
		return nil, false
	}
	effectSpan := shared.SpanOf(effect)
	return []EffectSyntax{{
		Kind:                    EffectPreventDamage,
		Span:                    sentence.Span,
		ClauseSpan:              effectSpan,
		VerbSpan:                effect[0].Span,
		Text:                    sentence.Text,
		Tokens:                  append([]shared.Token(nil), tokens...),
		Context:                 EffectContextController,
		PreventDamageThatAmount: amount,
		Exact:                   true,
	}}, true
}

func parseAdditionalLandPlaysEffect(sentence Sentence, tokens []shared.Token, _ Atoms) ([]EffectSyntax, bool) {
	words := make([]shared.Token, 0, len(tokens))
	for _, token := range tokens {
		if token.Kind == shared.Period {
			continue
		}
		words = append(words, token)
	}
	start := 0
	if len(words) >= 2 && equalWord(words[0], "you") && equalWord(words[1], "may") {
		start = 2
	}
	rest := words[start:]
	// Shortest match: "play an additional land this turn" (6 words).
	if len(rest) < 6 || !equalWord(rest[0], "play") {
		return nil, false
	}
	playToken := rest[0]
	rest = rest[1:]
	if equalWord(rest[0], "up") && len(rest) >= 2 && equalWord(rest[1], "to") {
		rest = rest[2:]
	}
	if len(rest) < 5 {
		return nil, false
	}
	count, ok := additionalLandCountWord(rest[0])
	if !ok || !equalWord(rest[1], "additional") {
		return nil, false
	}
	plural := count != 1
	landWord := "land"
	if plural {
		landWord = "lands"
	}
	if !equalWord(rest[2], landWord) ||
		!equalWord(rest[3], "this") ||
		!equalWord(rest[4], "turn") ||
		len(rest) != 5 {
		return nil, false
	}
	return []EffectSyntax{{
		Kind:       EffectAdditionalLandPlays,
		Span:       sentence.Span,
		ClauseSpan: sentence.Span,
		VerbSpan:   playToken.Span,
		Text:       sentence.Text,
		Tokens:     append([]shared.Token(nil), tokens...),
		Context:    EffectContextController,
		Duration:   EffectDurationThisTurn,
		Amount:     EffectAmountSyntax{Value: count, Known: true},
		Exact:      true,
	}}, true
}

// additionalLandCountWord reads the extra-land count from the determiner or
// number word preceding "additional land(s)": "a"/"an"/"one" mean a single extra
// land, and small cardinal words ("two", "three", ...) and integer literals give
// their value.
func additionalLandCountWord(token shared.Token) (int, bool) {
	if token.Kind == shared.Integer {
		value, err := strconv.Atoi(token.Text)
		if err != nil || value < 1 {
			return 0, false
		}
		return value, true
	}
	switch strings.ToLower(token.Text) {
	case "a", "an", "one":
		return 1, true
	default:
		return CardinalWordValue(token.Text)
	}
}

func parseHandLibraryPut(effect *EffectSyntax) HandLibraryPutSyntax {
	if effect.Kind != EffectPut ||
		effect.Context != EffectContextController ||
		!effect.Amount.Known || effect.Amount.Value < 1 ||
		effect.FromZone != zone.Hand ||
		effect.ToZone != zone.Library ||
		effect.Destination != EffectDestinationTop ||
		len(effect.Targets) != 0 ||
		!effectContainsWords(normalizedWords(effect.Tokens), "in", "any", "order") {
		return HandLibraryPutSyntax{}
	}
	return HandLibraryPutSyntax{Present: true}
}

// parseDigPut recognizes the impulse put clause "Put N <of them|of those cards>
// into your hand and the <rest|other> <into your graveyard|on the bottom of your
// library [in any order|in a random order]>." that follows an EffectDig look
// sentence, returning its structured fields. It returns the zero DigSyntax for
// every other effect. The structured fields it sets are revalidated
// byte-for-byte by exactDigPutEffectSyntax, so an over-broad match simply fails
// the exactness gate.
func parseDigPut(effect *EffectSyntax) DigSyntax {
	if effect.Kind != EffectPut {
		return DigSyntax{}
	}
	verb := slices.IndexFunc(effect.Tokens, func(token shared.Token) bool {
		return token.Span == effect.VerbSpan
	})
	if verb < 0 {
		return DigSyntax{}
	}
	clause := effect.Tokens[verb+1:]
	if len(clause) == 0 ||
		(!equalWord(clause[0], "one") && !equalWord(clause[0], "two") && !equalWord(clause[0], "three")) {
		return DigSyntax{}
	}
	i := 1
	var dig DigSyntax
	switch {
	case effectWordsAt(clause, i, "of", "them"):
		dig.Source = DigSourceThem
		i += 2
	case effectWordsAt(clause, i, "of", "those", "cards"):
		dig.Source = DigSourceThoseCards
		i += 3
	default:
		dig.Source = DigSourceNone
	}
	if !effectWordsAt(clause, i, "into", "your", "hand", "and", "the") {
		return DigSyntax{}
	}
	i += 5
	switch {
	case effectWordsAt(clause, i, "other"):
		dig.Singular = true
		i++
	case effectWordsAt(clause, i, "rest"):
		i++
	default:
		return DigSyntax{}
	}
	remainder, after, ok := digRemainderAt(clause, i)
	if !ok {
		return DigSyntax{}
	}
	dig.Remainder = remainder
	i = after
	if i < len(clause) && clause[i].Kind == shared.Period {
		i++
	}
	if i != len(clause) {
		return DigSyntax{}
	}
	dig.Put = true
	return dig
}

// digRemainderAt recognizes the remainder destination that follows "the
// <rest|other>" in an impulse put clause: "into your graveyard", or "on the
// bottom of your library" optionally trailed by an "in any order" / "in a random
// order" rider. It returns the matched remainder kind and the index just past
// the clause, or ok=false for any other wording.
func digRemainderAt(clause []shared.Token, start int) (DigRemainderKind, int, bool) {
	if effectWordsAt(clause, start, "into", "your", "graveyard") {
		return DigRemainderGraveyard, start + 3, true
	}
	if !effectWordsAt(clause, start, "on", "the", "bottom", "of", "your", "library") {
		return "", start, false
	}
	i := start + 6
	switch {
	case effectWordsAt(clause, i, "in", "any", "order"):
		return DigRemainderLibraryBottomAny, i + 3, true
	case effectWordsAt(clause, i, "in", "a", "random", "order"):
		return DigRemainderLibraryBottomRandom, i + 4, true
	default:
		return DigRemainderLibraryBottom, i, true
	}
}

// parseSearchSplitPut recognizes the split-destination put clause "put one
// <slot> and the other <slot>" that distributes the cards found by a preceding
// "up to two" library search across two single-card destination slots ("put one
// onto the battlefield tapped and the other into your hand"). It returns the
// zero SearchSplitSyntax for every other effect, including ordinary
// single-destination puts. Each slot is a hand or battlefield (optionally
// tapped) destination; any other wording fails closed. The structured fields it
// sets are revalidated byte-for-byte by the search exactness gate, so an
// over-broad match simply fails recognition.
func parseSearchSplitPut(effect *EffectSyntax) SearchSplitSyntax {
	if effect.Kind != EffectPut {
		return SearchSplitSyntax{}
	}
	verb := slices.IndexFunc(effect.Tokens, func(token shared.Token) bool {
		return token.Span == effect.VerbSpan
	})
	if verb < 0 {
		return SearchSplitSyntax{}
	}
	clause := effect.Tokens[verb+1:]
	if len(clause) == 0 || !equalWord(clause[0], "one") {
		return SearchSplitSyntax{}
	}
	first, i, ok := parseSearchSplitSlot(clause, 1)
	if !ok || !effectWordsAt(clause, i, "and", "the", "other") {
		return SearchSplitSyntax{}
	}
	i += 3
	second, i, ok := parseSearchSplitSlot(clause, i)
	if !ok {
		return SearchSplitSyntax{}
	}
	if i < len(clause) && clause[i].Kind == shared.Period {
		i++
	}
	if i != len(clause) {
		return SearchSplitSyntax{}
	}
	return SearchSplitSyntax{Present: true, First: first, Second: second}
}

// parseSearchSplitSlot reads one destination slot of a split put clause starting
// at index i: "onto the battlefield" with an optional trailing "tapped", or
// "into your hand". It returns the slot, the index just past it, and whether a
// slot was recognized.
func parseSearchSplitSlot(clause []shared.Token, i int) (SearchSplitSlot, int, bool) {
	switch {
	case effectWordsAt(clause, i, "onto", "the", "battlefield"):
		slot := SearchSplitSlot{ToZone: zone.Battlefield}
		i += 3
		if effectWordsAt(clause, i, "tapped") {
			slot.EntersTapped = true
			i++
		}
		return slot, i, true
	case effectWordsAt(clause, i, "into", "your", "hand"):
		return SearchSplitSlot{ToZone: zone.Hand}, i + 3, true
	default:
		return SearchSplitSlot{}, i, false
	}
}

// parseDamageRecipientPair recognizes the dual-recipient group-damage wording
// "deals N damage to each X and each Y" and returns the two recipient groups as
// separate selections. It returns nil for every other effect so the recipient
// stays single. The recipient is identified as the tokens after "damage to";
// it must split into exactly two "each <group>" phrases joined by a single
// top-level "and", and each phrase is parsed by the same parseSelection used for
// a lone group recipient. The downstream exactness gate reconstructs both halves
// and compares them byte-for-byte, so an over-broad split simply fails closed.
func parseDamageRecipientPair(kind EffectKind, clause []shared.Token, amount EffectAmountSyntax, atoms Atoms) []SelectionSyntax {
	if kind != EffectDealDamage {
		return nil
	}
	recipient, ok := damageRecipientTokens(clause)
	if !ok {
		recipient, ok = damageRecipientTokensAfterAmount(clause, amount)
		if !ok {
			return nil
		}
	}
	left, right, ok := splitEachAndEach(recipient)
	if !ok {
		return nil
	}
	leftSel := parseSelection(left, atoms)
	rightSel := parseSelection(right, atoms)
	applyOpponentTheyControl(&rightSel, right, atoms)
	return []SelectionSyntax{leftSel, rightSel}
}

// applyOpponentTheyControl scopes a dual-recipient group's second group to
// permanents the opponents control when the recipient ends in the pronoun "they
// control" — "each opponent and each creature [and planeswalker] they control"
// (Tectonic Hazard, Goblin Chainwhirler). The first recipient is "each
// opponent", so the trailing "they" binds to those opponents; parseSelection
// leaves the controller unset because "they control" is not a standalone control
// phrase. It records SelectionControllerOpponent with OpponentThey so lowering
// scopes to opponent-controlled permanents and the byte-exact reconstruction
// rebuilds the verbatim pronoun. It leaves any group that already carries an
// explicit controller untouched.
func applyOpponentTheyControl(sel *SelectionSyntax, tokens []shared.Token, atoms Atoms) {
	if sel.Controller != SelectionControllerAny {
		return
	}
	if _, ok := atoms.ControllerIn(shared.SpanOf(tokens)); ok {
		return
	}
	n := len(tokens)
	if n < 2 || !equalWord(tokens[n-2], "they") || !equalWord(tokens[n-1], "control") {
		return
	}
	sel.Controller = SelectionControllerOpponent
	sel.OpponentThey = true
}

// groupDamageRecipientFollowsAmount reports whether a deal-damage clause whose
// amount is a dynamic "equal to ..." phrase is followed by an "each <group>"
// recipient ("deals damage equal to its power to each opponent"). It gates the
// recipient-scoping rewrite so it applies only to group recipients and never
// disturbs the single-target form ("deals damage equal to its power to any
// target"), whose recipient is a chosen target parsed separately.
func groupDamageRecipientFollowsAmount(clause []shared.Token, amount EffectAmountSyntax) bool {
	recipient, ok := damageRecipientTokensAfterAmount(clause, amount)
	return ok && len(recipient) > 0 && equalWord(recipient[0], "each")
}

// damageRecipientTokensAfterAmount returns the recipient tokens of a deal-damage
// clause whose amount is a dynamic "equal to ..." phrase ("deals damage equal to
// its power to each other creature and each opponent."): everything after the
// first "to" that follows the amount span, with a trailing period removed. The
// plain damageRecipientTokens path fails for this wording because "damage" is
// followed by "equal" rather than "to", so the recipient is read from the "to"
// that introduces it after the amount. It fails closed when the amount is not a
// dynamic phrase or no recipient "to" follows it.
func damageRecipientTokensAfterAmount(clause []shared.Token, amount EffectAmountSyntax) ([]shared.Token, bool) {
	if amount.DynamicKind == EffectDynamicAmountNone {
		return nil, false
	}
	for i := 0; i+1 < len(clause); i++ {
		if clause[i].Span.Start.Offset < amount.Span.End.Offset {
			continue
		}
		if !equalWord(clause[i], "to") {
			continue
		}
		recipient := clause[i+1:]
		if len(recipient) > 0 && recipient[len(recipient)-1].Kind == shared.Period {
			recipient = recipient[:len(recipient)-1]
		}
		if len(recipient) == 0 {
			return nil, false
		}
		return recipient, true
	}
	return nil, false
}

// damageRecipientTokens returns the recipient tokens of a deal-damage clause:
// everything after the first "damage to", with a trailing period removed. It
// fails closed when "damage" is not immediately followed by "to" (for example
// the dynamic "damage equal to ... to ..." form), leaving such wordings to other
// paths.
func damageRecipientTokens(clause []shared.Token) ([]shared.Token, bool) {
	for i := 0; i+1 < len(clause); i++ {
		if equalWord(clause[i], "damage") && equalWord(clause[i+1], "to") {
			recipient := clause[i+2:]
			if len(recipient) > 0 && recipient[len(recipient)-1].Kind == shared.Period {
				recipient = recipient[:len(recipient)-1]
			}
			if len(recipient) == 0 {
				return nil, false
			}
			return recipient, true
		}
	}
	return nil, false
}

// damageRecipientReference recognizes a damage recipient that is the controller
// or owner of a referenced object (the prior removal target): "deals N damage to
// its controller", "... to its owner", "... to that <object>'s controller", or
// "... to that <object>'s owner". It uses the effect's own Tokens (the clause
// span) so the recipient is read from the verb clause alone. It returns None for
// every other recipient (a target, a group, or a dual recipient), leaving those
// to their existing paths.
// eachSourceDamageSyntax recognizes an "each <group> deals N damage to its
// controller/owner" effect, where every member of the subject group is the
// damage source dealing to the player who controls (or owns) it ("Each creature
// deals 1 damage to its controller."). It returns the parsed source-group
// selection and the recipient role (controller or owner). It fails closed
// (empty selection, None role) for every other shape: a non-damage effect, a
// subject that does not begin with "each" or does not parse to a recognized
// group, or a recipient that is not the bare "its controller"/"its owner".
func eachSourceDamageSyntax(kind EffectKind, subject, clause []shared.Token, amount EffectAmountSyntax, atoms Atoms) (SelectionSyntax, DamageRecipientReferenceKind) {
	if kind != EffectDealDamage || len(subject) == 0 || !equalWord(subject[0], "each") {
		return SelectionSyntax{}, DamageRecipientReferenceNone
	}
	recipient, ok := damageRecipientTokens(clause)
	if !ok || len(recipient) == 0 {
		return SelectionSyntax{}, DamageRecipientReferenceNone
	}
	var role DamageRecipientReferenceKind
	switch {
	case len(recipient) == 2 && equalWord(recipient[0], "its") && equalWord(recipient[1], "controller"):
		role = DamageRecipientReferenceController
	case len(recipient) == 2 && equalWord(recipient[0], "its") && equalWord(recipient[1], "owner"):
		role = DamageRecipientReferenceOwner
	case equalWord(recipient[0], "itself") && eachSelfPowerDamageAmount(amount):
		// "Each creature deals damage to itself equal to its power." The
		// recipient run begins with "itself" and is followed by the source-power
		// amount phrase; the per-member power is the amount, so the recipient is
		// the bare "itself" rather than a player.
		role = DamageRecipientReferenceItself
	default:
		return SelectionSyntax{}, DamageRecipientReferenceNone
	}
	selection := parseSelection(subject, atoms)
	if selection.Kind == SelectionUnknown {
		return SelectionSyntax{}, DamageRecipientReferenceNone
	}
	return selection, role
}

// eachSelfPowerDamageAmount reports whether amount is the source-power "equal to
// its power" form that pairs with the per-member "itself" recipient. It fails
// closed for every other amount so the self recipient stays unrecognized unless
// the per-member power amount accompanies it.
func eachSelfPowerDamageAmount(amount EffectAmountSyntax) bool {
	return amount.DynamicKind == EffectDynamicAmountSourcePower &&
		amount.DynamicForm == EffectDynamicAmountFormEqual &&
		amount.Multiplier == 1
}

func damageRecipientReference(effect *EffectSyntax) DamageRecipientReferenceKind {
	if effect.Kind != EffectDealDamage {
		return DamageRecipientReferenceNone
	}
	recipient, ok := damageRecipientTokens(effect.Tokens)
	if !ok {
		// A dynamic "deals damage equal to ... to <recipient>" clause separates
		// "damage" from "to" with the amount phrase, so the recipient is read
		// from the "to" that follows the amount span instead. Word order alone
		// gates this: only the recipient after the dynamic amount is considered,
		// so the amount's own referent ("its power", "that creature's toughness")
		// never bleeds into the recipient.
		recipient, ok = damageRecipientTokensAfterAmount(effect.Tokens, effect.Amount)
		if !ok {
			return DamageRecipientReferenceNone
		}
	}
	// "deals N damage to you" names the source's own controller. The lone "you"
	// recipient carries no object subject, so it is recognized before the
	// referenced-object controller/owner forms below.
	if len(recipient) == 1 && equalWord(recipient[0], "you") {
		return DamageRecipientReferenceYou
	}
	// "deals N damage to that player" names the triggering event's player, the
	// punisher recipient of "Whenever an opponent draws a card, ~ deals N damage
	// to that player." (Underworld Dreams, Megrim). The "that player" reference
	// binds to the event player downstream; lowering resolves it accordingly.
	if len(recipient) == 2 && equalWord(recipient[0], "that") && equalWord(recipient[1], "player") {
		return DamageRecipientReferenceThatPlayer
	}
	// "deals N damage to that creature" names the triggering event's related
	// combat permanent, the recipient of "Whenever this creature blocks or
	// becomes blocked by a creature, ~ deals N damage to that creature."
	// (Inferno Elemental). The "that creature" reference binds to the event's
	// related permanent downstream; lowering resolves it accordingly, and the
	// binding fails closed unless the trigger supplies a related combat permanent.
	if len(recipient) == 2 && equalWord(recipient[0], "that") && equalWord(recipient[1], "creature") {
		return DamageRecipientReferenceThatCreature
	}
	// "deals N damage to the player or planeswalker it's attacking" (or "... that
	// creature is attacking") names the defending player or planeswalker of the
	// triggering attack, the recipient of "Whenever this creature attacks, ~
	// deals N damage to the player or planeswalker it's attacking." (Scorch
	// Spitter, Cavalcade of Calamity). The recipient binds to the defending
	// player of the triggering attack event; lowering resolves it accordingly,
	// and the binding fails closed unless the trigger is an attack event.
	if attackedDefenderRecipient(recipient) {
		return DamageRecipientReferenceAttackedDefender
	}
	if len(recipient) < 2 {
		return DamageRecipientReferenceNone
	}
	role, ok := referencedControllerOwnerRecipient(recipient)
	if !ok {
		return DamageRecipientReferenceNone
	}
	return role
}

// attackedDefenderRecipient reports whether the recipient run names the player
// or planeswalker the triggering attacker is attacking, in either Oracle
// wording: "the player or planeswalker it's attacking" or "the player or
// planeswalker that creature is attacking". It fails closed for every other run
// so only these two exact defending-player recipients are recognized.
func attackedDefenderRecipient(recipient []shared.Token) bool {
	if len(recipient) < 5 {
		return false
	}
	if !equalWord(recipient[0], "the") ||
		!equalWord(recipient[1], "player") ||
		!equalWord(recipient[2], "or") ||
		!equalWord(recipient[3], "planeswalker") {
		return false
	}
	tail := recipient[4:]
	if len(tail) == 2 && equalWord(tail[0], "it's") && equalWord(tail[1], "attacking") {
		return true
	}
	if len(tail) == 4 &&
		equalWord(tail[0], "that") &&
		equalWord(tail[1], "creature") &&
		equalWord(tail[2], "is") &&
		equalWord(tail[3], "attacking") {
		return true
	}
	return false
}

// parseDamageRiders assembles the ordered follow-on damage riders of a
// deal-damage clause into one typed list, in Oracle order: the self rider, the
// target-controller/owner rider, then the second-target rider. Each detection
// helper fails closed for its non-matching wordings, so a clause with no rider
// yields an empty list. The order matches the order lowering emits the rider
// Damage instructions.
func parseDamageRiders(effect *EffectSyntax) []DamageRiderSyntax {
	var riders []DamageRiderSyntax
	if value, ok := damageSelfRider(effect); ok {
		riders = append(riders, DamageRiderSyntax{
			Recipient: DamageRiderRecipientYou,
			Value:     value,
		})
	}
	if value, role := damageTargetControllerRider(effect); role != DamageRecipientReferenceNone {
		riders = append(riders, DamageRiderSyntax{
			Recipient:     DamageRiderRecipientTargetController,
			ReferenceRole: role,
			Value:         value,
		})
	}
	if value, dynamic, ok := damageSecondTargetRider(effect); ok {
		riders = append(riders, DamageRiderSyntax{
			Recipient: DamageRiderRecipientSecondTarget,
			Value:     value,
			Dynamic:   dynamic,
		})
	}
	return riders
}

// damageSelfRider recognizes a "... and N damage to you" self-damage rider
// appended to a deal-damage clause whose primary recipient is its single target,
// as in "deals 4 damage to any target and 2 damage to you." It returns the fixed
// rider amount N (>= 1) and ok=true only when the clause ends with the exact
// "and <number> damage to you" suffix. It fails closed for every other ending
// (a non-"you" recipient, a missing leading "and", a non-numeric amount), so the
// dual-group "each X and each Y" recipient and the standalone "to you" recipient
// keep their existing paths.
func damageSelfRider(effect *EffectSyntax) (int, bool) {
	if effect.Kind != EffectDealDamage {
		return 0, false
	}
	tokens := effect.Tokens
	if len(tokens) > 0 && tokens[len(tokens)-1].Kind == shared.Period {
		tokens = tokens[:len(tokens)-1]
	}
	n := len(tokens)
	if n < 5 {
		return 0, false
	}
	if !equalWord(tokens[n-1], "you") ||
		!equalWord(tokens[n-2], "to") ||
		!equalWord(tokens[n-3], "damage") ||
		!equalWord(tokens[n-5], "and") {
		return 0, false
	}
	value, ok := damageRiderAmountValue(tokens[n-4])
	if !ok || value < 1 {
		return 0, false
	}
	return value, true
}

// damageRiderAmountValue reads the fixed numeric value of a self-damage rider
// amount token, accepting both an integer literal ("2") and a small cardinal
// word ("two"). It returns ok=false for any non-numeric token.
func damageRiderAmountValue(token shared.Token) (int, bool) {
	if token.Kind == shared.Integer {
		value, err := strconv.Atoi(token.Text)
		if err != nil {
			return 0, false
		}
		return value, true
	}
	return CardinalWordValue(token.Text)
}

// damageTargetControllerRider recognizes a "... and B damage to that creature's
// controller/owner" rider appended to a single-target deal-damage clause, as in
// "Chandra's Outrage deals 4 damage to target creature and 2 damage to that
// creature's controller." It returns the fixed rider amount B (>= 1) and the
// recipient role (controller or owner of the primary target). It fails closed
// (None) for every other ending, including the "to you" self rider and the
// dual-group "each X and each Y" recipient, which keep their existing paths.
func damageTargetControllerRider(effect *EffectSyntax) (int, DamageRecipientReferenceKind) {
	value, recipient, _ := targetControllerDamageRiderTokens(effect)
	return value, recipient
}

// targetControllerDamageRiderTokens detects the "... and B damage to that
// creature's controller/owner" rider suffix and returns the rider amount, the
// recipient role, and the recipient tokens (for exact reconstruction). It fails
// closed (ok=false) for every other ending.
func targetControllerDamageRiderTokens(effect *EffectSyntax) (int, DamageRecipientReferenceKind, []shared.Token) {
	if effect.Kind != EffectDealDamage {
		return 0, DamageRecipientReferenceNone, nil
	}
	tokens := effect.Tokens
	if len(tokens) > 0 && tokens[len(tokens)-1].Kind == shared.Period {
		tokens = tokens[:len(tokens)-1]
	}
	n := len(tokens)
	// The recipient phrase is "its controller/owner" (2 tokens) or "that
	// <noun>'s controller/owner" (3 tokens), preceded by "and <number> damage
	// to" (4 tokens).
	for _, recipientLen := range []int{2, 3} {
		if n < recipientLen+4 {
			continue
		}
		recipient := tokens[n-recipientLen:]
		role, ok := referencedControllerOwnerRecipient(recipient)
		if !ok {
			continue
		}
		head := n - recipientLen
		if !equalWord(tokens[head-4], "and") ||
			!equalWord(tokens[head-2], "damage") ||
			!equalWord(tokens[head-1], "to") {
			continue
		}
		value, ok := damageRiderAmountValue(tokens[head-3])
		if !ok || value < 1 {
			continue
		}
		return value, role, recipient
	}
	return 0, DamageRecipientReferenceNone, nil
}

// referencedControllerOwnerRecipient reports whether the recipient tokens name
// the controller or owner of a referenced object — "its controller", "its
// owner", "that <noun>'s controller", "that <noun>'s owner", "the <noun>'s
// controller", or "the <noun>'s owner" — and returns the matching recipient
// role. It fails closed (None) for any other phrase.
func referencedControllerOwnerRecipient(recipient []shared.Token) (DamageRecipientReferenceKind, bool) {
	if len(recipient) < 2 {
		return DamageRecipientReferenceNone, false
	}
	role := recipient[len(recipient)-1]
	subject := recipient[:len(recipient)-1]
	subjectIsReferencedObject := len(subject) == 1 && equalWord(subject[0], "its") ||
		len(subject) == 2 &&
			(equalWord(subject[0], "that") || equalWord(subject[0], "the")) &&
			referencePossessiveObjectNoun(subject[1])
	if !subjectIsReferencedObject {
		return DamageRecipientReferenceNone, false
	}
	switch {
	case equalWord(role, "controller"):
		return DamageRecipientReferenceController, true
	case equalWord(role, "owner"):
		return DamageRecipientReferenceOwner, true
	default:
		return DamageRecipientReferenceNone, false
	}
}

// damageSecondTargetRider recognizes a "... and B damage to <second target>"
// rider appended to a single-target deal-damage clause whose second clause names
// its own target, as in "Hungry Flames deals 3 damage to target creature and 2
// damage to target player or planeswalker." It requires the clause to carry
// exactly two parsed targets and the rider suffix "and <number> damage to" to
// land immediately before the second target's span. It returns the fixed rider
// amount B (>= 1), whether the rider amount is the variable "X" matching the
// primary dynamic amount ("deals X damage to any target and X damage to any
// other target", The Brothers' War chapter III), and ok=true, failing closed for
// every other shape so single-target and group-recipient clauses keep their
// existing paths.
func damageSecondTargetRider(effect *EffectSyntax) (value int, dynamic, ok bool) {
	if effect.Kind != EffectDealDamage || len(effect.Targets) != 2 {
		return 0, false, false
	}
	tokens := effect.Tokens
	if len(tokens) > 0 && tokens[len(tokens)-1].Kind == shared.Period {
		tokens = tokens[:len(tokens)-1]
	}
	secondStart := effect.Targets[1].Span.Start.Offset
	for i := 0; i+4 < len(tokens); i++ {
		if !equalWord(tokens[i], "and") {
			continue
		}
		if !equalWord(tokens[i+2], "damage") || !equalWord(tokens[i+3], "to") ||
			tokens[i+4].Span.Start.Offset != secondStart {
			continue
		}
		if equalWord(tokens[i+1], "x") {
			return 0, true, true
		}
		amount, valueOK := damageRiderAmountValue(tokens[i+1])
		if !valueOK || amount < 1 {
			continue
		}
		return amount, false, true
	}
	return 0, false, false
}

// splitEachAndEach splits recipient tokens at a single top-level "and" into two
// phrases that each begin with "each". It fails closed for any other shape (no
// "and", more than one "and", or a half that does not start with "each"), so
// single recipients and unsupported compounds are left to the single-recipient
// path.
func splitEachAndEach(recipient []shared.Token) (left, right []shared.Token, ok bool) {
	andIndex := -1
	for i := 0; i+1 < len(recipient); i++ {
		if !equalWord(recipient[i], "and") {
			continue
		}
		if !equalWord(recipient[i+1], "each") {
			continue
		}
		// Split on the first "and each" boundary so a union noun inside the
		// second group ("each creature and planeswalker they control") stays
		// whole; a second "and" then belongs to the union rather than a third
		// recipient. A boundary "and" not followed by "each" is part of a noun
		// union and never separates recipients.
		andIndex = i
		break
	}
	if andIndex <= 0 || andIndex >= len(recipient)-1 {
		return nil, nil, false
	}
	left = recipient[:andIndex]
	right = recipient[andIndex+1:]
	if !equalWord(left[0], "each") || !equalWord(right[0], "each") {
		return nil, nil, false
	}
	return left, right, true
}

func legacyExactManaBody(effect *EffectSyntax, sentence Sentence) bool {
	if effect.Kind != EffectAddMana || len(semanticEffectTokens(sentence.Tokens)) != len(sentence.Tokens) {
		return false
	}
	direct := len(effect.Tokens) > 0 && equalWord(effect.Tokens[0], "add")
	optionalController := len(effect.Tokens) > 2 &&
		effectWordsAt(effect.Tokens, 0, "you", "may", "add")
	if !direct && !optionalController {
		return false
	}
	return effect.Mana.AnyColor || effect.Mana.CommanderIdentity || effect.Mana.LandsProduce || effect.Mana.FilterPair || effect.Mana.ColorsAmongControlled || len(effect.Mana.Symbols) != 0
}

func effectWithinCondition(tokens []shared.Token, index int) bool {
	for i := index - 1; i >= 0; i-- {
		if tokens[i].Kind == shared.Comma || tokens[i].Kind == shared.Period || tokens[i].Kind == shared.Semicolon {
			return false
		}
		if equalWord(tokens[i], "if") || equalWord(tokens[i], "unless") {
			return true
		}
	}
	return false
}

// entersColorChoiceSyntax recognizes the self entry color-choice clause "choose
// a color ." (unconstrained) or "choose a color other than <color> ." (a single
// forbidden basic color, the Gate/Thriving land cycle) following an "As this
// <permanent> enters," verb. The enters verb is shared by many entry constructs,
// so this matches only these exact color-choice clauses; non-color choices fail
// closed. The returned color is the forbidden color for the "other than" form,
// or empty otherwise.
func entersColorChoiceSyntax(kind EffectKind, clause []shared.Token) (bool, mana.Color) {
	if kind != EffectEnterTapped {
		return false, ""
	}
	body := clause
	if len(body) > 0 && body[0].Kind == shared.Comma {
		body = body[1:]
	}
	if len(body) == 4 &&
		equalWord(body[0], "choose") &&
		equalWord(body[1], "a") &&
		equalWord(body[2], "color") &&
		body[3].Text == "." {
		return true, ""
	}
	if len(body) == 7 &&
		equalWord(body[0], "choose") &&
		equalWord(body[1], "a") &&
		equalWord(body[2], "color") &&
		equalWord(body[3], "other") &&
		equalWord(body[4], "than") &&
		body[6].Text == "." {
		if forbidden, ok := basicColorWord(body[5]); ok {
			return true, forbidden
		}
	}
	return false, ""
}

// entersTypeChoiceSyntax recognizes the self entry creature-type-choice clause
// "choose a creature type ." following an "As this <permanent> enters," verb.
// The enters verb is shared by many entry constructs, so this matches only this
// exact clause; other choices fail closed.
func entersTypeChoiceSyntax(kind EffectKind, clause []shared.Token) bool {
	if kind != EffectEnterTapped {
		return false
	}
	body := clause
	if len(body) > 0 && body[0].Kind == shared.Comma {
		body = body[1:]
	}
	return len(body) == 5 &&
		equalWord(body[0], "choose") &&
		equalWord(body[1], "a") &&
		equalWord(body[2], "creature") &&
		equalWord(body[3], "type") &&
		body[4].Text == "."
}

// basicColorWord maps a single English basic color word to its typed mana color.
// It fails closed on any other token so unrecognized color words leave the entry
// choice unconstrained.
func basicColorWord(token shared.Token) (mana.Color, bool) {
	switch {
	case equalWord(token, "white"):
		return mana.W, true
	case equalWord(token, "blue"):
		return mana.U, true
	case equalWord(token, "black"):
		return mana.B, true
	case equalWord(token, "red"):
		return mana.R, true
	case equalWord(token, "green"):
		return mana.G, true
	default:
		return "", false
	}
}

func entersWithCountersSyntax(kind EffectKind, clause []shared.Token) bool {
	if kind != EffectEnterTapped || len(clause) < 4 ||
		!equalWord(clause[0], "with") ||
		!equalWord(clause[len(clause)-3], "on") ||
		!equalWord(clause[len(clause)-2], "it") ||
		clause[len(clause)-1].Text != "." {
		return false
	}
	for _, token := range clause[1 : len(clause)-3] {
		if equalWord(token, "counter") || equalWord(token, "counters") {
			return true
		}
	}
	return false
}

// groupEntersTappedPermanentType maps a plural permanent-type noun used as the
// subject of a static "<permanents> enter tapped" replacement to its runtime
// card type. It reports ok=false for any word that is not a recognized
// permanent-type plural so the caller's type list fails closed.
func groupEntersTappedPermanentType(word string) (types.Card, bool) {
	switch strings.ToLower(word) {
	case "creatures":
		return types.Creature, true
	case "lands":
		return types.Land, true
	case "artifacts":
		return types.Artifact, true
	case "enchantments":
		return types.Enchantment, true
	case "planeswalkers":
		return types.Planeswalker, true
	default:
		return "", false
	}
}

// parseEntersAsCopyEffect recognizes a self enters-the-battlefield replacement
// that has the permanent enter as a copy of another permanent chosen as it
// enters ("You may have this creature enter the battlefield as a copy of any
// creature on the battlefield.", Clone; CR 706). The copied-permanent filter is
// the noun phrase after "as a copy of", up to an optional ", except <rider>"
// clause. Only the "isn't legendary" and "is an <type> in addition to its other
// types" copiable riders are recognized; any other rider fails closed so the
// card stays unsupported.
func parseEntersAsCopyEffect(sentence Sentence, tokens []shared.Token, atoms Atoms) ([]EffectSyntax, bool) {
	body := semanticEffectTokens(tokens)
	if len(body) < 5 || body[len(body)-1].Kind != shared.Period {
		return nil, false
	}
	// An "As this <permanent> enters," replacement prefix frames the temporary
	// "become a copy ... until end of turn" form (Cursed Mirror). Strip it so the
	// shared subject scan below sees the "you may have it ..." body, and record
	// that the enter verb was consumed by the prefix.
	viaAsEnters := false
	if afterPrefix, ok := entersAsCopyAsEntersPrefix(body); ok {
		body = body[afterPrefix:]
		viaAsEnters = true
	}
	words := normalizedWords(body)
	// Only a self enters-as-copy is supported: "You may have this <permanent>
	// enter ..." or "This <permanent> enters ...". Group forms such as
	// "Creatures you control enter as a copy of this creature." (Essence of the
	// Wild) have a different subject and fail closed.
	selfSubject := len(words) >= 1 && words[0] == "this" ||
		len(words) >= 3 && words[0] == "you" && words[1] == "may" && words[2] == "have"
	if !selfSubject {
		return nil, false
	}
	copyIndex := -1
	for i := 2; i+1 < len(body); i++ {
		if equalWord(body[i], "copy") && equalWord(body[i-1], "a") && equalWord(body[i+1], "of") &&
			(equalWord(body[i-2], "as") || equalWord(body[i-2], "become")) {
			copyIndex = i
			break
		}
	}
	if copyIndex < 0 {
		return nil, false
	}
	enters := viaAsEnters
	entersTapped := false
	for i := 0; i < copyIndex; i++ {
		if equalWord(body[i], "tapped") {
			// "enter tapped as a copy" (Vesuva) also taps the permanent as it
			// enters the battlefield as its chosen copy; record it so the copy
			// applies the tapped state after the optional choice is confirmed.
			entersTapped = true
		}
		if equalWord(body[i], "enter") || equalWord(body[i], "enters") {
			enters = true
		}
	}
	if !enters {
		return nil, false
	}
	filterStart := copyIndex + 2
	filterEnd := len(body) - 1
	var notLegendary bool
	var addTypes []types.Card
	var addSubtypes []types.Sub
	var addKeywords []KeywordKind
	var conditionalCounters []EntersAsCopyConditionalCounter
	if exceptIndex := entersAsCopyExceptIndex(body, filterStart); exceptIndex >= 0 {
		riders, ok := parseEntersAsCopyRider(body[exceptIndex+1:len(body)-1], atoms)
		if !ok {
			return nil, false
		}
		notLegendary = riders.notLegendary
		addTypes = riders.addTypes
		addSubtypes = riders.addSubtypes
		addKeywords = riders.addKeywords
		conditionalCounters = riders.conditionalCounters
		filterEnd = exceptIndex
	}
	filter := body[filterStart:filterEnd]
	for len(filter) > 0 && filter[len(filter)-1].Kind == shared.Comma {
		filter = filter[:len(filter)-1]
	}
	// "become a copy ... until end of turn" makes the copy a temporary effect
	// (Cursed Mirror); strip the trailing duration phrase and record it so
	// lowering scopes the copy effect to end of turn.
	untilEndOfTurn := false
	if trimmed, ok := trimTrailingUntilEndOfTurn(filter); ok {
		filter = trimmed
		untilEndOfTurn = true
	}
	// Only battlefield permanents can be copied by this runtime, so require an
	// explicit battlefield or "you control" scope; graveyard/hand sources such
	// as Body Double ("any creature card in a graveyard") fail closed.
	if !entersAsCopyFilterOnBattlefield(filter) {
		return nil, false
	}
	filter = trimTrailingZonePhrase(filter)
	if len(filter) == 0 {
		return nil, false
	}
	optional := words[0] == "you"
	effect := EffectSyntax{
		Kind:                     EffectEnterAsCopy,
		Context:                  EffectContextController,
		Span:                     sentence.Span,
		ClauseSpan:               sentence.Span,
		Text:                     sentence.Text,
		Tokens:                   append([]shared.Token(nil), body...),
		Selection:                parseSelection(filter, atoms),
		EntersAsCopy:             true,
		EntersAsCopyOptional:     optional,
		EntersAsCopyNotLegendary: notLegendary,
		EntersAsCopyAddTypes:     addTypes,
		EntersAsCopyAddSubtypes:  addSubtypes,

		EntersAsCopyConditionalCounters: conditionalCounters,
		EntersAsCopyUntilEndOfTurn:      untilEndOfTurn,
		EntersAsCopyAddKeywords:         addKeywords,
		EntersAsCopyTapped:              entersTapped,
	}
	return []EffectSyntax{effect}, true
}

// parseDevourEffect recognizes the canonical as-enters replacement that the
// printed "Devour N" keyword expands to (CR 702.81): "As this creature enters,
// you may sacrifice any number of creatures, then it enters with N +1/+1
// counters on it for each creature sacrificed." The wording is produced solely
// by expandDevourKeyword, so this matches that exact sentence and recovers the
// per-sacrificed-creature multiplier N; any other sentence fails closed.
func parseDevourEffect(sentence Sentence, tokens []shared.Token, atoms Atoms) ([]EffectSyntax, bool) {
	body := semanticEffectTokens(tokens)
	subject, n, ok := devourSentenceParse(body)
	if !ok {
		return nil, false
	}
	effect := EffectSyntax{
		Kind:                   EffectDevour,
		Context:                EffectContextController,
		Span:                   sentence.Span,
		ClauseSpan:             sentence.Span,
		Text:                   sentence.Text,
		Tokens:                 append([]shared.Token(nil), body...),
		EntersDevour:           true,
		EntersDevourMultiplier: n,
		EntersDevourType:       subject.cardType,
		EntersDevourSubtype:    subject.subtype,
	}
	effect.Exact = exactEffectSyntax(&effect)
	return []EffectSyntax{effect}, true
}

// devourSentenceParse matches the canonical Devour expansion token sequence and
// returns the sacrificed-permanent subject together with its +1/+1 counter
// multiplier N. The fixed framing words are validated and the sacrificed-noun
// positions identify the subject (creature, artifact, land, or Food); N is the
// integer that precedes the "+1/+1 counters" phrase. Any deviation from the
// canonical wording fails closed.
func devourSentenceParse(body []shared.Token) (devourSubject, int, bool) {
	words := normalizedWords(body)
	if len(words) != 22 {
		return devourSubject{}, 0, false
	}
	prefix := []string{"as", "this", "creature", "enters", "you", "may", "sacrifice", "any", "number", "of"}
	if !slices.Equal(words[:10], prefix) {
		return devourSubject{}, 0, false
	}
	mid := []string{"then", "it", "enters", "with", "counters", "on", "it", "for", "each"}
	if !slices.Equal(words[11:20], mid) {
		return devourSubject{}, 0, false
	}
	if words[21] != "sacrificed" {
		return devourSubject{}, 0, false
	}
	subject, ok := devourSubjectByNouns(words[10], words[20])
	if !ok {
		return devourSubject{}, 0, false
	}
	n, ok := devourSentenceMultiplier(body)
	if !ok {
		return devourSubject{}, 0, false
	}
	return subject, n, true
}

// devourSentenceMultiplier returns the +1/+1 counter multiplier N of a canonical
// Devour expansion: the integer that precedes the "+1/+1 counters" phrase.
func devourSentenceMultiplier(body []shared.Token) (int, bool) {
	for i := 0; i+5 < len(body); i++ {
		if body[i].Kind == shared.Integer &&
			body[i+1].Kind == shared.Plus &&
			body[i+2].Kind == shared.Integer && body[i+2].Text == "1" &&
			body[i+3].Kind == shared.Slash &&
			body[i+4].Kind == shared.Plus &&
			body[i+5].Kind == shared.Integer && body[i+5].Text == "1" {
			n, err := strconv.Atoi(body[i].Text)
			if err != nil || n <= 0 {
				return 0, false
			}
			return n, true
		}
	}
	return 0, false
}

// parseTributeEffect recognizes the canonical as-enters replacement that the
// printed "Tribute N" keyword expands to (CR 702.110): "As this creature enters,
// an opponent of your choice may put N +1/+1 counters on it." The wording is
// produced solely by expandTributeKeyword, so this matches that exact sentence
// and recovers the +1/+1 counter count N; any other sentence fails closed.
func parseTributeEffect(sentence Sentence, tokens []shared.Token, atoms Atoms) ([]EffectSyntax, bool) {
	body := semanticEffectTokens(tokens)
	n, ok := tributeSentenceCount(body)
	if !ok {
		return nil, false
	}
	effect := EffectSyntax{
		Kind:               EffectTribute,
		Context:            EffectContextController,
		Span:               sentence.Span,
		ClauseSpan:         sentence.Span,
		Text:               sentence.Text,
		Tokens:             append([]shared.Token(nil), body...),
		EntersTribute:      true,
		EntersTributeCount: n,
	}
	effect.Exact = exactEffectSyntax(&effect)
	return []EffectSyntax{effect}, true
}

// tributeSentenceCount matches the exact canonical Tribute expansion token
// sequence and returns its +1/+1 counter count N. N is the integer that precedes
// the "+1/+1 counters" phrase. Any deviation from the canonical wording fails
// closed.
func tributeSentenceCount(body []shared.Token) (int, bool) {
	words := normalizedWords(body)
	expected := []string{
		"as", "this", "creature", "enters",
		"an", "opponent", "of", "your", "choice", "may", "put",
		"counters", "on", "it",
	}
	if !slices.Equal(words, expected) {
		return 0, false
	}
	for i := 0; i+5 < len(body); i++ {
		if body[i].Kind == shared.Integer &&
			body[i+1].Kind == shared.Plus &&
			body[i+2].Kind == shared.Integer && body[i+2].Text == "1" &&
			body[i+3].Kind == shared.Slash &&
			body[i+4].Kind == shared.Plus &&
			body[i+5].Kind == shared.Integer && body[i+5].Text == "1" {
			n, err := strconv.Atoi(body[i].Text)
			if err != nil || n <= 0 {
				return 0, false
			}
			return n, true
		}
	}
	return 0, false
}

// "This artifact becomes a copy of target ... until end of turn.", Mirage
// Mirror; CR 706). The copied-permanent target is left as the ordinary "target"
// selector for the target machinery to extract; only the "until end of turn"
// duration and the "except it has this ability" / "except it has <keyword>"
// copiable riders are recognized, and any other rider fails closed.
func parseBecomeCopyEffect(sentence Sentence, tokens []shared.Token, atoms Atoms) ([]EffectSyntax, bool) {
	body := semanticEffectTokens(tokens)
	if len(body) < 6 || body[len(body)-1].Kind != shared.Period {
		return nil, false
	}
	words := normalizedWords(body)
	// The subject must be the source permanent itself ("This <permanent> becomes
	// a copy of ..."). Group or other-subject copies fail closed.
	if len(words) < 2 || words[0] != "this" {
		return nil, false
	}
	copyIndex := -1
	for i := 3; i+2 < len(body); i++ {
		if equalWord(body[i], "copy") && equalWord(body[i-1], "a") && equalWord(body[i-2], "becomes") &&
			equalWord(body[i+1], "of") && equalWord(body[i+2], "target") {
			copyIndex = i
			break
		}
	}
	if copyIndex < 0 {
		return nil, false
	}
	rest := body[copyIndex+3 : len(body)-1]
	var untilEndOfTurn, retainAbility bool
	var addKeywords []KeywordKind
	if exceptIndex := entersAsCopyExceptIndex(rest, 0); exceptIndex >= 0 {
		retain, keywords, ok := parseBecomeCopyRider(rest[exceptIndex+1:], atoms)
		if !ok {
			return nil, false
		}
		retainAbility = retain
		addKeywords = keywords
		rest = rest[:exceptIndex]
	}
	for len(rest) > 0 && rest[len(rest)-1].Kind == shared.Comma {
		rest = rest[:len(rest)-1]
	}
	if trimmed, ok := becomeCopyTrimUntilEndOfTurn(rest); ok {
		rest = trimmed
		untilEndOfTurn = true
	}
	// A target selector must remain ("land", "artifact, creature, ...").
	if len(rest) == 0 {
		return nil, false
	}
	effect := EffectSyntax{
		Kind:                         EffectBecomeCopy,
		Context:                      EffectContextController,
		Span:                         sentence.Span,
		ClauseSpan:                   sentence.Span,
		Text:                         sentence.Text,
		Tokens:                       append([]shared.Token(nil), body...),
		BecomeCopyUntilEndOfTurn:     untilEndOfTurn,
		BecomeCopyRetainsThisAbility: retainAbility,
		BecomeCopyAddKeywords:        addKeywords,
	}
	return []EffectSyntax{effect}, true
}

// parseBecomeTypeEffect recognizes a targeted continuous type-adding effect
// ("Target permanent becomes an artifact in addition to its other types until
// end of turn.", Liquimetal Torque, Liquimetal Coating; CR 613.1d). It also
// recognizes the additive color-and-type form ("Until end of turn, target
// creature you control becomes a blue artifact in addition to its other colors
// and types.", Unctus, Grand Metatect), where one or more leading color words
// precede the card-type words and the additive tail names "colors and types".
// The "until end of turn" duration may appear as a leading clause or a trailing
// phrase, but exactly one of the two; both or neither fail closed. The target
// selector before "becomes" is left as an ordinary target for the target
// machinery to extract. Only the additive "in addition to its other [colors
// and] types" form is recognized; the type-setting form ("becomes a <type>"
// without "in addition") and the permanent (no-duration) form fail closed so
// those cards stay unsupported. Each card-type word must be a recognized
// permanent card type and each color word a recognized color; any other word
// fails closed.
func parseBecomeTypeEffect(sentence Sentence, tokens []shared.Token, atoms Atoms) ([]EffectSyntax, bool) {
	body := semanticEffectTokens(tokens)
	if len(body) == 0 || body[len(body)-1].Kind != shared.Period {
		return nil, false
	}
	remaining, leadingDuration := stripLeadingDurationClause(body[:len(body)-1], atoms)
	leadingUntilEndOfTurn := leadingDuration == EffectDurationUntilEndOfTurn
	words := normalizedWords(remaining)
	if len(words) < 5 || words[0] != "target" {
		return nil, false
	}
	becomesIndex := -1
	for i, word := range words {
		if word == "becomes" {
			becomesIndex = i
			break
		}
	}
	if becomesIndex < 0 || becomesIndex+1 >= len(words) {
		return nil, false
	}
	// The words between "target" and "becomes" must be only the target noun
	// phrase. A connector or pump/grant verb there marks a compound effect
	// ("target nonartifact creature gets +1/+0 and becomes an artifact ...",
	// Thran Forge) whose other clause this recognizer cannot represent, so it
	// fails closed rather than silently dropping that clause.
	for _, word := range words[1:becomesIndex] {
		switch word {
		case "and", "then", "gets", "get", "gains", "gain", "loses", "lose":
			return nil, false
		}
	}
	rest := words[becomesIndex+1:]
	if rest[0] != "a" && rest[0] != "an" {
		return nil, false
	}
	rest = rest[1:]
	duration := []string{"until", "end", "of", "turn"}
	trailingUntilEndOfTurn := false
	if len(rest) >= len(duration) && slices.Equal(rest[len(rest)-len(duration):], duration) {
		trailingUntilEndOfTurn = true
		rest = rest[:len(rest)-len(duration)]
	}
	if leadingUntilEndOfTurn == trailingUntilEndOfTurn {
		return nil, false
	}
	additiveTypes := []string{"in", "addition", "to", "its", "other", "types"}
	additiveColorsTypes := []string{"in", "addition", "to", "its", "other", "colors", "and", "types"}
	addsColors := false
	switch {
	case len(rest) >= len(additiveColorsTypes)+1 &&
		slices.Equal(rest[len(rest)-len(additiveColorsTypes):], additiveColorsTypes):
		addsColors = true
		rest = rest[:len(rest)-len(additiveColorsTypes)]
	case len(rest) >= len(additiveTypes)+1 &&
		slices.Equal(rest[len(rest)-len(additiveTypes):], additiveTypes):
		rest = rest[:len(rest)-len(additiveTypes)]
	default:
		return nil, false
	}
	addColors := make([]Color, 0)
	for len(rest) > 0 {
		parsedColor, ok := recognizeColorWord(rest[0])
		if !ok {
			break
		}
		addColors = append(addColors, parsedColor)
		rest = rest[1:]
	}
	typeWords := rest
	if len(typeWords) == 0 {
		return nil, false
	}
	if addsColors != (len(addColors) > 0) {
		return nil, false
	}
	addTypes := make([]types.Card, 0, len(typeWords))
	for _, word := range typeWords {
		cardType, ok := entersAsCopyAddTypeWord(word)
		if !ok {
			return nil, false
		}
		addTypes = append(addTypes, cardType)
	}
	effect := EffectSyntax{
		Kind:                     EffectBecomeType,
		Context:                  EffectContextController,
		Span:                     sentence.Span,
		ClauseSpan:               sentence.Span,
		Text:                     sentence.Text,
		Tokens:                   append([]shared.Token(nil), body...),
		BecomeTypeAddTypes:       addTypes,
		BecomeTypeAddColors:      addColors,
		BecomeTypeUntilEndOfTurn: true,
	}
	return []EffectSyntax{effect}, true
}

// "Until end of turn, target <creature> loses all abilities and becomes a
// [colorless] <color>* <subtype> [creature] with base power and toughness N/N."
// (Turn to Frog, Snakeform, Gift of Tusks; CR 613). The leading "until end of
// turn" duration is required; the target selector before "loses" is left in the
// sentence for the target machinery to extract. The body must lose all
// abilities and set both a creature subtype and a literal base power/toughness;
// any other shape (no duration, a no-type "has base power and toughness" body,
// trailing riders such as "and gains flying", a permanent duration, or a
// non-creature card type) fails closed so those cards stay unsupported.
func parsePolymorphEffect(sentence Sentence, tokens []shared.Token, atoms Atoms) ([]EffectSyntax, bool) {
	body := semanticEffectTokens(tokens)
	if len(body) == 0 || body[len(body)-1].Kind != shared.Period {
		return nil, false
	}
	inner := body[:len(body)-1]
	remaining, duration := stripLeadingDurationClause(inner, atoms)
	if duration != EffectDurationUntilEndOfTurn {
		return nil, false
	}
	if len(remaining) == 0 || !equalWord(remaining[0], "target") {
		return nil, false
	}
	loseIndex := -1
	for i := 1; i+2 < len(remaining); i++ {
		if equalWord(remaining[i], "loses") && equalWord(remaining[i+1], "all") &&
			equalWord(remaining[i+2], "abilities") {
			loseIndex = i
			break
		}
	}
	if loseIndex < 1 {
		return nil, false
	}
	cursor := loseIndex + 3
	if !staticWordsAt(remaining, cursor, "and", "becomes") {
		return nil, false
	}
	cursor += 2
	if staticWordsAt(remaining, cursor, "a") || staticWordsAt(remaining, cursor, "an") {
		cursor++
	}
	var colorless bool
	if staticWordsAt(remaining, cursor, "colorless") {
		colorless = true
		cursor++
	}
	list, next, ok := parseStaticCharacteristicList(remaining, cursor, len(remaining), atoms)
	if !ok || len(list.subtypes) == 0 {
		return nil, false
	}
	for _, cardType := range list.cardTypes {
		if cardType != CardTypeCreature {
			return nil, false
		}
	}
	if !staticWordsAt(remaining, next, "with") {
		return nil, false
	}
	basePT, ok := parseStaticBasePowerToughnessAt(remaining, next+1)
	if !ok || basePT.next != len(remaining) {
		return nil, false
	}
	effect := EffectSyntax{
		Kind:                   EffectPolymorph,
		Context:                EffectContextController,
		Span:                   sentence.Span,
		ClauseSpan:             sentence.Span,
		Text:                   sentence.Text,
		Tokens:                 append([]shared.Token(nil), body...),
		PolymorphColors:        list.colors,
		PolymorphColorless:     colorless,
		PolymorphSubtypes:      list.subtypes,
		PolymorphBasePower:     basePT.power,
		PolymorphBaseToughness: basePT.toughness,
	}
	return []EffectSyntax{effect}, true
}

// parseNamedBecomePolymorphEffect recognizes the permanent named-become
// polymorph "<target subject> becomes a N/N [legendary] [<color>*] <subtype>
// creature named <Name> and loses all abilities." (The Curse of Fenric II). The
// subject selector before "becomes" is left in the sentence for the target
// machinery to extract. The body must set a literal base power/toughness, at
// least one creature subtype, an explicit name, and lose all abilities; the
// change is permanent. Any other shape — a leading duration, no name, no
// power/toughness, an "in addition" rider, a non-creature card type, or a
// quoted granted ability — fails closed so unrelated "becomes" wordings keep
// their own recognizers.
func parseNamedBecomePolymorphEffect(sentence Sentence, tokens []shared.Token, atoms Atoms) ([]EffectSyntax, bool) {
	body := semanticEffectTokens(tokens)
	if len(body) == 0 || body[len(body)-1].Kind != shared.Period {
		return nil, false
	}
	inner := body[:len(body)-1]
	becomeIndex := -1
	for i := 1; i < len(inner); i++ {
		if equalWord(inner[i], "becomes") {
			becomeIndex = i
			break
		}
	}
	if becomeIndex < 1 {
		return nil, false
	}
	if len(inner) < becomeIndex+5 ||
		!equalWord(inner[len(inner)-4], "and") ||
		!equalWord(inner[len(inner)-3], "loses") ||
		!equalWord(inner[len(inner)-2], "all") ||
		!equalWord(inner[len(inner)-1], "abilities") {
		return nil, false
	}
	mid := inner[becomeIndex+1 : len(inner)-4]
	cursor := 0
	if staticWordsAt(mid, cursor, "a") || staticWordsAt(mid, cursor, "an") {
		cursor++
	}
	if cursor+3 > len(mid) ||
		mid[cursor].Kind != shared.Integer ||
		mid[cursor+1].Kind != shared.Slash ||
		mid[cursor+2].Kind != shared.Integer {
		return nil, false
	}
	power, err := strconv.Atoi(mid[cursor].Text)
	if err != nil {
		return nil, false
	}
	toughness, err := strconv.Atoi(mid[cursor+2].Text)
	if err != nil {
		return nil, false
	}
	cursor += 3
	var supertypes []Supertype
	for cursor < len(mid) {
		super, ok := atoms.SupertypeAt(mid[cursor].Span)
		if !ok {
			break
		}
		supertypes = append(supertypes, super)
		cursor++
	}
	list, next, ok := parseStaticCharacteristicList(mid, cursor, len(mid), atoms)
	if !ok || len(list.subtypes) == 0 {
		return nil, false
	}
	for _, cardType := range list.cardTypes {
		if cardType != CardTypeCreature {
			return nil, false
		}
	}
	cursor = next
	if cursor >= len(mid) || !equalWord(mid[cursor], "named") {
		return nil, false
	}
	nameTokens := mid[cursor+1:]
	if len(nameTokens) == 0 {
		return nil, false
	}
	for _, token := range nameTokens {
		if equalWord(token, "with") {
			return nil, false
		}
	}
	name := joinedEffectText(nameTokens)
	if name == "" {
		return nil, false
	}
	effect := EffectSyntax{
		Kind:                   EffectPolymorph,
		Context:                EffectContextController,
		Span:                   sentence.Span,
		ClauseSpan:             sentence.Span,
		Text:                   sentence.Text,
		Tokens:                 append([]shared.Token(nil), body...),
		PolymorphColors:        list.colors,
		PolymorphSubtypes:      list.subtypes,
		PolymorphBasePower:     power,
		PolymorphBaseToughness: toughness,
		PolymorphName:          name,
		PolymorphSupertypes:    supertypes,
		PolymorphPermanent:     true,
	}
	return []EffectSyntax{effect}, true
}

// becomeCopyTrimUntilEndOfTurn removes a trailing "until end of turn" phrase from
// rest, reporting whether it was present.
func becomeCopyTrimUntilEndOfTurn(rest []shared.Token) ([]shared.Token, bool) {
	if len(rest) < 4 {
		return rest, false
	}
	tail := rest[len(rest)-4:]
	if equalWord(tail[0], "until") && equalWord(tail[1], "end") &&
		equalWord(tail[2], "of") && equalWord(tail[3], "turn") {
		return rest[:len(rest)-4], true
	}
	return rest, false
}

// parseBecomeCopyRider parses the copiable riders of a become-a-copy "except
// <rider>" clause. It recognizes "it has this ability" (RetainsThisAbility) and
// "it has <keyword>" keyword riders; any other rider fails closed.
func parseBecomeCopyRider(clause []shared.Token, atoms Atoms) (bool, []KeywordKind, bool) {
	words := normalizedWords(clause)
	// Drop a leading "it has" / "it's" / "it is" subject.
	switch {
	case len(words) >= 2 && words[0] == "it" && words[1] == "has":
		clause = clause[2:]
		words = words[2:]
	case len(words) >= 2 && words[0] == "it" && words[1] == "is":
		clause = clause[2:]
		words = words[2:]
	default:
		return false, nil, false
	}
	if len(words) == 2 && words[0] == "this" && words[1] == "ability" {
		return true, nil, true
	}
	keywords := scanKeywords(clause, atoms)
	if len(keywords) == 0 {
		return false, nil, false
	}
	kinds := make([]KeywordKind, 0, len(keywords))
	for _, keyword := range keywords {
		kinds = append(kinds, keyword.Kind)
	}
	return false, kinds, true
}

// entersAsCopyAsEntersPrefix reports whether body begins with an "As this
// <permanent> enters," replacement prefix and, when it does, returns the index
// of the first token after the introducing comma. The prefix must reach its
// enter verb before any comma so unrelated "As ..." clauses fail closed.
func entersAsCopyAsEntersPrefix(body []shared.Token) (int, bool) {
	if len(body) < 4 || !equalWord(body[0], "as") || !equalWord(body[1], "this") {
		return 0, false
	}
	for i := 2; i < len(body); i++ {
		if body[i].Kind == shared.Comma {
			return 0, false
		}
		if equalWord(body[i], "enter") || equalWord(body[i], "enters") {
			if i+1 < len(body) && body[i+1].Kind == shared.Comma {
				return i + 2, true
			}
			return 0, false
		}
	}
	return 0, false
}

// trimTrailingUntilEndOfTurn removes a trailing "until end of turn" phrase from a
// copied-permanent filter, reporting whether it was present. The phrase marks the
// temporary "become a copy ... until end of turn" copy duration (Cursed Mirror).
func trimTrailingUntilEndOfTurn(filter []shared.Token) ([]shared.Token, bool) {
	if len(filter) < 4 {
		return filter, false
	}
	tail := filter[len(filter)-4:]
	if equalWord(tail[0], "until") && equalWord(tail[1], "end") &&
		equalWord(tail[2], "of") && equalWord(tail[3], "turn") {
		return filter[:len(filter)-4], true
	}
	return filter, false
}

// entersAsCopyExceptIndex finds the index of the "except" word that introduces a
// copiable rider in an enters-as-copy clause, searching from start. It returns
// -1 when no rider is present.
func entersAsCopyExceptIndex(body []shared.Token, start int) int {
	for i := start; i < len(body); i++ {
		if equalWord(body[i], "except") {
			return i
		}
	}
	return -1
}

// entersAsCopyRiders collects the recognized copiable riders parsed from an
// enters-as-copy "except <rider>" clause.
type entersAsCopyRiders struct {
	addTypes            []types.Card
	addSubtypes         []types.Sub
	addKeywords         []KeywordKind
	notLegendary        bool
	conditionalCounters []EntersAsCopyConditionalCounter
}

// parseEntersAsCopyRider parses the recognized copiable riders of an
// enters-as-copy clause: "it isn't legendary" / "it's not legendary" sets the
// not-legendary flag, "it's an <type> in addition to its other types" adds the
// named card type, and "it enters with an additional <kind> counter on it if
// it's a <type>" adds a conditional copiable counter (Spark Double). Riders
// joined by commas or "and" are supported. Each clause must match a recognized
// template exactly; any other wording fails closed.
func parseEntersAsCopyRider(rider []shared.Token, atoms Atoms) (entersAsCopyRiders, bool) {
	clauses := splitEntersAsCopyRiderClauses(rider)
	if len(clauses) == 0 {
		return entersAsCopyRiders{}, false
	}
	var riders entersAsCopyRiders
	for _, clause := range clauses {
		words := normalizedWords(clause)
		if entersAsCopyNotLegendaryClause(words) {
			riders.notLegendary = true
			continue
		}
		if placement, ok := entersAsCopyConditionalCounterClause(clause); ok {
			riders.conditionalCounters = append(riders.conditionalCounters, placement)
			continue
		}
		if keyword, ok := entersAsCopyAddKeywordClause(clause, atoms); ok {
			riders.addKeywords = append(riders.addKeywords, keyword)
			continue
		}
		cardTypes, subtypes, typeOK := entersAsCopyAddTypeClause(words)
		if !typeOK {
			return entersAsCopyRiders{}, false
		}
		riders.addTypes = append(riders.addTypes, cardTypes...)
		riders.addSubtypes = append(riders.addSubtypes, subtypes...)
	}
	return riders, true
}

// entersAsCopyAddKeywordClause matches the "it has <keyword>" copiable rider on
// an enters-as-copy replacement ("except it has haste", Cursed Mirror) and
// returns the single granted keyword. It fails closed on any wording other than
// exactly "it has <keyword>" with one recognized keyword filling the clause.
func entersAsCopyAddKeywordClause(clause []shared.Token, atoms Atoms) (KeywordKind, bool) {
	if len(clause) < 3 || !equalWord(clause[0], "it") || !equalWord(clause[1], "has") {
		return KeywordUnknown, false
	}
	keywords := scanKeywords(clause[2:], atoms)
	if len(keywords) != 1 {
		return KeywordUnknown, false
	}
	return keywords[0].Kind, true
}

// splitEntersAsCopyRiderClauses splits a copiable-rider token run into individual
// clauses on top-level commas and "and" conjunctions.
func splitEntersAsCopyRiderClauses(rider []shared.Token) [][]shared.Token {
	var clauses [][]shared.Token
	var current []shared.Token
	flush := func() {
		if len(current) > 0 {
			clauses = append(clauses, current)
			current = nil
		}
	}
	for _, tok := range rider {
		if tok.Kind == shared.Comma || equalWord(tok, "and") {
			flush()
			continue
		}
		current = append(current, tok)
	}
	flush()
	return clauses
}

// entersAsCopyNotLegendaryClause reports whether a rider clause is the
// "it isn't legendary" / "it's not legendary" copiable rider.
func entersAsCopyNotLegendaryClause(words []string) bool {
	if len(words) < 2 || len(words) > 4 {
		return false
	}
	if words[len(words)-1] != "legendary" {
		return false
	}
	negation := false
	for _, word := range words[:len(words)-1] {
		switch word {
		case "it", "it's", "is":
		case "isn't", "not", "n't":
			negation = true
		default:
			return false
		}
	}
	return negation
}

// entersAsCopyAddTypeClause matches the "it's a <type...> in addition to its
// other types" copiable rider and returns the added card types and subtypes. The
// type run between the article and "in addition" may mix card types and
// subtypes in any order ("an artifact" Phyrexian Metamorph, "a Bird" Mockingbird,
// "a Synth artifact creature" Synth Infiltrator, "a Faerie Shapeshifter"
// Malleable Impostor). Each word must classify as either a recognized card type
// or a recognized subtype; any unrecognized word fails closed.
func entersAsCopyAddTypeClause(words []string) (cardTypes []types.Card, subtypes []types.Sub, ok bool) {
	switch {
	case len(words) >= 2 && words[0] == "it's":
		words = words[1:]
	case len(words) >= 3 && words[0] == "it" && (words[1] == "is" || words[1] == "'s"):
		words = words[2:]
	default:
		return nil, nil, false
	}
	if len(words) < 2 || (words[0] != "a" && words[0] != "an") {
		return nil, nil, false
	}
	words = words[1:]
	suffix := []string{"in", "addition", "to", "its", "other", "types"}
	// Newer printings scope the added types to the copied permanent's creature
	// types ("a Ninja in addition to its other creature types", Sakashima's
	// Student); the added types are subtypes either way, so accept the
	// "creature types" suffix as the same copiable add-type rider.
	creatureSuffix := []string{"in", "addition", "to", "its", "other", "creature", "types"}
	var typeWords []string
	switch {
	case len(words) > len(creatureSuffix) &&
		slices.Equal(words[len(words)-len(creatureSuffix):], creatureSuffix):
		typeWords = words[:len(words)-len(creatureSuffix)]
	case len(words) > len(suffix) &&
		slices.Equal(words[len(words)-len(suffix):], suffix):
		typeWords = words[:len(words)-len(suffix)]
	default:
		return nil, nil, false
	}
	for _, word := range typeWords {
		if cardType, typeOK := entersAsCopyAddTypeWord(word); typeOK {
			cardTypes = append(cardTypes, cardType)
			continue
		}
		if sub, subOK := recognizeSubtypePhrase(word); subOK {
			subtypes = append(subtypes, sub)
			continue
		}
		return nil, nil, false
	}
	if len(cardTypes) == 0 && len(subtypes) == 0 {
		return nil, nil, false
	}
	return cardTypes, subtypes, true
}

// entersAsCopyAddTypeWord maps a singular card-type word used in an
// enters-as-copy "in addition to its other types" rider to its card type. It
// fails closed on any other word.
func entersAsCopyAddTypeWord(word string) (types.Card, bool) {
	switch strings.ToLower(word) {
	case "artifact":
		return types.Artifact, true
	case "creature":
		return types.Creature, true
	case "enchantment":
		return types.Enchantment, true
	case "land":
		return types.Land, true
	default:
		return "", false
	}
}

// entersAsCopyConditionalCounterClause matches the conditional copiable counter
// rider "it enters with an additional <kind> counter on it if it's a <type>"
// (Spark Double) and returns a single counter placement guarded by the named
// card type. The counter kind is read at the token level (counterNameBefore)
// because symbol kinds such as "+1/+1" are dropped from normalized words. It
// fails closed on any other wording.
func entersAsCopyConditionalCounterClause(clause []shared.Token) (EntersAsCopyConditionalCounter, bool) {
	counterIndex := entersAsCopyWordIndex(clause, "counter", 0)
	if counterIndex < 0 {
		return EntersAsCopyConditionalCounter{}, false
	}
	kind, _, ok := counterNameBefore(clause, counterIndex)
	if !ok {
		return EntersAsCopyConditionalCounter{}, false
	}
	prefix := normalizedWords(clause[:counterIndex])
	if !slices.Contains(prefix, "enters") && !slices.Contains(prefix, "enter") {
		return EntersAsCopyConditionalCounter{}, false
	}
	if !slices.Contains(prefix, "with") {
		return EntersAsCopyConditionalCounter{}, false
	}
	ifIndex := entersAsCopyWordIndex(clause, "if", counterIndex)
	if ifIndex < 0 {
		return EntersAsCopyConditionalCounter{}, false
	}
	cardType, ok := entersAsCopyConditionalTypeTail(normalizedWords(clause[ifIndex:]))
	if !ok {
		return EntersAsCopyConditionalCounter{}, false
	}
	return EntersAsCopyConditionalCounter{Kind: kind, Amount: 1, IfType: cardType}, true
}

// entersAsCopyConditionalTypeTail parses the "if it's a <type>" / "if it is a
// <type>" tail of a conditional copiable counter rider into its card type. It
// fails closed on any other wording.
func entersAsCopyConditionalTypeTail(words []string) (types.Card, bool) {
	switch {
	case len(words) == 4 && words[0] == "if" && words[1] == "it's" && (words[2] == "a" || words[2] == "an"):
		return entersAsCopyConditionalTypeWord(words[3])
	case len(words) == 5 && words[0] == "if" && words[1] == "it" && words[2] == "is" && (words[3] == "a" || words[3] == "an"):
		return entersAsCopyConditionalTypeWord(words[4])
	default:
		return "", false
	}
}

// entersAsCopyConditionalTypeWord maps a singular card-type word used in a
// conditional copiable counter rider's "if it's a <type>" tail to its card
// type, including planeswalker (loyalty rider). It fails closed on any other
// word.
func entersAsCopyConditionalTypeWord(word string) (types.Card, bool) {
	switch strings.ToLower(word) {
	case "artifact":
		return types.Artifact, true
	case "creature":
		return types.Creature, true
	case "enchantment":
		return types.Enchantment, true
	case "land":
		return types.Land, true
	case "planeswalker":
		return types.Planeswalker, true
	default:
		return "", false
	}
}

// entersAsCopyWordIndex returns the index of the first token at or after start
// whose word equals word, or -1 when none matches.
func entersAsCopyWordIndex(tokens []shared.Token, word string, start int) int {
	for i := start; i < len(tokens); i++ {
		if equalWord(tokens[i], word) {
			return i
		}
	}
	return -1
}

// entersAsCopyFilterOnBattlefield reports whether a copy-filter token run is
// scoped to the battlefield, either via a trailing "on the battlefield" phrase
// or a controller relation ("you control", "an opponent controls", "your
// opponents control"). Filters without such a scope (for example "any creature
// card in a graveyard") fail closed because this replacement can only copy
// battlefield permanents.
func entersAsCopyFilterOnBattlefield(filter []shared.Token) bool {
	if len(filter) >= 3 &&
		equalWord(filter[len(filter)-3], "on") &&
		equalWord(filter[len(filter)-2], "the") &&
		equalWord(filter[len(filter)-1], "battlefield") {
		return true
	}
	for i := 0; i+1 < len(filter); i++ {
		if equalWord(filter[i], "you") && equalWord(filter[i+1], "control") {
			return true
		}
		if equalWord(filter[i], "opponent") && equalWord(filter[i+1], "controls") {
			return true
		}
		if equalWord(filter[i], "opponents") && equalWord(filter[i+1], "control") {
			return true
		}
	}
	return false
}

// trimTrailingZonePhrase drops a trailing "on the battlefield" zone phrase from
// a copy-filter token run so the filter selects the permanent type rather than a
// battlefield zone, which would make the selector unrepresentable.
func trimTrailingZonePhrase(filter []shared.Token) []shared.Token {
	if len(filter) >= 3 &&
		equalWord(filter[len(filter)-3], "on") &&
		equalWord(filter[len(filter)-2], "the") &&
		equalWord(filter[len(filter)-1], "battlefield") {
		return filter[:len(filter)-3]
	}
	return filter
}

// parseGroupEntersTappedEffect recognizes a static enters-tapped replacement that
// taps a group of OTHER permanents as they enter, such as "Creatures your
// opponents control enter tapped." (Authority of the Consuls), "Artifacts,
// creatures, and lands your opponents control enter the battlefield tapped."
// (Frozen Aether), or the unscoped "Permanents enter tapped." (Kismet family).
// The subject is a list of permanent-type plurals (or the catch-all
// "Permanents"), an optional controller scope, and the plural "enter [the
// battlefield] tapped" verb phrase. It matches the whole sentence exactly, so
// any other wording falls through to the generic effect grammar.
func parseGroupEntersTappedEffect(sentence Sentence, tokens []shared.Token) ([]EffectSyntax, bool) {
	body := semanticEffectTokens(tokens)
	if len(body) < 4 || body[len(body)-1].Kind != shared.Period {
		return nil, false
	}
	words := body[:len(body)-1]
	index := 0
	var cardTypes []types.Card
	if equalWord(words[0], "permanents") {
		index = 1
	} else {
		for index < len(words) {
			cardType, ok := groupEntersTappedPermanentType(words[index].Text)
			if !ok {
				break
			}
			cardTypes = append(cardTypes, cardType)
			index++
			for index < len(words) && (words[index].Kind == shared.Comma || equalWord(words[index], "and")) {
				index++
			}
		}
		if len(cardTypes) == 0 {
			return nil, false
		}
	}
	scope := EntersTappedGroupControllerEach
	switch {
	case index+2 < len(words) && equalWord(words[index], "your") &&
		equalWord(words[index+1], "opponents") && equalWord(words[index+2], "control"):
		scope = EntersTappedGroupControllerOpponents
		index += 3
	case index+2 < len(words) && equalWord(words[index], "an") &&
		equalWord(words[index+1], "opponent") && equalWord(words[index+2], "controls"):
		scope = EntersTappedGroupControllerOpponents
		index += 3
	case index+1 < len(words) && equalWord(words[index], "you") && equalWord(words[index+1], "control"):
		scope = EntersTappedGroupControllerYou
		index += 2
	default:
	}
	if index >= len(words) || !equalWord(words[index], "enter") {
		return nil, false
	}
	index++
	if index+1 < len(words) && equalWord(words[index], "the") && equalWord(words[index+1], "battlefield") {
		index += 2
	}
	if index >= len(words) || !equalWord(words[index], "tapped") {
		return nil, false
	}
	index++
	if index != len(words) {
		return nil, false
	}
	effect := EffectSyntax{
		Kind:         EffectEnterTapped,
		Context:      EffectContextController,
		Span:         sentence.Span,
		ClauseSpan:   sentence.Span,
		Text:         sentence.Text,
		Tokens:       append([]shared.Token(nil), tokens...),
		EntersTapped: true,
		GroupEntryModification: GroupEntryModificationSyntax{
			Kind:            GroupEntryModificationTapped,
			ControllerScope: scope,
			Types:           cardTypes,
		},
	}
	effect.Exact = exactEffectSyntax(&effect)
	return []EffectSyntax{effect}, true
}

// parseGroupEntersWithCountersEffect recognizes a static enters-with-counters
// replacement that adds counters to a group of the controller's permanents as
// they enter, e.g. "Each other creature you control enters with an additional
// vigilance counter on it." (Tayam, Luminous Enigma) or "Each planeswalker you
// control enters with an additional loyalty counter on it." (Oath of Gideon).
// The subject is an "Each <group> you control" noun phrase recognized by
// parseSelection; the predicate is parsed by
// parseGroupEntersWithCountersPredicate, which accepts both the fixed single-
// counter form and the dynamic "a number of additional <kind> counters on it
// equal to <amount>" form (Arwen, Weaver of Hope). Multi-counter and "X"
// quantities fail closed so the card stays unsupported.
func parseGroupEntersWithCountersEffect(sentence Sentence, tokens []shared.Token, atoms Atoms) ([]EffectSyntax, bool) {
	body := semanticEffectTokens(tokens)
	if len(body) < 8 || body[len(body)-1].Kind != shared.Period {
		return nil, false
	}
	words := body[:len(body)-1]
	if !equalWord(words[0], "each") {
		return nil, false
	}
	entersIndex := -1
	for i := 1; i < len(words); i++ {
		if equalWord(words[i], "enters") {
			entersIndex = i
			break
		}
	}
	if entersIndex < 0 {
		return nil, false
	}
	recipient := words[1:entersIndex]
	if len(recipient) == 0 {
		return nil, false
	}
	chosenEntryType := false
	if base, ok := cutChosenTypeQualifierSuffix(recipient); ok {
		recipient, chosenEntryType = base, true
	}
	if len(recipient) == 0 {
		return nil, false
	}
	if recipientHasRelativeClause(recipient) {
		return nil, false
	}
	predicate := words[entersIndex+1:]
	placement, ok := parseGroupEntersWithCountersPredicate(predicate, atoms)
	if !ok || !placement.CounterKnown {
		return nil, false
	}
	selection := parseSelection(recipient, atoms)
	if selection.Controller != SelectionControllerYou {
		return nil, false
	}
	selection.SubtypeFromEntryChoice = chosenEntryType
	effect := EffectSyntax{
		Kind:               EffectEnterTapped,
		Context:            EffectContextController,
		Span:               sentence.Span,
		ClauseSpan:         sentence.Span,
		Text:               sentence.Text,
		Tokens:             append([]shared.Token(nil), tokens...),
		EntersWithCounters: true,
		GroupEntryModification: GroupEntryModificationSyntax{
			Kind: GroupEntryModificationWithCounters,
		},
		Selection:    selection,
		CounterKind:  placement.CounterKind,
		CounterKnown: placement.CounterKnown,
		Amount:       placement.Amount,
	}
	effect.Exact = exactEffectSyntax(&effect)
	return []EffectSyntax{effect}, true
}

// groupEntersWithCountersPlacement is the parsed counter placement of a group
// enters-with-counters predicate: the counter kind, whether it was recognized,
// and the optional dynamic amount (empty for a fixed single counter).
type groupEntersWithCountersPlacement struct {
	CounterKind  counter.Kind
	CounterKnown bool
	Amount       EffectAmountSyntax
}

// parseGroupEntersWithCountersPredicate parses the predicate of a group
// enters-with-counters replacement ("... enters <predicate>.") into a counter
// kind and placement amount. It accepts the fixed single-counter form ("with an
// additional <kind> counter on it" — Tayam, Luminous Enigma) and the dynamic
// form ("with a number of additional <kind> counters on it equal to <amount>" —
// Arwen, Weaver of Hope). The dynamic amount is parsed by the shared
// dynamic-amount recognizer so every supported amount form unlocks. An empty
// returned amount means a fixed single counter. Any other shape, an "X" count,
// or an unrecognized counter kind fails closed.
func parseGroupEntersWithCountersPredicate(predicate []shared.Token, atoms Atoms) (groupEntersWithCountersPlacement, bool) {
	if len(predicate) < 5 || !equalWord(predicate[0], "with") {
		return groupEntersWithCountersPlacement{}, false
	}
	if equalWord(predicate[1], "a") && effectWordsAt(predicate, 2, "number", "of", "additional") {
		return parseGroupDynamicEntersWithCounters(predicate[5:], atoms)
	}
	if !equalWord(predicate[1], "an") ||
		!equalWord(predicate[2], "additional") ||
		!equalWord(predicate[len(predicate)-2], "on") ||
		!equalWord(predicate[len(predicate)-1], "it") {
		return groupEntersWithCountersPlacement{}, false
	}
	counterClause := predicate[3 : len(predicate)-2]
	if len(counterClause) < 2 {
		return groupEntersWithCountersPlacement{}, false
	}
	last := counterClause[len(counterClause)-1]
	if !equalWord(last, "counter") && !equalWord(last, "counters") {
		return groupEntersWithCountersPlacement{}, false
	}
	if equalWord(counterClause[0], "x") {
		return groupEntersWithCountersPlacement{}, false
	}
	counterKind, counterKnown := parseCounterPlacement(counterClause, atoms)
	return groupEntersWithCountersPlacement{CounterKind: counterKind, CounterKnown: counterKnown}, true
}

// parseGroupDynamicEntersWithCounters parses the tail "<kind> counters on it
// <amount>" of a dynamic group enters-with-counters predicate, after the leading
// "with a number of additional" prefix has been consumed. The counter kind
// precedes "counters on it" and the trailing dynamic amount is parsed by the
// shared dynamic-amount recognizer. It fails closed on an "X" count, an
// unrecognized counter kind, or a missing or unrecognized amount.
func parseGroupDynamicEntersWithCounters(tokens []shared.Token, atoms Atoms) (groupEntersWithCountersPlacement, bool) {
	onIndex := -1
	for i := 1; i+2 < len(tokens); i++ {
		if equalWord(tokens[i], "counters") &&
			equalWord(tokens[i+1], "on") && equalWord(tokens[i+2], "it") {
			onIndex = i
			break
		}
	}
	if onIndex < 1 {
		return groupEntersWithCountersPlacement{}, false
	}
	counterClause := tokens[:onIndex+1]
	if equalWord(counterClause[0], "x") {
		return groupEntersWithCountersPlacement{}, false
	}
	counterKind, counterKnown := parseCounterPlacement(counterClause, atoms)
	if !counterKnown {
		return groupEntersWithCountersPlacement{}, false
	}
	amount, _, ok := parseDynamicEffectAmount(tokens[onIndex+3:], atoms)
	if !ok || amount.DynamicKind == EffectDynamicAmountNone {
		return groupEntersWithCountersPlacement{}, false
	}
	return groupEntersWithCountersPlacement{CounterKind: counterKind, CounterKnown: counterKnown, Amount: amount}, true
}

// recipientHasRelativeClause reports whether a group enters-with-counters
// recipient phrase carries a relative or possessive clause ("of the chosen
// type", "that's a Wolf or a Werewolf", "that has an Adventure") that
// parseSelection cannot fully and reliably represent. Such phrases either drop
// silently (over-broadening the group) or map to characteristics whose runtime
// meaning is uncertain, so the recognizer fails closed on them and only accepts
// simple "[other] [non-<subtype>] [<subtype>] <type> you control" recipients.
func recipientHasRelativeClause(recipient []shared.Token) bool {
	for _, tok := range recipient {
		switch strings.ToLower(tok.Text) {
		case "that", "that's", "of", "with", "named", "chosen", "has", "have", "whose", "without":
			return true
		}
	}
	return false
}

// cutChosenTypeQualifierSuffix strips a trailing "of the chosen type" qualifier
// from a group recipient phrase ("each other creature you control of the chosen
// type") and returns the bare recipient with ok=true. The matched permanents
// must share the creature subtype the source permanent chose as it entered
// (Metallic Mimic); the caller records that as Selection.SubtypeFromEntryChoice.
func cutChosenTypeQualifierSuffix(recipient []shared.Token) ([]shared.Token, bool) {
	n := len(recipient)
	if n >= 4 && effectWordsAt(recipient, n-4, "of", "the", "chosen", "type") {
		return recipient[:n-4], true
	}
	return recipient, false
}

// entersTappedSelfSyntax recognizes a plain self enters-tapped clause. "This
// land enters tapped." or "Nyx Lotus enters tapped." The enters verb is
// shared by many entry constructs ("As ~ enters, choose ...", "enters with
// counters", "enters tapped and attacking"), so the qualifier following the
// verb must be exactly "tapped" (optionally "the battlefield tapped") to avoid
// classifying unrelated entry effects as a plain tapped entry.
func entersTappedSelfSyntax(kind EffectKind, clause []shared.Token) bool {
	if kind != EffectEnterTapped {
		return false
	}
	body := clause
	if len(body) >= 2 && equalWord(body[0], "the") && equalWord(body[1], "battlefield") {
		body = body[2:]
	}
	return len(body) == 2 && equalWord(body[0], "tapped") && body[1].Text == "."
}

// counterRecipientAttached reports that a counter-placement effect ("put ...
// counter(s) on enchanted creature" / "on equipped creature") targets the
// permanent the source Aura or Equipment is attached to. It gates on the counter
// verb and a known counter kind and matches only the bare attached-creature
// recipient; exact canonical reconstruction independently confirms the full
// clause wording, so any additional qualifier leaves the effect inexact and
// fails closed in lowering.
func counterRecipientAttached(kind EffectKind, counterKnown bool, clause []shared.Token) bool {
	if kind != EffectPut || !counterKnown {
		return false
	}
	return effectHasTokenWords(clause, "on", "enchanted", "creature") ||
		effectHasTokenWords(clause, "on", "equipped", "creature")
}

// fightSubjectAttached reports that a fight effect's fighter is the permanent the
// source Aura or Equipment is attached to ("enchanted creature fights up to one
// target creature an opponent controls", "equipped creature fights target
// creature"). The bare attached-creature subject precedes the fight verb, so it
// is matched against the subject tokens that lead the clause; the "<enchanted|
// equipped> creature" run appears only in the subject position, so any later
// object phrase cannot false-match. Every other fighting subject leaves it false
// so lowering keeps the existing source, event-permanent, and two-target fight
// shapes.
func fightSubjectAttached(kind EffectKind, subject []shared.Token) bool {
	if kind != EffectFight {
		return false
	}
	return effectHasTokenWords(subject, "enchanted", "creature") ||
		effectHasTokenWords(subject, "equipped", "creature")
}

// moveAllCountersClause reports the kind-agnostic "move all counters" form,
// where every counter on the source moves regardless of kind ("Move all
// counters from this permanent onto target creature."). It anchors on the
// literal "all counters" run so the specific-kind form ("a +1/+1 counter")
// keeps MoveCountersAll false and lowers through its named-kind path.
func moveAllCountersClause(clause []shared.Token) bool {
	return effectHasTokenWords(clause, "all", "counters")
}

// moveCountersDistributeClause reports the "move any number of <kind> counters
// ... onto other creatures" form, where the controller distributes the source's
// counters among a group of other creatures rather than a single target. It
// anchors on the literal "any number of" run so the single-target move forms
// keep MoveCountersDistribute false.
func moveCountersDistributeClause(clause []shared.Token) bool {
	return effectHasTokenWords(clause, "any", "number", "of")
}

// removeAllCountersClause reports the kind-agnostic "remove all counters" form,
// where every counter on the object is removed regardless of kind ("Remove all
// counters from target permanent."). It anchors on the literal "all counters"
// run so the specific-kind form ("a +1/+1 counter") and the kind-specific "all
// +1/+1 counters" form (whose words are non-adjacent) keep RemoveCountersAll
// false and lower through their fixed-count paths.
func removeAllCountersClause(clause []shared.Token) bool {
	return effectHasTokenWords(clause, "all", "counters")
}

// moveThoseCountersClause reports the counter-salvage form "put those counters
// on <destination>" or its singular-pronoun variant "put its counters on
// <destination>", where "those"/"its" name the counters a triggering permanent
// had as it left a zone ("When this creature dies, put its counters on target
// creature you control."). It anchors on the literal "those counters" or "its
// counters" run so an ordinary counter placement ("put a +1/+1 counter on ...")
// keeps MoveThoseCounters false. The kind-named singular form ("put its +1/+1
// counters on ...") keeps the pronoun and the noun non-adjacent and so stays
// out of this kind-agnostic salvage move.
func moveThoseCountersClause(clause []shared.Token) bool {
	return effectHasTokenWords(clause, "those", "counters") ||
		effectHasTokenWords(clause, "its", "counters")
}
