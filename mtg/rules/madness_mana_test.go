package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review: Comprehensive Rules effective 2026-09-25 (media.wizards.com
// MagicCompRules 20260925). CR 702.35a-b: the madness trigger lets the owner
// cast the exiled card by paying the madness cost rather than its mana cost,
// following the alternative-cost rules 601.2b and 601.2f-h; otherwise the card
// goes to the graveyard. CR 601.2f locks the total cost; CR 601.2g then lets
// the caster activate mana abilities if that total includes a mana payment;
// CR 733.1 reverses an unpaid cast and lets the player reverse its legal mana
// abilities. Oracle (pinned corpus): Arrogant Wurm dc7f0c33-80ea-453e-a237-
// 473f077f038e "Madness {2}{G}"; Waterlogged Grove 70fa2eba-565e-4fed-adc9-
// 7f5d9fcbf1fa "{T}, Pay 1 life: Add {G} or {U}."; Ashnod's Altar 4d18bcba-
// a346-445e-a182-6cc30b7e066d "Sacrifice a creature: Add {C}{C}."; Thalia,
// Guardian of Thraben 9b7f1d05-707c-4ed3-9f0e-8ced1232c2ee "Noncreature spells
// cost {1} more to cast." These scenarios cover Engine gap #3463.

// discardIntoMadness discards cardID from Player1's hand and puts the madness
// trigger on the stack.
func discardIntoMadness(t *testing.T, e *Engine, g *game.Game, cardID id.ID) {
	t.Helper()
	if !discardCardFromHand(g, game.Player1, cardID) || !g.Players[game.Player1].Exile.Contains(cardID) {
		t.Fatal("madness card was not discarded into exile")
	}
	if !e.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("madness trigger was not put on the stack")
	}
}

func requireMadnessCast(t *testing.T, g *game.Game, cardID id.ID) (*game.StackObject, game.Event) {
	t.Helper()
	obj, ok := g.Stack.Peek()
	if !ok || obj.Kind != game.StackSpell || obj.SourceID != cardID || g.Players[game.Player1].Exile.Contains(cardID) {
		t.Fatalf("stack top = %+v, want the madness spell cast from exile", obj)
	}
	for _, event := range g.Events {
		if event.Kind == game.EventSpellCast && event.CardID == cardID && event.FromZone == zone.Exile {
			return obj, event
		}
	}
	t.Fatal("no spell-cast event for the madness spell")
	return nil, game.Event{}
}

func requireMadnessNotCast(t *testing.T, g *game.Game, cardID id.ID) {
	t.Helper()
	player := g.Players[game.Player1]
	if g.Stack.Size() != 0 || player.Exile.Contains(cardID) || !player.Graveyard.Contains(cardID) {
		t.Fatalf("stack size=%d in exile=%t in graveyard=%t, want the uncast madness card in the graveyard",
			g.Stack.Size(), player.Exile.Contains(cardID), player.Graveyard.Contains(cardID))
	}
	for _, event := range g.Events {
		if event.Kind == game.EventSpellCast && event.CardID == cardID {
			t.Fatal("an uncast madness card produced a spell-cast event")
		}
	}
}

