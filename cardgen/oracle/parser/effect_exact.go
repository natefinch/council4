package parser

import (
	"fmt"
	"slices"
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func exactEffectSyntax(effect *EffectSyntax) bool {
	switch effect.Kind {
	case EffectDealDamage:
		return exactDamageEffectSyntax(effect)
	case EffectCounter:
		return exactCounterEffectSyntax(effect)
	case EffectCreate:
		return exactCreateTokenEffectSyntax(effect) ||
			exactCreateNamedTokenEffectSyntax(effect) ||
			exactCreateCopyTokenEffectSyntax(effect)
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
			exactMultiBounceEffectSyntax(effect) ||
			exactMassBounceEffectSyntax(effect) ||
			exactSelfBounceEffectSyntax(effect) ||
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
	amount := effectAmountSourceText(effect)
	if effect.Context == EffectContextController {
		// Imperative controller form: "Sacrifice a creature." or the rarer
		// "You sacrifice a creature." Both compile to EffectContextController.
		return strings.EqualFold(text, fmt.Sprintf("Sacrifice %s %s.", amount, noun)) ||
			strings.EqualFold(text, fmt.Sprintf("You sacrifice %s %s.", amount, noun))
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
	prefix := fmt.Sprintf("%s sacrifices %s %s", subject, amount, noun)
	return strings.EqualFold(text, prefix+".") ||
		strings.EqualFold(text, prefix+" of their choice.")
}

func exactSearchEffectSyntax(effect *EffectSyntax) bool {
	return searchUnsupportedDetail(effect) == ""
}

// searchUnsupportedDetail reconstructs the canonical library-search clause from
// the parsed Selection and count and compares it byte-for-byte against the
// source. It recognizes the bounded shapes the runtime models: a singular or
// "up to N" search of your own library for a plain card-type, a basic land, or a
// union of basic land subtypes (optionally "basic"), moved to hand or the
// battlefield (optionally tapped, optionally revealed first), ending with "then
// shuffle". It returns "" when the clause is supported, or a diagnostic detail
// otherwise. Every richer rider (graveyard search, "with different names",
// mana-value/power filters, "for each player", X counts) fails closed.
func searchUnsupportedDetail(effect *EffectSyntax) string {
	const prefix = "Search your library for "
	const shuffleSuffix = ", then shuffle."
	text := effect.Text
	// A resolving optional search ("You may search your library for ...") carries
	// its "you may" choice as effect.Optional. Strip that prefix and restore the
	// sentence-initial capital so the remaining clause reconstructs against the
	// same canonical "Search your library for ..." shape as a mandatory tutor;
	// the optionality is preserved separately by effect.Optional. This mirrors
	// the optional-prefix handling in exactEffectClauseText.
	if effect.Optional {
		if rest, ok := strings.CutPrefix(text, "You may "); ok {
			text = titleFirstEffectText(rest)
		} else if rest, ok := strings.CutPrefix(text, "you may "); ok {
			text = titleFirstEffectText(rest)
		}
	}
	if !strings.HasPrefix(text, prefix) || !strings.HasSuffix(text, shuffleSuffix) {
		return `the executable source backend supports only searches of your library ending with "then shuffle"`
	}
	rest := strings.TrimPrefix(text, prefix)

	consumed, amount, plural := searchCountPrefix(rest)
	if consumed == "" || !effect.Amount.Known || effect.Amount.Value != amount {
		return "the executable source backend supports only exact singular-card search wording"
	}
	rest = rest[len(consumed):]

	filter, ok := canonicalSearchFilter(effect.Selection)
	if !ok {
		return unsupportedSearchFilterDetail(rest)
	}
	noun := "card"
	if filter != "" {
		noun = filter + " card"
	}
	if plural {
		noun += "s"
	}
	destination, ok := strings.CutPrefix(rest, noun+", ")
	if !ok {
		return unsupportedSearchFilterDetail(rest)
	}
	if !searchDestinationSupported(destination, plural) {
		return "the executable source backend supports only exact hand or battlefield search destinations"
	}
	return ""
}

// searchCountPrefix consumes the count phrase that follows "for ". It accepts the
// singular articles "a "/"an " (amount 1) and the bounded "up to <word> " form
// (amount 2..10, plural). It returns the consumed literal (empty when the phrase
// is unrecognized) so the caller can keep reconstructing the clause
// byte-for-byte.
func searchCountPrefix(rest string) (consumed string, amount int, plural bool) {
	switch {
	case strings.HasPrefix(rest, "a "):
		return "a ", 1, false
	case strings.HasPrefix(rest, "an "):
		return "an ", 1, false
	case strings.HasPrefix(rest, "up to "):
		after := rest[len("up to "):]
		for n := 2; n <= 10; n++ {
			word, found := cardinalWord(n)
			if found && strings.HasPrefix(after, word+" ") {
				return "up to " + word + " ", n, true
			}
		}
		return "", 0, false
	default:
		return "", 0, false
	}
}

// unsupportedSearchFilterDetail extracts the printed filter (the text before
// " card") for a fail-closed diagnostic when the filter is outside the modeled
// envelope.
func unsupportedSearchFilterDetail(rest string) string {
	filter, _, ok := strings.Cut(rest, " card")
	if !ok {
		return "the executable source backend supports only exact singular-card search wording"
	}
	return fmt.Sprintf("unsupported library-search filter %q", filter)
}

// canonicalSearchFilter renders the modeled portion of a search filter (the text
// between the article and " card") from the parsed Selection, returning ok=false
// for any attribute the runtime SearchSpec cannot express. Supported filters are
// a plain card, a single card type (land/creature/artifact/enchantment/
// planeswalker), optionally "basic", a subtype union with no separate type noun
// ("Forest or Island", "Sliver", "Aura or Equipment"), and a subtype paired with
// a card type ("Myr creature", "Dragon creature").
func canonicalSearchFilter(sel SelectionSyntax) (string, bool) {
	if sel.Controller != SelectionControllerAny ||
		sel.All || sel.Another || sel.Other || sel.Attacking || sel.Blocking ||
		sel.Tapped || sel.Untapped || sel.Colorless || sel.Multicolored ||
		sel.Keyword != KeywordUnknown || sel.Zone != zone.None ||
		sel.MatchManaValue || sel.MatchPower || sel.MatchToughness ||
		len(sel.ExcludedTypes) != 0 || len(sel.SourceTypes) != 0 ||
		len(sel.ColorsAny) != 0 || len(sel.ExcludedColors) != 0 {
		return "", false
	}
	basic := false
	switch len(sel.Supertypes) {
	case 0:
	case 1:
		if sel.Supertypes[0] != SupertypeBasic {
			return "", false
		}
		basic = true
	default:
		return "", false
	}
	prefix := ""
	if basic {
		prefix = "basic "
	}
	base, ok := searchFilterTypeNoun(sel.Kind)
	if !ok {
		return "", false
	}
	if len(sel.SubtypesAny) > 0 {
		words := make([]string, 0, len(sel.SubtypesAny))
		for _, sub := range sel.SubtypesAny {
			words = append(words, string(sub))
		}
		subtypes := joinOrList(words)
		switch sel.Kind {
		case SelectionCard:
			// A subtype union with no separate type noun ("Sliver", "Forest or
			// Island", "Aura or Equipment"): the subtype implies the type, so the
			// runtime matches by subtype alone. A card-kind selection must not
			// carry a required card type, because the compiler drops a single
			// required type and the resulting spec would silently lose it. The
			// "basic" supertype is meaningful only for the basic land subtypes.
			if len(sel.RequiredTypesAny) != 0 {
				return "", false
			}
			if basic && !allBasicLandSubtypes(sel.SubtypesAny) {
				return "", false
			}
			return prefix + subtypes, true
		case SelectionCreature, SelectionArtifact, SelectionEnchantment, SelectionLand:
			// A subtype paired with a card type ("Myr creature", "Dragon
			// creature"): the runtime matches by both card type and subtype.
			// "basic" pairs only with a bare land, never with a typed subtype.
			if basic {
				return "", false
			}
			return subtypes + " " + base, true
		default:
			return "", false
		}
	}
	if basic && base != "land" {
		// "basic" without a subtype is meaningful only for "basic land".
		return "", false
	}
	return prefix + base, true
}

// searchFilterTypeNoun maps a selection kind to the printed card-type noun a
// search filter uses, returning ok=false for kinds the runtime SearchSpec cannot
// express. A plain card kind has an empty noun. Instant and sorcery are absent
// because they reach the parser as a card kind carrying a required card type the
// compiler drops, which would lose the type from the lowered spec.
func searchFilterTypeNoun(kind SelectionKind) (string, bool) {
	switch kind {
	case SelectionCard:
		return "", true
	case SelectionLand:
		return "land", true
	case SelectionCreature:
		return "creature", true
	case SelectionArtifact:
		return "artifact", true
	case SelectionEnchantment:
		return "enchantment", true
	case SelectionPlaneswalker:
		return "planeswalker", true
	default:
		return "", false
	}
}

// allBasicLandSubtypes reports whether every subtype in subs is one of the five
// basic land subtypes, the only subtypes a "basic" search-filter union may carry.
func allBasicLandSubtypes(subs []types.Sub) bool {
	for _, sub := range subs {
		switch sub {
		case types.Plains, types.Island, types.Swamp, types.Mountain, types.Forest:
		default:
			return false
		}
	}
	return true
}

// joinOrList renders a noun list with Oracle "or" punctuation: "A", "A or B", or
// "A, B, or C".
func joinOrList(words []string) string {
	switch len(words) {
	case 1:
		return words[0]
	case 2:
		return words[0] + " or " + words[1]
	default:
		return strings.Join(words[:len(words)-1], ", ") + ", or " + words[len(words)-1]
	}
}

// searchDestinationSupported reports whether the clause tail (everything after
// the noun phrase, through "then shuffle.") is one of the exact hand- or
// battlefield-destination wordings the runtime models, in its singular ("it"/
// "that card") or plural ("them"/"those cards") form.
func searchDestinationSupported(destination string, plural bool) bool {
	singular := []string{
		"put it into your hand, then shuffle.",
		"put that card into your hand, then shuffle.",
		"reveal it, put it into your hand, then shuffle.",
		"reveal that card, put it into your hand, then shuffle.",
		"put it onto the battlefield, then shuffle.",
		"put that card onto the battlefield, then shuffle.",
		"put it onto the battlefield tapped, then shuffle.",
		"put that card onto the battlefield tapped, then shuffle.",
	}
	pluralForms := []string{
		"put them into your hand, then shuffle.",
		"put those cards into your hand, then shuffle.",
		"reveal them, put them into your hand, then shuffle.",
		"reveal those cards, put them into your hand, then shuffle.",
		"put them onto the battlefield, then shuffle.",
		"put those cards onto the battlefield, then shuffle.",
		"put them onto the battlefield tapped, then shuffle.",
		"put those cards onto the battlefield tapped, then shuffle.",
	}
	if plural {
		return slices.Contains(pluralForms, destination)
	}
	return slices.Contains(singular, destination)
}

func exactLifeEffectSyntax(effect *EffectSyntax, controllerVerb, subjectVerb string) bool {
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
		} else if effect.Context == EffectContextPriorSubject && len(effect.Targets) == 0 &&
			effect.Amount.DynamicForm == EffectDynamicAmountFormNone {
			// The subject is elided: it is inherited from the prior effect in a
			// compound sentence ("Target player draws two cards and loses 2
			// life"). The clause reconstructs from the bare third-person verb,
			// matching how exactDamageEffectSyntax handles a prior-subject
			// damage clause with no own subject tokens. Restricted to a
			// self-contained amount (a fixed value or the spell's cost X): a
			// trailing "where X is ..." amount form defines a single X shared by
			// every effect in the sentence, but the parser binds that clause to
			// only one effect, so reconstructing the elided-subject clause in
			// isolation would not faithfully model the shared amount.
			prefixes = []string{subjectVerb}
		}
	case EffectContextEventPlayer, EffectContextReferencedPlayer:
		prefixes = []string{"They " + controllerVerb, "That player " + subjectVerb}
	case EffectContextReferencedObjectController:
		if subject := referencedControllerSubjectText(effect); subject != "" {
			prefixes = []string{subject + " " + subjectVerb}
		}
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
	if effect.StaticSubject.Kind != EffectStaticSubjectNone {
		return exactGroupTemporaryKeywordEffectSyntax(effect, text)
	}
	if effect.Context == EffectContextPriorSubject {
		// A singular prior subject ("it") reads "gains <kw> …"; a plural group
		// prior subject ("creatures you control") reads "gain <kw> …".
		middle, ok := strings.CutPrefix(text, "gains ")
		if !ok {
			middle, ok = strings.CutPrefix(text, "gain ")
		}
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
	if effect.Context == EffectContextSource {
		subject, ok := exactSelfSubjectReferenceText(effect.SubjectReferences)
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

// exactGroupTemporaryKeywordEffectSyntax recognizes a resolving keyword grant to
// a never-resolving creature or permanent group until end of turn ("Creatures
// you control gain trample until end of turn."). The subject is reconstructed
// byte-exactly from the tokens covered by the static-subject span, mirroring
// exactGroupModifyPTEffectSyntax. text is the lowercased clause text.
func exactGroupTemporaryKeywordEffectSyntax(effect *EffectSyntax, text string) bool {
	var subject []shared.Token
	for i := range effect.Tokens {
		if spanCovers(effect.StaticSubject.Span, effect.Tokens[i].Span) {
			subject = append(subject, effect.Tokens[i])
		}
	}
	if len(subject) == 0 {
		return false
	}
	subjectText := strings.ToLower(joinedEffectText(subject))
	// A plural group reads "gain"; the singular "each <permanent>" form reads
	// "gains". Try both so the reconstruction stays byte-exact with the source.
	for _, verb := range []string{" gain ", " gains "} {
		middle, ok := strings.CutPrefix(text, subjectText+verb)
		if !ok {
			continue
		}
		middle, ok = strings.CutSuffix(middle, " until end of turn.")
		if ok && middle != "" && exactTemporaryKeywordList(middle) {
			return true
		}
	}
	return false
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

// exactCreateTokenEffectSyntax recognizes vanilla creature-token creation:
// "Create <count> [tapped] <P>/<T> [colorless | <colors>] <Subtypes> [artifact |
// enchantment] creature token[s] [with <keyword>]." with a fixed
// power/toughness, up to two colors (or colorless), one or two creature
// subtypes, an optional leading artifact/enchantment permanent type, an optional
// "tapped" entry modifier, and an optional single creature keyword. It fails
// closed for every richer shape (attacking entry, quoted abilities, multiple
// keywords, modifiers, ...).
func exactCreateTokenEffectSyntax(effect *EffectSyntax) bool {
	forEach := effect.Amount.DynamicForm == EffectDynamicAmountFormForEach
	if (effect.Context != EffectContextController &&
		effect.Context != EffectContextReferencedObjectController) ||
		!effect.TokenPTKnown ||
		effect.Negated ||
		len(effect.Targets) != 0 {
		return false
	}
	if forEach {
		if effect.Context != EffectContextController ||
			effect.Amount.DynamicKind == EffectDynamicAmountNone ||
			effect.Amount.Multiplier != 1 {
			return false
		}
	} else if !effect.Amount.Known || effect.Amount.Value < 1 {
		return false
	}
	sel := effect.Selection
	if len(sel.SubtypesAny) < 1 || len(sel.SubtypesAny) > 2 ||
		len(sel.ColorsAny) > 2 ||
		len(sel.ExcludedTypes) != 0 || len(sel.ExcludedColors) != 0 ||
		len(sel.Supertypes) != 0 ||
		sel.Multicolored ||
		sel.MatchPower || sel.MatchToughness || sel.MatchManaValue ||
		sel.Untapped || sel.Attacking || sel.Blocking ||
		sel.All || sel.Another || sel.Other {
		return false
	}
	typeWords, ok := tokenCreatureTypeWords(sel)
	if !ok {
		return false
	}
	tappedPart := ""
	if sel.Tapped {
		tappedPart = "tapped "
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
	if sel.Colorless {
		if len(sel.ColorsAny) != 0 {
			return false
		}
		colorPart = "colorless "
	} else if len(sel.ColorsAny) > 0 {
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
	if !forEach && effect.Amount.Value != 1 {
		countWord, noun = effectAmountSourceText(effect), "tokens"
	}
	specBody := fmt.Sprintf("%s %s%d/%d %s%s %s %s%s",
		countWord, tappedPart, effect.TokenPower, effect.TokenToughness, colorPart,
		strings.Join(subtypeWords, " "), typeWords, noun, keywordPart)
	if effect.Context == EffectContextReferencedObjectController {
		subject := referencedControllerSubjectText(effect)
		if subject == "" {
			return false
		}
		return strings.EqualFold(exactEffectClauseText(effect), subject+" creates "+specBody+".")
	}
	if forEach {
		return strings.EqualFold(fullEffectClauseText(effect), effect.Amount.Text+", create "+specBody+".")
	}
	return strings.EqualFold(exactEffectClauseText(effect), "Create "+specBody+".")
}

// namedArtifactTokenSubtype reports whether sub is a predefined artifact token
// whose fixed Oracle ability the runtime CreateToken/TokenDef model already
// represents (Treasure, Food, Clue, Blood, Gold, Lander, Mutagen). Every other
// named token (Powerstone and Map, whose mana-restriction and explore-on-target
// abilities are not yet modeled, plus Incubator, Junk, ...) fails closed pending
// follow-up work.
func namedArtifactTokenSubtype(sub types.Sub) bool {
	switch sub {
	case types.Treasure, types.Food, types.Clue, types.Blood,
		types.Gold, types.Lander, types.Mutagen:
		return true
	default:
		return false
	}
}

// exactCreateNamedTokenEffectSyntax recognizes "Create a [tapped] <Named> token."
// for a predefined artifact token that carries no printed power/toughness
// (Treasure, Food, Clue, Blood), including a fixed count ("Create two Treasure
// tokens."), an optional "tapped" entry modifier ("Create a tapped Treasure
// token."), and the referenced-controller form ("Its controller creates a
// Treasure token."). It fails closed for every richer shape (colored, keyworded,
// per-each, or any other named token).
func exactCreateNamedTokenEffectSyntax(effect *EffectSyntax) bool {
	if (effect.Context != EffectContextController &&
		effect.Context != EffectContextReferencedObjectController) ||
		effect.TokenPTKnown || effect.TokenCopyOfTarget ||
		effect.Negated ||
		len(effect.Targets) != 0 ||
		effect.Amount.DynamicForm == EffectDynamicAmountFormForEach ||
		!effect.Amount.Known || effect.Amount.Value < 1 {
		return false
	}
	sel := effect.Selection
	if sel.Kind != SelectionUnknown ||
		len(sel.SubtypesAny) != 1 ||
		!namedArtifactTokenSubtype(sel.SubtypesAny[0]) ||
		sel.Keyword != KeywordUnknown ||
		len(sel.ColorsAny) != 0 || len(sel.ExcludedColors) != 0 ||
		len(sel.RequiredTypesAny) != 0 || len(sel.ExcludedTypes) != 0 ||
		len(sel.SourceTypes) != 0 || len(sel.Supertypes) != 0 ||
		sel.MatchPower || sel.MatchToughness || sel.MatchManaValue ||
		sel.Untapped || sel.Attacking || sel.Blocking ||
		sel.All || sel.Another || sel.Other ||
		sel.Colorless || sel.Multicolored {
		return false
	}
	tappedPart := ""
	if sel.Tapped {
		tappedPart = "tapped "
	}
	countWord, noun := "a", "token"
	if effect.Amount.Value != 1 {
		countWord, noun = effectAmountSourceText(effect), "tokens"
	}
	specBody := fmt.Sprintf("%s %s%s %s", countWord, tappedPart, string(sel.SubtypesAny[0]), noun)
	if effect.Context == EffectContextReferencedObjectController {
		subject := referencedControllerSubjectText(effect)
		if subject == "" {
			return false
		}
		return strings.EqualFold(exactEffectClauseText(effect), subject+" creates "+specBody+".")
	}
	return strings.EqualFold(exactEffectClauseText(effect), "Create "+specBody+".")
}

// through the trailing period, unlike exactEffectClauseText, which drops any
// pre-verb iteration prefix at the last comma. The create-token recognizer uses
// it so a typed "for each <X>," iterator is validated against the source rather
// than silently ignored.
func fullEffectClauseText(effect *EffectSyntax) string {
	text := joinedEffectText(effect.Tokens)
	if len(effect.Tokens) > 0 && effect.Tokens[len(effect.Tokens)-1].Kind != shared.Period {
		text += "."
	}
	return text
}

// exactCreateCopyTokenEffectSyntax recognizes "Create a token that's a copy of
// <target>." where the token copies the effect's single exact target object
// (e.g. "Create a token that's a copy of target creature you control."). It
// fails closed for every richer copy shape (modified copies, multiple tokens,
// non-target copy sources).
func exactCreateCopyTokenEffectSyntax(effect *EffectSyntax) bool {
	if effect.Context != EffectContextController ||
		effect.TokenPTKnown ||
		effect.Negated ||
		!effect.Amount.Known || effect.Amount.Value != 1 ||
		len(effect.Targets) != 1 ||
		!effect.Targets[0].Exact {
		return false
	}
	want := "Create a token that's a copy of " + effect.Targets[0].Text + "."
	return strings.EqualFold(exactEffectClauseText(effect), want)
}

// effect's verb (e.g. "Its controller", "That creature's controller") for a
// referenced-object-controller effect. It returns "" when the verb is not found.
func referencedControllerSubjectText(effect *EffectSyntax) string {
	verb := slices.IndexFunc(effect.Tokens, func(token shared.Token) bool {
		return token.Span == effect.VerbSpan
	})
	if verb <= 0 {
		return ""
	}
	return joinedEffectText(effect.Tokens[:verb])
}

// tokenCreatureTypeWords returns the Oracle card-type phrase for a created
// creature token ("creature", "artifact creature", or "enchantment creature")
// from the selection's required card types. The token must be a creature; at
// most one additional permanent type (artifact or enchantment) is allowed. It
// fails closed for every other required-type set.
func tokenCreatureTypeWords(sel SelectionSyntax) (string, bool) {
	required := sel.RequiredTypesAny
	if len(required) == 0 {
		required = []CardType{CardTypeCreature}
	}
	hasCreature := false
	extra := CardTypeUnknown
	for _, cardType := range required {
		switch cardType {
		case CardTypeCreature:
			hasCreature = true
		case CardTypeArtifact, CardTypeEnchantment:
			if extra != CardTypeUnknown {
				return "", false
			}
			extra = cardType
		default:
			return "", false
		}
	}
	if !hasCreature {
		return "", false
	}
	if extra == CardTypeUnknown {
		return "creature", true
	}
	word, ok := permanentCardTypeNoun(extra)
	if !ok {
		return "", false
	}
	return word + " creature", true
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
	case EffectContextEachPlayer:
		prefixes = []string{"Each player " + subjectVerb}
	case EffectContextEachOpponent:
		prefixes = []string{"Each opponent " + subjectVerb}
	case EffectContextTarget:
		if len(effect.Targets) == 1 && effect.Targets[0].Exact &&
			exactCardCountTargetPlayer(effect.Targets[0].Selection) {
			prefixes = []string{titleFirstEffectText(effect.Targets[0].Text) + " " + subjectVerb}
		}
	case EffectContextPriorSubject:
		if len(effect.Targets) == 1 && effect.Targets[0].Exact &&
			exactCardCountTargetPlayer(effect.Targets[0].Selection) {
			prefixes = []string{titleFirstEffectText(effect.Targets[0].Text) + " " + subjectVerb}
		} else {
			prefixes = []string{controllerVerb, subjectVerb}
		}
	case EffectContextEventPlayer, EffectContextReferencedPlayer:
		prefixes = []string{"They " + strings.TrimSuffix(subjectVerb, "s"), "That player " + subjectVerb}
	case EffectContextReferencedObjectController:
		if subject := referencedControllerSubjectText(effect); subject != "" {
			prefixes = []string{subject + " " + subjectVerb}
		}
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

// exactCardCountTargetPlayer reports whether a single-target selection for a
// draw/discard/mill clause is an unqualified "target player" or "target
// opponent". These are the only player targets the executable backend's
// playerTargetSpec lowers, so any other selector kind keeps the clause
// unsupported rather than approximating the recipient.
func exactCardCountTargetPlayer(selection SelectionSyntax) bool {
	return selection.Kind == SelectionPlayer || selection.Kind == SelectionOpponent
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
	case EffectDurationWhileControlledCreatureEnchanted:
		return strings.EqualFold(text, prefix+" for as long as that creature is enchanted.")
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
		subject = "It"
	case EffectContextSource:
		s, ok := exactSelfSubjectReferenceText(effect.References)
		if !ok {
			return false
		}
		subject = s
	default:
		return false
	}
	power := signedEffectAmountText(effect.PowerDelta)
	toughness := signedEffectAmountText(effect.ToughnessDelta)
	text := exactEffectClauseText(effect)
	if effect.Amount.DynamicKind == EffectDynamicAmountNone {
		prefix := fmt.Sprintf("%s gets %s/%s", subject, power, toughness)
		if strings.EqualFold(text, prefix+" until end of turn.") ||
			strings.EqualFold(text, prefix+".") ||
			strings.HasPrefix(strings.ToLower(text), strings.ToLower(prefix+" and gains ")) &&
				strings.HasSuffix(strings.ToLower(text), " until end of turn.") {
			return true
		}
		// A plural or "up to N" target distributes the same pump onto each chosen
		// creature with the distributive "<subject> each get <p>/<t> until end of
		// turn." wording ("Two target creatures each get -1/-1 until end of
		// turn."). The plural verb "get" and the "each" distributive word replace
		// the singular "gets", so reconstruct that form only for multi-target
		// cardinalities. When the body also grants a keyword ("… each get +1/+1
		// and gain lifelink until end of turn."), the keyword grant is split into
		// a separate prior-subject effect and the modify clause reads
		// "<subject> each get <p>/<t>." with the until-end-of-turn duration spread
		// onto it, mirroring the singular "<subject> gets <p>/<t>." form accepted
		// above.
		if effect.Context == EffectContextTarget && effect.Targets[0].Cardinality.Max >= 2 {
			distributivePrefix := fmt.Sprintf("%s each get %s/%s", subject, power, toughness)
			return strings.EqualFold(text, distributivePrefix+" until end of turn.") ||
				strings.EqualFold(text, distributivePrefix+".")
		}
		return false
	}
	switch effect.Amount.DynamicForm {
	case EffectDynamicAmountFormForEach:
		return strings.EqualFold(text, fmt.Sprintf("%s gets %s/%s %s until end of turn.", subject, power, toughness, effect.Amount.Text)) ||
			strings.EqualFold(text, fmt.Sprintf("%s gets %s/%s until end of turn %s.", subject, power, toughness, effect.Amount.Text))
	case EffectDynamicAmountFormWhereX:
		powerX := signedPTSideText(effect.PowerDelta)
		toughnessX := signedPTSideText(effect.ToughnessDelta)
		return strings.EqualFold(text, fmt.Sprintf("%s gets %s/%s until end of turn, %s.", subject, powerX, toughnessX, effect.Amount.Text))
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
	prefix := fmt.Sprintf(
		"%s get %s/%s",
		joinedEffectText(subject),
		signedEffectAmountText(effect.PowerDelta),
		signedEffectAmountText(effect.ToughnessDelta),
	)
	text := exactEffectClauseText(effect)
	if strings.EqualFold(text, prefix+" until end of turn.") {
		return true
	}
	// "<subject> get +N/+N and gain <keyword> until end of turn." splits the
	// modify and keyword grant into separate effects; the modify clause then
	// reads "<subject> get +N/+N." with the until-end-of-turn duration spread
	// onto it. Accept that form only when the duration was recognized.
	return effect.Duration == EffectDurationUntilEndOfTurn &&
		strings.EqualFold(text, prefix+".")
}

func exactCounterPlacementEffectSyntax(effect *EffectSyntax) bool {
	if !effect.CounterKnown {
		return false
	}
	object := ""
	switch {
	case len(effect.Targets) == 1 && effect.Targets[0].Exact:
		object = effect.Targets[0].Text
		// "Put a +1/+1 counter on each of up to two target creatures." places one
		// counter on each of several targets, so the canonical object reads "each
		// of <target>" for any genuine multi-target cardinality (Max >= 2). The
		// singular and "up to one" forms keep the bare target text.
		if effect.Targets[0].Cardinality.Max >= 2 {
			object = "each of " + object
		}
	case len(effect.Targets) == 0:
		var ok bool
		object, ok = exactObjectReferenceText(effect.References)
		if !ok {
			object, ok = exactSelfSubjectReferenceText(effect.References)
		}
		if !ok && len(effect.References) == 0 {
			// "Put a +1/+1 counter on each creature you control." — a group of
			// permanents rather than a single object.
			object, ok = exactGroupDamagePermanentRecipientText(effect.Selection)
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
