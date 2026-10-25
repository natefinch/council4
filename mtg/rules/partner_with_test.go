package rules

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review: Brallin, Skyshark Rider C20 4 printing
// 9b4fdcd5-5346-46a8-b1a4-84ddb74089e0 (oracle 464df46a-079b-47c1-95e7-930160abef2f);
// https://api.scryfall.com/cards/9b4fdcd5-5346-46a8-b1a4-84ddb74089e0/rulings
// returned eleven WotC rulings dated 2020-04-17. They spell out the full
// trigger ("target player may search their library for a card named [name],
// reveal it, put it into their hand, then shuffle"), that the target player
// searches their own library and reveals the found card even though the
// reminder text omits it, and that another player may be targeted. CR
// 2026-09-25 from the verified official copy: 702.124j defines "Partner with
// [name]" as the deck-construction permission plus "When this permanent enters,
// target player may search their library for a card named [name], reveal it,
// put it into their hand, then shuffle."; 701.23b lets a hidden-zone search for
// a stated quality fail to find; 603.6a covers the enters trigger.

// partnerWithAgent picks a player target, answers the target player's may
// decision, and finds the wanted card by name when it is offered. It records
// every choice it is asked to make.
type partnerWithAgent struct {
	simPassPolicy

	target   game.PlayerID
	accept   bool
	wanted   string
	requests []game.ChoiceRequest
}

func (a *partnerWithAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	a.requests = append(a.requests, request)
	switch request.Kind {
	case game.ChoiceMay:
		if a.accept {
			return []int{1}
		}
		return []int{0}
	case game.ChoiceTarget:
		for _, option := range request.Options {
			if slices.Contains(option.Targets, game.PlayerTarget(a.target)) {
				return []int{option.Index}
			}
		}
		return request.DefaultSelection
	default:
		for _, option := range request.Options {
			if a.wanted != "" && option.Label == a.wanted {
				return []int{option.Index}
			}
		}
		return []int{}
	}
}

func (a *partnerWithAgent) kinds() []game.ChoiceKind {
	var kinds []game.ChoiceKind
	for _, request := range a.requests {
		kinds = append(kinds, request.Kind)
	}
	return kinds
}

func partnerWithAgents(controller, other *partnerWithAgent) [game.NumPlayers]PlayerAgent {
	agents := simPassPolicies()
	agents[game.Player1] = controller
	if other != nil {
		agents[game.Player2] = other
	}
	return agents
}

func partnerWithFiller(g *game.Game, player game.PlayerID, count int) {
	for i := range count {
		addCardToLibrary(g, player, &game.CardDef{CardFace: game.CardFace{
			Name:  fmt.Sprintf("Partner filler %d", i),
			Types: []types.Card{types.Creature},
		}})
	}
}

func lookupPartnerWithCard(t *testing.T, registry *cards.Registry, name string) *game.CardDef {
	t.Helper()
	def, err := registry.LookupChecked(name)
	if err != nil {
		t.Fatalf("default registry lookup of %q failed: %v", name, err)
	}
	if def.Name != name {
		t.Fatalf("lookup of %q returned %q", name, def.Name)
	}
	return def
}

// castBrallinAndTrigger casts the production Brallin from Player1's hand,
// resolves it, and puts its partner-with enters trigger on the stack with
// Player1's agent choosing the trigger's target player.
func castBrallinAndTrigger(t *testing.T, g *game.Game, engine *Engine, agents [game.NumPlayers]PlayerAgent) *game.Permanent {
	t.Helper()
	registry := cards.NewDefaultRegistry()
	brallin := addCardToHand(g, game.Player1, lookupPartnerWithCard(t, registry, "Brallin, Skyshark Rider"))
	g.Players[game.Player1].ManaPool.Add(mana.R, 4)
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.CastSpell(brallin, nil, 0, nil)) {
		t.Fatal("casting Brallin failed")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	var permanent *game.Permanent
	for _, candidate := range g.Battlefield {
		if candidate.CardInstanceID == brallin {
			permanent = candidate
		}
	}
	if permanent == nil {
		t.Fatal("Brallin did not enter the battlefield")
	}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{}) {
		t.Fatal("Brallin's partner-with enters trigger was not put on the stack")
	}
	trigger, ok := g.Stack.Peek()
	if !ok || trigger.Kind != game.StackTriggeredAbility || trigger.SourceID != permanent.ObjectID {
		t.Fatal("top of stack is not Brallin's enters trigger")
	}
	return permanent
}

