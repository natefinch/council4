package rules

import (
	"fmt"
	"testing"

	"github.com/natefinch/council4/internal/carddata"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func TestCompiledActualCardQuantityResults(t *testing.T) {
	for _, tc := range []struct {
		name, text          string
		source              zone.Type
		owner               game.PlayerID
		matchType, missType types.Card
	}{
		{"mill", "Mill three cards. For each creature card milled this way, create a 1/1 green Insect creature token.", zone.Library, game.Player1, types.Creature, types.Land},
		{"discard", "Discard three cards. For each nonland card discarded this way, create a 1/1 red Human creature token.", zone.Hand, game.Player1, types.Creature, types.Land},
		{"zone group", "Exile target player's graveyard. For each creature card exiled this way, create a 2/2 black Zombie creature token.", zone.Graveyard, game.Player2, types.Creature, types.Land},
		{"choose batch", "Exile three cards from your graveyard. For each creature card exiled this way, create a 2/2 black Zombie creature token.", zone.Graveyard, game.Player1, types.Creature, types.Land},
		{"artifact-land union", "Exile target player's graveyard. For each artifact or land card exiled this way, create a 1/1 colorless Insect artifact creature token with flying.", zone.Graveyard, game.Player2, types.Artifact, types.Creature},
		{"opponent graveyard", "Exile target opponent's graveyard. For each creature card exiled this way, create a 1/1 colorless Spirit creature token.", zone.Graveyard, game.Player2, types.Creature, types.Land},
		{"mill creature artifact payload", "Mill three cards. For each creature card milled this way, create a 2/2 black Necron Warrior artifact creature token.", zone.Library, game.Player1, types.Creature, types.Land},
		{"two-card mill", "Mill two cards. For each creature card milled this way, create a 1/1 green Saproling creature token.", zone.Library, game.Player1, types.Creature, types.Land},
		{"discard intervening draw", "Discard two cards, then draw two cards. For each nonland card discarded this way, create a 1/1 red Elemental creature token.", zone.Hand, game.Player1, types.Creature, types.Land},
	} {
		for _, qualifying := range []int{0, 2} {
			t.Run(fmt.Sprintf("%s/%d", tc.name, qualifying), func(t *testing.T) {
				def := legacyCard(t, carddata.ScryfallCard{
					Name: "Quantity", Layout: "normal", TypeLine: "Sorcery", OracleText: tc.text,
				})
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				count := 3
				if tc.name == "two-card mill" || tc.name == "discard intervening draw" {
					count = 2
				}
				for i := range count {
					cardType := tc.missType
					if i < qualifying {
						cardType = tc.matchType
					}
					card := &game.CardDef{CardFace: game.CardFace{Name: fmt.Sprintf("Member%d", i), Types: []types.Card{cardType}}}
					switch tc.source {
					case zone.Library:
						addCardToLibrary(g, tc.owner, card)
					case zone.Hand:
						addCardToHand(g, tc.owner, card)
					case zone.Graveyard:
						addCardToGraveyard(g, tc.owner, card)
					default:
						t.Fatal("unmodeled test source")
					}
				}
				obj := &game.StackObject{Controller: game.Player1}
				if tc.owner == game.Player2 {
					obj.Targets = []game.Target{game.PlayerTarget(game.Player2)}
				}
				NewEngine(nil).resolveAbilityContentWithChoices(g, obj, def.SpellAbility.Val, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
				if got := len(g.Battlefield); got != qualifying {
					t.Fatalf("tokens=%d, want actual qualifying count %d (requested 3)", got, qualifying)
				}
				for _, token := range g.Battlefield {
					if !token.Token || token.Controller != game.Player1 || token.Owner != game.Player1 {
						t.Fatalf("token followed producer's physical owner instead of consumer: %#v", token)
					}
				}
			})
		}
	}
}

func TestCompiledActualCardQuantitySingleTargetIncarnation(t *testing.T) {
	def := legacyCard(t, carddata.ScryfallCard{
		Name: "Target history", Layout: "normal", TypeLine: "Instant",
		OracleText: "Exile target card from a graveyard. For each creature card exiled this way, you gain 2 life.",
	})
	for _, stale := range []bool{false, true} {
		t.Run(fmt.Sprintf("stale=%t", stale), func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			cardID := addCardToGraveyard(g, game.Player2, vanillaCreature("Target", 1, 1))
			target := currentCardTarget(t, g, cardID)
			if stale {
				if !moveCardBetweenZones(g, game.Player2, cardID, zone.Graveyard, zone.Hand) ||
					!moveCardBetweenZones(g, game.Player2, cardID, zone.Hand, zone.Graveyard) {
					t.Fatal("failed to create a different card incarnation")
				}
			}
			obj := &game.StackObject{Controller: game.Player1, Targets: []game.Target{target}}
			before := g.Players[game.Player1].Life
			NewEngine(nil).resolveAbilityContentWithChoices(g, obj, def.SpellAbility.Val, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
			want := before + 2
			if stale {
				want = before
			}
			if got := g.Players[game.Player1].Life; got != want {
				t.Fatalf("life=%d, want %d; actual move borrowed a different target incarnation", got, want)
			}
		})
	}
}

func TestActualCardQuantityCountsCommanderBeforeCommandZoneSBA(t *testing.T) {
	def := legacyCard(t, carddata.ScryfallCard{
		Name: "Exile destination", Layout: "normal", TypeLine: "Sorcery",
		OracleText: "Exile target player's graveyard. For each creature card exiled this way, create a 2/2 black Zombie creature token.",
	})
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	cardID := addCardToGraveyard(g, game.Player2, vanillaCreature("Commander", 1, 1))
	g.CommanderIDs[cardID] = true
	g.Players[game.Player2].CommanderInstanceID = cardID
	obj := &game.StackObject{Controller: game.Player1, Targets: []game.Target{game.PlayerTarget(game.Player2)}}
	NewEngine(nil).resolveAbilityContentWithChoices(g, obj, def.SpellAbility.Val, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if !g.Players[game.Player2].Exile.Contains(cardID) || len(g.Battlefield) != 1 {
		t.Fatal("commander's real exile arrival did not count for its payoff")
	}
	NewEngine(nil).applyStateBasedActions(g)
	if !g.Players[game.Player2].CommandZone.Contains(cardID) {
		t.Fatal("commander did not move to the command zone at the later SBA")
	}
}

func TestCompiledActualCardQuantityIndependentProducers(t *testing.T) {
	def := legacyCard(t, carddata.ScryfallCard{
		Name: "Independent", Layout: "normal", TypeLine: "Sorcery",
		OracleText: "Mill two cards. For each creature card milled this way, you gain 2 life. You gain 1 life. Mill two cards. For each creature card milled this way, you gain 3 life.",
	})
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	for range 2 {
		addCardToLibrary(g, game.Player1, vanillaCreature("First batch", 1, 1))
	}
	obj := &game.StackObject{Controller: game.Player1}
	before := g.Players[game.Player1].Life
	NewEngine(nil).resolveAbilityContentWithChoices(g, obj, def.SpellAbility.Val, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if got := g.Players[game.Player1].Life; got != before+5 {
		t.Fatalf("life=%d, want %d; second zero producer borrowed first batch", got, before+5)
	}
}

func TestCompiledActualCardQuantityPermanentCardDomain(t *testing.T) {
	for _, tc := range []struct {
		noun    string
		members [2]types.Card
		want    int
	}{
		{"card", [2]types.Card{types.Instant, types.Sorcery}, 2},
		{"permanent card", [2]types.Card{types.Instant, types.Sorcery}, 0},
		{"permanent card", [2]types.Card{types.Artifact, types.Land}, 2},
		{"nonland permanent card", [2]types.Card{types.Artifact, types.Land}, 1},
	} {
		t.Run(fmt.Sprintf("%s/%v", tc.noun, tc.members), func(t *testing.T) {
			def := legacyCard(t, carddata.ScryfallCard{
				Name: "Card domain", Layout: "normal", TypeLine: "Sorcery",
				OracleText: "Mill two cards. For each " + tc.noun + " milled this way, you gain 2 life.",
			})
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			for _, cardType := range tc.members {
				addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Types: []types.Card{cardType}}})
			}
			before := g.Players[game.Player1].Life
			NewEngine(nil).resolveAbilityContentWithChoices(g, &game.StackObject{Controller: game.Player1}, def.SpellAbility.Val, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
			if got := g.Players[game.Player1].Life; got != before+2*tc.want {
				t.Fatalf("life=%d, want %d; permanent-card universe collapsed into card wildcard", got, before+2*tc.want)
			}
		})
	}
}

func TestCompiledActualCardQuantityDestinationAndEventDomains(t *testing.T) {
	for _, tc := range []struct {
		name, text string
		discard    bool
		commander  bool
		want       int
	}{
		{"mill redirected", "Mill a card. For each creature card milled this way, you gain 2 life.", false, false, 2},
		{"mill commander", "Mill a card. For each creature card milled this way, you gain 2 life.", false, true, 2},
		{"discard redirected event", "Discard a card. For each creature card discarded this way, you gain 2 life.", true, false, 2},
		{"discard commander event", "Discard a card. For each creature card discarded this way, you gain 2 life.", true, true, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			def := legacyCard(t, carddata.ScryfallCard{Name: "Destination", Layout: "normal", TypeLine: "Sorcery", OracleText: tc.text})
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			owner := game.Player1
			if tc.commander {
				if tc.discard {
					addCommanderCardToHand(g, owner)
				} else {
					addCommanderCardToLibrary(g, owner)
				}
			} else {
				graveyardRedirectPermanent(g, game.Player1, game.TriggerControllerAny, false, types.Creature)
				if tc.discard {
					addCardToHand(g, owner, vanillaCreature("Redirected", 1, 1))
				} else {
					addCardToLibrary(g, owner, vanillaCreature("Redirected", 1, 1))
				}
			}
			before := g.Players[owner].Life
			NewEngine(nil).resolveAbilityContentWithChoices(g, &game.StackObject{Controller: owner}, def.SpellAbility.Val, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
			if got := g.Players[owner].Life; got != before+tc.want {
				t.Fatalf("life=%d, want %d; movement outcome and discard event conflated", got, before+tc.want)
			}
		})
	}
}

func actualCardQuantityReader(key game.ResultKey, addend int) game.Instruction {
	selection := game.Selection{RequiredTypes: []types.Card{types.Creature}}
	return game.Instruction{
		Primitive: game.GainLife{Player: game.ControllerReference(), Amount: game.Dynamic(game.DynamicAmount{
			Kind: game.DynamicAmountPreviousEffectResult, ResultKey: key, Selection: &selection, Addend: addend,
		})},
		ResultGate: opt.Val(game.InstructionResultGate{Key: key, Accepted: game.TriTrue}),
	}
}

func TestActualCardQuantityZeroAvailabilityAndLocalFrames(t *testing.T) {
	for _, tc := range []struct {
		name          string
		skip, decline bool
		want          int
	}{
		{"executed zero", false, false, 1},
		{"skipped unavailable", true, false, 0},
		{"declined unavailable", false, true, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			key := game.ResultKey("owned")
			obj := &game.StackObject{
				Controller:        game.Player1,
				ResolutionResults: map[string]game.InstructionResolutionResult{string(key): {Accepted: true, Succeeded: true}},
				ResolvedAmounts:   map[string]int{string(key): 99},
				ResolutionResultObjects: map[string][]game.ObjectSnapshot{string(key): {
					{CardID: 100, Types: []types.Card{types.Creature}},
					{CardID: 101, Types: []types.Card{types.Creature}},
				}},
			}
			publisher := game.Instruction{
				Primitive:     game.MoveTopOfLibrary{Player: game.ControllerReference(), Amount: game.Fixed(3), Destination: zone.Graveyard},
				PublishResult: key, LocalProducts: game.LocalProducts{Results: []game.ResultKey{key}},
				Optional: tc.decline,
			}
			if tc.skip {
				publisher.ConditionGate = "unavailable"
			}
			agents := [game.NumPlayers]PlayerAgent{}
			if tc.decline {
				agents[game.Player1] = &choiceOnlyAgent{choices: [][]int{{0}}}
			}
			before := g.Players[game.Player1].Life
			NewEngine(nil).resolveInstructionSequence(g, obj, []game.Instruction{publisher, actualCardQuantityReader(key, 1)}, agents, &TurnLog{})
			if got := g.Players[game.Player1].Life; got != before+tc.want {
				t.Fatalf("life=%d, want %d; zero/unavailable/stale batch confused", got, before+tc.want)
			}
			if obj.ResolvedAmounts[string(key)] != 99 || len(obj.ResolutionResultObjects[string(key)]) != 2 {
				t.Fatal("local result frame did not restore the outer scalar and snapshot cells")
			}
		})
	}
}

func TestActualCardQuantityRepeatedFrames(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCardToLibrary(g, game.Player1, vanillaCreature("First invocation", 1, 1))
	key := game.ResultKey("repeat")
	body := []game.Instruction{
		{Primitive: game.MoveTopOfLibrary{Player: game.ControllerReference(), Amount: game.Fixed(1), Destination: zone.Graveyard},
			PublishResult: key, LocalProducts: game.LocalProducts{Results: []game.ResultKey{key}}},
		actualCardQuantityReader(key, 1),
	}
	obj := &game.StackObject{Controller: game.Player1}
	before := g.Players[game.Player1].Life
	NewEngine(nil).resolveInstructionSequence(g, obj, []game.Instruction{{Primitive: game.RepeatProcess{
		Times: game.Fixed(2), Body: game.Mode{Sequence: body}.Ability(),
	}}}, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if got := g.Players[game.Player1].Life; got != before+3 {
		t.Fatalf("life=%d, want %d; repeated empty publisher borrowed prior invocation", got, before+3)
	}
	if len(obj.ResolutionResults) != 0 || len(obj.ResolutionResultObjects) != 0 {
		t.Fatal("repeated frame leaked producer products")
	}
}

func TestActualCardQuantityFrozenHistoryAndCardOnly(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	cardID := addCardToGraveyard(g, game.Player2, vanillaCreature("Frozen card", 1, 1))
	key := game.ResultKey("frozen")
	obj := &game.StackObject{Controller: game.Player1, Targets: []game.Target{currentCardTarget(t, g, cardID)}}
	engine := NewEngine(nil)
	publisher := game.Instruction{
		Primitive:     game.MoveCard{Card: game.CardReference{Kind: game.CardReferenceTarget}, FromZone: zone.Graveyard, Destination: zone.Exile},
		PublishResult: key,
	}
	engine.resolveInstructionWithChoices(g, obj, &publisher, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if !moveCardBetweenZones(g, game.Player2, cardID, zone.Exile, zone.Graveyard) {
		t.Fatal("failed to move historical member again")
	}
	g.CardInstances[cardID].Def.Types = []types.Card{types.Land}
	obj.ResolutionResultObjects[string(key)] = append(obj.ResolutionResultObjects[string(key)], game.ObjectSnapshot{
		ObjectID: g.IDGen.Next(), CardID: cardID, Types: []types.Card{types.Creature},
	})
	before := g.Players[game.Player1].Life
	reader := actualCardQuantityReader(key, 0)
	engine.resolveInstructionWithChoices(g, obj, &reader, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if got := g.Players[game.Player1].Life; got != before+1 {
		t.Fatalf("life=%d, want %d; live incarnation or departed permanent reinterpreted frozen cards", got, before+1)
	}
	if scalar := dynamicAmountValue(g, obj, game.Player1, game.DynamicAmount{Kind: game.DynamicAmountPreviousEffectResult, ResultKey: key}); scalar != obj.ResolvedAmounts[string(key)] {
		t.Fatal("nil-Selection scalar reader changed")
	}
}
