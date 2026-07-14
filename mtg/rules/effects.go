package rules

import (
	"github.com/natefinch/council4/mtg/game/zone"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/opt"
)

func (e *Engine) resolveSpellEffects(g *game.Game, obj *game.StackObject, card *game.CardInstance, log *TurnLog) {
	e.resolveSpellEffectsWithChoices(g, obj, card, [game.NumPlayers]PlayerAgent{}, log)
}

func (e *Engine) resolveSpellEffectsWithChoices(g *game.Game, obj *game.StackObject, card *game.CardInstance, agents [game.NumPlayers]PlayerAgent, log *TurnLog) {
	if card != nil && e.resolveCardImplementationSpell(g, obj, card, agents, log) {
		return
	}
	var spellDef *game.CardDef
	if card != nil {
		spellDef = cardFaceOrDefault(card, obj.Face)
	} else {
		var ok bool
		spellDef, ok = obj.SourceTokenDef.FaceDef(obj.Face)
		if !ok {
			return
		}
	}
	ability, ok := firstSpellAbility(spellDef)
	if obj.Overloaded && spellDef.Overload.Exists {
		ability = &spellDef.Overload.Val.SpellAbility
		ok = true
	}
	if !ok {
		return
	}
	if obj.GiftPromised {
		if gift, ok := spellGift(spellDef); ok {
			e.resolveAbilityContentWithChoices(g, obj, gift.Delivery, agents, log)
		}
	}
	e.resolveAbilityContentWithChoices(g, obj, *ability, agents, log)
	if obj.KickerPaid {
		if kicker, ok := spellKicker(spellDef); ok {
			e.resolveAbilityContentWithChoices(g, obj, kicker.BonusContent, agents, log)
		}
	}
	e.resolveSplicedContent(g, obj, agents, log)
}

// resolveSplicedContent resolves the spell effects spliced onto this Arcane spell
// (CR 702.47), in the order they were spliced, after the host spell's own effects.
// Each spliced content resolves against its own captured targets: obj.Targets and
// obj.TargetCounts are temporarily swapped for the spliced entry's targets (which
// are indexed from zero, matching the spliced content's own target references) and
// restored afterward.
func (e *Engine) resolveSplicedContent(g *game.Game, obj *game.StackObject, agents [game.NumPlayers]PlayerAgent, log *TurnLog) {
	if len(obj.SplicedContent) == 0 {
		return
	}
	savedTargets := obj.Targets
	savedCounts := obj.TargetCounts
	defer func() {
		obj.Targets = savedTargets
		obj.TargetCounts = savedCounts
	}()
	for i := range obj.SplicedContent {
		if i < len(obj.SplicedTargets) {
			obj.Targets = obj.SplicedTargets[i]
		} else {
			obj.Targets = nil
		}
		if i < len(obj.SplicedTargetCounts) {
			obj.TargetCounts = obj.SplicedTargetCounts[i]
		} else {
			obj.TargetCounts = nil
		}
		e.resolveAbilityContentWithChoices(g, obj, obj.SplicedContent[i], agents, log)
	}
}

func (e *Engine) resolveAbilityContentWithChoices(g *game.Game, obj *game.StackObject, content game.AbilityContent, agents [game.NumPlayers]PlayerAgent, log *TurnLog) {
	if len(content.Modes) == 0 {
		return
	}
	if !content.IsModal() {
		for i := range content.Modes[0].Sequence {
			e.resolveInstructionWithChoices(g, obj, &content.Modes[0].Sequence[i], agents, log)
		}
		return
	}
	allTargets := obj.Targets
	defer func() {
		obj.Targets = allTargets
	}()
	for chosenIndex, modeIndex := range obj.ChosenModes {
		if modeIndex < 0 || modeIndex >= len(content.Modes) {
			continue
		}
		obj.Targets = targetsForChosenMode(content, obj, allTargets, chosenIndex)
		for i := range content.Modes[modeIndex].Sequence {
			e.resolveInstructionWithChoices(g, obj, &content.Modes[modeIndex].Sequence[i], agents, log)
		}
	}
}

