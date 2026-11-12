package rules

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/cards/authored"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review: CR effective 2025-11-14 (official text from
// media.wizards.com, "MagicCompRules 20251114.txt") 122.1h: one or more
// finality counters create a single replacement effect, "If this permanent
// would be put into a graveyard from the battlefield, exile it instead."
// CR 614.1a/614.5/616.1 (replacement application, at most once, affected
// object's controller orders), 700.4 (dies), 603.6c/603.10a (leaves-the-
// battlefield triggers look back), 111.7 (tokens cease to exist), 903.9a
// (commander SBA option from exile). Rite of the Moth (DSK 229,
// a7fea839-d7d7-48b4-aeaf-133767532e32) rulings 2024-09-20 from
// https://api.scryfall.com/cards/0e79123f-accd-4193-8c72-b750e0d6f3fa/rulings:
// finality counters work on any permanent; other destinations such as hand
// are unaffected; they grant no ability, so ability loss does not stop the
// exile; multiple finality counters are redundant. Rite of the Moth itself is
// not admitted here; riteOfTheMothShape plays its spell shape to prove the
// shared rule.

// riteOfTheMothShape mirrors Rite of the Moth's spell: "Return target creature
// card from your graveyard to the battlefield with a finality counter on it."
// with Flashback {3}{W}{W}{B}.
func riteOfTheMothShape() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Black),
		CardFace: game.CardFace{
			Name:     "Rite of the Moth shape",
			ManaCost: opt.Val(cost.Mana{cost.O(1), cost.W, cost.B, cost.B}),
			Colors:   []color.Color{color.White, color.Black},
			Types:    []types.Card{types.Sorcery},
			StaticAbilities: []game.StaticAbility{{
				Text: "Flashback {3}{W}{W}{B}",
				KeywordAbilities: []game.KeywordAbility{
					game.FlashbackKeyword{Cost: cost.Mana{cost.O(3), cost.W, cost.W, cost.B}},
				},
			}},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{{
					MinTargets: 1, MaxTargets: 1,
					Constraint: "target creature card from your graveyard",
					Allow:      game.TargetAllowCard,
					TargetZone: zone.Graveyard,
					Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
				}},
				Sequence: []game.Instruction{{
					Primitive: game.PutOnBattlefield{
						Source:        game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget}),
						EntryCounters: []game.CounterPlacement{{Kind: counter.Finality, Amount: 1}},
					},
				}},
			}.Ability()),
		},
	}
}

var finalityRegistry = cards.NewDefaultRegistry()

func finalityRegistryCard(t *testing.T, name string) *game.CardDef {
	t.Helper()
	def := finalityRegistry.Lookup(name)
	if def == nil {
		t.Fatalf("production %s is missing", name)
	}
	return def
}

func finalityMainPhase() (*game.Game, *Engine) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	setMainPhasePriority(g, game.Player1)
	return g, NewEngineWithPrograms(nil, finalityRegistry.Programs())
}

func addFinalityPermanent(g *game.Game, controller game.PlayerID, def *game.CardDef, finality int) *game.Permanent {
	permanent := addCombatPermanent(g, controller, def)
	permanent.Counters.Add(counter.Finality, finality)
	return permanent
}

