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

// PayLifeDynamicAmount identifies a recognized rules-derived amount for a "pay
// life equal to ..." cost. The parser owns the Oracle wording; downstream
// layers consume the typed value.
type PayLifeDynamicAmount uint8

// Pay-life dynamic amounts recognized by the parser.
const (
	PayLifeDynamicAmountNone PayLifeDynamicAmount = iota
	// PayLifeDynamicCommanderColorIdentityCount recognizes "the number of
	// colors in your commanders' color identity" (War Room).
	PayLifeDynamicCommanderColorIdentityCount
	// PayLifeDynamicLifeGainedThisTurn recognizes "X, where X is the amount of
	// life you gained this turn" as the value of a "pay X life" cost (Tivash,
	// Gloom Summoner).
	PayLifeDynamicLifeGainedThisTurn
)

// Cost is the ordered, source-spanned typed cost the parser recognizes before an
// activated or loyalty ability's colon.
type Cost struct {
	Span       shared.Span     `json:"-"`
	Text       string          `json:",omitempty"`
	Components []CostComponent `json:",omitempty"`
	// Order is the cost phrase's dense source-order rank, used downstream to
	// test reference containment without byte offsets.
	Order shared.SourceOrder `json:"-"`
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

	// SecondObjectNoun is the second permanent-type noun of a two-type cost
	// union such as "sacrifice an artifact or creature." It is empty unless the
	// object names two permanent types joined by "or".
	SecondObjectNoun ObjectNoun `json:",omitempty"`

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

	// RemoveCounterAmong reports a "remove N counters from among <permanents>
	// you control" cost, where the removed counters are spread across the
	// chosen controlled permanents rather than taken from the ability's own
	// source. The object noun/controller fields carry the permanent constraint.
	RemoveCounterAmong bool `json:",omitempty"`

	// ExcludeSource reports that the cost object excludes the ability's own
	// source, recognized from the determiner "another" (e.g. "Sacrifice another
	// creature"). The compiler carries it onto the typed cost component.
	ExcludeSource bool `json:",omitempty"`

	// ChoiceGroup tags this component as one alternative of a printed "<cost> or
	// <cost>" choice (e.g. "sacrifice an artifact or discard a card"). Zero means
	// a mandatory standalone cost; components sharing a nonzero value are
	// alternatives of which exactly one is paid.
	ChoiceGroup uint8 `json:",omitempty"`

	// DiscardWholeHand reports a "discard your hand" cost object, where the
	// payer discards every card in their hand rather than a fixed count. The
	// compiler carries it onto the typed cost component.
	DiscardWholeHand bool `json:",omitempty"`

	// PayLifeDynamic names a recognized rules-derived amount for a "pay life
	// equal to ..." cost whose value is neither a fixed integer nor X. The
	// compiler maps it onto its typed dynamic-amount vocabulary.
	PayLifeDynamic PayLifeDynamicAmount `json:",omitempty"`

	// Order is the component's dense source-order rank, used downstream to test
	// reference containment without byte offsets.
	Order shared.SourceOrder `json:"-"`
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
	if abilityKind != AbilityLoyalty {
		if alternatives, ok := costChoiceAlternatives(parts); ok {
			for _, alternative := range alternatives {
				component := buildCostComponent(alternative, abilityKind, phrase, atoms)
				component.ChoiceGroup = 1
				cost.Components = append(cost.Components, component)
			}
			return cost
		}
	}
	for _, part := range parts {
		if len(part) == 0 {
			continue
		}
		cost.Components = append(cost.Components, buildCostComponent(part, abilityKind, phrase, atoms))
	}
	return cost
}

// buildCostComponent recognizes one cost operation's verb and typed object from
// a single comma- or choice-delimited token run.
func buildCostComponent(part []shared.Token, abilityKind AbilityKind, phrase Phrase, atoms Atoms) CostComponent {
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
	return component
}

// costVerbs lists the leading words the parser recognizes as cost operations.
// They distinguish a genuine "<cost> or <cost>" choice from a two-permanent-type
// union such as "sacrifice an artifact or creature", whose right side names a
// type rather than a verb.
var costVerbs = []string{
	"sacrifice", "discard", "pay", "reveal", "exile", "tap",
	"untap", "return", "mill", "put", "remove", "collect", "exert",
}

