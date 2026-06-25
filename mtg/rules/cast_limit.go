package rules

import "github.com/natefinch/council4/mtg/game"

// playerCastLimitThisTurn reports the tightest active per-turn cast limit that
// applies to playerID for a spell of spellDef, if any
// (RuleEffectCastLimitPerTurn; "Each player can't cast more than one spell each
// turn.", Rule of Law). A limit's optional SpellTypes/ExcludedSpellTypes filter
// is honored so a type-scoped cap ignores spells outside its scope. When several
// limits apply the smallest one governs.
func playerCastLimitThisTurn(g *game.Game, playerID game.PlayerID, spellDef *game.CardDef) (int, bool) {
	limit := 0
	limited := false
	effects := activeRuleEffects(g)
	for i := range effects {
		effect := &effects[i]
		if effect.Kind != game.RuleEffectCastLimitPerTurn ||
			!playerRelationMatches(effect.Controller, playerID, effect.AffectedPlayer) ||
			!actionRestrictionTurnActive(g, effect) {
			continue
		}
		if len(effect.SpellTypes) > 0 && !cardDefHasAnyType(spellDef, effect.SpellTypes) {
			continue
		}
		if len(effect.ExcludedSpellTypes) > 0 && cardDefHasAnyType(spellDef, effect.ExcludedSpellTypes) {
			continue
		}
		if !limited || effect.CastLimitPerTurn < limit {
			limit = effect.CastLimitPerTurn
			limited = true
		}
	}
	return limit, limited
}

// spellCastLimitReached reports whether an active per-turn cast limit forbids
// playerID from casting spellDef right now because they have already cast at
// least the limit this turn.
func spellCastLimitReached(g *game.Game, playerID game.PlayerID, spellDef *game.CardDef) bool {
	limit, limited := playerCastLimitThisTurn(g, playerID, spellDef)
	if !limited {
		return false
	}
	return spellsCastThisTurn(g, playerID) >= limit
}
