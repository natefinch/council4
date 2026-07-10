package parser

import (
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
)

func parseDamageTriggerEventClause(
	tokens []shared.Token,
	_ TriggerIntroductionKind,
	atoms Atoms,
	_ string,
) *TriggerEventClause {
	if syntaxWordsEqual(tokens, "you're", "dealt", "damage") ||
		syntaxWordsEqual(tokens, "you", "are", "dealt", "damage") ||
		syntaxWordsEqual(tokens, "you're", "dealt", "combat", "damage") ||
		syntaxWordsEqual(tokens, "you", "are", "dealt", "combat", "damage") ||
		syntaxWordsEqual(tokens, "you're", "dealt", "noncombat", "damage") ||
		syntaxWordsEqual(tokens, "you", "are", "dealt", "noncombat", "damage") {
		qualifier := TriggerEventCombatQualifier{Kind: TriggerEventCombatQualifierAny}
		switch {
		case slices.Contains(normalizedWords(tokens), "combat"):
			qualifier = TriggerEventCombatQualifier{Kind: TriggerEventCombatQualifierCombat, Span: tokens[len(tokens)-2].Span}
		case slices.Contains(normalizedWords(tokens), "noncombat"):
			qualifier = TriggerEventCombatQualifier{Kind: TriggerEventCombatQualifierNoncombat, Span: tokens[len(tokens)-2].Span}
		default:
		}
		player := playerSelectorFromKind(TriggerPlayerSelectorYou, tokens[0].Span)
		return &TriggerEventClause{
			Kind:            TriggerEventKindDamageDealt,
			Player:          player,
			CombatQualifier: qualifier,
			DamageRecipient: TriggerEventDamageRecipient{
				Kind:   TriggerEventDamageRecipientPlayer,
				Span:   shared.SpanOf(tokens),
				Player: player,
			},
		}
	}
	for _, template := range []struct {
		words     []string
		qualifier TriggerEventCombatQualifierKind
		plural    bool
	}{
		{words: []string{"is", "dealt", "combat", "damage"}, qualifier: TriggerEventCombatQualifierCombat},
		{words: []string{"is", "dealt", "noncombat", "damage"}, qualifier: TriggerEventCombatQualifierNoncombat},
		{words: []string{"is", "dealt", "damage"}},
		{words: []string{"are", "dealt", "combat", "damage"}, qualifier: TriggerEventCombatQualifierCombat, plural: true},
		{words: []string{"are", "dealt", "noncombat", "damage"}, qualifier: TriggerEventCombatQualifierNoncombat, plural: true},
		{words: []string{"are", "dealt", "damage"}, plural: true},
	} {
		prefix, ok := stripTokenSuffix(tokens, template.words...)
		if !ok {
			continue
		}
		subject := parsePermanentEventSubject(prefix, template.plural, atoms)
		if !subject.ok {
			return nil
		}
		return &TriggerEventClause{
			Kind:        TriggerEventKindDamageDealt,
			Subject:     subject.subject,
			Controller:  subject.controller,
			ExcludeSelf: subject.excludeSelf,
			OneOrMore:   subject.oneOrMore,
			CombatQualifier: TriggerEventCombatQualifier{
				Kind: template.qualifier,
				Span: shared.SpanOf(tokens[len(prefix):]),
			},
			DamageRecipient: TriggerEventDamageRecipient{
				Kind: TriggerEventDamageRecipientPermanent,
				Span: shared.SpanOf(tokens[len(prefix):]),
			},
		}
	}
	for _, template := range []struct {
		words     []string
		qualifier TriggerEventCombatQualifierKind
		plural    bool
	}{
		{words: []string{"deals", "combat", "damage"}, qualifier: TriggerEventCombatQualifierCombat},
		{words: []string{"deals", "noncombat", "damage"}, qualifier: TriggerEventCombatQualifierNoncombat},
		{words: []string{"deals", "damage"}},
		{words: []string{"deal", "combat", "damage"}, qualifier: TriggerEventCombatQualifierCombat, plural: true},
		{words: []string{"deal", "noncombat", "damage"}, qualifier: TriggerEventCombatQualifierNoncombat, plural: true},
		{words: []string{"deal", "damage"}, plural: true},
	} {
		index := syntaxWordsIndex(tokens, template.words...)
		if index <= 0 {
			continue
		}
		sourceTokens := tokens[:index]
		recipientTokens := tokens[index+len(template.words):]
		clause := parseDamageSourcePattern(sourceTokens, template.plural, atoms)
		if clause == nil {
			return nil
		}
		clause.Kind = TriggerEventKindDamageDealt
		clause.CombatQualifier = TriggerEventCombatQualifier{
			Kind: template.qualifier,
			Span: shared.SpanOf(tokens[index : index+len(template.words)]),
		}
		if len(recipientTokens) == 0 {
			return clause
		}
		recipientTokens, ok := cutSyntaxWords(recipientTokens, "to")
		if !ok {
			return nil
		}
		recipient, player, ok := parseDamageRecipient(recipientTokens, atoms)
		if !ok {
			return nil
		}
		clause.DamageRecipient = recipient
		clause.Player = player
		return clause
	}
	return nil
}

