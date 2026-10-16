package rules

import (
	"reflect"
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/cards/authored"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func warpWorldLibraryFixture(t *testing.T, libraries [game.NumPlayers][]*game.CardDef) (*game.Game, *Engine, map[string]id.ID) {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer, g.Turn.TurnNumber = game.Player3, 7
	cards := make(map[string]id.ID)
	for player, definitions := range libraries {
		for _, def := range definitions {
			cards[def.Name] = addCardToLibrary(g, game.PlayerID(player), def)
			g.Battlefield = append(g.Battlefield, &game.Permanent{
				ObjectID: g.IDGen.Next(), Owner: game.PlayerID(player), Controller: game.PlayerID(player), Token: true,
				TokenDef: vanillaCreature("Counting token", 1, 1),
			})
		}
	}
	catalog, err := cardprogram.NewCatalog(authored.WarpWorld.Programs)
	if err != nil {
		t.Fatal(err)
	}
	addImplementationSpellToStack(g, game.Player4, authored.WarpWorld.New(), nil)
	return g, NewEngineWithPrograms(nil, catalog), cards
}

func warpWorldAura(name string, enchanted types.Card) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name: name, Types: []types.Card{types.Enchantment}, Subtypes: []types.Sub{types.Aura},
		StaticAbilities: []game.StaticAbility{game.EnchantStaticAbility(&game.TargetSpec{
			Allow:     game.TargetAllowPermanent,
			Selection: opt.Val(game.Selection{RequiredTypesAny: []types.Card{enchanted}}),
		})},
	}}
}

func TestProductionWarpWorldPrintedPaymentAndClonedResolution(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	setMainPhasePriority(g, game.Player1)
	card := addCardToHand(g, game.Player1, authored.WarpWorld.New())
	addCombatPermanent(g, game.Player1, vanillaCreature("Warped creature", 2, 3))
	engine := removalTokenEngine(t, authored.WarpWorld)
	policies := simPassPolicies()
	proposal := action.CastSpell(card, nil, 0, nil)
	g.Players[game.Player1].ManaPool.Add(mana.R, 7)
	if _, accepted := engine.Simulator().Apply(g, game.Player1, proposal, policies); accepted {
		t.Fatal("Warp World was cast without its full printed eight-mana cost")
	}
	g.Players[game.Player1].ManaPool.Add(mana.R, 1)
	branch, accepted := engine.Simulator().Apply(g, game.Player1, proposal, policies)
	if !accepted || !g.Players[game.Player1].Hand.Contains(card) || branch.Players[game.Player1].Hand.Contains(card) {
		t.Fatal("printed Warp World cast was rejected or mutated its source")
	}
	resolved := engine.Simulator().ResolvePriority(branch, policies)
	if !resolved.Players[game.Player1].Graveyard.Contains(card) || len(resolved.Battlefield) != 1 || len(g.Battlefield) != 1 {
		t.Fatal("cloned production cast did not resolve the complete body")
	}
}

