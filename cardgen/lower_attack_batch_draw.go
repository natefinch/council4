package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// lowerAttackBatchEventPlayerDraw lowers the payoff of an attack-batch trigger
// whose effect makes the attacking player draw a card, gated on none of the
// attackers in that batch having attacked the ability controller — "Whenever
// another player attacks with two or more creatures, they draw a card if none
// of those creatures attacked you." (Firemane Commando). The "they" back-
// reference denotes the attacking player recorded on the triggering
// EventAttackerDeclared event, so the draw resolves for the event player. The
// trailing "if none of those creatures attacked you" gate is lowered to an
// effect condition that reads the declared attack batch from the resolving
// stack object's trigger event and holds when that batch declared no direct
// attack on the controller (attacks on another player, on any planeswalker, or
// on a battle do not count).
//
// It handles only the mandatory, non-modal, single fixed event-player card
// draw carrying that one gate and fails closed (ok == false) for anything else,
// so unrelated attack triggers fall through to the normal content lowering
// unchanged.
func lowerAttackBatchEventPlayerDraw(
	content compiler.AbilityContent,
	pattern game.TriggerPattern,
	optional bool,
) (game.AbilityContent, bool) {
	if optional {
		return game.AbilityContent{}, false
	}
	if pattern.Event != game.EventAttackerDeclared || !pattern.OneOrMore {
		return game.AbilityContent{}, false
	}
	if len(content.Effects) != 1 ||
		len(content.Targets) != 0 ||
		len(content.Modes) != 0 ||
		len(content.Keywords) != 0 ||
		len(content.Conditions) != 1 {
		return game.AbilityContent{}, false
	}
	effect := content.Effects[0]
	if effect.Kind != compiler.EffectDraw ||
		effect.Context != parser.EffectContextEventPlayer ||
		!effect.Exact ||
		effect.Negated ||
		effect.Optional ||
		!effect.Amount.Known ||
		effect.Amount.Value < 1 {
		return game.AbilityContent{}, false
	}
	condition := content.Conditions[0]
	if condition.Predicate != compiler.ConditionPredicateNoAttackerAttackedController ||
		condition.Negated {
		return game.AbilityContent{}, false
	}
	if !attackBatchDrawReferencesSupported(content.References) {
		return game.AbilityContent{}, false
	}
	gate, ok := lowerCondition(condition, conditionContextEffectGate)
	if !ok {
		return game.AbilityContent{}, false
	}
	return game.Mode{
		Sequence: []game.Instruction{{
			Primitive: game.Draw{
				Amount: game.Fixed(effect.Amount.Value),
				Player: game.EventPlayerReference(),
			},
			Condition: opt.Val(game.EffectCondition{
				Text:      condition.Text,
				Condition: opt.Val(gate),
			}),
		}},
	}.Ability(), true
}

// attackBatchDrawReferencesSupported reports whether the references carried by
// an attack-batch event-player draw are only the expected anaphors: a subject
// pronoun (they/them/their) that names the drawing attacker and the "those"
// pronoun that names the attack batch inside the gate. The draw always resolves
// for the event player recorded on the triggering attack-declared event
// (EffectContextEventPlayer), so this only guards the reference shape rather
// than the binding. Any other reference leaves the body unsupported so it fails
// closed. At least one subject pronoun must be present.
func attackBatchDrawReferencesSupported(references []compiler.CompiledReference) bool {
	sawEventPlayer := false
	for i := range references {
		reference := references[i]
		if reference.Kind != compiler.ReferencePronoun {
			return false
		}
		switch reference.Pronoun {
		case compiler.ReferencePronounThey,
			compiler.ReferencePronounTheir,
			compiler.ReferencePronounThem:
			sawEventPlayer = true
		case compiler.ReferencePronounThose:
		default:
			return false
		}
	}
	return sawEventPlayer
}
