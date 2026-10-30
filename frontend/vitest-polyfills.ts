// Polyfills for test environments that may lack browser APIs
if (typeof globalThis.BroadcastChannel === "undefined") {
  (globalThis as any).BroadcastChannel = class BroadcastChannel {
    name: string;
    constructor(name: string) {
      this.name = name;
    }
    postMessage() {}
    close() {}
    addEventListener() {}
    removeEventListener() {}
    onmessage = null;
  };
}
