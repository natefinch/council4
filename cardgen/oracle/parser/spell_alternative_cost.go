package parser

import (
	"slices"
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// SpellAlternativeCostKind identifies the rules change attached to an
// alternative spell cost.
type SpellAlternativeCostKind uint8

// Supported alternative spell-cost kinds.
const (
	SpellAlternativeCostUnknown SpellAlternativeCostKind = iota
	SpellAlternativeCostCommander
	SpellAlternativeCostOverload
	// SpellAlternativeCostPitch is the Force of Will family: exile a colored
	// card from hand (optionally paying extra life) rather than pay the
	// spell's mana cost.
	SpellAlternativeCostPitch
	// SpellAlternativeCostFlashback is the alternative-cost form of Flashback,
	// written "Flashback—<cost>" with an em dash before a non-mana (or compound)
	// cost. The cost itself is carried through the ability's CostSyntax; this
	// kind only marks the paragraph as granting Flashback for that cost.
	SpellAlternativeCostFlashback
	// SpellAlternativeCostEscape is the alternative-cost form of Escape, written
	// "Escape—<cost>, Exile N cards from your graveyard." with an em dash before
	// a compound cost. Like Flashback the cost is carried through the ability's
	// CostSyntax; this kind only marks the paragraph as granting Escape for that
	// cost (CR 702.139).
	SpellAlternativeCostEscape
	// SpellAlternativeCostDiscard is the Foil/Outbreak family: discard one or
	// more cards (each an optional subtype filter) from hand rather than pay the
	// spell's mana cost. The discards are carried as typed cost components on the
	// ability's CostSyntax, like every other non-mana cost.
	SpellAlternativeCostDiscard
	// SpellAlternativeCostBorderpost is "{1}, return a basic land you control"
	// rather than the spell's printed mana cost.
	SpellAlternativeCostBorderpost
	// SpellAlternativeCostFree is the "free spell" family: "[If <condition>,] you
	// may <non-mana payment> rather than pay this spell's mana cost", where the
	// payment is a single non-mana cost (pay life, sacrifice, tap, return, etc.)
	// carried through the ability's CostSyntax like every other non-mana cost.
	// Snuff Out ("If you control a Swamp, you may pay 4 life ...") is the
	// canonical member.
	SpellAlternativeCostFree
	// SpellAlternativeCostMana is the "conditional mana-only" family: "[If
	// <condition>,] you may pay {MANA} rather than pay this spell's mana cost",
	// where the replacement payment is a pure mana cost (including {0}) carried
	// on the SpellAlternativeCost's ManaCost field. The Trap cycle (Mindbreak
	// Trap, Lethargy Trap, Needlebite Trap, ...) and the unconditional Bringer
	// cycle are members. It fails closed on conditions this backend cannot
	// evaluate at cast time.
	SpellAlternativeCostMana
)

// SpellAlternativeCostCondition identifies a condition on an alternative spell cost.
type SpellAlternativeCostCondition uint8

// Supported alternative spell-cost conditions.
const (
	SpellAlternativeCostConditionUnknown SpellAlternativeCostCondition = iota
	SpellAlternativeCostConditionControlsCommander
	// SpellAlternativeCostConditionNotYourTurn gates a pitch alternative cost
	// behind "If it's not your turn," (Force of Negation).
	SpellAlternativeCostConditionNotYourTurn
	// SpellAlternativeCostConditionYourTurn gates a free alternative cost behind
	// "If it's your turn," (Mine Collapse).
	SpellAlternativeCostConditionYourTurn
	// SpellAlternativeCostConditionControlsSubtype gates a free alternative cost
	// behind "If you control a <subtype>," where the subtype is a basic land
	// type (Snuff Out's "If you control a Swamp,"). The subtype rides on the
	// SpellAlternativeCost's ConditionSubtype field.
	SpellAlternativeCostConditionControlsSubtype
	// SpellAlternativeCostConditionOpponentGainedLifeThisTurn gates a mana-only
	// alternative cost behind "If an opponent gained life this turn," (Needlebite
	// Trap). It is the life-gain mirror of the Spectacle life-loss condition.
	SpellAlternativeCostConditionOpponentGainedLifeThisTurn
	// SpellAlternativeCostConditionCreaturesAttacking gates a mana-only
	// alternative cost behind an attacking-creature count: "If N or more
	// creatures are attacking," (Lethargy Trap, Arrow Volley Trap) or "If exactly
	// one creature is attacking," (Pitfall Trap). The threshold rides on
	// ConditionCount and the exact-comparison flag on ConditionExactly.
	SpellAlternativeCostConditionCreaturesAttacking
)

// SpellAlternativeCost is typed syntax for a paragraph that offers an
// alternative to the spell's printed mana cost.
type SpellAlternativeCost struct {
	Span             shared.Span
	Kind             SpellAlternativeCostKind
	Condition        SpellAlternativeCostCondition
	ConditionSubtype types.Sub
	// ConditionCount is the attacking-creature threshold for a
	// SpellAlternativeCostConditionCreaturesAttacking condition; it is unused for
	// every other condition.
	ConditionCount int
	// ConditionExactly requires the attacking-creature count to equal
	// ConditionCount exactly ("If exactly one creature is attacking,") rather
	// than meet it as a minimum ("If N or more creatures are attacking,").
	ConditionExactly      bool
	WithoutPayingManaCost bool
	ManaCost              cost.Mana
	ReplaceTargetWithEach bool
}

func spellAlternativeCostClause(source string, body []shared.Token) (*SpellAlternativeCost, *Cost, bool) {
	if alternative, ok := overloadAlternativeCostClause(body); ok {
		return alternative, nil, true
	}
	if alternative, pitchCost, ok := pitchAlternativeCostClause(body); ok {
		return alternative, pitchCost, true
	}
	if alternative, discardCost, ok := discardAlternativeCostClause(body); ok {
		return alternative, discardCost, true
	}
	if alternative, returnCost, ok := borderpostAlternativeCostClause(source, body); ok {
		return alternative, returnCost, true
	}
	if alternative, ok := manaOnlyAlternativeCostClause(body); ok {
		return alternative, nil, true
	}
	words := []string{
		"if", "you", "control", "a", "commander", "you", "may", "cast",
		"this", "spell", "without", "paying", "its", "mana", "cost",
	}

	if len(body) != len(words)+2 {
		return nil, nil, false
	}
	for tokenIndex, wordIndex := 0, 0; tokenIndex < len(body); tokenIndex++ {
		switch tokenIndex {
		case 5:
			if body[tokenIndex].Kind != shared.Comma {
				return nil, nil, false
			}
		case 3:
			// The commander determiner is printed as either "a commander" or
			// "your commander"; both name the controller's commander.
			if body[tokenIndex].Kind != shared.Word ||
				(!equalWord(body[tokenIndex], "a") && !equalWord(body[tokenIndex], "your")) {
				return nil, nil, false
			}
			wordIndex++
		case len(body) - 1:
			if body[tokenIndex].Kind != shared.Period {
				return nil, nil, false
			}
		default:
			if body[tokenIndex].Kind != shared.Word || !equalWord(body[tokenIndex], words[wordIndex]) {
				return nil, nil, false
			}
			wordIndex++
		}
	}
	return &SpellAlternativeCost{
		Span:                  shared.SpanOf(body),
		Kind:                  SpellAlternativeCostCommander,
		Condition:             SpellAlternativeCostConditionControlsCommander,
		WithoutPayingManaCost: true,
	}, nil, true
}

func borderpostAlternativeCostClause(source string, body []shared.Token) (*SpellAlternativeCost, *Cost, bool) {
	const text = "You may pay {1} and return a basic land you control to its owner's hand rather than pay this spell's mana cost."
	if !strings.EqualFold(strings.TrimSpace(joinedEffectText(body)), text) {
		return nil, nil, false
	}
	rather := -1
	for i := range body {
		if effectWordsAt(body, i, "rather", "than", "pay") {
			rather = i
			break
		}
	}
	if rather <= 5 {
		return nil, nil, false
	}
	returnPhrase := phraseFromTokens(source, body[5:rather])
	parsed := Cost{
		Span: returnPhrase.Span,
		Text: returnPhrase.Text,
		Components: []CostComponent{{
			Kind:             CostComponentReturn,
			Span:             returnPhrase.Span,
			Text:             returnPhrase.Text,
			Amount:           "a",
			Object:           "a basic land you control to its owner's hand",
			AmountValue:      1,
			AmountKnown:      true,
			ObjectNoun:       ObjectNounLand,
			ObjectSupertype:  types.Basic,
			SupertypeKnown:   true,
			ObjectController: ControllerRelationYouControl,
			ToZone:           zone.Hand,
		}},
	}
	return &SpellAlternativeCost{
		Span:     shared.SpanOf(body),
		Kind:     SpellAlternativeCostBorderpost,
		ManaCost: cost.Mana{cost.O(1)},
	}, &parsed, true
}

// flashbackAlternativeCostClause recognizes the em-dash Flashback form
// "Flashback—<cost>", where the pre-dash label is exactly the Flashback keyword
// and the post-dash body is a non-mana or compound cost. The cost tokens become
// the paragraph's cost phrase so the shared cost machinery types them; the
// returned span covers the whole paragraph so its label and dash are accounted
// for in coverage. It returns ok=false when the label is not Flashback or the
// body is empty.
func flashbackAlternativeCostClause(source string, tokens []shared.Token, dash int) (*SpellAlternativeCost, Phrase, bool) {
	if !slices.Equal(normalizedWords(tokens[:dash]), []string{"flashback"}) {
		return nil, Phrase{}, false
	}
	clause := tokens[dash+1:]
	if period := shared.TopLevelIndex(clause, shared.Period); period >= 0 {
		clause = clause[:period]
	}
	if len(clause) == 0 {
		return nil, Phrase{}, false
	}
	return &SpellAlternativeCost{
		Span: shared.SpanOf(tokens),
		Kind: SpellAlternativeCostFlashback,
	}, phraseFromTokens(source, clause), true
}

// escapeAlternativeCostClause recognizes the em-dash Escape form
// "Escape—<cost>, Exile N cards from your graveyard.", where the pre-dash label
// is exactly the Escape keyword and the post-dash body is the compound escape
// cost (its mana cost plus the graveyard-exile additional cost). The cost tokens
// become the paragraph's cost phrase so the shared cost machinery types them;
// the returned span covers the whole paragraph so its label and dash are
// accounted for in coverage. It returns ok=false when the label is not Escape or
// the body is empty.
func escapeAlternativeCostClause(source string, tokens []shared.Token, dash int) (*SpellAlternativeCost, Phrase, bool) {
	if !slices.Equal(normalizedWords(tokens[:dash]), []string{"escape"}) {
		return nil, Phrase{}, false
	}
	clause := tokens[dash+1:]
	if period := shared.TopLevelIndex(clause, shared.Period); period >= 0 {
		clause = clause[:period]
	}
	if len(clause) == 0 {
		return nil, Phrase{}, false
	}
	return &SpellAlternativeCost{
		Span: shared.SpanOf(tokens),
		Kind: SpellAlternativeCostEscape,
	}, phraseFromTokens(source, clause), true
}

// discardAlternativeCostClause recognizes the Foil/Outbreak discard pitch
// family: "you may discard <card>[ and <card>] rather than pay this spell's mana
// cost", where each card is "a/an [<subtype>] card" or "another card". Each
// discarded card is emitted as a typed CostComponentDiscard from hand (with an
// optional subtype filter) so it lowers through the shared cost machinery.
func discardAlternativeCostClause(body []shared.Token) (*SpellAlternativeCost, *Cost, bool) {
	if !equalWordSequence(body, 0, "you", "may", "discard") {
		return nil, nil, false
	}
	cursor := 3
	var components []CostComponent
	for {
		component, next, ok := matchDiscardCardSpec(body, cursor)
		if !ok {
			return nil, nil, false
		}
		components = append(components, component)
		cursor = next
		if cursor < len(body) && equalWord(body[cursor], "and") {
			cursor++
			continue
		}
		break
	}
	if !equalWordSequence(body, cursor,
		"rather", "than", "pay", "this", "spell's", "mana", "cost") {
		return nil, nil, false
	}
	cursor += 7
	if cursor != len(body)-1 || body[cursor].Kind != shared.Period {
		return nil, nil, false
	}
	discardCost := &Cost{
		Span:       shared.SpanOf(body[3:cursor]),
		Components: components,
	}
	return &SpellAlternativeCost{
		Span: shared.SpanOf(body),
		Kind: SpellAlternativeCostDiscard,
	}, discardCost, true
}

// matchDiscardCardSpec parses one "a/an [<subtype>] card" or "another card"
// discard target starting at start, returning a typed CostComponentDiscard and
// the index immediately after the matched "card" noun.
func matchDiscardCardSpec(body []shared.Token, start int) (CostComponent, int, bool) {
	cursor := start
	if cursor >= len(body) || body[cursor].Kind != shared.Word ||
		(!equalWord(body[cursor], "a") && !equalWord(body[cursor], "an") &&
			!equalWord(body[cursor], "another")) {
		return CostComponent{}, 0, false
	}
	cursor++
	component := CostComponent{
		Kind:         CostComponentDiscard,
		AmountValue:  1,
		AmountKnown:  true,
		ObjectIsCard: true,
		ObjectNoun:   ObjectNounCard,
		SourceZone:   zone.Hand,
	}
	if cursor < len(body) && body[cursor].Kind == shared.Word && !equalWord(body[cursor], "card") {
		sub, ok := recognizeSubtypePhrase(body[cursor].Text)
		if !ok {
			return CostComponent{}, 0, false
		}
		component.SubtypesAny = []types.Sub{sub}
		cursor++
	}
	if cursor >= len(body) || !equalWord(body[cursor], "card") {
		return CostComponent{}, 0, false
	}
	cursor++
	component.Span = shared.SpanOf(body[start:cursor])
	return component, cursor, true
}

// freeAlternativeCostClause recognizes the "free spell" family: "[If
// <condition>,] you may <payment> rather than pay this spell's mana cost.",
// where <payment> is a single non-mana cost. It is the general form behind Snuff
// Out ("If you control a Swamp, you may pay 4 life ...") and the sacrifice-cost
// members (Crash, Fireblast, Flare of Malice, ...). The payment tokens are
// returned as a Phrase so the shared cost machinery (emitCost) types them with
// atoms, exactly like a Flashback or Ward cost. It fails closed on multi-part
// payments (any top-level "and", "or", or comma) and on unrecognized leading
// conditions so a mana-only or compound alternative cost is never mistaken for a
// free spell.
func freeAlternativeCostClause(source string, body []shared.Token) (*SpellAlternativeCost, Phrase, bool) {
	cursor := 0
	condition := SpellAlternativeCostConditionUnknown
	conditionSubtype := types.Sub("")
	switch {
	case equalWordSequence(body, cursor, "if", "it's", "not", "your", "turn") && commaAt(body, cursor+5):
		cursor += 6
		condition = SpellAlternativeCostConditionNotYourTurn
	case equalWordSequence(body, cursor, "if", "it's", "your", "turn") && commaAt(body, cursor+4):
		cursor += 5
		condition = SpellAlternativeCostConditionYourTurn
	case equalWordSequence(body, cursor, "if", "you", "control"):
		sub, next, ok := matchControlledBasicLandCondition(body, cursor+3)
		if !ok {
			return nil, Phrase{}, false
		}
		cursor = next
		condition = SpellAlternativeCostConditionControlsSubtype
		conditionSubtype = sub
	default:
		// No recognized leading condition: the free cost is ungated and the
		// cursor stays at the "you may" that must immediately follow.
	}
	if !equalWordSequence(body, cursor, "you", "may") {
		return nil, Phrase{}, false
	}
	cursor += 2
	rather := -1
	for i := cursor; i+7 <= len(body); i++ {
		if equalWordSequence(body, i, "rather", "than", "pay", "this", "spell's", "mana", "cost") {
			rather = i
			break
		}
	}
	if rather < 0 {
		return nil, Phrase{}, false
	}
	after := body[rather+7:]
	if len(after) != 1 || after[0].Kind != shared.Period {
		return nil, Phrase{}, false
	}
	payment := body[cursor:rather]
	if len(payment) == 0 || !singlePartPayment(payment) {
		return nil, Phrase{}, false
	}
	return &SpellAlternativeCost{
		Span:             shared.SpanOf(body),
		Kind:             SpellAlternativeCostFree,
		Condition:        condition,
		ConditionSubtype: conditionSubtype,
	}, phraseFromTokens(source, payment), true
}

// manaOnlyAlternativeCostClause recognizes the "conditional mana-only" family:
// "[If <condition>,] you may pay {MANA} rather than pay this spell's mana cost.",
// where the replacement payment is a pure mana cost (one or more mana symbols,
// including {0}). It backs the Trap cycle (Lethargy Trap, Needlebite Trap, ...)
// and the unconditional Bringer/Mastery cycle. The mana cost rides on the
// returned SpellAlternativeCost's ManaCost field, exactly like Overload.
//
// It fails closed on any leading condition this backend cannot evaluate at cast
// time, on a trailing "if ..." condition, and on any non-mana payment (which is
// left to the free/pitch/discard families), so an unmodeled Trap condition or a
// compound alternative cost is never approximated.
func manaOnlyAlternativeCostClause(body []shared.Token) (*SpellAlternativeCost, bool) {
	cursor := 0
	condition := SpellAlternativeCostConditionUnknown
	conditionCount := 0
	conditionExactly := false
	if equalWord(firstToken(body, cursor), "if") {
		next, matched, ok := matchManaAlternativeCondition(body, cursor)
		if !ok {
			return nil, false
		}
		cursor = next
		condition = matched.Kind
		conditionCount = matched.Count
		conditionExactly = matched.Exactly
	}
	if !equalWordSequence(body, cursor, "you", "may", "pay") {
		return nil, false
	}
	cursor += 3
	manaCost, end, ok := parseKeywordManaCost(body, cursor)
	if !ok || len(manaCost) == 0 {
		return nil, false
	}
	cursor = end
	if !equalWordSequence(body, cursor, "rather", "than", "pay", "this", "spell's", "mana", "cost") {
		return nil, false
	}
	cursor += 7
	if cursor != len(body)-1 || body[cursor].Kind != shared.Period {
		return nil, false
	}
	return &SpellAlternativeCost{
		Span:             shared.SpanOf(body),
		Kind:             SpellAlternativeCostMana,
		Condition:        condition,
		ConditionCount:   conditionCount,
		ConditionExactly: conditionExactly,
		ManaCost:         manaCost,
	}, true
}

// manaAlternativeCondition is a recognized leading gate for a mana-only
// alternative cost, bundling the typed condition kind with the attacking-creature
// threshold (Count) and whether that threshold is an exact match (Exactly).
type manaAlternativeCondition struct {
	Kind    SpellAlternativeCostCondition
	Count   int
	Exactly bool
}

// matchManaAlternativeCondition parses a recognized leading "If <condition>,"
// gate for a mana-only alternative cost, returning the index immediately after
// the condition's comma. It recognizes only the conditions this backend can
// evaluate at cast time; every other "If ..." wording fails closed.
func matchManaAlternativeCondition(body []shared.Token, start int) (next int, condition manaAlternativeCondition, ok bool) {
	// "If an opponent gained life this turn," (Needlebite Trap).
	if equalWordSequence(body, start, "if", "an", "opponent", "gained", "life", "this", "turn") &&
		commaAt(body, start+7) {
		return start + 8, manaAlternativeCondition{Kind: SpellAlternativeCostConditionOpponentGainedLifeThisTurn}, true
	}
	// "If exactly one creature is attacking," (Pitfall Trap).
	if equalWordSequence(body, start, "if", "exactly", "one", "creature", "is", "attacking") &&
		commaAt(body, start+6) {
		return start + 7, manaAlternativeCondition{Kind: SpellAlternativeCostConditionCreaturesAttacking, Count: 1, Exactly: true}, true
	}
	// "If <N> or more creatures are attacking," (Lethargy Trap, Arrow Volley Trap).
	if equalWord(firstToken(body, start), "if") &&
		firstToken(body, start+1).Kind == shared.Word {
		if count, ok := CardinalWordValue(body[start+1].Text); ok && count >= 1 &&
			equalWordSequence(body, start+2, "or", "more", "creatures", "are", "attacking") &&
			commaAt(body, start+7) {
			return start + 8, manaAlternativeCondition{Kind: SpellAlternativeCostConditionCreaturesAttacking, Count: count}, true
		}
	}
	return start, manaAlternativeCondition{Kind: SpellAlternativeCostConditionUnknown}, false
}

// firstToken returns the token at index, or a zero token when index is out of
// range, so leading-condition probes can inspect the first token without a
// bounds guard.
func firstToken(body []shared.Token, index int) shared.Token {
	if index < 0 || index >= len(body) {
		return shared.Token{}
	}
	return body[index]
}

// commaAt reports whether the token at index is a comma.
func commaAt(body []shared.Token, index int) bool {
	return index >= 0 && index < len(body) && body[index].Kind == shared.Comma
}

// matchControlledBasicLandCondition parses the "a/an <basic land subtype> ,"
// tail of a "If you control a Swamp," free-cost condition, returning the matched
// subtype and the index immediately after the comma.
func matchControlledBasicLandCondition(body []shared.Token, start int) (types.Sub, int, bool) {
	cursor := start
	if cursor >= len(body) || (!equalWord(body[cursor], "a") && !equalWord(body[cursor], "an")) {
		return "", 0, false
	}
	cursor++
	if cursor >= len(body) || body[cursor].Kind != shared.Word {
		return "", 0, false
	}
	sub, ok := basicLandSubtype(body[cursor].Text)
	if !ok {
		return "", 0, false
	}
	cursor++
	if !commaAt(body, cursor) {
		return "", 0, false
	}
	return sub, cursor + 1, true
}

// basicLandSubtype maps a basic land type word onto its typed subtype.
func basicLandSubtype(word string) (types.Sub, bool) {
	switch {
	case strings.EqualFold(word, "Plains"):
		return types.Plains, true
	case strings.EqualFold(word, "Island"):
		return types.Island, true
	case strings.EqualFold(word, "Swamp"):
		return types.Swamp, true
	case strings.EqualFold(word, "Mountain"):
		return types.Mountain, true
	case strings.EqualFold(word, "Forest"):
		return types.Forest, true
	default:
		return "", false
	}
}

// singlePartPayment reports that a free alternative cost's payment is a single
// cost component: it carries no top-level "and", "or", or comma that would split
// it into multiple components the free-cost lowering does not model.
func singlePartPayment(payment []shared.Token) bool {
	return len(splitTopLevelWord(payment, "and")) == 1 &&
		len(splitTopLevelWord(payment, "or")) == 1 &&
		shared.TopLevelIndex(payment, shared.Comma) < 0
}

func overloadAlternativeCostClause(body []shared.Token) (*SpellAlternativeCost, bool) {
	if len(body) < 2 || body[0].Kind != shared.Word || !equalWord(body[0], "overload") {
		return nil, false
	}
	manaCost, end, ok := parseKeywordManaCost(body, 1)
	if !ok || len(manaCost) == 0 {
		return nil, false
	}
	if end != len(body) && !canonicalOverloadReminder(body[end:]) {
		return nil, false
	}
	return &SpellAlternativeCost{
		Span:                  shared.SpanOf(body),
		Kind:                  SpellAlternativeCostOverload,
		ManaCost:              manaCost,
		ReplaceTargetWithEach: true,
	}, true
}

// pitchAlternativeCostClause recognizes the Force of Will pitch family:
// "[If it's not your turn, ] you may [pay N life and ] exile a/<count> <color>
// card[s] from your hand rather than pay this spell's mana cost". The non-mana
// payment is emitted as ordered typed cost components (an optional pay-life
// component followed by an exile-from-hand component) so it lowers through the
// same cost machinery as any other ability cost.
func pitchAlternativeCostClause(body []shared.Token) (*SpellAlternativeCost, *Cost, bool) {
	cursor := 0
	condition := SpellAlternativeCostConditionUnknown
	if equalWordSequence(body, cursor, "if", "it's", "not", "your", "turn") {
		cursor += 5
		if cursor >= len(body) || body[cursor].Kind != shared.Comma {
			return nil, nil, false
		}
		cursor++
		condition = SpellAlternativeCostConditionNotYourTurn
	}
	if !equalWordSequence(body, cursor, "you", "may") {
		return nil, nil, false
	}
	cursor += 2
	var components []CostComponent
	if cursor < len(body) && equalWord(body[cursor], "pay") {
		if cursor+3 >= len(body) ||
			body[cursor+1].Kind != shared.Integer ||
			!equalWord(body[cursor+2], "life") ||
			!equalWord(body[cursor+3], "and") {
			return nil, nil, false
		}
		value, ok := conditionNumberValue(body[cursor+1])
		if !ok || value <= 0 {
			return nil, nil, false
		}
		components = append(components, CostComponent{
			Kind:        CostComponentPayLife,
			Span:        shared.SpanOf(body[cursor : cursor+3]),
			AmountValue: value,
			AmountKnown: true,
		})
		cursor += 4
	}
	exileStart := cursor
	exile, ok := matchPitchExileClause(body, cursor)
	if !ok {
		return nil, nil, false
	}
	cursor = exile.next
	components = append(components, CostComponent{
		Kind:             CostComponentExile,
		Span:             shared.SpanOf(body[exileStart:cursor]),
		AmountValue:      exile.count,
		AmountKnown:      true,
		ObjectIsCard:     true,
		ObjectNoun:       ObjectNounCard,
		ObjectColor:      exile.color,
		ObjectColorKnown: true,
		SourceZone:       zone.Hand,
	})
	if !equalWordSequence(body, cursor,
		"from", "your", "hand", "rather", "than", "pay", "this", "spell's", "mana", "cost") {
		return nil, nil, false
	}
	cursor += 10
	if cursor != len(body)-1 || body[cursor].Kind != shared.Period {
		return nil, nil, false
	}
	pitchCost := &Cost{
		Span:       shared.SpanOf(body[exileStart:exile.next]),
		Components: components,
	}
	return &SpellAlternativeCost{
		Span:      shared.SpanOf(body),
		Kind:      SpellAlternativeCostPitch,
		Condition: condition,
	}, pitchCost, true
}

// pitchExileClause is the parsed "exile a/<count> <color> card[s]" segment of a
// pitch alternative cost: the card count, the required color, and the token
// index immediately after the matched "card"/"cards" noun.
type pitchExileClause struct {
	count int
	color Color
	next  int
}

// matchPitchExileClause parses "exile a <color> card" or "exile <count> <color>
// cards" starting at start.
func matchPitchExileClause(body []shared.Token, start int) (pitchExileClause, bool) {
	cursor := start
	if cursor >= len(body) || !equalWord(body[cursor], "exile") {
		return pitchExileClause{}, false
	}
	cursor++
	if cursor >= len(body) || body[cursor].Kind != shared.Word {
		return pitchExileClause{}, false
	}
	count := 0
	if equalWord(body[cursor], "a") {
		count = 1
	} else if value, ok := CardinalWordValue(body[cursor].Text); ok && value >= 1 {
		count = value
	} else {
		return pitchExileClause{}, false
	}
	cursor++
	if cursor >= len(body) || body[cursor].Kind != shared.Word {
		return pitchExileClause{}, false
	}
	color, ok := recognizeColorWord(body[cursor].Text)
	if !ok {
		return pitchExileClause{}, false
	}
	cursor++
	cardNoun := "card"
	if count > 1 {
		cardNoun = "cards"
	}
	if cursor >= len(body) || !equalWord(body[cursor], cardNoun) {
		return pitchExileClause{}, false
	}
	cursor++
	return pitchExileClause{count: count, color: color, next: cursor}, true
}

func equalWordSequence(body []shared.Token, start int, words ...string) bool {
	if start < 0 || start+len(words) > len(body) {
		return false
	}
	for offset, word := range words {
		if !equalWord(body[start+offset], word) {
			return false
		}
	}
	return true
}

func canonicalOverloadReminder(tokens []shared.Token) bool {
	if len(tokens) < 2 || tokens[0].Kind != shared.LeftParen || tokens[len(tokens)-1].Kind != shared.RightParen {
		return false
	}
	var normalized strings.Builder
	for _, token := range tokens {
		_, _ = normalized.WriteString(strings.ToLower(token.Text))
	}
	switch normalized.String() {
	case `(youmaycastthisspellforitsoverloadcost.ifyoudo,change"target"initstextto"each.")`,
		`(youmaycastthisspellforitsoverloadcost.ifyoudo,changeitstextbyreplacingallinstancesof"target"with"each.")`:
		return true
	default:
		return false
	}
}