func parseDamageSourcePattern(tokens []shared.Token, plural bool, atoms Atoms) *TriggerEventClause {
	working := tokens
	oneOrMore := false
	if rest, ok := cutSyntaxWords(working, "one", "or", "more"); ok {
		working = rest
		plural = true
		oneOrMore = true
	}
	if clause := parseSelfOrOtherDamageSource(working, oneOrMore, atoms); clause != nil {
		return clause
	}
	if syntaxWordsEqual(working, "a", "source") {
		return &TriggerEventClause{
			DamageSource: TriggerEventSubject{
				Kind: TriggerEventSubjectDamageSource,
				Span: shared.SpanOf(tokens),
			},
			OneOrMore: oneOrMore,
		}
	}
	if syntaxWordsEqual(working, "a", "spell") {
		return &TriggerEventClause{
			DamageSourceIsStackObject: true,
			StackObject: TriggerEventStackObject{
				Kind: TriggerEventStackObjectSpell,
				Span: shared.SpanOf(working),
			},
			OneOrMore: oneOrMore,
		}
	}
	if selection, controller, ok := parseDamageSpellSource(working, plural); ok {
		return &TriggerEventClause{
			DamageSourceIsStackObject:  true,
			DamageSourceSpellSelection: selection,
			StackObject: TriggerEventStackObject{
				Kind: TriggerEventStackObjectSpell,
				Span: shared.SpanOf(working),
			},
			Controller: controller,
			OneOrMore:  oneOrMore,
		}
	}
	subject := parsePermanentEventSubject(tokens, plural, atoms)
	if !subject.ok {
		return nil
	}
	return &TriggerEventClause{
		DamageSource: subject.subject,
		Controller:   subject.controller,
		ExcludeSelf:  subject.excludeSelf,
		OneOrMore:    subject.oneOrMore || oneOrMore,
	}
}

// parseSelfOrOtherDamageSource recognizes the combat-damage source union
// "this creature or another <Selection> you control" / "this creature or
// equipped creature", where the ability's own source and a self-excluding
// selection (or its equipped permanent) jointly satisfy the trigger. The union
// re-admits the source, so the self-excluding "another"/"equipped" subject is
// kept while SelfOrAnother widens it back to include the source itself.
func parseSelfOrOtherDamageSource(tokens []shared.Token, oneOrMore bool, atoms Atoms) *TriggerEventClause {
	_, count, ok := parseSelfSubject(tokens, atoms)
	if !ok || count >= len(tokens) || !equalWord(tokens[count], "or") {
		return nil
	}
	rest := tokens[count+1:]
	if attached, ok := parseAttachedEventSubject(rest); ok {
		return &TriggerEventClause{
			DamageSource:  attached,
			OneOrMore:     oneOrMore,
			SelfOrAnother: true,
		}
	}
	subject := parsePermanentEventSubject(rest, false, atoms)
	if !subject.ok || !subject.excludeSelf {
		return nil
	}
	return &TriggerEventClause{
		DamageSource:  subject.subject,
		Controller:    subject.controller,
		OneOrMore:     subject.oneOrMore || oneOrMore,
		SelfOrAnother: true,
	}
}