func castAndResolveFinalitySpell(t *testing.T, g *game.Game, engine *Engine, caster game.PlayerID, def *game.CardDef, from zone.Type, targets []game.Target) id.ID {
	t.Helper()
	setMainPhasePriority(g, caster)
	var spellID id.ID
	if from == zone.Graveyard {
		spellID = addCardToGraveyard(g, caster, def)
	} else {
		spellID = addCardToHand(g, caster, def)
	}
	g.Players[caster].ManaPool.Add(mana.W, 3)
	g.Players[caster].ManaPool.Add(mana.B, 3)
	g.Players[caster].ManaPool.Add(mana.G, 3)
	g.Players[caster].ManaPool.Add(mana.R, 3)
	g.Players[caster].ManaPool.Add(mana.U, 3)
	if !engine.applyAction(g, caster, action.CastSpellFromZone(spellID, from, targets, 0, nil)) {
		t.Fatalf("casting %s from %v failed", def.Name, from)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	return spellID
}

func requireFinalityExiled(t *testing.T, g *game.Game, permanent *game.Permanent) {
	t.Helper()
	if _, ok := permanentByObjectID(g, permanent.ObjectID); ok {
		t.Fatal("finality permanent is still on the battlefield")
	}
	owner := g.Players[permanent.Owner]
	if owner.Graveyard.Contains(permanent.CardInstanceID) {
		t.Fatal("finality permanent was put into its owner's graveyard")
	}
	if !owner.Exile.Contains(permanent.CardInstanceID) {
		t.Fatal("finality permanent was not exiled")
	}
	for _, event := range g.Events {
		if event.Kind == game.EventPermanentDied && event.PermanentID == permanent.ObjectID {
			t.Fatal("finality permanent emitted a dies event")
		}
	}
}

func TestFinalityCounterRiteShapeReturnsWithOneCounterFromHandAndFlashback(t *testing.T) {
	for _, from := range []zone.Type{zone.Hand, zone.Graveyard} {
		g, engine := finalityMainPhase()
		bearID := addCardToGraveyard(g, game.Player1, finalityRegistryCard(t, "Grizzly Bears"))
		riteID := castAndResolveFinalitySpell(t, g, engine, game.Player1, riteOfTheMothShape(), from,
			[]game.Target{game.CardTarget(bearID)})
		var bear *game.Permanent
		for _, permanent := range g.Battlefield {
			if permanent.CardInstanceID == bearID {
				bear = permanent
			}
		}
		if bear == nil {
			t.Fatalf("cast from %v: target creature card was not returned", from)
		}
		if got := bear.Counters.Get(counter.Finality); got != 1 {
			t.Fatalf("cast from %v: finality counters = %d, want 1", from, got)
		}
		if from == zone.Graveyard && !g.Players[game.Player1].Exile.Contains(riteID) {
			t.Fatal("flashback-cast spell was not exiled after resolving")
		}
	}
}

func TestFinalityCounterMurderExilesWithoutDiesButWithLeavesTrigger(t *testing.T) {
	g, engine := finalityMainPhase()
	bear := addFinalityPermanent(g, game.Player1, finalityRegistryCard(t, "Grizzly Bears"), 2)
	bloodArtist := addCombatPermanent(g, game.Player2, finalityRegistryCard(t, "Blood Artist"))
	cosmographer := addCombatPermanent(g, game.Player1, finalityRegistryCard(t, "Imperial Cosmographer"))

	castAndResolveFinalitySpell(t, g, engine, game.Player2, finalityRegistryCard(t, "Murder"), zone.Hand,
		[]game.Target{game.PermanentTarget(bear.ObjectID)})
	requireFinalityExiled(t, g, bear)

	engine.putTriggeredAbilitiesOnStackWithChoices(g, simPassPolicies(), &TurnLog{})
	if n := triggersFromSource(g, bloodArtist.ObjectID); n != 0 {
		t.Fatalf("Blood Artist dies triggers = %d, want 0", n)
	}
	if n := triggersFromSource(g, cosmographer.ObjectID); n != 1 {
		t.Fatalf("leaves-without-dying observer triggers = %d, want 1", n)
	}
}

// finalitySacrificeAgent pays a sacrifice cost with the named creature and,
// when replacement is set, applies the replacement whose label contains it.
type finalitySacrificeAgent struct {
	simPassPolicy
	name         string
	replacement  string
	replacements [][]string
}

func (a *finalitySacrificeAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	if request.Kind == game.ChoiceReplacement {
		var labels []string
		for _, option := range request.Options {
			labels = append(labels, option.Label)
		}
		a.replacements = append(a.replacements, labels)
		for _, option := range request.Options {
			if a.replacement != "" && strings.Contains(option.Label, a.replacement) {
				return []int{option.Index}
			}
		}
	}
	if request.Kind == game.ChoicePayment {
		for _, option := range request.Options {
			if strings.Contains(option.Label, a.name) {
				return []int{option.Index}
			}
		}
	}
	return request.DefaultSelection
}

func TestFinalityCounterLethalDamageAndSacrificeExile(t *testing.T) {
	t.Run("Shock", func(t *testing.T) {
		g, engine := finalityMainPhase()
		bear := addFinalityPermanent(g, game.Player1, finalityRegistryCard(t, "Grizzly Bears"), 1)
		castAndResolveFinalitySpell(t, g, engine, game.Player2, finalityRegistryCard(t, "Shock"), zone.Hand,
			[]game.Target{game.PermanentTarget(bear.ObjectID)})
		engine.applyStateBasedActions(g)
		requireFinalityExiled(t, g, bear)
	})
	t.Run("Viscera Seer", func(t *testing.T) {
		g, engine := finalityMainPhase()
		seer := addCombatPermanent(g, game.Player1, finalityRegistryCard(t, "Viscera Seer"))
		bear := addFinalityPermanent(g, game.Player1, finalityRegistryCard(t, "Grizzly Bears"), 1)
		addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Library card"}})
		agents := simPassPolicies()
		agents[game.Player1] = &finalitySacrificeAgent{name: "Grizzly Bears"}
		if !engine.applyActionWithChoices(g, game.Player1, action.ActivateAbility(seer.ObjectID, 0, nil, 0), agents, &TurnLog{}) {
			t.Fatal("Viscera Seer activation failed")
		}
		if _, ok := permanentByObjectID(g, seer.ObjectID); !ok {
			t.Fatal("Viscera Seer sacrificed itself instead of the finality creature")
		}
		requireFinalityExiled(t, g, bear)
	})
}

// TestFinalityCounterCompetesWithSelfReplacementByControllerChoice sacrifices a
// finality-countered Darksteel Colossus: its own shuffle replacement and the
// finality replacement both apply, and its controller chooses (CR 616.1).
func TestFinalityCounterCompetesWithSelfReplacementByControllerChoice(t *testing.T) {
	for _, test := range []struct {
		choose string
		want   zone.Type
	}{
		{choose: "Finality counter", want: zone.Exile},
		{choose: "shuffle", want: zone.Library},
	} {
		t.Run(test.choose, func(t *testing.T) {
			g, engine := finalityMainPhase()
			seer := addCombatPermanent(g, game.Player1, finalityRegistryCard(t, "Viscera Seer"))
			colossus := addFinalityPermanent(g, game.Player1, finalityRegistryCard(t, "Darksteel Colossus"), 1)
			addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Library card"}})
			agent := &finalitySacrificeAgent{name: "Darksteel Colossus", replacement: test.choose}
			agents := simPassPolicies()
			agents[game.Player1] = agent
			if !engine.applyActionWithChoices(g, game.Player1, action.ActivateAbility(seer.ObjectID, 0, nil, 0), agents, &TurnLog{}) {
				t.Fatal("Viscera Seer activation failed")
			}
			if len(agent.replacements) != 1 || len(agent.replacements[0]) != 2 {
				t.Fatalf("replacement choices = %q, want one choice between two effects", agent.replacements)
			}
			player := g.Players[game.Player1]
			inExile := player.Exile.Contains(colossus.CardInstanceID)
			inLibrary := player.Library.Contains(colossus.CardInstanceID)
			if player.Graveyard.Contains(colossus.CardInstanceID) || inExile != (test.want == zone.Exile) ||
				inLibrary != (test.want == zone.Library) {
				t.Fatalf("Colossus exile=%v library=%v, want %v", inExile, inLibrary, test.want)
			}
		})
	}
}

func TestFinalityCounterArtifactAndTokenExile(t *testing.T) {
	g, engine := finalityMainPhase()
	fabricator := addFinalityPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Fabricator artifact", Types: []types.Card{types.Artifact},
	}}, 1)
	castAndResolveFinalitySpell(t, g, engine, game.Player2, finalityRegistryCard(t, "Naturalize"), zone.Hand,
		[]game.Target{game.PermanentTarget(fabricator.ObjectID)})
	requireFinalityExiled(t, g, fabricator)

	token := addTokenCreaturePermanent(g, game.Player1, "Finality token")
	token.Counters.Add(counter.Finality, 1)
	castAndResolveFinalitySpell(t, g, engine, game.Player2, finalityRegistryCard(t, "Murder"), zone.Hand,
		[]game.Target{game.PermanentTarget(token.ObjectID)})
	engine.applyStateBasedActions(g)
	if _, ok := permanentByObjectID(g, token.ObjectID); ok {
		t.Fatal("finality token is still on the battlefield")
	}
	player := g.Players[game.Player1]
	if player.Graveyard.Contains(token.ObjectID) || player.Exile.Contains(token.ObjectID) {
		t.Fatal("finality token did not cease to exist after leaving the battlefield")
	}
	sawExile := false
	for _, event := range g.Events {
		if event.PermanentID != token.ObjectID {
			continue
		}
		if event.Kind == game.EventPermanentDied {
			t.Fatal("finality token died")
		}
		if event.Kind == game.EventZoneChanged && event.FromZone == zone.Battlefield {
			sawExile = event.ToZone == zone.Exile
		}
	}
	if !sawExile {
		t.Fatal("finality token did not move from the battlefield to exile")
	}
}

