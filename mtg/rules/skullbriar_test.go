package rules

import (
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

const skullbriarID = "088ac883-6f44-4046-bf19-81c7fd952e55"

func skullbriarGame(t *testing.T) (*game.Game, *Engine, *game.Permanent) {
	t.Helper()
	g, e, registry := lifetimeGame(t)
	return g, e, addCombatPermanent(g, game.Player1, lifetimeDefinition(t, registry, skullbriarID))
}

func skullCounters() counter.Set {
	var set counter.Set
	set.Add(counter.PlusOnePlusOne, 3)
	set.Add(counter.Charge, 2)
	set.Add(counter.Flying, 1)
	set.Add(counter.Lore, 2)
	return set
}

func assertSkullCounters(t *testing.T, got counter.Set) {
	t.Helper()
	want := skullCounters()
	if !reflect.DeepEqual(got.All(), want.All()) {
		t.Fatalf("counters = %v, want %v", got.All(), want.All())
	}
}

func TestSkullbriarProductionCombatTriggerAndSimulator(t *testing.T) {
	g, e, registry := lifetimeGame(t)
	def := lifetimeDefinition(t, registry, skullbriarID)
	card := addCardToHand(g, game.Player1, def)
	g.Players[game.Player1].ManaPool.Add(mana.B, 1)
	g.Players[game.Player1].ManaPool.Add(mana.G, 1)
	cast := action.CastSpell(card, nil, 0, nil)
	if !containsAction(e.Simulator().LegalActions(g, game.Player1), cast) || !e.applyAction(g, game.Player1, cast) {
		t.Fatal("production Skullbriar could not be cast for {B}{G}")
	}
	branch := e.Simulator().ResolvePriority(g, [game.NumPlayers]PlayerAgent{})
	source := permanentByCardID(branch, card)
	if source == nil || !hasKeyword(branch, source, game.Haste) || permanentByCardID(g, card) != nil {
		t.Fatal("Simulator failed isolated production resolution")
	}
	e.resolveTopOfStack(g, &TurnLog{})
	source = permanentByCardID(g, card)
	if source == nil || !hasKeyword(g, source, game.Haste) {
		t.Fatal("haste did not function on the newly cast creature")
	}
	g.Combat = &game.CombatState{Attackers: []game.AttackDeclaration{{
		Attacker: source.ObjectID, Target: game.AttackTarget{Player: game.Player2},
	}}}
	log := &TurnLog{}
	e.resolveCombatDamage(g, log)
	if g.Players[game.Player2].Life != 39 || !e.putTriggeredAbilitiesOnStack(g) || g.Stack.Size() != 1 {
		t.Fatal("actual player combat damage did not queue the production trigger")
	}
	e.resolveTopOfStack(g, log)
	if source.Counters.Get(counter.PlusOnePlusOne) != 1 || effectivePower(g, source) != 2 {
		t.Fatal("ordinary AddCounter trigger did not resolve")
	}
	dealPlayerDamage(g, source.CardInstanceID, source.ObjectID, game.Player1, game.Player2, 4, false)
	if e.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("noncombat damage incorrectly triggered Skullbriar")
	}
	// The ruling explicitly includes damage redirected to Skullbriar's controller.
	dealPlayerDamage(g, source.CardInstanceID, source.ObjectID, game.Player1, game.Player1, 7, true)
	if !e.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("combat damage to its controller did not trigger")
	}
	e.resolveTopOfStack(g, log)
	if source.Counters.Get(counter.PlusOnePlusOne) != 2 {
		t.Fatal("trigger scaled by damage instead of placing exactly one counter")
	}
}

