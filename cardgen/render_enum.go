package cardgen

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func renderManaSymbol(ctx *renderCtx, symbol cost.Symbol) (string, error) {
	ctx.need(importCost)
	switch symbol.Kind {
	case cost.ColoredSymbol:
		switch symbol.Color {
		case mana.W:
			return "cost.W", nil
		case mana.U:
			return "cost.U", nil
		case mana.B:
			return "cost.B", nil
		case mana.R:
			return "cost.R", nil
		case mana.G:
			return "cost.G", nil
		default:
			return "", fmt.Errorf("render: unsupported colored mana symbol %q", string(symbol.Color))
		}
	case cost.GenericSymbol:
		return fmt.Sprintf("cost.O(%d)", symbol.Generic), nil
	case cost.ColorlessSymbol:
		return "cost.C", nil
	case cost.VariableSymbol:
		return "cost.X", nil
	case cost.SnowSymbol:
		return "cost.S", nil
	case cost.HybridSymbol:
		ctx.need(importMana)
		first, err := renderManaColor(symbol.Color)
		if err != nil {
			return "", fmt.Errorf("render: unsupported hybrid mana color: %w", err)
		}
		second, err := renderManaColor(symbol.AltColor)
		if err != nil {
			return "", fmt.Errorf("render: unsupported hybrid mana alt color: %w", err)
		}
		return fmt.Sprintf("cost.HybridMana(%s, %s)", first, second), nil
	case cost.TwobridSymbol:
		ctx.need(importMana)
		c, err := renderManaColor(symbol.Color)
		if err != nil {
			return "", fmt.Errorf("render: unsupported twobrid mana color: %w", err)
		}
		return fmt.Sprintf("cost.Twobrid(%s)", c), nil
	case cost.PhyrexianSymbol:
		ctx.need(importMana)
		c, err := renderManaColor(symbol.Color)
		if err != nil {
			return "", fmt.Errorf("render: unsupported phyrexian mana color: %w", err)
		}
		return fmt.Sprintf("cost.PhyrexianMana(%s)", c), nil
	default:
		return "", fmt.Errorf("render: unsupported mana symbol kind %d", symbol.Kind)
	}
}

func renderManaColor(c mana.Color) (string, error) {
	switch c {
	case mana.W:
		return "mana.W", nil
	case mana.U:
		return "mana.U", nil
	case mana.B:
		return "mana.B", nil
	case mana.R:
		return "mana.R", nil
	case mana.G:
		return "mana.G", nil
	case mana.C:
		return "mana.C", nil
	default:
		return "", fmt.Errorf("render: unsupported mana color %q", string(c))
	}
}

func renderManaColorSlice(ctx *renderCtx, colors []mana.Color) (string, error) {
	ctx.need(importMana)
	literals := make([]string, 0, len(colors))
	for _, c := range colors {
		literal, err := renderManaColor(c)
		if err != nil {
			return "", err
		}
		literals = append(literals, literal)
	}
	return "[]mana.Color{" + strings.Join(literals, ", ") + "}", nil
}

func renderTypesCardSlice(ctx *renderCtx, cardTypes []types.Card) (string, error) {
	ctx.need(importTypes)
	literals := make([]string, 0, len(cardTypes))
	for _, cardType := range cardTypes {
		lit, err := cardTypeLiteral(cardType)
		if err != nil {
			return "", err
		}
		literals = append(literals, lit)
	}
	return "[]types.Card{" + strings.Join(literals, ", ") + "}", nil
}

// cardTypeLiteral returns the Go constant for a types.Card value. It errors for
// any card type not known to the renderer's supported subset, preventing silent
// emission of comment fallbacks.
func cardTypeLiteral(t types.Card) (string, error) {
	lit := CardTypeToLiteral(string(t))
	if strings.HasPrefix(lit, "/*") {
		return "", fmt.Errorf("render: unsupported card type %q", string(t))
	}
	return lit, nil
}

// supertypeLiteral returns the Go constant for a types.Super value. It errors
// for any supertype not known to the renderer's supported subset.
func supertypeLiteral(st types.Super) (string, error) {
	lit := SupertypeToLiteral(string(st))
	if strings.HasPrefix(lit, "/*") {
		return "", fmt.Errorf("render: unsupported supertype %q", string(st))
	}
	return lit, nil
}