// TestMadnessCastPaysWithPaymentTimeManaAbility casts the existing card Arrogant
// Wurm for madness {2}{G} with two colorless mana and Waterlogged Grove, whose
// life-paying mana ability is never activated automatically, as the only green
// source. The cast is proposed and the Grove pays {G} while paying (CR 601.2g).
// The cast runs in a clone without changing the original game. When the caster
// cannot decline the required Grove payment after announcing the cast
// (CR 601.2h, 2026-11-06).
func TestMadnessCastPaysWithPaymentTimeManaAbility(t *testing.T) {
	reg := cards.NewDefaultRegistry()
	e, g := newEffectiveCascadeGame()
	cardID := addCardToHand(g, game.Player1, reg.Lookup("Arrogant Wurm"))
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)
	addCombatPermanent(g, game.Player1, reg.Lookup("Waterlogged Grove"))
	discardIntoMadness(t, e, g, cardID)
	life := g.Players[game.Player1].Life

	t.Run("grove activated", func(t *testing.T) {
		clone := g.Clone()
		agent := &paymentManaAgent{ability: "Pay 1 life"}
		e.resolveTopOfStackWithChoices(clone, reboundKickerAgents(agent), &TurnLog{})

		obj, event := requireMadnessCast(t, clone, cardID)
		grove := reboundPermanentNamed(t, clone, "Waterlogged Grove")
		if agent.windows != 1 || agent.activated != 1 || !grove.Tapped || clone.Players[game.Player1].Life != life-1 {
			t.Fatalf("windows=%d activated=%d grove tapped=%t life=%d, want the Grove to pay {G}",
				agent.windows, agent.activated, grove.Tapped, clone.Players[game.Player1].Life)
		}
		if obj.ManaSpentToCast != 3 || event.ManaSpentToCast.Val != 3 || clone.Players[game.Player1].ManaPool.Total() != 0 {
			t.Fatalf("mana spent obj=%d event=%+v pool=%d, want the {2}{G} madness cost paid",
				obj.ManaSpentToCast, event.ManaSpentToCast, clone.Players[game.Player1].ManaPool.Total())
		}
		if !g.Players[game.Player1].Exile.Contains(cardID) || reboundPermanentNamed(t, g, "Waterlogged Grove").Tapped ||
			g.Players[game.Player1].ManaPool.Total() != 2 || g.Stack.Size() != 1 {
			t.Fatal("casting the madness spell in the clone changed the original game")
		}
	})

	t.Run("required grove cannot be declined", func(t *testing.T) {
		clone := g.Clone()
		agent := &paymentManaAgent{ability: "no such ability"}
		e.resolveTopOfStackWithChoices(clone, reboundKickerAgents(agent), &TurnLog{})

		requireMadnessCast(t, clone, cardID)
		if agent.windows != 1 || !reboundPermanentNamed(t, clone, "Waterlogged Grove").Tapped ||
			clone.Players[game.Player1].Life != life-1 || clone.Players[game.Player1].ManaPool.Total() != 0 {
			t.Fatalf("windows=%d life=%d pool=%d, want the required payment completed",
				agent.windows, clone.Players[game.Player1].Life, clone.Players[game.Player1].ManaPool.Total())
		}
	})
}

// TestMadnessCastLocksCostBeforeManaAbilities proves the madness total cost is
// locked before payment-time mana abilities (CR 601.2f-g): sacrificing the
// caster's own Thalia to Ashnod's Altar while paying does not remove the {1}
// tax already included in the total.
func TestMadnessCastLocksCostBeforeManaAbilities(t *testing.T) {
	reg := cards.NewDefaultRegistry()
	e, g := newEffectiveCascadeGame()
	cardID := addCardToHand(g, game.Player1, madnessSorcery(cost.Mana{cost.G}))
	forest := addBasicLandPermanent(g, game.Player1, types.Forest)
	thalia := addCombatPermanent(g, game.Player1, reg.Lookup("Thalia, Guardian of Thraben"))
	addCombatPermanent(g, game.Player1, reg.Lookup("Ashnod's Altar"))
	discardIntoMadness(t, e, g, cardID)
	agent := &paymentManaAgent{ability: "Sacrifice a creature"}

	e.resolveTopOfStackWithChoices(g, reboundKickerAgents(agent), &TurnLog{})

	obj, _ := requireMadnessCast(t, g, cardID)
	if _, onBattlefield := permanentByObjectID(g, thalia.ObjectID); agent.activated != 1 || onBattlefield {
		t.Fatalf("activated=%d thalia on battlefield=%t, want Thalia sacrificed for mana", agent.activated, onBattlefield)
	}
	if !forest.Tapped || obj.ManaSpentToCast != 2 || g.Players[game.Player1].ManaPool.Total() != 1 {
		t.Fatalf("forest tapped=%t mana spent=%d pool=%d, want the locked {1}{G} paid and {C} left",
			forest.Tapped, obj.ManaSpentToCast, g.Players[game.Player1].ManaPool.Total())
	}
}

