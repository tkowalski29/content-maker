# D9 — UI surfaces (opcjonalne, P2)

Data: 2026-08-22

## Status: nie zrealizowane w tym środowisku — uzasadnienie poniżej

To był deliverable P2 (opcjonalny) w brief'ie. Sprawdziłem wykonalność zamiast pominąć go milczeniem:

- Angular dev server **działa i odpowiada** (`http://localhost:4200/ibn4my9yhiyzpez7r1hydrz8jr/` → HTTP 200).
- Playwright/Chromium jest dostępny w środowisku (potwierdzone: `sembot-test-automation/node_modules/playwright`, MCP Playwright server per `.doc/testing.md`).
- **Brak** jednak udokumentowanych danych logowania do konta testowego/demo w tym devcontainerze (sprawdzone `.doc/*.md`, `.helpers/*.sh` — zero wzmianek o loginie/haśle/tokenie testowym).
- W bazie `mysql` jest jeden użytkownik wyglądający na testowy (`test@butsy.pl`, id 1) i kilka projektów testowych (`test.pl`, `test123123.pl` itd.), ale bez znanego hasła nie da się przejść przez UI-owy formularz logowania. Alternatywa (wygenerowanie tokenu Passport przez `php artisan tinker` i wstrzyknięcie go do `localStorage` Angulara) wymagałaby reverse-engineeringu dokładnego formatu tokenu/kluczy oczekiwanych przez frontend — nieudokumentowane, więc obarczone ryzykiem błędu i dodatkowym czasem bez pewności sukcesu.

## Rekomendacja

Jeśli zrzuty realnych ekranów (lista tasków, szczegóły taska, ewentualnie widok Procesów) są potrzebne do materiałów projektowych, najszybsza droga to:
1. Ktoś z zespołu (z realnym dostępem/loginem) robi 3-5 zrzutów ręcznie, **lub**
2. Przekazanie mi (albo innej sesji) danych logowania do konta demo — wtedy dokończę to w 10-15 minut przez Playwright.

Nie zgaduję wyglądu UI ani nie tworzę fikcyjnych mockupów podszywających się pod realny produkt — to byłoby sprzeczne z zasadą "nie zgaduj" z brief'u.
