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
	if card != nil && e.resolveCardImplementationSpell(g, obj, card, log) {
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
	e.resolveAbilityContentWithChoices(g, obj, *ability, agents, log)
	if obj.KickerPaid {
		if kicker, ok := spellKicker(spellDef); ok {
			e.resolveAbilityContentWithChoices(g, obj, kicker.BonusContent, agents, log)
		}
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

func spellKicker(card *game.CardDef) (game.KickerKeyword, bool) {
	if card == nil {
		return game.KickerKeyword{}, false
	}
	return card.KickerKeyword()
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
}

// record writes the resolution state into the stack object so that follow-up
// "if you do" / "that much" instructions see what actually happened
// (CR 608.2c; impossible actions CR 101.3).
func (res effectResolved) record(obj *game.StackObject, linkID string) {
	if res.accepted && res.succeeded {
		rememberEffectAmount(obj, linkID, res.amount)
		rememberEffectExcessDamage(obj, linkID, res.excessDamage)
	}
	rememberInstructionResolutionResult(obj, linkID, res.accepted, res.succeeded, res.amount)
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
		panic("rules: nil instruction primitive")
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
	handler := globalPrimitiveRegistry.dispatch(kind)
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

func (e *Engine) drawCards(g *game.Game, playerID game.PlayerID, amount int, log *TurnLog) bool {
	if amount <= 0 {
		return false
	}
	drew := false
	for range amount {
		cardID, ok := e.drawCard(g, playerID)
		drew = drew || ok
		log.addDraw(DrawLog{
			Player: playerID,
			CardID: cardID,
			Failed: !ok,
		})
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
			replacement.CounterMultiplier <= 1 &&
			replacement.CounterAddend == 0 &&
			replacement.DamageMultiplier <= 1 &&
			replacement.DamageAddend == 0 &&
			len(replacement.CreateOneOfEachTokens) == 0 &&
			!replacement.EntersTappedOthers &&
			!replacement.DrawFromEmptyLibraryWins {
			continue
		}
		replacement.ID = g.IDGen.Next()
		replacement.SourceObjectID = permanent.ObjectID
		replacement.SourceCardID = permanent.CardInstanceID
		replacement.Controller = effectiveController(g, permanent)
		replacement.Duration = game.DurationPermanent
		replacement.CreatedTurn = g.Turn.TurnNumber
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

func returnLinkedExiledObjects(e *Engine, g *game.Game, obj *game.StackObject, linkID string, controllerOverride opt.V[game.PlayerID], options permanentCreationOptions, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
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
		if !ok || !owner.Exile.Remove(ref.CardID) {
			continue
		}
		controller := card.Owner
		if controllerOverride.Exists {
			controller = controllerOverride.Val
		}
		if _, ok := createCardPermanentFaceWithOptions(e, g, card, controller, zone.Exile, game.FaceFront, nil, options, agents, log); ok {
			returned = true
		}
	}
	clearLinkedObjects(g, key)
	return returned
}
