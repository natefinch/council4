package cardgen

import (
	"fmt"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
)

// createTokenRecipientGroup maps a token-creation effect's player-group recipient
// context onto the runtime player group that receives the token. "Each player
// creates ..." widens to every player; "Each opponent creates ..." widens to the
// controller's opponents. Every other context (the controller, a referenced
// object's controller, a single target, "each other player") has no
// player-group recipient and fails closed.
func createTokenRecipientGroup(context parser.EffectContextKind) (game.PlayerGroupReference, bool) {
	switch context {
	case parser.EffectContextEachPlayer:
		return game.AllPlayersReference(), true
	case parser.EffectContextEachOpponent:
		return game.OpponentsReference(), true
	default:
		return game.PlayerGroupReference{}, false
	}
}

// lowerCreateTokenGroupRecipient lowers the player-group recipient form of a
// fixed-count token creation ("Each player creates a 1/1 white Soldier creature
// token.", "Each opponent creates a Treasure token.") to a single CreateToken
// whose RecipientGroup widens the recipient to every member of the group. Only
// the simple shape the controller path synthesizes is accepted: a fixed-count
// vanilla creature token or predefined artifact token with no targets,
// references, conditions, modes, choice, granted ability, copy, dynamic size, or
// published-link key. Any richer shape fails closed pending follow-up work under
// the token-creation epic.
func lowerCreateTokenGroupRecipient(ctx contentCtx, effect *compiler.CompiledEffect, group game.PlayerGroupReference, publishLinked game.LinkedKey, extraKeywords []parser.KeywordKind, keywordsOK bool) (game.AbilityContent, *shared.Diagnostic) {
	// The sole caller lowerCreateTokenSpellLinked passes its own single-effect
	// content (it already panics on len != 1) and its ctx.content.Effects[0] as
	// effect, and every caller of lowerCreateTokenSpellLinked guarantees that
	// sole effect is an EffectCreate. So neither a count other than one nor a
	// kind other than EffectCreate can reach here — either is a dispatch bug.
	if len(ctx.content.Effects) != 1 {
		panic(fmt.Sprintf(
			"lowerCreateTokenGroupRecipient: reached with %d effects; the EffectCreate dispatch is single-effect",
			len(ctx.content.Effects)))
	}
	if effect.Kind != compiler.EffectCreate {
		panic(fmt.Sprintf(
			"lowerCreateTokenGroupRecipient: reached with effect kind %v; every caller guarantees EffectCreate",
			effect.Kind))
	}
	if !effect.Exact ||
		effect.Negated ||
		effect.DelayedTiming != 0 ||
		!createTokenDurationOK(effect.Duration) ||
		!keywordsOK ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.References) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Modes) != 0 ||
		effect.TokenChoice ||
		effect.TokenPTVariableX ||
		effect.TokenGrantedAbility != nil ||
		len(effect.AdditionalTokens) != 0 ||
		publishLinked != "" {
		return game.AbilityContent{}, unsupportedTokenCreationDiagnostic(ctx)
	}
	def, ok := synthesizeCreatureTokenDef(effect, extraKeywords)
	if !ok && len(extraKeywords) == 0 {
		def, ok = synthesizeNamedArtifactTokenDef(effect)
	}
	if !ok {
		return game.AbilityContent{}, unsupportedTokenCreationDiagnostic(ctx)
	}
	amount, dynamicSize, ok := createTokenAmountAndSize(ctx, effect, game.ObjectReference{})
	if !ok || dynamicSize.Exists {
		return game.AbilityContent{}, unsupportedTokenCreationDiagnostic(ctx)
	}
	// "Each player who controls an artifact or enchantment creates ..." carries a
	// per-member conditional the parser captured onto the create effect. Project
	// the compiled qualifier to a game.Selection and thread it onto the recipient
	// group so the runtime keeps only members controlling a matching permanent.
	// The qualifier describes only the controlled permanent's characteristics
	// (Controller stays Any); per-member control is checked as the effect
	// resolves. An unrepresentable qualifier fails closed.
	if effect.RecipientControlsSelector != nil {
		selection, ok := SelectionForSelector(*effect.RecipientControlsSelector)
		if !ok {
			return game.AbilityContent{}, unsupportedTokenCreationDiagnostic(ctx)
		}
		group = group.ControllingMatching(selection)
	}
	return game.Mode{
		Sequence: []game.Instruction{{
			Primitive: game.CreateToken{
				Amount:         amount,
				Source:         game.TokenDef(def),
				RecipientGroup: group,
				EntryTapped:    effect.Selector.Tapped,
				EntryAttacking: effect.Selector.Attacking,
			},
		}},
	}.Ability(), nil
}
