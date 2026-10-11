package rules

import (
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestCompiledNamedDamageUsesOriginalPermanentAttribution(t *testing.T) {
	def := compileUnlessCard(t, cardgen.ScryfallCard{Name: "Prodigal Sorcerer", Layout: "normal",
		TypeLine: "Creature — Human Wizard", Power: new("1"), Toughness: new("1"),
		OracleText: "{T}: Prodigal Sorcerer deals 1 damage to any target."})
	content := def.ActivatedAbilities[0].Content
	damage, ok := content.Modes[0].Sequence[0].Primitive.(game.Damage)
	if !ok || !damage.DamageSource.Exists || damage.DamageSource.Val != game.SourcePermanentReference() {
		t.Fatalf("damage = %+v, want explicit original permanent attribution", damage)
	}
	for _, departed := range []bool{false, true} {
		t.Run(map[bool]string{false: "live changed controller", true: "departed and reincarnated"}[departed], func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			source := addCombatPermanent(g, game.Player1, def)
			g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
				ID: g.IDGen.Next(), AffectedObjectID: source.ObjectID, Layer: game.LayerAbility,
				AddKeywords: []game.Keyword{game.Lifelink},
			})
			obj := &game.StackObject{ID: g.IDGen.Next(), Kind: game.StackActivatedAbility, Controller: game.Player1,
				SourceID: source.ObjectID, SourceCardID: source.CardInstanceID,
				Targets: []game.Target{game.PlayerTarget(game.Player2)}}
			wantGainer := game.Player3
			if departed {
				r := newEffectResolver(NewEngine(nil), g, obj, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
				r.resolveInstruction(&game.Instruction{Primitive: game.MovePermanent{
					Object: game.SourcePermanentReference(), Destination: zone.Graveyard,
				}})
				returnObj := &game.StackObject{ID: g.IDGen.Next(), Controller: game.Player3,
					Targets: []game.Target{currentCardTarget(t, g, source.CardInstanceID)}}
				newEffectResolver(NewEngine(nil), g, returnObj, [game.NumPlayers]PlayerAgent{}, &TurnLog{}).
					resolveInstruction(&game.Instruction{Primitive: game.PutOnBattlefield{
						Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget}),
					}})
				returned := permanentForCard(g, source.CardInstanceID)
				if returned == nil || returned.ObjectID == source.ObjectID || hasKeyword(g, returned, game.Lifelink) {
					t.Fatal("fixture did not create a distinct, keyword-free reincarnation")
				}
				wantGainer = game.Player1
			} else {
				source.Controller = game.Player3
			}
			NewEngine(nil).resolveAbilityContentWithChoices(g, obj, content, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
			for _, player := range []game.PlayerID{game.Player1, game.Player2, game.Player3} {
				want := 40
				if player == wantGainer {
					want++
				}
				if player == game.Player2 {
					want--
				}
				if g.Players[player].Life != want {
					t.Fatalf("player %d life=%d want %d: damage did not use original/current-controller or LKI attribution",
						player, g.Players[player].Life, want)
				}
			}
		})
	}
	spell := compileUnlessCard(t, cardgen.ScryfallCard{Name: "Shock", Layout: "normal", TypeLine: "Instant",
		OracleText: "Shock deals 2 damage to any target."})
	spellDamage, ok := spell.SpellAbility.Val.Modes[0].Sequence[0].Primitive.(game.Damage)
	if !ok || spellDamage.DamageSource.Exists {
		t.Fatal("resolving spell acquired a permanent-source attribution")
	}
}
