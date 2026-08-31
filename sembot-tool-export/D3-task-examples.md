# D3 — Trzy przykłady tasków

Data: 2026-08-22 | Status danych: **wszystkie 3 przykłady są syntetyczne** — w tym środowisku baza `mysql` nie zawiera żadnych realnych, automatycznie wygenerowanych tasków (`task_triggers`: 0 rekordów; `tasks`: 28 rekordów, wszystkie ręczne). Przykłady zbudowane na bazie realnego schematu (D2) i realnych, aktywnych triggerów (D1/D5) — liczby i nazwa konta są zmyślone.

**⚠️ Każdy z poniższych przykładów jest oznaczony jako "Przykład demonstracyjny" i NIE pochodzi z konta realnego klienta.**

---

## Przykład 1 — Bezpieczeństwo: odrzucone reklamy

*Przykład demonstracyjny — dane zmyślone.*

Trigger źródłowy: `AdDisapproved` (kod 7, `app/Services/Projects/Triggers/AdDisapproved.php`), priorytet domyślny: high, harmonogram: co godzinę.

```json
{
  "title": "3 reklamy zostały odrzucone przez Google w kampanii \"Buty sportowe — Search PL\"",
  "type": "process",
  "status": "new",
  "priority": "high",
  "action_type": "AD_DISAPPROVED",
  "context": {
    "campaign_id": 118822001,
    "campaign_name": "Buty sportowe — Search PL",
    "disapproved_ads_count": 3,
    "policy_topic": "MISLEADING_PRICING"
  },
  "display_context": {
    "disapproved_ads_count": "Liczba odrzuconych reklam",
    "policy_topic": "Powód odrzucenia (polityka Google)"
  },
  "data": {
    "artifacts": {
      "1": {
        "type": "table",
        "columns": ["Reklama", "Grupa reklam", "Powód"],
        "rows": [
          ["Buty do biegania -30%", "Buty do biegania", "Wprowadzająca w błąd cena"],
          ["Promocja tenisówek", "Tenisówki", "Wprowadzająca w błąd cena"],
          ["Wyprzedaż butów zimowych", "Buty zimowe", "Wprowadzająca w błąd cena"]
        ]
      }
    }
  },
  "dostępne_akcje": ["reject_recommendations (nie dotyczy)", "mark_completed", "ignore"]
}
```

**Uwaga do treści:** dla tego typu problemu w kodzie **nie ma** dedykowanej akcji naprawczej (np. "popraw i wyślij ponownie") — dostępne są tylko ogólne akcje `mark_completed`/`ignore`. Naprawa wymaga ręcznej edycji reklamy poza Sembotem. Strona nie powinna sugerować, że Sembot naprawia to automatycznie.

---

## Przykład 2 — Oszczędności: wzrost kosztu konwersji

*Przykład demonstracyjny — dane zmyślone.*

Trigger źródłowy: `ConversionCostsChanges` (kod 22, `app/Services/Projects/Triggers/ConversionCostsChanges.php`), priorytet domyślny: high, harmonogram: tygodniowo.

```json
{
  "title": "Koszt konwersji w kampanii \"Kosmetyki — Shopping\" wzrósł o 42% w ciągu 14 dni",
  "type": "process",
  "status": "new",
  "priority": "high",
  "action_type": "CONVERSION_COSTS_INCREASE",
  "context": {
    "campaign_id": 118822045,
    "campaign_name": "Kosmetyki — Shopping",
    "previous_cost_per_conversion": 18.40,
    "current_cost_per_conversion": 26.15,
    "percent_change": 42.1
  },
  "display_context": {
    "previous_cost_per_conversion": "Koszt konwersji za poprzedni okres",
    "current_cost_per_conversion": "Koszt konwersji za obecny okres"
  },
  "data": {
    "artifacts": {
      "1": {
        "type": "kpi",
        "label": "Wzrost kosztu konwersji",
        "value": "+42%",
        "trend": "up_bad"
      },
      "2": {
        "type": "chart",
        "chart_type": "line",
        "metric": "cost_per_conversion",
        "window_days": 14
      }
    }
  },
  "dostępne_akcje": ["decrease_budget", "reduce_campaign_keywords_cost", "mark_completed", "ignore"]
}
```

**Uwaga do treści:** tu jest realna akcja naprawcza dostępna od razu na tasku (`decrease_budget`/`reduce_campaign_keywords_cost`) — ale jej wykonanie jest **natychmiastowe po kliknięciu**, bez podglądu/dry-run (patrz D4, Ścieżka A). Nie pokazywać tego jako "Sembot pokazuje dry-run przed zmianą" — to nieprawda dla tej ścieżki.

---

## Przykład 3 — Plan poprawy: spadek wyniku optymalizacji konta

*Przykład demonstracyjny — dane zmyślone.*

Trigger źródłowy: `AccountOptimizationScoreDecrease` (kod 41, `app/Services/Projects/Triggers/AccountOptimizationScoreDecrease.php`), priorytet domyślny: medium, harmonogram: codziennie.

```json
{
  "title": "Wynik optymalizacji konta spadł poniżej 70%",
  "type": "process",
  "status": "in_progress",
  "priority": "medium",
  "action_type": "ACCOUNT_OPTIMIZATION_SCORE_DECREASE",
  "context": {
    "account_id": 17,
    "current_score_percent": 64,
    "threshold_percent": 70
  },
  "display_context": {
    "current_score_percent": "Obecny wynik optymalizacji"
  },
  "data": {
    "artifacts": {
      "1": {
        "type": "alert",
        "condition_met": true,
        "metric": "optimization_score",
        "flagged_count": 1
      }
    }
  },
  "dostępne_akcje": ["dismiss_all_recommendations", "reject_recommendations", "mark_completed", "ignore", "reopen"]
}
```

**Uwaga do treści:** `status: in_progress` pokazuje, że user już zaczął się tym zajmować (nie jest to stan startowy `new`). `reopen` pojawia się wyłącznie gdy task jest `finished`/`ignored` — w tym przykładzie nie powinien być jeszcze dostępny; dodany tu tylko dla zilustrowania pełnej listy możliwych akcji w cyklu życia.

---

## ⚠️ Pytania otwarte
1. Czy do treści strony (hero/S6) potrzebne są realne, anonimizowane taski z prawdziwych kont (§14 brief przełożonego) — to osobny tor pracy, poza zasięgiem tej sesji (brak dostępu do produkcyjnych danych klientów z tego środowiska).
2. Pole `dostępne_akcje` w powyższych przykładach to moja adnotacja pomocnicza (nie jest to nazwa pola z bazy) — dodane dla jasności, do usunięcia jeśli przykłady mają iść 1:1 jako JSON do dev-a frontendowego.
