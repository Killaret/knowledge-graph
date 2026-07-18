import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, cleanup } from "@testing-library/svelte";
import RegisterForm from "./RegisterForm.svelte";

const mockRegister = vi.fn();

vi.mock("$shared/stores/auth.svelte.js", () => ({
  register: (...args: any[]) => mockRegister(...args),
  isLoading: () => false,
  error: () => null,
}));

vi.mock("$app/navigation", () => ({
  goto: vi.fn(),
}));

describe("RegisterForm", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    cleanup();
  });

  it("renders registration form with all fields", () => {
    render(RegisterForm);

    expect(screen.getByLabelText(/login/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/email/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/^password/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/confirm password/i)).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: /register/i }),
    ).toBeInTheDocument();
  });

  it("updates all form fields", async () => {
    render(RegisterForm);

    const loginInput = screen.getByLabelText(/login/i);
    const emailInput = screen.getByLabelText(/email/i);
    const passwordInput = screen.getByLabelText(/^password/i);
    const confirmInput = screen.getByLabelText(/confirm password/i);

    await fireEvent.input(loginInput, { target: { value: "newuser" } });
    await fireEvent.input(emailInput, {
      target: { value: "newuser@example.com" },
    });
    await fireEvent.input(passwordInput, { target: { value: "Password123!" } });
    await fireEvent.input(confirmInput, { target: { value: "Password123!" } });

    expect(loginInput).toHaveValue("newuser");
    expect(emailInput).toHaveValue("newuser@example.com");
    expect(passwordInput).toHaveValue("Password123!");
    expect(confirmInput).toHaveValue("Password123!");
  });

  it("shows password requirements when password is entered", async () => {
    render(RegisterForm);

    const passwordInput = screen.getByLabelText(/^password/i);
    await fireEvent.input(passwordInput, { target: { value: "weak" } });

    expect(screen.getByText(/password requirements/i)).toBeInTheDocument();
    expect(screen.getByText(/minimum 10 characters/i)).toBeInTheDocument();
    expect(screen.getByText(/uppercase letter/i)).toBeInTheDocument();
  });

  it("shows error when login is empty", async () => {
    const { container } = render(RegisterForm);

    const form = container.querySelector("form");
    await fireEvent.submit(form!);

    expect(screen.getByRole("alert")).toHaveTextContent(/login is required/i);
  });

  it("shows error when passwords do not match", async () => {
    render(RegisterForm);

    const passwordInput = screen.getByLabelText(/^password/i);
    const confirmInput = screen.getByLabelText(/confirm password/i);

    await fireEvent.input(passwordInput, { target: { value: "Password123!" } });
    await fireEvent.input(confirmInput, { target: { value: "Different123!" } });

    expect(screen.getByText(/passwords do not match/i)).toBeInTheDocument();
  });

  it("submits registration with valid data", async () => {
    mockRegister.mockResolvedValue(true);
    render(RegisterForm);

    const loginInput = screen.getByLabelText(/login/i);
    const emailInput = screen.getByLabelText(/email/i);
    const passwordInput = screen.getByLabelText(/^password/i);
    const confirmInput = screen.getByLabelText(/confirm password/i);
    const submitButton = screen.getByRole("button", { name: /register/i });

    await fireEvent.input(loginInput, { target: { value: "newuser" } });
    await fireEvent.input(emailInput, {
      target: { value: "newuser@example.com" },
    });
    await fireEvent.input(passwordInput, { target: { value: "Password123!" } });
    await fireEvent.input(confirmInput, { target: { value: "Password123!" } });
    await fireEvent.click(submitButton);

    expect(mockRegister).toHaveBeenCalledWith(
      "newuser",
      "Password123!",
      "newuser@example.com",
    );
  });

  it("has link to login page", () => {
    render(RegisterForm);

    const loginLink = screen.getByRole("link", { name: /sign in/i });
    expect(loginLink).toHaveAttribute("href", "/auth/login");
  });
});
