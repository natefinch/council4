package parser

import (
	"slices"
	"strconv"
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func parseTargets(tokens []shared.Token, atoms Atoms) []TargetSyntax {
	var targets []TargetSyntax
	for i, token := range tokens {
		plural := equalWord(token, "targets")
		if !equalWord(token, "target") && !plural {
			continue
		}
		start := i
		cardinality := TargetCardinalitySyntax{Min: 1, Max: 1}
		if enumStart, enumCard, ok := enumeratedTargetCardinality(tokens, i); ok {
			start = enumStart
			cardinality = enumCard
		} else {
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
					// "any other target" carries both the "any" determiner and
					// the "other" distinctness qualifier; keep "any" in the noun
					// phrase so the selection reconstructs as an "any target".
					if equalWord(tokens[i-1], "other") && i >= 2 && equalWord(tokens[i-2], "any") {
						start = i - 2
					}
				}
			default:
			}
		}
		// A bare plural "targets" with no recognized preceding cardinality is not
		// a target production; only "<cardinality> targets" (e.g. "any number of
		// targets", "one or two targets") names targets directly.
		if plural && start == i {
			continue
		}
		if target, ok := parseOpponentControlledArtifactEnchantmentOrNonbasicLandTarget(tokens, i, cardinality); ok {
			targets = append(targets, target)
			continue
		}
		if target, ok := parseQualifiedDisjunctivePermanentTarget(tokens, atoms, start, i, cardinality); ok {
			targets = append(targets, target)
			continue
		}
		// "under target player's control" / "under target opponent's control" is a
		// control rider on a put destination (Yavimaya Dryad, Evil Presents): the
		// target player is the permanent's new controller, not a target the
		// sentence selects on its own. The search/put clause reconstructs and
		// consumes these tokens, and lowering synthesizes the controller target
		// spec, so producing a target here would leave a malformed, unconsumed
		// "target player's control" selection. Skip it.
		if isControlRiderTarget(tokens, i) {
			continue
		}
		end := targetSyntaxEnd(tokens, atoms, i+1)
		selectionTokens := append([]shared.Token(nil), tokens[start:i]...)
		selectionTokens = append(selectionTokens, tokens[i+1:end]...)
		nameUnique := false
		if head, ok := splitSelectionNameUniqueTail(selectionTokens); ok {
			selectionTokens = head
			nameUnique = true
		}
		var sameNameGroup *SameNameGroupSyntax
		if head, group, ok := splitSelectionSameNameGroupTail(selectionTokens); ok {
			selectionTokens = head
			sameNameGroup = group
		}
		dealtDamage := false
		if head, ok := splitSelectionDealtDamageThisTurnTail(selectionTokens); ok {
			selectionTokens = head
			dealtDamage = true
		}
		otherThanSource := false
		if head, ok := splitSelectionOtherThanSelfTail(selectionTokens, atoms); ok {
			selectionTokens = head
			otherThanSource = true
		}
		var spellTargetRestrictions []SpellTargetRestriction
		if head, restrictions, ok := splitSpellTargetRestrictionTail(selectionTokens); ok {
			selectionTokens = head
			spellTargetRestrictions = restrictions
		}
		selection := parseSelection(selectionTokens, atoms)
		if otherThanSource {
			selection.Another = true
			selection.OtherThanSource = true
		}
		qualifierTokens := selectionTokens
		if _, head, ok := splitSelectionNamedTail(selectionTokens); ok {
			// parseSelection has already captured the "named <Name>" tail as
			// RequiredName; scan only the head for unsupported qualifiers so a
			// supported noun ("creature named Fenric") survives instead of being
			// wiped by the unrecognized name tokens (The Curse of Fenric III).
			qualifierTokens = head
		}
		if targetSelectionHasUnsupportedQualifier(qualifierTokens, atoms) {
			selection = SelectionSyntax{Span: selection.Span, Text: selection.Text}
		}
		selection.NameUniqueAmongControlled = nameUnique
		selection.DealtDamageThisTurn = dealtDamage
		selection.SameNameGroup = sameNameGroup
		if len(spellTargetRestrictions) > 0 && selection.Kind == SelectionSpell {
			selection.SpellTargetRestrictions = spellTargetRestrictions
		}
		if plural {
			// "targets" with no following noun means "any target" — a permanent
			// or a player (CR 115.4).
			selection = SelectionSyntax{
				Span: shared.SpanOf(tokens[start:end]),
				Text: joinedEffectText(tokens[start:end]),
				Kind: SelectionAny,
			}
		}
		targetTokens := tokens[start:end]
		if selection.Kind == SelectionUnknown && selectionIsBareTokenTarget(selection) {
			// A bare "target token" names any token permanent. parseSelection
			// leaves Kind unset for the token noun (it carries no card type), so
			// promote it to a permanent selection here, in the target-only path,
			// without disturbing token creation which shares parseSelection.
			selection.Kind = SelectionPermanent
		}
		if conjunctiveTypeTarget(selection) {
			selection.ConjunctiveTypes = true
		}
		exact := exactRuntimeTargetSyntax(targetTokens, cardinality, selection)
		targets = append(targets, TargetSyntax{
			Span:        shared.SpanOf(tokens[start:end]),
			ChoiceSpan:  exactTargetChoiceSpan(tokens, start, targetTokens, cardinality, selection),
			Text:        joinedEffectText(tokens[start:end]),
			Cardinality: cardinality,
			Selection:   selection,
			Exact:       exact,
		})
	}
	return targets
}

// selectionIsBareTokenTarget reports whether a target selection names a plain
// "token" with no narrowing card type, subtype, color, supertype, or keyword
// (e.g. "target token you control"). Such a selection denotes any token
// permanent; controller, "another", and combat qualifiers stay compatible with
// the permanent target it promotes to. Any type-bearing wording ("token
// creature", "Treasure token") sets one of these fields and is excluded.
func selectionIsBareTokenTarget(selection SelectionSyntax) bool {
	return selection.TokenOnly &&
		!selection.NonToken &&
		!selection.Colorless &&
		!selection.Multicolored &&
		selection.Keyword == KeywordUnknown &&
		selection.ExcludedKeyword == KeywordUnknown &&
		len(selection.RequiredTypesAny) == 0 &&
		len(selection.ExcludedTypes) == 0 &&
		len(selection.Supertypes) == 0 &&
		len(selection.ExcludedSupertypes) == 0 &&
		len(selection.ColorsAny) == 0 &&
		len(selection.ExcludedColors) == 0 &&
		len(selection.SubtypesAny) == 0 &&
		len(selection.ExcludedSubtypes) == 0
}

// isControlRiderTarget reports whether the "target" token at index i opens an
// "under target player's control" / "under target opponent's control" control
// rider (Yavimaya Dryad, Evil Presents). The preceding word is "under" and the
// following two tokens are a player/opponent possessive and "control". The
// search/put clause owns and reconstructs these tokens, so parseTargets skips
// the production rather than emit a malformed "target player's control" target.
func isControlRiderTarget(tokens []shared.Token, i int) bool {
	if i < 1 || i+2 >= len(tokens) {
		return false
	}
	if !equalWord(tokens[i-1], "under") {
		return false
	}
	if !equalWord(tokens[i+1], "player's") && !equalWord(tokens[i+1], "opponent's") {
		return false
	}
	return equalWord(tokens[i+2], "control")
}

func parseOpponentControlledArtifactEnchantmentOrNonbasicLandTarget(
	tokens []shared.Token,
	targetIndex int,
	cardinality TargetCardinalitySyntax,
) (TargetSyntax, bool) {
	if cardinality != (TargetCardinalitySyntax{Min: 1, Max: 1}) ||
		targetIndex+11 > len(tokens) ||
		!effectWordsAt(tokens, targetIndex, "target", "artifact") ||
		tokens[targetIndex+2].Kind != shared.Comma ||
		!equalWord(tokens[targetIndex+3], "enchantment") ||
		tokens[targetIndex+4].Kind != shared.Comma ||
		!effectWordsAt(tokens, targetIndex+5, "or", "nonbasic", "land", "an", "opponent", "controls") {
		return TargetSyntax{}, false
	}
	end := targetIndex + 11
	if end < len(tokens) && tokens[end].Kind != shared.Period {
		return TargetSyntax{}, false
	}
	targetTokens := tokens[targetIndex:end]
	return TargetSyntax{
		Span:        shared.SpanOf(targetTokens),
		Text:        joinedEffectText(targetTokens),
		Cardinality: cardinality,
		Selection: SelectionSyntax{
			Span:       shared.SpanOf(tokens[targetIndex+1 : end]),
			Text:       joinedEffectText(tokens[targetIndex+1 : end]),
			Kind:       SelectionPermanent,
			Controller: SelectionControllerOpponent,
			Alternatives: []SelectionSyntax{
				{Kind: SelectionArtifact, RequiredTypesAny: []CardType{CardTypeArtifact}},
				{Kind: SelectionEnchantment, RequiredTypesAny: []CardType{CardTypeEnchantment}},
				{
					Kind:               SelectionLand,
					RequiredTypesAny:   []CardType{CardTypeLand},
					ExcludedSupertypes: []Supertype{SupertypeBasic},
				},
			},
		},
		Exact: true,
	}, true
}

// enumeratedTargetCardinality recognizes the small fixed enumerations used by
// divided-damage wordings — "one or two" and "one, two, or three" — that precede
// the target word at index i. It returns the start index of the phrase and the
// inclusive count range, or ok=false when no enumeration is present.
func enumeratedTargetCardinality(tokens []shared.Token, i int) (int, TargetCardinalitySyntax, bool) {
	if i >= 3 &&
		equalWord(tokens[i-3], "one") &&
		equalWord(tokens[i-2], "or") &&
		equalWord(tokens[i-1], "two") {
		return i - 3, TargetCardinalitySyntax{Min: 1, Max: 2}, true
	}
	if i >= 6 &&
		equalWord(tokens[i-6], "one") &&
		tokens[i-5].Kind == shared.Comma &&
		equalWord(tokens[i-4], "two") &&
		tokens[i-3].Kind == shared.Comma &&
		equalWord(tokens[i-2], "or") &&
		equalWord(tokens[i-1], "three") {
		return i - 6, TargetCardinalitySyntax{Min: 1, Max: 3}, true
	}
	return 0, TargetCardinalitySyntax{}, false
}

func exactRuntimeTargetSyntax(tokens []shared.Token, cardinality TargetCardinalitySyntax, selection SelectionSyntax) bool {
	if exactChosenCreatureCardsInYourGraveyardTargetSyntax(tokens, cardinality, selection) {
		return true
	}
	if cardinality != (TargetCardinalitySyntax{Min: 1, Max: 1}) {
		text := joinedEffectText(tokens)
		// "Up to one target <noun>" (Min 0, Max 1) is a single optional target
		// slot, so its <noun> phrase carries the same qualifiers (tapped state,
		// excluded card type, mana-value rider, ...) the mandatory single-target
		// form reconstructs. Reuse that reconstruction on the phrase following
		// the "up to one " count, falling back to the plural multi-target
		// reconstruction for the count forms (plural, "other", type unions) it
		// does not cover.
		if cardinality == (TargetCardinalitySyntax{Min: 0, Max: 1}) {
			const upToOnePrefix = "up to one "
			if len(text) > len(upToOnePrefix) &&
				strings.EqualFold(text[:len(upToOnePrefix)], upToOnePrefix) &&
				exactSinglePermanentTargetSyntax(text[len(upToOnePrefix):], selection) {
				return true
			}
		}
		return exactMultiPermanentTargetSyntax(text, cardinality, selection)
	}
	return exactSinglePermanentTargetSyntax(joinedEffectText(tokens), selection)
}

