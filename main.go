package main

import (
	"fmt"
	"log"
)

func main() {
	fmt.Println("=== Skaletek Rule V2 Engine POC ===\n")

	events := seedEvents()
	rules := seedRules()

	compiled, err := compileRules(rules)
	if err != nil {
		log.Fatalf("rule compilation error: %v", err)
	}
	fmt.Printf("✓ Compiled %d rules\n\n", len(compiled))

	var allAlerts []Alert

	for _, event := range events {
		fmt.Printf("Processing event %s (template: %s)\n", event.ID, event.TemplateID)

		alerts, err := evaluateRules(event, compiled)
		if err != nil {
			log.Printf("evaluation error on event %s: %v", event.ID, err)
			continue
		}

		if len(alerts) == 0 {
			fmt.Println("  → no rules matched")
		}

		for _, alert := range alerts {
			fmt.Printf("  → ALERT [%s] %s: %s\n", alert.Severity, alert.RuleName, alert.Message)
			allAlerts = append(allAlerts, alert)
		}

		fmt.Println()
	}

	if err := persistAlerts(allAlerts, "results.json"); err != nil {
		log.Fatalf("failed to write results: %v", err)
	}

	fmt.Printf("=== Done: %d alerts generated, written to results.json ===\n", len(allAlerts))
}