func targetsForChosenMode(content game.AbilityContent, obj *game.StackObject, allTargets []game.Target, chosenIndex int) []game.Target {
	if len(obj.ChosenModes) == 1 {
		return allTargets
	}
	expectedCounts := len(content.SharedTargets)
	for _, modeIndex := range obj.ChosenModes {
		if modeIndex < 0 || modeIndex >= len(content.Modes) {
			continue
		}
		expectedCounts += len(content.Modes[modeIndex].Targets)
	}
	if expectedCounts != len(obj.TargetCounts) {
		panic("modal stack object target counts do not match chosen mode targets")
	}

	sharedTargetCount := sumTargetCounts(obj.TargetCounts[:len(content.SharedTargets)])
	countOffset := len(content.SharedTargets)
	targetOffset := sharedTargetCount
	if sumTargetCounts(obj.TargetCounts) != len(allTargets) {
		panic("modal stack object target counts do not match runtime targets")
	}
	for i, modeIndex := range obj.ChosenModes {
		if modeIndex < 0 || modeIndex >= len(content.Modes) {
			continue
		}
		nextCountOffset := countOffset + len(content.Modes[modeIndex].Targets)
		modeTargetCount := sumTargetCounts(obj.TargetCounts[countOffset:nextCountOffset])
		if i == chosenIndex {
			targets := append([]game.Target(nil), allTargets[:sharedTargetCount]...)
			return append(targets, allTargets[targetOffset:targetOffset+modeTargetCount]...)
		}
		countOffset = nextCountOffset
		targetOffset += modeTargetCount
	}
	panic("chosen mode target segment not found")
}

func sumTargetCounts(counts []int) int {
	total := 0
	for _, count := range counts {
		total += count
	}
	return total
}

func spellHasKicker(card *game.CardDef) bool {
	_, ok := spellKicker(card)
	return ok
}

func spellHasMultikicker(card *game.CardDef) bool {
	kicker, ok := spellKicker(card)
	return ok && kicker.Multi
}

func spellKicker(card *game.CardDef) (game.KickerKeyword, bool) {
	if card == nil {
		return game.KickerKeyword{}, false
	}
	return card.KickerKeyword()
}

func spellHasGift(card *game.CardDef) bool {
	_, ok := spellGift(card)
	return ok
}

// spellHasBargain reports whether a spell has the Bargain keyword (CR 702.166),
// so the rules layer offers the optional bargained cast that pays the Bargain
// additional cost and sets the resolving spell's bargained state.
func spellHasBargain(card *game.CardDef) bool {
	return card != nil && card.HasKeyword(game.Bargain)
}

// spellHasOffspring reports whether a spell has the Offspring keyword
// (CR 702.171), so the rules layer offers the optional offspring cast that pays
// the Offspring additional mana cost and sets the resolving spell's offspring
// state.
func spellHasOffspring(card *game.CardDef) bool {
	if card == nil {
		return false
	}
	_, ok := card.OffspringKeyword()
	return ok
}

func spellGift(card *game.CardDef) (game.GiftKeyword, bool) {
	if card == nil {
		return game.GiftKeyword{}, false
	}
	return card.GiftKeyword()
}

func firstSpellAbility(card *game.CardDef) (*game.AbilityContent, bool) {
	if card != nil && card.SpellAbility.Exists {
		return &card.SpellAbility.Val, true
	}
	return nil, false
}

func (e *Engine) resolveInstructionWithChoices(g *game.Game, obj *game.StackObject, instr *game.Instruction, agents [game.NumPlayers]PlayerAgent, log *TurnLog) {
	newEffectResolver(e, g, obj, agents, log).resolveInstruction(instr)
}

// effectResolver bundles the per-resolution context so the resolution body
// can be a method rather than a free function with five repeated parameters.
type effectResolver struct {
	engine             *Engine
	game               *game.Game
	obj                *game.StackObject
	agents             [game.NumPlayers]PlayerAgent
	log                *TurnLog
	currentInstruction *game.Instruction

	// groupOfferMember, when set, names the player currently being offered an
	// OptionalActorGroup instruction. It resolves PlayerReferenceGroupOfferMember
	// (the "them" of "Any player may have <source> deal N damage to them") to that
	// player while the group offer's primitive resolves for each accepting player.
	groupOfferMember opt.V[game.PlayerID]
}

func newEffectResolver(e *Engine, g *game.Game, obj *game.StackObject, agents [game.NumPlayers]PlayerAgent, log *TurnLog) *effectResolver {
	return &effectResolver{engine: e, game: g, obj: obj, agents: agents, log: log}
}