// exactSinglePermanentTargetSyntax reconstructs the canonical Oracle phrase for a
// single mandatory permanent (or spell/ability/player) target ("target tapped
// creature", "target nonland permanent", "target creature or planeswalker") and
// reports whether it round-trips byte-for-byte against text. It owns the full
// set of single-target qualifiers; exactRuntimeTargetSyntax also reuses it for
// the "up to one target <noun>" optional form after stripping the count words.
func exactSinglePermanentTargetSyntax(text string, selection SelectionSyntax) bool {
	if selection.RequiredName != "" {
		trimmed, had := strings.CutSuffix(text, " named "+selection.RequiredName)
		if !had {
			return false
		}
		text = trimmed
	}
	if selection.NameUniqueAmongControlled {
		trimmed, had := strings.CutSuffix(text, " "+nameUniqueAmongControlledClauseText)
		if !had {
			return false
		}
		text = trimmed
	}
	if selection.DealtDamageThisTurn {
		trimmed, had := strings.CutSuffix(text, " "+dealtDamageThisTurnClauseText)
		if !had {
			return false
		}
		text = trimmed
	}
	if selection.SameNameGroup != nil {
		trimmed, had := strings.CutSuffix(text, " "+selection.SameNameGroup.Text)
		if !had {
			return false
		}
		text = trimmed
	}
	if selection.OtherThanSource {
		index := strings.Index(strings.ToLower(text), " other than ")
		if index < 0 {
			return false
		}
		text = text[:index]
	}
	switch selection.Kind {
	case SelectionAny:
		if selection.Other {
			return text == "any other target"
		}
		return text == "any target"
	case SelectionPlayer:
		if selection.PlayerOrPlaneswalker {
			return strings.EqualFold(text, "target player or planeswalker")
		}
		return strings.EqualFold(text, "target player")
	case SelectionOpponent:
		if selection.PlayerOrPlaneswalker {
			return strings.EqualFold(text, "target opponent or planeswalker")
		}
		return strings.EqualFold(text, "target opponent")
	case SelectionActivatedAbility:
		return strings.EqualFold(text, "target activated ability") ||
			selectionHasCounterAbilityQualifier(selection)
	case SelectionTriggeredAbility:
		return strings.EqualFold(text, "target triggered ability") ||
			selectionHasCounterAbilityQualifier(selection)
	case SelectionActivatedOrTriggeredAbility:
		return strings.EqualFold(text, "target activated or triggered ability") ||
			selectionHasCounterAbilityQualifier(selection)
	case SelectionSpellActivatedOrTriggeredAbility:
		return strings.EqualFold(text, "target spell, activated ability, or triggered ability") ||
			strings.EqualFold(text, "target spell or ability") ||
			selectionHasCounterAbilityQualifier(selection)
	case SelectionTriggeredAbilityOrSpell:
		return selectionHasCounterAbilityQualifier(selection)
	case SelectionSpell:
		base := strings.ToLower(text)
		if len(selection.SpellTargetRestrictions) > 0 {
			clause, ok := spellTargetRestrictionsClause(selection.SpellTargetRestrictions)
			if !ok {
				return false
			}
			trimmed, had := strings.CutSuffix(base, clause)
			if !had {
				return false
			}
			base = trimmed
		}
		if selection.MatchManaValue {
			rider, ok := targetManaValueRider(selection.ManaValue)
			if !ok {
				return false
			}
			trimmed, had := strings.CutSuffix(base, rider)
			if !had {
				return false
			}
			base = trimmed
		}
		switch base {
		case "target spell", "target instant spell", "target sorcery spell", "target creature spell",
			"target artifact spell", "target noncreature spell":
			return true
		}
		if len(selection.RequiredTypesAny) >= 2 {
			return exactTypeUnionTargetSyntax(base, selection)
		}
		return exactSpellColorTargetSyntax(base, selection)
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

	if len(selection.RequiredTypesAny) >= 2 && !selection.ConjunctiveTypes {
		return exactTypeUnionTargetSyntax(text, selection)
	}
	if len(selection.SubtypesAny) >= 2 {
		return exactSubtypeUnionTargetSyntax(text, selection)
	}
	if len(selection.ExcludedTypes) > 0 && len(selection.ExcludedColors) > 0 {
		return exactExcludedTypeColorTargetSyntax(text, selection)
	}
	if len(selection.ExcludedTypes) > 0 {
		return exactExcludedTypeTargetSyntax(text, selection)
	}
	if len(selection.ExcludedColors) > 0 {
		return exactExcludedColorTargetSyntax(text, selection)
	}
	if len(selection.ExcludedSupertypes) > 0 {
		return exactExcludedSupertypeTargetSyntax(text, selection)
	}
	if len(selection.ExcludedSubtypes) > 0 {
		return exactExcludedSubtypeTargetSyntax(text, selection)
	}

	expected, ok := exactPermanentTargetText(selection)
	if !ok {
		return false
	}
	return strings.EqualFold(text, expected)
}

// targetManaValueRider reconstructs the " with mana value N", " with mana value
// N or less", or " with mana value N or greater" filter rider on a spell or
// permanent target from the parsed comparison. Only the exact, "or less", and
// "or greater" bounds the printed Oracle wording uses are modeled; every other
// comparison (less-than, greater-than, or an X-derived bound) fails closed.
func targetManaValueRider(mv compare.Int) (string, bool) {
	switch mv.Op {
	case compare.Equal:
		return " with mana value " + strconv.Itoa(mv.Value), true
	case compare.LessOrEqual:
		return " with mana value " + strconv.Itoa(mv.Value) + " or less", true
	case compare.GreaterOrEqual:
		return " with mana value " + strconv.Itoa(mv.Value) + " or greater", true
	default:
		return "", false
	}
}

func exactChosenCreatureCardsInYourGraveyardTargetSyntax(
	tokens []shared.Token,
	cardinality TargetCardinalitySyntax,
	selection SelectionSyntax,
) bool {
	if cardinality != (TargetCardinalitySyntax{Min: 2, Max: 2}) ||
		selection.Kind != SelectionCreature ||
		selection.Controller != SelectionControllerYou ||
		selection.Zone != zone.Graveyard ||
		chosenCreatureTargetHasScalarQualifiers(selection) ||
		chosenCreatureTargetHasListQualifiers(selection) {
		return false
	}
	return slices.Equal(selection.RequiredTypesAny, []CardType{CardTypeCreature}) &&
		strings.EqualFold(
			joinedEffectText(tokens),
			"two target creature cards in your graveyard",
		)
}

func chosenCreatureTargetHasScalarQualifiers(selection SelectionSyntax) bool {
	return selection.All ||
		selection.Another ||
		selection.Other ||
		selection.Attacking ||
		selection.Blocking ||
		selection.Tapped ||
		selection.Untapped ||
		selection.Colorless ||
		selection.Multicolored ||
		selection.MatchManaValue ||
		selection.MatchPower ||
		selection.MatchToughness ||
		selection.Keyword != KeywordUnknown ||
		selection.ExcludedKeyword != KeywordUnknown
}

func chosenCreatureTargetHasListQualifiers(selection SelectionSyntax) bool {
	return len(selection.ExcludedTypes) != 0 ||
		len(selection.SourceTypes) != 0 ||
		len(selection.Supertypes) != 0 ||
		len(selection.ExcludedSupertypes) != 0 ||
		len(selection.ColorsAny) != 0 ||
		len(selection.ExcludedColors) != 0 ||
		len(selection.SubtypesAny) != 0 ||
		len(selection.Alternatives) != 0
}

// exactChosenCardInYourGraveyardTargetSyntax reconstructs the canonical Oracle
// phrase for a single chosen card of one card type in the controller's
// graveyard ("target artifact card in your graveyard", Emry, Lurker of the
// Loch) and reports whether it round-trips byte-for-byte against the target
// tokens. It exists so the leading "Choose" verb of "Choose target <type> card
// in your graveyard." is consumed by exactTargetChoiceSpan; the target itself is
// already recognized through the normal graveyard-card target path.
//
// It accepts only the mandatory single-target cardinality, a controller-scoped
// graveyard zone, and exactly one required card type with no other qualifier, so
// any plural count, opponent or shared graveyard, additional type, color,
// supertype, subtype, name, mana-value, historic, or counter qualifier fails
// closed.
func exactChosenCardInYourGraveyardTargetSyntax(
	tokens []shared.Token,
	cardinality TargetCardinalitySyntax,
	selection SelectionSyntax,
) bool {
	if cardinality != (TargetCardinalitySyntax{Min: 1, Max: 1}) ||
		selection.Controller != SelectionControllerYou ||
		selection.Zone != zone.Graveyard ||
		selection.Historic ||
		selection.ConjunctiveTypes ||
		selection.NonToken ||
		selection.TokenOnly ||
		selection.BasicLandType ||
		selection.InclusiveOneOfEach ||
		selection.ManaValueX ||
		selection.RequiredName != "" ||
		selection.CounterRequired ||
		selection.CounterAny ||
		selection.CounterAbsent ||
		chosenCreatureTargetHasScalarQualifiers(selection) ||
		chosenCreatureTargetHasListQualifiers(selection) {
		return false
	}
	if len(selection.RequiredTypesAny) != 1 {
		return false
	}
	word, ok := cardTypeWord(selection.RequiredTypesAny[0])
	if !ok {
		return false
	}
	return strings.EqualFold(
		joinedEffectText(tokens),
		"target "+word+" card in your graveyard",
	)
}

func exactTargetChoiceSpan(
	tokens []shared.Token,
	start int,
	targetTokens []shared.Token,
	cardinality TargetCardinalitySyntax,
	selection SelectionSyntax,
) shared.Span {
	if start == 1 &&
		equalWord(tokens[0], "choose") &&
		(exactChosenCreatureCardsInYourGraveyardTargetSyntax(targetTokens, cardinality, selection) ||
			exactChosenCardInYourGraveyardTargetSyntax(targetTokens, cardinality, selection) ||
			exactRuntimeTargetSyntax(targetTokens, cardinality, selection)) {
		return tokens[0].Span
	}
	return shared.Span{}
}

// exactMultiPermanentTargetSyntax reconstructs the canonical Oracle phrase for a
// multi-target or optional permanent target the executable backend lowers to a
// single multi-target spec: "up to one target <noun>" (Min 0, Max 1), the fixed
// "<N> target <noun>s" (Min N, Max N), and the optional "up to <N> target
// <noun>s" (Min 0, Max N) for a small cardinal N, each with an optional plural
// "other" exclusion ("up to two other target creatures") and an optional single
// excluded card type ("up to two target nonland permanents"). It accepts only a
// plain permanent noun with those qualifiers and an optional controller clause,
// failing closed for every other qualifier so unsupported plural wordings keep
// failing the byte-exact round-trip.
func exactMultiPermanentTargetSyntax(text string, cardinality TargetCardinalitySyntax, selection SelectionSyntax) bool {
	prefix, plural, ok := multiTargetCardinalityPrefix(cardinality)
	if !ok {
		return false
	}
	if selection.All || selection.Another ||
		selection.Attacking || selection.Blocking || selection.Tapped || selection.Untapped ||
		selection.Keyword != KeywordUnknown || selection.Zone != zone.None ||
		selection.MatchManaValue || selection.MatchPower || selection.MatchToughness ||
		selection.PowerLessThanSource || selection.PowerGreaterThanSource ||
		selection.Colorless || selection.Multicolored ||
		len(selection.ColorsAny) != 0 || len(selection.ExcludedColors) != 0 ||
		len(selection.Supertypes) != 0 ||
		len(selection.SubtypesAny) != 0 {
		return false
	}
	// "any target" pluralizes to a bare "targets" head with no "target <noun>"
	// phrase ("two targets", "up to two targets"), unlike the permanent nouns
	// below. It accepts only the genuine plural cardinalities and no further
	// qualifier so a singular or qualified any-target wording fails closed.
	if selection.Kind == SelectionAny {
		if !plural || selection.Other || selection.PlayerOrPlaneswalker ||
			len(selection.RequiredTypesAny) != 0 || len(selection.ExcludedTypes) != 0 {
			return false
		}
		return strings.EqualFold(text, prefix+"targets")
	}
	// "target players" pluralizes the bare player head with no permanent noun
	// ("two target players", The Brothers' War chapter II). It accepts only the
	// genuine plural cardinalities and no further qualifier so a singular,
	// controller-scoped, "or planeswalker", or "other" wording fails closed.
	if selection.Kind == SelectionPlayer {
		if !plural || selection.Other || selection.PlayerOrPlaneswalker ||
			selection.Controller != SelectionControllerAny ||
			len(selection.RequiredTypesAny) != 0 || len(selection.ExcludedTypes) != 0 {
			return false
		}
		return strings.EqualFold(text, prefix+"target players")
	}
	// A card-type union ("artifact or enchantment") stands in for the permanent
	// noun and pluralizes every member ("two target artifacts or enchantments",
	// "up to one target creature or planeswalker"). The single-noun path below
	// rejects a multi-member RequiredTypesAny, so reconstruct the union here.
	if len(selection.RequiredTypesAny) >= 2 {
		return exactMultiPermanentUnionTargetSyntax(text, prefix, plural, selection)
	}
	noun, ok := permanentSelectionNoun(selection.Kind)
	if !ok || !selectionRedundantRequiredNoun(selection) {
		return false
	}
	// A single excluded card type renders as a "non<type>" prefix on the noun
	// ("nonland permanent"); pluralization still falls on the head noun so the
	// excluded prefix stays singular ("nonland permanents"). More than one
	// excluded type is an unrepresented shape and fails closed.
	excludedPrefix := ""
	switch len(selection.ExcludedTypes) {
	case 0:
	case 1:
		excludedNoun, ok := permanentCardTypeNoun(selection.ExcludedTypes[0])
		if !ok {
			return false
		}
		excludedPrefix = "non" + excludedNoun + " "
	default:
		return false
	}
	if plural {
		noun += "s"
	}
	// The plural "other" exclusion ("up to two other target creatures") reads
	// between the count words and "target"; "another" stays rejected above as a
	// singular shape the multi-target round-trip does not represent.
	otherWord := ""
	if selection.Other {
		otherWord = "other "
	}
	expected, ok := targetControllerSuffix(prefix+otherWord+"target "+excludedPrefix+noun, selection.Controller)
	if !ok {
		return false
	}
	return strings.EqualFold(text, expected)
}

// exactMultiPermanentUnionTargetSyntax reconstructs the canonical Oracle phrase
// for a multi-target or optional permanent target whose noun is a union of two or
// more permanent card types ("up to one target artifact or enchantment", "two
// target artifacts or enchantments", "up to two target creatures or
// planeswalkers"). Each union member pluralizes with the head when the
// cardinality is plural, joining as a bare "or" pair or an Oxford-comma list.
// It accepts an optional plural "other" exclusion and controller clause, failing
// closed for any subtype, excluded type, or other qualifier so unsupported union
// wordings keep failing the byte-exact round-trip. The lowering reuses the
// union-aware permanent target spec, which carries every member card type.
func exactMultiPermanentUnionTargetSyntax(text, prefix string, plural bool, selection SelectionSyntax) bool {
	if len(selection.ExcludedTypes) != 0 {
		return false
	}
	nouns := make([]string, 0, len(selection.RequiredTypesAny))
	for _, cardType := range selection.RequiredTypesAny {
		noun, ok := permanentCardTypeNoun(cardType)
		if !ok {
			return false
		}
		if plural {
			noun += "s"
		}
		nouns = append(nouns, noun)
	}
	otherWord := ""
	if selection.Other {
		otherWord = "other "
	}
	for _, conjunction := range []string{"or", "and/or"} {
		expected, ok := targetControllerSuffix(
			prefix+otherWord+"target "+joinUnionNounsSep(nouns, conjunction),
			selection.Controller,
		)
		if ok && strings.EqualFold(text, expected) {
			return true
		}
	}
	return false
}

// multiTargetCardinalityPrefix returns the canonical count words that precede
// "target" for a supported multi-target or optional cardinality, whether the
// target noun is plural, and whether the cardinality is one the round-trip
// represents. It reconstructs the unbounded "any number of" shape (Min 0,
// Max 99) as a plural count, reconstructs an adjacent "<lo> or <hi>" range (Min
// at least one, Max exactly one more, "one or two", "two or three") as a plural
// count, and fails closed for wider ranges and counts without a small-cardinal
// word.
func multiTargetCardinalityPrefix(c TargetCardinalitySyntax) (prefix string, plural, ok bool) {
	if c.Min == 0 && c.Max == 1 {
		return "up to one ", false, true
	}
	// The unbounded "any number of" shape (Min 0, Max 99) reconstructs its own
	// canonical count words and always pluralizes the target noun.
	if c.Min == 0 && c.Max == 99 {
		return "any number of ", true, true
	}
	if c.Max < 2 {
		return "", false, false
	}
	word, found := cardinalWord(c.Max)
	if !found {
		return "", false, false
	}
	if c.Min == 0 {
		return "up to " + word + " ", true, true
	}
	if c.Min == c.Max {
		return word + " ", true, true
	}
	// An adjacent count range ("one or two", "two or three") is the only bounded
	// minimum the round-trip represents: it reads as "<lo> or <hi> target". A
	// wider range ("one, two, or three") uses an Oxford-comma list this shape
	// does not model and fails closed.
	if c.Max == c.Min+1 {
		minWord, minFound := cardinalWord(c.Min)
		if !minFound {
			return "", false, false
		}
		return minWord + " or " + word + " ", true, true
	}
	return "", false, false
}

// cardinalWord renders a small cardinal count (1..10) as its Oracle number word,
// the inverse of CardinalWordValue. It fails closed for counts outside that
// range so unbounded or unusual cardinalities cannot reconstruct exact wording.
func cardinalWord(n int) (string, bool) {
	switch n {
	case 1:
		return "one", true
	case 2:
		return "two", true
	case 3:
		return "three", true
	case 4:
		return "four", true
	case 5:
		return "five", true
	case 6:
		return "six", true
	case 7:
		return "seven", true
	case 8:
		return "eight", true
	case 9:
		return "nine", true
	case 10:
		return "ten", true
	default:
		return "", false
	}
}

// exactSpellColorTargetSyntax reconstructs the canonical Oracle phrase for a
// color-qualified spell target the executable backend can represent: a single
// color ("target blue spell"), a single excluded color ("target nonblue spell"),
// "target colorless spell", or "target multicolored spell". It fails closed for
// any combination of color shapes, monocolored spells, type/subtype/supertype
// filters, or controller and zone qualifiers, keeping unsupported wordings out of
// the byte-exact round-trip.
func exactSpellColorTargetSyntax(text string, selection SelectionSyntax) bool {
	if selection.All || selection.Another || selection.Other ||
		selection.Attacking || selection.Blocking || selection.Tapped || selection.Untapped ||
		selection.Controller != SelectionControllerAny ||
		selection.Keyword != KeywordUnknown || selection.Zone != zone.None ||
		selection.MatchManaValue || selection.MatchPower || selection.MatchToughness ||
		selection.PowerLessThanSource || selection.PowerGreaterThanSource ||
		len(selection.RequiredTypesAny) != 0 || len(selection.ExcludedTypes) != 0 ||
		len(selection.Supertypes) != 0 || len(selection.SubtypesAny) != 0 {
		return false
	}
	colorShapes := len(selection.ColorsAny) + len(selection.ExcludedColors)
	if selection.Colorless {
		colorShapes++
	}
	if selection.Multicolored {
		colorShapes++
	}
	if colorShapes != 1 {
		return false
	}
	var qualifier string
	switch {
	case len(selection.ColorsAny) == 1:
		word, ok := colorWord(selection.ColorsAny[0])
		if !ok {
			return false
		}
		qualifier = word
	case len(selection.ExcludedColors) == 1:
		word, ok := colorWord(selection.ExcludedColors[0])
		if !ok {
			return false
		}
		qualifier = "non" + word
	case selection.Colorless:
		qualifier = "colorless"
	case selection.Multicolored:
		qualifier = "multicolored"
	default:
		return false
	}
	return strings.EqualFold(text, "target "+qualifier+" spell")
}

// exactPermanentTargetText reconstructs the canonical Oracle phrase for a single
// permanent target restricted only to qualifiers the executable backend can
// represent exactly: an "another"/"other" self-exclusion, a combat or tapped
// state, a single supertype, a single color, a single subtype that either
// qualifies an explicit type noun ("Beast creature") or stands in for it
// ("Soldier"), a "with power"/"with toughness" comparison, and a controller
// relation. It fails closed for every other qualifier so unsupported wordings
// keep failing the text-blind round-trip.
func exactPermanentTargetText(selection SelectionSyntax) (string, bool) {
	qualifier, ok := permanentSelectionQualifierWords(selection)
	if !ok {
		return "", false
	}
	var words []string
	switch {
	case selection.OtherThanSource:
		words = append(words, "target")
	case selection.Another:
		words = append(words, "another", "target")
	case selection.Other:
		words = append(words, "other", "target")
	default:
		words = append(words, "target")
	}
	words = append(words, qualifier...)
	return strings.Join(words, " "), true
}

// permanentSelectionQualifierWords reconstructs the canonical Oracle words that
// follow a single-permanent selection's leading determiner ("target", an
// article, or "another"): any combat/tapped state, a supertype, colors, a
// subtype, the permanent noun, the controller clause, and "with"/"without"
// qualifiers, in Oracle order. The determiner itself is supplied by the caller.
// It restricts to qualifiers the executable backend can represent exactly,
// failing closed for every other wording so unsupported selections keep failing
// the text-blind round-trip. See exactPermanentTargetText for the qualifier set.
func permanentSelectionQualifierWords(selection SelectionSyntax) ([]string, bool) {
	conjunctiveNoun, conjunctive := conjunctiveCreatureTargetNoun(selection)
	if selection.All || selection.Zone != zone.None ||
		selection.Colorless || selection.Multicolored ||
		len(selection.ExcludedColors) != 0 ||
		len(selection.ExcludedTypes) != 0 ||
		(len(selection.RequiredTypesAny) > 1 && !conjunctive) ||
		len(selection.SubtypesAny) > 1 ||
		len(selection.Supertypes) > 1 {
		return nil, false
	}
	combatWords, ok := selectionCombatStateWords(selection)
	if !ok {
		return nil, false
	}
	noun, hasNoun := permanentSelectionNoun(selection.Kind)
	if conjunctive {
		noun, hasNoun = conjunctiveNoun, true
	}
	if !hasNoun && selection.Kind != SelectionUnknown {
		return nil, false
	}
	// The parser records a permanent noun both as the selection Kind and as a
	// redundant single-element RequiredTypesAny. Accept only that redundant form
	// (a type inconsistent with the noun is not representable here).
	if len(selection.RequiredTypesAny) == 1 {
		requiredNoun, ok := permanentCardTypeNoun(selection.RequiredTypesAny[0])
		if !ok || !hasNoun || requiredNoun != noun {
			return nil, false
		}
	}
	words := append([]string(nil), combatWords...)
	if len(selection.Supertypes) == 1 {
		supertypeText, ok := supertypeWord(selection.Supertypes[0])
		if !ok {
			return nil, false
		}
		words = append(words, supertypeText)
	}
	if len(selection.ColorsAny) >= 1 {
		for i, colorValue := range selection.ColorsAny {
			colorText, ok := colorWord(colorValue)
			if !ok {
				return nil, false
			}
			if i > 0 {
				words = append(words, "or")
			}
			words = append(words, colorText)
		}
	}
	if len(selection.SubtypesAny) == 1 {
		words = append(words, string(selection.SubtypesAny[0]))
	}
	switch {
	case hasNoun:
		words = append(words, tokenQualifiedNoun(selection, noun)...)
	case len(selection.SubtypesAny) == 1:
	default:
		return nil, false
	}
	// The canonical Oracle ordering places the controller clause immediately
	// after the permanent noun and before any "with"/"without" qualifier, e.g.
	// "target creature you control without flying" and "target creature you
	// control with power 2". Reconstructing the controller clause here, rather
	// than as a trailing suffix, keeps those combined wordings byte-exact.
	controllerWords, ok := targetControllerWords(selection.Controller)
	if !ok {
		return nil, false
	}
	words = append(words, controllerWords...)
	keywordWords, ok := permanentKeywordQualifierWords(selection)
	if !ok {
		return nil, false
	}
	words = append(words, keywordWords...)
	numericWords, ok := permanentNumericQualifierWords(selection)
	if !ok {
		return nil, false
	}
	words = append(words, numericWords...)
	return words, true
}

// selectionCombatStateWords reconstructs the canonical Oracle combat/state
// qualifier words ("attacking", "blocking", "attacking or blocking", "tapped",
// "untapped") that precede a permanent noun. The mutually exclusive states are
// validated together: a permanent cannot be both tapped and untapped, nor carry
// a tapped/untapped state alongside a combat state. It is shared by the
// single-permanent reconstruction and the subtype-union reconstruction so both
// render the same prefix byte-exactly.
func selectionCombatStateWords(selection SelectionSyntax) ([]string, bool) {
	if (selection.Tapped && selection.Untapped) ||
		((selection.Tapped || selection.Untapped) && (selection.Attacking || selection.Blocking)) {
		return nil, false
	}
	switch {
	case selection.Attacking && selection.Blocking:
		return []string{"attacking", "or", "blocking"}, true
	case selection.Attacking:
		return []string{"attacking"}, true
	case selection.Blocking:
		return []string{"blocking"}, true
	case selection.Tapped:
		return []string{"tapped"}, true
	case selection.Untapped:
		return []string{"untapped"}, true
	default:
		return nil, true
	}
}

// tokenQualifiedNoun applies a selection's token adjective to its permanent
// noun, reconstructing the canonical Oracle wording. A bare "permanent" noun
// restricted to tokens prints as the single word "token" (Oracle never writes
// "token permanent"); a typed noun takes "token" as a trailing suffix ("artifact
// token", "creature token"), matching the printed Oracle order. A "nontoken"
// selector prefixes the noun directly ("nontoken creature", "nontoken
// permanent"). Selections without a token adjective return the noun unchanged.
func tokenQualifiedNoun(selection SelectionSyntax, noun string) []string {
	switch {
	case selection.TokenOnly && selection.Kind == SelectionPermanent:
		return []string{"token"}
	case selection.TokenOnly:
		return []string{noun, "token"}
	case selection.NonToken:
		return []string{"nontoken", noun}
	default:
		return []string{noun}
	}
}

// exactControlledBounceSelectionText reconstructs the canonical Oracle phrase for
// the permanent that a controlled-choice bounce returns: "a"/"an"/"another"
// followed by the same qualifier words an equivalent target would carry ("a red
// or green creature you control", "another permanent you control"). Only the
// "you control" relation is representable, because the chooser is the resolving
// controller picking from their own permanents; every other controller relation,
// and the "other" (mass-exclusion) determiner, fails closed.
func exactControlledBounceSelectionText(selection SelectionSyntax) (string, bool) {
	if selection.Controller != SelectionControllerYou || selection.Other {
		return "", false
	}
	qualifier, ok := permanentSelectionQualifierWords(selection)
	if !ok || len(qualifier) == 0 {
		return "", false
	}
	determiner := indefiniteArticle(qualifier[0])
	if selection.Another {
		determiner = "another"
	}
	return strings.Join(append([]string{determiner}, qualifier...), " "), true
}

// indefiniteArticle returns the English indefinite article ("a"/"an") for word.
// It uses the leading letter, which is exact for the permanent qualifiers the
// controlled-choice bounce reconstructs ("an artifact", "a creature"); a mismatch
// simply fails the byte-exact round-trip rather than mis-supporting a card.
func indefiniteArticle(word string) string {
	if word == "" {
		return "a"
	}
	switch word[0] {
	case 'a', 'e', 'i', 'o', 'u', 'A', 'E', 'I', 'O', 'U':
		return "an"
	}
	return "a"
}

// permanentKeywordQualifierWords reconstructs the "with <keyword>" clause of a
// permanent target whose selection carries one recognized keyword (e.g. "target
// creature with flying"). It returns no words when the selection has no keyword,
// and fails closed when a keyword coexists with a numeric "with ..." qualifier
// whose combined ordering the canonical phrasing cannot reproduce, keeping the
// text-blind round-trip honest.
func permanentKeywordQualifierWords(selection SelectionSyntax) ([]string, bool) {
	if selection.Keyword == KeywordUnknown && selection.ExcludedKeyword == KeywordUnknown {
		return nil, true
	}
	if selection.Keyword != KeywordUnknown && selection.ExcludedKeyword != KeywordUnknown {
		return nil, false
	}
	if selection.MatchManaValue || selection.MatchPower || selection.MatchToughness ||
		selection.PowerLessThanSource || selection.PowerGreaterThanSource {
		return nil, false
	}
	if selection.ExcludedKeyword != KeywordUnknown {
		word, ok := selection.ExcludedKeyword.OracleWord()
		if !ok {
			return nil, false
		}
		return []string{"without", word}, true
	}
	word, ok := selection.Keyword.OracleWord()
	if !ok {
		return nil, false
	}
	return []string{"with", word}, true
}

// permanentNumericQualifierWords reconstructs the "with mana value"/"with
// power"/"with toughness" clause of a permanent target. It returns no words when
// the selection carries no mana value, power, or toughness comparison, and fails
// closed for any comparison shape the canonical phrasing cannot reproduce,
// keeping the text-blind round-trip honest.
func permanentNumericQualifierWords(selection SelectionSyntax) ([]string, bool) {
	var clauses [][]string
	if selection.MatchManaValue {
		clause, ok := comparisonClauseWords("mana value", selection.ManaValue)
		if !ok {
			return nil, false
		}
		clauses = append(clauses, clause)
	}
	if selection.MatchPower {
		clause, ok := comparisonClauseWords("power", selection.Power)
		if !ok {
			return nil, false
		}
		clauses = append(clauses, clause)
	}
	if selection.MatchToughness {
		clause, ok := comparisonClauseWords("toughness", selection.Toughness)
		if !ok {
			return nil, false
		}
		clauses = append(clauses, clause)
	}
	if selection.PowerLessThanSource || selection.PowerGreaterThanSource {
		// "with lesser power" / "with greater power" compares the match to the
		// source permanent's power (Mentor). The adjective precedes the noun, so
		// it cannot share the "with <noun> N" ordering of the fixed comparisons;
		// fail closed when a fixed numeric clause is also present.
		if len(clauses) != 0 {
			return nil, false
		}
		adjective := "lesser"
		if selection.PowerGreaterThanSource {
			adjective = "greater"
		}
		return []string{"with", adjective, "power"}, true
	}
	if len(clauses) == 0 {
		return nil, true
	}
	words := []string{"with"}
	for i, clause := range clauses {
		if i > 0 {
			words = append(words, "and")
		}
		words = append(words, clause...)
	}
	return words, true
}

// comparisonClauseWords renders a single "<qualifier> N", "<qualifier> N or less",
// or "<qualifier> N or greater" clause. It fails closed for comparison operators
// without a canonical Oracle phrasing the round-trip can reproduce.
func comparisonClauseWords(qualifier string, comparison compare.Int) ([]string, bool) {
	value := strconv.Itoa(comparison.Value)
	switch comparison.Op {
	case compare.Equal:
		return []string{qualifier, value}, true
	case compare.LessOrEqual:
		return []string{qualifier, value, "or", "less"}, true
	case compare.GreaterOrEqual:
		return []string{qualifier, value, "or", "greater"}, true
	default:
		return nil, false
	}
}

// exactTypeUnionTargetSyntax recognizes a permanent target whose only restriction
// is a union of card types, e.g. "target creature or planeswalker" or "target
// artifact or enchantment you control". A single excluded card type is also
// supported, rendering as a "non<type>" qualifier on the union ("target
// noncreature artifact or noncreature enchantment"). It fails closed when any
// other qualifier (color, supertype, subtype, power, toughness, keyword, zone,
// combat or tapped state, "another"/"other") is present, or when any member is
// not a permanent card type.
func exactTypeUnionTargetSyntax(text string, selection SelectionSyntax) bool {
	if selection.All || selection.Another || selection.Other ||
		selection.Attacking || selection.Blocking || selection.Tapped || selection.Untapped ||
		selection.Keyword != KeywordUnknown || selection.ExcludedKeyword != KeywordUnknown ||
		selection.Zone != zone.None || selection.Colorless || selection.Multicolored ||
		selection.MatchPower || selection.MatchToughness ||
		len(selection.Supertypes) != 0 ||
		len(selection.ColorsAny) != 0 || len(selection.ExcludedColors) != 0 ||
		len(selection.SubtypesAny) != 0 || len(selection.SourceTypes) != 0 {
		return false
	}
	spellUnion := selection.Kind == SelectionSpell
	cardTypeNoun := permanentCardTypeNoun
	if spellUnion {
		if selection.MatchManaValue || len(selection.ExcludedTypes) != 0 {
			return false
		}
		cardTypeNoun = cardTypeWord
	} else if _, ok := permanentSelectionNoun(selection.Kind); !ok {
		return false
	}
	// A single excluded card type renders as a "non<type>" qualifier on the type
	// union. Oracle prints it either once before the whole union ("noncreature
	// artifact or enchantment") or repeated on every member ("noncreature
	// artifact or noncreature enchantment"); both describe the same selection, so
	// the round-trip reconstructs and accepts either rendering below.
	excludedPrefix := ""
	if len(selection.ExcludedTypes) != 0 {
		if len(selection.ExcludedTypes) != 1 {
			return false
		}
		excludedNoun, ok := permanentCardTypeNoun(selection.ExcludedTypes[0])
		if !ok {
			return false
		}
		excludedPrefix = "non" + excludedNoun + " "
	}
	nouns := make([]string, 0, len(selection.RequiredTypesAny))
	seen := make(map[CardType]bool, len(selection.RequiredTypesAny))
	for _, cardType := range selection.RequiredTypesAny {
		if seen[cardType] {
			return false
		}
		seen[cardType] = true
		noun, ok := cardTypeNoun(cardType)
		if !ok {
			return false
		}
		nouns = append(nouns, noun)
	}
	unions := []string{joinUnionNouns(nouns)}
	if excludedPrefix != "" {
		prefixed := make([]string, len(nouns))
		for i := range nouns {
			prefixed[i] = excludedPrefix + nouns[i]
		}
		unions = []string{excludedPrefix + joinUnionNouns(nouns), joinUnionNouns(prefixed)}
	}
	for _, union := range unions {
		expected, ok := typeUnionTargetExpected(union, spellUnion, selection)
		if ok && strings.EqualFold(text, expected) {
			return true
		}
	}
	return false
}

// typeUnionTargetExpected appends the spell, controller, and mana-value suffixes
// shared by every type-union target rendering to a reconstructed union noun
// phrase. A trailing "with mana value N or less/greater" qualifies the whole
// union ("target creature or planeswalker with mana value 3 or less"); every
// permanent has a mana value, so the qualifier applies uniformly to each member.
// Power and toughness are rejected by the caller because they exist only on
// creatures and would silently drop the non-creature members. Only the
// controller-free wording carries a mana-value qualifier, so a union mixing one
// with a controller clause fails the round-trip closed.
func typeUnionTargetExpected(union string, spellUnion bool, selection SelectionSyntax) (string, bool) {
	expected := "target " + union
	if spellUnion {
		expected += " spell"
	}
	switch selection.Controller {
	case SelectionControllerAny:
	case SelectionControllerYou:
		expected += " you control"
	case SelectionControllerOpponent:
		expected += " an opponent controls"
	case SelectionControllerNotYou:
		expected += " you don't control"
	default:
		return "", false
	}
	if selection.MatchManaValue {
		if selection.Controller != SelectionControllerAny {
			return "", false
		}
		clause, ok := comparisonClauseWords("mana value", selection.ManaValue)
		if !ok {
			return "", false
		}
		expected += " with " + strings.Join(clause, " ")
	}
	return expected, true
}

// joinUnionNouns renders a card-type union the way Oracle text does: a two-member
// union joins with a bare "or" ("artifact or enchantment"), while a union of
// three or more members uses an Oxford-comma list ("artifact, creature, or
// enchantment"). A single noun renders unchanged.
func joinUnionNouns(nouns []string) string {
	return joinUnionNounsSep(nouns, "or")
}

// joinUnionNounsSep renders a card-type union joining its final member with the
// given conjunction ("or" for the bare list, "and/or" for the inclusive-list
// wording some cards print — "artifacts, creatures, and/or lands"). The two
// conjunctions describe the same union, so the byte-exact round-trip tries both.
func joinUnionNounsSep(nouns []string, conjunction string) string {
	switch len(nouns) {
	case 0:
		return ""
	case 1:
		return nouns[0]
	case 2:
		return nouns[0] + " " + conjunction + " " + nouns[1]
	default:
		return strings.Join(nouns[:len(nouns)-1], ", ") + ", " + conjunction + " " + nouns[len(nouns)-1]
	}
}

// exactSubtypeUnionTargetSyntax recognizes a permanent target whose only
// restriction is a union of subtypes that stands in for the permanent noun, e.g.
// "target Skeleton, Vampire, or Zombie". It fails closed when any other
// qualifier (card type, color, supertype, power, toughness, keyword, zone,
// combat or tapped state, "another"/"other", or excluded types/colors) is
// present, so only the bare subtype union with an optional controller clause
// reconstructs byte-exact.
func exactSubtypeUnionTargetSyntax(text string, selection SelectionSyntax) bool {
	if selection.Kind != SelectionUnknown ||
		selection.All || selection.Another || selection.Other ||
		selection.Keyword != KeywordUnknown || selection.ExcludedKeyword != KeywordUnknown ||
		selection.Zone != zone.None || selection.Colorless || selection.Multicolored ||
		selection.MatchManaValue || selection.MatchPower || selection.MatchToughness ||
		selection.PowerLessThanSource || selection.PowerGreaterThanSource ||
		len(selection.RequiredTypesAny) != 0 || len(selection.ExcludedTypes) != 0 ||
		len(selection.Supertypes) != 0 ||
		len(selection.ColorsAny) != 0 || len(selection.ExcludedColors) != 0 {
		return false
	}
	combatWords, ok := selectionCombatStateWords(selection)
	if !ok {
		return false
	}
	nouns := make([]string, 0, len(selection.SubtypesAny))
	for _, subtype := range selection.SubtypesAny {
		nouns = append(nouns, string(subtype))
	}
	words := append([]string{"target"}, combatWords...)
	expected := strings.Join(append(words, joinUnionNouns(nouns)), " ")
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

// permanentCardTypeNoun returns the lowercase Oracle noun for a permanent card
// type. It fails closed for the non-permanent spell types (instant, sorcery).
func permanentCardTypeNoun(cardType CardType) (string, bool) {
	switch cardType {
	case CardTypeArtifact:
		return "artifact", true
	case CardTypeBattle:
		return "battle", true
	case CardTypeCreature:
		return "creature", true
	case CardTypeEnchantment:
		return "enchantment", true
	case CardTypeLand:
		return "land", true
	case CardTypePlaneswalker:
		return "planeswalker", true
	default:
		return "", false
	}
}

// conjunctiveCreatureTargetNoun returns the Oracle noun for a single permanent
// target whose required card types name a creature conjoined with one other
// permanent type ("artifact creature", "enchantment creature"). It is meaningful
// only when ConjunctiveTypes marks the type set as all-of rather than any-of, and
// fails closed otherwise.
func conjunctiveCreatureTargetNoun(selection SelectionSyntax) (string, bool) {
	if !selection.ConjunctiveTypes || len(selection.RequiredTypesAny) != 2 {
		return "", false
	}
	noun, ok := tokenCreatureTypeWords(selection)
	if !ok || noun == "creature" {
		return "", false
	}
	return noun, true
}

// conjunctiveTypeTarget reports whether a parsed target names two card types a
// permanent must carry at once ("artifact creature") rather than any one of them
// ("artifact or creature"). The two forms record the same RequiredTypesAny, so
// the conjunctive sense is recognized from the adjacent Oracle noun ("artifact
// creature") that the disjunctive "X or Y" wording never spells.
func conjunctiveTypeTarget(selection SelectionSyntax) bool {
	if len(selection.RequiredTypesAny) != 2 {
		return false
	}
	noun, ok := tokenCreatureTypeWords(selection)
	if !ok || noun == "creature" {
		return false
	}
	return strings.Contains(strings.ToLower(selection.Text), noun)
}

// selectionJoinsCardNounsWithAndOr reports whether a selection clause joins two
// or more singular "card" nouns with the "and/or" conjunction ("a Saga card
// and/or a land card"). The "and/or" token sequence marks the inclusive
// one-of-each wording, and requiring the singular "card" noun at least twice
// excludes a plural single-match union ("artifacts, creatures, and/or lands").
func selectionJoinsCardNounsWithAndOr(tokens []shared.Token) bool {
	andOr := false
	cardNouns := 0
	for i := range tokens {
		if i+2 < len(tokens) &&
			equalWord(tokens[i], "and") &&
			tokens[i+1].Kind == shared.Slash &&
			equalWord(tokens[i+2], "or") {
			andOr = true
		}
		if equalWord(tokens[i], "card") {
			cardNouns++
		}
	}
	return andOr && cardNouns >= 2
}

// permanentSelectionNoun returns the lowercase Oracle noun for a permanent
// selection kind. It fails closed for non-permanent selection kinds.
func permanentSelectionNoun(kind SelectionKind) (string, bool) {
	switch kind {
	case SelectionArtifact:
		return "artifact", true
	case SelectionBattle:
		return "battle", true
	case SelectionCreature:
		return "creature", true
	case SelectionEnchantment:
		return "enchantment", true
	case SelectionLand:
		return "land", true
	case SelectionPermanent:
		return "permanent", true
	case SelectionPlaneswalker:
		return "planeswalker", true
	default:
		return "", false
	}
}

// targetControllerSuffix appends the canonical controller clause for a target's
// controller relation, returning false for an unrecognized relation.
func targetControllerSuffix(expected string, controller SelectionController) (string, bool) {
	switch controller {
	case SelectionControllerAny:
		return expected, true
	case SelectionControllerYou:
		return expected + " you control", true
	case SelectionControllerOpponent:
		return expected + " an opponent controls", true
	case SelectionControllerNotYou:
		return expected + " you don't control", true
	default:
		return "", false
	}
}

// targetControllerWords returns the canonical controller clause for a target as a
// word slice, so callers can place it before trailing "with"/"without"
// qualifiers ("target creature you control without flying") rather than only at
// the end of the phrase. It fails closed for any unrecognized controller.
func targetControllerWords(controller SelectionController) ([]string, bool) {
	switch controller {
	case SelectionControllerAny:
		return nil, true
	case SelectionControllerYou:
		return []string{"you", "control"}, true
	case SelectionControllerOpponent:
		return []string{"an", "opponent", "controls"}, true
	case SelectionControllerNotYou:
		return []string{"you", "don't", "control"}, true
	default:
		return nil, false
	}
}

// exactExcludedTypeTargetSyntax recognizes a permanent target whose only
// restriction is a single excluded card type ("target nonland permanent",
// "target noncreature artifact"). It fails closed when any other qualifier is
// present or when more than one type is excluded.
// selectionRedundantRequiredNoun reports whether selection's RequiredTypesAny is
// either empty or the single redundant card-type the parser records alongside a
// permanent noun Kind (e.g. "creature" recorded both as Kind and RequiredTypesAny).
// Excluded-color/type target reconstruction renders from Kind, so it accepts only
// that redundant form.
func selectionRedundantRequiredNoun(selection SelectionSyntax) bool {
	if len(selection.RequiredTypesAny) == 0 {
		return true
	}
	if len(selection.RequiredTypesAny) != 1 {
		return false
	}
	noun, hasNoun := permanentSelectionNoun(selection.Kind)
	if !hasNoun {
		return false
	}
	requiredNoun, ok := permanentCardTypeNoun(selection.RequiredTypesAny[0])
	return ok && requiredNoun == noun
}

func exactExcludedColorTargetSyntax(text string, selection SelectionSyntax) bool {
	if selection.All || selection.Another || selection.Other ||
		selection.Attacking || selection.Blocking || selection.Tapped || selection.Untapped ||
		selection.Keyword != KeywordUnknown || selection.Zone != zone.None ||
		selection.MatchManaValue || selection.MatchPower || selection.MatchToughness ||
		selection.PowerLessThanSource || selection.PowerGreaterThanSource ||
		selection.Colorless || selection.Multicolored ||
		len(selection.Supertypes) != 0 ||
		len(selection.ColorsAny) != 0 || len(selection.ExcludedTypes) != 0 ||
		len(selection.SubtypesAny) != 0 {
		return false
	}
	if !selectionRedundantRequiredNoun(selection) {
		return false
	}
	if len(selection.ExcludedColors) != 1 {
		return false
	}
	excludedColor, ok := colorWord(selection.ExcludedColors[0])
	if !ok {
		return false
	}
	noun, ok := permanentSelectionNoun(selection.Kind)
	if !ok {
		return false
	}
	expected, ok := targetControllerSuffix("target non"+excludedColor+" "+noun, selection.Controller)
	if !ok {
		return false
	}
	return strings.EqualFold(text, expected)
}

func exactExcludedTypeTargetSyntax(text string, selection SelectionSyntax) bool {
	if selection.All ||
		selection.Attacking || selection.Blocking || selection.Tapped || selection.Untapped ||
		selection.Keyword != KeywordUnknown || selection.Zone != zone.None ||
		selection.MatchPower || selection.MatchToughness ||
		len(selection.Supertypes) != 0 ||
		len(selection.ColorsAny) != 0 || len(selection.ExcludedColors) != 0 ||
		len(selection.SubtypesAny) != 0 {
		return false
	}
	if !selectionRedundantRequiredNoun(selection) {
		return false
	}
	if len(selection.ExcludedTypes) != 1 {
		return false
	}
	excludedNoun, ok := permanentCardTypeNoun(selection.ExcludedTypes[0])
	if !ok {
		return false
	}
	noun, ok := permanentSelectionNoun(selection.Kind)
	if !ok {
		return false
	}
	// "another"/"other" prepend the self-exclusion word before "target" ("exile
	// another target nonland permanent", Oblivion Ring); the bare form keeps a
	// leading "target". permanentTargetSpec already lowers the Another predicate.
	prefix := "target"
	switch {
	case selection.Another:
		prefix = "another target"
	case selection.Other:
		prefix = "other target"
	default:
	}
	expected, ok := targetControllerSuffix(prefix+" non"+excludedNoun+" "+noun, selection.Controller)
	if !ok {
		return false
	}
	// A trailing "with mana value N or less/greater" qualifies the excluded-type
	// permanent ("target nonland permanent with mana value 3 or less"); every
	// permanent has a mana value, so the qualifier is faithful for any noun.
	// Power and toughness stay rejected above because they exist only on
	// creatures and would silently drop on a non-creature noun. The controller
	// clause already sits before this suffix in the reconstructed phrase.
	if selection.MatchManaValue {
		clause, ok := comparisonClauseWords("mana value", selection.ManaValue)
		if !ok {
			return false
		}
		expected += " with " + strings.Join(clause, " ")
	}
	return strings.EqualFold(text, expected)
}

// exactExcludedTypeColorTargetSyntax reconstructs the canonical Oracle phrase for
// a permanent target restricted by one excluded card type and one excluded color
// joined in a comma list ("target nonartifact, nonblack creature", Terror,
// Nekrataal, Shriekmaw) and compares it byte-exactly to the source text. The
// excluded card type renders first, then the excluded color, before the permanent
// noun, with an optional controller clause. It accepts exactly one excluded type
// and one excluded color on a redundant permanent noun, failing closed for every
// other qualifier so unsupported wordings keep failing the text-blind round-trip.
// Both exclusions lower to Selection.ExcludedTypes and Selection.ExcludedColors.
func exactExcludedTypeColorTargetSyntax(text string, selection SelectionSyntax) bool {
	if selection.All || selection.Another || selection.Other ||
		selection.Attacking || selection.Blocking || selection.Tapped || selection.Untapped ||
		selection.Keyword != KeywordUnknown || selection.ExcludedKeyword != KeywordUnknown ||
		selection.Zone != zone.None ||
		selection.MatchManaValue || selection.MatchPower || selection.MatchToughness ||
		selection.PowerLessThanSource || selection.PowerGreaterThanSource ||
		selection.Colorless || selection.Multicolored ||
		len(selection.Supertypes) != 0 || len(selection.ExcludedSupertypes) != 0 ||
		len(selection.ExcludedSubtypes) != 0 ||
		len(selection.ColorsAny) != 0 || len(selection.SubtypesAny) != 0 {
		return false
	}
	if !selectionRedundantRequiredNoun(selection) {
		return false
	}
	if len(selection.ExcludedTypes) != 1 || len(selection.ExcludedColors) != 1 {
		return false
	}
	excludedNoun, ok := permanentCardTypeNoun(selection.ExcludedTypes[0])
	if !ok {
		return false
	}
	excludedColor, ok := colorWord(selection.ExcludedColors[0])
	if !ok {
		return false
	}
	noun, ok := permanentSelectionNoun(selection.Kind)
	if !ok {
		return false
	}
	expected, ok := targetControllerSuffix("target non"+excludedNoun+", non"+excludedColor+" "+noun, selection.Controller)
	if !ok {
		return false
	}
	return strings.EqualFold(text, expected)
}

// a permanent target restricted by a single excluded supertype ("target nonbasic
// land", "target nonlegendary creature") and compares it byte-exactly to the
// source text. It accepts exactly one excluded supertype on a redundant permanent
// noun with an optional controller clause, failing closed for every other
// qualifier so unsupported wordings keep failing the text-blind round-trip.
func exactExcludedSupertypeTargetSyntax(text string, selection SelectionSyntax) bool {
	if selection.All || selection.Another || selection.Other ||
		selection.Attacking || selection.Blocking || selection.Tapped || selection.Untapped ||
		selection.Keyword != KeywordUnknown || selection.ExcludedKeyword != KeywordUnknown ||
		selection.Zone != zone.None ||
		selection.MatchManaValue || selection.MatchPower || selection.MatchToughness ||
		selection.PowerLessThanSource || selection.PowerGreaterThanSource ||
		selection.Colorless || selection.Multicolored ||
		len(selection.Supertypes) != 0 || len(selection.ExcludedTypes) != 0 ||
		len(selection.ColorsAny) != 0 || len(selection.ExcludedColors) != 0 ||
		len(selection.SubtypesAny) != 0 {
		return false
	}
	if !selectionRedundantRequiredNoun(selection) {
		return false
	}
	if len(selection.ExcludedSupertypes) != 1 {
		return false
	}
	excludedSuper, ok := supertypeWord(selection.ExcludedSupertypes[0])
	if !ok {
		return false
	}
	noun, ok := permanentSelectionNoun(selection.Kind)
	if !ok {
		return false
	}
	expected, ok := targetControllerSuffix("target non"+excludedSuper+" "+noun, selection.Controller)
	if !ok {
		return false
	}
	return strings.EqualFold(text, expected)
}

// exactExcludedSubtypeTargetSyntax reconstructs the canonical Oracle phrase for a
// permanent target restricted by a single excluded subtype ("target non-Spirit
// creature", "target non-Wall creature", "target non-Aura enchantment") and
// compares it byte-exactly to the source text. It mirrors
// exactExcludedSupertypeTargetSyntax: it accepts exactly one excluded subtype on
// a redundant permanent noun with an optional controller clause, failing closed
// for every other qualifier so unsupported wordings keep failing the text-blind
// round-trip. The single excluded subtype lowers to Selection.ExcludedSubtype.
func exactExcludedSubtypeTargetSyntax(text string, selection SelectionSyntax) bool {
	if selection.All || selection.Another || selection.Other ||
		selection.Attacking || selection.Blocking || selection.Tapped || selection.Untapped ||
		selection.Keyword != KeywordUnknown || selection.ExcludedKeyword != KeywordUnknown ||
		selection.Zone != zone.None ||
		selection.MatchManaValue || selection.MatchPower || selection.MatchToughness ||
		selection.PowerLessThanSource || selection.PowerGreaterThanSource ||
		selection.Colorless || selection.Multicolored ||
		len(selection.Supertypes) != 0 || len(selection.ExcludedSupertypes) != 0 ||
		len(selection.ExcludedTypes) != 0 ||
		len(selection.ColorsAny) != 0 || len(selection.ExcludedColors) != 0 ||
		len(selection.SubtypesAny) != 0 {
		return false
	}
	if !selectionRedundantRequiredNoun(selection) {
		return false
	}
	if len(selection.ExcludedSubtypes) != 1 {
		return false
	}
	noun, ok := permanentSelectionNoun(selection.Kind)
	if !ok {
		return false
	}
	expected, ok := targetControllerSuffix("target non-"+string(selection.ExcludedSubtypes[0])+" "+noun, selection.Controller)
	if !ok {
		return false
	}
	return strings.EqualFold(text, expected)
}

func targetSelectionHasUnsupportedQualifier(tokens []shared.Token, atoms Atoms) bool {
	dynStart, dynEnd, hasDyn := selectionManaValueDynamicSpan(tokens)
	for idx, token := range tokens {
		if hasDyn && idx >= dynStart && idx < dynEnd {
			continue
		}
		if token.Kind == shared.Integer || token.Kind == shared.Comma || token.Kind == shared.Slash ||
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
		"opponent", "opponent's", "opponents", "activated", "triggered", "source",
		"mana", "value", "power", "toughness", "equal", "less", "greater", "lesser",
		"battlefield", "graveyard", "hand", "library", "exile", "command",
		"historic", "single",
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
	for _, atom := range atoms.ExcludedSupertypes() {
		if covered(atom.Span) {
			return true
		}
	}
	for _, atom := range atoms.Subtypes() {
		if covered(atom.Span) {
			return true
		}
	}
	for _, atom := range atoms.ExcludedSubtypes() {
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

// orBackReferenceClauseFollowsAt reports whether the token at index i opens an
// anaphoric back-reference clause ("that creature", "that permanent", "it",
// "them", "those") that follows a sentence-level "or". Such a clause is a second
// alternative effect acting on the already-named target, not a member of a
// card-type union target ("target artifact or creature"), which always names a
// type noun rather than a demonstrative. Target scanning stops before it so the
// first effect's target noun phrase does not swallow the alternative clause.
func orBackReferenceClauseFollowsAt(tokens []shared.Token, i int) bool {
	if i >= len(tokens) {
		return false
	}
	return equalWord(tokens[i], "that") ||
		equalWord(tokens[i], "it") ||
		equalWord(tokens[i], "this") ||
		equalWord(tokens[i], "those") ||
		equalWord(tokens[i], "them") ||
		equalWord(tokens[i], "they")
}

func targetSyntaxEnd(tokens []shared.Token, atoms Atoms, start int) int {
	if end, ok := counterAbilityListEnd(tokens, start); ok {
		return end
	}
	end := start
	// A card-type or subtype union written as an Oxford-comma list ("artifact,
	// creature, or enchantment") embeds commas that would otherwise terminate
	// the target. Skip the scan past the whole list so the union's later members
	// join the target noun phrase; trailing qualifiers and the real clause
	// boundary are still found by the ordinary scan below.
	if unionEnd, ok := permanentUnionListEnd(tokens, atoms, start); ok {
		end = unionEnd
	}
	for end < len(tokens) {
		token := tokens[end]
		// The same-name restriction clause embeds "have ... you control", whose
		// "have" would otherwise terminate the target as an effect verb. Skip the
		// whole clause so the target noun phrase keeps it; parseTargets records
		// the restriction and strips the clause before parseSelection.
		if end > start && nameUniqueClauseStartsAt(tokens, end) {
			end += len(nameUniqueAmongControlledClause)
			continue
		}
		// The "that was dealt damage this turn" clause embeds "this turn", which
		// the trailing-duration check below would otherwise treat as the target's
		// boundary, splitting the clause off the noun phrase. Skip the whole
		// clause so the target keeps it; parseTargets strips it before
		// parseSelection and records the typed flag.
		if end > start && dealtDamageThisTurnClauseStartsAt(tokens, end) {
			end += len(dealtDamageThisTurnClause)
			continue
		}
		// A comma joining two negated card-type or subtype qualifiers ("non-Saga,
		// nonland permanent") is internal to the target's noun phrase, not a clause
		// boundary, so scanning continues past it to keep the whole filter on the
		// target. Requiring a negated qualifier immediately on both sides leaves an
		// ordinary comma terminating the target.
		if token.Kind == shared.Comma && negatedTypeListCommaAt(tokens, atoms, end, start) {
			end++
			continue
		}
		if token.Kind == shared.Comma || token.Kind == shared.Period || token.Kind == shared.Semicolon ||
			targetDestinationStartsAt(tokens, end) ||
			moveCounterDestinationStartsAt(tokens, end) ||
			(equalWord(token, "from") && end+1 < len(tokens) && equalWord(tokens[end+1], "combat")) ||
			equalWord(token, "unless") ||
			(equalWord(token, "equal") && end+1 < len(tokens) && equalWord(tokens[end+1], "to") &&
				(end < 2 || !equalWord(tokens[end-1], "or") || !equalWord(tokens[end-2], "than"))) ||
			(equalWord(token, "and") && end+2 < len(tokens) && equalWord(tokens[end+1], "you") && effectWordKind(tokens[end+2]) != EffectUnknown) ||
			selfDamageRiderFollowsAt(tokens, atoms, end) ||
			targetControllerDamageRiderFollowsAt(tokens, atoms, end) ||
			secondTargetDamageRiderFollowsAt(tokens, atoms, end) ||
			(equalWord(token, "and") && end+1 < len(tokens) &&
				(equalWord(tokens[end+1], "target") || equalWord(tokens[end+1], "targets"))) ||
			(equalWord(token, "or") && end+1 < len(tokens) && orBackReferenceClauseFollowsAt(tokens, end+1)) ||
			(equalWord(token, "or") && end+1 < len(tokens) && (equalWord(tokens[end+1], "remove") || equalWord(tokens[end+1], "removes"))) ||
			(equalWord(token, "and") && end+1 < len(tokens) && effectWordKind(tokens[end+1]) != EffectUnknown) ||
			(end > start && effectWordKind(token) != EffectUnknown) ||
			(end > start && equalWord(token, "becomes")) ||
			(end > start && cantBeBlockedThisTurnVerbAt(tokens, end)) ||
			(end > start && canAttackAsThoughDefenderVerbAt(tokens, end)) ||
			(end > start && cantBlockThisTurnVerbAt(tokens, end)) ||
			(end > start && cantAttackOrBlockThisTurnVerbAt(tokens, end)) ||
			(end > start && cantAttackThisTurnVerbAt(tokens, end)) ||
			(end > start && mustAttackTargetVerbAt(tokens, end)) ||
			(end > start && negatedNextUntapStepVerbAt(tokens, end)) ||
			(end > start && equalWord(token, "each") && end+1 < len(tokens) && effectWordKind(tokens[end+1]) != EffectUnknown) ||
			(equalWord(token, "until") && end+1 < len(tokens)) ||
			(equalWord(token, "this") && end+1 < len(tokens) && equalWord(tokens[end+1], "turn") &&
				thisTurnIsTrailingDuration(tokens, end)) ||
			(end > start && equalWord(token, "if")) ||
			(equalWord(token, "for") && effectWordsAt(tokens, end, "for", "as", "long", "as")) ||
			(equalWord(token, "as") && effectWordsAt(tokens, end, "as", "long", "as", "this")) {
			break
		}

		end++
	}

	return end
}

// negatedTypeListCommaAt reports whether the comma at index i joins two negated
// qualifiers within a single target noun phrase (e.g. the comma in "non-Saga,
// nonland permanent" or "nonartifact, nonblack creature"). Such a comma is
// internal to the noun phrase rather than a clause boundary, so target scanning
// continues past it. It requires a negated qualifier immediately on both sides,
// leaving an ordinary comma to terminate the target.
func negatedTypeListCommaAt(tokens []shared.Token, atoms Atoms, i, start int) bool {
	if i <= start || i+1 >= len(tokens) {
		return false
	}
	return isNegatedSelectionQualifier(tokens[i-1], atoms) && isNegatedSelectionQualifier(tokens[i+1], atoms)
}

// isNegatedSelectionQualifier reports whether the token begins a "non-<type>",
// "non-<subtype>", or "non<color>" exclusion qualifier ("nonland", "non-Saga",
// "nonblack").
func isNegatedSelectionQualifier(token shared.Token, atoms Atoms) bool {
	if _, ok := atoms.ExcludedCardTypeAt(token.Span); ok {
		return true
	}
	if _, ok := atoms.ExcludedSubtypeAt(token.Span); ok {
		return true
	}
	if _, ok := atoms.ExcludedColorAt(token.Span); ok {
		return true
	}
	return false
}

// is a trailing duration suffix on the target (e.g. "cast target ... card from
// your graveyard this turn") rather than part of an embedded amount clause (e.g.
// "the amount of life you lost this turn from your graveyard"). It is a trailing
// duration only when nothing of substance follows "turn": the next token is the
// clause boundary, the end of input, or a new effect verb.
func thisTurnIsTrailingDuration(tokens []shared.Token, i int) bool {
	after := i + 2
	if after >= len(tokens) {
		return true
	}
	next := tokens[after]
	if next.Kind == shared.Comma || next.Kind == shared.Period || next.Kind == shared.Semicolon {
		return true
	}
	return effectWordKind(next) != EffectUnknown
}

// self-damage rider begins at the "and" token at index i. Target scanning stops
// before the rider so the rider stays attached to the deal-damage clause (where
// the exactness gate reconstructs it and lowering emits a second damage to the
// source's controller) rather than being swallowed into the target noun phrase.
func selfDamageRiderFollowsAt(tokens []shared.Token, atoms Atoms, i int) bool {
	if i+4 >= len(tokens) || !equalWord(tokens[i], "and") {
		return false
	}
	if _, ok := effectNumber(tokens[i+1], atoms); !ok {
		return false
	}
	return equalWord(tokens[i+2], "damage") &&
		equalWord(tokens[i+3], "to") &&
		equalWord(tokens[i+4], "you")
}

// targetControllerDamageRiderFollowsAt reports whether a "... and N damage to
// that creature's controller/owner" rider begins at the "and" token at index i.
// Target scanning stops before the rider so the rider stays attached to the
// deal-damage clause (where the exactness gate reconstructs it and lowering
// emits a second damage to the primary target's controller or owner) rather
// than being swallowed into the target noun phrase. It accepts only the bounded
// "its controller/owner" and "that <noun>'s controller/owner" recipient phrases
// that immediately close the clause, so other "and ..." continuations are left
// to the ordinary scan.
func targetControllerDamageRiderFollowsAt(tokens []shared.Token, atoms Atoms, i int) bool {
	if i+4 >= len(tokens) || !equalWord(tokens[i], "and") {
		return false
	}
	if _, ok := effectNumber(tokens[i+1], atoms); !ok {
		return false
	}
	if !equalWord(tokens[i+2], "damage") || !equalWord(tokens[i+3], "to") {
		return false
	}
	for _, recipientLen := range []int{2, 3} {
		recipientEnd := i + 4 + recipientLen
		if recipientEnd > len(tokens) {
			continue
		}
		if recipientEnd < len(tokens) && tokens[recipientEnd].Kind != shared.Period {
			continue
		}
		if _, ok := referencedControllerOwnerRecipient(tokens[i+4 : recipientEnd]); ok {
			return true
		}
	}
	return false
}

// secondTargetDamageRiderFollowsAt reports whether a "... and N damage to target
// ..." rider — a second damage clause naming its own target — begins at the
// "and" token at index i. Target scanning stops before the rider so the first
// target's noun phrase does not swallow the second clause; the two targets are
// then parsed independently and lowering emits one Damage instruction each. It
// matches only the bounded "and <number-or-X> damage to target/targets" lead-in,
// so other "and ..." continuations are left to the ordinary scan.
func secondTargetDamageRiderFollowsAt(tokens []shared.Token, atoms Atoms, i int) bool {
	if i+4 >= len(tokens) || !equalWord(tokens[i], "and") {
		return false
	}
	if _, ok := effectNumber(tokens[i+1], atoms); !ok && !equalWord(tokens[i+1], "x") {
		return false
	}
	if !equalWord(tokens[i+2], "damage") || !equalWord(tokens[i+3], "to") {
		return false
	}
	// The second target opens with "target"/"targets" ("... and 2 damage to
	// target player") or with "any other" ("... and 1 damage to any other
	// target"), whose "other" distinctness qualifier precedes its own "target"
	// noun. Recognizing these openers stops the first target's noun phrase at
	// this rider instead of greedily absorbing the second clause. A bare "any
	// target" second clause is intentionally excluded; it stays unrepresented.
	if equalWord(tokens[i+4], "target") || equalWord(tokens[i+4], "targets") {
		return true
	}
	return equalWord(tokens[i+4], "any") && i+5 < len(tokens) && equalWord(tokens[i+5], "other")
}

// permanentUnionListEnd recognizes a permanent target whose noun phrase is a
// union of card-type or subtype nouns written as an Oxford-comma list
// ("artifact, creature, or enchantment", "Skeleton, Vampire, or Zombie")
// beginning at start. Each element is a single card-type or subtype noun
// separated by commas and a closing "or". It returns the index just past the
// final element and ok=true only when the list holds at least two elements, uses
// at least one comma, and closes with an "or"-joined element, so the ordinary
// single-noun target scan and the comma-free "X or Y" union are unaffected.
// Per-element qualifiers and non-noun words fail closed.
func permanentUnionListEnd(tokens []shared.Token, atoms Atoms, start int) (int, bool) {
	i := start
	elements := 0
	end := start
	sawComma := false
	prevSeparatorOr := false
	lastJoinedByOr := false
	for i < len(tokens) {
		if !unionMemberNoun(tokens[i], atoms) {
			break
		}
		elements++
		i++
		end = i
		lastJoinedByOr = prevSeparatorOr
		prevSeparatorOr = false
		consumedSeparator := false
		if i < len(tokens) && tokens[i].Kind == shared.Comma {
			sawComma = true
			i++
			consumedSeparator = true
		}
		if i+2 < len(tokens) && equalWord(tokens[i], "and") &&
			tokens[i+1].Kind == shared.Slash && equalWord(tokens[i+2], "or") {
			prevSeparatorOr = true
			i += 3
			consumedSeparator = true
		} else if i < len(tokens) && equalWord(tokens[i], "or") {
			prevSeparatorOr = true
			i++
			consumedSeparator = true
		}
		if !consumedSeparator {
			break
		}
	}
	if elements >= 2 && sawComma && lastJoinedByOr {
		return end, true
	}
	return start, false
}

// unionMemberNoun reports whether the token names a permanent card type or a
// subtype, the only two element kinds a permanent type/subtype union admits.
func unionMemberNoun(token shared.Token, atoms Atoms) bool {
	if _, ok := atoms.CardTypeAt(token.Span); ok {
		return true
	}
	_, ok := atoms.SubtypeAt(token.Span)
	return ok
}

func targetDestinationStartsAt(tokens []shared.Token, index int) bool {
	if index < 0 || index > len(tokens) {
		return false
	}
	for _, phrase := range [][]string{
		{"to", "its", "owner's", "hand"},
		{"to", "their", "owners", "'", "hands"},
		{"to", "your", "hand"},
		{"to", "their", "hand"},
		{"to", "their", "hands"},
		{"to", "the", "battlefield"},
		{"onto", "the", "battlefield"},
		{"into", "your", "graveyard"},
		{"into", "your", "library"},
		{"on", "top", "of", "your", "library"},
		{"on", "the", "top", "of", "your", "library"},
		{"on", "bottom", "of", "your", "library"},
		{"on", "the", "bottom", "of", "your", "library"},
		{"on", "top", "of", "its", "owner's", "library"},
		{"on", "the", "top", "of", "its", "owner's", "library"},
		{"on", "bottom", "of", "its", "owner's", "library"},
		{"on", "the", "bottom", "of", "its", "owner's", "library"},
	} {
		if _, ok := cutTokenPrefix(tokens[index:], phrase...); ok {
			return true
		}
	}
	return false
}

// moveCounterDestinationStartsAt reports whether a counter-move destination
// target phrase ("onto target ...", "onto a second target ...", "onto another
// target ...", "onto other target ...") begins at index. Source-target scanning
// stops before it so the first target ("Move a counter from target permanent you
// control ...") does not swallow the second, "onto"-introduced destination
// target; the two targets are then parsed independently and lowering emits a
// single MoveCounters reading the source target and placing onto the
// destination. The "onto the battlefield" zone destination is handled by
// targetDestinationStartsAt and is not a target, so it is excluded here.
func moveCounterDestinationStartsAt(tokens []shared.Token, index int) bool {
	if index < 0 || index > len(tokens) {
		return false
	}
	for _, phrase := range [][]string{
		{"onto", "target"},
		{"onto", "targets"},
		{"onto", "a", "target"},
		{"onto", "a", "second", "target"},
		{"onto", "another", "target"},
		{"onto", "other", "target"},
	} {
		if _, ok := cutTokenPrefix(tokens[index:], phrase...); ok {
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

// stackObjectSelectionKind recognizes the explicit spell/ability stack-object
// selection phrasings and reports the matching selection kind.
func stackObjectSelectionKind(words []string) (SelectionKind, bool) {
	switch {
	case slices.Equal(words, []string{"activated", "ability"}):
		return SelectionActivatedAbility, true
	case slices.Equal(words, []string{"triggered", "ability"}):
		return SelectionTriggeredAbility, true
	case slices.Equal(words, []string{"activated", "or", "triggered", "ability"}):
		return SelectionActivatedOrTriggeredAbility, true
	case slices.Equal(words, []string{"spell", "or", "ability"}):
		return SelectionSpellActivatedOrTriggeredAbility, true
	case slices.Equal(words, []string{"spell", "activated", "ability", "or", "triggered", "ability"}):
		return SelectionSpellActivatedOrTriggeredAbility, true
	default:
		return SelectionUnknown, false
	}
}

// splitSelectionNamedTail captures a "named <Name>" qualifier from a selection's
// tokens, returning the verbatim card name, the tokens preceding "named", and
// true when the qualifier is present. The name is joined from the source tokens
// after the first "named" word through the selection end (excluding a trailing
// period), mirroring parseTokenName. Splitting the name off keeps the trailing
// name words ("Trustworthy Scout") from being misread as subtypes; the
// byte-exact search reconstruction rebuilds and validates the qualifier, so a
// spurious capture fails closed there.
func splitSelectionNamedTail(tokens []shared.Token) (name string, head []shared.Token, ok bool) {
	named := -1
	for i, token := range tokens {
		if equalWord(token, "named") {
			named = i
			break
		}
	}
	if named < 0 || named == 0 {
		return "", nil, false
	}
	nameTokens := tokens[named+1:]
	if len(nameTokens) > 0 && nameTokens[len(nameTokens)-1].Kind == shared.Period {
		nameTokens = nameTokens[:len(nameTokens)-1]
	}
	if len(nameTokens) == 0 {
		return "", nil, false
	}
	return joinedEffectText(nameTokens), tokens[:named], true
}

// splitSelectionNameUniqueTail captures a trailing "that doesn't have the same
// name as another permanent you control" relative clause from a selection's
// tokens, returning the tokens preceding the clause and true when the clause is
// present (Yenna, Redtooth Regent). Splitting the clause off keeps its words
// from being misread by parseSelection or rejected as an unsupported qualifier;
// the caller records the restriction on SelectionSyntax.NameUniqueAmongControlled.
func splitSelectionNameUniqueTail(tokens []shared.Token) (head []shared.Token, ok bool) {
	if len(tokens) <= len(nameUniqueAmongControlledClause) {
		return nil, false
	}
	offset := len(tokens) - len(nameUniqueAmongControlledClause)
	if !nameUniqueClauseStartsAt(tokens, offset) {
		return nil, false
	}
	return tokens[:offset], true
}

// splitSelectionOtherThanSelfTail strips a trailing "other than <source name>"
// self-exclusion clause from a target's selection tokens ("target creature you
// control other than Rosie Cotton"), returning the head tokens that name the
// permanent. The excluded object must be the card's own name, matched through the
// atom self-name spans so the parser owns the name spelling, and must run to the
// end of the selection. It fails closed for any other trailing wording.
func splitSelectionOtherThanSelfTail(tokens []shared.Token, atoms Atoms) (head []shared.Token, ok bool) {
	for i := 0; i+2 < len(tokens); i++ {
		if !equalWord(tokens[i], "other") || !equalWord(tokens[i+1], "than") {
			continue
		}
		nameTokens := tokens[i+2:]
		span, found := atoms.SelfNameSpanStartingAt(nameTokens[0].Span)
		if !found || tokenCountForSpan(nameTokens, span) != len(nameTokens) {
			return nil, false
		}
		return tokens[:i], true
	}
	return nil, false
}

// nameUniqueAmongControlledClause is the relative clause that restricts a target
// to a permanent whose name is unique among the controller's permanents
// (Yenna, Redtooth Regent: "that doesn't have the same name as another
// permanent you control").
var nameUniqueAmongControlledClause = []string{
	"that", "doesn't", "have", "the", "same", "name",
	"as", "another", "permanent", "you", "control",
}

// nameUniqueAmongControlledClauseText is the canonical spelling of the
// name-uniqueness clause, used to reconstruct the exact target phrase.
const nameUniqueAmongControlledClauseText = "that doesn't have the same name as another permanent you control"

// nameUniqueClauseStartsAt reports whether nameUniqueAmongControlledClause
// begins at index i in tokens.
func nameUniqueClauseStartsAt(tokens []shared.Token, i int) bool {
	if i < 0 || i+len(nameUniqueAmongControlledClause) > len(tokens) {
		return false
	}
	for j, word := range nameUniqueAmongControlledClause {
		if !equalWord(tokens[i+j], word) {
			return false
		}
	}
	return true
}

// dealtDamageThisTurnClause is the relative clause that restricts a target to a
// permanent that was dealt damage during the current turn (Fatal Blow: "target
// creature that was dealt damage this turn").
var dealtDamageThisTurnClause = []string{
	"that", "was", "dealt", "damage", "this", "turn",
}

// dealtDamageThisTurnClauseText is the canonical spelling of the dealt-damage
// clause, used to reconstruct the exact target phrase.
const dealtDamageThisTurnClauseText = "that was dealt damage this turn"

// dealtDamageThisTurnClauseStartsAt reports whether dealtDamageThisTurnClause
// begins at index i in tokens.
func dealtDamageThisTurnClauseStartsAt(tokens []shared.Token, i int) bool {
	if i < 0 || i+len(dealtDamageThisTurnClause) > len(tokens) {
		return false
	}
	for j, word := range dealtDamageThisTurnClause {
		if !equalWord(tokens[i+j], word) {
			return false
		}
	}
	return true
}

// splitSelectionDealtDamageThisTurnTail strips a trailing "that was dealt damage
// this turn" clause from a target's selection tokens, returning the head tokens
// that name the permanent. It fails closed unless the clause runs to the end of
// the selection.
func splitSelectionDealtDamageThisTurnTail(tokens []shared.Token) (head []shared.Token, ok bool) {
	if len(tokens) <= len(dealtDamageThisTurnClause) {
		return nil, false
	}
	offset := len(tokens) - len(dealtDamageThisTurnClause)
	if !dealtDamageThisTurnClauseStartsAt(tokens, offset) {
		return nil, false
	}
	return tokens[:offset], true
}

// sameNameGroupClauseLength is the fixed token count of the same-name group
// clause "and all other <group> with the same name as that <noun>": the eleven
// words with one group noun and one back-reference noun.
const sameNameGroupClauseLength = 11

// splitSelectionSameNameGroupTail strips a trailing "and all other <group> with
// the same name as that <noun>" clause from a single-permanent target's
// selection tokens, returning the head tokens that name the target and the
// same-name group the effect additionally affects ("Destroy target nonland
// permanent and all other permanents with the same name as that permanent",
// Maelstrom Pulse; the Echoing cycle). The group noun is either the bare
// "permanents" (no card-type restriction) or one card-type plural; the trailing
// back-reference noun must be the matching singular permanent noun. It fails
// closed unless the clause runs to the end of the selection and leaves a
// non-empty head naming the target.
func splitSelectionSameNameGroupTail(tokens []shared.Token) (head []shared.Token, group *SameNameGroupSyntax, ok bool) {
	if len(tokens) <= sameNameGroupClauseLength {
		return nil, nil, false
	}
	c := len(tokens) - sameNameGroupClauseLength
	fixed := []struct {
		offset int
		word   string
	}{
		{0, "and"}, {1, "all"}, {2, "other"},
		{4, "with"}, {5, "the"}, {6, "same"}, {7, "name"}, {8, "as"}, {9, "that"},
	}
	for _, f := range fixed {
		if !equalWord(tokens[c+f.offset], f.word) {
			return nil, nil, false
		}
	}
	groupTypes, ok := sameNameGroupNounTypes(tokens[c+3])
	if !ok {
		return nil, nil, false
	}
	if !sameNameBackReferenceNoun(tokens[c+10]) {
		return nil, nil, false
	}
	clauseTokens := tokens[c:]
	return tokens[:c], &SameNameGroupSyntax{
		GroupTypes: groupTypes,
		Text:       joinedEffectText(clauseTokens),
		Span:       shared.SpanOf(clauseTokens),
	}, true
}

// sameNameGroupNounTypes maps the plural group noun of a same-name group clause
// to the card types it restricts the same-name permanents to. The bare
// "permanents" noun imposes no card-type restriction (an empty slice); a single
// card-type plural ("lands", "artifacts", "enchantments", ...) restricts to that
// type. It fails closed for any other word.
func sameNameGroupNounTypes(token shared.Token) ([]CardType, bool) {
	if equalWord(token, "permanents") {
		return nil, true
	}
	if cardType, ok := recognizeCardTypeWord(token.Text); ok {
		return []CardType{cardType}, true
	}
	return nil, false
}

// sameNameBackReferenceNoun reports whether token is the singular permanent noun
// that ends a same-name group clause ("... as that permanent" / "... as that
// land"). It accepts the bare "permanent" and every singular card-type noun.
func sameNameBackReferenceNoun(token shared.Token) bool {
	if equalWord(token, "permanent") {
		return true
	}
	_, ok := recognizeCardTypeWord(token.Text)
	return ok
}

func parseSelection(tokens []shared.Token, atoms Atoms) SelectionSyntax {
	if recognized, ok := counterAbilitySelectionSyntax(tokens, shared.SpanOf(tokens), joinedEffectText(tokens)); ok {
		return recognized
	}
	selection := SelectionSyntax{Span: shared.SpanOf(tokens), Text: joinedEffectText(tokens)}
	if name, head, ok := splitSelectionNamedTail(tokens); ok {
		selection.RequiredName = name
		tokens = head
	}
	words := normalizedWords(tokens)
	if kind, ok := stackObjectSelectionKind(words); ok {
		selection.Kind = kind
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
		if supertype, ok := atoms.ExcludedSupertypeAt(token.Span); ok && !slices.Contains(selection.ExcludedSupertypes, supertype) {
			selection.ExcludedSupertypes = append(selection.ExcludedSupertypes, supertype)
		}
		if qualifier, ok := atoms.ColorQualifierAt(token.Span); ok {
			switch qualifier {
			case ColorQualifierColorless:
				selection.Colorless = true
			case ColorQualifierMulticolored:
				selection.Multicolored = true
			default:
			}
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
	selection.ExcludedSubtypes = atoms.ExcludedSubtypesIn(span)
	if relation, ok := atoms.ControllerIn(span); ok {
		switch relation {
		case ControllerRelationYouControl:
			selection.Controller = SelectionControllerYou
		case ControllerRelationOpponentControls:
			selection.Controller = SelectionControllerOpponent
		case ControllerRelationEachOpponentControls:
			selection.Controller = SelectionControllerOpponent
			selection.OpponentEach = true
		case ControllerRelationYouDontControl:
			selection.Controller = SelectionControllerNotYou
		default:
		}
	}
	selection.Zone = firstZone(atoms, span, ZoneRoleFrom)
	if selection.Zone == zone.None {
		selection.Zone = firstZone(atoms, span, ZoneRolePlain)
	}
	if selection.Zone == zone.None &&
		len(words) >= 3 &&
		slices.Equal(words[len(words)-3:], []string{"in", "your", "graveyard"}) {
		selection.Zone = zone.Graveyard
	}
	switch {
	case effectContainsWords(words, "your", "graveyard"):
		selection.Controller = SelectionControllerYou
	case effectContainsWords(words, "opponent's", "graveyard"):
		selection.Controller = SelectionControllerOpponent
	default:
	}
	// "from a single graveyard" restricts every chosen card to one and the same
	// graveyard. The "single" qualifier keeps the any-graveyard owner relation;
	// it is recorded as its own flag so the byte-exact reconstruction can rebuild
	// the verbatim wording and lowering can carry the same-graveyard restriction.
	if selection.Zone == zone.Graveyard && effectContainsWords(words, "single", "graveyard") {
		selection.SingleGraveyard = true
	}
	selection.All = slices.Contains(words, "all")
	selection.Colored = effectContainsWords(words, "one", "or", "more", "colors") ||
		effectContainsWords(words, "one", "or", "more", "color")
	selection.Historic = slices.Contains(words, "historic")
	selection.Another = atoms.SelectionFlagIn(span, SelectionFlagAnother)
	selection.Other = atoms.SelectionFlagIn(span, SelectionFlagOther)
	selection.Attacking = atoms.SelectionFlagIn(span, SelectionFlagAttacking)
	selection.Blocking = atoms.SelectionFlagIn(span, SelectionFlagBlocking)
	selection.Tapped = atoms.SelectionFlagIn(span, SelectionFlagTapped)
	selection.Untapped = atoms.SelectionFlagIn(span, SelectionFlagUntapped)
	selection.NonToken = atoms.SelectionFlagIn(span, SelectionFlagNonToken)
	selection.TokenOnly = atoms.SelectionFlagIn(span, SelectionFlagToken)
	selection.EnteredThisTurn = effectContainsWords(words, "that", "entered", "this", "turn") ||
		effectContainsWords(words, "that", "entered", "the", "battlefield", "this", "turn")
	selection.DealtDamageThisTurn = effectContainsWords(words, "that", "was", "dealt", "damage", "this", "turn")
	if slices.Contains(words, "any") && selection.Kind == SelectionUnknown {
		selection.Kind = SelectionAny
	}
	if keyword, ok := atoms.KeywordSelectorIn(span, false); ok {
		selection.Keyword = keyword.Keyword
	}
	if keyword, ok := atoms.KeywordSelectorIn(span, true); ok {
		selection.ExcludedKeyword = keyword.Keyword
	}
	if match, ok := selectionCounterQualifier(tokens); ok {
		switch {
		case match.KindAbsent:
			selection.CounterKindAbsent = true
			selection.CounterKind = match.Kind
		case match.Absent:
			selection.CounterAbsent = true
		case match.Any:
			selection.CounterRequired = true
			selection.CounterAny = true
		default:
			selection.CounterRequired = true
			selection.CounterKind = match.Kind
		}
	}
	if (selection.Kind == SelectionPlayer && slices.Equal(words, []string{"player", "or", "planeswalker"})) ||
		(selection.Kind == SelectionOpponent && slices.Equal(words, []string{"opponent", "or", "planeswalker"})) {
		selection.PlayerOrPlaneswalker = true
	}
	parseSelectionChosenTypeQualifier(words, &selection)
	if !parseSelectionNumbers(tokens, atoms, &selection) {
		return SelectionSyntax{Span: span, Text: joinedEffectText(tokens)}
	}
	if alts, ok := disjunctiveSelectionAlternatives(tokens, atoms); ok {
		selection.Alternatives = alts
		selection.Kind = SelectionUnknown
		selection.RequiredTypesAny = nil
		selection.ExcludedTypes = nil
		selection.Supertypes = nil
		selection.ExcludedSupertypes = nil
		selection.SubtypesAny = nil
		selection.ExcludedSubtypes = nil
		selection.ColorsAny = nil
		selection.ExcludedColors = nil
		selection.Colorless = false
		selection.Multicolored = false
		selection.Colored = false
	}
	return selection
}

// disjunctiveSelectionAlternatives splits a selection phrase joined by a single
// top-level "or" into two type-dimension alternatives, but only when flattening
// the two sides into one selection would lose meaning. parseSelection otherwise
// merges every card type, supertype, and subtype across the whole phrase into a
// single selection. That flattening is correct for a pure type or subtype union
// ("artifact or creature", "Forest or Plains") because the runtime treats those
// dimensions as any-of, but it is lossy when the two sides carry different
// supertypes ("creature or basic land card", "basic land card or Gate card"):
// the runtime treats supertypes as all-of, so the flattened selection would
// force every match to satisfy both sides' supertypes at once. Only in that
// supertype-mismatch case does each side become its own alternative so the
// lowering can build a Selection.AnyOf. It fails closed for any phrase that is
// not exactly two supertype-divergent sides, leaving every other selection's
// existing flattened parse unchanged.
func disjunctiveSelectionAlternatives(tokens []shared.Token, atoms Atoms) ([]SelectionSyntax, bool) {
	if shared.TopLevelIndex(tokens, shared.Comma) >= 0 {
		return nil, false
	}
	orIndex := -1
	for i, token := range tokens {
		if !equalWord(token, "or") {
			continue
		}
		if i > 0 && tokens[i-1].Kind == shared.Slash {
			// "and/or" lexes as "and", "/", "or" and forms an InclusiveOneOfEach
			// union, not a single-choice disjunction; leave it to that handling.
			return nil, false
		}
		if orIndex >= 0 {
			return nil, false
		}
		orIndex = i
	}
	if orIndex <= 0 || orIndex >= len(tokens)-1 {
		return nil, false
	}
	left, ok := disjunctSelectionSide(tokens[:orIndex], atoms)
	if !ok {
		return nil, false
	}
	right, ok := disjunctSelectionSide(tokens[orIndex+1:], atoms)
	if !ok {
		return nil, false
	}
	if slices.Equal(left.Supertypes, right.Supertypes) {
		return nil, false
	}
	// A differing supertype is lossy only when it cannot distribute across both
	// sides. When neither side names a genuine card type ("basic Forest or
	// Plains [card]"), both sides are subtype-only and the leading supertype
	// belongs to the whole card, so the flattened "basic" + subtype-union parse
	// is correct and must be left intact. Splitting is required only when at
	// least one side carries a real card-type kind ("creature or basic land
	// card", "basic land card or Gate card"), where flattening would force every
	// type alternative to satisfy the other's supertype.
	if !disjunctSideTypedKind(left) && !disjunctSideTypedKind(right) {
		return nil, false
	}
	return []SelectionSyntax{left, right}, true
}

// disjunctSideTypedKind reports whether a disjunction side names a real card
// type kind (rather than a card matched purely by subtype), the signal that a
// differing supertype across the disjunction cannot distribute and the
// flattened parse would be lossy.
func disjunctSideTypedKind(side SelectionSyntax) bool {
	switch side.Kind {
	case SelectionCreature, SelectionLand, SelectionArtifact,
		SelectionEnchantment, SelectionPlaneswalker, SelectionPermanent:
		return true
	default:
		return false
	}
}

// disjunctSelectionSide parses one side of a disjunctive selection into a
// type-dimension-only SelectionSyntax, dropping any leading article so "a Gate
// card" parses like "Gate card". A subtype-only side ("Bird", "Gate") implies a
// card matched by its subtype, so its kind becomes SelectionCard. It fails
// closed unless the side names a card-type or permanent kind the search lowering
// can express, so an alternative never lowers to an empty match.
func disjunctSelectionSide(tokens []shared.Token, atoms Atoms) (SelectionSyntax, bool) {
	for len(tokens) > 0 &&
		(equalWord(tokens[0], "a") || equalWord(tokens[0], "an") || equalWord(tokens[0], "the")) {
		tokens = tokens[1:]
	}
	if len(tokens) == 0 {
		return SelectionSyntax{}, false
	}
	parsed := parseSelection(tokens, atoms)
	side := SelectionSyntax{
		Kind:               parsed.Kind,
		RequiredTypesAny:   parsed.RequiredTypesAny,
		ExcludedTypes:      parsed.ExcludedTypes,
		Supertypes:         parsed.Supertypes,
		ExcludedSupertypes: parsed.ExcludedSupertypes,
		SubtypesAny:        parsed.SubtypesAny,
		ExcludedSubtypes:   parsed.ExcludedSubtypes,
		ColorsAny:          parsed.ColorsAny,
		ExcludedColors:     parsed.ExcludedColors,
		Colorless:          parsed.Colorless,
		Multicolored:       parsed.Multicolored,
	}
	if side.Kind == SelectionUnknown && len(side.SubtypesAny) > 0 && len(side.RequiredTypesAny) == 0 {
		side.Kind = SelectionCard
	}
	if !disjunctSideExpressible(side) {
		return SelectionSyntax{}, false
	}
	return side, true
}

// disjunctSideExpressible reports whether a disjunction side names a card-type
// or permanent kind the search filter can carry, the precondition for splitting
// a selection into alternatives.
func disjunctSideExpressible(side SelectionSyntax) bool {
	switch side.Kind {
	case SelectionCard, SelectionCreature, SelectionLand, SelectionArtifact,
		SelectionEnchantment, SelectionPlaneswalker, SelectionPermanent:
		return true
	default:
		return false
	}
}

// parseSelectionChosenTypeQualifier records a trailing "of the chosen type" /
// "that aren't of the chosen type" qualifier on a selection. The matched
// permanents must (positive) or must not (negated) share the creature subtype a
// "Choose a creature type." effect selects earlier in the same resolution
// (Kindred Dominance: "Destroy all creatures that aren't of the chosen type.").
// The resolution-time chosen type is the only sense a one-shot effect body's "the
// chosen type" can denote; entry-time anthem groups take a separate static-subject
// parse path. It fails closed for any other trailing wording so unrelated
// selections are untouched.
func parseSelectionChosenTypeQualifier(words []string, selection *SelectionSyntax) {
	switch {
	case hasWordSuffix(words, "that", "aren't", "of", "the", "chosen", "type"):
		selection.SubtypeFromChosenTypeExcluded = true
	case hasWordSuffix(words, "of", "the", "chosen", "type"):
		selection.SubtypeFromChosenType = true
	default:
	}
}

// counterQualifierMatch records a parsed "with a/an <kind> counter on it/them"
// qualifier: Kind names the required counter, Any marks the kind-agnostic "with
// a counter on it" form (Rishkar) where any counter satisfies the filter, Absent
// marks the negated "with no counters on it/them" form (Damning Verdict) where
// the permanent must carry no counters, KindAbsent marks the kind-specific
// negated "without a <kind> counter on it/them" form (Wave Goodbye) where the
// permanent must carry no counter of Kind, and End is the token index just past
// the qualifier.
type counterQualifierMatch struct {
	Kind       counter.Kind
	Any        bool
	Absent     bool
	KindAbsent bool
	End        int
}

// counterQualifierKind detects a "with [a/an] <kind> counter(s) on it/them"
// qualifier or the negated "with no counters on it/them" qualifier beginning at
// index start and returns the parsed qualifier together with whether the phrase
// matched. The article is optional so the plural group form "with +1/+1
// counters on them" (Sphere Grid) matches alongside the singular "with a +1/+1
// counter on it". It fails closed when the phrase is absent so unrelated
// wordings keep their existing handling.
func counterQualifierKind(tokens []shared.Token, start int) (counterQualifierMatch, bool) {
	if effectWordsAt(tokens, start, "with", "no") {
		return noCounterQualifier(tokens, start)
	}
	if effectWordsAt(tokens, start, "without", "a") || effectWordsAt(tokens, start, "without", "an") {
		return excludedCounterQualifier(tokens, start)
	}
	if !effectWordsAt(tokens, start, "with") {
		return counterQualifierMatch{}, false
	}
	counterStart := start + 1
	if effectWordsAt(tokens, start, "with", "a") || effectWordsAt(tokens, start, "with", "an") {
		counterStart = start + 2
	}
	counterIndex := counterStart
	for counterIndex < len(tokens) &&
		!equalWord(tokens[counterIndex], "counter") && !equalWord(tokens[counterIndex], "counters") {
		counterIndex++
	}
	if counterIndex >= len(tokens) {
		return counterQualifierMatch{}, false
	}
	if !effectWordsAt(tokens, counterIndex+1, "on", "it") &&
		!effectWordsAt(tokens, counterIndex+1, "on", "them") {
		return counterQualifierMatch{}, false
	}
	if counterIndex == counterStart {
		// "with a counter on it/them" names no counter kind, so the qualifier
		// matches a permanent carrying a counter of any kind (Rishkar's "Each
		// creature you control with a counter on it has ...").
		return counterQualifierMatch{Any: true, End: counterIndex + 3}, true
	}
	kind, _, ok := counterNameBefore(tokens, counterIndex)
	if !ok {
		return counterQualifierMatch{}, false
	}
	return counterQualifierMatch{Kind: kind, End: counterIndex + 3}, true
}

// noCounterQualifier detects the negated "with no counter(s) on it/them"
// qualifier beginning at index start ("Destroy all creatures with no counters on
// them."). Only the bare, kind-agnostic negation is recognized: a named "with no
// <kind> counters" form is left unmatched so it fails closed rather than dropping
// the counter kind. It mirrors counterQualifierKind's "on it"/"on them" pronoun
// handling.
func noCounterQualifier(tokens []shared.Token, start int) (counterQualifierMatch, bool) {
	counterIndex := start + 2
	if counterIndex >= len(tokens) ||
		(!equalWord(tokens[counterIndex], "counter") && !equalWord(tokens[counterIndex], "counters")) {
		return counterQualifierMatch{}, false
	}
	if !effectWordsAt(tokens, counterIndex+1, "on", "it") &&
		!effectWordsAt(tokens, counterIndex+1, "on", "them") {
		return counterQualifierMatch{}, false
	}
	return counterQualifierMatch{Absent: true, End: counterIndex + 3}, true
}

// excludedCounterQualifier detects the kind-specific negated "without a/an
// <kind> counter on it/them" qualifier beginning at index start ("Return each
// creature without a +1/+1 counter on it to its owner's hand."). It requires a
// named counter kind: the kind-agnostic "without a counter" form names no kind
// and fails closed rather than dropping the restriction. It mirrors
// counterQualifierKind's "on it"/"on them" pronoun handling.
func excludedCounterQualifier(tokens []shared.Token, start int) (counterQualifierMatch, bool) {
	counterIndex := start + 2
	for counterIndex < len(tokens) &&
		!equalWord(tokens[counterIndex], "counter") && !equalWord(tokens[counterIndex], "counters") {
		counterIndex++
	}
	if counterIndex >= len(tokens) || counterIndex == start+2 {
		return counterQualifierMatch{}, false
	}
	if !effectWordsAt(tokens, counterIndex+1, "on", "it") &&
		!effectWordsAt(tokens, counterIndex+1, "on", "them") {
		return counterQualifierMatch{}, false
	}
	kind, _, ok := counterNameBefore(tokens, counterIndex)
	if !ok {
		return counterQualifierMatch{}, false
	}
	return counterQualifierMatch{Kind: kind, KindAbsent: true, End: counterIndex + 3}, true
}

// selectionCounterQualifier scans tokens for a "with a <kind> counter on
// it/them" qualifier (or its negated "with no counters on it/them" form) anywhere
// in a selection phrase and returns the parsed qualifier (its counter kind, the
// kind-agnostic "any counter" flag, or the "no counters" absence flag) together
// with whether any such qualifier matched.
func selectionCounterQualifier(tokens []shared.Token) (counterQualifierMatch, bool) {
	for i := range tokens {
		if match, found := counterQualifierKind(tokens, i); found {
			return match, true
		}
	}
	return counterQualifierMatch{}, false
}

func parseSelectionNumbers(tokens []shared.Token, atoms Atoms, selection *SelectionSyntax) bool {
	for i := range tokens {
		if i+2 < len(tokens) && effectWordsAt(tokens, i, "mana", "value") {
			if i >= 1 && equalWord(tokens[i-1], "total") {
				// "total mana value N or less" bounds the combined mana value of
				// the whole chosen set, not each card. Record it on the dedicated
				// total fields so lowering never mistakes it for a per-card filter.
				comparison, ok := parseSelectionNumberComparison(tokens[i+2:], atoms)
				if !ok {
					return false
				}
				selection.TotalManaValue = comparison
				selection.MatchTotalManaValue = true
				continue
			}
			if i+4 < len(tokens) && equalWord(tokens[i+2], "X") &&
				equalWord(tokens[i+3], "or") && equalWord(tokens[i+4], "less") {
				// "mana value X or less" bounds the match by the spell's chosen
				// {X}, which no fixed comparison can express. Record the operator
				// and flag the X-derived bound; lowering resolves it from X.
				selection.ManaValue = compare.Int{Op: compare.LessOrEqual}
				selection.MatchManaValue = true
				selection.ManaValueX = true
				continue
			}
			if kind, _, ok := parseSelectionManaValueDynamic(tokens, i+2); ok {
				// "mana value less than or equal to the amount of life you
				// (lost|gained) this turn" bounds the match by a turn-event life
				// total (Betor, Ancestor's Voice). Record the dynamic kind on its
				// own field; only the graveyard-card target reconstruction renders
				// it, so other contexts keep failing closed.
				selection.ManaValueDynamic = kind
				continue
			}
			comparison, ok := parseSelectionNumberComparison(tokens[i+2:], atoms)
			if !ok {
				return false
			}
			selection.ManaValue = comparison
			selection.MatchManaValue = true
			continue
		}
		if equalWord(tokens[i], "power") {
			if i >= 1 && equalWord(tokens[i-1], "lesser") {
				// "with lesser power" compares the match to the source
				// permanent's power, not a fixed number (Mentor). Record the
				// relative qualifier and skip the fixed-comparison parse.
				selection.PowerLessThanSource = true
				continue
			}
			if i >= 1 && equalWord(tokens[i-1], "greater") {
				selection.PowerGreaterThanSource = true
				continue
			}
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

// parseSelectionManaValueDynamic recognizes the "less than or equal to the
// amount of life you (lost|gained) this turn" upper bound that follows "mana
// value" in a graveyard-card target ("creature card with mana value less than or
// equal to the amount of life you lost this turn" — Betor, Ancestor's Voice),
// returning the dynamic life-total kind. It fails closed for any other operator
// or operand so the fixed and X-derived bounds keep their own paths.
func parseSelectionManaValueDynamic(tokens []shared.Token, start int) (EffectDynamicAmountKind, int, bool) {
	if !effectWordsAt(tokens, start, "less", "than", "or", "equal", "to") {
		return "", 0, false
	}
	idx := start + 5
	if !effectWordsAt(tokens, idx, "the") {
		return "", 0, false
	}
	idx++
	switch {
	case effectWordsAt(tokens, idx, "amount", "of", "life"):
		idx += 3
	case effectWordsAt(tokens, idx, "life"):
		idx++
	default:
		return "", 0, false
	}
	switch {
	case effectWordsAt(tokens, idx, "you've"):
		idx++
	case effectWordsAt(tokens, idx, "you", "have"):
		idx += 2
	case effectWordsAt(tokens, idx, "you"):
		idx++
	default:
		return "", 0, false
	}
	var kind EffectDynamicAmountKind
	switch {
	case effectWordsAt(tokens, idx, "lost"):
		kind = EffectDynamicAmountLifeLostThisTurn
	case effectWordsAt(tokens, idx, "gained"):
		kind = EffectDynamicAmountLifeGainedThisTurn
	default:
		return "", 0, false
	}
	idx++
	if !effectWordsAt(tokens, idx, "this", "turn") {
		return "", 0, false
	}
	idx += 2
	return kind, idx, true
}

// selectionManaValueDynamicSpan reports the token span of a "mana value less
// than or equal to the amount of life you (lost|gained) this turn" rider, so the
// unsupported-qualifier gate can treat the dynamic life words ("amount", "life",
// "lost", ...) as a recognized qualifier rather than rejecting the whole
// selection. It returns the half-open [start, end) index range over tokens.
func selectionManaValueDynamicSpan(tokens []shared.Token) (start, end int, ok bool) {
	for i := range tokens {
		if i+2 < len(tokens) && effectWordsAt(tokens, i, "mana", "value") {
			if _, end, ok := parseSelectionManaValueDynamic(tokens, i+2); ok {
				return i, end, true
			}
		}
	}
	return 0, 0, false
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

// staticGroupVerb reports whether token introduces a resolving plural creature
// or permanent group effect clause: "get"/"have" for a power/toughness or
// characteristic change, "gain" for a keyword grant ("Creatures you control gain
// trample until end of turn."), or "lose" for a keyword removal ("Permanents
// your opponents control lose hexproof until end of turn."). The keyword-grant
// and keyword-removal forms lower as one-shot continuous effects over the
// affected group, mirroring the "get" pump form.
func staticGroupVerb(token shared.Token) bool {
	return equalWord(token, "get") || equalWord(token, "have") ||
		equalWord(token, "gain") || equalWord(token, "lose")
}

// staticGroupVerbSingular reports the singular-subject counterparts of
// staticGroupVerb. The distributive "each creature" wording takes the singular
// verb ("Each creature gets/has/gains/loses ...") in place of the plural
// "get/have/gain/lose" used after "all creatures", so it introduces the same
// resolving group effect over every creature.
func staticGroupVerbSingular(token shared.Token) bool {
	return equalWord(token, "gets") || equalWord(token, "has") ||
		equalWord(token, "gains") || equalWord(token, "loses")
}

func parseEffectStaticSubject(tokens []shared.Token, atoms Atoms) EffectStaticSubjectSyntax {
	if subject, ok := parseChosenColorControlledGroupSubject(tokens, atoms); ok {
		return subject
	}
	if subject, ok := parseColoredControlledCreatureGroup(tokens); ok {
		return subject
	}
	if subject, ok := parseColoredBattlefieldCreatureGroup(tokens); ok {
		return subject
	}
	if subject, ok := parseKeywordFilteredCreatureGroupSubject(tokens); ok {
		return subject
	}
	if subject, ok := parseCounterFilteredCreatureGroupSubject(tokens); ok {
		return subject
	}
	if subject, ok := parseControlledCreatureSubtypeTokenGroupSubject(tokens, atoms); ok {
		return subject
	}
	if subject, ok := parseTypeFilteredControlledCreatureGroupSubject(tokens); ok {
		return subject
	}
	if subject, ok := parseChosenTypeControlledCreatureGroupSubject(tokens); ok {
		return subject
	}
	if subject, ok := parseFilteredControlledCreatureGroupSubject(tokens); ok {
		return subject
	}
	if subject, ok := parseRelativeClauseControlledSubtypeSubject(tokens, atoms); ok {
		return subject
	}
	if subject, ok := parseBattlefieldCreatureGroupSubject(tokens, atoms); ok {
		return subject
	}
	if subject, ok := parsePluralSubtypeGroupSubject(tokens, atoms); ok {
		return subject
	}
	switch {
	case len(tokens) >= 3 &&
		(equalWord(tokens[0], "enchanted") || equalWord(tokens[0], "equipped")) &&
		equalWord(tokens[1], "creature") &&
		(equalWord(tokens[2], "gets") || equalWord(tokens[2], "has")):
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectAttachedObject, Span: shared.SpanOf(tokens[:2])}
	case len(tokens) >= 4 && effectWordsAt(tokens, 0, "all", "other", "creatures") &&
		staticGroupVerb(tokens[3]):
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectAllOtherCreatures, Span: shared.SpanOf(tokens[:3])}
	case len(tokens) >= 3 && effectWordsAt(tokens, 0, "all", "creatures") &&
		staticGroupVerb(tokens[2]):
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectAllCreatures, Span: shared.SpanOf(tokens[:2])}
	case len(tokens) >= 3 && effectWordsAt(tokens, 0, "each", "creature") &&
		staticGroupVerbSingular(tokens[2]):
		// "Each creature gets ..." names every creature on the battlefield just as
		// "All creatures get ..." does, but with the singular "each creature" noun
		// and verb. It maps to the same all-creatures group.
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectAllCreatures, Span: shared.SpanOf(tokens[:2])}
	case len(tokens) >= 3 && effectWordsAt(tokens, 0, "attacking", "creatures") &&
		staticGroupVerb(tokens[2]):
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectAttackingCreatures, Span: shared.SpanOf(tokens[:2])}
	case len(tokens) >= 4 && effectWordsAt(tokens, 0, "each", "attacking", "creature") &&
		staticGroupVerbSingular(tokens[3]):
		// "Each attacking creature gets ..." is the singular distributive wording
		// for the same group as "Attacking creatures get ...".
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectAttackingCreatures, Span: shared.SpanOf(tokens[:3])}
	case len(tokens) >= 4 && effectWordsAt(tokens, 0, "other", "attacking", "creatures") &&
		staticGroupVerb(tokens[3]):
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectOtherAttackingCreatures, Span: shared.SpanOf(tokens[:3])}
	case len(tokens) >= 5 && effectWordsAt(tokens, 0, "each", "other", "attacking", "creature") &&
		staticGroupVerbSingular(tokens[4]):
		// "Each other attacking creature gets ..." is the Battle cry group: every
		// attacking creature except the source.
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectOtherAttackingCreatures, Span: shared.SpanOf(tokens[:4])}
	case len(tokens) >= 3 && effectWordsAt(tokens, 0, "blocking", "creatures") &&
		staticGroupVerb(tokens[2]):
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectBlockingCreatures, Span: shared.SpanOf(tokens[:2])}
	case len(tokens) >= 5 && effectWordsAt(tokens, 0, "other", "permanents", "you", "control") &&
		staticGroupVerb(tokens[4]):
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectOtherControlledPermanents, Span: shared.SpanOf(tokens[:4])}
	case len(tokens) >= 4 && effectWordsAt(tokens, 0, "permanents", "you", "control") &&
		staticGroupVerb(tokens[3]):
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectControlledPermanents, Span: shared.SpanOf(tokens[:3])}
	case len(tokens) >= 4 && effectWordsAt(tokens, 0, "commanders", "you", "control") &&
		staticGroupVerb(tokens[3]):
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectControlledCommanders, Span: shared.SpanOf(tokens[:3])}
	case len(tokens) >= 5 && effectWordsAt(tokens, 0, "other", "creatures", "you", "control") &&
		staticGroupVerb(tokens[4]):
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectOtherControlledCreatures, Span: shared.SpanOf(tokens[:4])}
	case len(tokens) >= 4 && effectWordsAt(tokens, 0, "creatures", "you", "control") &&
		staticGroupVerb(tokens[3]):
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectControlledCreatures, Span: shared.SpanOf(tokens[:3])}
	case len(tokens) >= 6 && effectWordsAt(tokens, 0, "each", "other", "creature") &&
		effectWordsAt(tokens, 3, "you", "control") &&
		staticGroupVerbSingular(tokens[5]):
		// "Each other creature you control gets ..." is the singular distributive
		// wording for the same source-excluding group as "Other creatures you
		// control get ...".
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectOtherControlledCreatures, Span: shared.SpanOf(tokens[:5])}
	case len(tokens) >= 5 && effectWordsAt(tokens, 0, "each", "creature") &&
		effectWordsAt(tokens, 2, "you", "control") &&
		staticGroupVerbSingular(tokens[4]):
		// "Each creature you control gets ..." is the singular distributive
		// wording for the same group as "Creatures you control get ...".
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectControlledCreatures, Span: shared.SpanOf(tokens[:4])}
	case len(tokens) >= 5 && effectWordsAt(tokens, 0, "permanents", "your", "opponents", "control") &&
		staticGroupVerb(tokens[4]):
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectOpponentControlledPermanents, Span: shared.SpanOf(tokens[:4])}
	case len(tokens) >= 5 && effectWordsAt(tokens, 0, "creatures", "your", "opponents", "control") &&
		staticGroupVerb(tokens[4]):
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectOpponentControlledCreatures, Span: shared.SpanOf(tokens[:4])}
	case len(tokens) >= 5 && effectWordsAt(tokens, 0, "each", "wall", "you", "control") &&
		(equalWord(tokens[4], "gets") || equalWord(tokens[4], "has")):
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectControlledWalls, Span: shared.SpanOf(tokens[:4]), Subtype: types.Wall, SubtypeKnown: true}
	case len(tokens) >= 4 && effectWordsAt(tokens, 0, "walls", "you", "control") &&
		staticGroupVerb(tokens[3]):
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectControlledWalls, Span: shared.SpanOf(tokens[:3]), Subtype: types.Wall, SubtypeKnown: true}
	case len(tokens) >= 4 && effectWordsAt(tokens, 0, "artifacts", "you", "control") &&
		staticGroupVerb(tokens[3]):
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectControlledArtifacts, Span: shared.SpanOf(tokens[:3])}
	case len(tokens) >= 4 && effectWordsAt(tokens, 0, "sagas", "you", "control") &&
		staticGroupVerb(tokens[3]):
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectControlledSagas, Span: shared.SpanOf(tokens[:3]), Subtype: types.Saga, SubtypeText: string(types.Saga), SubtypeKnown: true}
	case len(tokens) >= 4 && effectWordsAt(tokens, 0, "tokens", "you", "control") &&
		staticGroupVerb(tokens[3]):
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectControlledTokens, Span: shared.SpanOf(tokens[:3])}
	default:
		if subject := parseControlledPermanentSubtypeSubject(tokens, atoms); subject.Kind != EffectStaticSubjectNone {
			return subject
		}
		return parseControlledCreatureSubtypeSubject(tokens, atoms)
	}
}

