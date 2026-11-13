package rules

import (
	"encoding/json"
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

// Source review: Copper Gnomes USG 291 printing
// d5e326b7-6f6a-4249-a315-c5f017931c73 (oracle 4fd6d9af-4d53-4e10-bc23-672454803099),
// {2} Artifact Creature — Gnome, Oracle "{4}, Sacrifice this creature: You may
// put an artifact card from your hand onto the battlefield." from the pinned
// corpus; https://api.scryfall.com/cards/d5e326b7-6f6a-4249-a315-c5f017931c73/rulings
// returned 2004-10-04 (the artifact card is chosen on resolution) and
// 2007-05-01 (putting the card onto the battlefield is optional). Comprehensive
// Rules effective 2026-11-06 (SHA-256
// 524f5195a1e5f54bb13c735edf399ac717a4458e2713f398d1a3e6877186b829): 400.2
// (library and hand are hidden zones), 401.2 (a library is a face-down pile),
// 402.3 (a player can't look at another player's hand but may count it) and
// 110.2a (the put card enters under that player's control). These scenarios
// prove the single-chooser ChooseFromZone hidden-zone privacy fix (#3840).

// hiddenZonePicker records every request and selects, up to MaxChoices, the
// offered cards named in picks, accepting optional effects; otherwise it takes
// the default selection.
type hiddenZonePicker struct {
	simPassPolicy

	picks    []string
	requests []game.ChoiceRequest
}

func (a *hiddenZonePicker) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	a.requests = append(a.requests, request)
	var selected []int
	for _, option := range request.Options {
		if len(selected) == request.MaxChoices {
			break
		}
		if option.Card.Exists && slices.Contains(a.picks, option.Card.Val.Name) ||
			request.Kind == game.ChoiceMay && option.Label == "Yes" {
			selected = append(selected, option.Index)
		}
	}
	if len(selected) == 0 {
		return request.DefaultSelection
	}
	return selected
}

// assertNoHiddenNames fails when the returned game log names any hidden card.
func assertNoHiddenNames(t *testing.T, log *TurnLog, hidden []string) {
	t.Helper()
	data, err := json.Marshal(GameResult{Turns: []TurnLog{*log}})
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range hidden {
		if strings.Contains(string(data), name) {
			t.Fatalf("returned game log reveals hidden card %q", name)
		}
	}
}

// TestCopperGnomesHandChoiceStaysOutOfThePublicLog activates the production
// Copper Gnomes and puts an artifact card from its controller's hand onto the
// battlefield. The controller is offered every artifact card in their hand, but
// the public log records neither the offered cards nor how many there were
// (CR 400.2, 402.3); only the artifact that entered becomes public.
func TestCopperGnomesHandChoiceStaysOutOfThePublicLog(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	setMainPhasePriority(g, game.Player1)
	gnomes := addCombatPermanent(g, game.Player1, registryCardDef(t, "Copper Gnomes"))
	relic := addCardToHand(g, game.Player1, chooseFromZoneCardDef("Secret Relic", 3, []types.Card{types.Artifact}, nil))
	bauble := addCardToHand(g, game.Player1, chooseFromZoneCardDef("Secret Bauble", 1, []types.Card{types.Artifact}, nil))
	addCardToHand(g, game.Player1, chooseFromZoneCardDef("Secret Giant", 7, []types.Card{types.Creature}, nil))
	g.Players[game.Player1].ManaPool.Add(mana.C, 4)
	if !engine.applyAction(g, game.Player1, action.ActivateAbility(gnomes.ObjectID, 0, nil, 0)) {
		t.Fatal("Copper Gnomes activation failed")
	}
	if onBattlefieldByCard(g, gnomes.CardInstanceID) {
		t.Fatal("Copper Gnomes was not sacrificed to activate")
	}

	p1 := &hiddenZonePicker{picks: []string{"Secret Relic"}}
	agents := simPassPolicies()
	agents[game.Player1] = p1
	log := &TurnLog{}
	engine.resolveTopOfStackWithChoices(g, agents, log)

	var handChoice *game.ChoiceRequest
	for i := range p1.requests {
		if p1.requests[i].Kind == game.ChoiceResolution {
			handChoice = &p1.requests[i]
		}
	}
	if handChoice == nil || !handChoice.Private || len(handChoice.Options) != 2 {
		t.Fatalf("hand choice = %#v, want a private choice between both artifact cards", handChoice)
	}
	if permanent, ok := reanimatedPermanent(g, relic); !ok || permanent.Controller != game.Player1 {
		t.Fatal("the chosen artifact did not enter under its controller's control")
	}
	if !g.Players[game.Player1].Hand.Contains(bauble) {
		t.Fatal("the unchosen artifact left the hand")
	}
	for _, decision := range log.Choices {
		if decision.Request.Kind == game.ChoiceResolution {
			t.Fatalf("public log recorded the hidden hand choice %#v", decision)
		}
	}
	assertNoHiddenNames(t, log, []string{"Secret Bauble", "Secret Giant"})
}

