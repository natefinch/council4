package rules

import (
	"slices"
	"strings"

	"github.com/natefinch/council4/mtg/game/zone"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/rules/payment"
	"github.com/natefinch/council4/opt"
)

func effectiveCyclingCost(g *game.Game, playerID game.PlayerID, card *game.CardInstance, body *game.ActivatedAbility) *cost.Mana {
	if body == nil {
		return nil
	}
	return applyAbilityCostModifiers(manaCostPtr(body.ManaCost), cyclingCostModifiers(g, playerID, card, body))
}

func effectiveHandAbilityCost(g *game.Game, playerID game.PlayerID, card *game.CardInstance, body *game.ActivatedAbility) *cost.Mana {
	return manaCostPtr(effectiveActivatedAbilityCost(g, playerID, card, body))
}

func effectiveActivatedAbilityCost(g *game.Game, playerID game.PlayerID, card *game.CardInstance, body *game.ActivatedAbility) opt.V[cost.Mana] {
	if body == nil {
		return opt.V[cost.Mana]{}
	}
	modifiers := cyclingCostModifiers(g, playerID, card, body)
	for _, modifier := range body.CostModifiers {
		if !costModifierAppliesToAbility(g, modifier, playerID, card, body) {
			continue
		}
		if modifier.PerObjectReduction > 0 {
			count := countPermanentsMatchingGroup(g, nil, playerID, game.BattlefieldGroup(*modifier.CountSelection))
			modifier.GenericReduction += count * modifier.PerObjectReduction
			modifier.PerObjectReduction = 0
			modifier.CountSelection = nil
		}
		modifiers = append(modifiers, modifier)
	}
	effective := applyAbilityCostModifiers(manaCostPtr(body.ManaCost), modifiers)
	if effective == nil {
		return opt.V[cost.Mana]{}
	}
	return opt.Val(*effective)
}

func cyclingCostModifiers(g *game.Game, playerID game.PlayerID, card *game.CardInstance, body *game.ActivatedAbility) []game.CostModifier {
	var modifiers []game.CostModifier
	for _, modifier := range g.CostModifiers {
		if costModifierAppliesToAbility(g, modifier, playerID, card, body) {
			modifiers = append(modifiers, modifier)
		}
	}
	effects := activeRuleEffects(g)
	for i := range effects {
		effect := &effects[i]
		if effect.Kind != game.RuleEffectCostModifier ||
			!playerRelationMatches(effect.Controller, playerID, effect.AffectedPlayer) {
			continue
		}
		if costModifierAppliesToAbility(g, effect.CostModifier, playerID, card, body) {
			modifiers = append(modifiers, effect.CostModifier)
		}
	}
	return modifiers
}

func costModifierAppliesToAbility(g *game.Game, modifier game.CostModifier, playerID game.PlayerID, card *game.CardInstance, body *game.ActivatedAbility) bool {
	if modifier.Kind != game.CostModifierAbility {
		return false
	}
	if modifier.AbilityKeyword != game.KeywordNone && !game.BodyHasKeyword(body, modifier.AbilityKeyword) {
		return false
	}
	if !modifier.CardSelection.Empty() {
		if card == nil || !cardDefMatchesCostSelection(g, card.Def, modifier.CardSelection) {
			return false
		}
	}
	if modifier.FirstCycleEachTurn && playerCycledThisTurn(g, playerID) {
		return false
	}
	return true
}

func playerCycledThisTurn(g *game.Game, playerID game.PlayerID) bool {
	for _, event := range g.EventsThisTurn() {
		if event.Kind == game.EventCycled && event.Player == playerID {
			return true
		}
	}
	return false
}

func applyAbilityCostModifiers(manaCost *cost.Mana, modifiers []game.CostModifier) *cost.Mana {
	if len(modifiers) == 0 {
		return manaCost
	}
	baseCost := manaCost
	for i := range modifiers {
		if modifiers[i].SetManaCost.Exists {
			value := append(cost.Mana(nil), modifiers[i].SetManaCost.Val...)
			baseCost = &value
		}
	}
	generic := genericCostAmount(baseCost)
	minimum := 0
	setGeneric := (*int)(nil)
	for i := range modifiers {
		modifier := modifiers[i]
		if modifier.SetGeneric.Exists {
			setGeneric = &modifier.SetGeneric.Val
		}
		generic += modifier.GenericIncrease
		generic -= modifier.GenericReduction
		if modifier.MinimumGeneric > minimum {
			minimum = modifier.MinimumGeneric
		}
	}
	if setGeneric != nil {
		generic = *setGeneric
	}
	if generic < minimum {
		generic = minimum
	}
	if generic < 0 {
		generic = 0
	}
	return costWithGenericAmount(baseCost, generic)
}

