package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review: official Comprehensive Rules effective November 6, 2026
// (https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt),
// read 702.37a-g, 708.2-708.11, 701.40b-c, 701.58b-c, 116.2b and 122.6a.
// 702.37b: a megamorph cost is a morph cost, and "as this permanent is turned
// face up, put a +1/+1 counter on it if its megamorph cost was paid to turn it
// face up"; 708.11 applies that "as turned face up" ability while turning it
// face up; 701.40c/701.58c let a manifested or cloaked card with morph turn
// face up by either procedure. Pinned corpus Oracle text: Ainok Survivalist
// 7a5dae5f-29ab-4996-b236-c754de8769bc (printing 0f4b5b18), Aerie Bowmasters
// 43ace4d5-9006-4bad-bd17-3d368c20564d (printing a6db26b5) and Marsh Hulk
// 979a581b-22a0-4461-8638-d219587323d7 (printing ff347b25); each rulings_uri
// returned the same three 2015-02-25 rulings: turning face up and adding the
// counter is a special action, turning a manifested megamorph card face up
// for its mana cost adds no counter, and megamorph is a variant of morph.
// These cards are not registered yet; the fixtures below carry their exact
// Oracle text and only the abilities the scenarios exercise. Hardened Scales,
// Rattleclaw Mystic and Temur War Shaman are the registered cards.

// megamorphTestCard loads a registered legacy card by its exact name; the
// Oracle ID is recorded for the source review only.
func megamorphTestCard(t *testing.T, _ /* oracleID */, name string) *game.CardDef {
	t.Helper()
	def, err := cards.NewDefaultRegistry().LookupChecked(name)
	if err != nil || def == nil {
		t.Fatalf("registry lookup %s = %v/%v", name, def, err)
	}
	return def
}

func ainokSurvivalistFixture() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:      "Ainok Survivalist",
		ManaCost:  opt.Val(cost.Mana{cost.O(1), cost.G}),
		Colors:    []color.Color{color.Green},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Dog, types.Shaman},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 1}),
		StaticAbilities: []game.StaticAbility{{
			Text:             "Megamorph {1}{G}",
			KeywordAbilities: []game.KeywordAbility{game.MorphKeyword{Cost: cost.Mana{cost.O(1), cost.G}, Megamorph: true}},
		}},
		TriggeredAbilities: []game.TriggeredAbility{{
			Text: "When this creature is turned face up, destroy target artifact or enchantment an opponent controls.",
			Trigger: game.TriggerCondition{Type: game.TriggerWhen, Pattern: game.TriggerPattern{
				Event: game.EventPermanentTurnedFaceUp, Source: game.TriggerSourceSelf,
			}},
			Content: game.AbilityContent{MinModes: 1, MaxModes: 1, Modes: []game.Mode{{
				Targets: []game.TargetSpec{{
					MinTargets: 1, MaxTargets: 1,
					Constraint: "target artifact or enchantment an opponent controls",
					Allow:      game.TargetAllowPermanent,
					Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Enchantment}, Controller: game.ControllerOpponent}),
				}},
				Sequence: []game.Instruction{{Primitive: game.Destroy{Object: game.TargetPermanentReference(0)}}},
			}}},
		}},
		OracleText: `
			Megamorph {1}{G} (You may cast this card face down as a 2/2 creature for {3}. Turn it face up any time for its megamorph cost and put a +1/+1 counter on it.)
			When this creature is turned face up, destroy target artifact or enchantment an opponent controls.
		`,
	}}
}

func aerieBowmastersFixture() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:      "Aerie Bowmasters",
		ManaCost:  opt.Val(cost.Mana{cost.O(2), cost.G, cost.G}),
		Colors:    []color.Color{color.Green},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Dog, types.Archer},
		Power:     opt.Val(game.PT{Value: 3}),
		Toughness: opt.Val(game.PT{Value: 4}),
		StaticAbilities: []game.StaticAbility{
			game.ReachStaticBody,
			{
				Text:             "Megamorph {5}{G}",
				KeywordAbilities: []game.KeywordAbility{game.MorphKeyword{Cost: cost.Mana{cost.O(5), cost.G}, Megamorph: true}},
			},
		},
		OracleText: `
			Reach (This creature can block creatures with flying.)
			Megamorph {5}{G} (You may cast this card face down as a 2/2 creature for {3}. Turn it face up any time for its megamorph cost and put a +1/+1 counter on it.)
		`,
	}}
}

