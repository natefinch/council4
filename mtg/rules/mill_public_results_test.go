package rules

import (
	"fmt"
	"slices"
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestMillPublicResultDoesNotExposeHiddenCharacteristics(t *testing.T) {
	for _, destination := range []zone.Type{zone.Hand, zone.Library} {
		t.Run(destination.String(), func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			g.ReplacementEffects = append(g.ReplacementEffects, game.ReplacementEffect{
				ID: g.IDGen.Next(), Controller: game.Player1,
				MatchEvent:    game.EventZoneChanged,
				MatchFromZone: true, FromZone: zone.Library,
				MatchToZone: true, ToZone: zone.Graveyard,
				ReplaceToZone: destination,
			})
			cardID := addCardToLibrary(g, game.Player2, vanillaCreature("Hidden", 1, 1))
			obj := &game.StackObject{Controller: game.Player1, Targets: []game.Target{game.PlayerTarget(game.Player2)}}
			NewEngine(nil).resolveInstructionWithChoices(g, obj, &game.Instruction{
				Primitive: game.MoveTopOfLibrary{
					Player: game.TargetPlayerReference(0), Amount: game.Fixed(1), Destination: zone.Graveyard,
					PublishLinked: "graveyard-only",
				},
				PublishResult: "mill",
			}, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
			actual, exists := cardZone(g, cardID)
			if !exists || actual != destination || len(obj.ResolutionResultObjects["mill"]) != 0 {
				t.Fatal("hidden destination published printed characteristics or failed to move the card")
			}
			if len(linkedObjects(g, linkedObjectSourceKey(g, obj, "graveyard-only"))) != 0 {
				t.Fatal("hidden output widened the graveyard link")
			}
			departures := 1
			if destination == zone.Library {
				departures = 0
			}
			receipt := obj.ResolutionResults["mill"]
			if receipt.Amount != departures || receipt.Succeeded != (departures > 0) {
				t.Fatalf("receipt=%#v, want actual library departures %d", receipt, departures)
			}
		})
	}
}

func TestMillPublicResultSnapshotsAndLegacyLinks(t *testing.T) {
	for _, tc := range []struct {
		name      string
		members   []types.Card
		redirect  []types.Card
		commander bool
	}{
		{"zero", nil, nil, false},
		{"normal two", []types.Card{types.Creature, types.Land}, nil, false},
		{"all public exile", []types.Card{types.Creature, types.Land}, []types.Card{types.Creature, types.Land}, false},
		{"mixed destinations", []types.Card{types.Creature, types.Land}, []types.Card{types.Creature}, false},
		{"command", []types.Card{types.Creature}, nil, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			source := addCombatPermanent(g, game.Player3, &game.CardDef{CardFace: game.CardFace{Name: "Foreign owned source"}})
			if len(tc.redirect) > 0 {
				graveyardRedirectPermanent(g, game.Player1, game.TriggerControllerAny, false, tc.redirect...)
			}
			var members []id.ID
			var graveyardMembers []id.ID
			for _, cardType := range tc.members {
				cardID := addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
					Name: fmt.Sprintf("Member %d", len(members)), Types: []types.Card{cardType},
				}})
				members = append(members, cardID)
				if !slices.Contains(tc.redirect, cardType) {
					graveyardMembers = append(graveyardMembers, cardID)
				}
			}
			if tc.commander {
				g.CommanderIDs[members[0]] = true
				g.Players[game.Player2].CommanderInstanceID = members[0]
			}
			obj := &game.StackObject{
				Controller: game.Player1, SourceID: source.ObjectID, SourceCardID: source.CardInstanceID,
				Targets: []game.Target{game.PlayerTarget(game.Player2)},
			}
			publisher := game.Instruction{
				Primitive: game.MoveTopOfLibrary{
					Player: game.TargetPlayerReference(0), Amount: game.Fixed(3),
					Destination: zone.Graveyard, PublishLinked: "legacy-graveyard",
				},
				PublishResult: "actual-mill",
			}
			NewEngine(nil).resolveInstructionWithChoices(g, obj, &publisher, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
			snapshots := obj.ResolutionResultObjects["actual-mill"]
			if len(snapshots) != len(members) {
				t.Fatalf("snapshots=%d, want %d actual public departures", len(snapshots), len(members))
			}
			seen := make(map[id.ID]bool)
			for _, snapshot := range snapshots {
				if seen[snapshot.CardID] || !slices.Contains(members, snapshot.CardID) ||
					snapshot.ObjectID != 0 || snapshot.Owner != game.Player2 ||
					snapshot.Controller != game.Player2 {
					t.Fatalf("duplicated, foreign, or permanent-domain snapshot: %#v", snapshot)
				}
				seen[snapshot.CardID] = true
			}
			receipt := obj.ResolutionResults["actual-mill"]
			if !receipt.Accepted || receipt.Succeeded != (len(members) > 0) ||
				receipt.Amount != len(members) || obj.ResolvedAmounts["actual-mill"] != len(members) {
				t.Fatalf("receipt=%#v scalar=%d, want actual performed %d", receipt, obj.ResolvedAmounts["actual-mill"], len(members))
			}
			if _, available := obj.ResolvedAmounts["actual-mill"]; !available {
				t.Fatal("executed zero mill lost its known scalar availability")
			}
			if got := dynamicAmountValue(g, obj, game.Player1, game.DynamicAmount{
				Kind: game.DynamicAmountPreviousEffectResult, ResultKey: "actual-mill",
			}); got != len(members) {
				t.Fatalf("nil-Selection scalar=%d, want actual departure receipt %d", got, len(members))
			}
			linked := linkedObjects(g, linkedObjectSourceKey(g, obj, "legacy-graveyard"))
			if len(linked) != len(graveyardMembers) {
				t.Fatalf("legacy links=%d, want strict graveyard set %d", len(linked), len(graveyardMembers))
			}
			for _, member := range linked {
				if !slices.Contains(graveyardMembers, member.CardID) {
					t.Fatal("public redirect widened a legacy graveyard-only link")
				}
			}
		})
	}
}

