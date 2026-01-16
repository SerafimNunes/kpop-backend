document.addEventListener("alpine:init", () => {
  Alpine.data("vistoriaModule", () => ({
    generating: false,
    statusMessage: "",
    pgrsData: { setor: "", observacoes: "" },
    checklist: { segregacao: false, armazenamento: false, identificacao: false, vazamento: false },
    checklistItems: {
      segregacao: "Segregação na Fonte (NBR 10004)",
      armazenamento: "Armazenamento Coberto/Impermeável",
      identificacao: "Rotulagem (GHS/NBR 7500)",
      vazamento: "Kits de Mitigação/Contenção"
    },

    init() {
      console.log("[VISTORIA-CORE] Ready.");
      this.createFileInput();

      window.addEventListener("capture-photo", () => this.handleMediaDispatch("image/*"));
      window.addEventListener("record-video", () => this.handleMediaDispatch("video/*"));
      window.addEventListener("start-audio", () => this.handleMediaDispatch("audio/*"));

      window.addEventListener("socket:inspection_analysis_result", (event) => {
        localStorage.setItem("last_inspection_data", JSON.stringify(event.detail));
        window.dispatchEvent(new CustomEvent("auren-sync"));
        this.statusMessage = "Análise concluída. Dados enviados ao PGRS.";
        this.generating = false;
        alert("Vistoria finalizada com sucesso!");
      });

      // Listener for status messages from backend
      window.addEventListener("socket:status", (event) => {
        this.statusMessage = event.detail;
      });
      window.addEventListener("socket:error", (event) => {
        this.statusMessage = `Erro: ${event.detail}`;
        this.generating = false;
        alert(`Erro do Backend: ${event.detail}`);
      });
       window.addEventListener("socket:technical_insight", (event) => {
        this.generating = false;
        this.statusMessage = "Evidência analisada pela IA.";
      });
    },

    createFileInput() {
      const input = document.createElement("input");
      input.type = "file";
      input.style.display = "none";
      input.id = "media-capture-input";
      input.addEventListener("change", (event) => this.handleFileUpload(event));
      document.body.appendChild(input);
    },

    handleMediaDispatch(acceptType) {
      if (!window.AurenSocket || window.AurenSocket.readyState !== WebSocket.OPEN) {
        alert("⚠️ Conexão perdida com o Core.");
        return;
      }
      const fileInput = document.getElementById("media-capture-input");
      fileInput.accept = acceptType;
      fileInput.click();
    },

    handleFileUpload(event) {
      const file = event.target.files[0];
      if (!file) return;

      this.generating = true;
      this.statusMessage = `Enviando ${file.type}...`;

      const reader = new FileReader();
      reader.onload = (e) => {
        const fileData = e.target.result;
        
        // 1. Send metadata message
        window.AurenSocket.send(JSON.stringify({
          action: "media_capture_start",
          vistoriaID: 1, // Hardcoded for now, should be dynamic
          mimeType: file.type
        }));

        // 2. Send binary data
        window.AurenSocket.send(fileData);
      };
      reader.readAsArrayBuffer(file);
      
      // Reset file input value to allow capturing the same file again
      event.target.value = "";
    },

    async consolidateAnalysis() {
      if (!this.pgrsData.setor || this.pgrsData.observacoes.length < 10) {
        alert("Obrigatório: Setor e Relato Técnico detalhado.");
        return;
      }

      this.generating = true;
      this.statusMessage = "IA Auren auditando evidências...";

      window.AurenSocket.send(JSON.stringify({
        action: "consolidate_inspection",
        payload: {
          setor: this.pgrsData.setor,
          observacoes: this.pgrsData.observacoes,
          checklist: this.checklist,
          timestamp: new Date().toISOString()
        }
      }));
    }
  }));
});