// Impossible sacrifice/mana totals are not offered. With both lands the total
// is completable, so neither required payment can be declined into an unpaid
// cast (601.2h, 702.35a; 2026-11-06).
func TestMadnessCastOffersOnlyCompletablePayments(t *testing.T) {
	reg := cards.NewDefaultRegistry()
	for _, tc := range []struct {
		name          string
		artifact      bool
		thalias       int
		peatland      bool
		wantWindows   int
		wantActivated int
		wantCast      bool
	}{
		{name: "all costs paid", artifact: true, thalias: 1, wantWindows: 1, wantActivated: 1, wantCast: true},
		{name: "sacrifice impossible", artifact: false, thalias: 1},
		{name: "mana short", artifact: true, thalias: 2},
		{name: "both sources required", artifact: true, thalias: 2, peatland: true, wantWindows: 2, wantActivated: 1, wantCast: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, g := newEffectiveCascadeGame()
			def := madnessSorcery(cost.Mana{cost.G})
			def.AdditionalCosts = []cost.Additional{{
				Kind:               cost.AdditionalSacrifice,
				Text:               "Sacrifice an artifact",
				Amount:             1,
				MatchPermanentType: true,
				PermanentType:      types.Artifact,
			}}
			cardID := addCardToHand(g, game.Player1, def)
			g.Players[game.Player1].ManaPool.Add(mana.G, 1)
			var fodder *game.Permanent
			if tc.artifact {
				fodder = reboundTestArtifact(g)
			}
			for range tc.thalias {
				addCombatPermanent(g, game.Player2, reg.Lookup("Thalia, Guardian of Thraben"))
			}
			grove := addCombatPermanent(g, game.Player1, reg.Lookup("Waterlogged Grove"))
			var peatland *game.Permanent
			if tc.peatland {
				peatland = addCombatPermanent(g, game.Player1, reg.Lookup("Nurturing Peatland"))
			}
			discardIntoMadness(t, e, g, cardID)
			life := g.Players[game.Player1].Life
			agent := &paymentManaAgent{ability: "Add {G} or {U}"}

			e.resolveTopOfStackWithChoices(g, reboundKickerAgents(agent), &TurnLog{})

			if agent.windows != tc.wantWindows || agent.activated != tc.wantActivated {
				t.Fatalf("payment windows=%d activations=%d, want %d/%d", agent.windows, agent.activated, tc.wantWindows, tc.wantActivated)
			}
			if !tc.wantCast {
				requireMadnessNotCast(t, g, cardID)
				if grove.Tapped || g.Players[game.Player1].Life != life || g.Players[game.Player1].ManaPool.Total() != 1 {
					t.Fatalf("grove tapped=%t life=%d pool=%d, want no mana activation or payment left behind",
						grove.Tapped, g.Players[game.Player1].Life, g.Players[game.Player1].ManaPool.Total())
				}
				if peatland != nil && peatland.Tapped {
					t.Fatal("the Peatland was activated although the caster declined it")
				}
				if fodder != nil {
					if _, ok := permanentByObjectID(g, fodder.ObjectID); !ok {
						t.Fatal("the reversed proposal sacrificed the artifact")
					}
				}
				return
			}
			obj, _ := requireMadnessCast(t, g, cardID)
			if len(obj.SacrificedAsCostIDs) != 1 || !g.Players[game.Player1].Graveyard.Contains(fodder.CardInstanceID) {
				t.Fatalf("sacrificed as cost = %v, want the artifact sacrificed", obj.SacrificedAsCostIDs)
			}
			lifePaid, manaPaid := 1, 2
			if peatland != nil {
				lifePaid, manaPaid = 2, 3
			}
			if !grove.Tapped || g.Players[game.Player1].Life != life-lifePaid || obj.ManaSpentToCast != manaPaid ||
				peatland != nil && !peatland.Tapped {
				t.Fatalf("grove tapped=%t life=%d mana spent=%d, want the Grove to pay Thalia's tax",
					grove.Tapped, g.Players[game.Player1].Life, obj.ManaSpentToCast)
			}
		})
	}
}

