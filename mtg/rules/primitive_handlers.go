package rules

import (
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func (r *effectResolver) quantity(q game.Quantity) int {
	if q.IsDynamic() {
		return dynamicAmountValue(r.game, r.obj, stackObjectController(r.obj), q.DynamicAmount().Val)
	}
	return q.Value()
}

// quantityForPermanent evaluates a quantity whose dynamic amount may depend on
// the specific permanent being modified. The shared-creature-type-in-group count
// yields a different value per affected permanent (Shared Animosity counts the
// other attacking creatures sharing a creature type with the triggering
// attacker), so it is resolved relative to permanent; every other amount is
// independent of the affected permanent and delegates to quantity.
func (r *effectResolver) quantityForPermanent(q game.Quantity, permanent *game.Permanent) int {
	if q.IsDynamic() {
		dynamic := q.DynamicAmount().Val
		if dynamic.Kind == game.DynamicAmountSharedCreatureTypeCountInGroup {
			return dynamicAmountValueForPermanent(r.game, permanent, stackObjectController(r.obj), dynamic, 0)
		}
	}
	return r.quantity(q)
}

func (r *effectResolver) resolveObject(object game.ObjectReference) (*game.Permanent, bool) {
	resolved, ok := resolveObjectReference(r.game, r.obj, object)
	return resolved.permanent, ok && resolved.permanent != nil
}

func (r *effectResolver) resolvePlayer(player game.PlayerReference) (game.PlayerID, bool) {
	if player.Kind() == game.PlayerReferenceGroupOfferMember {
		if r.groupOfferMember.Exists {
			return r.groupOfferMember.Val, true
		}
		return 0, false
	}
	return resolvePlayerReference(r.game, r.obj, player)
}

func (r *effectResolver) recipientController(recipient game.PlayerReference) (game.PlayerID, bool) {
	if recipient.Kind() != game.PlayerReferenceNone {
		return r.resolvePlayer(recipient)
	}
	return r.obj.Controller, true
}

func (r *effectResolver) groupPermanents(group game.GroupReference) []*game.Permanent {
	ids := newReferenceResolver(r.game, r.obj).withGroupOfferMember(r.groupOfferMember).groupMembers(group)
	permanents := make([]*game.Permanent, 0, len(ids))
	for _, permanentID := range ids {
		if permanent, ok := permanentByObjectID(r.game, permanentID); ok {
			permanents = append(permanents, permanent)
		}
	}
	return permanents
}

func (r *effectResolver) groupPermanentsWithSource(group game.GroupReference, source *game.Permanent) []*game.Permanent {
	ids := newReferenceResolverWithSource(r.game, r.obj, source).withGroupOfferMember(r.groupOfferMember).groupMembers(group)
	permanents := make([]*game.Permanent, 0, len(ids))
	for _, permanentID := range ids {
		if permanent, ok := permanentByObjectID(r.game, permanentID); ok {
			permanents = append(permanents, permanent)
		}
	}
	return permanents
}

func (r *effectResolver) playerGroupMembers(group game.PlayerGroupReference) []game.PlayerID {
	return newReferenceResolver(r.game, r.obj).withGroupOfferMember(r.groupOfferMember).playerGroup(group)
}

func playersInAPNAPOrder(g *game.Game, players []game.PlayerID) []game.PlayerID {
	included := make(map[game.PlayerID]bool, len(players))
	for _, playerID := range players {
		included[playerID] = true
	}
	ordered := make([]game.PlayerID, 0, len(included))
	playerID := g.Turn.ActivePlayer
	for range game.NumPlayers {
		if included[playerID] {
			ordered = append(ordered, playerID)
		}
		playerID = g.TurnOrder.NextActivePlayer(playerID)
		if playerID == g.Turn.ActivePlayer {
			break
		}
	}
	return ordered
}

func handleDestroy(r *effectResolver, prim game.Destroy) effectResolved {
	res := effectResolved{accepted: true}
	targets := r.resolveObjectGroup(prim.Object, prim.Group)
	if !targets.single {
		destroyed := make([]*game.Permanent, 0, len(targets.permanents))
		for _, permanent := range targets.permanents {
			if hasKeyword(r.game, permanent, game.Indestructible) || replaceDestroyPermanent(r.game, permanent, prim.PreventRegeneration) {
				continue
			}
			destroyed = append(destroyed, permanent)
		}
		res.succeeded = movePermanentsToZoneSimultaneously(r.game, destroyed, zone.Graveyard)
		res.amount = len(destroyed)
		return res
	}
	if targets.resolved {
		_, res.succeeded = destroyPermanentInBatch(r.game, targets.permanents[0].ObjectID, 0, prim.PreventRegeneration)
	}
	return res
}

// handleAddMana resolves an "add mana" effect, putting the produced mana into the
// recipient's mana pool (CR 106.4: when an effect instructs a player to add mana,
// that mana goes into their mana pool). This is the effect of a mana ability or a
// mana-producing spell/ability (CR 106.3); the source of that mana is the spell
// itself, or for an ability the source of that ability (CR 106.3, with an
// ability's source identified by CR 113.7).
func handleAddMana(r *effectResolver, prim game.AddMana) effectResolved {
	res := effectResolved{accepted: true, amount: r.quantity(prim.Amount)}
	if res.amount <= 0 && !prim.Amount.IsDynamic() {
		res.amount = 1
	}
	recipientID := r.obj.Controller
	if prim.Player.Exists {
		resolved, ok := r.resolvePlayer(prim.Player.Val)
		if !ok {
			return res
		}
		recipientID = resolved
	}
	player, ok := playerByID(r.game, recipientID)
	if !ok || player.Eliminated {
		return res
	}
	addToPool := func(c mana.Color, amount int, snow bool) {
		switch {
		case prim.PersistUntilEndOfTurn && snow:
			player.ManaPool.AddPersistentSnow(c, amount)
		case prim.PersistUntilEndOfTurn:
			player.ManaPool.AddPersistent(c, amount)
		case snow:
			player.ManaPool.AddSnow(c, amount)
		default:
			player.ManaPool.Add(c, amount)
		}
	}
	if multiplier := tappedForManaProductionMultiplier(r.game, r.obj, recipientID); multiplier > 1 {
		res.amount *= multiplier
	}
	if prim.EachControlledColor != nil {
		snow := stackObjectSourceIsSnow(r.game, r.obj)
		for _, c := range controlledPermanentColors(r.game, recipientID, prim.EachControlledColor) {
			addToPool(c, res.amount, snow)
			res.succeeded = true
		}
		return res
	}
	if len(prim.CombinationColors) != 0 {
		snow := stackObjectSourceIsSnow(r.game, r.obj)
		for _, c := range r.combinationManaAllocation(recipientID, prim.CombinationColors, res.amount) {
			addToPool(c, 1, snow)
			res.succeeded = true
		}
		return res
	}
	manaColor := prim.ManaColor
	if choice, ok := linkedResolutionChoice(r.obj, string(prim.ChoiceFrom)); ok && choice.Kind == game.ResolutionChoiceMana {
		manaColor = choice.Color
	}
	if prim.EntryChoiceFrom != "" {
		choice, ok := linkedResolutionChoice(r.obj, string(prim.EntryChoiceFrom))
		if !ok || choice.Kind != game.ResolutionChoiceMana {
			return res
		}
		manaColor = choice.Color
	}
	chosenSubtype := types.Sub("")
	if prim.SpendRider.Exists {
		if key := prim.SpendRider.Val.ChosenSubtypeFrom; key != "" {
			choice, ok := linkedResolutionChoice(r.obj, string(key))
			if !ok || choice.Kind != game.ResolutionChoiceSubtype {
				return res
			}
			chosenSubtype = choice.Subtype
		}
	}
	snow := stackObjectSourceIsSnow(r.game, r.obj)
	addToPool(manaColor, res.amount, snow)
	if prim.SpendRider.Exists {
		unit := mana.Unit{Color: manaColor, Snow: snow}
		for i := 0; i < res.amount; i++ {
			player.ManaRiders = append(player.ManaRiders, game.ManaRiderInstance{
				Unit:           unit,
				Controller:     r.obj.Controller,
				SourceID:       r.obj.SourceCardID,
				SourceObjectID: r.obj.SourceID,
				ChosenSubtype:  chosenSubtype,
				Rider:          prim.SpendRider.Val,
			})
		}
	}
	res.succeeded = true
	return res
}

// tappedForManaProductionMultiplier returns the factor by which an activated
// mana ability's produced mana is scaled by a RuleEffectManaProductionMultiplier
// (Mana Reflection, Nyxbloom Ancient). It applies only when obj is an activated
// ability whose source permanent recipientID controls was tapped to pay for the
// mana, matching the "if you tap a permanent for mana" replacement; otherwise it
// returns 1. The payment path scales basic-land and planner-driven taps in the
// payment package, so this covers the standalone (floating) mana-ability path.
func tappedForManaProductionMultiplier(g *game.Game, obj *game.StackObject, recipientID game.PlayerID) int {
	if obj == nil || obj.Kind != game.StackActivatedAbility {
		return 1
	}
	multiplier := manaProductionMultiplierFor(g, recipientID)
	if multiplier <= 1 {
		return 1
	}
	permanent, ok := permanentByObjectID(g, obj.SourceID)
	if !ok || !permanent.Tapped || effectiveController(g, permanent) != recipientID {
		return 1
	}
	if !permanentTappedForManaIsCurrent(g, obj.SourceID) {
		return 1
	}
	return multiplier
}

// permanentTappedForManaIsCurrent reports whether the most recent tap event for
// permanentID recorded a tap that paid a mana ability (CR 106), i.e. the
// permanent's current tapped state was reached by tapping it for mana.
func permanentTappedForManaIsCurrent(g *game.Game, permanentID id.ID) bool {
	for i := len(g.Events) - 1; i >= 0; i-- {
		event := &g.Events[i]
		if event.Kind != game.EventPermanentTapped || event.PermanentID != permanentID {
			continue
		}
		return event.TappedForMana
	}
	return false
}

func handleAddCounter(r *effectResolver, prim game.AddCounter) effectResolved {
	if prim.AllKinds {
		return handleDoubleAllCounterKinds(r, prim)
	}
	res := effectResolved{accepted: true, amount: r.quantity(prim.Amount)}
	placementController := stackObjectController(r.obj)
	if prim.Group.Valid() && prim.DoubleKind {
		for _, permanent := range r.groupPermanents(prim.Group) {
			amount := permanent.Counters.Get(prim.CounterKind)
			if amount <= 0 {
				continue
			}
			if addCountersToPermanentControlledBy(r.game, placementController, permanent, prim.CounterKind, amount) {
				res.succeeded = true
			}
		}
		return res
	}
	if res.amount <= 0 {
		return res
	}
	if prim.Distribute {
		return r.addCountersDistributed(prim, res.amount)
	}
	if prim.Group.Valid() {
		if prim.ChooseOne {
			if permanent, ok := r.chooseOneGroupPermanent(prim.Group); ok {
				addCountersToPermanentControlledBy(r.game, placementController, permanent, prim.CounterKind, res.amount)
				res.succeeded = true
			}
			return res
		}
		for _, permanent := range r.groupPermanents(prim.Group) {
			if addCountersToPermanentControlledBy(r.game, placementController, permanent, prim.CounterKind, res.amount) {
				res.succeeded = true
			}
		}
		return res
	}
	permanent, ok := r.resolveObject(prim.Object)
	if ok {
		counterKind := prim.CounterKind
		if len(prim.KindChoices) != 0 {
			counterKind = r.chooseCounterKindToPlace(prim.KindChoices)
		}
		addCountersToPermanentControlledBy(r.game, placementController, permanent, counterKind, res.amount)
		res.succeeded = true
		if prim.PublishLinked != "" {
			rememberLinkedObject(
				r.game,
				linkedObjectSourceKey(r.game, r.obj, string(prim.PublishLinked)),
				permanentObjectBindingRef(permanent),
			)
		}
	}
	return res
}

// addCountersDistributed splits a fixed total of counters among the permanents
// chosen for a distribute-counters effect's target spec, each receiving at least
// one ("Distribute three +1/+1 counters among one, two, or three target
// creatures"). It mirrors damageDivided: the division spans every originally
// chosen target, an illegal target keeps its share but is placed no counters, and
// only still-legal permanents receive their allocation.
func (r *effectResolver) addCountersDistributed(prim game.AddCounter, total int) effectResolved {
	res := effectResolved{accepted: true, amount: total}
	targets := r.dividedTargets(prim.Object.TargetIndex())
	if len(targets) == 0 {
		return res
	}
	allocations := r.allocateCounters(total, targets)
	placementController := stackObjectController(r.obj)
	placedAny := false
	for i, entry := range targets {
		amount := allocations[i]
		// An illegal target keeps its share of the division but is placed no
		// counters; the amount is lost, never redistributed.
		if amount <= 0 || !entry.legal || entry.target.Kind != game.TargetPermanent {
			continue
		}
		permanent, found := permanentByObjectID(r.game, entry.target.PermanentID)
		if !found {
			continue
		}
		if addCountersToPermanentControlledBy(r.game, placementController, permanent, prim.CounterKind, amount) {
			placedAny = true
		}
	}
	res.succeeded = placedAny
	return res
}

// allocateCounters asks the controller to split total counters among every target
// chosen for the spec, returning one allocation per target. Each target receives
// at least one and the allocations sum to total, including targets that have
// since become illegal (whose share addCountersDistributed then drops). It
// mirrors allocateDividedDamage but raises a ChoiceCounterAllocation request.
func (r *effectResolver) allocateCounters(total int, targets []dividedDamageTarget) []int {
	n := len(targets)
	allocations := make([]int, n)
	if total < n {
		for i := range total {
			allocations[i] = 1
		}
		return allocations
	}
	options := make([]game.ChoiceOption, n)
	for i, entry := range targets {
		options[i] = game.ChoiceOption{Index: i, Label: entry.label}
	}
	request := game.ChoiceRequest{
		Kind:             game.ChoiceCounterAllocation,
		Player:           stackObjectController(r.obj),
		Prompt:           "Distribute counters among the chosen targets.",
		Options:          options,
		MinChoices:       total,
		MaxChoices:       total,
		DefaultSelection: defaultDividedAllocation(total, n),
	}
	selected := r.engine.chooseChoice(r.game, r.agents, request, r.log)
	for _, index := range selected {
		if index >= 0 && index < n {
			allocations[index]++
		}
	}
	return allocations
}

// combinationManaAllocation asks the recipient of an "add N mana in any
// combination of <colors>" effect to split the produced mana freely among the
// offered colors and returns one color per mana unit produced. Unlike counter
// and damage division, a color may receive zero, so the request uses the
// ChoiceManaCombination kind whose validity check permits empty shares. A
// non-positive amount produces no mana; a single-color set is added directly
// without a decision.
func (r *effectResolver) combinationManaAllocation(recipientID game.PlayerID, colors []mana.Color, amount int) []mana.Color {
	if amount <= 0 || len(colors) == 0 {
		return nil
	}
	if len(colors) == 1 {
		result := make([]mana.Color, amount)
		for i := range result {
			result[i] = colors[0]
		}
		return result
	}
	options := make([]game.ChoiceOption, len(colors))
	for i, c := range colors {
		options[i] = game.ChoiceOption{Index: i, Label: string(c)}
	}
	request := game.ChoiceRequest{
		Kind:             game.ChoiceManaCombination,
		Player:           recipientID,
		Prompt:           "Distribute the mana among the offered colors.",
		Options:          options,
		MinChoices:       amount,
		MaxChoices:       amount,
		DefaultSelection: defaultManaCombination(amount, len(colors)),
	}
	selected := r.engine.chooseChoice(r.game, r.agents, request, r.log)
	result := make([]mana.Color, 0, amount)
	for _, index := range selected {
		if index >= 0 && index < len(colors) {
			result = append(result, colors[index])
		}
	}
	return result
}

// defaultManaCombination spreads total mana round-robin across n colors so the
// engine has a valid ChoiceManaCombination fallback when no agent answers. The
// selection is a length-total multiset of color indices; round-robin keeps it
// deterministic without assuming every color receives a share.
func defaultManaCombination(total, n int) []int {
	if total <= 0 || n <= 0 {
		return nil
	}
	selected := make([]int, total)
	for i := range selected {
		selected[i] = i % n
	}
	return selected
}

// chooseCounterKindToPlace resolves which counter kind an "Put a <X> counter or a
// <Y> counter on it" effect places: the resolving controller chooses one of the
// listed kinds ("Put a +1/+1 counter or a loyalty counter on it.", Elspeth
// Conquers Death chapter III). The kinds are presented in listed order so the
// prompt is deterministic. An empty or unexpected selection falls back to the
// first kind.
func (r *effectResolver) chooseCounterKindToPlace(kinds []counter.Kind) counter.Kind {
	if len(kinds) == 0 {
		return 0
	}
	options := make([]game.ChoiceOption, len(kinds))
	for i, kind := range kinds {
		options[i] = game.ChoiceOption{
			Index: i,
			Label: kind.String() + " counter",
		}
	}
	selected := r.engine.chooseChoice(r.game, r.agents, game.ChoiceRequest{
		Kind:       game.ChoiceResolution,
		Player:     r.obj.Controller,
		Prompt:     "Choose a kind of counter",
		Options:    options,
		MinChoices: 1,
		MaxChoices: 1,
	}, r.log)
	if len(selected) != 1 || selected[0] < 0 || selected[0] >= len(kinds) {
		return kinds[0]
	}
	return kinds[selected[0]]
}

// handleDoubleAllCounterKinds doubles every kind of counter on the primitive's
// object ("double the number of each kind of counter on <permanent>", Vorel of
// the Hull Clade). It snapshots the counts before placing any counters so
// doubling one kind never feeds another, then adds, for each kind present, that
// many more counters. Kinds are visited in a deterministic order.
func handleDoubleAllCounterKinds(r *effectResolver, prim game.AddCounter) effectResolved {
	res := effectResolved{accepted: true}
	permanent, ok := r.resolveObject(prim.Object)
	if !ok {
		return res
	}
	placementController := stackObjectController(r.obj)
	counts := permanent.Counters.All()
	kinds := make([]counter.Kind, 0, len(counts))
	for kind := range counts {
		kinds = append(kinds, kind)
	}
	slices.Sort(kinds)
	for _, kind := range kinds {
		amount := counts[kind]
		if amount <= 0 {
			continue
		}
		if addCountersToPermanentControlledBy(r.game, placementController, permanent, kind, amount) {
			res.succeeded = true
			res.amount += amount
		}
	}
	return res
}

// handleAmass performs the amass keyword action (CR 701.44): it finds an Army
// the resolving controller controls and, if none exists, first creates a 0/0
// black Army creature token of the primitive's subtype, then puts Amount +1/+1
// counters on that Army.
func handleAmass(r *effectResolver, prim game.Amass) effectResolved {
	res := effectResolved{accepted: true, amount: r.quantity(prim.Amount)}
	if res.amount <= 0 {
		return res
	}
	controller := stackObjectController(r.obj)
	army := firstControlledArmy(r.game, controller)
	if army == nil {
		created, ok := createTokenPermanentsCollectingWithChoices(r.engine, r.game, controller, amassArmyTokenDef(prim.Subtype), 1, false, r.agents, r.log)
		if !ok || len(created) == 0 {
			return res
		}
		army = created[0]
	}
	if addCountersToPermanentControlledBy(r.game, controller, army, counter.PlusOnePlusOne, res.amount) {
		res.succeeded = true
	}
	return res
}

// handleBolster performs the bolster keyword action (CR 701.37): among creatures
// the resolving controller controls, it chooses one with the least toughness
// (the controller breaks a tie) and places Amount +1/+1 counters on it. It does
// nothing when the controller controls no creatures. When PublishLinked is set,
// the chosen creature is recorded under that key so a later linked effect can
// resolve it.
func handleBolster(r *effectResolver, prim game.Bolster) effectResolved {
	res := effectResolved{accepted: true, amount: r.quantity(prim.Amount)}
	if res.amount <= 0 {
		return res
	}
	controller := stackObjectController(r.obj)
	candidates := leastToughnessControlledCreatures(r, controller)
	if len(candidates) == 0 {
		return res
	}
	chosen, ok := r.chooseOnePermanentAmong(candidates, "Choose a creature with the least toughness")
	if !ok {
		return res
	}
	if addCountersToPermanentControlledBy(r.game, controller, chosen, counter.PlusOnePlusOne, res.amount) {
		res.succeeded = true
	}
	if prim.PublishLinked != "" {
		key := linkedObjectSourceKey(r.game, r.obj, string(prim.PublishLinked))
		clearLinkedObjects(r.game, key)
		rememberLinkedObject(r.game, key, permanentObjectBindingRef(chosen))
	}
	return res
}

// leastToughnessControlledCreatures returns every creature controller controls
// tied for the least effective toughness, the candidate set a bolster chooses
// from. Permanents without a defined toughness are excluded.
func leastToughnessControlledCreatures(r *effectResolver, controller game.PlayerID) []*game.Permanent {
	creatures := r.groupPermanents(game.BattlefieldGroup(game.Selection{
		RequiredTypes: []types.Card{types.Creature},
		Controller:    game.ControllerYou,
	}))
	least := 0
	haveLeast := false
	for _, permanent := range creatures {
		toughness, ok := effectiveToughness(r.game, permanent)
		if !ok {
			continue
		}
		if !haveLeast || toughness < least {
			least = toughness
			haveLeast = true
		}
	}
	if !haveLeast {
		return nil
	}
	candidates := make([]*game.Permanent, 0, len(creatures))
	for _, permanent := range creatures {
		toughness, ok := effectiveToughness(r.game, permanent)
		if ok && toughness == least {
			candidates = append(candidates, permanent)
		}
	}
	return candidates
}

// chooseOnePermanentAmong asks the resolving controller to choose exactly one
// permanent from candidates, mirroring chooseOneGroupPermanent but over a
// pre-narrowed slice (such as the least-toughness creatures a bolster selects
// among).
func (r *effectResolver) chooseOnePermanentAmong(candidates []*game.Permanent, prompt string) (*game.Permanent, bool) {
	if len(candidates) == 0 {
		return nil, false
	}
	options := make([]game.ChoiceOption, len(candidates))
	for i, permanent := range candidates {
		options[i] = game.ChoiceOption{
			Index: i,
			Label: permanentChoiceLabel(r.game, permanent),
			Card:  permanentChoiceInfo(r.game, permanent),
		}
	}
	selected := r.engine.chooseChoice(r.game, r.agents, game.ChoiceRequest{
		Kind:             game.ChoiceResolution,
		Player:           r.obj.Controller,
		Prompt:           prompt,
		Options:          options,
		MinChoices:       1,
		MaxChoices:       1,
		DefaultSelection: firstChoiceIndices(1),
	}, r.log)
	for _, idx := range selected {
		if idx >= 0 && idx < len(candidates) {
			return candidates[idx], true
		}
	}
	return nil, false
}

// firstControlledArmy returns the first Army permanent on the battlefield
// controlled by controller, or nil when the player controls no Army.
func firstControlledArmy(g *game.Game, controller game.PlayerID) *game.Permanent {
	for _, permanent := range g.Battlefield {
		if effectiveController(g, permanent) == controller && permanentHasSubtype(g, permanent, types.Army) {
			return permanent
		}
	}
	return nil
}

// amassArmyTokenDef builds the 0/0 black [subtype] Army creature token created
// when a player amasses without already controlling an Army (CR 701.44c). The
// named subtype precedes Army in the token's type line ("Orc Army").
func amassArmyTokenDef(subtype types.Sub) *game.CardDef {
	subtypes := []types.Sub{types.Army}
	name := "Army"
	if subtype != "" && subtype != types.Army {
		subtypes = []types.Sub{subtype, types.Army}
		name = string(subtype) + " Army"
	}
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      name,
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  subtypes,
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 0}),
		},
	}
}

