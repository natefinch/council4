package cardgen

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestTargetAnnouncementGeneratedNaturalRuntimeRoundTrip(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping generated-package runtime round-trip in short mode")
	}
	dir, pkg := writeCardRoundTripPackage(t, []*ScryfallCard{{
		Name: "Nadu, Winged Wisdom", Layout: "normal", TypeLine: "Legendary Creature — Bird Wizard",
		ManaCost: "{1}{G}{U}", Power: new("3"), Toughness: new("4"),
		OracleText: `Flying
Creatures you control have "Whenever this creature becomes the target of a spell or ability, reveal the top card of your library. If it's a land card, put it onto the battlefield. Otherwise, put it into your hand. This ability triggers only twice each turn."`,
	}, {
		Name: "RT Targeting Sacrifice", Layout: "normal", TypeLine: "Artifact",
		OracleText: "Sacrifice a creature: Target creature gets +1/+1 until end of turn.",
	}, {
		Name: "RT Repeat Target", Layout: "normal", TypeLine: "Artifact",
		OracleText: "{0}: Target creature gets +1/+1 until end of turn.",
	}, {
		Name: "RT Sacrifice Grant", Layout: "normal", TypeLine: "Artifact",
		OracleText: "Sacrifice a creature: Draw a card.",
	}, {
		Name: "RT Return Grant", Layout: "normal", TypeLine: "Instant", ManaCost: "{0}",
		OracleText: "Return target creature card from your graveyard to the battlefield.",
	}})
	source := fmt.Sprintf(`package %s

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/rules"
	"github.com/natefinch/council4/opt"
)

type announcementAgent struct {
	source, recipient id.ID
	costLabel string
	activated bool
}

func (a *announcementAgent) ChooseAction(_ rules.PlayerObservation, legal []action.Action) action.Action {
	if !a.activated {
		for _, candidate := range legal {
			body, ok := candidate.ActivateAbilityPayload()
			if ok && body.SourceID == a.source && len(body.Targets) == 1 &&
				body.Targets[0].Kind == game.TargetPermanent && body.Targets[0].PermanentID == a.recipient {
				a.activated = true
				return candidate
			}
		}
	}
	return action.Pass()
}

func (a *announcementAgent) ChooseChoice(_ rules.PlayerObservation, request game.ChoiceRequest) []int {
	for _, option := range request.Options {
		if a.costLabel != "" && strings.Contains(option.Label, a.costLabel) {
			return []int{option.Index}
		}
	}
	return request.DefaultSelection
}

func addAnnouncementPermanent(g *game.Game, def *game.CardDef, owner, controller game.PlayerID) *game.Permanent {
	card := g.IDGen.Next()
	g.CardInstances[card] = &game.CardInstance{ID:card, Def:def, Owner:owner}
	permanent := &game.Permanent{ObjectID:g.IDGen.Next(), CardInstanceID:card, Owner:owner, Controller:controller}
	g.Battlefield = append(g.Battlefield, permanent)
	return permanent
}

func TestGeneratedNaturalAnnouncementThenPayment(t *testing.T) {
	for _, sacrificeRecipient := range []bool{false,true} {
		t.Run(map[bool]string{false:"grant departure",true:"recipient departure"}[sacrificeRecipient], func(t *testing.T) {
			grant, actuatorDef := NaduWingedWisdom(), RTTargetingSacrifice()
			for _, def := range []*game.CardDef{grant,actuatorDef} {
				if issues := game.ValidateCardDef(def); len(issues)!=0 { t.Fatalf("invalid generated CardDef: %%+v",issues) }
			}
			grants := 0
			for _, static := range grant.StaticAbilities {
				for _, effect := range static.ContinuousEffects {
					for _, body := range effect.AddAbilities {
						trigger, ok := body.(*game.TriggeredAbility)
						if ok && trigger.Trigger.Pattern.Event == game.EventObjectBecameTarget &&
							trigger.Trigger.Pattern.Source == game.TriggerSourceSelf && trigger.MaxTriggersPerTurn == 2 {
							grants++
						}
					}
				}
			}
			if grants != 1 { t.Fatal("generated canonical recipient-local target grant/limit missing") }
			deck := make([]*game.CardDef,20)
			for i := range deck { deck[i]= &game.CardDef{CardFace:game.CardFace{Name:"Filler"}} }
			var configs [game.NumPlayers]game.PlayerConfig
			var agents [game.NumPlayers]rules.PlayerAgent
			for i:=range configs { configs[i].Deck=deck; agents[i]=&announcementAgent{} }
			engine := rules.NewEngine(nil)
			g := engine.NewGame(configs)
			nadu := addAnnouncementPermanent(g,grant,game.Player4,game.Player2)
			recipient := addAnnouncementPermanent(g,&game.CardDef{CardFace:game.CardFace{
				Name:"Original Recipient",Types:[]types.Card{types.Creature},
				Power:opt.Val(game.PT{Value:2}),Toughness:opt.Val(game.PT{Value:3}),
			}},game.Player1,game.Player2)
			actuator := addAnnouncementPermanent(g,actuatorDef,game.Player3,game.Player2)
			observed := g.IDGen.Next()
			g.CardInstances[observed]=&game.CardInstance{ID:observed,Owner:game.Player2,Def:&game.CardDef{CardFace:game.CardFace{Name:"Exact Observation"}}}
			g.Players[game.Player2].Library.Add(observed)
			agent := &announcementAgent{source:actuator.ObjectID,recipient:recipient.ObjectID,costLabel:"Nadu"}
			departed := nadu
			if sacrificeRecipient { agent.costLabel="Original Recipient"; departed=recipient }
			agents[game.Player2]=agent
			engine.RunGameWithTurnLimit(g,agents,1)
			if !agent.activated || !g.Players[departed.Owner].Graveyard.Contains(departed.CardInstanceID) ||
				!g.Players[game.Player2].Hand.Contains(observed) {
				t.Fatal("generated natural activation/payment did not retain exact observation after departure")
			}
			var targetEvents, activations []game.Event
			for _, event:=range g.Events {
				if event.Kind==game.EventObjectBecameTarget && event.PermanentID==recipient.ObjectID { targetEvents=append(targetEvents,event) }
				if event.Kind==game.EventAbilityActivated && event.PermanentID==actuator.ObjectID { activations=append(activations,event) }
			}
			if len(targetEvents)!=1 || len(activations)!=1 ||
				targetEvents[0].StackObjectID!=activations[0].StackObjectID ||
				!targetEvents[0].TriggeredAbilitiesCaptured || len(targetEvents[0].TriggeredAbilities)!=1 {
				t.Fatal("natural generated action duplicated or lost the exact announced target occurrence")
			}
			snapshot:=targetEvents[0].TriggeredAbilities[0]
			if snapshot.SourceID!=recipient.ObjectID || snapshot.SourceCardID!=recipient.CardInstanceID ||
				snapshot.Controller!=game.Player2 {
				t.Fatal("grant source/physical owner replaced recipient-local captured body/controller")
			}
		})
	}
}

type incarnationAgent struct {
	g *game.Game
	repeat, sacrifice, recipient, returnCard id.ID
	stage int
}

func (a *incarnationAgent) ChooseAction(_ rules.PlayerObservation, legal []action.Action) action.Action {
	if a.g.Stack.Size()!=0 || a.stage>=7 { return action.Pass() }
	for _,candidate:=range legal {
		if a.stage==3 {
			if cast,ok:=candidate.CastSpellPayload();ok && cast.CardID==a.returnCard {
				a.stage++;return candidate
			}
			continue
		}
		body,ok:=candidate.ActivateAbilityPayload()
		if !ok { continue }
		if a.stage==2 && body.SourceID==a.sacrifice { a.stage++;return candidate }
		if a.stage!=2 && body.SourceID==a.repeat && len(body.Targets)==1 &&
			body.Targets[0].PermanentID==a.recipient { a.stage++;return candidate }
	}
	return action.Pass()
}
func (*incarnationAgent) ChooseChoice(_ rules.PlayerObservation,r game.ChoiceRequest) []int {
	for _,option:=range r.Options { if strings.Contains(option.Label,"Nadu") { return []int{option.Index} } }
	return r.DefaultSelection
}

func TestGeneratedNaturalGrantReentryFreshTwoAndThirdRefused(t *testing.T) {
	deck:=make([]*game.CardDef,40)
	for i:=range deck { deck[i]=&game.CardDef{CardFace:game.CardFace{Name:"Nonland Observation"}} }
	var configs [game.NumPlayers]game.PlayerConfig
	var agents [game.NumPlayers]rules.PlayerAgent
	for i:=range configs { configs[i].Deck=deck;agents[i]=&announcementAgent{} }
	e:=rules.NewEngine(nil)
	g:=e.NewGame(configs)
	nadu:=addAnnouncementPermanent(g,NaduWingedWisdom(),game.Player2,game.Player2)
	recipient:=addAnnouncementPermanent(g,&game.CardDef{CardFace:game.CardFace{
		Name:"Exact Recipient",Types:[]types.Card{types.Creature},
		Power:opt.Val(game.PT{Value:2}),Toughness:opt.Val(game.PT{Value:3}),
	}},game.Player1,game.Player2)
	repeat:=addAnnouncementPermanent(g,RTRepeatTarget(),game.Player4,game.Player2)
	sacrifice:=addAnnouncementPermanent(g,RTSacrificeGrant(),game.Player3,game.Player2)
	cardID:=g.IDGen.Next()
	g.CardInstances[cardID]=&game.CardInstance{ID:cardID,Owner:game.Player2,Def:RTReturnGrant()}
	g.Players[game.Player2].Hand.Add(cardID)
	agent:=&incarnationAgent{g:g,repeat:repeat.ObjectID,sacrifice:sacrifice.ObjectID,
		recipient:recipient.ObjectID,returnCard:cardID}
	agents[game.Player2]=agent
	e.RunGameWithTurnLimit(g,agents,1)
	var returned *game.Permanent
	for _,p:=range g.Battlefield { if p.CardInstanceID==nadu.CardInstanceID { returned=p } }
	if agent.stage!=7 || returned==nil || returned.ObjectID==nadu.ObjectID {
		t.Fatal("normal generated legal action/sacrifice/cast did not complete exact source reentry")
	}
	old,newCount:=0,0
	for _,event:=range g.Events {
		if event.Kind!=game.EventObjectBecameTarget || event.PermanentID!=recipient.ObjectID { continue }
		if len(event.TriggeredAbilities)!=1 { t.Fatal("target event did not snapshot one grant") }
		origin:=event.TriggeredAbilities[0].TriggerOrigin()
		if origin.SourceID==nadu.ObjectID { old++ }
		if origin.SourceID==returned.ObjectID { newCount++ }
	}
	if old!=2 || newCount!=3 { t.Fatal("actual targeting did not preserve old versus new granting occurrence") }
	// Four recipient trigger observations plus the sacrifice actuator's draw;
	// the new grant's third announced target must not resolve an observation.
	reveals:=0
	for _,event:=range g.Events { if event.Kind==game.EventCardRevealed && event.Controller==game.Player2 { reveals++ } }
	if reveals!=4 { t.Fatalf("recipient observed %%d cards, want exactly four trigger uses across two grants",reveals) }
}
`, pkg)
	if err := os.WriteFile(filepath.Join(dir, "announcement_test.go"), []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	cmd := exec.CommandContext(context.Background(), "go", "test", "-v", "-count=1", "-timeout=60s", "./"+filepath.Base(dir))
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("target-announcement generated runtime round-trip: %v\n%s", err, output)
	}
	t.Logf("generated constructor runtime:\n%s", output)
}
