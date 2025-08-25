// Go WASM API Client
class GoWASMClient {
  constructor() {
    this.isReady = false;
    this.init();
  }

  async init() {
    try {
      // Initialize Go WASM
      const go = new Go();
      const result = await WebAssembly.instantiateStreaming(
        fetch("main.wasm"),
        go.importObject,
      );
      go.run(result.instance);

      // Wait a bit for Go to fully initialize
      setTimeout(() => {
        this.isReady = typeof window.goAPI !== "undefined";
        this.updateStatus();
      }, 100);
    } catch (error) {
      console.error("Failed to load Go WASM:", error);
      this.updateStatus("Failed to load Go WASM: " + error.message);
    }
  }

  updateStatus(errorMessage = null) {
    const statusIndicator = document.getElementById("statusIndicator");
    const statusText = document.getElementById("statusText");

    if (errorMessage) {
      statusIndicator.className = "status-dot error";
      statusText.textContent = errorMessage;
    } else if (this.isReady) {
      statusIndicator.className = "status-dot ready";
      statusText.textContent = "Go WASM API Ready";
    } else {
      statusIndicator.className = "status-dot loading";
      statusText.textContent = "Loading Go WASM API...";
    }
  }

  // API wrapper methods with error handling
  callAPI(functionName, ...args) {
    console.log(`Calling ${functionName} with args:`, args);
    
    if (!this.isReady) {
      console.error("Go WASM API not ready");
      return {
        success: false,
        error: "Go WASM API not ready yet",
      };
    }

    if (!window.goAPI) {
      console.error("window.goAPI is undefined");
      return {
        success: false,
        error: "Go API object not found",
      };
    }
    
    if (typeof window.goAPI[functionName] !== "function") {
      console.error(`Function ${functionName} not found on goAPI. Available functions:`, Object.keys(window.goAPI || {}));
      return {
        success: false,
        error: `Function ${functionName} not available`,
      };
    }

    try {
      console.log(`Invoking ${functionName}...`);
      const result = window.goAPI[functionName](...args);
      console.log(`Result from ${functionName}:`, result);
      return result || { success: false, error: "Function returned undefined" };
    } catch (error) {
      console.error(`Error calling ${functionName}:`, error);
      return {
        success: false,
        error: error.message,
      };
    }
  }

  // Utility methods
  displayResult(elementId, result) {
    if (typeof showResult === 'function') {
      // Use enhanced showResult function if available (from HTML)
      const isSuccess = result && result.success;
      showResult(elementId, result, isSuccess);
    } else {
      // Fallback for compatibility
      const element = document.getElementById(elementId);
      const pre = element.querySelector("pre");

      if (!result) {
        pre.textContent = "Error: No response from Go WASM API";
        element.classList.add("error");
      } else if (result.success) {
        pre.textContent = JSON.stringify(result, null, 2);
        element.classList.add("success");
      } else {
        pre.textContent = `Error: ${result.error || result.message || "Unknown error"}`;
        element.classList.add("error");
      }

      element.classList.remove("hidden");
      element.classList.add("show");
    }
  }

  // Parse comma-separated numbers
  parseNumbers(input) {
    return input
      .split(",")
      .map((s) => s.trim())
      .filter((s) => s !== "")
      .map((s) => parseFloat(s))
      .filter((n) => !isNaN(n));
  }
}

// Initialize client
const client = new GoWASMClient();

// Demo functions
function processText() {
  const input = document.getElementById("textInput").value;
  const result = client.callAPI("processData", input);
  client.displayResult("textResults", result);
}

function calculateStats() {
  const input = document.getElementById("numbersInput").value;
  const numbers = client.parseNumbers(input);

  if (numbers.length === 0) {
    client.displayResult("statsResults", {
      success: false,
      error: "No valid numbers found",
    });
    return;
  }

  const result = client.callAPI("calculateStats", numbers);
  client.displayResult("statsResults", result);
}

