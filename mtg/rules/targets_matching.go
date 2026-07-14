package rules

import (
	"slices"
	"strings"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func normalizeTargetSpec(spec *game.TargetSpec) game.TargetSpec {
	normalized := *spec
	if normalized.MinTargets == 0 && normalized.MaxTargets == 0 && normalized.Constraint != "" {
		normalized.MinTargets = 1
		normalized.MaxTargets = 1
	}
	return normalized
}

func targetSpecValid(spec *game.TargetSpec) bool {
	if spec.MinTargets < 0 || spec.MaxTargets < spec.MinTargets {
		return false
	}
	switch spec.Chooser {
	case game.TargetChooserController:
		return true
	case game.TargetChooserOpponent:
		return spec.MinTargets == 1 && spec.MaxTargets == 1
	default:
		return false
	}
}

func targetSpecUsesExternalChooser(spec *game.TargetSpec) bool {
	return spec.Chooser != game.TargetChooserController
}

func choosingOpponentsForTargetSpec(g *game.Game, controller game.PlayerID, source *game.CardDef, sourceObjectID id.ID, spec *game.TargetSpec) []game.PlayerID {
	if spec.Chooser != game.TargetChooserOpponent {
		return nil
	}
	var players []game.PlayerID
	current := controller
	for range game.NumPlayers - 1 {
		current = g.TurnOrder.NextPriority(current)
		if current == controller {
			break
		}
		if !isPlayerAlive(g, current) {
			continue
		}
		if len(targetCandidatesForSpecChosenBy(g, controller, current, source, sourceObjectID, game.Event{}, spec)) > 0 {
			players = append(players, current)
		}
	}
	return players
}

func externalChooserCouldChooseTarget(g *game.Game, controller game.PlayerID, source *game.CardDef, sourceObjectID id.ID, spec *game.TargetSpec, target game.Target) bool {
	for _, chooser := range choosingOpponentsForTargetSpec(g, controller, source, sourceObjectID, spec) {
		if slices.Contains(targetCandidatesForSpecChosenBy(g, controller, chooser, source, sourceObjectID, game.Event{}, spec), target) {
			return true
		}
	}
	return false
}

func targetSpecAllowsPlayers(spec *game.TargetSpec) bool {
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

func targetSpecAllowsPermanents(spec *game.TargetSpec) bool {
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

func targetSpecAllowsCards(spec *game.TargetSpec) bool {
	if spec.Allow != game.TargetAllowUnspecified {
		return spec.Allow&game.TargetAllowCard != 0
	}
	normalized := normalizedTargetConstraint(spec)
	return strings.Contains(normalized, "card") &&
		(strings.Contains(normalized, "graveyard") || strings.Contains(normalized, "library") || strings.Contains(normalized, "hand"))
}

func targetSpecAllowsStackObjects(spec *game.TargetSpec) bool {
	if spec.Allow != game.TargetAllowUnspecified {
		return spec.Allow&game.TargetAllowStackObject != 0
	}
	return false
}

func targetMatchesSpec(g *game.Game, controller game.PlayerID, sourceObjectID id.ID, triggerEvent game.Event, spec *game.TargetSpec, target game.Target) bool {
	switch target.Kind {
	case game.TargetPlayer:
		return playerTargetMatchesSpec(g, controller, spec, target.PlayerID)
	case game.TargetPermanent:
		return permanentTargetMatchesSpec(g, controller, sourceObjectID, triggerEvent, spec, target.PermanentID)
	case game.TargetCard:
		return cardTargetMatchesSpec(g, controller, triggerEvent, spec, target)
	case game.TargetStackObject:
		return stackObjectTargetMatchesSpec(g, controller, sourceObjectID, spec, target.StackObjectID)
	default:
		return false
	}
}

func stackObjectTargetMatchesSpec(g *game.Game, controller game.PlayerID, sourceObjectID id.ID, spec *game.TargetSpec, stackObjectID id.ID) bool {
	if !targetSpecAllowsStackObjects(spec) {
		return false
	}
	obj, ok := stackObjectByID(g, stackObjectID)
	if !ok || stackObjectID == sourceObjectID {
		return false
	}
	switch obj.Kind {
	case game.StackSpell, game.StackActivatedAbility, game.StackTriggeredAbility:
	default:
		return false
	}
	pred := spec.Predicate
	if !slices.Contains(pred.StackObjectKinds, obj.Kind) {
		return false
	}
	if !controllerRelationMatches(controller, stackObjectController(obj), pred.Controller) {
		return false
	}
	if !stackObjectSourceHasTypes(g, obj, pred.StackObjectSourceTypes) {
		return false
	}
	if !stackObjectSpellTargetsMatch(g, controller, obj, pred.SpellTargets) {
		return false
	}
	return stackObjectSpellQualifiersMatch(g, obj, pred)
}

// stackObjectSpellTargetsMatch enforces a "Counter target spell that targets
// <X>" restriction: the matched spell must have at least one chosen target
// satisfying one of the requirements. Abilities never satisfy the restriction,
// since only spells carry the "that targets" qualifier (CR 115.4). An empty
// requirement list imposes no restriction.
func stackObjectSpellTargetsMatch(g *game.Game, controller game.PlayerID, obj *game.StackObject, requirements []game.SpellTargetRequirement) bool {
	if len(requirements) == 0 {
		return true
	}
	if obj.Kind != game.StackSpell {
		return false
	}
	for i := range obj.Targets {
		for j := range requirements {
			if spellTargetSatisfiesRequirement(g, controller, obj.Targets[i], &requirements[j]) {
				return true
			}
		}
	}
	return false
}

// spellTargetSatisfiesRequirement reports whether one of a matched spell's
// chosen targets satisfies a single "that targets" requirement: a permanent
// requirement checks the targeted permanent's types and controller; a player
// requirement checks the targeted player's relation. Relations are evaluated
// relative to the player choosing the counter target.
func spellTargetSatisfiesRequirement(g *game.Game, controller game.PlayerID, target game.Target, requirement *game.SpellTargetRequirement) bool {
	switch requirement.Kind {
	case game.SpellTargetRequirementPlayer:
		if target.Kind != game.TargetPlayer {
			return false
		}
		return playerRelationMatches(controller, target.PlayerID, requirement.Player)
	case game.SpellTargetRequirementPermanent:
		if target.Kind != game.TargetPermanent {
			return false
		}
		permanent, ok := permanentByObjectID(g, target.PermanentID)
		if !ok || permanent.PhasedOut {
			return false
		}
		if !controllerRelationMatches(controller, effectiveController(g, permanent), requirement.Controller) {
			return false
		}
		for _, cardType := range requirement.RequiredTypes {
			if !permanentHasType(g, permanent, cardType) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

// stackObjectSpellQualifiersMatch enforces the spell-only restrictions a mixed
// stack-object target may carry. Abilities ignore spell qualifiers so a mixed
// "ability or qualified spell" target accepts any ability while restricting the
// spell choice (CR 115.4).
func stackObjectSpellQualifiersMatch(g *game.Game, obj *game.StackObject, pred game.TargetPredicate) bool {
	if pred.ManaValue.Exists && obj.Kind == game.StackSpell {
		manaValue, ok := stackObjectManaValue(g, obj)
		if !ok || !pred.ManaValue.Val.Matches(manaValue) {
			return false
		}
	}
	hasSpellQualifier := len(pred.SpellCardTypes) > 0 ||
		len(pred.SpellCardTypesAny) > 0 ||
		len(pred.ExcludedSpellCardTypes) > 0 ||
		len(pred.SpellSupertypes) > 0 ||
		len(pred.SpellColors) > 0 ||
		len(pred.SpellExcludedColors) > 0 ||
		pred.SpellMulticolored ||
		pred.SpellColorless
	if !hasSpellQualifier || obj.Kind != game.StackSpell {
		return true
	}
	chars, ok := stackObjectSourceChars(g, obj)
	if !ok {
		return false
	}
	cardTypes := chars.types
	if obj.FaceDown {
		cardTypes = []types.Card{types.Creature}
	}
	for _, cardType := range pred.SpellCardTypes {
		if !slices.Contains(cardTypes, cardType) {
			return false
		}
	}
	if len(pred.SpellCardTypesAny) > 0 && !slices.ContainsFunc(pred.SpellCardTypesAny, func(cardType types.Card) bool {
		return slices.Contains(cardTypes, cardType)
	}) {
		return false
	}
	for _, cardType := range pred.ExcludedSpellCardTypes {
		if slices.Contains(cardTypes, cardType) {
			return false
		}
	}
	for _, supertype := range pred.SpellSupertypes {
		if !slices.Contains(stackSpellSupertypes(g, obj), supertype) {
			return false
		}
	}
	if pred.SpellColorless && len(chars.colors) != 0 {
		return false
	}
	for _, c := range pred.SpellColors {
		if !slices.Contains(chars.colors, c) {
			return false
		}
	}
	for _, c := range pred.SpellExcludedColors {
		if slices.Contains(chars.colors, c) {
			return false
		}
	}
	if pred.SpellMulticolored && len(chars.colors) < 2 {
		return false
	}
	return true
}

// stackObjectSourceHasTypes reports whether the stack object's source has every
// listed card type, preferring the source permanent's current types and falling
// back to its printed face when the source has left the battlefield.
func stackObjectSourceHasTypes(g *game.Game, obj *game.StackObject, required []types.Card) bool {
	if len(required) == 0 {
		return true
	}
	if perm, ok := g.PermanentByID(obj.SourceID); ok {
		for _, cardType := range required {
			if !permanentHasType(g, perm, cardType) {
				return false
			}
		}
		return true
	}
	def, ok := stackObjectSourceDef(g, obj)
	if !ok {
		return false
	}
	effectiveTypes := def.Types
	// CR 702.103b: a bestowed spell is not a creature spell while on the stack.
	if obj.Kind == game.StackSpell && obj.Bestowed {
		effectiveTypes = game.BestowSpellTypes(effectiveTypes)
	}
	for _, cardType := range required {
		if !slices.Contains(effectiveTypes, cardType) {
			return false
		}
	}
	return true
}

func stackSpellSupertypes(g *game.Game, obj *game.StackObject) []types.Super {
	if obj.SourceTokenDef != nil {
		def, ok := obj.SourceTokenDef.FaceDef(obj.Face)
		if !ok {
			return nil
		}
		return def.Supertypes
	}
	_, spellDef, ok := cardInstanceFaceDef(g, obj.SourceID, obj.Face)
	if !ok {
		return nil
	}
	return spellDef.Supertypes
}

func cardTargetMatchesSpec(g *game.Game, controller game.PlayerID, triggerEvent game.Event, spec *game.TargetSpec, target game.Target) bool {
	if !targetSpecAllowsCards(spec) {
		return false
	}
	cardID := target.CardID
	card, ok := g.GetCardInstance(cardID)
	if !ok {
		return false
	}
	if target.CardZoneVersionSet && card.ZoneVersion != target.CardZoneVersion {
		return false
	}
	if spec.TargetZone != zone.None {
		actualZone, ok := cardZone(g, cardID)
		if !ok || actualZone != spec.TargetZone {
			return false
		}
	}
	sel := targetSelection(spec)
	if !sel.Empty() {
		subject := selectionSubject{
			kind:       subjectCard,
			g:          g,
			card:       card,
			controller: card.Owner,
			viewer:     controller,
			event:      triggerEvent,
		}
		if !matchSelection(&subject, &sel) {
			return false
		}
	}
	return true
}

func playerTargetMatchesSpec(g *game.Game, controller game.PlayerID, spec *game.TargetSpec, playerID game.PlayerID) bool {
	if !isPlayerAlive(g, playerID) || !targetSpecAllowsPlayers(spec) {
		return false
	}
	sel := targetSelection(spec)
	if sel.Player != game.PlayerAny {
		return selectionPlayerRelationMatches(sel.Player, playerID, controller)
	}
	normalized := normalizedTargetConstraint(spec)
	if strings.Contains(normalized, "opponent") && playerID == controller {
		return false
	}
	return true
}

func permanentTargetMatchesSpec(g *game.Game, controller game.PlayerID, sourceObjectID id.ID, triggerEvent game.Event, spec *game.TargetSpec, permanentID id.ID) bool {
	if !targetSpecAllowsPermanents(spec) {
		return false
	}
	permanent, ok := permanentByObjectID(g, permanentID)
	if !ok || permanent.PhasedOut {
		return false
	}
	sel := targetSelection(spec)
	if !sel.Empty() {
		values := effectivePermanentValues(g, permanent)
		subject := selectionSubject{
			kind:           subjectPermanent,
			g:              g,
			permanent:      permanent,
			values:         values,
			viewer:         controller,
			sourceObjectID: sourceObjectID,
			event:          triggerEvent,
			clampPower:     true,
		}
		if sel.Controller != game.ControllerAny {
			subject.controller = effectiveController(g, permanent)
		}
		if !matchSelection(&subject, &sel) {
			return false
		}
	}
	if sel.Controller == game.ControllerAny && !permanentConstraintControllerMatches(g, controller, spec, permanent) {
		return false
	}
	if normalizedTargetConstraint(spec) == "any target" {
		return permanentHasType(g, permanent, types.Creature) ||
			permanentHasType(g, permanent, types.Planeswalker) ||
			permanentHasType(g, permanent, types.Battle)
	}
	return permanentTypeMatchesSpec(g, spec, permanent)
}

// targetSelection returns the Selection a TargetSpec matches against. Permanent,
// card, and player characteristics live solely on the spec's Selection; the
// legacy TargetPredicate now carries only stack-object and spell qualifiers.
func targetSelection(spec *game.TargetSpec) game.Selection {
	if spec.Selection.Exists {
		return spec.Selection.Val
	}
	return game.Selection{}
}

func combatStateMatches(g *game.Game, permanent *game.Permanent, filter game.CombatStateFilter) bool {
	if filter == game.CombatStateAny {
		return true
	}
	attacking := false
	blocking := false
	if g.Combat != nil {
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

func targetProtectedFromSource(g *game.Game, controller game.PlayerID, source *game.CardDef, sourceObjectID id.ID, target game.Target) bool {
	if target.Kind == game.TargetPlayer {
		if playerProtectedFromSource(g, target.PlayerID, 0, sourceObjectID, source) {
			return true
		}
		var sourceColors []color.Color
		if chars, ok := sourceCharsForProtection(g, 0, sourceObjectID, source); ok {
			sourceColors = chars.colors
		}
		return playerUntargetableByRuleEffect(g, controller, target.PlayerID, sourceColors)
	}
	if target.Kind != game.TargetPermanent {
		return false
	}
	permanent, ok := permanentByObjectID(g, target.PermanentID)
	if !ok {
		return false
	}
	if hasKeyword(g, permanent, game.Shroud) {
		return true
	}
	if effectiveController(g, permanent) != controller {
		if hasKeyword(g, permanent, game.Hexproof) {
			return true
		}
		if from := permanentHexproofFromColors(g, permanent); len(from) > 0 {
			if chars, ok := sourceCharsForProtection(g, 0, sourceObjectID, source); ok &&
				colorsIntersect(chars.colors, from) {
				return true
			}
		}
	}
	// Use effective source characteristics when the source is a permanent on
	// the battlefield (CR 702.16c).
	if sourceObjectID != 0 {
		if sourcePermanent, ok2 := permanentByObjectID(g, sourceObjectID); ok2 {
			return permanentProtectedFromPermanentEffective(g, permanent, sourcePermanent)
		}
		// Stack spell: use the selected face's characteristics.
		if stackObj, ok2 := stackObjectByID(g, sourceObjectID); ok2 {
			if chars, ok3 := stackObjectSourceChars(g, stackObj); ok3 {
				return permanentProtectedFromChars(g, permanent, chars)
			}
		}
		// LKI fallback: covers departed permanents and resolved spells.
		if snapshot, ok2 := lastKnownObject(g, sourceObjectID); ok2 {
			return permanentProtectedFromChars(g, permanent, sourceChars{
				colors:   snapshot.Colors,
				types:    snapshot.Types,
				subtypes: snapshot.Subtypes,
			})
		}
	}
	// Fall back to the supplied face def (LKI, spell during announcement, etc.).
	return source != nil && permanentProtectedFromSourceDef(g, permanent, source)
}

// permanentHexproofFromColors returns the union of colors named by every
// "hexproof from [colors]" ability currently effective on the permanent
// (CR 702.11e). It reads the effective ability list so grants removed by "loses
// all abilities" are excluded, mirroring permanentProtectedFromChars.
func permanentHexproofFromColors(g *game.Game, permanent *game.Permanent) []color.Color {
	values := effectivePermanentValues(g, permanent)
	if !values.keywords.has(game.HexproofFrom) {
		return nil
	}
	var colors []color.Color
	for i := range values.abilities {
		body, ok := values.abilities[i].(*game.StaticAbility)
		if !ok {
			continue
		}
		hexproof, ok := game.StaticBodyHexproofFromKeyword(body)
		if !ok {
			continue
		}
		for _, c := range hexproof.FromColors {
			if !slices.Contains(colors, c) {
				colors = append(colors, c)
			}
		}
	}
	return colors
}

func permanentConstraintControllerMatches(g *game.Game, controller game.PlayerID, spec *game.TargetSpec, permanent *game.Permanent) bool {
	permanentController := effectiveController(g, permanent)
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

func permanentTypeMatchesSpec(g *game.Game, spec *game.TargetSpec, permanent *game.Permanent) bool {
	if spec.Selection.Exists && normalizedTargetConstraint(spec) == "" {
		return true
	}
	sel := targetSelection(spec)
	if len(sel.RequiredTypes) > 0 || len(sel.RequiredTypesAny) > 0 || len(sel.ExcludedTypes) > 0 {
		return true
	}
	// A subtype filter narrows membership to permanents carrying that subtype,
	// which implies a card type. matchSelection has already enforced it above, so
	// a bare-subtype target ("target Soldier you control") that sets no card-type
	// constraint is satisfied here rather than falling through to constraint-string
	// type inference, which cannot recognize a subtype as a type.
	if len(sel.SubtypesAny) > 0 {
		return true
	}
	normalized := normalizedTargetConstraint(spec)
	if spec.Allow != game.TargetAllowUnspecified && normalized == "" {
		if spec.Allow&game.TargetAllowPlayer != 0 {
			return permanentHasType(g, permanent, types.Creature) ||
				permanentHasType(g, permanent, types.Planeswalker) ||
				permanentHasType(g, permanent, types.Battle)
		}
		return len(effectivePermanentValues(g, permanent).types) > 0
	}
	if strings.Contains(normalized, "nonland permanent") {
		return !permanentHasType(g, permanent, types.Land)
	}
	if strings.Contains(normalized, "permanent") && !containsAnyPermanentTypeConstraint(normalized) {
		return len(effectivePermanentValues(g, permanent).types) > 0
	}
	// A bare "token" target ("target token you control") carries no card-type
	// keyword and selects any token permanent. The TokenOnly predicate has
	// already been enforced through matchSelection, so any permanent with a type
	// satisfies the type check here.
	if strings.Contains(normalized, "token") && !containsAnyPermanentTypeConstraint(normalized) {
		return len(effectivePermanentValues(g, permanent).types) > 0
	}
	allowedTypes := permanentTypesForConstraint(normalized)
	if len(allowedTypes) == 0 {
		return false
	}
	return slices.ContainsFunc(allowedTypes, func(cardType types.Card) bool {
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

func permanentTypesForConstraint(normalized string) []types.Card {
	var cardTypes []types.Card
	if strings.Contains(normalized, "creature") {
		cardTypes = append(cardTypes, types.Creature)
	}
	if strings.Contains(normalized, "artifact") {
		cardTypes = append(cardTypes, types.Artifact)
	}
	if strings.Contains(normalized, "enchantment") {
		cardTypes = append(cardTypes, types.Enchantment)
	}
	if strings.Contains(normalized, "land") {
		cardTypes = append(cardTypes, types.Land)
	}
	if strings.Contains(normalized, "planeswalker") {
		cardTypes = append(cardTypes, types.Planeswalker)
	}
	if strings.Contains(normalized, "battle") {
		cardTypes = append(cardTypes, types.Battle)
	}
	return cardTypes
}

func normalizedTargetConstraint(spec *game.TargetSpec) string {
	normalized := strings.ToLower(strings.TrimSpace(spec.Constraint))
	normalized = strings.TrimPrefix(normalized, "target ")
	return strings.Join(strings.Fields(normalized), " ")
}

func isPlayerAlive(g *game.Game, playerID game.PlayerID) bool {
	if playerID < 0 || int(playerID) >= len(g.Players) {
		return false
	}
	player := g.Players[playerID]
	return !player.Eliminated && !g.TurnOrder.IsEliminated(playerID)
}
