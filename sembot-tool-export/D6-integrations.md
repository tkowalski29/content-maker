# D6 — Realnie wspierane integracje / źródła danych

Data: 2026-08-21 | Źródło: `sembot-laravel` (`app/Types/Connections/ConnectionType.php`, `ConnectionServiceSlug.php`, `app/Services/Actions/Implementations/`, `app/Services/Connections/ConnectionExecMutationPolicy.php`, `app/Services/Projects/Triggers/`, `module-feed`)

## ⚠️ Rozróżnienie kluczowe dla treści strony

Sembot ma **więcej podłączalnych integracji niż aktywnych "procesów monitorujących"**. Automatyczny monitoring/taski (triggery) dziś czerpią dane **tylko z Google Ads, Google Merchant Center i Google Analytics 4** (`app/Types/Projects/TaskTriggerType.php:10-13,37-41` — formalnie zdefiniowane tylko te 3 typy + wewnętrzny `sembot`). Microsoft Advertising, Meta/Facebook Ads i Google Search Console są podłączalne i częściowo zapisywalne, ale **nie zasilają dziś silnika proaktywnych triggerów/tasków**. Strona nie powinna sugerować, że każda integracja = automatyczny monitoring.

## Tabela zbiorcza

| Integracja | Status | Zakres | Zasila triggery/taski automatycznie? | Kluczowy dowód (src) |
|---|---|---|---|---|
| **Google Ads** | `[IMPLEMENTED]` | read-write (pełny silnik akcji + MCP + campaign sync z feeda) | tak | `app/Services/Actions/Implementations/*.php` (24 klasy akcji); `ConnectionExecMutationPolicy::checkGoogleAds()` |
| **Google Merchant Center** | `[IMPLEMENTED]` | read-write (raporty/issues + push produktów przez Content API) | tak | `app/Jobs/Google/Merchant/GoogleMerchantCenterSyncProductsJob.php:134` |
| **Google Analytics 4** | `[IMPLEMENTED]` | read-only (GA4 Data API nie ma operacji zapisu) | tak | `app/Services/Projects/Triggers/AverageSessionDurationDecrease.php:20` |
| **Microsoft Advertising (Bing Ads)** | `[IMPLEMENTED]` | read-write (campaign sync Shopping z feeda przez SOAP + MCP mutate, węższy zakres akcji niż Google) | **nie** | `app/Jobs/Campaigns/Bing/BingSyncCampaign.php`; `ConnectionExecMutationPolicy::checkMicrosoftAds()` |
| **Meta / Facebook Ads** | `[PARTIAL]` | read-write ograniczony (raporty + katalog produktowy + Advantage+ Shopping campaigns + MCP status/budget only) — brak dedykowanego silnika akcji jak dla Google Ads | **nie** | `app/Services/Facebook/FacebookClient.php:107-134,198`; `ConnectionExecMutationPolicy::checkMetaAds()` |
| **Google Search Console** | `[IMPLEMENTED]` | read-only | **nie** | `app/Services/Google/SearchConsole/SearchConsoleService.php` |
| **Shopify** | `[IMPLEMENTED]` | read-only (import katalogu produktowego, źródło feedu wejściowego) | nie dotyczy (to źródło feedu, nie kanał reklamowy) | `ConnectionType::inputFeedConnections()`; `ParserShopifyGraphql.php` (module-feed) |
| **Feed produktowy (CSV/XML)** | `[IMPLEMENTED]` | read-write (import + eksport na Google/Facebook/Bing) | pośrednio — zasila jakość danych, na której pracują triggery GMC | `module-feed/laravel/src/Services/Parse`, `.../Generate` |
| **Gmail/IMAP** | `[IMPLEMENTED]`, ale **nie jest to integracja marketingowa** | wewnętrzne (parsowanie maili do systemu komentarzy/tasków supportowych) | nie | `app/Services/MailInbox/Inboxes/InboxGmail.php` — **nie umieszczać na stronie jako "integrację danych"** |

## Szczegóły per integracja

