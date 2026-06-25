package rules

import (
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// pendingTriggeredAbility is a triggered ability that has triggered (CR 603.2)
// but has not yet been put on the stack (CR 603.3). It records the controlling
// player and source needed to build the StackObject, the triggering event, and
// the mode/target choices made during preparation (CR 603.3c-d).
type pendingTriggeredAbility struct {
	controller                  game.PlayerID
	sourceID                    id.ID
	sourceCardID                id.ID
	sourceToken                 *game.CardDef
	face                        game.FaceIndex
	abilityIndex                int
	chosenModes                 []int
	targets                     []game.Target
	targetCounts                []int
	event                       game.Event
	hasEvent                    bool
	inline                      *game.TriggeredAbility
	sagaChapter                 bool
	wardTargetID                id.ID
	capturedTargetControllerLKI map[int]game.PlayerID
	capturedTargetManaValueLKI  map[int]int
	additionalTriggers          int
	triggerMultiplierCaptured   bool
	// ordinaryTrigger marks a triggered ability eligible for chosen-creature-type
	// trigger multiplication: an ordinary event-driven triggered ability of a
	// permanent (including keyword triggers such as ward, prowess, and exalted).
	// Saga chapter, madness, state, delayed, and synthetic mana-spend-rider
	// triggers are never multiplied and leave this false.
	ordinaryTrigger bool
}

func (e *Engine) putTriggeredAbilitiesOnStack(g *game.Game) bool {
	return e.putTriggeredAbilitiesOnStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, nil)
}

// putTriggeredAbilitiesOnStackWithChoices implements the process by which
// abilities that have triggered are put on the stack (CR 603.3). It runs "the
// next time a player would receive priority": it gathers everything that has
// triggered since the last check (ordinary event triggers per CR 603.2, plus
// madness, state, delayed, and mana-spend-rider triggers), applies trigger
// replacement effects (suppression and CR 603.2d-style multiplication), orders
// the results in APNAP order (CR 603.3b), and pushes each onto the stack as an
// object that isn't a card (CR 603.3). It reports whether anything was placed so
// the caller can re-check state-based actions and triggers (CR 603.3b).
func (e *Engine) putTriggeredAbilitiesOnStackWithChoices(g *game.Game, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	start := g.TriggerEventCursor
	if start < 0 || start > len(g.Events) {
		start = len(g.Events)
	}
	events := append([]game.Event(nil), g.Events[start:]...)
	g.TriggerEventCursor = len(g.Events)
	var pending []pendingTriggeredAbility
	// Detecting triggered abilities is a pure read over every permanent; frame
	// the whole detection block so the static-ability source set is built once,
	// and close it before any triggers are ordered or placed on the stack.
	func() {
		g.BeginStaticSourceFrame()
		defer g.EndStaticSourceFrame()
		pending = e.detectTriggeredAbilities(g, events)
		pending = append(pending, e.detectMadnessTriggeredAbilities(g, events)...)
		pending = append(pending, e.detectStateTriggeredAbilities(g)...)
		pending = append(pending, e.drainFiredManaSpendRiders(g)...)
		pending = append(pending, drainReadyDelayedTriggers(g, events)...)
		pending = append(pending, drainReadyEventDelayedTriggers(g, events)...)
	}()
	if len(pending) == 0 {
		return false
	}
	pending = suppressOpponentEnteringTriggers(g, pending)
	if len(pending) == 0 {
		return false
	}
	pending = multiplyAdditionalTriggers(g, pending)
	pending = limitPendingTriggeredAbilities(g, pending)
	if len(pending) == 0 {
		return false
	}
	orderedTriggers := e.orderTriggeredAbilitiesAPNAP(g, pending, agents, log)
	placed := false
	for i := range orderedTriggers {
		trigger := &orderedTriggers[i]
		if !e.prepareTriggeredAbility(g, trigger, agents, log) {
			releasePendingStateTriggerLatch(g, trigger)
			continue
		}

		obj := &game.StackObject{
			ID:                          g.IDGen.Next(),
			Kind:                        game.StackTriggeredAbility,
			SourceID:                    trigger.sourceID,
			Face:                        trigger.face,
			SourceCardID:                trigger.sourceCardID,
			SourceTokenDef:              trigger.sourceToken,
			AbilityIndex:                trigger.abilityIndex,
			TriggerEvent:                trigger.event,
			HasTriggerEvent:             trigger.hasEvent,
			InlineTrigger:               trigger.inline,
			SagaChapter:                 trigger.sagaChapter,
			WardTargetStackObjectID:     trigger.wardTargetID,
			Controller:                  trigger.controller,
			ChosenModes:                 append([]int(nil), trigger.chosenModes...),
			Targets:                     append([]game.Target(nil), trigger.targets...),
			TargetCounts:                append([]int(nil), trigger.targetCounts...),
			CapturedTargetControllerLKI: clonePlayerIDMap(trigger.capturedTargetControllerLKI),
			CapturedTargetManaValueLKI:  cloneIntMap(trigger.capturedTargetManaValueLKI),
		}
		if source, ok := permanentByObjectID(g, trigger.sourceID); ok {
			seedEntryChoices(obj, source)
		}
		pushAbilityToStack(g, obj)
		placed = true
	}
	return placed
}

// suppressOpponentEnteringTriggers drops the pending entering-caused triggered
// abilities of permanents controlled by an opponent of an active opponent-
// entering-trigger suppressor's controller ("Permanents entering don't cause
// abilities of permanents your opponents control to trigger.", Elesh Norn,
// Mother of Machines). A trigger is suppressed when it is an ordinary triggered
// ability whose triggering event is a permanent entering the battlefield and the
// triggered ability's controller is an opponent of a suppressor's controller.
// The suppressor's own controller's entering triggers are unaffected.
func suppressOpponentEnteringTriggers(g *game.Game, pending []pendingTriggeredAbility) []pendingTriggeredAbility {
	suppressors := make([]game.PlayerID, 0)
	effects := activeRuleEffects(g)
	for i := range effects {
		if effects[i].Kind == game.RuleEffectSuppressOpponentEnteringTriggers {
			suppressors = append(suppressors, effects[i].Controller)
		}
	}
	if len(suppressors) == 0 {
		return pending
	}
	kept := pending[:0]
	for i := range pending {
		trigger := &pending[i]
		if triggerSuppressedByOpponentEntering(trigger, suppressors) {
			continue
		}
		kept = append(kept, *trigger)
	}
	return kept
}