func genericCostAmount(manaCost *cost.Mana) int {
	if manaCost == nil {
		return 0
	}
	total := 0
	for _, symbol := range *manaCost {
		if symbol.Kind == cost.GenericSymbol {
			total += symbol.Generic
		}
	}
	return total
}

func costWithGenericAmount(manaCost *cost.Mana, generic int) *cost.Mana {
	var modified cost.Mana
	if generic > 0 {
		modified = append(modified, cost.O(generic))
	}
	if manaCost != nil {
		for _, symbol := range *manaCost {
			if symbol.Kind != cost.GenericSymbol {
				modified = append(modified, symbol)
			}
		}
	}
	return &modified
}

func abilityHasReturnUnblockedAttackerCost(costs []cost.Additional) bool {
	return len(costs) == 1 &&
		costs[0].Kind == cost.AdditionalReturnUnblockedAttacker &&
		payment.AdditionalCostAmount(costs[0]) == 1
}

func canAct(g *game.Game, playerID game.PlayerID) bool {
	return isPlayerAlive(g, playerID)
}

func canPlayAnyLand(g *game.Game, playerID game.PlayerID) bool {
	return canAct(g, playerID) &&
		playerID == g.Turn.ActivePlayer &&
		playerID == g.Turn.PriorityPlayer &&
		isSorcerySpeed(g, playerID) &&
		playerCanPlayLand(g, playerID)
}

// playerCanPlayLand reports whether the active player has not yet exhausted
// their land plays this turn, accounting for the one-land baseline plus any
// additional-land-play allowances granted by effects (Explore, Exploration,
// Azusa, etc.).
func playerCanPlayLand(g *game.Game, playerID game.PlayerID) bool {
	return g.Turn.LandsPlayedThisTurn < g.Turn.LandsAllowedThisTurn+additionalLandPlaysFor(g, playerID)
}

func canCastAtCurrentTiming(g *game.Game, playerID game.PlayerID, card *game.CardDef) bool {
	if card.HasType(types.Instant) || card.HasKeyword(game.Flash) {
		return true
	}
	if playerCanCastAsThoughFlash(g, playerID, card) {
		return true
	}
	return isSorcerySpeed(g, playerID)
}

func legalXValuesForCost(g *game.Game, playerID game.PlayerID, manaCost *cost.Mana) []int {
	if !costHasVariableMana(manaCost) {
		return []int{0}
	}
	var values []int
	for x := 0; x <= maxLegalXValue; x++ {
		if !paymentOrch.canPayGenericCost(g, payment.GenericRequest{PlayerID: playerID, Cost: manaCost, XValue: x}) {
			break
		}
		values = append(values, x)
	}
	return values
}

func legalXValuesForCostAndAdditional(g *game.Game, playerID game.PlayerID, manaCost *cost.Mana, additionalCosts []cost.Additional) []int {
	if !additionalCostsUseX(additionalCosts) {
		return legalXValuesForCost(g, playerID, manaCost)
	}
	upperBound := maxLegalXValue
	for _, additional := range additionalCosts {
		if additional.Kind != cost.AdditionalPayLife || !additional.AmountFromX {
			continue
		}
		player, ok := playerByID(g, playerID)
		if !ok {
			return nil
		}
		upperBound = player.Life
		break
	}
	var values []int
	for x := 0; x <= upperBound; x++ {
		if costHasVariableMana(manaCost) && !paymentOrch.canPayGenericCost(g, payment.GenericRequest{PlayerID: playerID, Cost: manaCost, XValue: x}) {
			break
		}
		values = append(values, x)
	}
	return values
}

func additionalCostsUseX(additionalCosts []cost.Additional) bool {
	for _, additional := range additionalCosts {
		if additional.AmountFromX {
			return true
		}
	}
	return false
}

func costHasVariableMana(manaCost *cost.Mana) bool {
	if manaCost == nil {
		return false
	}
	for _, symbol := range *manaCost {
		if symbol.Kind == cost.VariableSymbol {
			return true
		}
	}
	return false
}

func activatedAbilitySource(g *game.Game, playerID game.PlayerID, sourceID id.ID, abilityIndex int) (*game.Permanent, game.Ability, bool) {
	if abilityIndex < 0 {
		return nil, nil, false
	}
	permanent, ok := permanentByObjectID(g, sourceID)
	if !ok || permanent.PhasedOut || effectiveController(g, permanent) != playerID {
		return nil, nil, false
	}
	abilities := permanentEffectiveAbilities(g, permanent)
	if abilityIndex >= len(abilities) {
		return nil, nil, false
	}
	return permanent, abilities[abilityIndex], true
}

