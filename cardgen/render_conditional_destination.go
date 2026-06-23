package cardgen

import (
	"fmt"

	"github.com/natefinch/council4/mtg/game"
)

func (r Renderer) renderConditionalDestinationPlace(ctx *renderCtx, value game.ConditionalDestinationPlace) (string, error) {
	card, err := renderCardReference(value.Card)
	if err != nil {
		return "", err
	}
	fromZone, err := renderZone(value.FromZone)
	if err != nil {
		return "", err
	}
	elseZone, err := renderZone(value.Else)
	if err != nil {
		return "", err
	}
	ctx.need(importZone)
	fields := []string{
		fmt.Sprintf("Card: %s,", card),
		fmt.Sprintf("FromZone: %s,", fromZone),
	}
	if value.CardCondition.Exists {
		condition, condErr := r.renderCardSelection(ctx, value.CardCondition.Val)
		if condErr != nil {
			return "", condErr
		}
		ctx.need(importOpt)
		fields = append(fields, fmt.Sprintf("CardCondition: opt.Val(%s),", condition))
	}
	if value.Condition.Exists {
		condition, condErr := r.renderEffectCondition(ctx, &value.Condition.Val)
		if condErr != nil {
			return "", condErr
		}
		ctx.need(importOpt)
		fields = append(fields, fmt.Sprintf("Condition: opt.Val(%s),", condition))
	}
	if value.EntryTapped {
		fields = append(fields, "EntryTapped: true,")
	}
	fields = append(fields, fmt.Sprintf("Else: %s,", elseZone))
	if value.ElseBottom {
		fields = append(fields, "ElseBottom: true,")
	}
	if value.ElseOptional {
		fields = append(fields, "ElseOptional: true,")
	}
	return structLit("game.ConditionalDestinationPlace", fields), nil
}
