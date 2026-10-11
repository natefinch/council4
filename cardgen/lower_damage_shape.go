package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
)

func fixedSingleTargetDamageEffect(effect *compiler.CompiledEffect) bool {
	return effect.Kind == compiler.EffectDealDamage &&
		!effect.Negated &&
		!effect.Optional &&
		effect.DelayedTiming == 0 &&
		effect.Amount.Known &&
		effect.Amount.Value >= 1 &&
		effect.Amount.DynamicKind == compiler.DynamicAmountNone &&
		len(effect.DamageRiders) == 0 &&
		effect.DamageRecipient.Reference == parser.DamageRecipientReferenceNone &&
		effect.DamageRecipient.EachSourceRole == parser.DamageRecipientReferenceNone &&
		len(effect.DamageRecipient.GroupSelectors) == 0
}
