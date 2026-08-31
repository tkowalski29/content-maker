# Sembot.com — podsumowanie sesji (2026-08-21 → 2026-08-31)

> Dokument przygotowany przez Claude (AI) jako obiektywne podsumowanie faktycznych działań i ustaleń z tej konkretnej rozmowy — nie zawiera nowych analiz ani spekulacji poza tym, co faktycznie sprawdzono. Miejsca niepewne lub niepotwierdzone są jawnie oznaczone.
>
> Kontekst techniczny sesji: praca w devcontainerze z dostępem do kodu narzędzia Sembot (`sembot-laravel`, `sembot-angular`, moduły Composer), równolegle do innej sesji Claude (`service_proxy`) budującej samą stronę sembot.com bez dostępu do tego kodu.

---

## 0. Jak używać tego dokumentu

Ten plik ma służyć jako punkt startowy dla nowej sesji (prawdopodobnie na środowisku `new.sembot.com`). Sekcja 1 zawiera najważniejszą, nierozwiązaną kwestię — przeczytaj ją jako pierwszą, zanim zaczniesz cokolwiek kontynuować z reszty dokumentu.

---

## 1. ⚠️ Najważniejsza otwarta kwestia — do rozstrzygnięcia na starcie nowej sesji

Pod koniec tej rozmowy IT/PM zwrócili uwagę, że zrzuty ekranu i nagrania wykorzystane w materiałach wideo pochodzą ze środowiska dostępnego w **tym** devcontainerze (repo `sembot-laravel`/`sembot-angular`, branch `master`/`main`, zalogowane jako `app.sembot.com` wg konta testowego), a nie z **`new.sembot.com`**.

**Nie ustalono w tej rozmowie:**
- czy `new.sembot.com` to inny branch/repozytorium (inny kod) niż to, które analizowałem,
- czy to tylko inna domena/deployment tego samego kodu.

**Dlaczego to ważne:** jeśli to inny kod, to znaczna część ustaleń z sekcji 4 (zbudowana na analizie `sembot-laravel`/`sembot-angular` w tym środowisku) może dotyczyć **innego produktu** niż ten, który ma reprezentować `new.sembot.com`. Nie da się tego rozstrzygnąć z tej sesji — wymaga potwierdzenia u IT.

**Zalecana pierwsza czynność w nowej sesji:** ustalić dokładny stosunek między repozytorium dostępnym w nowym środowisku a `new.sembot.com`/`app.sembot.com`, i ocenić, które z ustaleń poniżej wymagają powtórzenia weryfikacji.

---

## 2. Cel projektu

Pomoc w tworzeniu nowej strony sembot.com wg narracji **"Process + Task first"**: gotowe procesy monitorują konta reklamowe, tworzą "taski" z dowodem, a Marketer AI pomaga je rozwiązać. Rola tej sesji (z dostępem do kodu narzędzia) to ugruntowywanie treści marketingowej w faktycznym stanie produktu — druga sesja Claude (`service_proxy`) buduje samą stronę i nie ma dostępu do tego kodu.

---

## 3. Dokumenty źródłowe otrzymane od przełożonego (w kolejności chronologicznej)

1. **`Sembot WWW — narracja i architektura strony v0.1.md`** (przekazany 2026-08-21, dokument datowany 2026-08-10) — pierwszy brief: narracja Process+Task, 4 obszary wartości (Bezpieczeństwo/Oszczędności/Plan poprawy/Okazje), architektura homepage, zasady copywritingu.

2. **Pakiet `Sembot-com-Handoff-2026-08-25-v2-all-TTD.zip`** (przekazany 2026-08-25) — pełniejszy, nowszy handoff zawierający:
   - `START_HERE.md` — nadrzędny brief wykonawczy z zatwierdzonymi decyzjami (CTA, trial, pricing, język PL→EN)
   - `AGENT_PROMPT.md` — instrukcje dla sesji budującej stronę (`service_proxy`), zasady bezwzględne (nie wymyślać funkcji/liczb/case studies)
   - `PUBLIC_PROCESS_CARDS.md` — pełna specyfikacja "Core 20" procesów (`GA-TTD-01–15`, `GMC-TTD-01–05`), każdy ze statusem: *"legacy do rewalidacji"* / *"final vX"* / *"koncept Core 20"*
   - `SOURCE_INDEX.md`, `OPEN_DECISIONS_TEMPLATE.md`, `CURRENT_SITE_SNAPSHOT.md`, referencje (ICP, roadmapa, analiza spójności)

   **Ważne ustalenie:** numeracja `GA-TTD-XX`/`GMC-TTD-XX` pochodzi z **wewnętrznego systemu planowania COO** ("Sembot Atomic Process Factory", widoczne w oryginalnych ścieżkach plików w pakiecie), NIE z kodu produktu. Nie da się jej znaleźć grepując repo.