// effectResolved captures the outcome of executing one effect: whether it was
// accepted by the player (for optional effects), whether it successfully
// applied, and the computed amount (used by linked "that much" follow-ups,
// CR 608.2c).
type effectResolved struct {
	accepted     bool
	succeeded    bool
	amount       int
	excessDamage int
	// acceptedActors is the set of players who accepted a group offer or Tempting
	// offer, published so a later consequence can branch on how many (and which)
	// members accepted. It is empty for a single-decider instruction.
	acceptedActors game.PlayerSet
}

// record writes the resolution state into the stack object so that follow-up
// "if you do" / "that much" instructions see what actually happened
// (CR 608.2c; impossible actions CR 101.3).
func (res effectResolved) record(obj *game.StackObject, linkID string) {
	if res.accepted && res.succeeded {
		rememberEffectAmount(obj, linkID, res.amount)
		rememberEffectExcessDamage(obj, linkID, res.excessDamage)
	}
	rememberInstructionResolutionResult(obj, linkID, res.accepted, res.succeeded, res.amount, res.acceptedActors)
}

func recordResultKey(obj *game.StackObject, key game.ResultKey, res effectResolved) {
	if key == "" {
		return
	}
	res.record(obj, string(key))
}

func (r *effectResolver) resolveInstruction(instr *game.Instruction) {
	if instr == nil {
		return
	}
	// Envelope: evaluate conditions first.
	if !effectConditionSatisfied(r.game, r.obj, instr.Condition) {
		return
	}
	if !cardConditionSatisfied(r.game, r.obj, instr.CardCondition) {
		return
	}
	if instr.ResultGate.Exists {
		if !instructionResultGateSatisfied(r.obj, instr.ResultGate.Val) {
			return
		}
	}
	if instr.Primitive == nil {
		// A Tempting offer with a multi-primitive shared body carries no
		// top-level primitive; resolveTemptingOffer runs the body instead.
		if !instr.TemptingOffer || !instr.Optional || !instr.OptionalActorGroup.Exists || len(instr.TemptingOfferBody) == 0 {
			panic("rules: nil instruction primitive")
		}
	}
	if instr.Optional && instr.OptionalActorGroup.Exists {
		if instr.TemptingOffer {
			r.resolveTemptingOffer(instr)
			return
		}
		r.resolveGroupOffer(instr)
		return
	}
	// Optional: ask the deciding player before executing.
	accepted := true
	if instr.Optional {
		decider := stackObjectController(r.obj)
		if instr.OptionalActor.Exists {
			actor, ok := resolvePlayerReference(r.game, r.obj, instr.OptionalActor.Val)
			if !ok {
				// The deciding player no longer exists (for example the
				// affected permanent's controller has left the game), so no one
				// can choose to perform the optional effect: skip it.
				if instr.PublishResult != "" {
					recordResultKey(r.obj, instr.PublishResult, effectResolved{accepted: false})
				}
				return
			}
			decider = actor
		}
		accepted = r.engine.chooseMay(r.game, r.agents, decider, "Apply optional effect?", r.log)
	}
	if !accepted {
		if instr.PublishResult != "" {
			recordResultKey(r.obj, instr.PublishResult, effectResolved{accepted: false})
		}
		return
	}
	kind := instr.Primitive.Kind()
	handler := globalPrimitiveRegistry().dispatch(kind)
	prev := r.currentInstruction
	r.currentInstruction = instr
	defer func() {
		r.currentInstruction = prev
	}()
	res := handler(r, instr.Primitive)
	if instr.PublishResult != "" {
		recordResultKey(r.obj, instr.PublishResult, res)
	}
}

// resolveGroupOffer resolves an OptionalActorGroup instruction: every player in
// the group is offered the effect in turn, and the primitive resolves once for
// each accepting player, with GroupOfferMemberReference() bound to that player.
// It publishes accepted=true when at least one player accepted so a following
// "If no one does" (gate Accepted TriFalse) or "If a player does" (gate Accepted
// TriTrue) consequence branches on the group's collective decision; the
// published Amount is the number who accepted and AcceptedActors is the exact
// set, the generic accepted-member publication a per-accepter consequence reads.
// It models the multiplayer "may have" offer (Browbeat, Book Burning, Vexing
// Devil).
func (r *effectResolver) resolveGroupOffer(instr *game.Instruction) {
	members := newReferenceResolver(r.game, r.obj).playerGroup(instr.OptionalActorGroup.Val)
	kind := instr.Primitive.Kind()
	handler := globalPrimitiveRegistry().dispatch(kind)
	prev := r.currentInstruction
	r.currentInstruction = instr
	prevMember := r.groupOfferMember
	defer func() {
		r.currentInstruction = prev
		r.groupOfferMember = prevMember
	}()
	anyAccepted := false
	anySucceeded := false
	var accepters game.PlayerSet
	for _, member := range members {
		if !r.engine.chooseMay(r.game, r.agents, member, "Apply optional effect?", r.log) {
			continue
		}
		anyAccepted = true
		accepters = accepters.With(member)
		r.groupOfferMember = opt.Val(member)
		res := handler(r, instr.Primitive)
		if res.succeeded {
			anySucceeded = true
		}
	}
	if instr.PublishResult != "" {
		recordResultKey(r.obj, instr.PublishResult, effectResolved{
			accepted:       anyAccepted,
			succeeded:      anySucceeded,
			amount:         accepters.Count(),
			acceptedActors: accepters,
		})
	}
}

