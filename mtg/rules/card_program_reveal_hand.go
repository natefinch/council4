package rules

import (
	"errors"
	"fmt"

	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// RevealHand has one player reveal their whole hand (CR 701.20a), then the
// source's controller chooses among the revealed cards matching the filter.
// Every card's reveal is a public announcement every player's observation
// keeps (AnnouncedCards), whether or not a choice follows, and the hand stays
// visible to every player until the choice is made; afterwards the hand is
// hidden again (CR 402.3) and the result's handles carry what was revealed.
// The choice request and its public log list the whole revealed hand, not only
// the matching options. Public records show each card's characteristics, never
// its persistent identity, so no other player can track a revealed card into a
// later hidden object such as a face-down foretold card (CR 406.3); the chooser
// picks by option index. Nothing leaves the hand (CR 701.20b): the chosen hand
// handles authorize one later MoveCards placement. A mandatory choice takes
// every matching card unasked when there are no more than Max, and nothing
// when none match (CR 608.2c-d).
func (c *programContext) RevealHand(reveal cardprogram.HandReveal) (cardprogram.HandRevealResult, error) {
	if err := c.valid(); err != nil {
		return cardprogram.HandRevealResult{}, err
	}
	if reveal.Max < 0 {
		return cardprogram.HandRevealResult{}, errors.New("hand reveal Max must not be negative")
	}
	if problems := reveal.Player.Validate(); len(problems) != 0 {
		return cardprogram.HandRevealResult{}, fmt.Errorf("invalid hand reveal player: %s", problems[0])
	}
	if problems := reveal.Filter.Validate(); len(problems) != 0 {
		return cardprogram.HandRevealResult{}, fmt.Errorf("invalid hand reveal filter: %s", problems[0])
	}
	player, available := resolvePlayerReference(c.game, c.obj, reveal.Player)
	if !available || !isPlayerAlive(c.game, player) {
		return cardprogram.HandRevealResult{}, nil
	}
	result := cardprogram.HandRevealResult{Available: true, Player: player}
	endReveal := c.game.OpenHandReveal(player)
	defer endReveal()
	chooser := c.obj.Controller
	var matching []*game.CardInstance
	var revealed []game.ChoiceCardInfo
	for _, cardID := range c.game.Players[player].Hand.All() {
		card, exists := c.game.GetCardInstance(cardID)
		if !exists {
			continue
		}
		emitEvent(c.game, game.Event{
			Kind:          game.EventCardRevealed,
			Announcement:  true,
			Unidentified:  true,
			SourceID:      stackObjectSourceID(c.obj),
			StackObjectID: stackObjectID(c.obj),
			Controller:    stackObjectController(c.obj),
			Player:        player,
			CardID:        cardID,
			FromZone:      zone.Hand,
			Amount:        1,
		})
		if info := revealedCardChoiceInfo(c.game, cardID); info.Exists {
			revealed = append(revealed, info.Val)
		}
		handle := c.mintFoundCard(card, zone.Hand)
		result.Revealed = append(result.Revealed, handle)
		if handCardMatchesSelection(c.game, card, reveal.Filter, chooser) {
			matching = append(matching, card)
			result.Matching = append(result.Matching, handle)
		}
	}
	chosen := c.chooseRevealedHandCards(reveal, chooser, matching, revealed)
	for _, index := range chosen {
		result.Chosen = append(result.Chosen, result.Matching[index])
	}
	return result, nil
}

// chooseRevealedHandCards returns the indices of the matching cards the
// chooser chooses. The hand is revealed to everyone, so the choice, with the
// whole revealed hand, is public.
func (c *programContext) chooseRevealedHandCards(reveal cardprogram.HandReveal, chooser game.PlayerID, matching []*game.CardInstance, revealed []game.ChoiceCardInfo) []int {
	maxChoices := min(reveal.Max, len(matching))
	minChoices := maxChoices
	if reveal.Optional {
		minChoices = 0
	}
	if maxChoices == 0 || !isPlayerAlive(c.game, chooser) {
		return nil
	}
	if minChoices == len(matching) {
		return firstChoiceIndices(minChoices)
	}
	options := make([]game.ChoiceOption, len(matching))
	for index, card := range matching {
		options[index] = game.ChoiceOption{
			Index: index,
			Label: cardChoiceLabel(c.game, card.ID),
			Card:  revealedCardChoiceInfo(c.game, card.ID),
		}
	}
	request := game.ChoiceRequest{
		Kind:             game.ChoiceZoneSelection,
		Player:           chooser,
		Prompt:           "Choose from the revealed hand.",
		Options:          options,
		MinChoices:       minChoices,
		MaxChoices:       maxChoices,
		DefaultSelection: firstChoiceIndices(maxChoices),
		RevealedCards:    revealed,
	}
	return c.engine.chooseChoice(c.game, c.agents, request, c.log)
}

// revealedCardChoiceInfo is a revealed card's public characteristics without
// its persistent CardID, which would let players correlate it with the card
// after it becomes a hidden object (CR 406.3).
func revealedCardChoiceInfo(g *game.Game, cardID id.ID) opt.V[game.ChoiceCardInfo] {
	info := cardChoiceInfo(g, cardID)
	info.Val.CardID = 0
	return info
}