func renderAdditionalKind(kind cost.AdditionalKind) (string, error) {
	switch kind {
	case cost.AdditionalSacrifice:
		return "cost.AdditionalSacrifice", nil
	case cost.AdditionalSacrificeSource:
		return "cost.AdditionalSacrificeSource", nil
	case cost.AdditionalDiscard:
		return "cost.AdditionalDiscard", nil
	case cost.AdditionalPayLife:
		return "cost.AdditionalPayLife", nil
	case cost.AdditionalExile:
		return "cost.AdditionalExile", nil
	case cost.AdditionalReveal:
		return "cost.AdditionalReveal", nil
	case cost.AdditionalTap:
		return "cost.AdditionalTap", nil
	case cost.AdditionalExileSource:
		return "cost.AdditionalExileSource", nil
	case cost.AdditionalUntap:
		return "cost.AdditionalUntap", nil
	case cost.AdditionalRemoveCounter:
		return "cost.AdditionalRemoveCounter", nil
	case cost.AdditionalReturnUnblockedAttacker:
		return "cost.AdditionalReturnUnblockedAttacker", nil
	case cost.AdditionalTapPermanents:
		return "cost.AdditionalTapPermanents", nil
	case cost.AdditionalEnergy:
		return "cost.AdditionalEnergy", nil
	case cost.AdditionalReturnToHand:
		return "cost.AdditionalReturnToHand", nil
	case cost.AdditionalExert:
		return "cost.AdditionalExert", nil
	case cost.AdditionalMill:
		return "cost.AdditionalMill", nil
	case cost.AdditionalPutCounter:
		return "cost.AdditionalPutCounter", nil
	case cost.AdditionalCollectEvidence:
		return "cost.AdditionalCollectEvidence", nil
	default:
		return "", fmt.Errorf("render: unsupported additional cost kind %d", kind)
	}
}

func renderAdditionalDynamicAmount(kind cost.AdditionalDynamicAmount) (string, error) {
	switch kind {
	case cost.AdditionalDynamicCommanderColorIdentityCount:
		return "cost.AdditionalDynamicCommanderColorIdentityCount", nil
	default:
		return "", fmt.Errorf("render: unsupported additional dynamic amount %d", kind)
	}
}

func renderCounterKind(kind counter.Kind) (string, error) {
	switch kind {
	case counter.PlusOnePlusOne:
		return "counter.PlusOnePlusOne", nil
	case counter.MinusOneMinusOne:
		return "counter.MinusOneMinusOne", nil
	case counter.Charge:
		return "counter.Charge", nil
	case counter.Loyalty:
		return "counter.Loyalty", nil
	case counter.Time:
		return "counter.Time", nil
	case counter.Defense:
		return "counter.Defense", nil
	case counter.Poison:
		return "counter.Poison", nil
	case counter.Lore:
		return "counter.Lore", nil
	case counter.Verse:
		return "counter.Verse", nil
	case counter.Shield:
		return "counter.Shield", nil
	case counter.Stun:
		return "counter.Stun", nil
	case counter.Finality:
		return "counter.Finality", nil
	case counter.Brick:
		return "counter.Brick", nil
	case counter.Page:
		return "counter.Page", nil
	case counter.Enlightened:
		return "counter.Enlightened", nil
	case counter.Oil:
		return "counter.Oil", nil
	case counter.Blood:
		return "counter.Blood", nil
	case counter.Indestructible:
		return "counter.Indestructible", nil
	case counter.Deathtouch:
		return "counter.Deathtouch", nil
	case counter.Flying:
		return "counter.Flying", nil
	case counter.FirstStrike:
		return "counter.FirstStrike", nil
	case counter.Hexproof:
		return "counter.Hexproof", nil
	case counter.Lifelink:
		return "counter.Lifelink", nil
	case counter.Menace:
		return "counter.Menace", nil
	case counter.Reach:
		return "counter.Reach", nil
	case counter.Trample:
		return "counter.Trample", nil
	case counter.Vigilance:
		return "counter.Vigilance", nil
	case counter.Energy:
		return "counter.Energy", nil
	case counter.Experience:
		return "counter.Experience", nil
	case counter.Burden:
		return "counter.Burden", nil
	case counter.Age:
		return "counter.Age", nil
	default:
		return "", fmt.Errorf("render: unsupported counter kind %d", kind)
	}
}

func renderTargetAllow(allow game.TargetAllow) string {
	var parts []string
	if allow&game.TargetAllowPermanent != 0 {
		parts = append(parts, "game.TargetAllowPermanent")
	}
	if allow&game.TargetAllowPlayer != 0 {
		parts = append(parts, "game.TargetAllowPlayer")
	}
	if allow&game.TargetAllowStackObject != 0 {
		parts = append(parts, "game.TargetAllowStackObject")
	}
	if allow&game.TargetAllowCard != 0 {
		parts = append(parts, "game.TargetAllowCard")
	}
	if len(parts) == 0 {
		return "game.TargetAllowUnspecified"
	}
	return strings.Join(parts, " | ")
}

