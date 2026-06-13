package parser

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// EffectKind identifies a resolving instruction. The parser owns the Oracle
// vocabulary which selects these values; consumers only map the typed value.
type EffectKind uint8

// Resolving effect kinds recognized by the parser.
const (
	EffectUnknown EffectKind = iota
	EffectAddMana
	EffectAttach
	EffectCast
	EffectCounter
	EffectCreate
	EffectDealDamage
	EffectDestroy
	EffectDiscard
	EffectDiscover
	EffectDouble
	EffectDraw
	EffectEnterTapped
	EffectEnterPrepared
	EffectExile
	EffectFight
	EffectGain
	EffectGainControl
	EffectGrantKeyword
	EffectInvestigate
	EffectExplore
	EffectLose
	EffectManifest
	EffectManifestDread
	EffectMill
	EffectModifyPT
	EffectPut
	EffectProliferate
	EffectRegenerate
	EffectReturn
	EffectReveal
	EffectSacrifice
	EffectScry
	EffectSurveil
	EffectSearch
	EffectShuffle
	EffectTap
	EffectUntap
	EffectTransform
)

// EffectDurationKind identifies a resolving effect's duration.
type EffectDurationKind uint8

// Resolving effect durations recognized by the parser.
const (
	EffectDurationNone EffectDurationKind = iota
	EffectDurationUntilEndOfTurn
	EffectDurationUntilYourNextTurn
	EffectDurationThisTurn
	EffectDurationThisCombat
	EffectDurationWhileSourceOnBattlefield
	EffectDurationWhileYouControlSource
)

// DelayedTimingKind identifies a delayed resolving instruction suffix.
type DelayedTimingKind uint8

// Delayed timings recognized by resolving-effect grammar.
const (
	DelayedTimingNone DelayedTimingKind = iota
	DelayedTimingNextEndStep
	DelayedTimingNextUpkeep
)

// EffectDestinationPosition identifies an ordered position in a destination
// zone.
type EffectDestinationPosition uint8

// Ordered destination positions recognized by resolving-effect grammar.
const (
	EffectDestinationUnspecified EffectDestinationPosition = iota
	EffectDestinationTop
	EffectDestinationBottom
)

// EffectDynamicAmountKind identifies a rules-derived amount.
type EffectDynamicAmountKind uint8

// Dynamic resolving amounts recognized by the parser.
const (
	EffectDynamicAmountNone EffectDynamicAmountKind = iota
	EffectDynamicAmountCount
	EffectDynamicAmountControllerLife
	EffectDynamicAmountOpponentCount
	EffectDynamicAmountSourcePower
	EffectDynamicAmountBasicLandTypes
)

// EffectDynamicAmountForm identifies how a dynamic amount is introduced.
type EffectDynamicAmountForm uint8

// Dynamic amount forms recognized by the parser.
const (
	EffectDynamicAmountFormNone EffectDynamicAmountForm = iota
	EffectDynamicAmountFormEqual
	EffectDynamicAmountFormForEach
	EffectDynamicAmountFormWhereX
)

// EffectAmountSyntax is a fixed or rules-derived source-spanned amount.
type EffectAmountSyntax struct {
	Span          shared.Span
	Text          string
	Value         int
	Known         bool
	VariableX     bool
	DynamicKind   EffectDynamicAmountKind
	DynamicForm   EffectDynamicAmountForm
	Multiplier    int
	ReferenceSpan shared.Span
	Selection     *SelectionSyntax
}

// EffectReplacementKind identifies how an instruction replaces an event.
type EffectReplacementKind uint8

// Resolving replacement modifiers recognized by the parser.
const (
	EffectReplacementNone EffectReplacementKind = iota
	EffectReplacementInstead
	EffectReplacementTwiceThatMany
	EffectReplacementThatMuchPlus
	EffectReplacementDoubleThat
)

// EffectReplacementSyntax is a source-spanned replacement modifier.
type EffectReplacementSyntax struct {
	Kind            EffectReplacementKind
	Span            shared.Span
	Amount          int
	EachCounterKind bool
}

// EffectManaSyntax describes exact add-mana output.
type EffectManaSyntax struct {
	Span            shared.Span
	Symbols         []string
	Choice          bool
	AnyColor        bool
	LegacyBodyExact bool
}

// EffectContextKind identifies the grammatical subject performing or receiving
// a resolving instruction.
type EffectContextKind uint8

// Resolving-effect contexts recognized by the parser.
const (
	EffectContextUnknown EffectContextKind = iota
	EffectContextController
	EffectContextTarget
	EffectContextEachOpponent
	EffectContextEachPlayer
	EffectContextEventPlayer
	EffectContextSource
	EffectContextReferencedObject
	EffectContextReferencedPlayer
	EffectContextPriorSubject
)

// SignedAmountSyntax is one signed half of a power/toughness change.
type SignedAmountSyntax struct {
	Span     shared.Span
	Value    int
	Known    bool
	Negative bool
}

// SelectionController identifies a selected object's controller.
type SelectionController uint8

// Selection controller relations.
const (
	SelectionControllerAny SelectionController = iota
	SelectionControllerYou
	SelectionControllerOpponent
	SelectionControllerNotYou
)

// SelectionKind identifies the broad object selected by a phrase.
type SelectionKind uint8

// Selection kinds recognized by resolving-effect grammar.
const (
	SelectionUnknown SelectionKind = iota
	SelectionAny
	SelectionPlayer
	SelectionOpponent
	SelectionArtifact
	SelectionCreature
	SelectionEnchantment
	SelectionLand
	SelectionPermanent
	SelectionCard
	SelectionSpell
	SelectionActivatedAbility
	SelectionTriggeredAbility
	SelectionActivatedOrTriggeredAbility
	SelectionSpellActivatedOrTriggeredAbility
	SelectionPlaneswalker
	SelectionBattle
)

// SelectionSyntax is a typed, source-spanned noun phrase.
type SelectionSyntax struct {
	Span             shared.Span
	Text             string
	Kind             SelectionKind
	Controller       SelectionController
	All              bool
	Another          bool
	Other            bool
	Attacking        bool
	Blocking         bool
	Tapped           bool
	Untapped         bool
	Keyword          KeywordKind
	Zone             zone.Type
	RequiredTypesAny []CardType
	ExcludedTypes    []CardType
	Supertypes       []Supertype
	ColorsAny        []Color
	ExcludedColors   []Color
	SubtypesAny      []types.Sub
	ManaValue        compare.Int
	MatchManaValue   bool
	Power            compare.Int
	MatchPower       bool
	Toughness        compare.Int
	MatchToughness   bool
}

// TargetCardinalitySyntax is an inclusive target-count range.
type TargetCardinalitySyntax struct {
	Min int
	Max int
}

// TargetSyntax is one typed target production.
type TargetSyntax struct {
	Span        shared.Span
	Text        string
	Cardinality TargetCardinalitySyntax
	Selection   SelectionSyntax
	Exact       bool
}

// EffectConnectionKind identifies how a resolving instruction is coordinated
// with the preceding instruction in the same sentence.
type EffectConnectionKind uint8

// Resolving-instruction connections recognized by the parser.
const (
	EffectConnectionNone EffectConnectionKind = iota
	EffectConnectionAnd
	EffectConnectionThen
)

// EffectSyntax is one typed resolving instruction. Text and Tokens remain
// lossless metadata; all meaning consumed downstream is carried by typed fields.
type EffectSyntax struct {
	Kind                    EffectKind
	Context                 EffectContextKind
	Connection              EffectConnectionKind
	ConnectionSpan          shared.Span
	Span                    shared.Span
	VerbSpan                shared.Span
	ClauseSpan              shared.Span
	Text                    string
	Tokens                  []shared.Token
	Duration                EffectDurationKind
	DelayedTiming           DelayedTimingKind
	Selection               SelectionSyntax
	Amount                  EffectAmountSyntax
	PowerDelta              SignedAmountSyntax
	ToughnessDelta          SignedAmountSyntax
	StaticSubject           EffectStaticSubjectSyntax
	CounterKind             counter.Kind
	CounterKnown            bool
	FromZone                zone.Type
	ToZone                  zone.Type
	Destination             EffectDestinationPosition
	EntersTapped            bool
	EntersTappedSelf        bool
	EntersWithCounters      bool
	UnderYourControl        bool
	CastAsAdventure         bool
	Negated                 bool
	Optional                bool
	OptionalSpan            shared.Span
	Symbol                  string
	Mana                    EffectManaSyntax
	Replacement             EffectReplacementSyntax
	References              []Reference
	SubjectReferences       []Reference
	Targets                 []TargetSyntax
	SubjectTargets          []TargetSyntax
	Payment                 EffectPaymentSyntax
	Exact                   bool
	RequiresOrderedLowering bool
	HasUnrecognizedSibling  bool
	UnsupportedDetail       string
}

// EffectPaymentPayerKind identifies who may pay a cost embedded in an effect.
type EffectPaymentPayerKind uint8

// Embedded-effect payers recognized by the parser.
const (
	EffectPaymentPayerUnknown EffectPaymentPayerKind = iota
	EffectPaymentPayerTargetController
)

// EffectPaymentSyntax is a source-spanned typed resolution payment.
type EffectPaymentSyntax struct {
	Span     shared.Span
	Payer    EffectPaymentPayerKind
	ManaCost cost.Mana
}

// EffectStaticSubjectKind identifies the group affected by a static resolving
// effect production.
type EffectStaticSubjectKind uint8

// Static effect subjects recognized by resolving-effect grammar.
const (
	EffectStaticSubjectNone EffectStaticSubjectKind = iota
	EffectStaticSubjectAttachedObject
	EffectStaticSubjectControlledCreatures
	EffectStaticSubjectOtherControlledCreatures
	EffectStaticSubjectControlledWalls
	EffectStaticSubjectControlledArtifacts
	EffectStaticSubjectControlledTokens
	EffectStaticSubjectOpponentControlledCreatures
	EffectStaticSubjectControlledCreatureSubtype
	EffectStaticSubjectOtherControlledCreatureSubtype
)

// EffectStaticSubjectSyntax is a source-spanned typed static-effect subject.
type EffectStaticSubjectSyntax struct {
	Kind         EffectStaticSubjectKind
	Span         shared.Span
	Subtype      types.Sub
	SubtypeText  string
	SubtypeKnown bool
}

func emitResolvingSyntax(abilities []Ability) {
	for i := range abilities {
		emitSentenceResolvingSyntax(abilities[i].Sentences, abilities[i].Atoms, abilities[i].ActivationRestrictions)
		if abilities[i].Modal == nil {
			continue
		}
		for j := range abilities[i].Modal.Options {
			mode := &abilities[i].Modal.Options[j]
			emitSentenceResolvingSyntax(mode.Sentences, mode.Atoms, nil)
		}
	}
}

