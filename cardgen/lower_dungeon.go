package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
)

// lowerVentureIntoDungeonSpell lowers the exact controller-scoped "venture into
// the dungeon." keyword action (CR 309.6) to a VentureIntoDungeon primitive.
func lowerVentureIntoDungeonSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	return lowerControllerKeywordAction(ctx, "venture into the dungeon", game.VentureIntoDungeon{Player: game.ControllerReference()})
}

// lowerVentureIntoUndercitySpell lowers the exact controller-scoped "venture into
// Undercity." keyword action to a VentureIntoUndercity primitive.
func lowerVentureIntoUndercitySpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	return lowerControllerKeywordAction(ctx, "venture into Undercity", game.VentureIntoUndercity{Player: game.ControllerReference()})
}

// lowerTakeInitiativeSpell lowers the exact controller-scoped "you take the
// initiative." keyword action (CR 720) to a TakeInitiative primitive.
func lowerTakeInitiativeSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	return lowerControllerKeywordAction(ctx, "take the initiative", game.TakeInitiative{Player: game.ControllerReference()})
}

// lowerControllerKeywordAction lowers an exact controller-scoped keyword action
// with no targets, references, conditions, keywords, or modes to a single
// controller-scoped primitive. Any other shape fails closed.
func lowerControllerKeywordAction(ctx contentCtx, name string, primitive game.Primitive) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	if effect.Negated || ctx.optional || !effect.Exact ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.References) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported "+name+" effect",
			"the executable source backend supports only the exact controller-scoped '"+name+"' effect",
		)
	}
	return game.Mode{
		Sequence: []game.Instruction{{Primitive: primitive}},
	}.Ability(), nil
}
