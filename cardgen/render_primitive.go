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
	var fields []string
	if len(value.Sources) > 0 {
		sources := make([]string, len(value.Sources))
		for i, source := range value.Sources {
			rendered, err := renderBattlefieldSource(source)
			if err != nil {
				return "", err
			}
			sources[i] = rendered
		}
		fields = append(fields, fmt.Sprintf("Sources: []game.BattlefieldSource{%s},", strings.Join(sources, ", ")))
	} else {
		source, err := renderBattlefieldSource(value.Source)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("Source: %s,", source))
	}
	if value.Recipient.Exists {
		recipient, err := r.renderPlayerReference(value.Recipient.Val)
		if err != nil {
			return "", err
		}
		ctx.need(importOpt)
		fields = append(fields, fmt.Sprintf("Recipient: opt.Val(%s),", recipient))
	}
	if len(value.ContinuousEffects) > 0 {
		effectLiterals := make([]string, 0, len(value.ContinuousEffects))
		for i := range value.ContinuousEffects {
			eff, err := r.renderContinuousEffect(ctx, &value.ContinuousEffects[i])
			if err != nil {
				return "", err
			}
			effectLiterals = append(effectLiterals, eff+",")
		}
		fields = append(fields, sliceField("ContinuousEffects", "game.ContinuousEffect", effectLiterals))
	}
	if value.EntryTapped {
		fields = append(fields, "EntryTapped: true,")
	}
	if len(value.EntryCounters) > 0 {
		counters, err := r.renderCounterPlacements(ctx, value.EntryCounters)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("EntryCounters: []game.CounterPlacement{%s},", strings.Join(counters, ", ")))
	}
	if value.PublishLinked != "" {
		fields = append(fields, fmt.Sprintf("PublishLinked: game.LinkedKey(%q),", string(value.PublishLinked)))
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

func (r Renderer) renderMoveCard(ctx *renderCtx, value game.MoveCard) (string, error) {
	fromZone, err := renderZone(value.FromZone)
	if err != nil {
		return "", err
	}
	destination, err := renderZone(value.Destination)
	if err != nil {
		return "", err
	}
	ctx.need(importZone)
	var reference string
	switch {
	case value.PlayerGroup.Kind != game.PlayerGroupReferenceNone:
		var group string
		switch value.PlayerGroup.Kind {
		case game.PlayerGroupReferenceOpponents:
			group = "game.OpponentsReference()"
		case game.PlayerGroupReferenceAllPlayers:
			group = "game.AllPlayersReference()"
		default:
			return "", fmt.Errorf("render: unsupported player group reference kind %d", value.PlayerGroup.Kind)
		}
		reference = fmt.Sprintf("PlayerGroup: %s,", group)
	case value.Player.Kind() != game.PlayerReferenceNone:
		player, err := r.renderPlayerReference(value.Player)
		if err != nil {
			return "", err
		}
		reference = fmt.Sprintf("Player: %s,", player)
	default:
		card, err := renderCardReference(value.Card)
		if err != nil {
			return "", err
		}
		reference = fmt.Sprintf("Card: %s,", card)
	}
	fields := []string{
		reference,
		fmt.Sprintf("FromZone: %s,", fromZone),
		fmt.Sprintf("Destination: %s,", destination),
	}
	if value.Amount.IsDynamic() || value.Amount.Value() != 0 {
		amount, err := r.renderQuantity(ctx, value.Amount)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("Amount: %s,", amount))
	}
	if value.DestinationBottom {
		fields = append(fields, "DestinationBottom: true,")
	}
	return structLit("game.MoveCard", fields), nil
}

func (r Renderer) renderMoveCommander(ctx *renderCtx, value game.MoveCommander) (string, error) {
	player, err := r.renderPlayerReference(value.Player)
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
		fmt.Sprintf("Destination: %s,", destination),
	}
	return structLit("game.MoveCommander", fields), nil
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

