package colab

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/playwright-community/playwright-go"
)

const keepAliveInterval = 2 * time.Minute

var (
	colabURL      string
	email         string
	password      string
	webhookURL    string
	sessionDir    string
	userID        string
	urlSent       bool // Flaga wskazująca, że URL został wysłany
	keepAliveOnce sync.Once
)

type UserConfig struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
	ColabURL string `json:"colab_url"`
}

type Config struct {
	WebhookURL string       `json:"webhook_url"`
	Users      []UserConfig `json:"users"`
}

func loadConfig(envPath, userID string) error {
	// Read env.json file
	data, err := os.ReadFile(envPath)
	if err != nil {
		return fmt.Errorf("nie można odczytać pliku %s: %w", envPath, err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("nie można sparsować pliku env.json: %w", err)
	}

	// Find user by ID
	var user *UserConfig
	for i := range config.Users {
		if config.Users[i].ID == userID {
			user = &config.Users[i]
			break
		}
	}

	if user == nil {
		return fmt.Errorf("nie znaleziono użytkownika z ID: %s", userID)
	}

	// Set variables
	webhookURL = config.WebhookURL
	colabURL = user.ColabURL
	email = user.Email
	password = user.Password

	return nil
}

func screenshotPath(filename string) string {
	if sessionDir != "" {
		return filepath.Join(sessionDir, filename)
	}
	return filename
}

func setupLogging() error {
	// Create debug directory with date subdirectory
	now := time.Now()
	sessionDir = filepath.Join("debug", now.Format("2006-01-02"), now.Format("150405"))
	os.MkdirAll(sessionDir, 0755)

	// Don't use logMessage here as it might not be ready yet
	fmt.Printf("📝 Katalog sesji: %s\n", sessionDir)
	fmt.Printf("📝 Screenshoty zapisywane w: %s\n", sessionDir)

	return nil
}

func killChromeProcesses() {
	fmt.Printf("🔪 Kończę istniejące instancje Chrome...\n")

	// Kill Chrome processes on macOS
	cmd := exec.Command("pkill", "-f", "chrome-user-data")
	err := cmd.Run()
	if err != nil {
		// Ignore error if no processes found
		fmt.Printf("   ℹ️  Nie znaleziono procesów Chrome do zakończenia\n")
	} else {
		fmt.Printf("   ✅ Zamknięto procesy Chrome\n")
	}

	// Wait a bit for processes to fully terminate
	time.Sleep(1 * time.Second)
}

func Run(args []string) int {
	prog := "colab"
	if len(args) > 0 && args[0] != "" {
		prog = filepath.Base(args[0])
	}

	// Define flags
	fs := flag.NewFlagSet(prog, flag.ExitOnError)
	envPath := fs.String("env", "", "ścieżka do pliku env.json (wymagane)")
	userIDFlag := fs.String("user", "", "ID użytkownika (wymagane)")

	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "❌ Użycie: %s -env <path> -user <user_id>\n", prog)
		fmt.Fprintf(fs.Output(), "   Przykład: %s -env ./env.json -user test1\n\n", prog)
		fmt.Fprintf(fs.Output(), "Flagi:\n")
		fs.PrintDefaults()
	}

	// Parse flags (skip program name)
	if err := fs.Parse(args[1:]); err != nil {
		return 1
	}

	// Validate required flags
	if *envPath == "" || *userIDFlag == "" {
		fs.Usage()
		return 1
	}

	userID = *userIDFlag
	fmt.Printf("🔑 Używasz konfiguracji użytkownika: %s\n", userID)
	fmt.Printf("📁 Plik konfiguracji: %s\n\n", *envPath)

	// Setup logging to file
	if err := setupLogging(); err != nil {
		log.Printf("Nie można skonfigurować logowania: %v", err)
		return 1
	}

	// Load configuration from env.json
	if err := loadConfig(*envPath, userID); err != nil {
		log.Printf("❌ Błąd ładowania konfiguracji: %v", err)
		return 1
	}

	fmt.Printf("✅ Załadowano konfigurację z pliku %s\n", *envPath)
	fmt.Printf("   User ID: %s\n", userID)
	fmt.Printf("   Email: %s\n", email)
	fmt.Printf("   Colab URL: %s\n", colabURL)
	if webhookURL != "" {
		fmt.Printf("   Webhook URL: %s\n", webhookURL)
	} else {
		fmt.Printf("   Webhook URL: nie ustawiony\n")
	}
	fmt.Printf("\n")

	// Kill existing Chrome processes
	killChromeProcesses()

	// Uruchom Playwright
	pw, err := playwright.Run()
	if err != nil {
		log.Printf("Nie można uruchomić Playwright: %v", err)
		return 1
	}
	defer pw.Stop()

	// Użyj dedykowanego katalogu dla Playwright (unikalny per użytkownik)
	userDataDir := filepath.Join(".", "chrome-user-data", userID)
	os.MkdirAll(userDataDir, 0755)

	// Konwertuj na ścieżkę bezwzględną
	userDataDir, err = filepath.Abs(userDataDir)
	if err != nil {
		log.Printf("Nie można uzyskać ścieżki bezwzględnej: %v", err)
		return 1
	}

	fmt.Printf("🔷 Używam Chrome z katalogiem danych: %s\n", userDataDir)
	fmt.Println("ℹ️  Sesja zostanie zapisana - przy kolejnym uruchomieniu będziesz automatycznie zalogowany!")
	fmt.Println("")

	// Uruchom Chrome w trybie widocznym
	context, err := pw.Chromium.LaunchPersistentContext(userDataDir, playwright.BrowserTypeLaunchPersistentContextOptions{
		Headless: playwright.Bool(false),
		Channel:  playwright.String("chrome"),
		Args: []string{
			"--disable-blink-features=AutomationControlled",
			"--start-maximized",
			"--no-first-run",
			"--disable-features=TranslateUI",
		},
		Viewport: &playwright.Size{
			Width:  1920,
			Height: 1080,
		},
	})
	if err != nil {
		log.Printf("Nie można uruchomić przeglądarki: %v", err)
		return 1
	}
	if context == nil {
		log.Printf("Kontekst przeglądarki jest nil")
		return 1
	}
	// defer context.Close() // Wyłączone - przeglądarka ma pozostać otwarta po znalezieniu URL

	// Otwórz nową stronę (context już ma dostępną przeglądarkę)
	page, err := context.NewPage()
	if err != nil {
		log.Printf("Nie można otworzyć strony: %v", err)
		return 1
	}

	// Listener konsoli - przechwytuj wszystkie logi
	var foundGradioURL string
	// Listener dla wszystkich komunikatów z konsoli
	consoleMessages := []string{}
	page.On("console", func(msg playwright.ConsoleMessage) {
		text := msg.Text()
		consoleMessages = append(consoleMessages, text)

		if strings.Contains(text, "gradio.live") || strings.Contains(text, "Running on") {
			// Wyciągnij URL z tekstu
			re := regexp.MustCompile(`(https?://[a-zA-Z0-9_.-]+\.gradio\.live[^\s\)]*)`)
			matches := re.FindStringSubmatch(text)
			if len(matches) > 0 {
				foundGradioURL = matches[1]
				fmt.Printf("🎉 Znaleziono URL w konsoli: %s\n", foundGradioURL)
			}
		}
	})

	// Listener dla żądań sieciowych - szukaj URL Gradio i sprawdź status
	page.On("response", func(response playwright.Response) {
		url := response.URL()
		if strings.Contains(url, "gradio.live") {
			fmt.Printf("🎉 Znaleziono URL w odpowiedzi sieciowej: %s\n", url)

			// Sprawdź status code
			status := response.Status()
			if status >= 200 && status < 300 {
				fmt.Printf("✅ URL zwraca status %d (OK)\n", status)

				// Wyciągnij podstawowy URL (tylko domena: subdomain.gradio.live)
				// Regex: https://(subdomain.gradio.live)/...
				var baseURL string
				re := regexp.MustCompile(`https?://([a-zA-Z0-9_.-]+\.gradio\.live)`)
				matches := re.FindStringSubmatch(url)
				if len(matches) > 1 {
					baseURL = "https://" + matches[1]
				} else {
					// Fallback - stare podejście
					baseURL = url
					if idx := strings.Index(baseURL, "//"); idx != -1 {
						baseURL = baseURL[idx+2:]
					}
					if idx := strings.Index(baseURL, "/"); idx != -1 {
						baseURL = baseURL[:idx]
					}
					baseURL = "https://" + baseURL
				}

				fmt.Printf("📤 Wysyłam na webhook: %s\n", baseURL)

				if webhookURL != "" && !urlSent {
					sendToWebhook(baseURL)
					urlSent = true
					fmt.Println("✅ URL został wysłany")
				} else if webhookURL == "" {
					fmt.Println("⚠️  WEBHOOK_URL nie jest ustawiony, pomijam wysyłkę")
				} else {
					fmt.Println("ℹ️  URL został już wcześniej wysłany, pomijam duplikat")
				}

				startKeepAlive(page)
				fmt.Println("🌐 Chrome pozostaje otwarty — keep-alive aktywny")
			} else {
				fmt.Printf("⚠️  URL zwraca status %d (pomijam)\n", status)
			}
		}
	})

	// Listener dla żądań - szukaj URL Gradio
	page.On("request", func(request playwright.Request) {
		url := request.URL()
		if strings.Contains(url, "gradio.live") {
			fmt.Printf("🎉 Znaleziono URL w żądaniu: %s\n", url)
		}
	})

	fmt.Println("🚀 Otwieranie Google Colab...")
	if _, err := page.Goto(colabURL, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
		Timeout:   playwright.Float(60000),
	}); err != nil {
		log.Printf("Nie można otworzyć strony Colab: %v", err)
		return 1
	}

	// Poczekaj dłużej, aby strona się całkowicie załadowała
	fmt.Println("⏳ Ładowanie strony...")
	time.Sleep(5 * time.Second)

	// Sprawdź, czy trzeba się zalogować
	fmt.Println("🔐 Sprawdzanie logowania...")
	if needsLogin(page) {
		fmt.Println("")
		fmt.Println("╔════════════════════════════════════════════════════════════════╗")
		fmt.Println("║  🔑 WYMAGANE LOGOWANIE                                        ║")
		fmt.Println("║                                                                ║")
		fmt.Println("║  Google blokuje automatyczne logowanie Playwright.            ║")
		fmt.Println("║  Musisz zalogować się RĘCZNIE w otwartej przeglądarce.        ║")
		fmt.Println("║                                                                ║")
		fmt.Println("║  ✓ Zaloguj się normalnie w przeglądarce                       ║")
		fmt.Println("║  ✓ Sesja zostanie zapisana w katalogu chrome-user-data/       ║")
		fmt.Println("║  ✓ Przy następnym uruchomieniu już będziesz zalogowany!       ║")
		fmt.Println("║                                                                ║")
		fmt.Println("║  ⏳ Czekam maksymalnie 5 minut na zalogowanie...              ║")
		fmt.Println("╚════════════════════════════════════════════════════════════════╝")
		fmt.Println("")

		if err := waitForManualLogin(page); err != nil {
			log.Printf("Błąd logowania: %v", err)
			return 1
		}
		fmt.Println("✅ Zalogowano pomyślnie! Sesja została zapisana.")
	} else {
		fmt.Println("✅ Już zalogowany (używam zapisanej sesji)!")
	}

	// Poczekaj na załadowanie notebooka
	fmt.Println("⏳ Czekam na załadowanie notebooka...")
	time.Sleep(10 * time.Second)

	// Zrób screenshot dla debugowania
	page.Screenshot(playwright.PageScreenshotOptions{
		Path: playwright.String(screenshotPath("1_notebook_loaded.png")),
	})
	fmt.Printf("📸 Screenshot 1: Notebook załadowany\n")

	// Zapisz HTML do pliku dla debugowania
	html, err := page.Content()
	if err == nil {
		htmlPath := screenshotPath("page_source.html")
		err = os.WriteFile(htmlPath, []byte(html), 0644)
		if err == nil {
			fmt.Printf("📄 HTML zapisany do: %s\n", htmlPath)
		}
	}

	// Zapisz logi konsoli do pliku
	if len(consoleMessages) > 0 {
		consolePath := screenshotPath("console_logs.txt")
		consoleText := strings.Join(consoleMessages, "\n")
		err = os.WriteFile(consolePath, []byte(consoleText), 0644)
		if err == nil {
			fmt.Printf("📋 Logi konsoli zapisane do: %s (%d linii)\n", consolePath, len(consoleMessages))
		}
	}

	// NAJPIERW sprawdź czy URL Gradio już istnieje (notebook może być już uruchomiony)
	fmt.Println("🔍 Sprawdzam czy notebook jest już uruchomiony...")
	existingURL, err := checkForExistingGradioURL(page)
	if err == nil && existingURL != "" {
		// Zapewnij że URL zawsze zaczyna się od https://
		if !strings.HasPrefix(existingURL, "https://") {
			if strings.HasPrefix(existingURL, "http://") {
				existingURL = strings.Replace(existingURL, "http://", "https://", 1)
			} else {
				existingURL = "https://" + existingURL
			}
		}

		fmt.Println("")
		fmt.Println("═══════════════════════════════════════════════════════")
		fmt.Println("✅ NOTEBOOK JUŻ URUCHOMIONY - ZNALEZIONO URL!")
		fmt.Println("")
		fmt.Printf("🔗 %s\n", existingURL)
		fmt.Println("")
		fmt.Println("═══════════════════════════════════════════════════════")
		fmt.Println("")

		// Zrób screenshot
		page.Screenshot(playwright.PageScreenshotOptions{
			Path: playwright.String(screenshotPath("2_existing_gradio.png")),
		})
		fmt.Printf("📸 Screenshot 2: Znaleziono istniejący URL Gradio\n")

		// Wyślij URL na webhook (tylko raz)
		if webhookURL != "" && !urlSent {
			sendToWebhook(existingURL)
			urlSent = true
		} else if webhookURL == "" {
			fmt.Println("ℹ️  WEBHOOK_URL nie jest ustawiony, pomijam wysyłkę")
		} else {
			fmt.Println("ℹ️  URL został już wcześniej wysłany, pomijam duplikat")
		}

		startKeepAlive(page)
		fmt.Println("")
		fmt.Println("⏳ Przeglądarka pozostanie otwarta...")
		fmt.Println("   Naciśnij Ctrl+C aby zakończyć")
		select {}
	}

	fmt.Println("ℹ️  Notebook nie jest uruchomiony lub nie znaleziono URL")
	fmt.Println("   Wykonuję pełny proces uruchomienia...")

	// Kliknij w menu Runtime
	fmt.Println("🖱️  Klikam w menu Runtime...")
	time.Sleep(2 * time.Second)

	err = page.Locator("text=Runtime").First().Click()
	if err != nil {
		log.Printf("Błąd kliknięcia Runtime: %v", err)
		return 1
	}

	time.Sleep(2 * time.Second)

	// Zrób screenshot po otwarciu menu
	page.Screenshot(playwright.PageScreenshotOptions{
		Path: playwright.String(screenshotPath("3_runtime_menu.png")),
	})
	fmt.Printf("📸 Screenshot 3: Menu Runtime otwarte\n")

	// Kliknij "Change runtime type"
	fmt.Println("🖱️  Klikam 'Change runtime type'...")
	time.Sleep(1 * time.Second)

	// Użyj last() żeby pominąć elementy w kodzie notebooka
	err = page.Locator("text=Change runtime type").Last().Click()
	if err != nil {
		log.Printf("Błąd kliknięcia Change runtime type: %v", err)
		return 1
	}

	time.Sleep(3 * time.Second)

	// Zrób screenshot dialogu
	page.Screenshot(playwright.PageScreenshotOptions{
		Path: playwright.String(screenshotPath("4_runtime_dialog.png")),
	})
	fmt.Printf("📸 Screenshot 4: Dialog zmiany runtime\n")

	// Wybierz T4 GPU z Hardware accelerator (radiobutton)
	fmt.Println("🎮 Wybieram T4 GPU...")
	time.Sleep(1 * time.Second)

	// Spróbuj różne selektory dla radiobutton T4 GPU
	err = page.Locator("text=T4 GPU").Click()
	if err != nil {
		log.Printf("Błąd wyboru T4 GPU: %v", err)
		return 1
	}

	time.Sleep(2 * time.Second)

	// Zrób screenshot po wyborze GPU
	page.Screenshot(playwright.PageScreenshotOptions{
		Path: playwright.String(screenshotPath("5_gpu_selected.png")),
	})
	fmt.Printf("📸 Screenshot 5: GPU T4 wybrany\n")

	// Kliknij Save
	fmt.Println("💾 Zapisuję zmiany (Save)...")
	time.Sleep(3 * time.Second) // Więcej czasu na załadowanie dialogu

	// Użyj Playwright locator - czeka automatycznie na element
	fmt.Println("  🔍 Szukam przycisku Save...")

	// Spróbuj po prostu kliknąć najprościej jak się da - w modalu musi być przycisk
	err = page.GetByRole("button", playwright.PageGetByRoleOptions{
		Name: "Save", // Zwykły string, nie wskaźnik
	}).Click(playwright.LocatorClickOptions{
		Timeout: playwright.Float(10000),
		Force:   playwright.Bool(true),
	})

	if err != nil {
		// Jeśli nie znalazł przez role, spróbuj przez tekst
		fmt.Println("  ⚠️  Nie znaleziono przez role, próbuję przez tekst...")
		err = page.GetByText("Save", playwright.PageGetByTextOptions{
			Exact: playwright.Bool(true),
		}).First().Click(playwright.LocatorClickOptions{
			Timeout: playwright.Float(10000),
			Force:   playwright.Bool(true),
		})

		if err != nil {
			log.Printf("Błąd kliknięcia Save: %v", err)
			return 1
		}
	}

	fmt.Println("  ✓ Kliknięto Save")

	time.Sleep(3 * time.Second)

	// Zrób screenshot po kliknięciu Save
	page.Screenshot(playwright.PageScreenshotOptions{
		Path: playwright.String(screenshotPath("6_after_save.png")),
	})
	fmt.Printf("📸 Screenshot 6: Po zapisaniu zmian runtime\n")

	// Poczekaj na restart runtime
	fmt.Println("⏳ Czekam na restart runtime (15 sekund)...")
	time.Sleep(15 * time.Second)

	// Zarządzaj sesjami - zamknij inne sesje przed Run all
	fmt.Println("🔌 Zarządzanie sesjami (zamykanie innych sesji)...")
	if err := manageSessions(page); err != nil {
		fmt.Printf("⚠️  Ostrzeżenie: Problem z zarządzaniem sesjami: %v\n", err)
		fmt.Println("   Kontynuuję dalej...")
	}

	// Kliknij przycisk "Run all" widoczny na stronie
	fmt.Println("▶️  Uruchamianie wszystkich komórek (Run all)...")
	fmt.Println("  🔍 Szukam przycisku 'Run all' na stronie...")
	time.Sleep(3 * time.Second)

	// Spróbuj najpierw znaleźć przycisk przez aria-label lub title
	clicked, err := page.Evaluate(`() => {
		// Szukaj przycisku z aria-label lub title zawierającym "Run all"
		const buttons = Array.from(document.querySelectorAll('button, [role="button"]'));

		for (const btn of buttons) {
			const ariaLabel = btn.getAttribute('aria-label') || '';
			const title = btn.getAttribute('title') || '';
			const text = btn.textContent || '';

			console.log('Sprawdzam przycisk:', {
				ariaLabel,
				title,
				text: text.trim().substring(0, 50),
				tagName: btn.tagName
			});

			if (ariaLabel.toLowerCase().includes('run all') ||
			    title.toLowerCase().includes('run all') ||
			    (text.trim().toLowerCase() === 'run all' && btn.offsetParent !== null)) {
				console.log('✓ Znalazłem przycisk Run all!');
				btn.click();
				return { success: true, method: 'button', label: ariaLabel || title || text.trim() };
			}
		}

		// Jeśli nie znaleziono przycisku, spróbuj skrótu klawiszowego
		console.log('Nie znaleziono przycisku, używam Ctrl+F9');
		return { success: false, method: 'not_found' };
	}`)
	if err != nil {
		log.Printf("Błąd szukania przycisku Run all: %v", err)
		return 1
	}

	fmt.Printf("  ✓ Rezultat: %v\n", clicked)

	// Jeśli nie znaleziono przycisku, użyj skrótu klawiszowego
	clickedMap, ok := clicked.(map[string]interface{})
	if ok && !clickedMap["success"].(bool) {
		fmt.Println("  ⌨️  Używam skrótu klawiszowego Ctrl+F9...")
		err = page.Keyboard().Press("Control+F9")
		if err != nil {
			log.Printf("Błąd użycia skrótu klawiszowego: %v", err)
			return 1
		}
	}

	time.Sleep(2 * time.Second)

	// Screenshot po Run all
	page.Screenshot(playwright.PageScreenshotOptions{
		Path: playwright.String(screenshotPath("7_run_all.png")),
	})
	fmt.Printf("📸 Screenshot 7: Po uruchomieniu Run all\n")

	// Czekaj na URL Gradio
	fmt.Println("")
	fmt.Println("🔗 Czekam na URL Gradio (maksymalnie 15 minut - modele się pobierają)...")
	fmt.Println("   ℹ️  To może potrwać 5-10 minut jeśli modele są pobierane po raz pierwszy")

	gradioURL := ""
	startTime := time.Now()
	maxWaitTime := 15 * time.Minute

	checkCount := 0
	lastProgressMinutes := -1

	for time.Since(startTime) < maxWaitTime {
		checkCount++
		elapsed := time.Since(startTime)
		elapsedMinutes := int(elapsed.Minutes())
		elapsedSeconds := int(elapsed.Seconds())

		// Pokaż progress co 30 sekund
		if elapsedMinutes != lastProgressMinutes && elapsedSeconds%30 == 0 {
			fmt.Printf("   ⏳ Minęło %d minut %d sekund, sprawdzam... (próba #%d)\n", elapsedMinutes, elapsedSeconds%60, checkCount)
			lastProgressMinutes = elapsedMinutes
		}

		// ZAWSZE scrolluj przed sprawdzeniem - agresywne scrollowanie
		page.Evaluate(`() => {
			// Scroll główny dokument - kilka razy aby załadować wszystkie elementy
			for (let i = 0; i < 10; i++) {
				window.scrollBy(0, 1000);
			}
			window.scrollTo(0, document.body.scrollHeight);

			// Scroll również wszystkie output containery
			const outputs = document.querySelectorAll('.output, .output_area, .output_wrapper, [class*="output"]');
			outputs.forEach(el => {
				if (el.scrollHeight > el.clientHeight) {
					el.scrollTop = el.scrollHeight;
				}
			});

			// Scroll wszystkie scrollowalne divy
			const scrollables = document.querySelectorAll('div[style*="overflow"]');
			scrollables.forEach(el => {
				if (el.scrollHeight > el.clientHeight) {
					el.scrollTop = el.scrollHeight;
				}
			});
		}`)
		time.Sleep(2 * time.Second) // Czekaj na załadowanie elementów

		// Szukaj linków zawierających gradio.live
		url, err := page.Evaluate(`() => {
			// Szukaj w linkach <a href> (również w iframe)
			const links = Array.from(document.querySelectorAll('a[href*="gradio.live"]'));
			if (links.length > 0) {
				console.log('✓ Znaleziono link gradio.live:', links[0].href);
				return { found: true, url: links[0].href, method: 'link' };
			}

			// PRZESZUKAJ IFRAmy!
			console.log('🔍 Przeszukuję iframe...');
			const iframes = document.querySelectorAll('iframe');
			console.log('Znaleziono iframe:', iframes.length);
			
			for (let i = 0; i < iframes.length; i++) {
				const iframe = iframes[i];
				console.log('Iframe #' + i + ' src:', iframe.src, 'class:', iframe.className);
				
				try {
					const iframeDoc = iframe.contentDocument || iframe.contentWindow.document;
					const iframeText = iframeDoc.body ? iframeDoc.body.textContent : iframeDoc.textContent;
					const textLength = iframeText ? iframeText.length : 0;
					console.log('Iframe #' + i + ' text length:', textLength);
					
					if (textLength > 0) {
						// Szukaj URL
						const gradioMatch = iframeText.match(/(https?:\/\/[a-zA-Z0-9_.-]+\.gradio\.live[^\s\)]*)/i);
						if (gradioMatch) {
							console.log('✓ Znaleziono URL w iframe #' + i + ':', gradioMatch[1]);
							return { found: true, url: gradioMatch[1], method: 'iframe_text', iframeIndex: i };
						}
						
						// Szukaj linków
						const iframeLinks = Array.from(iframeDoc.querySelectorAll('a[href*="gradio.live"]'));
						if (iframeLinks.length > 0) {
							console.log('✓ Znaleziono link w iframe #' + i + ':', iframeLinks[0].href);
							return { found: true, url: iframeLinks[0].href, method: 'iframe_link', iframeIndex: i };
						}
					}
				} catch (e) {
					console.log('Nie można odczytać iframe #' + i + ':', e.message);
				}
			}
			
			console.log('⚠️ Nie znaleziono URL w żadnym iframe');

			// Szukaj w zawartości tekstowej (również w ukrytych elementach)
			let allTexts = [];

			// 1. Pełny textContent dokumentu
			allTexts.push(document.body.textContent || '');

			// 2. Wszystkie output komórek Colab
			const outputSelectors = [
				'.output',
				'.output_area',
				'.output_wrapper',
				'.output_result',
				'.output_text',
				'.output_stream',
				'[class*="output"]',
				'[id*="output"]',
				'pre',
				'code',
				'div[class*="text"]',
				'div[class*="result"]'
			];

			outputSelectors.forEach(selector => {
				const elements = document.querySelectorAll(selector);
				elements.forEach(el => {
					allTexts.push(el.textContent || '');
					allTexts.push(el.innerText || '');
				});
			});

			// 3. Połącz wszystko
			const combinedText = allTexts.join(' ');

			// Szukaj URL gradio.live z różnymi wariantami (również z newline)
			const patterns = [
				/(https?:\/\/[a-zA-Z0-9_-]+\.gradio\.live[^\s\)]*)/i,
				/(https?:\/\/[a-f0-9-]+\.gradio\.live[^\s\)]*)/i,
				/(https:\/\/[a-z0-9\-]+\.[a-z0-9\-]+\.gradio\.live[^\s\)]*)/i
			];

			for (const pattern of patterns) {
				const match = combinedText.match(pattern);
				if (match) {
					let foundUrl = match[1].trim().replace(/\n/g, ''); // Usuń nowe linie
					// ZAWSZE używaj https://
					if (!foundUrl.startsWith('https://')) {
						if (foundUrl.startsWith('http://')) {
							foundUrl = foundUrl.replace('http://', 'https://');
						} else {
							foundUrl = 'https://' + foundUrl;
						}
					}
					console.log('✓ Znaleziono URL gradio.live w tekście:', foundUrl);
					return { found: true, url: foundUrl, method: 'text', pattern: pattern.toString() };
				}
			}
			
			// Sprawdź też czy jest tylko domena bez protokołu
			const domainOnlyPattern = /([a-z0-9_-]+\.gradio\.live)/i;
			const domainMatch = combinedText.match(domainOnlyPattern);
			if (domainMatch) {
				const foundUrl = 'https://' + domainMatch[1];
				console.log('✓ Znaleziono domenę gradio.live (bez protokołu), dodano https://:', foundUrl);
				return { found: true, url: foundUrl, method: 'text', pattern: 'domain_only' };
			}

			// DEBUG: Jeśli znaleziono "Running on", pokaż fragment
			const runningOnIndex = combinedText.toLowerCase().indexOf('running on');
			if (runningOnIndex >= 0) {
				const fragment = combinedText.substring(Math.max(0, runningOnIndex - 50), runningOnIndex + 300);
				console.log('DEBUG: Znaleziono "running on", fragment:', fragment);
			}
			
			return { found: false, debug: combinedText.length > 0 ? 'text_found' : 'no_text', combinedLength: combinedText.length };
		}`)

		if err == nil && url != nil {
			resultMap, ok := url.(map[string]interface{})
			if ok && resultMap["found"].(bool) {
				urlStr := resultMap["url"].(string)
				method := resultMap["method"].(string)
				fmt.Printf("   ✓ Znaleziono URL (metoda: %s, próba #%d)\n", method, checkCount)
				gradioURL = urlStr
				break
			} else if checkCount%12 == 0 { // Co minutę (12 * 5s = 60s)
				debugStr := ""
				if debugVal, hasDebug := resultMap["debug"]; hasDebug {
					debugStr = fmt.Sprintf(", debug: %v", debugVal)
				}
				if length, hasLength := resultMap["combinedLength"]; hasLength {
					debugStr += fmt.Sprintf(", text length: %v", length)
				}
				fmt.Printf("   🔍 Sprawdzam... (długość tekstu: %v%s)\n", resultMap["combinedLength"], debugStr)
			}
		}

		// Czekaj 5 sekund przed następnym sprawdzeniem
		time.Sleep(5 * time.Second)
	}

	if gradioURL != "" {
		// Zapewnij że URL zawsze zaczyna się od https://
		if !strings.HasPrefix(gradioURL, "https://") {
			if strings.HasPrefix(gradioURL, "http://") {
				gradioURL = strings.Replace(gradioURL, "http://", "https://", 1)
			} else {
				gradioURL = "https://" + gradioURL
			}
		}

		fmt.Println("")
		fmt.Println("═══════════════════════════════════════════════════════")
		fmt.Println("✅ ZNALEZIONO URL GRADIO!")
		fmt.Println("")
		fmt.Printf("🔗 %s\n", gradioURL)
		fmt.Println("")
		fmt.Println("═══════════════════════════════════════════════════════")
		fmt.Println("")

		// Zrób screenshot z URL
		page.Screenshot(playwright.PageScreenshotOptions{
			Path: playwright.String(screenshotPath("10_gradio_found.png")),
		})
		fmt.Printf("📸 Screenshot 10: Znaleziono URL Gradio!\n")

		// Wyślij URL na webhook (tylko raz)
		if webhookURL != "" && !urlSent {
			sendToWebhook(gradioURL)
			urlSent = true
		} else if webhookURL == "" {
			fmt.Println("ℹ️  WEBHOOK_URL nie jest ustawiony, pomijam wysyłkę")
		} else {
			fmt.Println("ℹ️  URL został już wcześniej wysłany, pomijam duplikat")
		}
	} else {
		fmt.Println("")
		fmt.Println("⚠️  Timeout - nie znaleziono URL Gradio w ciągu 15 minut")
		fmt.Println("   Sprawdź przeglądarkę manualnie - URL może się jeszcze pojawić")
	}

	// STOP - czekamy na ocenę
	startKeepAlive(page)
	fmt.Println("")
	fmt.Println("⏳ Przeglądarka pozostanie otwarta - możesz sprawdzić efekty...")
	fmt.Println("   Naciśnij Ctrl+C aby zakończyć")

	// Czekaj nieskończenie długo, aby użytkownik mógł sprawdzić
	select {}

	return 0
}

