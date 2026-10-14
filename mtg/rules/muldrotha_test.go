package rules

import (
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

const muldrothaID = "e4625704-1d52-44e4-804f-2f45644d76ac"

func muldrothaGame(t *testing.T) (*game.Game, *Engine, *game.Permanent) {
	t.Helper()
	g, e, registry := lifetimeGame(t)
	source := addCombatPermanent(g, game.Player1, lifetimeDefinition(t, registry, muldrothaID))
	return g, e, source
}

func muldrothaPermission(source *game.Permanent, cardType types.Card) game.PlayPermission {
	return game.PlayPermission{SourceID: source.ObjectID, Key: "permanent-types", Type: cardType}
}

func graveyardCast(card id.ID, source *game.Permanent, cardType types.Card) action.Action {
	return action.CastSpellFromZone(card, zone.Graveyard, nil, 0, nil).WithPlayPermission(muldrothaPermission(source, cardType))
}

func permanentWitness(cardTypes ...types.Card) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name: "Permanent witness", Types: cardTypes, ManaCost: opt.Val(cost.Mana{}),
		Power: opt.Val(game.PT{Value: 2}), Toughness: opt.Val(game.PT{Value: 2}),
	}}
}

func TestMuldrothaProductionCastingChoicesSimulatorAndAllTypes(t *testing.T) {
	g, e, registry := lifetimeGame(t)
	def := lifetimeDefinition(t, registry, muldrothaID)
	card := addCardToHand(g, game.Player1, def)
	g.Players[game.Player1].ManaPool.Add(mana.C, 3)
	g.Players[game.Player1].ManaPool.Add(mana.B, 1)
	g.Players[game.Player1].ManaPool.Add(mana.G, 1)
	g.Players[game.Player1].ManaPool.Add(mana.U, 1)
	cast := action.CastSpell(card, nil, 0, nil)
	if !containsAction(e.Simulator().LegalActions(g, game.Player1), cast) || !e.applyAction(g, game.Player1, cast) {
		t.Fatal("production Muldrotha could not be cast for its actual cost")
	}
	e.resolveTopOfStack(g, &TurnLog{})
	source := permanentByCardID(g, card)
	if source == nil || effectivePermanentValues(g, source).power != 6 || effectivePermanentValues(g, source).toughness != 6 {
		t.Fatal("production Muldrotha did not enter as a 6/6")
	}
	hybrid := addCardToGraveyard(g, game.Player1, permanentWitness(types.Artifact, types.Creature, types.Kindred))
	artifact := graveyardCast(hybrid, source, types.Artifact)
	creature := graveyardCast(hybrid, source, types.Creature)
	legal := e.Simulator().LegalActions(g, game.Player1)
	if !containsAction(legal, artifact) || !containsAction(legal, creature) ||
		containsAction(legal, graveyardCast(hybrid, source, types.Kindred)) ||
		len(g.PermanentTypePlayUsesThisTurn) != 0 {
		t.Fatal("multi-type enumeration spent usage, omitted choices, or offered Kindred")
	}
	branch, ok := e.Simulator().Apply(g, game.Player1, artifact, [game.NumPlayers]PlayerAgent{})
	if !ok || len(g.PermanentTypePlayUsesThisTurn) != 0 ||
		!branch.PermanentTypePlayUsesThisTurn[game.PlayPermissionUse{Permission: muldrothaPermission(source, types.Artifact), Player: game.Player1}] ||
		branch.PermanentTypePlayUsesThisTurn[game.PlayPermissionUse{Permission: muldrothaPermission(source, types.Creature), Player: game.Player1}] {
		t.Fatal("Simulator did not spend exactly one type on an isolated branch")
	}
	if !e.applyAction(g, game.Player1, artifact) {
		t.Fatal("live Engine rejected Simulator's permission choice")
	}
	e.resolveTopOfStack(g, &TurnLog{})
	second := addCardToGraveyard(g, game.Player1, permanentWitness(types.Artifact, types.Creature))
	if containsAction(e.legalActions(g, game.Player1), graveyardCast(second, source, types.Artifact)) ||
		!e.applyAction(g, game.Player1, graveyardCast(second, source, types.Creature)) {
		t.Fatal("artifact creature consumed both types or allowed reuse")
	}
	e.resolveTopOfStack(g, &TurnLog{})
	third := addCardToGraveyard(g, game.Player1, permanentWitness(types.Artifact, types.Creature))
	if e.applyAction(g, game.Player1, graveyardCast(third, source, types.Artifact)) ||
		e.applyAction(g, game.Player1, graveyardCast(third, source, types.Creature)) {
		t.Fatal("used type could be spent again")
	}
	for _, cardType := range []types.Card{types.Enchantment, types.Planeswalker, types.Battle} {
		spell := addCardToGraveyard(g, game.Player1, permanentWitness(cardType))
		act := graveyardCast(spell, source, cardType)
		if !containsAction(e.legalActions(g, game.Player1), act) || !e.applyAction(g, game.Player1, act) {
			t.Fatalf("unused %s was not available", cardType)
		}
		obj, _ := g.Stack.Peek()
		if !counterStackObject(g, obj.ID) || e.applyAction(g, game.Player1, act) {
			t.Fatalf("successful %s cast did not consume usage after counter", cardType)
		}
	}
	land := addCardToGraveyard(g, game.Player1, permanentWitness(types.Land))
	play := action.PlayLandFaceFromZone(land, zone.Graveyard, game.FaceFront).WithPlayPermission(muldrothaPermission(source, types.Land))
	if !containsAction(e.legalActions(g, game.Player1), play) || !e.applyAction(g, game.Player1, play) ||
		g.Turn.LandsPlayedThisTurn != 1 || len(g.PermanentTypePlayUsesThisTurn) != 6 {
		t.Fatal("land did not use the ordinary land play and its own allowance")
	}
	for _, cardType := range []types.Card{types.Instant, types.Sorcery} {
		spell := addCardToGraveyard(g, game.Player1, permanentWitness(cardType))
		if e.applyAction(g, game.Player1, graveyardCast(spell, source, cardType)) {
			t.Fatal("nonpermanent spell received a permanent-type permission")
		}
	}
}

