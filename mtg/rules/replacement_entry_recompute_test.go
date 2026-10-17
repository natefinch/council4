package rules

import (
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func TestEntryCopyRecomputesTriskelionCounters(t *testing.T) {
	registry := cards.NewDefaultRegistry()
	for _, doubling := range []bool{false, true} {
		t.Run(map[bool]string{false: "ordinary", true: "Doubling Season"}[doubling], func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			e := NewEngine(nil)
			addReplacementPermanent(t, g, game.Player1, registry.Lookup("Triskelion"))
			if doubling {
				addReplacementPermanent(t, g, game.Player1, registry.Lookup("Doubling Season"))
			}
			for range 4 {
				addBasicLandPermanent(g, game.Player1, types.Island)
			}
			cloneID := addCardToHand(g, game.Player1, registry.Lookup("Clone"))
			g.Turn.Phase, g.Turn.Step = game.PhasePrecombatMain, game.StepNone
			if !e.applyAction(g, game.Player1, action.CastSpell(cloneID, nil, 0, nil)) {
				t.Fatal("paid Clone cast failed")
			}
			before := g.Clone()
			policies := func() [game.NumPlayers]PlayerAgent {
				agents := simPassPolicies()
				agents[game.Player1] = &choiceOnlyAgent{choices: [][]int{{1}}}
				return agents
			}
			branch := e.Simulator().ResolvePriority(g, policies())
			if permanentByCardID(g, cloneID) != nil || g.Stack.Size() != 1 {
				t.Fatal("Simulator modified the original pending entry")
			}
			e.resolveTopOfStackWithChoices(g, policies(), &TurnLog{})
			replay := before.Clone()
			e.resolveTopOfStackWithChoices(replay, policies(), &TurnLog{})
			want := 3
			if doubling {
				want = 6
			}
			for _, result := range []*game.Game{g, branch, replay} {
				clone := permanentByCardID(result, cloneID)
				if clone == nil || permanentEffectiveName(result, clone) != "Triskelion" ||
					clone.Counters.Get(counter.PlusOnePlusOne) != want {
					t.Fatalf("Clone = %+v, want Triskelion with %d counters", clone, want)
				}
				total, arrivals := 0, 0
				for _, event := range result.Events {
					if event.PermanentID != clone.ObjectID {
						continue
					}
					if event.Kind == game.EventCountersAdded {
						total += event.Amount
					}
					if event.Kind == game.EventPermanentEnteredBattlefield {
						arrivals++
					}
				}
				if total != want || arrivals != 1 {
					t.Fatalf("counter event total/arrivals = %d/%d, want %d/1", total, arrivals, want)
				}
			}
			if !reflect.DeepEqual(g.Events, replay.Events) || !reflect.DeepEqual(g.ReplacementDecisions, replay.ReplacementDecisions) {
				t.Fatal("identical cloned entry branches did not replay deterministically")
			}
		})
	}
}

func entryCopyCreatureDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name: "Entry copier", Types: []types.Card{types.Creature},
		ReplacementAbilities: []game.ReplacementAbility{game.EntersAsCopyReplacement(
			"You may enter as a copy of a creature.",
			&game.Selection{RequiredTypes: []types.Card{types.Creature}},
			true, false, nil, false, nil, nil,
		)},
	}}
}

