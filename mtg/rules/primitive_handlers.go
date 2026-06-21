package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func (r *effectResolver) quantity(q game.Quantity) int {
	if q.IsDynamic() {
		return dynamicAmountValue(r.game, r.obj, stackObjectController(r.obj), q.DynamicAmount().Val)
	}
	return q.Value()
}

func (r *effectResolver) resolveObject(object game.ObjectReference) (*game.Permanent, bool) {
	resolved, ok := resolveObjectReference(r.game, r.obj, object)
	return resolved.permanent, ok && resolved.permanent != nil
}

func (r *effectResolver) resolvePlayer(player game.PlayerReference) (game.PlayerID, bool) {
	return resolvePlayerReference(r.game, r.obj, player)
}

func (r *effectResolver) recipientController(recipient game.PlayerReference) (game.PlayerID, bool) {
	if recipient.Kind() != game.PlayerReferenceNone {
		return r.resolvePlayer(recipient)
	}
	return r.obj.Controller, true
}

func (r *effectResolver) groupPermanents(group game.GroupReference) []*game.Permanent {
	ids := newReferenceResolver(r.game, r.obj).groupMembers(group)
	permanents := make([]*game.Permanent, 0, len(ids))
	for _, permanentID := range ids {
		if permanent, ok := permanentByObjectID(r.game, permanentID); ok {
			permanents = append(permanents, permanent)
		}
	}
	return permanents
}

func (r *effectResolver) groupPermanentsWithSource(group game.GroupReference, source *game.Permanent) []*game.Permanent {
	ids := newReferenceResolverWithSource(r.game, r.obj, source).groupMembers(group)
	permanents := make([]*game.Permanent, 0, len(ids))
	for _, permanentID := range ids {
		if permanent, ok := permanentByObjectID(r.game, permanentID); ok {
			permanents = append(permanents, permanent)
		}
	}
	return permanents
}

func (r *effectResolver) playerGroupMembers(group game.PlayerGroupReference) []game.PlayerID {
	return newReferenceResolver(r.game, r.obj).playerGroup(group)
}

func playersInAPNAPOrder(g *game.Game, players []game.PlayerID) []game.PlayerID {
	included := make(map[game.PlayerID]bool, len(players))
	for _, playerID := range players {
		included[playerID] = true
	}
	ordered := make([]game.PlayerID, 0, len(included))
	playerID := g.Turn.ActivePlayer
	for range game.NumPlayers {
		if included[playerID] {
			ordered = append(ordered, playerID)
		}
		playerID = g.TurnOrder.NextActivePlayer(playerID)
		if playerID == g.Turn.ActivePlayer {
			break
		}
	}
	return ordered
}

func handleDestroy(r *effectResolver, prim game.Destroy) effectResolved {
	res := effectResolved{accepted: true}
	if prim.Group.Valid() {
		permanents := r.groupPermanents(prim.Group)
		destroyed := make([]*game.Permanent, 0, len(permanents))
		for _, permanent := range permanents {
			if hasKeyword(r.game, permanent, game.Indestructible) || replaceDestroyPermanent(r.game, permanent, prim.PreventRegeneration) {
				continue
			}
			destroyed = append(destroyed, permanent)
		}
		res.succeeded = movePermanentsToZoneSimultaneously(r.game, destroyed, zone.Graveyard)
		return res
	}
	permanent, ok := r.resolveObject(prim.Object)
	if ok {
		_, res.succeeded = destroyPermanentInBatch(r.game, permanent.ObjectID, 0, prim.PreventRegeneration)
	}
	return res
}

