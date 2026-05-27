package rules

import (
	"strings"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/opt"
)

var paymentColors = []mana.Color{
	mana.White,
	mana.Blue,
	mana.Black,
	mana.Red,
	mana.Green,
	mana.Colorless,
}

const flashbackAlternativeLabel = "Flashback"

func manaCostPtr(cost opt.V[mana.Cost]) *mana.Cost {
	if !cost.Exists {
		return nil
	}
	return &cost.Val
}

type spellCostOption struct {
	index           int
	label           string
	card            *game.CardDef
	manaCost        *mana.Cost
	additionalCosts []game.AdditionalCost
}

func canPayCost(g *game.Game, playerID game.PlayerID, cost *mana.Cost) bool {
	return canPayCostWithX(g, playerID, cost, 0)
}

func canPayCostWithX(g *game.Game, playerID game.PlayerID, cost *mana.Cost, xValue int) bool {
	_, ok := buildPaymentPlan(g, playerID, cost, xValue, nil)
	return ok
}

func payCost(g *game.Game, playerID game.PlayerID, cost *mana.Cost) bool {
	return payCostWithX(g, playerID, cost, 0)
}

func payCostWithX(g *game.Game, playerID game.PlayerID, cost *mana.Cost, xValue int) bool {
	plan, ok := buildPaymentPlan(g, playerID, cost, xValue, nil)
	if !ok {
		return false
	}
	return applyPaymentPlan(g, playerID, plan)
}

func canPaySpellCosts(g *game.Game, playerID game.PlayerID, card *game.CardDef, xValue int) bool {
	return canPaySpellCostsWithKickerFromZone(g, playerID, 0, game.ZoneHand, card, xValue, false)
}

func canPaySpellCostsWithKicker(g *game.Game, playerID game.PlayerID, card *game.CardDef, xValue int, kickerPaid bool) bool {
	return canPaySpellCostsWithKickerFromZone(g, playerID, 0, game.ZoneHand, card, xValue, kickerPaid)
}

func canPaySpellCostsWithKickerFromZone(g *game.Game, playerID game.PlayerID, cardID id.ID, sourceZone game.ZoneType, card *game.CardDef, xValue int, kickerPaid bool) bool {
	for _, option := range spellCostOptionsForZoneAndKicker(card, sourceZone, kickerPaid) {
		if _, ok := buildSpellCostPlanForOption(g, playerID, cardID, sourceZone, option, xValue, nil); ok {
			return true
		}
	}
	return false
}

func paySpellCosts(g *game.Game, playerID game.PlayerID, card *game.CardDef, xValue int) ([]string, bool) {
	return paySpellCostsWithKickerAndPreferences(g, playerID, card, xValue, false, nil)
}

func paySpellCostsWithPreferences(g *game.Game, playerID game.PlayerID, card *game.CardDef, xValue int, prefs *paymentPreferences) ([]string, bool) {
	return paySpellCostsWithKickerAndPreferences(g, playerID, card, xValue, false, prefs)
}

func paySpellCostsWithKickerAndPreferences(g *game.Game, playerID game.PlayerID, card *game.CardDef, xValue int, kickerPaid bool, prefs *paymentPreferences) ([]string, bool) {
	return paySpellCostsWithKickerFromZoneAndPreferences(g, playerID, 0, game.ZoneHand, card, xValue, kickerPaid, prefs)
}

func paySpellCostsWithKickerFromZoneAndPreferences(g *game.Game, playerID game.PlayerID, cardID id.ID, sourceZone game.ZoneType, card *game.CardDef, xValue int, kickerPaid bool, prefs *paymentPreferences) ([]string, bool) {
	plan, ok := buildSpellCostPlanWithKickerFromZoneAndPreferences(g, playerID, cardID, sourceZone, card, xValue, kickerPaid, prefs)
	if !ok {
		return nil, false
	}
	player, ok := playerForCostPayment(g, playerID)
	if !ok || !additionalCostPlanStillValid(g, player, plan.additional) || !paymentPlanStillValid(g, player, plan.mana) {
		return nil, false
	}
	if !applyPaymentPlan(g, playerID, plan.mana) {
		return nil, false
	}
	if !applyAdditionalCostPlan(g, plan.additional) {
		panic("spell cost plan became invalid while paying additional costs")
	}
	return plan.additional.paid, true
}

type paymentPlan struct {
	poolSpend      map[mana.Unit]int
	manaTaps       []manaTap
	lifePayment    int
	symbolPayments []game.SymbolPayment
}

type spellCostPlan struct {
	mana       paymentPlan
	additional additionalCostPlan
	option     spellCostOption
}

type abilityCostPlan struct {
	mana       paymentPlan
	additional additionalCostPlan
	tapSource  bool
}

type additionalCostPlan struct {
	player     game.PlayerID
	paid       []string
	sacrifices []*game.Permanent
	discards   []id.ID
	lifePaid   int
}

type manaTap struct {
	permanent *game.Permanent
	color     mana.Color
	amount    int
	snow      bool
}

type manaSource struct {
	permanent *game.Permanent
	color     mana.Color
	amount    int
	snow      bool
}

func buildSpellCostPlan(g *game.Game, playerID game.PlayerID, card *game.CardDef, xValue int) (spellCostPlan, bool) {
	return buildSpellCostPlanWithPreferences(g, playerID, card, xValue, nil)
}

func buildSpellCostPlanWithPreferences(g *game.Game, playerID game.PlayerID, card *game.CardDef, xValue int, prefs *paymentPreferences) (spellCostPlan, bool) {
	return buildSpellCostPlanWithKickerAndPreferences(g, playerID, card, xValue, false, prefs)
}

func buildSpellCostPlanWithKickerAndPreferences(g *game.Game, playerID game.PlayerID, card *game.CardDef, xValue int, kickerPaid bool, prefs *paymentPreferences) (spellCostPlan, bool) {
	return buildSpellCostPlanWithKickerFromZoneAndPreferences(g, playerID, 0, game.ZoneHand, card, xValue, kickerPaid, prefs)
}