func revealedCards(g *game.Game, player game.PlayerID) []id.ID {
	var revealed []id.ID
	for _, event := range g.Events {
		if event.Kind == game.EventCardRevealed && event.Player == player {
			revealed = append(revealed, event.CardID)
		}
	}
	return revealed
}

// TestPartnerWithBrallinFetchesShabrazForController proves the production
// Brallin's enters trigger: its controller targets themself, accepts, finds the
// card named Shabraz, the Skyshark among other cards, reveals it, puts it into
// hand, and shuffles the library.
func TestPartnerWithBrallinFetchesShabrazForController(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	registry := cards.NewDefaultRegistry()
	partnerWithFiller(g, game.Player1, 10)
	shabraz := addCardToLibrary(g, game.Player1, lookupPartnerWithCard(t, registry, "Shabraz, the Skyshark"))
	partnerWithFiller(g, game.Player1, 10)
	agent := &partnerWithAgent{target: game.Player1, accept: true, wanted: "Shabraz, the Skyshark"}
	agents := partnerWithAgents(agent, nil)

	castBrallinAndTrigger(t, g, engine, agents)
	before := slices.Clone(g.Players[game.Player1].Library.All())
	before = slices.DeleteFunc(before, func(cardID id.ID) bool { return cardID == shabraz })
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	player := g.Players[game.Player1]
	if !player.Hand.Contains(shabraz) || player.Library.Contains(shabraz) {
		t.Fatal("accepted partner search did not move Shabraz from library to hand")
	}
	if got := revealedCards(g, game.Player1); !slices.Equal(got, []id.ID{shabraz}) {
		t.Fatalf("revealed cards = %v, want only Shabraz", got)
	}
	if countEventsOfKind(g.Events, game.EventLibrarySearched) != 1 {
		t.Fatal("accepted partner search did not search the library exactly once")
	}
	if slices.Equal(player.Library.All(), before) {
		t.Fatal("library was not shuffled after the partner search")
	}
	if !slices.Equal(agent.kinds(), []game.ChoiceKind{game.ChoiceTarget, game.ChoiceMay, game.ChoiceSearch}) {
		t.Fatalf("controller choices = %v, want target, may, search", agent.kinds())
	}
	search := agent.requests[2]
	if len(search.Options) != 1 || search.Options[0].Label != "Shabraz, the Skyshark" {
		t.Fatalf("search offered %#v, want only the card named Shabraz, the Skyshark", search.Options)
	}
	if !g.Stack.IsEmpty() {
		t.Fatal("stack is not empty after the partner trigger resolved")
	}
}

