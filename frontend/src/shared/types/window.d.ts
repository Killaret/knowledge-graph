export {};

declare global {
  interface Window {
    __SKIP_AUTH__?: boolean;
    __ACCESS_TOKEN__?: string;
  }
}