func TestEntryCopyAcquiresTappedAndColorChoiceAcrossEntryPaths(t *testing.T) {
	for _, path := range []string{"card", "token", "simultaneous"} {
		t.Run(path, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			e := NewEngine(nil)
			addReplacementPermanent(t, g, game.Player1, &game.CardDef{CardFace: game.CardFace{
				Name: "Copied entry choices", Types: []types.Card{types.Creature},
				ReplacementAbilities: []game.ReplacementAbility{
					game.EntersTappedReplacement("Enter tapped."),
					game.EntryColorChoiceReplacement("Choose a color."),
				},
			}})
			policy := &controlScript{choice: func(_ PlayerObservation, request game.ChoiceRequest) []int {
				if len(g.Battlefield) != 1 || request.Player != game.Player2 {
					t.Fatal("entry choice saw an entrant or used the owner instead of its controller")
				}
				if request.Prompt == "Choose a color." {
					return []int{4}
				}
				return fallbackChoice(request)
			}}
			agents := [game.NumPlayers]PlayerAgent{game.Player2: policy}
			def := entryCopyCreatureDef()
			start := len(g.Events)
			var permanent *game.Permanent
			var ok bool
			switch path {
			case "card":
				cardID := addCardToHand(g, game.Player1, def)
				permanent, ok = createCardPermanentWithChoices(e, g, g.CardInstances[cardID], game.Player2, zone.Hand, agents, &TurnLog{})
			case "token":
				permanent, ok = createTokenPermanentWithChoices(e, g, game.Player2, def, agents, &TurnLog{})
			case "simultaneous":
				cardID := addCardToHand(g, game.Player1, def)
				entry, prepared := prepareCardPermanentFaceForSimultaneousEntry(e, g, g.CardInstances[cardID],
					game.Player2, zone.Hand, game.FaceFront, nil, permanentCreationOptions{}, agents, &TurnLog{})
				ok = prepared
				if prepared {
					permanent = entry.permanent
					commitSimultaneousCardPermanentEntries(g, []preparedCardPermanentEntry{entry})
				}
			default:
				t.Fatalf("unknown entry path %q", path)
			}
			if !ok || permanent == nil || !permanent.Tapped || permanent.EntryChoices[game.EntryColorChoiceKey].Color != mana.G {
				t.Fatalf("entry = %+v/%v, want tapped with a fresh green choice", permanent, ok)
			}
			arrivals := 0
			for _, event := range g.Events[start:] {
				if event.Kind == game.EventPermanentTapped || event.Kind == game.EventPermanentUntapped {
					t.Fatal("entry emitted a tap/untap event")
				}
				if event.Kind == game.EventPermanentEnteredBattlefield {
					arrivals++
				}
			}
			if arrivals != 1 {
				t.Fatalf("arrival events = %d, want 1", arrivals)
			}
		})
	}
}

func TestEntryCopyPrecedenceDropsLostReplacements(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	e := NewEngine(nil)
	addReplacementPermanent(t, g, game.Player1, vanillaCreature("Plain copied creature", 2, 2))
	replacement := game.EntersAsCopyReplacement("Copy before other entry effects.",
		&game.Selection{RequiredTypes: []types.Card{types.Creature}}, false, false, nil, false, nil, nil).Replacement
	replacement.ID = g.IDGen.Next()
	g.ReplacementEffects = append(g.ReplacementEffects, replacement)
	def := &game.CardDef{CardFace: game.CardFace{
		Name: "Lose entry abilities", Types: []types.Card{types.Creature},
		ReplacementAbilities: []game.ReplacementAbility{
			game.EntersTappedReplacement("Enter tapped."),
			game.EntersWithCountersReplacement("Enter with three counters.", game.CounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 3}),
			game.EntersUnlessPaidElseZoneReplacement("Pay two life or go to the graveyard.",
				game.ResolutionPayment{AdditionalCosts: []cost.Additional{{Kind: cost.AdditionalPayLife, Amount: 2}}}, zone.Graveyard),
		},
		StaticAbilities: []game.StaticAbility{{Text: "Riot", KeywordAbilities: game.SimpleKeywords(game.Riot)}},
	}}
	policy := &controlScript{choice: func(_ PlayerObservation, request game.ChoiceRequest) []int {
		t.Fatalf("unexpected choice before forced copy or for lost ability: %+v", request)
		return nil
	}}
	cardID := addCardToHand(g, game.Player1, def)
	permanent, ok := createCardPermanentWithChoices(e, g, g.CardInstances[cardID], game.Player1,
		zone.Hand, [game.NumPlayers]PlayerAgent{game.Player1: policy}, &TurnLog{})
	if !ok || permanentEffectiveName(g, permanent) != "Plain copied creature" ||
		permanent.Tapped || permanent.Counters.Get(counter.PlusOnePlusOne) != 0 {
		t.Fatalf("stale entry abilities applied: %+v", permanent)
	}
}