// TestPartnerWithTargetOpponentSearchesOwnLibrary proves the target player, not
// the trigger's controller, decides and searches their own library, even after
// the source has left the battlefield.
func TestPartnerWithTargetOpponentSearchesOwnLibrary(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	registry := cards.NewDefaultRegistry()
	shabrazDef := lookupPartnerWithCard(t, registry, "Shabraz, the Skyshark")
	controllerShabraz := addCardToLibrary(g, game.Player1, shabrazDef)
	opponentShabraz := addCardToLibrary(g, game.Player2, shabrazDef)
	partnerWithFiller(g, game.Player2, 5)
	controller := &partnerWithAgent{target: game.Player2, accept: false, wanted: "Shabraz, the Skyshark"}
	opponent := &partnerWithAgent{accept: true, wanted: "Shabraz, the Skyshark"}
	agents := partnerWithAgents(controller, opponent)

	brallin := castBrallinAndTrigger(t, g, engine, agents)
	addEffectSpellToStack(g, game.Player2, game.MovePermanent{
		Object:      game.TargetPermanentReference(0),
		Destination: zone.Graveyard,
	}, []game.Target{game.PermanentTarget(brallin.ObjectID)})
	engine.resolveTopOfStack(g, &TurnLog{})
	if _, ok := permanentByObjectID(g, brallin.ObjectID); ok {
		t.Fatal("setup effect did not remove Brallin")
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if !g.Players[game.Player2].Hand.Contains(opponentShabraz) || g.Players[game.Player2].Library.Contains(opponentShabraz) {
		t.Fatal("target opponent did not put their own Shabraz into their hand")
	}
	if !g.Players[game.Player1].Library.Contains(controllerShabraz) || g.Players[game.Player1].Hand.Contains(controllerShabraz) {
		t.Fatal("trigger's controller searched their own library")
	}
	if got := revealedCards(g, game.Player2); !slices.Equal(got, []id.ID{opponentShabraz}) {
		t.Fatalf("opponent revealed %v, want only their Shabraz", got)
	}
	if !slices.Equal(controller.kinds(), []game.ChoiceKind{game.ChoiceTarget}) {
		t.Fatalf("controller choices = %v, want only the target choice", controller.kinds())
	}
	if !slices.Equal(opponent.kinds(), []game.ChoiceKind{game.ChoiceMay, game.ChoiceSearch}) {
		t.Fatalf("opponent choices = %v, want may then search", opponent.kinds())
	}
	for _, request := range opponent.requests {
		if request.Player != game.Player2 {
			t.Fatalf("choice %v was offered to %v, want the target player", request.Kind, request.Player)
		}
	}
}

// TestPartnerWithDeclineAndFailToFind proves the target player may decline the
// search entirely (no search, no reveal, no shuffle) or search and fail to find
// the named card although it is present (CR 701.23b).
func TestPartnerWithDeclineAndFailToFind(t *testing.T) {
	for _, test := range []struct {
		name         string
		accept       bool
		wantSearched int
		wantShuffled bool
	}{
		{name: "declined"},
		{name: "fail to find", accept: true, wantSearched: 1, wantShuffled: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			registry := cards.NewDefaultRegistry()
			partnerWithFiller(g, game.Player1, 10)
			shabraz := addCardToLibrary(g, game.Player1, lookupPartnerWithCard(t, registry, "Shabraz, the Skyshark"))
			partnerWithFiller(g, game.Player1, 10)
			agent := &partnerWithAgent{target: game.Player1, accept: test.accept}
			agents := partnerWithAgents(agent, nil)

			castBrallinAndTrigger(t, g, engine, agents)
			before := slices.Clone(g.Players[game.Player1].Library.All())
			engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

			player := g.Players[game.Player1]
			if !player.Library.Contains(shabraz) || player.Hand.Contains(shabraz) {
				t.Fatal("Shabraz left the library")
			}
			if got := countEventsOfKind(g.Events, game.EventLibrarySearched); got != test.wantSearched {
				t.Fatalf("library searches = %d, want %d", got, test.wantSearched)
			}
			if len(revealedCards(g, game.Player1)) != 0 {
				t.Fatal("a card was revealed without being found")
			}
			if shuffled := !slices.Equal(player.Library.All(), before); shuffled != test.wantShuffled {
				t.Fatalf("library shuffled = %v, want %v", shuffled, test.wantShuffled)
			}
			if test.accept {
				search := agent.requests[len(agent.requests)-1]
				if search.Kind != game.ChoiceSearch || search.MinChoices != 0 {
					t.Fatalf("search request = %#v, want an optional search", search)
				}
			}
		})
	}
}

// TestPartnerWithTriggerOnNoncreaturePermanent proves the trigger is a "this
// permanent enters" trigger (CR 702.124j), so a partner-with planeswalker
// (Rowan Kenrith, Will Kenrith) fetches its partner too.
func TestPartnerWithTriggerOnNoncreaturePermanent(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	walker := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:               "Partner Walker",
		Supertypes:         []types.Super{types.Legendary},
		Types:              []types.Card{types.Planeswalker},
		StaticAbilities:    []game.StaticAbility{game.PartnerWithStaticBody},
		TriggeredAbilities: []game.TriggeredAbility{game.PartnerWithTriggeredAbility("Partner Sibling")},
	}})
	sibling := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:       "Partner Sibling",
		Supertypes: []types.Super{types.Legendary},
		Types:      []types.Card{types.Planeswalker},
	}})
	agent := &partnerWithAgent{target: game.Player1, accept: true, wanted: "Partner Sibling"}
	agents := partnerWithAgents(agent, nil)

	emitEvent(g, game.Event{Kind: game.EventPermanentEnteredBattlefield, PermanentID: walker.ObjectID})
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{}) {
		t.Fatal("planeswalker's partner-with trigger was not put on the stack")
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	if !g.Players[game.Player1].Hand.Contains(sibling) {
		t.Fatal("planeswalker's partner-with trigger did not fetch its partner")
	}
}