// TestSingleChooserHiddenZoneChoicesStayOutOfThePublicLog verifies that every
// single-chooser grouping route keeps a hidden-zone decision out of the public
// log entirely (CR 400.2, 401.2, 402.3): the options, their count and the
// selection, including the split destination's follow-up prompts whose cards
// may go to a hidden zone. Only the cards moved to a public zone become public.
func TestSingleChooserHiddenZoneChoicesStayOutOfThePublicLog(t *testing.T) {
	creature := []types.Card{types.Creature}
	land := []types.Card{types.Land}
	splitToHand := opt.Val(game.ChooseSplitSlot{Destination: game.ChooseDestination{Zone: zone.Hand}})
	cases := []struct {
		name     string
		source   zone.Type
		cards    []*game.CardDef
		picks    []string
		env      game.ChooseFromZone
		entered  []string
		toHand   []string
		unchosen []string
	}{{
		name:   "across the set",
		source: zone.Hand,
		cards: []*game.CardDef{
			chooseFromZoneCardDef("Secret Elf", 1, creature, []types.Sub{types.Elf}),
			chooseFromZoneCardDef("Secret Giant", 7, creature, []types.Sub{types.Giant}),
			chooseFromZoneCardDef("Secret Spell", 1, []types.Card{types.Instant}, nil),
		},
		picks:    []string{"Secret Elf"},
		env:      game.ChooseFromZone{Filter: game.Selection{RequiredTypes: creature}, Quantity: game.Fixed(1)},
		entered:  []string{"Secret Elf"},
		unchosen: []string{"Secret Giant", "Secret Spell"},
	}, {
		name:   "shared subtype",
		source: zone.Hand,
		cards: []*game.CardDef{
			chooseFromZoneCardDef("Secret Elf", 1, creature, []types.Sub{types.Elf}),
			chooseFromZoneCardDef("Secret Druid", 2, creature, []types.Sub{types.Elf, types.Druid}),
			chooseFromZoneCardDef("Secret Giant", 7, creature, []types.Sub{types.Giant}),
		},
		picks: []string{"Secret Elf", "Secret Druid"},
		env: game.ChooseFromZone{Filter: game.Selection{RequiredTypes: creature}, Quantity: game.Fixed(2),
			Grouping: game.ChooseSharedSubtype},
		entered:  []string{"Secret Elf", "Secret Druid"},
		unchosen: []string{"Secret Giant"},
	}, {
		name:   "one of each named type",
		source: zone.Hand,
		cards: []*game.CardDef{
			chooseFromZoneCardDef("Secret Elf", 1, creature, nil),
			chooseFromZoneCardDef("Secret Giant", 7, creature, nil),
			chooseFromZoneCardDef("Secret Relic", 3, []types.Card{types.Artifact}, nil),
		},
		picks: []string{"Secret Elf", "Secret Relic"},
		env: game.ChooseFromZone{Filter: game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Artifact}},
			Quantity: game.Fixed(2), Grouping: game.ChooseOneOfEachNamedType},
		entered:  []string{"Secret Elf", "Secret Relic"},
		unchosen: []string{"Secret Giant"},
	}, {
		name:   "split destination, two cards",
		source: zone.Library,
		cards: []*game.CardDef{
			chooseFromZoneCardDef("Secret Forest", 0, land, nil),
			chooseFromZoneCardDef("Secret Mountain", 0, land, nil),
			chooseFromZoneCardDef("Secret Island", 0, land, nil),
		},
		picks: []string{"Secret Forest", "Secret Mountain"},
		env: game.ChooseFromZone{Filter: game.Selection{RequiredTypes: land}, Quantity: game.Fixed(2),
			Grouping: game.ChooseSplitDestination, SplitSecondary: splitToHand},
		entered:  []string{"Secret Forest"},
		toHand:   []string{"Secret Mountain"},
		unchosen: []string{"Secret Island"},
	}, {
		name:   "split destination, one card",
		source: zone.Library,
		cards: []*game.CardDef{
			chooseFromZoneCardDef("Secret Forest", 0, land, nil),
			chooseFromZoneCardDef("Secret Island", 0, land, nil),
		},
		picks: []string{"Secret Forest"},
		env: game.ChooseFromZone{Filter: game.Selection{RequiredTypes: land}, Quantity: game.Fixed(2),
			Grouping: game.ChooseSplitDestination, SplitSecondary: splitToHand},
		entered:  []string{"Secret Forest"},
		unchosen: []string{"Secret Island"},
	}}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			obj := chooseFromZoneSource(g)
			byName := map[string]id.ID{}
			for _, def := range tc.cards {
				if tc.source == zone.Library {
					byName[def.Name] = addCfzLibraryCard(g, game.Player1, def)
				} else {
					byName[def.Name] = addCardToHand(g, game.Player1, def)
				}
			}
			p1 := &hiddenZonePicker{picks: tc.picks}
			agents := simPassPolicies()
			agents[game.Player1] = p1
			log := &TurnLog{}
			env := tc.env
			env.Player = game.ControllerReference()
			env.SourceZone = tc.source
			env.Count = game.ChooseUpTo
			env.Destination = game.ChooseDestination{Zone: zone.Battlefield}
			newEffectResolver(NewEngine(nil), g, obj, agents, log).resolveChooseFromZone(env)

			if len(p1.requests) == 0 {
				t.Fatal("the chooser was not asked to choose")
			}
			for _, request := range p1.requests {
				if !request.Private {
					t.Fatalf("hidden-zone request %q is not private", request.Prompt)
				}
			}
			if len(log.Choices) != 0 {
				t.Fatalf("public log recorded hidden-zone decisions %#v", log.Choices)
			}
			for _, name := range tc.entered {
				if !onBattlefieldByCard(g, byName[name]) {
					t.Fatalf("chosen card %s did not enter the battlefield", name)
				}
			}
			for _, name := range tc.toHand {
				if !g.Players[game.Player1].Hand.Contains(byName[name]) {
					t.Fatalf("chosen card %s did not go to the hand", name)
				}
			}
			for _, name := range tc.unchosen {
				if onBattlefieldByCard(g, byName[name]) || tc.source == zone.Library && !g.Players[game.Player1].Library.Contains(byName[name]) {
					t.Fatalf("unchosen card %s left its zone", name)
				}
			}
			assertNoHiddenNames(t, log, append(slices.Clone(tc.unchosen), tc.toHand...))
		})
	}
}

