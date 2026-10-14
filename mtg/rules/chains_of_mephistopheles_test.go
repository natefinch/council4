package rules

import (
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/cards/authored"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// Scenarios use the pinned Oracle corpus (6da9ced0...) and official CR
// effective 2026-09-25: 121.2, 121.6, 614.5, 614.11, 616, 701.9 and 702.52.
type chainsChoices struct {
	abundanceChoices

	order []string
}

func (a *chainsChoices) ChooseChoice(obs PlayerObservation, request game.ChoiceRequest) []int {
	if request.Kind == game.ChoiceReplacement && len(a.order) > 0 {
		for _, option := range request.Options {
			if strings.Contains(option.Label, a.order[0]) {
				a.order = a.order[1:]
				a.requests = append(a.requests, request)
				return []int{option.Index}
			}
		}
	}
	if request.Prompt == "Choose cards to discard" {
		a.requests = append(a.requests, request)
		if obs.Player != request.Player || request.MinChoices != 1 || request.MaxChoices != 1 {
			panic("Chains did not ask the affected player for one discard")
		}
		return []int{len(request.Options) - 1}
	}
	if request.Kind == game.ChoiceZoneSelection {
		a.requests = append(a.requests, request)
		return []int{0}
	}
	return a.abundanceChoices.ChooseChoice(obs, request)
}

func challengeCastEnchantment(t *testing.T, g *game.Game, e *Engine, player game.PlayerID, def *game.CardDef) *game.Permanent {
	t.Helper()
	card := addCardToHand(g, player, def)
	g.Players[player].ManaPool.Add(mana.C, 2)
	switch {
	case slices.Contains(def.Colors, color.Black):
		g.Players[player].ManaPool.Add(mana.B, 1)
	case slices.Contains(def.Colors, color.Green):
		g.Players[player].ManaPool.Add(mana.G, 2)
	default:
		g.Players[player].ManaPool.Add(mana.W, 2)
	}
	if !e.applyAction(g, player, action.CastSpell(card, nil, 0, nil)) {
		t.Fatalf("%s could not be cast", def.Name)
	}
	e.resolveTopOfStack(g, &TurnLog{})
	p := permanentByCardID(g, card)
	if p == nil {
		t.Fatalf("%s did not resolve onto the battlefield", def.Name)
	}
	return p
}

func chainsGame(t *testing.T, count int) (*game.Game, *Engine) {
	t.Helper()
	registry := cards.NewDefaultRegistry()
	g, _ := setupBestowMain(t)
	e := NewEngineWithPrograms(nil, registry.Programs())
	for range count {
		def, err := registry.LookupOracleID(authored.ChainsOfMephistopheles.OracleID)
		if err != nil {
			t.Fatal(err)
		}
		challengeCastEnchantment(t, g, e, game.Player1, def)
	}
	return g, e
}

func chainsHand(g *game.Game, count int) []id.ID {
	var cardIDs []id.ID
	for range count {
		cardIDs = append(cardIDs, addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
			Name: "Discardable", Types: []types.Card{types.Instant},
		}}))
	}
	return cardIDs
}

func TestChainsDiscardSuccessEmptyHandAndReplacementDestination(t *testing.T) {
	t.Parallel()
	for _, hand := range []int{0, 1, 2} {
		for _, redirected := range []bool{false, true} {
			t.Run(strings.Repeat("card", hand)+"/exile="+map[bool]string{false: "no", true: "yes"}[redirected], func(t *testing.T) {
				t.Parallel()
				g, e := chainsGame(t, 1)
				held := chainsHand(g, hand)
				top := abundanceLibrary(g, "Top", "Keep")[0]
				if redirected {
					g.ReplacementEffects = append(g.ReplacementEffects, game.ReplacementEffect{
						ID: g.IDGen.Next(), MatchEvent: game.EventZoneChanged,
						FromZone: zone.Hand, MatchFromZone: true, ToZone: zone.Graveyard, ReplaceToZone: zone.Exile,
					})
				}
				a := &chainsChoices{}
				drew := e.drawCards(g, game.Player1, 1, [game.NumPlayers]PlayerAgent{a}, &TurnLog{})
				if drew != (hand > 0) || g.Players[game.Player1].Hand.Size() != hand ||
					g.Players[game.Player1].Library.Size() != 1 || g.FailedDraws[game.Player1] {
					t.Fatal("discard-success draw or empty-hand mill was not performed")
				}
				if hand == 0 {
					if !g.Players[game.Player1].Graveyard.Contains(top) ||
						len(abundanceEvents(g, game.EventCardDrawn)) != 0 {
						t.Fatal("empty hand did not mill without drawing")
					}
					return
				}
				destination := g.Players[game.Player1].Graveyard
				if redirected {
					destination = g.Players[game.Player1].Exile
				}
				if !destination.Contains(held[0]) ||
					!g.Players[game.Player1].Hand.Contains(top) ||
					!slices.Equal(abundanceEvents(g, game.EventCardDiscarded), []id.ID{held[0]}) ||
					!slices.Equal(abundanceEvents(g, game.EventCardDrawn), []id.ID{top}) {
					t.Fatal("discard choice/destination or actual draw events differ")
				}
			})
		}
	}
}

