package parser

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func exactEffectSyntax(effect *EffectSyntax) bool {
	switch effect.Kind {
	case EffectAddMana:
		return effect.Mana.ChosenColorDevotion || exactDynamicColorlessManaEffectSyntax(effect)
	case EffectDealDamage:
		return exactDamageEffectSyntax(effect) || exactSourcePowerDamageEffectSyntax(effect)
	case EffectCanAttackAsThoughDefender:
		return exactCanAttackAsThoughDefenderEffectSyntax(effect)
	case EffectCantBeBlocked:
		return exactCantBeBlockedEffectSyntax(effect)
	case EffectCantBlock:
		return exactCantBlockEffectSyntax(effect)
	case EffectCantAttack:
		return exactCantAttackEffectSyntax(effect)
	case EffectCantAttackOrBlock:
		return exactCantAttackOrBlockEffectSyntax(effect)
	case EffectMustAttack:
		return exactTargetMustAttackEffectSyntax(effect)
	case EffectCounter:
		return exactCounterEffectSyntax(effect)
	case EffectCopyStackObject:
		return exactCopyStackObjectEffectSyntax(effect)
	case EffectChooseNewTargets:
		return exactChooseNewTargetsEffectSyntax(effect)
	case EffectChooseCreatureType:
		return strings.EqualFold(exactEffectClauseText(effect), "Choose a creature type.")
	case EffectCreate:
		return exactCreateMultiTokenEffectSyntax(effect) ||
			exactCreateTokenEffectSyntax(effect) ||
			exactCreateTokenEachPlayerEffectSyntax(effect) ||
			exactCreateTokenForEachDestroyedThisWayEffectSyntax(effect) ||
			exactCreateTokenForEachExiledThisWayEffectSyntax(effect) ||
			exactCreateNamedTokenEffectSyntax(effect) ||
			exactCreatePredefinedTokenEffectSyntax(effect) ||
			exactCreateNamedTokenChoiceEffectSyntax(effect) ||
			exactCreateCopyTokenEffectSyntax(effect) ||
			exactCreateCopyTokenReferenceEffectSyntax(effect) ||
			exactCreateCopyTokenTriggeringSetEffectSyntax(effect) ||
			exactCreateCopyTokenAttachedEffectSyntax(effect)
	case EffectCreateEmblem:
		return exactCreateEmblemEffectSyntax(effect)
	case EffectDiscard:
		return exactCardCountEffectSyntax(effect, "Discard", "discards", false) ||
			effect.DiscardEntireHand ||
			effect.HandDiscard.AtRandom ||
			effect.RandomDiscard
	case EffectDestroy:
		return exactDirectTargetEffectSyntax(effect, "Destroy") ||
			exactMultiDistinctTargetEffectSyntax(effect, "Destroy") ||
			exactMassEffectSyntax(effect, "Destroy all ") ||
			exactMassEachEffectSyntax(effect, "Destroy each ") ||
			exactDestroyForEachPlayerEffectSyntax(effect) ||
			exactBackReferenceEffectSyntax(effect, "Destroy")
	case EffectDig:
		return exactDigLookEffectSyntax(effect)
	case EffectDraw:
		return exactCardCountEffectSyntax(effect, "Draw", "draws", true)
	case EffectEnterTapped:
		return exactLegacyFixedAmountSyntax(effect) ||
			effect.GroupEntryModification.Kind != GroupEntryModificationNone
	case EffectExile:
		return exactSourceSpellExileSyntax(effect) ||
			exactCounteredSpellExileSyntax(effect) ||
			exactExileUntilSourceLeavesEffectSyntax(effect) ||
			exactExileForEachPlayerUntilLeavesEffectSyntax(effect) ||
			exactExileTopOfLibrarySyntax(effect) ||
			exactExileEntireHandEffectSyntax(effect) ||
			exactExileAttachedEffectSyntax(effect) ||
			exactDirectTargetEffectSyntax(effect, "Exile") ||
			exactMultiDistinctTargetEffectSyntax(effect, "Exile") ||
			exactMassEffectSyntax(effect, "Exile all ") ||
			exactMassEachEffectSyntax(effect, "Exile each ") ||
			exactBackReferenceEffectSyntax(effect, "Exile") ||
			exactGraveyardExileEffectSyntax(effect) ||
			exactPlayerGraveyardExileEffectSyntax(effect)
	case EffectFight:
		return exactFightEffectSyntax(effect)
	case EffectExplore:
		return exactDirectPronounEffectSyntax(effect, "It explores.") ||
			exactSourceExploresEffectSyntax(effect) ||
			exactTargetExploresEffectSyntax(effect)
	case EffectGain:
		return exactLifeEffectSyntax(effect, "gain", "gains") ||
			exactTemporaryKeywordEffectSyntax(effect) ||
			exactDirectTargetKeywordGrantEffectSyntax(effect) ||
			exactControlledSourceKeywordGrantEffectSyntax(effect) ||
			exactBackReferenceTargetKeywordGrantEffectSyntax(effect) ||
			exactGainGrantedAbilityEffectSyntax(effect)
	case EffectGainControl:
		return exactGainControlEffectSyntax(effect) ||
			exactGiveControlEffectSyntax(effect)
	case EffectBecomeMonarch:
		return exactBecomeMonarchEffectSyntax(effect)
	case EffectInvestigate:
		return exactStandaloneActionEffectSyntax(effect, "Investigate")
	case EffectAmass:
		return exactAmassEffectSyntax(effect)
	case EffectRenown:
		return exactStandaloneActionEffectSyntax(effect, "renown")
	case EffectAdapt:
		return exactStandaloneActionEffectSyntax(effect, "Adapt")
	case EffectConnive:
		return exactConniveEffectSyntax(effect)
	case EffectLose:
		return exactLifeEffectSyntax(effect, "lose", "loses") ||
			exactLifeEffectSyntax(effect, "pay", "pays") ||
			exactTemporaryKeywordLossEffectSyntax(effect)
	case EffectLoseGame:
		return exactLoseGameEffectSyntax(effect)
	case EffectWinGame:
		return strings.EqualFold(exactEffectClauseText(effect), "You win the game.")
	case EffectManifest:
		return strings.EqualFold(exactEffectClauseText(effect), "Manifest the top card of your library.")
	case EffectManifestDread:
		return strings.EqualFold(exactEffectClauseText(effect), "Manifest dread.")
	case EffectMill:
		return exactCardCountEffectSyntax(effect, "Mill", "mills", true)
	case EffectMoveCounters:
		return exactMoveCountersEffectSyntax(effect)
	case EffectModifyPT:
		return exactModifyPTEffectSyntax(effect)
	case EffectGainPlayerCounter:
		return exactGainPlayerCounterEffectSyntax(effect)
	case EffectPut:
		return exactCounterPlacementEffectSyntax(effect) || exactGraveyardPutEffectSyntax(effect) ||
			exactCounteredSpellDestinationSyntax(effect) ||
			exactDigPutEffectSyntax(effect) || exactHandLibraryPutEffectSyntax(effect) ||
			exactPutThoseCountersEffectSyntax(effect) || exactPutThoseCardsIntoHandEffectSyntax(effect) ||
			exactBottomLinkedExiledCardsEffectSyntax(effect) ||
			exactPutLinkedExiledRestOnLibraryBottomEffectSyntax(effect) ||
			exactCounterExiledCardManaValueEffectSyntax(effect) ||
			exactDistributeCountersEffectSyntax(effect)
	case EffectProliferate:
		return exactStandaloneActionEffectSyntax(effect, "Proliferate")
	case EffectRemoveCounter:
		return exactRemoveCounterEffectSyntax(effect) || exactRemoveAllCountersEffectSyntax(effect)
	case EffectRegenerate:
		return exactDirectTargetEffectSyntax(effect, "Regenerate") ||
			exactRegenerateSelfEffectSyntax(effect) ||
			exactRegenerateAttachedEffectSyntax(effect) ||
			exactMassEffectSyntax(effect, "Regenerate all ") ||
			exactMassEachEffectSyntax(effect, "Regenerate each ")
	case EffectReorderLibraryTop:
		return exactLibraryTopReorderEffectSyntax(effect)
	case EffectReturn:
		return exactBounceEffectSyntax(effect) ||
			exactMultiBounceEffectSyntax(effect) ||
			exactDualBounceEffectSyntax(effect) ||
			exactMassBounceEffectSyntax(effect) ||
			exactMassEachBounceEffectSyntax(effect) ||
			exactControlledBounceEffectSyntax(effect) ||
			exactSelfBounceEffectSyntax(effect) ||
			exactGraveyardReturnEffectSyntax(effect) ||
			exactChosenCardsBattlefieldReturnEffectSyntax(effect) ||
			exactReturnExiledCardEffectSyntax(effect) ||
			exactReturnLinkedExiledToBattlefieldPartialEffectSyntax(effect) ||
			exactReturnExiledCardsToHandEffectSyntax(effect) ||
			exactReturnSourceAndExiledCardToHandEffectSyntax(effect) ||
			exactDirectPronounEffectSyntax(effect, "Return it to its owner's hand.")
	default:
		return exactEffectSyntaxTail(effect)
	}
}

// exactEffectSyntaxTail continues the exact-reconstruction dispatch for the
// remaining effect kinds. It is split out of exactEffectSyntax so neither
// function's maintainability index falls below the linter threshold.
func exactEffectSyntaxTail(effect *EffectSyntax) bool {
	switch effect.Kind {
	case EffectDevour:
		return effect.EntersDevour && effect.EntersDevourMultiplier > 0
	case EffectTribute:
		return effect.EntersTribute && effect.EntersTributeCount > 0
	case EffectSacrifice:
		return exactDirectPronounEffectSyntax(effect, "Sacrifice it.") ||
			exactSelfSacrificeEffectSyntax(effect) ||
			exactSacrificeChoiceEffectSyntax(effect) ||
			exactSacrificeMassEffectSyntax(effect)
	case EffectSearch:
		return exactSearchEffectSyntax(effect)
	case EffectScry:
		return exactControllerAmountEffectSyntax(effect, "Scry")
	case EffectSurveil:
		return exactControllerAmountEffectSyntax(effect, "Surveil")
	case EffectShuffle:
		return exactOptionalControllerShuffleEffectSyntax(effect) ||
			exactSourceSpellShuffleIntoLibrarySyntax(effect) ||
			exactControllerGraveyardShuffleIntoLibrarySyntax(effect)
	case EffectTap:
		return exactDirectTargetEffectSyntax(effect, "Tap") ||
			exactDirectReferenceEffectSyntax(effect, "Tap") ||
			exactTapAttachedEffectSyntax(effect) ||
			exactMassEffectSyntax(effect, "Tap all ") ||
			exactMassEachEffectSyntax(effect, "Tap each ")
	case EffectTapOrUntap:
		return exactDirectTargetEffectSyntax(effect, "Tap or untap")
	case EffectUntap:
		return exactDirectTargetEffectSyntax(effect, "Untap") ||
			exactDirectReferenceEffectSyntax(effect, "Untap") ||
			exactUntapAttachedEffectSyntax(effect) ||
			exactMassEffectSyntax(effect, "Untap all ") ||
			exactMassEachEffectSyntax(effect, "Untap each ") ||
			exactBoundedUntapEffectSyntax(effect) ||
			exactNegatedNextUntapStepSyntax(effect) ||
			exactTargetNextUntapStepSyntax(effect) ||
			exactSourceNextUntapStepSyntax(effect) ||
			exactPriorSubjectNextUntapStepSyntax(effect)
	case EffectTransform:
		return exactDirectTargetEffectSyntax(effect, "Transform")
	case EffectRemoveFromCombat:
		return exactRemoveFromCombatEffectSyntax(effect)
	default:
		return false
	}
}

func exactOptionalControllerShuffleEffectSyntax(effect *EffectSyntax) bool {
	if effect.Context != EffectContextController || !effect.Optional {
		return false
	}
	text := exactEffectClauseText(effect)
	return strings.EqualFold(text, "Shuffle.") ||
		strings.EqualFold(text, "Shuffle your library.")
}

// exactControllerGraveyardShuffleIntoLibrarySyntax recognizes the verbatim
// "Shuffle your graveyard into your library." (The Mending of Dominaria
// chapter III), a controller-scoped, non-optional shuffle whose graveyard
// source is carried by FromZone (see effectFromZone).
func exactControllerGraveyardShuffleIntoLibrarySyntax(effect *EffectSyntax) bool {
	if effect.Context != EffectContextController || effect.Optional {
		return false
	}
	if effect.FromZone != zone.Graveyard || effect.ToZone != zone.Library {
		return false
	}
	return strings.EqualFold(exactEffectClauseText(effect), "Shuffle your graveyard into your library.")
}

func exactLibraryTopReorderEffectSyntax(effect *EffectSyntax) bool {
	if effect.Context != EffectContextController || !effect.Amount.Known || effect.Amount.Value < 1 {
		return false
	}
	return strings.EqualFold(
		exactEffectClauseText(effect),
		fmt.Sprintf(
			"Look at the top %s cards of your library, then put them back in any order.",
			effectAmountSourceText(effect),
		),
	)
}

func exactDynamicColorlessManaEffectSyntax(effect *EffectSyntax) bool {
	return effect.Mana.DynamicColorless &&
		effect.Context == EffectContextController &&
		effect.DelayedTiming == DelayedTimingNextMain &&
		effect.Amount.DynamicKind == EffectDynamicAmountSourceManaValue &&
		effect.Amount.DynamicForm == EffectDynamicAmountFormEqual &&
		effect.Amount.Multiplier == 1 &&
		len(effect.References) == 1 &&
		strings.EqualFold(
			strings.TrimSpace(effect.Text),
			"At the beginning of your next main phase, add an amount of {C} equal to that spell's mana value.",
		)
}

// exactBoundedUntapEffectSyntax reconstructs the canonical "Untap up to N
// <permanent group>." clause from the parsed Selection and count and compares it
// byte-for-byte against the source. It recognizes the untargeted "up to N" range
// (Minimum 0, Maximum 2..10) of a permanent group the runtime ChooseUpTo untap
// models: a plain card-type or permanent noun (lands, creatures, artifacts,
// enchantments, planeswalkers, battles, permanents), optionally restricted by a
// controller clause ("you control", "an opponent controls", "you don't
// control"). Examples: "Untap up to two lands." (Snap), "Untap up to three
// lands." (Frantic Search), "Untap up to two creatures you control." Every
// richer qualifier — a subtype, color, supertype, tapped/untapped, attacking,
// mana-value, or keyword rider — fails closed so unsupported untap wordings keep
// failing the round-trip.
func exactBoundedUntapEffectSyntax(effect *EffectSyntax) bool {
	if effect.Context != EffectContextController ||
		!effect.Amount.RangeKnown ||
		effect.Amount.Minimum != 0 ||
		effect.Amount.Maximum < 2 ||
		effect.Amount.Maximum > 10 {
		return false
	}
	word, ok := cardinalWord(effect.Amount.Maximum)
	if !ok {
		return false
	}
	selection := effect.Selection
	if selection.All ||
		selection.Another ||
		selection.Other ||
		selection.Attacking ||
		selection.Blocking ||
		selection.Tapped ||
		selection.Untapped ||
		selection.Colorless ||
		selection.Multicolored ||
		selection.BasicLandType ||
		selection.PlayerOrPlaneswalker ||
		selection.MatchManaValue ||
		selection.MatchPower ||
		selection.MatchToughness ||
		selection.Keyword != KeywordUnknown ||
		selection.ExcludedKeyword != KeywordUnknown ||
		selection.Zone != zone.None ||
		len(selection.ExcludedTypes) != 0 ||
		len(selection.SourceTypes) != 0 ||
		len(selection.Supertypes) != 0 ||
		len(selection.ExcludedSupertypes) != 0 ||
		len(selection.ColorsAny) != 0 ||
		len(selection.ExcludedColors) != 0 ||
		len(selection.SubtypesAny) != 0 ||
		len(selection.Alternatives) != 0 ||
		!selectionRedundantRequiredNoun(selection) {
		return false
	}
	noun, ok := permanentSelectionNoun(selection.Kind)
	if !ok {
		return false
	}
	phrase, ok := targetControllerSuffix(noun+"s", selection.Controller)
	if !ok {
		return false
	}
	return strings.EqualFold(
		exactEffectClauseText(effect),
		fmt.Sprintf("Untap up to %s %s.", word, phrase),
	)
}

func exactHandLibraryPutEffectSyntax(effect *EffectSyntax) bool {
	if !effect.HandLibraryPut.Present ||
		effect.Amount.DynamicForm != EffectDynamicAmountFormNone ||
		!exactLegacyFixedAmountSyntax(effect) {
		return false
	}
	noun := "cards"
	if effect.Amount.Value == 1 {
		noun = "card"
	}
	return strings.EqualFold(
		exactEffectClauseText(effect),
		fmt.Sprintf(
			"Put %s %s from your hand on top of your library in any order.",
			effectAmountSourceText(effect),
			noun,
		),
	)
}

// exactSelfSacrificeEffectSyntax recognizes the controller sacrificing the
// source permanent itself by self-reference — "Sacrifice this creature.",
// "Sacrifice <this card's name>." — including the optional "You may sacrifice
// this creature." offer (exactEffectClauseText strips the "you may" prefix).
// The pronoun form "Sacrifice it." is handled by exactDirectPronounEffectSyntax.
func exactSelfSacrificeEffectSyntax(effect *EffectSyntax) bool {
	if effect.Context != EffectContextController ||
		len(effect.Targets) != 0 ||
		effect.Duration != EffectDurationNone ||
		effect.Negated {
		return false
	}
	object, ok := exactSelfSubjectReferenceText(effect.References)
	if !ok {
		return false
	}
	return strings.EqualFold(exactEffectClauseText(effect), "Sacrifice "+object+".")
}

