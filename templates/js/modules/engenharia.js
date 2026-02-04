document.addEventListener("alpine:init", () => {
  Alpine.data("engenhariaModule", () => ({
    init() {
      console.log("[ENG] Engenharia inicializada");
    },
  }));
});
