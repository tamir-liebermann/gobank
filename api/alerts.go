package api

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	tzevaadomWSURL = "wss://ws.tzevaadom.co.il/socket?platform=ANDROID"
	summaryWindow  = 12 * time.Hour
	reconnectDelay = 5 * time.Second
)

// tzevaadomMsg is the raw message format from the Tzevaadom WebSocket.
type tzevaadomMsg struct {
	Type   int      `json:"type"`
	Time   string   `json:"time"`
	Threat int      `json:"threat"`
	Cities []string `json:"cities"`
	IsDrill bool    `json:"isDrill"`
}

type timedAlert struct {
	alert  OrefAlert
	seenAt time.Time
}

type AlertPoller struct {
	mu         sync.Mutex
	seen       map[string]time.Time
	lastAlerts []OrefAlert
	history    []timedAlert
}

func NewAlertPoller() *AlertPoller {
	return &AlertPoller{seen: make(map[string]time.Time)}
}

func (p *AlertPoller) Start(store *SubscriberStore, broadcast func(*Subscriber, OrefAlert)) {
	for {
		if err := p.connect(store, broadcast); err != nil {
			log.Printf("WebSocket disconnected: %v — reconnecting in %s", err, reconnectDelay)
		}
		time.Sleep(reconnectDelay)
	}
}

func (p *AlertPoller) connect(store *SubscriberStore, broadcast func(*Subscriber, OrefAlert)) error {
	headers := map[string][]string{
		"Origin":     {"https://www.tzevaadom.co.il"},
		"User-Agent": {"Mozilla/5.0 (Linux; Android 10) AppleWebKit/537.36 Chrome/120.0 Mobile Safari/537.36"},
	}
	conn, _, err := websocket.DefaultDialer.Dial(tzevaadomWSURL, headers)
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}
	defer conn.Close()
	log.Println("Connected to Tzevaadom WebSocket")

	// keepalive ping
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			return fmt.Errorf("read: %w", err)
		}

		var raw tzevaadomMsg
		if err := json.Unmarshal(msg, &raw); err != nil {
			continue // skip non-alert messages (e.g. pongs, system messages)
		}

		// type 0 = alert, isDrill = skip drills
		if raw.Type != 0 || raw.IsDrill || len(raw.Cities) == 0 {
			continue
		}

		alert := p.toOrefAlert(raw)
		dedupKey := alert.ID
		if p.isDuplicate(dedupKey) {
			continue
		}
		p.markSeen(dedupKey)
		p.storeRecent(alert)

		log.Printf("New alert: %s — %s", alert.Title, strings.Join(alert.Data, ", "))
		for _, sub := range store.All() {
			go broadcast(sub, alert)
		}
	}
}

func (p *AlertPoller) toOrefAlert(raw tzevaadomMsg) OrefAlert {
	return OrefAlert{
		ID:    fmt.Sprintf("%s-%d", raw.Time, raw.Threat),
		Cat:   fmt.Sprintf("%d", raw.Threat),
		Title: threatTitle(raw.Threat),
		Data:  raw.Cities,
		Desc:  threatDesc(raw.Threat),
	}
}

func threatTitle(threat int) string {
	switch threat {
	case 1:
		return "ירי רקטות וטילים"
	case 2:
		return "חדירת כלי טיס עוין"
	case 3:
		return "רעידת אדמה"
	case 4:
		return "חשש לחדירת מחבלים"
	case 5:
		return "חומרים מסוכנים"
	case 6:
		return "התרעה בטחונית"
	case 13:
		return "ירי רקטות וטילים" // pre-alert
	default:
		return "התרעה"
	}
}

func threatDesc(threat int) string {
	switch threat {
	case 1, 13:
		return "היכנסו למרחב המוגן ושהו בו 10 דקות"
	case 2:
		return "היכנסו למרחב המוגן ושהו בו 10 דקות"
	case 4:
		return "היכנסו למבנה, נעלו את הדלת ועצרו את הכניסה"
	default:
		return "פעלו לפי הנחיות פיקוד העורף"
	}
}

