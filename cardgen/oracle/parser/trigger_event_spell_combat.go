package parser

import (
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

func parseSpellCastTriggerEventClause(
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
	selection, ok := parseTriggerEventSpellSelection(remaining)
	if !ok || selection.FromZone.Kind != TriggerEventZoneNone && actor.Kind != TriggerEventActorYou {
		return nil
	}
	if matchCopy && selection.FromZone.Kind != TriggerEventZoneNone {
		return nil
	}
	// "your Nth spell each turn" is a controller-scoped per-turn ordinal; it is
	// not combined with spell copies or other actors.
	if selection.Ordinal != 0 && (actor.Kind != TriggerEventActorYou || matchCopy) {
		return nil
	}
	return &TriggerEventClause{
		Kind:           TriggerEventKindSpellCast,
		Actor:          actor,
		SpellSelection: selection,
		MatchCopy:      matchCopy,
	}
}

func parseTriggerEventSpellSelection(tokens []shared.Token) (TriggerEventSpellSelection, bool) {
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
	case len(tokens) == 8 &&
		equalWord(tokens[0], "a") &&
		equalWord(tokens[1], "spell") &&
		equalWord(tokens[2], "with") &&
		equalWord(tokens[3], "mana") &&
		equalWord(tokens[4], "value") &&
		tokens[5].Kind == shared.Integer &&
		equalWord(tokens[6], "or") &&
		equalWord(tokens[7], "greater"):
		value := 0
		for _, r := range tokens[5].Text {
			if r < '0' || r > '9' {
				return TriggerEventSpellSelection{}, false
			}
			value = value*10 + int(r-'0')
		}
		selection.MatchManaValue = true
		selection.ManaValueAtLeast = value
		return selection, true
	case len(tokens) == 5 &&
		equalWord(tokens[0], "a") &&
		equalWord(tokens[1], "noncreature") &&
		tokens[2].Kind == shared.Comma &&
		equalWord(tokens[3], "nonland") &&
		equalWord(tokens[4], "spell"):
		selection.ExcludedTypes = []TriggerCardType{TriggerCardTypeCreature, TriggerCardTypeLand}
		return selection, true
	case len(tokens) == 5 &&
		(equalWord(tokens[0], "a") || equalWord(tokens[0], "an")) &&
		equalWord(tokens[2], "or") &&
		equalWord(tokens[4], "spell"):
		leftType, leftTypeOK := triggerCardType(tokens[1].Text)
		rightType, rightTypeOK := triggerCardType(tokens[3].Text)
		if leftTypeOK && rightTypeOK && leftType != TriggerCardTypeUnknown && rightType != TriggerCardTypeUnknown {
			if leftType != TriggerCardTypeInstant && leftType != TriggerCardTypeSorcery ||
				rightType != TriggerCardTypeInstant && rightType != TriggerCardTypeSorcery ||
				leftType == rightType {
				return TriggerEventSpellSelection{}, false
			}
			selection.TypesAny = []TriggerCardType{leftType, rightType}
			return selection, true
		}
		leftSub, leftSubOK := recognizeSubtypePhrase(strings.ToLower(tokens[1].Text))
		rightSub, rightSubOK := recognizeSubtypePhrase(strings.ToLower(tokens[3].Text))
		if leftSubOK && rightSubOK {
			if leftSub != "Spirit" && leftSub != "Arcane" ||
				rightSub != "Spirit" && rightSub != "Arcane" ||
				leftSub == rightSub {
				return TriggerEventSpellSelection{}, false
			}
			selection.SubtypesAny = []TriggerSubtype{leftSub, rightSub}
			return selection, true
		}
		return TriggerEventSpellSelection{}, false
	case len(tokens) == 2 && equalWord(tokens[0], "an"):
		cardType, ok := triggerCardType(tokens[1].Text)
		if !ok || cardType != TriggerCardTypeInstant {
			return TriggerEventSpellSelection{}, false
		}
		selection.Types = []TriggerCardType{cardType}
		return selection, true
	case len(tokens) == 3 && (equalWord(tokens[0], "a") || equalWord(tokens[0], "an")) && equalWord(tokens[2], "spell"):
		word := strings.ToLower(tokens[1].Text)
		if cardType, ok := triggerCardType(word); ok && cardType != TriggerCardTypeUnknown {
			selection.Types = []TriggerCardType{cardType}
			return selection, true
		}
		if color, ok := recognizeColorWord(word); ok {
			selection.ColorsAny = []TriggerColor{triggerColor(tokens[1].Text)}
			_ = color
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
	default:
		return TriggerEventSpellSelection{}, false
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
		recipient, player, ok := parseAttackRecipient(tokens[index+1:])
		if !ok {
			return nil
		}
		clause.AttackRecipient = recipient
		clause.Player = player
		return clause
	}
	if index := syntaxWordsIndex(tokens, "attack"); index > 0 {
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
	if actor.Kind == TriggerEventActorYou {
		if tokenWordsEqual(rest, "with", "one", "or", "more", "creatures") {
			return clause
		}
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
