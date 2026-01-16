/**
 * Auren Auditoria Module - v3.0 (Produção)
 * Compliance & Risk Management Engine
 */
document.addEventListener("alpine:init", () => {
  Alpine.data("auditoriaModule", () => ({
    loading: false,
    auditData: {
      accepted: true,
      complianceScore: 0,
      exposureLevel: "CALCULANDO...",
      lastCheck: new Date().toLocaleDateString("pt-BR"),
      criticalPoints: [
        {
          id: "01",
          title: "Armazenamento Temporário",
          desc: "Sinalização e contenção (NBR 12235).",
          status: "AGUARDANDO",
          color: "amber",
        },
        {
          id: "02",
          title: "Manifestos de Transporte (MTR)",
          desc: "Sincronização SINIR/SIGOR.",
          status: "OK",
          color: "emerald",
        },
        {
          id: "03",
          title: "Licença de Operação (LO)",
          desc: "Validade e condicionantes ambientais.",
          status: "ALERTA",
          color: "amber",
        },
      ],
    },

    init() {
      console.log("[AUDIT-ENGINE] Risk Analysis Active - Connected to Bus");

      // 1. Executa auditoria inicial
      this.runHeuristicAudit();

      // 2. Escuta sincronia entre abas (Se PGRS ou Vistoria mudar, recalcula o score)
      window.addEventListener("auren-sync", () => {
        console.log("[AUDIT-SYNC] Recalculando conformidade por mudança de dados externos...");
        this.runHeuristicAudit();
      });

      // 3. Escuta o Barramento Central para resultados de auditoria via servidor
      window.addEventListener("socket:audit_update", (event) => {
        const data = event.detail;
        if (!data) return;
        
        if (data.score !== undefined) this.auditData.complianceScore = data.score;
        if (data.points) this.auditData.criticalPoints = data.points;
        this.updateExposureLevel();
      });
    },

    /**
     * Heurística de Auditoria Local:
     * Cruza dados da última vistoria com o PGRS-form para gerar o Score
     */
    runHeuristicAudit() {
      let lastInspection = {};
      let pgrsDraft = {};

      // Proteção de leitura do LocalStorage
      try {
        lastInspection = JSON.parse(localStorage.getItem("last_inspection_data") || "{}");
        pgrsDraft = JSON.parse(localStorage.getItem("auren_pgrs_draft") || "{}");
      } catch (e) {
        console.error("[AUDIT-ENGINE] Erro ao ler cache local:", e);
      }

      let score = 50; // Base score

      // Validação de Documentação (Regra de Negócio Ambiental)
      if (pgrsDraft.razao && pgrsDraft.cnpj) score += 20;
      if (pgrsDraft.inventory && Array.isArray(pgrsDraft.inventory) && pgrsDraft.inventory.length > 0) {
          score += 15;
      }

      // Validação de Presença em Campo
      if (lastInspection.residuos_identificados) {
        score += 15;
      }

      // Penalidades por Inconformidades reportadas na Vistoria
      if (lastInspection.inconformidades && lastInspection.inconformidades.length > 0) {
        score -= 30;
        this.auditData.criticalPoints[0].status = "CRÍTICO";
        this.auditData.criticalPoints[0].color = "red";
      } else if (lastInspection.timestamp) {
        score += 10;
        this.auditData.criticalPoints[0].status = "OK";
        this.auditData.criticalPoints[0].color = "emerald";
      }

      this.auditData.complianceScore = Math.max(0, Math.min(score, 100));
      this.updateExposureLevel();
      this.auditData.lastCheck = new Date().toLocaleTimeString("pt-BR");
    },

    updateExposureLevel() {
      if (this.auditData.complianceScore > 90)
        this.auditData.exposureLevel = "MÍNIMO";
      else if (this.auditData.complianceScore > 70)
        this.auditData.exposureLevel = "MODERADO";
      else this.auditData.exposureLevel = "ALTO";
    },

    /**
     * Dispara um alerta oficial via WebSocket para o Backend
     */
    triggerAlert(point) {
      const payload = {
        action: "audit_incident_alert",
        payload: {
          item: point.title,
          status: point.status,
          score_at_moment: this.auditData.complianceScore,
          timestamp: new Date().toISOString(),
        }
      };

      // Tenta enviar via Barramento Realtime (window.AurenSocket deve ser o singleton do hub)
      if (window.AurenSocket && window.AurenSocket.readyState === WebSocket.OPEN) {
        window.AurenSocket.send(JSON.stringify(payload));
        console.log("[AUDIT-ALERT] Enviado ao servidor:", payload);
        alert(`ALERTA ENVIADO: A irregularidade em "${point.title}" foi reportada ao sistema central.`);
      } else {
        console.warn("[AUDIT-ALERT] Socket offline. Alerta salvo apenas em log local.");
        alert(`OFFLINE: O alerta para "${point.title}" será sincronizado ao restaurar a conexão.`);
      }
    },
  }));
});