// TestMadnessCastWithoutManaPaymentOpensNoManaWindow casts for a madness cost of
// {0}: the locked total includes no mana payment, so no payment-time mana
// ability window opens (CR 601.2g) although Waterlogged Grove is available.
func TestMadnessCastWithoutManaPaymentOpensNoManaWindow(t *testing.T) {
	reg := cards.NewDefaultRegistry()
	e, g := newEffectiveCascadeGame()
	cardID := addCardToHand(g, game.Player1, madnessSorcery(cost.Mana{cost.O(0)}))
	grove := addCombatPermanent(g, game.Player1, reg.Lookup("Waterlogged Grove"))
	discardIntoMadness(t, e, g, cardID)
	agent := &paymentManaAgent{ability: "Pay 1 life"}

	e.resolveTopOfStackWithChoices(g, reboundKickerAgents(agent), &TurnLog{})

	obj, _ := requireMadnessCast(t, g, cardID)
	if agent.windows != 0 || grove.Tapped || obj.ManaSpentToCast != 0 {
		t.Fatalf("windows=%d grove tapped=%t mana spent=%d, want a {0} madness cast without a mana window",
			agent.windows, grove.Tapped, obj.ManaSpentToCast)
	}
}

func TestMadnessAnnouncesAlternativeCostXOnStack(t *testing.T) {
	e, g := newEffectiveCascadeGame()
	def := numericDamageSpell()
	def.ManaCost = opt.Val(cost.Mana{cost.O(1), cost.R})
	def.StaticAbilities = append(def.StaticAbilities, game.StaticAbility{KeywordAbilities: []game.KeywordAbility{
		game.MadnessKeyword{Cost: cost.Mana{cost.X, cost.R}},
	}})
	card := addCardToHand(g, game.Player1, def)
	discardIntoMadness(t, e, g, card)
	g.Players[game.Player1].ManaPool.Add(mana.C, 21)
	g.Players[game.Player1].ManaPool.Add(mana.R, 1)
	var announced *game.StackObject
	numbers := 0
	agent := &controlScript{choice: func(_ PlayerObservation, request game.ChoiceRequest) []int {
		if request.Kind == game.ChoiceMay {
			return []int{1}
		}
		if request.Kind == game.ChoiceNumber || request.Kind == game.ChoiceTarget {
			top, ok := g.Stack.Peek()
			if !ok || top.SourceID != card || !top.Casting || g.Players[game.Player1].Exile.Contains(card) ||
				announced != nil && announced != top {
				t.Fatal("madness announcements did not observe the same Casting object")
			}
			announced = top
		}
		if request.Kind == game.ChoiceNumber {
			numbers++
			return []int{21}
		}
		if request.Kind == game.ChoiceTarget {
			for _, option := range request.Options {
				if len(option.Targets) == 1 && option.Targets[0] == game.PlayerTarget(game.Player2) {
					return []int{option.Index}
				}
			}
		}
		return request.DefaultSelection
	}}
	e.resolveTopOfStackWithChoices(g, reboundKickerAgents(agent), &TurnLog{})
	object, _ := requireMadnessCast(t, g, card)
	if object != announced || object.Casting || object.XValue != 21 || object.ManaSpentToCast != 22 || numbers != 1 {
		t.Fatal("madness used the printed cost instead of its announced alternative X")
	}
	e.resolveTopOfStack(g, &TurnLog{})
	if g.Players[game.Player2].Life != 19 {
		t.Fatal("madness did not resolve with the exact announced X")
	}
}