---

## 4. Ustalenia z kodu produktu

> ⚠️ Wszystko w tej sekcji dotyczy repo `sembot-laravel`/`sembot-angular` dostępnego w **tym** devcontainerze. Ważność tych ustaleń dla `new.sembot.com` zależy od rozstrzygnięcia kwestii z sekcji 1.

### 4.1 Koncept "task" z briefu — w kodzie to TRZY osobne mechanizmy, nie jeden

| Mechanizm | Co robi | Status |
|---|---|---|
| **`Task` + `TaskTrigger`** (klasyczny, ~47 detektorów w `app/Services/Projects/Triggers/`) | Monitoring → task z rekomendowaną akcją → user wykonuje ręcznie. Wykonanie akcji jest **natychmiastowe i synchroniczne**, bez dry-run | Działający, ale bez podglądu przed zmianą |
| **`Process`/`ProcessRun`** (nowszy, oparty o zewnętrzny n8n-runner, migracje z lipca 2026) | Proces → task z `data.artifacts` (alert/kpi/chart/table) w "drawerze", dedup, cooldown, powiadomienia | Infrastruktura działa; **katalog ~90 konkretnych procesów żyje w n8n-runnerze, poza tym repo** — nie da się go stąd zweryfikować |
| **`ConnectionExecMutateController` + `MutationAudit` + `McpWorkflowApproval`** | Prawdziwy dry-run→blast_radius→confirmation_token(TTL 5 min)→execute→audit z rollbackiem | Działający, ale tylko dla mutacji z czatu AI/MCP — nie wpięty w klasyczny `Task` ani `Process` |

**Wniosek:** wizualna "karta taska jako dowód z dry-run i approval", którą chce pokazać strona, to w rzeczywistości kompozycja dwóch/trzech różnych mechanizmów produktu, nie jeden gotowy obiekt z kodu.

### 4.2 Cross-check "Core 20" — żaden proces nie ma pełnego pokrycia w kodzie

Zweryfikowano wszystkie 20 procesów ze specyfikacji COO (`PUBLIC_PROCESS_CARDS.md`) względem kodu:

| ID | Werdykt | Najbliższy kod / uwaga |
|---|---|---|
| GA-TTD-01 (produkt bez sprzedaży) | **BRAK** | — |
| GA-TTD-02 (rentowna kampania ogr. budżetem) | **BRAK** | `BudgetExhausted` istnieje, ale bez elementu rentowności |
| GA-TTD-03 (budżet w słabszych kampaniach) | **BRAK** | — |
| GA-TTD-04 (kampania przestała się wyświetlać) | CZĘŚCIOWE | próg %, nie "dokładnie zero" |
| GA-TTD-05 (awaria pomiaru zakupów) | CZĘŚCIOWE | brak bramki ruchu, brak baseline |
| GA-TTD-06 (negative keyword blokuje popyt) | CZĘŚCIOWE, **martwy kod** | wyłączony w configu, `manageTask()` zakomentowany w pliku |
| GA-TTD-07 (produkt poza pokryciem) | **BRAK** | — |
| GA-TTD-08 (zbyt częste zmiany strategii) | BRAK / poza repo | realna logika może być w n8n, niesprawdzalna stąd |
| GA-TTD-09 (dodatkowa akcja Purchase) | **BRAK** | — |
| GA-TTD-10 (zero wyświetleń) | **BRAK** | kod-kandydat sam oznaczony komentarzem jako niedokończony |
| GA-TTD-11 (reklamy na niedziałającą stronę) | CZĘŚCIOWE | pojedynczy `Http::get()` bez retry, tylko kod 404 |
| GA-TTD-12 (Purchase wymusza stałą wartość) | **BRAK** | — |
| GA-TTD-13 (search term bez zakupu) | CZĘŚCIOWE | działa na keyword nie search term, brak pola kosztu |
| GA-TTD-14 (Smart Bidding za mało konwersji) | BRAK / poza repo | — |
| GA-TTD-15 (spend odbiega od planu) | CZĘŚCIOWE | kod istnieje, ale nieaktywny, per-kampania nie per-konto |
| GMC-TTD-01 (odrzucone produkty) | **BRAK progu** | legacy trigger tworzy task dla każdego issue, zero filtra |
| GMC-TTD-02 (SLA świeżości feedu) | CZĘŚCIOWE | surowe dane, zero logiki SLA |
| GMC-TTD-03 (spadek zaakceptowanych produktów) | CZĘŚCIOWE | surowe dane + mechanizm "poprzedni run", zero logiki progu w tym repo |
| GMC-TTD-04 (cena feed vs sklep) | **BRAK** | sprawdzone wszystkie tropy — zero kodu |
| GMC-TTD-05 (dostępność feed vs sklep) | **BRAK** | — |

