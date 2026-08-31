# D7 — Marketer AI: co realnie robi i jakie ma granice

Data: 2026-08-22 | Źródło: `sembot-laravel` (`app/Services/Ai/Chat/`, `app/Http/Controllers/Api/Connections/ConnectionExecMutateController.php`, `app/Models/Workflows/McpWorkflowApproval.php`, `config/assistant-config.php`, `docs/business/13-ai/DOC_MarketerAI_Flow.md`)

## Nazwa

**"Marketer AI" to realna, oficjalna nazwa produktowa** — nie tylko marketingowa. Używana konsekwentnie: `MarketerAiReplyNotification.php`, tłumaczenia (`notifications.php`: "Marketer AI odpowiedział w Twoim wątku"), UI (`pl.json`, dziesiątki wystąpień), dokumentacja biznesowa `docs/business/13-ai/DOC_MarketerAI_*.md`. W menu nawigacyjnym bywa też ogólnie "Asystent AI" — ale w czacie i powiadomieniach dominuje "Marketer AI".

## ⚠️ Ważne zastrzeżenie architektoniczne

Cały mechanizm "ChatFunctions" (`app/Services/Ai/Chat/ChatFunctions/*`) — czyli klasyczna, w pełni widoczna w tym repo lista narzędzi AI — jest **oficjalnie oznaczony jako deprecated**:

> `ChatFunction.php`: *"The ChatFunctions mechanism is no longer used by the marketer service. Marketer now consumes tools exclusively via MCP (see `sembot_public-mcp` binary). Do NOT add new ChatFunctions."*

Realny, dzisiejszy zestaw narzędzi Marketer AI żyje w osobnym serwisie Go (`sembot_public-mcp`), **poza zasięgiem tego repo/workspace**. Poniższa lista pokazuje więc **zakres kompetencji koncepcyjnie** (co AI historycznie/tematycznie potrafiło i prawdopodobnie nadal potrafi w podobnym zakresie), nie dokładną listę dzisiejszych narzędzi. Jeśli treść ma być w 100% aktualna co do nazw narzędzi, potrzebny jest dostęp do `sembot_public-mcp`.

## Zakres kompetencji (na bazie nazw dawnych ChatFunctions, pogrupowane tematycznie)

| Domena | Co potrafi (przykłady) | Status |
|---|---|---|
| Google Ads | odczyt kampanii/grup reklam/wydajności; **edycja**: budżet kampanii, strategia licytacji, status grupy reklam/kampanii, dodawanie negatywnych słów kluczowych | `[PARTIAL]` — realne narzędzia dziś żyją w MCP poza repo, tu widoczny tylko historyczny zakres |
| Microsoft Ads (Bing) | odczyt wydajności kampanii/grup/słów kluczowych/produktów/search terms | `[PARTIAL]`, tylko odczyt w tym co widać |
| Meta Ads | odczyt wydajności | `[PARTIAL]`, tylko odczyt |
| Google Analytics 4 | odczyt (audience, kampanie, zdarzenia, marka/nazwa produktu, nowi vs. powracający, źródło/medium, wiek/płeć) | `[PARTIAL]`, tylko odczyt |
| Google Merchant Center | odczyt (konkurencyjność cenowa, trendy shopping, top marki/kategorie/produkty) | `[PARTIAL]`, tylko odczyt |
| Katalog produktowy | zmiana etykiet, duplikowanie produktów, wykresy zagregowane, top produkty | `[PARTIAL]` |
| Monitor cen/fraz konkurencji | dodawanie i odczyt zadań monitorujących ceny/frazy | `[PARTIAL]` |
| SEO | ruch, dane słów kluczowych, SERP, widoczność strony | `[PARTIAL]` |
| Taski | tworzenie, wykonanie akcji na tasku, pobranie dostępnych akcji/sekcji/przypisanych osób | `[PARTIAL]` |
| Workflows | uruchomienie workflow | `[PARTIAL]` |
| Baza wiedzy | wyszukiwanie w bazie wiedzy Sembota | `[PARTIAL]` |

Aktualny wzorzec dodawania nowych narzędzi (potwierdzony w dokumentacji repo): endpoint Passport-scoped w `routes/public/v1.php` + narzędzie w `sembot_public-mcp/main.go` (Go, inne repo) + token przypięty przez `mcp_servers`/`agent_mcp_api_keys`.

## Mutacje: czy AI zmienia konto bez pytania człowieka?

**Zależy od trybu.** To jest istotne dla treści strony (§8 brief marketingowego: "kontrola i approval") — nie ma jednej, uniwersalnej odpowiedzi:

