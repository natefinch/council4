package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func terastodonStack(t *testing.T, count int) (*game.Game, *Engine, []*game.Permanent) {
	t.Helper()
	registry := cards.NewDefaultRegistry()
	def, err := registry.LookupOracleID("17975a20-2c13-4325-be1e-0bcda5063a2d")
	if err != nil {
		t.Fatal(err)
	}
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, def)
	var permanents []*game.Permanent
	var targets []game.Target
	for range count {
		permanent := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
			Name: "Target", Types: []types.Card{types.Artifact},
		}})
		permanents = append(permanents, permanent)
		targets = append(targets, game.PermanentTarget(permanent.ObjectID))
	}
	g.Stack.Push(&game.StackObject{
		ID: g.IDGen.Next(), Kind: game.StackTriggeredAbility,
		SourceID: source.ObjectID, SourceCardID: source.CardInstanceID, Controller: game.Player1,
		InlineTrigger: &def.TriggeredAbilities[0], Targets: targets, TargetCounts: []int{count},
	})
	return g, NewEngineWithPrograms(nil, registry.Programs()), permanents
}

func TestTerastodonProgramTargetsAndActualDepartures(t *testing.T) {
	t.Parallel()
	for count := range 4 {
		t.Run(string(rune('0'+count)), func(t *testing.T) {
			t.Parallel()
			g, engine, targets := terastodonStack(t, count)
			log := TurnLog{}
			engine.resolveTopOfStack(g, &log)
			if got := namedTokenCount(g, game.Player2, "Elephant"); got != count {
				t.Fatalf("Elephants = %d, want %d", got, count)
			}
			for _, target := range targets {
				if !g.Players[game.Player2].Graveyard.Contains(target.CardInstanceID) {
					t.Fatal("legal target was not destroyed")
				}
			}
			var simultaneous game.ObjectID
			for _, event := range g.Events {
				if event.Kind != game.EventZoneChanged || event.FromZone != zone.Battlefield || event.ToZone != zone.Graveyard {
					continue
				}
				if simultaneous == 0 {
					simultaneous = event.SimultaneousID
				}
				if event.SimultaneousID != simultaneous {
					t.Fatal("destruction was split into separate batches")
				}
			}
		})
	}
}

func TestTerastodonDoesNotRewardProtectedOrIllegalTargets(t *testing.T) {
	t.Parallel()
	for _, protection := range []string{"indestructible", "shield", "regeneration", "became creature", "left and returned"} {
		t.Run(protection, func(t *testing.T) {
			t.Parallel()
			g, engine, targets := terastodonStack(t, 2)
			protected := targets[0]
			switch protection {
			case "indestructible":
				card, _ := g.GetCardInstance(protected.CardInstanceID)
				def := *card.Def
				def.StaticAbilities = []game.StaticAbility{game.IndestructibleStaticBody}
				card.Def = &def
			case "shield":
				protected.Counters.Add(counter.Shield, 1)
			case "regeneration":
				protected.RegenerationShields = 1
			case "became creature":
				card, _ := g.GetCardInstance(protected.CardInstanceID)
				def := *card.Def
				def.Types = []types.Card{types.Creature}
				card.Def = &def
			case "left and returned":
				obj := &game.StackObject{
					ID: g.IDGen.Next(), Controller: game.Player2,
					Targets: []game.Target{game.PermanentTarget(protected.ObjectID)},
				}
				resolveInstruction(engine, g, obj, game.MovePermanent{
					Object: game.TargetPermanentReference(0), Destination: zone.Exile, PublishLinked: "blink",
				}, nil)
				resolveInstruction(engine, g, obj, game.PutOnBattlefield{
					Source: game.LinkedBattlefieldSource("blink"),
				}, nil)
				returned := permanentByCardID(g, protected.CardInstanceID)
				if returned == nil || returned.ObjectID == protected.ObjectID {
					t.Fatal("fixture did not create a new battlefield incarnation")
				}
			default:
				t.Fatalf("unknown protection fixture %q", protection)
			}

			engine.resolveTopOfStack(g, &TurnLog{})
			if got := namedTokenCount(g, game.Player2, "Elephant"); got != 1 {
				t.Fatalf("Elephants = %d, want 1", got)
			}
		})
	}
}

func TestTerastodonOptionalDestructionAtResolution(t *testing.T) {
	t.Parallel()
	g, engine, targets := terastodonStack(t, 3)
	log := TurnLog{}
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{0}}},
	}, &log)
	if len(log.Choices) != 1 || log.Choices[0].Request.Kind != game.ChoiceMay || log.Choices[0].UsedFallback {
		t.Fatal("optional destruction was not a real resolution-time choice")
	}
	for _, target := range targets {
		if permanentByCardID(g, target.CardInstanceID) == nil {
			t.Fatal("declining destruction removed a target")
		}
	}
	if namedTokenCount(g, game.Player2, "Elephant") != 0 {
		t.Fatal("declining destruction created tokens")
	}
}