// triggerSuppressedByOpponentEntering reports whether an entering-caused trigger
// belongs to an opponent of any suppressor's controller.
func triggerSuppressedByOpponentEntering(trigger *pendingTriggeredAbility, suppressors []game.PlayerID) bool {
	if !trigger.ordinaryTrigger || !trigger.hasEvent || !eventEntersBattlefield(&trigger.event) {
		return false
	}
	for _, controller := range suppressors {
		if controller != trigger.controller {
			return true
		}
	}
	return false
}

// multiplyAdditionalTriggers expands the pending triggered abilities by the extra
// occurrences granted by trigger-multiplying replacement effects: the
// chosen-creature-type doublers and the entering-permanent doublers
// (Panharmonicon, Yarok, Ancient Greenwarden). Each additional occurrence is an
// identical copy of the trigger, placed on the stack alongside the original.
func multiplyAdditionalTriggers(g *game.Game, pending []pendingTriggeredAbility) []pendingTriggeredAbility {
	originals := append([]pendingTriggeredAbility(nil), pending...)
	multiplied := make([]pendingTriggeredAbility, 0, len(originals))
	for i := range originals {
		trigger := originals[i]
		multiplied = append(multiplied, trigger)
		additional := trigger.additionalTriggers
		if !trigger.triggerMultiplierCaptured {
			additional = capturedChosenCreatureTypeAdditionalTriggerCount(g, &trigger)
		}
		additional += enteringPermanentAdditionalTriggerCount(g, &trigger)
		additional += controlledPermanentAdditionalTriggerCount(g, &trigger)
		for range additional {
			multiplied = append(multiplied, trigger)
		}
	}
	return multiplied
}

// enteringPermanentAdditionalTriggerCount counts the additional occurrences an
// ordinary triggered ability gains from active entering-permanent trigger
// doublers (Panharmonicon, Yarok, Ancient Greenwarden). It applies when the
// trigger's own event is a permanent entering the battlefield whose card type
// satisfies the doubler's filter and the doubler's controller controls the
// triggered ability's source. The count is read from the live rule effects; a
// doubler that leaves the battlefield before the triggers are placed does not
// contribute.
func enteringPermanentAdditionalTriggerCount(g *game.Game, trigger *pendingTriggeredAbility) int {
	if !trigger.ordinaryTrigger || !trigger.hasEvent || !eventEntersBattlefield(&trigger.event) {
		return 0
	}
	count := 0
	effects := activeRuleEffects(g)
	for i := range effects {
		effect := &effects[i]
		if effect.Kind != game.RuleEffectAdditionalTriggerForEnteringPermanent ||
			effect.SourceObjectID == 0 ||
			effect.Controller != trigger.controller {
			continue
		}
		if enteringPermanentMatchesFilter(g, &trigger.event, effect.PermanentTypes) {
			count++
		}
	}
	return count
}

// controlledPermanentAdditionalTriggerCount counts the additional occurrences an
// ordinary triggered ability gains from active controlled-permanent trigger
// doublers ("If a triggered ability of a legendary creature you control
// triggers, that ability triggers an additional time.", Annie Joins Up; Katara,
// the Fearless; Splinter, Radical Rat). It applies when the triggered ability's
// source is a permanent the doubler's controller controls and that permanent
// matches the doubler's source-permanent selection filter. Unlike the
// chosen-type and entering-permanent doublers this family includes the doubler's
// own triggers ("a ... you control", not "another"). The count is read from the
// live rule effects; the source permanent's type, supertype, and subtype are
// taken from its current state, falling back to last-known information once it
// has left the battlefield (so a dying creature's leaves-the-battlefield trigger
// still doubles).
func controlledPermanentAdditionalTriggerCount(g *game.Game, trigger *pendingTriggeredAbility) int {
	if !trigger.ordinaryTrigger || !trigger.hasEvent {
		return 0
	}
	count := 0
	effects := activeRuleEffects(g)
	for i := range effects {
		effect := &effects[i]
		if effect.Kind != game.RuleEffectAdditionalTriggerForControlledPermanent ||
			effect.SourceObjectID == 0 ||
			effect.Controller != trigger.controller {
			continue
		}
		if controlledTriggerSourceMatches(g, effect, trigger.sourceID) {
			count++
		}
	}
	return count
}

// controlledTriggerSourceMatches reports whether the triggered ability's source
// is a permanent controlled by the doubler's controller that satisfies the
// doubler's source-permanent selection filter. It checks the live permanent
// first and falls back to last-known information for a source that has left the
// battlefield.
func controlledTriggerSourceMatches(g *game.Game, effect *game.RuleEffect, sourceID id.ID) bool {
	if permanent, ok := permanentByObjectID(g, sourceID); ok {
		return effectiveController(g, permanent) == effect.Controller &&
			permanentMatchesTriggerSourceFilter(g, &effect.AffectedSelection, permanent)
	}
	snapshot, ok := lastKnownObject(g, sourceID)
	return ok &&
		snapshot.Controller == effect.Controller &&
		snapshotMatchesTriggerSourceFilter(&effect.AffectedSelection, &snapshot)
}

