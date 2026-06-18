package rules

import (
	"slices"
	"strings"

	"github.com/natefinch/council4/mtg/game/zone"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
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
	values := effectivePermanentValues(g, permanent)
	return append([]game.Ability(nil), values.abilities...)
}

func permanentEffectiveName(g *game.Game, permanent *game.Permanent) string {
	return effectivePermanentValues(g, permanent).name
}

func effectiveController(g *game.Game, permanent *game.Permanent) game.PlayerID {
	values := basePermanentValues(g, permanent)
	sources := staticAbilitySources(g)
	effects := orderContinuousEffects(continuousEffectsForLayer(g, permanent, &values, game.LayerControl, sources))
	for i := range effects {
		effect := &effects[i]
		if effect.NewController.Exists {
			values.controller = effect.NewController.Val
		}
	}
	return values.controller
}

func effectivePermanentValues(g *game.Game, permanent *game.Permanent) permanentEffectiveValues {
	values := basePermanentValues(g, permanent)
	applyContinuousLayers(g, permanent, &values)
	applyCounterAndTemporaryValues(permanent, &values)
	for _, keyword := range keywordCounters(permanent) {
		values.keywords[keyword] = true
	}
	return values
}

func basePermanentValues(g *game.Game, permanent *game.Permanent) permanentEffectiveValues {
	values := permanentEffectiveValues{keywords: make(map[game.Keyword]bool)}
	values.controller = permanent.Controller
	if permanent.FaceDown {
		values.types = []types.Card{types.Creature}
		if permanent.FaceDownKind == game.FaceDownDisguise {
			values.abilities = []game.Ability{faceDownDisguiseWardBody()}
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
				values.abilities = append(values.abilities, faceDownDisguiseWardBody())
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
			if basePermanentHasType(g, permanent, types.Creature) {
				count++
			}
		}
		return count
	default:
	}
	return 0
}