func marshHulkFixture() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:      "Marsh Hulk",
		ManaCost:  opt.Val(cost.Mana{cost.O(4), cost.B, cost.B}),
		Colors:    []color.Color{color.Black},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Zombie, types.Ogre},
		Power:     opt.Val(game.PT{Value: 4}),
		Toughness: opt.Val(game.PT{Value: 6}),
		StaticAbilities: []game.StaticAbility{{
			Text:             "Megamorph {6}{B}",
			KeywordAbilities: []game.KeywordAbility{game.MorphKeyword{Cost: cost.Mana{cost.O(6), cost.B}, Megamorph: true}},
		}},
		OracleText: `
			Megamorph {6}{B} (You may cast this card face down as a 2/2 creature for {3}. Turn it face up any time for its megamorph cost and put a +1/+1 counter on it.)
		`,
	}}
}

// megamorphMayPolicy accepts every optional choice and records the requests.
type megamorphMayPolicy struct {
	simPassPolicy

	payment  []int
	requests []game.ChoiceRequest
}

func (p *megamorphMayPolicy) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	p.requests = append(p.requests, request)
	switch request.Kind {
	case game.ChoiceMay:
		return []int{1}
	case game.ChoicePayment:
		if p.payment != nil && request.Prompt == "Choose turn face-up cost" {
			return p.payment
		}
	}
	return request.DefaultSelection
}

func megamorphAgents(policy *megamorphMayPolicy) [game.NumPlayers]PlayerAgent {
	agents := simPassPolicies()
	agents[game.Player1] = policy
	return agents
}

func resolveMegamorphTriggers(t *testing.T, g *game.Game, engine *Engine, agents [game.NumPlayers]PlayerAgent) {
	t.Helper()
	engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{})
	for !g.Stack.IsEmpty() {
		engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
		engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{})
	}
}

func TestAinokSurvivalistCastFaceDownThenMegamorphFaceUp(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToHand(g, game.Player1, ainokSurvivalistFixture())
	relic := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Opposing Relic", Types: []types.Card{types.Enchantment}}})
	ownRelic := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Own Relic", Types: []types.Card{types.Artifact}}})
	g.Players[game.Player1].ManaPool.Add(mana.C, 3)
	g.Turn.ActivePlayer = game.Player1
	g.Turn.PriorityPlayer = game.Player1
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if !engine.applyAction(g, game.Player1, actionBuild.castFaceDown(cardID, game.FaceFront, game.FaceDownMorph)) {
		t.Fatal("megamorph card could not be cast face down for {3}")
	}
	if got := g.Players[game.Player1].ManaPool.Total(); got != 0 {
		t.Fatalf("mana left after face-down cast = %d, want {3} spent", got)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	var ainok *game.Permanent
	for _, permanent := range g.Battlefield {
		if permanent.CardInstanceID == cardID {
			ainok = permanent
		}
	}
	if ainok == nil || !ainok.FaceDown || ainok.FaceDownKind != game.FaceDownMorph {
		t.Fatalf("face-down Ainok permanent = %+v", ainok)
	}
	seen := false
	for _, view := range observe(g, game.Player2).Battlefield() {
		if view.ObjectID != ainok.ObjectID {
			continue
		}
		seen = true
		if !view.FaceDown || view.Name != "" || view.Power != 2 || view.Toughness != 2 || len(view.FaceDownCards) != 0 {
			t.Fatalf("opponent view = %+v, want hidden nameless 2/2", view)
		}
	}
	if !seen {
		t.Fatal("opponent observation omitted the face-down creature")
	}
	if len(effectivePermanentValues(g, ainok).colors) != 0 {
		t.Fatal("face-down megamorph creature is not colorless")
	}

	g.Players[game.Player1].ManaPool.Add(mana.G, 2)
	agents := megamorphAgents(&megamorphMayPolicy{})
	if !engine.applyActionWithChoices(g, game.Player1, actionBuild.turnFaceUp(ainok.ObjectID), agents, &TurnLog{}) {
		t.Fatal("turn face up for megamorph {1}{G} failed")
	}
	toughness, _ := effectiveToughness(g, ainok)
	if ainok.FaceDown || permanentEffectiveName(g, ainok) != "Ainok Survivalist" || effectivePower(g, ainok) != 3 || toughness != 2 {
		t.Fatalf("face-up Ainok = %q %d/%d faceDown=%t, want Ainok Survivalist 3/2", permanentEffectiveName(g, ainok), effectivePower(g, ainok), toughness, ainok.FaceDown)
	}
	if got := ainok.Counters.Get(counter.PlusOnePlusOne); got != 1 {
		t.Fatalf("+1/+1 counters = %d, want 1", got)
	}
	resolveMegamorphTriggers(t, g, engine, agents)
	if _, ok := permanentByObjectID(g, relic.ObjectID); ok {
		t.Fatal("turned-face-up trigger did not destroy the opponent's enchantment")
	}
	if _, ok := permanentByObjectID(g, ownRelic.ObjectID); !ok {
		t.Fatal("turned-face-up trigger destroyed its controller's own artifact")
	}
}

func TestMegamorphCounterIsModifiedByHardenedScales(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addReplacementPermanent(t, g, game.Player1, megamorphTestCard(t, "a1f3da21-af6d-450e-bf0b-985d158418e6", "Hardened Scales"))
	bowmasters := addFaceDownPermanent(g, game.Player1, aerieBowmastersFixture(), game.FaceDownMorph)
	g.Players[game.Player1].ManaPool.Add(mana.G, 6)
	g.Turn.PriorityPlayer = game.Player1

	if !engine.applyAction(g, game.Player1, actionBuild.turnFaceUp(bowmasters.ObjectID)) {
		t.Fatal("turn face up for megamorph {5}{G} failed")
	}
	if got := bowmasters.Counters.Get(counter.PlusOnePlusOne); got != 2 {
		t.Fatalf("+1/+1 counters with Hardened Scales = %d, want 2", got)
	}
	if toughness, _ := effectiveToughness(g, bowmasters); effectivePower(g, bowmasters) != 5 || toughness != 6 {
		t.Fatalf("Aerie Bowmasters = %d/%d, want 5/6", effectivePower(g, bowmasters), toughness)
	}
}

func TestRattleclawMysticMorphFaceUpGetsNoCounter(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	mystic := addFaceDownPermanent(g, game.Player1, megamorphTestCard(t, "84c9d8b8-8dd9-41fc-a2b6-af1ffb9090f9", "Rattleclaw Mystic"), game.FaceDownMorph)
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)
	g.Turn.PriorityPlayer = game.Player1
	agents := megamorphAgents(&megamorphMayPolicy{})

	if !engine.applyActionWithChoices(g, game.Player1, actionBuild.turnFaceUp(mystic.ObjectID), agents, &TurnLog{}) {
		t.Fatal("turn face up for morph {2} failed")
	}
	if got := mystic.Counters.Get(counter.PlusOnePlusOne); got != 0 {
		t.Fatalf("+1/+1 counters after morph = %d, want 0", got)
	}
	resolveMegamorphTriggers(t, g, engine, agents)
	pool := g.Players[game.Player1].ManaPool
	if pool.Amount(mana.G) != 1 || pool.Amount(mana.U) != 1 || pool.Amount(mana.R) != 1 {
		t.Fatalf("mana pool after trigger = G%d U%d R%d, want {G}{U}{R}", pool.Amount(mana.G), pool.Amount(mana.U), pool.Amount(mana.R))
	}
}