func handleAddPlayerCounter(r *effectResolver, prim game.AddPlayerCounter) effectResolved {
	res := effectResolved{accepted: true, amount: r.quantity(prim.Amount)}
	if res.amount <= 0 {
		return res
	}
	controller := stackObjectController(r.obj)
	if prim.PlayerGroup.Kind != game.PlayerGroupReferenceNone {
		for _, playerID := range r.playerGroupMembers(prim.PlayerGroup) {
			player, ok := playerByID(r.game, playerID)
			if !ok || player.Eliminated {
				continue
			}
			if addCountersToPlayerControlledBy(r.game, controller, player, prim.CounterKind, res.amount) {
				res.succeeded = true
			}
		}
		return res
	}
	playerID, ok := r.resolvePlayer(prim.Player)
	if !ok {
		return res
	}
	player, ok := playerByID(r.game, playerID)
	if !ok || player.Eliminated {
		return res
	}
	if addCountersToPlayerControlledBy(r.game, controller, player, prim.CounterKind, res.amount) {
		res.succeeded = true
	}
	return res
}

func handleMoveCounters(r *effectResolver, prim game.MoveCounters) effectResolved {
	if prim.Distribute {
		return r.distributeMoveCounters(prim)
	}
	res := effectResolved{accepted: true, amount: r.quantity(prim.Amount)}
	destination, ok := r.resolveObject(prim.Object)
	if !ok {
		return res
	}
	counters, source, ok := effectCounterSource(r.game, r.obj, prim.Source)
	if !ok || counters.IsEmpty() || source != nil && source.ObjectID == destination.ObjectID {
		return res
	}
	if prim.AllKinds {
		for kind, amount := range counters.All() {
			addCountersToPermanentControlledBy(r.game, stackObjectController(r.obj), destination, kind, amount)
			if source != nil {
				source.Counters.Remove(kind, amount)
			}
		}
		res.succeeded = true
		return res
	}
	moveKind := prim.CounterKind
	if prim.ChooseKind {
		chosen, ok := r.chooseCounterKindToMove(counters)
		if !ok {
			return res
		}
		moveKind = chosen
	}
	available := counters.Get(moveKind)
	moved := min(available, res.amount)
	if moved <= 0 {
		return res
	}
	addCountersToPermanentControlledBy(r.game, stackObjectController(r.obj), destination, moveKind, moved)
	if source != nil {
		source.Counters.Remove(moveKind, moved)
	}
	res.succeeded = true
	return res
}

