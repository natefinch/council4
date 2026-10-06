package rules

import (
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestCompiledReferenceSubjectKillerEnteredIdentity(t *testing.T) {
	def := compiledReferenceSubjectOriginal(t, "Killer Instinct")
	sequence := def.TriggeredAbilities[0].Content.Modes[0].Sequence
	grant, ok := sequence[2].Primitive.(game.ApplyContinuous)
	if !ok || !grant.Object.Exists || grant.Object.Val != game.LinkedObjectReference("sequence-effect-1-product") {
		t.Errorf("complete CardDef haste must name the entered product; primitive=%T object=%v", sequence[2].Primitive, grant.Object)
	}
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player2, def)
	source.Owner = game.Player1
	g.CardInstances[source.CardInstanceID].Owner = game.Player1
	creature := referenceSubjectCreature(t)
	observed := addCardToLibrary(g, game.Player2, creature)
	decoy := addCardToLibrary(g, game.Player1, creature)
	rival := addCombatPermanent(g, game.Player3, creature)
	engine := NewEngine(nil)
	g.AppendEvent(game.Event{Kind: game.EventBeginningOfStep, Step: game.StepUpkeep, Player: game.Player2, Controller: game.Player2})
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("complete compiled upkeep trigger did not fire")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	entered := permanentForCard(g, observed)
	if entered == nil {
		t.Fatal("qualifying observed card did not actually enter")
	}
	view := findPermanentView(t, observe(g, game.Player2), entered.ObjectID)
	if !view.HasKeyword(game.Haste) {
		t.Error("actual entered creature did not gain haste")
	}
	if entered.Owner != game.Player2 || entered.Controller != game.Player2 ||
		!g.Players[game.Player1].Library.Contains(decoy) ||
		findPermanentView(t, observe(g, game.Player2), rival.ObjectID).HasKeyword(game.Haste) {
		t.Error("entry/haste used a decoy, source owner, or wrong controller")
	}
	t.Logf("source=%d observedCard=%d enteredObject=%d version=%d haste=%v",
		source.ObjectID, observed, entered.ObjectID, g.CardInstances[observed].ZoneVersion, view.HasKeyword(game.Haste))
}

func TestCompiledReferenceSubjectDeceiverSourceIdentity(t *testing.T) {
	for _, tc := range []struct {
		name    string
		keyword game.Keyword
	}{
		{"Callous Deceiver", game.Flying},
		{"Feral Deceiver", game.Trample},
		{"Brutal Deceiver", game.FirstStrike},
	} {
		t.Run(tc.name, func(t *testing.T) {
			def := compiledReferenceSubjectOriginal(t, tc.name)
			if len(def.ActivatedAbilities) != 2 {
				t.Fatal("complete original lost an activated ability")
			}
			sequence := def.ActivatedAbilities[1].Content.Modes[0].Sequence
			pt, ptOK := sequence[1].Primitive.(game.ModifyPT)
			grant, grantOK := sequence[2].Primitive.(game.ApplyContinuous)
			if !ptOK || pt.Object != game.SourcePermanentReference() ||
				!grantOK || !grant.Object.Exists || grant.Object.Val != pt.Object {
				t.Errorf("complete CardDef PT and keyword must share original source: PT=%v keyword=%v",
					pt.Object, grant.Object)
			}
			for _, lifecycle := range []string{"live", "reentered", "clone reentered"} {
				t.Run(lifecycle, func(t *testing.T) {
					g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
					controller := game.Player2
					g.Turn.PriorityPlayer = controller
					g.Players[controller].ManaPool.Add(mana.C, 2)
					source := addCombatPermanent(g, controller, def)
					source.Owner = game.Player1
					g.CardInstances[source.CardInstanceID].Owner = game.Player1
					decoy := addCombatPermanent(g, controller, def)
					forest := compileUnlessCard(t, cardgen.ScryfallCard{
						Name: "Forest", Layout: "normal", TypeLine: "Basic Land \u2014 Forest", OracleText: "{T}: Add {G}.",
					})
					observed := addCardToLibrary(g, controller, forest)
					engine := NewEngine(nil)
					obj := activateOrdinal(t, g, engine, source.ObjectID, 1, nil, nil)
					if obj.SourceID != source.ObjectID || obj.SourceCardID != source.CardInstanceID {
						t.Fatal("actual activation did not capture exact source identity")
					}
					if lifecycle == "clone reentered" {
						g = g.Clone()
						obj, _ = g.Stack.Peek()
					}
					recipient := source
					if lifecycle != "live" {
						if !movePermanentToZone(g, permanentForCard(g, source.CardInstanceID), zone.Exile) {
							t.Fatal("actual source departure failed")
						}
						r := effectResolver{game: g, engine: engine, obj: obj}
						var entered bool
						recipient, entered = r.putResolvedCardOnBattlefieldValue(g.CardInstances[source.CardInstanceID],
							zone.Exile, controller, nil, permanentCreationOptions{})
						if !entered || recipient.ObjectID == source.ObjectID {
							t.Fatal("actual reentry did not create a different incarnation")
						}
					}
					engine.resolveTopOfStack(g, &TurnLog{})
					view := findPermanentView(t, observe(g, controller), recipient.ObjectID)
					want := lifecycle == "live"
					if view.HasKeyword(tc.keyword) != want {
						t.Errorf("keyword=%v want %v; original ObjectID=%d reached ObjectID=%d",
							view.HasKeyword(tc.keyword), want, source.ObjectID, recipient.ObjectID)
					}
					if findPermanentView(t, observe(g, controller), decoy.ObjectID).HasKeyword(tc.keyword) ||
						!g.Players[controller].Library.Contains(observed) {
						t.Error("keyword/observation used an independent source or wrong card")
					}
					t.Logf("lifecycle=%s source=%d reached=%d card=%d version=%d keyword=%v",
						lifecycle, source.ObjectID, recipient.ObjectID, source.CardInstanceID,
						g.CardInstances[source.CardInstanceID].ZoneVersion, view.HasKeyword(tc.keyword))
				})
			}
		})
	}
}

