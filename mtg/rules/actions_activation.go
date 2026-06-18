package rules

import (
	"slices"
	"strings"

	"github.com/natefinch/council4/mtg/game/zone"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/rules/payment"
)

func effectiveCyclingCost(g *game.Game, playerID game.PlayerID, card *game.CardInstance, body *game.ActivatedAbility) *cost.Mana {
	if body == nil {
		return nil
	}
	return applyAbilityCostModifiers(manaCostPtr(body.ManaCost), cyclingCostModifiers(g, playerID, card, body))
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
	if modifier.MatchCardType {
		if card == nil || card.Def == nil || !card.Def.HasType(modifier.CardType) {
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
		g.Turn.CanPlayLand()
}

func canCastAtCurrentTiming(g *game.Game, playerID game.PlayerID, card *game.CardDef) bool {
	if card.HasType(types.Instant) || card.HasKeyword(game.Flash) {
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
	var values []int
	for x := 0; x <= maxLegalXValue; x++ {
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
	body, ok := frontDef.BodyAt(abilityIndex).(game.ActivatedAbility)
	if !ok {
		return nil, game.ActivatedAbility{}, false
	}
	return card, body, true
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
	if !game.BodyHasKeyword(body, game.Equip) && body.Timing != game.SorceryOnly {
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
	return paymentOrch.canPayGenericCost(g, payment.GenericRequest{PlayerID: playerID, Cost: manaCostPtr(body.ManaCost)})
}

func canActivateGeneralAbility(g *game.Game, playerID game.PlayerID, permanent *game.Permanent, body *game.ActivatedAbility, abilityIndex int, targets []game.Target, xValue int) bool {
	return canActivateGeneralAbilityWithModes(g, playerID, permanent, body, abilityIndex, targets, xValue, nil)
}

func canActivateGeneralAbilityWithModes(g *game.Game, playerID game.PlayerID, permanent *game.Permanent, body *game.ActivatedAbility, abilityIndex int, targets []game.Target, xValue int, chosenModes []int) bool {
	if body == nil || !canAct(g, playerID) || playerID != g.Turn.PriorityPlayer || permanent.PhasedOut || effectiveController(g, permanent) != playerID {
		return false
	}
	if game.BodyHasKeyword(body, game.Equip) || !bodyFunctionsOnBattlefield(body) {
		return false
	}
	if !activatedAbilityTimingAllows(g, playerID, body.Timing) || activatedAbilityUsedThisTurn(g, permanent.ObjectID, abilityIndex, body.Timing) {
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
	return paymentOrch.buildAbilityCostPlan(g, payment.AbilityRequest{
		PlayerID:         playerID,
		Source:           permanent,
		ManaCost:         body.ManaCost,
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
	if !ok || !game.BodyHasKeyword(gotAbility, game.Cycling) {
		return false
	}
	return paymentOrch.canPayGenericCost(g, payment.GenericRequest{PlayerID: playerID, Cost: effectiveCyclingCost(g, playerID, card, body)})
}

func canActivateGraveyardAbility(g *game.Game, playerID game.PlayerID, cardID id.ID, body *game.ActivatedAbility, abilityIndex int, targets []game.Target, xValue int) bool {
	return canActivateGraveyardAbilityWithModes(g, playerID, cardID, body, abilityIndex, targets, xValue, nil)
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
	if len(game.BodyTargets(body)) != 0 || !manaBodyHasAddManaEffect(body) || !manaBodyChoicesAvailable(g, playerID, body) {
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
			default:
				return false
			}
		}
		return hasAddMana
	}
	return false
}

func manaBodyChoicesAvailable(g *game.Game, playerID game.PlayerID, body *game.ManaAbility) bool {
	if body == nil {
		return false
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
			_, values := resolutionChoiceOptions(g, nil, choicePlayer, &primitive.Choice)
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
		if !ok || addMana.EntryChoiceFrom == "" {
			continue
		}
		if !permanentEntryChoiceAvailable(permanent, addMana.EntryChoiceFrom) {
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
	if !player.Hand.Contains(cardID) {
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