func TestMuldrothaLandLimitsTimingAndArtifactLandChoice(t *testing.T) {
	g, e, source := muldrothaGame(t)
	land := addCardToGraveyard(g, game.Player1, permanentWitness(types.Artifact, types.Land))
	play := action.PlayLandFaceFromZone(land, zone.Graveyard, game.FaceFront)
	artifact := play.WithPlayPermission(muldrothaPermission(source, types.Artifact))
	if !containsAction(e.legalActions(g, game.Player1), artifact) ||
		!containsAction(e.legalActions(g, game.Player1), play.WithPlayPermission(muldrothaPermission(source, types.Land))) {
		t.Fatal("artifact land did not offer both permanent type choices")
	}
	g.Turn.LandsPlayedThisTurn = 1
	if e.applyAction(g, game.Player1, artifact) || len(g.PermanentTypePlayUsesThisTurn) != 0 {
		t.Fatal("permission bypassed land limit or failed play spent quota")
	}
	g.Turn.LandsPlayedThisTurn = 0
	g.Turn.Phase = game.PhaseEnding
	if e.applyAction(g, game.Player1, artifact) {
		t.Fatal("permission bypassed land timing")
	}
	g.Turn.Phase = game.PhasePrecombatMain
	if !e.applyAction(g, game.Player1, artifact) {
		t.Fatal("legal artifact land play failed")
	}
	g.RuleEffects = append(g.RuleEffects, game.RuleEffect{
		Kind: game.RuleEffectAdditionalLandPlays, Controller: game.Player1,
		AffectedPlayer: game.PlayerYou, AdditionalLandPlays: 2,
	})
	next := addCardToGraveyard(g, game.Player1, permanentWitness(types.Land))
	if !e.applyAction(g, game.Player1, action.PlayLandFaceFromZone(next, zone.Graveyard, game.FaceFront).WithPlayPermission(muldrothaPermission(source, types.Land))) {
		t.Fatal("unchosen land type was consumed by artifact land")
	}
	third := addCardToGraveyard(g, game.Player1, permanentWitness(types.Land))
	if e.applyAction(g, game.Player1, action.PlayLandFaceFromZone(third, zone.Graveyard, game.FaceFront)) {
		t.Fatal("additional land allowance incorrectly reset per-source quota")
	}
	hand := addCardToHand(g, game.Player1, permanentWitness(types.Land))
	if !e.applyAction(g, game.Player1, action.PlayLand(hand)) {
		t.Fatal("graveyard quota affected ordinary hand land plays")
	}
}