func parseDamageSpellSource(
	tokens []shared.Token,
	plural bool,
) (TriggerEventSpellSelection, TriggerController, bool) {
	working, controller, ok := stripControllerSuffix(tokens)
	if !ok {
		return TriggerEventSpellSelection{}, ControllerAny, false
	}
	if plural && len(working) > 0 && equalWord(working[0], "other") {
		working = working[1:]
	}
	articleLength := 0
	if !plural {
		if len(working) == 0 ||
			(!equalWord(working[0], "a") && !equalWord(working[0], "an")) {
			return TriggerEventSpellSelection{}, ControllerAny, false
		}
		articleLength = 1
	}
	noun := "spell"
	if plural {
		noun = "spells"
	}
	if len(working) <= articleLength || !equalWord(working[len(working)-1], noun) {
		return TriggerEventSpellSelection{}, ControllerAny, false
	}
	phrase := working[articleLength : len(working)-1]
	selection := TriggerEventSpellSelection{Span: shared.SpanOf(working)}
	switch {
	case len(phrase) == 0:
		return selection, controller, true
	case syntaxWordsEqual(phrase, "noncreature"):
		selection.ExcludedTypes = []TriggerCardType{TriggerCardTypeCreature}
		return selection, controller, true
	case len(phrase) == 3 && equalWord(phrase[1], "or"):
		left, leftOK := triggerCardType(phrase[0].Text)
		right, rightOK := triggerCardType(phrase[2].Text)
		if !leftOK || !rightOK ||
			left != TriggerCardTypeInstant && left != TriggerCardTypeSorcery ||
			right != TriggerCardTypeInstant && right != TriggerCardTypeSorcery ||
			left == right {
			return TriggerEventSpellSelection{}, ControllerAny, false
		}
		selection.TypesAny = []TriggerCardType{left, right}
		return selection, controller, true
	case len(phrase) == 1:
		cardType, typeOK := triggerCardType(phrase[0].Text)
		if !typeOK || cardType == TriggerCardTypeUnknown {
			return TriggerEventSpellSelection{}, ControllerAny, false
		}
		selection.Types = []TriggerCardType{cardType}
		return selection, controller, true
	default:
		return TriggerEventSpellSelection{}, ControllerAny, false
	}
}

func parseDamageRecipient(
	tokens []shared.Token,
	atoms Atoms,
) (TriggerEventDamageRecipient, TriggerPlayerSelector, bool) {
	if span, count, ok := parseSelfSubject(tokens, atoms); ok && count == len(tokens) {
		return TriggerEventDamageRecipient{
			Kind:     TriggerEventDamageRecipientPermanent,
			Span:     span,
			IsSource: true,
		}, TriggerPlayerSelector{}, true
	}
	switch {
	case syntaxWordsEqual(tokens, "you"):
		player := playerSelectorFromKind(TriggerPlayerSelectorYou, tokens[0].Span)
		return TriggerEventDamageRecipient{
			Kind:   TriggerEventDamageRecipientPlayer,
			Span:   tokens[0].Span,
			Player: player,
		}, player, true
	case syntaxWordsEqual(tokens, "an", "opponent"), syntaxWordsEqual(tokens, "one", "of", "your", "opponents"):
		player := playerSelectorFromKind(TriggerPlayerSelectorOpponent, shared.SpanOf(tokens))
		return TriggerEventDamageRecipient{
			Kind:   TriggerEventDamageRecipientPlayer,
			Span:   shared.SpanOf(tokens),
			Player: player,
		}, player, true
	case syntaxWordsEqual(tokens, "a", "player"), syntaxWordsEqual(tokens, "one", "or", "more", "players"):
		player := playerSelectorFromKind(TriggerPlayerSelectorAny, shared.SpanOf(tokens))
		return TriggerEventDamageRecipient{
			Kind:   TriggerEventDamageRecipientPlayer,
			Span:   shared.SpanOf(tokens),
			Player: player,
		}, player, true
	case syntaxWordsEqual(tokens, "a", "player", "or", "planeswalker"):
		return TriggerEventDamageRecipient{
			Kind: TriggerEventDamageRecipientPlayer | TriggerEventDamageRecipientPermanent,
			Span: shared.SpanOf(tokens),
			Selection: TriggerSelection{
				RequiredTypes: []TriggerCardType{TriggerCardTypePlaneswalker},
			},
		}, TriggerPlayerSelector{}, true
	case syntaxWordsEqual(tokens, "a", "player", "or", "battle"):
		return TriggerEventDamageRecipient{
			Kind: TriggerEventDamageRecipientPlayer | TriggerEventDamageRecipientPermanent,
			Span: shared.SpanOf(tokens),
			Selection: TriggerSelection{
				RequiredTypes: []TriggerCardType{TriggerCardTypeBattle},
			},
		}, TriggerPlayerSelector{}, true
	case syntaxWordsEqual(tokens, "any", "target"):
		return TriggerEventDamageRecipient{
			Kind: TriggerEventDamageRecipientPlayer | TriggerEventDamageRecipientPermanent,
			Span: shared.SpanOf(tokens),
			Selection: TriggerSelection{
				RequiredTypesAny: []TriggerCardType{
					TriggerCardTypeCreature,
					TriggerCardTypePlaneswalker,
					TriggerCardTypeBattle,
				},
			},
		}, TriggerPlayerSelector{}, true
	}
	selection, ok := parseRelatedSelectionPhrase(tokens)
	if !ok {
		return TriggerEventDamageRecipient{}, TriggerPlayerSelector{}, false
	}
	player := TriggerPlayerSelector{}
	switch selection.Controller {
	case ControllerYou:
		player = playerSelectorFromKind(TriggerPlayerSelectorYou, shared.SpanOf(tokens))
	case ControllerOpponent:
		player = playerSelectorFromKind(TriggerPlayerSelectorOpponent, shared.SpanOf(tokens))
	default:
	}
	return TriggerEventDamageRecipient{
		Kind:      TriggerEventDamageRecipientPermanent,
		Span:      shared.SpanOf(tokens),
		Player:    player,
		Selection: selection,
	}, player, true
}

