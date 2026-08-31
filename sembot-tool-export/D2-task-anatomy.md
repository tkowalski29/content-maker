# D2 — Anatomia obiektu task

Data: 2026-08-21 | Źródło: `sembot-laravel` (`app/Models/Projects/Task.php`, `Domain/Task/`, `app/Services/Projects/*`, `app/Http/Controllers/Api/Connections/ConnectionExecMutateController.php`, `app/Http/Controllers/Api/ActionController.php`, `app/Services/Projects/ProcessTaskService.php`)

## ⚠️ Zastrzeżenie wstępne — przeczytać przed użyciem w treści strony

W kodzie **nie istnieje jeden spójny obiekt "task"** z wbudowanym dry-run + kwotą impact + approval, tak jak sugeruje narracja "proces→task→dry-run→approval→wykonanie→weryfikacja". Są **trzy równoległe mechanizmy**, które razem odpowiadają za to, co strona chce pokazać jako "kartę taska":

| Mechanizm | Status | Co robi | Czego brakuje względem narracji |
|---|---|---|---|
| **Model `Task` + klasyczne "task triggery"** (47 detektorów PHP) | `[IMPLEMENTED]` | monitoring → task z rekomendowaną akcją → user sam wykonuje | brak wbudowanego dry-run/preview, brak pola kwoty/impact jako kolumny |
| **System `Process`/`ProcessRun`/n8n-runner** (nowszy, migracje lipiec 2026) | `[IMPLEMENTED]` jako infrastruktura | proces → task z `data.artifacts` (alert/kpi/chart/table) w drawerze, dedup, cooldown, powiadomienia | katalog konkretnych ~90 procesów żyje w n8n-runnerze poza tym repo — patrz [[D1]] (wstrzymane) |
| **`ConnectionExecMutateController` + `MutationAudit` + `McpWorkflowApproval`** | `[IMPLEMENTED]` | pełny dry-run→blast_radius→confirmation_token(TTL 5 min)→execute→audit z rollback | działa dla mutacji z czatu AI/MCP, **nie jest wpięty** bezpośrednio w model `Task` ani `Process` |

Rekomendacja dla contentu: nie przedstawiać "dry-run + approval" jako czegoś co dzieje się wewnątrz każdego taska. To, co strona pokazuje jako jedną kartę, w produkcie jest kompozycją tych mechanizmów — user może dostać task (z pierwszych dwóch systemów) i **osobno**, w rozmowie z Marketer AI, uruchomić prawdziwy dry-run/approval (trzeci mechanizm) dla wykonania zmiany.

---

## 1. Model `Task` (`app/Models/Projects/Task.php`, tabela `tasks`)

| Pole | Typ | Opis | Zawsze obecne? | Przykładowa wartość (zmyślona) |
|---|---|---|---|---|
| `id` | bigint | identyfikator | tak | 90210 |
| `parent_id` | bigint, nullable | task-rodzic (subtaski procesu) | nie | null / 90200 |
| `project_id` | bigint | konto/projekt klienta | tak | 4821 |
| `account_id` | bigint | użytkownik-właściciel | tak | 17 |
| `task_trigger_id` | bigint, nullable | link do triggera klasycznego systemu | nie (tylko `type=process` starego systemu) | 1024 |
| `process_id` | bigint, nullable | link do `Process` (nowy system) — pole ustawiane przez `ProcessTaskService::linkToProcess()`, nie widoczne wprost w `DESCRIBE tasks` bo bywa dodawane osobną migracją | nie | 55 |
| `section_id` | bigint, nullable | sekcja/kolumna w UI zadań | nie | 3 |
| `title` | text | tytuł po ludzku | tak | "Kampania Search PL wyczerpała dzienny budżet" |
| `description` | longtext | opis | nie | — |
| `context` | JSON (`array` cast) | surowe dane robocze — patrz §3 | nie (per-trigger) | `{"campaign_id": 123, "campaign_cost_diff": 850.0}` |
| `display_context` | JSON (`json` cast, z accessorem tłumaczącym) | mapowanie kluczy `context` → etykiety PL/EN — patrz §3 | nie | `{"campaign_cost_diff": "Różnica kosztu kampanii"}` |
| `data` | JSON | **artifact payload** nowego systemu Procesów — patrz §3 | tylko dla `type=process` z nowego systemu | `{"artifacts": {"1": {"type": "alert", ...}}}` |
| `metrics_run_id` | bigint, nullable, FK → `workflow_runs` | osobny kanał KPI dla data-driven workflows | nie | 771 |
| `author_id` | bigint, nullable | kto utworzył (jeśli manualny) | nie | null |
| `type` | enum string | patrz §2 | tak | `process` |
| `status` | enum string | patrz §2 | tak | `new` |
| `priority` | enum (`very_low`..`very_high`) | tak | tak | `high` |
| `action_type` | varchar, nullable | klucz z `ActionTypes` mapujący na dostępne akcje — patrz §4 | nie | `EXHAUSTED_BUDGET` |
| `due_date`, `started_at`, `finished_at` | timestamp, nullable | daty cyklu życia | nie | — |
| `interval` | longtext, nullable | harmonogram powtarzania (dla tasków cyklicznych) | nie | — |
| `order`, `subtask_order`, `is_pinned` | int/bool | UI: kolejność, przypięcie | tak/nie | — |
| `deleted_at`, `deleted_by` | soft-delete | tak (soft delete) | nie | — |