type muldrothaChoiceAgent struct {
	requests []game.ChoiceRequest
	choose   func(game.ChoiceRequest) []int
}

func (*muldrothaChoiceAgent) ChooseAction(_ PlayerObservation, _ []action.Action) action.Action {
	return action.Pass()
}

func (a *muldrothaChoiceAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	a.requests = append(a.requests, request)
	if a.choose != nil {
		return a.choose(request)
	}
	return request.DefaultSelection
}

func TestMuldrothaFailedPaymentIllegalAnnouncementsAndDirectChoice(t *testing.T) {
	g, e, source := muldrothaGame(t)
	def := permanentWitness(types.Artifact, types.Creature)
	def.ManaCost = opt.Val(cost.Mana{cost.O(1)})
	def.AlternativeCosts = []cost.Alternative{{Label: "Witness alternative", ManaCost: opt.Val(cost.Mana{cost.O(2)})}}
	card := addCardToGraveyard(g, game.Player1, def)
	valid := graveyardCast(card, source, types.Artifact)
	for _, bad := range []game.PlayPermission{
		{SourceID: source.ObjectID + 100, Key: "permanent-types", Type: types.Artifact},
		{SourceID: source.ObjectID, Key: "wrong", Type: types.Artifact},
		muldrothaPermission(source, types.Enchantment),
	} {
		if e.applyAction(g, game.Player1, valid.WithPlayPermission(bad)) {
			t.Fatal("fabricated permission accepted")
		}
	}
	if e.applyAction(g, game.Player1, valid) || len(g.PermanentTypePlayUsesThisTurn) != 0 {
		t.Fatal("unpayable proposal spent allowance")
	}
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)
	agent := &muldrothaChoiceAgent{choose: func(request game.ChoiceRequest) []int {
		if request.Kind == game.ChoicePayment {
			g.Players[game.Player1].ManaPool.Empty()
		}
		return request.DefaultSelection
	}}
	agents := [game.NumPlayers]PlayerAgent{game.Player1: agent}
	if e.applyActionWithChoices(g, game.Player1, valid, agents, &TurnLog{}) ||
		!g.Players[game.Player1].Graveyard.Contains(card) || g.Stack.Size() != 0 ||
		len(g.PermanentTypePlayUsesThisTurn) != 0 {
		t.Fatal("late failed payment did not roll back without spending permission")
	}
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	agent.choose = func(request game.ChoiceRequest) []int {
		for i, option := range request.Options {
			if strings.Contains(option.Label, "Creature permission") {
				return []int{i}
			}
		}
		return request.DefaultSelection
	}
	if !e.applyActionWithChoices(g, game.Player1, action.CastSpellFromZone(card, zone.Graveyard, nil, 0, nil), agents, &TurnLog{}) ||
		!g.PermanentTypePlayUsesThisTurn[game.PlayPermissionUse{Permission: muldrothaPermission(source, types.Creature), Player: game.Player1}] ||
		g.PermanentTypePlayUsesThisTurn[game.PlayPermissionUse{Permission: muldrothaPermission(source, types.Artifact), Player: game.Player1}] {
		t.Fatal("direct Engine cast did not offer the player's source/type choice")
	}
}