// chooseCounterKindToMove resolves which single counter kind a "Move a counter"
// or "Remove a counter" effect acts on: the lone kind present on the source, or
// the controller's choice when the source carries counters of more than one
// kind. The candidate kinds are presented in counter-kind order so the prompt is
// deterministic.
func (r *effectResolver) chooseCounterKindToMove(counters counter.Set) (counter.Kind, bool) {
	var kinds []counter.Kind
	for kind, amount := range counters.All() {
		if amount > 0 {
			kinds = append(kinds, kind)
		}
	}
	if len(kinds) == 0 {
		return 0, false
	}
	slices.Sort(kinds)
	if len(kinds) == 1 {
		return kinds[0], true
	}
	options := make([]game.ChoiceOption, len(kinds))
	for i, kind := range kinds {
		options[i] = game.ChoiceOption{
			Index: i,
			Label: kind.String() + " counter",
		}
	}
	selected := r.engine.chooseChoice(r.game, r.agents, game.ChoiceRequest{
		Kind:       game.ChoiceResolution,
		Player:     r.obj.Controller,
		Prompt:     "Choose a kind of counter",
		Options:    options,
		MinChoices: 1,
		MaxChoices: 1,
	}, r.log)
	if len(selected) != 1 || selected[0] < 0 || selected[0] >= len(kinds) {
		return kinds[0], true
	}
	return kinds[selected[0]], true
}

