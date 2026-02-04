document.addEventListener("alpine:init", () => {
  Alpine.data("vistoriaModule", () => ({
    init() {
      console.log("[VIST] Vistoria inicializada");
    },
  }));
});