func TestEntryCopiedPaymentCanRedirectEveryEntryPath(t *testing.T) {
	registry := cards.NewDefaultRegistry()
	steel, mox := registry.Lookup("Sculpting Steel"), registry.Lookup("Mox Diamond")
	if steel == nil || mox == nil {
		t.Fatal("registered Sculpting Steel or Mox Diamond missing")
	}
	for _, path := range []string{"spell", "effect", "simultaneous", "token"} {
		for _, paid := range []bool{false, true} {
			t.Run(path+map[bool]string{false: "/unpaid", true: "/paid"}[paid], func(t *testing.T) {
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				e := NewEngine(nil)
				addCombatPermanent(g, game.Player1, mox)
				var paymentCard game.ObjectID
				if paid {
					paymentCard = addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
						Name: "Payment land", Types: []types.Card{types.Land},
					}})
				}
				cardID := addCardToHand(g, game.Player1, steel)
				start := len(g.Events)
				switch path {
				case "spell":
					for range 3 {
						addBasicLandPermanent(g, game.Player1, types.Island)
					}
					g.Turn.Phase, g.Turn.Step = game.PhasePrecombatMain, game.StepNone
					if !e.applyAction(g, game.Player1, action.CastSpell(cardID, nil, 0, nil)) {
						t.Fatal("paid Sculpting Steel cast failed")
					}
					e.resolveTopOfStack(g, &TurnLog{})
				case "effect":
					r := effectResolver{engine: e, game: g, obj: &game.StackObject{Controller: game.Player1}, log: &TurnLog{}}
					r.putResolvedCardOnBattlefieldValue(g.CardInstances[cardID], zone.Hand, game.Player1, nil, permanentCreationOptions{})
				case "simultaneous":
					r := effectResolver{engine: e, game: g, obj: &game.StackObject{Controller: game.Player1}, log: &TurnLog{}}
					r.putResolvedCardsOnBattlefieldCollecting([]resolvedBattlefieldCard{{
						card: g.CardInstances[cardID], fromZone: zone.Hand, controller: game.Player1,
					}}, nil, permanentCreationOptions{})
				case "token":
					g.Players[game.Player1].Hand.Remove(cardID)
					createTokenPermanentWithChoices(e, g, game.Player1, steel, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
				default:
					t.Fatalf("unknown entry path %q", path)
				}
				if path != "token" {
					permanent := permanentByCardID(g, cardID)
					if (permanent != nil) != paid || g.Players[game.Player1].Hand.Contains(cardID) ||
						g.Players[game.Player1].Graveyard.Contains(cardID) == paid {
						t.Fatalf("copied payment entry/hand/graveyard = %v/%v/%v", permanent,
							g.Players[game.Player1].Hand.Contains(cardID), g.Players[game.Player1].Graveyard.Contains(cardID))
					}
				}
				arrivals, redirects, created, resolved := 0, 0, 0, 0
				for _, event := range g.Events[start:] {
					if event.Kind == game.EventSpellResolved && event.CardID == cardID {
						resolved++
					}
					if event.CardID != cardID && (path != "token" || event.TokenDef != steel) {
						continue
					}
					if event.Kind == game.EventPermanentEnteredBattlefield {
						arrivals++
					}
					if event.Kind == game.EventZoneChanged && event.ToZone == zone.Graveyard {
						redirects++
					}
					if event.Kind == game.EventTokenCreated {
						created++
					}
				}
				wantArrivals, wantRedirects := 0, 1
				if paid {
					wantArrivals, wantRedirects = 1, 0
				}
				if arrivals != wantArrivals || redirects != wantRedirects || path == "token" && created != 1 {
					t.Fatalf("arrival/redirect/created = %d/%d/%d", arrivals, redirects, created)
				}
				if path == "spell" && resolved != 1 {
					t.Fatalf("spell resolved events = %d, want 1 even when entry is replaced (CR 608.3)", resolved)
				}
				if paid && !g.Players[game.Player1].Graveyard.Contains(paymentCard) {
					t.Fatal("copied payment did not consume the land")
				}
			})
		}
	}
}

func TestEntryCopiedPaymentSharesSimultaneousResources(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	e := NewEngine(nil)
	registry := cards.NewDefaultRegistry()
	addCombatPermanent(g, game.Player1, registry.Lookup("Mox Diamond"))
	landID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Only payment land", Types: []types.Card{types.Land},
	}})
	var resolved []resolvedBattlefieldCard
	for range 2 {
		cardID := addCardToLibrary(g, game.Player1, registry.Lookup("Sculpting Steel"))
		resolved = append(resolved, resolvedBattlefieldCard{
			card: g.CardInstances[cardID], fromZone: zone.Library, controller: game.Player1,
		})
	}
	r := effectResolver{engine: e, game: g, obj: &game.StackObject{Controller: game.Player1}, log: &TurnLog{}}
	entered := r.putResolvedCardsOnBattlefieldCollecting(resolved, nil, permanentCreationOptions{})
	if len(entered) != 1 || len(g.Players[game.Player1].Graveyard.All()) != 2 ||
		!g.Players[game.Player1].Graveyard.Contains(landID) || len(g.Players[game.Player1].Library.All()) != 0 {
		t.Fatalf("cohort reused payment or lost a diverted card: entered %d, graveyard %v, library %v",
			len(entered), g.Players[game.Player1].Graveyard.All(), g.Players[game.Player1].Library.All())
	}
}

