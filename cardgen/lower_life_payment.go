package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/opt"
)

func lowerLifePaymentClause(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	payer, payerOK := lifePaymentClausePayer(ctx)
	if !effect.Exact || !effect.LifeObject ||
		effect.LifePayment != parser.EffectLifePaymentOptional ||
		!payerOK || effect.Player != parser.EffectPlayerNone ||
		effect.Negated || effect.DelayedTiming != 0 || effect.Duration != 0 ||
		!effect.Amount.Known || effect.Amount.Value <= 0 ||
		len(ctx.content.Effects) != 1 || len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Keywords) != 0 || len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, contentDiagnostic(ctx, "unsupported life payment",
			"resolution life payment requires one optional fixed cost with a proven singular payer")
	}
	return game.Mode{Sequence: []game.Instruction{{Primitive: game.Pay{
		Payment: game.ResolutionPayment{
			Payer: opt.Val(payer),
			AdditionalCosts: []cost.Additional{{
				Kind: cost.AdditionalPayLife, Amount: effect.Amount.Value,
			}},
		},
	}}}}.Ability(), nil
}

func lifePaymentClausePayer(ctx contentCtx) (game.PlayerReference, bool) {
	effect := ctx.content.Effects[0]
	if effect.Context == parser.EffectContextController && len(ctx.content.References) == 0 {
		return game.ControllerReference(), true
	}
	if len(ctx.content.References) != 1 || len(effect.SubjectReferences) != 1 || ctx.triggerOneOrMore {
		return game.PlayerReference{}, false
	}
	reference := ctx.content.References[0]
	if !ctx.content.OwnsSubject(reference) || reference.NodeID != effect.SubjectReferences[0].NodeID {
		return game.PlayerReference{}, false
	}
	if effect.Context == parser.EffectContextReferencedObjectController &&
		reference.Binding == compiler.ReferenceBindingEventPermanent &&
		reference.SubjectDomain() == compiler.ReferenceSubjectPermanent {
		return game.ObjectControllerReference(game.EventPermanentReference()), true
	}
	if (effect.Context == parser.EffectContextEventPlayer || effect.Context == parser.EffectContextReferencedPlayer) &&
		reference.Binding == compiler.ReferenceBindingEventPlayer &&
		reference.SubjectDomain() == compiler.ReferenceSubjectPlayer {
		return game.EventPlayerReference(), true
	}
	return game.PlayerReference{}, false
}
