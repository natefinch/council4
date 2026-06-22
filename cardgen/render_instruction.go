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
	fields := []string{sliceField("Modes", "game.Mode", modeElements)}
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
	return structLit("game.AbilityContent", fields), nil
}

func (r Renderer) renderMode(ctx *renderCtx, mode game.Mode) (string, error) {
	var fields []string
	if mode.Text != "" {
		fields = append(fields, fmt.Sprintf("Text: %q,", mode.Text))
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
		condition, err := r.renderCardCondition(ctx, instruction.CardCondition.Val)
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
	case game.PrimitiveGroupSourceDamage:
		return r.renderGroupSourceDamage(ctx, primitive)
	case game.PrimitiveDraw, game.PrimitiveDiscard, game.PrimitiveMill,
		game.PrimitiveScry, game.PrimitiveSurveil, game.PrimitiveGainLife,
		game.PrimitiveLoseLife, game.PrimitiveReorderLibraryTop,
		game.PrimitiveExileTopOfLibrary:
		return r.renderPlayerAmountPrimitive(ctx, primitive)
	case game.PrimitivePlayerLosesGame:
		value, ok := primitive.(game.PlayerLosesGame)
		if !ok {
			return "", errors.New("render: internal error: PlayerLosesGame kind has unexpected concrete type")
		}
		return r.renderPlayerLosesGame(value)
	case game.PrimitivePlayerWinsGame:
		value, ok := primitive.(game.PlayerWinsGame)
		if !ok {
			return "", errors.New("render: internal error: PlayerWinsGame kind has unexpected concrete type")
		}
		return r.renderPlayerWinsGame(value)
	case game.PrimitiveInvestigate, game.PrimitiveProliferate, game.PrimitiveManifest:
		return r.renderStandalonePrimitive(ctx, primitive)
	case game.PrimitiveAmass:
		value, ok := primitive.(game.Amass)
		if !ok {
			return "", errors.New("render: internal error: Amass kind has unexpected concrete type")
		}
		return r.renderAmass(ctx, value)
	case game.PrimitiveRenown:
		value, ok := primitive.(game.Renown)
		if !ok {
			return "", errors.New("render: internal error: Renown kind has unexpected concrete type")
		}
		return r.renderRenown(ctx, value)
	case game.PrimitiveBecomeSaddled:
		value, ok := primitive.(game.BecomeSaddled)
		if !ok {
			return "", errors.New("render: internal error: BecomeSaddled kind has unexpected concrete type")
		}
		return r.renderBecomeSaddled(ctx, value)
	case game.PrimitiveDig:
		value, ok := primitive.(game.Dig)
		if !ok {
			return "", errors.New("render: internal error: Dig kind has unexpected concrete type")
		}
		return r.renderDigPrimitive(ctx, value)
	case game.PrimitiveDestroy, game.PrimitiveBounce, game.PrimitiveUntap,
		game.PrimitiveTap, game.PrimitiveExile, game.PrimitivePhaseOut:
		return r.renderObjectOrGroupPrimitive(ctx, primitive)
	case game.PrimitiveRegenerate, game.PrimitiveExplore,
		game.PrimitiveCounterObject, game.PrimitiveSacrifice, game.PrimitiveSkipNextUntap,
		game.PrimitiveChooseNewTargets:
		return r.renderObjectPrimitive(primitive)
	case game.PrimitiveCopyStackObject:
		value, ok := primitive.(game.CopyStackObject)
		if !ok {
			return "", errors.New("render: internal error: CopyStackObject kind has unexpected concrete type")
		}
		return r.renderCopyStackObjectPrimitive(value)
	case game.PrimitiveAttach:
		return r.renderAttachPrimitive(primitive)
	case game.PrimitiveBecomeCopy:
		value, ok := primitive.(game.BecomeCopy)
		if !ok {
			return "", errors.New("render: internal error: BecomeCopy kind has unexpected concrete type")
		}
		return r.renderBecomeCopy(value)
	case game.PrimitiveSearch:
		value, ok := primitive.(game.Search)
		if !ok {
			return "", errors.New("render: internal error: Search kind has unexpected concrete type")
		}
		return r.renderSearchPrimitive(ctx, value)
	case game.PrimitiveReveal:
		value, ok := primitive.(game.Reveal)
		if !ok {
			return "", errors.New("render: internal error: Reveal kind has unexpected concrete type")
		}
		return r.renderRevealPrimitive(ctx, value)
	case game.PrimitiveExileFromHand:
		value, ok := primitive.(game.ExileFromHand)
		if !ok {
			return "", errors.New("render: internal error: ExileFromHand kind has unexpected concrete type")
		}
		return r.renderExileFromHand(ctx, value)
	case game.PrimitivePutFromHand:
		value, ok := primitive.(game.PutFromHand)
		if !ok {
			return "", errors.New("render: internal error: PutFromHand kind has unexpected concrete type")
		}
		return r.renderPutFromHand(ctx, value)
	case game.PrimitivePutHandOnLibraryThenDraw:
		return r.renderPutHandOnLibraryThenDraw(primitive)
	case game.PrimitiveCastForFree:
		value, ok := primitive.(game.CastForFree)
		if !ok {
			return "", errors.New("render: internal error: CastForFree kind has unexpected concrete type")
		}
		return r.renderCastForFree(ctx, value)
	case game.PrimitiveReturnFromGraveyard:
		value, ok := primitive.(game.ReturnFromGraveyard)
		if !ok {
			return "", errors.New("render: internal error: ReturnFromGraveyard kind has unexpected concrete type")
		}
		return r.renderReturnFromGraveyard(ctx, value)
	case game.PrimitiveMassReturnFromGraveyard:
		value, ok := primitive.(game.MassReturnFromGraveyard)
		if !ok {
			return "", errors.New("render: internal error: MassReturnFromGraveyard kind has unexpected concrete type")
		}
		return r.renderMassReturnFromGraveyard(ctx, value)
	case game.PrimitiveMassReanimationExchange:
		value, ok := primitive.(game.MassReanimationExchange)
		if !ok {
			return "", errors.New("render: internal error: MassReanimationExchange kind has unexpected concrete type")
		}
		return r.renderMassReanimationExchange(ctx, value)
	case game.PrimitiveShufflePermanentIntoLibrary:
		value, ok := primitive.(game.ShufflePermanentIntoLibrary)
		if !ok {
			return "", errors.New("render: internal error: ShufflePermanentIntoLibrary kind has unexpected concrete type")
		}
		return r.renderShufflePermanentIntoLibrary(value)
	case game.PrimitiveShuffleSpellIntoLibrary:
		if _, ok := primitive.(game.ShuffleSpellIntoLibrary); !ok {
			return "", errors.New("render: internal error: ShuffleSpellIntoLibrary kind has unexpected concrete type")
		}
		return "game.ShuffleSpellIntoLibrary{}", nil
	case game.PrimitivePutPermanentOnLibrary:
		value, ok := primitive.(game.PutPermanentOnLibrary)
		if !ok {
			return "", errors.New("render: internal error: PutPermanentOnLibrary kind has unexpected concrete type")
		}
		return r.renderPutPermanentOnLibrary(value)
	case game.PrimitiveShuffleLibrary:
		value, ok := primitive.(game.ShuffleLibrary)
		if !ok {
			return "", errors.New("render: internal error: ShuffleLibrary kind has unexpected concrete type")
		}
		return r.renderShuffleLibrary(value)
	case game.PrimitiveLookAtLibraryTop:
		value, ok := primitive.(game.LookAtLibraryTop)
		if !ok {
			return "", errors.New("render: internal error: LookAtLibraryTop kind has unexpected concrete type")
		}
		return r.renderLookAtLibraryTop(value)
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
		value, ok := primitive.(game.AddMana)
		if !ok {
			return "", errors.New("render: internal error: AddMana kind has unexpected concrete type")
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
	case game.PrimitiveMoveCounters:
		value, ok := primitive.(game.MoveCounters)
		if !ok {
			return "", errors.New("render: internal error: MoveCounters kind has unexpected concrete type")
		}
		return r.renderMoveCounters(ctx, &value)
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
	case game.PrimitiveMoveCommander:
		value, ok := primitive.(game.MoveCommander)
		if !ok {
			return "", errors.New("render: internal error: MoveCommander kind has unexpected concrete type")
		}
		return r.renderMoveCommander(ctx, value)
	case game.PrimitiveGrantCastPermission:
		value, ok := primitive.(game.GrantCastPermission)
		if !ok {
			return "", errors.New("render: internal error: GrantCastPermission kind has unexpected concrete type")
		}
		return r.renderGrantCastPermission(ctx, value)
	case game.PrimitiveImpulseExile:
		value, ok := primitive.(game.ImpulseExile)
		if !ok {
			return "", errors.New("render: internal error: ImpulseExile kind has unexpected concrete type")
		}
		return r.renderImpulseExile(ctx, value)
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
	case game.PrimitiveRevealUntil:
		value, ok := primitive.(game.RevealUntil)
		if !ok {
			return "", errors.New("render: internal error: RevealUntil kind has unexpected concrete type")
		}
		return r.renderRevealUntil(ctx, &value)
	case game.PrimitivePunisherEachLoseLife:
		value, ok := primitive.(game.PunisherEachLoseLife)
		if !ok {
			return "", errors.New("render: internal error: PunisherEachLoseLife kind has unexpected concrete type")
		}
		return r.renderPunisherEachLoseLife(ctx, &value)
	case game.PrimitiveRepeatProcess:
		value, ok := primitive.(game.RepeatProcess)
		if !ok {
			return "", errors.New("render: internal error: RepeatProcess kind has unexpected concrete type")
		}
		return r.renderRepeatProcess(ctx, &value)
	case game.PrimitiveCreateToken:
		value, ok := primitive.(game.CreateToken)
		if !ok {
			return "", errors.New("render: internal error: CreateToken kind has unexpected concrete type")
		}
		return r.renderCreateToken(ctx, value)
	case game.PrimitivePreventDamage:
		value, ok := primitive.(game.PreventDamage)
		if !ok {
			return "", errors.New("render: internal error: PreventDamage kind has unexpected concrete type")
		}
		return r.renderPreventDamage(ctx, value)
	default:
		return "", fmt.Errorf("render: unsupported primitive kind %d", primitive.Kind())
	}
}

func (Renderer) renderCardCondition(ctx *renderCtx, condition game.CardCondition) (string, error) {
	card, err := renderCardReference(condition.Card)
	if err != nil {
		return "", err
	}
	fields := []string{fmt.Sprintf("Card: %s,", card)}
	if condition.RequirePermanentCard {
		fields = append(fields, "RequirePermanentCard: true,")
	}
	if len(condition.Types) != 0 {
		typesRendered := make([]string, 0, len(condition.Types))
		for _, cardType := range condition.Types {
			rendered, err := cardTypeLiteral(cardType)
			if err != nil {
				return "", err
			}
			typesRendered = append(typesRendered, rendered)
		}
		ctx.need(importTypes)
		fields = append(fields, fmt.Sprintf("Types: []types.Card{%s},", strings.Join(typesRendered, ", ")))
	}
	if len(condition.Supertypes) != 0 || len(condition.SubtypesAny) != 0 {
		return "", errors.New("render: unsupported CardCondition supertype or subtype filters")
	}
	if condition.ChosenSubtypeFrom != "" {
		switch condition.ChosenSubtypeFrom {
		case game.EntryTypeChoiceKey:
			fields = append(fields, "ChosenSubtypeFrom: game.EntryTypeChoiceKey,")
		default:
			fields = append(fields, fmt.Sprintf("ChosenSubtypeFrom: game.ChoiceKey(%q),", condition.ChosenSubtypeFrom))
		}
	}
	return structLit("game.CardCondition", fields), nil
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

func (r Renderer) renderExileFromHand(ctx *renderCtx, value game.ExileFromHand) (string, error) {
	player, err := r.renderPlayerReference(value.Player)
	if err != nil {
		return "", err
	}
	amount, err := r.renderQuantity(ctx, value.Amount)
	if err != nil {
		return "", err
	}
	selection, err := r.renderSelection(ctx, value.Selection)
	if err != nil {
		return "", err
	}
	fields := []string{
		fmt.Sprintf("Player: %s,", player),
		fmt.Sprintf("Selection: %s,", selection),
		fmt.Sprintf("Amount: %s,", amount),
	}
	if value.PublishLinked != "" {
		fields = append(fields, fmt.Sprintf("PublishLinked: game.LinkedKey(%q),", string(value.PublishLinked)))
	}
	return structLit("game.ExileFromHand", fields), nil
}

func (r Renderer) renderPutFromHand(ctx *renderCtx, value game.PutFromHand) (string, error) {
	player, err := r.renderPlayerReference(value.Player)
	if err != nil {
		return "", err
	}
	amount, err := r.renderQuantity(ctx, value.Amount)
	if err != nil {
		return "", err
	}
	selection, err := r.renderSelection(ctx, value.Selection)
	if err != nil {
		return "", err
	}
	fields := []string{
		fmt.Sprintf("Player: %s,", player),
		fmt.Sprintf("Selection: %s,", selection),
		fmt.Sprintf("Amount: %s,", amount),
	}
	if value.EntersTapped {
		fields = append(fields, "EntersTapped: true,")
	}
	return structLit("game.PutFromHand", fields), nil
}

func (r Renderer) renderPutHandOnLibraryThenDraw(primitive game.Primitive) (string, error) {
	value, ok := primitive.(game.PutHandOnLibraryThenDraw)
	if !ok {
		return "", errors.New("render: internal error: PutHandOnLibraryThenDraw kind has unexpected concrete type")
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

func (r Renderer) renderCastForFree(ctx *renderCtx, value game.CastForFree) (string, error) {
	ctx.need(importZone)
	player, err := r.renderPlayerReference(value.Player)
	if err != nil {
		return "", err
	}
	selection, err := r.renderSelection(ctx, value.Selection)
	if err != nil {
		return "", err
	}
	sourceZone, err := renderZone(value.Zone)
	if err != nil {
		return "", err
	}
	return structLit("game.CastForFree", []string{
		fmt.Sprintf("Player: %s,", player),
		fmt.Sprintf("Selection: %s,", selection),
		fmt.Sprintf("Zone: %s,", sourceZone),
	}), nil
}

func (r Renderer) renderReturnFromGraveyard(ctx *renderCtx, value game.ReturnFromGraveyard) (string, error) {
	player, err := r.renderPlayerReference(value.Player)
	if err != nil {
		return "", err
	}
	amount, err := r.renderQuantity(ctx, value.Amount)
	if err != nil {
		return "", err
	}
	selection, err := r.renderSelection(ctx, value.Selection)
	if err != nil {
		return "", err
	}
	fields := []string{
		fmt.Sprintf("Player: %s,", player),
		fmt.Sprintf("Selection: %s,", selection),
		fmt.Sprintf("Amount: %s,", amount),
	}
	return structLit("game.ReturnFromGraveyard", fields), nil
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
	if value.Spec.CardType.Exists {
		cardType, err := cardTypeLiteral(value.Spec.CardType.Val)
		if err != nil {
			return "", err
		}
		ctx.need(importOpt)
		ctx.need(importTypes)
		specFields = append(specFields, fmt.Sprintf("CardType: opt.Val(%s),", cardType))
	}
	if len(value.Spec.CardTypesAny) > 0 {
		cardTypes := make([]string, 0, len(value.Spec.CardTypesAny))
		for _, value := range value.Spec.CardTypesAny {
			cardType, err := cardTypeLiteral(value)
			if err != nil {
				return "", err
			}
			cardTypes = append(cardTypes, cardType)
		}
		ctx.need(importTypes)
		specFields = append(specFields, fmt.Sprintf("CardTypesAny: []types.Card{%s},", strings.Join(cardTypes, ", ")))
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
	if len(value.Spec.ColorsAny) > 0 {
		colorLits, err := colorValueLiterals(value.Spec.ColorsAny)
		if err != nil {
			return "", err
		}
		ctx.need(importColor)
		specFields = append(specFields, fmt.Sprintf("ColorsAny: []color.Color{%s},", colorLits))
	}
	if value.Spec.MaxManaValue.Exists {
		ctx.need(importOpt)
		specFields = append(specFields, fmt.Sprintf("MaxManaValue: opt.Val(%d),", value.Spec.MaxManaValue.Val))
	}
	if value.Spec.MaxManaValueFromX {
		specFields = append(specFields, "MaxManaValueFromX: true,")
	}
	if value.Spec.MaxPower.Exists {
		ctx.need(importOpt)
		specFields = append(specFields, fmt.Sprintf("MaxPower: opt.Val(%d),", value.Spec.MaxPower.Val))
	}
	if value.Spec.MinPower.Exists {
		ctx.need(importOpt)
		specFields = append(specFields, fmt.Sprintf("MinPower: opt.Val(%d),", value.Spec.MinPower.Val))
	}
	if value.Spec.MaxToughness.Exists {
		ctx.need(importOpt)
		specFields = append(specFields, fmt.Sprintf("MaxToughness: opt.Val(%d),", value.Spec.MaxToughness.Val))
	}
	if value.Spec.MinToughness.Exists {
		ctx.need(importOpt)
		specFields = append(specFields, fmt.Sprintf("MinToughness: opt.Val(%d),", value.Spec.MinToughness.Val))
	}
	if value.Spec.Name != "" {
		specFields = append(specFields, fmt.Sprintf("Name: %q,", value.Spec.Name))
	}
	if value.Spec.Reveal {
		specFields = append(specFields, "Reveal: true,")
	}
	if value.Spec.EntersTapped {
		specFields = append(specFields, "EntersTapped: true,")
	}
	if value.Spec.SharedSubtype {
		specFields = append(specFields, "SharedSubtype: true,")
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
		fmt.Sprintf("Player: %s,", player),
		fmt.Sprintf("Spec: %s,", structLit("game.SearchSpec", specFields)),
		fmt.Sprintf("Amount: %s,", amount),
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