func TestFinalityCounterOtherDestinationsAndRemovedCounter(t *testing.T) {
	g, engine := finalityMainPhase()
	bounced := addFinalityPermanent(g, game.Player1, finalityRegistryCard(t, "Grizzly Bears"), 1)
	castAndResolveFinalitySpell(t, g, engine, game.Player2, finalityRegistryCard(t, "Unsummon"), zone.Hand,
		[]game.Target{game.PermanentTarget(bounced.ObjectID)})
	if !g.Players[game.Player1].Hand.Contains(bounced.CardInstanceID) {
		t.Fatal("Unsummon did not return the finality creature to its owner's hand")
	}

	removed := addFinalityPermanent(g, game.Player1, finalityRegistryCard(t, "Grizzly Bears"), 1)
	removed.Counters.Remove(counter.Finality, 1)
	castAndResolveFinalitySpell(t, g, engine, game.Player2, finalityRegistryCard(t, "Murder"), zone.Hand,
		[]game.Target{game.PermanentTarget(removed.ObjectID)})
	if !g.Players[game.Player1].Graveyard.Contains(removed.CardInstanceID) {
		t.Fatal("creature whose finality counter was removed did not go to the graveyard")
	}
	if countEventsOfKind(g.Events, game.EventPermanentDied) != 1 {
		t.Fatal("creature without a finality counter did not die")
	}
}

