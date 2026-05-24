package rules

import (
	"slices"
	"strings"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
)

func targetChoicesForSpell(g *game.Game, controller game.PlayerID, card *game.CardDef) [][]game.Target {
	specs := spellTargetSpecs(card)
	if len(specs) == 0 {
		return [][]game.Target{nil}
	}
	if len(specs) != 1 || !isSingleTargetSpec(specs[0]) {
		return nil
	}

	var choices [][]game.Target
	spec := specs[0]
	if targetSpecAllowsPlayers(spec) {
		for playerID := game.Player1; playerID < game.NumPlayers; playerID++ {
			target := game.PlayerTarget(playerID)
			if targetMatchesSpec(g, controller, spec, target) {
				choices = append(choices, []game.Target{target})
			}
		}
	}
	if targetSpecAllowsPermanents(spec) {
		for _, permanent := range g.Battlefield {
			if permanent == nil {
				continue
			}
			target := game.PermanentTarget(permanent.ObjectID)
			if targetMatchesSpec(g, controller, spec, target) {
				choices = append(choices, []game.Target{target})
			}
		}
	}
	return choices
}

func targetsValidForSpell(g *game.Game, controller game.PlayerID, card *game.CardDef, targets []game.Target) bool {
	specs := spellTargetSpecs(card)
	if len(specs) == 0 {
		return len(targets) == 0
	}
	if len(specs) != len(targets) {
		return false
	}
	for i, spec := range specs {
		if !isSingleTargetSpec(spec) || !targetMatchesSpec(g, controller, spec, targets[i]) {
			return false
		}
	}
	return true
}

func spellHasAnyLegalTargets(g *game.Game, card *game.CardDef, controller game.PlayerID, targets []game.Target) bool {
	specs := spellTargetSpecs(card)
	if len(specs) == 0 {
		return true
	}
	if len(specs) != len(targets) {
		return false
	}
	for i, spec := range specs {
		if targetMatchesSpec(g, controller, spec, targets[i]) {
			return true
		}
	}
	return false
}

func spellTargetSpecs(card *game.CardDef) []game.TargetSpec {
	ability := firstSpellAbility(card)
	if ability == nil {
		return nil
	}
	return ability.Targets
}

func isSingleTargetSpec(spec game.TargetSpec) bool {
	return spec.MinTargets == 1 &&
		spec.MaxTargets == 1
}

func targetSpecAllowsPlayers(spec game.TargetSpec) bool {
	normalized := normalizedTargetConstraint(spec)
	return normalized == "player" ||
		normalized == "target player" ||
		normalized == "opponent" ||
		normalized == "target opponent" ||
		normalized == "any target"
}

func targetSpecAllowsPermanents(spec game.TargetSpec) bool {
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

func targetMatchesSpec(g *game.Game, controller game.PlayerID, spec game.TargetSpec, target game.Target) bool {
	switch target.Kind {
	case game.TargetPlayer:
		return playerTargetMatchesSpec(g, controller, spec, target.PlayerID)
	case game.TargetPermanent:
		return permanentTargetMatchesSpec(g, controller, spec, target.PermanentID)
	default:
		return false
	}
}

func playerTargetMatchesSpec(g *game.Game, controller game.PlayerID, spec game.TargetSpec, playerID game.PlayerID) bool {
	if !isPlayerAlive(g, playerID) || !targetSpecAllowsPlayers(spec) {
		return false
	}
	normalized := normalizedTargetConstraint(spec)
	if strings.Contains(normalized, "opponent") && playerID == controller {
		return false
	}
	return true
}

func permanentTargetMatchesSpec(g *game.Game, controller game.PlayerID, spec game.TargetSpec, permanentID id.ID) bool {
	if !targetSpecAllowsPermanents(spec) {
		return false
	}
	permanent := permanentByObjectID(g, permanentID)
	if permanent == nil || permanent.PhasedOut {
		return false
	}
	if !permanentControllerMatchesSpec(g, controller, spec, permanent) {
		return false
	}
	if normalizedTargetConstraint(spec) == "any target" {
		card := permanentCardDef(g, permanent)
		return card != nil && (card.HasType(game.TypeCreature) || card.HasType(game.TypePlaneswalker) || card.HasType(game.TypeBattle))
	}
	return permanentTypeMatchesSpec(g, spec, permanent)
}

func permanentControllerMatchesSpec(g *game.Game, controller game.PlayerID, spec game.TargetSpec, permanent *game.Permanent) bool {
	normalized := normalizedTargetConstraint(spec)
	switch {
	case strings.Contains(normalized, "you control") || strings.Contains(normalized, "controlled by you"):
		return permanent.Controller == controller
	case strings.Contains(normalized, "opponent controls") ||
		strings.Contains(normalized, "opponents control") ||
		strings.Contains(normalized, "controlled by an opponent") ||
		strings.Contains(normalized, "controlled by opponent"):
		return permanent.Controller != controller && isPlayerAlive(g, permanent.Controller)
	default:
		return true
	}
}

func permanentTypeMatchesSpec(g *game.Game, spec game.TargetSpec, permanent *game.Permanent) bool {
	card := permanentCardDef(g, permanent)
	if card == nil {
		return false
	}
	normalized := normalizedTargetConstraint(spec)
	if strings.Contains(normalized, "nonland permanent") {
		return !card.HasType(game.TypeLand)
	}
	if strings.Contains(normalized, "permanent") && !containsAnyPermanentTypeConstraint(normalized) {
		return card.IsPermanent()
	}
	allowedTypes := permanentTypesForConstraint(normalized)
	if len(allowedTypes) == 0 {
		return false
	}
	return slices.ContainsFunc(allowedTypes, card.HasType)
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