// resolveTemptingOffer resolves the "Tempting offer" ability-word idiom (the
// Tempt cycle). The controller performs the primitive once as a base ("you do
// X"), then every member of the offered group is asked in turn ("each opponent
// may do X for themselves"), and for each accepting member the controller
// performs the primitive one additional time ("for each opponent who does, you
// do X again"). The primitive addresses the acting player through
// GroupOfferMemberReference(): the runtime binds it to the controller for the
// base and reward resolutions and to each accepting member for that member's own
// resolution. PublishResult reports accepted=true when at least one member
// accepted; the published Amount is the number who accepted (which the
// controller repeat matches) and AcceptedActors is the exact set, so a future
// per-accepter consequence can read which members accepted.
func (r *effectResolver) resolveTemptingOffer(instr *game.Instruction) {
	controller := stackObjectController(r.obj)
	members := newReferenceResolver(r.game, r.obj).playerGroup(instr.OptionalActorGroup.Val)
	prev := r.currentInstruction
	r.currentInstruction = instr
	prevMember := r.groupOfferMember
	defer func() {
		r.currentInstruction = prev
		r.groupOfferMember = prevMember
	}()
	// Base: the controller performs the effect for themselves.
	r.groupOfferMember = opt.Val(controller)
	anySucceeded := r.runTemptingOfferBody(instr)
	anyAccepted := false
	var accepters game.PlayerSet
	// Each member of the group is offered the effect for themselves.
	for _, member := range members {
		if !r.engine.chooseMay(r.game, r.agents, member, "Apply optional effect?", r.log) {
			continue
		}
		anyAccepted = true
		accepters = accepters.With(member)
		r.groupOfferMember = opt.Val(member)
		if r.runTemptingOfferBody(instr) {
			anySucceeded = true
		}
	}
	// For each accepting member, the controller performs the effect again.
	for range accepters.Count() {
		r.groupOfferMember = opt.Val(controller)
		if r.runTemptingOfferBody(instr) {
			anySucceeded = true
		}
	}
	if instr.PublishResult != "" {
		recordResultKey(r.obj, instr.PublishResult, effectResolved{
			accepted:       anyAccepted,
			succeeded:      anySucceeded,
			amount:         accepters.Count(),
			acceptedActors: accepters,
		})
	}
}

// runTemptingOfferBody performs one resolution of a Tempting offer's shared
// effect body for the currently bound acting player (r.groupOfferMember). It
// runs the single Primitive when the offer carries one, or every instruction of
// TemptingOfferBody in order when the shared body is a multi-primitive sequence
// (Tempt with Bunnies's "draw a card and create a token"). It returns whether
// any part of the body did something rules-relevant. Each body instruction's
// primitive is dispatched with r.currentInstruction bound to that instruction so
// per-instruction primitive state (amounts, linked keys) resolves against it.
func (r *effectResolver) runTemptingOfferBody(instr *game.Instruction) bool {
	if len(instr.TemptingOfferBody) == 0 {
		handler := globalPrimitiveRegistry().dispatch(instr.Primitive.Kind())
		return handler(r, instr.Primitive).succeeded
	}
	outer := r.currentInstruction
	defer func() { r.currentInstruction = outer }()
	succeeded := false
	for i := range instr.TemptingOfferBody {
		body := &instr.TemptingOfferBody[i]
		r.currentInstruction = body
		handler := globalPrimitiveRegistry().dispatch(body.Primitive.Kind())
		if handler(r, body.Primitive).succeeded {
			succeeded = true
		}
	}
	return succeeded
}

