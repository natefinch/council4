package cardgen

import (
	"errors"
	"fmt"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
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
		default:
			return "", fmt.Errorf("render: unsupported modal choice bonus condition %d", content.ModeChoiceBonus.Condition)
		}
		fields = append(fields, fmt.Sprintf(
			"ModeChoiceBonus: game.ModeChoiceBonus{Condition: %s, AdditionalMaxModes: %d},",
			condition,
			content.ModeChoiceBonus.AdditionalMaxModes,
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
	if gate.AmountRange.Exists {
		fields = append(fields, fmt.Sprintf(
			"AmountRange: opt.Val(game.IntRange{Min: %d, Max: %d}),",
			gate.AmountRange.Val.Min, gate.AmountRange.Val.Max,
		))
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
	case game.PrimitiveGroupSourceDamage:
		return r.renderGroupSourceDamage(ctx, primitive)
	case game.PrimitiveGroupSelfPowerDamage:
		return r.renderGroupSelfPowerDamage(ctx, primitive)
	case game.PrimitiveDraw, game.PrimitiveDiscard, game.PrimitiveMill,
		game.PrimitiveScry, game.PrimitiveSurveil, game.PrimitiveGainLife,
		game.PrimitiveLoseLife, game.PrimitiveReorderLibraryTop,
		game.PrimitiveExileTopOfLibrary:
		return r.renderPlayerAmountPrimitive(ctx, primitive)
	case game.PrimitivePlayerLosesGame:
		value, err := assertPrimitive[game.PlayerLosesGame](primitive)
		if err != nil {
			return "", err
		}
		return r.renderPlayerLosesGame(value)
	case game.PrimitivePlayerWinsGame:
		value, err := assertPrimitive[game.PlayerWinsGame](primitive)
		if err != nil {
			return "", err
		}
		return r.renderPlayerWinsGame(value)
	case game.PrimitiveInvestigate, game.PrimitiveProliferate, game.PrimitiveManifest,
		game.PrimitiveDiscoverCards:
		return r.renderStandalonePrimitive(ctx, primitive)
	case game.PrimitiveAmass:
		value, err := assertPrimitive[game.Amass](primitive)
		if err != nil {
			return "", err
		}
		return r.renderAmass(ctx, value)
	case game.PrimitiveRenown:
		value, err := assertPrimitive[game.Renown](primitive)
		if err != nil {
			return "", err
		}
		return r.renderRenown(ctx, value)
	case game.PrimitiveAdapt:
		value, err := assertPrimitive[game.Adapt](primitive)
		if err != nil {
			return "", err
		}
		return r.renderAdapt(ctx, value)
	case game.PrimitiveMonstrosity:
		value, err := assertPrimitive[game.Monstrosity](primitive)
		if err != nil {
			return "", err
		}
		return r.renderMonstrosity(ctx, value)
	case game.PrimitiveConnive:
		value, err := assertPrimitive[game.Connive](primitive)
		if err != nil {
			return "", err
		}
		return r.renderConnive(ctx, value)
	case game.PrimitiveBecomeSaddled:
		value, err := assertPrimitive[game.BecomeSaddled](primitive)
		if err != nil {
			return "", err
		}
		return r.renderBecomeSaddled(ctx, value)
	case game.PrimitiveDig:
		value, err := assertPrimitive[game.Dig](primitive)
		if err != nil {
			return "", err
		}
		return r.renderDigPrimitive(ctx, value)
	case game.PrimitiveDestroy, game.PrimitiveBounce, game.PrimitiveUntap,
		game.PrimitiveTap, game.PrimitiveTapOrUntap, game.PrimitiveExile, game.PrimitivePhaseOut,
		game.PrimitiveRegenerate, game.PrimitiveSkipNextUntap, game.PrimitiveGoad:
		return r.renderObjectOrGroupPrimitive(ctx, primitive)
	case game.PrimitiveExplore,
		game.PrimitiveCounterObject, game.PrimitiveSacrifice,
		game.PrimitiveChooseNewTargets, game.PrimitiveRemoveFromCombat,
		game.PrimitiveTransform:
		return r.renderObjectPrimitive(primitive)
	case game.PrimitiveCopyStackObject:
		value, err := assertPrimitive[game.CopyStackObject](primitive)
		if err != nil {
			return "", err
		}
		return r.renderCopyStackObjectPrimitive(value)
	case game.PrimitiveAttach:
		return r.renderAttachPrimitive(primitive)
	case game.PrimitiveBecomeCopy:
		value, err := assertPrimitive[game.BecomeCopy](primitive)
		if err != nil {
			return "", err
		}
		return r.renderBecomeCopy(value)
	case game.PrimitiveSearch:
		value, err := assertPrimitive[game.Search](primitive)
		if err != nil {
			return "", err
		}
		return r.renderSearchPrimitive(ctx, value)
	case game.PrimitiveReveal:
		value, err := assertPrimitive[game.Reveal](primitive)
		if err != nil {
			return "", err
		}
		return r.renderRevealPrimitive(ctx, value)
	case game.PrimitiveExileEntireHand:
		value, err := assertPrimitive[game.ExileEntireHand](primitive)
		if err != nil {
			return "", err
		}
		return r.renderExileEntireHand(value)
	case game.PrimitiveReturnExiledCardsToHand:
		value, err := assertPrimitive[game.ReturnExiledCardsToHand](primitive)
		if err != nil {
			return "", err
		}
		return r.renderReturnExiledCardsToHand(value)
	case game.PrimitiveExileForEachPlayer:
		value, err := assertPrimitive[game.ExileForEachPlayer](primitive)
		if err != nil {
			return "", err
		}
		return r.renderExileForEachPlayer(ctx, value)
	case game.PrimitiveChampionExile:
		value, err := assertPrimitive[game.ChampionExile](primitive)
		if err != nil {
			return "", err
		}
		return r.renderChampionExile(ctx, value)
	case game.PrimitiveReturnLinkedExiledCardsToBattlefield:
		value, err := assertPrimitive[game.ReturnLinkedExiledCardsToBattlefield](primitive)
		if err != nil {
			return "", err
		}
		return r.renderReturnLinkedExiledCardsToBattlefield(ctx, value)
	case game.PrimitiveDestroyForEachPlayer:
		value, err := assertPrimitive[game.DestroyForEachPlayer](primitive)
		if err != nil {
			return "", err
		}
		return r.renderDestroyForEachPlayer(ctx, value)
	case game.PrimitiveEachPlayerChooseDestroy:
		value, err := assertPrimitive[game.EachPlayerChooseDestroy](primitive)
		if err != nil {
			return "", err
		}
		return r.renderEachPlayerChooseDestroy(ctx, value)
	case game.PrimitiveCreateTokenForEachDestroyed:
		value, err := assertPrimitive[game.CreateTokenForEachDestroyed](primitive)
		if err != nil {
			return "", err
		}
		return r.renderCreateTokenForEachDestroyed(ctx, value)
	case game.PrimitiveExileForEachOpponent:
		value, err := assertPrimitive[game.ExileForEachOpponent](primitive)
		if err != nil {
			return "", err
		}
		return r.renderExileForEachOpponent(ctx, value)
	case game.PrimitiveDrawForEachExiled:
		value, err := assertPrimitive[game.DrawForEachExiled](primitive)
		if err != nil {
			return "", err
		}
		return r.renderDrawForEachExiled(value)
	case game.PrimitiveRemoveTargetsForToken:
		value, err := assertPrimitive[game.RemoveTargetsForToken](primitive)
		if err != nil {
			return "", err
		}
		return r.renderRemoveTargetsForToken(value)
	case game.PrimitiveChooseFromZone:
		value, err := assertPrimitive[game.ChooseFromZone](primitive)
		if err != nil {
			return "", err
		}
		return r.renderChooseFromZone(ctx, value)
	case game.PrimitivePutHandOnLibraryThenDraw:
		return r.renderPutHandOnLibraryThenDraw(primitive)
	case game.PrimitiveDiscardThenDraw:
		return r.renderDiscardThenDraw(primitive)
	case game.PrimitiveDiscardUnlessType:
		return r.renderDiscardUnlessType(ctx, primitive)
	case game.PrimitiveCastForFree:
		value, err := assertPrimitive[game.CastForFree](primitive)
		if err != nil {
			return "", err
		}
		return r.renderCastForFree(ctx, value)
	case game.PrimitiveMassReturnFromGraveyard:
		value, err := assertPrimitive[game.MassReturnFromGraveyard](primitive)
		if err != nil {
			return "", err
		}
		return r.renderMassReturnFromGraveyard(ctx, value)
	case game.PrimitiveMassReanimationExchange:
		value, err := assertPrimitive[game.MassReanimationExchange](primitive)
		if err != nil {
			return "", err
		}
		return r.renderMassReanimationExchange(ctx, value)
	case game.PrimitiveShufflePermanentIntoLibrary:
		value, err := assertPrimitive[game.ShufflePermanentIntoLibrary](primitive)
		if err != nil {
			return "", err
		}
		return r.renderShufflePermanentIntoLibrary(value)
	case game.PrimitiveShuffleSpellIntoLibrary:
		if _, err := assertPrimitive[game.ShuffleSpellIntoLibrary](primitive); err != nil {
			return "", err
		}
		return "game.ShuffleSpellIntoLibrary{}", nil
	case game.PrimitivePutPermanentOnLibrary:
		value, err := assertPrimitive[game.PutPermanentOnLibrary](primitive)
		if err != nil {
			return "", err
		}
		return r.renderPutPermanentOnLibrary(value)
	case game.PrimitivePutLinkedExiledCardsInLibrary:
		value, err := assertPrimitive[game.PutLinkedExiledCardsInLibrary](primitive)
		if err != nil {
			return "", err
		}
		return renderPutLinkedExiledCardsInLibrary(value), nil
	case game.PrimitiveShuffleLibrary:
		value, err := assertPrimitive[game.ShuffleLibrary](primitive)
		if err != nil {
			return "", err
		}
		return r.renderShuffleLibrary(value)
	case game.PrimitiveShuffleGraveyardIntoLibrary:
		value, err := assertPrimitive[game.ShuffleGraveyardIntoLibrary](primitive)
		if err != nil {
			return "", err
		}
		return r.renderShuffleGraveyardIntoLibrary(value)
	case game.PrimitiveLookAtHand:
		value, err := assertPrimitive[game.LookAtHand](primitive)
		if err != nil {
			return "", err
		}
		return r.renderLookAtHand(value)
	case game.PrimitiveChooseDiscardFromHand:
		value, err := assertPrimitive[game.ChooseDiscardFromHand](primitive)
		if err != nil {
			return "", err
		}
		return r.renderChooseDiscardFromHand(ctx, value)
	case game.PrimitiveLookAtLibraryTop:
		value, err := assertPrimitive[game.LookAtLibraryTop](primitive)
		if err != nil {
			return "", err
		}
		return r.renderLookAtLibraryTop(value)
	case game.PrimitiveConditionalDestinationPlace:
		value, err := assertPrimitive[game.ConditionalDestinationPlace](primitive)
		if err != nil {
			return "", err
		}
		return r.renderConditionalDestinationPlace(ctx, value)
	default:
		return r.renderPrimitiveExtra(ctx, primitive)
	}
}

// renderPrimitiveExtra renders the remaining primitive kinds not handled by
// renderPrimitive. The dispatch is split across two functions purely to keep
// each function's size manageable; together they cover every primitive kind.
func (r Renderer) renderPrimitiveExtra(ctx *renderCtx, primitive game.Primitive) (string, error) {
	switch primitive.Kind() {
	case game.PrimitiveAddMana:
		value, err := assertPrimitive[game.AddMana](primitive)
		if err != nil {
			return "", err
		}
		return r.renderAddMana(ctx, &value)
	default:
		return r.renderPrimitiveTail(ctx, primitive)
	}
}

// renderPrimitiveTail handles the remaining primitive kinds not dispatched by
// renderPrimitive, keeping each switch small enough to stay maintainable.
func (r Renderer) renderPrimitiveTail(ctx *renderCtx, primitive game.Primitive) (string, error) {
	switch primitive.Kind() {
	case game.PrimitiveAddCounter:
		value, err := assertPrimitive[game.AddCounter](primitive)
		if err != nil {
			return "", err
		}
		return r.renderAddCounter(ctx, &value)
	case game.PrimitiveAddPlayerCounter:
		value, err := assertPrimitive[game.AddPlayerCounter](primitive)
		if err != nil {
			return "", err
		}
		return r.renderAddPlayerCounter(ctx, &value)
	case game.PrimitiveMoveCounters:
		value, err := assertPrimitive[game.MoveCounters](primitive)
		if err != nil {
			return "", err
		}
		return r.renderMoveCounters(ctx, &value)
	case game.PrimitiveRemoveCounter:
		value, err := assertPrimitive[game.RemoveCounter](primitive)
		if err != nil {
			return "", err
		}
		return r.renderRemoveCounter(ctx, &value)
	case game.PrimitiveModifyPT:
		value, err := assertPrimitive[game.ModifyPT](primitive)
		if err != nil {
			return "", err
		}
		return r.renderModifyPT(ctx, &value)
	case game.PrimitiveFight:
		return r.renderFightPrimitive(primitive)
	case game.PrimitiveChoose:
		value, err := assertPrimitive[game.Choose](primitive)
		if err != nil {
			return "", err
		}
		return r.renderChoose(ctx, value)
	case game.PrimitivePay:
		value, err := assertPrimitive[game.Pay](primitive)
		if err != nil {
			return "", err
		}
		return r.renderPay(ctx, value)
	case game.PrimitivePayRepeatedly:
		value, err := assertPrimitive[game.PayRepeatedly](primitive)
		if err != nil {
			return "", err
		}
		return r.renderPayRepeatedly(ctx, value)
	case game.PrimitivePutOnBattlefield:
		value, err := assertPrimitive[game.PutOnBattlefield](primitive)
		if err != nil {
			return "", err
		}
		return r.renderPutOnBattlefield(ctx, value)
	case game.PrimitiveMoveCard:
		value, err := assertPrimitive[game.MoveCard](primitive)
		if err != nil {
			return "", err
		}
		return r.renderMoveCard(ctx, value)
	case game.PrimitiveMoveCommander:
		value, err := assertPrimitive[game.MoveCommander](primitive)
		if err != nil {
			return "", err
		}
		return r.renderMoveCommander(ctx, value)
	case game.PrimitiveGrantCastPermission:
		value, err := assertPrimitive[game.GrantCastPermission](primitive)
		if err != nil {
			return "", err
		}
		return r.renderGrantCastPermission(ctx, value)
	case game.PrimitiveExileForPlay:
		value, err := assertPrimitive[game.ExileForPlay](primitive)
		if err != nil {
			return "", err
		}
		return r.renderExileForPlay(ctx, value)
	case game.PrimitiveExilePermanentForPlay:
		value, err := assertPrimitive[game.ExilePermanentForPlay](primitive)
		if err != nil {
			return "", err
		}
		return r.renderExilePermanentForPlay(value)
	case game.PrimitivePlayChosenExiledCard:
		value, err := assertPrimitive[game.PlayChosenExiledCard](primitive)
		if err != nil {
			return "", err
		}
		return r.renderPlayChosenExiledCard(ctx, value)
	case game.PrimitiveImpulseExile:
		value, err := assertPrimitive[game.ImpulseExile](primitive)
		if err != nil {
			return "", err
		}
		return r.renderImpulseExile(ctx, value)
	case game.PrimitiveExileLibraryUntilNonlandCast:
		value, err := assertPrimitive[game.ExileLibraryUntilNonlandCast](primitive)
		if err != nil {
			return "", err
		}
		return r.renderExileLibraryUntilNonlandCast(value)
	case game.PrimitiveHideawayExile:
		value, err := assertPrimitive[game.HideawayExile](primitive)
		if err != nil {
			return "", err
		}
		amount, err := r.renderQuantity(ctx, value.Amount)
		if err != nil {
			return "", err
		}
		return structLit("game.HideawayExile", []string{
			fmt.Sprintf("Amount: %s,", amount),
		}), nil
	case game.PrimitivePlayHideawayCard:
		if _, err := assertPrimitive[game.PlayHideawayCard](primitive); err != nil {
			return "", err
		}
		return "game.PlayHideawayCard{}", nil
	case game.PrimitiveCreateDelayedTrigger:
		value, err := assertPrimitive[game.CreateDelayedTrigger](primitive)
		if err != nil {
			return "", err
		}
		return r.renderCreateDelayedTrigger(ctx, value)
	case game.PrimitiveCreateReflexiveTrigger:
		value, err := assertPrimitive[game.CreateReflexiveTrigger](primitive)
		if err != nil {
			return "", err
		}
		return r.renderCreateReflexiveTrigger(ctx, value)
	case game.PrimitiveCreateReplacement:
		value, err := assertPrimitive[game.CreateReplacement](primitive)
		if err != nil {
			return "", err
		}
		return r.renderCreateReplacement(ctx, value)
	case game.PrimitiveCreateEmblem:
		value, err := assertPrimitive[game.CreateEmblem](primitive)
		if err != nil {
			return "", err
		}
		return r.renderCreateEmblem(ctx, value)
	case game.PrimitiveApplyContinuous:
		value, err := assertPrimitive[game.ApplyContinuous](primitive)
		if err != nil {
			return "", err
		}
		return r.renderApplyContinuousPrimitive(ctx, value)
	case game.PrimitiveApplyRule:
		value, err := assertPrimitive[game.ApplyRule](primitive)
		if err != nil {
			return "", err
		}
		return r.renderApplyRulePrimitive(ctx, value)
	case game.PrimitivePlayerMayPayGenericOrRule:
		value, err := assertPrimitive[game.PlayerMayPayGenericOrRule](primitive)
		if err != nil {
			return "", err
		}
		return r.renderPlayerMayPayGenericOrRule(ctx, value)
	case game.PrimitiveSacrificePermanents:
		value, err := assertPrimitive[game.SacrificePermanents](primitive)
		if err != nil {
			return "", err
		}
		return r.renderSacrificePermanents(ctx, &value)
	case game.PrimitiveRevealUntil:
		value, err := assertPrimitive[game.RevealUntil](primitive)
		if err != nil {
			return "", err
		}
		return r.renderRevealUntil(ctx, &value)
	case game.PrimitivePileSplit:
		value, err := assertPrimitive[game.PileSplit](primitive)
		if err != nil {
			return "", err
		}
		return r.renderPileSplit(ctx, &value)
	case game.PrimitiveRevealTopPartition:
		value, err := assertPrimitive[game.RevealTopPartition](primitive)
		if err != nil {
			return "", err
		}
		return r.renderRevealTopPartition(ctx, &value)
	case game.PrimitivePunisherEachLoseLife:
		value, err := assertPrimitive[game.PunisherEachLoseLife](primitive)
		if err != nil {
			return "", err
		}
		return r.renderPunisherEachLoseLife(ctx, &value)
	case game.PrimitiveRepeatProcess:
		value, err := assertPrimitive[game.RepeatProcess](primitive)
		if err != nil {
			return "", err
		}
		return r.renderRepeatProcess(ctx, &value)
	case game.PrimitiveCreateToken:
		value, err := assertPrimitive[game.CreateToken](primitive)
		if err != nil {
			return "", err
		}
		return r.renderCreateToken(ctx, value)
	case game.PrimitivePreventDamage:
		value, err := assertPrimitive[game.PreventDamage](primitive)
		if err != nil {
			return "", err
		}
		return r.renderPreventDamage(ctx, value)
	case game.PrimitiveAddExtraPhases:
		value, err := assertPrimitive[game.AddExtraPhases](primitive)
		if err != nil {
			return "", err
		}
		return renderAddExtraPhases(value), nil
	case game.PrimitiveRollDie:
		value, err := assertPrimitive[game.RollDie](primitive)
		if err != nil {
			return "", err
		}
		return renderRollDie(value), nil
	case game.PrimitiveSetClassLevel:
		value, err := assertPrimitive[game.SetClassLevel](primitive)
		if err != nil {
			return "", err
		}
		return r.renderSetClassLevel(ctx, value)
	case game.PrimitiveBecomeMonarch:
		value, err := assertPrimitive[game.BecomeMonarch](primitive)
		if err != nil {
			return "", err
		}
		return r.renderBecomeMonarch(value)
	case game.PrimitiveCantBecomeMonarch:
		value, err := assertPrimitive[game.CantBecomeMonarch](primitive)
		if err != nil {
			return "", err
		}
		return r.renderCantBecomeMonarch(value)
	case game.PrimitiveRingTempts:
		value, err := assertPrimitive[game.RingTempts](primitive)
		if err != nil {
			return "", err
		}
		return r.renderRingTempts(value)
	case game.PrimitiveVote:
		value, err := assertPrimitive[game.Vote](primitive)
		if err != nil {
			return "", err
		}
		return renderVote(value), nil
	case game.PrimitivePartitionExiledCostCards:
		value, err := assertPrimitive[game.PartitionExiledCostCards](primitive)
		if err != nil {
			return "", err
		}
		return renderPartitionExiledCostCards(value), nil
	default:
		return "", fmt.Errorf("render: unsupported primitive kind %d", primitive.Kind())
	}
}

// renderAddExtraPhases renders the AddExtraPhases primitive, emitting only the
// set phase flags so the literal matches the typed effect.
func renderAddExtraPhases(value game.AddExtraPhases) string {
	var fields []string
	if value.Combat {
		fields = append(fields, "Combat: true,")
	}
	if value.Main {
		fields = append(fields, "Main: true,")
	}
	if value.Beginning {
		fields = append(fields, "Beginning: true,")
	}
	return structLit("game.AddExtraPhases", fields)
}

// renderRollDie renders the RollDie primitive, emitting its die size so the
// literal matches the typed effect.
func renderRollDie(value game.RollDie) string {
	return structLit("game.RollDie", []string{fmt.Sprintf("Sides: %d,", value.Sides)})
}

// renderVote renders the Vote primitive, emitting its named option labels so the
// literal matches the typed effect.
func renderVote(value game.Vote) string {
	options := make([]string, len(value.Options))
	for i, option := range value.Options {
		options[i] = fmt.Sprintf("%q,", option)
	}
	return structLit("game.Vote", []string{sliceField("Options", "string", options)})
}

func renderPartitionExiledCostCards(value game.PartitionExiledCostCards) string {
	var fields []string
	if value.ChooserOpponent {
		fields = append(fields, "ChooserOpponent: true,")
	}
	if value.ChosenToLibraryBottom {
		fields = append(fields, "ChosenToLibraryBottom: true,")
	}
	if value.OtherEntersTapped {
		fields = append(fields, "OtherEntersTapped: true,")
	}
	return structLit("game.PartitionExiledCostCards", fields)
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

func (r Renderer) renderRevealPrimitive(ctx *renderCtx, value game.Reveal) (string, error) {
	if value.Card.Kind != game.CardReferenceNone {
		card, err := renderCardReference(value.Card)
		if err != nil {
			return "", err
		}
		return structLit("game.Reveal", []string{fmt.Sprintf("Card: %s,", card)}), nil
	}
	amount, err := r.renderQuantity(ctx, value.Amount)
	if err != nil {
		return "", err
	}
	player, err := r.renderPlayerReference(value.Player)
	if err != nil {
		return "", err
	}
	fields := []string{
		fmt.Sprintf("Amount: %s,", amount),
		fmt.Sprintf("Player: %s,", player),
	}
	if value.Recipient.Exists {
		recipient, err := r.renderPlayerReference(value.Recipient.Val)
		if err != nil {
			return "", err
		}
		ctx.need(importOpt)
		fields = append(fields, fmt.Sprintf("Recipient: opt.Val(%s),", recipient))
	}
	if value.PublishLinked != "" {
		fields = append(fields, fmt.Sprintf("PublishLinked: game.LinkedKey(%q),", string(value.PublishLinked)))
	}
	return structLit("game.Reveal", fields), nil
}

func (r Renderer) renderExileEntireHand(value game.ExileEntireHand) (string, error) {
	player, err := r.renderPlayerReference(value.Player)
	if err != nil {
		return "", err
	}
	return structLit("game.ExileEntireHand", []string{
		fmt.Sprintf("Player: %s,", player),
		fmt.Sprintf("LinkedKey: game.LinkedKey(%q),", string(value.LinkedKey)),
	}), nil
}

func (Renderer) renderReturnExiledCardsToHand(value game.ReturnExiledCardsToHand) (string, error) {
	return structLit("game.ReturnExiledCardsToHand", []string{
		fmt.Sprintf("LinkedKey: game.LinkedKey(%q),", string(value.LinkedKey)),
	}), nil
}

func (r Renderer) renderExileForEachPlayer(ctx *renderCtx, value game.ExileForEachPlayer) (string, error) {
	chooser, err := r.renderPlayerReference(value.Chooser)
	if err != nil {
		return "", err
	}
	selection, err := r.renderSelection(ctx, value.Selection)
	if err != nil {
		return "", err
	}
	return structLit("game.ExileForEachPlayer", []string{
		fmt.Sprintf("Chooser: %s,", chooser),
		fmt.Sprintf("Selection: %s,", selection),
		fmt.Sprintf("LinkedKey: game.LinkedKey(%q),", string(value.LinkedKey)),
	}), nil
}

func (r Renderer) renderChampionExile(ctx *renderCtx, value game.ChampionExile) (string, error) {
	selection, err := r.renderSelection(ctx, value.Selection)
	if err != nil {
		return "", err
	}
	return structLit("game.ChampionExile", []string{
		fmt.Sprintf("Selection: %s,", selection),
		fmt.Sprintf("LinkedKey: game.LinkedKey(%q),", string(value.LinkedKey)),
	}), nil
}

func (r Renderer) renderReturnLinkedExiledCardsToBattlefield(ctx *renderCtx, value game.ReturnLinkedExiledCardsToBattlefield) (string, error) {
	chooser, err := r.renderPlayerReference(value.Chooser)
	if err != nil {
		return "", err
	}
	amount, err := r.renderQuantity(ctx, value.Amount)
	if err != nil {
		return "", err
	}
	fields := []string{
		fmt.Sprintf("Chooser: %s,", chooser),
		fmt.Sprintf("LinkedKey: game.LinkedKey(%q),", string(value.LinkedKey)),
		fmt.Sprintf("Amount: %s,", amount),
	}
	if value.RestToLibraryBottom {
		fields = append(fields, "RestToLibraryBottom: true,")
	}
	return structLit("game.ReturnLinkedExiledCardsToBattlefield", fields), nil
}

func (r Renderer) renderDestroyForEachPlayer(ctx *renderCtx, value game.DestroyForEachPlayer) (string, error) {
	chooser, err := r.renderPlayerReference(value.Chooser)
	if err != nil {
		return "", err
	}
	selection, err := r.renderSelection(ctx, value.Selection)
	if err != nil {
		return "", err
	}
	return structLit("game.DestroyForEachPlayer", []string{
		fmt.Sprintf("Chooser: %s,", chooser),
		fmt.Sprintf("Selection: %s,", selection),
		fmt.Sprintf("LinkedKey: game.LinkedKey(%q),", string(value.LinkedKey)),
	}), nil
}

func (r Renderer) renderEachPlayerChooseDestroy(ctx *renderCtx, value game.EachPlayerChooseDestroy) (string, error) {
	selection, err := r.renderSelection(ctx, value.Selection)
	if err != nil {
		return "", err
	}
	fields := []string{fmt.Sprintf("Selection: %s,", selection)}
	if value.Optional {
		fields = append(fields, "Optional: true,")
	}
	if value.PreventRegeneration {
		fields = append(fields, "PreventRegeneration: true,")
	}
	return structLit("game.EachPlayerChooseDestroy", fields), nil
}

func (r Renderer) renderCreateTokenForEachDestroyed(ctx *renderCtx, value game.CreateTokenForEachDestroyed) (string, error) {
	source, err := r.renderTokenSource(ctx, value.Source)
	if err != nil {
		return "", err
	}
	return structLit("game.CreateTokenForEachDestroyed", []string{
		fmt.Sprintf("Source: %s,", source),
		fmt.Sprintf("LinkedKey: game.LinkedKey(%q),", string(value.LinkedKey)),
	}), nil
}

func (r Renderer) renderExileForEachOpponent(ctx *renderCtx, value game.ExileForEachOpponent) (string, error) {
	chooser, err := r.renderPlayerReference(value.Chooser)
	if err != nil {
		return "", err
	}
	selection, err := r.renderSelection(ctx, value.Selection)
	if err != nil {
		return "", err
	}
	return structLit("game.ExileForEachOpponent", []string{
		fmt.Sprintf("Chooser: %s,", chooser),
		fmt.Sprintf("Selection: %s,", selection),
		fmt.Sprintf("LinkedKey: game.LinkedKey(%q),", string(value.LinkedKey)),
	}), nil
}

func (Renderer) renderDrawForEachExiled(value game.DrawForEachExiled) (string, error) {
	return structLit("game.DrawForEachExiled", []string{
		fmt.Sprintf("LinkedKey: game.LinkedKey(%q),", string(value.LinkedKey)),
	}), nil
}

func (Renderer) renderRemoveTargetsForToken(value game.RemoveTargetsForToken) (string, error) {
	fields := []string{}
	if value.Exile {
		fields = append(fields, "Exile: true,")
	}
	if value.PreventRegeneration {
		fields = append(fields, "PreventRegeneration: true,")
	}
	fields = append(fields, fmt.Sprintf("LinkedKey: game.LinkedKey(%q),", string(value.LinkedKey)))
	return structLit("game.RemoveTargetsForToken", fields), nil
}

// renderChooseFromZone renders the canonical game.ChooseFromZone envelope, the
// single primitive every choose-from-zone family (exile from hand/graveyard, put
// from hand, return from graveyard) now lowers to. It emits every field the
// lowerers can set so the rendered literal reproduces the envelope exactly. Only
// non-zero fields are emitted, matching the zero-value semantics of the runtime.
func (r Renderer) renderChooseFromZone(ctx *renderCtx, value game.ChooseFromZone) (string, error) {
	ctx.need(importZone)
	player, err := r.renderPlayerReference(value.Player)
	if err != nil {
		return "", err
	}
	selection, err := r.renderSelection(ctx, value.Filter)
	if err != nil {
		return "", err
	}
	sourceZone, err := renderZone(value.SourceZone)
	if err != nil {
		return "", err
	}
	destZone, err := renderZone(value.Destination.Zone)
	if err != nil {
		return "", err
	}
	fields := []string{
		fmt.Sprintf("Player: %s,", player),
		fmt.Sprintf("SourceZone: %s,", sourceZone),
	}
	if value.AllOwners {
		fields = append(fields, "AllOwners: true,")
	}
	fields = append(fields, fmt.Sprintf("Filter: %s,", selection))
	if value.Count != game.ChooseAnyNumber {
		amount, err := r.renderQuantity(ctx, value.Quantity)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("Quantity: %s,", amount))
	}
	switch value.Count {
	case game.ChooseExactly:
	case game.ChooseUpTo:
		fields = append(fields, "Count: game.ChooseUpTo,")
	case game.ChooseAnyNumber:
		fields = append(fields, "Count: game.ChooseAnyNumber,")
	default:
		return "", fmt.Errorf("render: unsupported ChooseFromZone count %d", value.Count)
	}
	destination, err := r.renderChooseDestination(value.Destination, destZone)
	if err != nil {
		return "", err
	}
	fields = append(fields, fmt.Sprintf("Destination: %s,", destination))
	riders, ok, err := r.renderChooseRiders(ctx, value.Riders)
	if err != nil {
		return "", err
	}
	if ok {
		fields = append(fields, fmt.Sprintf("Riders: %s,", riders))
	}
	if value.Prompt != "" {
		fields = append(fields, fmt.Sprintf("Prompt: %q,", value.Prompt))
	}
	return structLit("game.ChooseFromZone", fields), nil
}

// renderChooseDestination renders a game.ChooseDestination given its
// already-rendered zone literal.
func (Renderer) renderChooseDestination(destination game.ChooseDestination, destZone string) (string, error) {
	fields := []string{fmt.Sprintf("Zone: %s,", destZone)}
	switch destination.Position {
	case game.ChoosePositionDefault:
	case game.ChoosePositionTop:
		fields = append(fields, "Position: game.ChoosePositionTop,")
	default:
		return "", fmt.Errorf("render: unsupported ChooseFromZone destination position %d", destination.Position)
	}
	return structLit("game.ChooseDestination", fields), nil
}

// renderChooseRiders renders a game.ChooseRiders, returning ok=false when no
// rider is set so the caller can omit the field entirely.
func (Renderer) renderChooseRiders(ctx *renderCtx, riders game.ChooseRiders) (string, bool, error) {
	var fields []string
	if riders.EntersTapped {
		fields = append(fields, "EntersTapped: true,")
	}
	if riders.EntersAttacking {
		fields = append(fields, "EntersAttacking: true,")
	}
	if riders.UnderOwnerControl {
		fields = append(fields, "UnderOwnerControl: true,")
	}
	if riders.DestinationBottom {
		fields = append(fields, "DestinationBottom: true,")
	}
	if riders.Reveal {
		fields = append(fields, "Reveal: true,")
	}
	if riders.FromLinked != "" {
		fields = append(fields, fmt.Sprintf("FromLinked: game.LinkedKey(%q),", string(riders.FromLinked)))
	}
	if riders.PublishLinked != "" {
		fields = append(fields, fmt.Sprintf("PublishLinked: game.LinkedKey(%q),", string(riders.PublishLinked)))
	}
	if riders.PublishObjectScoped {
		fields = append(fields, "PublishObjectScoped: true,")
	}
	if riders.MaxTotalManaValue.Exists {
		ctx.need(importOpt)
		fields = append(fields, fmt.Sprintf("MaxTotalManaValue: opt.Val(%d),", riders.MaxTotalManaValue.Val))
	}
	if riders.MaxManaValueFromX {
		fields = append(fields, "MaxManaValueFromX: true,")
	}
	if len(fields) == 0 {
		return "", false, nil
	}
	return structLit("game.ChooseRiders", fields), true, nil
}

func (r Renderer) renderPutHandOnLibraryThenDraw(primitive game.Primitive) (string, error) {
	value, err := assertPrimitive[game.PutHandOnLibraryThenDraw](primitive)
	if err != nil {
		return "", err
	}
	player, err := r.renderPlayerReference(value.Player)
	if err != nil {
		return "", err
	}
	fields := []string{
		fmt.Sprintf("Player: %s,", player),
	}
	if value.Bottom {
		fields = append(fields, "Bottom: true,")
	}
	if value.DrawOffset != 0 {
		fields = append(fields, fmt.Sprintf("DrawOffset: %d,", value.DrawOffset))
	}
	return structLit("game.PutHandOnLibraryThenDraw", fields), nil
}

func (r Renderer) renderDiscardThenDraw(primitive game.Primitive) (string, error) {
	value, err := assertPrimitive[game.DiscardThenDraw](primitive)
	if err != nil {
		return "", err
	}
	player, err := r.renderPlayerReference(value.Player)
	if err != nil {
		return "", err
	}
	fields := []string{
		fmt.Sprintf("Player: %s,", player),
	}
	if value.Max != 0 {
		fields = append(fields, fmt.Sprintf("Max: %d,", value.Max))
	}
	if value.DrawOffset != 0 {
		fields = append(fields, fmt.Sprintf("DrawOffset: %d,", value.DrawOffset))
	}
	return structLit("game.DiscardThenDraw", fields), nil
}

func (r Renderer) renderDiscardUnlessType(ctx *renderCtx, primitive game.Primitive) (string, error) {
	value, err := assertPrimitive[game.DiscardUnlessType](primitive)
	if err != nil {
		return "", err
	}
	player, err := r.renderPlayerReference(value.Player)
	if err != nil {
		return "", err
	}
	exempt, err := renderTypesCardSlice(ctx, value.ExemptTypes)
	if err != nil {
		return "", err
	}
	fields := []string{
		fmt.Sprintf("Player: %s,", player),
		fmt.Sprintf("Amount: %d,", value.Amount),
		fmt.Sprintf("ExemptTypes: %s,", exempt),
	}
	return structLit("game.DiscardUnlessType", fields), nil
}

func (r Renderer) renderCastForFree(ctx *renderCtx, value game.CastForFree) (string, error) {
	ctx.need(importZone)
	player, err := r.renderPlayerReference(value.Player)
	if err != nil {
		return "", err
	}
	sourceZone, err := renderZone(value.Zone)
	if err != nil {
		return "", err
	}
	fields := []string{fmt.Sprintf("Player: %s,", player)}
	if value.Card.Kind == game.CardReferenceNone {
		selection, err := r.renderSelection(ctx, value.Selection)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("Selection: %s,", selection))
	}
	fields = append(fields, fmt.Sprintf("Zone: %s,", sourceZone))
	if value.Card.Kind != game.CardReferenceNone {
		card, err := renderCardReference(value.Card)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("Card: %s,", card))
	}
	if value.ExileOnResolution {
		fields = append(fields, "ExileOnResolution: true,")
	}
	return structLit("game.CastForFree", fields), nil
}

