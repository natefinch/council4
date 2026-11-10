package main

import (
	"testing"

	"github.com/natefinch/council4/internal/carddata"
)

func TestCardFamiliesGroupBySingleNewParagraph(t *testing.T) {
	card := func(id, name, typeLine, text string) carddata.ScryfallCard {
		return carddata.ScryfallCard{OracleID: id, Name: name, TypeLine: typeLine, OracleText: text}
	}
	corpus := []carddata.ScryfallCard{
		card("s1", "Serra Angel", "Creature — Angel", "Flying\nVigilance"),
		card("s2", "Llanowar Elves", "Creature — Elf Druid", "{T}: Add {G}."),
		card("a", "Chomping Kavu", "Creature — Kavu", "Backup 1 (Put a +1/+1 counter on target creature.)\nFlying"),
		card("b", "Saiba Cryptomancer", "Creature — Human Wizard", "Backup 2\nVigilance"),
		card("c", "Fyndhorn Elves", "Creature — Elf Druid", "{T}: Add {G}."),
		card("d", "Lonely Novelty", "Creature — Ooze", "Fuse\nFlying"),
		card("e", "Two New Things", "Creature — Ooze", "Backup 1\nFuse"),
		card("f", "Claimed Backup", "Creature — Ooze", "Backup 3"),
	}
	queue := manifest{Records: []record{
		{OracleID: "s1", Status: "installed-compiled"},
		{OracleID: "s2", Status: "implemented-verified"},
		{OracleID: "a", Name: "Chomping Kavu", Revision: "ra", Status: "unclaimed"},
		{OracleID: "b", Name: "Saiba Cryptomancer", Revision: "rb", Status: "unclaimed", Decks: 2},
		{OracleID: "c", Name: "Fyndhorn Elves", Revision: "rc", Status: "unclaimed"},
		{OracleID: "d", Name: "Lonely Novelty", Revision: "rd", Status: "unclaimed"},
		{OracleID: "e", Name: "Two New Things", Revision: "re", Status: "unclaimed"},
		{OracleID: "f", Name: "Claimed Backup", Revision: "rf", Status: "claimed", Batch: "x"},
	}}
	got := cardFamilies(queue, corpus)
	if len(got) != 2 {
		t.Fatalf("families = %+v, want backup and no-new-paragraph families only", got)
	}
	backup := got[0]
	if backup.Key != "backup n" || backup.Decks != 2 || len(backup.Cards) != 2 ||
		backup.Cards[0].Name != "Saiba Cryptomancer" || backup.Cards[1].Revision != "ra" {
		t.Errorf("backup family = %+v", backup)
	}
	if got[1].Key != "" || len(got[1].Cards) != 1 || got[1].Cards[0].Name != "Fyndhorn Elves" {
		t.Errorf("no-new-paragraph family = %+v", got[1])
	}
}
