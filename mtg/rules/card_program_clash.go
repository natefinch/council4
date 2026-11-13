package rules

import (
	"fmt"
	"slices"

	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Clash makes the source's controller clash with an opponent (CR 701.30b):
// both players reveal their top cards at the same time, then in APNAP order
// each decides whether to put that card on the bottom of their library, then
// the cards move at the same time (CR 701.30c). An empty library reveals
// nothing and is never a draw. The winner revealed a card whose mana value is
// greater than every other revealed card (CR 701.30d).
func (c *programContext) Clash(opponent opt.V[game.PlayerReference]) (cardprogram.ClashResult, error) {
	if err := c.valid(); err != nil {
		return cardprogram.ClashResult{}, err
	}
	controller := c.source.Controller
	if !isPlayerAlive(c.game, controller) {
		return cardprogram.ClashResult{}, nil
	}
	chosen, ok, err := c.clashOpponent(controller, opponent)
	if err != nil || !ok {
		return cardprogram.ClashResult{}, err
	}
	result := cardprogram.ClashResult{Available: true, Opponent: chosen}
	revealed := map[game.PlayerID]id.ID{}
	for _, player := range triggerAPNAPPlayers(c.game) {
		if player != controller && player != chosen {
			continue
		}
		reveal := cardprogram.ClashReveal{Player: player}
		if cards := revealCardIDs(c.game, c.obj, player, zone.Library, 1); len(cards) == 1 {
			reveal.Card = clashCardInfo(c.game, cards[0])
			revealed[player] = cards[0]
		}
		result.Reveals = append(result.Reveals, reveal)
	}
	// Both cards are revealed before anyone chooses, so every chooser sees
	// every revealed card along with earlier placements.
	var decisions []game.ClashDecision
	for _, reveal := range result.Reveals {
		if reveal.Card.Exists {
			decisions = append(decisions, game.ClashDecision{Player: reveal.Player, Card: reveal.Card.Val})
		}
	}
	decision := 0
	for i := range result.Reveals {
		reveal := &result.Reveals[i]
		if !reveal.Card.Exists {
			continue
		}
		request := libraryChoiceRequest(game.ChoiceResolution, reveal.Player,
			"Clash: choose where to put the revealed card.", []string{"top", "bottom"})
		request.Subject = reveal.Card
		request.ClashDecisions = slices.Clone(decisions)
		selected := c.engine.chooseChoice(c.game, c.agents, request, c.log)
		reveal.Bottom = len(selected) == 1 && selected[0] == 1
		decisions[decision].Decided = true
		decisions[decision].Bottom = reveal.Bottom
		decision++
	}
	c.game.RecordClashReveals(decisions)
	for _, reveal := range result.Reveals {
		if reveal.Bottom {
			commitScryPlacement(c.game, scryPlacement{player: reveal.Player, bottom: []id.ID{revealed[reveal.Player]}})
		}
	}
	result.Winner = clashWinner(result.Reveals)
	emitClashEvents(c.game, result)
	return result, nil
}

// emitClashEvents records that each participant clashed, after the whole
// clash has ended (CR 701.30b-d), so "whenever you clash" abilities of both
// players trigger with the won/lost fact captured on each event.
func emitClashEvents(g *game.Game, result cardprogram.ClashResult) {
	simultaneousID := g.IDGen.Next()
	for _, reveal := range result.Reveals {
		opponent := result.Opponent
		for _, other := range result.Reveals {
			if other.Player != reveal.Player {
				opponent = other.Player
			}
		}
		emitEvent(g, game.Event{
			Kind:           game.EventClashed,
			Controller:     reveal.Player,
			Player:         reveal.Player,
			ClashOpponent:  opponent,
			ClashWon:       result.Won(reveal.Player),
			SimultaneousID: simultaneousID,
		})
	}
}

func (c *programContext) clashOpponent(controller game.PlayerID, reference opt.V[game.PlayerReference]) (game.PlayerID, bool, error) {
	if reference.Exists {
		player, available, err := c.Player(reference.Val)
		if err != nil || !available || player == controller || !isPlayerAlive(c.game, player) {
			return 0, false, err
		}
		return player, true, nil
	}
	opponents := aliveOpponents(c.game, controller)
	switch len(opponents) {
	case 0:
		return 0, false, nil
	case 1:
		return opponents[0], true, nil
	}
	options := make([]game.ChoiceOption, 0, len(opponents))
	for i, opponent := range opponents {
		options = append(options, game.ChoiceOption{
			Index:   i,
			Label:   fmt.Sprintf("Player %d", opponent+1),
			Targets: []game.Target{game.PlayerTarget(opponent)},
		})
	}
	selected := c.engine.chooseChoice(c.game, c.agents, game.ChoiceRequest{
		Kind:             game.ChoicePlayer,
		Player:           controller,
		Prompt:           "Clash: choose an opponent.",
		Options:          options,
		MinChoices:       1,
		MaxChoices:       1,
		DefaultSelection: []int{0},
	}, c.log)
	if len(selected) != 1 || selected[0] < 0 || selected[0] >= len(opponents) {
		return opponents[0], true, nil
	}
	return opponents[selected[0]], true, nil
}

// clashCardInfo reports a revealed library card's public characteristics with
// its mana value in the library (CR 709.4b).
func clashCardInfo(g *game.Game, cardID id.ID) opt.V[game.ChoiceCardInfo] {
	info := cardChoiceInfo(g, cardID)
	if card, ok := g.GetCardInstance(cardID); ok && info.Exists {
		info.Val.ManaValue = offStackCardManaValue(card)
	}
	return info
}

func clashWinner(reveals []cardprogram.ClashReveal) opt.V[game.PlayerID] {
	var winner opt.V[game.PlayerID]
	for i, reveal := range reveals {
		if !reveal.Card.Exists {
			continue
		}
		highest := true
		for j, other := range reveals {
			if j != i && other.Card.Exists && other.Card.Val.ManaValue >= reveal.Card.Val.ManaValue {
				highest = false
				break
			}
		}
		if highest {
			winner = opt.Val(reveal.Player)
		}
	}
	return winner
}