// isManaAbilityActivation reports whether activate selects a mana ability —
// one that produces mana and resolves without using the stack — whether printed
// on a battlefield permanent or on a card activated from hand. It mirrors the
// mana-ability dispatch in applyActivateAbility so the turn log can flag the
// action without re-running it.
func isManaAbilityActivation(g *game.Game, playerID game.PlayerID, activate action.ActivateAbilityAction) bool {
	if _, body, ok := activatedAbilitySource(g, playerID, activate.SourceID, activate.AbilityIndex); ok {
		if _, isMana := body.(*game.ManaAbility); isMana {
			return true
		}
	}
	if _, _, ok := handManaAbilitySource(g, playerID, activate.SourceID, activate.AbilityIndex); ok {
		return true
	}
	return false
}

func cyclingAbilitySource(g *game.Game, playerID game.PlayerID, sourceID id.ID, abilityIndex int) (*game.CardInstance, game.ActivatedAbility, bool) {
	return handActivatedAbilitySource(g, playerID, sourceID, abilityIndex)
}

func handActivatedAbilitySource(g *game.Game, playerID game.PlayerID, sourceID id.ID, abilityIndex int) (*game.CardInstance, game.ActivatedAbility, bool) {
	if abilityIndex < 0 {
		return nil, game.ActivatedAbility{}, false
	}

	player, ok := playerByID(g, playerID)
	if !ok || !player.Hand.Contains(sourceID) {
		return nil, game.ActivatedAbility{}, false
	}
	card, ok := g.GetCardInstance(sourceID)
	if !ok {
		return nil, game.ActivatedAbility{}, false
	}
	effectiveAbilities := effectiveHandActivatedAbilities(g, playerID, card)
	for i := range effectiveAbilities {
		effective := &effectiveAbilities[i]
		if effective.index == abilityIndex {
			return card, effective.body, true
		}
	}
	return nil, game.ActivatedAbility{}, false
}

// handManaAbilitySource returns the hand card and ManaAbility addressed by a
// canonical ability index, when the card is in the player's hand. It mirrors
// handActivatedAbilitySource for the mana-from-hand family (Simian/Elvish
// Spirit Guide).
func handManaAbilitySource(g *game.Game, playerID game.PlayerID, sourceID id.ID, abilityIndex int) (*game.CardInstance, game.ManaAbility, bool) {
	if abilityIndex < 0 {
		return nil, game.ManaAbility{}, false
	}
	player, ok := playerByID(g, playerID)
	if !ok || !player.Hand.Contains(sourceID) {
		return nil, game.ManaAbility{}, false
	}
	card, ok := g.GetCardInstance(sourceID)
	if !ok {
		return nil, game.ManaAbility{}, false
	}
	def := cardFaceOrDefault(card, game.FaceFront)
	for i := range def.ManaAbilities {
		if def.ManaAbilityIndex(i) == abilityIndex {
			return card, def.ManaAbilities[i], true
		}
	}
	return nil, game.ManaAbility{}, false
}

// abilityHasExileThisCardFromHandCost reports whether costs is exactly the
// "Exile this card from your hand" self-exile cost that gates the hand mana
// ability family. Any other cost shape fails closed.
func abilityHasExileThisCardFromHandCost(costs []cost.Additional) bool {
	return len(costs) == 1 &&
		costs[0].Kind == cost.AdditionalExileSource &&
		costs[0].Source == zone.Hand
}

// canActivateHandManaAbility reports whether the player may activate a mana
// ability printed on a card in their hand whose only cost is exiling that card
// from hand. It mirrors canActivateManaAbility but is keyed on a hand card
// rather than a battlefield permanent: there is no mana cost, no target, and no
// permanent to seed entry choices, so the only producible content is fixed
// add-mana output.
func canActivateHandManaAbility(g *game.Game, playerID game.PlayerID, cardID id.ID, body *game.ManaAbility, abilityIndex int) bool {
	if body == nil || !canAct(g, playerID) || playerID != g.Turn.PriorityPlayer {
		return false
	}
	if game.BodyFunctionZone(body) != zone.Hand ||
		!abilityHasExileThisCardFromHandCost(body.AdditionalCosts) ||
		(body.ManaCost.Exists && len(body.ManaCost.Val) != 0) {
		return false
	}
	if len(game.BodyTargets(body)) != 0 || !manaBodyHasAddManaEffect(body) {
		return false
	}
	if !activatedAbilityTimingAllows(g, playerID, body.Timing) ||
		activatedAbilityUsedThisTurn(g, cardID, abilityIndex, body.Timing) {
		return false
	}
	if !activationConditionSatisfied(g, playerID, nil, body.ActivationCondition) {
		return false
	}
	_, _, ok := handManaAbilitySource(g, playerID, cardID, abilityIndex)
	return ok
}

