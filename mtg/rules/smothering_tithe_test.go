package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func TestDrawTaxTreasurePaymentOutcomes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		canPay         bool
		acceptsPayment bool
		wantTreasures  int
		wantChoices    int
	}{
		{name: "pays", canPay: true, acceptsPayment: true, wantChoices: 1},
		{name: "declines", canPay: true, wantTreasures: 1, wantChoices: 1},
		{name: "cannot pay", wantTreasures: 1},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			token := testTreasureToken()
			addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
				Event:  game.EventCardDrawn,
				Player: game.TriggerPlayerOpponent,
			}, eventPlayerTaxedTreasureInstructions(cost.Mana{cost.O(2)}, token), nil)
			addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
			if tc.canPay {
				addBasicLandPermanent(g, game.Player2, types.Forest)
				addBasicLandPermanent(g, game.Player2, types.Forest)
			}

			if _, ok := engine.drawCard(g, game.Player2, false); !ok {
				t.Fatal("drawCard() = false")
			}
			if !engine.putTriggeredAbilitiesOnStack(g) {
				t.Fatal("draw trigger was not put on the stack")
			}
			choice := []int{0}
			if tc.acceptsPayment {
				choice = []int{1}
			}
			agents := [game.NumPlayers]PlayerAgent{
				game.Player2: &choiceOnlyAgent{choices: [][]int{choice}},
			}
			log := TurnLog{}
			engine.resolveTopOfStackWithChoices(g, agents, &log)

			if got := countTokenDef(g, token); got != tc.wantTreasures {
				t.Fatalf("Treasures = %d, want %d", got, tc.wantTreasures)
			}
			if len(log.Choices) != tc.wantChoices {
				t.Fatalf("choices = %+v, want %d", log.Choices, tc.wantChoices)
			}
			if tc.wantChoices == 1 && log.Choices[0].Request.Player != game.Player2 {
				t.Fatalf("payment choice player = %v, want drawing player", log.Choices[0].Request.Player)
			}
			if tc.acceptsPayment && g.Players[game.Player2].ManaPool.Total() != 0 {
				t.Fatalf("payer mana pool = %v, want payment consumed", g.Players[game.Player2].ManaPool)
			}
		})
	}
}

func TestDrawTaxCanBePaidWithGeneratedTreasures(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	token := testTreasureToken()
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:  game.EventCardDrawn,
		Player: game.TriggerPlayerOpponent,
	}, eventPlayerTaxedTreasureInstructions(cost.Mana{cost.O(2)}, token), nil)
	treasures, ok := createTokenPermanentsCollectingWithChoices(
		engine,
		g,
		game.Player2,
		token,
		2,
		false,
		[game.NumPlayers]PlayerAgent{},
		nil,
	)
	if !ok {
		t.Fatal("creating payment Treasures failed")
	}
	addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})

	if _, ok = engine.drawCard(g, game.Player2, false); !ok {
		t.Fatal("drawCard() = false")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("draw trigger was not put on the stack")
	}
	agents := [game.NumPlayers]PlayerAgent{
		game.Player2: &choiceOnlyAgent{choices: [][]int{{1}}},
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if got := countTokenDef(g, token); got != 0 {
		t.Fatalf("Treasures = %d, want both payment Treasures sacrificed and no new Treasure", got)
	}
	if got := g.Players[game.Player2].ManaPool.Total(); got != 0 {
		t.Fatalf("payer mana pool = %d, want Treasure mana consumed", got)
	}
	for _, treasure := range treasures {
		tapped := false
		sacrificed := false
		for _, event := range g.Events {
			tapped = tapped || event.Kind == game.EventPermanentTapped && event.PermanentID == treasure.ObjectID
			sacrificed = sacrificed || event.Kind == game.EventPermanentSacrificed && event.PermanentID == treasure.ObjectID
		}
		if !tapped || !sacrificed {
			t.Fatalf("Treasure %v events: tapped=%v sacrificed=%v, want both", treasure.ObjectID, tapped, sacrificed)
		}
	}
}

func TestDrawTaxTreasureEachOpponentAndEachCard(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	token := testTreasureToken()
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:  game.EventCardDrawn,
		Player: game.TriggerPlayerOpponent,
	}, eventPlayerTaxedTreasureInstructions(cost.Mana{cost.O(2)}, token), nil)
	for _, card := range []string{"P2 first", "P2 second"} {
		addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: card}})
	}
	addCardToLibrary(g, game.Player3, &game.CardDef{CardFace: game.CardFace{Name: "P3"}})
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Controller"}})

	for _, player := range []game.PlayerID{game.Player2, game.Player2, game.Player3, game.Player1} {
		if _, ok := engine.drawCard(g, player, false); !ok {
			t.Fatalf("drawCard(%v) = false", player)
		}
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("opponent draw triggers were not put on the stack")
	}
	if got := g.Stack.Size(); got != 3 {
		t.Fatalf("stack size = %d, want one trigger for each of three opponent card draws", got)
	}
	for g.Stack.Size() > 0 {
		engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	}
	if got := countTokenDef(g, token); got != 3 {
		t.Fatalf("Treasures = %d, want 3", got)
	}
}

