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
// A grant matches when its granted keyword equals keyword and
// spellKeywordGrantMatches accepts the spell. A one-shot next-spell grant stays
// active here; it is consumed separately, and only when a matching spell is
// actually cast.
func spellGrantedKeyword(g *game.Game, playerID game.PlayerID, card *game.CardDef, cardID id.ID, fromZone zone.Type, keyword game.Keyword) bool {
	if card == nil || keyword == game.KeywordNone {
		return false
	}
	for _, effect := range activeRuleEffects(g) {
		if effect.Kind == game.RuleEffectGrantSpellKeyword &&
			effect.GrantedKeyword == keyword &&
			spellKeywordGrantMatches(g, &effect, playerID, card, cardID, fromZone) {
			return true
		}
	}
	return false
}

// spellKeywordGrantMatches reports whether a RuleEffectGrantSpellKeyword applies
// to a spell caster casts from fromZone whose characteristics are card and whose
// card instance is cardID (zero for a cast copy, which is never a commander).
// The controller relation must resolve to the caster, a set CastFromZone must
// equal fromZone, CardSelection.MatchCommander requires the card to carry the
// commander designation (CR 903.3, an attribute of the card rather than a
// characteristic), and the rest of the selection reads the spell's
// characteristics as characteristicDefinedFace computes them.
func spellKeywordGrantMatches(g *game.Game, effect *game.RuleEffect, caster game.PlayerID, card *game.CardDef, cardID id.ID, fromZone zone.Type) bool {
	if card == nil || !controllerRelationMatches(effect.Controller, caster, effect.AffectedController) {
		return false
	}
	if effect.CastFromZone != zone.None && effect.CastFromZone != fromZone {
		return false
	}
	selection := effect.CardSelection
	if selection.MatchCommander {
		if cardID == 0 || !g.CommanderIDs[cardID] {
			return false
		}
		selection.MatchCommander = false
	}
	return selection.Empty() || cardDefMatchesCostSelection(g, characteristicDefinedFace(card), selection)
}

// consumeNextSpellKeywordGrantEffects applies and consumes any one-shot "The
// next spell you cast this turn has <keyword>." rule effects (Archway of
// Innovation) that match the spell just cast from fromZone. Each matching effect
// is removed from the game's rule effects so later spells are unaffected; the
// grant already fed the payment planner while the spell's cost was being paid,
// so nothing needs to persist onto the spell. It is a no-op when no such
// one-shot grant is active.
//
// Only a matching spell that is actually cast reaches this point: it runs from
// emitSpellCastEvents after the spell is on the stack and its costs are paid, so
// activating abilities, land plays, failed cast attempts, and nonmatching spells
// never consume the grant.
func consumeNextSpellKeywordGrantEffects(g *game.Game, obj *game.StackObject, fromZone zone.Type) {
	if len(g.RuleEffects) == 0 {
		return
	}
	spellDef, ok := spellDefForStackObject(g, obj)
	if !ok {
		return
	}
	cardID := obj.SourceID
	if obj.SourceTokenDef != nil {
		cardID = 0
	}
	kept := g.RuleEffects[:0]
	for i := range g.RuleEffects {
		effect := &g.RuleEffects[i]
		if effect.Kind != game.RuleEffectGrantSpellKeyword ||
			!effect.AppliesToNextSpellOnly ||
			!spellKeywordGrantMatches(g, effect, obj.Controller, spellDef, cardID, fromZone) {
			kept = append(kept, *effect)
			continue
		}
	}
	g.RuleEffects = kept
}
