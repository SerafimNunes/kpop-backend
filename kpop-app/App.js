import React, { useState, useEffect, useCallback } from "react";
import {
  StyleSheet,
  Text,
  View,
  TouchableOpacity,
  Dimensions,
  StatusBar,
  Alert,
  PermissionsAndroid,
  Platform,
} from "react-native";
import Voice from "@react-native-voice/voice";

const { width } = Dimensions.get("window");

export default function App() {
  const [estaOuvindo, setEstaOuvindo] = useState(false);
  const [traducao, setTraducao] = useState("Aguardando voz...");
  const [textoOriginal, setTextoOriginal] = useState("");

  // Solicitar permissão explicitamente (Android)
  const checarPermissao = async () => {
    if (Platform.OS === "android") {
      try {
        const granted = await PermissionsAndroid.request(
          PermissionsAndroid.PERMISSIONS.RECORD_AUDIO,
          {
            title: "Acesso ao Microfone",
            message: "ARMY, o app precisa do microfone para traduzir as lives!",
            buttonPositive: "OK",
          }
        );
        return granted === PermissionsAndroid.RESULTS.GRANTED;
      } catch (err) {
        console.warn(err);
        return false;
      }
    }
    return true;
  };

  const onSpeechResults = useCallback((e) => {
    if (e.value && e.value.length > 0) {
      const fala = e.value[0];
      setTextoOriginal(fala);
      enviarParaTraducao(fala);
    }
  }, []);

  const onSpeechError = useCallback((e) => {
    console.log("Erro de voz:", e);
    setEstaOuvindo(false);
    if (e.error?.message?.includes("permiss")) {
      Alert.alert("Erro", "Permissão de microfone negada no sistema.");
    }
  }, []);

  useEffect(() => {
    // Configurar listeners
    Voice.onSpeechResults = onSpeechResults;
    Voice.onSpeechError = onSpeechError;

    // Pedir permissão ao abrir o app
    checarPermissao();

    StatusBar.setHidden(true);

    return () => {
      // Cleanup rigoroso
      Voice.destroy().then(Voice.removeAllListeners);
    };
  }, [onSpeechResults, onSpeechError]);

  const enviarParaTraducao = async (texto) => {
    try {
      const response = await fetch("http://192.168.110.225:8080/traduzir-app", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ texto }),
      });

      const data = await response.json();
      setTraducao(data.traducao || "Não consegui entender, Oppa...");
    } catch (e) {
      setTraducao("Erro de conexão com o PC...");
    }
  };

  const alternarEscuta = async () => {
    try {
      if (estaOuvindo) {
        await Voice.stop();
        setEstaOuvindo(false);
      } else {
        const temPermissao = await checarPermissao();
        if (!temPermissao) {
          Alert.alert("Erro", "Microfone não autorizado.");
          return;
        }

        setTextoOriginal("");
        setTraducao("Ouvindo coreano...");
        await Voice.start("ko-KR");
        setEstaOuvindo(true);
      }
    } catch (e) {
      console.error(e);
      setEstaOuvindo(false);
    }
  };

  return (
    <View style={styles.container} pointerEvents="box-none">
      <View style={styles.floatingOverlay}>
        <View style={styles.header}>
          <Text style={styles.branding}>K-POP STUDIO AI</Text>
          <Text style={styles.original} numberOfLines={1}>
            {textoOriginal || "Pronto para traduzir..."}
          </Text>
        </View>

        <Text style={styles.traducao}>{traducao}</Text>

        <TouchableOpacity
          onPress={alternarEscuta}
          activeOpacity={0.7}
          style={[styles.miniBotao, estaOuvindo && styles.btnAtivo]}
        >
          <View style={estaOuvindo ? styles.micOn : styles.micOff} />
          <Text style={styles.btnTexto}>{estaOuvindo ? "STOP" : "LIVE"}</Text>
        </TouchableOpacity>
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: "#121212",
    justifyContent: "flex-end",
    alignItems: "center",
    paddingBottom: 40,
  },
  floatingOverlay: {
    width: width * 0.95,
    backgroundColor: "rgba(10, 10, 10, 0.92)",
    borderRadius: 25,
    padding: 18,
    borderWidth: 1.5,
    borderColor: "#E11D48",
    flexDirection: "row",
    alignItems: "center",
    elevation: 20,
    shadowColor: "#E11D48",
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.3,
    shadowRadius: 8,
  },
  header: { flex: 1.2 },
  branding: {
    color: "#E11D48",
    fontSize: 9,
    fontWeight: "900",
    letterSpacing: 1,
    marginBottom: 2,
  },
  original: { color: "#888", fontSize: 11, fontStyle: "italic" },
  traducao: {
    flex: 3,
    color: "#fff",
    fontSize: 15,
    fontWeight: "600",
    marginHorizontal: 10,
    textAlign: "center",
  },
  miniBotao: {
    width: 55,
    height: 55,
    borderRadius: 28,
    backgroundColor: "#1a1a1a",
    justifyContent: "center",
    alignItems: "center",
    borderWidth: 1,
    borderColor: "#333",
  },
  btnAtivo: {
    backgroundColor: "#E11D48",
    borderColor: "#fff",
  },
  btnTexto: { color: "#fff", fontSize: 10, fontWeight: "bold", marginTop: 2 },
  micOn: { width: 8, height: 8, borderRadius: 4, backgroundColor: "#fff" },
  micOff: { width: 8, height: 8, borderRadius: 4, backgroundColor: "#E11D48" },
});