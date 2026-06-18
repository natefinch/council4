package parser

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/shared"
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
