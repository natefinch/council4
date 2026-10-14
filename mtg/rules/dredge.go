package rules

import (
	"fmt"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/zone"
)

// dredgeCandidate is one Dredge card eligible to replace a draw.
type dredgeCandidate struct {
	cardID id.ID
	count  int
}

// eligibleDredgeCandidates collects the Dredge cards in playerID's graveyard
// whose mill count does not exceed the player's library size (CR 702.52c). A
// player whose library is too small to mill cannot apply that card's dredge.
func eligibleDredgeCandidates(g *game.Game, playerID game.PlayerID) []dredgeCandidate {
	player, ok := playerByID(g, playerID)
	if !ok {
		return nil
	}
	librarySize := player.Library.Size()
	var candidates []dredgeCandidate
	for _, cardID := range player.Graveyard.All() {
		card, ok := g.GetCardInstance(cardID)
		if !ok {
			continue
		}
		count, ok := cardFaceOrDefault(card, game.FaceFront).DredgeCount()
		if !ok || count <= 0 || count > librarySize {
			continue
		}
		candidates = append(candidates, dredgeCandidate{cardID: cardID, count: count})
	}
	return candidates
}

// tryDredge offers playerID the choice to replace an imminent draw with the
// Dredge of one eligible graveyard card. When the player chooses a card it mills
// that card's amount and returns the card to hand, and tryDredge reports true so
// the caller skips the draw. Declining (the default) reports false. The choice
// defaults to declining so deterministic play never auto-dredges.
func (e *Engine) tryDredge(g *game.Game, playerID game.PlayerID, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	candidates := eligibleDredgeCandidates(g, playerID)
	if len(candidates) == 0 {
		return false
	}
	options := make([]game.ChoiceOption, len(candidates))
	for i, candidate := range candidates {
		label := fmt.Sprintf("Dredge %d", candidate.count)
		if card, ok := g.GetCardInstance(candidate.cardID); ok {
			label = fmt.Sprintf("Dredge %d (%s)", candidate.count, card.Def.Name)
		}
		options[i] = game.ChoiceOption{
			Index: i,
			Label: label,
			Card:  cardChoiceInfo(g, candidate.cardID),
		}
	}
	selected := e.chooseChoice(g, agents, game.ChoiceRequest{
		Kind:             game.ChoiceZoneSelection,
		Player:           playerID,
		Prompt:           "Dredge a card from your graveyard instead of drawing?",
		Options:          options,
		MinChoices:       0,
		MaxChoices:       1,
		DefaultSelection: nil,
	}, log)
	if len(selected) == 0 {
		return false
	}
	index := selected[0]
	if index < 0 || index >= len(candidates) {
		return false
	}
	chosen := candidates[index]
	millCards(g, playerID, chosen.count)
	moveCardBetweenZones(g, playerID, chosen.cardID, zone.Graveyard, zone.Hand)
	return true
}