func handleAddMana(r *effectResolver, prim game.AddMana) effectResolved {
	res := effectResolved{accepted: true, amount: r.quantity(prim.Amount)}
	if res.amount <= 0 && !prim.Amount.IsDynamic() {
		res.amount = 1
	}
	recipientID := r.obj.Controller
	if prim.Player.Exists {
		resolved, ok := r.resolvePlayer(prim.Player.Val)
		if !ok {
			return res
		}
		recipientID = resolved
	}
	player, ok := playerByID(r.game, recipientID)
	if !ok || player.Eliminated {
		return res
	}
	if prim.EachControlledColor != nil {
		snow := stackObjectSourceIsSnow(r.game, r.obj)
		for _, c := range controlledPermanentColors(r.game, recipientID, prim.EachControlledColor) {
			if snow {
				player.ManaPool.AddSnow(c, res.amount)
			} else {
				player.ManaPool.Add(c, res.amount)
			}
			res.succeeded = true
		}
		return res
	}
	manaColor := prim.ManaColor
	if choice, ok := linkedResolutionChoice(r.obj, string(prim.ChoiceFrom)); ok && choice.Kind == game.ResolutionChoiceMana {
		manaColor = choice.Color
	}
	if prim.EntryChoiceFrom != "" {
		choice, ok := linkedResolutionChoice(r.obj, string(prim.EntryChoiceFrom))
		if !ok || choice.Kind != game.ResolutionChoiceMana {
			return res
		}
		manaColor = choice.Color
	}
	chosenSubtype := types.Sub("")
	if prim.SpendRider.Exists {
		if key := prim.SpendRider.Val.ChosenSubtypeFrom; key != "" {
			choice, ok := linkedResolutionChoice(r.obj, string(key))
			if !ok || choice.Kind != game.ResolutionChoiceSubtype {
				return res
			}
			chosenSubtype = choice.Subtype
		}
	}
	snow := stackObjectSourceIsSnow(r.game, r.obj)
	if snow {
		player.ManaPool.AddSnow(manaColor, res.amount)
	} else {
		player.ManaPool.Add(manaColor, res.amount)
	}
	if prim.SpendRider.Exists {
		unit := mana.Unit{Color: manaColor, Snow: snow}
		for i := 0; i < res.amount; i++ {
			player.ManaRiders = append(player.ManaRiders, game.ManaRiderInstance{
				Unit:           unit,
				Controller:     r.obj.Controller,
				SourceID:       r.obj.SourceCardID,
				SourceObjectID: r.obj.SourceID,
				ChosenSubtype:  chosenSubtype,
				Rider:          prim.SpendRider.Val,
			})
		}
	}
	res.succeeded = true
	return res
}

func handleAddCounter(r *effectResolver, prim game.AddCounter) effectResolved {
	res := effectResolved{accepted: true, amount: r.quantity(prim.Amount)}
	if res.amount <= 0 {
		return res
	}
	placementController := stackObjectController(r.obj)
	if prim.Group.Valid() {
		for _, permanent := range r.groupPermanents(prim.Group) {
			if addCountersToPermanentControlledBy(r.game, placementController, permanent, prim.CounterKind, res.amount) {
				res.succeeded = true
			}
		}
		return res
	}
	permanent, ok := r.resolveObject(prim.Object)
	if ok {
		addCountersToPermanentControlledBy(r.game, placementController, permanent, prim.CounterKind, res.amount)
		res.succeeded = true
	}
	return res
}

func handleAddPlayerCounter(r *effectResolver, prim game.AddPlayerCounter) effectResolved {
	res := effectResolved{accepted: true, amount: r.quantity(prim.Amount)}
	playerID, ok := r.resolvePlayer(prim.Player)
	if !ok || res.amount <= 0 {
		return res
	}
	player, ok := playerByID(r.game, playerID)
	if !ok || player.Eliminated {
		return res
	}
	if addCountersToPlayerControlledBy(r.game, stackObjectController(r.obj), player, prim.CounterKind, res.amount) {
		res.succeeded = true
	}
	return res
}

func handleMoveCounters(r *effectResolver, prim game.MoveCounters) effectResolved {
	res := effectResolved{accepted: true, amount: r.quantity(prim.Amount)}
	destination, ok := r.resolveObject(prim.Object)
	if !ok {
		return res
	}
	counters, source, ok := effectCounterSource(r.game, r.obj, prim.Source)
	if !ok || counters.IsEmpty() || source != nil && source.ObjectID == destination.ObjectID {
		return res
	}
	for kind, amount := range counters.All() {
		addCountersToPermanentControlledBy(r.game, stackObjectController(r.obj), destination, kind, amount)
		if source != nil {
			source.Counters.Remove(kind, amount)
		}
	}
	res.succeeded = true
	return res
}