### Google Ads
Najgłębsza integracja. Odczyt: raporty/GAQL w `Domain/GoogleAds/`. Zapis: 24 klasy akcji w `app/Services/Actions/Implementations/` (np. `IncreaseBudget.php`, `PauseKeyword.php`, `RemoveKeyword.php`, `ExcludeKeyword.php`) wykonywane przez SDK `Google\Ads\GoogleAds\V25`. Osobno: MCP allow-list na mutacje (`ConnectionExecMutationPolicy::checkGoogleAds()`) z tieringiem ryzyka low/medium/high i hard-gate na reset Smart Bidding. Kampanie Shopping/PMax/Search budowane automatycznie z feeda (`app/Jobs/Strategies/Campaigns/Google/Sync{Shopping,Pmax,Search}Campaign.php`).

### Microsoft Advertising (Bing Ads)
**Nie jest to integracja "tylko czytająca"** — potwierdzony realny SOAP-owy zapis kampanii Shopping z feeda (`BingSyncCampaign.php`: `UpdateCampaignsRequest`, `AddAdGroupsRequest`, `ApplyProductPartitionActionsRequest`) plus kontrolowany kanał mutacji ad-hoc przez MCP (allow-list: `UpdateCampaigns`, `UpdateAdGroups`, `UpdateKeywords`, `AddBudgets`, `UpdateBidStrategies` jako high-risk; denylist na `Delete*`/`Remove*`). Health check oznaczony `//tbd` w kodzie — niedopracowany, ale samo połączenie i mutacje działają produkcyjnie.

### Google Merchant Center
Odczyt: issues konta/produktu (`Domain/GoogleMerchantCenter/Commands/*`, wykorzystywane przez triggery `AccountGmcWarning`, `ProductGmcWarning`). Zapis: realny push produktów przez Content API (`GoogleMerchantCenterSyncProductsJob.php`). CSS (Comparison Shopping Service) jako rozszerzenie: `app/Services/Google/Css/GoogleCssClient.php` — osobna synchronizacja lokalnych produktów do CSS.

### Google Analytics 4
Czysto odczytowa — sama natura GA4 Data API nie oferuje zapisu poza konfiguracją property, której Sembot nie rusza. Zasila 2 triggery (`HighRejectionRate`, `AverageSessionDurationDecrease`) i 8 getterów w Automations.

### Meta / Facebook Ads
`[PARTIAL]` — świadomie nie `[IMPLEMENTED]` na równi z Google Ads: jest odczyt raportowy, zapis katalogu produktowego do Commerce Manager, tworzenie kampanii Advantage+ Shopping, i wąski MCP-proxy (tylko pola `status`, `daily_budget`, `lifetime_budget`, `bid_amount`, `name`) — ale **brak** dedykowanych granularnych akcji (nie ma `PauseAdFacebook`/`IncreaseBudgetFacebook` analogicznych do Google Ads), i kanał nie zasila triggerów/tasków.

### Google Search Console
Wyłącznie odczyt (sama natura GSC API). Osobny `ConnectionServiceSlug::GOOGLE_SEARCH_CONSOLE`, dashboard w `DashboardSearchConsoleController.php`. Nie zasila triggerów.

### Shopify
Źródło feedu wejściowego (import katalogu produktów przez GraphQL Admin API), nie kanał do którego Sembot zapisuje z powrotem. Brak dowodów na zapis do Shopify w przeszukanym kodzie.

### Feed produktowy
Formaty wejściowe: CSV, XML (w tym streamowany), Shopify GraphQL API. Formaty/kierunki wyjściowe (eksport z Sembota): Google Shopping/PMax, Facebook Catalog, Bing Shopping campaigns — wszystkie potwierdzone realnymi jobami synchronizacji.

---

## ⚠️ Pytania otwarte

1. Czy strona ma w ogóle wspominać Microsoft Advertising/Meta jako integracje, skoro nie zasilają dziś silnika automatycznych triggerów/tasków (mogą być mylące przy narracji "procesy pilnują kont") — czy pokazać je osobno jako "podłącz dane, kampanie i feed", bez obietnicy automatycznego monitoringu?
2. Gmail/IMAP nie powinien pojawiać się w sekcji integracji marketingowych — czy w ogóle wspominać go na stronie, czy pominąć całkowicie?
