export const icons = {
  sun: '<svg viewBox="0 0 24 24" aria-hidden="true"><circle cx="12" cy="12" r="4"/><path d="M12 2v2m0 16v2m10-10h-2M4 12H2m17.07-7.07-1.41 1.41M6.34 17.66l-1.41 1.41m14.14 0-1.41-1.41M6.34 6.34 4.93 4.93"/></svg>',
  moon: '<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M20.2 15.2A8.5 8.5 0 0 1 8.8 3.8 8.5 8.5 0 1 0 20.2 15.2Z"/></svg>',
  edit: '<svg viewBox="0 0 24 24" aria-hidden="true"><path d="m15 5 4 4M4 20l4-.8L19 8a2.12 2.12 0 0 0-3-3L5 16l-1 4Z"/></svg>',
  trash: '<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M4 7h16m-10 4v6m4-6v6M6 7l1 13h10l1-13M9 7V4h6v3"/></svg>',
  check: '<svg viewBox="0 0 24 24" aria-hidden="true"><path d="m5 12 4 4L19 6"/></svg>',
  error: '<svg viewBox="0 0 24 24" aria-hidden="true"><circle cx="12" cy="12" r="9"/><path d="M12 8v4m0 4h.01"/></svg>',
  success: '<svg viewBox="0 0 24 24" aria-hidden="true"><circle cx="12" cy="12" r="9"/><path d="m8 12 2.5 2.5L16 9"/></svg>'
};

export const text = {
  ru: {
    title: "Задачи — Daymark", description: "Простое и удобное управление задачами.",
    filters: "Фильтры задач",
    all: "Все задачи", active: "Активные", completed: "Выполненные",
    heading: "Порядок в делах —<br><span>ясность в мыслях.</span>", subtitle: "Освободите время для главного.",
    progress: "Ваш прогресс", composerTitle: "Название задачи", composerDescription: "Описание задачи",
    newTask: "Что нужно сделать?", note: "Добавить заметку...", add: "Добавить",
    remainingOne: "Осталась {n} задача", remainingFew: "Осталось {n} задачи", remainingMany: "Осталось {n} задач",
    allCaughtUp: "Всё сделано", started: "Начнём с чистого листа", doneOf: "Выполнено {done} из {total}",
    empty: "Пока здесь пусто", emptyCopy: "Добавьте задачу выше и задайте направление своему дню.",
    emptyActive: "Отлично! Активных задач нет.", emptyActiveCopy: "Вы разобрались со всеми задачами.",
    emptyCompleted: "Выполненных задач пока нет.", emptyCompletedCopy: "Завершите задачу — она появится здесь.",
    footer: "Маленькие шаги — большие перемены.", editHint: "Дважды щёлкните по задаче, чтобы изменить её",
    loading: "Загрузка задач", markDone: "Отметить задачу выполненной", markActive: "Вернуть задачу в активные",
    edit: "Изменить задачу", remove: "Удалить задачу", save: "Сохранить", cancel: "Отмена",
    cancelEdit: "Отменить редактирование", invalidTitle: "У задачи должно быть название.",
    connection: "Не удалось связаться с сервером. Проверьте подключение.",
    requestFailed: "Не удалось выполнить запрос ({status}).", notFound: "Задача не найдена.",
    validation: "Проверьте правильность заполнения задачи.", taskAdded: "Задача добавлена",
    taskUpdated: "Задача обновлена", taskRemoved: "Задача удалена", languageLabel: "Switch to English",
    dueAt: "Срок: {date} {time}", dueDate: "Выбрать дату и время срока",
    taskDue: "Пора делать: {title}", taskDueMessage: "Срок задачи истек",
    themeLight: "Включить светлую тему", themeDark: "Включить тёмную тему", dateLocale: "ru-RU",
    completeAll: "Выполнить все активные задачи", allCompleted: "Все активные задачи выполнены",
    bulkPartial: "Выполнено {completed}; не удалось выполнить {failed}."
  },
  en: {
    title: "Tasks — Daymark", description: "A simple, focused place to manage your tasks.",
    filters: "Task filters",
    all: "All tasks", active: "In progress", completed: "Completed",
    heading: "A clear mind<br>a <span>clear list.</span>", subtitle: "Make room for what matters today.",
    progress: "Daily progress", composerTitle: "Task title", composerDescription: "Task details",
    newTask: "What needs your attention?", note: "Add a note...", add: "Add task",
    remainingOne: "{n} task left", remainingFew: "{n} tasks left", remainingMany: "{n} tasks left",
    allCaughtUp: "All caught up", started: "A fresh start", doneOf: "{done} of {total} completed",
    empty: "Nothing on your list. Yet.", emptyCopy: "Add a task above and give your day a little direction.",
    emptyActive: "All clear. Nice work.", emptyActiveCopy: "You’ve made it through everything on your list.",
    emptyCompleted: "No completed tasks just yet.", emptyCompletedCopy: "Finish a task and it’ll find its way here.",
    footer: "Small steps, meaningful days.", editHint: "Double-click a task to edit it",
    loading: "Loading tasks", markDone: "Mark task complete", markActive: "Mark task active",
    edit: "Edit task", remove: "Delete task", save: "Save", cancel: "Cancel",
    cancelEdit: "Cancel editing", invalidTitle: "A task needs a title.",
    connection: "Could not reach the server. Check your connection and try again.",
    requestFailed: "Request failed ({status}).", notFound: "Task not found.",
    validation: "Please check the task details.", taskAdded: "Task added",
    taskUpdated: "Task updated", taskRemoved: "Task removed", languageLabel: "Switch to Russian",
    dueDate: "Choose a due date and time", themeLight: "Switch to light theme",
    themeDark: "Switch to dark theme", dateLocale: "en-US",
    completeAll: "Complete all active tasks", allCompleted: "All active tasks completed",
    bulkPartial: "Completed {completed}; could not complete {failed}."
  }
};

export function remainingLabel(count, language) {
  if (language === "en") return count === 1 ? "remainingOne" : "remainingMany";
  const mod100 = count % 100;
  if (mod100 >= 11 && mod100 <= 14) return "remainingMany";
  const mod10 = count % 10;
  if (mod10 === 1) return "remainingOne";
  if (mod10 >= 2 && mod10 <= 4) return "remainingFew";
  return "remainingMany";
}
