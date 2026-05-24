package rules

import (
	"slices"
	"strings"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
)

func targetChoicesForSpell(g *game.Game, controller game.PlayerID, card *game.CardDef, chosenModes []int) [][]game.Target {
	specs := spellTargetSpecs(card, chosenModes)
	return targetChoicesForSpecs(g, controller, card, 0, specs)
}

func targetChoicesForAbility(g *game.Game, controller game.PlayerID, ability *game.AbilityDef) [][]game.Target {
	return targetChoicesForAbilityFromSource(g, controller, nil, ability)
}

func targetChoicesForAbilityFromSource(g *game.Game, controller game.PlayerID, source *game.CardDef, ability *game.AbilityDef) [][]game.Target {
	return targetChoicesForAbilityFromSourceObject(g, controller, source, 0, ability)
}

func targetChoicesForAbilityFromSourceObject(g *game.Game, controller game.PlayerID, source *game.CardDef, sourceObjectID id.ID, ability *game.AbilityDef) [][]game.Target {
	if ability == nil {
		return nil
	}
	return targetChoicesForSpecs(g, controller, source, sourceObjectID, ability.Targets)
}

func targetChoicesForSpecs(g *game.Game, controller game.PlayerID, source *game.CardDef, sourceObjectID id.ID, specs []game.TargetSpec) [][]game.Target {
	if len(specs) == 0 {
		return [][]game.Target{nil}
	}
	var result [][]game.Target
	appendTargetChoicesForSpec(g, controller, source, sourceObjectID, specs, 0, nil, &result)
	return result
}

func appendTargetChoicesForSpec(g *game.Game, controller game.PlayerID, source *game.CardDef, sourceObjectID id.ID, specs []game.TargetSpec, specIndex int, prefix []game.Target, result *[][]game.Target) {
	if specIndex >= len(specs) {
		*result = append(*result, append([]game.Target(nil), prefix...))
		return
	}
	spec := normalizeTargetSpec(specs[specIndex])
	if !targetSpecRangeValid(spec) {
		return
	}
	candidates := targetCandidatesForSpec(g, controller, source, sourceObjectID, spec)
	maxTargets := min(spec.MaxTargets, len(candidates))
	for _, count := range targetCountsForChoices(spec.MinTargets, maxTargets) {
		for _, combination := range targetCombinations(candidates, count) {
			next := append(append([]game.Target(nil), prefix...), combination...)
			appendTargetChoicesForSpec(g, controller, source, sourceObjectID, specs, specIndex+1, next, result)
		}
	}
}

func targetCountsForChoices(minTargets int, maxTargets int) []int {
	var counts []int
	start := minTargets
	if minTargets == 0 && maxTargets > 0 {
		start = 1
	}
	for count := start; count <= maxTargets; count++ {
		counts = append(counts, count)
	}
	if minTargets == 0 && start > 0 {
		counts = append(counts, 0)
	}
	return counts
}

func targetCandidatesForSpec(g *game.Game, controller game.PlayerID, source *game.CardDef, sourceObjectID id.ID, spec game.TargetSpec) []game.Target {
	var candidates []game.Target
	if targetSpecAllowsPlayers(spec) {
		for playerID := game.Player1; playerID < game.NumPlayers; playerID++ {
			target := game.PlayerTarget(playerID)
			if targetMatchesSpec(g, controller, sourceObjectID, spec, target) && !targetProtectedFromSource(g, source, target) {
				candidates = append(candidates, target)
			}
		}
	}
	if targetSpecAllowsPermanents(spec) {
		for _, permanent := range g.Battlefield {
			if permanent == nil {
				continue
			}
			target := game.PermanentTarget(permanent.ObjectID)
			if targetMatchesSpec(g, controller, sourceObjectID, spec, target) && !targetProtectedFromSource(g, source, target) {
				candidates = append(candidates, target)
			}
		}
	}
	return candidates
}

func targetCombinations(candidates []game.Target, count int) [][]game.Target {
	if count == 0 {
		return [][]game.Target{nil}
	}
	if count > len(candidates) {
		return nil
	}
	var result [][]game.Target
	var walk func(start int, chosen []game.Target)
	walk = func(start int, chosen []game.Target) {
		if len(chosen) == count {
			result = append(result, append([]game.Target(nil), chosen...))
			return
		}
		need := count - len(chosen)
		for i := start; i <= len(candidates)-need; i++ {
			walk(i+1, append(chosen, candidates[i]))
		}
	}
	walk(0, nil)
	return result
}