func emitSentenceResolvingSyntax(sentences []Sentence, atoms Atoms, restrictions []ActivationRestriction) {
	legacyEffects := 0
	currentEffects := 0
	unrecognizedSibling := false
	for i := range sentences {
		if sentences[i].StaticRule != nil || spanInsideActivationRestriction(sentences[i].Span, restrictions) {
			continue
		}
		tokens := semanticEffectTokens(sentences[i].Tokens)
		count := legacyEffectCount(tokens, atoms)
		legacyEffects += count
		sentences[i].LegacyEffects = count > 0
		sentences[i].Targets = parseTargets(tokens, atoms)
		sentences[i].Effects = parseEffects(sentences[i], tokens, atoms)
		currentEffects += len(sentences[i].Effects)
		if len(tokens) > 0 && len(sentences[i].Effects) == 0 &&
			len(atoms.KeywordsWithin(tokens)) == 0 && count == 0 &&
			!effectWordsAt(tokens, 0, "activate", "only", "if") {
			unrecognizedSibling = true
		}
	}
	if currentEffects == 1 && unrecognizedSibling {
		for i := range sentences {
			for j := range sentences[i].Effects {
				sentences[i].Effects[j].Exact = false
				sentences[i].Effects[j].HasUnrecognizedSibling = true
			}
		}
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

func spanInsideActivationRestriction(span shared.Span, restrictions []ActivationRestriction) bool {
	for i := range restrictions {
		if spanCovers(restrictions[i].Span, span) || spanCovers(span, restrictions[i].Span) {
			return true
		}
	}
	return false
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

func parseEffects(sentence Sentence, tokens []shared.Token, atoms Atoms) []EffectSyntax {
	indices := effectIndices(tokens, atoms)
	requiresOrderedLowering := legacyEffectCount(tokens, atoms) > 1
	effects := make([]EffectSyntax, 0, len(indices))
	for effectIndex, tokenIndex := range indices {
		clauseEnd := resolvingClauseEnd(tokens, indices, effectIndex)
		ownershipStart := resolvingClauseStart(tokens, indices, effectIndex)
		ownership := tokens[ownershipStart:clauseEnd]
		clause := tokens[tokenIndex+1 : clauseEnd]
		clause, delayed := cutDelayedTiming(clause)
		power, toughness := parsePTChange(clause)
		counterKind, counterKnown := parseCounterPlacement(clause, atoms)
		span := shared.SpanOf(clause)
		ownershipSpan := shared.SpanOf(ownership)
		toZone := firstZone(atoms, span, ZoneRoleTo)
		if ambiguousZoneChoice(ownership, atoms, span) {
			toZone = zone.None
		}
		staticSubject := parseEffectStaticSubject(ownership, atoms)
		payment := parseEffectPayment(tokens)
		connection, connectionSpan := effectConnection(tokens, indices, effectIndex)
		optional, optionalSpan := effectOptional(tokens, tokenIndex)
		context := effectContextAt(tokens, tokenIndex, atoms)
		if effectIndex > 0 && !effectHasExplicitSubject(tokens, tokenIndex) &&
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
		kind := effectKindAt(tokens, tokenIndex)
		effects = append(effects, EffectSyntax{
			Kind:                    kind,
			Context:                 context,
			Connection:              connection,
			ConnectionSpan:          connectionSpan,
			Span:                    sentence.Span,
			VerbSpan:                tokens[tokenIndex].Span,
			ClauseSpan:              ownershipSpan,
			Text:                    sentence.Text,
			Tokens:                  append([]shared.Token(nil), ownership...),
			Duration:                parseEffectDuration(durationTokens, atoms),
			DelayedTiming:           delayed,
			Selection:               parseSelection(clause, atoms),
			Amount:                  parseEffectAmount(kind, clause, atoms),
			PowerDelta:              power,
			ToughnessDelta:          toughness,
			StaticSubject:           staticSubject,
			CounterKind:             counterKind,
			CounterKnown:            counterKnown,
			FromZone:                firstZone(atoms, span, ZoneRoleFrom),
			ToZone:                  toZone,
			Destination:             parseEffectDestination(ownership),
			EntersTapped:            effectWordsAtAny(ownership, "battlefield", "tapped"),
			EntersTappedSelf:        entersTappedSelfSyntax(kind, clause),
			EntersWithCounters:      entersWithCountersSyntax(kind, clause),
			UnderYourControl:        effectContainsWords(shared.NormalizedWords(ownership), "under", "your", "control"),
			CastAsAdventure:         effectContainsWords(shared.NormalizedWords(clause), "as", "an", "adventure"),
			Negated:                 effectIsNegated(tokens, tokenIndex),
			Optional:                optional,
			OptionalSpan:            optionalSpan,
			Symbol:                  firstEffectSymbol(clause),
			Mana:                    parseEffectMana(kind, clause, nextConnection != EffectConnectionNone),
			Replacement:             parseEffectReplacement(ownership, atoms),
			References:              referencesInSpan(atoms, ownershipSpan),
			SubjectReferences:       referencesInSpan(atoms, shared.SpanOf(tokens[ownershipStart:tokenIndex])),
			Targets:                 targetsInSpan(sentence.Targets, ownershipSpan),
			SubjectTargets:          targetsInSpan(sentence.Targets, shared.SpanOf(tokens[ownershipStart:tokenIndex])),
			Payment:                 payment,
			RequiresOrderedLowering: requiresOrderedLowering,
		})
	}

	for i := range effects {
		effects[i].Exact = exactEffectSyntax(&effects[i])
		effects[i].Mana.LegacyBodyExact = legacyExactManaBody(&effects[i], sentence)
		if effects[i].Kind == EffectSearch {
			effects[i].UnsupportedDetail = searchUnsupportedDetail(&effects[i])
		}
	}
	return effects
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
	return effect.Mana.AnyColor || len(effect.Mana.Symbols) != 0
}

func legacyEffectCount(tokens []shared.Token, atoms Atoms) int {
	count := 0
	for i := range tokens {
		if legacyEffectKindAt(tokens, i) != EffectUnknown &&
			!atoms.SelfNameAt(tokens[i].Span) &&
			!effectWithinCondition(tokens, i) {
			count++
		}
	}
	return count
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

func legacyEffectKindAt(tokens []shared.Token, index int) EffectKind {
	if equalWord(tokens[index], "look") {
		return EffectManifestDread
	}
	kind := effectWordKind(tokens[index])
	switch {
	case kind == EffectGrantKeyword && index >= 2 &&
		(equalWord(tokens[index-2], "opponent") || equalWord(tokens[index-2], "opponents")) &&
		equalWord(tokens[index-1], "you"):
		return EffectUnknown
	case kind == EffectEnterTapped && index+1 < len(tokens) && equalWord(tokens[index+1], "prepared"):
		return EffectEnterPrepared
	case kind == EffectCast && index > 0 && (equalWord(tokens[index-1], "was") || equalWord(tokens[index-1], "were")):
		return EffectUnknown
	case kind == EffectCounter && !counterVerbAt(tokens, index):
		return EffectUnknown
	case kind == EffectGain && index+1 < len(tokens) && equalWord(tokens[index+1], "control"):
		return EffectGainControl
	case kind == EffectDouble && index+1 < len(tokens) && equalWord(tokens[index+1], "strike"):
		return EffectUnknown
	case kind == EffectGrantKeyword && priorPTChange(tokens, index):
		return EffectUnknown
	default:
		return kind
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

// entersTappedSelfSyntax recognizes a self enters-tapped instruction such as
// "This land enters tapped." or "Nyx Lotus enters tapped." The enters verb is
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

func exactEffectSyntax(effect *EffectSyntax) bool {
	switch effect.Kind {
	case EffectDealDamage:
		return exactDamageEffectSyntax(effect)
	case EffectCounter:
		return exactCounterEffectSyntax(effect)
	case EffectDiscard:
		return exactCardCountEffectSyntax(effect, "Discard", "discards", false)
	case EffectDestroy:
		return exactDirectTargetEffectSyntax(effect, "Destroy") ||
			exactMassEffectSyntax(effect, "Destroy all ") ||
			exactDirectPronounEffectSyntax(effect, "Destroy it.")
	case EffectDraw:
		return exactCardCountEffectSyntax(effect, "Draw", "draws", true)
	case EffectEnterTapped:
		return exactLegacyFixedAmountSyntax(effect)
	case EffectExile:
		return exactDirectTargetEffectSyntax(effect, "Exile") ||
			exactMassEffectSyntax(effect, "Exile all ") ||
			exactDirectPronounEffectSyntax(effect, "Exile it.")
	case EffectFight:
		return exactFightEffectSyntax(effect)
	case EffectExplore:
		return exactDirectPronounEffectSyntax(effect, "It explores.")
	case EffectGain:
		return exactLifeEffectSyntax(effect, "gain", "gains") ||
			exactTemporaryKeywordEffectSyntax(effect)
	case EffectGainControl:
		return exactGainControlEffectSyntax(effect)
	case EffectInvestigate:
		return exactStandaloneActionEffectSyntax(effect, "Investigate")
	case EffectLose:
		return exactLifeEffectSyntax(effect, "lose", "loses")
	case EffectManifest:
		return strings.EqualFold(exactEffectClauseText(effect), "Manifest the top card of your library.")
	case EffectManifestDread:
		return strings.EqualFold(exactEffectClauseText(effect), "Manifest dread.")
	case EffectMill:
		return exactCardCountEffectSyntax(effect, "Mill", "mills", true)
	case EffectModifyPT:
		return exactModifyPTEffectSyntax(effect)
	case EffectPut:
		return exactCounterPlacementEffectSyntax(effect) || exactGraveyardPutEffectSyntax(effect)
	case EffectProliferate:
		return exactStandaloneActionEffectSyntax(effect, "Proliferate")
	case EffectRegenerate:
		return exactDirectTargetEffectSyntax(effect, "Regenerate")
	case EffectReturn:
		return exactBounceEffectSyntax(effect) ||
			exactGraveyardReturnEffectSyntax(effect) ||
			exactDirectPronounEffectSyntax(effect, "Return it to its owner's hand.")
	case EffectSacrifice:
		return exactDirectPronounEffectSyntax(effect, "Sacrifice it.") ||
			exactSacrificeChoiceEffectSyntax(effect)
	case EffectSearch:
		return exactSearchEffectSyntax(effect)
	case EffectScry:
		return exactControllerAmountEffectSyntax(effect, "Scry")
	case EffectSurveil:
		return exactControllerAmountEffectSyntax(effect, "Surveil")
	case EffectTap:
		return exactDirectTargetEffectSyntax(effect, "Tap") || exactDirectReferenceEffectSyntax(effect, "Tap")
	case EffectUntap:
		return exactDirectTargetEffectSyntax(effect, "Untap") ||
			exactDirectReferenceEffectSyntax(effect, "Untap") ||
			exactNegatedNextUntapStepSyntax(effect)
	case EffectTransform:
		return exactDirectTargetEffectSyntax(effect, "Transform")
	default:
		return false
	}
}

func exactSacrificeChoiceEffectSyntax(effect *EffectSyntax) bool {
	if !effect.Amount.Known || effect.Amount.Value < 1 || effect.Amount.Value > 2 {
		return false
	}
	subject := ""
	switch effect.Context {
	case EffectContextEachOpponent:
		subject = "Each opponent"
	case EffectContextEachPlayer:
		subject = "Each player"
	case EffectContextTarget:
		if len(effect.Targets) != 1 || !effect.Targets[0].Exact {
			return false
		}
		subject = titleFirstEffectText(effect.Targets[0].Text)
	default:
		return false
	}
	noun := ""
	switch effect.Selection.Kind {
	case SelectionArtifact:
		noun = "artifact"
	case SelectionCreature:
		noun = "creature"
	case SelectionEnchantment:
		noun = "enchantment"
	case SelectionLand:
		noun = "land"
	case SelectionPermanent:
		noun = "permanent"
	default:
		return false
	}
	if effect.Amount.Value > 1 {
		noun += "s"
	}
	text := exactEffectClauseText(effect)
	prefix := fmt.Sprintf("%s sacrifices %s %s", subject, effectAmountSourceText(effect), noun)
	return strings.EqualFold(text, prefix+".") ||
		strings.EqualFold(text, prefix+" of their choice.")
}

func exactSearchEffectSyntax(effect *EffectSyntax) bool {
	return searchUnsupportedDetail(effect) == ""
}

func searchUnsupportedDetail(effect *EffectSyntax) string {
	text := effect.Text
	if !strings.HasPrefix(text, "Search your library for ") || !strings.HasSuffix(text, ", then shuffle.") {
		return `the executable source backend supports only searches of your library ending with "then shuffle"`
	}
	rest := strings.TrimPrefix(text, "Search your library for ")
	rest = strings.TrimPrefix(rest, "a ")
	rest = strings.TrimPrefix(rest, "an ")
	filter := ""
	if !strings.HasPrefix(rest, "card,") {
		var ok bool
		filter, _, ok = strings.Cut(rest, " card,")
		if !ok {
			return "the executable source backend supports only exact singular-card search wording"
		}
	}
	switch filter {
	case "", "basic land", "land", "creature", "artifact", "enchantment",
		"Forest", "Plains", "Island", "Swamp", "Mountain",
		"Forest or Plains", "Plains, Island, Swamp, or Mountain":
	default:
		return fmt.Sprintf("unsupported library-search filter %q", filter)
	}
	for _, suffix := range []string{
		", put it into your hand, then shuffle.",
		", put that card into your hand, then shuffle.",
		", reveal it, put it into your hand, then shuffle.",
		", reveal that card, put it into your hand, then shuffle.",
		", put it onto the battlefield, then shuffle.",
		", put that card onto the battlefield, then shuffle.",
		", put it onto the battlefield tapped, then shuffle.",
		", put that card onto the battlefield tapped, then shuffle.",
	} {
		if strings.HasSuffix(text, suffix) {
			return ""
		}
	}
	return "the executable source backend supports only exact hand or battlefield search destinations"
}

func exactLifeEffectSyntax(effect *EffectSyntax, controllerVerb, subjectVerb string) bool {
	if effect.Optional {
		return false
	}

	var prefixes []string
	switch effect.Context {
	case EffectContextController:
		prefixes = []string{"You " + controllerVerb}
	case EffectContextEachOpponent:
		prefixes = []string{"Each opponent " + subjectVerb}
	case EffectContextEachPlayer:
		prefixes = []string{"Each player " + subjectVerb}
	case EffectContextTarget, EffectContextPriorSubject:
		if len(effect.Targets) == 1 && effect.Targets[0].Exact {
			prefixes = []string{titleFirstEffectText(effect.Targets[0].Text) + " " + subjectVerb}
		}
	case EffectContextEventPlayer, EffectContextReferencedPlayer:
		prefixes = []string{"They " + controllerVerb, "That player " + subjectVerb}
	default:
	}
	text := exactEffectClauseText(effect)
	for _, prefix := range prefixes {
		if exactAmountEffectText(text, prefix, "life", effect.Amount, effectAmountSourceText(effect)) {
			return true
		}
	}
	return false
}

func exactTemporaryKeywordEffectSyntax(effect *EffectSyntax) bool {
	if effect.Duration != EffectDurationUntilEndOfTurn {
		return false
	}
	text := strings.ToLower(exactEffectClauseText(effect))
	if effect.Context == EffectContextPriorSubject {
		middle, ok := strings.CutPrefix(text, "gains ")
		if !ok {
			return false
		}
		middle, ok = strings.CutSuffix(middle, " until end of turn.")
		return ok && exactTemporaryKeywordList(middle)
	}
	if effect.Context == EffectContextReferencedObject {
		subject, ok := exactObjectReferenceText(effect.SubjectReferences)
		if !ok {
			return false
		}
		middle, ok := strings.CutPrefix(text, strings.ToLower(subject)+" gains ")
		if !ok {
			return false
		}
		middle, ok = strings.CutSuffix(middle, " until end of turn.")
		return ok && exactTemporaryKeywordList(middle)
	}
	if len(effect.Targets) != 1 || !effect.Targets[0].Exact {
		return false
	}
	if prefix, suffix, ok := strings.Cut(text, " and gains "); ok &&
		strings.HasPrefix(prefix, strings.ToLower(effect.Targets[0].Text)+" gets ") {
		middle, suffixOK := strings.CutSuffix(suffix, " until end of turn.")
		return suffixOK && exactTemporaryKeywordList(middle)
	}
	prefix := strings.ToLower(effect.Targets[0].Text) + " gains "
	middle, ok := strings.CutPrefix(text, prefix)
	if !ok {
		return false
	}
	middle, ok = strings.CutSuffix(middle, " until end of turn.")
	if !ok || middle == "" {
		return false
	}
	return exactTemporaryKeywordList(middle)
}

func exactTemporaryKeywordList(text string) bool {
	text = strings.ReplaceAll(strings.ToLower(text), ", and ", ", ")
	text = strings.ReplaceAll(text, " and ", ", ")
	for keyword := range strings.SplitSeq(text, ", ") {
		switch keyword {
		case "deathtouch", "double strike", "first strike", "flying", "haste",
			"hexproof", "indestructible", "lifelink", "menace", "reach", "shroud", "trample", "vigilance":
		default:
			return false
		}
	}
	return true
}

func exactCardCountEffectSyntax(effect *EffectSyntax, controllerVerb, subjectVerb string, allowDynamic bool) bool {
	if effect.Amount.Known && !exactLegacyFixedAmountSyntax(effect) {
		return false
	}
	if effect.Kind == EffectMill && effect.Amount.DynamicKind == EffectDynamicAmountControllerLife {
		return false
	}
	var prefixes []string
	switch effect.Context {
	case EffectContextController:
		prefixes = []string{controllerVerb}
	case EffectContextTarget:
		if len(effect.Targets) == 1 && effect.Targets[0].Exact &&
			effect.Targets[0].Selection.Kind == SelectionPlayer {
			prefixes = []string{titleFirstEffectText(effect.Targets[0].Text) + " " + subjectVerb}
		}
	case EffectContextPriorSubject:
		if len(effect.Targets) == 1 && effect.Targets[0].Exact &&
			effect.Targets[0].Selection.Kind == SelectionPlayer {
			prefixes = []string{titleFirstEffectText(effect.Targets[0].Text) + " " + subjectVerb}
		} else {
			prefixes = []string{controllerVerb, subjectVerb}
		}
	case EffectContextEventPlayer, EffectContextReferencedPlayer:
		prefixes = []string{"They " + strings.TrimSuffix(subjectVerb, "s"), "That player " + subjectVerb}
	default:
	}
	text := exactEffectClauseText(effect)
	for _, prefix := range prefixes {
		if exactCountedNounEffectText(text, prefix, "card", "cards", effect.Amount, effectAmountSourceText(effect), allowDynamic) {
			return true
		}
	}
	return false
}

func exactGainControlEffectSyntax(effect *EffectSyntax) bool {
	if effect.Negated {
		return false
	}
	object := ""
	switch {
	case effect.Context == EffectContextController &&
		len(effect.Targets) == 1 &&
		effect.Targets[0].Exact:
		object = effect.Targets[0].Text
	case (effect.Context == EffectContextController || effect.Context == EffectContextPriorSubject) &&
		len(effect.Targets) == 0 &&
		len(effect.References) == 1 &&
		effect.References[0].Kind == ReferencePronoun &&
		effect.References[0].Pronoun == PronounIt:
		object = "it"
	default:
		return false
	}
	prefix := "Gain control of " + object
	text := exactEffectClauseText(effect)
	switch effect.Duration {
	case EffectDurationNone:
		return strings.EqualFold(text, prefix+".")
	case EffectDurationUntilEndOfTurn:
		return strings.EqualFold(text, prefix+" until end of turn.")
	case EffectDurationWhileYouControlSource:
		return exactGainControlControlledSourceDuration(text, prefix)
	case EffectDurationWhileSourceOnBattlefield:
		return exactGainControlBattlefieldSourceDuration(text, prefix)
	default:
		return false
	}
}

func exactGainControlControlledSourceDuration(text, prefix string) bool {
	const namedSourcePrefix = " for as long as you control "
	if suffix, ok := strings.CutPrefix(strings.ToLower(text), strings.ToLower(prefix+namedSourcePrefix)); ok {
		return suffix != "." && strings.HasSuffix(suffix, ".")
	}
	for _, noun := range []string{"artifact", "creature", "enchantment", "land", "permanent", "planeswalker"} {
		if strings.EqualFold(text, prefix+namedSourcePrefix+"this "+noun+".") {
			return true
		}
	}
	return false
}

func exactGainControlBattlefieldSourceDuration(text, prefix string) bool {
	for _, noun := range []string{"artifact", "creature", "enchantment", "land", "permanent", "planeswalker"} {
		for _, verb := range []string{"is", "remains"} {
			if strings.EqualFold(text, prefix+" as long as this "+noun+" "+verb+" on the battlefield.") {
				return true
			}
		}
	}
	return false
}

func exactControllerAmountEffectSyntax(effect *EffectSyntax, verb string) bool {
	return effect.Context == EffectContextController &&
		effect.Amount.Known &&
		strings.EqualFold(
			exactEffectClauseText(effect),
			fmt.Sprintf("%s %s.", verb, effectAmountSourceText(effect)),
		)
}

func exactStandaloneActionEffectSyntax(effect *EffectSyntax, verb string) bool {
	if effect.Context != EffectContextController || !effect.Amount.Known {
		return false
	}
	text := exactEffectClauseText(effect)
	if effect.Amount.Value == 1 && strings.EqualFold(text, verb+".") {
		return true
	}
	amount := effectAmountSourceText(effect)
	return strings.EqualFold(text, fmt.Sprintf("%s %s.", verb, amount)) ||
		strings.EqualFold(text, fmt.Sprintf("%s %s times.", verb, amount))
}

func exactLegacyFixedAmountSyntax(effect *EffectSyntax) bool {
	if !effect.Amount.Known || effect.Amount.Value <= 4 {
		return true
	}
	for _, token := range effect.Tokens {
		if token.Span == effect.Amount.Span {
			return token.Kind == shared.Integer
		}
	}
	return false
}

func exactAmountEffectText(text, prefix, noun string, amount EffectAmountSyntax, amountText string) bool {
	switch amount.DynamicForm {
	case EffectDynamicAmountFormNone:
		return strings.EqualFold(text, fmt.Sprintf("%s %s %s.", prefix, amountText, noun))
	case EffectDynamicAmountFormEqual:
		return strings.EqualFold(text, fmt.Sprintf("%s %s %s.", prefix, noun, amount.Text))
	case EffectDynamicAmountFormForEach:
		return strings.EqualFold(text, fmt.Sprintf("%s %d %s %s.", prefix, amount.Multiplier, noun, amount.Text))
	case EffectDynamicAmountFormWhereX:
		return strings.EqualFold(text, fmt.Sprintf("%s X %s, %s.", prefix, noun, amount.Text))
	default:
		return false
	}
}

func exactCountedNounEffectText(
	text, prefix, singular, plural string,
	amount EffectAmountSyntax,
	amountText string,
	allowDynamic bool,
) bool {
	if amount.DynamicForm == EffectDynamicAmountFormNone {
		noun := plural
		if amount.Known && amount.Value == 1 {
			noun = singular
		}
		return strings.EqualFold(text, fmt.Sprintf("%s %s %s.", prefix, amountText, noun))
	}
	if !allowDynamic {
		return false
	}
	switch amount.DynamicForm {
	case EffectDynamicAmountFormEqual:
		return strings.EqualFold(text, fmt.Sprintf("%s %s %s.", prefix, plural, amount.Text))
	case EffectDynamicAmountFormForEach:
		noun := plural
		if amount.Multiplier == 1 {
			noun = singular
		}
		return strings.EqualFold(text, fmt.Sprintf("%s %d %s %s.", prefix, amount.Multiplier, noun, amount.Text)) ||
			(amount.Multiplier == 1 && strings.EqualFold(text, fmt.Sprintf("%s a %s %s.", prefix, noun, amount.Text)))
	case EffectDynamicAmountFormWhereX:
		return strings.EqualFold(text, fmt.Sprintf("%s X %s, %s.", prefix, plural, amount.Text))
	default:
		return false
	}
}

func exactModifyPTEffectSyntax(effect *EffectSyntax) bool {
	if effect.Optional || effect.Duration != EffectDurationUntilEndOfTurn {
		return false
	}
	if effect.StaticSubject.Kind != EffectStaticSubjectNone {
		return exactGroupModifyPTEffectSyntax(effect)
	}
	subject := ""
	switch effect.Context {
	case EffectContextTarget:
		if len(effect.Targets) != 1 || !effect.Targets[0].Exact {
			return false
		}
		subject = titleFirstEffectText(effect.Targets[0].Text)
	case EffectContextReferencedObject:
		if effect.Amount.DynamicKind != EffectDynamicAmountNone {
			return false
		}
		subject = "It"
	default:
		return false
	}
	power := signedEffectAmountText(effect.PowerDelta)
	toughness := signedEffectAmountText(effect.ToughnessDelta)
	text := exactEffectClauseText(effect)
	if effect.Amount.DynamicKind == EffectDynamicAmountNone {
		prefix := fmt.Sprintf("%s gets %s/%s", subject, power, toughness)
		return strings.EqualFold(text, prefix+" until end of turn.") ||
			strings.EqualFold(text, prefix+".") ||
			strings.HasPrefix(strings.ToLower(text), strings.ToLower(prefix+" and gains ")) &&
				strings.HasSuffix(strings.ToLower(text), " until end of turn.")
	}
	switch effect.Amount.DynamicForm {
	case EffectDynamicAmountFormForEach:
		return strings.EqualFold(text, fmt.Sprintf("%s gets %s/%s %s until end of turn.", subject, power, toughness, effect.Amount.Text)) ||
			strings.EqualFold(text, fmt.Sprintf("%s gets %s/%s until end of turn %s.", subject, power, toughness, effect.Amount.Text))
	case EffectDynamicAmountFormWhereX:
		return strings.EqualFold(text, fmt.Sprintf("%s gets +X/+X until end of turn, %s.", subject, effect.Amount.Text))
	default:
		return false
	}
}

func exactGroupModifyPTEffectSyntax(effect *EffectSyntax) bool {
	if effect.Amount.DynamicKind != EffectDynamicAmountNone {
		return false
	}
	var subject []shared.Token
	for i := range effect.Tokens {
		if spanCovers(effect.StaticSubject.Span, effect.Tokens[i].Span) {
			subject = append(subject, effect.Tokens[i])
		}
	}
	if len(subject) == 0 {
		return false
	}
	return strings.EqualFold(
		exactEffectClauseText(effect),
		fmt.Sprintf(
			"%s get %s/%s until end of turn.",
			joinedEffectText(subject),
			signedEffectAmountText(effect.PowerDelta),
			signedEffectAmountText(effect.ToughnessDelta),
		),
	)
}

func exactCounterPlacementEffectSyntax(effect *EffectSyntax) bool {
	if !effect.CounterKnown {
		return false
	}
	object := ""
	switch {
	case len(effect.Targets) == 1 && effect.Targets[0].Exact:
		object = effect.Targets[0].Text
	case len(effect.Targets) == 0:
		var ok bool
		object, ok = exactObjectReferenceText(effect.References)
		if !ok {
			return false
		}
	default:
		return false
	}
	noun := "counters"
	if effect.Amount.Known && effect.Amount.Value == 1 {
		noun = "counter"
	}
	text := exactEffectClauseText(effect)
	prefix := fmt.Sprintf("Put %s %s %s on %s", effectAmountSourceText(effect), effect.CounterKind.String(), noun, object)
	if strings.EqualFold(text, prefix+".") {
		return true
	}
	return effect.Amount.DynamicForm == EffectDynamicAmountFormWhereX &&
		strings.EqualFold(text, prefix+", "+effect.Amount.Text+".")
}

func effectAmountSourceText(effect *EffectSyntax) string {
	if effect.Amount.VariableX || effect.Amount.DynamicForm == EffectDynamicAmountFormWhereX {
		return "X"
	}
	for _, token := range effect.Tokens {
		if token.Span == effect.Amount.Span {
			return token.Text
		}
	}
	return effect.Amount.Text
}

func exactGraveyardReturnEffectSyntax(effect *EffectSyntax) bool {
	text := exactEffectClauseText(effect)
	if len(effect.Targets) == 0 {
		switch {
		case strings.EqualFold(text, "Return this card from your graveyard to your hand."),
			strings.EqualFold(text, "Return this card from your graveyard to the battlefield."),
			strings.EqualFold(text, "Return this card from your graveyard to the battlefield tapped."):
			return true
		case strings.HasPrefix(strings.ToLower(text), "return this card from your graveyard to the battlefield with "),
			strings.HasPrefix(strings.ToLower(text), "return this card from your graveyard to the battlefield tapped with "):
			return effect.CounterKnown && effect.CounterKind == counter.PlusOnePlusOne
		default:
			return false
		}
	}
	if len(effect.Targets) != 1 || !exactGraveyardCardTargetSyntax(effect.Targets[0]) {
		return false
	}
	prefix := "Return " + effect.Targets[0].Text
	for _, suffix := range []string{
		" to your hand.",
		" to the battlefield.",
		" to the battlefield tapped.",
		" to the battlefield under your control.",
		" to the battlefield tapped under your control.",
		" on top of your library.",
		" on the top of your library.",
		" on bottom of your library.",
		" on the bottom of your library.",
	} {
		if strings.EqualFold(text, prefix+suffix) {
			return true
		}
	}
	return false
}

func exactGraveyardPutEffectSyntax(effect *EffectSyntax) bool {
	if len(effect.Targets) != 1 || !exactGraveyardCardTargetSyntax(effect.Targets[0]) {
		return false
	}
	text := exactEffectClauseText(effect)
	prefix := "Put " + effect.Targets[0].Text
	for _, suffix := range []string{
		" onto the battlefield.",
		" onto the battlefield under your control.",
		" on top of your library.",
		" on the top of your library.",
		" on bottom of your library.",
		" on the bottom of your library.",
	} {
		if strings.EqualFold(text, prefix+suffix) {
			return true
		}
	}
	return false
}

func exactGraveyardCardTargetSyntax(target TargetSyntax) bool {
	if target.Selection.Zone != zone.Graveyard ||
		target.Selection.Other {
		return false
	}
	cardinalityOne := target.Cardinality == (TargetCardinalitySyntax{Min: 1, Max: 1}) ||
		target.Cardinality == (TargetCardinalitySyntax{Min: 0, Max: 1})
	text := strings.ToLower(target.Text)
	text = strings.TrimPrefix(text, "up to one ")
	text = strings.TrimPrefix(text, "up to two ")
	text = strings.TrimPrefix(text, "another ")
	if !strings.HasPrefix(text, "target ") {
		return false
	}
	for _, noun := range []string{
		"card", "creature card", "artifact card", "enchantment card", "land card",
		"planeswalker card", "instant or sorcery card",
	} {
		for _, owner := range []string{"your graveyard", "a graveyard", "an opponent's graveyard"} {
			if cardinalityOne && (text == "target "+noun+" from "+owner ||
				exactGraveyardManaValueTarget(text, noun, owner)) {
				return true
			}
		}
	}
	if target.Cardinality == (TargetCardinalitySyntax{Min: 0, Max: 2}) &&
		text == "target cards with cycling from your graveyard" {
		return true
	}
	if cardinalityOne && len(target.Selection.SubtypesAny) == 1 {
		subtype := strings.ToLower(string(target.Selection.SubtypesAny[0]))
		for _, owner := range []string{"your graveyard", "a graveyard", "an opponent's graveyard"} {
			if text == "target "+subtype+" card from "+owner {
				return true
			}
		}
	}
	return false
}

func exactGraveyardManaValueTarget(text, noun, owner string) bool {
	prefix := "target " + noun + " with mana value "
	suffix := " or less from " + owner
	value, ok := strings.CutSuffix(strings.TrimPrefix(text, prefix), suffix)
	if !ok || !strings.HasPrefix(text, prefix) {
		return false
	}
	_, err := strconv.Atoi(value)
	return err == nil
}

func titleFirstEffectText(text string) string {
	if text == "" {
		return ""
	}
	return strings.ToUpper(text[:1]) + text[1:]
}

func signedEffectAmountText(amount SignedAmountSyntax) string {
	if amount.Negative {
		return fmt.Sprintf("-%d", amount.Value)
	}
	return fmt.Sprintf("+%d", amount.Value)
}

func exactCounterEffectSyntax(effect *EffectSyntax) bool {
	if exactDirectTargetEffectSyntax(effect, "Counter") {
		return true
	}
	return len(effect.Targets) == 1 &&
		effect.Targets[0].Exact &&
		effect.Payment.Payer == EffectPaymentPayerTargetController &&
		len(effect.Payment.ManaCost) > 0 &&
		strings.EqualFold(
			exactEffectClauseText(effect),
			"Counter "+effect.Targets[0].Text+" unless its controller pays "+effect.Payment.ManaCost.String()+".",
		)
}

func exactDirectTargetEffectSyntax(effect *EffectSyntax, verb string) bool {
	return len(effect.Targets) == 1 &&
		effect.Targets[0].Exact &&
		strings.EqualFold(exactEffectClauseText(effect), verb+" "+effect.Targets[0].Text+".")
}

func exactNegatedNextUntapStepSyntax(effect *EffectSyntax) bool {
	if !effect.Negated || effect.Context != EffectContextUnknown ||
		len(effect.Targets) != 0 || len(effect.References) != 0 {
		return false
	}
	words := shared.NormalizedWords(effect.Tokens)
	verb := slices.Index(words, "untap")
	return verb == 4 &&
		slices.Equal(words[:verb], []string{"lands", "you", "control", "don't"}) &&
		slices.Equal(words[verb+1:], []string{"during", "your", "next", "untap", "step"})
}

func exactBounceEffectSyntax(effect *EffectSyntax) bool {
	return len(effect.Targets) == 1 &&
		effect.Targets[0].Exact &&
		strings.EqualFold(exactEffectClauseText(effect), "Return "+effect.Targets[0].Text+" to its owner's hand.")
}

func exactDirectPronounEffectSyntax(effect *EffectSyntax, exact string) bool {
	return len(effect.Targets) == 0 &&
		effect.Duration == EffectDurationNone &&
		strings.EqualFold(exactEffectClauseText(effect), exact)
}

func exactDirectReferenceEffectSyntax(effect *EffectSyntax, verb string) bool {
	if len(effect.Targets) != 0 || effect.Optional || effect.Duration != EffectDurationNone {
		return false
	}
	object, ok := exactObjectReferenceText(effect.References)
	return ok && strings.EqualFold(exactEffectClauseText(effect), verb+" "+object+".")
}

func exactObjectReferenceText(references []Reference) (string, bool) {
	if len(references) != 1 {
		return "", false
	}
	switch references[0].Kind {
	case ReferenceThatObject:
	case ReferencePronoun:
		if references[0].Pronoun != PronounIt {
			return "", false
		}
	default:
		return "", false
	}
	return joinedEffectText(references[0].Tokens), true
}

func exactFightEffectSyntax(effect *EffectSyntax) bool {
	if effect.Context == EffectContextPriorSubject &&
		len(effect.Targets) == 1 &&
		effect.Targets[0].Exact &&
		strings.EqualFold(exactEffectClauseText(effect), "fights "+effect.Targets[0].Text+".") {
		return true
	}
	return len(effect.Targets) == 2 &&
		effect.Targets[0].Exact &&
		effect.Targets[1].Exact &&
		strings.EqualFold(exactEffectClauseText(effect), effect.Targets[0].Text+" fights "+effect.Targets[1].Text+".")
}

func exactMassEffectSyntax(effect *EffectSyntax, prefix string) bool {
	text := exactEffectClauseText(effect)
	if !strings.HasPrefix(strings.ToLower(text), strings.ToLower(prefix)) || !strings.HasSuffix(text, ".") {
		return false
	}
	phrase := text[len(prefix) : len(text)-1]
	return exactMassGroupPhrase(phrase)
}

func exactMassGroupPhrase(phrase string) bool {
	if phrase == "" || strings.TrimSpace(phrase) != phrase {
		return false
	}
	phrase = strings.ToLower(phrase)
	hadControllerSuffix := false
	for _, suffix := range []string{" you don't control", " your opponents control", " you control"} {
		if remainder, ok := strings.CutSuffix(phrase, suffix); ok {
			phrase = remainder
			hadControllerSuffix = true
			break
		}
	}
	if exactMassNumericPhrase(phrase) {
		return true
	}
	if !hadControllerSuffix {
		if keyword, ok := strings.CutPrefix(phrase, "creatures with "); ok {
			return keyword != "" &&
				!strings.Contains(keyword, " ") &&
				exactTemporaryKeywordList(keyword)
		}
	}
	if exactMassBaseNoun(phrase) {
		return true
	}
	for _, prefix := range []string{
		"other ", "tapped ", "nonland ", "nonartifact ", "noncreature ", "nonenchantment ",
		"white ", "blue ", "black ", "red ", "green ", "nonwhite ", "nonblue ", "nonblack ", "nonred ", "nongreen ",
	} {
		if remainder, ok := strings.CutPrefix(phrase, prefix); ok {
			return exactMassBaseNoun(remainder)
		}
	}
	return false
}

func exactMassBaseNoun(phrase string) bool {
	switch phrase {
	case "creatures", "artifacts", "enchantments", "lands", "planeswalkers", "permanents",
		"creatures and lands", "creatures and planeswalkers", "artifacts and enchantments",
		"artifacts and creatures", "artifacts, creatures, and enchantments",
		"artifacts, creatures, and lands":
		return true
	default:
		return false
	}
}

func exactMassNumericPhrase(phrase string) bool {
	for _, qualifier := range []string{"mana value", "power", "toughness"} {
		comparison, ok := strings.CutPrefix(phrase, "creatures with "+qualifier+" ")
		if !ok {
			continue
		}
		parts := strings.Fields(comparison)
		switch {
		case len(parts) == 1:
			_, err := strconv.Atoi(parts[0])
			return err == nil
		case len(parts) == 3 && parts[0] == "equal" && parts[1] == "to":
			_, err := strconv.Atoi(parts[2])
			return err == nil
		case len(parts) == 3 && parts[1] == "or" && (parts[2] == "less" || parts[2] == "greater"):
			_, err := strconv.Atoi(parts[0])
			return err == nil
		}
	}
	return false
}

func exactEffectClauseText(effect *EffectSyntax) string {
	verb := slices.IndexFunc(effect.Tokens, func(token shared.Token) bool {
		return token.Span == effect.VerbSpan
	})
	if verb < 0 {
		return ""
	}
	start := effectSubjectStart(effect.Tokens, verb)
	if effect.Optional && effectWordsAt(effect.Tokens, start, "you", "may") && start+2 == verb {
		start = verb
	}
	text := joinedEffectText(effect.Tokens[start:])
	if len(effect.Tokens) > 0 && effect.Tokens[len(effect.Tokens)-1].Kind != shared.Period {
		text += "."
	}
	if effect.DelayedTiming != DelayedTimingNone {
		for _, suffix := range []string{
			" at the beginning of the next end step.",
			" at the beginning of the next turn's upkeep.",
		} {
			if prefix, ok := strings.CutSuffix(text, suffix); ok {
				return prefix + "."
			}
		}
	}
	return text
}

func exactDamageEffectSyntax(effect *EffectSyntax) bool {
	verb := slices.IndexFunc(effect.Tokens, func(token shared.Token) bool {
		return token.Span == effect.VerbSpan
	})
	if verb < 0 {
		return false
	}
	subjectStart := effectSubjectStart(effect.Tokens, verb)
	subjectTokens := effect.Tokens[subjectStart:verb]
	subject := ""
	if len(subjectTokens) == 0 {
		if effect.Context != EffectContextPriorSubject {
			return false
		}
	} else {
		subject = joinedEffectText(subjectTokens)
		subjectSpan := shared.SpanOf(subjectTokens)
		exactSubject := false
		for _, reference := range effect.SubjectReferences {
			if !spanCovers(subjectSpan, reference.Span) {
				continue
			}
			exactSubject = reference.Kind == ReferenceSelfName ||
				reference.Kind == ReferencePronoun && reference.Pronoun == PronounIt
		}
		if !exactSubject {
			return false
		}
	}
	verbText := effect.Tokens[verb].Text
	if !equalWord(effect.Tokens[verb], "deal") && !equalWord(effect.Tokens[verb], "deals") {
		return false
	}
	prefix := verbText
	if subject != "" {
		prefix = subject + " " + verbText
	}
	text := joinedEffectText(effect.Tokens[subjectStart:])
	if len(effect.Tokens) > 0 && effect.Tokens[len(effect.Tokens)-1].Kind != shared.Period {
		text += "."
	}
	if len(effect.Targets) == 0 {
		if !effect.Amount.Known {
			return false
		}
		recipient := ""
		switch {
		case effect.Selection.Kind == SelectionOpponent && !effect.Selection.Other:
			recipient = "each opponent"
		case effect.Selection.Kind == SelectionPlayer && !effect.Selection.Other:
			recipient = "each player"
		case effect.Selection.Kind == SelectionCreature && !effect.Selection.Other:
			recipient = "each creature"
		case effect.Selection.Kind == SelectionCreature && effect.Selection.Other:
			recipient = "each other creature"
		default:
			return false
		}
		return text == fmt.Sprintf("%s %d damage to %s.", prefix, effect.Amount.Value, recipient)
	}
	if len(effect.Targets) != 1 || !effect.Targets[0].Exact {
		return false
	}
	target := effect.Targets[0].Text
	switch effect.Amount.DynamicForm {
	case EffectDynamicAmountFormNone:
		amount := "X"
		if effect.Amount.Known {
			amount = strconv.Itoa(effect.Amount.Value)
		} else if !effect.Amount.VariableX {
			return false
		}
		return text == fmt.Sprintf("%s %s damage to %s.", prefix, amount, target)
	case EffectDynamicAmountFormEqual:
		return text == fmt.Sprintf("%s damage %s to %s.", prefix, effect.Amount.Text, target)
	case EffectDynamicAmountFormForEach:
		return text == fmt.Sprintf("%s %d damage %s to %s.", prefix, effect.Amount.Multiplier, effect.Amount.Text, target)
	case EffectDynamicAmountFormWhereX:
		return text == fmt.Sprintf("%s X damage to %s, %s.", prefix, target, effect.Amount.Text)
	default:
		return false
	}
}

func durationScopesAcrossAnd(current, next EffectKind) bool {
	return temporaryModifierEffect(current) && temporaryModifierEffect(next)
}

func temporaryModifierEffect(kind EffectKind) bool {
	switch kind {
	case EffectModifyPT, EffectGain, EffectGrantKeyword:
		return true
	default:
		return false
	}
}

func targetsInSpan(targets []TargetSyntax, span shared.Span) []TargetSyntax {
	var result []TargetSyntax
	for _, target := range targets {
		if target.Span.Start.Offset >= span.Start.Offset && target.Span.End.Offset <= span.End.Offset {
			result = append(result, target)
		}
	}
	return result
}

func resolvingClauseStart(tokens []shared.Token, indices []int, effectIndex int) int {
	if effectIndex == 0 {
		return 0
	}
	for i := indices[effectIndex] - 1; i > indices[effectIndex-1]; i-- {
		if tokens[i].Kind == shared.Comma || tokens[i].Kind == shared.Semicolon ||
			equalWord(tokens[i], "then") || equalWord(tokens[i], "and") {
			return i + 1
		}
	}
	return 0
}

func parseEffectReplacement(tokens []shared.Token, atoms Atoms) EffectReplacementSyntax {
	if len(tokens) < 2 ||
		!equalWord(tokens[len(tokens)-2], "instead") ||
		tokens[len(tokens)-1].Kind != shared.Period {
		return EffectReplacementSyntax{}
	}
	replacement := EffectReplacementSyntax{
		Kind: EffectReplacementInstead,
		Span: tokens[len(tokens)-2].Span,
	}
	if replacementHasUnsupportedSelectionModifier(tokens, atoms) {
		return replacement
	}
	twiceMany := effectHasTokenWords(tokens, "twice", "that", "many")
	thatMuchPlus := effectHasTokenWords(tokens, "that", "much", "damage", "plus")
	doubleThat := effectHasTokenWords(tokens, "double", "that", "damage") ||
		effectHasTokenWords(tokens, "twice", "that", "damage")
	if boolCount(twiceMany, thatMuchPlus, doubleThat) != 1 {
		return replacement
	}
	switch {
	case twiceMany:
		replacement.Kind = EffectReplacementTwiceThatMany
	case thatMuchPlus:
		for i := range tokens {
			if !equalWord(tokens[i], "plus") || i+1 >= len(tokens) {
				continue
			}
			if amount, ok := effectNumber(tokens[i+1], atoms); ok {
				replacement.Kind = EffectReplacementThatMuchPlus
				replacement.Amount = amount
			}
			break
		}
	case doubleThat:
		replacement.Kind = EffectReplacementDoubleThat
	default:
	}
	replacement.EachCounterKind = effectHasTokenWords(tokens, "each", "of", "those", "kinds", "of", "counters")
	return replacement
}

func replacementHasUnsupportedSelectionModifier(tokens []shared.Token, atoms Atoms) bool {
	selection := parseSelection(tokens, atoms)
	return selection.Controller != SelectionControllerAny ||
		selection.Another || selection.Other || selection.Attacking || selection.Blocking ||
		selection.Tapped || selection.Untapped || selection.Keyword != KeywordUnknown ||
		selection.Zone != zone.None ||
		selection.MatchManaValue || selection.MatchPower || selection.MatchToughness ||
		len(selection.ExcludedTypes) != 0 || len(selection.Supertypes) != 0 ||
		len(selection.ColorsAny) != 0 || len(selection.ExcludedColors) != 0 ||
		len(selection.SubtypesAny) != 0
}

func boolCount(values ...bool) int {
	count := 0
	for _, value := range values {
		if value {
			count++
		}
	}
	return count
}

func effectHasTokenWords(tokens []shared.Token, words ...string) bool {
	for i := range tokens {
		if effectWordsAt(tokens, i, words...) {
			return true
		}
	}
	return false
}

func parseEffectMana(kind EffectKind, tokens []shared.Token, connected bool) EffectManaSyntax {
	if kind != EffectAddMana || len(tokens) == 0 {
		return EffectManaSyntax{}
	}
	body := tokens
	if tokens[len(tokens)-1].Kind == shared.Period {
		body = tokens[:len(tokens)-1]
	} else if !connected {
		return EffectManaSyntax{}
	}
	if len(body) == 5 && effectWordsAt(body, 0, "one", "mana", "of", "any", "color") {
		return EffectManaSyntax{Span: shared.SpanOf(body), AnyColor: true}
	}
	var symbols []string
	choice := false
	expectSymbol := true
	for i := 0; i < len(body); i++ {
		token := body[i]
		if expectSymbol {
			if token.Kind != shared.Symbol {
				return EffectManaSyntax{}
			}
			symbols = append(symbols, token.Text)
			expectSymbol = false
			continue
		}
		switch {
		case token.Kind == shared.Symbol:
			if choice {
				return EffectManaSyntax{}
			}
			symbols = append(symbols, token.Text)
		case token.Kind == shared.Comma:
			if len(symbols) != 1 && !choice {
				return EffectManaSyntax{}
			}
			choice = true
			expectSymbol = true
			if i+1 < len(body) && equalWord(body[i+1], "or") {
				i++
			}
		case equalWord(token, "or"):
			if len(symbols) != 1 && !choice {
				return EffectManaSyntax{}
			}
			choice = true
			expectSymbol = true
		default:
			return EffectManaSyntax{}
		}
	}
	if len(symbols) == 0 || expectSymbol || choice && len(symbols) < 2 {
		return EffectManaSyntax{}
	}
	return EffectManaSyntax{Span: shared.SpanOf(body), Symbols: symbols, Choice: choice}
}

func effectConnection(tokens []shared.Token, indices []int, effectIndex int) (EffectConnectionKind, shared.Span) {
	if effectIndex == 0 {
		return EffectConnectionNone, shared.Span{}
	}
	for i := indices[effectIndex] - 1; i > indices[effectIndex-1]; i-- {
		switch {
		case equalWord(tokens[i], "then"):
			return EffectConnectionThen, tokens[i].Span
		case equalWord(tokens[i], "and"):
			return EffectConnectionAnd, tokens[i].Span
		}
	}
	return EffectConnectionNone, shared.Span{}
}

func effectOptional(tokens []shared.Token, index int) (bool, shared.Span) {
	start := max(0, index-3)
	for i, token := range tokens[start:index] {
		if equalWord(token, "may") {
			span := token.Span
			tokenIndex := start + i
			if tokenIndex > 0 && equalWord(tokens[tokenIndex-1], "you") {
				span.Start = tokens[tokenIndex-1].Span.Start
			}
			return true, span
		}
	}
	return false, shared.Span{}
}

func parseEffectDestination(tokens []shared.Token) EffectDestinationPosition {
	words := shared.NormalizedWords(tokens)
	switch {
	case effectContainsWords(words, "on", "top", "of", "your", "library") ||
		effectContainsWords(words, "on", "the", "top", "of", "your", "library"):
		return EffectDestinationTop
	case effectContainsWords(words, "on", "bottom", "of", "your", "library") ||
		effectContainsWords(words, "on", "the", "bottom", "of", "your", "library"):
		return EffectDestinationBottom
	default:
		return EffectDestinationUnspecified
	}
}

func effectWordsAtAny(tokens []shared.Token, first, second string) bool {
	for i := range tokens {
		if equalWord(tokens[i], first) {
			for _, token := range tokens[i+1:] {
				if equalWord(token, second) {
					return true
				}
			}
		}
	}
	return false
}

func effectContextAt(tokens []shared.Token, index int, atoms Atoms) EffectContextKind {
	for _, token := range tokens {
		if equalWord(token, "random") || equalWord(token, "named") {
			return EffectContextUnknown
		}
	}
	start := effectSubjectStart(tokens, index)
	subject := tokens[start:index]
	for len(subject) > 0 && equalWord(subject[0], "then") {
		subject = subject[1:]
	}
	if len(subject) == 0 {
		return EffectContextController
	}
	words := shared.NormalizedWords(subject)
	if len(words) == 0 {
		return EffectContextUnknown
	}
	switch {
	case effectContainsWords(words, "each", "opponent") || effectContainsWords(words, "each", "opponents"):
		return EffectContextEachOpponent
	case effectContainsWords(words, "each", "player"):
		return EffectContextEachPlayer
	case effectContainsWords(words, "target"):
		return EffectContextTarget
	case len(words) >= 2 && words[len(words)-2] == "that" && words[len(words)-1] == "player":
		return EffectContextReferencedPlayer
	case words[len(words)-1] == "they":
		return EffectContextEventPlayer
	case words[len(words)-1] == "you" || len(words) >= 2 && words[len(words)-2] == "you" && words[len(words)-1] == "may":
		return EffectContextController
	}
	span := shared.SpanOf(subject)
	for _, reference := range atoms.References() {
		if !spanCovers(span, reference.Span) {
			continue
		}
		switch {
		case reference.Kind == ReferenceSelfName || reference.Kind == ReferenceThisObject:
			return EffectContextSource
		case reference.Kind == ReferencePronoun && reference.Pronoun == PronounThey:
			return EffectContextEventPlayer
		case reference.Kind == ReferenceThatObject:
			return EffectContextReferencedObject
		case reference.Kind == ReferenceThatPlayer:
			return EffectContextReferencedPlayer
		case reference.Kind == ReferencePronoun && reference.Pronoun == PronounIt:
			return EffectContextReferencedObject
		}
	}
	return EffectContextUnknown
}

func effectHasExplicitSubject(tokens []shared.Token, index int) bool {
	return effectSubjectStart(tokens, index) < index
}

func effectSubjectStart(tokens []shared.Token, index int) int {
	start := 0
	for i := range index {
		if tokens[i].Kind == shared.Comma || tokens[i].Kind == shared.Period || tokens[i].Kind == shared.Semicolon ||
			equalWord(tokens[i], "then") || equalWord(tokens[i], "and") {
			start = i + 1
		}
	}
	return start
}

func parseEffectPayment(tokens []shared.Token) EffectPaymentSyntax {
	for i := range tokens {
		if !effectWordsAt(tokens, i, "unless", "its", "controller", "pays") {
			continue
		}
		manaCost, end, ok := parseKeywordManaCost(tokens, i+4)
		if !ok || end >= len(tokens) || tokens[end].Kind != shared.Period || end != len(tokens)-1 {
			return EffectPaymentSyntax{}
		}
		return EffectPaymentSyntax{
			Span:     shared.SpanOf(tokens[i:end]),
			Payer:    EffectPaymentPayerTargetController,
			ManaCost: manaCost,
		}
	}
	return EffectPaymentSyntax{}
}

func effectIndices(tokens []shared.Token, atoms Atoms) []int {
	var result []int
	for i := range tokens {
		if effectKindAt(tokens, i) != EffectUnknown &&
			!atoms.SelfNameAt(tokens[i].Span) &&
			!effectNounAt(tokens, i) {
			result = append(result, i)
		}
	}
	return result
}

func effectNounAt(tokens []shared.Token, index int) bool {
	return index > 0 && index+1 < len(tokens) &&
		equalWord(tokens[index], "untap") &&
		equalWord(tokens[index-1], "next") &&
		equalWord(tokens[index+1], "step")
}

func resolvingClauseEnd(tokens []shared.Token, indices []int, effectIndex int) int {
	start := indices[effectIndex] + 1
	end := len(tokens)
	for _, next := range indices[effectIndex+1:] {
		for i := next - 1; i >= start; i-- {
			if tokens[i].Kind == shared.Comma || tokens[i].Kind == shared.Semicolon {
				end = i
				break
			}
			if equalWord(tokens[i], "then") || equalWord(tokens[i], "and") {
				end = i
				if i > start && tokens[i-1].Kind == shared.Comma {
					end--
				}
				break
			}
		}
		if end != len(tokens) {
			break
		}
	}
	for i := start; i < end; i++ {
		if equalWord(tokens[i], "if") || equalWord(tokens[i], "unless") ||
			(i+1 < end && equalWord(tokens[i], "only") && equalWord(tokens[i+1], "if")) {
			return i
		}
	}
	return end
}

func effectKindAt(tokens []shared.Token, index int) EffectKind {
	kind := effectWordKind(tokens[index])
	switch {
	case equalWord(tokens[index], "manifest"):
		switch {
		case effectWordsAt(tokens, index+1, "dread") && len(tokens) == index+3 && tokens[index+2].Kind == shared.Period:
			return EffectManifestDread
		case effectWordsAt(tokens, index+1, "the", "top", "card", "of", "your", "library") &&
			len(tokens) == index+8 && tokens[index+7].Kind == shared.Period:
			return EffectManifest
		default:
			return EffectManifest
		}
	case equalWord(tokens[index], "look"):
		if manifestDreadLookInstruction(tokens[index:]) {
			return EffectManifestDread
		}
		return EffectManifestDread
	case kind == EffectGrantKeyword && index >= 2 &&
		(equalWord(tokens[index-2], "opponent") || equalWord(tokens[index-2], "opponents")) &&
		equalWord(tokens[index-1], "you"):
		return EffectUnknown
	case kind == EffectEnterTapped && index+1 < len(tokens) && equalWord(tokens[index+1], "prepared"):
		return EffectEnterPrepared
	case kind == EffectCast && index > 0 && (equalWord(tokens[index-1], "was") || equalWord(tokens[index-1], "were")):
		return EffectUnknown
	case kind == EffectCounter && !counterVerbAt(tokens, index):
		return EffectUnknown
	case kind == EffectGain && index+1 < len(tokens) && equalWord(tokens[index+1], "control"):
		return EffectGainControl
	case kind == EffectDouble && index+1 < len(tokens) && equalWord(tokens[index+1], "strike"):
		return EffectUnknown
	case kind == EffectGrantKeyword && priorPTChange(tokens, index):
		return EffectUnknown
	default:
		return kind
	}
}

func effectWordKind(token shared.Token) EffectKind {
	if token.Kind != shared.Word {
		return EffectUnknown
	}
	switch strings.ToLower(token.Text) {
	case "add", "adds":
		return EffectAddMana
	case "attach", "attaches":
		return EffectAttach
	case "cast", "casts":
		return EffectCast
	case "counter", "counters":
		return EffectCounter
	case "create", "creates":
		return EffectCreate
	case "deal", "deals":
		return EffectDealDamage
	case "destroy", "destroys":
		return EffectDestroy
	case "discard", "discards":
		return EffectDiscard
	case "discover", "discovers":
		return EffectDiscover
	case "double", "doubles":
		return EffectDouble
	case "draw", "draws":
		return EffectDraw
	case "enters":
		return EffectEnterTapped
	case "exile", "exiles":
		return EffectExile
	case "fight", "fights":
		return EffectFight
	case "gain", "gains":
		return EffectGain
	case "has", "have":
		return EffectGrantKeyword
	case "investigate", "investigates":
		return EffectInvestigate
	case "explore", "explores":
		return EffectExplore
	case "lose", "loses":
		return EffectLose
	case "manifest":
		return EffectManifest
	case "mill", "mills":
		return EffectMill
	case "get", "gets":
		return EffectModifyPT
	case "put", "puts":
		return EffectPut
	case "proliferate", "proliferates":
		return EffectProliferate
	case "regenerate", "regenerates":
		return EffectRegenerate
	case "return", "returns":
		return EffectReturn
	case "reveal", "reveals":
		return EffectReveal
	case "sacrifice", "sacrifices":
		return EffectSacrifice
	case "scry", "scries":
		return EffectScry
	case "surveil", "surveils":
		return EffectSurveil
	case "search", "searches":
		return EffectSearch
	case "shuffle", "shuffles":
		return EffectShuffle
	case "tap", "taps":
		return EffectTap
	case "untap", "untaps":
		return EffectUntap
	case "transform", "transforms":
		return EffectTransform
	default:
		return EffectUnknown
	}
}

func manifestDreadLookInstruction(tokens []shared.Token) bool {
	return len(tokens) == 10 &&
		effectWordsAt(tokens, 0, "look", "at", "the", "top", "two", "cards", "of", "your", "library") &&
		tokens[9].Kind == shared.Period
}

func counterVerbAt(tokens []shared.Token, index int) bool {
	if index == 0 {
		return true
	}
	previous := tokens[index-1]
	if previous.Kind == shared.Comma || previous.Kind == shared.Period || previous.Kind == shared.Semicolon ||
		equalWord(previous, "then") || equalWord(previous, "may") || equalWord(previous, "can") {
		return true
	}
	return index+1 < len(tokens) &&
		(equalWord(tokens[index+1], "target") || equalWord(tokens[index+1], "it") || equalWord(tokens[index+1], "that"))
}

func priorPTChange(tokens []shared.Token, index int) bool {
	for i := range index {
		if equalWord(tokens[i], "get") || equalWord(tokens[i], "gets") {
			power, toughness := parsePTChange(tokens[i+1 : index])
			return power.Known && toughness.Known
		}
	}
	return false
}

func effectIsNegated(tokens []shared.Token, index int) bool {
	start := max(0, index-3)
	for i, token := range tokens[start:index] {
		if equalWord(token, "can't") || equalWord(token, "cannot") ||
			equalWord(token, "doesn't") || equalWord(token, "don't") || equalWord(token, "not") {
			for _, following := range tokens[start+i+1 : index] {
				if equalWord(following, "control") {
					return false
				}
			}
			return true
		}
	}
	return false
}

func parseEffectDuration(tokens []shared.Token, atoms Atoms) EffectDurationKind {
	words := shared.NormalizedWords(tokens)
	switch {
	case effectContainsWords(words, "until", "the", "end", "of", "your", "next", "turn"):
		return EffectDurationUntilYourNextTurn
	case effectContainsWords(words, "until", "end", "of", "turn"):
		return EffectDurationUntilEndOfTurn
	case effectContainsWords(words, "until", "your", "next", "turn"):
		return EffectDurationUntilYourNextTurn
	case effectContainsWords(words, "this", "combat"):
		return EffectDurationThisCombat
	case effectContainsWords(words, "this", "turn"):
		return EffectDurationThisTurn
	case effectContainsWords(words, "as", "long", "as", "this") &&
		(effectContainsWords(words, "remains", "on", "the", "battlefield") ||
			effectContainsWords(words, "is", "on", "the", "battlefield")):
		return EffectDurationWhileSourceOnBattlefield
	case effectContainsWords(words, "for", "as", "long", "as", "you", "control", "this"):
		return EffectDurationWhileYouControlSource
	}
	for i := 0; i+6 < len(tokens); i++ {
		if !effectWordsAt(tokens, i, "for", "as", "long", "as", "you", "control") {
			continue
		}
		nameSpan, ok := atoms.SelfNameSpanStartingAt(tokens[i+6].Span)
		if !ok {
			continue
		}
		end := i + 6
		for end < len(tokens) && spanCovers(nameSpan, tokens[end].Span) {
			end++
		}
		if end == len(tokens)-1 && tokens[end].Kind == shared.Period {
			return EffectDurationWhileYouControlSource
		}
	}
	return EffectDurationNone
}

func cutDelayedTiming(tokens []shared.Token) ([]shared.Token, DelayedTimingKind) {
	end := len(tokens)
	if end > 0 && tokens[end-1].Kind == shared.Period {
		end--
	}
	for _, suffix := range []struct {
		words  []string
		timing DelayedTimingKind
	}{
		{[]string{"at", "the", "beginning", "of", "the", "next", "end", "step"}, DelayedTimingNextEndStep},
		{[]string{"at", "the", "beginning", "of", "the", "next", "turn's", "upkeep"}, DelayedTimingNextUpkeep},
	} {
		start := end - len(suffix.words)
		if start >= 0 && effectWordsAt(tokens, start, suffix.words...) {
			return append(append([]shared.Token(nil), tokens[:start]...), tokens[end:]...), suffix.timing
		}
	}
	return tokens, DelayedTimingNone
}

func parsePTChange(tokens []shared.Token) (power, toughness SignedAmountSyntax) {
	for i := 0; i+4 < len(tokens); i++ {
		power, powerOK := parseSignedAmount(tokens[i], tokens[i+1])
		toughness, toughnessOK := parseSignedAmount(tokens[i+3], tokens[i+4])
		if powerOK && tokens[i+2].Kind == shared.Slash && toughnessOK {
			return power, toughness
		}
	}
	return SignedAmountSyntax{}, SignedAmountSyntax{}
}

func parseSignedAmount(sign, amount shared.Token) (SignedAmountSyntax, bool) {
	if amount.Kind != shared.Integer || sign.Kind != shared.Plus && sign.Kind != shared.Minus {
		return SignedAmountSyntax{}, false
	}
	value, err := strconv.Atoi(amount.Text)
	if err != nil {
		return SignedAmountSyntax{}, false
	}
	return SignedAmountSyntax{
		Span:     shared.Span{Start: sign.Span.Start, End: amount.Span.End},
		Value:    value,
		Known:    true,
		Negative: sign.Kind == shared.Minus,
	}, true
}

func parseEffectAmount(kind EffectKind, tokens []shared.Token, atoms Atoms) EffectAmountSyntax {
	if amount, attempted, ok := parseDynamicEffectAmount(tokens, atoms); attempted {
		if ok {
			return amount
		}
		return EffectAmountSyntax{}
	}
	if kind == EffectEnterTapped {
		for i, token := range tokens {
			if equalWord(token, "with") && i+1 < len(tokens) && equalWord(tokens[i+1], "X") {
				return EffectAmountSyntax{Span: tokens[i+1].Span, VariableX: true}
			}
		}
	}
	for _, token := range tokens {
		if token.Kind != shared.Word {
			continue
		}
		if equalWord(token, "X") {
			return EffectAmountSyntax{Span: token.Span, VariableX: true}
		}
		break
	}
	for i, token := range tokens {
		if value, ok := effectNumber(token, atoms); ok && value > 0 {
			if i > 0 && tokens[i-1].Kind == shared.Minus {
				return EffectAmountSyntax{}
			}
			return EffectAmountSyntax{Span: token.Span, Value: value, Known: true}
		}
		if equalWord(token, "a") || equalWord(token, "an") {
			return EffectAmountSyntax{Span: token.Span, Value: 1, Known: true}
		}
	}
	for _, token := range tokens {
		if token.Kind == shared.Symbol {
			return EffectAmountSyntax{Span: token.Span, Value: 1, Known: true}
		}
	}
	if (kind == EffectInvestigate || kind == EffectProliferate) &&
		len(tokens) == 1 && tokens[0].Kind == shared.Period {
		return EffectAmountSyntax{Value: 1, Known: true}
	}
	return EffectAmountSyntax{}
}

type dynamicAmountPrefix struct {
	form       EffectDynamicAmountForm
	start      int
	multiplier int
	plural     bool
	count      bool
}

type dynamicAmountSubject struct {
	amount EffectAmountSyntax
	end    int
	plural bool
	count  bool
}

func parseDynamicEffectAmount(tokens []shared.Token, atoms Atoms) (amount EffectAmountSyntax, attempted, ok bool) {
	var matches []EffectAmountSyntax
	for i := range tokens {
		prefix, prefixOK := parseDynamicAmountPrefix(tokens, i, atoms)
		if !prefixOK {
			continue
		}
		attempted = true
		subject, subjectOK := parseDynamicAmountSubject(tokens, prefix.start, atoms)
		if !subjectOK || subject.count != prefix.count || subject.count && subject.plural != prefix.plural {
			continue
		}
		match := subject.amount
		match.DynamicForm = prefix.form
		match.Multiplier = prefix.multiplier
		match.Span = shared.SpanOf(tokens[i:subject.end])
		match.Text = joinedEffectText(tokens[i:subject.end])
		matches = append(matches, match)
	}
	if len(matches) != 1 {
		return EffectAmountSyntax{}, attempted, false
	}
	return matches[0], true, true
}

func parseDynamicAmountPrefix(tokens []shared.Token, index int, atoms Atoms) (dynamicAmountPrefix, bool) {
	switch {
	case effectWordsAt(tokens, index, "equal", "to", "twice", "the", "number", "of"):
		return dynamicAmountPrefix{EffectDynamicAmountFormEqual, index + 6, 2, true, true}, true
	case effectWordsAt(tokens, index, "equal", "to", "the", "number", "of"):
		return dynamicAmountPrefix{EffectDynamicAmountFormEqual, index + 5, 1, true, true}, true
	case effectWordsAt(tokens, index, "for", "each"):
		return dynamicAmountPrefix{EffectDynamicAmountFormForEach, index + 2, precedingEffectMultiplier(tokens[:index], atoms), false, true}, true
	case effectWordsAt(tokens, index, "equal", "to"):
		return dynamicAmountPrefix{EffectDynamicAmountFormEqual, index + 2, 1, false, false}, true
	case effectWordsAt(tokens, index, "where", "X", "is", "twice", "the", "number", "of"):
		return dynamicAmountPrefix{EffectDynamicAmountFormWhereX, index + 7, 2, true, true}, true
	case effectWordsAt(tokens, index, "where", "X", "is", "the", "number", "of"):
		return dynamicAmountPrefix{EffectDynamicAmountFormWhereX, index + 6, 1, true, true}, true
	case effectWordsAt(tokens, index, "where", "X", "is"):
		return dynamicAmountPrefix{EffectDynamicAmountFormWhereX, index + 3, 1, false, false}, true
	default:
		return dynamicAmountPrefix{}, false
	}
}

func precedingEffectMultiplier(tokens []shared.Token, atoms Atoms) int {
	multiplier := 0
	for _, token := range tokens {
		value, ok := effectNumber(token, atoms)
		if !ok || value == 0 {
			continue
		}
		if multiplier != 0 && multiplier != value {
			return 0
		}
		multiplier = value
	}
	if multiplier == 0 {
		return 1
	}
	return multiplier
}

func parseDynamicAmountSubject(tokens []shared.Token, start int, atoms Atoms) (dynamicAmountSubject, bool) {
	if start >= len(tokens) {
		return dynamicAmountSubject{}, false
	}
	switch {
	case effectWordsAt(tokens, start, "your", "life", "total") && dynamicAmountBoundary(tokens, start+3):
		return dynamicAmountSubject{
			amount: EffectAmountSyntax{DynamicKind: EffectDynamicAmountControllerLife},
			end:    start + 3,
		}, true
	case effectWordsAt(tokens, start, "its", "power") && dynamicAmountBoundary(tokens, start+2):
		return dynamicAmountSubject{
			amount: EffectAmountSyntax{DynamicKind: EffectDynamicAmountSourcePower, ReferenceSpan: tokens[start].Span},
			end:    start + 2,
		}, true
	case effectWordsAt(tokens, start, "this", "creature") &&
		start+4 < len(tokens) && tokens[start+2].Kind == shared.Apostrophe &&
		equalWord(tokens[start+3], "s") && equalWord(tokens[start+4], "power") &&
		dynamicAmountBoundary(tokens, start+5):
		return dynamicAmountSubject{
			amount: EffectAmountSyntax{DynamicKind: EffectDynamicAmountSourcePower, ReferenceSpan: shared.SpanOf(tokens[start : start+2])},
			end:    start + 5,
		}, true
	case start+2 < len(tokens) && equalWord(tokens[start], "this") &&
		strings.EqualFold(tokens[start+1].Text, "creature's") &&
		equalWord(tokens[start+2], "power") && dynamicAmountBoundary(tokens, start+3):
		return dynamicAmountSubject{
			amount: EffectAmountSyntax{DynamicKind: EffectDynamicAmountSourcePower, ReferenceSpan: shared.SpanOf(tokens[start : start+2])},
			end:    start + 3,
		}, true
	case effectWordsAt(tokens, start, "basic", "land", "type", "among", "lands", "you", "control") &&
		dynamicAmountBoundary(tokens, start+7):
		return dynamicAmountSubject{
			amount: EffectAmountSyntax{DynamicKind: EffectDynamicAmountBasicLandTypes},
			end:    start + 7, count: true,
		}, true
	case effectWordsAt(tokens, start, "basic", "land", "types", "among", "lands", "you", "control") &&
		dynamicAmountBoundary(tokens, start+7):
		return dynamicAmountSubject{
			amount: EffectAmountSyntax{DynamicKind: EffectDynamicAmountBasicLandTypes},
			end:    start + 7, count: true, plural: true,
		}, true
	}
	if subject, ok := parseDynamicCountSubject(tokens, start, atoms); ok {
		return subject, true
	}
	nameSpan, ok := atoms.SelfNameSpanStartingAt(tokens[start].Span)
	if !ok {
		return dynamicAmountSubject{}, false
	}
	end := start
	for end < len(tokens) && tokens[end].Span.End.Offset <= nameSpan.End.Offset {
		end++
	}
	if end < len(tokens) && equalWord(tokens[end], "power") && dynamicAmountBoundary(tokens, end+1) {
		return dynamicAmountSubject{
			amount: EffectAmountSyntax{DynamicKind: EffectDynamicAmountSourcePower, ReferenceSpan: nameSpan},
			end:    end + 1,
		}, true
	}
	if end+2 < len(tokens) && tokens[end].Kind == shared.Apostrophe &&
		equalWord(tokens[end+1], "s") && equalWord(tokens[end+2], "power") &&
		dynamicAmountBoundary(tokens, end+3) {
		return dynamicAmountSubject{
			amount: EffectAmountSyntax{DynamicKind: EffectDynamicAmountSourcePower, ReferenceSpan: nameSpan},
			end:    end + 3,
		}, true
	}
	return dynamicAmountSubject{}, false
}

func parseDynamicCountSubject(tokens []shared.Token, start int, atoms Atoms) (dynamicAmountSubject, bool) {
	if equalWord(tokens[start], "card") || equalWord(tokens[start], "cards") {
		if subject, ok := parseDynamicCardCountSubject(tokens, start, atoms); ok {
			return subject, true
		}
	}
	noun, ok := atoms.ObjectNounAt(tokens[start].Span)
	if !ok {
		return dynamicAmountSubject{}, false
	}
	plural := strings.HasSuffix(strings.ToLower(tokens[start].Text), "s")
	if noun == ObjectNounOpponent {
		end := start + 1
		if effectWordsAt(tokens, end, "you", "have") {
			end += 2
		}
		if dynamicAmountBoundary(tokens, end) {
			return dynamicAmountSubject{
				amount: EffectAmountSyntax{DynamicKind: EffectDynamicAmountOpponentCount},
				end:    end, count: true, plural: plural,
			}, true
		}
		return dynamicAmountSubject{}, false
	}
	if !slices.Contains([]ObjectNoun{
		ObjectNounArtifact, ObjectNounCreature, ObjectNounEnchantment, ObjectNounLand, ObjectNounPermanent,
	}, noun) {
		return dynamicAmountSubject{}, false
	}
	end := start + 1
	for _, suffix := range [][]string{{"you", "control"}, {"your", "opponents", "control"}, {"on", "the", "battlefield"}} {
		if !effectWordsAt(tokens, end, suffix...) || !dynamicAmountBoundary(tokens, end+len(suffix)) {
			continue
		}
		subjectEnd := end + len(suffix)
		selection := parseSelection(tokens[start:subjectEnd], atoms)
		return dynamicAmountSubject{
			amount: EffectAmountSyntax{DynamicKind: EffectDynamicAmountCount, Selection: &selection},
			end:    subjectEnd, count: true, plural: plural,
		}, true
	}
	return dynamicAmountSubject{}, false
}

func parseDynamicCardCountSubject(tokens []shared.Token, start int, atoms Atoms) (dynamicAmountSubject, bool) {
	end := start + 1
	if end >= len(tokens) {
		return dynamicAmountSubject{}, false
	}
	keyword, ok := atoms.KeywordSelectorStartingAt(tokens[end].Span)
	if !ok || keyword.Excluded || keyword.Keyword != KeywordCycling {
		return dynamicAmountSubject{}, false
	}
	for end < len(tokens) && tokens[end].Span.End.Offset <= keyword.Span.End.Offset {
		end++
	}
	if !effectWordsAt(tokens, end, "in", "your", "graveyard") || !dynamicAmountBoundary(tokens, end+3) {
		return dynamicAmountSubject{}, false
	}
	end += 3
	selection := parseSelection(tokens[start:end], atoms)
	selection.Kind = SelectionCard
	selection.Controller = SelectionControllerYou
	selection.Zone = zone.Graveyard
	return dynamicAmountSubject{
		amount: EffectAmountSyntax{DynamicKind: EffectDynamicAmountCount, Selection: &selection},
		end:    end, count: true, plural: strings.EqualFold(tokens[start].Text, "cards"),
	}, true
}

func dynamicAmountBoundary(tokens []shared.Token, end int) bool {
	if end >= len(tokens) {
		return true
	}
	if tokens[end].Kind == shared.Comma || tokens[end].Kind == shared.Period {
		return true
	}
	return equalWord(tokens[end], "to") || equalWord(tokens[end], "until")
}

func effectNumber(token shared.Token, atoms Atoms) (int, bool) {
	if token.Kind == shared.Integer {
		value, err := strconv.Atoi(token.Text)
		return value, err == nil
	}
	return atoms.CardinalAt(token.Span)
}

func parseCounterPlacement(tokens []shared.Token, atoms Atoms) (counter.Kind, bool) {
	for _, token := range tokens {
		if equalWord(token, "and") {
			return counter.Kind(0), false
		}
	}
	span := shared.SpanOf(tokens)
	var kinds []counter.Kind
	for _, atom := range atoms.Counters() {
		if spanCovers(span, atom.Span) {
			kinds = append(kinds, atom.Kind)
		}
	}
	if len(kinds) != 1 {
		return counter.Kind(0), false
	}
	kind := kinds[0]
	return kind, kind.Valid() && kind != counter.Stun && kind != counter.Finality
}

func firstZone(atoms Atoms, span shared.Span, role ZoneRole) zone.Type {
	result := zone.None
	for _, atom := range atoms.Zones() {
		if atom.Role != role || !spanCovers(span, atom.Span) {
			continue
		}
		if result != zone.None && atom.Zone != result {
			return zone.None
		}
		result = atom.Zone
	}
	return result
}

func firstEffectSymbol(tokens []shared.Token) string {
	for _, token := range tokens {
		if token.Kind == shared.Symbol {
			return token.Text
		}
	}
	return ""
}

func referencesInSpan(atoms Atoms, span shared.Span) []Reference {
	var references []Reference
	for _, reference := range atoms.References() {
		if spanCovers(span, reference.Span) {
			references = append(references, reference)
		}
	}
	return references
}

func parseTargets(tokens []shared.Token, atoms Atoms) []TargetSyntax {
	var targets []TargetSyntax
	for i, token := range tokens {
		if !equalWord(token, "target") {
			continue
		}
		start := i
		cardinality := TargetCardinalitySyntax{Min: 1, Max: 1}
		switch {
		case i >= 3 && effectWordsAt(tokens, i-3, "any", "number", "of"):
			start = i - 3
			cardinality = TargetCardinalitySyntax{Min: 0, Max: 99}
		case i >= 4 && effectWordsAt(tokens, i-4, "up", "to") &&
			(effectWordsAt(tokens, i-1, "another") || effectWordsAt(tokens, i-1, "other")):
			start = i - 4
			cardinality.Min = 0
			var ok bool
			cardinality.Max, ok = effectNumber(tokens[i-2], atoms)
			if !ok || cardinality.Max < 1 {
				cardinality = TargetCardinalitySyntax{}
			}
		case i >= 3 && effectWordsAt(tokens, i-3, "up", "to"):
			start = i - 3
			cardinality.Min = 0
			var ok bool
			cardinality.Max, ok = effectNumber(tokens[i-1], atoms)
			if !ok || cardinality.Max < 1 {
				cardinality = TargetCardinalitySyntax{}
			}
		case i >= 1:
			if count, ok := effectNumber(tokens[i-1], atoms); ok && count > 0 {
				start = i - 1
				cardinality = TargetCardinalitySyntax{Min: count, Max: count}
			} else if equalWord(tokens[i-1], "any") ||
				equalWord(tokens[i-1], "another") ||
				equalWord(tokens[i-1], "other") {
				start = i - 1
			}
		default:
		}
		end := targetSyntaxEnd(tokens, i+1)
		selectionTokens := append([]shared.Token(nil), tokens[start:i]...)
		selectionTokens = append(selectionTokens, tokens[i+1:end]...)
		selection := parseSelection(selectionTokens, atoms)
		if targetSelectionHasUnsupportedQualifier(selectionTokens, atoms) {
			selection = SelectionSyntax{Span: selection.Span, Text: selection.Text}
		}
		targets = append(targets, TargetSyntax{
			Span:        shared.SpanOf(tokens[start:end]),
			Text:        joinedEffectText(tokens[start:end]),
			Cardinality: cardinality,
			Selection:   selection,
			Exact:       exactRuntimeTargetSyntax(tokens[start:end], cardinality, selection),
		})
	}
	return targets
}

func exactRuntimeTargetSyntax(tokens []shared.Token, cardinality TargetCardinalitySyntax, selection SelectionSyntax) bool {
	if cardinality != (TargetCardinalitySyntax{Min: 1, Max: 1}) {
		return false
	}
	text := joinedEffectText(tokens)
	switch selection.Kind {
	case SelectionAny:
		return text == "any target"
	case SelectionPlayer:
		return strings.EqualFold(text, "target player")
	case SelectionOpponent:
		return strings.EqualFold(text, "target opponent")
	case SelectionActivatedAbility:
		return strings.EqualFold(text, "target activated ability")
	case SelectionTriggeredAbility:
		return strings.EqualFold(text, "target triggered ability")
	case SelectionActivatedOrTriggeredAbility:
		return strings.EqualFold(text, "target activated or triggered ability")
	case SelectionSpellActivatedOrTriggeredAbility:
		return strings.EqualFold(text, "target spell, activated ability, or triggered ability")
	case SelectionSpell:
		switch strings.ToLower(text) {
		case "target spell", "target instant spell", "target sorcery spell", "target creature spell",
			"target artifact spell", "target noncreature spell":
			return true
		}
		return false
	case SelectionCreature:
		if strings.EqualFold(text, "target creature spell") {
			return true
		}
	case SelectionArtifact:
		if strings.EqualFold(text, "target artifact spell") {
			return true
		}
	default:
	}

	noun := ""
	switch selection.Kind {
	case SelectionArtifact:
		noun = "artifact"
	case SelectionBattle:
		noun = "battle"
	case SelectionCreature:
		noun = "creature"
	case SelectionEnchantment:
		noun = "enchantment"
	case SelectionLand:
		noun = "land"
	case SelectionPermanent:
		noun = "permanent"
	case SelectionPlaneswalker:
		noun = "planeswalker"
	default:
		return false
	}
	if selection.Another || selection.Other ||
		(selection.Tapped && selection.Untapped) ||
		((selection.Tapped || selection.Untapped) && (selection.Attacking || selection.Blocking)) {
		return false
	}
	expected := "target "
	switch {
	case selection.Attacking && selection.Blocking:
		expected += "attacking or blocking "
	case selection.Attacking:
		expected += "attacking "
	case selection.Blocking:
		expected += "blocking "
	case selection.Tapped:
		expected += "tapped "
	case selection.Untapped:
		expected += "untapped "
	default:
	}
	expected += noun
	switch selection.Controller {
	case SelectionControllerAny:
	case SelectionControllerYou:
		expected += " you control"
	case SelectionControllerOpponent:
		expected += " an opponent controls"
	case SelectionControllerNotYou:
		expected += " you don't control"
	default:
		return false
	}
	return strings.EqualFold(text, expected)
}

func targetSelectionHasUnsupportedQualifier(tokens []shared.Token, atoms Atoms) bool {
	for _, token := range tokens {
		if token.Kind == shared.Integer || token.Kind == shared.Comma ||
			selectionGrammarWord(token) || selectionAtomCoversToken(atoms, token) {
			continue
		}
		return true
	}
	return false
}

func selectionGrammarWord(token shared.Token) bool {
	for _, word := range []string{
		"a", "an", "all", "any", "number", "of", "up", "to", "or", "and",
		"with", "without", "from", "in", "your", "you", "control", "controls", "don't",
		"opponent", "opponent's", "opponents", "activated", "triggered",
		"mana", "value", "power", "toughness", "equal", "less", "greater",
		"battlefield", "graveyard", "hand", "library", "exile", "command",
	} {
		if equalWord(token, word) {
			return true
		}
	}
	return false
}

func selectionAtomCoversToken(atoms Atoms, token shared.Token) bool {
	covered := func(span shared.Span) bool {
		return spanCovers(span, token.Span)
	}
	for _, atom := range atoms.Colors() {
		if covered(atom.Span) {
			return true
		}
	}
	for _, atom := range atoms.ExcludedColors() {
		if covered(atom.Span) {
			return true
		}
	}
	for _, atom := range atoms.ColorQualifiers() {
		if covered(atom.Span) {
			return true
		}
	}
	for _, atom := range atoms.CardTypes() {
		if covered(atom.Span) {
			return true
		}
	}
	for _, atom := range atoms.ExcludedTypes() {
		if covered(atom.Span) {
			return true
		}
	}
	for _, atom := range atoms.Supertypes() {
		if covered(atom.Span) {
			return true
		}
	}
	for _, atom := range atoms.Subtypes() {
		if covered(atom.Span) {
			return true
		}
	}
	for _, atom := range atoms.ObjectNouns() {
		if covered(atom.Span) {
			return true
		}
	}
	for _, atom := range atoms.Zones() {
		if covered(atom.Span) {
			return true
		}
	}
	for _, atom := range atoms.Cardinals() {
		if covered(atom.Span) {
			return true
		}
	}
	for _, atom := range atoms.SelectionFlags() {
		if covered(atom.Span) {
			return true
		}
	}
	for _, atom := range atoms.Controllers() {
		if covered(atom.Span) {
			return true
		}
	}
	for _, atom := range atoms.KeywordSelectors() {
		if covered(atom.Span) {
			return true
		}
	}
	return false
}

func targetSyntaxEnd(tokens []shared.Token, start int) int {
	if start+8 <= len(tokens) &&
		effectWordsAt(tokens, start, "spell") &&
		tokens[start+1].Kind == shared.Comma &&
		effectWordsAt(tokens, start+2, "activated", "ability") &&
		tokens[start+4].Kind == shared.Comma &&
		effectWordsAt(tokens, start+5, "or", "triggered", "ability") {
		return start + 8
	}
	end := start
	for end < len(tokens) {
		token := tokens[end]
		if token.Kind == shared.Comma || token.Kind == shared.Period || token.Kind == shared.Semicolon ||
			targetDestinationStartsAt(tokens, end) ||
			equalWord(token, "unless") ||
			(equalWord(token, "and") && end+2 < len(tokens) && equalWord(tokens[end+1], "you") && effectWordKind(tokens[end+2]) != EffectUnknown) ||
			(equalWord(token, "and") && end+1 < len(tokens) && effectWordKind(tokens[end+1]) != EffectUnknown) ||
			(end > start && effectWordKind(token) != EffectUnknown) ||
			(equalWord(token, "until") && end+1 < len(tokens)) ||
			(equalWord(token, "for") && effectWordsAt(tokens, end, "for", "as", "long", "as")) ||
			(equalWord(token, "as") && effectWordsAt(tokens, end, "as", "long", "as", "this")) {
			break
		}

		end++
	}

	return end
}

func targetDestinationStartsAt(tokens []shared.Token, index int) bool {
	for _, phrase := range [][]string{
		{"to", "its", "owner's", "hand"},
		{"to", "your", "hand"},
		{"to", "their", "hand"},
		{"to", "the", "battlefield"},
		{"onto", "the", "battlefield"},
		{"into", "your", "graveyard"},
		{"into", "your", "library"},
		{"on", "top", "of", "your", "library"},
		{"on", "the", "top", "of", "your", "library"},
		{"on", "bottom", "of", "your", "library"},
		{"on", "the", "bottom", "of", "your", "library"},
	} {
		if effectWordsAt(tokens, index, phrase...) {
			return true
		}
	}
	return false
}

func ambiguousZoneChoice(tokens []shared.Token, atoms Atoms, span shared.Span) bool {
	zones := atoms.Zones()
	for i, first := range zones {
		if !spanCovers(span, first.Span) {
			continue
		}
		for _, second := range zones[i+1:] {
			if first.Zone == second.Zone || !spanCovers(span, second.Span) {
				continue
			}
			for _, token := range tokens {
				if token.Span.Start.Offset >= first.Span.End.Offset &&
					token.Span.End.Offset <= second.Span.Start.Offset &&
					equalWord(token, "or") {
					return true
				}
			}
		}
	}
	return false
}

func parseSelection(tokens []shared.Token, atoms Atoms) SelectionSyntax {
	selection := SelectionSyntax{Span: shared.SpanOf(tokens), Text: joinedEffectText(tokens)}
	words := shared.NormalizedWords(tokens)
	switch {
	case slices.Equal(words, []string{"activated", "ability"}):
		selection.Kind = SelectionActivatedAbility
	case slices.Equal(words, []string{"triggered", "ability"}):
		selection.Kind = SelectionTriggeredAbility
	case slices.Equal(words, []string{"activated", "or", "triggered", "ability"}):
		selection.Kind = SelectionActivatedOrTriggeredAbility
	case slices.Equal(words, []string{"spell", "activated", "ability", "or", "triggered", "ability"}):
		selection.Kind = SelectionSpellActivatedOrTriggeredAbility
	default:
	}
	for _, token := range tokens {
		if noun, ok := atoms.ObjectNounAt(token.Span); ok && selection.Kind == SelectionUnknown {
			selection.Kind = selectionKindForNoun(noun)
		}
		if cardType, ok := atoms.CardTypeAt(token.Span); ok && !slices.Contains(selection.RequiredTypesAny, cardType) {
			selection.RequiredTypesAny = append(selection.RequiredTypesAny, cardType)
		}
		if cardType, ok := atoms.ExcludedCardTypeAt(token.Span); ok && !slices.Contains(selection.ExcludedTypes, cardType) {
			selection.ExcludedTypes = append(selection.ExcludedTypes, cardType)
		}
		if colorValue, ok := atoms.ColorAt(token.Span); ok && !slices.Contains(selection.ColorsAny, colorValue) {
			selection.ColorsAny = append(selection.ColorsAny, colorValue)
		}
		if colorValue, ok := atoms.ExcludedColorAt(token.Span); ok && !slices.Contains(selection.ExcludedColors, colorValue) {
			selection.ExcludedColors = append(selection.ExcludedColors, colorValue)
		}
		if supertype, ok := atoms.SupertypeAt(token.Span); ok && !slices.Contains(selection.Supertypes, supertype) {
			selection.Supertypes = append(selection.Supertypes, supertype)
		}
	}
	for _, token := range tokens {
		if noun, ok := atoms.ObjectNounAt(token.Span); ok && noun == ObjectNounSpell &&
			selection.Kind != SelectionActivatedAbility &&
			selection.Kind != SelectionTriggeredAbility &&
			selection.Kind != SelectionActivatedOrTriggeredAbility &&
			selection.Kind != SelectionSpellActivatedOrTriggeredAbility {
			selection.Kind = SelectionSpell
			break
		}
	}
	for _, token := range tokens {
		if noun, ok := atoms.ObjectNounAt(token.Span); ok && noun == ObjectNounAbility &&
			selection.Kind != SelectionActivatedAbility &&
			selection.Kind != SelectionTriggeredAbility &&
			selection.Kind != SelectionActivatedOrTriggeredAbility &&
			selection.Kind != SelectionSpellActivatedOrTriggeredAbility {
			selection.Kind = SelectionUnknown
			break
		}
	}
	span := shared.SpanOf(tokens)
	selection.SubtypesAny = atoms.SubtypesIn(span)
	if relation, ok := atoms.ControllerIn(span); ok {
		switch relation {
		case ControllerRelationYouControl:
			selection.Controller = SelectionControllerYou
		case ControllerRelationOpponentControls:
			selection.Controller = SelectionControllerOpponent
		case ControllerRelationYouDontControl:
			selection.Controller = SelectionControllerNotYou
		default:
		}
	}
	selection.Zone = firstZone(atoms, span, ZoneRoleFrom)
	if selection.Zone == zone.None {
		selection.Zone = firstZone(atoms, span, ZoneRolePlain)
	}
	switch {
	case effectContainsWords(words, "your", "graveyard"):
		selection.Controller = SelectionControllerYou
	case effectContainsWords(words, "opponent's", "graveyard"):
		selection.Controller = SelectionControllerOpponent
	default:
	}
	selection.All = slices.Contains(words, "all")
	selection.Another = atoms.SelectionFlagIn(span, SelectionFlagAnother)
	selection.Other = atoms.SelectionFlagIn(span, SelectionFlagOther)
	selection.Attacking = atoms.SelectionFlagIn(span, SelectionFlagAttacking)
	selection.Blocking = atoms.SelectionFlagIn(span, SelectionFlagBlocking)
	selection.Tapped = atoms.SelectionFlagIn(span, SelectionFlagTapped)
	selection.Untapped = atoms.SelectionFlagIn(span, SelectionFlagUntapped)
	if slices.Contains(words, "any") && selection.Kind == SelectionUnknown {
		selection.Kind = SelectionAny
	}
	if keyword, ok := atoms.KeywordSelectorIn(span, false); ok {
		selection.Keyword = keyword.Keyword
	}
	if !parseSelectionNumbers(tokens, atoms, &selection) {
		return SelectionSyntax{Span: span, Text: joinedEffectText(tokens)}
	}
	return selection
}

func parseSelectionNumbers(tokens []shared.Token, atoms Atoms, selection *SelectionSyntax) bool {
	for i := range tokens {
		if i+2 < len(tokens) && effectWordsAt(tokens, i, "mana", "value") {
			comparison, ok := parseSelectionNumberComparison(tokens[i+2:], atoms)
			if !ok {
				return false
			}
			selection.ManaValue = comparison
			selection.MatchManaValue = true
			continue
		}
		if equalWord(tokens[i], "power") {
			comparison, ok := parseSelectionNumberComparison(tokens[i+1:], atoms)
			if !ok {
				return false
			}
			selection.Power = comparison
			selection.MatchPower = true
			continue
		}
		if equalWord(tokens[i], "toughness") {
			comparison, ok := parseSelectionNumberComparison(tokens[i+1:], atoms)
			if !ok {
				return false
			}
			selection.Toughness = comparison
			selection.MatchToughness = true
		}
	}
	return true
}

func parseSelectionNumberComparison(tokens []shared.Token, atoms Atoms) (compare.Int, bool) {
	if len(tokens) == 0 {
		return compare.Int{}, false
	}
	if value, ok := effectNumber(tokens[0], atoms); ok {
		if len(tokens) >= 3 && equalWord(tokens[1], "or") {
			switch {
			case equalWord(tokens[2], "less"):
				return compare.Int{Op: compare.LessOrEqual, Value: value}, true
			case equalWord(tokens[2], "greater"):
				return compare.Int{Op: compare.GreaterOrEqual, Value: value}, true
			}
		}
		return compare.Int{Op: compare.Equal, Value: value}, true
	}
	if len(tokens) >= 3 && effectWordsAt(tokens, 0, "equal", "to") {
		if value, ok := effectNumber(tokens[2], atoms); ok {
			return compare.Int{Op: compare.Equal, Value: value}, true
		}
	}
	return compare.Int{}, false
}

func parseEffectStaticSubject(tokens []shared.Token, atoms Atoms) EffectStaticSubjectSyntax {
	subtype := func(index int) (types.Sub, bool) {
		if index >= len(tokens) {
			return "", false
		}
		value, ok := atoms.SubtypeAt(tokens[index].Span)
		return value, ok && SubtypeMatchesAnyRuntimeCardType(value, []types.Card{types.Creature, types.Kindred})
	}
	switch {
	case len(tokens) >= 3 &&
		(equalWord(tokens[0], "enchanted") || equalWord(tokens[0], "equipped")) &&
		equalWord(tokens[1], "creature") &&
		(equalWord(tokens[2], "gets") || equalWord(tokens[2], "has")):
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectAttachedObject, Span: shared.SpanOf(tokens[:2])}
	case len(tokens) >= 5 && effectWordsAt(tokens, 0, "other", "creatures", "you", "control") &&
		(equalWord(tokens[4], "get") || equalWord(tokens[4], "have")):
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectOtherControlledCreatures, Span: shared.SpanOf(tokens[:4])}
	case len(tokens) >= 4 && effectWordsAt(tokens, 0, "creatures", "you", "control") &&
		(equalWord(tokens[3], "get") || equalWord(tokens[3], "have")):
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectControlledCreatures, Span: shared.SpanOf(tokens[:3])}
	case len(tokens) >= 5 && effectWordsAt(tokens, 0, "creatures", "your", "opponents", "control") &&
		(equalWord(tokens[4], "get") || equalWord(tokens[4], "have")):
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectOpponentControlledCreatures, Span: shared.SpanOf(tokens[:4])}
	case len(tokens) >= 5 && effectWordsAt(tokens, 0, "each", "wall", "you", "control") &&
		(equalWord(tokens[4], "gets") || equalWord(tokens[4], "has")):
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectControlledWalls, Span: shared.SpanOf(tokens[:4]), Subtype: types.Wall, SubtypeKnown: true}
	case len(tokens) >= 4 && effectWordsAt(tokens, 0, "walls", "you", "control") &&
		(equalWord(tokens[3], "get") || equalWord(tokens[3], "have")):
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectControlledWalls, Span: shared.SpanOf(tokens[:3]), Subtype: types.Wall, SubtypeKnown: true}
	case len(tokens) >= 4 && effectWordsAt(tokens, 0, "artifacts", "you", "control") &&
		(equalWord(tokens[3], "get") || equalWord(tokens[3], "have")):
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectControlledArtifacts, Span: shared.SpanOf(tokens[:3])}
	case len(tokens) >= 4 && effectWordsAt(tokens, 0, "tokens", "you", "control") &&
		(equalWord(tokens[3], "get") || equalWord(tokens[3], "have")):
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectControlledTokens, Span: shared.SpanOf(tokens[:3])}
	case len(tokens) >= 5 && equalWord(tokens[0], "other") && effectWordsAt(tokens, 2, "you", "control", "have"):
		value, ok := subtype(1)
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectOtherControlledCreatureSubtype, Span: shared.SpanOf(tokens[:4]), Subtype: value, SubtypeText: tokens[1].Text, SubtypeKnown: ok}
	case len(tokens) >= 4 && effectWordsAt(tokens, 1, "you", "control", "have"):
		value, ok := subtype(0)
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectControlledCreatureSubtype, Span: shared.SpanOf(tokens[:3]), Subtype: value, SubtypeText: tokens[0].Text, SubtypeKnown: ok}
	default:
		return EffectStaticSubjectSyntax{}
	}
}

