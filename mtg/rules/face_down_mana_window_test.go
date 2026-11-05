package rules

import (
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// Source review: issue #3479. Default-registry cards checked against the
// pinned corpus: War Behemoth 4792a05e-b557-4132-8303-235feefe0d99 "Morph
// {4}{W} (You may cast this card face down as a 2/2 creature for {3}. ...)";
// Alley Assailant (Disguise {4}{B}{B}, see morph_cost_test.go); Waterlogged
// Grove 70fa2eba-565e-4fed-adc9-7f5d9fcbf1fa "{T}, Pay 1 life: Add {G} or
// {U}."; Ashnod's Altar 4d18bcba-a346-445e-a182-6cc30b7e066d "Sacrifice a
// creature: Add {C}{C}."; Lodestone Golem a698c9f5-325a-49f6-8902-eef09f4260c5
// "Nonartifact spells cost {1} more to cast."; Heartless Summoning
// fa1be67e-a07e-42ce-88be-446acd643dc6 "Creature spells you cast cost {2} less
// to cast."; Grizzly Bears 14c8f55d-d177-4c25-a931-ebeb9e6062a0 (vanilla 2/2
// creature); Muldrotha, the Gravetide e4625704-1d52-44e4-804f-2f45644d76ac
// (cast a permanent spell of each permanent type from your graveyard); Millikin
// fa4dffda-6f04-4d0b-829d-28a1a5794dee "{T}, Mill a card: Add {C}. (Activate
// only as an instant. ...)", not a mana ability (605.1a, ruling 2026-08-05);
// Deathbloom Ritualist
// 45ff816d-1eb7-4985-90ec-f44802a696fc "{T}: Add X mana of any one color,
// where X is the number of creature cards in your graveyard."; Forest
// 2e781bfd-f8cb-4d20-b977-4359f3a21311.
// Comprehensive Rules effective 2026-09-25: 702.37c and 702.168b turn the card
// face down as a 2/2 with no name, subtypes or mana cost and put it onto the
// stack, paying {3} following the alternative-cost rules; 601.2a moves the
// card to the stack first; 601.2f locks the total cost including increases
// and reductions; 601.2g "If the total cost includes a mana payment, the
// player then has a chance to activate mana abilities ... before costs are
// paid"; 601.2h pays it; 404.2 says players "normally can't change" a
// graveyard's order; 733.1 reverses a cast that can't be completed, returning
// the spell to the zone it came from and letting the player reverse legal mana
// abilities activated while making it; 733.2 lets the player keep priority
// and redo or take another action. 733.1 also says "Players may not reverse
// actions that moved cards ... from a library to any zone other than the
// stack". 400.2: "Library and hand are hidden zones"; 401.2: "Players can't
// look at or change the order of cards in a library."

// faceDownWindowAgent casts cardID face down whenever offered, and otherwise
// acts as its paymentManaAgent while paying. It records whether any choice it
// was asked while casting named the hidden card.
type faceDownWindowAgent struct {
	paymentManaAgent
	cardID      id.ID
	casts       int
	namedHidden bool
}

func (a *faceDownWindowAgent) ChooseAction(_ PlayerObservation, legal []action.Action) action.Action {
	for _, act := range legal {
		if payload, ok := act.CastFaceDownPayload(); ok && payload.CardID == a.cardID {
			a.casts++
			return act
		}
	}
	return action.Pass()
}

func (a *faceDownWindowAgent) ChooseChoice(obs PlayerObservation, request game.ChoiceRequest) []int {
	for _, option := range request.Options {
		if option.Card.Exists && option.Card.Val.CardID == a.cardID {
			a.namedHidden = true
		}
	}
	return a.paymentManaAgent.ChooseChoice(obs, request)
}

func requireFaceDownSpell(t *testing.T, g *game.Game, cardID id.ID, kind game.FaceDownKind, from zone.Type, spent int) {
	t.Helper()
	obj, ok := g.Stack.Peek()
	if !ok || obj.Kind != game.StackSpell || !obj.FaceDown || obj.SourceID != cardID || obj.FaceDownKind != kind || obj.SourceZone != from {
		t.Fatalf("stack top = %+v, want a face-down spell cast from %v", obj, from)
	}
	if obj.ManaSpentToCast != spent {
		t.Fatalf("ManaSpentToCast = %d, want %d", obj.ManaSpentToCast, spent)
	}
	assertEvent(t, g.Events, game.EventSpellCast, func(event game.Event) bool {
		return event.CardID == cardID && event.FromZone == from && event.ToZone == zone.Stack && event.FaceDown &&
			event.ManaSpentToCast.Exists && event.ManaSpentToCast.Val == spent &&
			event.ManaValue.Exists && event.ManaValue.Val == 0 && len(event.Colors) == 0 &&
			len(event.CardTypes) == 1 && event.CardTypes[0] == types.Creature
	})
}

func requireNoFaceDownCast(t *testing.T, g *game.Game, cardID id.ID) {
	t.Helper()
	if g.Stack.Size() != 0 {
		t.Fatalf("stack size = %d, want the reversed face-down spell removed", g.Stack.Size())
	}
	for _, event := range g.Events {
		if event.Kind == game.EventSpellCast && event.CardID == cardID {
			t.Fatal("a reversed face-down cast produced a spell-cast event")
		}
	}
}

// TestFaceDownHandCastPaysWithPaymentTimeManaAbility gives the caster two
// colorless mana and Waterlogged Grove, whose life-paying mana ability is never
// activated automatically. The {3} face-down cast of War Behemoth (morph) and
// Alley Assailant (disguise) is offered, the Grove pays the last {1} while
// paying (CR 601.2g), and the spell is a face-down 2/2 cast from hand. The
// cast runs in a clone without changing the original game. Declining the
// window reverses the proposal (CR 733.1): the card returns to hand with the
// mana and life unspent.
func TestFaceDownHandCastPaysWithPaymentTimeManaAbility(t *testing.T) {
	for _, tc := range []struct {
		card string
		kind game.FaceDownKind
	}{
		{card: "War Behemoth", kind: game.FaceDownMorph},
		{card: "Alley Assailant", kind: game.FaceDownDisguise},
	} {
		t.Run(tc.card, func(t *testing.T) {
			g, engine, cardID := faceDownCostGame(t, tc.card, mana.C, 2)
			addCombatPermanent(g, game.Player1, productionChoiceCard(t, "Waterlogged Grove"))
			life := g.Players[game.Player1].Life
			if !faceDownCastOffered(engine, g, cardID, tc.kind) {
				t.Fatal("face-down cast payable with Waterlogged Grove was not offered")
			}
			cast := actionBuild.castFaceDown(cardID, game.FaceFront, tc.kind)

			t.Run("grove activated", func(t *testing.T) {
				clone := g.Clone()
				agent := &faceDownWindowAgent{paymentManaAgent: paymentManaAgent{ability: "Pay 1 life"}, cardID: cardID}
				if !engine.applyActionWithChoices(clone, game.Player1, cast, reboundKickerAgents(agent), &TurnLog{}) {
					t.Fatal("face-down cast paid with Waterlogged Grove failed")
				}
				requireFaceDownSpell(t, clone, cardID, tc.kind, zone.Hand, 3)
				grove := reboundPermanentNamed(t, clone, "Waterlogged Grove")
				if agent.windows != 1 || agent.activated != 1 || !grove.Tapped || clone.Players[game.Player1].Life != life-1 ||
					clone.Players[game.Player1].ManaPool.Total() != 0 || clone.Players[game.Player1].Hand.Contains(cardID) {
					t.Fatalf("windows=%d activated=%d grove tapped=%t life=%d pool=%d, want the Grove to pay the last {1}",
						agent.windows, agent.activated, grove.Tapped, clone.Players[game.Player1].Life, clone.Players[game.Player1].ManaPool.Total())
				}
				if agent.namedHidden {
					t.Fatal("a payment choice named the hidden face-down card")
				}
				if !g.Players[game.Player1].Hand.Contains(cardID) || reboundPermanentNamed(t, g, "Waterlogged Grove").Tapped ||
					g.Players[game.Player1].ManaPool.Total() != 2 || g.Stack.Size() != 0 {
					t.Fatal("casting the face-down spell in the clone changed the original game")
				}
				engine.resolveTopOfStack(clone, &TurnLog{})
				permanent := permanentByCardID(clone, cardID)
				if permanent == nil || !permanent.FaceDown || !faceDownIsTwoTwo(clone, permanent) {
					t.Fatalf("face-down spell resolved as %+v, want a face-down 2/2", permanent)
				}
			})

			t.Run("window declined", func(t *testing.T) {
				clone := g.Clone()
				agent := &faceDownWindowAgent{paymentManaAgent: paymentManaAgent{ability: "no such ability"}, cardID: cardID}
				var attempts castAttempts
				outcome, _ := engine.attemptActionWithChoices(clone, game.Player1, cast, reboundKickerAgents(agent), &TurnLog{}, &attempts)
				if outcome != actionReversed {
					t.Fatalf("outcome = %v, want the unpaid face-down cast reversed", outcome)
				}
				requireNoFaceDownCast(t, clone, cardID)
				if agent.windows != 1 || !clone.Players[game.Player1].Hand.Contains(cardID) ||
					reboundPermanentNamed(t, clone, "Waterlogged Grove").Tapped ||
					clone.Players[game.Player1].Life != life || clone.Players[game.Player1].ManaPool.Total() != 2 {
					t.Fatalf("windows=%d in hand=%t life=%d pool=%d, want the proposal reversed without payment",
						agent.windows, clone.Players[game.Player1].Hand.Contains(cardID),
						clone.Players[game.Player1].Life, clone.Players[game.Player1].ManaPool.Total())
				}
			})
		})
	}
}

// TestFaceDownCastReversalsInPriorityLoop runs a caster who chooses the
// face-down cast of War Behemoth whenever it is offered but never activates
// Waterlogged Grove while paying. Each unpaid cast is reversed and the caster
// keeps priority (CR 733.2); once the repeated proposal is withheld the cast is
// no longer offered, so the caster passes and the step ends with the card
// still in hand and nothing paid.
func TestFaceDownCastReversalsInPriorityLoop(t *testing.T) {
	g, engine, cardID := faceDownCostGame(t, "War Behemoth", mana.C, 2)
	addCombatPermanent(g, game.Player1, productionChoiceCard(t, "Waterlogged Grove"))
	life := g.Players[game.Player1].Life
	agent := &faceDownWindowAgent{paymentManaAgent: paymentManaAgent{ability: "no such ability"}, cardID: cardID}

	engine.runPriorityLoop(g, reboundKickerAgents(agent), &TurnLog{})

	if agent.casts < 2 || agent.windows != agent.casts-1 {
		t.Fatalf("casts=%d windows=%d, want repeated reversed attempts and then exhaustion", agent.casts, agent.windows)
	}
	requireNoFaceDownCast(t, g, cardID)
	if !g.Players[game.Player1].Hand.Contains(cardID) || g.Players[game.Player1].Life != life ||
		g.Players[game.Player1].ManaPool.Total() != 2 || reboundPermanentNamed(t, g, "Waterlogged Grove").Tapped {
		t.Fatal("reversed face-down casts left a payment behind")
	}
}

// TestFaceDownCastLocksCostBeforeManaAbilities casts War Behemoth face down
// under the caster's own Lodestone Golem: the {4} total is locked before
// payment-time mana abilities (CR 601.2f-g), so sacrificing the Golem to
// Ashnod's Altar while paying does not remove the {1} tax already included.
func TestFaceDownCastLocksCostBeforeManaAbilities(t *testing.T) {
	g, engine, cardID := faceDownCostGame(t, "War Behemoth", mana.C, 2)
	golem := addCombatPermanent(g, game.Player1, productionChoiceCard(t, "Lodestone Golem"))
	addCombatPermanent(g, game.Player1, productionChoiceCard(t, "Ashnod's Altar"))
	if !faceDownCastOffered(engine, g, cardID, game.FaceDownMorph) {
		t.Fatal("face-down cast payable with Ashnod's Altar was not offered")
	}
	agent := &faceDownWindowAgent{paymentManaAgent: paymentManaAgent{ability: "Sacrifice a creature"}, cardID: cardID}

	if !engine.applyActionWithChoices(g, game.Player1, actionBuild.castFaceDown(cardID, game.FaceFront, game.FaceDownMorph), reboundKickerAgents(agent), &TurnLog{}) {
		t.Fatal("face-down cast paid with Ashnod's Altar failed")
	}

	requireFaceDownSpell(t, g, cardID, game.FaceDownMorph, zone.Hand, 4)
	if _, onBattlefield := permanentByObjectID(g, golem.ObjectID); agent.activated != 1 || onBattlefield {
		t.Fatalf("activated=%d golem on battlefield=%t, want Lodestone Golem sacrificed for mana", agent.activated, onBattlefield)
	}
	if got := g.Players[game.Player1].ManaPool.Total(); got != 0 {
		t.Fatalf("mana left = %d, want the locked {4} total paid", got)
	}
}

// TestFaceDownReducedToZeroOpensNoManaWindow casts War Behemoth face down
// with two Heartless Summonings: the {3} alternative cost reduced by {4} locks
// a total with no mana payment, so no payment-time mana ability window opens
// (CR 601.2g) although Waterlogged Grove is available.
func TestFaceDownReducedToZeroOpensNoManaWindow(t *testing.T) {
	g, engine, cardID := faceDownCostGame(t, "War Behemoth", mana.C, 0)
	addCombatPermanent(g, game.Player1, productionChoiceCard(t, "Heartless Summoning"))
	addCombatPermanent(g, game.Player1, productionChoiceCard(t, "Heartless Summoning"))
	grove := addCombatPermanent(g, game.Player1, productionChoiceCard(t, "Waterlogged Grove"))
	if !faceDownCastOffered(engine, g, cardID, game.FaceDownMorph) {
		t.Fatal("face-down cast reduced to {0} was not offered")
	}
	agent := &faceDownWindowAgent{paymentManaAgent: paymentManaAgent{ability: "Pay 1 life"}, cardID: cardID}

	if !engine.applyActionWithChoices(g, game.Player1, actionBuild.castFaceDown(cardID, game.FaceFront, game.FaceDownMorph), reboundKickerAgents(agent), &TurnLog{}) {
		t.Fatal("face-down cast reduced to {0} failed")
	}

	requireFaceDownSpell(t, g, cardID, game.FaceDownMorph, zone.Hand, 0)
	if agent.windows != 0 || grove.Tapped {
		t.Fatalf("windows=%d grove tapped=%t, want a {0} face-down cast without a mana window", agent.windows, grove.Tapped)
	}
}

// TestFaceDownGraveyardCastPaysWithPaymentTimeManaAbility casts War Behemoth
// face down from the graveyard through Muldrotha's creature permission with
// two colorless mana and Waterlogged Grove. The cast is offered and the Grove
// pays while paying (CR 601.2g); the spell keeps its graveyard origin and the
// permission's use is recorded. Declining the window reverses the cast
// (CR 733.1): the card returns to its place in the graveyard, below Grizzly
// Bears (CR 404.2), and the permission stays unused.
func TestFaceDownGraveyardCastPaysWithPaymentTimeManaAbility(t *testing.T) {
	g, e, source := muldrothaGame(t)
	cardID := addCardToGraveyard(g, game.Player1, productionChoiceCard(t, "War Behemoth"))
	addCardToGraveyard(g, game.Player1, productionChoiceCard(t, "Grizzly Bears"))
	graveyard := g.Players[game.Player1].Graveyard.All()
	addCombatPermanent(g, game.Player1, productionChoiceCard(t, "Waterlogged Grove"))
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)
	permission := muldrothaPermission(source, types.Creature)
	cast := action.CastFaceDownFromZone(cardID, zone.Graveyard, game.FaceFront, game.FaceDownMorph).WithPlayPermission(permission)
	use := game.PlayPermissionUse{Permission: permission, Player: game.Player1}
	life := g.Players[game.Player1].Life
	if !containsAction(e.legalActions(g, game.Player1), cast) {
		t.Fatal("graveyard face-down cast payable with Waterlogged Grove was not offered")
	}

	t.Run("grove activated", func(t *testing.T) {
		clone := g.Clone()
		agent := &faceDownWindowAgent{paymentManaAgent: paymentManaAgent{ability: "Pay 1 life"}, cardID: cardID}
		if !e.applyActionWithChoices(clone, game.Player1, cast, reboundKickerAgents(agent), &TurnLog{}) {
			t.Fatal("graveyard face-down cast paid with Waterlogged Grove failed")
		}
		requireFaceDownSpell(t, clone, cardID, game.FaceDownMorph, zone.Graveyard, 3)
		if agent.windows != 1 || agent.activated != 1 || clone.Players[game.Player1].Life != life-1 ||
			clone.Players[game.Player1].Graveyard.Contains(cardID) || !clone.PermanentTypePlayUsesThisTurn[use] {
			t.Fatalf("windows=%d activated=%d life=%d permission used=%t, want the Grove to pay through Muldrotha",
				agent.windows, agent.activated, clone.Players[game.Player1].Life, clone.PermanentTypePlayUsesThisTurn[use])
		}
	})

	t.Run("window declined", func(t *testing.T) {
		clone := g.Clone()
		agent := &faceDownWindowAgent{paymentManaAgent: paymentManaAgent{ability: "no such ability"}, cardID: cardID}
		if e.applyActionWithChoices(clone, game.Player1, cast, reboundKickerAgents(agent), &TurnLog{}) {
			t.Fatal("unpaid graveyard face-down cast succeeded")
		}
		requireNoFaceDownCast(t, clone, cardID)
		if agent.windows != 1 || !clone.Players[game.Player1].Graveyard.Contains(cardID) ||
			clone.Players[game.Player1].Life != life || clone.Players[game.Player1].ManaPool.Total() != 2 ||
			reboundPermanentNamed(t, clone, "Waterlogged Grove").Tapped || clone.PermanentTypePlayUsesThisTurn[use] {
			t.Fatal("reversed graveyard face-down cast left the card, a payment or the permission use behind")
		}
		if got := clone.Players[game.Player1].Graveyard.All(); !slices.Equal(got, graveyard) {
			t.Fatalf("graveyard after reversal = %v, want its order %v kept (CR 404.2)", got, graveyard)
		}
	})
}

