package rules

import (
	"slices"

	"github.com/natefinch/council4/mtg/game/zone"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
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
	// Optional: ask the controller before executing.
	accepted := true
	if instr.Optional {
		accepted = r.engine.chooseMay(r.game, r.agents, stackObjectController(r.obj), "Apply optional effect?", r.log)
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

func resolveCardReference(g *game.Game, obj *game.StackObject, ref game.CardReference) (id.ID, zone.Type, bool) {
	switch ref.Kind {
	case game.CardReferenceSource:
		if obj == nil || obj.SourceCardID == 0 {
			return 0, zone.None, false
		}
		sourceZone, ok := cardZone(g, obj.SourceCardID)
		if ok && obj.SourceZone != zone.None {
			card, cardOK := g.GetCardInstance(obj.SourceCardID)
			if !cardOK || sourceZone != obj.SourceZone || card.ZoneVersion != obj.SourceZoneVersion {
				return 0, zone.None, false
			}
		}
		return obj.SourceCardID, sourceZone, ok
	case game.CardReferenceEvent:
		if obj == nil || !obj.HasTriggerEvent || obj.TriggerEvent.CardID == 0 {
			return 0, zone.None, false
		}
		card, ok := g.GetCardInstance(obj.TriggerEvent.CardID)
		if !ok ||
			obj.TriggerEvent.CardZoneVersion == 0 ||
			card.ZoneVersion != obj.TriggerEvent.CardZoneVersion {
			return 0, zone.None, false
		}
		eventZone, ok := cardZone(g, obj.TriggerEvent.CardID)
		return obj.TriggerEvent.CardID, eventZone, ok
	case game.CardReferenceLinked:
		for _, linked := range linkedObjects(g, linkedObjectSourceKey(g, obj, ref.LinkID)) {
			if linked.CardID == 0 {
				continue
			}
			if linkedZone, ok := cardZone(g, linked.CardID); ok {
				return linked.CardID, linkedZone, true
			}
		}
		return 0, zone.None, false
	case game.CardReferenceTarget:
		if obj == nil {
			return 0, zone.None, false
		}
		for _, target := range obj.Targets {
			if target.Kind != game.TargetCard || target.CardID == 0 {
				continue
			}
			card, ok := g.GetCardInstance(target.CardID)
			if !ok ||
				!target.CardZoneVersionSet ||
				card.ZoneVersion != target.CardZoneVersion {
				return 0, zone.None, false
			}
			targetZone, ok := cardZone(g, target.CardID)
			return target.CardID, targetZone, ok
		}
		return 0, zone.None, false
	default:
		return 0, zone.None, false
	}
}

func cardZone(g *game.Game, cardID id.ID) (zone.Type, bool) {
	for _, player := range g.Players {
		if player.Library.Contains(cardID) {
			return zone.Library, true
		}
		if player.Hand.Contains(cardID) {
			return zone.Hand, true
		}
		if player.Graveyard.Contains(cardID) {
			return zone.Graveyard, true
		}
		if player.Exile.Contains(cardID) {
			return zone.Exile, true
		}
		if player.CommandZone.Contains(cardID) {
			return zone.Command, true
		}
	}
	return zone.None, false
}

func buildTokenCopyDef(g *game.Game, obj *game.StackObject, spec game.TokenCopySpec) (*game.CardDef, bool) {
	var source *game.CardDef
	switch spec.Source {
	case game.TokenCopySourceSourceCard:
		cardID := stackObjectSourceID(obj)
		card, ok := g.GetCardInstance(cardID)
		if !ok {
			return nil, false
		}
		source = cardFaceOrDefault(card, game.FaceFront)
	case game.TokenCopySourceObject:
		resolved, ok := resolveObjectReference(g, obj, spec.Object)
		if !ok {
			return nil, false
		}
		switch {
		case resolved.permanent != nil:
			var ok bool
			source, ok = permanentCopyDef(g, resolved.permanent)
			if !ok {
				return nil, false
			}
		case resolved.snapshot.TokenDef != nil:
			source = resolved.snapshot.TokenDef
		case resolved.snapshot.CardID != 0:
			card, ok := g.GetCardInstance(resolved.snapshot.CardID)
			if !ok {
				return nil, false
			}
			source = cardFaceOrDefault(card, resolved.snapshot.Face)
		default:
		}
	default:
		return nil, false
	}
	token := copyCardDef(source)
	if spec.SetName != "" {
		token.Name = spec.SetName
	}
	if len(spec.SetColors) > 0 {
		token.Colors = append([]color.Color(nil), spec.SetColors...)
	}
	if len(spec.SetTypes) > 0 {
		token.Types = append([]types.Card(nil), spec.SetTypes...)
	}
	if len(spec.SetSubtypes) > 0 {
		token.Subtypes = append([]types.Sub(nil), spec.SetSubtypes...)
	}
	if spec.SetPower.Exists {
		token.Power = spec.SetPower
		token.DynamicPower = opt.V[game.DynamicValue]{}
	}
	if spec.SetToughness.Exists {
		token.Toughness = spec.SetToughness
		token.DynamicToughness = opt.V[game.DynamicValue]{}
	}
	if spec.NoManaCost {
		token.ManaCost = opt.V[cost.Mana]{}
	}
	if spec.NoPrintedText {
		token.OracleText = ""
		clearCardFaceAbilities(&token.CardFace)
	}
	return token, true
}

func permanentCopyDef(g *game.Game, permanent *game.Permanent) (*game.CardDef, bool) {
	if permanent.FaceDown {
		pt := opt.Val(game.PT{Value: 2})
		def := &game.CardDef{CardFace: game.CardFace{
			Types:     []types.Card{types.Creature},
			Power:     pt,
			Toughness: pt,
		}}
		if permanent.FaceDownKind == game.FaceDownDisguise {
			def.StaticAbilities = []game.StaticAbility{faceDownDisguiseWardBody()}
		}
		return def, true
	}
	top, ok := permanentCardDef(g, permanent)
	if !ok {
		return nil, false
	}
	copied := copyCardDef(top)
	for _, component := range permanent.MergedCards {
		if component.FaceDown {
			if component.FaceDownKind == game.FaceDownDisguise {
				copied.StaticAbilities = append(copied.StaticAbilities, faceDownDisguiseWardBody())
			}
			continue
		}
		var def *game.CardDef
		if component.TokenDef != nil {
			def, ok = component.TokenDef.FaceDef(component.Face)
		} else {
			var card *game.CardInstance
			card, ok = g.GetCardInstance(component.CardInstanceID)
			if ok {
				def, ok = cardFaceDef(card, component.Face)
			}
		}
		if !ok {
			continue
		}
		appendCardFaceAbilities(&copied.CardFace, &def.CardFace)
	}
	return copied, true
}

func appendCardFaceAbilities(dst, src *game.CardFace) {
	dst.ActivatedAbilities = append(dst.ActivatedAbilities, src.ActivatedAbilities...)
	dst.ManaAbilities = append(dst.ManaAbilities, src.ManaAbilities...)
	dst.LoyaltyAbilities = append(dst.LoyaltyAbilities, src.LoyaltyAbilities...)
	dst.TriggeredAbilities = append(dst.TriggeredAbilities, src.TriggeredAbilities...)
	dst.ChapterAbilities = append(dst.ChapterAbilities, src.ChapterAbilities...)
	dst.ReplacementAbilities = append(dst.ReplacementAbilities, src.ReplacementAbilities...)
	dst.StaticAbilities = append(dst.StaticAbilities, src.StaticAbilities...)
}

func copyCardDef(source *game.CardDef) *game.CardDef {
	copied := *source
	copied.Colors = append([]color.Color(nil), source.Colors...)
	copied.ColorIdentity = source.ColorIdentity
	copied.Supertypes = append([]types.Super(nil), source.Supertypes...)
	copied.Types = append([]types.Card(nil), source.Types...)
	copied.Subtypes = append([]types.Sub(nil), source.Subtypes...)
	copyCardFaceAbilityFields(&copied.CardFace, &source.CardFace)
	if source.Back.Exists {
		copied.Back = opt.Val(copyCardFace(&source.Back.Val))
	}
	return &copied
}

func copyCardFace(source *game.CardFace) game.CardFace {
	copied := *source
	copied.Colors = append([]color.Color(nil), source.Colors...)
	copied.Supertypes = append([]types.Super(nil), source.Supertypes...)
	copied.Types = append([]types.Card(nil), source.Types...)
	copied.Subtypes = append([]types.Sub(nil), source.Subtypes...)
	copyCardFaceAbilityFields(&copied, source)
	return copied
}

func copyCardFaceAbilityFields(dst, src *game.CardFace) {
	dst.SpellAbility = src.SpellAbility
	dst.ActivatedAbilities = append([]game.ActivatedAbility(nil), src.ActivatedAbilities...)
	dst.ManaAbilities = append([]game.ManaAbility(nil), src.ManaAbilities...)
	dst.LoyaltyAbilities = append([]game.LoyaltyAbility(nil), src.LoyaltyAbilities...)
	dst.TriggeredAbilities = append([]game.TriggeredAbility(nil), src.TriggeredAbilities...)
	dst.ReplacementAbilities = append([]game.ReplacementAbility(nil), src.ReplacementAbilities...)
	dst.StaticAbilities = append([]game.StaticAbility(nil), src.StaticAbilities...)
}

func clearCardFaceAbilities(face *game.CardFace) {
	face.ClearAbilities()
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

func stackObjectSourceIsSnow(g *game.Game, obj *game.StackObject) bool {
	permanent, ok := permanentByObjectID(g, obj.SourceID)
	return ok && permanentIsSnow(g, permanent)
}

func permanentIsSnow(g *game.Game, permanent *game.Permanent) bool {
	return permanentHasSupertype(g, permanent, types.Snow)
}

func dynamicAmountValue(g *game.Game, obj *game.StackObject, controller game.PlayerID, dynamic game.DynamicAmount) int {
	amount := 0
	switch dynamic.Kind {
	case game.DynamicAmountConstant:
		amount = dynamic.Constant
	case game.DynamicAmountX:
		if obj != nil {
			amount = obj.XValue
		}
	case game.DynamicAmountTargetPower:
		if obj == nil {
			break
		}
		if resolved, ok := resolveObjectReference(g, obj, dynamic.Object); ok && resolved.permanent != nil {
			permanent := resolved.permanent
			amount = effectivePower(g, permanent)
		}
	case game.DynamicAmountTargetToughness:
		if obj == nil {
			break
		}
		if resolved, ok := resolveObjectReference(g, obj, dynamic.Object); ok && resolved.permanent != nil {
			permanent := resolved.permanent
			if toughness, ok := effectiveToughness(g, permanent); ok {
				amount = toughness
			}
		}
	case game.DynamicAmountTargetManaValue:
		if obj == nil {
			break
		}
		if resolved, ok := resolveObjectReference(g, obj, dynamic.Object); ok && resolved.permanent != nil {
			permanent := resolved.permanent
			if def, ok := permanentCardDef(g, permanent); ok {
				amount = def.ManaValue()
			}
		}
	case game.DynamicAmountTargetCounters:
		if obj == nil {
			break
		}
		if resolved, ok := resolveObjectReference(g, obj, dynamic.Object); ok && resolved.permanent != nil {
			permanent := resolved.permanent
			amount = permanent.Counters.Get(dynamic.CounterKind)
		}
	case game.DynamicAmountControllerLife:
		if player, ok := playerByID(g, controller); ok {
			amount = player.Life
		}
	case game.DynamicAmountControllerHandSize:
		if player, ok := playerByID(g, controller); ok {
			amount = player.Hand.Size()
		}
	case game.DynamicAmountControllerGraveyardSize:
		if player, ok := playerByID(g, controller); ok {
			amount = player.Graveyard.Size()
		}
	case game.DynamicAmountCountSelector:
		amount = countPermanentsMatchingGroup(g, obj, controller, dynamic.Group)
	case game.DynamicAmountPreviousEffectResult:
		key := dynamicResultKey(dynamic)
		if obj != nil && key != "" {
			amount = obj.ResolvedAmounts[key]
		}
	case game.DynamicAmountPreviousEffectExcessDamage:
		key := dynamicResultKey(dynamic)
		if obj != nil && key != "" {
			amount = obj.ResolvedExcessDamage[key]
		}
	case game.DynamicAmountOpponentCount:
		amount = len(aliveOpponents(g, controller))
	case game.DynamicAmountEventDamage:
		if obj != nil && obj.HasTriggerEvent {
			amount = obj.TriggerEvent.Amount
		}
	case game.DynamicAmountObjectPower:
		if obj == nil {
			break
		}
		if resolved, ok := resolveObjectReference(g, obj, dynamic.Object); ok {
			amount = resolvedObjectPower(g, &resolved)
		}
	default:
	}
	multiplier := dynamic.Multiplier
	if multiplier == 0 {
		multiplier = 1
	}
	return amount * multiplier
}

func dynamicResultKey(dynamic game.DynamicAmount) string {
	return string(dynamic.ResultKey)
}

func resolvedObjectPower(g *game.Game, resolved *resolvedObjectReference) int {
	if resolved.permanent != nil {
		return effectivePower(g, resolved.permanent)
	}
	if resolved.snapshot.Power.Exists {
		return resolved.snapshot.Power.Val
	}
	return 0
}

func rememberEffectAmount(obj *game.StackObject, linkID string, amount int) {
	if linkID == "" {
		return
	}
	if obj.ResolvedAmounts == nil {
		obj.ResolvedAmounts = make(map[string]int)
	}
	obj.ResolvedAmounts[linkID] = amount
}

func rememberEffectExcessDamage(obj *game.StackObject, linkID string, excessDamage int) {
	if linkID == "" || excessDamage <= 0 {
		return
	}
	if obj.ResolvedExcessDamage == nil {
		obj.ResolvedExcessDamage = make(map[string]int)
	}
	obj.ResolvedExcessDamage[linkID] = excessDamage
}

func effectCounterSource(g *game.Game, obj *game.StackObject, source game.CounterSourceSpec) (counter.Set, *game.Permanent, bool) {
	switch source.Kind {
	case game.CounterSourceTarget:
		resolved, ok := resolveObjectReference(g, obj, source.Object)
		if !ok || resolved.permanent == nil {
			return counter.Set{}, nil, false
		}
		permanent := resolved.permanent
		return cloneCounters(permanent.Counters), permanent, true
	case game.CounterSourceEventPermanent:
		if !obj.HasTriggerEvent || obj.TriggerEvent.PermanentID == 0 {
			return counter.Set{}, nil, false
		}
		// Zone-change triggers such as "put those counters on..." use the
		// triggering permanent's current state or its last-known information if it
		// has already left the battlefield (CR 603.10, CR 122).
		if permanent, ok := permanentByObjectID(g, obj.TriggerEvent.PermanentID); ok {
			return cloneCounters(permanent.Counters), permanent, true
		}
		if snapshot, ok := lastKnownObject(g, obj.TriggerEvent.PermanentID); ok {
			return cloneCounters(snapshot.Counters), nil, true
		}
	default:
	}
	return counter.Set{}, nil, false
}

func effectConditionSatisfied(g *game.Game, obj *game.StackObject, condition opt.V[game.EffectCondition]) bool {
	if !condition.Exists {
		return true
	}
	cond := condition.Val
	if cond.PermanentType.Exists {
		resolved, ok := resolveObjectReference(g, obj, cond.Object)
		if !ok || resolved.permanent == nil {
			return false
		}
		permanent := resolved.permanent
		matches := permanentHasType(g, permanent, cond.PermanentType.Val)
		if cond.Negate {
			matches = !matches
		}
		if !matches {
			return false
		}
	}
	if !conditionSatisfied(g, conditionContext{
		controller: stackObjectController(obj),
		obj:        obj,
	}, cond.Condition) {
		return false
	}
	return true
}

func cardConditionSatisfied(g *game.Game, obj *game.StackObject, condition opt.V[game.CardCondition]) bool {
	if !condition.Exists {
		return true
	}
	cond := condition.Val
	if cond.Card.Kind != game.CardReferenceLinked || cond.Card.LinkID == "" {
		return false
	}
	for _, ref := range linkedObjects(g, linkedObjectSourceKey(g, obj, cond.Card.LinkID)) {
		if ref.CardID == 0 {
			continue
		}
		card, ok := g.GetCardInstance(ref.CardID)
		if ok && cardMatchesCondition(card.Def, condition) {
			return true
		}
	}
	return false
}

func cardMatchesCondition(card *game.CardDef, condition opt.V[game.CardCondition]) bool {
	if !condition.Exists {
		return true
	}
	if card == nil {
		return false
	}
	cond := condition.Val
	if cond.RequirePermanentCard && !card.IsPermanent() {
		return false
	}
	face := card.DefaultFace()
	for _, cardType := range cond.Types {
		if !face.HasType(cardType) {
			return false
		}
	}
	for _, supertype := range cond.Supertypes {
		if !face.HasSupertype(supertype) {
			return false
		}
	}
	if len(cond.SubtypesAny) > 0 && !slices.ContainsFunc(cond.SubtypesAny, face.HasSubtype) {
		return false
	}
	return true
}

func instructionResultGateSatisfied(obj *game.StackObject, gate game.InstructionResultGate) bool {
	if gate.Key == "" {
		return true
	}
	if obj == nil || obj.ResolutionResults == nil {
		return false
	}
	result, ok := obj.ResolutionResults[string(gate.Key)]
	if !ok {
		return false
	}
	if gate.Accepted != game.TriAny && (gate.Accepted == game.TriTrue) != result.Accepted {
		return false
	}
	if gate.Succeeded != game.TriAny && (gate.Succeeded == game.TriTrue) != result.Succeeded {
		return false
	}
	return true
}

func rememberInstructionResolutionResult(obj *game.StackObject, linkID string, accepted, succeeded bool, amount int) {
	if obj == nil || linkID == "" {
		return
	}
	if obj.ResolutionResults == nil {
		obj.ResolutionResults = make(map[string]game.InstructionResolutionResult)
	}
	obj.ResolutionResults[linkID] = game.InstructionResolutionResult{
		Accepted:  accepted,
		Succeeded: succeeded,
		Amount:    amount,
	}
}

func damageSourceIDs(g *game.Game, obj *game.StackObject) (sourceID, sourceObjectID id.ID) {
	switch obj.Kind {
	case game.StackActivatedAbility, game.StackTriggeredAbility:
		if obj.SourceCardID != 0 {
			if permanent, ok := permanentByObjectID(g, obj.SourceID); ok && permanent.CardInstanceID == obj.SourceCardID {
				return obj.SourceCardID, obj.SourceID
			}
			return obj.SourceCardID, 0
		}
		permanent, ok := permanentByObjectID(g, obj.SourceID)
		if !ok {
			return 0, obj.SourceID
		}
		return permanent.CardInstanceID, permanent.ObjectID
	default:
		return obj.SourceID, 0
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
	if permanent, ok := permanentByObjectID(g, obj.SourceID); ok {
		return permanent, true
	}
	sourceCardID := obj.SourceCardID
	if sourceCardID == 0 {
		sourceCardID = obj.SourceID
	}
	for _, permanent := range g.Battlefield {
		if permanent.CardInstanceID == sourceCardID {
			return permanent, true
		}
	}
	return nil, false
}

func firstPermanentControlledBy(g *game.Game, controller game.PlayerID) (*game.Permanent, bool) {
	for _, permanent := range g.Battlefield {
		if effectiveController(g, permanent) == controller {
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

func returnLinkedExiledObjects(e *Engine, g *game.Game, obj *game.StackObject, linkID string, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
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
		if _, ok := createCardPermanentWithChoices(e, g, card, obj.Controller, zone.Exile, agents, log); ok {
			returned = true
		}
	}
	clearLinkedObjects(g, key)
	return returned
}

func createTokenPermanent(g *game.Game, controller game.PlayerID, token *game.CardDef) (*game.Permanent, bool) {
	return createTokenPermanentWithChoices(NewEngine(nil), g, controller, token, [game.NumPlayers]PlayerAgent{}, nil)
}

func createTokenPermanentWithChoices(e *Engine, g *game.Game, controller game.PlayerID, token *game.CardDef, agents [game.NumPlayers]PlayerAgent, log *TurnLog) (*game.Permanent, bool) {
	if token == nil {
		return nil, false
	}
	objectID := g.IDGen.Next()
	permanent := &game.Permanent{
		ObjectID:      objectID,
		Owner:         controller,
		Controller:    controller,
		SummoningSick: entersSummoningSick(token),
		Prepared:      token.EntersPrepared,
		Token:         true,
		TokenDef:      token,
	}
	initializePermanentCounters(permanent, token)
	initializeReadAhead(e, g, permanent, agents, log)
	applyEnterBattlefieldReplacementEffects(enterBattlefieldContext{
		engine: e,
		agents: agents,
		log:    log,
	}, g, permanent, zone.None)
	g.Battlefield = append(g.Battlefield, permanent)
	if lore := permanent.Counters.Get(counter.Lore); lore > 0 {
		emitCounterAddedEvent(g, permanent, counter.Lore, 0, lore)
	}
	event := game.Event{
		Controller:  controller,
		Player:      controller,
		PermanentID: objectID,
		TokenName:   token.Name,
		TokenDef:    token,
		FromZone:    zone.None,
		ToZone:      zone.Battlefield,
	}
	emitZoneChangeEvent(g, event)
	event.Kind = game.EventPermanentEnteredBattlefield
	emitEvent(g, event)
	return permanent, true
}
