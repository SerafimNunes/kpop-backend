/**
 * Auren Platform - Core Engine v2.6 (Production Ready)
 * Responsabilidade: Barramento Central de Dados e WebSocket Resiliente.
 */
function aurenPlatform() {
  return {
    currentTab: "dashboard",
    sections: {},
    connected: false,
    token:
      new URLSearchParams(window.location.search).get("token") ||
      "AUREN-PLATFORM-DEMO",

    menuItems: [
      {
        id: "dashboard",
        label: "Dashboard",
        icon: '<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path d="M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zM14 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zM14 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z"/></svg>',
      },
      {
        id: "business",
        label: "Business",
        icon: '<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/></svg>',
      },
      {
        id: "intelligence",
        label: "Insight Engine",
        icon: '<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path d="M13 10V3L4 14h7v7l9-11h-7z"/></svg>',
      },
      {
        id: "pgrs-form",
        label: "Software PGRS",
        icon: '<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"/></svg>',
      },
      {
        id: "vistoria",
        label: "Checklist",
        icon: '<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"/></svg>',
      },
      {
        id: "auditoria",
        label: "Auditoria",
        icon: '<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04a11.357 11.357 0 00-1.026 5.405c0 5.992 4.137 11.03 9.644 12.315a12.164 12.164 0 009.644-12.315c0-1.879-.451-3.65-1.246-5.208z"/></svg>',
      },
    ],

    async init() {
      console.log("[AUREN-CORE] Booting...");

      // 1. Inicia Socket
      this.initSocket();

      // 2. Observer de Sincronia de Estado
      window.addEventListener("storage", (e) => {
        if (e.key.startsWith("auren_state_")) {
          window.dispatchEvent(
            new CustomEvent("auren-sync", { detail: e.key })
          );
        }
      });

      // 3. Carregamento Paralelo de Seções
      const loadSections = this.menuItems.map(async (item) => {
        try {
          const r = await fetch(
            `./sections/${item.id}.html?token=${this.token}`
          );
          this.sections[item.id] = r.ok
            ? await r.text()
            : `<div class="p-10 text-red-500">Falha na Seção: ${item.id}</div>`;
        } catch (err) {
          console.error(`Erro ao carregar ${item.id}`, err);
        }
      });
      await Promise.all(loadSections);
    },

    initSocket() {
      const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
      const wsUrl = `${protocol}//${window.location.host}/ws/engine?token=${this.token}`;

      const socket = new WebSocket(wsUrl);

      socket.onopen = () => {
        this.connected = true;
        // ATUALIZAÇÃO CRÍTICA: Expõe o objeto socket aberto globalmente
        window.AurenSocket = socket;
        console.log("✅ [BARRAMENTO CENTRAL] Conectado e Globalizado.");
        window.dispatchEvent(new CustomEvent("socket:ready"));
      };

      socket.onmessage = (event) => {
        try {
          const msg = JSON.parse(event.data);
          // Broadcast para os módulos Alpine (Business, Vistoria, etc)
          window.dispatchEvent(
            new CustomEvent(`socket:${msg.type}`, { detail: msg.payload })
          );
        } catch (e) {
          console.error("Erro no processamento do frame:", e);
        }
      };

      socket.onclose = () => {
        this.connected = false;
        window.AurenSocket = null; // Limpa referência para evitar erros de envio
        console.warn("⚠️ [BARRAMENTO] Offline. Reconectando...");
        setTimeout(() => this.initSocket(), 3000);
      };

      socket.onerror = (err) => {
        console.error("Erro fatal no barramento:", err);
      };
    },

    navigate(tabId) {
      this.currentTab = tabId;
      window.location.hash = tabId;
    },
  };
}
