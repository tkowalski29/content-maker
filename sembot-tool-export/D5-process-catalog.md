# Katalog procesów pilnujących kont

Data: 2026-08-22 | Źródło: `sembot-laravel`, `app/Services/Projects/Triggers/*.php`, `data/trigger/*/codes.php` + `defaults.php`

## ⚠️ Zakres

Ten katalog opisuje **stary, klasyczny system triggerów** (dziś jedyny w pełni widoczny z tego repo). Nowy system `Process` (n8n-runner, ok. 90 procesów) ma analogiczną rolę ("gotowy monitor sytuacji biznesowej" z briefu marketingowego), ale jego katalog dojdzie osobno od usera — ten dokument będzie wtedy wymagał uzupełnienia.

Wszystkie poniższe: `[IMPLEMENTED]`, źródło danych = Google Ads chyba że zaznaczono inaczej. Harmonogram = domyślny `run_schedule` (user może go zmienić per projekt).

| Proces | Co monitoruje | Źródło danych | Częstotliwość (domyślna) | Jaki task produkuje |
|---|---|---|---|---|
| Wyczerpany budżet | `cost >= daily budget` kampanii | Google Ads | co godzinę | task z priorytetem medium, kampania która wyczerpała budżet |
| Spadek CTR | CTR poniżej progu w oknie dni | Google Ads | tygodniowo | task ze spadkiem CTR per kampania/grupa |
| Spadek liczby konwersji | spadek konwersji >X% w 7 dni | Google Ads | codziennie | task z listą dotkniętych kampanii |
| Zmiany w śr. CPC | wzrost CPC powyżej progu | Google Ads | tygodniowo | task z kampaniami/grupami o wzroście CPC |
| Niska efektywność słów kluczowych | dużo kliknięć, mało konwersji | Google Ads | tygodniowo | task z listą słów kluczowych do przeglądu |
| Zakończenie kampanii | <X dni do końca kampanii | Google Ads | codziennie | task przypominający o kończącej się kampanii |
| Ograniczenia reklam (odrzucenia) | reklamy odrzucone przez Google | Google Ads | co godzinę, priorytet high | task z listą odrzuconych reklam |
| Skuteczność reklam | różnice konwersji/CTR w grupie reklam | Google Ads | tygodniowo | task z rekomendacją które reklamy poprawić |
| Spadek udziału w wyświetleniach | search impression share poniżej progu | Google Ads | codziennie | task z utraconym potencjałem |
| Nieaktualne linki | wykryte nieaktualne linki w reklamach | Google Ads | co godzinę, priorytet high | task z listą reklam do poprawy |
| Zmiana konkurencji | nowi konkurenci na głównych słowach kluczowych | Google Ads | tygodniowo | task informacyjny o nowej konkurencji |
| Zmiany w jakości reklam | spadek oceny jakości reklam | Google Ads | codziennie | task z listą reklam o niższej jakości |
| Brakujące rozszerzenia | reklamy bez pełnego zestawu rozszerzeń | Google Ads | tygodniowo | task z rekomendacją dodania rozszerzeń |
| Optymalizacja dla urządzeń | jedno urządzenie znacząco słabsze | Google Ads | tygodniowo | task z rekomendacją zmiany stawek per urządzenie |
| Problemy z geo-targetingiem `[nieukończony wg komentarza w kodzie]` | kliknięcia spoza wybranej lokalizacji | Google Ads | tygodniowo, priorytet high | task z listą nietargetowanych regionów |
| Optymalizacja harmonogramu godzinowego | najgorsza godzina dnia w ostatnich 24h | Google Ads | codziennie | task z rekomendacją harmonogramu |
| Brak aktywności w kampanii | spadek kliknięć >X% | Google Ads | codziennie, priorytet high | task alarmujący o martwej kampanii |
| Różnice mobile-desktop | różnica konwersji/ROAS/CPC | Google Ads | tygodniowo | task z rekomendacją zmiany stawek mobile |
| Wzrost kosztu konwersji | koszt konwersji wzrósł >X% | Google Ads | tygodniowo, priorytet high | task z alarmem kosztowym |
| Spadek jakości słów kluczowych | quality score poniżej progu | Google Ads | tygodniowo | task z listą słów do poprawy |
| Reklamy z nieaktualną frazą | fraza sugerująca przeterminowaną ofertę | Google Ads | tygodniowo | task z listą reklam do aktualizacji |
| Grupy reklam bez aktywnych reklam | puste grupy reklam | Google Ads | tygodniowo | task z listą pustych grup |
| Wzrost CPM | CPM wzrósł >X% | Google Ads | tygodniowo, priorytet high | task z alarmem kosztowym |
| Zmiany w kosztach reklam wideo | koszt wideo wzrósł >X% | Google Ads | tygodniowo | task z alarmem kosztowym |
| Limit wykluczających słów kluczowych | zbliżanie się do limitu | Google Ads | tygodniowo | task techniczny |
| Spadek wyniku optymalizacji konta/kampanii | Google Ads optimization score spadł | Google Ads | codziennie | task z rekomendacją poprawy |
| Optymalizacja harmonogramu tygodniowego | najgorszy dzień tygodnia | Google Ads | tygodniowo | task z rekomendacją harmonogramu |
| Brak konwersji z tagów reklam | tagi reklam nie rejestrują konwersji | Google Ads | codziennie | task alarmujący |
| Wykluczenie brandu | brand niewykluczony z kampanii non-brand | Google Ads | codziennie | task z rekomendacją dodania wykluczenia |
| Wyłączone auto-tagowanie | brak gclid | Google Ads | codziennie | task techniczny |
| Reklama kieruje do 404 | martwy URL w aktywnej reklamie | Google Ads | tygodniowo | task alarmujący |
| Reklama display w aplikacjach mobilnych | niechciane miejsce wyświetlania | Google Ads | codziennie | task z rekomendacją wykluczenia |
| Włączone automatyczne rekomendacje Google | Google może samo zmieniać konto | Google Ads | codziennie | task ostrzegawczy |
| Powiadomienia o produktach w GMC | problemy/błędy produktowe | Google Merchant Center | priorytet high | task z listą produktów z problemem |
| Powiadomienia o koncie w GMC | problemy/błędy konta | Google Merchant Center | priorytet high | task alarmujący |
| Wysoki współczynnik odrzuceń | bounce rate powyżej progu | Google Analytics 4 | tygodniowo | task z alarmem UX |
| Spadek średniego czasu sesji | spadek >X% | Google Analytics 4 | tygodniowo | task z alarmem UX |
| Brak marki w tytule produktu | tytuł produktu bez marki z atrybutu `brand` | feed / wewnętrzne (Sembot) | — | task z listą produktów do poprawy tytułu |
| [AI] Brak parametru "brand" | odwrotność powyższego, wykryte przez AI | feed / wewnętrzne (Sembot) | — | task z listą produktów |
| Błąd połączenia z usługą | błąd integracji w projekcie | wewnętrzne (Sembot) | — | task techniczny |
| Brak zmian w historii konta | brak jakiejkolwiek zmiany na koncie w oknie dni | wewnętrzne (Sembot) | codziennie | task sygnalizujący zaniedbane konto |

## Wyłączone / `[PLANNED]` (istnieją w kodzie, ale nie działają dziś)

| Proces | Co miałby monitorować | Status |
|---|---|---|
| Brakujące tagi śledzące | brak tagów konwersji na stronie docelowej | `[PLANNED]` — zakomentowany w `codes.php` |
| Słowa kluczowe zablokowane przez wykluczenia | słowa kluczowe blokowane przez własne negatywne słowa kluczowe | `[PLANNED]` — zakomentowany |
| Zmiany w dostępności produktów konwertujących | produkt z konwersjami zniknął z oferty | `[PLANNED]` — zakomentowany, komentarz w kodzie: "brak statusu produktu, dodać z GMC" |

## ⚠️ Pytania otwarte
1. Katalog ~90 procesów nowego systemu (n8n-runner) — czeka na dane od usera, ten dokument wymaga wtedy scalenia.
2. Harmonogramy w tabeli to wartości domyślne — user może je zmienić per projekt (`run_schedule` jest polem instancji, nie stałą).
