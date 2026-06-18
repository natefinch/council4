package cardgen

import (
	"errors"
	"fmt"
	"strings"

	"github.com/natefinch/council4/mtg/game"
)

func (r Renderer) renderCreateDelayedTrigger(ctx *renderCtx, value game.CreateDelayedTrigger) (string, error) {
	timing, err := renderDelayedTriggerTiming(value.Trigger.Timing)
	if err != nil {
		return "", err
	}
	content, err := r.renderAbilityContent(ctx, value.Trigger.Content)
	if err != nil {
		return "", err
	}
	triggerFields := []string{
		fmt.Sprintf("Timing: %s,", timing),
		fmt.Sprintf("Content: %s,", content),
	}
	if value.Trigger.Optional {
		triggerFields = append(triggerFields, "Optional: true,")
	}
	return structLit("game.CreateDelayedTrigger", []string{
		fmt.Sprintf("Trigger: %s,", structLit("game.DelayedTriggerDef", triggerFields)),
	}), nil
}

func (r Renderer) renderPutOnBattlefield(ctx *renderCtx, value game.PutOnBattlefield) (string, error) {
	source, err := renderBattlefieldSource(value.Source)
	if err != nil {
		return "", err
	}
	fields := []string{fmt.Sprintf("Source: %s,", source)}
	if value.Recipient.Exists {
		recipient, err := r.renderPlayerReference(value.Recipient.Val)
		if err != nil {
			return "", err
		}
		ctx.need(importOpt)
		fields = append(fields, fmt.Sprintf("Recipient: opt.Val(%s),", recipient))
	}
	if len(value.ContinuousEffects) > 0 {
		return "", errors.New("render: unsupported PutOnBattlefield continuous effects")
	}
	if value.EntryTapped {
		fields = append(fields, "EntryTapped: true,")
	}
	if len(value.EntryCounters) > 0 {
		counters, err := renderCounterPlacements(ctx, value.EntryCounters)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("EntryCounters: []game.CounterPlacement{%s},", strings.Join(counters, ", ")))
	}
	return structLit("game.PutOnBattlefield", fields), nil
}

func renderBattlefieldSource(source game.BattlefieldSource) (string, error) {
	if ref, ok := source.CardRef(); ok {
		rendered, err := renderCardReference(ref)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("game.CardBattlefieldSource(%s)", rendered), nil
	}
	if key, ok := source.LinkedKey(); ok {
		return fmt.Sprintf("game.LinkedBattlefieldSource(game.LinkedKey(%q))", string(key)), nil
	}
	return "", errors.New("render: unsupported battlefield source")
}

func (Renderer) renderMoveCard(ctx *renderCtx, value game.MoveCard) (string, error) {
	card, err := renderCardReference(value.Card)
	if err != nil {
		return "", err
	}
	fromZone, err := renderZone(value.FromZone)
	if err != nil {
		return "", err
	}
	destination, err := renderZone(value.Destination)
	if err != nil {
		return "", err
	}
	ctx.need(importZone)
	fields := []string{
		fmt.Sprintf("Card: %s,", card),
		fmt.Sprintf("FromZone: %s,", fromZone),
		fmt.Sprintf("Destination: %s,", destination),
	}
	if value.DestinationBottom {
		fields = append(fields, "DestinationBottom: true,")
	}
	return structLit("game.MoveCard", fields), nil
}

func (Renderer) renderGrantCastPermission(ctx *renderCtx, value game.GrantCastPermission) (string, error) {
	card, err := renderCardReference(value.Card)
	if err != nil {
		return "", err
	}
	fromZone, err := renderZone(value.FromZone)
	if err != nil {
		return "", err
	}
	duration, err := renderDuration(value.Duration)
	if err != nil {
		return "", err
	}
	if value.Face != game.FaceAlternate {
		return "", fmt.Errorf("render: unsupported cast-permission face %d", value.Face)
	}
	ctx.need(importZone)
	return structLit("game.GrantCastPermission", []string{
		fmt.Sprintf("Card: %s,", card),
		fmt.Sprintf("FromZone: %s,", fromZone),
		"Face: game.FaceAlternate,",
		fmt.Sprintf("Duration: %s,", duration),
	}), nil
}

