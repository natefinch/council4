package rules

import (
	"slices"
	"strings"

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
	abilities  []game.AbilityDef
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

func permanentEffectiveAbilities(g *game.Game, permanent *game.Permanent) []game.AbilityDef {
	values := effectivePermanentValues(g, permanent)
	return append([]game.AbilityDef(nil), values.abilities...)
}

func permanentEffectiveName(g *game.Game, permanent *game.Permanent) string {
	return effectivePermanentValues(g, permanent).name
}

func effectiveController(g *game.Game, permanent *game.Permanent) game.PlayerID {
	values := basePermanentValues(g, permanent)
	effects := orderContinuousEffects(continuousEffectsForLayer(g, permanent, &values, game.LayerControl))
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
			values.abilities = []game.AbilityDef{faceDownDisguiseWardAbility()}
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
	values.abilities = append([]game.AbilityDef(nil), card.AbilityDefs()...)
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
	for _, layer := range []game.ContinuousLayer{
		game.LayerCopy,
		game.LayerControl,
		game.LayerText,
		game.LayerType,
		game.LayerColor,
		game.LayerAbility,
		game.LayerPowerToughnessSet,
		game.LayerPowerToughnessModify,
		game.LayerPowerToughnessSwitch,
	} {
		effects := continuousEffectsForLayer(g, permanent, values, layer)
		ordered := orderContinuousEffects(effects)
		for i := range ordered {
			applyContinuousEffect(g, permanent, values, &ordered[i])
		}
	}
}

func continuousEffectsForLayer(g *game.Game, permanent *game.Permanent, values *permanentEffectiveValues, layer game.ContinuousLayer) []game.ContinuousEffect {
	var effects []game.ContinuousEffect
	for i := range g.ContinuousEffects {
		effect := &g.ContinuousEffects[i]
		if effect.Layer == layer && continuousEffectApplies(g, permanent, values, effect) {
			effects = append(effects, *effect)
		}
	}
	effects = append(effects, staticAbilityContinuousEffectsForLayer(g, permanent, values, layer)...)
	return effects
}

type staticAbilitySource struct {
	permanent  *game.Permanent
	card       *game.CardDef
	cardID     id.ID
	controller game.PlayerID
	timestamp  game.Timestamp
}

func staticAbilityContinuousEffectsForLayer(g *game.Game, permanent *game.Permanent, values *permanentEffectiveValues, layer game.ContinuousLayer) []game.ContinuousEffect {
	var effects []game.ContinuousEffect
	for _, source := range staticAbilitySources(g, layer) {
		effects = append(effects, staticAbilitySourceContinuousEffects(g, source, permanent, values, layer)...)
	}
	return effects
}

func staticAbilitySources(g *game.Game, layer game.ContinuousLayer) []staticAbilitySource {
	var sources []staticAbilitySource
	for _, permanent := range g.Battlefield {
		card, ok := staticAbilityPermanentCardDef(g, permanent)
		if !ok {
			continue
		}
		if !staticAbilityCardHasLayer(card, true, layer) {
			continue
		}
		controller := permanent.Controller
		if layer != game.LayerControl {
			controller = effectiveController(g, permanent)
		}
		sources = append(sources, staticAbilitySource{
			permanent:  permanent,
			card:       card,
			cardID:     permanent.CardInstanceID,
			controller: controller,
			timestamp:  permanent.Timestamp(),
		})
	}
	for playerID := range game.PlayerID(game.NumPlayers) {
		player := g.Players[playerID]
		for _, cardID := range player.Graveyard.All() {
			card, ok := g.GetCardInstance(cardID)
			if !ok {
				continue
			}
			def := staticAbilityCardInstanceDef(card, game.FaceFront)
			if !staticAbilityCardHasLayer(def, false, layer) {
				continue
			}
			sources = append(sources, staticAbilitySource{
				card:       def,
				cardID:     card.ID,
				controller: card.Owner,
				timestamp:  game.Timestamp(card.ID),
			})
		}
	}
	return sources
}

func staticAbilitySourceContinuousEffects(g *game.Game, source staticAbilitySource, permanent *game.Permanent, values *permanentEffectiveValues, layer game.ContinuousLayer) []game.ContinuousEffect {
	var effects []game.ContinuousEffect
	abilities := source.card.AbilityDefs()
	for i := range abilities {
		ability := &abilities[i]
		if !staticAbilityFunctionsFromSource(ability, source) {
			continue
		}
		if !conditionSatisfied(g, conditionContext{
			controller:             source.controller,
			source:                 source.permanent,
			useBaseCharacteristics: true,
		}, ability.Condition) {
			continue
		}
		if body, ok := ability.StaticBody(); ok {
			for i := range body.ContinuousEffects {
				template := &body.ContinuousEffects[i]
				if template.Layer != layer {
					continue
				}
				staticEffect := *template
				staticEffect.SourceObjectID = sourceObjectID(source)
				staticEffect.SourceCardID = source.cardID
				staticEffect.Controller = source.controller
				staticEffect.Timestamp = source.timestamp
				if continuousEffectApplies(g, permanent, values, &staticEffect) {
					effects = append(effects, staticEffect)
				}
			}
		}
		for i := range ability.Effects {
			effect := &ability.Effects[i]
			if layer == game.LayerPowerToughnessModify && effect.Type == game.EffectModifyPT && permanentValuesMatchSelectorForSource(source.permanent, source.controller, permanent, values, effect.Selector) {
				powerDelta := effect.PowerDelta
				if effect.DynamicAmount.Exists {
					powerDelta = dynamicAmountValue(g, nil, source.controller, effect.DynamicAmount.Val)
				}
				effects = append(effects, game.ContinuousEffect{
					SourceObjectID:   sourceObjectID(source),
					SourceCardID:     source.cardID,
					Controller:       source.controller,
					Timestamp:        source.timestamp,
					AffectedObjectID: permanent.ObjectID,
					Layer:            game.LayerPowerToughnessModify,
					PowerDelta:       powerDelta,
					ToughnessDelta:   effect.ToughnessDelta,
				})
			}
			if effect.Type != game.EffectApplyContinuous {
				continue
			}
			for i := range effect.ContinuousEffects {
				template := &effect.ContinuousEffects[i]
				if template.Layer != layer {
					continue
				}
				staticEffect := *template
				staticEffect.SourceObjectID = sourceObjectID(source)
				staticEffect.SourceCardID = source.cardID
				staticEffect.Controller = source.controller
				staticEffect.Timestamp = source.timestamp
				if continuousEffectApplies(g, permanent, values, &staticEffect) {
					effects = append(effects, staticEffect)
				}
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
	abilities := card.AbilityDefs()
	for i := range abilities {
		ability := &abilities[i]
		if !staticAbilityFunctionsInZone(ability, onBattlefield) {
			continue
		}
		if staticAbilityHasEffectForLayer(ability, layer) {
			return true
		}
	}
	return false
}

func staticAbilityHasEffectForLayer(ability *game.AbilityDef, layer game.ContinuousLayer) bool {
	if body, ok := ability.StaticBody(); ok {
		for i := range body.ContinuousEffects {
			if body.ContinuousEffects[i].Layer == layer {
				return true
			}
		}
	}
	for i := range ability.Effects {
		effect := &ability.Effects[i]
		if effect.Type == game.EffectModifyPT && layer == game.LayerPowerToughnessModify {
			return true
		}
		if effect.Type != game.EffectApplyContinuous {
			continue
		}
		for i := range effect.ContinuousEffects {
			template := &effect.ContinuousEffects[i]
			if template.Layer == layer {
				return true
			}
		}
	}
	return false
}

func staticAbilityFunctionsFromSource(ability *game.AbilityDef, source staticAbilitySource) bool {
	return staticAbilityFunctionsInZone(ability, source.permanent != nil)
}

func staticAbilityFunctionsInZone(ability *game.AbilityDef, onBattlefield bool) bool {
	if ability == nil || !ability.IsStatic() {
		return false
	}
	if onBattlefield {
		return abilityFunctionsOnBattlefield(ability)
	}
	return ability.FunctionZone() == game.ZoneGraveyard
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
	if effect.Selector == game.EffectSelectorNone {
		return false
	}
	source, _ := permanentByObjectID(g, effect.SourceObjectID)
	return permanentValuesMatchSelectorForSource(source, effect.Controller, permanent, values, effect.Selector)
}

func permanentValuesMatchSelectorForSource(source *game.Permanent, controller game.PlayerID, permanent *game.Permanent, values *permanentEffectiveValues, selector game.EffectSelector) bool {
	switch selector {
	case game.EffectSelectorAllCreatures:
		return valuesHasType(values, types.Creature)
	case game.EffectSelectorAllArtifacts:
		return valuesHasType(values, types.Artifact)
	case game.EffectSelectorAllEnchantments:
		return valuesHasType(values, types.Enchantment)
	case game.EffectSelectorAllNonlandPermanents:
		return !valuesHasType(values, types.Land)
	case game.EffectSelectorAllPermanents:
		return true
	case game.EffectSelectorCreaturesYouControl:
		return values.controller == controller && valuesHasType(values, types.Creature)
	case game.EffectSelectorOtherCreaturesYouControl:
		return source != nil && permanent.ObjectID != source.ObjectID && values.controller == controller && valuesHasType(values, types.Creature)
	case game.EffectSelectorEquippedCreature:
		return source != nil && source.AttachedTo.Exists && permanent.ObjectID == source.AttachedTo.Val
	default:
		return false
	}
}

func valuesHasType(values *permanentEffectiveValues, cardType types.Card) bool {
	return slices.Contains(values.types, cardType)
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
		values.abilities = append(values.abilities, effect.AddAbilities...)
		for i := range effect.AddAbilities {
			effect.AddAbilities[i].AddKeywordKindsTo(values.keywords)
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
			powerDelta = dynamicAmountValue(g, nil, effect.Controller, effect.PowerDeltaDynamic.Val)
		}
		toughnessDelta := effect.ToughnessDelta
		if effect.ToughnessDeltaDynamic.Exists {
			toughnessDelta = dynamicAmountValue(g, nil, effect.Controller, effect.ToughnessDeltaDynamic.Val)
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
	values.abilities = append([]game.AbilityDef(nil), copyValues.Abilities...)
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
	values.keywords = make(map[game.Keyword]bool)
	for i := range values.abilities {
		values.abilities[i].AddKeywordKindsTo(values.keywords)
	}
}

func abilityFunctionsOnBattlefield(ability *game.AbilityDef) bool {
	zone := ability.FunctionZone()
	return zone == game.ZoneNone || zone == game.ZoneBattlefield
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