func selectionKindForNoun(noun ObjectNoun) SelectionKind {
	switch noun {
	case ObjectNounArtifact:
		return SelectionArtifact
	case ObjectNounCard:
		return SelectionCard
	case ObjectNounCreature:
		return SelectionCreature
	case ObjectNounEnchantment:
		return SelectionEnchantment
	case ObjectNounLand:
		return SelectionLand
	case ObjectNounOpponent:
		return SelectionOpponent
	case ObjectNounPermanent:
		return SelectionPermanent
	case ObjectNounPlaneswalker:
		return SelectionPlaneswalker
	case ObjectNounPlayer:
		return SelectionPlayer
	case ObjectNounSpell:
		return SelectionSpell
	default:
		return SelectionUnknown
	}
}

func effectWordsAt(tokens []shared.Token, start int, words ...string) bool {
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

func effectContainsWords(words []string, sequence ...string) bool {
	for i := 0; i+len(sequence) <= len(words); i++ {
		if slices.Equal(words[i:i+len(sequence)], sequence) {
			return true
		}
	}
	return false
}

func joinedEffectText(tokens []shared.Token) string {
	var builder strings.Builder
	for i, token := range tokens {
		if i > 0 && token.Span.Start.Offset > tokens[i-1].Span.End.Offset {
			_ = builder.WriteByte(' ')
		}
		_, _ = builder.WriteString(token.Text)
	}
	return builder.String()
}
