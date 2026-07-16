package rules

import (
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func emitEvent(g *game.Game, event game.Event) {
	if event.Kind == game.EventDamageDealt && event.DamageSourceName == "" {
		event.DamageSourceName = damageEventSourceName(g, &event)
	}
	if event.Kind == game.EventDamageDealt && event.SourceObjectID != 0 {
		if source, ok := permanentByObjectID(g, event.SourceObjectID); ok {
			event.DamageSourceHadTrample = hasKeyword(g, source, game.Trample)
		} else if snapshot, ok := lastKnownObject(g, event.SourceObjectID); ok {
			event.DamageSourceHadTrample = slices.Contains(snapshot.Keywords, game.Trample)
		}
	}
	if event.Kind == game.EventSpellCast && event.PlayerEventOrdinalThisTurn == 0 {
		event.PlayerEventOrdinalThisTurn = nextSpellCastOrdinalThisTurn(g, event.Controller)
	}
	if event.Kind == game.EventCardDiscarded && event.PlayerEventOrdinalThisTurn == 0 {
		event.PlayerEventOrdinalThisTurn = discardBatchOrdinalThisTurn(g, event.Player, event.SimultaneousID)
	}
	if event.Kind == game.EventCycled && event.PlayerEventOrdinalThisTurn == 0 {
		event.PlayerEventOrdinalThisTurn = nextPlayerEventOrdinalThisTurn(g, game.EventCycled, event.Player)
	}
	if event.Kind == game.EventCardDrawn || event.Kind == game.EventBeginningOfStep {
		event.TriggeredAbilities = captureEventTriggeredAbilities(g, event)
		event.TriggeredAbilitiesCaptured = true
	}
	if !event.TriggeredAbilitiesCaptured {
		if doublers := captureChosenTypeTriggerDoublers(g); len(doublers) > 0 {
			event.ChosenTypeTriggerDoublers = &game.ChosenTypeTriggerDoublerSnapshot{Doublers: doublers}
		}
	}
	g.AppendEvent(event)
}

func nextPlayerEventOrdinalThisTurn(g *game.Game, kind game.EventKind, playerID game.PlayerID) int {
	return eventsThisTurnWindow(g).nextOrdinal(eventKindPlayer(kind, playerID))
}

// discardBatchOrdinalThisTurn reports the per-turn ordinal position of the
// discard occurrence the event about to be emitted belongs to, counting
// distinct discard "batches" by the same player this turn (CR 701.8e; "the
// first time you discard one or more cards each turn", Rielle). Cards discarded
// together share a nonzero SimultaneousID and form one occurrence; a single
// discard carries SimultaneousID 0 and is its own occurrence. The first
// occurrence each turn is ordinal 1, so a first-each-turn discard trigger gates
// on PlayerEventOrdinalThisTurn == 1 regardless of how many cards it includes.
func discardBatchOrdinalThisTurn(g *game.Game, playerID game.PlayerID, simultaneousID id.ID) int {
	window := eventsThisTurnWindow(g)
	discardedByPlayer := eventKindPlayer(game.EventCardDiscarded, playerID)
	if simultaneousID != 0 {
		for i := range window {
			if discardedByPlayer(window[i]) && window[i].SimultaneousID == simultaneousID {
				return window[i].PlayerEventOrdinalThisTurn
			}
		}
	}
	return window.distinctBatches(discardedByPlayer) + 1
}

// nextSpellCastOrdinalThisTurn reports the per-turn ordinal position of the
// spell about to be cast by controller, counting only prior EventSpellCast
// events this turn (CR 700.6, "Nth spell each turn"). Spell copies emit
// EventSpellCopied and are deliberately excluded so copies do not advance the
// count.
func nextSpellCastOrdinalThisTurn(g *game.Game, controller game.PlayerID) int {
	return eventsThisTurnWindow(g).nextOrdinal(eventKindController(game.EventSpellCast, controller))
}

func emitZoneChangeEvent(g *game.Game, event game.Event) game.Event {
	if event.CardID != 0 && event.CardZoneVersion == 0 {
		if card, ok := g.GetCardInstance(event.CardID); ok {
			card.ZoneVersion++
			event.CardZoneVersion = card.ZoneVersion
		}
	}
	if event.FromZone == zone.Exile && event.ToZone != zone.Exile {
		delete(g.AdventureCards, event.CardID)
		delete(g.SuspendedCards, event.CardID)
		delete(g.PlottedCards, event.CardID)
		delete(g.ForetoldCards, event.CardID)
		delete(g.ExileCounters, event.CardID)
		delete(g.ExileCounterExiledBy, event.CardID)
	}
	if event.CardID != 0 && event.FromZone != event.ToZone {
		clearCardCastPermissions(g, event.CardID, event.FromZone)
	}
	event.Kind = game.EventZoneChanged
	emitEvent(g, event)
	return event
}

func clearCardCastPermissions(g *game.Game, cardID game.ObjectID, fromZone zone.Type) {
	kept := g.RuleEffects[:0]
	for i := range g.RuleEffects {
		effect := &g.RuleEffects[i]
		if (effect.Kind == game.RuleEffectCastFromZone || effect.Kind == game.RuleEffectPlayFromZone) &&
			effect.AffectedCardID == cardID &&
			effect.CastFromZone == fromZone {
			continue
		}
		kept = append(kept, *effect)
	}
	g.RuleEffects = kept
}

func markCurrentTurnEventStart(g *game.Game) {
	index := g.Turn.TurnNumber - 1
	for len(g.EventTurnStarts) <= index {
		g.EventTurnStarts = append(g.EventTurnStarts, len(g.Events))
	}
	g.EventTurnStarts[index] = len(g.Events)
	g.TriggerEventCursor = len(g.Events)
}

func emitPermanentTappedEvent(g *game.Game, permanent *game.Permanent) {
	emitEvent(g, permanentTappedEvent(g, permanent, true))
}

func permanentTappedEvent(g *game.Game, permanent *game.Permanent, tapped bool) game.Event {
	kind := game.EventPermanentUntapped
	if tapped {
		kind = game.EventPermanentTapped
	}
	return game.Event{
		Kind:        kind,
		Controller:  effectiveController(g, permanent),
		Player:      permanent.Owner,
		CardID:      permanent.CardInstanceID,
		PermanentID: permanent.ObjectID,
		TokenName:   permanentTokenName(permanent),
		TokenDef:    permanent.TokenDef,
	}
}

func emitPermanentUntappedEvent(g *game.Game, permanent *game.Permanent) {
	emitEvent(g, permanentTappedEvent(g, permanent, false))
}

func setPermanentsTappedSimultaneously(g *game.Game, permanents []*game.Permanent, tapped bool) bool {
	var changed []*game.Permanent
	for _, permanent := range permanents {
		if permanent != nil && permanent.Tapped != tapped {
			changed = append(changed, permanent)
			permanent.Tapped = tapped
		}
	}
	if len(changed) == 0 {
		return false
	}
	simultaneousID := g.IDGen.Next()
	for _, permanent := range changed {
		event := permanentTappedEvent(g, permanent, tapped)
		event.SimultaneousID = simultaneousID
		emitEvent(g, event)
	}
	return true
}

func sacrificePermanent(g *game.Game, permanent *game.Permanent) bool {
	return sacrificePermanentsSimultaneously(g, []*game.Permanent{permanent})
}

func sacrificePermanentsSimultaneously(g *game.Game, permanents []*game.Permanent) bool {
	if len(permanents) == 0 {
		return false
	}
	simultaneousID := g.IDGen.Next()
	events := make([]game.Event, 0, len(permanents))
	for _, permanent := range permanents {
		if permanent == nil {
			continue
		}
		events = append(events, game.Event{
			Kind:           game.EventPermanentSacrificed,
			SimultaneousID: simultaneousID,
			Controller:     effectiveController(g, permanent),
			Player:         effectiveController(g, permanent),
			CardID:         permanent.CardInstanceID,
			PermanentID:    permanent.ObjectID,
			TokenName:      permanentTokenName(permanent),
			TokenDef:       permanent.TokenDef,
		})
	}
	if !movePermanentsToZoneSimultaneously(g, permanents, zone.Graveyard) {
		return false
	}
	succeeded := false
	for _, event := range events {
		if _, stillOnBattlefield := permanentByObjectID(g, event.PermanentID); stillOnBattlefield {
			continue
		}
		emitEvent(g, event)
		succeeded = true
	}
	return succeeded
}

func setPermanentTapped(g *game.Game, permanent *game.Permanent, tapped bool) {
	if permanent.Tapped == tapped {
		return
	}
	permanent.Tapped = tapped
	if tapped {
		emitPermanentTappedEvent(g, permanent)
		return
	}
	emitPermanentUntappedEvent(g, permanent)
}

// setPermanentTappedForMana taps a permanent and records tapped-for-mana
// provenance on the emitted event so "is tapped for mana" triggers (Wild Growth
// and the mana-additional aura family) can fire.
func setPermanentTappedForMana(g *game.Game, permanent *game.Permanent) {
	if permanent.Tapped {
		return
	}
	permanent.Tapped = true
	event := permanentTappedEvent(g, permanent, true)
	event.TappedForMana = true
	emitEvent(g, event)
}

// manaPoolColorSnapshot records, per color, how much mana the player currently
// holds. producedManaColorsSince diffs against it to learn which types a mana
// ability just added.
func manaPoolColorSnapshot(g *game.Game, playerID game.PlayerID) map[mana.Color]int {
	snapshot := make(map[mana.Color]int)
	player, ok := playerByID(g, playerID)
	if !ok {
		return snapshot
	}
	for unit, amount := range player.ManaPool.Units() {
		snapshot[unit.Color] += amount
	}
	return snapshot
}

// producedManaColorsSince returns the distinct mana types whose count in the
// player's pool grew relative to before, in WUBRG order with colorless last. It
// reports exactly the types a just-resolved mana ability added, backing the
// "one mana of any type that land produced" mana-doubler trigger.
func producedManaColorsSince(g *game.Game, playerID game.PlayerID, before map[mana.Color]int) []mana.Color {
	after := manaPoolColorSnapshot(g, playerID)
	var produced []mana.Color
	for _, c := range []mana.Color{mana.W, mana.U, mana.B, mana.R, mana.G, mana.C} {
		if after[c] > before[c] {
			produced = append(produced, c)
		}
	}
	return produced
}

// producedManaAmountSince returns the total number of mana units the player's
// pool grew by relative to before, summing the per-type increases. It reports how
// much mana a just-resolved mana ability added, so EventManaProduced can carry
// the aggregate amount even when the ability added several mixed units.
func producedManaAmountSince(g *game.Game, playerID game.PlayerID, before map[mana.Color]int) int {
	after := manaPoolColorSnapshot(g, playerID)
	total := 0
	for c, amount := range after {
		if delta := amount - before[c]; delta > 0 {
			total += delta
		}
	}
	return total
}

// manaAbilityTappedSourceSince reports whether the source permanent was tapped
// for mana while the events at or after eventsBefore were emitted, so an
// EventManaProduced event can record whether the production tapped its source.
func manaAbilityTappedSourceSince(g *game.Game, objectID id.ID, eventsBefore int) bool {
	for i := eventsBefore; i < len(g.Events); i++ {
		event := g.Events[i]
		if event.Kind == game.EventPermanentTapped && event.TappedForMana && event.PermanentID == objectID {
			return true
		}
	}
	return false
}

// recordTappedForManaProduced annotates the most recent tapped-for-mana event
// for permanentID with the mana types its tap produced, so a "one mana of any
// type that land produced" trigger (Mirari's Wake) can mirror them at
// resolution. It is a no-op when no colors were produced or no matching event
// is found.
func recordTappedForManaProduced(g *game.Game, permanentID id.ID, colors []mana.Color) {
	if len(colors) == 0 {
		return
	}
	for i := len(g.Events) - 1; i >= 0; i-- {
		event := &g.Events[i]
		if event.Kind == game.EventPermanentTapped && event.TappedForMana && event.PermanentID == permanentID {
			event.ProducedManaColors = append(event.ProducedManaColors, colors...)
			return
		}
	}
}

// manaProducedSource captures the last-known identity of a mana ability's source
// permanent so an EventManaProduced event carries correct provenance even when
// the source sacrificed itself to produce the mana (CR 603.10). It is snapshotted
// before the ability's cost is paid, so a sacrifice-for-mana land still reports
// that it was a land after it has left the battlefield.
type manaProducedSource struct {
	sourceID   id.ID
	objectID   id.ID
	controller game.PlayerID
	isLand     bool
	tokenName  string
	tokenDef   *game.CardDef
}

// captureManaProducedSource snapshots a mana ability's source identity for a
// later EventManaProduced emission.
func captureManaProducedSource(g *game.Game, permanent *game.Permanent) manaProducedSource {
	return manaProducedSource{
		sourceID:   permanent.CardInstanceID,
		objectID:   permanent.ObjectID,
		controller: effectiveController(g, permanent),
		isLand:     permanentHasType(g, permanent, types.Land),
		tokenName:  permanentTokenName(permanent),
		tokenDef:   permanent.TokenDef,
	}
}

// emitManaProducedEvent emits the authoritative "an ability added mana" event
// (EventManaProduced, CR 106.1 / 605) for a mana ability that added colors to
// recipient's pool, carrying the source's provenance, the produced types and
// total amount, whether the source was a land, and whether it tapped as part of
// the production. It is a no-op when the ability added no mana, so an ability
// that could add mana but did not (an empty color set) fires no trigger. Only
// activated mana abilities reach this path; additional-mana triggered abilities
// resolve on the stack, so they never emit it and cannot recursively retrigger
// themselves.
func emitManaProducedEvent(g *game.Game, src manaProducedSource, recipient game.PlayerID, colors []mana.Color, amount int, tappedForMana bool) {
	if amount <= 0 || len(colors) == 0 {
		return
	}
	emitEvent(g, game.Event{
		Kind:               game.EventManaProduced,
		SourceID:           src.sourceID,
		SourceObjectID:     src.objectID,
		PermanentID:        src.objectID,
		Controller:         src.controller,
		Player:             recipient,
		ProducedManaColors: append([]mana.Color(nil), colors...),
		Amount:             amount,
		ManaSourceIsLand:   src.isLand,
		TappedForMana:      tappedForMana,
		ManaAbility:        true,
		TokenName:          src.tokenName,
		TokenDef:           src.tokenDef,
	})
}

func emitTargetEvents(g *game.Game, obj *game.StackObject) {
	for _, target := range obj.Targets {
		event := game.Event{
			Kind:          game.EventObjectBecameTarget,
			StackObjectID: obj.ID,
			Controller:    obj.Controller,
			Target:        target,
		}
		event.SourceID, event.SourceObjectID = damageSourceIDs(g, obj)
		switch target.Kind {
		case game.TargetPermanent:
			event.PermanentID = target.PermanentID
		case game.TargetPlayer:
			event.Player = target.PlayerID
		default:
		}
		emitEvent(g, event)
	}
	emitCrimeEvent(g, obj)
}

// emitCrimeEvent emits an EventCrimeCommitted when putting obj on the stack
// constitutes committing a crime (CR 700.15). A crime is committed once per
// spell or ability put on the stack, regardless of how many qualifying targets
// it has, so this fires at most one event per push.
func emitCrimeEvent(g *game.Game, obj *game.StackObject) {
	if !committedCrime(g, obj) {
		return
	}
	sourceID, sourceObjectID := damageSourceIDs(g, obj)
	emitEvent(g, game.Event{
		Kind:           game.EventCrimeCommitted,
		SourceID:       sourceID,
		SourceObjectID: sourceObjectID,
		StackObjectID:  obj.ID,
		Controller:     obj.Controller,
		Player:         obj.Controller,
	})
}

// committedCrime reports whether obj targets one or more opponents of its
// controller, objects an opponent controls (permanents or spells/abilities on
// the stack), or cards in an opponent's graveyard (CR 700.15a). Targets that no
// longer resolve to a known object are ignored.
func committedCrime(g *game.Game, obj *game.StackObject) bool {
	for _, target := range obj.Targets {
		switch target.Kind {
		case game.TargetPlayer:
			if target.PlayerID != obj.Controller {
				return true
			}
		case game.TargetPermanent:
			if permanent, ok := permanentByObjectID(g, target.PermanentID); ok &&
				permanent.Controller != obj.Controller {
				return true
			}
		case game.TargetStackObject:
			if stackObj, ok := stackObjectByID(g, target.StackObjectID); ok &&
				stackObj.Controller != obj.Controller {
				return true
			}
		case game.TargetCard:
			if cardInOpponentGraveyard(g, obj.Controller, target.CardID) {
				return true
			}
		default:
		}
	}
	return false
}

// cardInOpponentGraveyard reports whether cardID currently sits in the
// graveyard of a player other than controller.
func cardInOpponentGraveyard(g *game.Game, controller game.PlayerID, cardID id.ID) bool {
	for opponent := range game.PlayerID(game.NumPlayers) {
		if opponent == controller {
			continue
		}
		player, ok := playerByID(g, opponent)
		if !ok {
			continue
		}
		if player.Graveyard.Contains(cardID) {
			return true
		}
	}
	return false
}

func emitAbilityActivatedEvent(g *game.Game, obj *game.StackObject, permanentID game.ObjectID, manaAbility bool) {
	emitEvent(g, game.Event{
		Kind:           game.EventAbilityActivated,
		SourceID:       obj.SourceCardID,
		SourceObjectID: permanentID,
		StackObjectID:  obj.ID,
		AbilityIndex:   obj.AbilityIndex,
		ManaAbility:    manaAbility,
		Controller:     obj.Controller,
		Player:         obj.Controller,
		CardID:         obj.SourceCardID,
		PermanentID:    permanentID,
		TokenDef:       obj.SourceTokenDef,
	})
}
