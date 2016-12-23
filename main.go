package main

import (
	"log"

	"database/sql"
	"os"

	"strings"

	_ "github.com/lib/pq"
)

var DATABASE_URL = os.Getenv("DATABASE_URL")

func main() {
	d := NewDispatcher()
	d.HandleFunc("container", "start", Start)
	d.HandleFunc("container", "die", Die)
	d.HandleFunc("image", "pull", UpdateImage)
	d.HandleFunc("image", "tag", UpdateImage)
	log.Fatal(d.Run())
}

func Start(e Event) {
	name := e["Actor"].(map[string]interface{})["Attributes"].(map[string]interface{})["name"].(string)
	log.Printf("Handling start event for container: %s\n", name)
	db, err := sql.Open("postgres", DATABASE_URL)
	if err != nil {
		log.Printf("DB connection error: %s\n", err)
		return
	}
	defer db.Close()
	id, err := decodeHashID(strings.Split(name, "_")[1])
	if err != nil {
		log.Printf("Error decoding hash id: %s\n", err)
		return
	}
	_, err = db.Exec(`INSERT INTO server_run_statistics (server_id, start) VALUES ($1, CURRENT_TIMESTAMP)`, id)
	if err != nil {
		log.Printf("Error inserting server statistics: %s\n", err)
	}
}

func Die(e Event) {
	attrs := e["Actor"].(map[string]interface{})["Attributes"].(map[string]interface{})
	name := attrs["name"].(string)
	log.Printf("Handling die event for container: %s\n", name)
	db, err := sql.Open("postgres", DATABASE_URL)
	if err != nil {
		log.Printf("DB connection error: %s\n", err)
		return
	}
	defer db.Close()
	exitCode := attrs["exitCode"]
	id, err := decodeHashID(strings.Split(name, "_")[1])
	if err != nil {
		log.Printf("Error decoding hash id: %s\n", err)
		return
	}
	_, err = db.Exec(`UPDATE server_run_statistics SET stop = CURRENT_TIMESTAMP, exit_code = $1
	 WHERE server_run_statistics.server_id = $2 AND server_run_statistics.stop IS NULL`, exitCode, id)
	if err != nil {
		log.Printf("Error inserting server statistics: %s\n", err)
	}
}

func UpdateImage(e Event) {
	image := e["Actor"].(map[string]interface{})["Attributes"].(map[string]interface{})["name"].(string)
	imageName := strings.Split(image, ":")[0]
	log.Printf("Handling image update: %s\n", imageName)
	db, err := sql.Open("postgres", DATABASE_URL)
	if err != nil {
		log.Printf("DB connection error : %s\n", err)
		return
	}
	defer db.Close()
	rows, err := db.Query(`SELECT servers.id
		FROM servers
		  JOIN environment_type ON servers.environment_type_id = environment_type.id
		WHERE environment_type.image_name LIKE $1 || '%';`, imageName)
	if err != nil {
		log.Printf("Query error: %s\n", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var id ServerId
		if err = rows.Scan(&id); err != nil {
			log.Printf("Row scan error: %s\n", err)
			continue
		}
		if !checkIfSent(id.Hash()) {
			go sendNotification(id)
			go setCache(id)
		}
	}
	if err = rows.Err(); err != nil {
		log.Printf("Rows error: %s\n", err)
	}
}