func exactSacrificeChoiceEffectSyntax(effect *EffectSyntax) bool {
	// Spelled cardinal amounts run from "one" through "ten"; the runtime
	// SacrificePermanents primitive carries a fixed count, so every amount in
	// that range reconstructs and lowers uniformly across subjects.
	if !effect.Amount.Known || effect.Amount.Value < 1 || effect.Amount.Value > 10 {
		return false
	}
	noun, ok := sacrificeChoiceNoun(&effect.Selection, effect.Amount.Value > 1)
	if !ok {
		return false
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
	case EffectContextEachOtherPlayer:
		subject = "Each other player"
	case EffectContextEachPlayer:
		subject = "Each player"
	case EffectContextDefendingPlayer:
		subject = "Defending player"
	case EffectContextReferencedPlayer:
		subject = "That player"
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

// sacrificeChoiceNoun reconstructs the permanent noun phrase of a sacrifice
// effect ("creature", "nontoken creature", "creature or planeswalker",
// "nonland permanent", "Blood token"), optionally pluralized. It fails closed
// for selector shapes the runtime sacrifice selection cannot express so the
// effect stays unsupported.
func sacrificeChoiceNoun(selection *SelectionSyntax, plural bool) (string, bool) {
	base, ok := sacrificeChoiceBaseNoun(selection, plural)
	if !ok {
		return "", false
	}
	// The token-suffix forms ("Blood token", "a token") carry the token qualifier
	// as a trailing word, which sacrificeChoiceBaseNoun has already appended; a
	// card-type token ("token creature") instead leads with the qualifier.
	switch {
	case selection.NonToken:
		base = "nontoken " + base
	case selection.TokenOnly && !sacrificeTokenSuffixForm(selection):
		base = "token " + base
	default:
	}
	return base, true
}

// sacrificeTokenSuffixForm reports whether the selector names a token by a
// subtype ("Blood token") or by no type at all ("a token"), where Oracle places
// the "token" word last. A token named by a card type ("token creature") keeps
// the qualifier leading and is excluded here.
func sacrificeTokenSuffixForm(selection *SelectionSyntax) bool {
	return selection.TokenOnly &&
		selection.Kind == SelectionUnknown &&
		len(selection.RequiredTypesAny) == 0 &&
		len(selection.ExcludedTypes) == 0 &&
		len(selection.SubtypesAny) <= 1
}

// sacrificeChoiceBaseNoun maps the selector kind (or a card-type union such as
// "creature or planeswalker") to its printed noun, rejecting selectors carrying
// qualifiers the sacrifice reconstruction does not model.
func sacrificeChoiceBaseNoun(selection *SelectionSyntax, plural bool) (string, bool) {
	if len(selection.ExcludedSubtypes) > 1 {
		return "", false
	}
	if len(selection.RequiredTypesAny) > 1 {
		if len(selection.ExcludedSubtypes) != 0 {
			return "", false
		}
		words := make([]string, 0, len(selection.RequiredTypesAny))
		for _, cardType := range selection.RequiredTypesAny {
			word, ok := searchFilterCardTypeWord(cardType)
			if !ok {
				return "", false
			}
			if plural {
				word += "s"
			}
			words = append(words, word)
		}
		return joinOrList(words), true
	}
	// A token named by no type at all ("a token") or by a subtype ("a Blood
	// token") renders the "token" word last. The bare form matches any token; the
	// subtype form matches that token subtype through the SubtypesAny filter.
	if sacrificeTokenSuffixForm(selection) && len(selection.ExcludedSubtypes) == 0 {
		noun := "token"
		if len(selection.SubtypesAny) == 1 {
			noun = string(selection.SubtypesAny[0]) + " token"
		}
		if plural {
			noun += "s"
		}
		return noun, true
	}
	// A bare subtype names the permanent by its subtype alone, with no card-type
	// kind: an artifact token ("a Treasure"/"a Food"), a land type ("a Forest"),
	// or a creature subtype ("a Goblin"). The subtype noun is printed verbatim
	// ("Treasure"/"Treasures"); the runtime sacrifice selection matches it
	// through the SubtypesAny filter. Irregular plurals ("Elves") simply fail the
	// byte-exact reconstruction and stay unsupported.
	if selection.Kind == SelectionUnknown &&
		len(selection.RequiredTypesAny) == 0 &&
		len(selection.ExcludedTypes) == 0 &&
		len(selection.ExcludedSubtypes) == 0 &&
		len(selection.SubtypesAny) == 1 {
		noun := string(selection.SubtypesAny[0])
		if plural {
			noun += "s"
		}
		return noun, true
	}
	noun := ""
	switch selection.Kind {
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
		return "", false
	}
	if plural {
		noun += "s"
	}
	// A single excluded card type renders as a "non<type>" prefix ("nonland
	// permanent", "noncreature artifact", "nonartifact creature"). The runtime
	// sacrifice selection matches it through the ExcludedTypes filter. More than
	// one excluded type has no canonical sacrifice wording, so it fails closed.
	switch len(selection.ExcludedTypes) {
	case 0:
	case 1:
		if len(selection.ExcludedSubtypes) != 0 {
			return "", false
		}
		word, ok := searchFilterCardTypeWord(selection.ExcludedTypes[0])
		if !ok {
			return "", false
		}
		noun = "non" + word + " " + noun
	default:
		return "", false
	}
	// A single excluded creature subtype renders as a hyphenated "non-<Subtype>"
	// prefix ("non-Zombie creature", "non-Demon creature"). The runtime sacrifice
	// selection matches it through the ExcludedSubtype filter. It never combines
	// with an excluded card type (rejected above), and more than one excluded
	// subtype has no canonical sacrifice wording (rejected at entry).
	if len(selection.ExcludedSubtypes) == 1 {
		noun = "non-" + string(selection.ExcludedSubtypes[0]) + " " + noun
	}
	return noun, true
}

func exactSearchEffectSyntax(effect *EffectSyntax) bool {
	return analyzeSearchClause(effect).detail == ""
}

// searchUnsupportedDetail reports the fail-closed diagnostic for a library-search
// clause, or "" when the clause is supported. See analyzeSearchClause for the
// recognized envelope.
func searchUnsupportedDetail(effect *EffectSyntax) string {
	return analyzeSearchClause(effect).detail
}

// searchHeterogeneousSlotSubtypes recognizes a search-and-put clause whose noun
// phrase names two distinct singular card slots joined by a plain "and" — "a
// Forest card and a Plains card" (Krosan Verge). Each slot is one basic land
// subtype, and the searching player finds one card matching each slot. It
// reconstructs the canonical "a <Sub0> card and a <Sub1> card" noun phrase
// byte-for-byte from the parsed subtype union and the source text, returning the
// per-slot subtypes only on an exact match. It fails closed (returns nil) for
// every other count, filter, supertype, name, mana-value rider, optional or
// non-controller searcher, "or" union, "and/or" inclusive join, repeated
// subtype, or non-basic-land subtype, so an ordinary single-filter search is
// never reinterpreted as a multi-slot one.
func searchHeterogeneousSlotSubtypes(effect *EffectSyntax) []types.Sub {
	if effect.Optional || effect.Context != EffectContextController {
		return nil
	}
	if !effect.Amount.Known || effect.Amount.Value != 1 {
		return nil
	}
	sel := effect.Selection
	if sel.Kind != SelectionCard ||
		len(sel.SubtypesAny) != 2 ||
		sel.BasicLandType ||
		len(sel.Alternatives) != 0 ||
		len(sel.RequiredTypesAny) != 0 ||
		len(sel.Supertypes) != 0 ||
		len(sel.ColorsAny) != 0 ||
		len(sel.ExcludedSubtypes) != 0 ||
		sel.RequiredName != "" ||
		sel.MatchManaValue || sel.MatchPower || sel.MatchToughness ||
		sel.Colorless || sel.Multicolored {
		return nil
	}
	if !allBasicLandSubtypes(sel.SubtypesAny) || sel.SubtypesAny[0] == sel.SubtypesAny[1] {
		return nil
	}
	prefix, text := searchClausePrefix(effect)
	if !strings.HasPrefix(text, prefix) {
		return nil
	}
	rest := strings.TrimPrefix(text, prefix)
	noun := "a " + string(sel.SubtypesAny[0]) + " card and a " + string(sel.SubtypesAny[1]) + " card"
	if !strings.HasPrefix(rest, noun+", ") {
		return nil
	}
	return slices.Clone(sel.SubtypesAny)
}

// searchSharedSubtypeRider reports whether a library-search clause carries the
// supported "that share a land type" correlation rider, requiring every found
// card to share a land subtype with the others. It is the structured companion
// to the byte-exact reconstruction in analyzeSearchClause; both agree because
// they share that one recognizer.
func searchSharedSubtypeRider(effect *EffectSyntax) bool {
	return analyzeSearchClause(effect).sharedSubtype
}

// searchDestinationPosition reports the ordered destination carried by an exact
// search clause. The zero value denotes the ordinary hand/battlefield families.
func searchDestinationPosition(effect *EffectSyntax) EffectDestinationPosition {
	return analyzeSearchClause(effect).destinationPosition
}

// searchControlRider reports the controller rider an exact search-and-put clause
// carries ("under target player's control"). The zero value denotes no rider:
// the found card enters under the searching player's control.
func searchControlRider(effect *EffectSyntax) SearchControlRider {
	return analyzeSearchClause(effect).control
}

// searchClauseAnalysis carries the structured outcome of analyzeSearchClause: a
// fail-closed diagnostic detail (empty when supported) and the riders the
// recognized clause carries.
type searchClauseAnalysis struct {
	detail              string
	sharedSubtype       bool
	destinationPosition EffectDestinationPosition
	control             SearchControlRider
}

// analyzeSearchClause reconstructs the canonical library-search clause from the
// parsed Selection and count and compares it byte-for-byte against the source.
// It recognizes the bounded shapes the runtime models: a singular or "up to N"
// search of your own library for a plain card-type, a basic land, a union of
// basic land subtypes (optionally "basic"), a permanent card (optionally with a
// subtype, e.g. "Rebel permanent"), optionally a "legendary" supertype, and
// optionally a "with mana value N", "with mana value N or less", or "with mana
// value N or greater" rider, optionally a "with power/
// toughness N or less/greater" rider, optionally a "that share a
// land type" correlation rider on a multi-card land search, moved to hand or the
// battlefield (optionally tapped, optionally revealed first), ending with "then
// shuffle". It returns detail="" when the clause is supported, or a diagnostic
// detail otherwise, plus whether the correlation rider was recognized. Every
// richer rider (graveyard search, "with different names", X-derived mana-value
// bounds, "for each player", X counts) fails closed.
func analyzeSearchClause(effect *EffectSyntax) searchClauseAnalysis {
	var sharedSubtype bool
	prefix, text := searchClausePrefix(effect)
	if !strings.HasPrefix(text, prefix) {
		return searchClauseAnalysis{detail: `the executable source backend supports only exact searches of your library`, sharedSubtype: false, destinationPosition: EffectDestinationUnspecified, control: SearchControlRiderNone}
	}
	rest := strings.TrimPrefix(text, prefix)

	consumed, amount, plural := searchCountPrefix(rest)
	dynamic := consumed == "up to X "
	switch {
	case consumed == "":
		return searchClauseAnalysis{detail: "the executable source backend supports only exact singular-card search wording", sharedSubtype: false, destinationPosition: EffectDestinationUnspecified, control: SearchControlRiderNone}
	case dynamic:
		// The "up to X" count defers its bound to a resolving "where X is ..."
		// rules-derived amount the parser recognized on this effect. Require that
		// typed dynamic amount so an unrecognized "X" fails closed.
		if effect.Amount.DynamicForm != EffectDynamicAmountFormWhereX ||
			effect.Amount.DynamicKind == EffectDynamicAmountNone {
			return searchClauseAnalysis{detail: "the executable source backend supports only exact singular-card search wording", sharedSubtype: false, destinationPosition: EffectDestinationUnspecified, control: SearchControlRiderNone}
		}
	case !effect.Amount.Known || effect.Amount.Value != amount:
		return searchClauseAnalysis{detail: "the executable source backend supports only exact singular-card search wording", sharedSubtype: false, destinationPosition: EffectDestinationUnspecified, control: SearchControlRiderNone}
	default:
	}
	rest = rest[len(consumed):]

	noun := ""
	switch {
	case strings.HasPrefix(rest, "land card with a basic land type"):
		if effect.Selection.Kind != SelectionLand {
			return searchClauseAnalysis{detail: unsupportedSearchFilterDetail(rest), sharedSubtype: false, destinationPosition: EffectDestinationUnspecified, control: SearchControlRiderNone}
		}
		if !effect.Selection.BasicLandType {
			if len(effect.Selection.Supertypes) != 1 ||
				effect.Selection.Supertypes[0] != SupertypeBasic {
				return searchClauseAnalysis{detail: unsupportedSearchFilterDetail(rest), sharedSubtype: false, destinationPosition: EffectDestinationUnspecified, control: SearchControlRiderNone}
			}
			effect.Selection.Supertypes = nil
			effect.Selection.BasicLandType = true
		} else if len(effect.Selection.Supertypes) != 0 {
			return searchClauseAnalysis{detail: unsupportedSearchFilterDetail(rest), sharedSubtype: false, destinationPosition: EffectDestinationUnspecified, control: SearchControlRiderNone}
		}
		noun = "land card with a basic land type"
		if plural {
			noun += "s"
		}
	case len(effect.Selection.Alternatives) > 0:
		disjunction, ok := canonicalSearchDisjunctionNoun(effect.Selection, rest, plural)
		if !ok {
			return searchClauseAnalysis{detail: unsupportedSearchFilterDetail(rest), sharedSubtype: false, destinationPosition: EffectDestinationUnspecified, control: SearchControlRiderNone}
		}
		noun = disjunction
	default:
		filter, ok := canonicalSearchFilter(effect.Selection)
		if !ok {
			return searchClauseAnalysis{detail: unsupportedSearchFilterDetail(rest), sharedSubtype: false, destinationPosition: EffectDestinationUnspecified, control: SearchControlRiderNone}
		}
		noun = "card"
		if filter != "" {
			noun = filter + " card"
		}
		if plural {
			noun += "s"
		}
	}
	if effect.Selection.RequiredName != "" {
		noun += " named " + effect.Selection.RequiredName
	}
	riderText := ""
	numericRiders := 0
	if effect.Selection.MatchManaValue {
		numericRiders++
		rider, ok := searchManaValueRider(effect.Selection)
		if !ok {
			return searchClauseAnalysis{detail: unsupportedSearchFilterDetail(rest), sharedSubtype: false, destinationPosition: EffectDestinationUnspecified, control: SearchControlRiderNone}
		}
		riderText = rider
	}
	if effect.Selection.MatchPower {
		numericRiders++
		rider, ok := searchCharacteristicRider("power", effect.Selection.Power)
		if !ok {
			return searchClauseAnalysis{detail: unsupportedSearchFilterDetail(rest), sharedSubtype: false, destinationPosition: EffectDestinationUnspecified, control: SearchControlRiderNone}
		}
		riderText = rider
	}
	if effect.Selection.MatchToughness {
		numericRiders++
		rider, ok := searchCharacteristicRider("toughness", effect.Selection.Toughness)
		if !ok {
			return searchClauseAnalysis{detail: unsupportedSearchFilterDetail(rest), sharedSubtype: false, destinationPosition: EffectDestinationUnspecified, control: SearchControlRiderNone}
		}
		riderText = rider
	}
	if numericRiders > 1 {
		return searchClauseAnalysis{detail: unsupportedSearchFilterDetail(rest), sharedSubtype: false, destinationPosition: EffectDestinationUnspecified, control: SearchControlRiderNone}
	}
	afterNoun, ok := strings.CutPrefix(rest, noun+riderText)
	if !ok {
		return searchClauseAnalysis{detail: unsupportedSearchFilterDetail(rest), sharedSubtype: false, destinationPosition: EffectDestinationUnspecified, control: SearchControlRiderNone}
	}
	if dynamic {
		// The "up to X" count's resolving bound is printed inline as a trailing
		// "where X is ..." clause the parser captured as the effect amount. Strip
		// its verbatim text from the reconstruction so the destination match
		// resumes at the put phrase; a clause whose amount text does not round-trip
		// fails closed.
		stripped, ok := strings.CutPrefix(afterNoun, ", "+effect.Amount.Text)
		if !ok {
			return searchClauseAnalysis{detail: unsupportedSearchFilterDetail(rest), sharedSubtype: false, destinationPosition: EffectDestinationUnspecified, control: SearchControlRiderNone}
		}
		afterNoun = stripped
	}
	if afterNoun == "." {
		// A two-sentence search ("Search your library for <filter>[, where X is
		// ...]. Put those cards onto the battlefield, then shuffle.") ends the
		// search sentence after the filter and any count phrase; its destination
		// is a separate following put effect that lowering validates and lowers as
		// the search destination. The search clause itself is exact.
		return searchClauseAnalysis{detail: "", sharedSubtype: false, destinationPosition: EffectDestinationUnspecified, control: SearchControlRiderNone}
	}
	if remainder, ok := strings.CutPrefix(afterNoun, searchSharedSubtypeRiderText); ok {
		// "that share a land type" correlates the found cards: each must share a
		// land subtype with the others. It is modeled only for the two-card basic
		// land search ("up to two basic land cards"), where the subtype is
		// meaningful and the runtime can enforce a legal pair (Myriad Landscape);
		// any other count or filter fails closed.
		if amount != 2 || effect.Selection.Kind != SelectionLand {
			return searchClauseAnalysis{detail: "the executable source backend supports the shared-land-type rider only on a two-card basic-land search", sharedSubtype: false, destinationPosition: EffectDestinationUnspecified, control: SearchControlRiderNone}
		}
		afterNoun = remainder
		sharedSubtype = true
	}
	destination, ok := strings.CutPrefix(afterNoun, ", ")
	if !ok {
		return searchClauseAnalysis{detail: unsupportedSearchFilterDetail(rest), sharedSubtype: false, destinationPosition: EffectDestinationUnspecified, control: SearchControlRiderNone}
	}
	if base, control := stripSearchControlRider(destination); control != SearchControlRiderNone {
		// "put it onto the battlefield ... under target player's control" makes the
		// found permanent enter under the named target player's control rather than
		// the searching player's. The rider attaches only to a battlefield put, so
		// the rider-free base must be a supported battlefield destination; any other
		// base (hand, library top, a split put) fails closed.
		if !sharedSubtype && searchDestinationSupported(base, plural) && strings.Contains(base, "onto the battlefield") {
			return searchClauseAnalysis{detail: "", sharedSubtype: false, destinationPosition: EffectDestinationUnspecified, control: control}
		}
		return searchClauseAnalysis{detail: "the executable source backend supports the \"under target player's control\" rider only on a battlefield search destination", sharedSubtype: false, destinationPosition: EffectDestinationUnspecified, control: SearchControlRiderNone}
	}
	if searchSplitDestinationSupported(destination) {
		// A split destination ("put one ... and the other ...") distributes the
		// found cards across two single-card slots, so it requires exactly the
		// two-card "up to two" search; any other count fails closed. The
		// correlation rider is not modeled in combination with a split
		// destination, so reject that pairing.
		if amount != 2 || sharedSubtype {
			return searchClauseAnalysis{detail: "the executable source backend supports a split search destination only for an \"up to two\" search", sharedSubtype: false, destinationPosition: EffectDestinationUnspecified, control: SearchControlRiderNone}
		}
		return searchClauseAnalysis{detail: "", sharedSubtype: false, destinationPosition: EffectDestinationUnspecified, control: SearchControlRiderNone}
	}
	if searchDestinationSupported(destination, plural) {
		return searchClauseAnalysis{detail: "", sharedSubtype: sharedSubtype, destinationPosition: EffectDestinationUnspecified, control: SearchControlRiderNone}
	}
	if base, ok := stripSearchRiderClause(destination); ok && searchDestinationSupported(base, plural) {
		// A supported rider ("discard a card at random", "you lose N life") may
		// sit between the put phrase and the trailing "then shuffle." The rider is
		// compiled as its own effect that lowering validates and lowers after the
		// search; here we only confirm the base destination is one the runtime
		// models so the search clause itself stays exact.
		return searchClauseAnalysis{detail: "", sharedSubtype: sharedSubtype, destinationPosition: EffectDestinationUnspecified, control: SearchControlRiderNone}
	}
	if base, ok := stripSearchLifeGainRider(destination); ok && searchDestinationSupported(base, plural) {
		// A "you gain N life" reward may close the search sentence, joined to the
		// trailing "then shuffle" by "and" — "..., then shuffle and you gain 1
		// life." (the Cabaretti Courtyard tapped-fetch land cycle). The life gain
		// is compiled as its own effect that lowering validates and lowers after
		// the search; here we only confirm the rider-free base destination is one
		// the runtime models so the search clause itself stays exact.
		return searchClauseAnalysis{detail: "", sharedSubtype: sharedSubtype, destinationPosition: EffectDestinationUnspecified, control: SearchControlRiderNone}
	}
	if !plural && searchTopDestinationSupported(destination) && !sharedSubtype {
		return searchClauseAnalysis{detail: "", sharedSubtype: false, destinationPosition: EffectDestinationTop, control: SearchControlRiderNone}
	}
	return searchClauseAnalysis{detail: "the executable source backend supports only exact hand, battlefield, or singular library-top search destinations", sharedSubtype: false, destinationPosition: EffectDestinationUnspecified, control: SearchControlRiderNone}
}

// searchShuffleSuffix is the canonical trailing clause every shuffle-terminated
// library-search destination ends with.
const searchShuffleSuffix = ", then shuffle."

// stripSearchRiderClause removes a recognized rider clause inserted between a
// search's put phrase and its trailing "then shuffle." It returns the
// rider-free destination and true when a supported rider is present, so the base
// destination can be matched against the destination whitelist. Supported riders
// mirror the riders lowering can lower after the search: a random discard and a
// fixed controller life loss.
func stripSearchRiderClause(destination string) (string, bool) {
	head, ok := strings.CutSuffix(destination, searchShuffleSuffix)
	if !ok {
		return destination, false
	}
	for _, rider := range []string{"discard a card at random"} {
		if base, ok := strings.CutSuffix(head, ", "+rider); ok {
			return base + searchShuffleSuffix, true
		}
	}
	if base, ok := stripSearchLifeLossRider(head); ok {
		return base + searchShuffleSuffix, true
	}
	return destination, false
}

// stripSearchLifeLossRider removes a "you lose N life" rider (N a positive
// integer) from the end of a search destination head.
func stripSearchLifeLossRider(head string) (string, bool) {
	idx := strings.LastIndex(head, ", you lose ")
	if idx < 0 {
		return head, false
	}
	amount, ok := strings.CutSuffix(head[idx+len(", you lose "):], " life")
	if !ok || amount == "" {
		return head, false
	}
	if _, err := strconv.Atoi(amount); err != nil {
		return head, false
	}
	return head[:idx], true
}

// stripSearchLifeGainRider removes a trailing "and you gain N life" reward (N a
// positive integer) that closes a search sentence after its "then shuffle" — the
// Cabaretti Courtyard tapped-fetch land cycle ends "..., then shuffle and you
// gain 1 life." It returns the rider-free destination (ending again at "then
// shuffle.") and true when the rider is present, so the base destination can be
// matched against the destination whitelist. The life gain is compiled as its
// own effect that lowering lowers after the search.
func stripSearchLifeGainRider(destination string) (string, bool) {
	head, ok := strings.CutSuffix(destination, " life.")
	if !ok {
		return destination, false
	}
	idx := strings.LastIndex(head, " and you gain ")
	if idx < 0 {
		return destination, false
	}
	amount := head[idx+len(" and you gain "):]
	if amount == "" {
		return destination, false
	}
	if _, err := strconv.Atoi(amount); err != nil {
		return destination, false
	}
	base := head[:idx]
	if !strings.HasSuffix(base, "then shuffle") {
		return destination, false
	}
	return base + ".", true
}

// rider that follows the searched noun phrase, requiring every found card to
// share a land subtype with the others (Myriad Landscape).
const searchSharedSubtypeRiderText = " that share a land type"

// stripSearchControlRider removes an "under target player's control" /
// "under target opponent's control" controller rider that sits between a search
// clause's battlefield put phrase and its trailing "then shuffle." It returns
// the rider-free destination and the recognized rider, or the unchanged
// destination and SearchControlRiderNone when no rider is present. The found
// permanent enters under the named target player's control instead of the
// searching player's.
func stripSearchControlRider(destination string) (string, SearchControlRider) {
	head, ok := strings.CutSuffix(destination, searchShuffleSuffix)
	if !ok {
		return destination, SearchControlRiderNone
	}
	for _, rider := range []struct {
		text  string
		rider SearchControlRider
	}{
		{" under target player's control", SearchControlRiderTargetPlayer},
		{" under target opponent's control", SearchControlRiderTargetOpponent},
	} {
		if base, ok := strings.CutSuffix(head, rider.text); ok {
			return base + searchShuffleSuffix, rider.rider
		}
	}
	return destination, SearchControlRiderNone
}

// searchClausePrefix selects the canonical "search ... library for " prefix the
// clause must reconstruct against and returns it alongside the (possibly
// normalized) source text to match. Four searcher forms are recognized:
//
//   - The affected-permanent's-controller optional rider on a removal spell —
//     "Exile target creature. Its controller may search their library for a basic
//     land card, ... then shuffle." The parser marks the clause Optional; the
//     literal "Its controller may search their library" is reconstructed
//     verbatim (the "may" is kept, not stripped) so the byte-exact comparison
//     still holds, and the executable backend routes the search-or-decline
//     choice to that affected player rather than the spell's controller.
//   - A "You may search your library ..." optional self-tutor, whose "you may"
//     choice is carried by effect.Optional. The prefix is stripped and the
//     sentence-initial capital restored so the remaining clause reconstructs
//     against the same canonical "Search your library for ..." shape as a
//     mandatory tutor; the optionality is preserved separately by
//     effect.Optional. This mirrors the optional-prefix handling in
//     exactEffectClauseText.
//   - A mandatory controller tutor — "Search your library for ...".
//   - A mandatory controller tutor embedded after a leading clause, most often a
//     triggered ability's condition — "When this creature enters, search your
//     library for ...". Such a clause is not sentence-initial, so its verb is
//     lowercase ("search"); it reconstructs against the lowercase prefix and
//     lowers to the same search as a sentence-initial tutor.
//
// Any other searcher wording falls through to the controller prefix and fails
// the prefix check in the caller (fail closed).
func searchClausePrefix(effect *EffectSyntax) (prefix, text string) {
	const controllerPrefix = "Search your library for "
	const lowerControllerPrefix = "search your library for "
	const affectedPlayerPrefix = "That player may search their library for "
	text = trimLeadingInterveningCondition(effect.Text)
	text = stripMandatoryReflexiveConnector(text)
	// A clause-initial "instead" marks a conditional replacement search ("If
	// <condition>, instead search your library ..."); strip it so the search
	// wording that follows reconstructs against the canonical prefix. The
	// replacement relationship itself is carried by effect.Replacement.
	if rest, ok := strings.CutPrefix(text, "instead "); ok {
		text = rest
	}
	// A targeted-player searcher ("Target player searches their library for ...",
	// "Target opponent searches their library for ...") performs the search from
	// that single target player's library. The clause reads in the third person
	// ("searches", "puts", "shuffles", "their hand"); normalize the searcher's
	// verbs and possessives to the canonical controller second-person form and
	// reconstruct against the standard "Search your library for ..." prefix so the
	// shared recognizer validates the count, filter, and destination. Lowering
	// resolves the searcher to the ability's target player. The "may" optional and
	// plural-subject forms are not this single-target shape and fall through.
	if !effect.Optional && effect.Context == EffectContextTarget &&
		len(effect.Targets) == 1 &&
		exactCardCountTargetPlayer(effect.Targets[0].Selection) {
		subjectPrefix := effect.Targets[0].Text + " searches their library for "
		if rest, ok := strings.CutPrefix(text, subjectPrefix); ok {
			return controllerPrefix, controllerPrefix + normalizeThirdPersonSearchRest(rest)
		}
	}
	// An each-player searcher ("Each player searches their library for ...")
	// performs a symmetric search from every player's own library; the found
	// card enters under each searching player's control. The clause reads in the
	// third person ("searches", "puts", "shuffles"); normalize the searcher's
	// verbs to the canonical controller second-person form and reconstruct
	// against the standard "Search your library for ..." prefix so the shared
	// recognizer validates the count, filter, and destination. Lowering resolves
	// the searcher to the all-players group. The "may" optional form is not this
	// shape and falls through.
	if !effect.Optional && effect.Context == EffectContextEachPlayer {
		const eachPlayerPrefix = "Each player searches their library for "
		if rest, ok := strings.CutPrefix(text, eachPlayerPrefix); ok {
			return controllerPrefix, controllerPrefix + normalizeThirdPersonSearchRest(rest)
		}
	}
	// A referenced-object-controller searcher ("Its controller may search …",
	// "That land's controller may search …") reconstructs its prefix from the
	// subject reference's verbatim text, so any possessive object form — not just
	// the creature pronoun "Its" — round-trips byte-exactly to the same search.
	if effect.Optional && effect.Context == EffectContextReferencedObjectController &&
		len(effect.SubjectReferences) == 1 {
		riderPrefix := effect.SubjectReferences[0].Text + " controller may search their library for "
		if strings.HasPrefix(text, riderPrefix) {
			return riderPrefix, text
		}
	}
	if effect.Optional && strings.HasPrefix(text, affectedPlayerPrefix) {
		return affectedPlayerPrefix, text
	}
	if effect.Optional {
		if rest, ok := strings.CutPrefix(text, "You may "); ok {
			text = titleFirstEffectText(rest)
		} else if rest, ok := strings.CutPrefix(text, "you may "); ok {
			text = titleFirstEffectText(rest)
		}
	}
	if strings.HasPrefix(text, lowerControllerPrefix) {
		return lowerControllerPrefix, text
	}
	return controllerPrefix, text
}

// normalizeThirdPersonSearchRest rewrites the third-person verbs and possessive
// of a single target player's search clause ("... searches their library for"
// having already been consumed) into the canonical controller second-person form
// so the shared search recognizer can validate the remaining count, filter, and
// destination unchanged. Only the fixed search-clause verbs (reveal, put,
// shuffle) and the destination possessive ("into their hand/graveyard") are
// conjugated; every other word, including the searched filter and any "named"
// clause, is left untouched.
func normalizeThirdPersonSearchRest(rest string) string {
	for _, sub := range []struct{ from, to string }{
		{"reveals ", "reveal "},
		{"puts ", "put "},
		{"shuffles ", "shuffle "},
		{"shuffles.", "shuffle."},
		{"into their hand", "into your hand"},
		{"into their graveyard", "into your graveyard"},
	} {
		rest = strings.ReplaceAll(rest, sub.from, sub.to)
	}
	return rest
}

// trimLeadingInterveningCondition removes a leading intervening-if condition
// clause ("if ...,", "unless ...,", "as long as ...,", "only if ...,") from a
// search effect's text so the byte-exact clause reconstruction can match the
// search wording that follows. A triggered ability that gates its search behind
// such a condition (Land Tax: "...if an opponent controls more lands than you,
// you may search...") keeps the condition on the search effect's text; the
// condition itself is recognized and lowered separately, so the search clause
// must be analyzed without it. The clause ends at the first comma, matching the
// parser's own condition-clause boundary (conditionClauseEnd).
func trimLeadingInterveningCondition(text string) string {
	lower := strings.ToLower(text)
	for _, intro := range []string{"if ", "unless ", "as long as ", "only if "} {
		if strings.HasPrefix(lower, intro) {
			if _, after, ok := strings.Cut(text, ", "); ok {
				return after
			}
		}
	}
	return text
}

// stripMandatoryReflexiveConnector removes a leading mandatory-reflexive "When
// you do, " connector from a search effect's text so the byte-exact clause
// reconstruction can match the search wording that follows. The parser leaves a
// "When you do," after a *mandatory* action in-sentence and resolves its trailing
// effect unconditionally (parseConditionIntro only converts the reflexive into an
// "if you did" gate when a preceding "you may" makes the action optional). The
// connector's literal presence therefore guarantees the unconditional form, as on
// the Cabaretti Courtyard tapped-fetch land cycle ("When this land enters,
// sacrifice it. When you do, search your library for ..."), so stripping it for
// reconstruction is safe; the always-performed sequencing is unchanged.
func stripMandatoryReflexiveConnector(text string) string {
	if rest, ok := strings.CutPrefix(text, "When you do, "); ok {
		return rest
	}
	return text
}

// searchCountPrefix consumes the count phrase that follows "for ". It accepts the
// singular articles "a "/"an " (amount 1) and the bounded "up to <word> " form
// (amount 2..10, plural). The dynamic "up to X " form (amount 0, plural) defers
// its bound to the effect's resolving "where X is ..." count, validated by the
// caller against the typed amount. It returns the consumed literal (empty when
// the phrase is unrecognized) so the caller can keep reconstructing the clause
// byte-for-byte.
func searchCountPrefix(rest string) (consumed string, amount int, plural bool) {
	switch {
	case strings.HasPrefix(rest, "a "):
		return "a ", 1, false
	case strings.HasPrefix(rest, "an "):
		return "an ", 1, false
	case strings.HasPrefix(rest, "up to X "):
		return "up to X ", 0, true
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

// searchManaValueRider reconstructs the "with mana value N", "with mana value N
// or less", "with mana value N or greater", or "with mana value X or less"
// filter rider from the parsed selection. A fixed exact, upper, or lower bound
// mirrors the concrete Selection.ManaValue comparison the SearchSpec carries; an
// X-derived bound mirrors SearchSpec.MaxManaValueFromX (Green Sun's Zenith,
// Wargate). The runtime evaluates the bound with compare.Int.Matches for any of
// these operators, exactly as the power/toughness riders do. Every other
// comparison fails closed.
func searchManaValueRider(sel SelectionSyntax) (string, bool) {
	if sel.ManaValueX {
		return " with mana value X or less", true
	}
	switch sel.ManaValue.Op {
	case compare.Equal:
		return fmt.Sprintf(" with mana value %d", sel.ManaValue.Value), true
	case compare.LessOrEqual:
		return fmt.Sprintf(" with mana value %d or less", sel.ManaValue.Value), true
	case compare.GreaterOrEqual:
		return fmt.Sprintf(" with mana value %d or greater", sel.ManaValue.Value), true
	default:
		return "", false
	}
}

// searchCharacteristicRider reconstructs a "with <characteristic> N or less" or
// "with <characteristic> N or greater" filter rider from a parsed power or
// toughness comparison, mirroring SearchSpec.Max/Min Power and Toughness. The
// less-or-equal bound mirrors the Max field; the greater-or-equal bound mirrors
// the Min field. Every other comparison (exact, less-than, greater-than) fails
// closed.
func searchCharacteristicRider(characteristic string, bound compare.Int) (string, bool) {
	switch bound.Op {
	case compare.LessOrEqual:
		return fmt.Sprintf(" with %s %d or less", characteristic, bound.Value), true
	case compare.GreaterOrEqual:
		return fmt.Sprintf(" with %s %d or greater", characteristic, bound.Value), true
	default:
		return "", false
	}
}

// canonicalSearchFilter renders the modeled portion of a search filter (the text
// between the article and " card") from the parsed Selection, returning ok=false
// for any attribute the runtime SearchSpec cannot express. Supported filters are
// a plain card, a single card type (land/creature/artifact/enchantment/
// planeswalker), a permanent card, optionally "basic" or "legendary", a subtype
// union with no separate type noun ("Forest or Island", "Sliver", "Aura or
// Equipment"), and a subtype paired with a card type or "permanent" ("Myr
// creature", "Dragon creature", "Rebel permanent"). An optional "with mana value
// N or less" or "with power/toughness N or less/greater" rider is reconstructed
// by the caller, not here.
// canonicalSearchDisjunctionNoun reconstructs the noun phrase of a two-sided
// disjunctive search filter ("creature or basic land card", "basic land card or
// a Gate card") whose sides parsed into Selection.Alternatives. Each side
// reconstructs through canonicalSearchFilter; the two are joined by "or" with
// the trailing "card[s]" placed after one or both sides, matching the variant
// Oracle wordings. The candidate whose joined form prefixes the source text is
// returned so the reconstruction stays byte-exact; it fails closed when no
// candidate matches or a side is not an expressible filter.
func canonicalSearchDisjunctionNoun(sel SelectionSyntax, rest string, plural bool) (string, bool) {
	if len(sel.Alternatives) != 2 {
		return "", false
	}
	first, ok := canonicalSearchFilter(sel.Alternatives[0])
	if !ok || first == "" {
		return "", false
	}
	second, ok := canonicalSearchFilter(sel.Alternatives[1])
	if !ok || second == "" {
		return "", false
	}
	card := "card"
	if plural {
		card = "cards"
	}
	candidates := []string{
		first + " or " + second + " " + card,
		first + " " + card + " or " + second + " " + card,
		first + " " + card + " or a " + second + " " + card,
		first + " " + card + " or an " + second + " " + card,
	}
	for _, candidate := range candidates {
		if strings.HasPrefix(rest, candidate) {
			return candidate, true
		}
	}
	return "", false
}

func canonicalSearchFilter(sel SelectionSyntax) (string, bool) {
	if sel.Controller != SelectionControllerAny ||
		sel.All || sel.Another || sel.Other || sel.Attacking || sel.Blocking ||
		sel.Tapped || sel.Untapped || sel.Multicolored ||
		sel.Keyword != KeywordUnknown || sel.Zone != zone.None ||
		len(sel.ExcludedTypes) != 0 || len(sel.SourceTypes) != 0 ||
		len(sel.ExcludedColors) != 0 {
		return "", false
	}
	colorStr := ""
	if len(sel.ColorsAny) > 0 {
		words := make([]string, 0, len(sel.ColorsAny))
		for _, c := range sel.ColorsAny {
			word, ok := colorWord(c)
			if !ok {
				return "", false
			}
			words = append(words, word)
		}
		colorStr = joinOrList(words)
	}
	if sel.Colorless {
		if colorStr != "" {
			return "", false
		}
		colorStr = "colorless"
	}
	basic, legendary := false, false
	switch len(sel.Supertypes) {
	case 0:
	case 1:
		switch sel.Supertypes[0] {
		case SupertypeBasic:
			basic = true
		case SupertypeLegendary:
			legendary = true
		default:
			return "", false
		}
	default:
		return "", false
	}
	prefix := ""
	switch {
	case basic:
		prefix = "basic "
	case legendary:
		prefix = "legendary "
	default:
	}
	// A required card-type union ("instant or sorcery", "artifact or
	// enchantment") reconstructs from one word per type. A single required card
	// type also takes this path when the selection is a plain card kind
	// (SelectionCard), modeling instant- and sorcery-card tutors ("a sorcery
	// card", "an instant card") whose type has no dedicated SelectionKind; the
	// compiler keeps that single type for SelectionCard so the lowered spec
	// preserves it. Typed card kinds (creature, artifact) keep their single type
	// in Kind and reconstruct through searchFilterTypeNoun below.
	if len(sel.SubtypesAny) == 0 && !basic && !legendary && colorStr == "" && sel.Kind != SelectionSpell &&
		(len(sel.RequiredTypesAny) > 1 ||
			(len(sel.RequiredTypesAny) == 1 && sel.Kind == SelectionCard)) {
		words := make([]string, 0, len(sel.RequiredTypesAny))
		for _, cardType := range sel.RequiredTypesAny {
			word, ok := searchFilterCardTypeWord(cardType)
			if !ok {
				return "", false
			}
			words = append(words, word)
		}
		return joinOrList(words), true
	}
	base, ok := searchFilterTypeNoun(sel.Kind)
	if !ok {
		return "", false
	}

	if len(sel.SubtypesAny) > 0 {
		if colorStr != "" {
			return "", false
		}
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
		case SelectionCreature, SelectionArtifact, SelectionEnchantment, SelectionLand, SelectionPermanent:
			// A subtype paired with a card type or "permanent" ("Myr creature",
			// "Dragon creature", "Rebel permanent"): the runtime matches by both
			// the type (or permanent-ness) and the subtype. "basic" pairs only
			// with a bare land, never with a typed subtype; "legendary" may prefix
			// the union.
			if basic {
				return "", false
			}
			return prefix + subtypes + " " + base, true
		default:
			return "", false
		}
	}
	if basic && base != "land" {
		// "basic" without a subtype is meaningful only for "basic land".
		return "", false
	}
	if colorStr != "" {
		if base == "" {
			return prefix + colorStr, true
		}
		return prefix + colorStr + " " + base, true
	}
	return prefix + base, true
}

func searchFilterCardTypeWord(cardType CardType) (string, bool) {
	switch cardType {
	case CardTypeLand:
		return "land", true
	case CardTypeCreature:
		return "creature", true
	case CardTypeArtifact:
		return "artifact", true
	case CardTypeEnchantment:
		return "enchantment", true
	case CardTypePlaneswalker:
		return "planeswalker", true
	case CardTypeInstant:
		return "instant", true
	case CardTypeSorcery:
		return "sorcery", true
	default:
		return "", false
	}
}

// searchFilterTypeNoun maps a selection kind to the printed card-type noun a
// search filter uses, returning ok=false for kinds the runtime SearchSpec cannot
// express. A plain card kind has an empty noun; an instant- or sorcery-card
// filter carries its type as a single RequiredTypesAny entry on SelectionCard
// and reconstructs through the card-type-word path in canonicalSearchFilter
// rather than here.
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
	case SelectionPermanent:
		return "permanent", true
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
		"put it into your graveyard, then shuffle.",
		"put that card into your graveyard, then shuffle.",
		"reveal it, put it into your graveyard, then shuffle.",
		"reveal that card, put it into your graveyard, then shuffle.",
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
		"put them into your graveyard, then shuffle.",
		"put those cards into your graveyard, then shuffle.",
		"reveal them, put them into your graveyard, then shuffle.",
		"reveal those cards, put them into your graveyard, then shuffle.",
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

func searchTopDestinationSupported(destination string) bool {
	// "put <object> on top" sends the found card to the top of the library after
	// shuffling. The found card is named by an interchangeable demonstrative
	// ("it", "that card", or "the card"), optionally after a reveal that names it
	// the same way. Every combination denotes the same put-on-top destination.
	objects := []string{"it", "that card", "the card"}
	for _, put := range objects {
		if destination == "then shuffle and put "+put+" on top." {
			return true
		}
		for _, revealed := range objects {
			if destination == "reveal "+revealed+", then shuffle and put "+put+" on top." {
				return true
			}
		}
	}
	return false
}

// searchSplitDestinationSupported reports whether the clause tail is one of the
// split-destination wordings the runtime models: the two found cards are
// revealed (optionally) and distributed across two single-card slots, "put one
// <slot> and the other <slot>", where each slot is a hand or battlefield
// (optionally tapped) destination. It models Cultivate and Kodama's Reach. The
// typed slot assignment is carried separately on the EffectPut clause's
// SearchSplit field (parseSearchSplitPut); this gate only confirms the byte-exact
// envelope so lowering may consume those typed fields.
func searchSplitDestinationSupported(destination string) bool {
	const slotA = "one onto the battlefield tapped"
	const slotB = "the other into your hand"
	bodies := []string{
		"put " + slotA + " and " + slotB,
		"put one into your hand and the other onto the battlefield tapped",
		"put one onto the battlefield and the other into your hand",
		"put one into your hand and the other onto the battlefield",
	}
	reveals := []string{
		"",
		"reveal those cards, ",
		"reveal them, ",
	}
	for _, reveal := range reveals {
		for _, body := range bodies {
			if destination == reveal+body+", then shuffle." {
				return true
			}
		}
	}
	return false
}

// exactLoseGameEffectSyntax reconstructs the byte-exact "<subject> lose(s) the
// game." clause for each supported losing player: the controller ("You lose the
// game."), the referenced triggering player ("They lose the game.", "That player
// loses the game."), and a single exact target ("Target player loses the
// game."). The subject phrasing mirrors exactLifeEffectSyntax so a granted "...
// that player loses the game." trigger and a targeted lose-game spell both
// reconstruct without the compiler inspecting wording.
func exactLoseGameEffectSyntax(effect *EffectSyntax) bool {
	var prefixes []string
	switch effect.Context {
	case EffectContextController:
		prefixes = []string{"You lose"}
	case EffectContextEventPlayer, EffectContextReferencedPlayer:
		prefixes = []string{"They lose", "That player loses"}
	case EffectContextTarget, EffectContextPriorSubject:
		if len(effect.Targets) == 1 && effect.Targets[0].Exact {
			prefixes = []string{titleFirstEffectText(effect.Targets[0].Text) + " loses"}
		}
	default:
	}
	text := exactEffectClauseText(effect)
	for _, prefix := range prefixes {
		if strings.EqualFold(text, prefix+" the game.") {
			return true
		}
	}
	return false
}

// exactGainGrantedAbilityEffectSyntax reconstructs the byte-exact "<subject>
// gains." head of a resolving ability grant ("This creature gains \"Whenever
// this creature deals combat damage to a player, that player loses the
// game.\""). The quoted body is stripped from the clause text to a bare "gains",
// mirroring how a token's quoted rider strips to a bare "with"; its tokens are
// recognized and covered through the parsed granted ability's own inner
// document, so the compiler lowers the conferred ability without inspecting
// wording. It applies only once attachGainGrantedAbilities has bound the quoted
// ability.
// exactCreateEmblemEffectSyntax reconstructs an "You get an emblem with
// \"...\"" clause from its typed emblem abilities. The quoted ability text is
// recognized and covered through each parsed inner document, so the clause
// strips to a bare "You get an emblem with" the way a gain rider strips to
// "gains". It applies only once attachEmblemEffects has bound the abilities.
func exactCreateEmblemEffectSyntax(effect *EffectSyntax) bool {
	if len(effect.EmblemAbilities) == 0 {
		return false
	}
	return strings.EqualFold(exactEffectClauseText(effect), "You get an emblem with.")
}

func exactGainGrantedAbilityEffectSyntax(effect *EffectSyntax) bool {
	if effect.GainGrantedAbility == nil {
		return false
	}
	verb := slices.IndexFunc(effect.Tokens, func(token shared.Token) bool {
		return token.Span == effect.VerbSpan
	})
	if verb < 0 || !equalWord(effect.Tokens[verb], "gains") {
		return false
	}
	subjectStart := effectSubjectStart(effect.Tokens, verb, effectSelfNameSpans(effect))
	subjectTokens := effect.Tokens[subjectStart:verb]
	if len(subjectTokens) == 0 {
		return false
	}
	subject := joinedEffectText(subjectTokens)
	expected := subject + " gains."
	return strings.EqualFold(exactEffectClauseText(effect), expected)
}

// cardsNamedSelfInGraveyardsAmount reports whether amount is a self-named
// graveyard count ("for each card named <this card> in [each|your] graveyard"),
// across either every graveyard or only the controller's. The count is
// self-contained per clause (it reads the source card's own name at resolution,
// not a shared X), so an elided-subject clause carrying it reconstructs
// faithfully in isolation.
func cardsNamedSelfInGraveyardsAmount(amount EffectAmountSyntax) bool {
	return amount.DynamicKind == EffectDynamicAmountCardsNamedSelfInGraveyards ||
		amount.DynamicKind == EffectDynamicAmountCardsNamedSelfInControllerGraveyard
}

func exactLifeEffectSyntax(effect *EffectSyntax, controllerVerb, subjectVerb string) bool {
	var prefixes []string
	switch effect.Context {
	case EffectContextController:
		prefixes = []string{"You " + controllerVerb, titleFirstEffectText(controllerVerb)}
	case EffectContextEachOpponent:
		prefixes = []string{"Each opponent " + subjectVerb}
	case EffectContextEachOtherPlayer:
		prefixes = []string{"Each other player " + subjectVerb}
	case EffectContextEachPlayer:
		prefixes = []string{"Each player " + subjectVerb}
	case EffectContextDefendingPlayer:
		prefixes = []string{"Defending player " + subjectVerb}
	case EffectContextTarget, EffectContextPriorSubject:
		if len(effect.Targets) == 1 && effect.Targets[0].Exact {
			prefixes = []string{titleFirstEffectText(effect.Targets[0].Text) + " " + subjectVerb}
		} else if effect.Context == EffectContextPriorSubject && len(effect.Targets) == 0 &&
			(effect.Amount.DynamicForm == EffectDynamicAmountFormNone ||
				cardsNamedSelfInGraveyardsAmount(effect.Amount)) {
			// The subject is elided: it is inherited from the prior effect in a
			// compound sentence ("Target player draws two cards and loses 2
			// life"). The clause reconstructs from the bare third-person verb,
			// matching how exactDamageEffectSyntax handles a prior-subject
			// damage clause with no own subject tokens. Restricted to a
			// self-contained amount (a fixed value, the spell's cost X, or a
			// self-named graveyard count): a trailing "where X is ..." amount
			// form defines a single X shared by every effect in the sentence,
			// but the parser binds that clause to only one effect, so
			// reconstructing the elided-subject clause in isolation would not
			// faithfully model the shared amount. A "for each card named <this
			// card> in [each|your] graveyard" count, by contrast, is independent
			// per clause ("Target player gains 4 life, then gains 4 life for each
			// card named Life Burst in each graveyard", Life Burst).
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
	if effect.Duration == EffectDurationUntilYourNextTurn &&
		effect.Context == EffectContextController {
		return strings.EqualFold(
			exactEffectClauseText(effect),
			"You gain protection from everything until your next turn.",
		)
	}
	return exactTemporaryKeywordChangeSyntax(effect, "gain", "gains", true) ||
		exactPermanentKeywordGrantEffectSyntax(effect)
}

// exactPermanentKeywordGrantEffectSyntax recognizes a resolving keyword grant
// with no duration to the object an earlier clause acted on ("It gains haste."),
// the temporary-reanimation haste rider on "Return target creature card ... It
// gains haste. Exile it at the beginning of the next end step." (Whip of Erebos
// and the many graveyard-return cards that grant the reanimated creature haste).
// The grant carries no "until end of turn" suffix, so it persists while the
// object remains on the battlefield; a sibling cleanup clause exiles it at the
// next end step. Only the singular referenced-object back-reference ("it" / "that
// creature") is matched, so a static group anthem ("Creatures you control gain
// trample.") never reaches this no-duration path.
func exactPermanentKeywordGrantEffectSyntax(effect *EffectSyntax) bool {
	if effect.Duration != "" || effect.Context != EffectContextReferencedObject {
		return false
	}
	subject, ok := exactObjectReferenceText(effect.SubjectReferences)
	if !ok {
		return false
	}
	text := strings.ToLower(exactEffectClauseText(effect))
	middle, ok := strings.CutPrefix(text, strings.ToLower(subject)+" gains ")
	if !ok {
		return false
	}
	body, ok := strings.CutSuffix(middle, ".")
	return ok && body != "" && (exactTemporaryKeywordList(body) || exactKeywordChoiceList(body))
}

// keywordGrantIsChoice reports whether a gain effect grants a disjunctive
// keyword choice ("gains banding, first strike, or trample") rather than a
// conjunctive list. It extracts the keyword body after the "gains " verb (for a
// direct target, a referenced object, the source, or a prior subject) and tests
// it against the disjunctive keyword-choice grammar. It returns false for any
// effect that is not a recognized keyword-choice grant.
func keywordGrantIsChoice(effect *EffectSyntax) bool {
	if effect.Kind != EffectGain {
		return false
	}
	text := strings.ToLower(exactEffectClauseText(effect))
	_, after, ok := strings.Cut(text, " gains ")
	if !ok {
		// A prior-subject grant reads "gains <keywords> ..." with no subject noun.
		after, ok = strings.CutPrefix(text, "gains ")
		if !ok {
			return false
		}
	}
	body, ok := strings.CutSuffix(after, " until end of turn.")
	if !ok {
		body, ok = strings.CutSuffix(after, ".")
		if !ok {
			return false
		}
	}
	return exactKeywordChoiceList(body)
}

// exactBackReferenceTargetKeywordGrantEffectSyntax recognizes a resolving
// keyword grant with no duration whose subject is a back-reference to the
// ability's target ("... or that creature gains banding, first strike, or
// trample."). The compiler binds "that creature" to the target, so the effect
// carries EffectContextTarget with a subject reference and no target noun phrase
// of its own (the targeted noun is owned by the sibling alternative clause).
// Both a conjunctive keyword list (grant all) and a disjunctive keyword choice
// (grant one chosen at resolution) are recognized; the lowering distinguishes
// them from the connective recorded on the effect. The own-target form ("Target
// creature gains ...") is handled by exactDirectTargetKeywordGrantEffectSyntax.
func exactBackReferenceTargetKeywordGrantEffectSyntax(effect *EffectSyntax) bool {
	if effect.Duration != "" || effect.Context != EffectContextTarget || len(effect.Targets) != 0 {
		return false
	}
	subject, ok := exactObjectReferenceText(effect.SubjectReferences)
	if !ok {
		return false
	}
	text := strings.ToLower(exactEffectClauseText(effect))
	middle, ok := strings.CutPrefix(text, strings.ToLower(subject)+" gains ")
	if !ok {
		return false
	}
	body, ok := strings.CutSuffix(middle, ".")
	return ok && body != "" && (exactTemporaryKeywordList(body) || exactKeywordChoiceList(body))
}

// exactDirectTargetKeywordGrantEffectSyntax recognizes a resolving keyword grant
// with no duration to a single targeted creature ("Target creature gains first
// strike.", "Target creature gains banding, first strike, or trample."). The
// grant carries no "until end of turn" suffix, so it persists indefinitely while
// the creature remains on the battlefield (the runtime applies it with
// DurationPermanent). Both a conjunctive keyword list ("first strike and
// trample", grant all) and a disjunctive keyword choice ("banding, first strike,
// or trample", grant one chosen at resolution) are recognized; the lowering
// distinguishes them from the connective recorded on the effect.
func exactDirectTargetKeywordGrantEffectSyntax(effect *EffectSyntax) bool {
	if effect.Duration != "" || effect.Context != EffectContextTarget {
		return false
	}
	if len(effect.Targets) != 1 || !effect.Targets[0].Exact {
		return false
	}
	text := strings.ToLower(exactEffectClauseText(effect))
	middle, ok := strings.CutPrefix(text, strings.ToLower(effect.Targets[0].Text)+" gains ")
	if !ok {
		return false
	}
	body, ok := strings.CutSuffix(middle, ".")
	if !ok || body == "" {
		return false
	}
	return exactTemporaryKeywordList(body) || exactKeywordChoiceList(body)
}

// exactControlledSourceKeywordGrantEffectSyntax recognizes a resolving keyword
// grant to a single targeted permanent that lasts as long as the source remains
// under its controller's control ("Target creature you control gains
// indestructible for as long as you control this Saga.", Tale of Tinúviel
// chapter I). It mirrors exactDirectTargetKeywordGrantEffectSyntax for the
// no-duration grant but matches the "for as long as you control this <noun>"
// duration suffix instead of a bare period, so the grant reconstructs
// byte-exactly. Both a conjunctive keyword list and a disjunctive keyword choice
// are recognized.
func exactControlledSourceKeywordGrantEffectSyntax(effect *EffectSyntax) bool {
	if effect.Context != EffectContextTarget ||
		len(effect.Targets) != 1 || !effect.Targets[0].Exact {
		return false
	}
	if effect.Duration != EffectDurationWhileYouControlSource {
		return false
	}
	text := strings.ToLower(exactEffectClauseText(effect))
	middle, ok := strings.CutPrefix(text, strings.ToLower(effect.Targets[0].Text)+" gains ")
	if !ok {
		return false
	}
	return exactKeywordControlledSourceDurationBody(middle)
}

// exactKeywordControlledSourceDurationBody validates the keyword list preceding a
// "for as long as you control this <noun>." suffix on a keyword grant, mirroring
// exactGainControlControlledSourceDuration's named-source acceptance.
func exactKeywordControlledSourceDurationBody(middle string) bool {
	const suffix = " for as long as you control this "
	index := strings.Index(middle, suffix)
	if index <= 0 {
		return false
	}
	noun, ok := strings.CutSuffix(middle[index+len(suffix):], ".")
	if !ok || noun == "" || strings.ContainsRune(noun, ' ') {
		return false
	}
	body := middle[:index]
	return exactTemporaryKeywordList(body) || exactKeywordChoiceList(body)
}

// until end of turn ("Permanents your opponents control lose hexproof and
// indestructible until end of turn.", "Target creature loses flying until end of
// turn."). It mirrors exactTemporaryKeywordEffectSyntax with the "lose"/"loses"
// verbs, so a removal reconstructs byte-exactly for the same subject shapes a
// keyword grant supports.
func exactTemporaryKeywordLossEffectSyntax(effect *EffectSyntax) bool {
	return exactTemporaryKeywordChangeSyntax(effect, "lose", "loses", false)
}

// exactTemporaryKeywordChangeSyntax reconstructs the byte-exact form of a
// resolving until-end-of-turn keyword change clause for the supplied plural verb
// ("gain"/"lose") and singular verb ("gains"/"loses"). It covers every affected
// subject shape: a never-resolving creature or permanent group, a prior subject
// ("it"/"creatures you control"), a referenced object, the source permanent, and
// a single exact target (including the combined "<target> gets +N/+N and
// gains/loses ..." pump form).
//
// allowChoice permits a disjunctive "your choice of <list>" keyword body on the
// single-subject shapes (source, referenced object, and one exact target) the
// lowering can realize as a one-of-N modal grant. The grant verb passes true; the
// loss verb passes false, since a keyword-loss choice is not lowered.
func exactTemporaryKeywordChangeSyntax(effect *EffectSyntax, pluralVerb, singularVerb string, allowChoice bool) bool {
	validBody := func(body string) bool {
		return exactTemporaryKeywordList(body) || (allowChoice && exactKeywordChoiceList(body))
	}
	if effect.Duration != EffectDurationUntilEndOfTurn {
		return false
	}
	text := strings.ToLower(exactEffectClauseText(effect))
	if effect.StaticSubject.Kind != EffectStaticSubjectNone {
		return exactGroupTemporaryKeywordEffectSyntax(effect, text, pluralVerb, singularVerb)
	}
	if effect.Context == EffectContextPriorSubject {
		// A singular prior subject ("it") reads "<singularVerb> <kw> …"; a plural
		// group prior subject ("creatures you control") reads "<pluralVerb> <kw>
		// …".
		middle, ok := strings.CutPrefix(text, singularVerb+" ")
		if !ok {
			middle, ok = strings.CutPrefix(text, pluralVerb+" ")
		}
		if !ok {
			return false
		}
		middle, ok = cutTemporaryKeywordDurationSuffix(effect, middle)
		return ok && exactTemporaryKeywordList(middle)
	}
	if effect.Context == EffectContextReferencedObject {
		subject, ok := exactObjectReferenceText(effect.SubjectReferences)
		if !ok {
			return false
		}
		middle, ok := strings.CutPrefix(text, strings.ToLower(subject)+" "+singularVerb+" ")
		if !ok {
			return false
		}
		middle, ok = cutTemporaryKeywordDurationSuffix(effect, middle)
		return ok && validBody(middle)
	}
	if effect.Context == EffectContextSource {
		subject, ok := exactSelfSubjectReferenceText(effect.SubjectReferences)
		if !ok {
			return false
		}
		middle, ok := strings.CutPrefix(text, strings.ToLower(subject)+" "+singularVerb+" ")
		if !ok {
			return false
		}
		middle, ok = cutTemporaryKeywordDurationSuffix(effect, middle)
		return ok && validBody(middle)
	}
	// "Those creatures gain <keyword> until end of turn." grants a keyword to a
	// group named by the demonstrative back-reference "those" (Inspiring Call).
	// The group noun is reconstructed from the preceding count clause at
	// lowering, so only the demonstrative subject and the keyword tail are
	// validated here; the whole clause is still consumed byte-exactly.
	if exactThoseSubjectReference(effect.SubjectReferences) {
		body, ok := strings.CutPrefix(text, "those ")
		if !ok {
			return false
		}
		noun, keywords, ok := strings.Cut(body, " gain ")
		if !ok || noun == "" {
			return false
		}
		keywords, ok = cutTemporaryKeywordDurationSuffix(effect, keywords)
		return ok && exactTemporaryKeywordList(keywords)
	}
	if len(effect.Targets) != 1 || !effect.Targets[0].Exact {
		return false
	}
	// A plural ("two target creatures") or optional-multi ("up to two target
	// creatures") target distributes the change with "each <pluralVerb>": "Up to
	// two target creatures each gain lifelink until end of turn." and the
	// combined "Up to two target creatures each get +N/+N and gain trample until
	// end of turn." pump form. The singular "<target> <singularVerb>" path below
	// owns the one-target cardinality.
	if effect.Targets[0].Cardinality.Max >= 2 {
		if prefix, suffix, ok := strings.Cut(text, " and "+pluralVerb+" "); ok &&
			strings.HasPrefix(prefix, strings.ToLower(effect.Targets[0].Text)+" each get ") {
			middle, suffixOK := cutTemporaryKeywordDurationSuffix(effect, suffix)
			return suffixOK && exactTemporaryKeywordList(middle)
		}
		eachMiddle, ok := strings.CutPrefix(text, strings.ToLower(effect.Targets[0].Text)+" each "+pluralVerb+" ")
		if !ok {
			return false
		}
		body, suffixOK := cutTemporaryKeywordDurationSuffix(effect, eachMiddle)
		return suffixOK && body != "" && exactTemporaryKeywordList(body)
	}
	if prefix, suffix, ok := strings.Cut(text, " and "+singularVerb+" "); ok &&
		strings.HasPrefix(prefix, strings.ToLower(effect.Targets[0].Text)+" gets ") {
		middle, suffixOK := cutTemporaryKeywordDurationSuffix(effect, suffix)
		return suffixOK && exactTemporaryKeywordList(middle)
	}
	prefix := strings.ToLower(effect.Targets[0].Text) + " " + singularVerb + " "
	middle, ok := strings.CutPrefix(text, prefix)
	if !ok {
		return false
	}
	middle, ok = cutTemporaryKeywordDurationSuffix(effect, middle)
	if !ok || middle == "" {
		return false
	}
	return validBody(middle)
}

// cutTemporaryKeywordDurationSuffix strips the until-end-of-turn duration suffix
// from a keyword-change clause body. A plain grant ends "… until end of turn.";
// a grant that shares a "+X/+X … where X is …" pump's amount carries the trailing
// "where X is …" clause bound to this keyword effect (The Weatherseed Treaty
// chapter III: "gains trample until end of turn, where X is the number of basic
// land types …"), so the suffix becomes "… until end of turn, <amount>.". The
// combined +N/+N-and-gain lowering reads that shared amount to resolve the pump's
// X; here it is consumed so the clause reconstructs byte-exactly.
//
// A sentence-leading "Until end of turn," supplies the duration before the
// subject ("Until end of turn, target creature gains trample."), so the clause
// body itself carries no trailing suffix and ends with a bare period. The leading
// duration is consumed by the effect's recorded Duration, so accept the bare
// "<body>." form only when that duration is until end of turn, mirroring the
// power/toughness reconstruction in exactGroupModifyPTBody. A static no-duration
// grant (Duration "") never reaches this bare branch, so an indefinite anthem is
// still rejected here.
func cutTemporaryKeywordDurationSuffix(effect *EffectSyntax, body string) (string, bool) {
	if effect.Amount.DynamicForm == EffectDynamicAmountFormWhereX && effect.Amount.Text != "" {
		suffix := " until end of turn, " + strings.ToLower(effect.Amount.Text) + "."
		if middle, ok := strings.CutSuffix(body, suffix); ok {
			return middle, true
		}
	}
	if middle, ok := strings.CutSuffix(body, " until end of turn."); ok {
		return middle, true
	}
	if effect.Duration == EffectDurationUntilEndOfTurn {
		return strings.CutSuffix(body, ".")
	}
	return "", false
}

// exactGroupTemporaryKeywordEffectSyntax recognizes a resolving keyword grant to
// a never-resolving creature or permanent group until end of turn ("Creatures
// you control gain trample until end of turn."). The subject is reconstructed
// byte-exactly from the tokens covered by the static-subject span, mirroring
// exactGroupModifyPTEffectSyntax. text is the lowercased clause text.
func exactGroupTemporaryKeywordEffectSyntax(effect *EffectSyntax, text, pluralVerb, singularVerb string) bool {
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
	// A plural group reads "<pluralVerb>"; the singular "each <permanent>" form
	// reads "<singularVerb>". Try both so the reconstruction stays byte-exact
	// with the source.
	for _, verb := range []string{" " + pluralVerb + " ", " " + singularVerb + " "} {
		middle, ok := strings.CutPrefix(text, subjectText+verb)
		if !ok {
			continue
		}
		if body, ok := strings.CutSuffix(middle, " until end of turn."); ok && body != "" && exactTemporaryKeywordList(body) {
			return true
		}
		// A keyword-first mass pump ("Creatures you control gain trample and get
		// +X/+X until end of turn, …") splits the keyword grant off without its
		// own duration suffix; the until-end-of-turn duration is spread onto this
		// effect and the following modify clause carries the suffix. Accept the
		// bare "<subject> gain <keywords>." form only when the duration was
		// recognized so a static anthem (no duration) never matches.
		if effect.Duration == EffectDurationUntilEndOfTurn {
			if body, ok := strings.CutSuffix(middle, "."); ok && body != "" && exactTemporaryKeywordList(body) {
				return true
			}
		}
	}
	return false
}

// exactCanAttackAsThoughDefenderEffectSyntax recognizes the temporary combat
// permission "<source> can attack this turn as though it didn't have defender."
// that lets the source creature attack despite its defender keyword until end of
// turn. Only the source subject is recognized ("This creature ..." or the card's
// own name), an activated or triggered self grant (Krotiq Nestguard, Skyclave
// Squid, Returned Phalanx). The clause is reconstructed byte-exactly from the
// source subject text so every deviation fails closed: a different subject, a
// different duration (the trailing "this turn" is fixed), or any added rider
// leaves the clause non-exact.
func exactCanAttackAsThoughDefenderEffectSyntax(effect *EffectSyntax) bool {
	if effect.Duration != EffectDurationThisTurn {
		return false
	}
	if effect.Context != EffectContextSource && effect.Context != EffectContextPriorSubject {
		return false
	}
	clause := exactEffectClauseText(effect)
	if effect.Context == EffectContextPriorSubject {
		return strings.EqualFold(clause, "can attack this turn as though it didn't have defender.")
	}
	subject, ok := exactSelfSubjectReferenceText(effect.SubjectReferences)
	return ok && strings.EqualFold(clause, subject+" can attack this turn as though it didn't have defender.")
}

// exactCantBeBlockedEffectSyntax recognizes the temporary combat-evasion
// resolving effect "<subject> can't be blocked this turn." that grants the
// subject creature(s) a "can't be blocked" restriction until end of turn. Three
// subject shapes are recognized, each reconstructed byte-exactly so every
// deviation fails closed:
//
//   - a target noun phrase ("Target creature can't be blocked this turn.",
//     "Up to one target creature can't be blocked this turn.") with single,
//     plural, or optional cardinality, like the sibling can't-block recognizer;
//   - the source itself ("This creature can't be blocked this turn." / the
//     card's own name), a self grant on an activated ability; and
//   - a prior-subject sequence clause ("... and can't be blocked this turn.")
//     whose subject noun is inherited from the preceding clause and so carries
//     only the bare predicate here.
//
// Every other deviation leaves the clause non-exact: a different duration (the
// trailing "this turn" is fixed), a "can't be blocked except by ..." qualifier,
// a "by more than one creature" rider, or any "can't block" / "can't attack"
// wording.
func exactCantBeBlockedEffectSyntax(effect *EffectSyntax) bool {
	if effect.Duration != EffectDurationThisTurn {
		return false
	}
	clause := exactEffectClauseText(effect)
	switch effect.Context {
	case EffectContextTarget:
		return len(effect.Targets) == 1 &&
			effect.Targets[0].Exact &&
			effect.Targets[0].Cardinality.Min >= 0 &&
			effect.Targets[0].Cardinality.Max >= 1 &&
			effect.Targets[0].Cardinality.Min <= effect.Targets[0].Cardinality.Max &&
			strings.EqualFold(clause, effect.Targets[0].Text+" can't be blocked this turn.")
	case EffectContextSource:
		subject, ok := exactSelfSubjectReferenceText(effect.SubjectReferences)
		return ok && strings.EqualFold(clause, subject+" can't be blocked this turn.")
	case EffectContextPriorSubject:
		return strings.EqualFold(clause, "can't be blocked this turn.")
	default:
		return false
	}
}

// exactCantBlockEffectSyntax recognizes the temporary combat-restriction
// resolving effect "<targets> can't block this turn." that grants the targeted
// creature(s) a "can't block" restriction until end of turn. Unlike the
// single-target can't-be-blocked recognizer it accepts the multi-target and
// optional cardinalities ("Up to three target creatures can't block this turn.",
// "One or two target creatures can't block this turn.") the lowering applies
// per target. The clause is reconstructed byte-exactly from the target's own
// text so every deviation fails closed: a different duration (the trailing "this
// turn" is fixed), a "can't block creatures you control" protected-object
// qualifier, an "except" rider, or any "can't be blocked" / "can't attack"
// wording leaves the clause non-exact.
func exactCantBlockEffectSyntax(effect *EffectSyntax) bool {
	return effect.Duration == EffectDurationThisTurn &&
		effect.Context == EffectContextTarget &&
		len(effect.Targets) == 1 &&
		effect.Targets[0].Exact &&
		effect.Targets[0].Cardinality.Min >= 0 &&
		effect.Targets[0].Cardinality.Max >= 1 &&
		effect.Targets[0].Cardinality.Min <= effect.Targets[0].Cardinality.Max &&
		strings.EqualFold(
			exactEffectClauseText(effect),
			effect.Targets[0].Text+" can't block this turn.",
		)
}

// exactCantAttackEffectSyntax recognizes the temporary combat-restriction
// resolving effect "<targets> can't attack this turn." that grants the targeted
// creature(s) a "can't attack" restriction until end of turn. It mirrors the
// sibling can't-block recognizer, accepting the single-target and optional
// multi-target cardinalities ("Up to two target creatures can't attack this
// turn.") the lowering applies per target. The clause is reconstructed
// byte-exactly from the target's own text, so every deviation fails closed (a
// different duration, a "can't attack you"/"can't attack unless ..." rider, or
// the combined "can't attack or block this turn." form recognized separately).
func exactCantAttackEffectSyntax(effect *EffectSyntax) bool {
	return exactTargetCombatClause(effect, "can't attack this turn.")
}

// exactCantAttackOrBlockEffectSyntax recognizes the combined temporary
// combat-restriction resolving effect "<targets> can't attack or block this
// turn." (Thundersong Trumpeter, Off Balance), which grants the targeted
// creature(s) both a "can't attack" and a "can't block" restriction until end of
// turn. The clause is reconstructed byte-exactly from the target's own text, so
// any other duration, rider, or qualifier fails closed.
func exactCantAttackOrBlockEffectSyntax(effect *EffectSyntax) bool {
	return exactTargetCombatClause(effect, "can't attack or block this turn.")
}

// exactTargetMustAttackEffectSyntax recognizes the temporary single-target
// forced-attack resolving effect "<target> attacks this turn if able." (Kookus,
// Norritt), which forces the targeted creature to attack this turn. The clause
// text reconstruction strips the trailing "if able" qualifier, so the predicate
// matched here is "attacks this turn."; the verb recognizer already required the
// full "attacks this turn if able" wording before this kind is assigned. The
// group "<group> attack this turn if able." form (owned by the dedicated group
// recognizer) and the directed "attacks <player> this turn if able" form both
// fail closed because they never reach this single-target recognizer.
func exactTargetMustAttackEffectSyntax(effect *EffectSyntax) bool {
	return effect.Duration == EffectDurationThisTurn &&
		effect.Context == EffectContextTarget &&
		len(effect.Targets) == 1 &&
		effect.Targets[0].Exact &&
		effect.Targets[0].Cardinality.Min == 1 &&
		effect.Targets[0].Cardinality.Max == 1 &&
		strings.EqualFold(
			exactEffectClauseText(effect),
			effect.Targets[0].Text+" attacks this turn.",
		)
}

// exactTargetCombatClause reports whether effect is an exact, this-turn,
// creature-target combat clause whose reconstructed text is exactly the
// target's own text followed by predicate. It backs the temporary
// combat-restriction recognizers (can't attack, can't attack or block) that
// share the same this-turn, byte-exact reconstruction shape and the optional
// multi-target cardinalities applied per target by lowering.
func exactTargetCombatClause(effect *EffectSyntax, predicate string) bool {
	return effect.Duration == EffectDurationThisTurn &&
		effect.Context == EffectContextTarget &&
		len(effect.Targets) == 1 &&
		effect.Targets[0].Exact &&
		effect.Targets[0].Cardinality.Min >= 0 &&
		effect.Targets[0].Cardinality.Max >= 1 &&
		effect.Targets[0].Cardinality.Min <= effect.Targets[0].Cardinality.Max &&
		strings.EqualFold(
			exactEffectClauseText(effect),
			effect.Targets[0].Text+" "+predicate,
		)
}

func exactTemporaryKeywordList(text string) bool {
	items := splitGrantKeywordItems(text)
	for _, keyword := range items {
		if !grantableKeywordWord(keyword) {
			return false
		}
	}
	return len(items) > 0
}

// exactKeywordChoiceList recognizes a disjunctive list of two or more grantable
// keywords joined by "or" ("first strike or trample", "banding, first strike, or
// trample"). It also accepts the explicit "your choice of <list>" header that
// modern templating prefixes to the same disjunction ("your choice of vigilance,
// lifelink, or haste"). The disjunction means the controller chooses exactly one
// of the listed keywords at resolution, distinct from the conjunctive list
// recognized by exactTemporaryKeywordList where every keyword is granted. It
// requires at least one "or" connective so a single keyword or an "and" list
// never matches here.
func exactKeywordChoiceList(text string) bool {
	text = strings.ToLower(text)
	text = strings.TrimPrefix(text, "your choice of ")
	if !strings.Contains(text, " or ") {
		return false
	}
	text = strings.ReplaceAll(text, ", or ", ", ")
	text = strings.ReplaceAll(text, " or ", ", ")
	count := 0
	for keyword := range strings.SplitSeq(text, ", ") {
		if !grantableKeywordWord(keyword) {
			return false
		}
		count++
	}
	return count >= 2
}

// grantableKeywordWord reports whether a lowercase Oracle phrase names a
// non-parameterized keyword (or a fully-specified protection variant) the
// executable backend can grant. Protection phrases are validated structurally by
// grantableProtectionPhrase so every protected predicate the keyword parser can
// recognize and the lowering can reduce to a static mechanic — a color list, the
// each-color/everything/monocolored/multicolored/chosen-color quantifiers, a
// card-type list, or a creature/land subtype list — is grantable.
func grantableKeywordWord(keyword string) bool {
	switch keyword {
	case "deathtouch", "double strike", "fear", "first strike", "flying", "haste",
		"banding", "hexproof", "indestructible", "intimidate", "lifelink", "menace", "reach", "shadow", "shroud", "trample", "vigilance",
		"horsemanship", "infect", "skulk", "wither",
		"landwalk", "plainswalk", "islandwalk", "swampwalk", "mountainwalk", "forestwalk", "desertwalk", "nonbasic landwalk":
		return true
	default:
		return grantableProtectionPhrase(keyword)
	}
}

// exactCreateTokenEffectSyntax recognizes vanilla creature-token creation:
// "Create <count> [tapped] <P>/<T> [colorless | <colors>] <Subtypes> [artifact |
// enchantment] creature token[s] [with <keyword>] [named <Name>] [that's/that are
// [tapped and] attacking]." with a fixed power/toughness, up to two colors (or
// colorless), one or two creature subtypes, an optional leading
// artifact/enchantment permanent type, an optional "tapped" entry modifier, an
// optional single creature keyword, an optional explicit Oracle name ("... named
// <Name>"), and an optional trailing attacking-entry clause (CR 508.4). The token
// count may be a fixed number, the spell's variable X, a "for each <iterator>"
// per-object count (in either leading or trailing position), a "number of ...
// equal to <dynamic>" count, or a "where X is <dynamic>" count. It fails closed
// for every richer shape (a "blocking" entry, quoted abilities, multiple
// keywords, modifiers, ...); a name followed by a quoted granted-ability rider
// ("... named X with \"...\"") fails closed via parseTokenName. The recipient may
// be the spell's controller ("Create ..."), a referenced object's controller
// ("Its controller creates ..."), or a single targeted player ("Target opponent
// creates ...", "Target player creates ..."); the targeted-player form accepts
// fixed counts only.
// exactCreateTokenRecipientContext validates the create-token effect's recipient
// context. It returns targetRecipient=true for the "Target opponent/player
// creates ..." form, and ok=false when the context (or its target shape) is not
// a supported recipient. A targeted-player recipient requires exactly one exact
// player-or-opponent target; the controller and referenced-object-controller
// forms must carry no target.
func exactCreateTokenRecipientContext(effect *EffectSyntax) (targetRecipient, ok bool) {
	targetRecipient = effect.Context == EffectContextTarget
	if effect.Context != EffectContextController &&
		effect.Context != EffectContextReferencedObjectController &&
		!targetRecipient {
		return false, false
	}
	if targetRecipient {
		if len(effect.Targets) != 1 ||
			!effect.Targets[0].Exact ||
			!exactCardCountTargetPlayer(effect.Targets[0].Selection) {
			return false, false
		}
	} else if len(effect.Targets) != 0 {
		return false, false
	}
	return targetRecipient, true
}

// creatureTokenSpecBody validates a creature-token effect's selection and, on
// success, returns a builder that renders the canonical token spec body for a
// given count word and noun ("a"/"token", "X"/"tokens", ...). It returns ok=false
// for any selection a vanilla creature token cannot represent. The builder folds
// in the leading "tapped" adjective, color words, subtypes, permanent-type words,
// a single "with <keyword>[ and <keyword>]" rider, an explicit "named <Name>",
// and a trailing "that's/that are [tapped and] attacking" entry clause (CR
// 508.4).
func creatureTokenSpecBody(effect *EffectSyntax) (func(countWord, noun string) string, bool) {
	sel := effect.Selection
	if len(sel.SubtypesAny) < 1 || len(sel.SubtypesAny) > 2 ||
		len(sel.ColorsAny) > 2 ||
		len(sel.ExcludedTypes) != 0 || len(sel.ExcludedColors) != 0 ||
		sel.Multicolored ||
		sel.MatchPower || sel.MatchToughness || sel.MatchManaValue ||
		sel.Untapped || sel.Blocking ||
		sel.All || sel.Another || sel.Other {
		return nil, false
	}
	supertypePart, ok := tokenSupertypePart(sel)
	if !ok {
		return nil, false
	}
	typeWords, ok := tokenCreatureTypeWords(sel)
	if !ok {
		return nil, false
	}
	keywordPart, ok := tokenKeywordPart(effect.TokenKeywords)
	if !ok {
		return nil, false
	}
	// A token's quoted granted ability ("... token with \"When this token dies,
	// ...\"") is stripped from the clause text to a bare trailing connector;
	// reconstruct that connector so the byte-exact comparison succeeds. With no
	// keyword rider the connector is the bare "with" the granted ability follows;
	// with a keyword rider the granted ability is the final item of the
	// "with <keyword>[, <keyword>], and \"...\"" list, so the connector is the
	// "and"/", and" that joins it to the keyword words.
	grantedPart := ""
	if effect.TokenGrantedAbility != nil {
		if len(effect.TokenKeywords) == 0 {
			grantedPart = " with"
		} else {
			rider, ok := tokenKeywordGrantedRiderPart(effect.TokenKeywords)
			if !ok {
				return nil, false
			}
			keywordPart = rider
		}
	}
	colorPart, ok := tokenColorPart(sel)
	if !ok {
		return nil, false
	}
	ptPart := fmt.Sprintf("%d/%d", effect.TokenPower, effect.TokenToughness)
	if effect.TokenPTVariableX {
		ptPart = "X/X"
	}
	subtypeWords := make([]string, 0, len(sel.SubtypesAny))
	for _, sub := range sel.SubtypesAny {
		subtypeWords = append(subtypeWords, string(sub))
	}
	subtypeJoin := strings.Join(subtypeWords, " ")
	namePart := ""
	if effect.TokenName != "" && !effect.TokenNameLeading {
		namePart = " named " + effect.TokenName
	}
	// The leading "Create <Name>, a ..." form prints the name as a "<Name>, "
	// prefix on the token spec; record it so the closure renders it ahead of the
	// count word.
	leadingNamePart := ""
	if effect.TokenName != "" && effect.TokenNameLeading {
		leadingNamePart = effect.TokenName + ", "
	}
	// A token entering attacking carries a trailing "that's/that are [tapped
	// and] attacking" relative clause; its "tapped" modifier lives in that clause
	// rather than as a leading adjective, so the leading tapped slot is cleared
	// whenever the attacking clause is present.
	tappedPart := ""
	if sel.Tapped && !sel.Attacking {
		tappedPart = "tapped "
	}
	return func(countWord, noun string) string {
		return fmt.Sprintf("%s%s %s%s%s %s%s %s %s%s%s%s%s",
			leadingNamePart, countWord, tappedPart, supertypePart, ptPart, colorPart,
			subtypeJoin, typeWords, noun, keywordPart, grantedPart, namePart,
			tokenAttackClause(sel, noun, effect.AttackDefender))
	}, true
}

// tokenSupertypePart renders a created creature token's canonical supertype words
// ("legendary "), or "" when the token has no supertype. It accepts only the
// Legendary supertype the named-token forms print; any other supertype fails
// closed.
func tokenSupertypePart(sel SelectionSyntax) (string, bool) {
	if len(sel.Supertypes) == 0 {
		return "", true
	}
	words := make([]string, 0, len(sel.Supertypes))
	for _, supertype := range sel.Supertypes {
		if supertype != SupertypeLegendary {
			return "", false
		}
		word, ok := supertypeWord(supertype)
		if !ok {
			return "", false
		}
		words = append(words, word)
	}
	return strings.Join(words, " ") + " ", true
}

// tokenKeywordPart renders the canonical "with <keyword>[ and <keyword>]" rider
// for a created token's bare creature keywords, or ok=false if any keyword is not
// a representable bare creature keyword.
func tokenKeywordPart(keywords []KeywordKind) (string, bool) {
	if len(keywords) == 0 {
		return "", true
	}
	words := make([]string, 0, len(keywords))
	for _, kw := range keywords {
		if !tokenCreatureKeyword(kw) {
			return "", false
		}
		word, ok := kw.OracleWord()
		if !ok {
			return "", false
		}
		words = append(words, word)
	}
	return " with " + joinKeywordWords(words), true
}

// tokenKeywordGrantedRiderPart renders the "with <keyword>[, <keyword>], and"
// rider for a created token that carries both bare creature keywords and a
// trailing quoted granted ability, where the quoted ability is the final item of
// the with-list ("... token with flying and \"...\"", "... token with flying,
// indestructible, and \"...\""). The quoted body is stripped from the clause
// text, so the rider ends at the connector that would precede it. It joins the
// keyword words and the omitted final item, then drops that placeholder, leaving
// the trailing "and"/", and". It returns ok=false if any keyword is not a
// representable bare creature keyword.
func tokenKeywordGrantedRiderPart(keywords []KeywordKind) (string, bool) {
	words := make([]string, 0, len(keywords)+1)
	for _, kw := range keywords {
		if !tokenCreatureKeyword(kw) {
			return "", false
		}
		word, ok := kw.OracleWord()
		if !ok {
			return "", false
		}
		words = append(words, word)
	}
	joined := joinKeywordWords(append(words, "\x00"))
	joined = strings.TrimRight(strings.TrimSuffix(joined, "\x00"), " ")
	return " with " + joined, true
}

// tokenColorPart renders a created token's canonical color words ("colorless " or
// "white and blue "), or ok=false for an unrepresentable color selection.
func tokenColorPart(sel SelectionSyntax) (string, bool) {
	if sel.Colorless {
		if len(sel.ColorsAny) != 0 {
			return "", false
		}
		return "colorless ", true
	}
	if len(sel.ColorsAny) == 0 {
		return "", true
	}
	words := make([]string, 0, len(sel.ColorsAny))
	for _, c := range sel.ColorsAny {
		word, ok := colorWord(c)
		if !ok {
			return "", false
		}
		words = append(words, word)
	}
	return strings.Join(words, " and ") + " ", true
}

// tokenAttackClause renders the trailing attacking-entry relative clause for a
// created token, or "" when the token does not enter attacking. The relative
// pronoun matches the count noun ("that's" for a single "token", "that are" for
// "tokens"), and the clause includes "tapped and" when the token also enters
// tapped. An explicit attacked defender (CR 508.4) recorded in defender appends
// the recognized "that player[ or a planeswalker they control]" / "that
// opponent" tail so the byte-exact reconstruction covers it.
func tokenAttackClause(sel SelectionSyntax, noun string, defender AttackDefenderKind) string {
	if !sel.Attacking {
		return ""
	}
	relative := "that are"
	if noun == "token" {
		relative = "that's"
	}
	clause := " " + relative + " attacking"
	if sel.Tapped {
		clause = " " + relative + " tapped and attacking"
	}
	switch defender {
	case AttackDefenderThatPlayerOrPlaneswalker:
		clause += " that player or a planeswalker they control"
	case AttackDefenderThatPlayer:
		clause += " that player"
	case AttackDefenderThatOpponent:
		clause += " that opponent"
	default:
	}
	return clause
}

func exactCreateTokenEffectSyntax(effect *EffectSyntax) bool {
	targetRecipient, ok := exactCreateTokenRecipientContext(effect)
	if !ok || (!effect.TokenPTKnown && !effect.TokenPTVariableX) || effect.Negated {
		return false
	}
	specBody, ok := creatureTokenSpecBody(effect)
	if !ok {
		return false
	}
	// The referenced-object-controller form ("Its controller creates ...") and
	// the targeted-player form ("Target opponent creates ...") both name their
	// creating player as the clause subject and accept fixed counts only; dynamic
	// counts attach to the controller form.
	if effect.Context == EffectContextReferencedObjectController || targetRecipient {
		if effect.Amount.DynamicForm != EffectDynamicAmountFormNone ||
			!effect.Amount.Known || effect.Amount.Value < 1 {
			return false
		}
		subject := referencedControllerSubjectText(effect)
		if subject == "" {
			return false
		}
		countWord, noun := "a", "token"
		if effect.Amount.Value != 1 {
			countWord, noun = effectAmountSourceText(effect), "tokens"
		}
		return strings.EqualFold(exactEffectClauseText(effect), subject+" creates "+specBody(countWord, noun)+".")
	}
	switch effect.Amount.DynamicForm {
	case EffectDynamicAmountFormNone:
		if effect.Amount.VariableX {
			return createTokenControllerClauseMatches(exactEffectClauseText(effect), specBody("X", "tokens")+".")
		}
		if effect.Amount.DynamicKind == EffectDynamicAmountTriggeringCombatDamage {
			return createTokenControllerClauseMatches(exactEffectClauseText(effect), specBody("that many", "tokens")+".")
		}
		if !effect.Amount.Known || effect.Amount.Value < 1 {
			return false
		}
		countWord, noun := "a", "token"
		if effect.TokenPTVariableX {
			countWord = "an"
		}
		if effect.Amount.Value != 1 {
			countWord, noun = effectAmountSourceText(effect), "tokens"
		}
		return createTokenControllerClauseMatches(exactEffectClauseText(effect), specBody(countWord, noun)+".")
	case EffectDynamicAmountFormForEach:
		if effect.Amount.DynamicKind == EffectDynamicAmountNone || effect.Amount.Multiplier != 1 {
			return false
		}
		spec := specBody("a", "token")
		return createTokenControllerForEachClauseMatches(fullEffectClauseText(effect), spec, effect.Amount.Text)
	case EffectDynamicAmountFormEqual:
		if effect.Amount.DynamicKind == EffectDynamicAmountNone {
			return false
		}
		// Oracle wording for these counts retains the "You" subject ("You create
		// a number of X tokens equal to ..."), which effectSubjectStart does not
		// strip, so accept both the bare and "You" prefixes.
		spec := specBody("a number of", "tokens")
		return createTokenControllerClauseMatches(exactEffectClauseText(effect), spec+" "+effect.Amount.Text+".")
	case EffectDynamicAmountFormWhereX:
		if effect.Amount.DynamicKind == EffectDynamicAmountNone && !effect.Amount.VariableX {
			return false
		}
		if effect.TokenPTVariableX {
			// "Create an X/X ... token, where X is <dynamic>": the variable X sizes
			// the token's printed power and toughness rather than counting tokens, so
			// the count is the singular article and the trailing "where X is" clause
			// binds the size. One such token is created.
			return createTokenControllerClauseMatches(exactEffectClauseText(effect),
				specBody("an", "token")+", "+effect.Amount.Text+".")
		}
		return createTokenControllerClauseMatches(exactEffectClauseText(effect),
			specBody("X", "tokens")+", "+effect.Amount.Text+".")
	default:
		return false
	}
}

// namedArtifactTokenSubtype reports whether sub is a predefined artifact token
// whose fixed Oracle ability the runtime CreateToken/TokenDef model already
// represents (Treasure, Food, Clue, Blood, Gold, Lander, Mutagen, Map, Junk,
// Powerstone). Every other named token (Incubator, whose transform ability is
// not yet modeled) fails closed pending follow-up work.
func namedArtifactTokenSubtype(sub types.Sub) bool {
	switch sub {
	case types.Treasure, types.Food, types.Clue, types.Blood,
		types.Gold, types.Lander, types.Mutagen, types.Map, types.Junk,
		types.Powerstone:
		return true
	default:
		return false
	}
}

// exactCreateNamedTokenEffectSyntax recognizes "Create a [tapped] <Named> token."
// for a predefined artifact token that carries no printed power/toughness
// (Treasure, Food, Clue, Blood), including a fixed count ("Create two Treasure
// tokens."), an optional "tapped" entry modifier ("Create a tapped Treasure
// token."), a "for each <iterator>" count ("Create a Treasure token for each
// artifact you control."), the referenced-controller form ("Its controller
// creates a Treasure token."), and the targeted-player form ("Target opponent
// creates two Treasure tokens."). It fails closed for every richer shape
// (colored, keyworded, or any other named token).
func exactCreateNamedTokenEffectSyntax(effect *EffectSyntax) bool {
	targetRecipient, ok := exactCreateTokenRecipientContext(effect)
	if !ok ||
		effect.TokenPTKnown || effect.TokenCopyOfTarget ||
		effect.Negated {
		return false
	}
	// The spell's variable X count ("Create X Treasure tokens.") and the
	// "for each <iterator>" count attach only to the controller form; the
	// referenced-object-controller and targeted-player forms accept fixed
	// counts only, mirroring the creature-token path.
	controllerForm := effect.Context != EffectContextReferencedObjectController && !targetRecipient
	variableCount := effect.Amount.VariableX &&
		effect.Amount.DynamicForm == EffectDynamicAmountFormNone && controllerForm
	dynamicCombatDamageCount := effect.Amount.DynamicKind == EffectDynamicAmountTriggeringCombatDamage &&
		effect.Amount.DynamicForm == EffectDynamicAmountFormNone && controllerForm
	// A "number of <Named> tokens equal to <dynamic>" count ("Create a number of
	// Food tokens equal to the number of opponents you have.", "Create a number
	// of tapped Treasure tokens equal to its power.") mirrors the creature-token
	// path's FormEqual handling: any non-None dynamic kind the count lowerer
	// already represents drives the token count. The die-roll-result count
	// ("equal to the result") is one such FormEqual dynamic.
	equalDynamicCount := effect.Amount.DynamicForm == EffectDynamicAmountFormEqual &&
		effect.Amount.DynamicKind != EffectDynamicAmountNone && controllerForm
	forEachCount := effect.Amount.DynamicForm == EffectDynamicAmountFormForEach &&
		effect.Amount.DynamicKind != EffectDynamicAmountNone &&
		effect.Amount.Multiplier == 1 && controllerForm
	whereXDynamicCount := effect.Amount.DynamicForm == EffectDynamicAmountFormWhereX &&
		(effect.Amount.DynamicKind != EffectDynamicAmountNone || effect.Amount.VariableX) &&
		effect.Amount.Multiplier == 1 && controllerForm
	if !variableCount && !dynamicCombatDamageCount && !equalDynamicCount && !forEachCount &&
		!whereXDynamicCount && (!effect.Amount.Known || effect.Amount.Value < 1) {
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
	if equalDynamicCount {
		spec := fmt.Sprintf("a number of %s%s tokens", tappedPart, string(sel.SubtypesAny[0]))
		return createTokenControllerClauseMatches(exactEffectClauseText(effect), spec+" "+effect.Amount.Text+".")
	}
	if forEachCount {
		spec := fmt.Sprintf("a %s%s token", tappedPart, string(sel.SubtypesAny[0]))
		return createTokenControllerForEachClauseMatches(fullEffectClauseText(effect), spec, effect.Amount.Text)
	}
	if whereXDynamicCount {
		spec := fmt.Sprintf("X %s%s tokens", tappedPart, string(sel.SubtypesAny[0]))
		return createTokenControllerClauseMatches(exactEffectClauseText(effect),
			spec+", "+effect.Amount.Text+".")
	}
	countWord, noun := "a", "token"
	switch {
	case variableCount:
		countWord, noun = "X", "tokens"
	case dynamicCombatDamageCount:
		countWord, noun = "that many", "tokens"
	case effect.Amount.Value != 1:
		countWord, noun = effectAmountSourceText(effect), "tokens"
	default:
	}
	specBody := fmt.Sprintf("%s %s%s %s", countWord, tappedPart, string(sel.SubtypesAny[0]), noun)
	if effect.Context == EffectContextReferencedObjectController || targetRecipient {
		subject := referencedControllerSubjectText(effect)
		if subject == "" {
			return false
		}
		return strings.EqualFold(exactEffectClauseText(effect), subject+" creates "+specBody+".")
	}
	return createTokenControllerClauseMatches(exactEffectClauseText(effect), specBody+".")
}

// exactCreatePredefinedTokenEffectSyntax recognizes "Create a [tapped] <Name>
// token." for a predefined named token whose name is a card name rather than a
// card subtype (Mutavault). Such a token carries no printed power/toughness,
// color, subtype, keyword, or count modifier in its create clause; the name
// alone identifies it (its characteristics live in its own definition). It
// accepts only the controller recipient form with a single fixed count and an
// optional leading "tapped" adjective, and byte-checks the reconstructed
// "Create a [tapped] <Name> token." clause. Every richer shape (a count other
// than one, a non-controller recipient, or any selection qualifier beyond
// "tapped") fails closed.
func exactCreatePredefinedTokenEffectSyntax(effect *EffectSyntax) bool {
	if effect.TokenPredefinedName == "" ||
		effect.Context != EffectContextController ||
		effect.TokenPTKnown || effect.TokenCopyOfTarget ||
		effect.TokenChoice || effect.Negated ||
		len(effect.Targets) != 0 ||
		!effect.Amount.Known || effect.Amount.Value != 1 {
		return false
	}
	sel := effect.Selection
	if sel.Kind != SelectionUnknown ||
		len(sel.SubtypesAny) != 0 ||
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
	specBody := fmt.Sprintf("a %s%s token", tappedPart, effect.TokenPredefinedName)
	return createTokenControllerClauseMatches(exactEffectClauseText(effect), specBody+".")
}

// exactCreateNamedTokenChoiceEffectSyntax recognizes an N-way (N >= 2) choice
// among predefined artifact tokens (Treasure, Food, Clue, Blood, ...), each
// named by its own subtype with no printed power/toughness. It accepts both the
// bare two-way "Create a <A> token or a <B> token." and the "your choice of"
// list form "Create your choice of a <A> token, a <B> token, or a <C> token."
// (an Oxford-comma list ending in "or", any count >= 2), plus the
// referenced-controller and targeted-player recipient variants and the
// lowercase-verb form used inside embedded trigger/ability bodies. The effect
// creates exactly one of the alternatives; lowering emits a choose-one modal
// ability. It fails closed for every richer shape (colored, keyworded, tapped,
// counts other than one, or any non-predefined token).
func exactCreateNamedTokenChoiceEffectSyntax(effect *EffectSyntax) bool {
	targetRecipient, ok := exactCreateTokenRecipientContext(effect)
	if !ok || !effect.TokenChoice ||
		effect.TokenPTKnown || effect.TokenCopyOfTarget ||
		effect.Negated ||
		effect.Amount.DynamicForm == EffectDynamicAmountFormForEach ||
		!effect.Amount.Known || effect.Amount.Value != 1 {
		return false
	}
	sel := effect.Selection
	if sel.Kind != SelectionUnknown ||
		len(sel.SubtypesAny) < 2 ||
		sel.Keyword != KeywordUnknown ||
		len(sel.ColorsAny) != 0 || len(sel.ExcludedColors) != 0 ||
		len(sel.RequiredTypesAny) != 0 || len(sel.ExcludedTypes) != 0 ||
		len(sel.SourceTypes) != 0 || len(sel.Supertypes) != 0 ||
		sel.MatchPower || sel.MatchToughness || sel.MatchManaValue ||
		sel.Tapped || sel.Untapped || sel.Attacking || sel.Blocking ||
		sel.All || sel.Another || sel.Other ||
		sel.Colorless || sel.Multicolored {
		return false
	}
	seen := make(map[types.Sub]bool, len(sel.SubtypesAny))
	for _, sub := range sel.SubtypesAny {
		if seen[sub] || !namedArtifactTokenSubtype(sub) {
			return false
		}
		seen[sub] = true
	}
	listBody := namedTokenChoiceListBody(sel.SubtypesAny)
	clause := exactEffectClauseText(effect)
	verbBody := func(verb string) bool {
		return strings.EqualFold(clause, verb+" "+listBody+".") ||
			strings.EqualFold(clause, verb+" your choice of "+listBody+".")
	}
	if effect.Context == EffectContextReferencedObjectController || targetRecipient {
		subject := referencedControllerSubjectText(effect)
		if subject == "" {
			return false
		}
		return verbBody(subject + " creates")
	}
	return verbBody("Create")
}

// namedTokenChoiceListBody renders the canonical alternatives list for a
// named-token choice: "a <A> token or a <B> token" for two alternatives and an
// Oxford-comma list ending in "or" for three or more ("a <A> token, a <B>
// token, or a <C> token").
func namedTokenChoiceListBody(subtypes []types.Sub) string {
	items := make([]string, 0, len(subtypes))
	for _, sub := range subtypes {
		items = append(items, "a "+string(sub)+" token")
	}
	if len(items) == 2 {
		return items[0] + " or " + items[1]
	}
	return strings.Join(items[:len(items)-1], ", ") + ", or " + items[len(items)-1]
}

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
// <target>[, except <it/the token> isn't legendary]." where the token copies the
// effect's single exact target object (e.g. "Create a token that's a copy of
// target creature you control."). The optional "except ... isn't legendary"
// modifier is recorded on TokenCopyDropLegendary. It fails closed for every
// richer copy shape (other copy modifiers, multiple tokens, non-target copy
// sources).
func exactCreateCopyTokenEffectSyntax(effect *EffectSyntax) bool {
	if effect.Context != EffectContextController ||
		effect.Negated ||
		!createCopyTokenCountKnown(effect) ||
		len(effect.Targets) != 1 ||
		!effect.Targets[0].Exact {
		return false
	}
	base, rider, ok := copyTokenExceptModifier(effect, exactEffectClauseText(effect))
	if !ok {
		return false
	}
	entersTapped, matched := createCopyTokenClauseMatches(effect, base, effect.Targets[0].Text)
	if !matched {
		return false
	}
	effect.TokenCopyDropLegendary = rider.dropLegendary
	effect.TokenCopyGrantKeywords = rider.grantKeywords
	effect.TokenCopyEntersTapped = entersTapped
	if rider.override != nil {
		applyCopyTokenOverride(effect, *rider.override)
	}
	return true
}

// createCopyTokenCountKnown reports whether the effect creates a fixed, positive
// number of copy tokens (one or more). Copy-token shapes accept a known integer
// count and reject dynamic counts ("X tokens", "that many tokens"), which the
// copy-token backend does not yet lower.
func createCopyTokenCountKnown(effect *EffectSyntax) bool {
	return effect.Amount.Known &&
		effect.Amount.Value >= 1 &&
		!effect.Amount.VariableX &&
		effect.Amount.DynamicForm == EffectDynamicAmountFormNone
}

// createCopyTokenClause builds the canonical create-copy-token clause for a known
// count, choosing the singular "Create a token that's a copy of <source>." for a
// count of one and the plural "Create <count> tokens that are copies of
// <source>." (Saw in Half, Gruff Triplets) otherwise. The plural count word is
// the effect's verbatim source text so the comparison matches the printed
// grammatical number. When tapped is set the "tapped" entry adjective is inserted
// ("Create a tapped token that's a copy of <source>.", Compy Swarm).
func createCopyTokenClause(effect *EffectSyntax, source string, tapped bool) string {
	tappedWord := ""
	if tapped {
		tappedWord = "tapped "
	}
	if effect.Amount.Value == 1 {
		return "Create a " + tappedWord + "token that's a copy of " + source + "."
	}
	return "Create " + effectAmountSourceText(effect) + " " + tappedWord + "tokens that are copies of " + source + "."
}

// createCopyTokenClauseMatches reports whether base equals the canonical create-
// copy-token clause for source, accepting the optional "tapped" entry modifier.
// It returns whether the tapped variant matched and whether either variant
// matched at all.
func createCopyTokenClauseMatches(effect *EffectSyntax, base, source string) (tapped, ok bool) {
	if strings.EqualFold(base, createCopyTokenClause(effect, source, false)) {
		return false, true
	}
	if strings.EqualFold(base, createCopyTokenClause(effect, source, true)) {
		return true, true
	}
	return false, false
}

// exactCreateCopyTokenReferenceEffectSyntax reports whether the effect is
// "Create a token that's a copy of <reference>[ instead][, except <it/the token>
// isn't legendary]." where the copy source is an explicit reference ("this
// creature", the card's own name, or the pronoun "it") rather than a grammatical
// target. The trailing " instead" suffix (the conditional-replacement form,
// recorded separately in the effect's Replacement) is stripped before
// comparison. An "except it isn't legendary" modifier adds a second pronoun
// reference; only the copy-source reference completes the base clause, and any
// remaining references must be the modifier's pronoun. It requires no targets, a
// single token, and the controller recipient.
func exactCreateCopyTokenReferenceEffectSyntax(effect *EffectSyntax) bool {
	if effect.Context != EffectContextController ||
		effect.Negated ||
		!createCopyTokenCountKnown(effect) ||
		len(effect.Targets) != 0 ||
		len(effect.References) == 0 {
		return false
	}
	base, rider, ok := copyTokenExceptModifier(effect, exactEffectClauseText(effect))
	if !ok {
		return false
	}
	clause := strings.TrimSuffix(base, ".")
	clause = strings.TrimSuffix(clause, " instead")
	sourceIndex := -1
	entersTapped := false
	for i := range effect.References {
		if !copyTokenReferenceSupported(effect.References[i]) {
			continue
		}
		tapped, matched := createCopyTokenClauseMatches(effect, clause+".", effect.References[i].Text)
		if matched {
			sourceIndex = i
			entersTapped = tapped
			break
		}
	}
	if sourceIndex < 0 {
		return false
	}
	for i := range effect.References {
		if i == sourceIndex {
			continue
		}
		if effect.References[i].Kind != ReferencePronoun {
			return false
		}
	}
	effect.TokenCopyDropLegendary = rider.dropLegendary
	effect.TokenCopyGrantKeywords = rider.grantKeywords
	effect.TokenCopyEntersTapped = entersTapped
	if rider.override != nil {
		applyCopyTokenOverride(effect, *rider.override)
	}
	return true
}

// exactCreateCopyTokenTriggeringSetEffectSyntax reports whether the effect is
// "Create a token that's a copy of one of them[, except <it/the token> isn't
// legendary]." where "them" denotes the set of permanents that triggered the
// enclosing ability and the controller chooses one to copy ("Whenever one or
// more other creatures you control enter, ... create a token that's a copy of
// one of them.", Twilight Diviner). The copy source is the controller-chosen
// member of the triggering event batch; the source phrase "one of them" is
// fixed and the effect's references must be the "them"/"they" pronouns naming
// that set (plus any modifier pronoun). It requires no targets, the controller
// recipient, and a known positive count.
func exactCreateCopyTokenTriggeringSetEffectSyntax(effect *EffectSyntax) bool {
	if effect.Context != EffectContextController ||
		effect.TokenPTKnown ||
		effect.Negated ||
		!createCopyTokenCountKnown(effect) ||
		len(effect.Targets) != 0 ||
		!referencesIncludeThemPronoun(effect.References) {
		return false
	}
	base, rider, ok := copyTokenExceptModifier(effect, exactEffectClauseText(effect))
	if !ok {
		return false
	}
	entersTapped, matched := createCopyTokenClauseMatches(effect, base, "one of them")
	if !matched {
		return false
	}
	for i := range effect.References {
		if effect.References[i].Kind != ReferencePronoun {
			return false
		}
	}
	effect.TokenCopyDropLegendary = rider.dropLegendary
	effect.TokenCopyGrantKeywords = rider.grantKeywords
	effect.TokenCopyEntersTapped = entersTapped
	return true
}

// referencesIncludeThemPronoun reports whether references holds the "them" (or
// "they") pronoun that names the triggering-event set for a "copy of one of
// them" token create.
func referencesIncludeThemPronoun(references []Reference) bool {
	for i := range references {
		if references[i].Kind == ReferencePronoun &&
			(references[i].Pronoun == PronounThem || references[i].Pronoun == PronounThey) {
			return true
		}
	}
	return false
}

// exactCreateCopyTokenAttachedEffectSyntax reports whether the effect is "Create
// a token that's a copy of equipped creature." or "... enchanted creature." (the
// permanent the source Equipment or Aura is attached to), with an optional
// "except <it/the token> isn't legendary" modifier recorded on
// TokenCopyDropLegendary. Any references must be the modifier's pronoun.
func exactCreateCopyTokenAttachedEffectSyntax(effect *EffectSyntax) bool {
	if effect.Context != EffectContextController ||
		effect.Negated ||
		!createCopyTokenCountKnown(effect) ||
		len(effect.Targets) != 0 {
		return false
	}
	base, rider, ok := copyTokenExceptModifier(effect, exactEffectClauseText(effect))
	if !ok {
		return false
	}
	equippedTapped, equippedOK := createCopyTokenClauseMatches(effect, base, "equipped creature")
	enchantedTapped, enchantedOK := createCopyTokenClauseMatches(effect, base, "enchanted creature")
	if !equippedOK && !enchantedOK {
		return false
	}
	for i := range effect.References {
		if effect.References[i].Kind != ReferencePronoun {
			return false
		}
	}
	effect.TokenCopyDropLegendary = rider.dropLegendary
	effect.TokenCopyGrantKeywords = rider.grantKeywords
	effect.TokenCopyEntersTapped = equippedTapped || enchantedTapped
	if rider.override != nil {
		applyCopyTokenOverride(effect, *rider.override)
	}
	return true
}

// exactCreateCopyTokenForEachEffectSyntax reports whether the effect is a
// per-each copy-token create whose copy source is each member of a controlled
// battlefield group: "For each <permanent filter> you control, create a token
// that's a copy of that permanent." (Second Harvest) or the "... a copy of it"
// variant. The created token copies each iterated permanent in turn; the
// trailing "that permanent"/"that token"/"it" reference names the per-iteration
// member rather than a single fixed source. It returns the iterated group
// selection, parsed from the pre-verb "For each <group>," prefix, when the shape
// matches. It requires the controller recipient, a single (per-iteration) token,
// no targets, and no fixed power/toughness.
func exactCreateCopyTokenForEachEffectSyntax(effect *EffectSyntax, atoms Atoms) (*SelectionSyntax, bool) {
	if effect.Kind != EffectCreate ||
		effect.Context != EffectContextController ||
		effect.TokenPTKnown ||
		effect.Negated ||
		!effect.Amount.Known || effect.Amount.Value != 1 ||
		len(effect.Targets) != 0 {
		return nil, false
	}
	verb := slices.IndexFunc(effect.Tokens, func(token shared.Token) bool {
		return token.Span == effect.VerbSpan
	})
	if verb < 0 {
		return nil, false
	}
	group, ok := parseForEachControlledGroup(effect.Tokens[:verb], atoms)
	if !ok {
		return nil, false
	}
	base, rider, ok := copyTokenExceptModifier(effect, copyForEachClauseText(effect, verb))
	if !ok {
		return nil, false
	}
	entersTapped, matched := copyForEachSourcePhrase(base)
	if !matched {
		return nil, false
	}
	effect.TokenCopyDropLegendary = rider.dropLegendary
	effect.TokenCopyGrantKeywords = rider.grantKeywords
	effect.TokenCopyEntersTapped = entersTapped
	return group, true
}

// parseForEachControlledGroup parses a leading "For each <permanent filter> you
// control," iteration prefix into the controlled battlefield group it iterates.
// It reuses parseDynamicAmountPrefix to recognize the "for each" form and
// parseSelection to model the group's filter, requiring the trailing comma and a
// "you control" controller scope so only controlled groups qualify.
func parseForEachControlledGroup(pre []shared.Token, atoms Atoms) (*SelectionSyntax, bool) {
	prefix, ok := parseDynamicAmountPrefix(pre, 0, atoms)
	if !ok || prefix.form != EffectDynamicAmountFormForEach {
		return nil, false
	}
	if len(pre) == 0 || pre[len(pre)-1].Kind != shared.Comma {
		return nil, false
	}
	groupTokens := pre[prefix.start : len(pre)-1]
	if len(groupTokens) == 0 {
		return nil, false
	}
	selection := parseSelection(groupTokens, atoms)
	if selection.Controller != SelectionControllerYou {
		return nil, false
	}
	return &selection, true
}

// copyForEachClauseText reconstructs the post-prefix create clause ("Create a
// token that's a copy of that permanent.") from the verb token onward, restoring
// the trailing period the sentence split may have dropped.
func copyForEachClauseText(effect *EffectSyntax, verb int) string {
	text := joinedEffectText(effect.Tokens[verb:])
	if len(effect.Tokens) > 0 && effect.Tokens[len(effect.Tokens)-1].Kind != shared.Period {
		text += "."
	}
	return text
}

// copyForEachSourcePhrase reports whether base names a per-each copy source that
// refers to the iterated group member ("that permanent", "that token", "that
// creature", or the pronoun "it"), and whether the optional "tapped" entry
// modifier is present.
func copyForEachSourcePhrase(base string) (tapped, ok bool) {
	switch {
	case strings.EqualFold(base, "Create a token that's a copy of that permanent."),
		strings.EqualFold(base, "Create a token that's a copy of that token."),
		strings.EqualFold(base, "Create a token that's a copy of that creature."),
		strings.EqualFold(base, "Create a token that's a copy of it."):
		return false, true
	case strings.EqualFold(base, "Create a tapped token that's a copy of that permanent."),
		strings.EqualFold(base, "Create a tapped token that's a copy of that token."),
		strings.EqualFold(base, "Create a tapped token that's a copy of that creature."),
		strings.EqualFold(base, "Create a tapped token that's a copy of it."):
		return true, true
	default:
		return false, false
	}
}

// copyTokenExceptRider holds the recognized copiable modifiers of a copy-token
// "except <rider>" clause: whether the copy drops its legendary supertype and
// the keyword abilities it gains.
type copyTokenExceptRider struct {
	dropLegendary bool
	grantKeywords []KeywordKind
	override      *copyTokenOverride
}

// copyTokenExceptModifier splits a copy-token clause into its base "Create a
// token that's a copy of <source>." text and the recognized trailing copiable
// modifiers. A clause with no ", except" suffix returns the clause unchanged
// with no modifiers. The bare "except <it/the token> isn't legendary" form is
// recognized directly. A richer "except <the token/it> has <keyword>[ and ...][,
// and] <it/the token> isn't legendary" rider (Irenicus's Vile Duplication) is
// recognized from the effect's rider tokens, returning the granted copiable
// keywords. Any other except modifier (power/toughness, added types, quoted
// abilities) is unrecognized and returns ok=false so the copy fails closed.
func copyTokenExceptModifier(effect *EffectSyntax, clause string) (base string, rider copyTokenExceptRider, ok bool) {
	body, hadPeriod := strings.CutSuffix(clause, ".")
	if !hadPeriod {
		return clause, copyTokenExceptRider{}, true
	}
	head, except, found := strings.Cut(body, ", except ")
	if !found {
		return clause, copyTokenExceptRider{}, true
	}
	switch normalizeApostrophes(strings.ToLower(strings.TrimSpace(except))) {
	case "it isn't legendary", "it is not legendary", "it's not legendary",
		"the token isn't legendary", "the token is not legendary":
		return head + ".", copyTokenExceptRider{dropLegendary: true}, true
	}
	if override, ok := copyTokenExceptOverride(effect); ok {
		return head + ".", copyTokenExceptRider{dropLegendary: override.dropLegendary, override: &override}, true
	}
	drop, keywords, riderOK := copyTokenExceptRiderTokens(effect)
	if !riderOK {
		return "", copyTokenExceptRider{}, false
	}
	return head + ".", copyTokenExceptRider{dropLegendary: drop, grantKeywords: keywords}, true
}

// copyTokenExceptRiderTokens parses the copiable rider that follows a copy-token
// "except" clause from the effect's tokens. It splits the rider after the final
// "except" word into "and"/comma-separated sub-clauses and accepts only the
// "<it/the token> isn't legendary" drop-legendary clause and one or more
// "<it/the token> has <keyword>" keyword-grant clauses. A quoted granted ability
// or parenthetical in the rider ("...except it has haste and \"At the beginning
// of the end step, sacrifice this token.\"", Electroduplicate/Heat Shimmer)
// carries semantics the keyword/legendary recognizer cannot represent, so it
// fails closed rather than silently dropping it. It also fails closed when any
// sub-clause is unrecognized or when no modifier at all is recognized, so the
// copy stays unsupported rather than silently dropping rider semantics.
func copyTokenExceptRiderTokens(effect *EffectSyntax) (dropLegendary bool, grantKeywords []KeywordKind, ok bool) {
	raw := effect.Tokens
	exceptIndex := -1
	for i := range raw {
		if equalWord(raw[i], "except") {
			exceptIndex = i
		}
	}
	if exceptIndex < 0 {
		return false, nil, false
	}
	rider := raw[exceptIndex+1:]
	for _, token := range rider {
		if token.Kind == shared.Quote || token.Kind == shared.LeftParen {
			return false, nil, false
		}
	}
	for len(rider) > 0 && rider[len(rider)-1].Kind == shared.Period {
		rider = rider[:len(rider)-1]
	}
	// A rider that ends in a dangling conjunction or comma signals that trailing
	// content (a quoted granted ability that the clause builder dropped from the
	// effect tokens) was elided, so fail closed rather than recognizing only the
	// surviving prefix.
	if len(rider) > 0 {
		last := rider[len(rider)-1]
		if last.Kind == shared.Comma || equalWord(last, "and") || equalWord(last, "or") {
			return false, nil, false
		}
	}
	clauses := splitEntersAsCopyRiderClauses(rider)
	if len(clauses) == 0 {
		return false, nil, false
	}
	for _, clause := range clauses {
		if copyTokenNotLegendaryClause(normalizedWords(clause)) {
			dropLegendary = true
			continue
		}
		if keyword, kok := copyTokenHasKeywordClause(clause); kok {
			grantKeywords = append(grantKeywords, keyword)
			continue
		}
		return false, nil, false
	}
	if !dropLegendary && len(grantKeywords) == 0 {
		return false, nil, false
	}
	return dropLegendary, grantKeywords, true
}

// copyTokenNotLegendaryClause reports whether the rider sub-clause words are an
// "<it/the token> isn't legendary" copiable rider. The "the token" subject is
// normalized to the "it" subject the shared enters-as-copy recognizer accepts.
func copyTokenNotLegendaryClause(words []string) bool {
	if len(words) >= 2 && words[0] == "the" && words[1] == "token" {
		words = append([]string{"it"}, words[2:]...)
	}
	return entersAsCopyNotLegendaryClause(words)
}

// copyTokenHasKeywordClause recognizes an "<it/the token> has <keyword>" copiable
// rider sub-clause and returns the single granted keyword. The keyword name must
// consume the entire remainder of the clause so trailing words (parameters or
// extra text) fail closed.
func copyTokenHasKeywordClause(clause []shared.Token) (KeywordKind, bool) {
	rest, ok := copyTokenHasSubject(clause)
	if !ok {
		return KeywordUnknown, false
	}
	kind, width, ok := recognizeKeywordNameAt(rest, 0)
	if !ok || width != len(rest) {
		return KeywordUnknown, false
	}
	return kind, true
}

// copyTokenHasSubject strips a leading "it has" or "the token has" subject from a
// rider sub-clause, returning the remaining keyword-name tokens.
func copyTokenHasSubject(clause []shared.Token) ([]shared.Token, bool) {
	switch {
	case len(clause) >= 3 && equalWord(clause[0], "it") && equalWord(clause[1], "has"):
		return clause[2:], true
	case len(clause) >= 4 && equalWord(clause[0], "the") && equalWord(clause[1], "token") && equalWord(clause[2], "has"):
		return clause[3:], true
	default:
		return nil, false
	}
}

// normalizeApostrophes converts curly apostrophes to straight ones so modifier
// matching is independent of the source's apostrophe spelling.
func normalizeApostrophes(text string) string {
	return strings.ReplaceAll(text, "\u2019", "'")
}

// copyTokenReferenceSupported reports whether a reference can name the copy
// source of a copy-of-reference token: an explicit self reference ("this
// creature"/"this permanent" or the card's own name), the pronoun "it", or a
// "that <permanent>" antecedent ("that creature", "that permanent", "that
// token") that resolves to a single triggering permanent ("Whenever a nontoken
// Zombie you control enters, create a token that's a copy of that creature." —
// Necroduality). The compiler binds the "that" antecedent and the lowering fails
// closed when it does not resolve to a supported single object.
func copyTokenReferenceSupported(reference Reference) bool {
	switch reference.Kind {
	case ReferenceThisObject, ReferenceSelfName:
		return true
	case ReferenceThatObject:
		return copyTokenThatAntecedentText(reference.Text)
	case ReferencePronoun:
		return reference.Pronoun == PronounIt
	default:
		return false
	}
}

// copyTokenThatAntecedentText reports whether text is a "that <permanent>"
// antecedent the copy-of-reference token form accepts as its copy source. Only
// the bare permanent antecedents qualify; richer "that creature an opponent
// controls" phrasings carry extra words and are rejected so the copy fails
// closed.
func copyTokenThatAntecedentText(text string) bool {
	switch normalizeApostrophes(strings.ToLower(strings.TrimSpace(text))) {
	case "that creature", "that permanent", "that token":
		return true
	default:
		return false
	}
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

// joinKeywordWords joins token keyword Oracle words the way Oracle text lists a
// token's "with" rider: a single word as-is, two words joined by "and", and three
// or more in an Oxford-comma series ("flying, vigilance, and indestructible").
func joinKeywordWords(words []string) string {
	switch len(words) {
	case 0:
		return ""
	case 1:
		return words[0]
	case 2:
		return words[0] + " and " + words[1]
	default:
		return strings.Join(words[:len(words)-1], ", ") + ", and " + words[len(words)-1]
	}
}

// tokenCreatureKeyword reports whether a keyword is a creature combat/evergreen
// keyword that is safe to grant a synthesized creature token through its typed
// static-ability body. The landwalk evasion family (CR 702.14) is included: each
// typed variant ("islandwalk", "swampwalk", ...) and the generic and nonbasic
// forms carry a fixed typed static body the runtime already models.
func tokenCreatureKeyword(k KeywordKind) bool {
	switch k {
	case KeywordChangeling, KeywordFlying, KeywordFirstStrike, KeywordDoubleStrike, KeywordDeathtouch,
		KeywordHaste, KeywordHexproof, KeywordIndestructible, KeywordLifelink,
		KeywordMenace, KeywordReach, KeywordTrample, KeywordVigilance,
		KeywordDefender, KeywordShroud, KeywordWither, KeywordInfect, KeywordProwess,
		KeywordLandwalk, KeywordPlainswalk, KeywordIslandwalk, KeywordSwampwalk,
		KeywordMountainwalk, KeywordForestwalk, KeywordDesertwalk, KeywordNonbasicLandwalk:
		return true
	default:
		return false
	}
}

func exactCardCountEffectSyntax(effect *EffectSyntax, controllerVerb, subjectVerb string, allowDynamic bool) bool {
	// Card-count amounts are routinely spelled out well above the four-card
	// ceiling the legacy fixed-amount shortcut once enforced ("draws seven
	// cards", "mills thirteen cards"). The exact-reconstruction below rebuilds
	// the clause from the original amount token, so a spelled count round-trips
	// byte-for-byte regardless of magnitude; a non-cardinal or otherwise
	// unreconstructable amount simply fails that comparison and stays inexact.
	// This function only ever runs for draw, discard, and mill effects, so the
	// reconstruction is the sole authority for every spelled count.
	if effect.Kind == EffectMill && effect.Amount.DynamicKind == EffectDynamicAmountControllerLife {
		return false
	}
	prefixes := cardCountSubjectPrefixes(effect, controllerVerb, subjectVerb)
	text := exactEffectClauseText(effect)
	singular, plural := "card", "cards"
	if effect.Additional {
		singular, plural = "additional card", "additional cards"
	}
	for _, prefix := range prefixes {
		if exactCountedNounEffectText(text, prefix, singular, plural, effect.Amount, effectAmountSourceText(effect), allowDynamic) {
			return true
		}
	}
	return false
}

// cardCountSubjectPrefixes builds the accepted subject-clause prefix(es) for a
// draw, discard, or mill card-count effect, keyed off the effect's subject
// context. controllerVerb is the controller-voice verb ("Discard"), subjectVerb
// the third-person voice ("discards"). The returned prefixes are compared
// against the reconstructed clause text by the exact-card-count and random
// discard recognizers; an unrecognized context yields no prefixes.
func cardCountSubjectPrefixes(effect *EffectSyntax, controllerVerb, subjectVerb string) []string {
	switch effect.Context {
	case EffectContextController:
		return []string{controllerVerb, "You " + controllerVerb}
	case EffectContextEachPlayer:
		return inabilityAwarePrefixes(effect, "Each player", subjectVerb)
	case EffectContextEachOtherPlayer:
		return inabilityAwarePrefixes(effect, "Each other player", subjectVerb)
	case EffectContextEachOpponent:
		return inabilityAwarePrefixes(effect, "Each opponent", subjectVerb)
	case EffectContextDefendingPlayer:
		return []string{"Defending player " + subjectVerb}
	case EffectContextTarget:
		if len(effect.Targets) == 1 && effect.Targets[0].Exact &&
			exactCardCountTargetPlayer(effect.Targets[0].Selection) {
			return []string{titleFirstEffectText(effect.Targets[0].Text) + " " + subjectVerb}
		}
	case EffectContextControllerAndTarget:
		if len(effect.Targets) == 1 && effect.Targets[0].Exact &&
			exactCardCountTargetPlayer(effect.Targets[0].Selection) {
			// exactEffectClauseText drops the leading "You and" at its "and"
			// split, so the reconstructed clause begins at the target subject:
			// "target opponent each draw a card".
			return []string{effect.Targets[0].Text + " each " + strings.ToLower(controllerVerb)}
		}
	case EffectContextPriorSubject:
		if len(effect.Targets) == 1 && effect.Targets[0].Exact &&
			exactCardCountTargetPlayer(effect.Targets[0].Selection) {
			return []string{titleFirstEffectText(effect.Targets[0].Text) + " " + subjectVerb}
		}
		return []string{controllerVerb, subjectVerb}
	case EffectContextEventPlayer, EffectContextReferencedPlayer:
		return []string{"They " + strings.TrimSuffix(subjectVerb, "s"), "That player " + subjectVerb}
	case EffectContextReferencedObjectController:
		if subject := referencedControllerSubjectText(effect); subject != "" {
			if effect.Optional && effect.Amount.RangeKnown &&
				effect.DelayedTiming == DelayedTimingNextUpkeep {
				subject = strings.TrimSuffix(subject, " may")
				return []string{subject + " may " + strings.TrimSuffix(subjectVerb, "s")}
			}
			return []string{subject + " " + subjectVerb}
		}
	default:
	}
	return nil
}

// exactNonControllerRandomDiscardSyntax reconstructs the canonical
// "<subject> discards <N> card(s) at random." wording for a fixed-count random
// discard by a non-controller subject (each player, each opponent, the
// defending player, a target player, or the "that player"/"they" anaphor).
// Controller random discards are recognized separately by
// exactControllerRandomDiscardSyntax and travel on HandDiscard, so this helper
// skips the controller voice. The "at random" suffix marks the random variant,
// distinguishing it from the player-choice discard exactCardCountEffectSyntax
// recognizes. Only a plain unqualified card selection of a known positive fixed
// count round-trips; every filtered, ranged, dynamic, or whole-hand discard
// fails the reconstruction and stays inexact.
func exactNonControllerRandomDiscardSyntax(effect *EffectSyntax) bool {
	if effect.Kind != EffectDiscard ||
		effect.Context == EffectContextController ||
		effect.DiscardEntireHand ||
		effect.Negated ||
		!effect.Amount.Known || effect.Amount.Value < 1 || effect.Amount.RangeKnown ||
		effect.Amount.DynamicForm != EffectDynamicAmountFormNone {
		return false
	}
	noun := "cards"
	if effect.Amount.Value == 1 {
		noun = "card"
	}
	text := exactEffectClauseText(effect)
	amountText := effectAmountSourceText(effect)
	for _, prefix := range cardCountSubjectPrefixes(effect, "Discard", "discards") {
		if strings.EqualFold(text, fmt.Sprintf("%s %s %s at random.", prefix, amountText, noun)) {
			return true
		}
	}
	return false
}

// drawAdditionalCardsQualifier reports whether a draw clause counts "additional"
// cards ("draw two additional cards", "draw an additional card") — the
// extra-draw wording on draw-step triggers such as Sylvan Library. Drawing N
// additional cards is mechanically a plain draw of N cards; the flag only lets
// exact reconstruction restore the "additional" word. It is false for every
// non-draw effect and every draw without the qualifier.
func drawAdditionalCardsQualifier(effect *EffectSyntax) bool {
	if effect.Kind != EffectDraw {
		return false
	}
	for i := 0; i+1 < len(effect.Tokens); i++ {
		if equalWord(effect.Tokens[i], "additional") &&
			(equalWord(effect.Tokens[i+1], "card") || equalWord(effect.Tokens[i+1], "cards")) {
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
// inabilityAwarePrefixes builds the accepted subject prefix(es) for an
// each-player card-count clause. For a "who can't" fallback rider ("Each player
// who can't discards a card.") the subject carries the relative clause, so the
// accepted prefix is "<subject> who can't <verb>"; otherwise it is the plain
// "<subject> <verb>".
func inabilityAwarePrefixes(effect *EffectSyntax, subject, subjectVerb string) []string {
	if effect.FallbackOnInability {
		return []string{
			subject + " who can't " + subjectVerb,
			subject + " who cannot " + subjectVerb,
		}
	}
	return []string{subject + " " + subjectVerb}
}

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

// exactGiveControlEffectSyntax recognizes the give-control forms whose subject
// is a target player who gains control of a permanent. It covers the two-target
// spell "<target player> gains control of <target permanent>." (Donate,
// Harmless Offering, Wrong Turn) and the source self-gift "<target player>
// gains control of this <object>." (Jinxed Idol, Avarice Amulet, Measure of
// Wickedness), where the controlled object is the ability's own source. The
// clause is exact only when its verbatim reconstruction matches, so any other
// wording leaves it non-exact and lowering fails closed.
func exactGiveControlEffectSyntax(effect *EffectSyntax) bool {
	if effect.Negated || effect.Context != EffectContextTarget {
		return false
	}
	if len(effect.Targets) == 0 || !effect.Targets[0].Exact {
		return false
	}
	suffix := ""
	switch effect.Duration {
	case EffectDurationNone:
		suffix = "."
	case EffectDurationUntilEndOfTurn:
		suffix = " until end of turn."
	default:
		return false
	}
	prefix := effect.Targets[0].Text + " gains control of "
	text := exactEffectClauseText(effect)
	switch {
	case len(effect.Targets) == 2 && len(effect.References) == 0 && effect.Targets[1].Exact:
		return strings.EqualFold(text, prefix+effect.Targets[1].Text+suffix)
	case len(effect.Targets) == 1 && len(effect.References) == 1 &&
		effect.References[0].Kind == ReferenceThisObject:
		rest, ok := strings.CutPrefix(strings.ToLower(text), strings.ToLower(prefix))
		if !ok {
			return false
		}
		object, ok := strings.CutSuffix(rest, suffix)
		if !ok {
			return false
		}
		return giveControlThisObjectNoun(object)
	default:
		return false
	}
}

// giveControlThisObjectNoun reports whether object is a single-word "this
// <noun>" self reference (e.g. "this artifact", "this enchantment", "this
// equipment"). The reference kind already binds the noun to the ability's
// source, so only the shape needs confirming.
func giveControlThisObjectNoun(object string) bool {
	rest, ok := strings.CutPrefix(object, "this ")
	return ok && rest != "" && !strings.Contains(rest, " ")
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
	lower := strings.ToLower(text)
	suffix, ok := strings.CutPrefix(lower, strings.ToLower(prefix)+" for as long as this ")
	if !ok {
		suffix, ok = strings.CutPrefix(lower, strings.ToLower(prefix)+" as long as this ")
	}
	if !ok {
		return false
	}
	self, rest, ok := strings.Cut(suffix, " ")
	if !ok || self == "" {
		return false
	}
	// "this <type>" is a single-word self reference (e.g. "creature", "saga",
	// "aura"); the runtime already carries the source-on-battlefield duration,
	// so only the trailing battlefield phrase needs verbatim confirmation.
	return rest == "remains on the battlefield." || rest == "is on the battlefield."
}

// exactControllerAmountEffectSyntax reconstructs a controller scry/surveil
// clause ("Scry 2.", "Surveil 1.") and its dynamic and prior-subject variants,
// comparing the result byte-for-byte against the printed clause. Beyond the
// fixed literal count it now also restores:
//   - a "where X is <count>" dynamic amount ("Scry X, where X is the number of
//     Zombies you control.");
//   - an explicit "You " subject ("Then you scry 2.") and a prior-subject
//     continuation joined by "then"/"and" ("…, then scry 1."), both of which
//     denote the same controller action.
//
// The dynamic count and prior-subject recipient are carried as typed fields
// (Amount and Context), so the lowering re-resolves them without reading the
// wording. A ranged "up to" amount has no scry/surveil form and stays inexact.
func exactControllerAmountEffectSyntax(effect *EffectSyntax, verb string) bool {
	switch effect.Context {
	case EffectContextController, EffectContextPriorSubject:
	default:
		return false
	}
	if effect.Amount.RangeKnown {
		return false
	}
	text := exactEffectClauseText(effect)
	for _, prefix := range []string{verb, "You " + verb} {
		if exactControllerAmountClauseText(text, prefix, effect.Amount, effectAmountSourceText(effect)) {
			return true
		}
	}
	return false
}

// exactControllerAmountClauseText reports whether the printed scry/surveil clause
// text matches the reconstruction for the given subject prefix and amount form.
// Comparisons are case-insensitive so the mid-sentence lowercase verb of a
// then-joined continuation ("scry 1.") matches the same reconstruction as the
// sentence-initial form ("Scry 1.").
func exactControllerAmountClauseText(text, prefix string, amount EffectAmountSyntax, amountText string) bool {
	switch amount.DynamicForm {
	case EffectDynamicAmountFormNone:
		if !amount.Known {
			return false
		}
		return strings.EqualFold(text, fmt.Sprintf("%s %s.", prefix, amountText))
	case EffectDynamicAmountFormWhereX:
		return strings.EqualFold(text, fmt.Sprintf("%s X, %s.", prefix, amount.Text))
	default:
		return false
	}
}

// exactDigLookEffectSyntax reconstructs the impulse look clause "Look at the top
// <number> cards of your library." and compares it byte-for-byte. It requires a
// fixed looked-at count of at least two (a dig looks at more cards than it
// takes), so variable ("X") and single-card forms fail closed.
func exactDigLookEffectSyntax(effect *EffectSyntax) bool {
	if effect.Context != EffectContextController || !effect.Amount.Known || effect.Amount.Value < 2 {
		return false
	}
	return strings.EqualFold(
		exactEffectClauseText(effect),
		fmt.Sprintf("Look at the top %s cards of your library.", effectAmountSourceText(effect)),
	)
}

// exactDigPutEffectSyntax reconstructs the impulse put clause "Put <number> <of
// them|of those cards> into your hand and the <rest|other> <into your
// graveyard|on the bottom of your library [in any order|in a random order]>."
// and compares it byte-for-byte. The structured fields come from parseDigPut; a
// fixed take count of one to three is required so variable forms fail closed.
func exactDigPutEffectSyntax(effect *EffectSyntax) bool {
	if !effect.Dig.Put || effect.Context != EffectContextController ||
		!effect.Amount.Known || effect.Amount.Value < 1 || effect.Amount.Value > 3 {
		return false
	}
	source := digSourceText(effect.Dig.Source)
	remainder := "rest"
	if effect.Dig.Singular {
		remainder = "other"
	}
	want := fmt.Sprintf(
		"Put %s%s into your hand and the %s %s.",
		effectAmountSourceText(effect), source, remainder, digRemainderText(effect.Dig.Remainder),
	)
	return strings.EqualFold(exactEffectClauseText(effect), want)
}

// exactPutThoseCardsIntoHandEffectSyntax reconstructs the "put a card from among
// those cards into your hand." consequence (Ripples of Undeath), where "those
// cards" denotes a card set produced by an earlier clause in the same ability
// (such as a preceding mill). It requires a controller-context single-card put
// into the hand carrying exactly one "those" pronoun reference, and matches the
// canonical clause text byte-for-byte; any filter or count variation fails closed
// so only the plain any-card form is recognized.
func exactPutThoseCardsIntoHandEffectSyntax(effect *EffectSyntax) bool {
	if effect.Context != EffectContextController ||
		effect.ToZone != zone.Hand ||
		!effect.Amount.Known || effect.Amount.Value != 1 {
		return false
	}
	if len(effect.References) != 1 ||
		effect.References[0].Kind != ReferencePronoun ||
		effect.References[0].Pronoun != PronounThose {
		return false
	}
	return strings.EqualFold(exactEffectClauseText(effect),
		"put a card from among those cards into your hand.")
}

// recorded, so the exactness gate can compare it byte-for-byte.
func digRemainderText(remainder DigRemainderKind) string {
	switch remainder {
	case DigRemainderLibraryBottom:
		return "on the bottom of your library"
	case DigRemainderLibraryBottomAny:
		return "on the bottom of your library in any order"
	case DigRemainderLibraryBottomRandom:
		return "on the bottom of your library in a random order"
	default:
		return "into your graveyard"
	}
}

// digSourceText renders the connector that links the impulse take count to the
// looked-at cards ("of them" or "of those cards"); an unset source yields the
// empty string so the exactness gate rejects the clause.
func digSourceText(source DigSourceKind) string {
	switch source {
	case DigSourceThem:
		return " of them"
	case DigSourceThoseCards:
		return " of those cards"
	default:
		return ""
	}
}

// exactConniveEffectSyntax reconstructs a connive keyword-action clause whose
// subject is the conniving permanent itself ("this creature connives.",
// "<this card's name> connives.", and the rarer numeric "<subject> connives N."
// form). It requires the source-scoped self subject and, when a count is
// printed, a fixed count of at least one; the variable form fails closed. The
// parenthetical reminder text is excluded from the parsed clause upstream.
func exactConniveEffectSyntax(effect *EffectSyntax) bool {
	if effect.Context != EffectContextSource {
		return false
	}
	subject, ok := exactSelfSubjectReferenceText(effect.SubjectReferences)
	if !ok {
		return false
	}
	text := exactEffectClauseText(effect)
	if !effect.Amount.Known {
		return strings.EqualFold(text, subject+" connives.")
	}
	if effect.Amount.Value < 1 {
		return false
	}
	return strings.EqualFold(text, fmt.Sprintf("%s connives %s.", subject, effectAmountSourceText(effect)))
}

// exactSourceExploresEffectSyntax recognizes the explore keyword action whose
// subject is the source permanent itself ("This creature explores." / "<name>
// explores."), the counterpart to the pronoun form "It explores." handled by
// exactDirectPronounEffectSyntax. The reminder text is excluded upstream. Only a
// bare explore is exact here; the repeated form ("explores, then it explores
// again.") and the variable form ("explores X times.") fail closed.
func exactSourceExploresEffectSyntax(effect *EffectSyntax) bool {
	if effect.Context != EffectContextSource ||
		len(effect.Targets) != 0 ||
		effect.Duration != EffectDurationNone {
		return false
	}
	subject, ok := exactSelfSubjectReferenceText(effect.SubjectReferences)
	if !ok {
		return false
	}
	return strings.EqualFold(exactEffectClauseText(effect), subject+" explores.")
}

// exactTargetExploresEffectSyntax recognizes the explore keyword action whose
// subject is a single target permanent ("Target creature you control
// explores."), the counterpart to the source forms handled by
// exactSourceExploresEffectSyntax. The reminder text is excluded upstream. Only
// a bare explore is exact; the repeated form ("explores, then it explores
// again.") fails closed.
func exactTargetExploresEffectSyntax(effect *EffectSyntax) bool {
	if effect.Context != EffectContextTarget ||
		effect.Duration != EffectDurationNone ||
		len(effect.Targets) != 1 ||
		!effect.Targets[0].Exact {
		return false
	}
	return strings.EqualFold(
		exactEffectClauseText(effect),
		effect.Targets[0].Text+" explores.",
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

// exactAmassEffectSyntax reconstructs an amass keyword-action clause ("Amass
// Orcs N." / "Amass Zombies N." / "Amass N.") and compares it byte-for-byte. It
// requires a fixed count of at least one, so variable ("X") forms fail closed.
// The subtype word printed in a typed clause must recognize to the parsed
// AmassSubtype; the untyped "Amass N." form is only exact for the default
// Zombie Army subtype.
func exactAmassEffectSyntax(effect *EffectSyntax) bool {
	if effect.Context != EffectContextController ||
		!effect.Amount.Known || effect.Amount.Value < 1 ||
		effect.AmassSubtype == "" {
		return false
	}
	amount := effectAmountSourceText(effect)
	text := exactEffectClauseText(effect)
	if strings.EqualFold(text, fmt.Sprintf("Amass %s.", amount)) {
		return effect.AmassSubtype == types.Zombie
	}
	subtype := amassSubtypeSourceText(effect)
	if subtype == "" {
		return false
	}
	return strings.EqualFold(text, fmt.Sprintf("Amass %s %s.", subtype, amount))
}

// amassSubtypeSourceText returns the printed subtype word that follows the
// "amass" verb when it recognizes to the parsed AmassSubtype, or "" when no such
// word is present (the untyped "Amass N." form).
func amassSubtypeSourceText(effect *EffectSyntax) string {
	for i, token := range effect.Tokens {
		if !equalWord(token, "amass") {
			continue
		}
		if i+1 < len(effect.Tokens) && effect.Tokens[i+1].Kind == shared.Word {
			word := effect.Tokens[i+1]
			if sub, ok := recognizeSubtypePhrase(word.Text); ok && sub == effect.AmassSubtype {
				return word.Text
			}
		}
		return ""
	}
	return ""
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
		if amount.RangeKnown {
			noun := plural
			if amount.Maximum == 1 {
				noun = singular
			}
			return strings.EqualFold(text, fmt.Sprintf("%s up to %s %s.", prefix, amountText, noun))
		}
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
	case EffectDynamicAmountFormHalfLibrary:
		// The half-library amount's noun is the milling player's library, carried
		// whole in amount.Text ("half their library, rounded down"), so the clause
		// reconstructs as the bare subject verb followed by that phrase with no
		// counted "cards" noun: "Target player mills half their library, rounded
		// down."
		return strings.EqualFold(text, fmt.Sprintf("%s %s.", prefix, amount.Text))
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
		if s, ok := exactObjectReferenceText(modifyPTSubjectReferences(effect)); ok {
			subject = titleFirstEffectText(s)
		} else {
			subject = "It"
		}
	case EffectContextSource:
		s, ok := exactSelfSubjectReferenceText(modifyPTSubjectReferences(effect))
		if !ok {
			return false
		}
		subject = s
	case EffectContextPriorSubject:
		return exactPriorSubjectGroupModifyPTEffectSyntax(effect)
	default:
		return false
	}
	power := signedPTSideText(effect.PowerDelta)
	toughness := signedPTSideText(effect.ToughnessDelta)
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
		// The "an additional" wording introduces a second, stacking pump on a
		// creature an earlier clause already modified ("Target creature gets
		// -1/-1 until end of turn. It gets an additional -1/-1 until end of turn
		// for each card named Compound Fracture in your graveyard.", Compound
		// Fracture; Growth Cycle). Mechanically it is a plain dynamic pump, so
		// reconstruct both the bare and "an additional" verb phrasings.
		for _, verb := range []string{"gets", "gets an additional"} {
			if strings.EqualFold(text, fmt.Sprintf("%s %s %s/%s %s until end of turn.", subject, verb, power, toughness, effect.Amount.Text)) ||
				strings.EqualFold(text, fmt.Sprintf("%s %s %s/%s until end of turn %s.", subject, verb, power, toughness, effect.Amount.Text)) {
				return true
			}
		}
		return false
	case EffectDynamicAmountFormWhereX:
		powerX := signedPTSideText(effect.PowerDelta)
		toughnessX := signedPTSideText(effect.ToughnessDelta)
		return strings.EqualFold(text, fmt.Sprintf("%s gets %s/%s until end of turn, %s.", subject, powerX, toughnessX, effect.Amount.Text))
	default:
		return false
	}
}

func exactGroupModifyPTEffectSyntax(effect *EffectSyntax) bool {
	var subject []shared.Token
	for i := range effect.Tokens {
		if spanCovers(effect.StaticSubject.Span, effect.Tokens[i].Span) {
			subject = append(subject, effect.Tokens[i])
		}
	}
	if len(subject) == 0 {
		return false
	}
	// The distributive "each creature" subject takes the singular verb ("Each
	// creature gets ..."), unlike the plural "all creatures get ..."; pick the
	// verb form from the subject so the round-trip reconstructs the source text.
	verb := "get"
	if equalWord(subject[0], "each") {
		verb = "gets"
	}
	prefix := fmt.Sprintf(
		"%s %s %s/%s",
		joinedEffectText(subject),
		verb,
		signedPTSideText(effect.PowerDelta),
		signedPTSideText(effect.ToughnessDelta),
	)
	return exactGroupModifyPTBody(effect, prefix)
}

// exactPriorSubjectGroupModifyPTEffectSyntax recognizes the modify clause of a
// keyword-first mass pump ("Creatures you control gain trample and get +X/+X
// until end of turn, where X is …"). The preceding keyword clause names the
// affected group, so this clause inherits that subject (EffectContextPriorSubject)
// and reads "get <p>/<t> …" with no subject prefix.
func exactPriorSubjectGroupModifyPTEffectSyntax(effect *EffectSyntax) bool {
	prefix := fmt.Sprintf(
		"get %s/%s",
		signedPTSideText(effect.PowerDelta),
		signedPTSideText(effect.ToughnessDelta),
	)
	return exactGroupModifyPTBody(effect, prefix)
}

// exactGroupModifyPTBody matches the until-end-of-turn body of a group
// power/toughness change against prefix (the reconstructed "<subject> get
// <p>/<t>" or, for a prior-subject clause, "get <p>/<t>"). It accepts the bare
// fixed form, the keyword-split fixed form (no duration suffix, spread from a
// sibling clause), and the two dynamic-amount shapes ("… for each …" and "…
// where X is …") so both standalone and conjoined mass pumps with a fixed or
// dynamic amount are recognized.
func exactGroupModifyPTBody(effect *EffectSyntax, prefix string) bool {
	text := exactEffectClauseText(effect)
	if effect.Amount.DynamicKind == EffectDynamicAmountNone {
		if strings.EqualFold(text, prefix+" until end of turn.") {
			return true
		}
		return effect.Duration == EffectDurationUntilEndOfTurn &&
			strings.EqualFold(text, prefix+".")
	}
	switch effect.Amount.DynamicForm {
	case EffectDynamicAmountFormForEach:
		if strings.EqualFold(text, fmt.Sprintf("%s %s until end of turn.", prefix, effect.Amount.Text)) ||
			strings.EqualFold(text, fmt.Sprintf("%s until end of turn %s.", prefix, effect.Amount.Text)) {
			return true
		}
		// A sentence-leading "Until end of turn," supplies the duration, so the
		// clause itself carries no suffix ("Until end of turn, creatures you
		// control … get +N/+N for each …").
		return effect.Duration == EffectDurationUntilEndOfTurn &&
			strings.EqualFold(text, fmt.Sprintf("%s %s.", prefix, effect.Amount.Text))
	case EffectDynamicAmountFormWhereX:
		if strings.EqualFold(text, fmt.Sprintf("%s until end of turn, %s.", prefix, effect.Amount.Text)) {
			return true
		}
		return effect.Duration == EffectDurationUntilEndOfTurn &&
			strings.EqualFold(text, fmt.Sprintf("%s, %s.", prefix, effect.Amount.Text))
	default:
		return false
	}
}

// exactMoveCountersEffectSyntax recognizes the supported counter-movement form:
// moving counters off the source permanent onto a single target permanent
// ("Move a +1/+1 counter from this creature onto target creature.", "Move all
// counters from this permanent onto target creature."). The source is a single
// self reference (the effect's own permanent — "this <object>" or the card's
// own name), the destination is the single exact target, and the move is either
// a single specific-kind counter (CounterKnown, Amount one) or the kind-agnostic
// "all counters" form (MoveCountersAll). The "onto other creatures" group
// distribution (Forgotten Ancient) carries no target and stays unrecognized.
func exactMoveCountersEffectSyntax(effect *EffectSyntax) bool {
	if effect.MoveCountersDistribute {
		return exactMoveCountersDistributeEffectSyntax(effect)
	}
	if effect.MoveCountersFromTarget {
		return exactMoveCountersFromTargetEffectSyntax(effect)
	}
	if len(effect.Targets) != 1 ||
		!effect.Targets[0].Exact ||
		effect.Targets[0].Cardinality.Max != 1 {
		return false
	}
	source, ok := exactSelfSubjectReferenceText(effect.References)
	if !ok {
		return false
	}
	dest := effect.Targets[0].Text
	text := exactEffectClauseText(effect)
	if effect.MoveCountersAll {
		return strings.EqualFold(
			text,
			fmt.Sprintf("Move all counters from %s onto %s.", source, dest),
		)
	}
	if !effect.CounterKnown || !effect.Amount.Known || effect.Amount.Value != 1 {
		return false
	}
	return strings.EqualFold(
		text,
		fmt.Sprintf("Move %s %s counter from %s onto %s.",
			effectAmountSourceText(effect), effect.CounterKind.String(), source, dest),
	)
}

// exactRemoveCounterEffectSyntax recognizes the resolving counter-removal effect
// "Remove <amount> [<kind> ]counter(s) from <object>." (Ferropede, "remove a
// counter from target permanent."; Thrull Parasite, "Remove a counter from
// target nonland permanent."; "Remove two +1/+1 counters from target creature").
// The object is a single recognized target permanent (single or "up to one"
// cardinality). The count is a fixed positive amount; the counter is either a
// named recognized kind or, in the kind-unspecified "a counter" form, a counter
// of any kind the controller chooses at resolution. The clause is reconstructed
// and matched byte-exact, so the mass "all counters" form, dynamic counts, and
// any referenced or pronoun-object shape stay non-exact and fail closed.
func exactRemoveCounterEffectSyntax(effect *EffectSyntax) bool {
	if !effect.Amount.Known || effect.Amount.Value < 1 {
		return false
	}
	// The kind-unspecified "a counter" form removes one counter of a single
	// controller-chosen kind, so a plural unspecified count ("two counters") has
	// no single-choice resolution and is left non-exact to fail closed.
	if !effect.CounterKnown && effect.Amount.Value != 1 {
		return false
	}
	object, ok := exactRemoveCounterObjectText(effect)
	if !ok {
		return false
	}
	noun := "counters"
	if effect.Amount.Value == 1 {
		noun = "counter"
	}
	text := exactEffectClauseText(effect)
	if effect.CounterKnown {
		return strings.EqualFold(
			text,
			fmt.Sprintf("Remove %s %s %s from %s.",
				effectAmountSourceText(effect), effect.CounterKind.String(), noun, object),
		)
	}
	return strings.EqualFold(
		text,
		fmt.Sprintf("Remove %s %s from %s.",
			effectAmountSourceText(effect), noun, object),
	)
}

// exactRemoveAllCountersEffectSyntax recognizes the kind-agnostic mass removal
// "Remove all counters from <object>." (Vampire Hexmage, "Remove all counters
// from target permanent."). The object is a single recognized target permanent
// or a lone source/self reference, reusing exactRemoveCounterObjectText. The
// clause is reconstructed and matched byte-exact, so the fixed-count, dynamic,
// and kind-specific removals stay out of this path and lower through their own
// recognizers.
func exactRemoveAllCountersEffectSyntax(effect *EffectSyntax) bool {
	if !effect.RemoveCountersAll || effect.CounterKnown {
		return false
	}
	object, ok := exactRemoveCounterObjectText(effect)
	if !ok {
		return false
	}
	return strings.EqualFold(
		exactEffectClauseText(effect),
		fmt.Sprintf("Remove all counters from %s.", object),
	)
}

// exactRemoveCounterObjectText returns the rendered object a counter is removed
// from, for either a single exact target permanent ("target creature") or a lone
// source/self reference ("this creature"/"this artifact"/the card's own name, as
// in "Remove a -1/-1 counter from this creature."). It fails closed for every
// other shape — multiple or inexact targets, a targeted cardinality above one, a
// group ("each creature you control"), or a "from it" pronoun paired with delayed
// timing — so unrepresentable removals keep the wording unsupported.
func exactRemoveCounterObjectText(effect *EffectSyntax) (string, bool) {
	switch {
	case len(effect.Targets) == 1 && len(effect.References) == 0:
		if !effect.Targets[0].Exact || effect.Targets[0].Cardinality.Max != 1 {
			return "", false
		}
		return effect.Targets[0].Text, true
	case len(effect.Targets) == 0 && len(effect.References) == 1:
		if object, ok := exactObjectReferenceText(effect.References); ok {
			return object, true
		}
		return exactSelfSubjectReferenceText(effect.References)
	}
	return "", false
}

// exactPutThoseCountersEffectSyntax recognizes the counter-salvage form "put
// those counters on <destination>" and its singular-pronoun variant "put its
// counters on <destination>", where "those"/"its" name the counters a
// triggering permanent had as it left a zone ("Whenever a creature you control
// leaves the battlefield, if it had counters on it, put those counters on target
// creature you control.", "When this creature dies, put its counters on target
// creature you control."). The destination is either a single/optional exact
// target permanent or the source permanent itself ("this <object>" or the card's
// own name). The clause is reconstructed and matched byte-exact so any
// unrepresentable destination keeps the wording unsupported.
func exactPutThoseCountersEffectSyntax(effect *EffectSyntax) bool {
	if !effect.MoveThoseCounters {
		return false
	}
	text := exactEffectClauseText(effect)
	if len(effect.Targets) == 1 {
		if !effect.Targets[0].Exact || effect.Targets[0].Cardinality.Max != 1 {
			return false
		}
		return putThoseCountersClauseMatches(text, effect.Targets[0].Text)
	}
	if len(effect.Targets) != 0 {
		return false
	}
	dest, ok := putThoseCountersSelfText(effect.References)
	if !ok {
		return false
	}
	return putThoseCountersClauseMatches(text, dest)
}

// putThoseCountersClauseMatches reports whether text is the kind-agnostic
// counter-salvage placement "Put those counters on <dest>." or its
// singular-pronoun variant "Put its counters on <dest>." Both pronouns name a
// triggering permanent's counters for the same salvage move, so either wording
// round-trips to the same MoveThoseCounters effect.
func putThoseCountersClauseMatches(text, dest string) bool {
	for _, pronoun := range []string{"those", "its"} {
		if strings.EqualFold(text, fmt.Sprintf("Put %s counters on %s.", pronoun, dest)) {
			return true
		}
	}
	return false
}

// putThoseCountersSelfText returns the rendered text of a self destination for
// the counter-salvage form when the references carry exactly one source
// self-reference ("this <object>" or the card's own name) alongside the
// salvage's "it"/"those" back-references. It fails closed when no self
// reference, or more than one, is present.
func putThoseCountersSelfText(references []Reference) (string, bool) {
	text := ""
	count := 0
	for _, reference := range references {
		if reference.Kind == ReferenceThisObject || reference.Kind == ReferenceSelfName {
			text = joinedEffectText(reference.Tokens)
			count++
		}
	}
	if count != 1 {
		return "", false
	}
	return text, true
}

// exactMoveCountersFromTargetEffectSyntax recognizes the two-target counter-move
// form, where the counters are read from a first chosen target permanent and
// placed onto a second chosen target permanent ("Move a counter from target
// permanent you control onto a second target permanent." — Nesting Grounds,
// "Move a +1/+1 counter from target creature onto a second target creature." —
// Daghatar, "Move all counters from target creature onto another target
// creature." — Fate Transfer). The source target is the first exact single
// permanent target and the destination is the second; both selections must be
// exactly representable. The moved counter is one named-kind counter, the
// kind-agnostic "all counters" form, or one counter of a kind the controller
// chooses ("a counter"). It reconstructs the full clause from typed pieces and
// accepts only an exact round-trip, with the destination's "a second" determiner
// admitted alongside the bare and "another"/"other" determiners, so any
// unrepresentable wording (a relational "with the same controller" destination,
// a fixed count other than one) keeps the effect inexact.
func exactMoveCountersFromTargetEffectSyntax(effect *EffectSyntax) bool {
	single := TargetCardinalitySyntax{Min: 1, Max: 1}
	if len(effect.Targets) != 2 ||
		!effect.Targets[0].Exact || effect.Targets[0].Cardinality != single ||
		!effect.Targets[1].Exact || effect.Targets[1].Cardinality != single {
		return false
	}
	source, ok := exactPermanentTargetText(effect.Targets[0].Selection)
	if !ok {
		return false
	}
	dest, ok := exactPermanentTargetText(effect.Targets[1].Selection)
	if !ok {
		return false
	}
	var kindPhrase string
	switch {
	case effect.MoveCountersAll:
		kindPhrase = "all counters"
	case effect.CounterKnown:
		if !effect.Amount.Known || effect.Amount.Value != 1 {
			return false
		}
		kindPhrase = fmt.Sprintf("%s %s counter",
			effectAmountSourceText(effect), effect.CounterKind.String())
	default:
		if effect.Amount.Known && effect.Amount.Value != 1 {
			return false
		}
		kindPhrase = fmt.Sprintf("%s counter", effectAmountSourceText(effect))
	}
	destForms := []string{dest}
	destSelection := effect.Targets[1].Selection
	if !destSelection.Another && !destSelection.Other {
		destForms = append(destForms, "a second "+dest)
	}
	text := exactEffectClauseText(effect)
	for _, destForm := range destForms {
		if strings.EqualFold(text,
			fmt.Sprintf("Move %s from %s onto %s.", kindPhrase, source, destForm)) {
			return true
		}
	}
	return false
}

// exactMoveCountersDistributeEffectSyntax recognizes the "move any number of
// <kind> counters from <source> onto other creatures" form (Forgotten Ancient),
// where the controller distributes the source's counters among a group of other
// creatures rather than a single target. The source is a single self reference,
// the destination is the "other creatures" group, the counter kind is known, and
// the effect carries no target.
func exactMoveCountersDistributeEffectSyntax(effect *EffectSyntax) bool {
	if len(effect.Targets) != 0 || !effect.CounterKnown {
		return false
	}
	source, ok := exactSelfSubjectReferenceText(effect.References)
	if !ok {
		return false
	}
	group, ok := exactMoveCountersDistributeGroupText(effect.Selection)
	if !ok {
		return false
	}
	return strings.EqualFold(
		exactEffectClauseText(effect),
		fmt.Sprintf("Move any number of %s counters from %s onto %s.",
			effect.CounterKind.String(), source, group),
	)
}

// exactMoveCountersDistributeGroupText renders the destination phrase for the
// distributed move-counters form. It recognizes only the bare "other creatures"
// group (creature kind, "other" qualifier, no other selector qualifiers) and
// fails closed for every other shape so an unrepresentable group keeps the
// wording unsupported.
func exactMoveCountersDistributeGroupText(selection SelectionSyntax) (string, bool) {
	if selection.Kind != SelectionCreature || !selection.Other {
		return "", false
	}
	if selection.All || selection.Another || selection.Controller != SelectionControllerAny ||
		selection.Attacking || selection.Blocking || selection.Tapped || selection.Untapped ||
		selection.NonToken || selection.TokenOnly || selection.Colorless || selection.Multicolored ||
		selection.BasicLandType || selection.MatchManaValue || selection.MatchPower ||
		selection.MatchToughness || selection.CounterRequired ||
		selection.Keyword != KeywordUnknown || selection.ExcludedKeyword != KeywordUnknown ||
		selection.Zone != zone.None ||
		len(selection.ExcludedTypes) != 0 || len(selection.Supertypes) != 0 ||
		len(selection.ExcludedSupertypes) != 0 || len(selection.ColorsAny) != 0 ||
		len(selection.ExcludedColors) != 0 || len(selection.SubtypesAny) != 0 ||
		len(selection.ExcludedSubtypes) != 0 || len(selection.Alternatives) != 0 {
		return "", false
	}
	for _, required := range selection.RequiredTypesAny {
		if required != CardTypeCreature {
			return "", false
		}
	}
	return "other creatures", true
}

func exactCounterPlacementEffectSyntax(effect *EffectSyntax) bool {
	if !effect.CounterKnown && len(effect.CounterKindChoices) < 2 {
		return false
	}
	objects := []string{}
	switch {
	case len(effect.Targets) == 1 && (effect.Targets[0].Exact || counterPlacementTargetIsPlayerControls(effect.Targets[0])):
		// A "<group> target <player|opponent> controls" placement reconstructs
		// the recipient group jointly with its targeted player; the targeted
		// player supplies the group's controller and does not receive the
		// counter itself. Try that reconstruction first, then fall back to the
		// standard exact single target (which covers a player-counter placement
		// such as "Put a poison counter on target player.", whose recipient is
		// the targeted player).
		if object, ok := counterPlacementTargetPlayerControlsObject(effect); ok {
			objects = append(objects, object)
		} else if effect.Targets[0].Exact {
			object := effect.Targets[0].Text
			// "Put a +1/+1 counter on each of up to two target creatures."
			// places one counter on each of several targets, so the canonical
			// object reads "each of <target>" for any genuine multi-target
			// cardinality (Max >= 2). The singular and "up to one" forms keep
			// the bare target text.
			if effect.Targets[0].Cardinality.Max >= 2 {
				object = "each of " + object
			}
			objects = append(objects, object)
		} else {
			return false
		}
	case len(effect.Targets) == 0:
		if effect.CounterRecipientAttached {
			// The attached recipient is "enchanted creature" (Aura) or
			// "equipped creature" (Equipment); both lower to the source's
			// attached-permanent reference, so offer both candidates and let the
			// byte-exact text match select the one the source printed.
			objects = append(objects, "enchanted creature", "equipped creature")
			break
		}
		// A trailing dynamic count ("… where X is the number of +1/+1 counters
		// on this creature") carries its own referent inside the amount span;
		// that referent names the counted subject, not the placement recipient,
		// so exclude it before reconstructing the recipient.
		recipientRefs := referencesOutsideSpan(effect.References, effect.Amount.Span)
		// A filtered group recipient with a "with a <kind> counter on it/them"
		// qualifier ("each creature you control with a +1/+1 counter on it")
		// carries a trailing pronoun referent inside its own selection span that
		// names the filtered permanent, not the recipient. Exclude it only for
		// that qualifier so a recipient that genuinely is a referenced object
		// ("Put a +1/+1 counter on this creature.") keeps its referent.
		if effect.Selection.CounterRequired || effect.Selection.CounterAny {
			recipientRefs = referencesOutsideSpan(recipientRefs, effect.Selection.Span)
		}
		if object, ok := exactObjectReferenceText(recipientRefs); ok {
			objects = append(objects, object)
		} else if object, ok := exactSelfSubjectReferenceText(recipientRefs); ok {
			objects = append(objects, object)
		} else if len(recipientRefs) == 0 {
			// A non-target recipient is either a group ("each creature you
			// control") or a single chooser ("a creature you control"); try both
			// reconstructions and accept whichever matches the source text.
			if object, ok := exactGroupDamagePermanentRecipientText(effect.Selection); ok {
				objects = append(objects, object)
			}
			if object, ok := exactSingularChosenPermanentRecipientText(effect.Selection); ok {
				objects = append(objects, object)
			}
		}
		if len(objects) == 0 {
			return false
		}
	default:
		return false
	}
	for _, object := range objects {
		if counterPlacementTextMatches(effect, object) {
			return true
		}
		if len(effect.CounterKindChoices) >= 2 && counterPlacementChoiceTextMatches(effect, object) {
			return true
		}
	}
	return false
}

// counterPlacementTargetIsPlayerControls reports whether a counter-placement
// target is the bare single player or opponent of a "<group> target <player|
// opponent> controls" group recipient. The targeted player supplies the
// recipient group's controller relationship rather than receiving the counter
// itself, so this target is reconstructed jointly with the recipient group
// (counterPlacementTargetPlayerControlsObject) rather than through the standard
// exact-target path.
func counterPlacementTargetIsPlayerControls(target TargetSyntax) bool {
	switch target.Selection.Kind {
	case SelectionPlayer, SelectionOpponent:
	default:
		return false
	}
	return target.Cardinality.Min == 1 && target.Cardinality.Max == 1 &&
		!target.Selection.Other &&
		target.Selection.Controller == SelectionControllerAny
}

// counterPlacementTargetPlayerControlsObject reconstructs the recipient phrase
// for a group counter placement whose group is every permanent a single targeted
// player controls ("each creature target player controls", Meadowboon; "each
// creature target opponent controls"). The targeted player is the effect's sole
// target and carries no qualifier of its own; the recipient permanent group is
// the effect's selection with no controller of its own (the controller
// relationship is supplied by the targeted player). It returns the combined
// "<group> target <player|opponent> controls" object only for the bare single
// player or opponent target; every other shape fails closed so the byte-exact
// round-trip keeps unrepresentable wordings unsupported.
func counterPlacementTargetPlayerControlsObject(effect *EffectSyntax) (string, bool) {
	target := effect.Targets[0]
	if target.Cardinality.Min != 1 || target.Cardinality.Max != 1 {
		return "", false
	}
	var controlsPhrase string
	switch target.Selection.Kind {
	case SelectionPlayer:
		controlsPhrase = "target player controls"
	case SelectionOpponent:
		controlsPhrase = "target opponent controls"
	default:
		return "", false
	}
	if target.Selection.Other || target.Selection.Controller != SelectionControllerAny {
		return "", false
	}
	if effect.Selection.Controller != SelectionControllerAny {
		return "", false
	}
	group, ok := exactGroupDamagePermanentRecipientText(effect.Selection)
	if !ok {
		return "", false
	}
	return group + " " + controlsPhrase, true
}

// "Distribute N <kind> counters among <cardinality> target creatures" form: a
// fixed (or X) total of counters split among the chosen targets, at least one
// each, the counter analog of divided damage. It is detected by the same
// byte-exact reconstruction the exactness gate uses, so the parser sets the
// DistributeCounters flag only for wordings the executable backend lowers.
func distributeCountersEffect(effect *EffectSyntax) bool {
	return effect.Kind == EffectPut && exactDistributeCountersEffectSyntax(effect)
}

// exactDistributeCountersEffectSyntax reconstructs the canonical "Distribute
// <amount> <kind> counters among <cardinality> target creatures[ you control]."
// clause and compares it byte-for-byte to the source. It supports a fixed total
// of at least one or the spell's variable X, the enumerated and "any number of"
// cardinalities divided damage recognizes, and a plain creature target
// optionally restricted to "you control". Every other shape fails closed,
// leaving the byte-exact round-trip to reject the wording.
func exactDistributeCountersEffectSyntax(effect *EffectSyntax) bool {
	if !effect.CounterKnown || effect.Negated || effect.Optional || len(effect.Targets) != 1 {
		return false
	}
	amountText, ok := distributeCountersAmountText(effect.Amount)
	if !ok {
		return false
	}
	cardinality, ok := dividedCardinalityPhrase(effect.Targets[0].Cardinality)
	if !ok {
		return false
	}
	noun, ok := distributeCountersTargetNoun(effect.Targets[0].Selection)
	if !ok {
		return false
	}
	expected := fmt.Sprintf("Distribute %s %s counters among %s %s.",
		amountText, effect.CounterKind.String(), cardinality, noun)
	return strings.EqualFold(exactEffectClauseText(effect), expected)
}

// distributeCountersAmountText reconstructs the canonical amount token for a
// distribute counters clause: the spelled-out cardinal word for a fixed total of
// at least one (Oracle wording writes "Distribute three +1/+1 counters", not the
// digit form), or "X" for the spell's bare variable X. It fails closed for a
// non-positive or out-of-range fixed total and for every dynamic amount form, so
// those wordings keep failing the round-trip.
func distributeCountersAmountText(amount EffectAmountSyntax) (string, bool) {
	if amount.DynamicForm != EffectDynamicAmountFormNone ||
		amount.DynamicKind != EffectDynamicAmountNone ||
		amount.Addend != 0 || amount.Multiplier != 0 {
		return "", false
	}
	switch {
	case amount.Known:
		return cardinalWord(amount.Value)
	case amount.VariableX:
		return "X", true
	default:
		return "", false
	}
}

// counters effect splits among. It supports the plain "target creatures" and the
// "target creatures you control" controller restriction, failing closed for
// every other selector the distributed counter placement does not model.
func distributeCountersTargetNoun(selection SelectionSyntax) (string, bool) {
	if selection.Kind != SelectionCreature {
		return "", false
	}
	if selection.All || selection.Another || selection.Other ||
		selection.Tapped || selection.Untapped ||
		selection.Colorless || selection.Multicolored ||
		selection.MatchManaValue || selection.MatchPower || selection.MatchToughness ||
		selection.PowerLessThanSource || selection.PowerGreaterThanSource ||
		selection.TokenOnly || selection.NonToken ||
		selection.Attacking || selection.Blocking ||
		selection.Zone != zone.None ||
		len(selection.ExcludedTypes) != 0 ||
		len(selection.ExcludedColors) != 0 ||
		len(selection.ColorsAny) != 0 ||
		len(selection.SubtypesAny) != 0 ||
		len(selection.Supertypes) != 0 ||
		len(selection.ExcludedSupertypes) != 0 ||
		len(selection.ExcludedSubtypes) != 0 {
		return "", false
	}
	if required := selection.RequiredTypesAny; len(required) > 1 ||
		(len(required) == 1 && required[0] != CardTypeCreature) {
		return "", false
	}
	switch selection.Controller {
	case SelectionControllerAny:
		return "target creatures", true
	case SelectionControllerYou:
		return "target creatures you control", true
	default:
		return "", false
	}
}

// counterPlacementChoiceTextMatches reconstructs the controller-choice counter
// placement clause "Put a <X> counter or a <Y> counter on <object>." from the
// recognized choice kinds and reports whether the printed effect text matches it
// byte-for-byte. Each named kind is a single counter ("a <kind> counter") and the
// kinds are joined with " or " in source order; any richer amount or wording
// fails closed because the canonical text would not match.
func counterPlacementChoiceTextMatches(effect *EffectSyntax, object string) bool {
	if !effect.Amount.Known || effect.Amount.Value != 1 {
		return false
	}
	parts := make([]string, 0, len(effect.CounterKindChoices))
	for _, kind := range effect.CounterKindChoices {
		parts = append(parts, "a "+kind.String()+" counter")
	}
	prefix := "Put " + strings.Join(parts, " or ") + " on " + object
	return strings.EqualFold(exactEffectClauseText(effect), prefix+".")
}

// counterPlacementSingleChoiceRecipient reports whether an exact non-target
// counter placement names a single chosen group member ("a creature you
// control") rather than a distributive group ("each creature you control"). The
// two forms compile to identical selectors, so lowering relies on this flag to
// emit a single-choice placement instead of a group placement.
func counterPlacementSingleChoiceRecipient(effect *EffectSyntax) bool {
	if effect.Kind != EffectPut || !effect.CounterKnown || len(effect.Targets) != 0 {
		return false
	}
	if effect.CounterRecipientAttached {
		return false
	}
	object, ok := exactSingularChosenPermanentRecipientText(effect.Selection)
	if !ok {
		return false
	}
	return counterPlacementTextMatches(effect, object)
}

func counterPlacementTextMatches(effect *EffectSyntax, object string) bool {
	noun := "counters"
	if effect.Amount.Known && effect.Amount.Value == 1 {
		noun = "counter"
	}
	text := exactEffectClauseText(effect)
	// The "equal to <amount>" form ("Put a number of +1/+1 counters on it equal
	// to the amount of life you gained this turn …") states the dynamic count as
	// a trailing "a number of … <amount>" clause rather than a leading numeral,
	// so its canonical text places the amount phrase after the object.
	if effect.Amount.DynamicForm == EffectDynamicAmountFormEqual {
		equalPrefix := fmt.Sprintf("Put a number of %s counters on %s", effect.CounterKind.String(), object)
		return strings.EqualFold(text, equalPrefix+" "+effect.Amount.Text+".")
	}
	// The "for each" form ("Put a +1/+1 counter on target creature for each Elf
	// you control.") places one counter per counted object and states its count
	// as a trailing "for each <iterator>" clause that the amount captured
	// verbatim. Only the multiplier-one form prints the bare "a <kind> counter"
	// count word, so a richer multiplier or an unrecognized iterator fails
	// closed.
	if effect.Amount.DynamicForm == EffectDynamicAmountFormForEach {
		if effect.Amount.DynamicKind == EffectDynamicAmountNone || effect.Amount.Multiplier != 1 {
			return false
		}
		forEachPrefix := fmt.Sprintf("Put a %s counter on %s", effect.CounterKind.String(), object)
		return strings.EqualFold(text, forEachPrefix+" "+effect.Amount.Text+".")
	}
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

// exactGainPlayerCounterEffectSyntax recognizes the player-counter gain effect
// "You get {E}…{E}." (energy symbols) or "<recipient> gets <N> <kind> counter(s)."
// (a named player-only counter). The optional "(N energy counters)" reminder is
// already stripped. The energy form is controller-only; the named form supports
// the controller plus the defending player, the triggering "that player", a
// single targeted player, and the "each opponent"/"each player" groups, mirroring
// the recipients exactLifeEffectSyntax accepts.
// Every other recipient or richer form fails closed; the lowering re-resolves the
// recipient from the same typed context.
func exactGainPlayerCounterEffectSyntax(effect *EffectSyntax) bool {
	if !effect.Amount.Known || effect.Amount.Value < 1 {
		return false
	}
	if effect.Context == EffectContextController &&
		len(effect.Targets) == 0 && len(effect.References) == 0 {
		verb := slices.IndexFunc(effect.Tokens, func(token shared.Token) bool {
			return token.Span == effect.VerbSpan
		})
		if verb >= 0 {
			rest := effect.Tokens[verb+1:]
			if len(rest) > 0 && rest[len(rest)-1].Kind == shared.Period {
				rest = rest[:len(rest)-1]
			}
			if len(rest) == effect.Amount.Value && allEnergySymbols(rest) {
				return true
			}
		}
	}
	if !effect.CounterKnown || !effect.CounterKind.PlayerOnly() {
		return false
	}
	noun := effect.CounterKind.String()
	return exactPlayerCounterRecipientText(effect, noun+" counter", noun+" counters")
}

// exactPlayerCounterRecipientText reconstructs the named player-counter gain
// clause "<recipient> gets <N> <singular|plural>." for each supported recipient
// subject and reports whether the printed effect text matches byte-for-byte. The
// recipient set matches the references the lowering resolves (controller,
// defending player, the triggering "that player"/"they", the referenced object's
// controller "Its controller", a lone targeted player, and the "each
// opponent"/"each player" groups); any other subject yields no prefix and fails
// closed.
func exactPlayerCounterRecipientText(effect *EffectSyntax, singular, plural string) bool {
	var prefixes []string
	switch effect.Context {
	case EffectContextController:
		prefixes = []string{"You get"}
	case EffectContextEachOpponent:
		prefixes = []string{"Each opponent gets"}
	case EffectContextEachOtherPlayer:
		prefixes = []string{"Each other player gets"}
	case EffectContextEachPlayer:
		prefixes = []string{"Each player gets"}
	case EffectContextDefendingPlayer:
		prefixes = []string{"Defending player gets"}
	case EffectContextEventPlayer, EffectContextReferencedPlayer:
		prefixes = []string{"They get", "That player gets"}
	case EffectContextReferencedObjectController:
		if subject := referencedControllerSubjectText(effect); subject != "" {
			prefixes = []string{subject + " gets"}
		}
	case EffectContextTarget, EffectContextPriorSubject:
		if len(effect.Targets) == 1 && effect.Targets[0].Exact {
			prefixes = []string{titleFirstEffectText(effect.Targets[0].Text) + " gets"}
		}
	default:
	}
	text := exactEffectClauseText(effect)
	amountText := effectAmountSourceText(effect)
	for _, prefix := range prefixes {
		if exactCountedNounEffectText(text, prefix, singular, plural, effect.Amount, amountText, false) {
			return true
		}
	}
	return false
}
