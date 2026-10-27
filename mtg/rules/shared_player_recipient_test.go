package rules

import (
	"testing"

	"github.com/natefinch/council4/internal/carddata"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestCompiledSharedPlayerRecipientSecondSlot(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, action string
		life, hand   int
		library      int
	}{
		{"life", "Target opponent loses 2 life.", 38, 0, 3},
		{"draw", "Target opponent draws two cards.", 40, 2, 1},
		{"mill", "Target opponent mills two cards.", 40, 0, 1},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			def := legacyCard(t, carddata.ScryfallCard{
				Name: "Second Recipient", Layout: "normal", TypeLine: "Instant",
				OracleText: "Tap target creature. " + tt.action,
			})
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			decoy := addCombatCreaturePermanent(g, game.Player2)
			for player := range g.Players {
				for range 3 {
					addCardToLibrary(g, game.PlayerID(player), evidenceCard("Library card", 1))
				}
			}
			obj := &game.StackObject{Kind: game.StackSpell, Controller: game.Player1,
				Targets:      []game.Target{game.PermanentTarget(decoy.ObjectID), game.PlayerTarget(game.Player3)},
				TargetCounts: []int{1, 1}}
			NewEngine(nil).resolveAbilityContentWithChoices(g, obj, def.SpellAbility.Val, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
			if !decoy.Tapped {
				t.Fatal("independent earlier target was not tapped")
			}
			for player, state := range g.Players {
				life, hand, library := 40, 0, 3
				if game.PlayerID(player) == game.Player3 {
					life, hand, library = tt.life, tt.hand, tt.library
				}
				if state.Life != life || state.Hand.Size() != hand || state.Library.Size() != library {
					t.Fatalf("player %d: life/hand/library=%d/%d/%d, want %d/%d/%d",
						player, state.Life, state.Hand.Size(), state.Library.Size(), life, hand, library)
				}
			}
		})
	}
}

func TestCompiledSharedPlayerRecipientDepartedControllerAndOwner(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, action string
		player       game.PlayerID
		life, hand   int
		library      int
	}{
		{"controller life", "Its controller gains 2 life.", game.Player3, 42, 0, 3},
		{"owner life", "Its owner gains 2 life.", game.Player2, 42, 0, 3},
		{"controller draw", "Its controller draws a card.", game.Player3, 40, 1, 2},
		{"controller mill", "Its controller mills two cards.", game.Player3, 40, 0, 1},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			def := legacyCard(t, carddata.ScryfallCard{
				Name: "Departed Recipient", Layout: "normal", TypeLine: "Instant",
				OracleText: "Tap target artifact. Destroy target creature. Scry 1. " + tt.action,
			})
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			decoy := addCombatPermanent(g, game.Player4, &game.CardDef{CardFace: game.CardFace{
				Name: "Decoy", Types: []types.Card{types.Artifact},
			}})
			departing := addCombatCreaturePermanent(g, game.Player2)
			departing.Controller = game.Player3
			for player := range g.Players {
				for range 3 {
					addCardToLibrary(g, game.PlayerID(player), evidenceCard("Library card", 1))
				}
			}
			obj := &game.StackObject{Kind: game.StackSpell, Controller: game.Player1,
				Targets:      []game.Target{game.PermanentTarget(decoy.ObjectID), game.PermanentTarget(departing.ObjectID)},
				TargetCounts: []int{1, 1}}
			NewEngine(nil).resolveAbilityContentWithChoices(g, obj, def.SpellAbility.Val, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
			if !decoy.Tapped {
				t.Fatal("earlier artifact target was not tapped")
			}
			if _, present := permanentByObjectID(g, departing.ObjectID); present {
				t.Fatal("creature target did not depart")
			}
			if !g.Players[game.Player2].Graveyard.Contains(departing.CardInstanceID) {
				t.Fatal("destroyed card did not go to its physical owner's graveyard")
			}
			for player, state := range g.Players {
				life, hand, library := 40, 0, 3
				if game.PlayerID(player) == tt.player {
					life, hand, library = tt.life, tt.hand, tt.library
				}
				if state.Life != life || state.Hand.Size() != hand || state.Library.Size() != library {
					t.Fatalf("player %d: life/hand/library=%d/%d/%d, want %d/%d/%d",
						player, state.Life, state.Hand.Size(), state.Library.Size(), life, hand, library)
				}
			}
		})
	}
}

