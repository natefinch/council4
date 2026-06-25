package parser

import (
	"slices"
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

func parseSpellCastTriggerEventClause(
	tokens []shared.Token,
	intro TriggerIntroductionKind,
	atoms Atoms,
	_ string,
) *TriggerEventClause {
	if intro != TriggerIntroductionWhenever {
		return nil
	}
	var actor TriggerEventActor
	var remaining []shared.Token
	switch {
	case len(tokens) >= 2 && equalWord(tokens[0], "you") && equalWord(tokens[1], "cast"):
		actor = TriggerEventActor{Kind: TriggerEventActorYou, Span: shared.SpanOf(tokens[:2])}
		remaining = tokens[2:]
	case len(tokens) >= 3 && equalWord(tokens[0], "a") && equalWord(tokens[1], "player") && equalWord(tokens[2], "casts"):
		actor = TriggerEventActor{Kind: TriggerEventActorPlayer, Span: shared.SpanOf(tokens[:3])}
		remaining = tokens[3:]
	case len(tokens) >= 3 && equalWord(tokens[0], "an") && equalWord(tokens[1], "opponent") && equalWord(tokens[2], "casts"):
		actor = TriggerEventActor{Kind: TriggerEventActorOpponent, Span: shared.SpanOf(tokens[:3])}
		remaining = tokens[3:]
	default:
		return nil
	}
	// "Whenever you cast or copy ..." (magecraft) also matches spell copies. The
	// "or copy" wording only appears with the controller-scoped "you" actor.
	matchCopy := false
	if actor.Kind == TriggerEventActorYou && len(remaining) >= 2 &&
		equalWord(remaining[0], "or") && equalWord(remaining[1], "copy") {
		matchCopy = true
		remaining = remaining[2:]
	}
	// "Whenever you cast a spell that targets this creature" (Heroic) restricts
	// the cast trigger to spells targeting the source permanent. The broader
	// "that targets a creature you control" / "...an opponent controls" forms
	// restrict it to spells targeting a permanent matching a selection. The
	// "that targets <relation>" suffix is stripped before the spell selection is
	// parsed so the remaining "a spell" filter parses normally.
	spellTargetsSource := false
	var spellTargetSelection *TriggerSelection
	if index := syntaxWordsIndex(remaining, "that", "targets"); index > 0 {
		targetTokens := remaining[index+2:]
		if _, count, selfOK := parseSelfSubject(targetTokens, atoms); selfOK && count == len(targetTokens) {
			spellTargetsSource = true
			remaining = remaining[:index]
		} else if relation, relationOK := parseSpellTargetSelection(targetTokens); relationOK {
			spellTargetSelection = &relation
			remaining = remaining[:index]
		}
	}
	// "Whenever you cast a spell during your turn" / "during an opponent's turn"
	// restricts the trigger to spells cast on the controller's own turn or on a
	// turn that isn't theirs. The trailing timing phrase is stripped before the
	// spell selection is parsed.
	turnRelation, remaining := cutSpellCastTurnRelation(remaining)
	selection, ok := parseTriggerEventSpellSelection(remaining)
	actorOrdinal := false
	if !ok {
		selection, ok = parseOrdinalSpellSelectionForActor(remaining, actor.Kind)
		actorOrdinal = ok
	}
	if !ok && turnRelation != TriggerCastTurnRelationNone && actor.Kind == TriggerEventActorYou {
		// "your first spell during each opponent's turn" carries the per-turn
		// "each turn" reset in the timing phrase, so the remaining "your Nth
		// spell" run lacks the trailing "each turn" the ordinary ordinal form
		// requires.
		selection, ok = parseYourOrdinalSpell(remaining)
	}
	if !ok || selection.FromZone.Kind != TriggerEventZoneNone && actor.Kind != TriggerEventActorYou {
		return nil
	}
	if matchCopy && selection.FromZone.Kind != TriggerEventZoneNone {
		return nil
	}
	// A spell copy is not cast, so the "cast or copy" magecraft form does not
	// compose with the "from anywhere other than their hand" cast-provenance
	// restriction.
	if matchCopy && selection.CastNotFromHand {
		return nil
	}
	// "your Nth spell each turn" is a controller-scoped per-turn ordinal; it is
	// not combined with spell copies. The "a player"/"an opponent" actors carry
	// their own "their Nth spell each turn" ordinal via parseOrdinalSpellSelectionForActor.
	if selection.Ordinal != 0 && matchCopy {
		return nil
	}
	if selection.Ordinal != 0 && actor.Kind != TriggerEventActorYou && !actorOrdinal {
		return nil
	}
	return &TriggerEventClause{
		Kind:                  TriggerEventKindSpellCast,
		Actor:                 actor,
		SpellSelection:        selection,
		MatchCopy:             matchCopy,
		SpellTargetsSource:    spellTargetsSource,
		SpellTargetSelection:  spellTargetSelection,
		SpellCastTurnRelation: turnRelation,
	}
}

// parseSpellTargetSelection recognizes the permanent target relation in a "that
// targets ..." spell-cast suffix (for example "a creature you control" or "a
// creature an opponent controls"). It strips the leading article and defers to
// parseTriggerSelection, which folds the controller relation into the returned
// selection. The self-target form ("this creature") is handled by the caller
// through parseSelfSubject; "another"-scoped relations are not recognized here
// because the bare selection cannot carry the exclude-source restriction.
func parseSpellTargetSelection(tokens []shared.Token) (TriggerSelection, bool) {
	if len(tokens) < 2 {
		return TriggerSelection{}, false
	}
	switch {
	case equalWord(tokens[0], "a"), equalWord(tokens[0], "an"):
		tokens = tokens[1:]
	default:
		return TriggerSelection{}, false
	}
	return parseTriggerSelection(tokens)
}

// cutSpellCastTurnRelation strips a trailing "during your turn" / "during an
// opponent's turn" / "during each opponent's turn" timing phrase from a
// spell-cast trigger's token run, reporting the relation and the remaining
// tokens. It reports TriggerCastTurnRelationNone with the original tokens when
// no timing phrase is present.
func cutSpellCastTurnRelation(tokens []shared.Token) (TriggerCastTurnRelation, []shared.Token) {
	if rest, ok := stripTokenSuffix(tokens, "during", "your", "turn"); ok {
		return TriggerCastTurnRelationYourTurn, rest
	}
	if rest, ok := stripTokenSuffix(tokens, "during", "an", "opponent's", "turn"); ok {
		return TriggerCastTurnRelationNotYourTurn, rest
	}
	if rest, ok := stripTokenSuffix(tokens, "during", "each", "opponent's", "turn"); ok {
		return TriggerCastTurnRelationNotYourTurn, rest
	}
	return TriggerCastTurnRelationNone, tokens
}

// parseYourOrdinalSpell resolves a controller-scoped "your Nth spell" run that
// lacks the trailing "each turn" because a "during each opponent's turn" timing
// phrase supplied the per-turn reset ("your first spell during each opponent's
// turn").
func parseYourOrdinalSpell(tokens []shared.Token) (TriggerEventSpellSelection, bool) {
	if len(tokens) != 3 || !equalWord(tokens[0], "your") || !equalWord(tokens[2], "spell") {
		return TriggerEventSpellSelection{}, false
	}
	ordinal, ok := OrdinalWordValue(tokens[1].Text)
	if !ok {
		return TriggerEventSpellSelection{}, false
	}
	return TriggerEventSpellSelection{Span: shared.SpanOf(tokens), Ordinal: ordinal}, true
}

// parseOrdinalSpellSelectionForActor resolves a non-controller "their Nth spell
// each turn" ordinal cast trigger for the "a player"/"an opponent" actors, in
// either the unfiltered form ("their second spell each turn") or a single-noun
// filtered form ("their first creature spell each turn"). The controller-scoped
// "you" actor uses the "your Nth spell each turn" production instead.
func parseOrdinalSpellSelectionForActor(tokens []shared.Token, actor TriggerEventActorKind) (TriggerEventSpellSelection, bool) {
	if actor != TriggerEventActorOpponent && actor != TriggerEventActorPlayer {
		return TriggerEventSpellSelection{}, false
	}
	if len(tokens) != 5 && len(tokens) != 6 {
		return TriggerEventSpellSelection{}, false
	}
	if !equalWord(tokens[0], "their") {
		return TriggerEventSpellSelection{}, false
	}
	ordinal, ok := OrdinalWordValue(tokens[1].Text)
	if !ok {
		return TriggerEventSpellSelection{}, false
	}
	if !syntaxWordsEqual(tokens[len(tokens)-3:], "spell", "each", "turn") {
		return TriggerEventSpellSelection{}, false
	}
	selection := TriggerEventSpellSelection{Span: shared.SpanOf(tokens)}
	if len(tokens) == 6 {
		selection, ok = parseSingleNounSpellSelection(selection, tokens[2])
		if !ok {
			return TriggerEventSpellSelection{}, false
		}
	}
	selection.Ordinal = ordinal
	return selection, true
}

func parseTriggerEventSpellSelection(tokens []shared.Token) (TriggerEventSpellSelection, bool) {
	full := tokens
	fromEntryChoice := false
	if rest, ok := cutTriggerSpellChosenTypeSuffix(tokens); ok {
		tokens = rest
		fromEntryChoice = true
	}
	castNotFromHand := false
	if rest, ok := cutTriggerSpellNotFromHandSuffix(tokens); ok {
		tokens = rest
		castNotFromHand = true
	}
	manaValueAtLeast := 0
	manaValueAtMost := 0
	matchManaValue := false
	matchManaValueAtMost := false
	if rest, value, ok := cutTriggerSpellManaValueAtLeastSuffix(tokens); ok {
		tokens = rest
		manaValueAtLeast = value
		matchManaValue = true
	} else if rest, value, ok := cutTriggerSpellManaValueAtMostSuffix(tokens); ok {
		tokens = rest
		manaValueAtMost = value
		matchManaValue = true
		matchManaValueAtMost = true
	}
	selection, ok := parseTriggerEventSpellSelectionFilter(tokens)
	if !ok {
		return TriggerEventSpellSelection{}, false
	}
	if matchManaValue {
		// The base filter must not already carry a mana-value, ordinal, or
		// from-zone qualifier; those forms compose differently and fail closed.
		if selection.MatchManaValue ||
			selection.Ordinal != 0 ||
			selection.FromZone.Kind != TriggerEventZoneNone {
			return TriggerEventSpellSelection{}, false
		}
		selection.MatchManaValue = true
		if matchManaValueAtMost {
			selection.ManaValueAtMost = manaValueAtMost
		} else {
			selection.ManaValueAtLeast = manaValueAtLeast
		}
		selection.Span = shared.SpanOf(full)
	}
	if fromEntryChoice {
		selection.SubtypeFromEntryChoice = true
		selection.Span = shared.SpanOf(full)
	}
	if castNotFromHand {
		// The cast-provenance restriction does not compose with a mana-value,
		// ordinal, or from-zone qualifier; those forms fail closed.
		if selection.MatchManaValue ||
			selection.Ordinal != 0 ||
			selection.FromZone.Kind != TriggerEventZoneNone {
			return TriggerEventSpellSelection{}, false
		}
		selection.CastNotFromHand = true
		selection.Span = shared.SpanOf(full)
	}
	return selection, true
}

// cutTriggerSpellNotFromHandSuffix strips a trailing "from anywhere other than
// their hand" / "from anywhere other than your hand" cast-provenance phrase from
// a spell-selection token run, reporting the remaining filter tokens. The suffix
// restricts the spell-cast trigger to spells cast from a zone other than the
// caster's hand (Ghostly Pilferer). It reports false when the suffix is absent.
func cutTriggerSpellNotFromHandSuffix(tokens []shared.Token) ([]shared.Token, bool) {
	n := len(tokens)
	if n < 6 {
		return nil, false
	}
	if !equalWord(tokens[n-6], "from") ||
		!equalWord(tokens[n-5], "anywhere") ||
		!equalWord(tokens[n-4], "other") ||
		!equalWord(tokens[n-3], "than") ||
		(!equalWord(tokens[n-2], "their") && !equalWord(tokens[n-2], "your")) ||
		!equalWord(tokens[n-1], "hand") {
		return nil, false
	}
	return tokens[:n-6], true
}

// cutTriggerSpellManaValueAtLeastSuffix strips a trailing "with mana value N or
// greater" qualifier from a spell-selection token run, reporting the remaining
// noun-phrase tokens and the integer threshold. The suffix lowers to the
// Selection mana-value-at-least filter and composes with the typed, colored,
// colorless, subtype, and disjunction noun phrases the remaining tokens parse
// into ("a creature spell with mana value 6 or greater", "a colorless spell
// with mana value 7 or greater"). It reports false when the suffix is absent or
// malformed.
func cutTriggerSpellManaValueAtLeastSuffix(tokens []shared.Token) ([]shared.Token, int, bool) {
	n := len(tokens)
	if n < 6 {
		return nil, 0, false
	}
	if !equalWord(tokens[n-6], "with") ||
		!equalWord(tokens[n-5], "mana") ||
		!equalWord(tokens[n-4], "value") ||
		tokens[n-3].Kind != shared.Integer ||
		!equalWord(tokens[n-2], "or") ||
		!equalWord(tokens[n-1], "greater") {
		return nil, 0, false
	}
	value, ok := integerTokenValue(tokens[n-3])
	if !ok {
		return nil, 0, false
	}
	return tokens[:n-6], value, true
}

// cutTriggerSpellManaValueAtMostSuffix strips a trailing "with mana value N or
// less" (or "or fewer") qualifier from a spell-selection token run, reporting
// the remaining noun-phrase tokens and the integer upper bound. It is the
// symmetric counterpart of cutTriggerSpellManaValueAtLeastSuffix: the suffix
// lowers to the Selection mana-value-at-most filter and composes with the same
// typed, colored, colorless, subtype, and disjunction noun phrases ("a creature
// spell with mana value 3 or less"). It reports false when the suffix is absent
// or malformed.
func cutTriggerSpellManaValueAtMostSuffix(tokens []shared.Token) ([]shared.Token, int, bool) {
	n := len(tokens)
	if n < 6 {
		return nil, 0, false
	}
	if !equalWord(tokens[n-6], "with") ||
		!equalWord(tokens[n-5], "mana") ||
		!equalWord(tokens[n-4], "value") ||
		tokens[n-3].Kind != shared.Integer ||
		!equalWord(tokens[n-2], "or") ||
		(!equalWord(tokens[n-1], "less") && !equalWord(tokens[n-1], "fewer")) {
		return nil, 0, false
	}
	value, ok := integerTokenValue(tokens[n-3])
	if !ok {
		return nil, 0, false
	}
	return tokens[:n-6], value, true
}

// integerTokenValue reads a base-ten integer literal token into its value,
// reporting false for any non-digit character.
func integerTokenValue(token shared.Token) (int, bool) {
	value := 0
	for _, r := range token.Text {
		if r < '0' || r > '9' {
			return 0, false
		}
		value = value*10 + int(r-'0')
	}
	return value, true
}

// cutTriggerSpellChosenTypeSuffix strips a trailing "of the chosen type" phrase
// from a spell-selection token run, reporting the remaining filter tokens. The
// suffix ties the cast spell to the creature subtype the source permanent chose
// as it entered (Vanquisher's Banner).
func cutTriggerSpellChosenTypeSuffix(tokens []shared.Token) ([]shared.Token, bool) {
	n := len(tokens)
	if n < 4 {
		return nil, false
	}
	if syntaxWordsEqual(tokens[n-4:], "of", "the", "chosen", "type") {
		return tokens[:n-4], true
	}
	return nil, false
}

func parseTriggerEventSpellSelectionFilter(tokens []shared.Token) (TriggerEventSpellSelection, bool) {
	selection := TriggerEventSpellSelection{Span: shared.SpanOf(tokens)}
	switch {
	case syntaxWordsEqual(tokens, "a", "spell"):
		return selection, true
	case len(tokens) == 5 &&
		equalWord(tokens[0], "your") &&
		equalWord(tokens[2], "spell") &&
		equalWord(tokens[3], "each") &&
		equalWord(tokens[4], "turn"):
		ordinal, ok := OrdinalWordValue(tokens[1].Text)
		if !ok {
			return TriggerEventSpellSelection{}, false
		}
		selection.Ordinal = ordinal
		return selection, true
	case syntaxWordsEqual(tokens, "a", "kicked", "spell"):
		selection.Kicker = true
		return selection, true
	case syntaxWordsEqual(tokens, "a", "historic", "spell"):
		selection.Historic = true
		return selection, true
	case syntaxWordsEqual(tokens, "a", "spell", "from", "your", "graveyard"):
		selection.FromZone = TriggerEventZone{
			Kind: TriggerEventZoneGraveyard,
			Span: shared.SpanOf(tokens[3:]),
		}
		return selection, true
	case len(tokens) == 5 &&
		equalWord(tokens[0], "a") &&
		equalWord(tokens[1], "noncreature") &&
		tokens[2].Kind == shared.Comma &&
		equalWord(tokens[3], "nonland") &&
		equalWord(tokens[4], "spell"):
		selection.ExcludedTypes = []TriggerCardType{TriggerCardTypeCreature, TriggerCardTypeLand}
		return selection, true
	case len(tokens) >= 5 &&
		(equalWord(tokens[0], "a") || equalWord(tokens[0], "an")) &&
		equalWord(tokens[len(tokens)-1], "spell") &&
		containsWord(tokens, "or"):
		return parseSpellSelectionDisjunction(selection, tokens[1:len(tokens)-1])
	case len(tokens) == 2 && equalWord(tokens[0], "an"):
		cardType, ok := triggerCardType(tokens[1].Text)
		if !ok || cardType != TriggerCardTypeInstant {
			return TriggerEventSpellSelection{}, false
		}
		selection.Types = []TriggerCardType{cardType}
		return selection, true
	case len(tokens) == 3 && (equalWord(tokens[0], "a") || equalWord(tokens[0], "an")) && equalWord(tokens[2], "spell"):
		return parseSingleNounSpellSelection(selection, tokens[1])
	default:
		return TriggerEventSpellSelection{}, false
	}
}

// parseSpellSelectionDisjunction resolves a homogeneous "<noun> or <noun>
// spell" / "<noun>, <noun>, or <noun> spell" disjunction filter, where the nouns
// are passed without the leading article or trailing "spell". Every noun must
// resolve to a card type, or every noun must resolve to a color, or every noun
// must resolve to a card subtype; mixing categories is not expressible in the
// runtime selection model and fails closed, as do duplicate or unrecognized
// nouns. Card types lower to TypesAny, colors to ColorsAny, and subtypes to
// SubtypesAny, each of which the runtime matches as a union.
func parseSpellSelectionDisjunction(selection TriggerEventSpellSelection, nouns []shared.Token) (TriggerEventSpellSelection, bool) {
	words, ok := splitDisjunctionNouns(nouns)
	if !ok || len(words) < 2 {
		return TriggerEventSpellSelection{}, false
	}
	cardTypes := make([]TriggerCardType, 0, len(words))
	allCardTypes := true
	for _, word := range words {
		cardType, ok := triggerCardType(word.Text)
		if !ok || cardType == TriggerCardTypeUnknown {
			allCardTypes = false
			break
		}
		if slices.Contains(cardTypes, cardType) {
			return TriggerEventSpellSelection{}, false
		}
		cardTypes = append(cardTypes, cardType)
	}
	if allCardTypes {
		selection.TypesAny = cardTypes
		return selection, true
	}
	colors := make([]TriggerColor, 0, len(words))
	allColors := true
	for _, word := range words {
		if _, ok := recognizeColorWord(word.Text); !ok {
			allColors = false
			break
		}
		colorValue := triggerColor(word.Text)
		if colorValue == TriggerColorUnknown || slices.Contains(colors, colorValue) {
			return TriggerEventSpellSelection{}, false
		}
		colors = append(colors, colorValue)
	}
	if allColors {
		selection.ColorsAny = colors
		return selection, true
	}
	subtypes := make([]TriggerSubtype, 0, len(words))
	for _, word := range words {
		subtype, ok := recognizeSubtypePhrase(word.Text)
		if !ok || slices.Contains(subtypes, subtype) {
			return TriggerEventSpellSelection{}, false
		}
		subtypes = append(subtypes, subtype)
	}
	selection.SubtypesAny = subtypes
	return selection, true
}

// splitDisjunctionNouns extracts the noun tokens from a disjunction run such as
// "Aura or Equipment" or "Aura, Equipment, or Vehicle", requiring exactly one
// "or" connector that precedes the final noun. Commas separate earlier nouns.
// Malformed runs (missing "or", trailing separators, adjacent nouns) fail
// closed.
func splitDisjunctionNouns(tokens []shared.Token) ([]shared.Token, bool) {
	var nouns []shared.Token
	sawOr := false
	expectNoun := true
	for _, tok := range tokens {
		switch {
		case tok.Kind == shared.Comma:
			if expectNoun {
				return nil, false
			}
			expectNoun = true
		case equalWord(tok, "or"):
			if sawOr {
				return nil, false
			}
			sawOr = true
			expectNoun = true
		default:
			if !expectNoun {
				return nil, false
			}
			nouns = append(nouns, tok)
			expectNoun = false
		}
	}
	if !sawOr || expectNoun {
		return nil, false
	}
	return nouns, true
}

// containsWord reports whether any token equals the given lowercase word.
func containsWord(tokens []shared.Token, word string) bool {
	for _, tok := range tokens {
		if equalWord(tok, word) {
			return true
		}
	}
	return false
}

// parseSingleNounSpellSelection resolves the "a <noun> spell" spell-selection
// form, where the single noun is a card type, a color, the colorless or
// multicolored cardinality, a non-<type> exclusion, or a card subtype. The noun
// token retains its original capitalization so subtype recognition can own
// canonicalization. Card types, colors, and the colorless/multicolored/non-
// forms are tried before subtypes, and an unrecognized noun fails closed.
func parseSingleNounSpellSelection(selection TriggerEventSpellSelection, noun shared.Token) (TriggerEventSpellSelection, bool) {
	word := strings.ToLower(noun.Text)
	if cardType, ok := triggerCardType(word); ok && cardType != TriggerCardTypeUnknown {
		selection.Types = []TriggerCardType{cardType}
		return selection, true
	}
	if _, ok := recognizeColorWord(word); ok {
		selection.ColorsAny = []TriggerColor{triggerColor(noun.Text)}
		return selection, true
	}
	switch word {
	case "colorless":
		selection.Colorless = true
		return selection, true
	case "multicolored":
		selection.Multicolored = true
		return selection, true
	}
	if rest, ok := strings.CutPrefix(word, "non"); ok {
		cardType, cardTypeOK := triggerCardType(rest)
		if cardTypeOK && cardType != TriggerCardTypeUnknown {
			selection.ExcludedTypes = []TriggerCardType{cardType}
			return selection, true
		}
	}
	// "a <Subtype> spell" (for example "an Elf spell" or "a Goblin spell")
	// matches every spell that carries the named card subtype. The single
	// subtype lowers to a one-element SubtypesAny set, exactly like the
	// two-subtype union recognized above, so the runtime spell-cast filter
	// reuses its existing subtype-membership matching.
	if subtype, ok := recognizeSubtypePhrase(noun.Text); ok {
		selection.SubtypesAny = []TriggerSubtype{subtype}
		return selection, true
	}
	return TriggerEventSpellSelection{}, false
}

func parseAbilityActivatedTriggerEventClause(
	tokens []shared.Token,
	intro TriggerIntroductionKind,
	_ Atoms,
	_ string,
) *TriggerEventClause {
	if intro != TriggerIntroductionWhenever {
		return nil
	}
	var actor TriggerEventActor
	var remaining []shared.Token
	switch {
	case len(tokens) >= 2 && equalWord(tokens[0], "you") && equalWord(tokens[1], "activate"):
		actor = TriggerEventActor{Kind: TriggerEventActorYou, Span: shared.SpanOf(tokens[:2])}
		remaining = tokens[2:]
	case len(tokens) >= 3 && equalWord(tokens[0], "an") && equalWord(tokens[1], "opponent") && equalWord(tokens[2], "activates"):
		actor = TriggerEventActor{Kind: TriggerEventActorOpponent, Span: shared.SpanOf(tokens[:3])}
		remaining = tokens[3:]
	case len(tokens) >= 3 && equalWord(tokens[0], "a") && equalWord(tokens[1], "player") && equalWord(tokens[2], "activates"):
		actor = TriggerEventActor{Kind: TriggerEventActorPlayer, Span: shared.SpanOf(tokens[:3])}
		remaining = tokens[3:]
	default:
		return nil
	}
	remaining, ok := stripTokenSuffix(remaining, "that", "isn't", "a", "mana", "ability")
	if !ok {
		return nil
	}
	clause := &TriggerEventClause{
		Kind:               TriggerEventKindAbilityActivated,
		Actor:              actor,
		ExcludeManaAbility: true,
	}
	if syntaxWordsEqual(remaining, "an", "ability") {
		return clause
	}
	remaining, ok = cutSyntaxWords(remaining, "an", "ability", "of")
	if !ok {
		return nil
	}
	selection, ok := parseSingleSelectionPhrase(remaining)
	if !ok {
		return nil
	}
	clause.SourceSelection = selection
	return clause
}

func parseAttackBlockTriggerEventClause(
	tokens []shared.Token,
	_ TriggerIntroductionKind,
	atoms Atoms,
	_ string,
) *TriggerEventClause {
	if clause := parsePlayerAttackTriggerEventClause(tokens); clause != nil {
		return clause
	}
	if index := syntaxWordsIndex(tokens, "becomes", "blocked", "by"); index > 0 {
		subject := parsePermanentEventSubject(tokens[:index], false, atoms)
		if !subject.ok || subject.oneOrMore {
			return nil
		}
		selection, ok := parseRelatedSelectionPhrase(tokens[index+3:])
		if !ok || !basicCreatureTriggerSelection(selection) {
			return nil
		}
		return &TriggerEventClause{
			Kind:             TriggerEventKindBecameBlocked,
			Subject:          subject.subject,
			Controller:       subject.controller,
			ExcludeSelf:      subject.excludeSelf,
			RelatedSelection: selection,
		}
	}
	if prefix, ok := stripTokenSuffix(tokens, "becomes", "blocked"); ok {
		subject := parsePermanentEventSubject(prefix, false, atoms)
		if !subject.ok || subject.oneOrMore {
			return nil
		}
		return &TriggerEventClause{
			Kind:        TriggerEventKindBecameBlocked,
			Subject:     subject.subject,
			Controller:  subject.controller,
			ExcludeSelf: subject.excludeSelf,
		}
	}
	if index := syntaxWordsIndex(tokens, "blocks"); index > 0 {
		subject := parsePermanentEventSubject(tokens[:index], false, atoms)
		if !subject.ok || subject.oneOrMore {
			return nil
		}
		clause := &TriggerEventClause{
			Kind:        TriggerEventKindBlock,
			Subject:     subject.subject,
			Controller:  subject.controller,
			ExcludeSelf: subject.excludeSelf,
		}
		if index+1 == len(tokens) {
			return clause
		}
		related, ok := parseRelatedSelectionPhrase(tokens[index+1:])
		if !ok || !selectionHasType(related, TriggerCardTypeCreature) {
			return nil
		}
		clause.RelatedSelection = related
		return clause
	}
	if index := syntaxWordsIndex(tokens, "block"); index > 0 {
		subject := parsePermanentEventSubject(tokens[:index], true, atoms)
		if !subject.ok {
			return nil
		}
		clause := &TriggerEventClause{
			Kind:        TriggerEventKindBlock,
			Subject:     subject.subject,
			Controller:  subject.controller,
			ExcludeSelf: subject.excludeSelf,
			OneOrMore:   subject.oneOrMore,
		}
		if index+1 == len(tokens) {
			return clause
		}
		related, ok := parseRelatedSelectionPhrase(tokens[index+1:])
		if !ok || !selectionHasType(related, TriggerCardTypeCreature) {
			return nil
		}
		clause.RelatedSelection = related
		return clause
	}
	if index := syntaxWordsIndex(tokens, "attacks"); index > 0 {
		subject := parsePermanentEventSubject(tokens[:index], false, atoms)
		if !subject.ok || subject.oneOrMore {
			return nil
		}
		clause := &TriggerEventClause{
			Kind:        TriggerEventKindAttack,
			Subject:     subject.subject,
			Controller:  subject.controller,
			ExcludeSelf: subject.excludeSelf,
		}
		if index+1 == len(tokens) {
			return clause
		}
		if syntaxWordsEqual(tokens[index+1:], "alone") {
			clause.AttackAlone = true
			return clause
		}
		if syntaxWordsEqual(tokens[index+1:], "and", "isn't", "blocked") ||
			syntaxWordsEqual(tokens[index+1:], "and", "is", "not", "blocked") {
			clause.Kind = TriggerEventKindAttacksUnblocked
			return clause
		}
		if syntaxWordsEqual(tokens[index+1:], "while", "saddled") {
			clause.AttackWhileSaddled = true
			return clause
		}
		recipient, player, ok := parseAttackRecipient(tokens[index+1:])
		if !ok {
			return nil
		}
		clause.AttackRecipient = recipient
		clause.Player = player
		return clause
	}
	if index := syntaxWordsIndex(tokens, "attack"); index > 0 {
		if subjectTokens, count, ok := attackerCountFromOtherCreaturesSuffix(tokens[:index]); ok && index+1 == len(tokens) {
			subject := parsePermanentEventSubject(subjectTokens, false, atoms)
			if subject.ok && !subject.oneOrMore && !subject.excludeSelf &&
				subject.subject.Kind == TriggerEventSubjectSelf {
				return &TriggerEventClause{
					Kind:                 TriggerEventKindAttack,
					Subject:              subject.subject,
					Controller:           subject.controller,
					AttackerCountAtLeast: count,
				}
			}
		}
		subject := parsePermanentEventSubject(tokens[:index], true, atoms)
		if !subject.ok {
			return nil
		}
		clause := &TriggerEventClause{
			Kind:        TriggerEventKindAttack,
			Subject:     subject.subject,
			Controller:  subject.controller,
			ExcludeSelf: subject.excludeSelf,
			OneOrMore:   subject.oneOrMore,
		}
		if index+1 == len(tokens) {
			return clause
		}
		recipient, player, ok := parseAttackRecipient(tokens[index+1:])
		if !ok {
			return nil
		}
		clause.AttackRecipient = recipient
		clause.Player = player
		return clause
	}
	return nil
}

func parsePlayerAttackTriggerEventClause(tokens []shared.Token) *TriggerEventClause {
	var actor TriggerEventActor
	var rest []shared.Token
	switch {
	case len(tokens) >= 2 && syntaxWordsEqual(tokens[:2], "you", "attack"):
		actor = TriggerEventActor{Kind: TriggerEventActorYou, Span: tokens[0].Span}
		rest = tokens[2:]
	case len(tokens) >= 3 && syntaxWordsEqual(tokens[:3], "an", "opponent", "attacks"):
		actor = TriggerEventActor{Kind: TriggerEventActorOpponent, Span: shared.SpanOf(tokens[:2])}
		rest = tokens[3:]
	case len(tokens) >= 3 && syntaxWordsEqual(tokens[:3], "a", "player", "attacks"):
		actor = TriggerEventActor{Kind: TriggerEventActorPlayer, Span: shared.SpanOf(tokens[:2])}
		rest = tokens[3:]
	default:
		return nil
	}
	clause := &TriggerEventClause{
		Kind:      TriggerEventKindAttack,
		Actor:     actor,
		OneOrMore: true,
	}
	if len(rest) == 0 {
		return clause
	}
	if tokenWordsEqual(rest, "with", "creatures") ||
		tokenWordsEqual(rest, "with", "one", "or", "more", "creatures") {
		return clause
	}
	if actor.Kind == TriggerEventActorYou {
		if count, ok := attackWithCreatureCount(rest); ok {
			clause.AttackerCountAtLeast = count
			return clause
		}
	}
	recipient, player, ok := parseAttackRecipient(rest)
	if !ok || recipient.Kind != TriggerEventAttackRecipientPlayer {
		return nil
	}
	clause.OneOrMorePerAttackTarget = true
	clause.Player = player
	clause.AttackRecipient = recipient
	return clause
}

// attackWithCreatureCount recognizes "with <N> or more creatures" for a
// controller-scoped attack trigger and returns the minimum attacker count N. It
// fails closed for "one" (the unrestricted wording handled by its own branch)
// and any count outside the small cardinal-word range.
func attackWithCreatureCount(tokens []shared.Token) (int, bool) {
	if len(tokens) != 5 ||
		!equalWord(tokens[0], "with") ||
		!equalWord(tokens[2], "or") ||
		!equalWord(tokens[3], "more") ||
		!equalWord(tokens[4], "creatures") {
		return 0, false
	}
	count, ok := CardinalWordValue(tokens[1].Text)
	if !ok || count < 2 {
		return 0, false
	}
	return count, true
}

// attackerCountFromOtherCreaturesSuffix recognizes the Battalion-style infix
// "<subject> and at least <N> other creatures" that precedes the "attack" verb
// and returns the subject tokens together with the total minimum attacker count
// N+1 (the source creature plus N other creatures). It fails closed when the
// suffix is absent or the cardinal word is out of range.
func attackerCountFromOtherCreaturesSuffix(tokens []shared.Token) ([]shared.Token, int, bool) {
	const suffixLen = 6 // and at least <N> other creatures
	if len(tokens) <= suffixLen {
		return nil, 0, false
	}
	suffix := tokens[len(tokens)-suffixLen:]
	if !equalWord(suffix[0], "and") ||
		!equalWord(suffix[1], "at") ||
		!equalWord(suffix[2], "least") ||
		!equalWord(suffix[4], "other") ||
		!equalWord(suffix[5], "creatures") {
		return nil, 0, false
	}
	others, ok := CardinalWordValue(suffix[3].Text)
	if !ok || others < 1 {
		return nil, 0, false
	}
	return tokens[:len(tokens)-suffixLen], others + 1, true
}
