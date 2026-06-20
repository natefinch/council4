package cardgen

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// lowerReminderManaAbility preserves a parenthesized reminder mana ability such
// as "({T}: Add {R} or {G}.)" and consumes other rules-free reminder abilities.
func lowerReminderManaAbility(
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) (abilityLowering, *shared.Diagnostic) {
	unsupported := func() *shared.Diagnostic {
		return executableDiagnostic(
			ability,
			"unsupported reminder ability",
			"the executable source backend does not yet lower reminder abilities",
		)
	}
	innerDocument, innerDiags, ok := syntax.ReminderInner()
	if !ok {
		return abilityLowering{}, unsupported()
	}
	innerComp, compilerDiags := compiler.Compile(innerDocument, compiler.Context{})
	innerDiags = append(append([]shared.Diagnostic(nil), innerDiags...), compilerDiags...)
	if len(innerComp.Abilities) == 1 && isSemanticManaAbility(innerComp.Abilities[0]) {
		if len(innerDiags) != 0 ||
			len(innerComp.Syntax.Abilities) != 1 ||
			innerComp.Abilities[0].Kind != compiler.AbilityActivated {
			return abilityLowering{}, unsupported()
		}
		manaAbility, diagnostic := lowerManaAbility(
			"",
			innerComp.Abilities[0],
			&innerComp.Syntax.Abilities[0],
		)
		if diagnostic != nil {
			return abilityLowering{}, unsupported()
		}
		// The compiled reminder ability has no independent semantic elements;
		// all content is filtered as parenthesized. The consumed counts are all
		// zero, matching the empty CompiledAbility fields.
		return abilityLowering{
			manaAbility: opt.Val(manaAbility),
			consumed:    semanticConsumption{},
			sourceSpans: []shared.Span{ability.Span},
		}, nil
	}

	// Non-mana reminder abilities carry no semantic content beyond their
	// parenthesized explanation.
	return abilityLowering{
		sourceSpans: []shared.Span{ability.Span},
	}, nil
}

// lowerManaAbility lowers an activated mana ability into a game.ManaAbility.
// It accepts the same supported cost shapes as ordinary activated abilities,
// plus supported fixed-symbol, choice, and any-color mana output bodies. A
// single fixed self-damage rider ("<CARDNAME> deals N damage to you") may
// accompany the add-mana effect, modelling painlands, the painland Talismans,
// and similar self-damaging mana sources; the lowered content already carries
// the source-dealt Damage instruction. Unrecognised costs and bodies remain
// fail-closed.
func lowerManaAbility(
	cardName string,
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) (game.ManaAbility, *shared.Diagnostic) {
	if len(ability.Content.Modes) != 0 {
		return game.ManaAbility{}, executableDiagnostic(
			ability,
			"unsupported activation modes",
			"the Payment Planner cannot safely choose modes for a mana ability",
		)
	}
	shell, diagnostic := lowerActivationShell(cardName, ability, syntax)
	if diagnostic != nil {
		return game.ManaAbility{}, diagnostic
	}
	if len(shell.semanticContent.Effects) < 1 || len(shell.semanticContent.Effects) > 2 ||
		shell.semanticContent.Effects[0].Kind != compiler.EffectAddMana ||
		shell.semanticContent.Effects[0].Negated ||
		len(shell.semanticContent.Keywords) != 0 ||
		len(shell.semanticContent.Targets) != 0 ||
		(len(shell.semanticContent.Effects) == 2 &&
			!isSelfDamageToControllerRider(&shell.semanticContent.Effects[1]) &&
			!isManaSpendRider(&shell.semanticContent.Effects[1])) {
		return game.ManaAbility{}, executableDiagnostic(
			ability,
			"unsupported mana effect",
			"the executable source backend supports only exact non-targeting add-mana content, optionally with a fixed self-damage rider, in mana abilities",
		)
	}
	if shell.semanticContent.Effects[0].HasUnrecognizedSibling {
		return game.ManaAbility{}, executableDiagnostic(
			ability,
			"unsupported mana symbol",
			"the executable source backend cannot lower this add-mana content",
		)
	}
	if shell.zoneOfFunction != zone.Battlefield {
		return game.ManaAbility{}, executableDiagnostic(
			ability,
			"unsupported activation zone",
			"the Payment Planner supports mana abilities only on the battlefield",
		)
	}

	return game.ManaAbility{
		Text:                shell.text,
		ManaCost:            shell.manaCost,
		Content:             shell.content,
		Timing:              shell.timing,
		ActivationCondition: shell.activationCondition,
		AdditionalCosts:     shell.additionalCosts,
	}, nil
}

