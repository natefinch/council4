package rules

import (
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/mtg/rules/payment"
	"github.com/natefinch/council4/opt"
)

// Resolution supplies permission and cost substitution, not another casting
// procedure. Timing and priority are owned by the resolving effect (608.2g).
func (e *Engine) resolutionSpell(g *game.Game, source *game.Player, card *game.CardInstance, cast action.CastSpellAction, request payment.SpellRequest, spell freeCastSpell, chooseBranches bool, agents [game.NumPlayers]PlayerAgent, log *TurnLog) actionOutcome {
	player := request.PlayerID
	def := request.Card
	object := &game.StackObject{
		Kind: game.StackSpell, SourceID: card.ID, Face: cast.Face, Controller: player,
		SourceZone: cast.SourceZone, Casting: true,
		ExileInsteadOfGraveyard: spell.exileOnResolution, Suspend: spell.suspend,
		CastDuringControllerMainPhase: player == g.Turn.ActivePlayer && g.Turn.IsMainPhase(),
	}
	var sourceID game.PlayerID
	if source != nil {
		sourceID = source.ID
	}
	if spell.copyCard {
		object.Copy, object.CastCopy = true, true
		object.CopyOwner = opt.Val(player)
		object.SourceCardID = card.ID
		object.SourceTokenDef = copyCardDef(card.Def)
		request.CardID = 0
	}
	update := func() {
		object.ChosenModes = slices.Clone(cast.ChosenModes)
		object.XValue, object.KickerPaid, object.KickerCount = cast.XValue, cast.KickerPaid, cast.KickerCount
		object.Bargained, object.OffspringPaid = cast.Bargained, cast.Offspring
		object.GiftPromised, object.GiftRecipient = cast.GiftPromised, cast.GiftRecipient
		request.XValue, request.KickerPaid, request.KickerCount = cast.XValue, cast.KickerPaid, cast.KickerCount
		request.Bargained, request.Offspring = cast.Bargained, cast.Offspring
		request.ChosenModes = cast.ChosenModes
	}
	var partial payment.Partial
	var splice spliceCastResult
	var events []game.Event
	descriptor := costedAction{
		object: object, usesStack: true,
		cost: activationCost{present: func(g *game.Game) bool { return costedSpellPresent(g, object) }},
		available: func(g *game.Game) bool {
			view := priorityCastView(g, object, sourceID)
			return len(e.resolutionSpellChoices(view, cast, request, chooseBranches)) > 0
		},
		announcements: func(g *game.Game, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
			choices := e.resolutionSpellChoices(g, cast, request, chooseBranches)
			if len(choices) == 0 {
				return false
			}
			if len(choices) > 1 {
				options := make([]freeCastOption, len(choices))
				for index, choice := range choices {
					options[index] = freeCastOption{cardID: card.ID, cast: choice}
				}
				index := 0
				if spellFreeRequest(request) {
					var ok bool
					index, ok = e.chooseFreeCastOption(g, player, card, cast.SourceZone, options, false, agents, log)
					if !ok {
						return false
					}
				} else {
					announcements := make([]game.ChoiceOption, len(choices))
					for i, choice := range choices {
						announcements[i] = game.ChoiceOption{Index: i, Label: freeCastOptionLabel(def, choice)}
					}
					index = e.chooseChoice(g, agents, game.ChoiceRequest{
						Kind: game.ChoiceModal, Player: player, Prompt: "Choose spell announcements.",
						Options: announcements, MinChoices: 1, MaxChoices: 1, DefaultSelection: []int{0},
					}, log)[0]
				}
				cast = choices[index]
			} else {
				cast = choices[0]
			}
			update()
			var ok bool
			announcedDef := def
			if request.Alternative.Exists && !spellFreeRequest(request) {
				copy := *def
				copy.ManaCost = request.Alternative.Val.ManaCost
				announcedDef = &copy
			}
			cast, ok = e.chooseCastAnnouncements(g, player, announcedDef, cast, spellFreeRequest(request), agents, log,
				func(announced action.CastSpellAction) { cast = announced; update() })
			update()
			if !ok {
				return false
			}
			_, ok = e.pendingSpellPayment(g, def, cast)(request, payment.Partial{})
			if !ok {
				g.Execution.InvalidChoices++
			}
			return ok
		},
		proposal: func(g *game.Game, agents [game.NumPlayers]PlayerAgent, log *TurnLog) (payment.Request, payment.Partial, bool) {
			complete := e.pendingSpellPayment(g, def, cast)
			_, announced, alternative, ok := e.choosePrioritySpellPayment(g, []castRoute{{}}, []payment.SpellRequest{request}, complete, agents, log)
			partial = announced
			object.AlternativeCost = alternative
			if ok && cast.SourceZone == zone.Hand && spellIsArcane(def) {
				splice = e.choosePrioritySplice(g, player, request.CardID, request, partial, alternative, agents, log)
				request.SpliceManaCosts = splice.manaCosts
				object.SplicedContent = splice.contents
			}
			return payment.ForSpell(request), partial, ok
		},
		targets: func(g *game.Game, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
			complete := func(request payment.SpellRequest, partial payment.Partial) (payment.Plan, bool) {
				return e.completeSplicedPayment(g, player, def, splice.cards, request, partial)
			}
			var ok bool
			branch := castBranchForObject(object)
			cast.Targets, ok = e.chooseCostedSpellTargets(g, player, def, cast, branch, request, partial, agents, log, complete)
			if !ok {
				return false
			}
			object.Targets = slices.Clone(cast.Targets)
			object.TargetCounts, ok = spellTargetCounts(g, player, def, cast.ChosenModes, cast.Targets, branch)
			if !ok || !e.completePrioritySpliceTargets(g, player, def, cast, &splice, request, partial, agents, log) {
				return false
			}
			object.SplicedTargets, object.SplicedTargetCounts = splice.targets, splice.targetCounts
			request.Targets = append(slices.Clone(cast.Targets), slices.Concat(splice.targets...)...)
			events = prepareTargetEvents(g, object)
			return true
		},
		lock: func(*game.Game) payment.Request { return payment.ForSpell(request) },
		commit: func(g *game.Game, paid payment.ApplyResult, _ [game.NumPlayers]PlayerAgent, _ *TurnLog, _ int) {
			result := paid.Spell
			object.Casting = false
			object.Blitzed = blitzAlternativeChosen(def, result.AlternativeIndex)
			object.BuybackPaid = result.BuybackPaid
			object.ConspirePaid = result.ConspirePaid
			object.AdditionalCostsPaid = result.AdditionalCostsPaid
			object.SacrificedAsCostIDs = result.SacrificedIDs
			object.PaidCostSubjects = result.PaidCostSubjects
			object.ColorsOfManaSpentToCast = distinctManaColorsSpent(result.PoolSpend)
			object.ManaSpentByColorToCast = manaSpentByColor(result.PoolSpend)
			object.ManaSpentToCast = totalManaSpent(result.PoolSpend)
			object.ManaFromCreaturesSpentToCast = creatureManaSpent(result.PoolSpend)
			typeLine := stackSpellTypeLine(g, object, def)
			emitSpellCastEventsWithTargets(g, object, game.Event{
				SourceID: card.ID, StackObjectID: object.ID, Controller: player,
				CardID: request.CardID, TokenDef: object.SourceTokenDef, Face: cast.Face,
				CardTypes: typeLine.types, CardSupertypes: typeLine.supertypes, CardSubtypes: typeLine.subtypes,
				Colors: spellColors(def), ManaValue: stackObjectKnownManaValue(g, object),
				ManaSpentToCast: opt.Val(object.ManaSpentToCast), ManaFromCreaturesSpentToCast: opt.Val(object.ManaFromCreaturesSpentToCast),
				KickerPaid: cast.KickerPaid, FromZone: cast.SourceZone, ToZone: zone.Stack,
			}, events)
			resolveSpellCastManaSpendRiders(g, player, paid.ManaBefore, result.PoolSpend, def, object)
		},
	}
	if !spell.copyCard {
		descriptor.spellSource = &costedSpellSource{player: sourceID, card: card.ID, zone: cast.SourceZone}
	}
	return e.costedAction(g, player, descriptor, agents, log)
}