// permanentMatchesTriggerSourceFilter reports whether a live permanent satisfies
// the type, supertype, and subtype filter carried by a controlled-permanent
// trigger doubler's selection. RequiredTypes and Supertypes are conjunctive;
// SubtypesAny is disjunctive. An empty selection matches any permanent.
func permanentMatchesTriggerSourceFilter(g *game.Game, selection *game.Selection, permanent *game.Permanent) bool {
	for _, cardType := range selection.RequiredTypes {
		if !permanentHasType(g, permanent, cardType) {
			return false
		}
	}
	for _, supertype := range selection.Supertypes {
		if !permanentHasSupertype(g, permanent, supertype) {
			return false
		}
	}
	if len(selection.SubtypesAny) == 0 {
		return true
	}
	for _, subtype := range selection.SubtypesAny {
		if permanentHasSubtype(g, permanent, subtype) {
			return true
		}
	}
	return false
}

// snapshotMatchesTriggerSourceFilter mirrors permanentMatchesTriggerSourceFilter
// against last-known information for a source permanent that has left the
// battlefield.
func snapshotMatchesTriggerSourceFilter(selection *game.Selection, snapshot *game.ObjectSnapshot) bool {
	for _, cardType := range selection.RequiredTypes {
		if !slices.Contains(snapshot.Types, cardType) {
			return false
		}
	}
	for _, supertype := range selection.Supertypes {
		if !slices.Contains(snapshot.Supertypes, supertype) {
			return false
		}
	}
	if len(selection.SubtypesAny) == 0 {
		return true
	}
	for _, subtype := range selection.SubtypesAny {
		if slices.Contains(snapshot.Subtypes, subtype) {
			return true
		}
	}
	return false
}

func eventEntersBattlefield(event *game.Event) bool {
	return event.Kind == game.EventZoneChanged &&
		event.ToZone == zone.Battlefield &&
		event.PermanentID != 0
}

// enteringPermanentMatchesFilter reports whether the permanent that entered in
// event has a card type in filter. An empty filter matches any entering
// permanent ("a permanent" — Yarok). The entered permanent's current types are
// used while it remains on the battlefield, falling back to last-known
// information once it has left.
func enteringPermanentMatchesFilter(g *game.Game, event *game.Event, filter []types.Card) bool {
	if len(filter) == 0 {
		return true
	}
	if permanent, ok := permanentByObjectID(g, event.PermanentID); ok {
		for _, cardType := range filter {
			if permanentHasType(g, permanent, cardType) {
				return true
			}
		}
		return false
	}
	snapshot, ok := lastKnownObject(g, event.PermanentID)
	if !ok {
		return false
	}
	for _, cardType := range filter {
		if slices.Contains(snapshot.Types, cardType) {
			return true
		}
	}
	return false
}

// captureChosenTypeTriggerDoublers snapshots the active chosen-creature-type
// trigger doublers when an event is emitted, recording each doubler's source,
// controller, and chosen subtype. Ordinary triggered abilities that the event
// produces are multiplied from this snapshot at resolution, so the multiplier
// stays authoritative even if a doubler later changes controller or chosen type
// or leaves the battlefield before the triggers are placed on the stack.
func captureChosenTypeTriggerDoublers(g *game.Game) []game.ChosenTypeTriggerDoubler {
	effects := activeRuleEffects(g)
	var doublers []game.ChosenTypeTriggerDoubler
	for i := range effects {
		effect := &effects[i]
		if effect.Kind != game.RuleEffectAdditionalTriggerForChosenCreatureType || effect.SourceObjectID == 0 {
			continue
		}
		doubler, ok := permanentByObjectID(g, effect.SourceObjectID)
		if !ok {
			continue
		}
		choice, ok := doubler.EntryChoices[game.EntryTypeChoiceKey]
		if !ok || choice.Kind != game.ResolutionChoiceSubtype || choice.Subtype == "" {
			continue
		}
		doublers = append(doublers, game.ChosenTypeTriggerDoubler{
			SourceID:   effect.SourceObjectID,
			Controller: effect.Controller,
			Subtype:    choice.Subtype,
		})
	}
	return doublers
}

// liveChosenCreatureTypeAdditionalTriggerCount counts the additional occurrences
// an ordinary triggered ability gains from currently-active chosen-creature-type
// doublers. It runs at event emission (see captureEventTriggeredAbilities), when
// the live doubler set is the event-time state.
func liveChosenCreatureTypeAdditionalTriggerCount(g *game.Game, trigger *pendingTriggeredAbility) int {
	if !trigger.ordinaryTrigger {
		return 0
	}
	count := 0
	effects := activeRuleEffects(g)
	for i := range effects {
		if chosenCreatureTypeDoublerMatches(g, &effects[i], trigger) {
			count++
		}
	}
	return count + simultaneouslyDepartedChosenTypeDoublerCount(g, trigger, nil)
}

// capturedChosenCreatureTypeAdditionalTriggerCount counts the additional
// occurrences an ordinary triggered ability gains, using the doubler set,
// controller, and chosen subtype captured on the trigger's event when it was
// emitted. Capturing at emission keeps the multiplier authoritative even when a
// doubler's controller or chosen type changes, or the doubler leaves, before the
// triggers are ordered and placed on the stack. Only ordinary event-driven
// triggered abilities qualify; chapter, madness, state, delayed, and synthetic
// triggers are excluded.
func capturedChosenCreatureTypeAdditionalTriggerCount(g *game.Game, trigger *pendingTriggeredAbility) int {
	if !trigger.ordinaryTrigger || !trigger.hasEvent {
		return 0
	}
	seen := make(map[id.ID]bool)
	count := 0
	if snapshot := trigger.event.ChosenTypeTriggerDoublers; snapshot != nil {
		for i := range snapshot.Doublers {
			doubler := &snapshot.Doublers[i]
			if seen[doubler.SourceID] || !capturedChosenTypeDoublerMatches(g, doubler, trigger) {
				continue
			}
			seen[doubler.SourceID] = true
			count++
		}
	}
	return count + simultaneouslyDepartedChosenTypeDoublerCount(g, trigger, seen)
}

func capturedChosenTypeDoublerMatches(g *game.Game, doubler *game.ChosenTypeTriggerDoubler, trigger *pendingTriggeredAbility) bool {
	if doubler.SourceID == 0 ||
		doubler.SourceID == trigger.sourceID ||
		doubler.Controller != trigger.controller ||
		doubler.Subtype == "" {
		return false
	}
	return triggerSourceHadChosenCreatureType(g, trigger.sourceID, doubler.Subtype)
}