func renderPlayerRelation(relation game.PlayerRelation) (string, error) {
	switch relation {
	case game.PlayerAny:
		return "game.PlayerAny", nil
	case game.PlayerYou:
		return "game.PlayerYou", nil
	case game.PlayerOpponent:
		return "game.PlayerOpponent", nil
	case game.PlayerNotYou:
		return "game.PlayerNotYou", nil
	default:
		return "", fmt.Errorf("render: unsupported player relation %d", relation)
	}
}

func renderTriggerType(triggerType game.TriggerType) (string, error) {
	switch triggerType {
	case game.TriggerWhen:
		return "game.TriggerWhen", nil
	case game.TriggerWhenever:
		return "game.TriggerWhenever", nil
	case game.TriggerAt:
		return "game.TriggerAt", nil
	case game.TriggerState:
		return "game.TriggerState", nil
	default:
		return "", fmt.Errorf("render: unsupported trigger type %d", triggerType)
	}
}

func renderStep(step game.Step) (string, error) {
	switch step {
	case game.StepUpkeep:
		return "game.StepUpkeep", nil
	case game.StepDraw:
		return "game.StepDraw", nil
	case game.StepBeginningOfCombat:
		return "game.StepBeginningOfCombat", nil
	case game.StepEndOfCombat:
		return "game.StepEndOfCombat", nil
	case game.StepEnd:
		return "game.StepEnd", nil
	case game.StepPrecombatMain:
		return "game.StepPrecombatMain", nil
	case game.StepPostcombatMain:
		return "game.StepPostcombatMain", nil
	default:
		return "", fmt.Errorf("render: unsupported step %d", step)
	}
}

func renderTriggerSource(source game.TriggerSourceFilter) (string, error) {
	switch source {
	case game.TriggerSourceSelf:
		return "game.TriggerSourceSelf", nil
	case game.TriggerSourceAttachedPermanent:
		return "game.TriggerSourceAttachedPermanent", nil
	default:
		return "", fmt.Errorf("render: unsupported trigger source %d", source)
	}
}

func renderTriggerSubject(subject game.TriggerSubjectObject) (string, error) {
	switch subject {
	case game.TriggerSubjectPermanent:
		return "game.TriggerSubjectPermanent", nil
	case game.TriggerSubjectBlockedAttacker:
		return "game.TriggerSubjectBlockedAttacker", nil
	case game.TriggerSubjectDamageSource:
		return "game.TriggerSubjectDamageSource", nil
	default:
		return "", fmt.Errorf("render: unsupported trigger subject %d", subject)
	}
}

func renderTriggerController(controller game.TriggerControllerFilter) (string, error) {
	switch controller {
	case game.TriggerControllerYou:
		return "game.TriggerControllerYou", nil
	case game.TriggerControllerOpponent:
		return "game.TriggerControllerOpponent", nil
	default:
		return "", fmt.Errorf("render: unsupported trigger controller filter %d", controller)
	}
}

func renderTriggerPlayer(player game.TriggerPlayerFilter) (string, error) {
	switch player {
	case game.TriggerPlayerYou:
		return "game.TriggerPlayerYou", nil
	case game.TriggerPlayerOpponent:
		return "game.TriggerPlayerOpponent", nil
	default:
		return "", fmt.Errorf("render: unsupported trigger player filter %d", player)
	}
}