**Ważny niuans:** te statusy nie są "błędem" strony ani przełożonego — `PUBLIC_PROCESS_CARDS.md` sam oznacza je jako "legacy do rewalidacji"/"koncept". To jest właśnie informacja, którą wymaga ich własny "Publication gate" przed publikacją.

### 4.3 Marketer AI — realne możliwości i granice

- **Nazwa "Marketer AI" jest realna produktowo**, nie tylko marketingowa — używana konsekwentnie w kodzie, tłumaczeniach, powiadomieniach.
- Mechanizm "task → czat AI" **NIE przekazuje pełnego kontekstu taska automatycznie** — przekazywany jest wyłącznie `task_id` jako tekst w promptcie startowym; AI samo dociąga szczegóły przez function-calling, jeśli ich potrzebuje. Model `ChatThread` nie ma nawet pola `task_id` (brak trwałego powiązania wątku z taskiem).
- Realne mechanizmy ograniczające AI: Passport scope'y (osobne dla odczytu/zapisu), whitelist dozwolonych mutacji, rate limiting (10/min, 100/dzień) — nie tekstowy "system prompt bezpieczeństwa" (ten został usunięty z bazy w kwietniu 2026, prawdopodobnie żyje teraz w osobnym serwisie `sembot_public-mcp`, poza tym repo).
- Istnieje tryb **auto-approve** na poziomie workspace — jeśli włączony, AI może wykonać mutację bez pytania człowieka za każdym razem (to nie jest domyślne zachowanie).

### 4.4 Integracje

| Integracja | Status | Zasila automatyczny monitoring (triggery)? |
|---|---|---|
| Google Ads | IMPLEMENTED, read-write | Tak |
| Google Merchant Center | IMPLEMENTED, read-write | Tak |
| Google Analytics 4 | IMPLEMENTED, read-only | Tak |
| Microsoft Advertising | IMPLEMENTED, read-write (węższy zakres) | **Nie** |
| Meta/Facebook Ads | PARTIAL | **Nie** |
| Google Search Console | IMPLEMENTED, read-only | Nie dotyczy |
| Shopify | IMPLEMENTED, read-only (import feedu) | Nie dotyczy |

**Kluczowy wniosek:** automatyczny monitoring (triggery) działa dziś tylko na danych z Google Ads + GMC + GA4 — Microsoft Ads i Meta są podłączalne, ale nie zasilają silnika detekcji.

### 4.5 Pozostałe dostarczone pliki (D1, D3, D5, D8, D9)

- **D1** — pełny katalog 45 aktywnych triggerów (starego systemu) zmapowany na 4 obszary wartości (interpretacja własna, nie fakt z kodu)
- **D3** — 3 syntetyczne przykłady tasków oparte na realnym schemacie (baza dev nie miała żadnych realnych, automatycznie wygenerowanych tasków do wykorzystania)
- **D5** — katalog procesów (ta sama treść co D1, w formacie "co monitoruje / jak często / jaki task produkuje")
- **D8** — glosariusz: task="Zadanie", trigger="Wyzwalacz" (niespójnie też "Trigger"), nowy system="Proces", dry-run=zapożyczone słowo (czasem "PREVIEW" w promptach czatu), workspace="przestrzeń robocza" (unikać słowa "konto" — niejednoznaczne w samym kodzie)
- **D9** — sprawdzone i **udokumentowane jako niewykonalne** w tamtym momencie (brak danych logowania do konta demo) — later odblokowane, patrz sekcja 5

Pełna treść D1-D9 jest w PDF, patrz sekcja 6.

---

## 5. Materiały wideo — chronologia prób i finalny zestaw