// TestPartnerWithRegisteredCardsFetchTheirPartner enters every registered
// "Partner with" card and proves its trigger fetches exactly the card its
// keyword names, and that the name agrees with the commander pairing half.
func TestPartnerWithRegisteredCardsFetchTheirPartner(t *testing.T) {
	registry := cards.NewDefaultRegistry()
	pairs := []struct{ card, partner string }{
		{"Alphinaud Leveilleur", "Alisaie Leveilleur"},
		{"Blaring Recruiter", "Blaring Captain"},
		{"Brallin, Skyshark Rider", "Shabraz, the Skyshark"},
		{"Cazur, Ruthless Stalker", "Ukkima, Stalking Shadow"},
		{"Chakram Retriever", "Chakram Slinger"},
		{"Chakram Slinger", "Chakram Retriever"},
		{"Evie Frye", "Jacob Frye"},
		{"Impetuous Protege", "Proud Mentor"},
		{"Jenny Flint", "Madame Vastra"},
		{"Kamber, the Plunderer", "Laurine, the Diversion"},
		{"Krav, the Unredeemed", "Regna, the Redeemer"},
		{"Laurine, the Diversion", "Kamber, the Plunderer"},
		{"Ley Weaver", "Lore Weaver"},
		{"Lore Weaver", "Ley Weaver"},
		{"Madame Vastra", "Jenny Flint"},
		{"Merry, Warden of Isengard", "Pippin, Warden of Isengard"},
		{"Pippin, Warden of Isengard", "Merry, Warden of Isengard"},
		{"Proud Mentor", "Impetuous Protege"},
		{"Rocksteady, Mutant Marauder", "Bebop, Skull & Crossbones"},
		{"Shabraz, the Skyshark", "Brallin, Skyshark Rider"},
		{"Silvar, Devourer of the Free", "Trynn, Champion of Freedom"},
		{"Soulblade Corrupter", "Soulblade Renewer"},
		{"Timin, Youthful Geist", "Rhoda, Geist Avenger"},
		{"Toothy, Imaginary Friend", "Pir, Imaginative Rascal"},
		{"Trynn, Champion of Freedom", "Silvar, Devourer of the Free"},
		{"Ukkima, Stalking Shadow", "Cazur, Ruthless Stalker"},
		{"Virtus the Veiled", "Gorm the Great"},
	}
	for _, pair := range pairs {
		t.Run(pair.card, func(t *testing.T) {
			def := lookupPartnerWithCard(t, registry, pair.card)
			if !def.HasKeyword(game.PartnerWith) {
				t.Fatal("card does not have the PartnerWith keyword")
			}
			if !strings.Contains(def.OracleText, "Partner with "+pair.partner) {
				t.Fatalf("oracle text %q does not name partner %q", def.OracleText, pair.partner)
			}
			// Nonlegendary partner-with cards (Battlebond's Proud Mentor) can't be
			// commanders, so only legendary ones have a commander pairing half.
			if got := partnerWithName(def); isLegendaryCreature(def) && got != pair.partner {
				t.Fatalf("commander partner name = %q, want %q", got, pair.partner)
			}
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			permanent := addCombatPermanent(g, game.Player1, def)
			partnerWithFiller(g, game.Player1, 3)
			partner := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
				Name:       pair.partner,
				Supertypes: []types.Super{types.Legendary},
				Types:      []types.Card{types.Creature},
				Power:      opt.Val(game.PT{Value: 1}),
				Toughness:  opt.Val(game.PT{Value: 1}),
			}})
			agent := &partnerWithAgent{target: game.Player1, accept: true, wanted: pair.partner}
			agents := partnerWithAgents(agent, nil)

			emitEvent(g, game.Event{Kind: game.EventPermanentEnteredBattlefield, PermanentID: permanent.ObjectID})
			if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{}) {
				t.Fatal("no enters trigger was put on the stack")
			}
			for !g.Stack.IsEmpty() {
				engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
			}
			if !g.Players[game.Player1].Hand.Contains(partner) {
				t.Fatalf("entering %s did not fetch %s", pair.card, pair.partner)
			}
			if got := g.Players[game.Player1].Hand.Size(); got != 1 {
				t.Fatalf("hand size = %d, want only the partner", got)
			}
		})
	}
}