func TestCompiledSharedPlayerRecipientEdictOutcome(t *testing.T) {
	t.Parallel()
	def := legacyCard(t, carddata.ScryfallCard{
		Name: "Self-Inflicted Wound", Layout: "normal", TypeLine: "Sorcery",
		OracleText: "Target opponent sacrifices a green or white creature of their choice. If that player does, they lose 2 life.",
	})
	for _, succeeds := range []bool{false, true} {
		t.Run(map[bool]string{false: "no matching creature", true: "actual sacrifice"}[succeeds], func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			if succeeds {
				creatureDef := legacyCard(t, carddata.ScryfallCard{
					Name: "Green Creature", Layout: "normal", TypeLine: "Creature",
					ManaCost: "{G}", Colors: []string{"G"}, Power: new("2"), Toughness: new("2"),
				})
				creature := addCombatPermanent(g, game.Player2, creatureDef)
				creature.Controller = game.Player3
			}
			obj := &game.StackObject{Kind: game.StackSpell, Controller: game.Player1,
				Targets: []game.Target{game.PlayerTarget(game.Player3)}, TargetCounts: []int{1}}
			result, available := NewEngine(nil).resolveAbilityContentReceipt(g, obj, def.SpellAbility.Val,
				[game.NumPlayers]PlayerAgent{}, &TurnLog{}, "if-you-do")
			if !available || result.Succeeded != succeeds {
				t.Fatalf("actual sacrifice result=%+v available=%v, want succeeded=%v", result, available, succeeds)
			}
			if succeeds && g.Players[game.Player2].Graveyard.Size() != 1 {
				t.Fatal("actual sacrificed card did not reach its physical owner's graveyard")
			}
			for player, state := range g.Players {
				want := 40
				if succeeds && game.PlayerID(player) == game.Player3 {
					want = 38
				}

				if state.Life != want {
					t.Fatalf("player %d life=%d, want %d", player, state.Life, want)
				}
			}
		})
	}
}

func TestCompiledSharedPlayerRecipientZoneEventPlayer(t *testing.T) {
	t.Parallel()
	def := legacyCard(t, carddata.ScryfallCard{
		Name: "Zone Event Recipient", Layout: "normal", TypeLine: "Enchantment",
		OracleText: "Whenever a card is put into an opponent's graveyard from anywhere, you may have that player lose 2 life. If you do, you gain 2 life.",
	})
	for _, accept := range []bool{false, true} {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		source := addCombatPermanent(g, game.Player1, def)
		card := addCardToGraveyard(g, game.Player2, evidenceCard("Event card", 1))
		obj := &game.StackObject{Kind: game.StackTriggeredAbility, Controller: game.Player1,
			SourceID: source.ObjectID, HasTriggerEvent: true,
			TriggerEvent: game.Event{Kind: game.EventZoneChanged, CardID: card, Player: game.Player2,
				Controller: game.Player3, FromZone: zone.Library, ToZone: zone.Graveyard}}
		agents := [game.NumPlayers]PlayerAgent{game.Player1: &scopedMayAgent{accept: []bool{accept}}}
		NewEngine(nil).resolveAbilityContentWithChoices(g, obj, def.TriggeredAbilities[0].Content, agents, &TurnLog{})
		for player, state := range g.Players {
			want := 40
			if accept && game.PlayerID(player) == game.Player1 {
				want = 42
			}
			if accept && game.PlayerID(player) == game.Player2 {
				want = 38
			}
			if state.Life != want {
				t.Fatalf("accept=%v player %d life=%d, want %d", accept, player, state.Life, want)
			}
		}
	}
}

