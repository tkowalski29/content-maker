# Taksonomia typów tasków

Data: 2026-08-21 | Źródło: `sembot-laravel`, `data/trigger/{google_ads,gmc,google_analytics_4,sembot}/codes.php` + `defaults.php` + `resources/lang/pl/tasks.php`

## ⚠️ Zakres tego dokumentu — przeczytać przed użyciem

Ten katalog obejmuje **wyłącznie stary, klasyczny system "task triggerów"** (45 aktywnych detektorów PHP, `app/Services/Projects/Triggers/`). Istnieje też **nowszy system `Process`** (n8n-runner, ok. 90 procesów) — architektonicznie bliższy narracji strony, ale jego katalog żyje poza tym repo. User dostarczy tę listę osobno; gdy dojdzie, ten dokument wymaga aktualizacji/scalenia.

**Kategorie (Bezpieczeństwo/Oszczędności/Plan poprawy/Okazje) nie istnieją jako pole w kodzie** — to moja interpretacja na bazie opisu triggera, nie fakt z bazy danych. Część triggerów jest dwuznaczna (oznaczone niżej) — decyzja, do której kategorii je przypisać, zależy od narracji, jaką zespół marketingowy chce nadać.

Wszystkie poniższe triggery mają status `[IMPLEMENTED]` (aktywne w `codes.php`), chyba że zaznaczono inaczej.

---

## Obszar: Bezpieczeństwo

- **Spadek liczby konwersji** (`Conversions`, kod 3) — liczba konwersji spadła o więcej niż X% w 7 dni — trigger: codzienny odczyt Google Ads — src: `app/Services/Projects/Triggers/Conversions.php`
- **Zakończenie kampanii** (`CampaignEndsSoon`, kod 6) — pozostało mniej niż X dni do końca kampanii — trigger: daily — src: `CampaignEndsSoon.php`
- **Ograniczenia reklam** (`AdDisapproved`, kod 7) — Google odrzucił reklamy — trigger: hourly, priorytet high — src: `AdDisapproved.php`
- **Nieaktualne linki** (`InappropriateLink`, kod 10) — nieaktualne linki w reklamach — trigger: hourly, priorytet high — src: `InappropriateLink.php`
- **Zmiana konkurencji** (`NewCompetitor`, kod 11) — nowi konkurenci dla głównych słów kluczowych — trigger: weekly — src: `NewCompetitor.php` *(dwuznaczne — może też być "Okazja")*
- **Problemy z geo-targetingiem** (`GeoTargetingIssues`, kod 15) — znacząca ilość kliknięć spoza wybranej lokalizacji — trigger: weekly, priorytet high — src: `GeoTargetingIssues.php` *(w kodzie komentarz "nieukończony — brak danych na koncie", traktować ostrożnie)*
- **Zmiany w wyświetleniach** (`ViewChanges`, kod 16) — spadek wyświetleń o X% w 7 dni — trigger: weekly — src: `ViewChanges.php`
- **Brak aktywności w kampanii** (`CampaignNoActivity`, kod 18) — spadek kliknięć o X%, priorytet high — trigger: daily — src: `CampaignNoActivity.php`
- **Reklamy z nieaktualną frazą** (`OutdatedAd`, kod 24) — reklamy promujące wydarzenia/oferty z nieaktualną frazą — trigger: weekly — src: `OutdatedAd.php`
- **Grupy reklam bez aktywnych reklam** (`NoActiveAdsInGroup`, kod 25) — trigger: weekly — src: `NoActiveAdsInGroup.php`
- **Limit wykluczających słów kluczowych** (`NegativeKeywordsLimit`, kod 30) — zbliżanie się do limitu — trigger: weekly — src: `NegativeKeywordsLimit.php`
- **Brak konwersji z tagów reklam** (`NoConversionsFromAdTags`, kod 47) — trigger: daily — src: `NoConversionsFromAdTags.php`
- **Wyłączone automatyczne tagowanie** (`DisabledAutoTagging`, kod 50) — brak gclid — trigger: daily — src: `DisabledAutoTagging.php`
- **Reklama kieruje do 404** (`AdsLeadTo404`, kod 51) — trigger: weekly — src: `AdsLeadTo404.php`
- **Reklama display w aplikacjach mobilnych** (`DisplayAdAppearsInMobileApps`, kod 52) — trigger: daily — src: `DisplayAdAppearsInMobileApps.php`
- **Włączone automatyczne rekomendacje Google** (`AutomaticRecommendationsEnabled`, kod 54) — Google może samo zmieniać ustawienia konta — trigger: daily — src: `AutomaticRecommendationsEnabled.php`
- **Powiadomienia o produktach w GMC** (`ProductGmcWarning`, kod 1gmc) — problemy/błędy produktowe w Merchant Center — trigger: priorytet high — src: `ProductGmcWarning.php`
- **Powiadomienia o koncie w GMC** (`AccountGmcWarning`, kod 2gmc) — trigger: priorytet high — src: `AccountGmcWarning.php`
- **Błąd połączenia z usługą** (`ProjectConnectionError`, kod 3sembot) — trigger: wewnętrzny — src: `ProjectConnectionError.php`
- **Brakujące tagi śledzące** (`MissingTags`, kod 19) — `[PLANNED]`, zakomentowany w kodzie — konwersje przestały zbierać dane — src: `data/trigger/google_ads/codes.php:19` (wyłączony)
- **Zmiany w dostępności produktów konwertujących** (`ConvertingProductNotAvailable`, kod 39) — `[PLANNED]`, zakomentowany, komentarz w kodzie "brak statusu produktu, dodać z GMC" — src: `data/trigger/google_ads/codes.php:42`

