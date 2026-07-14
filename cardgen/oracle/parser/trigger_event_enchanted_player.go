package parser

import (
	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// parseEnchantedPlayerAttackedTriggerEventClause recognizes the passive combat
// trigger "enchanted player is attacked" (Curse of Opulence, Curse of
// Disturbance, and their siblings). The clause has no attacker subject or
// recipient of its own: the recipient is the player the source Aura enchants,
// carried through lowering by the typed EnchantedPlayerIsAttacked flag. It
// matches an attacker-declared event once per combat (OneOrMore) in which one or
// more creatures are declared attacking that player directly; an attack on that
// player's planeswalker or battle is not an attack on the player (CR 508.1),
// which the runtime enforces from the flag.
func parseEnchantedPlayerAttackedTriggerEventClause(
	tokens []shared.Token,
	_ TriggerIntroductionKind,
	_ Atoms,
	_ string,
) *TriggerEventClause {
	if !syntaxWordsEqual(tokens, "enchanted", "player", "is", "attacked") {
		return nil
	}
	return &TriggerEventClause{
		Kind:                      TriggerEventKindAttack,
		EnchantedPlayerIsAttacked: true,
		OneOrMore:                 true,
	}
}
