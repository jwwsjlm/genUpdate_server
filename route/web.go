package route

const indexHTML = `<!doctype html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>GenUpdate 更新中心</title>
  <style>
    :root {
      color-scheme: light;
      --bg: #f6f7f9;
      --panel: #ffffff;
      --text: #1f2937;
      --muted: #6b7280;
      --line: #d9dee7;
      --accent: #0f766e;
      --accent-strong: #0b5f59;
      --warn: #b45309;
      --shadow: 0 18px 46px rgba(31, 41, 55, .08);
      font-family: Inter, "Segoe UI", "Microsoft YaHei", Arial, sans-serif;
    }
    * { box-sizing: border-box; }
    body { margin: 0; background: var(--bg); color: var(--text); min-width: 320px; }
    button, input { font: inherit; }
    a { color: inherit; text-decoration: none; }
    .shell { max-width: 1180px; margin: 0 auto; padding: 24px; }
    .topbar { display: flex; align-items: center; justify-content: space-between; gap: 16px; min-height: 64px; margin-bottom: 18px; }
    .brand { display: flex; align-items: center; gap: 12px; min-width: 0; }
    .mark { width: 42px; height: 42px; border-radius: 8px; display: grid; place-items: center; background: #102a43; color: #fff; font-weight: 800; letter-spacing: 0; flex: 0 0 auto; }
    h1 { font-size: 24px; line-height: 1.2; margin: 0; letter-spacing: 0; }
    .subtitle { color: var(--muted); font-size: 13px; margin-top: 5px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
    .actions { display: flex; gap: 8px; align-items: center; flex-wrap: wrap; justify-content: flex-end; }
    .btn { border: 1px solid var(--line); background: var(--panel); color: var(--text); min-height: 38px; padding: 0 12px; border-radius: 8px; display: inline-flex; align-items: center; gap: 8px; cursor: pointer; box-shadow: 0 1px 0 rgba(31, 41, 55, .04); }
    .btn.primary { background: var(--accent); color: #fff; border-color: var(--accent); }
    .btn:hover { border-color: #aeb8c6; }
    .btn.primary:hover { background: var(--accent-strong); }
    .token-panel { background: var(--panel); border: 1px solid var(--line); border-radius: 8px; padding: 14px; margin-bottom: 18px; display: grid; grid-template-columns: minmax(0, 1fr) auto auto; gap: 10px; align-items: center; box-shadow: 0 1px 0 rgba(31, 41, 55, .03); }
    .token-panel[hidden] { display: none; }
    .token-panel input { width: 100%; height: 38px; border: 1px solid var(--line); border-radius: 8px; padding: 0 12px; outline: none; background: #fbfcfd; font-family: "Cascadia Mono", Consolas, monospace; font-size: 13px; }
    .token-status { color: var(--muted); font-size: 12px; white-space: nowrap; }
    .stats { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); gap: 12px; margin-bottom: 18px; }
    .stat { background: var(--panel); border: 1px solid var(--line); border-radius: 8px; padding: 16px; box-shadow: 0 1px 0 rgba(31, 41, 55, .03); min-width: 0; }
    .stat .label { color: var(--muted); font-size: 12px; margin-bottom: 9px; }
    .stat .value { font-size: 24px; font-weight: 750; overflow-wrap: anywhere; }
    .layout { display: grid; grid-template-columns: 300px minmax(0, 1fr); gap: 16px; align-items: start; }
    .sidebar, .content { background: var(--panel); border: 1px solid var(--line); border-radius: 8px; box-shadow: var(--shadow); min-width: 0; }
    .sidebar { overflow: hidden; }
    .search { padding: 14px; border-bottom: 1px solid var(--line); }
    .search input, .file-filter { width: 100%; height: 40px; border: 1px solid var(--line); border-radius: 8px; padding: 0 12px; outline: none; background: #fbfcfd; }
    .search input:focus, .file-filter:focus { border-color: var(--accent); box-shadow: 0 0 0 3px rgba(15, 118, 110, .12); }
    .app-list { max-height: calc(100vh - 262px); overflow: auto; padding: 6px; }
    .app-item { width: 100%; border: 0; background: transparent; border-radius: 8px; padding: 11px 10px; display: grid; gap: 5px; text-align: left; cursor: pointer; color: var(--text); }
    .app-item:hover { background: #f2f5f7; }
    .app-item.active { background: #e8f5f3; }
    .app-line { display: flex; align-items: center; justify-content: space-between; gap: 8px; min-width: 0; }
    .app-name { font-weight: 700; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
    .pill { color: var(--accent-strong); background: #dff3f0; border: 1px solid #bde3de; border-radius: 999px; padding: 2px 8px; font-size: 12px; flex: 0 0 auto; }
    .app-meta { color: var(--muted); font-size: 12px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
    .content { overflow: hidden; }
    .detail-head { padding: 18px; border-bottom: 1px solid var(--line); display: flex; align-items: flex-start; justify-content: space-between; gap: 16px; }
    .detail-title { font-size: 22px; font-weight: 780; margin-bottom: 7px; overflow-wrap: anywhere; }
    .detail-note { color: var(--muted); line-height: 1.55; max-width: 760px; overflow-wrap: anywhere; }
    .toolbar { padding: 14px 18px; display: flex; align-items: center; justify-content: space-between; gap: 12px; border-bottom: 1px solid var(--line); background: #fbfcfd; }
    .file-filter { max-width: 360px; background: #fff; }
    .table-wrap { overflow: auto; }
    table { width: 100%; border-collapse: collapse; min-width: 760px; }
    th, td { border-bottom: 1px solid var(--line); padding: 12px 14px; text-align: left; vertical-align: middle; font-size: 13px; }
    th { color: var(--muted); background: #fbfcfd; font-weight: 650; position: sticky; top: 0; z-index: 1; }
    .path { font-weight: 650; overflow-wrap: anywhere; }
    .hash { font-family: "Cascadia Mono", Consolas, monospace; color: #4b5563; max-width: 260px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
    .download { color: var(--accent-strong); font-weight: 700; }
    .empty, .error { padding: 34px 18px; color: var(--muted); text-align: center; }
    .error { color: var(--warn); }
    @media (max-width: 860px) {
      .shell { padding: 16px; }
      .topbar { align-items: flex-start; flex-direction: column; }
      .actions { justify-content: flex-start; }
      .stats { grid-template-columns: repeat(2, minmax(0, 1fr)); }
      .layout { grid-template-columns: 1fr; }
      .app-list { max-height: 260px; }
      .detail-head, .toolbar { flex-direction: column; align-items: stretch; }
      .token-panel { grid-template-columns: 1fr; }
    }
    @media (max-width: 520px) {
      .stats { grid-template-columns: 1fr; }
      h1 { font-size: 21px; }
      .stat .value { font-size: 21px; }
      .btn { width: 100%; justify-content: center; }
    }
  </style>
</head>
<body>
  <main class="shell">
    <header class="topbar">
      <div class="brand">
        <div class="mark">GU</div>
        <div>
          <h1>GenUpdate 更新中心</h1>
          <div class="subtitle" id="serverInfo">正在读取服务信息</div>
        </div>
      </div>
      <div class="actions">
        <button class="btn" id="generateTokenBtn" type="button">生成 Token</button>
        <button class="btn" id="refreshBtn" type="button">刷新</button>
        <a class="btn primary" href="/version" target="_blank" rel="noreferrer">版本接口</a>
      </div>
    </header>
    <section class="token-panel" id="tokenPanel" aria-label="Token generator" hidden>
      <input id="tokenOutput" type="text" readonly spellcheck="false">
      <button class="btn primary" id="copyTokenBtn" type="button">复制</button>
      <div class="token-status" id="tokenStatus">32 bytes · base64url</div>
    </section>
    <section class="stats" aria-label="统计信息">
      <div class="stat"><div class="label">软件数量</div><div class="value" id="totalApps">-</div></div>
      <div class="stat"><div class="label">文件数量</div><div class="value" id="totalFiles">-</div></div>
      <div class="stat"><div class="label">文件总量</div><div class="value" id="totalBytes">-</div></div>
      <div class="stat"><div class="label">清单缓存</div><div class="value" id="cacheAge">-</div></div>
    </section>
    <section class="layout">
      <aside class="sidebar" aria-label="软件列表">
        <div class="search"><input id="appSearch" type="search" placeholder="搜索软件"></div>
        <div class="app-list" id="appList"></div>
      </aside>
      <section class="content" aria-label="软件详情">
        <div class="detail-head">
          <div>
            <div class="detail-title" id="detailTitle">请选择软件</div>
            <div class="detail-note" id="detailNote">左侧列表会展示当前可更新的软件。</div>
          </div>
          <a class="btn" id="manifestLink" href="#" target="_blank" rel="noreferrer">清单 JSON</a>
        </div>
        <div class="toolbar">
          <input class="file-filter" id="fileSearch" type="search" placeholder="筛选文件名、路径或 SHA256">
          <div class="app-meta" id="fileSummary">-</div>
        </div>
        <div class="table-wrap" id="fileTable"></div>
      </section>
    </section>
  </main>
  <script>
    const state = { apps: [], selected: "", query: "", fileQuery: "", version: null };
    const el = (id) => document.getElementById(id);

    function formatBytes(bytes) {
      if (!bytes) return "0 B";
      const units = ["B", "KB", "MB", "GB", "TB"];
      let value = Number(bytes);
      let index = 0;
      while (value >= 1024 && index < units.length - 1) {
        value /= 1024;
        index += 1;
      }
      return value.toFixed(value >= 10 || index === 0 ? 0 : 1) + " " + units[index];
    }

    function formatDate(value) {
      if (!value) return "-";
      const date = new Date(value);
      if (Number.isNaN(date.getTime())) return value;
      return date.toLocaleString();
    }

    function currentApp() {
      return state.apps.find((app) => app.fileName === state.selected) || state.apps[0];
    }

    async function loadData() {
      try {
        const responses = await Promise.all([fetch("/api/apps"), fetch("/version")]);
        if (!responses[0].ok) throw new Error("应用清单读取失败");
        const appsData = await responses[0].json();
        state.apps = appsData.apps || [];
        state.version = responses[1].ok ? await responses[1].json() : null;
        if (!state.apps.some((app) => app.fileName === state.selected)) {
          state.selected = state.apps[0] ? state.apps[0].fileName : "";
        }
        renderStats(appsData);
        renderAppList();
        renderDetail();
      } catch (err) {
        el("appList").innerHTML = '<div class="error">' + escapeHTML(err.message) + '</div>';
        el("fileTable").innerHTML = '<div class="error">无法加载更新清单，请检查服务日志。</div>';
      }
    }

    function renderStats(data) {
      el("totalApps").textContent = data.totalApps ?? state.apps.length;
      el("totalFiles").textContent = data.totalFiles ?? "-";
      el("totalBytes").textContent = formatBytes(data.totalBytes || 0);
      el("cacheAge").textContent = state.version?.cacheMaxAge || "-";
      const build = state.version || {};
      el("serverInfo").textContent = "版本 " + (build.version || "dev") + " · 提交 " + (build.commit || "unknown");
    }

    function renderAppList() {
      const query = state.query.trim().toLowerCase();
      const apps = state.apps.filter((app) => {
        const note = app.ReleaseNote || {};
        return [app.fileName, note.appName, note.version, note.description]
          .filter(Boolean).join(" ").toLowerCase().includes(query);
      });
      if (!apps.length) {
        el("appList").innerHTML = '<div class="empty">没有匹配的软件</div>';
        return;
      }
      el("appList").innerHTML = apps.map((app) => {
        const note = app.ReleaseNote || {};
        const files = app.fileList || [];
        const name = note.appName || app.fileName;
        const active = app.fileName === state.selected ? " active" : "";
        const bytes = files.reduce((sum, file) => sum + (file.size || 0), 0);
        return '<button class="app-item' + active + '" type="button" data-app="' + escapeAttr(app.fileName) + '">' +
          '<span class="app-line"><span class="app-name">' + escapeHTML(name) + '</span>' +
          '<span class="pill">' + escapeHTML(note.version || "1.0.0") + '</span></span>' +
          '<span class="app-meta">' + files.length + ' 个文件 · ' + formatBytes(bytes) + '</span></button>';
      }).join("");
      document.querySelectorAll(".app-item").forEach((item) => {
        item.addEventListener("click", () => {
          state.selected = item.dataset.app;
          renderAppList();
          renderDetail();
        });
      });
    }

    function renderDetail() {
      const app = currentApp();
      if (!app) {
        el("detailTitle").textContent = "暂无软件";
        el("detailNote").textContent = "把软件目录放入 update 后，服务会自动生成清单。";
        el("manifestLink").href = "#";
        el("fileSummary").textContent = "-";
        el("fileTable").innerHTML = '<div class="empty">暂无文件</div>';
        return;
      }
      const note = app.ReleaseNote || {};
      const files = app.fileList || [];
      const name = note.appName || app.fileName;
      const description = note.description && note.description !== "null" ? note.description : "暂无更新说明";
      el("detailTitle").textContent = name + " · " + (note.version || "1.0.0");
      el("detailNote").textContent = description;
      el("manifestLink").href = "/updateList/" + encodeURIComponent(app.fileName);
      const fileQuery = state.fileQuery.trim().toLowerCase();
      const visible = files.filter((file) => [file.path, file.name, file.sha256]
        .filter(Boolean).join(" ").toLowerCase().includes(fileQuery));
      const totalBytes = files.reduce((sum, file) => sum + (file.size || 0), 0);
      el("fileSummary").textContent = visible.length + "/" + files.length + " 个文件 · " + formatBytes(totalBytes);
      if (!visible.length) {
        el("fileTable").innerHTML = '<div class="empty">没有匹配的文件</div>';
        return;
      }
      el("fileTable").innerHTML = '<table><thead><tr><th>文件</th><th>大小</th><th>修改时间</th><th>SHA256</th><th>操作</th></tr></thead><tbody>' +
        visible.map((file) => '<tr><td><div class="path">' + escapeHTML(file.path || file.name || "") + '</div></td>' +
        '<td>' + formatBytes(file.size || 0) + '</td><td>' + escapeHTML(formatDate(file.modTime)) + '</td>' +
        '<td><div class="hash" title="' + escapeAttr(file.sha256 || "") + '">' + escapeHTML(file.sha256 || "-") + '</div></td>' +
        '<td><a class="download" href="' + escapeAttr(file.downloadURL || "#") + '">下载</a></td></tr>').join("") +
        '</tbody></table>';
    }

    function escapeHTML(value) {
      return String(value).replace(/[&<>"']/g, (char) => ({
        "&": "&amp;",
        "<": "&lt;",
        ">": "&gt;",
        '"': "&quot;",
        "'": "&#39;"
      })[char]);
    }

    function escapeAttr(value) {
      return escapeHTML(value);
    }

    function generateToken(byteLength = 32) {
      const bytes = new Uint8Array(byteLength);
      crypto.getRandomValues(bytes);
      let binary = "";
      bytes.forEach((byte) => {
        binary += String.fromCharCode(byte);
      });
      return btoa(binary).replace(/\+/g, "-").replace(/\//g, "_").replace(/=+$/g, "");
    }

    async function copyToken() {
      const token = el("tokenOutput").value;
      if (!token) return;
      try {
        await navigator.clipboard.writeText(token);
        el("tokenStatus").textContent = "已复制";
      } catch {
        el("tokenOutput").select();
        document.execCommand("copy");
        el("tokenStatus").textContent = "已复制";
      }
    }

    el("generateTokenBtn").addEventListener("click", () => {
      el("tokenOutput").value = generateToken();
      el("tokenPanel").hidden = false;
      el("tokenStatus").textContent = "32 bytes · base64url";
    });
    el("copyTokenBtn").addEventListener("click", copyToken);
    el("refreshBtn").addEventListener("click", loadData);
    el("appSearch").addEventListener("input", (event) => {
      state.query = event.target.value;
      renderAppList();
    });
    el("fileSearch").addEventListener("input", (event) => {
      state.fileQuery = event.target.value;
      renderDetail();
    });
    loadData();
  </script>
</body>
</html>`
