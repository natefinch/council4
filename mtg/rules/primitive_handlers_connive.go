package rules

import (
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
)

// handleConnive resolves the connive keyword action (CR 702.154): the conniving
// permanent's controller draws prim.Amount cards, then discards prim.Amount
// cards, and a +1/+1 counter is placed on prim.Object for each nonland card
// discarded this way.
func handleConnive(r *effectResolver, prim game.Connive) effectResolved {
	res := effectResolved{accepted: true, amount: r.quantity(prim.Amount)}
	if res.amount <= 0 {
		return res
	}
	playerID, ok := r.resolvePlayer(prim.Player)
	if !ok {
		return res
	}
	r.engine.drawCards(r.game, playerID, res.amount, r.agents, r.log)
	nonland := r.conniveDiscard(playerID, res.amount)
	if nonland > 0 {
		if permanent, found := r.resolveObject(prim.Object); found {
			addCountersToPermanentControlledBy(r.game, stackObjectController(r.obj), permanent, counter.PlusOnePlusOne, nonland)
		}
	}
	res.succeeded = true
	return res
}

// conniveDiscard makes playerID discard amount chosen cards from hand and
// returns how many of the discarded cards were nonland cards. It mirrors the
// ordinary resolution discard but reports the nonland count the connive action
// needs to place +1/+1 counters (CR 702.154).
func (r *effectResolver) conniveDiscard(playerID game.PlayerID, amount int) int {
	player, ok := playerByID(r.game, playerID)
	if !ok {
		return 0
	}
	candidates := player.Hand.All()
	amount = min(amount, len(candidates))
	if amount <= 0 {
		return 0
	}
	options := make([]game.ChoiceOption, len(candidates))
	for i, cardID := range candidates {
		options[i] = game.ChoiceOption{
			Index: i,
			Label: cardChoiceLabel(r.game, cardID),
			Card:  cardChoiceInfo(r.game, cardID),
		}
	}
	selected := r.engine.chooseChoice(r.game, r.agents, game.ChoiceRequest{
		Kind:             game.ChoiceResolution,
		Player:           playerID,
		Prompt:           "Choose cards to discard",
		Options:          options,
		MinChoices:       amount,
		MaxChoices:       amount,
		DefaultSelection: firstChoiceIndices(amount),
	}, r.log)
	simultaneousID := r.game.IDGen.Next()
	nonland := 0
	for _, idx := range selected {
		if idx < 0 || idx >= len(candidates) {
			continue
		}
		cardID := candidates[idx]
		wasNonland := conniveCardIsNonland(r.game, cardID)
		if discardCardFromHandInBatch(r.game, playerID, cardID, simultaneousID) && wasNonland {
			nonland++
		}
	}
	return nonland
}

// conniveCardIsNonland reports whether the named card is a nonland card, read
// from its front face before it leaves the hand.
func conniveCardIsNonland(g *game.Game, cardID id.ID) bool {
	card, ok := g.GetCardInstance(cardID)
	if !ok {
		return false
	}
	return !slices.Contains(cardFaceOrDefault(card, game.FaceFront).Types, types.Land)
}
