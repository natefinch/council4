package rules

import (
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

type conditionContext struct {
	controller             game.PlayerID
	source                 *game.Permanent
	event                  *game.Event
	obj                    *game.StackObject
	useBaseCharacteristics bool
	characteristicsBefore  game.ContinuousLayer
}

// conditionParametersNegative reports whether any numeric condition parameter is
// negative, which is structurally invalid and must fail closed.
func conditionParametersNegative(cond *game.Condition) bool {
	return cond.ControllerLifeAtLeast < 0 ||
		cond.AnyPlayerLifeAtMost < 0 ||
		cond.AnyOpponentPoisonAtLeast < 0 ||
		cond.ControllerHandSizeExactly.Exists && cond.ControllerHandSizeExactly.Val < 0 ||
		cond.OpponentCountAtLeast < 0 ||
		cond.ControllerGraveyardCardCountAtLeast < 0 ||
		cond.ControllerGraveyardCardTypeCountAtLeast < 0 ||
		cond.ControllerBasicLandTypeCountAtLeast < 0 ||
		cond.ControllerCreaturePowerDiversityAtLeast < 0 ||
		cond.ControllerControls.MinCount < 0 ||
		cond.ControlsMatching.Exists && cond.ControlsMatching.Val.MinCount < 0 ||
		cond.AnyOpponentControls.Exists && cond.AnyOpponentControls.Val.MinCount < 0 ||
		cond.OpponentsControl.Exists && cond.OpponentsControl.Val.MinCount < 0
}

func conditionSatisfied(g *game.Game, ctx conditionContext, condition opt.V[game.Condition]) bool {
	if !condition.Exists || condition.Val.Empty() {
		return true
	}
	cond := condition.Val
	if conditionParametersNegative(&cond) {
		return false
	}
	matches := true
	if cond.ControlsMatching.Exists {
		matches = matches && controllerControlsMatchingSelection(g, ctx, cond.ControlsMatching.Val)
	} else if !cond.ControllerControls.Empty() {
		matches = matches && controllerControlsMatchingSelection(g, ctx, controlSelectionFromFilter(cond.ControllerControls))
	}
	if cond.ControllerLifeAtLeast > 0 {
		player, ok := playerByID(g, ctx.controller)
		matches = matches && ok && player.Life >= cond.ControllerLifeAtLeast
	}
	if cond.ControllerHandSizeAtLeast > 0 {
		player, ok := playerByID(g, ctx.controller)
		matches = matches && ok && cardInstanceCount(g, player.Hand.All()) >= cond.ControllerHandSizeAtLeast
	}
	if cond.ControllerHandSizeExactly.Exists {
		player, ok := playerByID(g, ctx.controller)
		matches = matches && ok && cardInstanceCount(g, player.Hand.All()) == cond.ControllerHandSizeExactly.Val
	}
	if cond.AnyOpponentPoisonAtLeast > 0 {
		matches = matches && anyOpponentPoisonAtLeast(g, ctx.controller, cond.AnyOpponentPoisonAtLeast)
	}
	if cond.AnyPlayerLifeAtMost > 0 {
		matches = matches && anyPlayerLifeAtMost(g, cond.AnyPlayerLifeAtMost)
	}
	if cond.OpponentCountAtLeast > 0 {
		matches = matches && len(aliveOpponents(g, ctx.controller)) >= cond.OpponentCountAtLeast
	}
	if cond.ControllerHandEmpty {
		player, ok := playerByID(g, ctx.controller)
		matches = matches && ok && cardInstanceCount(g, player.Hand.All()) == 0
	}
	if cond.ControllerGraveyardCardCountAtLeast > 0 {
		player, ok := playerByID(g, ctx.controller)
		matches = matches && ok && cardInstanceCount(g, player.Graveyard.All()) >= cond.ControllerGraveyardCardCountAtLeast
	}
	if cond.ControllerGraveyardCardTypeCountAtLeast > 0 {
		matches = matches && controllerGraveyardCardTypeCount(g, ctx.controller) >= cond.ControllerGraveyardCardTypeCountAtLeast
	}
	if cond.ControllerBasicLandTypeCountAtLeast > 0 {
		matches = matches && controllerBasicLandTypeCount(g, ctx) >= cond.ControllerBasicLandTypeCountAtLeast
	}
	if cond.ControllerCreaturePowerDiversityAtLeast > 0 {
		matches = matches && controllerCreaturePowerDiversity(g, ctx) >= cond.ControllerCreaturePowerDiversityAtLeast
	}
	if cond.AnyOpponentControls.Exists {
		matches = matches && anyOpponentControlsMatchingSelection(g, ctx, cond.AnyOpponentControls.Val)
	}
	if cond.OpponentsControl.Exists {
		matches = matches && playersControlMatchingSelection(g, ctx, aliveOpponents(g, ctx.controller), cond.OpponentsControl.Val)
	}
	if cond.ControlComparison.Exists {
		matches = matches && controlCountComparisonSatisfied(g, ctx, cond.ControlComparison.Val)
	}
	if cond.Object.Exists || cond.ObjectMatches.Exists || len(cond.Types) > 0 {
		matches = matches && conditionObjectMatches(g, ctx, &cond)
	}
	if cond.EventPermanentNameUniqueAmongControlledAndGraveyardCreatures {
		matches = matches && eventPermanentNameUniqueAmongControlledAndGraveyardCreatures(g, ctx)
	}
	if cond.SourceClassLevelAtLeast > 0 {
		matches = matches && ctx.source != nil && ctx.source.ClassLevel >= cond.SourceClassLevelAtLeast
	}
	if cond.SourceClassLevelLessThan > 0 {
		matches = matches && ctx.source != nil && ctx.source.ClassLevel < cond.SourceClassLevelLessThan
	}
	if cond.SourceNotMonstrous {
		matches = matches && ctx.source != nil && !ctx.source.Monstrous
	}
	if cond.SourceTributeNotPaid {
		matches = matches && ctx.source != nil && !ctx.source.TributePaid
	}
	if cond.ControllerHasMaxSpeed {
		player, ok := playerByID(g, ctx.controller)
		matches = matches && ok && player.Speed >= 4
	}
	if cond.TargetEnteredThisTurn.Exists {
		matches = matches && conditionTargetEnteredThisTurn(g, ctx, cond.TargetEnteredThisTurn.Val)
	}
	if cond.CastFromZone.Exists {
		matches = matches && ctx.obj != nil && !ctx.obj.Copy && ctx.obj.SourceZone == cond.CastFromZone.Val
	}
	if cond.CastDuringControllerMainPhase {
		matches = matches && ctx.obj != nil && !ctx.obj.Copy && ctx.obj.CastDuringControllerMainPhase
	}
	if cond.SpellWasKicked {
		matches = matches && ctx.obj != nil && !ctx.obj.Copy && ctx.obj.KickerPaid
	}
	if cond.ControllerCreatedTokenThisTurn {
		matches = matches && controllerCreatedTokenThisTurn(g, ctx.controller)
	}
	if cond.EventHistory.Exists {
		matches = matches && conditionEventHistorySatisfied(g, ctx, &cond.EventHistory.Val)
	}
	if cond.ControllerControlsCommander {
		matches = matches && playerControlsCommander(g, ctx.controller)
	}
	if cond.Negate {
		return !matches
	}
	return matches
}

func cardInstanceCount(g *game.Game, objectIDs []id.ID) int {
	count := 0
	for _, objectID := range objectIDs {
		if _, ok := g.GetCardInstance(objectID); ok {
			count++
		}
	}
	return count
}

func controllerGraveyardCardTypeCount(g *game.Game, controller game.PlayerID) int {
	player, ok := playerByID(g, controller)
	if !ok {
		return 0
	}
	distinct := make(map[types.Card]bool)
	for _, cardID := range player.Graveyard.All() {
		card, ok := g.GetCardInstance(cardID)
		if !ok {
			continue
		}
		for _, cardType := range cardFaceOrDefault(card, game.FaceFront).Types {
			distinct[cardType] = true
		}
		if card.Def.Layout == game.LayoutSplit && card.Def.Alternate.Exists {
			for _, cardType := range card.Def.Alternate.Val.Types {
				distinct[cardType] = true
			}
		}
	}
	return len(distinct)
}

func controllerBasicLandTypeCount(g *game.Game, ctx conditionContext) int {
	basicLandTypes := [...]types.Sub{types.Plains, types.Island, types.Swamp, types.Mountain, types.Forest}
	distinct := make(map[types.Sub]bool)
	for _, permanent := range g.Battlefield {
		if permanent.PhasedOut {
			continue
		}
		values := permanentValuesForCondition(g, permanent, ctx)
		if values.controller != ctx.controller || !slices.Contains(values.types, types.Land) {
			continue
		}
		for _, subtype := range basicLandTypes {
			if slices.Contains(values.subtypes, subtype) {
				distinct[subtype] = true
			}
		}
	}
	return len(distinct)
}

func controllerCreaturePowerDiversity(g *game.Game, ctx conditionContext) int {
	distinct := make(map[int]bool)
	for _, permanent := range g.Battlefield {
		if permanent.PhasedOut {
			continue
		}
		values := permanentValuesForCondition(g, permanent, ctx)
		if values.controller == ctx.controller &&
			slices.Contains(values.types, types.Creature) &&
			values.powerOK {
			distinct[values.power] = true
		}
	}
	return len(distinct)
}

func permanentValuesForCondition(g *game.Game, permanent *game.Permanent, ctx conditionContext) permanentEffectiveValues {
	switch {
	case ctx.useBaseCharacteristics:
		return basePermanentValues(g, permanent)
	case ctx.characteristicsBefore != 0:
		values := permanentValuesBeforeLayer(g, permanent, ctx.characteristicsBefore)
		applyCounterAndTemporaryValues(permanent, &values)
		return values
	default:
		return effectivePermanentValues(g, permanent)
	}
}

func conditionObjectMatches(g *game.Game, ctx conditionContext, cond *game.Condition) bool {
	if cond == nil || cond.ObjectMatches.Exists && !cond.Object.Exists {
		return false
	}
	obj := ctx.obj
	if obj == nil && (ctx.event != nil || ctx.source != nil) {
		obj = &game.StackObject{Controller: ctx.controller}
		if ctx.event != nil {
			obj.HasTriggerEvent = true
			obj.TriggerEvent = *ctx.event
		}
		if ctx.source != nil {
			obj.SourceID = ctx.source.ObjectID
		}
	}
	ref := game.EventPermanentReference()
	if cond.Object.Exists {
		ref = cond.Object.Val
	}
	if obj == nil {
		return false
	}
	resolved, ok := resolveObjectReference(g, obj, ref)
	if !ok {
		return false
	}
	if cond.ObjectMatches.Exists &&
		!resolvedObjectMatchesConditionSelection(g, ctx, &resolved, &cond.ObjectMatches.Val) {
		return false
	}
	for _, cardType := range cond.Types {
		if !resolvedObjectHasType(g, &resolved, cardType) {
			return false
		}
	}
	return true
}

func resolvedObjectMatchesConditionSelection(
	g *game.Game,
	ctx conditionContext,
	resolved *resolvedObjectReference,
	selection *game.Selection,
) bool {
	if resolved == nil || selection == nil {
		return false
	}
	if resolved.permanent != nil {
		values := permanentValuesForCondition(g, resolved.permanent, ctx)
		subject := selectionSubject{
			kind:      subjectPermanent,
			g:         g,
			permanent: resolved.permanent,
			values:    &values,
			viewer:    ctx.controller,
		}
		if selection.Controller != game.ControllerAny {
			subject.controller = values.controller
		}
		if ctx.source != nil {
			subject.sourceObjectID = ctx.source.ObjectID
		}
		return matchSelection(&subject, selection)
	}
	if resolved.stack != nil {
		colors, ok := stackObjectColors(g, resolved.stack)
		if !ok {
			return false
		}
		subject := selectionSubject{
			kind:   subjectCastSpell,
			g:      g,
			event:  game.Event{Colors: colors},
			viewer: ctx.controller,
		}
		return matchSelection(&subject, selection)
	}
	if resolved.snapshot.ObjectID == 0 {
		return false
	}
	subject := selectionSubject{
		kind:   subjectEventPermanent,
		g:      g,
		event:  game.Event{PermanentID: resolved.snapshot.ObjectID},
		viewer: ctx.controller,
	}
	if selection.Controller != game.ControllerAny {
		subject.controller = resolved.snapshot.Controller
	}
	if ctx.source != nil {
		subject.sourceObjectID = ctx.source.ObjectID
	}
	return matchSelection(&subject, selection)
}

func resolvedObjectHasType(g *game.Game, resolved *resolvedObjectReference, cardType types.Card) bool {
	if resolved.permanent != nil {
		return permanentHasType(g, resolved.permanent, cardType)
	}
	return slices.Contains(resolved.snapshot.Types, cardType)
}

func controlSelectionFromFilter(filter game.PermanentFilter) game.SelectionCount {
	return game.SelectionCount{
		Selection:  filter.Selection(),
		MinCount:   filter.MinCount,
		TotalPower: filter.TotalPower,
	}
}

func controllerControlsMatchingSelection(g *game.Game, ctx conditionContext, control game.SelectionCount) bool {
	return playersControlMatchingSelection(g, ctx, []game.PlayerID{ctx.controller}, control)
}

func anyOpponentControlsMatchingSelection(g *game.Game, ctx conditionContext, control game.SelectionCount) bool {
	for _, opponent := range aliveOpponents(g, ctx.controller) {
		if playersControlMatchingSelection(g, ctx, []game.PlayerID{opponent}, control) {
			return true
		}
	}
	return false
}

func playersControlMatchingSelection(g *game.Game, ctx conditionContext, controllers []game.PlayerID, control game.SelectionCount) bool {
	if control.MinCount < 0 {
		return false
	}
	want := control.MinCount
	if want <= 0 {
		want = 1
	}
	allowed := make(map[game.PlayerID]bool, len(controllers))
	for _, controller := range controllers {
		allowed[controller] = true
	}
	count := 0
	totalPower := 0
	sel := control.Selection
	var distinctNames map[string]bool
	if control.DistinctNames.Exists {
		distinctNames = make(map[string]bool)
	}
	for _, permanent := range g.Battlefield {
		if permanent.PhasedOut {
			continue
		}
		if ctx.useBaseCharacteristics {
			if !allowed[permanent.Controller] {
				continue
			}
		}
		values := permanentValuesForCondition(g, permanent, ctx)
		if !ctx.useBaseCharacteristics && !allowed[values.controller] {
			continue
		}
		subject := selectionSubject{
			kind:      subjectPermanent,
			g:         g,
			permanent: permanent,
			values:    &values,
			viewer:    ctx.controller,
			useBase:   ctx.useBaseCharacteristics,
		}
		if sel.Controller != game.ControllerAny {
			if ctx.useBaseCharacteristics {
				subject.controller = permanent.Controller
			} else {
				subject.controller = values.controller
			}
		}
		if ctx.source != nil {
			subject.sourceObjectID = ctx.source.ObjectID
		}
		if !matchSelection(&subject, &sel) {
			continue
		}
		count++
		if control.TotalPower.Exists {
			powerValues := &values
			if ctx.useBaseCharacteristics {
				effective := effectivePermanentValues(g, permanent)
				powerValues = &effective
			}
			if powerValues.powerOK {
				totalPower += powerValues.power
			}
		}
		if distinctNames != nil && values.name != "" {
			distinctNames[values.name] = true
		}
		if count >= want && !control.DistinctNames.Exists {
			if !control.TotalPower.Exists || control.TotalPower.Val.Matches(totalPower) {
				return true
			}
		}
	}
	if count < want {
		return false
	}
	if control.TotalPower.Exists && !control.TotalPower.Val.Matches(totalPower) {
		return false
	}
	if control.DistinctNames.Exists && !control.DistinctNames.Val.Matches(len(distinctNames)) {
		return false
	}
	return control.TotalPower.Exists || control.DistinctNames.Exists
}

// countPlayerMatchingSelection counts permanents matching sel controlled by the
// given player, mirroring the subject construction of
// playersControlMatchingSelection without any count threshold.
func countPlayerMatchingSelection(g *game.Game, ctx conditionContext, player game.PlayerID, sel game.Selection) int {
	count := 0
	for _, permanent := range g.Battlefield {
		if permanent.PhasedOut {
			continue
		}
		if ctx.useBaseCharacteristics {
			if permanent.Controller != player {
				continue
			}
		}
		values := permanentValuesForCondition(g, permanent, ctx)
		if !ctx.useBaseCharacteristics && values.controller != player {
			continue
		}
		subject := selectionSubject{
			kind:      subjectPermanent,
			g:         g,
			permanent: permanent,
			values:    &values,
			viewer:    ctx.controller,
			useBase:   ctx.useBaseCharacteristics,
		}
		if sel.Controller != game.ControllerAny {
			if ctx.useBaseCharacteristics {
				subject.controller = permanent.Controller
			} else {
				subject.controller = values.controller
			}
		}
		if ctx.source != nil {
			subject.sourceObjectID = ctx.source.ObjectID
		}
		if matchSelection(&subject, &sel) {
			count++
		}
	}
	return count
}

// controlCountComparisonSatisfied evaluates a cross-player control-count
// comparison. The opponent-scoped side is quantified existentially
// (ControlPlayerAnyOpponent) or universally (ControlPlayerEachOpponent); with no
// opponents the universal form is vacuously true and the existential form false.
// A ControlPlayerTriggeringPlayer side names the specific player tied to the
// triggering event, so the comparison resolves against that one player.
func controlCountComparisonSatisfied(g *game.Game, ctx conditionContext, cmp game.ControlCountComparison) bool {
	youCount := countPlayerMatchingSelection(g, ctx, ctx.controller, cmp.Selection)
	if cmp.Left == game.ControlPlayerTriggeringPlayer || cmp.Right == game.ControlPlayerTriggeringPlayer {
		if ctx.event == nil {
			return false
		}
		triggeringCount := countPlayerMatchingSelection(g, ctx, ctx.event.Controller, cmp.Selection)
		left := youCount
		if cmp.Left != game.ControlPlayerController {
			left = triggeringCount
		}
		right := youCount
		if cmp.Right != game.ControlPlayerController {
			right = triggeringCount
		}
		return compare.Int{Op: cmp.Op, Value: right}.Matches(left)
	}
	universal := cmp.Left == game.ControlPlayerEachOpponent || cmp.Right == game.ControlPlayerEachOpponent
	opponents := aliveOpponents(g, ctx.controller)
	if len(opponents) == 0 {
		return universal
	}
	for _, opponent := range opponents {
		opponentCount := countPlayerMatchingSelection(g, ctx, opponent, cmp.Selection)
		left := youCount
		if cmp.Left != game.ControlPlayerController {
			left = opponentCount
		}
		right := youCount
		if cmp.Right != game.ControlPlayerController {
			right = opponentCount
		}
		satisfied := compare.Int{Op: cmp.Op, Value: right}.Matches(left)
		if universal && !satisfied {
			return false
		}
		if !universal && satisfied {
			return true
		}
	}
	return universal
}
func anyPlayerLifeAtMost(g *game.Game, maximum int) bool {
	for playerID := range game.PlayerID(game.NumPlayers) {
		player, ok := playerByID(g, playerID)
		if ok && !player.Eliminated && player.Life <= maximum {
			return true
		}
	}
	return false
}

func anyOpponentPoisonAtLeast(g *game.Game, controller game.PlayerID, minimum int) bool {
	for _, opponent := range aliveOpponents(g, controller) {
		player, ok := playerByID(g, opponent)
		if ok && player.PoisonCounters >= minimum {
			return true
		}
	}
	return false
}

func conditionTargetEnteredThisTurn(g *game.Game, ctx conditionContext, targetIndex int) bool {
	if ctx.obj == nil {
		return false
	}
	permanent, ok := effectPermanentAt(g, ctx.obj, targetIndex)
	if !ok {
		return false
	}
	return permanentEnteredThisTurn(g, permanent.ObjectID)
}

// permanentEnteredThisTurn reports whether the permanent identified by id entered
// the battlefield during the current turn, scanning this turn's events for its
// enter-the-battlefield event.
func permanentEnteredThisTurn(g *game.Game, permanentID id.ID) bool {
	for _, event := range g.EventsThisTurn() {
		if event.Kind == game.EventPermanentEnteredBattlefield && event.PermanentID == permanentID {
			return true
		}
	}
	return false
}

func eventPermanentNameUniqueAmongControlledAndGraveyardCreatures(g *game.Game, ctx conditionContext) bool {
	if ctx.event == nil || ctx.event.PermanentID == 0 {
		return false
	}
	resolved, ok := resolvePermanentOrLastKnown(g, ctx.event.PermanentID)
	if !ok {
		return false
	}
	name := resolvedObjectName(g, &resolved)
	if name == "" {
		return false
	}
	for _, permanent := range g.Battlefield {
		if !activeBattlefieldPermanent(permanent) ||
			permanent.ObjectID == ctx.event.PermanentID ||
			effectiveController(g, permanent) != ctx.controller ||
			!permanentHasType(g, permanent, types.Creature) {
			continue
		}
		if def, ok := permanentCardDef(g, permanent); ok && def.Name == name {
			return false
		}
	}
	player, ok := playerByID(g, ctx.controller)
	if !ok {
		return false
	}
	for _, cardID := range player.Graveyard.All() {
		card, ok := g.GetCardInstance(cardID)
		if !ok {
			continue
		}
		def := cardFaceOrDefault(card, game.FaceFront)
		if def.Name == name && def.HasType(types.Creature) {
			return false
		}
	}
	return true
}

func resolvedObjectName(g *game.Game, resolved *resolvedObjectReference) string {
	if resolved.permanent != nil {
		if resolved.permanent.Token {
			return permanentTokenName(resolved.permanent)
		}
		return permanentEffectiveName(g, resolved.permanent)
	}
	return resolved.snapshot.Name
}

func activationConditionSatisfied(g *game.Game, playerID game.PlayerID, permanent *game.Permanent, condition opt.V[game.Condition]) bool {
	return conditionSatisfied(g, conditionContext{
		controller: playerID,
		source:     permanent,
	}, condition)
}

// controllerCreatedTokenThisTurn reports whether the controller created at least
// one token during the current turn. Token creation is recorded in the per-turn
// event log as a battlefield-entry event carrying the token's definition, so the
// scan reuses that log rather than tracking separate mutable state.
func controllerCreatedTokenThisTurn(g *game.Game, controller game.PlayerID) bool {
	for _, event := range g.EventsThisTurn() {
		if event.Kind == game.EventPermanentEnteredBattlefield &&
			event.TokenDef != nil &&
			event.Controller == controller {
			return true
		}
	}
	return false
}

// conditionEventHistorySatisfied returns true when the chosen turn's event
// history contains at least hist.MinCount events matching hist.Pattern (at least
// one when MinCount is zero). The source permanent is passed to
// triggerMatchesEvent so controller-relative filters (TriggerControllerYou,
// TriggerPlayerYou, etc.) resolve correctly. A nil source permanent fails closed
// for any pattern that references the source (such filters can never match
// without one); source-agnostic patterns, such as "a creature died this turn"
// gating a resolving instant, evaluate against the event log directly.
func conditionEventHistorySatisfied(g *game.Game, ctx conditionContext, hist *game.EventHistoryCondition) bool {
	if ctx.source == nil && eventHistoryPatternNeedsSource(&hist.Pattern) {
		return false
	}
	var events []game.Event
	switch hist.Window {
	case game.EventHistoryCurrentTurn:
		events = g.EventsThisTurn()
	case game.EventHistoryPreviousTurn:
		events = g.EventsPreviousTurn()
	default:
		return false
	}
	want := max(hist.MinCount, 1)
	matches := 0
	for _, event := range events {
		if triggerMatchesEvent(g, ctx.source, &hist.Pattern, event) {
			matches++
			if matches >= want {
				return true
			}
		}
	}
	return false
}

// eventHistoryPatternNeedsSource reports whether a trigger pattern consults the
// ability's source permanent (its controller, identity, or attachment) to match
// an event. Such patterns can never match without a source, so a source-agnostic
// caller (a resolving instant gating on event history) fails them closed instead
// of dereferencing a nil source.
func eventHistoryPatternNeedsSource(pattern *game.TriggerPattern) bool {
	return pattern.Controller != game.TriggerControllerAny ||
		pattern.CauseController != game.TriggerControllerAny ||
		pattern.Player != game.TriggerPlayerAny ||
		pattern.Source != game.TriggerSourceAny ||
		pattern.ExcludeSelf ||
		pattern.SubjectSelectionOrSelf ||
		pattern.DamageRecipientIsSource ||
		pattern.SpellTargetsSource ||
		!pattern.StepPlayerSourceAttachedSelection.Empty()
}
