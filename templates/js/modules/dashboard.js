document.addEventListener("alpine:init", () => {
  Alpine.data("dashboardModule", () => ({
    leads: [],
    radarNews: [],
    semaforo: [],
    projetos: [],
    stats: { tempoIA: 0 },

    init() {
      console.log("[DASHBOARD] Iniciando");
      this.sendCommand("dashboard_request");

      window.addEventListener("socket:dashboard_data", (e) => {
        if (e.detail) {
          this.leads = e.detail.leads || [];
          this.radarNews = e.detail.radar || [];
          this.semaforo = e.detail.semaforo || [];
          this.projetos = e.detail.projetos || [];
          this.stats.tempoIA = e.detail.tempoIA || 0;
        }
      });
    },
  }));
});
