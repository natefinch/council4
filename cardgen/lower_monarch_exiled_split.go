package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

// lowerMonarchExiledCardSplitSequence lowers the resolution of Coin of Fate's
// activated ability: "An opponent chooses one of the exiled cards. You put that
// card on the bottom of your library and return the other to the battlefield
// tapped. You become the monarch." The two cost-exiled creature cards are chosen
// between by an opponent; the chosen card ("that card") goes to the bottom of its
// owner's library and the other returns to the battlefield tapped under the
// controller, then the controller becomes the monarch. The opponent's choice is
// implicit in the compiled body (its sentence carries no effect), so lowering
// synthesizes it through the PartitionExiledCostCards primitive rather than a
// reference. It fails closed for any shape it does not fully model.
func lowerMonarchExiledCardSplitSequence(ctx contentCtx) (game.AbilityContent, bool) {
	content := ctx.content
	if ctx.optional ||
		len(content.Effects) != 3 ||
		len(content.Targets) != 0 ||
		len(content.Conditions) != 0 ||
		len(content.Keywords) != 0 ||
		len(content.Modes) != 0 {
		return game.AbilityContent{}, false
	}
	if !monarchSplitChosenToLibrary(content.Effects[0]) ||
		!monarchSplitOtherToBattlefield(content.Effects[1]) ||
		!monarchSplitBecomeMonarch(content.Effects[2]) {
		return game.AbilityContent{}, false
	}
	// The resolving body's only reference is the put clause's "that card"
	// (validated above); require exactly it so no unmodeled reference is
	// silently dropped. The cost's "this artifact" sacrifice reference is not
	// part of the resolution content.
	if len(content.References) != 1 ||
		content.References[0].Kind != compiler.ReferenceThatObject {
		return game.AbilityContent{}, false
	}
	return game.Mode{Sequence: []game.Instruction{
		{Primitive: game.PartitionExiledCostCards{
			ChooserOpponent:       true,
			ChosenToLibraryBottom: true,
			OtherEntersTapped:     true,
		}},
		{Primitive: game.BecomeMonarch{Player: game.ControllerReference()}},
	}}.Ability(), true
}

// monarchSplitChosenToLibrary reports whether the effect is the controller
// putting the opponent-chosen exiled card ("that card") on the bottom of its
// owner's library. The opponent-chooser role is carried by the parser-recognized
// ExiledCardSplitOpponentChooses flag, which must be set so a "You choose..."
// variant fails closed.
func monarchSplitChosenToLibrary(effect compiler.CompiledEffect) bool {
	return effect.Kind == compiler.EffectPut &&
		!effect.Negated &&
		!effect.Optional &&
		effect.ExiledCardSplitOpponentChooses &&
		effect.Context == parser.EffectContextController &&
		effect.ToZone == zone.Library &&
		effect.Destination == parser.EffectDestinationBottom &&
		len(effect.Targets) == 0 &&
		len(effect.References) == 1 &&
		effect.References[0].Kind == compiler.ReferenceThatObject &&
		effect.References[0].Binding == compiler.ReferenceBindingSource
}

// monarchSplitOtherToBattlefield reports whether the effect returns "the other"
// exiled card to the battlefield tapped under the controller's control. "The
// other" is carried by the selector's Other flag, not by a reference, so the
// clause holds no references of its own.
func monarchSplitOtherToBattlefield(effect compiler.CompiledEffect) bool {
	return effect.Kind == compiler.EffectReturn &&
		!effect.Negated &&
		!effect.Optional &&
		effect.Context == parser.EffectContextController &&
		effect.ToZone == zone.Battlefield &&
		effect.EntersTapped &&
		effect.Selector.Other &&
		len(effect.Targets) == 0 &&
		len(effect.References) == 0
}

// monarchSplitBecomeMonarch reports whether the effect is the controller
// becoming the monarch.
func monarchSplitBecomeMonarch(effect compiler.CompiledEffect) bool {
	return effect.Kind == compiler.EffectBecomeMonarch &&
		!effect.Negated &&
		effect.Context == parser.EffectContextController &&
		len(effect.Targets) == 0 &&
		len(effect.References) == 0
}
