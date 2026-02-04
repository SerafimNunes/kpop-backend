/**
 * Auren Intelligence Engine - v3.1
 * Professional RAG & Multimodal Analysis Module
 */
document.addEventListener('alpine:init', () => {
    Alpine.data('intelligenceEngine', () => ({
        sourceUrl: '',
        loading: false,
        tokenUsage: 0,
        chatInput: '',
        statusMessage: 'Aguardando Mídia Técnica',
        chatHistory: [
            { role: 'assistant', content: 'Olá! Sou o motor de IA da Auren. Posso analisar leis, documentos técnicos ou vídeos para o seu PGRS. Cole uma URL ou suba um arquivo para começar.' }
        ],
        insightData: null,

        init() {
            console.log("🧠 [INTEL-ENGINE] Sistema de Pesquisa Acoplado.");

            // 1. Escuta progresso da IA (Token Usage)
            window.addEventListener('socket:ai_progress', (e) => {
                if (this.loading) {
                    this.tokenUsage = e.detail.percentage || 40;
                    this.statusMessage = e.detail.message || "Processando camadas...";
                }
            });

            // 2. Recebimento do Report Final
            window.addEventListener('socket:technical_insight', (e) => {
                this.insightData = {
                    summary: e.detail.summary || 'Resumo indisponível.',
                    pgrsImpact: e.detail.pgrsImpact || 'Não mapeado.',
                    legalBase: e.detail.legalBase || 'Não citada.',
                    opportunity: e.detail.opportunity || 'Sem observações.'
                };
                this.loading = false;
                this.tokenUsage = 100;
                this.statusMessage = "Análise Concluída";
                
                this.chatHistory.push({ 
                    role: 'assistant', 
                    content: 'A análise técnica foi concluída. O relatório completo está disponível no painel ao lado.' 
                });
                this.scrollToBottom();
            });

            // 3. Chat Contextual
            window.addEventListener('socket:ai_chat_response', (e) => {
                this.chatHistory.push({ role: 'assistant', content: e.detail.content });
                this.scrollToBottom();
            });
        },

        isVideo(url) {
            return url && (url.includes('youtube.com') || url.includes('youtu.be'));
        },

        getEmbedUrl(url) {
            if (!url) return '';
            let videoId = url.includes('v=') ? url.split('v=')[1].split('&')[0] : url.split('youtu.be/')[1];
            return `https://www.youtube.com/embed/${videoId}`;
        },

        async processSource() {
            if (!this.sourceUrl || !window.AurenSocket) return;
            
            this.loading = true;
            this.insightData = null;
            this.tokenUsage = 10;
            this.statusMessage = "Conectando ao núcleo de IA...";

            window.AurenSocket.send(JSON.stringify({
                action: "process_intelligence_source",
                payload: {
                    url: this.sourceUrl,
                    type: this.isVideo(this.sourceUrl) ? "video" : "web_source",
                    timestamp: new Date().toISOString()
                }
            }));
        },

        sendChatMessage() {
            if (!this.chatInput.trim() || this.loading) return;
            
            const q = this.chatInput;
            this.chatHistory.push({ role: 'user', content: q });
            this.chatInput = '';

            window.AurenSocket.send(JSON.stringify({
                action: "ai_chat_query",
                payload: { query: q, context: "intelligence_module" }
            }));
            this.scrollToBottom();
        },

        handleFileUpload(event) {
            const file = event.target.files[0];
            if (!file || !window.AurenSocket) return;

            this.loading = true;
            this.statusMessage = `Enviando ${file.name}...`;
            
            const reader = new FileReader();
            reader.onload = (e) => {
                // Envia cabeçalho do arquivo
                window.AurenSocket.send(JSON.stringify({
                    action: "upload_knowledge",
                    fileName: file.name,
                    fileSize: file.size
                }));
                // Envia binário
                window.AurenSocket.send(e.target.result);
            };
            reader.readAsArrayBuffer(file);
            event.target.value = '';
        },

        scrollToBottom() {
            setTimeout(() => {
                const c = document.getElementById('chatContainer');
                if (c) c.scrollTop = c.scrollHeight;
            }, 100);
        }
    }));
});