func targetsValidForSpell(g *game.Game, controller game.PlayerID, card *game.CardDef, chosenModes []int, targets []game.Target) bool {
	specs := spellTargetSpecs(card, chosenModes)
	return targetsValidForSpecs(g, controller, card, 0, specs, targets)
}

func targetsValidForAbility(g *game.Game, controller game.PlayerID, ability *game.AbilityDef, targets []game.Target) bool {
	return targetsValidForAbilityFromSource(g, controller, nil, ability, targets)
}

func targetsValidForAbilityFromSource(g *game.Game, controller game.PlayerID, source *game.CardDef, ability *game.AbilityDef, targets []game.Target) bool {
	return targetsValidForAbilityFromSourceObject(g, controller, source, 0, ability, targets)
}

func targetsValidForAbilityFromSourceObject(g *game.Game, controller game.PlayerID, source *game.CardDef, sourceObjectID id.ID, ability *game.AbilityDef, targets []game.Target) bool {
	if ability == nil {
		return len(targets) == 0
	}
	return targetsValidForSpecs(g, controller, source, sourceObjectID, ability.Targets, targets)
}

func targetsValidForSpecs(g *game.Game, controller game.PlayerID, source *game.CardDef, sourceObjectID id.ID, specs []game.TargetSpec, targets []game.Target) bool {
	if len(specs) == 0 {
		return len(targets) == 0
	}
	return targetsValidForSpecFrom(g, controller, source, sourceObjectID, specs, targets, 0, 0)
}

func targetsValidForSpecFrom(g *game.Game, controller game.PlayerID, source *game.CardDef, sourceObjectID id.ID, specs []game.TargetSpec, targets []game.Target, specIndex int, targetIndex int) bool {
	if specIndex == len(specs) {
		return targetIndex == len(targets)
	}
	spec := normalizeTargetSpec(specs[specIndex])
	if !targetSpecRangeValid(spec) {
		return false
	}
	remaining := len(targets) - targetIndex
	maxTargets := min(spec.MaxTargets, remaining)
	for count := spec.MinTargets; count <= maxTargets; count++ {
		slice := targets[targetIndex : targetIndex+count]
		if targetsMatchSpecSlice(g, controller, source, sourceObjectID, spec, slice) &&
			targetsValidForSpecFrom(g, controller, source, sourceObjectID, specs, targets, specIndex+1, targetIndex+count) {
			return true
		}
	}
	return false
}

func targetsMatchSpecSlice(g *game.Game, controller game.PlayerID, source *game.CardDef, sourceObjectID id.ID, spec game.TargetSpec, targets []game.Target) bool {
	if len(targets) < spec.MinTargets || len(targets) > spec.MaxTargets {
		return false
	}
	seen := make(map[game.Target]bool, len(targets))
	for _, target := range targets {
		if seen[target] || !targetMatchesSpec(g, controller, sourceObjectID, spec, target) || targetProtectedFromSource(g, source, target) {
			return false
		}
		seen[target] = true
	}
	return true
}

func spellHasAnyLegalTargets(g *game.Game, card *game.CardDef, controller game.PlayerID, chosenModes []int, targets []game.Target) bool {
	specs := spellTargetSpecs(card, chosenModes)
	return hasAnyLegalTargetForSpecs(g, controller, card, 0, specs, targets)
}

func abilityHasAnyLegalTargets(g *game.Game, ability *game.AbilityDef, controller game.PlayerID, targets []game.Target) bool {
	return abilityHasAnyLegalTargetsFromSource(g, nil, ability, controller, targets)
}

func abilityHasAnyLegalTargetsFromSource(g *game.Game, source *game.CardDef, ability *game.AbilityDef, controller game.PlayerID, targets []game.Target) bool {
	return abilityHasAnyLegalTargetsFromSourceObject(g, source, 0, ability, controller, targets)
}

func abilityHasAnyLegalTargetsFromSourceObject(g *game.Game, source *game.CardDef, sourceObjectID id.ID, ability *game.AbilityDef, controller game.PlayerID, targets []game.Target) bool {
	if ability == nil {
		return len(targets) == 0
	}
	return hasAnyLegalTargetForSpecs(g, controller, source, sourceObjectID, ability.Targets, targets)
}