func TestSkullbriarDirectZoneCircuitAndNewIdentity(t *testing.T) {
	t.Parallel()
	g, e, source := skullbriarGame(t)
	source.Counters = skullCounters()
	oldID := source.ObjectID
	oldCard, _ := g.GetCardInstance(source.CardInstanceID)
	oldVersion := oldCard.ZoneVersion
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		AffectedObjectID: oldID, Layer: game.LayerPowerToughnessModify, PowerDelta: 9, ToughnessDelta: 9,
	})
	linked := game.LinkedObjectRef{ObjectID: oldID, CardID: oldCard.ID, CardZoneVersion: oldVersion, CardZoneVersionSet: true}
	if !movePermanentToZone(g, source, zone.Graveyard) {
		t.Fatal("battlefield departure failed")
	}
	assertSkullCounters(t, g.CardCounters[oldCard.ID])
	snapshot, _ := lastKnownObject(g, oldID)
	assertSkullCounters(t, snapshot.Counters)
	if snapshot.Power.Val != 13 {
		t.Fatal("counter retention corrupted departure LKI")
	}
	for _, change := range [][2]zone.Type{{zone.Graveyard, zone.Exile}, {zone.Exile, zone.Command}, {zone.Command, zone.Graveyard}} {
		if !moveCardBetweenZones(g, game.Player1, oldCard.ID, change[0], change[1]) {
			t.Fatalf("move %v failed", change)
		}
		assertSkullCounters(t, g.CardCounters[oldCard.ID])
	}
	g.Players[game.Player1].Graveyard.Remove(oldCard.ID)
	eventsBefore := len(g.Events)
	entrant, ok := createCardPermanent(g, oldCard, game.Player1, zone.Graveyard)
	if !ok || entrant.ObjectID == oldID || oldCard.ZoneVersion <= oldVersion ||
		effectivePower(g, entrant) != 4 || !hasKeyword(g, entrant, game.Flying) {
		t.Fatal("reentry did not retain counters on a genuinely new object")
	}
	assertSkullCounters(t, entrant.Counters)
	if _, exists := g.CardCounters[oldCard.ID]; exists {
		t.Fatal("counter containers duplicated between card and permanent")
	}
	for _, event := range g.Events[eventsBefore:] {
		if event.Kind == game.EventCountersAdded {
			t.Fatal("retained counters were logged as placed/removed")
		}
	}
	if linkedCardInstanceExists(g, linked) {
		t.Fatal("zone-incarnation-bound card link followed reentry")
	}
	stale := &game.StackObject{SourceID: oldID, Controller: game.Player1}
	resolveInstruction(e, g, stale, game.AddCounter{Object: game.SourcePermanentReference(), CounterKind: counter.Charge, Amount: game.Fixed(7)}, nil)
	stale.Targets = []game.Target{game.PermanentTarget(oldID)}
	resolveInstruction(e, g, stale, game.AddCounter{Object: game.TargetPermanentReference(0), CounterKind: counter.Charge, Amount: game.Fixed(7)}, nil)
	rememberLinkedObject(g, linkedObjectSourceKey(g, stale, "stale"), linked)
	resolveInstruction(e, g, stale, game.AddCounter{Object: game.LinkedObjectReference("stale"), CounterKind: counter.Charge, Amount: game.Fixed(7)}, nil)
	assertSkullCounters(t, entrant.Counters)
	for _, destination := range []zone.Type{zone.Hand, zone.Library} {
		if !movePermanentToZone(g, entrant, destination) {
			t.Fatal("hand/library arrival failed")
		}
		set := g.CardCounters[oldCard.ID]
		if !set.IsEmpty() {
			t.Fatal("actual hand/library arrival retained counters")
		}
		from, _ := destinationZone(g, game.Player1, destination)
		from.Remove(oldCard.ID)
		entrant, ok = createCardPermanent(g, oldCard, game.Player1, destination)
		if !ok || !entrant.Counters.IsEmpty() {
			t.Fatal("lost counters reappeared")
		}
		entrant.Counters = skullCounters()
	}
}

func linkedCardInstanceExists(g *game.Game, ref game.LinkedObjectRef) bool {
	_, ok := linkedCardInstance(g, ref)
	return ok
}