// distributeMoveCounters implements the "move any number of <kind> counters from
// this permanent onto other creatures" form (CR, e.g. Forgotten Ancient). The
// controller distributes the source's counters of the given kind among the
// permanents of Group one at a time, choosing a destination for each counter or
// stopping early, until the source runs out of counters of that kind.
func (r *effectResolver) distributeMoveCounters(prim game.MoveCounters) effectResolved {
	res := effectResolved{accepted: true}
	counters, source, ok := effectCounterSource(r.game, r.obj, prim.Source)
	if !ok || source == nil {
		return res
	}
	available := counters.Get(prim.CounterKind)
	candidates := r.groupPermanents(*prim.Group)
	if available <= 0 || len(candidates) == 0 {
		return res
	}
	placementController := stackObjectController(r.obj)
	for available > 0 {
		options := make([]game.ChoiceOption, len(candidates))
		for i, permanent := range candidates {
			options[i] = game.ChoiceOption{
				Index: i,
				Label: permanentChoiceLabel(r.game, permanent),
				Card:  permanentChoiceInfo(r.game, permanent),
			}
		}
		selected := r.engine.chooseChoice(r.game, r.agents, game.ChoiceRequest{
			Kind:       game.ChoiceResolution,
			Player:     r.obj.Controller,
			Prompt:     "Choose a creature to move a counter onto, or stop",
			Options:    options,
			MinChoices: 0,
			MaxChoices: 1,
		}, r.log)
		if len(selected) == 0 {
			break
		}
		index := selected[0]
		if index < 0 || index >= len(candidates) {
			break
		}
		addCountersToPermanentControlledBy(r.game, placementController, candidates[index], prim.CounterKind, 1)
		source.Counters.Remove(prim.CounterKind, 1)
		available--
		res.amount++
		res.succeeded = true
	}
	return res
}

func handleApplyContinuous(r *effectResolver, prim game.ApplyContinuous) effectResolved {
	res := effectResolved{accepted: true}
	if prim.ChooseFrom.Valid() {
		effects := r.resolveChosenColorProtection(prim.ContinuousEffects)
		for _, permanent := range r.chooseApplyContinuousPermanents(prim) {
			if applyTypedContinuousEffects(r.game, r.obj, permanent, effects, prim.Duration) {
				res.succeeded = true
			}
		}
		return res
	}
	var permanent *game.Permanent
	if prim.Object.Exists {
		permanent, _ = r.resolveObject(prim.Object.Val)
	}
	effects := r.resolveChosenColorProtection(prim.ContinuousEffects)
	res.succeeded = applyTypedContinuousEffects(r.game, r.obj, permanent, effects, prim.Duration)
	if prim.PublishLinked != "" && permanent != nil {
		// This publish path records exactly one permanent (the multi-permanent
		// ChooseFrom path returned above), so clear any permanent a prior
		// resolution published under this source-and-link-scoped key before
		// remembering this resolution's permanent. The key is constant across
		// repeated same-turn activations, so without clearing a repeatable
		// "target creature gains ... until end of turn. Sacrifice it at the
		// beginning of the next end step." (Krovikan Elementalist) or a
		// per-spell reanimation (Moira and Teshar) would leave the key holding
		// [obj1, obj2]; the schedule-time capture resolves the first entry, so
		// every activation captures obj1 and the later permanent leaks. Mirrors
		// the other single-binding publish sites (handleCreateToken,
		// handlePutOnBattlefield) that clear before remembering.
		key := linkedObjectSourceKey(r.game, r.obj, string(prim.PublishLinked))
		clearLinkedObjects(r.game, key)
		rememberLinkedObject(r.game, key, permanentLinkedObjectRef(permanent))
	}
	return res
}

// chooseApplyContinuousPermanents prompts the resolving controller to choose up
// to the primitive's dynamic amount of distinct permanents from its candidate
// group ("up to that many target lands you control", Primal Adversary). It
// returns the chosen permanents, or nil when the amount or candidate set is
// empty, so the continuous effect applies to nothing.
func (r *effectResolver) chooseApplyContinuousPermanents(prim game.ApplyContinuous) []*game.Permanent {
	amount := r.quantity(prim.ChooseUpTo)
	if amount <= 0 {
		return nil
	}
	candidates := r.groupPermanents(prim.ChooseFrom)
	maxChoices := min(amount, len(candidates))
	if maxChoices == 0 {
		return nil
	}
	options := make([]game.ChoiceOption, len(candidates))
	for i, permanent := range candidates {
		options[i] = game.ChoiceOption{
			Index: i,
			Label: permanentChoiceLabel(r.game, permanent),
			Card:  permanentChoiceInfo(r.game, permanent),
		}
	}
	prompt := prim.Prompt
	if prompt == "" {
		prompt = "Choose permanents"
	}
	selected := r.engine.chooseChoice(r.game, r.agents, game.ChoiceRequest{
		Kind:             game.ChoiceResolution,
		Player:           r.obj.Controller,
		Prompt:           prompt,
		Options:          options,
		MinChoices:       0,
		MaxChoices:       maxChoices,
		DefaultSelection: firstChoiceIndices(maxChoices),
	}, r.log)
	chosen := make([]*game.Permanent, 0, len(selected))
	for _, idx := range selected {
		if idx >= 0 && idx < len(candidates) {
			chosen = append(chosen, candidates[idx])
		}
	}
	return chosen
}

// resolveChosenColorProtection rewrites any granted "protection from the color
// of your choice" ability into protection from a concrete color chosen by the
// resolving ability's controller. The choice is made once as the ability
// resolves; the returned templates are freshly cloned where a rewrite happens so
// the card definition's shared continuous-effect template is left untouched.
func (r *effectResolver) resolveChosenColorProtection(templates []game.ContinuousEffect) []game.ContinuousEffect {
	result := templates
	cloned := false
	for i := range templates {
		for j, ability := range templates[i].AddAbilities {
			static, ok := ability.(*game.StaticAbility)
			if !ok {
				continue
			}
			prot, ok := game.StaticBodyProtectionKeyword(static)
			if !ok || !prot.ChosenColor {
				continue
			}
			chosen, ok := r.chooseProtectionColor(r.obj.Controller)
			if !ok {
				continue
			}
			resolved := game.ProtectionFromColorsStaticAbility(chosen)
			if !cloned {
				result = append([]game.ContinuousEffect(nil), templates...)
				cloned = true
			}
			abilities := append([]game.Ability(nil), result[i].AddAbilities...)
			abilities[j] = &resolved
			result[i].AddAbilities = abilities
		}
	}
	return result
}

// chooseProtectionColor prompts the player to pick one of the five colors for a
// chosen-color protection grant.
func (r *effectResolver) chooseProtectionColor(controller game.PlayerID) (color.Color, bool) {
	engine := r.engine
	if engine == nil {
		engine = NewEngine(nil)
	}
	choice := game.ResolutionChoice{
		Kind:   game.ResolutionChoiceMana,
		Prompt: "Choose a color.",
		Colors: []mana.Color{mana.W, mana.U, mana.B, mana.R, mana.G},
	}
	result, ok := engine.chooseEntryColor(r.game, r.agents, controller, &choice, r.log)
	if !ok {
		return "", false
	}
	return manaColor(result.Color)
}

func handleApplyRule(r *effectResolver, prim game.ApplyRule) effectResolved {
	return effectResolved{
		accepted:  true,
		succeeded: createRuleEffectTemplates(r.game, r.obj, prim.Object, prim.RuleEffects, prim.Duration),
	}
}

func handleModifyPT(r *effectResolver, prim game.ModifyPT) effectResolved {
	res := effectResolved{accepted: true}
	permanent, ok := r.resolveObject(prim.Object)
	if !ok || prim.Duration != game.DurationUntilEndOfTurn {
		return res
	}
	powerDelta := r.quantityForPermanent(prim.PowerDelta, permanent)
	toughnessDelta := r.quantityForPermanent(prim.ToughnessDelta, permanent)
	r.game.ContinuousEffects = append(r.game.ContinuousEffects, untilEndOfTurnPTContinuousEffect(r.game, r.obj, permanent, powerDelta, toughnessDelta))
	if prim.PublishLinked != "" {
		// Records exactly one permanent, so clear any permanent a prior
		// resolution published under this source-and-link-scoped key before
		// remembering this one; see handleApplyContinuous for why single-binding
		// publish sites must clear so each same-turn activation's schedule-time
		// capture binds to the permanent this activation acted on.
		key := linkedObjectSourceKey(r.game, r.obj, string(prim.PublishLinked))
		clearLinkedObjects(r.game, key)
		rememberLinkedObject(r.game, key, permanentLinkedObjectRef(permanent))
	}
	res.succeeded = true
	return res
}

func handleTap(r *effectResolver, prim game.Tap) effectResolved {
	res := effectResolved{accepted: true}
	targets := r.resolveObjectGroup(prim.Object, prim.Group)
	if !targets.single {
		res.succeeded = setPermanentsTappedSimultaneously(r.game, targets.permanents, true)
		return res
	}
	if targets.resolved {
		setPermanentTapped(r.game, targets.permanents[0], true)
		res.succeeded = true
	}
	return res
}