func chosenCreatureTypeDoublerMatches(g *game.Game, effect *game.RuleEffect, trigger *pendingTriggeredAbility) bool {
	if effect == nil ||
		effect.Kind != game.RuleEffectAdditionalTriggerForChosenCreatureType ||
		effect.SourceObjectID == 0 ||
		trigger.sourceID == effect.SourceObjectID ||
		trigger.controller != effect.Controller {
		return false
	}
	doubler, ok := permanentByObjectID(g, effect.SourceObjectID)
	if !ok {
		return false
	}
	choice, ok := doubler.EntryChoices[game.EntryTypeChoiceKey]
	if !ok || choice.Kind != game.ResolutionChoiceSubtype || choice.Subtype == "" {
		return false
	}
	return triggerSourceHadChosenCreatureType(g, trigger.sourceID, choice.Subtype)
}

func triggerSourceHadChosenCreatureType(g *game.Game, sourceID id.ID, subtype types.Sub) bool {
	source, ok := permanentByObjectID(g, sourceID)
	if ok {
		return permanentHasType(g, source, types.Creature) &&
			permanentHasSubtype(g, source, subtype)
	}
	snapshot, ok := lastKnownObject(g, sourceID)
	return ok &&
		slices.Contains(snapshot.Types, types.Creature) &&
		slices.Contains(snapshot.Subtypes, subtype)
}

// simultaneouslyDepartedChosenTypeDoublerCount counts doublers that left the
// battlefield in the same simultaneous batch as the trigger source, recovering
// their chosen type and doubler count from last-known information. seen records
// doubler object IDs already counted (for example by the event-time snapshot) so
// a doubler is never counted twice; it may be nil.
func simultaneouslyDepartedChosenTypeDoublerCount(g *game.Game, trigger *pendingTriggeredAbility, seen map[id.ID]bool) int {
	if !trigger.hasEvent || trigger.event.SimultaneousID == 0 {
		return 0
	}
	if seen == nil {
		seen = make(map[id.ID]bool)
	}
	count := 0
	for _, event := range g.Events {
		if event.SimultaneousID != trigger.event.SimultaneousID ||
			event.FromZone != zone.Battlefield ||
			event.PermanentID == 0 ||
			event.PermanentID == trigger.sourceID ||
			seen[event.PermanentID] {
			continue
		}
		seen[event.PermanentID] = true
		snapshot, ok := lastKnownObject(g, event.PermanentID)
		if !ok || snapshot.Controller != trigger.controller {
			continue
		}
		choice, ok := snapshot.EntryChoices[game.EntryTypeChoiceKey]
		if !ok || choice.Kind != game.ResolutionChoiceSubtype || choice.Subtype == "" {
			continue
		}
		if triggerSourceHadChosenCreatureType(g, trigger.sourceID, choice.Subtype) {
			count += snapshotChosenCreatureTypeTriggerMultiplierCount(&snapshot)
		}
	}
	return count
}

func snapshotChosenCreatureTypeTriggerMultiplierCount(snapshot *game.ObjectSnapshot) int {
	if snapshot == nil {
		return 0
	}
	count := 0
	for _, kind := range snapshot.RuleEffectKinds {
		if kind == game.RuleEffectAdditionalTriggerForChosenCreatureType {
			count++
		}
	}
	return count
}

// detectMadnessTriggeredAbilities produces the madness triggered ability for any
// card that was discarded to exile this batch (CR 702.35a: "When this card is
// exiled this way, its owner may cast it by paying [cost]..."). The madness
// static ability has already replaced the discard's destination with exile; this
// is the second, triggered half of the keyword (CR 603.2).
func (*Engine) detectMadnessTriggeredAbilities(g *game.Game, events []game.Event) []pendingTriggeredAbility {
	var pending []pendingTriggeredAbility
	for _, event := range events {
		if event.Kind != game.EventCardDiscarded || event.ToZone != zone.Exile || event.CardID == 0 {
			continue
		}
		card, ok := g.GetCardInstance(event.CardID)
		if !ok {
			continue
		}
		cost, ok := madnessCostForCard(cardFaceOrDefault(card, event.Face))
		if !ok {
			continue
		}
		pending = append(pending, pendingTriggeredAbility{
			controller:   event.Player,
			sourceID:     event.CardID,
			sourceCardID: event.CardID,
			face:         game.FaceFront,
			inline: &game.TriggeredAbility{
				Text:             "Madness",
				KeywordAbilities: []game.KeywordAbility{game.MadnessKeyword{Cost: cost}},
			},
			event:    event,
			hasEvent: true,
		})
	}
	return pending
}

// detectTriggeredAbilities scans for ordinary triggered abilities whose trigger
// event matches one of the given events (CR 603.2: whenever a game event matches
// a triggered ability's trigger event, that ability automatically triggers).
// Every active battlefield permanent is checked, plus sources that may have
// already left the battlefield: leaves-the-battlefield and simultaneous
// zone-change triggers look back in time (CR 603.10), while damage triggers fire
// at the moment damage is dealt even if the source or recipient has since moved
// (CR 120.4b).
func (*Engine) detectTriggeredAbilities(g *game.Game, events []game.Event) []pendingTriggeredAbility {
	// Detection is a pure read that scans every permanent for each event, so a
	// static-source frame avoids rescanning the battlefield for static-ability
	// sources on every permanent it inspects.
	g.BeginStaticSourceFrame()
	defer g.EndStaticSourceFrame()
	var pending []pendingTriggeredAbility
	for _, event := range events {
		if event.TriggeredAbilitiesCaptured {
			pending = append(pending, pendingTriggeredAbilitiesFromEvent(event)...)
			continue
		}
		for _, permanent := range g.Battlefield {
			if !activeBattlefieldPermanent(permanent) {
				continue
			}
			pending = append(pending, detectTriggeredAbilitiesFromPermanent(g, permanent, event)...)
		}
		if source, ok := leftBattlefieldTriggerSource(g, event); ok {
			pending = append(pending, detectTriggeredAbilitiesFromPermanent(g, source, event)...)
		}
		pending = append(pending, cycledCardSelfTriggers(g, event)...)
		for _, source := range simultaneousLeftBattlefieldTriggerSources(g, event, events) {
			pending = append(pending, detectTriggeredAbilitiesFromPermanent(g, source, event)...)
		}
		if source, ok := damageSourceTriggerSource(g, event); ok {
			pending = append(pending, detectTriggeredAbilitiesFromPermanent(g, source, event)...)
		}
		if source, ok := damageRecipientTriggerSource(g, event); ok {
			pending = append(pending, detectTriggeredAbilitiesFromPermanent(g, source, event)...)
		}
		for _, source := range damageAttachedTriggerSources(g, event) {
			pending = append(pending, detectTriggeredAbilitiesFromPermanent(g, source, event)...)
		}
	}
	return coalescePendingTriggeredAbilities(g, pending)
}