func buildSpellCostPlanWithKickerFromZoneAndPreferences(g *game.Game, playerID game.PlayerID, cardID id.ID, sourceZone game.ZoneType, card *game.CardDef, xValue int, kickerPaid bool, prefs *paymentPreferences) (spellCostPlan, bool) {
	options := spellCostOptionsForZoneAndKicker(card, sourceZone, kickerPaid)
	if len(options) == 0 {
		return spellCostPlan{}, false
	}
	if prefs != nil {
		for _, option := range options {
			if option.index == prefs.alternativeIndex {
				return buildSpellCostPlanForOption(g, playerID, cardID, sourceZone, option, xValue, prefs)
			}
		}
		return spellCostPlan{}, false
	}
	for _, option := range options {
		if plan, ok := buildSpellCostPlanForOption(g, playerID, cardID, sourceZone, option, xValue, nil); ok {
			return plan, true
		}
	}
	return spellCostPlan{}, false
}

func buildSpellCostPlanForOption(g *game.Game, playerID game.PlayerID, cardID id.ID, sourceZone game.ZoneType, option spellCostOption, xValue int, prefs *paymentPreferences) (spellCostPlan, bool) {
	option = applyCostModifiers(g, costModificationContext{player: playerID, card: option.card, cardID: cardID, sourceZone: sourceZone, option: option})
	plan := spellCostPlan{option: option}
	additional, ok := buildAdditionalCostPlanForCosts(g, playerID, option.additionalCosts, prefs)
	if !ok {
		return plan, false
	}
	excluded := make(map[id.ID]bool)
	for _, sacrifice := range additional.sacrifices {
		excluded[sacrifice.ObjectID] = true
	}
	manaPlan, ok := buildPaymentPlanWithPreferences(g, playerID, option.manaCost, xValue, excluded, prefs)
	if !ok {
		return plan, false
	}
	plan.additional = additional
	plan.mana = manaPlan
	return plan, true
}

func applyCostModifiers(g *game.Game, context costModificationContext) spellCostOption {
	context.option.manaCost = applyGenericCostModifiers(context.option.manaCost, costModifiersForContext(g, context))
	return context.option
}

func costModifiersForContext(g *game.Game, context costModificationContext) []game.CostModifier {
	var modifiers []game.CostModifier
	for _, modifier := range g.CostModifiers {
		if modifier.Kind != game.CostModifierSpell {
			continue
		}
		if modifier.MatchCardType && (context.card == nil || !context.card.HasType(modifier.CardType)) {
			continue
		}
		modifiers = append(modifiers, modifier)
	}
	if context.sourceZone == game.ZoneCommand && context.cardID != 0 {
		player, ok := playerByID(g, context.player)
		if ok && player.CommanderInstanceID == context.cardID && player.CommanderTax() > 0 {
			modifiers = append(modifiers, game.CostModifier{
				Kind:            game.CostModifierSpell,
				GenericIncrease: player.CommanderTax(),
			})
		}
	}
	modifiers = append(modifiers, staticCostModifiersForContext(g, context)...)
	return modifiers
}

func applyGenericCostModifiers(cost *mana.Cost, modifiers []game.CostModifier) *mana.Cost {
	if len(modifiers) == 0 {
		return cost
	}
	generic := genericCostAmount(cost)
	minimum := 0
	set := (*int)(nil)
	for _, modifier := range modifiers {
		if modifier.SetGeneric.Exists {
			set = &modifier.SetGeneric.Val
		}
		generic += modifier.GenericIncrease
		generic -= modifier.GenericReduction
		if modifier.MinimumGeneric > minimum {
			minimum = modifier.MinimumGeneric
		}
	}
	if set != nil {
		generic = *set
	}
	if generic < minimum {
		generic = minimum
	}
	if generic < 0 {
		generic = 0
	}
	return costWithGenericAmount(cost, generic)
}

func genericCostAmount(cost *mana.Cost) int {
	if cost == nil {
		return 0
	}
	total := 0
	for _, symbol := range *cost {
		if symbol.Kind == mana.GenericSymbol {
			total += symbol.Generic
		}
	}
	return total
}

func costWithGenericAmount(cost *mana.Cost, generic int) *mana.Cost {
	var modified mana.Cost
	if generic > 0 {
		modified = append(modified, mana.GenericMana(generic))
	}
	if cost != nil {
		for _, symbol := range *cost {
			if symbol.Kind != mana.GenericSymbol {
				modified = append(modified, symbol)
			}
		}
	}
	return &modified
}

func buildAbilityCostPlan(g *game.Game, playerID game.PlayerID, source *game.Permanent, ability *game.AbilityDef, xValue int) (abilityCostPlan, bool) {
	return buildAbilityCostPlanWithPreferences(g, playerID, source, ability, xValue, nil)
}

func buildAbilityCostPlanWithPreferences(g *game.Game, playerID game.PlayerID, source *game.Permanent, ability *game.AbilityDef, xValue int, prefs *paymentPreferences) (abilityCostPlan, bool) {
	plan := abilityCostPlan{}
	if source == nil || ability == nil {
		return plan, false
	}
	if xValue != 0 && !costHasVariableMana(manaCostPtr(ability.ManaCost)) {
		return plan, false
	}
	tapSource := hasTapCost(ability)
	if tapSource && !canTapPermanentForAbility(g, source) {
		return plan, false
	}
	additional, ok := buildAdditionalCostPlanForCosts(g, playerID, abilityAdditionalCosts(ability), prefs)
	if !ok {
		return plan, false
	}
	excluded := make(map[id.ID]bool)
	if tapSource {
		excluded[source.ObjectID] = true
	}
	for _, sacrifice := range additional.sacrifices {
		excluded[sacrifice.ObjectID] = true
	}
	manaPlan, ok := buildPaymentPlanWithPreferences(g, playerID, manaCostPtr(ability.ManaCost), xValue, excluded, prefs)
	if !ok {
		return plan, false
	}
	plan.mana = manaPlan
	plan.additional = additional
	plan.tapSource = tapSource
	return plan, true
}