func handleApplyContinuous(r *effectResolver, prim game.ApplyContinuous) effectResolved {
	res := effectResolved{accepted: true}
	var permanent *game.Permanent
	if prim.Object.Exists {
		permanent, _ = r.resolveObject(prim.Object.Val)
	}
	res.succeeded = applyTypedContinuousEffects(r.game, r.obj, permanent, prim.ContinuousEffects, prim.Duration)
	if prim.PublishLinked != "" && permanent != nil {
		rememberLinkedObject(
			r.game,
			linkedObjectSourceKey(r.game, r.obj, string(prim.PublishLinked)),
			permanentLinkedObjectRef(permanent),
		)
	}
	return res
}

func handleApplyRule(r *effectResolver, prim game.ApplyRule) effectResolved {
	return effectResolved{
		accepted:  true,
		succeeded: createRuleEffectTemplates(r.game, r.obj, prim.Object, prim.RuleEffects, prim.Duration),
	}
}

func handleModifyPT(r *effectResolver, prim game.ModifyPT) effectResolved {
	res := effectResolved{accepted: true}
	permanent, ok := r.resolveObject(prim.Object)
	if !ok || prim.Duration != game.DurationUntilEndOfTurn {
		return res
	}
	powerDelta := r.quantity(prim.PowerDelta)
	toughnessDelta := r.quantity(prim.ToughnessDelta)
	r.game.ContinuousEffects = append(r.game.ContinuousEffects, untilEndOfTurnPTContinuousEffect(r.game, r.obj, permanent, powerDelta, toughnessDelta))
	if prim.PublishLinked != "" {
		rememberLinkedObject(
			r.game,
			linkedObjectSourceKey(r.game, r.obj, string(prim.PublishLinked)),
			permanentLinkedObjectRef(permanent),
		)
	}
	res.succeeded = true
	return res
}

func handleTap(r *effectResolver, prim game.Tap) effectResolved {
	res := effectResolved{accepted: true}
	if prim.Group.Valid() {
		res.succeeded = setPermanentsTappedSimultaneously(r.game, r.groupPermanents(prim.Group), true)
		return res
	}
	if permanent, ok := r.resolveObject(prim.Object); ok {
		setPermanentTapped(r.game, permanent, true)
		res.succeeded = true
	}
	return res
}

func handleStartEngines(r *effectResolver, prim game.StartEngines) effectResolved {
	res := effectResolved{accepted: true}
	playerID, ok := r.resolvePlayer(prim.Player)
	if ok {
		res.succeeded = startEngines(r.game, playerID)
	}
	return res
}

func handleSetClassLevel(r *effectResolver, prim game.SetClassLevel) effectResolved {
	res := effectResolved{accepted: true, amount: r.quantity(prim.Amount)}
	permanent, ok := r.resolveObject(prim.Object)
	if ok && res.amount > permanent.ClassLevel {
		permanent.ClassLevel = res.amount
		res.succeeded = true
	}
	return res
}

func handleMonstrosity(r *effectResolver, prim game.Monstrosity) effectResolved {
	res := effectResolved{accepted: true, amount: r.quantity(prim.Amount)}
	permanent, ok := r.resolveObject(prim.Object)
	if ok && !permanent.Monstrous {
		if res.amount > 0 {
			addCountersToPermanentControlledBy(r.game, stackObjectController(r.obj), permanent, counter.PlusOnePlusOne, res.amount)
		}
		permanent.Monstrous = true
		res.succeeded = true
	}
	return res
}

func handlePay(r *effectResolver, prim game.Pay) effectResolved {
	payment := prim.Payment
	if payment.Prompt == "" {
		payment.Prompt = prim.Prompt
	}
	accepted, succeeded := r.engine.resolveResolutionPaymentValue(r.game, r.obj, &payment, r.agents, r.log)
	return effectResolved{accepted: accepted, succeeded: succeeded}
}

func handleChoose(r *effectResolver, prim game.Choose) effectResolved {
	succeeded := r.engine.resolveResolutionChoiceValue(r.game, r.obj, &prim.Choice, string(prim.PublishChoice), r.agents, r.log)
	return effectResolved{accepted: true, succeeded: succeeded}
}