func TestMuldrothaTimingProhibitionsAndOwnership(t *testing.T) {
	for _, restriction := range []string{"other turn", "end step", "stack occupied", "spell prohibited", "zone prohibited", "opponent graveyard", "no priority"} {
		t.Run(restriction, func(t *testing.T) {
			g, e, source := muldrothaGame(t)
			card := addCardToGraveyard(g, game.Player1, permanentWitness(types.Creature))
			act := graveyardCast(card, source, types.Creature)
			switch restriction {
			case "other turn":
				g.Turn.ActivePlayer = game.Player2
			case "end step":
				g.Turn.Phase = game.PhaseEnding
			case "stack occupied":
				g.Stack.Push(&game.StackObject{ID: g.IDGen.Next(), Kind: game.StackTriggeredAbility})
			case "spell prohibited":
				g.RuleEffects = append(g.RuleEffects, game.RuleEffect{Kind: game.RuleEffectCantCastSpells, Controller: game.Player1, AffectedPlayer: game.PlayerYou})
			case "zone prohibited":
				g.RuleEffects = append(g.RuleEffects, game.RuleEffect{Kind: game.RuleEffectCantCastFromZones, Controller: game.Player1, AffectedPlayer: game.PlayerYou, CantCastFromZones: []zone.Type{zone.Graveyard}})
			case "opponent graveyard":
				g.Players[game.Player1].Graveyard.Remove(card)
				g.Players[game.Player2].Graveyard.Add(card)
				instance, _ := g.GetCardInstance(card)
				instance.Owner = game.Player2
			case "no priority":
				g.Turn.PriorityPlayer = game.Player2
			default:
				t.Fatalf("unknown restriction %q", restriction)
			}
			if containsAction(e.legalActions(g, game.Player1), act) || e.applyAction(g, game.Player1, act) ||
				len(g.PermanentTypePlayUsesThisTurn) != 0 {
				t.Fatal("permission bypassed restriction or illegal action spent allowance")
			}
		})
	}
	g, e, source := muldrothaGame(t)
	def := permanentWitness(types.Creature)
	def.StaticAbilities = []game.StaticAbility{game.FlashStaticBody}
	card := addCardToGraveyard(g, game.Player1, def)
	g.Turn.Phase = game.PhaseEnding
	if !e.applyAction(g, game.Player1, graveyardCast(card, source, types.Creature)) {
		t.Fatal("normal flash permission was not respected")
	}
}

func TestMuldrothaSourcesControlBlinkTurnsAndClone(t *testing.T) {
	g, e, source := muldrothaGame(t)
	secondDef := lifetimeDefinition(t, cards.NewDefaultRegistry(), muldrothaID)
	secondDef.Supertypes = nil // A nonlegendary copy can coexist with Muldrotha.
	second := addCombatPermanent(g, game.Player1, secondDef)
	firstCard := addCardToGraveyard(g, game.Player1, permanentWitness(types.Creature))
	if !e.applyAction(g, game.Player1, graveyardCast(firstCard, source, types.Creature)) {
		t.Fatal("first source failed")
	}
	obj, _ := g.Stack.Peek()
	counterStackObject(g, obj.ID)
	if !e.applyAction(g, game.Player1, graveyardCast(firstCard, second, types.Creature)) {
		t.Fatal("second source shared the first source's allowance")
	}
	obj, _ = g.Stack.Peek()
	counterStackObject(g, obj.ID)
	before := g.Clone()
	source.Controller = game.Player2
	if len(permanentTypePlayPermissions(g, game.Player1, zone.Graveyard, permanentWitness(types.Creature))) != 0 {
		t.Fatal("lost control retained permission")
	}
	// During this turn player 2 cannot use the permission, even on a source
	// with no usage for them. Returning control doesn't reset player 1's use.
	if len(permanentTypePlayPermissions(g, game.Player2, zone.Graveyard, permanentWitness(types.Creature))) != 0 {
		t.Fatal("control change ignored controller-turn restriction")
	}
	source.Controller = game.Player1
	if e.applyAction(g, game.Player1, graveyardCast(firstCard, source, types.Creature)) {
		t.Fatal("regaining the same source reset usage")
	}
	movePermanentToZone(g, source, zone.Exile)
	card, _ := g.GetCardInstance(source.CardInstanceID)
	g.Players[game.Player1].Exile.Remove(card.ID)
	reentered, ok := createCardPermanent(g, card, game.Player1, zone.Exile)
	if !ok || reentered.ObjectID == source.ObjectID ||
		!e.applyAction(g, game.Player1, graveyardCast(firstCard, reentered, types.Creature)) {
		t.Fatal("new incarnation did not get a new allowance")
	}
	if e.applyAction(before, game.Player1, graveyardCast(firstCard, source, types.Creature)) {
		t.Fatal("clone lost preexisting usage")
	}
	obj, _ = g.Stack.Peek()
	counterStackObject(g, obj.ID)
	oldTurn := g.Turn.TurnNumber
	g.Turn.ExtraTurns = append(g.Turn.ExtraTurns, game.Player1)
	e.advanceToNextTurn(g)
	g.Turn.Phase, g.Turn.Step, g.Turn.PriorityPlayer = game.PhasePrecombatMain, game.StepNone, game.Player1
	if g.Turn.ActivePlayer != game.Player1 || g.Turn.TurnNumber != oldTurn+1 ||
		len(g.PermanentTypePlayUsesThisTurn) != 0 ||
		!e.applyAction(g, game.Player1, graveyardCast(firstCard, reentered, types.Creature)) {
		t.Fatal("extra turn did not reset source usage")
	}
	obj, _ = g.Stack.Peek()
	counterStackObject(g, obj.ID)
	e.advanceToNextTurn(g)
	reentered.Controller = game.Player2
	g.Turn.Phase, g.Turn.Step, g.Turn.PriorityPlayer = game.PhasePrecombatMain, game.StepNone, game.Player2
	other := addCardToGraveyard(g, game.Player2, permanentWitness(types.Creature))
	if len(g.PermanentTypePlayUsesThisTurn) != 0 ||
		!e.applyAction(g, game.Player2, graveyardCast(other, reentered, types.Creature)) {
		t.Fatal("normal next turn/current controller allowance failed")
	}
	// Shared operation explicitly isolates players, even on the same source.
	permission := muldrothaPermission(reentered, types.Artifact)
	recordPermanentTypePlay(g, game.Player1, permission)
	if !slices.Contains(permanentTypePlayPermissions(g, game.Player2, zone.Graveyard, permanentWitness(types.Artifact)), permission) {
		t.Fatal("source usage was global rather than per player")
	}
}

