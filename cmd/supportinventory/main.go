// Command supportinventory joins production support with frozen compiler evidence.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/natefinch/council4/internal/supportinventory"
	"github.com/natefinch/council4/mtg/cards"
)

func main() {
	corpus := flag.String("corpus", ".cardwork/corpus/oracle-cards.json", "pinned corpus")
	compiler := flag.String("compiler-report", ".cardwork/preservation/compile-report.json", "existing same-input compiler evidence; no compilation")
	report := flag.String("report", ".cardwork/support-inventory.json", "merged identity/status evidence")
	docs := flag.Bool("docs", false, "update canonical production support docs")
	flag.Parse()
	inventory, err := supportinventory.Load(*corpus, *compiler, cards.NewDefaultRegistry())
	if err != nil {
		panic(err)
	}
	data, err := json.MarshalIndent(inventory, "", "  ")
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile(*report, data, 0o600); err != nil {
		panic(err)
	}
	if *docs {
		if err := inventory.WriteDocs("."); err != nil {
			panic(err)
		}
	}
	fmt.Printf("Production %d entries; compiler %d/%d; evidence %s\n", len(inventory.Production), inventory.Accepted, inventory.Eligible, *report)
}
