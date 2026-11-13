package rules

import (
	"reflect"
	"testing"

	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review: pinned-corpus Oracle text for Dredge the Mire (C20
// 10e286b0-7245-4c30-a7a9-6d4a9c49b0a7, {3}{B} Sorcery, "Each opponent chooses
// a creature card in their graveyard. Put those cards onto the battlefield
// under your control."), Exhume (fbe61f74-1b3c-4e12-8758-7029872c9ff1, {1}{B}
// Sorcery, "Each player puts a creature card from their graveyard onto the
// battlefield.") and Show and Tell (b83a3ba0-249e-4c39-bbf0-cb005413f7d2,
// "Each player may put an artifact, creature, enchantment, or land card from
// their hand onto the battlefield."). Scryfall /cards/<id>/rulings: Dredge the
// Mire 2020-04-17 (the next opponent in turn order, or the active opponent,
// chooses first, then each other opponent in turn order knowing the earlier
// choices; all cards are returned at the same time); Exhume has none; Show and
// Tell 2008-04-01 (current player first, then turn order; a chosen card need
// not be revealed so long as it is clear which card was chosen; all cards enter
// simultaneously). Comprehensive Rules effective 2026-11-06 (SHA-256
// 524f5195a1e5f54bb13c735edf399ac717a4458e2713f398d1a3e6877186b829): 101.4
// (APNAP choices, then simultaneous actions), 101.4a (choices from a hidden zone
// may stay face down), 101.4b (players know earlier choices), 110.2a (a
// permanent enters under the control of the player who put it there unless the
// effect says otherwise) and 603.6a (enters triggers see every permanent that
// entered in the same event). These fixtures prove the shared
// ChooseFromZone.Choosers operation (#3830); no card is admitted.

// groupZoneChoiceRecorder records each request and picks the named card, or
// nothing when pick is empty. check runs at choice time.
type groupZoneChoiceRecorder struct {
	simPassPolicy

	pick     string
	requests []game.ChoiceRequest
	check    func()
}

func (a *groupZoneChoiceRecorder) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	a.requests = append(a.requests, request)
	if a.check != nil {
		a.check()
	}
	for _, option := range request.Options {
		if a.pick != "" && option.Card.Exists && option.Card.Val.Name == a.pick {
			return []int{option.Index}
		}
	}
	return nil
}

func dredgeTheMireFixture() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     "Dredge the Mire",
		ManaCost: opt.Val(cost.Mana{cost.O(3), cost.B}),
		Types:    []types.Card{types.Sorcery},
		SpellAbility: opt.Val(game.Mode{Sequence: []game.Instruction{{Primitive: game.ChooseFromZone{
			Choosers:    opt.Val(game.OpponentsReference()),
			SourceZone:  zone.Graveyard,
			Filter:      game.Selection{RequiredTypes: []types.Card{types.Creature}},
			Quantity:    game.Fixed(1),
			Count:       game.ChooseExactly,
			Destination: game.ChooseDestination{Zone: zone.Battlefield},
			Riders:      game.ChooseRiders{EntryController: opt.Val(game.ControllerReference())},
		}}}}.Ability()),
	}}
}

// creatureWatcherFixture is a creature whose controller gains 1 life whenever
// another creature enters.
func creatureWatcherFixture(name string) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:      name,
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
		TriggeredAbilities: []game.TriggeredAbility{{
			Trigger: game.TriggerCondition{Type: game.TriggerWhenever, Pattern: game.TriggerPattern{
				Event:            game.EventPermanentEnteredBattlefield,
				ExcludeSelf:      true,
				SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
			}},
			Content: game.Mode{Sequence: []game.Instruction{{
				Primitive: game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(1)},
			}}}.Ability(),
		}},
	}}
}

func groupChoiceEntryBatches(t *testing.T, g *game.Game, cards ...id.ID) []id.ID {
	t.Helper()
	var batches []id.ID
	for _, cardID := range cards {
		permanent, ok := reanimatedPermanent(g, cardID)
		if !ok {
			t.Fatalf("card %v did not enter the battlefield", cardID)
		}
		for _, event := range g.Events {
			if event.Kind == game.EventPermanentEnteredBattlefield && event.PermanentID == permanent.ObjectID {
				batches = append(batches, event.SimultaneousID)
			}
		}
	}
	return batches
}

