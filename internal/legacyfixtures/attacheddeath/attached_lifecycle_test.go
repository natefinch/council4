// Legacy regression fixture preserved when the Oracle compiler was retired.
package fixtures

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/mtg/rules"
	"github.com/natefinch/council4/opt"
)

func fixturePermanent(g *game.Game, def *game.CardDef, owner, controller game.PlayerID) *game.Permanent {
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{ID: cardID, Def: def, Owner: owner}
	p := &game.Permanent{ObjectID: g.IDGen.Next(), CardInstanceID: cardID, Owner: owner, Controller: controller}
	g.Battlefield = append(g.Battlefield, p)
	return p
}

func fixtureGame() (*rules.Engine, *game.Game) {
	deck := make([]*game.CardDef, 30)
	for i := range deck {
		deck[i] = &game.CardDef{CardFace: game.CardFace{Name: "Filler"}}
	}
	var configs [game.NumPlayers]game.PlayerConfig
	for i := range configs {
		configs[i].Deck = deck
	}
	e := rules.NewEngine(nil)
	return e, e.NewGame(configs)
}

func hostDef(name string, subtype types.Sub) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{Name: name, Types: []types.Card{types.Creature},
		Subtypes: []types.Sub{subtype}, Power: opt.Val(game.PT{Value: 2}), Toughness: opt.Val(game.PT{Value: 3})}}
}

type lifecycleAgent struct {
	g                      *game.Game
	sacrificeSource, quest id.ID
	activated              bool
}

func (a *lifecycleAgent) ChooseAction(_ rules.PlayerObservation, legal []action.Action) action.Action {
	if a.g != nil && !a.activated {
		queued := false
		for _, obj := range a.g.Stack.Objects() {
			if obj.Kind == game.StackTriggeredAbility && obj.SourceID == a.quest {
				queued = true
			}
		}
		if queued {
			for _, candidate := range legal {
				body, ok := candidate.ActivateAbilityPayload()
				if ok && body.SourceID == a.sacrificeSource {
					a.activated = true
					return candidate
				}
			}
		}
	}
	return action.Pass()
}

func (*lifecycleAgent) ChooseChoice(_ rules.PlayerObservation, r game.ChoiceRequest) []int {
	for _, option := range r.Options {
		if strings.Contains(option.Label, "Quest for Ula") || strings.Contains(option.Label, "Exact Hand Card") {
			return []int{option.Index}
		}
	}
	return r.DefaultSelection
}

func TestGeneratedGhoulishNaturalHostDeathSBADelayedAttachment(t *testing.T) {
	e, g := fixtureGame()
	def := GhoulishImpetus()
	if issues := game.ValidateCardDef(def); len(issues) != 0 {
		t.Fatalf("invalid generated Aura: %+v", issues)
	}
	aura := fixturePermanent(g, def, game.Player3, game.Player2)
	host := fixturePermanent(g, hostDef("Original Host", types.Human), game.Player4, game.Player1)
	survivor := fixturePermanent(g, hostDef("Legal New Host", types.Human), game.Player4, game.Player1)
	aura.AttachedTo = opt.Val(host.ObjectID)
	host.Attachments = []id.ID{aura.ObjectID}
	host.MarkedDamage = 100
	var agents [game.NumPlayers]rules.PlayerAgent
	for i := range agents {
		agents[i] = &lifecycleAgent{}
	}
	e.RunGameWithTurnLimit(g, agents, 1)
	var returned *game.Permanent
	for _, p := range g.Battlefield {
		if p.CardInstanceID == aura.CardInstanceID {
			returned = p
		}
	}
	if returned == nil || returned.ObjectID == aura.ObjectID || returned.Owner != game.Player3 ||
		returned.Controller != game.Player3 || !returned.AttachedTo.Exists || returned.AttachedTo.Val != survivor.ObjectID {
		t.Fatal("generated natural completed host death/Aura SBA/delayed owner return/legal attachment failed")
	}
	captured := 0
	for _, event := range g.Events {
		if event.Kind == game.EventPermanentDied && event.PermanentID == host.ObjectID {
			if event.CardID != host.CardInstanceID || event.CardZoneVersion == 0 {
				t.Fatal("death payload lost completed host incarnation")
			}
			for _, body := range event.TriggeredAbilities {
				if body.SourceID == aura.ObjectID {
					captured++
					if body.SourceCardID != aura.CardInstanceID || body.Controller != game.Player2 ||
						body.SourceZoneVersion != 0 {
						t.Fatal("host event capture was rebound or mutated by later departure")
					}
				}
			}
		}
	}
	if captured != 1 {
		t.Fatal("original watcher was lost or duplicated")
	}
}