// parseControlledPermanentSubtypeSubject recognizes a controlled-permanent group
// named directly by a single non-creature permanent subtype noun ("Foods you
// control have ...", "Other Clues you control have ..."). The subtype must
// resolve to an artifact, enchantment, land, planeswalker, or battle subtype and
// must not be a creature subtype, so the controlled-creature subtype productions
// keep owning every creature-typed group noun. The leading "Other" maps to the
// source-excluding subject kind. It returns the zero subject for any other shape.
func parseControlledPermanentSubtypeSubject(tokens []shared.Token, atoms Atoms) EffectStaticSubjectSyntax {
	permanentSubtype := func(index int) (types.Sub, bool) {
		if index >= len(tokens) {
			return "", false
		}
		value, ok := atoms.SubtypeAt(tokens[index].Span)
		if !ok {
			return "", false
		}
		if SubtypeMatchesAnyRuntimeCardType(value, []types.Card{types.Creature, types.Kindred}) {
			return "", false
		}
		if !SubtypeMatchesAnyRuntimeCardType(value, []types.Card{types.Artifact, types.Enchantment, types.Land, types.Planeswalker, types.Battle}) {
			return "", false
		}
		return value, true
	}
	switch {
	case len(tokens) >= 5 && equalWord(tokens[0], "other") &&
		effectWordsAt(tokens, 2, "you", "control") &&
		(equalWord(tokens[4], "have") || equalWord(tokens[4], "get")):
		value, ok := permanentSubtype(1)
		if !ok {
			return EffectStaticSubjectSyntax{}
		}
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectOtherControlledPermanentSubtype, Span: shared.SpanOf(tokens[:4]), Subtype: value, SubtypeText: tokens[1].Text, SubtypeKnown: true}
	case len(tokens) >= 4 && effectWordsAt(tokens, 1, "you", "control") &&
		(equalWord(tokens[3], "have") || equalWord(tokens[3], "get")):
		value, ok := permanentSubtype(0)
		if !ok {
			return EffectStaticSubjectSyntax{}
		}
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectControlledPermanentSubtype, Span: shared.SpanOf(tokens[:3]), Subtype: value, SubtypeText: tokens[0].Text, SubtypeKnown: true}
	default:
		return EffectStaticSubjectSyntax{}
	}
}

