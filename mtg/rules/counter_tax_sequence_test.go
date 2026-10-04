package rules

import (
	"strconv"
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

func compileCounterTaxCard(t *testing.T, name, text, typeLine string) *game.CardDef {
	t.Helper()
	defs, diagnostics, err := cardgen.CompileCardDefs(&cardgen.ScryfallCard{
		Name: name, Layout: "normal", TypeLine: typeLine, OracleText: text,
	})
	if err != nil || len(diagnostics) != 0 || len(defs) != 1 {
		t.Fatalf("compile: err=%v diagnostics=%#v defs=%d", err, diagnostics, len(defs))
	}
	return defs[0]
}

func pushCounterTaxSpell(g *game.Game, def *game.CardDef, target *game.StackObject) *game.StackObject {
	cardID := addCardToHand(g, game.Player1, def)
	g.Players[game.Player1].Hand.Remove(cardID)
	resolving := &game.StackObject{
		ID: g.IDGen.Next(), Kind: game.StackSpell, SourceID: cardID,
		Controller: game.Player1, XValue: 3,
		Targets: []game.Target{game.StackObjectTarget(target.ID)}, TargetCounts: []int{1},
	}
	g.Stack.Push(resolving)
	return resolving
}

func TestCompiledCounterTaxIndependentRider(t *testing.T) {
	t.Parallel()
	for _, tax := range []string{"{3}", "{X}"} {
		for _, outer := range []string{"", "If you control a creature, "} {
			def := compileCounterTaxCard(t, "Tax Probe", outer+
				"Counter target spell unless its controller pays "+tax+". You gain 2 life.", "Instant")
			tests := []struct {
				name                         string
				mana                         int
				decline, protected, creature bool
				wantCounter                  bool
				wantSpent                    int
			}{
				{"paid", 3, false, false, true, false, 3},
				{"declined", 3, true, false, true, true, 0},
				{"cannot pay", 2, false, false, true, true, 0},
				{"counter impossible", 3, true, true, true, false, 0},
				{"counter impossible but pays", 3, false, true, true, false, 3},
				{"outer false", 3, false, false, false, false, 0},
			}
			for _, tt := range tests {
				t.Run(tax+"/"+outer+"/"+tt.name, func(t *testing.T) {
					t.Parallel()
					g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
					g.Players[game.Player1].ManaPool.Add(mana.C, 5)
					g.Players[game.Player2].ManaPool.Add(mana.C, tt.mana)
					target := addStackSpell(g, game.Player2, "Target Spell", []types.Card{types.Sorcery})
					g.CardInstances[target.SourceID].Owner = game.Player3
					if tt.protected {
						g.CardInstances[target.SourceID].Def.StaticAbilities = []game.StaticAbility{game.CantBeCounteredStaticBody}
					}
					if tt.creature {
						addCreaturePermanent(g, game.Player1)
					}
					pushCounterTaxSpell(g, def, target)
					agents := [game.NumPlayers]PlayerAgent{}
					if tt.decline {
						agents[game.Player2] = &choiceOnlyAgent{choices: [][]int{{0}}}
					}
					log := TurnLog{}
					NewEngine(nil).resolveTopOfStackWithChoices(g, agents, &log)
					wantCounter, wantSpent := tt.wantCounter, tt.wantSpent
					if !tt.creature && outer == "" {
						wantSpent = 3
					}
					if _, remains := stackObjectByID(g, target.ID); remains == wantCounter {
						t.Fatalf("spell remains=%v, want counter=%v", remains, wantCounter)
					}
					if g.Players[game.Player3].Graveyard.Contains(target.SourceID) != wantCounter {
						t.Fatal("counter did not preserve card ownership")
					}
					if got := g.Players[game.Player1].Life; got != 42 {
						t.Fatalf("independent rider life=%d, want 42", got)
					}
					if got := g.Players[game.Player2].ManaPool.Total(); got != tt.mana-wantSpent {
						t.Fatalf("target controller mana=%d, want %d", got, tt.mana-wantSpent)
					}
					if got := g.Players[game.Player1].ManaPool.Total(); got != 5 {
						t.Fatalf("ability controller paid the tax: mana=%d", got)
					}
					for _, choice := range log.Choices {
						if choice.Request.Player != game.Player2 {
							t.Fatalf("tax choice offered to %v, want target controller", choice.Request.Player)
						}
					}
					if outer != "" && !tt.creature && len(log.Choices) != 0 {
						t.Fatal("false outer condition still offered payment")
					}
				})
			}
		}
	}
}

func TestCompiledCounterTaxControllerRiderAfterCounter(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"Mindswipe", "Frightful Delusion"} {
		for _, pay := range []bool{false, true} {
			t.Run(name+"/pay="+strconv.FormatBool(pay), func(t *testing.T) {
				t.Parallel()
				text := "Counter target spell unless its controller pays {X}. Mindswipe deals X damage to that spell's controller."
				if name == "Frightful Delusion" {
					text = "Counter target spell unless its controller pays {1}. That player discards a card."
				}
				def := compileCounterTaxCard(t, name, text, "Instant")
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				target := addStackSpell(g, game.Player2, "Target Spell", []types.Card{types.Sorcery})
				g.CardInstances[target.SourceID].Owner = game.Player3
				g.Players[game.Player2].ManaPool.Add(mana.C, 3)
				discard := addCardToHand(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Discarded"}})
				resolving := pushCounterTaxSpell(g, def, target)
				agents := [game.NumPlayers]PlayerAgent{}
				if !pay {
					agents[game.Player2] = &choiceOnlyAgent{choices: [][]int{{0}, {0}}}
				}
				NewEngine(nil).resolveTopOfStackWithChoices(g, agents, &TurnLog{})
				if _, remains := stackObjectByID(g, target.ID); remains != pay {
					t.Fatalf("target remains=%v, payment=%v", remains, pay)
				}
				if name == "Mindswipe" && g.Players[game.Player2].Life != 37 {
					t.Fatalf("controller rider life=%d, want 37", g.Players[game.Player2].Life)
				}
				if name == "Frightful Delusion" && !g.Players[game.Player2].Graveyard.Contains(discard) {
					t.Fatal("controller did not discard independently of counter success")
				}
				if !pay && resolving.TargetControllerLKI[0] != game.Player2 {
					t.Fatal("counter lost target controller LKI")
				}
				if g.Players[game.Player3].Life != 40 {
					t.Fatal("rider affected spell owner instead of controller")
				}
			})
		}
	}
}

func TestCompiledCounterTaxEventPayment(t *testing.T) {
	t.Parallel()
	text := "Whenever this enchantment becomes the target of a spell or ability, counter that spell or ability unless its controller pays {3}. You gain 2 life."
	def := compileCounterTaxCard(t, "Tax Watcher", text, "Enchantment")
	for _, pay := range []bool{false, true} {
		t.Run("pay="+strconv.FormatBool(pay), func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			source := addCombatPermanent(g, game.Player1, def)
			target := addStackSpell(g, game.Player2, "Target Spell", []types.Card{types.Sorcery})
			g.Players[game.Player2].ManaPool.Add(mana.C, 3)
			g.Stack.Push(&game.StackObject{
				ID: g.IDGen.Next(), Kind: game.StackTriggeredAbility, Controller: game.Player1,
				SourceID: source.ObjectID, SourceCardID: source.CardInstanceID,
				InlineTrigger: &def.TriggeredAbilities[0], HasTriggerEvent: true,
				TriggerEvent: game.Event{StackObjectID: target.ID},
			})
			agents := [game.NumPlayers]PlayerAgent{}
			if !pay {
				agents[game.Player2] = &choiceOnlyAgent{choices: [][]int{{0}}}
			}
			NewEngine(nil).resolveTopOfStackWithChoices(g, agents, &TurnLog{})
			if _, remains := stackObjectByID(g, target.ID); remains != pay {
				t.Fatalf("event spell remains=%v, payment=%v", remains, pay)
			}
			if g.Players[game.Player1].Life != 42 {
				t.Fatal("event tax skipped independent rider")
			}
		})
	}
}