func TestSkullbriarCommanderSBAStackCounterAndRecast(t *testing.T) {
	t.Parallel()
	for _, destination := range []zone.Type{zone.Graveyard, zone.Exile, zone.Hand, zone.Library} {
		t.Run(destination.String(), func(t *testing.T) {
			t.Parallel()
			g, e, source := skullbriarGame(t)
			source.Counters = skullCounters()
			cardID := source.CardInstanceID
			trackCommanderID(g, game.Player1, cardID)
			if !movePermanentToZone(g, source, destination) {
				t.Fatal("commander departure failed")
			}
			e.applyStateBasedActions(g)
			if !g.Players[game.Player1].CommandZone.Contains(cardID) {
				t.Fatal("commander did not reach command by replacement/SBA")
			}
			assertSkullCounters(t, g.CardCounters[cardID])
			g.Players[game.Player1].ManaPool.Add(mana.B, 2)
			g.Players[game.Player1].ManaPool.Add(mana.G, 2)
			g.Players[game.Player1].ManaPool.Add(mana.C, 2)
			cast := action.CastSpellFromZone(cardID, zone.Command, nil, 0, nil)
			branch, ok := e.Simulator().Apply(g, game.Player1, cast, [game.NumPlayers]PlayerAgent{})
			if !ok || g.Stack.Size() != 0 {
				t.Fatal("Simulator command cast was not isolated")
			}
			assertSkullCounters(t, branch.CardCounters[cardID])
			if !e.applyAction(g, game.Player1, cast) {
				t.Fatal("command cast failed")
			}
			obj, _ := g.Stack.Peek()
			stackSnapshot := snapshotStackSpell(g, obj)
			assertSkullCounters(t, stackSnapshot.Counters)
			if stackSnapshot.Power.Val != 4 || !counterStackObject(g, obj.ID) {
				t.Fatal("countered permanent spell lost its stack characteristics")
			}
			assertSkullCounters(t, g.CardCounters[cardID])
			e.applyStateBasedActions(g)
			if !e.applyAction(g, game.Player1, cast) {
				t.Fatal("recast from command with ordinary commander tax failed")
			}
			e.resolveTopOfStack(g, &TurnLog{})
			assertSkullCounters(t, permanentByCardID(g, cardID).Counters)
		})
	}
}

func TestSkullbriarActualReplacementDestination(t *testing.T) {
	t.Parallel()
	g, _, source := skullbriarGame(t)
	source.Counters = skullCounters()
	graveyardRedirectCounterPermanent(g, game.Player2, counter.Void)
	if !movePermanentToZone(g, source, zone.Graveyard) || !g.Players[game.Player1].Exile.Contains(source.CardInstanceID) {
		t.Fatal("graveyard redirect failed")
	}
	set := g.CardCounters[source.CardInstanceID]
	if set.Get(counter.PlusOnePlusOne) != 3 || set.Get(counter.Void) != 1 {
		t.Fatal("redirect did not preserve old counters and place its new marker")
	}
	if !moveCardBetweenZones(g, game.Player1, source.CardInstanceID, zone.Exile, zone.Graveyard) {
		t.Fatal("off-battlefield redirect failed")
	}
	set = g.CardCounters[source.CardInstanceID]
	if set.Get(counter.PlusOnePlusOne) != 3 || set.Get(counter.Void) != 2 {
		t.Fatal("same-destination exile incarnation lost retained or replacement counters")
	}
	// Counter-and-bounce uses the actual hand destination, not graveyard.
	g.Players[game.Player1].Exile.Remove(source.CardInstanceID)
	card, _ := g.GetCardInstance(source.CardInstanceID)
	obj := &game.StackObject{ID: g.IDGen.Next(), Kind: game.StackSpell, SourceID: card.ID, Controller: game.Player1}
	emitZoneChangeEvent(g, game.Event{CardID: card.ID, FromZone: zone.Exile, ToZone: zone.Stack})
	g.Stack.Push(obj)
	if !counterStackObjectWithDestination(g, obj.ID, false, game.CounteredSpellHand) {
		t.Fatal("stack bounce failed")
	}
	set = g.CardCounters[card.ID]
	if !g.Players[game.Player1].Hand.Contains(card.ID) || !set.IsEmpty() {
		t.Fatal("actual stack-to-hand destination retained counters")
	}
}

