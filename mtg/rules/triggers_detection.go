package rules

import (
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

type pendingTriggeredAbility struct {
	controller          game.PlayerID
	sourceID            id.ID
	sourceCardID        id.ID
	sourceToken         *game.CardDef
	face                game.FaceIndex
	abilityIndex        int
	targets             []game.Target
	targetCounts        []int
	event               game.Event
	hasEvent            bool
	inline              *game.TriggeredAbility
	sagaChapter         bool
	wardTargetID        id.ID
	targetControllerLKI map[int]game.PlayerID
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
	}()
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
			TargetControllerLKI:     clonePlayerIDMap(trigger.targetControllerLKI),
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
			pending = append(pending, detectTriggeredAbilitiesFromPermanent(g, permanent, event)...)
		}
		if source, ok := leftBattlefieldTriggerSource(g, event); ok {
			pending = append(pending, detectTriggeredAbilitiesFromPermanent(g, source, event)...)
		}
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
	return filterPendingTriggeredAbilities(g, pending)
}

func captureEventTriggeredAbilities(g *game.Game, event game.Event) []game.EventTriggeredAbility {
	g.BeginStaticSourceFrame()
	defer g.EndStaticSourceFrame()
	var captured []game.EventTriggeredAbility
	for _, permanent := range g.Battlefield {
		triggers := detectTriggeredAbilitiesFromPermanent(g, permanent, event)
		for i := range triggers {
			trigger := &triggers[i]
			captured = append(captured, game.EventTriggeredAbility{
				Controller:     trigger.controller,
				SourceID:       trigger.sourceID,
				SourceCardID:   trigger.sourceCardID,
				SourceTokenDef: trigger.sourceToken,
				Face:           trigger.face,
				AbilityIndex:   trigger.abilityIndex,
				Ability:        trigger.inline,
			})
		}
	}
	return captured
}

func pendingTriggeredAbilitiesFromEvent(event game.Event) []pendingTriggeredAbility {
	pending := make([]pendingTriggeredAbility, 0, len(event.TriggeredAbilities))
	for _, captured := range event.TriggeredAbilities {
		pending = append(pending, pendingTriggeredAbility{
			controller:   captured.Controller,
			sourceID:     captured.SourceID,
			sourceCardID: captured.SourceCardID,
			sourceToken:  captured.SourceTokenDef,
			face:         captured.Face,
			abilityIndex: captured.AbilityIndex,
			inline:       captured.Ability,
			event:        event,
			hasEvent:     true,
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
				controller:   controller,
				sourceID:     permanent.ObjectID,
				sourceCardID: permanent.CardInstanceID,
				sourceToken:  permanent.TokenDef,
				face:         permanent.Face,
				abilityIndex: i,
				inline:       triggered,
				event:        event,
				hasEvent:     true,
			})
			continue
		}
		static, ok := body.(*game.StaticAbility)
		if !ok || !game.BodyHasKeyword(static, game.Ward) {
			continue
		}
		if ward, ok := wardTriggerForEvent(permanent, controller, static, event); ok {
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