func renderEventKind(event game.EventKind) (string, error) {
	switch event {
	case game.EventDamageDealt:
		return "game.EventDamageDealt", nil
	case game.EventCardDrawn:
		return "game.EventCardDrawn", nil
	case game.EventAttackerBecameBlocked:
		return "game.EventAttackerBecameBlocked", nil
	case game.EventAttackerDeclared:
		return "game.EventAttackerDeclared", nil
	case game.EventBlockerDeclared:
		return "game.EventBlockerDeclared", nil
	case game.EventSpellCast:
		return "game.EventSpellCast", nil
	case game.EventLifeGained:
		return "game.EventLifeGained", nil
	case game.EventLifeLost:
		return "game.EventLifeLost", nil
	case game.EventPermanentEnteredBattlefield:
		return "game.EventPermanentEnteredBattlefield", nil
	case game.EventPermanentDied:
		return "game.EventPermanentDied", nil
	case game.EventZoneChanged:
		return "game.EventZoneChanged", nil
	case game.EventCardDiscarded:
		return "game.EventCardDiscarded", nil
	case game.EventCycled:
		return "game.EventCycled", nil
	case game.EventPermanentMutated:
		return "game.EventPermanentMutated", nil
	case game.EventPermanentTapped:
		return "game.EventPermanentTapped", nil
	case game.EventPermanentUntapped:
		return "game.EventPermanentUntapped", nil
	case game.EventPermanentTurnedFaceUp:
		return "game.EventPermanentTurnedFaceUp", nil
	case game.EventPermanentSacrificed:
		return "game.EventPermanentSacrificed", nil
	case game.EventScry:
		return "game.EventScry", nil
	case game.EventSurveil:
		return "game.EventSurveil", nil
	case game.EventAbilityActivated:
		return "game.EventAbilityActivated", nil
	case game.EventObjectBecameTarget:
		return "game.EventObjectBecameTarget", nil
	case game.EventCountersAdded:
		return "game.EventCountersAdded", nil
	case game.EventBeginningOfStep:
		return "game.EventBeginningOfStep", nil
	case game.EventTokenCreated:
		return "game.EventTokenCreated", nil
	default:
		return "", fmt.Errorf("render: unsupported event kind %d", event)
	}
}

func renderDamageRecipient(recipient game.DamageRecipientKind) (string, error) {
	const known = game.DamageRecipientPlayer | game.DamageRecipientPermanent
	if recipient == game.DamageRecipientNone || recipient&^known != 0 {
		return "", fmt.Errorf("render: unsupported damage recipient %d", recipient)
	}
	var values []string
	if recipient&game.DamageRecipientPlayer != 0 {
		values = append(values, "game.DamageRecipientPlayer")
	}
	if recipient&game.DamageRecipientPermanent != 0 {
		values = append(values, "game.DamageRecipientPermanent")
	}
	return strings.Join(values, " | "), nil
}

func renderAttackRecipient(recipient game.AttackRecipientKind) (string, error) {
	const known = game.AttackRecipientPlayer |
		game.AttackRecipientPlaneswalker |
		game.AttackRecipientBattle
	if recipient == game.AttackRecipientAny || recipient&^known != 0 {
		return "", fmt.Errorf("render: unsupported attack recipient %d", recipient)
	}
	var values []string
	if recipient&game.AttackRecipientPlayer != 0 {
		values = append(values, "game.AttackRecipientPlayer")
	}
	if recipient&game.AttackRecipientPlaneswalker != 0 {
		values = append(values, "game.AttackRecipientPlaneswalker")
	}
	if recipient&game.AttackRecipientBattle != 0 {
		values = append(values, "game.AttackRecipientBattle")
	}
	return strings.Join(values, " | "), nil
}

func renderDuration(duration game.EffectDuration) (string, error) {
	switch duration {
	case game.DurationPermanent:
		return "game.DurationPermanent", nil
	case game.DurationUntilEndOfTurn:
		return "game.DurationUntilEndOfTurn", nil
	case game.DurationUntilYourNextTurn:
		return "game.DurationUntilYourNextTurn", nil
	case game.DurationThisTurn:
		return "game.DurationThisTurn", nil
	case game.DurationUntilEndOfYourNextTurn:
		return "game.DurationUntilEndOfYourNextTurn", nil
	case game.DurationForAsLongAsSourceOnBattlefield:
		return "game.DurationForAsLongAsSourceOnBattlefield", nil
	case game.DurationForAsLongAsYouControlSource:
		return "game.DurationForAsLongAsYouControlSource", nil
	case game.DurationForAsLongAsControlledCreatureEnchanted:
		return "game.DurationForAsLongAsControlledCreatureEnchanted", nil
	default:
		return "", fmt.Errorf("render: unsupported effect duration %d", duration)
	}
}

func renderDelayedTriggerTiming(timing game.DelayedTriggerTiming) (string, error) {
	switch timing {
	case game.DelayedAtBeginningOfNextEndStep:
		return "game.DelayedAtBeginningOfNextEndStep", nil
	case game.DelayedAtBeginningOfNextUpkeep:
		return "game.DelayedAtBeginningOfNextUpkeep", nil
	case game.DelayedAtBeginningOfNextMainPhase:
		return "game.DelayedAtBeginningOfNextMainPhase", nil
	default:
		return "", fmt.Errorf("render: unsupported delayed trigger timing %d", timing)
	}
}

