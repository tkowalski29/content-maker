# D10 — Cross-check: Core 20 (GA-TTD/GMC-TTD) vs kod produktu

Data: 2026-08-26 | Kontekst: przełożony (COO) dostarczył pakiet `Sembot-com-Handoff-2026-08-25` z pełną specyfikacją 20 procesów ("Core 20": `GA-TTD-01–15`, `GMC-TTD-01–05`) planowanych do publikacji na sembot.com. Specyfikacje żyją w zewnętrznym systemie planowania COO ("Sembot Atomic Process Factory"), NIE w kodzie. Ten dokument sprawdza dla każdego z 20, czy i jak bardzo jest to dziś ugruntowane w `sembot-laravel`/`module-feed`.

## ⚠️ Cel i granice tego dokumentu
To NIE jest ocena czy proces "zadziała" — to sprawdzenie czy istnieje w kodzie logika realizująca dokładnie to, co opisuje specyfikacja (progi, pola API, warunki brzegowe). Status `publication_status` z `PUBLIC_PROCESS_CARDS.md` jest wewnętrzny — ten dokument dostarcza materiał do wypełnienia punktu "potwierdzonego statusu implementacji" z ich własnego "Publication gate", nie zastępuje decyzji Product.

## Google Ads

| ID | Werdykt | Najbliższy kod | Czym się różni |
|---|---|---|---|
| GA-TTD-01 | BRAK | `ConvertingProductNotAvailable` (wyłączony, sprawdza odwrotny warunek) | brak logiki "koszt bez zakupu na poziomie produktu" |
| GA-TTD-02 | BRAK | `BudgetExhausted` | brak elementu rentowności/ROAS |
| GA-TTD-03 | BRAK | — | zero logiki porównania alokacji budżetu między kampaniami |
| GA-TTD-04 | CZĘŚCIOWE | `CampaignNoActivity.php` (clicks, próg %), `ViewChanges.php` (impressions, próg %) | oba to progi procentowe, nie deterministyczny gate "impressions == 0" |
| GA-TTD-05 | CZĘŚCIOWE | `NoConversionsFromAdTags.php` | brak bramki minimalnego ruchu, brak baseline historycznego, brak filtra do kategorii Purchase |
| GA-TTD-06 | CZĘŚCIOWE (martwy kod) | `NegativeKeywords.php` | wyłączony w `codes.php`; `manageTask()` zakomentowany w samym pliku — nigdy nie tworzy taska nawet gdyby włączony; logika to naiwne przecięcie tekstu, nie match-type semantics |
| GA-TTD-07 | BRAK | — | zero pojęcia "produkt strategiczny" w kodzie |
| GA-TTD-08 | BRAK / poza repo | `NoChangesInAccountHistory.php` (odwrotna logika: alarmuje przy BRAKU zmian) | opis etykiety AI (`auto_category_change_history_alert`) pasuje tematycznie, ale realny workflow żyje w n8n-runnerze, niedostępny |
| GA-TTD-09 | BRAK | — | zero wzmianek "measurement policy"/`goal_config_level` w całym repo |
| GA-TTD-10 | BRAK | `ConvertingProductNotAvailable.php` (wyłączony) | autorzy sami oznaczyli komentarzem jako niedokończony szkielet |
| GA-TTD-11 | CZĘŚCIOWE | `AdsLeadTo404.php` + `TaskTriggerService::checkUrl()` | pojedynczy `Http::get()`, bez retry/timeout/backoff, tylko kod 404 (nie 5xx/przekierowania/connection refused) — spec wymaga "stabilnego dowodu" z powtórnymi próbami |
| GA-TTD-12 | BRAK | — | zero śladu `always_use_default_value` w repo |
| GA-TTD-13 | CZĘŚCIOWE | `ExcludedKeywords.php` | działa na `keyword_view` (istniejące słowo kluczowe), nie `search_term_view`; brak pola kosztu w warunku w ogóle |
| GA-TTD-14 | BRAK / poza repo | opis etykiety `auto_category_smart_bidding_health` | plik workflow n8n nie istnieje w tym repo, nie da się zweryfikować progów |
| GA-TTD-15 | CZĘŚCIOWE | `workflow-templates/budget_deviation_v2.json` | realna logika projekcji istnieje, ale: `status: inactive`, opt-in per klient, działa per-kampania (nie na poziomie planu miesiąca konta) |

## Merchant Center / feed

| ID | Werdykt | Uwaga |
|---|---|---|
| GMC-TTD-01 | BRAK progu | legacy trigger (`ProductGmcWarning`) tworzy task dla KAŻDEGO issue z API — potwierdzone testem `GmcTaskTriggersTest.php` (`count(issues) === count(created_tasks)`), zero filtra po revenue/transakcjach/clicks |
| GMC-TTD-02 | CZĘŚCIOWE | `ProcessSourcesController::gmcDatafeedStatuses` daje surowy `last_upload_date`; zero logiki SLA/harmonogramu/maintenance window. `module-feed SyncProfiler` to fałszywy trop — mierzy wydajność faz przetwarzania, nie świeżość |
| GMC-TTD-03 | CZĘŚCIOWE | current snapshot + generyczny mechanizm "poprzedni run" (`ProcessRunController::previousRun`) istnieją; sama logika progu/delty prawdopodobnie liczona w n8n, poza tym repo |
| GMC-TTD-04 | BRAK | sprawdzone wszystkie tropy (unit pricing validator, CSS sync, scraper cen konkurencji) — żaden nie porównuje ceny feedu z ceną sklepu |
| GMC-TTD-05 | BRAK | istniejący kod łączy dostępność z feedu z wydatkami Google Ads (inny cel: `oos_products_with_spend`), nie feed vs sklep |

## Wniosek ogólny

Żaden z 20 procesów nie ma dziś w kodzie logiki spełniającej dokładnie warunki brzegowe ze specyfikacji. Tam gdzie coś jest tematycznie bliskie, to w większości: **(a)** martwy/wyłączony kod legacy triggerów (GA-TTD-06, częściowo 09/10), **(b)** surowe dane bez logiki decyzyjnej — próg/SLA/delta prawdopodobnie liczony w zewnętrznym n8n-runnerze poza zasięgiem tego repo (GA-TTD-08/14/15, GMC-TTD-02/03), albo **(c)** całkowity brak jakiegokolwiek zalążka (GA-TTD-01/02/03/07/12, GMC-TTD-01/04/05).

**Dla Publication gate:** żaden z 20 procesów nie powinien dziś dostać zaznaczenia "potwierdzony status implementacji" na podstawie tego repo. Dla GA-TTD-08/14/15 i GMC-TTD-02/03 rekomendacja to potwierdzenie u zespołu odpowiedzialnego za n8n-runner — realna logika może tam istnieć, ale nie da się tego zweryfikować z `/workspace`.
