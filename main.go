package main

import (
	"database/sql"
	"encoding/binary"
	"encoding/json"
	"fmt"
	_ "github.com/lib/pq"
	"gopkg.in/matryer/try.v1"
	"log"
	"mac2ip/config"
	"net"
	"net/http"
	"strconv"
	"time"
)

const (
	table    = "devices"
	counter  = "counter"
	maxValue = 2147483647
)

func fetchIP(db *sql.DB, hw net.HardwareAddr) *string {
	row := db.QueryRow("SELECT ip FROM "+table+" WHERE mac=$1", hw.String())
	var ip string
	err := row.Scan(&ip)
	switch {
	case err == sql.ErrNoRows:
		return nil
	case err != nil:
		log.Fatalf("query error: %v\n", err)
	}

	return &ip
}

func IntToIP(nn uint32) (string, error) {
	if nn > maxValue {
		return "", fmt.Errorf("IP value out of range: %v", nn)
	}
	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, nn)
	return ip.String(), nil
}

func registerDevice(db *sql.DB, hw net.HardwareAddr) *string {
	var c uint32
	err := db.QueryRow("SELECT nextval($1)", counter).Scan(&c)
	if err != nil {
		log.Fatalf("failed to fetch counter: %v\n", err)
	}

	ip, err := IntToIP(c)
	if err != nil {
		log.Fatal(err)
	}

	err = db.QueryRow("INSERT INTO "+table+" (mac, ip) VALUES ($1, $2) RETURNING ip", hw.String(), ip).Scan(&ip)
	if err != nil {
		log.Fatalf("failed to register device: %v\n", err)
	}

	return &ip
}

func tryQuery(db *sql.DB, query string) error {
	return try.Do(func(attempt int) (bool, error) {
		_, err := db.Query(query)
		if err != nil {
			if attempt == 5 {
				return false, err
			}
			timeout := time.Duration(5*attempt) * time.Second
			log.Printf("failed to execute query %v, retrying in %v...", query, timeout)
			time.Sleep(timeout)
		}
		return true, err
	})
}

func getIP(w http.ResponseWriter, r *http.Request, db *sql.DB) *string {
	type request struct {
		MAC string `json:"mac"`
	}

	decoder := json.NewDecoder(r.Body)
	var s request
	err := decoder.Decode(&s)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("expected JSON request of format { \"mac\": \"11:22:33:44:55:66\" }"))
		return nil
	}

	hw, err := net.ParseMAC(s.MAC)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("failed to parse MAC"))
		return nil
	}

	ip := fetchIP(db, hw)
	if ip == nil {
		ip = registerDevice(db, hw)
	}

	if ip == nil {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "failed to fetch IP for mac %s", s.MAC)
	}

	return ip
}

func main() {
	conf := config.GetConf()
	connStr := fmt.Sprintf("dbname=%s user=%s password=%s host=%s sslmode=disable", conf.Name, conf.User, conf.Pass, conf.Host)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	s := strconv.Itoa(maxValue)
	if err := tryQuery(db, "CREATE SEQUENCE IF NOT EXISTS "+counter+" MAXVALUE "+s); err != nil {
		log.Fatal("failed to create IP sequence")
	}

	if err := tryQuery(db, "CREATE TABLE IF NOT EXISTS "+table+" (mac macaddr PRIMARY KEY, ip inet UNIQUE)"); err != nil {
		log.Fatal("failed to create table")
	}

	http.Handle("/ip", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getIP(w, r, db)
		if ip == nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		type response struct {
			IP string `json:"ip"`
		}

		p := response{*ip}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(p)
	}))

	http.Handle("/ipxe", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getIP(w, r, db)
		if ip == nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "#!ipxe\nchain http://%s/%s.ipxe\n", conf.Instance, *ip)
	}))

	log.Fatal(http.ListenAndServe(":8080", nil))
}