// manaSequenceAgent activates, in each payment-time mana ability window, the
// first of abilities it has not yet activated whose label is offered, and
// otherwise finishes activating mana abilities.
type manaSequenceAgent struct {
	abilities []string
	activated []string
}

func (*manaSequenceAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (a *manaSequenceAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	if request.Prompt != "Activate a mana ability while paying?" {
		return request.DefaultSelection
	}
	for _, ability := range a.abilities[len(a.activated):] {
		for _, option := range request.Options {
			if strings.Contains(option.Label, ability) {
				a.activated = append(a.activated, ability)
				return []int{option.Index}
			}
		}
	}
	return request.DefaultSelection
}

// TestFaceDownCastOfferedWhenManaAbilitiesCombine gives the caster no mana,
// Waterlogged Grove, Ashnod's Altar and Grizzly Bears: neither ability alone
// can pay the locked {3}, but Grove's one mana and the Altar's two together
// can (CR 601.2g-h), so the face-down cast is offered and pays with both.
func TestFaceDownCastOfferedWhenManaAbilitiesCombine(t *testing.T) {
	g, engine, cardID := faceDownCostGame(t, "War Behemoth", mana.C, 0)
	addCombatPermanent(g, game.Player1, productionChoiceCard(t, "Waterlogged Grove"))
	addCombatPermanent(g, game.Player1, productionChoiceCard(t, "Ashnod's Altar"))
	bears := addCombatPermanent(g, game.Player1, productionChoiceCard(t, "Grizzly Bears"))
	life := g.Players[game.Player1].Life
	if !faceDownCastOffered(engine, g, cardID, game.FaceDownMorph) {
		t.Fatal("face-down cast payable with Waterlogged Grove and Ashnod's Altar together was not offered")
	}
	agent := &manaSequenceAgent{abilities: []string{"Pay 1 life", "Sacrifice a creature"}}

	if !engine.applyActionWithChoices(g, game.Player1, actionBuild.castFaceDown(cardID, game.FaceFront, game.FaceDownMorph), reboundKickerAgents(agent), &TurnLog{}) {
		t.Fatal("face-down cast paid with Waterlogged Grove and Ashnod's Altar failed")
	}

	requireFaceDownSpell(t, g, cardID, game.FaceDownMorph, zone.Hand, 3)
	_, bearsOnBattlefield := permanentByObjectID(g, bears.ObjectID)
	if len(agent.activated) != 2 || bearsOnBattlefield || !reboundPermanentNamed(t, g, "Waterlogged Grove").Tapped ||
		g.Players[game.Player1].Life != life-1 || g.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatalf("activated=%v bears on battlefield=%t life=%d pool=%d, want Grove and the Altar to pay {3}",
			agent.activated, bearsOnBattlefield, g.Players[game.Player1].Life, g.Players[game.Player1].ManaPool.Total())
	}
}