func TestWarpWorldAuraAPNAPAndOwnerChosenRemainder(t *testing.T) {
	firstHost, secondHost := vanillaCreature("First host", 2, 3), vanillaCreature("Second host", 2, 3)
	firstHost.StaticAbilities, secondHost.StaticAbilities = []game.StaticAbility{game.ShroudStaticBody}, []game.StaticAbility{game.ShroudStaticBody}
	g, engine, cards := warpWorldLibraryFixture(t, [game.NumPlayers][]*game.CardDef{
		game.Player1: {firstHost, warpWorldAura("First Aura", types.Creature)},
		game.Player3: {secondHost, warpWorldAura("Third Aura", types.Creature)},
		game.Player2: {
			{CardFace: game.CardFace{Name: "Same-stage enchantment", Types: []types.Card{types.Enchantment}}},
			warpWorldAura("Unplaceable Aura", types.Enchantment),
			{CardFace: game.CardFace{Name: "Remainder instant", Types: []types.Card{types.Instant}}},
			{CardFace: game.CardFace{Name: "Remainder walker", Types: []types.Card{types.Planeswalker}}},
		},
	})
	var auraOrder []game.PlayerID
	var bottomOrder []id.ID
	var policies [game.NumPlayers]PlayerAgent
	for player := range policies {
		policies[player] = &controlScript{choice: func(_ PlayerObservation, request game.ChoiceRequest) []int {
			if g.Stack.Size() != 0 {
				t.Fatal("a triggered ability was placed on the stack during resolution")
			}
			if request.Prompt == "Choose what the Aura enchants." {
				auraOrder = append(auraOrder, request.Player)
				if len(g.Battlefield) != 2 || len(request.Options) != 2 {
					t.Fatal("Aura choices did not observe exactly the pre-stage-one creature hosts")
				}
				return []int{len(request.Options) - 1}
			}
			if request.Kind == game.ChoiceOrder {
				if request.Player != game.Player2 {
					t.Fatal("a different player chose the revealed remainder order")
				}
				var selected []int
				for index := len(request.Options) - 1; index >= 0; index-- {
					selected = append(selected, index)
					bottomOrder = append(bottomOrder, request.Options[index].Card.Val.CardID)
				}
				return selected
			}
			return request.DefaultSelection
		}}
	}
	engine.resolveTopOfStackWithChoices(g, policies, &TurnLog{})
	engine.applyStateBasedActions(g)
	if !slices.Equal(auraOrder, []game.PlayerID{game.Player3, game.Player1}) {
		t.Fatalf("Aura decisions = %v, want active player3 then player1", auraOrder)
	}
	for _, name := range []string{"First Aura", "Third Aura"} {
		permanent := permanentForCard(g, cards[name])
		if permanent == nil || !permanent.AttachedTo.Exists {
			t.Fatalf("%s did not attach to a previous-stage shrouded creature", name)
		}
	}
	if permanentForCard(g, cards["Unplaceable Aura"]) != nil || len(bottomOrder) != 3 ||
		!slices.Equal(g.Players[game.Player2].Library.All(), bottomOrder) {
		t.Fatal("same-stage Aura or owner-ordered remainder was lost, duplicated or illegally placed")
	}
}

func TestWarpWorldNewReplacementSourceOnlyAffectsLaterStage(t *testing.T) {
	g, engine, cards := warpWorldLibraryFixture(t, [game.NumPlayers][]*game.CardDef{
		game.Player1: {
			{CardFace: game.CardFace{
				Name: "New entry replacement", Types: []types.Card{types.Artifact},
				ReplacementAbilities: []game.ReplacementAbility{game.EntersTappedGroupReplacement("Other permanents enter tapped.", game.TriggerControllerAny)},
			}},
			vanillaCreature("Same-stage creature", 2, 3),
		},
		game.Player2: {{CardFace: game.CardFace{Name: "Later-stage enchantment", Types: []types.Card{types.Enchantment}}}},
	})
	engine.resolveTopOfStack(g, &TurnLog{})
	for _, name := range []string{"New entry replacement", "Same-stage creature"} {
		permanent := permanentForCard(g, cards[name])
		if permanent == nil || permanent.Tapped {
			t.Fatalf("%s was affected by a newly entering replacement source", name)
		}
	}
	later := permanentForCard(g, cards["Later-stage enchantment"])
	if later == nil || !later.Tapped {
		t.Fatal("stage-one replacement source did not affect a later-stage entry")
	}
}

func TestWarpWorldOptionalEntryPaymentUsesOwnerHand(t *testing.T) {
	for _, payable := range []bool{false, true} {
		g, engine, cards := warpWorldLibraryFixture(t, [game.NumPlayers][]*game.CardDef{
			game.Player1: {optionalDiscardElseGraveyardArtifact(), vanillaCreature("Other entrant", 2, 3)},
		})
		var land id.ID
		if payable {
			land = addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Payment land", Types: []types.Card{types.Land}}})
		}
		engine.resolveTopOfStack(g, &TurnLog{})
		mox := permanentForCard(g, cards["Mox Diamond"])
		if payable {
			if mox == nil || !g.Players[game.Player1].Graveyard.Contains(land) {
				t.Fatal("paid entry did not use the entering card's owner's hand")
			}
		} else if mox != nil || !g.Players[game.Player1].Graveyard.Contains(cards["Mox Diamond"]) {
			t.Fatal("unpayable entry replacement did not put the actual card in its graveyard")
		}
	}
}

