package rules

import (
	"errors"
	"fmt"
	"slices"

	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// RevealHand has one player reveal their whole hand (CR 701.20a), then the
// source's controller chooses among the revealed cards matching the filter,
// and, when the reveal offers it, that player's graveyard cards matching its
// own filter, in one choice ("a nonland card from it or a card from their
// graveyard"). Every card's reveal is a public announcement every player's
// observation keeps (AnnouncedCards), whether or not a choice follows, and the
// hand stays visible to every player until the choice is made; afterwards the
// hand is hidden again (CR 402.3) and the result's handles carry what was
// revealed. The choice request and its public log list the whole revealed
// hand, not only the matching options. Public records show each revealed
// card's characteristics, never its persistent identity, so no other player
// can track a revealed card into a later hidden object such as a face-down
// foretold card (CR 406.3); the chooser picks by option index. Graveyard
// options are public cards (CR 400.2) and keep their identities. Nothing moves (CR
// 701.20b): the chosen handles authorize one later MoveCards placement, which
// moves hand and graveyard cards together. A mandatory choice takes every
// matching card unasked when there are no more than Max, and nothing when none
// match (CR 608.2c-d); All takes every match unasked. Chooser, when set, names
// another player who makes the choice ("That player exiles a card from it");
// the filter stays relative to the source's controller. KeepRevealed keeps the
// hand revealed until the program finishes or a later reveal finds it empty,
// rather than only through this choice, and every later RevealHand of that hand
// in the same program continues that one reveal (CR 701.20a): it shows the
// hand as it now is and announces only cards it has not already revealed, so
// alternating choices ("Repeat this process until all cards in that hand have
// been exiled") are made from one revealed hand. Ending at an empty hand keeps
// cards the program later returns to it hidden (CR 402.3).
func (c *programContext) RevealHand(reveal cardprogram.HandReveal) (cardprogram.HandRevealResult, error) {
	if err := c.valid(); err != nil {
		return cardprogram.HandRevealResult{}, err
	}
	if reveal.Max < 0 {
		return cardprogram.HandRevealResult{}, errors.New("hand reveal Max must not be negative")
	}
	if reveal.All && (reveal.Max != 0 || reveal.Optional) {
		return cardprogram.HandRevealResult{}, errors.New("hand reveal All excludes Max and Optional")
	}
	if problems := reveal.Player.Validate(); len(problems) != 0 {
		return cardprogram.HandRevealResult{}, fmt.Errorf("invalid hand reveal player: %s", problems[0])
	}
	if problems := reveal.Filter.Validate(); len(problems) != 0 {
		return cardprogram.HandRevealResult{}, fmt.Errorf("invalid hand reveal filter: %s", problems[0])
	}
	if reveal.Graveyard.Exists {
		if problems := reveal.Graveyard.Val.Validate(); len(problems) != 0 {
			return cardprogram.HandRevealResult{}, fmt.Errorf("invalid hand reveal graveyard filter: %s", problems[0])
		}
	}
	if reveal.Chooser.Exists {
		if problems := reveal.Chooser.Val.Validate(); len(problems) != 0 {
			return cardprogram.HandRevealResult{}, fmt.Errorf("invalid hand reveal chooser: %s", problems[0])
		}
	}
	player, available := resolvePlayerReference(c.game, c.obj, reveal.Player)
	if !available || !isPlayerAlive(c.game, player) {
		return cardprogram.HandRevealResult{}, nil
	}
	result := cardprogram.HandRevealResult{Available: true, Player: player}
	controller := c.obj.Controller
	chooser, hasChooser := controller, true
	if reveal.Chooser.Exists {
		chooser, hasChooser = resolvePlayerReference(c.game, c.obj, reveal.Chooser.Val)
	}
	kept := c.keptHandReveal(player)
	if kept == nil {
		endReveal := c.game.OpenHandReveal(player)
		if reveal.KeepRevealed {
			kept = &keptHandReveal{player: player, announced: make(map[game.ZoneCardSnapshot]bool), end: endReveal}
			c.keptReveals = append(c.keptReveals, kept)
		} else {
			defer endReveal()
		}
	}
	var candidates []revealChoiceCandidate
	var revealed []game.ChoiceCardInfo
	for _, cardID := range c.game.Players[player].Hand.All() {
		card, exists := c.game.GetCardInstance(cardID)
		if !exists {
			continue
		}
		object := game.ZoneCardSnapshot{CardID: cardID, ZoneVersion: card.ZoneVersion}
		if kept == nil || !kept.announced[object] {
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
			if kept != nil {
				kept.announced[object] = true
			}
		}
		if info := revealedCardChoiceInfo(c.game, cardID); info.Exists {
			revealed = append(revealed, info.Val)
		}
		handle := c.mintFoundCard(card, zone.Hand)
		result.Revealed = append(result.Revealed, handle)
		if handCardMatchesSelection(c.game, card, reveal.Filter, controller) {
			result.Matching = append(result.Matching, handle)
			candidates = append(candidates, revealChoiceCandidate{handle: handle, info: revealedCardChoiceInfo(c.game, cardID)})
		}
	}
	if reveal.Graveyard.Exists {
		for _, cardID := range c.game.Players[player].Graveyard.All() {
			card, exists := c.game.GetCardInstance(cardID)
			if !exists || !handCardMatchesSelection(c.game, card, reveal.Graveyard.Val, controller) {
				continue
			}
			handle := c.mintFoundCard(card, zone.Graveyard)
			result.GraveyardMatching = append(result.GraveyardMatching, handle)
			candidates = append(candidates, revealChoiceCandidate{handle: handle, info: cardChoiceInfo(c.game, cardID)})
		}
	}
	if kept != nil && len(result.Revealed) == 0 {
		c.endKeptHandReveal(kept)
	}
	if !hasChooser {
		return result, nil
	}
	for _, index := range c.chooseRevealedCards(reveal, chooser, candidates, revealed) {
		result.Chosen = append(result.Chosen, candidates[index].handle)
	}
	return result, nil
}

// keptHandReveal is one hand a program keeps revealed until it finishes
// (HandReveal.KeepRevealed), with the hand objects that reveal has announced.
type keptHandReveal struct {
	player    game.PlayerID
	announced map[game.ZoneCardSnapshot]bool
	end       func()
}

func (c *programContext) keptHandReveal(player game.PlayerID) *keptHandReveal {
	for _, kept := range c.keptReveals {
		if kept.player == player {
			return kept
		}
	}
	return nil
}

// endKeptHandReveal ends a kept reveal once it finds the hand empty, so cards
// later put into that hand, which were never revealed, stay hidden (CR 402.3).
func (c *programContext) endKeptHandReveal(kept *keptHandReveal) {
	kept.end()
	c.keptReveals = slices.DeleteFunc(c.keptReveals, func(other *keptHandReveal) bool { return other == kept })
}

// endKeptHandReveals hides every hand the program still keeps revealed (CR
// 402.3).
func (c *programContext) endKeptHandReveals() {
	for _, kept := range c.keptReveals {
		kept.end()
	}
	c.keptReveals = nil
}

// revealChoiceCandidate is one card the controller may choose after a hand
// reveal, with the public characteristics its choice option shows.
type revealChoiceCandidate struct {
	handle cardprogram.FoundCard
	info   opt.V[game.ChoiceCardInfo]
}

// chooseRevealedCards returns the indices of the candidates the chooser
// chooses. The hand is revealed to everyone, so the choice, with the whole
// revealed hand, is public.
func (c *programContext) chooseRevealedCards(reveal cardprogram.HandReveal, chooser game.PlayerID, candidates []revealChoiceCandidate, revealed []game.ChoiceCardInfo) []int {
	if reveal.All {
		return firstChoiceIndices(len(candidates))
	}
	maxChoices := min(reveal.Max, len(candidates))
	minChoices := maxChoices
	if reveal.Optional {
		minChoices = 0
	}
	if maxChoices == 0 || !isPlayerAlive(c.game, chooser) {
		return nil
	}
	if minChoices == len(candidates) {
		return firstChoiceIndices(minChoices)
	}
	prompt := "Choose from the revealed hand."
	if reveal.Graveyard.Exists {
		prompt = "Choose from the revealed hand or the graveyard."
	}
	options := make([]game.ChoiceOption, len(candidates))
	for index, candidate := range candidates {
		label := cardChoiceLabel(c.game, candidate.handle.Identity.CardID)
		if reveal.Graveyard.Exists {
			label += " (" + candidate.handle.Zone.String() + ")"
		}
		options[index] = game.ChoiceOption{Index: index, Label: label, Card: candidate.info}
	}
	request := game.ChoiceRequest{
		Kind:             game.ChoiceZoneSelection,
		Player:           chooser,
		Prompt:           prompt,
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