func captureEventTriggeredAbilities(g *game.Game, event game.Event) []game.EventTriggeredAbility {
	g.BeginStaticSourceFrame()
	defer g.EndStaticSourceFrame()
	var captured []game.EventTriggeredAbility
	for _, permanent := range g.Battlefield {
		if !activeBattlefieldPermanent(permanent) {
			continue
		}
		triggers := detectTriggeredAbilitiesFromPermanent(g, permanent, event)
		for i := range triggers {
			trigger := &triggers[i]
			captured = append(captured, game.EventTriggeredAbility{
				Controller:                trigger.controller,
				SourceID:                  trigger.sourceID,
				SourceCardID:              trigger.sourceCardID,
				SourceTokenDef:            trigger.sourceToken,
				Face:                      trigger.face,
				AbilityIndex:              trigger.abilityIndex,
				Ability:                   trigger.inline,
				AdditionalTriggers:        liveChosenCreatureTypeAdditionalTriggerCount(g, trigger),
				TriggerMultiplierCaptured: true,
			})
		}
	}
	return captured
}

func pendingTriggeredAbilitiesFromEvent(event game.Event) []pendingTriggeredAbility {
	pending := make([]pendingTriggeredAbility, 0, len(event.TriggeredAbilities))
	for _, captured := range event.TriggeredAbilities {
		pending = append(pending, pendingTriggeredAbility{
			controller:                captured.Controller,
			sourceID:                  captured.SourceID,
			sourceCardID:              captured.SourceCardID,
			sourceToken:               captured.SourceTokenDef,
			face:                      captured.Face,
			abilityIndex:              captured.AbilityIndex,
			inline:                    captured.Ability,
			event:                     event,
			hasEvent:                  true,
			additionalTriggers:        captured.AdditionalTriggers,
			triggerMultiplierCaptured: captured.TriggerMultiplierCaptured,
			ordinaryTrigger:           true,
		})
	}
	return pending
}

func simultaneousLeftBattlefieldTriggerSources(g *game.Game, event game.Event, events []game.Event) []*game.Permanent {
	if event.SimultaneousID == 0 {
		return nil
	}
	seen := make(map[id.ID]bool)
	var sources []*game.Permanent
	for _, candidate := range events {
		if candidate.SimultaneousID != event.SimultaneousID ||
			candidate.PermanentID == event.PermanentID ||
			seen[candidate.PermanentID] {
			continue
		}
		source, ok := leftBattlefieldTriggerSource(g, candidate)
		if !ok {
			continue
		}
		seen[candidate.PermanentID] = true
		sources = append(sources, source)
	}
	return sources
}

func coalescePendingTriggeredAbilities(g *game.Game, pending []pendingTriggeredAbility) []pendingTriggeredAbility {
	if len(pending) == 0 {
		return pending
	}
	filtered := make([]pendingTriggeredAbility, 0, len(pending))
	seenOneOrMore := make(map[triggerBatchKey]bool)
	for i := range pending {
		trigger := &pending[i]
		ability, ok := pendingTriggerAbility(g, trigger)
		if !ok {
			continue
		}
		if ability.Trigger.Pattern.OneOrMore && trigger.event.SimultaneousID != 0 {
			key := triggerBatchKey{
				sourceID:     trigger.sourceID,
				abilityIndex: trigger.abilityIndex,
				event:        trigger.event.Kind,
				controller:   trigger.controller,
				simultaneous: trigger.event.SimultaneousID,
			}
			if ability.Trigger.Pattern.OneOrMorePerAttackTarget {
				key.attackTarget = trigger.event.AttackTarget
			}
			if seenOneOrMore[key] {
				continue
			}
			seenOneOrMore[key] = true
		}
		filtered = append(filtered, *trigger)
	}
	return filtered
}

func limitPendingTriggeredAbilities(g *game.Game, pending []pendingTriggeredAbility) []pendingTriggeredAbility {
	filtered := make([]pendingTriggeredAbility, 0, len(pending))
	for i := range pending {
		trigger := &pending[i]
		ability, ok := pendingTriggerAbility(g, trigger)
		if !ok {
			continue
		}
		if ability.MaxTriggersPerTurn > 0 {
			key := game.TriggeredAbilityUse{SourceID: trigger.sourceID, AbilityIndex: trigger.abilityIndex}
			if g.TriggeredAbilitiesThisTurn == nil {
				g.TriggeredAbilitiesThisTurn = make(map[game.TriggeredAbilityUse]int)
			}
			if g.TriggeredAbilitiesThisTurn[key] >= ability.MaxTriggersPerTurn {
				continue
			}
			g.TriggeredAbilitiesThisTurn[key]++
		}
		filtered = append(filtered, *trigger)
	}
	return filtered
}

type triggerBatchKey struct {
	sourceID     id.ID
	abilityIndex int
	event        game.EventKind
	controller   game.PlayerID
	simultaneous id.ID
	attackTarget game.AttackTarget
}

