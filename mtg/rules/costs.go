package rules

import (
	"strings"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
)

var paymentColors = []mana.Color{
	mana.White,
	mana.Blue,
	mana.Black,
	mana.Red,
	mana.Green,
	mana.Colorless,
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
	_, ok := buildSpellCostPlan(g, playerID, card, xValue)
	return ok
}

func paySpellCosts(g *game.Game, playerID game.PlayerID, card *game.CardDef, xValue int) ([]string, bool) {
	plan, ok := buildSpellCostPlan(g, playerID, card, xValue)
	if !ok {
		return nil, false
	}
	player := playerForCostPayment(g, playerID)
	if player == nil || !additionalCostPlanStillValid(g, player, plan.additional) || !paymentPlanStillValid(g, player, plan.mana) {
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
	poolSpend map[mana.Color]int
	manaTaps  []manaTap
}

type spellCostPlan struct {
	mana       paymentPlan
	additional additionalCostPlan
}

type abilityCostPlan struct {
	mana       paymentPlan
	additional additionalCostPlan
	tapSource  bool
}

type additionalCostPlan struct {
	paid      []string
	sacrifice *game.Permanent
}

type manaTap struct {
	permanent *game.Permanent
	color     mana.Color
	amount    int
}

type manaSource struct {
	permanent *game.Permanent
	color     mana.Color
	amount    int
}

func buildSpellCostPlan(g *game.Game, playerID game.PlayerID, card *game.CardDef, xValue int) (spellCostPlan, bool) {
	plan := spellCostPlan{}
	additional, ok := buildAdditionalCostPlanForCost(g, playerID, spellAdditionalCost(card))
	if !ok {
		return plan, false
	}
	excluded := make(map[id.ID]bool)
	if additional.sacrifice != nil {
		excluded[additional.sacrifice.ObjectID] = true
	}
	manaPlan, ok := buildPaymentPlan(g, playerID, card.ManaCost, xValue, excluded)
	if !ok {
		return plan, false
	}
	plan.additional = additional
	plan.mana = manaPlan
	return plan, true
}

func buildAbilityCostPlan(g *game.Game, playerID game.PlayerID, source *game.Permanent, ability *game.AbilityDef, xValue int) (abilityCostPlan, bool) {
	plan := abilityCostPlan{}
	if source == nil || ability == nil {
		return plan, false
	}
	if xValue != 0 && !costHasVariableMana(ability.ManaCost) {
		return plan, false
	}
	tapSource := hasTapCost(ability)
	if tapSource && !canTapPermanentForAbility(g, source) {
		return plan, false
	}
	additional, ok := buildAdditionalCostPlanForCost(g, playerID, ability.AdditionalCost)
	if !ok {
		return plan, false
	}
	excluded := make(map[id.ID]bool)
	if tapSource {
		excluded[source.ObjectID] = true
	}
	if additional.sacrifice != nil {
		excluded[additional.sacrifice.ObjectID] = true
	}
	manaPlan, ok := buildPaymentPlan(g, playerID, ability.ManaCost, xValue, excluded)
	if !ok {
		return plan, false
	}
	plan.mana = manaPlan
	plan.additional = additional
	plan.tapSource = tapSource
	return plan, true
}

func buildPaymentPlan(g *game.Game, playerID game.PlayerID, cost *mana.Cost, xValue int, exclude map[id.ID]bool) (paymentPlan, bool) {
	plan := paymentPlan{poolSpend: make(map[mana.Color]int)}
	player := playerForCostPayment(g, playerID)
	if player == nil {
		return plan, false
	}
	colored, generic, ok := costRequirements(cost, xValue)
	if !ok {
		return plan, false
	}

	pool := snapshotPool(player)
	manaSources := availableManaSources(g, playerID, exclude)

	for _, color := range paymentColors {
		need := colored[color]
		if need == 0 {
			continue
		}

		spent := spendSnapshot(pool, color, need)
		need -= spent
		if spent > 0 {
			plan.poolSpend[color] += spent
		}
		for need > 0 {
			source := takeManaSource(manaSources, color)
			if source == nil {
				return plan, false
			}
			plan.manaTaps = append(plan.manaTaps, manaTap{permanent: source.permanent, color: color, amount: source.amount})
			pool[color] += source.amount
			spent := spendSnapshot(pool, color, need)
			need -= spent
			plan.poolSpend[color] += spent
		}
	}

	remainingGeneric := generic
	for _, color := range paymentColors {
		if remainingGeneric == 0 {
			break
		}
		spent := spendSnapshot(pool, color, remainingGeneric)
		remainingGeneric -= spent
		if spent > 0 {
			plan.poolSpend[color] += spent
		}
	}

	for remainingGeneric > 0 {
		source := takeAnyManaSource(manaSources)
		if source == nil {
			return plan, false
		}
		plan.manaTaps = append(plan.manaTaps, manaTap{permanent: source.permanent, color: source.color, amount: source.amount})
		pool[source.color] += source.amount
		spent := spendSnapshot(pool, source.color, remainingGeneric)
		remainingGeneric -= spent
		plan.poolSpend[source.color] += spent
	}

	return plan, true
}

func buildAdditionalCostPlan(g *game.Game, playerID game.PlayerID, card *game.CardDef) (additionalCostPlan, bool) {
	return buildAdditionalCostPlanForCost(g, playerID, spellAdditionalCost(card))
}

func buildAdditionalCostPlanForCost(g *game.Game, playerID game.PlayerID, cost string) (additionalCostPlan, bool) {
	plan := additionalCostPlan{}
	if cost == "" || isTapCost(cost) {
		return plan, true
	}
	matches, ok := sacrificeCostMatcher(cost)
	if !ok {
		return plan, false
	}
	permanent := chooseSacrificePermanent(g, playerID, matches)
	if permanent == nil {
		return plan, false
	}
	plan.paid = []string{cost}
	plan.sacrifice = permanent
	return plan, true
}

func spellAdditionalCost(card *game.CardDef) string {
	if card == nil {
		return ""
	}
	for _, ability := range card.Abilities {
		if ability.Kind == game.SpellAbility && ability.AdditionalCost != "" {
			return ability.AdditionalCost
		}
	}
	return ""
}

func sacrificeCostMatcher(cost string) (func(*game.CardDef) bool, bool) {
	normalized := strings.TrimSuffix(strings.ToLower(strings.TrimSpace(cost)), ".")
	switch normalized {
	case "sacrifice a creature":
		return func(card *game.CardDef) bool { return card != nil && card.HasType(game.TypeCreature) }, true
	case "sacrifice an artifact":
		return func(card *game.CardDef) bool { return card != nil && card.HasType(game.TypeArtifact) }, true
	case "sacrifice an enchantment":
		return func(card *game.CardDef) bool { return card != nil && card.HasType(game.TypeEnchantment) }, true
	case "sacrifice a land":
		return func(card *game.CardDef) bool { return card != nil && card.HasType(game.TypeLand) }, true
	case "sacrifice a permanent":
		return func(card *game.CardDef) bool { return card != nil }, true
	default:
		return nil, false
	}
}

func chooseSacrificePermanent(g *game.Game, playerID game.PlayerID, matches func(*game.CardDef) bool) *game.Permanent {
	if g == nil || matches == nil {
		return nil
	}
	for _, permanent := range g.Battlefield {
		if permanent == nil || permanent.Controller != playerID {
			continue
		}
		if matches(permanentCardDef(g, permanent)) {
			return permanent
		}
	}
	return nil
}

func additionalCostPlanStillValid(g *game.Game, player *game.Player, plan additionalCostPlan) bool {
	if player == nil {
		return false
	}
	if plan.sacrifice == nil {
		return true
	}
	permanent := permanentByObjectID(g, plan.sacrifice.ObjectID)
	return permanent != nil && permanent.Controller == player.ID && permanent == plan.sacrifice
}

func applyAdditionalCostPlan(g *game.Game, plan additionalCostPlan) bool {
	if plan.sacrifice == nil {
		return true
	}
	return movePermanentToZone(g, plan.sacrifice, game.ZoneGraveyard)
}

func payAbilityCosts(g *game.Game, playerID game.PlayerID, source *game.Permanent, ability *game.AbilityDef, xValue int) (abilityCostPlan, bool) {
	plan, ok := buildAbilityCostPlan(g, playerID, source, ability, xValue)
	if !ok {
		return plan, false
	}
	player := playerForCostPayment(g, playerID)
	if player == nil || !abilityCostPlanStillValid(g, player, source, plan) {
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
	if player == nil || source == nil {
		return false
	}
	if plan.tapSource && !canTapPermanentForAbility(g, source) {
		return false
	}
	return additionalCostPlanStillValid(g, player, plan.additional) &&
		paymentPlanStillValid(g, player, plan.mana)
}

func applyPaymentPlan(g *game.Game, playerID game.PlayerID, plan paymentPlan) bool {
	player := playerForCostPayment(g, playerID)
	if player == nil || !paymentPlanStillValid(g, player, plan) {
		return false
	}
	for _, tap := range plan.manaTaps {
		if !tapPermanentForMana(g, tap.permanent, tap.color, tap.amount) {
			panic("payment plan became invalid while tapping mana sources")
		}
	}
	for _, color := range paymentColors {
		amount := plan.poolSpend[color]
		if amount > 0 && !player.ManaPool.Spend(color, amount) {
			panic("payment plan became invalid while spending mana")
		}
	}
	return true
}

func paymentPlanStillValid(g *game.Game, player *game.Player, plan paymentPlan) bool {
	tappedMana := make(map[mana.Color]int)
	for _, tap := range plan.manaTaps {
		if tap.permanent == nil || tap.permanent.Tapped || tap.permanent.Controller != player.ID {
			return false
		}
		color, amount, ok := permanentManaOutput(g, tap.permanent)
		if !ok || color != tap.color || amount != tap.amount {
			return false
		}
		tappedMana[tap.color] += tap.amount
	}
	for _, color := range paymentColors {
		if player.ManaPool.Amount(color)+tappedMana[color] < plan.poolSpend[color] {
			return false
		}
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

func snapshotPool(player *game.Player) map[mana.Color]int {
	pool := make(map[mana.Color]int)
	for _, color := range paymentColors {
		pool[color] = player.ManaPool.Amount(color)
	}
	return pool
}

func spendSnapshot(pool map[mana.Color]int, color mana.Color, amount int) int {
	if amount <= 0 {
		return 0
	}
	spent := min(pool[color], amount)
	pool[color] -= spent
	return spent
}

func availableManaSources(g *game.Game, playerID game.PlayerID, exclude map[id.ID]bool) map[mana.Color][]manaSource {
	available := make(map[mana.Color][]manaSource)
	if g == nil {
		return available
	}
	for _, permanent := range g.Battlefield {
		if permanent == nil || permanent.Controller != playerID || permanent.Tapped || exclude[permanent.ObjectID] {
			continue
		}
		color, amount, ok := permanentManaOutput(g, permanent)
		if !ok {
			continue
		}
		available[color] = append(available[color], manaSource{permanent: permanent, color: color, amount: amount})
	}
	return available
}

func tapPermanentForMana(g *game.Game, permanent *game.Permanent, color mana.Color, amount int) bool {
	if g == nil || permanent == nil || permanent.Tapped {
		return false
	}
	player := playerForCostPayment(g, permanent.Controller)
	if player == nil {
		return false
	}
	sourceColor, sourceAmount, ok := permanentManaOutput(g, permanent)
	if !ok || sourceColor != color || sourceAmount != amount {
		return false
	}
	permanent.Tapped = true
	player.ManaPool.Add(color, amount)
	return true
}

func permanentManaOutput(g *game.Game, permanent *game.Permanent) (mana.Color, int, bool) {
	if color, ok := basicLandManaColor(g, permanent); ok {
		return color, 1, true
	}
	_, ability, ok := simpleTapManaAbility(g, permanent)
	if !ok {
		return 0, 0, false
	}
	amount := ability.Effects[0].Amount
	if amount <= 0 {
		amount = 1
	}
	return ability.Effects[0].ManaColor, amount, true
}

func basicLandManaColor(g *game.Game, permanent *game.Permanent) (mana.Color, bool) {
	card := g.GetCardInstance(permanent.CardInstanceID)
	if card == nil || card.Def == nil || !card.Def.HasType(game.TypeLand) {
		return 0, false
	}
	for _, landType := range basicLandTypes {
		if card.Def.HasSubtype(landType.subtype) || strings.EqualFold(card.Def.Name, landType.subtype) {
			return landType.color, true
		}
	}
	return 0, false
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
	card := permanentCardDef(g, permanent)
	if card == nil {
		return 0, nil, false
	}
	for i := range card.Abilities {
		ability := &card.Abilities[i]
		if ability.Kind == game.ActivatedAbility &&
			ability.IsManaAbility &&
			hasTapCost(ability) &&
			ability.ManaCost == nil &&
			len(ability.Targets) == 0 &&
			len(ability.Effects) == 1 &&
			ability.Effects[0].Type == game.EffectAddMana {
			if card.HasType(game.TypeCreature) && permanent.SummoningSick {
				return 0, nil, false
			}
			return i, ability, true
		}
	}
	return 0, nil, false
}

func takeManaSource(sources map[mana.Color][]manaSource, color mana.Color) *manaSource {
	if len(sources[color]) == 0 {
		return nil
	}
	source := sources[color][0]
	sources[color] = sources[color][1:]
	return &source
}

func takeAnyManaSource(sources map[mana.Color][]manaSource) *manaSource {
	for _, color := range paymentColors {
		if source := takeManaSource(sources, color); source != nil {
			return source
		}
	}
	return nil
}

func playerForCostPayment(g *game.Game, playerID game.PlayerID) *game.Player {
	if g == nil || playerID < 0 || int(playerID) >= len(g.Players) {
		return nil
	}
	player := g.Players[playerID]
	if player == nil || player.Eliminated || g.TurnOrder.IsEliminated(playerID) {
		return nil
	}
	return player
}
