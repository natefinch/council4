package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
)

// lowerCreateEmblemContent lowers a single "You get an emblem with \"...\""
// effect into a game.CreateEmblem primitive carrying the runtime abilities the
// emblem confers. Each quoted ability was parsed once by the parser; this
// recursively compiles and lowers each inner document, mirroring the
// reminder-mana-ability pattern, and fails closed if any inner ability does not
// lower.
func lowerCreateEmblemContent(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	if ctx.optional ||
		!effect.Exact ||
		effect.Negated ||
		effect.Context != parser.EffectContextController ||
		len(effect.EmblemAbilities) == 0 ||
		ctx.content.Unconsumed() {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported emblem effect",
			"the executable source backend supports only a controller \"You get an emblem with \\\"...\\\"\" effect whose quoted abilities all lower",
		)
	}
	var abilities []game.Ability
	for i := range effect.EmblemAbilities {
		granted := effect.EmblemAbilities[i]
		lowered, ok := lowerEmblemQuotedAbilities(&granted)
		if !ok {
			return game.AbilityContent{}, contentDiagnostic(
				ctx,
				"unsupported emblem ability",
				"the executable source backend does not yet lower one of this emblem's quoted abilities",
			)
		}
		abilities = append(abilities, lowered...)
	}
	return game.Mode{Sequence: []game.Instruction{{Primitive: game.CreateEmblem{
		EmblemAbilities: abilities,
	}}}}.Ability(), nil
}

// lowerEmblemQuotedAbilities compiles and lowers one quoted emblem ability into
// the runtime abilities it produces. The parser parsed the quoted body once;
// this recursive compile + lower mirrors lowerStaticGrantedQuotedAbility but
// also returns static abilities (an emblem's "Creatures you control have base
// power and toughness 9/9" is a static), so the emblem carries every category
// of inner ability.
func lowerEmblemQuotedAbilities(granted *parser.StaticGrantedAbilitySyntax) ([]game.Ability, bool) {
	innerDocument, innerDiags := granted.Inner()
	if len(innerDiags) != 0 {
		return nil, false
	}
	innerComp, compilerDiags := compiler.Compile(innerDocument, compiler.Context{})
	if len(compilerDiags) != 0 ||
		len(innerComp.Abilities) != 1 ||
		len(innerComp.Syntax.Abilities) != 1 {
		return nil, false
	}
	lowered, diagnostic := lowerExecutableAbility("", false, nil, innerComp.Abilities[0], &innerComp.Syntax.Abilities[0])
	if diagnostic != nil {
		return nil, false
	}
	return loweredAbilityToEmblemAbilities(lowered)
}

// loweredAbilityToEmblemAbilities collects the lowered ability categories an
// emblem may carry — static, activated, mana, triggered, and replacement
// abilities — in Oracle order. It fails closed when the inner ability lowered to
// a spell, overload, loyalty, chapter, additional/alternative cost, or
// characteristic rider an emblem cannot hold, or when no ability was produced.
func loweredAbilityToEmblemAbilities(lowered abilityLowering) ([]game.Ability, bool) {
	if lowered.spellAbility.Exists ||
		lowered.overloadCost.Exists ||
		len(lowered.additionalCosts) != 0 ||
		len(lowered.alternativeCosts) != 0 ||
		lowered.loyaltyAbility.Exists ||
		lowered.chapterAbility.Exists ||
		lowered.entersPrepared ||
		lowered.dynamicPower.Exists ||
		lowered.dynamicToughness.Exists {
		return nil, false
	}
	var abilities []game.Ability
	for i := range lowered.staticAbilities {
		body := lowered.staticAbilities[i].Body
		abilities = append(abilities, &body)
	}
	if lowered.activatedAbility.Exists {
		ability := lowered.activatedAbility.Val
		abilities = append(abilities, &ability)
	}
	if lowered.manaAbility.Exists {
		ability := lowered.manaAbility.Val
		abilities = append(abilities, &ability)
	}
	if lowered.triggeredAbility.Exists {
		ability := lowered.triggeredAbility.Val
		abilities = append(abilities, &ability)
	}
	if lowered.replacementAbility.Exists {
		ability := lowered.replacementAbility.Val
		abilities = append(abilities, &ability)
	}
	if len(abilities) == 0 {
		return nil, false
	}
	return abilities, true
}
