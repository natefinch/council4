package cardgen

import (
	"fmt"
	"slices"
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func lowerEnchantAbility(
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) (game.StaticAbility, bool, *shared.Diagnostic) {
	if len(ability.Content.Keywords) != 1 || ability.Content.Keywords[0].Kind != parser.KeywordEnchant {
		return game.StaticAbility{}, false, nil
	}
	keyword := ability.Content.Keywords[0]
	target, ok := enchantTargetSpec(keyword.EnchantTarget)
	if !ok ||
		ability.Kind != compiler.AbilityStatic ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		len(ability.Content.Effects) != 0 ||
		len(ability.Content.References) != 0 ||
		ability.AbilityWord != "" {
		return game.StaticAbility{}, true, executableDiagnostic(
			ability,
			"unsupported Enchant ability",
			"the executable source backend supports only exact Enchant with a supported target kind",
		)
	}
	if !keywordOnlyCovered(syntax, keyword) {
		return game.StaticAbility{}, true, executableDiagnostic(
			ability,
			"unsupported Enchant ability",
			"the executable source backend supports only exact Enchant with a supported target kind",
		)
	}
	return game.EnchantStaticAbility(&target), true, nil
}

func lowerProtectionAbility(
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) (game.StaticAbility, bool, *shared.Diagnostic) {
	if len(ability.Content.Keywords) != 1 || ability.Content.Keywords[0].Kind != parser.KeywordProtection {
		return game.StaticAbility{}, false, nil
	}
	// If the ability has effects, it is a grant (e.g., "Enchanted creature has
	// protection from X") — defer to Static Declaration lowering instead.
	if len(ability.Content.Effects) > 0 {
		return game.StaticAbility{}, false, nil
	}
	keyword := ability.Content.Keywords[0]

	// Common structural checks for all protection variants.
	structureOK := ability.Kind == compiler.AbilityStatic &&
		ability.Cost == nil &&
		ability.Trigger == nil &&
		len(ability.Content.Targets) == 0 &&
		len(ability.Content.Conditions) == 0 &&
		len(ability.Content.Effects) == 0 &&
		len(ability.Content.References) == 0 &&
		ability.AbilityWord == ""

	unsupported := func() (game.StaticAbility, bool, *shared.Diagnostic) {
		return game.StaticAbility{}, true, executableDiagnostic(
			ability,
			"unsupported Protection ability",
			"the executable source backend supports only exact fixed-predicate protection",
		)
	}

	if !structureOK {
		return unsupported()
	}

	// Validate that the syntax tokens are fully covered by the keyword span.
	if !keywordOnlyCovered(syntax, keyword) {
		return unsupported()
	}

	if !keyword.ProtectionKnown || !protectionKeywordRuntimeSupported(keyword.Protection) {
		return unsupported()
	}
	return staticAbilityFromProtectionKeyword(keyword.Protection, ability.Text), true, nil
}

func protectionKeywordRuntimeSupported(prot game.ProtectionKeyword) bool {
	for _, sub := range prot.FromSubtypes {
		if !parser.SubtypeMatchesAnyRuntimeCardType(sub, []types.Card{types.Creature, types.Land}) {
			return false
		}
	}
	return true
}

// lowerKeywordDispatch tries Enchant, Protection, Cumulative upkeep, Equip,
// Cycling, Landcycling, Ninjutsu, Mutate, and Flashback — the
// single-keyword special cases that each produce a full abilityLowering.
// Returns (lowering, true, nil) on success, (lowering, true, diag) on a
// recognized-but-rejected attempt, and ({}, false, nil) when no attempt matches.
func lowerKeywordDispatch(
	creatureSubtypes []types.Sub,
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) (abilityLowering, bool, *shared.Diagnostic) {
	if enchantAbility, ok, diag := lowerEnchantAbility(ability, syntax); ok {
		if diag != nil {
			return abilityLowering{}, true, diag
		}
		return keywordStaticLowering(&enchantAbility, ability, syntax), true, nil
	}
	if protectionAbility, ok, diag := lowerProtectionAbility(ability, syntax); ok {
		if diag != nil {
			return abilityLowering{}, true, diag
		}
		return keywordStaticLowering(&protectionAbility, ability, syntax), true, nil
	}
	if cumulativeAbility, ok, diag := lowerCumulativeUpkeepAbility(ability, syntax); ok {
		if diag != nil {
			return abilityLowering{}, true, diag
		}
		return keywordTriggeredLowering(&cumulativeAbility, ability, syntax), true, nil
	}
	if fabricateAbility, ok, diag := lowerFabricateAbility(ability, syntax); ok {
		if diag != nil {
			return abilityLowering{}, true, diag
		}
		return keywordTriggeredLowering(&fabricateAbility, ability, syntax), true, nil
	}
	if soulshiftAbility, ok, diag := lowerSoulshiftAbility(ability, syntax); ok {
		if diag != nil {
			return abilityLowering{}, true, diag
		}
		return keywordTriggeredLowering(&soulshiftAbility, ability, syntax), true, nil
	}
	if undyingPersistAbility, ok, diag := lowerUndyingPersistAbility(ability, syntax); ok {
		if diag != nil {
			return abilityLowering{}, true, diag
		}
		return keywordTriggeredLowering(&undyingPersistAbility, ability, syntax), true, nil
	}
	if dethroneAbility, ok, diag := lowerDethroneAbility(ability, syntax); ok {
		if diag != nil {
			return abilityLowering{}, true, diag
		}
		return keywordTriggeredLowering(&dethroneAbility, ability, syntax), true, nil
	}
	if startEnginesAbility, ok, diag := lowerStartEnginesAbility(ability, syntax); ok {
		if diag != nil {
			return abilityLowering{}, true, diag
		}
		return keywordTriggeredLowering(&startEnginesAbility, ability, syntax), true, nil
	}
	if flankingAbility, ok, diag := lowerFlankingAbility(ability, syntax); ok {
		if diag != nil {
			return abilityLowering{}, true, diag
		}
		return keywordTriggeredLowering(&flankingAbility, ability, syntax), true, nil
	}
	if trainingAbility, ok, diag := lowerTrainingAbility(ability, syntax); ok {
		if diag != nil {
			return abilityLowering{}, true, diag
		}
		return keywordTriggeredLowering(&trainingAbility, ability, syntax), true, nil
	}
	if livingWeaponAbility, ok, diag := lowerLivingWeaponAbility(ability, syntax); ok {
		if diag != nil {
			return abilityLowering{}, true, diag
		}
		return keywordTriggeredLowering(&livingWeaponAbility, ability, syntax), true, nil
	}
	if equipAbility, ok, diag := lowerEquipAbility(ability, syntax); ok {
		if diag != nil {
			return abilityLowering{}, true, diag
		}
		return keywordActivatedLowering(&equipAbility, ability, syntax), true, nil
	}
	if cyclingAbility, ok, diag := lowerCyclingAbility(ability, syntax); ok {
		if diag != nil {
			return abilityLowering{}, true, diag
		}
		return keywordActivatedLowering(&cyclingAbility, ability, syntax), true, nil
	}
	if scavengeAbility, ok, diag := lowerScavengeAbility(ability, syntax); ok {
		if diag != nil {
			return abilityLowering{}, true, diag
		}
		return keywordActivatedLowering(&scavengeAbility, ability, syntax), true, nil
	}
	if unearthAbility, ok, diag := lowerUnearthAbility(ability, syntax); ok {
		if diag != nil {
			return abilityLowering{}, true, diag
		}
		return keywordActivatedLowering(&unearthAbility, ability, syntax), true, nil
	}
	if outlastAbility, ok, diag := lowerOutlastAbility(ability, syntax); ok {
		if diag != nil {
			return abilityLowering{}, true, diag
		}
		return keywordActivatedLowering(&outlastAbility, ability, syntax), true, nil
	}
	if saddleAbility, ok, diag := lowerSaddleAbility(ability, syntax); ok {
		if diag != nil {
			return abilityLowering{}, true, diag
		}
		return keywordActivatedLowering(&saddleAbility, ability, syntax), true, nil
	}
	if crewAbility, ok, diag := lowerCrewAbility(ability, syntax); ok {
		if diag != nil {
			return abilityLowering{}, true, diag
		}
		return keywordActivatedLowering(&crewAbility, ability, syntax), true, nil
	}
	if eternalizeAbility, ok, diag := lowerEternalizeAbility(creatureSubtypes, ability, syntax); ok {
		if diag != nil {
			return abilityLowering{}, true, diag
		}
		return keywordActivatedLowering(&eternalizeAbility, ability, syntax), true, nil
	}
	if embalmAbility, ok, diag := lowerEmbalmAbility(creatureSubtypes, ability, syntax); ok {
		if diag != nil {
			return abilityLowering{}, true, diag
		}
		return keywordActivatedLowering(&embalmAbility, ability, syntax), true, nil
	}
	if landcyclingAbility, ok, diag := lowerLandcyclingAbility(ability, syntax); ok {
		if diag != nil {
			return abilityLowering{}, true, diag
		}
		return keywordActivatedLowering(&landcyclingAbility, ability, syntax), true, nil
	}
	if ninjutsuAbility, ok, diag := lowerNinjutsuAbility(ability, syntax); ok {
		if diag != nil {
			return abilityLowering{}, true, diag
		}
		return keywordActivatedLowering(&ninjutsuAbility, ability, syntax), true, nil
	}
	if mutateAbility, ok, diag := lowerMutateAbility(ability, syntax); ok {
		if diag != nil {
			return abilityLowering{}, true, diag
		}
		return keywordStaticLowering(&mutateAbility, ability, syntax), true, nil
	}
	if flashbackAbility, ok, diag := lowerFlashbackAbility(ability, syntax); ok {
		if diag != nil {
			return abilityLowering{}, true, diag
		}
		return keywordStaticLowering(&flashbackAbility, ability, syntax), true, nil
	}
	if evokeLowering, ok, diag := lowerEvokeAbility(ability, syntax); ok {
		if diag != nil {
			return abilityLowering{}, true, diag
		}
		return evokeLowering, true, nil
	}
	if spectacleLowering, ok, diag := lowerSpectacleAbility(ability, syntax); ok {
		if diag != nil {
			return abilityLowering{}, true, diag
		}
		return spectacleLowering, true, nil
	}
	if dredgeAbility, ok, diag := lowerDredgeAbility(ability, syntax); ok {
		if diag != nil {
			return abilityLowering{}, true, diag
		}
		return keywordStaticLowering(&dredgeAbility, ability, syntax), true, nil
	}
	if bloodthirstAbility, ok, diag := lowerBloodthirstAbility(ability, syntax); ok {
		if diag != nil {
			return abilityLowering{}, true, diag
		}
		return keywordReplacementLowering(&bloodthirstAbility, ability, syntax), true, nil
	}
	if rampageAbility, ok, diag := lowerRampageAbility(ability, syntax); ok {
		if diag != nil {
			return abilityLowering{}, true, diag
		}
		return keywordTriggeredLowering(&rampageAbility, ability, syntax), true, nil
	}
	return abilityLowering{}, false, nil
}

func lowerCumulativeUpkeepAbility(
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) (game.TriggeredAbility, bool, *shared.Diagnostic) {
	if len(ability.Content.Keywords) != 1 || ability.Content.Keywords[0].Kind != parser.KeywordCumulativeUpkeep {
		return game.TriggeredAbility{}, false, nil
	}
	keyword := ability.Content.Keywords[0]
	manaCost, fixed := fixedKeywordManaCost(keyword)
	if !fixed ||
		ability.Kind != compiler.AbilityStatic ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		len(ability.Content.Effects) != 0 ||
		len(ability.Content.References) != 0 ||
		ability.AbilityWord != "" ||
		!keywordOnlyCovered(syntax, keyword) {
		return game.TriggeredAbility{}, true, executableDiagnostic(
			ability,
			"unsupported Cumulative upkeep ability",
			"the executable source backend supports only exact cumulative upkeep with one fixed mana cost",
		)
	}
	return game.CumulativeUpkeepTriggeredAbility(manaCost), true, nil
}

func lowerFabricateAbility(
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) (game.TriggeredAbility, bool, *shared.Diagnostic) {
	if len(ability.Content.Keywords) != 1 || ability.Content.Keywords[0].Kind != parser.KeywordFabricate {
		return game.TriggeredAbility{}, false, nil
	}
	keyword := ability.Content.Keywords[0]
	if keyword.ParameterKind != parser.KeywordParameterInteger ||
		keyword.Integer < 1 ||
		ability.Kind != compiler.AbilityStatic ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		len(ability.Content.Effects) != 0 ||
		len(ability.Content.References) != 0 ||
		ability.AbilityWord != "" ||
		!keywordOnlyCovered(syntax, keyword) {
		return game.TriggeredAbility{}, true, executableDiagnostic(
			ability,
			"unsupported Fabricate ability",
			"the executable source backend supports only exact Fabricate with one integer parameter",
		)
	}
	return game.FabricateTriggeredAbility(keyword.Integer), true, nil
}

// lowerSoulshiftAbility lowers the Soulshift N keyword (CR 702.46) to its
// canonical dies trigger that optionally returns a target Spirit card with mana
// value N or less from the controller's graveyard to their hand. Only the exact
// keyword with a fixed positive integer and no other rules text is supported.
func lowerSoulshiftAbility(
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) (game.TriggeredAbility, bool, *shared.Diagnostic) {
	if len(ability.Content.Keywords) != 1 || ability.Content.Keywords[0].Kind != parser.KeywordSoulshift {
		return game.TriggeredAbility{}, false, nil
	}
	keyword := ability.Content.Keywords[0]
	if keyword.ParameterKind != parser.KeywordParameterInteger ||
		keyword.Integer < 1 ||
		ability.Kind != compiler.AbilityStatic ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		len(ability.Content.Effects) != 0 ||
		len(ability.Content.References) != 0 ||
		ability.AbilityWord != "" ||
		!keywordOnlyCovered(syntax, keyword) {
		return game.TriggeredAbility{}, true, executableDiagnostic(
			ability,
			"unsupported Soulshift ability",
			"the executable source backend supports only exact Soulshift with one integer parameter",
		)
	}
	return game.SoulshiftTriggeredAbility(keyword.Integer), true, nil
}

// lowerUndyingPersistAbility lowers a printed Undying (CR 702.92) or Persist
// (CR 702.78) keyword to its canonical dies-return-with-counter triggered
// ability. Both keywords are printed bare (their reminder text is stripped), so
// the lowering expands the keyword to the reusable typed body. It supports only
// the exact keyword with no other rules text.
func lowerUndyingPersistAbility(
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) (game.TriggeredAbility, bool, *shared.Diagnostic) {
	if len(ability.Content.Keywords) != 1 {
		return game.TriggeredAbility{}, false, nil
	}
	keyword := ability.Content.Keywords[0]
	var body game.TriggeredAbility
	switch keyword.Kind {
	case parser.KeywordUndying:
		body = game.UndyingTriggeredBody
	case parser.KeywordPersist:
		body = game.PersistTriggeredBody
	default:
		return game.TriggeredAbility{}, false, nil
	}
	if keyword.ParameterKind != parser.KeywordParameterNone ||
		(ability.Kind != compiler.AbilityStatic && ability.Kind != compiler.AbilitySpell) ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		len(ability.Content.Effects) != 0 ||
		len(ability.Content.References) != 0 ||
		ability.AbilityWord != "" ||
		!keywordOnlyCovered(syntax, keyword) {
		return game.TriggeredAbility{}, true, executableDiagnostic(
			ability,
			"unsupported "+keyword.Name+" ability",
			"the executable source backend supports only the exact "+keyword.Name+" keyword",
		)
	}
	return body, true, nil
}

// lowerDethroneAbility lowers a printed Dethrone (CR 702.103) keyword to its
// canonical attacks-the-most-life-player triggered ability. Dethrone is printed
// bare (its reminder text is stripped), so the lowering expands the keyword to
// the reusable typed body. It supports only the exact keyword with no other
// rules text.
func lowerDethroneAbility(
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) (game.TriggeredAbility, bool, *shared.Diagnostic) {
	if len(ability.Content.Keywords) != 1 || ability.Content.Keywords[0].Kind != parser.KeywordDethrone {
		return game.TriggeredAbility{}, false, nil
	}
	keyword := ability.Content.Keywords[0]
	if keyword.ParameterKind != parser.KeywordParameterNone ||
		(ability.Kind != compiler.AbilityStatic && ability.Kind != compiler.AbilitySpell) ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		len(ability.Content.Effects) != 0 ||
		len(ability.Content.References) != 0 ||
		ability.AbilityWord != "" ||
		!keywordOnlyCovered(syntax, keyword) {
		return game.TriggeredAbility{}, true, executableDiagnostic(
			ability,
			"unsupported "+keyword.Name+" ability",
			"the executable source backend supports only the exact "+keyword.Name+" keyword",
		)
	}
	return game.DethroneTriggeredBody, true, nil
}

// lowerStartEnginesAbility lowers a printed "Start your engines!" (CR 702.179)
// keyword to its canonical enters-the-battlefield triggered ability, which runs
// the StartEngines primitive to seed the controller's speed to 1 if they have
// none. The keyword is printed with reminder text (stripped before lowering),
// so the lowering expands the bare keyword to the reusable typed body. Every
// printed instance is on a permanent, so the enters-the-battlefield trigger
// applies universally. It supports only the exact keyword with no other rules
// text.
func lowerStartEnginesAbility(
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) (game.TriggeredAbility, bool, *shared.Diagnostic) {
	if len(ability.Content.Keywords) != 1 || ability.Content.Keywords[0].Kind != parser.KeywordStartEngines {
		return game.TriggeredAbility{}, false, nil
	}
	keyword := ability.Content.Keywords[0]
	if keyword.ParameterKind != parser.KeywordParameterNone ||
		(ability.Kind != compiler.AbilityStatic && ability.Kind != compiler.AbilitySpell) ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		len(ability.Content.Effects) != 0 ||
		len(ability.Content.References) != 0 ||
		ability.AbilityWord != "" ||
		!keywordOnlyCovered(syntax, keyword) {
		return game.TriggeredAbility{}, true, executableDiagnostic(
			ability,
			"unsupported "+keyword.Name+" ability",
			"the executable source backend supports only the exact "+keyword.Name+" keyword",
		)
	}
	return game.StartEnginesTriggeredBody, true, nil
}

// lowerFlankingAbility lowers a printed Flanking (CR 702.25) keyword to its
// canonical becomes-blocked triggered ability. Flanking is printed bare (its
// reminder text is stripped), so the lowering expands the keyword to the
// reusable typed body. It supports only the exact keyword with no other rules
// text.
// lowerTrainingAbility lowers a printed Training (CR 702.150) keyword to its
// canonical attacks-with-greater-power triggered ability. Training is printed
// bare (its reminder text is stripped), so the lowering expands the keyword to
// the reusable typed body. It supports only the exact keyword with no other
// rules text.
func lowerTrainingAbility(
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) (game.TriggeredAbility, bool, *shared.Diagnostic) {
	if len(ability.Content.Keywords) != 1 || ability.Content.Keywords[0].Kind != parser.KeywordTraining {
		return game.TriggeredAbility{}, false, nil
	}
	keyword := ability.Content.Keywords[0]
	if keyword.ParameterKind != parser.KeywordParameterNone ||
		(ability.Kind != compiler.AbilityStatic && ability.Kind != compiler.AbilitySpell) ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		len(ability.Content.Effects) != 0 ||
		len(ability.Content.References) != 0 ||
		ability.AbilityWord != "" ||
		!keywordOnlyCovered(syntax, keyword) {
		return game.TriggeredAbility{}, true, executableDiagnostic(
			ability,
			"unsupported "+keyword.Name+" ability",
			"the executable source backend supports only the exact "+keyword.Name+" keyword",
		)
	}
	return game.TrainingTriggeredBody, true, nil
}

func lowerFlankingAbility(
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) (game.TriggeredAbility, bool, *shared.Diagnostic) {
	if len(ability.Content.Keywords) != 1 || ability.Content.Keywords[0].Kind != parser.KeywordFlanking {
		return game.TriggeredAbility{}, false, nil
	}
	keyword := ability.Content.Keywords[0]
	if keyword.ParameterKind != parser.KeywordParameterNone ||
		(ability.Kind != compiler.AbilityStatic && ability.Kind != compiler.AbilitySpell) ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		len(ability.Content.Effects) != 0 ||
		len(ability.Content.References) != 0 ||
		ability.AbilityWord != "" ||
		!keywordOnlyCovered(syntax, keyword) {
		return game.TriggeredAbility{}, true, executableDiagnostic(
			ability,
			"unsupported "+keyword.Name+" ability",
			"the executable source backend supports only the exact "+keyword.Name+" keyword",
		)
	}
	return game.FlankingTriggeredBody, true, nil
}

// lowerLivingWeaponAbility lowers a printed Living weapon (CR 702.91) keyword to
// its canonical enters-the-battlefield triggered ability: create a 0/0 black
// Phyrexian Germ creature token, then attach this Equipment to it. Living weapon
// is printed bare (its reminder text is stripped), so the lowering expands the
// keyword to the reusable typed body. It supports only the exact keyword with no
// other rules text.
func lowerLivingWeaponAbility(
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) (game.TriggeredAbility, bool, *shared.Diagnostic) {
	if len(ability.Content.Keywords) != 1 || ability.Content.Keywords[0].Kind != parser.KeywordLivingWeapon {
		return game.TriggeredAbility{}, false, nil
	}
	keyword := ability.Content.Keywords[0]
	if keyword.ParameterKind != parser.KeywordParameterNone ||
		(ability.Kind != compiler.AbilityStatic && ability.Kind != compiler.AbilitySpell) ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		len(ability.Content.Effects) != 0 ||
		len(ability.Content.References) != 0 ||
		ability.AbilityWord != "" ||
		!keywordOnlyCovered(syntax, keyword) {
		return game.TriggeredAbility{}, true, executableDiagnostic(
			ability,
			"unsupported "+keyword.Name+" ability",
			"the executable source backend supports only the exact "+keyword.Name+" keyword",
		)
	}
	return game.LivingWeaponTriggeredAbility(), true, nil
}

func keywordStaticLowering(
	body *game.StaticAbility,
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) abilityLowering {
	spans := keywordSpans(ability, syntax)
	return abilityLowering{
		staticAbilities: []loweredStaticAbility{{Body: *body}},
		consumed:        semanticConsumption{keywords: 1},
		sourceSpans:     spans,
	}
}

func keywordActivatedLowering(
	body *game.ActivatedAbility,
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) abilityLowering {
	spans := keywordSpans(ability, syntax)
	return abilityLowering{
		activatedAbility: opt.Val(*body),
		consumed:         semanticConsumption{keywords: 1},
		sourceSpans:      spans,
	}
}

func keywordTriggeredLowering(
	body *game.TriggeredAbility,
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) abilityLowering {
	return abilityLowering{
		triggeredAbility: opt.Val(*body),
		consumed:         semanticConsumption{keywords: 1},
		sourceSpans:      keywordSpans(ability, syntax),
	}
}

func keywordReplacementLowering(
	body *game.ReplacementAbility,
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) abilityLowering {
	return abilityLowering{
		replacementAbility: opt.Val(*body),
		consumed:           semanticConsumption{keywords: 1},
		sourceSpans:        keywordSpans(ability, syntax),
	}
}

// lowerBloodthirstAbility lowers the Bloodthirst N keyword (CR 702.54) to a
// conditional enters-with-counters replacement. Only the canonical fixed-N form
// is supported; the rare "Bloodthirst X" variant (Indoraptor) carries no
// integer parameter and stays unsupported.
func lowerBloodthirstAbility(
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) (game.ReplacementAbility, bool, *shared.Diagnostic) {
	if len(ability.Content.Keywords) != 1 || ability.Content.Keywords[0].Kind != parser.KeywordBloodthirst {
		return game.ReplacementAbility{}, false, nil
	}
	keyword := ability.Content.Keywords[0]
	if keyword.ParameterKind != parser.KeywordParameterInteger ||
		keyword.Integer <= 0 ||
		ability.Kind != compiler.AbilityStatic ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		len(ability.Content.Effects) != 0 ||
		len(ability.Content.References) != 0 ||
		ability.AbilityWord != "" ||
		!keywordOnlyCovered(syntax, keyword) {
		return game.ReplacementAbility{}, true, executableDiagnostic(
			ability,
			"unsupported Bloodthirst ability",
			"the executable source backend supports only exact \"Bloodthirst N\" with a fixed positive amount",
		)
	}
	return game.BloodthirstReplacement(keyword.Name+" "+keyword.Parameter, keyword.Integer), true, nil
}

// lowerRampageAbility lowers the Rampage N keyword (CR 702.23) to its canonical
// becomes-blocked triggered ability, which gives the source +N/+N until end of
// turn for each creature blocking it beyond the first. Only the exact keyword
// with a fixed positive integer and no other rules text is supported.
func lowerRampageAbility(
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) (game.TriggeredAbility, bool, *shared.Diagnostic) {
	if len(ability.Content.Keywords) != 1 || ability.Content.Keywords[0].Kind != parser.KeywordRampage {
		return game.TriggeredAbility{}, false, nil
	}
	keyword := ability.Content.Keywords[0]
	if keyword.ParameterKind != parser.KeywordParameterInteger ||
		keyword.Integer < 1 ||
		ability.Kind != compiler.AbilityStatic ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		len(ability.Content.Effects) != 0 ||
		len(ability.Content.References) != 0 ||
		ability.AbilityWord != "" ||
		!keywordOnlyCovered(syntax, keyword) {
		return game.TriggeredAbility{}, true, executableDiagnostic(
			ability,
			"unsupported Rampage ability",
			"the executable source backend supports only exact \"Rampage N\" with a fixed positive amount",
		)
	}
	return game.RampageTriggeredAbility(keyword.Integer), true, nil
}

func keywordSpans(ability compiler.CompiledAbility, syntax *parser.Ability) []shared.Span {
	spans := []shared.Span{ability.Content.Keywords[0].Span}
	for _, reminder := range syntax.Reminders {
		spans = append(spans, reminder.Span)
	}
	return spans
}

// staticAbilityFromProtectionKeyword builds a StaticAbility from a
// ProtectionKeyword using the appropriate factory function.
func staticAbilityFromProtectionKeyword(prot game.ProtectionKeyword, text string) game.StaticAbility {
	switch {
	case prot.Everything:
		return game.ProtectionFromEverythingStaticAbility()
	case prot.EachColor:
		return game.ProtectionFromEachColorStaticAbility()
	case prot.ChosenColor:
		return game.ProtectionFromChosenColorStaticAbility()
	case prot.Multicolored:
		return game.ProtectionFromMulticoloredStaticAbility()
	case prot.Monocolored:
		return game.ProtectionFromMonocoloredStaticAbility()
	case len(prot.FromTypes) > 0:
		return game.ProtectionFromTypesStaticAbility(prot.FromTypes...)
	case len(prot.FromSubtypes) > 0:
		return game.ProtectionFromSubtypesStaticAbility(prot.FromSubtypes...)
	case len(prot.FromColors) > 0:
		return game.ProtectionFromColorsStaticAbility(prot.FromColors...)
	default:
		panic(fmt.Sprintf("lower: empty ProtectionKeyword for %q", text))
	}
}

func enchantTargetSpec(target compiler.CompiledEnchantTarget) (game.TargetSpec, bool) {
	if !target.Known {
		return game.TargetSpec{}, false
	}
	spec := game.TargetSpec{
		MinTargets: 1,
		MaxTargets: 1,
	}
	switch {
	case target.Player:
		spec.Allow = game.TargetAllowPlayer
		spec.Constraint = "player"
		return spec, true
	case target.Opponent:
		spec.Allow = game.TargetAllowPlayer
		spec.Predicate.Player = game.PlayerOpponent
		spec.Constraint = "opponent"
		return spec, true
	case target.Permanent:
		spec.Allow = game.TargetAllowPermanent
		spec.Constraint = "permanent"
		return spec, true
	}
	spec.Allow = game.TargetAllowPermanent
	switch {
	case len(target.Subtypes) == 0:
		spec.Constraint = enchantConstraintText(target)
		spec.Predicate.PermanentTypes = slices.Clone(target.CardTypes)
	case len(target.CardTypes) == 0:
		spec.Constraint = enchantConstraintText(target)
		spec.Predicate.Subtypes = slices.Clone(target.Subtypes)
	default:
		// A union mixing card types and subtypes ("creature or Vehicle") is a
		// disjunction across two characteristic families, which a single
		// Selection cannot express conjunctively; AnyOf restores the "match any
		// alternative" meaning. The Constraint is intentionally left empty: the
		// runtime permanent-type matcher re-parses a non-empty Constraint and
		// cannot recognize a subtype as a card type, so an empty Constraint keeps
		// the Selection authoritative for attachment legality.
		spec.Selection = opt.Val(game.Selection{
			AnyOf: []game.Selection{
				{RequiredTypesAny: slices.Clone(target.CardTypes)},
				{SubtypesAny: slices.Clone(target.Subtypes)},
			},
		})
	}
	return spec, true
}

// enchantConstraintText renders the display Constraint for a permanent Enchant
// target from its typed card types and subtypes. The structured Allow,
// Predicate, and Selection fields drive legality; Constraint is display only.
func enchantConstraintText(target compiler.CompiledEnchantTarget) string {
	words := make([]string, 0, len(target.CardTypes)+len(target.Subtypes))
	for _, cardType := range target.CardTypes {
		words = append(words, strings.ToLower(string(cardType)))
	}
	for _, subtype := range target.Subtypes {
		words = append(words, strings.ToLower(string(subtype)))
	}
	return strings.Join(words, " or ")
}

func lowerEquipAbility(
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) (game.ActivatedAbility, bool, *shared.Diagnostic) {
	if len(ability.Content.Keywords) != 1 || ability.Content.Keywords[0].Kind != parser.KeywordEquip {
		return game.ActivatedAbility{}, false, nil
	}
	keyword := ability.Content.Keywords[0]
	if keyword.ParameterKind != parser.KeywordParameterManaCost ||
		ability.Kind != compiler.AbilityStatic ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		len(ability.Content.Effects) != 0 ||
		len(ability.Content.References) != 0 ||
		ability.AbilityWord != "" {
		return game.ActivatedAbility{}, true, executableDiagnostic(
			ability,
			"unsupported Equip ability",
			"the executable source backend supports only exact Equip with a mana cost",
		)
	}
	if len(keyword.ManaCost) == 0 {
		return game.ActivatedAbility{}, true, executableDiagnostic(
			ability,
			"unsupported Equip ability",
			"the executable source backend supports only exact Equip with a mana cost",
		)
	}
	if !keywordOnlyCovered(syntax, keyword) {
		return game.ActivatedAbility{}, true, executableDiagnostic(
			ability,
			"unsupported Equip ability",
			"the executable source backend supports only exact Equip with a mana cost",
		)
	}
	if keyword.EquipRestriction != nil {
		return game.EquipRestrictedActivatedAbility(
			slices.Clone(keyword.ManaCost),
			slices.Clone(keyword.EquipRestriction.Supertypes),
			slices.Clone(keyword.EquipRestriction.Subtypes),
		), true, nil
	}
	return game.EquipActivatedAbility(slices.Clone(keyword.ManaCost)), true, nil
}

// lowerEternalizeFamilyAbility lowers an Eternalize or Embalm keyword ability to
// its canonical graveyard-activated token-copy ability. creatureSubtypes are the
// card's printed creature subtypes (Zombie already removed), which the builder
// re-adds alongside the Zombie type the keyword grants. It mirrors Cycling: only
// an exact keyword with a fixed mana cost and no other rules text is supported.
func lowerEternalizeFamilyAbility(
	creatureSubtypes []types.Sub,
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
	kind parser.KeywordKind,
	name string,
	build func(cost.Mana, ...types.Sub) game.ActivatedAbility,
) (game.ActivatedAbility, bool, *shared.Diagnostic) {
	if len(ability.Content.Keywords) != 1 || ability.Content.Keywords[0].Kind != kind {
		return game.ActivatedAbility{}, false, nil
	}
	keyword := ability.Content.Keywords[0]
	manaCost, fixed := fixedKeywordManaCost(keyword)
	if !fixed ||
		(ability.Kind != compiler.AbilityStatic && ability.Kind != compiler.AbilitySpell) ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		len(ability.Content.Effects) != 0 ||
		len(ability.Content.References) != 0 ||
		ability.AbilityWord != "" ||
		!keywordOnlyCovered(syntax, keyword) {
		return game.ActivatedAbility{}, true, executableDiagnostic(
			ability,
			"unsupported "+name+" ability",
			"the executable source backend supports only exact "+name+" with a fixed mana cost",
		)
	}
	return build(manaCost, creatureSubtypes...), true, nil
}

func lowerEternalizeAbility(
	creatureSubtypes []types.Sub,
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) (game.ActivatedAbility, bool, *shared.Diagnostic) {
	return lowerEternalizeFamilyAbility(
		creatureSubtypes, ability, syntax,
		parser.KeywordEternalize, "Eternalize", game.EternalizeActivatedBody,
	)
}

func lowerEmbalmAbility(
	creatureSubtypes []types.Sub,
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) (game.ActivatedAbility, bool, *shared.Diagnostic) {
	return lowerEternalizeFamilyAbility(
		creatureSubtypes, ability, syntax,
		parser.KeywordEmbalm, "Embalm", game.EmbalmActivatedBody,
	)
}

func lowerCyclingAbility(
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) (game.ActivatedAbility, bool, *shared.Diagnostic) {
	if len(ability.Content.Keywords) != 1 || ability.Content.Keywords[0].Kind != parser.KeywordCycling {
		return game.ActivatedAbility{}, false, nil
	}
	keyword := ability.Content.Keywords[0]
	if keyword.ParameterKind == parser.KeywordParameterNone &&
		(len(ability.Content.Targets) != 0 || len(ability.Content.Effects) != 0 || len(ability.Content.References) != 0) {
		return game.ActivatedAbility{}, false, nil
	}
	if keyword.ParameterKind != parser.KeywordParameterManaCost ||
		(ability.Kind != compiler.AbilityStatic && ability.Kind != compiler.AbilitySpell) ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		len(ability.Content.Effects) != 0 ||
		len(ability.Content.References) != 0 ||
		ability.AbilityWord != "" {
		return game.ActivatedAbility{}, true, executableDiagnostic(
			ability,
			"unsupported Cycling ability",
			"the executable source backend supports only exact Cycling with a mana cost",
		)
	}
	if len(keyword.ManaCost) == 0 {
		return game.ActivatedAbility{}, true, executableDiagnostic(
			ability,
			"unsupported Cycling ability",
			"the executable source backend supports only exact Cycling with a mana cost",
		)
	}
	if !keywordOnlyCovered(syntax, keyword) {
		return game.ActivatedAbility{}, true, executableDiagnostic(
			ability,
			"unsupported Cycling ability",
			"the executable source backend supports only exact Cycling with a mana cost",
		)
	}
	return game.CyclingActivatedAbility(slices.Clone(keyword.ManaCost)), true, nil
}

// lowerOutlastAbility lowers an Outlast keyword with a mana cost to its
// canonical activated ability (CR 702.105): "[cost], {T}: Put a +1/+1 counter
// on this creature. Activate only as a sorcery." It mirrors lowerCyclingAbility:
// only an isolated, parameterized Outlast keyword is supported.
func lowerOutlastAbility(
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) (game.ActivatedAbility, bool, *shared.Diagnostic) {
	if len(ability.Content.Keywords) != 1 || ability.Content.Keywords[0].Kind != parser.KeywordOutlast {
		return game.ActivatedAbility{}, false, nil
	}
	keyword := ability.Content.Keywords[0]
	if keyword.ParameterKind != parser.KeywordParameterManaCost ||
		(ability.Kind != compiler.AbilityStatic && ability.Kind != compiler.AbilitySpell) ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		len(ability.Content.Effects) != 0 ||
		len(ability.Content.References) != 0 ||
		ability.AbilityWord != "" {
		return game.ActivatedAbility{}, true, executableDiagnostic(
			ability,
			"unsupported Outlast ability",
			"the executable source backend supports only exact Outlast with a mana cost",
		)
	}
	if len(keyword.ManaCost) == 0 {
		return game.ActivatedAbility{}, true, executableDiagnostic(
			ability,
			"unsupported Outlast ability",
			"the executable source backend supports only exact Outlast with a mana cost",
		)
	}
	if !keywordOnlyCovered(syntax, keyword) {
		return game.ActivatedAbility{}, true, executableDiagnostic(
			ability,
			"unsupported Outlast ability",
			"the executable source backend supports only exact Outlast with a mana cost",
		)
	}
	return game.OutlastActivatedAbility(slices.Clone(keyword.ManaCost)), true, nil
}

// lowerSaddleAbility lowers a Saddle N keyword to its canonical activated
// ability (CR 702.166): "Tap any number of other creatures you control with
// total power N or more: This Mount becomes saddled until end of turn. Saddle
// only as a sorcery." It mirrors lowerOutlastAbility: only an isolated,
// integer-parameterized Saddle keyword is supported.
func lowerSaddleAbility(
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) (game.ActivatedAbility, bool, *shared.Diagnostic) {
	if len(ability.Content.Keywords) != 1 || ability.Content.Keywords[0].Kind != parser.KeywordSaddle {
		return game.ActivatedAbility{}, false, nil
	}
	keyword := ability.Content.Keywords[0]
	if keyword.ParameterKind != parser.KeywordParameterInteger ||
		keyword.Integer < 1 ||
		(ability.Kind != compiler.AbilityStatic && ability.Kind != compiler.AbilitySpell) ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		len(ability.Content.Effects) != 0 ||
		len(ability.Content.References) != 0 ||
		ability.AbilityWord != "" {
		return game.ActivatedAbility{}, true, executableDiagnostic(
			ability,
			"unsupported Saddle ability",
			"the executable source backend supports only exact Saddle with a positive integer",
		)
	}
	if !keywordOnlyCovered(syntax, keyword) {
		return game.ActivatedAbility{}, true, executableDiagnostic(
			ability,
			"unsupported Saddle ability",
			"the executable source backend supports only exact Saddle with a positive integer",
		)
	}
	return game.SaddleActivatedAbility(keyword.Integer), true, nil
}

// lowerCrewAbility lowers a Crew N keyword to its canonical activated ability
// (CR 702.122): "Tap any number of creatures you control with total power N or
// more: This Vehicle becomes an artifact creature until end of turn." It mirrors
// lowerSaddleAbility: only an isolated, integer-parameterized Crew keyword is
// supported.
func lowerCrewAbility(
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) (game.ActivatedAbility, bool, *shared.Diagnostic) {
	if len(ability.Content.Keywords) != 1 || ability.Content.Keywords[0].Kind != parser.KeywordCrew {
		return game.ActivatedAbility{}, false, nil
	}
	keyword := ability.Content.Keywords[0]
	if keyword.ParameterKind != parser.KeywordParameterInteger ||
		keyword.Integer < 1 ||
		(ability.Kind != compiler.AbilityStatic && ability.Kind != compiler.AbilitySpell) ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		len(ability.Content.Effects) != 0 ||
		len(ability.Content.References) != 0 ||
		ability.AbilityWord != "" {
		return game.ActivatedAbility{}, true, executableDiagnostic(
			ability,
			"unsupported Crew ability",
			"the executable source backend supports only exact Crew with a positive integer",
		)
	}
	if !keywordOnlyCovered(syntax, keyword) {
		return game.ActivatedAbility{}, true, executableDiagnostic(
			ability,
			"unsupported Crew ability",
			"the executable source backend supports only exact Crew with a positive integer",
		)
	}
	return game.CrewActivatedAbility(keyword.Integer), true, nil
}

// landcyclingKeywordKinds maps each typed landcycling keyword to the library
// search filter its reminder text describes. Plain Landcycling finds any land;
// Basic landcycling finds a basic land; each typed variant finds a basic land
// of its own land type.
var landcyclingKeywordKinds = map[parser.KeywordKind]game.SearchSpec{
	parser.KeywordLandcycling:      {Filter: game.Selection{RequiredTypes: []types.Card{types.Land}}},
	parser.KeywordBasicLandcycling: {Filter: game.Selection{RequiredTypes: []types.Card{types.Land}, Supertypes: []types.Super{types.Basic}}},
	parser.KeywordPlainscycling:    {Filter: game.Selection{SubtypesAny: []types.Sub{types.Plains}}},
	parser.KeywordIslandcycling:    {Filter: game.Selection{SubtypesAny: []types.Sub{types.Island}}},
	parser.KeywordSwampcycling:     {Filter: game.Selection{SubtypesAny: []types.Sub{types.Swamp}}},
	parser.KeywordMountaincycling:  {Filter: game.Selection{SubtypesAny: []types.Sub{types.Mountain}}},
	parser.KeywordForestcycling:    {Filter: game.Selection{SubtypesAny: []types.Sub{types.Forest}}},
}

func lowerLandcyclingAbility(
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) (game.ActivatedAbility, bool, *shared.Diagnostic) {
	if len(ability.Content.Keywords) != 1 {
		return game.ActivatedAbility{}, false, nil
	}
	keyword := ability.Content.Keywords[0]
	spec, ok := landcyclingKeywordKinds[keyword.Kind]
	if !ok {
		return game.ActivatedAbility{}, false, nil
	}
	manaCost, fixed := fixedKeywordManaCost(keyword)
	if !fixed ||
		(ability.Kind != compiler.AbilityStatic && ability.Kind != compiler.AbilitySpell) ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		len(ability.Content.Effects) != 0 ||
		len(ability.Content.References) != 0 ||
		ability.AbilityWord != "" ||
		!keywordOnlyCovered(syntax, keyword) {
		return game.ActivatedAbility{}, true, executableDiagnostic(
			ability,
			"unsupported landcycling ability",
			"the executable source backend supports only exact typed landcycling with a mana cost",
		)
	}
	return game.LandcyclingActivatedAbility(manaCost, spec), true, nil
}

func lowerNinjutsuAbility(
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) (game.ActivatedAbility, bool, *shared.Diagnostic) {
	if len(ability.Content.Keywords) != 1 || ability.Content.Keywords[0].Kind != parser.KeywordNinjutsu {
		return game.ActivatedAbility{}, false, nil
	}
	keyword := ability.Content.Keywords[0]
	if keyword.ParameterKind != parser.KeywordParameterManaCost ||
		(ability.Kind != compiler.AbilityStatic && ability.Kind != compiler.AbilitySpell) ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		len(ability.Content.Effects) != 0 ||
		len(ability.Content.References) != 0 ||
		ability.AbilityWord != "" {
		return game.ActivatedAbility{}, true, executableDiagnostic(
			ability,
			"unsupported Ninjutsu ability",
			"the executable source backend supports only exact Ninjutsu with a mana cost",
		)
	}
	if len(keyword.ManaCost) == 0 {
		return game.ActivatedAbility{}, true, executableDiagnostic(
			ability,
			"unsupported Ninjutsu ability",
			"the executable source backend supports only exact Ninjutsu with a mana cost",
		)
	}
	if !keywordOnlyCovered(syntax, keyword) {
		return game.ActivatedAbility{}, true, executableDiagnostic(
			ability,
			"unsupported Ninjutsu ability",
			"the executable source backend supports only exact Ninjutsu with a mana cost",
		)
	}
	return game.NinjutsuActivatedAbility(slices.Clone(keyword.ManaCost)), true, nil
}

func lowerMutateAbility(
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) (game.StaticAbility, bool, *shared.Diagnostic) {
	if len(ability.Content.Keywords) != 1 || ability.Content.Keywords[0].Kind != parser.KeywordMutate {
		return game.StaticAbility{}, false, nil
	}
	keyword := ability.Content.Keywords[0]
	if keyword.ParameterKind != parser.KeywordParameterManaCost ||
		(ability.Kind != compiler.AbilityStatic && ability.Kind != compiler.AbilitySpell) ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		len(ability.Content.Effects) != 0 ||
		len(ability.Content.References) != 0 ||
		ability.AbilityWord != "" {
		return game.StaticAbility{}, true, executableDiagnostic(
			ability,
			"unsupported Mutate ability",
			"the executable source backend supports only exact Mutate with a mana cost",
		)
	}
	if len(keyword.ManaCost) == 0 {
		return game.StaticAbility{}, true, executableDiagnostic(
			ability,
			"unsupported Mutate ability",
			"the executable source backend supports only exact Mutate with a mana cost",
		)
	}
	if !keywordOnlyCovered(syntax, keyword) {
		return game.StaticAbility{}, true, executableDiagnostic(
			ability,
			"unsupported Mutate ability",
			"the executable source backend supports only exact Mutate with a mana cost",
		)
	}
	return game.MutateStaticAbility(slices.Clone(keyword.ManaCost)), true, nil
}

func lowerFlashbackAbility(
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) (game.StaticAbility, bool, *shared.Diagnostic) {
	if len(ability.Content.Keywords) != 1 || ability.Content.Keywords[0].Kind != parser.KeywordFlashback {
		return game.StaticAbility{}, false, nil
	}
	keyword := ability.Content.Keywords[0]
	manaCost, fixed := fixedKeywordManaCost(keyword)
	if !fixed ||
		(ability.Kind != compiler.AbilityStatic && ability.Kind != compiler.AbilitySpell) ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		len(ability.Content.Effects) != 0 ||
		len(ability.Content.References) != 0 ||
		ability.AbilityWord != "" ||
		!keywordOnlyCovered(syntax, keyword) {
		return game.StaticAbility{}, true, executableDiagnostic(
			ability,
			"unsupported Flashback ability",
			"the executable source backend supports only exact Flashback with a fixed mana cost",
		)
	}
	return game.StaticAbility{
		Text:             keyword.Name + " " + keyword.Parameter,
		KeywordAbilities: []game.KeywordAbility{game.FlashbackKeyword{Cost: manaCost}},
	}, true, nil
}

// lowerEvokeAbility lowers the Evoke keyword (CR 702.74) into two lowered
// pieces: an "Evoke" alternative spell cost the payment machinery auto-offers at
// cast, and the canonical evoke-sacrifice triggered ability that sacrifices the
// permanent when it enters if its evoke cost was paid. Only the exact keyword
// with a fixed mana cost and no other rules text is supported; the em-dash
// "Evoke—<non-mana cost>" variant carries no mana parameter and stays
// unsupported.
func lowerEvokeAbility(
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) (abilityLowering, bool, *shared.Diagnostic) {
	if len(ability.Content.Keywords) != 1 || ability.Content.Keywords[0].Kind != parser.KeywordEvoke {
		return abilityLowering{}, false, nil
	}
	keyword := ability.Content.Keywords[0]
	manaCost, fixed := fixedKeywordManaCost(keyword)
	if !fixed ||
		(ability.Kind != compiler.AbilityStatic && ability.Kind != compiler.AbilitySpell) ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		len(ability.Content.Effects) != 0 ||
		len(ability.Content.References) != 0 ||
		ability.AbilityWord != "" ||
		!keywordOnlyCovered(syntax, keyword) {
		return abilityLowering{}, true, executableDiagnostic(
			ability,
			"unsupported Evoke ability",
			"the executable source backend supports only exact Evoke with a fixed mana cost",
		)
	}
	return abilityLowering{
		triggeredAbility: opt.Val(game.EvokeSacrificeTriggeredAbility()),
		alternativeCosts: []cost.Alternative{{
			Label:    "Evoke",
			ManaCost: opt.Val(manaCost),
		}},
		consumed:    semanticConsumption{keywords: 1},
		sourceSpans: keywordSpans(ability, syntax),
	}, true, nil
}

// lowerSpectacleAbility lowers the Spectacle keyword (CR 702.107): "Spectacle
// <cost>" lets the spell be cast for its spectacle cost rather than its mana
// cost if an opponent lost life this turn. It produces a single conditional
// alternative cost; unlike Evoke there is no entry sacrifice trigger. Only the
// exact keyword with a fixed mana cost and no other rules text is supported;
// anything else fails closed.
func lowerSpectacleAbility(
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) (abilityLowering, bool, *shared.Diagnostic) {
	if len(ability.Content.Keywords) != 1 || ability.Content.Keywords[0].Kind != parser.KeywordSpectacle {
		return abilityLowering{}, false, nil
	}
	keyword := ability.Content.Keywords[0]
	manaCost, fixed := fixedKeywordManaCost(keyword)
	if !fixed ||
		(ability.Kind != compiler.AbilityStatic && ability.Kind != compiler.AbilitySpell) ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		len(ability.Content.Effects) != 0 ||
		len(ability.Content.References) != 0 ||
		ability.AbilityWord != "" ||
		!keywordOnlyCovered(syntax, keyword) {
		return abilityLowering{}, true, executableDiagnostic(
			ability,
			"unsupported Spectacle ability",
			"the executable source backend supports only exact Spectacle with a fixed mana cost",
		)
	}
	return abilityLowering{
		alternativeCosts: []cost.Alternative{{
			Label:     "Spectacle",
			ManaCost:  opt.Val(manaCost),
			Condition: cost.AlternativeConditionOpponentLostLifeThisTurn,
		}},
		consumed:    semanticConsumption{keywords: 1},
		sourceSpans: keywordSpans(ability, syntax),
	}, true, nil
}

// lowerDredgeAbility lowers the Dredge N keyword (CR 702.52) to its canonical
// graveyard-functioning static ability. While the card is in its owner's
// graveyard, the runtime offers to replace one of that player's draws with
// milling N cards and returning this card to hand. Only the exact keyword with a
// fixed positive integer and no other rules text is supported; anything else
// (the rare "Opponent dredge" variant, a non-integer parameter) fails closed.
func lowerDredgeAbility(
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) (game.StaticAbility, bool, *shared.Diagnostic) {
	if len(ability.Content.Keywords) != 1 || ability.Content.Keywords[0].Kind != parser.KeywordDredge {
		return game.StaticAbility{}, false, nil
	}
	keyword := ability.Content.Keywords[0]
	if keyword.ParameterKind != parser.KeywordParameterInteger ||
		keyword.Integer <= 0 ||
		(ability.Kind != compiler.AbilityStatic && ability.Kind != compiler.AbilitySpell) ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		len(ability.Content.Effects) != 0 ||
		len(ability.Content.References) != 0 ||
		ability.AbilityWord != "" ||
		!keywordOnlyCovered(syntax, keyword) {
		return game.StaticAbility{}, true, executableDiagnostic(
			ability,
			"unsupported Dredge ability",
			"the executable source backend supports only exact \"Dredge N\" with a fixed positive amount",
		)
	}
	return game.DredgeStaticAbility(keyword.Integer), true, nil
}

func simpleStaticKeyword(keyword compiler.CompiledKeyword) (game.Keyword, bool) {
	if keyword.ParameterKind != parser.KeywordParameterNone {
		return 0, false
	}
	body, ok := keywordStaticBodies[keyword.Kind]
	if !ok || len(body.Body.KeywordAbilities) != 1 {
		return 0, false
	}
	simple, ok := body.Body.KeywordAbilities[0].(game.SimpleKeyword)
	if !ok || !mixedStaticKeywordImplemented(simple.Kind) {
		return 0, false
	}
	return simple.Kind, true
}

func mixedStaticKeywords(keywords []compiler.CompiledKeyword) ([]game.Keyword, bool) {
	result := make([]game.Keyword, 0, len(keywords))
	for _, keyword := range keywords {
		simple, ok := simpleStaticKeyword(keyword)
		if !ok {
			return nil, false
		}
		result = append(result, simple)
	}
	return result, true
}

// partitionTemporaryKeywords splits keyword grants into simple keyword enum
// values and granted ability bodies. Protection keywords lower to static ability
// bodies so the grant carries their full characteristics; every other keyword
// must reduce to a simple keyword. It fails closed for anything else.
func partitionTemporaryKeywords(keywords []compiler.CompiledKeyword) ([]game.Keyword, []game.Ability, bool) {
	simpleKeywords := make([]game.Keyword, 0, len(keywords))
	var abilities []game.Ability
	for _, keyword := range keywords {
		if keyword.Kind == parser.KeywordProtection {
			if !keyword.ProtectionKnown {
				return nil, nil, false
			}
			ability := staticAbilityFromProtectionKeyword(keyword.Protection, keyword.Text)
			abilities = append(abilities, &ability)
			continue
		}
		simple, ok := simpleStaticKeyword(keyword)
		if !ok {
			return nil, nil, false
		}
		simpleKeywords = append(simpleKeywords, simple)
	}
	return simpleKeywords, abilities, true
}

// abilityKeywordsExcludingSelectorPredicates returns the ability's keyword grants
// with the keyword atoms that actually function as selector predicates removed.
// A keyword written inside a target or effect-selector noun phrase ("deals 1
// damage to target creature with flying", "each creature with flying") is
// recorded both as that selector's Keyword and, redundantly, as a semantic
// ability keyword; the latter would otherwise make damage and other spell
// lowerings treat the ability as if it granted the keyword. Only keyword atoms
// whose source span falls inside a selector phrase that carries the same keyword
// are excluded, so a genuine standalone keyword grant elsewhere on the ability is
// preserved.
func abilityKeywordsExcludingSelectorPredicates(content compiler.AbilityContent) []compiler.CompiledKeyword {
	filtered := make([]compiler.CompiledKeyword, 0, len(content.Keywords))
	for _, keyword := range content.Keywords {
		if keywordIsSelectorPredicate(content, keyword) {
			continue
		}
		filtered = append(filtered, keyword)
	}
	return filtered
}

// keywordIsSelectorPredicate reports whether a keyword atom is a "with
// <keyword>" selector predicate rather than a granted ability keyword. It holds
// when a target or effect selector carries the same keyword kind and the keyword
// atom's span lies within that selector's source span.
func keywordIsSelectorPredicate(content compiler.AbilityContent, keyword compiler.CompiledKeyword) bool {
	if keyword.ParameterKind != parser.KeywordParameterNone {
		return false
	}
	for i := range content.Targets {
		target := &content.Targets[i]
		if target.Selector.Keyword == keyword.Kind && spanContains(target.Span, keyword.Span) {
			return true
		}
	}
	for i := range content.Effects {
		effect := &content.Effects[i]
		if effect.Selector.Keyword == keyword.Kind && spanContains(effect.Span, keyword.Span) {
			return true
		}
		if effect.Amount.Selector().Keyword == keyword.Kind && spanContains(effect.Span, keyword.Span) {
			return true
		}
	}
	return false
}

// spanContains reports whether outer fully covers inner by byte offset.
func spanContains(outer, inner shared.Span) bool {
	return inner.Start.Offset >= outer.Start.Offset && inner.End.Offset <= outer.End.Offset
}

func mixedStaticKeywordImplemented(keyword game.Keyword) bool {
	switch keyword {
	case game.Deathtouch,
		game.Banding,
		game.Defender,
		game.DoubleStrike,
		game.FirstStrike,
		game.Flying,
		game.Haste,
		game.Hexproof,
		game.Indestructible,
		game.Lifelink,
		game.Menace,
		game.Reach,
		game.Shroud,
		game.Trample,
		game.Vigilance,
		game.Wither,
		game.Horsemanship,
		game.Shadow,
		game.Infect,
		game.Exalted,
		game.Riot,
		game.Evolve,
		game.Unleash,
		game.Fear,
		game.Skulk,
		game.Intimidate:
		return true
	default:
		return false
	}
}

func resolvingStaticSubjectGroup(effect *compiler.CompiledEffect) (game.GroupReference, bool) {
	// One-shot mass effects do not yet lower a color-filtered affected group;
	// fail closed rather than silently dropping the color constraint. Color
	// filtering is supported only for never-resolving static declarations.
	if effect.StaticSubjectHasColorFilter() {
		return game.GroupReference{}, false
	}
	selection := game.Selection{Controller: game.ControllerYou}
	switch effect.StaticSubject {
	case compiler.StaticSubjectAllCreatures:
		return game.BattlefieldGroup(game.Selection{
			RequiredTypes: []types.Card{types.Creature},
		}), true
	case compiler.StaticSubjectAllOtherCreatures:
		return game.BattlefieldGroupExcluding(
			game.Selection{RequiredTypes: []types.Card{types.Creature}},
			game.SourcePermanentReference(),
		), true
	case compiler.StaticSubjectAttackingCreatures:
		return game.BattlefieldGroup(game.Selection{
			RequiredTypes: []types.Card{types.Creature},
			CombatState:   game.CombatStateAttacking,
		}), true
	case compiler.StaticSubjectOtherAttackingCreatures:
		return game.BattlefieldGroupExcluding(
			game.Selection{
				RequiredTypes: []types.Card{types.Creature},
				CombatState:   game.CombatStateAttacking,
			},
			game.SourcePermanentReference(),
		), true
	case compiler.StaticSubjectBlockingCreatures:
		return game.BattlefieldGroup(game.Selection{
			RequiredTypes: []types.Card{types.Creature},
			CombatState:   game.CombatStateBlocking,
		}), true
	case compiler.StaticSubjectControlledPermanents:
	case compiler.StaticSubjectOpponentControlledPermanents:
		selection.Controller = game.ControllerOpponent
	case compiler.StaticSubjectControlledCreatures:
		selection.RequiredTypes = []types.Card{types.Creature}
	case compiler.StaticSubjectOtherControlledCreatures:
		selection.RequiredTypes = []types.Card{types.Creature}
		return game.BattlefieldGroupExcluding(selection, game.SourcePermanentReference()), true
	case compiler.StaticSubjectControlledWalls:
		selection.SubtypesAny = []types.Sub{types.Wall}
	case compiler.StaticSubjectControlledArtifacts:
		selection.RequiredTypes = []types.Card{types.Artifact}
	case compiler.StaticSubjectControlledTokens:
		selection.TokenOnly = true
	case compiler.StaticSubjectOpponentControlledCreatures:
		selection.RequiredTypes = []types.Card{types.Creature}
		selection.Controller = game.ControllerOpponent
	case compiler.StaticSubjectControlledCreatureSubtype:
		if !effect.StaticSubjectSubKnown() {
			return game.GroupReference{}, false
		}
		if effect.StaticSubjectSubExcluded() {
			selection.RequiredTypes = []types.Card{types.Creature}
			selection.ExcludedSubtype = effect.StaticSubjectSub()
		} else {
			selection.SubtypesAny = []types.Sub{effect.StaticSubjectSub()}
		}
	case compiler.StaticSubjectOtherControlledCreatureSubtype:
		if !effect.StaticSubjectSubKnown() {
			return game.GroupReference{}, false
		}
		if effect.StaticSubjectSubExcluded() {
			selection.RequiredTypes = []types.Card{types.Creature}
			selection.ExcludedSubtype = effect.StaticSubjectSub()
		} else {
			selection.SubtypesAny = []types.Sub{effect.StaticSubjectSub()}
		}
		return game.BattlefieldGroupExcluding(selection, game.SourcePermanentReference()), true
	default:
		return game.GroupReference{}, false
	}
	return game.BattlefieldGroup(selection), true
}

func lowerKeywordAbility(
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) ([]loweredStaticAbility, *shared.Diagnostic) {
	for _, keyword := range ability.Content.Keywords {
		if keyword.Kind == parser.KeywordDevoid && !syntax.DevoidRecognized {
			return nil, executableDiagnostic(
				ability,
				"unsupported Devoid ability",
				"the executable source backend supports only exact \"Devoid (This card has no color.)\" abilities",
			)
		}
		if keyword.Kind == parser.KeywordReadAhead && !syntax.ReadAheadRecognized {
			return nil, executableDiagnostic(
				ability,
				"unsupported Read ahead ability",
				"the executable source backend supports only the canonical Read ahead ability and reminder text",
			)
		}
	}
	if len(ability.Content.Modes) > 0 {
		return nil, executableDiagnostic(
			ability,
			"unsupported modal ability",
			"the executable source backend does not yet lower modal abilities",
		)
	}
	if !rulesFreeAbilityWordLabel(ability.AbilityWord) {
		return nil, executableDiagnostic(
			ability,
			"unsupported ability word",
			fmt.Sprintf("the executable source backend does not yet lower the %q ability word", ability.AbilityWord),
		)
	}
	if len(ability.Content.Keywords) == 0 {
		return nil, executableDiagnostic(
			ability,
			"unsupported static ability",
			"the executable source backend does not yet lower non-keyword static rules text",
		)
	}
	bodies := make([]loweredStaticAbility, 0, len(ability.Content.Keywords))
	for _, keyword := range ability.Content.Keywords {
		if keyword.ParameterKind != parser.KeywordParameterNone {
			if body, ok, diag := lowerParameterizedKeywordToStaticAbility(ability, keyword); ok {
				if diag != nil {
					return nil, diag
				}
				bodies = append(bodies, loweredStaticAbility{Body: body})
				continue
			}
			return nil, executableDiagnostic(
				ability,
				"unsupported parameterized keyword",
				fmt.Sprintf(
					"the executable source backend does not yet lower %s with parameter %q",
					keyword.Name,
					keyword.Parameter,
				),
			)
		}
		body, ok := keywordStaticBodies[keyword.Kind]
		if !ok {
			return nil, executableDiagnostic(
				ability,
				"unsupported keyword ability",
				fmt.Sprintf(
					"the executable source backend has no reusable game template for %s",
					keyword.Name,
				),
			)
		}
		bodies = append(bodies, body)
	}
	if len(ability.Content.Targets) > 0 ||
		len(ability.Content.Conditions) > 0 ||
		len(ability.Content.Effects) > 0 ||
		len(ability.Content.References) > 0 {
		return nil, mixedKeywordDiagnostic(contentCtx{span: ability.Span, content: ability.Content})
	}
	for _, span := range syntax.CoverageSpans() {
		if (syntax.AbilityWord != nil && span == syntax.AbilityWord.SeparatorSpan) ||
			spanCoveredByAbilityWord(span, syntax.AbilityWord) ||
			spanCoveredByKeyword(span, ability.Content.Keywords) ||
			spanCoveredByDelimited(span, syntax.Reminders) ||
			spanIsKeywordListSemicolon(span, syntax.Tokens) {
			continue
		}
		return nil, mixedKeywordDiagnostic(contentCtx{span: ability.Span, content: ability.Content})
	}
	return bodies, nil
}

func rulesFreeAbilityWordLabel(label string) bool {
	switch label {
	case "",
		"Addendum",
		"Bloodrush",
		"Celebration",
		"Channel",
		"Converge",
		"Corrupted",
		"Coven",
		"Delirium",
		"Domain",
		"Enrage",
		"Ferocious",
		"Flurry",
		"Formidable",
		"Hellbent",
		"Inspired",
		"Kinship",
		"Lieutenant",
		"Magecraft",
		"Metalcraft",
		"Morbid",
		"Opus",
		"Parley",
		"Raid",
		"Revolt",
		"Survival",
		"Threshold",
		"Vivid",
		"Void",
		"Will of the council":
		return true
	default:
		return false
	}
}

func syntaxWithoutAbilityWord(syntax *parser.Ability) parser.Ability {
	result := *syntax
	if result.AbilityWord == nil {
		return result
	}
	result.Tokens = parser.TokensFrom(result.Tokens, result.AbilityWord.SeparatorSpan.End.Offset)
	return result
}

func spellBodyWithoutAbilityWord(
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) (compiler.CompiledAbility, parser.Ability, bool) {
	if ability.AbilityWord == "" {
		return ability, *syntax, true
	}
	if !rulesFreeAbilityWordLabel(ability.AbilityWord) || syntax.AbilityWord == nil {
		return compiler.CompiledAbility{}, parser.Ability{}, false
	}
	body := syntaxWithoutAbilityWord(syntax)
	if len(body.Tokens) == 0 {
		return compiler.CompiledAbility{}, parser.Ability{}, false
	}
	start := body.Tokens[0].Span.Start
	offset := start.Offset - ability.Span.Start.Offset
	if offset < 0 || offset >= len(ability.Text) {
		return compiler.CompiledAbility{}, parser.Ability{}, false
	}
	ability.Text = strings.TrimSpace(ability.Text[offset:])
	ability.Span.Start = start
	ability.AbilityWord = ""
	body.Span.Start = start
	body.Text = ability.Text
	body.AbilityWord = nil
	return ability, body, true
}

func tokensWithoutSpans(tokens []shared.Token, spans ...shared.Span) []shared.Token {
	return slices.DeleteFunc(append([]shared.Token(nil), tokens...), func(token shared.Token) bool {
		return spanCovered(token.Span, spans)
	})
}

// lowerParameterizedKeywordToStaticAbility handles lowering of a single
// parameterized keyword (Ward, Protection, and others) to a static ability.
// Returns (body, true, nil) on success, ({}, true, diag) on a recognised but
// unsupported form, and ({}, false, nil) when no handler matches.
func lowerParameterizedKeywordToStaticAbility(
	ability compiler.CompiledAbility,
	keyword compiler.CompiledKeyword,
) (game.StaticAbility, bool, *shared.Diagnostic) {
	switch keyword.Kind {
	case parser.KeywordWard:
		if keyword.ParameterKind == parser.KeywordParameterManaCost && len(keyword.ManaCost) > 0 {
			return game.WardStaticAbility(slices.Clone(keyword.ManaCost)), true, nil
		}
	case parser.KeywordProtection:
		if keyword.ProtectionKnown {
			return staticAbilityFromProtectionKeyword(keyword.Protection, ""), true, nil
		}
	default:
	}
	if body, ok := lowerParameterizedStaticKeyword(keyword); ok {
		return body, true, nil
	}
	return game.StaticAbility{}, false, nil
}

func lowerParameterizedStaticKeyword(keyword compiler.CompiledKeyword) (game.StaticAbility, bool) {
	body := game.StaticAbility{Text: keyword.Name + " " + keyword.Parameter}
	switch keyword.Kind {
	case parser.KeywordKicker:
		manaCost, ok := fixedKeywordManaCost(keyword)
		if !ok {
			return game.StaticAbility{}, false
		}
		body.KeywordAbilities = []game.KeywordAbility{game.KickerKeyword{Cost: manaCost}}
	case parser.KeywordMultikicker:
		manaCost, ok := fixedKeywordManaCost(keyword)
		if !ok {
			return game.StaticAbility{}, false
		}
		body.KeywordAbilities = []game.KeywordAbility{game.KickerKeyword{Cost: manaCost, Multi: true}}
	case parser.KeywordMadness:
		manaCost, ok := fixedKeywordManaCost(keyword)
		if !ok {
			return game.StaticAbility{}, false
		}
		body.KeywordAbilities = []game.KeywordAbility{game.MadnessKeyword{Cost: manaCost}}
	case parser.KeywordFlashback:
		manaCost, ok := fixedKeywordManaCost(keyword)
		if !ok {
			return game.StaticAbility{}, false
		}
		body.KeywordAbilities = []game.KeywordAbility{game.FlashbackKeyword{Cost: manaCost}}
	case parser.KeywordMorph:
		manaCost, ok := fixedKeywordManaCost(keyword)
		if !ok {
			return game.StaticAbility{}, false
		}
		body.KeywordAbilities = []game.KeywordAbility{game.MorphKeyword{Cost: manaCost}}
	case parser.KeywordDisguise:
		manaCost, ok := fixedKeywordManaCost(keyword)
		if !ok {
			return game.StaticAbility{}, false
		}
		body.KeywordAbilities = []game.KeywordAbility{game.DisguiseKeyword{Cost: manaCost}}
	case parser.KeywordToxic:
		if keyword.ParameterKind != parser.KeywordParameterInteger || keyword.Integer <= 0 {
			return game.StaticAbility{}, false
		}
		body.KeywordAbilities = []game.KeywordAbility{game.ToxicKeyword{Amount: keyword.Integer}}
	default:
		return game.StaticAbility{}, false
	}
	return body, true
}

func fixedKeywordManaCost(keyword compiler.CompiledKeyword) (cost.Mana, bool) {
	if keyword.ParameterKind != parser.KeywordParameterManaCost || len(keyword.ManaCost) == 0 {
		return nil, false
	}
	for _, symbol := range keyword.ManaCost {
		if symbol.Kind == cost.VariableSymbol {
			return nil, false
		}
	}
	return slices.Clone(keyword.ManaCost), true
}

func lowerScavengeAbility(
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) (game.ActivatedAbility, bool, *shared.Diagnostic) {
	if len(ability.Content.Keywords) != 1 || ability.Content.Keywords[0].Kind != parser.KeywordScavenge {
		return game.ActivatedAbility{}, false, nil
	}
	keyword := ability.Content.Keywords[0]
	if keyword.ParameterKind != parser.KeywordParameterManaCost ||
		len(keyword.ManaCost) == 0 ||
		(ability.Kind != compiler.AbilityStatic && ability.Kind != compiler.AbilitySpell) ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		len(ability.Content.Effects) != 0 ||
		len(ability.Content.References) != 0 ||
		ability.AbilityWord != "" ||
		!keywordOnlyCovered(syntax, keyword) {
		return game.ActivatedAbility{}, true, executableDiagnostic(
			ability,
			"unsupported Scavenge ability",
			"the executable source backend supports only exact Scavenge with a mana cost",
		)
	}
	return game.ScavengeActivatedAbility(slices.Clone(keyword.ManaCost)), true, nil
}

// lowerUnearthAbility lowers the Unearth keyword (CR 702.83) to its canonical
// graveyard-activated ability: at sorcery speed the controller pays the unearth
// cost to return this card from their graveyard to the battlefield with haste,
// exiling it at the next end step. Only the exact keyword with a fixed mana cost
// and no other rules text is supported; anything else fails closed.
func lowerUnearthAbility(
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) (game.ActivatedAbility, bool, *shared.Diagnostic) {
	if len(ability.Content.Keywords) != 1 || ability.Content.Keywords[0].Kind != parser.KeywordUnearth {
		return game.ActivatedAbility{}, false, nil
	}
	keyword := ability.Content.Keywords[0]
	if keyword.ParameterKind != parser.KeywordParameterManaCost ||
		len(keyword.ManaCost) == 0 ||
		(ability.Kind != compiler.AbilityStatic && ability.Kind != compiler.AbilitySpell) ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		len(ability.Content.Effects) != 0 ||
		len(ability.Content.References) != 0 ||
		ability.AbilityWord != "" ||
		!keywordOnlyCovered(syntax, keyword) {
		return game.ActivatedAbility{}, true, executableDiagnostic(
			ability,
			"unsupported Unearth ability",
			"the executable source backend supports only exact Unearth with a mana cost",
		)
	}
	return game.UnearthActivatedAbility(slices.Clone(keyword.ManaCost)), true, nil
}
