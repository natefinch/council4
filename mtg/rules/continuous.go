package rules

import (
	"reflect"
	"slices"
	"strings"

	"github.com/natefinch/council4/mtg/game/zone"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func ptPtr(v opt.V[game.PT]) *game.PT {
	if !v.Exists {
		return nil
	}
	return &v.Val
}

func dynamicValuePtr(v opt.V[game.DynamicValue]) *game.DynamicValue {
	if !v.Exists {
		return nil
	}
	return &v.Val
}

type permanentEffectiveValues struct {
	name       string
	oracleText string
	colors     []color.Color
	supertypes []types.Super
	types      []types.Card
	subtypes   []types.Sub
	abilities  []game.Ability
	controller game.PlayerID

	power            int
	powerOK          bool
	powerPT          *game.PT
	dynamicPower     *game.DynamicValue
	toughness        int
	toughnessOK      bool
	toughnessPT      *game.PT
	dynamicToughness *game.DynamicValue
	keywords         map[game.Keyword]bool
}

func effectivePower(g *game.Game, permanent *game.Permanent) int {
	values := effectivePermanentValues(g, permanent)
	if !values.powerOK {
		return 0
	}
	return max(0, values.power)
}

func effectiveToughness(g *game.Game, permanent *game.Permanent) (int, bool) {
	values := effectivePermanentValues(g, permanent)
	return values.toughness, values.toughnessOK
}

func hasKeyword(g *game.Game, permanent *game.Permanent, keyword game.Keyword) bool {
	return effectivePermanentValues(g, permanent).keywords[keyword]
}

func permanentHasType(g *game.Game, permanent *game.Permanent, cardType types.Card) bool {
	return slices.Contains(effectivePermanentValues(g, permanent).types, cardType)
}

func permanentHasSubtype(g *game.Game, permanent *game.Permanent, subtype types.Sub) bool {
	return slices.Contains(effectivePermanentValues(g, permanent).subtypes, subtype)
}

func permanentHasSupertype(g *game.Game, permanent *game.Permanent, supertype types.Super) bool {
	return slices.Contains(effectivePermanentValues(g, permanent).supertypes, supertype)
}

func permanentEffectiveColors(g *game.Game, permanent *game.Permanent) []color.Color {
	values := effectivePermanentValues(g, permanent)
	return append([]color.Color(nil), values.colors...)
}

func permanentEffectiveAbilities(g *game.Game, permanent *game.Permanent) []game.Ability {
	return append([]game.Ability(nil), permanentEffectiveAbilitiesView(g, permanent)...)
}

// permanentEffectiveAbilitiesView returns the cached ability slice for
// read-only internal use. Callers must not modify the slice or its elements.
func permanentEffectiveAbilitiesView(g *game.Game, permanent *game.Permanent) []game.Ability {
	values := effectivePermanentValues(g, permanent)
	return values.abilities
}

func permanentEffectiveName(g *game.Game, permanent *game.Permanent) string {
	return effectivePermanentValues(g, permanent).name
}

// frameCache holds derived values memoized for the duration of one pure-read
// static-source frame. Because a frame only ever wraps reads that do not mutate
// the game state these values depend on, every entry is valid for the whole
// frame. Cached permanentEffectiveValues must be treated as read-only by
// callers; every accessor copies out the fields it returns.
type frameCache struct {
	sources      []staticAbilitySource
	sourcesBuilt bool
	values       map[id.ID]permanentEffectiveValues
	controllers  map[id.ID]game.PlayerID
}

// frameCacheFor returns the frame cache for the current frame, creating it on
// first use, or nil when no frame is open (so callers recompute every time).
func frameCacheFor(g *game.Game) *frameCache {
	if !g.InStaticSourceFrame() {
		return nil
	}
	if v, ok := g.StaticSourceFrameValue(); ok {
		if fc, ok := v.(*frameCache); ok {
			return fc
		}
	}
	fc := &frameCache{}
	g.SetStaticSourceFrameValue(fc)
	return fc
}

func effectiveController(g *game.Game, permanent *game.Permanent) game.PlayerID {
	fc := frameCacheFor(g)
	if fc != nil {
		if controller, ok := fc.controllers[permanent.ObjectID]; ok {
			return controller
		}
	}
	values := basePermanentValues(g, permanent)
	sources := staticAbilitySources(g)
	effects := orderContinuousEffects(continuousEffectsForLayer(g, permanent, &values, game.LayerControl, sources))
	for i := range effects {
		effect := &effects[i]
		if effect.NewController.Exists {
			values.controller = effect.NewController.Val
		}
	}
	if fc != nil {
		if fc.controllers == nil {
			fc.controllers = make(map[id.ID]game.PlayerID)
		}
		fc.controllers[permanent.ObjectID] = values.controller
	}
	return values.controller
}

func effectivePermanentValues(g *game.Game, permanent *game.Permanent) permanentEffectiveValues {
	fc := frameCacheFor(g)
	if fc != nil {
		if values, ok := fc.values[permanent.ObjectID]; ok {
			return values
		}
	}
	values := basePermanentValues(g, permanent)
	baseSubtypes := append([]types.Sub(nil), values.subtypes...)
	applyContinuousLayers(g, permanent, &values)
	applyAddedBasicLandManaAbilities(&values, baseSubtypes)
	applyCounterAndTemporaryValues(permanent, &values)
	for _, keyword := range keywordCounters(permanent) {
		values.keywords[keyword] = true
	}
	if fc != nil {
		if fc.values == nil {
			fc.values = make(map[id.ID]permanentEffectiveValues)
		}
		fc.values[permanent.ObjectID] = values
	}
	return values
}

func basePermanentValues(g *game.Game, permanent *game.Permanent) permanentEffectiveValues {
	values := permanentEffectiveValues{keywords: make(map[game.Keyword]bool)}
	values.controller = permanent.Controller
	if permanent.FaceDown {
		values.types = []types.Card{types.Creature}
		if permanent.FaceDownKind == game.FaceDownDisguise {
			ward := faceDownDisguiseWardBody()
			values.abilities = []game.Ability{&ward}
			rebuildKeywords(&values)
		}
		values.power, values.powerOK = 2, true
		values.toughness, values.toughnessOK = 2, true
		return values
	}
	card, ok := permanentCardDef(g, permanent)
	if !ok {
		return values
	}
	values.name = card.Name
	values.oracleText = card.OracleText
	values.colors = append([]color.Color(nil), card.Colors...)
	values.supertypes = append([]types.Super(nil), card.Supertypes...)
	values.types = append([]types.Card(nil), card.Types...)
	values.subtypes = append([]types.Sub(nil), card.Subtypes...)
	values.abilities = make([]game.Ability, 0, card.AbilityCount())
	for i := 0; i < card.AbilityCount(); i++ {
		values.abilities = append(values.abilities, card.BodyAt(i))
	}
	for _, component := range permanent.MergedCards {
		if component.FaceDown {
			if component.FaceDownKind == game.FaceDownDisguise {
				ward := faceDownDisguiseWardBody()
				values.abilities = append(values.abilities, &ward)
			}
			continue
		}
		if component.TokenDef != nil {
			componentDef, ok := component.TokenDef.FaceDef(component.Face)
			if !ok {
				continue
			}
			for i := 0; i < componentDef.AbilityCount(); i++ {
				values.abilities = append(values.abilities, componentDef.BodyAt(i))
			}
			continue
		}
		instance, ok := g.GetCardInstance(component.CardInstanceID)
		if !ok {
			continue
		}
		componentDef, ok := cardFaceDef(instance, component.Face)
		if !ok {
			continue
		}
		for i := 0; i < componentDef.AbilityCount(); i++ {
			values.abilities = append(values.abilities, componentDef.BodyAt(i))
		}
	}
	if card.Power.Exists {
		values.powerPT = ptPtr(card.Power)
		values.dynamicPower = dynamicValuePtr(card.DynamicPower)
		values.power, values.powerOK = ptValue(g, values.controller, values.powerPT, values.dynamicPower)
	}
	if card.Toughness.Exists {
		values.toughnessPT = ptPtr(card.Toughness)
		values.dynamicToughness = dynamicValuePtr(card.DynamicToughness)
		values.toughness, values.toughnessOK = ptValue(g, values.controller, values.toughnessPT, values.dynamicToughness)
	}
	rebuildKeywords(&values)
	return values
}

func ptValue(g *game.Game, controller game.PlayerID, pt *game.PT, dynamic *game.DynamicValue) (int, bool) {
	if pt == nil {
		return 0, false
	}
	if !pt.IsStar {
		return pt.Value, true
	}
	if dynamic == nil {
		return 0, false
	}
	return dynamicValue(g, controller, dynamic), true
}

func dynamicValue(g *game.Game, controller game.PlayerID, dynamic *game.DynamicValue) int {
	if dynamic == nil {
		return 0
	}
	return dynamicValueBase(g, controller, dynamic) + dynamic.Offset
}

func dynamicValueBase(g *game.Game, controller game.PlayerID, dynamic *game.DynamicValue) int {
	switch dynamic.Kind {
	case game.DynamicValueConstant:
		return dynamic.Value
	case game.DynamicValueControllerHandSize:
		if player, ok := playerByID(g, controller); ok {
			return player.Hand.Size()
		}
	case game.DynamicValueControllerGraveyardSize:
		if player, ok := playerByID(g, controller); ok {
			return player.Graveyard.Size()
		}
	case game.DynamicValueControllerCreatureCount:
		return countControlledPermanentsWithType(g, controller, types.Creature)
	case game.DynamicValueControllerLandCount:
		return countControlledPermanentsWithType(g, controller, types.Land)
	case game.DynamicValueControllerArtifactCount:
		return countControlledPermanentsWithType(g, controller, types.Artifact)
	case game.DynamicValueAllBattlefieldCreatureCount:
		count := 0
		for _, permanent := range g.Battlefield {
			if activeBattlefieldPermanent(permanent) && basePermanentHasType(g, permanent, types.Creature) {
				count++
			}
		}
		return count
	case game.DynamicValueAllGraveyardsSize:
		return allGraveyardsCardCount(g, types.Creature, false)
	case game.DynamicValueCreatureCardsInAllGraveyards:
		return allGraveyardsCardCount(g, types.Creature, true)
	case game.DynamicValueCardTypesAmongAllGraveyards:
		return allGraveyardsCardTypeCount(g)
	case game.DynamicValueControllerCreatureCardsInGraveyard:
		return controllerGraveyardCardCount(g, controller, func(card *game.CardInstance) bool {
			return graveyardCardHasType(card, types.Creature)
		})
	case game.DynamicValueControllerInstantOrSorceryCardsInGraveyard:
		return controllerGraveyardCardCount(g, controller, func(card *game.CardInstance) bool {
			return graveyardCardHasType(card, types.Instant) || graveyardCardHasType(card, types.Sorcery)
		})
	case game.DynamicValueControllerLandCardsInGraveyard:
		return controllerGraveyardCardCount(g, controller, func(card *game.CardInstance) bool {
			return graveyardCardHasType(card, types.Land)
		})
	case game.DynamicValueControllerPermanentCardsInGraveyard:
		return controllerGraveyardCardCount(g, controller, graveyardCardIsPermanent)
	case game.DynamicValueControllerCardTypesInGraveyard:
		return controllerGraveyardCardTypeCount(g, controller)
	case game.DynamicValueControllerSubtypeCount:
		return countControlledPermanentsWithSubtype(g, controller, dynamic.Subtype)
	case game.DynamicValueControllerColorPermanentCount:
		return countControlledPermanentsWithColor(g, controller, dynamic.Color)
	case game.DynamicValueControllerBasicLandTypeCount:
		return controllerBasicLandSubtypeCount(g, controller)
	case game.DynamicValueControllerLifeTotal:
		if player, ok := playerByID(g, controller); ok {
			return player.Life
		}
	case game.DynamicValueAllPlayersHandSize:
		count := 0
		for _, player := range g.Players {
			count += player.Hand.Size()
		}
		return count
	case game.DynamicValueControllerCardsDrawnThisTurn:
		return cardsDrawnThisTurn(g, controller)
	default:
	}
	return 0
}

// allGraveyardsCardCount counts cards across every player's graveyard. When
// filterByType is true, only cards whose front (or split alternate) face has
// cardType are counted; otherwise every card is counted.
func allGraveyardsCardCount(g *game.Game, cardType types.Card, filterByType bool) int {
	count := 0
	for _, player := range g.Players {
		for _, cardID := range player.Graveyard.All() {
			card, ok := g.GetCardInstance(cardID)
			if !ok {
				continue
			}
			if !filterByType || graveyardCardHasType(card, cardType) {
				count++
			}
		}
	}
	return count
}

// allGraveyardsCardTypeCount counts the distinct card types among all cards in
// every player's graveyard (CR 208.2, Tarmogoyf).
func allGraveyardsCardTypeCount(g *game.Game) int {
	distinct := make(map[types.Card]bool)
	for _, player := range g.Players {
		for _, cardID := range player.Graveyard.All() {
			card, ok := g.GetCardInstance(cardID)
			if !ok {
				continue
			}
			for _, cardType := range graveyardCardTypes(card) {
				distinct[cardType] = true
			}
		}
	}
	return len(distinct)
}

// graveyardCardTypes returns the card types of a graveyard card's front face,
// including a split card's alternate face.
func graveyardCardTypes(card *game.CardInstance) []types.Card {
	cardTypes := append([]types.Card(nil), cardFaceOrDefault(card, game.FaceFront).Types...)
	if card.Def.Layout == game.LayoutSplit && card.Def.Alternate.Exists {
		cardTypes = append(cardTypes, card.Def.Alternate.Val.Types...)
	}
	return cardTypes
}

func graveyardCardHasType(card *game.CardInstance, cardType types.Card) bool {
	return slices.Contains(graveyardCardTypes(card), cardType)
}

func countControlledPermanentsWithType(g *game.Game, controller game.PlayerID, cardType types.Card) int {
	count := 0
	for _, permanent := range g.Battlefield {
		if activeBattlefieldPermanent(permanent) &&
			permanent.Controller == controller &&
			basePermanentHasType(g, permanent, cardType) {
			count++
		}
	}
	return count
}

// countControlledPermanentsWithSubtype counts the active permanents the given
// player controls whose printed front face has the subtype ("the number of
// Swamps you control"). It reads printed subtypes only, so it does not depend on
// continuous layers and cannot recurse into power/toughness computation.
func countControlledPermanentsWithSubtype(g *game.Game, controller game.PlayerID, subtype types.Sub) int {
	count := 0
	for _, permanent := range g.Battlefield {
		if !activeBattlefieldPermanent(permanent) || permanent.Controller != controller || permanent.FaceDown {
			continue
		}
		if card, ok := permanentCardDef(g, permanent); ok && card.HasSubtype(subtype) {
			count++
		}
	}
	return count
}

// countControlledPermanentsWithColor counts the active permanents the given
// player controls whose printed front face includes the color ("the number of
// red permanents you control"). It reads printed colors only, so it does not
// depend on continuous layers and cannot recurse into power/toughness
// computation.
func countControlledPermanentsWithColor(g *game.Game, controller game.PlayerID, c color.Color) int {
	count := 0
	for _, permanent := range g.Battlefield {
		if !activeBattlefieldPermanent(permanent) || permanent.Controller != controller || permanent.FaceDown {
			continue
		}
		if card, ok := permanentCardDef(g, permanent); ok && slices.Contains(card.Colors, c) {
			count++
		}
	}
	return count
}

// basicLandTypes lists the five basic land subtypes (CR 305.6) counted by
// "the number of basic land types among lands you control".
var basicLandTypes = []types.Sub{types.Plains, types.Island, types.Swamp, types.Mountain, types.Forest}

// controllerBasicLandSubtypeCount counts the distinct basic land subtypes
// present among the active lands the given player controls, reading printed
// subtypes only to avoid recursing into continuous layers.
func controllerBasicLandSubtypeCount(g *game.Game, controller game.PlayerID) int {
	count := 0
	for _, subtype := range basicLandTypes {
		if countControlledPermanentsWithSubtype(g, controller, subtype) > 0 {
			count++
		}
	}
	return count
}

// controllerGraveyardCardCount counts the cards in the given player's graveyard
// that satisfy match.
func controllerGraveyardCardCount(g *game.Game, controller game.PlayerID, match func(*game.CardInstance) bool) int {
	player, ok := playerByID(g, controller)
	if !ok {
		return 0
	}
	count := 0
	for _, cardID := range player.Graveyard.All() {
		card, ok := g.GetCardInstance(cardID)
		if !ok {
			continue
		}
		if match(card) {
			count++
		}
	}
	return count
}

// permanentCardTypes lists the card types that make a card a permanent card
// (CR 110.4a) for "the number of permanent cards in your graveyard".
var permanentCardTypes = []types.Card{
	types.Artifact, types.Creature, types.Enchantment, types.Land, types.Planeswalker, types.Battle,
}

func graveyardCardIsPermanent(card *game.CardInstance) bool {
	for _, cardType := range permanentCardTypes {
		if graveyardCardHasType(card, cardType) {
			return true
		}
	}
	return false
}

func basePermanentHasType(g *game.Game, permanent *game.Permanent, cardType types.Card) bool {
	if permanent.FaceDown {
		return cardType == types.Creature
	}
	card, ok := permanentCardDef(g, permanent)
	return ok && card.HasType(cardType)
}

func applyContinuousLayers(g *game.Game, permanent *game.Permanent, values *permanentEffectiveValues) {
	sources := staticAbilitySources(g)
	for _, layer := range continuousLayers {
		if layer == game.LayerType && values.keywords[game.Changeling] {
			values.subtypes = append([]types.Sub(nil), types.SubtypesForType(types.Creature)...)
		}
		effects := continuousEffectsForLayer(g, permanent, values, layer, sources)
		ordered := orderContinuousEffects(effects)
		for i := range ordered {
			applyContinuousEffect(g, permanent, values, &ordered[i])
		}
	}
}

func permanentValuesBeforeLayer(g *game.Game, permanent *game.Permanent, stop game.ContinuousLayer) permanentEffectiveValues {
	values := basePermanentValues(g, permanent)
	sources := staticAbilitySources(g)
	for _, layer := range continuousLayers {
		if layer == stop {
			break
		}
		if layer == game.LayerType && values.keywords[game.Changeling] {
			values.subtypes = append([]types.Sub(nil), types.SubtypesForType(types.Creature)...)
		}
		effects := continuousEffectsForLayer(g, permanent, &values, layer, sources)
		ordered := orderContinuousEffects(effects)
		for i := range ordered {
			applyContinuousEffect(g, permanent, &values, &ordered[i])
		}
	}
	return values
}

var continuousLayers = [...]game.ContinuousLayer{
	game.LayerCopy,
	game.LayerControl,
	game.LayerText,
	game.LayerType,
	game.LayerColor,
	game.LayerAbility,
	game.LayerPowerToughnessSet,
	game.LayerPowerToughnessModify,
	game.LayerPowerToughnessSwitch,
}

func continuousEffectsForLayer(g *game.Game, permanent *game.Permanent, values *permanentEffectiveValues, layer game.ContinuousLayer, sources []staticAbilitySource) []game.ContinuousEffect {
	var effects []game.ContinuousEffect
	for i := range g.ContinuousEffects {
		effect := &g.ContinuousEffects[i]
		if effect.Layer == layer && continuousEffectApplies(g, permanent, values, effect) {
			effects = append(effects, *effect)
		}
	}
	effects = append(effects, staticAbilityContinuousEffectsForLayer(g, permanent, values, layer, sources)...)
	return effects
}

type staticAbilitySource struct {
	permanent  *game.Permanent
	card       *game.CardDef
	cardID     id.ID
	controller game.PlayerID
	timestamp  game.Timestamp
}

func staticAbilityContinuousEffectsForLayer(g *game.Game, permanent *game.Permanent, values *permanentEffectiveValues, layer game.ContinuousLayer, sources []staticAbilitySource) []game.ContinuousEffect {
	var effects []game.ContinuousEffect
	for _, source := range sources {
		if !staticAbilityCardHasLayer(source.card, source.permanent != nil, layer) {
			continue
		}
		if source.permanent != nil && layer != game.LayerControl {
			source.controller = effectiveController(g, source.permanent)
		}
		effects = append(effects, staticAbilitySourceContinuousEffects(g, source, permanent, values, layer)...)
	}
	return effects
}

func staticAbilitySources(g *game.Game) []staticAbilitySource {
	fc := frameCacheFor(g)
	if fc != nil {
		if !fc.sourcesBuilt {
			fc.sources = buildStaticAbilitySources(g)
			fc.sourcesBuilt = true
		}
		return fc.sources
	}
	return buildStaticAbilitySources(g)
}

func buildStaticAbilitySources(g *game.Game) []staticAbilitySource {
	var sources []staticAbilitySource
	for _, permanent := range g.Battlefield {
		if !activeBattlefieldPermanent(permanent) {
			continue
		}
		visitPermanentStaticAbilityComponents(g, permanent, func(component permanentAbilityComponent) {
			if !staticAbilityCardHasContinuousEffects(component.card, true) {
				return
			}
			sources = append(sources, staticAbilitySource{
				permanent:  permanent,
				card:       component.card,
				cardID:     component.cardID,
				controller: permanent.Controller,
				timestamp:  permanent.Timestamp(),
			})
		})
	}
	for playerID := range game.PlayerID(game.NumPlayers) {
		player := g.Players[playerID]
		player.Graveyard.Range(func(cardID id.ID) bool {
			card, ok := g.GetCardInstance(cardID)
			if !ok {
				return true
			}
			component, ok := staticAbilityCardInstanceComponent(card, game.FaceFront)
			if !ok || len(component.card.StaticAbilities) == 0 ||
				!staticAbilityCardHasContinuousEffects(component.card, false) {
				return true
			}
			sources = append(sources, staticAbilitySource{
				card:       component.card,
				cardID:     card.ID,
				controller: card.Owner,
				timestamp:  game.Timestamp(card.ID),
			})
			return true
		})
	}
	return sources
}

type permanentAbilityComponent struct {
	card   *game.CardDef
	cardID id.ID
}

func visitPermanentStaticAbilityComponents(g *game.Game, permanent *game.Permanent, visit func(permanentAbilityComponent)) {
	component, ok := staticAbilityPermanentComponent(g, permanent)
	if !ok {
		return
	}
	if len(component.card.StaticAbilities) > 0 {
		visit(component)
	}
	for _, merged := range permanent.MergedCards {
		if merged.FaceDown {
			continue
		}
		if merged.TokenDef != nil {
			def, ok := merged.TokenDef.FaceDef(merged.Face)
			if ok && len(def.StaticAbilities) > 0 {
				visit(permanentAbilityComponent{card: def})
			}
			continue
		}
		instance, ok := g.GetCardInstance(merged.CardInstanceID)
		if !ok {
			continue
		}
		component, ok := staticAbilityCardInstanceComponent(instance, merged.Face)
		if ok && len(component.card.StaticAbilities) > 0 {
			visit(component)
		}
	}
}

func staticAbilitySourceContinuousEffects(g *game.Game, source staticAbilitySource, permanent *game.Permanent, values *permanentEffectiveValues, layer game.ContinuousLayer) []game.ContinuousEffect {
	var effects []game.ContinuousEffect
	for i := range source.card.StaticAbilities {
		body := &source.card.StaticAbilities[i]
		if !staticAbilityFunctionsFromSource(body, source) || !staticAbilityHasEffectForLayer(body, layer) {
			continue
		}
		if !conditionSatisfied(g, conditionContext{
			controller:            source.controller,
			source:                source.permanent,
			characteristicsBefore: layer,
		}, body.Condition) {
			continue
		}
		for i := range body.ContinuousEffects {
			template := &body.ContinuousEffects[i]
			if template.Layer != layer {
				continue
			}
			if template.AffectedSource && !template.Group.Empty() {
				continue
			}
			staticEffect := *template
			staticEffect.SourceObjectID = sourceObjectID(source)
			staticEffect.SourceCardID = source.cardID
			staticEffect.Controller = source.controller
			staticEffect.Timestamp = source.timestamp
			if staticEffect.Layer == game.LayerControl && staticEffect.NewController.Exists {
				staticEffect.NewController = opt.Val(source.controller)
			}
			if template.AffectedSource {
				if source.permanent == nil {
					continue
				}
				staticEffect.AffectedObjectID = source.permanent.ObjectID
			}
			if continuousEffectApplies(g, permanent, values, &staticEffect) {
				effects = append(effects, staticEffect)
			}
		}
	}
	return effects
}

func staticAbilityPermanentComponent(g *game.Game, permanent *game.Permanent) (permanentAbilityComponent, bool) {
	if permanent.FaceDown {
		return permanentAbilityComponent{}, false
	}
	if permanent.Token {
		if permanent.TokenDef == nil {
			return permanentAbilityComponent{}, false
		}
		if !permanent.TokenDef.Back.Exists && permanent.Face == game.FaceFront {
			return permanentAbilityComponent{card: permanent.TokenDef}, true
		}
		def, ok := permanent.TokenDef.FaceDef(permanent.Face)
		return permanentAbilityComponent{card: def}, ok
	}
	card, ok := g.GetCardInstance(permanent.CardInstanceID)
	if !ok {
		return permanentAbilityComponent{}, false
	}
	if !card.Def.Back.Exists && permanent.Face == game.FaceFront {
		return permanentAbilityComponent{
			card:   card.Def,
			cardID: permanent.CardInstanceID,
		}, true
	}
	def, ok := card.Def.FaceDef(permanent.Face)
	return permanentAbilityComponent{
		card:   def,
		cardID: permanent.CardInstanceID,
	}, ok
}

func staticAbilityCardInstanceComponent(card *game.CardInstance, face game.FaceIndex) (permanentAbilityComponent, bool) {
	if card == nil {
		return permanentAbilityComponent{}, false
	}
	def, ok := cardFaceDef(card, face)
	return permanentAbilityComponent{card: def, cardID: card.ID}, ok
}

func staticAbilityHasEffectForLayer(body *game.StaticAbility, layer game.ContinuousLayer) bool {
	if body == nil {
		return false
	}
	for i := range body.ContinuousEffects {
		if body.ContinuousEffects[i].Layer == layer {
			return true
		}
	}
	return false
}

func staticAbilityFunctionsFromSource(body *game.StaticAbility, source staticAbilitySource) bool {
	return staticAbilityFunctionsInZone(body, source.permanent != nil)
}

func staticAbilityFunctionsInZone(body *game.StaticAbility, onBattlefield bool) bool {
	if body == nil {
		return false
	}
	if onBattlefield {
		return bodyFunctionsOnBattlefield(body)
	}
	return body.ZoneOfFunction == zone.Graveyard
}

func sourceObjectID(source staticAbilitySource) id.ID {
	if source.permanent == nil {
		return 0
	}
	return source.permanent.ObjectID
}

func continuousEffectApplies(g *game.Game, permanent *game.Permanent, values *permanentEffectiveValues, effect *game.ContinuousEffect) bool {
	if effect.AffectedObjectID != 0 {
		return effect.AffectedObjectID == permanent.ObjectID
	}
	if !effect.Group.Valid() {
		return false
	}
	source, _ := permanentByObjectID(g, effect.SourceObjectID)
	resolver := newReferenceResolverWithSource(g, &game.StackObject{Controller: effect.Controller}, source)
	switch effect.Group.Domain() {
	case game.GroupDomainAttachedObject:
		anchor, ok := effect.Group.Anchor()
		if !ok {
			return false
		}
		resolvedAnchor, ok := resolver.object(anchor)
		if !ok || resolvedAnchor.permanent == nil || !resolvedAnchor.permanent.AttachedTo.Exists {
			return false
		}
		return resolvedAnchor.permanent.AttachedTo.Val == permanent.ObjectID
	case game.GroupDomainBattlefield:
		return continuousSelectionApplies(g, resolver, effect.Group, source, permanent, values, effect.Controller)
	case game.GroupDomainObjectControlled:
		anchor, ok := effect.Group.Anchor()
		if !ok {
			return false
		}
		resolvedAnchor, ok := resolver.object(anchor)
		if !ok {
			return false
		}
		anchorController, ok := resolvedAnchor.controller(g)
		if !ok {
			return false
		}
		if effectiveController(g, permanent) != anchorController {
			return false
		}
		return continuousSelectionApplies(g, resolver, effect.Group, source, permanent, values, effect.Controller)
	default:
		return false
	}
}

// continuousSelectionApplies checks whether permanent satisfies the group's
// Selection and exclusion without allocating a members slice, shared by the
// Battlefield and ObjectControlled continuous-effect domains.
func continuousSelectionApplies(g *game.Game, resolver referenceResolver, group game.GroupReference, source, permanent *game.Permanent, values *permanentEffectiveValues, controller game.PlayerID) bool {
	sel := group.Selection()
	if sel.ExcludeSource && source == nil {
		return false
	}
	exclude, hasExclude := group.Exclusion()
	if hasExclude {
		excluded, ok := resolver.object(exclude)
		if ok && excluded.permanent != nil && excluded.permanent.ObjectID == permanent.ObjectID {
			return false
		}
		excludedID, _ := resolver.objectIdentityID(exclude)
		if excludedID != 0 && excludedID == permanent.ObjectID {
			return false
		}
	}
	subject := selectionSubject{
		kind:      subjectPermanent,
		g:         g,
		permanent: permanent,
		values:    values,
		viewer:    controller,
	}
	if sel.Controller != game.ControllerAny {
		subject.controller = values.controller
	}
	if source != nil {
		subject.sourceObjectID = source.ObjectID
	}
	return matchSelection(&subject, &sel)
}

func orderContinuousEffects(effects []game.ContinuousEffect) []game.ContinuousEffect {
	if len(effects) <= 1 {
		return effects
	}
	ordered := append([]game.ContinuousEffect(nil), effects...)
	for i := 1; i < len(ordered); i++ {
		for j := i; j > 0 && compareContinuousEffects(&ordered[j], &ordered[j-1]) < 0; j-- {
			ordered[j], ordered[j-1] = ordered[j-1], ordered[j]
		}
	}
	remaining := append([]game.ContinuousEffect(nil), ordered...)
	result := make([]game.ContinuousEffect, 0, len(ordered))
	applied := make(map[id.ID]bool, len(ordered))
	for len(remaining) > 0 {
		progress := false
		for i := 0; i < len(remaining); {
			if dependenciesSatisfied(&remaining[i], applied, remaining) {
				result = append(result, remaining[i])
				if remaining[i].ID != 0 {
					applied[remaining[i].ID] = true
				}
				remaining = append(remaining[:i], remaining[i+1:]...)
				progress = true
				continue
			}
			i++
		}
		if !progress {
			return append(result, remaining...)
		}
	}
	return result
}

func compareContinuousEffects(left, right *game.ContinuousEffect) int {
	if left.Timestamp < right.Timestamp {
		return -1
	}
	if left.Timestamp > right.Timestamp {
		return 1
	}
	if left.ID < right.ID {
		return -1
	}
	if left.ID > right.ID {
		return 1
	}
	return 0
}

func dependenciesSatisfied(effect *game.ContinuousEffect, applied map[id.ID]bool, remaining []game.ContinuousEffect) bool {
	for _, dependency := range effect.DependsOn {
		if dependency == 0 || applied[dependency] {
			continue
		}
		for i := range remaining {
			other := &remaining[i]
			if other.ID == dependency {
				return false
			}
		}
	}
	return true
}

func applyContinuousEffect(g *game.Game, permanent *game.Permanent, values *permanentEffectiveValues, effect *game.ContinuousEffect) {
	switch effect.Layer {
	case game.LayerCopy:
		if effect.CopyValues.Exists {
			applyCopyValues(g, permanent, values, &effect.CopyValues.Val)
		}
	case game.LayerControl:
		if effect.NewController.Exists {
			values.controller = effect.NewController.Val
			recalculateDynamicPT(g, values)
		}
	case game.LayerText:
		if effect.SetName != "" {
			values.name = effect.SetName
		}
		if effect.TextFrom != "" {
			values.oracleText = strings.ReplaceAll(values.oracleText, effect.TextFrom, effect.TextTo)
		}
	case game.LayerType:
		applyTypeLayer(g, values, effect)
	case game.LayerColor:
		if effect.SetColorless {
			values.colors = nil
		} else if effect.SetColors != nil {
			values.colors = append([]color.Color(nil), effect.SetColors...)
		}
		values.colors = removeColors(values.colors, effect.RemoveColors)
		values.colors = appendUniqueColors(values.colors, effect.AddColors...)
	case game.LayerAbility:
		if effect.RemoveAllAbilities {
			values.abilities = nil
			clear(values.keywords)
		}
		for _, body := range effect.AddAbilities {
			values.abilities = append(values.abilities, body)
			game.BodyAddKeywordKindsTo(body, values.keywords)
		}
		for _, keyword := range effect.RemoveKeywords {
			values.keywords[keyword] = false
		}
		for _, keyword := range effect.AddKeywords {
			values.keywords[keyword] = true
		}
	case game.LayerPowerToughnessSet:
		if effect.SetPower.Exists {
			values.powerPT = ptPtr(effect.SetPower)
			values.dynamicPower = nil
			values.power, values.powerOK = ptValue(g, values.controller, values.powerPT, nil)
		} else if effect.SetPowerDynamic.Exists {
			set := game.PT{Value: dynamicAmountValueForPermanent(g, permanent, effect.Controller, effect.SetPowerDynamic.Val, effect.Layer)}
			values.powerPT = &set
			values.dynamicPower = nil
			values.power, values.powerOK = ptValue(g, values.controller, values.powerPT, nil)
		}
		if effect.SetToughness.Exists {
			values.toughnessPT = ptPtr(effect.SetToughness)
			values.dynamicToughness = nil
			values.toughness, values.toughnessOK = ptValue(g, values.controller, values.toughnessPT, nil)
		} else if effect.SetToughnessDynamic.Exists {
			set := game.PT{Value: dynamicAmountValueForPermanent(g, permanent, effect.Controller, effect.SetToughnessDynamic.Val, effect.Layer)}
			values.toughnessPT = &set
			values.dynamicToughness = nil
			values.toughness, values.toughnessOK = ptValue(g, values.controller, values.toughnessPT, nil)
		}
	case game.LayerPowerToughnessModify:
		powerDelta := effect.PowerDelta
		if effect.PowerDeltaDynamic.Exists {
			powerDelta = dynamicAmountValueForPermanent(g, permanent, effect.Controller, effect.PowerDeltaDynamic.Val, effect.Layer)
		}
		toughnessDelta := effect.ToughnessDelta
		if effect.ToughnessDeltaDynamic.Exists {
			toughnessDelta = dynamicAmountValueForPermanent(g, permanent, effect.Controller, effect.ToughnessDeltaDynamic.Val, effect.Layer)
		}
		if values.powerOK {
			if effect.DoublePower {
				powerDelta += values.power
			}
			values.power += powerDelta
		}
		if values.toughnessOK {
			if effect.DoubleToughness {
				toughnessDelta += values.toughness
			}
			values.toughness += toughnessDelta
		}
	case game.LayerPowerToughnessSwitch:
		values.power, values.toughness = values.toughness, values.power
		values.powerOK, values.toughnessOK = values.toughnessOK, values.powerOK
	default:
	}
}

func applyCopyValues(g *game.Game, permanent *game.Permanent, values *permanentEffectiveValues, copyValues *game.CopyableValues) {
	if copyValues == nil {
		return
	}
	values.name = copyValues.Name
	values.oracleText = copyValues.OracleText
	values.colors = append([]color.Color(nil), copyValues.Colors...)
	values.supertypes = append([]types.Super(nil), copyValues.Supertypes...)
	values.types = append([]types.Card(nil), copyValues.Types...)
	values.subtypes = append([]types.Sub(nil), copyValues.Subtypes...)
	values.abilities = append([]game.Ability(nil), copyValues.Abilities...)
	values.powerPT = ptPtr(copyValues.Power)
	values.dynamicPower = dynamicValuePtr(copyValues.DynamicPower)
	values.toughnessPT = ptPtr(copyValues.Toughness)
	values.dynamicToughness = dynamicValuePtr(copyValues.DynamicToughness)
	values.power, values.powerOK = ptValue(g, values.controller, values.powerPT, values.dynamicPower)
	values.toughness, values.toughnessOK = ptValue(g, values.controller, values.toughnessPT, values.dynamicToughness)
	rebuildKeywords(values)
}

func recalculateDynamicPT(g *game.Game, values *permanentEffectiveValues) {
	if values == nil {
		return
	}
	// LayerControl runs before P/T layers, so this cannot erase later P/T effects.
	if values.powerPT != nil && values.powerPT.IsStar {
		values.power, values.powerOK = ptValue(g, values.controller, values.powerPT, values.dynamicPower)
	}
	if values.toughnessPT != nil && values.toughnessPT.IsStar {
		values.toughness, values.toughnessOK = ptValue(g, values.controller, values.toughnessPT, values.dynamicToughness)
	}
}

func applyTypeLayer(g *game.Game, values *permanentEffectiveValues, effect *game.ContinuousEffect) {
	if effect.SetSupertypes != nil {
		values.supertypes = append([]types.Super(nil), effect.SetSupertypes...)
	}
	values.supertypes = removeSupertypes(values.supertypes, effect.RemoveSupertypes)
	values.supertypes = appendUniqueSupertypes(values.supertypes, effect.AddSupertypes...)

	if effect.SetTypes != nil {
		values.types = append([]types.Card(nil), effect.SetTypes...)
	}
	values.types = removeTypes(values.types, effect.RemoveTypes)
	values.types = appendUniqueTypes(values.types, effect.AddTypes...)

	if effect.SetSubtypes != nil {
		values.subtypes = append([]types.Sub(nil), effect.SetSubtypes...)
	}
	values.subtypes = removeSubtypes(values.subtypes, effect.RemoveSubtypes)
	values.subtypes = appendUniqueSubtypes(values.subtypes, effect.AddSubtypes...)
	if effect.AddEveryCreatureType {
		values.subtypes = appendUniqueSubtypes(values.subtypes, types.SubtypesForType(types.Creature)...)
	}
	if effect.AddEveryBasicLandType {
		values.subtypes = appendUniqueSubtypes(values.subtypes, basicLandSubtypes[:]...)
	}
	if effect.AddSubtypeFromEntryChoice != "" {
		if source, ok := permanentByObjectID(g, effect.SourceObjectID); ok {
			if choice, ok := source.EntryChoices[effect.AddSubtypeFromEntryChoice]; ok &&
				choice.Kind == game.ResolutionChoiceSubtype &&
				choice.Subtype != "" {
				values.subtypes = appendUniqueSubtypes(values.subtypes, choice.Subtype)
			}
		}
	}
}

// basicLandSubtypes enumerates the five basic land subtypes (CR 305.6). It
// backs both the basic-land-type count condition and the "every basic land
// type" continuous type-grant (Dryad of the Ilysian Grove, Prismatic Omen).
var basicLandSubtypes = [...]types.Sub{types.Plains, types.Island, types.Swamp, types.Mountain, types.Forest}

// applyAddedBasicLandManaAbilities grants the intrinsic mana ability that a
// basic land type confers (CR 305.6) for every basic land subtype a continuous
// effect added to this permanent beyond its base subtypes. Printed basic land
// faces already carry their mana ability, so only subtypes gained from outside
// the base face contribute; this keeps a continuous "Each land is a Forest in
// addition to its other land types" static from duplicating the mana ability on
// lands that already have the granted type while letting any other land tap for
// the new color.
func applyAddedBasicLandManaAbilities(values *permanentEffectiveValues, baseSubtypes []types.Sub) {
	for _, subtype := range values.subtypes {
		manaColor, ok := basicLandSubtypeManaColor(subtype)
		if !ok || slices.Contains(baseSubtypes, subtype) {
			continue
		}
		ability := game.TapManaAbility(manaColor)
		if slices.ContainsFunc(values.abilities, func(existing game.Ability) bool {
			body, ok := existing.(*game.ManaAbility)
			return ok && reflect.DeepEqual(*body, ability)
		}) {
			continue
		}
		values.abilities = append(values.abilities, &ability)
	}
}

// basicLandSubtypeManaColor maps a basic land subtype onto the mana color its
// intrinsic ability produces, failing closed for any non-basic subtype.
func basicLandSubtypeManaColor(subtype types.Sub) (mana.Color, bool) {
	switch subtype {
	case types.Plains:
		return mana.W, true
	case types.Island:
		return mana.U, true
	case types.Swamp:
		return mana.B, true
	case types.Mountain:
		return mana.R, true
	case types.Forest:
		return mana.G, true
	default:
		return mana.C, false
	}
}

func applyCounterAndTemporaryValues(permanent *game.Permanent, values *permanentEffectiveValues) {
	counterDelta := powerToughnessCounterDelta(permanent)
	if values.powerOK {
		values.power += counterDelta + permanent.TemporaryPowerModifier
	}
	if values.toughnessOK {
		values.toughness += counterDelta + permanent.TemporaryToughnessModifier
	}
}

func rebuildKeywords(values *permanentEffectiveValues) {
	if values.keywords == nil {
		values.keywords = make(map[game.Keyword]bool)
	} else {
		clear(values.keywords)
	}
	for _, body := range values.abilities {
		game.BodyAddKeywordKindsTo(body, values.keywords)
	}
}

func bodyFunctionsOnBattlefield(body game.Ability) bool {
	functionZone := game.BodyFunctionZone(body)
	return functionZone == zone.None || functionZone == zone.Battlefield
}

func powerToughnessCounterDelta(permanent *game.Permanent) int {
	return permanent.Counters.Get(counter.PlusOnePlusOne) - permanent.Counters.Get(counter.MinusOneMinusOne)
}

func keywordCounters(permanent *game.Permanent) []game.Keyword {
	var keywords []game.Keyword
	for _, keyword := range []game.Keyword{
		game.Deathtouch,
		game.FirstStrike,
		game.Flying,
		game.Hexproof,
		game.Indestructible,
		game.Lifelink,
		game.Menace,
		game.Reach,
		game.Trample,
		game.Vigilance,
	} {
		counterKind, ok := keywordCounterKind(keyword)
		if ok && permanent.Counters.Get(counterKind) > 0 {
			keywords = append(keywords, keyword)
		}
	}
	return keywords
}

func removeColors(colors, remove []color.Color) []color.Color {
	return slices.DeleteFunc(colors, func(color color.Color) bool {
		return slices.Contains(remove, color)
	})
}

func appendUniqueColors(colors []color.Color, add ...color.Color) []color.Color {
	for _, clr := range add {
		if !slices.Contains(colors, clr) {
			colors = append(colors, clr)
		}
	}
	return colors
}

func removeSupertypes(supertypes, remove []types.Super) []types.Super {
	return slices.DeleteFunc(supertypes, func(supertype types.Super) bool {
		return slices.Contains(remove, supertype)
	})
}

func appendUniqueSupertypes(supertypes []types.Super, add ...types.Super) []types.Super {
	for _, supertype := range add {
		if !slices.Contains(supertypes, supertype) {
			supertypes = append(supertypes, supertype)
		}
	}
	return supertypes
}

func removeTypes(cardTypes, remove []types.Card) []types.Card {
	return slices.DeleteFunc(cardTypes, func(cardType types.Card) bool {
		return slices.Contains(remove, cardType)
	})
}

func appendUniqueTypes(cardTypes []types.Card, add ...types.Card) []types.Card {
	for _, cardType := range add {
		if !slices.Contains(cardTypes, cardType) {
			cardTypes = append(cardTypes, cardType)
		}
	}
	return cardTypes
}

func removeSubtypes(values, remove []types.Sub) []types.Sub {
	return slices.DeleteFunc(values, func(value types.Sub) bool {
		return slices.Contains(remove, value)
	})
}

func appendUniqueSubtypes(values []types.Sub, add ...types.Sub) []types.Sub {
	for _, value := range add {
		if !slices.Contains(values, value) {
			values = append(values, value)
		}
	}
	return values
}