**Brak pola kwoty/impact jako kolumny SQL.** Sprawdzone w całej historii 27 migracji tabeli `tasks` — nie ma `amount`/`impact`/`value`/`cost`/`savings`. To, co strona nazywa "kwotą problemu", zawsze żyje w jednym z JSON-ów (§3), pod kluczem który **każdy trigger/proces wybiera sam** — nie ma jednego gwarantowanego pola.

## 2. Enumy typu i statusu

**`TaskType`** (`app/Types/Projects/TaskType.php:11-17`) — 7 wartości:
```
workflow | agent | summary | process | manual | ai_proposed | template
```
Grupowania: `standardTypes()` = wszystkie oprócz `template`; `templateTypes()` = tylko `template`.

**`TaskStatus`** (`app/Types/Projects/TaskStatus.php`) — 6 wartości:
```
info | new | in_progress | ignored | finished | canceled
```
Grupowania: `unresolved()`, `resolved()`, `terminal()`.

**Cykl życia (state machine, ze zdarzeń w kodzie):**
```
NEW → IN_PROGRESS → FINISHED
              ↘  IGNORED / CANCELED
FINISHED/IGNORED → (REOPEN akcja) → NEW/IN_PROGRESS
```
`REOPEN` jako dostępna akcja pojawia się tylko gdy task jest już `FINISHED`/`IGNORED` (`ActionTypes.php:748-758`). `MARK_COMPLETED` i `IGNORE` są dostępne na każdym tasku niezależnie od jego `action_type` (`ActionTypes.php:189-205`).

**`TaskPriority`**: `very_low | low | medium | high | very_high`.

## 3. Trzy JSON-y — role i różnice

| Pole | Rola | Kto wypełnia | Uwaga |
|---|---|---|---|
| `context` | surowe dane robocze — używane RÓWNOCZEŚNIE do wyświetlenia i do wykonania akcji (np. `IncreaseBudget` czyta stąd `campaign_id`) | `Domain\Task\Commands\CreateTaskHandler`/`UpdateTaskHandler`, per-trigger (np. `CampaignCostChanges.php:39-61`) | freeform, brak walidowanego schematu |
| `display_context` | mapa "który klucz z `context` pokazać i pod jaką etykietą" — surowa wartość to nazwa klucza tłumaczenia (`resources/lang/pl/tasks.php:131-184`), accessor w modelu tłumaczy ją na tekst | jw. | to mechanizm renderowania, nie sam artefakt |
| `data.artifacts` | **właściwy "artifact payload"** z narracji marketingowej — bloki `alert`/`kpi`/`chart`/`table` renderowane w drawerze taska przez `proxy_sembot` | `ProcessTaskService::applyData()`, nadpisywane w całości przy KAŻDYM runie procesu | **struktura NIE jest walidowana w Laravelu** (komentarz wprost w `Task.php`) — kształt zależy od konkretnego workflow node w n8n |
| `workflow_runs.metrics_json` (przez `metrics_run_id`) | osobny kanał KPI ("Task drawer KPI tiles") dla tasków powiązanych z data-driven workflows | n8n-runner | równoległy, nie scalony z `data.artifacts` |

Przykład realnego kształtu `data.artifacts` (z testów `ProcessInternalEndpointsTest.php`, kształt kontraktu, nie realny biznesowy przypadek):
```json
{
  "artifacts": {
    "1": {"type": "alert", "condition_met": true, "flagged_count": 3, "metric": "ctr"},
    "2": {"type": "kpi", "impact": {"clicks": -120}}
  }
}
```

## 4. Wykonanie akcji — brak dry-run na poziomie Task