func detectTriggeredAbilitiesFromPermanent(g *game.Game, permanent *game.Permanent, event game.Event) []pendingTriggeredAbility {
	var pending []pendingTriggeredAbility
	controller := effectiveController(g, permanent)
	for i, body := range permanentEffectiveAbilities(g, permanent) {
		if chapter, ok := body.(*game.ChapterAbility); ok {
			if event.Kind != game.EventCountersAdded ||
				event.PermanentID != permanent.ObjectID ||
				event.CounterKind != counter.Lore {
				continue
			}
			for _, number := range chapter.Chapters {
				if !sagaChapterTriggeredByEvent(permanent, event, number) {
					continue
				}
				pending = append(pending, pendingTriggeredAbility{
					controller:   controller,
					sourceID:     permanent.ObjectID,
					sourceCardID: permanent.CardInstanceID,
					sourceToken:  permanent.TokenDef,
					face:         permanent.Face,
					abilityIndex: i,
					inline: &game.TriggeredAbility{
						Text:    chapter.Text,
						Content: chapter.Content,
					},
					sagaChapter: true,
					event:       event,
					hasEvent:    true,
				})
				break
			}
			continue
		}
		if triggered, ok := body.(*game.TriggeredAbility); ok {
			trigger := &triggered.Trigger
			if !triggerMatchesEvent(g, permanent, &trigger.Pattern, event) || !triggerInterveningIf(g, permanent, controller, trigger, &event) {
				continue
			}
			pending = append(pending, pendingTriggeredAbility{
				controller:      controller,
				sourceID:        permanent.ObjectID,
				sourceCardID:    permanent.CardInstanceID,
				sourceToken:     permanent.TokenDef,
				face:            permanent.Face,
				abilityIndex:    i,
				inline:          triggered,
				event:           event,
				hasEvent:        true,
				ordinaryTrigger: true,
			})
			continue
		}
		static, ok := body.(*game.StaticAbility)
		if !ok || !game.BodyHasKeyword(static, game.Ward) {
			continue
		}
		if ward, ok := wardTriggerForEvent(permanent, controller, static, event); ok {
			pending = append(pending, pendingTriggeredAbility{
				controller:      controller,
				sourceID:        permanent.ObjectID,
				sourceCardID:    permanent.CardInstanceID,
				sourceToken:     permanent.TokenDef,
				face:            permanent.Face,
				inline:          ward,
				event:           event,
				hasEvent:        true,
				wardTargetID:    event.StackObjectID,
				ordinaryTrigger: true,
			})
		}
	}
	if prowess, ok := prowessTriggerForEvent(g, permanent, controller, event); ok {
		pending = append(pending, pendingTriggeredAbility{
			controller:      controller,
			sourceID:        permanent.ObjectID,
			sourceCardID:    permanent.CardInstanceID,
			sourceToken:     permanent.TokenDef,
			face:            permanent.Face,
			inline:          prowess,
			event:           event,
			hasEvent:        true,
			ordinaryTrigger: true,
		})
	}
	if exalted, ok := exaltedTriggerForEvent(g, permanent, controller, event); ok {
		pending = append(pending, pendingTriggeredAbility{
			controller:      controller,
			sourceID:        permanent.ObjectID,
			sourceCardID:    permanent.CardInstanceID,
			sourceToken:     permanent.TokenDef,
			face:            permanent.Face,
			inline:          exalted,
			event:           event,
			hasEvent:        true,
			ordinaryTrigger: true,
		})
	}
	if evolve, ok := evolveTriggerForEvent(g, permanent, controller, event); ok {
		pending = append(pending, pendingTriggeredAbility{
			controller:      controller,
			sourceID:        permanent.ObjectID,
			sourceCardID:    permanent.CardInstanceID,
			sourceToken:     permanent.TokenDef,
			face:            permanent.Face,
			inline:          evolve,
			event:           event,
			hasEvent:        true,
			ordinaryTrigger: true,
		})
	}
	return pending
}

func wardTriggerForEvent(permanent *game.Permanent, controller game.PlayerID, wardBody *game.StaticAbility, event game.Event) (*game.TriggeredAbility, bool) {
	if event.Kind != game.EventObjectBecameTarget || event.PermanentID != permanent.ObjectID || event.StackObjectID == 0 {
		return nil, false
	}
	wardCost, ok := game.BodyKeywordAbility(wardBody, game.Ward)
	if !ok || event.Controller == controller {
		return nil, false
	}
	ward, ok := wardCost.(game.WardKeyword)
	if !ok {
		return nil, false
	}
	return &game.TriggeredAbility{
		Text:             "Ward",
		KeywordAbilities: []game.KeywordAbility{game.WardKeyword{Cost: ward.Cost, AdditionalCosts: ward.AdditionalCosts}},
	}, true
}

func prowessTriggerForEvent(g *game.Game, permanent *game.Permanent, controller game.PlayerID, event game.Event) (*game.TriggeredAbility, bool) {
	if event.Kind != game.EventSpellCast || event.Controller != controller || !hasKeyword(g, permanent, game.Prowess) {
		return nil, false
	}
	if slices.Contains(event.CardTypes, types.Creature) {
		return nil, false
	}
	instr := game.Instruction{Primitive: game.ModifyPT{
		Object:         game.SourcePermanentReference(),
		PowerDelta:     game.Fixed(1),
		ToughnessDelta: game.Fixed(1),
		Duration:       game.DurationUntilEndOfTurn,
	}}
	return &game.TriggeredAbility{
		Text: "Prowess",
		Content: game.Mode{
			Sequence: []game.Instruction{instr},
		}.Ability(),
	}, true
}

