document.addEventListener("DOMContentLoaded", () => {
  const token = localStorage.getItem("token");

  function parseJwt(t) {
    if (!t) return null;
    try {
      const base64Url = t.split(".")[1];
      const base64 = base64Url.replace(/-/g, "+").replace(/_/g, "/");
      const json = decodeURIComponent(
        atob(base64)
          .split("")
          .map((c) => "%" + ("00" + c.charCodeAt(0).toString(16)).slice(-2))
          .join(""),
      );
      return JSON.parse(json);
    } catch (e) {
      return null;
    }
  }

  const signoutBtn = document.getElementById("signout-btn");
  if (signoutBtn) {
    signoutBtn.addEventListener("click", () => {
      localStorage.removeItem("token");
      window.location.href = "/login";
    });
  }

  const loginForm = document.getElementById("login-form");
  if (loginForm) {
    if (token) {
      window.location.href = "/";
      return;
    }
    loginForm.addEventListener("submit", async (e) => {
      e.preventDefault();
      const login = loginForm.login.value.trim();
      const password = loginForm.password.value.trim();

      try {
        const res = await fetch("/api/v1/auth/sign-in", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ login, password }),
        });
        const data = await res.json();

        if (res.ok) {
          localStorage.setItem("token", data.result);
          window.location.href = "/";
        } else {
          alert(data.error || "Login error");
        }
      } catch (err) {
        alert("Network error: " + err.message);
      }
    });
  }

  const signupForm = document.getElementById("signup-form");
  if (signupForm) {
    if (token) {
      window.location.href = "/";
      return;
    }
    signupForm.addEventListener("submit", async (e) => {
      e.preventDefault();
      const login = signupForm.login.value.trim();
      const password = signupForm.password.value.trim();
      const role = signupForm.role.value;

      try {
        const res = await fetch("/api/v1/auth/sign-up", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ login, password, role }),
        });
        const data = await res.json();

        if (res.ok) {
          localStorage.setItem("token", data.result);
          window.location.href = "/";
        } else {
          alert(data.error || "Registration error");
        }
      } catch (err) {
        alert("Network error: " + err.message);
      }
    });
  }

  const tbody = document.getElementById("items-tbody");
  if (!tbody) return;

  if (!token) {
    window.location.href = "/login";
    return;
  }

  const dashboard = document.getElementById("dashboard");
  const addBtn = document.getElementById("add-item-btn");
  const loginLink = document.getElementById("login-link");
  const signupLink = document.getElementById("signup-link");

  let currentToken = token;
  let currentRole = null;

  function updateAuthUI() {
    const payload = parseJwt(currentToken);
    if (!payload) return;

    currentRole = payload.role || payload.Role;
    const userId = payload.sub || payload.Subject;

    document.getElementById("current-role").textContent = currentRole;
    document.getElementById("current-user-id").textContent = userId;

    loginLink.style.display = "none";
    signupLink.style.display = "none";
    signoutBtn.style.display = "inline-block";

    if (currentRole === "manager" || currentRole === "admin") {
      addBtn.style.display = "inline-block";
    }
  }

  async function loadItems() {
    try {
      const res = await fetch("/api/v1/items", {
        headers: { Authorization: `Bearer ${currentToken}` },
      });
      const data = await res.json();
      if (res.ok) {
        renderItems(data.result || []);
      } else {
        if (data.error?.toLowerCase().includes("token")) {
          localStorage.removeItem("token");
          window.location.href = "/login";
        } else {
          alert(data.error || "Error loading items");
        }
      }
    } catch (err) {
      alert("Network error");
    }
  }

  function renderItems(items) {
    tbody.innerHTML = "";

    const isAdmin = currentRole === "admin";
    const canSeeDeleted = isAdmin;

    items.forEach((item) => {
      if (!canSeeDeleted && item.deleted_at) return;

      const tr = document.createElement("tr");
      if (item.deleted_at) tr.classList.add("deleted-row");

      const nameHtml = item.deleted_at
        ? `<span class="deleted" title="Deleted on ${new Date(item.deleted_at).toLocaleString("en-US")}">${item.name}</span>`
        : item.name;

      const showEdit =
        (currentRole === "manager" || currentRole === "admin") &&
        !item.deleted_at;
      const showDelete = currentRole === "admin" && !item.deleted_at;
      const showHistory = currentRole === "admin";

      tr.innerHTML = `
                 <td>${item.id}</td>
                 <td>${nameHtml}</td>
                 <td>${item.description || "-"}</td>
                 <td>${item.quantity}</td>
                 <td>${item.price}</td>
                 <td>${new Date(item.updated_at).toLocaleString("en-US")}</td>
                 <td>
                    ${showEdit ? `<button onclick="editItem(${item.id})" class="btn-small">Edit</button>` : ""}
                    ${showDelete ? `<button onclick="deleteItem(${item.id})" class="btn-small danger">Delete</button>` : ""}
                    ${showHistory ? `<button onclick="showHistory(${item.id}, '${item.name.replace(/'/g, "\\'")}')" class="btn-small">History</button>` : ""}
                 </td>
            `;
      tbody.appendChild(tr);
    });
  }

  window.editItem = async (id) => {
    try {
      const res = await fetch(`/api/v1/items/${id}`, {
        headers: { Authorization: `Bearer ${currentToken}` },
      });
      const data = await res.json();
      if (!res.ok) throw new Error(data.error);

      const item = data.result;
      document.getElementById("edit-item-id").value = item.id;
      document.getElementById("item-name").value = item.name;
      document.getElementById("item-description").value =
        item.description || "";
      document.getElementById("item-quantity").value = item.quantity;
      document.getElementById("item-price").value = item.price;
      document.getElementById("modal-title").textContent = "Edit Item";
      document.getElementById("item-modal").style.display = "flex";
    } catch (err) {
      alert(err.message);
    }
  };

  window.deleteItem = async (id) => {
    if (!confirm("Delete item?")) return;
    try {
      const res = await fetch(`/api/v1/items/${id}`, {
        method: "DELETE",
        headers: { Authorization: `Bearer ${currentToken}` },
      });
      if (res.ok) loadItems();
      else {
        const data = await res.json();
        alert(data.error || "Delete error");
      }
    } catch (err) {
      alert("An error occurred");
    }
  };

  const itemForm = document.getElementById("item-form");
  if (itemForm) {
    itemForm.addEventListener("submit", async (e) => {
      e.preventDefault();
      const id = document.getElementById("edit-item-id").value;
      const method = id ? "PUT" : "POST";
      const url = id ? `/api/v1/items/${id}` : "/api/v1/items";

      const body = {
        name: document.getElementById("item-name").value,
        description: document.getElementById("item-description").value,
        quantity: parseInt(document.getElementById("item-quantity").value),
        price: parseFloat(document.getElementById("item-price").value),
      };

      try {
        const res = await fetch(url, {
          method,
          headers: {
            "Content-Type": "application/json",
            Authorization: `Bearer ${currentToken}`,
          },
          body: JSON.stringify(body),
        });
        if (res.ok) {
          closeModal("item-modal");
          loadItems();
        } else {
          const data = await res.json();
          alert(data.error || "Save error");
        }
      } catch (err) {
        alert("An error occurred");
      }
    });
  }

  if (addBtn) {
    addBtn.addEventListener("click", () => {
      document.getElementById("edit-item-id").value = "";
      document.getElementById("item-name").value = "";
      document.getElementById("item-description").value = "";
      document.getElementById("item-quantity").value = "";
      document.getElementById("item-price").value = "";
      document.getElementById("modal-title").textContent = "Add Item";
      document.getElementById("item-modal").style.display = "flex";
    });
  }

  let currentHistoryItemId = null;

  window.showHistory = (id, name) => {
    currentHistoryItemId = id;
    document.getElementById("history-item-title").textContent =
      `#${id} — ${name}`;
    document.getElementById("history-modal").style.display = "flex";
    loadHistory();
  };

  window.loadHistory = async () => {
    if (!currentHistoryItemId) return;
    const params = new URLSearchParams();
    ["h-from", "h-to", "h-user", "h-action", "h-limit"].forEach((id) => {
      const val = document.getElementById(id).value;
      if (val) params.append(id.replace("h-", ""), val);
    });

    try {
      const res = await fetch(
        `/api/v1/items/${currentHistoryItemId}/history?${params.toString()}`,
        { headers: { Authorization: `Bearer ${currentToken}` } },
      );
      const data = await res.json();
      if (res.ok) renderHistory(data.result || []);
      else alert(data.error);
    } catch (err) {
      alert("An error occurred");
    }
  };

  function renderHistory(history) {
    const tbody = document.getElementById("history-tbody");
    tbody.innerHTML = "";

    history.forEach((h) => {
      const tr = document.createElement("tr");
      const oldStr = h.old_data
        ? JSON.stringify(h.old_data, null, 2).slice(0, 120) + "..."
        : "-";
      const newStr = h.new_data
        ? JSON.stringify(h.new_data, null, 2).slice(0, 120) + "..."
        : "-";

      tr.innerHTML = `
                 <td>${h.id}</td>
                 <td>${h.user_id}</td>
                 <td>${h.action}</td>
                 <td>${new Date(h.changed_at).toLocaleString("en-US")}</td>
                 <td><pre class="history-pre">${oldStr}</pre></td>
                 <td><pre class="history-pre">${newStr}</pre></td>
                 <td>
                    ${
                      h.action === "UPDATE"
                        ? `<button class="diff-btn btn-small" data-old='${JSON.stringify(h.old_data || {})}' data-new='${JSON.stringify(h.new_data || {})}'>Diff</button>`
                        : "-"
                    }
                 </td>
            `;
      tbody.appendChild(tr);
    });

    document.querySelectorAll(".diff-btn").forEach((btn) => {
      btn.addEventListener("click", () => {
        const oldData = JSON.parse(btn.dataset.old || "{}");
        const newData = JSON.parse(btn.dataset.new || "{}");
        showDiff(oldData, newData);
      });
    });
  }

  window.showDiff = (oldData, newData) => {
    document.getElementById("diff-old").textContent = JSON.stringify(
      oldData,
      null,
      2,
    );
    document.getElementById("diff-new").textContent = JSON.stringify(
      newData,
      null,
      2,
    );
    document.getElementById("diff-modal").style.display = "flex";
  };

  window.exportHistoryCSV = async () => {
    if (!currentHistoryItemId) return;
    const params = new URLSearchParams();
    ["h-from", "h-to", "h-user", "h-action", "h-limit"].forEach((id) => {
      const val = document.getElementById(id).value;
      if (val) params.append(id.replace("h-", ""), val);
    });
    params.append("export", "csv");

    try {
      const res = await fetch(
        `/api/v1/items/${currentHistoryItemId}/history?${params.toString()}`,
        { headers: { Authorization: `Bearer ${currentToken}` } },
      );
      if (!res.ok) throw new Error("Export error");
      const blob = await res.blob();
      const url = URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = `item_${currentHistoryItemId}_history_${new Date().toISOString().slice(0, 10)}.csv`;
      a.click();
      URL.revokeObjectURL(url);
    } catch (err) {
      alert("Export error");
    }
  };

  window.closeModal = (id) => {
    document.getElementById(id).style.display = "none";
  };

  document.querySelectorAll(".modal").forEach((modal) => {
    modal.addEventListener("click", (e) => {
      if (e.target === modal) closeModal(modal.id);
    });
  });

  dashboard.style.display = "block";
  updateAuthUI();
  loadItems();
});