func (e *Engine) drawCards(g *game.Game, playerID game.PlayerID, amount int, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	if amount <= 0 {
		return false
	}
	drew := false
	for range amount {
		if e.drawCardWithReplacements(g, playerID, agents, log, false) {
			drew = true
		}
	}
	return drew
}

// chooseSacrificePermanentsForPlayer makes playerID choose amount permanents that
// satisfy sel from the battlefield. If there are fewer or equal eligible
// permanents than amount, it chooses all of them without asking.
func (e *Engine) chooseSacrificePermanentsForPlayer(g *game.Game, resolver referenceResolver, playerID game.PlayerID, amount int, sel game.Selection, agents [game.NumPlayers]PlayerAgent, log *TurnLog) []*game.Permanent {
	var candidates []*game.Permanent
	for _, permanent := range g.Battlefield {
		if effectiveController(g, permanent) != playerID {
			continue
		}
		if !resolver.permanentMatchesGroupSelection(&sel, nil, permanent) {
			continue
		}
		if permanentCantBeSacrificed(g, permanent) {
			continue
		}
		candidates = append(candidates, permanent)
	}
	if len(candidates) == 0 || amount <= 0 {
		return nil
	}
	if len(candidates) <= amount {
		return candidates
	}
	options := make([]game.ChoiceOption, len(candidates))
	for i, permanent := range candidates {
		options[i] = game.ChoiceOption{Index: i, Label: permanentChoiceLabel(g, permanent), Card: permanentChoiceInfo(g, permanent)}
	}
	request := game.ChoiceRequest{
		Kind:             game.ChoicePayment,
		Player:           playerID,
		Prompt:           "Choose a permanent to sacrifice",
		Options:          options,
		MinChoices:       amount,
		MaxChoices:       amount,
		DefaultSelection: firstChoiceIndices(amount),
	}
	selected := e.chooseChoice(g, agents, request, log)
	chosen := make([]*game.Permanent, 0, len(selected))
	for _, idx := range selected {
		if idx >= 0 && idx < len(candidates) {
			chosen = append(chosen, candidates[idx])
		}
	}
	return chosen
}

// chooseAnyNumberToSacrificeForPlayer makes playerID choose any number (none up
// to all) of the permanents they control that satisfy sel, modeling "sacrifice
// any number of <permanents>". The controller may decline entirely, so the
// choice's minimum is zero; the chosen permanents are returned for the caller to
// sacrifice and count.
func (e *Engine) chooseAnyNumberToSacrificeForPlayer(g *game.Game, resolver referenceResolver, playerID game.PlayerID, sel game.Selection, agents [game.NumPlayers]PlayerAgent, log *TurnLog) []*game.Permanent {
	var candidates []*game.Permanent
	for _, permanent := range g.Battlefield {
		if effectiveController(g, permanent) != playerID {
			continue
		}
		if !resolver.permanentMatchesGroupSelection(&sel, nil, permanent) {
			continue
		}
		if permanentCantBeSacrificed(g, permanent) {
			continue
		}
		candidates = append(candidates, permanent)
	}
	if len(candidates) == 0 {
		return nil
	}
	options := make([]game.ChoiceOption, len(candidates))
	for i, permanent := range candidates {
		options[i] = game.ChoiceOption{Index: i, Label: permanentChoiceLabel(g, permanent), Card: permanentChoiceInfo(g, permanent)}
	}
	request := game.ChoiceRequest{
		Kind:             game.ChoicePayment,
		Player:           playerID,
		Prompt:           "Choose any number of permanents to sacrifice",
		Options:          options,
		MinChoices:       0,
		MaxChoices:       len(candidates),
		DefaultSelection: nil,
	}
	selected := e.chooseChoice(g, agents, request, log)
	chosen := make([]*game.Permanent, 0, len(selected))
	for _, idx := range selected {
		if idx >= 0 && idx < len(candidates) {
			chosen = append(chosen, candidates[idx])
		}
	}
	return chosen
}

