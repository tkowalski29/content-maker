# Status eksportu contentu Sembot → sembot.com

Ostatnia aktualizacja: 2026-08-21

## Gotowe
- **D1-task-taxonomy.md** — 45 triggerów starego systemu zmapowane na 4 obszary wartości (interpretacja, nie fakt z kodu — zaznaczone).
- **D2-task-anatomy.md** — pełna anatomia obiektu `Task`, 3 mechanizmy (klasyczne triggery / system Process+n8n / dry-run+approval przez MCP), brak pola kwoty jako kolumny.
- **D3-task-examples.md** — 3 syntetyczne przykłady (Bezpieczeństwo/Oszczędności/Plan poprawy), oparte o realne triggery i schemat, dane zmyślone.
- **D4-process-loop.md** — 3 ścieżki (klasyczny Task bez dry-run / MCP z pełnym dry-run+approval / nowy system Process).
- **D5-process-catalog.md** — katalog 45 aktywnych + 3 wyłączonych triggerów z częstotliwością i produkowanym taskiem.
- **D6-integrations.md** — katalog integracji ze statusem IMPLEMENTED/PARTIAL i zakresem read/write.
- **D9-ui-surfaces.md** — sprawdzone i udokumentowane jako niewykonalne bez danych logowania do konta demo (brak w środowisku); nie zgadywane.

## Świadomie niepełne — czeka na dane od usera (user zdecydował 2026-08-22: pominąć na razie, dosłać później)
- D1 i D5 obejmują tylko **stary** system (47 triggerów PHP). Nowy system Procesów (~90, n8n-runner) jest poza zasięgiem tego repo — user dostarczy listę osobno, wtedy oba pliki wymagają scalenia/aktualizacji.

## Gotowe (dokończone)
- **D7-marketer-ai.md** — zakres kompetencji (z zastrzeżeniem: realne narzędzia dziś żyją w osobnym repo `sembot_public-mcp`, poza zasięgiem tego audytu), mechanizm mutacji (dry-run/token/audit + tryb auto-approve + McpWorkflowApproval), granice (brak system promptu w tym repo — access control zamiast tego).
- **D8-glossary.md** — task=Zadanie, trigger=Wyzwalacz(/Trigger niespójnie), nowy system=Proces, Marketer AI (nazwa realna), dry-run (zapożyczone + "PREVIEW"/"POTWIERDZAM" w promptach), approval=zatwierdzenie, workspace=przestrzeń robocza vs projekt (nie "konto"), 4 obszary wartości = brak w kodzie, autorska rama marketingowa.

## Wszystko z P0+P1+P2 zrobione (D1-D9)
Jedyny świadomy brak: pełny katalog nowego systemu Procesów (~90, n8n-runner) w D1/D5 — czeka na listę od usera, do scalenia gdy dojdzie. Wszystko inne kompletne i zapisane w `/workspace/sembot-tool-export/`.

## Następny krok
Gdy user potwierdzi że to wszystko, zrobić PDF (HTML → Playwright → publikacja na content-maker → link raw.githubusercontent.com) ze wszystkich plików D1-D9.

## Kluczowe ustalenia do pamiętania
1. Koncept "task" z briefu marketingowego (proces→dry-run→approval→wykonanie jako jedna rzecz) **nie istnieje jako jeden mechanizm** — patrz D2 dla pełnego rozbicia na 3 systemy.
2. 3 nazwane procesy TTD-first z briefu przełożonego (§6.6) **nie istnieją w kodzie** pod żadną nazwą — ani w starym systemie triggerów, ani w nowym systemie Procesów, ani w n8n workflow-templates widocznych w tym repo. Oznaczone jako `[PLANNED]` do czasu wyjaśnienia z przełożonym.
3. Automatyczny monitoring/taski dziś działają tylko na Google Ads + GMC + GA4 — Microsoft Ads/Meta/GSC są podłączalne, ale nie zasilają triggerów.
4. Finalny output ma trafić do usera jako PDF (HTML → Playwright → publikacja na `tkowalski29/content-maker` → link raw.githubusercontent.com) — dopiero na końcu, gdy wszystkie pliki D1-D9 (przynajmniej P0) będą gotowe.
