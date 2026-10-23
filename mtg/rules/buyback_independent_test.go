package rules

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Independent #170 review at f2c7f5c90c4d38aef2271b5ed81f639e4a55188d.
// Read the supplied verified CR effective 2026-09-25 (SHA-256
// 8d860e451f20f38865b725b42d82feb714c725373dd8f3b32b8652b3eeb070ca):
// 601.2, 608.2b/n, 702.27a, 707.10/12, 400.7, 614.15, 616.1,
// 702.34a, 702.40, 118.8/9. Read pinned Reiterate printing
// 99302c41-434f-41ff-a61a-2c3681a0c135 and its full seven WotC
// 2021-03-19 rulings. The minimal fixture here isolates Engine cost and copy
// probes; production Reiterate behavior is exercised in
// buyback_interaction_test.go.

func independentBuybackAgent(t *testing.T, choose func(game.ChoiceRequest) string) [game.NumPlayers]PlayerAgent {
	t.Helper()
	return [game.NumPlayers]PlayerAgent{game.Player1: &authoredCopyChoiceAgent{
		choose: func(request game.ChoiceRequest) []int {
			label := choose(request)
			if label == "" {
				return request.DefaultSelection
			}
			for _, option := range request.Options {
				if option.Label == label {
					return []int{option.Index}
				}
			}
			t.Fatalf("choice %q not offered: %+v", label, request)
			return nil
		},
	}}
}

func independentChooseBuyback(request game.ChoiceRequest) string {
	if request.Kind == game.ChoicePayment {
		for _, option := range request.Options {
			if strings.Contains(option.Label, "Buyback") {
				return option.Label
			}
		}
	}
	return ""
}

func TestBuybackIndependentStormCopies(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := buybackEngine(t)
	subject := castBuybackSubject(t, g, engine)
	def := reiterateFixture()
	def.StaticAbilities = append(def.StaticAbilities, game.StaticAbility{
		KeywordAbilities: game.SimpleKeywords(game.Storm),
	})
	discarded := addCardToHand(g, game.Player1, vanillaCreature("Storm payment subject", 2, 2))
	def.AdditionalCosts = []cost.Additional{{
		Kind: cost.AdditionalDiscard, Amount: 1, SubjectKey: "storm-discard",
	}}
	cardID := addCardToHand(g, game.Player1, def)
	fundBuyback(g, 4)
	log := &TurnLog{}
	if !engine.applyActionWithChoices(g, game.Player1, action.CastSpell(cardID,
		[]game.Target{game.StackObjectTarget(subject.ID)}, 0, nil),
		independentBuybackAgent(t, independentChooseBuyback), log) {
		t.Fatal("paid original with storm failed to cast")
	}
	objects := g.Stack.Objects()
	if len(objects) != 3 || !objects[1].BuybackPaid || !objects[2].Copy {
		t.Fatalf("expected subject, paid original, and real storm copy: %+v", objects)
	}
	for _, branch := range []*game.Game{g, g.Clone()} {
		copied, _ := branch.Stack.Peek()
		if !copied.BuybackPaid {
			t.Error("storm copy lost the original's paid buyback decision (CR 707.10)")
		}
		if len(copied.PaidCostSubjects) != 1 ||
			copied.PaidCostSubjects[0].Snapshot.CardID != discarded {
			t.Fatal("storm copy lost its original payment subject")
		}
		engine.resolveTopOfStack(branch, &TurnLog{})
		if branch.Players[game.Player1].Hand.Contains(cardID) ||
			branch.CardInstances[cardID].ZoneVersion != 1 ||
			branch.Players[game.Player1].ManaPool.Total() != 0 {
			t.Fatal("storm copy moved the physical original or paid again")
		}
		engine.resolveTopOfStack(branch, &TurnLog{}) // storm copy's subject copy
		engine.resolveTopOfStack(branch, &TurnLog{}) // physical Reiterate
		if !branch.Players[game.Player1].Hand.Contains(cardID) ||
			branch.CardInstances[cardID].ZoneVersion != 2 {
			t.Fatal("physical original failed to return after its storm copy resolved")
		}
	}
}

