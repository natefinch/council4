package agent

import (
	"slices"

	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/rules"
)

// Mana-sequencing weights for GenericStrategy (see
// docs/research/COMMANDER-AGENT-PLAYBOOK.md §5).
const (
	// scoreColorFix rewards a land that adds a colour the agent's hand needs but
	// cannot yet produce, so the agent fixes its mana before adding another
	// source of a colour it already has. It is applied once per newly enabled
	// colour, so a dual that fixes two needs outranks a basic that fixes one.
	scoreColorFix = 60.0

	// scoreHoldUp discourages tapping out on a sorcery-speed play when the agent
	// holds a cheaper instant it wants to keep mana open for. It is smaller than
	// a creature's deployment value, so a high-impact play still overrides it and
	// the agent never simply refuses to develop.
	scoreHoldUp = 30.0
)

// manaAvailability summarises the mana the agent could produce right now from
// its untapped sources: a coarse total and the set of colours available.
type manaAvailability struct {
	total  int
	colors map[color.Color]bool
}

func (m manaAvailability) produces(c color.Color) bool {
	return m.colors[c]
}

// availableMana counts the agent's untapped mana sources and the colours they
// can produce. Each source contributes one mana, a deliberate approximation:
// the heuristic cares about roughly how much mana is open, not exact ramp.
// Summoning-sick creatures are excluded because they cannot tap for mana yet.
func availableMana(obs rules.PlayerObservation) manaAvailability {
	avail := manaAvailability{colors: make(map[color.Color]bool)}
	battlefield := obs.Battlefield()
	for i := range battlefield {
		permanent := battlefield[i]
		if permanent.Controller != obs.Player || !permanent.ProducesMana {
			continue
		}
		if permanent.Tapped || permanent.PhasedOut {
			continue
		}
		if permanent.SummoningSick && isCreaturePermanent(permanent) {
			continue
		}
		avail.total++
		for _, c := range permanent.ProducesColors {
			avail.colors[c] = true
		}
	}
	return avail
}

// scoreLandPlay scores a land drop, rewarding a land that fixes a colour the
// agent's hand needs but cannot yet produce, so it avoids colour screw.
func scoreLandPlay(obs rules.PlayerObservation, act action.Action) float64 {
	score := scorePlayLand
	play, ok := act.PlayLandPayload()
	if !ok {
		return score
	}
	land, ok := handCard(obs, play.CardID)
	if !ok || len(land.ProducesColors) == 0 {
		return score
	}
	needs := handColorNeeds(obs)
	if len(needs) == 0 {
		return score
	}
	have := availableMana(obs)
	for _, c := range land.ProducesColors {
		if needs[c] && !have.produces(c) {
			score += scoreColorFix
		}
	}
	return score
}

// handColorNeeds is the set of colours the agent's non-land hand cards require
// to cast, approximated by their printed colours.
func handColorNeeds(obs rules.PlayerObservation) map[color.Color]bool {
	needs := make(map[color.Color]bool)
	hand := obs.Hand()
	for i := range hand {
		if slices.Contains(hand[i].Types, types.Land) {
			continue
		}
		for _, c := range hand[i].Colors {
			needs[c] = true
		}
	}
	return needs
}

// holdUpPenalty returns the penalty for a sorcery-speed cast that would tap the
// agent below the mana it needs to keep open for its cheapest held instant, so
// it tends to leave interaction mana up. Casting the reactive spell itself is
// never penalised, and the penalty does not apply when the agent could not hold
// the instant up regardless of this play.
func holdUpPenalty(obs rules.PlayerObservation, card rules.CardView, personality Personality) float64 {
	if slices.Contains(card.Types, types.Instant) {
		return 0
	}
	reserve, ok := cheapestHeldInstantCost(obs)
	if !ok {
		return 0
	}
	avail := availableMana(obs).total
	if avail < reserve {
		return 0
	}
	if avail-card.ManaValue < reserve {
		return scoreHoldUp * personality.holdUpScale()
	}
	return 0
}

// cheapestHeldInstantCost returns the lowest mana value among the instants in
// the agent's hand, reporting false when it holds none.
func cheapestHeldInstantCost(obs rules.PlayerObservation) (int, bool) {
	best := 0
	found := false
	hand := obs.Hand()
	for i := range hand {
		if !slices.Contains(hand[i].Types, types.Instant) {
			continue
		}
		if !found || hand[i].ManaValue < best {
			best = hand[i].ManaValue
			found = true
		}
	}
	return best, found
}
