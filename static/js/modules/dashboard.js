document.addEventListener("alpine:init", () => {
  Alpine.data("dashboardModule", () => ({
    // --- ESTADO LOCAL (será preenchido pelo backend) ---
    stats: { tempoIA: 0 },
    semaforo: [],
    leads: [],
    radarNews: [],
    projetos: [],

    init() {
        console.log("🚀 Dashboard Module: Online");

        // Solicita os dados iniciais ao motor via WebSocket
        // A função sendCommand está no escopo global do Alpine, em auren-core.js
        this.sendCommand('dashboard_request_sync');

        // --- LISTENERS DE SOCKET ---

        // Listener para receber os dados iniciais do dashboard
        window.addEventListener("socket:dashboard_init", (e) => {
            console.log("Dashboard data received:", e.detail);
            const data = e.detail;
            this.semaforo = data.semaforo || [];
            this.leads = data.leads || [];
            this.radarNews = data.radarNews || [];
            this.projetos = data.projetos || [];
            if (data.stats) {
                this.stats = { ...this.stats, ...data.stats };
            }
        });
        
        // Listeners para atualizações em tempo real
        window.addEventListener("socket:HUNTER_UPDATE", (e) => {
            if (e.detail && e.detail.empresa) {
                this.leads.unshift(e.detail);
                if (this.leads.length > 5) this.leads.pop();
                this.addNotification(`Hunter: Novo Lead Detectado (${e.detail.empresa})`, 'success');
            }
        });

        window.addEventListener("socket:RADAR_ALERT", (e) => {
            this.radarNews.unshift(e.detail);
            if (this.radarNews.length > 5) this.radarNews.pop();
            this.addNotification("Radar: Nova atualização legislativa detectada", 'info');
        });

        window.addEventListener("socket:TELEMETRY", (e) => {
            if (e.detail && e.detail.latency) {
                this.stats.tempoIA = e.detail.latency;
            }
        });
    },
  }));
});