type indexedHandActivatedAbility struct {
	index int
	body  game.ActivatedAbility
}

func effectiveHandActivatedAbilities(g *game.Game, playerID game.PlayerID, card *game.CardInstance) []indexedHandActivatedAbility {
	if card == nil {
		return nil
	}
	frontDef := cardFaceOrDefault(card, game.FaceFront)
	abilities := make([]indexedHandActivatedAbility, 0, len(frontDef.ActivatedAbilities))
	seenCyclingCosts := []cost.Mana{}
	for i := range frontDef.ActivatedAbilities {
		body := frontDef.ActivatedAbilities[i]
		if cyclingCost, ok := game.ActivatedBodyCyclingCost(&body); ok {
			seenCyclingCosts = append(seenCyclingCosts, append(cost.Mana(nil), cyclingCost...))
		}
		abilities = append(abilities, indexedHandActivatedAbility{
			index: frontDef.ActivatedAbilityIndex(i),
			body:  body,
		})
	}

	nextIndex := frontDef.AbilityCount()
	activeEffects := activeRuleEffects(g)
	for i := range activeEffects {
		effect := &activeEffects[i]
		if effect.Kind != game.RuleEffectGrantHandCardAbility ||
			!playerRelationMatches(effect.Controller, playerID, effect.AffectedPlayer) ||
			!handCardMatchesSelection(g, card, effect.CardSelection, effect.Controller) {
			continue
		}
		cyclingCost, ok := game.ActivatedBodyCyclingCost(&effect.GrantedAbility)
		if !ok || slices.ContainsFunc(seenCyclingCosts, func(existing cost.Mana) bool {
			return slices.Equal(existing, cyclingCost)
		}) {
			continue
		}
		seenCyclingCosts = append(seenCyclingCosts, append(cost.Mana(nil), cyclingCost...))
		abilities = append(abilities, indexedHandActivatedAbility{
			index: nextIndex,
			body:  effect.GrantedAbility,
		})
		nextIndex++
	}
	return abilities
}

func handCardMatchesSelection(g *game.Game, card *game.CardInstance, selection game.Selection, viewer game.PlayerID) bool {
	if card == nil || card.Def == nil {
		return false
	}
	return matchSelection(&selectionSubject{
		kind:   subjectCard,
		g:      g,
		card:   card,
		viewer: viewer,
	}, &selection)
}

func graveyardAbilitySource(g *game.Game, playerID game.PlayerID, sourceID id.ID, abilityIndex int) (*game.CardInstance, game.ActivatedAbility, bool) {
	if abilityIndex < 0 {
		return nil, game.ActivatedAbility{}, false
	}
	player, ok := playerByID(g, playerID)
	if !ok || !player.Graveyard.Contains(sourceID) {
		return nil, game.ActivatedAbility{}, false
	}
	card, ok := g.GetCardInstance(sourceID)
	if !ok {
		return nil, game.ActivatedAbility{}, false
	}
	frontDef := cardFaceOrDefault(card, game.FaceFront)
	body, ok := frontDef.BodyAt(abilityIndex).(*game.ActivatedAbility)
	if !ok {
		return nil, game.ActivatedAbility{}, false
	}
	return card, *body, true
}

func canActivateEquipAbility(g *game.Game, playerID game.PlayerID, permanent *game.Permanent, body *game.ActivatedAbility, abilityIndex int, targets []game.Target, xValue int) bool {
	return canActivateEquipAbilityWithModes(g, playerID, permanent, body, abilityIndex, targets, xValue, nil)
}

func canActivateEquipAbilityWithModes(g *game.Game, playerID game.PlayerID, permanent *game.Permanent, body *game.ActivatedAbility, abilityIndex int, targets []game.Target, xValue int, chosenModes []int) bool {
	if body == nil || !canAct(g, playerID) || playerID != g.Turn.PriorityPlayer || permanent.PhasedOut || effectiveController(g, permanent) != playerID {
		return false
	}
	if xValue != 0 || !bodyFunctionsOnBattlefield(body) || !isEquipmentPermanent(g, permanent) {
		return false
	}
	if !bodyAttachesLikeEquip(body) && body.Timing != game.SorceryOnly {
		return false
	}
	if !isSorcerySpeed(g, playerID) || abilityHasNonTapAdditionalCosts(body.AdditionalCosts) || activatedAbilityUsedThisTurn(g, permanent.ObjectID, abilityIndex, body.Timing) {
		return false
	}
	if !activationConditionSatisfied(g, playerID, permanent, body.ActivationCondition) {
		return false
	}
	card, ok := permanentCardDef(g, permanent)
	if !ok || !modesValidForBody(body, chosenModes) ||
		!targetsValidForBodyFromSourceObjectWithModes(g, playerID, card, permanent.ObjectID, body, chosenModes, targets) {
		return false
	}
	if len(targets) != 1 || targets[0].Kind != game.TargetPermanent {
		return false
	}
	target, ok := permanentByObjectID(g, targets[0].PermanentID)
	if !ok || effectiveController(g, target) != playerID || !canAttachPermanent(g, permanent, target) {
		return false
	}
	sourceCard, _ := g.GetCardInstance(permanent.CardInstanceID)
	return paymentOrch.canPayGenericCost(g, payment.GenericRequest{
		PlayerID: playerID,
		Cost:     manaCostPtr(effectiveActivatedAbilityCost(g, playerID, sourceCard, body)),
	})
}