func hasAnyLegalTargetForSpecs(g *game.Game, controller game.PlayerID, source *game.CardDef, sourceObjectID id.ID, specs []game.TargetSpec, targets []game.Target) bool {
	if len(specs) == 0 {
		return true
	}
	return targetsValidForSpecs(g, controller, source, sourceObjectID, specs, targets)
}

func spellTargetSpecs(card *game.CardDef, chosenModes []int) []game.TargetSpec {
	ability := firstSpellAbility(card)
	if ability == nil {
		return nil
	}
	if len(ability.Modes) > 0 {
		if !modesValidForAbility(ability, chosenModes) {
			return nil
		}
		var specs []game.TargetSpec
		for _, modeIndex := range chosenModes {
			specs = append(specs, ability.Modes[modeIndex].Targets...)
		}
		return specs
	}
	return ability.Targets
}

func modeChoicesForSpell(card *game.CardDef) [][]int {
	ability := firstSpellAbility(card)
	if ability == nil || len(ability.Modes) == 0 {
		return [][]int{nil}
	}
	choices := make([][]int, 0, len(ability.Modes))
	for i := range ability.Modes {
		choices = append(choices, []int{i})
	}
	return choices
}

func modesValidForSpell(card *game.CardDef, chosenModes []int) bool {
	ability := firstSpellAbility(card)
	if ability == nil {
		return len(chosenModes) == 0
	}
	return modesValidForAbility(ability, chosenModes)
}

func modesValidForAbility(ability *game.AbilityDef, chosenModes []int) bool {
	if ability == nil {
		return len(chosenModes) == 0
	}
	if len(ability.Modes) == 0 {
		return len(chosenModes) == 0
	}
	if len(chosenModes) != 1 {
		return false
	}
	return chosenModes[0] >= 0 && chosenModes[0] < len(ability.Modes)
}

func normalizeTargetSpec(spec game.TargetSpec) game.TargetSpec {
	if spec.MinTargets == 0 && spec.MaxTargets == 0 && spec.Constraint != "" {
		spec.MinTargets = 1
		spec.MaxTargets = 1
	}
	return spec
}

func targetSpecRangeValid(spec game.TargetSpec) bool {
	return spec.MinTargets >= 0 && spec.MaxTargets >= spec.MinTargets
}

func targetSpecAllowsPlayers(spec game.TargetSpec) bool {
	if spec.Allow != game.TargetAllowUnspecified {
		return spec.Allow&game.TargetAllowPlayer != 0
	}
	normalized := normalizedTargetConstraint(spec)
	return normalized == "player" ||
		normalized == "target player" ||
		normalized == "opponent" ||
		normalized == "target opponent" ||
		normalized == "any target"
}

func targetSpecAllowsPermanents(spec game.TargetSpec) bool {
	if spec.Allow != game.TargetAllowUnspecified {
		return spec.Allow&game.TargetAllowPermanent != 0
	}
	normalized := normalizedTargetConstraint(spec)
	if normalized == "any target" {
		return true
	}
	if strings.Contains(normalized, "permanent") ||
		strings.Contains(normalized, "creature") ||
		strings.Contains(normalized, "artifact") ||
		strings.Contains(normalized, "enchantment") ||
		strings.Contains(normalized, "land") ||
		strings.Contains(normalized, "planeswalker") ||
		strings.Contains(normalized, "battle") {
		return true
	}
	return false
}

func targetMatchesSpec(g *game.Game, controller game.PlayerID, sourceObjectID id.ID, spec game.TargetSpec, target game.Target) bool {
	switch target.Kind {
	case game.TargetPlayer:
		return playerTargetMatchesSpec(g, controller, spec, target.PlayerID)
	case game.TargetPermanent:
		return permanentTargetMatchesSpec(g, controller, sourceObjectID, spec, target.PermanentID)
	default:
		return false
	}
}

func playerTargetMatchesSpec(g *game.Game, controller game.PlayerID, spec game.TargetSpec, playerID game.PlayerID) bool {
	if !isPlayerAlive(g, playerID) || !targetSpecAllowsPlayers(spec) {
		return false
	}
	switch spec.Predicate.Player {
	case game.PlayerYou:
		return playerID == controller
	case game.PlayerOpponent, game.PlayerNotYou:
		return playerID != controller
	}
	normalized := normalizedTargetConstraint(spec)
	if strings.Contains(normalized, "opponent") && playerID == controller {
		return false
	}
	return true
}