func handleGainLife(r *effectResolver, prim game.GainLife) effectResolved {
	perPlayer := r.quantity(prim.Amount)
	res := effectResolved{accepted: true, amount: perPlayer}
	if perPlayer <= 0 {
		return res
	}
	if prim.PlayerGroup.Kind != game.PlayerGroupReferenceNone {
		res.amount = 0
		for _, playerID := range r.playerGroupMembers(prim.PlayerGroup) {
			gained := gainLife(r.game, playerID, perPlayer)
			res.amount += gained
			res.succeeded = gained > 0 || res.succeeded
		}
		return res
	}
	playerID, ok := r.resolvePlayer(prim.Player)
	if ok {
		res.succeeded = gainLife(r.game, playerID, res.amount) > 0
	}
	return res
}

func handleLoseLife(r *effectResolver, prim game.LoseLife) effectResolved {
	perPlayer := r.quantity(prim.Amount)
	res := effectResolved{accepted: true, amount: perPlayer}
	if perPlayer <= 0 {
		return res
	}
	if prim.PlayerGroup.Kind != game.PlayerGroupReferenceNone {
		res.amount = 0
		for _, playerID := range r.playerGroupMembers(prim.PlayerGroup) {
			lost := loseLife(r.game, playerID, perPlayer)
			res.amount += lost
			res.succeeded = lost > 0 || res.succeeded
		}
		return res
	}
	playerID, ok := r.resolvePlayer(prim.Player)
	if ok {
		res.succeeded = loseLife(r.game, playerID, res.amount) > 0
	}
	return res
}

func handlePlayerLosesGame(r *effectResolver, prim game.PlayerLosesGame) effectResolved {
	res := effectResolved{accepted: true}
	playerID, ok := r.resolvePlayer(prim.Player)
	if !ok {
		return res
	}
	if player, ok := playerByID(r.game, playerID); ok && player.Eliminated {
		return res
	}
	r.game.MarkedToLoseGame[playerID] = true
	res.succeeded = true
	return res
}

func handlePlayerWinsGame(r *effectResolver, prim game.PlayerWinsGame) effectResolved {
	res := effectResolved{accepted: true}
	winnerID, ok := r.resolvePlayer(prim.Player)
	if !ok {
		return res
	}
	if player, ok := playerByID(r.game, winnerID); ok && player.Eliminated {
		return res
	}
	for _, player := range r.game.Players {
		if player.ID == winnerID || player.Eliminated {
			continue
		}
		r.game.MarkedToLoseGame[player.ID] = true
		res.succeeded = true
	}
	return res
}

func handleUntap(r *effectResolver, prim game.Untap) effectResolved {
	res := effectResolved{accepted: true}
	if prim.ChooseUpTo {
		for _, permanent := range r.chooseUntapPermanents(prim) {
			setPermanentTapped(r.game, permanent, false)
			res.succeeded = true
		}
		return res
	}
	if prim.Group.Valid() {
		res.succeeded = setPermanentsTappedSimultaneously(r.game, r.groupPermanents(prim.Group), false)
		return res
	}
	if permanent, ok := r.resolveObject(prim.Object); ok {
		setPermanentTapped(r.game, permanent, false)
		res.succeeded = true
	}
	return res
}

func (r *effectResolver) chooseUntapPermanents(prim game.Untap) []*game.Permanent {
	amount := r.quantity(prim.Amount)
	if amount <= 0 {
		return nil
	}
	candidates := r.groupPermanents(prim.Group)
	maxChoices := min(amount, len(candidates))
	if maxChoices == 0 {
		return nil
	}
	options := make([]game.ChoiceOption, len(candidates))
	for i, permanent := range candidates {
		options[i] = game.ChoiceOption{
			Index: i,
			Label: permanentChoiceLabel(r.game, permanent),
			Card:  permanentChoiceInfo(r.game, permanent),
		}
	}
	selected := r.engine.chooseChoice(r.game, r.agents, game.ChoiceRequest{
		Kind:             game.ChoiceResolution,
		Player:           r.obj.Controller,
		Prompt:           "Choose permanents to untap",
		Options:          options,
		MinChoices:       0,
		MaxChoices:       maxChoices,
		DefaultSelection: firstChoiceIndices(maxChoices),
	}, r.log)
	chosen := make([]*game.Permanent, 0, len(selected))
	for _, idx := range selected {
		if idx >= 0 && idx < len(candidates) {
			chosen = append(chosen, candidates[idx])
		}
	}
	return chosen
}