func canActivateGeneralAbility(g *game.Game, playerID game.PlayerID, permanent *game.Permanent, body *game.ActivatedAbility, abilityIndex int, targets []game.Target, xValue int) bool {
	return canActivateGeneralAbilityWithModes(g, playerID, permanent, body, abilityIndex, targets, xValue, nil)
}

func canActivateGeneralAbilityWithModes(g *game.Game, playerID game.PlayerID, permanent *game.Permanent, body *game.ActivatedAbility, abilityIndex int, targets []game.Target, xValue int, chosenModes []int) bool {
	if body == nil || !canAct(g, playerID) || playerID != g.Turn.PriorityPlayer || permanent.PhasedOut || effectiveController(g, permanent) != playerID {
		return false
	}
	if bodyAttachesLikeEquip(body) || !bodyFunctionsOnBattlefield(body) {
		return false
	}
	if !activatedAbilityTimingAllows(g, playerID, body.Timing) || activatedAbilityUsedThisTurn(g, permanent.ObjectID, abilityIndex, body.Timing) {
		return false
	}
	if !activationConditionSatisfied(g, playerID, permanent, body.ActivationCondition) {
		return false
	}
	if abilityActivationProhibited(g, playerID, permanent) {
		return false
	}
	card, ok := permanentCardDef(g, permanent)
	if !ok || !modesValidForBody(body, chosenModes) ||
		!targetsValidForBodyFromSourceObjectWithModes(g, playerID, card, permanent.ObjectID, body, chosenModes, targets) {
		return false
	}
	sourceCard, _ := g.GetCardInstance(permanent.CardInstanceID)
	return paymentOrch.buildAbilityCostPlan(g, payment.AbilityRequest{
		PlayerID:         playerID,
		Source:           permanent,
		ManaCost:         effectiveActivatedAbilityCost(g, playerID, sourceCard, body),
		AdditionalCosts:  abilityAdditionalCosts(body.AdditionalCosts),
		AlternativeCosts: append([]cost.Alternative(nil), body.AlternativeCosts...),
		XValue:           xValue,
	})
}

func canActivateCyclingAbility(g *game.Game, playerID game.PlayerID, cardID id.ID, body *game.ActivatedAbility, abilityIndex int, targets []game.Target, xValue int) bool {
	if body == nil || !canAct(g, playerID) || playerID != g.Turn.PriorityPlayer {
		return false
	}

	if xValue != 0 || abilityIndex < 0 || !game.BodyHasKeyword(body, game.Cycling) {
		return false
	}
	if body.Timing != game.NoTimingRestriction || !abilityHasDiscardThisCardCost(body.AdditionalCosts) {
		return false
	}
	if len(targets) != 0 || len(game.BodyTargets(body)) != 0 {
		return false
	}
	card, gotAbility, ok := cyclingAbilitySource(g, playerID, cardID, abilityIndex)
	if !ok || !game.BodyHasKeyword(&gotAbility, game.Cycling) {
		return false
	}
	return paymentOrch.canPayGenericCost(g, payment.GenericRequest{PlayerID: playerID, Cost: effectiveCyclingCost(g, playerID, card, body)})
}

func canActivateGraveyardAbility(g *game.Game, playerID game.PlayerID, cardID id.ID, body *game.ActivatedAbility, abilityIndex int, targets []game.Target, xValue int) bool {
	return canActivateGraveyardAbilityWithModes(g, playerID, cardID, body, abilityIndex, targets, xValue, nil)
}

