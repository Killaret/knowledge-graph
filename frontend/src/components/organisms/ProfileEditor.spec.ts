import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, fireEvent, cleanup, waitFor } from "@testing-library/svelte";
import ProfileEditor from "./ProfileEditor.svelte";
import { goto } from "$app/navigation";
import { updateMe, deleteMe, updateSetting } from "$shared/api/users";
import { currentUser, updateUserInfo, logout } from "$shared/stores/auth.svelte.js";
import type { User } from "$shared/types";

const mockUser: User = {
  id: "1",
  login: "testuser",
  email: "test@example.com",
  role: "user",
  created_at: new Date().toISOString(),
};

vi.mock("$shared/stores/auth.svelte.js", () => ({
  currentUser: vi.fn(),
  updateUserInfo: vi.fn(),
  logout: vi.fn(),
  isLoading: () => false,
  error: () => null,
  isAuthenticated: () => true,
}));

vi.mock("$shared/api/users", () => ({
  updateMe: vi.fn(),
  deleteMe: vi.fn(),
  updateSetting: vi.fn(),
  getSettings: vi.fn(),
  getMe: vi.fn(),
}));

describe("ProfileEditor", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    cleanup();
    localStorage.clear();
    vi.mocked(currentUser).mockReturnValue(mockUser);
    vi.mocked(updateUserInfo).mockResolvedValue(undefined);
    vi.mocked(logout).mockResolvedValue(undefined);
    vi.mocked(updateMe).mockResolvedValue({ ...mockUser });
    vi.mocked(deleteMe).mockResolvedValue(undefined);
    vi.mocked(updateSetting).mockResolvedValue(undefined);
  });

  afterEach(() => {
    cleanup();
  });

  it("renders current user data", () => {
    render(ProfileEditor);

    const nameInput = screen.getByDisplayValue("testuser") as HTMLInputElement;
    const emailInput = screen.getByLabelText(/email/i) as HTMLInputElement;

    expect(nameInput).toHaveValue("testuser");
    expect(nameInput).toBeDisabled();
    expect(emailInput).toHaveValue("test@example.com");
  });

  it("saves email changes successfully", async () => {
    render(ProfileEditor);

    const emailInput = screen.getByLabelText(/email/i);
    const saveButton = screen.getByRole("button", { name: /save/i });

    await fireEvent.input(emailInput, { target: { value: "new@example.com" } });
    await fireEvent.click(saveButton);

    await waitFor(() => {
      expect(updateMe).toHaveBeenCalledWith({ email: "new@example.com" });
      expect(updateUserInfo).toHaveBeenCalled();
    });

    expect(screen.getByText(/saved/i)).toBeInTheDocument();
  });

  it("saves with empty email as undefined", async () => {
    render(ProfileEditor);

    const emailInput = screen.getByLabelText(/email/i);
    const saveButton = screen.getByRole("button", { name: /save/i });

    await fireEvent.input(emailInput, { target: { value: "" } });
    await fireEvent.click(saveButton);

    await waitFor(() => {
      expect(updateMe).toHaveBeenCalledWith({ email: undefined });
    });
  });

  it("shows error when save fails", async () => {
    vi.mocked(updateMe).mockRejectedValue(new Error("update failed"));

    render(ProfileEditor);

    const saveButton = screen.getByRole("button", { name: /save/i });
    await fireEvent.click(saveButton);

    await waitFor(() => {
      expect(screen.getByRole("alert")).toBeInTheDocument();
    });
  });

  it("changes locale and persists to backend", async () => {
    const reloadSpy = vi.spyOn(window.location, "reload").mockImplementation(() => {});

    render(ProfileEditor);

    const localeSelect = screen.getByLabelText(/language/i) as HTMLSelectElement;
    await fireEvent.change(localeSelect, { target: { value: "ru" } });

    await waitFor(() => {
      expect(updateSetting).toHaveBeenCalledWith("preferred_language", "ru");
      expect(localStorage.getItem("locale")).toBe("ru");
    });

    reloadSpy.mockRestore();
  });

  it("recovers from locale save error", async () => {
    vi.mocked(updateSetting).mockRejectedValue(new Error("locale save failed"));
    const reloadSpy = vi.spyOn(window.location, "reload").mockImplementation(() => {});

    render(ProfileEditor);

    const localeSelect = screen.getByLabelText(/language/i) as HTMLSelectElement;
    await fireEvent.change(localeSelect, { target: { value: "ru" } });

    await waitFor(() => {
      expect(updateSetting).toHaveBeenCalled();
      expect(localStorage.getItem("locale")).toBe("ru");
    });

    reloadSpy.mockRestore();
  });

  it("opens and cancels delete account modal", async () => {
    render(ProfileEditor);

    const deleteButton = screen.getByRole("button", { name: /delete account/i });
    await fireEvent.click(deleteButton);

    expect(screen.getByRole("button", { name: /cancel/i })).toBeInTheDocument();

    await fireEvent.click(screen.getByRole("button", { name: /cancel/i }));

    await waitFor(() => {
      expect(screen.queryByRole("button", { name: /cancel/i })).not.toBeInTheDocument();
    });
  });

  it("deletes account and redirects to login", async () => {
    render(ProfileEditor);

    const deleteButton = screen.getByRole("button", { name: /delete account/i });
    await fireEvent.click(deleteButton);

    const passwordInput = screen.getByLabelText(/password for confirmation/i);
    await fireEvent.input(passwordInput, { target: { value: "secret123" } });

    const confirmDelete = screen.getByRole("button", { name: /confirm delete/i });
    await fireEvent.click(confirmDelete);

    await waitFor(() => {
      expect(deleteMe).toHaveBeenCalledWith("secret123");
      expect(logout).toHaveBeenCalled();
      expect(goto).toHaveBeenCalledWith("/auth/login");
    });
  });

  it("shows error when account deletion fails", async () => {
    vi.mocked(deleteMe).mockRejectedValue(new Error("delete failed"));

    render(ProfileEditor);

    const deleteButton = screen.getByRole("button", { name: /delete account/i });
    await fireEvent.click(deleteButton);

    const passwordInput = screen.getByLabelText(/password for confirmation/i);
    await fireEvent.input(passwordInput, { target: { value: "secret123" } });

    const confirmDelete = screen.getByRole("button", { name: /confirm delete/i });
    await fireEvent.click(confirmDelete);

    await waitFor(() => {
      expect(screen.getByText(/DELETE_ERROR/)).toBeInTheDocument();
    });
  });
});
