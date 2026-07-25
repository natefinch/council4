package parser

import (
	"strings"

	"github.com/natefinch/council4/mtg/game/zone"
)

// DistributiveRemovalScope names the player group a distributive removal clause
// walks.
type DistributiveRemovalScope string

// Distributive removal scopes.
const (
	// DistributiveScopeEachPlayer is the "For each player," prefix.
	DistributiveScopeEachPlayer DistributiveRemovalScope = "EachPlayer"
	// DistributiveScopeEachOpponent is the "For each opponent," prefix.
	DistributiveScopeEachOpponent DistributiveRemovalScope = "EachOpponent"
)

// DistributiveRemovalVerb names the removal a distributive clause applies.
type DistributiveRemovalVerb string

// Distributive removal verbs.
const (
	// DistributiveVerbExile is the "exile" wording.
	DistributiveVerbExile DistributiveRemovalVerb = "Exile"
	// DistributiveVerbDestroy is the "destroy" wording.
	DistributiveVerbDestroy DistributiveRemovalVerb = "Destroy"
)

// DistributiveRemoval marks the distributive removal clause
//
//	"For each <group>, <verb> up to one [other] target <permanent> that player
//	controls[ until <source> leaves the battlefield]."
//
// Each player's permanents form an independent "up to one" pool, and the "that
// player" reference is the distribution anchor rather than a target. Lowering
// turns the whole family into one game.ForEachPlayer primitive; a paired payoff
// clause consumes the linked removed set.
//
// It replaced four booleans, each with its own recognizer that differed only in
// the verb word and the group word. Parameterizing them means a scope and verb
// combination no card had yet - "For each opponent, destroy up to one target
// creature that player controls." - parses without new parser code.
type DistributiveRemoval struct {
	Scope DistributiveRemovalScope
	Verb  DistributiveRemovalVerb
	// UntilSourceLeaves marks the exile-until-leaves duration, whose clause ends
	// "until <source> leaves the battlefield" and carries a second self
	// reference as the duration anchor. Lowering pairs it with a return.
	UntilSourceLeaves bool `json:",omitempty"`
}

// distributiveScopePrefixes maps each recognized clause prefix to its scope.
var distributiveScopePrefixes = map[string]DistributiveRemovalScope{
	"for each player, ":   DistributiveScopeEachPlayer,
	"for each opponent, ": DistributiveScopeEachOpponent,
}

// distributiveRemovalVerbs maps each removal effect kind to its clause verb.
var distributiveRemovalVerbs = map[EffectKind]DistributiveRemovalVerb{
	EffectExile:   DistributiveVerbExile,
	EffectDestroy: DistributiveVerbDestroy,
}

// exactDistributiveRemovalEffectSyntax recognizes every "For each <group>,
// <verb> up to one target <permanent> that player controls" clause, in both the
// plain form (The Curse of Fenric, Unexplained Absence, King Solomon's Frogs)
// and the exile-until-leaves Saga form (Vault 13: Dweller's Journey, Battle at
// the Helvault).
//
// The shape is one clause-level check shared by every combination: an
// unnegated, non-optional controller-context removal with no duration or zone
// movement, exactly one "up to one" target, and the distribution anchor
// references its wording requires. Only the leading group phrase and the
// reconstructed clause text vary, so both are parameters rather than copies.
// Any other removal shape leaves the clause non-exact so lowering fails closed.
func exactDistributiveRemovalEffectSyntax(effect *EffectSyntax) bool {
	verb, ok := distributiveRemovalVerbs[effect.Kind]
	if !ok || effect.Negated || effect.Optional {
		return false
	}
	if effect.Context != EffectContextController {
		return false
	}
	if effect.Duration != EffectDurationNone || effect.FromZone != zone.None || effect.ToZone != zone.None {
		return false
	}
	if len(effect.Targets) != 1 ||
		effect.Targets[0].Cardinality.Min != 0 ||
		effect.Targets[0].Cardinality.Max != 1 {
		return false
	}
	scope, ok := distributiveClauseScope(effect)
	if !ok {
		return false
	}

	// The until-leaves form carries a second self reference as its duration
	// anchor and ends with the matching phrase; the plain form carries only the
	// distribution anchor and ends at the target.
	if sourceRef, ok := exileForEachPlayerReferences(effect.References); ok && verb == DistributiveVerbExile {
		expected := "Exile " + effect.Targets[0].Text + " until " + sourceRef.Text + " leaves the battlefield."
		if !strings.EqualFold(exactEffectClauseText(effect), expected) {
			return false
		}
		stripSourceSubtypeContamination(&effect.Selection, sourceRef)
		effect.DistributiveRemoval = &DistributiveRemoval{Scope: scope, Verb: verb, UntilSourceLeaves: true}
		return true
	}
	if !onlyThatPlayerAnchor(effect.References) {
		return false
	}
	if !strings.EqualFold(exactEffectClauseText(effect), string(verb)+" "+effect.Targets[0].Text+".") {
		return false
	}
	effect.DistributiveRemoval = &DistributiveRemoval{Scope: scope, Verb: verb}
	return true
}

// distributiveClauseScope reads the leading group phrase. It checks both the
// raw clause text and the reconstructed full clause text because an intervening
// condition ("If you cast it, for each opponent, ...") is stripped from one but
// not the other, and either spelling is the same distribution.
func distributiveClauseScope(effect *EffectSyntax) (DistributiveRemovalScope, bool) {
	texts := [...]string{effect.Text, fullEffectClauseText(effect)}
	for _, text := range texts {
		normalized := strings.ToLower(strings.TrimSpace(text))
		for prefix, scope := range distributiveScopePrefixes {
			if strings.HasPrefix(normalized, prefix) {
				return scope, true
			}
		}
	}
	return "", false
}

// onlyThatPlayerAnchor confirms the clause carries exactly the one "that player"
// anchor its wording requires. That anchor is the per-player distribution
// reference rather than a resolving object, so lowering consumes it in place of
// a target binding.
func onlyThatPlayerAnchor(references []Reference) bool {
	return len(references) == 1 && references[0].Kind == ReferenceThatPlayer
}