func TestDredgeTheMireOpponentsChooseInTurnOrderThenCardsEnterTogether(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	casterCreature := addCardToGraveyard(g, game.Player1, chooseFromZoneCardDef("Caster Bear", 2, []types.Card{types.Creature}, nil))
	watcherA := addCardToGraveyard(g, game.Player2, creatureWatcherFixture("Watcher A"))
	p2Other := addCardToGraveyard(g, game.Player2, chooseFromZoneCardDef("Second Bear", 2, []types.Card{types.Creature}, nil))
	watcherB := addCardToGraveyard(g, game.Player3, creatureWatcherFixture("Watcher B"))
	addCardToGraveyard(g, game.Player3, chooseFromZoneCardDef("Shock", 1, []types.Card{types.Instant}, nil))
	addCardToGraveyard(g, game.Player4, chooseFromZoneCardDef("Divination", 3, []types.Card{types.Sorcery}, nil))

	dredge := addCardToHand(g, game.Player1, dredgeTheMireFixture())
	g.Players[game.Player1].ManaPool.Add(mana.B, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 3)
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.CastSpell(dredge, nil, 0, nil)) {
		t.Fatal("Dredge the Mire cast failed")
	}

	p1 := &groupZoneChoiceRecorder{}
	p2 := &groupZoneChoiceRecorder{pick: "Watcher A"}
	p4 := &groupZoneChoiceRecorder{}
	// Player3 chooses after Player2, before Player2's choice has moved.
	p3 := &groupZoneChoiceRecorder{pick: "Watcher B", check: func() {
		if !g.Players[game.Player2].Graveyard.Contains(watcherA) {
			t.Fatal("Player2's chosen card moved before every opponent chose")
		}
	}}
	agents := simPassPolicies()
	agents[game.Player1], agents[game.Player2], agents[game.Player3], agents[game.Player4] = p1, p2, p3, p4
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if len(p1.requests) != 0 || len(p4.requests) != 0 {
		t.Fatalf("caster requests = %d, empty-graveyard opponent requests = %d, want none", len(p1.requests), len(p4.requests))
	}
	if len(p2.requests) != 1 || len(p2.requests[0].Options) != 2 || p2.requests[0].MinChoices != 1 ||
		p2.requests[0].ZoneChoiceDecisions != nil {
		t.Fatalf("Player2 requests = %#v, want one first choice between two creature cards", p2.requests)
	}
	if len(p3.requests) != 1 || len(p3.requests[0].Options) != 1 {
		t.Fatalf("Player3 requests = %#v, want one choice of its only creature card", p3.requests)
	}
	wantEarlier := []game.ZoneChoiceDecision{{Player: game.Player2, Count: 1, Cards: []id.ID{watcherA}}}
	if got := p3.requests[0].ZoneChoiceDecisions; !reflect.DeepEqual(got, wantEarlier) {
		t.Fatalf("Player3 saw earlier choices %#v, want %#v", got, wantEarlier)
	}
	for _, cardID := range []id.ID{watcherA, watcherB} {
		permanent, ok := reanimatedPermanent(g, cardID)
		if !ok || permanent.Controller != game.Player1 {
			t.Fatalf("chosen card %v entered = %v, want under the caster's control", cardID, ok)
		}
	}
	if !g.Players[game.Player2].Graveyard.Contains(p2Other) || !g.Players[game.Player1].Graveyard.Contains(casterCreature) {
		t.Fatal("an unchosen creature card left its graveyard")
	}
	if batches := groupChoiceEntryBatches(t, g, watcherA, watcherB); len(batches) != 2 || batches[0] == 0 || batches[0] != batches[1] {
		t.Fatalf("entry batches = %v, want one simultaneous entry", batches)
	}

	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, simPassPolicies(), &TurnLog{}) {
		t.Fatal("the entering watchers did not see each other")
	}
	if got := g.Stack.Size(); got != 2 {
		t.Fatalf("watcher triggers = %d, want each watcher to see the other enter (2)", got)
	}
	life := g.Players[game.Player1].Life
	for range 2 {
		engine.resolveTopOfStackWithChoices(g, simPassPolicies(), &TurnLog{})
	}
	if got := g.Players[game.Player1].Life; got != life+2 {
		t.Fatalf("caster life = %d, want %d", got, life+2)
	}
}

