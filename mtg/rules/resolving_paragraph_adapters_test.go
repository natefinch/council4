package rules

import (
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestCompiledResolvingParagraphDamageAndPT(t *testing.T) {
	for _, test := range []struct {
		name, text  string
		base, bonus int
		damage      bool
	}{
		{"Ember Volley", "Ember Volley deals 3 damage to any target.\nMorbid — Ember Volley deals 5 damage instead if a creature died this turn.", 3, 5, true},
		{"PT Alternative", "Target creature gets -1/-1 until end of turn.\nMorbid — That creature gets -13/-13 until end of turn instead if a creature died this turn.", 1, 13, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			def := compileUnlessCard(t, cardgen.ScryfallCard{
				Name: test.name, Layout: "normal", TypeLine: "Instant", OracleText: test.text,
			})
			for _, died := range []bool{false, true} {
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				recipient := addCombatPermanent(g, game.Player3, vanillaCreature("Recipient", 20, 20))
				if died {
					emitCreatureDiedEvent(g)
				}
				obj := &game.StackObject{Controller: game.Player2,
					Targets: []game.Target{game.PermanentTarget(recipient.ObjectID)}}
				NewEngine(nil).resolveAbilityContentWithChoices(g, obj, def.SpellAbility.Val,
					[game.NumPlayers]PlayerAgent{}, &TurnLog{})
				want := test.base
				if died {
					want = test.bonus
				}
				if test.damage {
					if recipient.MarkedDamage != want {
						t.Fatalf("died=%v damage=%d want=%d", died, recipient.MarkedDamage, want)
					}
				} else if toughness, ok := effectiveToughness(g, recipient); !ok || toughness != 20-want {
					t.Fatalf("died=%v toughness=%d ok=%v want=%d", died, toughness, ok, 20-want)
				}
			}
		})
	}
}

func TestCompiledResolvingParagraphExpandedMana(t *testing.T) {
	def := compileUnlessCard(t, cardgen.ScryfallCard{
		Name: "Expanded Mana Alternative", Layout: "normal", TypeLine: "Instant",
		OracleText: "Add {B}{B}{B}.\nThreshold — Add {B}{B}{B}{B}{B} instead if there are seven or more cards in your graveyard.",
	})
	for _, graveyardSize := range []int{6, 7} {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		for range graveyardSize {
			addCardToGraveyard(g, game.Player2, vanillaCreature("Grave", 1, 1))
		}
		NewEngine(nil).resolveAbilityContentWithChoices(g, &game.StackObject{Controller: game.Player2},
			def.SpellAbility.Val, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
		want := 3
		if graveyardSize == 7 {
			want = 5
		}
		if got := g.Players[game.Player2].ManaPool.Amount(mana.B); got != want {
			t.Fatalf("graveyard=%d mana=%d want=%d", graveyardSize, got, want)
		}
		if g.Players[game.Player1].ManaPool.Total() != 0 {
			t.Fatal("production leaked to another controller")
		}
	}
}

func TestCompiledResolvingParagraphCounterPaymentAlternative(t *testing.T) {
	def := compileUnlessCard(t, cardgen.ScryfallCard{
		Name: "Counter Payment Alternative", Layout: "normal", TypeLine: "Instant",
		OracleText: "Counter target noncreature spell unless its controller pays {1}.\nFerocious — If you control a creature with power 4 or greater, counter that spell instead.",
	})
	for _, ferocious := range []bool{false, true} {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		g.Players[game.Player2].ManaPool.Add(mana.C, 1)
		if ferocious {
			addCombatPermanent(g, game.Player1, vanillaCreature("Qualifying", 4, 4))
		}
		target := addStackSpell(g, game.Player2, "Target", []types.Card{types.Sorcery})
		g.CardInstances[target.SourceID].Owner = game.Player3
		pushCounterTaxSpell(g, def, target)
		log := TurnLog{}
		NewEngine(nil).resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &log)
		if _, remains := stackObjectByID(g, target.ID); remains == ferocious {
			t.Fatalf("ferocious=%v target remains=%v", ferocious, remains)
		}
		if got := g.Players[game.Player2].ManaPool.Total(); got != 0 && !ferocious || got != 1 && ferocious {
			t.Fatalf("ferocious=%v payer mana=%d", ferocious, got)
		}
		if ferocious && (len(log.Choices) != 0 || !g.Players[game.Player3].Graveyard.Contains(target.SourceID)) {
			t.Fatal("override charged payment or lost physical owner")
		}
	}
}
