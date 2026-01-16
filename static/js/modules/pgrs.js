/**
 * Auren PGRS Module - v3.0 (Produção Sênior)
 * Compliance: Lei 12.305/2010 | NBR 10.004
 * Barramento de Eventos e Persistência Ativa
 */
document.addEventListener("alpine:init", () => {
  Alpine.data("pgrsModule", () => ({
    generating: false,
    statusMessage: "PGRS pronto.",
    lastSync: null,

    pgrsData: {
      razao: "",
      cnpj: "",
      cnae: "",
      responsavel: "",
      registro: "",
      inventory: [],
    },

    init() {
      // 1. Persistência de Estado
      const saved = localStorage.getItem("auren_pgrs_draft");
      if (saved) this.pgrsData = JSON.parse(saved);

      this.$watch("pgrsData", (val) => {
        localStorage.setItem("auren_pgrs_draft", JSON.stringify(val));
      });

      // 2. Listener de Sincronia Global (Vistoria -> PGRS)
      window.addEventListener("auren-sync", () => {
        console.log("[PGRS-CORE] Sinal de sincronia detectado.");
        this.importFromInspection(true); // silent import
      });

      // 3. Listeners de Barramento Central (WebSocket)
      window.addEventListener("socket:status", (e) => {
        if (this.generating) this.statusMessage = e.detail;
      });

      window.addEventListener("socket:pgrs_report_ready", (e) => {
        this.generating = false;
        alert("✅ Laudo PGRS Gerado e Assinado Digitalmente!");
        console.log("Payload de Saída:", e.detail);
      });

      window.addEventListener("socket:error", (e) => {
        this.generating = false;
        alert("❌ Falha crítica no Core: " + e.detail);
      });
    },

    importFromInspection(silent = false) {
      const raw = localStorage.getItem("last_inspection_data");
      if (!raw) {
        if (!silent) alert("Nenhuma vistoria pendente para este cliente.");
        return;
      }

      try {
        const data = JSON.parse(raw);
        // Transformação de Dados: Vistoria -> Schema PGRS
        const newItems = data.residuos_identificados.map((r) => ({
          nome: r.nome,
          classe: r.classe_sugerida || "IIA",
          destino: r.destino_padrao || "Reciclagem/Aterro",
          quantidade: r.estimativa_mensal || "A apurar",
        }));

        // Merge inteligente (não apaga o que já existe, apenas soma)
        this.pgrsData.inventory = [...this.pgrsData.inventory, ...newItems];
        this.lastSync = new Date().toLocaleTimeString();
        
        if (!silent) alert("Dados de campo importados com sucesso!");
      } catch (e) {
        console.error("Erro na integração de módulos:", e);
      }
    },

    generateReport() {
      // Validação Jurídica Pré-Envio
      if (!this.pgrsData.razao || !this.pgrsData.cnpj || this.pgrsData.inventory.length === 0) {
        alert("Erro de Compliance: Dados cadastrais e inventário são obrigatórios.");
        return;
      }

      if (!window.AurenSocket || window.AurenSocket.readyState !== WebSocket.OPEN) {
        alert("⚠️ Core Engine Offline. Verifique o Barramento Central.");
        return;
      }

      this.generating = true;
      this.statusMessage = "Protocolando documento na IA Auren...";

      window.AurenSocket.send(JSON.stringify({
        action: "generate_pgrs_report",
        payload: { ...this.pgrsData, timestamp: new Date().toISOString() }
      }));
    },

    addInventoryRow() {
      this.pgrsData.inventory.push({ nome: "", classe: "", destino: "", quantidade: "" });
    }
  }));
});