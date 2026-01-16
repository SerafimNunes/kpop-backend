/**
 * Auren Dashboard Module - v2.0
 * Gerenciamento de KPIs, Gráficos de Faturamento e Status Operacional
 */
document.addEventListener('alpine:init', () => {
    Alpine.data('dashboardModule', () => ({
        revenueChart: null,
        
        // Mock de dados operacionais (Pode ser substituído por fetch futuro)
        stats: {
            pgrsEmitidos: 24,
            metaMensal: 30,
            tempoIA: 4.2,
            aguardandoRT: 7,
            receitaMensal: 14200,
            receitaPendente: 3800
        },

        // Dados do gráfico de faturamento
        revenueData: {
            labels: ["Jul", "Ago", "Set", "Out", "Nov", "Dez"],
            datasets: [{
                label: "Faturamento Mensal (R$)",
                data: [10000, 12000, 11500, 13000, 14200, 15500],
                borderColor: "#d4af37",
                backgroundColor: "rgba(212, 175, 55, 0.1)",
                fill: true,
                tension: 0.4,
                borderWidth: 3,
                pointRadius: 4,
                pointBackgroundColor: "#d4af37"
            }]
        },

        /**
         * Inicialização do componente
         */
        init() {
            // Pequeno delay para garantir que o canvas foi injetado pelo auren-core
            setTimeout(() => {
                this.renderRevenueChart();
            }, 100);
        },

        /**
         * Renderização do Chart.js
         */
        renderRevenueChart() {
            const canvas = document.getElementById("revenueChart");
            if (!canvas) return;

            if (this.revenueChart) this.revenueChart.destroy();

            const ctx = canvas.getContext('2d');
            this.revenueChart = new Chart(ctx, {
                type: "line",
                data: this.revenueData,
                options: {
                    responsive: true,
                    maintainAspectRatio: false,
                    plugins: { 
                        legend: { display: false },
                        tooltip: {
                            backgroundColor: '#161b22',
                            titleColor: '#d4af37',
                            bodyColor: '#fff',
                            borderColor: '#30363d',
                            borderWidth: 1
                        }
                    },
                    scales: {
                        y: {
                            grid: { color: "rgba(255,255,255,0.05)", drawBorder: false },
                            ticks: { color: "#71717a", font: { size: 10 } },
                        },
                        x: { 
                            grid: { display: false }, 
                            ticks: { color: "#71717a", font: { size: 10 } } 
                        },
                    },
                },
            });
        }
    }));
});