func TestManifestedMarshHulkOffersManaCostAndMegamorphCost(t *testing.T) {
	for _, tc := range []struct {
		name     string
		pick     int
		counters int
		power    int
		left     int
	}{
		{name: "mana cost", pick: 0, counters: 0, power: 4, left: 1},
		{name: "megamorph cost", pick: 1, counters: 1, power: 5, left: 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			hulkID := addCardToLibrary(g, game.Player1, marshHulkFixture())
			shamanID := addCardToHand(g, game.Player1, megamorphTestCard(t, "fb683e18-50b8-4821-9a1a-86c6dea63ac9", "Temur War Shaman"))
			g.Players[game.Player1].ManaPool.Add(mana.G, 6)
			g.Turn.ActivePlayer = game.Player1
			g.Turn.PriorityPlayer = game.Player1
			g.Turn.Phase = game.PhasePrecombatMain
			g.Turn.Step = game.StepNone
			policy := &megamorphMayPolicy{payment: []int{tc.pick}}
			agents := megamorphAgents(policy)

			if !engine.applyActionWithChoices(g, game.Player1, actionBuild.castSpell(shamanID, zone.Hand, game.FaceFront, nil, 0, nil), agents, &TurnLog{}) {
				t.Fatal("Temur War Shaman cast failed")
			}
			engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
			resolveMegamorphTriggers(t, g, engine, agents)
			var hulk *game.Permanent
			for _, permanent := range g.Battlefield {
				if permanent.CardInstanceID == hulkID {
					hulk = permanent
				}
			}
			if hulk == nil || !hulk.FaceDown || hulk.FaceDownKind != game.FaceDownManifest {
				t.Fatalf("manifested Marsh Hulk = %+v", hulk)
			}

			g.Players[game.Player1].ManaPool.Add(mana.B, 7)
			if !engine.applyActionWithChoices(g, game.Player1, actionBuild.turnFaceUp(hulk.ObjectID), agents, &TurnLog{}) {
				t.Fatal("turn manifested Marsh Hulk face up failed")
			}
			var labels []string
			for _, request := range policy.requests {
				if request.Prompt == "Choose turn face-up cost" {
					for _, option := range request.Options {
						labels = append(labels, option.Label)
					}
				}
			}
			if want := []string{"Pay {4}{B}{B} (mana cost)", "Pay {6}{B} (megamorph)"}; !slices.Equal(labels, want) {
				t.Fatalf("turn face-up cost options = %q, want %q", labels, want)
			}
			if hulk.FaceDown || permanentEffectiveName(g, hulk) != "Marsh Hulk" {
				t.Fatalf("Marsh Hulk faceDown=%t name=%q", hulk.FaceDown, permanentEffectiveName(g, hulk))
			}
			if got := hulk.Counters.Get(counter.PlusOnePlusOne); got != tc.counters {
				t.Fatalf("+1/+1 counters = %d, want %d", got, tc.counters)
			}
			if effectivePower(g, hulk) != tc.power {
				t.Fatalf("Marsh Hulk power = %d, want %d", effectivePower(g, hulk), tc.power)
			}
			if got := g.Players[game.Player1].ManaPool.Amount(mana.B); got != tc.left {
				t.Fatalf("black mana left = %d, want %d", got, tc.left)
			}
		})
	}
}

