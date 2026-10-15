package rules

import (
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

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
			_, ok := linkedCardInstance(g, linked)
			if !ok {
				continue
			}
			if linkedZone, ok := cardZone(g, linked.CardID); ok {
				return linked.CardID, linkedZone, true
			}
		}
		return 0, zone.None, false
	case game.CardReferenceTarget:
		if obj == nil || ref.TargetIndex < 0 {
			return 0, zone.None, false
		}
		wantIndex := remapCardTargetSlot(g, obj, ref.TargetIndex)
		cardTargetIndex := 0
		for _, target := range obj.Targets {
			if !targetOccupiesCardReferenceSlot(target) {
				continue
			}
			if cardTargetIndex != wantIndex {
				cardTargetIndex++
				continue
			}
			if target.Kind != game.TargetCard || target.CardID == 0 {
				return 0, zone.None, false
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
	case game.CardReferenceCaptured:
		if obj == nil || obj.CapturedCardID == 0 {
			return 0, zone.None, false
		}
		card, ok := g.GetCardInstance(obj.CapturedCardID)
		if !ok ||
			obj.CapturedCardZoneVersion == 0 ||
			card.ZoneVersion != obj.CapturedCardZoneVersion {
			return 0, zone.None, false
		}
		capturedZone, ok := cardZone(g, obj.CapturedCardID)
		return obj.CapturedCardID, capturedZone, ok
	default:
		return 0, zone.None, false
	}
}

