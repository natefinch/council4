package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/zone"
)

// untilSourceLeavesSource returns the battlefield ObjectID of the permanent
// whose triggered or activated ability is resolving an "exile ... until this
// leaves the battlefield" instruction. It reports false when that permanent has
// already left the battlefield, in which case CR 610.3a-b says the object
// doesn't move. A resolving spell has no battlefield source for this duration.
func untilSourceLeavesSource(g *game.Game, obj *game.StackObject) (id.ID, bool) {
	if obj == nil || (obj.Kind != game.StackActivatedAbility && obj.Kind != game.StackTriggeredAbility) {
		return 0, false
	}
	permanent, ok := permanentByObjectID(g, obj.SourceID)
	if !ok || (obj.SourceCardID != 0 && permanent.CardInstanceID != obj.SourceCardID) {
		return 0, false
	}
	return permanent.ObjectID, true
}

// startUntilSourceLeaves begins an instruction's "until this leaves the
// battlefield" exile. Without the duration it returns a nil recorder and
// proceeds. With it, proceed is false when the source already left, because
// CR 610.3a-b says nothing is exiled.
func (r *effectResolver) startUntilSourceLeaves(untilSourceLeaves bool) (until *untilSourceLeavesRecorder, proceed bool) {
	if !untilSourceLeaves {
		return nil, true
	}
	sourceObjectID, present := untilSourceLeavesSource(r.game, r.obj)
	if !present {
		return nil, false
	}
	return &untilSourceLeavesRecorder{sourceObjectID: sourceObjectID}, true
}

// untilSourceLeavesRecorder accumulates the exact card incarnations one
// UntilSourceLeaves move put into exile.
type untilSourceLeavesRecorder struct {
	sourceObjectID id.ID
	objects        []game.LinkedObjectRef
}

// add records every card component of a moved permanent that reached exile.
// Tokens have no card component and cease to exist, so they never return.
func (u *untilSourceLeavesRecorder) add(g *game.Game, result permanentZoneMoveResult) {
	if u == nil || !result.moved || result.destination != zone.Exile {
		return
	}
	for _, arrival := range result.arrivals {
		if from, ok := cardZone(g, arrival.CardID); !ok || from != zone.Exile {
			continue
		}
		u.objects = append(u.objects, game.LinkedObjectRef{
			ObjectID:           result.permanent.ObjectID,
			CardID:             arrival.CardID,
			CardZoneVersion:    arrival.ZoneVersion,
			CardZoneVersionSet: true,
		})
	}
}

// commit stores the duration on the game. If the source left the battlefield
// during the move itself (it exiled itself), the return happens at the next
// flush, immediately after this instruction.
func (u *untilSourceLeavesRecorder) commit(g *game.Game) {
	if u == nil || len(u.objects) == 0 {
		return
	}
	g.UntilSourceLeavesExiles = append(g.UntilSourceLeavesExiles, game.UntilSourceLeavesExile{
		SourceObjectID: u.sourceObjectID,
		Objects:        u.objects,
	})
}

// returnUntilSourceLeavesExiles performs CR 610.3's second one-shot effect for
// every recorded source that is no longer on the battlefield: each recorded
// card still in exile as the same incarnation returns to the battlefield under
// its owner's control (CR 610.3c). All returns created by the same flush are
// one simultaneous event (CR 610.3d). It is called immediately after events
// that can remove a permanent (after each resolving instruction, at the start
// of each state-based-action pass and before triggered abilities are put on the
// stack), so the return precedes any priority and never uses the stack. It
// reports whether any record was consumed.
func (e *Engine) returnUntilSourceLeavesExiles(g *game.Game, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	if len(g.UntilSourceLeavesExiles) == 0 {
		return false
	}
	var departed []game.UntilSourceLeavesExile
	kept := g.UntilSourceLeavesExiles[:0:0]
	for _, record := range g.UntilSourceLeavesExiles {
		if _, onBattlefield := permanentByObjectID(g, record.SourceObjectID); onBattlefield {
			kept = append(kept, record)
			continue
		}
		departed = append(departed, record)
	}
	if len(departed) == 0 {
		return false
	}
	// Consume the records before moving anything so a nested flush during the
	// return (an entering permanent's replacement or choice) cannot repeat it.
	g.UntilSourceLeavesExiles = kept
	var resolved []resolvedBattlefieldCard
	seen := make(map[id.ID]bool)
	for _, record := range departed {
		for _, ref := range record.Objects {
			if seen[ref.CardID] {
				continue
			}
			card, ok := linkedCardInstance(g, ref)
			if !ok {
				continue
			}
			if from, ok := cardZone(g, card.ID); !ok || from != zone.Exile {
				continue
			}
			seen[ref.CardID] = true
			resolved = append(resolved, resolvedBattlefieldCard{card: card, fromZone: zone.Exile, controller: card.Owner})
		}
	}
	if len(resolved) > 0 {
		r := newEffectResolver(e, g, nil, agents, log)
		r.putResolvedCardsOnBattlefieldCollecting(resolved, nil, permanentCreationOptions{})
	}
	return true
}

// returnUntilSourceLeavesExilesInContext flushes pending returns outside a
// resolving instruction, using the game's current choice context for any
// entry choices (an Aura's attachment) and replacement choices.
func (e *Engine) returnUntilSourceLeavesExilesInContext(g *game.Game) bool {
	if len(g.UntilSourceLeavesExiles) == 0 {
		return false
	}
	var agents [game.NumPlayers]PlayerAgent
	var log *TurnLog
	if ctx, ok := replacementChoiceContextFor(g); ok {
		agents, log = ctx.agents, ctx.log
	}
	return e.returnUntilSourceLeavesExiles(g, agents, log)
}
