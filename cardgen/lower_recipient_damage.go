package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
)

// Reference consumption and quantity/recipient binding belong to the factory;
// the frame consumes the common untargeted, undivided damage envelope.
func lowerRecipientDamageSpell(
	ctx contentCtx,
	recipient parser.DamageRecipientReferenceKind,
	detail string,
	primitiveFactory func(compiler.CompiledEffect, []compiler.CompiledReference) (game.Damage, bool),
) (game.AbilityContent, *shared.Diagnostic) {
	assertUndividedRecipientDamageDispatch(ctx, recipient)
	effect := ctx.content.Effects[0]
	unsupported := func() (game.AbilityContent, *shared.Diagnostic) {
		return game.AbilityContent{}, contentDiagnostic(ctx, "unsupported damage spell", detail)
	}
	if !effect.Exact ||
		effect.Negated ||
		len(ctx.content.Targets) != 0 ||
		len(effect.DamageRecipient.GroupSelectors) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(abilityKeywordsExcludingSelectorPredicates(ctx.content)) != 0 ||
		len(ctx.content.Modes) != 0 {
		return unsupported()
	}
	damage, ok := primitiveFactory(effect, ctx.content.References)
	if !ok {
		return unsupported()
	}
	return game.Mode{
		Sequence: []game.Instruction{{Primitive: damage}},
	}.Ability(), nil
}
