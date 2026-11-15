package rules

import (
	"fmt"
	"maps"
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/rules/payment"
)

// CR 605.1b: a nontargeted trigger that can add mana and triggers on an
// activated mana ability, or mana being added, is a mana ability. Merely
// becoming tapped (City of Brass) does not meet that criterion.
func triggeredManaAbility(body *game.TriggeredAbility, event game.Event) bool {
	if body == nil || len(game.BodyTargets(body)) != 0 {
		return false
	}
	pattern := body.Trigger.Pattern
	eligible := pattern.Event == game.EventPermanentTapped && pattern.RequireTappedForMana && event.TappedForMana ||
		pattern.Event == game.EventManaProduced ||
		pattern.Event == game.EventAbilityActivated && event.ManaAbility
	if !eligible {
		return false
	}
	for _, mode := range body.Content.Modes {
		for _, instruction := range mode.Sequence {
			if instruction.Primitive != nil && instruction.Primitive.Kind() == game.PrimitiveAddMana {
				return true
			}
		}
	}
	return false
}

func manaEventTriggers(g *game.Game, event game.Event) []pendingTriggeredAbility {
	if event.ManaTriggersResolved {
		return nil
	}
	var triggers []pendingTriggeredAbility
	if event.TriggeredAbilitiesCaptured {
		triggers = pendingTriggeredAbilitiesFromEvent(event)
	} else {
		triggers = eventTriggeredAbilities(g, event, []game.Event{event})
	}
	triggers = slices.DeleteFunc(triggers, func(trigger pendingTriggeredAbility) bool { return !triggeredManaAbility(trigger.inline, event) })
	triggers = multiplyAdditionalTriggers(g, triggers)
	for _, delayed := range g.DelayedTriggers {
		if !delayed.EventPattern.Exists || delayed.Window == game.DelayedWindowThisTurn && delayed.CreatedTurn != g.Turn.TurnNumber {
			continue
		}
		body := delayed.Ability
		body.Trigger.Pattern = delayed.EventPattern.Val
		if !triggeredManaAbility(&body, event) || len(matchEventDelayedTriggerEvents(g, &delayed, []game.Event{event})) == 0 {
			continue
		}
		source, _ := permanentByObjectID(g, delayed.SourceObjectID)
		if !triggerInterveningIf(g, source, delayed.Controller, &body.Trigger, &event) {
			continue
		}
		trigger := firedDelayedTrigger(&delayed, &body)
		trigger.event, trigger.hasEvent = event, true
		trigger.manaDelayedID, trigger.manaDelayedOneShot = delayed.ID, delayed.OneShot
		triggers = append(triggers, trigger)
	}
	return triggers
}

func manaTriggerObject(g *game.Game, trigger pendingTriggeredAbility) *game.StackObject {
	obj := &game.StackObject{Kind: game.StackTriggeredAbility, Controller: trigger.controller, SourceID: trigger.sourceID, SourceCardID: trigger.sourceCardID, SourceTokenDef: trigger.sourceToken, Face: trigger.face, AbilityIndex: trigger.abilityIndex, TriggerEvent: trigger.event, HasTriggerEvent: true, InlineTrigger: trigger.inline}
	if len(trigger.entryChoices) != 0 {
		seedResolutionEntryChoices(obj, trigger.entryChoices)
	} else if source, ok := permanentByObjectID(g, trigger.sourceID); ok {
		seedEntryChoices(obj, source)
	}
	return obj
}