func canActivateHandAbilityWithModes(g *game.Game, playerID game.PlayerID, cardID id.ID, body *game.ActivatedAbility, abilityIndex int, targets []game.Target, xValue int, chosenModes []int) bool {
	if body == nil || !canAct(g, playerID) || playerID != g.Turn.PriorityPlayer {
		return false
	}
	if game.BodyFunctionZone(body) != zone.Hand ||
		game.BodyHasKeyword(body, game.Cycling) ||
		game.BodyHasKeyword(body, game.Ninjutsu) ||
		len(body.AdditionalCosts) != 1 ||
		!abilityHasDiscardThisCardCost(body.AdditionalCosts) {
		return false
	}
	if !activatedAbilityTimingAllows(g, playerID, body.Timing) ||
		activatedAbilityUsedThisTurn(g, cardID, abilityIndex, body.Timing) {
		return false
	}
	if !activationConditionSatisfied(g, playerID, nil, body.ActivationCondition) {
		return false
	}
	card, gotAbility, ok := handActivatedAbilitySource(g, playerID, cardID, abilityIndex)
	if !ok || game.BodyFunctionZone(&gotAbility) != zone.Hand {
		return false
	}
	def := cardFaceOrDefault(card, game.FaceFront)
	if !modesValidForBody(body, chosenModes) ||
		!targetsValidForBodyFromSourceObjectWithModes(g, playerID, def, 0, body, chosenModes, targets) {
		return false
	}
	return paymentOrch.canPayGenericCost(g, payment.GenericRequest{
		PlayerID: playerID,
		Cost:     effectiveHandAbilityCost(g, playerID, card, body),
		XValue:   xValue,
	})
}

func canActivateGraveyardAbilityWithModes(g *game.Game, playerID game.PlayerID, cardID id.ID, body *game.ActivatedAbility, abilityIndex int, targets []game.Target, xValue int, chosenModes []int) bool {
	if body == nil || !canAct(g, playerID) || playerID != g.Turn.PriorityPlayer {
		return false
	}
	if game.BodyFunctionZone(body) != zone.Graveyard {
		return false
	}
	if !activatedAbilityTimingAllows(g, playerID, body.Timing) || activatedAbilityUsedThisTurn(g, cardID, abilityIndex, body.Timing) {
		return false
	}
	if !activationConditionSatisfied(g, playerID, nil, body.ActivationCondition) {
		return false
	}
	card, _, ok := graveyardAbilitySource(g, playerID, cardID, abilityIndex)
	if !ok {
		return false
	}
	def := cardFaceOrDefault(card, game.FaceFront)
	if !modesValidForBody(body, chosenModes) ||
		!targetsValidForBodyFromSourceObjectWithModes(g, playerID, def, 0, body, chosenModes, targets) {
		return false
	}
	return paymentOrch.buildAbilityCostPlan(g, payment.AbilityRequest{
		PlayerID:         playerID,
		SourceCardID:     cardID,
		SourceZone:       zone.Graveyard,
		ManaCost:         body.ManaCost,
		AdditionalCosts:  abilityAdditionalCosts(body.AdditionalCosts),
		AlternativeCosts: append([]cost.Alternative(nil), body.AlternativeCosts...),
		XValue:           xValue,
	})
}

func canActivateManaAbility(g *game.Game, playerID game.PlayerID, permanent *game.Permanent, body *game.ManaAbility, abilityIndex int) bool {
	if body == nil || !canAct(g, playerID) || playerID != g.Turn.PriorityPlayer || permanent.PhasedOut || effectiveController(g, permanent) != playerID {
		return false
	}
	if !bodyFunctionsOnBattlefield(body) {
		return false
	}
	if len(game.BodyTargets(body)) != 0 || !manaBodyHasAddManaEffect(body) || !manaBodyChoicesAvailable(g, playerID, permanent, body) {
		return false
	}
	if !manaBodyEntryChoicesAvailable(permanent, body) {
		return false
	}
	if !activatedAbilityTimingAllows(g, playerID, body.Timing) ||
		activatedAbilityUsedThisTurn(g, permanent.ObjectID, abilityIndex, body.Timing) {
		return false
	}
	if !activationConditionSatisfied(g, playerID, permanent, body.ActivationCondition) {
		return false
	}
	if abilityActivationProhibited(g, playerID, permanent) {
		return false
	}
	return paymentOrch.buildAbilityCostPlan(g, payment.AbilityRequest{
		PlayerID:        playerID,
		Source:          permanent,
		ManaCost:        body.ManaCost,
		AdditionalCosts: abilityAdditionalCosts(body.AdditionalCosts),
	})
}

