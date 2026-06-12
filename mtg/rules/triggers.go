package rules

import (
	"fmt"
	"slices"

	"github.com/natefinch/council4/mtg/game/zone"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
)

type pendingTriggeredAbility struct {
	controller   game.PlayerID
	sourceID     id.ID
	sourceCardID id.ID
	sourceToken  *game.CardDef
	face         game.FaceIndex
	abilityIndex int
	targets      []game.Target
	targetCounts []int
	event        game.Event
	hasEvent     bool
	inline       *game.TriggeredAbility
	sagaChapter  bool
	wardTargetID id.ID
}

func (e *Engine) putTriggeredAbilitiesOnStack(g *game.Game) bool {
	return e.putTriggeredAbilitiesOnStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, nil)
}

func (e *Engine) putTriggeredAbilitiesOnStackWithChoices(g *game.Game, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	start := g.TriggerEventCursor
	if start < 0 || start > len(g.Events) {
		start = len(g.Events)
	}
	events := append([]game.Event(nil), g.Events[start:]...)
	g.TriggerEventCursor = len(g.Events)
	pending := e.detectTriggeredAbilities(g, events)
	pending = append(pending, e.detectMadnessTriggeredAbilities(g, events)...)
	pending = append(pending, e.detectStateTriggeredAbilities(g)...)
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
			ID:                      g.IDGen.Next(),
			Kind:                    game.StackTriggeredAbility,
			SourceID:                trigger.sourceID,
			Face:                    trigger.face,
			SourceCardID:            trigger.sourceCardID,
			SourceTokenDef:          trigger.sourceToken,
			AbilityIndex:            trigger.abilityIndex,
			TriggerEvent:            trigger.event,
			HasTriggerEvent:         trigger.hasEvent,
			InlineTrigger:           trigger.inline,
			SagaChapter:             trigger.sagaChapter,
			WardTargetStackObjectID: trigger.wardTargetID,
			Controller:              trigger.controller,
			Targets:                 append([]game.Target(nil), trigger.targets...),
			TargetCounts:            append([]int(nil), trigger.targetCounts...),
		}
		pushAbilityToStack(g, obj)
		placed = true
	}
	return placed
}

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