// TestManifestEqualManaAndMegamorphCostsStayDistinct proves equal mana costs
// from different sources are not merged: only the megamorph choice adds the
// counter (CR 702.37b; 2015-02-25 megamorph manifest ruling).
func TestManifestEqualManaAndMegamorphCostsStayDistinct(t *testing.T) {
	for _, tc := range []struct {
		pick     int
		counters int
	}{{pick: 0, counters: 0}, {pick: 1, counters: 1}} {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		def := manifestMorphCreature(cost.Mana{cost.O(1), cost.G}, cost.Mana{cost.O(1), cost.G})
		def.StaticAbilities[0].KeywordAbilities = []game.KeywordAbility{game.MorphKeyword{Cost: cost.Mana{cost.O(1), cost.G}, Megamorph: true}}
		permanent := addFaceDownPermanent(g, game.Player1, def, game.FaceDownManifest)
		g.Players[game.Player1].ManaPool.Add(mana.G, 2)
		g.Turn.PriorityPlayer = game.Player1
		policy := &megamorphMayPolicy{payment: []int{tc.pick}}

		if !engine.applyActionWithChoices(g, game.Player1, actionBuild.turnFaceUp(permanent.ObjectID), megamorphAgents(policy), &TurnLog{}) {
			t.Fatal("turn face up failed")
		}
		if len(policy.requests) == 0 || len(policy.requests[0].Options) != 2 ||
			policy.requests[0].Options[0].Label == policy.requests[0].Options[1].Label {
			t.Fatalf("cost choice requests = %+v, want two distinctly labelled options", policy.requests)
		}
		if got := permanent.Counters.Get(counter.PlusOnePlusOne); got != tc.counters {
			t.Fatalf("pick %d: +1/+1 counters = %d, want %d", tc.pick, got, tc.counters)
		}
	}
}

func TestTemurWarShamanFightSeesMegamorphCounter(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, megamorphTestCard(t, "fb683e18-50b8-4821-9a1a-86c6dea63ac9", "Temur War Shaman"))
	ainok := addFaceDownPermanent(g, game.Player1, ainokSurvivalistFixture(), game.FaceDownMorph)
	pt := game.PT{Value: 1}
	four := game.PT{Value: 4}
	wall := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Opposing Wall",
		Types: []types.Card{types.Creature}, Power: opt.Val(pt), Toughness: opt.Val(four)}})
	g.Players[game.Player1].ManaPool.Add(mana.G, 2)
	g.Turn.PriorityPlayer = game.Player1
	agents := megamorphAgents(&megamorphMayPolicy{})

	if !engine.applyActionWithChoices(g, game.Player1, actionBuild.turnFaceUp(ainok.ObjectID), agents, &TurnLog{}) {
		t.Fatal("turn face up for megamorph failed")
	}
	counterAt, faceUpAt := -1, -1
	for i, event := range g.Events {
		switch {
		case event.Kind == game.EventCountersAdded && event.PermanentID == ainok.ObjectID:
			counterAt = i
		case event.Kind == game.EventPermanentTurnedFaceUp && event.PermanentID == ainok.ObjectID:
			faceUpAt = i
		}
	}
	if counterAt < 0 || faceUpAt < 0 || counterAt > faceUpAt {
		t.Fatalf("counter event index %d, turned-face-up event index %d; want counter first", counterAt, faceUpAt)
	}
	resolveMegamorphTriggers(t, g, engine, agents)
	if wall.MarkedDamage != 3 {
		t.Fatalf("fight damage dealt by Ainok = %d, want 3 from the countered 3/2", wall.MarkedDamage)
	}
	if ainok.MarkedDamage != 1 {
		t.Fatalf("fight damage dealt to Ainok = %d, want 1", ainok.MarkedDamage)
	}
}