func targetOccupiesCardReferenceSlot(target game.Target) bool {
	return target.Kind == game.TargetCard ||
		target.Kind == game.TargetDeferred &&
			target.DeferredKindSet &&
			target.DeferredKind == game.TargetCard
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

// halveRoundUp returns half of n rounded up (toward positive infinity), the
// rounding "Round up each time." applies to a copy's halved power and toughness
// (Saw in Half). For a negative n it rounds toward zero, matching ceil(n/2).
func halveRoundUp(n int) int {
	if n >= 0 {
		return (n + 1) / 2
	}
	return -((-n) / 2)
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
		if spec.HalvePowerToughnessRoundUp {
			spec.SetPower = opt.Val(game.PT{Value: halveRoundUp(resolvedObjectPower(g, &resolved))})
			spec.SetToughness = opt.Val(game.PT{Value: halveRoundUp(resolvedObjectToughness(g, &resolved))})
		}
		switch {
		case resolved.permanent != nil:
			var ok bool
			source, ok = permanentTokenCopyDef(g, resolved.permanent)
			if !ok {
				return nil, false
			}
		case resolved.snapshot.CopiableDef != nil:
			source = resolved.snapshot.CopiableDef
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
	case game.TokenCopySourceLinkedExiledCard:
		cardID, _, ok := linkedExiledCardInExile(g, obj, string(spec.LinkID))
		if !ok {
			return nil, false
		}
		card, ok := g.GetCardInstance(cardID)
		if !ok {
			return nil, false
		}
		source = cardFaceOrDefault(card, game.FaceFront)
	default:
		return nil, false
	}
	return applyTokenCopyOverrides(source, spec)
}

// applyTokenCopyOverrides copies source and applies a copy spec's explicit
// copy-modifying exceptions (name, colors, types, power/toughness, mana cost,
// printed text, legendary drop, and granted keywords). It is shared by the
// single-source copy path and the per-each group copy path, which both copy a
// resolved source CardDef before applying the same overrides.
func applyTokenCopyOverrides(source *game.CardDef, spec game.TokenCopySpec) (*game.CardDef, bool) {
	if source == nil {
		return nil, false
	}
	token := copyCardDef(source)
	if source.Back.Exists {
		back, ok := applyTokenCopyOverrides(&game.CardDef{CardFace: source.Back.Val}, spec)
		if !ok {
			return nil, false
		}
		token.Back = opt.Val(back.CardFace)
	}
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
	if len(spec.AddColors) > 0 {
		token.Colors = appendUniqueColors(token.Colors, spec.AddColors...)
	}
	if len(spec.AddTypes) > 0 {
		token.Types = appendUniqueTypes(token.Types, spec.AddTypes...)
	}
	if len(spec.AddSubtypes) > 0 {
		token.Subtypes = appendUniqueSubtypes(token.Subtypes, spec.AddSubtypes...)
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
	if spec.SetNotLegendary {
		token.Supertypes = slices.DeleteFunc(
			append([]types.Super(nil), token.Supertypes...),
			func(super types.Super) bool { return super == types.Legendary },
		)
	}
	for _, keyword := range spec.AddKeywords {
		body, ok := game.KeywordStaticBody(keyword)
		if !ok {
			return nil, false
		}
		token.StaticAbilities = append(token.StaticAbilities, body)
	}
	return token, true
}

// CR 707.8a copies both faces, with layer-1 changes applied to each face.
func permanentTokenCopyDef(g *game.Game, permanent *game.Permanent) (*game.CardDef, bool) {
	visible, ok := permanentCopyDef(g, permanent)
	if !ok || permanent.FaceDown {
		return visible, ok
	}
	physical, ok := physicalPermanentDef(g, permanent)
	if !ok || !physical.Back.Exists || physical.Layout == game.LayoutMeld {
		return visible, true
	}
	frontSource, backSource := *permanent, *permanent
	frontSource.Face, backSource.Face = game.FaceFront, game.FaceBack
	front, frontOK := permanentCopyDef(g, &frontSource)
	back, backOK := permanentCopyDef(g, &backSource)
	if !frontOK || !backOK {
		return nil, false
	}
	result := copyCardDef(physical)
	result.CardFace = front.CardFace
	result.Back = opt.Val(back.CardFace)
	result.Layout = game.LayoutDoubleFacedToken
	return result, true
}

func permanentCopyDef(g *game.Game, permanent *game.Permanent) (*game.CardDef, bool) {
	if permanent.FaceDown {
		def := faceDownCopyDef(permanent)
		if permanent.FaceDownKind == game.FaceDownDisguise || permanent.FaceDownKind == game.FaceDownCloak {
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
			if component.FaceDownKind == game.FaceDownDisguise || component.FaceDownKind == game.FaceDownCloak {
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
	applyPermanentCopyLayer(g, permanent, copied)
	return copied, true
}

func faceDownCopyDef(permanent *game.Permanent) *game.CardDef {
	pt := game.PT{Value: 2}
	face := game.CardFace{
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(pt),
		Toughness: opt.Val(pt),
	}
	if permanent.FaceDownCharacteristics.Exists {
		characteristics := permanent.FaceDownCharacteristics.Val
		face.Name = characteristics.Name
		face.Colors = slices.Clone(characteristics.Colors)
		face.Supertypes = slices.Clone(characteristics.Supertypes)
		if len(characteristics.Types) > 0 {
			face.Types = slices.Clone(characteristics.Types)
		}
		face.Subtypes = slices.Clone(characteristics.Subtypes)
		face.Power = opt.Val(characteristics.Power)
		face.Toughness = opt.Val(characteristics.Toughness)
	}
	return &game.CardDef{CardFace: face}
}

func effectivePermanentManaValue(g *game.Game, permanent *game.Permanent) (int, bool) {
	// CR 202.3b / 712.8e: real transforming permanents and full double-faced
	// tokens use their front mana cost. An individual-face copy does not.
	if permanent.Face == game.FaceBack && !permanent.FaceDown && !permanentHasCopyLayer(g, permanent) {
		if physical, ok := physicalPermanentDef(g, permanent); ok && physical.IsTransformingDoubleFaced() {
			return physical.ManaValue(), true
		}
	}
	def, ok := permanentCopyDef(g, permanent)
	if !ok {
		return 0, false
	}
	return def.ManaValue(), true
}

// applyPermanentCopyLayer overlays active layer-1 copy effects onto a copied card
// definition. This makes subsequent copies and copy tokens snapshot the
// permanent's current copiable values rather than its underlying physical card.
func applyPermanentCopyLayer(g *game.Game, permanent *game.Permanent, copied *game.CardDef) {
	if g == nil || permanent == nil || copied == nil {
		return
	}
	for _, effect := range permanentCopyEffects(g, permanent) {
		applyCopyValuesToCardDef(copied, &effect.CopyValues.Val)
	}
}

func permanentCopyEffects(g *game.Game, permanent *game.Permanent) []game.ContinuousEffect {
	var effects []game.ContinuousEffect
	for i := range g.ContinuousEffects {
		effect := &g.ContinuousEffects[i]
		if effect.Layer != game.LayerCopy ||
			effect.AffectedObjectID != permanent.ObjectID ||
			!effect.CopyValues.Exists {
			continue
		}
		effects = append(effects, *effect)
	}
	return orderContinuousEffects(effects)
}

func permanentHasCopyLayer(g *game.Game, permanent *game.Permanent) bool {
	if g == nil || permanent == nil {
		return false
	}
	for i := range g.ContinuousEffects {
		effect := &g.ContinuousEffects[i]
		if effect.Layer == game.LayerCopy &&
			effect.AffectedObjectID == permanent.ObjectID &&
			effect.CopyValues.Exists {
			return true
		}
	}
	return false
}

func applyCopyValuesToCardDef(def *game.CardDef, values *game.CopyableValues) {
	if def == nil || values == nil {
		return
	}
	def.Name = values.Name
	def.ManaCost = cloneManaCost(values.ManaCost)
	def.Colors = append([]color.Color(nil), values.Colors...)
	def.Supertypes = append([]types.Super(nil), values.Supertypes...)
	def.Types = append([]types.Card(nil), values.Types...)
	def.Subtypes = append([]types.Sub(nil), values.Subtypes...)
	def.Power = values.Power
	def.Toughness = values.Toughness
	def.DynamicPower = values.DynamicPower
	def.DynamicToughness = values.DynamicToughness
	def.OracleText = values.OracleText
	def.SpellAbility = opt.V[game.AbilityContent]{}
	def.ActivatedAbilities = nil
	def.ManaAbilities = nil
	def.LoyaltyAbilities = nil
	def.TriggeredAbilities = nil
	def.ChapterAbilities = nil
	def.ReplacementAbilities = nil
	def.StaticAbilities = nil
	for _, ability := range values.Abilities {
		switch body := ability.(type) {
		case *game.AbilityContent:
			def.SpellAbility = opt.Val(*body)
		case *game.ActivatedAbility:
			def.ActivatedAbilities = append(def.ActivatedAbilities, *body)
		case *game.ManaAbility:
			def.ManaAbilities = append(def.ManaAbilities, *body)
		case *game.LoyaltyAbility:
			def.LoyaltyAbilities = append(def.LoyaltyAbilities, *body)
		case *game.TriggeredAbility:
			def.TriggeredAbilities = append(def.TriggeredAbilities, *body)
		case *game.ChapterAbility:
			def.ChapterAbilities = append(def.ChapterAbilities, *body)
		case *game.ReplacementAbility:
			def.ReplacementAbilities = append(def.ReplacementAbilities, *body)
		case *game.StaticAbility:
			def.StaticAbilities = append(def.StaticAbilities, *body)
		default:
		}
	}
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
	dst.Overload = src.Overload
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

func createTokenPermanent(g *game.Game, controller game.PlayerID, token *game.CardDef) (*game.Permanent, bool) {
	var engine *Engine
	var agents [game.NumPlayers]PlayerAgent
	var log *TurnLog
	if context, available := replacementChoiceContextFor(g); available {
		engine, agents, log = context.engine, context.agents, context.log
	} else {
		engine = NewEngine(nil)
	}
	created, ok := createTokenPermanentsCollectingWithChoices(engine, g, controller, token, 1, false, agents, log)
	if !ok || len(created) == 0 {
		return nil, false
	}
	return created[0], true
}

func createTokenPermanentsWithChoices(e *Engine, g *game.Game, controller game.PlayerID, token *game.CardDef, amount int, tapped bool, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	_, ok := createTokenPermanentsCollectingWithChoices(e, g, controller, token, amount, tapped, agents, log)
	return ok
}

// createTokenPermanentsCollectingWithChoices creates amount tokens for the
// controller, returning the created permanents in creation order. It shares the
// replacement-amount adjustment and simultaneous-entry grouping of
// createTokenPermanentsWithChoices and is the variant callers use when they must
// act on the freshly created tokens (for example, to put them onto the
// battlefield attacking).
func createTokenPermanentsCollectingWithChoices(e *Engine, g *game.Game, controller game.PlayerID, token *game.CardDef, amount int, tapped bool, agents [game.NumPlayers]PlayerAgent, log *TurnLog, faces ...game.FaceIndex) ([]*game.Permanent, bool) {
	return createTokenPermanentsCollectingWithEntryChoices(e, g, controller, token, amount, tapped, nil, agents, log, faces...)
}

func createTokenPermanentsCollectingAttachedWithChoices(e *Engine, g *game.Game, controller game.PlayerID, token *game.CardDef, amount int, tapped bool, attachedTo *game.Permanent, agents [game.NumPlayers]PlayerAgent, log *TurnLog, faces ...game.FaceIndex) ([]*game.Permanent, bool) {
	if attachedTo == nil {
		return nil, false
	}
	return createTokenPermanentsCollectingWithEntryChoices(e, g, controller, token, amount, tapped, attachedTo, agents, log, faces...)
}

func createTokenPermanentsCollectingWithEntryChoices(e *Engine, g *game.Game, controller game.PlayerID, token *game.CardDef, amount int, tapped bool, attachedTo *game.Permanent, agents [game.NumPlayers]PlayerAgent, log *TurnLog, faces ...game.FaceIndex) ([]*game.Permanent, bool) {
	previous, _ := g.ChoiceContext()
	e.setReplacementChoiceContext(g, agents, log)
	defer g.SetChoiceContext(previous)
	if token == nil || departedPlayer(g, controller) {
		return nil, false
	}
	face := game.FaceFront
	if len(faces) > 0 {
		face = faces[0]
	}
	plans := planTokenCreation(g, controller, token, amount, face)
	total := 0
	for _, plan := range plans {
		total += plan.count
	}
	if total <= 0 {
		return nil, false
	}
	simultaneousID := tokenCreationSimultaneousID(g, total)
	return createPlannedTokenEntries(e, g, plans, simultaneousID, tapped, attachedTo, agents, log)
}

func createPlannedTokenEntries(e *Engine, g *game.Game, plans []tokenCreationPlan, simultaneousID id.ID, tapped bool, attachedTo *game.Permanent, agents [game.NumPlayers]PlayerAgent, log *TurnLog) ([]*game.Permanent, bool) {
	var entries []preparedTokenEntry
	for _, plan := range plans {
		if departedPlayer(g, plan.controller) {
			continue
		}
		for range plan.count {
			entry, ok := prepareTokenEntry(e, g, plan.controller, plan.def, tapped, attachedTo, agents, log, plan.face)
			if !ok {
				return nil, false
			}
			entries = append(entries, entry)
		}
	}
	return commitTokenEntries(g, entries, simultaneousID), true
}

func createTokenPermanentWithChoices(e *Engine, g *game.Game, controller game.PlayerID, token *game.CardDef, agents [game.NumPlayers]PlayerAgent, log *TurnLog) (*game.Permanent, bool) {
	return createTokenPermanentWithChoicesInBatch(e, g, controller, token, 0, false, agents, log)
}

func tokenCreationSimultaneousID(g *game.Game, amount int) id.ID {
	if amount > 1 {
		return g.IDGen.Next()
	}
	return 0
}

func createTokenPermanentWithChoicesInBatch(e *Engine, g *game.Game, controller game.PlayerID, token *game.CardDef, simultaneousID id.ID, tapped bool, agents [game.NumPlayers]PlayerAgent, log *TurnLog) (*game.Permanent, bool) {
	return createTokenPermanentWithEntryChoicesInBatch(e, g, controller, token, simultaneousID, tapped, nil, agents, log)
}

func createTokenPermanentWithEntryChoicesInBatch(e *Engine, g *game.Game, controller game.PlayerID, token *game.CardDef, simultaneousID id.ID, tapped bool, attachedTo *game.Permanent, agents [game.NumPlayers]PlayerAgent, log *TurnLog, faces ...game.FaceIndex) (*game.Permanent, bool) {
	entry, ok := prepareTokenEntry(e, g, controller, token, tapped, attachedTo, agents, log, faces...)
	if !ok {
		return nil, false
	}
	created := commitTokenEntries(g, []preparedTokenEntry{entry}, simultaneousID)
	return created[0], true
}

type preparedTokenEntry struct {
	permanent  *game.Permanent
	controller game.PlayerID
	visible    *game.CardDef
	token      *game.CardDef
	attachedTo *game.Permanent
}

func prepareTokenEntry(e *Engine, g *game.Game, controller game.PlayerID, token *game.CardDef, tapped bool, attachedTo *game.Permanent, agents [game.NumPlayers]PlayerAgent, log *TurnLog, faces ...game.FaceIndex) (preparedTokenEntry, bool) {
	if token == nil || departedPlayer(g, controller) {
		return preparedTokenEntry{}, false
	}
	face := game.FaceFront
	if len(faces) > 0 {
		face = faces[0]
	}
	visible, ok := token.FaceDefView(face)
	if !ok {
		return preparedTokenEntry{}, false
	}
	objectID := g.IDGen.Next()
	permanent := &game.Permanent{
		ObjectID:      objectID,
		Owner:         controller,
		Controller:    controller,
		SummoningSick: true,
		Prepared:      visible.EntersPrepared,
		Token:         true,
		TokenDef:      token,
		Face:          face,
		Transformed:   face == game.FaceBack,
	}
	initializePermanentCounters(permanent, visible)
	initializeReadAhead(e, g, permanent, agents, log)
	applyEnterBattlefieldReplacementEffects(enterBattlefieldContext{
		engine: e,
		agents: agents,
		log:    log,
	}, g, permanent, zone.None)
	if tapped {
		permanent.Tapped = true
	}
	return preparedTokenEntry{
		permanent: permanent, controller: controller, visible: visible, token: token, attachedTo: attachedTo,
	}, true
}

func commitTokenEntries(g *game.Game, entries []preparedTokenEntry, simultaneousID id.ID) []*game.Permanent {
	created := make([]*game.Permanent, 0, len(entries))
	for _, entry := range entries {
		registerPermanentReplacementEffects(g, entry.permanent)
	}
	for _, entry := range entries {
		g.Battlefield = append(g.Battlefield, entry.permanent)
		created = append(created, entry.permanent)
	}
	for _, entry := range entries {
		permanent := entry.permanent
		if entry.attachedTo != nil {
			attachPermanent(g, permanent, entry.attachedTo)
		}
		if lore := permanent.Counters.Get(counter.Lore); lore > 0 {
			emitCounterAddedEvent(g, permanent, effectiveController(g, permanent), counter.Lore, 0, lore)
		}
		event := game.Event{
			Controller: entry.controller, Player: entry.controller,
			PermanentID: permanent.ObjectID, TokenName: entry.visible.Name, TokenDef: entry.token,
			FromZone: zone.None, ToZone: zone.Battlefield, SimultaneousID: simultaneousID,
		}
		emitZoneChangeEvent(g, event)
		event.Kind = game.EventPermanentEnteredBattlefield
		emitEvent(g, event)
		event.Kind = game.EventTokenCreated
		emitEvent(g, event)
	}
	return created
}