func (e *Engine) detectTriggeredAbilities(g *game.Game, events []game.Event) []pendingTriggeredAbility {
	var pending []pendingTriggeredAbility
	for _, event := range events {
		for _, permanent := range g.Battlefield {
			pending = append(pending, e.detectTriggeredAbilitiesFromPermanent(g, permanent, event)...)
		}
		if source, ok := leftBattlefieldTriggerSource(g, event); ok {
			pending = append(pending, e.detectTriggeredAbilitiesFromPermanent(g, source, event)...)
		}
		for _, source := range simultaneousLeftBattlefieldTriggerSources(g, event, events) {
			pending = append(pending, e.detectTriggeredAbilitiesFromPermanent(g, source, event)...)
		}
		if source, ok := damageSourceTriggerSource(g, event); ok {
			pending = append(pending, e.detectTriggeredAbilitiesFromPermanent(g, source, event)...)
		}
		if source, ok := damageRecipientTriggerSource(g, event); ok {
			pending = append(pending, e.detectTriggeredAbilitiesFromPermanent(g, source, event)...)
		}
		for _, source := range damageAttachedTriggerSources(g, event) {
			pending = append(pending, e.detectTriggeredAbilitiesFromPermanent(g, source, event)...)
		}
	}
	return filterPendingTriggeredAbilities(g, pending)
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

func filterPendingTriggeredAbilities(g *game.Game, pending []pendingTriggeredAbility) []pendingTriggeredAbility {
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

func (*Engine) detectTriggeredAbilitiesFromPermanent(g *game.Game, permanent *game.Permanent, event game.Event) []pendingTriggeredAbility {
	var pending []pendingTriggeredAbility
	controller := effectiveController(g, permanent)
	for i, body := range permanentEffectiveAbilities(g, permanent) {
		if chapter, ok := body.(game.ChapterAbility); ok {
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
		if triggered, ok := body.(game.TriggeredAbility); ok {
			trigger := &triggered.Trigger
			if !triggerMatchesEvent(g, permanent, &trigger.Pattern, event) || !triggerInterveningIf(g, permanent, controller, trigger, &event) {
				continue
			}
			pending = append(pending, pendingTriggeredAbility{
				controller:   controller,
				sourceID:     permanent.ObjectID,
				sourceCardID: permanent.CardInstanceID,
				sourceToken:  permanent.TokenDef,
				face:         permanent.Face,
				abilityIndex: i,
				inline:       &triggered,
				event:        event,
				hasEvent:     true,
			})
			continue
		}
		static, ok := body.(game.StaticAbility)
		if !ok || !game.BodyHasKeyword(static, game.Ward) {
			continue
		}
		if ward, ok := wardTriggerForEvent(permanent, controller, &static, event); ok {
			pending = append(pending, pendingTriggeredAbility{
				controller:   controller,
				sourceID:     permanent.ObjectID,
				sourceCardID: permanent.CardInstanceID,
				sourceToken:  permanent.TokenDef,
				face:         permanent.Face,
				inline:       ward,
				event:        event,
				hasEvent:     true,
				wardTargetID: event.StackObjectID,
			})
		}
	}
	if prowess, ok := prowessTriggerForEvent(g, permanent, controller, event); ok {
		pending = append(pending, pendingTriggeredAbility{
			controller:   controller,
			sourceID:     permanent.ObjectID,
			sourceCardID: permanent.CardInstanceID,
			sourceToken:  permanent.TokenDef,
			face:         permanent.Face,
			inline:       prowess,
			event:        event,
			hasEvent:     true,
		})
	}
	if exalted, ok := exaltedTriggerForEvent(g, permanent, controller, event); ok {
		pending = append(pending, pendingTriggeredAbility{
			controller:   controller,
			sourceID:     permanent.ObjectID,
			sourceCardID: permanent.CardInstanceID,
			sourceToken:  permanent.TokenDef,
			face:         permanent.Face,
			inline:       exalted,
			event:        event,
			hasEvent:     true,
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
		KeywordAbilities: []game.KeywordAbility{game.WardKeyword{Cost: ward.Cost}},
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

func (*Engine) detectStateTriggeredAbilities(g *game.Game) []pendingTriggeredAbility {
	if g.StateTriggerLatches == nil {
		g.StateTriggerLatches = make(map[game.StateTriggerKey]bool)
	}
	var pending []pendingTriggeredAbility
	seen := make(map[game.StateTriggerKey]bool)
	for _, permanent := range g.Battlefield {
		controller := effectiveController(g, permanent)
		for i, body := range permanentEffectiveAbilities(g, permanent) {
			triggeredBody, ok := body.(game.TriggeredAbility)
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
				inline:       &triggeredBody,
			})
		}
	}
	for key := range g.StateTriggerLatches {
		if !seen[key] {
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
	if _, ok := permanentByObjectID(g, event.SourceObjectID); ok {
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
	if _, ok := permanentByObjectID(g, event.PermanentID); ok {
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
			if _, ok := permanentByObjectID(g, attachmentID); ok {
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

func (e *Engine) triggerTargets(g *game.Game, controller game.PlayerID, source *game.CardDef, sourceObjectID id.ID, ability *game.TriggeredAbility, agents [game.NumPlayers]PlayerAgent, log *TurnLog) ([]game.Target, bool) {
	result := targetChoicesForBodyFromSourceObject(g, controller, source, sourceObjectID, *ability)
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
	if pattern.Event == game.EventUnknown || pattern.Event != event.Kind {
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

	// Trigger patterns are checked when the triggering event is processed, and
	// LTB/dies checks may need last-known information for the moved permanent
	// (CR 603.2, CR 603.6c, CR 603.10).
	sourceController := effectiveController(g, source)
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
	if pattern.PlayerEventOrdinalThisTurn > 0 &&
		pattern.PlayerEventOrdinalThisTurn != event.PlayerEventOrdinalThisTurn {
		return false
	}
	if !triggerCombatPatternMatches(g, sourceController, source, pattern, event) {
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
		if !triggerSelectionMatches(g, sourceController, event, event.PermanentID, &subjectSel) {
			return false
		}
	}
	if cardSel := triggerCardSelection(pattern); !cardSel.Empty() {
		subject := selectionSubject{
			kind:      subjectCastSpell,
			g:         g,
			event:     event,
			cardTypes: eventSpellCardTypes(g, event),
		}
		if !matchSelection(&subject, &cardSel) {
			return false
		}
	}
	if pattern.MatchStackObjectKind && !eventStackObjectKindMatches(g, event, pattern.StackObjectKind) {
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
	if !pattern.DamageRecipientSelection.Empty() &&
		event.DamageRecipient == game.DamageRecipientPermanent &&
		!triggerSelectionMatches(g, viewer, event, event.PermanentID, &pattern.DamageRecipientSelection) {
		return false
	}
	if !pattern.DamageSourceSelection.Empty() &&
		!triggerSelectionMatches(g, viewer, event, event.SourceObjectID, &pattern.DamageSourceSelection) {
		return false
	}
	if !pattern.RelatedSubjectSelection.Empty() &&
		!triggerSelectionMatches(g, viewer, event, event.RelatedPermanentID, &pattern.RelatedSubjectSelection) {
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
	return recipientID == 0 || triggerSelectionMatches(g, viewer, event, recipientID, selection)
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
	return triggerSelectionMatches(g, viewer, event, source.AttachedTo.Val, selection)
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

func triggerSelectionMatches(g *game.Game, viewer game.PlayerID, event game.Event, objectID id.ID, selection *game.Selection) bool {
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
		kind:       subjectEventPermanent,
		g:          g,
		event:      subjectEvent,
		controller: controller,
		viewer:     viewer,
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
	if !conditionSatisfied(g, conditionContext{
		controller: controller,
		source:     source,
		event:      event,
	}, trigger.InterveningCondition) {
		return false
	}
	return true
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

func (e *Engine) orderTriggeredAbilitiesAPNAP(g *game.Game, triggers []pendingTriggeredAbility, agents [game.NumPlayers]PlayerAgent, log *TurnLog) []pendingTriggeredAbility {
	if len(triggers) == 0 {
		return triggers
	}
	ordered := make([]pendingTriggeredAbility, 0, len(triggers))
	used := make([]bool, len(triggers))
	for _, playerID := range triggerAPNAPPlayers(g) {
		var playerTriggers []pendingTriggeredAbility
		for i := range triggers {
			trigger := &triggers[i]
			if trigger.controller == playerID {
				playerTriggers = append(playerTriggers, *trigger)
				used[i] = true
			}
		}
		ordered = append(ordered, e.chooseTriggerOrder(g, playerID, playerTriggers, agents, log)...)
	}
	for i := range triggers {
		if !used[i] {
			ordered = append(ordered, triggers[i])
		}
	}
	return ordered
}

func (e *Engine) prepareTriggeredAbility(g *game.Game, trigger *pendingTriggeredAbility, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	source, _ := pendingTriggerSourceDef(g, trigger)
	ability, ok := pendingTriggerAbilityFromDef(source, trigger)
	if !ok {
		return false
	}
	targets, ok := e.triggerTargets(g, trigger.controller, source, trigger.sourceID, ability, agents, log)
	if !ok {
		return false
	}
	trigger.targets = targets
	targetCounts, ok := bodyTargetCounts(g, trigger.controller, source, trigger.sourceID, ability, targets)
	if !ok {
		panic("validated triggered ability targets could not be segmented")
	}
	trigger.targetCounts = targetCounts
	return true
}

func releasePendingStateTriggerLatch(g *game.Game, trigger *pendingTriggeredAbility) {
	if trigger.inline == nil || !trigger.inline.Trigger.State.Exists {
		return
	}
	deleteStateTriggerLatch(g, trigger.sourceID, trigger.sourceCardID, trigger.abilityIndex)
}

func pendingTriggerAbility(g *game.Game, trigger *pendingTriggeredAbility) (*game.TriggeredAbility, bool) {
	source, _ := pendingTriggerSourceDef(g, trigger)
	return pendingTriggerAbilityFromDef(source, trigger)
}

func pendingTriggerSourceDef(g *game.Game, trigger *pendingTriggeredAbility) (*game.CardDef, bool) {
	if trigger.sourceCardID != 0 {
		if card, ok := g.GetCardInstance(trigger.sourceCardID); ok {
			return card.Def.FaceDef(trigger.face)
		}
		return nil, false
	}
	if trigger.sourceToken == nil {
		return nil, false
	}
	return trigger.sourceToken.FaceDef(trigger.face)
}

func pendingTriggerAbilityFromDef(def *game.CardDef, trigger *pendingTriggeredAbility) (*game.TriggeredAbility, bool) {
	if trigger.inline != nil {
		return trigger.inline, true
	}
	if def == nil {
		return nil, false
	}
	body := def.BodyAt(trigger.abilityIndex)
	triggered, ok := body.(game.TriggeredAbility)
	if !ok {
		return nil, false
	}
	return &triggered, true
}

func (e *Engine) chooseTriggerOrder(g *game.Game, playerID game.PlayerID, triggers []pendingTriggeredAbility, agents [game.NumPlayers]PlayerAgent, log *TurnLog) []pendingTriggeredAbility {
	if len(triggers) <= 1 {
		return triggers
	}
	options := make([]game.ChoiceOption, 0, len(triggers))
	for i := range triggers {
		trigger := &triggers[i]
		options = append(options, game.ChoiceOption{
			Index: i,
			Label: fmt.Sprintf("source=%v ability=%d", trigger.sourceID, trigger.abilityIndex),
		})
	}
	selected := e.chooseChoice(g, agents, orderChoiceRequest(playerID, "Order triggered abilities.", options), log)
	ordered := make([]pendingTriggeredAbility, 0, len(triggers))
	used := make([]bool, len(triggers))
	for _, index := range selected {
		if index < 0 || index >= len(triggers) || used[index] {
			continue
		}
		ordered = append(ordered, triggers[index])
		used[index] = true
	}
	for i := range triggers {
		if !used[i] {
			ordered = append(ordered, triggers[i])
		}
	}
	return ordered
}

func triggerAPNAPPlayers(g *game.Game) []game.PlayerID {
	players := make([]game.PlayerID, 0, game.NumPlayers)
	playerID := g.Turn.ActivePlayer
	for range int(game.NumPlayers) {
		if playerID < 0 || playerID >= game.NumPlayers {
			break
		}
		players = append(players, playerID)
		playerID = g.TurnOrder.NextPriority(playerID)
		if slices.Contains(players, playerID) {
			break
		}
	}
	return players
}