// TestFaceDownCastNotOfferedWhenManaAbilitiesFallShort keeps a face-down
// cast unoffered when the caster's payment-time mana abilities can't make its
// locked total payable (CR 601.2f-h): Waterlogged Grove adds only one mana
// toward {3} from an empty or one-mana pool, from hand or from the graveyard
// through Muldrotha; and sacrificing Lodestone Golem to Ashnod's Altar can't
// pay the {4} total locked while the Golem's tax applied. Checking leaves the
// game unchanged, and a direct attempt fails without paying anything.
func TestFaceDownCastNotOfferedWhenManaAbilitiesFallShort(t *testing.T) {
	for _, tc := range []struct {
		name       string
		pool       int
		permanents []string
		fromGrave  bool
	}{
		{name: "grove from empty pool", permanents: []string{"Waterlogged Grove"}},
		{name: "grove from one mana", pool: 1, permanents: []string{"Waterlogged Grove"}},
		{name: "altar sacrificing taxing golem", pool: 1, permanents: []string{"Lodestone Golem", "Ashnod's Altar"}},
		{name: "graveyard grove from empty pool", permanents: []string{"Waterlogged Grove"}, fromGrave: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var (
				g      *game.Game
				engine *Engine
				cardID id.ID
				cast   action.Action
				from   = zone.Hand
			)
			if tc.fromGrave {
				var source *game.Permanent
				g, engine, source = muldrothaGame(t)
				cardID = addCardToGraveyard(g, game.Player1, productionChoiceCard(t, "War Behemoth"))
				g.Players[game.Player1].ManaPool.Add(mana.C, tc.pool)
				cast = action.CastFaceDownFromZone(cardID, zone.Graveyard, game.FaceFront, game.FaceDownMorph).
					WithPlayPermission(muldrothaPermission(source, types.Creature))
				from = zone.Graveyard
			} else {
				g, engine, cardID = faceDownCostGame(t, "War Behemoth", mana.C, tc.pool)
				cast = actionBuild.castFaceDown(cardID, game.FaceFront, game.FaceDownMorph)
			}
			for _, name := range tc.permanents {
				addCombatPermanent(g, game.Player1, productionChoiceCard(t, name))
			}
			life, events, nextID, battlefield := g.Players[game.Player1].Life, len(g.Events), g.IDGen.Current(), len(g.Battlefield)

			for _, act := range engine.legalActions(g, game.Player1) {
				if payload, ok := act.CastFaceDownPayload(); ok && payload.CardID == cardID {
					t.Fatalf("face-down cast %+v offered though its mana abilities can't pay the total", payload)
				}
			}
			if len(g.Events) != events || g.IDGen.Current() != nextID || len(g.Battlefield) != battlefield ||
				g.Players[game.Player1].Life != life || g.Players[game.Player1].ManaPool.Total() != tc.pool ||
				reboundPermanentNamed(t, g, tc.permanents[0]).Tapped {
				t.Fatal("checking whether the face-down cast is payable changed the game")
			}

			agent := &manaSequenceAgent{abilities: []string{"Pay 1 life", "Sacrifice a creature"}}
			outcome, _ := engine.attemptActionWithChoices(g, game.Player1, cast, reboundKickerAgents(agent), &TurnLog{}, nil)
			if outcome != actionFailed || len(agent.activated) != 0 {
				t.Fatalf("outcome=%v activated=%v, want the illegal face-down cast to fail before any payment", outcome, agent.activated)
			}
			requireNoFaceDownCast(t, g, cardID)
			source := g.Players[game.Player1].Hand
			if from == zone.Graveyard {
				source = g.Players[game.Player1].Graveyard
			}
			if !source.Contains(cardID) || g.Players[game.Player1].Life != life ||
				g.Players[game.Player1].ManaPool.Total() != tc.pool || len(g.Battlefield) != battlefield {
				t.Fatal("the failed face-down cast changed the game")
			}
		})
	}
}

