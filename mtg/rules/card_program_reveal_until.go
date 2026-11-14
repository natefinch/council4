package rules

import (
	"errors"
	"fmt"

	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// RevealUntil runs the shared reveal-until-to-library-bottom process for one
// player. The revealed cards were reordered, so each became a new object (CR
// 701.20d): the returned handles report what was revealed, and every handle to
// those cards in this invocation stops authorizing placements.
func (c *programContext) RevealUntil(prim game.RevealUntil) (cardprogram.UntilResult, error) {
	if err := c.valid(); err != nil {
		return cardprogram.UntilResult{}, err
	}
	if prim.Destination != zone.Library || prim.PlayerGroup.Kind != game.PlayerGroupReferenceNone {
		return cardprogram.UntilResult{}, errors.New("authored RevealUntil requires one Player and a Library Destination")
	}
	if err := game.ValidatePrimitive(prim, c.specs); err != nil {
		return cardprogram.UntilResult{}, fmt.Errorf("invalid authored operation: %w", err)
	}
	player, available := resolvePlayerReference(c.game, c.obj, prim.Player)
	if !available || !isPlayerAlive(c.game, player) {
		return cardprogram.UntilResult{}, nil
	}
	result := cardprogram.UntilResult{Available: true}
	describe := func(cardID game.ObjectID) cardprogram.FoundCard {
		card, _ := c.game.GetCardInstance(cardID)
		c.revokeFoundCard(cardID)
		return c.describeFoundCard(card)
	}
	hit, others := revealUntilToLibraryBottom(c.game, c.obj, player, prim.Until)
	if hit.Exists {
		result.Hit = opt.Val(describe(hit.Val))
	}
	for _, cardID := range others {
		result.Others = append(result.Others, describe(cardID))
	}
	return result, nil
}