func TestSimultaneousAuraDoesNotPrematurelyModifyHost(t *testing.T) {
	for _, reverse := range []bool{false, true} {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		host := addCombatPermanent(g, game.Player1, vanillaCreature("Unmodified host", 2, 3))
		counterCard := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
			Name: "Count modified on entry", Types: []types.Card{types.Enchantment},
			ReplacementAbilities: []game.ReplacementAbility{game.EntersWithCountersReplacement(
				"Enters with a counter for each modified creature.", game.CounterPlacement{
					Kind: counter.PlusOnePlusOne, Dynamic: opt.Val(&game.DynamicAmount{
						Kind: game.DynamicAmountCountSelector, Multiplier: 1,
						Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, MatchModified: true}),
					}),
				})},
		}})
		addCardToLibrary(g, game.Player1, warpWorldAura("Pending Aura", types.Creature))
		context := &programContext{
			engine: NewEngine(nil), game: g, obj: &game.StackObject{Controller: game.Player1},
			active: true, log: &TurnLog{},
		}
		revealed, err := context.Execute(game.Reveal{Player: game.ControllerReference(), Amount: game.Fixed(2)})
		if err != nil {
			t.Fatal(err)
		}
		if reverse {
			slices.Reverse(revealed.Cards)
		}
		arrivals, err := context.MoveCards(cardprogram.CardPlacement{Cards: revealed.Cards, Destination: zone.Battlefield})
		entered := permanentForCard(g, counterCard)
		if err != nil || len(arrivals) != 2 || entered == nil || entered.Counters.Get(counter.PlusOnePlusOne) != 0 ||
			len(host.Attachments) != 1 {
			t.Fatal("entry preparation observed a pending Aura or failed to publish its final host attachment")
		}
	}
}

func TestExplicitRevealPublishesNativeHandle(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	card := addCardToLibrary(g, game.Player1, vanillaCreature("Observed card", 2, 3))
	context := &programContext{
		engine: NewEngine(nil), game: g, obj: &game.StackObject{
			ID: g.IDGen.Next(), SourceID: card, SourceCardID: card, Controller: game.Player1,
		},
		active: true, log: &TurnLog{},
	}
	if _, err := context.Execute(game.LookAtLibraryTop{Player: game.ControllerReference(), PublishLinked: "observed"}); err != nil {
		t.Fatal(err)
	}
	revealed, err := context.Execute(game.Reveal{Card: game.CardReference{Kind: game.CardReferenceLinked, LinkID: "observed"}})
	if err != nil || !revealed.Succeeded || len(revealed.Objects) != 1 || len(revealed.Cards) != 1 ||
		revealed.Cards[0].Identity.CardID != card {
		t.Fatalf("explicit-card reveal did not publish an exact handle: %+v, %v", revealed, err)
	}
}

func TestWarpWorldCountsCommanderDivertedFromLibrary(t *testing.T) {
	g, engine, _ := warpWorldLibraryFixture(t, [game.NumPlayers][]*game.CardDef{
		game.Player1: {vanillaCreature("First reveal", 2, 3), vanillaCreature("Second reveal", 2, 3)},
	})
	addCardToLibrary(g, game.Player1, vanillaCreature("Third reveal", 2, 3))
	commander := addCommanderPermanent(g, game.Player1)
	commander.Controller = game.Player2
	owner := &choiceOnlyAgent{choices: [][]int{{1}}}
	policies := simPassPolicies()
	policies[game.Player1] = owner
	engine.resolveTopOfStackWithChoices(g, policies, &TurnLog{})
	count := 0
	for _, event := range g.Events {
		if event.Kind == game.EventCardRevealed && event.Player == game.Player1 {
			count++
		}
	}
	if count != 3 || !g.Players[game.Player1].CommandZone.Contains(commander.CardInstanceID) || owner.next != 1 {
		t.Fatal("diverted commander did not count under its pre-move owner")
	}
}