func (e *Engine) resolveImmediateManaTriggers(g *game.Game, start int, agents [game.NumPlayers]PlayerAgent, log *TurnLog) {
	for index := start; index < len(g.Events); index++ {
		triggers := manaEventTriggers(g, g.Events[index])
		if len(triggers) == 0 {
			continue
		}
		g.Events[index].ManaTriggersResolved = true
		triggers = limitPendingTriggeredAbilities(g, triggers)
		for _, trigger := range triggers {
			if trigger.manaDelayedOneShot {
				g.DelayedTriggers = slices.DeleteFunc(g.DelayedTriggers, func(delayed game.DelayedTrigger) bool { return delayed.ID == trigger.manaDelayedID })
			}
			obj := manaTriggerObject(g, trigger)
			obj.ID = g.IDGen.Next()
			var before [game.NumPlayers]map[mana.Color]int
			for player := range before {
				before[player] = manaPoolColorSnapshot(g, game.PlayerID(player))
			}
			e.resolveTriggeredAbilityWithChoices(g, obj, agents, log)
			log.addTriggeredAbility(TriggeredAbilityLog{StackObjectID: obj.ID, Controller: obj.Controller, SourceID: obj.SourceID, SourceName: stackObjectSourceName(g, obj), AbilityText: trigger.inline.Text})
			for player := range before {
				produced := producedManaColorsSince(g, game.PlayerID(player), before[player])
				if len(produced) != 0 {
					source := manaProducedSource{sourceID: obj.SourceCardID, objectID: obj.SourceID, controller: obj.Controller, tokenDef: obj.SourceTokenDef}
					emitManaProducedEvent(g, source, game.PlayerID(player), produced, producedManaAmountSince(g, game.PlayerID(player), before[player]), false)
				}
			}
		}
	}
}

func forecastTriggeredMana(g *game.Game, events []game.Event, payer game.PlayerID) ([]payment.ManaProduction, map[id.ID]bool, bool) {
	var result []payment.ManaProduction
	spent := make(map[id.ID]bool)
	paths := make([]map[string]bool, len(events))
	for index := 0; index < len(events); index++ {
		view := g.ResourceView(game.ResourceProjection{SpentDelayed: spent})
		view.BeginStaticSourceFrame()
		defer view.EndStaticSourceFrame()
		triggers := manaEventTriggers(view, events[index])
		if len(triggers) == 0 {
			continue
		}
		event := events[index]
		signature := fmt.Sprint(event.Kind, event.SourceID, event.SourceObjectID, event.Player, event.ProducedManaColors, event.TappedForMana)
		if paths[index][signature] {
			return nil, nil, false
		}
		path := maps.Clone(paths[index])
		if path == nil {
			path = make(map[string]bool)
		}
		path[signature] = true
		for _, trigger := range triggers {
			if trigger.inline.Optional || trigger.inline.MaxTriggersPerTurn > 0 || !payment.ResourceManaCondition(trigger.inline.Trigger.InterveningCondition) {
				return nil, nil, false
			}
			if trigger.manaDelayedOneShot {
				spent[trigger.manaDelayedID] = true
			}
			obj := manaTriggerObject(g, trigger)
			// A triggered "any type that land produced" choice is deterministic
			// when its triggering tap produced exactly one type.
			for _, mode := range trigger.inline.Content.Modes {
				for _, instruction := range mode.Sequence {
					if choose, ok := instruction.Primitive.(game.Choose); ok {
						if choose.Choice.AtRandom {
							return nil, nil, false
						}
						_, choices := resolutionChoiceRequest(g, obj, &choose.Choice)
						if len(choices) != 1 {
							return nil, nil, false
						}
						for _, choice := range choices {
							rememberResolutionChoice(obj, string(choose.PublishChoice), choice)
						}
					}
				}
			}
			production, effects, ok := forecastManaContent(g, obj, trigger.inline.Content, payer)
			// Forecast only deterministic mana-only triggers. Other consequences
			// or selectable colors remain opaque rather than being approximated.
			if !ok || len(effects.Permanents) != 0 {
				return nil, nil, false
			}
			for player, after := range effects.Players {
				before, _ := playerByID(g, player)
				if after.Life != before.Life || after.Energy != before.EnergyCounters {
					return nil, nil, false
				}
			}
			result = append(result, production...)
			colors, amount := productionColors(production)
			if amount > 0 {
				event := game.Event{Kind: game.EventManaProduced, SourceID: obj.SourceCardID, SourceObjectID: obj.SourceID, PermanentID: obj.SourceID, Controller: obj.Controller, Player: payer, ProducedManaColors: colors, Amount: amount, ManaAbility: true}
				events = append(events, prepareEvent(g, event))
				paths = append(paths, path)
			}
		}
	}
	return result, spent, true
}

func productionColors(production []payment.ManaProduction) ([]mana.Color, int) {
	var colors []mana.Color
	total := 0
	for _, unit := range production {
		if unit.Amount <= 0 {
			continue
		}
		total += unit.Amount
		if !slices.Contains(colors, unit.Unit.Color) {
			colors = append(colors, unit.Unit.Color)
		}
	}
	return colors, total
}