func startKeepAlive(page playwright.Page) {
	keepAliveOnce.Do(func() {
		fmt.Println("🛟 Startuję keep-alive — co 2 minuty wysyłam aktywność do Colaba")

		go func() {
			ticker := time.NewTicker(keepAliveInterval)
			defer ticker.Stop()

			for {
				<-ticker.C

				_, err := page.Evaluate(`() => {
					try {
						window.scrollBy(0, 50);
						window.scrollBy(0, -50);
						const event = new MouseEvent('mousemove', { bubbles: true, clientX: 1, clientY: 1 });
						if (document.body) {
							document.body.dispatchEvent(event);
						}
					} catch (error) {
						console.log('Keep-alive exception', error?.message || error);
					}
					return true;
				}`)

				if err != nil {
					errText := err.Error()
					fmt.Printf("⚠️  Keep-alive błąd: %v\n", err)
					if strings.Contains(strings.ToLower(errText), "target closed") ||
						strings.Contains(strings.ToLower(errText), "browser has been closed") {
						fmt.Println("🛑 Przerywam keep-alive — przeglądarka została zamknięta")
						return
					}
					continue
				}

				fmt.Printf("💓 Keep-alive ping wysłany (%s)\n", time.Now().Format("15:04:05"))
			}
		}()
	})
}