func TestGeneratedGhoulishNaturalNoLegalHostKeepsOriginalGraveyardIncarnation(t *testing.T) {
	e, g := fixtureGame()
	aura := fixturePermanent(g, GhoulishImpetus(), game.Player3, game.Player2)
	host := fixturePermanent(g, hostDef("Only Host", types.Human), game.Player4, game.Player1)
	aura.AttachedTo = opt.Val(host.ObjectID)
	host.Attachments = []id.ID{aura.ObjectID}
	host.MarkedDamage = 100
	var agents [game.NumPlayers]rules.PlayerAgent
	for i := range agents {
		agents[i] = &lifecycleAgent{}
	}
	e.RunGameWithTurnLimit(g, agents, 1)
	departures, entries := 0, 0
	var departedVersion uint64
	for _, event := range g.Events {
		if event.CardID != aura.CardInstanceID {
			continue
		}
		if event.Kind == game.EventZoneChanged && event.FromZone == zone.Battlefield && event.ToZone == zone.Graveyard {
			departures++
			departedVersion = event.CardZoneVersion
		}
		if event.Kind == game.EventPermanentEnteredBattlefield {
			entries++
		}
	}
	if departures != 1 || entries != 0 || departedVersion == 0 ||
		g.CardInstances[aura.CardInstanceID].ZoneVersion != departedVersion ||
		!g.Players[game.Player3].Graveyard.Contains(aura.CardInstanceID) {
		t.Fatal("normal generated delayed no-host refusal changed the completed original graveyard incarnation")
	}
	for _, p := range g.Battlefield {
		if p.CardInstanceID == aura.CardInstanceID {
			t.Fatal("no-host refusal created a battlefield incarnation")
		}
	}
}

func TestGeneratedQuestNaturalQueuedSourceDeparture(t *testing.T) {
	for _, subtype := range []types.Sub{types.Kraken, types.Leviathan, types.Octopus, types.Serpent} {
		t.Run(string(subtype), func(t *testing.T) {
			e, g := fixtureGame()
			questDef, actuatorDef := QuestForUlaSTemple(), RTEnchantmentSacrifice()
			for _, def := range []*game.CardDef{questDef, actuatorDef} {
				if issues := game.ValidateCardDef(def); len(issues) != 0 {
					t.Fatalf("invalid generated CardDef: %+v", issues)
				}
			}
			quest := fixturePermanent(g, questDef, game.Player1, game.Player2)
			quest.Counters.Add(counter.Quest, 3)
			actuator := fixturePermanent(g, actuatorDef, game.Player4, game.Player2)
			cardID := g.IDGen.Next()
			g.CardInstances[cardID] = &game.CardInstance{ID: cardID, Owner: game.Player2, Def: hostDef("Exact Hand Card", subtype)}
			g.Players[game.Player2].Hand.Add(cardID)
			var agents [game.NumPlayers]rules.PlayerAgent
			for i := range agents {
				agents[i] = &lifecycleAgent{}
			}
			agent := &lifecycleAgent{g: g, sacrificeSource: actuator.ObjectID, quest: quest.ObjectID}
			agents[game.Player2] = agent
			e.RunGameWithTurnLimit(g, agents, 1)
			if !agent.activated || !g.Players[game.Player1].Graveyard.Contains(quest.CardInstanceID) {
				t.Fatal("generated legal sacrifice did not depart original source after actual end-step queue")
			}
			found := false
			for _, p := range g.Battlefield {
				if p.CardInstanceID == cardID {
					found = p.Owner == game.Player2 && p.Controller == game.Player2
				}
			}
			if !found {
				t.Fatal("generated resolution lost original-source LKI condition or controller hand")
			}
		})
	}
}
