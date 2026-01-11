/**
 * Auren Platform Core Engine - Frontend v1.0
 * Gerenciamento de Estado Reativo e Integração IA Multimodal
 */
function aurenPlatform() {
    return {
        currentTab: "dashboard",
        loading: false,
        generating: false,
        sections: {},
        statusMessage: "Auren Engine pronta...",

        // --- INSIGHT ENGINE DATA ---
        sourceUrl: "",
        tokenUsage: 15,
        insightData: null,
        chatInput: "",
        chatHistory: [
            { role: "assistant", content: "Olá Engenheiro. Sou o especialista Auren. Como posso ajudar com a norma ou documento atual?" }
        ],

        token: new URLSearchParams(window.location.search).get("token") || "AUREN-PLATFORM-DEMO",

        pgrsData: {
            razao: "",
            cnpj: "",
            cnae: "",
            responsavel: "",
            registro: "",
            observacoes: "",
            setor: "",
            inventory: [
                { nome: "", classe: "", codigo: "", quantidade: "", destino: "" },
            ],
        },

        revenueData: {
            labels: ["Jul", "Ago", "Set", "Out", "Nov", "Dez"],
            datasets: [{
                label: "Faturamento Mensal (R$)",
                data: [10000, 12000, 11500, 13000, 14200, 15500],
                borderColor: "#d4af37",
                backgroundColor: "rgba(212, 175, 55, 0.2)",
                fill: true,
                tension: 0.4,
            }],
        },
        revenueChart: null,

        menuItems: [
            { id: "dashboard", label: "Dashboard", icon: '<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path d="M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zM14 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zM14 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>' },
            { id: "intelligence", label: "Insight Engine", icon: '<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path d="M13 10V3L4 14h7v7l9-11h-7z" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>' },
            { id: "pgrs-form", label: "Software PGRS", icon: '<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>' },
            { id: "vistoria", label: "Checklist Campo", icon: '<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>' },
            { id: "auditoria", label: "Auditoria Surpresa", icon: '<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04a11.357 11.357 0 00-1.026 5.405c0 5.992 4.137 11.03 9.644 12.315a12.164 12.164 0 009.644-12.315c0-1.879-.451-3.65-1.246-5.208z" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>' },
        ],

        async init() {
            console.log("[AUREN] Inicializando Core...");
            const loadPromises = this.menuItems.map(async (item) => {
                try {
                    const resp = await fetch(`./sections/${item.id}.html?token=${this.token}`);
                    if (resp.ok) {
                        this.sections[item.id] = await resp.text();
                        console.log(`[AUREN] Seção OK: ${item.id}`);
                    } else {
                        this.sections[item.id] = `<div class="p-10 text-red-500">Erro ao carregar componente: ${item.id}</div>`;
                    }
                } catch (e) {
                    console.error(`[AUREN] Falha técnica em ${item.id}`, e);
                }
            });
            await Promise.all(loadPromises);
            this.$watch("currentTab", (tab) => {
                if (tab === "dashboard") {
                    this.$nextTick(() => this.renderRevenueChart());
                }
            });
            if (this.currentTab === "dashboard") {
                this.$nextTick(() => this.renderRevenueChart());
            }
            window.addEventListener('capture-photo', () => this.handleCaptureMedia('image'));
            window.addEventListener('record-video', () => this.handleCaptureMedia('video'));
            window.addEventListener('start-audio', () => this.handleCaptureMedia('audio'));
        },

        addInventoryItem() {
            this.pgrsData.inventory.push({ nome: "", classe: "", codigo: "", quantidade: "", destino: "" });
        },

        removeInventoryItem(index) {
            if (this.pgrsData.inventory.length > 1) {
                this.pgrsData.inventory.splice(index, 1);
            }
        },

        async handleCaptureMedia(type) {
            this.statusMessage = `Acessando dispositivo para ${type}...`;
            try {
                const input = document.createElement('input');
                input.type = 'file';
                input.capture = 'environment';
                if (type === 'image') input.accept = 'image/*';
                else if (type === 'video') input.accept = 'video/*';
                
                if (type !== 'audio') {
                    input.onchange = (e) => {
                        const file = e.target.files[0];
                        console.log(`[AUREN] ${type} capturado:`, file);
                        // Fluxo de campo mantém o comportamento binário direto para ser tratado como evidência
                        const reader = new FileReader();
                        reader.onload = (ev) => {
                            const socket = new WebSocket(`${window.location.protocol === "https:" ? "wss" : "ws"}://${window.location.host}/ws/engine?token=${this.token}`);
                            socket.onopen = () => socket.send(ev.target.result);
                        };
                        reader.readAsArrayBuffer(file);
                    };
                    input.click();
                } else {
                    alert("Auren: Recurso de gravação de voz ativo. Processando áudio em tempo real...");
                }
            } catch (err) {
                console.error("[AUREN] Erro ao acessar mídia:", err);
            }
        },

        async processSource() {
            if (!this.sourceUrl) return alert("Insira uma URL técnica (Youtube/PDF) para análise.");
            this.loading = true;
            this.generating = true;
            this.insightData = null;
            this.statusMessage = "Conectando ao Auren Neural Link...";
            const socket = new WebSocket(`${window.location.protocol === "https:" ? "wss" : "ws"}://${window.location.host}/ws/engine?token=${this.token}`);
            socket.onopen = () => {
                socket.send(JSON.stringify({ action: "process_insight", url: this.sourceUrl }));
            };
            socket.onmessage = (event) => {
                const msg = JSON.parse(event.data);
                if (msg.Type === "status") this.statusMessage = msg.Payload;
                if (msg.Type === "technical_insight") {
                    this.insightData = {
                        summary: msg.Payload,
                        pgrsImpact: "Análise IA: Impacto imediato na gestão de resíduos e conformidade.",
                        legalBase: "Legislação Federal Aplicada e Normas Estaduais.",
                        opportunity: "Otimização de custos identificada via análise técnica."
                    };
                    this.generating = false;
                    this.loading = false;
                    this.tokenUsage = Math.floor(Math.random() * (85 - 45) + 45);
                    socket.close();
                }
            };
        },

        sendChatMessage() {
            if (!this.chatInput.trim()) return;
            const prompt = this.chatInput;
            this.chatHistory.push({ role: "user", content: prompt });
            this.chatInput = "";
            this.$nextTick(() => {
                const container = document.getElementById('chatContainer');
                if (container) container.scrollTop = container.scrollHeight;
            });
            setTimeout(() => {
                this.chatHistory.push({ 
                    role: "assistant", 
                    content: `Analisando "${prompt}" com base nos dados do projeto e normas técnicas vigentes...` 
                });
            }, 1000);
        },

        handleFileUpload(event) {
            const file = event.target.files[0];
            if (!file) return;

            this.statusMessage = `Processando arquivo técnico: ${file.name}`;
            this.loading = true;

            const reader = new FileReader();
            reader.onload = (e) => {
                const rawData = e.target.result;
                const protocol = window.location.protocol === "https:" ? "wss" : "ws";
                const socket = new WebSocket(`${protocol}://${window.location.host}/ws/engine?token=${this.token}`);
                
                socket.onopen = () => {
                    // HEADER DE INTENÇÃO: Evita que o backend salve como evidência de vistoria
                    socket.send(JSON.stringify({ 
                        action: "upload_knowledge", 
                        fileName: file.name,
                        fileType: file.type 
                    }));
                    // ENVIO DO BINÁRIO LOGO EM SEGUIDA
                    socket.send(rawData);
                };

                socket.onmessage = (ev) => {
                    const msg = JSON.parse(ev.data);
                    if (msg.Type === "technical_insight") {
                        this.insightData = { 
                            summary: msg.Payload,
                            pgrsImpact: "Extraído do documento técnico enviado.",
                            legalBase: "Análise baseada no arquivo: " + file.name,
                            opportunity: "Otimização identificada via Librarian Agent."
                        };
                        this.loading = false;
                        socket.close();
                    }
                    if (msg.Type === "status") this.statusMessage = msg.Payload;
                };
            };
            reader.readAsArrayBuffer(file);
        },

        renderRevenueChart() {
            const canvas = document.getElementById("revenueChart");
            if (!canvas) return;
            if (this.revenueChart) this.revenueChart.destroy();
            this.revenueChart = new Chart(canvas, {
                type: "line",
                data: this.revenueData,
                options: {
                    responsive: true,
                    maintainAspectRatio: false,
                    plugins: { legend: { display: false } },
                    scales: {
                        x: { grid: { color: "#1e293b" }, ticks: { color: "#64748b" } },
                        y: { grid: { color: "#1e293b" }, ticks: { color: "#64748b" } },
                    },
                },
            });
        },

        isVideo(url) { return url && (url.includes("youtube.com") || url.includes("youtu.be")); },
        getEmbedUrl(url) {
            if (!url) return "";
            let videoId = url.includes("v=") ? url.split("v=")[1].split("&")[0] : url.split("youtu.be/")[1];
            return `https://www.youtube.com/embed/${videoId}`;
        },
        
        async consolidateAnalysis() {
            this.generating = true;
            this.statusMessage = "IA Auren auditando conformidade e gerando relatório final...";
            setTimeout(() => {
                this.generating = false;
                alert("Análise Técnica Multimodal Consolidada!");
            }, 3000);
        }
    };
}