func handleSkipNextUntap(r *effectResolver, prim game.SkipNextUntap) effectResolved {
	res := effectResolved{accepted: true}
	if permanent, ok := r.resolveObject(prim.Object); ok {
		permanent.Exerted = true
		res.succeeded = true
	}
	return res
}

func handleProliferate(r *effectResolver, prim game.Proliferate) effectResolved {
	amount := r.quantity(prim.Amount)
	if amount <= 0 {
		amount = 1
	}
	res := effectResolved{accepted: true, amount: amount}
	for range amount {
		if r.engine.resolveProliferate(r.game, r.obj, r.agents, r.log) {
			res.succeeded = true
		}
	}
	return res
}

func handleExplore(r *effectResolver, prim game.Explore) effectResolved {
	res := effectResolved{accepted: true}
	permanent, ok := r.resolveObject(prim.Creature)
	if !ok {
		return res
	}
	controller := effectiveController(r.game, permanent)
	if r.engine.exploreCreature(r.game, r.obj, r.agents, r.log, controller, permanent) {
		res.succeeded = true
	}
	return res
}

func handleGoad(r *effectResolver, prim game.Goad) effectResolved {
	res := effectResolved{accepted: true}
	if permanent, ok := r.resolveObject(prim.Object); ok && permanentHasType(r.game, permanent, types.Creature) {
		goadPermanent(r.game, permanent, r.obj.Controller)
		res.succeeded = true
	}
	return res
}

func handleRemoveCounter(r *effectResolver, prim game.RemoveCounter) effectResolved {
	res := effectResolved{accepted: true, amount: r.quantity(prim.Amount)}
	if res.amount <= 0 {
		return res
	}
	if prim.Group.Valid() {
		for _, permanent := range r.groupPermanents(prim.Group) {
			permanent.Counters.Remove(prim.CounterKind, res.amount)
			res.succeeded = true
		}
		return res
	}
	if permanent, ok := r.resolveObject(prim.Object); ok {
		permanent.Counters.Remove(prim.CounterKind, res.amount)
		res.succeeded = true
	}
	return res
}

func handlePhaseOut(r *effectResolver, prim game.PhaseOut) effectResolved {
	res := effectResolved{accepted: true}
	var roots []*game.Permanent
	if prim.Group.Valid() {
		roots = append(roots, r.groupPermanents(prim.Group)...)
	} else if permanent, ok := r.resolveObject(prim.Object); ok {
		roots = append(roots, permanent)
	}
	phaseOutRoots := make([]phaseOutRoot, 0, len(roots))
	for _, permanent := range roots {
		phaseOutRoots = append(phaseOutRoots, phaseOutRoot{
			permanent:  permanent,
			phaseInFor: effectiveController(r.game, permanent),
		})
	}
	res.succeeded = phaseOutPermanentTrees(r.game, phaseOutRoots)
	return res
}

func phaseOutPermanentTree(g *game.Game, permanent *game.Permanent, phaseInFor game.PlayerID, phased map[game.ObjectID]bool) bool {
	return phaseOutPermanentTreesWithSeen(g, []phaseOutRoot{{
		permanent:  permanent,
		phaseInFor: phaseInFor,
	}}, phased)
}

type phaseOutRoot struct {
	permanent  *game.Permanent
	phaseInFor game.PlayerID
}

type phaseOutCandidate struct {
	permanent  *game.Permanent
	phaseInFor game.PlayerID
	controller game.PlayerID
	snapshot   game.ObjectSnapshot
	event      game.Event
}

func phaseOutPermanentTrees(g *game.Game, roots []phaseOutRoot) bool {
	return phaseOutPermanentTreesWithSeen(g, roots, make(map[game.ObjectID]bool))
}