func TestDrawTaxTreasureTriggersWhenSourceLeavesBeforeTriggerProcessing(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	token := testTreasureToken()
	source := addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:  game.EventCardDrawn,
		Player: game.TriggerPlayerOpponent,
	}, eventPlayerTaxedTreasureInstructions(cost.Mana{cost.O(2)}, token), nil)
	addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})

	if _, ok := engine.drawCard(g, game.Player2, false); !ok {
		t.Fatal("drawCard() = false")
	}
	movePermanentToZone(g, source, zone.Graveyard)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("draw trigger was lost when its source left before trigger processing")
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.SourceID != source.ObjectID || obj.Controller != game.Player1 {
		t.Fatalf("trigger = %+v, want source %v controlled by Player1 at trigger time", obj, source.ObjectID)
	}
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if got := countTokenDef(g, token); got != 1 {
		t.Fatalf("Treasures = %d, want trigger controller to create one after source left", got)
	}
}

func TestDrawTaxTreasureKeepsTriggerTimeControllerAndAPNAPOrder(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	token := testTreasureToken()
	first := addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:  game.EventCardDrawn,
		Player: game.TriggerPlayerOpponent,
	}, eventPlayerTaxedTreasureInstructions(cost.Mana{cost.O(2)}, token), nil)
	second := addTriggeredPermanent(g, game.Player2, &game.TriggerPattern{
		Event:  game.EventCardDrawn,
		Player: game.TriggerPlayerOpponent,
	}, eventPlayerTaxedTreasureInstructions(cost.Mana{cost.O(2)}, token), nil)
	addCardToLibrary(g, game.Player4, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})

	if _, ok := engine.drawCard(g, game.Player4, false); !ok {
		t.Fatal("drawCard() = false")
	}
	first.Controller = game.Player3
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("draw triggers were not put on the stack")
	}
	objects := g.Stack.Objects()
	if len(objects) != 2 {
		t.Fatalf("stack size = %d, want 2", len(objects))
	}
	if objects[0].SourceID != first.ObjectID || objects[0].Controller != game.Player1 {
		t.Fatalf("bottom trigger = %+v, want first source with trigger-time controller Player1", objects[0])
	}
	if objects[1].SourceID != second.ObjectID || objects[1].Controller != game.Player2 {
		t.Fatalf("top trigger = %+v, want Player2 source above Player1 in APNAP order", objects[1])
	}
}

func eventPlayerTaxedTreasureInstructions(manaCost cost.Mana, token *game.CardDef) []game.Instruction {
	return []game.Instruction{
		{
			Primitive: game.Pay{Payment: game.ResolutionPayment{
				Prompt:   "Pay " + manaCost.String() + "?",
				Payer:    opt.Val(game.EventPlayerReference()),
				ManaCost: opt.Val(slices.Clone(manaCost)),
			}},
			PublishResult: "unless-paid",
		},
		{
			Primitive: game.CreateToken{Amount: game.Fixed(1), Source: game.TokenDef(token)},
			ResultGate: opt.Val(game.InstructionResultGate{
				Key:       "unless-paid",
				Succeeded: game.TriFalse,
			}),
		},
	}
}

func testTreasureToken() *game.CardDef {
	ability := game.TapManaChoiceAbility(mana.W, mana.U, mana.B, mana.R, mana.G)
	ability.Text = "{T}, Sacrifice this artifact: Add one mana of any color."
	ability.AdditionalCosts = append(ability.AdditionalCosts, cost.Additional{
		Kind:               cost.AdditionalSacrificeSource,
		Text:               "Sacrifice this artifact",
		Amount:             1,
		MatchPermanentType: true,
		PermanentType:      types.Artifact,
	})
	return &game.CardDef{CardFace: game.CardFace{
		Name:          string(types.Treasure),
		Types:         []types.Card{types.Artifact},
		Subtypes:      []types.Sub{types.Treasure},
		ManaAbilities: []game.ManaAbility{ability},
	}}
}

func countTokenDef(g *game.Game, def *game.CardDef) int {
	count := 0
	for _, permanent := range g.Battlefield {
		if permanent.Token && permanent.TokenDef == def && permanent.Controller == game.Player1 {
			count++
		}
	}
	return count
}