### Co próbowano po drodze (wersje zarzucone/superseded)
1. Pierwsza 5-sekundowa próbka z **wymyślonym logo** ("Sembot." wielką literą w Poppins) — **błąd**, poprawiony po sprawdzeniu prawdziwego logo na sembot.com/css.sembot.com (prawdziwe: małe litery "sembot", niestandardowy krój, kolor `#F5C313` zgadzał się przypadkowo)
2. Pełny 40s film v1/v2 — zrekonstruowane UI w Remotion na ciemnym tle, karty "unoszące się" w pustej przestrzeni — po samoocenie jako ekspert content/sales: wygląda jak animowana prezentacja, nie jak prawdziwy produkt; złamana ciągłość między scenami; disclaimer zjadał pierwsze sekundy hooka
3. Film v3 — po znalezieniu danych logowania dev w `CLAUDE.md` tego repo (`test@butsy.pl`/`passpass`) udało się zalogować do **prawdziwego** panelu Sembota (po drodze naprawiono dwa realne błędy środowiska: routing API przez proxy do `nginx:80`, i brakującą kolumnę `last_login_provider` w tabeli `users` — dodaną ręcznym, bezpiecznym `ALTER TABLE`) i przechwycić realne zrzuty (sidebar, top bar, drawer szczegółów taska). Film przebudowany z tymi realnymi elementami jako tło

### ⚠️ Uwaga bezpieczeństwa z tego etapu
Podczas eksploracji panelu (`/panel/.../project-connections`) ujawniły się **prawdziwe adresy e-mail pracowników Sembota** (nie klienta) podpięte do testowego projektu. Nie trafiły do żadnego finalnego materiału, ale warto o tym pamiętać przy dalszej eksploracji podobnych kont testowych.

### Finalny, dostarczony zestaw plików (zoptymalizowany pod wagę strony, bez dźwięku)

| Plik | Format | Waga | Link |
|---|---|---|---|
| `sembot-produkt.webm` | VP9, 1920×1080, główny | 1.05 MB | https://raw.githubusercontent.com/tkowalski29/content-maker/main/sembot-produkt.webm |
| `sembot-produkt.mp4` | h264, 1920×1080, fallback Safari | 2.5 MB | https://raw.githubusercontent.com/tkowalski29/content-maker/main/sembot-produkt.mp4 |
| `sembot-produkt-poster.webp` | klatka z tabelą dowodów + badge "Wymaga decyzji" | 38 KB | https://raw.githubusercontent.com/tkowalski29/content-maker/main/sembot-produkt-poster.webp |
| `sembot-produkt-loop.webm` | 8s bezszwowa pętla, muted | 142 KB | https://raw.githubusercontent.com/tkowalski29/content-maker/main/sembot-produkt-loop.webm |

**Sposób użycia wg ustaleń z sesją budującą stronę:** hero zostaje lekką interaktywną kartą taska (LCP), 40-sekundowy film ląduje jako "Zobacz, jak to działa" na kliknięcie (bez autoplay).

**Uwaga do przemyślenia w nowym środowisku:** ten materiał bazuje na zrzutach z `app.sembot.com` (patrz sekcja 1) — jeśli okaże się, że `new.sembot.com` ma inny wygląd/UI, materiał wideo prawdopodobnie wymaga przeróbki z nowymi zrzutami.

### Techniczna notatka: Remotion vs Playwright
To narzędzia komplementarne, nie konkurencyjne: **Playwright** przechwytuje prawdziwe ekrany (autentyczność), **Remotion** składa z nich historię (napisy, tempo, branding, disclaimer). Remotion Player "na żywo" (bez eksportu do pliku wideo) był rozważany jako sposób na zmniejszenie wagi strony, ale **odrzucony** — strona sembot.com jest w Go+templ, bez Reacta, więc nie da się osadzić żywego komponentu React.

---

## 6. Eksport treści z kodu — plik PDF (D1-D9)

Pełna treść wszystkich 9 plików (D1-D9) + plik STATUS, w formie PDF:
**https://raw.githubusercontent.com/tkowalski29/content-maker/main/sembot-tool-export-2026-08-25.pdf**

⚠️ **Ten PDF NIE zawiera D10** (cross-check "Core 20" powstał później) — pełna treść D10 jest w sekcji 4.2 powyżej tego dokumentu.

---

## 7. Przegląd zbudowanej strony (stan na 2026-08-31)

Adres w momencie przeglądu: `http://172.16.10.1/jarek_rzepa_05_20_1630/` (staging, dostępny z tego devcontainera bez VPN, bo ta sama sieć Dockera co user łączy się przez VPN z zewnątrz; `noindex,nofollow` poprawnie ustawione).