func sendToWebhook(url string) {
	fmt.Println("📤 Wysyłam URL na webhook...")

	payload := map[string]string{
		"url":       url,
		"timestamp": time.Now().Format(time.RFC3339),
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("❌ Błąd serializacji JSON: %v\n", err)
		return
	}

	req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("❌ Błąd tworzenia żądania: %v\n", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ Błąd wysyłki na webhook: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		fmt.Println("✅ URL został wysłany na webhook pomyślnie")
	} else {
		fmt.Printf("⚠️  Webhook zwrócił status: %d\n", resp.StatusCode)
	}
}

func needsLogin(page playwright.Page) bool {
	// Sprawdź, czy jest na stronie jakiś element związany z logowaniem
	emailInput, _ := page.QuerySelector("input[type='email']")
	signInBtn, _ := page.QuerySelector("button:has-text('Sign in'), a:has-text('Sign in')")
	return emailInput != nil || signInBtn != nil
}

func waitForManualLogin(page playwright.Page) error {
	// Czekaj maksymalnie 5 minut
	timeout := time.After(5 * time.Minute)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return fmt.Errorf("timeout: użytkownik nie zalogował się w ciągu 5 minut")
		case <-ticker.C:
			// Sprawdź czy użytkownik jest już zalogowany
			// Jeśli nie ma już formularza logowania i jesteśmy na stronie Colab
			if !needsLogin(page) && strings.Contains(page.URL(), "colab.research.google.com") {
				// Dodatkowe sprawdzenie - czy załadował się notebook
				time.Sleep(2 * time.Second)
				return nil
			}
			fmt.Print(".")
		}
	}
}

