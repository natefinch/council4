package rules

import (
	"fmt"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

const counterOutcomeExile = "If that spell is countered this way, exile it instead of putting it into its owner's graveyard."

func TestCompiledCounterDestinationAndActualOutcome(t *testing.T) {
	t.Parallel()
	for _, destination := range []struct {
		name, text string
		zone       zone.Type
	}{
		{"graveyard", "", zone.Graveyard},
		{"exile", counterOutcomeExile, zone.Exile},
		{"hand", "If that spell is countered this way, put it into its owner's hand instead of into that player's graveyard.", zone.Hand},
		{"library", "If that spell is countered this way, put it on top of its owner's library instead of into that player's graveyard.", zone.Library},
	} {
		for _, tax := range []string{"{3}", "{X}"} {
			def := compileCounterTaxCard(t, "Counter Outcome Probe",
				"If you control a creature, counter target spell unless its controller pays "+tax+". "+
					destination.text+" If that spell is countered this way, you gain 2 life. You gain 1 life.", "Instant")
			for _, tt := range []struct {
				name                string
				mana                int
				decline, protected  bool
				creature, succeeded bool
				spent               int
			}{
				{"paid", 3, false, false, true, false, 3},
				{"declined", 3, true, false, true, true, 0},
				{"cannot pay", 2, false, false, true, true, 0},
				{"uncounterable declined", 3, true, true, true, false, 0},
				{"uncounterable paid", 3, false, true, true, false, 3},
				{"skipped counter", 3, false, false, false, false, 0},
			} {
				t.Run(destination.name+"/"+tax+"/"+tt.name, func(t *testing.T) {
					t.Parallel()
					g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
					target := addStackSpell(g, game.Player2, "Target Spell", []types.Card{types.Sorcery})
					card := g.CardInstances[target.SourceID]
					card.Owner = game.Player3
					version := card.ZoneVersion
					if tt.protected {
						card.Def.StaticAbilities = []game.StaticAbility{game.CantBeCounteredStaticBody}
					}
					if tt.creature {
						addCreaturePermanent(g, game.Player1)
					}
					g.Players[game.Player2].ManaPool.Add(mana.C, tt.mana)
					pushCounterTaxSpell(g, def, target)
					agents := [game.NumPlayers]PlayerAgent{}
					if tt.decline {
						agents[game.Player2] = &choiceOnlyAgent{choices: [][]int{{0}}}
					}
					engine := NewEngine(nil)
					log := TurnLog{}
					engine.resolveTopOfStackWithChoices(g, agents, &log)
					wantLife := 41
					if tt.succeeded {
						wantLife += 2
					}
					if g.Players[game.Player1].Life != wantLife {
						t.Fatalf("life = %d, want %d; success=%v", g.Players[game.Player1].Life, wantLife, tt.succeeded)
					}
					if _, remains := stackObjectByID(g, target.ID); remains == tt.succeeded {
						t.Fatalf("target remains=%v; success=%v", remains, tt.succeeded)
					}
					if got := g.Players[game.Player2].ManaPool.Total(); got != tt.mana-tt.spent {
						t.Fatalf("payer mana=%d, want %d", got, tt.mana-tt.spent)
					}
					moves := 0
					for _, event := range g.Events {
						if event.Kind != game.EventZoneChanged || event.CardID != card.ID {
							continue
						}
						moves++
						if !tt.succeeded || event.FromZone != zone.Stack || event.ToZone != destination.zone {
							t.Fatalf("counter zone event = %#v", event)
						}
					}
					wantMoves := 0
					if tt.succeeded {
						wantMoves = 1
					}
					if moves != wantMoves || card.ZoneVersion != version+uint64(wantMoves) {
						t.Fatalf("moves=%d version=%d, want %d/%d", moves, card.ZoneVersion, wantMoves, version+uint64(wantMoves))
					}
					if tt.succeeded {
						switch destination.zone {
						case zone.Graveyard:
							if !g.Players[game.Player3].Graveyard.Contains(card.ID) {
								t.Fatal("wrong graveyard owner")
							}
						case zone.Exile:
							if !g.Players[game.Player3].Exile.Contains(card.ID) {
								t.Fatal("wrong exile owner")
							}
						case zone.Hand:
							if !g.Players[game.Player3].Hand.Contains(card.ID) {
								t.Fatal("wrong hand owner")
							}
						case zone.Library:
							top, ok := g.Players[game.Player3].Library.Top()
							if !ok || top != card.ID {
								t.Fatal("wrong library top or owner")
							}
						default:
							t.Fatalf("unexpected counter destination: %v", destination.zone)
						}
					} else if target.ExileOnResolution || target.CounteredDestination != game.CounteredSpellGraveyard {
						t.Fatal("failed/skipped counter leaked its destination into later resolution")
					}
					if tt.protected {
						engine.resolveTopOfStack(g, &TurnLog{})
						if !g.Players[game.Player3].Graveyard.Contains(card.ID) {
							t.Fatal("uncounterable spell's later resolution was redirected")
						}
					}
				})
			}
		}
	}
}

func TestCompiledIndependentCounterOutcomesAndOccurrences(t *testing.T) {
	t.Parallel()
	text := "Exile target creature. Counter target spell. " + counterOutcomeExile +
		" If that spell is countered this way, you gain 2 life. Counter target spell. " +
		"If that spell is countered this way, put it into its owner's hand instead of into that player's graveyard. " +
		"If that spell is countered this way, you gain 5 life. You gain 1 life."
	def := compileCounterTaxCard(t, "Independent Counter Probe", text, "Instant")
	for _, firstProtected := range []bool{false, true} {
		for _, secondProtected := range []bool{false, true} {
			t.Run(fmt.Sprintf("%v/%v", firstProtected, secondProtected), func(t *testing.T) {
				t.Parallel()
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				permanent := addCreaturePermanent(g, game.Player4)
				first := addStackSpell(g, game.Player2, "First", []types.Card{types.Sorcery})
				second := addStackSpell(g, game.Player3, "Second", []types.Card{types.Sorcery})
				g.CardInstances[first.SourceID].Owner = game.Player4
				for _, pair := range []struct {
					object    *game.StackObject
					protected bool
				}{{first, firstProtected}, {second, secondProtected}} {
					if pair.protected {
						g.CardInstances[pair.object.SourceID].Def.StaticAbilities = []game.StaticAbility{game.CantBeCounteredStaticBody}
					}
				}
				resolving := pushCounterTaxSpell(g, def, first)
				resolving.Targets = []game.Target{
					game.PermanentTarget(permanent.ObjectID),
					game.StackObjectTarget(first.ID), game.StackObjectTarget(second.ID),
				}
				resolving.TargetCounts = []int{1, 1, 1}
				NewEngine(nil).resolveTopOfStack(g, &TurnLog{})
				wantLife := 41
				if !firstProtected {
					wantLife += 2
				}
				if !secondProtected {
					wantLife += 5
				}
				if g.Players[game.Player1].Life != wantLife || !g.Players[game.Player4].Exile.Contains(permanent.CardInstanceID) {
					t.Fatalf("life=%d, want %d; independent exile was lost or rebound", g.Players[game.Player1].Life, wantLife)
				}
				if g.Players[game.Player4].Exile.Contains(first.SourceID) == firstProtected ||
					g.Players[game.Player3].Hand.Contains(second.SourceID) == secondProtected {
					t.Fatal("independent destination/outcome cross-bound")
				}
			})
		}
	}
}

func TestCompiledCounterCopyPublishesSuccessWithoutMovingOriginal(t *testing.T) {
	t.Parallel()
	def := compileCounterTaxCard(t, "Copy Counter Probe",
		"Counter target spell. "+counterOutcomeExile+" If that spell is countered this way, you gain 2 life.", "Instant")
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	original := addStackSpell(g, game.Player2, "Original", []types.Card{types.Sorcery})
	copyObject := *original
	copyObject.ID, copyObject.Copy = g.IDGen.Next(), true
	g.Stack.Push(&copyObject)
	pushCounterTaxSpell(g, def, &copyObject)
	NewEngine(nil).resolveTopOfStack(g, &TurnLog{})
	if _, exists := stackObjectByID(g, copyObject.ID); exists {
		t.Fatal("countered copy remains")
	}
	if _, exists := stackObjectByID(g, original.ID); !exists || original.ExileOnResolution {
		t.Fatal("countered copy changed original identity")
	}
	if g.Players[game.Player1].Life != 42 {
		t.Fatal("success on a countered copy was not published")
	}
	for _, event := range g.Events {
		if event.Kind == game.EventZoneChanged && event.CardID == original.SourceID {
			t.Fatalf("copy fabricated a physical card move: %#v", event)
		}
	}
}

func TestCompiledCounterGoneTargetDoesNotPublishSuccess(t *testing.T) {
	t.Parallel()
	def := compileCounterTaxCard(t, "Gone Counter Probe",
		"Tap target creature. Counter target spell. "+counterOutcomeExile+
			" If that spell is countered this way, you gain 2 life. You gain 1 life.", "Instant")
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	permanent := addCreaturePermanent(g, game.Player4)
	target := addStackSpell(g, game.Player2, "Gone Spell", []types.Card{types.Sorcery})
	resolving := pushCounterTaxSpell(g, def, target)
	resolving.Targets = []game.Target{game.PermanentTarget(permanent.ObjectID), game.StackObjectTarget(target.ID)}
	resolving.TargetCounts = []int{1, 1}
	if !counterStackObject(g, target.ID) {
		t.Fatal("could not remove counter target")
	}
	source := g.CardInstances[resolving.SourceID]
	source.Owner = game.Player3
	NewEngine(nil).resolveTopOfStack(g, &TurnLog{})
	if g.Players[game.Player1].Life != 41 || !permanent.Tapped {
		t.Fatal("gone target ran success consequence or suppressed an independent rider")
	}
	if !g.Players[game.Player2].Graveyard.Contains(target.SourceID) ||
		!g.Players[game.Player3].Graveyard.Contains(source.ID) ||
		g.Players[game.Player2].Exile.Contains(target.SourceID) {
		t.Fatal("counter modifier confused target and resolving source identities")
	}
}

func TestCounterEventDestinationFailureDoesNotLeak(t *testing.T) {
	t.Parallel()
	for _, protected := range []bool{false, true} {
		t.Run(fmt.Sprint(protected), func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			target := addStackSpell(g, game.Player2, "Event Target", []types.Card{types.Sorcery})
			if protected {
				g.CardInstances[target.SourceID].Def.StaticAbilities = []game.StaticAbility{game.CantBeCounteredStaticBody}
			}
			source := addCombatPermanent(g, game.Player1, &game.CardDef{
				CardFace: game.CardFace{Name: "Event Source", Types: []types.Card{types.Artifact}},
			})
			trigger := game.TriggeredAbility{Content: game.Mode{Sequence: []game.Instruction{
				{Primitive: game.CounterObject{Object: game.EventStackObjectReference(), ExileInstead: true}, PublishResult: "counter"},
				{Primitive: game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(2)},
					ResultGate: opt.Val(game.InstructionResultGate{Key: "counter", Succeeded: game.TriTrue})},
			}}.Ability()}
			g.Stack.Push(&game.StackObject{
				ID: g.IDGen.Next(), Kind: game.StackTriggeredAbility, Controller: game.Player1,
				SourceID: source.ObjectID, SourceCardID: source.CardInstanceID,
				InlineTrigger: &trigger, HasTriggerEvent: true,
				TriggerEvent: game.Event{StackObjectID: target.ID},
			})
			engine := NewEngine(nil)
			engine.resolveTopOfStack(g, &TurnLog{})
			if protected {
				if target.ExileOnResolution || g.Players[game.Player1].Life != 40 {
					t.Fatal("failed event counter leaked destination or success")
				}
				engine.resolveTopOfStack(g, &TurnLog{})
				if !g.Players[game.Player2].Graveyard.Contains(target.SourceID) {
					t.Fatal("failed event counter redirected normal resolution")
				}
			} else if g.Players[game.Player1].Life != 42 || !g.Players[game.Player2].Exile.Contains(target.SourceID) {
				t.Fatal("event counter did not publish actual success or apply destination")
			}
		})
	}
}