func damageSourceIDs(g *game.Game, obj *game.StackObject) (sourceID, sourceObjectID id.ID) {
	switch obj.Kind {
	case game.StackActivatedAbility, game.StackTriggeredAbility:
		if obj.SourceCardID != 0 {
			if permanent, ok := permanentByObjectID(g, obj.SourceID); ok && permanent.CardInstanceID == obj.SourceCardID {
				return obj.SourceCardID, obj.SourceID
			}
			// Permanent has left the battlefield. Preserve obj.SourceID so that
			// protection checks can consult LKI for its last-known characteristics.
			return obj.SourceCardID, obj.SourceID
		}
		permanent, ok := permanentByObjectID(g, obj.SourceID)
		if !ok {
			return 0, obj.SourceID
		}
		return permanent.CardInstanceID, permanent.ObjectID
	default:
		// For StackSpell, include the stack object's own ID as sourceObjectID so
		// that protection checks can use the selected face via LKI even after the
		// object has been removed from the stack during resolution.
		return obj.SourceID, obj.ID
	}
}

type effectDamageSource struct {
	sourceID       id.ID
	sourceObjectID id.ID
	controller     game.PlayerID
	permanent      *game.Permanent
	deathtouch     bool
	lifelink       bool
}

func applyDamageSourceKeywordEffects(g *game.Game, source effectDamageSource, damaged *game.Permanent, damage int) {
	if damage <= 0 {
		return
	}
	if source.deathtouch {
		damaged.MarkedDeathtouchDamage = true
	}
	applyDamageSourceLifelink(g, source, damage)
}

func applyDamageSourceLifelink(g *game.Game, source effectDamageSource, damage int) {
	if damage <= 0 || !source.lifelink {
		return
	}
	if source.controller < 0 || int(source.controller) >= len(g.Players) {
		return
	}
	gainLife(g, source.controller, damage)
}

func registerPermanentReplacementEffects(g *game.Game, permanent *game.Permanent) {
	def, ok := permanentCardDef(g, permanent)
	if !ok {
		return
	}
	for i := range def.ReplacementAbilities {
		replacement := def.ReplacementAbilities[i].Replacement
		if replacement.TokenMultiplier <= 1 &&
			replacement.TokenAddend == 0 &&
			replacement.CounterMultiplier <= 1 &&
			replacement.CounterAddend == 0 &&
			replacement.DamageMultiplier <= 1 &&
			replacement.DamageAddend == 0 &&
			replacement.DamagePreventAmount == 0 &&
			replacement.LifeGainMultiplier <= 1 &&
			replacement.LifeGainAddend == 0 &&
			replacement.LifeLossMultiplier <= 1 &&
			replacement.LifeLossAddend == 0 &&
			len(replacement.CreateOneOfEachTokens) == 0 &&
			replacement.TokenReplaceDef == nil &&
			!replacement.EntersTappedOthers &&
			!replacement.EntersUntappedOthers &&
			!replacement.EntersWithCountersOthers &&
			replacement.DrawCardMultiplier <= 1 &&
			replacement.DrawCardDigLook <= 0 &&
			!replacement.DrawFromEmptyLibraryWins &&
			!replacement.DamagePreventAll &&
			!replacement.ContinuousZoneRedirect {
			continue
		}
		replacement.ID = g.IDGen.Next()
		replacement.SourceObjectID = permanent.ObjectID
		replacement.SourceCardID = permanent.CardInstanceID
		replacement.Controller = effectiveController(g, permanent)
		replacement.Duration = game.DurationPermanent
		replacement.CreatedTurn = g.Turn.TurnNumber
		if replacement.CounterRecipientSelf {
			replacement.AffectedObjectID = permanent.ObjectID
		}
		if replacement.DamageRecipientSelf {
			replacement.AffectedObjectID = permanent.ObjectID
		}
		// DamageRecipientAttached is scoped dynamically: the Equipment/Aura enters
		// unattached and attaches later without re-registration, so AffectedObjectID
		// is left unset here and the damage-replacement match resolves the recipient
		// from the source's current AttachedTo (matchingDamageReplacementEffects).
		g.ReplacementEffects = append(g.ReplacementEffects, replacement)
	}
}

// countPermanentsMatchingGroup counts battlefield permanents in a GroupReference.
func countPermanentsMatchingGroup(g *game.Game, obj *game.StackObject, controller game.PlayerID, group game.GroupReference) int {
	resolverObj := obj
	if resolverObj == nil {
		resolverObj = &game.StackObject{Controller: controller}
	}
	return len(newReferenceResolver(g, resolverObj).groupMembers(group))
}

func effectPermanentAt(g *game.Game, obj *game.StackObject, targetIndex int) (*game.Permanent, bool) {
	return effectPermanentTarget(g, obj, targetIndex)
}

