document.addEventListener("alpine:init", () => {
  Alpine.data("vistoriaModule", () => ({
    // --- STATE ---
    generating: false,
    progress: 0,
    statusMessage: "Aguardando Evidência",
    pgrsData: { setor: "", observacoes: "" },
    checklist: { segregacao: false, armazenamento: false, identificacao: false, vazamento: false },
    previousInspections: [], // Novo estado para vistorias anteriores
    
    // --- STATIC DATA ---
    checklistItems: {
      segregacao: "Segregação (NBR 10004)",
      armazenamento: "Área Coberta/Impermeável",
      identificacao: "Rotulagem GHS/NBR 7500",
      vazamento: "Kits de Contenção"
    },

    // --- LIFECYCLE & LISTENERS ---
    init() {
      console.log("🛠️ [VISTORIA-CORE] Operacional.");
      this.createFileInput();
      this.sendCommand('vistoria_request_sync'); // Solicita dados iniciais

      // Listener para dados iniciais
      window.addEventListener("socket:vistoria_init", (e) => {
        if (e.detail && e.detail.inspections) {
          this.previousInspections = e.detail.inspections;
        }
      });
      
      // Listener para resultado final da consolidação
      window.addEventListener("socket:inspection_analysis_result", (e) => {
        this.generating = false;
        this.progress = 100;
        this.statusMessage = "Auditoria Consolidada!";
        this.pgrsData.observacoes = e.detail.observacoes_consolidadas;
        localStorage.setItem("last_inspection_data", JSON.stringify(e.detail));
        this.addNotification("Auditoria Digital Consolidada com Sucesso!", "success");
        this.sendCommand('vistoria_request_sync'); // Re-sincroniza a lista
      });

      // Listener para insights de mídia (análise de imagem/vídeo)
      window.addEventListener("socket:technical_insight", (e) => {
        this.statusMessage = "Análise de Mídia Concluída.";
        this.pgrsData.observacoes += `\n[IA Media Insight]: ${e.detail}`;
        this.progress = 100;
        setTimeout(() => { if(!this.generating) this.progress = 0; }, 2000);
      });

      // Listener para atualizações de status durante o processamento
      window.addEventListener("socket:status", (e) => {
        if(this.generating) {
          this.statusMessage = e.detail;
          this.progress = Math.min(95, this.progress + 5);
        }
      });
      
      // Listener para erros
      window.addEventListener("socket:error", (e) => {
        this.generating = false;
        this.progress = 0;
        this.statusMessage = "Erro na Operação";
        this.addNotification(e.detail, "error");
      });
    },

    // --- UI & FILE HANDLING ---
    createFileInput() {
      if (document.getElementById("media-capture-input")) return;
      const input = document.createElement("input");
      input.type = "file";
      input.id = "media-capture-input";
      input.className = "hidden";
      input.addEventListener("change", (e) => this.handleFileUpload(e));
      document.body.appendChild(input);
    },

    handleMediaDispatch(type) {
      if (!this.pgrsData.setor) {
        this.addNotification("Informe o Setor antes de enviar a evidência.", "error");
        return;
      }
      document.getElementById("media-capture-input").accept = type;
      document.getElementById("media-capture-input").click();
    },

    handleFileUpload(e) {
      const file = e.target.files[0];
      if (!file) return;

      this.generating = true;
      this.progress = 10;
      this.statusMessage = "Processando arquivo...";

      // 1. Envia o cabeçalho JSON via sendCommand
      this.sendCommand("media_capture_start", {
        mimeType: file.type,
        setor: this.pgrsData.setor,
        fileName: file.name
      });
      
      // 2. Envia o binário diretamente pelo socket
      const reader = new FileReader();
      reader.onload = (event) => {
        if (window.AurenSocket && window.AurenSocket.readyState === WebSocket.OPEN) {
          window.AurenSocket.send(event.target.result);
          this.statusMessage = "Analisando evidência com Visão Computacional...";
          this.progress = 40;
        } else {
          this.addNotification("Conexão perdida ao enviar arquivo.", "error");
          this.generating = false;
        }
      };
      reader.readAsArrayBuffer(file);
      e.target.value = ""; // Reseta o input
    },

    // --- ACTIONS ---
    consolidateAnalysis() {
      if (!this.pgrsData.setor) {
          this.addNotification("Informe o Setor ou Ponto de Geração.", "error");
          return;
      }
      
      this.generating = true;
      this.progress = 10;
      this.statusMessage = "Consolidando Laudo Técnico...";
      
      this.sendCommand("consolidate_inspection", {
        ...this.pgrsData,
        checklist: this.checklist,
        timestamp: new Date().toISOString()
      });
    },
    
    // --- HELPERS ---
    addNotification(text, type) {
      const global = Alpine.store('auren');
      if (global) global.addNotification(text, type);
    }
  }));
});