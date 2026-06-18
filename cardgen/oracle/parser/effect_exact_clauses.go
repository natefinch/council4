package parser

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/zone"
)

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

// exactGraveyardCardTargetSyntax reports whether target is a graveyard-card
// target the executable backend can lower exactly. It reconstructs the canonical
// Oracle noun phrase ("Return target <noun> from <owner> graveyard") byte-for-
// byte from the selection's typed fields and compares it to target.Text,
// accepting a single card-type, a union of card types ("creature or enchantment
// card"), a permanent card, a single color, a colorless or multicolored card, a
// single subtype, or the plain "card" noun, with an optional "with mana value N
// or less" qualifier and an optional multi-target or "up to N" count. It fails
// closed for every qualifier the canonical phrasing cannot render (power,
// toughness, keyword, supertype, excluded types or colors, combinations), so an
// unrepresentable target keeps failing rather than lowering to a wrong predicate.
func exactGraveyardCardTargetSyntax(target TargetSyntax) bool {
	sel := target.Selection
	if sel.Zone != zone.Graveyard || sel.Other {
		return false
	}
	if target.Cardinality == (TargetCardinalitySyntax{Min: 0, Max: 2}) &&
		strings.EqualFold(target.Text, "up to two target cards with cycling from your graveyard") {
		return true
	}
	if sel.All || sel.Attacking || sel.Blocking || sel.Tapped || sel.Untapped ||
		sel.Keyword != KeywordUnknown || sel.MatchPower || sel.MatchToughness ||
		len(sel.ExcludedTypes) != 0 || len(sel.SourceTypes) != 0 ||
		len(sel.Supertypes) != 0 || len(sel.ExcludedColors) != 0 {
		return false
	}
	owner, ok := graveyardOwnerSuffix(sel.Controller)
	if !ok {
		return false
	}
	prefix, plural, ok := graveyardCardCardinalityPrefix(target.Cardinality, sel.Another)
	if !ok {
		return false
	}
	noun, ok := graveyardCardNoun(sel)
	if !ok {
		return false
	}
	if plural {
		noun += "s"
	}
	manaClause := ""
	if sel.MatchManaValue {
		// The canonical mana-value qualifier follows the singular noun; the
		// multi-target plural noun never carries it in printed Oracle wording.
		if plural {
			return false
		}
		clause, ok := graveyardManaValueClause(sel.ManaValue)
		if !ok {
			return false
		}
		manaClause = clause
	}
	return strings.EqualFold(target.Text, prefix+noun+manaClause+" "+owner)
}

// graveyardOwnerSuffix renders the canonical "from <owner> graveyard" clause for
// a graveyard-card target's controller relation: "your" for the controller,
// "a" for any graveyard, and "an opponent's" for an opponent. It fails closed for
// the "you don't control" relation, which has no graveyard-owner phrasing.
func graveyardOwnerSuffix(controller SelectionController) (string, bool) {
	switch controller {
	case SelectionControllerYou:
		return "from your graveyard", true
	case SelectionControllerAny:
		return "from a graveyard", true
	case SelectionControllerOpponent:
		return "from an opponent's graveyard", true
	default:
		return "", false
	}
}

// graveyardCardCardinalityPrefix returns the canonical count words preceding the
// graveyard-card noun, whether that noun is plural, and whether the cardinality
// is one the round-trip represents. Single targets render "target " (or
// "another target " for a self-exclusion); optional and multi-target counts reuse
// multiTargetCardinalityPrefix for "up to one ", "up to <N> ", and "<N> ". It
// fails closed for a self-exclusion combined with a multi-target count, which has
// no canonical phrasing.
func graveyardCardCardinalityPrefix(c TargetCardinalitySyntax, another bool) (prefix string, plural, ok bool) {
	if c == (TargetCardinalitySyntax{Min: 1, Max: 1}) {
		if another {
			return "another target ", false, true
		}
		return "target ", false, true
	}
	if another {
		return "", false, false
	}
	countPrefix, plural, ok := multiTargetCardinalityPrefix(c)
	if !ok {
		return "", false, false
	}
	return countPrefix + "target ", plural, true
}