func spellFreeRequest(request payment.SpellRequest) bool {
	return request.Alternative.Exists && request.Alternative.Val.Label == freeCastAlternativeCost().Label
}

// The list contains announcements, never bundled target decisions. Each one
// has an exact pending-target/payment witness in the CR 601.2a view.
func (e *Engine) resolutionSpellChoices(g *game.Game, cast action.CastSpellAction, request payment.SpellRequest, branches bool) []action.CastSpellAction {
	candidates := []action.CastSpellAction{cast}
	if branches {
		candidates = freeCastBranchOptions(g, request.PlayerID, cast.CardID, cast.SourceZone, cast.Face, request.Card)
	}
	var choices []action.CastSpellAction
	for _, candidate := range candidates {
		if additionalCostsRequirePositiveX(request.Card.AdditionalCosts) && candidate.XValue == 0 {
			candidate.XValue = 1
		}
		modes := [][]int{candidate.ChosenModes}
		if !modesValidForSpellAtBranch(g, request.PlayerID, request.Card, candidate.ChosenModes, castBranchForCast(candidate)) {
			modes = modeChoicesForSpellAtBranch(g, request.PlayerID, request.Card, castBranchForCast(candidate))
		}
		baseX := candidate.XValue
		for _, mode := range modes {
			candidate.XValue = baseX
			candidate.ChosenModes = mode
			for _, x := range resolutionSpellXValues(g, request, candidate) {
				candidate.XValue = x
				next := request
				next.XValue, next.KickerPaid, next.KickerCount = candidate.XValue, candidate.KickerPaid, candidate.KickerCount
				next.Bargained, next.Offspring, next.ChosenModes = candidate.Bargained, candidate.Offspring, mode
				if !castRouteAllowsX(request.Card, candidate.XValue, next) || !castGiftRecipientValid(g, request.PlayerID, candidate) {
					continue
				}
				view := resolutionAnnouncementView(g, candidate)
				if _, ok := e.pendingSpellPayment(view, request.Card, candidate)(next, payment.Partial{}); ok {
					choices = append(choices, candidate)
					break
				}
			}
		}
	}
	return choices
}

