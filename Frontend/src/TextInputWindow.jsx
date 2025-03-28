import { useState } from "react";

function TextInputWindow() {
  const [inputValue, setInputValue] = useState("");
  const [displayText, setDisplayText] = useState("");
  const [loading, setLoading] = useState(false);  // Para mostrar un estado de carga

  // Función para mostrar el texto ingresado
  const handleShowText = async () => {
    if (inputValue.trim()) {
      setLoading(true); // Activar el estado de carga
      try {
        // Enviar el texto al backend
        const response = await fetch("http://localhost:8080/endpoint", {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({ message: inputValue }), // Enviar el mensaje
        });

        if (response.ok) {
          const data = await response.json();
          setDisplayText(`Resultado: ${data.result}`); // Suponiendo que el backend devuelva un JSON con el resultado
        } else {
          alert("Error al procesar el mensaje en el backend.");
        }
      } catch (error) {
        console.error("Error al conectar con el backend:", error);
        alert("Hubo un error en la conexión con el backend.");
      } finally {
        setLoading(false); // Desactivar el estado de carga
      }
    } else {
      alert("Por favor, ingrese un mensaje.");
    }
  };

  // Función para limpiar el área de entrada y salida
  const handleClear = () => {
    setInputValue("");
    setDisplayText("");
  };

  // Función para cargar el archivo de texto
  const handleFileUpload = (event) => {
    const file = event.target.files[0];
    if (file && file.type === "text/plain") {
      const reader = new FileReader();
      reader.onload = () => {
        setInputValue(reader.result);
      };
      reader.readAsText(file);
    } else {
      alert("Cargar archivo.");
    }
  };

  return (
    <div className="flex flex-col min-h-screen bg-gray-100">
      {/* Área de entrada en la parte superior */}
      <div className="w-full p-4 bg-white shadow-md">
        <label
          htmlFor="message"
          className="block mb-2 text-sm font-medium text-gray-900"
        >
          Entrada:
        </label>
        <textarea
          id="message"
          rows="4"
          value={inputValue}
          onChange={(e) => setInputValue(e.target.value)}
          className="block p-2.5 w-full text-sm text-gray-900 bg-gray-50 rounded-lg border border-gray-300 focus:ring-blue-500 focus:border-blue-500"
          placeholder="Texto de entrada..."
        ></textarea>
      </div>

      {/* Área de botones */}
      <div className="flex justify-between p-4">
        <button
          onClick={handleShowText}
          className="bg-blue-500 text-white p-2 rounded-lg hover:bg-blue-600 transition"
        >
          {loading ? "Cargando..." : "Mostrar Texto"}
        </button>
        <button
          onClick={handleClear}
          className="bg-gray-500 text-white p-2 rounded-lg hover:bg-gray-600 transition"
        >
          Limpiar
        </button>
        <input
          type="file"
          accept=".txt"
          onChange={handleFileUpload}
          className="p-2 rounded-lg bg-gray-200 text-sm text-gray-900"
        />
      </div>

      {/* Área de salida */}
      {displayText && (
        <div className="w-full p-4 bg-white shadow-md mt-4">
          <h3 className="text-lg font-semibold text-gray-900">Salida:</h3>
          <p className="text-sm text-gray-700">{displayText}</p>
        </div>
      )}
    </div>
  );
}

export default TextInputWindow;
