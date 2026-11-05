package rules

import (
	"fmt"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/zone"
)

// Source review: Engine gap #3636. Default-registry cards checked against the
// pinned corpus 6da9ced0…a1b4 (oracle_id): Waterlogged Grove
// 70fa2eba-565e-4fed-adc9-7f5d9fcbf1fa "{T}, Pay 1 life: Add {G} or {U}. / {1},
// {T}, Sacrifice this land: Draw a card."; Sol Ring
// 6ad8011d-3471-4369-9d68-b264cc027487 {1} "{T}: Add {C}{C}."; Elvish Spirit
// Guide 6b0e23cf-7d68-4329-86db-7adc26abd86b "Exile this card from your hand:
// Add {G}."; Reckless Lackey 73733fa5-d292-40ec-8f17-05ba5feb40ab "… create a
// Treasure token. (It's an artifact with "{T}, Sacrifice this token: Add one
// mana of any color.")"; Thalia, Guardian of Thraben
// 9b7f1d05-707c-4ed3-9f0e-8ced1232c2ee "Noncreature spells cost {1} more to
// cast."; Conduit of Worlds ed14be15-8f8d-4fe3-a147-f5da8ed873bf "{T}: Choose
// target nonland permanent card in your graveyard. If you haven't cast a spell
// this turn, you may cast that card. …". Comprehensive Rules effective
// 2026-09-25: 601.2g "If the total cost includes a mana payment, the player then
// has a chance to activate mana abilities"; 605.3a a player may activate a mana
// ability "whenever they are casting a spell … that requires a mana payment";
// 733.1 a reversed cast may reverse legal mana abilities activated while
// casting it, but not one that moved a card to a library. Silent Clearing
// fd45063f-c83c-431c-9104-f139c497ec0d "{T}, Pay 1 life: Add {W} or {B}. …";
// Phyrexian Altar 8d02b297-97c4-4379-9862-0a462400f66f "Sacrifice a creature:
// Add one mana of any color."; Blightsteel Colossus
// e80772e2-8623-4094-81a2-70828b2b151c "If Blightsteel Colossus would be put
// into a graveyard from anywhere, reveal … and shuffle it into its owner's
// library instead."; Future Sight 93a0e938-b916-47ec-9803-369e885e253c "You
// may … cast spells from the top of your library."; Village Rites
// 365548fb-5acc-4a8a-b20b-26d28b7d029f "As an additional cost to cast this
// spell, sacrifice a creature. Draw two cards.";
// 109.1 a token and a permanent are each an object; 111.1 a token represents a
// permanent not represented by a card; 400.2 "Library and hand are hidden
// zones"; 708.2 a face-down permanent has only its face-down characteristics.

const manaWindowPrompt = "Activate a mana ability while paying?"

// manaSourcePickAgent accepts optional casts and, in each payment-time mana
// ability window, activates the next of picks: the option whose source is that
// permanent object or, when the pick isn't on the battlefield, that hand card.
// It records every window request it is shown.
type manaSourcePickAgent struct {
	picks   []id.ID
	windows []game.ChoiceRequest
}

