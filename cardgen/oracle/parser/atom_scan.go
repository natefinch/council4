package parser

import (
	"slices"
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// collectAtoms recognizes every reusable atom within the semantic tokens of a
// syntax node and returns them with their source spans. Reminder and quoted
// spans are excluded so that recognized meaning matches the semantic tokens the
// compiler consumes.
func collectAtoms(tokens []shared.Token, reminders, quoted []Delimited, cardName string) Atoms {
	tokens = atomSemanticTokens(tokens, reminders, quoted)
	atoms := Atoms{
		references:        collectReferences(tokens, cardName),
		selfNameSpans:     collectSelfNameSpans(tokens, cardName),
		sourceNameSpans:   collectSourceNameSpans(tokens, cardName),
		sourceMarkerSpans: collectSourceMarkerSpans(tokens),
	}
	for _, token := range tokens {
		if token.Kind != shared.Word {
			continue
		}
		if color, ok := recognizeColorWord(token.Text); ok {
			appendAtomColor(&atoms, color, token.Span)
		}
		if rest, ok := strings.CutPrefix(strings.ToLower(token.Text), "non"); ok {
			if color, colorOK := recognizeColorWord(rest); colorOK {
				appendAtomExcludedColor(&atoms, color, token.Span)
			}
			if cardType, typeOK := recognizeCardTypeWord(rest); typeOK {
				appendAtomExcludedType(&atoms, cardType, token.Span)
			}
			if supertype, superOK := recognizeSupertypeWord(rest); superOK {
				appendAtomExcludedSupertype(&atoms, supertype, token.Span)
			}
		}
		if qualifier, ok := recognizeColorQualifierWord(token.Text); ok {
			appendAtomColorQualifier(&atoms, qualifier, token.Span)
		}
		if cardType, ok := recognizeCardTypeWord(token.Text); ok {
			appendAtomCardType(&atoms, cardType, token.Span)
		}
		if supertype, ok := recognizeSupertypeWord(token.Text); ok {
			appendAtomSupertype(&atoms, supertype, token.Span)
		}
		if noun, ok := recognizeObjectNoun(token); ok {
			appendAtomObjectNoun(&atoms, noun, token.Span)
		}
		if value, ok := CardinalWordValue(token.Text); ok {
			appendAtomCardinal(&atoms, value, token.Span)
		}
		if value, ok := OrdinalWordValue(token.Text); ok {
			appendAtomOrdinal(&atoms, value, token.Span)
		}
		if flag, ok := recognizeSelectionFlag(token.Text); ok {
			appendAtomSelectionFlag(&atoms, flag, token.Span)
		}
	}
	for _, atom := range scanSubtypes(tokens) {
		appendAtomSubtype(&atoms, atom.Identity, atom.Span)
	}
	for _, atom := range scanControllerRelations(tokens) {
		appendAtomController(&atoms, atom.Relation, atom.Span)
	}
	for _, atom := range scanZones(tokens) {
		appendAtomZone(&atoms, atom.Zone, atom.Role, atom.Span)
	}
	for _, atom := range scanCounters(tokens) {
		appendAtomCounter(&atoms, atom.Kind, atom.Span)
	}
	atoms.keywords = scanKeywords(tokens, atoms)
	atoms.keywordSelectors = scanKeywordSelectors(tokens)
	return atoms
}

func atomSemanticTokens(tokens []shared.Token, reminders, quoted []Delimited) []shared.Token {
	if len(reminders) == 0 && len(quoted) == 0 {
		return tokens
	}
	excluded := append(append([]Delimited(nil), reminders...), quoted...)
	result := make([]shared.Token, 0, len(tokens))
	for _, token := range tokens {
		skip := false
		for _, delimiter := range excluded {
			if token.Span.Start.Offset >= delimiter.Span.Start.Offset &&
				token.Span.End.Offset <= delimiter.Span.End.Offset {
				skip = true
				break
			}
		}
		if !skip {
			result = append(result, token)
		}
	}
	return result
}

// scanZones recognizes origin and destination zone phrases and emits a zone atom
// for each occurrence in source order.
func scanZones(tokens []shared.Token) []ZoneAtom {
	var atoms []ZoneAtom
	for i := range tokens {
		switch {
		case equalWord(tokens[i], "from") && i+1 < len(tokens):
			if zoneValue, ok := zonePhrase(tokens[i+1:]); ok {
				atoms = appendZone(atoms, tokens, i, zoneValue, ZoneRoleFrom)
			}
		case (equalWord(tokens[i], "to") || equalWord(tokens[i], "into") || equalWord(tokens[i], "onto")) && i+1 < len(tokens):
			if zoneValue, ok := zonePhrase(tokens[i+1:]); ok {
				atoms = appendZone(atoms, tokens, i, zoneValue, ZoneRoleTo)
			}
		case equalWord(tokens[i], "on") && i+3 < len(tokens) &&
			(equalWord(tokens[i+1], "top") || equalWord(tokens[i+1], "bottom")) &&
			equalWord(tokens[i+2], "of"):
			if zoneValue, ok := zonePhrase(tokens[i+3:]); ok {
				atoms = appendZone(atoms, tokens, i, zoneValue, ZoneRoleTo)
			}
		case equalWord(tokens[i], "on") && i+4 < len(tokens) &&
			equalWord(tokens[i+1], "the") &&
			(equalWord(tokens[i+2], "top") || equalWord(tokens[i+2], "bottom")) &&
			equalWord(tokens[i+3], "of"):
			if zoneValue, ok := zonePhrase(tokens[i+4:]); ok {
				atoms = appendZone(atoms, tokens, i, zoneValue, ZoneRoleTo)
			}
		case equalWord(tokens[i], "or") && i+1 < len(tokens):
			if zoneValue, ok := zonePhrase(tokens[i+1:]); ok {
				atoms = appendZone(atoms, tokens, i, zoneValue, ZoneRoleTo)
			}
		default:
		}
	}
	return atoms
}

func zonePhrase(tokens []shared.Token) (zone.Type, bool) {
	switch {
	case graveyardZonePhrase(tokens):
		return zone.Graveyard, true
	case battlefieldZonePhrase(tokens):
		return zone.Battlefield, true
	case handZonePhrase(tokens):
		return zone.Hand, true
	case libraryZonePhrase(tokens):
		return zone.Library, true
	case exileZonePhrase(tokens):
		return zone.Exile, true
	case commandZonePhrase(tokens):
		return zone.Command, true
	default:
		return zone.None, false
	}
}

func appendZone(atoms []ZoneAtom, tokens []shared.Token, i int, value zone.Type, role ZoneRole) []ZoneAtom {
	return append(atoms, ZoneAtom{
		Zone: value,
		Role: role,
		Span: tokens[i].Span,
	})
}

// counterKindNames lists the counter kinds the parser recognizes by name, in the
// priority order the compiler historically matched them.
var counterKindNames = []counter.Kind{
	counter.PlusOnePlusOne,
	counter.MinusOneMinusOne,
	counter.Loyalty,
	counter.Charge,
	counter.Time,
	counter.Defense,
	counter.Poison,
	counter.Lore,
	counter.Verse,
	counter.Shield,
	counter.Stun,
	counter.Finality,
	counter.Brick,
	counter.Page,
	counter.Enlightened,
	counter.Oil,
	counter.Blood,
	counter.Indestructible,
	counter.Deathtouch,
	counter.Flying,
	counter.FirstStrike,
	counter.Hexproof,
	counter.Lifelink,
	counter.Menace,
	counter.Reach,
	counter.Trample,
	counter.Vigilance,
	counter.Energy,
	counter.Experience,
	counter.Burden,
	counter.Age,
}

// scanCounters emits a counter atom for each "<kind> counter(s)" phrase, spanning
// the kind-name tokens that immediately precede the counter noun.
func scanCounters(tokens []shared.Token) []CounterAtom {
	var atoms []CounterAtom
	for i := range tokens {
		if !equalWord(tokens[i], "counter") && !equalWord(tokens[i], "counters") {
			continue
		}
		if kind, span, ok := counterNameBefore(tokens, i); ok {
			atoms = append(atoms, CounterAtom{Kind: kind, Span: span})
		}
	}
	return atoms
}

func counterNameBefore(tokens []shared.Token, counterIndex int) (counter.Kind, shared.Span, bool) {
	for start := counterIndex - 1; start >= 0; start-- {
		name := strings.ToLower(joinTokens(tokens[start:counterIndex]))
		if kind, ok := counterKindAlias(name); ok {
			return kind, shared.SpanOf(tokens[start:counterIndex]), true
		}
		for _, kind := range counterKindNames {
			if name == kind.String() {
				return kind, shared.SpanOf(tokens[start:counterIndex]), true
			}
		}
	}
	return 0, shared.Span{}, false
}

func counterKindAlias(name string) (counter.Kind, bool) {
	switch name {
	case "storage", "fuse":
		return counter.Charge, true
	default:
		return 0, false
	}
}

func scanControllerRelations(tokens []shared.Token) []ControllerRelationAtom {
	patterns := []struct {
		words    []string
		relation ControllerRelation
	}{
		{[]string{"you", "control"}, ControllerRelationYouControl},
		{[]string{"you", "don't", "control"}, ControllerRelationYouDontControl},
		{[]string{"an", "opponent", "controls"}, ControllerRelationOpponentControls},
		{[]string{"your", "opponents", "control"}, ControllerRelationOpponentControls},
		{[]string{"you", "own"}, ControllerRelationYouOwn},
		{[]string{"an", "opponent", "owns"}, ControllerRelationOpponentOwns},
	}
	var atoms []ControllerRelationAtom
	for i := range tokens {
		for _, pattern := range patterns {
			if atomWordsAt(tokens, i, pattern.words...) {
				atoms = append(atoms, ControllerRelationAtom{Relation: pattern.relation, Span: shared.SpanOf(tokens[i : i+len(pattern.words)])})
			}
		}
	}
	return atoms
}

func allWordTokens(tokens []shared.Token) bool {
	for _, token := range tokens {
		if token.Kind != shared.Word {
			return false
		}
	}
	return len(tokens) > 0
}

func atomWordsAt(tokens []shared.Token, start int, words ...string) bool {
	if start < 0 || start+len(words) > len(tokens) {
		return false
	}
	for i, word := range words {
		if !equalWord(tokens[start+i], word) {
			return false
		}
	}
	return true
}

func scanSubtypes(tokens []shared.Token) []SubtypeAtom {
	var atoms []SubtypeAtom
	used := make([]bool, len(tokens))
	for width := 3; width >= 1; width-- {
		for i := 0; i+width <= len(tokens); i++ {
			if slices.Contains(used[i:i+width], true) || !allWordTokens(tokens[i:i+width]) {
				continue
			}
			if identity, ok := recognizeSubtypePhrase(joinTokens(tokens[i : i+width])); ok {
				atoms = append(atoms, SubtypeAtom{Identity: identity, Span: shared.SpanOf(tokens[i : i+width])})
				for j := i; j < i+width; j++ {
					used[j] = true
				}
			}
		}
	}
	slices.SortFunc(atoms, func(a, b SubtypeAtom) int {
		return a.Span.Start.Offset - b.Span.Start.Offset
	})
	return atoms
}

var subtypeCardFamilies = []types.Card{
	types.Artifact,
	types.Battle,
	types.Creature,
	types.Enchantment,
	types.Instant,
	types.Kindred,
	types.Land,
	types.Planeswalker,
	types.Sorcery,
	types.Plane,
	types.Dungeon,
}

// recognizeSubtypePhrase resolves an Oracle subtype phrase to its canonical
// typed identity, owning capitalization, multiword, and plural normalization.
func recognizeSubtypePhrase(phrase string) (types.Sub, bool) {
	phrase = strings.TrimSpace(phrase)
	if phrase == "" {
		return "", false
	}
	candidates := subtypeIdentityCandidates(phrase)
	for _, candidate := range candidates {
		sub := types.Sub(candidate)
		for _, cardType := range subtypeCardFamilies {
			if types.KnownSubtypeForType(cardType, sub) {
				return sub, true
			}
		}
	}
	return "", false
}

func subtypeIdentityCandidates(phrase string) []string {
	lower := strings.ToLower(phrase)
	switch lower {
	case "children":
		return []string{string(types.Child)}
	case "mice":
		return []string{string(types.Mouse)}
	}
	seen := map[string]struct{}{}
	var candidates []string
	add := func(candidate string) {
		if candidate == "" {
			return
		}
		if _, ok := seen[candidate]; ok {
			return
		}
		seen[candidate] = struct{}{}
		candidates = append(candidates, candidate)
	}
	add(phrase)
	add(titleCaseWord(phrase))
	if strings.Contains(phrase, " ") {
		hyphenated := strings.ReplaceAll(phrase, " ", "-")
		add(hyphenated)
		add(titleCaseWord(hyphenated))
	}
	words := strings.Fields(phrase)
	if len(words) > 0 {
		last := words[len(words)-1]
		for _, singular := range SingularNounForms(last) {
			if singular == last {
				continue
			}
			candidateWords := append([]string(nil), words...)
			candidateWords[len(candidateWords)-1] = singular
			candidate := strings.Join(candidateWords, " ")
			add(candidate)
			add(titleCaseWord(candidate))
			if strings.Contains(candidate, " ") {
				hyphenated := strings.ReplaceAll(candidate, " ", "-")
				add(hyphenated)
				add(titleCaseWord(hyphenated))
			}
		}
	}
	for _, singular := range SingularNounForms(phrase) {
		if singular != phrase {
			add(singular)
			add(titleCaseWord(singular))
		}
	}
	return candidates
}

func titleCaseWord(word string) string {
	if word == "" {
		return ""
	}
	parts := strings.Fields(word)
	if len(parts) > 1 {
		for i := range parts {
			parts[i] = titleCaseWord(parts[i])
		}
		return strings.Join(parts, " ")
	}
	return strings.ToUpper(word[:1]) + strings.ToLower(word[1:])
}

func joinTokens(tokens []shared.Token) string {
	if len(tokens) == 0 {
		return ""
	}
	var builder strings.Builder
	for i, token := range tokens {
		if i > 0 && atomNeedsSpace(tokens[i-1], token) {
			_ = builder.WriteByte(' ')
		}
		_, _ = builder.WriteString(token.Text)
	}
	return builder.String()
}

// atomNeedsSpace mirrors the compiler's needsSemanticSpace so that the joined
// counter-name text matches the spelling the compiler historically compared.
func atomNeedsSpace(previous, current shared.Token) bool {
	if current.Kind == shared.Comma || current.Kind == shared.Period || current.Kind == shared.Colon ||
		current.Kind == shared.Semicolon || current.Kind == shared.RightParen ||
		previous.Kind == shared.LeftParen || previous.Kind == shared.Quote || current.Kind == shared.Quote {
		return false
	}
	if previous.Kind == shared.Plus || previous.Kind == shared.Minus || previous.Kind == shared.Slash ||
		current.Kind == shared.Slash {
		return false
	}
	return previous.Kind != shared.Symbol && current.Kind != shared.Symbol
}
