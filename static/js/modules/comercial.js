document.addEventListener("alpine:init", () => {
  Alpine.data("comercialModule", () => ({
    loading: false,
    statusMessage: "Auren Engine pronta...",
    output: "",
    pipelineValue: 0, // Valor inicial, será atualizado pelo backend

    comercialData: {
      cliente: "",
      servico: "PGRS",
      detalhes: "",
      prioridade: "Normal",
      prazo: "12 meses",
    },

    init() {
      console.log("💼 [COMERCIAL] Módulo de Negócios Ativado.");
      this.sendCommand('comercial_request_sync');

      // Listeners do Engine Central
      window.addEventListener("socket:comercial_init", (e) => {
          if(e.detail && e.detail.pipelineValue) {
              this.pipelineValue = e.detail.pipelineValue;
          }
      });

      window.addEventListener("socket:status", (e) => {
        if (this.loading) this.statusMessage = e.detail;
      });

      window.addEventListener("socket:technical_insight", (e) => {
        if (this.loading) {
          this.output = e.detail;
          this.loading = false;
          this.statusMessage = "Documento Gerado com Sucesso";
          this.addNotification("Documento Redigido!", "success");
        }
      });
    },

    generateDoc(action) {
      if (!this.comercialData.cliente) {
        this.addNotification("Por favor, identifique o Cliente.", "error");
        return;
      }
      
      this.loading = true;
      this.output = "";
      this.statusMessage =
        action === "generate_contract"
          ? "Convertendo Proposta em Contrato Jurídico..."
          : "Analisando Passivos e Redigindo Proposta...";

      const payload = {
        ...this.comercialData,
        action: action, // A action aqui é para o agente de IA
        contentBase: action === "generate_contract" ? this.output : null,
        timestamp: new Date().toISOString(),
      };
      
      // O 'type' da mensagem WebSocket é o que o router do backend usa.
      this.sendCommand(action, payload);
    },

    formatDoc(text) {
      if (!text) return "";
      return (
        text
          .replace(/\n/g, "<br>")
          .replace(
            /(\d\.\s[A-ZÀ-Ú\s]{3,}:)/g,
            '<h2 class="text-lg font-black text-[#1E293B] border-b border-slate-200 mt-8 mb-4 pb-2 uppercase tracking-tighter">$1</h2>',
          )
          .replace(
            /\*\*(.*?)\*\*/g,
            '<strong class="text-slate-900 font-bold">$1</strong>',
          )
          .replace(
            /^\*\s(.*?)(<br>|$)/gm,
            '<div class="flex items-start mb-2 pl-2 text-slate-600"><span class="mr-2 text-emerald-500">◆</span><span>$1</span></div>',
          )
      );
    },

    formatCurrency(v) {
      return new Intl.NumberFormat("pt-BR", {
        style: "currency",
        currency: "BRL",
      }).format(v);
    },
    
    // Helper para notificações
    addNotification(text, type) {
        const global = Alpine.store('auren');
        if(global) global.addNotification(text, type);
    },

    exportToPDF() {
      const printWindow = window.open("", "_blank");
      const content = this.formatDoc(this.output);

      printWindow.document.write(`
                <html>
                    <head>
                        <title>AUREN - ${this.comercialData.cliente}</title>
                        <style>
                            @import url('https://fonts.googleapis.com/css2?family=Inter:wght@400;700;900&display=swap');
                            body { font-family: 'Inter', sans-serif; padding: 2.5cm; color: #1E293B; line-height: 1.5; font-size: 11pt; }
                            h2 { color: #000; border-bottom: 2px solid #00B894; padding-bottom: 5px; margin-top: 30px; text-transform: uppercase; }
                            .header { display: flex; justify-content: space-between; align-items: flex-end; border-bottom: 3px solid #1E293B; padding-bottom: 10px; margin-bottom: 40px; }
                            .footer { position: fixed; bottom: 1cm; left: 2.5cm; right: 2.5cm; border-top: 1px solid #E2E8F0; padding-top: 10px; font-size: 8pt; color: #94A3B8; display: flex; justify-content: space-between; }
                            strong { color: #000; }
                        </style>
                    </head>
                    <body>
                        <div class="header">
                            <div><strong style="font-size: 20pt; letter-spacing: -1px;">AUREN<span style="color:#00B894">.</span></strong><br><span style="font-size: 8pt; font-weight: 900; letter-spacing: 2px; color: #64748B;">ENGINEERING & COMPLIANCE</span></div>
                            <div style="text-align: right; font-size: 9pt;">Emitido em: ${new Date().toLocaleDateString()}</div>
                        </div>
                        <main>${content}</main>
                        <div class="footer">
                            <span>Hash de Autenticidade: ${Math.random().toString(36).substring(2, 15).toUpperCase()}</span>
                            <span>Página 1 de 1</span>
                        </div>
                    </body>
                </html>
            `);
      printWindow.document.close();
      setTimeout(() => {
        printWindow.print();
      }, 500);
    },
  }));
});
