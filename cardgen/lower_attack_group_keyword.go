package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// lowerAttackGroupKeywordGrant lowers the payoff of an attack trigger whose
// effect grants a keyword to the attacking creatures that caused it, as in
// "Whenever one or more creatures you control attack, they gain indestructible
// until end of turn." (Angelic Guardian). The plural back-reference
// "they"/"those creatures" denotes the attacking creatures you control. The
// lowering approximates that set by re-deriving it at resolution: it grants the
// keyword to the "attacking creatures you control" battlefield group via a single
// ApplyContinuous until end of turn. The runtime snapshots the group's members at
// resolution (CR 611.2c), so each creature keeps the keyword for the rest of the
// turn even if it later leaves combat.
//
// This re-derivation matches the declared-attacker set in the common case, but is
// not a faithful binding of the exact objects that triggered the ability: a
// creature put onto the battlefield attacking by another same-declaration trigger
// resolving first would be wrongly included, and a declared attacker removed from
// combat before this effect resolves would be wrongly excluded. Both require
// specific card combinations and stack ordering; a precise binding would need a
// runtime "the attacking creatures that caused this trigger" group reference,
// tracked as future work.
//
// It handles only the mandatory, non-modal, unconditional single keyword-grant
// shape and fails closed (ok == false) for anything else, so unrelated attack
// triggers fall through to the normal content lowering unchanged.
func lowerAttackGroupKeywordGrant(
	content compiler.AbilityContent,
	pattern game.TriggerPattern,
	optional bool,
) (game.AbilityContent, bool) {
	if optional {
		return game.AbilityContent{}, false
	}
	if pattern.Event != game.EventAttackerDeclared ||
		pattern.Controller != game.TriggerControllerYou ||
		!pattern.OneOrMore {
		return game.AbilityContent{}, false
	}
	if len(content.Effects) != 1 ||
		len(content.Targets) != 0 ||
		len(content.Conditions) != 0 ||
		len(content.Modes) != 0 ||
		len(content.References) != 1 {
		return game.AbilityContent{}, false
	}
	effect := content.Effects[0]
	if effect.Kind != compiler.EffectGain ||
		effect.Negated ||
		effect.KeywordGrantChoice ||
		effect.Duration != compiler.DurationUntilEndOfTurn ||
		effect.StaticSubject != compiler.StaticSubjectNone {
		return game.AbilityContent{}, false
	}
	reference := content.References[0]
	if reference.Kind != compiler.ReferencePronoun ||
		(reference.Pronoun != compiler.ReferencePronounThey &&
			reference.Pronoun != compiler.ReferencePronounThose) {
		return game.AbilityContent{}, false
	}
	keywords, abilities, ok := partitionTemporaryKeywords(content.Keywords)
	if !ok || (len(keywords) == 0 && len(abilities) == 0) {
		return game.AbilityContent{}, false
	}
	group := game.BattlefieldGroup(game.Selection{
		RequiredTypes: []types.Card{types.Creature},
		Controller:    game.ControllerYou,
		CombatState:   game.CombatStateAttacking,
	})
	return game.Mode{
		Sequence: []game.Instruction{{
			Primitive: game.ApplyContinuous{
				ContinuousEffects: []game.ContinuousEffect{{
					Layer:        game.LayerAbility,
					Group:        group,
					AddKeywords:  keywords,
					AddAbilities: abilities,
				}},
				Duration: game.DurationUntilEndOfTurn,
			},
		}},
	}.Ability(), true
}
