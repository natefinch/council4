package cardgen

import (
	"errors"
	"fmt"

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
	fields := []string{sliceField("Modes", "game.Mode", modeElements)}
	if content.MinModes != 0 {
		fields = append(fields, fmt.Sprintf("MinModes: %d,", content.MinModes))
	}
	if content.MaxModes != 0 {
		fields = append(fields, fmt.Sprintf("MaxModes: %d,", content.MaxModes))
	}
	return structLit("game.AbilityContent", fields), nil
}

func (r Renderer) renderMode(ctx *renderCtx, mode game.Mode) (string, error) {
	var fields []string
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
	primitive, err := r.renderPrimitive(ctx, instruction.Primitive)
	if err != nil {
		return "", err
	}
	fields := []string{fmt.Sprintf("Primitive: %s,", primitive)}
	if instruction.Condition.Exists {
		condition, err := r.renderEffectCondition(ctx, &instruction.Condition.Val)
		if err != nil {
			return "", err
		}
		ctx.need(importOpt)
		fields = append(fields, fmt.Sprintf("Condition: opt.Val(%s),", condition))
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
	return structLit("game.InstructionResultGate", fields), nil
}

func (r Renderer) renderPrimitive(ctx *renderCtx, primitive game.Primitive) (string, error) {
	if primitive == nil {
		return "", errors.New("render: nil primitive")
	}
	switch primitive.Kind() {
	case game.PrimitiveDamage:
		return r.renderDamagePrimitive(ctx, primitive)
	case game.PrimitiveDraw, game.PrimitiveDiscard, game.PrimitiveMill,
		game.PrimitiveScry, game.PrimitiveSurveil, game.PrimitiveGainLife,
		game.PrimitiveLoseLife:
		return r.renderPlayerAmountPrimitive(ctx, primitive)
	case game.PrimitiveInvestigate, game.PrimitiveProliferate, game.PrimitiveManifest:
		return r.renderStandalonePrimitive(ctx, primitive)
	case game.PrimitiveDig:
		value, ok := primitive.(game.Dig)
		if !ok {
			return "", errors.New("render: internal error: Dig kind has unexpected concrete type")
		}
		return r.renderDigPrimitive(ctx, value)
	case game.PrimitiveDestroy, game.PrimitiveBounce, game.PrimitiveUntap,
		game.PrimitiveTap, game.PrimitiveExile:
		return r.renderObjectOrGroupPrimitive(ctx, primitive)
	case game.PrimitiveRegenerate, game.PrimitiveExplore,
		game.PrimitiveCounterObject, game.PrimitiveSacrifice, game.PrimitiveSkipNextUntap:
		return r.renderObjectPrimitive(primitive)
	case game.PrimitiveSearch:
		value, ok := primitive.(game.Search)
		if !ok {
			return "", errors.New("render: internal error: Search kind has unexpected concrete type")
		}
		return r.renderSearchPrimitive(ctx, value)
	case game.PrimitiveAddMana:
		value, ok := primitive.(game.AddMana)
		if !ok {
			return "", errors.New("render: internal error: AddMana kind has unexpected concrete type")
		}
		return r.renderAddMana(ctx, &value)
	case game.PrimitiveAddCounter:
		value, ok := primitive.(game.AddCounter)
		if !ok {
			return "", errors.New("render: internal error: AddCounter kind has unexpected concrete type")
		}
		return r.renderAddCounter(ctx, &value)
	case game.PrimitiveAddPlayerCounter:
		value, ok := primitive.(game.AddPlayerCounter)
		if !ok {
			return "", errors.New("render: internal error: AddPlayerCounter kind has unexpected concrete type")
		}
		return r.renderAddPlayerCounter(ctx, &value)
	case game.PrimitiveModifyPT:
		value, ok := primitive.(game.ModifyPT)
		if !ok {
			return "", errors.New("render: internal error: ModifyPT kind has unexpected concrete type")
		}
		return r.renderModifyPT(ctx, &value)
	case game.PrimitiveFight:
		return r.renderFightPrimitive(primitive)
	case game.PrimitiveChoose:
		value, ok := primitive.(game.Choose)
		if !ok {
			return "", errors.New("render: internal error: Choose kind has unexpected concrete type")
		}
		return r.renderChoose(ctx, value)
	case game.PrimitivePay:
		value, ok := primitive.(game.Pay)
		if !ok {
			return "", errors.New("render: internal error: Pay kind has unexpected concrete type")
		}
		return r.renderPay(ctx, value)
	case game.PrimitivePutOnBattlefield:
		value, ok := primitive.(game.PutOnBattlefield)
		if !ok {
			return "", errors.New("render: internal error: PutOnBattlefield kind has unexpected concrete type")
		}
		return r.renderPutOnBattlefield(ctx, value)
	case game.PrimitiveMoveCard:
		value, ok := primitive.(game.MoveCard)
		if !ok {
			return "", errors.New("render: internal error: MoveCard kind has unexpected concrete type")
		}
		return r.renderMoveCard(ctx, value)
	case game.PrimitiveGrantCastPermission:
		value, ok := primitive.(game.GrantCastPermission)
		if !ok {
			return "", errors.New("render: internal error: GrantCastPermission kind has unexpected concrete type")
		}
		return r.renderGrantCastPermission(ctx, value)
	case game.PrimitiveCreateDelayedTrigger:
		value, ok := primitive.(game.CreateDelayedTrigger)
		if !ok {
			return "", errors.New("render: internal error: CreateDelayedTrigger kind has unexpected concrete type")
		}
		return r.renderCreateDelayedTrigger(ctx, value)
	case game.PrimitiveApplyContinuous:
		value, ok := primitive.(game.ApplyContinuous)
		if !ok {
			return "", errors.New("render: internal error: ApplyContinuous kind has unexpected concrete type")
		}
		return r.renderApplyContinuousPrimitive(ctx, value)
	case game.PrimitiveApplyRule:
		value, ok := primitive.(game.ApplyRule)
		if !ok {
			return "", errors.New("render: internal error: ApplyRule kind has unexpected concrete type")
		}
		return r.renderApplyRulePrimitive(ctx, value)
	case game.PrimitiveSacrificePermanents:
		value, ok := primitive.(game.SacrificePermanents)
		if !ok {
			return "", errors.New("render: internal error: SacrificePermanents kind has unexpected concrete type")
		}
		return r.renderSacrificePermanents(ctx, &value)
	case game.PrimitiveCreateToken:
		value, ok := primitive.(game.CreateToken)
		if !ok {
			return "", errors.New("render: internal error: CreateToken kind has unexpected concrete type")
		}
		return r.renderCreateToken(ctx, value)
	default:
		return "", fmt.Errorf("render: unsupported primitive kind %d", primitive.Kind())
	}
}

func (r Renderer) renderSearchPrimitive(ctx *renderCtx, value game.Search) (string, error) {
	player, err := r.renderPlayerReference(value.Player)
	if err != nil {
		return "", err
	}
	amount, err := r.renderQuantity(ctx, value.Amount)
	if err != nil {
		return "", err
	}
	source, err := renderZone(value.Spec.SourceZone)
	if err != nil {
		return "", err
	}
	destination, err := renderZone(value.Spec.Destination)
	if err != nil {
		return "", err
	}
	ctx.need(importZone)
	specFields := []string{
		fmt.Sprintf("SourceZone: %s,", source),
		fmt.Sprintf("Destination: %s,", destination),
	}
	if value.Spec.CardType.Exists {
		cardType, err := cardTypeLiteral(value.Spec.CardType.Val)
		if err != nil {
			return "", err
		}
		ctx.need(importOpt)
		ctx.need(importTypes)
		specFields = append(specFields, fmt.Sprintf("CardType: opt.Val(%s),", cardType))
	}
	if value.Spec.Permanent {
		specFields = append(specFields, "Permanent: true,")
	}
	if value.Spec.Supertype.Exists {
		supertype, err := supertypeLiteral(value.Spec.Supertype.Val)
		if err != nil {
			return "", err
		}
		ctx.need(importOpt)
		ctx.need(importTypes)
		specFields = append(specFields, fmt.Sprintf("Supertype: opt.Val(%s),", supertype))
	}
	if len(value.Spec.SubtypesAny) > 0 {
		subtypes, err := renderSubtypeArguments(ctx, value.Spec.SubtypesAny)
		if err != nil {
			return "", err
		}
		specFields = append(specFields, fmt.Sprintf("SubtypesAny: []types.Sub{%s},", subtypes))
	}
	if value.Spec.MaxManaValue.Exists {
		ctx.need(importOpt)
		specFields = append(specFields, fmt.Sprintf("MaxManaValue: opt.Val(%d),", value.Spec.MaxManaValue.Val))
	}
	if value.Spec.Reveal {
		specFields = append(specFields, "Reveal: true,")
	}
	if value.Spec.EntersTapped {
		specFields = append(specFields, "EntersTapped: true,")
	}
	if value.Spec.SplitDestination.Exists {
		splitZone, err := renderZone(value.Spec.SplitDestination.Val.Zone)
		if err != nil {
			return "", err
		}
		ctx.need(importOpt)
		ctx.need(importZone)
		splitFields := []string{fmt.Sprintf("Zone: %s,", splitZone)}
		if value.Spec.SplitDestination.Val.EntersTapped {
			splitFields = append(splitFields, "EntersTapped: true,")
		}
		specFields = append(specFields, fmt.Sprintf("SplitDestination: opt.Val(%s),", structLit("game.SearchDestination", splitFields)))
	}
	return structLit("game.Search", []string{
		fmt.Sprintf("Player: %s,", player),
		fmt.Sprintf("Spec: %s,", structLit("game.SearchSpec", specFields)),
		fmt.Sprintf("Amount: %s,", amount),
	}), nil
}

func (r Renderer) renderApplyContinuousPrimitive(ctx *renderCtx, value game.ApplyContinuous) (string, error) {
	var fields []string
	if value.Object.Exists {
		obj, err := r.renderObjectReference(value.Object.Val)
		if err != nil {
			return "", err
		}
		ctx.need(importOpt)
		fields = append(fields, fmt.Sprintf("Object: opt.Val(%s),", obj))
	}
	effectLiterals := make([]string, 0, len(value.ContinuousEffects))
	for i := range value.ContinuousEffects {
		eff, err := r.renderContinuousEffect(ctx, &value.ContinuousEffects[i])
		if err != nil {
			return "", err
		}
		effectLiterals = append(effectLiterals, eff+",")
	}
	fields = append(fields, sliceField("ContinuousEffects", "game.ContinuousEffect", effectLiterals))
	duration, err := renderDuration(value.Duration)
	if err != nil {
		return "", err
	}
	fields = append(fields, fmt.Sprintf("Duration: %s,", duration))
	if value.PublishLinked != "" {
		fields = append(fields, fmt.Sprintf("PublishLinked: game.LinkedKey(%q),", string(value.PublishLinked)))
	}
	return structLit("game.ApplyContinuous", fields), nil
}

// renderApplyRulePrimitive renders an ApplyRule instruction, the resolving form
// of a rule-effect grant such as "Target creature can't be blocked this turn."
// It mirrors renderApplyContinuousPrimitive: an optional target object, the
// carried RuleEffect declarations, and the duration.
func (r Renderer) renderApplyRulePrimitive(ctx *renderCtx, value game.ApplyRule) (string, error) {
	var fields []string
	if value.Object.Exists {
		obj, err := r.renderObjectReference(value.Object.Val)
		if err != nil {
			return "", err
		}
		ctx.need(importOpt)
		fields = append(fields, fmt.Sprintf("Object: opt.Val(%s),", obj))
	}
	ruleEffects := make([]string, 0, len(value.RuleEffects))
	for i := range value.RuleEffects {
		eff, err := r.renderRuleEffect(ctx, &value.RuleEffects[i])
		if err != nil {
			return "", err
		}
		ruleEffects = append(ruleEffects, eff+",")
	}
	fields = append(fields, sliceField("RuleEffects", "game.RuleEffect", ruleEffects))
	duration, err := renderDuration(value.Duration)
	if err != nil {
		return "", err
	}
	fields = append(fields, fmt.Sprintf("Duration: %s,", duration))
	return structLit("game.ApplyRule", fields), nil
}