func TestCompiledSharedPlayerRecipientGroups(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		text       string
		hand, life int
		library    int
		all        bool
	}{
		{"Each opponent loses 2 life.", 0, 38, 3, false},
		{"Each player draws two cards.", 2, 40, 1, true},
		{"Each opponent mills two cards.", 0, 40, 1, false},
	} {
		t.Run(tt.text, func(t *testing.T) {
			t.Parallel()
			def := legacyCard(t, carddata.ScryfallCard{
				Name: "Group Recipients", Layout: "normal", TypeLine: "Instant", OracleText: tt.text,
			})
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			for player := range g.Players {
				for range 3 {
					addCardToLibrary(g, game.PlayerID(player), evidenceCard("Group library", 1))
				}
			}
			NewEngine(nil).resolveAbilityContentWithChoices(g, &game.StackObject{Kind: game.StackSpell, Controller: game.Player1},
				def.SpellAbility.Val, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
			for player, state := range g.Players {
				hand, life, library := tt.hand, tt.life, tt.library
				if !tt.all && game.PlayerID(player) == game.Player1 {
					hand, life, library = 0, 40, 3
				}
				if state.Hand.Size() != hand || state.Life != life || state.Library.Size() != library {
					t.Fatalf("player %d: hand/life/library=%d/%d/%d, want %d/%d/%d",
						player, state.Hand.Size(), state.Life, state.Library.Size(), hand, life, library)
				}
			}
		})
	}
}

func TestCompiledSharedPlayerRecipientConditionalOptionalAction(t *testing.T) {
	t.Parallel()
	def := legacyCard(t, carddata.ScryfallCard{
		Name: "Scoped Recipients", Layout: "normal", TypeLine: "Instant",
		OracleText: "If you control a creature, you may draw a card and gain 2 life. You lose 1 life.",
	})
	for _, tt := range []struct {
		name                  string
		creature, accept      bool
		life, hand, decisions int
	}{
		{"accepted", true, true, 41, 1, 1},
		{"declined", true, false, 39, 0, 1},
		{"condition false", false, true, 39, 0, 0},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			addCardToLibrary(g, game.Player1, evidenceCard("Optional draw", 1))
			if tt.creature {
				addCombatCreaturePermanent(g, game.Player1)
			}
			agent := &scopedMayAgent{accept: []bool{tt.accept}}
			agents := [game.NumPlayers]PlayerAgent{game.Player1: agent}
			NewEngine(nil).resolveAbilityContentWithChoices(g, &game.StackObject{Kind: game.StackSpell, Controller: game.Player1},
				def.SpellAbility.Val, agents, &TurnLog{})
			if g.Players[game.Player1].Life != tt.life || g.Players[game.Player1].Hand.Size() != tt.hand || agent.next != tt.decisions {
				t.Fatalf("life/hand/decisions=%d/%d/%d, want %d/%d/%d", g.Players[game.Player1].Life,
					g.Players[game.Player1].Hand.Size(), agent.next, tt.life, tt.hand, tt.decisions)
			}
			for player := game.Player2; player < game.NumPlayers; player++ {
				if g.Players[player].Life != 40 || g.Players[player].Hand.Size() != 0 {
					t.Fatal("controller's optional action leaked to another player")
				}
			}
		})
	}
}

func TestCompiledSharedPlayerRecipientStackController(t *testing.T) {
	t.Parallel()
	def := legacyCard(t, carddata.ScryfallCard{
		Name: "Stack Recipients", Layout: "normal", TypeLine: "Instant",
		OracleText: "Counter target spell. Counter target spell. Its controller draws a card.",
	})
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	decoy := addStackSpell(g, game.Player2, "Decoy spell", []types.Card{types.Instant})
	target := addStackSpell(g, game.Player3, "Actual spell", []types.Card{types.Instant})
	addCardToLibrary(g, game.Player2, evidenceCard("Decoy library", 1))
	addCardToLibrary(g, game.Player3, evidenceCard("Actual library", 1))
	obj := &game.StackObject{Kind: game.StackSpell, Controller: game.Player1,
		Targets:      []game.Target{game.StackObjectTarget(decoy.ID), game.StackObjectTarget(target.ID)},
		TargetCounts: []int{1, 1}}
	NewEngine(nil).resolveAbilityContentWithChoices(g, obj, def.SpellAbility.Val, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if g.Players[game.Player3].Hand.Size() != 1 || g.Players[game.Player2].Hand.Size() != 0 || g.Players[game.Player1].Hand.Size() != 0 {
		t.Fatal("countered stack target's controller was confused with another target or resolving controller")
	}
}