## Obszar: Oszczędności

- **Zmiany w śr. CPC** (`CpcChanges`, kod 4) — wzrost średniego CPC powyżej progów — trigger: weekly — src: `CpcChanges.php`
- **Niska efektywność słów kluczowych** (`ExcludedKeywords`, kod 5) — dużo kliknięć, mało konwersji — trigger: weekly — src: `ExcludedKeywords.php`
- **Wzrost kosztu konwersji** (`ConversionCostsChanges`, kod 22) — trigger: weekly, priorytet high — src: `ConversionCostsChanges.php`
- **Wzrost CPM** (`CpmChanges`, kod 29) — trigger: weekly, priorytet high — src: `CpmChanges.php`
- **Zmiany w kosztach reklam wideo** (`VideoAdCostsIncrease`, kod 32) — trigger: weekly — src: `VideoAdCostsIncrease.php`
- **Zmiany w koszcie kampanii** (`CampaignCostChanges`, kod 38) — koszt odbiega od poprzedniego dnia — trigger: daily — src: `CampaignCostChanges.php`
- **Wykluczenie brandu** (`BrandExclusion`, kod 49) — brand nie wykluczony z kampanii non-brand — trigger: daily — src: `BrandExclusion.php`
- **Słowa kluczowe zablokowane przez wykluczenia** (`NegativeKeywords`, kod 26) — `[PLANNED]`, zakomentowany — src: `data/trigger/google_ads/codes.php:26`

## Obszar: Plan poprawy