func parseCounterTriggerEventClause(
	tokens []shared.Token,
	intro TriggerIntroductionKind,
	atoms Atoms,
	_ string,
) *TriggerEventClause {
	if intro != TriggerIntroductionWhenever || len(tokens) == 0 {
		return nil
	}
	if equalWord(tokens[0], "you") && len(tokens) > 1 && equalWord(tokens[1], "put") {
		return parseActiveCounterTriggerEventClause(tokens, tokens[2:], atoms)
	}
	if index := syntaxWordsIndex(tokens, "counter", "is", "put", "on"); index > 1 && equalWord(tokens[0], "a") {
		return buildCounterTriggerEventClause(tokens, tokens[index+4:], atoms, false)
	}
	if index := syntaxWordsIndex(tokens, "counters", "are", "put", "on"); index > 3 && syntaxWordsEqual(tokens[:3], "one", "or", "more") {
		return buildCounterTriggerEventClause(tokens, tokens[index+4:], atoms, true)
	}
	return nil
}

// parseActiveCounterTriggerEventClause recognizes the active-voice
// counter-placement trigger "you put [one or more|a] <kind> counter[s] on
// <subject>" (Exemplar of Light, Terrasymbiosis). The acting player is the
// ability's controller, recorded as a "you" cause controller so the trigger
// fires only for counters the controller places. quantifierTokens are the
// tokens after the leading "you put".
func parseActiveCounterTriggerEventClause(tokens, quantifierTokens []shared.Token, atoms Atoms) *TriggerEventClause {
	oneOrMore := false
	boundary := "counter"
	rest, ok := cutSyntaxWords(quantifierTokens, "one", "or", "more")
	if ok {
		oneOrMore = true
		boundary = "counters"
	} else if rest, ok = cutSyntaxWords(quantifierTokens, "a"); !ok {
		return nil
	}
	index := syntaxWordsIndex(rest, boundary, "on")
	if index < 1 {
		return nil
	}
	subjectTokens := rest[index+2:]
	if len(subjectTokens) == 0 {
		return nil
	}
	clause := buildCounterTriggerEventClause(tokens, subjectTokens, atoms, oneOrMore)
	if clause == nil {
		return nil
	}
	clause.CauseController = TriggerEventActorYou
	return clause
}