func (r Renderer) renderMassReturnFromGraveyard(ctx *renderCtx, value game.MassReturnFromGraveyard) (string, error) {
	player, err := r.renderPlayerReference(value.Player)
	if err != nil {
		return "", err
	}
	selection, err := r.renderSelection(ctx, value.Selection)
	if err != nil {
		return "", err
	}
	destination, err := renderZone(value.Destination)
	if err != nil {
		return "", err
	}
	ctx.need(importZone)
	fields := []string{
		fmt.Sprintf("Player: %s,", player),
		fmt.Sprintf("Selection: %s,", selection),
		fmt.Sprintf("Destination: %s,", destination),
	}
	if value.EntryTapped {
		fields = append(fields, "EntryTapped: true,")
	}
	if value.SourceGroup.Kind != game.PlayerGroupReferenceNone {
		var group string
		switch value.SourceGroup.Kind {
		case game.PlayerGroupReferenceOpponents:
			group = "game.OpponentsReference()"
		case game.PlayerGroupReferenceAllPlayers:
			group = "game.AllPlayersReference()"
		default:
			return "", fmt.Errorf("render: unsupported player group reference kind %d", value.SourceGroup.Kind)
		}
		fields = append(fields, fmt.Sprintf("SourceGroup: %s,", group))
	}
	if value.ControlledByOwner {
		fields = append(fields, "ControlledByOwner: true,")
	}
	if value.FromTriggerBatch {
		fields = append(fields, "FromTriggerBatch: true,")
	}
	return structLit("game.MassReturnFromGraveyard", fields), nil
}

