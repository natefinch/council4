package cardgen

import (
	"errors"
	"fmt"
	"strings"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

func (r Renderer) renderCreateReplacement(ctx *renderCtx, value game.CreateReplacement) (string, error) {
	if value.Replacement == nil {
		return "", errors.New("render: CreateReplacement has no replacement effect")
	}
	replacement, err := r.renderCreateReplacementEffect(ctx, *value.Replacement)
	if err != nil {
		return "", err
	}
	fields := []string{
		fmt.Sprintf("Replacement: %s,", replacement),
	}
	if value.Object.Kind() != game.ObjectReferenceNone {
		object, objErr := r.renderObjectReference(value.Object)
		if objErr != nil {
			return "", objErr
		}
		fields = append(fields, fmt.Sprintf("Object: %s,", object))
	}
	if value.Duration != game.DurationPermanent {
		duration, durErr := renderDuration(value.Duration)
		if durErr != nil {
			return "", durErr
		}
		fields = append(fields, fmt.Sprintf("Duration: %s,", duration))
	}
	return structLit("game.CreateReplacement", fields), nil
}

// renderCreateReplacementEffect renders the dynamically created
// ReplacementEffect as a pointer to a struct literal. It supports the
// zone-change redirect shape produced for the leaves-the-battlefield exile
// replacement (MatchEvent EventZoneChanged, a from-zone match, and a
// ReplaceToZone redirect) and the future-cast enters-with-counters shape
// (MatchEvent EventPermanentEnteredBattlefield with EntersWithCounters).
func (r Renderer) renderCreateReplacementEffect(ctx *renderCtx, replacement game.ReplacementEffect) (string, error) {
	event, err := renderEventKind(replacement.MatchEvent)
	if err != nil {
		return "", err
	}
	fields := []string{
		fmt.Sprintf("MatchEvent: %s,", event),
	}
	if replacement.Description != "" {
		fields = append(fields, fmt.Sprintf("Description: %q,", replacement.Description))
	}
	if replacement.MatchFromZone {
		fromZone, zoneErr := renderZone(replacement.FromZone)
		if zoneErr != nil {
			return "", zoneErr
		}
		ctx.need(importZone)
		fields = append(fields,
			"MatchFromZone: true,",
			fmt.Sprintf("FromZone: %s,", fromZone),
		)
	}
	if replacement.MatchToZone {
		toZone, zoneErr := renderZone(replacement.ToZone)
		if zoneErr != nil {
			return "", zoneErr
		}
		ctx.need(importZone)
		fields = append(fields,
			"MatchToZone: true,",
			fmt.Sprintf("ToZone: %s,", toZone),
		)
	}
	if replacement.ReplaceToZone != zone.None {
		replaceZone, zoneErr := renderZone(replacement.ReplaceToZone)
		if zoneErr != nil {
			return "", zoneErr
		}
		ctx.need(importZone)
		fields = append(fields, fmt.Sprintf("ReplaceToZone: %s,", replaceZone))
	}
	if replacement.AffectedObjectMustBeCreature {
		fields = append(fields, "AffectedObjectMustBeCreature: true,")
	}
	if len(replacement.EntersWithCounters) > 0 {
		placements, placeErr := r.renderCounterPlacements(ctx, replacement.EntersWithCounters)
		if placeErr != nil {
			return "", placeErr
		}
		fields = append(fields, fmt.Sprintf("EntersWithCounters: []game.CounterPlacement{%s},", strings.Join(placements, ", ")))
	}
	return "&" + structLit("game.ReplacementEffect", fields), nil
}

func (r Renderer) renderCreateDelayedTrigger(ctx *renderCtx, value game.CreateDelayedTrigger) (string, error) {
	content, err := r.renderAbilityContent(ctx, value.Trigger.Content)
	if err != nil {
		return "", err
	}
	var triggerFields []string
	if value.Trigger.EventPattern.Exists {
		pattern, err := r.renderTriggerPattern(ctx, &value.Trigger.EventPattern.Val)
		if err != nil {
			return "", err
		}
		window, err := renderDelayedTriggerWindow(value.Trigger.Window)
		if err != nil {
			return "", err
		}
		ctx.need(importOpt)
		triggerFields = append(triggerFields, fmt.Sprintf("EventPattern: opt.Val(%s),", pattern))
		if value.Trigger.OneShot {
			triggerFields = append(triggerFields, "OneShot: true,")
		}
		triggerFields = append(triggerFields, fmt.Sprintf("Window: %s,", window))
		if value.Trigger.DamageSourceObject.Exists {
			object, err := r.renderObjectReference(value.Trigger.DamageSourceObject.Val)
			if err != nil {
				return "", err
			}
			triggerFields = append(triggerFields, fmt.Sprintf("DamageSourceObject: opt.Val(%s),", object))
		}
		if value.Trigger.CapturedAttackerObject.Exists {
			object, err := r.renderObjectReference(value.Trigger.CapturedAttackerObject.Val)
			if err != nil {
				return "", err
			}
			triggerFields = append(triggerFields, fmt.Sprintf("CapturedAttackerObject: opt.Val(%s),", object))
		}
		if value.Trigger.CapturedDyingObject.Exists {
			object, err := r.renderObjectReference(value.Trigger.CapturedDyingObject.Val)
			if err != nil {
				return "", err
			}
			triggerFields = append(triggerFields, fmt.Sprintf("CapturedDyingObject: opt.Val(%s),", object))
		}
		if value.Trigger.InterveningCondition.Exists {
			condition, err := r.renderControllerControlsCondition(ctx, &value.Trigger.InterveningCondition.Val, "delayed trigger intervening")
			if err != nil {
				return "", err
			}
			triggerFields = append(triggerFields, fmt.Sprintf("InterveningCondition: opt.Val(%s),", condition))
		}
	} else {
		timing, err := renderDelayedTriggerTiming(value.Trigger.Timing)
		if err != nil {
			return "", err
		}
		triggerFields = append(triggerFields, fmt.Sprintf("Timing: %s,", timing))
		if value.Trigger.CapturedObject.Exists {
			object, err := r.renderObjectReference(value.Trigger.CapturedObject.Val)
			if err != nil {
				return "", err
			}
			ctx.need(importOpt)
			triggerFields = append(triggerFields, fmt.Sprintf("CapturedObject: opt.Val(%s),", object))
		}
		if value.Trigger.CapturedObjectGroup.Exists {
			group, err := r.renderObjectReference(value.Trigger.CapturedObjectGroup.Val)
			if err != nil {
				return "", err
			}
			ctx.need(importOpt)
			triggerFields = append(triggerFields, fmt.Sprintf("CapturedObjectGroup: opt.Val(%s),", group))
		}
	}
	triggerFields = append(triggerFields, fmt.Sprintf("Content: %s,", content))
	if value.Trigger.Optional {
		triggerFields = append(triggerFields, "Optional: true,")
	}
	return structLit("game.CreateDelayedTrigger", []string{
		fmt.Sprintf("Trigger: %s,", structLit("game.DelayedTriggerDef", triggerFields)),
	}), nil
}

// renderCreateReflexiveTrigger renders a reflexive-trigger instruction (CR
// 603.11): the enabling action publishes its result, and this instruction — gated
// on that success — puts a reflexive triggered ability carrying the deferred
// consequence on the stack, its targets chosen then. The body is rendered exactly
// like any ability content, so its target selection resolves after the enabling
// action rather than up front.
func (r Renderer) renderCreateReflexiveTrigger(ctx *renderCtx, value game.CreateReflexiveTrigger) (string, error) {
	content, err := r.renderAbilityContent(ctx, value.Trigger.Content)
	if err != nil {
		return "", err
	}
	return structLit("game.CreateReflexiveTrigger", []string{
		fmt.Sprintf("Trigger: %s,", structLit("game.ReflexiveTriggerDef", []string{
			fmt.Sprintf("Content: %s,", content),
		})),
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
	if value.EntryTransformed {
		fields = append(fields, "EntryTransformed: true,")
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
	if len(value.LinkedReturnZones) > 0 {
		zones := make([]string, len(value.LinkedReturnZones))
		for i, z := range value.LinkedReturnZones {
			rendered, err := renderZone(z)
			if err != nil {
				return "", err
			}
			zones[i] = rendered
		}
		ctx.need(importZone)
		fields = append(fields, fmt.Sprintf("LinkedReturnZones: []zone.Type{%s},", strings.Join(zones, ", ")))
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
	if value.Counter.Exists {
		kind, err := renderCounterKind(value.Counter.Val)
		if err != nil {
			return "", err
		}
		ctx.need(importCounter)
		ctx.need(importOpt)
		fields = append(fields, fmt.Sprintf("Counter: opt.Val(%s),", kind))
	}
	if value.PublishLinked != "" {
		fields = append(fields, fmt.Sprintf("PublishLinked: game.LinkedKey(%q),", string(value.PublishLinked)))
	}
	if value.PublishLinkedObjectScoped {
		fields = append(fields, "PublishLinkedObjectScoped: true,")
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
	face, err := renderFaceIndex(value.Face)
	if err != nil {
		return "", err
	}
	ctx.need(importZone)
	return structLit("game.GrantCastPermission", []string{
		fmt.Sprintf("Card: %s,", card),
		fmt.Sprintf("FromZone: %s,", fromZone),
		fmt.Sprintf("Face: %s,", face),
		fmt.Sprintf("Duration: %s,", duration),
	}), nil
}

// renderFaceIndex renders a card face index to its game enum literal. Only the
// front and alternate faces appear in lowered cast permissions (a normal
// graveyard cast vs. an Adventure cast); any other face is unsupported.
func renderFaceIndex(face game.FaceIndex) (string, error) {
	switch face {
	case game.FaceFront:
		return "game.FaceFront", nil
	case game.FaceAlternate:
		return "game.FaceAlternate", nil
	default:
		return "", fmt.Errorf("render: unsupported cast-permission face %d", face)
	}
}

func (Renderer) renderExileForPlay(ctx *renderCtx, value game.ExileForPlay) (string, error) {
	fromZone, err := renderZone(value.FromZone)
	if err != nil {
		return "", err
	}
	duration, err := renderDuration(value.Duration)
	if err != nil {
		return "", err
	}
	ctx.need(importZone)
	var fields []string
	if value.SelectFromBatch {
		fields = append(fields, "SelectFromBatch: true,")
	} else {
		card, err := renderCardReference(value.Card)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("Card: %s,", card))
	}
	fields = append(fields,
		fmt.Sprintf("FromZone: %s,", fromZone),
		fmt.Sprintf("Duration: %s,", duration))
	if value.Cast {
		fields = append(fields, "Cast: true,")
	}
	return structLit("game.ExileForPlay", fields), nil
}

// renderExilePermanentForPlay renders an ExilePermanentForPlay primitive: the
// target permanent reference and, when set, the source-keyed linked set the
// exiled card joins (Prowl, Stoic Strategist).
func (r Renderer) renderExilePermanentForPlay(value game.ExilePermanentForPlay) (string, error) {
	object, err := r.renderObjectReference(value.Object)
	if err != nil {
		return "", err
	}
	fields := []string{fmt.Sprintf("Object: %s,", object)}
	if value.LinkedKey != "" {
		fields = append(fields, fmt.Sprintf("LinkedKey: game.LinkedKey(%q),", string(value.LinkedKey)))
	}
	return structLit("game.ExilePermanentForPlay", fields), nil
}

// renderCopyCard renders a CopyCard primitive: the copying player and the
// object-scoped imprint link key naming the exiled card to copy (Isochron
// Scepter, Spellbinder).
func (r Renderer) renderCopyCard(value game.CopyCard) (string, error) {
	player, err := r.renderPlayerReference(value.Player)
	if err != nil {
		return "", err
	}
	fields := []string{
		fmt.Sprintf("Player: %s,", player),
		fmt.Sprintf("LinkID: %q,", value.LinkID),
	}
	return structLit("game.CopyCard", fields), nil
}

// renderPlayLinkedExiledCard renders a PlayLinkedExiledCard primitive: the
// casting player, the object-scoped imprint link key, and the copy and
// free-cast riders (Isochron Scepter, Spellbinder).
func (r Renderer) renderPlayLinkedExiledCard(value game.PlayLinkedExiledCard) (string, error) {
	player, err := r.renderPlayerReference(value.Player)
	if err != nil {
		return "", err
	}
	fields := []string{
		fmt.Sprintf("Player: %s,", player),
		fmt.Sprintf("LinkID: %q,", value.LinkID),
	}
	if value.Copy {
		fields = append(fields, "Copy: true,")
	}
	if value.WithoutPayingManaCost {
		fields = append(fields, "WithoutPayingManaCost: true,")
	}
	return structLit("game.PlayLinkedExiledCard", fields), nil
}

// renderPlayChosenExiledCard renders a PlayChosenExiledCard primitive: the
// choosing player, the source zone, the owner scope, the optional marker-counter
// filter, the play window, and the free-cast rider (Dauthi Voidwalker).
func (r Renderer) renderPlayChosenExiledCard(ctx *renderCtx, value game.PlayChosenExiledCard) (string, error) {
	player, err := r.renderPlayerReference(value.Player)
	if err != nil {
		return "", err
	}
	zoneLit, err := renderZone(value.Zone)
	if err != nil {
		return "", err
	}
	scope, err := renderPlayerRelation(value.OwnerScope)
	if err != nil {
		return "", err
	}
	duration, err := renderDuration(value.Duration)
	if err != nil {
		return "", err
	}
	ctx.need(importZone)
	fields := []string{
		fmt.Sprintf("Player: %s,", player),
		fmt.Sprintf("Zone: %s,", zoneLit),
		fmt.Sprintf("OwnerScope: %s,", scope),
	}
	if value.Counter.Exists {
		kind, err := renderCounterKind(value.Counter.Val)
		if err != nil {
			return "", err
		}
		ctx.need(importCounter)
		ctx.need(importOpt)
		fields = append(fields, fmt.Sprintf("Counter: opt.Val(%s),", kind))
	}
	fields = append(fields, fmt.Sprintf("Duration: %s,", duration))
	if value.WithoutPayingManaCost {
		fields = append(fields, "WithoutPayingManaCost: true,")
	}
	return structLit("game.PlayChosenExiledCard", fields), nil
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
	return structLit("game.ImpulseExile", impulseExileFields(player, amount, duration, value)), nil
}

// impulseExileFields renders the ImpulseExile struct fields, appending the
// cast-only permission, the any-type-mana rider, and the source-keyed linked set
// only when the play permission sets them, so every other impulse exile renders
// unchanged.
func impulseExileFields(player, amount, duration string, value game.ImpulseExile) []string {
	fields := []string{
		fmt.Sprintf("Player: %s,", player),
		fmt.Sprintf("Amount: %s,", amount),
		fmt.Sprintf("Duration: %s,", duration),
	}
	if value.SpendAnyMana {
		fields = append(fields, "SpendAnyMana: true,")
	}
	if value.Cast {
		fields = append(fields, "Cast: true,")
	}
	if value.PublishLinked != "" {
		fields = append(fields, fmt.Sprintf("PublishLinked: game.LinkedKey(%q),", string(value.PublishLinked)))
	}
	return fields
}

func (r Renderer) renderExileLibraryUntilNonlandCast(value game.ExileLibraryUntilNonlandCast) (string, error) {
	player, err := r.renderPlayerReference(value.Player)
	if err != nil {
		return "", err
	}
	return structLit("game.ExileLibraryUntilNonlandCast", []string{
		fmt.Sprintf("Player: %s,", player),
	}), nil
}

func (r Renderer) renderExileTopEachLibraryCastFree(ctx *renderCtx, value game.ExileTopEachLibraryCastFree) (string, error) {
	amount, err := r.renderQuantity(ctx, value.Amount)
	if err != nil {
		return "", err
	}
	return structLit("game.ExileTopEachLibraryCastFree", []string{
		fmt.Sprintf("Amount: %s,", amount),
	}), nil
}

func (r Renderer) renderIterativeLibraryProcess(ctx *renderCtx, value game.IterativeLibraryProcess) (string, error) {
	player, err := r.renderPlayerReference(value.Player)
	if err != nil {
		return "", err
	}
	var stop string
	switch value.Stop {
	case game.IterativeLibraryStopChosenName:
		stop = "game.IterativeLibraryStopChosenName"
	case game.IterativeLibraryStopDuplicateName:
		stop = "game.IterativeLibraryStopDuplicateName"
	default:
		return "", fmt.Errorf("render: unsupported iterative library stop %d", value.Stop)
	}
	fields := []string{
		fmt.Sprintf("Player: %s,", player),
		fmt.Sprintf("Stop: %s,", stop),
	}
	if value.PreExile.IsDynamic() || value.PreExile.Value() != 0 {
		preExile, err := r.renderQuantity(ctx, value.PreExile)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("PreExile: %s,", preExile))
	}
	if value.ChooseName {
		fields = append(fields, "ChooseName: true,")
	}
	if value.Reveal {
		fields = append(fields, "Reveal: true,")
	}
	if value.OptionalTake {
		fields = append(fields, "OptionalTake: true,")
	}
	if value.AllowAbsentName {
		fields = append(fields, "AllowAbsentName: true,")
	}
	return structLit("game.IterativeLibraryProcess", fields), nil
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
	if len(value.KindChoices) != 0 {
		ctx.need(importCounter)
		amount, err := r.renderQuantity(ctx, value.Amount)
		if err != nil {
			return "", err
		}
		object, err := r.renderObjectReference(value.Object)
		if err != nil {
			return "", err
		}
		kinds := make([]string, 0, len(value.KindChoices))
		for _, kind := range value.KindChoices {
			lit, err := renderCounterKind(kind)
			if err != nil {
				return "", err
			}
			kinds = append(kinds, lit)
		}
		fields := []string{
			fmt.Sprintf("Amount: %s,", amount),
			fmt.Sprintf("Object: %s,", object),
			fmt.Sprintf("KindChoices: []counter.Kind{%s},", strings.Join(kinds, ", ")),
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
	if value.DoubleKind {
		fields = nil
	}
	if value.Group.Domain() != 0 {
		group, err := r.renderGroupReference(ctx, value.Group)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("Group: %s,", group))
		if value.ChooseOne {
			fields = append(fields, "ChooseOne: true,")
		}
	} else {
		object, err := r.renderObjectReference(value.Object)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("Object: %s,", object))
	}
	fields = append(fields, fmt.Sprintf("CounterKind: %s,", kind))
	if value.Distribute {
		fields = append(fields, "Distribute: true,")
	}
	if value.DoubleKind {
		fields = append(fields, "DoubleKind: true,")
	}
	if value.PublishLinked != "" {
		fields = append(fields, fmt.Sprintf("PublishLinked: game.LinkedKey(%q),", string(value.PublishLinked)))
	}
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
	if value.RecipientGroup.Kind != game.PlayerGroupReferenceNone {
		var group string
		switch value.RecipientGroup.Kind {
		case game.PlayerGroupReferenceOpponents:
			group = "game.OpponentsReference()"
		case game.PlayerGroupReferenceAllPlayers:
			group = "game.AllPlayersReference()"
		default:
			return "", fmt.Errorf("render: unsupported player group reference kind %d", value.RecipientGroup.Kind)
		}
		fields = append(fields, fmt.Sprintf("RecipientGroup: %s,", group))
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
// token CardDef var or a copy of the effect's target object. Copy specs may
// carry characteristic-overriding exceptions ("except it's a 1/1 green Frog",
// "except it's an artifact in addition to its other types"); their power,
// toughness, color, card-type, subtype, and keyword overrides are rendered
// below. Eternalize-style specs that rename, drop mana cost, or drop printed
// text are built directly in card code, never rendered, so they fail closed.
func (r Renderer) renderTokenSource(ctx *renderCtx, source game.TokenSource) (string, error) {
	if def, ok := source.TokenDefRef(); ok {
		return fmt.Sprintf("game.TokenDef(%s)", ctx.tokenDefVar(def)), nil
	}
	spec, ok := source.TokenCopy()
	if !ok || spec.SetName != "" || spec.NoManaCost || spec.NoPrintedText {
		return "", errors.New("render: unsupported CreateToken token source")
	}
	switch spec.Source {
	case game.TokenCopySourceObject:
		return r.renderTokenCopyObjectSource(ctx, spec)
	case game.TokenCopySourceEachInGroup:
		return r.renderTokenCopyForEachSource(ctx, spec)
	case game.TokenCopySourceChosenFromTriggerBatch:
		return renderTokenCopyTriggeringSetSource(ctx, spec)
	case game.TokenCopySourceChosenControlledCreatureToken:
		return renderTokenCopyPopulateSource(ctx, spec)
	default:
		return "", errors.New("render: unsupported CreateToken token source")
	}

}

func renderTokenCopyPopulateSource(ctx *renderCtx, spec game.TokenCopySpec) (string, error) {
	fields := []string{
		"Source: game.TokenCopySourceChosenControlledCreatureToken,",
	}
	fields, err := appendTokenCopyModifierFields(ctx, fields, spec)
	if err != nil {
		return "", err
	}
	rendered, err := renderTokenCopyKeywordField(fields, spec)
	if err != nil {
		return "", err
	}
	return "game.TokenCopyOf(" + structLit("game.TokenCopySpec", rendered) + ")", nil
}

func renderTokenCopyTriggeringSetSource(ctx *renderCtx, spec game.TokenCopySpec) (string, error) {
	fields := []string{
		"Source: game.TokenCopySourceChosenFromTriggerBatch,",
	}
	fields, err := appendTokenCopyModifierFields(ctx, fields, spec)
	if err != nil {
		return "", err
	}
	rendered, err := renderTokenCopyKeywordField(fields, spec)
	if err != nil {
		return "", err
	}
	return structLit("game.TokenCopyOf(game.TokenCopySpec", rendered) + ")", nil
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
	fields, err = appendTokenCopyModifierFields(ctx, fields, spec)
	if err != nil {
		return "", err
	}
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
	fields, err = appendTokenCopyModifierFields(ctx, fields, spec)
	if err != nil {
		return "", err
	}
	rendered, err := renderTokenCopyKeywordField(fields, spec)
	if err != nil {
		return "", err
	}
	return structLit("game.TokenCopyOf(game.TokenCopySpec", rendered) + ")", nil
}

// appendTokenCopyModifierFields renders the copy-token spec's
// characteristic-overriding exception fields: the legendary drop, the
// power/toughness override, the replacing color/type/subtype overrides ("except
// it's a 1/1 green Frog"), and the additive color/type/subtype overrides
// ("except it's an artifact in addition to its other types").
func appendTokenCopyModifierFields(ctx *renderCtx, fields []string, spec game.TokenCopySpec) ([]string, error) {
	if spec.SetNotLegendary {
		fields = append(fields, "SetNotLegendary: true,")
	}
	if spec.SetPower.Exists {
		ctx.need(importOpt)
		fields = append(fields, fmt.Sprintf("SetPower: opt.Val(%s),", renderPTValue(spec.SetPower.Val)))
	}
	if spec.SetToughness.Exists {
		ctx.need(importOpt)
		fields = append(fields, fmt.Sprintf("SetToughness: opt.Val(%s),", renderPTValue(spec.SetToughness.Val)))
	}
	if spec.HalvePowerToughnessRoundUp {
		fields = append(fields, "HalvePowerToughnessRoundUp: true,")
	}
	if len(spec.SetColors) != 0 {
		literal, err := renderColorSlice(ctx, spec.SetColors)
		if err != nil {
			return nil, err
		}
		fields = append(fields, fmt.Sprintf("SetColors: %s,", literal))
	}
	if len(spec.SetTypes) != 0 {
		literal, err := renderTypesCardSlice(ctx, spec.SetTypes)
		if err != nil {
			return nil, err
		}
		fields = append(fields, fmt.Sprintf("SetTypes: %s,", literal))
	}
	if len(spec.SetSubtypes) != 0 {
		literal, err := renderSubtypeSlice(ctx, spec.SetSubtypes)
		if err != nil {
			return nil, err
		}
		fields = append(fields, fmt.Sprintf("SetSubtypes: %s,", literal))
	}
	if len(spec.AddColors) != 0 {
		literal, err := renderColorSlice(ctx, spec.AddColors)
		if err != nil {
			return nil, err
		}
		fields = append(fields, fmt.Sprintf("AddColors: %s,", literal))
	}
	if len(spec.AddTypes) != 0 {
		literal, err := renderTypesCardSlice(ctx, spec.AddTypes)
		if err != nil {
			return nil, err
		}
		fields = append(fields, fmt.Sprintf("AddTypes: %s,", literal))
	}
	if len(spec.AddSubtypes) != 0 {
		literal, err := renderSubtypeSlice(ctx, spec.AddSubtypes)
		if err != nil {
			return nil, err
		}
		fields = append(fields, fmt.Sprintf("AddSubtypes: %s,", literal))
	}
	return fields, nil
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
	kind, err := renderCounterKind(value.CounterKind)
	if err != nil {
		return "", err
	}
	ctx.need(importCounter)
	amount, err := r.renderQuantity(ctx, value.Amount)
	if err != nil {
		return "", err
	}
	var reference string
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
		reference = fmt.Sprintf("PlayerGroup: %s,", group)
	} else {
		player, err := r.renderPlayerReference(value.Player)
		if err != nil {
			return "", err
		}
		reference = fmt.Sprintf("Player: %s,", player)
	}
	return structLit("game.AddPlayerCounter", []string{
		fmt.Sprintf("Amount: %s,", amount),
		reference,
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

// renderRemoveCounter renders a RemoveCounter primitive: a fixed amount removed
// of either a named CounterKind or a controller-chosen kind (ChooseKind),
// modeling the "remove a counter from target permanent" family (Ferropede).
// The kind-agnostic mass form (AllKinds) removes every counter regardless of
// kind ("Remove all counters from target permanent.", Vampire Hexmage) and
// carries neither an amount nor a kind. The counter is removed from a single
// referenced Object, or from every permanent of a Group when Group is set
// ("Remove a -1/-1 counter from each creature you control.", Heartmender).
func (r Renderer) renderRemoveCounter(ctx *renderCtx, value *game.RemoveCounter) (string, error) {
	if value.Group.Valid() {
		return r.renderRemoveCounterGroup(ctx, value)
	}
	object, err := r.renderObjectReference(value.Object)
	if err != nil {
		return "", err
	}
	if value.AllKinds {
		return structLit("game.RemoveCounter", []string{
			fmt.Sprintf("Object: %s,", object),
			"AllKinds: true,",
		}), nil
	}
	amount, err := r.renderQuantity(ctx, value.Amount)
	if err != nil {
		return "", err
	}
	fields := []string{
		fmt.Sprintf("Amount: %s,", amount),
		fmt.Sprintf("Object: %s,", object),
	}
	if value.ChooseKind {
		fields = append(fields, "ChooseKind: true,")
		return structLit("game.RemoveCounter", fields), nil
	}
	kind, err := renderCounterKind(value.CounterKind)
	if err != nil {
		return "", err
	}
	ctx.need(importCounter)
	fields = append(fields, fmt.Sprintf("CounterKind: %s,", kind))
	return structLit("game.RemoveCounter", fields), nil
}

// renderRemoveCounterGroup renders the group form of RemoveCounter, in which the
// counter is removed from every permanent selected by a GroupReference rather
// than from a single Object ("Remove a -1/-1 counter from each creature you
// control.", Heartmender). ChooseKind is ignored for group removals, matching
// the runtime, so only the named-kind and kind-agnostic (AllKinds) forms are
// rendered here.
func (r Renderer) renderRemoveCounterGroup(ctx *renderCtx, value *game.RemoveCounter) (string, error) {
	group, err := r.renderGroupReference(ctx, value.Group)
	if err != nil {
		return "", err
	}
	if value.AllKinds {
		return structLit("game.RemoveCounter", []string{
			fmt.Sprintf("Group: %s,", group),
			"AllKinds: true,",
		}), nil
	}
	amount, err := r.renderQuantity(ctx, value.Amount)
	if err != nil {
		return "", err
	}
	kind, err := renderCounterKind(value.CounterKind)
	if err != nil {
		return "", err
	}
	ctx.need(importCounter)
	return structLit("game.RemoveCounter", []string{
		fmt.Sprintf("Amount: %s,", amount),
		fmt.Sprintf("Group: %s,", group),
		fmt.Sprintf("CounterKind: %s,", kind),
	}), nil
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
	value, err := assertPrimitive[game.GroupSourceDamage](primitive)
	if err != nil {
		return "", err
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

func (r Renderer) renderGroupSelfPowerDamage(ctx *renderCtx, primitive game.Primitive) (string, error) {
	value, err := assertPrimitive[game.GroupSelfPowerDamage](primitive)
	if err != nil {
		return "", err
	}
	group, err := r.renderGroupReference(ctx, value.Group)
	if err != nil {
		return "", err
	}
	return structLit("game.GroupSelfPowerDamage", []string{
		fmt.Sprintf("Group: %s,", group),
	}), nil
}

func (r Renderer) renderDamagePrimitive(ctx *renderCtx, primitive game.Primitive) (string, error) {
	value, err := assertPrimitive[game.Damage](primitive)
	if err != nil {
		return "", err
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
	if value.ExcessRecipient.Valid() {
		excess, err := r.renderDamageRecipient(ctx, value.ExcessRecipient)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("ExcessRecipient: %s,", excess))
	}
	return structLit("game.Damage", fields), nil
}

func (r Renderer) renderExchangeLifeTotalWithSourceCharacteristic(
	value game.ExchangeLifeTotalWithSourceCharacteristic,
) (string, error) {
	player, err := r.renderPlayerReference(value.Player)
	if err != nil {
		return "", err
	}
	var characteristic string
	switch value.Characteristic {
	case game.SourcePower:
		characteristic = "game.SourcePower"
	case game.SourceToughness:
		characteristic = "game.SourceToughness"
	default:
		return "", fmt.Errorf("render: unsupported source power/toughness %d", value.Characteristic)
	}
	return structLit("game.ExchangeLifeTotalWithSourceCharacteristic", []string{
		fmt.Sprintf("Player: %s,", player),
		fmt.Sprintf("Characteristic: %s,", characteristic),
	}), nil
}

func (r Renderer) renderPlayerAmountPrimitive(ctx *renderCtx, primitive game.Primitive) (string, error) {
	var typeName string
	var amount game.Quantity
	var player game.PlayerReference
	switch primitive.Kind() {
	case game.PrimitiveDraw:
		value, err := assertPrimitive[game.Draw](primitive)
		if err != nil {
			return "", err
		}
		if value.PlayerGroup.Kind != game.PlayerGroupReferenceNone {
			return r.renderAmountPlayerGroup(ctx, "game.Draw", value.Amount, value.PlayerGroup)
		}
		typeName, amount, player = "game.Draw", value.Amount, value.Player
	case game.PrimitiveDiscard:
		value, err := assertPrimitive[game.Discard](primitive)
		if err != nil {
			return "", err
		}
		if value.EntireHand {
			return r.renderDiscardEntireHand(value)
		}
		if value.PlayerGroup.Kind != game.PlayerGroupReferenceNone {
			return r.renderDiscardAmountGroup(ctx, value)
		}
		if value.AtRandom {
			return r.renderDiscardAmountPlayer(ctx, value)
		}
		typeName, amount, player = "game.Discard", value.Amount, value.Player
	case game.PrimitiveMill:
		value, err := assertPrimitive[game.Mill](primitive)
		if err != nil {
			return "", err
		}
		if value.PlayerGroup.Kind != game.PlayerGroupReferenceNone {
			return r.renderAmountPlayerGroup(ctx, "game.Mill", value.Amount, value.PlayerGroup)
		}
		if value.PublishLinked != "" {
			return r.renderMillLinked(ctx, value)
		}
		typeName, amount, player = "game.Mill", value.Amount, value.Player
	case game.PrimitiveExileTopOfLibrary:
		value, err := assertPrimitive[game.ExileTopOfLibrary](primitive)
		if err != nil {
			return "", err
		}
		return r.renderExileTopOfLibrary(ctx, value)
	case game.PrimitiveScry:
		value, err := assertPrimitive[game.Scry](primitive)
		if err != nil {
			return "", err
		}
		typeName, amount, player = "game.Scry", value.Amount, value.Player
	case game.PrimitiveSurveil:
		value, err := assertPrimitive[game.Surveil](primitive)
		if err != nil {
			return "", err
		}
		typeName, amount, player = "game.Surveil", value.Amount, value.Player
	case game.PrimitiveGainLife:
		value, err := assertPrimitive[game.GainLife](primitive)
		if err != nil {
			return "", err
		}
		if value.PlayerGroup.Kind != game.PlayerGroupReferenceNone {
			return r.renderAmountPlayerGroup(ctx, "game.GainLife", value.Amount, value.PlayerGroup)
		}
		typeName, amount, player = "game.GainLife", value.Amount, value.Player
	case game.PrimitiveLoseLife:
		value, err := assertPrimitive[game.LoseLife](primitive)
		if err != nil {
			return "", err
		}
		if value.PlayerGroup.Kind != game.PlayerGroupReferenceNone {
			return r.renderAmountPlayerGroup(ctx, "game.LoseLife", value.Amount, value.PlayerGroup)
		}
		typeName, amount, player = "game.LoseLife", value.Amount, value.Player
	case game.PrimitiveReorderLibraryTop:
		value, err := assertPrimitive[game.ReorderLibraryTop](primitive)
		if err != nil {
			return "", err
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

func (r Renderer) renderBecomeMonarch(value game.BecomeMonarch) (string, error) {
	rendered, err := r.renderPlayerReference(value.Player)
	if err != nil {
		return "", err
	}
	return structLit("game.BecomeMonarch", []string{fmt.Sprintf("Player: %s,", rendered)}), nil
}

func (r Renderer) renderCantBecomeMonarch(value game.CantBecomeMonarch) (string, error) {
	rendered, err := r.renderPlayerReference(value.Player)
	if err != nil {
		return "", err
	}
	return structLit("game.CantBecomeMonarch", []string{fmt.Sprintf("Player: %s,", rendered)}), nil
}

func (Renderer) renderGainCityBlessing(_ game.GainCityBlessing) (string, error) {
	return "game.GainCityBlessing{}", nil
}

func (r Renderer) renderRingTempts(value game.RingTempts) (string, error) {
	rendered, err := r.renderPlayerReference(value.Player)
	if err != nil {
		return "", err
	}
	return structLit("game.RingTempts", []string{fmt.Sprintf("Player: %s,", rendered)}), nil
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

func (r Renderer) renderShuffleGraveyardIntoLibrary(value game.ShuffleGraveyardIntoLibrary) (string, error) {
	if value.PlayerGroup.Kind != game.PlayerGroupReferenceNone {
		group, err := renderPlayerGroupReference(value.PlayerGroup)
		if err != nil {
			return "", err
		}
		return structLit("game.ShuffleGraveyardIntoLibrary", []string{
			fmt.Sprintf("PlayerGroup: %s,", group),
		}), nil
	}
	player, err := r.renderPlayerReference(value.Player)
	if err != nil {
		return "", err
	}
	fields := []string{
		fmt.Sprintf("Player: %s,", player),
	}
	if value.IncludeHand {
		fields = append(fields, "IncludeHand: true,")
	}
	return structLit("game.ShuffleGraveyardIntoLibrary", fields), nil
}

func (r Renderer) renderMillLinked(ctx *renderCtx, value game.Mill) (string, error) {
	amount, err := r.renderQuantity(ctx, value.Amount)
	if err != nil {
		return "", err
	}
	player, err := r.renderPlayerReference(value.Player)
	if err != nil {
		return "", err
	}
	return structLit("game.Mill", []string{
		fmt.Sprintf("Amount: %s,", amount),
		fmt.Sprintf("Player: %s,", player),
		fmt.Sprintf("PublishLinked: game.LinkedKey(%q),", string(value.PublishLinked)),
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

func (r Renderer) renderChooseDiscardFromHand(ctx *renderCtx, value game.ChooseDiscardFromHand) (string, error) {
	player, err := r.renderPlayerReference(value.Player)
	if err != nil {
		return "", err
	}
	fields := []string{fmt.Sprintf("Player: %s,", player)}
	if value.ExcludeCreature {
		fields = append(fields, "ExcludeCreature: true,")
	}
	if value.ExcludeLand {
		fields = append(fields, "ExcludeLand: true,")
	}
	if value.MaxManaValue.Exists {
		ctx.need(importOpt)
		fields = append(fields, fmt.Sprintf("MaxManaValue: opt.Val(%d),", value.MaxManaValue.Val))
	}
	renderedSelection, err := r.renderSelection(ctx, value.Selection)
	if err != nil {
		return "", err
	}
	if renderedSelection != "game.Selection{}" {
		fields = append(fields, fmt.Sprintf("Selection: %s,", renderedSelection))
	}
	return structLit("game.ChooseDiscardFromHand", fields), nil
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
		value, err := assertPrimitive[game.Investigate](primitive)
		if err != nil {
			return "", err
		}
		amount, err := r.renderQuantity(ctx, value.Amount)
		if err != nil {
			return "", err
		}
		return structLit("game.Investigate", []string{fmt.Sprintf("Amount: %s,", amount)}), nil
	case game.PrimitiveProliferate:
		value, err := assertPrimitive[game.Proliferate](primitive)
		if err != nil {
			return "", err
		}
		amount, err := r.renderQuantity(ctx, value.Amount)
		if err != nil {
			return "", err
		}
		return structLit("game.Proliferate", []string{fmt.Sprintf("Amount: %s,", amount)}), nil
	case game.PrimitiveDiscoverCards:
		value, err := assertPrimitive[game.DiscoverCards](primitive)
		if err != nil {
			return "", err
		}
		amount, err := r.renderQuantity(ctx, value.Amount)
		if err != nil {
			return "", err
		}
		return structLit("game.DiscoverCards", []string{fmt.Sprintf("Amount: %s,", amount)}), nil
	case game.PrimitiveManifest:
		value, err := assertPrimitive[game.Manifest](primitive)
		if err != nil {
			return "", err
		}
		var fields []string
		if value.Dread {
			fields = append(fields, "Dread: true,")
		}
		if value.Cloak {
			fields = append(fields, "Cloak: true,")
		}
		if value.Player.Kind() != game.PlayerReferenceNone {
			player, err := r.renderPlayerReference(value.Player)
			if err != nil {
				return "", err
			}
			fields = append(fields, fmt.Sprintf("Player: %s,", player))
		}
		if value.PublishLinked != "" {
			fields = append(fields, fmt.Sprintf("PublishLinked: game.LinkedKey(%q),", string(value.PublishLinked)))
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

// renderBolster renders a Bolster primitive, emitting its fixed +1/+1 counter
// count and, when present, the linked key under which the chosen creature is
// published for a later linked effect.
func (r Renderer) renderBolster(ctx *renderCtx, value game.Bolster) (string, error) {
	amount, err := r.renderQuantity(ctx, value.Amount)
	if err != nil {
		return "", err
	}
	fields := []string{fmt.Sprintf("Amount: %s,", amount)}
	if value.PublishLinked != "" {
		fields = append(fields, fmt.Sprintf("PublishLinked: game.LinkedKey(%q),", string(value.PublishLinked)))
	}
	return structLit("game.Bolster", fields), nil
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

// renderAdapt renders an Adapt primitive, emitting the adapted-creature
// reference and the fixed +1/+1 counter count.
func (r Renderer) renderAdapt(ctx *renderCtx, value game.Adapt) (string, error) {
	object, err := r.renderObjectReference(value.Object)
	if err != nil {
		return "", err
	}
	amount, err := r.renderQuantity(ctx, value.Amount)
	if err != nil {
		return "", err
	}
	return structLit("game.Adapt", []string{
		fmt.Sprintf("Object: %s,", object),
		fmt.Sprintf("Amount: %s,", amount),
	}), nil
}

// renderMonstrosity renders a Monstrosity primitive, emitting the source
// permanent reference and the fixed +1/+1 counter count.
func (r Renderer) renderMonstrosity(ctx *renderCtx, value game.Monstrosity) (string, error) {
	object, err := r.renderObjectReference(value.Object)
	if err != nil {
		return "", err
	}
	amount, err := r.renderQuantity(ctx, value.Amount)
	if err != nil {
		return "", err
	}
	return structLit("game.Monstrosity", []string{
		fmt.Sprintf("Object: %s,", object),
		fmt.Sprintf("Amount: %s,", amount),
	}), nil
}

// renderConnive renders a Connive primitive, emitting the conniving permanent's
// object reference, the drawing/discarding player reference, and the fixed
// connive count.
func (r Renderer) renderConnive(ctx *renderCtx, value game.Connive) (string, error) {
	object, err := r.renderObjectReference(value.Object)
	if err != nil {
		return "", err
	}
	player, err := r.renderPlayerReference(value.Player)
	if err != nil {
		return "", err
	}
	amount, err := r.renderQuantity(ctx, value.Amount)
	if err != nil {
		return "", err
	}
	return structLit("game.Connive", []string{
		fmt.Sprintf("Object: %s,", object),
		fmt.Sprintf("Player: %s,", player),
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
	if value.Filter.Exists {
		selection, selErr := r.renderSelection(ctx, value.Filter.Val)
		if selErr != nil {
			return "", selErr
		}
		ctx.need(importOpt)
		fields = append(fields, fmt.Sprintf("Filter: opt.Val(%s),", selection))
	}
	if value.TakeUpTo {
		fields = append(fields, "TakeUpTo: true,")
	}
	if value.Reveal {
		fields = append(fields, "Reveal: true,")
	}
	if value.Destination != zone.None {
		destination, zoneErr := renderZone(value.Destination)
		if zoneErr != nil {
			return "", zoneErr
		}
		ctx.need(importZone)
		fields = append(fields, fmt.Sprintf("Destination: %s,", destination))
	}
	if value.EntersTapped {
		fields = append(fields, "EntersTapped: true,")
	}
	return structLit("game.Dig", fields), nil
}

func (r Renderer) renderObjectOrGroupPrimitive(ctx *renderCtx, primitive game.Primitive) (string, error) {
	switch primitive.Kind() {
	case game.PrimitiveDestroy:
		value, err := assertPrimitive[game.Destroy](primitive)
		if err != nil {
			return "", err
		}
		return r.renderDestroy(ctx, value)
	case game.PrimitiveBounce:
		value, err := assertPrimitive[game.Bounce](primitive)
		if err != nil {
			return "", err
		}
		return r.renderBounce(ctx, value)
	case game.PrimitiveUntap:
		value, err := assertPrimitive[game.Untap](primitive)
		if err != nil {
			return "", err
		}
		return r.renderUntap(ctx, value)
	case game.PrimitiveTap:
		value, err := assertPrimitive[game.Tap](primitive)
		if err != nil {
			return "", err
		}
		return r.renderObjectOrGroup(ctx, "game.Tap", value.Object, value.Group)
	case game.PrimitiveSkipNextUntap:
		value, err := assertPrimitive[game.SkipNextUntap](primitive)
		if err != nil {
			return "", err
		}
		return r.renderObjectOrGroup(ctx, "game.SkipNextUntap", value.Object, value.Group)
	case game.PrimitiveTapOrUntap:
		value, err := assertPrimitive[game.TapOrUntap](primitive)
		if err != nil {
			return "", err
		}
		return r.renderObjectOrGroup(ctx, "game.TapOrUntap", value.Object, game.GroupReference{})
	case game.PrimitiveExile:
		value, err := assertPrimitive[game.Exile](primitive)
		if err != nil {
			return "", err
		}
		return r.renderExile(ctx, value)
	case game.PrimitivePhaseOut:
		value, err := assertPrimitive[game.PhaseOut](primitive)
		if err != nil {
			return "", err
		}
		return r.renderObjectOrGroup(ctx, "game.PhaseOut", value.Object, value.Group)
	case game.PrimitiveRegenerate:
		value, err := assertPrimitive[game.Regenerate](primitive)
		if err != nil {
			return "", err
		}
		return r.renderObjectOrGroup(ctx, "game.Regenerate", value.Object, value.Group)
	case game.PrimitiveGoad:
		value, err := assertPrimitive[game.Goad](primitive)
		if err != nil {
			return "", err
		}
		return r.renderObjectOrGroup(ctx, "game.Goad", value.Object, value.Group)
	case game.PrimitiveSacrifice:
		value, err := assertPrimitive[game.Sacrifice](primitive)
		if err != nil {
			return "", err
		}
		return r.renderSacrifice(ctx, value)
	default:
		return "", fmt.Errorf("render: unsupported object or group primitive kind %d", primitive.Kind())
	}
}

func (r Renderer) renderSacrifice(ctx *renderCtx, value game.Sacrifice) (string, error) {
	if !value.ByItsController {
		return r.renderObjectOrGroup(ctx, "game.Sacrifice", value.Object, value.Group)
	}
	// "That object's current controller sacrifices it" applies to the
	// single-object form only, so render the object reference alongside the flag.
	rendered, err := r.renderObjectReference(value.Object)
	if err != nil {
		return "", err
	}
	return structLit("game.Sacrifice", []string{
		fmt.Sprintf("Object: %s,", rendered),
		"ByItsController: true,",
	}), nil
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
	// A linked exile over a group (mass blink) remembers every exiled permanent
	// under the key so a later linked return brings the whole group back; render
	// the group reference alongside the key rather than a single object.
	if value.Group.Domain() != 0 {
		rendered, err := r.renderGroupReference(ctx, value.Group)
		if err != nil {
			return "", err
		}
		return structLit("game.Exile", []string{
			fmt.Sprintf("Group: %s,", rendered),
			fmt.Sprintf("ExileLinkedKey: game.LinkedKey(%q),", string(value.ExileLinkedKey)),
		}), nil
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
	case game.PrimitiveExplore:
		value, err := assertPrimitive[game.Explore](primitive)
		if err != nil {
			return "", err
		}
		fieldName = "Creature"
		typeName, object = "game.Explore", value.Creature
	case game.PrimitiveCounterObject:
		value, err := assertPrimitive[game.CounterObject](primitive)
		if err != nil {
			return "", err
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
		if value.Destination != game.CounteredSpellGraveyard {
			literal, err := counteredSpellDestinationLiteral(value.Destination)
			if err != nil {
				return "", err
			}
			rendered, err := r.renderObjectReference(value.Object)
			if err != nil {
				return "", err
			}
			return structLit("game.CounterObject", []string{
				fmt.Sprintf("Object: %s,", rendered),
				fmt.Sprintf("Destination: %s,", literal),
			}), nil
		}
		typeName, object = "game.CounterObject", value.Object
	case game.PrimitiveChooseNewTargets:
		value, err := assertPrimitive[game.ChooseNewTargets](primitive)
		if err != nil {
			return "", err
		}
		typeName, object = "game.ChooseNewTargets", value.Object
	case game.PrimitiveRemoveFromCombat:
		value, err := assertPrimitive[game.RemoveFromCombat](primitive)
		if err != nil {
			return "", err
		}
		typeName, object = "game.RemoveFromCombat", value.Object
	case game.PrimitiveTransform:
		value, err := assertPrimitive[game.Transform](primitive)
		if err != nil {
			return "", err
		}
		typeName, object = "game.Transform", value.Object
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
	if value.Chooser.Exists {
		chooser, err := r.renderPlayerReference(value.Chooser.Val)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("Chooser: opt.Val(%s),", chooser))
	}
	return structLit("game.CopyStackObject", fields), nil
}

func (r Renderer) renderFightPrimitive(primitive game.Primitive) (string, error) {
	value, err := assertPrimitive[game.Fight](primitive)
	if err != nil {
		return "", err
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
	value, err := assertPrimitive[game.Attach](primitive)
	if err != nil {
		return "", err
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
		renderedGroup, err := renderPlayerGroupReference(value.PlayerGroup)
		if err != nil {
			return "", err
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
	renderedGroup, err := renderPlayerGroupReference(group)
	if err != nil {
		return "", err
	}
	return structLit(typeName, []string{
		fmt.Sprintf("Amount: %s,", renderedAmount),
		fmt.Sprintf("PlayerGroup: %s,", renderedGroup),
	}), nil
}

// renderExileTopOfLibrary renders an ExileTopOfLibrary primitive, preserving the
// optional named exile counter the shared amount/player renderers drop. Without
// a counter it renders byte-identically to the shared amount/player(-group)
// renderer (Counter omitted when unset).
func (r Renderer) renderExileTopOfLibrary(ctx *renderCtx, value game.ExileTopOfLibrary) (string, error) {
	renderedAmount, err := r.renderQuantity(ctx, value.Amount)
	if err != nil {
		return "", err
	}
	fields := []string{fmt.Sprintf("Amount: %s,", renderedAmount)}
	if value.PlayerGroup.Kind != game.PlayerGroupReferenceNone {
		renderedGroup, err := renderPlayerGroupReference(value.PlayerGroup)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("PlayerGroup: %s,", renderedGroup))
	} else {
		player, err := r.renderPlayerReference(value.Player)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("Player: %s,", player))
	}
	if value.Counter.Exists {
		kind, err := renderCounterKind(value.Counter.Val)
		if err != nil {
			return "", err
		}
		ctx.need(importCounter)
		ctx.need(importOpt)
		fields = append(fields, fmt.Sprintf("Counter: opt.Val(%s),", kind))
	}
	if value.FaceDown {
		fields = append(fields, "FaceDown: true,")
	}
	return structLit("game.ExileTopOfLibrary", fields), nil
}

// renderDiscardAmountPlayer renders a single-player discard, preserving the
// "at random" selection flag the shared amount/player renderer drops. The
// non-random case routes through renderAmountPlayer, so only the random variant
// reaches here.
func (r Renderer) renderDiscardAmountPlayer(ctx *renderCtx, value game.Discard) (string, error) {
	renderedAmount, err := r.renderQuantity(ctx, value.Amount)
	if err != nil {
		return "", err
	}
	player, err := r.renderPlayerReference(value.Player)
	if err != nil {
		return "", err
	}
	fields := []string{
		fmt.Sprintf("Amount: %s,", renderedAmount),
		fmt.Sprintf("Player: %s,", player),
	}
	if value.AtRandom {
		fields = append(fields, "AtRandom: true,")
	}
	return structLit("game.Discard", fields), nil
}

// renderDiscardAmountGroup renders a player-group discard, preserving the "at
// random" selection flag. A non-random group discard renders byte-identically to
// the shared amount/player-group renderer (AtRandom omitted when false).
func (r Renderer) renderDiscardAmountGroup(ctx *renderCtx, value game.Discard) (string, error) {
	renderedAmount, err := r.renderQuantity(ctx, value.Amount)
	if err != nil {
		return "", err
	}
	renderedGroup, err := renderPlayerGroupReference(value.PlayerGroup)
	if err != nil {
		return "", err
	}
	fields := []string{
		fmt.Sprintf("Amount: %s,", renderedAmount),
		fmt.Sprintf("PlayerGroup: %s,", renderedGroup),
	}
	if value.AtRandom {
		fields = append(fields, "AtRandom: true,")
	}
	return structLit("game.Discard", fields), nil
}

func (r Renderer) renderSacrificePermanents(ctx *renderCtx, value *game.SacrificePermanents) (string, error) {
	renderedSelection, err := r.renderSelection(ctx, value.Selection)
	if err != nil {
		return "", err
	}
	var fields []string
	switch {
	case value.All:
		fields = append(fields, "All: true,")
	case value.AnyNumber:
		fields = append(fields, "AnyNumber: true,")
	default:
		renderedAmount, err := r.renderQuantity(ctx, value.Amount)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("Amount: %s,", renderedAmount))
	}
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
	if value.PublishLinked != "" {
		fields = append(fields, fmt.Sprintf("PublishLinked: game.LinkedKey(%q),", string(value.PublishLinked)))
	}
	if value.PublishObjectBinding {
		fields = append(fields, "PublishObjectBinding: true,")
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
	if value.MatchToDestinationRestRandomBottom {
		fields = append(fields, "MatchToDestinationRestRandomBottom: true,")
	}
	return structLit("game.RevealUntil", fields), nil
}

func (r Renderer) renderPileSplit(ctx *renderCtx, value *game.PileSplit) (string, error) {
	player, err := r.renderPlayerReference(value.Player)
	if err != nil {
		return "", err
	}
	renderedAmount, err := r.renderQuantity(ctx, value.Amount)
	if err != nil {
		return "", err
	}
	kept, err := renderZone(value.Kept)
	if err != nil {
		return "", err
	}
	other, err := renderZone(value.Other)
	if err != nil {
		return "", err
	}
	ctx.need(importZone)
	fields := []string{
		fmt.Sprintf("Player: %s,", player),
		fmt.Sprintf("Amount: %s,", renderedAmount),
	}
	if value.SeparatorOpponent {
		fields = append(fields, "SeparatorOpponent: true,")
	}
	if value.ChooserOpponent {
		fields = append(fields, "ChooserOpponent: true,")
	}
	fields = append(fields,
		fmt.Sprintf("Kept: %s,", kept),
		fmt.Sprintf("Other: %s,", other),
	)
	return structLit("game.PileSplit", fields), nil
}

func (r Renderer) renderRevealTopPartition(ctx *renderCtx, value *game.RevealTopPartition) (string, error) {
	player, err := r.renderPlayerReference(value.Player)
	if err != nil {
		return "", err
	}
	renderedAmount, err := r.renderQuantity(ctx, value.Amount)
	if err != nil {
		return "", err
	}
	selection, err := r.renderSelection(ctx, value.Selection)
	if err != nil {
		return "", err
	}
	fields := []string{
		fmt.Sprintf("Player: %s,", player),
		fmt.Sprintf("Amount: %s,", renderedAmount),
		fmt.Sprintf("Selection: %s,", selection),
	}
	if value.Remainder == game.DigRemainderLibraryBottom {
		fields = append(fields, "Remainder: game.DigRemainderLibraryBottom,")
	}
	return structLit("game.RevealTopPartition", fields), nil
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
	if value.DiscardCount > 0 {
		fields = append(fields, fmt.Sprintf("DiscardCount: %d,", value.DiscardCount))
	}
	if value.ControllerDrawEach {
		fields = append(fields, "ControllerDrawEach: true,")
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
	fields := []string{
		"ChooseUpTo: true,",
		fmt.Sprintf("Amount: %s,", renderedAmount),
		fmt.Sprintf("Group: %s,", renderedGroup),
	}
	if value.Chooser.Kind() != game.PlayerReferenceNone {
		renderedChooser, err := r.renderPlayerReference(value.Chooser)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("Chooser: %s,", renderedChooser))
	}
	return structLit("game.Untap", fields), nil
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
	if len(value.CombinationColors) > 0 {
		colorsLiteral, err := renderManaColorSlice(ctx, value.CombinationColors)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("CombinationColors: %s,", colorsLiteral))
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
	if value.PersistUntilEndOfTurn {
		fields = append(fields, "PersistUntilEndOfTurn: true,")
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
	if _, ok := value.AnyTarget.AnyTargetObjectReference(); ok {
		recipient, err := r.renderDamageRecipient(ctx, value.AnyTarget)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("AnyTarget: %s,", recipient))
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
	if value.OneShot {
		fields = append(fields, "OneShot: true,")
	}
	if len(value.SourceColors) > 0 {
		colors, err := renderColorSlice(ctx, value.SourceColors)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("SourceColors: %s,", colors))
	}
	if value.RedirectPreventedToSourceController {
		fields = append(fields, "RedirectPreventedToSourceController: true,")
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
