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
	"github.com/natefinch/council4/opt"
)

// Source review: issue #3584. Default-registry cards checked against the
// pinned corpus (oracle_id): Millikin fa4dffda-6f04-4d0b-829d-28a1a5794dee
// "{T}, Mill a card: Add {C}. (Activate only as an instant. ...)"; Waterlogged
// Grove 70fa2eba-565e-4fed-adc9-7f5d9fcbf1fa "{T}, Pay 1 life: Add {G} or
// {U}. ..."; Muldrotha, the Gravetide e4625704-1d52-44e4-804f-2f45644d76ac
// "During each of your turns, you may play a land and cast a permanent spell of
// each permanent type from your graveyard. ..."; Grizzly Bears
// 14c8f55d-d177-4c25-a931-ebeb9e6062a0 {1}{G}; War Behemoth
// 4792a05e-b557-4132-8303-235feefe0d99 "Morph {4}{W} (You may cast this card
// face down as a 2/2 creature for {3}. ...)"; Gemrazer
// 3dfb0c0a-b68f-43b9-8475-28d0192fc4ed "Mutate {1}{G}{G} (If you cast this
// spell for its mutate cost, put it over or under target non-Human creature you
// own. ...)"; Conduit of Worlds ed14be15-8f8d-4fe3-a147-f5da8ed873bf "{T}:
// Choose target nonland permanent card in your graveyard. If you haven't cast a
// spell this turn, you may cast that card. ..."; Sol Ring
// 6ad8011d-3471-4369-9d68-b264cc027487 {1}; Thalia, Guardian of Thraben
// 9b7f1d05-707c-4ed3-9f0e-8ced1232c2ee "Noncreature spells cost {1} more to
// cast."; Gods' Hall Guardian 7d22e071-4af4-4182-b387-ecca873d00cc {5}{W}
// "Vigilance / Foretell {3}{W} (During your turn, you may pay {2} and exile
// this card from your hand face down. Cast it on a later turn for its foretell
// cost.)"; Forest b34bb2dc-c1af-4d77-b0b3-a0fb342a5fc6.
// Comprehensive Rules effective 2026-09-25: 601.2a moves the card to the stack
// first; 601.2g lets the caster activate mana abilities before paying; 733.1
// "the entire action is reversed ... If the action was casting a spell, the
// spell returns to the zone it came from", and "Players may not reverse
// actions that moved cards ... from a library to any zone other than the
// stack", so Millikin's mill stays; 404.2 "players normally can't change"
// a graveyard's order; 406.3 a card exiled face down stays face down and
// players can't look at it unless an instruction allows it.

// sourceRestoreAgent accepts optional casts and activates only Millikin's mana
// ability while paying, leaving the cost short so the cast is reversed after
// the mill.
type sourceRestoreAgent struct {
	milled int
}