func TestEntryReplacementOrderIsChosenAndRecorded(t *testing.T) {
	for _, untapFirst := range []bool{false, true} {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		e := NewEngine(nil)
		addReplacementPermanent(t, g, game.Player2, &game.CardDef{CardFace: game.CardFace{
			Name: "Untapped entry", Types: []types.Card{types.Enchantment},
			ReplacementAbilities: []game.ReplacementAbility{game.EntersUntappedGroupReplacement(
				"Lands enter untapped.", game.TriggerControllerAny, types.Land)},
		}})
		index := 0
		if untapFirst {
			index = 1
		}
		policy := &controlScript{choice: func(_ PlayerObservation, request game.ChoiceRequest) []int {
			if request.Kind != game.ChoiceReplacement || request.Player != game.Player2 || len(request.Options) != 2 {
				t.Fatalf("replacement request = %+v", request)
			}
			return []int{index}
		}}
		def := &game.CardDef{CardFace: game.CardFace{
			Name: "Tapped entry", Types: []types.Card{types.Land},
			ReplacementAbilities: []game.ReplacementAbility{game.EntersTappedReplacement("This land enters tapped.")},
		}}
		cardID := addCardToHand(g, game.Player1, def)
		log := &TurnLog{}
		permanent, ok := createCardPermanentWithChoices(e, g, g.CardInstances[cardID], game.Player2,
			zone.Hand, [game.NumPlayers]PlayerAgent{game.Player2: policy}, log)
		if !ok || permanent.Tapped != untapFirst || len(policy.choices) != 1 {
			t.Fatalf("entry tapped/choices = %v/%d, want %v/1", permanent.Tapped, len(policy.choices), untapFirst)
		}
		if len(g.ReplacementDecisions) != 1 {
			t.Fatalf("replacement decisions = %+v", g.ReplacementDecisions)
		}
		decision := g.ReplacementDecisions[0]
		if decision.Player != game.Player2 || decision.UsedFallback || !slices.Equal(decision.Selected, []int{index}) {
			t.Fatalf("recorded replacement decision = %+v", decision)
		}
		if len(log.Choices) != 1 || log.Choices[0].Request.Kind != game.ChoiceReplacement ||
			!slices.Equal(log.Choices[0].Selected, decision.Selected) {
			t.Fatalf("replacement selection was not replay-logged: %+v", log.Choices)
		}
	}
}

func TestEntryReplacementDeclinedCopyAndPaymentAreNotRepeated(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	e := NewEngine(nil)
	addReplacementPermanent(t, g, game.Player1, vanillaCreature("Copy candidate", 2, 2))
	def := entryCopyCreatureDef()
	def.ReplacementAbilities = append(def.ReplacementAbilities,
		game.EntersTappedUnlessPaidReplacement("Pay two life or enter tapped.", game.ResolutionPayment{
			AdditionalCosts: []cost.Additional{{Kind: cost.AdditionalPayLife, Amount: 2}},
		}),
		game.EntersWithCountersReplacement("Enter with three counters.", game.CounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 3}),
	)
	copyChoices, paymentChoices := 0, 0
	policy := &controlScript{choice: func(_ PlayerObservation, request game.ChoiceRequest) []int {
		if strings.Contains(request.Prompt, "enter as a copy") {
			copyChoices++
			return []int{0}
		}
		if strings.Contains(request.Prompt, "Pay enter-the-battlefield cost") {
			paymentChoices++
			return []int{1}
		}
		return fallbackChoice(request)
	}}
	cardID := addCardToHand(g, game.Player1, def)
	permanent, ok := createCardPermanentWithChoices(e, g, g.CardInstances[cardID], game.Player1,
		zone.Hand, [game.NumPlayers]PlayerAgent{game.Player1: policy}, &TurnLog{})
	if !ok || permanent.Tapped || permanent.Counters.Get(counter.PlusOnePlusOne) != 3 ||
		g.Players[game.Player1].Life != 38 || copyChoices != 1 || paymentChoices != 1 {
		t.Fatalf("entry = %+v; life/copy/payment = %d/%d/%d", permanent, g.Players[game.Player1].Life, copyChoices, paymentChoices)
	}
}

