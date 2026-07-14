package cardgen

import (
	"fmt"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
)

// goadCreatedTokensLinkKey links the tokens a group-recipient copy create
// produces to the following rest-of-game goad so exactly those tokens are goaded
// (Life of the Party: "each opponent creates a token that's a copy of it. The
// tokens are goaded for the rest of the game."). The runtime scopes the key per
// source object, so a fixed string is unambiguous across cards.
const goadCreatedTokensLinkKey game.LinkedKey = "goad-created-tokens"

// lowerCreateCopyTokenGroupGoadSequence lowers a group-recipient copy-of-reference
// token creation carrying the folded "The tokens are goaded for the rest of the
// game." rider (Life of the Party) into a two-instruction sequence: each opponent
// creates a copy of the referenced permanent, every created token is remembered
// under a link key, and a following goad binds exactly those linked tokens for
// the rest of the game. Only the exact shape is accepted — a group recipient, a
// copy-of-reference source with supported copy modifiers, the single copied
// reference, and no targets, conditions, modes, or outer link; any richer shape
// fails closed.
func lowerCreateCopyTokenGroupGoadSequence(ctx contentCtx, effect *compiler.CompiledEffect, publishLinked game.LinkedKey) (game.AbilityContent, *shared.Diagnostic) {
	// Reached only from lowerCreateTokenSpellLinked's goad-rider branch, which
	// runs on its single-effect content, so a different count is a dispatch bug.
	if len(ctx.content.Effects) != 1 {
		panic(fmt.Sprintf(
			"lowerCreateCopyTokenGroupGoadSequence: reached with %d effects; the goad-rider dispatch is single-effect",
			len(ctx.content.Effects)))
	}
	group, ok := createTokenRecipientGroup(effect.Context)
	if !ok {
		return game.AbilityContent{}, unsupportedTokenCreationDiagnostic(ctx)
	}
	if !effect.Exact ||
		effect.Negated ||
		effect.DelayedTiming != 0 ||
		effect.Duration != compiler.DurationNone ||
		!effect.TokenCopyOfReference ||
		publishLinked != "" ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.References) < 1 ||
		!tokenCopyAuxiliaryReferencesOK(ctx.content.References[1:]) ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != len(effect.TokenCopyGrantKeywords) ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, unsupportedTokenCreationDiagnostic(ctx)
	}
	object, ok := lowerObjectReference(
		ctx.content.References[0],
		referenceLoweringContext{AllowSource: true, AllowTarget: true, AllowEvent: true},
	)
	if !ok {
		return game.AbilityContent{}, unsupportedTokenCreationDiagnostic(ctx)
	}
	spec, ok := tokenCopyModifiers(effect, object)
	if !ok {
		return game.AbilityContent{}, unsupportedTokenCreationDiagnostic(ctx)
	}
	amount, ok := createTokenAmount(ctx, effect, game.ObjectReference{})
	if !ok {
		return game.AbilityContent{}, unsupportedTokenCreationDiagnostic(ctx)
	}
	return game.Mode{
		Sequence: []game.Instruction{
			{Primitive: game.CreateToken{
				Amount:         amount,
				Source:         game.TokenCopyOf(spec),
				RecipientGroup: group,
				EntryTapped:    effect.TokenCopyEntersTapped,
				PublishLinked:  goadCreatedTokensLinkKey,
			}},
			{Primitive: game.Goad{
				Group:      game.LinkedObjectsGroup(goadCreatedTokensLinkKey),
				RestOfGame: true,
			}},
		},
	}.Ability(), nil
}