func permanentTargetMatchesSpec(g *game.Game, controller game.PlayerID, sourceObjectID id.ID, spec game.TargetSpec, permanentID id.ID) bool {
	if !targetSpecAllowsPermanents(spec) {
		return false
	}
	permanent := permanentByObjectID(g, permanentID)
	if permanent == nil || permanent.PhasedOut {
		return false
	}
	if spec.Predicate.Another && sourceObjectID != 0 && permanent.ObjectID == sourceObjectID {
		return false
	}
	if !permanentControllerMatchesSpec(g, controller, spec, permanent) {
		return false
	}
	if !structuredPermanentPredicateMatches(g, spec.Predicate, permanent) {
		return false
	}
	if normalizedTargetConstraint(spec) == "any target" {
		return permanentHasType(g, permanent, game.TypeCreature) ||
			permanentHasType(g, permanent, game.TypePlaneswalker) ||
			permanentHasType(g, permanent, game.TypeBattle)
	}
	return permanentTypeMatchesSpec(g, spec, permanent)
}

func structuredPermanentPredicateMatches(g *game.Game, predicate game.TargetPredicate, permanent *game.Permanent) bool {
	if len(predicate.PermanentTypes) > 0 && !slices.ContainsFunc(predicate.PermanentTypes, func(cardType game.CardType) bool {
		return permanentHasType(g, permanent, cardType)
	}) {
		return false
	}
	if slices.ContainsFunc(predicate.ExcludedTypes, func(cardType game.CardType) bool {
		return permanentHasType(g, permanent, cardType)
	}) {
		return false
	}
	colors := permanentEffectiveColors(g, permanent)
	if len(predicate.Colors) > 0 && !slices.ContainsFunc(predicate.Colors, func(color mana.Color) bool {
		return slices.Contains(colors, color)
	}) {
		return false
	}
	if slices.ContainsFunc(predicate.ExcludedColors, func(color mana.Color) bool {
		return slices.Contains(colors, color)
	}) {
		return false
	}
	if predicate.Tapped == game.TriTrue && !permanent.Tapped {
		return false
	}
	if predicate.Tapped == game.TriFalse && permanent.Tapped {
		return false
	}
	if !combatStateMatches(g, permanent, predicate.CombatState) {
		return false
	}
	if predicate.Keyword != game.KeywordNone && !hasKeyword(g, permanent, predicate.Keyword) {
		return false
	}
	if predicate.ExcludedKeyword != game.KeywordNone && hasKeyword(g, permanent, predicate.ExcludedKeyword) {
		return false
	}
	if predicate.ManaValue != nil {
		def := permanentCardDef(g, permanent)
		if def == nil || !intComparisonMatches(def.ManaValue, *predicate.ManaValue) {
			return false
		}
	}
	if predicate.Power != nil && !intComparisonMatches(effectivePower(g, permanent), *predicate.Power) {
		return false
	}
	if predicate.Toughness != nil {
		toughness, ok := effectiveToughness(g, permanent)
		if !ok || !intComparisonMatches(toughness, *predicate.Toughness) {
			return false
		}
	}
	return true
}

func combatStateMatches(g *game.Game, permanent *game.Permanent, filter game.CombatStateFilter) bool {
	if filter == game.CombatStateAny {
		return true
	}
	attacking := false
	blocking := false
	if g != nil && g.Combat != nil && permanent != nil {
		attacking = slices.ContainsFunc(g.Combat.Attackers, func(declaration game.AttackDeclaration) bool {
			return declaration.Attacker == permanent.ObjectID
		})
		blocking = slices.ContainsFunc(g.Combat.Blockers, func(declaration game.BlockDeclaration) bool {
			return declaration.Blocker == permanent.ObjectID
		})
	}
	switch filter {
	case game.CombatStateAttacking:
		return attacking
	case game.CombatStateBlocking:
		return blocking
	case game.CombatStateAttackingOrBlocking:
		return attacking || blocking
	default:
		return true
	}
}

func intComparisonMatches(value int, comparison game.IntComparison) bool {
	switch comparison.Op {
	case game.CompareEqual:
		return value == comparison.Value
	case game.CompareLessOrEqual:
		return value <= comparison.Value
	case game.CompareGreaterOrEqual:
		return value >= comparison.Value
	case game.CompareLessThan:
		return value < comparison.Value
	case game.CompareGreaterThan:
		return value > comparison.Value
	default:
		return true
	}
}

