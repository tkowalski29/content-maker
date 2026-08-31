# D8 — Glosariusz nazewnictwa produktowego

Data: 2026-08-22 | Źródło: `sembot-laravel` (`resources/lang/{pl,en}/*.php`), `sembot-angular` (`src/assets/i18n/{pl,en}.json`)

Cel: żeby copy strony mówiło językiem produktu, nie wymyślonym.

| Pojęcie z brief'u marketingowego | Jak nazywa to produkt (PL) | Jak nazywa to produkt (EN) | Uwaga |
|---|---|---|---|
| Task | **Zadanie** / **Zadania** (menu) | Task / Tasks | Model: `App\Models\Projects\Task` |
| Proces (stary system, 47 triggerów) | **Wyzwalacz** (czasem niespójnie zostawione jako "Trigger") | Trigger | Model: `TaskTrigger`. Niespójność w kodzie: część miejsc w UI używa spolszczonego "Wyzwalacz", część zostawia "Trigger" po angielsku — do ujednolicenia zanim trafi na stronę |
| Proces (nowy system, n8n) | **Proces** / **Procesy** | Process / Processes | Model: `App\Models\Projects\Process` — docblock wprost: "Nowy system 'Procesy', równoległy do TaskTrigger" |
| Marketer AI | **Marketer AI** (nazwa własna, nie tłumaczona) | Marketer AI | Realna nazwa produktowa, używana konsekwentnie w kodzie/UI/notyfikacjach, nie tylko marketingowo. W menu nawigacyjnym bywa też ogólnie "Asystent AI" |
| Dry-run | **dry-run** (zapożyczone, nietłumaczone) | dry-run | Występuje dosłownie w `resources/lang/{pl,en}/auth.php` i `workflows.php`. W warstwie promptów czatu bywa też nazywane **"PREVIEW"** + słowo-klucz **"POTWIERDZAM"** do wykonania — brak jednej ujednoliconej nazwy user-facing |
| Blast radius | *(brak w tłumaczeniach PL/EN)* | blast_radius | Pojęcie API/deweloperskie (pole w response JSON), nie ma dedykowanego labela UI |
| Approval | **zatwierdzenie** | approval / Approve / Reject | `McpWorkflowApproval` — notyfikacja "Wymagane zatwierdzenie: :summary", przyciski Approve/Reject w UI (dokładne PL tłumaczenie przycisków nie zweryfikowane) |
| Konto (z brief'u marketingowego) | **brak jako oficjalny termin produktowy** — patrz niżej | — | Nie używać "konto" jako terminu 1:1 z produktem — patrz rozbicie workspace/projekt poniżej |
| — | **Przestrzeń robocza** (workspace) | Workspace | Nadrzędna jednostka rozliczeniowa/organizacyjna. `resources/lang/pl/workspace.php` |
| — | **Projekt** | Project | Podrzędna jednostka wewnątrz workspace'u — zwykle 1 sklep/domena. Sluga unikalne per workspace, nie globalnie |
| Obszar wartości (Bezpieczeństwo/Oszczędności/Plan poprawy/Okazje) | **brak w kodzie** | — | Nie istnieje jako enum/etykieta/grupowanie w produkcie — to autorska rama contentowa zespołu marketingowego. Można używać, ale nie przedstawiać jako "nazwa z produktu" |

## ⚠️ Ważne rozróżnienie: "konto" vs "workspace" vs "projekt"

W kodzie kolumna `account_id` jest **niejednoznaczna** — na modelu `Task` jednocześnie ma dwie różne relacje: raz czytana jako `workspace()`, raz jako `account()` → `User`. To legacy nazewnictwo bazodanowe, nie oficjalny termin produktowy. **Rekomendacja: na stronie używać "przestrzeń robocza" (workspace) i "projekt", unikać słowa "konto"** jako terminu produktowego — w produkcie samym jest ono źródłem niejednoznaczności, nie warto przenosić tego na stronę.

## ⚠️ Pytania otwarte
1. Niespójność "Wyzwalacz" vs "Trigger" w UI — do ujednolicenia z zespołem produktowym przed użyciem w treści.
2. Dokładne polskie tłumaczenie przycisków "Approve"/"Reject" nie zweryfikowane (nie sprawdzone w plikach frontendowych poza tym co znaleziono).
3. "PREVIEW"/"POTWIERDZAM" jako alternatywne nazwy dla konceptu dry-run w warstwie promptów czatu — czy to ma iść na stronę, czy trzymać się wyłącznie "dry-run".