func TestChainsMultipleCopiesAndNestedDrawLineage(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name                                    string
		hand, draws, discards, mills, remaining int
		double                                  bool
	}{
		{name: "two discards then draw", hand: 2, draws: 1, discards: 2, remaining: 1},
		{name: "one discard then stop at mill", hand: 1, discards: 1, mills: 1},
		{name: "empty stops at first mill", mills: 1},
		{name: "sibling lineages", hand: 3, draws: 2, discards: 4, remaining: 1, double: true},
		{name: "sibling mills independent", hand: 1, discards: 1, mills: 2, double: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			g, e := chainsGame(t, 2)
			chainsHand(g, test.hand)
			abundanceLibrary(g, "A", "B", "C", "D", "Keep")
			if test.double {
				drawDoublingPermanent(g, game.Player1, 2, false)
			}
			a := &chainsChoices{order: []string{"doubler"}}
			e.drawCards(g, game.Player1, 1, [game.NumPlayers]PlayerAgent{a}, &TurnLog{})
			if len(abundanceEvents(g, game.EventCardDrawn)) != test.draws ||
				len(abundanceEvents(g, game.EventCardDiscarded)) != test.discards ||
				g.Players[game.Player1].Graveyard.Size() != test.discards+test.mills ||
				g.Players[game.Player1].Hand.Size() != test.remaining ||
				g.Players[game.Player1].Library.Size() != 5-test.draws-test.mills {
				t.Fatalf("draw/discard/hand/graveyard = %d/%d/%d/%d", len(abundanceEvents(g, game.EventCardDrawn)),
					len(abundanceEvents(g, game.EventCardDiscarded)), g.Players[game.Player1].Hand.Size(), g.Players[game.Player1].Graveyard.Size())
			}
			if _, exists := g.ChoiceContext(); exists {
				t.Fatal("replacement lineage escaped its invocation")
			}
		})
	}
}

func TestChainsAbundanceAndDredgeCompeteAtEachDerivedDraw(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name                                    string
		chains, hand, discard, grave, finalHand int
		abundance, dredge                       bool
		order                                   []string
	}{
		{name: "Abundance first", chains: 2, hand: 2, abundance: true, finalHand: 3, order: []string{"choose land or nonland"}},
		{name: "two Chains then Abundance", chains: 2, hand: 2, abundance: true, discard: 2, grave: 2, finalHand: 1, order: []string{"that player discards", "that player discards", "choose land or nonland"}},
		{name: "empty Chains stops Abundance", chains: 2, abundance: true, grave: 1, order: []string{"that player discards"}},
		{name: "dredge first", chains: 2, hand: 2, dredge: true, grave: 2, finalHand: 3, order: []string{"Dredge"}},
		{name: "Chains then dredge", chains: 2, hand: 2, dredge: true, discard: 1, grave: 3, finalHand: 2, order: []string{"that player discards", "Dredge"}},
		{name: "both Chains then dredge", chains: 2, hand: 2, dredge: true, discard: 2, grave: 4, finalHand: 1, order: []string{"that player discards", "that player discards", "Dredge"}},
		{name: "all three replacements", chains: 2, hand: 2, abundance: true, dredge: true, discard: 1, grave: 2, finalHand: 2, order: []string{"that player discards", "choose land or nonland"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			g, e := chainsGame(t, test.chains)
			if test.abundance {
				challengeCastEnchantment(t, g, e, game.Player1, authored.Abundance.New())
			}
			chainsHand(g, test.hand)
			abundanceLibrary(g, "LandA", "LandB", "LandC", "Keep")
			if test.dredge {
				addCardToGraveyard(g, game.Player1, dredgeCardDef(2))
			}
			a := &chainsChoices{order: slices.Clone(test.order)}
			e.drawCards(g, game.Player1, 1, [game.NumPlayers]PlayerAgent{a}, &TurnLog{})
			if len(abundanceEvents(g, game.EventCardDrawn)) != 0 ||
				len(abundanceEvents(g, game.EventCardDiscarded)) != test.discard ||
				g.Players[game.Player1].Graveyard.Size() != test.grave ||
				g.Players[game.Player1].Hand.Size() != test.finalHand {
				t.Fatalf("pending choices/draws/discards/grave/hand = %v/%d/%d/%d/%d", a.order,
					len(abundanceEvents(g, game.EventCardDrawn)), len(abundanceEvents(g, game.EventCardDiscarded)),
					g.Players[game.Player1].Graveyard.Size(), g.Players[game.Player1].Hand.Size())
			}
		})
	}
}