// parseControlledCreatureSubtypeSubject recognizes the controlled-creature group
// subjects filtered by a creature subtype: "[Other] <Subtype> creatures you
// control get/have ..." and the "non-<Subtype>" exclusion form ("Non-Human
// creatures you control get ..."). It is split out of parseEffectStaticSubject
// to keep that grammar dispatcher's maintainability within bounds.
func parseControlledCreatureSubtypeSubject(tokens []shared.Token, atoms Atoms) EffectStaticSubjectSyntax {
	subtype := func(index int) (types.Sub, bool) {
		if index >= len(tokens) {
			return "", false
		}
		value, ok := atoms.SubtypeAt(tokens[index].Span)
		return value, ok && SubtypeMatchesAnyRuntimeCardType(value, []types.Card{types.Creature, types.Kindred})
	}
	subtypeKnown := func(index int) bool {
		_, ok := subtype(index)
		return ok
	}
	excludedSubtype := func(index int) (types.Sub, bool) {
		if index >= len(tokens) {
			return "", false
		}
		value, ok := atoms.ExcludedSubtypeAt(tokens[index].Span)
		return value, ok && SubtypeMatchesAnyRuntimeCardType(value, []types.Card{types.Creature, types.Kindred})
	}
	excludedSubtypeKnown := func(index int) bool {
		_, ok := excludedSubtype(index)
		return ok
	}
	switch {
	case len(tokens) >= 6 && equalWord(tokens[0], "other") && equalWord(tokens[2], "creatures") &&
		effectWordsAt(tokens, 3, "you", "control") &&
		staticGroupVerb(tokens[5]) &&
		subtypeKnown(1):
		value, _ := subtype(1)
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectOtherControlledCreatureSubtype, Span: shared.SpanOf(tokens[:5]), Subtype: value, SubtypeText: tokens[1].Text, SubtypeKnown: true}
	case len(tokens) >= 5 && equalWord(tokens[1], "creatures") &&
		effectWordsAt(tokens, 2, "you", "control") &&
		staticGroupVerb(tokens[4]) &&
		subtypeKnown(0):
		value, _ := subtype(0)
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectControlledCreatureSubtype, Span: shared.SpanOf(tokens[:4]), Subtype: value, SubtypeText: tokens[0].Text, SubtypeKnown: true}
	case len(tokens) >= 5 && equalWord(tokens[1], "creatures") &&
		effectWordsAt(tokens, 2, "you", "control") &&
		staticGroupVerb(tokens[4]) &&
		excludedSubtypeKnown(0):
		value, _ := excludedSubtype(0)
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectControlledCreatureSubtype, Span: shared.SpanOf(tokens[:4]), Subtype: value, SubtypeText: tokens[0].Text, SubtypeKnown: true, ExcludedSubtype: true}
	case len(tokens) >= 5 && equalWord(tokens[0], "other") && effectWordsAt(tokens, 2, "you", "control") &&
		staticGroupVerb(tokens[4]):
		value, ok := subtype(1)
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectOtherControlledCreatureSubtype, Span: shared.SpanOf(tokens[:4]), Subtype: value, SubtypeText: tokens[1].Text, SubtypeKnown: ok}
	case len(tokens) >= 4 && effectWordsAt(tokens, 1, "you", "control") &&
		staticGroupVerb(tokens[3]):
		value, ok := subtype(0)
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectControlledCreatureSubtype, Span: shared.SpanOf(tokens[:3]), Subtype: value, SubtypeText: tokens[0].Text, SubtypeKnown: ok}
	default:
		return EffectStaticSubjectSyntax{}
	}
}

