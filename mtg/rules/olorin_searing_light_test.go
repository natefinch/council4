package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

const testCorrelatedOpponentExileKey = game.LinkedKey("test-correlated-opponent-exile")

func olorinTestContent() game.AbilityContent {
	member := game.GroupOfferMemberReference()
	return game.Mode{Sequence: []game.Instruction{
		{
			Primitive: game.ExileForEachOpponent{
				Chooser:      member,
				Selection:    game.Selection{RequiredTypes: []types.Card{types.Creature}},
				LinkedKey:    testCorrelatedOpponentExileKey,
				Required:     true,
				Extremum:     game.PermanentChoiceGreatestPower,
				Simultaneous: true,
			},
		},
		{
			Primitive: game.Damage{
				Amount: game.Dynamic(game.DynamicAmount{
					Kind:   game.DynamicAmountObjectPower,
					Object: game.LinkedObjectReference(string(testCorrelatedOpponentExileKey)),
					Player: &member,
				}),
				Recipient: game.PlayerDamageRecipient(member),
			},
			Condition: opt.Val(game.EffectCondition{Condition: opt.Val(game.Condition{
				ControllerGraveyardInstantOrSorceryCountAtLeast: 2,
			})}),
			ForEachPlayerGroup: opt.Val(game.OpponentsReference()),
		},
	}}.Ability()
}

type olorinTestResult struct {
	game   *game.Game
	tieA   *game.Permanent
	tieB   *game.Permanent
	token  *game.Permanent
	agents [game.NumPlayers]*choiceOnlyAgent
	obj    *game.StackObject
}

func resolveOlorinTest(t *testing.T, mastery bool) olorinTestResult {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer = game.Player2
	engine := NewEngine(nil)

	addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	tieA := addCombatCreaturePermanentWithPower(g, game.Player2, 5)
	tieB := addCombatCreaturePermanentWithPower(g, game.Player2, 5)
	token := addCombatTokenCreaturePermanent(g, game.Player4, 4)
	addCombatCreaturePermanentWithPower(g, game.Player4, 3)

	if mastery {
		addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Types: []types.Card{types.Instant}}})
		addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Types: []types.Card{types.Sorcery}}})
	}
	obj := &game.StackObject{
		ID:           g.IDGen.Next(),
		Kind:         game.StackSpell,
		Controller:   game.Player1,
		SourceCardID: g.IDGen.Next(),
	}
	agents := [game.NumPlayers]*choiceOnlyAgent{
		game.Player2: {choices: [][]int{{1}}},
		game.Player4: {choices: [][]int{{0}}},
	}
	var playerAgents [game.NumPlayers]PlayerAgent
	for i := range agents {
		playerAgents[i] = agents[i]
	}
	engine.resolveAbilityContentWithChoices(g, obj, olorinTestContent(), playerAgents, &TurnLog{})
	return olorinTestResult{game: g, tieA: tieA, tieB: tieB, token: token, agents: agents, obj: obj}
}

func TestOlorinsSearingLightMultiplayerChoicesAreCorrelatedAndSimultaneous(t *testing.T) {
	result := resolveOlorinTest(t, true)
	g, tieA, tieB, token, agents, obj := result.game, result.tieA, result.tieB, result.token, result.agents, result.obj
	if agents[game.Player1] != nil || agents[game.Player3] != nil {
		t.Fatal("players without a required choice unexpectedly received an agent")
	}
	if agents[game.Player2].next != 1 || agents[game.Player4].next != 1 {
		t.Fatalf("choice counts = Player2 %d, Player4 %d", agents[game.Player2].next, agents[game.Player4].next)
	}
	if permanentByCardID(g, tieA.CardInstanceID) == nil {
		t.Fatal("unchosen tied creature left the battlefield")
	}
	if permanentByCardID(g, tieB.CardInstanceID) != nil {
		t.Fatal("chosen tied creature remained on the battlefield")
	}
	if permanentByObjectIDMustNil(g, token.ObjectID) != nil {
		t.Fatal("chosen token remained on the battlefield")
	}
	refs := linkedObjects(g, linkedObjectSourceKey(g, obj, string(testCorrelatedOpponentExileKey)))
	if len(refs) != 2 ||
		!refs[0].CorrelatedPlayer.Exists || refs[0].CorrelatedPlayer.Val != game.Player2 ||
		!refs[1].CorrelatedPlayer.Exists || refs[1].CorrelatedPlayer.Val != game.Player4 {
		t.Fatalf("correlated refs = %#v", refs)
	}
	var simultaneous game.ObjectID
	moves := 0
	for _, event := range g.Events {
		if event.Kind != game.EventZoneChanged || event.FromZone != zone.Battlefield ||
			(event.PermanentID != tieB.ObjectID && event.PermanentID != token.ObjectID) {
			continue
		}
		moves++
		if simultaneous == 0 {
			simultaneous = event.SimultaneousID
		} else if event.SimultaneousID != simultaneous {
			t.Fatalf("move batch IDs differ: %v and %v", simultaneous, event.SimultaneousID)
		}
	}
	if moves != 2 || simultaneous == 0 {
		t.Fatalf("simultaneous exile events = %d, batch %v", moves, simultaneous)
	}
}

