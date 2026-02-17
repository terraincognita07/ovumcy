# TASK: Fix Delete Confirmation Dialog

**Проблема:** Кнопка Delete удаляет данные БЕЗ подтверждения - критическая UX проблема!

**Severity:** HIGH - пользователь может случайно потерять данные
**Статус:** Completed (2026-02-17)

---

## Что нужно исправить:

### Файл: `web/templates/day_editor_partial.html`

**Найди кнопку Delete (примерно line 61):**

```html
<button type="button"
        hx-delete="/api/days/..."
        class="...">
    <span x-text="translations.delete"></span>
</button>
ДОБАВЬ атрибут hx-confirm:

xml
<button type="button"
        hx-delete="/api/days/{{.Date}}"
        hx-confirm="{{ if eq .Lang "ru" }}Удалить запись за этот день? Это действие нельзя отменить.{{ else }}Delete this day's entry? This cannot be undone.{{ end }}"
        hx-target="#day-editor"
        hx-swap="innerHTML"
        class="px-4 py-2 bg-red-50 text-red-600 rounded-lg hover:bg-red-100 transition-colors border border-red-200">
    <span x-text="translations.delete"></span>
</button>
Или через Alpine.js (более чистый вариант):

xml
<button type="button"
        @click="if(confirm(translations.confirmDelete)) { 
            fetch('/api/days/{{.Date}}', {method: 'DELETE'}).then(() => {
                htmx.trigger('#calendar-grid', 'refresh');
                document.getElementById('day-editor').innerHTML = '<p>Day deleted</p>';
            })
        }"
        class="px-4 py-2 bg-red-50 text-red-600 rounded-lg hover:bg-red-100 transition-colors border border-red-200">
    <span x-text="translations.delete"></span>
</button>
i18n ключи (ПРОВЕРЬ что они есть):
internal/i18n/en.json (line 99+):
json
"confirmDelete": "Delete this day's entry? This cannot be undone."
internal/i18n/ru.json (line 99+):
json
"confirmDelete": "Удалить запись за этот день? Это действие нельзя отменить."
Проверка:
Запусти go run cmd/lume/main.go

Открой календарь, выбери день с данными

Нажми "Удалить"

Должно появиться окно подтверждения с текстом на правильном языке

Отмена → ничего не происходит

Подтверждение → день удаляется

Коммит:
bash
git add .
git commit -m "fix: Add confirmation dialog for delete day entry

- Add hx-confirm or Alpine.js confirm() before delete
- Prevent accidental data loss
- i18n support for confirmation text (RU/EN)"
Severity: HIGH → FIXED ✅
После этого фича будет полностью безопасной для пользователя.
