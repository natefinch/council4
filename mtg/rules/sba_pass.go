package rules

import (
	"cmp"
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

type stateBasedDeparture struct {
	permanent  *game.Permanent
	snapshot   game.ObjectSnapshot
	move       opt.V[PermanentDeathReason]
	destroy    bool
	sacrifice  bool
	unattached bool
}

type stateBasedAttachment struct {
	permanent *game.Permanent
	unbestow  bool
}

type stateBasedCounterPair struct {
	permanent *game.Permanent
	count     int
}

type stateBasedToken struct {
	cards *zone.Zone
	card  id.ID
}

type stateBasedCommander struct {
	card        game.ZoneCardSnapshot
	owner       game.PlayerID
	event       game.Event
	replacement zoneChangeReplacementResult
}

type stateBasedActionPass struct {
	departures  []stateBasedDeparture
	attachments []stateBasedAttachment
	counters    []stateBasedCounterPair
	tokens      []stateBasedToken
	losses      []LossLog
	commanders  []stateBasedCommander
}

func prepareStateBasedActionPass(g *game.Game, commanders []game.Event) stateBasedActionPass {
	g.BeginStaticSourceFrame()
	defer g.EndStaticSourceFrame()
	pass := stateBasedActionPass{
		departures: permanentStateBasedDepartures(g),
		counters:   counterStateBasedActions(g),
		tokens:     tokenStateBasedActions(g),
		losses:     playerStateBasedLosses(g),
	}
	legends := legendaryRuleStateBasedActionGroups(g)
	attachments, changes := attachmentStateBasedActions(g)
	pass.attachments = changes
	pass.addDepartures(attachments)
	pass.addDepartures(stateBasedDeparturesForIDs(g, roleRuleStateBasedActionCandidates(g), PermanentDeathReasonRoleRule))
	for _, player := range triggerAPNAPPlayers(g) {
		pass.commanders = append(pass.commanders, prepareCommanderStateBasedActionsForPlayer(g, commanders, player)...)
		pass.addDepartures(stateBasedDeparturesForIDs(g, chooseLegendaryRuleStateBasedActions(g, legends, player), PermanentDeathReasonLegendaryRule))
	}
	return pass
}

func (pass *stateBasedActionPass) addDepartures(departures []stateBasedDeparture) {
	for _, departure := range departures {
		index := slices.IndexFunc(pass.departures, func(other stateBasedDeparture) bool {
			return other.permanent.ObjectID == departure.permanent.ObjectID
		})
		if index < 0 {
			pass.departures = append(pass.departures, departure)
			continue
		}
		member := &pass.departures[index]
		member.destroy = member.destroy || departure.destroy
		member.sacrifice = member.sacrifice || departure.sacrifice
		member.unattached = member.unattached || departure.unattached
		if !member.move.Exists {
			member.move = departure.move
		}
	}
}

func permanentStateBasedDepartures(g *game.Game) []stateBasedDeparture {
	var pending []stateBasedDeparture
	g.BeginStaticSourceFrame()
	defer g.EndStaticSourceFrame()
	for _, permanent := range g.Battlefield {
		if !activeBattlefieldPermanent(permanent) {
			continue
		}
		reason, applies := permanentDeathReason(g, permanent)
		if !applies {
			continue
		}
		departure := stateBasedDeparture{permanent: permanent, snapshot: snapshotPermanent(g, permanent, zone.Battlefield)}
		if permanentDeathBypassesDestroy(reason) {
			departure.move = opt.Val(reason)
			departure.sacrifice = reason == PermanentDeathReasonSagaComplete
		} else {
			departure.destroy = true
		}
		pending = append(pending, departure)
	}
	return pending
}

func stateBasedDeparturesForIDs(g *game.Game, ids []id.ID, reason PermanentDeathReason) []stateBasedDeparture {
	pending := make([]stateBasedDeparture, 0, len(ids))
	for _, objectID := range ids {
		if permanent, ok := permanentByObjectID(g, objectID); ok {
			pending = append(pending, stateBasedDeparture{
				permanent: permanent, snapshot: snapshotPermanent(g, permanent, zone.Battlefield), move: opt.Val(reason),
			})
		}
	}
	return pending
}

func attachmentStateBasedActions(g *game.Game) ([]stateBasedDeparture, []stateBasedAttachment) {
	var departures []stateBasedDeparture
	var changes []stateBasedAttachment
	g.BeginStaticSourceFrame()
	defer g.EndStaticSourceFrame()
	for _, permanent := range g.Battlefield {
		if !activeBattlefieldPermanent(permanent) {
			continue
		}
		if permanent.AttachedToPlayer.Exists {
			if !auraCanAttachToPlayer(g, permanent, permanent.AttachedToPlayer.Val) {
				departures = append(departures, stateBasedDeparture{
					permanent: permanent, snapshot: snapshotPermanent(g, permanent, zone.Battlefield),
					move: opt.Val(PermanentDeathReasonIllegalAura),
				})
			}
			continue
		}
		unattached := !permanent.AttachedTo.Exists
		if !unattached {
			host, exists := permanentByObjectID(g, permanent.AttachedTo.Val)
			if exists && canAttachPermanent(g, permanent, host) {
				continue
			}
		}
		switch {
		case permanent.Bestowed:
			changes = append(changes, stateBasedAttachment{permanent: permanent, unbestow: true})
		case isAuraPermanent(g, permanent):
			departures = append(departures, stateBasedDeparture{
				permanent: permanent, snapshot: snapshotPermanent(g, permanent, zone.Battlefield),
				move: opt.Val(PermanentDeathReasonIllegalAura), unattached: unattached,
			})
		case !unattached:
			changes = append(changes, stateBasedAttachment{permanent: permanent})
		default:
		}
	}
	return departures, changes
}

func counterStateBasedActions(g *game.Game) []stateBasedCounterPair {
	var pending []stateBasedCounterPair
	for _, permanent := range g.Battlefield {
		if activeBattlefieldPermanent(permanent) {
			count := min(permanent.Counters.Get(counter.PlusOnePlusOne), permanent.Counters.Get(counter.MinusOneMinusOne))
			if count > 0 {
				pending = append(pending, stateBasedCounterPair{permanent: permanent, count: count})
			}
		}
	}
	return pending
}

func tokenStateBasedActions(g *game.Game) []stateBasedToken {
	var pending []stateBasedToken
	for _, player := range g.Players {
		for _, cards := range []*zone.Zone{&player.Library, &player.Hand, &player.Graveyard, &player.Exile, &player.CommandZone} {
			for _, card := range cards.All() {
				if g.CardInstances[card] == nil {
					pending = append(pending, stateBasedToken{cards: cards, card: card})
				}
			}
		}
	}
	return pending
}

func playerStateBasedLosses(g *game.Game) []LossLog {
	var pending []LossLog
	for _, player := range g.Players {
		if !player.Eliminated && (player.Life <= 0 || player.HasLethalPoison() ||
			player.HasLethalCommanderDamage() || g.FailedDraws[player.ID] || g.MarkedToLoseGame[player.ID]) {
			pending = append(pending, LossLog{Player: player.ID, Reason: lossReason(g, player)})
		}
	}
	return pending
}

func prepareCommanderStateBasedActions(g *game.Game, events []game.Event) []stateBasedCommander {
	var pending []stateBasedCommander
	for _, owner := range triggerAPNAPPlayers(g) {
		pending = append(pending, prepareCommanderStateBasedActionsForPlayer(g, events, owner)...)
	}
	return pending
}

func prepareCommanderStateBasedActionsForPlayer(g *game.Game, events []game.Event, owner game.PlayerID) []stateBasedCommander {
	var pending []stateBasedCommander
	for _, event := range events {
		card, ok := g.GetCardInstance(event.CardID)
		if !ok || card.Owner != owner || card.ZoneVersion != event.CardZoneVersion || !chooseCommanderCommandZone(g, card.ID) {
			continue
		}
		move := game.Event{Kind: game.EventZoneChanged, Controller: owner, Player: owner, CardID: card.ID,
			FromZone: event.ToZone, ToZone: zone.Command}
		replacement := replacementZoneChange(g, move)
		replacement.destination = commanderReplacementDestination(g, card.ID, replacement.destination)
		pending = append(pending, stateBasedCommander{
			card:  game.ZoneCardSnapshot{CardID: card.ID, ZoneVersion: card.ZoneVersion},
			owner: owner, event: move, replacement: replacement,
		})
	}
	return pending
}

func (pass *stateBasedActionPass) apply(g *game.Game, batchID func() id.ID) (bool, []PermanentDeathLog) {
	changed := false
	var moves []preparedPermanentZoneMove
	var replacements []plannedDestroyReplacement
	planned := make(map[id.ID]bool)
	cancellations := make(map[id.ID]int, len(pass.counters))
	for _, pair := range pass.counters {
		cancellations[pair.permanent.ObjectID] = pair.count
	}
	selected := make(map[id.ID]stateBasedDeparture, len(pass.departures))
	for _, departure := range pass.departures {
		selected[departure.permanent.ObjectID] = departure
	}
	// Replacement choices and all departure snapshots precede any pass mutation.
	order := slices.Clone(pass.departures)
	rank := make(map[game.PlayerID]int)
	for index, player := range triggerAPNAPPlayers(g) {
		rank[player] = index
	}
	slices.SortStableFunc(order, func(first, second stateBasedDeparture) int {
		firstRank, firstKnown := rank[first.snapshot.Controller]
		secondRank, secondKnown := rank[second.snapshot.Controller]
		if !firstKnown {
			firstRank = len(rank)
		}
		if !secondKnown {
			secondRank = len(rank)
		}
		return cmp.Compare(firstRank, secondRank)
	})
	for _, departure := range order {
		planned[departure.permanent.ObjectID] = true
		move := departure.move.Exists
		if departure.destroy {
			replacement, replaced := planDestroyPermanentReplacement(g, departure.permanent, false, batchID(), destructionFromStateBasedAction)
			if replaced {
				replacements = append(replacements, plannedDestroyReplacement{permanent: departure.permanent, replacement: replacement})
			} else {
				move = true
			}
		}
		if move {
			prepared, ok := preparePermanentZoneMove(g, departure.permanent, zone.Graveyard)
			if !ok {
				panic("applicable state-based departure could not be prepared")
			}
			prepared.snapshot = departure.snapshot
			prepared.event.SimultaneousID = batchID()
			moves = append(moves, prepared)
		}
	}
	// Umbra armor's additional destructions belong to this event. Plan them
	// while the pre-pass battlefield still exists, including their replacements.
	for index := 0; index < len(replacements); index++ {
		replacement := replacements[index].replacement
		if replacement.kind != destroyUmbraArmor || planned[replacement.auraID] {
			continue
		}
		planned[replacement.auraID] = true
		aura, ok := permanentByObjectID(g, replacement.auraID)
		if !ok {
			panic("chosen umbra armor source disappeared before the pass")
		}
		destroyed, additional := planDestroyPermanents(g, []*game.Permanent{aura}, false, batchID())
		replacements = append(replacements, additional...)
		moves = append(moves, preparePermanentZoneMovesInBatch(g, destroyed, zone.Graveyard, batchID())...)
	}
	// Canceled counters cannot travel with a counter-retaining card, but its
	// pre-pass LKI retains them, including additional replacement departures.
	for index := range moves {
		move := &moves[index]
		if count := cancellations[move.permanent.ObjectID]; count > 0 {
			move.retained.Remove(counter.PlusOnePlusOne, count)
			move.retained.Remove(counter.MinusOneMinusOne, count)
			for component := range move.componentMoves {
				move.componentMoves[component].retained.Remove(counter.PlusOnePlusOne, count)
				move.componentMoves[component].retained.Remove(counter.MinusOneMinusOne, count)
			}
		}
	}
	var sacrifices []game.Event
	for _, move := range moves {
		if departure, exists := selected[move.permanent.ObjectID]; exists && departure.sacrifice {
			sacrifices = append(sacrifices, prepareEvent(g, game.Event{
				Kind: game.EventPermanentSacrificed, Controller: departure.snapshot.Controller,
				Player: departure.snapshot.Controller, PermanentID: move.permanent.ObjectID,
				CardID: move.permanent.CardInstanceID, Face: departure.snapshot.Face, FaceDown: departure.snapshot.FaceDown,
				FromZone: zone.Battlefield,
				ToZone:   move.actualDestination, SimultaneousID: batchID(),
			}))
		}
	}
	for _, pair := range pass.counters {
		pair.permanent.Counters.Remove(counter.PlusOnePlusOne, pair.count)
		pair.permanent.Counters.Remove(counter.MinusOneMinusOne, pair.count)
		changed = true
	}
	for _, attachment := range pass.attachments {
		if attachment.unbestow {
			attachment.permanent.Bestowed = false
		}
		detachPermanent(g, attachment.permanent)
		changed = true
	}
	batch := &destroyBatch{game: g, simultaneousID: 0, planned: planned}
	if len(replacements) > 0 {
		batch.simultaneousID = batchID()
	}
	for _, replacement := range replacements {
		batch.changed = applyDestroyReplacement(g, replacement.permanent, replacement.replacement, batch) || batch.changed
	}
	for _, queued := range batch.queued {
		if !planned[queued] {
			panic("unplanned additional state-based destruction")
		}
	}
	batch.applyQueued()
	changed = changed || batch.changed
	results := applyPreparedPermanentZoneMoves(g, moves)
	var deaths []PermanentDeathLog
	for index, result := range results {
		if !result.moved {
			continue
		}
		changed = true
		departure, selectedDeparture := selected[result.permanent.ObjectID]
		if !selectedDeparture {
			continue
		}
		if departure.unattached {
			certifyAttachedDeathDeparture(g, &moves[index], true)
		}
		if result.destination == zone.Command {
			continue
		}
		reason := PermanentDeathReasonLethalDamage
		if departure.move.Exists {
			reason = departure.move.Val
		}
		deaths = append(deaths, PermanentDeathLog{
			Permanent: result.permanent.ObjectID, SourceID: result.permanent.CardInstanceID,
			TokenName: permanentTokenName(result.permanent), Owner: departure.snapshot.Owner,
			Controller: departure.snapshot.Controller, Reason: reason,
		})
	}
	for _, event := range sacrifices {
		for index, result := range results {
			if result.moved && result.permanent.ObjectID == event.PermanentID {
				event.CardZoneVersion = moves[index].event.CardZoneVersion
				emitPreparedEvent(g, event)
				break
			}
		}
	}
	for _, commander := range pass.commanders {
		card, exists := g.GetCardInstance(commander.card.CardID)
		if exists && card.ZoneVersion == commander.card.ZoneVersion {
			changed = moveCardBetweenZonesAfterReplacement(g, commander.owner, card.ID, commander.event.FromZone,
				commander.replacement, commander.event, zoneMoveOptions{}, batchID()) || changed
		}
	}
	for _, token := range pass.tokens {
		if token.cards.Remove(token.card) {
			delete(g.CardCounters, token.card)
			changed = true
		}
	}
	for _, loss := range pass.losses {
		if !markPlayerEliminated(g, loss.Player) {
			panic("selected simultaneous loss could not be applied")
		}
		changed = true
	}
	for _, loss := range pass.losses {
		// Departure cleanup is an immediate consequence of losing, after the
		// selected simultaneous pass, never postponed to a priority window.
		passInitiativeOnElimination(g, loss.Player)
		cleanupEliminatedPlayer(g, loss.Player)
	}
	for _, player := range g.Players {
		if player.Eliminated {
			delete(g.FailedDraws, player.ID)
			delete(g.MarkedToLoseGame, player.ID)
		}
	}
	if g.Mode != game.RunModeGoldfish && len(pass.losses) > 0 && g.TurnOrder.ActivePlayerCount() == 0 {
		for _, loss := range pass.losses {
			g.DrawParticipants[loss.Player] = true
		}
	}
	return changed, deaths
}
