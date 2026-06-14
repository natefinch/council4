package rules

import (
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
			if linkedZone, ok := cardZone(g, linked.CardID); ok {
				return linked.CardID, linkedZone, true
			}
		}
		return 0, zone.None, false
	case game.CardReferenceTarget:
		if obj == nil {
			return 0, zone.None, false
		}
		cardTargetIndex := 0
		for _, target := range obj.Targets {
			if target.Kind != game.TargetCard || target.CardID == 0 {
				continue
			}
			if cardTargetIndex != ref.TargetIndex {
				cardTargetIndex++
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

func createTokenPermanent(g *game.Game, controller game.PlayerID, token *game.CardDef) (*game.Permanent, bool) {
	amount := replacementTokenCreationAmount(g, controller, 1)
	simultaneousID := tokenCreationSimultaneousID(g, amount)
	var first *game.Permanent
	for range amount {
		permanent, ok := createTokenPermanentWithChoicesInBatch(NewEngine(nil), g, controller, token, simultaneousID, [game.NumPlayers]PlayerAgent{}, nil)
		if !ok {
			return nil, false
		}
		if first == nil {
			first = permanent
		}
	}
	return first, first != nil
}

func createTokenPermanentsWithChoices(e *Engine, g *game.Game, controller game.PlayerID, token *game.CardDef, amount int, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	amount = replacementTokenCreationAmount(g, controller, amount)
	if amount <= 0 {
		return false
	}
	simultaneousID := tokenCreationSimultaneousID(g, amount)
	for range amount {
		if _, ok := createTokenPermanentWithChoicesInBatch(e, g, controller, token, simultaneousID, agents, log); !ok {
			return false
		}
	}
	return true
}

func createTokenPermanentWithChoices(e *Engine, g *game.Game, controller game.PlayerID, token *game.CardDef, agents [game.NumPlayers]PlayerAgent, log *TurnLog) (*game.Permanent, bool) {
	return createTokenPermanentWithChoicesInBatch(e, g, controller, token, 0, agents, log)
}

func tokenCreationSimultaneousID(g *game.Game, amount int) id.ID {
	if amount > 1 {
		return g.IDGen.Next()
	}
	return 0
}

func createTokenPermanentWithChoicesInBatch(e *Engine, g *game.Game, controller game.PlayerID, token *game.CardDef, simultaneousID id.ID, agents [game.NumPlayers]PlayerAgent, log *TurnLog) (*game.Permanent, bool) {
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
	registerPermanentReplacementEffects(g, permanent)
	initializeReadAhead(e, g, permanent, agents, log)
	applyEnterBattlefieldReplacementEffects(enterBattlefieldContext{
		engine: e,
		agents: agents,
		log:    log,
	}, g, permanent, zone.None)
	g.Battlefield = append(g.Battlefield, permanent)
	if lore := permanent.Counters.Get(counter.Lore); lore > 0 {
		emitCounterAddedEvent(g, permanent, effectiveController(g, permanent), counter.Lore, 0, lore)
	}
	event := game.Event{
		Controller:     controller,
		Player:         controller,
		PermanentID:    objectID,
		TokenName:      token.Name,
		TokenDef:       token,
		FromZone:       zone.None,
		ToZone:         zone.Battlefield,
		SimultaneousID: simultaneousID,
	}
	emitZoneChangeEvent(g, event)
	event.Kind = game.EventPermanentEnteredBattlefield
	emitEvent(g, event)
	return permanent, true
}
