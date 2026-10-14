package rules

import (
	"errors"
	"slices"

	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/game"
)

func (c *programContext) Player(reference game.PlayerReference) (game.PlayerID, bool, error) {
	if err := c.valid(); err != nil {
		return 0, false, err
	}
	if err := game.ValidateScopedPlayerReference(reference, c.specs); err != nil {
		return 0, false, err
	}
	player, available := resolvePlayerReference(c.game, c.obj, reference)
	return player, available, nil
}

func (c *programContext) Players(reference game.PlayerGroupReference) ([]game.PlayerID, error) {
	if err := c.valid(); err != nil {
		return nil, err
	}
	if problems := reference.Validate(); len(problems) != 0 {
		return nil, errors.New(problems[0])
	}
	members := newReferenceResolver(c.game, c.obj).playerGroup(reference)
	ordered := make([]game.PlayerID, 0, len(members))
	for _, player := range triggerAPNAPPlayers(c.game) {
		if slices.Contains(members, player) {
			ordered = append(ordered, player)
		}
	}
	return ordered, nil
}

func (c *programContext) Copy(operation game.CopyStackObject, player game.PlayerID) (cardprogram.Copies, error) {
	if err := c.valid(); err != nil {
		return cardprogram.Copies{}, err
	}
	if player < 0 || int(player) >= game.NumPlayers || operation.Chooser.Exists {
		return cardprogram.Copies{}, errors.New("authored copies require one explicit valid copier")
	}
	if err := game.ValidatePrimitive(operation, c.specs); err != nil {
		return cardprogram.Copies{}, err
	}
	resolver := newEffectResolver(c.engine, c.game, c.obj, c.agents, c.log)
	original, available := copyStackObjectSource(resolver, operation.Object)
	outcome := cardprogram.Copies{Available: available}
	if !available || !isPlayerAlive(c.game, player) {
		return outcome, nil
	}
	_, copies := copyStackObjectForPlayer(resolver, operation, original, player)
	outcome.StackIDs = slices.Clone(copies)
	return outcome, nil
}