func checkForExistingGradioURL(page playwright.Page) (string, error) {
	// Agresywne scrollowanie - dokładnie tak samo jak w głównej pętli
	page.Evaluate(`() => {
		// Scroll główny dokument - kilka razy aby załadować wszystkie elementy
		for (let i = 0; i < 10; i++) {
			window.scrollBy(0, 1000);
		}
		window.scrollTo(0, document.body.scrollHeight);

		// Scroll również wszystkie output containery
		const outputs = document.querySelectorAll('.output, .output_area, .output_wrapper, [class*="output"]');
		outputs.forEach(el => {
			if (el.scrollHeight > el.clientHeight) {
				el.scrollTop = el.scrollHeight;
			}
		});

		// Scroll wszystkie scrollowalne divy
		const scrollables = document.querySelectorAll('div[style*="overflow"]');
		scrollables.forEach(el => {
			if (el.scrollHeight > el.clientHeight) {
				el.scrollTop = el.scrollHeight;
			}
		});
	}`)
	time.Sleep(3 * time.Second) // Czekaj na załadowanie elementów

	// Zrób screenshot po scrollowaniu
	page.Screenshot(playwright.PageScreenshotOptions{
		Path: playwright.String(screenshotPath("0_after_scroll.png")),
	})
	fmt.Printf("📸 Screenshot po scrollowaniu\n")

	// Szukaj URL Gradio - dokładnie tak samo jak w głównej pętli
	url, err := page.Evaluate(`() => {
		// Szukaj w linkach (również w iframe)
		const links = Array.from(document.querySelectorAll('a[href*="gradio.live"]'));
		if (links.length > 0) {
			console.log('✓ Znaleziono link gradio.live:', links[0].href);
			return links[0].href;
		}

		// PRZESZUKAJ IFRAmy!
		console.log('🔍 Przeszukuję iframe...');
		const iframes = document.querySelectorAll('iframe');
		console.log('Znaleziono iframe:', iframes.length);
		
		for (let i = 0; i < iframes.length; i++) {
			const iframe = iframes[i];
			console.log('Iframe #' + i + ' src:', iframe.src, 'class:', iframe.className);
			
			try {
				const iframeDoc = iframe.contentDocument || iframe.contentWindow.document;
				const iframeText = iframeDoc.body ? iframeDoc.body.textContent : iframeDoc.textContent;
				const textLength = iframeText ? iframeText.length : 0;
				console.log('Iframe #' + i + ' text length:', textLength);
				
				if (textLength > 0) {
					// Szukaj URL
					const gradioMatch = iframeText.match(/(https?:\/\/[a-zA-Z0-9_.-]+\.gradio\.live[^\s\)]*)/i);
					if (gradioMatch) {
						console.log('✓ Znaleziono URL w iframe #' + i + ':', gradioMatch[1]);
						return gradioMatch[1];
					}
					
					// Szukaj linków
					const iframeLinks = Array.from(iframeDoc.querySelectorAll('a[href*="gradio.live"]'));
					if (iframeLinks.length > 0) {
						console.log('✓ Znaleziono link w iframe #' + i + ':', iframeLinks[0].href);
						return iframeLinks[0].href;
					}
				}
			} catch (e) {
				console.log('Nie można odczytać iframe #' + i + ':', e.message);
			}
		}
		
		console.log('⚠️ Nie znaleziono URL w żadnym iframe');

		// Przeszukaj WSZYSTKIE elementy na stronie - także te ukryte
		let allTexts = [];

		// 1. Pełny textContent dokumentu
		allTexts.push(document.body.textContent || '');

		// 2. Wszystkie output komórek Colab
		const outputSelectors = [
			'.output',
			'.output_area',
			'.output_wrapper',
			'.output_result',
			'.output_text',
			'.output_stream',
			'[class*="output"]',
			'[id*="output"]',
			'pre',
			'code',
			'div[class*="text"]',
			'div[class*="result"]'
		];

		outputSelectors.forEach(selector => {
			const elements = document.querySelectorAll(selector);
			elements.forEach(el => {
				allTexts.push(el.textContent || '');
				allTexts.push(el.innerText || '');
			});
		});

		// 3. Połącz wszystko
		const combinedText = allTexts.join(' ');

		// Szukaj URL gradio.live z różnymi wariantami (również z newline)
		const patterns = [
			/(https?:\/\/[a-zA-Z0-9_-]+\.gradio\.live[^\s\)]*)/i,
			/(https?:\/\/[a-f0-9-]+\.gradio\.live[^\s\)]*)/i,
			/(https:\/\/[a-z0-9\-]+\.[a-z0-9\-]+\.gradio\.live[^\s\)]*)/i
		];

		for (const pattern of patterns) {
			const match = combinedText.match(pattern);
			if (match) {
				let foundUrl = match[1].trim().replace(/\n/g, ''); // Usuń nowe linie
				// ZAWSZE używaj https://
				if (!foundUrl.startsWith('https://')) {
					if (foundUrl.startsWith('http://')) {
						foundUrl = foundUrl.replace('http://', 'https://');
					} else {
						foundUrl = 'https://' + foundUrl;
					}
				}
				console.log('✓ Znaleziono URL gradio.live w tekście:', foundUrl);
				return foundUrl;
			}
		}
		
		// Sprawdź też czy jest tylko domena bez protokołu (nie powinno się zdarzać, ale na wszelki wypadek)
		const domainOnlyPattern = /([a-z0-9_-]+\.gradio\.live)/i;
		const domainMatch = combinedText.match(domainOnlyPattern);
		if (domainMatch) {
			const foundUrl = 'https://' + domainMatch[1];
			console.log('✓ Znaleziono domenę gradio.live (bez protokołu), dodano https://:', foundUrl);
			return foundUrl;
		}

		return "";
	}`)

	if err != nil {
		return "", err
	}

	if urlStr, ok := url.(string); ok && urlStr != "" {
		return urlStr, nil
	}

	return "", nil
}