func buildPaymentPlan(g *game.Game, playerID game.PlayerID, cost *mana.Cost, xValue int, exclude map[id.ID]bool) (paymentPlan, bool) {
	return buildPaymentPlanWithPreferences(g, playerID, cost, xValue, exclude, nil)
}

func buildPaymentPlanWithPreferences(g *game.Game, playerID game.PlayerID, cost *mana.Cost, xValue int, exclude map[id.ID]bool, prefs *paymentPreferences) (paymentPlan, bool) {
	plan := paymentPlan{poolSpend: make(map[mana.Unit]int)}
	player, ok := playerForCostPayment(g, playerID)
	if !ok {
		return plan, false
	}
	pool := snapshotPool(player)
	manaSources := availableManaSources(g, playerID, exclude)
	if xValue < 0 {
		return plan, false
	}
	if cost == nil {
		return plan, true
	}

	for _, symbol := range *cost {
		switch symbol.Kind {
		case mana.ColoredSymbol:
			if !payColoredSymbol(&plan, pool, manaSources, symbol, symbol.Color, game.SymbolPaymentMana) {
				return plan, false
			}
		case mana.ColorlessSymbol:
			if !payColoredSymbol(&plan, pool, manaSources, symbol, mana.Colorless, game.SymbolPaymentMana) {
				return plan, false
			}
		}
	}
	for _, symbol := range *cost {
		if symbol.Kind == mana.SnowSymbol {
			if !paySnowSymbol(&plan, pool, manaSources, symbol) {
				return plan, false
			}
		}
	}
	for _, symbol := range *cost {
		switch symbol.Kind {
		case mana.HybridSymbol:
			if !payHybridSymbol(&plan, pool, manaSources, symbol) {
				return plan, false
			}
		case mana.MonoHybridSymbol:
			if !payMonoHybridSymbol(&plan, pool, manaSources, symbol) {
				return plan, false
			}
		case mana.PhyrexianSymbol:
			if !payPhyrexianSymbol(player, &plan, pool, manaSources, symbol, prefs) {
				return plan, false
			}
		}
	}
	for _, symbol := range *cost {
		switch symbol.Kind {
		case mana.GenericSymbol:
			if !payGenericSymbol(&plan, pool, manaSources, symbol, symbol.Generic, game.SymbolPaymentGeneric) {
				return plan, false
			}
		case mana.VariableSymbol:
			if !payGenericSymbol(&plan, pool, manaSources, symbol, xValue, game.SymbolPaymentX) {
				return plan, false
			}
		default:
			if symbol.Kind != mana.ColoredSymbol &&
				symbol.Kind != mana.ColorlessSymbol &&
				symbol.Kind != mana.SnowSymbol &&
				symbol.Kind != mana.HybridSymbol &&
				symbol.Kind != mana.MonoHybridSymbol &&
				symbol.Kind != mana.PhyrexianSymbol {
				return plan, false
			}
		}
	}
	return plan, true
}

func buildAdditionalCostPlan(g *game.Game, playerID game.PlayerID, card *game.CardDef) (additionalCostPlan, bool) {
	return buildAdditionalCostPlanForCosts(g, playerID, spellAdditionalCosts(card), nil)
}

func buildAdditionalCostPlanForCosts(g *game.Game, playerID game.PlayerID, costs []game.AdditionalCost, prefs *paymentPreferences) (additionalCostPlan, bool) {
	plan := additionalCostPlan{player: playerID}
	for _, cost := range costs {
		amount := additionalCostAmount(cost)
		switch cost.Kind {
		case game.AdditionalCostUnknown:
			if cost.Text == "" {
				continue
			}
			return plan, false
		case game.AdditionalCostTap:
			continue
		case game.AdditionalCostSacrifice:
			chosen := preferredSacrificePermanents(g, playerID, cost, amount, plan.sacrifices, prefs)
			if len(chosen) != amount {
				return plan, false
			}
			plan.sacrifices = append(plan.sacrifices, chosen...)
			plan.paid = append(plan.paid, additionalCostText(cost))
		case game.AdditionalCostDiscard:
			chosen := preferredDiscardCards(g, playerID, cost, amount, plan.discards, prefs)
			if len(chosen) != amount {
				return plan, false
			}
			plan.discards = append(plan.discards, chosen...)
			plan.paid = append(plan.paid, additionalCostText(cost))
		case game.AdditionalCostPayLife:
			player, ok := playerByID(g, playerID)
			if !ok || player.Life < amount {
				return plan, false
			}
			plan.lifePaid += amount
			plan.paid = append(plan.paid, additionalCostText(cost))
		default:
			return plan, false
		}
	}
	return plan, true
}

func spellAdditionalCosts(card *game.CardDef) []game.AdditionalCost {
	if card == nil {
		return nil
	}
	for _, ability := range card.Abilities {
		if ability.Kind == game.SpellAbility {
			return abilityAdditionalCosts(&ability)
		}
	}
	return nil
}

func spellCostOptions(card *game.CardDef) []spellCostOption {
	return spellCostOptionsForKicker(card, false)
}

func spellCostOptionsForKicker(card *game.CardDef, kickerPaid bool) []spellCostOption {
	return spellCostOptionsForZoneAndKicker(card, game.ZoneHand, kickerPaid)
}