func TestMillLegacyReturnStaysGraveyardOnly(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	graveyardRedirectPermanent(g, game.Player1, game.TriggerControllerAny, false, types.Creature)
	exiled := addCardToLibrary(g, game.Player1, vanillaCreature("Redirected", 1, 1))
	graveyard := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Types: []types.Card{types.Land}}})
	got := millCards(g, game.Player1, 2)
	if !slices.Equal(got, []id.ID{graveyard}) || !g.Players[game.Player1].Exile.Contains(exiled) {
		t.Fatalf("legacy millCards=%v, want only actual graveyard card %v", got, graveyard)
	}
}

func TestMillActualSuccessCompatibilityWrapper(t *testing.T) {
	for _, tc := range []struct {
		name                string
		count               int
		redirect, unchanged bool
	}{
		{"zero", 0, false, false},
		{"one", 1, false, false},
		{"two", 2, false, false},
		{"public redirect", 2, true, false},
		{"no library departure", 1, false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			if tc.redirect {
				graveyardRedirectPermanent(g, game.Player1, game.TriggerControllerAny, false)
			}
			for range tc.count {
				def := vanillaCreature("Compatibility member", 1, 1)
				if tc.unchanged {
					def = selfLibraryReplacementCardDef()
				}
				addCardToLibrary(g, game.Player1, def)
			}
			cards, succeeded := millCardsWithResult(g, game.Player1, 3)
			wantMembers := tc.count
			if tc.redirect || tc.unchanged {
				wantMembers = 0
			}
			if len(cards) != wantMembers || succeeded != (tc.count > 0 && !tc.unchanged) {
				t.Fatalf("legacy members=%d success=%t, want members=%d actual departure=%t",
					len(cards), succeeded, wantMembers, tc.count > 0 && !tc.unchanged)
			}
		})
	}
}

func TestMillPublicGroupRequestIsStablePerPlayer(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	for _, playerID := range []game.PlayerID{game.Player2, game.Player3} {
		for range 3 {
			addCardToLibrary(g, playerID, vanillaCreature("Per-player member", 1, 1))
		}
	}
	obj := &game.StackObject{Controller: game.Player1}
	NewEngine(nil).resolveInstructionWithChoices(g, obj, &game.Instruction{
		Primitive:     game.MoveTopOfLibrary{PlayerGroup: game.OpponentsReference(), Amount: game.Fixed(2), Destination: zone.Graveyard},
		PublishResult: "stable",
	}, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if g.Players[game.Player2].Library.Size() != 1 || g.Players[game.Player3].Library.Size() != 1 ||
		obj.ResolvedAmounts["stable"] != 4 || len(obj.ResolutionResultObjects["stable"]) != 4 {
		t.Fatal("accumulated actual result replaced the request used by a later player")
	}
}

func TestCompiledMilledCardFiltersIncludePublicRedirects(t *testing.T) {
	for _, noun := range []string{"card", "creature card", "land card"} {
		for _, qualifying := range []int{0, 2} {
			t.Run(fmt.Sprintf("%s/%d", noun, qualifying), func(t *testing.T) {
				def := compileUnlessCard(t, cardgen.ScryfallCard{
					Name: "Public mill quantity", Layout: "normal", TypeLine: "Sorcery",
					OracleText: "Mill three cards. For each " + noun + " milled this way, create a 1/1 green Insect creature token.",
				})
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				graveyardRedirectPermanent(g, game.Player1, game.TriggerControllerAny, false)
				for i := range 2 {
					cardType := types.Land
					if i < qualifying {
						cardType = types.Creature
					}
					addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Types: []types.Card{cardType}}})
				}
				obj := &game.StackObject{Controller: game.Player1}
				NewEngine(nil).resolveAbilityContentWithChoices(g, obj, def.SpellAbility.Val, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
				var want int
				switch noun {
				case "creature card":
					want = qualifying
				case "land card":
					want = 2 - qualifying
				default:
					want = 2
				}
				if g.Players[game.Player1].Exile.Size() != 2 || len(g.Battlefield) != want+1 {
					t.Fatalf("exiled=%d permanents=%d, want actual redirected batch 2 and tokens %d plus redirect source",
						g.Players[game.Player1].Exile.Size(), len(g.Battlefield), want)
				}
			})
		}
	}
}