func buildCounterTriggerEventClause(
	tokens, subjectTokens []shared.Token,
	atoms Atoms,
	oneOrMore bool,
) *TriggerEventClause {
	counterKind, counterSpan, ok := triggerEventCounterIn(tokens, atoms)
	if !ok {
		return nil
	}
	eventCounter := TriggerEventCounter{Kind: counterKind, Span: counterSpan}
	if syntaxWordsEqual(subjectTokens, "this", "creature") || syntaxWordsEqual(subjectTokens, "this", "permanent") {
		return &TriggerEventClause{
			Kind:      TriggerEventKindCounterAdded,
			Subject:   TriggerEventSubject{Kind: TriggerEventSubjectSelf, Span: shared.SpanOf(subjectTokens)},
			Counter:   eventCounter,
			OneOrMore: oneOrMore,
		}
	}
	subject := parsePermanentEventSubject(subjectTokens, false, atoms)
	if !subject.ok {
		return nil
	}
	if subject.subject.Kind == TriggerEventSubjectSelf {
		return &TriggerEventClause{
			Kind:      TriggerEventKindCounterAdded,
			Subject:   subject.subject,
			Counter:   eventCounter,
			OneOrMore: oneOrMore,
		}
	}
	if subject.subject.Kind != TriggerEventSubjectSelection {
		return nil
	}
	return &TriggerEventClause{
		Kind:        TriggerEventKindCounterAdded,
		Subject:     subject.subject,
		Controller:  subject.controller,
		ExcludeSelf: subject.excludeSelf,
		Counter:     eventCounter,
		OneOrMore:   oneOrMore,
	}
}

func triggerEventCounterIn(tokens []shared.Token, atoms Atoms) (TriggerEventCounterKind, shared.Span, bool) {
	kind, span, ok := atoms.CounterIn(shared.SpanOf(tokens))
	if !ok {
		return TriggerEventCounterAny, shared.Span{}, false
	}
	switch kind {
	case counter.PlusOnePlusOne:
		return TriggerEventCounterPlusOnePlusOne, span, true
	case counter.MinusOneMinusOne:
		return TriggerEventCounterMinusOneMinusOne, span, true
	case counter.Lore:
		return TriggerEventCounterLore, span, true
	default:
		return TriggerEventCounterAny, shared.Span{}, false
	}
}

func parsePermanentStateTriggerEventClause(
	tokens []shared.Token,
	intro TriggerIntroductionKind,
	atoms Atoms,
	_ string,
) *TriggerEventClause {
	for _, template := range []struct {
		suffix    []string
		kind      TriggerEventKind
		allowWhen bool
	}{
		{suffix: []string{"becomes", "tapped"}, kind: TriggerEventKindBecomesTapped, allowWhen: true},
		{suffix: []string{"becomes", "untapped"}, kind: TriggerEventKindBecomesUntapped, allowWhen: true},
		{suffix: []string{"is", "turned", "face", "up"}, kind: TriggerEventKindTurnedFaceUp, allowWhen: true},
	} {
		if intro != TriggerIntroductionWhenever && (intro != TriggerIntroductionWhen || !template.allowWhen) {
			continue
		}
		prefix, ok := stripTokenSuffix(tokens, template.suffix...)
		if !ok {
			continue
		}
		if span, count, ok := parseSelfSubject(prefix, atoms); ok && count == len(prefix) {
			return &TriggerEventClause{
				Kind:    template.kind,
				Subject: TriggerEventSubject{Kind: TriggerEventSubjectSelf, Span: span},
			}
		}
		if intro != TriggerIntroductionWhenever && template.kind != TriggerEventKindTurnedFaceUp {
			return nil
		}
		subject := parsePermanentEventSubject(prefix, false, atoms)
		if !subject.ok || subject.oneOrMore || subject.subject.Kind == TriggerEventSubjectSelf {
			return nil
		}
		return &TriggerEventClause{
			Kind:        template.kind,
			Subject:     subject.subject,
			Controller:  subject.controller,
			ExcludeSelf: subject.excludeSelf,
		}
	}
	return nil
}