1. **Domyślnie: dry-run + confirmation token.** `ConnectionExecMutateController` — 6-warstwowy mechanizm bezpieczeństwa: walidacja scope → `dry_run=true` zwraca `blast_radius` + jednorazowy token (TTL 5 min) → dopiero wywołanie z `dry_run=false` + poprawnym tokenem wykonuje zmianę, zapisując `snapshot_before` do audytu → rate limit 10/min, 100/dzień na workspace → mutacje wysokiego ryzyka wymagają `acknowledge_high_risk=true`. `[IMPLEMENTED]`.
2. **Ale: to nie zawsze jest potwierdzenie CZŁOWIEKA.** Rytuał dry-run→token jest po stronie protokołu MCP/LLM — model sam może wykonać obie rundy bez pytania usera w UI, **chyba że** workspace ma **wyłączony** auto-approve.
3. **Auto-approve na poziomie workspace.** Jeśli workspace ma włączone `auto_approve_<platforma>` (lub legacy `auto_approve_workflows`), AI może pominąć rundę tokena i wykonać mutację od razu (audit dostaje wtedy `auto_approved=true`). `[IMPLEMENTED]` — to oznacza, że **jest tryb, w którym AI faktycznie zmienia konto bez pytania człowieka za każdym razem**, ale wyłącznie jako świadomy wybór właściciela workspace'u, nie domyślne zachowanie.
4. **Prawdziwe, jawne "Zatwierdź/Odrzuć" w UI: `McpWorkflowApproval`.** Dla mutacji w workflow, które zatrzymują się na węźle `mcp_call` — user dostaje powiadomienie in-app + push z przyciskami Approve/Reject, i dopiero jego decyzja odblokowuje dalsze wykonanie. `[IMPLEMENTED]`. To jedyne miejsce z pewnością, że wymagane jest jawne kliknięcie człowieka (poza auto-approve).
5. **Rollback.** Audit trail (`mcp_mutation_audit`) trzyma `snapshot_before/after` przez 24h i pozwala cofnąć mutację. `[IMPLEMENTED]`.

**Rekomendacja dla treści:** pisać "zmiany wymagają zatwierdzenia, chyba że firma świadomie włączy tryb auto-approve dla danej platformy" — nie "Sembot zawsze pyta o zgodę". To jest zgodne z zasadą z brief'u przełożonego (§16: "automatycznie poprawia konto — tylko po wskazaniu zakresu gotowych akcji i approval"), ale trzeba dodać niuans auto-approve.

## Granice — czego AI nie robi bez człowieka / guardrails

- **Brak w tym repo tekstowego "system promptu bezpieczeństwa".** Kolumna `system_prompt` została usunięta z tabeli `agents` migracją (`2026_04_01_000003_refactor_agents_table.php`) — nowe agenty tworzone są bez wbudowanego promptu w tym repo. Realny prompt najprawdopodobniej żyje w `sembot_public-mcp` lub warstwie orkiestracji LLM poza zasięgiem tego audytu — **brak potwierdzenia w kodzie, nie zgaduję treści**.
- **Legacy prompt** (`config/assistant-config.php`, praktycznie martwy) zawierał tylko instrukcję routingu do subagentów, nie guardrail bezpieczeństwa.
- **Realne, egzekwowalne granice to nie prompt, tylko access control:**
  - Passport scope'y osobno dla odczytu i zapisu każdej domeny (`products-api-read`/`write`, `ads-execution`/`ads-execution-write`, `tasks-write` itd.) — token agenta/MCP decyduje co narzędzie może w ogóle zawołać.
  - `ConnectionExecMutationPolicy` — whitelist dozwolonych mutacji per platforma + hard-deny operacji destrukcyjnych (np. reset Smart Bidding, DELETE).
  - Rate limiting: 10 mutacji/min, 100/dzień na workspace.
  - Osobny, czysto odczytowy kanał (`ConnectionExecController`) odrzuca mutujące żądania z 403.
  - Moderacja treści usera (hate speech/violence) przed przetworzeniem wiadomości — to ogranicza input usera, nie output/akcje AI.

**Rekomendacja dla treści:** zamiast "AI ma wbudowane zasady etyczne", pisać uczciwie "dostęp AI do zmian jest ograniczony przez uprawnienia (scope), whitelistę dozwolonych operacji i limity — nie przez instrukcję tekstową, którą AI mogłoby zignorować".

## `MarketerAiReplyNotification` — kiedy user dostaje info o odpowiedzi

Wysyłana **przy każdej** odpowiedzi AI w wątku czatu (nie tylko przy specjalnych zdarzeniach), wyłącznie in-app (celowo bez maila — komentarz w kodzie: *"marketer odpowiada często; mailowy spam byłby agresywny"*), z wyciągiem odpowiedzi (max 120 znaków). Domyślnie włączona, user może wyłączyć w ustawieniach.

---

## ⚠️ Pytania otwarte
1. Dokładna, aktualna lista narzędzi AI żyje w `sembot_public-mcp` (Go, inne repo) — potrzebny dostęp, jeśli treść ma być precyzyjna co do "co dziś Marketer AI potrafi zrobić".
2. Treść realnego system promptu/guardrails — nie znaleziona w tym repo, prawdopodobnie żyje poza nim. Nie zgaduję treści.
3. Czy strona ma wspominać tryb auto-approve (AI działa bez pytania człowieka, gdy workspace to włączy) — to prawdziwa funkcja, ale wymaga ostrożnego sformułowania, żeby nie zaprzeczyć obietnicy "kontrola i approval".