func countControlledPermanentsWithType(g *game.Game, controller game.PlayerID, cardType types.Card) int {
	count := 0
	for _, permanent := range g.Battlefield {
		if permanent.Controller == controller && basePermanentHasType(g, permanent, cardType) {
			count++
		}
	}
	return count
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
	var sources []staticAbilitySource
	for _, permanent := range g.Battlefield {
		for _, component := range permanentStaticAbilityComponents(g, permanent) {
			if !staticAbilityCardHasContinuousEffects(component.card, true) {
				continue
			}
			sources = append(sources, staticAbilitySource{
				permanent:  permanent,
				card:       component.card,
				cardID:     component.cardID,
				controller: permanent.Controller,
				timestamp:  permanent.Timestamp(),
			})
		}
	}
	for playerID := range game.PlayerID(game.NumPlayers) {
		player := g.Players[playerID]
		player.Graveyard.Range(func(cardID id.ID) bool {
			card, ok := g.GetCardInstance(cardID)
			if !ok {
				return true
			}
			def := staticAbilityCardInstanceDef(card, game.FaceFront)
			if !staticAbilityCardHasContinuousEffects(def, false) {
				return true
			}
			sources = append(sources, staticAbilitySource{
				card:       def,
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

func permanentStaticAbilityComponents(g *game.Game, permanent *game.Permanent) []permanentAbilityComponent {
	card, ok := staticAbilityPermanentCardDef(g, permanent)
	if !ok {
		return nil
	}
	components := []permanentAbilityComponent{{card: card, cardID: permanent.CardInstanceID}}
	for _, merged := range permanent.MergedCards {
		if merged.FaceDown {
			continue
		}
		if merged.TokenDef != nil {
			def, ok := merged.TokenDef.FaceDef(merged.Face)
			if ok {
				components = append(components, permanentAbilityComponent{card: def})
			}
			continue
		}
		instance, ok := g.GetCardInstance(merged.CardInstanceID)
		if !ok {
			continue
		}
		def, ok := cardFaceDef(instance, merged.Face)
		if ok {
			components = append(components, permanentAbilityComponent{card: def, cardID: merged.CardInstanceID})
		}
	}
	return components
}

func staticAbilitySourceContinuousEffects(g *game.Game, source staticAbilitySource, permanent *game.Permanent, values *permanentEffectiveValues, layer game.ContinuousLayer) []game.ContinuousEffect {
	var effects []game.ContinuousEffect
	for i := range source.card.StaticAbilities {
		body := &source.card.StaticAbilities[i]
		if !staticAbilityFunctionsFromSource(body, source) {
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

func staticAbilityPermanentCardDef(g *game.Game, permanent *game.Permanent) (*game.CardDef, bool) {
	if permanent.FaceDown {
		return nil, false
	}
	if permanent.Token {
		if permanent.TokenDef == nil {
			return nil, false
		}
		if !permanent.TokenDef.Back.Exists && permanent.Face == game.FaceFront {
			return permanent.TokenDef, true
		}
		return permanent.TokenDef.FaceDef(permanent.Face)
	}
	card, ok := g.GetCardInstance(permanent.CardInstanceID)
	if !ok {
		return nil, false
	}
	if !card.Def.Back.Exists && permanent.Face == game.FaceFront {
		return card.Def, true
	}
	return card.Def.FaceDef(permanent.Face)
}

func staticAbilityCardInstanceDef(card *game.CardInstance, face game.FaceIndex) *game.CardDef {
	if card == nil {
		return nil
	}
	if !card.Def.Back.Exists && face == game.FaceFront {
		return card.Def
	}
	return cardFaceOrDefault(card, face)
}

func staticAbilityCardHasLayer(card *game.CardDef, onBattlefield bool, layer game.ContinuousLayer) bool {
	if card == nil {
		return false
	}
	for i := range card.StaticAbilities {
		body := &card.StaticAbilities[i]
		if !staticAbilityFunctionsInZone(body, onBattlefield) {
			continue
		}
		if staticAbilityHasEffectForLayer(body, layer) {
			return true
		}
	}
	return false
}

func staticAbilityCardHasContinuousEffects(card *game.CardDef, onBattlefield bool) bool {
	if card == nil {
		return false
	}
	for i := range card.StaticAbilities {
		body := &card.StaticAbilities[i]
		if staticAbilityFunctionsInZone(body, onBattlefield) && len(body.ContinuousEffects) > 0 {
			return true
		}
	}
	return false
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
		if effect.TextFrom != "" {
			values.oracleText = strings.ReplaceAll(values.oracleText, effect.TextFrom, effect.TextTo)
		}
	case game.LayerType:
		applyTypeLayer(values, effect)
	case game.LayerColor:
		if effect.SetColors != nil {
			values.colors = append([]color.Color(nil), effect.SetColors...)
		}
		values.colors = removeColors(values.colors, effect.RemoveColors)
		values.colors = appendUniqueColors(values.colors, effect.AddColors...)
	case game.LayerAbility:
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
		}
		if effect.SetToughness.Exists {
			values.toughnessPT = ptPtr(effect.SetToughness)
			values.dynamicToughness = nil
			values.toughness, values.toughnessOK = ptValue(g, values.controller, values.toughnessPT, nil)
		}
	case game.LayerPowerToughnessModify:
		powerDelta := effect.PowerDelta
		if effect.PowerDeltaDynamic.Exists {
			powerDelta = dynamicAmountValueBeforeLayer(g, nil, effect.Controller, effect.PowerDeltaDynamic.Val, effect.Layer)
		}
		toughnessDelta := effect.ToughnessDelta
		if effect.ToughnessDeltaDynamic.Exists {
			toughnessDelta = dynamicAmountValueBeforeLayer(g, nil, effect.Controller, effect.ToughnessDeltaDynamic.Val, effect.Layer)
		}
		if values.powerOK {
			values.power += powerDelta
		}
		if values.toughnessOK {
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

func applyTypeLayer(values *permanentEffectiveValues, effect *game.ContinuousEffect) {
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