func (r Renderer) renderMassReanimationExchange(ctx *renderCtx, value game.MassReanimationExchange) (string, error) {
	selection, err := r.renderSelection(ctx, value.Selection)
	if err != nil {
		return "", err
	}
	fields := []string{
		fmt.Sprintf("Selection: %s,", selection),
	}
	return structLit("game.MassReanimationExchange", fields), nil
}

func (r Renderer) renderSetClassLevel(ctx *renderCtx, value game.SetClassLevel) (string, error) {
	object, err := r.renderObjectReference(value.Object)
	if err != nil {
		return "", err
	}
	amount, err := r.renderQuantity(ctx, value.Amount)
	if err != nil {
		return "", err
	}
	return structLit("game.SetClassLevel", []string{
		fmt.Sprintf("Object: %s,", object),
		fmt.Sprintf("Amount: %s,", amount),
	}), nil
}

func (r Renderer) renderShufflePermanentIntoLibrary(value game.ShufflePermanentIntoLibrary) (string, error) {
	object, err := r.renderObjectReference(value.Object)
	if err != nil {
		return "", err
	}
	return structLit("game.ShufflePermanentIntoLibrary", []string{
		fmt.Sprintf("Object: %s,", object),
	}), nil
}

func (r Renderer) renderPutPermanentOnLibrary(value game.PutPermanentOnLibrary) (string, error) {
	object, err := r.renderObjectReference(value.Object)
	if err != nil {
		return "", err
	}
	fields := []string{fmt.Sprintf("Object: %s,", object)}
	if value.Bottom {
		fields = append(fields, "Bottom: true,")
	}
	return structLit("game.PutPermanentOnLibrary", fields), nil
}