func renderResolutionChoiceKind(kind game.ResolutionChoiceKind) (string, error) {
	switch kind {
	case game.ResolutionChoiceMana:
		return "game.ResolutionChoiceMana", nil
	case game.ResolutionChoiceCardType:
		return "game.ResolutionChoiceCardType", nil
	case game.ResolutionChoicePlayer:
		return "game.ResolutionChoicePlayer", nil
	case game.ResolutionChoiceCard:
		return "game.ResolutionChoiceCard", nil
	case game.ResolutionChoiceNumber:
		return "game.ResolutionChoiceNumber", nil
	default:
		return "", fmt.Errorf("render: unsupported resolution choice kind %d", kind)
	}
}

func renderResolutionChoiceColorSource(source game.ResolutionChoiceColorSource) (string, error) {
	switch source {
	case game.ResolutionChoiceColorSourceStatic:
		return "game.ResolutionChoiceColorSourceStatic", nil
	case game.ResolutionChoiceColorSourceCommanderIdentity:
		return "game.ResolutionChoiceColorSourceCommanderIdentity", nil
	case game.ResolutionChoiceColorSourceLandsProduce:
		return "game.ResolutionChoiceColorSourceLandsProduce", nil
	case game.ResolutionChoiceColorSourceLinkedExileColors:
		return "game.ResolutionChoiceColorSourceLinkedExileColors", nil
	case game.ResolutionChoiceColorSourceControlledPermanentColors:
		return "game.ResolutionChoiceColorSourceControlledPermanentColors", nil
	default:
		return "", fmt.Errorf("render: unsupported resolution choice color source %d", source)
	}
}

func renderManaSpendConditionKind(kind game.ManaSpendConditionKind) (string, error) {
	switch kind {
	case game.ManaSpendCastCommanderCreatureType:
		return "game.ManaSpendCastCommanderCreatureType", nil
	case game.ManaSpendCastChosenCreatureType:
		return "game.ManaSpendCastChosenCreatureType", nil
	case game.ManaSpendCastLegendarySpell:
		return "game.ManaSpendCastLegendarySpell", nil
	case game.ManaSpendCastOrActivateChosenCreatureType:
		return "game.ManaSpendCastOrActivateChosenCreatureType", nil
	case game.ManaSpendCastCreatureSpell:
		return "game.ManaSpendCastCreatureSpell", nil
	default:
		return "", fmt.Errorf("render: unsupported mana spend condition kind %d", kind)
	}
}

func renderManaSpendRestrictionKind(kind game.ManaSpendRestrictionKind) (string, error) {
	switch kind {
	case game.ManaSpendUnrestricted:
		return "game.ManaSpendUnrestricted", nil
	case game.ManaSpendRestrictedToCondition:
		return "game.ManaSpendRestrictedToCondition", nil
	default:
		return "", fmt.Errorf("render: unsupported mana spend restriction kind %d", kind)
	}
}

func renderZone(zoneType zone.Type) (string, error) {
	switch zoneType {
	case zone.Battlefield:
		return "zone.Battlefield", nil
	case zone.Hand:
		return "zone.Hand", nil
	case zone.Graveyard:
		return "zone.Graveyard", nil
	case zone.Library:
		return "zone.Library", nil
	case zone.Exile:
		return "zone.Exile", nil
	default:
		return "", fmt.Errorf("render: unsupported zone %d", zoneType)
	}
}

// renderText renders a string field value, preferring a raw backtick literal for
// multi-line text and falling back to a quoted literal when the text already
// contains a backtick.
func renderText(text string) string {
	if strings.ContainsRune(text, '`') {
		return strconv.Quote(text)
	}
	if strings.ContainsRune(text, '\n') {
		return "`" + text + "`"
	}
	return strconv.Quote(text)
}

func structLit(typeName string, fields []string) string {
	if len(fields) == 0 {
		return typeName + "{}"
	}
	return typeName + "{\n" + strings.Join(fields, "\n") + "\n}"
}

func sliceLit(elementType string, elements []string) string {
	if len(elements) == 0 {
		return "[]" + elementType + "{}"
	}
	return "[]" + elementType + "{\n" + strings.Join(elements, "\n") + "\n}"
}

func sliceField(fieldName, elementType string, elements []string) string {
	return fieldName + ": " + sliceLit(elementType, elements) + ","
}

// compactStructLit renders a struct literal on a single line so that gofmt
// preserves it inline. Each field must be a "Key: value" fragment without a
// trailing comma.
func compactStructLit(typeName string, fields []string) string {
	return typeName + "{" + strings.Join(fields, ", ") + "}"
}