func manaBodyHasAddManaEffect(body *game.ManaAbility) bool {
	if body == nil {
		return false
	}
	if sequence, ok := manaBodyInstructionSequence(body); ok {
		hasAddMana := false
		for i := range sequence {
			if sequence[i].Primitive == nil {
				return false
			}
			switch sequence[i].Primitive.Kind() {
			case game.PrimitiveAddMana:
				hasAddMana = true
			case game.PrimitiveChoose:
				choice, ok := sequence[i].Primitive.(game.Choose)
				if !ok || choice.Choice.Kind != game.ResolutionChoiceMana {
					return false
				}
			case game.PrimitiveDamage:
				damage, ok := sequence[i].Primitive.(game.Damage)
				if !ok || !isSelfControllerDamageRider(damage) {
					return false
				}
			case game.PrimitiveGainLife:
				gain, ok := sequence[i].Primitive.(game.GainLife)
				if !ok || !isSelfControllerGainLifeRider(gain) {
					return false
				}
			default:
				return false
			}
		}
		return hasAddMana
	}
	return false
}

// isSelfControllerDamageRider reports whether a Damage instruction is a mana
// source's "deals N damage to you" rider: a non-divided amount the source
// permanent deals to its own controller. Painlands, the painland Talismans,
// Ancient Tomb, and Tarnished Citadel carry it. CR 605.1a keeps such abilities
// mana abilities because the rider neither targets nor stops them from adding
// mana, so the rider must not disqualify the ability from immediate resolution.
func isSelfControllerDamageRider(damage game.Damage) bool {
	if damage.Divided || !damage.DamageSource.Exists ||
		damage.DamageSource.Val.Kind() != game.ObjectReferenceSourcePermanent {
		return false
	}
	player, ok := damage.Recipient.PlayerReference()
	return ok && player.Kind() == game.PlayerReferenceController
}

// isSelfControllerGainLifeRider reports whether a GainLife instruction is a mana
// source's "You gain N life" rider: a non-group amount of life gained by the
// ability's own controller. The Great Henge carries it. CR 605.1a keeps such
// abilities mana abilities because the rider neither targets nor stops them from
// adding mana, so the rider must not disqualify the ability from immediate
// resolution.
func isSelfControllerGainLifeRider(gain game.GainLife) bool {
	return gain.Player.Kind() == game.PlayerReferenceController
}

func manaBodyChoicesAvailable(g *game.Game, playerID game.PlayerID, permanent *game.Permanent, body *game.ManaAbility) bool {
	if body == nil {
		return false
	}
	// A synthetic activated-ability object lets resolution choices that depend on
	// the source permanent's object identity (e.g. an imprinted card's colors)
	// resolve their options for the activation legality check, mirroring the
	// object built when the ability actually resolves.
	var obj *game.StackObject
	if permanent != nil {
		obj = &game.StackObject{
			Kind:         game.StackActivatedAbility,
			SourceID:     permanent.ObjectID,
			SourceCardID: permanent.CardInstanceID,
		}
	}
	if sequence, ok := manaBodyInstructionSequence(body); ok {
		for i := range sequence {
			if sequence[i].Primitive == nil {
				return false
			}
			if sequence[i].Primitive.Kind() != game.PrimitiveChoose {
				continue
			}
			primitive, ok := sequence[i].Primitive.(game.Choose)
			if !ok {
				return false
			}
			choicePlayer := resolutionChoicePlayer(playerID, &primitive.Choice)
			_, values := resolutionChoiceOptions(g, obj, choicePlayer, &primitive.Choice)
			if len(values) == 0 {
				return false
			}
		}
		return true
	}
	return true
}

func manaBodyInstructionSequence(body *game.ManaAbility) ([]game.Instruction, bool) {
	if body == nil {
		return nil, false
	}
	if len(body.Content.Modes) == 0 || body.Content.IsModal() {
		return nil, false
	}
	return body.Content.Modes[0].Sequence, true
}

// manaBodyEntryChoicesAvailable reports whether every entry-time choice a mana
// ability reads (AddMana{EntryChoiceFrom:...}) was recorded on the source
// permanent as it entered. A permanent missing the choice cannot produce that
// mana, so the ability is not activatable.
func manaBodyEntryChoicesAvailable(permanent *game.Permanent, body *game.ManaAbility) bool {
	sequence, ok := manaBodyInstructionSequence(body)
	if !ok {
		return true
	}
	for i := range sequence {
		addMana, ok := sequence[i].Primitive.(game.AddMana)
		if !ok {
			continue
		}
		if addMana.EntryChoiceFrom != "" && !permanentEntryChoiceAvailable(permanent, addMana.EntryChoiceFrom) {
			return false
		}
		if addMana.SpendRider.Exists &&
			addMana.SpendRider.Val.ChosenSubtypeFrom != "" &&
			!permanentEntryChoiceAvailable(permanent, addMana.SpendRider.Val.ChosenSubtypeFrom) {
			return false
		}
	}
	return true
}