// TestFaceDownCastOfferDoesNotRevealLibrary gives the caster one mana,
// Millikin and Deathbloom Ritualist. Millikin's ability mills a card, so it
// isn't a mana ability (CR 605.1a) and can't be activated while paying (CR
// 601.2g): the Ritualist, with no creature card in the graveyard, can't make
// up the locked {3}, so the face-down cast isn't offered whatever the library
// holds and in whatever order (CR 401.2), and checking leaves the game
// unchanged. Activating Millikin with priority first puts its ability on the
// stack (CR 602.2), milling the top card publicly as a cost; once it resolves,
// a milled creature card lets the Ritualist pay the rest of {3} and the cast
// is offered and made, while a milled land doesn't.
func TestFaceDownCastOfferDoesNotRevealLibrary(t *testing.T) {
	for _, library := range [][]string{
		{"Grizzly Bears"}, {"Forest"}, {"Forest", "Grizzly Bears"}, {"Grizzly Bears", "Forest"},
	} {
		t.Run(strings.Join(library, " then "), func(t *testing.T) {
			g, engine, cardID := faceDownCostGame(t, "War Behemoth", mana.C, 1)
			millikin := addCombatPermanent(g, game.Player1, productionChoiceCard(t, "Millikin"))
			addCombatPermanent(g, game.Player1, productionChoiceCard(t, "Deathbloom Ritualist"))
			for _, name := range library {
				addCardToLibrary(g, game.Player1, productionChoiceCard(t, name))
			}
			order, events, nextID := g.Players[game.Player1].Library.All(), len(g.Events), g.IDGen.Current()

			if faceDownCastOffered(engine, g, cardID, game.FaceDownMorph) {
				t.Fatal("face-down cast offered as if Millikin could pay while casting")
			}
			if !slices.Equal(g.Players[game.Player1].Library.All(), order) || g.Players[game.Player1].Graveyard.Size() != 0 ||
				len(g.Events) != events || g.IDGen.Current() != nextID || g.Players[game.Player1].ManaPool.Total() != 1 || millikin.Tapped {
				t.Fatal("checking whether the face-down cast is payable changed the game")
			}
			if len(library) != 1 {
				return
			}

			agent := &manaSequenceAgent{abilities: []string{"number of creature cards"}}
			agents := reboundKickerAgents(agent)
			if !engine.applyActionWithChoices(g, game.Player1, sourceActivation(t, engine, g, millikin.ObjectID), agents, &TurnLog{}) {
				t.Fatal("activating Millikin failed")
			}
			if obj, ok := g.Stack.Peek(); !ok || obj.Kind != game.StackActivatedAbility || !g.Players[game.Player1].Graveyard.Contains(order[0]) {
				t.Fatalf("stack top = %+v, want Millikin's ability on the stack after milling the top card", obj)
			}
			engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
			if g.Stack.Size() != 0 || g.Players[game.Player1].ManaPool.Total() != 2 {
				t.Fatalf("pool = %d, want Millikin's {C} added on resolution", g.Players[game.Player1].ManaPool.Total())
			}

			offered := faceDownCastOffered(engine, g, cardID, game.FaceDownMorph)
			if offered != (library[0] == "Grizzly Bears") {
				t.Fatalf("face-down cast offered = %t after milling %s", offered, library[0])
			}
			if !offered {
				return
			}
			cast := actionBuild.castFaceDown(cardID, game.FaceFront, game.FaceDownMorph)
			if outcome, _ := engine.attemptActionWithChoices(g, game.Player1, cast, agents, &TurnLog{}, nil); outcome != actionApplied || len(agent.activated) != 1 {
				t.Fatalf("outcome=%v activated=%v, want the Ritualist to pay the rest of {3}", outcome, agent.activated)
			}
			requireFaceDownSpell(t, g, cardID, game.FaceDownMorph, zone.Hand, 3)
		})
	}
}