// parseControlledCreatureSubtypeTokenGroupSubject recognizes the controlled
// creature-token group subjects filtered by a named creature subtype:
// "[Other] <Subtype> tokens you control get/have ...". It is the token-only
// sibling of the bare "<Subtype> creatures you control" group, matching the
// Amass Zombie cycle's "Zombie tokens you control have <keyword>" anthem (Eternal
// Skylord, Dreadhorde Twins, Vizier of the Scorpion, Gleaming Overseer). The
// named subtype rides SubtypesAny; the leading "Other" maps to the
// source-excluding subject kind. The subtype must resolve to a creature or
// kindred subtype so the noncreature permanent-token groups keep falling through
// to their own productions. It returns false for any other shape.
func parseControlledCreatureSubtypeTokenGroupSubject(tokens []shared.Token, atoms Atoms) (EffectStaticSubjectSyntax, bool) {
	creatureSubtype := func(index int) (types.Sub, bool) {
		if index >= len(tokens) {
			return "", false
		}
		value, ok := atoms.SubtypeAt(tokens[index].Span)
		return value, ok && SubtypeMatchesAnyRuntimeCardType(value, []types.Card{types.Creature, types.Kindred})
	}
	switch {
	case len(tokens) >= 6 && equalWord(tokens[0], "other") && equalWord(tokens[2], "tokens") &&
		effectWordsAt(tokens, 3, "you", "control") &&
		(equalWord(tokens[5], "have") || equalWord(tokens[5], "get")):
		value, ok := creatureSubtype(1)
		if !ok {
			return EffectStaticSubjectSyntax{}, false
		}
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectOtherControlledCreatureSubtypeTokens, Span: shared.SpanOf(tokens[:5]), Subtype: value, SubtypeText: tokens[1].Text, SubtypeKnown: true, SubtypesAny: []types.Sub{value}}, true
	case len(tokens) >= 5 && equalWord(tokens[1], "tokens") &&
		effectWordsAt(tokens, 2, "you", "control") &&
		(equalWord(tokens[4], "have") || equalWord(tokens[4], "get")):
		value, ok := creatureSubtype(0)
		if !ok {
			return EffectStaticSubjectSyntax{}, false
		}
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectControlledCreatureSubtypeTokens, Span: shared.SpanOf(tokens[:4]), Subtype: value, SubtypeText: tokens[0].Text, SubtypeKnown: true, SubtypesAny: []types.Sub{value}}, true
	default:
		return EffectStaticSubjectSyntax{}, false
	}
}

