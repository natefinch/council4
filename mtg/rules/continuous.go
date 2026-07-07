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

	// beforeValues memoizes permanentValuesBeforeLayer per (permanent, stop
	// layer). Evaluating a static ability's condition during layer application
	// scans the battlefield and recomputes each permanent's characteristics "as of
	// an earlier layer", and that scan repeats for every permanent evaluated, every
	// layer, and every conditional static ability — the dominant cost of a large
	// board. Within a frame the game state is fixed, so the "before layer L" values
	// of a permanent never change and are memoized here. Cached values are
	// read-only to callers except for the scalar power/toughness a caller may add
	// counters/modifiers to on its own by-value copy.
	beforeValues map[beforeLayerKey]permanentEffectiveValues

	// ruleEffects memoizes activeRuleEffects: the active rule effects on the
	// battlefield (and stack/graveyard/exile). Building that set rescans every
	// permanent's static abilities and evaluates each condition, and the engine
	// asks for it constantly (every legality, combat, and trigger check). Within a
	// frame the set is fixed, so it is built once. It is clipped (cap == len) so a
	// caller that appends reallocates rather than corrupting the shared slice;
	// callers otherwise treat it read-only.
	ruleEffects      []game.RuleEffect
	ruleEffectsBuilt bool
}

// beforeLayerKey identifies a permanentValuesBeforeLayer result: a permanent and
// the layer the computation stops before.
type beforeLayerKey struct {
	object id.ID
	stop   game.ContinuousLayer
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

// effectiveController resolves a permanent's controller by applying only the
// control-changing effects of layer 2 (CR 613.1b). It is separated from the full
// layer computation because many layer effects are controller-relative, so the
// controller must be known before the later layers are applied.
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
		if effect.NewControllerIsMonarch {
			if monarch := livingMonarch(g); monarch.Exists {
				values.controller = monarch.Val
			}
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

// effectivePermanentValues computes a permanent's characteristics by the
// continuous-effect layer system (CR 613.1): start with the permanent's own
// printed/defined values, then apply each layer in order. Counters and temporary
// power/toughness modifiers are applied within power/toughness layer 7c during
// the layer pass (see applyContinuousLayers), not afterward. Results are memoized
// per static-source frame.
//
// A characteristic-defining effect can depend on the very characteristic it
// defines — for example a creature whose power is set to "the greatest power
// among creatures you control", a group that includes itself, or two such
// creatures that measure each other. Computing the permanent then re-enters this
// function for the same permanent; the game's characteristic-computation guard
// detects that loop and breaks it by returning base (pre-layer) values, so a real
// board state produces a finite answer instead of overflowing the stack (CR
// 613.8c: a dependency loop is applied in timestamp order, i.e. without
// re-entering).
func effectivePermanentValues(g *game.Game, permanent *game.Permanent) permanentEffectiveValues {
	fc := frameCacheFor(g)
	if fc != nil {
		if values, ok := fc.values[permanent.ObjectID]; ok {
			return values
		}
	}
	if g.BeginCharacteristicComputation(permanent.ObjectID) {
		return basePermanentValues(g, permanent)
	}
	defer g.EndCharacteristicComputation(permanent.ObjectID)

	values := basePermanentValues(g, permanent)
	baseSubtypes := append([]types.Sub(nil), values.subtypes...)
	applyContinuousLayers(g, permanent, &values)
	applyAddedBasicLandManaAbilities(&values, baseSubtypes)
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

// basePermanentValues returns the permanent's characteristics before any
// continuous effects are applied: the starting point of the layer system, i.e.
// the values printed on the card or defined by the effect that created a token
// or copy (CR 613.1). Face-down permanents start as 2/2 creatures with no text,
// name, subtypes, or mana cost (CR 613.2b, CR 708.2a).
func basePermanentValues(g *game.Game, permanent *game.Permanent) permanentEffectiveValues {
	values := permanentEffectiveValues{keywords: make(map[game.Keyword]bool)}
	values.controller = permanent.Controller
	if permanent.FaceDown {
		values.types = []types.Card{types.Creature}
		if permanent.FaceDownKind == game.FaceDownDisguise || permanent.FaceDownKind == game.FaceDownCloak {
			ward := faceDownDisguiseWardBody()
			values.abilities = []game.Ability{&ward}
			rebuildKeywords(permanent, &values)
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
			if component.FaceDownKind == game.FaceDownDisguise || component.FaceDownKind == game.FaceDownCloak {
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
	rebuildKeywords(permanent, &values)
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

// applyContinuousLayers applies the continuous effects that affect a permanent
// in the order prescribed by the layer system (CR 613.1): each layer in turn,
// with the effects within a layer ordered by timestamp and dependency (see
// orderContinuousEffects). The Changeling special case adds every creature type
// in the type-changing layer 4 (CR 613.1d) before type-dependent effects apply.
//
// CR 613.4c: +1/+1 / -1/-1 counters and temporary "until end of turn" power and
// toughness modifiers are part of power/toughness layer 7c, alongside the 7c
// modifying continuous effects. They are injected here as a single synthetic 7c
// effect timestamped with the permanent's own timestamp (CR 613.7c counters are
// timestamped when placed; the engine does not track per-counter timestamps, so
// it uses the permanent's timestamp), so they are ordered by timestamp with the
// other 7c effects and applied before the layer 7d power/toughness switch. This
// matters only for non-commutative interactions (a doubling 7c effect, or an
// asymmetric temporary modifier combined with a 7d switch); additive modifiers
// commute, so the result is unchanged for them.
func applyContinuousLayers(g *game.Game, permanent *game.Permanent, values *permanentEffectiveValues) {
	sources := staticAbilitySources(g)
	counterEffect, hasCounterEffect := counterAndTemporaryEffect(permanent)
	for _, layer := range continuousLayers {
		if layer == game.LayerType && values.keywords[game.Changeling] {
			values.subtypes = append([]types.Sub(nil), types.SubtypesForType(types.Creature)...)
		}
		effects := continuousEffectsForLayer(g, permanent, values, layer, sources)
		if layer == game.LayerPowerToughnessModify && hasCounterEffect {
			effects = append(effects, counterEffect)
		}
		ordered := orderContinuousEffects(effects)
		for i := range ordered {
			applyContinuousEffect(g, permanent, values, &ordered[i])
		}
	}
}

// counterAndTemporaryEffect builds the synthetic layer-7c continuous effect that
// represents a permanent's +1/+1 / -1/-1 counters and temporary power/toughness
// modifiers (CR 613.4c), so they interleave with the other 7c effects by
// timestamp. It reports false when there is no net modifier.
func counterAndTemporaryEffect(permanent *game.Permanent) (game.ContinuousEffect, bool) {
	counterPower, counterToughness := powerToughnessCounterDelta(permanent)
	powerDelta := counterPower + permanent.TemporaryPowerModifier
	toughnessDelta := counterToughness + permanent.TemporaryToughnessModifier
	if powerDelta == 0 && toughnessDelta == 0 {
		return game.ContinuousEffect{}, false
	}
	return game.ContinuousEffect{
		AffectedObjectID: permanent.ObjectID,
		Timestamp:        permanent.Timestamp(),
		Layer:            game.LayerPowerToughnessModify,
		PowerDelta:       powerDelta,
		ToughnessDelta:   toughnessDelta,
	}, true
}

func permanentValuesBeforeLayer(g *game.Game, permanent *game.Permanent, stop game.ContinuousLayer) permanentEffectiveValues {
	fc := frameCacheFor(g)
	key := beforeLayerKey{object: permanent.ObjectID, stop: stop}
	if fc != nil {
		if cached, ok := fc.beforeValues[key]; ok {
			return cached
		}
	}
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
	if fc != nil {
		if fc.beforeValues == nil {
			fc.beforeValues = make(map[beforeLayerKey]permanentEffectiveValues)
		}
		fc.beforeValues[key] = values
	}
	return values
}

// continuousLayers is the order in which continuous effects are applied
// (CR 613.1): layer 1 copy (613.1a), layer 2 control (613.1b), layer 3 text
// (613.1c), layer 4 type (613.1d), layer 5 color (613.1e), layer 6 ability
// (613.1f), then the layer 7 power/toughness sublayers (613.4): 7b set effects
// (613.4b), 7c modifying effects and counters (613.4c), and 7d power/toughness
// switch (613.4d). Layer 7a characteristic-defining P/T (613.4a) is folded into
// the base values; layer 7c counters and temporary modifiers are injected into
// the 7c pass as a synthetic effect (see counterAndTemporaryEffect).
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
			rewriteChosenColorProtectionEffect(&staticEffect, source.permanent)
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

// rewriteChosenColorProtectionEffect resolves a granted "protection from the
// chosen color" ability on a continuous ability-layer effect to protection from
// the concrete color the granting source chose as it entered the battlefield
// (stored under EntryColorChoiceKey). The granted body is freshly built where a
// rewrite happens so the shared card-definition template is left untouched. When
// the source has made no recorded color choice the marker is left in place; the
// unresolved ChosenColor marker matches no source, so the grant fails closed.
func rewriteChosenColorProtectionEffect(effect *game.ContinuousEffect, source *game.Permanent) {
	if effect.Layer != game.LayerAbility || source == nil {
		return
	}
	choice, ok := source.EntryChoices[game.EntryColorChoiceKey]
	if !ok || choice.Kind != game.ResolutionChoiceMana {
		return
	}
	chosen, ok := manaColor(choice.Color)
	if !ok {
		return
	}
	cloned := false
	for j, ability := range effect.AddAbilities {
		static, ok := ability.(*game.StaticAbility)
		if !ok {
			continue
		}
		prot, ok := game.StaticBodyProtectionKeyword(static)
		if !ok || !prot.ChosenColor {
			continue
		}
		resolved := game.ProtectionFromColorsStaticAbility(chosen)
		if !cloned {
			effect.AddAbilities = append([]game.Ability(nil), effect.AddAbilities...)
			cloned = true
		}
		effect.AddAbilities[j] = &resolved
	}
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

// orderContinuousEffects orders the effects within a single layer. Effects are
// first sorted by timestamp, breaking ties by ID (CR 613.7: an effect with an
// earlier timestamp is applied first). Dependencies then override timestamp
// order (CR 613.8): each step applies the earliest-timestamp effect whose
// dependencies have already been applied, reevaluating after each one
// (CR 613.8c), so a dependent effect applies just after its dependencies
// orderContinuousEffects orders the effects within a single layer. Effects are
// first sorted by timestamp, breaking ties by ID (CR 613.7: an effect with an
// earlier timestamp is applied first). Dependencies then override timestamp
// order (CR 613.8): each step applies the earliest-timestamp applicable effect,
// reevaluating after each one (CR 613.8c), so a dependent effect applies just
// after its dependencies (CR 613.8b). Effects that form a dependency loop are
// treated as a strongly connected component: CR 613.8b ignores the mutual
// dependencies within the loop and applies the loop's effects in timestamp order,
// but the loop as a whole still waits until every dependency any of its members
// has on an effect outside the loop has been applied.
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
	for len(remaining) > 0 {
		// CR 613.8c: reevaluate after each application. Applied effects are
		// removed from remaining, so a dependency that is no longer present is
		// already satisfied. Scanning the timestamp-sorted remaining front to
		// back takes the earliest-timestamp applicable effect.
		adjacency := dependencyIndexAdjacency(remaining)
		reachable := reachabilityMatrix(adjacency)
		next := -1
		for i := range remaining {
			if effectComponentApplicable(i, adjacency, reachable) {
				next = i
				break
			}
		}
		if next == -1 {
			// Defensive: the condensation of the dependency graph is a DAG, so a
			// non-empty set always has an applicable component; fall back to
			// timestamp order to guarantee progress if that ever fails.
			return append(result, remaining...)
		}
		result = append(result, remaining[next])
		remaining = append(remaining[:next], remaining[next+1:]...)
	}
	return result
}

// dependencyIndexAdjacency builds the dependency graph among the remaining effects
// keyed by their index in remaining: adjacency[i] holds the indices of the effects
// that effect i depends on that are still remaining. A dependency that is no longer
// remaining (already applied, or never present) is omitted, and effects without an
// ID can't be depended on so they have no incoming edges.
func dependencyIndexAdjacency(remaining []game.ContinuousEffect) [][]int {
	indexByID := make(map[id.ID]int, len(remaining))
	for i := range remaining {
		if remaining[i].ID != 0 {
			indexByID[remaining[i].ID] = i
		}
	}
	adjacency := make([][]int, len(remaining))
	for i := range remaining {
		for _, dependency := range remaining[i].DependsOn {
			if dependency == 0 {
				continue
			}
			if j, ok := indexByID[dependency]; ok {
				adjacency[i] = append(adjacency[i], j)
			}
		}
	}
	return adjacency
}

// reachabilityMatrix returns reachable where reachable[i][j] reports whether
// effect j is reachable from effect i along dependency edges.
func reachabilityMatrix(adjacency [][]int) [][]bool {
	reachable := make([][]bool, len(adjacency))
	for i := range adjacency {
		reachable[i] = make([]bool, len(adjacency))
		var walk func(node int)
		walk = func(node int) {
			for _, next := range adjacency[node] {
				if !reachable[i][next] {
					reachable[i][next] = true
					walk(next)
				}
			}
		}
		walk(i)
	}
	return reachable
}

// sameComponent reports whether effects i and j are in the same strongly
// connected component of the dependency graph (each reaches the other), i.e. they
// are in the same dependency loop.
func sameComponent(i, j int, reachable [][]bool) bool {
	return i == j || (reachable[i][j] && reachable[j][i])
}

// effectComponentApplicable reports whether effect i can be applied now (CR 613.8):
// its strongly connected component has no remaining dependency on an effect outside
// the component. Within a component the mutual dependencies are ignored (CR 613.8b),
// so once the component's external dependencies are applied its members become
// applicable in timestamp order; an effect that depends on a loop member from
// outside the loop is its own component and still waits for that member.
func effectComponentApplicable(i int, adjacency [][]int, reachable [][]bool) bool {
	for j := range adjacency {
		if !sameComponent(i, j, reachable) {
			continue
		}
		for _, dependency := range adjacency[j] {
			if !sameComponent(i, dependency, reachable) {
				return false
			}
		}
	}
	return true
}

// compareContinuousEffects orders two effects by timestamp, breaking ties by ID
// for a deterministic order (CR 613.7: earlier timestamp applies first).
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

// applyContinuousEffect applies a single continuous effect to a permanent's
// in-progress values according to its layer (CR 613.1): copy (layer 1), control
// (layer 2), text (layer 3, CR 612), type (layer 4), color (layer 5), ability
// (layer 6), and the power/toughness sublayers (layer 7, CR 613.4).
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
		if effect.NewControllerIsMonarch {
			if monarch := livingMonarch(g); monarch.Exists {
				values.controller = monarch.Val
				recalculateDynamicPT(g, values)
			}
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
		// Layer 7b: effects that set power and/or toughness to a specific value
		// (CR 613.4b).
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
		// Layer 7c: effects that modify (but don't set) power and/or toughness,
		// including doubling (CR 613.4c). Counters and temporary modifiers, also
		// part of 7c, are injected here as a synthetic effect during the final
		// layer pass (see counterAndTemporaryEffect).
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
		// Layer 7d: effects that switch a creature's power and toughness
		// (CR 613.4d).
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
	rebuildKeywords(permanent, values)
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

// applyCounterAndTemporaryValues applies +1/+1 and -1/-1 counters and temporary
// power/toughness modifiers as a flat addition. It is used when evaluating a
// permanent's characteristics for a static ability's condition, where the current
// power/toughness (including counters) is needed at a bounded layer boundary. The
// final effective values instead inject these as a timestamped layer-7c effect
// during the layer pass (see counterAndTemporaryEffect), so they order correctly
// with non-commutative 7c effects and the 7d switch (CR 613.4c).
func applyCounterAndTemporaryValues(permanent *game.Permanent, values *permanentEffectiveValues) {
	counterPower, counterToughness := powerToughnessCounterDelta(permanent)
	if values.powerOK {
		values.power += counterPower + permanent.TemporaryPowerModifier
	}
	if values.toughnessOK {
		values.toughness += counterToughness + permanent.TemporaryToughnessModifier
	}
}

func rebuildKeywords(permanent *game.Permanent, values *permanentEffectiveValues) {
	if values.keywords == nil {
		values.keywords = make(map[game.Keyword]bool)
	} else {
		clear(values.keywords)
	}
	for _, body := range values.abilities {
		if static, ok := body.(*game.StaticAbility); ok &&
			!levelBandKeywordConditionSatisfied(permanent, static.Condition) {
			continue
		}
		game.BodyAddKeywordKindsTo(body, values.keywords)
	}
}

func levelBandKeywordConditionSatisfied(
	permanent *game.Permanent,
	condition opt.V[game.Condition],
) bool {
	if !condition.Exists ||
		condition.Val.SourceLevelCountersAtLeast == 0 &&
			condition.Val.SourceLevelCountersLessThan == 0 {
		return true
	}
	levels := permanent.Counters.Get(counter.Level)
	return levels >= condition.Val.SourceLevelCountersAtLeast &&
		(condition.Val.SourceLevelCountersLessThan == 0 ||
			levels < condition.Val.SourceLevelCountersLessThan)
}

func bodyFunctionsOnBattlefield(body game.Ability) bool {
	functionZone := game.BodyFunctionZone(body)
	return functionZone == zone.None || functionZone == zone.Battlefield
}

func powerToughnessCounterDelta(permanent *game.Permanent) (power, toughness int) {
	return permanent.Counters.PowerToughnessDelta()
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