Ścieżka UI: `GET /actions/{taskId}` (lista dostępnych akcji, `ActionController::actionsList()`) → user klika → `POST /actions/` (`ActionController::process()`) → **natychmiastowe, synchroniczne** wykonanie (`ActionService::process()` woła `execute()` od razu; dispatch na kolejkę jest zakomentowany w kodzie). Realny call idzie od razu do zewnętrznego API (np. `IncreaseBudget::execute()` → `GoogleAdsCampaignBudgetService`), i task jest od razu oznaczany `FINISHED`.

**Brak jakiegokolwiek podglądu/potwierdzenia na tym poziomie.** To przeciwieństwo mechanizmu MCP opisanego niżej. Istnieje endpoint `/actions/async`, ale nie jest to ścieżka wywoływana przez główny flow.

Pełna lista 22 akcji (`ActionNames.php:11-32`): `exclude_keyword, pause_keyword, increase_budget, decrease_budget, reduce_keyword_cost, increase_keyword_cost, reduce_campaign_keywords_cost, increase_campaign_keywords_cost, reject_recommendations, pause_ad, start_ad, mark_completed, ignore, reopen, reduce_mobile_bid, increase_mobile_bid, reduce_device_bid, increase_device_bid, increase_ad_group_keywords_cost, update_keyword_match_type, auto_tagging_modifier, dismiss_all_recommendations`.

## 5. Prawdziwy dry-run + approval — ale gdzie indziej

`ConnectionExecMutateController` (`app/Http/Controllers/Api/Connections/ConnectionExecMutateController.php`) + `MutationAudit` (`app/Models/Mcp/MutationAudit.php`) + `McpWorkflowApproval` (`app/Models/Workflows/McpWorkflowApproval.php`):

```
dry_run (oblicza blast_radius) → confirmation_token (TTL 5 min) → execute → MutationAudit (snapshot_before/after, rollback możliwy)
```
`McpWorkflowApproval` ma status `pending → approved/rejected/expired`. To jest **jedyne miejsce w kodzie**, gdzie faktycznie istnieje mechanizm "pokaż co się zmieni, zanim to zrobisz" z audytem i rollbackiem — ale działa dla mutacji inicjowanych przez czat AI / MCP, nie jest automatycznie wpięty w klasyczny `Task` ani w subtaski `Process`.

## 6. System `Process`/`ProcessRun` (nowszy mechanizm)

- `Process` (`app/Models/Projects/Process.php`, tabela `processes`) — **instancja** procesu per projekt: `code`, `template_slug`, `params`, `task_template`, `run_schedule`. Definicja treści/warunku (`meta.process`: nazwa, opis, metric, threshold, `artifact_blocks`) żyje w n8n-runnerze, katalogowana przez `ProcessCatalog::fetchFromRunner()` — **poza zasięgiem tego repo**.
- `ProcessRun` (`app/Models/Projects/ProcessRun.php`, tabela `process_runs`) — jeden run procesu, `belongsTo(Task)`.
- `ProcessTaskService::upsertTask()` — szuka jednego otwartego task-rodzica per proces; jeśli brak, tworzy (`type=process`); subtaski mergowane po `(search_column, search_value)` w `context`.
- **Dedup + cooldown**: `dedupe_key` + `cooldown_days` niesione przez runner w payloadzie — jeśli task zamknięty niedawno w oknie cooldown, nowy task nie powstaje.
- **`persistence`**: licznik ile runów z rzędu problem się utrzymuje — dopisywany przy każdym runie, wg komentarza w kodzie "wejście pod przyszły severity/priority_score".
- Endpointy są jawnie oznaczone w kodzie jako KONTRAKT z n8n-runnerem: "nie zmieniać ścieżek/pól bez uzgodnienia z runnerem" (`ProcessRunController.php:16-25`).

---

## ⚠️ Pytania otwarte

1. Katalog konkretnych ~90 procesów (nazwy, warunki, artifact_blocks) żyje w n8n-runnerze, poza zasięgiem tego repo — potrzebny eksport/dostęp stamtąd, żeby D1/D5 mogły go opisać (obecnie wstrzymane, czeka na dane od usera).
2. Czy `process_id` na `Task` to realna kolumna, czy relacja liczona inaczej (nie znalazłem jej w `DESCRIBE tasks` bezpośrednio zapytanym do bazy — może być dodana późniejszą migracją nieobjętą tym zapytaniem; do potwierdzenia przy pisaniu finalnej wersji, jeśli ten szczegół ma trafić 1:1 na stronę).
3. Nie ma potwierdzenia, czy/kiedy mechanizm dry-run+approval (§5) zostanie wpięty bezpośrednio w `Task`/`Process` — obecnie to osobna ścieżka tylko dla czatu AI/MCP. Strona nie powinna sugerować, że KAŻDY task ma wbudowany dry-run, dopóki to się nie zmieni w kodzie.