func exaltedTriggerForEvent(g *game.Game, permanent *game.Permanent, controller game.PlayerID, event game.Event) (*game.TriggeredAbility, bool) {
	if event.Kind != game.EventAttackerDeclared || event.Controller != controller || !hasKeyword(g, permanent, game.Exalted) {
		return nil, false
	}
	if g.Combat == nil || len(g.Combat.Attackers) != 1 {
		return nil, false
	}
	instr := game.Instruction{Primitive: game.ModifyPT{
		Object:         game.EventPermanentReference(),
		PowerDelta:     game.Fixed(1),
		ToughnessDelta: game.Fixed(1),
		Duration:       game.DurationUntilEndOfTurn,
	}}
	return &game.TriggeredAbility{
		Text: "Exalted",
		Content: game.Mode{
			Sequence: []game.Instruction{instr},
		}.Ability(),
		KeywordAbilities: game.SimpleKeywords(game.Exalted),
	}, true
}

func evolveTriggerForEvent(g *game.Game, permanent *game.Permanent, controller game.PlayerID, event game.Event) (*game.TriggeredAbility, bool) {
	if event.Kind != game.EventPermanentEnteredBattlefield || event.Controller != controller || !hasKeyword(g, permanent, game.Evolve) {
		return nil, false
	}
	if event.PermanentID == permanent.ObjectID {
		return nil, false
	}
	entered, ok := permanentByObjectID(g, event.PermanentID)
	if !ok || !permanentHasType(g, entered, types.Creature) {
		return nil, false
	}
	if !evolveGreater(g, entered, permanent) {
		return nil, false
	}
	instr := game.Instruction{Primitive: game.AddCounter{
		Object:      game.SourcePermanentReference(),
		CounterKind: counter.PlusOnePlusOne,
		Amount:      game.Fixed(1),
	}}
	return &game.TriggeredAbility{
		Text: "Evolve",
		Content: game.Mode{
			Sequence: []game.Instruction{instr},
		}.Ability(),
		KeywordAbilities: game.SimpleKeywords(game.Evolve),
	}, true
}

// evolveGreater reports whether the entered creature has greater power or
// greater toughness than the evolve creature (CR 702.100b). Toughness counts
// only when both creatures have a defined toughness, so an undefined "*"
// toughness never satisfies the comparison on that axis.
func evolveGreater(g *game.Game, entered, evolve *game.Permanent) bool {
	if effectivePower(g, entered) > effectivePower(g, evolve) {
		return true
	}
	enteredToughness, enteredOK := effectiveToughness(g, entered)
	evolveToughness, evolveOK := effectiveToughness(g, evolve)
	return enteredOK && evolveOK && enteredToughness > evolveToughness
}

func (*Engine) detectStateTriggeredAbilities(g *game.Game) []pendingTriggeredAbility {
	if g.StateTriggerLatches == nil {
		g.StateTriggerLatches = make(map[game.StateTriggerKey]bool)
	}
	var pending []pendingTriggeredAbility
	seen := make(map[game.StateTriggerKey]bool)
	for _, permanent := range g.Battlefield {
		if !activeBattlefieldPermanent(permanent) {
			continue
		}
		controller := effectiveController(g, permanent)
		for i, body := range permanentEffectiveAbilities(g, permanent) {
			triggeredBody, ok := body.(*game.TriggeredAbility)
			if !ok {
				continue
			}
			if !triggeredBody.Trigger.State.Exists {
				continue
			}
			key := game.StateTriggerKey{
				SourceObjectID: permanent.ObjectID,
				SourceCardID:   permanent.CardInstanceID,
				AbilityIndex:   i,
			}
			seen[key] = true
			if !stateTriggerConditionSatisfied(g, controller, &triggeredBody.Trigger.State.Val) {
				continue
			}
			if g.StateTriggerLatches[key] {
				continue
			}
			// State triggers do not trigger again until the ability leaves the stack
			// or fails to enter it (CR 603.8).
			g.StateTriggerLatches[key] = true
			pending = append(pending, pendingTriggeredAbility{
				controller:   controller,
				sourceID:     permanent.ObjectID,
				sourceCardID: permanent.CardInstanceID,
				sourceToken:  permanent.TokenDef,
				face:         permanent.Face,
				abilityIndex: i,
				inline:       triggeredBody,
			})
		}
	}
	for key := range g.StateTriggerLatches {
		if !seen[key] {
			if source, ok := permanentByObjectID(g, key.SourceObjectID); ok && source.PhasedOut {
				continue
			}
			delete(g.StateTriggerLatches, key)
		}
	}
	return pending
}

func stateTriggerConditionSatisfied(g *game.Game, controller game.PlayerID, condition *game.StateTriggerCondition) bool {
	if condition == nil {
		return false
	}
	if condition.MatchControllerLifeLessOrEqual {
		player, ok := playerByID(g, controller)
		if !ok || player.Life > condition.ControllerLifeLessOrEqual {
			return false
		}
	}
	return true
}

// cycledCardSelfTriggers detects a cycled card's own "When you cycle this card"
// triggered abilities (CR 702.29e). These abilities function from the graveyard
// the card is put into as it is cycled, so the ordinary battlefield scan in
// detectTriggeredAbilities never sees them. Only the cycled card's self-source
// cycle triggers are considered; its other abilities do not function from the
// graveyard.
func cycledCardSelfTriggers(g *game.Game, event game.Event) []pendingTriggeredAbility {
	if event.Kind != game.EventCycled || event.CardID == 0 {
		return nil
	}
	card, ok := g.GetCardInstance(event.CardID)
	if !ok {
		return nil
	}
	def, ok := cardFaceDef(card, game.FaceFront)
	if !ok {
		return nil
	}
	source := &game.Permanent{
		ObjectID:       event.SourceID,
		CardInstanceID: card.ID,
		Owner:          card.Owner,
		Controller:     event.Controller,
		Face:           game.FaceFront,
	}
	var pending []pendingTriggeredAbility
	for i := range def.TriggeredAbilities {
		triggered := &def.TriggeredAbilities[i]
		pattern := &triggered.Trigger.Pattern
		if pattern.Event != game.EventCycled || pattern.Source != game.TriggerSourceSelf {
			continue
		}
		if !triggerMatchesEventForController(g, source, event.Controller, pattern, event) ||
			!triggerInterveningIf(g, source, event.Controller, &triggered.Trigger, &event) {
			continue
		}
		pending = append(pending, pendingTriggeredAbility{
			controller:      event.Controller,
			sourceID:        event.SourceID,
			sourceCardID:    card.ID,
			face:            game.FaceFront,
			abilityIndex:    i,
			inline:          triggered,
			event:           event,
			hasEvent:        true,
			ordinaryTrigger: true,
		})
	}
	return pending
}