// costChoiceAlternatives detects an additional cost printed as a choice of
// alternatives joined by "or" and returns the alternatives' token runs. It
// recognizes the two-way form "<cost> or <cost>" (a single comma part) and the
// Oxford form "<cost>, <cost>, or <cost>" (comma parts whose last begins with
// "or"). It requires every alternative to begin with a recognized cost verb so a
// two-type union like "sacrifice an artifact or creature" is not split.
func costChoiceAlternatives(parts [][]shared.Token) ([][]shared.Token, bool) {
	if len(parts) >= 2 {
		last := parts[len(parts)-1]
		if rest, ok := stripLeadingOr(last); ok {
			alternatives := make([][]shared.Token, 0, len(parts))
			alternatives = append(alternatives, parts[:len(parts)-1]...)
			alternatives = append(alternatives, rest)
			if allCostVerbLed(alternatives) {
				return alternatives, true
			}
		}
		return nil, false
	}
	if len(parts) == 1 {
		segments := splitTopLevelWord(parts[0], "or")
		if len(segments) == 2 && allCostVerbLed(segments) {
			return segments, true
		}
	}
	return nil, false
}

// stripLeadingOr drops a leading "or" word, reporting whether one was present.
func stripLeadingOr(tokens []shared.Token) ([]shared.Token, bool) {
	if len(tokens) == 0 || !equalWord(tokens[0], "or") {
		return nil, false
	}
	return tokens[1:], true
}

// allCostVerbLed reports that every alternative is non-empty and begins with a
// recognized cost verb.
func allCostVerbLed(alternatives [][]shared.Token) bool {
	for _, alternative := range alternatives {
		if !costVerbLed(alternative) {
			return false
		}
	}
	return len(alternatives) > 0
}

// costVerbLed reports that the token run begins with a recognized cost verb.
func costVerbLed(tokens []shared.Token) bool {
	words := normalizedWords(tokens)
	return len(words) > 0 && slices.Contains(costVerbs, words[0])
}

// splitTopLevelWord splits tokens on each occurrence of the given lowercase word
// at parenthesis/quote depth zero.
func splitTopLevelWord(tokens []shared.Token, word string) [][]shared.Token {
	var parts [][]shared.Token
	start := 0
	depth := 0
	quoted := false
	for i, token := range tokens {
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
			if depth == 0 && !quoted && equalWord(token, word) {
				parts = append(parts, tokens[start:i])
				start = i + 1
			}
		}
	}
	return append(parts, tokens[start:])
}