func renderCardReference(reference game.CardReference) (string, error) {
	switch reference.Kind {
	case game.CardReferenceEvent:
		if reference.LinkID != "" {
			return "", errors.New("render: event card reference has LinkID")
		}
		return "game.CardReference{Kind: game.CardReferenceEvent}", nil
	case game.CardReferenceSource:
		if reference.LinkID != "" {
			return "", errors.New("render: source card reference has LinkID")
		}
		return "game.CardReference{Kind: game.CardReferenceSource}", nil
	case game.CardReferenceTarget:
		if reference.LinkID != "" {
			return "", errors.New("render: target card reference has LinkID")
		}
		if reference.TargetIndex < 0 {
			return "", errors.New("render: target card reference has negative TargetIndex")
		}
		if reference.TargetIndex != 0 {
			return fmt.Sprintf("game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: %d}", reference.TargetIndex), nil
		}
		return "game.CardReference{Kind: game.CardReferenceTarget}", nil
	case game.CardReferenceLinked:
		if reference.LinkID == "" {
			return "", errors.New("render: linked card reference has no LinkID")
		}
		return fmt.Sprintf("game.CardReference{Kind: game.CardReferenceLinked, LinkID: %q}", reference.LinkID), nil
	default:
		return "", fmt.Errorf("render: unsupported card reference kind %d", reference.Kind)
	}
}

func (r Renderer) renderAddCounter(ctx *renderCtx, value *game.AddCounter) (string, error) {
	kind, err := renderCounterKind(value.CounterKind)
	if err != nil {
		return "", err
	}
	ctx.need(importCounter)
	amount, err := r.renderQuantity(ctx, value.Amount)
	if err != nil {
		return "", err
	}
	fields := []string{
		fmt.Sprintf("Amount: %s,", amount),
	}
	if value.Group.Domain() != 0 {
		group, err := r.renderGroupReference(ctx, value.Group)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("Group: %s,", group))
	} else {
		object, err := r.renderObjectReference(value.Object)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("Object: %s,", object))
	}
	fields = append(fields, fmt.Sprintf("CounterKind: %s,", kind))
	return structLit("game.AddCounter", fields), nil
}

func (r Renderer) renderCreateToken(ctx *renderCtx, value game.CreateToken) (string, error) {
	source, err := r.renderTokenSource(ctx, value.Source)
	if err != nil {
		return "", err
	}
	amount, err := r.renderQuantity(ctx, value.Amount)
	if err != nil {
		return "", err
	}
	fields := []string{
		fmt.Sprintf("Amount: %s,", amount),
		fmt.Sprintf("Source: %s,", source),
	}
	if value.Recipient.Exists {
		recipient, err := r.renderPlayerReference(value.Recipient.Val)
		if err != nil {
			return "", err
		}
		ctx.need(importOpt)
		fields = append(fields, fmt.Sprintf("Recipient: opt.Val(%s),", recipient))
	}
	return structLit("game.CreateToken", fields), nil
}

// renderTokenSource renders a CreateToken's TokenSource: either a synthesized
// token CardDef var or a copy of the effect's target object. Richer copy specs
// (Eternalize-style overrides) are built directly in card code, never rendered,
// so they fail closed here.
func (r Renderer) renderTokenSource(ctx *renderCtx, source game.TokenSource) (string, error) {
	if def, ok := source.TokenDefRef(); ok {
		return fmt.Sprintf("game.TokenDef(%s)", ctx.tokenDefVar(def)), nil
	}
	spec, ok := source.TokenCopy()
	if !ok || spec.Source != game.TokenCopySourceObject ||
		spec.SetName != "" || len(spec.SetColors) != 0 || len(spec.SetTypes) != 0 ||
		len(spec.SetSubtypes) != 0 || spec.SetPower.Exists || spec.SetToughness.Exists ||
		spec.NoManaCost || spec.NoPrintedText {
		return "", errors.New("render: unsupported CreateToken token source")
	}
	object, err := r.renderObjectReference(spec.Object)
	if err != nil {
		return "", err
	}
	return structLit("game.TokenCopyOf(game.TokenCopySpec", []string{
		"Source: game.TokenCopySourceObject,",
		fmt.Sprintf("Object: %s,", object),
	}) + ")", nil
}