// parseRelativeClauseControlledSubtypeSubject recognizes a controlled-creature
// group narrowed by a relative-clause disjunction of creature subtypes:
// "[Each] [other] creature[s] you control that's a <Subtype> [or [a]
// <Subtype>]...". Tribal anthems that name more than one creature type ("Each
// other creature you control that's a Wolf or a Werewolf gets +1/+1.") use this
// form. Every named subtype rides SubtypesAny so the affected group matches a
// permanent that has any one of them, exactly like the single-subtype "Other
// <Sub> creatures you control" group. A leading "other" maps to the
// source-excluding subject kind. Only the multi-subtype disjunction is accepted;
// the single-subtype relative clause keeps falling through to the existing
// subject productions.
func parseRelativeClauseControlledSubtypeSubject(tokens []shared.Token, atoms Atoms) (EffectStaticSubjectSyntax, bool) {
	idx := 0
	excluded := false
	if idx < len(tokens) && equalWord(tokens[idx], "each") {
		idx++
	}
	if idx < len(tokens) && equalWord(tokens[idx], "other") {
		excluded = true
		idx++
	}
	if idx >= len(tokens) || (!equalWord(tokens[idx], "creature") && !equalWord(tokens[idx], "creatures")) {
		return EffectStaticSubjectSyntax{}, false
	}
	idx++
	if !effectWordsAt(tokens, idx, "you", "control") {
		return EffectStaticSubjectSyntax{}, false
	}
	idx += 2
	switch {
	case idx < len(tokens) && equalWord(tokens[idx], "that's"):
		idx++
	case effectWordsAt(tokens, idx, "that", "is"), effectWordsAt(tokens, idx, "that", "are"):
		idx += 2
	default:
		return EffectStaticSubjectSyntax{}, false
	}
	subs, end, ok := parseControlledCreatureSubtypeOrList(tokens, idx, atoms)
	if !ok {
		return EffectStaticSubjectSyntax{}, false
	}
	kind := EffectStaticSubjectControlledCreatureSubtype
	if excluded {
		kind = EffectStaticSubjectOtherControlledCreatureSubtype
	}
	return EffectStaticSubjectSyntax{
		Kind:         kind,
		Span:         shared.SpanOf(tokens[:end]),
		Subtype:      subs[0],
		SubtypeText:  string(subs[0]),
		SubtypeKnown: true,
		SubtypesAny:  subs,
	}, true
}

