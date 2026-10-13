export function createNotifications(t) {
  const bridge = window.Android;
  const timers = new Map();
  const maxDelay = 2_147_000_000;
  let enabled = false;

  function isAvailable() {
    return typeof bridge !== "undefined" && typeof bridge.showNotification === "function";
  }

  async function requestPermission() {
    if (!isAvailable() || typeof bridge.requestNotificationPermission !== "function") return false;
    enabled = bridge.requestNotificationPermission() === true;
    return enabled;
  }

  function cancel(taskID) {
    const timer = timers.get(taskID);
    if (timer !== undefined) window.clearTimeout(timer);
    timers.delete(taskID);
  }

  function schedule(task) {
    cancel(task.id);
    if (!enabled || task.done || !task.due_at || !isAvailable()) return;

    const dueAt = new Date(task.due_at).getTime();
    if (!Number.isFinite(dueAt) || dueAt <= Date.now()) return;

    const delay = Math.min(dueAt - Date.now(), maxDelay);
    timers.set(task.id, window.setTimeout(() => {
      if (Date.now() < dueAt) {
        schedule(task);
        return;
      }
      timers.delete(task.id);
      bridge.showNotification(notificationID(task.id), t("taskDue", { title: task.title }), t("taskDueMessage", { title: task.title }));
    }, delay));
  }

  function sync(tasks) {
    for (const taskID of timers.keys()) cancel(taskID);
    tasks.forEach(schedule);
  }

  return { isAvailable, requestPermission, schedule, cancel, sync };
}

function notificationID(value) {
  let hash = 0;
  for (let i = 0; i < value.length; i++) hash = (Math.imul(hash, 31) + value.charCodeAt(i)) | 0;
  return hash & 0x7fffffff;
}