// isSelfDamageToControllerRider reports whether effect is exactly a
// "<CARDNAME> deals N damage to you" rider, the only non-mana effect a mana
// ability may carry. It accepts only a fixed positive amount of source-dealt
// damage to the source's own controller with no target, no divided damage, and
// no additional damage riders, so unrelated deal-damage clauses cannot ride
// into a mana ability. This models the painlands ("This land deals 1 damage to
// you."), the painland Talismans, Ancient Tomb, and Tarnished Citadel, whose
// lowered content already carries the matching self-source Damage instruction.
func isSelfDamageToControllerRider(effect *compiler.CompiledEffect) bool {
	return effect.Kind == compiler.EffectDealDamage &&
		effect.Exact &&
		!effect.Negated &&
		!effect.Optional &&
		!effect.Divided &&
		!effect.HasUnrecognizedSibling &&
		effect.DamageRecipientReference == parser.DamageRecipientReferenceYou &&
		len(effect.DamageRecipientSelectors) == 0 &&
		len(effect.Targets) == 0 &&
		!effect.HasSelfDamageRider &&
		!effect.HasSecondTargetDamageRider &&
		effect.TargetControllerDamageRiderRecipient == parser.DamageRecipientReferenceNone &&
		effect.Duration == compiler.DurationNone &&
		effect.DelayedTiming == 0 &&
		effect.Amount.Known &&
		!effect.Amount.VariableX &&
		effect.Amount.DynamicKind == compiler.DynamicAmountNone &&
		effect.Amount.Value >= 1
}

// isManaSpendRider reports whether effect is one of the closed, fully modeled
// mana-spend rider shapes.
func isManaSpendRider(effect *compiler.CompiledEffect) bool {
	if effect.Kind != compiler.EffectManaSpendRider ||
		!effect.Exact ||
		effect.Negated ||
		effect.Optional ||
		effect.HasUnrecognizedSibling ||
		effect.ManaSpendRider == nil {
		return false
	}
	return isCommanderScryManaSpendRider(effect.ManaSpendRider) ||
		isChosenTypeUncounterableManaSpendRider(effect.ManaSpendRider)
}

func isCommanderScryManaSpendRider(rider *compiler.CompiledManaSpendRider) bool {
	return rider.Condition == parser.ManaSpendCastCommanderCreatureType &&
		rider.Effect == parser.ManaSpendRiderEffectScry &&
		!rider.Restricted &&
		rider.ScryAmount >= 1
}

func isChosenTypeUncounterableManaSpendRider(rider *compiler.CompiledManaSpendRider) bool {
	return rider.Condition == parser.ManaSpendCastChosenCreatureType &&
		rider.Effect == parser.ManaSpendRiderEffectCantBeCountered &&
		rider.Restricted &&
		rider.ScryAmount == 0
}

// lowerManaSpendRiderContent lowers a typed add-mana effect and its exact
// spend-linked rider without consulting source text.
func lowerManaSpendRiderContent(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	manaEffect := ctx.content.Effects[0]
	if ctx.optional ||
		manaEffect.Negated ||
		manaEffect.DelayedTiming != 0 ||
		manaEffect.Duration != compiler.DurationNone ||
		manaEffect.HasUnrecognizedSibling {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported mana effect",
			"the executable source backend supports mana-spend riders only on exact modeled add-mana effects",
		)
	}
	riderEffect := ctx.content.Effects[1].ManaSpendRider
	if isCommanderScryManaSpendRider(riderEffect) {
		if !manaEffect.Mana.CommanderIdentity {
			return game.AbilityContent{}, contentDiagnostic(
				ctx,
				"unsupported mana effect",
				"the commander-creature-type rider requires an exact commander-identity add-mana effect",
			)
		}
		rider := game.ManaSpendRider{
			Condition: game.ManaSpendCastCommanderCreatureType,
			Effect: game.Mode{Sequence: []game.Instruction{
				{
					Primitive: game.Scry{
						Amount: game.Fixed(riderEffect.ScryAmount),
						Player: game.ControllerReference(),
					},
				},
			}},
		}
		return game.TapManaCommanderIdentityWithSpendRiderAbility(ctx.text, rider).Content, nil
	}
	if isChosenTypeUncounterableManaSpendRider(riderEffect) {
		if !manaEffect.Mana.AnyColor {
			return game.AbilityContent{}, contentDiagnostic(
				ctx,
				"unsupported mana effect",
				"the chosen-creature-type rider requires an exact any-color add-mana effect",
			)
		}
		rider := game.ManaSpendRider{
			Condition:         game.ManaSpendCastChosenCreatureType,
			Restriction:       game.ManaSpendRestrictedToCondition,
			SpellRuleEffect:   game.RuleEffectCantBeCountered,
			ChosenSubtypeFrom: game.EntryTypeChoiceKey,
		}
		return game.TapManaChoiceWithSpendRiderAbility(
			ctx.text,
			rider,
			mana.W, mana.U, mana.B, mana.R, mana.G,
		).Content, nil
	}
	return game.AbilityContent{}, contentDiagnostic(
		ctx,
		"unsupported mana effect",
		"the executable source backend cannot lower this mana-spend rider",
	)
}

