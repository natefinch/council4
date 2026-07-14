package rules

import "github.com/natefinch/council4/mtg/game"

// additionalSpellCopies returns the active copy addends controlled by player and
// whether their additional copies may choose new targets.
func additionalSpellCopies(g *game.Game, player game.PlayerID) (int, bool) {
	addend := 0
	mayChooseNewTargets := false
	for i := range g.ReplacementEffects {
		replacement := &g.ReplacementEffects[i]
		if replacementCurrentController(g, replacement) != player ||
			replacement.SpellCopyAddend <= 0 ||
			!replacementSourceIsActive(g, replacement) {
			continue
		}
		addend += replacement.SpellCopyAddend
		mayChooseNewTargets = mayChooseNewTargets ||
			replacement.SpellCopyAdditionalMayChooseNewTargets
	}
	return addend, mayChooseNewTargets
}