// renderPutLinkedExiledCardsInLibrary renders the linked disposal primitive,
// emitting the consumed link key and the bottom flag when set so the literal
// matches the typed effect.
func renderPutLinkedExiledCardsInLibrary(value game.PutLinkedExiledCardsInLibrary) string {
	fields := []string{fmt.Sprintf("LinkedKey: game.LinkedKey(%q),", string(value.LinkedKey))}
	if value.Bottom {
		fields = append(fields, "Bottom: true,")
	}
	return structLit("game.PutLinkedExiledCardsInLibrary", fields)
}

func (r Renderer) renderSearchPrimitive(ctx *renderCtx, value game.Search) (string, error) {
	amount, err := r.renderQuantity(ctx, value.Amount)
	if err != nil {
		return "", err
	}
	source, err := renderZone(value.Spec.SourceZone)
	if err != nil {
		return "", err
	}
	ctx.need(importZone)
	specFields := []string{
		fmt.Sprintf("SourceZone: %s,", source),
	}
	if value.Spec.Destination != zone.None {
		destination, destErr := renderZone(value.Spec.Destination)
		if destErr != nil {
			return "", destErr
		}
		specFields = append(specFields, fmt.Sprintf("Destination: %s,", destination))
	}
	if value.Spec.DestinationPosition == game.SearchPositionTop {
		specFields = append(specFields, "DestinationPosition: game.SearchPositionTop,")
	}
	switch value.Spec.FailToFindPolicy {
	case game.SearchFailToFindDefault:
	case game.SearchMayFailToFind:
		specFields = append(specFields, "FailToFindPolicy: game.SearchMayFailToFind,")
	case game.SearchMustFindIfAvailable:
		specFields = append(specFields, "FailToFindPolicy: game.SearchMustFindIfAvailable,")
	default:
		return "", errors.New("render: unsupported search fail-to-find policy")
	}
	if !value.Spec.Filter.Empty() {
		filter, err := r.renderSelection(ctx, value.Spec.Filter)
		if err != nil {
			return "", err
		}
		specFields = append(specFields, fmt.Sprintf("Filter: %s,", filter))
	}
	if value.Spec.MaxManaValueFromX {
		specFields = append(specFields, "MaxManaValueFromX: true,")
	}
	if value.Spec.Name != "" {
		specFields = append(specFields, fmt.Sprintf("Name: %q,", value.Spec.Name))
	}
	if value.Spec.Reveal {
		specFields = append(specFields, "Reveal: true,")
	}
	if value.Spec.AlsoGraveyard {
		specFields = append(specFields, "AlsoGraveyard: true,")
	}
	if value.Spec.RevealOnly {
		specFields = append(specFields, "RevealOnly: true,")
	}
	if value.Spec.EntersTapped {
		specFields = append(specFields, "EntersTapped: true,")
	}
	if value.Spec.SharedSubtype {
		specFields = append(specFields, "SharedSubtype: true,")
	}
	if value.Spec.DifferentNames {
		specFields = append(specFields, "DifferentNames: true,")
	}
	if len(value.Spec.SlotFilters) != 0 {
		slotLits := make([]string, 0, len(value.Spec.SlotFilters))
		for _, slot := range value.Spec.SlotFilters {
			slotLit, slotErr := r.renderSelection(ctx, slot)
			if slotErr != nil {
				return "", slotErr
			}
			slotLits = append(slotLits, slotLit+",")
		}
		specFields = append(specFields, fmt.Sprintf("SlotFilters: %s,", sliceLit("game.Selection", slotLits)))
	}
	if value.Spec.SplitDestination.Exists {
		splitZone, err := renderZone(value.Spec.SplitDestination.Val.Zone)
		if err != nil {
			return "", err
		}
		ctx.need(importOpt)
		ctx.need(importZone)
		splitFields := []string{fmt.Sprintf("Zone: %s,", splitZone)}
		if value.Spec.SplitDestination.Val.Position == game.SearchPositionTop {
			splitFields = append(splitFields, "Position: game.SearchPositionTop,")
		}
		if value.Spec.SplitDestination.Val.EntersTapped {
			splitFields = append(splitFields, "EntersTapped: true,")
		}
		specFields = append(specFields, fmt.Sprintf("SplitDestination: opt.Val(%s),", structLit("game.SearchDestination", splitFields)))
	}
	fields := []string{
		fmt.Sprintf("Spec: %s,", structLit("game.SearchSpec", specFields)),
		fmt.Sprintf("Amount: %s,", amount),
	}
	if value.PlayerGroup.Kind != game.PlayerGroupReferenceNone {
		var group string
		switch value.PlayerGroup.Kind {
		case game.PlayerGroupReferenceOpponents:
			group = "game.OpponentsReference()"
		case game.PlayerGroupReferenceAllPlayers:
			group = "game.AllPlayersReference()"
		default:
			return "", fmt.Errorf("render: unsupported player group reference kind %d", value.PlayerGroup.Kind)
		}
		fields = append([]string{fmt.Sprintf("PlayerGroup: %s,", group)}, fields...)
	} else {
		player, err := r.renderPlayerReference(value.Player)
		if err != nil {
			return "", err
		}
		fields = append([]string{fmt.Sprintf("Player: %s,", player)}, fields...)
	}
	if value.Controller.Exists {
		controller, err := r.renderPlayerReference(value.Controller.Val)
		if err != nil {
			return "", err
		}
		ctx.need(importOpt)
		fields = append(fields, fmt.Sprintf("Controller: opt.Val(%s),", controller))
	}
	if value.PublishLinked != "" {
		fields = append(fields, fmt.Sprintf("PublishLinked: game.LinkedKey(%q),", value.PublishLinked))
	}
	return structLit("game.Search", fields), nil
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
	if value.ChooseFrom.Valid() {
		group, err := r.renderGroupReference(ctx, value.ChooseFrom)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("ChooseFrom: %s,", group))
		amount, err := r.renderQuantity(ctx, value.ChooseUpTo)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("ChooseUpTo: %s,", amount))
	}
	if value.Prompt != "" {
		fields = append(fields, fmt.Sprintf("Prompt: %q,", value.Prompt))
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

// renderPlayerMayPayGenericOrRule renders a PlayerMayPayGenericOrRule
// instruction: the payer, the generic mana amount, the rule effects installed on
// non-payment, and their duration.
func (r Renderer) renderPlayerMayPayGenericOrRule(ctx *renderCtx, value game.PlayerMayPayGenericOrRule) (string, error) {
	player, err := r.renderPlayerReference(value.Player)
	if err != nil {
		return "", err
	}
	amount, err := r.renderQuantity(ctx, value.Amount)
	if err != nil {
		return "", err
	}
	fields := []string{
		fmt.Sprintf("Player: %s,", player),
		fmt.Sprintf("Amount: %s,", amount),
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
	return structLit("game.PlayerMayPayGenericOrRule", fields), nil
}