var exhumeFixtureProgram = game.ProgramReference{ID: "test/exhume/spell/0", Version: "1"}

func TestExhumeProgramEachPlayerReturnsOwnCreatureTogether(t *testing.T) {
	catalog, err := cardprogram.NewCatalog([]cardprogram.Program{{
		Reference: exhumeFixtureProgram,
		Resolve: func(ctx cardprogram.Context) error {
			_, err := ctx.Execute(game.ChooseFromZone{
				Choosers:    opt.Val(game.AllPlayersReference()),
				SourceZone:  zone.Graveyard,
				Filter:      game.Selection{RequiredTypes: []types.Card{types.Creature}},
				Quantity:    game.Fixed(1),
				Count:       game.ChooseExactly,
				Destination: game.ChooseDestination{Zone: zone.Battlefield},
			})
			return err
		},
	}})
	if err != nil {
		t.Fatal(err)
	}
	engine := NewEngineWithPrograms(nil, catalog)
	exhume := &game.CardDef{CardFace: game.CardFace{
		Name:     "Exhume",
		ManaCost: opt.Val(cost.Mana{cost.O(1), cost.B}),
		Types:    []types.Card{types.Sorcery},
		SpellAbility: opt.Val(game.Mode{
			Text:    "Each player puts a creature card from their graveyard onto the battlefield.",
			Program: exhumeFixtureProgram,
		}.Ability()),
	}}

	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	setMainPhasePriority(g, game.Player2)
	returned := map[game.PlayerID]id.ID{
		game.Player1: addCardToGraveyard(g, game.Player1, creatureWatcherFixture("Watcher One")),
		game.Player2: addCardToGraveyard(g, game.Player2, chooseFromZoneCardDef("Caster Bear", 2, []types.Card{types.Creature}, nil)),
		game.Player3: addCardToGraveyard(g, game.Player3, creatureWatcherFixture("Watcher Three")),
	}
	spell := addCardToHand(g, game.Player2, exhume)
	g.Players[game.Player2].ManaPool.Add(mana.B, 1)
	g.Players[game.Player2].ManaPool.Add(mana.C, 1)
	if !engine.applyAction(g, game.Player2, action.CastSpell(spell, nil, 0, nil)) {
		t.Fatal("Exhume cast failed")
	}

	var order []game.PlayerID
	agents := simPassPolicies()
	for _, player := range []game.PlayerID{game.Player1, game.Player2, game.Player3, game.Player4} {
		agents[player] = &authoredCopyChoiceAgent{choose: func(request game.ChoiceRequest) []int {
			order = append(order, request.Player)
			return request.DefaultSelection
		}}
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if want := []game.PlayerID{game.Player2, game.Player3, game.Player1}; !reflect.DeepEqual(order, want) {
		t.Fatalf("choice order = %v, want APNAP from the active player %v", order, want)
	}
	for owner, cardID := range returned {
		permanent, ok := reanimatedPermanent(g, cardID)
		if !ok || permanent.Controller != owner {
			t.Fatalf("player %v's creature entered = %v, want under its owner's control", owner, ok)
		}
	}
	batches := groupChoiceEntryBatches(t, g, returned[game.Player1], returned[game.Player2], returned[game.Player3])
	if len(batches) != 3 || batches[0] == 0 || batches[0] != batches[1] || batches[1] != batches[2] {
		t.Fatalf("entry batches = %v, want one simultaneous entry", batches)
	}
	engine.putTriggeredAbilitiesOnStackWithChoices(g, simPassPolicies(), &TurnLog{})
	if got := g.Stack.Size(); got != 4 {
		t.Fatalf("watcher triggers = %d, want each watcher to see both other creatures (4)", got)
	}
}

func TestShowAndTellHiddenHandChoicesShowOnlyCounts(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	obj := chooseFromZoneSource(g)
	setMainPhasePriority(g, game.Player1)
	permanentTypes := []types.Card{types.Artifact, types.Creature, types.Enchantment, types.Land}
	p1Artifact := addCardToHand(g, game.Player1, chooseFromZoneCardDef("Hidden Relic", 3, []types.Card{types.Artifact}, nil))
	p2Creature := addCardToHand(g, game.Player2, chooseFromZoneCardDef("Hidden Giant", 7, []types.Card{types.Creature}, nil))
	p3Land := addCardToHand(g, game.Player3, chooseFromZoneCardDef("Hidden Grove", 0, []types.Card{types.Land}, nil))
	addCardToHand(g, game.Player4, chooseFromZoneCardDef("Hidden Spell", 1, []types.Card{types.Instant}, nil))

	p1 := &groupZoneChoiceRecorder{pick: "Hidden Relic"}
	p2 := &groupZoneChoiceRecorder{check: func() {
		if !g.Players[game.Player1].Hand.Contains(p1Artifact) {
			t.Fatal("Player1's chosen card moved before every player chose")
		}
	}}
	p3 := &groupZoneChoiceRecorder{pick: "Hidden Grove"}
	p4 := &groupZoneChoiceRecorder{}
	agents := [game.NumPlayers]PlayerAgent{game.Player1: p1, game.Player2: p2, game.Player3: p3, game.Player4: p4}
	resolver := newEffectResolver(NewEngine(nil), g, obj, agents, &TurnLog{})
	resolver.resolveChooseFromZone(game.ChooseFromZone{
		Choosers:    opt.Val(game.AllPlayersReference()),
		SourceZone:  zone.Hand,
		Filter:      game.Selection{RequiredTypesAny: permanentTypes},
		Quantity:    game.Fixed(1),
		Count:       game.ChooseUpTo,
		Destination: game.ChooseDestination{Zone: zone.Battlefield},
		Riders:      game.ChooseRiders{PublishLinked: "shown"},
	})

	if len(p1.requests) != 1 || !p1.requests[0].Private {
		t.Fatalf("Player1 requests = %#v, want one choice kept out of the public log", p1.requests)
	}
	if len(p2.requests) != 1 || p2.requests[0].MinChoices != 0 || !p2.requests[0].Private {
		t.Fatalf("Player2 requests = %#v, want one private optional choice", p2.requests)
	}
	wantP2 := []game.ZoneChoiceDecision{{Player: game.Player1, Count: 1}}
	if got := p2.requests[0].ZoneChoiceDecisions; !reflect.DeepEqual(got, wantP2) {
		t.Fatalf("Player2 saw %#v, want only Player1's face-down count %#v", got, wantP2)
	}
	wantP3 := []game.ZoneChoiceDecision{{Player: game.Player1, Count: 1}, {Player: game.Player2}}
	if len(p3.requests) != 1 || !reflect.DeepEqual(p3.requests[0].ZoneChoiceDecisions, wantP3) {
		t.Fatalf("Player3 requests = %#v, want earlier face-down counts %#v", p3.requests, wantP3)
	}
	if len(p4.requests) != 0 {
		t.Fatal("a player with no matching hand card was asked to choose")
	}
	if permanent, ok := reanimatedPermanent(g, p1Artifact); !ok || permanent.Controller != game.Player1 {
		t.Fatal("Player1's chosen artifact did not enter under its chooser's control")
	}
	if !g.Players[game.Player2].Hand.Contains(p2Creature) {
		t.Fatal("Player2 declined but its creature left its hand")
	}
	if batches := groupChoiceEntryBatches(t, g, p1Artifact, p3Land); len(batches) != 2 || batches[0] == 0 || batches[0] != batches[1] {
		t.Fatalf("entry batches = %v, want one simultaneous entry", batches)
	}
	refs := linkedObjects(g, linkedObjectSourceKey(g, obj, "shown"))
	if len(refs) != 2 || refs[0].CardID != p1Artifact || refs[1].CardID != p3Land {
		t.Fatalf("linked objects = %v, want the whole chosen set", refs)
	}
}

// TestGroupHiddenZoneChoicesStayOutOfThePublicLog verifies that no part of a
// Choosers member's hidden-zone decision reaches the public log, for every
// grouping a group choice can use: whether the member was asked, how many
// options they had and what they chose would reveal how many matching cards
// they hold (CR 400.2, 101.4a). Only the moved cards become public.
func TestGroupHiddenZoneChoicesStayOutOfThePublicLog(t *testing.T) {
	groupings := map[string]game.ChooseFromZone{
		"across the set": {
			Filter: game.Selection{RequiredTypes: []types.Card{types.Creature}},
		},
		"one of each named type": {
			Filter:   game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Artifact}},
			Grouping: game.ChooseOneOfEachNamedType,
		},
	}
	for name, env := range groupings {
		t.Run(name, func(t *testing.T) {
			for eligible := range 3 {
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				obj := chooseFromZoneSource(g)
				for range eligible {
					addCardToHand(g, game.Player1, chooseFromZoneCardDef("Held Creature", 2, []types.Card{types.Creature}, nil))
				}
				shown := addCardToHand(g, game.Player2, chooseFromZoneCardDef("Shown Creature", 2, []types.Card{types.Creature}, nil))
				p1 := &groupZoneChoiceRecorder{}
				agents := simPassPolicies()
				agents[game.Player1], agents[game.Player2] = p1, &groupZoneChoiceRecorder{pick: "Shown Creature"}
				log := &TurnLog{}
				env.Choosers = opt.Val(game.AllPlayersReference())
				env.SourceZone = zone.Hand
				env.Quantity = game.Fixed(1)
				env.Count = game.ChooseUpTo
				env.Destination = game.ChooseDestination{Zone: zone.Battlefield}
				newEffectResolver(NewEngine(nil), g, obj, agents, log).resolveChooseFromZone(env)

				if eligible > 0 && len(p1.requests) == 0 {
					t.Fatalf("eligible=%d: Player1 was not asked to choose", eligible)
				}
				if len(log.Choices) != 0 {
					t.Fatalf("eligible=%d: public log recorded hidden-zone decisions %#v", eligible, log.Choices)
				}
				if _, ok := reanimatedPermanent(g, shown); !ok {
					t.Fatalf("eligible=%d: Player2's chosen card did not enter", eligible)
				}
			}
		})
	}
}