func manageSessions(page playwright.Page) error {
	fmt.Printf("  🖱️  Otwieranie menu Runtime...\n")
	time.Sleep(2 * time.Second)

	// Kliknij w menu Runtime
	err := page.Locator("text=Runtime").First().Click()
	if err != nil {
		return fmt.Errorf("nie można kliknąć menu Runtime: %w", err)
	}

	time.Sleep(2 * time.Second)

	// Screenshot po otwarciu menu
	page.Screenshot(playwright.PageScreenshotOptions{
		Path: playwright.String(screenshotPath("11_runtime_menu.png")),
	})
	fmt.Printf("  📸 Screenshot 11: Menu Runtime otwarte\n")

	// Kliknij "Manage sessions"
	fmt.Printf("  🔍 Klikanie 'Manage sessions'...\n")
	err = page.Locator("text=Manage sessions").First().Click(playwright.LocatorClickOptions{
		Timeout: playwright.Float(5000),
	})
	if err != nil {
		return fmt.Errorf("nie można kliknąć 'Manage sessions': %w", err)
	}

	time.Sleep(3 * time.Second)

	// Screenshot dialogu manage sessions
	page.Screenshot(playwright.PageScreenshotOptions{
		Path: playwright.String(screenshotPath("12_manage_sessions_dialog.png")),
	})
	fmt.Printf("  📸 Screenshot 12: Dialog Manage sessions\n")

	// Kliknij przycisk "Terminate other sessions" w modalu
	fmt.Printf("  🗑️  Klikanie 'Terminate other sessions'...\n")
	err = page.Locator("button:has-text('Terminate other sessions')").First().Click(playwright.LocatorClickOptions{
		Timeout: playwright.Float(5000),
	})
	if err != nil {
		fmt.Printf("  ⚠️  Nie można kliknąć 'Terminate other sessions' - może nie ma innych sesji, kontynuuję...\n")
	} else {
		time.Sleep(2 * time.Second)

		// Screenshot po kliknięciu
		page.Screenshot(playwright.PageScreenshotOptions{
			Path: playwright.String(screenshotPath("13_terminated.png")),
		})
		fmt.Printf("  📸 Screenshot 13: Terminate clicked\n")
	}

	// Zamknij dialog
	fmt.Printf("  🔙 Zamykam dialog...\n")
	err = page.Locator("button:has-text('Close'), button:has-text('Done')").First().Click(playwright.LocatorClickOptions{
		Timeout: playwright.Float(5000),
	})
	if err != nil {
		// Try ESC key
		page.Keyboard().Press("Escape")
	}

	time.Sleep(2 * time.Second)

	fmt.Printf("  ✅ Zarządzanie sesjami zakończone\n")
	return nil
}
