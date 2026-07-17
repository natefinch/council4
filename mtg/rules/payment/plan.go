package payment

import (
	"maps"

	"github.com/natefinch/council4/mtg/game/zone"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// paymentPlan describes the concrete steps needed to pay a mana cost.
type paymentPlan struct {
	poolSpend      map[mana.Unit]int
	manaTaps       []manaTap
	convokeTaps    []*game.Permanent
	improviseTaps  []*game.Permanent
	delveExiles    []id.ID
	lifePayment    int
	symbolPayments []game.SymbolPayment
}

// spellCostPlan is the full payment plan for a spell, including mana and
// additional costs.
type spellCostPlan struct {
	mana       paymentPlan
	additional additionalCostPlan
	option     spellCostOption
}

// abilityCostPlan is the full payment plan for an activated ability.
type abilityCostPlan struct {
	mana       paymentPlan
	additional additionalCostPlan
	tapSource  bool
}

// manaTap records a planned tap of a mana-producing permanent.
type manaTap struct {
	permanent    *game.Permanent
	color        mana.Color
	amount       int
	snow         bool
	fromCreature bool
	untap        bool
	sacrifice    bool
	abilityIndex int
	timing       game.TimingRestriction
	flexibility  int
}

// manaSource is a candidate mana-producing permanent used during plan building.
type manaSource struct {
	permanent    *game.Permanent
	color        mana.Color
	amount       int
	snow         bool
	fromCreature bool
	untap        bool
	sacrifice    bool
	abilityIndex int
	timing       game.TimingRestriction
	flexibility  int
}

type manaSourceRestrictions struct {
	excluded                 map[id.ID]bool
	nonSacrificingOnly       map[id.ID]bool
	reservedGraveyardCardIDs map[id.ID]bool
}

// paymentColors is the deterministic ordering used when spending mana. Callers
// must consume mana sources through this slice rather than ranging over maps.
var paymentColors = []mana.Color{
	mana.W,
	mana.U,
	mana.B,
	mana.R,
	mana.G,
	mana.C,
}

func canPayCostWithX(s State, playerID game.PlayerID, manaCost *cost.Mana, xValue int) bool {
	_, ok := buildPaymentPlan(s, playerID, manaCost, xValue, nil)
	return ok
}

func canPaySpellCosts(s State, req SpellRequest) bool {
	for _, option := range spellCostOptionsForRequest(s, req) {
		if _, ok := buildSpellCostPlanForOption(s, req.PlayerID, req.CardID, req.SourceZone, option, req.XValue, req.Targets, req.Bestowed, nil); ok {
			return true
		}
	}
	return false
}

func paySpellCosts(s State, req SpellRequest) (SpellPaymentResult, bool) {
	plan, ok := buildSpellCostPlan(s, req)
	if !ok {
		return SpellPaymentResult{}, false
	}
	player, ok := s.Player(req.PlayerID)
	if !ok || !paymentApplicationReady(s, player, plan.mana, plan.additional) {
		return SpellPaymentResult{}, false
	}
	applyPaymentPlan(s, req.PlayerID, plan.mana)
	applyAdditionalCostPlan(s, plan.additional)
	return SpellPaymentResult{
		AdditionalCostsPaid: plan.additional.paid,
		SacrificedIDs:       sacrificedPermanentIDs(plan.additional),
		PoolSpend:           clonePoolSpend(plan.mana.poolSpend),
		CastPermission:      plan.option.castPermission,
	}, true
}

// clonePoolSpend returns an independent copy of a plan's per-unit pool spend,
// reporting exactly how much pool mana of each exact unit (color and snow
// provenance) the plan consumed. The rules engine uses it to resolve mana-spend
// riders against the precise units spent rather than inferring spend from gross
// pool deltas that mid-payment mana production could mask. Entries with a
// non-positive amount are dropped. It returns nil for an empty spend so callers
// holding no riders allocate nothing.
func clonePoolSpend(poolSpend map[mana.Unit]int) map[mana.Unit]int {
	if len(poolSpend) == 0 {
		return nil
	}
	cloned := make(map[mana.Unit]int, len(poolSpend))
	for unit, amount := range poolSpend {
		if amount > 0 {
			cloned[unit] = amount
		}
	}
	if len(cloned) == 0 {
		return nil
	}
	return cloned
}

func buildSpellCostPlan(s State, req SpellRequest) (spellCostPlan, bool) {
	options := spellCostOptionsForRequest(s, req)
	if len(options) == 0 {
		return spellCostPlan{}, false
	}
	if req.Prefs != nil {
		for _, option := range options {
			if option.index == req.Prefs.AlternativeIndex {
				return buildSpellCostPlanForOption(s, req.PlayerID, req.CardID, req.SourceZone, option, req.XValue, req.Targets, req.Bestowed, req.Prefs)
			}
		}
		return spellCostPlan{}, false
	}
	for _, option := range options {
		if plan, ok := buildSpellCostPlanForOption(s, req.PlayerID, req.CardID, req.SourceZone, option, req.XValue, req.Targets, req.Bestowed, nil); ok {
			return plan, true
		}
	}
	return spellCostPlan{}, false
}

func buildAbilityCostPlan(s State, req AbilityRequest) (abilityCostPlan, bool) {
	plan := abilityCostPlan{}
	if req.Source == nil && req.SourceCardID == 0 {
		return plan, false
	}
	if req.XValue != 0 && !costHasVariableMana(manaCostPtr(req.ManaCost)) && !additionalCostsUseX(req.AdditionalCosts) {
		return plan, false
	}
	tapSource := hasTapCostOf(req.AdditionalCosts)
	if tapSource && (req.Source == nil || !canTapForAbility(s, req.Source)) {
		return plan, false
	}
	sourceCardID := req.SourceCardID
	sourceZone := req.SourceZone
	if req.Source != nil && sourceCardID == 0 {
		sourceCardID = req.Source.CardInstanceID
		sourceZone = zone.Battlefield
	}
	additional, ok := buildAdditionalCostPlanForCosts(s, req.PlayerID, req.AdditionalCosts, req.XValue, clonePreferences(req.Prefs), req.Source, sourceCardID, sourceZone, 0)
	if !ok {
		return plan, false
	}
	manaPlan, ok := buildPaymentPlanWithPreferences(s, req.PlayerID, manaCostPtr(req.ManaCost), req.XValue, abilityManaRestrictions(additional, tapSource, req.Source, true), spendContext{abilitySource: req.Source}, clonePreferences(req.Prefs))
	if !ok {
		additional, manaPlan, ok = retryAbilityCostPlanAvoidingManaTapConflict(s, req, sourceCardID, sourceZone, tapSource, additional)
		if !ok {
			return plan, false
		}
	}
	plan.mana = manaPlan
	plan.additional = additional
	plan.tapSource = tapSource
	return plan, true
}

func retryAbilityCostPlanAvoidingManaTapConflict(s State, req AbilityRequest, sourceCardID id.ID, sourceZone zone.Type, tapSource bool, previous additionalCostPlan) (additionalCostPlan, paymentPlan, bool) {
	if len(previous.permanentsToTap) == 0 {
		return additionalCostPlan{}, paymentPlan{}, false
	}
	manaPlan, ok := buildPaymentPlanWithPreferences(s, req.PlayerID, manaCostPtr(req.ManaCost), req.XValue, abilityManaRestrictions(previous, tapSource, req.Source, false), spendContext{abilitySource: req.Source}, clonePreferences(req.Prefs))
	if !ok {
		return additionalCostPlan{}, paymentPlan{}, false
	}
	additional, ok := buildAdditionalCostPlanForCosts(s, req.PlayerID, req.AdditionalCosts, req.XValue, tapRetryPreferences(req.Prefs), req.Source, sourceCardID, sourceZone, 0, paymentPlanTappedPermanents(manaPlan)...)
	if !ok {
		return additionalCostPlan{}, paymentPlan{}, false
	}
	manaPlan, ok = buildPaymentPlanWithPreferences(s, req.PlayerID, manaCostPtr(req.ManaCost), req.XValue, abilityManaRestrictions(additional, tapSource, req.Source, true), spendContext{abilitySource: req.Source}, clonePreferences(req.Prefs))
	if !ok {
		return additionalCostPlan{}, paymentPlan{}, false
	}
	return additional, manaPlan, true
}

func additionalCostsUseX(costs []cost.Additional) bool {
	for _, additional := range costs {
		if additional.AmountFromX {
			return true
		}
	}
	return false
}

func abilityManaRestrictions(additional additionalCostPlan, tapSource bool, source *game.Permanent, includeTapPermanents bool) manaSourceRestrictions {
	restrictions := additionalManaRestrictions(nil, additional, includeTapPermanents)
	if tapSource && source != nil {
		restrictions.excluded[source.ObjectID] = true
	}
	return restrictions
}

func additionalManaRestrictions(base map[id.ID]bool, additional additionalCostPlan, includeTapPermanents bool) manaSourceRestrictions {
	restrictions := manaSourceRestrictions{
		excluded:                 make(map[id.ID]bool),
		nonSacrificingOnly:       make(map[id.ID]bool),
		reservedGraveyardCardIDs: make(map[id.ID]bool),
	}
	maps.Copy(restrictions.excluded, base)
	for _, permanent := range additional.manaExcluded {
		restrictions.excluded[permanent.ObjectID] = true
	}
	for _, sacrifice := range additional.sacrifices {
		restrictions.nonSacrificingOnly[sacrifice.ObjectID] = true
	}
	for _, permanent := range additional.exilePermanents {
		restrictions.nonSacrificingOnly[permanent.ObjectID] = true
	}
	for _, returned := range additional.returnsToHand {
		restrictions.nonSacrificingOnly[returned.permanent.ObjectID] = true
	}
	for _, exile := range additional.exiles {
		if exile.zone == zone.Graveyard {
			restrictions.reservedGraveyardCardIDs[exile.cardID] = true
		}
	}
	for _, evidence := range additional.evidence {
		for _, card := range evidence.cards {
			restrictions.reservedGraveyardCardIDs[card.cardID] = true
		}
	}
	if includeTapPermanents {
		for _, permanent := range additional.permanentsToTap {
			restrictions.excluded[permanent.ObjectID] = true
		}
	}
	return restrictions
}

func tapRetryPreferences(prefs *Preferences) *Preferences {
	cloned := clonePreferences(prefs)
	if cloned != nil {
		cloned.TapChoices = nil
	}
	return cloned
}

func paymentPlanTappedPermanents(plan paymentPlan) []*game.Permanent {
	permanents := make([]*game.Permanent, 0, len(plan.manaTaps)+len(plan.convokeTaps)+len(plan.improviseTaps))
	for _, tap := range plan.manaTaps {
		permanents = append(permanents, tap.permanent)
	}
	permanents = append(permanents, plan.convokeTaps...)
	permanents = append(permanents, plan.improviseTaps...)
	return permanents
}

// AbilityCostPayment carries the results of paying an ability's activation cost:
// the per-unit pool mana consumed (for mana-spend rider resolution), the object
// IDs of permanents sacrificed as a cost, and the card-instance IDs of cards
// exiled as a cost so the caller can record them on the resolving stack object.
type AbilityCostPayment struct {
	PoolSpend     map[mana.Unit]int
	SacrificedIDs []id.ID
	TappedIDs     []id.ID
	ExiledIDs     []id.ID
}

func payAbilityCosts(s State, req AbilityRequest) (AbilityCostPayment, bool) {
	plan, ok := buildAbilityCostPlan(s, req)
	if !ok {
		return AbilityCostPayment{}, false
	}
	player, ok := s.Player(req.PlayerID)
	if !ok || !abilityCostPlanStillValid(s, player, req.Source, plan) {
		return AbilityCostPayment{}, false
	}
	applyPaymentPlan(s, req.PlayerID, plan.mana)
	if plan.tapSource && !tapForAbility(s, req.Source, req.ForMana) {
		panic("ability source became untappable after prevalidation")
	}
	applyAdditionalCostPlan(s, plan.additional)
	return AbilityCostPayment{
		PoolSpend:     clonePoolSpend(plan.mana.poolSpend),
		SacrificedIDs: sacrificedPermanentIDs(plan.additional),
		TappedIDs:     tappedPermanentIDs(plan.additional),
		ExiledIDs:     exiledCardIDs(plan.additional),
	}, true
}

// exiledCardIDs returns the card-instance IDs of cards exiled by the
// additional-cost plan, in plan order, so a resolution effect can act on the
// cost-exiled cards ("An opponent chooses one of the exiled cards ...").
func exiledCardIDs(plan additionalCostPlan) []id.ID {
	if len(plan.exiles) == 0 {
		return nil
	}
	ids := make([]id.ID, 0, len(plan.exiles))
	for _, exiled := range plan.exiles {
		ids = append(ids, exiled.cardID)
	}
	return ids
}

// sacrificedPermanentIDs returns the object IDs of permanents sacrificed by the
// additional-cost plan, in plan order, so an effect can read the sacrificed
// permanent's last-known information ("the sacrificed creature's power").
func sacrificedPermanentIDs(plan additionalCostPlan) []id.ID {
	if len(plan.sacrifices) == 0 {
		return nil
	}
	ids := make([]id.ID, 0, len(plan.sacrifices))
	for _, sacrifice := range plan.sacrifices {
		ids = append(ids, sacrifice.ObjectID)
	}
	return ids
}

// tappedPermanentIDs returns the object IDs of permanents tapped to pay
// additional costs, in plan order. It deliberately excludes mana-source,
// convoke, and improvise taps, which are not AdditionalTapPermanents costs.
func tappedPermanentIDs(plan additionalCostPlan) []id.ID {
	if len(plan.permanentsToTap) == 0 {
		return nil
	}
	ids := make([]id.ID, 0, len(plan.permanentsToTap))
	for _, permanent := range plan.permanentsToTap {
		ids = append(ids, permanent.ObjectID)
	}
	return ids
}

func canPayGenericCost(s State, req GenericRequest) bool {
	if len(req.AdditionalCosts) > 0 {
		if _, ok := buildGenericCostPlan(s, req); !ok {
			return false
		}
		return true
	}
	if len(req.Exclude) > 0 {
		_, ok := buildPaymentPlan(s, req.PlayerID, req.Cost, req.XValue, req.Exclude)
		return ok
	}
	return canPayCostWithX(s, req.PlayerID, req.Cost, req.XValue)
}

func payGenericCost(s State, req GenericRequest) (poolSpend map[mana.Unit]int, ok bool) {
	if len(req.AdditionalCosts) > 0 {
		plan, ok := buildGenericCostPlan(s, req)
		if !ok {
			return nil, false
		}
		player, ok := s.Player(req.PlayerID)
		if !ok || !paymentApplicationReady(s, player, plan.mana, plan.additional) {
			return nil, false
		}
		applyPaymentPlan(s, req.PlayerID, plan.mana)
		applyAdditionalCostPlan(s, plan.additional)
		return clonePoolSpend(plan.mana.poolSpend), true
	}
	plan, ok := buildPaymentPlanWithPreferences(s, req.PlayerID, req.Cost, req.XValue, manaSourceRestrictions{excluded: req.Exclude}, spendContext{spell: req.Spell}, req.Prefs)
	if !ok {
		return nil, false
	}
	player, ok := s.Player(req.PlayerID)
	if !ok || !paymentPlanStillValid(s, player, plan) {
		return nil, false
	}
	applyPaymentPlan(s, req.PlayerID, plan)
	return clonePoolSpend(plan.poolSpend), true
}

func buildGenericCostPlan(s State, req GenericRequest) (spellCostPlan, bool) {
	plan := spellCostPlan{}
	additional, ok := buildAdditionalCostPlanForCosts(s, req.PlayerID, req.AdditionalCosts, req.XValue, clonePreferences(req.Prefs), req.Source, req.SourceCardID, zone.None, 0)
	if !ok {
		return plan, false
	}
	manaPlan, ok := buildPaymentPlanWithPreferences(s, req.PlayerID, req.Cost, req.XValue, additionalManaRestrictions(req.Exclude, additional, true), spendContext{spell: req.Spell}, clonePreferences(req.Prefs))
	if !ok {
		additional, manaPlan, ok = retryGenericCostPlanAvoidingManaTapConflict(s, req, additional)
		if !ok {
			return plan, false
		}
	}
	plan.additional = additional
	plan.mana = manaPlan
	return plan, true
}

func retryGenericCostPlanAvoidingManaTapConflict(s State, req GenericRequest, previous additionalCostPlan) (additionalCostPlan, paymentPlan, bool) {
	if len(previous.permanentsToTap) == 0 {
		return additionalCostPlan{}, paymentPlan{}, false
	}
	manaPlan, ok := buildPaymentPlanWithPreferences(s, req.PlayerID, req.Cost, req.XValue, additionalManaRestrictions(req.Exclude, previous, false), spendContext{spell: req.Spell}, clonePreferences(req.Prefs))
	if !ok {
		return additionalCostPlan{}, paymentPlan{}, false
	}
	additional, ok := buildAdditionalCostPlanForCosts(s, req.PlayerID, req.AdditionalCosts, req.XValue, tapRetryPreferences(req.Prefs), req.Source, req.SourceCardID, zone.None, 0, paymentPlanTappedPermanents(manaPlan)...)
	if !ok {
		return additionalCostPlan{}, paymentPlan{}, false
	}
	manaPlan, ok = buildPaymentPlanWithPreferences(s, req.PlayerID, req.Cost, req.XValue, additionalManaRestrictions(req.Exclude, additional, true), spendContext{spell: req.Spell}, clonePreferences(req.Prefs))
	if !ok {
		return additionalCostPlan{}, paymentPlan{}, false
	}
	return additional, manaPlan, true
}

func buildSpellCostPlanForOption(s State, playerID game.PlayerID, cardID id.ID, sourceZone zone.Type, option spellCostOption, xValue int, targets []game.Target, bestowed bool, prefs *Preferences) (spellCostPlan, bool) {
	option = applyCostModifiers(s, costModificationContext{player: playerID, card: option.card, cardID: cardID, sourceZone: sourceZone, targets: targets, bargained: option.bargained, bestowed: bestowed, option: option})
	plan := spellCostPlan{option: option}
	if xValue < 0 ||
		xValue != 0 && !costHasVariableMana(option.manaCost) && !additionalCostsUseX(option.additionalCosts) {
		return plan, false
	}
	additional, ok := buildAdditionalCostPlanForCosts(s, playerID, option.additionalCosts, xValue, clonePreferences(prefs), nil, cardID, sourceZone, cardID)
	if !ok {
		return plan, false
	}
	manaPlan, ok := buildSpellManaPlanForOption(s, playerID, cardID, sourceZone, option, xValue, additionalManaRestrictions(nil, additional, true), clonePreferences(prefs))
	if !ok {
		additional, manaPlan, ok = retrySpellCostPlanAvoidingManaTapConflict(s, playerID, cardID, sourceZone, option, xValue, prefs, additional)
		if !ok {
			return plan, false
		}
	}
	plan.additional = additional
	plan.mana = manaPlan
	return plan, true
}

func retrySpellCostPlanAvoidingManaTapConflict(s State, playerID game.PlayerID, cardID id.ID, sourceZone zone.Type, option spellCostOption, xValue int, prefs *Preferences, previous additionalCostPlan) (additionalCostPlan, paymentPlan, bool) {
	if len(previous.permanentsToTap) == 0 {
		return additionalCostPlan{}, paymentPlan{}, false
	}
	manaPlan, ok := buildSpellManaPlanForOption(s, playerID, cardID, sourceZone, option, xValue, additionalManaRestrictions(nil, previous, false), clonePreferences(prefs))
	if !ok {
		return additionalCostPlan{}, paymentPlan{}, false
	}
	additional, ok := buildAdditionalCostPlanForCosts(s, playerID, option.additionalCosts, xValue, tapRetryPreferences(prefs), nil, cardID, sourceZone, cardID, paymentPlanTappedPermanents(manaPlan)...)
	if !ok {
		return additionalCostPlan{}, paymentPlan{}, false
	}
	manaPlan, ok = buildSpellManaPlanForOption(s, playerID, cardID, sourceZone, option, xValue, additionalManaRestrictions(nil, additional, true), clonePreferences(prefs))
	if !ok {
		return additionalCostPlan{}, paymentPlan{}, false
	}
	return additional, manaPlan, true
}

func buildSpellManaPlanForOption(s State, playerID game.PlayerID, cardID id.ID, sourceZone zone.Type, option spellCostOption, xValue int, restrictions manaSourceRestrictions, prefs *Preferences) (paymentPlan, bool) {
	manaPlan, ok := buildPaymentPlanWithPreferences(s, playerID, option.manaCost, xValue, restrictions, spendContext{spell: option.card}, prefs)
	if ok {
		return manaPlan, true
	}
	convokeTaps, convokedCost, convokeOK := convokePayment(s, playerID, option.manaCost, xValue, restrictions.excluded)
	if spellHasCostKeyword(s, playerID, option.card, cardID, sourceZone, game.Convoke) && convokeOK {
		convokeRestrictions := cloneManaSourceRestrictions(restrictions)
		for _, permanent := range convokeTaps {
			convokeRestrictions.excluded[permanent.ObjectID] = true
		}
		manaPlan, ok = buildPaymentPlanWithPreferences(s, playerID, convokedCost, xValue, convokeRestrictions, spendContext{spell: option.card}, prefs)
		if ok {
			manaPlan.convokeTaps = convokeTaps
			return manaPlan, true
		}
	}
	if spellHasCostKeyword(s, playerID, option.card, cardID, sourceZone, game.Improvise) {
		improviseTaps, improvisedCost, improviseOK := improvisePayment(s, playerID, option.manaCost, xValue, restrictions.excluded)
		if improviseOK {
			improviseRestrictions := cloneManaSourceRestrictions(restrictions)
			for _, permanent := range improviseTaps {
				improviseRestrictions.excluded[permanent.ObjectID] = true
			}
			manaPlan, ok = buildPaymentPlanWithPreferences(s, playerID, improvisedCost, xValue, improviseRestrictions, spendContext{spell: option.card}, prefs)
			if ok {
				manaPlan.improviseTaps = improviseTaps
				return manaPlan, true
			}
		}
	}
	if spellHasCostKeyword(s, playerID, option.card, cardID, sourceZone, game.Delve) {
		delveExiles, generic, delveOK := delveCandidates(s, playerID, option.manaCost, xValue, cardID, sourceZone, restrictions.reservedGraveyardCardIDs)
		for exiledCount := 1; delveOK && exiledCount <= min(generic, len(delveExiles)); exiledCount++ {
			delvedCost := costWithGenericRequirement(option.manaCost, generic-exiledCount)
			manaPlan, ok = buildPaymentPlanWithPreferences(s, playerID, delvedCost, 0, restrictions, spendContext{spell: option.card}, prefs)
			if ok {
				manaPlan.delveExiles = append([]id.ID(nil), delveExiles[:exiledCount]...)
				return manaPlan, true
			}
		}
	}
	return paymentPlan{}, false
}

// spellHasCostKeyword reports whether the spell being cast carries keyword,
// either natively on its card face or via an active RuleEffectGrantSpellKeyword
// that grants it to spells the caster casts ("Nonartifact spells you cast have
// improvise.", Inspiring Statuary; "The next spell you cast this turn has
// improvise.", Archway of Innovation). Native and granted keywords compose
// idempotently: a card that already has the keyword is unaffected by a grant, so
// the improvise/convoke/delve payment path is enabled exactly once. It centralizes
// the keyword check so any granted cost-affecting keyword is honored before costs
// are paid.
func spellHasCostKeyword(s State, playerID game.PlayerID, card *game.CardDef, cardID id.ID, sourceZone zone.Type, keyword game.Keyword) bool {
	if card.HasKeyword(keyword) {
		return true
	}
	return s.SpellHasGrantedKeyword(playerID, card, cardID, sourceZone, keyword)
}

func buildPaymentPlan(s State, playerID game.PlayerID, manaCost *cost.Mana, xValue int, exclude map[id.ID]bool) (paymentPlan, bool) {
	return buildPaymentPlanWithPreferences(s, playerID, manaCost, xValue, manaSourceRestrictions{excluded: exclude}, spendContext{}, nil)
}

// effectiveManaSymbols rewrites the cost's colored symbols into Phyrexian symbols
// of the same color when an active RuleEffectPayLifeForColoredMana lets playerID
// pay 2 life rather than that mana ("For each {B} in a cost, you may pay 2 life
// rather than pay that mana.", K'rrik). It returns manaCost unchanged when no
// symbol is affected, so unaffected costs allocate nothing and pay exactly as
// before.
func effectiveManaSymbols(s State, playerID game.PlayerID, manaCost cost.Mana) []cost.Symbol {
	converted := false
	for i := range manaCost {
		if manaCost[i].Kind == cost.ColoredSymbol && s.PayLifeForManaColor(playerID, manaCost[i].Color) {
			converted = true
			break
		}
	}
	if !converted {
		return manaCost
	}
	symbols := make([]cost.Symbol, len(manaCost))
	for i := range manaCost {
		symbol := manaCost[i]
		if symbol.Kind == cost.ColoredSymbol && s.PayLifeForManaColor(playerID, symbol.Color) {
			symbol = cost.PhyrexianMana(symbol.Color)
		}
		symbols[i] = symbol
	}
	return symbols
}

// EffectiveManaCost returns manaCost with each colored symbol the player may pay
// life for instead rewritten to the equivalent Phyrexian symbol (CR for K'rrik's
// "For each {B} in a cost, ..." static). Callers that enumerate Phyrexian payment
// choices use it so the choice order matches the payment plan's symbol order.
func EffectiveManaCost(s State, playerID game.PlayerID, manaCost cost.Mana) cost.Mana {
	return effectiveManaSymbols(s, playerID, manaCost)
}

func buildPaymentPlanWithPreferences(s State, playerID game.PlayerID, manaCost *cost.Mana, xValue int, restrictions manaSourceRestrictions, ctx spendContext, prefs *Preferences) (paymentPlan, bool) {
	plan := paymentPlan{poolSpend: make(map[mana.Unit]int)}
	player, ok := s.Player(playerID)
	if !ok {
		return plan, false
	}
	pool := snapshotPool(s, player, ctx)
	manaSources := availableManaSources(s, playerID, restrictions)
	if xValue < 0 {
		return plan, false
	}
	if manaCost == nil {
		return plan, true
	}

	symbols := effectiveManaSymbols(s, playerID, *manaCost)

	for _, symbol := range symbols {
		switch symbol.Kind {
		case cost.ColoredSymbol:
			if !payColoredSymbol(&plan, pool, manaSources, symbol, symbol.Color, game.SymbolPaymentMana) {
				return plan, false
			}
		case cost.ColorlessSymbol:
			if !payColoredSymbol(&plan, pool, manaSources, symbol, mana.C, game.SymbolPaymentMana) {
				return plan, false
			}
		default:
		}
	}
	for _, symbol := range symbols {
		if symbol.Kind == cost.SnowSymbol {
			if !paySnowSymbol(&plan, pool, manaSources, symbol) {
				return plan, false
			}
		}
	}
	for _, symbol := range symbols {
		switch symbol.Kind {
		case cost.HybridSymbol:
			if !payHybridSymbol(&plan, pool, manaSources, symbol) {
				return plan, false
			}
		case cost.TwobridSymbol:
			if !payMonoHybridSymbol(&plan, pool, manaSources, symbol) {
				return plan, false
			}
		case cost.PhyrexianSymbol:
			if !payPhyrexianSymbol(player, &plan, pool, manaSources, symbol, prefs, s.CanPayLife(playerID)) {
				return plan, false
			}
		case cost.PhyrexianGenericSymbol:
			if !payPhyrexianGenericSymbol(player, &plan, pool, manaSources, symbol, prefs, s.CanPayLife(playerID)) {
				return plan, false
			}
		default:
		}
	}
	for _, symbol := range symbols {
		switch symbol.Kind {
		case cost.GenericSymbol:
			if !payGenericSymbol(&plan, pool, manaSources, symbol, symbol.Generic, game.SymbolPaymentGeneric) {
				return plan, false
			}
		case cost.VariableSymbol:
			if !payGenericSymbol(&plan, pool, manaSources, symbol, xValue, game.SymbolPaymentX) {
				return plan, false
			}
		default:
			if symbol.Kind != cost.ColoredSymbol &&
				symbol.Kind != cost.ColorlessSymbol &&
				symbol.Kind != cost.SnowSymbol &&
				symbol.Kind != cost.HybridSymbol &&
				symbol.Kind != cost.TwobridSymbol &&
				symbol.Kind != cost.PhyrexianSymbol &&
				symbol.Kind != cost.PhyrexianGenericSymbol {
				return plan, false
			}
		}
	}
	return plan, true
}

func paymentPlanStillValid(s State, player *game.Player, plan paymentPlan) bool {
	tappedMana := make(map[mana.Unit]int)
	for _, tap := range plan.manaTaps {
		current, ok := s.PermanentByObjectID(tap.permanent.ObjectID)
		if !ok || current != tap.permanent ||
			tap.permanent.Tapped != tap.untap ||
			s.EffectiveController(tap.permanent) != player.ID {
			return false
		}
		output, ok := permanentManaOutputForActivation(s, tap.permanent, tap)
		if !ok ||
			output.color != tap.color ||
			output.amount != tap.amount ||
			output.snow != tap.snow ||
			output.fromCreature != tap.fromCreature ||
			output.untap != tap.untap ||
			output.sacrifice != tap.sacrifice ||
			output.abilityIndex != tap.abilityIndex ||
			output.timing != tap.timing {
			return false
		}
		tappedMana[mana.Unit{Color: tap.color, Snow: tap.snow, FromCreature: tap.fromCreature}] += tap.amount
	}
	for _, permanent := range plan.convokeTaps {
		if !canConvokeWith(s, player.ID, permanent, nil) {
			return false
		}
	}
	for _, permanent := range plan.improviseTaps {
		if !canImproviseWith(s, player.ID, permanent, nil) {
			return false
		}
	}
	for _, cardID := range plan.delveExiles {
		if !player.Graveyard.Contains(cardID) {
			return false
		}
	}
	for _, unit := range paymentUnitOrder() {
		if player.ManaPool.Units()[unit]+tappedMana[unit] < plan.poolSpend[unit] {
			return false
		}
	}
	return player.Life >= plan.lifePayment && (plan.lifePayment == 0 || s.CanPayLife(player.ID))
}

// paymentApplicationReady reports whether a fully built mana + additional cost
// plan can be applied without any post-mutation failure. It re-checks every
// resource the appliers consume and, crucially, the player's combined life cost
// across both plans (a Phyrexian mana symbol and an additional "pay N life" cost
// can each draw from the same life total, yet each plan's own validity check
// only sees its own share). Establishing readiness before any mutation lets
// applyPaymentPlan and applyAdditionalCostPlan treat the plan as a validated
// contract and panic on any inconsistency, so a clean failure here leaves game
// state byte-for-byte unchanged.
func paymentApplicationReady(s State, player *game.Player, manaPlan paymentPlan, additionalPlan additionalCostPlan) bool {
	if !additionalCostPlanStillValid(s, player, additionalPlan) || !paymentPlanStillValid(s, player, manaPlan) {
		return false
	}
	totalLife := manaPlan.lifePayment + additionalPlan.lifePaid
	return totalLife == 0 || (player.Life >= totalLife && s.CanPayLife(player.ID))
}

func abilityCostPlanStillValid(s State, player *game.Player, source *game.Permanent, plan abilityCostPlan) bool {
	if plan.tapSource && !canTapForAbility(s, source) {
		return false
	}
	return paymentApplicationReady(s, player, plan.mana, plan.additional)
}

func clonePaymentPlan(plan paymentPlan) paymentPlan {
	plan.poolSpend = cloneUnitCounts(plan.poolSpend)
	plan.manaTaps = append([]manaTap(nil), plan.manaTaps...)
	plan.symbolPayments = append([]game.SymbolPayment(nil), plan.symbolPayments...)
	return plan
}

func cloneUnitCounts(units map[mana.Unit]int) map[mana.Unit]int {
	clone := make(map[mana.Unit]int, len(units))
	maps.Copy(clone, units)
	return clone
}

func cloneManaSourceRestrictions(restrictions manaSourceRestrictions) manaSourceRestrictions {
	clone := manaSourceRestrictions{
		excluded:                 make(map[id.ID]bool, len(restrictions.excluded)),
		nonSacrificingOnly:       make(map[id.ID]bool, len(restrictions.nonSacrificingOnly)),
		reservedGraveyardCardIDs: make(map[id.ID]bool, len(restrictions.reservedGraveyardCardIDs)),
	}
	maps.Copy(clone.excluded, restrictions.excluded)
	maps.Copy(clone.nonSacrificingOnly, restrictions.nonSacrificingOnly)
	maps.Copy(clone.reservedGraveyardCardIDs, restrictions.reservedGraveyardCardIDs)
	return clone
}

func replaceUnitCounts(dst, src map[mana.Unit]int) {
	for unit := range dst {
		delete(dst, unit)
	}
	maps.Copy(dst, src)
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

func costRequirements(manaCost *cost.Mana, xValue int) (colored map[mana.Color]int, generic int, ok bool) {
	colored = make(map[mana.Color]int)
	if xValue < 0 {
		return nil, 0, false
	}
	if manaCost == nil {
		return colored, 0, true
	}

	generic = 0
	for _, symbol := range *manaCost {
		switch symbol.Kind {
		case cost.ColoredSymbol:
			colored[symbol.Color]++
		case cost.ColorlessSymbol:
			colored[mana.C]++
		case cost.GenericSymbol:
			generic += symbol.Generic
		case cost.VariableSymbol:
			generic += xValue
		default:
			return nil, 0, false
		}
	}
	return colored, generic, true
}

// spendContext identifies what a mana payment is for, so a restricted
// mana-spend rider can decide whether its tagged mana may be admitted to the
// pool for this payment. A spell is set when casting a spell; abilitySource is
// set when paying an activated ability's costs (the ability's source permanent).
// A zero spendContext is an unidentified payment that no restricted rider mana
// may pay.
type spendContext struct {
	spell         *game.CardDef
	abilitySource *game.Permanent
}

func snapshotPool(s State, player *game.Player, ctx spendContext) map[mana.Unit]int {
	pool := player.ManaPool.Units()
	for _, rider := range player.ManaRiders {
		if rider.Rider.Restriction != game.ManaSpendRestrictedToCondition {
			continue
		}
		if restrictedManaCanPay(s, rider, ctx) {
			continue
		}
		if pool[rider.Unit] <= 1 {
			delete(pool, rider.Unit)
		} else {
			pool[rider.Unit]--
		}
	}
	return pool
}

func restrictedManaCanPay(s State, rider game.ManaRiderInstance, ctx spendContext) bool {
	switch rider.Rider.Condition {
	case game.ManaSpendCastChosenCreatureType:
		return rider.MatchesChosenCreatureType(ctx.spell)
	case game.ManaSpendCastOrActivateChosenCreatureType:
		return rider.MatchesChosenCreatureType(ctx.spell) ||
			abilitySourceIsChosenCreatureType(s, rider, ctx.abilitySource)
	case game.ManaSpendCastLegendarySpell:
		return ctx.spell != nil && ctx.spell.HasSupertype(types.Legendary)
	case game.ManaSpendCastArtifactSpell:
		// Powerstone: usable for anything except a nonartifact spell cast. A
		// non-spell payment (ability cost; ctx.spell is nil) is always allowed.
		return ctx.spell == nil || ctx.spell.HasType(types.Artifact)
	case game.ManaSpendCastArtifactSpellOnly:
		// Castle Doom, Mishra's Workshop: spendable only to cast an artifact
		// spell. A non-spell payment (ability cost; ctx.spell is nil) is not an
		// artifact spell, so the tagged mana cannot pay for it.
		return ctx.spell != nil && ctx.spell.HasType(types.Artifact)
	case game.ManaSpendCastOrActivateArtifact:
		// Power Depot, Cargo Ship: spendable to cast an artifact spell or to
		// activate an ability of an artifact permanent.
		return (ctx.spell != nil && ctx.spell.HasType(types.Artifact)) ||
			abilitySourceIsArtifact(s, ctx.abilitySource)
	case game.ManaSpendActivateArtifactAbility:
		// Soldevi Machinist: spendable only to activate an ability of an artifact
		// permanent, never to cast a spell.
		return ctx.spell == nil && abilitySourceIsArtifact(s, ctx.abilitySource)
	case game.ManaSpendCastArtifactOrActivateAbility:
		// Guidelight Optimizer, Automated Artificer: spendable to cast an artifact
		// spell or to activate any activated ability.
		return (ctx.spell != nil && ctx.spell.HasType(types.Artifact)) ||
			(ctx.spell == nil && ctx.abilitySource != nil)
	case game.ManaSpendCastCreatureSpell:
		// Beastcaller Savant: spendable only to cast a creature spell. A
		// non-spell payment (ability cost; ctx.spell is nil) is not a creature
		// spell, so the tagged mana cannot pay for it.
		return ctx.spell != nil && ctx.spell.HasType(types.Creature)
	case game.ManaSpendCastOrActivateCreature:
		// Castle Garenbrig: spendable to cast a creature spell or to activate an
		// ability of a creature permanent.
		return (ctx.spell != nil && ctx.spell.HasType(types.Creature)) ||
			abilitySourceIsCreature(s, ctx.abilitySource)
	case game.ManaSpendCastInstantOrSorcerySpell:
		// Vodalian Arcanist: spendable only to cast an instant or sorcery spell.
		// A non-spell payment (ability cost; ctx.spell is nil) is neither, so the
		// tagged mana cannot pay for it.
		return ctx.spell != nil &&
			(ctx.spell.HasType(types.Instant) || ctx.spell.HasType(types.Sorcery))
	case game.ManaSpendCastNoncreatureSpell:
		// Nardole, Resourceful Cyborg: spendable only to cast a noncreature
		// spell. A non-spell payment (ability cost; ctx.spell is nil) is not a
		// spell cast, so the tagged mana cannot pay for it.
		return ctx.spell != nil && !ctx.spell.HasType(types.Creature)
	case game.ManaSpendCastMulticoloredSpell:
		// Pillar of the Paruns: spendable only to cast a multicolored spell (two
		// or more colors). A non-spell payment (ability cost; ctx.spell is nil)
		// is not a spell cast, so the tagged mana cannot pay for it.
		return ctx.spell != nil && len(ctx.spell.Colors) >= 2
	case game.ManaSpendCastMonocoloredSpellOfChosenColor:
		// Throne of Eldraine: spendable only to cast a monocolored spell whose
		// single color matches the tagged mana's chosen color. A non-spell
		// payment (ability cost; ctx.spell is nil) is not a spell cast, so the
		// tagged mana cannot pay for it. Shares the predicate with the resolve
		// path so planning and firing never diverge.
		return rider.MatchesMonocoloredChosenColorSpell(ctx.spell)
	case game.ManaSpendCastPlaneswalkerSpell:
		// Interplanar Beacon: spendable only to cast a planeswalker spell. A
		// non-spell payment (ability cost; ctx.spell is nil) is not a planeswalker
		// spell, so the tagged mana cannot pay for it.
		return ctx.spell != nil && ctx.spell.HasType(types.Planeswalker)
	default:
		return false
	}
}

// abilitySourceIsChosenCreatureType reports whether source is a creature
// permanent of the subtype captured on the rider's mana unit, so the rider's
// tagged mana may pay to activate that source's ability (Secluded Courtyard).
func abilitySourceIsChosenCreatureType(s State, rider game.ManaRiderInstance, source *game.Permanent) bool {
	return source != nil &&
		types.KnownSubtypeForType(types.Creature, rider.ChosenSubtype) &&
		s.PermanentHasType(source, types.Creature) &&
		s.PermanentHasSubtype(source, rider.ChosenSubtype)
}

// abilitySourceIsArtifact reports whether source is an artifact permanent, so an
// artifact-restricted rider's tagged mana may pay to activate that source's
// ability (Power Depot, Soldevi Machinist).
func abilitySourceIsArtifact(s State, source *game.Permanent) bool {
	return source != nil && s.PermanentHasType(source, types.Artifact)
}

// abilitySourceIsCreature reports whether source is a creature permanent, so a
// creature-restricted rider's tagged mana may pay to activate that source's
// ability (Castle Garenbrig).
func abilitySourceIsCreature(s State, source *game.Permanent) bool {
	return source != nil && s.PermanentHasType(source, types.Creature)
}

// hasTapCostOf reports whether the cost list has a tap additional cost.
func hasTapCostOf(additionalCosts []cost.Additional) bool {
	for _, addCost := range additionalCosts {
		if addCost.Kind == cost.AdditionalTap {
			return true
		}
	}
	return false
}

// costHasVariableMana reports whether the cost contains an X (variable) symbol.
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

// manaCostPtr returns a pointer to the mana cost value, or nil if it does not exist.
func manaCostPtr(manaCost opt.V[cost.Mana]) *cost.Mana {
	if !manaCost.Exists {
		return nil
	}
	return &manaCost.Val
}

// canTapForAbility reports whether the permanent can be tapped as an ability cost.
func canTapForAbility(s State, p *game.Permanent) bool {
	if p.Tapped {
		return false
	}
	return summoningSicknessAllowsAbilityCost(s, p)
}

func canUntapForAbility(s State, p *game.Permanent) bool {
	if !p.Tapped {
		return false
	}
	return summoningSicknessAllowsAbilityCost(s, p)
}

// summoningSicknessAllowsAbilityCost reports whether a permanent may pay a {T} or
// {Q} cost in one of its own activated abilities. A creature normally can't while
// it has been under its controller's control for less than a full turn (CR
// 302.6), modeled by the summoning-sickness flag, but an active
// RuleEffectActivateAbilitiesAsThoughHaste its controller controls lifts that
// restriction (CR 702.10c, Thousand-Year Elixir).
func summoningSicknessAllowsAbilityCost(s State, p *game.Permanent) bool {
	if !s.PermanentHasType(p, types.Creature) || !p.SummoningSick {
		return true
	}
	return s.ActivateAbilitiesAsThoughHaste(s.EffectiveController(p))
}

// tapForAbility taps a permanent as an ability cost. forMana records
// tapped-for-mana provenance when the cost belongs to a mana ability.
func tapForAbility(s State, p *game.Permanent, forMana bool) bool {
	if !canTapForAbility(s, p) {
		return false
	}
	if forMana {
		s.SetTappedForMana(p)
		return true
	}
	s.SetTapped(p, true)
	return true
}

// canConvokeWith reports whether the permanent can be used for convoke.
func canConvokeWith(s State, playerID game.PlayerID, p *game.Permanent, exclude map[id.ID]bool) bool {
	if exclude[p.ObjectID] || p.Tapped || p.PhasedOut || s.EffectiveController(p) != playerID {
		return false
	}
	return s.PermanentHasType(p, types.Creature)
}

// canImproviseWith reports whether the permanent can be used for improvise.
func canImproviseWith(s State, playerID game.PlayerID, p *game.Permanent, exclude map[id.ID]bool) bool {
	if exclude[p.ObjectID] || p.Tapped || p.PhasedOut || s.EffectiveController(p) != playerID {
		return false
	}
	return s.PermanentHasType(p, types.Artifact)
}