func lowerAddManaContent(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	if content, ok := lowerTargetOpponentHandMana(ctx); ok {
		return content, nil
	}
	if !effect.Mana.LegacyBodyExact && (effect.Mana.AnyColor || effect.Mana.CommanderIdentity || effect.Mana.LandsProduce || len(effect.Mana.Symbols) != 0) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported mana symbol",
			"the executable source backend cannot lower this add-mana content",
		)
	}
	if ctx.optional ||
		effect.Negated ||
		effect.DelayedTiming != 0 ||
		effect.Duration != compiler.DurationNone ||
		ctx.content.Unconsumed() {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported mana effect",
			"the executable source backend supports only exact unconditional add-mana content",
		)
	}
	content, ok := typedManaEffectContent(effect.Mana)
	if !ok {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported mana symbol",
			"the executable source backend cannot lower this add-mana content",
		)
	}
	return content, nil
}

func lowerTargetOpponentHandMana(ctx contentCtx) (game.AbilityContent, bool) {
	if ctx.optional ||
		len(ctx.content.Effects) != 1 ||
		len(ctx.content.Targets) != 1 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.References) != 0 {
		return game.AbilityContent{}, false
	}
	effect := ctx.content.Effects[0]
	selector := effect.Amount.Selector()
	selection, characteristicsOK := dynamicCountCharacteristics(selector)
	if !effect.Exact ||
		effect.Negated ||
		effect.Optional ||
		effect.DelayedTiming != 0 ||
		effect.Duration != compiler.DurationNone ||
		effect.Context != parser.EffectContextController ||
		effect.Amount.DynamicKind != compiler.DynamicAmountCount ||
		effect.Amount.DynamicForm != compiler.DynamicAmountForEach ||
		effect.Amount.Multiplier != 1 ||
		selector.Kind != compiler.SelectorCard ||
		selector.Controller != compiler.ControllerOpponent ||
		selector.Zone != zone.Hand ||
		!characteristicsOK ||
		!selection.Empty() ||
		!effect.Mana.ColorsKnown ||
		len(effect.Mana.Colors) != 1 ||
		effect.Mana.Colors[0] != mana.R ||
		effect.Mana.Choice ||
		effect.Mana.AnyColor {
		return game.AbilityContent{}, false
	}
	target, ok := playerTargetSpec(ctx.content.Targets[0])
	if !ok || target.Predicate.Player != game.PlayerOpponent {
		return game.AbilityContent{}, false
	}
	player := game.TargetPlayerReference(0)
	return game.Mode{
		Targets: []game.TargetSpec{target},
		Sequence: []game.Instruction{{Primitive: game.AddMana{
			Amount: game.Dynamic(game.DynamicAmount{
				Kind:       game.DynamicAmountCountCardsInZone,
				Multiplier: 1,
				Player:     &player,
				CardZone:   zone.Hand,
				Selection:  &game.Selection{},
			}),
			ManaColor: mana.R,
		}}},
	}.Ability(), true
}

func typedManaEffectContent(effect compiler.CompiledEffectMana) (game.AbilityContent, bool) {
	if effect.FilterPair {
		if len(effect.FilterColors) != 2 {
			return game.AbilityContent{}, false
		}
		return game.TwoColorFilterManaAbility(effect.FilterColors[0], effect.FilterColors[1]).Content, true
	}
	if effect.ChosenColor {
		if effect.ChosenColorFixedKnown {
			return game.TapFixedOrChosenColorManaAbility("", effect.ChosenColorFixed).Content, true
		}
		return game.TapChosenColorManaAbility("").Content, true
	}
	if effect.CommanderIdentity {
		return game.TapManaCommanderIdentityAbility().Content, true
	}
	if effect.LandsProduce {
		relation, ok := landsProduceRelation(effect.LandsProduceScope)
		if !ok {
			return game.AbilityContent{}, false
		}
		return game.TapManaLandsProduceAbility(relation, effect.LandsProduceAnyType).Content, true
	}
	if effect.AnyColor {
		return game.TapManaChoiceAbility(mana.W, mana.U, mana.B, mana.R, mana.G).Content, true
	}
	if !effect.ColorsKnown {
		return game.AbilityContent{}, false
	}
	colors := effect.Colors
	if effect.Choice && len(colors) >= 2 {
		return game.TapManaChoiceAbility(colors...).Content, true
	}
	if !effect.Choice && len(colors) > 0 {
		return manaFixedContent(colors), true
	}
	return game.AbilityContent{}, false
}

