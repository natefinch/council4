package cardgen

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestSacrificeEventGeneratedNaturalRuntimeRoundTrip(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping generated-package runtime round-trip in short mode")
	}
	cards := []*ScryfallCard{{
		Name: "It That Betrays", Layout: "normal", TypeLine: "Creature — Eldrazi",
		ManaCost: "{12}", Power: new("11"), Toughness: new("11"),
		OracleText: "Annihilator 2 (Whenever this creature attacks, defending player sacrifices two permanents of their choice.)\n" +
			"Whenever an opponent sacrifices a nontoken permanent, put that card onto the battlefield under your control.",
	}, {
		Name: "RT Owner Watcher", Layout: "normal", TypeLine: "Creature", Power: new("1"), Toughness: new("1"),
		OracleText: "Whenever an opponent sacrifices a nontoken permanent, put that card onto the battlefield.",
	}, {
		Name: "RT Sacrifice Subject", Layout: "normal", TypeLine: "Creature", Power: new("2"), Toughness: new("3"),
		OracleText: "At the beginning of your upkeep, sacrifice this creature.",
	}}
	dir, pkg := writeCardRoundTripPackage(t, cards)
	source := fmt.Sprintf(`package %s

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/rules"
)

type passAgent struct{}

func (passAgent) ChooseAction(rules.PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (passAgent) ChooseChoice(_ rules.PlayerObservation, request game.ChoiceRequest) []int {
	return request.DefaultSelection
}

func addPermanent(g *game.Game, def *game.CardDef, owner, controller game.PlayerID) *game.Permanent {
	card := g.IDGen.Next()
	g.CardInstances[card] = &game.CardInstance{ID: card, Def: def, Owner: owner}
	permanent := &game.Permanent{ObjectID: g.IDGen.Next(), CardInstanceID: card, Owner: owner, Controller: controller}
	g.Battlefield = append(g.Battlefield, permanent)
	return permanent
}

func TestGeneratedNaturalOpponentSacrificeAndRecipient(t *testing.T) {
	for _, explicit := range []bool{false, true} {
		t.Run(map[bool]string{false:"default owner",true:"printed controller"}[explicit], func(t *testing.T) {
			watcher := RTOwnerWatcher()
			if explicit { watcher = ItThatBetrays() }
			subject := RTSacrificeSubject()
			for _, def := range []*game.CardDef{watcher, subject} {
				if issues := game.ValidateCardDef(def); len(issues) != 0 {
					t.Fatalf("generated CardDef invalid: %%+v", issues)
				}
			}
			found := false
			for _, ability := range watcher.TriggeredAbilities {
				if ability.Trigger.Pattern.Event != game.EventPermanentSacrificed { continue }
				put, ok := ability.Content.Modes[0].Sequence[0].Primitive.(game.PutOnBattlefield)
				if !ok || put.Source != game.CardBattlefieldSource(game.CardReference{Kind:game.CardReferenceEvent}) ||
					put.Recipient.Exists != explicit || explicit && put.Recipient.Val != game.ControllerReference() {
					t.Fatal("generated exact event source or clause-owned recipient lost")
				}
				found = true
			}
			if !found { t.Fatal("generated sacrifice ability missing") }
			deck := make([]*game.CardDef, 20)
			for i := range deck { deck[i] = &game.CardDef{CardFace:game.CardFace{Name:"Filler"}} }
			var configs [game.NumPlayers]game.PlayerConfig
			var agents [game.NumPlayers]rules.PlayerAgent
			for i := range configs { configs[i].Deck = deck; agents[i] = passAgent{} }
			engine := rules.NewEngine(nil)
			g := engine.NewGame(configs)
			source := addPermanent(g, watcher, game.Player4, game.Player2)
			victim := addPermanent(g, subject, game.Player3, game.Player1)
			decoy := g.IDGen.Next()
			g.CardInstances[decoy] = &game.CardInstance{ID:decoy, Def:subject, Owner:game.Player3}
			g.Players[game.Player3].Graveyard.Add(decoy)
			engine.RunGameWithTurnLimit(g, agents, 1)
			var returned *game.Permanent
			for _, permanent := range g.Battlefield {
				if permanent.CardInstanceID == victim.CardInstanceID { returned = permanent }
			}
			want := game.Player3
			if explicit { want = game.Player2 }
			if returned == nil || returned.Controller != want || returned.Owner != game.Player3 ||
				returned.ObjectID == victim.ObjectID || !g.Players[game.Player3].Graveyard.Contains(decoy) {
				t.Fatal("natural generated gameplay did not return exact card under printed/default controller")
			}
			var events []game.Event
			for _, event := range g.Events {
				if event.Kind == game.EventPermanentSacrificed && event.CardID == victim.CardInstanceID { events = append(events,event) }
			}
			if len(events) != 1 || events[0].PermanentID != victim.ObjectID || events[0].CardZoneVersion == 0 ||
				events[0].Controller != game.Player1 || events[0].Player != game.Player1 ||
				!events[0].TriggeredAbilitiesCaptured || len(events[0].TriggeredAbilities) != 1 {
				t.Fatal("natural sacrifice did not snapshot exact once-only physical event and watcher")
			}
			trigger := events[0].TriggeredAbilities[0]
			if trigger.SourceID != source.ObjectID || trigger.Controller != game.Player2 ||
				trigger.SourceCardID != source.CardInstanceID {
				t.Fatal("owner/event player replaced generated ability controller/source")
			}
			if card := g.CardInstances[victim.CardInstanceID]; card.ZoneVersion <= events[0].CardZoneVersion {
				t.Fatal("battlefield return did not create a later incarnation than immutable sacrifice event")
			}
			var ids []id.ID
			for _, permanent := range g.Battlefield { if permanent.CardInstanceID == victim.CardInstanceID { ids = append(ids,permanent.ObjectID) } }
			if len(ids) != 1 { t.Fatal("one event produced duplicate physical permanents") }
		})
	}
}
`, pkg)
	if err := os.WriteFile(filepath.Join(dir, "semantic_test.go"), []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	cmd := exec.CommandContext(context.Background(), "go", "test", "-v", "-count=1", "-timeout=60s", "./"+filepath.Base(dir))
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("sacrifice-event generated runtime round-trip: %v\n%s", err, output)
	}
	t.Logf("generated constructor runtime:\n%s", output)
}
