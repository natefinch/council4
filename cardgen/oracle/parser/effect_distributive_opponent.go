package parser

import (
	"strings"

	"github.com/natefinch/council4/mtg/game/zone"
)

// exactExileForEachOpponentEffectSyntax recognizes the distributive enters
// clause "for each opponent, exile up to one target permanent that player
// controls with mana value 3 or greater." (King Solomon's Frogs). The leading
// "for each opponent," distributes a single "up to one" target pool across every
// opponent; the controller chooses one eligible permanent each opponent controls
// at resolution. The "that player" reference is the distributive anchor rather
// than a second object.
//
// It mirrors exactDestroyForEachPlayerEffectSyntax exactly except that it
// distributes across opponents (not every player) and exiles rather than
// destroys. effectSubjectStart drops the "for each opponent," prefix from the
// reconstructed clause text, so the recognizer confirms that prefix on the full
// effect clause text (with any leading intervening condition such as "if you
// cast it," already stripped) and rebuilds the remainder from the single target.
// Any other exile shape leaves the clause non-exact so lowering fails closed.
func exactExileForEachOpponentEffectSyntax(effect *EffectSyntax) bool {
	if effect.Kind != EffectExile || effect.Negated || effect.Optional {
		return false
	}
	if effect.Context != EffectContextController {
		return false
	}
	if effect.Duration != EffectDurationNone || effect.FromZone != zone.None || effect.ToZone != zone.None {
		return false
	}
	if len(effect.Targets) != 1 {
		return false
	}
	if effect.Targets[0].Cardinality.Min != 0 || effect.Targets[0].Cardinality.Max != 1 {
		return false
	}
	if !exileForEachOpponentReferences(effect.References) {
		return false
	}
	if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(fullEffectClauseText(effect))), "for each opponent, ") {
		return false
	}
	if !strings.EqualFold(exactEffectClauseText(effect), "Exile "+effect.Targets[0].Text+".") {
		return false
	}
	effect.ExileForEachOpponent = true
	return true
}

// exileForEachOpponentReferences confirms the distributive exile clause carries
// exactly the one "that player" anchor its wording requires. That anchor is the
// per-opponent distribution reference rather than a resolving object, so the
// lowering consumes it in place of a target binding.
func exileForEachOpponentReferences(references []Reference) bool {
	return len(references) == 1 && references[0].Kind == ReferenceThatPlayer
}

// exactDrawForEachExiledThisWayEffectSyntax recognizes the per-controller payoff
// clause "For each permanent exiled this way, its controller draws a card."
// (King Solomon's Frogs). The "exiled this way" set is the permanents a sibling
// distributive exile removed; each one's controller draws a single card, so the
// count is one per exiled permanent rather than a multiplier on the draw. The
// "its controller" subject is the referenced-object-controller form already
// modeled by the card-count machinery; only the "For each permanent exiled this
// way," distribution prefix is new.
//
// It mirrors exactCreateTokenForEachExiledThisWayEffectSyntax except that the
// payoff draws a card ("permanent") rather than creating a token ("creature").
// effectSubjectStart drops that prefix from the reconstructed clause text, so the
// recognizer confirms the prefix on the full effect text and rebuilds the
// remainder from the referenced-controller subject. Any other draw shape leaves
// the clause non-exact so lowering fails closed.
func exactDrawForEachExiledThisWayEffectSyntax(effect *EffectSyntax) bool {
	if effect.Kind != EffectDraw || effect.Negated || effect.Optional {
		return false
	}
	if effect.Context != EffectContextReferencedObjectController {
		return false
	}
	if len(effect.Targets) != 0 {
		return false
	}
	if effect.Amount.DynamicForm != EffectDynamicAmountFormNone ||
		!effect.Amount.Known || effect.Amount.Value != 1 {
		return false
	}
	if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(fullEffectClauseText(effect))),
		"for each permanent exiled this way, ") {
		return false
	}
	subject := referencedControllerSubject(effect)
	if subject == "" {
		return false
	}
	if !strings.EqualFold(exactEffectClauseText(effect), subject+" draws a card.") {
		return false
	}
	effect.DrawForEachExiledThisWay = true
	return true
}
