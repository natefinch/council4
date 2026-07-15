package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/zone"
)

// performOpeningHandBattlefieldActions runs the pregame opening-hand action
// window (CR 103.6). Once the mulligan process is complete, the starting player,
// then each other player in turn order, may begin the game with eligible cards
// from their opening hand on the battlefield (CR 103.6a) — the Leyline cycle's
// "If this card is in your opening hand, you may begin the game with it on the
// battlefield."
//
// It runs after opening hands and any mulligans and before the first turn, while
// no player has priority and nothing is on the stack. Accepted cards are put onto
// the battlefield as new permanents so their static abilities and replacement
// effects apply from the start of the game. It is called before
// markCurrentTurnEventStart advances the trigger cursor (see events.go), so the
// entering events are positioned before that cursor and never place an
// enters-the-battlefield triggered ability on the stack — matching how the same
// mechanism excludes opening-hand draws from triggering, and matching the pregame
// procedure in which no ability can be put on the stack until the first turn
// begins.
func (e *Engine) performOpeningHandBattlefieldActions(g *game.Game, agents [game.NumPlayers]PlayerAgent) {
	for _, playerID := range openingHandActionOrder(g) {
		e.performPlayerOpeningHandBattlefield(g, agents, playerID)
	}
}

// openingHandActionOrder lists the players who may take opening-hand actions in
// the CR 103.6 order: the starting player (the initial active player) first, then
// each other player in turn order. It walks the fixed seating order once from the
// starting player's seat, so each seat is visited exactly once; eliminated seats
// are skipped, so a goldfish game processes only its single active seat.
func openingHandActionOrder(g *game.Game) []game.PlayerID {
	seats := g.TurnOrder.Order
	start := 0
	for i, p := range seats {
		if p == g.Turn.ActivePlayer {
			start = i
			break
		}
	}
	order := make([]game.PlayerID, 0, len(seats))
	for i := range len(seats) {
		p := seats[(start+i)%len(seats)]
		if !g.TurnOrder.IsEliminated(p) {
			order = append(order, p)
		}
	}
	return order
}

// performPlayerOpeningHandBattlefield asks one player about each eligible card in
// their opening hand. The hand is snapshotted first: accepting a card mutates the
// hand, and a card added by a move must not itself be offered. Duplicate copies
// are distinct instances and are each considered once, in hand order, so the
// window is deterministic (CR 103.6a lets the player order these actions freely;
// hand order is a fixed, reproducible resolution of that freedom).
func (e *Engine) performPlayerOpeningHandBattlefield(g *game.Game, agents [game.NumPlayers]PlayerAgent, playerID game.PlayerID) {
	player, ok := playerByID(g, playerID)
	if !ok || player.Eliminated {
		return
	}
	for _, cardID := range append([]id.ID(nil), player.Hand.All()...) {
		e.considerOpeningHandBattlefieldCard(g, agents, playerID, cardID)
	}
}

// considerOpeningHandBattlefieldCard offers a single opening-hand card to its
// owner and, on acceptance, moves it from hand to the battlefield. It leaves the
// card untouched — with no partial mutation — when the card is not eligible, has
// left the hand between enumeration and resolution, the player declines or no
// valid decision is made, or the entry itself does not happen.
func (e *Engine) considerOpeningHandBattlefieldCard(g *game.Game, agents [game.NumPlayers]PlayerAgent, playerID game.PlayerID, cardID id.ID) {
	player, ok := playerByID(g, playerID)
	if !ok {
		return
	}
	// Only cards still in this player's opening hand are eligible; a card may have
	// left the hand since the snapshot was taken.
	if !player.Hand.Contains(cardID) {
		return
	}
	card, ok := g.GetCardInstance(cardID)
	if !ok || card.Def == nil || !card.Def.BeginsGameOnBattlefield() {
		return
	}
	if !e.chooseBeginGameOnBattlefield(g, agents, playerID, cardID) {
		return
	}
	// Re-validate before mutating. Agents observe read-only state, so the card
	// cannot have moved, but this guard guarantees no partial state change.
	if !player.Hand.Contains(cardID) {
		return
	}
	// A card that begins the game on the battlefield is put there by its owner,
	// who becomes its controller (CR 110.2). Remove it from hand first, then build
	// the permanent through the shared helper so ownership, controller,
	// enters-the-battlefield replacements, counters, static abilities, events, and
	// logs stay coherent. Commander invariants are preserved trivially: a commander
	// begins in the command zone (CR 903.6), never in the opening hand, so this
	// path never touches it.
	if !removeCardFromZone(g, card.Owner, cardID, zone.Hand) {
		return
	}
	if _, placed := createCardPermanentFaceWithChoices(e, g, card, playerID, zone.Hand, game.FaceFront, agents, nil); !placed {
		// The entry did not happen (for example an entry from hand was
		// prohibited). Return the card to its owner's hand so no card is lost and
		// no partial state remains.
		player.Hand.Add(cardID)
	}
}

// chooseBeginGameOnBattlefield asks playerID whether to begin the game with
// cardID on the battlefield. The action is optional (CR 103.6, "may"), so the
// deterministic fallback declines: with no agent, no ChoiceAgent, or an invalid
// answer, the card stays hidden in hand rather than being revealed. The request
// carries only the acting player's own card and is sent only to that player, and
// it is not recorded in a shared turn log, so a declined card never leaks an
// opponent's hidden hand contents through the game's public logs; an accepted
// card is revealed only by entering the battlefield.
func (e *Engine) chooseBeginGameOnBattlefield(g *game.Game, agents [game.NumPlayers]PlayerAgent, playerID game.PlayerID, cardID id.ID) bool {
	request := mayChoiceRequest(playerID, "You may begin the game with this card on the battlefield.")
	// Decline by default: the action is optional (CR 103.6), so with no agent, no
	// ChoiceAgent, or an invalid answer the card stays hidden in hand rather than
	// being revealed. This overrides mayChoiceRequest's accept-by-default, which
	// exists to preserve legacy behavior for optional triggered abilities.
	request.DefaultSelection = []int{0}
	request.Subject = cardChoiceInfo(g, cardID)
	selected := e.chooseChoice(g, agents, request, nil)
	return len(selected) == 1 && selected[0] == 1
}