func TestRevealedPlacementRejectsStaleAndForeignHandles(t *testing.T) {
	for _, scenario := range []string{"foreign invocation", "zone incarnation", "shuffle correlation", "diverted entry"} {
		t.Run(scenario, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			card := addCardToLibrary(g, game.Player1, vanillaCreature("Exact revealed card", 2, 3))
			context := &programContext{
				engine: NewEngine(nil), game: g, obj: &game.StackObject{Controller: game.Player1},
				active: true, log: &TurnLog{},
			}
			revealed, err := context.Execute(game.Reveal{Player: game.ControllerReference(), Amount: game.Fixed(1)})
			if err != nil || len(revealed.Cards) != 1 {
				t.Fatal("fixture did not mint a revealed handle")
			}
			switch scenario {
			case "foreign invocation":
				context.found = nil
			case "zone incarnation":
				moveCardBetweenZones(g, game.Player1, card, zone.Library, zone.Hand)
				moveCardBetweenZones(g, game.Player1, card, zone.Hand, zone.Library)
			case "shuffle correlation":
				g.Players[game.Player1].Library.Shuffle(g.RNG)
			case "diverted entry":
				g.ReplacementEffects = append(g.ReplacementEffects, game.ReplacementEffect{
					MatchEvent: game.EventZoneChanged, MatchFromZone: true, FromZone: zone.Library,
					MatchToZone: true, ToZone: zone.Battlefield, ReplaceToZone: zone.Exile,
				})
			default:
				t.Fatal("unknown scenario")
			}
			arrived, err := context.MoveCards(cardprogram.CardPlacement{Cards: revealed.Cards, Destination: zone.Battlefield})
			if scenario == "foreign invocation" {
				if err == nil {
					t.Fatal("another invocation's handle was accepted")
				}
				return
			}
			if err != nil || len(arrived) != 0 || permanentForCard(g, card) != nil {
				t.Fatal("stale or replaced handle reported a battlefield arrival")
			}
			if scenario == "diverted entry" && !g.Players[game.Player1].Exile.Contains(card) {
				t.Fatal("entry redirection lost the actual card")
			}
		})
	}
}

func TestWarpWorldTriggersWaitAndObserveOnlyTheirEntryStages(t *testing.T) {
	trigger := game.TriggeredAbility{
		Trigger: game.TriggerCondition{Type: game.TriggerWhenever, Pattern: game.TriggerPattern{
			Event: game.EventPermanentEnteredBattlefield, Controller: game.TriggerControllerAny,
		}},
		Content: game.Mode{Sequence: []game.Instruction{{
			Primitive: game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(1)},
		}}}.Ability(),
	}
	first := vanillaCreature("First-stage witness", 2, 3)
	first.TriggeredAbilities = []game.TriggeredAbility{trigger}
	second := &game.CardDef{CardFace: game.CardFace{
		Name: "Second-stage witness", Types: []types.Card{types.Enchantment}, TriggeredAbilities: []game.TriggeredAbility{trigger},
	}}
	g, engine, cards := warpWorldLibraryFixture(t, [game.NumPlayers][]*game.CardDef{
		game.Player1: {first, vanillaCreature("Other first-stage entrant", 2, 3), second,
			{CardFace: game.CardFace{Name: "Other second-stage entrant", Types: []types.Card{types.Enchantment}}}},
	})
	engine.resolveTopOfStack(g, &TurnLog{})
	if g.Stack.Size() != 0 || g.Players[game.Player1].Life != 40 {
		t.Fatal("entry triggers were put on the stack or resolved inside Warp World")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) || g.Stack.Size() != 6 {
		t.Fatalf("queued %d triggers, want first-stage witness4 and second-stage witness2", g.Stack.Size())
	}
	var firstCount, secondCount int
	for _, object := range g.Stack.Objects() {
		if object.SourceCardID == cards["First-stage witness"] {
			firstCount++
		}
		if object.SourceCardID == cards["Second-stage witness"] {
			secondCount++
		}
	}
	if firstCount != 4 || secondCount != 2 {
		t.Fatal("a later-stage witness saw an earlier stage, or a simultaneous cohort was serialized")
	}
}