// X-dependent target legality changes at the candidates' characteristic
// values. Testing those breakpoints finds a minimum-cost witness even when
// the first legal X is above the action sampler's range.
func resolutionSpellXValues(g *game.Game, request payment.SpellRequest, cast action.CastSpellAction) []int {
	values := []int{cast.XValue}
	if spellFreeRequest(request) && costHasVariableMana(manaCostPtr(request.Card.ManaCost)) {
		return values
	}
	manaCost := request.Card.ManaCost
	if request.Alternative.Exists {
		manaCost = request.Alternative.Val.ManaCost
	}
	variable := costHasVariableMana(manaCostPtr(manaCost)) || additionalCostsUseX(request.Card.AdditionalCosts)
	if !variable {
		return values
	}
	specs := spellTargetSpecs(request.Card, cast.ChosenModes, castBranchForCast(cast))
	manaValueBound := specsBindTargetsToManaValueX(specs)
	powerBound := slices.ContainsFunc(specs, func(spec game.TargetSpec) bool { return spec.PowerAtMostX })
	if !manaValueBound && !powerBound {
		return values
	}
	add := func(value int, ok bool) {
		if ok && value >= cast.XValue && !slices.Contains(values, value) {
			values = append(values, value)
		}
	}
	for _, permanent := range g.Battlefield {
		if manaValueBound {
			add(effectivePermanentManaValue(g, permanent))
		}
		if powerBound {
			pt := effectivePermanentValues(g, permanent)
			add(pt.power, pt.powerOK)
		}
	}
	slices.Sort(values)
	return values
}

func resolutionAnnouncementView(g *game.Game, cast action.CastSpellAction) *game.Game {
	view := g.ResourceView(game.ResourceProjection{})
	view.Stack = game.Stack{}
	objects := g.Stack.Objects()
	for index, object := range objects {
		if index == len(objects)-1 && object.Casting {
			proposed := *object
			proposed.ChosenModes = slices.Clone(cast.ChosenModes)
			proposed.XValue, proposed.KickerPaid, proposed.KickerCount = cast.XValue, cast.KickerPaid, cast.KickerCount
			proposed.Bargained, proposed.OffspringPaid = cast.Bargained, cast.Offspring
			proposed.GiftPromised, proposed.GiftRecipient = cast.GiftPromised, cast.GiftRecipient
			object = &proposed
		}
		view.Stack.Push(object)
	}
	return view
}
