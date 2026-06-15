package parser

import (
	"fmt"
	"slices"
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

func exactEffectSyntax(effect *EffectSyntax) bool {
	switch effect.Kind {
	case EffectDealDamage:
		return exactDamageEffectSyntax(effect)
	case EffectCounter:
		return exactCounterEffectSyntax(effect)
	case EffectCreate:
		return exactCreateTokenEffectSyntax(effect)
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
		prefixes = []string{"You " + controllerVerb, titleFirstEffectText(controllerVerb)}
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

// exactCreateTokenEffectSyntax recognizes the simplest vanilla creature-token
// creation: "Create a <P>/<T> [color] <Subtype> creature token." with a single
// token, fixed power/toughness, at most one color, exactly one creature subtype,
// and no other qualifier. It fails closed for every richer shape.
func exactCreateTokenEffectSyntax(effect *EffectSyntax) bool {
	if (effect.Context != EffectContextController &&
		effect.Context != EffectContextReferencedObjectController) ||
		!effect.TokenPTKnown ||
		!effect.Amount.Known || effect.Amount.Value < 1 ||
		effect.Negated || effect.Optional ||
		len(effect.Targets) != 0 {
		return false
	}
	sel := effect.Selection
	if sel.Kind != SelectionCreature ||
		len(sel.SubtypesAny) < 1 || len(sel.SubtypesAny) > 2 ||
		len(sel.ColorsAny) > 2 ||
		len(sel.ExcludedTypes) != 0 || len(sel.ExcludedColors) != 0 ||
		len(sel.Supertypes) != 0 ||
		sel.MatchPower || sel.MatchToughness || sel.MatchManaValue ||
		sel.Tapped || sel.Untapped || sel.Attacking || sel.Blocking ||
		sel.All || sel.Another || sel.Other {
		return false
	}
	keywordPart := ""
	if sel.Keyword != KeywordUnknown {
		if !tokenCreatureKeyword(sel.Keyword) {
			return false
		}
		word, ok := sel.Keyword.OracleWord()
		if !ok {
			return false
		}
		keywordPart = " with " + word
	}
	colorPart := ""
	if len(sel.ColorsAny) > 0 {
		words := make([]string, 0, len(sel.ColorsAny))
		for _, c := range sel.ColorsAny {
			word, ok := colorWord(c)
			if !ok {
				return false
			}
			words = append(words, word)
		}
		colorPart = strings.Join(words, " and ") + " "
	}
	subtypeWords := make([]string, 0, len(sel.SubtypesAny))
	for _, sub := range sel.SubtypesAny {
		subtypeWords = append(subtypeWords, string(sub))
	}
	countWord, noun := "a", "token"
	if effect.Amount.Value != 1 {
		countWord, noun = effectAmountSourceText(effect), "tokens"
	}
	specBody := fmt.Sprintf("%s %d/%d %s%s creature %s%s",
		countWord, effect.TokenPower, effect.TokenToughness, colorPart,
		strings.Join(subtypeWords, " "), noun, keywordPart)
	if effect.Context == EffectContextReferencedObjectController {
		subject := createTokenRecipientSubjectText(effect)
		if subject == "" {
			return false
		}
		return strings.EqualFold(exactEffectClauseText(effect), subject+" creates "+specBody+".")
	}
	return strings.EqualFold(exactEffectClauseText(effect), "Create "+specBody+".")
}

// createTokenRecipientSubjectText returns the rendered subject phrase before the
// "creates" verb (e.g. "Its controller", "That creature's controller") for a
// referenced-object-controller token creation. It returns "" when the verb is
// not found.
func createTokenRecipientSubjectText(effect *EffectSyntax) string {
	verb := slices.IndexFunc(effect.Tokens, func(token shared.Token) bool {
		return token.Span == effect.VerbSpan
	})
	if verb <= 0 {
		return ""
	}
	return joinedEffectText(effect.Tokens[:verb])
}

// tokenCreatureKeyword reports whether a keyword is a creature combat/evergreen
// keyword that is safe to grant a synthesized creature token through its typed
// static-ability body.
func tokenCreatureKeyword(k KeywordKind) bool {
	switch k {
	case KeywordFlying, KeywordFirstStrike, KeywordDoubleStrike, KeywordDeathtouch,
		KeywordHaste, KeywordHexproof, KeywordIndestructible, KeywordLifelink,
		KeywordMenace, KeywordReach, KeywordTrample, KeywordVigilance,
		KeywordDefender, KeywordShroud, KeywordWither, KeywordInfect, KeywordProwess:
		return true
	default:
		return false
	}
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
		prefixes = []string{controllerVerb, "You " + controllerVerb}
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
	case EffectContextSource:
		if effect.Amount.DynamicKind != EffectDynamicAmountNone {
			return false
		}
		if len(effect.References) != 1 || effect.References[0].Kind != ReferenceThisObject {
			return false
		}
		subject = joinedEffectText(effect.References[0].Tokens)
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
			object, ok = exactSourceObjectReferenceText(effect.References)
		}
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