func TestBuybackIndependentCastCardCopyOffersPayment(t *testing.T) {
	for _, copyCard := range []bool{false, true} {
		t.Run(map[bool]string{false: "physical-free-cast-control", true: "cast-copy-of-exiled-card"}[copyCard], func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := buybackEngine(t)
			subject := castBuybackSubject(t, g, engine)
			cardID := addCardToHand(g, game.Player1, reiterateFixture())
			if !moveCardBetweenZones(g, game.Player1, cardID, zone.Hand, zone.Exile) {
				t.Fatal("could not put the source card in exile")
			}
			g.Players[game.Player1].ManaPool.Add(mana.C, 3)
			paymentOffers := 0
			agents := independentBuybackAgent(t, func(request game.ChoiceRequest) string {
				label := independentChooseBuyback(request)
				if label != "" {
					paymentOffers++
				}
				return label
			})
			log := &TurnLog{}
			ok := false
			if copyCard {
				ok = engine.castFreeCopyOfCard(g, game.Player1, cardID, agents, log)
			} else {
				ok = engine.castFreeOption(g, g.Players[game.Player1], game.Player1,
					g.CardInstances[cardID], action.CastSpellAction{
						CardID: cardID, SourceZone: zone.Exile,
						Targets: []game.Target{game.StackObjectTarget(subject.ID)},
					}, false, agents, log)
			}
			if !ok {
				t.Fatal("otherwise legal free cast failed")
			}
			obj, _ := g.Stack.Peek()
			if paymentOffers != 1 || !obj.BuybackPaid || obj.ManaSpentToCast != 3 ||
				g.Players[game.Player1].ManaPool.Total() != 0 {
				t.Errorf("cast must offer/pay optional buyback: offers=%d paid=%v spent=%d pool=%d",
					paymentOffers, obj.BuybackPaid, obj.ManaSpentToCast,
					g.Players[game.Player1].ManaPool.Total())
			}
			if obj.Copy != copyCard {
				t.Fatal("casting-copy identity wrong")
			}
			engine.resolveTopOfStack(g, log)
			if g.Stack.Size() != 2 || g.Players[game.Player1].Hand.Contains(cardID) == copyCard ||
				g.Players[game.Player1].Exile.Contains(cardID) != copyCard {
				t.Fatal("resolution must copy the target without moving an imprinted physical card")
			}
		})
	}
}

func TestBuybackIndependentReplacementChainAndClone(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := buybackEngine(t)
	castBuybackSubject(t, g, engine)
	cardID := addCardToHand(g, game.Player4, reiterateFixture())
	discarded := addCardToHand(g, game.Player1, vanillaCreature("Clone payment subject", 1, 1))
	g.CardInstances[cardID].Def.AdditionalCosts = []cost.Additional{{
		Kind: cost.AdditionalDiscard, Amount: 1, SubjectKey: "clone-discard",
	}}
	fundBuyback(g, 4)
	if !engine.castPaidSpellFromSource(g, g.Players[game.Player4], game.Player1,
		cardID, zone.Hand, independentBuybackAgent(t, independentChooseBuyback), &TurnLog{}) {
		t.Fatal("controller could not cast the owner's spell during resolution")
	}
	original, _ := g.Stack.Peek()
	if !original.BuybackPaid || original.PaidCostSubjects[0].Snapshot.CardID != discarded {
		t.Fatal("paid control setup failed")
	}
	for _, redirect := range []struct {
		from, to zone.Type
		name     string
	}{
		{zone.Graveyard, zone.Exile, "Graveyard to exile"},
		{zone.Hand, zone.Graveyard, "Hand to graveyard"},
		{zone.Hand, zone.Library, "Hand to library"},
	} {
		g.ReplacementEffects = append(g.ReplacementEffects, game.ReplacementEffect{
			ID: g.IDGen.Next(), Description: redirect.name, MatchEvent: game.EventZoneChanged,
			MatchFromZone: true, FromZone: zone.Stack,
			MatchToZone: true, ToZone: redirect.from, ReplaceToZone: redirect.to,
		})
	}
	for _, path := range []string{"library", "cycle", "exile"} {
		t.Run(path, func(t *testing.T) {
			branch := g.Clone()
			cloned, _ := branch.Stack.Peek()
			cloned.PaidCostSubjects[0].Key = path
			choices := 0
			agents := independentBuybackAgent(t, func(request game.ChoiceRequest) string {
				if request.Kind != game.ChoiceReplacement {
					return ""
				}
				if request.Player != game.Player1 {
					t.Fatal("replacement chooser must be spell controller, not owner")
				}
				choices++
				if choices == 1 {
					if path == "exile" {
						return "Graveyard to exile"
					}
					return "Buyback"
				}
				if path == "cycle" {
					return "Hand to graveyard"
				}
				return "Hand to library"
			})
			engine.resolveTopOfStackWithChoices(branch, agents, &TurnLog{})
			owner := branch.Players[game.Player4]
			if owner.Library.Contains(cardID) != (path == "library") ||
				owner.Exile.Contains(cardID) != (path != "library") ||
				owner.Hand.Contains(cardID) || branch.Players[game.Player1].Hand.Contains(cardID) {
				t.Fatal("replacement chain reached the wrong owner's destination")
			}
			wantChoices := 2
			if path == "exile" {
				wantChoices = 1
			}
			if choices != wantChoices || len(branch.ReplacementEffects) != 3 ||
				branch.CardInstances[cardID].ZoneVersion != 2 ||
				original.PaidCostSubjects[0].Key != "clone-discard" || g.Stack.Size() != 2 ||
				g.CardInstances[cardID].ZoneVersion != 1 {
				t.Fatal("replacement re-evaluation, local lifetime, or clone isolation failed")
			}
		})
	}
}