func targetProtectedFromSource(g *game.Game, source *game.CardDef, target game.Target) bool {
	if source == nil || target.Kind != game.TargetPermanent {
		return false
	}
	permanent := permanentByObjectID(g, target.PermanentID)
	return permanentProtectedFromSourceDef(g, permanent, source)
}

func permanentControllerMatchesSpec(g *game.Game, controller game.PlayerID, spec game.TargetSpec, permanent *game.Permanent) bool {
	permanentController := effectiveController(g, permanent)
	switch spec.Predicate.Controller {
	case game.ControllerYou:
		return permanentController == controller
	case game.ControllerOpponent, game.ControllerNotYou:
		return permanentController != controller && isPlayerAlive(g, permanentController)
	}
	normalized := normalizedTargetConstraint(spec)
	switch {
	case strings.Contains(normalized, "you control") || strings.Contains(normalized, "controlled by you"):
		return permanentController == controller
	case strings.Contains(normalized, "opponent controls") ||
		strings.Contains(normalized, "opponents control") ||
		strings.Contains(normalized, "controlled by an opponent") ||
		strings.Contains(normalized, "controlled by opponent"):
		return permanentController != controller && isPlayerAlive(g, permanentController)
	default:
		return true
	}
}

func permanentTypeMatchesSpec(g *game.Game, spec game.TargetSpec, permanent *game.Permanent) bool {
	if permanent == nil {
		return false
	}
	if len(spec.Predicate.PermanentTypes) > 0 || len(spec.Predicate.ExcludedTypes) > 0 {
		return true
	}
	normalized := normalizedTargetConstraint(spec)
	if spec.Allow != game.TargetAllowUnspecified && normalized == "" {
		if spec.Allow&game.TargetAllowPlayer != 0 {
			return permanentHasType(g, permanent, game.TypeCreature) ||
				permanentHasType(g, permanent, game.TypePlaneswalker) ||
				permanentHasType(g, permanent, game.TypeBattle)
		}
		return len(effectivePermanentValues(g, permanent).types) > 0
	}
	if strings.Contains(normalized, "nonland permanent") {
		return !permanentHasType(g, permanent, game.TypeLand)
	}
	if strings.Contains(normalized, "permanent") && !containsAnyPermanentTypeConstraint(normalized) {
		return len(effectivePermanentValues(g, permanent).types) > 0
	}
	allowedTypes := permanentTypesForConstraint(normalized)
	if len(allowedTypes) == 0 {
		return false
	}
	return slices.ContainsFunc(allowedTypes, func(cardType game.CardType) bool {
		return permanentHasType(g, permanent, cardType)
	})
}

func containsAnyPermanentTypeConstraint(normalized string) bool {
	return strings.Contains(normalized, "creature") ||
		strings.Contains(normalized, "artifact") ||
		strings.Contains(normalized, "enchantment") ||
		strings.Contains(normalized, "land") ||
		strings.Contains(normalized, "planeswalker") ||
		strings.Contains(normalized, "battle")
}

func permanentTypesForConstraint(normalized string) []game.CardType {
	var types []game.CardType
	if strings.Contains(normalized, "creature") {
		types = append(types, game.TypeCreature)
	}
	if strings.Contains(normalized, "artifact") {
		types = append(types, game.TypeArtifact)
	}
	if strings.Contains(normalized, "enchantment") {
		types = append(types, game.TypeEnchantment)
	}
	if strings.Contains(normalized, "land") {
		types = append(types, game.TypeLand)
	}
	if strings.Contains(normalized, "planeswalker") {
		types = append(types, game.TypePlaneswalker)
	}
	if strings.Contains(normalized, "battle") {
		types = append(types, game.TypeBattle)
	}
	return types
}

func normalizedTargetConstraint(spec game.TargetSpec) string {
	normalized := strings.ToLower(strings.TrimSpace(spec.Constraint))
	normalized = strings.TrimPrefix(normalized, "target ")
	return strings.Join(strings.Fields(normalized), " ")
}

func isPlayerAlive(g *game.Game, playerID game.PlayerID) bool {
	if g == nil || playerID < 0 || int(playerID) >= len(g.Players) {
		return false
	}
	player := g.Players[playerID]
	return player != nil && !player.Eliminated && !g.TurnOrder.IsEliminated(playerID)
}