func (r Renderer) renderAddPlayerCounter(ctx *renderCtx, value *game.AddPlayerCounter) (string, error) {
	player, err := r.renderPlayerReference(value.Player)
	if err != nil {
		return "", err
	}
	kind, err := renderCounterKind(value.CounterKind)
	if err != nil {
		return "", err
	}
	ctx.need(importCounter)
	amount, err := r.renderQuantity(ctx, value.Amount)
	if err != nil {
		return "", err
	}
	return structLit("game.AddPlayerCounter", []string{
		fmt.Sprintf("Amount: %s,", amount),
		fmt.Sprintf("Player: %s,", player),
		fmt.Sprintf("CounterKind: %s,", kind),
	}), nil
}

func (r Renderer) renderDamagePrimitive(ctx *renderCtx, primitive game.Primitive) (string, error) {
	value, ok := primitive.(game.Damage)
	if !ok {
		return "", errors.New("render: internal error: Damage kind has unexpected concrete type")
	}
	recipient, err := r.renderDamageRecipient(ctx, value.Recipient)
	if err != nil {
		return "", err
	}
	amount, err := r.renderQuantity(ctx, value.Amount)
	if err != nil {
		return "", err
	}
	fields := []string{
		fmt.Sprintf("Amount: %s,", amount),
		fmt.Sprintf("Recipient: %s,", recipient),
	}
	if value.Divided {
		fields = append(fields, "Divided: true,")
	}
	if value.DamageSource.Exists {
		source, err := r.renderObjectReference(value.DamageSource.Val)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("DamageSource: opt.Val(%s),", source))
		ctx.need(importOpt)
	}
	return structLit("game.Damage", fields), nil
}

func (r Renderer) renderPlayerAmountPrimitive(ctx *renderCtx, primitive game.Primitive) (string, error) {
	var typeName string
	var amount game.Quantity
	var player game.PlayerReference
	switch primitive.Kind() {
	case game.PrimitiveDraw:
		value, ok := primitive.(game.Draw)
		if !ok {
			return "", errors.New("render: internal error: Draw kind has unexpected concrete type")
		}
		typeName, amount, player = "game.Draw", value.Amount, value.Player
	case game.PrimitiveDiscard:
		value, ok := primitive.(game.Discard)
		if !ok {
			return "", errors.New("render: internal error: Discard kind has unexpected concrete type")
		}
		typeName, amount, player = "game.Discard", value.Amount, value.Player
	case game.PrimitiveMill:
		value, ok := primitive.(game.Mill)
		if !ok {
			return "", errors.New("render: internal error: Mill kind has unexpected concrete type")
		}
		typeName, amount, player = "game.Mill", value.Amount, value.Player
	case game.PrimitiveScry:
		value, ok := primitive.(game.Scry)
		if !ok {
			return "", errors.New("render: internal error: Scry kind has unexpected concrete type")
		}
		typeName, amount, player = "game.Scry", value.Amount, value.Player
	case game.PrimitiveSurveil:
		value, ok := primitive.(game.Surveil)
		if !ok {
			return "", errors.New("render: internal error: Surveil kind has unexpected concrete type")
		}
		typeName, amount, player = "game.Surveil", value.Amount, value.Player
	case game.PrimitiveGainLife:
		value, ok := primitive.(game.GainLife)
		if !ok {
			return "", errors.New("render: internal error: GainLife kind has unexpected concrete type")
		}
		if value.PlayerGroup.Kind != game.PlayerGroupReferenceNone {
			return r.renderAmountPlayerGroup(ctx, "game.GainLife", value.Amount, value.PlayerGroup)
		}
		typeName, amount, player = "game.GainLife", value.Amount, value.Player
	case game.PrimitiveLoseLife:
		value, ok := primitive.(game.LoseLife)
		if !ok {
			return "", errors.New("render: internal error: LoseLife kind has unexpected concrete type")
		}
		if value.PlayerGroup.Kind != game.PlayerGroupReferenceNone {
			return r.renderAmountPlayerGroup(ctx, "game.LoseLife", value.Amount, value.PlayerGroup)
		}
		typeName, amount, player = "game.LoseLife", value.Amount, value.Player
	default:
		return "", fmt.Errorf("render: unsupported player amount primitive kind %d", primitive.Kind())
	}
	rendered, err := r.renderPlayerReference(player)
	if err != nil {
		return "", err
	}
	return r.renderAmountPlayer(ctx, typeName, amount, rendered)
}