func TestBuybackIndependentAdditionalChoiceAlternative(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := buybackEngine(t)
	subject := castBuybackSubject(t, g, engine)
	def := reiterateFixture()
	def.AlternativeCosts = []cost.Alternative{{
		Label: "Reduced alternative", ManaCost: opt.Val(cost.Mana{cost.R}),
	}}
	def.AdditionalCostChoices = []cost.AdditionalChoice{{Options: []cost.AdditionalChoiceOption{
		{Label: "Pay 2 life", Costs: []cost.Additional{{Kind: cost.AdditionalPayLife, Amount: 2}}},
		{Label: "Pay 2 mana", Mana: cost.Mana{cost.O(2)}},
	}}}
	cardID := addCardToHand(g, game.Player1, def)
	fundBuyback(g, 6)
	agents := independentBuybackAgent(t, func(request game.ChoiceRequest) string {
		if request.Kind == game.ChoicePayment && request.Prompt == "Choose spell cost" {
			if len(request.Options) != 8 {
				t.Fatalf("expected full normal/alternative/additional/buyback product: %+v", request)
			}
			return "Reduced alternative + Pay 2 life + Buyback"
		}
		return ""
	})
	if !engine.applyActionWithChoices(g, game.Player1, action.CastSpell(cardID,
		[]game.Target{game.StackObjectTarget(subject.ID)}, 0, nil), agents, &TurnLog{}) {
		t.Fatal("combined additional-choice/alternative/buyback cast failed")
	}
	obj, _ := g.Stack.Peek()
	if !obj.BuybackPaid || obj.ManaSpentToCast != 4 ||
		g.Players[game.Player1].Life != 38 || g.Players[game.Player1].ManaPool.Total() != 4 {
		t.Fatal("additional choice displaced the alternative or buyback costs")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if !g.Players[game.Player1].Hand.Contains(cardID) {
		t.Fatal("combined payment lost buyback replacement")
	}
}

func TestBuybackIndependentCostChoiceGroupsStayIndependent(t *testing.T) {
	for _, buybackGroup := range []uint8{1, 2} {
		t.Run(map[uint8]string{1: "component-local-group-ids", 2: "distinct-group-control"}[buybackGroup], func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := buybackEngine(t)
			subject := castBuybackSubject(t, g, engine)
			def := reiterateFixture()
			def.AdditionalCosts = []cost.Additional{
				{Kind: cost.AdditionalPayLife, Amount: 2, ChoiceGroup: 1},
				{Kind: cost.AdditionalSacrifice, Amount: 1, ChoiceGroup: 1,
					MatchPermanentType: true, PermanentType: types.Artifact},
			}
			def.StaticAbilities[0].KeywordAbilities = []game.KeywordAbility{game.BuybackKeyword{
				Cost: cost.Mana{cost.O(0)},
				AdditionalCosts: []cost.Additional{
					{Kind: cost.AdditionalPayLife, Amount: 3, ChoiceGroup: buybackGroup},
					{Kind: cost.AdditionalDiscard, Amount: 1, ChoiceGroup: buybackGroup},
				},
			}}
			if issues := game.ValidateCardDef(def); len(issues) != 0 {
				t.Fatalf("independent non-mana cost components rejected: %+v", issues)
			}
			cardID := addCardToHand(g, game.Player1, def)
			fundBuyback(g, 1)
			if !engine.applyActionWithChoices(g, game.Player1, action.CastSpell(cardID,
				[]game.Target{game.StackObjectTarget(subject.ID)}, 0, nil),
				independentBuybackAgent(t, independentChooseBuyback), &TurnLog{}) {
				t.Fatal("payable combined non-mana buyback cast failed")
			}
			obj, _ := g.Stack.Peek()
			if !obj.BuybackPaid || obj.ManaSpentToCast != 3 {
				t.Fatal("paid announcement setup failed")
			}
			// No artifact or other hand card exists: both independent cost
			// components must be paid in life, for five life in total.
			if g.Players[game.Player1].Life != 35 || len(obj.AdditionalCostsPaid) != 2 {
				t.Errorf("buyback collapsed with mandatory cost: life=%d paid=%v; want life=35 and both costs",
					g.Players[game.Player1].Life, obj.AdditionalCostsPaid)
			}
			engine.resolveTopOfStack(g, &TurnLog{})
			if !g.Players[game.Player1].Hand.Contains(cardID) {
				t.Fatal("paid spell did not return via buyback")
			}
		})
	}
}