func TestWarpWorldLeaveTriggersKeepPreMoveSourceIdentity(t *testing.T) {
	witness := func(name string) *game.CardDef {
		return &game.CardDef{CardFace: game.CardFace{
			Name: name, Types: []types.Card{types.Enchantment},
			TriggeredAbilities: []game.TriggeredAbility{{
				Trigger: game.TriggerCondition{Type: game.TriggerWhenever, Pattern: game.TriggerPattern{
					Event: game.EventZoneChanged, Controller: game.TriggerControllerAny,
					MatchFromZone: true, FromZone: zone.Battlefield, MatchToZone: true, ToZone: zone.Library,
				}},
				Content: game.Mode{Sequence: []game.Instruction{{
					Primitive: game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(1)},
				}}}.Ability(),
			}},
		}}
	}
	g, engine, _ := warpWorldLibraryFixture(t, [game.NumPlayers][]*game.CardDef{
		game.Player1: {vanillaCreature("Revealed creature", 2, 3), witness("Later leave witness")},
	})
	earlier := addCombatPermanent(g, game.Player1, witness("Earlier leave witness"))
	oldID := earlier.ObjectID
	engine.resolveTopOfStack(g, &TurnLog{})
	if !engine.putTriggeredAbilitiesOnStack(g) || g.Stack.Size() != 3 {
		t.Fatalf("queued %d departure triggers, want only three from the original source", g.Stack.Size())
	}
	for _, object := range g.Stack.Objects() {
		if object.SourceID != oldID {
			t.Fatal("a later-stage or reentered witness observed an earlier battlefield departure")
		}
	}
}

func TestWarpWorldSimulatorCloneIndependence(t *testing.T) {
	g, engine, _ := warpWorldLibraryFixture(t, [game.NumPlayers][]*game.CardDef{
		game.Player1: {vanillaCreature("Clone creature", 2, 3)},
		game.Player2: {{CardFace: game.CardFace{Name: "Clone enchantment", Types: []types.Card{types.Enchantment}}}},
	})
	before := g.Clone()
	simulated := engine.Simulator().ResolvePriority(g, simPassPolicies())
	if !reflect.DeepEqual(g.Battlefield, before.Battlefield) || !reflect.DeepEqual(g.Events, before.Events) || g.Stack.Size() != 1 {
		t.Fatal("simulated Warp World changed its source")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if len(g.Battlefield) != len(simulated.Battlefield) {
		t.Fatal("live and Simulator staged entries differ")
	}
	for _, permanent := range g.Battlefield {
		counterpart := permanentForCard(simulated, permanent.CardInstanceID)
		if counterpart == nil || counterpart.Owner != permanent.Owner || counterpart.Controller != permanent.Controller {
			t.Fatal("live and Simulator lost an exact owned entry")
		}
	}
}

func TestCapturedPlayerReferenceRetainsActorIdentity(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.TurnNumber = 7
	g.PlayerControlEffects = []game.PlayerControlEffect{{
		ID: g.IDGen.Next(), Player: game.Player2, Controller: game.Player3, Turn: 7,
	}}
	context := &programContext{
		engine: NewEngine(nil), game: g, obj: &game.StackObject{Controller: game.Player4},
		active: true, log: &TurnLog{},
	}
	for player := range game.PlayerID(game.NumPlayers) {
		ref := game.CapturedPlayerReference(player)
		if problems := ref.Validate(); len(problems) != 0 {
			t.Fatal(problems)
		}
		resolved, available, err := context.Player(ref)
		if err != nil || !available || resolved != player {
			t.Fatal("captured seat changed to controller or decision maker")
		}
	}
	for _, invalid := range []game.PlayerID{-1, game.NumPlayers} {
		if len(game.CapturedPlayerReference(invalid).Validate()) == 0 {
			t.Fatal("invalid captured seat was accepted")
		}
	}
	g.Players[game.Player2].Eliminated = true
	if _, available, err := context.Player(game.CapturedPlayerReference(game.Player2)); err != nil || available {
		t.Fatal("captured departed seat remained available")
	}
}

func TestRevealedHandlesSurviveEarlierEntryStage(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Creature", Types: []types.Card{types.Creature},
		Power: opt.Val(game.PT{Value: 2}), Toughness: opt.Val(game.PT{Value: 2}),
	}})
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Enchantment", Types: []types.Card{types.Enchantment},
	}})
	context := &programContext{
		engine: NewEngine(nil), game: g, obj: &game.StackObject{Controller: game.Player1},
		active: true, log: &TurnLog{},
	}
	revealed, err := context.Execute(game.Reveal{Player: game.ControllerReference(), Amount: game.Fixed(2)})
	if err != nil || len(revealed.Cards) != 2 {
		t.Fatalf("reveal did not publish both observed incarnations: %+v, %v", revealed, err)
	}
	first, err := context.MoveCards(cardprogram.CardPlacement{Cards: revealed.Cards[1:], Destination: zone.Battlefield})
	if err != nil || len(first) != 1 {
		t.Fatalf("first stage did not enter its card: %+v, %v", first, err)
	}
	second, err := context.MoveCards(cardprogram.CardPlacement{Cards: revealed.Cards[:1], Destination: zone.Battlefield})
	if err != nil || len(second) != 1 {
		t.Fatalf("second stage lost its still-revealed handle: %+v, %v", second, err)
	}
}