// graveyardCardNoun reconstructs the singular graveyard-card noun ("creature
// card", "creature or enchantment card", "permanent card", "green card",
// "multicolored card", "colorless card", "Zombie card", or the plain "card")
// from selection's typed fields. It accepts exactly one restriction category so
// combinations it could not render in canonical order fail closed.
func graveyardCardNoun(sel SelectionSyntax) (string, bool) {
	hasTypes := len(sel.RequiredTypesAny) > 0
	hasColors := len(sel.ColorsAny) > 0
	hasColorQualifier := sel.Colorless || sel.Multicolored
	hasSubtype := len(sel.SubtypesAny) > 0
	isPermanent := sel.Kind == SelectionPermanent

	categories := 0
	for _, present := range []bool{hasTypes, hasColors, hasColorQualifier, hasSubtype, isPermanent} {
		if present {
			categories++
		}
	}
	if categories > 1 {
		return "", false
	}
	switch {
	case hasTypes:
		return graveyardCardTypeNoun(sel)
	case isPermanent:
		return "permanent card", true
	case hasColors:
		if len(sel.ColorsAny) != 1 {
			return "", false
		}
		word, ok := colorWord(sel.ColorsAny[0])
		if !ok {
			return "", false
		}
		return word + " card", true
	case hasColorQualifier:
		if sel.Colorless && sel.Multicolored {
			return "", false
		}
		if sel.Colorless {
			return "colorless card", true
		}
		return "multicolored card", true
	case hasSubtype:
		if len(sel.SubtypesAny) != 1 {
			return "", false
		}
		return string(sel.SubtypesAny[0]) + " card", true
	default:
		if sel.Kind != SelectionCard {
			return "", false
		}
		return "card", true
	}
}

// graveyardCardTypeNoun reconstructs the card-type noun ("creature card",
// "creature or enchantment card"). A single type must be carried by the selection
// Kind so lowering's Kind-to-type mapping reproduces it (this excludes the
// instant and sorcery types, whose single-type form the compiler does not retain
// and which would otherwise lower to an unrestricted card). A union of two or
// more types is carried explicitly by the compiler, so each member is rendered
// from its card-type word and joined with " or ".
func graveyardCardTypeNoun(sel SelectionSyntax) (string, bool) {
	if len(sel.RequiredTypesAny) == 1 {
		noun, ok := permanentSelectionNoun(sel.Kind)
		if !ok {
			return "", false
		}
		word, ok := cardTypeWord(sel.RequiredTypesAny[0])
		if !ok || word != noun {
			return "", false
		}
		return noun + " card", true
	}
	words := make([]string, 0, len(sel.RequiredTypesAny))
	for _, cardType := range sel.RequiredTypesAny {
		word, ok := cardTypeWord(cardType)
		if !ok {
			return "", false
		}
		words = append(words, word)
	}
	return strings.Join(words, " or ") + " card", true
}