func spellCostOptionsForZoneAndKicker(card *game.CardDef, sourceZone game.ZoneType, kickerPaid bool) []spellCostOption {
	if card == nil {
		return nil
	}
	ability, ok := firstSpellAbility(card)
	if !ok {
		return []spellCostOption{{index: 0, label: "Normal cost", card: card, manaCost: manaCostPtr(card.ManaCost)}}
	}
	requiredAdditional := abilityAdditionalCosts(ability)
	options := []spellCostOption{
		{
			index:           0,
			label:           "Normal cost",
			card:            card,
			manaCost:        spellManaCostWithKicker(manaCostPtr(card.ManaCost), ability, kickerPaid),
			additionalCosts: append([]game.AdditionalCost(nil), requiredAdditional...),
		},
	}
	for i, alternative := range ability.AlternativeCosts {
		if isFlashbackAlternative(alternative) && sourceZone != game.ZoneGraveyard {
			continue
		}
		if sourceZone == game.ZoneGraveyard && !isFlashbackAlternative(alternative) {
			continue
		}
		additional := append([]game.AdditionalCost(nil), requiredAdditional...)
		additional = append(additional, alternative.AdditionalCosts...)
		label := alternative.Label
		if label == "" {
			label = "Alternative cost"
		}
		options = append(options, spellCostOption{
			index:           i + 1,
			label:           label,
			card:            card,
			manaCost:        spellManaCostWithKicker(manaCostPtr(alternative.ManaCost), ability, kickerPaid),
			additionalCosts: additional,
		})
	}
	if sourceZone == game.ZoneGraveyard {
		return options[1:]
	}
	return options
}

func isFlashbackAlternative(alternative game.AlternativeCost) bool {
	return strings.EqualFold(strings.TrimSpace(alternative.Label), flashbackAlternativeLabel)
}

func spellManaCostWithKicker(base *mana.Cost, ability *game.AbilityDef, kickerPaid bool) *mana.Cost {
	if !kickerPaid || ability == nil || !ability.KickerCost.Exists {
		return base
	}
	combined := mana.Cost{}
	if base != nil {
		combined = append(combined, (*base)...)
	}
	combined = append(combined, ability.KickerCost.Val...)
	return &combined
}

func abilityAdditionalCosts(ability *game.AbilityDef) []game.AdditionalCost {
	if ability == nil {
		return nil
	}
	return append([]game.AdditionalCost(nil), ability.AdditionalCosts...)
}

func sacrificeCostMatcher(cost string) (func(*game.CardDef) bool, bool) {
	typed, ok := sacrificeAdditionalCost(cost)
	if !ok {
		return nil, false
	}
	return additionalCostCardMatcher(typed), true
}

func sacrificeAdditionalCost(cost string) (game.AdditionalCost, bool) {
	normalized := strings.TrimSuffix(strings.ToLower(strings.TrimSpace(cost)), ".")
	switch normalized {
	case "sacrifice a creature":
		return game.AdditionalCost{Kind: game.AdditionalCostSacrifice, Text: cost, Amount: 1, MatchPermanentType: true, PermanentType: game.TypeCreature}, true
	case "sacrifice an artifact":
		return game.AdditionalCost{Kind: game.AdditionalCostSacrifice, Text: cost, Amount: 1, MatchPermanentType: true, PermanentType: game.TypeArtifact}, true
	case "sacrifice an enchantment":
		return game.AdditionalCost{Kind: game.AdditionalCostSacrifice, Text: cost, Amount: 1, MatchPermanentType: true, PermanentType: game.TypeEnchantment}, true
	case "sacrifice a land":
		return game.AdditionalCost{Kind: game.AdditionalCostSacrifice, Text: cost, Amount: 1, MatchPermanentType: true, PermanentType: game.TypeLand}, true
	case "sacrifice a permanent":
		return game.AdditionalCost{Kind: game.AdditionalCostSacrifice, Text: cost, Amount: 1}, true
	default:
		return game.AdditionalCost{}, false
	}
}

func chooseSacrificePermanent(g *game.Game, playerID game.PlayerID, matches func(*game.CardDef) bool) (*game.Permanent, bool) {
	if matches == nil {
		return nil, false
	}
	for _, permanent := range g.Battlefield {
		if effectiveController(g, permanent) != playerID {
			continue
		}
		def, ok := permanentCardDef(g, permanent)
		if ok && matches(def) {
			return permanent, true
		}
	}
	return nil, false
}

func chooseSacrificePermanents(g *game.Game, playerID game.PlayerID, cost game.AdditionalCost, amount int, alreadyChosen []*game.Permanent) []*game.Permanent {
	chosenIDs := make(map[id.ID]bool)
	for _, permanent := range alreadyChosen {
		chosenIDs[permanent.ObjectID] = true
	}
	var chosen []*game.Permanent
	for _, permanent := range g.Battlefield {
		if effectiveController(g, permanent) != playerID || chosenIDs[permanent.ObjectID] {
			continue
		}
		if additionalCostMatchesPermanent(g, permanent, cost) {
			chosen = append(chosen, permanent)
			if len(chosen) == amount {
				return chosen
			}
		}
	}
	return chosen
}

func preferredSacrificePermanents(g *game.Game, playerID game.PlayerID, cost game.AdditionalCost, amount int, alreadyChosen []*game.Permanent, prefs *paymentPreferences) []*game.Permanent {
	if prefs == nil || len(prefs.sacrificeChoices) == 0 {
		return chooseSacrificePermanents(g, playerID, cost, amount, alreadyChosen)
	}
	chosenIDs := make(map[id.ID]bool)
	for _, permanent := range alreadyChosen {
		chosenIDs[permanent.ObjectID] = true
	}
	var chosen []*game.Permanent
	var consumed int
	for _, permanentID := range prefs.sacrificeChoices {
		permanent, ok := permanentByObjectID(g, permanentID)
		if !ok || effectiveController(g, permanent) != playerID || chosenIDs[permanentID] || !additionalCostMatchesPermanent(g, permanent, cost) {
			return nil
		}
		chosen = append(chosen, permanent)
		chosenIDs[permanentID] = true
		consumed++
		if len(chosen) == amount {
			prefs.sacrificeChoices = prefs.sacrificeChoices[consumed:]
			return chosen
		}
	}
	return nil
}