func handleTapOrUntap(r *effectResolver, prim game.TapOrUntap) effectResolved {
	res := effectResolved{accepted: true}
	permanent, ok := r.resolveObject(prim.Object)
	if !ok {
		return res
	}
	selected := r.engine.chooseChoice(r.game, r.agents, game.ChoiceRequest{
		Kind:   game.ChoiceResolution,
		Player: r.obj.Controller,
		Prompt: "Tap or untap the permanent?",
		Options: []game.ChoiceOption{
			{Index: 0, Label: "Tap"},
			{Index: 1, Label: "Untap"},
		},
		MinChoices: 1,
		MaxChoices: 1,
	}, r.log)
	tap := len(selected) != 1 || selected[0] != 1
	setPermanentTapped(r.game, permanent, tap)
	res.succeeded = true
	return res
}

func handleStartEngines(r *effectResolver, prim game.StartEngines) effectResolved {
	res := effectResolved{accepted: true}
	playerID, ok := r.resolvePlayer(prim.Player)
	if ok {
		res.succeeded = startEngines(r.game, playerID)
	}
	return res
}

func handleSetClassLevel(r *effectResolver, prim game.SetClassLevel) effectResolved {
	res := effectResolved{accepted: true, amount: r.quantity(prim.Amount)}
	permanent, ok := r.resolveObject(prim.Object)
	if ok && res.amount > permanent.ClassLevel {
		previous := permanent.ClassLevel
		permanent.ClassLevel = res.amount
		res.succeeded = true
		for level := previous + 1; level <= res.amount; level++ {
			emitEvent(r.game, game.Event{
				Kind:           game.EventClassLevelGained,
				SourceObjectID: permanent.ObjectID,
				PermanentID:    permanent.ObjectID,
				SourceID:       permanent.CardInstanceID,
				Controller:     effectiveController(r.game, permanent),
				Amount:         level,
			})
		}
	}
	return res
}

func handleMonstrosity(r *effectResolver, prim game.Monstrosity) effectResolved {
	res := effectResolved{accepted: true, amount: r.quantity(prim.Amount)}
	permanent, ok := r.resolveObject(prim.Object)
	if ok && !permanent.Monstrous {
		if res.amount > 0 {
			addCountersToPermanentControlledBy(r.game, stackObjectController(r.obj), permanent, counter.PlusOnePlusOne, res.amount)
		}
		permanent.Monstrous = true
		res.succeeded = true
	}
	return res
}

func handleRenown(r *effectResolver, prim game.Renown) effectResolved {
	res := effectResolved{accepted: true, amount: r.quantity(prim.Amount)}
	permanent, ok := r.resolveObject(prim.Object)
	if ok && !permanent.Renowned {
		if res.amount > 0 {
			addCountersToPermanentControlledBy(r.game, stackObjectController(r.obj), permanent, counter.PlusOnePlusOne, res.amount)
		}
		permanent.Renowned = true
		res.succeeded = true
	}
	return res
}

func handleAdapt(r *effectResolver, prim game.Adapt) effectResolved {
	res := effectResolved{accepted: true, amount: r.quantity(prim.Amount)}
	permanent, ok := r.resolveObject(prim.Object)
	if ok && permanent.Counters.Get(counter.PlusOnePlusOne) == 0 {
		if res.amount > 0 {
			addCountersToPermanentControlledBy(r.game, stackObjectController(r.obj), permanent, counter.PlusOnePlusOne, res.amount)
		}
		res.succeeded = true
	}
	return res
}

func handleBecomeSaddled(r *effectResolver, prim game.BecomeSaddled) effectResolved {
	res := effectResolved{accepted: true}
	permanent, ok := r.resolveObject(prim.Object)
	if ok && !permanent.Saddled {
		permanent.Saddled = true
		res.succeeded = true
	}
	return res
}

func handlePay(r *effectResolver, prim game.Pay) effectResolved {
	payment := prim.Payment
	if payment.Prompt == "" {
		payment.Prompt = prim.Prompt
	}
	accepted, succeeded := r.engine.resolveResolutionPaymentValue(r.game, r.obj, &payment, r.agents, r.log)
	return effectResolved{accepted: accepted, succeeded: succeeded}
}

// maxResolutionPayRepeatCount bounds how many times an unbounded PayRepeatedly
// cost may be paid in one resolution, matching the Multikicker enumeration cap so
// a free or fully-affordable cost cannot iterate without limit. A PayRepeatedly
// carrying a MaxCount instead caps at that rules-derived bound, which is itself a
// finite triggering quantity.
const maxResolutionPayRepeatCount = 20

func handlePayRepeatedly(r *effectResolver, prim game.PayRepeatedly) effectResolved {
	limit := maxResolutionPayRepeatCount
	if prim.MaxCount.Exists && prim.MaxCount.Val != nil {
		limit = max(0, dynamicAmountValue(r.game, r.obj, stackObjectController(r.obj), *prim.MaxCount.Val))
	}
	count := 0
	for count < limit {
		payment := prim.Payment
		if payment.Prompt == "" {
			payment.Prompt = prim.Prompt
		}
		accepted, succeeded := r.engine.resolveResolutionPaymentValue(r.game, r.obj, &payment, r.agents, r.log)
		if !accepted || !succeeded {
			break
		}
		count++
	}
	rememberResolutionChoice(r.obj, string(prim.PublishCount), game.ResolutionChoiceResult{
		Kind:   game.ResolutionChoiceNumber,
		Number: count,
	})
	return effectResolved{accepted: true, succeeded: count > 0, amount: count}
}

func handleChoose(r *effectResolver, prim game.Choose) effectResolved {
	succeeded := r.engine.resolveResolutionChoiceValue(r.game, r.obj, &prim.Choice, string(prim.PublishChoice), r.agents, r.log)
	return effectResolved{accepted: true, succeeded: succeeded}
}

func handleGainLife(r *effectResolver, prim game.GainLife) effectResolved {
	perPlayer := r.quantity(prim.Amount)
	res := effectResolved{accepted: true, amount: perPlayer}
	if perPlayer <= 0 {
		return res
	}
	if prim.PlayerGroup.Kind != game.PlayerGroupReferenceNone {
		res.amount = 0
		for _, playerID := range r.playerGroupMembers(prim.PlayerGroup) {
			gained := gainLife(r.game, playerID, perPlayer)
			res.amount += gained
			res.succeeded = gained > 0 || res.succeeded
		}
		return res
	}
	playerID, ok := r.resolvePlayer(prim.Player)
	if ok {
		res.succeeded = gainLife(r.game, playerID, res.amount) > 0
	}
	return res
}

func handleLoseLife(r *effectResolver, prim game.LoseLife) effectResolved {
	perPlayer := r.quantity(prim.Amount)
	res := effectResolved{accepted: true, amount: perPlayer}
	if perPlayer <= 0 {
		return res
	}

	if prim.PlayerGroup.Kind != game.PlayerGroupReferenceNone {
		res.amount = 0
		for _, playerID := range r.playerGroupMembers(prim.PlayerGroup) {
			lost := loseLife(r.game, playerID, perPlayer)
			res.amount += lost
			res.succeeded = lost > 0 || res.succeeded
		}
		return res
	}

	playerID, ok := r.resolvePlayer(prim.Player)
	if ok {
		res.succeeded = loseLife(r.game, playerID, res.amount) > 0
	}
	return res
}

func handleExchangeLifeTotalWithSourceCharacteristic(
	r *effectResolver,
	prim game.ExchangeLifeTotalWithSourceCharacteristic,
) effectResolved {
	res := effectResolved{accepted: true}
	playerID, ok := r.resolvePlayer(prim.Player)
	if !ok {
		return res
	}
	player, ok := playerByID(r.game, playerID)
	if !ok || player.Eliminated {
		return res
	}
	source, ok := sourcePermanent(r.game, r.obj)
	if !ok {
		return res
	}
	values := effectivePermanentValues(r.game, source)
	var characteristic int
	var effect game.ContinuousEffect
	effect.Layer = game.LayerPowerToughnessSet
	switch prim.Characteristic {
	case game.SourcePower:
		if !values.powerOK {
			return res
		}
		characteristic = values.power
		effect.SetPower = opt.Val(game.PT{Value: player.Life})
	case game.SourceToughness:
		if !values.toughnessOK {
			return res
		}
		characteristic = values.toughness
		effect.SetToughness = opt.Val(game.PT{Value: player.Life})
	default:
		return res
	}
	if player.Life != characteristic &&
		playerRuleEffectActive(r.game, playerID, game.RuleEffectLifeTotalCantChange) {
		return res
	}
	if characteristic > player.Life && !canGainLife(r.game, playerID) {
		return res
	}
	if !applyTypedContinuousEffects(
		r.game,
		r.obj,
		source,
		[]game.ContinuousEffect{effect},
		game.DurationPermanent,
	) {
		return res
	}
	if characteristic > player.Life {
		res.amount = gainLife(r.game, playerID, characteristic-player.Life)
	} else if characteristic < player.Life {
		res.amount = loseLife(r.game, playerID, player.Life-characteristic)
	}
	res.succeeded = true
	return res
}