func (r Renderer) renderImpulseExile(ctx *renderCtx, value game.ImpulseExile) (string, error) {
	player, err := r.renderPlayerReference(value.Player)
	if err != nil {
		return "", err
	}
	amount, err := r.renderQuantity(ctx, value.Amount)
	if err != nil {
		return "", err
	}
	duration, err := renderDuration(value.Duration)
	if err != nil {
		return "", err
	}
	return structLit("game.ImpulseExile", []string{
		fmt.Sprintf("Player: %s,", player),
		fmt.Sprintf("Amount: %s,", amount),
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
	if value.AllKinds {
		object, err := r.renderObjectReference(value.Object)
		if err != nil {
			return "", err
		}
		fields := []string{
			fmt.Sprintf("Object: %s,", object),
			"AllKinds: true,",
		}
		return structLit("game.AddCounter", fields), nil
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
	if value.EntryTapped {
		fields = append(fields, "EntryTapped: true,")
	}
	if value.EntryAttacking {
		fields = append(fields, "EntryAttacking: true,")
	}
	if value.Power.Exists {
		power, err := r.renderQuantity(ctx, value.Power.Val)
		if err != nil {
			return "", err
		}
		ctx.need(importOpt)
		fields = append(fields, fmt.Sprintf("Power: opt.Val(%s),", power))
	}
	if value.Toughness.Exists {
		toughness, err := r.renderQuantity(ctx, value.Toughness.Val)
		if err != nil {
			return "", err
		}
		ctx.need(importOpt)
		fields = append(fields, fmt.Sprintf("Toughness: opt.Val(%s),", toughness))
	}
	if value.PublishLinked != "" {
		fields = append(fields, fmt.Sprintf("PublishLinked: game.LinkedKey(%q),", string(value.PublishLinked)))
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
	if !ok ||
		spec.SetName != "" || len(spec.SetColors) != 0 || len(spec.SetTypes) != 0 ||
		len(spec.SetSubtypes) != 0 || spec.SetPower.Exists || spec.SetToughness.Exists ||
		spec.NoManaCost || spec.NoPrintedText {
		return "", errors.New("render: unsupported CreateToken token source")
	}
	switch spec.Source {
	case game.TokenCopySourceObject:
		return r.renderTokenCopyObjectSource(ctx, spec)
	case game.TokenCopySourceEachInGroup:
		return r.renderTokenCopyForEachSource(ctx, spec)
	default:
		return "", errors.New("render: unsupported CreateToken token source")
	}
}

func (r Renderer) renderTokenCopyObjectSource(ctx *renderCtx, spec game.TokenCopySpec) (string, error) {
	object, err := r.renderObjectReference(spec.Object)
	if err != nil {
		return "", err
	}
	fields := []string{
		"Source: game.TokenCopySourceObject,",
		fmt.Sprintf("Object: %s,", object),
	}
	fields = appendTokenCopyModifierFields(fields, spec)
	rendered, err := renderTokenCopyKeywordField(fields, spec)
	if err != nil {
		return "", err
	}
	return structLit("game.TokenCopyOf(game.TokenCopySpec", rendered) + ")", nil
}

func (r Renderer) renderTokenCopyForEachSource(ctx *renderCtx, spec game.TokenCopySpec) (string, error) {
	group, err := r.renderGroupReference(ctx, *spec.Group)
	if err != nil {
		return "", err
	}
	fields := []string{
		"Source: game.TokenCopySourceEachInGroup,",
		fmt.Sprintf("Group: game.GroupRef(%s),", group),
	}
	fields = appendTokenCopyModifierFields(fields, spec)
	rendered, err := renderTokenCopyKeywordField(fields, spec)
	if err != nil {
		return "", err
	}
	return structLit("game.TokenCopyOf(game.TokenCopySpec", rendered) + ")", nil
}

func appendTokenCopyModifierFields(fields []string, spec game.TokenCopySpec) []string {
	if spec.SetNotLegendary {
		fields = append(fields, "SetNotLegendary: true,")
	}
	return fields
}

func renderTokenCopyKeywordField(fields []string, spec game.TokenCopySpec) ([]string, error) {
	if len(spec.AddKeywords) == 0 {
		return fields, nil
	}
	rendered := make([]string, 0, len(spec.AddKeywords))
	for _, keyword := range spec.AddKeywords {
		literal, err := renderKeyword(keyword)
		if err != nil {
			return nil, err
		}
		rendered = append(rendered, literal)
	}
	return append(fields, fmt.Sprintf("AddKeywords: []game.Keyword{%s},", strings.Join(rendered, ", "))), nil
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

func (r Renderer) renderMoveCounters(ctx *renderCtx, value *game.MoveCounters) (string, error) {
	if value.Distribute {
		group, err := r.renderGroupReference(ctx, *value.Group)
		if err != nil {
			return "", err
		}
		source, err := r.renderCounterSourceSpec(value.Source)
		if err != nil {
			return "", err
		}
		kind, err := renderCounterKind(value.CounterKind)
		if err != nil {
			return "", err
		}
		ctx.need(importCounter)
		return structLit("game.MoveCounters", []string{
			fmt.Sprintf("CounterKind: %s,", kind),
			fmt.Sprintf("Source: %s,", source),
			fmt.Sprintf("Group: game.GroupRef(%s),", group),
			"Distribute: true,",
		}), nil
	}
	object, err := r.renderObjectReference(value.Object)
	if err != nil {
		return "", err
	}
	source, err := r.renderCounterSourceSpec(value.Source)
	if err != nil {
		return "", err
	}
	fields := []string{}
	if value.AllKinds {
		fields = append(fields,
			fmt.Sprintf("Object: %s,", object),
			fmt.Sprintf("Source: %s,", source),
			"AllKinds: true,",
		)
		return structLit("game.MoveCounters", fields), nil
	}
	amount, err := r.renderQuantity(ctx, value.Amount)
	if err != nil {
		return "", err
	}
	if value.ChooseKind {
		fields = append(fields,
			fmt.Sprintf("Amount: %s,", amount),
			fmt.Sprintf("Object: %s,", object),
			fmt.Sprintf("Source: %s,", source),
			"ChooseKind: true,",
		)
		return structLit("game.MoveCounters", fields), nil
	}
	kind, err := renderCounterKind(value.CounterKind)
	if err != nil {
		return "", err
	}
	ctx.need(importCounter)
	fields = append(fields,
		fmt.Sprintf("Amount: %s,", amount),
		fmt.Sprintf("Object: %s,", object),
		fmt.Sprintf("CounterKind: %s,", kind),
		fmt.Sprintf("Source: %s,", source),
	)
	return structLit("game.MoveCounters", fields), nil
}

func (r Renderer) renderCounterSourceSpec(source game.CounterSourceSpec) (string, error) {
	switch source.Kind {
	case game.CounterSourceSelf:
		return "game.CounterSourceSpec{Kind: game.CounterSourceSelf}", nil
	case game.CounterSourceEventPermanent:
		return "game.CounterSourceSpec{Kind: game.CounterSourceEventPermanent}", nil
	case game.CounterSourceTarget:
		object, err := r.renderObjectReference(source.Object)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("game.CounterSourceSpec{Kind: game.CounterSourceTarget, Object: %s}", object), nil
	default:
		return "", fmt.Errorf("render: unsupported counter source kind %d", source.Kind)
	}
}

func (r Renderer) renderGroupSourceDamage(ctx *renderCtx, primitive game.Primitive) (string, error) {
	value, ok := primitive.(game.GroupSourceDamage)
	if !ok {
		return "", errors.New("render: internal error: GroupSourceDamage kind has unexpected concrete type")
	}
	group, err := r.renderGroupReference(ctx, value.Group)
	if err != nil {
		return "", err
	}
	amount, err := r.renderQuantity(ctx, value.Amount)
	if err != nil {
		return "", err
	}
	fields := []string{
		fmt.Sprintf("Group: %s,", group),
		fmt.Sprintf("Amount: %s,", amount),
	}
	if value.ToOwner {
		fields = append(fields, "ToOwner: true,")
	}
	return structLit("game.GroupSourceDamage", fields), nil
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
		if value.PlayerGroup.Kind != game.PlayerGroupReferenceNone {
			return r.renderAmountPlayerGroup(ctx, "game.Draw", value.Amount, value.PlayerGroup)
		}
		typeName, amount, player = "game.Draw", value.Amount, value.Player
	case game.PrimitiveDiscard:
		value, ok := primitive.(game.Discard)
		if !ok {
			return "", errors.New("render: internal error: Discard kind has unexpected concrete type")
		}
		if value.EntireHand {
			return r.renderDiscardEntireHand(value)
		}
		if value.PlayerGroup.Kind != game.PlayerGroupReferenceNone {
			return r.renderAmountPlayerGroup(ctx, "game.Discard", value.Amount, value.PlayerGroup)
		}
		typeName, amount, player = "game.Discard", value.Amount, value.Player
	case game.PrimitiveMill:
		value, ok := primitive.(game.Mill)
		if !ok {
			return "", errors.New("render: internal error: Mill kind has unexpected concrete type")
		}
		if value.PlayerGroup.Kind != game.PlayerGroupReferenceNone {
			return r.renderAmountPlayerGroup(ctx, "game.Mill", value.Amount, value.PlayerGroup)
		}
		typeName, amount, player = "game.Mill", value.Amount, value.Player
	case game.PrimitiveExileTopOfLibrary:
		value, ok := primitive.(game.ExileTopOfLibrary)
		if !ok {
			return "", errors.New("render: internal error: ExileTopOfLibrary kind has unexpected concrete type")
		}
		if value.PlayerGroup.Kind != game.PlayerGroupReferenceNone {
			return r.renderAmountPlayerGroup(ctx, "game.ExileTopOfLibrary", value.Amount, value.PlayerGroup)
		}
		typeName, amount, player = "game.ExileTopOfLibrary", value.Amount, value.Player
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
	case game.PrimitiveReorderLibraryTop:
		value, ok := primitive.(game.ReorderLibraryTop)
		if !ok {
			return "", errors.New("render: internal error: ReorderLibraryTop kind has unexpected concrete type")
		}
		typeName, amount, player = "game.ReorderLibraryTop", value.Amount, value.Player
	default:
		return "", fmt.Errorf("render: unsupported player amount primitive kind %d", primitive.Kind())
	}
	rendered, err := r.renderPlayerReference(player)
	if err != nil {
		return "", err
	}
	return r.renderAmountPlayer(ctx, typeName, amount, rendered)
}

func (r Renderer) renderPlayerLosesGame(value game.PlayerLosesGame) (string, error) {
	rendered, err := r.renderPlayerReference(value.Player)
	if err != nil {
		return "", err
	}
	return structLit("game.PlayerLosesGame", []string{fmt.Sprintf("Player: %s,", rendered)}), nil
}

func (r Renderer) renderPlayerWinsGame(value game.PlayerWinsGame) (string, error) {
	rendered, err := r.renderPlayerReference(value.Player)
	if err != nil {
		return "", err
	}
	return structLit("game.PlayerWinsGame", []string{fmt.Sprintf("Player: %s,", rendered)}), nil
}

func (r Renderer) renderShuffleLibrary(value game.ShuffleLibrary) (string, error) {
	player, err := r.renderPlayerReference(value.Player)
	if err != nil {
		return "", err
	}
	return structLit("game.ShuffleLibrary", []string{
		fmt.Sprintf("Player: %s,", player),
	}), nil
}

func (r Renderer) renderLookAtHand(value game.LookAtHand) (string, error) {
	player, err := r.renderPlayerReference(value.Player)
	if err != nil {
		return "", err
	}
	return structLit("game.LookAtHand", []string{
		fmt.Sprintf("Player: %s,", player),
	}), nil
}

func (r Renderer) renderLookAtLibraryTop(value game.LookAtLibraryTop) (string, error) {
	if value.PublishLinked == "" {
		return "", errors.New("render: LookAtLibraryTop has no PublishLinked")
	}
	player, err := r.renderPlayerReference(value.Player)
	if err != nil {
		return "", err
	}
	return structLit("game.LookAtLibraryTop", []string{
		fmt.Sprintf("Player: %s,", player),
		fmt.Sprintf("PublishLinked: game.LinkedKey(%q),", string(value.PublishLinked)),
	}), nil
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
		if value.Player.Kind() != game.PlayerReferenceNone {
			player, err := r.renderPlayerReference(value.Player)
			if err != nil {
				return "", err
			}
			fields = append(fields, fmt.Sprintf("Player: %s,", player))
		}
		return structLit("game.Manifest", fields), nil
	default:
		return "", fmt.Errorf("render: unsupported standalone primitive kind %d", primitive.Kind())
	}
}

// renderAmass renders an Amass primitive, emitting its fixed count and the
// named Army creature subtype.
func (r Renderer) renderAmass(ctx *renderCtx, value game.Amass) (string, error) {
	amount, err := r.renderQuantity(ctx, value.Amount)
	if err != nil {
		return "", err
	}
	fields := []string{fmt.Sprintf("Amount: %s,", amount)}
	if value.Subtype != "" {
		ctx.need(importTypes)
		lit := SubtypeToLiteral(string(value.Subtype), []string{"Creature"})
		if strings.HasPrefix(lit, "/*") {
			return "", fmt.Errorf("render: unsupported amass subtype %q", string(value.Subtype))
		}
		fields = append(fields, fmt.Sprintf("Subtype: %s,", lit))
	}
	return structLit("game.Amass", fields), nil
}

// renderRenown renders a Renown primitive, emitting the renowned-permanent
// reference and the fixed +1/+1 counter count.
func (r Renderer) renderRenown(ctx *renderCtx, value game.Renown) (string, error) {
	object, err := r.renderObjectReference(value.Object)
	if err != nil {
		return "", err
	}
	amount, err := r.renderQuantity(ctx, value.Amount)
	if err != nil {
		return "", err
	}
	return structLit("game.Renown", []string{
		fmt.Sprintf("Object: %s,", object),
		fmt.Sprintf("Amount: %s,", amount),
	}), nil
}

// renderBecomeSaddled renders a BecomeSaddled primitive, emitting the saddled
// Mount reference.
func (r Renderer) renderBecomeSaddled(_ *renderCtx, value game.BecomeSaddled) (string, error) {
	object, err := r.renderObjectReference(value.Object)
	if err != nil {
		return "", err
	}
	return structLit("game.BecomeSaddled", []string{
		fmt.Sprintf("Object: %s,", object),
	}), nil
}

func (r Renderer) renderDigPrimitive(ctx *renderCtx, value game.Dig) (string, error) {
	player, err := r.renderPlayerReference(value.Player)
	if err != nil {
		return "", err
	}
	look, err := r.renderQuantity(ctx, value.Look)
	if err != nil {
		return "", err
	}
	take, err := r.renderQuantity(ctx, value.Take)
	if err != nil {
		return "", err
	}
	fields := []string{
		fmt.Sprintf("Player: %s,", player),
		fmt.Sprintf("Look: %s,", look),
		fmt.Sprintf("Take: %s,", take),
	}
	if value.Remainder == game.DigRemainderLibraryBottom {
		fields = append(fields, "Remainder: game.DigRemainderLibraryBottom,")
	}
	return structLit("game.Dig", fields), nil
}

func (r Renderer) renderObjectOrGroupPrimitive(ctx *renderCtx, primitive game.Primitive) (string, error) {
	switch primitive.Kind() {
	case game.PrimitiveDestroy:
		value, ok := primitive.(game.Destroy)
		if !ok {
			return "", errors.New("render: internal error: Destroy kind has unexpected concrete type")
		}
		return r.renderDestroy(ctx, value)
	case game.PrimitiveBounce:
		value, ok := primitive.(game.Bounce)
		if !ok {
			return "", errors.New("render: internal error: Bounce kind has unexpected concrete type")
		}
		return r.renderBounce(ctx, value)
	case game.PrimitiveUntap:
		value, ok := primitive.(game.Untap)
		if !ok {
			return "", errors.New("render: internal error: Untap kind has unexpected concrete type")
		}
		return r.renderUntap(ctx, value)
	case game.PrimitiveTap:
		value, ok := primitive.(game.Tap)
		if !ok {
			return "", errors.New("render: internal error: Tap kind has unexpected concrete type")
		}
		return r.renderObjectOrGroup(ctx, "game.Tap", value.Object, value.Group)
	case game.PrimitiveExile:
		value, ok := primitive.(game.Exile)
		if !ok {
			return "", errors.New("render: internal error: Exile kind has unexpected concrete type")
		}
		return r.renderExile(ctx, value)
	case game.PrimitivePhaseOut:
		value, ok := primitive.(game.PhaseOut)
		if !ok {
			return "", errors.New("render: internal error: PhaseOut kind has unexpected concrete type")
		}
		return r.renderObjectOrGroup(ctx, "game.PhaseOut", value.Object, value.Group)
	default:
		return "", fmt.Errorf("render: unsupported object or group primitive kind %d", primitive.Kind())
	}
}

func (r Renderer) renderDestroy(ctx *renderCtx, value game.Destroy) (string, error) {
	if !value.PreventRegeneration {
		return r.renderObjectOrGroup(ctx, "game.Destroy", value.Object, value.Group)
	}
	var reference string
	if value.Group.Domain() != 0 {
		rendered, err := r.renderGroupReference(ctx, value.Group)
		if err != nil {
			return "", err
		}
		reference = fmt.Sprintf("Group: %s,", rendered)
	} else {
		rendered, err := r.renderObjectReference(value.Object)
		if err != nil {
			return "", err
		}
		reference = fmt.Sprintf("Object: %s,", rendered)
	}
	return structLit("game.Destroy", []string{
		reference,
		"PreventRegeneration: true,",
	}), nil
}

func (r Renderer) renderExile(ctx *renderCtx, value game.Exile) (string, error) {
	if value.SourceSpell {
		if value.Object.Kind() != game.ObjectReferenceNone || value.Group.Domain() != 0 || value.ExileLinkedKey != "" {
			return "", errors.New("render: source-spell exile cannot set an object, group, or linked key")
		}
		return structLit("game.Exile", []string{"SourceSpell: true,"}), nil
	}
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

// renderBecomeCopy renders a BecomeCopy primitive, including its optional
// until-end-of-turn duration, retain-this-ability flag, and copiable keyword
// riders.
func (r Renderer) renderBecomeCopy(value game.BecomeCopy) (string, error) {
	var fields []string
	if value.Card.Kind != game.CardReferenceNone {
		card, err := renderCardReference(value.Card)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("Card: %s,", card))
	} else {
		object, err := r.renderObjectReference(value.Object)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("Object: %s,", object))
	}
	if value.UntilEndOfTurn {
		fields = append(fields, "UntilEndOfTurn: true,")
	}
	if value.RetainsThisAbility {
		fields = append(fields, "RetainsThisAbility: true,")
	}
	if len(value.AddKeywords) > 0 {
		rendered := make([]string, 0, len(value.AddKeywords))
		for _, keyword := range value.AddKeywords {
			literal, err := renderKeyword(keyword)
			if err != nil {
				return "", err
			}
			rendered = append(rendered, literal)
		}
		fields = append(fields, fmt.Sprintf("AddKeywords: []game.Keyword{%s},", strings.Join(rendered, ", ")))
	}
	return structLit("game.BecomeCopy", fields), nil
}

func (r Renderer) renderObjectPrimitive(primitive game.Primitive) (string, error) {
	var typeName string
	fieldName := "Object"
	var object game.ObjectReference
	switch primitive.Kind() {
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
		if value.ExileInstead {
			rendered, err := r.renderObjectReference(value.Object)
			if err != nil {
				return "", err
			}
			return structLit("game.CounterObject", []string{
				fmt.Sprintf("Object: %s,", rendered),
				"ExileInstead: true,",
			}), nil
		}
		typeName, object = "game.CounterObject", value.Object
	case game.PrimitiveChooseNewTargets:
		value, ok := primitive.(game.ChooseNewTargets)
		if !ok {
			return "", errors.New("render: internal error: ChooseNewTargets kind has unexpected concrete type")
		}
		typeName, object = "game.ChooseNewTargets", value.Object
	case game.PrimitiveSacrifice:
		value, ok := primitive.(game.Sacrifice)
		if !ok {
			return "", errors.New("render: internal error: Sacrifice kind has unexpected concrete type")
		}
		typeName, object = "game.Sacrifice", value.Object
	case game.PrimitiveSkipNextUntap:
		value, ok := primitive.(game.SkipNextUntap)
		if !ok {
			return "", errors.New("render: internal error: SkipNextUntap kind has unexpected concrete type")
		}
		typeName, object = "game.SkipNextUntap", value.Object
	default:
		return "", fmt.Errorf("render: unsupported object primitive kind %d", primitive.Kind())
	}
	rendered, err := r.renderObjectReference(object)
	if err != nil {
		return "", err
	}
	return structLit(typeName, []string{fmt.Sprintf("%s: %s,", fieldName, rendered)}), nil
}

func (r Renderer) renderCopyStackObjectPrimitive(value game.CopyStackObject) (string, error) {
	rendered, err := r.renderObjectReference(value.Object)
	if err != nil {
		return "", err
	}
	fields := []string{fmt.Sprintf("Object: %s,", rendered)}
	if value.MayChooseNewTargets {
		fields = append(fields, "MayChooseNewTargets: true,")
	}
	return structLit("game.CopyStackObject", fields), nil
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
func (r Renderer) renderAttachPrimitive(primitive game.Primitive) (string, error) {
	value, ok := primitive.(game.Attach)
	if !ok {
		return "", errors.New("render: internal error: Attach kind has unexpected concrete type")
	}
	attachment, err := r.renderObjectReference(value.Attachment)
	if err != nil {
		return "", err
	}
	target, err := r.renderObjectReference(value.Target)
	if err != nil {
		return "", err
	}
	return structLit("game.Attach", []string{
		fmt.Sprintf("Attachment: %s,", attachment),
		fmt.Sprintf("Target: %s,", target),
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

// renderDiscardEntireHand renders a "discard their hand" primitive, which sets
// EntireHand and leaves Amount unset.
func (r Renderer) renderDiscardEntireHand(value game.Discard) (string, error) {
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
		return structLit("game.Discard", []string{
			"EntireHand: true,",
			fmt.Sprintf("PlayerGroup: %s,", renderedGroup),
		}), nil
	}
	player, err := r.renderPlayerReference(value.Player)
	if err != nil {
		return "", err
	}
	return structLit("game.Discard", []string{
		"EntireHand: true,",
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
	if value.Fallback.Kind != game.SacrificeFallbackNone {
		renderedFallback, err := r.renderSacrificeFallback(ctx, value.Fallback)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("Fallback: %s,", renderedFallback))
	}
	return structLit("game.SacrificePermanents", fields), nil
}

func (r Renderer) renderRevealUntil(ctx *renderCtx, value *game.RevealUntil) (string, error) {
	destination, err := renderZone(value.Destination)
	if err != nil {
		return "", err
	}
	ctx.need(importZone)
	renderedSelection, err := r.renderSelection(ctx, value.Until)
	if err != nil {
		return "", err
	}
	var fields []string
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
		fields = append(fields, fmt.Sprintf("Until: %s,", renderedSelection))
	}
	fields = append(fields, fmt.Sprintf("Destination: %s,", destination))
	return structLit("game.RevealUntil", fields), nil
}

func (r Renderer) renderPunisherEachLoseLife(ctx *renderCtx, value *game.PunisherEachLoseLife) (string, error) {
	renderedAmount, err := r.renderQuantity(ctx, value.Amount)
	if err != nil {
		return "", err
	}
	var renderedGroup string
	switch value.PlayerGroup.Kind {
	case game.PlayerGroupReferenceOpponents:
		renderedGroup = "game.OpponentsReference()"
	case game.PlayerGroupReferenceAllPlayers:
		renderedGroup = "game.AllPlayersReference()"
	default:
		return "", fmt.Errorf("render: unsupported player group reference kind %d", value.PlayerGroup.Kind)
	}
	fields := []string{
		fmt.Sprintf("PlayerGroup: %s,", renderedGroup),
		fmt.Sprintf("Amount: %s,", renderedAmount),
	}
	if value.AllowSacrifice {
		fields = append(fields, "AllowSacrifice: true,")
		renderedSelection, err := r.renderSelection(ctx, value.SacrificeSelection)
		if err != nil {
			return "", err
		}
		if renderedSelection != "game.Selection{}" {
			fields = append(fields, fmt.Sprintf("SacrificeSelection: %s,", renderedSelection))
		}
	}
	if value.AllowDiscard {
		fields = append(fields, "AllowDiscard: true,")
	}
	return structLit("game.PunisherEachLoseLife", fields), nil
}

func (r Renderer) renderRepeatProcess(ctx *renderCtx, value *game.RepeatProcess) (string, error) {
	renderedTimes, err := r.renderQuantity(ctx, value.Times)
	if err != nil {
		return "", err
	}
	renderedBody, err := r.renderAbilityContent(ctx, value.Body)
	if err != nil {
		return "", err
	}
	return structLit("game.RepeatProcess", []string{
		fmt.Sprintf("Times: %s,", renderedTimes),
		fmt.Sprintf("Body: %s,", renderedBody),
	}), nil
}

func (r Renderer) renderSacrificeFallback(ctx *renderCtx, value game.SacrificeFallback) (string, error) {
	renderedAmount, err := r.renderQuantity(ctx, value.Amount)
	if err != nil {
		return "", err
	}
	var kind string
	switch value.Kind {
	case game.SacrificeFallbackDiscard:
		kind = "game.SacrificeFallbackDiscard"
	case game.SacrificeFallbackLoseLife:
		kind = "game.SacrificeFallbackLoseLife"
	default:
		return "", fmt.Errorf("render: unsupported sacrifice fallback kind %d", value.Kind)
	}
	fields := []string{
		fmt.Sprintf("Kind: %s,", kind),
		fmt.Sprintf("Amount: %s,", renderedAmount),
	}
	return structLit("game.SacrificeFallback", fields), nil
}

func (r Renderer) renderBounce(ctx *renderCtx, value game.Bounce) (string, error) {
	if !value.ControlledChoice {
		return r.renderObjectOrGroup(ctx, "game.Bounce", value.Object, value.Group)
	}
	renderedAmount, err := r.renderQuantity(ctx, value.Amount)
	if err != nil {
		return "", err
	}
	renderedGroup, err := r.renderGroupReference(ctx, value.Group)
	if err != nil {
		return "", err
	}
	fields := []string{
		"ControlledChoice: true,",
		fmt.Sprintf("Amount: %s,", renderedAmount),
		fmt.Sprintf("Group: %s,", renderedGroup),
	}
	return structLit("game.Bounce", fields), nil
}

func (r Renderer) renderUntap(ctx *renderCtx, value game.Untap) (string, error) {
	if !value.ChooseUpTo {
		return r.renderObjectOrGroup(ctx, "game.Untap", value.Object, value.Group)
	}
	renderedAmount, err := r.renderQuantity(ctx, value.Amount)
	if err != nil {
		return "", err
	}
	renderedGroup, err := r.renderGroupReference(ctx, value.Group)
	if err != nil {
		return "", err
	}
	return structLit("game.Untap", []string{
		"ChooseUpTo: true,",
		fmt.Sprintf("Amount: %s,", renderedAmount),
		fmt.Sprintf("Group: %s,", renderedGroup),
	}), nil
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
	if value.SpendRider.Exists {
		rider, err := r.renderManaSpendRider(ctx, value.SpendRider.Val)
		if err != nil {
			return "", err
		}
		ctx.need(importOpt)
		fields = append(fields, fmt.Sprintf("SpendRider: opt.Val(%s),", rider))
	}
	if value.Player.Exists {
		playerRef, err := r.renderPlayerReference(value.Player.Val)
		if err != nil {
			return "", err
		}
		ctx.need(importOpt)
		fields = append(fields, fmt.Sprintf("Player: opt.Val(%s),", playerRef))
	}
	return structLit("game.AddMana", fields), nil
}

// renderManaSpendRider renders spend-linked semantics tagged onto produced mana.
func (r Renderer) renderManaSpendRider(ctx *renderCtx, rider game.ManaSpendRider) (string, error) {
	condition, err := renderManaSpendConditionKind(rider.Condition)
	if err != nil {
		return "", err
	}
	fields := []string{
		fmt.Sprintf("Condition: %s,", condition),
	}
	if rider.Restriction != game.ManaSpendUnrestricted {
		restriction, err := renderManaSpendRestrictionKind(rider.Restriction)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("Restriction: %s,", restriction))
	}
	if len(rider.Effect.Sequence) > 0 {
		effect, err := r.renderMode(ctx, rider.Effect)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("Effect: %s,", effect))
	}
	if rider.SpellRuleEffect != game.RuleEffectNone {
		ruleEffect, err := renderRuleEffectKind(rider.SpellRuleEffect)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("SpellRuleEffect: %s,", ruleEffect))
	}
	if rider.ChosenSubtypeFrom != "" {
		switch rider.ChosenSubtypeFrom {
		case game.EntryTypeChoiceKey:
			fields = append(fields, "ChosenSubtypeFrom: game.EntryTypeChoiceKey,")
		default:
			fields = append(fields, fmt.Sprintf("ChosenSubtypeFrom: game.ChoiceKey(%q),", rider.ChosenSubtypeFrom))
		}
	}
	if len(rider.SpellGainsKeywords) > 0 {
		elements := make([]string, 0, len(rider.SpellGainsKeywords))
		for _, keyword := range rider.SpellGainsKeywords {
			literal, err := renderKeyword(keyword)
			if err != nil {
				return "", err
			}
			elements = append(elements, literal+",")
		}
		fields = append(fields, sliceField("SpellGainsKeywords", "game.Keyword", elements))
	}
	return structLit("game.ManaSpendRider", fields), nil
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

func (r Renderer) renderPreventDamage(ctx *renderCtx, value game.PreventDamage) (string, error) {
	var fields []string
	if value.Object.Kind() != game.ObjectReferenceNone {
		object, err := r.renderObjectReference(value.Object)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("Object: %s,", object))
	}
	if value.Player.Kind() != game.PlayerReferenceNone {
		player, err := r.renderPlayerReference(value.Player)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("Player: %s,", player))
	}
	if !value.All {
		amount, err := r.renderQuantity(ctx, value.Amount)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("Amount: %s,", amount))
	}
	if value.All {
		fields = append(fields, "All: true,")
	}
	if value.CombatOnly {
		fields = append(fields, "CombatOnly: true,")
	}
	if value.BySource {
		fields = append(fields, "BySource: true,")
	}
	if value.Global {
		fields = append(fields, "Global: true,")
	}
	return structLit("game.PreventDamage", fields), nil
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

func (r Renderer) renderResolutionChoice(ctx *renderCtx, choice game.ResolutionChoice) (string, error) {
	kind, err := renderResolutionChoiceKind(choice.Kind)
	if err != nil {
		return "", err
	}
	fields := []string{fmt.Sprintf("Kind: %s,", kind)}
	if choice.Prompt != "" {
		fields = append(fields, fmt.Sprintf("Prompt: %q,", choice.Prompt))
	}
	if choice.PlayerReference != nil {
		player, err := r.renderPlayerReference(*choice.PlayerReference)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("PlayerReference: func() *game.PlayerReference { ref := %s; return &ref }(),", player))
	}
	if choice.ColorSource != game.ResolutionChoiceColorSourceStatic {
		source, err := renderResolutionChoiceColorSource(choice.ColorSource)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("ColorSource: %s,", source))
	}
	if choice.PlayerRelation != game.PlayerAny {
		relation, err := renderPlayerRelation(choice.PlayerRelation)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("PlayerRelation: %s,", relation))
	}
	if choice.IncludeColorless {
		fields = append(fields, "IncludeColorless: true,")
	}
	if choice.Selection != nil && !choice.Selection.Empty() {
		selection, err := r.renderSelection(ctx, *choice.Selection)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("Selection: &%s,", selection))
	}
	if choice.Kind == game.ResolutionChoiceNumber {
		fields = append(fields,
			fmt.Sprintf("MinNumber: %d,", choice.MinNumber),
			fmt.Sprintf("MaxNumber: %d,", choice.MaxNumber),
		)
	}
	if len(choice.Colors) > 0 {
		ctx.need(importMana)
		colors, err := renderManaColorSlice(ctx, choice.Colors)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("Colors: %s,", colors))
	}
	if choice.Kind == game.ResolutionChoiceSubtype {
		ctx.need(importTypes)
		lit, err := cardTypeLiteral(choice.SubtypeOfType)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("SubtypeOfType: %s,", lit))
	}
	return structLit("game.ResolutionChoice", fields), nil
}