// TestGroupChoicesMoveToAnotherZoneTogether verifies that a group choice whose
// destination is not the battlefield still moves every chosen card in one
// simultaneous event (CR 101.4, 603.2c) and publishes the whole moved set.
func TestGroupChoicesMoveToAnotherZoneTogether(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	obj := chooseFromZoneSource(g)
	first := addCardToGraveyard(g, game.Player1, chooseFromZoneCardDef("First Bear", 2, []types.Card{types.Creature}, nil))
	second := addCardToGraveyard(g, game.Player3, chooseFromZoneCardDef("Third Bear", 2, []types.Card{types.Creature}, nil))
	agents := simPassPolicies()
	agents[game.Player3] = &groupZoneChoiceRecorder{pick: "Third Bear", check: func() {
		if !g.Players[game.Player1].Graveyard.Contains(first) {
			t.Fatal("Player1's chosen card moved before every player chose")
		}
	}}
	newEffectResolver(NewEngine(nil), g, obj, agents, &TurnLog{}).resolveChooseFromZone(game.ChooseFromZone{
		Choosers:    opt.Val(game.AllPlayersReference()),
		SourceZone:  zone.Graveyard,
		Filter:      game.Selection{RequiredTypes: []types.Card{types.Creature}},
		Quantity:    game.Fixed(1),
		Count:       game.ChooseExactly,
		Destination: game.ChooseDestination{Zone: zone.Hand},
		Riders:      game.ChooseRiders{PublishLinked: "returned"},
	})

	if !g.Players[game.Player1].Hand.Contains(first) || !g.Players[game.Player3].Hand.Contains(second) {
		t.Fatal("each chosen card did not return to its owner's hand")
	}
	var batches []id.ID
	for _, event := range g.Events {
		if event.Kind == game.EventZoneChanged && event.ToZone == zone.Hand && (event.CardID == first || event.CardID == second) {
			batches = append(batches, event.SimultaneousID)
		}
	}
	if len(batches) != 2 || batches[0] == 0 || batches[0] != batches[1] {
		t.Fatalf("hand arrival batches = %v, want one simultaneous move", batches)
	}
	refs := linkedObjects(g, linkedObjectSourceKey(g, obj, "returned"))
	if len(refs) != 2 || refs[0].CardID != first || refs[1].CardID != second {
		t.Fatalf("linked objects = %v, want the whole moved set", refs)
	}
}
