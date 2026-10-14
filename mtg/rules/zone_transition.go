package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/zone"
)

func retainsCounters(g *game.Game, abilities []game.Ability, ctx conditionContext, from, to zone.Type) bool {
	if to == zone.Hand || to == zone.Library {
		return false
	}
	for _, ability := range abilities {
		provider, ok := ability.(game.ZoneTransitionPolicyProvider)
		if !ok {
			continue
		}
		policy, scope, condition := provider.ZoneTransitionDeclaration()
		if scope != zone.None && scope != from || !conditionSatisfied(g, ctx, condition) {
			continue
		}
		switch policy {
		case game.ZoneTransitionRetainCountersExceptHandLibrary:
			return true
		default:
			continue
		}
	}
	return false
}

// cardZoneAbilities evaluates the departing card, not the arriving permanent.
// Battlefield copy/ability-loss effects do not follow a card across CR 400.7.
func cardZoneAbilities(g *game.Game, card *game.CardInstance, from zone.Type, face game.FaceIndex, faceDown bool) []game.Ability {
	if card == nil || card.Def == nil {
		return nil
	}
	if faceDown {
		return nil
	}
	if from != zone.Stack {
		face = game.FaceFront
	}
	def, ok := card.Def.FaceView(face)
	if !ok {
		return nil
	}
	var abilities []game.Ability
	for i := range def.AbilityCount() {
		abilities = append(abilities, def.BodyAt(i))
	}
	var effects []game.ContinuousEffect
	for _, effect := range g.ContinuousEffects {
		if cardZoneAbilityEffectApplies(g, card, from, &effect) {
			effects = append(effects, effect)
		}
	}
	for _, source := range g.Battlefield {
		if !activeBattlefieldPermanent(source) {
			continue
		}
		controller := effectiveController(g, source)
		for _, ability := range permanentEffectiveAbilitiesView(g, source) {
			provider, ok := ability.(game.StaticContinuousEffectProvider)
			if !ok {
				continue
			}
			scope, condition, declarations := provider.StaticContinuousDeclarations()
			if scope != zone.None && scope != zone.Battlefield ||
				!conditionSatisfied(g, conditionContext{controller: controller, source: source}, condition) {
				continue
			}
			for _, declaration := range declarations {
				declaration.Controller = controller
				declaration.Timestamp = source.Timestamp()
				if cardZoneAbilityEffectApplies(g, card, from, &declaration) {
					effects = append(effects, declaration)
				}
			}
		}
	}
	for _, effect := range orderContinuousEffects(effects) {
		if effect.RemoveAllAbilities {
			abilities = nil
		}
		abilities = append(abilities, effect.AddAbilities...)
	}
	return abilities
}

func cardZoneAbilityEffectApplies(g *game.Game, card *game.CardInstance, from zone.Type, effect *game.ContinuousEffect) bool {
	if effect.CardZone != from || effect.Layer != game.LayerAbility {
		return false
	}

	subject := selectionSubject{kind: subjectCard, g: g, card: card, controller: card.Owner, viewer: effect.Controller}
	return effect.CardSelection == nil || matchSelection(&subject, effect.CardSelection)
}

func cardFaceDownInZone(g *game.Game, card *game.CardInstance, from zone.Type) bool {
	if cards, ok := destinationZone(g, card.Owner, from); ok {
		return cards.IsFaceDown(card.ID)
	}
	return false
}

func cardRetainsCounters(g *game.Game, card *game.CardInstance, from, to zone.Type, face game.FaceIndex, faceDown bool) bool {
	return retainsCounters(g, cardZoneAbilities(g, card, from, face, faceDown),
		conditionContext{controller: card.Owner}, from, to)
}

func retainedEntryCounters(g *game.Game, card *game.CardInstance, from zone.Type, face game.FaceIndex, faceDown bool) counter.Set {
	faceDown = faceDown || cardFaceDownInZone(g, card, from)
	set := g.CardCounters[card.ID]
	if set.IsEmpty() || !cardRetainsCounters(g, card, from, zone.Battlefield, face, faceDown) {
		return counter.Set{}
	}
	return cloneCounters(set)
}

func addRetainedCounters(permanent *game.Permanent, retained counter.Set) {
	for kind, n := range retained.All() {
		permanent.Counters.Add(kind, n)
	}
}

func storeDepartingCounters(g *game.Game, cardID id.ID, retained counter.Set) {
	delete(g.CardCounters, cardID)
	if !retained.IsEmpty() {
		if g.CardCounters == nil {
			g.CardCounters = make(map[id.ID]counter.Set)
		}
		g.CardCounters[cardID] = cloneCounters(retained)
	}
}

// Off-battlefield changes are finalized here by every zone operation, including
// Commander SBAs and library helpers. Entry/departure transfer is
// prepared separately, before removing the effective source.
func finalizeCardCounterTransition(g *game.Game, event game.Event) {
	if event.CardID == 0 {
		return
	}

	if event.ToZone == zone.Battlefield {
		delete(g.CardCounters, event.CardID)
		return
	}
	if event.FromZone == zone.Battlefield {
		return
	}
	set := g.CardCounters[event.CardID]
	if set.IsEmpty() {
		return
	}
	card, ok := g.GetCardInstance(event.CardID)
	if !ok || !cardRetainsCounters(g, card, event.FromZone, event.ToZone, event.Face, event.FaceDown) {
		delete(g.CardCounters, event.CardID)
	}
}

// Casting moves the card before costs are paid (CR 601.2a). Capture the source
// abilities before a cost can sacrifice an ability-suppressing permanent.
// The prior counter container also supports undoing a failed proposal (CR 728).
// Authorized casting from face-down exile turns the source face up before
// announcement (CR 406.3a), so the departing card has its printed abilities.
type castCounterTransition struct {
	cardID   id.ID
	previous counter.Set
	retained counter.Set
}

func prepareCastCounterTransition(g *game.Game, card *game.CardInstance, from zone.Type) castCounterTransition {
	transition := castCounterTransition{cardID: card.ID, previous: cloneCounters(g.CardCounters[card.ID])}
	if !transition.previous.IsEmpty() && cardRetainsCounters(g, card, from, zone.Stack, game.FaceFront, false) {
		transition.retained = cloneCounters(transition.previous)
	}
	return transition
}

func (t castCounterTransition) apply(g *game.Game) {
	storeDepartingCounters(g, t.cardID, t.retained)
}

func (t castCounterTransition) undo(g *game.Game) {
	storeDepartingCounters(g, t.cardID, t.previous)
}