func (r Renderer) renderStandalonePrimitive(ctx *renderCtx, primitive game.Primitive) (string, error) {
	switch primitive.Kind() {
	case game.PrimitiveInvestigate:
		value, ok := primitive.(game.Investigate)
		if !ok {
			return "", errors.New("render: internal error: Investigate kind has unexpected concrete type")
		}
		amount, err := r.renderQuantity(ctx, value.Amount)
		if err != nil {
			return "", err
		}
		return structLit("game.Investigate", []string{fmt.Sprintf("Amount: %s,", amount)}), nil
	case game.PrimitiveProliferate:
		value, ok := primitive.(game.Proliferate)
		if !ok {
			return "", errors.New("render: internal error: Proliferate kind has unexpected concrete type")
		}
		amount, err := r.renderQuantity(ctx, value.Amount)
		if err != nil {
			return "", err
		}
		return structLit("game.Proliferate", []string{fmt.Sprintf("Amount: %s,", amount)}), nil
	case game.PrimitiveManifest:
		value, ok := primitive.(game.Manifest)
		if !ok {
			return "", errors.New("render: internal error: Manifest kind has unexpected concrete type")
		}
		var fields []string
		if value.Dread {
			fields = append(fields, "Dread: true,")
		}
		return structLit("game.Manifest", fields), nil
	default:
		return "", fmt.Errorf("render: unsupported standalone primitive kind %d", primitive.Kind())
	}
}

func (r Renderer) renderObjectOrGroupPrimitive(ctx *renderCtx, primitive game.Primitive) (string, error) {
	switch primitive.Kind() {
	case game.PrimitiveDestroy:
		value, ok := primitive.(game.Destroy)
		if !ok {
			return "", errors.New("render: internal error: Destroy kind has unexpected concrete type")
		}
		return r.renderObjectOrGroup(ctx, "game.Destroy", value.Object, value.Group)
	case game.PrimitiveBounce:
		value, ok := primitive.(game.Bounce)
		if !ok {
			return "", errors.New("render: internal error: Bounce kind has unexpected concrete type")
		}
		return r.renderObjectOrGroup(ctx, "game.Bounce", value.Object, value.Group)
	case game.PrimitiveUntap:
		value, ok := primitive.(game.Untap)
		if !ok {
			return "", errors.New("render: internal error: Untap kind has unexpected concrete type")
		}
		return r.renderObjectOrGroup(ctx, "game.Untap", value.Object, value.Group)
	case game.PrimitiveExile:
		value, ok := primitive.(game.Exile)
		if !ok {
			return "", errors.New("render: internal error: Exile kind has unexpected concrete type")
		}
		return r.renderExile(ctx, value)
	default:
		return "", fmt.Errorf("render: unsupported object or group primitive kind %d", primitive.Kind())
	}
}

func (r Renderer) renderExile(ctx *renderCtx, value game.Exile) (string, error) {
	if value.ExileLinkedKey == "" {
		return r.renderObjectOrGroup(ctx, "game.Exile", value.Object, value.Group)
	}
	if value.Group.Domain() != 0 {
		return "", errors.New("render: linked exile requires one object")
	}
	rendered, err := r.renderObjectReference(value.Object)
	if err != nil {
		return "", err
	}
	return structLit("game.Exile", []string{
		fmt.Sprintf("Object: %s,", rendered),
		fmt.Sprintf("ExileLinkedKey: game.LinkedKey(%q),", string(value.ExileLinkedKey)),
	}), nil
}

