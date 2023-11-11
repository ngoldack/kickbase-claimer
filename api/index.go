package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

type league struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type authResponse struct {
	Token   string   `json:"token"`
	Leagues []league `json:"leagues"`
}

func Handler(w http.ResponseWriter, r *http.Request) {
	body, err := json.Marshal(map[string]any{
		"email":    os.Getenv("KICKBASE_EMAIL"),
		"password": os.Getenv("KICKBASE_PASSWORD"),
		"ext":      false,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp, err := http.Post("https://api.kickbase.com/user/login", "application/json", bytes.NewReader(body))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if resp.StatusCode != 200 {
		http.Error(w, "Login failed", http.StatusInternalServerError)
		return
	}

	var authResponse authResponse
	err = json.NewDecoder(resp.Body).Decode(&authResponse)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	errs := make([]error, 0)
	for _, league := range authResponse.Leagues {
		req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("https://api.kickbase.com/leagues/%s/collectgift", league.ID), nil)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		req.Header.Add("Cookie", "kkstrauth="+authResponse.Token)

		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		if resp.StatusCode != 200 {
			errs = append(errs, err)
			continue
		}

		log.Printf("Collected gift for league %s (%s)", league.Name, league.ID)
	}

	if len(errs) > 0 {
		http.Error(w, "Errors occurred", http.StatusInternalServerError)
		for _, err := range errs {
			log.Printf("Error: %s", err.Error())
		}
		return
	}

	log.Printf("Successfully collected gifts for %d leagues", len(authResponse.Leagues))
	w.WriteHeader(http.StatusOK)
}
