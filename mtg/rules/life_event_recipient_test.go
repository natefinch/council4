package rules

import (
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestCompiledLifeEventPlayerIsQualifiedPermanentController(t *testing.T) {
	for _, tt := range []struct {
		name, text           string
		subject              types.Card
		dies, optional       bool
		loss, controllerGain int
	}{
		{"Polluted Bonds", "Whenever a land an opponent controls enters, that player loses 2 life and you gain 2 life.", types.Land, false, false, 2, 2},
		{"Assault Intercessor", "First strike, menace\nChainsword — Whenever a creature an opponent controls dies, that player loses 2 life.", types.Creature, true, false, 2, 0},
		{"Blood Seeker", "Whenever a creature an opponent controls enters, you may have that player lose 1 life.", types.Creature, false, true, 1, 0},
		{"Nightshade Harvester", "Whenever a land an opponent controls enters, that player loses 1 life. Put a +1/+1 counter on this creature.", types.Land, false, false, 1, 0},
		{"Massacre Wurm", "When this creature enters, creatures your opponents control get -2/-2 until end of turn.\nWhenever a creature an opponent controls dies, that player loses 2 life.", types.Creature, true, false, 2, 0},
		{"Suture Priest", "Whenever another creature you control enters, you may gain 1 life.\nWhenever a creature an opponent controls enters, you may have that player lose 1 life.", types.Creature, false, true, 1, 0},
	} {
		t.Run(tt.name, func(t *testing.T) {
			def := compileUnlessCard(t, cardgen.ScryfallCard{
				Name: tt.name, Layout: "normal", TypeLine: "Creature", Power: new("4"), Toughness: new("4"), OracleText: tt.text,
			})
			for _, controlled := range []bool{false, true} {
				for _, accept := range []bool{false, true} {
					g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
					engine := NewEngine(nil)
					source := addCombatPermanent(g, game.Player1, def)
					subject := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
						Name: "Event subject", Types: []types.Card{tt.subject},
					}})
					subject.Controller = game.Player3
					if controlled {
						subject.Controller = game.Player1
					}
					if tt.dies {
						destroyPermanent(g, subject.ObjectID)
					} else {
						emitEvent(g, game.Event{Kind: game.EventPermanentEnteredBattlefield, PermanentID: subject.ObjectID,
							CardID: subject.CardInstanceID, Controller: subject.Controller, Player: game.Player2})
					}
					triggered := engine.putTriggeredAbilitiesOnStack(g)
					if controlled {
						// Suture Priest's separate positive trigger may fire.
						if tt.name != "Suture Priest" && triggered {
							t.Fatal("controlled subject passed opponent-controller qualifier")
						}
						continue
					}
					if !triggered {
						t.Fatal("qualified opponent-controlled subject did not trigger")
					}
					resolveInstruction(engine, g, &game.StackObject{Controller: game.Player1, SourceID: source.ObjectID},
						game.MovePermanent{Object: game.SourcePermanentReference(), Destination: zone.Exile}, nil)
					if !tt.dies {
						resolveInstruction(engine, g, &game.StackObject{Controller: game.Player3, SourceID: subject.ObjectID},
							game.MovePermanent{Object: game.SourcePermanentReference(), Destination: zone.Exile}, nil)
					}
					agents := [game.NumPlayers]PlayerAgent{game.Player1: &scopedMayAgent{accept: []bool{accept}}}
					engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
					for player, state := range g.Players {
						want := 40
						if !tt.optional || accept {
							if game.PlayerID(player) == game.Player1 {
								want += tt.controllerGain
							}
							if game.PlayerID(player) == game.Player3 {
								want -= tt.loss
							}
						}
						if state.Life != want {
							t.Fatalf("accept=%v player%d life=%d, want%d; owner2 must not become controller3", accept, player, state.Life, want)
						}
					}
				}
			}
		})
	}
}

func TestCompiledLifeMonarchPlayerUsesEventPlayer(t *testing.T) {
	def := compileUnlessCard(t, cardgen.ScryfallCard{
		Name: "Knights of the Black Rose", Layout: "normal", TypeLine: "Creature",
		Power: new("4"), Toughness: new("4"),
		OracleText: "When this creature enters, you become the monarch.\n" +
			"Whenever an opponent becomes the monarch, if you were the monarch as the turn began, that player loses 2 life and you gain 2 life.",
	})
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, def)
	obj := &game.StackObject{Controller: game.Player1, SourceID: source.ObjectID, Kind: game.StackTriggeredAbility,
		HasTriggerEvent: true, TriggerEvent: game.Event{Kind: game.EventBecameMonarch, Player: game.Player2, Controller: game.Player3}}
	NewEngine(nil).resolveAbilityContentWithChoices(g, obj, def.TriggeredAbilities[1].Content, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if g.Players[game.Player2].Life != 38 || g.Players[game.Player1].Life != 42 || g.Players[game.Player3].Life != 40 {
		t.Fatal("monarch event's player was confused with event controller")
	}
}
