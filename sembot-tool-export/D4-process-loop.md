# D4 — Pętla: wykrycie → task → dry-run → approval → wykonanie → weryfikacja

Data: 2026-08-22 | Źródło: `sembot-laravel` (patrz też D2-task-anatomy.md dla pełnego rozbicia mechanizmów)

## ⚠️ Kluczowe zastrzeżenie

**Nie ma jednej pętli.** W kodzie są dwie różne, częściowo niezależne ścieżki. Strona nie powinna przedstawiać ich jako jednego, jednolitego przepływu — poniżej opisane osobno, z jasnym zaznaczeniem gdzie się różnią.

---

## Ścieżka A — klasyczny Task (bez dry-run)

```
1. Wykrycie      — trigger uruchamia się wg harmonogramu (np. hourly/daily/weekly),
                    odpytuje Google Ads/GMC/GA4 — src: app/Services/Projects/Triggers/*.php,
                    scheduler: app/Console/Commands (task-triggers:run)
2. Utworzenie taska — Domain\Task\Commands\CreateTaskHandler, task dostaje
                    title/context/display_context/action_type — src: Domain/Task/Commands/CreateTaskHandler.php
3. Dry-run/preview — BRAK na tym poziomie. Task od razu pokazuje sugerowaną akcję
                    z domyślnymi parametrami (np. "zwiększ budżet o 5%") —
                    src: app/Services/Actions/ActionTypes.php (defaults per action)
4. Approval      — user widzi listę dostępnych akcji (GET /actions/{taskId}) i wybiera —
                    to jest jedyny "approval": kliknięcie w UI, bez podglądu skutków —
                    src: app/Http/Controllers/Api/ActionController.php::actionsList()
5. Wykonanie     — POST /actions/ wykonuje NATYCHMIAST i SYNCHRONICZNIE, wywołuje
                    realne API (np. Google Ads), task automatycznie → FINISHED —
                    src: ActionController.php::process() → ActionService::process()
6. Weryfikacja/feedback — BRAK zautomatyzowanego mechanizmu weryfikacji efektu na tym
                    poziomie (żadna z 47 klas triggerów nie sprawdza "czy poprzednia
                    akcja pomogła") — jeśli problem wróci, powstanie nowy, osobny task
```

**Dla treści strony:** ta ścieżka NIE ma dry-run ani formalnego "zatwierdzenia ze zrozumieniem skutków" — user klika przycisk akcji i zmiana dzieje się od razu. To ważne ograniczenie względem obietnicy "approval przed zmianą konta" z brief'u marketingowego.

---

## Ścieżka B — mutacje przez czat AI / MCP (pełny dry-run + approval)

```
1. Żądanie zmiany — user rozmawia z AI chat (Marketer AI) lub AI proponuje zmianę
                    w kontekście taska — src: app/Services/Ai/Chat/ChatFunctions/
2. Dry-run       — ConnectionExecMutateController oblicza blast_radius (zasięg zmiany)
                    zamiast wykonywać ją od razu — src: ConnectionExecMutateController.php
3. Confirmation token — token z TTL 5 minut, user/AI musi go użyć żeby faktycznie
                    wykonać zmianę w tym oknie czasowym — src: ConnectionExecMutateController.php
4. Approval      — McpWorkflowApproval: status pending → approved/rejected/expired —
                    src: app/Models/Workflows/McpWorkflowApproval.php
5. Wykonanie     — dopiero po approval, z użyciem confirmation_token —
                    src: ConnectionExecMutateController.php
6. Weryfikacja   — MutationAudit zapisuje snapshot_before/after, można zrobić rollback —
                    src: app/Models/Mcp/MutationAudit.php (pola: was_dry_run, blast_radius,
                    rolled_back_at)
```

**Dla treści strony:** to jest ścieżka, która faktycznie realizuje obietnicę "dry-run→approval→audit→rollback" z brief'u marketingowego — ale działa dla mutacji inicjowanych przez AI chat/MCP, nie jest to coś co dzieje się automatycznie dla każdego taska ze Ścieżki A.

---

## Ścieżka C — nowy system Process (n8n) — mechanika runów, nie pojedynczej akcji

```
1. Wykrycie     — proces uruchamia się wg harmonogramu w n8n-runnerze, warunek
                  (metric/threshold) zdefiniowany w katalogu runnera (poza tym repo)
2. Run          — POST .../runs/{runUuid}/task (kontrakt z n8n) tworzy/aktualizuje
                  task-rodzica — src: app/Http/Controllers/Api/Projects/ProcessRunController.php
3. Dedup+cooldown — jeśli podobny task zamknięty niedawno (w oknie cooldown_days),
                  nowy task NIE powstaje — src: ProcessTaskService::recentlyClosedTask()
4. Artefakty    — dane runu (alert/kpi/chart/table) zapisywane w tasks.data.artifacts,
                  NADPISYWANE przy każdym kolejnym runie — src: ProcessTaskService::applyData()
5. Finalizacja  — runner woła finalizeRun (status, condition_met, flagged_count,
                  metrics_json, impact_json) — src: ProcessRunController.php::finalizeRun()
6. Powiadomienie — ProcessRunNotifier::runFinalized() — src: app/Notifications/Processes/
7. Wykonanie/approval — NIE jest częścią tego kontraktu; jeśli user chce wykonać
                  akcję, przechodzi do Ścieżki A (klasyczne akcje na tasku) lub
                  Ścieżki B (przez AI chat)
```

**Dla treści strony:** to jest mechanika najbliższa "regularnemu procesowi z pamięcią" (dedup, cooldown, licznik `persistence` — ile runów problem się utrzymuje) — ale wykonanie akcji po wykryciu problemu i tak przechodzi przez Ścieżkę A albo B, nie ma tu własnego kroku approval.

---

## ⚠️ Pytania otwarte
1. Czy strona ma pokazywać uproszczoną, jedną pętlę (upraszczając te 3 ścieżki), czy uczciwie rozdzielić "wykrycie i task" (silne, działające) od "dry-run i approval" (silne, ale osobne, tylko dla mutacji AI/MCP)?
2. Czy i kiedy planowane jest wpięcie dry-run/approval bezpośrednio w wykonanie akcji ze Ścieżki A — brak potwierdzenia w kodzie, do wyjaśnienia z zespołem produktowym.
