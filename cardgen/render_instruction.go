package cardgen

import (
	"errors"
	"fmt"
	"strings"

	"github.com/natefinch/council4/mtg/game"
)

func (r Renderer) renderAbilityContent(ctx *renderCtx, content game.AbilityContent) (string, error) {
	if !content.IsModal() {
		mode, err := r.renderMode(ctx, content.Modes[0])
		if err != nil {
			return "", err
		}
		return mode + ".Ability()", nil
	}
	return r.renderModalAbilityContent(ctx, content)
}

// renderModalAbilityContent renders a modal game.AbilityContent with multiple
// modes, MinModes, and MaxModes as a game.AbilityContent struct literal.
func (r Renderer) renderModalAbilityContent(ctx *renderCtx, content game.AbilityContent) (string, error) {
	if len(content.Modes) == 0 {
		return "", errors.New("render: modal ability content has no modes")
	}
	modeElements := make([]string, 0, len(content.Modes))
	for i := range content.Modes {
		rendered, err := r.renderMode(ctx, content.Modes[i])
		if err != nil {
			return "", err
		}
		modeElements = append(modeElements, rendered+",")
	}
	var fields []string
	if len(content.SharedTargets) > 0 {
		sharedElements := make([]string, 0, len(content.SharedTargets))
		for i := range content.SharedTargets {
			rendered, err := r.renderTargetSpec(ctx, &content.SharedTargets[i])
			if err != nil {
				return "", err
			}
			sharedElements = append(sharedElements, rendered+",")
		}
		fields = append(fields, sliceField("SharedTargets", "game.TargetSpec", sharedElements))
	}
	fields = append(fields, sliceField("Modes", "game.Mode", modeElements))
	if content.RandomModes {
		fields = append(fields, "RandomModes: true,")
	}
	if content.ModesUniquePerTurn {
		fields = append(fields, "ModesUniquePerTurn: true,")
	}
	if content.MinModes != 0 {
		fields = append(fields, fmt.Sprintf("MinModes: %d,", content.MinModes))
	}
	if content.MaxModes != 0 {
		fields = append(fields, fmt.Sprintf("MaxModes: %d,", content.MaxModes))
	}
	if content.ModeChoiceBonus != (game.ModeChoiceBonus{}) {
		condition := ""
		switch content.ModeChoiceBonus.Condition {
		case game.ModeChoiceConditionControlsCommander:
			condition = "game.ModeChoiceConditionControlsCommander"
		case game.ModeChoiceConditionSpellKicked:
			condition = "game.ModeChoiceConditionSpellKicked"
		default:
			return "", fmt.Errorf("render: unsupported modal choice bonus condition %d", content.ModeChoiceBonus.Condition)
		}
		bonusFields := []string{fmt.Sprintf("Condition: %s", condition)}
		if content.ModeChoiceBonus.AdditionalMaxModes != 0 {
			bonusFields = append(bonusFields, fmt.Sprintf("AdditionalMaxModes: %d", content.ModeChoiceBonus.AdditionalMaxModes))
		}
		if content.ModeChoiceBonus.ReplaceRange {
			bonusFields = append(bonusFields,
				"ReplaceRange: true",
				fmt.Sprintf("MinModes: %d", content.ModeChoiceBonus.MinModes),
				fmt.Sprintf("MaxModes: %d", content.ModeChoiceBonus.MaxModes),
			)
		}
		fields = append(fields, fmt.Sprintf(
			"ModeChoiceBonus: game.ModeChoiceBonus{%s},",
			strings.Join(bonusFields, ", "),
		))
	}
	if content.EscalateCost.Exists {
		renderedCost, err := r.renderManaCost(ctx, content.EscalateCost.Val)
		if err != nil {
			return "", err
		}
		ctx.need(importOpt)
		fields = append(fields, fmt.Sprintf("EscalateCost: opt.Val(%s),", renderedCost))
	}
	return structLit("game.AbilityContent", fields), nil
}

func (r Renderer) renderMode(ctx *renderCtx, mode game.Mode) (string, error) {
	var fields []string
	if mode.Text != "" {
		fields = append(fields, fmt.Sprintf("Text: %q,", mode.Text))
	}
	if mode.Cost.Exists {
		renderedCost, err := r.renderManaCost(ctx, mode.Cost.Val)
		if err != nil {
			return "", err
		}
		ctx.need(importOpt)
		fields = append(fields, fmt.Sprintf("Cost: opt.Val(%s),", renderedCost))
	}
	if len(mode.Targets) > 0 {
		elements := make([]string, 0, len(mode.Targets))
		for i := range mode.Targets {
			rendered, err := r.renderTargetSpec(ctx, &mode.Targets[i])
			if err != nil {
				return "", err
			}
			elements = append(elements, rendered+",")
		}
		fields = append(fields, sliceField("Targets", "game.TargetSpec", elements))
	}
	elements := make([]string, 0, len(mode.Sequence))
	for i := range mode.Sequence {
		rendered, err := r.renderInstruction(ctx, &mode.Sequence[i])
		if err != nil {
			return "", err
		}
		elements = append(elements, rendered+",")
	}
	fields = append(fields, sliceField("Sequence", "game.Instruction", elements))
	return structLit("game.Mode", fields), nil
}

