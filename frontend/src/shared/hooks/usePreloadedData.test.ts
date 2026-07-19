// Unit тесты для хуков usePreloadedData
import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import {
  getGraphWithPreload,
  getAchievementsWithPreload,
  usePreloadedDataStatus,
  useInstantData,
  loadAppData,
} from "./usePreloadedData";
import { PreloadService } from "$shared/services/PreloadService";
import * as graphApi from "$shared/api/graph";
import * as usersApi from "$shared/api/users";
import { UserPoints } from "$shared/lib/domain";
import {
  mockGraphData,
  mockAchievementsData,
  mockPersonalAchievementsData,
  mockGraphError,
  mockAchievementsError,
} from "$shared/services/__mocks__/PreloadService.mocks";

// Мокаем зависимости
vi.mock("$shared/api/graph", () => ({
  getFullGraphData: vi.fn(),
}));

vi.mock("$shared/api/users", () => ({
  getAllAchievements: vi.fn(),
  getMyAchievements: vi.fn(),
}));

describe("usePreloadedData Hooks", () => {
  beforeEach(() => {
    vi.clearAllMocks();

    // Полный сброс PreloadService
    delete (PreloadService as any).instance;

    // Получаем новый инстанс сервиса
    const service = PreloadService;

    // Принудительно очищаем все внутренние состояния
    (service as any).preloadedGraph = null;
    (service as any).preloadedAchievements = null;
    (service as any).isPreloading = false;
    (service as any).preloadPromise = null;

    // Очищаем кэш
    service.clearCache();

    vi.mocked(graphApi.getFullGraphData).mockResolvedValue(mockGraphData);
    vi.mocked(usersApi.getAllAchievements).mockResolvedValue(
      mockAchievementsData,
    );
    vi.mocked(usersApi.getMyAchievements).mockResolvedValue(
      mockPersonalAchievementsData,
    );
  });

  afterEach(() => {
    PreloadService.clearCache();
  });

  describe("getGraphWithPreload", () => {
    it("should return preloaded data when available", async () => {
      // Предзагружаем данные
      await PreloadService.startPreload();

      const result = await getGraphWithPreload(500);

      expect(result).toEqual(mockGraphData);
      // API не должен вызываться, так как есть предзагруженные данные
      expect(graphApi.getFullGraphData).not.toHaveBeenCalledWith(500);
    });

    it("should fetch from server when no preloaded data", async () => {
      const result = await getGraphWithPreload(500);

      expect(result).toEqual(mockGraphData);
      expect(graphApi.getFullGraphData).toHaveBeenCalledWith(500);
    });

    it("should use default limit when not specified", async () => {
      const result = await getGraphWithPreload();

      expect(result).toEqual(mockGraphData);
      expect(graphApi.getFullGraphData).toHaveBeenCalledWith(1000);
    });

    it("should handle server errors gracefully", async () => {
      vi.mocked(graphApi.getFullGraphData).mockRejectedValue(mockGraphError);

      await expect(getGraphWithPreload()).rejects.toThrow(mockGraphError);
    });

    it("should log when using preloaded data", async () => {
      const consoleSpy = vi.spyOn(console, "log").mockImplementation(() => {});

      await PreloadService.startPreload();
      await getGraphWithPreload();

      expect(consoleSpy).toHaveBeenCalledWith(
        "[usePreloadedData] Using preloaded graph data",
      );

      consoleSpy.mockRestore();
    });

    it("should log when loading from server", async () => {
      const consoleSpy = vi.spyOn(console, "log").mockImplementation(() => {});

      await getGraphWithPreload();

      expect(consoleSpy).toHaveBeenCalledWith(
        "[usePreloadedData] Loading graph data from server",
      );

      consoleSpy.mockRestore();
    });
  });

  describe("getAchievementsWithPreload", () => {
    it("should fetch public achievements from server", async () => {
      const result = await getAchievementsWithPreload(false);

      expect(result).toEqual({
        achievements: mockAchievementsData.achievements,
      });
      expect(usersApi.getAllAchievements).toHaveBeenCalledTimes(1);
      expect(usersApi.getMyAchievements).not.toHaveBeenCalled();
    });

    it("should fetch personal achievements from server when usePersonal is true", async () => {
      const result = await getAchievementsWithPreload(true);

      expect(result).toEqual(mockPersonalAchievementsData);
      expect(usersApi.getMyAchievements).toHaveBeenCalledTimes(1);
      expect(usersApi.getAllAchievements).not.toHaveBeenCalled();
    });

    it("should handle API errors gracefully", async () => {
      vi.mocked(usersApi.getAllAchievements).mockRejectedValue(
        mockAchievementsError,
      );

      await expect(getAchievementsWithPreload(false)).rejects.toThrow(
        mockAchievementsError,
      );
    });

    it("should log when loading achievements from server", async () => {
      const consoleSpy = vi.spyOn(console, "log").mockImplementation(() => {});

      await getAchievementsWithPreload(false);

      expect(consoleSpy).toHaveBeenCalledWith(
        "[usePreloadedData] Loading all achievements from server",
      );

      consoleSpy.mockRestore();
    });
  });

  describe("usePreloadedDataStatus", () => {
    it("should return correct status when no data is preloaded", () => {
      const status = usePreloadedDataStatus();

      expect(status).toEqual({
        hasData: false,
        hasGraph: false,
        hasAchievements: false,
      });
    });

    it("should return correct status when graph is preloaded", async () => {
      await PreloadService.startPreload();

      const status = usePreloadedDataStatus();

      expect(status).toEqual({
        hasData: true,
        hasGraph: true,
        hasAchievements: false,
      });
    });
  });

  describe("useInstantData", () => {
    it("should return empty data when nothing is preloaded", () => {
      const instantData = useInstantData();

      expect(instantData).toEqual({
        graph: { nodes: [], links: [] },
        delta: null,
        achievements: [],
        hasInstantData: false,
        isDataReady: false,
      });
    });

    it("should return preloaded graph when available", async () => {
      await PreloadService.startPreload();

      const instantData = useInstantData();

      expect(instantData.graph).toEqual(mockGraphData);
      expect(instantData.achievements).toEqual([]);
      expect(instantData.hasInstantData).toBe(true);
      expect(instantData.isDataReady).toBe(true);
    });

    it("should return partial data when only one type is preloaded", async () => {
      // Предзагружаем только граф
      vi.mocked(usersApi.getAllAchievements).mockRejectedValue(
        mockAchievementsError,
      );

      await PreloadService.startPreload();

      const instantData = useInstantData();

      expect(instantData.graph).toEqual(mockGraphData);
      expect(instantData.achievements).toEqual([]);
      expect(instantData.hasInstantData).toBe(true);
      expect(instantData.isDataReady).toBe(true);
    });
  });

  describe("loadAppData", () => {
    it("should load graph from preloaded data and achievements from server", async () => {
      await PreloadService.startPreload();

      const result = await loadAppData();

      expect(result.graph).toEqual(mockGraphData);
      expect(result.achievements).toEqual(mockAchievementsData.achievements);
      expect(result.usedPreloaded.graph).toBe(true);
      expect(result.usedPreloaded.achievements).toBe(false);
      expect(result.totalPoints).toBeUndefined();
      expect(result.userPoints).toBeInstanceOf(UserPoints);
      expect(result.userPoints?.computedTotal).toBe(25);
    });

    it("should load data from server when no preloaded data", async () => {
      const result = await loadAppData();

      expect(result.graph).toEqual(mockGraphData);
      expect(result.achievements).toEqual(mockAchievementsData.achievements);
      expect(result.usedPreloaded.graph).toBe(false);
      expect(result.usedPreloaded.achievements).toBe(false);
    });

    it("should load personal achievements when specified", async () => {
      await PreloadService.startPreload();

      const result = await loadAppData({ usePersonalAchievements: true });

      expect(result.achievements).toEqual(
        mockPersonalAchievementsData.achievements,
      );
      expect(result.totalPoints).toBe(35);
      expect(result.userPoints).toBeInstanceOf(UserPoints);
      expect(result.userPoints?.total).toBe(35);
      expect(result.userPoints?.isConsistent).toBe(true);
      expect(result.usedPreloaded.achievements).toBe(false); // Всегда false для персональных
    });

    it("should use custom limit for graph", async () => {
      const result = await loadAppData({ limit: 500 });

      expect(result.graph).toEqual(mockGraphData);
      expect(graphApi.getFullGraphData).toHaveBeenCalledWith(500);
    });

    it("should handle API errors when fallbackToServer is false", async () => {
      vi.mocked(graphApi.getFullGraphData).mockRejectedValue(mockGraphError);
      vi.mocked(usersApi.getAllAchievements).mockRejectedValue(
        mockAchievementsError,
      );

      await expect(loadAppData({ fallbackToServer: false })).rejects.toThrow();
    });

    it("should load data in parallel", async () => {
      const startTime = Date.now();

      await loadAppData();

      const endTime = Date.now();
      const duration = endTime - startTime;

      // Параллельная загрузка должна быть быстрой
      expect(duration).toBeLessThan(100);
    });

    it("should mix preloaded and fresh data correctly", async () => {
      // Предзагружаем только граф
      vi.mocked(usersApi.getAllAchievements).mockRejectedValue(
        mockAchievementsError,
      );
      await PreloadService.startPreload();

      // Восстанавливаем мок для достижений
      vi.mocked(usersApi.getAllAchievements).mockResolvedValue(
        mockAchievementsData,
      );

      const result = await loadAppData();

      expect(result.usedPreloaded.graph).toBe(true);
      expect(result.usedPreloaded.achievements).toBe(false);
      expect(result.graph).toEqual(mockGraphData);
      expect(result.achievements).toEqual(mockAchievementsData.achievements);
    });

    it("should use default options when not specified", async () => {
      const result = await loadAppData({});

      expect(result.graph).toEqual(mockGraphData);
      expect(result.achievements).toEqual(mockAchievementsData.achievements);
      expect(graphApi.getFullGraphData).toHaveBeenCalledWith(1000);
    });
  });
});