func phaseOutPermanentTreesWithSeen(g *game.Game, roots []phaseOutRoot, phased map[game.ObjectID]bool) bool {
	var candidates []phaseOutCandidate
	for _, root := range roots {
		collectPhaseOutPermanentTree(g, root.permanent, root.phaseInFor, phased, &candidates)
	}
	if len(candidates) == 0 {
		return false
	}

	g.BeginStaticSourceFrame()
	for i := range candidates {
		candidate := &candidates[i]
		candidate.controller = effectiveController(g, candidate.permanent)
		candidate.snapshot = snapshotPermanent(g, candidate.permanent, zone.Battlefield)
	}
	g.EndStaticSourceFrame()

	var simultaneousID id.ID
	if len(candidates) > 1 {
		simultaneousID = g.IDGen.Next()
	}
	for i := range candidates {
		candidate := &candidates[i]
		event := game.Event{
			Kind:           game.EventPermanentPhasedOut,
			Controller:     candidate.controller,
			Player:         candidate.phaseInFor,
			PermanentID:    candidate.permanent.ObjectID,
			CardID:         candidate.permanent.CardInstanceID,
			SimultaneousID: simultaneousID,
		}
		event.TriggeredAbilities = captureEventTriggeredAbilities(g, event)
		event.TriggeredAbilitiesCaptured = true
		candidate.event = event
	}

	for i := range candidates {
		candidate := &candidates[i]
		rememberLastKnown(g, &candidate.snapshot)
		candidate.permanent.PhasedOut = true
		candidate.permanent.PhasedOutFor = candidate.phaseInFor
		candidate.permanent.PhaseInScheduled = true
		removePermanentFromCombat(g, candidate.permanent.ObjectID)
	}
	for i := range candidates {
		emitEvent(g, candidates[i].event)
	}
	return true
}

func collectPhaseOutPermanentTree(
	g *game.Game,
	permanent *game.Permanent,
	phaseInFor game.PlayerID,
	phased map[game.ObjectID]bool,
	candidates *[]phaseOutCandidate,
) {
	if permanent == nil || permanent.PhasedOut || phased[permanent.ObjectID] {
		return
	}
	phased[permanent.ObjectID] = true
	*candidates = append(*candidates, phaseOutCandidate{
		permanent:  permanent,
		phaseInFor: phaseInFor,
	})
	for _, attachmentID := range permanent.Attachments {
		if attachment, ok := permanentByObjectID(g, attachmentID); ok {
			collectPhaseOutPermanentTree(g, attachment, phaseInFor, phased, candidates)
		}
	}
}

func handleRegenerate(r *effectResolver, prim game.Regenerate) effectResolved {
	res := effectResolved{accepted: true}
	if permanent, ok := r.resolveObject(prim.Object); ok {
		permanent.RegenerationShields++
		res.succeeded = true
	}
	return res
}

func handleAttach(r *effectResolver, prim game.Attach) effectResolved {
	res := effectResolved{accepted: true}
	attachment, ok := r.resolveObject(prim.Attachment)
	if !ok {
		return res
	}
	target, ok := r.resolveObject(prim.Target)
	if !ok {
		return res
	}
	if attachPermanent(r.game, attachment, target) {
		res.succeeded = true
	}
	return res
}

func handleSkipStep(r *effectResolver, prim game.SkipStep) effectResolved {
	res := effectResolved{accepted: true}
	if playerID, ok := r.resolvePlayer(prim.Player); ok {
		scheduleSkipStep(r.game, playerID, prim.Step)
		res.succeeded = true
	}
	return res
}

func handleCreateEmblem(r *effectResolver, prim game.CreateEmblem) effectResolved {
	r.game.Emblems = append(r.game.Emblems, game.Emblem{
		Owner:     r.obj.Controller,
		Abilities: append([]game.Ability(nil), prim.EmblemAbilities...),
	})
	return effectResolved{accepted: true, succeeded: true}
}

func handleCreateDelayedTrigger(r *effectResolver, prim game.CreateDelayedTrigger) effectResolved {
	return effectResolved{accepted: true, succeeded: scheduleDelayedTrigger(r.game, r.obj, &prim.Trigger)}
}