// parseControlledCreatureSubtypeOrList parses a disjunctive list of creature
// subtypes beginning at start ("a Wolf or a Werewolf"). Each alternative is a
// single subtype atom optionally preceded by "a"/"an" and separated by "or". It
// returns the resolved subtypes and the token index just past the list, failing
// closed unless every alternative resolves to a creature or kindred subtype and
// at least two are named.
func parseControlledCreatureSubtypeOrList(tokens []shared.Token, start int, atoms Atoms) ([]types.Sub, int, bool) {
	var subs []types.Sub
	idx := start
	for {
		if idx < len(tokens) && (equalWord(tokens[idx], "a") || equalWord(tokens[idx], "an")) {
			idx++
		}
		if idx >= len(tokens) {
			return nil, 0, false
		}
		value, ok := atoms.SubtypeAt(tokens[idx].Span)
		if !ok || !SubtypeMatchesAnyRuntimeCardType(value, []types.Card{types.Creature, types.Kindred}) {
			return nil, 0, false
		}
		subs = append(subs, value)
		idx++
		if idx < len(tokens) && equalWord(tokens[idx], "or") {
			idx++
			continue
		}
		break
	}
	if len(subs) < 2 {
		return nil, 0, false
	}
	return subs, idx, true
}

// doublePTObject is the parsed object of a power/toughness doubling effect: the
// affected group together with which characteristics double.
type doublePTObject struct {
	Subject         EffectStaticSubjectSyntax
	DoublePower     bool
	DoubleToughness bool
}

// parseDoublePTObject recognizes the object of a power/toughness doubling
// effect: "the power and toughness of <group>", "the power of <group>", or "the
// toughness of <group>" (with a trailing duration and period the caller already
// scopes elsewhere). It returns the affected group as a static subject together
// with which characteristics double. Only the controlled-creatures and
// all-creatures groups are recognized; every other object (a player's life, a
// counter count, a single target) returns ok=false so the doubling effect fails
// closed.
func parseDoublePTObject(tokens []shared.Token, atoms Atoms) (doublePTObject, bool) {
	if len(tokens) < 2 || !equalWord(tokens[0], "the") {
		return doublePTObject{}, false
	}
	var object doublePTObject
	index := 1
	switch {
	case equalWord(tokens[index], "power"):
		object.DoublePower = true
		index++
		if index+1 < len(tokens) && equalWord(tokens[index], "and") && equalWord(tokens[index+1], "toughness") {
			object.DoubleToughness = true
			index += 2
		}
	case equalWord(tokens[index], "toughness"):
		object.DoubleToughness = true
		index++
	default:
		return doublePTObject{}, false
	}
	if index >= len(tokens) || !equalWord(tokens[index], "of") {
		return doublePTObject{}, false
	}
	group, groupOK := doubleGroupStaticSubject(tokens[index+1:], atoms)
	if !groupOK {
		return doublePTObject{}, false
	}
	object.Subject = group
	return object, true
}

// doubleCountersObject describes the object of a counter-doubling effect: the
// single counter Kind to double (zero/unused when AllKinds doubles every kind),
// whether every kind of counter is doubled (AllKinds), and whether the doubled
// permanent is a "target ..." object (Target) rather than the source itself.
type doubleCountersObject struct {
	Kind     counter.Kind
	AllKinds bool
	Target   bool
}

// parseDoubleCountersObject recognizes the object of a counter-doubling effect:
// "the number of <kind> counters on <object>" ("double the number of +1/+1
// counters on this creature", Mossborn Hydra; "... on target creature", Gilder
// Bairn) and "the number of each kind of counter on <object>" ("double the
// number of each kind of counter on target artifact, creature, or land", Vorel
// of the Hull Clade). The object is the source itself ("this <permanent>" / "it"
// / the card's own name) or a "target ..." permanent whose target the sentence's
// target scanner owns; any other object returns ok=false so the effect fails
// closed.
func parseDoubleCountersObject(tokens []shared.Token, atoms Atoms) (doubleCountersObject, bool) {
	rest, ok := cutTokenPrefix(tokens, "the", "number", "of")
	if !ok {
		return doubleCountersObject{}, false
	}
	if afterOn, okAll := cutTokenPrefix(rest, "each", "kind", "of", "counter", "on"); okAll {
		target, okScope := doubleCountersObjectScope(afterOn, atoms)
		if !okScope {
			return doubleCountersObject{}, false
		}
		return doubleCountersObject{AllKinds: true, Target: target}, true
	}
	for _, atom := range atoms.Counters() {
		if len(rest) == 0 || atom.Span.Start.Offset != rest[0].Span.Start.Offset {
			continue
		}
		counterNoun := 0
		for counterNoun < len(rest) && rest[counterNoun].Span.End.Offset <= atom.Span.End.Offset {
			counterNoun++
		}
		if counterNoun >= len(rest) ||
			(!equalWord(rest[counterNoun], "counter") && !equalWord(rest[counterNoun], "counters")) ||
			counterNoun+1 >= len(rest) || !equalWord(rest[counterNoun+1], "on") {
			continue
		}
		target, okScope := doubleCountersObjectScope(rest[counterNoun+2:], atoms)
		if !okScope {
			continue
		}
		return doubleCountersObject{Kind: atom.Kind, Target: target}, true
	}
	return doubleCountersObject{}, false
}

// doubleCountersObjectScope reports whether the tokens after "on" name a
// permanent the counter-doubling effect can resolve against and whether that
// object is a target. The source itself ("this <permanent>" / "it" / the card's
// own name, consuming the whole object) returns target=false; a "target ..."
// object whose target the sentence's target scanner owns returns target=true.
func doubleCountersObjectScope(object []shared.Token, atoms Atoms) (target, ok bool) {
	if len(object) == 0 {
		return false, false
	}
	if equalWord(object[0], "target") {
		return true, true
	}
	_, end, okSource := sourceCounterReferenceSpan(object, 0, atoms)
	if !okSource {
		return false, false
	}
	trailingOK := end == len(object) ||
		(end == len(object)-1 && object[end].Kind == shared.Period)
	if !trailingOK {
		return false, false
	}
	return false, true
}

// doubling object's "of <group>" tail: "each creature you control" / "creatures
// you control" (the controlled-creatures group) and "each creature" / "all
// creatures" (every creature on the battlefield). Unlike parseEffectStaticSubject
// these forms are not anchored to a trailing group verb and accept the singular
// "each creature" wording, so they are recognized here rather than reused.
func doubleGroupStaticSubject(tokens []shared.Token, atoms Atoms) (EffectStaticSubjectSyntax, bool) {
	_ = atoms
	start := 0
	hasEach := false
	if len(tokens) > 0 && (equalWord(tokens[0], "each") || equalWord(tokens[0], "all")) {
		hasEach = true
		start = 1
	}
	rest := tokens[start:]
	switch {
	case effectWordsAt(rest, 0, "creature", "you", "control"):
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectControlledCreatures, Span: shared.SpanOf(tokens[:start+3])}, true
	case effectWordsAt(rest, 0, "creatures", "you", "control"):
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectControlledCreatures, Span: shared.SpanOf(tokens[:start+3])}, true
	case hasEach && effectWordsAt(rest, 0, "creature"):
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectAllCreatures, Span: shared.SpanOf(tokens[:start+1])}, true
	case hasEach && effectWordsAt(rest, 0, "creatures"):
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectAllCreatures, Span: shared.SpanOf(tokens[:start+1])}, true
	default:
		return EffectStaticSubjectSyntax{}, false
	}
}

func parseBattlefieldCreatureGroupSubject(tokens []shared.Token, atoms Atoms) (EffectStaticSubjectSyntax, bool) {
	subtypeAt := func(index int) (types.Sub, bool) {
		if index >= len(tokens) {
			return "", false
		}
		value, ok := atoms.SubtypeAt(tokens[index].Span)
		return value, ok && SubtypeMatchesAnyRuntimeCardType(value, []types.Card{types.Creature, types.Kindred})
	}
	excludedSubtypeAt := func(index int) (types.Sub, bool) {
		if index >= len(tokens) {
			return "", false
		}
		value, ok := atoms.ExcludedSubtypeAt(tokens[index].Span)
		return value, ok && SubtypeMatchesAnyRuntimeCardType(value, []types.Card{types.Creature, types.Kindred})
	}
	switch {
	case len(tokens) >= 5 && effectWordsAt(tokens, 0, "attacking", "creatures") &&
		effectWordsAt(tokens, 2, "you", "control") &&
		(equalWord(tokens[4], "get") || equalWord(tokens[4], "have")):
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectControlledAttackingCreatures, Span: shared.SpanOf(tokens[:4])}, true
	case len(tokens) >= 5 && equalWord(tokens[0], "all") && equalWord(tokens[2], "creatures") &&
		(equalWord(tokens[3], "get") || equalWord(tokens[3], "have")):
		if value, ok := subtypeAt(1); ok {
			return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectAllCreatureSubtype, Span: shared.SpanOf(tokens[:3]), Subtype: value, SubtypeText: tokens[1].Text, SubtypeKnown: true}, true
		}
		if value, ok := excludedSubtypeAt(1); ok {
			return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectAllCreatureSubtype, Span: shared.SpanOf(tokens[:3]), Subtype: value, SubtypeText: tokens[1].Text, SubtypeKnown: true, ExcludedSubtype: true}, true
		}
	case len(tokens) >= 5 && equalWord(tokens[0], "other") && equalWord(tokens[2], "creatures") &&
		(equalWord(tokens[3], "get") || equalWord(tokens[3], "have")):
		if value, ok := subtypeAt(1); ok {
			return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectOtherCreatureSubtype, Span: shared.SpanOf(tokens[:3]), Subtype: value, SubtypeText: tokens[1].Text, SubtypeKnown: true}, true
		}
	// Bare "<Subtype> creatures get/have ..." names every creature of that
	// subtype on the battlefield ("Goblin creatures get +1/+1 until end of
	// turn."), the unprefixed sibling of "All <Subtype> creatures get ...".
	case len(tokens) >= 3 && equalWord(tokens[1], "creatures") &&
		(equalWord(tokens[2], "get") || equalWord(tokens[2], "have")):
		if value, ok := subtypeAt(0); ok {
			return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectAllCreatureSubtype, Span: shared.SpanOf(tokens[:2]), Subtype: value, SubtypeText: tokens[0].Text, SubtypeKnown: true}, true
		}
		// "Non-<Subtype> creatures get/have ..." names every creature on the
		// battlefield that does not carry the named subtype ("Non-Elf creatures
		// get -2/-2 until end of turn.").
		if value, ok := excludedSubtypeAt(0); ok {
			return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectAllCreatureSubtype, Span: shared.SpanOf(tokens[:2]), Subtype: value, SubtypeText: tokens[0].Text, SubtypeKnown: true, ExcludedSubtype: true}, true
		}
	case len(tokens) >= 4 && effectWordsAt(tokens, 0, "creature", "tokens") &&
		(equalWord(tokens[2], "get") || equalWord(tokens[2], "have")):
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectBattlefieldCreatureTokens, Span: shared.SpanOf(tokens[:2])}, true
	default:
	}
	return EffectStaticSubjectSyntax{}, false
}

