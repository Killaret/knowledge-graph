import { describe, it, expect, vi, beforeEach, beforeAll, afterAll } from "vitest";
import { render, screen, fireEvent, cleanup } from "@testing-library/svelte";
import LoginForm from "./LoginForm.svelte";
import { goto } from "$app/navigation";

const mockLogin = vi.fn();
const mockLoginWithApiKey = vi.fn();

const envCache: Record<string, string | undefined> = {};

function enableFeatureFlags() {
  envCache.VITE_API_KEY_ENABLED = import.meta.env.VITE_API_KEY_ENABLED as string | undefined;
  envCache.VITE_YANDEX_ENABLED = import.meta.env.VITE_YANDEX_ENABLED as string | undefined;
  import.meta.env.VITE_API_KEY_ENABLED = "true";
  import.meta.env.VITE_YANDEX_ENABLED = "true";
}

function restoreFeatureFlags() {
  import.meta.env.VITE_API_KEY_ENABLED = envCache.VITE_API_KEY_ENABLED;
  import.meta.env.VITE_YANDEX_ENABLED = envCache.VITE_YANDEX_ENABLED;
}

vi.mock("$shared/stores/auth.svelte.js", () => ({
  login: (...args: any[]) => mockLogin(...args),
  loginWithApiKey: (...args: any[]) => mockLoginWithApiKey(...args),
  isLoading: () => false,
  error: () => null,
}));

vi.mock("$app/navigation", () => ({
  goto: vi.fn(),
}));

describe("LoginForm", () => {
  beforeAll(() => {
    enableFeatureFlags();
  });

  afterAll(() => {
    restoreFeatureFlags();
  });

  beforeEach(() => {
    vi.clearAllMocks();
    cleanup();
  });

  it("renders login form with default fields", () => {
    render(LoginForm);

    expect(screen.getByLabelText(/login/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/password/i)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /^Sign in$/i })).toBeInTheDocument();
  });

  it("updates login and password fields", async () => {
    render(LoginForm);

    const loginInput = screen.getByLabelText(/login/i);
    const passwordInput = screen.getByLabelText(/password/i);

    await fireEvent.input(loginInput, { target: { value: "testuser" } });
    await fireEvent.input(passwordInput, { target: { value: "password123" } });

    expect(loginInput).toHaveValue("testuser");
    expect(passwordInput).toHaveValue("password123");
  });

  it("shows error when fields are empty", async () => {
    const { container } = render(LoginForm);

    const form = container.querySelector("form");
    await fireEvent.submit(form!);

    expect(screen.getByRole("alert")).toHaveTextContent(/please enter login and password/i);
  });

  it("shows error when login fails", async () => {
    mockLogin.mockResolvedValue(false);
    render(LoginForm);

    const loginInput = screen.getByLabelText(/login/i);
    const passwordInput = screen.getByLabelText(/password/i);

    await fireEvent.input(loginInput, { target: { value: "testuser" } });
    await fireEvent.input(passwordInput, { target: { value: "password123" } });
    await fireEvent.click(screen.getByRole("button", { name: /^Sign in$/i }));

    expect(screen.getByRole("alert")).toHaveTextContent(/invalid credentials/i);
  });

  it("submits login with valid credentials and navigates on success", async () => {
    mockLogin.mockResolvedValue(true);
    render(LoginForm);

    const loginInput = screen.getByLabelText(/login/i);
    const passwordInput = screen.getByLabelText(/password/i);
    const submitButton = screen.getByRole("button", { name: /^Sign in$/i });

    await fireEvent.input(loginInput, { target: { value: "testuser" } });
    await fireEvent.input(passwordInput, { target: { value: "password123" } });
    await fireEvent.click(submitButton);

    expect(mockLogin).toHaveBeenCalledWith("testuser", "password123");
    expect(goto).toHaveBeenCalledWith("/");
  });

  it("calls onSuccess instead of goto when login succeeds", async () => {
    const onSuccess = vi.fn();
    mockLogin.mockResolvedValue(true);

    render(LoginForm, { props: { onSuccess } });

    await fireEvent.input(screen.getByLabelText(/login/i), { target: { value: "testuser" } });
    await fireEvent.input(screen.getByLabelText(/password/i), { target: { value: "password123" } });
    await fireEvent.click(screen.getByRole("button", { name: /^Sign in$/i }));

    expect(onSuccess).toHaveBeenCalled();
    expect(goto).not.toHaveBeenCalled();
  });

  it("switches to API key mode and submits", async () => {
    mockLoginWithApiKey.mockResolvedValue(true);
    render(LoginForm);

    const apiMode = screen.getByRole("button", { name: /api key/i });
    await fireEvent.click(apiMode);

    const keyInput = screen.getByLabelText(/api key/i);
    await fireEvent.input(keyInput, { target: { value: "secret-key" } });
    await fireEvent.click(screen.getByRole("button", { name: /^Sign in$/i }));

    expect(mockLoginWithApiKey).toHaveBeenCalledWith("secret-key");
  });

  it("shows error when API key login fails", async () => {
    mockLoginWithApiKey.mockResolvedValue(false);
    render(LoginForm);

    const apiMode = screen.getByRole("button", { name: /api key/i });
    await fireEvent.click(apiMode);

    const keyInput = screen.getByLabelText(/api key/i);
    await fireEvent.input(keyInput, { target: { value: "wrong-key" } });
    await fireEvent.click(screen.getByRole("button", { name: /^Sign in$/i }));

    expect(screen.getByRole("alert")).toHaveTextContent(/invalid api key/i);
  });

  it("renders Yandex login button when enabled", () => {
    render(LoginForm);
    expect(screen.getByRole("button", { name: /yandex/i })).toBeInTheDocument();
  });

  it("has link to register page", () => {
    render(LoginForm);

    const registerLink = screen.getByRole("link", { name: /register/i });
    expect(registerLink).toHaveAttribute("href", "/auth/register");
  });

  it("calls onRegister when register link is clicked", async () => {
    const onRegister = vi.fn();
    const { container } = render(LoginForm, { props: { onRegister } });

    const registerLink = screen.getByRole("link", { name: /register/i });
    await fireEvent.click(registerLink);

    expect(onRegister).toHaveBeenCalled();
  });

  it("has link to forgot password page", () => {
    render(LoginForm);

    const forgotLink = screen.getByRole("link", { name: /forgot password/i });
    expect(forgotLink).toHaveAttribute("href", "/auth/forgot-password");
  });
});