func TestSkullbriarRedirectedHandLibraryArrivals(t *testing.T) {
	t.Parallel()
	for _, destination := range []zone.Type{zone.Hand, zone.Library, zone.Exile} {
		t.Run(destination.String(), func(t *testing.T) {
			t.Parallel()
			g, e, source := skullbriarGame(t)
			source.Counters = skullCounters()
			resolveInstruction(e, g, &game.StackObject{Controller: game.Player1}, game.CreateReplacement{
				Replacement: &game.ReplacementEffect{
					MatchEvent: game.EventZoneChanged, MatchToZone: true, ToZone: zone.Graveyard,
					ReplaceToZone: destination,
				},
			}, nil)
			if !movePermanentToZone(g, source, zone.Graveyard) {
				t.Fatal("redirected departure failed")
			}
			actual, _ := destinationZone(g, game.Player1, destination)
			if !actual.Contains(source.CardInstanceID) {
				t.Fatal("replacement destination not reached")
			}
			set := g.CardCounters[source.CardInstanceID]
			if destination == zone.Exile {
				assertSkullCounters(t, set)
				// An off-battlefield move must also evaluate the replaced destination.
				g.ReplacementEffects[0].ReplaceToZone = zone.Hand
				if !moveCardBetweenZones(g, game.Player1, source.CardInstanceID, zone.Exile, zone.Graveyard) {
					t.Fatal("off-battlefield redirected move failed")
				}
				set = g.CardCounters[source.CardInstanceID]
				if !g.Players[game.Player1].Hand.Contains(source.CardInstanceID) || !set.IsEmpty() {
					t.Fatal("off-battlefield hand arrival retained counters")
				}
			} else if !set.IsEmpty() {
				t.Fatal("retention used intended graveyard instead of actual hand/library")
			}
		})
	}
}

func TestSkullbriarCopyingAwayAndFaceDownSource(t *testing.T) {
	t.Parallel()
	g, _, source := skullbriarGame(t)
	source.Counters = skullCounters()
	ordinary := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	becomeStaticCopy(t, g, source, ordinary, false)
	movePermanentToZone(g, source, zone.Graveyard)
	set := g.CardCounters[source.CardInstanceID]
	if !set.IsEmpty() {
		t.Fatal("Skullbriar copied away retained its printed policy")
	}
	card, _ := g.GetCardInstance(source.CardInstanceID)
	g.Players[game.Player1].Graveyard.Remove(card.ID)
	source, _ = createCardPermanent(g, card, game.Player1, zone.Graveyard)
	source.Counters = skullCounters()
	source.FaceDown = true
	movePermanentToZone(g, source, zone.Exile)
	set = g.CardCounters[card.ID]
	if !set.IsEmpty() {
		t.Fatal("face-down battlefield object had its printed retention ability")
	}
}

func TestSkullbriarSourceZoneSuppressionBeforeCastingCostsAndUndo(t *testing.T) {
	t.Parallel()
	g, e, muldrotha := muldrothaGame(t)
	production := lifetimeDefinition(t, cards.NewDefaultRegistry(), skullbriarID)
	witness := *production
	witness.AdditionalCosts = []cost.Additional{{
		Kind: cost.AdditionalSacrifice, Amount: 1, MatchPermanentType: true, PermanentType: types.Creature,
	}}
	cardID := addCardToGraveyard(g, game.Player1, &witness)
	g.CardCounters[cardID] = skullCounters()
	jailer := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Ability suppressor", Types: []types.Card{types.Creature},
		Power: opt.Val(game.PT{Value: 2}), Toughness: opt.Val(game.PT{Value: 1}),
		StaticAbilities: []game.StaticAbility{{ContinuousEffects: []game.ContinuousEffect{{
			Layer: game.LayerAbility, CardZone: zone.Graveyard, RemoveAllAbilities: true,
		}}}},
	}})
	card, _ := g.GetCardInstance(cardID)
	transition := prepareCastCounterTransition(g, card, zone.Graveyard)
	transition.apply(g)
	set := g.CardCounters[cardID]
	if !set.IsEmpty() {
		t.Fatal("casting proposal ignored departing graveyard ability loss")
	}
	transition.undo(g)
	assertSkullCounters(t, g.CardCounters[cardID])
	g.Players[game.Player1].ManaPool.Add(mana.B, 1)
	g.Players[game.Player1].ManaPool.Add(mana.G, 1)
	agent := &muldrothaChoiceAgent{choose: func(request game.ChoiceRequest) []int {
		for i, option := range request.Options {
			if option.Card.Exists && option.Card.Val.CardID == jailer.CardInstanceID {
				return []int{i}
			}
		}
		return request.DefaultSelection
	}}
	if !e.applyActionWithChoices(g, game.Player1, graveyardCast(cardID, muldrotha, types.Creature),
		[game.NumPlayers]PlayerAgent{game.Player1: agent}, &TurnLog{}) {
		t.Fatal("counter-bearing graveyard witness could not pay its actual sacrifice cost")
	}
	set = g.CardCounters[cardID]
	if _, exists := permanentByObjectID(g, jailer.ObjectID); exists || !set.IsEmpty() {
		t.Fatal("post-payment source abilities resurrected counters lost at CR 601.2a")
	}
	e.resolveTopOfStack(g, &TurnLog{})
	if entrant := permanentByCardID(g, cardID); entrant == nil || !entrant.Counters.IsEmpty() {
		t.Fatal("counters lost before costs reappeared at resolution")
	}
}

