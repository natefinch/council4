package rules

import (
	"fmt"
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/mtg/rules/payment"
	"github.com/natefinch/council4/opt"
)

func TestPaymentCompletionRegisteredSixMasticoreDiscardAssignments(t *testing.T) {
	for _, tc := range []struct {
		landFirst bool
		hasLand   bool
	}{{true, true}, {false, true}, {false, false}} {
		t.Run(fmt.Sprintf("land-first=%v/land=%v", tc.landFirst, tc.hasLand), func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			addPermanentForSBA(g, game.Player1, registryCardDef(t, "Six"))
			def := registryCardDef(t, "Lesser Masticore")
			card := addCardToGraveyard(g, game.Player1, def)
			var land id.ID
			if tc.hasLand && tc.landFirst {
				land = addCardToHand(g, game.Player1, registryCardDef(t, "Forest"))
			}
			nonland := addCardToHand(g, game.Player1, registryCardDef(t, "Cathartic Reunion"))
			if tc.hasLand && !tc.landFirst {
				land = addCardToHand(g, game.Player1, registryCardDef(t, "Forest"))
			}
			if !tc.hasLand {
				addCardToHand(g, game.Player1, discardableCard("Second nonland"))
			}
			g.Players[game.Player1].ManaPool.Add(mana.C, 2)
			setMainPhasePriority(g, game.Player1)
			planner := paymentOrch.planner(g)
			req := payment.ForSpell(payment.SpellRequest{PlayerID: game.Player1, CardID: card, Card: def, SourceZone: zone.Graveyard, CastPermissions: []payment.SpellCastPermission{payment.SpellCastPermissionEscape}})
			partial := payment.Partial{Option: opt.Val(1)} // Six's granted retrace.
			plan, ok, _ := planner.Complete(req, partial)
			if ok != tc.hasLand {
				t.Fatalf("joint retrace completion = %v, want %v", ok, tc.hasLand)
			}
			if tc.hasLand {
				if !slices.Equal(plan.Components()[0].Objects, []id.ID{nonland}) ||
					!slices.Equal(plan.Components()[1].Objects, []id.ID{land}) {
					t.Fatalf("joint discard assignment = %v, want nonland then land", plan.Components())
				}
				options := planner.Options(req, partial, payment.Step{Kind: payment.StepComponent, Index: 0})
				if len(options) != 1 || !slices.Equal(options[0].Component.Objects, []id.ID{nonland}) {
					t.Fatalf("unrestricted discard choices = %+v, want only nonland", options)
				}
			}

			engine := NewEngine(nil)
			act := action.CastSpellFromZone(card, zone.Graveyard, nil, 0, nil)
			if engine.applyAction(g, game.Player1, act) != tc.hasLand {
				t.Fatal("registered retrace casting disagreed with the complete payment")
			}
			if tc.hasLand && (!g.Players[game.Player1].Graveyard.Contains(nonland) || !g.Players[game.Player1].Graveyard.Contains(land)) {
				t.Fatal("the cast did not discard both selected cards")
			}
		})
	}
}

func TestPaymentCompletionRegisteredMillSourcesRequireLibraryCard(t *testing.T) {
	for _, name := range []string{"Millikin", "Deranged Assistant"} {
		for _, fuel := range []string{"empty", "ordinary", "replacement"} {
			t.Run(name+"/"+fuel, func(t *testing.T) {
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				e := NewEngine(nil)
				def := registryCardDef(t, name)
				source := addCombatPermanent(g, game.Player1, def)
				source.SummoningSick = false
				setMainPhasePriority(g, game.Player1)
				var card id.ID
				if fuel == "ordinary" {
					card = addCardToLibrary(g, game.Player1, discardableCard("Mill fuel"))
				} else if fuel == "replacement" {
					card = addCardToLibrary(g, game.Player1, registryCardDef(t, "Blightsteel Colossus"))
				}
				act := action.ActivateAbility(source.ObjectID, def.ActivatedAbilityIndex(0), nil, 0)
				want := fuel != "empty"
				if containsAction(e.legalActions(g, game.Player1), act) != want || e.applyAction(g, game.Player1, act) != want {
					t.Fatal("registered mill-cost activation did not enforce library affordability")
				}
				if !want {
					if source.Tapped || !g.Stack.IsEmpty() {
						t.Fatal("rejected mill-cost activation mutated the source or stack")
					}
					return
				}
				if !source.Tapped || g.Stack.IsEmpty() {
					t.Fatal("affordable library-moving ability was not activated on the stack")
				}
				e.resolveTopOfStack(g, &TurnLog{})
				if g.Players[game.Player1].ManaPool.Total() != 1 {
					t.Fatal("the registered source did not add its mana on resolution")
				}
				if fuel == "replacement" && !g.Players[game.Player1].Library.Contains(card) ||
					fuel == "ordinary" && !g.Players[game.Player1].Graveyard.Contains(card) {
					t.Fatal("mill payment bypassed the applicable destination behavior")
				}
			})
		}
	}
}

func TestPaymentCompletionMultiplePhyrexianSymbolsShareLifeBudget(t *testing.T) {
	for _, life := range []int{1, 3, 4} {
		t.Run(fmt.Sprint(life), func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			g.Players[game.Player1].Life = life
			g.Players[game.Player1].ManaPool.Add(mana.W, 2)
			manaCost := cost.Mana{cost.O(2), cost.PhyrexianMana(mana.W), cost.PhyrexianMana(mana.U)}
			req := payment.ForGeneric(payment.GenericRequest{PlayerID: game.Player1, Cost: &manaCost})
			plan, ok, _ := paymentOrch.planner(g).Complete(req, payment.Partial{})
			if ok != (life == 4) {
				t.Fatalf("complete two-Phyrexian payment with %d life = %v", life, ok)
			}
			if !ok {
				return
			}
			options := paymentOrch.planner(g).Options(req, payment.Partial{}, payment.Step{Kind: payment.StepPhyrexian, Index: 0})
			if len(options) != 1 || !options[0].PayLife {
				t.Fatalf("Phyrexian choices = %+v, want only life", options)
			}
			requireApplyPayment(t, g, plan)
			if g.Players[game.Player1].Life != 0 || g.Players[game.Player1].ManaPool.Total() != 0 {
				t.Fatal("the complete plan did not pay its exact life/mana allocation")
			}
		})
	}
}

