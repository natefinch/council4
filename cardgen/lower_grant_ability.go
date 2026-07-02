package cardgen

import (
	"fmt"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// lowerGainGrantedAbilitySpell lowers a resolving grant of a quoted triggered
// ability to the source permanent ("This creature gains \"Whenever this creature
// deals combat damage to a player, that player loses the game.\"", the second
// chapter of Summon: Primal Odin) into a permanent ApplyContinuous that adds the
// granted ability for as long as the source remains on the battlefield. It
// mirrors lowerPermanentKeywordGrantSpell, substituting the recursively lowered
// quoted ability for a keyword, and recurses through lowerStaticGrantedQuotedAbility
// so the conferred triggered ability lowers from typed data. It fails closed for
// any subject other than the source and for any non-permanent grant.
func lowerGainGrantedAbilitySpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	unsupported := func() (game.AbilityContent, *shared.Diagnostic) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported keyword or ability grant",
			"the executable source backend does not yet lower spells that grant a keyword or quoted ability",
		)
	}
	// lowerGainGrantedAbilitySpell is reached only from the default arm of
	// lowerImmediateSingleEffectSpellTail, which lowerImmediateSingleEffectSpell
	// dispatches solely in single-effect context (the len==1 gate at
	// lower_spell.go:297, the delayed len==1 gate, RepeatBody==1, and
	// contextForEffect's one-effect slice), so an effect count other than one is
	// a dispatch bug rather than an unsupported card.
	if len(ctx.content.Effects) != 1 {
		panic(fmt.Sprintf(
			"lowerGainGrantedAbilitySpell: reached with %d effects; single-effect dispatch guarantees exactly one",
			len(ctx.content.Effects)))
	}
	effect := ctx.content.Effects[0]
	if len(ctx.content.Targets) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.References) != 1 ||
		ctx.content.References[0].Binding != compiler.ReferenceBindingSource ||
		effect.GainGrantedAbility == nil ||
		effect.Context != parser.EffectContextSource ||
		!effect.Exact ||
		effect.Negated ||
		effect.Optional ||
		ctx.optional ||
		effect.StaticSubject != compiler.StaticSubjectNone ||
		effect.Duration != compiler.DurationNone {
		return unsupported()
	}
	ability, ok := lowerStaticGrantedQuotedAbility(effect.GainGrantedAbility)
	if !ok {
		return unsupported()
	}
	object, ok := lowerObjectReference(ctx.content.References[0], referenceLoweringContext{AllowSource: true})
	if !ok {
		return unsupported()
	}
	return game.Mode{
		Sequence: []game.Instruction{{
			Primitive: game.ApplyContinuous{
				Object: opt.Val(object),
				ContinuousEffects: []game.ContinuousEffect{{
					Layer:        game.LayerAbility,
					AddAbilities: []game.Ability{ability},
				}},
				Duration: game.DurationPermanent,
			},
		}},
	}.Ability(), nil
}