// TestSingleChooserPublicZoneChoiceStaysLogged verifies that a choice from a
// public zone is still recorded in the public log with its cards, since every
// player can already see them (CR 400.2).
func TestSingleChooserPublicZoneChoiceStaysLogged(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	obj := chooseFromZoneSource(g)
	addCfzGraveyardCard(g, game.Player1, chooseFromZoneCardDef("Open Elf", 1, []types.Card{types.Creature}, nil))
	addCfzGraveyardCard(g, game.Player1, chooseFromZoneCardDef("Open Giant", 7, []types.Card{types.Creature}, nil))
	agents := simPassPolicies()
	agents[game.Player1] = &hiddenZonePicker{picks: []string{"Open Giant"}}
	log := &TurnLog{}
	newEffectResolver(NewEngine(nil), g, obj, agents, log).resolveChooseFromZone(game.ChooseFromZone{
		Player:      game.ControllerReference(),
		SourceZone:  zone.Graveyard,
		Filter:      game.Selection{RequiredTypes: []types.Card{types.Creature}},
		Quantity:    game.Fixed(1),
		Count:       game.ChooseUpTo,
		Destination: game.ChooseDestination{Zone: zone.Battlefield},
	})

	if len(log.Choices) != 1 || log.Choices[0].Request.Private || len(log.Choices[0].Request.Options) != 2 ||
		log.Choices[0].Request.Options[0].Label == "Private option" {
		t.Fatalf("public-zone choice log = %#v, want one public entry naming both cards", log.Choices)
	}
}
