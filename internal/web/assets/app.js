import { icons, remainingLabel, text } from "./translations.js";
import { createUI } from "./ui.js";
import { createNotifications } from "./notifications.js";

(() => {
  "use strict";

  const $ = (selector, root = document) => root.querySelector(selector);
  const list = $("#task-list");
  const template = $("#task-template");
  const tasksByID = new Map();
  let filter = "all";
  let loaded = false;
  let loadSequence = 0;
  let bulkCompleting = false;
  let language = localStorage.getItem("daymark-language") || "ru";
  if (!text[language]) language = "ru";

  function t(key, values = {}) {
    return Object.entries(values).reduce((value, [name, replacement]) => value.replaceAll(`{${name}}`, replacement), text[language][key]);
  }
  const { initTheme, request, toast } = createUI(t, icons);
  const notifications = createNotifications(t);
  let dueAtInput = null;

  function applyLanguage() {
    const labels = text[language];
    document.documentElement.lang = language;
    document.title = labels.title;
    $('meta[name="description"]').content = labels.description;
    $(".nav-list").setAttribute("aria-label", labels.filters);
    document.querySelectorAll("[data-filter]").forEach(button => {
      $("span:not(.nav-count)", button).textContent = labels[button.dataset.filter];
    });
    $("#complete-all").setAttribute("aria-label", labels.completeAll);
    $("#complete-all").title = labels.completeAll;
    $("#page-label").textContent = labels[filter];
    $("#page-title").innerHTML = labels.heading;
    $(".page-subtitle").textContent = labels.subtitle;
    $(".progress-card").setAttribute("aria-label", labels.progress);
    $(".progress-card strong").textContent = labels.progress;
    $("#title").placeholder = labels.newTask;
    $("#title").previousElementSibling.textContent = labels.composerTitle;
    $("#description").placeholder = labels.note;
    $("#description").previousElementSibling.textContent = labels.composerDescription;
    $("#submit-btn span").textContent = labels.add;
    $("#list-heading").textContent = labels[filter];
    $("#skeleton").setAttribute("aria-label", labels.loading);
    $(".page-footer span:first-child").textContent = labels.footer;
    $(".page-footer span:last-child").textContent = labels.editHint;
    $("#language-toggle").textContent = language === "ru" ? "EN" : "RU";
    $("#language-toggle").setAttribute("aria-label", labels.languageLabel);
    $("#theme-toggle").setAttribute("aria-label", document.documentElement.dataset.theme === "dark" ? labels.themeLight : labels.themeDark);
    $("#today-label").textContent = new Intl.DateTimeFormat(labels.dateLocale, {
      weekday: "long", month: "long", day: "numeric"
    }).format(new Date());
    if (loaded) render();
  }

  function visibleTasks() {
    return [...tasksByID.values()]
      .filter(task => filter === "all" || (filter === "active" ? !task.done : task.done))
      .sort((a, b) => new Date(b.created_at) - new Date(a.created_at));
  }

  function updateSummary(tasks) {
    const total = tasksByID.size;
    const remaining = [...tasksByID.values()].filter(task => !task.done).length;
    const complete = total - remaining;
    const progress = total ? Math.round(complete / total * 100) : 0;
    $("#all-count").textContent = total;
    $("#visible-count").textContent = tasks.length;
    $("#remaining-count").textContent = remaining ? t(remainingLabel(remaining), { n: remaining }) : t("allCaughtUp");
    $("#progress-value").textContent = `${progress}%`;
    $("#progress-fill").setAttribute("stroke-dashoffset", String(119.38 * (1 - progress / 100)));
    $("#progress-caption").textContent = total ? t("doneOf", { done: complete, total }) : t("started");
    $("#complete-all").disabled = bulkCompleting || remaining === 0;
  }

  function render() {
    const tasks = visibleTasks();
    list.replaceChildren();
    updateSummary(tasks);
    tasks.forEach(task => list.appendChild(createTaskElement(task)));
    const empty = tasks.length === 0;
    notifications.sync([...tasksByID.values()]);
    $("#empty-state").classList.toggle("hidden", !empty);
    $("#empty-title").textContent = tasksByID.size === 0
      ? t("empty")
      : filter === "active" ? t("emptyActive") : t("emptyCompleted");
    $("#empty-copy").textContent = tasksByID.size === 0
      ? t("emptyCopy")
      : filter === "active" ? t("emptyActiveCopy") : t("emptyCompletedCopy");
  }

  function createTaskElement(task) {
    const node = template.content.firstElementChild.cloneNode(true);
    node.dataset.id = task.id;
    node.classList.toggle("is-complete", task.done);
    const checkbox = $(".check-button", node);
    checkbox.innerHTML = icons.check;
    checkbox.setAttribute("aria-label", task.done ? t("markActive") : t("markDone"));
    checkbox.disabled = bulkCompleting;
    $(".task-title", node).textContent = task.title;
    const description = $(".task-description", node);
    description.textContent = task.description || "";
    description.classList.toggle("hidden", !task.description);
    const meta = $(".task-meta", node);
    if (task.due_at) {
      const due = new Date(task.due_at);
      meta.textContent = t("dueAt", { date: due.toLocaleDateString(t("dateLocale")), time: due.toLocaleTimeString(t("dateLocale"), { hour: "2-digit", minute: "2-digit" }) });
      const isOverdue = !task.done && due < new Date();
      meta.classList.toggle("overdue", isOverdue);
    } else {
      meta.textContent = "";
    }
    $(".edit-button", node).innerHTML = icons.edit;
    $(".delete-button", node).innerHTML = icons.trash;
    $(".edit-button", node).setAttribute("aria-label", t("edit"));
    $(".edit-button", node).title = t("edit");
    $(".delete-button", node).setAttribute("aria-label", t("remove"));
    $(".delete-button", node).title = t("remove");
    $(".edit-input", node).setAttribute("aria-label", t("composerTitle"));
    $(".edit-description-input", node).setAttribute("aria-label", t("composerDescription"));
    $(".edit-description-input", node).placeholder = t("note");
    if (task.due_at) {
      const due = new Date(task.due_at);
      const local = due.toISOString().slice(0, 16);
      $(".edit-due-input", node).value = local;
    }
    $(".save-edit", node).textContent = t("save");
    $(".cancel-edit", node).textContent = t("cancel");
    $(".cancel-edit", node).setAttribute("aria-label", t("cancelEdit"));
    checkbox.addEventListener("click", () => toggleTask(task.id));
    $(".delete-button", node).disabled = bulkCompleting;
    $(".edit-button", node).disabled = bulkCompleting;
    $(".delete-button", node).addEventListener("click", () => deleteTask(task.id));
    $(".edit-button", node).addEventListener("click", () => beginEdit(node, task));
    $(".task-title", node).addEventListener("dblclick", () => beginEdit(node, task));
    $(".task-title", node).addEventListener("keydown", event => {
      if (event.key === "Enter") beginEdit(node, task);
    });
    $(".edit-form", node).addEventListener("submit", event => saveEdit(event, node, task));
    $(".cancel-edit", node).addEventListener("click", () => endEdit(node));
    return node;
  }

  function beginEdit(node, task) {
    const form = $(".edit-form", node);
    const input = $(".edit-input", node);
    input.value = task.title;
    $(".edit-description-input", node).value = task.description || "";
    $(".task-title", node).classList.add("hidden");
    form.classList.remove("hidden");
    input.focus();
    input.select();
    form.onkeydown = event => {
      if (event.key === "Escape") {
        event.preventDefault();
        endEdit(node);
      }
    };
  }

  function endEdit(node) {
    $(".task-title", node).classList.remove("hidden");
    $(".edit-form", node).classList.add("hidden");
    $(".edit-form", node).onkeydown = null;
  }

  async function saveEdit(event, node, task) {
    event.preventDefault();
    const title = $(".edit-input", node).value.trim();
    const description = $(".edit-description-input", node).value.trim();
    const dueValue = $(".edit-due-input", node).value;
    const dueAt = dueValue ? new Date(dueValue).toISOString() : null;
    if (!title) {
      toast(t("invalidTitle"), "error");
      return;
    }
    const previous = { title: task.title, description: task.description, due_at: task.due_at };
    task.title = title;
    task.description = description;
    task.due_at = dueAt;
    endEdit(node);
    render();
    try {
      const updated = await request(`/${encodeURIComponent(task.id)}`, {
        method: "PATCH", body: JSON.stringify({ title, description, due_at: dueAt })
      });
      tasksByID.set(updated.id, updated);
      render();
      toast(t("taskUpdated"));
    } catch (error) {
      task.title = previous.title;
      task.description = previous.description;
      task.due_at = previous.due_at;
      render();
      toast(error.message, "error");
    }
  }

  async function toggleTask(id) {
    if (bulkCompleting) return;
    const task = tasksByID.get(id);
    if (!task) return;
    const previous = task.done;
    task.done = !previous;
    render();
    try {
      const updated = await request(`/${encodeURIComponent(id)}`, {
        method: "PATCH", body: JSON.stringify({ done: task.done })
      });
      tasksByID.set(updated.id, updated);
      render();
    } catch (error) {
      task.done = previous;
      render();
      toast(error.message, "error");
    }
  }

  async function completeAllTasks() {
    if (bulkCompleting) return;
    const activeTasks = [...tasksByID.values()].filter(task => !task.done);
    if (!activeTasks.length) return;

    bulkCompleting = true;
    activeTasks.forEach(task => { task.done = true; });
    render();

    const failedTasks = [];
    for (const task of activeTasks) {
      try {
        const updated = await request(`/${encodeURIComponent(task.id)}`, {
          method: "PATCH", body: JSON.stringify({ done: true })
        });
        tasksByID.set(updated.id, updated);
      } catch (error) {
        task.done = false;
        failedTasks.push(error);
      }
    }

    bulkCompleting = false;
    render();
    if (failedTasks.length) {
      toast(t("bulkPartial", {
        completed: activeTasks.length - failedTasks.length,
        failed: failedTasks.length
      }) + ` ${failedTasks[0].message}`, "error");
    } else {
      toast(t("allCompleted"));
    }
  }

  async function deleteTask(id) {
    const task = tasksByID.get(id);
    if (!task) return;
    const item = [...list.children].find(element => element.dataset.id === id);
    if (item) {
      item.classList.add("task-removing");
      await new Promise(resolve => window.setTimeout(resolve, 180));
    }
    tasksByID.delete(id);
    render();
    try {
      await request(`/${encodeURIComponent(id)}`, { method: "DELETE" });
      toast(t("taskRemoved"));
    } catch (error) {
      tasksByID.set(id, task);
      render();
      toast(error.message, "error");
    }
  }

  async function loadTasks(showSkeleton = false) {
    const sequence = ++loadSequence;
    if (showSkeleton) $("#skeleton").classList.remove("hidden");
    try {
      const tasks = await request("");
      if (sequence !== loadSequence) return;
      tasksByID.clear();
      (Array.isArray(tasks) ? tasks : []).forEach(task => tasksByID.set(task.id, task));
      render();
      loaded = true;
    } catch (error) {
      if (sequence === loadSequence) toast(error.message, "error");
    } finally {
      if (sequence === loadSequence) {
        $("#skeleton").classList.add("hidden");
        if (!loaded) $("#empty-state").classList.add("hidden");
      }
    }
  }

  function setup() {
    initTheme();
    dueAtInput = $("#due-at");
    $("#complete-all").addEventListener("click", completeAllTasks);
    $("#language-toggle").addEventListener("click", () => {
      language = language === "ru" ? "en" : "ru";
      localStorage.setItem("daymark-language", language);
      applyLanguage();
    });
    applyLanguage();
    $("#task-form").addEventListener("submit", async event => {
      event.preventDefault();
      const titleInput = $("#title");
      const title = titleInput.value.trim();
      const description = $("#description").value.trim();
      const dueAt = dueAtInput.value ? new Date(dueAtInput.value).toISOString() : null;
      if (!title) {
        titleInput.focus();
        return;
      }
      const button = $("#submit-btn");
      button.disabled = true;
      const temporaryID = `pending-${Date.now()}`;
      const optimisticTask = {
        id: temporaryID,
        title,
        description,
        due_at: dueAt,
        done: false,
        created_at: new Date().toISOString()
      };
      tasksByID.set(temporaryID, optimisticTask);
      render();
      try {
        const created = await request("", {
          method: "POST", body: JSON.stringify({ title, description, due_at: dueAt })
        });
        tasksByID.delete(temporaryID);
        tasksByID.set(created.id, created);
        $("#task-form").reset();
        titleInput.focus();
        render();
        toast(t("taskAdded"));
      } catch (error) {
        tasksByID.delete(temporaryID);
        render();
        toast(error.message, "error");
      } finally {
        button.disabled = false;
      }
    });
    document.querySelectorAll("[data-filter]").forEach(button => {
      button.addEventListener("click", () => {
        filter = button.dataset.filter;
        document.querySelectorAll("[data-filter]").forEach(item => item.classList.toggle("active", item === button));
        $("#page-label").textContent = t(filter);
        $("#list-heading").textContent = t(filter);
        render();
      });
    });
    notifications.requestPermission().then(granted => {
      if (granted) notifications.sync([...tasksByID.values()]);
    });
    loadTasks(true);
  }

  document.addEventListener("DOMContentLoaded", setup);
})();