func (*manaSourcePickAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (a *manaSourcePickAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	if request.Kind == game.ChoiceMay {
		return []int{1}
	}
	if request.Prompt != manaWindowPrompt {
		return request.DefaultSelection
	}
	a.windows = append(a.windows, cloneChoiceRequest(request))
	if len(a.picks) == 0 {
		return request.DefaultSelection
	}
	pick := a.picks[0]
	for _, option := range request.Options {
		if option.SourcePermanent == pick || option.SourcePermanent == 0 && option.Card.Exists && option.Card.Val.CardID == pick {
			a.picks = a.picks[1:]
			return []int{option.Index}
		}
	}
	return request.DefaultSelection
}

// requireManaSourceOption finds the window option for source and checks its
// public identity: a permanent's object ID and card instance (zero for a token)
// or a hand card's instance with no permanent.
func requireManaSourceOption(t *testing.T, request game.ChoiceRequest, name string, permanent, card id.ID) game.ChoiceOption {
	t.Helper()
	var found []game.ChoiceOption
	for _, option := range request.Options[1:] {
		if option.SourcePermanent == permanent && option.Card.Exists && option.Card.Val.CardID == card {
			found = append(found, option)
		}
	}
	if len(found) != 1 {
		t.Fatalf("options %+v: found %d for %s (permanent %d, card %d), want 1", request.Options, len(found), name, permanent, card)
	}
	if found[0].Card.Val.Name != name || !strings.HasPrefix(found[0].Label, name) {
		t.Fatalf("option %+v, want it to name %s", found[0], name)
	}
	return found[0]
}

func requireDistinctManaLabels(t *testing.T, request game.ChoiceRequest) {
	t.Helper()
	var labels []string
	for _, option := range request.Options {
		if slices.Contains(labels, option.Label) {
			t.Fatalf("label %q offered twice in %+v", option.Label, request.Options)
		}
		labels = append(labels, option.Label)
	}
}

// requireManaWindowsPrivate checks no payment-time mana window, whose presence
// and option count can reveal a hand card, is logged, and no logged choice
// names the hidden hand card (CR 400.2).
func requireManaWindowsPrivate(t *testing.T, log *TurnLog, hidden string) {
	t.Helper()
	for _, choice := range log.Choices {
		request := choice.Request
		if request.Kind == game.ChoicePayment {
			t.Fatalf("logged mana window %+v, want it kept out of the log", request)
		}
		for _, option := range request.Options {
			if strings.Contains(option.Label, hidden) || option.Card.Exists && option.Card.Val.Name == hidden {
				t.Fatalf("logged option %+v reveals %s in hand", option, hidden)
			}
		}
	}
}

// TestManaWindowLogHidesHandMaskingOptions pays for Sol Ring with {C} already
// floating while Player1 holds one hidden card, either Grizzly Bears or Elvish
// Spirit Guide, on otherwise identical boards, and declines every
// payment-time mana ability window. The Spirit Guide adds an option to a window
// the untapped Grove opens anyway, or opens the only window when the Grove is
// tapped; either way the public log must be identical, so neither a window's
// presence nor its option count reveals the hand card (CR 400.2, 601.2g). This
// holds for a cast while resolving Conduit of Worlds and a cast with priority.
func TestManaWindowLogHidesHandMaskingOptions(t *testing.T) {
	type route struct {
		name string
		cast func(t *testing.T, tapGrove bool, hidden string) (*TurnLog, int)
	}
	routes := []route{
		{name: "resolution cast", cast: func(t *testing.T, tapGrove bool, hidden string) (*TurnLog, int) {
			e, g, cardID := conduitPaidCastGame(t, "Sol Ring")
			if tapGrove {
				tapManaWindowGroves(g)
			}
			addCardToHand(g, game.Player1, registryCardDef(t, hidden))
			g.Players[game.Player1].ManaPool.Add(mana.C, 1)
			agent := &manaSourcePickAgent{}
			log := &TurnLog{}
			e.resolveTopOfStackWithChoices(g, reboundKickerAgents(agent), log)
			requirePaidCast(t, g, cardID, zone.Graveyard, 1)
			return log, len(agent.windows)
		}},
		{name: "priority cast", cast: func(t *testing.T, tapGrove bool, hidden string) (*TurnLog, int) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			addCombatPermanent(g, game.Player1, registryCardDef(t, "Waterlogged Grove"))
			if tapGrove {
				tapManaWindowGroves(g)
			}
			addCardToHand(g, game.Player1, registryCardDef(t, hidden))
			cardID := addCardToHand(g, game.Player1, registryCardDef(t, "Sol Ring"))
			g.Players[game.Player1].ManaPool.Add(mana.C, 1)
			giveMainPhasePriority(g, game.Player1)
			e := NewEngine(nil)
			var cast action.Action
			for _, act := range e.legalActions(g, game.Player1) {
				if castOfCardFrom(zone.Hand)(act, cardID) {
					cast = act
				}
			}
			agent := &manaSourcePickAgent{}
			log := &TurnLog{}
			if outcome, _ := e.attemptActionWithChoices(g, game.Player1, cast, reboundKickerAgents(agent), log, nil); outcome != actionApplied {
				t.Fatalf("outcome = %v, want the cast applied", outcome)
			}
			if obj, ok := g.Stack.Peek(); !ok || obj.SourceID != cardID || obj.SourceZone != zone.Hand || obj.ManaSpentToCast != 1 {
				t.Fatal("Sol Ring isn't on the stack cast from hand with {C} spent")
			}
			return log, len(agent.windows)
		}},
	}
	for _, r := range routes {
		for _, tapGrove := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s grove tapped %t", r.name, tapGrove), func(t *testing.T) {
				bearsLog, bearsWindows := r.cast(t, tapGrove, "Grizzly Bears")
				guideLog, guideWindows := r.cast(t, tapGrove, "Elvish Spirit Guide")
				if guideWindows == 0 || tapGrove != (bearsWindows == 0) {
					t.Fatalf("windows shown: Bears %d, Spirit Guide %d; want the Spirit Guide's shown and the Bears' only with an untapped Grove", bearsWindows, guideWindows)
				}
				requireManaWindowsPrivate(t, guideLog, "Elvish Spirit Guide")
				requirePaymentManaActivationsLogged(t, guideLog)
				if !reflect.DeepEqual(bearsLog.Choices, guideLog.Choices) || !reflect.DeepEqual(bearsLog.Entries, guideLog.Entries) {
					t.Fatalf("public logs differ by hidden hand card:\nBears %+v\nSpirit Guide %+v", bearsLog.Entries, guideLog.Entries)
				}
			})
		}
	}
}

