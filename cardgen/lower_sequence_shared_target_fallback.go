package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
)

func lowerSharedTargetSequenceClause(
	cardName string,
	ctx contentCtx,
	syntax *parser.Ability,
	allowEventPronoun bool,
) (game.AbilityContent, *shared.Diagnostic) {
	content, diagnostic := lowerSequenceClauseContent(cardName, ctx, ctx.content, ctx.optional, syntax, allowEventPronoun)
	if diagnostic == nil {
		return content, nil
	}
	ctx.content.Targets = nil
	return lowerSequenceClauseContent(cardName, ctx, ctx.content, ctx.optional, syntax, allowEventPronoun)
}