// parseTappedForManaTriggerEventClause recognizes "Whenever <subject> is tapped
// for mana" (Wild Growth and the mana-additional aura family), the active-voice
// "Whenever you tap <subject> for mana" (Forbidden Orchard), and the generic
// any-player "Whenever a player taps <subject> for mana" / opponent-scoped
// "Whenever an opponent taps <subject> for mana" (Manabarbs, Mana Flare, War's
// Toll) forms, reusing the becomes-tapped event family with the TappedForMana
// provenance flag set. The any-player form leaves the tapped subject's
// controller unrestricted; the opponent form restricts it to an opponent. The
// active-voice form accepts the source itself ("you tap this land for mana"),
// which is equivalent to the passive self form.
// stripTappedForManaSuffix recognizes the mana-provenance suffix of a
// tapped-for-mana trigger. It accepts the unrestricted "for mana" (returning an
// empty color, matching any produced type) as well as a specific mana symbol
// such as "for {C}" or "for {G}" (returning that color, restricting the trigger
// to taps that produced it). It returns the tokens preceding the suffix.
func stripTappedForManaSuffix(tokens []shared.Token) ([]shared.Token, mana.Color, bool) {
	if inner, ok := stripTokenSuffix(tokens, "for", "mana"); ok {
		return inner, "", true
	}
	if len(tokens) == 0 {
		return nil, "", false
	}
	last := tokens[len(tokens)-1]
	if last.Kind != shared.Symbol {
		return nil, "", false
	}
	color, ok := effectManaColor(last.Text)
	if !ok {
		return nil, "", false
	}
	inner, ok := stripTokenSuffix(tokens[:len(tokens)-1], "for")
	if !ok {
		return nil, "", false
	}
	return inner, color, true
}

func parseTappedForManaTriggerEventClause(
	tokens []shared.Token,
	intro TriggerIntroductionKind,
	atoms Atoms,
	_ string,
) *TriggerEventClause {
	if intro != TriggerIntroductionWhenever {
		return nil
	}
	if rest, ok := cutSyntaxWords(tokens, "you", "tap"); ok {
		inner, color, ok := stripTappedForManaSuffix(rest)
		if !ok {
			return nil
		}
		if span, count, ok := parseSelfSubject(inner, atoms); ok && count == len(inner) {
			return &TriggerEventClause{
				Kind:               TriggerEventKindBecomesTapped,
				Subject:            TriggerEventSubject{Kind: TriggerEventSubjectSelf, Span: span},
				TappedForMana:      true,
				TappedForManaColor: color,
			}
		}
		subject := parsePermanentEventSubject(inner, false, atoms)
		if !subject.ok || subject.oneOrMore || subject.subject.Kind == TriggerEventSubjectSelf {
			return nil
		}
		controller := subject.controller
		if !mergeTriggerController(&controller, ControllerYou) {
			return nil
		}
		return &TriggerEventClause{
			Kind:               TriggerEventKindBecomesTapped,
			Subject:            subject.subject,
			Controller:         controller,
			ExcludeSelf:        subject.excludeSelf,
			TappedForMana:      true,
			TappedForManaColor: color,
		}
	}
	for _, actor := range []struct {
		words      []string
		controller TriggerController
	}{
		{words: []string{"a", "player", "taps"}, controller: ControllerAny},
		{words: []string{"an", "opponent", "taps"}, controller: ControllerOpponent},
	} {
		rest, ok := cutSyntaxWords(tokens, actor.words...)
		if !ok {
			continue
		}
		inner, color, ok := stripTappedForManaSuffix(rest)
		if !ok {
			return nil
		}
		subject := parsePermanentEventSubject(inner, false, atoms)
		if !subject.ok || subject.oneOrMore || subject.subject.Kind == TriggerEventSubjectSelf {
			return nil
		}
		controller := subject.controller
		if !mergeTriggerController(&controller, actor.controller) {
			return nil
		}
		return &TriggerEventClause{
			Kind:               TriggerEventKindBecomesTapped,
			Subject:            subject.subject,
			Controller:         controller,
			ExcludeSelf:        subject.excludeSelf,
			TappedForMana:      true,
			TappedForManaColor: color,
		}
	}
	afterFor, color, ok := stripTappedForManaSuffix(tokens)
	if !ok {
		return nil
	}
	prefix, ok := stripTokenSuffix(afterFor, "is", "tapped")
	if !ok {
		return nil
	}
	if span, count, ok := parseSelfSubject(prefix, atoms); ok && count == len(prefix) {
		return &TriggerEventClause{
			Kind:               TriggerEventKindBecomesTapped,
			Subject:            TriggerEventSubject{Kind: TriggerEventSubjectSelf, Span: span},
			TappedForMana:      true,
			TappedForManaColor: color,
		}
	}
	subject := parsePermanentEventSubject(prefix, false, atoms)
	if !subject.ok || subject.oneOrMore || subject.subject.Kind == TriggerEventSubjectSelf {
		return nil
	}
	return &TriggerEventClause{
		Kind:               TriggerEventKindBecomesTapped,
		Subject:            subject.subject,
		Controller:         subject.controller,
		ExcludeSelf:        subject.excludeSelf,
		TappedForMana:      true,
		TappedForManaColor: color,
	}
}

