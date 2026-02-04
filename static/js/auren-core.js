/**
 * AUREN PLATFORM | CORE ENGINE V1.4
 * Gerenciamento centralizado de estado, sockets e navegação.
 * CORRIGIDO: sendCommand agora usa 'action' em vez de 'type'
 */

function aurenPlatform() {
  return {
    // --- ESTADO DE NAVEGAÇÃO ---
    currentTab: window.location.hash.replace("#", "") || "dashboard",
    sections: {},

    // --- CONEXÃO E SEGURANÇA ---
    connected: false,
    token:
      new URLSearchParams(window.location.search).get("token") ||
      "AUREN-PLATFORM-DEMO",

    // --- FEEDBACK E INTERFACE ---
    notifications: [],
    globalStats: {
      balance: 0,
      criticalProjects: 0,
      lastUpdate: "Iniciando...",
    },

    // --- MENU DE SISTEMA ---
    menuItems: [
      {
        id: "dashboard",
        label: "Dashboard",
        icon: '<svg fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zM14 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zM14 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z"/></svg>',
      },
      {
        id: "comercial",
        label: "Comercial",
        icon: '<svg fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/></svg>',
      },
      {
        id: "financeiro",
        label: "Financeiro",
        icon: '<svg fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 17v-2m3 2v-4m3 4v-6m2 10H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"/></svg>',
      },
      {
        id: "vistoria",
        label: "Vistoria",
        icon: '<svg fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"/></svg>',
      },
      {
        id: "engenharia",
        label: "Engenharia",
        icon: '<svg fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z"/></svg>',
      },
    ],

    // --- INICIALIZAÇÃO ---
    async init() {
      this.initSocket();
      await this.loadModule(this.currentTab);
    },

    // --- MOTOR DE COMUNICAÇÃO (WEBSOCKET) ---
    sendCommand(command, payload = {}) {
      if (!this.connected || !window.AurenSocket) {
        this.addNotification(
          "Sem conexão com o motor. Ação não enviada.",
          "error",
        );
        console.error(
          "Attempted to send command without a connection:",
          command,
        );
        return;
      }
      try {
        const message = JSON.stringify({
          action: command, // ✅ CORRIGIDO: usar 'action' para compatibilidade com handler WebSocket
          payload: payload,
        });
        window.AurenSocket.send(message);
        console.log(`[AUREN] Comando enviado: ${command}`, payload);
      } catch (e) {
        this.addNotification(
          "Erro ao codificar mensagem para o motor.",
          "error",
        );
        console.error("Failed to stringify or send message:", e);
      }
    },

    initSocket() {
      const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
      const wsUrl = `${protocol}//${window.location.host}/ws/engine?token=${this.token}`;
      const socket = new WebSocket(wsUrl);

      socket.onopen = () => {
        this.connected = true;
        this.addNotification(
          "Conexão estabelecida com o Motor Semântico",
          "success",
        );
        window.AurenSocket = socket;
      };

      socket.onmessage = (event) => {
        try {
          const msg = JSON.parse(event.data);

          // Tratamento de feedback visual global
          if (msg.type === "error") {
            this.addNotification(msg.payload, "error");
          } else if (msg.type === "status") {
            this.addNotification(msg.payload, "info");
          }

          // Atualização de estatísticas globais (DRE/Projetos)
          if (msg.type === "global_update") {
            this.globalStats = { ...this.globalStats, ...msg.payload };
          }

          // Dispatch para módulos específicos escutarem (Ex: comercial.js)
          window.dispatchEvent(
            new CustomEvent(`socket:${msg.type}`, { detail: msg.payload }),
          );
        } catch (e) {
          console.error("Erro no processamento do frame:", e);
        }
      };

      socket.onclose = () => {
        this.connected = false;
        this.addNotification(
          "Conexão com o servidor perdida. Tentando reconectar...",
          "error",
        );
        setTimeout(() => this.initSocket(), 3000);
      };
    },

    // --- GERENCIAMENTO DE INTERFACE E CARREGAMENTO ---
    async loadModule(tabId) {
      if (this.sections[tabId]) return;

      try {
        const r = await fetch(`./sections/${tabId}.html?token=${this.token}`);
        if (r.ok) {
          this.sections[tabId] = await r.text();
          this.globalStats.lastUpdate = new Date().toLocaleTimeString("pt-BR");
        } else {
          this.sections[tabId] =
            `<div class="p-10 text-red-500 font-mono text-center">⚠️ Módulo [${tabId}] não encontrado no servidor.</div>`;
        }
      } catch (err) {
        console.error(`Erro ao carregar seção ${tabId}:`, err);
        this.addNotification("Falha ao carregar módulo do servidor", "error");
      }
    },

    navigate(tabId) {
      this.currentTab = tabId;
      window.location.hash = tabId;
      this.loadModule(tabId);
    },

    // --- SISTEMA DE NOTIFICAÇÕES (UX) ---
    addNotification(text, type = "info") {
      const id = Date.now();
      this.notifications.push({ id, text, type });

      // Auto-remover após 5 segundos
      setTimeout(() => {
        this.notifications = this.notifications.filter((n) => n.id !== id);
      }, 5000);
    },

    // --- UTILITÁRIOS ---
    formatCurrency(value) {
      return new Intl.NumberFormat("pt-BR", {
        style: "currency",
        currency: "BRL",
      }).format(value);
    },
  };
}
