package rules

import (
	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

func (c *programContext) Exchange(first, second game.ObjectReference) (cardprogram.Exchange, error) {
	if err := c.valid(); err != nil {
		return cardprogram.Exchange{}, err
	}
	for _, reference := range []game.ObjectReference{first, second} {
		if err := game.ValidateScopedObjectReference(reference, c.specs); err != nil {
			return cardprogram.Exchange{}, err
		}
	}
	resolver := newEffectResolver(c.engine, c.game, c.obj, c.agents, c.log)
	a, firstAvailable := resolver.resolveObject(first)
	b, secondAvailable := resolver.resolveObject(second)
	outcome := cardprogram.Exchange{
		Available: firstAvailable && secondAvailable && activeBattlefieldPermanent(a) && activeBattlefieldPermanent(b),
	}
	if !outcome.Available {
		return outcome, nil
	}
	outcome.Before = []game.ObjectSnapshot{currentProgramSnapshot(c.game, a), currentProgramSnapshot(c.game, b)}
	if a.ObjectID == b.ObjectID {
		return outcome, nil
	}
	firstController, secondController := effectiveController(c.game, a), effectiveController(c.game, b)
	if firstController == secondController ||
		!isPlayerAlive(c.game, firstController) || !isPlayerAlive(c.game, secondController) {
		return outcome, nil
	}
	sourceID, sourceObjectID := damageSourceIDs(c.game, c.obj)
	timestamp := game.Timestamp(c.game.IDGen.Next())
	effects := []game.ContinuousEffect{
		{
			ID: c.game.IDGen.Next(), SourceCardID: sourceID, SourceObjectID: sourceObjectID,
			Controller: c.obj.Controller, Timestamp: timestamp, CreatedTurn: c.game.Turn.TurnNumber,
			AffectedObjectID: a.ObjectID, Layer: game.LayerControl, NewController: opt.Val(secondController),
			Duration: game.DurationPermanent,
		},
		{
			ID: c.game.IDGen.Next(), SourceCardID: sourceID, SourceObjectID: sourceObjectID,
			Controller: c.obj.Controller, Timestamp: timestamp, CreatedTurn: c.game.Turn.TurnNumber,
			AffectedObjectID: b.ObjectID, Layer: game.LayerControl, NewController: opt.Val(firstController),
			Duration: game.DurationPermanent,
		},
	}
	c.game.ContinuousEffects = append(c.game.ContinuousEffects, effects...)
	a.SummoningSick, b.SummoningSick = true, true
	removePermanentFromCombat(c.game, a.ObjectID)
	removePermanentFromCombat(c.game, b.ObjectID)
	outcome.Exchanged = true
	outcome.After = []game.ObjectSnapshot{currentProgramSnapshot(c.game, a), currentProgramSnapshot(c.game, b)}
	return outcome, nil
}
