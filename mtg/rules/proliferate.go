package rules

import (
	"cmp"
	"fmt"
	"slices"
	"strings"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
)

type proliferateRecipient struct {
	permanent *game.Permanent
	player    *game.Player
	kinds     []counter.Kind
}

func (e *Engine) resolveProliferate(g *game.Game, obj *game.StackObject, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	recipients := proliferateRecipients(g)
	if len(recipients) == 0 {
		return false
	}
	controller := stackObjectController(obj)
	selected := slices.Clone(e.chooseChoice(g, agents, proliferateChoiceRequest(g, controller, recipients), log))
	if len(selected) == 0 {
		return false
	}
	// The selection is a set, not a placement order. When there are
	// simultaneous replacement choices, the affected players choose in APNAP
	// order (CR 616.1); use option order within a seat for determinism.
	var rank [game.NumPlayers]int
	for index, player := range triggerAPNAPPlayers(g) {
		rank[player] = index
	}
	recipientRank := func(index int) int {
		recipient := recipients[index]
		if recipient.permanent != nil {
			return rank[effectiveController(g, recipient.permanent)]
		}
		return rank[recipient.player.ID]
	}
	slices.SortFunc(selected, func(a, b int) int {
		if order := cmp.Compare(recipientRank(a), recipientRank(b)); order != 0 {
			return order
		}
		return cmp.Compare(a, b)
	})

	// CR 701.34a: choose all recipients first, then place the counters together.
	// Resolve replacements against the pre-placement state and publish facts
	// only after every counter has been placed.
	type placement struct {
		recipient proliferateRecipient
		kind      counter.Kind
		amount    int
	}
	var placements []placement
	for _, index := range selected {
		recipient := recipients[index]
		for _, kind := range recipient.kinds {
			var amount int
			if recipient.permanent != nil {
				amount = replacementPermanentCounterPlacementAmount(g, controller, recipient.permanent, kind, 1)
			} else {
				amount = replacementPlayerCounterPlacementAmount(g, controller, recipient.player.ID, kind, 1)
			}
			if amount > 0 {
				placements = append(placements, placement{recipient, kind, amount})
			}
		}
	}
	if len(placements) == 0 {
		return false
	}
	simultaneousID := g.IDGen.Next()
	events := make([]game.Event, 0, len(placements))
	for _, p := range placements {
		if p.recipient.permanent != nil {
			events = append(events, placePermanentCounters(g, p.recipient.permanent, controller, p.kind, p.amount, simultaneousID))
		} else if event, ok := placePlayerCounters(p.recipient.player, controller, p.kind, p.amount, simultaneousID); ok {
			events = append(events, event)
		}
	}
	for _, event := range events {
		emitEvent(g, event)
	}
	return len(events) > 0
}

func proliferateRecipients(g *game.Game) []proliferateRecipient {
	var recipients []proliferateRecipient
	for _, permanent := range g.Battlefield {
		if !activeBattlefieldPermanent(permanent) || permanent.Counters.IsEmpty() {
			continue
		}
		kinds := sortedCounterKinds(permanent.Counters.All())
		if len(kinds) > 0 {
			recipients = append(recipients, proliferateRecipient{permanent: permanent, kinds: kinds})
		}
	}
	for _, player := range g.Players {
		if player == nil || player.Eliminated {
			continue
		}
		var counters []counter.Kind
		if player.PoisonCounters > 0 {
			counters = append(counters, counter.Poison)
		}
		if player.EnergyCounters > 0 {
			counters = append(counters, counter.Energy)
		}
		if player.ExperienceCounters > 0 {
			counters = append(counters, counter.Experience)
		}
		if len(counters) > 0 {
			recipients = append(recipients, proliferateRecipient{player: player, kinds: counters})
		}
	}
	return recipients
}

func sortedCounterKinds(counts map[counter.Kind]int) []counter.Kind {
	kinds := make([]counter.Kind, 0, len(counts))
	for kind, amount := range counts {
		if amount > 0 {
			kinds = append(kinds, kind)
		}
	}
	slices.Sort(kinds)
	return kinds
}

func proliferateChoiceRequest(g *game.Game, player game.PlayerID, recipients []proliferateRecipient) game.ChoiceRequest {
	options := make([]game.ChoiceOption, len(recipients))
	for i, recipient := range recipients {
		option := game.ChoiceOption{Index: i}
		if p := recipient.permanent; p != nil {
			option.Card = permanentChoiceInfo(g, p)
			option.Label = fmt.Sprintf("Permanent %d", p.ObjectID)
			if option.Card.Exists && option.Card.Val.Name != "" {
				option.Label = fmt.Sprintf("%s (%s)", option.Card.Val.Name, option.Label)
			}
		} else {
			option.Label = fmt.Sprintf("Player %d", recipient.player.ID)
			if recipient.player.Name != "" {
				option.Label = fmt.Sprintf("%s (%s)", recipient.player.Name, option.Label)
			}
		}
		names := make([]string, len(recipient.kinds))
		for j, kind := range recipient.kinds {
			names[j] = kind.String()
		}
		option.Label += ": " + strings.Join(names, ", ")
		options[i] = option
	}
	return game.ChoiceRequest{
		Kind:       game.ChoiceProliferate,
		Player:     player,
		Prompt:     "Proliferate: choose any number of permanents and/or players with counters.",
		Options:    options,
		MinChoices: 0,
		MaxChoices: len(options),
	}
}