// requirePaymentManaActivationsLogged checks the public log records one action
// entry for each mana ability activated while paying and not reversed, by its
// source, in order (CR 601.2g, 733.1), and that these entries aren't counted
// among the priority actions in Actions. It returns the logged entries.
func requirePaymentManaActivationsLogged(t *testing.T, log *TurnLog, sources ...id.ID) []ActionLog {
	t.Helper()
	var logged []id.ID
	var entries []ActionLog
	for _, entry := range log.Entries {
		if entry.Kind != TurnLogEntryAction || !entry.paymentMana {
			continue
		}
		activation, ok := entry.Action.Action.ActivateAbilityPayload()
		if !ok || !entry.Action.ManaAbility || entry.Action.Player != game.Player1 {
			t.Fatalf("payment-time entry %+v, want a Player1 mana ability activation", entry.Action)
		}
		logged = append(logged, activation.SourceID)
		entries = append(entries, entry.Action)
	}
	if !slices.Equal(logged, sources) {
		t.Fatalf("logged payment-time activations of %v, want %v", logged, sources)
	}
	for _, actionLog := range log.Actions {
		if actionLog.ManaAbility {
			t.Fatalf("Actions has mana ability %+v, want only priority actions", actionLog)
		}
	}
	return entries
}

// castOnceAgent takes cast the first time it is legal and otherwise passes,
// answering choices as its sacrificingAgent does.
type castOnceAgent struct {
	*sacrificingAgent
	cast  action.Action
	casts int
}

func (a *castOnceAgent) ChooseAction(_ PlayerObservation, legal []action.Action) action.Action {
	if a.casts == 0 && actionsContain(legal, a.cast) {
		a.casts++
		return a.cast
	}
	return action.Pass()
}

// TestReversedCastLogsKeptPaymentManaActivations runs the priority loop for a
// player who casts Village Rites off the top of the library with Future Sight
// under Thalia, activates mana abilities while paying (CR 601.2g), and leaves
// the cost unpaid, so the cast is reversed (CR 733.1) and the player then
// passes. The cast and each reversed activation leave the public log, while
// each activation that can't be or isn't reversed stays logged as performed:
//   - Silent Clearing, Phyrexian Altar sacrificing Blightsteel Colossus, which
//     is shuffled into its owner's library, then Waterlogged Grove: only the
//     Altar's activation is kept.
//   - Silent Clearing, then Sacrifice Filter spending the Clearing's mana and
//     sacrificing the Colossus: both are kept.
func TestReversedCastLogsKeptPaymentManaActivations(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name  string
		order []string
		kept  []string
	}{
		{name: "lands around Phyrexian Altar", order: []string{clearingMana, altarMana, groveMana}, kept: []string{"Phyrexian Altar"}},
		{name: "land spent on Sacrifice Filter", order: []string{clearingMana, "Add {C}{C}"}, kept: []string{"Silent Clearing", "Sacrifice Filter"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			g, e, rites, _ := villageRitesLibraryGame(t, "Future Sight", 0)
			addCombatPermanent(g, game.Player2, registryCardDef(t, "Thalia, Guardian of Thraben"))
			for _, name := range []string{"Silent Clearing", "Waterlogged Grove"} {
				addCombatPermanent(g, game.Player1, registryCardDef(t, name))
			}
			addCombatPermanent(g, game.Player1, colossusFilterDef())
			colossus := addAltarAndColossus(t, g, game.Player1)
			cast := action.CastSpellFaceFromZone(rites, zone.Library, game.FaceFront, nil, 0, nil)
			order := &manaOrderAgent{order: test.order, colors: []string{"W", "G"}}
			agent := &castOnceAgent{sacrificingAgent: &sacrificingAgent{next: order, victim: "Blightsteel Colossus"}, cast: cast}
			log := &TurnLog{}

			e.runPriorityLoop(g, reboundKickerAgents(agent), log)

			if agent.casts != 1 || order.activated != len(test.order) || agent.sacrifices != 1 {
				t.Fatalf("casts = %d, activated = %d, sacrifices = %d; want one cast with every activation", agent.casts, order.activated, agent.sacrifices)
			}
			if top, _ := g.Players[game.Player1].Library.Top(); top != rites || !g.Stack.IsEmpty() {
				t.Fatal("Village Rites wasn't returned to the top of the library")
			}
			requireColossusShuffled(t, g, colossus)
			var kept []id.ID
			for _, name := range test.kept {
				kept = append(kept, reboundPermanentNamed(t, g, name).ObjectID)
			}
			requirePaymentManaActivationsLogged(t, log, kept...)
			for _, entry := range log.Entries {
				if entry.Kind == TurnLogEntryAction && entry.Action.Action.Kind == action.ActionCastSpell {
					t.Fatalf("reversed cast logged: %+v", entry.Action)
				}
			}
		})
	}
}

