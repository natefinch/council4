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
	return enumLiteral(manaColorLiterals, "mana color", c)
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
	return enumLiteralNonZero(costAdditionalKindLiterals, "additional cost kind", kind)
}

func renderAdditionalDynamicAmount(kind cost.AdditionalDynamicAmount) (string, error) {
	return enumLiteralNonZero(costAdditionalDynamicAmountLiterals, "additional dynamic amount", kind)
}

func renderPowerContributionKind(kind cost.PowerContributionKind) (string, error) {
	return enumLiteral(costPowerContributionKindLiterals, "power contribution kind", kind)
}

func renderCounterKind(kind counter.Kind) (string, error) {
	return enumLiteral(counterKindLiterals, "counter kind", kind)
}

// renderTargetAllow renders the target categories a target slot allows. Unlike
// the recipient bitmasks the zero value is meaningful: it names the unspecified
// constant rather than failing.
func renderTargetAllow(allow game.TargetAllow) (string, error) {
	if allow == game.TargetAllowUnspecified {
		return "game.TargetAllowUnspecified", nil
	}
	return bitmaskLiteral(gameTargetAllowFlags, "target allow", allow)
}

func renderPlayerRelation(relation game.PlayerRelation) (string, error) {
	return enumLiteral(gamePlayerRelationLiterals, "player relation", relation)
}

func renderOwnerRelation(relation game.OwnerRelation) (string, error) {
	return enumLiteral(gameOwnerRelationLiterals, "owner relation", relation)
}

func renderTriggerType(triggerType game.TriggerType) (string, error) {
	return enumLiteral(gameTriggerTypeLiterals, "trigger type", triggerType)
}

func renderStep(step game.Step) (string, error) {
	return enumLiteralNonZero(gameStepLiterals, "step", step)
}

func renderTriggerSource(source game.TriggerSourceFilter) (string, error) {
	return enumLiteralNonZero(gameTriggerSourceFilterLiterals, "trigger source filter", source)
}

func renderTriggerSubject(subject game.TriggerSubjectObject) (string, error) {
	return enumLiteralNonZero(gameTriggerSubjectObjectLiterals, "trigger subject", subject)
}

func renderTriggerController(controller game.TriggerControllerFilter) (string, error) {
	return enumLiteralNonZero(gameTriggerControllerFilterLiterals, "trigger controller filter", controller)
}

func renderTriggerPlayer(player game.TriggerPlayerFilter) (string, error) {
	return enumLiteralNonZero(gameTriggerPlayerFilterLiterals, "trigger player filter", player)
}

// renderEventKind returns the Go expression naming an event kind, refusing the
// kinds recorded in unrenderedEventKinds. The zero value is the unset sentinel
// and never renders.
func renderEventKind(event game.EventKind) (string, error) {
	if unrenderedEventKinds[event] {
		return "", fmt.Errorf("render: event kind %d is produced only by runtime machinery and is not emitted into generated source", int(event))
	}
	return enumLiteralNonZero(gameEventKindLiterals, "event kind", event)
}

// renderDamageRecipient renders the recipient flags a damage trigger matches.
// The zero value means no recipient was specified, which is never a value the
// trigger pattern should carry.
func renderDamageRecipient(recipient game.DamageRecipientKind) (string, error) {
	if recipient == game.DamageRecipientNone {
		return "", fmt.Errorf("render: damage recipient %d specifies no recipient", int(recipient))
	}
	return bitmaskLiteral(gameDamageRecipientKindFlags, "damage recipient", recipient)
}

// renderAttackRecipient renders the recipient flags an attack trigger matches.
// The zero value means the trigger matches any recipient and renders no field.
func renderAttackRecipient(recipient game.AttackRecipientKind) (string, error) {
	if recipient == game.AttackRecipientAny {
		return "", fmt.Errorf("render: attack recipient %d matches any recipient and renders no field", int(recipient))
	}
	return bitmaskLiteral(gameAttackRecipientKindFlags, "attack recipient", recipient)
}

func renderDuration(duration game.EffectDuration) (string, error) {
	return enumLiteral(gameEffectDurationLiterals, "effect duration", duration)
}

func renderResolutionChoiceKind(kind game.ResolutionChoiceKind) (string, error) {
	return enumLiteralNonZero(gameResolutionChoiceKindLiterals, "resolution choice kind", kind)
}

func renderZone(zoneType zone.Type) (string, error) {
	return enumLiteralNonZero(zoneTypeLiterals, "zone", zoneType)
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