func TestCompiledReferenceSubjectSelfDeathCardIdentity(t *testing.T) {
	for _, name := range []string{"Doombot Harbinger", "Flitting Guerrilla"} {
		t.Run(name, func(t *testing.T) {
			def := compiledReferenceSubjectOriginal(t, name)
			for _, lifecycle := range []string{"accept", "decline", "stale card"} {
				t.Run(lifecycle, func(t *testing.T) {
					g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
					source := addCombatPermanent(g, game.Player2, def)
					source.Owner = game.Player1
					g.CardInstances[source.CardInstanceID].Owner = game.Player1
					decoy := addCombatPermanent(g, game.Player3, def)
					reward := addCardToGraveyard(g, game.Player2, referenceSubjectCreature(t))
					engine := NewEngine(nil)
					if _, ok := destroyPermanent(g, source.ObjectID); !ok || !engine.putTriggeredAbilitiesOnStack(g) {
						t.Fatal("complete original self-death trigger did not fire")
					}
					agent := &libraryPaymentAgent{accept: lifecycle != "decline"}
					if lifecycle == "stale card" {
						if !moveCardBetweenZones(g, source.Owner, source.CardInstanceID, zone.Graveyard, zone.Hand) ||
							!moveCardBetweenZones(g, source.Owner, source.CardInstanceID, zone.Hand, zone.Graveyard) {
							t.Fatal("actual card incarnation change failed")
						}
					}
					resolveStackWithTriggers(engine, g, [game.NumPlayers]PlayerAgent{game.Player2: agent})
					wantExile := lifecycle == "accept"
					if g.Players[source.Owner].Exile.Contains(source.CardInstanceID) != wantExile ||
						g.Players[source.Owner].Graveyard.Contains(source.CardInstanceID) == wantExile {
						t.Errorf("self-death %s: exile=%v graveyard=%v, want exile=%v",
							lifecycle, g.Players[source.Owner].Exile.Contains(source.CardInstanceID),
							g.Players[source.Owner].Graveyard.Contains(source.CardInstanceID), wantExile)
					}
					if permanentForCard(g, decoy.CardInstanceID) == nil {
						t.Error("self-card operation moved another source")
					}
					if name == "Doombot Harbinger" &&
						g.Players[game.Player2].Hand.Contains(reward) != wantExile {
						t.Error("self-death consequence did not return the controller's chosen card")
					}
					if name == "Flitting Guerrilla" &&
						g.Players[game.Player2].Library.Contains(reward) != wantExile {
						t.Error("self-death consequence did not put the controller's chosen card on the library")
					}
					t.Logf("self-death=%s originalObject=%d card=%d version=%d exile=%v prompts=%d",
						lifecycle, source.ObjectID, source.CardInstanceID, g.CardInstances[source.CardInstanceID].ZoneVersion,
						g.Players[source.Owner].Exile.Contains(source.CardInstanceID), len(agent.prompts))
				})
			}
		})
	}
}