func leftBattlefieldTriggerSource(g *game.Game, event game.Event) (*game.Permanent, bool) {
	if event.FromZone != zone.Battlefield || event.PermanentID == 0 {
		return nil, false
	}
	if _, ok := permanentByObjectID(g, event.PermanentID); ok {
		return nil, false
	}
	if event.CardID != 0 {
		snapshot, ok := lastKnownObject(g, event.PermanentID)
		if !ok {
			return nil, false
		}
		return &game.Permanent{
			ObjectID:       event.PermanentID,
			CardInstanceID: event.CardID,
			Owner:          event.Player,
			Controller:     event.Controller,
			Face:           event.Face,
			FaceDown:       snapshot.FaceDown,
			FaceDownFace:   snapshot.FaceDownFace,
			FaceDownKind:   snapshot.FaceDownKind,
			MergedCards:    append([]game.MergedCard(nil), snapshot.MergedCards...),
		}, true
	}
	if event.TokenDef == nil {
		return nil, false
	}
	snapshot, ok := lastKnownObject(g, event.PermanentID)
	if !ok {
		return nil, false
	}
	return &game.Permanent{
		ObjectID:     event.PermanentID,
		Owner:        event.Player,
		Controller:   event.Controller,
		Face:         event.Face,
		FaceDown:     snapshot.FaceDown,
		FaceDownFace: snapshot.FaceDownFace,
		FaceDownKind: snapshot.FaceDownKind,
		MergedCards:  append([]game.MergedCard(nil), snapshot.MergedCards...),
		Token:        true,
		TokenDef:     event.TokenDef,
	}, true
}

func damageSourceTriggerSource(g *game.Game, event game.Event) (*game.Permanent, bool) {
	if event.Kind != game.EventDamageDealt || event.SourceObjectID == 0 {
		return nil, false
	}
	if permanent, ok := permanentByObjectID(g, event.SourceObjectID); ok && activeBattlefieldPermanent(permanent) {
		return nil, false
	}
	snapshot, ok := lastKnownObject(g, event.SourceObjectID)
	if !ok {
		return nil, false
	}
	sourceCardID := event.SourceID
	if sourceCardID == 0 {
		sourceCardID = snapshot.CardID
	}
	return &game.Permanent{
		ObjectID:       event.SourceObjectID,
		CardInstanceID: sourceCardID,
		Owner:          snapshot.Owner,
		Controller:     event.Controller,
		Face:           snapshot.Face,
		FaceDown:       snapshot.FaceDown,
		FaceDownFace:   snapshot.FaceDownFace,
		FaceDownKind:   snapshot.FaceDownKind,
		MergedCards:    append([]game.MergedCard(nil), snapshot.MergedCards...),
		Token:          snapshot.TokenDef != nil || snapshot.CardID == 0,
		TokenDef:       snapshot.TokenDef,
	}, true
}

func damageRecipientTriggerSource(g *game.Game, event game.Event) (*game.Permanent, bool) {
	if event.Kind != game.EventDamageDealt || event.PermanentID == 0 || event.PermanentID == event.SourceObjectID {
		return nil, false
	}
	if permanent, ok := permanentByObjectID(g, event.PermanentID); ok && activeBattlefieldPermanent(permanent) {
		return nil, false
	}
	snapshot, ok := lastKnownObject(g, event.PermanentID)
	if !ok {
		return nil, false
	}
	sourceCardID := event.CardID
	if sourceCardID == 0 {
		sourceCardID = snapshot.CardID
	}
	return &game.Permanent{
		ObjectID:       event.PermanentID,
		CardInstanceID: sourceCardID,
		Owner:          snapshot.Owner,
		Controller:     snapshot.Controller,
		Face:           snapshot.Face,
		FaceDown:       snapshot.FaceDown,
		FaceDownFace:   snapshot.FaceDownFace,
		FaceDownKind:   snapshot.FaceDownKind,
		MergedCards:    append([]game.MergedCard(nil), snapshot.MergedCards...),
		Token:          snapshot.TokenDef != nil || snapshot.CardID == 0,
		TokenDef:       snapshot.TokenDef,
	}, true
}

func damageAttachedTriggerSources(g *game.Game, event game.Event) []*game.Permanent {
	if event.Kind != game.EventDamageDealt {
		return nil
	}
	var sources []*game.Permanent
	seen := make(map[id.ID]bool)
	for _, subjectID := range []id.ID{event.SourceObjectID, event.PermanentID} {
		subject, ok := lastKnownObject(g, subjectID)
		if !ok {
			continue
		}
		for _, attachmentID := range subject.Attachments {
			if attachmentID == 0 || seen[attachmentID] {
				continue
			}
			seen[attachmentID] = true
			if permanent, ok := permanentByObjectID(g, attachmentID); ok && activeBattlefieldPermanent(permanent) {
				continue
			}
			snapshot, ok := lastKnownObject(g, attachmentID)
			if !ok {
				continue
			}
			sources = append(sources, &game.Permanent{
				ObjectID:       snapshot.ObjectID,
				CardInstanceID: snapshot.CardID,
				Owner:          snapshot.Owner,
				Controller:     snapshot.Controller,
				Face:           snapshot.Face,
				FaceDown:       snapshot.FaceDown,
				FaceDownFace:   snapshot.FaceDownFace,
				FaceDownKind:   snapshot.FaceDownKind,
				MergedCards:    append([]game.MergedCard(nil), snapshot.MergedCards...),
				Token:          snapshot.TokenDef != nil || snapshot.CardID == 0,
				TokenDef:       snapshot.TokenDef,
			})
		}
	}
	return sources
}