func TestEntryReplacementRecomputesGroupApplicability(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addReplacementPermanent(t, g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Creature entry counters", Types: []types.Card{types.Enchantment},
		ReplacementAbilities: []game.ReplacementAbility{game.EntersWithCountersGroupReplacement(
			"Creatures you control enter with a counter.",
			&game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou},
			game.CounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 1},
		)},
	}})
	addReplacementPermanent(t, g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Animate entering artifacts", Types: []types.Card{types.Enchantment},
		ReplacementAbilities: []game.ReplacementAbility{game.EntersBecomesGroupReplacement(
			"Artifacts enter as creatures.", game.EntersBecomesGroupParams{
				Controller: game.TriggerControllerYou, SubjectTypes: []types.Card{types.Artifact},
				AddTypes: []types.Card{types.Creature}, BasePower: opt.Val(7), BaseToughness: opt.Val(7),
			},
		)},
	}})
	permanent := addReplacementPermanent(t, g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Animated on entry", Types: []types.Card{types.Artifact},
	}})
	if !permanentHasType(g, permanent, types.Creature) || permanent.Counters.Get(counter.PlusOnePlusOne) != 1 {
		t.Fatalf("newly applicable creature replacement was missed or repeated: %+v", permanent)
	}
}

func TestEntryCopyRechecksContinuousAbilityRemoval(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	e := NewEngine(nil)
	addBloodMoon(t, g)
	addReplacementPermanent(t, g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Nonbasic choice land", Types: []types.Card{types.Land},
		ReplacementAbilities: []game.ReplacementAbility{
			game.EntersTappedReplacement("Enter tapped."),
			game.EntryColorChoiceReplacement("Choose a color."),
		},
	}})
	def := entryCopyCreatureDef()
	def.ReplacementAbilities[0].Replacement.EntersAsCopySelection = &game.Selection{RequiredTypes: []types.Card{types.Land}}
	policy := &controlScript{choice: func(_ PlayerObservation, request game.ChoiceRequest) []int {
		if request.Kind == game.ChoiceReplacement || request.Prompt == "Choose a color." {
			t.Fatalf("ability erased by the would-be battlefield's Blood Moon was offered: %+v", request)
		}
		return fallbackChoice(request)
	}}
	cardID := addCardToHand(g, game.Player1, def)
	permanent, ok := createCardPermanentWithChoices(e, g, g.CardInstances[cardID], game.Player1,
		zone.Hand, [game.NumPlayers]PlayerAgent{game.Player1: policy}, &TurnLog{})
	if !ok || !permanentHasSubtype(g, permanent, types.Mountain) || permanent.Tapped || len(permanent.EntryChoices) != 0 {
		t.Fatalf("copy did not use current would-be characteristics: %+v", permanent)
	}
}

func TestEntryCopySimultaneousCohortKeepsSourcesAndChoicesPending(t *testing.T) {
	for _, reverse := range []bool{false, true} {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		e := NewEngine(nil)
		registry := cards.NewDefaultRegistry()
		addReplacementPermanent(t, g, game.Player1, registry.Lookup("Triskelion"))
		cloneID := addCardToLibrary(g, game.Player1, registry.Lookup("Clone"))
		sourceID := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
			Name: "Pending tap source", Types: []types.Card{types.Artifact},
			ReplacementAbilities: []game.ReplacementAbility{game.EntersTappedGroupReplacement(
				"Other permanents enter tapped.", game.TriggerControllerAny)},
		}})
		resolved := []resolvedBattlefieldCard{
			{card: g.CardInstances[cloneID], fromZone: zone.Library, controller: game.Player1},
			{card: g.CardInstances[sourceID], fromZone: zone.Library, controller: game.Player1},
		}
		if reverse {
			slices.Reverse(resolved)
		}
		policy := &controlScript{choice: func(_ PlayerObservation, request game.ChoiceRequest) []int {
			if len(g.Battlefield) != 1 {
				t.Fatal("choice observed a partially published simultaneous cohort")
			}
			return fallbackChoice(request)
		}}
		r := effectResolver{engine: e, game: g, obj: &game.StackObject{Controller: game.Player1},
			agents: [game.NumPlayers]PlayerAgent{game.Player1: policy}, log: &TurnLog{}}
		arrivals := r.putResolvedCardsOnBattlefieldCollecting(resolved, nil, permanentCreationOptions{})
		clone := permanentByCardID(g, cloneID)
		if len(arrivals) != 2 || clone == nil || clone.Tapped || clone.Counters.Get(counter.PlusOnePlusOne) != 3 {
			t.Fatalf("simultaneous Clone = %+v, arrivals = %d", clone, len(arrivals))
		}
		var cohort game.Event
		count := 0
		for _, event := range g.Events {
			if event.Kind != game.EventPermanentEnteredBattlefield || event.CardID != cloneID && event.CardID != sourceID {
				continue
			}
			if cohort.SimultaneousID != 0 && event.SimultaneousID != cohort.SimultaneousID {
				t.Fatal("cohort arrival correlation was lost")
			}
			cohort = event
			count++
		}
		if count != 2 || cohort.SimultaneousID == 0 {
			t.Fatalf("cohort arrivals = %d, simultaneous ID = %d", count, cohort.SimultaneousID)
		}
		later := addReplacementPermanent(t, g, game.Player1, vanillaCreature("Later creature", 2, 2))
		if !later.Tapped {
			t.Fatal("committed replacement source did not affect a later entry")
		}
	}
}

