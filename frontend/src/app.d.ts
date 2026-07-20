// See https://svelte.dev/docs/kit/types#app.d.ts
// for information about these interfaces
declare global {
  namespace App {
    // interface Error {}
    // interface Locals {}
    // interface PageData {}
    // interface PageState {}
    // interface Platform {}
  }

  interface Window {
    __SKIP_AUTH__?: boolean;
    __graphCanvas?: {
      getSimulationNodes: () => { id: string; title: string; type?: string }[];
      transform: { x: number; y: number; k: number };
    };
  }

  interface Navigator {
    deviceMemory?: number;
  }

  var anomalyConfig:
    { reality_rift?: { core_color: string; glow_color: string } } | undefined;

  interface ImportMeta {
    readonly env: ImportMetaEnv;
  }

  interface ImportMetaEnv {
    readonly VITE_API_URL?: string;
    readonly VITE_GRAPH_SERVICE_URL?: string;
  }
}

export {};