func hasTapCostOf(additionalCosts []cost.Additional) bool {
	for _, addCost := range additionalCosts {
		if addCost.Kind == cost.AdditionalTap {
			return true
		}
	}
	return false
}

func abilityHasNonTapAdditionalCosts(additionalCosts []cost.Additional) bool {
	for _, addCost := range additionalCosts {
		if addCost.Kind != cost.AdditionalTap {
			return true
		}
	}
	return false
}

func abilityHasDiscardThisCardCost(costs []cost.Additional) bool {
	if len(costs) != 1 {
		return false
	}
	addCost := costs[0]
	if addCost.Kind != cost.AdditionalDiscard || payment.AdditionalCostAmount(addCost) != 1 {
		return false
	}
	if addCost.Text != "" {
		return strings.TrimSuffix(strings.ToLower(strings.TrimSpace(addCost.Text)), ".") == "discard this card"
	}
	return addCost.Source == zone.Hand
}

func canTapPermanentForAbility(g *game.Game, permanent *game.Permanent) bool {
	if permanent.Tapped {
		return false
	}
	return !permanentHasType(g, permanent, types.Creature) || !permanent.SummoningSick
}

func tapPermanentForAbility(g *game.Game, permanent *game.Permanent) bool {
	if !canTapPermanentForAbility(g, permanent) {
		return false
	}
	setPermanentTapped(g, permanent, true)
	return true
}

func activatedAbilityTimingAllows(g *game.Game, playerID game.PlayerID, timing game.TimingRestriction) bool {
	switch timing {
	case game.NoTimingRestriction, game.OncePerTurn:
		return true
	case game.SorceryOnly, game.SorceryOncePerTurn:
		return isSorcerySpeed(g, playerID)
	case game.DuringCombat:
		return g.Turn.Phase == game.PhaseCombat
	case game.DuringUpkeep:
		return g.Turn.ActivePlayer == playerID &&
			g.Turn.Phase == game.PhaseBeginning &&
			g.Turn.Step == game.StepUpkeep
	case game.DuringYourTurn:
		return g.Turn.ActivePlayer == playerID
	default:
		return false
	}
}

func activatedAbilityUsedThisTurn(g *game.Game, sourceID id.ID, abilityIndex int, timing game.TimingRestriction) bool {
	if !abilityHasOncePerTurnRestriction(timing) {
		return false
	}
	return g.ActivatedAbilitiesThisTurn[game.ActivatedAbilityUse{
		SourceID:     sourceID,
		AbilityIndex: abilityIndex,
	}]
}

func recordActivatedAbilityUse(g *game.Game, sourceID id.ID, abilityIndex int, timing game.TimingRestriction) {
	if !abilityHasOncePerTurnRestriction(timing) {
		return
	}
	if g.ActivatedAbilitiesThisTurn == nil {
		g.ActivatedAbilitiesThisTurn = make(map[game.ActivatedAbilityUse]bool)
	}
	g.ActivatedAbilitiesThisTurn[game.ActivatedAbilityUse{SourceID: sourceID, AbilityIndex: abilityIndex}] = true
}

func abilityHasOncePerTurnRestriction(timing game.TimingRestriction) bool {
	return timing == game.OncePerTurn || timing == game.SorceryOncePerTurn
}

func isSorcerySpeed(g *game.Game, playerID game.PlayerID) bool {
	return playerID == g.Turn.ActivePlayer &&
		g.Turn.IsMainPhase() &&
		g.Turn.Step == game.StepNone &&
		g.Stack.IsEmpty()
}

func landCardInstance(g *game.Game, player *game.Player, cardID id.ID) (*game.CardInstance, bool) {
	return landCardInstanceFace(g, player, cardID, game.FaceFront)
}

func landCardInstanceFace(g *game.Game, player *game.Player, cardID id.ID, face game.FaceIndex) (*game.CardInstance, bool) {
	return landCardInstanceFaceFromZone(g, player, cardID, zone.Hand, face)
}

func landCardInstanceFaceFromZone(g *game.Game, player *game.Player, cardID id.ID, sourceZone zone.Type, face game.FaceIndex) (*game.CardInstance, bool) {
	cards, ok := playerCardsInZone(player, sourceZone)
	if !ok || !cards.Contains(cardID) {
		return nil, false
	}
	card, ok := g.GetCardInstance(cardID)
	if !ok || !card.Def.CanChooseLandFace(face) {
		return nil, false
	}
	return card, true
}

func entersSummoningSick(card *game.CardDef) bool {
	return !card.HasKeyword(game.Haste)
}

func isSupportedSpell(card *game.CardDef) bool {
	return !card.HasType(types.Land) &&
		(card.IsPermanent() ||
			card.HasType(types.Instant) ||
			card.HasType(types.Sorcery))
}
