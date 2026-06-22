package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
)

// lowerDestroyThenSearchSequence lowers the "Destroy target <permanent>.
// <library-search group>" family (Field of Ruin, Tectonic Edge, Ghost
// Quarter): an exact single-target destruction followed by one library-search
// group performed either by the controller ("Search your library ...") or by
// every player ("Each player searches their library ..."). The search group is
// a multi-effect run (search, put, then shuffle) that the per-effect
// ordered-sequence loop cannot keep together, so this dedicated lowerer groups
// it via searchGroupSpec and emits the destruction followed by one Search
// primitive. It fails closed (ok=false) for any shape it cannot model exactly.
func lowerDestroyThenSearchSequence(ctx contentCtx) (game.AbilityContent, bool) {
	content := ctx.content
	if ctx.optional ||
		len(content.Modes) != 0 ||
		len(content.Keywords) != 0 ||
		len(content.Conditions) != 0 ||
		len(content.Targets) != 1 ||
		len(content.Effects) < 4 ||
		content.Effects[0].Kind != compiler.EffectDestroy {
		return game.AbilityContent{}, false
	}
	destroy := &content.Effects[0]
	if destroy.Negated ||
		!destroy.Exact ||
		destroy.Context != parser.EffectContextController ||
		content.Targets[0].Cardinality.Min != 1 ||
		content.Targets[0].Cardinality.Max != 1 {
		return game.AbilityContent{}, false
	}
	groups, _, ok := splitSequenceSearchGroups(content.Effects[1:])
	if !ok || len(groups) != 1 {
		return game.AbilityContent{}, false
	}
	player, searcherGroup, ok := searchGroupSearcher(&content.Effects[1])
	if !ok {
		return game.AbilityContent{}, false
	}
	groupSearcher := searcherGroup.Kind != game.PlayerGroupReferenceNone
	for i := range content.References {
		ref := content.References[i]
		if ref.Binding == compiler.ReferenceBindingPriorInstructionResult {
			continue
		}
		// "Each player searches their library ..." — the "their" possessive is
		// realized by the all-players group searcher and needs no per-reference
		// lowering.
		if groupSearcher && isPlayerPronoun(ref.Pronoun) {
			continue
		}
		return game.AbilityContent{}, false
	}
	targetSpec, ok := permanentTargetSpec(content.Targets[0])
	if !ok {
		return game.AbilityContent{}, false
	}
	searchSeq, ok := searchGroupInstructionsWithSearcher(groups[0], player, searcherGroup)
	if !ok {
		return game.AbilityContent{}, false
	}
	sequence := []game.Instruction{{Primitive: game.Destroy{
		Object:              game.TargetPermanentReference(0),
		PreventRegeneration: destroy.PreventRegeneration,
	}}}
	sequence = append(sequence, searchSeq...)
	return game.Mode{Targets: []game.TargetSpec{targetSpec}, Sequence: sequence}.Ability(), true
}

// searchGroupSearcher resolves the player(s) that perform a library-search group
// from the search effect's subject context: the controller ("Search your
// library ...") or every player ("Each player searches their library ..."). It
// fails closed for a target-player searcher, which requires an ability target
// spec the grouped-sequence callers do not thread.
func searchGroupSearcher(search *compiler.CompiledEffect) (game.PlayerReference, game.PlayerGroupReference, bool) {
	switch search.Context {
	case parser.EffectContextController:
		return game.ControllerReference(), game.PlayerGroupReference{}, true
	case parser.EffectContextEachPlayer:
		return game.PlayerReference{}, game.AllPlayersReference(), true
	default:
		return game.PlayerReference{}, game.PlayerGroupReference{}, false
	}
}

// searchGroupInstructionsWithSearcher builds the runtime instructions for one
// library-search group performed by the given single player or player group. A
// group carrying an in-clause rider is not modeled here and fails closed.
func searchGroupInstructionsWithSearcher(
	group searchGroup,
	player game.PlayerReference,
	searcherGroup game.PlayerGroupReference,
) ([]game.Instruction, bool) {
	if group.RiderIndex != 0 {
		return nil, false
	}
	return []game.Instruction{{Primitive: game.Search{
		Player:      player,
		PlayerGroup: searcherGroup,
		Spec:        group.Spec,
		Amount:      game.Fixed(group.Amount),
	}}}, true
}