func chooseDiscardCards(g *game.Game, playerID game.PlayerID, cost game.AdditionalCost, amount int, alreadyChosen []id.ID) []id.ID {
	player, ok := playerByID(g, playerID)
	if !ok {
		return nil
	}
	chosenIDs := make(map[id.ID]bool)
	for _, cardID := range alreadyChosen {
		chosenIDs[cardID] = true
	}
	var chosen []id.ID
	for _, cardID := range player.Hand.All() {
		if chosenIDs[cardID] {
			continue
		}
		card, ok := g.GetCardInstance(cardID)
		if !ok || !additionalCostMatchesCard(cardFaceOrDefault(card, game.FaceFront), cost) {
			continue
		}
		chosen = append(chosen, cardID)
		if len(chosen) == amount {
			return chosen
		}
	}
	return chosen
}

func preferredDiscardCards(g *game.Game, playerID game.PlayerID, cost game.AdditionalCost, amount int, alreadyChosen []id.ID, prefs *paymentPreferences) []id.ID {
	if prefs == nil || len(prefs.discardChoices) == 0 {
		return chooseDiscardCards(g, playerID, cost, amount, alreadyChosen)
	}
	player, ok := playerByID(g, playerID)
	if !ok {
		return nil
	}
	chosenIDs := make(map[id.ID]bool)
	for _, cardID := range alreadyChosen {
		chosenIDs[cardID] = true
	}
	var chosen []id.ID
	var consumed int
	for _, cardID := range prefs.discardChoices {
		card, ok := g.GetCardInstance(cardID)
		if !ok || !player.Hand.Contains(cardID) || chosenIDs[cardID] || !additionalCostMatchesCard(cardFaceOrDefault(card, game.FaceFront), cost) {
			return nil
		}
		chosen = append(chosen, cardID)
		chosenIDs[cardID] = true
		consumed++
		if len(chosen) == amount {
			prefs.discardChoices = prefs.discardChoices[consumed:]
			return chosen
		}
	}
	return nil
}

func additionalCostMatchesPermanent(g *game.Game, permanent *game.Permanent, cost game.AdditionalCost) bool {
	if cost.MatchPermanentType && !permanentHasType(g, permanent, cost.PermanentType) {
		return false
	}
	return true
}

func additionalCostMatchesCard(card *game.CardDef, cost game.AdditionalCost) bool {
	if card == nil {
		return false
	}
	if cost.MatchCardType && !card.HasType(cost.CardType) {
		return false
	}
	return true
}

func additionalCostCardMatcher(cost game.AdditionalCost) func(*game.CardDef) bool {
	return func(card *game.CardDef) bool {
		return additionalCostMatchesCard(card, game.AdditionalCost{
			MatchCardType: cost.MatchPermanentType,
			CardType:      cost.PermanentType,
		})
	}
}

func additionalCostAmount(cost game.AdditionalCost) int {
	if cost.Amount > 0 {
		return cost.Amount
	}
	return 1
}

func additionalCostText(cost game.AdditionalCost) string {
	if cost.Text != "" {
		return cost.Text
	}
	switch cost.Kind {
	case game.AdditionalCostSacrifice:
		return "Sacrifice a permanent"
	case game.AdditionalCostDiscard:
		return "Discard a card"
	case game.AdditionalCostPayLife:
		return "Pay life"
	case game.AdditionalCostExile:
		return "Exile a card"
	case game.AdditionalCostReveal:
		return "Reveal a card"
	case game.AdditionalCostTap:
		return "{T}"
	default:
		return "Additional cost"
	}
}

func additionalCostPlanStillValid(g *game.Game, player *game.Player, plan additionalCostPlan) bool {
	for _, sacrifice := range plan.sacrifices {
		permanent, ok := permanentByObjectID(g, sacrifice.ObjectID)
		if !ok || effectiveController(g, permanent) != player.ID || permanent != sacrifice {
			return false
		}
	}
	for _, cardID := range plan.discards {
		if !player.Hand.Contains(cardID) {
			return false
		}
	}
	if plan.lifePaid > 0 && player.Life < plan.lifePaid {
		return false
	}
	return true
}

func applyAdditionalCostPlan(g *game.Game, plan additionalCostPlan) bool {
	for _, sacrifice := range plan.sacrifices {
		if !movePermanentToZone(g, sacrifice, game.ZoneGraveyard) {
			return false
		}
	}
	for _, cardID := range plan.discards {
		card, ok := g.GetCardInstance(cardID)
		if !ok || !discardCardFromHand(g, card.Owner, cardID) {
			return false
		}
	}
	if plan.lifePaid > 0 {
		player, ok := playerByID(g, plan.player)
		if !ok || player.Life < plan.lifePaid {
			return false
		}
		loseLife(g, plan.player, plan.lifePaid)
	}
	return true
}

func payAbilityCosts(g *game.Game, playerID game.PlayerID, source *game.Permanent, ability *game.AbilityDef, xValue int) (abilityCostPlan, bool) {
	return payAbilityCostsWithPreferences(g, playerID, source, ability, xValue, nil)
}

func payAbilityCostsWithPreferences(g *game.Game, playerID game.PlayerID, source *game.Permanent, ability *game.AbilityDef, xValue int, prefs *paymentPreferences) (abilityCostPlan, bool) {
	plan, ok := buildAbilityCostPlanWithPreferences(g, playerID, source, ability, xValue, prefs)
	if !ok {
		return plan, false
	}
	player, ok := playerForCostPayment(g, playerID)
	if !ok || !abilityCostPlanStillValid(g, player, source, plan) {
		return plan, false
	}
	if !applyPaymentPlan(g, playerID, plan.mana) {
		return plan, false
	}
	if plan.tapSource {
		if !tapPermanentForAbility(g, source) {
			return plan, false
		}
	}
	if !applyAdditionalCostPlan(g, plan.additional) {
		panic("ability cost plan became invalid while paying additional costs")
	}
	return plan, true
}

func abilityCostPlanStillValid(g *game.Game, player *game.Player, source *game.Permanent, plan abilityCostPlan) bool {
	if source == nil {
		return false
	}
	if plan.tapSource && !canTapPermanentForAbility(g, source) {
		return false
	}
	return additionalCostPlanStillValid(g, player, plan.additional) &&
		paymentPlanStillValid(g, player, plan.mana)
}