func TestChainsFirstActualDrawExemptionEachOwnDrawStep(t *testing.T) {
	t.Parallel()
	g, e := chainsGame(t, 1)
	challengeCastEnchantment(t, g, e, game.Player1, authored.Abundance.New())
	chainsHand(g, 2)
	top := abundanceLibrary(g, "LandA", "LandB", "LandC", "Keep")
	g.Turn.Step = game.StepDraw
	emitBeginningOfStepEvent(g, game.StepDraw)
	a := &chainsChoices{abundanceChoices: abundanceChoices{accepts: []bool{true, false, false, false}}}
	agents := [game.NumPlayers]PlayerAgent{a}
	e.drawCards(g, game.Player1, 1, agents, &TurnLog{})
	e.drawCards(g, game.Player1, 1, agents, &TurnLog{})
	if !slices.Equal(abundanceEvents(g, game.EventCardDrawn), []id.ID{top[1]}) ||
		len(abundanceEvents(g, game.EventCardDiscarded)) != 0 {
		t.Fatal("putting a card into hand consumed the first actual draw exemption")
	}
	e.drawCards(g, game.Player1, 1, agents, &TurnLog{})
	if len(abundanceEvents(g, game.EventCardDiscarded)) != 1 {
		t.Fatal("a later independent draw retained the exemption")
	}
	emitBeginningOfStepEvent(g, game.StepDraw)
	e.drawCards(g, game.Player1, 1, agents, &TurnLog{})
	if len(abundanceEvents(g, game.EventCardDiscarded)) != 1 || len(abundanceEvents(g, game.EventCardDrawn)) != 3 {
		t.Fatal("an extra draw step failed to start a fresh first-draw exemption")
	}
}

func TestChainsOpponentDrawSourceDepartureCloneAndEmptyLibrary(t *testing.T) {
	t.Parallel()
	g, e := chainsGame(t, 1)
	source := g.Battlefield[0]
	addCardToHand(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Opponent discard"}})
	addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Opponent top"}})
	e.drawCards(g, game.Player2, 1, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if len(abundanceEvents(g, game.EventCardDiscarded)) != 1 ||
		g.Players[game.Player2].Hand.Size() != 1 {
		t.Fatal("Chains was restricted to its controller")
	}
	chainsHand(g, 1)
	cloned := g.Clone()
	e.drawCards(cloned, game.Player1, 1, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if !cloned.FailedDraws[game.Player1] || g.FailedDraws[game.Player1] || g.Players[game.Player1].Hand.Size() != 1 {
		t.Fatal("discard followed by a failed draw, or clone isolation, was lost")
	}
	for _, card := range g.Players[game.Player1].Hand.All() {
		g.Players[game.Player1].Hand.Remove(card)
		g.Players[game.Player1].Graveyard.Add(card)
	}
	e.drawCards(g, game.Player1, 1, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if g.FailedDraws[game.Player1] {
		t.Fatal("milling an empty library caused a failed-draw loss")
	}
	abundanceLibrary(g, "A", "Keep")
	movePermanentToZone(g, source, zone.Graveyard)
	e.drawCards(g, game.Player1, 1, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if g.Players[game.Player1].Hand.Size() != 1 {
		t.Fatal("departed Chains still replaced a draw")
	}
}
