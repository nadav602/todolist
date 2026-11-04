<template>
    <div class="app">
        <header class="app-header">
            <div class="header-title">
                <img
                    v-if="titleImage"
                    :src="titleImage"
                    alt="Title icon"
                    class="title-image"
                />
                <h1>{{ appTitle }}</h1>
            </div>
            <div class="header-actions">
                <button class="btn" @click="fetchTodos" :disabled="loading">
                    Reload
                </button>
                <button
                    class="icon-btn"
                    @click="toggleTheme"
                    :title="
                        theme === 'dark'
                            ? 'Switch to light theme'
                            : 'Switch to dark theme'
                    "
                    aria-label="Toggle theme"
                >
                    <!-- Sun/Moon icon -->
                    <svg
                        xmlns="http://www.w3.org/2000/svg"
                        width="18"
                        height="18"
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="2"
                        stroke-linecap="round"
                        stroke-linejoin="round"
                    >
                        <circle
                            v-if="theme === 'light'"
                            cx="12"
                            cy="12"
                            r="5"
                        ></circle>
                        <path
                            v-else
                            d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"
                        ></path>
                    </svg>
                </button>
                <span v-if="loading" class="hint">Loading…</span>
                <span v-if="error" class="error">{{ error }}</span>
            </div>
        </header>

        <section class="card">
            <h2 class="card-title">Add Todo</h2>
            <div class="form-row">
                <input
                    class="input task-input"
                    v-model="newTask"
                    type="text"
                    placeholder="What needs doing?"
                    @keyup.enter="addTodo"
                    aria-label="New todo task"
                />
                <button
                    class="btn primary"
                    @click="addTodo"
                    :disabled="loading"
                >
                    Add
                </button>
            </div>
        </section>

        <section class="toolbar">
            <label for="filter" class="label">Filter:</label>
            <select
                id="filter"
                class="select"
                v-model="filterStatus"
                aria-label="Filter by status"
            >
                <option value="">All</option>
                <option v-for="s in STATUS_OPTIONS" :key="s" :value="s">
                    {{ s }}
                </option>
            </select>
            <label for="sort" class="label">Sort:</label>
            <select
                id="sort"
                class="select"
                v-model="sortBy"
                aria-label="Sort by field"
            >
                <option value="">None</option>
                <option value="id">ID</option>
                <option value="status">Status</option>
            </select>
            <button
                class="btn"
                @click="toggleReverse"
                :title="sortReverse ? 'Ascending' : 'Descending'"
                aria-label="Toggle sort order"
            >
                {{ sortReverse ? "↓" : "↑" }}
            </button>
        </section>

        <div v-if="filteredTodos.length === 0" class="empty">
            No todos to show. Add one above!
        </div>

        <ul class="todo-list">
            <li v-for="todo in filteredTodos" :key="todo.id" class="todo-item">
                <div class="todo-content">
                    <div class="todo-main">
                        <div class="todo-id">#{{ todo.id }}</div>
                        <input
                            class="input task-input"
                            :value="todo.task"
                            @change="updateTodoTask(todo, $event.target.value)"
                            aria-label="Edit task"
                        />
                    </div>
                    <div class="todo-meta">
                        <label class="label">Status</label>
                        <select
                            class="select"
                            :value="todo.status || 'Todo'"
                            @change="
                                updateTodoStatus(todo, $event.target.value)
                            "
                            aria-label="Edit status"
                        >
                            <option
                                v-for="s in STATUS_OPTIONS"
                                :key="s"
                                :value="s"
                            >
                                {{ s }}
                            </option>
                        </select>
                        <button
                            class="icon-btn danger"
                            @click="deleteTodo(todo.id)"
                            :disabled="loading"
                            aria-label="Delete todo"
                            title="Delete"
                        >
                            <svg
                                xmlns="http://www.w3.org/2000/svg"
                                width="18"
                                height="18"
                                viewBox="0 0 24 24"
                                fill="none"
                                stroke="currentColor"
                                stroke-width="2"
                                stroke-linecap="round"
                                stroke-linejoin="round"
                            >
                                <polyline points="3 6 5 6 21 6"></polyline>
                                <path
                                    d="M19 6l-1 14a2 2 0 0 1-2 2H8a2 2 0 0 1-2-2L5 6"
                                ></path>
                                <path d="M10 11v6"></path>
                                <path d="M14 11v6"></path>
                                <path
                                    d="M9 6V4a2 2 0 0 1 2-2h2a2 2 0 0 1 2 2v2"
                                ></path>
                            </svg>
                        </button>
                    </div>
                </div>
            </li>
        </ul>
    </div>
</template>

<script>
import axios from "axios";

const STATUS_OPTIONS = ["Todo", "In Progress", "Complete"];