func applyPaymentPlan(g *game.Game, playerID game.PlayerID, plan paymentPlan) bool {
	player, ok := playerForCostPayment(g, playerID)
	if !ok || !paymentPlanStillValid(g, player, plan) {
		return false
	}
	for _, tap := range plan.manaTaps {
		if !tapPermanentForMana(g, tap.permanent, tap.color, tap.amount, tap.snow) {
			panic("payment plan became invalid while tapping mana sources")
		}
	}
	for _, color := range paymentColors {
		for _, snow := range []bool{false, true} {
			unit := mana.Unit{Color: color, Snow: snow}
			amount := plan.poolSpend[unit]
			if amount > 0 && !player.ManaPool.SpendMatching(amount, func(candidate mana.Unit) bool { return candidate == unit }) {
				panic("payment plan became invalid while spending mana")
			}
		}
	}
	if plan.lifePayment > 0 {
		if player.Life < plan.lifePayment {
			return false
		}
		loseLife(g, playerID, plan.lifePayment)
	}
	return true
}

func paymentPlanStillValid(g *game.Game, player *game.Player, plan paymentPlan) bool {
	tappedMana := make(map[mana.Unit]int)
	for _, tap := range plan.manaTaps {
		if tap.permanent.Tapped || effectiveController(g, tap.permanent) != player.ID {
			return false
		}
		output, ok := permanentManaOutput(g, tap.permanent)
		if !ok || output.color != tap.color || output.amount != tap.amount || output.snow != tap.snow {
			return false
		}
		tappedMana[mana.Unit{Color: tap.color, Snow: tap.snow}] += tap.amount
	}
	for _, color := range paymentColors {
		for _, snow := range []bool{false, true} {
			unit := mana.Unit{Color: color, Snow: snow}
			if player.ManaPool.Units()[unit]+tappedMana[unit] < plan.poolSpend[unit] {
				return false
			}
		}
	}
	if player.Life < plan.lifePayment {
		return false
	}
	return true
}

func costRequirements(cost *mana.Cost, xValue int) (map[mana.Color]int, int, bool) {
	colored := make(map[mana.Color]int)
	if xValue < 0 {
		return nil, 0, false
	}
	if cost == nil {
		return colored, 0, true
	}

	generic := 0
	for _, symbol := range *cost {
		switch symbol.Kind {
		case mana.ColoredSymbol:
			colored[symbol.Color]++
		case mana.ColorlessSymbol:
			colored[mana.Colorless]++
		case mana.GenericSymbol:
			generic += symbol.Generic
		case mana.VariableSymbol:
			generic += xValue
		default:
			return nil, 0, false
		}
	}
	return colored, generic, true
}

func snapshotPool(player *game.Player) map[mana.Unit]int {
	return player.ManaPool.Units()
}

func payColoredSymbol(plan *paymentPlan, pool map[mana.Unit]int, sources map[mana.Color][]manaSource, symbol mana.Symbol, color mana.Color, method game.SymbolPaymentMethod) bool {
	if !paySpecificMana(plan, pool, sources, color) {
		return false
	}
	plan.symbolPayments = append(plan.symbolPayments, game.SymbolPayment{
		Symbol: symbol,
		Method: method,
		Color:  color,
	})
	return true
}

func paySpecificMana(plan *paymentPlan, pool map[mana.Unit]int, sources map[mana.Color][]manaSource, color mana.Color) bool {
	if spendUnitFromSnapshot(plan, pool, mana.Unit{Color: color}, 1) {
		return true
	}
	if source, ok := takeNonSnowManaSource(sources, color); ok {
		plan.manaTaps = append(plan.manaTaps, manaTap{permanent: source.permanent, color: source.color, amount: source.amount, snow: source.snow})
		pool[mana.Unit{Color: source.color, Snow: source.snow}] += source.amount
		return paySpecificMana(plan, pool, sources, color)
	}
	if spendUnitFromSnapshot(plan, pool, mana.Unit{Color: color, Snow: true}, 1) {
		return true
	}
	source, ok := takeManaSource(sources, color)
	if !ok {
		return false
	}
	plan.manaTaps = append(plan.manaTaps, manaTap{permanent: source.permanent, color: source.color, amount: source.amount, snow: source.snow})
	pool[mana.Unit{Color: source.color, Snow: source.snow}] += source.amount
	return paySpecificMana(plan, pool, sources, color)
}

func payGenericSymbol(plan *paymentPlan, pool map[mana.Unit]int, sources map[mana.Color][]manaSource, symbol mana.Symbol, amount int, method game.SymbolPaymentMethod) bool {
	if amount < 0 {
		return false
	}
	if !payGenericMana(plan, pool, sources, amount) {
		return false
	}
	plan.symbolPayments = append(plan.symbolPayments, game.SymbolPayment{
		Symbol:        symbol,
		Method:        method,
		GenericAmount: amount,
	})
	return true
}

func payGenericMana(plan *paymentPlan, pool map[mana.Unit]int, sources map[mana.Color][]manaSource, amount int) bool {
	remaining := amount
	for remaining > 0 {
		if spendAnyUnitFromSnapshot(plan, pool) {
			remaining--
			continue
		}
		source, ok := takeAnyManaSource(sources)
		if !ok {
			return false
		}
		plan.manaTaps = append(plan.manaTaps, manaTap{permanent: source.permanent, color: source.color, amount: source.amount, snow: source.snow})
		pool[mana.Unit{Color: source.color, Snow: source.snow}] += source.amount
	}
	return true
}

func payHybridSymbol(plan *paymentPlan, pool map[mana.Unit]int, sources map[mana.Color][]manaSource, symbol mana.Symbol) bool {
	if trySymbolPayment(plan, pool, sources, func(trialPlan *paymentPlan, trialPool map[mana.Unit]int, trialSources map[mana.Color][]manaSource) bool {
		return payColoredSymbol(trialPlan, trialPool, trialSources, symbol, symbol.Color, game.SymbolPaymentHybridFirst)
	}) {
		return true
	}
	return trySymbolPayment(plan, pool, sources, func(trialPlan *paymentPlan, trialPool map[mana.Unit]int, trialSources map[mana.Color][]manaSource) bool {
		return payColoredSymbol(trialPlan, trialPool, trialSources, symbol, symbol.AltColor, game.SymbolPaymentHybridSecond)
	})
}

