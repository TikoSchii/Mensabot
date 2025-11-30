package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/PuerkitoBio/goquery"
)

// ----------------------------------------
// Datentyp für ein Gericht
// ----------------------------------------
type Meal struct {
	Category  string // z.B. "Angebot 1"
	Name      string // Name des Gerichts
	PriceStud string // Preis für Studierende
}

// ----------------------------------------
// Funktion: Webseite abrufen + Menü extrahieren
// ----------------------------------------
func fetchMenu() ([]Meal, error) {

	// Seite abrufen
	resp, err := http.Get("https://stwwb.webspeiseplan.de/Menu")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP Status %d", resp.StatusCode)
	}

	// HTML-Dokument mit goquery einlesen
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	meals := []Meal{}

	// Jeden Menü-Block finden
	doc.Find(".meal-wrapper .meal").Each(func(i int, s *goquery.Selection) {

		m := Meal{}

		// Kategorie (Angebot 1, Angebot 2, …)
		m.Category = s.Find(".categoryName").Text()

		// Name des Gerichts
		m.Name = s.Find(".mealNameWrapper").Text()

		// Preise extrahieren
		s.Find(".price-row").Each(func(j int, pr *goquery.Selection) {

			label := pr.Find(".price-label").Text() // z.B. "Studierende"
			value := pr.Find(".price-value").Text() // z.B. "2,15 €"

			if label == "Studierende" {
				m.PriceStud = value
			}

		})

		// Nur speichern, wenn der Name nicht leer ist
		if m.Name != "" {
			meals = append(meals, m)
		}
	})

	return meals, nil
}

// ----------------------------------------
// Menü schön formatieren für Telegram
// ----------------------------------------
func formatMeals(meals []Meal) string {

	// Wenn keine Gerichte vorhanden → Hinweistext zurückgeben
	if len(meals) == 0 {
		return "Heute gibt es keinen Speiseplan."
	}

	msg := "🍽 *Heutiger Mensaplan*\n\n"

	for _, m := range meals {
		msg += fmt.Sprintf(
			"*%s*\n%s\n👤 Studierende: %s\n\n",
			m.Category,
			m.Name,
			m.PriceStud,
		)
	}

	return msg
}

// ----------------------------------------
// Nachricht per Telegram senden
// ----------------------------------------
func sendTelegram(token, chatID, text string) error {

	// Telegram API URL zum Nachrichten versenden
	api := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)

	// HTTP POST Anfrage an Telegram senden
	_, err := http.PostForm(api, url.Values{
		"chat_id":    {chatID},
		"text":       {text},
		"parse_mode": {"Markdown"}, // erlaubt fette Schrift usw.
	})

	return err
}

// ----------------------------------------
// Hauptprogramm
// ----------------------------------------
func main() {

	botToken := "8590472718:AAG0mFAIjn8j2nF_X1Y5T6nqMMhPJYPRY3w"
	chatID := "8479860473"

	meals, err := fetchMenu()
	if err != nil {
		log.Fatal("Fehler beim Abrufen des Menüs:", err)
	}

	// Formatieren (liefert automatisch „kein Menü“-Nachricht falls leer)
	msg := formatMeals(meals)

	// Nachricht senden
	err = sendTelegram(botToken, chatID, msg)
	if err != nil {
		log.Fatal("Fehler beim Senden an Telegram:", err)
	}

	// Log-Ausgabe im Terminal
	if len(meals) == 0 {
		fmt.Println("Heute: Kein Menü gefunden → Hinweis an Telegram gesendet.")
	} else {
		fmt.Println("Menü erfolgreich an Telegram gesendet!")
	}
}
