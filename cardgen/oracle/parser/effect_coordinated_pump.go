package parser

import "github.com/natefinch/council4/cardgen/oracle/shared"

// coordinatedGroupSourceExcludingKind maps a controlled/battlefield creature
// group subject onto the source-EXCLUDING variant that names the same group
// minus the effect's own permanent. The coordinated "<self> and <group> each
// get <p>/<t>" pump (Alandra, Sky Dreamer) pumps the source once on its own and
// every other group member through the excluding variant, so the source is never
// counted twice. Only groups whose "other" variant lowers to a
// source-excluding battlefield group are mapped; every other group kind returns
// ok=false so the coordinated shape fails closed rather than risk double-pumping
// the source.
func coordinatedGroupSourceExcludingKind(kind EffectStaticSubjectKind) (EffectStaticSubjectKind, bool) {
	switch kind {
	case EffectStaticSubjectControlledCreatures:
		return EffectStaticSubjectOtherControlledCreatures, true
	case EffectStaticSubjectControlledCreatureSubtype:
		return EffectStaticSubjectOtherControlledCreatureSubtype, true
	case EffectStaticSubjectAllCreatures:
		return EffectStaticSubjectAllOtherCreatures, true
	case EffectStaticSubjectAllCreatureSubtype:
		return EffectStaticSubjectOtherCreatureSubtype, true
	case EffectStaticSubjectAttackingCreatures:
		return EffectStaticSubjectOtherAttackingCreatures, true
	default:
		return "", false
	}
}

// parseCoordinatedSelfGroupSubject recognizes the subject of a coordinated
// power/toughness pump that names the source permanent alongside a controlled
// creature group: "<self> and <group> each get <p>/<t> …" (Alandra, Sky Dreamer:
// "Alandra and Drakes you control each get +X/+X until end of turn, where X is
// the number of cards in your hand."). <self> is the card's own name or a "this
// creature"/"this permanent" self reference; <group> is any creature group whose
// source-excluding variant lowers to a battlefield group that excludes the
// source. It returns the group's source-excluding EffectStaticSubjectSyntax with
// its Span widened to cover "<group> each" so the exact-reconstruction round-trips
// the "<group> each get <p>/<t>" body (the leading "<self> and" is represented by
// the caller's CoordinatedSourceSubject flag, which drives the separate source
// pump). Any other subject shape fails closed.
func parseCoordinatedSelfGroupSubject(tokens []shared.Token, atoms Atoms) (EffectStaticSubjectSyntax, bool) {
	selfEnd, ok := coordinatedSelfPrefix(tokens, atoms)
	if !ok {
		return EffectStaticSubjectSyntax{}, false
	}
	if selfEnd >= len(tokens) || !equalWord(tokens[selfEnd], "and") {
		return EffectStaticSubjectSyntax{}, false
	}
	groupStart := selfEnd + 1
	eachIndex := -1
	for i := groupStart; i+1 < len(tokens); i++ {
		if equalWord(tokens[i], "each") && staticGroupVerb(tokens[i+1]) {
			eachIndex = i
			break
		}
	}
	if eachIndex <= groupStart {
		return EffectStaticSubjectSyntax{}, false
	}
	groupTokens := tokens[groupStart:eachIndex]
	subject, ok := coordinatedGroupSubjectWords(groupTokens, atoms)
	if !ok {
		return EffectStaticSubjectSyntax{}, false
	}
	excluding, ok := coordinatedGroupSourceExcludingKind(subject.Kind)
	if !ok {
		return EffectStaticSubjectSyntax{}, false
	}
	subject.Kind = excluding
	subject.Span = shared.SpanOf(tokens[groupStart : eachIndex+1])
	return subject, true
}

// coordinatedSelfPrefix reports the token count of a leading self reference in a
// coordinated subject: the card's own name (one or more tokens covered by a
// self-name span) or a "this creature"/"this permanent" self reference. It
// returns ok=false when the subject does not begin with a self reference.
func coordinatedSelfPrefix(tokens []shared.Token, atoms Atoms) (int, bool) {
	if len(tokens) >= 2 && equalWord(tokens[0], "this") &&
		(equalWord(tokens[1], "creature") || equalWord(tokens[1], "permanent")) {
		return 2, true
	}
	selfNames := atoms.SelfNameSpans()
	end := 0
	for end < len(tokens) && spanWithinAny(tokens[end].Span, selfNames) {
		end++
	}
	if end == 0 {
		return 0, false
	}
	return end, true
}

// coordinatedGroupSubjectWords recognizes a static creature-group subject
// standing on its own by reusing the shared parseEffectStaticSubject grammar: it
// appends a synthetic "have" verb so the group sub-parsers, which require a
// trailing group verb, engage over exactly the group tokens. Unlike
// recognizeGroupSubjectWords (used by the this-turn can't-block rule effect,
// which drops subtype and other refinements), it preserves subtype-filtered
// groups because the coordinated pump lowers them faithfully. It fails closed on
// an unrepresentable group.
func coordinatedGroupSubjectWords(groupTokens []shared.Token, atoms Atoms) (EffectStaticSubjectSyntax, bool) {
	if len(groupTokens) == 0 {
		return EffectStaticSubjectSyntax{}, false
	}
	synthetic := append(append([]shared.Token(nil), groupTokens...), shared.Token{Kind: shared.Word, Text: "have"})
	subject := parseEffectStaticSubject(synthetic, atoms)
	if subject.Kind == EffectStaticSubjectNone {
		return EffectStaticSubjectSyntax{}, false
	}
	return subject, true
}