func TestMuldrothaEffectiveCopiedGrantedAndRemovedAbilities(t *testing.T) {
	g, e, source := muldrothaGame(t)
	host := addCombatPermanent(g, game.Player1, permanentWitness(types.Artifact))
	becomeStaticCopy(t, g, host, source, false)
	movePermanentToZone(g, source, zone.Exile)
	card := addCardToGraveyard(g, game.Player1, permanentWitness(types.Creature))
	act := graveyardCast(card, host, types.Creature)
	if !containsAction(e.legalActions(g, game.Player1), act) {
		t.Fatal("copied static body did not supply permission")
	}
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID: g.IDGen.Next(), Timestamp: 100000, Layer: game.LayerAbility,
		AffectedObjectID: host.ObjectID, RemoveAllAbilities: true,
	})
	if containsAction(e.legalActions(g, game.Player1), act) || e.applyAction(g, game.Player1, act) {
		t.Fatal("removed copied permission still functioned")
	}
	def := lifetimeDefinition(t, cards.NewDefaultRegistry(), muldrothaID)
	body := def.StaticAbilities[0]
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID: g.IDGen.Next(), Timestamp: 200000, Layer: game.LayerAbility,
		AffectedObjectID: host.ObjectID, AddAbilities: []game.Ability{&body},
	})
	if !e.applyAction(g, game.Player1, act) {
		t.Fatal("externally granted effective permission was ignored")
	}
	obj, _ := g.Stack.Peek()
	counterStackObject(g, obj.ID)
	// Layer-6 loss/restoration, unlike blink, retains the source's usage.
	if e.applyAction(g, game.Player1, act) {
		t.Fatal("restoring permission reset source usage")
	}
}