func handlePlayerLosesGame(r *effectResolver, prim game.PlayerLosesGame) effectResolved {
	res := effectResolved{accepted: true}
	playerID, ok := r.resolvePlayer(prim.Player)
	if !ok {
		return res
	}
	if player, ok := playerByID(r.game, playerID); ok && player.Eliminated {
		return res
	}
	r.game.MarkedToLoseGame[playerID] = true
	res.succeeded = true
	return res
}

func handlePlayerWinsGame(r *effectResolver, prim game.PlayerWinsGame) effectResolved {
	res := effectResolved{accepted: true}
	winnerID, ok := r.resolvePlayer(prim.Player)
	if !ok {
		return res
	}
	if player, ok := playerByID(r.game, winnerID); ok && player.Eliminated {
		return res
	}
	res.succeeded = markPlayerWinsGame(r.game, winnerID)
	return res
}

// markPlayerWinsGame marks every other still-active player to lose the game so
// the named player wins (CR 104.2a). It returns whether any opponent was marked.
func markPlayerWinsGame(g *game.Game, winnerID game.PlayerID) bool {
	if player, ok := playerByID(g, winnerID); ok && player.Eliminated {
		return false
	}
	marked := false
	for _, player := range g.Players {
		if player.ID == winnerID || player.Eliminated {
			continue
		}
		g.MarkedToLoseGame[player.ID] = true
		marked = true
	}
	return marked
}

func handleUntap(r *effectResolver, prim game.Untap) effectResolved {
	res := effectResolved{accepted: true}
	if prim.ChooseUpTo {
		for _, permanent := range r.chooseUntapPermanents(prim) {
			setPermanentTapped(r.game, permanent, false)
			res.succeeded = true
		}
		return res
	}
	if prim.Group.Valid() {
		res.succeeded = setPermanentsTappedSimultaneously(r.game, r.groupPermanents(prim.Group), false)
		return res
	}
	if permanent, ok := r.resolveObject(prim.Object); ok {
		setPermanentTapped(r.game, permanent, false)
		res.succeeded = true
	}
	return res
}

// chooseOneGroupPermanent prompts the resolving controller to choose exactly one
// permanent from the group ("a creature you control"). It returns false when the
// group has no member.
func (r *effectResolver) chooseOneGroupPermanent(group game.GroupReference) (*game.Permanent, bool) {
	candidates := r.groupPermanents(group)
	if len(candidates) == 0 {
		return nil, false
	}
	options := make([]game.ChoiceOption, len(candidates))
	for i, permanent := range candidates {
		options[i] = game.ChoiceOption{
			Index: i,
			Label: permanentChoiceLabel(r.game, permanent),
			Card:  permanentChoiceInfo(r.game, permanent),
		}
	}
	selected := r.engine.chooseChoice(r.game, r.agents, game.ChoiceRequest{
		Kind:             game.ChoiceResolution,
		Player:           r.obj.Controller,
		Prompt:           "Choose a permanent",
		Options:          options,
		MinChoices:       1,
		MaxChoices:       1,
		DefaultSelection: firstChoiceIndices(1),
	}, r.log)
	for _, idx := range selected {
		if idx >= 0 && idx < len(candidates) {
			return candidates[idx], true
		}
	}
	return nil, false
}

func (r *effectResolver) chooseUntapPermanents(prim game.Untap) []*game.Permanent {
	amount := r.quantity(prim.Amount)
	if amount <= 0 {
		return nil
	}
	candidates := r.groupPermanents(prim.Group)
	maxChoices := min(amount, len(candidates))
	if maxChoices == 0 {
		return nil
	}
	chooser := r.obj.Controller
	if prim.Chooser.Kind() != game.PlayerReferenceNone {
		playerID, ok := r.resolvePlayer(prim.Chooser)
		if !ok {
			return nil
		}
		chooser = playerID
	}
	options := make([]game.ChoiceOption, len(candidates))
	for i, permanent := range candidates {
		options[i] = game.ChoiceOption{
			Index: i,
			Label: permanentChoiceLabel(r.game, permanent),
			Card:  permanentChoiceInfo(r.game, permanent),
		}
	}
	selected := r.engine.chooseChoice(r.game, r.agents, game.ChoiceRequest{
		Kind:             game.ChoiceResolution,
		Player:           chooser,
		Prompt:           "Choose permanents to untap",
		Options:          options,
		MinChoices:       0,
		MaxChoices:       maxChoices,
		DefaultSelection: firstChoiceIndices(maxChoices),
	}, r.log)
	chosen := make([]*game.Permanent, 0, len(selected))
	for _, idx := range selected {
		if idx >= 0 && idx < len(candidates) {
			chosen = append(chosen, candidates[idx])
		}
	}
	return chosen
}

func handleSkipNextUntap(r *effectResolver, prim game.SkipNextUntap) effectResolved {
	res := effectResolved{accepted: true}
	targets := r.resolveObjectGroup(prim.Object, prim.Group)
	if !targets.single {
		for _, permanent := range targets.permanents {
			permanent.Exerted = true
			res.succeeded = true
		}
		return res
	}
	if targets.resolved {
		targets.permanents[0].Exerted = true
		res.succeeded = true
	}
	return res
}

func handleRemoveFromCombat(r *effectResolver, prim game.RemoveFromCombat) effectResolved {
	res := effectResolved{accepted: true}
	if permanent, ok := r.resolveObject(prim.Object); ok {
		removePermanentFromCombat(r.game, permanent.ObjectID)
		res.succeeded = true
	}
	return res
}

func handleProliferate(r *effectResolver, prim game.Proliferate) effectResolved {
	amount := r.quantity(prim.Amount)
	if amount <= 0 {
		amount = 1
	}
	res := effectResolved{accepted: true, amount: amount}
	for range amount {
		if r.engine.resolveProliferate(r.game, r.obj, r.agents, r.log) {
			res.succeeded = true
		}
	}
	return res
}

func handleExplore(r *effectResolver, prim game.Explore) effectResolved {
	res := effectResolved{accepted: true}
	permanent, ok := r.resolveObject(prim.Creature)
	if !ok {
		return res
	}
	controller := effectiveController(r.game, permanent)
	if r.engine.exploreCreature(r.game, r.obj, r.agents, r.log, controller, permanent) {
		res.succeeded = true
	}
	return res
}

func handleGoad(r *effectResolver, prim game.Goad) effectResolved {
	res := effectResolved{accepted: true}
	targets := r.resolveObjectGroup(prim.Object, prim.Group)
	if !targets.single {
		for _, permanent := range targets.permanents {
			if permanentHasType(r.game, permanent, types.Creature) {
				goadPermanent(r.game, permanent, r.obj.Controller, prim.RestOfGame)
				res.succeeded = true
			}
		}
		return res
	}
	if targets.resolved && permanentHasType(r.game, targets.permanents[0], types.Creature) {
		goadPermanent(r.game, targets.permanents[0], r.obj.Controller, prim.RestOfGame)
		res.succeeded = true
	}
	return res
}

func handleRemoveCounter(r *effectResolver, prim game.RemoveCounter) effectResolved {
	if prim.AllKinds {
		return handleRemoveAllCounters(r, prim)
	}
	res := effectResolved{accepted: true, amount: r.quantity(prim.Amount)}
	if res.amount <= 0 {
		return res
	}
	if prim.Group.Valid() {
		for _, permanent := range r.groupPermanents(prim.Group) {
			permanent.Counters.Remove(prim.CounterKind, res.amount)
			res.succeeded = true
		}
		return res
	}
	if permanent, ok := r.resolveObject(prim.Object); ok {
		kind := prim.CounterKind
		if prim.ChooseKind {
			chosen, ok := r.chooseCounterKindToMove(permanent.Counters)
			if !ok {
				return res
			}
			kind = chosen
		}
		permanent.Counters.Remove(kind, res.amount)
		res.succeeded = true
	}
	return res
}

// handleRemoveAllCounters resolves the kind-agnostic mass form "remove all
// counters from <permanent>" (Vampire Hexmage), clearing every counter of every
// kind from the referenced permanent or group regardless of count.
func handleRemoveAllCounters(r *effectResolver, prim game.RemoveCounter) effectResolved {
	res := effectResolved{accepted: true}
	if prim.Group.Valid() {
		for _, permanent := range r.groupPermanents(prim.Group) {
			res.amount += permanent.Counters.RemoveAll()
			res.succeeded = true
		}
		return res
	}
	if permanent, ok := r.resolveObject(prim.Object); ok {
		res.amount = permanent.Counters.RemoveAll()
		res.succeeded = true
	}
	return res
}

func handlePhaseOut(r *effectResolver, prim game.PhaseOut) effectResolved {
	res := effectResolved{accepted: true}
	roots := r.resolveObjectGroup(prim.Object, prim.Group).permanents
	phaseOutRoots := make([]phaseOutRoot, 0, len(roots))
	for _, permanent := range roots {
		phaseOutRoots = append(phaseOutRoots, phaseOutRoot{
			permanent:  permanent,
			phaseInFor: effectiveController(r.game, permanent),
		})
	}
	res.succeeded = phaseOutPermanentTrees(r.game, phaseOutRoots)
	return res
}

func phaseOutPermanentTree(g *game.Game, permanent *game.Permanent, phaseInFor game.PlayerID, phased map[game.ObjectID]bool) bool {
	return phaseOutPermanentTreesWithSeen(g, []phaseOutRoot{{
		permanent:  permanent,
		phaseInFor: phaseInFor,
	}}, phased)
}