export default {
    name: "App",
    data() {
        const rawApiBase =
            (typeof window !== "undefined" &&
                window.APP_CONFIG &&
                window.APP_CONFIG.API_BASE_URL) ||
            "";
        const apiBase = String(rawApiBase || "");
        const base = apiBase.endsWith("/") ? apiBase : apiBase + "/";

        // Load persisted theme or default to dark
        const savedTheme =
            (typeof localStorage !== "undefined" &&
                localStorage.getItem("theme")) ||
            "dark";

        return {
            apiBase: base,
            http: axios.create({
                baseURL: base,
                headers: { "Content-Type": "application/json" },
            }),
            todos: [],
            loading: false,
            error: null,
            newTask: "",
            newStatus: "Todo", // default; not shown in Add Todo form
            filterStatus: "",
            sortBy: "",
            sortReverse: false,
            theme: savedTheme,
        };
    },
    computed: {
        STATUS_OPTIONS() {
            return STATUS_OPTIONS;
        },
        appTitle() {
            const title =
                (typeof window !== "undefined" &&
                    window.APP_CONFIG &&
                    window.APP_CONFIG.TITLE) ||
                null;
            return title || "My TODOS";
        },
        titleImage() {
            return "/title-image.png";
        },
        filteredTodos() {
            const status = (this.filterStatus || "").trim();
            let result = this.todos;

            // Apply filter
            if (status) {
                result = result.filter((t) => (t.status || "Todo") === status);
            }

            // Apply sorting
            if (this.sortBy) {
                result = [...result].sort((a, b) => {
                    let aVal = a[this.sortBy];
                    let bVal = b[this.sortBy];

                    if (this.sortBy === "id") {
                        aVal = Number(aVal);
                        bVal = Number(bVal);
                    } else if (this.sortBy === "status") {
                        aVal = String(aVal || "Todo");
                        bVal = String(bVal || "Todo");
                    }

                    if (aVal < bVal) return this.sortReverse ? 1 : -1;
                    if (aVal > bVal) return this.sortReverse ? -1 : 1;
                    return 0;
                });
            }

            return result;
        },
    },
    async created() {
        this.applyTheme(this.theme);
        await this.fetchTodos().catch((err) => {
            const msg = err?.message || "fetchTodos failed";
            console.error("[App.vue] created() fetchTodos error:", msg, err);
        });
    },
    methods: {
        toggleTheme() {
            this.theme = this.theme === "dark" ? "light" : "dark";
            this.applyTheme(this.theme);
            try {
                localStorage.setItem("theme", this.theme);
            } catch {}
        },
        applyTheme(theme) {
            const root = document.documentElement;
            root.setAttribute("data-theme", theme);
        },
        _fmtErr(err, fallback) {
            const msg =
                err?.response?.data?.error ||
                err?.message ||
                fallback ||
                "Unexpected error";
            return msg;
        },
        async fetchTodos() {
            this.loading = true;
            this.error = null;
            try {
                const { data } = await this.http.get("");
                this.todos = Array.isArray(data)
                    ? data.map((t) => ({ ...t, status: t.status || "Todo" }))
                    : [];
            } catch (err) {
                this.error = this._fmtErr(err, "Failed to load todos");
            } finally {
                this.loading = false;
            }
        },
        async addTodo() {
            const task = this.newTask.trim();
            if (!task) {
                this.error = "Task cannot be empty";
                return;
            }
            this.loading = true;
            this.error = null;
            try {
                const { data } = await this.http.post("", { task });
                const created = data;
                const desiredStatus = this.newStatus || "Todo";
                if (desiredStatus && desiredStatus !== created.status) {
                    await this.http.patch(`${created.id}`, {
                        status: desiredStatus,
                    });
                    created.status = desiredStatus;
                }
                this.todos.unshift(created);
                this.newTask = "";
                this.newStatus = "Todo";
            } catch (err) {
                this.error = this._fmtErr(err, "Failed to add todo");
            } finally {
                this.loading = false;
            }
        },
        async updateTodoStatus(todo, status) {
            if (!STATUS_OPTIONS.includes(status)) {
                this.error = "Invalid status";
                return;
            }
            this.loading = true;
            this.error = null;
            try {
                await this.http.patch(`${todo.id}`, { status });
                todo.status = status;
            } catch (err) {
                this.error = this._fmtErr(err, "Failed to update status");
            } finally {
                this.loading = false;
            }
        },
        async updateTodoTask(todo, newTask) {
            const task = String(newTask || "").trim();
            if (!task) {
                this.error = "Task cannot be empty";
                return;
            }
            this.loading = true;
            this.error = null;
            try {
                await this.http.patch(`${todo.id}`, { task });
                todo.task = task;
            } catch (err) {
                this.error = this._fmtErr(err, "Failed to update task");
            } finally {
                this.loading = false;
            }
        },
        async deleteTodo(id) {
            this.loading = true;
            this.error = null;
            try {
                await this.http.delete(`${id}`);
                this.todos = this.todos.filter((t) => t.id !== id);
            } catch (err) {
                this.error = this._fmtErr(err, "Failed to delete todo");
            } finally {
                this.loading = false;
            }
        },
        toggleReverse() {
            this.sortReverse = !this.sortReverse;
        },
    },
};
</script>

<style>
:root {
    /* Dark theme (default) */
    --bg: #0f172a;
    --panel: #111827;
    --muted: #6b7280;
    --text: #e5e7eb;
    --primary: #2563eb;
    --danger: #dc2626;
    --border: #1f2937;
    --accent: #374151;

    /* Controls */
    --control-bg: #0b1220;
    --control-border: var(--accent);
    --control-hover-bg: #0d1426;
}

