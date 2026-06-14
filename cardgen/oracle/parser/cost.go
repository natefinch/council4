package parser

import (
	"slices"
	"strconv"
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// This file owns activation/loyalty cost grammar. The parser recognizes the
// comma-separated cost components of an activated, loyalty, or otherwise
// cost-bearing ability and emits a source-spanned typed Cost. Cost verbs,
// object structure, suffixes, and the loyalty sign are Oracle spelling the
// parser owns here; the compiler consumes the typed Cost mechanically and never
// re-reads cost text to derive meaning.

// CostComponentKind identifies one comma-separated cost operation.
type CostComponentKind string

// Cost component kinds recognized by the parser.
const (
	CostComponentUnknown         CostComponentKind = ""
	CostComponentMana            CostComponentKind = "CostComponentMana"
	CostComponentTap             CostComponentKind = "CostComponentTap"
	CostComponentUntap           CostComponentKind = "CostComponentUntap"
	CostComponentSacrifice       CostComponentKind = "CostComponentSacrifice"
	CostComponentDiscard         CostComponentKind = "CostComponentDiscard"
	CostComponentPayLife         CostComponentKind = "CostComponentPayLife"
	CostComponentExile           CostComponentKind = "CostComponentExile"
	CostComponentRemoveCounter   CostComponentKind = "CostComponentRemoveCounter"
	CostComponentReveal          CostComponentKind = "CostComponentReveal"
	CostComponentTapPermanents   CostComponentKind = "CostComponentTapPermanents"
	CostComponentEnergy          CostComponentKind = "CostComponentEnergy"
	CostComponentReturn          CostComponentKind = "CostComponentReturn"
	CostComponentExert           CostComponentKind = "CostComponentExert"
	CostComponentMill            CostComponentKind = "CostComponentMill"
	CostComponentPutCounter      CostComponentKind = "CostComponentPutCounter"
	CostComponentCollectEvidence CostComponentKind = "CostComponentCollectEvidence"
	CostComponentLoyalty         CostComponentKind = "CostComponentLoyalty"
)

// Cost is the ordered, source-spanned typed cost the parser recognizes before an
// activated or loyalty ability's colon.
type Cost struct {
	Span       shared.Span     `json:"-"`
	Text       string          `json:",omitempty"`
	Components []CostComponent `json:",omitempty"`
}

// CostComponent is one typed comma-separated cost operation. The compiler maps
// these typed fields onto its semantic cost IR; the string fields (Text, Symbol,
// Amount, Object) are retained rendering/diagnostic metadata and a genuine mana
// literal (Symbol), never re-parsed for structural meaning.
type CostComponent struct {
	Kind   CostComponentKind `json:",omitempty"`
	Span   shared.Span       `json:"-"`
	Text   string            `json:",omitempty"`
	Symbol string            `json:",omitempty"`
	Amount string            `json:",omitempty"`
	Object string            `json:",omitempty"`

	AmountValue int  `json:",omitempty"`
	AmountKnown bool `json:",omitempty"`
	AmountFromX bool `json:",omitempty"`

	// ObjectNoun is the recognized object noun, if any. ObjectIsCard reports
	// that the object is selected as a card (rather than a permanent), which
	// changes how the compiler maps the noun onto a selector and card type.
	ObjectNoun   ObjectNoun `json:",omitempty"`
	ObjectIsCard bool       `json:",omitempty"`

	ObjectSupertype  types.Super        `json:",omitempty"`
	SupertypeKnown   bool               `json:",omitempty"`
	ObjectColor      Color              `json:",omitempty"`
	ObjectColorKnown bool               `json:",omitempty"`
	ObjectController ControllerRelation `json:",omitempty"`
	RequireTapped    bool               `json:",omitempty"`
	RequireUntapped  bool               `json:",omitempty"`
	SourceZone       zone.Type          `json:",omitempty"`
	ToZone           zone.Type          `json:",omitempty"`
	SourceSelf       bool               `json:",omitempty"`
	CounterKind      counter.Kind       `json:",omitempty"`
	CounterKindKnown bool               `json:",omitempty"`
	SubtypesAny      []types.Sub        `json:",omitempty"`
}

// emitCost fills each ability's typed Cost from its cost phrase and atoms.
func emitCost(abilities []Ability) {
	for i := range abilities {
		ability := &abilities[i]
		if ability.costPhrase == nil {
			continue
		}
		cost := parseCost(*ability.costPhrase, ability.Kind, ability.Atoms)
		ability.CostSyntax = &cost
	}
}

// parseCost recognizes the typed cost components of a cost phrase. Loyalty
// abilities carry a single signed/variable loyalty amount; every other ability
// splits its cost on top-level commas and recognizes each component's verb and
// typed object.
func parseCost(phrase Phrase, abilityKind AbilityKind, atoms Atoms) Cost {
	cost := Cost{Span: phrase.Span, Text: phrase.Text}
	parts := splitTopLevelTokens(phrase.Tokens, shared.Comma)
	for _, part := range parts {
		if len(part) == 0 {
			continue
		}
		component := CostComponent{
			Kind: CostComponentUnknown,
			Span: shared.SpanOf(part),
			Text: shared.SliceSpan(phrase.Text, costRelativeSpan(shared.SpanOf(part), phrase.Span.Start.Offset)),
		}
		if abilityKind == AbilityLoyalty {
			component.Kind = CostComponentLoyalty
			component.Amount = costJoinedTokenText(part)
			if value, ok := signedLoyaltyAmount(component.Amount); ok {
				component.AmountValue = value
				component.AmountKnown = true
			} else if isVariableLoyaltyAmount(component.Amount) {
				component.AmountFromX = true
			}
		} else {
			words := normalizedWords(part)
			switch {
			case len(part) == 1 && part[0].Kind == shared.Symbol && strings.EqualFold(part[0].Text, "{T}"):
				component.Kind = CostComponentTap
				component.Symbol = part[0].Text
			case len(part) == 1 && part[0].Kind == shared.Symbol && strings.EqualFold(part[0].Text, "{Q}"):
				component.Kind = CostComponentUntap
				component.Symbol = part[0].Text
			case startsWords(words, "sacrifice"):
				component.Kind = CostComponentSacrifice
				component.Object = wordsAfterFirst(part)
			case startsWords(words, "discard"):
				component.Kind = CostComponentDiscard
				component.Object = wordsAfterFirst(part)
			case startsWords(words, "pay") && slices.Contains(words, "life"):
				component.Kind = CostComponentPayLife
				component.Amount = firstInteger(part)
			case startsWords(words, "pay") && allEnergySymbols(part[1:]):
				component.Kind = CostComponentEnergy
				component.Amount = strconv.Itoa(len(part) - 1)
				component.AmountValue = len(part) - 1
				component.AmountKnown = true
			case startsWords(words, "return") && slices.Contains(words, "hand"):
				component.Kind = CostComponentReturn
				component.Object = wordsAfterFirst(part)
			case startsWords(words, "reveal"):
				component.Kind = CostComponentReveal
				component.Object = wordsAfterFirst(part)
			case startsWords(words, "exert"):
				component.Kind = CostComponentExert
				component.Object = wordsAfterFirst(part)
			case startsWords(words, "mill"):
				component.Kind = CostComponentMill
				component.Object = wordsAfterFirst(part)
			case startsWords(words, "put") && containsNoun(words, "counter"):
				component.Kind = CostComponentPutCounter
				component.Object = wordsAfterFirst(part)
			case startsWords(words, "collect", "evidence") && len(part) == 3 && positiveIntegerWord(firstInteger(part)):
				component.Kind = CostComponentCollectEvidence
				component.Amount = firstInteger(part)
			case startsWords(words, "exile"):
				component.Kind = CostComponentExile
				component.Object = wordsAfterFirst(part)
			case startsWords(words, "remove") && (slices.Contains(words, "counter") || slices.Contains(words, "counters")):
				component.Kind = CostComponentRemoveCounter
				component.Object = wordsAfterFirst(part)
			case startsWords(words, "tap"):
				component.Kind = CostComponentTapPermanents
				component.Object = wordsAfterFirst(part)
			case allSymbols(part):
				component.Kind = CostComponentMana
				component.Symbol = costJoinedTokenText(part)
			default:
			}
		}
		parseCostAtoms(&component, part, atoms)
		cost.Components = append(cost.Components, component)
	}
	return cost
}

func parseCostAtoms(component *CostComponent, tokens []shared.Token, atoms Atoms) {
	if len(tokens) == 0 {
		return
	}
	object := tokens[1:]
	switch component.Kind {
	case CostComponentPayLife, CostComponentCollectEvidence:
		annotateIntegerCostAmount(component)
	case CostComponentEnergy:
		component.AmountValue = len(tokens) - 1
		component.AmountKnown = component.AmountValue > 0
	case CostComponentSacrifice:
		if costSelfReference(object, atoms, false) {
			component.SourceSelf = true
			return
		}
		annotateExactCostObject(component, object, atoms, false)
	case CostComponentDiscard:
		annotateExactCostObject(component, object, atoms, true)
	case CostComponentExile:
		if costSelfReference(object, atoms, false) {
			component.SourceSelf = true
			return
		}
		annotateExileCostObject(component, object, atoms)
	case CostComponentExert:
		component.SourceSelf = costSelfReference(object, atoms, true)
	case CostComponentMill:
		annotateMillCostObject(component, object, atoms)
	case CostComponentPutCounter:
		annotatePutCounterCostObject(component, object, atoms)
	case CostComponentRemoveCounter:
		annotateRemoveCounterCostObject(component, object, atoms)
	case CostComponentReveal:
		annotateRevealCostObject(component, object, atoms)
	case CostComponentReturn:
		annotateReturnCostObject(component, object, atoms)
	case CostComponentTapPermanents:
		annotateTapPermanentsCostObject(component, object, atoms)
	default:
	}
}

func annotateIntegerCostAmount(component *CostComponent) {
	amount, err := strconv.Atoi(component.Amount)
	if err == nil && amount > 0 {
		component.AmountValue = amount
		component.AmountKnown = true
	}
}

func costAmountAt(component *CostComponent, token shared.Token, atoms Atoms, allowX bool) bool {
	switch {
	case allowX && equalWord(token, "x"):
		component.AmountFromX = true
		return true
	case equalWord(token, "a"), equalWord(token, "an"):
		component.AmountValue = 1
		component.AmountKnown = true
		return true
	case token.Kind == shared.Integer:
		if value, err := strconv.Atoi(token.Text); err == nil && value > 0 {
			component.AmountValue = value
			component.AmountKnown = true
			return true
		}
	default:
		if value, ok := atoms.CardinalAt(token.Span); ok {
			component.AmountValue = value
			component.AmountKnown = true
			return true
		}
	}
	return false
}

// costObjectNounAccepted reports whether a noun is a recognized cost object
// noun. The parser records the typed noun; the compiler maps it onto a selector.
func costObjectNounAccepted(noun ObjectNoun) bool {
	switch noun {
	case ObjectNounArtifact, ObjectNounCreature, ObjectNounEnchantment,
		ObjectNounLand, ObjectNounPermanent, ObjectNounCard:
		return true
	default:
		return false
	}
}

func annotateCostObjectNoun(component *CostComponent, noun ObjectNoun) bool {
	if !costObjectNounAccepted(noun) {
		return false
	}
	component.ObjectNoun = noun
	return true
}

func annotateExactCostObject(component *CostComponent, object []shared.Token, atoms Atoms, cardObject bool) {
	words := object
	if cardObject {
		if len(words) < 2 || !costCardNoun(words[len(words)-1], atoms) {
			return
		}
		words = words[:len(words)-1]
	}
	if len(words) == 4 {
		if !equalWord(words[2], "you") || !equalWord(words[3], "control") {
			return
		}
		words = words[:2]
	}
	if cardObject && len(words) == 1 {
		if costAmountAt(component, words[0], atoms, false) {
			component.ObjectNoun = ObjectNounCard
			component.ObjectIsCard = true
		}
		return
	}
	if len(words) != 2 || !costAmountAt(component, words[0], atoms, false) {
		return
	}
	noun, ok := atoms.ObjectNounAt(words[1].Span)
	if !ok {
		return
	}
	if cardObject && noun == ObjectNounPermanent {
		return
	}
	if cardObject {
		component.ObjectNoun = noun
		component.ObjectIsCard = true
		return
	}
	annotateCostObjectNoun(component, noun)
}

func costCardNoun(token shared.Token, atoms Atoms) bool {
	noun, ok := atoms.ObjectNounAt(token.Span)
	return ok && noun == ObjectNounCard
}

func costSelfReference(tokens []shared.Token, atoms Atoms, allowIt bool) bool {
	if len(tokens) == 0 {
		return false
	}
	span := shared.SpanOf(tokens)
	for _, reference := range atoms.ReferencesIn(span) {
		if reference.Span != span {
			continue
		}
		switch reference.Kind {
		case ReferenceSelfName:
			return true
		case ReferencePronoun:
			if allowIt && len(tokens) == 1 && equalWord(tokens[0], "it") {
				return true
			}
		case ReferenceThisObject:
			if len(tokens) != 2 {
				continue
			}
			noun, ok := atoms.ObjectNounAt(tokens[1].Span)
			if !ok {
				continue
			}
			switch noun {
			case ObjectNounArtifact, ObjectNounCreature, ObjectNounEnchantment,
				ObjectNounLand, ObjectNounPermanent, ObjectNounToken:
				return true
			default:
			}
		default:
		}
	}
	return false
}

func annotateMillCostObject(component *CostComponent, object []shared.Token, atoms Atoms) {
	if len(object) != 2 || !costAmountAt(component, object[0], atoms, false) || !costCardNoun(object[1], atoms) {
		return
	}
	component.ObjectNoun = ObjectNounCard
	component.ObjectIsCard = true
}

func annotatePutCounterCostObject(component *CostComponent, object []shared.Token, atoms Atoms) {
	counterIndex := singleCounterWordIndex(object)
	if counterIndex <= 1 || counterIndex+2 >= len(object) || !equalWord(object[counterIndex+1], "on") {
		return
	}
	if !costAmountAt(component, object[0], atoms, false) ||
		!costSelfReference(object[counterIndex+2:], atoms, true) {
		return
	}
	kind, ok := exactCostCounterKind(object[1:counterIndex], atoms, putCounterCostKinds())
	if !ok {
		return
	}
	component.CounterKind = kind
	component.CounterKindKnown = true
	component.SourceSelf = true
}

func annotateRemoveCounterCostObject(component *CostComponent, object []shared.Token, atoms Atoms) {
	counterIndex := singleCounterWordIndex(object)
	if counterIndex <= 1 || counterIndex+2 >= len(object) || !equalWord(object[counterIndex+1], "from") {
		return
	}
	if !costAmountAt(component, object[0], atoms, false) ||
		!costSelfReference(object[counterIndex+2:], atoms, true) {
		return
	}
	kind, ok := exactCostCounterKind(object[1:counterIndex], atoms, removeCounterCostKinds())
	if !ok {
		return
	}
	component.CounterKind = kind
	component.CounterKindKnown = true
	component.SourceSelf = true
}

func singleCounterWordIndex(tokens []shared.Token) int {
	index := -1
	for i, token := range tokens {
		if !equalWord(token, "counter") && !equalWord(token, "counters") {
			continue
		}
		if index >= 0 {
			return -1
		}
		index = i
	}
	return index
}

func exactCostCounterKind(tokens []shared.Token, atoms Atoms, allowed []counter.Kind) (counter.Kind, bool) {
	if len(tokens) == 0 {
		return 0, false
	}
	kind, span, ok := atoms.CounterIn(shared.SpanOf(tokens))
	if !ok || span != shared.SpanOf(tokens) || !slices.Contains(allowed, kind) {
		return 0, false
	}
	return kind, true
}

func putCounterCostKinds() []counter.Kind {
	return []counter.Kind{counter.PlusOnePlusOne, counter.MinusOneMinusOne, counter.Charge, counter.Verse, counter.Blood}
}

func removeCounterCostKinds() []counter.Kind {
	return []counter.Kind{
		counter.PlusOnePlusOne, counter.MinusOneMinusOne, counter.Loyalty, counter.Charge,
		counter.Time, counter.Defense, counter.Lore, counter.Verse, counter.Shield,
		counter.Stun, counter.Finality, counter.Brick, counter.Page, counter.Enlightened,
		counter.Oil, counter.Blood, counter.Indestructible, counter.Deathtouch,
		counter.Flying, counter.FirstStrike, counter.Hexproof, counter.Lifelink,
		counter.Menace, counter.Reach, counter.Trample, counter.Vigilance,
	}
}

func annotateRevealCostObject(component *CostComponent, object []shared.Token, atoms Atoms) {
	if len(object) >= 4 &&
		equalWord(object[len(object)-4], "that") &&
		equalWord(object[len(object)-3], "share") &&
		equalWord(object[len(object)-2], "a") &&
		equalWord(object[len(object)-1], "color") {
		object = object[:len(object)-4]
	}
	if len(object) < 5 ||
		!equalWord(object[len(object)-3], "from") ||
		!equalWord(object[len(object)-2], "your") ||
		!equalWord(object[len(object)-1], "hand") {
		return
	}
	if z, ok := atoms.ZoneIn(shared.SpanOf(object[len(object)-3:]), ZoneRoleFrom); !ok || z != zone.Hand {
		return
	}
	prefix := object[:len(object)-3]
	if len(prefix) < 2 || len(prefix) > 3 || !costAmountAt(component, prefix[0], atoms, true) {
		return
	}
	if len(prefix) == 3 {
		colorAtom, ok := atoms.ColorAt(prefix[1].Span)
		if !ok || colorAtom == ColorUnknown {
			return
		}
		component.ObjectColor = colorAtom
		component.ObjectColorKnown = true
		prefix = append(prefix[:1], prefix[2])
	}
	if !costCardNoun(prefix[1], atoms) {
		return
	}
	component.ObjectNoun = ObjectNounCard
	component.ObjectIsCard = true
	component.SourceZone = zone.Hand
}

func annotateReturnCostObject(component *CostComponent, object []shared.Token, atoms Atoms) {
	if len(object) < 6 ||
		!equalWord(object[len(object)-6], "you") ||
		!equalWord(object[len(object)-5], "control") ||
		!equalWord(object[len(object)-4], "to") ||
		!strings.EqualFold(object[len(object)-2].Text, "owner's") ||
		!equalWord(object[len(object)-1], "hand") {
		return
	}
	pronoun, ok := atoms.PronounAt(object[len(object)-3].Span)
	if !ok || pronoun != PronounIts && pronoun != PronounTheir {
		return
	}
	if z, ok := atoms.ZoneIn(shared.SpanOf(object[len(object)-4:]), ZoneRoleTo); !ok || z != zone.Hand {
		return
	}
	prefix := object[:len(object)-6]
	if len(prefix) < 2 || !costAmountAt(component, prefix[0], atoms, false) {
		return
	}
	prefix = prefix[1:]
	if len(prefix) > 0 && equalWord(prefix[0], "tapped") {
		component.RequireTapped = true
		prefix = prefix[1:]
	}
	if annotateCostPermanentObject(component, prefix, atoms, true, []types.Card{types.Land, types.Creature, types.Artifact, types.Enchantment}) {
		component.ObjectController = ControllerRelationYouControl
		component.ToZone = zone.Hand
	}
}

func annotateTapPermanentsCostObject(component *CostComponent, object []shared.Token, atoms Atoms) {
	if len(object) < 5 ||
		!costAmountAt(component, object[0], atoms, false) ||
		!equalWord(object[1], "untapped") ||
		!equalWord(object[len(object)-2], "you") ||
		!equalWord(object[len(object)-1], "control") {
		return
	}
	if annotateCostPermanentObject(component, object[2:len(object)-2], atoms, false, []types.Card{types.Creature, types.Artifact}) {
		component.RequireUntapped = true
		component.ObjectController = ControllerRelationYouControl
	}
}

func annotateCostPermanentObject(component *CostComponent, object []shared.Token, atoms Atoms, allowSnowLand bool, subtypeTypes []types.Card) bool {
	if len(object) == 0 {
		return false
	}
	if allowSnowLand && len(object) == 2 && equalWord(object[0], "snow") {
		noun, ok := atoms.ObjectNounAt(object[1].Span)
		supertype, superOK := atoms.SupertypeAt(object[0].Span)
		if !ok || noun != ObjectNounLand || !superOK || supertype != SupertypeSnow {
			return false
		}
		component.ObjectNoun = ObjectNounLand
		component.ObjectSupertype = types.Snow
		component.SupertypeKnown = true
		return true
	}
	if len(object) == 1 {
		if noun, ok := atoms.ObjectNounAt(object[0].Span); ok {
			return annotateCostObjectNoun(component, noun)
		}
		if sub, ok := atoms.SubtypeAt(object[0].Span); ok {
			if !SubtypeMatchesAnyRuntimeCardType(sub, subtypeTypes) {
				return false
			}
			component.SubtypesAny = []types.Sub{sub}
			return true
		}
	}
	return false
}

func annotateExileCostObject(component *CostComponent, object []shared.Token, atoms Atoms) {
	if len(object) < 5 ||
		!equalWord(object[len(object)-3], "from") ||
		!equalWord(object[len(object)-2], "your") ||
		!equalWord(object[len(object)-1], "graveyard") {
		return
	}
	if z, ok := atoms.ZoneIn(shared.SpanOf(object[len(object)-3:]), ZoneRoleFrom); !ok || z != zone.Graveyard {
		return
	}
	// Record the graveyard source zone as soon as the "from your graveyard"
	// suffix is confirmed: the compiler reads this typed zone to decide an
	// activated ability's source zone even when the exiled object is the source
	// itself (e.g. "Exile this card from your graveyard"), which the object
	// recognition below does not classify as a card selector.
	component.SourceZone = zone.Graveyard
	prefix := object[:len(object)-3]
	switch {
	case len(prefix) == 2 && exileCardAmount(component, prefix[0], atoms) && costCardNoun(prefix[1], atoms):
		component.ObjectNoun = ObjectNounCard
		component.ObjectIsCard = true
	case len(prefix) == 3 && exileTypedCardAmount(component, prefix[0]) && costCardNoun(prefix[2], atoms):
		noun, ok := atoms.ObjectNounAt(prefix[1].Span)
		if !ok || !costCardTypeNounAccepted(noun) {
			return
		}
		component.ObjectNoun = noun
		component.ObjectIsCard = true
	default:
		return
	}
}

// costCardTypeNounAccepted reports whether a noun names a card type the cost
// grammar can carry as a typed card-object type.
func costCardTypeNounAccepted(noun ObjectNoun) bool {
	switch noun {
	case ObjectNounArtifact, ObjectNounCreature, ObjectNounEnchantment, ObjectNounLand:
		return true
	default:
		return false
	}
}

func exileCardAmount(component *CostComponent, token shared.Token, atoms Atoms) bool {
	if equalWord(token, "a") || equalWord(token, "an") {
		component.AmountValue = 1
		component.AmountKnown = true
		return true
	}
	if value, ok := atoms.CardinalAt(token.Span); ok && value == 2 {
		component.AmountValue = 2
		component.AmountKnown = true
		return true
	}
	return false
}

func exileTypedCardAmount(component *CostComponent, token shared.Token) bool {
	if !equalWord(token, "a") && !equalWord(token, "an") {
		return false
	}
	component.AmountValue = 1
	component.AmountKnown = true
	return true
}

// signedLoyaltyAmount converts a fixed loyalty cost amount such as "+1", "−2"
// (Unicode minus U+2212), or "0" into a signed integer. It reports false for
// variable costs (e.g. "+X") and malformed input. The loyalty cost sign is a
// literal semantic value of the cost grammar.
func signedLoyaltyAmount(amount string) (int, bool) {
	if amount == "" {
		return 0, false
	}
	rest := amount
	sign := 1
	switch {
	case strings.HasPrefix(rest, "+"):
		rest = rest[1:]
	case strings.HasPrefix(rest, "\u2212"):
		sign = -1
		rest = rest[len("\u2212"):]
	case strings.HasPrefix(rest, "-"):
		sign = -1
		rest = rest[1:]
	default:
	}
	n, err := strconv.Atoi(rest)
	if err != nil {
		return 0, false
	}
	return sign * n, true
}

// isVariableLoyaltyAmount reports whether a loyalty cost amount is a variable
// cost such as "+X", "−X", or "X".
func isVariableLoyaltyAmount(amount string) bool {
	rest := amount
	rest = strings.TrimPrefix(rest, "+")
	rest = strings.TrimPrefix(rest, "\u2212")
	rest = strings.TrimPrefix(rest, "-")
	return strings.EqualFold(rest, "X")
}

func costJoinedTokenText(tokens []shared.Token) string {
	var builder strings.Builder
	for _, token := range tokens {
		_, _ = builder.WriteString(token.Text)
	}
	return builder.String()
}

func costRelativeSpan(span shared.Span, base int) shared.Span {
	span.Start.Offset -= base
	span.End.Offset -= base
	return span
}

func wordsAfterFirst(tokens []shared.Token) string {
	if len(tokens) < 2 {
		return ""
	}
	return costJoinedSourceText(tokens[1:])
}

func costJoinedSourceText(tokens []shared.Token) string {
	if len(tokens) == 0 {
		return ""
	}
	var builder strings.Builder
	for i, token := range tokens {
		if i > 0 && costNeedsSemanticSpace(tokens[i-1], token) {
			_ = builder.WriteByte(' ')
		}
		_, _ = builder.WriteString(token.Text)
	}
	return builder.String()
}

func costNeedsSemanticSpace(previous, current shared.Token) bool {
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

func firstInteger(tokens []shared.Token) string {
	for _, token := range tokens {
		if token.Kind == shared.Integer {
			return token.Text
		}
	}
	return ""
}

func positiveIntegerWord(word string) bool {
	amount, err := strconv.Atoi(word)
	return err == nil && amount > 0
}

func startsWords(words []string, expected ...string) bool {
	if len(words) < len(expected) {
		return false
	}
	for i := range expected {
		if words[i] != expected[i] {
			return false
		}
	}
	return true
}

func containsNoun(words []string, singular string) bool {
	return slices.Contains(words, singular) || slices.Contains(words, singular+"s")
}

func allSymbols(tokens []shared.Token) bool {
	if len(tokens) == 0 {
		return false
	}
	for _, token := range tokens {
		if token.Kind != shared.Symbol {
			return false
		}
	}
	return true
}

func allEnergySymbols(tokens []shared.Token) bool {
	if len(tokens) == 0 {
		return false
	}
	for _, token := range tokens {
		if token.Kind != shared.Symbol || !strings.EqualFold(token.Text, "{E}") {
			return false
		}
	}
	return true
}
