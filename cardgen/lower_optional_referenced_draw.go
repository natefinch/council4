package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// lowerOptionalReferencedPlayerDraw lowers a single-effect optional draw whose
// drawing player is named by the triggering event rather than the ability's
// controller — the "<referenced player> may draw a card" family:
//
//	Whenever this creature attacks, defending player may draw a card.
//	  (Sibilant Spirit, Harbor Guardian)
//	Whenever a creature deals combat damage to one of your opponents, its
//	controller may draw a card. (Edric, Spymaster of Trest; Synapse Sliver)
//	Whenever a creature dies, that creature's controller may draw a card.
//	  (Fecundity)
//
// The drawing player is also the player who decides whether to draw, so the
// produced instruction carries that player as both Draw.Player and
// OptionalActor: the engine asks that player — not the ability's controller —
// whether to apply the optional draw (CR 603.3, the controller of the ability
// is not the affected player here).
//
// It fails closed (ok=false) unless the body is exactly one optional,
// non-negated, fixed-count draw by a supported referenced player with no
// targets, modes, or keywords, and at most one resolution condition that lowers
// through the shared effect-gate path. The referenced player resolves to:
//
//   - DefendingPlayerReference for the defending-player context (no object
//     reference: the defending player is derived from the attack event).
//   - ObjectControllerReference(EventPermanentReference()) for the
//     referenced-object-controller context whose lone subject reference binds
//     the triggering event permanent ("its controller", "that creature's
//     controller").
func lowerOptionalReferencedPlayerDraw(ctx contentCtx) (game.AbilityContent, bool) {
	if ctx.optional ||
		len(ctx.content.Effects) != 1 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Conditions) > 1 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, false
	}
	effect := ctx.content.Effects[0]
	if effect.Kind != compiler.EffectDraw ||
		!effect.Optional ||
		effect.Negated ||
		effect.DelayedTiming != 0 ||
		!effect.Amount.Known ||
		effect.Amount.Value < 1 {
		return game.AbilityContent{}, false
	}
	// A trailing/leading resolution condition ("... may draw a card if its power
	// is greater than each other creature's power"; Selvala, Heart of the Wilds)
	// gates the optional draw when the ability resolves. Lower it through the
	// shared effect-gate path and drop its condition-spanned references before
	// resolving the drawing player, so the "its power" pronoun inside the gate is
	// not mistaken for the draw's own referenced player.
	var gate opt.V[game.EffectCondition]
	drawCtx := ctx
	if len(ctx.content.Conditions) == 1 {
		condition, ok := lowerCondition(ctx.content.Conditions[0], conditionContextEffectGate)
		if !ok {
			return game.AbilityContent{}, false
		}
		gate = opt.Val(game.EffectCondition{Condition: opt.Val(condition)})
		drawCtx.content = contentWithoutConditionSpannedReferences(ctx.content)
	}
	player, ok := optionalDrawReferencedPlayer(drawCtx, drawCtx.content.Effects[0])
	if !ok {
		return game.AbilityContent{}, false
	}
	return game.Mode{Sequence: []game.Instruction{{
		Primitive: game.Draw{
			Player: player,
			Amount: game.Fixed(effect.Amount.Value),
		},
		Optional:      true,
		OptionalActor: opt.Val(player),
		Condition:     gate,
	}}}.Ability(), true
}

// optionalDrawReferencedPlayer resolves the referenced player who draws (and
// decides) for lowerOptionalReferencedPlayerDraw, matching the supported
// trigger-referenced player contexts. It fails closed for any other context or
// reference shape so an unrecognized subject leaves the body unsupported rather
// than drawing for the wrong player.
func optionalDrawReferencedPlayer(
	ctx contentCtx,
	effect compiler.CompiledEffect,
) (game.PlayerReference, bool) {
	switch effect.Context {
	case parser.EffectContextDefendingPlayer:
		if len(ctx.content.References) != 0 {
			return game.PlayerReference{}, false
		}
		return game.DefendingPlayerReference(), true
	case parser.EffectContextReferencedObjectController:
		if len(ctx.content.References) != 1 ||
			ctx.content.References[0].Binding != compiler.ReferenceBindingEventPermanent {
			return game.PlayerReference{}, false
		}
		return game.ObjectControllerReference(game.EventPermanentReference()), true
	default:
		return game.PlayerReference{}, false
	}
}