func sourcePermanent(g *game.Game, obj *game.StackObject) (*game.Permanent, bool) {
	if obj == nil {
		return nil, false
	}
	if permanent, ok := permanentByObjectID(g, obj.SourceID); ok && !permanent.PhasedOut {
		return permanent, true
	}
	return nil, false
}

func firstPermanentControlledBy(g *game.Game, controller game.PlayerID) (*game.Permanent, bool) {
	for _, permanent := range g.Battlefield {
		if activeBattlefieldPermanent(permanent) && effectiveController(g, permanent) == controller {
			return permanent, true
		}
	}
	return nil, false
}

func permanentLinkedObjectRef(permanent *game.Permanent) game.LinkedObjectRef {
	if permanent.CardInstanceID == 0 {
		return game.LinkedObjectRef{}
	}
	return game.LinkedObjectRef{ObjectID: permanent.ObjectID, CardID: permanent.CardInstanceID}
}

// permanentObjectBindingRef returns a linked-object ref that preserves the
// permanent's ObjectID even for a token (CardInstanceID == 0), so an
// object-identity binding survives for a token permanent. Unlike
// permanentLinkedObjectRef, it does not require a card instance because its
// consumers resolve the captured permanent by ObjectID: the
// AddCounter.PublishLinked attacker binding resolves it while it remains on the
// battlefield, and the distributive removal payoffs (ExileForEachOpponent's draw,
// DestroyForEachPlayer's and RemoveTargetsForToken's per-controller token) resolve
// its last-known controller by ObjectID after it has left. A token has a stable
// ObjectID that is never reused, so all these bindings stay correct.
func permanentObjectBindingRef(permanent *game.Permanent) game.LinkedObjectRef {
	return game.LinkedObjectRef{ObjectID: permanent.ObjectID, CardID: permanent.CardInstanceID}
}

// returnLinkedNonBattlefieldObjects returns every linked object recorded under
// linkID from the first of returnZones that currently holds it. Each object is
// matched by its last-known-object snapshot so a stale reference can never
// resurrect a different card that happens to reuse the card id.
//
// returnZones scopes which zones a return may consult, which is a correctness
// requirement rather than an optimization: an exile-until or blink return
// (Palace Jailer, Oblivion Ring) passes {zone.Exile} only and must do nothing
// once its card has left exile — even if a same-id card now sits in the owner's
// graveyard — whereas a sacrifice-then-return effect (Heart-Shaped Herb) passes
// {zone.Graveyard} to return the card it just put there.
func returnLinkedNonBattlefieldObjects(e *Engine, g *game.Game, obj *game.StackObject, linkID string, returnZones []zone.Type, controllerOverride opt.V[game.PlayerID], options permanentCreationOptions, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	key := linkedObjectSourceKey(g, obj, linkID)
	returned := false
	for _, ref := range linkedObjects(g, key) {
		if snapshot, ok := lastKnownObject(g, ref.ObjectID); !ok || snapshot.CardID != ref.CardID {
			continue
		}
		card, ok := g.GetCardInstance(ref.CardID)
		if !ok {
			continue
		}
		owner, ok := playerByID(g, card.Owner)
		if !ok {
			continue
		}
		fromZone, ok := removeLinkedCardFromZones(owner, ref.CardID, returnZones)
		if !ok {
			continue
		}
		controller := card.Owner
		if controllerOverride.Exists {
			controller = controllerOverride.Val
		}
		if _, ok := createCardPermanentFaceWithOptions(e, g, card, controller, fromZone, game.FaceFront, nil, options, agents, log); ok {
			returned = true
		}
	}
	clearLinkedObjects(g, key)
	return returned
}

// removeLinkedCardFromZones removes cardID from the first of zones that holds it
// and reports that zone so the re-entry uses the correct origin for CR 603/614
// zone-change events. It only consults the given zones, so a caller that permits
// exile alone never reanimates a card that has moved on to the graveyard.
func removeLinkedCardFromZones(owner *game.Player, cardID id.ID, zones []zone.Type) (zone.Type, bool) {
	for _, z := range zones {
		switch z {
		case zone.Exile:
			if owner.Exile.Remove(cardID) {
				return zone.Exile, true
			}
		case zone.Graveyard:
			if owner.Graveyard.Remove(cardID) {
				return zone.Graveyard, true
			}
		default:
			// Other zones cannot back a linked-source return; skip them.
		}
	}
	return zone.None, false
}
