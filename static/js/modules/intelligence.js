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
            console.log("[INTEL-ENGINE] Sênior Mode Active");

            // 1. Escuta o progresso real via WebSocket
            window.addEventListener('socket:ai_progress', (event) => {
                if (this.loading) {
                    this.tokenUsage = event.detail.percentage || 0;
                    this.statusMessage = event.detail.message || "Analisando contexto...";
                }
            });

            // 2. Entrega do Insight Técnico (Onde o resumo e leis aparecem)
            window.addEventListener('socket:technical_insight', (event) => {
                console.log("[INTEL-ENGINE] Insight Received:", event.detail);
                this.insightData = {
                    summary: event.detail.summary || 'Resumo não disponível.',
                    pgrsImpact: event.detail.pgrsImpact || 'Impacto não calculado.',
                    legalBase: event.detail.legalBase || 'Base legal não identificada.',
                    opportunity: event.detail.opportunity || 'Sem oportunidades mapeadas no momento.'
                };
                this.loading = false;
                this.tokenUsage = 100;
                this.statusMessage = "Análise Concluída";

                this.chatHistory.push({ 
                    role: 'assistant', 
                    content: 'A análise técnica foi concluída. Você pode conferir o relatório ao lado ou me fazer perguntas específicas sobre o conteúdo.' 
                });
                this.scrollToBottom();
            });

            // 3. Respostas do Chat
            window.addEventListener('socket:ai_chat_response', (event) => {
                this.chatHistory.push({ 
                    role: 'assistant', 
                    content: event.detail.content 
                });
                this.scrollToBottom();
            });

            // 4. Tratamento de Erro Global
            window.addEventListener('socket:error', (event) => {
                this.loading = false;
                this.tokenUsage = 0;
                alert("Erro no Motor de IA: " + event.detail);
            });
        },

        isVideo(url) {
            if (!url) return false;
            return url.includes('youtube.com') || url.includes('youtu.be');
        },

        getEmbedUrl(url) {
            if (!url) return '';
            let videoId = '';
            try {
                if (url.includes('v=')) {
                    videoId = url.split('v=')[1].split('&')[0];
                } else if (url.includes('youtu.be/')) {
                    videoId = url.split('youtu.be/')[1];
                }
                return `https://www.youtube.com/embed/${videoId}`;
            } catch (e) {
                return '';
            }
        },

        async processSource() {
            if (!this.sourceUrl) return;

            if (!window.AurenSocket || window.AurenSocket.readyState !== WebSocket.OPEN) {
                return alert("⚠️ Conexão com o Barramento Auren perdida. Recarregue a página.");
            }
            
            this.loading = true;
            this.tokenUsage = 5; // Início visual
            this.insightData = null;
            this.statusMessage = "Iniciando captura de dados...";

            window.AurenSocket.send(JSON.stringify({
                action: "process_intelligence_source",
                payload: {
                    url: this.sourceUrl,
                    type: this.isVideo(this.sourceUrl) ? "video" : "document",
                    timestamp: new Date().toISOString()
                }
            }));
        },

        sendChatMessage() {
            if (!this.chatInput.trim() || this.loading) return;

            const userQuery = this.chatInput;
            this.chatHistory.push({ role: 'user', content: userQuery });
            this.chatInput = '';

            window.AurenSocket.send(JSON.stringify({
                action: "ai_chat_query",
                payload: {
                    query: userQuery,
                    context: this.insightData ? "document_analysis" : "general_compliance"
                }
            }));

            this.scrollToBottom();
        },

        handleFileUpload(event) {
            const file = event.target.files[0];
            if (!file) return;

            if (!window.AurenSocket || window.AurenSocket.readyState !== WebSocket.OPEN) {
                return alert("⚠️ Conexão com o Barramento Auren perdida. Recarregue a página.");
            }

            this.loading = true;
            this.insightData = null;
            this.statusMessage = `Enviando ${file.name}...`;
            this.tokenUsage = 5;

            const reader = new FileReader();
            reader.onload = (e) => {
                // 1. Enviar metadados primeiro
                window.AurenSocket.send(JSON.stringify({
                    action: "upload_knowledge",
                    fileName: file.name
                }));

                // 2. Enviar o conteúdo binário do arquivo
                window.AurenSocket.send(e.target.result);
            };
            
            reader.onerror = (e) => {
                this.loading = false;
                this.statusMessage = "Falha ao ler o arquivo.";
                console.error("FileReader error:", e);
                alert("Não foi possível ler o arquivo selecionado.");
            };

            reader.readAsArrayBuffer(file);
            
            // Limpa o valor do input para permitir o upload do mesmo arquivo novamente
            event.target.value = '';
        },

        scrollToBottom() {
            setTimeout(() => {
                const container = document.getElementById('chatContainer');
                if (container) container.scrollTop = container.scrollHeight;
            }, 100);
        }
    }));
});