package rules

import (
	"errors"
	"fmt"

	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/game"
)

// DelayCardTrigger schedules a fixed-timing delayed trigger bound to one card
// handle this invocation minted. The trigger captures the handle's exact zone
// incarnation, so its content's CardReferenceCaptured finds nothing once that
// card has changed zones (CR 400.7, 603.7c). The handle need not still be
// available: a delayed trigger still triggers after its object left, and then
// does nothing.
func (c *programContext) DelayCardTrigger(request cardprogram.CardDelayedTrigger) (bool, error) {
	if err := c.valid(); err != nil {
		return false, err
	}
	handle := request.Card
	minted, available := c.found[handle.Identity.CardID]
	if !available || minted != handle.Identity || c.libraryEpoch[handle.Identity.CardID] != handle.LibraryEpoch {
		return false, errors.New("card delayed trigger requires a handle from this invocation")
	}
	def := request.Trigger
	if def.EventPattern.Exists || def.CapturedCard.Exists || def.CapturedObject.Exists || def.CapturedObjectGroup.Exists {
		return false, errors.New("card delayed trigger requires a fixed timing bound only to its card handle")
	}
	// The handle supplies the captured card that the intervening "if you
	// haven't cast the card" requires.
	check := def
	check.InterveningIfCapturedCardNotCast = false
	primitive := game.CreateDelayedTrigger{Trigger: check}
	if err := game.ValidatePrimitive(primitive, c.specs); err != nil {
		return false, fmt.Errorf("invalid card delayed trigger: %w", err)
	}
	if err := game.ValidateLinkedProducerOperation(primitive, c.obj.ResolvingLinkedProducers); err != nil {
		return false, err
	}
	if err := c.engine.validateProgramReferences(primitive.ProgramReferences()); err != nil {
		return false, err
	}
	if !scheduleDelayedTrigger(c.game, c.obj, &def) {
		return false, nil
	}
	scheduled := &c.game.DelayedTriggers[len(c.game.DelayedTriggers)-1]
	scheduled.CapturedCardID = handle.Identity.CardID
	scheduled.CapturedCardZoneVersion = handle.Identity.ZoneVersion
	return true, nil
}
