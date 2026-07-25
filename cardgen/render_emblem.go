package cardgen

import (
	"fmt"

	"github.com/natefinch/council4/mtg/game"
)

// renderEmblemAbility renders one ability an emblem confers, dispatching to the
// concrete ability renderer through comma-ok type assertions.
func (r Renderer) renderEmblemAbility(ctx *renderCtx, ability game.Ability) (string, error) {
	if body, ok := ability.(*game.StaticAbility); ok {
		return r.renderStaticAbility(ctx, body, nil)
	}
	if body, ok := ability.(*game.ManaAbility); ok {
		return r.renderManaAbility(ctx, body)
	}
	if body, ok := ability.(*game.TriggeredAbility); ok {
		return r.renderTriggeredAbility(ctx, body)
	}
	if body, ok := ability.(*game.ActivatedAbility); ok {
		return r.renderActivatedAbility(ctx, body)
	}
	if body, ok := ability.(*game.ReplacementAbility); ok {
		return r.renderReplacementAbility(ctx, body)
	}
	return "", fmt.Errorf("render: unsupported emblem ability: %T", ability)
}