### 7.1 Zgodność treści z kodem — najważniejsze ustalenia

| Priorytet | Ustalenie |
|---|---|
| 🔴 Krytyczne | Wyciekłe notatki redakcyjne wprost w treści strony, odnoszące się do moich plików: *"realne dane: liczby klientów/agencji (D08), case studies... (D09)... do potwierdzenia przez przełożonego"* (homepage) i *"zakres wielokontowy — do potwierdzenia (D10)"* (`/dla-agencji`). Wizualnie oznaczone badge'em "W BUDOWIE" (mniej groźne niż wyglądało w surowym HTML), ale referencje do wewnętrznych ID plików nie powinny zostać w finalnej wersji |
| 🟠 Overclaim | *"Marketer AI zaczyna rozmowę z pełnym kontekstem sprawy"* — powtórzone 4× na stronie. Zweryfikowane w kodzie: przekazywany jest tylko `task_id`, nie pełny kontekst (patrz 4.3) |
| 🟠 Overclaim | `/optymalizator-feeda` przedstawia GMC-TTD-04/05 (cena/dostępność feed vs sklep) jako działające procesy bez żadnego zastrzeżenia — w kodzie to całkowity BRAK (patrz 4.2) |
| 🟡 Do decyzji Product | Wszystkie 12 przykładowych procesów pokazanych na stronie (homepage + `/procesy` + `/dla-ecommerce`) mają w D10 status BRAK lub CZĘŚCIOWE — żaden nie przeszedłby dziś ich własnego "Publication gate" wymagającego potwierdzonej implementacji |
| 🟢 Dobrze | FAQ i sekcja kontroli poprawnie komunikują "tryb odczytu, zero automatycznych zmian bez zgody" — zgodne z kodem |
| 🟢 Dobrze | Hedging przy integracjach ("GA4 i Microsoft Advertising — w zakresie potwierdzonej gotowości procesów") — zgodny z ustaleniami D6 |

### 7.2 Ważne dopowiedzenie: skąd wzięły się te claimy

Sprawdzono źródła — **większość "overclaimów" pochodzi wprost z briefu przełożonego**, nie jest wymysłem sesji budującej stronę:
- 12 przykładowych procesów = dokładnie te wskazane w `AGENT_PROMPT.md` i `PUBLIC_PROCESS_CARDS.md` (sekcja "Strony kategorii")
- "Pełny kontekst sprawy" = dosłowny nagłówek z `START_HERE.md`, Sekcja 7

**Wyjątki, które NIE pochodzą z briefu** (prawdopodobnie artefakt pisania strony):
- Notatki "(D08)/(D09)/(D10)" — format nazewnictwa pasuje do MOICH plików, nie do numeracji z briefu (`D01-D16` w innym kontekście)
- Brak zastrzeżeń przy GMC-TTD-04/05 na `/optymalizator-feeda` — w `PUBLIC_PROCESS_CARDS.md` te procesy MAJĄ sekcję "Ograniczenie", która zgubiła się przy skracaniu do strony

### 7.3 Ocena UX/design (perspektywa web designera 2026)

**Ogólnie: mocne wykonanie** jak na stronę bez frameworka JS (Go+templ+czysty CSS, zero build-stepu):

| Obszar | Ocena |
|---|---|
| Waga strony | 🟢 ~144 KB, 17 requestów, zero błędów konsoli — bardzo lekko. Jedyny minus: 8 plików fontów Google (~140 KB) |
| System komponentu "task-card" | 🟢 Reużywany identycznie w hero/anatomii/przykładach — dobre myślenie systemowe |
| Dostępność | 🟢 `:focus-visible`, skip-link, `prefers-reduced-motion` obsłużony, kontrast WCAG AA skomentowany wprost w CSS |
| Responsywność desktop | 🟢 17 media queries, breakpointy dopasowane per-sekcja |
| **Bug na mobile** | 🔴 Kwota w karcie taska (np. "12 400 zł") **łamie się** na viewport 390px — "zł" spada do nowej linii z dziedziczonym żółtym podkreśleniem, wygląda jak osobny błąd. Do naprawy: `white-space: nowrap` w CSS `.task-card__amount` |
| SEO/meta | 🟢 Canonical, OG, hreflang, JSON-LD, poprawny `noindex,nofollow` na stagingu |

### 7.4 Konwersja / CRO — temat poruszony, analiza NIE wykonana