func TestPaymentCompletionInvalidPartialNeverSilentlyDefaults(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	manaCost := cost.Mana{cost.O(1)}
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	req := payment.ForGeneric(payment.GenericRequest{PlayerID: game.Player1, Cost: &manaCost})
	for _, partial := range []payment.Partial{
		{Option: opt.Val(99)}, {Phyrexian: map[int]bool{0: true}},
		{Groups: map[uint8]int{1: 0}}, {Components: map[int]payment.ComponentChoice{0: {Objects: []id.ID{999}}}},
	} {
		if _, ok, reason := paymentOrch.planner(g).Complete(req, partial); ok || reason == "" {
			t.Fatal("invalid partial choices silently became a payable default")
		}
	}
}

func TestPaymentCompletionJointPowerTapKeepsLaterReturnPayable(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Power source"}})
	first := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Artifact creature", Types: []types.Card{types.Creature, types.Artifact}, Power: opt.Val(game.PT{Value: 2}), Toughness: opt.Val(game.PT{Value: 2})}})
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Other creature", Types: []types.Card{types.Creature}, Power: opt.Val(game.PT{Value: 1}), Toughness: opt.Val(game.PT{Value: 1})}})
	req := payment.ForAbility(payment.AbilityRequest{PlayerID: game.Player1, Source: source, AdditionalCosts: []cost.Additional{
		{Kind: cost.AdditionalTapPermanents, TotalPowerAtLeast: 3, MatchPermanentType: true, PermanentType: types.Creature},
		{Kind: cost.AdditionalReturnToHand, MatchPermanentType: true, PermanentType: types.Artifact},
	}})
	plan := requireCompletePayment(t, g, req, payment.Partial{})
	requireApplyPayment(t, g, plan)
	if !first.Tapped || !g.Players[game.Player1].Hand.Contains(first.CardInstanceID) {
		t.Fatal("power tap and subsequent return were not planned and paid jointly")
	}
}

func TestPaymentCompletionThresholdAppliesEverySelectedObject(t *testing.T) {
	for _, kind := range []cost.AdditionalKind{cost.AdditionalTapPermanents, cost.AdditionalExile, cost.AdditionalCollectEvidence} {
		t.Run(fmt.Sprint(kind), func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Threshold source"}})
			additional := cost.Additional{Kind: kind, Amount: 2}
			var selected []id.ID
			for i := range 3 {
				def := &game.CardDef{CardFace: game.CardFace{Name: fmt.Sprint("Threshold fuel", i), ManaCost: opt.Val(cost.Mana{cost.O(1)}),
					Types: []types.Card{types.Creature}, Power: opt.Val(game.PT{Value: 1}), Toughness: opt.Val(game.PT{Value: 1})}}
				if kind == cost.AdditionalTapPermanents {
					selected = append(selected, addCombatPermanent(g, game.Player1, def).ObjectID)
				} else {
					selected = append(selected, addCardToGraveyard(g, game.Player1, def))
				}
			}
			if kind == cost.AdditionalTapPermanents {
				additional.TotalPowerAtLeast = 2
				additional.MatchPermanentType, additional.PermanentType = true, types.Creature
			} else if kind == cost.AdditionalExile {
				additional.TotalManaValueAtLeast = 2
			}
			req := payment.ForAbility(payment.AbilityRequest{PlayerID: game.Player1, Source: source, AdditionalCosts: []cost.Additional{additional}})
			partial := payment.Partial{Components: map[int]payment.ComponentChoice{0: {Objects: selected, Finished: true}}}
			plan := requireCompletePayment(t, g, req, partial)
			result := requireApplyPayment(t, g, plan)
			if kind == cost.AdditionalTapPermanents {
				if !slices.Equal(result.Ability.TappedIDs, selected) {
					t.Fatalf("paid taps = %v, want every selected object %v", result.Ability.TappedIDs, selected)
				}
			} else if g.Players[game.Player1].Exile.Size() != 3 || g.Players[game.Player1].Graveyard.Size() != 0 {
				t.Fatal("a complete threshold plan silently omitted selected cards")
			}
		})
	}
}

func TestPaymentCompletionRejectsCardAfterZoneRoundTrip(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	card := addCardToHand(g, game.Player1, discardableCard("New object after returning"))
	req := payment.ForGeneric(payment.GenericRequest{PlayerID: game.Player1, AdditionalCosts: []cost.Additional{{Kind: cost.AdditionalDiscard}}})
	plan := requireCompletePayment(t, g, req, payment.Partial{})
	state := &rulesPaymentState{g: g}
	if !state.MoveCard(game.Player1, card, zone.Hand, zone.Graveyard) ||
		!state.MoveCard(game.Player1, card, zone.Graveyard, zone.Hand) {
		t.Fatal("zone round-trip setup failed")
	}
	if _, ok, reason := paymentOrch.planner(g).Apply(plan); ok || reason == "" {
		t.Fatal("the old witness followed a selected card to its new object")
	}
	if !g.Players[game.Player1].Hand.Contains(card) {
		t.Fatal("stale witness partially paid its discard cost")
	}
}