func tapManaWindowGroves(g *game.Game) {
	for _, permanent := range g.Battlefield {
		if permanentChoiceLabel(g, permanent) == "Waterlogged Grove" {
			permanent.Tapped = true
		}
	}
}

// TestResolutionCastManaWindowIdentifiesSources casts Sol Ring through Conduit
// of Worlds with two Waterlogged Groves and Elvish Spirit Guide in hand. Each
// payment-time option (CR 601.2g, 605.3a) names its source: each Grove by its
// own object and card, the Spirit Guide by its hand card. Choosing the second
// Grove activates that Grove, not the first, and the log hides the hand card.
func TestResolutionCastManaWindowIdentifiesSources(t *testing.T) {
	e, g, cardID := conduitPaidCastGame(t, "Sol Ring", "Waterlogged Grove")
	var groves []*game.Permanent
	for _, permanent := range g.Battlefield {
		if permanentChoiceLabel(g, permanent) == "Waterlogged Grove" {
			groves = append(groves, permanent)
		}
	}
	if len(groves) != 2 {
		t.Fatalf("groves = %d, want 2", len(groves))
	}
	guide := addCardToHand(g, game.Player1, registryCardDef(t, "Elvish Spirit Guide"))
	life := g.Players[game.Player1].Life

	agent := &manaSourcePickAgent{picks: []id.ID{groves[1].ObjectID}}
	log := &TurnLog{}
	e.resolveTopOfStackWithChoices(g, reboundKickerAgents(agent), log)

	requirePaidCast(t, g, cardID, zone.Graveyard, 1)
	if len(agent.windows) == 0 {
		t.Fatal("no payment-time mana ability window opened")
	}
	window := agent.windows[0]
	requireDistinctManaLabels(t, window)
	for _, grove := range groves {
		option := requireManaSourceOption(t, window, "Waterlogged Grove", grove.ObjectID, grove.CardInstanceID)
		if !strings.Contains(option.Label, fmt.Sprintf("[%d]", grove.ObjectID)) {
			t.Fatalf("label %q doesn't distinguish Grove %d", option.Label, grove.ObjectID)
		}
	}
	requireManaSourceOption(t, window, "Elvish Spirit Guide", 0, guide)
	if groves[0].Tapped || !groves[1].Tapped || g.Players[game.Player1].Life != life-1 {
		t.Fatalf("first tapped=%t second tapped=%t life=%d, want only the chosen Grove activated", groves[0].Tapped, groves[1].Tapped, g.Players[game.Player1].Life)
	}
	if !g.Players[game.Player1].Hand.Contains(guide) {
		t.Fatal("Elvish Spirit Guide left the hand although not chosen")
	}
	requireManaWindowsPrivate(t, log, "Elvish Spirit Guide")
	if logged := requirePaymentManaActivationsLogged(t, log, groves[1].ObjectID); !strings.Contains(logged[0].AbilityText, "Pay 1 life: Add {G} or {U}") {
		t.Fatalf("logged ability text %q, want the Grove's mana ability", logged[0].AbilityText)
	}
}

