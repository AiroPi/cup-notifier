package main

import (
	"fmt"
	"net/http"
	"io"
	"encoding/json"
	"strings"
	"log"
	"crypto/tls"
	"time"
	"os"
	apprise "github.com/unraid/apprise-go"
)

var cache = make(map[string]map[string]any)

func checkForUpdates(notifier *apprise.Apprise) {
	if os.Getenv("INSECURE_SKIP_VERIFY") == "true" {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	baseUrl, _ := strings.CutSuffix(os.Getenv("CUP_URL"), "/")
	resp, err := http.Get(baseUrl + "/api/v3/json")
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	var dat map[string]any
	err = json.Unmarshal(body, &dat)
	if err != nil {
		log.Fatalln(err)
	}

	newCache := make(map[string]map[string]any)
	alerts := make(map[string][]string)
	for _, image := range dat["images"].([]any) {
		img := image.(map[string]any)
		parts := img["parts"].(map[string]any)
		result := img["result"].(map[string]any)
		server := "bluewhale"
		if s, ok := img["server"].(string); ok {
			server = s
		}

		if result["error"] != nil {
			continue
		}
		if result["has_update"].(bool) {
			info := result["info"].(map[string]any)

			key := parts["registry"].(string) + "/" + parts["repository"].(string)
			info_type := info["type"].(string)

			if _, exists := newCache[key]; !exists {
				newCache[key] = make(map[string]any)
				newCache[key]["hosts"] = make(map[string]map[string]string)
			}

			hosts := newCache[key]["hosts"].(map[string]map[string]string)
			var oldHosts map[string]map[string]string
			if oldImage, ok := cache[key]; ok {
				oldHosts = oldImage["hosts"].(map[string]map[string]string)
			}

			if info_type == "digest" {
				hosts[server] = map[string]string{"type": "digest"}
				if oldHosts == nil || oldHosts[server] == nil {
					alerts[key] = append(alerts[key], server)
				}
			} else if info_type == "version" {
				newVersion := info["new_version"].(string)

				hosts[server] = map[string]string{
					"type": "version", 
					"current": info["current_version"].(string), 
					"new": newVersion,
				}

				if oldHosts == nil || oldHosts[server] == nil || oldHosts[server]["new"] != newVersion {
					alerts[key] = append(alerts[key], server)
				}
			}
		}
	}

	cache = newCache
	for k, v := range alerts {
		title := fmt.Sprintf("New updates for %v", k)
		var content []string

		for _, h := range v {
			infos := cache[k]["hosts"].(map[string]map[string]string)[h]
			content = append(content, fmt.Sprintf(" - %v (%v -> %v)\n", h, infos["current"], infos["new"]))
		}

		fmt.Println(title)
		fmt.Println(strings.Join(content, "\n"))
		if err := notifier.Send(
			strings.Join(content, "\n"),
			apprise.WithTitle(title),
			apprise.WithNotifyType(apprise.NotifyInfo),
		); err != nil {
			log.Fatal(err)
		}
	}
}

func main() {
	for {
		notifier := apprise.New()
		endpoints := os.Getenv("NOTIFICATION_URLS")
		if endpoints != "" {
			for _, url := range(strings.Split(endpoints, ",")) {
				if err := notifier.Add(url); err != nil {
					log.Fatal(err)
				}
			}
		} else {
			log.Fatal("NOTIFICATION_URLS not defined")
		}

		if os.Getenv("CUP_URL") == "" {
			log.Fatal("CUP_URL not defined")
		}

		fmt.Println("Checking updates...")
		checkForUpdates(notifier)
		time.Sleep(5 * time.Minute)
	}
}