// graveyardManaValueClause renders the canonical " with mana value N or less"
// qualifier from a mana-value comparison. It accepts only the "or less" bound the
// printed Oracle wording uses, failing closed for any other comparison operator.
func graveyardManaValueClause(manaValue compare.Int) (string, bool) {
	if manaValue.Op != compare.LessOrEqual {
		return "", false
	}
	return " with mana value " + strconv.Itoa(manaValue.Value) + " or less", true
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
	words := normalizedWords(effect.Tokens)
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

// exactMultiBounceEffectSyntax recognizes the plural battlefield bounce
// "Return <N target permanents> to their owners' hands." (and the optional "up
// to N" form) that the executable backend lowers to one multi-target spec with
// one Bounce per slot. It accepts only the exact plural possessive destination
// for a multi-target permanent, failing closed for every other wording so the
// single-target "to its owner's hand" path is untouched.
func exactMultiBounceEffectSyntax(effect *EffectSyntax) bool {
	return len(effect.Targets) == 1 &&
		effect.Targets[0].Exact &&
		effect.Targets[0].Cardinality.Max >= 2 &&
		strings.EqualFold(exactEffectClauseText(effect), "Return "+effect.Targets[0].Text+" to their owners' hands.")
}

// exactSelfBounceEffectSyntax recognizes "Return this <object> to its owner's
// hand." where the subject is the source permanent itself (e.g. Etherium-Horn
// Sorcerer's "Return this creature to its owner's hand.").
func exactSelfBounceEffectSyntax(effect *EffectSyntax) bool {
	if len(effect.Targets) != 0 || len(effect.References) == 0 ||
		effect.References[0].Kind != ReferenceThisObject {
		return false
	}
	subject := joinedEffectText(effect.References[0].Tokens)
	return strings.EqualFold(exactEffectClauseText(effect), "Return "+subject+" to its owner's hand.")
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

// exactSelfSubjectReferenceText returns the rendered text of a single source
// self-reference, either "this <object>" (ReferenceThisObject) or the card's own
// name (ReferenceSelfName), used to recognize effects whose subject is the source
// permanent itself (e.g. "This creature gains flying until end of turn.").
func exactSelfSubjectReferenceText(references []Reference) (string, bool) {
	if len(references) != 1 {
		return "", false
	}
	switch references[0].Kind {
	case ReferenceThisObject, ReferenceSelfName:
		return joinedEffectText(references[0].Tokens), true
	}
	return "", false
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
				reference.Kind == ReferenceThisObject ||
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
	if effect.Divided {
		return exactDividedDamageText(effect, prefix, text)
	}
	if len(effect.Targets) == 0 {
		if !effect.Amount.Known {
			return false
		}
		if len(effect.DamageRecipientPair) == 2 {
			first, ok := exactGroupDamageRecipientText(effect.DamageRecipientPair[0])
			if !ok {
				return false
			}
			second, ok := exactGroupDamageRecipientText(effect.DamageRecipientPair[1])
			if !ok {
				return false
			}
			return text == fmt.Sprintf("%s %d damage to %s and %s.", prefix, effect.Amount.Value, first, second)
		}
		recipient, ok := exactGroupDamageRecipientText(effect.Selection)
		if !ok {
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
		return text == fmt.Sprintf("%s damage %s to %s.", prefix, effect.Amount.Text, target) ||
			text == fmt.Sprintf("%s damage to %s %s.", prefix, target, effect.Amount.Text)
	case EffectDynamicAmountFormForEach:
		return text == fmt.Sprintf("%s %d damage %s to %s.", prefix, effect.Amount.Multiplier, effect.Amount.Text, target)
	case EffectDynamicAmountFormWhereX:
		return text == fmt.Sprintf("%s X damage to %s, %s.", prefix, target, effect.Amount.Text)
	default:
		return false
	}
}

// dividedDamageEffect reports whether a damage effect uses the "divided as you
// choose among <targets>" wording, where a fixed total is split among the
// chosen targets (CR 601.2d).
func dividedDamageEffect(effect *EffectSyntax) bool {
	if effect.Kind != EffectDealDamage {
		return false
	}
	return effectContainsWords(normalizedWords(effect.Tokens), "divided", "as", "you", "choose", "among")
}

// exactDividedDamageText reconstructs the canonical "deals N damage divided as
// you choose among <cardinality> <noun>" clause and compares it byte-for-byte to
// the source. It supports only a fixed total and the cardinality and target
// nouns the executable backend can represent exactly, failing closed otherwise.
func exactDividedDamageText(effect *EffectSyntax, prefix, text string) bool {
	if !effect.Amount.Known || effect.Amount.Value < 1 ||
		effect.Amount.DynamicForm != EffectDynamicAmountFormNone ||
		effect.Amount.DynamicKind != EffectDynamicAmountNone ||
		effect.Negated ||
		len(effect.Targets) != 1 {
		return false
	}
	cardinality, ok := dividedCardinalityPhrase(effect.Targets[0].Cardinality)
	if !ok {
		return false
	}
	noun, ok := dividedTargetNoun(effect.Targets[0].Selection)
	if !ok {
		return false
	}
	expected := fmt.Sprintf("%s %d damage divided as you choose among %s %s.",
		prefix, effect.Amount.Value, cardinality, noun)
	return text == expected
}

// dividedCardinalityPhrase reconstructs the cardinal phrase that introduces the
// divided targets. It supports the enumerated "one or two" and "one, two, or
// three" ranges and the unbounded "any number of" form.
func dividedCardinalityPhrase(cardinality TargetCardinalitySyntax) (string, bool) {
	switch cardinality {
	case TargetCardinalitySyntax{Min: 1, Max: 2}:
		return "one or two", true
	case TargetCardinalitySyntax{Min: 1, Max: 3}:
		return "one, two, or three", true
	case TargetCardinalitySyntax{Min: 0, Max: 99}:
		return "any number of", true
	default:
		return "", false
	}
}

// dividedTargetNoun reconstructs the target noun phrase for divided damage. It
// supports "targets" (any target) and "target creatures" with no further
// qualifiers, failing closed for every other selector.
func dividedTargetNoun(selection SelectionSyntax) (string, bool) {
	switch selection.Kind {
	case SelectionAny:
		return "targets", true
	case SelectionCreature:
		if dividedPlainCreatureSelection(selection) {
			return "target creatures", true
		}
	default:
	}
	return "", false
}

// dividedPlainCreatureSelection reports that a creature selection carries no
// qualifier beyond its card type, so it reconstructs as a bare "target
// creatures" phrase.
func dividedPlainCreatureSelection(selection SelectionSyntax) bool {
	if selection.All || selection.Another || selection.Other ||
		selection.Attacking || selection.Blocking ||
		selection.Tapped || selection.Untapped ||
		selection.Colorless || selection.Multicolored ||
		selection.MatchManaValue || selection.MatchPower || selection.MatchToughness ||
		selection.Keyword != KeywordUnknown ||
		selection.Zone != zone.None ||
		selection.Controller != SelectionControllerAny ||
		len(selection.ExcludedTypes) != 0 ||
		len(selection.ExcludedColors) != 0 ||
		len(selection.ColorsAny) != 0 ||
		len(selection.SubtypesAny) != 0 ||
		len(selection.Supertypes) != 0 {
		return false
	}
	switch len(selection.RequiredTypesAny) {
	case 0:
		return true
	case 1:
		return selection.RequiredTypesAny[0] == CardTypeCreature
	default:
		return false
	}
}

// exactGroupDamageRecipientText reconstructs the canonical Oracle recipient
// phrase for a fixed-amount group damage spell ("each opponent", "each
// creature your opponents control", "each attacking creature", "each Goblin",
// "each nonartifact creature"). The caller compares the reconstruction against
// the literal source text, so any selector qualifier this renderer cannot
// represent makes the comparison fail and keeps the wording unsupported rather
// than silently dropping or approximating the filter.
func exactGroupDamageRecipientText(selection SelectionSyntax) (string, bool) {
	switch {
	case selection.Kind == SelectionOpponent && !selection.Other:
		return "each opponent", true
	case selection.Kind == SelectionPlayer && !selection.Other:
		return "each player", true
	}
	return exactGroupDamagePermanentRecipientText(selection)
}

// exactGroupDamagePermanentRecipientText reconstructs the recipient phrase for a
// group damage spell whose recipients form a single filtered permanent group. It
// renders only the controller, combat, tapped, single-color, single-subtype,
// single-excluded-type, keyword, and "other" qualifiers the executable backend
// can represent exactly, and fails closed for every other qualifier.
func exactGroupDamagePermanentRecipientText(selection SelectionSyntax) (string, bool) {
	if selection.All || selection.Another || selection.Zone != zone.None ||
		selection.Keyword != KeywordUnknown ||
		selection.MatchManaValue || selection.MatchPower || selection.MatchToughness ||
		selection.Colorless || selection.Multicolored ||
		len(selection.Supertypes) != 0 ||
		len(selection.ExcludedColors) != 0 ||
		len(selection.RequiredTypesAny) > 1 ||
		len(selection.ColorsAny) > 1 ||
		len(selection.SubtypesAny) > 1 ||
		len(selection.ExcludedTypes) > 1 {
		return "", false
	}
	if (selection.Attacking && selection.Blocking) ||
		(selection.Tapped && selection.Untapped) ||
		((selection.Tapped || selection.Untapped) && (selection.Attacking || selection.Blocking)) {
		return "", false
	}
	noun, hasNoun := permanentSelectionNoun(selection.Kind)
	if !hasNoun && selection.Kind != SelectionUnknown {
		return "", false
	}
	// The parser records a permanent noun both as the selection Kind and as a
	// redundant single-element RequiredTypesAny. Accept only that redundant form
	// (a union or a type inconsistent with the noun is not representable here).
	if len(selection.RequiredTypesAny) == 1 {
		requiredNoun, ok := permanentCardTypeNoun(selection.RequiredTypesAny[0])
		if !ok || !hasNoun || requiredNoun != noun {
			return "", false
		}
	}
	words := []string{"each"}
	if selection.Other {
		words = append(words, "other")
	}
	switch {
	case selection.Attacking:
		words = append(words, "attacking")
	case selection.Blocking:
		words = append(words, "blocking")
	case selection.Tapped:
		words = append(words, "tapped")
	case selection.Untapped:
		words = append(words, "untapped")
	default:
	}
	if len(selection.ColorsAny) == 1 {
		colorText, ok := colorWord(selection.ColorsAny[0])
		if !ok {
			return "", false
		}
		words = append(words, colorText)
	}
	if len(selection.SubtypesAny) == 1 {
		words = append(words, string(selection.SubtypesAny[0]))
	}
	if len(selection.ExcludedTypes) == 1 {
		if !hasNoun {
			return "", false
		}
		excludedNoun, ok := permanentCardTypeNoun(selection.ExcludedTypes[0])
		if !ok {
			return "", false
		}
		words = append(words, "non"+excludedNoun)
	}
	if hasNoun {
		words = append(words, noun)
	} else if len(selection.SubtypesAny) != 1 {
		return "", false
	}
	switch selection.Controller {
	case SelectionControllerAny:
	case SelectionControllerYou:
		words = append(words, "you", "control")
	case SelectionControllerOpponent:
		words = append(words, "your", "opponents", "control")
	case SelectionControllerNotYou:
		words = append(words, "you", "don't", "control")
	default:
		return "", false
	}
	return strings.Join(words, " "), true
}