func TestBuybackIndependentPartialTargetsAndNormalX(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	a := addCombatPermanent(g, game.Player2, vanillaCreature("Illegal target", 3, 3))
	b := addCombatPermanent(g, game.Player2, vanillaCreature("Legal target", 3, 3))
	def := &game.CardDef{CardFace: game.CardFace{
		Name: "Independent partial-target buyback probe", Types: []types.Card{types.Instant},
		ManaCost: opt.Val(cost.Mana{cost.X}),
		StaticAbilities: []game.StaticAbility{{ZoneOfFunction: zone.Stack,
			KeywordAbilities: []game.KeywordAbility{game.BuybackKeyword{
				Cost:            cost.Mana{cost.O(1)},
				AdditionalCosts: []cost.Additional{{Kind: cost.AdditionalPayLife, Amount: 1}},
			}},
		}},
		SpellAbility: opt.Val(game.Mode{
			Targets: []game.TargetSpec{{
				MinTargets: 2, MaxTargets: 2, Allow: game.TargetAllowPermanent,
				Selection: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
			}},
			Sequence: []game.Instruction{
				{Primitive: game.Damage{Amount: game.Fixed(1), Recipient: game.AnyTargetDamageRecipient(0)}},
				{Primitive: game.Damage{Amount: game.Fixed(1), Recipient: game.AnyTargetDamageRecipient(1)}},
			},
		}.Ability()),
	}}
	cardID := addCardToHand(g, game.Player1, def)
	setMainPhasePriority(g, game.Player1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 3)
	if !engine.applyActionWithChoices(g, game.Player1, action.CastSpell(cardID,
		[]game.Target{game.PermanentTarget(a.ObjectID), game.PermanentTarget(b.ObjectID)}, 2, nil),
		independentBuybackAgent(t, independentChooseBuyback), &TurnLog{}) {
		t.Fatal("otherwise legal normal-X spell with fixed buyback failed")
	}
	movePermanentToZone(g, a, zone.Graveyard)
	engine.resolveTopOfStack(g, &TurnLog{})
	if b.MarkedDamage != 1 || !g.Players[game.Player1].Hand.Contains(cardID) ||
		g.Players[game.Player1].Life != 39 || g.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatal("partially legal spell must resolve and return after paying fixed buyback")
	}
}