func handleCreateReplacement(r *effectResolver, prim game.CreateReplacement) effectResolved {
	replacement := *prim.Replacement
	replacement.ID = r.game.IDGen.Next()
	replacement.Controller = r.obj.Controller
	replacement.SourceCardID, replacement.SourceObjectID = damageSourceIDs(r.game, r.obj)
	replacement.CreatedTurn = r.game.Turn.TurnNumber
	if prim.Duration != game.DurationPermanent {
		replacement.Duration = prim.Duration
	}
	r.game.ReplacementEffects = append(r.game.ReplacementEffects, replacement)
	return effectResolved{accepted: true, succeeded: true}
}

func applyTypedContinuousEffects(g *game.Game, obj *game.StackObject, permanent *game.Permanent, templates []game.ContinuousEffect, duration game.EffectDuration) bool {
	if len(templates) == 0 {
		return false
	}
	// Source-tied durations require the resolving ability to come from a
	// battlefield permanent (an activated or triggered ability).  Spell sources
	// are not permanents, so a source-tied duration would be immediately stale.
	// Fail closed here rather than silently creating a permanent-duration effect.
	if duration == game.DurationForAsLongAsSourceOnBattlefield ||
		duration == game.DurationForAsLongAsYouControlSource {
		if obj.Kind != game.StackActivatedAbility && obj.Kind != game.StackTriggeredAbility {
			return false
		}
		if _, ok := permanentByObjectID(g, obj.SourceID); !ok {
			return false
		}
	}
	sourceID, sourceObjectID := damageSourceIDs(g, obj)
	timestamp := game.Timestamp(g.IDGen.Next())
	applied := false
	for i := range templates {
		runtimeEffect := templates[i]
		runtimeEffect.SourceCardID = sourceID
		runtimeEffect.SourceObjectID = sourceObjectID
		runtimeEffect.Controller = obj.Controller
		runtimeEffect.Timestamp = timestamp
		runtimeEffect.CreatedTurn = g.Turn.TurnNumber
		snapshotContinuousX(g, obj, &runtimeEffect)
		if duration != game.DurationPermanent {
			runtimeEffect.Duration = duration
		}
		if runtimeEffect.Duration == game.DurationUntilYourNextTurn && runtimeEffect.ExpiresFor == game.Player1 {
			runtimeEffect.ExpiresFor = obj.Controller
		}
		if runtimeEffect.NewController.Exists && runtimeEffect.NewController.Val == game.Player1 {
			runtimeEffect.NewController = opt.Val(obj.Controller)
		}
		if runtimeEffect.Group.Valid() {
			members := newReferenceResolver(g, obj).groupMembers(runtimeEffect.Group)
			runtimeEffect.Group = game.GroupReference{}
			for _, objectID := range members {
				memberEffect := runtimeEffect
				memberEffect.ID = g.IDGen.Next()
				memberEffect.AffectedObjectID = objectID
				g.ContinuousEffects = append(g.ContinuousEffects, memberEffect)
			}
			applied = true
			continue
		}
		if runtimeEffect.AffectedObjectID == 0 {
			if permanent == nil {
				continue
			}
			runtimeEffect.AffectedObjectID = permanent.ObjectID
		}
		runtimeEffect.ID = g.IDGen.Next()
		g.ContinuousEffects = append(g.ContinuousEffects, runtimeEffect)
		applied = true
	}
	return applied
}

// snapshotContinuousX locks a one-shot continuous effect's dynamic power and
// toughness deltas to fixed values at resolution. A mass pump such as
// "Creatures you control get +X/+X until end of turn, where X is the number of
// creatures you control." computes X once when the spell or ability resolves,
// so every dynamic delta kind (the spell's X, a battlefield count, a greatest
// characteristic, …) is evaluated here and frozen rather than re-evaluated each
// time the continuous effect applies.
func snapshotContinuousX(g *game.Game, obj *game.StackObject, effect *game.ContinuousEffect) {
	if effect.PowerDeltaDynamic.Exists {
		effect.PowerDelta += dynamicAmountValue(g, obj, obj.Controller, effect.PowerDeltaDynamic.Val)
		effect.PowerDeltaDynamic.Exists = false
	}
	if effect.ToughnessDeltaDynamic.Exists {
		effect.ToughnessDelta += dynamicAmountValue(g, obj, obj.Controller, effect.ToughnessDeltaDynamic.Val)
		effect.ToughnessDeltaDynamic.Exists = false
	}
}