:root[data-theme="light"] {
    /* Light theme */
    --bg: #f8fafc;
    --panel: #ffffff;
    --muted: #6b7280;
    --text: #0f172a;
    --primary: #2563eb;
    --danger: #dc2626;
    --border: #e5e7eb;
    --accent: #d1d5db;

    /* Controls */
    --control-bg: #ffffff;
    --control-border: var(--accent);
    --control-hover-bg: #f3f4f6; /* light hover */
}

* {
    box-sizing: border-box;
}

html,
body {
    margin: 0;
    padding: 0;
    background: var(--bg);
    color: var(--text);
    font-family:
        system-ui,
        -apple-system,
        Segoe UI,
        Roboto,
        sans-serif;
}

.app {
    max-width: 840px;
    margin: 32px auto;
    padding: 0 16px;
}

.app-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    margin-bottom: 16px;
}

.header-title {
    display: flex;
    align-items: center;
    gap: 10px;
}

.title-image {
    height: 64px;
    width: 64px;
    object-fit: contain;
    flex-shrink: 0;
}

.app-header h1 {
    margin: 0;
    font-size: 28px;
    letter-spacing: 0.2px;
}

.header-actions {
    display: flex;
    gap: 12px;
    align-items: center;
}

.card {
    background: var(--panel);
    border: 1px solid var(--border);
    border-radius: 12px;
    padding: 16px;
    margin: 12px 0 20px;
    box-shadow: 0 1px 0 rgba(0, 0, 0, 0.15);
}

.card-title {
    margin: 0 0 12px;
    font-size: 18px;
    color: var(--text);
}

.toolbar {
    display: flex;
    align-items: center;
    gap: 10px;
    margin: 8px 0 16px;
}

.form-row {
    display: flex;
    gap: 10px;
    align-items: center;
    flex-wrap: wrap;
}
.form-row .task-input {
    flex: 1;
    min-width: 320px;
}

.input,
.select {
    background: var(--control-bg);
    color: var(--text);
    border: 1px solid var(--control-border);
    border-radius: 8px;
    padding: 8px 10px;
    outline: none;
}

.input:focus,
.select:focus {
    border-color: var(--primary);
    box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.2);
}

.task-input {
    flex: 1;
    min-width: 240px;
}

.btn {
    appearance: none;
    border: 1px solid var(--control-border);
    background: var(--control-bg);
    color: var(--text);
    padding: 8px 12px;
    border-radius: 8px;
    cursor: pointer;
    transition:
        transform 0.05s ease,
        background 0.2s ease,
        border-color 0.2s ease;
}

.btn:hover {
    background: var(--control-hover-bg);
    border-color: var(--text);
}

.btn:active {
    transform: translateY(1px);
}

.btn.primary {
    background: var(--primary);
    border-color: var(--primary);
    color: #fff; /* ensure readable text in both themes */
}

.btn.primary:hover {
    background: #1f57d0;
    border-color: #1f57d0;
}

.btn.danger {
    background: var(--danger);
    border-color: var(--danger);
}

/* Icon button style for inline delete */
.icon-btn {
    appearance: none;
    border: 1px solid var(--control-border);
    background: var(--control-bg);
    color: var(--text);
    padding: 6px;
    border-radius: 8px;
    cursor: pointer;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    transition:
        transform 0.05s ease,
        background 0.2s ease,
        border-color 0.2s ease;
}

.icon-btn:hover {
    background: var(--control-hover-bg);
    border-color: var(--text);
}

.icon-btn.danger {
    color: #fca5a5;
    border-color: var(--danger);
}

.icon-btn.danger:hover {
    background: #b91c1c;
    color: #fff;
    border-color: #b91c1c;
}

.btn.danger:hover {
    background: #b91c1c;
    border-color: #b91c1c;
}

.label {
    color: var(--muted);
    font-size: 14px;
}

.hint {
    color: var(--muted);
}

.error {
    color: #fca5a5;
}

.empty {
    color: var(--muted);
    font-style: italic;
    padding: 16px;
    text-align: center;
}

.todo-list {
    list-style: none;
    padding: 0;
    margin: 0;
}

.todo-item {
    background: var(--panel);
    border: 1px solid var(--border);
    border-radius: 12px;
    padding: 12px;
    margin-bottom: 10px;
}

.todo-content {
    display: flex;
    gap: 12px;
    align-items: flex-start;
    justify-content: space-between;
    flex-wrap: wrap;
}

.todo-main {
    display: flex;
    gap: 10px;
    align-items: center;
    flex: 1;
}

.todo-id {
    font-weight: 600;
    color: var(--muted);
}

.todo-meta {
    display: flex;
    gap: 8px;
    align-items: center;
}

.todo-actions {
    margin-top: 8px;
    display: flex;
    justify-content: flex-end;
}
.app-footer {
    margin-top: 24px;
    color: var(--muted);
    text-align: center;
}
</style>