func parseCostAtoms(component *CostComponent, tokens []shared.Token, atoms Atoms) {
	if len(tokens) == 0 {
		return
	}
	object := tokens[1:]
	switch component.Kind {
	case CostComponentPayLife:
		annotatePayLifeCostAmount(component, tokens, atoms)
	case CostComponentCollectEvidence:
		annotateIntegerCostAmount(component)
	case CostComponentEnergy:
		component.AmountValue = len(tokens) - 1
		component.AmountKnown = component.AmountValue > 0
	case CostComponentSacrifice:
		if costSelfReference(object, atoms, false) {
			component.SourceSelf = true
			return
		}
		annotateSacrificeCostObject(component, object, atoms)
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

func annotatePayLifeCostAmount(component *CostComponent, tokens []shared.Token, atoms Atoms) {
	if len(tokens) == 3 && equalWord(tokens[0], "pay") && equalWord(tokens[2], "life") &&
		costAmountAt(component, tokens[1], atoms, true) {
		return
	}
	if startsWords(normalizedWords(tokens),
		"pay", "life", "equal", "to", "the", "number", "of",
		"colors", "in", "your", "commanders", "color", "identity") {
		component.PayLifeDynamic = PayLifeDynamicCommanderColorIdentityCount
		return
	}
	annotateIntegerCostAmount(component)
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
	if component.Kind == CostComponentDiscard && discardWholeHandObject(object) {
		component.DiscardWholeHand = true
		component.SourceZone = zone.Hand
		return
	}
	if costSelfReference(object, atoms, false) {
		component.AmountKnown = true
		component.AmountValue = 1
		component.ObjectNoun = ObjectNounCard
		component.ObjectIsCard = true
		component.SourceSelf = true
		if component.Kind == CostComponentDiscard {
			component.SourceZone = zone.Hand
		}
		return
	}
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

// discardWholeHandObject reports whether a discard cost object names the whole
// hand ("discard your hand"), which the payer satisfies by discarding every
// card in hand rather than a fixed count.
func discardWholeHandObject(object []shared.Token) bool {
	return len(object) == 2 && equalWord(object[0], "your") && equalWord(object[1], "hand")
}

// costSelfExileNoun reports whether a noun token following "this" names the
// source object for a self-exile cost ("Exile this card/creature from your
// hand"). It accepts the generic "card" noun and the permanent card-type nouns
// so single-type self-exilers (e.g. "Exile this creature from your hand") match
// the same source-exile cost as the generic "this card" wording.
func costSelfExileNoun(token shared.Token, atoms Atoms) bool {
	noun, ok := atoms.ObjectNounAt(token.Span)
	if !ok {
		return false
	}
	switch noun {
	case ObjectNounCard, ObjectNounCreature, ObjectNounArtifact,
		ObjectNounEnchantment, ObjectNounLand, ObjectNounPermanent:
		return true
	default:
		return false
	}
}

// sacrificeSubtypeFamilies lists the permanent card types whose subtypes a
// sacrifice cost object may name, e.g. "Sacrifice a Goblin" or "Sacrifice three
// Treasures". A named subtype must belong to one of these families to be
// recognized as a sacrifice constraint.
var sacrificeSubtypeFamilies = []types.Card{
	types.Artifact,
	types.Battle,
	types.Creature,
	types.Enchantment,
	types.Land,
	types.Planeswalker,
}

// annotateSacrificeCostObject recognizes a battlefield sacrifice cost object.
// It accepts an amount word (or the determiner "another", which means one and
// excludes the source) followed by a single object noun or permanent subtype,
// with an optional "you control" suffix. Unrecognized objects leave the
// component bare so lowering fails closed.
// costSelfSubtypeNoun reports whether token names the source by its own
// permanent subtype (e.g. "Aura" in "Sacrifice this Aura") or by the Equipment
// noun, which costSelfReference does not classify as a self marker. A card that
// names "this <its own subtype>" in a cost always refers to its own source.
func costSelfSubtypeNoun(token shared.Token, atoms Atoms) bool {
	if noun, ok := atoms.ObjectNounAt(token.Span); ok {
		return noun == ObjectNounEquipment
	}
	_, ok := atoms.SubtypeAt(token.Span)
	return ok
}

func annotateSacrificeCostObject(component *CostComponent, object []shared.Token, atoms Atoms) {
	words := object
	if len(words) == 2 && equalWord(words[0], "this") && costSelfSubtypeNoun(words[1], atoms) {
		component.SourceSelf = true
		return
	}
	if len(words) >= 2 && equalWord(words[0], "another") {
		component.ExcludeSource = true
		component.AmountValue = 1
		component.AmountKnown = true
		words = words[1:]
	} else {
		if len(words) < 2 || !costAmountAt(component, words[0], atoms, false) {
			return
		}
		words = words[1:]
		// A count followed by "other" excludes the source, e.g. "Sacrifice two
		// other creatures." It carries the announced amount, unlike "another."
		if len(words) >= 2 && equalWord(words[0], "other") {
			component.ExcludeSource = true
			words = words[1:]
		}
	}
	if len(words) >= 2 && equalWord(words[len(words)-2], "you") && equalWord(words[len(words)-1], "control") {
		words = words[:len(words)-2]
	}
	if len(words) >= 2 {
		if colorAtom, ok := atoms.ColorAt(words[0].Span); ok && colorAtom != ColorUnknown {
			component.ObjectColor = colorAtom
			component.ObjectColorKnown = true
			words = words[1:]
		}
	}
	if first, second, ok := costTwoTypeUnionNouns(words); ok {
		if annotateCostTwoTypeUnionObject(component, first, second, atoms) {
			return
		}
		if annotateCostTwoSubtypeUnion(component, first, second, atoms, sacrificeSubtypeFamilies) {
			return
		}
		clearSacrificeCostObject(component)
		return
	}
	if annotateCostSubtypeWithNoun(component, words, atoms, sacrificeSubtypeFamilies) {
		return
	}
	if len(words) != 1 {
		clearSacrificeCostObject(component)
		return
	}
	if noun, ok := atoms.ObjectNounAt(words[0].Span); ok {
		if !annotateCostObjectNoun(component, noun) {
			clearSacrificeCostObject(component)
		}
		return
	}
	if sub, ok := atoms.SubtypeAt(words[0].Span); ok &&
		SubtypeMatchesAnyRuntimeCardType(sub, sacrificeSubtypeFamilies) {
		component.SubtypesAny = []types.Sub{sub}
		return
	}
	clearSacrificeCostObject(component)
}

// annotateCostTwoSubtypeUnion recognizes a cost object that names two
// alternative permanent subtypes joined by "or" (e.g. "Orc or Goblin", "Forest
// or Plains"). Each subtype must be defined for one of the supplied permanent
// families. The two subtypes lower to a SubtypesAny set matched with OR
// semantics by the rules. Anything else leaves the object bare so lowering
// fails closed.
func annotateCostTwoSubtypeUnion(component *CostComponent, first, second shared.Token, atoms Atoms, families []types.Card) bool {
	firstSub, ok := atoms.SubtypeAt(first.Span)
	if !ok || !SubtypeMatchesAnyRuntimeCardType(firstSub, families) {
		return false
	}
	secondSub, ok := atoms.SubtypeAt(second.Span)
	if !ok || secondSub == firstSub || !SubtypeMatchesAnyRuntimeCardType(secondSub, families) {
		return false
	}
	component.SubtypesAny = []types.Sub{firstSub, secondSub}
	return true
}

// annotateCostSubtypeWithNoun recognizes a cost object that names a permanent
// subtype followed by a permanent-type noun, e.g. "Goblin creature" or "Blood
// token." The subtype must be defined for the noun's card type (or, for the
// generic "permanent"/"token" nouns, for any of the supplied families). The
// trailing noun is descriptive; the subtype alone constrains the cost.
func annotateCostSubtypeWithNoun(component *CostComponent, words []shared.Token, atoms Atoms, families []types.Card) bool {
	if len(words) != 2 {
		return false
	}
	sub, ok := atoms.SubtypeAt(words[0].Span)
	if !ok {
		return false
	}
	noun, ok := atoms.ObjectNounAt(words[1].Span)
	if !ok || !costSubtypeMatchesNoun(sub, noun, families) {
		return false
	}
	component.SubtypesAny = []types.Sub{sub}
	return true
}

// costSubtypeMatchesNoun reports whether a subtype belongs to the card type
// named by a permanent-type noun. The generic "permanent" and "token" nouns
// carry no single type, so the subtype need only be defined for one of the
// supplied families.
func costSubtypeMatchesNoun(sub types.Sub, noun ObjectNoun, families []types.Card) bool {
	switch noun {
	case ObjectNounArtifact:
		return types.KnownSubtypeForType(types.Artifact, sub)
	case ObjectNounCreature:
		return types.KnownSubtypeForType(types.Creature, sub)
	case ObjectNounEnchantment:
		return types.KnownSubtypeForType(types.Enchantment, sub)
	case ObjectNounLand:
		return types.KnownSubtypeForType(types.Land, sub)
	case ObjectNounPermanent, ObjectNounToken:
		return SubtypeMatchesAnyRuntimeCardType(sub, families)
	default:
		return false
	}
}

// clearSacrificeCostObject resets the amount, source-exclusion, and color
// fields a sacrifice cost object set before its noun failed recognition, so an
// unrecognized object never lowers to a partial cost.
func clearSacrificeCostObject(component *CostComponent) {
	component.AmountValue = 0
	component.AmountKnown = false
	component.ExcludeSource = false
	component.ObjectColor = ColorUnknown
	component.ObjectColorKnown = false
}

// annotateCostTwoTypeUnionObject recognizes a cost object that names two
// alternative permanent types joined by "or" (e.g. "an artifact or creature",
// "a creature or planeswalker"). The first noun must be a type the permanent
// selector maps directly; the second noun may also be planeswalker. Anything
// else leaves the object bare so lowering fails closed. It is shared by the
// sacrifice and tap-permanents cost grammars, whose unions lower identically.
func annotateCostTwoTypeUnionObject(component *CostComponent, first, second shared.Token, atoms Atoms) bool {
	firstNoun, ok := atoms.ObjectNounAt(first.Span)
	if !ok || !costSelectorPermanentTypeNoun(firstNoun) {
		return false
	}
	secondNoun, ok := atoms.ObjectNounAt(second.Span)
	if !ok || !costUnionPermanentTypeNoun(secondNoun) || secondNoun == firstNoun {
		return false
	}
	component.ObjectNoun = firstNoun
	component.SecondObjectNoun = secondNoun
	return true
}

// costTwoTypeUnionNouns recognizes a two-type cost union's noun tokens. It
// accepts "A or B" and "A and/or B" (the latter lexed as "and", "/", "or"),
// each with an optional article before the second noun (e.g. "creature or an
// enchantment"). It returns the first and second noun tokens and whether the
// tokens form such a union. Type recognition is left to the caller.
func costTwoTypeUnionNouns(tokens []shared.Token) (first, second shared.Token, ok bool) {
	if len(tokens) < 3 {
		return shared.Token{}, shared.Token{}, false
	}
	first = tokens[0]
	rest := tokens[1:]
	switch {
	case equalWord(rest[0], "or"):
		rest = rest[1:]
	case len(rest) >= 3 && equalWord(rest[0], "and") && rest[1].Kind == shared.Slash && equalWord(rest[2], "or"):
		rest = rest[3:]
	default:
		return shared.Token{}, shared.Token{}, false
	}
	if len(rest) == 2 && (equalWord(rest[0], "a") || equalWord(rest[0], "an")) {
		rest = rest[1:]
	}
	if len(rest) != 1 {
		return shared.Token{}, shared.Token{}, false
	}
	return first, rest[0], true
}

// costSelectorPermanentTypeNoun reports whether a noun names a permanent card
// type the compiler maps onto a permanent selector with a card type (artifact,
// creature, enchantment, land). It is the set valid as the first member of a
// two-type cost union, where the selector and primary type are derived from the
// noun.
func costSelectorPermanentTypeNoun(noun ObjectNoun) bool {
	switch noun {
	case ObjectNounArtifact, ObjectNounCreature, ObjectNounEnchantment, ObjectNounLand:
		return true
	default:
		return false
	}
}

// costUnionPermanentTypeNoun reports whether a noun names a permanent card type
// valid as the second member of a two-type cost union. It extends the selector
// set with planeswalker, which only contributes an alternative card type and
// needs no selector. The bare "permanent" and "card" nouns are excluded because
// a union with them carries no extra constraint.
func costUnionPermanentTypeNoun(noun ObjectNoun) bool {
	return costSelectorPermanentTypeNoun(noun) || noun == ObjectNounPlaneswalker
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
				ObjectNounLand, ObjectNounPermanent, ObjectNounToken, ObjectNounCard:
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
	if counterIndex < 1 || counterIndex+2 >= len(object) || !equalWord(object[counterIndex+1], "from") {
		return
	}
	kindTokens := object[1:counterIndex]
	rest := object[counterIndex+2:]
	if annotateRemoveCounterAmongObject(component, object[0], kindTokens, rest, atoms) {
		return
	}
	if annotateRemoveCounterPermanentObject(component, object[0], kindTokens, rest, atoms) {
		return
	}
	annotateRemoveCounterSourceObject(component, object[0], kindTokens, rest, atoms)
}

// annotateRemoveCounterSourceObject recognizes the single-source cost "Remove N
// <kind> counters from <this permanent>", which removes the counters from the
// ability's own source. It requires an explicit counter kind.
func annotateRemoveCounterSourceObject(component *CostComponent, amount shared.Token, kindTokens, rest []shared.Token, atoms Atoms) {
	if len(kindTokens) == 0 {
		return
	}
	if !costAmountAt(component, amount, atoms, false) ||
		!costSelfReference(rest, atoms, true) {
		return
	}
	kind, ok := exactCostCounterKind(kindTokens, atoms, removeCounterCostKinds())
	if !ok {
		return
	}
	component.CounterKind = kind
	component.CounterKindKnown = true
	component.SourceSelf = true
}

// annotateRemoveCounterAmongObject recognizes the spread cost "Remove N <kind>
// counters from among <permanents> you control", where the removed counters are
// distributed across the chosen controlled permanents. The amount may be a fixed
// count or X; the counter kind, when named, must be a recognized kind.
func annotateRemoveCounterAmongObject(component *CostComponent, amount shared.Token, kindTokens, rest []shared.Token, atoms Atoms) bool {
	if len(rest) < 4 ||
		!equalWord(rest[0], "among") ||
		!equalWord(rest[len(rest)-2], "you") ||
		!equalWord(rest[len(rest)-1], "control") {
		return false
	}
	typeTokens := rest[1 : len(rest)-2]
	if !annotateCostPermanentObject(component, typeTokens, atoms, false, sacrificeSubtypeFamilies) {
		return false
	}
	if !costAmountAt(component, amount, atoms, true) {
		clearRemoveCounterAmongObject(component)
		return false
	}
	if len(kindTokens) > 0 {
		kind, ok := exactCostCounterKind(kindTokens, atoms, removeCounterCostKinds())
		if !ok {
			clearRemoveCounterAmongObject(component)
			return false
		}
		component.CounterKind = kind
		component.CounterKindKnown = true
	}
	component.ObjectController = ControllerRelationYouControl
	component.RemoveCounterAmong = true
	return true
}

// annotateRemoveCounterPermanentObject recognizes the single-permanent cost
// "Remove a <kind> counter from a permanent you control", which removes one
// counter from a single permanent the payer controls. It reuses the spread
// removal machinery with a fixed amount of one: a single-permanent removal is a
// spread removal that happens to draw its one counter from one chosen
// permanent. The amount must be a fixed count; the counter kind, when named,
// must be a recognized kind, and the permanent constraint comes from the named
// noun.
func annotateRemoveCounterPermanentObject(component *CostComponent, amount shared.Token, kindTokens, rest []shared.Token, atoms Atoms) bool {
	if len(rest) < 4 ||
		equalWord(rest[0], "among") ||
		!equalWord(rest[len(rest)-2], "you") ||
		!equalWord(rest[len(rest)-1], "control") {
		return false
	}
	body := rest[:len(rest)-2]
	if !equalWord(body[0], "a") && !equalWord(body[0], "an") {
		return false
	}
	typeTokens := body[1:]
	if !annotateCostPermanentObject(component, typeTokens, atoms, false, sacrificeSubtypeFamilies) {
		return false
	}
	if !costAmountAt(component, amount, atoms, false) || component.AmountValue != 1 {
		clearRemoveCounterAmongObject(component)
		return false
	}
	if len(kindTokens) > 0 {
		kind, ok := exactCostCounterKind(kindTokens, atoms, removeCounterCostKinds())
		if !ok {
			clearRemoveCounterAmongObject(component)
			return false
		}
		component.CounterKind = kind
		component.CounterKindKnown = true
	}
	component.ObjectController = ControllerRelationYouControl
	component.RemoveCounterAmong = true
	return true
}

// among-removal cost may have set so the component falls back to bare and
// lowering fails closed.
func clearRemoveCounterAmongObject(component *CostComponent) {
	component.ObjectNoun = ObjectNounUnknown
	component.ObjectController = ControllerRelationUnknown
	component.SubtypesAny = nil
	component.ObjectSupertype = ""
	component.SupertypeKnown = false
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
	middle := object[2 : len(object)-2]
	if first, second, ok := costTwoTypeUnionNouns(middle); ok {
		if annotateCostTwoTypeUnionObject(component, first, second, atoms) {
			component.RequireUntapped = true
			component.ObjectController = ControllerRelationYouControl
		}
		return
	}
	if annotateCostPermanentObject(component, middle, atoms, false, sacrificeSubtypeFamilies) {
		component.RequireUntapped = true
		component.ObjectController = ControllerRelationYouControl
	}
}

func annotateCostPermanentObject(component *CostComponent, object []shared.Token, atoms Atoms, allowSnowLand bool, subtypeTypes []types.Card) bool {
	if len(object) == 0 {
		return false
	}
	if len(object) == 2 {
		if supertype, ok := atoms.SupertypeAt(object[0].Span); ok {
			return annotateCostSupertypeObject(component, supertype, object[1], atoms, allowSnowLand)
		}
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

// annotateCostSupertypeObject recognizes a cost permanent object named by a
// supertype and noun, such as "legendary creature." Snow is recognized only for
// lands in snow-permitting contexts, preserving the established snow-land cost
// grammar; every other supertype constrains any object noun.
func annotateCostSupertypeObject(component *CostComponent, supertype Supertype, nounToken shared.Token, atoms Atoms, allowSnowLand bool) bool {
	runtimeSuper, ok := runtimeSupertype(supertype)
	if !ok {
		return false
	}
	noun, ok := atoms.ObjectNounAt(nounToken.Span)
	if !ok {
		return false
	}
	if supertype == SupertypeSnow && (!allowSnowLand || noun != ObjectNounLand) {
		return false
	}
	if !annotateCostObjectNoun(component, noun) {
		return false
	}
	component.ObjectSupertype = runtimeSuper
	component.SupertypeKnown = true
	return true
}

func annotateExileCostObject(component *CostComponent, object []shared.Token, atoms Atoms) {
	if len(object) < 5 ||
		!equalWord(object[len(object)-3], "from") ||
		!equalWord(object[len(object)-2], "your") {
		return
	}
	var sourceZone zone.Type
	switch {
	case equalWord(object[len(object)-1], "graveyard"):
		sourceZone = zone.Graveyard
	case equalWord(object[len(object)-1], "hand"):
		sourceZone = zone.Hand
	default:
		return
	}
	if z, ok := atoms.ZoneIn(shared.SpanOf(object[len(object)-3:]), ZoneRoleFrom); !ok || z != sourceZone {
		return
	}
	// Record the source zone as soon as the "from your <zone>" suffix is
	// confirmed: the compiler reads this typed zone to decide an activated
	// ability's source zone even when the exiled object is the source itself
	// (e.g. "Exile this card from your hand"), which the object recognition
	// below does not classify as a card selector.
	component.SourceZone = sourceZone
	prefix := object[:len(object)-3]
	switch {
	case len(prefix) == 2 && equalWord(prefix[0], "this") && costSelfExileNoun(prefix[1], atoms):
		// "Exile this card/creature from your hand or graveyard" exiles the
		// source card itself; the compiler routes this to an
		// AdditionalExileSource cost paid from the source zone recorded above.
		component.SourceSelf = true
	case len(prefix) == 2 && exileCardAmount(component, prefix[0], atoms) && costCardNoun(prefix[1], atoms):
		component.ObjectNoun = ObjectNounCard
		component.ObjectIsCard = true
	case len(prefix) == 3 && exileCardAmount(component, prefix[0], atoms) &&
		equalWord(prefix[1], "other") && costCardNoun(prefix[2], atoms):
		// "Exile N other cards from your graveyard" (Escape): "other" excludes
		// the escaping card itself, which is still in the graveyard while the
		// cost is being paid and must not be exiled to satisfy its own cost.
		component.ObjectNoun = ObjectNounCard
		component.ObjectIsCard = true
		component.ExcludeSource = true
	case len(prefix) == 3 && exileCardAmount(component, prefix[0], atoms) && costCardNoun(prefix[2], atoms):
		if noun, ok := atoms.ObjectNounAt(prefix[1].Span); ok {
			if !costCardTypeNounAccepted(noun) {
				return
			}
			component.ObjectNoun = noun
			component.ObjectIsCard = true
			return
		}
		if sub, ok := atoms.SubtypeAt(prefix[1].Span); ok &&
			SubtypeMatchesAnyRuntimeCardType(sub, exileCardSubtypeFamilies) {
			component.ObjectNoun = ObjectNounCard
			component.ObjectIsCard = true
			component.SubtypesAny = []types.Sub{sub}
			return
		}
		return
	default:
		return
	}
}

// exileCardSubtypeFamilies lists the card-type families whose subtypes a
// graveyard exile cost object may name, e.g. "Exile an Elf card from your
// graveyard." A named subtype must belong to one of these families to be
// recognized as an exile constraint.
var exileCardSubtypeFamilies = []types.Card{
	types.Artifact,
	types.Battle,
	types.Creature,
	types.Enchantment,
	types.Instant,
	types.Land,
	types.Planeswalker,
	types.Sorcery,
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
	if equalWord(token, "x") {
		component.AmountFromX = true
		return true
	}
	if value, ok := atoms.CardinalAt(token.Span); ok && value > 0 {
		component.AmountValue = value
		component.AmountKnown = true
		return true
	}
	return false
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