type phaseOutRoot struct {
	permanent  *game.Permanent
	phaseInFor game.PlayerID
}

type phaseOutCandidate struct {
	permanent  *game.Permanent
	phaseInFor game.PlayerID
	controller game.PlayerID
	snapshot   game.ObjectSnapshot
	event      game.Event
}

func phaseOutPermanentTrees(g *game.Game, roots []phaseOutRoot) bool {
	return phaseOutPermanentTreesWithSeen(g, roots, make(map[game.ObjectID]bool))
}

func phaseOutPermanentTreesWithSeen(g *game.Game, roots []phaseOutRoot, phased map[game.ObjectID]bool) bool {
	var candidates []phaseOutCandidate
	for _, root := range roots {
		collectPhaseOutPermanentTree(g, root.permanent, root.phaseInFor, phased, &candidates)
	}
	if len(candidates) == 0 {
		return false
	}
	normalizePhaseOutAttachmentSchedules(candidates)

	g.BeginStaticSourceFrame()
	for i := range candidates {
		candidate := &candidates[i]
		candidate.controller = effectiveController(g, candidate.permanent)
		candidate.snapshot = snapshotPermanent(g, candidate.permanent, zone.Battlefield)
	}
	g.EndStaticSourceFrame()

	var simultaneousID id.ID
	if len(candidates) > 1 {
		simultaneousID = g.IDGen.Next()
	}
	for i := range candidates {
		candidate := &candidates[i]
		event := game.Event{
			Kind:           game.EventPermanentPhasedOut,
			Controller:     candidate.controller,
			Player:         candidate.phaseInFor,
			PermanentID:    candidate.permanent.ObjectID,
			CardID:         candidate.permanent.CardInstanceID,
			SimultaneousID: simultaneousID,
		}
		event.TriggeredAbilities = captureEventTriggeredAbilities(g, event)
		event.TriggeredAbilitiesCaptured = true
		candidate.event = event
	}

	for i := range candidates {
		candidate := &candidates[i]
		rememberLastKnown(g, &candidate.snapshot)
		candidate.permanent.PhasedOut = true
		candidate.permanent.PhasedOutFor = candidate.phaseInFor
		candidate.permanent.PhaseInScheduled = true
		removePermanentFromCombat(g, candidate.permanent.ObjectID)
	}
	for i := range candidates {
		emitEvent(g, candidates[i].event)
	}
	return true
}

func normalizePhaseOutAttachmentSchedules(candidates []phaseOutCandidate) {
	byID := make(map[game.ObjectID]*phaseOutCandidate, len(candidates))
	for i := range candidates {
		byID[candidates[i].permanent.ObjectID] = &candidates[i]
	}
	var inheritSchedule func(*phaseOutCandidate, game.PlayerID)
	inheritSchedule = func(candidate *phaseOutCandidate, phaseInFor game.PlayerID) {
		candidate.phaseInFor = phaseInFor
		for _, attachmentID := range candidate.permanent.Attachments {
			if attachment := byID[attachmentID]; attachment != nil {
				inheritSchedule(attachment, phaseInFor)
			}
		}
	}
	for i := range candidates {
		candidate := &candidates[i]
		if candidate.permanent.AttachedTo.Exists && byID[candidate.permanent.AttachedTo.Val] != nil {
			continue
		}
		inheritSchedule(candidate, candidate.phaseInFor)
	}
}

func collectPhaseOutPermanentTree(
	g *game.Game,
	permanent *game.Permanent,
	phaseInFor game.PlayerID,
	phased map[game.ObjectID]bool,
	candidates *[]phaseOutCandidate,
) {
	if permanent == nil || permanent.PhasedOut || phased[permanent.ObjectID] {
		return
	}
	phased[permanent.ObjectID] = true
	*candidates = append(*candidates, phaseOutCandidate{
		permanent:  permanent,
		phaseInFor: phaseInFor,
	})
	for _, attachmentID := range permanent.Attachments {
		if attachment, ok := permanentByObjectID(g, attachmentID); ok {
			collectPhaseOutPermanentTree(g, attachment, phaseInFor, phased, candidates)
		}
	}
}

func handleRegenerate(r *effectResolver, prim game.Regenerate) effectResolved {
	res := effectResolved{accepted: true}
	if prim.Group.Valid() {
		permanents := r.groupPermanents(prim.Group)
		for _, permanent := range permanents {
			permanent.RegenerationShields++
		}
		res.succeeded = len(permanents) > 0
		res.amount = len(permanents)
		return res
	}
	if permanent, ok := r.resolveObject(prim.Object); ok {
		permanent.RegenerationShields++
		res.succeeded = true
	}
	return res
}

// handleBecomeCopy makes the source permanent become a copy of the targeted
// permanent (CR 706) by registering a LayerCopy continuous effect on the source.
// RetainsThisAbility keeps the source's own become-a-copy ability so it can copy
// again (Thespian's Stage), AddKeywords applies "except it has <keyword>" riders,
// and UntilEndOfTurn limits the copy to end of turn (Mirage Mirror); otherwise
// the copy lasts for as long as the source remains on the battlefield.
func handleBecomeCopy(r *effectResolver, prim game.BecomeCopy) effectResolved {
	res := effectResolved{accepted: true}
	g := r.game
	source, ok := sourcePermanent(g, r.obj)
	if !ok {
		return res
	}
	def, ok := becomeCopyTargetDef(g, r, prim)
	if !ok {
		return res
	}
	values := copyableValuesFromDef(def)
	for _, keyword := range prim.AddKeywords {
		body, ok := game.KeywordStaticBody(keyword)
		if !ok {
			continue
		}
		values.Abilities = append(values.Abilities, &body)
	}
	if prim.RetainsThisAbility {
		if sourceDef, ok := permanentCopyDef(g, source); ok {
			values.Abilities = append(values.Abilities, becomeCopyRetainedAbilities(r, sourceDef)...)
		}
	}
	controller := effectiveController(g, source)
	duration := game.DurationForAsLongAsSourceOnBattlefield
	if prim.UntilEndOfTurn {
		duration = game.DurationUntilEndOfTurn
	}
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID:               g.IDGen.Next(),
		SourceObjectID:   source.ObjectID,
		SourceCardID:     source.CardInstanceID,
		Controller:       controller,
		Timestamp:        source.Timestamp(),
		Duration:         duration,
		CreatedTurn:      g.Turn.TurnNumber,
		AffectedObjectID: source.ObjectID,
		Layer:            game.LayerCopy,
		CopyValues:       opt.Val(values),
	})
	res.succeeded = true
	return res
}

// becomeCopyTargetDef resolves the copiable card definition the source becomes a
// copy of. The copy target is either a battlefield permanent (Thespian's Stage,
// Mirage Mirror) or a card in a non-battlefield zone such as a graveyard
// (Shifting Woodland), distinguished by which of the primitive's Object or Card
// reference is set.
func becomeCopyTargetDef(g *game.Game, r *effectResolver, prim game.BecomeCopy) (*game.CardDef, bool) {
	if prim.Card.Kind != game.CardReferenceNone {
		cardID, _, ok := resolveCardReference(g, r.obj, prim.Card)
		if !ok {
			return nil, false
		}
		_, def, ok := cardInstanceFaceDef(g, cardID, game.FaceFront)
		return def, ok
	}
	target, ok := r.resolveObject(prim.Object)
	if !ok {
		return nil, false
	}
	return permanentCopyDef(g, target)
}

// becomeCopyRetainedAbilities returns the ability an "except it has this ability"
// copy keeps: the ability that produced the become-a-copy effect ("this
// ability"). When a triggered ability is resolving (Court of Vantress's upkeep
// ability), that exact ability rides on the stack object, so the copy retains
// the whole triggered ability. Otherwise "this ability" is the source's
// activated become-a-copy ability (Thespian's Stage), which lets the copy become
// a copy again.
func becomeCopyRetainedAbilities(r *effectResolver, def *game.CardDef) []game.Ability {
	if r.obj.Kind == game.StackTriggeredAbility && r.obj.InlineTrigger != nil {
		return []game.Ability{r.obj.InlineTrigger}
	}
	var abilities []game.Ability
	for i := range def.ActivatedAbilities {
		ability := &def.ActivatedAbilities[i]
		if activatedAbilityHasBecomeCopy(ability) {
			abilities = append(abilities, ability)
		}
	}
	return abilities
}

// activatedAbilityHasBecomeCopy reports whether an activated ability's instruction
// sequence contains a BecomeCopy primitive.
func activatedAbilityHasBecomeCopy(ability *game.ActivatedAbility) bool {
	for i := range ability.Content.Modes {
		mode := &ability.Content.Modes[i]
		for j := range mode.Sequence {
			if mode.Sequence[j].Primitive != nil && mode.Sequence[j].Primitive.Kind() == game.PrimitiveBecomeCopy {
				return true
			}
		}
	}
	return false
}

func handleAttach(r *effectResolver, prim game.Attach) effectResolved {
	res := effectResolved{accepted: true}
	attachment, ok := r.resolveObject(prim.Attachment)
	if !ok {
		return res
	}
	target, ok := r.resolveObject(prim.Target)
	if !ok {
		return res
	}
	if attachPermanent(r.game, attachment, target) {
		res.succeeded = true
	}
	return res
}