func TestTerastodonUsesDepartureControllerNotOwnerOrAnnouncementController(t *testing.T) {
	t.Parallel()
	g, engine, targets := terastodonStack(t, 2)
	targets[0].Controller = game.Player3
	engine.resolveTopOfStack(g, &TurnLog{})
	if !g.Players[game.Player2].Graveyard.Contains(targets[0].CardInstanceID) ||
		namedTokenCount(g, game.Player3, "Elephant") != 1 ||
		namedTokenCount(g, game.Player2, "Elephant") != 1 {
		t.Fatal("token recipient did not use departed permanent's last controller")
	}
}

func TestTerastodonUsesActualGraveyardArrival(t *testing.T) {
	t.Parallel()
	g, engine, targets := terastodonStack(t, 2)
	g.ReplacementEffects = append(g.ReplacementEffects, game.ReplacementEffect{
		ID: g.IDGen.Next(), Description: "exile instead", MatchEvent: game.EventZoneChanged,
		MatchFromZone: true, FromZone: zone.Battlefield,
		MatchToZone: true, ToZone: zone.Graveyard, ReplaceToZone: zone.Exile,
	})
	engine.resolveTopOfStack(g, &TurnLog{})
	for _, target := range targets {
		if !g.Players[game.Player2].Exile.Contains(target.CardInstanceID) {
			t.Fatal("zone replacement did not apply")
		}
	}
	if namedTokenCount(g, game.Player2, "Elephant") != 0 {
		t.Fatal("intended graveyard destination was mistaken for an actual arrival")
	}
}

func TestTerastodonRewardsCommanderArrivalBeforeSBA(t *testing.T) {
	t.Parallel()
	g, engine, targets := terastodonStack(t, 1)
	target := targets[0]
	g.CommanderIDs[target.CardInstanceID] = true
	engine.resolveTopOfStack(g, &TurnLog{})
	if !g.Players[game.Player2].Graveyard.Contains(target.CardInstanceID) ||
		namedTokenCount(g, game.Player2, "Elephant") != 1 {
		t.Fatal("commander's real graveyard arrival did not create an Elephant")
	}
	engine.applyStateBasedActions(g)
	if !g.Players[game.Player2].CommandZone.Contains(target.CardInstanceID) ||
		namedTokenCount(g, game.Player2, "Elephant") != 1 {
		t.Fatal("command-zone SBA erased the qualifying graveyard arrival")
	}
}

func TestTerastodonTokenArrivalAndTokenReplacements(t *testing.T) {
	t.Parallel()
	g, engine, _ := terastodonStack(t, 0)
	token := &game.CardDef{CardFace: game.CardFace{Name: "Treasure", Types: []types.Card{types.Artifact}}}
	createTokenPermanentsWithChoices(engine, g, game.Player2, token, 1, false, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	var target *game.Permanent
	for _, permanent := range g.Battlefield {
		if permanent.Token {
			target = permanent
		}
	}
	addReplacementPermanent(t, g, game.Player2, tokenDoublingReplacementCardDef())
	obj, _ := g.Stack.Peek()
	obj.Targets, obj.TargetCounts = []game.Target{game.PermanentTarget(target.ObjectID)}, []int{1}
	engine.resolveTopOfStack(g, &TurnLog{})
	if !g.Players[game.Player2].Graveyard.Contains(target.ObjectID) ||
		namedTokenCount(g, game.Player2, "Elephant") != 2 {
		t.Fatal("token graveyard arrival or normal token-doubling replacement was bypassed")
	}
	engine.applyStateBasedActions(g)
	if g.Players[game.Player2].Graveyard.Contains(target.ObjectID) {
		t.Fatal("departed token did not cease to exist at SBA")
	}
}

func TestTerastodonSimultaneousDeparturesRetainOtherLeavingTriggers(t *testing.T) {
	t.Parallel()
	g, engine, targets := terastodonStack(t, 2)
	for _, target := range targets {
		card, _ := g.GetCardInstance(target.CardInstanceID)
		def := *card.Def
		def.TriggeredAbilities = []game.TriggeredAbility{{
			Trigger: game.TriggerCondition{Pattern: game.TriggerPattern{
				Event: game.EventZoneChanged, ExcludeSelf: true,
				Controller: game.TriggerControllerYou, RequirePermanentTypes: []types.Card{types.Artifact},
				MatchFromZone: true, FromZone: zone.Battlefield, MatchToZone: true, ToZone: zone.Graveyard,
			}},
			Content: game.Mode{Sequence: []game.Instruction{{
				Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()},
			}}}.Ability(),
		}}
		card.Def = &def
	}
	startLife := g.Players[game.Player2].Life
	engine.resolveTopOfStack(g, &TurnLog{})
	if g.Players[game.Player2].Life != startLife || !g.Stack.IsEmpty() {
		t.Fatal("leaves triggers were processed while the authored batch was resolving")
	}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{}) || g.Stack.Size() != 2 {
		t.Fatal("simultaneously departing sources did not see each other leave")
	}
	for !g.Stack.IsEmpty() {
		engine.resolveTopOfStack(g, &TurnLog{})
	}
	if g.Players[game.Player2].Life != startLife+2 {
		t.Fatal("departed-source triggers did not retain their controllers and bodies")
	}
}
