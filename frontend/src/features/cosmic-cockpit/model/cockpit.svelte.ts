import { browser } from "$app/environment";
import type {
  CockpitPanelPosition,
  CockpitPanelState,
  CockpitPanelsState,
  CockpitSettings,
} from "./cockpit.types";

const STORAGE_KEY = "cockpit-settings";

const defaultPanel: CockpitPanelState = {
  open: false,
  pinned: false,
  hovering: false,
};

const defaultSettings: CockpitSettings = {
  hoverDelay: 350,
  edgeSensitivity: 40,
  autoCollapse: true,
  reducedMotion: false,
};

function loadSettings(): CockpitSettings {
  if (!browser) return { ...defaultSettings };
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (raw) {
      const parsed = JSON.parse(raw) as Partial<CockpitSettings>;
      return {
        hoverDelay: parsed.hoverDelay ?? defaultSettings.hoverDelay,
        edgeSensitivity: parsed.edgeSensitivity ?? defaultSettings.edgeSensitivity,
        autoCollapse: parsed.autoCollapse ?? defaultSettings.autoCollapse,
        reducedMotion: parsed.reducedMotion ?? defaultSettings.reducedMotion,
      };
    }
  } catch {
    // ignore corrupt localStorage
  }
  return { ...defaultSettings };
}

function persistSettings(settings: CockpitSettings) {
  if (!browser) return;
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(settings));
  } catch {
    // ignore
  }
}

function createCockpitStore() {
  const initialSettings = loadSettings();

  let firstPerson = $state(false);
  let activeCluster = $state<string | null>(null);
  let fps = $state(0);
  let lastSyncAt = $state<number | null>(null);
  let syncing = $state(false);
  let hoverDelay = $state(initialSettings.hoverDelay);
  let edgeSensitivity = $state(initialSettings.edgeSensitivity);
  let autoCollapse = $state(initialSettings.autoCollapse);
  let reducedMotion = $state(initialSettings.reducedMotion);

  const panels = $state<CockpitPanelsState>({
    top: { ...defaultPanel },
    bottom: { ...defaultPanel },
    left: { ...defaultPanel },
    right: { ...defaultPanel },
  });

  function getSettings(): CockpitSettings {
    return {
      hoverDelay,
      edgeSensitivity,
      autoCollapse,
      reducedMotion,
    };
  }

  function setPanel<K extends keyof CockpitPanelsState>(key: K, patch: Partial<CockpitPanelState>) {
    panels[key] = { ...panels[key], ...patch };
  }

  function togglePin(position: CockpitPanelPosition) {
    const panel = panels[position];
    setPanel(position, { pinned: !panel.pinned, open: true });
  }

  function openPanel(position: CockpitPanelPosition) {
    setPanel(position, { open: true });
  }

  function closePanel(position: CockpitPanelPosition) {
    const panel = panels[position];
    if (panel.pinned) return;
    setPanel(position, { open: false, hovering: false });
  }

  function hoverPanel(position: CockpitPanelPosition, hovering: boolean) {
    setPanel(position, { hovering });
  }

  function setFirstPerson(value: boolean) {
    firstPerson = value;
  }

  function toggleFirstPerson() {
    firstPerson = !firstPerson;
  }

  function exitFirstPerson() {
    firstPerson = false;
  }

  function setFps(value: number) {
    fps = value;
  }

  function setSyncing(value: boolean) {
    syncing = value;
    if (!value) {
      lastSyncAt = Date.now();
    }
  }

  function setActiveCluster(cluster: string | null) {
    activeCluster = cluster;
  }

  return {
    get firstPerson() {
      return firstPerson;
    },
    get activeCluster() {
      return activeCluster;
    },
    get fps() {
      return fps;
    },
    get lastSyncAt() {
      return lastSyncAt;
    },
    get syncing() {
      return syncing;
    },
    get hoverDelay() {
      return hoverDelay;
    },
    set hoverDelay(value: number) {
      hoverDelay = value;
    },
    get edgeSensitivity() {
      return edgeSensitivity;
    },
    set edgeSensitivity(value: number) {
      edgeSensitivity = value;
    },
    get autoCollapse() {
      return autoCollapse;
    },
    set autoCollapse(value: boolean) {
      autoCollapse = value;
    },
    get reducedMotion() {
      return reducedMotion;
    },
    set reducedMotion(value: boolean) {
      reducedMotion = value;
    },
    get panels() {
      return panels;
    },
    get settings() {
      return getSettings();
    },
    saveSettings() {
      persistSettings(getSettings());
    },
    setPanel,
    togglePin,
    openPanel,
    closePanel,
    hoverPanel,
    setFirstPerson,
    toggleFirstPerson,
    exitFirstPerson,
    setFps,
    setSyncing,
    setActiveCluster,
  };
}

export const cockpitStore = createCockpitStore();