func TestMuldrothaBestowAlternativeCostsAndSourceSacrifice(t *testing.T) {
	g, e, source := muldrothaGame(t)
	card := addCardToGraveyard(g, game.Player1, bestowTestCreature())
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	targets := []game.Target{game.PermanentTarget(source.ObjectID)}
	bestow := action.CastBestowSpellFaceFromZone(card, zone.Graveyard, game.FaceFront, targets, 0, nil)
	enchantment := bestow.WithPlayPermission(muldrothaPermission(source, types.Enchantment))
	if containsAction(e.legalActions(g, game.Player1), bestow.WithPlayPermission(muldrothaPermission(source, types.Creature))) ||
		e.applyAction(g, game.Player1, bestow.WithPlayPermission(muldrothaPermission(source, types.Creature))) {
		t.Fatal("bestow permission used printed creature types")
	}
	if !e.applyAction(g, game.Player1, enchantment) {
		t.Fatal("legal bestow cast under enchantment permission failed")
	}
	e.resolveTopOfStack(g, &TurnLog{})
	creature := addCardToGraveyard(g, game.Player1, permanentWitness(types.Creature))
	if !e.applyAction(g, game.Player1, graveyardCast(creature, source, types.Creature)) {
		t.Fatal("bestow consumed the creature allowance")
	}
	e.resolveTopOfStack(g, &TurnLog{})
	def := permanentWitness(types.Artifact)
	def.ManaCost = opt.Val(cost.Mana{cost.O(9)})
	def.AlternativeCosts = []cost.Alternative{{Label: "Witness free alternative", ManaCost: opt.Val(cost.Mana{})}}
	def.AdditionalCosts = []cost.Additional{{Kind: cost.AdditionalSacrifice, Amount: 1, MatchPermanentType: true, PermanentType: types.Creature}}
	relic := addCardToGraveyard(g, game.Player1, def)
	agent := &muldrothaChoiceAgent{choose: func(request game.ChoiceRequest) []int {
		for i, option := range request.Options {
			if option.Card.Exists && option.Card.Val.CardID == source.CardInstanceID {
				return []int{i}
			}
		}
		return request.DefaultSelection
	}}
	if !e.applyActionWithChoices(g, game.Player1, graveyardCast(relic, source, types.Artifact), [game.NumPlayers]PlayerAgent{game.Player1: agent}, &TurnLog{}) ||
		permanentByCardID(g, source.CardInstanceID) != nil ||
		!g.PermanentTypePlayUsesThisTurn[game.PlayPermissionUse{Permission: muldrothaPermission(source, types.Artifact), Player: game.Player1}] {
		t.Fatal("ordinary alternative/additional costs failed when source was sacrificed during payment")
	}
}

func TestMuldrothaIndependentPermissionAndCopiesDoNotSpendQuota(t *testing.T) {
	g, e, source := muldrothaGame(t)
	def := permanentWitness(types.Artifact)
	card := addCardToGraveyard(g, game.Player1, def)
	g.RuleEffects = append(g.RuleEffects, game.RuleEffect{
		Kind: game.RuleEffectCastFromZone, Controller: game.Player1, AffectedPlayer: game.PlayerYou,
		AffectedCardID: card, CastFromZone: zone.Graveyard,
	})
	plain := action.CastSpellFromZone(card, zone.Graveyard, nil, 0, nil)
	scoped := graveyardCast(card, source, types.Artifact)
	if !containsAction(e.legalActions(g, game.Player1), plain) || !containsAction(e.legalActions(g, game.Player1), scoped) ||
		!e.applyAction(g, game.Player1, plain) || len(g.PermanentTypePlayUsesThisTurn) != 0 {
		t.Fatal("independent graveyard permission was not separately selectable")
	}
	obj, _ := g.Stack.Peek()
	copyObject := *obj
	copyObject.ID, copyObject.Copy = g.IDGen.Next(), true
	g.Stack.Push(&copyObject)
	e.resolveTopOfStack(g, &TurnLog{})
	if len(g.PermanentTypePlayUsesThisTurn) != 0 {
		t.Fatal("resolving a spell copy inferred a graveyard cast")
	}
	counterStackObject(g, obj.ID)
	if !e.applyAction(g, game.Player1, scoped) || len(g.PermanentTypePlayUsesThisTurn) != 1 {
		t.Fatal("choosing scoped permission did not consume its type")
	}
}

