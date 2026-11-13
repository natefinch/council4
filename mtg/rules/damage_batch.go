package rules

import (
	"cmp"
	"fmt"
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

type pendingDamage struct {
	event      damageEvent
	applied    map[id.ID]bool
	shields    map[int]bool
	redirected bool
}

type damagePreventionConsequence struct {
	permanent  *game.Permanent
	remove     opt.V[counter.Kind]
	add        int
	controller game.PlayerID
	riders     preventedDamageRiders
}

type damageBatch struct {
	game         *game.Game
	members      []pendingDamage
	usedCounters map[id.ID]bool
	after        []damagePreventionConsequence
	// spentOneShots holds one-shot shields that prevented damage in this batch
	// outside a wider damage event scope; they stay applicable to the batch's
	// other recipients and are used up when the batch's modifications end.
	spentOneShots []int
	// simultaneousID groups the batch's damage events; a redirection that adds
	// a damage event to the batch allocates one when the batch had none.
	simultaneousID id.ID
}

type damageModifierKind uint8

const (
	damageProtection damageModifierKind = iota
	damageShield
	damageShieldCounter
	damageRedirection
	damageReplacement
)

type damageModifier struct {
	kind        damageModifierKind
	member      int
	shield      int
	replacement game.ReplacementEffect
	label       string
}

func prepareDamageBatch(g *game.Game, events []damageEvent, simultaneousID id.ID) *damageBatch {
	batch := &damageBatch{
		game:           g,
		usedCounters:   make(map[id.ID]bool),
		simultaneousID: simultaneousID,
	}
	for _, event := range events {
		event.amount = max(0, event.amount)
		event.simultaneousID = simultaneousID
		batch.members = append(batch.members, pendingDamage{
			event: event, applied: make(map[id.ID]bool), shields: make(map[int]bool),
		})
	}
	for {
		modifiers := batch.modifiers()
		if len(modifiers) == 0 {
			break
		}
		chooser := replacementDecisionPlayer(g, batch.members[modifiers[0].member].event)
		// CR 616.1: simultaneous affected players make their choices in APNAP.
		for _, player := range triggerAPNAPPlayers(g) {
			index := slices.IndexFunc(modifiers, func(modifier damageModifier) bool {
				return replacementDecisionPlayer(g, batch.members[modifier.member].event) == player
			})
			if index >= 0 {
				chooser = player
				break
			}
		}
		var choices []damageModifier
		var labels []string
		for _, modifier := range modifiers {
			if replacementDecisionPlayer(g, batch.members[modifier.member].event) == chooser {
				choices = append(choices, modifier)
				label := modifier.label
				if len(batch.members) > 1 {
					event := batch.members[modifier.member].event
					if event.permanent != nil {
						label += fmt.Sprintf(" to %s [%d]", permanentEffectiveName(g, event.permanent), event.permanent.ObjectID)
					} else {
						label += fmt.Sprintf(" to player %d", event.player)
					}
					if modifier.kind == damageReplacement || modifier.kind == damageProtection {
						label += fmt.Sprintf(" from %s [%d]", damageEventSourceName(g, &game.Event{
							SourceID: event.sourceID, SourceObjectID: event.sourceObjectID,
						}), event.sourceObjectID)
					}
				}
				labels = append(labels, label)
			}
		}
		chosen := 0
		if len(choices) > 1 {
			chosen = chooseReplacementDecision(g, chooser, labels).Selected[0]
		}
		batch.apply(choices[chosen])
	}
	for _, index := range batch.spentOneShots {
		if riders, ok := takePreventedDamageRiders(&g.PreventionShields[index]); ok {
			batch.after = append(batch.after, damagePreventionConsequence{riders: riders})
		}
		exhaustOneShotShield(&g.PreventionShields[index])
	}
	g.PreventionShields = compactPreventionShields(g.PreventionShields)
	return batch
}

func (batch *damageBatch) modifiers() []damageModifier {
	var result []damageModifier
	for index := range batch.members {
		member := &batch.members[index]
		event := member.event
		if event.amount <= 0 {
			continue
		}
		preventable := !damageEventCantBePrevented(batch.game, event)
		if preventable && damageEventProtected(batch.game, event) {
			result = append(result, damageModifier{kind: damageProtection, member: index, label: "protection"})
		}
		// Unpreventable damage can still be redirected (CR 615.12, 614.9).
		for _, shield := range applicablePreventionShieldIndices(batch.game, event) {
			if member.shields[shield] {
				continue
			}
			if batch.game.PreventionShields[shield].Redirect != game.DamageRedirectNone {
				result = append(result, damageModifier{kind: damageShield, member: index, shield: shield, label: "damage redirection"})
			} else if preventable {
				result = append(result, damageModifier{kind: damageShield, member: index, shield: shield, label: "prevention shield"})
			}
		}
		if event.permanent != nil && event.permanent.Counters.Get(counter.Shield) > 0 &&
			!batch.usedCounters[event.permanent.ObjectID] {
			result = append(result, damageModifier{kind: damageShieldCounter, member: index, label: "shield counter"})
		}
		if event.permanent == nil {
			if _, ok := playerDamageRedirectPermanent(batch.game, event.player); ok {
				result = append(result, damageModifier{kind: damageRedirection, member: index, label: "damage redirection"})
			}
		}
		// Unpreventable damage still invokes prevention's additional effects
		// (CR 615.12), but never consumes a finite prevention shield.
		for _, replacement := range matchingDamageReplacementEffects(batch.game, event, member.applied, true) {
			result = append(result, damageModifier{
				kind: damageReplacement, member: index, replacement: replacement,
				label: replacementEffectLabels([]game.ReplacementEffect{replacement})[0],
			})
		}
	}
	var unique []damageModifier
	for _, candidate := range result {
		duplicate := slices.ContainsFunc(unique, func(other damageModifier) bool {
			if candidate.kind != other.kind {
				return false
			}
			switch candidate.kind {
			case damageShieldCounter, damageRedirection:
				return sameDamageRecipient(batch.members[candidate.member].event, batch.members[other.member].event)
			case damageShield:
				return candidate.shield == other.shield &&
					sameDamageRecipient(batch.members[candidate.member].event, batch.members[other.member].event)
			case damageReplacement:
				return candidate.replacement.DamagePreventAll && candidate.replacement.ID == other.replacement.ID &&
					sameDamageRecipient(batch.members[candidate.member].event, batch.members[other.member].event)
			default:
				return candidate.member == other.member
			}
		})
		if !duplicate {
			unique = append(unique, candidate)
		}
	}
	return unique
}

func (batch *damageBatch) prevent(index, amount int) int {
	event := &batch.members[index].event
	if damageEventCantBePrevented(batch.game, *event) {
		return 0
	}
	amount = min(event.amount, amount)
	if amount > 0 {
		emitDamagePreventedEvent(batch.game, *event, amount)
		event.amount -= amount
	}
	return amount
}

func sameDamageRecipient(first, second damageEvent) bool {
	if first.permanent != nil || second.permanent != nil {
		return first.permanent != nil && second.permanent != nil &&
			first.permanent.ObjectID == second.permanent.ObjectID
	}
	return first.player == second.player
}

func (batch *damageBatch) apply(modifier damageModifier) {
	member := &batch.members[modifier.member]
	switch modifier.kind {
	case damageProtection:
		batch.prevent(modifier.member, member.event.amount)
	case damageShieldCounter:
		permanent := member.event.permanent
		batch.usedCounters[permanent.ObjectID] = true
		for index := range batch.members {
			if sameDamageRecipient(member.event, batch.members[index].event) {
				batch.prevent(index, batch.members[index].event.amount)
			}
		}
		batch.after = append(batch.after, damagePreventionConsequence{permanent: permanent, remove: opt.Val(counter.Shield)})
	case damageShield:
		batch.applyShield(modifier)
	case damageRedirection:
		permanent, ok := playerDamageRedirectPermanent(batch.game, member.event.player)
		if !ok {
			panic("applicable damage redirection disappeared")
		}
		player := member.event.player
		for index := range batch.members {
			other := &batch.members[index]
			if other.event.amount > 0 && other.event.permanent == nil && other.event.player == player {
				other.event.permanent = permanent
				other.event.player = 0
				other.redirected = true
			}
		}
	case damageReplacement:
		replacement := modifier.replacement
		if replacement.DamagePreventAll {
			prevented := 0
			recipient := member.event
			for index := range batch.members {
				other := &batch.members[index]
				if !sameDamageRecipient(recipient, other.event) {
					continue
				}
				matches := matchingDamageReplacementEffects(batch.game, other.event, other.applied, true)
				if !slices.ContainsFunc(matches, func(effect game.ReplacementEffect) bool { return effect.ID == replacement.ID }) {
					continue
				}
				other.applied[replacement.ID] = true
				prevented += batch.prevent(index, other.event.amount)
			}
			if recipient.permanent != nil {
				if replacement.DamagePreventedBecomesPlusOneCounters && prevented > 0 {
					batch.after = append(batch.after, damagePreventionConsequence{
						permanent: recipient.permanent, add: prevented, controller: replacement.Controller,
					})
				}
				if replacement.DamagePreventedRemovesPlusOneCounter {
					batch.after = append(batch.after, damagePreventionConsequence{
						permanent: recipient.permanent, remove: opt.Val(counter.PlusOnePlusOne),
					})
				}
			}
			return
		}
		member.applied[replacement.ID] = true
		if replacement.DamagePreventAmount > 0 {
			batch.prevent(modifier.member, replacement.DamagePreventAmount)
		} else {
			if replacement.DamageMultiplier > 1 {
				member.event.amount *= replacement.DamageMultiplier
			}
			member.event.amount = max(0, member.event.amount+replacement.DamageAddend)
		}
	default:
		panic("unknown damage modifier")
	}
}

func (batch *damageBatch) applyShield(modifier damageModifier) {
	shield := &batch.game.PreventionShields[modifier.shield]
	if shield.Redirect != game.DamageRedirectNone {
		batch.applyRedirectShield(modifier)
		return
	}
	var indices []int
	total := 0
	recipient := batch.members[modifier.member].event
	for index, member := range batch.members {
		if member.event.amount > 0 && sameDamageRecipient(recipient, member.event) && preventionShieldApplies(batch.game, *shield, member.event) &&
			!damageEventCantBePrevented(batch.game, member.event) &&
			preventionShieldSourceMatches(batch.game, shield, member.event) {
			indices = append(indices, index)
			total += member.event.amount
		}
	}
	prevented := make([]int, len(indices))
	switch {
	case shield.HalfRoundedDown:
		remaining := total / 2
		for position, index := range indices {
			prevented[position] = min(remaining, batch.members[index].event.amount)
			remaining -= prevented[position]
			batch.members[index].shields[modifier.shield] = true
		}
	case shield.All:
		for position, index := range indices {
			prevented[position] = batch.members[index].event.amount
			batch.members[index].shields[modifier.shield] = true
		}
	default:
		remaining := min(total, shield.Amount)
		shield.Amount -= remaining
		// Each choice allocates one point, so partial shares are expressible
		// without offering more prevention than any source would deal (615.7).
		for remaining > 0 {
			var options []game.ChoiceOption
			for position, index := range indices {
				event := batch.members[index].event
				if prevented[position] < event.amount {
					options = append(options, damageSourceOption(batch.game, event, position))
				}
			}
			position := batch.choosePreventionSource(replacementDecisionPlayer(batch.game, recipient), options,
				"Choose which source's damage to prevent next.")
			amount := 1
			if _, choicesAvailable := replacementChoiceContextFor(batch.game); !choicesAvailable || len(options) == 1 {
				amount = min(remaining, batch.members[indices[position]].event.amount-prevented[position])
			}
			prevented[position] += amount
			remaining -= amount
		}
	}
	total = 0
	for position, index := range indices {
		event := batch.members[index].event
		amount := batch.prevent(index, prevented[position])
		total += amount
		shield.Prevented += amount
		// A rider's condition on the prevented source is checked as the damage
		// is prevented; its effect waits until the damage event ends (CR 615.5).
		if amount > 0 && len(shield.PreventedRiders) > 0 && (len(shield.RiderSourceColors) == 0 ||
			preventionSourceCharacteristicsMatch(damageSourceColors(batch.game, event), nil, shield.RiderSourceColors, nil)) {
			shield.RiderAmount += amount
			shield.RiderSourceObjectID = event.sourceObjectID
			shield.RiderPlayer = event.controller
		}
	}
	// A one-shot shield is used up by the damage event it prevents, after all
	// of the chosen source's simultaneous in-scope damage has been prevented;
	// one that prevents nothing, including half that rounds down to nothing,
	// remains (CR 609.7b).
	if shield.OneShot && total > 0 {
		if scope := recipient.preventionScope; scope != 0 {
			shield.SpentScope = scope
		} else if !slices.Contains(batch.spentOneShots, modifier.shield) {
			batch.spentOneShots = append(batch.spentOneShots, modifier.shield)
		}
	}
}

// applyRedirectShield applies a one-shot redirection shield to all of its chosen
// source's damage to the modifier's recipient in this batch (CR 614.9). Each
// redirected damage event keeps its source, so it is still combat damage when
// it was, and remains subject to replacement and prevention effects that apply
// to its new recipient (CR 616.2), but never to this shield again (CR 614.5).
func (batch *damageBatch) applyRedirectShield(modifier damageModifier) {
	g := batch.game
	shield := &g.PreventionShields[modifier.shield]
	recipient := batch.members[modifier.member].event
	var indices []int
	for index, member := range batch.members {
		if member.event.amount > 0 && !member.shields[modifier.shield] && sameDamageRecipient(recipient, member.event) &&
			preventionShieldApplies(g, *shield, member.event) && preventionShieldSourceMatches(g, shield, member.event) {
			indices = append(indices, index)
		}
	}
	for _, index := range indices {
		member := &batch.members[index]
		member.shields[modifier.shield] = true
		switch shield.Redirect {
		case game.DamageRedirectToController:
			member.event.permanent, member.event.player = nil, shield.RedirectPlayer
		case game.DamageRedirectToDamageSourceController:
			member.event.permanent, member.event.player = nil, member.event.controller
		case game.DamageRedirectToObject:
			permanent, ok := permanentByObjectID(g, shield.RedirectPermanentID)
			if !ok {
				panic("applicable damage redirection destination disappeared")
			}
			member.event.permanent, member.event.player = permanent, 0
		case game.DamageRedirectAlsoToDamageSourceController:
			// The original damage is still dealt; the shield's own source deals
			// that much damage to the chosen source's controller in the same
			// replaced event.
			batch.addSimultaneousDamage(damageEvent{
				sourceID: shield.RedirectSourceID, sourceObjectID: shield.RedirectSourceObjectID,
				controller: shield.Controller, player: member.event.controller, amount: member.event.amount,
			})
			continue
		default:
			panic("unknown damage redirect")
		}
		member.redirected = true
	}
	if len(indices) == 0 {
		return
	}
	if scope := recipient.preventionScope; scope != 0 {
		shield.SpentScope = scope
	} else if !slices.Contains(batch.spentOneShots, modifier.shield) {
		batch.spentOneShots = append(batch.spentOneShots, modifier.shield)
	}
}

// addSimultaneousDamage adds a damage event a replacement creates to the batch,
// dealt simultaneously with the event it replaced.
func (batch *damageBatch) addSimultaneousDamage(event damageEvent) {
	if event.amount <= 0 || !isPlayerAlive(batch.game, event.player) {
		return
	}
	if batch.simultaneousID == 0 {
		batch.simultaneousID = batch.game.IDGen.Next()
		for index := range batch.members {
			batch.members[index].event.simultaneousID = batch.simultaneousID
		}
	}
	event.simultaneousID = batch.simultaneousID
	batch.members = append(batch.members, pendingDamage{
		event: event, applied: make(map[id.ID]bool), shields: make(map[int]bool),
	})
}

func damageSourceOption(g *game.Game, event damageEvent, index int) game.ChoiceOption {
	return game.ChoiceOption{
		Index: index, Label: fmt.Sprintf("Damage from %s [%d]", damageEventSourceName(g, &game.Event{
			SourceID: event.sourceID, SourceObjectID: event.sourceObjectID,
		}), event.sourceObjectID), Card: damageSourceChoiceInfo(g, event),
	}
}

func (batch *damageBatch) choosePreventionSource(chooser game.PlayerID, options []game.ChoiceOption, prompt string) int {
	selected := options[0].Index
	if ctx, ok := replacementChoiceContextFor(batch.game); ok && len(options) > 1 {
		return ctx.engine.chooseChoice(batch.game, ctx.agents, game.ChoiceRequest{
			Kind: game.ChoicePreventionAllocation, Player: chooser, Prompt: prompt, Options: options,
			MinChoices: 1, MaxChoices: 1, DefaultSelection: []int{selected},
		}, ctx.log)[0]
	}
	return selected
}

func damageSourceChoiceInfo(g *game.Game, event damageEvent) opt.V[game.ChoiceCardInfo] {
	if source, ok := permanentByObjectID(g, event.sourceObjectID); ok {
		values := effectivePermanentValues(g, source)
		manaValue, _ := effectivePermanentManaValue(g, source)
		return opt.Val(game.ChoiceCardInfo{
			CardID: event.sourceID, Name: permanentEffectiveName(g, source), ManaValue: manaValue,
			Types: slices.Clone(values.types), Colors: slices.Clone(values.colors),
		})
	}
	return cardChoiceInfo(g, event.sourceID)
}

func (batch *damageBatch) finish() {
	// CR 615.5: prevention's additional effects happen immediately after the
	// original damage event, not while simultaneous assignments are processed.
	for _, consequence := range batch.after {
		if consequence.permanent != nil {
			if consequence.remove.Exists {
				consequence.permanent.Counters.Remove(consequence.remove.Val, 1)
			}
			if consequence.add > 0 {
				addCountersToPermanentControlledBy(batch.game, consequence.controller, consequence.permanent, counter.PlusOnePlusOne, consequence.add)
			}
		} else if consequence.riders.amount > 0 {
			applyPreventedDamageRiders(batch.game, consequence.riders)
		}
	}
}

type preparedDamageResult struct {
	damage      damageModificationResult
	event       game.Event
	deathtouch  bool
	counterKind opt.V[counter.Kind]
	counters    int
	lifeLoss    int
	lifelink    bool
	commanders  []id.ID
	toxic       int
	poison      int
	initiative  bool
	ring        bool
}

func prepareDamageResult(g *game.Game, member pendingDamage, combat bool) preparedDamageResult {
	event := member.event
	result := preparedDamageResult{damage: damageModificationResult{event: event, amount: event.amount, redirected: member.redirected}}
	if event.amount <= 0 {
		return result
	}
	result.event = damageReplacementEvent(g, event)
	result.event.SimultaneousID = event.simultaneousID
	switch {
	case event.permanent != nil:
		result.event.Player = effectiveController(g, event.permanent)
		result.deathtouch = damageSourceHasDeathtouch(g, event.sourceObjectID)
		switch {
		case damageSourceUsesMinusOneCounters(g, event.sourceObjectID) && permanentHasType(g, event.permanent, types.Creature):
			result.counterKind = opt.Val(counter.MinusOneMinusOne)
			result.counters = replacementPermanentCounterPlacementAmount(g, event.controller, event.permanent, result.counterKind.Val, event.amount)
		case permanentHasType(g, event.permanent, types.Planeswalker):
			result.counterKind = opt.Val(counter.Loyalty)
			result.counters = event.amount
		case permanentHasType(g, event.permanent, types.Battle):
			result.counterKind = opt.Val(counter.Defense)
			result.counters = event.amount
		default:
		}
	case damageSourceHasKeyword(g, event.sourceID, event.sourceObjectID, game.Infect):
		result.poison = replacementPlayerCounterPlacementAmount(g, event.controller, event.player, counter.Poison, event.amount)
	case !playerRuleEffectActive(g, event.player, game.RuleEffectDamageDoesntCauseLifeLoss):
		result.lifeLoss = prepareLifeLoss(g, event.player, event.amount)
	default:
	}
	result.event = prepareEvent(g, result.event)
	if combat && event.combatDamage {
		source, ok := permanentByObjectID(g, event.sourceObjectID)
		if !ok {
			panic("combat damage source disappeared before its simultaneous event")
		}
		result.lifelink = hasKeyword(g, source, game.Lifelink)
		if event.permanent == nil {
			result.commanders = sourceCommanderIDs(g, source)
			for _, body := range permanentEffectiveAbilities(g, source) {
				if amount, ok := game.BodyToxicAmount(body); ok {
					result.toxic += amount
				}
			}
			if result.toxic > 0 {
				result.toxic = replacementPlayerCounterPlacementAmount(g, event.controller, event.player, counter.Poison, result.toxic)
			}
			result.initiative = g.Players[event.player].HasInitiative && event.player != event.controller
			result.ring = g.Players[event.controller].RingLevel >= ringMaxLevel && isRingBearer(g, source)
		}
	}
	return result
}

// resolveDamageBatch deals events as one simultaneous damage event and reports
// each event's result in order. Damage events a replacement added to the batch
// are dealt with them but not reported.
func resolveDamageBatch(g *game.Game, events []damageEvent, simultaneousID id.ID, combat bool, log *TurnLog) []damageModificationResult {
	batch := prepareDamageBatch(g, events, simultaneousID)
	simultaneousID = batch.simultaneousID
	results := make([]preparedDamageResult, len(batch.members))
	type lifelinkGain struct {
		source     id.ID
		controller game.PlayerID
		amount     int
	}
	var gains []lifelinkGain
	order := make([]int, len(batch.members))
	for index := range order {
		order[index] = index
	}
	rank := make(map[game.PlayerID]int)
	for index, player := range triggerAPNAPPlayers(g) {
		rank[player] = index
	}
	slices.SortStableFunc(order, func(first, second int) int {
		return cmp.Compare(rank[replacementDecisionPlayer(g, batch.members[first].event)],
			rank[replacementDecisionPlayer(g, batch.members[second].event)])
	})
	for _, index := range order {
		member := batch.members[index]
		result := prepareDamageResult(g, member, combat)
		results[index] = result
		if result.lifelink && member.event.amount > 0 {
			position := slices.IndexFunc(gains, func(gain lifelinkGain) bool { return gain.source == member.event.sourceObjectID })
			if position < 0 {
				gains = append(gains, lifelinkGain{source: member.event.sourceObjectID, controller: member.event.controller})
				position = len(gains) - 1
			}
			gains[position].amount += member.event.amount
		}
	}
	for index := range gains {
		gains[index].amount = prepareLifeGain(g, gains[index].controller, gains[index].amount)
	}
	// Commit every result before observing their consequences or publishing
	// facts. Damage triggers were captured before its results (CR 120.4b-c).
	var actual []damageModificationResult
	type damageFact struct {
		event    game.Event
		prepared bool
	}
	var facts []damageFact
	for _, result := range results {
		actual = append(actual, result.damage)
		event := result.damage.event
		if event.amount <= 0 {
			continue
		}
		if event.permanent != nil {
			switch {
			case result.counterKind.Exists && result.counterKind.Val == counter.MinusOneMinusOne:
				if result.counters > 0 {
					previous := event.permanent.Counters.Get(counter.MinusOneMinusOne)
					event.permanent.Counters.Add(counter.MinusOneMinusOne, result.counters)
					facts = append(facts, damageFact{event: counterAddedEvent(g, event.permanent, event.controller, counter.MinusOneMinusOne, previous, result.counters, simultaneousID)})
				}
			case result.counterKind.Exists:
				kind := result.counterKind.Val
				previous := event.permanent.Counters.Get(kind)
				if removed := event.permanent.Counters.Remove(kind, result.counters); removed > 0 {
					facts = append(facts, damageFact{event: counterRemovedEvent(event.permanent, event.controller, kind, previous, removed, simultaneousID)})
				}
			default:
				event.permanent.MarkedDamage += event.amount
			}
			if result.deathtouch {
				event.permanent.MarkedDeathtouchDamage = true
			}
		} else {
			if result.poison > 0 {
				added, ok := placePlayerCounters(g.Players[event.player], event.controller, counter.Poison, result.poison, simultaneousID)
				if !ok {
					panic("prepared Infect poison placement failed")
				}
				facts = append(facts, damageFact{event: added})
			}
			if result.lifeLoss > 0 {
				facts = append(facts, damageFact{event: commitLifeLoss(g, event.player, result.lifeLoss, simultaneousID)})
			}
			for _, commander := range result.commanders {
				g.Players[event.player].CommanderDamage[commander] += event.amount
			}
			if result.toxic > 0 {
				added, ok := placePlayerCounters(g.Players[event.player], event.controller, counter.Poison, result.toxic, simultaneousID)
				if !ok {
					panic("prepared toxic placement failed")
				}
				facts = append(facts, damageFact{event: added})
			}
		}
		facts = append(facts, damageFact{event: result.event, prepared: true})
		if combat && event.combatDamage {
			if event.permanent != nil {
				log.addCreatureDamage(CreatureDamageLog{
					SourcePermanent: event.sourceObjectID, SourceID: event.sourceID, Controller: event.controller,
					DamagedPermanent: event.permanent.ObjectID, DamagedSourceID: event.permanent.CardInstanceID,
					DamagedController: result.event.Player, Damage: event.amount,
				})
			} else {
				log.addCombatDamage(CombatDamageLog{
					Attacker: event.sourceObjectID, SourceID: event.sourceID, Controller: event.controller,
					DefendingPlayer: event.player, Damage: event.amount,
				})
			}
		}
	}
	for _, gain := range gains {
		if gain.amount > 0 {
			facts = append(facts, damageFact{event: commitLifeGain(g, gain.controller, gain.amount, simultaneousID)})
		}
	}
	for _, fact := range facts {
		if fact.prepared {
			emitPreparedEvent(g, fact.event)
		} else {
			emitEvent(g, fact.event)
		}
	}
	batch.finish()
	actual = actual[:len(events)]
	for _, result := range results {
		if result.initiative {
			setInitiative(g, result.damage.event.controller)
		}
		if result.ring {
			for _, player := range g.Players {
				if player.ID != result.damage.event.controller && !player.Eliminated {
					loseLife(g, player.ID, ringBearerLoseLifeOnDamage)
				}
			}
		}
	}
	return actual
}
