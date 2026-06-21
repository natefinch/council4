package rules

import (
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func (e *Engine) triggerTargets(g *game.Game, controller game.PlayerID, source *game.CardDef, sourceObjectID id.ID, ability *game.TriggeredAbility, chosenModes []int, agents [game.NumPlayers]PlayerAgent, log *TurnLog) ([]game.Target, bool) {
	result := targetChoicesForBodyFromSourceObjectWithModes(g, controller, source, sourceObjectID, ability, chosenModes)
	switch result.kind {
	case targetNoLegalChoices, targetInvalidSpec:
		return nil, false
	}
	choices := result.choices
	if len(choices) == 1 {
		return append([]game.Target(nil), choices[0]...), true
	}
	selected := e.chooseChoice(g, agents, targetChoiceRequest(controller, "Choose triggered ability targets.", choices), log)
	if len(selected) != 1 || selected[0] < 0 || selected[0] >= len(choices) {
		return append([]game.Target(nil), choices[0]...), true
	}
	return append([]game.Target(nil), choices[selected[0]]...), true
}

func triggerMatchesEvent(g *game.Game, source *game.Permanent, pattern *game.TriggerPattern, event game.Event) bool {
	if pattern.Event == game.EventUnknown || !patternMatchesEventKind(pattern, event.Kind) {
		return false
	}
	// Payment-time mana activations do not emit this event yet, so an
	// unrestricted pattern would silently under-trigger.
	if pattern.Event == game.EventAbilityActivated && !pattern.ExcludeManaAbility {
		return false
	}
	if pattern.Event == game.EventZoneChanged && event.PermanentID == 0 {
		return false
	}
	if pattern.RequireTappedForMana && !event.TappedForMana {
		return false
	}

	// Trigger patterns are checked when the triggering event is processed, and
	// LTB/dies checks may need last-known information for the moved permanent
	// (CR 603.2, CR 603.6c, CR 603.10). A nil source has no controller; that
	// only reaches here for source-agnostic event-history conditions evaluated
	// during a spell's resolution, where source-relative filters fail closed via
	// the out-of-range controller sentinel below.
	sourceController := game.PlayerID(-1)
	if source != nil {
		sourceController = effectiveController(g, source)
	}
	var sourceObjectID id.ID
	if source != nil {
		sourceObjectID = source.ObjectID
	}
	subjectController := event.Controller
	if subject, ok := triggerSubjectPermanent(g, pattern.Subject, event); ok {
		subjectController = effectiveController(g, subject)
	}
	if !triggerControllerMatches(sourceController, pattern.Controller, subjectController) {
		return false
	}
	if !triggerControllerMatches(sourceController, pattern.CauseController, event.Controller) {
		return false
	}
	if !triggerSourceMatches(g, source, pattern.Source, pattern.Subject, event) {
		return false
	}
	if pattern.ExcludeSelf && triggerSourceMatches(g, source, game.TriggerSourceSelf, pattern.Subject, event) {
		return false
	}
	if !triggerPlayerMatches(sourceController, pattern.Player, event.Player) {
		return false
	}
	if !pattern.StepPlayerSourceAttachedSelection.Empty() &&
		!stepPlayerSourceAttachedMatches(g, sourceController, source, event, &pattern.StepPlayerSourceAttachedSelection) {
		return false
	}
	if pattern.MatchFromZone && pattern.FromZone != event.FromZone {
		return false
	}
	if pattern.MatchToZone && pattern.ToZone != event.ToZone {
		return false
	}
	if pattern.ExcludeToZone && pattern.ToZone == event.ToZone {
		return false
	}
	if pattern.MatchFaceDown && pattern.FaceDown != event.FaceDown {
		return false
	}
	if pattern.RequireKickerPaid && !event.KickerPaid {
		return false
	}
	if pattern.RequireHistoric && !eventSpellHistoric(event) {
		return false
	}
	if pattern.ExcludeManaAbility && event.ManaAbility {
		return false
	}
	cardSel := triggerCardSelection(pattern)
	filteredSpellOrdinal := pattern.Event == game.EventSpellCast && !cardSel.Empty()
	if pattern.PlayerEventOrdinalThisTurn > 0 && !filteredSpellOrdinal &&
		pattern.PlayerEventOrdinalThisTurn != event.PlayerEventOrdinalThisTurn {
		return false
	}
	if !triggerCombatPatternMatches(g, sourceController, source, pattern, event) {
		return false
	}
	if !triggerAttackerCountMatches(g, pattern, event) {
		return false
	}
	if pattern.MatchCounterKind && pattern.CounterKind != event.CounterKind {
		return false
	}
	if pattern.Event == game.EventBeginningOfStep {
		if pattern.Step == game.StepNone || pattern.Step != event.Step {
			return false
		}
	}
	if subjectSel := triggerSubjectSelection(pattern); !subjectSel.Empty() {
		matched := triggerSelectionMatches(g, sourceController, event, event.PermanentID, &subjectSel, sourceObjectID)
		if !matched && pattern.SubjectSelectionOrSelf {
			matched = triggerSourceMatches(g, source, game.TriggerSourceSelf, pattern.Subject, event)
		}
		if !matched {
			return false
		}
	}
	if !cardSel.Empty() {
		subject := selectionSubject{
			kind:           subjectCastSpell,
			g:              g,
			event:          event,
			cardTypes:      eventSpellCardTypes(g, event),
			sourceObjectID: sourceObjectID,
		}
		if !matchSelection(&subject, &cardSel) {
			return false
		}
		if pattern.PlayerEventOrdinalThisTurn > 0 &&
			pattern.PlayerEventOrdinalThisTurn != filteredSpellCastOrdinalThisTurn(g, event, &cardSel) {
			return false
		}
	}
	if pattern.MatchStackObjectKind && event.Kind == pattern.Event &&
		!eventStackObjectKindMatches(g, event, pattern.StackObjectKind) {
		return false
	}
	if pattern.SpellTargetsSource && !spellTargetsSource(g, source, event) {
		return false
	}
	if pattern.SpellTargetPattern.Exists && !spellTargetsPattern(g, sourceController, pattern.SpellTargetAllow, pattern.SpellTargetPattern.Val, event) {
		return false
	}
	return true
}

func filteredSpellCastOrdinalThisTurn(g *game.Game, event game.Event, selection *game.Selection) int {
	start := 0
	index := g.Turn.TurnNumber - 1
	if index >= 0 && index < len(g.EventTurnStarts) {
		start = g.EventTurnStarts[index]
	}
	ordinal := 0
	for _, candidate := range g.Events[start:] {
		if candidate.Kind != game.EventSpellCast || candidate.Controller != event.Controller {
			continue
		}
		subject := selectionSubject{
			kind:      subjectCastSpell,
			g:         g,
			event:     candidate,
			cardTypes: eventSpellCardTypes(g, candidate),
		}
		if matchSelection(&subject, selection) {
			ordinal++
		}
		if candidate.PlayerEventOrdinalThisTurn == event.PlayerEventOrdinalThisTurn {
			break
		}
	}
	return ordinal
}

// patternMatchesEventKind reports whether the pattern's event family covers the
// given event kind. A spell-cast pattern with MatchSpellCopy also covers
// EventSpellCopied so "cast or copy" (magecraft) triggers fire on spell copies
// without widening ordinary cast-only triggers.
func patternMatchesEventKind(pattern *game.TriggerPattern, kind game.EventKind) bool {
	if pattern.Event == kind {
		return true
	}
	if pattern.UnionEvent != game.EventUnknown && pattern.UnionEvent == kind {
		return true
	}
	return pattern.MatchSpellCopy &&
		pattern.Event == game.EventSpellCast &&
		kind == game.EventSpellCopied
}

func triggerCombatPatternMatches(g *game.Game, viewer game.PlayerID, source *game.Permanent, pattern *game.TriggerPattern, event game.Event) bool {
	if pattern.DamageRecipient != game.DamageRecipientNone && pattern.DamageRecipient&event.DamageRecipient == 0 {
		return false
	}
	if pattern.DamageRecipientIsSource && !damageRecipientIsSource(source, event) {
		return false
	}
	if pattern.RequireCombatDamage && !event.CombatDamage {
		return false
	}
	if pattern.RequireNonCombatDamage && event.CombatDamage {
		return false
	}
	if !attackRecipientMatches(pattern.AttackRecipient, event) ||
		!attackRecipientSelectionMatches(g, viewer, &pattern.AttackRecipientSelection, event) ||
		!damageRecipientTypesMatch(g, pattern.DamageRecipientTypes, event) {
		return false
	}
	if pattern.AttackedPlayerHasMostLife && !attackedPlayerHasMostLife(g, event) {
		return false
	}
	if !pattern.DamageRecipientSelection.Empty() &&
		event.DamageRecipient == game.DamageRecipientPermanent &&
		!triggerSelectionMatches(g, viewer, event, event.PermanentID, &pattern.DamageRecipientSelection, source.ObjectID) {
		return false
	}
	if !pattern.DamageSourceSelection.Empty() &&
		!triggerSelectionMatches(g, viewer, event, event.SourceObjectID, &pattern.DamageSourceSelection, source.ObjectID) {
		return false
	}
	if !pattern.RelatedSubjectSelection.Empty() &&
		!triggerSelectionMatches(g, viewer, event, event.RelatedPermanentID, &pattern.RelatedSubjectSelection, source.ObjectID) {
		return false
	}
	if pattern.DamageRecipientCombatState == game.CombatStateAny {
		return true
	}
	permanent, ok := permanentByObjectID(g, event.PermanentID)
	return event.DamageRecipient == game.DamageRecipientPermanent &&
		ok &&
		combatStateMatches(g, permanent, pattern.DamageRecipientCombatState)
}

// triggerAttackerCountMatches enforces attacker-count combat relations against
// the attackers declared this combat. "Attacks alone" requires exactly one
// attacker, and AttackerCountAtLeast requires at least that many. The full
// declaration is recorded in g.Combat.Attackers before any attacker-declared
// event is processed, so the count is authoritative when the trigger matches.
func triggerAttackerCountMatches(g *game.Game, pattern *game.TriggerPattern, event game.Event) bool {
	if !pattern.AttackAlone && pattern.AttackerCountAtLeast == 0 {
		return true
	}
	if event.Kind != game.EventAttackerDeclared || g.Combat == nil {
		return false
	}
	attackers := len(g.Combat.Attackers)
	if pattern.AttackAlone && attackers != 1 {
		return false
	}
	return pattern.AttackerCountAtLeast == 0 || attackers >= pattern.AttackerCountAtLeast
}

func damageRecipientIsSource(source *game.Permanent, event game.Event) bool {
	return source.ObjectID != 0 && event.PermanentID == source.ObjectID ||
		source.CardInstanceID != 0 && event.CardID == source.CardInstanceID
}

func attackRecipientSelectionMatches(g *game.Game, viewer game.PlayerID, selection *game.Selection, event game.Event) bool {
	if selection.Empty() {
		return true
	}
	recipientID := event.AttackTarget.PlaneswalkerID
	if recipientID == 0 {
		recipientID = event.AttackTarget.BattleID
	}
	return recipientID == 0 || triggerSelectionMatches(g, viewer, event, recipientID, selection, 0)
}

func damageRecipientTypesMatch(g *game.Game, required []types.Card, event game.Event) bool {
	if len(required) == 0 {
		return true
	}
	if event.DamageRecipient != game.DamageRecipientPermanent {
		return false
	}
	for _, cardType := range required {
		if !eventPermanentHasType(g, event, cardType) {
			return false
		}
	}
	return true
}

func stepPlayerSourceAttachedMatches(g *game.Game, viewer game.PlayerID, source *game.Permanent, event game.Event, selection *game.Selection) bool {
	if !source.AttachedTo.Exists {
		return false
	}
	attached, ok := resolvePermanentOrLastKnown(g, source.AttachedTo.Val)
	if !ok || attached.permanent == nil || effectiveController(g, attached.permanent) != event.Player {
		return false
	}
	return triggerSelectionMatches(g, viewer, event, source.AttachedTo.Val, selection, source.ObjectID)
}

// attackedPlayerHasMostLife reports whether an attacker-declared event targets a
// player whose life total is at least every non-eliminated player's life total
// (dethrone CR 702.103). Attacks against planeswalkers or battles never match.
func attackedPlayerHasMostLife(g *game.Game, event game.Event) bool {
	if event.Kind != game.EventAttackerDeclared ||
		event.AttackTarget.PlaneswalkerID != 0 ||
		event.AttackTarget.BattleID != 0 {
		return false
	}
	attacked, ok := playerByID(g, event.AttackTarget.Player)
	if !ok || attacked.Eliminated {
		return false
	}
	for playerID := range game.PlayerID(game.NumPlayers) {
		other, ok := playerByID(g, playerID)
		if !ok || other.Eliminated {
			continue
		}
		if other.Life > attacked.Life {
			return false
		}
	}
	return true
}

func attackRecipientMatches(filter game.AttackRecipientKind, event game.Event) bool {
	if filter == game.AttackRecipientAny {
		return true
	}
	if event.Kind != game.EventAttackerDeclared {
		return false
	}
	switch {
	case event.AttackTarget.PlaneswalkerID != 0:
		return filter&game.AttackRecipientPlaneswalker != 0
	case event.AttackTarget.BattleID != 0:
		return filter&game.AttackRecipientBattle != 0
	default:
		return filter&game.AttackRecipientPlayer != 0
	}
}

func triggerSelectionMatches(g *game.Game, viewer game.PlayerID, event game.Event, objectID id.ID, selection *game.Selection, sourceObjectID id.ID) bool {
	if objectID == 0 {
		return false
	}
	subjectEvent := event
	if objectID != event.PermanentID {
		subjectEvent.PermanentID = objectID
		subjectEvent.CardID = 0
		subjectEvent.TokenName = ""
		subjectEvent.TokenDef = nil
		if objectID == event.SourceObjectID {
			subjectEvent.CardID = event.SourceID
		}
	}
	controller := event.Controller
	if resolved, ok := resolvePermanentOrLastKnown(g, objectID); ok && resolved.permanent != nil {
		controller = effectiveController(g, resolved.permanent)
	}
	subject := selectionSubject{
		kind:           subjectEventPermanent,
		g:              g,
		event:          subjectEvent,
		controller:     controller,
		viewer:         viewer,
		sourceObjectID: sourceObjectID,
	}
	return matchSelection(&subject, selection)
}

func eventStackObjectKindMatches(g *game.Game, event game.Event, kind game.StackObjectKind) bool {
	if event.StackObjectID == 0 {
		return false
	}
	obj, ok := stackObjectByID(g, event.StackObjectID)
	return ok && obj.Kind == kind
}

func eventPermanentIsToken(g *game.Game, event game.Event) bool {
	if event.PermanentID != 0 {
		if permanent, ok := permanentByObjectID(g, event.PermanentID); ok {
			return permanent.Token
		}
		if snapshot, ok := lastKnownObject(g, event.PermanentID); ok {
			return snapshot.TokenDef != nil || snapshot.CardID == 0
		}
	}
	return event.TokenDef != nil || (event.CardID == 0 && event.TokenName != "")
}

func triggerInterveningIf(g *game.Game, source *game.Permanent, controller game.PlayerID, trigger *game.TriggerCondition, event *game.Event) bool {
	if trigger == nil {
		return true
	}
	// Intervening "if" conditions are checked both as the event triggers and as
	// the ability resolves (CR 603.4).
	if trigger.InterveningIfControllerLifeAtLeast != 0 {
		player, ok := playerByID(g, controller)
		if !ok || player.Life < trigger.InterveningIfControllerLifeAtLeast {
			return false
		}
	}
	if trigger.InterveningIfEventPermanentHadCounters && !eventPermanentHadCounters(g, event) {
		return false
	}
	if trigger.InterveningIfEventPermanentHadNoCounterKind.Exists &&
		!eventPermanentHadNoCounterKind(g, event, trigger.InterveningIfEventPermanentHadNoCounterKind.Val) {
		return false
	}
	if trigger.InterveningIfEventPermanentWasKicked && (event == nil || !event.KickerPaid) {
		return false
	}
	if trigger.InterveningIfEventPermanentWasCast && (event == nil || !event.EnterWasCast) {
		return false
	}
	if trigger.InterveningIfEventPermanentWasCastByController &&
		(event == nil || !event.EnterWasCast || !event.EnterHasCastController ||
			event.EnterCastController != controller) {
		return false
	}
	if trigger.InterveningIfEventPermanentEnteredOrCastFromGraveyard &&
		!eventEnteredOrCastFromGraveyard(event) {
		return false
	}
	if trigger.InterveningIfEventPermanentEnteredOrCastFromControllerGraveyard &&
		!eventEnteredOrCastFromControllerGraveyard(event, controller) {
		return false
	}
	if !conditionSatisfied(g, conditionContext{
		controller: controller,
		source:     source,
		event:      event,
	}, trigger.InterveningCondition) {
		return false
	}
	return true
}

// eventEnteredOrCastFromGraveyard reports whether the entering permanent of an
// enter event came from any graveyard, either by entering the battlefield
// directly from a graveyard (reanimation) or by being cast from a graveyard
// (escape, flashback). It backs the any-graveyard "if they entered or were cast
// from a graveyard" intervening condition.
func eventEnteredOrCastFromGraveyard(event *game.Event) bool {
	if event == nil {
		return false
	}
	if event.FromZone == zone.Graveyard {
		return true
	}
	return event.EnterWasCast && event.EnterCastFromZone == zone.Graveyard
}

// eventEnteredOrCastFromControllerGraveyard is the controller-scoped form
// backing "if it entered from your graveyard or you cast it from your
// graveyard". A card always rests in its owner's graveyard (CR 404.2), so the
// source graveyard belongs to the trigger controller exactly when the entering
// card's owner is that controller. The cast branch additionally requires the
// controller to be the caster.
func eventEnteredOrCastFromControllerGraveyard(event *game.Event, controller game.PlayerID) bool {
	if event == nil || event.Player != controller {
		return false
	}
	if event.FromZone == zone.Graveyard {
		return true
	}
	return event.EnterWasCast && event.EnterCastFromZone == zone.Graveyard &&
		event.EnterHasCastController && event.EnterCastController == controller
}

func triggerControllerMatches(sourceController game.PlayerID, filter game.TriggerControllerFilter, eventController game.PlayerID) bool {
	switch filter {
	case game.TriggerControllerYou:
		return eventController == sourceController
	case game.TriggerControllerOpponent:
		return eventController != sourceController && eventController >= 0 && eventController < game.NumPlayers
	default:
		return true
	}
}

func triggerSourceMatches(g *game.Game, source *game.Permanent, filter game.TriggerSourceFilter, subject game.TriggerSubjectObject, event game.Event) bool {
	if filter == game.TriggerSourceAttachedPermanent {
		return triggerSourceAttachedPermanentMatchesSubject(g, source, event, subject)
	}
	if filter != game.TriggerSourceSelf {
		return true
	}
	if subject == game.TriggerSubjectDamageSource {
		return (source.ObjectID != 0 && event.SourceObjectID == source.ObjectID) ||
			(source.CardInstanceID != 0 && event.SourceID == source.CardInstanceID)
	}
	if subject == game.TriggerSubjectPermanent {
		return (source.ObjectID != 0 && event.PermanentID == source.ObjectID) ||
			(source.CardInstanceID != 0 && event.CardID == source.CardInstanceID)
	}
	subjectID := triggerSubjectObjectID(event, subject)
	return (source.ObjectID != 0 && event.SourceObjectID == source.ObjectID) ||
		(source.ObjectID != 0 && subjectID == source.ObjectID) ||
		(source.CardInstanceID != 0 && event.SourceID == source.CardInstanceID) ||
		(source.CardInstanceID != 0 && event.CardID == source.CardInstanceID)
}

func triggerSourceAttachedPermanentMatches(g *game.Game, source *game.Permanent, event game.Event) bool {
	return triggerSourceAttachedPermanentMatchesSubject(g, source, event, game.TriggerSubjectDefault)
}

func triggerSourceAttachedPermanentMatchesSubject(g *game.Game, source *game.Permanent, event game.Event, subject game.TriggerSubjectObject) bool {
	subjectID := triggerSubjectObjectID(event, subject)
	if source.ObjectID == 0 || subjectID == 0 {
		return false
	}
	if source.AttachedTo.Exists && source.AttachedTo.Val == subjectID {
		return true
	}
	if snapshot, ok := lastKnownObject(g, subjectID); ok {
		return slices.Contains(snapshot.Attachments, source.ObjectID)
	}
	return false
}

func triggerSubjectObjectID(event game.Event, subject game.TriggerSubjectObject) id.ID {
	switch subject {
	case game.TriggerSubjectBlockedAttacker:
		return event.BlockedAttackerID
	case game.TriggerSubjectDamageSource:
		return event.SourceObjectID
	default:
		return event.PermanentID
	}
}

func triggerSubjectPermanent(g *game.Game, subject game.TriggerSubjectObject, event game.Event) (*game.Permanent, bool) {
	objectID := triggerSubjectObjectID(event, subject)
	if objectID == 0 {
		return nil, false
	}
	if permanent, ok := permanentByObjectID(g, objectID); ok {
		return permanent, true
	}
	resolved, ok := resolvePermanentOrLastKnown(g, objectID)
	if !ok {
		return nil, false
	}
	return resolved.permanent, resolved.permanent != nil
}

func spellTargetsSource(g *game.Game, source *game.Permanent, event game.Event) bool {
	if event.Kind != game.EventSpellCast || source.ObjectID == 0 {
		return false
	}
	obj, ok := stackObjectByID(g, event.StackObjectID)
	if !ok {
		return false
	}
	for _, target := range obj.Targets {
		if target.Kind == game.TargetPermanent && target.PermanentID == source.ObjectID {
			return true
		}
	}
	return false
}

func spellTargetsPattern(g *game.Game, controller game.PlayerID, allow game.TargetAllow, predicate game.TargetPredicate, event game.Event) bool {
	if event.Kind != game.EventSpellCast {
		return false
	}
	obj, ok := stackObjectByID(g, event.StackObjectID)
	if !ok {
		return false
	}
	spec := game.TargetSpec{
		Allow:     allow,
		Predicate: predicate,
	}
	for _, target := range obj.Targets {
		if targetMatchesSpec(g, controller, 0, &spec, target) {
			return true
		}
	}
	return false
}

func triggerPlayerMatches(sourceController game.PlayerID, filter game.TriggerPlayerFilter, eventPlayer game.PlayerID) bool {
	switch filter {
	case game.TriggerPlayerYou:
		return eventPlayer == sourceController
	case game.TriggerPlayerOpponent:
		return eventPlayer != sourceController && eventPlayer >= 0 && eventPlayer < game.NumPlayers
	default:
		return true
	}
}

func eventPermanentHasType(g *game.Game, event game.Event, cardType types.Card) bool {
	if event.PermanentID != 0 {
		if permanent, ok := permanentByObjectID(g, event.PermanentID); ok {
			return permanentHasType(g, permanent, cardType)
		}
		// Leaves-the-battlefield and dies triggers look back at the permanent's
		// last existence on the battlefield (CR 603.10).
		if snapshot, ok := lastKnownObject(g, event.PermanentID); ok {
			return slices.Contains(snapshot.Types, cardType)
		}
	}
	if event.CardID != 0 {
		if card, ok := g.GetCardInstance(event.CardID); ok {
			return cardFaceOrDefault(card, game.FaceFront).HasType(cardType)
		}
	}
	if event.TokenDef != nil {
		return event.TokenDef.HasType(cardType)
	}
	return false
}

// triggerSubjectSelection returns the Selection a trigger pattern matches its
// event subject permanent against, preferring the explicit SubjectSelection and
// otherwise adapting the legacy permanent-type and non-token filters.
func triggerSubjectSelection(pattern *game.TriggerPattern) game.Selection {
	if !pattern.SubjectSelection.Empty() {
		return pattern.SubjectSelection
	}
	return game.Selection{
		RequiredTypes: pattern.RequirePermanentTypes,
		ExcludedTypes: pattern.ExcludePermanentTypes,
		NonToken:      pattern.RequireNonToken,
	}
}

// triggerCardSelection returns the Selection a trigger pattern matches a cast
// spell's card types against, preferring the explicit CardSelection and
// otherwise adapting the legacy card-type filters.
func triggerCardSelection(pattern *game.TriggerPattern) game.Selection {
	if !pattern.CardSelection.Empty() {
		return pattern.CardSelection
	}
	return game.Selection{
		RequiredTypes: pattern.RequireCardTypes,
		ExcludedTypes: pattern.ExcludeCardTypes,
	}
}

func eventSpellHistoric(event game.Event) bool {
	return slices.Contains(event.CardTypes, types.Artifact) ||
		slices.Contains(event.CardSupertypes, types.Legendary) ||
		slices.Contains(event.CardSubtypes, types.Saga)
}

// eventSpellCardTypes resolves the card types a spell-cast event matches against,
// using the event's recorded types and falling back to the front face.
func eventSpellCardTypes(g *game.Game, event game.Event) []types.Card {
	cardTypes := event.CardTypes
	if len(cardTypes) == 0 && event.CardID != 0 {
		if card, ok := g.GetCardInstance(event.CardID); ok {
			cardTypes = cardFaceOrDefault(card, game.FaceFront).Types
		}
	}
	return cardTypes
}

func eventPermanentHadCounters(g *game.Game, event *game.Event) bool {
	if event == nil || event.PermanentID == 0 {
		return false
	}
	if permanent, ok := permanentByObjectID(g, event.PermanentID); ok {
		return !permanent.Counters.IsEmpty()
	}
	if snapshot, ok := lastKnownObject(g, event.PermanentID); ok {
		return !snapshot.Counters.IsEmpty()
	}
	return false
}

func eventPermanentHadNoCounterKind(g *game.Game, event *game.Event, kind counter.Kind) bool {
	if event == nil || event.PermanentID == 0 {
		return false
	}
	if permanent, ok := permanentByObjectID(g, event.PermanentID); ok {
		return !permanent.Counters.Has(kind)
	}
	if snapshot, ok := lastKnownObject(g, event.PermanentID); ok {
		return !snapshot.Counters.Has(kind)
	}
	return false
}