// manaFixedContent builds AbilityContent that adds one mana of each color in
// the given order. For a single color this produces a single AddMana
// instruction identical to game.TapManaAbility.
func manaFixedContent(colors []mana.Color) game.AbilityContent {
	seq := make([]game.Instruction, 0, len(colors))
	for _, c := range colors {
		seq = append(seq, game.Instruction{
			Primitive: game.AddMana{
				Amount:    game.Fixed(1),
				ManaColor: c,
			},
		})
	}
	return game.Mode{Sequence: seq}.Ability()
}

// landsProduceRelation maps a parsed lands-produce scope to the game player
// relation its mana ability scopes to. It fails closed on any unrecognized
// scope so an unmodeled wording cannot lower to a mislabeled ability.
func landsProduceRelation(scope parser.ManaLandsProduceScope) (game.PlayerRelation, bool) {
	switch scope {
	case parser.ManaLandsProduceYou:
		return game.PlayerYou, true
	case parser.ManaLandsProduceOpponent:
		return game.PlayerOpponent, true
	default:
		return game.PlayerAny, false
	}
}

func choiceTapManaAbility(colorNames []string) game.ManaAbility {
	colors := make([]mana.Color, 0, len(colorNames))
	for _, name := range colorNames {
		if manaColor, ok := manaColorValue(name); ok {
			colors = append(colors, manaColor)
		}
	}
	return game.TapManaChoiceAbility(colors...)
}

func manaCostHasVariableSymbol(manaCost cost.Mana) bool {
	for _, symbol := range manaCost {
		if symbol.Kind == cost.VariableSymbol {
			return true
		}
	}
	return false
}

func manaColorValue(name string) (mana.Color, bool) {
	switch name {
	case "W":
		return mana.W, true
	case "U":
		return mana.U, true
	case "B":
		return mana.B, true
	case "R":
		return mana.R, true
	case "G":
		return mana.G, true
	case "C":
		return mana.C, true
	default:
		return "", false
	}
}

// parseManaCostValue parses a Scryfall mana cost string (e.g., "{2}{W}") into a
// typed cost.Mana value. Empty input yields a nil cost.
func parseManaCostValue(s string) (cost.Mana, error) {
	if s == "" {
		return nil, nil
	}
	matches := manaSymbolRe.FindAllStringSubmatch(s, -1)
	if len(matches) == 0 {
		return nil, nil
	}
	out := make(cost.Mana, 0, len(matches))
	for _, match := range matches {
		symbol, err := parseManaSymbolValue(match[1])
		if err != nil {
			return nil, fmt.Errorf("unsupported mana symbol {%s} in cost %q: %w", match[1], s, err)
		}
		out = append(out, symbol)
	}
	return out, nil
}

func parseManaSymbolValue(sym string) (cost.Symbol, error) {
	switch sym {
	case "X":
		return cost.X, nil
	case "C":
		return cost.C, nil
	case "S":
		return cost.S, nil
	case "W":
		return cost.W, nil
	case "U":
		return cost.U, nil
	case "B":
		return cost.B, nil
	case "R":
		return cost.R, nil
	case "G":
		return cost.G, nil
	default:
	}
	if before, ok := strings.CutSuffix(sym, "/P"); ok {
		manaColor, ok := manaColorValue(before)
		if !ok {
			return cost.Symbol{}, fmt.Errorf("unsupported mana symbol: %s", sym)
		}
		return cost.PhyrexianMana(manaColor), nil
	}
	if strings.Contains(sym, "/") {
		parts := strings.SplitN(sym, "/", 2)
		if _, err := strconv.Atoi(parts[0]); err == nil {
			manaColor, ok := manaColorValue(parts[1])
			if !ok {
				return cost.Symbol{}, fmt.Errorf("unsupported mana symbol: %s", sym)
			}
			return cost.Twobrid(manaColor), nil
		}
		first, ok := manaColorValue(parts[0])
		second, ok2 := manaColorValue(parts[1])
		if !ok || !ok2 {
			return cost.Symbol{}, fmt.Errorf("unsupported mana symbol: %s", sym)
		}
		return cost.HybridMana(first, second), nil
	}
	if n, err := strconv.Atoi(sym); err == nil {
		return cost.O(n), nil
	}
	return cost.Symbol{}, fmt.Errorf("unsupported mana symbol: %s", sym)
}