func TestCompiledMilledResultConditionIncludesPublicRedirect(t *testing.T) {
	def := compileUnlessCard(t, cardgen.ScryfallCard{
		Name: "Public mill condition", Layout: "normal", TypeLine: "Sorcery",
		OracleText: "Mill a card. If a creature card was milled this way, you gain 2 life.",
	})
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	graveyardRedirectPermanent(g, game.Player1, game.TriggerControllerAny, false, types.Creature)
	addCardToLibrary(g, game.Player1, vanillaCreature("Redirected", 1, 1))
	NewEngine(nil).resolveAbilityContentWithChoices(g, &game.StackObject{Controller: game.Player1},
		def.SpellAbility.Val, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if g.Players[game.Player1].Life != 42 {
		t.Fatal("typed milled-this-way condition lost a public redirected card")
	}
}

func TestMillPublicResultGroupCountsActualPlayers(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	graveyardRedirectPermanent(g, game.Player1, game.TriggerControllerAny, false, types.Creature)
	for _, playerID := range []game.PlayerID{game.Player1, game.Player2, game.Player3} {
		addCardToLibrary(g, playerID, vanillaCreature("Group member", 1, 1))
	}

	obj := &game.StackObject{Controller: game.Player1}
	NewEngine(nil).resolveInstructionWithChoices(g, obj, &game.Instruction{
		Primitive: game.MoveTopOfLibrary{
			PlayerGroup: game.OpponentsReference(), Amount: game.Fixed(3), Destination: zone.Graveyard,
		},
		PublishResult: "group",
	}, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if len(obj.ResolutionResultObjects["group"]) != 2 || obj.ResolvedAmounts["group"] != 2 ||
		g.Players[game.Player1].Library.Size() != 1 {
		t.Fatal("group result counted requested cards, omitted public redirects, or milled the source controller")
	}
}

func TestMillPublicResultRepeatRestoresOuterCells(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	graveyardRedirectPermanent(g, game.Player1, game.TriggerControllerAny, false, types.Creature)
	addCardToLibrary(g, game.Player1, vanillaCreature("First public invocation", 1, 1))
	key := game.ResultKey("public-repeat")
	obj := &game.StackObject{
		Controller:              game.Player1,
		ResolutionResults:       map[string]game.InstructionResolutionResult{string(key): {Accepted: true, Succeeded: true}},
		ResolvedAmounts:         map[string]int{string(key): 99},
		ResolutionResultObjects: map[string][]game.ObjectSnapshot{string(key): {{CardID: 100, Types: []types.Card{types.Creature}}}},
	}
	body := []game.Instruction{
		{Primitive: game.MoveTopOfLibrary{Player: game.ControllerReference(), Amount: game.Fixed(3), Destination: zone.Graveyard},
			PublishResult: key, LocalProducts: game.LocalProducts{Results: []game.ResultKey{key}}},
		actualCardQuantityReader(key, 1),
	}
	before := g.Players[game.Player1].Life
	NewEngine(nil).resolveInstructionSequence(g, obj, []game.Instruction{{Primitive: game.RepeatProcess{
		Times: game.Fixed(2), Body: game.Mode{Sequence: body}.Ability(),
	}}}, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if g.Players[game.Player1].Life != before+3 || obj.ResolvedAmounts[string(key)] != 99 ||
		len(obj.ResolutionResultObjects[string(key)]) != 1 || obj.ResolutionResultObjects[string(key)][0].CardID != 100 {
		t.Fatal("public mill result borrowed an older invocation or lost the outer result frame")
	}
}