func (p *AlertPoller) isDuplicate(id string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	_, exists := p.seen[id]
	return exists
}

func (p *AlertPoller) markSeen(id string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.seen[id] = time.Now()
	// cleanup old entries
	if len(p.seen) > 500 {
		cutoff := time.Now().Add(-1 * time.Hour)
		for k, t := range p.seen {
			if t.Before(cutoff) {
				delete(p.seen, k)
			}
		}
	}
}

func (p *AlertPoller) storeRecent(alert OrefAlert) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.lastAlerts = append([]OrefAlert{alert}, p.lastAlerts...)
	if len(p.lastAlerts) > 20 {
		p.lastAlerts = p.lastAlerts[:20]
	}
	p.history = append(p.history, timedAlert{alert: alert, seenAt: time.Now()})
}

// History returns all alerts seen in the last 12 hours.
func (p *AlertPoller) History() []OrefAlert {
	p.mu.Lock()
	defer p.mu.Unlock()
	cutoff := time.Now().Add(-summaryWindow)
	var result []OrefAlert
	for _, ta := range p.history {
		if ta.seenAt.After(cutoff) {
			result = append(result, ta.alert)
		}
	}
	return result
}

func (p *AlertPoller) RecentAlerts() []OrefAlert {
	p.mu.Lock()
	defer p.mu.Unlock()
	result := make([]OrefAlert, len(p.lastAlerts))
	copy(result, p.lastAlerts)
	return result
}

// broadcastAlert sends an alert to a single subscriber with location-based prioritization.
func (api *ApiManager) broadcastAlert(sub *Subscriber, alert OrefAlert) {
	var msg string
	if sub.Language == "en" {
		translated, err := translateAreas(alert.Data)
		if err == nil {
			msg = buildMessage(sub, alert, translated, true)
		} else {
			msg = buildMessage(sub, alert, alert.Data, false)
		}
	} else {
		msg = buildMessage(sub, alert, alert.Data, false)
	}
	to := sub.Phone
	if !strings.HasPrefix(to, "whatsapp:") {
		to = "whatsapp:" + to
	}
	if err := api.sendWhatsAppMessage(to, msg); err != nil {
		log.Printf("broadcast failed for %s: %v", sub.Phone, err)
	}
}

func buildMessage(sub *Subscriber, alert OrefAlert, areas []string, inEnglish bool) string {
	areaStr := strings.Join(areas, ", ")
	prefix := "📢"
	urgent := false

	if sub.Location != "" {
		for _, area := range alert.Data {
			if strings.Contains(area, sub.Location) || strings.Contains(sub.Location, area) {
				urgent = true
				break
			}
		}
		if !urgent && inEnglish {
			for _, area := range areas {
				if strings.EqualFold(area, sub.Location) {
					urgent = true
					break
				}
			}
		}
	}

	if urgent {
		prefix = "🚨 *YOUR AREA*"
	}

	if inEnglish {
		return fmt.Sprintf("%s\n*%s*\nAreas: %s\n%s", prefix, threatTitleEn(alert.Cat), areaStr, alert.Desc)
	}
	return fmt.Sprintf("%s *%s*\nאזורים: %s\n%s", prefix, alert.Title, areaStr, alert.Desc)
}

func threatTitleEn(cat string) string {
	switch cat {
	case "1", "13":
		return "Missile/Rocket Alert"
	case "2":
		return "Hostile Aircraft Infiltration"
	case "4":
		return "Terrorist Infiltration"
	case "5":
		return "Hazardous Materials"
	default:
		return "Security Alert"
	}
}

func translateAreas(areas []string) ([]string, error) {
	prompt := fmt.Sprintf(
		"Translate these Israeli city/area names from Hebrew to English. Return a comma-separated list only, no extra text: %s",
		strings.Join(areas, ", "),
	)
	result, err := claudeComplete("You are a translator.", prompt)
	if err != nil {
		return nil, err
	}
	return strings.Split(result, ", "), nil
}