// TestPriorityCastManaWindowActivatesChosenSource casts Sol Ring from hand
// with priority, its total {2} with an opponent's Thalia, with two Waterlogged
// Groves, a Treasure token made from Reckless Lackey's token and Elvish Spirit
// Guide in hand. The token is identified by its object with no card (CR 111.1),
// and each chosen source is the one activated while paying (CR 601.2g-h).
func TestPriorityCastManaWindowActivatesChosenSource(t *testing.T) {
	for _, tc := range []struct {
		name  string
		picks func(groves []*game.Permanent, treasure *game.Permanent, guide id.ID) []id.ID
	}{
		{name: "token and hand card", picks: func(_ []*game.Permanent, treasure *game.Permanent, guide id.ID) []id.ID {
			return []id.ID{treasure.ObjectID, guide}
		}},
		{name: "second Grove and token", picks: func(groves []*game.Permanent, treasure *game.Permanent, _ id.ID) []id.ID {
			return []id.ID{groves[1].ObjectID, treasure.ObjectID}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			groves := []*game.Permanent{
				addCombatPermanent(g, game.Player1, registryCardDef(t, "Waterlogged Grove")),
				addCombatPermanent(g, game.Player1, registryCardDef(t, "Waterlogged Grove")),
			}
			treasure, ok := createTokenPermanent(g, game.Player1, recklessLackeyTreasure(t))
			if !ok {
				t.Fatal("creating the Treasure token failed")
			}
			addCombatPermanent(g, game.Player2, registryCardDef(t, "Thalia, Guardian of Thraben"))
			guide := addCardToHand(g, game.Player1, registryCardDef(t, "Elvish Spirit Guide"))
			cardID := addCardToHand(g, game.Player1, registryCardDef(t, "Sol Ring"))
			g.Players[game.Player1].Life = 20
			giveMainPhasePriority(g, game.Player1)
			e := NewEngine(nil)
			cast, offered := action.Action{}, false
			for _, act := range e.legalActions(g, game.Player1) {
				if castOfCardFrom(zone.Hand)(act, cardID) {
					cast, offered = act, true
				}
			}
			if !offered {
				t.Fatal("Sol Ring cast not offered")
			}

			picks := tc.picks(groves, treasure, guide)
			agent := &manaSourcePickAgent{picks: slices.Clone(picks)}
			log := &TurnLog{}
			if outcome, _ := e.attemptActionWithChoices(g, game.Player1, cast, reboundKickerAgents(agent), log, nil); outcome != actionApplied {
				t.Fatalf("outcome = %v, want the cast applied", outcome)
			}
			spellOnStackFrom(zone.Hand, false)(t, g, cardID)
			if len(agent.picks) != 0 {
				t.Fatalf("unactivated picks %v", agent.picks)
			}
			window := agent.windows[0]
			requireDistinctManaLabels(t, window)
			requireManaSourceOption(t, window, "Treasure", treasure.ObjectID, 0)
			requireManaSourceOption(t, window, "Elvish Spirit Guide", 0, guide)
			for _, grove := range groves {
				requireManaSourceOption(t, window, "Waterlogged Grove", grove.ObjectID, grove.CardInstanceID)
			}

			_, treasureRemains := permanentByObjectID(g, treasure.ObjectID)
			guideExiled := g.Players[game.Player1].Exile.Contains(guide)
			lifePaid := 0
			for i, grove := range groves {
				chosen := slices.Contains(picks, grove.ObjectID)
				if grove.Tapped != chosen {
					t.Fatalf("Grove %d tapped=%t, want %t", i, grove.Tapped, chosen)
				}
				if chosen {
					lifePaid++
				}
			}
			if treasureRemains != !slices.Contains(picks, treasure.ObjectID) || guideExiled != slices.Contains(picks, guide) {
				t.Fatalf("Treasure remains=%t guide exiled=%t, want only the picks %v activated", treasureRemains, guideExiled, picks)
			}
			if g.Players[game.Player1].Life != 20-lifePaid {
				t.Fatalf("life = %d, want %d", g.Players[game.Player1].Life, 20-lifePaid)
			}
			hidden := "Elvish Spirit Guide"
			if guideExiled {
				hidden = "no hidden card"
			}
			requireManaWindowsPrivate(t, log, hidden)
			requirePaymentManaActivationsLogged(t, log, picks...)
		})
	}
}

// recklessLackeyTreasure is the Treasure token Reckless Lackey's ability
// creates.
func recklessLackeyTreasure(t *testing.T) *game.CardDef {
	t.Helper()
	for _, ability := range registryCardDef(t, "Reckless Lackey").ActivatedAbilities {
		for _, mode := range ability.Content.Modes {
			for _, instruction := range mode.Sequence {
				if create, ok := instruction.Primitive.(game.CreateToken); ok {
					if def, ok := create.Source.TokenDefRef(); ok {
						return def
					}
				}
			}
		}
	}
	t.Fatal("Reckless Lackey creates no Treasure token")
	return nil
}