// TestFaceDownGraveyardCastProbesWithCardOnStack casts War Behemoth face down
// from the graveyard through Muldrotha with two mana and Deathbloom Ritualist.
// The card moves to the stack before mana abilities are activated (CR 601.2a,
// 601.2g), so the Ritualist doesn't count it: alone in the graveyard it adds
// nothing and the cast isn't offered, while with Grizzly Bears beside it the
// Ritualist adds one mana and the cast is offered and paid.
func TestFaceDownGraveyardCastProbesWithCardOnStack(t *testing.T) {
	for _, withBears := range []bool{false, true} {
		t.Run(map[bool]string{false: "behemoth alone", true: "behemoth and bears"}[withBears], func(t *testing.T) {
			g, engine, source := muldrothaGame(t)
			if withBears {
				addCardToGraveyard(g, game.Player1, productionChoiceCard(t, "Grizzly Bears"))
			}
			cardID := addCardToGraveyard(g, game.Player1, productionChoiceCard(t, "War Behemoth"))
			addCombatPermanent(g, game.Player1, productionChoiceCard(t, "Deathbloom Ritualist"))
			g.Players[game.Player1].ManaPool.Add(mana.C, 2)
			cast := action.CastFaceDownFromZone(cardID, zone.Graveyard, game.FaceFront, game.FaceDownMorph).
				WithPlayPermission(muldrothaPermission(source, types.Creature))
			graveyard := g.Players[game.Player1].Graveyard.All()

			offered := containsAction(engine.legalActions(g, game.Player1), cast)
			if offered != withBears {
				t.Fatalf("offered=%t, want %t: the Ritualist mustn't count the card being cast", offered, withBears)
			}
			if !slices.Equal(g.Players[game.Player1].Graveyard.All(), graveyard) || g.Stack.Size() != 0 {
				t.Fatal("checking whether the face-down cast is payable changed the game")
			}
			agent := &manaSequenceAgent{abilities: []string{"number of creature cards"}}
			outcome, _ := engine.attemptActionWithChoices(g, game.Player1, cast, reboundKickerAgents(agent), &TurnLog{}, nil)
			if !withBears {
				if outcome != actionFailed || len(agent.activated) != 0 || !slices.Equal(g.Players[game.Player1].Graveyard.All(), graveyard) {
					t.Fatalf("outcome=%v activated=%v, want the illegal cast to fail untouched", outcome, agent.activated)
				}
				requireNoFaceDownCast(t, g, cardID)
				return
			}
			if outcome != actionApplied || len(agent.activated) != 1 || g.Players[game.Player1].ManaPool.Total() != 0 {
				t.Fatalf("outcome=%v activated=%v pool=%d, want the Ritualist's one mana to pay the rest",
					outcome, agent.activated, g.Players[game.Player1].ManaPool.Total())
			}
			requireFaceDownSpell(t, g, cardID, game.FaceDownMorph, zone.Graveyard, 3)
		})
	}
}