function validateInput() {
  const input = document.getElementById("validateInput").value;
  const type = document.getElementById("validationType").value;
  const result = client.callAPI("validateInput", input, type);
  client.displayResult("validateResults", result);
}

function generateID() {
  const type = document.getElementById("idType").value;
  const result = client.callAPI("generateID", type);
  client.displayResult("idResults", result);
}

function formatJSON() {
  const input = document.getElementById("jsonInput").value;
  const result = client.callAPI("formatJSON", input);
  client.displayResult("jsonResults", result);
}

function getVersion() {
  const result = client.callAPI("getVersion");
  client.displayResult("versionResults", result);
}

// WebSocket example (commented out since we're not using a backend)
// This shows how you would integrate WebSockets with Go WASM processing
/*
function setupWebSocketExample() {
    const ws = new WebSocket('wss://echo.websocket.org');

    ws.onopen = function() {
        console.log('WebSocket connected');
        ws.send('Hello from Go WASM client!');
    };

    ws.onmessage = function(event) {
        // Process WebSocket data with Go WASM
        const processed = client.callAPI('processData', event.data);
        console.log('Processed WebSocket message:', processed);
    };

    ws.onclose = function() {
        console.log('WebSocket disconnected');
    };

    ws.onerror = function(error) {
        console.error('WebSocket error:', error);
    };
}
*/

// IndexedDB integration example
class LocalStorage {
  constructor() {
    this.dbName = "GoWASMApp";
    this.version = 1;
    this.db = null;
  }

  async init() {
    return new Promise((resolve, reject) => {
      const request = indexedDB.open(this.dbName, this.version);

      request.onerror = () => reject(request.error);
      request.onsuccess = () => {
        this.db = request.result;
        resolve(this.db);
      };

      request.onupgradeneeded = (event) => {
        const db = event.target.result;
        if (!db.objectStoreNames.contains("results")) {
          const store = db.createObjectStore("results", {
            keyPath: "id",
            autoIncrement: true,
          });
          store.createIndex("timestamp", "timestamp", { unique: false });
        }
      };
    });
  }

  async saveResult(data) {
    if (!this.db) await this.init();

    const transaction = this.db.transaction(["results"], "readwrite");
    const store = transaction.objectStore("results");

    return store.add({
      ...data,
      timestamp: Date.now(),
    });
  }

  async getResults(limit = 10) {
    if (!this.db) await this.init();

    const transaction = this.db.transaction(["results"], "readonly");
    const store = transaction.objectStore("results");
    const index = store.index("timestamp");

    return new Promise((resolve, reject) => {
      const request = index.openCursor(null, "prev");
      const results = [];

      request.onsuccess = (event) => {
        const cursor = event.target.result;
        if (cursor && results.length < limit) {
          results.push(cursor.value);
          cursor.continue();
        } else {
          resolve(results);
        }
      };

      request.onerror = () => reject(request.error);
    });
  }
}

// Initialize local storage
const localStorage = new LocalStorage();

// Auto-save results (optional)
const originalDisplayResult = client.displayResult;
client.displayResult = function (elementId, result) {
  originalDisplayResult.call(this, elementId, result);

  // Save successful results to IndexedDB
  if (result && result.success) {
    localStorage
      .saveResult({
        operation: elementId.replace("Results", ""),
        result: result,
      })
      .catch(console.error);
  }
};

// Keyboard shortcuts
document.addEventListener("keydown", function (event) {
  // Ctrl+Enter or Cmd+Enter to process text
  if ((event.ctrlKey || event.metaKey) && event.key === "Enter") {
    const activeElement = document.activeElement;

    if (activeElement.id === "textInput") {
      processText();
    } else if (activeElement.id === "numbersInput") {
      calculateStats();
    } else if (activeElement.id === "validateInput") {
      validateInput();
    } else if (activeElement.id === "jsonInput") {
      formatJSON();
    }

    event.preventDefault();
  }
});

// Show keyboard shortcuts hint
console.log(
  "💡 Tip: Use Ctrl+Enter (or Cmd+Enter) to execute functions when focused on input fields",
);