func TestFinalityCounterSurvivesHumilityAbilityLoss(t *testing.T) {
	g, engine := finalityMainPhase()
	addCombatPermanent(g, game.Player2, authored.Humility.New())
	seer := addFinalityPermanent(g, game.Player1, finalityRegistryCard(t, "Thalakos Seer"), 1)
	if values := effectivePermanentValues(g, seer); len(values.abilities) != 0 || values.toughness != 1 {
		t.Fatalf("Humility did not remove abilities and set 1/1: %+v", values)
	}
	castAndResolveFinalitySpell(t, g, engine, game.Player2, finalityRegistryCard(t, "Shock"), zone.Hand,
		[]game.Target{game.PermanentTarget(seer.ObjectID)})
	engine.applyStateBasedActions(g)
	requireFinalityExiled(t, g, seer)
}

func TestFinalityCounterCommanderExiledThenCommandZoneSBA(t *testing.T) {
	g, engine := finalityMainPhase()
	commander := addFinalityPermanent(g, game.Player1, finalityRegistryCard(t, "Grizzly Bears"), 1)
	g.CommanderIDs[commander.CardInstanceID] = true
	g.Players[game.Player1].CommanderInstanceID = commander.CardInstanceID

	castAndResolveFinalitySpell(t, g, engine, game.Player2, finalityRegistryCard(t, "Murder"), zone.Hand,
		[]game.Target{game.PermanentTarget(commander.ObjectID)})
	requireFinalityExiled(t, g, commander)
	engine.applyStateBasedActions(g)
	player := g.Players[game.Player1]
	if player.Exile.Contains(commander.CardInstanceID) || !player.CommandZone.Contains(commander.CardInstanceID) {
		t.Fatal("exiled commander was not moved to the command zone by the CR 903.9a option")
	}
}

func TestFinalityCounterSurvivesCloneAndReplaysIdentically(t *testing.T) {
	g, engine := finalityMainPhase()
	bear := addFinalityPermanent(g, game.Player1, finalityRegistryCard(t, "Grizzly Bears"), 1)
	clone := g.Clone()
	cloned, ok := permanentByObjectID(clone, bear.ObjectID)
	if !ok || cloned.Counters.Get(counter.Finality) != 1 {
		t.Fatal("finality counter did not survive the game clone")
	}
	murder := finalityRegistryCard(t, "Murder")
	for _, run := range []*game.Game{g, clone} {
		castAndResolveFinalitySpell(t, run, engine, game.Player2, murder, zone.Hand,
			[]game.Target{game.PermanentTarget(bear.ObjectID)})
		requireFinalityExiled(t, run, bear)
	}
	if len(g.Events) != len(clone.Events) {
		t.Fatalf("original and cloned replays emitted %d and %d events", len(g.Events), len(clone.Events))
	}
	for i := range g.Events {
		if g.Events[i].Kind != clone.Events[i].Kind || g.Events[i].ToZone != clone.Events[i].ToZone {
			t.Fatalf("event %d differs between original and cloned replay", i)
		}
	}
}