func TestSkullbriarAbilityLossCopiesAndTokens(t *testing.T) {
	t.Parallel()
	g, e, source := skullbriarGame(t)
	source.Counters = skullCounters()
	def, _ := permanentCardDef(g, source)
	ordinary := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	ordinary.Counters = skullCounters()
	becomeStaticCopy(t, g, ordinary, source, false)
	if !movePermanentsToZoneSimultaneously(g, []*game.Permanent{source, ordinary}, zone.Graveyard) {
		t.Fatal("simultaneous copy departure failed")
	}
	assertSkullCounters(t, g.CardCounters[source.CardInstanceID])
	assertSkullCounters(t, g.CardCounters[ordinary.CardInstanceID])
	if !moveCardBetweenZones(g, ordinary.Owner, ordinary.CardInstanceID, zone.Graveyard, zone.Exile) {
		t.Fatal("copied physical card's next move failed")
	}
	set := g.CardCounters[ordinary.CardInstanceID]
	if !set.IsEmpty() {
		t.Fatal("temporary copy policy persisted into next incarnation")
	}
	jailerDef := &game.CardDef{CardFace: game.CardFace{Name: "Graveyard suppression", StaticAbilities: []game.StaticAbility{{
		ContinuousEffects: []game.ContinuousEffect{{Layer: game.LayerAbility, CardZone: zone.Graveyard, RemoveAllAbilities: true}},
	}}}}
	if issues := game.ValidateCardDef(jailerDef); len(issues) != 0 {
		t.Fatalf("typed graveyard suppression is not valid: %+v", issues)
	}
	jailer := addCombatPermanent(g, game.Player2, jailerDef)
	g.Players[game.Player1].Graveyard.Remove(source.CardInstanceID)
	card, _ := g.GetCardInstance(source.CardInstanceID)
	entrant, ok := createCardPermanent(g, card, game.Player1, zone.Graveyard)
	if !ok || !entrant.Counters.IsEmpty() {
		t.Fatal("Yixlid Jailer-style source-zone ability loss did not remove counters")
	}
	movePermanentToZone(g, jailer, zone.Exile)
	entrant.Counters = skullCounters()
	movePermanentToZone(g, entrant, zone.Graveyard)
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		Layer: game.LayerAbility, Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}}), RemoveAllAbilities: true,
	})
	g.Players[game.Player1].Graveyard.Remove(card.ID)
	entrant, ok = createCardPermanent(g, card, game.Player1, zone.Graveyard)
	if !ok {
		t.Fatal("reentry under Humility-style ability loss failed")
	}
	assertSkullCounters(t, entrant.Counters)
	if !movePermanentToZone(g, entrant, zone.Exile) {
		t.Fatal("abilityless departure failed")
	}
	set = g.CardCounters[card.ID]
	if !set.IsEmpty() {
		t.Fatal("printed policy overrode effective battlefield ability loss")
	}
	g.ContinuousEffects = nil
	createTokenPermanentsWithChoices(e, g, game.Player2, def, 1, false, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	token := g.Battlefield[len(g.Battlefield)-1]
	token.Counters = skullCounters()
	if !movePermanentToZone(g, token, zone.Exile) {
		t.Fatal("token copy departure failed")
	}
	assertSkullCounters(t, g.CardCounters[token.ObjectID])
	e.applyStateBasedActions(g)
	set = g.CardCounters[token.ObjectID]
	if g.Players[game.Player2].Exile.Contains(token.ObjectID) || !set.IsEmpty() {
		t.Fatal("counter retention prevented token from ceasing to exist")
	}
}