func payMonoHybridSymbol(plan *paymentPlan, pool map[mana.Unit]int, sources map[mana.Color][]manaSource, symbol mana.Symbol) bool {
	if trySymbolPayment(plan, pool, sources, func(trialPlan *paymentPlan, trialPool map[mana.Unit]int, trialSources map[mana.Color][]manaSource) bool {
		return payColoredSymbol(trialPlan, trialPool, trialSources, symbol, symbol.Color, game.SymbolPaymentMonoHybridColor)
	}) {
		return true
	}
	return trySymbolPayment(plan, pool, sources, func(trialPlan *paymentPlan, trialPool map[mana.Unit]int, trialSources map[mana.Color][]manaSource) bool {
		return payGenericSymbol(trialPlan, trialPool, trialSources, symbol, 2, game.SymbolPaymentMonoHybridGeneric)
	})
}

func paySnowSymbol(plan *paymentPlan, pool map[mana.Unit]int, sources map[mana.Color][]manaSource, symbol mana.Symbol) bool {
	if !paySnowMana(plan, pool, sources) {
		return false
	}
	plan.symbolPayments = append(plan.symbolPayments, game.SymbolPayment{
		Symbol: symbol,
		Method: game.SymbolPaymentSnow,
		Snow:   true,
	})
	return true
}

func paySnowMana(plan *paymentPlan, pool map[mana.Unit]int, sources map[mana.Color][]manaSource) bool {
	if spendAnySnowUnitFromSnapshot(plan, pool) {
		return true
	}
	source, ok := takeAnySnowManaSource(sources)
	if !ok {
		return false
	}
	plan.manaTaps = append(plan.manaTaps, manaTap{permanent: source.permanent, color: source.color, amount: source.amount, snow: source.snow})
	pool[mana.Unit{Color: source.color, Snow: source.snow}] += source.amount
	return spendAnySnowUnitFromSnapshot(plan, pool)
}

func payPhyrexianSymbol(player *game.Player, plan *paymentPlan, pool map[mana.Unit]int, sources map[mana.Color][]manaSource, symbol mana.Symbol, prefs *paymentPreferences) bool {
	if prefs != nil && prefs.nextPhyrexianLifeChoice() {
		if player.Life-plan.lifePayment < 2 {
			return false
		}
		plan.lifePayment += 2
		plan.symbolPayments = append(plan.symbolPayments, game.SymbolPayment{
			Symbol:   symbol,
			Method:   game.SymbolPaymentPhyrexianLife,
			LifePaid: 2,
		})
		return true
	}
	if trySymbolPayment(plan, pool, sources, func(trialPlan *paymentPlan, trialPool map[mana.Unit]int, trialSources map[mana.Color][]manaSource) bool {
		return payColoredSymbol(trialPlan, trialPool, trialSources, symbol, symbol.Color, game.SymbolPaymentPhyrexianMana)
	}) {
		return true
	}
	if player.Life-plan.lifePayment < 2 {
		return false
	}
	plan.lifePayment += 2
	plan.symbolPayments = append(plan.symbolPayments, game.SymbolPayment{
		Symbol:   symbol,
		Method:   game.SymbolPaymentPhyrexianLife,
		LifePaid: 2,
	})
	return true
}

func spendUnitFromSnapshot(plan *paymentPlan, pool map[mana.Unit]int, unit mana.Unit, amount int) bool {
	if amount <= 0 {
		return true
	}
	if pool[unit] < amount {
		return false
	}
	pool[unit] -= amount
	plan.poolSpend[unit] += amount
	return true
}

func spendAnyUnitFromSnapshot(plan *paymentPlan, pool map[mana.Unit]int) bool {
	for _, unit := range paymentUnitOrder() {
		if spendUnitFromSnapshot(plan, pool, unit, 1) {
			return true
		}
	}
	return false
}

func spendAnySnowUnitFromSnapshot(plan *paymentPlan, pool map[mana.Unit]int) bool {
	for _, unit := range paymentUnitOrder() {
		if !unit.Snow {
			continue
		}
		if spendUnitFromSnapshot(plan, pool, unit, 1) {
			return true
		}
	}
	return false
}

func paymentUnitOrder() []mana.Unit {
	var units []mana.Unit
	for _, color := range paymentColors {
		units = append(units, mana.Unit{Color: color})
	}
	for _, color := range paymentColors {
		units = append(units, mana.Unit{Color: color, Snow: true})
	}
	return units
}

func trySymbolPayment(plan *paymentPlan, pool map[mana.Unit]int, sources map[mana.Color][]manaSource, pay func(*paymentPlan, map[mana.Unit]int, map[mana.Color][]manaSource) bool) bool {
	trialPlan := clonePaymentPlan(*plan)
	trialPool := cloneUnitCounts(pool)
	trialSources := cloneManaSources(sources)
	if !pay(&trialPlan, trialPool, trialSources) {
		return false
	}
	*plan = trialPlan
	replaceUnitCounts(pool, trialPool)
	replaceManaSources(sources, trialSources)
	return true
}

func clonePaymentPlan(plan paymentPlan) paymentPlan {
	plan.poolSpend = cloneUnitCounts(plan.poolSpend)
	plan.manaTaps = append([]manaTap(nil), plan.manaTaps...)
	plan.symbolPayments = append([]game.SymbolPayment(nil), plan.symbolPayments...)
	return plan
}

func cloneUnitCounts(units map[mana.Unit]int) map[mana.Unit]int {
	clone := make(map[mana.Unit]int, len(units))
	for unit, amount := range units {
		clone[unit] = amount
	}
	return clone
}

