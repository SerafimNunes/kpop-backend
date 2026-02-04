document.addEventListener("alpine:init", () => {
  Alpine.data("financeiroModule", () => ({
    init() {
      console.log("[FIN] Financeiro inicializado");
    },
  }));
});