func TestSkullbriarOrdinaryCloneAndParallelCounterIsolation(t *testing.T) {
	t.Parallel()
	g, e, source := skullbriarGame(t)
	source.Counters = skullCounters()
	ordinary := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	ordinary.Counters = skullCounters()
	movePermanentsToZoneSimultaneously(g, []*game.Permanent{source, ordinary}, zone.Graveyard)
	card, _ := g.GetCardInstance(source.CardInstanceID)
	ordinaryCard, _ := g.GetCardInstance(ordinary.CardInstanceID)
	set := g.CardCounters[ordinaryCard.ID]
	if !set.IsEmpty() {
		t.Fatal("ordinary creature retained counters")
	}
	subject := selectionSubject{kind: subjectCard, g: g, card: card}
	if !matchSelection(&subject, &game.Selection{
		Power: opt.Val(compare.Int{Op: compare.Equal, Value: 4}), Toughness: opt.Val(compare.Int{Op: compare.Equal, Value: 4}),
		RequiredCounter: counter.Charge, RequiredCounterCount: opt.Val(compare.Int{Op: compare.Equal, Value: 2}),
	}) {
		t.Fatal("off-battlefield counters did not affect actual card selection")
	}
	clone := g.Clone()
	clone.AddExileCounter(card.ID, counter.Charge, 8)
	assertSkullCounters(t, g.CardCounters[card.ID])
	for i := range 8 {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()
			branch := g.Clone()
			branch.Players[game.Player1].Graveyard.Remove(card.ID)
			branchCard, _ := branch.GetCardInstance(card.ID)
			entrant, ok := createCardPermanent(branch, branchCard, game.Player1, zone.Graveyard)
			if !ok {
				t.Fatal("clone reentry failed")
			}
			entrant.Counters.Add(counter.Charge, i)
			e.applyStateBasedActions(branch)
			assertSkullCounters(t, g.CardCounters[card.ID])
		})
	}
}

func TestSkullbriarPolicyRenderer(t *testing.T) {
	def := lifetimeDefinition(t, cards.NewDefaultRegistry(), skullbriarID)
	source, err := (cardgen.Renderer{}).RenderCardSource(
		&cardgen.ScryfallCard{Name: def.Name, Layout: "normal"}, []*game.CardDef{def}, nil, "test")
	if err != nil || !strings.Contains(source, "game.ZoneTransitionRetainCountersExceptHandLibrary") {
		t.Fatalf("policy renderer: %v", err)
	}
}

func TestSkullbriarCopiedAndRemovedOffBattlefieldSuppressionSource(t *testing.T) {
	t.Parallel()
	g, _, source := skullbriarGame(t)
	source.Counters = skullCounters()
	movePermanentToZone(g, source, zone.Graveyard)
	jailer := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name: "Suppression blueprint", StaticAbilities: []game.StaticAbility{{ContinuousEffects: []game.ContinuousEffect{{
			Layer: game.LayerAbility, CardZone: zone.Graveyard, RemoveAllAbilities: true,
		}}}},
	}})
	host := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	becomeStaticCopy(t, g, host, jailer, false)
	movePermanentToZone(g, jailer, zone.Exile)
	card, _ := g.GetCardInstance(source.CardInstanceID)
	if cardRetainsCounters(g, card, zone.Graveyard, zone.Battlefield, game.FaceFront, false) {
		t.Fatal("copied static source was not discovered from effective ability bodies")
	}
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		AffectedObjectID: host.ObjectID, Layer: game.LayerAbility, RemoveAllAbilities: true,
	})
	if !cardRetainsCounters(g, card, zone.Graveyard, zone.Battlefield, game.FaceFront, false) {
		t.Fatal("abilityless copied suppressor still disabled Skullbriar's departing ability")
	}
	assertSkullCounters(t, g.CardCounters[card.ID])
}

