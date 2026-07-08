package parser

import (
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// emitDelayedTriggerEffects rewrites a single-sentence ability whose leading
// clause is a cast-event "this turn" delayed-trigger preamble ("Whenever you
// cast a spell this turn, ...", "When you next cast a creature spell this turn,
// ...") into one EffectDelayedTrigger effect carrying the sentence reparsed as a
// nested triggered ability with its turn window stripped. The preamble's cast
// verb otherwise reads as a spurious resolving cast effect that blocks lowering,
// and the post-comma body would read as an immediate effect rather than a
// delayed trigger. Rewriting fails closed: an ability the recognizer does not
// match, or whose stripped body does not reparse to exactly one triggered
// ability, is left untouched.
func emitDelayedTriggerEffects(abilities []Ability, instantOrSorcery bool) {
	for i := range abilities {
		rewriteDelayedTriggerAbility(&abilities[i])
		rewriteCapturedCombatDamageDelayedTrigger(&abilities[i])
		rewriteCapturedAttacksMonarchDelayedTrigger(&abilities[i])
		rewriteCapturedDiesMonarchDelayedTrigger(&abilities[i])
		rewriteSpellTriggeredThisTurnDelayedAbility(&abilities[i], instantOrSorcery)
	}
}

// rewriteSpellTriggeredThisTurnDelayedAbility converts a spell paragraph that
// parsed as a standing triggered ability whose trigger ends with "this turn"
// ("Whenever one or more creatures you control deal combat damage to one or more
// players this turn, you become the monarch.", Forth Eorlingas!'s second
// paragraph) into a single EffectDelayedTrigger spell effect. On an instant or
// sorcery such a paragraph is not a standing triggered ability but a delayed
// trigger the spell sets up as it resolves; the parser strips the "this turn"
// preamble into ability.Trigger, so this rebuilds the nested triggered ability
// from ability.Text (which retains the full preamble) and reclassifies the
// ability as a spell effect that the backend merges into the spell. It fails
// closed on any non-spell ability, a trigger that does not end with "this turn",
// or a body that does not reparse to exactly one plain triggered ability.
func rewriteSpellTriggeredThisTurnDelayedAbility(ability *Ability, instantOrSorcery bool) {
	if !instantOrSorcery ||
		ability.Kind != AbilityTriggered ||
		ability.Trigger == nil ||
		len(ability.Sentences) != 1 {
		return
	}
	triggerLower := strings.ToLower(strings.TrimSpace(ability.Trigger.Text))
	if !strings.HasSuffix(triggerLower, "this turn") {
		return
	}
	if !strings.Contains(triggerLower, "cast") && !strings.Contains(triggerLower, "combat damage") {
		return
	}
	inner, oneShot, ok := delayedTriggerInnerText(ability.Text)
	if !ok {
		return
	}
	granted, ok := parseDelayedTriggerAbility(inner)
	if !ok {
		return
	}
	sentence := &ability.Sentences[0]
	effect := EffectSyntax{
		Kind:                  EffectDelayedTrigger,
		Span:                  ability.Span,
		VerbSpan:              ability.Span,
		ClauseSpan:            ability.Span,
		Text:                  ability.Text,
		DelayedTriggerAbility: &granted,
		DelayedTriggerOneShot: oneShot,
	}
	ability.Kind = AbilitySpell
	ability.Trigger = nil
	sentence.Effects = []EffectSyntax{effect}
	sentence.Targets = nil
	sentence.LegacyEffects = false
	ability.SemanticReferences = nil
	ability.SemanticKeywords = nil
	ability.ConditionBoundaries = nil
	ability.EventHistoryConditions = nil
	ability.ConditionClauses = nil
	ability.ConditionSegments = nil
}

// rewriteCapturedCombatDamageDelayedTrigger rewrites a trailing "Whenever that
// creature deals combat damage to a player this turn, <body>" rider sentence
// into one EffectDelayedTrigger whose source binds to the permanent an earlier
// clause in the same ability acted on ("... target creature ... Whenever that
// creature deals combat damage to a player this turn, you draw a card."). The
// back-reference subject ("that creature") otherwise reads as a spurious
// resolving combat-damage effect that blocks lowering. The rider is reparsed in
// the self ("this creature") form so it carries an ordinary combat-damage
// trigger pattern; lowering rebinds that pattern's source to the captured
// object via the ability's preserved back-reference. The ability's earlier
// sentences and its semantic references are left untouched so the reference
// still resolves to the antecedent target. It fails closed: any ability whose
// trailing sentence the recognizer does not match, or whose reparsed self-form
// is not exactly one triggered ability, is left unchanged.
func rewriteCapturedCombatDamageDelayedTrigger(ability *Ability) {
	if len(ability.Sentences) < 2 {
		return
	}
	last := &ability.Sentences[len(ability.Sentences)-1]
	tokens := semanticEffectTokens(last.Tokens)
	comma := shared.TopLevelIndex(tokens, shared.Comma)
	if comma <= 0 {
		return
	}
	lead := tokens[:comma]
	if !isDelayedThisTurnPreamble(lead) || !leadBindsThatCombatDamageToPlayer(lead) {
		return
	}
	inner, ok := capturedCombatDamageInnerText(last.Text)
	if !ok {
		return
	}
	granted, ok := parseDelayedTriggerAbility(inner)
	if !ok {
		return
	}
	var references, subjectReferences []Reference
	for i := range last.Effects {
		references = append(references, last.Effects[i].References...)
		subjectReferences = append(subjectReferences, last.Effects[i].SubjectReferences...)
	}
	last.Effects = []EffectSyntax{{
		Kind:                           EffectDelayedTrigger,
		Span:                           last.Span,
		VerbSpan:                       last.Span,
		ClauseSpan:                     last.Span,
		Text:                           last.Text,
		References:                     references,
		SubjectReferences:              subjectReferences,
		DelayedTriggerAbility:          &granted,
		DelayedTriggerBindDamageSource: true,
	}}
	last.Targets = nil
	last.LegacyEffects = false
}

// leadBindsThatCombatDamageToPlayer reports whether a delayed-trigger preamble
// names a back-referenced permanent ("that creature") that deals combat damage
// to a player, the captured-object combat-damage shape the rider rewriter binds.
func leadBindsThatCombatDamageToPlayer(lead []shared.Token) bool {
	if len(lead) < 2 || !equalWord(lead[1], "that") {
		return false
	}
	hasCombatDamage := false
	hasPlayer := false
	for i := range lead {
		if i+2 < len(lead) &&
			equalWord(lead[i], "deals") &&
			equalWord(lead[i+1], "combat") &&
			equalWord(lead[i+2], "damage") {
			hasCombatDamage = true
		}
		if equalWord(lead[i], "player") {
			hasPlayer = true
		}
	}
	return hasCombatDamage && hasPlayer
}

// capturedCombatDamageInnerText reconstructs the self-form triggered-ability
// source of a captured-object combat-damage rider by stripping the "this turn"
// window and rewriting the "that <noun>" back-reference subject to "this
// <noun>" so the result is an ordinary source-self combat-damage trigger
// ("Whenever that creature deals combat damage to a player this turn, you draw
// a card." -> "Whenever this creature deals combat damage to a player, you draw
// a card."). Lowering rebinds the resulting pattern's source to the captured
// object, so the self form supplies only the combat-damage event shape. It
// fails closed on any other preamble.
func capturedCombatDamageInnerText(text string) (inner string, ok bool) {
	trimmed := strings.TrimSpace(text)
	comma := strings.Index(trimmed, ",")
	if comma <= 0 {
		return "", false
	}
	preamble := strings.TrimSpace(trimmed[:comma])
	body := trimmed[comma:]
	lowered := strings.ToLower(preamble)
	if !strings.HasSuffix(lowered, "this turn") {
		return "", false
	}
	preamble = strings.TrimSpace(preamble[:len(preamble)-len("this turn")])
	lowered = strings.ToLower(preamble)
	var rest string
	switch {
	case strings.HasPrefix(lowered, "whenever that "):
		rest = preamble[len("whenever that "):]
	case strings.HasPrefix(lowered, "when that "):
		rest = preamble[len("when that "):]
	default:
		return "", false
	}
	if !strings.Contains(strings.ToLower(rest), "deals combat damage") {
		return "", false
	}
	return "Whenever this " + rest + body, true
}

// rewriteCapturedAttacksMonarchDelayedTrigger rewrites a trailing "Whenever that
// creature attacks the monarch this turn, <body>" rider sentence into one
// EffectDelayedTrigger whose attacker-declared event binds to the permanent an
// earlier clause in the same ability acted on ("... target creature ... Whenever
// that creature attacks the monarch this turn, it gains double strike and
// trample until end of turn."). The back-reference subject ("that creature")
// otherwise reads as a spurious keyword grant on a referenced object that drops
// the attacks-the-monarch condition. The rider is reparsed in the self ("this
// creature") form so it carries an ordinary attacker-declared trigger pattern
// against the monarch; lowering rebinds that pattern to the captured object via
// the ability's preserved back-reference. The rider's body keywords belong to
// the reparsed nested ability, so the outer ability's semantic keywords are
// cleared to keep the sequence fully consumed. It fails closed: any ability
// whose trailing sentence the recognizer does not match, or whose reparsed
// self-form is not exactly one triggered ability, is left unchanged.
func rewriteCapturedAttacksMonarchDelayedTrigger(ability *Ability) {
	if len(ability.Sentences) < 2 {
		return
	}
	last := &ability.Sentences[len(ability.Sentences)-1]
	tokens := semanticEffectTokens(last.Tokens)
	comma := shared.TopLevelIndex(tokens, shared.Comma)
	if comma <= 0 {
		return
	}
	lead := tokens[:comma]
	if !isDelayedThisTurnPreamble(lead) || !leadBindsThatAttacksMonarch(lead) {
		return
	}
	inner, ok := capturedAttacksMonarchInnerText(last.Text)
	if !ok {
		return
	}
	granted, ok := parseDelayedTriggerAbility(inner)
	if !ok {
		return
	}
	var references, subjectReferences []Reference
	for i := range last.Effects {
		references = append(references, last.Effects[i].References...)
		subjectReferences = append(subjectReferences, last.Effects[i].SubjectReferences...)
	}
	last.Effects = []EffectSyntax{{
		Kind:                       EffectDelayedTrigger,
		Span:                       last.Span,
		VerbSpan:                   last.Span,
		ClauseSpan:                 last.Span,
		Text:                       last.Text,
		References:                 references,
		SubjectReferences:          subjectReferences,
		DelayedTriggerAbility:      &granted,
		DelayedTriggerBindAttacker: true,
	}}
	last.Targets = nil
	last.LegacyEffects = false
	ability.SemanticKeywords = nil
}

// leadBindsThatAttacksMonarch reports whether a delayed-trigger preamble names a
// back-referenced permanent ("that creature") that attacks the monarch, the
// captured-object attacks-the-monarch shape the rider rewriter binds.
func leadBindsThatAttacksMonarch(lead []shared.Token) bool {
	if len(lead) < 2 || !equalWord(lead[1], "that") {
		return false
	}
	hasAttacks := false
	hasMonarch := false
	for i := range lead {
		if equalWord(lead[i], "attacks") {
			hasAttacks = true
		}
		if equalWord(lead[i], "monarch") {
			hasMonarch = true
		}
	}
	return hasAttacks && hasMonarch
}

// capturedAttacksMonarchInnerText reconstructs the self-form triggered-ability
// source of a captured-object attacks-the-monarch rider by stripping the "this
// turn" window and rewriting the "that <noun>" back-reference subject to "this
// <noun>" so the result is an ordinary source-self attacker-declared trigger
// ("Whenever that creature attacks the monarch this turn, it gains double strike
// and trample until end of turn." -> "Whenever this creature attacks the
// monarch, it gains double strike and trample until end of turn."). Lowering
// rebinds the resulting pattern to the captured object, so the self form
// supplies only the attacks-the-monarch event shape. It fails closed on any
// other preamble.
func capturedAttacksMonarchInnerText(text string) (inner string, ok bool) {
	trimmed := strings.TrimSpace(text)
	comma := strings.Index(trimmed, ",")
	if comma <= 0 {
		return "", false
	}
	preamble := strings.TrimSpace(trimmed[:comma])
	body := trimmed[comma:]
	lowered := strings.ToLower(preamble)
	if !strings.HasSuffix(lowered, "this turn") {
		return "", false
	}
	preamble = strings.TrimSpace(preamble[:len(preamble)-len("this turn")])
	lowered = strings.ToLower(preamble)
	var rest string
	switch {
	case strings.HasPrefix(lowered, "whenever that "):
		rest = preamble[len("whenever that "):]
	case strings.HasPrefix(lowered, "when that "):
		rest = preamble[len("when that "):]
	default:
		return "", false
	}
	if !strings.Contains(strings.ToLower(rest), "attacks the monarch") {
		return "", false
	}
	return "Whenever this " + rest + body, true
}

// rewriteCapturedDiesMonarchDelayedTrigger rewrites a trailing "When the
// creature an opponent controls dies this turn, if you control your commander,
// you become the monarch." rider sentence into one EffectDelayedTrigger whose
// permanent-died event binds to the permanent an earlier clause in the same
// resolution acted on: the second fight target ("... it fights target creature
// an opponent controls. When the creature an opponent controls dies this turn,
// ..."). The definite-description subject ("the creature an opponent controls")
// otherwise reads as a spurious resolving become-the-monarch effect gated by an
// intervening commander condition the ordered-sequence lowerer cannot gate. The
// rider is reparsed in the self ("this creature") form so it carries an ordinary
// permanent-died trigger pattern with its "if you control your commander"
// intervening condition; lowering rebinds that pattern to the captured object
// and preserves the intervening condition. The rider's condition belongs to the
// reparsed nested ability, so the outer ability's conditions inside the rider
// sentence are cleared to keep the sequence fully consumed. It fails closed: any
// ability whose trailing sentence the recognizer does not match, or whose
// reparsed self-form is not exactly one triggered ability, is left unchanged.
func rewriteCapturedDiesMonarchDelayedTrigger(ability *Ability) {
	if len(ability.Sentences) < 2 {
		return
	}
	last := &ability.Sentences[len(ability.Sentences)-1]
	tokens := semanticEffectTokens(last.Tokens)
	comma := shared.TopLevelIndex(tokens, shared.Comma)
	if comma <= 0 {
		return
	}
	lead := tokens[:comma]
	if !isDelayedThisTurnPreamble(lead) || !leadBindsDiesOpponentControls(lead) {
		return
	}
	inner, ok := capturedDiesInnerText(last.Text)
	if !ok {
		return
	}
	granted, ok := parseDelayedTriggerAbility(inner)
	if !ok {
		return
	}
	var references, subjectReferences []Reference
	for i := range last.Effects {
		references = append(references, last.Effects[i].References...)
		subjectReferences = append(subjectReferences, last.Effects[i].SubjectReferences...)
	}
	last.Effects = []EffectSyntax{{
		Kind:                          EffectDelayedTrigger,
		Span:                          last.Span,
		VerbSpan:                      last.Span,
		ClauseSpan:                    last.Span,
		Text:                          last.Text,
		References:                    references,
		SubjectReferences:             subjectReferences,
		DelayedTriggerAbility:         &granted,
		DelayedTriggerBindDyingObject: true,
	}}
	last.Targets = nil
	last.LegacyEffects = false
	stripConditionsInSpan(ability, last.Span)
}

// leadBindsDiesOpponentControls reports whether a delayed-trigger preamble names
// the opponent's creature ("the creature an opponent controls") dying, the
// captured-object permanent-died shape the rider rewriter binds.
func leadBindsDiesOpponentControls(lead []shared.Token) bool {
	hasDies := false
	hasOpponent := false
	hasCreature := false
	for i := range lead {
		switch {
		case equalWord(lead[i], "dies"):
			hasDies = true
		case equalWord(lead[i], "opponent"):
			hasOpponent = true
		case equalWord(lead[i], "creature"):
			hasCreature = true
		default:
		}
	}
	return hasDies && hasOpponent && hasCreature
}

// capturedDiesInnerText reconstructs the self-form triggered-ability source of a
// captured-object permanent-died rider by stripping the "this turn" window and
// rewriting the definite-description subject "the creature an opponent controls"
// to "this creature" so the result is an ordinary source-self permanent-died
// trigger ("When the creature an opponent controls dies this turn, if you
// control your commander, you become the monarch." -> "When this creature dies,
// if you control your commander, you become the monarch."). Lowering rebinds the
// resulting pattern to the captured object, so the self form supplies only the
// permanent-died event shape and the intervening condition. It fails closed on
// any other preamble.
func capturedDiesInnerText(text string) (inner string, ok bool) {
	trimmed := strings.TrimSpace(text)
	comma := strings.Index(trimmed, ",")
	if comma <= 0 {
		return "", false
	}
	preamble := strings.TrimSpace(trimmed[:comma])
	body := trimmed[comma:]
	lowered := strings.ToLower(preamble)
	if !strings.HasSuffix(lowered, "this turn") {
		return "", false
	}
	preamble = strings.TrimSpace(preamble[:len(preamble)-len("this turn")])
	lowered = strings.ToLower(preamble)
	const subject = "when the creature an opponent controls "
	if !strings.HasPrefix(lowered, subject) {
		return "", false
	}
	rest := strings.TrimSpace(preamble[len(subject):])
	if strings.ToLower(rest) != "dies" {
		return "", false
	}
	return "When this creature dies" + body, true
}

// stripConditionsInSpan removes every condition boundary, segment, clause, and
// event-history condition whose source position falls within span. It clears a
// rewritten delayed-trigger rider's conditions from the outer ability so its
// "if you control your commander" intervening condition lives only in the
// reparsed nested ability rather than surfacing as an outer per-effect gate.
func stripConditionsInSpan(ability *Ability, span shared.Span) {
	within := func(offset int) bool {
		return offset >= span.Start.Offset && offset < span.End.Offset
	}
	boundaries := ability.ConditionBoundaries[:0]
	for _, boundary := range ability.ConditionBoundaries {
		if !within(boundary.Start.Offset) {
			boundaries = append(boundaries, boundary)
		}
	}
	ability.ConditionBoundaries = boundaries
	segments := ability.ConditionSegments[:0]
	for _, segment := range ability.ConditionSegments {
		if !within(segment.Span.Start.Offset) {
			segments = append(segments, segment)
		}
	}
	ability.ConditionSegments = segments
	clauses := ability.ConditionClauses[:0]
	for _, clause := range ability.ConditionClauses {
		if !within(clause.Span.Start.Offset) {
			clauses = append(clauses, clause)
		}
	}
	ability.ConditionClauses = clauses
	history := ability.EventHistoryConditions[:0]
	for _, condition := range ability.EventHistoryConditions {
		if !within(condition.Span.Start.Offset) {
			history = append(history, condition)
		}
	}
	ability.EventHistoryConditions = history
}

func rewriteDelayedTriggerAbility(ability *Ability) {
	if len(ability.Sentences) != 1 {
		return
	}
	sentence := &ability.Sentences[0]
	tokens := semanticEffectTokens(sentence.Tokens)
	comma := shared.TopLevelIndex(tokens, shared.Comma)
	if comma <= 0 {
		return
	}
	lead := tokens[:comma]
	if !isDelayedThisTurnPreamble(lead) || (!leadMentionsCast(lead) && !leadMentionsCombatDamage(lead)) {
		return
	}
	body := delayedTriggerBodyEffect(sentence, tokens[comma].Span)
	if body == nil {
		return
	}
	inner, oneShot, ok := delayedTriggerInnerText(sentence.Text)
	if !ok {
		return
	}
	granted, ok := parseDelayedTriggerAbility(inner)
	if !ok {
		return
	}
	effect := EffectSyntax{
		Kind:                  EffectDelayedTrigger,
		Span:                  body.Span,
		VerbSpan:              body.VerbSpan,
		ClauseSpan:            body.ClauseSpan,
		Text:                  body.Text,
		DelayedTriggerAbility: &granted,
		DelayedTriggerOneShot: oneShot,
	}
	sentence.Effects = []EffectSyntax{effect}
	sentence.Targets = nil
	sentence.LegacyEffects = false
	ability.SemanticReferences = nil
	ability.SemanticKeywords = nil
	ability.ConditionBoundaries = nil
	ability.EventHistoryConditions = nil
	ability.ConditionClauses = nil
	ability.ConditionSegments = nil
}

func leadMentionsCast(lead []shared.Token) bool {
	for i := range lead {
		if equalWord(lead[i], "cast") {
			return true
		}
	}
	return false
}

// leadMentionsCombatDamage reports whether a delayed-trigger preamble names a
// combat-damage event ("Whenever one or more creatures you control deal combat
// damage to one or more players this turn, ...", Forth Eorlingas!), so the
// spell or ability that sets it up schedules a combat-damage delayed trigger
// rather than reading the preamble as an immediate resolving effect.
func leadMentionsCombatDamage(lead []shared.Token) bool {
	for i := range lead {
		if i+2 < len(lead) &&
			(equalWord(lead[i], "deal") || equalWord(lead[i], "deals")) &&
			equalWord(lead[i+1], "combat") &&
			equalWord(lead[i+2], "damage") {
			return true
		}
	}
	return false
}

// delayedTriggerBodyEffect returns the lone represented effect whose clause
// begins after the preamble comma at commaSpan, the post-comma body whose spans
// the rewritten EffectDelayedTrigger reuses so coverage credits the body clause.
// It returns nil when no single such effect exists so a body the parser split
// across multiple clauses fails closed.
func delayedTriggerBodyEffect(sentence *Sentence, commaSpan shared.Span) *EffectSyntax {
	var match *EffectSyntax
	for i := range sentence.Effects {
		effect := &sentence.Effects[i]
		if effect.Kind == EffectUnknown ||
			effect.ClauseSpan.Start.Offset < commaSpan.End.Offset {
			continue
		}
		if match != nil {
			return nil
		}
		match = effect
	}
	return match
}

// delayedTriggerInnerText reconstructs the nested triggered-ability source of a
// delayed "this turn" cast preamble by stripping the turn window and the "next"
// one-shot marker and normalizing the trigger introducer to "Whenever you cast"
// so the result is an ordinary triggered ability the pipeline parses ("Whenever
// you cast a spell this turn, <body>" -> "Whenever you cast a spell, <body>";
// "When you next cast a creature spell this turn, <body>" -> "Whenever you cast
// a creature spell, <body>"; "The next time you cast a creature spell this turn,
// <body>" -> "Whenever you cast a creature spell, <body>"). The delayed trigger
// reuses only the inner trigger pattern, so normalizing "When"/"the next time"
// to "Whenever" preserves the matched event while avoiding the provenance slot a
// one-shot "When you cast" trigger otherwise requires. oneShot reports the
// "next" forms that fire only on the first match. It fails closed on any other
// preamble shape.
func delayedTriggerInnerText(text string) (inner string, oneShot bool, ok bool) {
	trimmed := strings.TrimSpace(text)
	comma := strings.Index(trimmed, ",")
	if comma <= 0 {
		return "", false, false
	}
	preamble := strings.TrimSpace(trimmed[:comma])
	body := trimmed[comma:]
	lowered := strings.ToLower(preamble)
	if !strings.HasSuffix(lowered, "this turn") {
		return "", false, false
	}
	preamble = strings.TrimSpace(preamble[:len(preamble)-len("this turn")])
	lowered = strings.ToLower(preamble)
	switch {
	case strings.HasPrefix(lowered, "the next time you cast"):
		oneShot = true
		preamble = "Whenever you cast" + preamble[len("the next time you cast"):]
	case strings.HasPrefix(lowered, "when you next cast"):
		oneShot = true
		preamble = "Whenever you cast" + preamble[len("when you next cast"):]
	case strings.HasPrefix(lowered, "whenever you cast"):
	case strings.HasPrefix(lowered, "when you cast"):
		preamble = "Whenever you cast" + preamble[len("when you cast"):]
	case strings.Contains(lowered, "deal combat damage") || strings.Contains(lowered, "deals combat damage"):
		// A combat-damage delayed trigger keeps its whole "Whenever <subject> deal
		// combat damage to <recipient>" preamble ("Whenever one or more creatures
		// you control deal combat damage to one or more players this turn, ...",
		// Forth Eorlingas!); only the turn window is stripped. A "When" introducer
		// is normalized to "Whenever" so the reparsed inner ability is an ordinary
		// repeating triggered ability.
		switch {
		case strings.HasPrefix(lowered, "whenever "):
		case strings.HasPrefix(lowered, "when "):
			preamble = "Whenever " + preamble[len("when "):]
		default:
			return "", false, false
		}
	default:
		return "", false, false
	}
	return strings.TrimSpace(preamble) + body, oneShot, true
}

// parseDelayedTriggerAbility reparses the reconstructed inner ability text
// through the same pipeline so downstream layers lower the delayed trigger from
// the typed inner document. It mirrors parseStaticGrantedAbility but takes raw
// text rather than a quoted token, and requires exactly one triggered ability so
// any other shape fails closed.
func parseDelayedTriggerAbility(text string) (StaticGrantedAbilitySyntax, bool) {
	document, diagnostics := Parse(text, Context{})
	if len(document.Abilities) != 1 ||
		document.Abilities[0].Kind != AbilityTriggered {
		return StaticGrantedAbilitySyntax{}, false
	}
	return StaticGrantedAbilitySyntax{
		Text:        text,
		document:    document,
		diagnostics: diagnostics,
	}, true
}
