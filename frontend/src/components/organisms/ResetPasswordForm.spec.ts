import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, cleanup, waitFor } from "@testing-library/svelte";
import ResetPasswordForm from "./ResetPasswordForm.svelte";
import { resetPassword } from "$shared/api/auth";
import { goto } from "$app/navigation";

vi.mock("$shared/api/auth", () => ({
  resetPassword: vi.fn(),
}));

describe("ResetPasswordForm", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    cleanup();
    vi.mocked(resetPassword).mockResolvedValue(undefined);
  });

  it("renders reset form with token", () => {
    render(ResetPasswordForm, { props: { token: "reset-token-123" } });

    expect(screen.getByLabelText(/new password/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/confirm password/i)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /save new password/i })).toBeInTheDocument();
  });

  it("shows password requirements when typing", async () => {
    render(ResetPasswordForm, { props: { token: "reset-token-123" } });

    const newPassword = screen.getByLabelText(/new password/i);
    await fireEvent.input(newPassword, { target: { value: "weak" } });

    expect(screen.getByText(/minimum 10 characters/i)).toBeInTheDocument();
    expect(screen.getByText(/uppercase letter/i)).toBeInTheDocument();
    expect(screen.getByText(/lowercase letter/i)).toBeInTheDocument();
    expect(screen.getByText(/number/i)).toBeInTheDocument();
    expect(screen.getByText(/special character/i)).toBeInTheDocument();
  });

  it("shows error when password requirements are not met", async () => {
    render(ResetPasswordForm, { props: { token: "reset-token-123" } });

    const newPassword = screen.getByLabelText(/new password/i);
    const confirmPassword = screen.getByLabelText(/confirm password/i);

    await fireEvent.input(newPassword, { target: { value: "short" } });
    await fireEvent.input(confirmPassword, { target: { value: "short" } });

    const form = document.querySelector(".reset-form")!;
    await fireEvent.submit(form);

    await waitFor(() => {
      expect(screen.getByText(/does not meet requirements/i)).toBeInTheDocument();
    });

    expect(resetPassword).not.toHaveBeenCalled();
  });

  it("shows error when passwords do not match", async () => {
    render(ResetPasswordForm, { props: { token: "reset-token-123" } });

    const newPassword = screen.getByLabelText(/new password/i);
    const confirmPassword = screen.getByLabelText(/confirm password/i);

    await fireEvent.input(newPassword, { target: { value: "StrongPass1!" } });
    await fireEvent.input(confirmPassword, { target: { value: "StrongPass2!" } });

    const form = document.querySelector(".reset-form");
    await fireEvent.submit(form!);

    await waitFor(() => {
      expect(screen.getAllByText(/passwords do not match/i).length).toBeGreaterThan(0);
    });

    expect(resetPassword).not.toHaveBeenCalled();
  });

  it("submits reset with valid matching password", async () => {
    render(ResetPasswordForm, { props: { token: "reset-token-123" } });

    const newPassword = screen.getByLabelText(/new password/i);
    const confirmPassword = screen.getByLabelText(/confirm password/i);

    await fireEvent.input(newPassword, { target: { value: "StrongPass1!" } });
    await fireEvent.input(confirmPassword, { target: { value: "StrongPass1!" } });

    const form = document.querySelector(".reset-form");
    await fireEvent.submit(form!);

    await waitFor(() => {
      expect(resetPassword).toHaveBeenCalledWith("reset-token-123", "StrongPass1!");
      expect(screen.getByText(/password changed successfully/i)).toBeInTheDocument();
    });
  });

  it("disables submit while loading", async () => {
    vi.mocked(resetPassword).mockImplementation(
      () => new Promise((resolve) => setTimeout(resolve, 50))
    );

    render(ResetPasswordForm, { props: { token: "reset-token-123" } });

    const newPassword = screen.getByLabelText(/new password/i);
    const confirmPassword = screen.getByLabelText(/confirm password/i);

    await fireEvent.input(newPassword, { target: { value: "StrongPass1!" } });
    await fireEvent.input(confirmPassword, { target: { value: "StrongPass1!" } });

    const submitButton = screen.getByRole("button", { name: /save new password/i });
    await fireEvent.click(submitButton);

    await waitFor(() => {
      expect(submitButton).toHaveTextContent(/saving/i);
    });
  });

  it("shows error when reset request fails", async () => {
    vi.mocked(resetPassword).mockRejectedValue(new Error("reset failed"));

    render(ResetPasswordForm, { props: { token: "reset-token-123" } });

    const newPassword = screen.getByLabelText(/new password/i);
    const confirmPassword = screen.getByLabelText(/confirm password/i);

    await fireEvent.input(newPassword, { target: { value: "StrongPass1!" } });
    await fireEvent.input(confirmPassword, { target: { value: "StrongPass1!" } });

    const form = document.querySelector(".reset-form");
    await fireEvent.submit(form!);

    await waitFor(() => {
      expect(screen.getByText(/reset failed/i)).toBeInTheDocument();
    });
  });
});