// parseFilteredControlledCreatureGroupSubject recognizes controller-permanent
// creature group subjects that carry a single bounded non-color filter the
// continuous matcher can express: "Creature tokens you control get/have ..."
// (token-only), "Legendary creatures you control get/have ..." (the Legendary
// supertype), "Nonlegendary creatures you control get/have ..." (the excluded
// Legendary supertype), "Untapped creatures you control get/have ..." (untapped
// state), "Modified creatures you control get/have ..." (modified: a counter,
// Aura, or Equipment), and "Other tapped creatures you control get/have ..."
// (tapped state excluding the source). It returns the typed subject, or false so
// callers fall through to the bare grammar. It fails closed for "Tapped"
// battlefield-wide forms that have no Selection representation.
// parseChosenTypeControlledCreatureGroupSubject recognizes the chosen-type
// anthem group subjects "[Other] creatures you control of the chosen type
// get/have/gain ...", the affected group of cards that buff only the controlled
// creatures whose type matches the source permanent's entry-time creature-type
// choice (Patchwork Banner, Adaptive Automaton, Obelisk of Urd). It also
// recognizes the battlefield-wide forms "[All] creatures of the chosen type
// get/have ..." (Shared Triumph, Engineered Plague) and the opponent-only form
// "Creatures of the chosen type your opponents control get/have ..." (Plague
// Engineer). It returns false so callers fall through to the bare
// controlled-creature grammar.
func parseChosenTypeControlledCreatureGroupSubject(tokens []shared.Token) (EffectStaticSubjectSyntax, bool) {
	switch {
	case len(tokens) >= 9 && effectWordsAt(tokens, 0, "other", "creatures", "you", "control", "of", "the", "chosen", "type") &&
		staticGroupVerb(tokens[8]):
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectOtherControlledCreaturesChosenType, Span: shared.SpanOf(tokens[:8])}, true
	case len(tokens) >= 8 && effectWordsAt(tokens, 0, "creatures", "you", "control", "of", "the", "chosen", "type") &&
		staticGroupVerb(tokens[7]):
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectControlledCreaturesChosenType, Span: shared.SpanOf(tokens[:7])}, true
	case len(tokens) >= 9 && effectWordsAt(tokens, 0, "creatures", "of", "the", "chosen", "type", "your", "opponents", "control") &&
		staticGroupVerb(tokens[8]):
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectOpponentControlledCreaturesChosenType, Span: shared.SpanOf(tokens[:8])}, true
	case len(tokens) >= 7 && effectWordsAt(tokens, 0, "all", "creatures", "of", "the", "chosen", "type") &&
		staticGroupVerb(tokens[6]):
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectAllCreaturesChosenType, Span: shared.SpanOf(tokens[:6])}, true
	case len(tokens) >= 6 && effectWordsAt(tokens, 0, "creatures", "of", "the", "chosen", "type") &&
		staticGroupVerb(tokens[5]):
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectAllCreaturesChosenType, Span: shared.SpanOf(tokens[:5])}, true
	default:
	}
	return EffectStaticSubjectSyntax{}, false
}

// parseChosenColorControlledGroupSubject recognizes a static anthem group carrying
// a trailing "of the chosen color" qualifier ("Creatures you control of the
// chosen color get ...", Heraldic Banner). It strips the four qualifier tokens,
// re-parses the bare group head, and records ChosenColorFromEntry on the result
// so the affected group is constrained to permanents sharing the source
// permanent's entry-time color choice. It composes over any base group whose
// noun phrase the qualifier immediately follows, failing closed otherwise.
func parseChosenColorControlledGroupSubject(tokens []shared.Token, atoms Atoms) (EffectStaticSubjectSyntax, bool) {
	const qualifierWidth = 4
	for index := 1; index+qualifierWidth < len(tokens); index++ {
		if !effectWordsAt(tokens, index, "of", "the", "chosen", "color") {
			continue
		}
		if !staticGroupVerb(tokens[index+qualifierWidth]) {
			return EffectStaticSubjectSyntax{}, false
		}
		base := make([]shared.Token, 0, len(tokens)-qualifierWidth)
		base = append(base, tokens[:index]...)
		base = append(base, tokens[index+qualifierWidth:]...)
		group := parseEffectStaticSubject(base, atoms)
		if group.Kind == EffectStaticSubjectNone || group.ChosenColorFromEntry {
			return EffectStaticSubjectSyntax{}, false
		}
		if tokensCoveredCount(base, group.Span) != index {
			return EffectStaticSubjectSyntax{}, false
		}
		group.Span = shared.SpanOf(tokens[:index+qualifierWidth])
		group.ChosenColorFromEntry = true
		return group, true
	}
	return EffectStaticSubjectSyntax{}, false
}

func parseFilteredControlledCreatureGroupSubject(tokens []shared.Token) (EffectStaticSubjectSyntax, bool) {
	switch {
	case len(tokens) >= 5 && effectWordsAt(tokens, 0, "creature", "tokens", "you", "control") &&
		(equalWord(tokens[4], "get") || equalWord(tokens[4], "have")):
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectControlledCreatureTokens, Span: shared.SpanOf(tokens[:4])}, true
	case len(tokens) >= 5 && effectWordsAt(tokens, 0, "legendary", "creatures", "you", "control") &&
		(equalWord(tokens[4], "get") || equalWord(tokens[4], "have")):
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectControlledLegendaryCreatures, Span: shared.SpanOf(tokens[:4])}, true
	case len(tokens) >= 5 && effectWordsAt(tokens, 0, "nonlegendary", "creatures", "you", "control") &&
		(equalWord(tokens[4], "get") || equalWord(tokens[4], "have")):
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectControlledNonlegendaryCreatures, Span: shared.SpanOf(tokens[:4])}, true
	case len(tokens) >= 5 && effectWordsAt(tokens, 0, "commander", "creatures", "you", "control") &&
		(equalWord(tokens[4], "get") || equalWord(tokens[4], "have")):
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectControlledCommanderCreatures, Span: shared.SpanOf(tokens[:4])}, true
	case len(tokens) >= 5 && effectWordsAt(tokens, 0, "commander", "creatures", "you", "own") &&
		(equalWord(tokens[4], "get") || equalWord(tokens[4], "have")):
		// "Commander creatures you own ..." is the Background buff group (Folk
		// Hero, Raised by Giants). A Background grants its abilities to a
		// commander its controller owns and controls, so this owned-commander
		// wording maps to the same controlled-commander affected group.
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectControlledCommanderCreatures, Span: shared.SpanOf(tokens[:4])}, true
	case len(tokens) >= 5 && effectWordsAt(tokens, 0, "untapped", "creatures", "you", "control") &&
		(equalWord(tokens[4], "get") || equalWord(tokens[4], "have")):
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectControlledUntappedCreatures, Span: shared.SpanOf(tokens[:4])}, true
	case len(tokens) >= 5 && effectWordsAt(tokens, 0, "modified", "creatures", "you", "control") &&
		(equalWord(tokens[4], "get") || equalWord(tokens[4], "have")):
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectControlledModifiedCreatures, Span: shared.SpanOf(tokens[:4])}, true
	case len(tokens) >= 6 && effectWordsAt(tokens, 0, "other", "tapped", "creatures", "you", "control") &&
		(equalWord(tokens[5], "get") || equalWord(tokens[5], "have")):
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectOtherControlledTappedCreatures, Span: shared.SpanOf(tokens[:5])}, true
	default:
	}
	return EffectStaticSubjectSyntax{}, false
}

// parseTypeFilteredControlledCreatureGroupSubject recognizes controller-permanent
// creature group subjects that carry a single bounded card-type or token filter
// the continuous matcher can express: "[Other] artifact creatures you control
// get/have ..." (the conjunctive artifact-creature type line) and "[Other]
// nontoken creatures you control get/have ..." (the non-token state). It returns
// the typed subject, mirroring the bare controlled creature group forms with the
// extra filter attached, or false so callers fall through to the bare grammar.
func parseTypeFilteredControlledCreatureGroupSubject(tokens []shared.Token) (EffectStaticSubjectSyntax, bool) {
	switch {
	case len(tokens) >= 6 && effectWordsAt(tokens, 0, "other", "artifact", "creatures", "you", "control") &&
		(equalWord(tokens[5], "get") || equalWord(tokens[5], "have")):
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectOtherControlledArtifactCreatures, Span: shared.SpanOf(tokens[:5])}, true
	case len(tokens) >= 5 && effectWordsAt(tokens, 0, "artifact", "creatures", "you", "control") &&
		(equalWord(tokens[4], "get") || equalWord(tokens[4], "have")):
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectControlledArtifactCreatures, Span: shared.SpanOf(tokens[:4])}, true
	case len(tokens) >= 6 && effectWordsAt(tokens, 0, "other", "nontoken", "creatures", "you", "control") &&
		(equalWord(tokens[5], "get") || equalWord(tokens[5], "have")):
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectOtherControlledNontokenCreatures, Span: shared.SpanOf(tokens[:5])}, true
	case len(tokens) >= 5 && effectWordsAt(tokens, 0, "nontoken", "creatures", "you", "control") &&
		(equalWord(tokens[4], "get") || equalWord(tokens[4], "have")):
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectControlledNontokenCreatures, Span: shared.SpanOf(tokens[:4])}, true
	default:
	}
	return EffectStaticSubjectSyntax{}, false
}

// parseKeywordFilteredCreatureGroupSubject recognizes a creature group carrying a
// single "with <keyword>"/"without <keyword>" filter the continuous matcher can
// express. It recognizes the battlefield-wide ("Creatures with flying get ..."),
// controlled ("Creatures you control with flying get ..."), excluding-source
// ("Other creatures you control with flying get ..."), opponent-controlled
// ("Creatures with flying your opponents control get ..."), and negated
// ("Creatures without flying get ...") forms, reusing the matching bare creature
// subject kind with the keyword predicate attached. It returns false (so callers
// fall through to the bare grammar) when no recognizable keyword qualifies the
// group or the group shape is not one of these closed forms.
func parseKeywordFilteredCreatureGroupSubject(tokens []shared.Token) (EffectStaticSubjectSyntax, bool) {
	idx := 0
	excludeSource := false
	if len(tokens) > 0 && equalWord(tokens[0], "other") {
		excludeSource = true
		idx = 1
	}
	if idx >= len(tokens) || !equalWord(tokens[idx], "creatures") {
		return EffectStaticSubjectSyntax{}, false
	}
	idx++
	controlled := false
	if idx+1 < len(tokens) && equalWord(tokens[idx], "you") && equalWord(tokens[idx+1], "control") {
		controlled = true
		idx += 2
	}
	filter, ok := staticGroupKeywordFilterAt(tokens, idx)
	if !ok {
		return EffectStaticSubjectSyntax{}, false
	}
	kind, end, ok := staticKeywordGroupKind(tokens, filter.end, controlled, excludeSource)
	if !ok {
		return EffectStaticSubjectSyntax{}, false
	}
	subject := EffectStaticSubjectSyntax{Kind: kind, Span: shared.SpanOf(tokens[:end])}
	if filter.excluded {
		subject.ExcludedKeyword = filter.keyword
	} else {
		subject.Keyword = filter.keyword
	}
	return subject, true
}

// parseCounterFilteredCreatureGroupSubject recognizes a controller-creature group
// constrained to members carrying a counter: "Each [other] creature you control
// with a <kind> counter on it has/gets ..." (singular) and "[Other] creatures you
// control with a <kind> counter on them have/get ..." (plural). These are the
// counter-matters anthem subjects (Abzan Falconer, Ainok Bond-Kin, Bramblewood
// Paragon). It records the required counter on the subject and fails closed for
// any other shape so callers fall through to the bare grammar.
func parseCounterFilteredCreatureGroupSubject(tokens []shared.Token) (EffectStaticSubjectSyntax, bool) {
	head, ok := counterGroupNounPhrase(tokens)
	if !ok {
		return EffectStaticSubjectSyntax{}, false
	}
	idx := head.next
	if !effectWordsAt(tokens, idx, "you", "control") {
		return EffectStaticSubjectSyntax{}, false
	}
	idx += 2
	match, ok := counterQualifierKind(tokens, idx)
	if !ok || match.Absent {
		// The negated "with no counters" qualifier has no modeled
		// counter-filtered group subject; fail closed so it is not mistaken for a
		// required-counter group.
		return EffectStaticSubjectSyntax{}, false
	}
	if !counterGroupVerbAt(tokens, match.End, head.singular) {
		return EffectStaticSubjectSyntax{}, false
	}
	groupKind := EffectStaticSubjectControlledCreatures
	if head.excludeSource {
		groupKind = EffectStaticSubjectOtherControlledCreatures
	}
	subject := EffectStaticSubjectSyntax{
		Kind:            groupKind,
		Span:            shared.SpanOf(tokens[:match.End]),
		CounterRequired: true,
	}
	if match.Any {
		subject.CounterAny = true
	} else {
		subject.CounterKind = match.Kind
	}
	return subject, true
}

// counterGroupHead is the leading noun phrase of a counter-matters anthem
// subject: the token index just past it, whether the source is excluded
// ("other"), and whether the phrase is the singular "each creature" form.
type counterGroupHead struct {
	next          int
	excludeSource bool
	singular      bool
}

// counterGroupNounPhrase recognizes the leading noun phrase of a counter-matters
// anthem subject. It fails closed for any other noun phrase.
func counterGroupNounPhrase(tokens []shared.Token) (counterGroupHead, bool) {
	switch {
	case effectWordsAt(tokens, 0, "each", "other", "creature"):
		return counterGroupHead{next: 3, excludeSource: true, singular: true}, true
	case effectWordsAt(tokens, 0, "each", "creature"):
		return counterGroupHead{next: 2, singular: true}, true
	case effectWordsAt(tokens, 0, "other", "creatures"):
		return counterGroupHead{next: 2, excludeSource: true}, true
	case effectWordsAt(tokens, 0, "creatures"):
		return counterGroupHead{next: 1}, true
	default:
		return counterGroupHead{}, false
	}
}

// counterGroupVerbAt reports whether the token at index introduces the group
// effect verb that follows a counter-matters anthem subject: the singular
// "has"/"gets" after "each creature", or the plural "have"/"get".
func counterGroupVerbAt(tokens []shared.Token, index int, singular bool) bool {
	if index >= len(tokens) {
		return false
	}
	if singular {
		return equalWord(tokens[index], "has") || equalWord(tokens[index], "gets")
	}
	return equalWord(tokens[index], "have") || equalWord(tokens[index], "get")
}

// staticKeywordGroupKind resolves the static subject kind for a keyword-filtered
// creature group whose keyword qualifier ends at token index end, given whether
// the group is controller-scoped and whether it excludes the source. It also
// recognizes the trailing "your opponents control" that turns an otherwise
// battlefield-wide group into an opponent-controlled one, returning the updated
// end index. It requires the group clause to continue with a "get"/"have" verb
// and fails closed otherwise.
func staticKeywordGroupKind(tokens []shared.Token, end int, controlled, excludeSource bool) (EffectStaticSubjectKind, int, bool) {
	verbAt := func(i int) bool {
		return i < len(tokens) && (equalWord(tokens[i], "get") || equalWord(tokens[i], "have"))
	}
	switch {
	case controlled && excludeSource && verbAt(end):
		return EffectStaticSubjectOtherControlledCreatures, end, true
	case controlled && verbAt(end):
		return EffectStaticSubjectControlledCreatures, end, true
	case excludeSource && verbAt(end):
		return EffectStaticSubjectAllOtherCreatures, end, true
	case !excludeSource && effectWordsAt(tokens, end, "your", "opponents", "control") && verbAt(end+3):
		return EffectStaticSubjectOpponentControlledCreatures, end + 3, true
	case !excludeSource && verbAt(end):
		return EffectStaticSubjectAllCreatures, end, true
	default:
		return EffectStaticSubjectNone, end, false
	}
}

// staticGroupKeywordFilter is a recognized "with <keyword>" / "without <keyword>"
// qualifier on an affected creature group: the keyword kind, whether it is an
// exclusion ("without"), and the token index just past the keyword name.
type staticGroupKeywordFilter struct {
	keyword  KeywordKind
	excluded bool
	end      int
}

// staticGroupKeywordFilterAt recognizes a "with <keyword>" or "without <keyword>"
// group filter beginning at token index i. It fails closed for any other word or
// an unrecognized keyword name.
func staticGroupKeywordFilterAt(tokens []shared.Token, i int) (staticGroupKeywordFilter, bool) {
	if i >= len(tokens) {
		return staticGroupKeywordFilter{end: i}, false
	}
	excluded := false
	switch {
	case equalWord(tokens[i], "with"):
	case equalWord(tokens[i], "without"):
		excluded = true
	default:
		return staticGroupKeywordFilter{end: i}, false
	}
	kind, width, ok := recognizeKeywordNameAt(tokens, i+1)
	if !ok {
		return staticGroupKeywordFilter{end: i}, false
	}
	return staticGroupKeywordFilter{keyword: kind, excluded: excluded, end: i + 1 + width}, true
}

// staticGroupColorFilter is a recognized color constraint on an affected creature
// group, holding the disjunctive single colors and the colorless/multicolored
// color-family qualifiers.
type staticGroupColorFilter struct {
	colors       []Color
	colorless    bool
	multicolored bool
}

// parseColoredControlledCreatureGroup recognizes a controller-permanent creature
// group carrying a color filter: "[Other] <color> creatures you control
// get/have ...". It returns the typed subject, mirroring the bare controlled and
// other-controlled creature group forms with the color predicate attached. It
// fails closed for any non-color qualifier so callers fall through to the bare
// grammar.
func parseColoredControlledCreatureGroup(tokens []shared.Token) (EffectStaticSubjectSyntax, bool) {
	colorIndex, kind, spanEnd := 0, EffectStaticSubjectControlledCreatures, 4
	if len(tokens) >= 1 && equalWord(tokens[0], "other") {
		colorIndex, kind, spanEnd = 1, EffectStaticSubjectOtherControlledCreatures, 5
	}
	filter, width, ok := staticColorFilterAt(tokens, colorIndex)
	if !ok {
		return EffectStaticSubjectSyntax{}, false
	}
	creature := colorIndex + width
	if len(tokens) < creature+4 ||
		!equalWord(tokens[creature], "creatures") ||
		!effectWordsAt(tokens, creature+1, "you", "control") ||
		!staticGroupVerb(tokens[creature+3]) {
		return EffectStaticSubjectSyntax{}, false
	}
	return EffectStaticSubjectSyntax{
		Kind:         kind,
		Span:         shared.SpanOf(tokens[:spanEnd]),
		Colors:       filter.colors,
		Colorless:    filter.colorless,
		Multicolored: filter.multicolored,
	}, true
}

// parseColoredBattlefieldCreatureGroup recognizes a battlefield-wide creature
// group carrying a color filter: "[Other] <color> creatures get/have ...". It
// reuses the all-creature and all-other-creature subject kinds with the color
// predicate attached, so the affected group spans every matching permanent
// regardless of controller. It is tried only after the controlled color form, so
// "you control" variants never reach here. It fails closed for any non-color
// qualifier so callers fall through to the bare grammar.
func parseColoredBattlefieldCreatureGroup(tokens []shared.Token) (EffectStaticSubjectSyntax, bool) {
	colorIndex, kind, spanEnd := 0, EffectStaticSubjectAllCreatures, 2
	if len(tokens) >= 1 && equalWord(tokens[0], "other") {
		colorIndex, kind, spanEnd = 1, EffectStaticSubjectAllOtherCreatures, 3
	}
	filter, width, ok := staticColorFilterAt(tokens, colorIndex)
	if !ok {
		return EffectStaticSubjectSyntax{}, false
	}
	creature := colorIndex + width
	if len(tokens) < creature+2 ||
		!equalWord(tokens[creature], "creatures") ||
		!staticGroupVerb(tokens[creature+1]) {
		return EffectStaticSubjectSyntax{}, false
	}
	return EffectStaticSubjectSyntax{
		Kind:         kind,
		Span:         shared.SpanOf(tokens[:spanEnd]),
		Colors:       filter.colors,
		Colorless:    filter.colorless,
		Multicolored: filter.multicolored,
	}, true
}

// at index, returning the typed color filter and its token width. A bare color
// word ("red") yields a one-element colors slice; "colorless" and "multicolored"
// yield the matching qualifier flag. It fails closed for any other word,
// including "monocolored", which no Selection color filter can represent.
func staticColorFilterAt(tokens []shared.Token, index int) (staticGroupColorFilter, int, bool) {
	if index < 0 || index >= len(tokens) {
		return staticGroupColorFilter{}, 0, false
	}
	if value, ok := recognizeColorWord(tokens[index].Text); ok {
		return staticGroupColorFilter{colors: []Color{value}}, 1, true
	}
	switch qualifier, ok := recognizeColorQualifierWord(tokens[index].Text); {
	case ok && qualifier == ColorQualifierColorless:
		return staticGroupColorFilter{colorless: true}, 1, true
	case ok && qualifier == ColorQualifierMulticolored:
		return staticGroupColorFilter{multicolored: true}, 1, true
	}
	return staticGroupColorFilter{}, 0, false
}

func selectionKindForNoun(noun ObjectNoun) SelectionKind {
	switch noun {
	case ObjectNounArtifact:
		return SelectionArtifact
	case ObjectNounCard:
		return SelectionCard
	case ObjectNounCommander:
		return SelectionCommander
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