func (r Renderer) renderObjectPrimitive(primitive game.Primitive) (string, error) {
	var typeName string
	fieldName := "Object"
	var object game.ObjectReference
	switch primitive.Kind() {
	case game.PrimitiveTap:
		value, ok := primitive.(game.Tap)
		if !ok {
			return "", errors.New("render: internal error: Tap kind has unexpected concrete type")
		}
		typeName, object = "game.Tap", value.Object
	case game.PrimitiveRegenerate:
		value, ok := primitive.(game.Regenerate)
		if !ok {
			return "", errors.New("render: internal error: Regenerate kind has unexpected concrete type")
		}
		typeName, object = "game.Regenerate", value.Object
	case game.PrimitiveExplore:
		value, ok := primitive.(game.Explore)
		if !ok {
			return "", errors.New("render: internal error: Explore kind has unexpected concrete type")
		}
		fieldName = "Creature"
		typeName, object = "game.Explore", value.Creature
	case game.PrimitiveCounterObject:
		value, ok := primitive.(game.CounterObject)
		if !ok {
			return "", errors.New("render: internal error: CounterObject kind has unexpected concrete type")
		}
		typeName, object = "game.CounterObject", value.Object
	case game.PrimitiveSacrifice:
		value, ok := primitive.(game.Sacrifice)
		if !ok {
			return "", errors.New("render: internal error: Sacrifice kind has unexpected concrete type")
		}
		typeName, object = "game.Sacrifice", value.Object
	default:
		return "", fmt.Errorf("render: unsupported object primitive kind %d", primitive.Kind())
	}
	rendered, err := r.renderObjectReference(object)
	if err != nil {
		return "", err
	}
	return structLit(typeName, []string{fmt.Sprintf("%s: %s,", fieldName, rendered)}), nil
}

func (r Renderer) renderFightPrimitive(primitive game.Primitive) (string, error) {
	value, ok := primitive.(game.Fight)
	if !ok {
		return "", errors.New("render: internal error: Fight kind has unexpected concrete type")
	}
	object, err := r.renderObjectReference(value.Object)
	if err != nil {
		return "", err
	}
	related, err := r.renderObjectReference(value.RelatedObject)
	if err != nil {
		return "", err
	}
	return structLit("game.Fight", []string{
		fmt.Sprintf("Object: %s,", object),
		fmt.Sprintf("RelatedObject: %s,", related),
	}), nil
}

func (r Renderer) renderAmountPlayer(
	ctx *renderCtx,
	typeName string,
	amount game.Quantity,
	player string,
) (string, error) {
	renderedAmount, err := r.renderQuantity(ctx, amount)
	if err != nil {
		return "", err
	}
	return structLit(typeName, []string{
		fmt.Sprintf("Amount: %s,", renderedAmount),
		fmt.Sprintf("Player: %s,", player),
	}), nil
}

func (r Renderer) renderAmountPlayerGroup(
	ctx *renderCtx,
	typeName string,
	amount game.Quantity,
	group game.PlayerGroupReference,
) (string, error) {
	renderedAmount, err := r.renderQuantity(ctx, amount)
	if err != nil {
		return "", err
	}
	var renderedGroup string
	switch group.Kind {
	case game.PlayerGroupReferenceOpponents:
		renderedGroup = "game.OpponentsReference()"
	case game.PlayerGroupReferenceAllPlayers:
		renderedGroup = "game.AllPlayersReference()"
	default:
		return "", fmt.Errorf("render: unsupported player group reference kind %d", group.Kind)
	}
	return structLit(typeName, []string{
		fmt.Sprintf("Amount: %s,", renderedAmount),
		fmt.Sprintf("PlayerGroup: %s,", renderedGroup),
	}), nil
}

func (r Renderer) renderSacrificePermanents(ctx *renderCtx, value *game.SacrificePermanents) (string, error) {
	renderedAmount, err := r.renderQuantity(ctx, value.Amount)
	if err != nil {
		return "", err
	}
	renderedSelection, err := r.renderSelection(ctx, value.Selection)
	if err != nil {
		return "", err
	}
	fields := []string{fmt.Sprintf("Amount: %s,", renderedAmount)}
	if value.PlayerGroup.Kind != game.PlayerGroupReferenceNone {
		var renderedGroup string
		switch value.PlayerGroup.Kind {
		case game.PlayerGroupReferenceOpponents:
			renderedGroup = "game.OpponentsReference()"
		case game.PlayerGroupReferenceAllPlayers:
			renderedGroup = "game.AllPlayersReference()"
		default:
			return "", fmt.Errorf("render: unsupported player group reference kind %d", value.PlayerGroup.Kind)
		}
		fields = append(fields, fmt.Sprintf("PlayerGroup: %s,", renderedGroup))
	} else {
		player, err := r.renderPlayerReference(value.Player)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("Player: %s,", player))
	}
	if renderedSelection != "game.Selection{}" {
		fields = append(fields, fmt.Sprintf("Selection: %s,", renderedSelection))
	}
	return structLit("game.SacrificePermanents", fields), nil
}