func (*sourceRestoreAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (a *sourceRestoreAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	switch {
	case request.Kind == game.ChoiceMay:
		return []int{1}
	case request.Prompt == "Activate a mana ability while paying?" && a.milled == 0:
		for _, option := range request.Options {
			if strings.Contains(option.Label, "Mill a card") {
				a.milled++
				return []int{option.Index}
			}
		}
	}
	return request.DefaultSelection
}

func addCardToGraveyardBottom(g *game.Game, playerID game.PlayerID, def *game.CardDef) id.ID {
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{ID: cardID, Def: def, Owner: playerID}
	g.Players[playerID].Graveyard.AddToBottom(cardID)
	return cardID
}

// graveyardRestoreRoute casts a default-registry card out of Player1's
// graveyard on one cast route, with Millikin and Waterlogged Grove able to pay
// together but Millikin alone not.
type graveyardRestoreRoute struct {
	name  string
	build func(t *testing.T) (*game.Game, id.ID, func([game.NumPlayers]PlayerAgent))
}

func graveyardRestoreRoutes() []graveyardRestoreRoute {
	thalias := func(t *testing.T, g *game.Game, n int) {
		for range n {
			addCombatPermanent(g, game.Player2, productionChoiceCard(t, "Thalia, Guardian of Thraben"))
		}
	}
	return []graveyardRestoreRoute{
		{name: "face-down cast through Muldrotha", build: func(t *testing.T) (*game.Game, id.ID, func([game.NumPlayers]PlayerAgent)) {
			g, e, source := muldrothaGame(t)
			cardID := addCardToGraveyard(g, game.Player1, productionChoiceCard(t, "War Behemoth"))
			g.Players[game.Player1].ManaPool.Add(mana.C, 1)
			cast := action.CastFaceDownFromZone(cardID, zone.Graveyard, game.FaceFront, game.FaceDownMorph).
				WithPlayPermission(muldrothaPermission(source, types.Creature))
			return g, cardID, func(agents [game.NumPlayers]PlayerAgent) {
				if outcome, _ := e.attemptActionWithChoices(g, game.Player1, cast, agents, &TurnLog{}, nil); outcome != actionReversed {
					t.Fatalf("outcome = %v, want the unpaid face-down cast reversed", outcome)
				}
				if len(g.PermanentTypePlayUsesThisTurn) != 0 {
					t.Fatal("the reversed cast spent Muldrotha's permission")
				}
			}
		}},
		{name: "paid cast through Conduit of Worlds", build: func(t *testing.T) (*game.Game, id.ID, func([game.NumPlayers]PlayerAgent)) {
			e, g, cardID := conduitPaidCastGame(t, "Sol Ring")
			thalias(t, g, 1)
			return g, cardID, func(agents [game.NumPlayers]PlayerAgent) {
				e.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
				if spellCastProhibited(g, game.Player1, vanillaCreatureDef()) {
					t.Fatal("Conduit's can't-cast-additional lock applied after a reversed cast")
				}
			}
		}},
		{name: "free cast", build: func(t *testing.T) (*game.Game, id.ID, func([game.NumPlayers]PlayerAgent)) {
			e, g := newEffectiveCascadeGame()
			addCombatPermanent(g, game.Player1, productionChoiceCard(t, "Waterlogged Grove"))
			cardID := addCardToGraveyard(g, game.Player1, productionChoiceCard(t, "Sol Ring"))
			thalias(t, g, 2)
			return g, cardID, func(agents [game.NumPlayers]PlayerAgent) {
				e.chooseAndCastFreeSpellFromSource(g, g.Players[game.Player1], game.Player1, cardID, zone.Graveyard,
					game.Selection{}, opt.V[int]{}, false, agents, &TurnLog{})
			}
		}},
	}
}

// TestReversedGraveyardCastKeepsGraveyardOrder casts a card from the middle of
// its owner's graveyard on each route that opens a payment-time mana ability
// window: a face-down cast through Muldrotha, a paid cast through Conduit of
// Worlds and a free cast. Millikin and Waterlogged Grove could pay together, so
// the cast is proposed, but the caster activates only Millikin, milling Forest
// onto the graveyard (CR 601.2g), and the cost goes unpaid. The cast is
// reversed (CR 733.1): Millikin's mill stays, and the card returns between the
// cards that were above and below it rather than on top (CR 404.2), with the
// milled Forest still on top.
func TestReversedGraveyardCastKeepsGraveyardOrder(t *testing.T) {
	for _, route := range graveyardRestoreRoutes() {
		t.Run(route.name, func(t *testing.T) {
			g, cardID, run := route.build(t)
			addCombatPermanent(g, game.Player1, productionChoiceCard(t, "Millikin"))
			if !slices.ContainsFunc(g.Battlefield, func(permanent *game.Permanent) bool {
				card, ok := g.GetCardInstance(permanent.CardInstanceID)
				return ok && card.Def.Name == "Waterlogged Grove"
			}) {
				addCombatPermanent(g, game.Player1, productionChoiceCard(t, "Waterlogged Grove"))
			}
			above := addCardToGraveyard(g, game.Player1, productionChoiceCard(t, "Grizzly Bears"))
			below := addCardToGraveyardBottom(g, game.Player1, productionChoiceCard(t, "Forest"))
			milled := addCardToLibrary(g, game.Player1, productionChoiceCard(t, "Forest"))
			if got, want := g.Players[game.Player1].Graveyard.All(), []id.ID{above, cardID, below}; !slices.Equal(got, want) {
				t.Fatalf("graveyard before = %v, want %v", got, want)
			}
			life := g.Players[game.Player1].Life
			agent := &sourceRestoreAgent{}

			run(reboundKickerAgents(agent))

			if agent.milled != 1 || !reboundPermanentNamed(t, g, "Millikin").Tapped || g.Players[game.Player1].Library.Size() != 0 {
				t.Fatal("Millikin didn't mill while paying")
			}
			if g.Stack.Size() != 0 || reboundPermanentNamed(t, g, "Waterlogged Grove").Tapped || g.Players[game.Player1].Life != life {
				t.Fatal("the unpaid cast left a spell or a Grove payment behind")
			}
			for _, event := range g.Events {
				if event.Kind == game.EventSpellCast && event.CardID == cardID {
					t.Fatal("the reversed cast produced a spell-cast event")
				}
			}
			if got, want := g.Players[game.Player1].Graveyard.All(), []id.ID{milled, above, cardID, below}; !slices.Equal(got, want) {
				t.Fatalf("graveyard after reversal = %v, want %v: the card back in its place under the milled card (CR 404.2)", got, want)
			}
		})
	}
}

// TestReversedPriorityGraveyardCastKeepsGraveyardOrder casts Vault Skirge
// ({1}{B/P}) from the middle of its owner's graveyard with priority through
// Muldrotha, announcing {B} for its Phyrexian symbol with only {C} in pool. The
// announcement is binding (CR 601.2b, 601.2h), so the cast is reversed (CR
// 733.1) and the card returns between the cards that were above and below it
// (CR 404.2) without spending Muldrotha's permission. Paying 2 life on the next
// attempt casts it.
func TestReversedPriorityGraveyardCastKeepsGraveyardOrder(t *testing.T) {
	g, e, source := muldrothaGame(t)
	cardID := addCardToGraveyard(g, game.Player1, productionChoiceCard(t, "Vault Skirge"))
	above := addCardToGraveyard(g, game.Player1, productionChoiceCard(t, "Grizzly Bears"))
	below := addCardToGraveyardBottom(g, game.Player1, productionChoiceCard(t, "Forest"))
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	graveyard := &g.Players[game.Player1].Graveyard
	if got, want := graveyard.All(), []id.ID{above, cardID, below}; !slices.Equal(got, want) {
		t.Fatalf("graveyard before = %v, want %v", got, want)
	}
	life := g.Players[game.Player1].Life
	cast := graveyardCast(cardID, source, types.Creature)
	var attempts castAttempts

	outcome, _ := e.attemptActionWithChoices(g, game.Player1, cast, reboundKickerAgents(&phyrexianWindowAgent{payMana: true}), &TurnLog{}, &attempts)

	if outcome != actionReversed || g.Stack.Size() != 0 || g.Players[game.Player1].Life != life ||
		g.Players[game.Player1].ManaPool.Total() != 1 || len(g.PermanentTypePlayUsesThisTurn) != 0 {
		t.Fatalf("outcome=%v stack=%d life=%d pool=%d permission uses=%d, want the cast with {B} announced reversed",
			outcome, g.Stack.Size(), g.Players[game.Player1].Life, g.Players[game.Player1].ManaPool.Total(), len(g.PermanentTypePlayUsesThisTurn))
	}
	if got, want := graveyard.All(), []id.ID{above, cardID, below}; !slices.Equal(got, want) {
		t.Fatalf("graveyard after reversal = %v, want %v: the card back in its place (CR 404.2)", got, want)
	}

	outcome, _ = e.attemptActionWithChoices(g, game.Player1, cast, reboundKickerAgents(&phyrexianWindowAgent{payLife: true}), &TurnLog{}, &attempts)

	if obj, ok := g.Stack.Peek(); outcome != actionApplied || !ok || obj.SourceID != cardID || g.Players[game.Player1].Life != life-2 {
		t.Fatalf("outcome=%v stack top=%+v life=%d, want Vault Skirge cast for {C} and 2 life", outcome, obj, g.Players[game.Player1].Life)
	}
}

// TestReversedFreeCastFromExileStaysFaceDownInPlace foretells Doomskar between
// two other exiled cards, then on a later turn casts it from exile without
// paying its mana cost while two Thalias add {2}. Millikin and Waterlogged
// Grove could pay that together, but only Millikin is activated, so the cast is
// reversed (CR 733.1): the card returns to its place in exile still face down
// (CR 406.3) and foretold, and its foretell cast is still offered.
func TestReversedFreeCastFromExileStaysFaceDownInPlace(t *testing.T) {
	e, g := newEffectiveCascadeGame()
	below := addCardToExile(g, game.Player1, productionChoiceCard(t, "Forest"))
	cardID := addCardToHand(g, game.Player1, productionChoiceCard(t, "Doomskar"))
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)
	setSorceryTiming(g, game.Player1)
	g.Turn.TurnNumber = 3
	if !e.applyAction(g, game.Player1, action.ForetellCard(cardID)) {
		t.Fatal("foretelling Doomskar failed")
	}
	above := addCardToExile(g, game.Player1, productionChoiceCard(t, "Grizzly Bears"))
	g.Turn.TurnNumber = 4
	for range 2 {
		addCombatPermanent(g, game.Player2, productionChoiceCard(t, "Thalia, Guardian of Thraben"))
	}
	addCombatPermanent(g, game.Player1, productionChoiceCard(t, "Millikin"))
	grove := addCombatPermanent(g, game.Player1, productionChoiceCard(t, "Waterlogged Grove"))
	addCardToLibrary(g, game.Player1, productionChoiceCard(t, "Forest"))
	exile := &g.Players[game.Player1].Exile
	if got, want := exile.All(), []id.ID{above, cardID, below}; !slices.Equal(got, want) || !exile.IsFaceDown(cardID) {
		t.Fatalf("exile before = %v face down=%t, want %v with the foretold card face down", got, exile.IsFaceDown(cardID), want)
	}
	agent := &sourceRestoreAgent{}

	e.chooseAndCastFreeSpellFromSource(g, g.Players[game.Player1], game.Player1, cardID, zone.Exile,
		game.Selection{}, opt.V[int]{}, false, reboundKickerAgents(agent), &TurnLog{})

	if agent.milled != 1 || g.Stack.Size() != 0 || grove.Tapped {
		t.Fatalf("milled=%d stack=%d grove tapped=%t, want the unpaid free cast reversed after Millikin's mill", agent.milled, g.Stack.Size(), grove.Tapped)
	}
	if got, want := exile.All(), []id.ID{above, cardID, below}; !slices.Equal(got, want) {
		t.Fatalf("exile after reversal = %v, want %v", got, want)
	}
	if !exile.IsFaceDown(cardID) {
		t.Fatal("the reversed foretold card returned to exile face up")
	}
	g.Players[game.Player1].ManaPool.Add(mana.W, 2)
	g.Players[game.Player1].ManaPool.Add(mana.C, 3)
	cast := action.CastSpellFaceFromZone(cardID, zone.Exile, game.FaceFront, nil, 0, nil)
	if !cardIsForetoldInExile(g, game.Player1, cardID) || !containsAction(e.legalActions(g, game.Player1), cast) {
		t.Fatal("the reversed foretold card lost its foretell cast")
	}
}