- **Spadek CTR** (`FluctuationsCtr`, kod 2) — trigger: weekly — src: `FluctuationsCtr.php`
- **Skuteczność reklam** (`AdEffectiveness`, kod 8) — różnice konwersji/CTR w grupie reklam — trigger: weekly — src: `AdEffectiveness.php`
- **Zmiany w jakości reklam** (`AdQualityDecrease`, kod 12) — trigger: daily — src: `AdQualityDecrease.php`
- **Brakujące rozszerzenia reklam** (`MissingExtensions`, kod 13) — trigger: weekly — src: `MissingExtensions.php`
- **Optymalizacja dla urządzeń** (`OptimizationForDevices`, kod 14) — jedno urządzenie znacząco słabsze — trigger: weekly — src: `OptimizationForDevices.php`
- **Optymalizacja harmonogramu godzinowego** (`HourlyScheduleOptimization`, kod 17) — najgorsza godzina dnia — trigger: daily — src: `HourlyScheduleOptimization.php`
- **Różnice mobile-desktop** (`MobileCampaignEffectiveness`, kod 21) — różnica konwersji/ROAS/CPC — trigger: weekly — src: `MobileCampaignEffectiveness.php`
- **Spadek jakości słów kluczowych** (`LowQualityKeywords`, kod 23) — trigger: weekly — src: `LowQualityKeywords.php`
- **Brak harmonogramu reklam** (`NoAdSchedule`, kod 35) — trigger: daily — src: `NoAdSchedule.php`
- **Spadek wyniku optymalizacji konta** (`AccountOptimizationScoreDecrease`, kod 41) — trigger: daily — src: `AccountOptimizationScoreDecrease.php`
- **Spadek wyniku optymalizacji kampanii** (`CampaignOptimizationScoreDecrease`, kod 42) — trigger: daily — src: `CampaignOptimizationScoreDecrease.php`
- **Optymalizacja harmonogramu tygodniowego** (`WeeklyScheduleOptimization`, kod 45) — najgorszy dzień tygodnia — trigger: weekly — src: `WeeklyScheduleOptimization.php`
- **Wysoki współczynnik odrzuceń** (`HighRejectionRate`, kod 1ga, GA4) — trigger: weekly — src: `HighRejectionRate.php`
- **Spadek średniego czasu sesji** (`AverageSessionDurationDecrease`, kod 2ga, GA4) — trigger: weekly — src: `AverageSessionDurationDecrease.php`
- **Brak marki w tytule produktu** (`NoBrandInTitle`, kod 1sembot) — jakość feedu — trigger: wewnętrzny — src: `NoBrandInTitle.php`
- **[AI] Brak parametru "brand"** (`MissingBrandParameterAi`, kod 4sembot) — jakość feedu, wersja AI — trigger: wewnętrzny — src: `MissingBrandParameterAi.php`

## Obszar: Okazje

- **Wyczerpany budżet** (`BudgetExhausted`, kod 1) — dzienny budżet wyczerpany przed końcem dnia, czyli kampania traci potencjalny ruch — trigger: hourly — src: `BudgetExhausted.php` *(dwuznaczne — brak elementu "rentowności", patrz zastrzeżenie w D2/rozmowa o TTD-first; można też widzieć jako Bezpieczeństwo)*
- **Spadek udziału w wyświetleniach** (`SearchImpressionShareDecrease`, kod 9) — utracony potencjał do odzyskania — trigger: daily — src: `SearchImpressionShareDecrease.php`
- **Wzrost popularności fraz** (`ChangingTrends`, kod 28) — trigger: weekly — src: `ChangingTrends.php`

## Taski poza tymi 4 obszarami (jeśli są)

- **Brak zmian w historii konta** (`NoChangesInAccountHistory`, kod 53) — sygnał "konto jest zaniedbane", nie pasuje jednoznacznie do żadnego z 4 obszarów — bliżej meta-monitoringu pracy niż konkretnego ryzyka/oszczędności/okazji — src: `NoChangesInAccountHistory.php`

---

## Podsumowanie liczbowe
- 45 aktywnych triggerów (38 Google Ads + 2 GMC + 2 GA4 + 3 Sembot).
- 3 triggery `[PLANNED]` (zakomentowane w kodzie, nie działają dziś): `MissingTags`, `NegativeKeywords`, `ConvertingProductNotAvailable`.
- Kilka triggerów ma niejednoznaczną kategorię — wymaga decyzji zespołu marketingowego, nie da się rozstrzygnąć z samego kodu.

## ⚠️ Pytania otwarte
1. Nowy system Procesów (~90, n8n-runner) — czeka na listę od usera. Ten dokument będzie wymagał scalenia/aktualizacji po jej otrzymaniu.
2. Które z dwuznacznych triggerów (`BudgetExhausted`, `NewCompetitor`, `CampaignCostChanges`) mają trafić do której kategorii — decyzja narracyjna, nie techniczna.
3. Czy triggery `[PLANNED]` w ogóle powinny się pojawić na stronie (jako "już wkrótce"), czy pominąć do czasu wdrożenia.