func TestMuldrothaFaceDownUsesAnnouncedCreatureTypeAndPlanner(t *testing.T) {
	g, e, source := muldrothaGame(t)
	def := morphCreature(cost.Mana{cost.O(2)})
	def.Types = []types.Card{types.Land}
	card := addCardToGraveyard(g, game.Player1, def)
	permission := muldrothaPermission(source, types.Creature)
	cast := action.CastFaceDownFromZone(card, zone.Graveyard, game.FaceFront, game.FaceDownMorph).WithPlayPermission(permission)
	g.Players[game.Player1].ManaPool.Add(mana.C, 3)
	g.RuleEffects = append(g.RuleEffects, game.RuleEffect{
		Kind: game.RuleEffectCostModifier, Controller: game.Player1, AffectedPlayer: game.PlayerYou,
		CostModifier: game.CostModifier{Kind: game.CostModifierSpell, GenericIncrease: 1},
	})
	if containsAction(e.legalActions(g, game.Player1), cast) || e.applyAction(g, game.Player1, cast) {
		t.Fatal("face-down graveyard cast bypassed cost increase")
	}
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	for range 4 {
		g.Players[game.Player1].ManaRiders = append(g.Players[game.Player1].ManaRiders, game.ManaRiderInstance{
			Unit: mana.Unit{Color: mana.C}, Controller: game.Player1,
			Rider: game.ManaSpendRider{Condition: game.ManaSpendCastOrActivateCreature, Restriction: game.ManaSpendRestrictedToCondition},
		})
	}
	if !containsAction(e.Simulator().LegalActions(g, game.Player1), cast) {
		t.Fatalf("face down missing from legal actions: %+v; can cast = %v", e.legalActions(g, game.Player1), canCastGraveyardFaceDown(g, game.Player1, card, game.FaceFront, game.FaceDownMorph, permission))
	}
	if !e.applyAction(g, game.Player1, cast) {
		t.Fatal("legal face-down cast failed application")
	}
	if g.Players[game.Player1].ManaPool.Total() != 0 || len(g.Players[game.Player1].ManaRiders) != 0 ||
		!g.PermanentTypePlayUsesThisTurn[game.PlayPermissionUse{Permission: permission, Player: game.Player1}] ||
		g.PermanentTypePlayUsesThisTurn[game.PlayPermissionUse{Permission: muldrothaPermission(source, types.Land), Player: game.Player1}] {
		t.Fatal("morph of a land did not use only the announced creature type and full payment")
	}
	obj, _ := g.Stack.Peek()
	if !obj.FaceDown || obj.SourceZone != zone.Graveyard {
		t.Fatal("graveyard face-down spell lost its origin/characteristics")
	}
	e.resolveTopOfStack(g, &TurnLog{})
	permanent := permanentByCardID(g, card)
	if permanent == nil || !permanent.FaceDown || !permanentHasType(g, permanent, types.Creature) {
		t.Fatal("morph did not resolve as a face-down creature")
	}
	land := addCardToGraveyard(g, game.Player1, permanentWitness(types.Land))
	if !e.applyAction(g, game.Player1, action.PlayLandFaceFromZone(land, zone.Graveyard, game.FaceFront).WithPlayPermission(muldrothaPermission(source, types.Land))) {
		t.Fatal("morph incorrectly consumed land allowance")
	}
}

func TestMuldrothaMutateAndFailedTargets(t *testing.T) {
	g, e, source := muldrothaGame(t)
	card := addCardToGraveyard(g, game.Player1, mutateCard())
	g.Players[game.Player1].ManaPool.Add(mana.G, 1)
	cast := action.CastMutateSpellFromZone(card, zone.Graveyard, source.ObjectID).WithPlayPermission(muldrothaPermission(source, types.Creature))
	invalid := action.CastMutateSpellFromZone(card, zone.Graveyard, source.ObjectID+100).WithPlayPermission(muldrothaPermission(source, types.Creature))
	if e.applyAction(g, game.Player1, invalid) || len(g.PermanentTypePlayUsesThisTurn) != 0 ||
		!containsAction(e.legalActions(g, game.Player1), cast) || !e.applyAction(g, game.Player1, cast) {
		t.Fatal("mutate alternative bypassed targets, consumed a failed proposal, or lost graveyard permission")
	}
	e.resolveTopOfStack(g, &TurnLog{})
	if len(g.PermanentTypePlayUsesThisTurn) != 1 || len(source.MergedCards) != 1 {
		t.Fatal("graveyard mutate did not resolve through the existing merged-permanent path")
	}

	g, e, source = muldrothaGame(t)
	card = addCardToGraveyard(g, game.Player1, mutateCard())
	recordPermanentTypePlay(g, game.Player1, muldrothaPermission(source, types.Creature))
	g.RuleEffects = append(g.RuleEffects, game.RuleEffect{
		Kind: game.RuleEffectCastFromZone, Controller: game.Player1, AffectedPlayer: game.PlayerYou,
		AffectedCardID: card, CastFromZone: zone.Graveyard,
	})
	g.Players[game.Player1].ManaPool.Add(mana.G, 1)
	independent := action.CastMutateSpellFromZone(card, zone.Graveyard, source.ObjectID)
	branch, ok := e.Simulator().Apply(g, game.Player1, independent, [game.NumPlayers]PlayerAgent{})
	if !ok || len(branch.PermanentTypePlayUsesThisTurn) != 1 ||
		!g.Players[game.Player1].Graveyard.Contains(card) {
		t.Fatal("independent mutate checked normal cost or incorrectly spent exhausted scoped allowance")
	}
}
