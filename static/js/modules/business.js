/**
 * Auren Business Module - v3.8 (Gold Stability Edition)
 * Foco: Sincronização Total, Estética de Engenharia e Correção de Estouro de Páginas
 */
document.addEventListener("alpine:init", () => {
  Alpine.data("businessModule", () => ({
    loading: false,
    statusMessage: "Auren Engine pronta...",
    businessOutput: "",
    businessData: {
      cliente: "",
      servico: "PGRS",
      detalhes: "",
    },

    /**
     * Inicialização e Acoplamento ao Barramento Central (Socket)
     */
    init() {
      // 1. Escuta atualizações de status enviadas pelo Core
      window.addEventListener("socket:status", (event) => {
        if (this.loading) this.statusMessage = event.detail;
      });

      // 2. Escuta o resultado técnico final da IA
      window.addEventListener("socket:technical_insight", (event) => {
        if (this.loading) {
          console.log("[Business] Recebendo documento. Aplicando higienização e formatação...");
          this.businessOutput = event.detail;
          this.loading = false;
          this.statusMessage = "Documento redigido com sucesso.";
        }
      });

      // 3. Escuta erros globais do servidor
      window.addEventListener("socket:error", (event) => {
        if (this.loading) {
          alert("Erro na IA: " + event.detail);
          this.loading = false;
          this.statusMessage = "Falha na geração.";
        }
      });

      console.log("[Business] Módulo acoplado com sucesso e pronto para operação.");
    },

    /**
     * Orquestração de Geração de Documentos via WebSocket
     */
    async generateBusinessDoc(action) {
      if (!this.businessData.cliente && action === "generate_proposal") {
        return alert("Erro: Identificação do Cliente é obrigatória.");
      }

      if (!window.AurenSocket || window.AurenSocket.readyState !== WebSocket.OPEN) {
        return alert("⚠️ Sistema Offline. Verifique sua conexão com o servidor Go.");
      }

      this.loading = true;
      this.businessOutput = ""; // Limpa visualização anterior
      this.statusMessage = action === "generate_contract" 
          ? "IA Auren transformando proposta em contrato jurídico..." 
          : "IA Auren redigindo documento técnico de engenharia...";

      window.AurenSocket.send(
        JSON.stringify({
          action: action,
          cliente: this.businessData.cliente,
          servico: this.businessData.servico,
          detalhes: this.businessData.detalhes,
          propostaAtual: this.businessOutput,
          timestamp: new Date().toISOString()
        })
      );
    },

    /**
     * Formatador Markdown Seguro (Resolve erro de tabelas e estouro de página)
     */
    formatDoc(text) {
      if (!text) return "";

      // Remove sequências de caracteres de separação que podem quebrar o PDF
      let cleanedText = text.replace(/-{10,}/g, '');

      return cleanedText
        .replace(/\n/g, "<br>")
        // Títulos de Seção (Ex: 1. OBJETIVO)
        .replace(/(\d\.\s[A-ZÀ-Ú\s]{3,}:)/g, '<h2 class="text-xl font-black text-slate-900 border-b-2 border-amber-500 mt-10 mb-4 pb-1 uppercase tracking-tighter">$1</h2>')
        // Subtítulos de Fases (Ex: Fase 1:)
        .replace(/(Fase\s\d+:.*?)(<br>|$)/g, '<h3 class="text-md font-bold text-amber-600 mt-4 mb-2">$1</h3>')
        // Negritos de ênfase
        .replace(/\*\*(.*?)\*\*/g, '<strong class="text-slate-900 font-extrabold">$1</strong>')
        // Listas Técnicas com ícones
        .replace(/^\*\s(.*?)(<br>|$)/gm, '<div class="flex items-start mb-2 pl-4 text-slate-700"><span class="mr-3 text-amber-500">◈</span><span>$1</span></div>')
        // Renderizador de Tabelas Seguro (MD Pipes para HTML Grid)
        .replace(/^\|(.+)\|$/gm, (match) => {
           // Ignora linhas de separação Markdown que causam estouro de página
           if (match.includes('---') || match.includes(':--')) return "";
           
           const cells = match.split('|').filter(c => c.trim() !== "");
           if (cells.length === 0) return "";
           
           return `<div class="grid grid-cols-${cells.length} border-b border-slate-200 py-2 px-4 bg-white hover:bg-slate-50">
             ${cells.map(c => `<span class="text-[11px] font-medium text-slate-700">${c.trim()}</span>`).join('')}
           </div>`;
        });
    },

    /**
     * Exportador PDF de Alta Fidelidade (Marca d'água e Layout de Engenharia)
     */
    exportToPDF() {
      if (!this.businessOutput) return;
      
      const content = this.formatDoc(this.businessOutput);
      const cliente = this.businessData.cliente || "Proposta_Tecnica_Auren";
      const printWindow = window.open("", "_blank");

      const goldStyle = `
        <style>
          @import url('https://fonts.googleapis.com/css2?family=Cinzel:wght@700&family=Inter:wght@400;600;800&display=swap');
          * { box-sizing: border-box; -webkit-print-color-adjust: exact; }
          body { 
            font-family: 'Inter', sans-serif; 
            color: #1a202c; 
            line-height: 1.7; 
            margin: 0; padding: 0; 
          }
          .page { padding: 2.5cm; position: relative; min-height: 100vh; }
          .watermark { 
            position: fixed; top: 50%; left: 50%; transform: translate(-50%, -50%) rotate(-45deg);
            font-size: 70px; color: rgba(226, 232, 240, 0.3); font-weight: 900; z-index: -1; 
            pointer-events: none; white-space: nowrap; text-transform: uppercase;
          }
          h1.brand { font-family: 'Cinzel', serif; font-size: 30px; letter-spacing: 3px; margin: 0; font-weight: 700; }
          h2 { page-break-after: avoid; }
          .grid { display: grid; gap: 5px; }
          .footer { 
            position: fixed; bottom: 1cm; left: 2.5cm; right: 2.5cm; 
            font-size: 9px; color: #94a3b8; border-top: 1px solid #e2e8f0; 
            padding-top: 10px; text-align: center;
          }
          @media print {
            .page { padding: 0; margin: 0; }
            .no-print { display: none; }
            @page { margin: 2cm; size: auto; }
          }
        </style>
      `;

      printWindow.document.write(`
        <html>
          <head>
            <title>AUREN - ${cliente}</title>
            ${goldStyle}
          </head>
          <body>
            <div class="page">
              <div class="watermark">Auren Consultoria</div>
              <header style="display: flex; justify-content: space-between; align-items: center; border-bottom: 4px solid #1a202c; padding-bottom: 20px; margin-bottom: 40px;">
                <div>
                  <h1 class="brand">AUREN<span style="color:#d4af37">.</span></h1>
                  <p style="font-size: 9px; text-transform: uppercase; font-weight: 800; color: #64748b; margin: 0; letter-spacing: 1px;">Environmental Engineering & Intelligence</p>
                </div>
                <div style="text-align: right; line-height: 1.2;">
                  <p style="font-size: 10px; font-weight: 900; margin: 0;">PROPOSTA TÉCNICA Nº ${Math.random().toString(36).substring(7).toUpperCase()}</p>
                  <p style="font-size: 9px; color: #64748b; margin: 5px 0 0 0;">EMISSÃO: ${new Date().toLocaleDateString("pt-BR")}</p>
                </div>
              </header>
              <main style="font-size: 13px;">${content}</main>
              <footer class="footer">
                Auren Platform ® - Sistema de Inteligência em Engenharia Ambiental. <br>
                Este documento é um instrumento técnico confidencial sob responsabilidade da Auren Consultoria.
              </footer>
            </div>
          </body>
        </html>
      `);

      printWindow.document.close();
      
      printWindow.onload = () => { 
        setTimeout(() => { 
          printWindow.print(); 
          // printWindow.close(); // Opcional: fechar após imprimir
        }, 600); 
      };
    }
  }));
});