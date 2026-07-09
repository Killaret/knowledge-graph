import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, cleanup } from '@testing-library/svelte';
import LoginForm from './LoginForm.svelte';

const mockLogin = vi.fn();
const mockLoginWithApiKey = vi.fn();

vi.mock('$lib/stores/auth.svelte.js', () => ({
  login: (...args: any[]) => mockLogin(...args),
  loginWithApiKey: (...args: any[]) => mockLoginWithApiKey(...args),
  isLoading: () => false,
  error: () => null
}));

vi.mock('$app/navigation', () => ({
  goto: vi.fn()
}));

describe('LoginForm', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    cleanup();
  });

  it('renders login form with default fields', () => {
    render(LoginForm);

    expect(screen.getByLabelText(/login/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/password/i)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /sign in/i })).toBeInTheDocument();
  });

  it('updates login and password fields', async () => {
    render(LoginForm);

    const loginInput = screen.getByLabelText(/login/i);
    const passwordInput = screen.getByLabelText(/password/i);

    await fireEvent.input(loginInput, { target: { value: 'testuser' } });
    await fireEvent.input(passwordInput, { target: { value: 'password123' } });

    expect(loginInput).toHaveValue('testuser');
    expect(passwordInput).toHaveValue('password123');
  });

  it('shows error when fields are empty', async () => {
    const { container } = render(LoginForm);

    const form = container.querySelector('form');
    await fireEvent.submit(form!);

    expect(screen.getByRole('alert')).toHaveTextContent(/please enter login and password/i);
  });

  it('submits login with valid credentials', async () => {
    mockLogin.mockResolvedValue(true);
    render(LoginForm);

    const loginInput = screen.getByLabelText(/login/i);
    const passwordInput = screen.getByLabelText(/password/i);
    const submitButton = screen.getByRole('button', { name: /sign in/i });

    await fireEvent.input(loginInput, { target: { value: 'testuser' } });
    await fireEvent.input(passwordInput, { target: { value: 'password123' } });
    await fireEvent.click(submitButton);

    expect(mockLogin).toHaveBeenCalledWith('testuser', 'password123');
  });

  it('has link to register page', () => {
    render(LoginForm);

    const registerLink = screen.getByRole('link', { name: /register/i });
    expect(registerLink).toHaveAttribute('href', '/auth/register');
  });

  it('has link to forgot password page', () => {
    render(LoginForm);

    const forgotLink = screen.getByRole('link', { name: /forgot password/i });
    expect(forgotLink).toHaveAttribute('href', '/auth/forgot-password');
  });
});