func (r Renderer) renderInstruction(ctx *renderCtx, instruction *game.Instruction) (string, error) {
	var fields []string
	if instruction.Primitive != nil {
		primitive, err := r.renderPrimitive(ctx, instruction.Primitive)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("Primitive: %s,", primitive))
	}
	if instruction.Condition.Exists {
		condition, err := r.renderEffectCondition(ctx, &instruction.Condition.Val)
		if err != nil {
			return "", err
		}
		ctx.need(importOpt)
		fields = append(fields, fmt.Sprintf("Condition: opt.Val(%s),", condition))
	}
	if instruction.CardCondition.Exists {
		condition, err := r.renderCardSelection(ctx, instruction.CardCondition.Val)
		if err != nil {
			return "", err
		}
		ctx.need(importOpt)
		fields = append(fields, fmt.Sprintf("CardCondition: opt.Val(%s),", condition))
	}
	if instruction.ResultGate.Exists {
		gate, err := renderInstructionResultGate(instruction.ResultGate.Val)
		if err != nil {
			return "", err
		}
		ctx.need(importOpt)
		fields = append(fields, fmt.Sprintf("ResultGate: opt.Val(%s),", gate))
	}
	if instruction.Optional {
		fields = append(fields, "Optional: true,")
	}
	if instruction.OptionalActor.Exists {
		actor, err := r.renderPlayerReference(instruction.OptionalActor.Val)
		if err != nil {
			return "", err
		}
		ctx.need(importOpt)
		fields = append(fields, fmt.Sprintf("OptionalActor: opt.Val(%s),", actor))
	}
	if instruction.OptionalActorGroup.Exists {
		group, err := renderPlayerGroupReference(instruction.OptionalActorGroup.Val)
		if err != nil {
			return "", err
		}
		ctx.need(importOpt)
		fields = append(fields, fmt.Sprintf("OptionalActorGroup: opt.Val(%s),", group))
	}
	if instruction.ForEachPlayerGroup.Exists {
		group, err := renderPlayerGroupReference(instruction.ForEachPlayerGroup.Val)
		if err != nil {
			return "", err
		}
		ctx.need(importOpt)
		fields = append(fields, fmt.Sprintf("ForEachPlayerGroup: opt.Val(%s),", group))
	}
	if instruction.TemptingOffer {
		fields = append(fields, "TemptingOffer: true,")
	}
	if len(instruction.TemptingOfferBody) > 0 {
		var bodyLits []string
		for i := range instruction.TemptingOfferBody {
			bodyLit, err := r.renderInstruction(ctx, &instruction.TemptingOfferBody[i])
			if err != nil {
				return "", err
			}
			bodyLits = append(bodyLits, bodyLit+",")
		}
		fields = append(fields, fmt.Sprintf("TemptingOfferBody: []game.Instruction{%s},", strings.Join(bodyLits, "\n")))
	}
	if instruction.PublishResult != "" {
		fields = append(fields, fmt.Sprintf("PublishResult: game.ResultKey(%q),", string(instruction.PublishResult)))
	}
	if instruction.Description != "" {
		fields = append(fields, fmt.Sprintf("Description: %q,", instruction.Description))
	}
	return structLit("", fields), nil
}

func (r Renderer) renderEffectCondition(ctx *renderCtx, condition *game.EffectCondition) (string, error) {
	var fields []string
	if condition.Object.Kind() != game.ObjectReferenceNone {
		object, err := r.renderObjectReference(condition.Object)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("Object: %s,", object))
	}
	if condition.PermanentType.Exists {
		ctx.need(importTypes)
		cardType, err := cardTypeLiteral(condition.PermanentType.Val)
		if err != nil {
			return "", err
		}
		ctx.need(importOpt)
		fields = append(fields, fmt.Sprintf("PermanentType: opt.Val(%s),", cardType))
	}
	if condition.Negate {
		fields = append(fields, "Negate: true,")
	}
	if condition.Condition.Exists {
		nested, err := r.renderControllerControlsCondition(ctx, &condition.Condition.Val, "effect condition")
		if err != nil {
			return "", err
		}
		ctx.need(importOpt)
		fields = append(fields, fmt.Sprintf("Condition: opt.Val(%s),", nested))
	}
	return structLit("game.EffectCondition", fields), nil
}

func renderInstructionResultGate(gate game.InstructionResultGate) (string, error) {
	var fields []string
	if gate.Key != "" {
		fields = append(fields, fmt.Sprintf("Key: %q,", gate.Key))
	}
	if gate.Accepted != game.TriAny {
		accepted, err := renderTriState(gate.Accepted)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("Accepted: %s,", accepted))
	}
	if gate.Succeeded != game.TriAny {
		succeeded, err := renderTriState(gate.Succeeded)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("Succeeded: %s,", succeeded))
	}
	if gate.SearchedLibrary != game.TriAny {
		searchedLibrary, err := renderTriState(gate.SearchedLibrary)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("SearchedLibrary: %s,", searchedLibrary))
	}
	if gate.AmountRange.Exists {
		fields = append(fields, fmt.Sprintf(
			"AmountRange: opt.Val(game.IntRange{Min: %d, Max: %d}),",
			gate.AmountRange.Val.Min, gate.AmountRange.Val.Max,
		))
	}
	return structLit("game.InstructionResultGate", fields), nil
}

func (r Renderer) renderPrimitive(ctx *renderCtx, primitive game.Primitive) (string, error) {
	return r.renderGamePrimitiveValue(ctx, primitive)
}

func (r Renderer) renderCardSelection(ctx *renderCtx, condition game.CardSelection) (string, error) {
	card, err := renderCardReference(condition.Card)
	if err != nil {
		return "", err
	}
	fields := []string{fmt.Sprintf("Card: %s,", card)}
	if !condition.Selection.Empty() {
		selection, err := r.renderSelection(ctx, condition.Selection)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("Selection: %s,", selection))
	}
	return structLit("game.CardSelection", fields), nil
}
