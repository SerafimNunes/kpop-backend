document.addEventListener("alpine:init", () => {
  Alpine.data("engenhariaModule", () => ({
    // --- STATE ---
    activeTab: "auditoria",
    generating: false,
    statusMessage: "Pronto para iniciar",
    projects: [], // Para listar projetos existentes
    auditData: {
      complianceScore: 0,
      exposureLevel: "CALCULANDO...",
      aiSuggestion: "Aguardando dados...",
      criticalPoints: [
        { id: "Doc", title: "Dados Cadastrais", desc: "CNPJ e Razão Social", status: "PENDENTE", color: "amber" },
        { id: "Inv", title: "Inventário de Resíduos", desc: "Classificação NBR 10.004", status: "CRÍTICO", color: "red" },
        { id: "Field", title: "Evidência de Campo", desc: "Dados da Vistoria Mobile", status: "PENDENTE", color: "amber" },
      ],
    },
    pgrsData: {
      razao: "",
      cnpj: "",
      cnae: "",
      inventory: [],
    },
    aiAuditResult: "",
    aiGeneratedReport: "",

    // --- LIFECYCLE & LISTENERS ---
    init() {
      console.log("🧬 [ENGENHARIA-CORE] Lab Técnico Operacional.");
      this.sendCommand('engenharia_request_sync');
      this.loadPersistedData();
      this.runHeuristicAudit();

      // Listeners do WebSocket
      window.addEventListener("socket:engenharia_init", (e) => {
          if (e.detail && e.detail.projects) {
              this.projects = e.detail.projects;
          }
      });

      window.addEventListener("socket:status", (e) => {
        if (this.generating) this.statusMessage = e.detail;
      });

      window.addEventListener("socket:technical_insight", (e) => {
        if (this.generating) this.aiAuditResult = e.detail;
      });

      window.addEventListener("socket:pgrs_report_ready", (e) => {
        this.aiGeneratedReport = e.detail.replace(/\n/g, "<br>");
        this.statusMessage = "Laudo Finalizado!";
        this.generating = false;
        this.addNotification("Documento Gerado pela IA", "success");
      });

      window.addEventListener("socket:error", (e) => {
        if (this.generating) {
          this.statusMessage = `Erro na IA: ${e.detail}`;
          this.generating = false;
          this.addNotification(e.detail, "error");
        }
      });

      window.addEventListener("auren-sync", () => this.importFromInspection(true));
      this.$watch("pgrsData", (val) => {
        localStorage.setItem("auren_pgrs_draft", JSON.stringify(val));
        this.runHeuristicAudit();
      });
    },

    // --- DATA HANDLING ---
    loadPersistedData() {
      const savedPGRS = localStorage.getItem("auren_pgrs_draft");
      if (savedPGRS) {
        try { this.pgrsData = JSON.parse(savedPGRS); } 
        catch (e) { console.error("Erro no Parse do cache", e); }
      }
    },

    runHeuristicAudit() {
      // ... (lógica de auditoria mantida)
    },

    addInventoryRow() {
      this.pgrsData.inventory.push({ nome: "", classe: "", destino: "" });
    },
    
    removeInventoryRow(index) {
        this.pgrsData.inventory.splice(index, 1);
    },

    importFromInspection(silent = false) {
      // ... (lógica de importação mantida)
    },

    // --- ACTIONS ---
    generateReport() {
      if (!this.pgrsData.razao || this.pgrsData.inventory.length === 0) {
        this.addNotification("Preencha dados da empresa e inventário.", "error");
        return;
      }

      this.generating = true;
      this.aiGeneratedReport = "";
      this.aiAuditResult = "";
      this.statusMessage = "Engine: Processando Camada Técnica...";

      this.sendCommand("generate_pgrs_report", {
        ...this.pgrsData,
        auditScore: this.auditData.complianceScore,
        exposure: this.auditData.exposureLevel,
      });
    },

    // --- HELPERS ---
    addNotification(text, type) {
        const global = Alpine.store('auren');
        if(global) global.addNotification(text, type);
    },
    getScoreColor(score) { /* ... */ },
    getScoreBorder(score) { /* ... */ },
    getScoreBg(score) { /* ... */ },
  }));
});