func (r Renderer) renderObjectOrGroup(ctx *renderCtx, typeName string, object game.ObjectReference, group game.GroupReference) (string, error) {
	if group.Domain() != 0 {
		rendered, err := r.renderGroupReference(ctx, group)
		if err != nil {
			return "", err
		}
		return structLit(typeName, []string{fmt.Sprintf("Group: %s,", rendered)}), nil
	}
	rendered, err := r.renderObjectReference(object)
	if err != nil {
		return "", err
	}
	return structLit(typeName, []string{fmt.Sprintf("Object: %s,", rendered)}), nil
}

func (r Renderer) renderAddMana(ctx *renderCtx, value *game.AddMana) (string, error) {
	amount, err := r.renderQuantity(ctx, value.Amount)
	if err != nil {
		return "", err
	}
	fields := []string{fmt.Sprintf("Amount: %s,", amount)}
	if value.ManaColor != "" {
		ctx.need(importMana)
		colorLiteral, err := renderManaColor(value.ManaColor)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("ManaColor: %s,", colorLiteral))
	}
	if value.ChoiceFrom != "" {
		fields = append(fields, fmt.Sprintf("ChoiceFrom: game.ChoiceKey(%q),", string(value.ChoiceFrom)))
	}
	if value.EntryChoiceFrom != "" {
		fields = append(fields, fmt.Sprintf("EntryChoiceFrom: game.ChoiceKey(%q),", string(value.EntryChoiceFrom)))
	}
	return structLit("game.AddMana", fields), nil
}

func (r Renderer) renderModifyPT(ctx *renderCtx, value *game.ModifyPT) (string, error) {
	object, err := r.renderObjectReference(value.Object)
	if err != nil {
		return "", err
	}
	duration, err := renderDuration(value.Duration)
	if err != nil {
		return "", err
	}
	power, err := r.renderQuantity(ctx, value.PowerDelta)
	if err != nil {
		return "", err
	}
	toughness, err := r.renderQuantity(ctx, value.ToughnessDelta)
	if err != nil {
		return "", err
	}
	fields := []string{
		fmt.Sprintf("Object: %s,", object),
		fmt.Sprintf("PowerDelta: %s,", power),
		fmt.Sprintf("ToughnessDelta: %s,", toughness),
		fmt.Sprintf("Duration: %s,", duration),
	}
	if value.PublishLinked != "" {
		fields = append(fields, fmt.Sprintf("PublishLinked: game.LinkedKey(%q),", string(value.PublishLinked)))
	}
	return structLit("game.ModifyPT", fields), nil
}

func (r Renderer) renderChoose(ctx *renderCtx, value game.Choose) (string, error) {
	choice, err := r.renderResolutionChoice(ctx, value.Choice)
	if err != nil {
		return "", err
	}
	fields := []string{fmt.Sprintf("Choice: %s,", choice)}
	if value.PublishChoice != "" {
		fields = append(fields, fmt.Sprintf("PublishChoice: game.ChoiceKey(%q),", string(value.PublishChoice)))
	}
	return structLit("game.Choose", fields), nil
}

func (Renderer) renderResolutionChoice(ctx *renderCtx, choice game.ResolutionChoice) (string, error) {
	kind, err := renderResolutionChoiceKind(choice.Kind)
	if err != nil {
		return "", err
	}
	fields := []string{fmt.Sprintf("Kind: %s,", kind)}
	if len(choice.Colors) > 0 {
		ctx.need(importMana)
		colors, err := renderManaColorSlice(ctx, choice.Colors)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("Colors: %s,", colors))
	}
	return structLit("game.ResolutionChoice", fields), nil
}