func handleSkipStep(r *effectResolver, prim game.SkipStep) effectResolved {
	res := effectResolved{accepted: true}
	if playerID, ok := r.resolvePlayer(prim.Player); ok {
		scheduleSkipStep(r.game, playerID, prim.Step)
		res.succeeded = true
	}
	return res
}

func handleAddExtraPhases(r *effectResolver, prim game.AddExtraPhases) effectResolved {
	if prim.Beginning {
		r.game.Turn.ExtraPhases = append(r.game.Turn.ExtraPhases, game.PhaseBeginning)
	}
	if prim.Combat {
		r.game.Turn.ExtraPhases = append(r.game.Turn.ExtraPhases, game.PhaseCombat)
	}
	if prim.Main {
		r.game.Turn.ExtraPhases = append(r.game.Turn.ExtraPhases, game.PhasePostcombatMain)
	}
	return effectResolved{accepted: true, succeeded: true}
}

func handleRollDie(r *effectResolver, prim game.RollDie) effectResolved {
	if prim.Sides < 2 {
		return effectResolved{accepted: true, succeeded: false}
	}
	roll := r.engine.rng.IntN(prim.Sides) + 1
	return effectResolved{accepted: true, succeeded: true, amount: roll}
}

func handleCreateEmblem(r *effectResolver, prim game.CreateEmblem) effectResolved {
	r.game.Emblems = append(r.game.Emblems, game.Emblem{
		Owner:     r.obj.Controller,
		Abilities: append([]game.Ability(nil), prim.EmblemAbilities...),
	})
	return effectResolved{accepted: true, succeeded: true}
}

func handleCreateDelayedTrigger(r *effectResolver, prim game.CreateDelayedTrigger) effectResolved {
	return effectResolved{accepted: true, succeeded: scheduleDelayedTrigger(r.game, r.obj, &prim.Trigger)}
}

func handleCreateReflexiveTrigger(r *effectResolver, prim game.CreateReflexiveTrigger) effectResolved {
	return effectResolved{accepted: true, succeeded: queueReflexiveTrigger(r.game, r.obj, &prim.Trigger)}
}

func handleCreateReplacement(r *effectResolver, prim game.CreateReplacement) effectResolved {
	replacement := *prim.Replacement
	replacement.ID = r.game.IDGen.Next()
	replacement.Controller = r.obj.Controller
	replacement.SourceCardID, replacement.SourceObjectID = damageSourceIDs(r.game, r.obj)
	replacement.CreatedTurn = r.game.Turn.TurnNumber
	if prim.Duration != game.DurationPermanent {
		replacement.Duration = prim.Duration
	}
	if prim.Object.Kind() == game.ObjectReferenceEventStackObject {
		cardID, ok := triggeringSpellCardID(r, prim.Object)
		if !ok {
			return effectResolved{accepted: true, succeeded: false}
		}
		replacement.AffectedCardID = cardID
	} else if prim.Object.Kind() != game.ObjectReferenceNone {
		permanent, ok := r.resolveObject(prim.Object)
		if !ok || permanent == nil {
			return effectResolved{accepted: true, succeeded: false}
		}
		replacement.AffectedObjectID = permanent.ObjectID
	}
	r.game.ReplacementEffects = append(r.game.ReplacementEffects, replacement)
	return effectResolved{accepted: true, succeeded: true}
}

// triggeringSpellCardID resolves the stable card instance ID of the spell named
// by an event-stack-object reference ("that creature"/"that spell") on a
// spell-cast trigger. A future-cast enters-with-counters replacement binds to
// this card ID rather than an object ID because a permanent spell gains a fresh
// object ID as it resolves onto the battlefield, so only the preserved card
// instance ID identifies the entering permanent. It fails closed when the
// reference does not denote a card-backed spell still on the stack.
func triggeringSpellCardID(r *effectResolver, ref game.ObjectReference) (id.ID, bool) {
	stackObjectID, ok := copyStackObjectSourceID(r.game, r.obj, ref)
	if !ok {
		return 0, false
	}
	stackObject, ok := stackObjectByID(r.game, stackObjectID)
	if !ok || stackObject.Kind != game.StackSpell || stackObject.SourceID == 0 {
		return 0, false
	}
	return stackObject.SourceID, true
}

func applyTypedContinuousEffects(g *game.Game, obj *game.StackObject, permanent *game.Permanent, templates []game.ContinuousEffect, duration game.EffectDuration) bool {
	if len(templates) == 0 {
		return false
	}
	// Source-tied durations require the resolving ability to come from a
	// battlefield permanent (an activated or triggered ability).  Spell sources
	// are not permanents, so a source-tied duration would be immediately stale.
	// Fail closed here rather than silently creating a permanent-duration effect.
	if duration == game.DurationForAsLongAsSourceOnBattlefield ||
		duration == game.DurationForAsLongAsYouControlSource {
		if obj.Kind != game.StackActivatedAbility && obj.Kind != game.StackTriggeredAbility {
			return false
		}
		if _, ok := permanentByObjectID(g, obj.SourceID); !ok {
			return false
		}
	}
	sourceID, sourceObjectID := damageSourceIDs(g, obj)
	timestamp := game.Timestamp(g.IDGen.Next())
	applied := false
	for i := range templates {
		runtimeEffect := templates[i]
		runtimeEffect.SourceCardID = sourceID
		runtimeEffect.SourceObjectID = sourceObjectID
		runtimeEffect.Controller = obj.Controller
		runtimeEffect.Timestamp = timestamp
		runtimeEffect.CreatedTurn = g.Turn.TurnNumber
		snapshotContinuousX(g, obj, &runtimeEffect)
		if duration != game.DurationPermanent {
			runtimeEffect.Duration = duration
		}
		if runtimeEffect.Duration == game.DurationUntilYourNextTurn && runtimeEffect.ExpiresFor == game.Player1 {
			runtimeEffect.ExpiresFor = obj.Controller
		}
		if runtimeEffect.ExpiresForRef.Exists {
			expiresFor, ok := resolvePlayerReference(g, obj, runtimeEffect.ExpiresForRef.Val)
			if !ok {
				continue
			}
			runtimeEffect.ExpiresFor = expiresFor
			runtimeEffect.ExpiresForRef = opt.V[game.PlayerReference]{}
		}
		if runtimeEffect.NewController.Exists && runtimeEffect.NewController.Val == game.Player1 {
			runtimeEffect.NewController = opt.Val(obj.Controller)
		}
		if runtimeEffect.NewControllerRef.Exists {
			newController, ok := resolvePlayerReference(g, obj, runtimeEffect.NewControllerRef.Val)
			if !ok {
				continue
			}
			runtimeEffect.NewController = opt.Val(newController)
			runtimeEffect.NewControllerRef = opt.V[game.PlayerReference]{}
		}
		if runtimeEffect.Group.Valid() {
			members := newReferenceResolver(g, obj).groupMembers(runtimeEffect.Group)
			runtimeEffect.Group = game.GroupReference{}
			for _, objectID := range members {
				memberEffect := runtimeEffect
				memberEffect.ID = g.IDGen.Next()
				memberEffect.AffectedObjectID = objectID
				g.ContinuousEffects = append(g.ContinuousEffects, memberEffect)
			}
			applied = true
			continue
		}
		if runtimeEffect.AffectedObjectID == 0 {
			if permanent == nil {
				continue
			}
			runtimeEffect.AffectedObjectID = permanent.ObjectID
		}
		runtimeEffect.ID = g.IDGen.Next()
		g.ContinuousEffects = append(g.ContinuousEffects, runtimeEffect)
		applied = true
	}
	return applied
}

// snapshotContinuousX locks a one-shot continuous effect's dynamic power and
// toughness deltas and dynamic base-P/T sets to fixed values at resolution. A
// mass pump such as "Creatures you control get +X/+X until end of turn, where X
// is the number of creatures you control." computes X once when the spell or
// ability resolves, and a base-P/T set such as Mirror Entity's "creatures you
// control have base power and toughness X/X until end of turn" likewise locks X
// to the cost paid. Every dynamic kind (the spell or ability's X, a battlefield
// count, a greatest characteristic, …) is evaluated here and frozen rather than
// re-evaluated each time the continuous effect applies.
func snapshotContinuousX(g *game.Game, obj *game.StackObject, effect *game.ContinuousEffect) {
	if effect.SetPowerDynamic.Exists {
		effect.SetPower = opt.Val(game.PT{Value: dynamicAmountValue(g, obj, obj.Controller, effect.SetPowerDynamic.Val)})
		effect.SetPowerDynamic.Exists = false
	}
	if effect.SetToughnessDynamic.Exists {
		effect.SetToughness = opt.Val(game.PT{Value: dynamicAmountValue(g, obj, obj.Controller, effect.SetToughnessDynamic.Val)})
		effect.SetToughnessDynamic.Exists = false
	}
	if effect.PowerDeltaDynamic.Exists {
		effect.PowerDelta += dynamicAmountValue(g, obj, obj.Controller, effect.PowerDeltaDynamic.Val)
		effect.PowerDeltaDynamic.Exists = false
	}
	if effect.ToughnessDeltaDynamic.Exists {
		effect.ToughnessDelta += dynamicAmountValue(g, obj, obj.Controller, effect.ToughnessDeltaDynamic.Val)
		effect.ToughnessDeltaDynamic.Exists = false
	}
}
