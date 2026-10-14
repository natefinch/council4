package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards/authored"
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/opt"
)

// CR 707.8a: a token copy preserves both copiable faces and enters the same face up.
func TestAuthoredCreatureCopyDoubleFacedTokens(t *testing.T) {
	t.Parallel()
	for _, entry := range []cardset.Entry{authored.CacklingCounterpart, authored.FatedInfatuation} {
		t.Run(entry.Name, func(t *testing.T) {
			t.Parallel()
			for _, face := range []game.FaceIndex{game.FaceFront, game.FaceBack} {
				t.Run(map[game.FaceIndex]string{game.FaceFront: "front", game.FaceBack: "back"}[face], func(t *testing.T) {
					t.Parallel()
					g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
					g.Turn.ActivePlayer = game.Player2
					engine := choicesProgramEngine(t)
					front := vanillaCreature("Front creature", 2, 3)
					front.ManaCost = opt.Val(cost.Mana{cost.O(3)})
					front.Layout = game.LayoutTransform
					back := vanillaCreature("Back creature", 5, 6)
					back.ReplacementAbilities = []game.ReplacementAbility{
						game.EntersWithCountersReplacement("Enters with a counter.", game.CounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 1}),
					}
					front.Back = opt.Val(back.CardFace)
					source := addCombatPermanent(g, game.Player1, front)
					source.Face, source.Transformed = face, face == game.FaceBack
					addImplementationSpellToStack(g, game.Player1, entry.New(), []game.Target{game.PermanentTarget(source.ObjectID)})
					engine.resolveTopOfStack(g, &TurnLog{})
					token := g.Battlefield[len(g.Battlefield)-1]
					if !token.Token || !token.TokenDef.Back.Exists || token.Face != face ||
						!token.TokenDef.IsTransformingDoubleFaced() || token.TokenDef.Name != front.Name ||
						token.TokenDef.Back.Val.Name != back.Name {
						t.Fatalf("copy lost either face or entry orientation: %+v", token)
					}
					if mv, known := effectivePermanentManaValue(g, token); !known || mv != 3 {
						t.Fatal("full double-faced token lost its front mana value")
					}
					wantCounter := 0
					if face == game.FaceBack {
						wantCounter = 1
					}
					if token.Counters.Get(counter.PlusOnePlusOne) != wantCounter {
						t.Fatal("entry replacement used the wrong face")
					}
					branch := g.Clone()
					copied, _ := permanentByObjectID(branch, token.ObjectID)
					if !transformPermanent(branch, copied) || copied.Face == face {
						t.Fatal("token could not transform")
					}
					if mv, known := effectivePermanentManaValue(branch, copied); !known || mv != 3 {
						t.Fatal("transforming a full double-faced token changed its mana value")
					}
					wantPower := 5
					if face == game.FaceBack {
						wantPower = 2
					}
					if effectivePower(branch, copied) != wantPower+wantCounter || token.Face != face || source.Face != face {
						t.Fatal("transformation changed the original or lost the other face's values")
					}
					addImplementationSpellToStack(g, game.Player1, entry.New(), []game.Target{game.PermanentTarget(token.ObjectID)})
					engine.resolveTopOfStack(g, &TurnLog{})
					second := g.Battlefield[len(g.Battlefield)-1]
					if !second.TokenDef.Back.Exists || second.Face != face || !transformPermanent(g, second) {
						t.Fatal("copy of a double-faced token lost its faces")
					}
				})
			}
		})
	}
}
