package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/zone"
)

// spellGrantedKeyword reports whether an active RuleEffectGrantSpellKeyword
// confers keyword on the spell playerID is casting. It backs the payment
// planner's cost-affecting keyword check, so a spell a static or one-shot grant
// touches ("Nonartifact spells you cast have improvise.", Inspiring Statuary;
// "The next spell you cast this turn has improvise.", Archway of Innovation) can
// pay with the granted keyword's machinery in addition to any keyword the card
// carries natively.
//
// A grant matches when its granted keyword equals keyword, its controller
// relation resolves to the caster, and its card selection (empty for an
// unfiltered "spells you cast" grant) accepts the spell's printed
// characteristics. A one-shot next-spell grant stays active here; it is consumed
// separately, and only when a matching spell is actually cast.
func spellGrantedKeyword(g *game.Game, playerID game.PlayerID, card *game.CardDef, _ id.ID, _ zone.Type, keyword game.Keyword) bool {
	if card == nil || keyword == game.KeywordNone {
		return false
	}
	for _, effect := range activeRuleEffects(g) {
		if effect.Kind != game.RuleEffectGrantSpellKeyword ||
			effect.GrantedKeyword != keyword ||
			!controllerRelationMatches(effect.Controller, playerID, effect.AffectedController) {
			continue
		}
		if !effect.CardSelection.Empty() && !cardDefMatchesCostSelection(g, card, effect.CardSelection) {
			continue
		}
		return true
	}
	return false
}

// consumeNextSpellKeywordGrantEffects applies and consumes any one-shot "The
// next spell you cast this turn has <keyword>." rule effects (Archway of
// Innovation) whose controller relation and card selection match the spell just
// cast. Each matching effect is removed from the game's rule effects so later
// spells are unaffected; the grant already fed the payment planner while the
// spell's cost was being paid, so nothing needs to persist onto the spell. It is
// a no-op when no such one-shot grant is active.
//
// Only a matching spell that is actually cast reaches this point: it runs from
// emitSpellCastEvents after the spell is on the stack and its costs are paid, so
// activating abilities, land plays, failed cast attempts, and nonmatching spells
// never consume the grant.
func consumeNextSpellKeywordGrantEffects(g *game.Game, obj *game.StackObject) {
	if len(g.RuleEffects) == 0 {
		return
	}
	spellDef, ok := spellDefForStackObject(g, obj)
	if !ok {
		return
	}
	kept := g.RuleEffects[:0]
	for i := range g.RuleEffects {
		effect := &g.RuleEffects[i]
		if effect.Kind != game.RuleEffectGrantSpellKeyword ||
			!effect.AppliesToNextSpellOnly ||
			!controllerRelationMatches(effect.Controller, obj.Controller, effect.AffectedController) ||
			(!effect.CardSelection.Empty() && !cardDefMatchesCostSelection(g, spellDef, effect.CardSelection)) {
			kept = append(kept, *effect)
			continue
		}
	}
	g.RuleEffects = kept
}
