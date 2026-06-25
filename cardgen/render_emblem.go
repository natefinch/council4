package cardgen

import (
	"fmt"

	"github.com/natefinch/council4/mtg/game"
)

// renderCreateEmblem renders a game.CreateEmblem primitive, emitting each of the
// emblem's conferred abilities through the matching ability renderer. Each
// element is wrapped in new(...) because game.Ability is implemented on pointer
// receivers, so the slice must hold addressable pointers.
func (r Renderer) renderCreateEmblem(ctx *renderCtx, value game.CreateEmblem) (string, error) {
	elements := make([]string, 0, len(value.EmblemAbilities))
	for _, ability := range value.EmblemAbilities {
		rendered, err := r.renderEmblemAbility(ctx, ability)
		if err != nil {
			return "", err
		}
		elements = append(elements, "new("+rendered+"),")
	}
	field := sliceField("EmblemAbilities", "game.Ability", elements)
	return structLit("game.CreateEmblem", []string{field}), nil
}

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