func parseSacrificeTriggerEventClause(
	tokens []shared.Token,
	_ TriggerIntroductionKind,
	atoms Atoms,
	_ string,
) *TriggerEventClause {
	for _, actor := range []struct {
		words []string
		kind  TriggerEventActorKind
	}{
		{words: []string{"you", "sacrifice"}, kind: TriggerEventActorYou},
		{words: []string{"an", "opponent", "sacrifices"}, kind: TriggerEventActorOpponent},
		{words: []string{"a", "player", "sacrifices"}, kind: TriggerEventActorPlayer},
	} {
		remaining, ok := cutSyntaxWords(tokens, actor.words...)
		if !ok {
			continue
		}
		subject := parsePermanentEventSubject(remaining, false, atoms)
		if !subject.ok || subject.subject.Kind == TriggerEventSubjectAttached {
			return nil
		}
		return &TriggerEventClause{
			Kind:        TriggerEventKindSacrificed,
			Actor:       TriggerEventActor{Kind: actor.kind, Span: shared.SpanOf(tokens[:len(actor.words)])},
			Subject:     subject.subject,
			Controller:  subject.controller,
			ExcludeSelf: subject.excludeSelf,
			OneOrMore:   subject.oneOrMore,
		}
	}
	return nil
}

func parseMutateTriggerEventClause(
	tokens []shared.Token,
	intro TriggerIntroductionKind,
	_ Atoms,
	_ string,
) *TriggerEventClause {
	if intro != TriggerIntroductionWhenever || !syntaxWordsEqual(tokens, "this", "creature", "mutates") {
		return nil
	}
	return &TriggerEventClause{
		Kind: TriggerEventKindMutated,
		Subject: TriggerEventSubject{
			Kind: TriggerEventSubjectSelf,
			Span: shared.SpanOf(tokens[:2]),
		},
	}
}

func parseBecameTargetTriggerEventClause(
	tokens []shared.Token,
	_ TriggerIntroductionKind,
	atoms Atoms,
	_ string,
) *TriggerEventClause {
	index := syntaxWordsIndex(tokens, "becomes", "the", "target", "of")
	if index <= 0 {
		return nil
	}
	subject := parsePermanentEventSubject(tokens[:index], false, atoms)
	if !subject.ok || subject.oneOrMore {
		return nil
	}
	cause := tokens[index+4:]
	firstTimeEachTurn := false
	var firstTimeSpan shared.Span
	if endsWithSyntaxWords(cause, "for", "the", "first", "time", "each", "turn") {
		ordinal := cause[len(cause)-6:]
		firstTimeSpan = shared.SpanOf(ordinal)
		cause = cause[:len(cause)-6]
		firstTimeEachTurn = true
	}
	causeController := TriggerEventActorUnknown
	switch {
	case endsWithSyntaxWords(cause, "you", "control"):
		cause = cause[:len(cause)-2]
		causeController = TriggerEventActorYou
	case endsWithSyntaxWords(cause, "an", "opponent", "controls"):
		cause = cause[:len(cause)-3]
		causeController = TriggerEventActorOpponent
	default:
	}
	var stackObject TriggerEventStackObject
	switch {
	case syntaxWordsEqual(cause, "a", "spell"):
		stackObject = TriggerEventStackObject{Kind: TriggerEventStackObjectSpell, Span: shared.SpanOf(cause)}
	case syntaxWordsEqual(cause, "a", "spell", "or", "ability"):
		stackObject = TriggerEventStackObject{Kind: TriggerEventStackObjectAny, Span: shared.SpanOf(cause)}
	default:
		return nil
	}
	return &TriggerEventClause{
		Kind:                  TriggerEventKindBecameTarget,
		Subject:               subject.subject,
		Controller:            subject.controller,
		ExcludeSelf:           subject.excludeSelf,
		StackObject:           stackObject,
		CauseController:       causeController,
		FirstTimeEachTurn:     firstTimeEachTurn,
		FirstTimeEachTurnSpan: firstTimeSpan,
	}
}
