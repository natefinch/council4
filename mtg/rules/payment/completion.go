package payment

import (
	"maps"
	"reflect"
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

type requestKind uint8

const (
	spellPayment requestKind = iota + 1
	abilityPayment
	genericPayment
)

// Request describes a costed action. Use ForSpell, ForAbility or ForGeneric;
// the zero value is invalid. Card definitions are immutable shared definitions,
// while every per-game reference retained here is an ID, not a game pointer.
type Request struct {
	kind     requestKind
	spell    SpellRequest
	ability  AbilityRequest
	generic  GenericRequest
	sourceID id.ID
	locked   bool
}

// ForSpell describes the spell's announced costs, including fixed targets.
func ForSpell(req SpellRequest) Request {
	req.ChosenModes = slices.Clone(req.ChosenModes)
	req.Targets = slices.Clone(req.Targets)
	req.CastPermissions = slices.Clone(req.CastPermissions)
	if req.Alternative.Exists {
		req.Alternative.Val = copyAlternative(req.Alternative.Val)
	}
	req.SpliceManaCosts = slices.Clone(req.SpliceManaCosts)
	for i := range req.SpliceManaCosts {
		req.SpliceManaCosts[i] = slices.Clone(req.SpliceManaCosts[i])
	}
	if req.PreparedCost != nil {
		prepared := *req.PreparedCost
		prepared.option.manaCost = copyManaCost(prepared.option.manaCost)
		prepared.option.additionalCosts = slices.Clone(prepared.option.additionalCosts)
		prepared.Summary.ManaCost = copyManaCost(prepared.Summary.ManaCost)
		prepared.Summary.AdditionalCosts = slices.Clone(prepared.Summary.AdditionalCosts)
		req.PreparedCost = &prepared
	}
	req.Prefs = clonePreferences(req.Prefs)
	return Request{kind: spellPayment, spell: req}
}

// ForAbility describes an activation without retaining its mutable source.
func ForAbility(req AbilityRequest) Request {
	sourceID := id.ID(0)
	if req.Source != nil {
		sourceID = req.Source.ObjectID
	}
	req.Source = nil
	req.ManaCost.Val = slices.Clone(req.ManaCost.Val)
	req.AdditionalCosts = slices.Clone(req.AdditionalCosts)
	req.CostModifiers = slices.Clone(req.CostModifiers)
	for i := range req.CostModifiers {
		modifier := &req.CostModifiers[i]
		modifier.SetManaCost.Val = slices.Clone(modifier.SetManaCost.Val)
		modifier.ManaIncrease = slices.Clone(modifier.ManaIncrease)
		modifier.ColoredIncrease = slices.Clone(modifier.ColoredIncrease)
		modifier.ColoredReduction = slices.Clone(modifier.ColoredReduction)
		modifier.PerTargetBeyondFirstIncrease = slices.Clone(modifier.PerTargetBeyondFirstIncrease)
	}
	req.AlternativeCosts = slices.Clone(req.AlternativeCosts)
	for i := range req.AlternativeCosts {
		req.AlternativeCosts[i] = copyAlternative(req.AlternativeCosts[i])
	}
	req.Prefs = clonePreferences(req.Prefs)
	return Request{kind: abilityPayment, ability: req, sourceID: sourceID}
}

// ForGeneric describes a rule/effect payment without retaining its source.
func ForGeneric(req GenericRequest) Request {
	sourceID := id.ID(0)
	if req.Source != nil {
		sourceID = req.Source.ObjectID
	}
	req.Source = nil
	req.Cost = copyManaCost(req.Cost)
	req.AdditionalCosts = slices.Clone(req.AdditionalCosts)
	req.Exclude = cloneManaSourceRestrictions(manaSourceRestrictions{excluded: req.Exclude}).excluded
	req.Prefs = clonePreferences(req.Prefs)
	return Request{kind: genericPayment, generic: req, sourceID: sourceID}
}

func copyManaCost(manaCost *cost.Mana) *cost.Mana {
	if manaCost == nil {
		return nil
	}
	copied := slices.Clone(*manaCost)
	return &copied
}

func copyAlternative(alternative cost.Alternative) cost.Alternative {
	alternative.ManaCost.Val = slices.Clone(alternative.ManaCost.Val)
	alternative.AdditionalCosts = slices.Clone(alternative.AdditionalCosts)
	return alternative
}

// Partial fixes any subset of payment choices. Missing entries are undecided,
// not defaults. An Allocation fixes an exact previously selected mana payment.
type Partial struct {
	Option        opt.V[int]
	X             opt.V[int]
	Phyrexian     map[int]bool
	Components    map[int]ComponentChoice
	Groups        map[uint8]int
	Convoke       opt.V[[]id.ID]
	Improvise     opt.V[[]id.ID]
	Delve         opt.V[[]id.ID]
	Allocation    *ManaAllocation
	ManaAbilities map[ManaAbilityKey]bool
	joint         bool
	poolOnly      bool
}

type ManaAbilityKey struct {
	Source       id.ID
	AbilityIndex int
}

// ManaActivation records one selected source's activation and exact output.
type ManaActivation struct {
	Source       id.ID
	AbilityIndex int
	Color        mana.Color
	Amount       int
	Snow         bool
	FromCreature bool
	Untap        bool
	Sacrifice    bool
	SourceZone   zone.Type
	CostPlan     *Plan
	Choices      map[game.ChoiceKey]game.ResolutionChoiceResult
	Production   []ManaProduction
	Effects      game.ResourceProjection
}

// ManaAllocation is an independent snapshot of exact units spent, activations,
// nonmana mechanics, symbol payments and life paid.
type ManaAllocation struct {
	PoolSpend   map[mana.Unit]int
	Activations []ManaActivation
	Convoke     []id.ID
	Improvise   []id.ID
	Delve       []id.ID
	Symbols     []game.SymbolPayment
	Life        int
}

// Plan is a complete, immutable payment witness. Its private data contains no
// mutable game pointers. Accessors return independent copies; it can be applied
// to a clone of the game that produced it.
type Plan struct {
	request      Request
	option       int
	x            int
	components   map[int]ComponentChoice
	groups       map[uint8]int
	allocation   ManaAllocation
	cardVersions map[id.ID]uint64
	valid        bool
}

// CostOption returns the selected cost option's stable index.
func (p Plan) CostOption() int { return p.option }

// XValue returns the announced X.
func (p Plan) XValue() int { return p.x }

// Components returns the selections indexed by the original cost component.
func (p Plan) Components() map[int]ComponentChoice { return cloneComponents(p.components) }

// Mana returns an independent copy of the exact mana payment.
func (p Plan) Mana() ManaAllocation { return cloneAllocation(p.allocation) }

// ApplyResult carries payment evidence for spell, ability and generic callers.
type ApplyResult struct {
	Spell   SpellPaymentResult
	Ability AbilityCostPayment
	// ManaBefore is the pool immediately before the final cost is spent,
	// including planned production. Rider consumers must use this receipt
	// rather than a snapshot taken before mana abilities were activated.
	ManaBefore map[mana.Unit]int
}

type completion struct {
	mana       paymentPlan
	additional additionalCostPlan
	option     spellCostOption
	source     *game.Permanent
	tapSource  bool
	forMana    bool
	player     game.PlayerID
}

// Complete returns a deterministic witness when at least one joint payment
// satisfies partial. It performs only reads, never clone probes or mutations.
func (p Planner) Complete(request Request, partial Partial) (Plan, bool, string) {
	plan, _, ok := p.complete(request, partial)
	if !ok {
		return Plan{}, false, "no complete legal payment satisfies the fixed choices"
	}
	return plan, true, ""
}

func (p Planner) complete(request Request, partial Partial) (Plan, completion, bool) {
	partial.joint = true
	var plan Plan
	var result completion
	var source *game.Permanent
	if request.sourceID != 0 {
		var ok bool
		source, ok = p.s.PermanentByObjectID(request.sourceID)
		if !ok {
			return plan, result, false
		}
	}
	var player game.PlayerID
	var cardID, castingID id.ID
	var sourceZone zone.Type
	var options []spellCostOption
	var prefs *Preferences
	var x int
	var ctx spendContext
	var excluded map[id.ID]bool
	var tapSource, forMana bool
	switch request.kind {
	case spellPayment:
		req := request.spell
		if req.Card == nil {
			return plan, result, false
		}
		if partial.X.Exists {
			req.XValue = partial.X.Val
		}
		player, cardID, castingID, sourceZone, x, prefs = req.PlayerID, req.CardID, req.CardID, req.SourceZone, req.XValue, req.Prefs
		if req.PreparedCost != nil {
			options = spellCostOptionsForRequest(p.s, req)
		} else {
			for _, prepared := range p.PrepareSpellCosts(req) {
				options = append(options, prepared.option)
			}
		}
	case abilityPayment:
		req := request.ability
		req.Source = source
		if source == nil && req.SourceCardID == 0 ||
			source != nil && (source.PhasedOut || p.s.EffectiveController(source) != req.PlayerID) {
			return plan, result, false
		}
		if partial.X.Exists {
			req.XValue = partial.X.Val
		}
		player, cardID, sourceZone, x, prefs = req.PlayerID, req.SourceCardID, req.SourceZone, req.XValue, req.Prefs
		cardID, sourceZone = abilityCostSource(req)
		base := manaCostPtr(req.ManaCost)
		for _, modifier := range req.CostModifiers {
			if modifier.SetManaCost.Exists {
				base = &modifier.SetManaCost.Val
			}
		}
		if !request.locked && x != 0 && !costHasVariableMana(base) && !additionalCostsUseX(req.AdditionalCosts) {
			return plan, result, false
		}
		resolved, ok := resolveManaCostX(base, x)
		if !ok {
			return plan, result, false
		}
		resolved = applyManaCostModifiers(resolved, req.CostModifiers)
		options = []spellCostOption{{manaCost: resolved, additionalCosts: req.AdditionalCosts}}
		ctx.abilitySource = source
		tapSource, forMana = hasTapCostOf(req.AdditionalCosts), req.ForMana
		if tapSource && (source == nil || !canTapForAbility(p.s, source)) {
			return plan, result, false
		}
	case genericPayment:
		req := request.generic
		player, cardID, x, prefs = req.PlayerID, req.SourceCardID, req.XValue, req.Prefs
		if partial.X.Exists {
			x = partial.X.Val
		}
		options = []spellCostOption{{manaCost: req.Cost, additionalCosts: req.AdditionalCosts}}
		ctx.spell, excluded = req.Spell, req.Exclude
	default:
		return plan, result, false
	}
	if x < 0 {
		return plan, result, false
	}
	for _, option := range options {
		if partial.Option.Exists && option.index != partial.Option.Val ||
			prefs != nil && option.index != prefs.AlternativeIndex {
			continue
		}
		if request.kind != spellPayment {
			option.additionalCosts = lockAdditionalAmounts(p.s, player, option.additionalCosts)
		}
		currentCtx := ctx
		var spell *spellCostOption
		if request.kind == spellPayment {
			currentCtx = option.spendContext()
			spell = &option
		}
		var witness additionalWitness
		sawAdditional := false
		tryResources := func(state State, prefix []ManaActivation, projection game.ResourceProjection) (paymentPlan, bool) {
			currentSource := source
			if source != nil {
				var ok bool
				currentSource, ok = state.PermanentByObjectID(source.ObjectID)
				if !ok || tapSource && !canTapForAbility(state, currentSource) {
					return paymentPlan{}, false
				}
			}
			tail := partial
			if partial.Allocation != nil {
				allocation := cloneAllocation(*partial.Allocation)
				if len(prefix) > len(allocation.Activations) {
					return paymentPlan{}, false
				}
				allocation.Activations = slices.Clone(allocation.Activations[len(prefix):])
				if len(allocation.Activations) == 0 {
					allocation.Activations = nil
				}
				tail.Allocation = &allocation
			}
			tail.ManaAbilities = maps.Clone(partial.ManaAbilities)
			for _, activation := range prefix {
				key := ManaAbilityKey{Source: activation.Source, AbilityIndex: activation.AbilityIndex}
				if tail.ManaAbilities[key] {
					delete(tail.ManaAbilities, key)
				}
			}
			selected, ok := completeAdditional(state, player, option.additionalCosts, x, prefs, currentSource, cardID, sourceZone, castingID, partial.Components, partial.Groups, nil, func(witness additionalWitness) bool {
				sawAdditional = true
				for index := range partial.Components {
					if _, selected := witness.components[index]; !selected {
						return false
					}
				}
				for group := range partial.Groups {
					if _, selected := witness.groups[group]; !selected {
						return false
					}
				}
				restrictions := additionalManaRestrictions(excluded, witness.plan, true)
				restrictions.reserved = &witness.plan
				if tapSource {
					restrictions.excluded[source.ObjectID] = true
				}
				manaPlan, ok := completeMana(state, player, option.manaCost, x, restrictions, currentCtx, prefs, witness.plan.lifePaid, spell, cardID, sourceZone, tail)
				if !ok {
					return false
				}
				manaPlan.activations = append(slices.Clone(prefix), manaPlan.activations...)
				manaPlan.projection = mergeProjection(projection, manaPlan.projection)
				if !partial.acceptsMana(manaPlan) || !jointPaymentReady(p.s, player, manaPlan, witness.plan) {
					return false
				}
				result = completion{mana: manaPlan, additional: witness.plan, option: option, source: source, tapSource: tapSource, forMana: forMana, player: player}
				return true
			})
			if ok {
				witness = selected
			}
			return result.mana, ok
		}
		_, ok := tryResources(p.s, nil, game.ResourceProjection{})
		if !ok && !sawAdditional {
			if state, supported := p.s.(ManaState); supported {
				restrictions := manaSourceRestrictions{resourceCompletion: tryResources}
				_, ok = completeManaActivations(state, player, nil, 0, restrictions, spendContext{}, nil, 0, nil, 0, zone.None, partial)
			}
		}
		if !ok {
			continue
		}
		// Retain the locked spell total, not live modifiers (CR 601.2f).
		if request.kind == spellPayment {
			option.manaCost = copyManaCost(option.manaCost)
			option.additionalCosts = slices.Clone(option.additionalCosts)
			request.spell.PreparedCost = &PreparedSpellCost{option: option}
		} else if request.kind == abilityPayment {
			request.ability.AdditionalCosts = slices.Clone(option.additionalCosts)
			request.ability.ManaCost = opt.V[cost.Mana]{}
			if option.manaCost != nil {
				request.ability.ManaCost = opt.Val(slices.Clone(*option.manaCost))
			}
			request.ability.CostModifiers = nil
			request.ability.AlternativeCosts = nil
		} else {
			request.generic.AdditionalCosts = slices.Clone(option.additionalCosts)
		}
		request.locked = true
		versions := make(map[id.ID]uint64)
		recordVersion := func(cardID id.ID) {
			if card, ok := p.s.CardInstance(cardID); ok {
				versions[cardID] = card.ZoneVersion
			}
		}
		if request.kind == spellPayment {
			recordVersion(request.spell.CardID)
		}
		for index, component := range witness.components {
			switch option.additionalCosts[index].Kind {
			case cost.AdditionalDiscard, cost.AdditionalExile, cost.AdditionalReveal, cost.AdditionalCollectEvidence:
				for _, card := range component.Objects {
					recordVersion(card)
				}
			case cost.AdditionalExileSource:
				if sourceZone != zone.Battlefield {
					recordVersion(cardID)
				}
			}
		}
		for _, activation := range result.mana.activations {
			if activation.SourceZone == zone.Hand {
				recordVersion(activation.Source)
			}
		}
		return Plan{request: request, option: option.index, x: x, components: witness.components, groups: witness.groups, allocation: freezeMana(result.mana), cardVersions: versions, valid: true}, result, true
	}
	return plan, result, false
}

func lockAdditionalAmounts(s State, player game.PlayerID, costs []cost.Additional) []cost.Additional {
	result := make([]cost.Additional, 0, len(costs))
	for _, additional := range costs {
		if additional.AmountDynamic != cost.AdditionalDynamicAmountNone {
			additional.Amount = s.AdditionalDynamicAmountValue(player, additional.AmountDynamic)
			additional.AmountDynamic = cost.AdditionalDynamicAmountNone
			additional.AmountFromX = false
			if additional.Amount == 0 {
				continue
			}
		}
		result = append(result, additional)
	}
	return result
}

// Apply validates the exact witness against current state before any mutation.
// A stale or invalid plan is rejected without selecting a substitute payment.
func (p Planner) Apply(plan Plan) (ApplyResult, bool, string) {
	if view, ok := p.s.(interface{ ResourceReadOnly() bool }); ok && view.ResourceReadOnly() {
		return ApplyResult{}, false, "cannot apply a payment to a read-only resource view"
	}
	if !plan.valid {
		return ApplyResult{}, false, "invalid payment plan"
	}
	for cardID, version := range plan.cardVersions {
		card, ok := p.s.CardInstance(cardID)
		if !ok || card.ZoneVersion != version {
			return ApplyResult{}, false, "payment plan references a changed card object"
		}
	}
	allocation := plan.Mana()
	partial := Partial{Option: opt.Val(plan.option), X: opt.Val(plan.x), Components: plan.Components(), Groups: cloneGroups(plan.groups), Allocation: &allocation}
	_, current, ok := p.complete(plan.request, partial)
	if !ok {
		return ApplyResult{}, false, "payment plan is stale"
	}
	var before map[mana.Unit]int
	if len(current.mana.activations) > 0 {
		for _, activation := range current.mana.activations {
			if !p.s.(ManaState).ActivatePlannedMana(current.player, activation) {
				panic("prevalidated planned mana activation failed")
			}
		}
		payer, _ := p.s.Player(current.player)
		before = payer.ManaPool.Units()
		current.mana.activations = nil
		current.mana.projection = game.ResourceProjection{}
	}
	for i, permanent := range current.mana.convokeTaps {
		current.mana.convokeTaps[i], _ = p.s.PermanentByObjectID(permanent.ObjectID)
	}
	for i, permanent := range current.mana.improviseTaps {
		current.mana.improviseTaps[i], _ = p.s.PermanentByObjectID(permanent.ObjectID)
	}
	applyPaymentPlan(p.s, current.player, current.mana)
	if before == nil {
		payer, _ := p.s.Player(current.player)
		before = payer.ManaPool.Units()
		for unit, amount := range current.mana.poolSpend {
			before[unit] += amount
		}
	}
	if state, ok := p.s.(ManaState); ok && plan.request.kind != spellPayment {
		state.ConsumePaymentRiders(current.player, current.source, before, current.mana.poolSpend)
	}
	if current.tapSource && !tapForAbility(p.s, current.source, current.forMana) {
		panic("prevalidated costed-action source became untappable")
	}
	current.additional = translateAdditionalPlan(p.s, current.additional)
	subjects := applyAdditionalCostPlan(p.s, current.additional)
	spell := SpellPaymentResult{
		AlternativeIndex: current.option.alternativeIndex, BuybackPaid: current.option.buybackPaid,
		AdditionalCostsPaid: slices.Clone(current.additional.paid), PaidCostSubjects: subjects,
		SacrificedIDs: sacrificedPermanentIDs(current.additional), ReturnedAttackerIDs: returnedAttackerIDs(current.additional),
		PoolSpend: clonePoolSpend(current.mana.poolSpend), CastPermission: current.option.castPermission, Sneaked: current.option.sneak,
	}
	ability := AbilityCostPayment{
		PoolSpend: clonePoolSpend(current.mana.poolSpend), PaidCostSubjects: subjects,
		SacrificedIDs: sacrificedPermanentIDs(current.additional), TappedIDs: tappedPermanentIDs(current.additional), ExiledIDs: exiledCardIDs(current.additional),
	}
	return ApplyResult{Spell: spell, Ability: ability, ManaBefore: before}, true, ""
}

func freezeMana(plan paymentPlan) ManaAllocation {
	result := ManaAllocation{PoolSpend: clonePoolSpend(plan.poolSpend), Life: plan.lifePayment, Symbols: slices.Clone(plan.symbolPayments), Delve: slices.Clone(plan.delveExiles)}
	result.Activations = slices.Clone(plan.activations)
	for _, activation := range plan.manaTaps {
		result.Activations = append(result.Activations, ManaActivation{
			Source: activation.permanent.ObjectID, AbilityIndex: activation.abilityIndex, Color: activation.color,
			Amount: activation.amount, Snow: activation.snow, FromCreature: activation.fromCreature, Untap: activation.untap, Sacrifice: activation.sacrifice,
		})
	}
	for _, permanent := range plan.convokeTaps {
		result.Convoke = append(result.Convoke, permanent.ObjectID)
	}
	for _, permanent := range plan.improviseTaps {
		result.Improvise = append(result.Improvise, permanent.ObjectID)
	}
	return result
}

func cloneAllocation(a ManaAllocation) ManaAllocation {
	a.PoolSpend = clonePoolSpend(a.PoolSpend)
	a.Activations = slices.Clone(a.Activations)
	for i := range a.Activations {
		activation := &a.Activations[i]
		activation.Choices = maps.Clone(activation.Choices)
		activation.Production = slices.Clone(activation.Production)
		for j := range activation.Production {
			if rider := activation.Production[j].Rider; rider != nil {
				copied := *rider
				activation.Production[j].Rider = &copied
			}
		}
		activation.Effects = cloneProjection(activation.Effects)
		if activation.CostPlan != nil {
			copied := *activation.CostPlan
			copied.allocation = cloneAllocation(copied.allocation)
			copied.components = cloneComponents(copied.components)
			copied.groups = cloneGroups(copied.groups)
			copied.cardVersions = maps.Clone(copied.cardVersions)
			activation.CostPlan = &copied
		}
	}
	a.Convoke = slices.Clone(a.Convoke)
	a.Improvise = slices.Clone(a.Improvise)
	a.Delve = slices.Clone(a.Delve)
	a.Symbols = slices.Clone(a.Symbols)
	return a
}

func (p Partial) allowsPoolUnit(unit mana.Unit, spent int) bool {
	return p.Allocation == nil || spent < p.Allocation.PoolSpend[unit]
}

func (p Partial) allowsManaSource(source manaSource) bool {
	if p.Allocation == nil {
		return true
	}
	return slices.ContainsFunc(p.Allocation.Activations, func(a ManaActivation) bool {
		return a.Source == source.permanent.ObjectID && a.AbilityIndex == source.abilityIndex && a.Color == source.color && a.Amount == source.amount &&
			a.Snow == source.snow && a.FromCreature == source.fromCreature && a.Untap == source.untap && a.Sacrifice == source.sacrifice
	})
}

func (p Partial) allowsMechanic(object id.ID, selected opt.V[[]id.ID]) bool {
	return !selected.Exists || slices.Contains(selected.Val, object)
}

func (p Partial) acceptsMana(plan paymentPlan) bool {
	allocation := freezeMana(plan)
	for key, use := range p.ManaAbilities {
		used := slices.ContainsFunc(allocation.Activations, func(a ManaActivation) bool { return a.Source == key.Source && a.AbilityIndex == key.AbilityIndex })
		if used != use {
			return false
		}
	}
	if p.Allocation != nil && !reflect.DeepEqual(allocation, *p.Allocation) {
		return false
	}
	return (!p.Convoke.Exists || sameIDSet(allocation.Convoke, p.Convoke.Val)) &&
		(!p.Improvise.Exists || sameIDSet(allocation.Improvise, p.Improvise.Val)) &&
		(!p.Delve.Exists || sameIDSet(allocation.Delve, p.Delve.Val))
}

func sameIDSet(a, b []id.ID) bool {
	a, b = slices.Clone(a), slices.Clone(b)
	slices.Sort(a)
	slices.Sort(b)
	return slices.Equal(a, b)
}

// StepKind identifies the next payment decision.
type StepKind uint8

const (
	StepCostOption StepKind = iota
	StepX
	StepPhyrexian
	StepComponent
	StepCostGroup
	StepManaAbility
)

// Step specifies a decision. Index is the symbol/component/group index.
// Values supplies the action's legal X candidates (which may be bounded by its
// effect); other steps derive their candidate choices from the cost itself.
type Step struct {
	Kind   StepKind
	Index  int
	Values []int
	Source id.ID
}

// Choice is one value for a Step, with a complete payment witness.
type Choice struct {
	Value          int
	PayLife        bool
	Component      ComponentChoice
	Plan           Plan
	UseManaAbility bool
	Label          string
}

// With returns a partial with one additional choice fixed, leaving p untouched.
func (p Partial) With(step Step, choice Choice) Partial {
	p.Components = cloneComponents(p.Components)
	p.Groups = cloneGroups(p.Groups)
	phyrexian := make(map[int]bool, len(p.Phyrexian))
	for index, value := range p.Phyrexian {
		phyrexian[index] = value
	}
	p.Phyrexian = phyrexian
	p.ManaAbilities = maps.Clone(p.ManaAbilities)
	if p.ManaAbilities == nil {
		p.ManaAbilities = make(map[ManaAbilityKey]bool)
	}
	switch step.Kind {
	case StepCostOption:
		p.Option = opt.Val(choice.Value)
	case StepX:
		p.X = opt.Val(choice.Value)
	case StepPhyrexian:
		p.Phyrexian[step.Index] = choice.PayLife
	case StepComponent:
		p.Components[step.Index] = cloneComponent(choice.Component)
	case StepCostGroup:
		p.Groups[uint8(step.Index)] = choice.Value
	case StepManaAbility:
		p.ManaAbilities[ManaAbilityKey{Source: step.Source, AbilityIndex: step.Index}] = choice.UseManaAbility
	}
	return p
}

// Options offers only choices with a witness completion.
func (p Planner) Options(request Request, partial Partial, step Step) []Choice {
	var choices []Choice
	add := func(choice Choice) {
		plan, ok, _ := p.Complete(request, partial.With(step, choice))
		if ok {
			choice.Plan = plan
			choices = append(choices, choice)
		}
	}
	switch step.Kind {
	case StepManaAbility:
		label := "mana ability"
		if state, ok := p.s.(ManaState); ok {
			label = state.ManaAbilityLabel(step.Source, step.Index)
		}
		add(Choice{UseManaAbility: false, Label: "Do not use " + label})
		add(Choice{UseManaAbility: true, Label: "Use " + label})
	case StepX:
		for _, value := range step.Values {
			add(Choice{Value: value})
		}
	case StepPhyrexian:
		add(Choice{PayLife: false})
		add(Choice{PayLife: true})
	case StepCostOption:
		if request.kind == spellPayment {
			for _, option := range p.SpellCostOptions(request.spell) {
				add(Choice{Value: option.Index})
			}
		} else {
			add(Choice{})
		}
	case StepComponent, StepCostGroup:
		plan, current, ok := p.complete(request, partial)
		if !ok {
			return nil
		}
		costs := current.option.additionalCosts
		if step.Kind == StepCostGroup {
			for index, additional := range costs {
				if int(additional.ChoiceGroup) == step.Index {
					add(Choice{Value: index})
				}
			}
		} else if step.Index >= 0 && step.Index < len(costs) {
			cardID, castingID := request.generic.SourceCardID, id.ID(0)
			if request.kind == spellPayment {
				cardID, castingID = request.spell.CardID, request.spell.CardID
			} else if request.kind == abilityPayment {
				cardID = request.ability.SourceCardID
				if current.source != nil && cardID == 0 {
					cardID = current.source.CardInstanceID
				}
			}
			var candidates []Choice
			collect := func(state State, _ []ManaActivation, _ game.ResourceProjection) (paymentPlan, bool) {
				source := current.source
				if source != nil {
					source, _ = state.PermanentByObjectID(source.ObjectID)
				}
				New(state).nextComponentChoices(current.player, costs[step.Index], plan.x, source, cardID, castingID, partial.Components[step.Index], plan.components[step.Index], func(choice Choice) {
					if !slices.ContainsFunc(candidates, func(candidate Choice) bool { return reflect.DeepEqual(candidate.Component, choice.Component) }) {
						candidates = append(candidates, choice)
					}
				})
				return paymentPlan{}, false
			}
			collect(p.s, nil, game.ResourceProjection{})
			if state, supported := p.s.(ManaState); supported && !partial.Components[step.Index].Finished {
				collect(state.ProjectResources(current.mana.projection), nil, game.ResourceProjection{})
				// One witness can hide another source or counter kind. Discover
				// choices across resource projections, then validate each jointly.
				if !state.OrdinaryManaOnly(current.player) || costs[step.Index].RequireTapped {
					restrictions := manaSourceRestrictions{resourceCompletion: collect}
					completeManaActivations(state, current.player, nil, 0, restrictions, spendContext{}, nil, 0, nil, 0, zone.None, partial)
				}
			}
			for _, choice := range candidates {
				add(choice)
			}
		}
	}
	return choices
}

// A component step chooses one more object/counter, not an entire Cartesian
// subset. Threshold payments can finish once sufficient or select more.
func (p Planner) nextComponentChoices(player game.PlayerID, additional cost.Additional, x int, source *game.Permanent, sourceCardID, castingCardID id.ID, prefix, witness ComponentChoice, add func(Choice)) {
	if prefix.Finished {
		return
	}
	amount := additionalCostAmountIn(p.s, player, additional, x)
	var candidates []id.ID
	threshold := 0
	var weight func(id.ID) int
	switch additional.Kind {
	case cost.AdditionalSacrifice, cost.AdditionalTapPermanents, cost.AdditionalReturnToHand, cost.AdditionalReturnUnblockedAttacker:
		candidates = permanentIDs(CandidatePermanentsForCost(p.s, player, additional, returnCostSource(p.s, additional, source, sourceCardID)))
		threshold = additional.TotalPowerAtLeast
		weight = func(object id.ID) int {
			permanent, ok := p.s.PermanentByObjectID(object)
			if !ok {
				return 0
			}
			return permanentPowerContribution(p.s, permanent, additional.PowerContribution)
		}
	case cost.AdditionalDiscard, cost.AdditionalExile, cost.AdditionalReveal, cost.AdditionalCollectEvidence:
		if additional.Random {
			prefix.Finished = true
			add(Choice{Component: prefix})
			return
		}
		candidates = CandidateCardsForCost(p.s, player, additional, sourceCardID, castingCardID)
		threshold = additional.TotalManaValueAtLeast
		if additional.Kind == cost.AdditionalCollectEvidence {
			threshold = amount
		}
		weight = func(card id.ID) int {
			value, _ := evidenceCardManaValue(p.s, card)
			return value
		}
	case cost.AdditionalRemoveCounter, cost.AdditionalRemoveCounterAmong:
		if additional.RemoveAll {
			amount = len(witness.Counters)
		}
		if len(prefix.Counters) == amount {
			prefix.Finished = true
			add(Choice{Component: prefix})
			return
		}
		permanents := CandidatePermanentsForCost(p.s, player, additional, source)
		if additional.Kind == cost.AdditionalRemoveCounter {
			if source == nil {
				return
			}
			permanents = []*game.Permanent{source}
		}
		for _, permanent := range permanents {
			for _, kind := range RemovableCounterKinds(permanent, additional) {
				next := cloneComponent(prefix)
				next.Counters = append(next.Counters, CounterRemovalChoice{Permanent: permanent.ObjectID, Kind: kind})
				next.Finished = len(next.Counters) == amount
				add(Choice{Component: next})
			}
		}
		return
	default:
		prefix.Finished = true
		add(Choice{Component: prefix})
		return
	}
	enough := len(prefix.Objects) == amount
	if threshold > 0 {
		total := 0
		for _, object := range prefix.Objects {
			total += weight(object)
		}
		enough = total >= threshold
	}
	if enough {
		finished := cloneComponent(prefix)
		finished.Finished = true
		add(Choice{Component: finished})
		if threshold == 0 {
			return
		}
	}
	for _, object := range candidates {
		if slices.Contains(prefix.Objects, object) {
			continue
		}
		next := cloneComponent(prefix)
		next.Objects = append(next.Objects, object)
		next.Finished = threshold == 0 && len(next.Objects) == amount
		add(Choice{Component: next})
	}
}