func TestWarpWorldOwnedCountsAndCrossPlayerStages(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	first := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Stage-one creature", Types: []types.Card{types.Creature},
		Power: opt.Val(game.PT{Value: 2}), Toughness: opt.Val(game.PT{Value: 2}),
	}})
	multitype := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name: "Stage-one artifact enchantment", Types: []types.Card{types.Artifact, types.Enchantment},
	}})
	multitype.Controller = game.Player1
	for _, player := range []game.PlayerID{game.Player1, game.Player2} {
		g.Battlefield = append(g.Battlefield, &game.Permanent{
			ObjectID: g.IDGen.Next(), Owner: player, Controller: game.Player2, Token: true,
			TokenDef: &game.CardDef{CardFace: game.CardFace{Name: "Owned token", Types: []types.Card{types.Creature}}},
		})
		addCardToLibrary(g, player, &game.CardDef{CardFace: game.CardFace{
			Name: "Stage-two enchantment", Types: []types.Card{types.Enchantment},
		}})
	}
	catalog, err := cardprogram.NewCatalog(authored.WarpWorld.Programs)
	if err != nil {
		t.Fatal(err)
	}
	engine := NewEngineWithPrograms(nil, catalog)
	addImplementationSpellToStack(g, game.Player1, authored.WarpWorld.New(), nil)
	engine.resolveTopOfStack(g, &TurnLog{})
	var counts [game.NumPlayers]int
	var firstBatch, secondBatch id.ID
	entries := 0
	for _, event := range g.Events {
		if event.Kind == game.EventCardRevealed {
			counts[event.Player]++
		}
		if event.Kind != game.EventPermanentEnteredBattlefield {
			continue
		}
		entries++
		card, _ := g.GetCardInstance(event.CardID)
		if card.ID == first.CardInstanceID || card.ID == multitype.CardInstanceID {
			if firstBatch != 0 && firstBatch != event.SimultaneousID {
				t.Fatal("stage one was serialized by player")
			}
			firstBatch = event.SimultaneousID
		} else {
			if secondBatch != 0 && secondBatch != event.SimultaneousID {
				t.Fatal("stage two was serialized by player")
			}
			secondBatch = event.SimultaneousID
		}
	}
	if counts[game.Player1] != 2 || counts[game.Player2] != 2 || entries != 4 ||
		firstBatch == 0 || secondBatch == 0 || firstBatch == secondBatch {
		for _, event := range g.Events {
			t.Logf("event kind %v card %d source %d from %v to %v batch %d", event.Kind, event.CardID, event.SourceID, event.FromZone, event.ToZone, event.SimultaneousID)
		}
		for cardID, card := range g.CardInstances {
			t.Logf("card %d %s types %v version %d owner %d", cardID, card.Def.Name, card.Def.Types, card.ZoneVersion, card.Owner)
		}
		t.Fatalf("wrong ownership/count/stages: reveal %v, entries %d, batches %d/%d", counts, entries, firstBatch, secondBatch)
	}
}