func replaceUnitCounts(dst, src map[mana.Unit]int) {
	for unit := range dst {
		delete(dst, unit)
	}
	for unit, amount := range src {
		dst[unit] = amount
	}
}

func availableManaSources(g *game.Game, playerID game.PlayerID, exclude map[id.ID]bool) map[mana.Color][]manaSource {
	available := make(map[mana.Color][]manaSource)
	for _, permanent := range g.Battlefield {
		if effectiveController(g, permanent) != playerID || permanent.Tapped || exclude[permanent.ObjectID] {
			continue
		}
		output, ok := permanentManaOutput(g, permanent)
		if !ok {
			continue
		}
		available[output.color] = append(available[output.color], manaSource{
			permanent: permanent,
			color:     output.color,
			amount:    output.amount,
			snow:      output.snow,
		})
	}
	return available
}

func tapPermanentForMana(g *game.Game, permanent *game.Permanent, color mana.Color, amount int, snow bool) bool {
	if permanent.Tapped {
		return false
	}
	player, ok := playerForCostPayment(g, effectiveController(g, permanent))
	if !ok {
		return false
	}
	output, ok := permanentManaOutput(g, permanent)
	if !ok || output.color != color || output.amount != amount || output.snow != snow {
		return false
	}
	setPermanentTapped(g, permanent, true)
	if output.snow {
		player.ManaPool.AddSnow(color, amount)
	} else {
		player.ManaPool.Add(color, amount)
	}
	return true
}

type manaOutput struct {
	color  mana.Color
	amount int
	snow   bool
}

func permanentManaOutput(g *game.Game, permanent *game.Permanent) (manaOutput, bool) {
	if color, ok := basicLandManaColor(g, permanent); ok {
		return manaOutput{color: color, amount: 1, snow: permanentIsSnow(g, permanent)}, true
	}
	_, ability, ok := simpleTapManaAbility(g, permanent)
	if !ok {
		return manaOutput{}, false
	}
	amount := ability.Effects[0].Amount
	if amount <= 0 {
		amount = 1
	}
	return manaOutput{color: ability.Effects[0].ManaColor, amount: amount, snow: permanentIsSnow(g, permanent)}, true
}

func basicLandManaColor(g *game.Game, permanent *game.Permanent) (mana.Color, bool) {
	card, ok := permanentCardDef(g, permanent)
	if !ok || !card.HasType(game.TypeLand) {
		return 0, false
	}
	for _, landType := range basicLandTypes {
		if card.HasSubtype(landType.subtype) || strings.EqualFold(card.Name, landType.subtype) {
			return landType.color, true
		}
	}
	return 0, false
}

func permanentIsSnow(g *game.Game, permanent *game.Permanent) bool {
	return permanentHasSupertype(g, permanent, game.Snow)
}

var basicLandTypes = []struct {
	subtype string
	color   mana.Color
}{
	{subtype: "Plains", color: mana.White},
	{subtype: "Island", color: mana.Blue},
	{subtype: "Swamp", color: mana.Black},
	{subtype: "Mountain", color: mana.Red},
	{subtype: "Forest", color: mana.Green},
}

func simpleTapManaAbility(g *game.Game, permanent *game.Permanent) (int, *game.AbilityDef, bool) {
	card, ok := permanentCardDef(g, permanent)
	if !ok {
		return 0, nil, false
	}
	for i := range card.Abilities {
		ability := &card.Abilities[i]
		if ability.Kind == game.ActivatedAbility &&
			ability.IsManaAbility &&
			hasTapCost(ability) &&
			!ability.ManaCost.Exists &&
			len(ability.Targets) == 0 &&
			len(ability.Effects) == 1 &&
			ability.Effects[0].Type == game.EffectAddMana {
			if permanentHasType(g, permanent, game.TypeCreature) && permanent.SummoningSick {
				return 0, nil, false
			}
			return i, ability, true
		}
	}
	return 0, nil, false
}

func takeManaSource(sources map[mana.Color][]manaSource, color mana.Color) (manaSource, bool) {
	if source, ok := takeNonSnowManaSource(sources, color); ok {
		return source, true
	}
	if len(sources[color]) > 0 {
		source := sources[color][0]
		sources[color] = sources[color][1:]
		return source, true
	}
	return manaSource{}, false
}

func takeNonSnowManaSource(sources map[mana.Color][]manaSource, color mana.Color) (manaSource, bool) {
	for i, source := range sources[color] {
		if source.snow {
			continue
		}
		sources[color] = append(sources[color][:i], sources[color][i+1:]...)
		return source, true
	}
	return manaSource{}, false
}

func takeAnyManaSource(sources map[mana.Color][]manaSource) (manaSource, bool) {
	for _, color := range paymentColors {
		if source, ok := takeManaSource(sources, color); ok {
			return source, true
		}
	}
	return manaSource{}, false
}

func takeAnySnowManaSource(sources map[mana.Color][]manaSource) (manaSource, bool) {
	for _, color := range paymentColors {
		for i, source := range sources[color] {
			if !source.snow {
				continue
			}
			sources[color] = append(sources[color][:i], sources[color][i+1:]...)
			return source, true
		}
	}
	return manaSource{}, false
}

func cloneManaSources(sources map[mana.Color][]manaSource) map[mana.Color][]manaSource {
	clone := make(map[mana.Color][]manaSource, len(sources))
	for color, colorSources := range sources {
		clone[color] = append([]manaSource(nil), colorSources...)
	}
	return clone
}

func replaceManaSources(dst, src map[mana.Color][]manaSource) {
	for color := range dst {
		delete(dst, color)
	}
	for color, colorSources := range src {
		dst[color] = append([]manaSource(nil), colorSources...)
	}
}

func playerForCostPayment(g *game.Game, playerID game.PlayerID) (*game.Player, bool) {
	if playerID < 0 || int(playerID) >= len(g.Players) {
		return nil, false
	}
	player := g.Players[playerID]
	if player.Eliminated || g.TurnOrder.IsEliminated(playerID) {
		return nil, false
	}
	return player, true
}