func TestEntryTokenInitialTappedStateCanBeReplaced(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addReplacementPermanent(t, g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Untapped tokens", Types: []types.Card{types.Enchantment},
		ReplacementAbilities: []game.ReplacementAbility{game.EntersUntappedGroupReplacement(
			"Permanents enter untapped.", game.TriggerControllerAny)},
	}})
	permanent, ok := createTokenPermanentWithChoicesInBatch(NewEngine(nil), g, game.Player1,
		vanillaCreature("Tapped by creation", 2, 2), 0, true, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if !ok || permanent.Tapped {
		t.Fatalf("token creation overrode its entry replacement: %+v", permanent)
	}
}

func TestEntryCopyAcquiredCopyAbilityHasItsOwnIdentity(t *testing.T) {
	for _, copyAgain := range []bool{false, true} {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		e := NewEngine(nil)
		addCombatPermanent(g, game.Player1, entryCopyCreatureDef())
		addCombatPermanent(g, game.Player1, vanillaCreature("Final creature", 2, 2))
		def := entryCopyCreatureDef()
		def.ReplacementAbilities[0].Replacement.EntersAsCopyConditionalCounters = []game.ConditionalCounterPlacement{{
			Kind: counter.PlusOnePlusOne, Amount: 1, IfType: types.Creature,
		}}
		optionalChoices, targetChoices := 0, 0
		policy := &controlScript{choice: func(_ PlayerObservation, request game.ChoiceRequest) []int {
			if request.Prompt == "Have this permanent enter as a copy?" {
				optionalChoices++
				if optionalChoices > 2 {
					t.Fatal("an already handled copy ability applied again")
				}
				if optionalChoices == 2 && !copyAgain {
					return []int{0}
				}
				return []int{1}
			}
			if request.Prompt == "Choose a permanent to copy" {
				targetChoices++
				return []int{targetChoices - 1}
			}
			return fallbackChoice(request)
		}}
		cardID := addCardToHand(g, game.Player1, def)
		permanent, ok := createCardPermanentWithChoices(e, g, g.CardInstances[cardID], game.Player1,
			zone.Hand, [game.NumPlayers]PlayerAgent{game.Player1: policy}, &TurnLog{})
		wantName, wantCounters := "Entry copier", 1
		if copyAgain {
			wantName, wantCounters = "Final creature", 0
		}
		if !ok || permanentEffectiveName(g, permanent) != wantName ||
			permanent.Counters.Get(counter.PlusOnePlusOne) != wantCounters || optionalChoices != 2 {
			t.Fatalf("chained entry = %+v; optional choices = %d; want %s with %d counters", permanent, optionalChoices, wantName, wantCounters)
		}
	}
}

func TestEntryReplacementDistinctDeclarationsApplyOnceEach(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	ability := game.EntersWithCountersReplacement("Enter with three counters.",
		game.CounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 3})
	permanent := addReplacementPermanent(t, g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Two distinct replacements", Types: []types.Card{types.Creature},
		ReplacementAbilities: []game.ReplacementAbility{ability, ability},
	}})
	if permanent.Counters.Get(counter.PlusOnePlusOne) != 6 || len(g.ReplacementDecisions) != 1 {
		t.Fatalf("distinct entry effects conflated or repeated: %+v / %+v", permanent, g.ReplacementDecisions)
	}
}
