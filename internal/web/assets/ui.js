const API = "/api/tasks";
const $ = (selector) => document.querySelector(selector);

export function createUI(t, icons) {
  function setTheme(theme) {
    const dark = theme === "auto" ? matchMedia("(prefers-color-scheme: dark)").matches : theme === "dark";
    document.documentElement.dataset.theme = dark ? "dark" : "light";
    localStorage.setItem("daymark-theme", theme);
    $("#theme-toggle").innerHTML = dark ? icons.sun : icons.moon;
    $("#theme-toggle").setAttribute("aria-label", dark ? t("themeLight") : t("themeDark"));
  }

  function initTheme() {
    const saved = localStorage.getItem("daymark-theme");
    const media = matchMedia("(prefers-color-scheme: dark)");
    setTheme(saved === "dark" || saved === "light" ? saved : "auto");
    media.addEventListener("change", () => {
      if (localStorage.getItem("daymark-theme") === "auto") setTheme("auto");
    });
    $("#theme-toggle").addEventListener("click", () => {
      const current = localStorage.getItem("daymark-theme");
      setTheme(current === "auto" ? "dark" : current === "dark" ? "light" : "auto");
    });
  }

  function toast(message, kind = "success") {
    const node = document.createElement("div");
    node.className = `toast toast-${kind}`;
    node.innerHTML = `${kind === "error" ? icons.error : icons.success}<span></span>`;
    $(".toast-region").appendChild(node);
    node.querySelector("span").textContent = message;
    requestAnimationFrame(() => node.classList.add("toast-visible"));
    window.setTimeout(() => {
      node.classList.remove("toast-visible");
      node.addEventListener("transitionend", () => node.remove(), { once: true });
    }, 3200);
  }

  async function request(path = "", options = {}) {
    let response;
    try {
      response = await fetch(API + path, {
        ...options,
        headers: { "Content-Type": "application/json", ...options.headers }
      });
    } catch {
      throw new Error(t("connection"));
    }
    if (!response.ok) {
      let message = t("requestFailed", { status: response.status });
      try {
        const body = await response.json();
        const codes = { NOT_FOUND: "notFound", VALIDATION_ERROR: "validation", INTERNAL_ERROR: "requestFailed" };
        message = codes[body.error?.code] ? t(codes[body.error.code], { status: response.status }) : message;
      } catch {}
      throw new Error(message);
    }
    return response.status === 204 ? null : response.json();
  }

  return { initTheme, request, toast };
}
