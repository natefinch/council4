package eval

import (
	"strconv"
	"strings"

	"github.com/natefinch/council4/mtg/game/cost"
)

// Describe renders a short, human-readable gloss of what an ability costs and
// does, derived from the value-oriented IR rather than the engine primitives —
// for example "sacrifice a creature, draw a card". It is a coarse summary for
// logs and reports, not a substitute for the printed oracle text: consequences
// the IR does not model contribute nothing, so an ability with only unmodeled
// effects yields an empty string. Phrases are ordered costs first, then effects,
// matching how an ability reads.
func Describe(ability ScorableAbility) string {
	phrases := make([]string, 0, len(ability.Costs)+len(ability.Effect))
	for i := range ability.Costs {
		if phrase := describeCost(ability.Costs[i]); phrase != "" {
			phrases = append(phrases, phrase)
		}
	}
	for i := range ability.Effect {
		if phrase := describeEffect(ability.Effect[i]); phrase != "" {
			phrases = append(phrases, phrase)
		}
	}
	return strings.Join(phrases, ", ")
}

// describeEffect glosses one effect atom. It returns "" for an unmodeled
// (neutral) consequence so Describe omits it rather than emitting filler.
func describeEffect(atom EffectAtom) string {
	switch atom.Kind {
	case EffectCardsDrawn:
		return audience(atom.Affected, "draw", "draws") + " " + countNoun(atom, "card")
	case EffectCardsLost:
		return audience(atom.Affected, "lose", "loses") + " " + countNoun(atom, "card")
	case EffectLifeGained:
		return audience(atom.Affected, "gain", "gains") + " " + amountNoun(atom, "life", "life")
	case EffectLifeLost:
		return audience(atom.Affected, "lose", "loses") + " " + amountNoun(atom, "life", "life")
	case EffectDamageDealt:
		return "deal " + amountNoun(atom, "damage", "damage")
	case EffectPermanentRemoved:
		return "remove a permanent"
	case EffectPermanentTapped:
		return "tap a permanent"
	case EffectManaAdded:
		return "add " + amountNoun(atom, "mana", "mana")
	case EffectTokenCreated:
		return "create " + countNoun(atom, "token")
	case EffectCounterAdded:
		return "put " + countNoun(atom, "counter")
	case EffectCardTutored:
		return "search your library"
	default:
		return ""
	}
}

// describeCost glosses one additional cost. It prefers the cost's preserved
// human-readable text when present, and otherwise synthesizes a phrase for the
// resource-spending kinds. It returns "" for the ubiquitous, low-information
// tap/untap activation costs and any kind it does not model, so Describe omits
// them and the gloss stays focused on costs that change the decision.
func describeCost(c cost.Additional) string {
	if text := strings.TrimSpace(c.Text); text != "" {
		return strings.ToLower(text)
	}
	switch c.Kind {
	case cost.AdditionalSacrifice:
		return "sacrifice " + costCountNoun(c, permanentNoun(c))
	case cost.AdditionalSacrificeSource:
		return "sacrifice this"
	case cost.AdditionalDiscard:
		return "discard " + costCountNoun(c, "card")
	case cost.AdditionalExile:
		return "exile " + costCountNoun(c, "card")
	case cost.AdditionalPayLife:
		return "pay " + strconv.Itoa(costAmount(c)) + " life"
	default:
		return ""
	}
}

// audience renders the subject for an effect whose value sign depends on who it
// affects. The controller (or an unknown audience, scored as the controller's
// own effect) reads as an imperative with no subject; an opponent audience names
// the opponent with a conjugated verb.
func audience(affected Affected, imperative, thirdPerson string) string {
	if affected == AffectedEachOpponent {
		return "each opponent " + thirdPerson
	}
	return imperative
}

// countNoun renders a counted noun ("a card", "2 cards", "X tokens"), pluralized
// by amount and using "X" for a dynamic amount.
func countNoun(atom EffectAtom, noun string) string {
	if atom.IsDynamic {
		return "X " + noun + "s"
	}
	n := atom.Amount
	if n <= 1 {
		return "a " + noun
	}
	return strconv.Itoa(n) + " " + noun + "s"
}

// amountNoun renders a mass-noun amount ("3 damage", "X life"), which is not
// pluralized and uses "X" for a dynamic amount.
func amountNoun(atom EffectAtom, singular, plural string) string {
	noun := plural
	if singular == plural {
		noun = singular
	}
	if atom.IsDynamic {
		return "X " + noun
	}
	n := atom.Amount
	if n <= 0 {
		n = 1
	}
	return strconv.Itoa(n) + " " + noun
}

// costCountNoun renders a counted cost noun ("a creature", "3 cards").
func costCountNoun(c cost.Additional, noun string) string {
	n := costAmount(c)
	if n <= 1 {
		return "a " + noun
	}
	return strconv.Itoa(n) + " " + noun + "s"
}

// costAmount returns a cost's required count, treating the zero default as one
// (CR: zero means one for object and card costs).
func costAmount(c cost.Additional) int {
	if c.Amount <= 0 {
		return 1
	}
	return c.Amount
}

// permanentNoun names the permanent a sacrifice cost consumes, using the
// constrained type when the cost matches one and "permanent" otherwise.
func permanentNoun(c cost.Additional) string {
	if c.MatchPermanentType && c.PermanentType != "" {
		return strings.ToLower(string(c.PermanentType))
	}
	return "permanent"
}
