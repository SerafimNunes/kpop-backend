function financeiroModule() {
    return {
        isModalOpen: false,
        isTaxModalOpen: false,
        isTaxLoading: false,
        taxAdvice: "",
        currentTxnForAdvice: null,
        searchTerm: "",
        
        newTransaction: {
            description: '',
            amount: null,
            category: 'Operacional',
            type: 'saida',
            status: 'pendente'
        },

        metrics: [
            { id: 'receita', label: 'Receita Bruta (Mês)', value: 0, trend: 0, icon: '📈' },
            { id: 'despesa', label: 'Burn Rate / Despesas', value: 0, trend: 0, icon: '📉' },
            { id: 'impostos', label: 'Provisão de Impostos', value: 0, trend: 0, icon: '⚖️' },
            { id: 'ebitda', label: 'EBITDA Projetado', value: 0, trend: 0, icon: '💎' }
        ],
        
        transactions: [],

        init() {
            console.log("💰 [FINANCEIRO] Engine de Ativos Pronto.");
            this.requestInitialData();

            window.addEventListener('socket:financeiro_init', (e) => {
                const data = e.detail;
                // Corrigido: Backend manda um array, não um objeto mapeado.
                if (data.metrics) {
                    this.metrics[0].value = data.metrics.find(m => m.label === 'Receita Bruta')?.value || 0;
                    this.metrics[1].value = data.metrics.find(m => m.label === 'Despesas')?.value || 0;
                }
                if (data.transactions) {
                    this.transactions = data.transactions.map(tx => ({
                        ...tx, 
                        date: new Date(tx.Date || tx.date).toLocaleDateString('pt-BR') // Aceita ambos os casings
                    }));
                }
            });

            window.addEventListener('socket:financeiro_update', (e) => {
                // Adiciona a nova transação no topo da lista
                const newTx = e.detail;
                this.transactions.unshift({
                    ...newTx,
                    date: new Date(newTx.timestamp).toLocaleDateString('pt-BR')
                });
                this.addNotification("Novo lançamento registrado!", "success");
                this.requestInitialData(); // Re-sincroniza as métricas
            });

            window.addEventListener('socket:tax_advice_ready', (e) => {
                this.taxAdvice = e.detail.advice;
                this.isTaxLoading = false;
            });
        },

        get filteredTransactions() {
            if (!this.searchTerm) return this.transactions;
            return this.transactions.filter(tx => 
                tx.description.toLowerCase().includes(this.searchTerm.toLowerCase()) ||
                tx.category.toLowerCase().includes(this.searchTerm.toLowerCase())
            );
        },

        getTaxAdvice(txn) {
            this.currentTxnForAdvice = txn;
            this.isTaxModalOpen = true;
            this.isTaxLoading = true;
            this.taxAdvice = "";
            // Refatorado para usar sendCommand
            this.sendCommand("analyze_tax_impact", { transactionId: txn.id, amount: txn.amount, category: txn.category });
        },

        requestInitialData() {
            // Refatorado para usar sendCommand
            this.sendCommand("financeiro_request_sync");
        },

        saveTransaction() {
            if (!this.newTransaction.description || !this.newTransaction.amount) {
                this.addNotification("Preencha descrição e valor.", "error");
                return;
            }
            // Refatorado para usar sendCommand
            this.sendCommand("create_transaction", { ...this.newTransaction, timestamp: new Date().toISOString() });
            
            this.closeModal();
            this.addNotification("Lançamento enviado para processamento.", "info");
        },

        formatCurrency(v) {
            return new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' }).format(v);
        },

        // Disponibiliza o helper de notificação no escopo do módulo
        addNotification(text, type) {
            // Acessa a função global do aurenPlatform()
            const global = Alpine.store('auren');
            if(global) global.addNotification(text, type);
        },

        closeModal() { 
            this.isModalOpen = false;
            // Reseta o formulário
            this.newTransaction = { description: '', amount: null, category: 'Operacional', type: 'saida', status: 'pendente' };
        },
        closeTaxModal() { this.isTaxModalOpen = false; }
    }
}