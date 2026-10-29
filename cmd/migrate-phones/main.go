package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/config"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/database"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/validation"
)

func main() {
	apply := flag.Bool("apply", false, "write changes to the database")
	flag.Parse()

	cfg := config.Load()
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer database.Close()

	// --- 1. Empty-string phones: dry-run count, then NULL them if applying ---
	var emptyCount int64
	if err := db.Raw(`SELECT COUNT(*) FROM users WHERE phone = ''`).Scan(&emptyCount).Error; err != nil {
		log.Fatalf("count empty phones: %v", err)
	}
	fmt.Printf("empty-string phones: %d\n", emptyCount)

	if *apply && emptyCount > 0 {
		if err := db.Exec(`UPDATE users SET phone = NULL, updated_at = NOW() WHERE phone = ''`).Error; err != nil {
			log.Fatalf("normalize empty phones: %v", err)
		}
		log.Printf("normalized %d empty-string phones to NULL", emptyCount)
	}
	fmt.Println()

	// --- 2. Normalize remaining non-null, non-empty phones to E.164 ---
	v := validation.New()

	type row struct {
		ID    string
		Phone *string
	}

	var rows []row
	if err := db.Raw(`SELECT id, phone FROM users WHERE phone IS NOT NULL AND phone <> '' ORDER BY created_at ASC`).
		Scan(&rows).Error; err != nil {
		log.Fatalf("select users: %v", err)
	}

	var changed, invalid, unchanged int
	seen := make(map[string]string)

	for _, r := range rows {
		if r.Phone == nil || *r.Phone == "" {
			continue
		}

		normalized, err := v.Phone.NormalizeAndValidate(*r.Phone)
		if err != nil {
			invalid++
			log.Printf("SKIP id=%s raw=%q err=%v", r.ID, *r.Phone, err)
			continue
		}

		if prevID, dup := seen[normalized]; dup {
			invalid++
			log.Printf("DUPLICATE id=%s raw=%q normalized=%q collides with user %s",
				r.ID, *r.Phone, normalized, prevID)
			continue
		}
		seen[normalized] = r.ID

		if normalized == *r.Phone {
			unchanged++
			continue
		}

		changed++
		fmt.Printf("id=%s  %q -> %q\n", r.ID, *r.Phone, normalized)

		if *apply {
			if err := db.Exec(`UPDATE users SET phone = ?, updated_at = NOW() WHERE id = ?`,
				normalized, r.ID).Error; err != nil {
				log.Printf("UPDATE failed for id=%s: %v", r.ID, err)
			}
		}
	}

	fmt.Printf("\nsummary: total=%d unchanged=%d changed=%d invalid=%d apply=%v\n",
		len(rows), unchanged, changed, invalid, *apply)

	if invalid > 0 && *apply {
		log.Fatal("aborting: unparseable or duplicate rows remain; fix them before applying the migration")
	}
}