Ustalono ramę: rola **specjalisty CRO (Conversion Rate Optimization) dla landing page'y pod płatny ruch B2B SaaS** (łączy SEM message-match + UX friction/psychologię konwersji) — user zaakceptował kierunek, ale **właściwa analiza konwersji nie została jeszcze zrobiona w tej sesji**.

---

## 8. Narzędzia i procedury środowiskowe

> Poniższe mogą, ale nie muszą przenieść się 1:1 do nowego środowiska — flagowane osobno.

- **Generowanie PDF do pobrania:** HTML (CSS inline, w `/workspace/docs/`) → render przez Playwright/Chromium (zainstalowany ręcznie w tym środowisku pod tym userem, nie było domyślnie dostępne) → PDF → push do publicznego repo `tkowalski29/content-maker` → link `raw.githubusercontent.com`. To jedyna potwierdzona działająca metoda hostowania plików do pobrania z tego typu środowiska (localhost i GitHub Releases nie działają).
- **Remotion** (generowanie wideo z kodu React) — zainstalowany i przetestowany w `/tmp`, działa bez dodatkowej konfiguracji (własny headless Chromium + enkoder, nie wymaga systemowego ffmpeg).
- **Dane logowania dev** (`test@butsy.pl` / `passpass`) znalezione w `CLAUDE.md` repo `sembot-angular` **tego** środowiska — prawdopodobnie nieadekwatne dla nowego środowiska, nie zakładać że będą identyczne.
- **Napotkane i naprawione błędy tego środowiska** (nie zakładać, że wystąpią w nowym): rozjazd routingu API przez proxy `local` (`environment.local.ts` miał zaszytą ścieżkę kanału), brakująca kolumna `last_login_provider` w bazie mimo odnotowanej migracji.

---

## 9. Nierozwiązane pytania / TODO dla nowej sesji

1. **[Priorytet 1]** Ustalić relację `new.sembot.com` ↔ repo w nowym środowisku (patrz sekcja 1) i ocenić, co z sekcji 4 wymaga powtórzenia.
2. Lista ~90 konkretnych procesów z n8n-runnera — user miał dosłać, nie dostarczono w tej sesji; D1/D5 bazują tylko na starym systemie 47 triggerów.
3. Trzecie narzędzie konkurencyjne, "bdos.ai" (Karol Dziedzic) — zidentyfikowane przez WebSearch (integracje: Google Ads, GA4, Merchant Center, Keyword Planner; ~200 użytkowników miesiąc po premierze), ale głębszy research porównawczy odłożony na później.
4. Naprawa buga mobile (łamanie "zł" w karcie taska) — zgłoszona, nie naprawiona (leży po stronie sesji budującej stronę, ma dostęp do CSS).
5. Wyczyszczenie wyciekłych notatek "(D08)/(D09)/(D10)" ze strony — zgłoszone, nie naprawione.
6. Analiza CRO/konwersji — rama ustalona (specjalista CRO), analiza nie wykonana.
7. Bezpieczne zrzuty/nagrania z prawdziwej produkcji przez Playwright — omówione zasady bezpieczeństwa (dedykowane konto demo, zero klikania w przyciski akcji, anonimizacja), nie wykonane; wymaga danych dostępowych od IT.
8. Blok A (struktura/IA strony) i Blok D (pełny research konkurencji: Optmyzr, Channable) z pierwotnego planu 4-blokowego — nie rozpoczęte w tej sesji.

---

## 10. Zapisane pamięci sesji (auto-memory, dostępne w przyszłych rozmowach w tym środowisku)

Dla porządku — te notatki żyją w systemie pamięci Claude powiązanym z tym userem/środowiskiem, nie w tym pliku:

- `project_sembot-com-website-plan.md` — ogólny plan 4-blokowy pomocy przy stronie
- `project_sembot-tool-export-brief.md` — status brief'u D1-D10, kluczowe ustalenie o 3 mechanizmach (Task/Process/MCP)
- `reference_pdf-generation-procedure.md` — procedura generowania PDF opisana w sekcji 8
- `feedback_mattermost-dot-message.md` — user komunikuje się przez Mattermost, pojedyncza kropka jako wiadomość = prośba o pokazanie poprzedniej odpowiedzi, nie pusta wiadomość

---

*Koniec dokumentu. Wygenerowano na podstawie faktycznego przebiegu rozmowy — bez dodawania informacji spoza tego, co faktycznie ustalono, sprawdzono lub dostarczono.*
