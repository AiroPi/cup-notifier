package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	apprise "github.com/unraid/apprise-go"
)

var cache = make(map[string]map[string]any)

func checkForUpdates(notifier *apprise.Apprise) (err error) {
	if os.Getenv("INSECURE_SKIP_VERIFY") == "true" {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	baseURL, _ := strings.CutSuffix(os.Getenv("CUP_URL"), "/")
	resp, err := http.Get(baseURL + "/api/v3/json")
	if err != nil {
		return fmt.Errorf("unable to get cup data: %w", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected cup status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("unable to read cup response: %w", err)
	}
	var dat map[string]any
	err = json.Unmarshal(body, &dat)
	if err != nil {
		return fmt.Errorf("unable to decode cup response: %w", err)
	}

	newCache := make(map[string]map[string]any)
	alerts := make(map[string][]string)

	// we map all images (from cup API)
	for _, image := range dat["images"].([]any) {
		img := image.(map[string]any)
		parts := img["parts"].(map[string]any)
		result := img["result"].(map[string]any)

		// hardcoded special case for multi-host cup : the "server" key is null for the server which got the request.
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
			infoType := info["type"].(string)

			// image can be present multiple time (multi-host) so we initialize the cache only once.
			if _, exists := newCache[key]; !exists {
				newCache[key] = make(map[string]any)
				newCache[key]["url"] = img["url"].(string)
				newCache[key]["hosts"] = make(map[string]map[string]string)
			}

			hosts := newCache[key]["hosts"].(map[string]map[string]string)
			var oldHosts map[string]map[string]string
			if oldImage, ok := cache[key]; ok {
				oldHosts = oldImage["hosts"].(map[string]map[string]string)
			}

			switch infoType {
			case "digest":
				hosts[server] = map[string]string{"type": "digest"}
				if oldHosts == nil || oldHosts[server] == nil {
					alerts[key] = append(alerts[key], server)
				}
			case "version":
				newVersion := info["new_version"].(string)

				hosts[server] = map[string]string{
					"type":    "version",
					"current": info["current_version"].(string),
					"new":     newVersion,
				}

				if oldHosts == nil || oldHosts[server] == nil || oldHosts[server]["new"] != newVersion {
					alerts[key] = append(alerts[key], server)
				}
			}
		}
	}

	cache = newCache
	for k, v := range alerts {
		var title string
		if cache[k]["url"] == nil {
			title = fmt.Sprintf("**New updates for `%v`:**", k)
		} else {
			title = fmt.Sprintf("**New updates for [`%v`](%v):**", k, cache[k]["url"])
		}
		var content []string

		for _, h := range v {
			infos := cache[k]["hosts"].(map[string]map[string]string)[h]
			if infos["current"] != "" {
				content = append(content, fmt.Sprintf(" - *%v* (`%v` → `%v`)", h, infos["current"], infos["new"]))
			} else {
				content = append(content, fmt.Sprintf(" - *%v* (digest change)", h))
			}
		}

		fmt.Println(title)
		fmt.Println(strings.Join(content, "\n"))
		if err := notifier.Send(
			strings.Join(content, "\n"),
			apprise.WithTitle(title),
			apprise.WithNotifyType(apprise.NotifyInfo),
		); err != nil {
			log.Println(err)
		}
	}

	return err
}

func main() {
	notifier := apprise.New()
	endpoints := os.Getenv("NOTIFICATION_URLS")
	if endpoints != "" {
		for url := range strings.SplitSeq(endpoints, ",") {
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

	for {
		fmt.Println("Checking updates...")
		if err := checkForUpdates(notifier); err != nil {
			log.Println(err)
		}
		time.Sleep(5 * time.Minute)
	}
}