func TestSkullbriarFaceDownExileDepartures(t *testing.T) {
	t.Parallel()
	for _, destination := range []zone.Type{zone.Graveyard, zone.Battlefield, zone.Command} {
		t.Run(destination.String(), func(t *testing.T) {
			t.Parallel()
			g, e, source := skullbriarGame(t)
			source.Counters = skullCounters()
			movePermanentToZone(g, source, zone.Exile)
			g.Players[game.Player1].Exile.SetFaceDown(source.CardInstanceID, true)
			card, _ := g.GetCardInstance(source.CardInstanceID)
			if destination == zone.Battlefield {
				r := &effectResolver{engine: e, game: g, obj: &game.StackObject{Controller: game.Player1}, log: &TurnLog{}}
				entrant, ok := r.putResolvedCardOnBattlefieldValue(card, zone.Exile, game.Player1, nil, permanentCreationOptions{})
				if !ok || !entrant.Counters.IsEmpty() {
					t.Fatal("face-down exiled source incorrectly had a retention ability on reentry")
				}
			} else {
				if !moveCardBetweenZones(g, game.Player1, card.ID, zone.Exile, destination) {
					t.Fatal("face-down exile departure failed")
				}
				set := g.CardCounters[card.ID]
				if !set.IsEmpty() {
					t.Fatal("face-down exiled source incorrectly retained counters")
				}
			}
		})
	}
}

func TestSkullbriarAuthorizedCastTurnsFaceDownExileUpBeforeAnnouncement(t *testing.T) {
	t.Parallel()
	g, e, source := skullbriarGame(t)
	source.Counters = skullCounters()
	movePermanentToZone(g, source, zone.Exile)
	cardID := source.CardInstanceID
	g.Players[game.Player1].Exile.SetFaceDown(cardID, true)
	g.RuleEffects = append(g.RuleEffects, game.RuleEffect{
		Kind: game.RuleEffectCastFromZone, Controller: game.Player1, AffectedPlayer: game.PlayerYou,
		AffectedCardID: cardID, CastFromZone: zone.Exile,
	})
	g.Players[game.Player1].ManaPool.Add(mana.B, 1)
	g.Players[game.Player1].ManaPool.Add(mana.G, 1)
	if !e.applyAction(g, game.Player1, action.CastSpellFromZone(cardID, zone.Exile, nil, 0, nil)) {
		t.Fatal("explicit permission did not allow casting the exiled card")
	}
	assertSkullCounters(t, g.CardCounters[cardID])
	e.resolveTopOfStack(g, &TurnLog{})
	assertSkullCounters(t, permanentByCardID(g, cardID).Counters)
}

func TestSkullbriarRetainedMinusCountersAndNoPlacementReplacement(t *testing.T) {
	t.Parallel()
	g, e, source := skullbriarGame(t)
	addReplacementPermanent(t, g, game.Player1, anyCounterDoublingReplacementCardDef())
	source.Counters.Add(counter.MinusOneMinusOne, 1)
	movePermanentToZone(g, source, zone.Graveyard)
	card, _ := g.GetCardInstance(source.CardInstanceID)
	g.Players[game.Player1].Graveyard.Remove(card.ID)
	entrant, ok := createCardPermanent(g, card, game.Player1, zone.Graveyard)
	if !ok || entrant.Counters.Get(counter.MinusOneMinusOne) != 1 {
		t.Fatal("retention was treated as counter placement by the shared doubling replacement")
	}
	e.applyStateBasedActions(g)
	set := g.CardCounters[card.ID]
	if permanentByCardID(g, card.ID) != nil || !g.Players[game.Player1].Graveyard.Contains(card.ID) ||
		set.Get(counter.MinusOneMinusOne) != 1 {
		t.Fatal("retained -1/-1 counter did not cause ordinary zero-toughness death and remain on the new card")
	}
}