func permanentByObjectIDMustNil(g *game.Game, objectID game.ObjectID) *game.Permanent {
	permanent, _ := permanentByObjectID(g, objectID)
	return permanent
}

func TestOlorinsSearingLightSpellMasteryAndTokenLKI(t *testing.T) {
	for _, test := range []struct {
		name        string
		mastery     bool
		wantPlayer2 int
		wantPlayer3 int
		wantPlayer4 int
	}{
		{name: "off", mastery: false, wantPlayer2: 40, wantPlayer3: 40, wantPlayer4: 40},
		{name: "on", mastery: true, wantPlayer2: 35, wantPlayer3: 40, wantPlayer4: 36},
	} {
		t.Run(test.name, func(t *testing.T) {
			result := resolveOlorinTest(t, test.mastery)
			g, obj := result.game, result.obj
			if g.Players[game.Player2].Life != test.wantPlayer2 ||
				g.Players[game.Player3].Life != test.wantPlayer3 ||
				g.Players[game.Player4].Life != test.wantPlayer4 {
				t.Fatalf("life totals = %d/%d/%d", g.Players[game.Player2].Life, g.Players[game.Player3].Life, g.Players[game.Player4].Life)
			}
			if !test.mastery {
				return
			}
			damageEvents := 0
			for _, event := range g.Events {
				if event.Kind != game.EventDamageDealt {
					continue
				}
				damageEvents++
				if event.Controller != game.Player1 || event.SourceObjectID != obj.ID {
					t.Fatalf("damage source/controller = %#v", event)
				}
			}

			if damageEvents != 2 {
				t.Fatalf("damage events = %d, want 2", damageEvents)
			}
		})
	}
}

func TestDynamicAmountsBindForEachPlayerGroupMember(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Players[game.Player2].Life = 17
	addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	addCombatCreaturePermanentWithPower(g, game.Player2, 3)
	addCombatCreaturePermanentWithPower(g, game.Player3, 4)
	addCardToGraveyard(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Types: []types.Card{types.Instant}}})

	obj := &game.StackObject{Controller: game.Player1}
	member := game.GroupOfferMemberReference()
	groupMember := opt.Val(game.Player2)
	tests := []struct {
		name    string
		dynamic game.DynamicAmount
		want    int
	}{
		{
			name:    "player life",
			dynamic: game.DynamicAmount{Kind: game.DynamicAmountPlayerLife, Player: &member},
			want:    17,
		},
		{
			name: "controlled group count",
			dynamic: game.DynamicAmount{
				Kind:  game.DynamicAmountCountSelector,
				Group: game.PlayerControlledGroup(member, game.Selection{RequiredTypes: []types.Card{types.Creature}}),
			},
			want: 2,
		},
		{
			name: "cards in zone",
			dynamic: game.DynamicAmount{
				Kind:      game.DynamicAmountCountCardsInZone,
				Player:    &member,
				CardZone:  zone.Graveyard,
				Selection: &game.Selection{RequiredTypes: []types.Card{types.Instant}},
			},
			want: 1,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := dynamicAmountValueBeforeLayerWithGroupMember(
				g,
				opt.Val(obj),
				game.Player1,
				test.dynamic,
				0,
				groupMember,
			)
			if got != test.want {
				t.Fatalf("amount = %d, want %d", got, test.want)
			}
		})
	}
}
